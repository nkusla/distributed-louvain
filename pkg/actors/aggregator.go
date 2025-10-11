package actors

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/messages"
)

type AggregatorActor struct {
	*actor.BaseActor
	coordinator   actor.PID
	edgeWeights   map[string]float64 // Map of edge key -> total weight
}

func NewAggregatorActor(pid actor.PID, system *actor.ActorSystem, coordinator actor.PID, numPartitions int) *AggregatorActor {
	return &AggregatorActor{
		BaseActor:    actor.NewBaseActor(pid, system, 1000),
		coordinator:  coordinator,
		edgeWeights:  make(map[string]float64),
	}
}

func (a *AggregatorActor) Start(ctx context.Context) {
	a.Ctx, a.Cancel = context.WithCancel(ctx)
	a.Wg.Add(1)

	go func() {
		defer a.Wg.Done()
		a.run()
	}()
}

func (a *AggregatorActor) run() {
	log.Printf("[Aggregator %s] Started", a.PID().ActorID)

	for {
		select {
		case <-a.Ctx.Done():
			log.Printf("[Aggregator %s] Shutting down", a.PID().ActorID)
			return
		case msg, ok := <-a.Mailbox.Receive():
			if !ok {
				return
			}
			a.Receive(a.Ctx, msg)
		}
	}
}

func (a *AggregatorActor) Receive(ctx context.Context, msg actor.Message) {
	switch m := msg.(type) {
	case *messages.StartPhase2:
		a.resetCounters()
	case *messages.EdgeAggregate:
		a.handleEdgeAggregate(m)
	case *messages.AlgorithmComplete:
		a.handleAlgorithmComplete(m)
	default:
		log.Printf("[Aggregator %s] Received unknown message type: %s", a.PID().ActorID, msg.Type())
	}
}

func (a *AggregatorActor) handleEdgeAggregate(msg *messages.EdgeAggregate) {
	edgeKey := makeEdgeKey(msg.CommunityU, msg.CommunityV)

	a.edgeWeights[edgeKey] += msg.Weight

	log.Printf("[Aggregator %s] Received edge (%d,%d) weight %.2f from partition %d, total: %.2f",
		a.PID().ActorID, msg.CommunityU, msg.CommunityV, msg.Weight, msg.PartitionID, a.edgeWeights[edgeKey])
}

func (a *AggregatorActor) resetCounters() {
	a.edgeWeights = make(map[string]float64)
	log.Printf("[Aggregator %s] Reset edge weights for new aggregation phase", a.PID().ActorID)
}

func (a *AggregatorActor) CompleteAggregation() {
	superEdges := make([]messages.SuperEdge, 0, len(a.edgeWeights))
	for edgeKey, weight := range a.edgeWeights {
		var commU, commV int
		parseEdgeKey(edgeKey, &commU, &commV)

		superEdges = append(superEdges, messages.SuperEdge{
			CommunityU: commU,
			CommunityV: commV,
			Weight:     weight,
		})
	}

	log.Printf("[Aggregator %s] Aggregation complete. Total super-edges: %d", a.PID().ActorID, len(superEdges))

	a.Send(a.coordinator, &messages.AggregationComplete{
		ActorID:    a.PID().ActorID,
		SuperEdges: superEdges,
		Sender:     a.PID(),
	})
}

func (a *AggregatorActor) handleAlgorithmComplete(msg *messages.AlgorithmComplete) {
	log.Printf("[Aggregator %s] Algorithm complete! Final modularity: %.6f",
		a.PID().ActorID, msg.FinalModularity)

	a.Stop()
}

func makeEdgeKey(u, v int) string {
	if u > v {
		u, v = v, u
	}

	return fmt.Sprintf("%d-%d", u, v)
}

func parseEdgeKey(key string, u, v *int) {
	parts := strings.Split(key, "-")
	if len(parts) != 2 {
		return // or handle error appropriately
	}

	*u, _ = strconv.Atoi(parts[0])
	*v, _ = strconv.Atoi(parts[1])
}
