package actor

import "fmt"

type PID struct {
	NodeID  string
	ActorID string
}

func NewPID(nodeID string, actorID string) PID {
	return PID{
		NodeID:  nodeID,
		ActorID: actorID,
	}
}

func (p PID) String() string {
	return fmt.Sprintf("%s/%s", p.NodeID, p.ActorID)
}

func (p PID) IsLocal(nodeID string) bool {
	return p.NodeID == nodeID
}

func (p PID) IsZero() bool {
	return p.NodeID == "" && p.ActorID == ""
}

func ParsePID(s string) (PID, error) {
	var nodeID, actorID string
	_, err := fmt.Sscanf(s, "%s/%s", &nodeID, &actorID)
	if err != nil {
		return PID{}, fmt.Errorf("invalid PID format: %s", s)
	}
	return PID{NodeID: nodeID, ActorID: actorID}, nil
}
