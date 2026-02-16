// Package repositories implements PostgreSQL-backed domain repositories.
// my-go-backend/post_service/src/posts/infra/repositories/outbox/outbox_repo.go
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"my-go-backend/post_service/src/posts/domain/outbox"
	"my-go-backend/post_service/src/posts/infra/models"
	"my-go-backend/post_service/src/posts/ports"

	"github.com/google/uuid"
)

// dbExecutor defines the minimal interface needed for database operations.
// This allows the repository to work with both *sql.DB and *sql.Tx.
type dbExecutor interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}

// OutboxRepositoryImpl implements the outbox.OutboxRepository interface using PostgreSQL.
type OutboxRepositoryImpl struct {
	db dbExecutor
}

// NewOutboxRepository returns a new PostgreSQL-backed outbox repository using a database connection.
func NewOutboxRepository(db *sql.DB) *OutboxRepositoryImpl {
	return &OutboxRepositoryImpl{db: db}
}

// NewOutboxRepositoryTx returns a new PostgreSQL-backed outbox repository using a transaction.
func NewOutboxRepositoryTx(tx *sql.Tx) *OutboxRepositoryImpl {
	return &OutboxRepositoryImpl{db: tx}
}

// Save persists an outbox event in the same transaction as the domain state change.
// Note: For true transactional consistency, this should be called within a shared transaction.
func (r *OutboxRepositoryImpl) Save(ctx context.Context, event *outbox.OutboxEvent) error {
	// Convert domain event to DB model
	// We use aggregate_version = 0 since it's not exposed in domain yet (but table requires it)
	model := models.OutboxModel{
		ID:               event.ID,
		EventType:        event.EventType,
		EventPayload:     []byte(event.EventPayload),
		AggregateID:      event.AggregateID,
		AggregateType:    event.AggregateType,
		AggregateVersion: 0, // or derive from aggregate if versioning is added later
		TraceID:          event.TraceID,
		Metadata:         []byte(event.Metadata),
		CreatedAt:        event.CreatedAt,
		RetryCount:       uint16(event.RetryCount),
		// PublishedAt, ProcessedAt, ErrorMessage are NULL initially
	}

	_, err := r.db.ExecContext(ctx, ports.QueryInsertOutboxEvent,
		model.ID,
		model.EventType,
		model.EventPayload,
		model.AggregateID,
		model.AggregateType,
		model.AggregateVersion,
		model.TraceID,
		model.Metadata,
		model.CreatedAt,
		model.RetryCount,
	)
	if err != nil {
		return outbox.NewOutboxSaveError(event.EventType, event.AggregateID, err.Error())
	}
	return nil
}

// GetUnpublishedEvents retrieves unpublished events, locking them for processing.
// IMPORTANT: This method should be called within a transaction to respect the FOR UPDATE lock.
// Use NewOutboxRepositoryTx() when polling for events.
func (r *OutboxRepositoryImpl) GetUnpublishedEvents(ctx context.Context, limit int) ([]*outbox.OutboxEvent, error) {
	rows, err := r.db.QueryContext(ctx, ports.QuerySelectUnpublishedOutboxEvents, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unpublished outbox events: %w", err)
	}
	defer rows.Close()

	var events []*outbox.OutboxEvent
	for rows.Next() {
		var model models.OutboxModel
		err := rows.Scan(
			&model.ID,
			&model.EventType,
			&model.EventPayload,
			&model.AggregateID,
			&model.AggregateType,
			&model.AggregateVersion,
			&model.TraceID,
			&model.Metadata,
			&model.CreatedAt,
			&model.RetryCount,
			&model.PublishedAt,
			&model.ProcessedAt,
			&model.ErrorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan outbox row: %w", err)
		}

		event := &outbox.OutboxEvent{
			ID:            model.ID,
			EventType:     model.EventType,
			EventPayload:  json.RawMessage(model.EventPayload),
			AggregateID:   model.AggregateID,
			AggregateType: model.AggregateType,
			TraceID:       model.TraceID,
			Metadata:      json.RawMessage(model.Metadata),
			CreatedAt:     model.CreatedAt,
			RetryCount:    int(model.RetryCount),
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return events, nil
}

// MarkAsPublished marks an event as successfully delivered.
func (r *OutboxRepositoryImpl) MarkAsPublished(ctx context.Context, outboxID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, ports.QueryMarkOutboxAsPublished, outboxID)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event as published: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return outbox.NewOutboxNotFoundError(outboxID)
	}
	return nil
}

// MarkAsFailed records a publishing failure and increments retry count.
func (r *OutboxRepositoryImpl) MarkAsFailed(ctx context.Context, outboxID uuid.UUID, errorMsg string) error {
	result, err := r.db.ExecContext(ctx, ports.QueryMarkOutboxAsFailed, errorMsg, outboxID)
	if err != nil {
		return fmt.Errorf("failed to mark outbox event as failed: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return outbox.NewOutboxNotFoundError(outboxID)
	}
	return nil
}
