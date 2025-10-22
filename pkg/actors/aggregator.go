package actors

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"strconv"
	"strings"

	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/graph"
	"github.com/distributed-louvain/pkg/messages"
)

type AggregatorActor struct {
	*actor.BaseActor
	coordinator   actor.PID
	edgeWeights   map[string]int // Map of edge key -> total weight
}

func NewAggregatorActor(pid actor.PID, system *actor.ActorSystem, coordinator actor.PID, numPartitions int) *AggregatorActor {
	return &AggregatorActor{
		BaseActor:    actor.NewBaseActor(pid, system, 1000),
		coordinator:  coordinator,
		edgeWeights:  make(map[string]int),
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
	log.Printf("[%s] Started", a.PID().ActorID)

	for {
		select {
		case <-a.Ctx.Done():
			log.Printf("[%s] Shutting down", a.PID().ActorID)
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
	case *messages.CompleteAggregation:
		a.CompleteAggregation()
	case *messages.AlgorithmComplete:
		a.handleAlgorithmComplete(m)
	default:
		log.Printf("[%s] Received unknown message type: %s", a.PID().ActorID, msg.Type())
	}
}

func (a *AggregatorActor) handleEdgeAggregate(msg *messages.EdgeAggregate) {
	edgeKey := makeEdgeKey(msg.CommunityU, msg.CommunityV)

	a.edgeWeights[edgeKey] += msg.Weight

	log.Printf("[%s] Received community edge (%d,%d) weight %d from %s, total: %d",
		a.PID().ActorID, msg.CommunityU, msg.CommunityV, msg.Weight, msg.Sender.ActorID, a.edgeWeights[edgeKey])
}

func (a *AggregatorActor) resetCounters() {
	a.edgeWeights = make(map[string]int)
	log.Printf("[%s] Reset edge weights for new aggregation phase", a.PID().ActorID)
}

func (a *AggregatorActor) CompleteAggregation() {
	partitionEdges := make(map[actor.PID][]graph.Edge)
	for edgeKey, weight := range a.edgeWeights {
		var commU, commV int
		if err := parseEdgeKey(edgeKey, &commU, &commV); err != nil {
			log.Printf("[%s] Error parsing edge key: %s", a.PID().ActorID, err)
			continue
		}

		targetPartition, err := a.getTargetPartition(commU)
		if err != nil {
			log.Printf("[%s] Error getting target partition for u=%d: %v", a.PID().ActorID, commU, err)
			continue
		}

		partitionEdges[targetPartition] = append(partitionEdges[targetPartition], graph.Edge{U: commU, V: commV, W: weight})

		targetPartition, err = a.getTargetPartition(commV)
		if err != nil {
			log.Printf("[%s] Error getting target partition for v=%d: %v", a.PID().ActorID, commV, err)
			continue
		}

		partitionEdges[targetPartition] = append(partitionEdges[targetPartition], graph.Edge{U: commV, V: commU, W: weight})
	}

	for partitionPID, edges := range partitionEdges {
		a.Send(partitionPID, &messages.AggregationResult{
			Edges: edges,
		})
	}

	a.Send(a.coordinator, &messages.AggregationComplete{
		Sender: a.PID(),
	})
}

func (a *AggregatorActor) handleAlgorithmComplete(msg *messages.AlgorithmComplete) {
	log.Printf("[%s] Algorithm complete! Final modularity: %.6f",
		a.PID().ActorID, msg.FinalModularity)
}

func (a *AggregatorActor) getTargetPartition(u int) (actor.PID, error) {
	partitions := a.System.GetActors(actor.PartitionType)
	if len(partitions) == 0 {
		return actor.PID{}, fmt.Errorf("no partition actors available")
	}

	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%d", u)))
	hash := h.Sum32()

	targetIndex := int(hash) % len(partitions)
	return partitions[targetIndex], nil
}

func makeEdgeKey(u, v int) string {
	if u > v {
		u, v = v, u
	}

	return fmt.Sprintf("%d-%d", u, v)
}

func parseEdgeKey(key string, u, v *int) error {
	parts := strings.Split(key, "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid edge key: %s", key)
	}

	*u, _ = strconv.Atoi(parts[0])
	*v, _ = strconv.Atoi(parts[1])
	return nil
}
