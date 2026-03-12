// github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events/kafka/dispatcher.go
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"
	domain_event "github.com/alphaxad9/my-go-backend/post_service/src/posts/domain/events"
	events "github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events"

	"github.com/google/uuid"
)

// Dispatcher routes Kafka events to local in-memory event bus
type Dispatcher struct {
	eventBus events.EventBus
}

func NewDispatcher(eventBus events.EventBus) *Dispatcher {
	return &Dispatcher{eventBus: eventBus}
}

// RegisterHandlers registers all supported post event types
func (d *Dispatcher) RegisterHandlers(consumer *KafkaConsumer) {
	consumer.RegisterHandler(domain_event.PostEventTypeCreated, d.handlePostCreated)
	consumer.RegisterHandler(domain_event.PostEventTypeUpdated, d.handlePostUpdated)
	consumer.RegisterHandler(domain_event.PostEventTypeDeleted, d.handlePostDeleted)
	consumer.RegisterHandler(domain_event.PostEventTypeVisibilityToggled, d.handlePostVisibilityToggled)
	consumer.RegisterHandler(domain_event.PostEventTypeLiked, d.handlePostLiked)
	consumer.RegisterHandler(domain_event.PostEventTypeUnliked, d.handlePostUnliked)
	consumer.RegisterHandler(domain_event.PostEventTypeCommented, d.handlePostCommented)
	consumer.RegisterHandler(domain_event.PostEventTypeCommentRemoved, d.handlePostCommentRemoved)
}

// --- Handlers ---

