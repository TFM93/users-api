package domain

import (
	"context"

	"github.com/google/uuid"
)

type (
	// NotificationService is an interface for sending events as notifications.
	NotificationService interface {
		// Publish sends the given event to the notification system
		// In case of failure it returns domain.ErrNotificationNotSent or domain.ErrFailedToProcessData
		Publish(context.Context, *Event) error
	}

	// Event represents an Event in the domain model
	Event struct {
		ID      uuid.UUID
		Type    string
		Payload []byte
	}
)
