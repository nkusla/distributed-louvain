package actor

import "fmt"

type PID struct {
	MachineID  string
	ActorID string
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

func ParsePID(s string) (PID, error) {
	var machineId, actorID string
	_, err := fmt.Sscanf(s, "%s/%s", &machineId, &actorID)
	if err != nil {
		return PID{}, fmt.Errorf("invalid PID format: %s", s)
	}
	return PID{MachineID: machineId, ActorID: actorID}, nil
}
