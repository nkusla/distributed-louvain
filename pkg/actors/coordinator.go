package actors

import (
	"context"
	"log"

	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/crdt"
	"github.com/distributed-louvain/pkg/graph"
	"github.com/distributed-louvain/pkg/messages"
)

type CoordinatorActor struct {
	*actor.BaseActor
	currentPhase    int
	iteration       int
	maxIterations   int
	totalModularity float64
	prevModularity  float64
	completedActors map[string]bool
	nodeset         *crdt.NodeSet
}

func NewCoordinatorActor(pid actor.PID, system *actor.ActorSystem, maxIterations int) *CoordinatorActor {
	return &CoordinatorActor{
		BaseActor:       actor.NewBaseActor(pid, system, 1000),
		completedActors: make(map[string]bool),
		currentPhase:    0,
		nodeset:         crdt.NewNodeSet(),
		iteration:       0,
		maxIterations:   maxIterations,
	}
}

func (c *CoordinatorActor) Start(ctx context.Context) {
	c.Ctx, c.Cancel = context.WithCancel(ctx)
	c.Wg.Add(1)

	go func() {
		defer c.Wg.Done()
		c.run()
	}()
}

func (c *CoordinatorActor) run() {
	log.Printf("[Coordinator] Started")

	for {
		select {
		case <-c.Ctx.Done():
			log.Printf("[Coordinator] Shutting down")
			return
		case msg, ok := <-c.Mailbox.Receive():
			if !ok {
				return
			}
			c.Receive(c.Ctx, msg)
		}
	}
}

func (c *CoordinatorActor) Receive(ctx context.Context, msg actor.Message) {
	switch m := msg.(type) {
	case *messages.InitialPartitionCreationComplete:
		c.handleInitialPartitionCreationComplete(m)
	case *messages.LocalOptimizationComplete:
		c.handleLocalOptimizationComplete(m)
	case *messages.AggregationComplete:
		c.handleAggregationComplete(m)
	case *messages.RedistributionComplete:
		c.handleRedistributionComplete(m)
	case *messages.Phase2Complete:
		c.handlePhase2Complete(m)
	default:
		log.Printf("[Coordinator] Received unknown message type: %s", msg.Type())
	}
}

func (c *CoordinatorActor) StartAlgorithm(edges []graph.Edge, totalGraphWeight int) {
	log.Printf("[Coordinator] Starting Louvain algorithm")

	partitionPIDs := c.System.GetActors(actor.PartitionType)
	numPartitions := len(partitionPIDs)

	if numPartitions == 0 {
		log.Printf("[Coordinator] No partition actors available")
		return
	}

	partitionEdges := make([][]graph.Edge, numPartitions)
	for i := range partitionEdges {
		partitionEdges[i] = make([]graph.Edge, 0)
	}

	for _, edge := range edges {
		partitionU := edge.U % numPartitions
		partitionEdges[partitionU] = append(partitionEdges[partitionU], edge)

		partitionV := edge.V % numPartitions
		reversedEdge := graph.NewEdge(edge.V, edge.U, edge.W)
		partitionEdges[partitionV] = append(partitionEdges[partitionV], reversedEdge)
	}

	for i, pid := range partitionPIDs {
		log.Printf("[Coordinator] Sending %d edges to partition %s", len(partitionEdges[i]), pid)
		c.Send(pid, &messages.InitialPartitionCreation{
			Edges: partitionEdges[i],
			TotalGraphWeight: totalGraphWeight,
		})
	}
}

func (c *CoordinatorActor) handleInitialPartitionCreationComplete(msg *messages.InitialPartitionCreationComplete) {
	c.completedActors[msg.Sender.String()] = true

	if len(c.completedActors) == len(c.System.GetActors(actor.PartitionType)) {
		log.Printf("[Coordinator] Initial partition creation complete")
		c.startPhase1()
	}
}

func (c *CoordinatorActor) startPhase1() {
	c.currentPhase = 1
	c.completedActors = make(map[string]bool)
	c.nodeset.Clear()

	log.Printf("[Coordinator] Starting Phase 1: Local Optimization")

	for _, pid := range c.System.GetActors(actor.PartitionType) {
		c.Send(pid, &messages.StartPhase1{})
	}
}

func (c *CoordinatorActor) handleLocalOptimizationComplete(msg *messages.LocalOptimizationComplete) {
	c.completedActors[msg.Sender.String()] = true
	c.nodeset.Merge(msg.NodeSet)

	log.Printf("[Coordinator] Local optimization complete from %s", msg.Sender)

	if len(c.completedActors) == len(c.System.GetActors(actor.PartitionType)) {
		c.checkConvergence()
	}
}

func (c *CoordinatorActor) checkConvergence() {
	transitions := c.nodeset.GetAll()
	for _, transition := range transitions {
		c.totalModularity += transition.ModularityDelta
	}

	improvement := c.totalModularity - c.prevModularity
	log.Printf("[Coordinator] Iteration %d complete. Modularity: %.6f (improvement: %.6f)",
		c.iteration, c.totalModularity, improvement)

	if improvement < 1e-6 || c.iteration >= c.maxIterations {
		log.Printf("[Coordinator] Algorithm converged!")
		c.completeAlgorithm()
		return
	}

	c.prevModularity = c.totalModularity
	c.startPhase2()
}

func (c *CoordinatorActor) startPhase2() {
	c.currentPhase = 2
	c.completedActors = make(map[string]bool)

	log.Printf("[Coordinator] Starting Phase 2: Aggregation")

	c.System.Broadcast(c.PID(), actor.PartitionType, &messages.StartPhase2{})

	c.System.Broadcast(c.PID(), actor.AggregatorType, &messages.StartPhase2{})
}

func (c *CoordinatorActor) handleAggregationComplete(msg *messages.AggregationComplete) {
	c.completedActors[msg.Sender.String()] = true

	log.Printf("[Coordinator] Aggregation complete from %s", msg.Sender)

	if len(c.completedActors) == len(c.System.GetActors(actor.AggregatorType)) {
		c.startRedistribution()
	}
}

func (c *CoordinatorActor) startRedistribution() {
	log.Printf("[Coordinator] Starting redistribution")

	c.iteration++
}

func (c *CoordinatorActor) handleRedistributionComplete(msg *messages.RedistributionComplete) {
	c.completedActors[msg.Sender.String()] = true

	log.Printf("[Coordinator] Redistribution complete from %s", msg.Sender)

	if len(c.completedActors) == len(c.System.GetActors(actor.PartitionType)) {
		c.iteration++
		c.startPhase1()
	}
}

func (c *CoordinatorActor) handlePhase2Complete(msg *messages.Phase2Complete) {
	c.completedActors[msg.Sender.String()] = true

	log.Printf("[Coordinator] Phase 2 complete from %s", msg.Sender)
}

func (c *CoordinatorActor) completeAlgorithm() {
	msg := &messages.AlgorithmComplete{
		FinalModularity: c.totalModularity,
		Iterations:      c.iteration,
	}

	c.System.Broadcast(c.PID(), actor.PartitionType, msg)
	c.System.Broadcast(c.PID(), actor.AggregatorType, msg)

	log.Println("\n=== ALGORITHM COMPLETE ===")
	log.Printf("Final Modularity: %.6f\n", c.totalModularity)
	log.Printf("Total Iterations: %d\n", c.iteration)
	log.Println("==========================")
}
