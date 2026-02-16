// Package models contains database-level structs that map directly to tables.
// These are used only by infrastructure repositories—not exposed to domain logic.
// my-go-backend/post_service/src/posts/infra/models/outbox_model.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// OutboxModel represents a row in the 'event_outbox' table.
// Used exclusively for data persistence and retrieval via the infrastructure layer.
type OutboxModel struct {
	ID              uuid.UUID `db:"id"`
	EventType       string    `db:"event_type"`
	EventPayload    []byte    `db:"event_payload"` // stored as JSONB or TEXT
	AggregateID     uuid.UUID `db:"aggregate_id"`
	AggregateType   string    `db:"aggregate_type"`
	AggregateVersion uint32   `db:"aggregate_version"`
	TraceID         *uuid.UUID `db:"trace_id"`
	Metadata        []byte    `db:"metadata"` // stored as JSONB or TEXT
	CreatedAt       time.Time `db:"created_at"`
	PublishedAt     *time.Time `db:"published_at"`
	ProcessedAt     *time.Time `db:"processed_at"`
	RetryCount      uint16    `db:"retry_count"`
	ErrorMessage    *string   `db:"error_message"`
}