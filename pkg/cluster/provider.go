package cluster

import (
	"context"
	"sync"

	"github.com/distributed-louvain/pkg/actor"
)

type MachineInfo struct {
	ID      string
	Address string
}

type SimpleProvider struct {
	machineID    string
	machines     map[string]*MachineInfo
	aggregators  []actor.PID       // Shared list of aggregator actors
	partitions   []actor.PID       // Shared list of partition actors
	mu           sync.RWMutex
}

func NewSimpleProvider(machineID string) *SimpleProvider {
	return &SimpleProvider{
		machineID:   machineID,
		machines:    make(map[string]*MachineInfo),
		aggregators: make([]actor.PID, 0),
		partitions:  make([]actor.PID, 0),
	}
}

func (p *SimpleProvider) Start(ctx context.Context) error {
	p.machines[p.machineID] = &MachineInfo{
		ID:      p.machineID,
		Address: "localhost:8080",
	}
	return nil
}

func (p *SimpleProvider) RegisterActor(pid actor.PID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.actorMap[pid.ActorID] = pid.MachineID

	return nil
}

func (p *SimpleProvider) RegisterAggregator(pid actor.PID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, existing := range p.aggregators {
		if existing.String() == pid.String() {
			return nil // Already registered
		}
	}

	p.aggregators = append(p.aggregators, pid)

	return nil
}

func (p *SimpleProvider) RegisterPartition(pid actor.PID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, existing := range p.partitions {
		if existing.String() == pid.String() {
			return nil // Already registered
		}
	}

	p.partitions = append(p.partitions, pid)

	return nil
}

func (p *SimpleProvider) GetAggregators() []actor.PID {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]actor.PID, len(p.aggregators))
	copy(result, p.aggregators)
	return result
}

func (p *SimpleProvider) GetPartitions() []actor.PID {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]actor.PID, len(p.partitions))
	copy(result, p.partitions)
	return result
}

func (p *SimpleProvider) RegisterMachine(machineID, address string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.machines[machineID]; !exists {
		p.machines[machineID] = &MachineInfo{
			ID:      machineID,
			Address: address,
		}
	}
}

func (p *SimpleProvider) Stop() error {
	return nil
}
