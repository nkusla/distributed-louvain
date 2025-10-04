package actor

import (
	"sync"
)

type Message interface {
	Type() string
}

type Mailbox struct {
	messages chan Message
	mutex       sync.RWMutex
	closed   bool
}

func NewMailbox(size int) *Mailbox {
	return &Mailbox{
		messages: make(chan Message, size),
	}
}

func (m *Mailbox) Send(msg Message) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.closed {
		return ErrMailboxClosed
	}

	select {
	case m.messages <- msg:
		return nil
	default:
		return ErrMailboxFull
	}
}

func (m *Mailbox) Receive() <-chan Message {
	return m.messages
}

func (m *Mailbox) Close() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.closed {
		m.closed = true
		close(m.messages)
	}
}
