// github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events/eventbus.go
package eventbus

import (
	"context"
	"log/slog"
	"sync"

	"github.com/alphaxad9/my-go-backend/post_service/src/shared"
)

// EventHandler is a function that processes a domain event.
// It should be non-blocking and safe to run concurrently.
type EventHandler func(ctx context.Context, event shared.DomainEvent) error

// EventBus defines the interface for publishing and subscribing to domain events.
type EventBus interface {
	Publish(ctx context.Context, event shared.DomainEvent) error
	Subscribe(eventType string, handler EventHandler)
}

// InMemoryEventBus is a thread-safe, in-memory implementation of EventBus.
// It supports concurrent publishing and dynamic handler registration.
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewInMemoryEventBus creates and returns a new instance of InMemoryEventBus.
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish delivers the event to all handlers subscribed to its EventType().
// Handlers are executed concurrently via goroutines.
// Individual handler errors are logged but do not propagate or stop other handlers.
// The method itself always returns nil to ensure fire-and-forget semantics.
func (b *InMemoryEventBus) Publish(ctx context.Context, event shared.DomainEvent) error {
	eventType := event.EventType()

	b.mu.RLock()
	handlerList := b.handlers[eventType]
	// Copy slice to avoid holding lock during execution
	handlers := make([]EventHandler, len(handlerList))
	copy(handlers, handlerList)
	b.mu.RUnlock()

	if len(handlers) == 0 {
		slog.Debug("No handlers registered for event", "event_type", eventType, "event_id", event.ID())
		return nil
	}

	slog.Info("Publishing domain event", "event_type", eventType, "event_id", event.ID())

	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h EventHandler, e shared.DomainEvent) {
			defer wg.Done()
			if err := h(ctx, e); err != nil {
				slog.Error("Event handler failed",
					"error", err,
					"event_type", e.EventType(),
					"event_id", e.ID(),
					"aggregate_id", e.AggregateID(),
				)
			}
		}(handler, event)
	}

	wg.Wait()
	return nil
}

// Subscribe registers an EventHandler for a specific event type string.
// Event types should match the value returned by DomainEvent.EventType().
// Multiple handlers can subscribe to the same event type.
func (b *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) {
	if handler == nil {
		slog.Warn("Attempted to subscribe nil handler", "event_type", eventType)
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	slog.Debug("Subscribed handler to event", "event_type", eventType)
}
