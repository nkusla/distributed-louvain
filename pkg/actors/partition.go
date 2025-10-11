package actors

import (
	"context"
	"log"
	"sync"

	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/crdt"
	"github.com/distributed-louvain/pkg/graph"
	"github.com/distributed-louvain/pkg/messages"
)

type PartitionActor struct {
	*actor.BaseActor
	nodeSet           *crdt.NodeSet
	currentPhase      int
	partition         *graph.Graph
	totalGraphWeight  int
	coordinator       actor.PID
	mu                sync.RWMutex
	totalNodePairs    int
	processedNodePairs int
}

func NewPartitionActor(pid actor.PID, system *actor.ActorSystem, coordinatorPID actor.PID) *PartitionActor {
	return &PartitionActor{
		BaseActor:       actor.NewBaseActor(pid, system, 1000),
		coordinator:     coordinatorPID,
		nodeSet:         crdt.NewNodeSet(),
		currentPhase:    0,
		partition:       graph.NewGraph(),
		totalGraphWeight: 0,
	}
}

func (p *PartitionActor) Start(ctx context.Context) {
	p.Ctx, p.Cancel = context.WithCancel(ctx)
	p.Wg.Add(1)

	go func() {
		defer p.Wg.Done()
		p.run()
	}()
}

func (p *PartitionActor) run() {
	log.Printf("[Partition %d] Started", p.PID().ActorID)

	for {
		select {
		case <-p.Ctx.Done():
			log.Printf("[Partition %d] Shutting down", p.PID().ActorID)
			return
		case msg, ok := <-p.Mailbox.Receive():
			if !ok {
				return
			}
			p.Receive(p.Ctx, msg)
		}
	}
}

func (p *PartitionActor) Receive(ctx context.Context, msg actor.Message) {
	switch m := msg.(type) {
	case *messages.InitialPartitionCreation:
		p.handleInitialPartitionCreation(m)
	case *messages.StartPhase1:
		p.handleStartPhase1()
	case *messages.DegreeRequest:
		p.handleDegreeRequest(m)
	case *messages.DegreeResponse:
		p.handleDegreeResponse(m)
	case *messages.AlgorithmComplete:
		p.handleAlgorithmComplete(m)
	default:
		log.Printf("[Partition %d] Received unknown message type: %s", p.PID().ActorID, msg.Type())
	}
}

func (p *PartitionActor) handleInitialPartitionCreation(msg *messages.InitialPartitionCreation) {
	p.partition.AddEdges(msg.Edges)
	p.totalGraphWeight = msg.TotalGraphWeight
	log.Printf("[Partition %d] Received %d edges", p.PID().ActorID, len(msg.Edges))
	p.Send(p.coordinator, &messages.InitialPartitionCreationComplete{
		Sender: p.PID(),
	})
}

func (p *PartitionActor) handleStartPhase1() {
	p.currentPhase = 1
	p.totalNodePairs = len(p.partition.Adj)
	p.processedNodePairs = 0
	p.totalNodePairs = 0

	log.Printf("[Partition %s] Starting Phase 1", p.PID().ActorID)

	for nodeId, neighbors := range p.partition.Adj {
		for _, neighbor := range neighbors {
			p.totalNodePairs++

			if _, exists := p.partition.Degree[neighbor.NodeID]; !exists {
				p.System.Broadcast(p.PID(), actor.PartitionType, &messages.DegreeRequest{
					NodeID: nodeId,
					NeighborID: neighbor.NodeID,
					Sender: p.PID(),
				})

				continue
			}

			modularityDelta := p.calculateModularityDelta(nodeId, neighbor.NodeID)

			if modularityDelta > 0 {
				p.nodeSet.Add(nodeId, neighbor.NodeID, modularityDelta)
			}

			p.processedNodePairs++
		}
	}
}

func (p *PartitionActor) handleDegreeRequest(msg *messages.DegreeRequest) {
	p.Send(msg.Sender, &messages.DegreeResponse{
		NodeID: msg.NodeID,
		NeighborID: msg.NeighborID,
		Degree: p.partition.Degree[msg.NeighborID],
	})
}

func (p *PartitionActor) handleDegreeResponse(msg *messages.DegreeResponse) {
	p.partition.Degree[msg.NeighborID] = msg.Degree

	modularityDelta := p.calculateModularityDelta(msg.NodeID, msg.NeighborID)

	if modularityDelta > 0 {
		p.nodeSet.Add(msg.NodeID, msg.NeighborID, modularityDelta)
	}

	p.processedNodePairs++
	p.checkLocalOptimizationComplete()
}


func (p *PartitionActor) calculateModularityDelta(nodeU int, nodeV int) float64 {
	sumIn := p.partition.GetWeight(nodeU, nodeV)
	sumTot := p.partition.Degree[nodeU] + p.partition.Degree[nodeV] - 2 * sumIn

	modularityDelta := float64(sumIn) / float64(2 * p.totalGraphWeight) - float64(sumTot) / float64(4 * p.totalGraphWeight * p.totalGraphWeight)

	return modularityDelta
}

func (p *PartitionActor) checkLocalOptimizationComplete() {
	if p.processedNodePairs < p.totalNodePairs{
		return
	}

	log.Printf("[Partition %s] Local optimization complete. Processed %d node pairs",
			p.PID().ActorID, p.processedNodePairs)

	p.Send(p.coordinator, &messages.LocalOptimizationComplete{
			NodeSet: p.nodeSet,
			Sender: p.PID(),
	})
}

func (p *PartitionActor) handleAlgorithmComplete(msg *messages.AlgorithmComplete) {
	log.Printf("[Partition %s] Algorithm complete! Final modularity: %.6f",
		p.PID().ActorID, msg.FinalModularity)

	p.Stop()
}