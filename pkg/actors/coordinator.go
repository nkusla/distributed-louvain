package actors

import (
	"context"
	"log"

	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/messages"
	"github.com/distributed-louvain/pkg/crdt"
)

type CoordinatorActor struct {
	*actor.BaseActor
	currentPhase    int
	iteration       int
	totalModularity float64
	prevModularity  float64
	completedActors map[string]bool
	nodeset         *crdt.NodeSet
}

func NewCoordinatorActor(pid actor.PID, system *actor.ActorSystem) *CoordinatorActor {
	return &CoordinatorActor{
		BaseActor:       actor.NewBaseActor(pid, system, 1000),
		completedActors: make(map[string]bool),
		currentPhase:    0,
		nodeset:         crdt.NewNodeSet(),
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
	case *messages.Phase1Complete:
		c.handlePhase1Complete(m)
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

func (c *CoordinatorActor) StartAlgorithm() {
	log.Printf("[Coordinator] Starting Louvain algorithm - Iteration %d", c.iteration)
	c.startPhase1()
}

func (c *CoordinatorActor) startPhase1() {
	c.currentPhase = 1
	c.completedActors = make(map[string]bool)
	c.nodeset.Clear()

	log.Printf("[Coordinator] Starting Phase 1: Local Optimization")

	// Broadcast to all partition actors
	for _, pid := range c.System.GetActors(actor.PartitionType) {
		c.Send(pid, &messages.StartPhase1{})
	}
}

func (c *CoordinatorActor) handlePhase1Complete(msg *messages.Phase1Complete) {
	c.completedActors[msg.Sender.String()] = true
	c.totalModularity += msg.ModularityGain

	log.Printf("[Coordinator] Phase 1 complete from %s (gain: %.6f)", msg.Sender, msg.ModularityGain)

	// Check if all partition actors have completed
	if len(c.completedActors) == len(c.System.GetActors(actor.PartitionType)) {
		c.checkConvergence()
	}
}

func (c *CoordinatorActor) checkConvergence() {
	improvement := c.totalModularity - c.prevModularity
	log.Printf("[Coordinator] Iteration %d complete. Modularity: %.6f (improvement: %.6f)",
		c.iteration, c.totalModularity, improvement)

	if improvement < 1e-6 {
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

	// Tell partition actors to aggregate
	for _, pid := range c.System.GetActors(actor.PartitionType) {
		c.Send(pid, &messages.StartPhase2{})
	}
}

func (c *CoordinatorActor) handleAggregationComplete(msg *messages.AggregationComplete) {
	c.completedActors[msg.Sender.String()] = true

	log.Printf("[Coordinator] Aggregation complete from %s", msg.Sender)

	// Check if all aggregators have completed
	if len(c.completedActors) == len(c.System.GetActors(actor.AggregatorType)) {
		c.startRedistribution()
	}
}

func (c *CoordinatorActor) startRedistribution() {
	log.Printf("[Coordinator] Starting redistribution")

	// TODO: Collect super-edges from aggregators and redistribute to partition actors
	// For now, just start next iteration
	c.iteration++
	c.startPhase1()
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

	for _, pid := range c.System.GetActors(actor.PartitionType) {
		c.Send(pid, msg)
	}
	for _, pid := range c.System.GetActors(actor.AggregatorType) {
		c.Send(pid, msg)
	}

	log.Println("\n=== ALGORITHM COMPLETE ===")
	log.Printf("Final Modularity: %.6f\n", c.totalModularity)
	log.Printf("Total Iterations: %d\n", c.iteration)
	log.Println("==========================")
}
