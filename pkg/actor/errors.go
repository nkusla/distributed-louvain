package actor

import "errors"

var (
	ErrActorNotFound = errors.New("actor not found")

	ErrMailboxFull = errors.New("mailbox is full")

	ErrMailboxClosed = errors.New("mailbox is closed")

	ErrSystemShutdown = errors.New("actor system is shutting down")

	ErrRemoteDeliveryFailed = errors.New("remote message delivery failed")
)
