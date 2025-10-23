package actor

import (
	"fmt"
	"strings"
)

type PID struct {
	MachineID string `json:"machine_id"`
	ActorID   string `json:"actor_id"`
}

func NewPID(machineId string, actorID string) PID {
	return PID{
		MachineID:  machineId,
		ActorID: actorID,
	}
}

func (p PID) String() string {
	return fmt.Sprintf("%s/%s", p.MachineID, p.ActorID)
}

func (p PID) IsLocal(machineId string) bool {
	return p.MachineID == machineId
}

func (p PID) IsZero() bool {
	return p.MachineID == "" && p.ActorID == ""
}

func (p PID) Equal(other PID) bool {
	return p.MachineID == other.MachineID && p.ActorID == other.ActorID
}

func ParsePID(s string) (PID, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return PID{}, fmt.Errorf("invalid PID format: %s", s)
	}
	return PID{MachineID: parts[0], ActorID: parts[1]}, nil
}
