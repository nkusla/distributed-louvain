package actor
import (
	"context"
	"fmt"
	"sync"
)

type Provider interface {
	GetActors(actorType ActorType) []PID
	MachineID() string
	Send(to PID, msg Message) error
	Start(ctx context.Context) error
	Stop() error
}

type ActorSystem struct {
	actors    map[string]Actor
	mu        sync.RWMutex
	provider Provider
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewActorSystem(machineId string, provider Provider) *ActorSystem {
	ctx, cancel := context.WithCancel(context.Background())
	return &ActorSystem{
		actors:    make(map[string]Actor),
		provider: provider,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (s *ActorSystem) Start() error {
	if s.provider != nil {
		return s.provider.Start(s.ctx)
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
	if to.IsLocal(s.provider.MachineID()) {
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
	if s.provider == nil {
		return fmt.Errorf("no transport configured for remote delivery")
	}
	return s.provider.Send(to, msg)
}

func (s *ActorSystem) Broadcast(from PID, actorType ActorType, msg Message) {
	actors := s.GetActors(actorType)

	for _, actorPID := range actors {
		if actorPID.Equal(from) {
			continue
		}

		s.Send(actorPID, msg)
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

	if s.provider != nil {
		s.provider.Stop()
	}
}


func (s *ActorSystem) GetActors(actorType ActorType) []PID {
	if s.provider != nil {
		return s.provider.GetActors(actorType)
	}
	return []PID{}
}
