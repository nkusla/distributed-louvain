package actors

import (
	"context"
	"fmt"
	"hash/fnv"
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
	log.Printf("[%s] Started", p.PID().ActorID)

	for {
		select {
		case <-p.Ctx.Done():
			log.Printf("[%s] Shutting down", p.PID().ActorID)
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
	case *messages.LocalOptimizationComplete:
		p.handleLocalOptimizationComplete(m)
	case *messages.AggregationResult:
		p.handleAggregationResult(m)
	case *messages.StartPhase2:
		p.handleStartPhase2()
	case *messages.AlgorithmComplete:
		p.handleAlgorithmComplete(m)
	default:
		log.Printf("[%s] Received unknown message type: %s", p.PID().ActorID, msg.Type())
	}
}

func (p *PartitionActor) handleInitialPartitionCreation(msg *messages.InitialPartitionCreation) {
	p.partition.AddEdges(msg.Edges)
	p.totalGraphWeight = msg.TotalGraphWeight
	log.Printf("[%s] Received %d edges", p.PID().ActorID, len(msg.Edges))
	p.Send(p.coordinator, &messages.InitialPartitionCreationComplete{
		Sender: p.PID(),
	})
}

func (p *PartitionActor) handleStartPhase1() {
	p.currentPhase = 1
	p.processedNodePairs = 0
	p.totalNodePairs = 0

	log.Printf("[%s] Starting Phase 1", p.PID().ActorID)

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

	p.checkLocalOptimizationComplete()
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

	log.Printf("[%s] Local optimization complete. Processed %d node pairs",
			p.PID().ActorID, p.processedNodePairs)

	p.System.Broadcast(p.PID(), actor.PartitionType, &messages.LocalOptimizationComplete{
		NodeSet: p.nodeSet,
		Sender: p.PID(),
	})

	p.Send(p.coordinator, &messages.LocalOptimizationComplete{
		NodeSet: p.nodeSet,
		Sender: p.PID(),
	})
}

func (p *PartitionActor) handleLocalOptimizationComplete(msg *messages.LocalOptimizationComplete) {
	p.nodeSet.Merge(msg.NodeSet)
}

func (p *PartitionActor) handleStartPhase2() {
	p.currentPhase = 2

	for nodeID := range p.partition.Adj {
		for _, neighbor := range p.partition.Adj[nodeID] {

			communityU := nodeID
			communityV := neighbor.NodeID
			weight := neighbor.Weight

			if transition, exists := p.nodeSet.Get(nodeID); exists {
				communityU = transition.CommunityID
			}
			if transition, exists := p.nodeSet.Get(neighbor.NodeID); exists {
				communityV = transition.CommunityID
			}

			if communityU == communityV {
				continue
			}

			targetAggregator, err := p.getTargetAggregator(communityU, communityV)
			if err != nil {
				log.Printf("[%s] Error getting target aggregator: %v", p.PID().ActorID, err)
				continue
			}

			p.Send(targetAggregator, &messages.EdgeAggregate{
				CommunityU: communityU,
				CommunityV: communityV,
				Weight: weight,
				Sender: p.PID(),
			})

			log.Printf("[%s] Sent edge (%d,%d) weight %d to aggregator %s",
				p.PID().ActorID, communityU, communityV, weight, targetAggregator.ActorID)
		}
	}

	p.partition = graph.NewGraph()
}

func (p *PartitionActor) handleAggregationResult(msg *messages.AggregationResult) {
	p.partition.AddEdges(msg.Edges)
}

func (p *PartitionActor) handleAlgorithmComplete(msg *messages.AlgorithmComplete) {
	log.Printf("[%s] Algorithm complete! Final modularity: %.6f",
		p.PID().ActorID, msg.FinalModularity)

	p.Stop()
}

func (p *PartitionActor) getTargetAggregator(communityU, communityV int) (actor.PID, error) {
	aggregators := p.System.GetActors(actor.AggregatorType)
	if len(aggregators) == 0 {
		return actor.PID{}, fmt.Errorf("no aggregator actors available")
	}

	u, v := communityU, communityV
	if u > v {
		u, v = v, u
	}

	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%d-%d", u, v)))
	hash := h.Sum32()

	targetIndex := int(hash) % len(aggregators)
	return aggregators[targetIndex], nil
}