import "github.com/distributed-louvain/pkg/actor"

// Phase 2: Aggregation Messages

type StartPhase2 struct{}
func (m *StartPhase2) Type() string { return "StartPhase2" }

type EdgeAggregate struct {
	CommunityU int
	CommunityV int
	Weight     float64
	PartitionID int
	Sender     actor.PID
}
func (m *EdgeAggregate) Type() string { return "EdgeAggregate" }

type AggregationComplete struct {
	AggregatorID int
	SuperEdges   []SuperEdge
	Sender       actor.PID
}
func (m *AggregationComplete) Type() string { return "AggregationComplete" }

type StartRedistribution struct {
	NewGraph GraphData
}
func (m *StartRedistribution) Type() string { return "StartRedistribution" }

type RedistributionComplete struct {
	PartitionID int
	NodesCount  int
	EdgesCount  int
	Sender      actor.PID
}
func (m *RedistributionComplete) Type() string { return "RedistributionComplete" }

// Helper types

type SuperEdge struct {
	CommunityU int
	CommunityV int
	Weight     float64
}

type GraphData struct {
	Nodes []NodeData
	Edges []EdgeData
}

type NodeData struct {
	ID        int
	Community int
}

type EdgeData struct {
	From   int
	To     int
	Weight float64
}