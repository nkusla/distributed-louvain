package actor

import (
	"context"
	"sync"
)

type Actor interface {
	PID() PID
	Receive(ctx context.Context, msg Message)
	Start(ctx context.Context)
	Stop()
	GetMailbox() *Mailbox
}

type BaseActor struct {
	pid     PID
	Mailbox *Mailbox
	System  *ActorSystem
	Ctx     context.Context
	Cancel  context.CancelFunc
	Wg      sync.WaitGroup
}

func NewBaseActor(pid PID, system *ActorSystem, mailboxSize int) *BaseActor {
	return &BaseActor{
		pid:     pid,
		Mailbox: NewMailbox(mailboxSize),
		System:  system,
	}
}

func (a *BaseActor) PID() PID {
	return a.pid
}

func (a *BaseActor) Send(to PID, msg Message) error {
	return a.System.Send(to, msg)
}

func (a *BaseActor) Stop() {
	if a.Cancel != nil {
		a.Cancel()
	}
	a.Mailbox.Close()
	a.Wg.Wait()
}

func (a *BaseActor) GetMailbox() *Mailbox {
	return a.Mailbox
}

func (a *BaseActor) Start(ctx context.Context) {
	// This should be overridden by concrete actors
}

func (a *BaseActor) Receive(ctx context.Context, msg Message) {
	// This should be overridden by concrete actors
}
