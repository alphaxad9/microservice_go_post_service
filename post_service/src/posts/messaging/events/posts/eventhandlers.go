// my-go-backend/post_service/src/posts/messaging/events/posts/eventhandlers.go
package handlers

import (
	"context"
	"log/slog"

	domain_event "my-go-backend/post_service/src/posts/domain/events"
	eventbus "my-go-backend/post_service/src/posts/messaging/events"
	"my-go-backend/post_service/src/shared"
)

type PostEventHandler struct {
	eventBus eventbus.EventBus
	// Add dependencies here if needed (e.g., repo, client)
}

func NewPostEventHandler(eventBus eventbus.EventBus) *PostEventHandler {
	return &PostEventHandler{
		eventBus: eventBus,
	}
}

func (h *PostEventHandler) RegisterSubscriptions() {
	h.eventBus.Subscribe(domain_event.PostEventTypeCreated, h.handlePostCreated)
	h.eventBus.Subscribe(domain_event.PostEventTypeUpdated, h.handlePostUpdated)
	h.eventBus.Subscribe(domain_event.PostEventTypeDeleted, h.handlePostDeleted)
	h.eventBus.Subscribe(domain_event.PostEventTypeVisibilityToggled, h.handlePostVisibilityToggled)
	h.eventBus.Subscribe(domain_event.PostEventTypeLiked, h.handlePostLiked)
	h.eventBus.Subscribe(domain_event.PostEventTypeUnliked, h.handlePostUnliked)
	h.eventBus.Subscribe(domain_event.PostEventTypeCommented, h.handlePostCommented)
	h.eventBus.Subscribe(domain_event.PostEventTypeCommentRemoved, h.handlePostCommentRemoved)
}

func (h *PostEventHandler) handlePostCreated(ctx context.Context, event shared.DomainEvent) error {
	created, ok := event.(*domain_event.PostCreatedEvent)
	if !ok {
		slog.Warn("Unexpected event type in handlePostCreated", "actual_type", eventType(event))
		return nil
	}

	slog.Info("Handling post created",
		"post_id", created.AggregateID(),
		"author_id", created.AuthorID.String(),
		"community_id", created.CommunityID.String(),
	)
	return nil
}

func (h *PostEventHandler) handlePostUpdated(ctx context.Context, event shared.DomainEvent) error {
	updated, ok := event.(*domain_event.PostUpdatedEvent)
	if !ok {
		slog.Warn("Unexpected event type in handlePostUpdated", "actual_type", eventType(event))
		return nil
	}
	slog.Debug("Handling post updated", "post_id", updated.AggregateID())
	return nil
}

func (h *PostEventHandler) handlePostDeleted(ctx context.Context, event shared.DomainEvent) error {
	deleted, ok := event.(*domain_event.PostDeletedEvent)
	if !ok {
		slog.Warn("Unexpected event type in handlePostDeleted", "actual_type", eventType(event))
		return nil
	}
	slog.Info("Handling post deleted", "post_id", deleted.AggregateID())
	return nil
}

func (h *PostEventHandler) handlePostVisibilityToggled(ctx context.Context, event shared.DomainEvent) error {
	toggled, ok := event.(*domain_event.PostVisibilityToggledEvent)
	if !ok {
		slog.Warn("Unexpected event type in handlePostVisibilityToggled", "actual_type", eventType(event))
		return nil
	}
	slog.Debug("Handling post visibility toggled", "post_id", toggled.AggregateID(), "is_public", toggled.IsPublic)
	return nil
}

func (h *PostEventHandler) handlePostLiked(ctx context.Context, event shared.DomainEvent) error {
	liked, ok := event.(*domain_event.PostLikedEvent)
	if !ok {
		slog.Warn("Unexpected event type in handlePostLiked", "actual_type", eventType(event))
		return nil
	}
	slog.Debug("Handling post liked", "post_id", liked.AggregateID(), "likes_count", liked.LikesCount)
	return nil
}

func (h *PostEventHandler) handlePostUnliked(ctx context.Context, event shared.DomainEvent) error {
	unliked, ok := event.(*domain_event.PostUnlikedEvent)
	if !ok {
		slog.Warn("Unexpected event type in handlePostUnliked", "actual_type", eventType(event))
		return nil
	}
	slog.Debug("Handling post unliked", "post_id", unliked.AggregateID(), "likes_count", unliked.LikesCount)
	return nil
}

func (h *PostEventHandler) handlePostCommented(ctx context.Context, event shared.DomainEvent) error {
	commented, ok := event.(*domain_event.PostCommentedEvent)
	if !ok {
		slog.Warn("Unexpected event type in handlePostCommented", "actual_type", eventType(event))
		return nil
	}
	slog.Debug("Handling post commented", "post_id", commented.AggregateID(), "comment_count", commented.CommentCount)
	return nil
}

func (h *PostEventHandler) handlePostCommentRemoved(ctx context.Context, event shared.DomainEvent) error {
	removed, ok := event.(*domain_event.PostCommentRemovedEvent)
	if !ok {
		slog.Warn("Unexpected event type in handlePostCommentRemoved", "actual_type", eventType(event))
		return nil
	}
	slog.Debug("Handling post comment removed", "post_id", removed.AggregateID(), "comment_count", removed.CommentCount)
	return nil
}

func eventType(e shared.DomainEvent) string {
	if e == nil {
		return "<nil>"
	}
	return e.EventType()
}
