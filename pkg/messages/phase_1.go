package messages

import (
	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/crdt"
)

// Phase 1: Local Optimization Messages

type StartPhase1 struct{}
func (m *StartPhase1) Type() string { return "StartPhase1" }

type DegreeRequest struct {
	NodeID			int `json:"node_id"`
	NeighborID	int `json:"neighbor_id"`
	Sender	actor.PID `json:"sender"`
}
func (m *DegreeRequest) Type() string { return "DegreeRequest" }

type DegreeResponse struct {
	NodeID			int `json:"node_id"`
	NeighborID	int `json:"neighbor_id"`
	Degree			int `json:"degree"`
}
func (m *DegreeResponse) Type() string { return "DegreeResponse" }

type LocalOptimizationComplete struct {
	NodeSet	*crdt.NodeSet `json:"node_set"`
	Sender	actor.PID `json:"sender"`
}
func (m *LocalOptimizationComplete) Type() string { return "LocalOptimizationComplete" }
