package messages

import (
	"github.com/distributed-louvain/pkg/graph"
	"github.com/distributed-louvain/pkg/actor"
)

type InitialPartitionCreation struct{
	Edges []graph.Edge `json:"edges"`
	TotalGraphWeight int `json:"total_graph_weight"`
}
func (m *InitialPartitionCreation) Type() string { return "InitialPartitionCreation" }

type InitialPartitionCreationComplete struct{
	Sender actor.PID `json:"sender"`
}
func (m *InitialPartitionCreationComplete) Type() string { return "InitialPartitionCreationComplete" }
