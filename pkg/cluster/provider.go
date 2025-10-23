package cluster

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/distributed-louvain/pkg/actor"
)

type SimpleProvider struct {
	machineID    string
	machines     map[string]string
	transport    *Transport
	coordinator  actor.PID
	actorMap     map[actor.ActorType][]actor.PID
	mu           sync.RWMutex
}

func NewSimpleProvider(machineID string, useTransportLayer bool) *SimpleProvider {
	var p = &SimpleProvider{
		machineID:   machineID,
		machines:    make(map[string]string),
		coordinator: actor.PID{},
		actorMap:    make(map[actor.ActorType][]actor.PID),
	}

	if useTransportLayer {
		p.transport = NewTransport(machineID, 8080)
	}

	return p
}

func (p *SimpleProvider) MachineID() string {
	return p.machineID
}

func (p *SimpleProvider) Start(ctx context.Context) error {
	p.machines[p.machineID] = "localhost:8080"

	if p.transport != nil {
		p.transport.Start(ctx)
	}

	return nil
}

func (p *SimpleProvider) SetActorSystem(system *actor.ActorSystem) {
	if p.transport != nil {
		p.transport.SetActorSystem(system)
	}
}

func (p *SimpleProvider) SetCoordinator(coordinator actor.PID) {
	p.coordinator = coordinator
}

func (p *SimpleProvider) RegisterActor(actorType actor.ActorType, pid actor.PID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.actorMap[actorType] = append(p.actorMap[actorType], pid)

  actors := p.actorMap[actorType]

	// Sort by MachineID first, then by ActorID for deterministic ordering
	sort.Slice(actors, func(i, j int) bool {
		if actors[i].MachineID != actors[j].MachineID {
			return actors[i].MachineID < actors[j].MachineID
		}
		return actors[i].ActorID < actors[j].ActorID
	})

	return nil
}

func (p *SimpleProvider) GetCoordinator() actor.PID {
	return p.coordinator
}

func (p *SimpleProvider) GetActors(actorType actor.ActorType) []actor.PID {
	p.mu.RLock()
	defer p.mu.RUnlock()

	actors := make([]actor.PID, len(p.actorMap[actorType]))
	copy(actors, p.actorMap[actorType])
	return actors
}

func (p *SimpleProvider) RegisterMachine(machineID, address string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.machines[machineID]; !exists {
		p.machines[machineID] = address
	}
}

func (p *SimpleProvider) Send(to actor.PID, msg actor.Message) error {
	if p.transport != nil {
		return p.transport.Send(to, p.machines[to.MachineID], msg)
	}
	return fmt.Errorf("transport layer not enabled")
}

func (p *SimpleProvider) Stop() error {
	if p.transport != nil {
		p.transport.Stop()
	}
	return nil
}
