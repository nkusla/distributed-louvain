package messages

import (
	"github.com/distributed-louvain/pkg/graph"
	"github.com/distributed-louvain/pkg/actor"
)

type InitialPartitionCreation struct{
	Edges []graph.Edge
	TotalGraphWeight int
}
func (m *InitialPartitionCreation) Type() string { return "InitialPartitionCreation" }

type InitialPartitionCreationComplete struct{
	Sender actor.PID
}
func (m *InitialPartitionCreationComplete) Type() string { return "InitialPartitionCreationComplete" }
