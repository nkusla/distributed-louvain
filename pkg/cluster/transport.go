package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/distributed-louvain/pkg/actor"
)

type Transport struct {
	machineID string
	system    *actor.ActorSystem
}

func NewTransport(machineID string) *Transport {
	return &Transport{
		machineID: machineID,
	}
}

func (t *Transport) SetSystem(system *actor.ActorSystem) {
	t.system = system
}

func (t *Transport) Start(ctx context.Context) error {
	log.Printf("[Transport] Started for machine %s", t.machineID)
	return nil
}

func (t *Transport) Send(to actor.PID, msg actor.Message) error {
	// For now, this is a simple local transport
	// In production, this would serialize and send over network (gRPC)

	if to.MachineID == t.machineID {
		// Local delivery - should not happen here
		return fmt.Errorf("transport should only handle remote messages")
	}

	// Serialize message (in production, use protobuf)
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	log.Printf("[Transport] Would send %d bytes to %s (type: %s)", len(data), to, msg.Type())

	// TODO: Implement actual network send via gRPC
	// For now, just log the attempt

	return nil
}

func (t *Transport) Stop() error {
	log.Printf("[Transport] Stopped")
	return nil
}
