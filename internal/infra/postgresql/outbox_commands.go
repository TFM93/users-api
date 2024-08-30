package postgresql

import (
	"context"
	"fmt"
	"users/internal/domain"
	log "users/pkg/logger"
	"users/pkg/postgresql"
)

type outboxCommandsRepo struct {
	pg postgresql.Interface
	l  log.Interface
}

// NewOutboxCommandsRepo creates a new instance of outboxCommandsRepo that satisfies the domain.OutboxRepoCommands interface
func NewOutboxCommandsRepo(pg postgresql.Interface, logger log.Interface) domain.OutboxRepoCommands {
	ur := &outboxCommandsRepo{pg: pg, l: logger}
	return ur
}

func (r outboxCommandsRepo) db(ctx context.Context) postgresql.DBProvider {
	tx, ok := ctx.Value(domain.TxKey).(postgresql.Tx)
	if ok {
		return tx
	}
	return r.pg.GetPool()
}

// GetUnprocessed fetches the unprocessed events from the database.
// The amount of entries in the response is capped by the limit param
func (r outboxCommandsRepo) GetUnprocessed(ctx context.Context, limit int32) ([]*domain.Event, error) {
	query := `SELECT id, event_type, payload 
		FROM outbox 
		WHERE processed_at IS NULL 
		FOR UPDATE SKIP LOCKED 
		LIMIT $1`
	rows, err := r.db(ctx).Query(ctx, query, limit)
	if err != nil {
		r.l.Debug(fmt.Errorf("failed to fetch unprocessed outbox events: %w", err))
		return nil, domain.ErrInternal
	}
	defer rows.Close()

	var events []*domain.Event
	// NOTE: as per version 5 of pgx this can be done relying on generics:
	// https://donchev.is/post/working-with-postgresql-in-go-using-pgx/
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(&event.ID, &event.Type, &event.Payload); err != nil {
			r.l.Error(fmt.Errorf("failed to scan row: %w", err))
			return nil, domain.ErrFailedToProcessData
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		r.l.Error(fmt.Errorf("row iteration error: %w", err))
		return nil, domain.ErrFailedToProcessData
	}
	return events, nil
}

// MarkAsProcessed updates outbox table.
func (r outboxCommandsRepo) MarkAsProcessed(ctx context.Context, id string) error {
	query := `UPDATE outbox SET processed_at=NOW() WHERE id=$1;`
	_, err := r.db(ctx).Exec(ctx, query, id)
	if err != nil {
		r.l.Error(fmt.Errorf("failed to update outbox: %w", err))
		return domain.ErrInternal
	}
	return err
}

// AddEvent creates a new event in the database.
// If an internal error occurs, it logs the error and returns domain.ErrInternal
func (r outboxCommandsRepo) AddEvent(ctx context.Context, event *domain.Event) (id string, err error) {
	query := `INSERT INTO outbox (event_type, payload) VALUES ($1, $2) RETURNING id`
	err = r.db(ctx).QueryRow(ctx, query, event.Type, event.Payload).Scan(&id)
	if err != nil {
		r.l.Error(fmt.Errorf("failed to create event: %w", err))
		return id, domain.ErrInternal
	}
	return id, nil
}
