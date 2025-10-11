package actor
import (
	"context"
	"fmt"
	"sync"
)

type Transport interface {
	Send(to PID, msg Message) error
	Start(ctx context.Context) error
	Stop() error
}

type Provider interface {
	GetAggregators() []PID
	GetPartitions() []PID
	FindActor(actorID string) (PID, error)
	Start(ctx context.Context) error
	Stop() error
}

type ActorSystem struct {
	machineId    string
	actors    map[string]Actor
	mu        sync.RWMutex
	transport Transport
	provider Provider
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewActorSystem(machineId string, transport Transport, provider Provider) *ActorSystem {
	ctx, cancel := context.WithCancel(context.Background())
	return &ActorSystem{
		machineId:    machineId,
		actors:    make(map[string]Actor),
		transport: transport,
		provider: provider,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (s *ActorSystem) MachineID() string {
	return s.machineId
}

func (s *ActorSystem) Start() error {
	if s.transport != nil {
		return s.transport.Start(s.ctx)
	}
	return nil
}

func (s *ActorSystem) Register(actor Actor) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pid := actor.PID()
	if _, exists := s.actors[pid.ActorID]; exists {
		return fmt.Errorf("actor %s already registered", pid.ActorID)
	}

	s.actors[pid.ActorID] = actor
	return nil
}

func (s *ActorSystem) Unregister(actorID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.actors, actorID)
}

func (s *ActorSystem) Send(to PID, msg Message) error {
	if to.IsLocal(s.machineId) {
		return s.localDeliver(to, msg)
	}
	return s.remoteDeliver(to, msg)
}

func (s *ActorSystem) localDeliver(to PID, msg Message) error {
	s.mu.RLock()
	actor, exists := s.actors[to.ActorID]
	s.mu.RUnlock()

	if !exists {
		return ErrActorNotFound
	}

	mailbox := actor.GetMailbox()
	if mailbox != nil {
		return mailbox.Send(msg)
	}

	go actor.Receive(s.ctx, msg)
	return nil
}

func (s *ActorSystem) remoteDeliver(to PID, msg Message) error {
	if s.transport == nil {
		return fmt.Errorf("no transport configured for remote delivery")
	}
	return s.transport.Send(to, msg)
}

func (s *ActorSystem) GetActor(actorID string) (Actor, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	actor, exists := s.actors[actorID]
	return actor, exists
}

func (s *ActorSystem) Broadcast(msg Message) {
	s.mu.RLock()
	actors := make([]Actor, 0, len(s.actors))
	for _, actor := range s.actors {
		actors = append(actors, actor)
	}
	s.mu.RUnlock()

	for _, actor := range actors {
		go actor.Receive(s.ctx, msg)
	}
}

func (s *ActorSystem) Shutdown() {
	s.cancel()

	s.mu.RLock()
	actors := make([]Actor, 0, len(s.actors))
	for _, actor := range s.actors {
		actors = append(actors, actor)
	}
	s.mu.RUnlock()

	for _, actor := range actors {
		actor.Stop()
	}

	if s.transport != nil {
		s.transport.Stop()
	}
}

func (s *ActorSystem) GetAggregators() []PID {
	if s.provider != nil {
		return s.provider.GetAggregators()
	}
	return []PID{}
}

func (s *ActorSystem) GetPartitions() []PID {
	if s.provider != nil {
		return s.provider.GetPartitions()
	}
	return []PID{}
}

func (s *ActorSystem) FindActor(actorID string) (PID, error) {
	if s.provider != nil {
		return s.provider.FindActor(actorID)
	}
	return PID{}, fmt.Errorf("no cluster provider available")
}
