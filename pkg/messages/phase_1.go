package messages

import "github.com/distributed-louvain/pkg/actor"

// Phase 1: Local Optimization Messages

type StartPhase1 struct{}
func (m *StartPhase1) Type() string { return "StartPhase1" }

type DegreeRequest struct {
	NodeID  int
	ReplyTo actor.PID
}
func (m *DegreeRequest) Type() string { return "DegreeRequest" }

type DegreeResponse struct {
	NodeID int
	Degree int
}
func (m *DegreeResponse) Type() string { return "DegreeResponse" }

type CommunityRequest struct {
	NodeID  int
	ReplyTo actor.PID
}
func (m *CommunityRequest) Type() string { return "CommunityRequest" }

type CommunityResponse struct {
	NodeID    int
	Community int
}
func (m *CommunityResponse) Type() string { return "CommunityResponse" }

type NodeTransition struct {
	NodeID           int
	OldCommunity     int
	NewCommunity     int
	ModularityDelta  float64
	Sender           actor.PID
}
func (m *NodeTransition) Type() string { return "NodeTransition" }

type Phase1Complete struct {
	PartitionID    int
	ModularityGain float64
	Sender         actor.PID
}
func (m *Phase1Complete) Type() string { return "Phase1Complete" }