func (d *Dispatcher) handlePostCreated(ctx context.Context, eventType string, payload json.RawMessage) error {
	var fullEvent struct {
		ID          string                 `json:"id"`
		AggregateID string                 `json:"aggregate_id"`
		Timestamp   string                 `json:"timestamp"`
		Payload     map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(payload, &fullEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Parse required fields
	title := fullEvent.Payload["title"].(string)
	content := fullEvent.Payload["content"].(string)
	isPublic := fullEvent.Payload["is_public"].(bool)
	authorID := uuid.MustParse(fullEvent.Payload["author_id"].(string))
	communityID := uuid.MustParse(fullEvent.Payload["community_id"].(string))

	// Parse timestamp
	ts, err := time.Parse(time.RFC3339Nano, fullEvent.Timestamp)
	if err != nil {
		ts = time.Now().UTC()
	}

	// Reconstruct aggregate
	agg := domain.ReconstructPostAggregate(
		uuid.MustParse(fullEvent.AggregateID),
		authorID,
		communityID,
		title,
		content,
		isPublic,
		0, // likesCount
		0, // commentCount
		ts,
		ts,
	)

	event := domain_event.NewPostCreatedEvent(agg)
	return d.eventBus.Publish(ctx, event)
}

func (d *Dispatcher) handlePostUpdated(ctx context.Context, eventType string, payload json.RawMessage) error {
	var fullEvent struct {
		ID          string                 `json:"id"`
		AggregateID string                 `json:"aggregate_id"`
		Timestamp   string                 `json:"timestamp"`
		Payload     map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(payload, &fullEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	title := fullEvent.Payload["title"].(string)
	content := fullEvent.Payload["content"].(string)

	ts, err := time.Parse(time.RFC3339Nano, fullEvent.Timestamp)
	if err != nil {
		ts = time.Now().UTC()
	}

	// We only need ID for event creation
	agg := domain.ReconstructPostAggregate(
		uuid.MustParse(fullEvent.AggregateID),
		uuid.Nil, // placeholder
		uuid.Nil, // placeholder
		title,
		content,
		false, // placeholder
		0, 0, ts, ts,
	)

	event := domain_event.NewPostUpdatedEvent(agg)
	return d.eventBus.Publish(ctx, event)
}

func (d *Dispatcher) handlePostDeleted(ctx context.Context, eventType string, payload json.RawMessage) error {
	var fullEvent struct {
		ID          string                 `json:"id"`
		AggregateID string                 `json:"aggregate_id"`
		Timestamp   string                 `json:"timestamp"`
		Payload     map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(payload, &fullEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	title := fullEvent.Payload["title"].(string)
	authorID := uuid.MustParse(fullEvent.Payload["author_id"].(string))
	communityID := uuid.MustParse(fullEvent.Payload["community_id"].(string))

	ts, err := time.Parse(time.RFC3339Nano, fullEvent.Timestamp)
	if err != nil {
		ts = time.Now().UTC()
	}

	agg := domain.ReconstructPostAggregate(
		uuid.MustParse(fullEvent.AggregateID),
		authorID,
		communityID,
		title,
		"", // content not needed
		false,
		0, 0, ts, ts,
	)

	event := domain_event.NewPostDeletedEvent(agg)
	return d.eventBus.Publish(ctx, event)
}

func (d *Dispatcher) handlePostVisibilityToggled(ctx context.Context, eventType string, payload json.RawMessage) error {
	var fullEvent struct {
		ID          string                 `json:"id"`
		AggregateID string                 `json:"aggregate_id"`
		Timestamp   string                 `json:"timestamp"`
		Payload     map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(payload, &fullEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	isPublic := fullEvent.Payload["is_public"].(bool)

	ts, err := time.Parse(time.RFC3339Nano, fullEvent.Timestamp)
	if err != nil {
		ts = time.Now().UTC()
	}

	agg := domain.ReconstructPostAggregate(
		uuid.MustParse(fullEvent.AggregateID),
		uuid.Nil, uuid.Nil, "", "", isPublic,
		0, 0, ts, ts,
	)

	event := domain_event.NewPostVisibilityToggledEvent(agg)
	return d.eventBus.Publish(ctx, event)
}

func (d *Dispatcher) handlePostLiked(ctx context.Context, eventType string, payload json.RawMessage) error {
	var fullEvent struct {
		ID          string                 `json:"id"`
		AggregateID string                 `json:"aggregate_id"`
		Timestamp   string                 `json:"timestamp"`
		Payload     map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(payload, &fullEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	likesCount := int(fullEvent.Payload["likes_count"].(float64))

	ts, err := time.Parse(time.RFC3339Nano, fullEvent.Timestamp)
	if err != nil {
		ts = time.Now().UTC()
	}

	agg := domain.ReconstructPostAggregate(
		uuid.MustParse(fullEvent.AggregateID),
		uuid.Nil, uuid.Nil, "", "", false,
		likesCount, 0, ts, ts,
	)

	event := domain_event.NewPostLikedEvent(agg)
	return d.eventBus.Publish(ctx, event)
}

func (d *Dispatcher) handlePostUnliked(ctx context.Context, eventType string, payload json.RawMessage) error {
	var fullEvent struct {
		ID          string                 `json:"id"`
		AggregateID string                 `json:"aggregate_id"`
		Timestamp   string                 `json:"timestamp"`
		Payload     map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(payload, &fullEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	likesCount := int(fullEvent.Payload["likes_count"].(float64))

	ts, err := time.Parse(time.RFC3339Nano, fullEvent.Timestamp)
	if err != nil {
		ts = time.Now().UTC()
	}

	agg := domain.ReconstructPostAggregate(
		uuid.MustParse(fullEvent.AggregateID),
		uuid.Nil, uuid.Nil, "", "", false,
		likesCount, 0, ts, ts,
	)

	event := domain_event.NewPostUnlikedEvent(agg)
	return d.eventBus.Publish(ctx, event)
}

func (d *Dispatcher) handlePostCommented(ctx context.Context, eventType string, payload json.RawMessage) error {
	var fullEvent struct {
		ID          string                 `json:"id"`
		AggregateID string                 `json:"aggregate_id"`
		Timestamp   string                 `json:"timestamp"`
		Payload     map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(payload, &fullEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	commentCount := int(fullEvent.Payload["comment_count"].(float64))

	ts, err := time.Parse(time.RFC3339Nano, fullEvent.Timestamp)
	if err != nil {
		ts = time.Now().UTC()
	}

	agg := domain.ReconstructPostAggregate(
		uuid.MustParse(fullEvent.AggregateID),
		uuid.Nil, uuid.Nil, "", "", false,
		0, commentCount, ts, ts,
	)

	event := domain_event.NewPostCommentedEvent(agg)
	return d.eventBus.Publish(ctx, event)
}

func (d *Dispatcher) handlePostCommentRemoved(ctx context.Context, eventType string, payload json.RawMessage) error {
	var fullEvent struct {
		ID          string                 `json:"id"`
		AggregateID string                 `json:"aggregate_id"`
		Timestamp   string                 `json:"timestamp"`
		Payload     map[string]interface{} `json:"payload"`
	}

	if err := json.Unmarshal(payload, &fullEvent); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	commentCount := int(fullEvent.Payload["comment_count"].(float64))

	ts, err := time.Parse(time.RFC3339Nano, fullEvent.Timestamp)
	if err != nil {
		ts = time.Now().UTC()
	}

	agg := domain.ReconstructPostAggregate(
		uuid.MustParse(fullEvent.AggregateID),
		uuid.Nil, uuid.Nil, "", "", false,
		0, commentCount, ts, ts,
	)

	event := domain_event.NewPostCommentRemovedEvent(agg)
	return d.eventBus.Publish(ctx, event)
}
