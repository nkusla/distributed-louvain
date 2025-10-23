package actors

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"path/filepath"
	"os"

	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/crdt"
	"github.com/distributed-louvain/pkg/graphio"
	"github.com/distributed-louvain/pkg/graph"
	"github.com/distributed-louvain/pkg/messages"
)

type CoordinatorActor struct {
	*actor.BaseActor
	currentPhase      int
	iteration         int
	maxIterations     int
	prevModularity    float64
	completedActors   map[string]bool
	nodeset           *crdt.NodeSet
	nodeCommunityMap  map[int]int // maps node ID to its current community ID
}

func NewCoordinatorActor(pid actor.PID, system *actor.ActorSystem, maxIterations int) *CoordinatorActor {
	return &CoordinatorActor{
		BaseActor:        actor.NewBaseActor(pid, system, 1000),
		completedActors:  make(map[string]bool),
		currentPhase:     0,
		nodeset:          crdt.NewNodeSet(),
		iteration:        0,
		maxIterations:    maxIterations,
		nodeCommunityMap: make(map[int]int),
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
	log.Printf("[coordinator] Started")

	for {
		select {
		case <-c.Ctx.Done():
			log.Printf("[coordinator] Shutting down")
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
	case *messages.StartAlgorithmRequest:
		c.handleStartAlgorithmRequest(m)
	case *messages.InitialPartitionCreationComplete:
		c.handleInitialPartitionCreationComplete(m)
	case *messages.LocalOptimizationComplete:
		c.handleLocalOptimizationComplete(m)
	case *messages.EdgeAggregateComplete:
		c.handleEdgeAggregateComplete(m)
	case *messages.AggregationComplete:
		c.handleAggregationComplete(m)
	default:
		log.Printf("[coordinator] Received unknown message type: %s", msg.Type())
	}
}

func (c *CoordinatorActor) handleStartAlgorithmRequest(msg *messages.StartAlgorithmRequest) {
  dataPath := os.Getenv("DATA_PATH")
	if dataPath == "" {
		log.Printf("[coordinator] DATA_PATH is not set")
		return
	}

	log.Printf("[coordinator] Starting Louvain algorithm with data from %s", dataPath)

	edges, totalGraphWeight, err := graph.LoadGraphData(dataPath)
	if err != nil {
		log.Printf("[coordinator] Failed to load graph data: %v", err)
		return
	}
	for _, edge := range edges {
		c.nodeCommunityMap[edge.U] = edge.U
		c.nodeCommunityMap[edge.V] = edge.V
	}

	partitionPIDs := c.System.GetActors(actor.PartitionType)
	numPartitions := len(partitionPIDs)

	if numPartitions == 0 {
		log.Printf("[coordinator] No partition actors available")
		return
	}

	partitionEdges := make([][]graph.Edge, numPartitions)
	for i := range partitionEdges {
		partitionEdges[i] = make([]graph.Edge, 0)
	}

	for _, edge := range edges {
		partitionU := edge.U % numPartitions
		partitionV := edge.V % numPartitions

		// Send edge to partition containing node U
		partitionEdges[partitionU] = append(partitionEdges[partitionU], edge)

		// Only send reversed edge if nodes are in different partitions
		if partitionU != partitionV {
			reversedEdge := graph.NewEdge(edge.V, edge.U, edge.W)
			partitionEdges[partitionV] = append(partitionEdges[partitionV], reversedEdge)
		}
	}

	for i, pid := range partitionPIDs {
		log.Printf("[coordinator] Sending %d edges to partition %s", len(partitionEdges[i]), pid)
		c.Send(pid, &messages.InitialPartitionCreation{
			Edges: partitionEdges[i],
			TotalGraphWeight: totalGraphWeight,
		})
	}
}

func (c *CoordinatorActor) handleInitialPartitionCreationComplete(msg *messages.InitialPartitionCreationComplete) {
	c.completedActors[msg.Sender.String()] = true

	if len(c.completedActors) == len(c.System.GetActors(actor.PartitionType)) {
		log.Printf("[coordinator] Initial partition creation complete")
		c.startPhase1()
	}
}

func (c *CoordinatorActor) startPhase1() {
	c.currentPhase = 1
	c.completedActors = make(map[string]bool)

	c.nodeset.Project()

	// Update node community map with the projected transitions
	for nodeID, communityID := range c.nodeCommunityMap {
		for _, transition := range c.nodeset.GetAll() {
			if communityID == transition.NodeID {
				c.nodeCommunityMap[nodeID] = transition.CommunityID
				break
			}
		}
	}

	c.nodeset.Clear()

	log.Printf("[coordinator] Starting Phase 1: Local Optimization")

	for _, pid := range c.System.GetActors(actor.PartitionType) {
		c.Send(pid, &messages.StartPhase1{})
	}
}

func (c *CoordinatorActor) handleLocalOptimizationComplete(msg *messages.LocalOptimizationComplete) {
	c.completedActors[msg.Sender.String()] = true
	c.nodeset.Merge(msg.NodeSet)

	log.Printf("[coordinator] Local optimization complete from %s", msg.Sender)

	if len(c.completedActors) == len(c.System.GetActors(actor.PartitionType)) {
		c.checkConvergence()
	}
}

func (c *CoordinatorActor) checkConvergence() {
	transitions := c.nodeset.GetAll()

	// Calculate modularity for this iteration only (don't accumulate)
	currentModularity := 0.0
	for _, transition := range transitions {
		currentModularity += transition.ModularityDelta
	}

	improvement := currentModularity - c.prevModularity
	log.Printf("[coordinator] Iteration %d complete. Modularity: %.6f (improvement: %.6f)",
		c.iteration, currentModularity, improvement)

	if improvement < 1e-6 || c.iteration >= c.maxIterations {
		log.Printf("[coordinator] Algorithm converged!")
		c.prevModularity = currentModularity
		c.completeAlgorithm()
		return
	}

	c.prevModularity = currentModularity
	c.startPhase2()
}

func (c *CoordinatorActor) startPhase2() {
	c.currentPhase = 2
	c.completedActors = make(map[string]bool)

	log.Printf("[coordinator] Starting Phase 2: Aggregation")

	c.System.Broadcast(c.PID(), actor.PartitionType, &messages.StartPhase2{})

	c.System.Broadcast(c.PID(), actor.AggregatorType, &messages.StartPhase2{})
}

func (c *CoordinatorActor) handleEdgeAggregateComplete(msg *messages.EdgeAggregateComplete) {
	c.completedActors[msg.Sender.String()] = true

	log.Printf("[coordinator] Edge aggregation complete from %s", msg.Sender)

	if len(c.completedActors) == len(c.System.GetActors(actor.PartitionType)) {
		log.Printf("[coordinator] All partitions completed edge aggregation, telling aggregators to complete")
		// Tell all aggregators to complete their aggregation
		c.completedActors = make(map[string]bool)
		for _, pid := range c.System.GetActors(actor.AggregatorType) {
			c.Send(pid, &messages.CompleteAggregation{})
		}
	}
}

func (c *CoordinatorActor) handleAggregationComplete(msg *messages.AggregationComplete) {
	c.completedActors[msg.Sender.String()] = true

	log.Printf("[coordinator] Aggregation complete from %s", msg.Sender)

	if len(c.completedActors) == len(c.System.GetActors(actor.AggregatorType)) {
		c.iteration++
		c.startPhase1()
	}
}

func (c *CoordinatorActor) completeAlgorithm() {
	msg := &messages.AlgorithmComplete{
		FinalModularity: c.prevModularity,
		Iterations:      c.iteration,
	}

	c.System.Broadcast(c.PID(), actor.PartitionType, msg)
	c.System.Broadcast(c.PID(), actor.AggregatorType, msg)

	// Save community assignments to CSV
	c.saveCommunityAssignments()

	logStr := fmt.Sprintf(
		"\n====== ALGORITHM COMPLETE ======\n"+
		"Final Modularity: %.6f\n"+
		"Total Iterations: %d / %d\n"+
		"================================",
		c.prevModularity, c.iteration, c.maxIterations)

	log.Println(logStr)
}

func (c *CoordinatorActor) saveCommunityAssignments() {
	filePath := filepath.Join("output", "community.csv")
	headers := []string{"nodeId", "communityId"}

	if err := graphio.WriteIntMapToCSV(filePath, headers, c.nodeCommunityMap); err != nil {
		log.Printf("[coordinator] Error saving community assignments: %v", err)
		return
	}

	log.Printf("[coordinator] Community assignments saved to %s", filePath)
}

func (c *CoordinatorActor) getTargetPartitionForNode(nodeID int) (actor.PID, error) {
	partitions := c.System.GetActors(actor.PartitionType)
	if len(partitions) == 0 {
		return actor.PID{}, fmt.Errorf("no partition actors available")
	}

	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%d", nodeID)))
	hash := h.Sum32()

	targetIndex := int(hash) % len(partitions)
	return partitions[targetIndex], nil
}