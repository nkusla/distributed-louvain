package messages

import "github.com/distributed-louvain/pkg/actor"
import "github.com/distributed-louvain/pkg/graph"

// Phase 2: Aggregation Messages

type StartPhase2 struct{}
func (m *StartPhase2) Type() string { return "StartPhase2" }

type EdgeAggregate struct {
	CommunityU int
	CommunityV int
	Weight     int
	Sender     actor.PID
}
func (m *EdgeAggregate) Type() string { return "EdgeAggregate" }

type EdgeAggregateComplete struct {
	Sender actor.PID
}
func (m *EdgeAggregateComplete) Type() string { return "EdgeAggregateComplete" }

type AggregationResult struct {
	Edges 	[]graph.Edge
}
func (m *AggregationResult) Type() string { return "AggregationResult" }

type AggregationComplete struct {
	Sender actor.PID
}
func (m *AggregationComplete) Type() string { return "AggregationComplete" }
