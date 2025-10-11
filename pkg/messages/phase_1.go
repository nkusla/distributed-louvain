package messages

import (
	"github.com/distributed-louvain/pkg/actor"
	"github.com/distributed-louvain/pkg/crdt"
)

// Phase 1: Local Optimization Messages

type StartPhase1 struct{}
func (m *StartPhase1) Type() string { return "StartPhase1" }

type DegreeRequest struct {
	NodeID			int
	NeighborID	int
	Sender	actor.PID
}
func (m *DegreeRequest) Type() string { return "DegreeRequest" }

type DegreeResponse struct {
	NodeID			int
	NeighborID	int
	Degree			int
}
func (m *DegreeResponse) Type() string { return "DegreeResponse" }

type LocalOptimizationComplete struct {
	NodeSet	*crdt.NodeSet
	Sender	actor.PID
}
func (m *LocalOptimizationComplete) Type() string { return "LocalOptimizationComplete" }

type Phase1Complete struct {
}
func (m *Phase1Complete) Type() string { return "Phase1Complete" }
