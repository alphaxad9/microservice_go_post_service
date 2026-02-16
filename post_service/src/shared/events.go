//my-go-backend/post_service/src/shared/events.go

package shared

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent is the shared interface for all domain events across bounded contexts.
// It ensures every event carries essential metadata for routing, tracing, and replay.
type DomainEvent interface {
	// EventType returns a stable, namespaced identifier (e.g., "package.shipped").
	EventType() string

	// AggregateID returns the unique ID of the aggregate that emitted this event.
	AggregateID() string

	// AggregateType returns the type of the emitting aggregate (e.g., "Package", "Payment").
	AggregateType() string

	// ID returns the unique event ID (UUID).
	ID() string

	// Timestamp returns when the event occurred (UTC).
	Timestamp() time.Time

	// ToMap returns a JSON-serializable representation of the event payload + metadata.
	// Implementations should embed BaseEvent or call BaseEvent.ToMap().
	ToMap() map[string]interface{}
}

// BaseEvent provides common fields and methods for all concrete domain events.
// Embed this struct anonymously in your event structs.
type BaseEvent struct {
	id            string
	aggregateID   string
	aggregateType string
	eventType     string
	timestamp     time.Time
}

// NewBaseEvent constructs a BaseEvent with standard defaults.
func NewBaseEvent(aggregateID, aggregateType, eventType string) BaseEvent {
	return BaseEvent{
		id:            uuid.New().String(),
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		eventType:     eventType,
		timestamp:     time.Now().UTC(),
	}
}

// ID implements DomainEvent.ID().
func (e BaseEvent) ID() string { return e.id }

// AggregateID implements DomainEvent.AggregateID().
func (e BaseEvent) AggregateID() string { return e.aggregateID }

// AggregateType implements DomainEvent.AggregateType().
func (e BaseEvent) AggregateType() string { return e.aggregateType }

// EventType implements DomainEvent.EventType().
func (e BaseEvent) EventType() string { return e.eventType }

// Timestamp implements DomainEvent.Timestamp().
func (e BaseEvent) Timestamp() time.Time { return e.timestamp }

// ToMap returns a serializable map including metadata.
// Concrete events should override this by merging their own payload.
func (e BaseEvent) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":             e.id,
		"event_type":     e.eventType,
		"aggregate_id":   e.aggregateID,
		"aggregate_type": e.aggregateType,
		"timestamp":      e.timestamp.Format(time.RFC3339Nano),
		"payload":        map[string]interface{}{}, // override in concrete event
	}
}

// NewBaseEventWithID creates a BaseEvent with explicit ID and timestamp.
// Use this when reconstructing events from external sources (e.g., Kafka).
func NewBaseEventWithID(id, aggregateID, aggregateType, eventType string, timestamp time.Time) BaseEvent {
	return BaseEvent{
		id:            id,
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		eventType:     eventType,
		timestamp:     timestamp,
	}
}
