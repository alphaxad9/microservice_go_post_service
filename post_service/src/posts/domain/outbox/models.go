package outbox

import (
	"encoding/json"
	exceptons "github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"
	"time"

	"github.com/google/uuid"
)

// OutboxEvent represents a domain event staged for reliable delivery (e.g., to a message broker).
// It is immutable by design—once created, its fields should not change.
type OutboxEvent struct {
	ID            uuid.UUID       `json:"id"`
	EventType     string          `json:"event_type"`
	EventPayload  json.RawMessage `json:"event_payload"`
	AggregateID   uuid.UUID       `json:"aggregate_id"`
	AggregateType string          `json:"aggregate_type"`
	TraceID       *uuid.UUID      `json:"trace_id,omitempty"`
	Metadata      json.RawMessage `json:"metadata"`
	CreatedAt     time.Time       `json:"created_at"`
	RetryCount    int             `json:"retry_count"`
}

// NewOutboxEvent creates a new outbox event with safe serialization of payload and metadata.
// Payload and metadata must be JSON-serializable.
func NewOutboxEvent(
	eventType string,
	payload any,
	aggregateID uuid.UUID,
	aggregateType string,
	traceID *uuid.UUID,
	metadata map[string]any,
) (*OutboxEvent, error) {
	if eventType == "" {
		return nil, exceptons.NewValidationFailed(map[string]string{
			"event_type": "Event type is required",
		})
	}
	if aggregateID == uuid.Nil {
		return nil, exceptons.NewValidationFailed(map[string]string{
			"aggregate_id": "Valid aggregate ID is required",
		})
	}
	if aggregateType == "" {
		return nil, exceptons.NewValidationFailed(map[string]string{
			"aggregate_type": "Aggregate type is required",
		})
	}

	// Serialize payload safely
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, exceptons.NewDomainError("failed to serialize event payload", err)
	}

	// Ensure metadata is always a map; default to empty if nil
	if metadata == nil {
		metadata = make(map[string]any)
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, exceptons.NewDomainError("failed to serialize metadata", err)
	}

	now := time.Now().UTC()
	id := uuid.New()

	return &OutboxEvent{
		ID:            id,
		EventType:     eventType,
		EventPayload:  payloadBytes,
		AggregateID:   aggregateID,
		AggregateType: aggregateType,
		TraceID:       traceID,
		Metadata:      metadataBytes,
		CreatedAt:     now,
		RetryCount:    0,
	}, nil
}

// IncrementRetry increments the retry count and updates internal state.
// Note: Although OutboxEvent is conceptually immutable, this method supports retry mechanics
// during processing. Use cautiously—ideally, retries are handled by replacing the record.
func (e *OutboxEvent) IncrementRetry() {
	e.RetryCount++
	// CreatedAt is intentionally NOT updated—retry does not change event origin time
}

// GetEventPayload returns the deserialized event payload.
func (e *OutboxEvent) GetEventPayload(target any) error {
	return json.Unmarshal(e.EventPayload, target)
}

// GetMetadata returns the deserialized metadata.
func (e *OutboxEvent) GetMetadata(target any) error {
	return json.Unmarshal(e.Metadata, target)
}
