package events

import (
	"github.com/alphaxad9/my-go-backend/post_service/src/shared"

	post_aggregate "github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"

	"github.com/google/uuid"
)

// -----------------------
// Post Event Types (Constants)
// -----------------------

const (
	PostEventTypeCreated           = "go_post.created"
	PostEventTypeUpdated           = "go_post.updated"
	PostEventTypeVisibilityToggled = "go_post.visibility_toggled"
	PostEventTypeLiked             = "go_post.liked"
	PostEventTypeUnliked           = "go_post.unliked"
	PostEventTypeCommented         = "go_post.commented"
	PostEventTypeCommentRemoved    = "go_post.comment_removed"
	PostEventTypeDeleted           = "go_post.deleted"
)

// -----------------------
// Base Post Event
// -----------------------

// PostEvent embeds shared.BaseEvent and adds post-specific context.
// All concrete post events should embed this anonymously.
type PostEvent struct {
	shared.BaseEvent
}

// NewPostEvent creates a base post event with common metadata.
func NewPostEvent(postID uuid.UUID, eventType string) PostEvent {
	return PostEvent{
		BaseEvent: shared.NewBaseEvent(
			postID.String(),
			"Post",
			eventType,
		),
	}
}

// -----------------------
// Concrete Post Events
// -----------------------

// PostCreatedEvent is emitted when a new post is created.
type PostCreatedEvent struct {
	PostEvent
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	AuthorID    uuid.UUID `json:"author_id"`
	CommunityID uuid.UUID `json:"community_id"`
	IsPublic    bool      `json:"is_public"`
}

func NewPostCreatedEvent(agg *post_aggregate.PostAggregate) *PostCreatedEvent {
	return &PostCreatedEvent{
		PostEvent:   NewPostEvent(agg.ID(), PostEventTypeCreated),
		Title:       agg.Title(),
		Content:     agg.Content(),
		AuthorID:    agg.AuthorID(),
		CommunityID: agg.CommunityID(),
		IsPublic:    agg.IsPublic(),
	}
}

func (e *PostCreatedEvent) ToMap() map[string]interface{} {
	base := e.BaseEvent.ToMap()
	base["payload"] = map[string]interface{}{
		"title":        e.Title,
		"content":      e.Content,
		"author_id":    e.AuthorID.String(),
		"community_id": e.CommunityID.String(),
		"is_public":    e.IsPublic,
	}
	return base
}

// PostUpdatedEvent is emitted when post content is modified.
type PostUpdatedEvent struct {
	PostEvent
	Title   string `json:"title"`
	Content string `json:"content"`
}

func NewPostUpdatedEvent(agg *post_aggregate.PostAggregate) *PostUpdatedEvent {
	return &PostUpdatedEvent{
		PostEvent: NewPostEvent(agg.ID(), PostEventTypeUpdated),
		Title:     agg.Title(),
		Content:   agg.Content(),
	}
}

func (e *PostUpdatedEvent) ToMap() map[string]interface{} {
	base := e.BaseEvent.ToMap()
	base["payload"] = map[string]interface{}{
		"title":   e.Title,
		"content": e.Content,
	}
	return base
}

// PostVisibilityToggledEvent is emitted when visibility changes.
type PostVisibilityToggledEvent struct {
	PostEvent
	IsPublic bool `json:"is_public"`
}

func NewPostVisibilityToggledEvent(agg *post_aggregate.PostAggregate) *PostVisibilityToggledEvent {
	return &PostVisibilityToggledEvent{
		PostEvent: NewPostEvent(agg.ID(), PostEventTypeVisibilityToggled),
		IsPublic:  agg.IsPublic(),
	}
}

func (e *PostVisibilityToggledEvent) ToMap() map[string]interface{} {
	base := e.BaseEvent.ToMap()
	base["payload"] = map[string]interface{}{
		"is_public": e.IsPublic,
	}
	return base
}

// PostLikedEvent is emitted when a user likes a post.
type PostLikedEvent struct {
	PostEvent
	LikesCount int `json:"likes_count"`
}

func NewPostLikedEvent(agg *post_aggregate.PostAggregate) *PostLikedEvent {
	return &PostLikedEvent{
		PostEvent:  NewPostEvent(agg.ID(), PostEventTypeLiked),
		LikesCount: agg.LikesCount(),
	}
}

func (e *PostLikedEvent) ToMap() map[string]interface{} {
	base := e.BaseEvent.ToMap()
	base["payload"] = map[string]interface{}{
		"likes_count": e.LikesCount,
	}
	return base
}

// PostUnlikedEvent is emitted when a user removes their like.
type PostUnlikedEvent struct {
	PostEvent
	LikesCount int `json:"likes_count"`
}

func NewPostUnlikedEvent(agg *post_aggregate.PostAggregate) *PostUnlikedEvent {
	return &PostUnlikedEvent{
		PostEvent:  NewPostEvent(agg.ID(), PostEventTypeUnliked),
		LikesCount: agg.LikesCount(),
	}
}

func (e *PostUnlikedEvent) ToMap() map[string]interface{} {
	base := e.BaseEvent.ToMap()
	base["payload"] = map[string]interface{}{
		"likes_count": e.LikesCount,
	}
	return base
}

// PostCommentedEvent is emitted when a comment is added.
type PostCommentedEvent struct {
	PostEvent
	CommentCount int `json:"comment_count"`
}

func NewPostCommentedEvent(agg *post_aggregate.PostAggregate) *PostCommentedEvent {
	return &PostCommentedEvent{
		PostEvent:    NewPostEvent(agg.ID(), PostEventTypeCommented),
		CommentCount: agg.CommentCount(),
	}
}

func (e *PostCommentedEvent) ToMap() map[string]interface{} {
	base := e.BaseEvent.ToMap()
	base["payload"] = map[string]interface{}{
		"comment_count": e.CommentCount,
	}
	return base
}

// PostCommentRemovedEvent is emitted when a comment is removed.
type PostCommentRemovedEvent struct {
	PostEvent
	CommentCount int `json:"comment_count"`
}

func NewPostCommentRemovedEvent(agg *post_aggregate.PostAggregate) *PostCommentRemovedEvent {
	return &PostCommentRemovedEvent{
		PostEvent:    NewPostEvent(agg.ID(), PostEventTypeCommentRemoved),
		CommentCount: agg.CommentCount(),
	}
}

func (e *PostCommentRemovedEvent) ToMap() map[string]interface{} {
	base := e.BaseEvent.ToMap()
	base["payload"] = map[string]interface{}{
		"comment_count": e.CommentCount,
	}
	return base
}

// PostDeletedEvent is emitted when a post is deleted.
type PostDeletedEvent struct {
	PostEvent
	Title       string    `json:"title"`
	AuthorID    uuid.UUID `json:"author_id"`
	CommunityID uuid.UUID `json:"community_id"`
}

func NewPostDeletedEvent(agg *post_aggregate.PostAggregate) *PostDeletedEvent {
	return &PostDeletedEvent{
		PostEvent:   NewPostEvent(agg.ID(), PostEventTypeDeleted),
		Title:       agg.Title(),
		AuthorID:    agg.AuthorID(),
		CommunityID: agg.CommunityID(),
	}
}

func (e *PostDeletedEvent) ToMap() map[string]interface{} {
	base := e.BaseEvent.ToMap()
	base["payload"] = map[string]interface{}{
		"title":        e.Title,
		"author_id":    e.AuthorID.String(),
		"community_id": e.CommunityID.String(),
	}
	return base
}
