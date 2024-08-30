package domain

import (
	"context"
)

type (
	// OutboxRepoCommands is an interface for persisting events
	OutboxRepoCommands interface {
		// GetUnprocessed fetches a list of events up to the provided limit
		// Parameters:
		//   limit: Maximum number of events to return
		// Returns a slice of event objects and an error if the operation fails
		GetUnprocessed(ctx context.Context, limit int32) ([]*Event, error)

		// MarkAsProcessed marks a persisted event as processed
		// If an internal error occurs, it logs the error and returns domain.ErrInternal
		MarkAsProcessed(ctx context.Context, id string) error

		// AddEvent creates a new event in the database.
		// Returns the created Event's ID and an error in case of failure
		// If an internal error occurs, it logs the error and returns domain.ErrInternal
		AddEvent(ctx context.Context, event *Event) (string, error)
	}
)
