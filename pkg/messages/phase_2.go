package messages

import "github.com/distributed-louvain/pkg/actor"
import "github.com/distributed-louvain/pkg/graph"

// Phase 2: Aggregation Messages

type StartPhase2 struct{}
func (m *StartPhase2) Type() string { return "StartPhase2" }

type EdgeAggregate struct {
	CommunityU int `json:"community_u"`
	CommunityV int `json:"community_v"`
	Weight     int `json:"weight"`
	Sender     actor.PID `json:"sender"`
}
func (m *EdgeAggregate) Type() string { return "EdgeAggregate" }

type EdgeAggregateComplete struct {
	Sender actor.PID `json:"sender"`
}
func (m *EdgeAggregateComplete) Type() string { return "EdgeAggregateComplete" }

type AggregationResult struct {
	Edges 	[]graph.Edge `json:"edges"`
}
func (m *AggregationResult) Type() string { return "AggregationResult" }

type AggregationComplete struct {
	Sender actor.PID `json:"sender"`
}
func (m *AggregationComplete) Type() string { return "AggregationComplete" }

type CompleteAggregation struct{}
func (m *CompleteAggregation) Type() string { return "CompleteAggregation" }