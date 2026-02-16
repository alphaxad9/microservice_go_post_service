// Package outbox defines the domain model and repository contract for the transactional outbox pattern.
package outbox

import (
	"context"

	"github.com/google/uuid"
)

// OutboxRepository defines the domain-level interface for persisting and managing outbox events.
// It is used by application services across all domains (e.g., user, post, payment) to reliably
// stage domain events for asynchronous publishing.
//
// Implementations must be transactionally consistent with the originating domain operation
// (i.e., saving an event must happen in the same DB transaction as the state change).
type OutboxRepository interface {
	// Save persists an outbox event atomically with the domain state change.
	// Called within the same database transaction as the aggregate update.
	Save(ctx context.Context, event *OutboxEvent) error

	// GetUnpublishedEvents fetches up to 'limit' unpublished events for processing.
	// Must be safe for concurrent consumers (e.g., using SELECT FOR UPDATE SKIP LOCKED).
	// Returns events ordered by CreatedAt (oldest first).
	GetUnpublishedEvents(ctx context.Context, limit int) ([]*OutboxEvent, error)

	// MarkAsPublished marks an event as successfully delivered.
	// Sets PublishedAt and ProcessedAt to now (UTC).
	MarkAsPublished(ctx context.Context, outboxID uuid.UUID) error

	// MarkAsFailed records a publishing failure for observability and retry logic.
	// Increments retry count and stores the error message.
	// If max retries are exceeded, the caller should handle dead-letter logic.
	MarkAsFailed(ctx context.Context, outboxID uuid.UUID, errorMsg string) error
}
