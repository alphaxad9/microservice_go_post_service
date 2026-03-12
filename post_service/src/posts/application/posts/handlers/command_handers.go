// Package handlers provides command handlers for post mutations.
// / github.com/alphaxad9/my-go-backend/post_service/src/posts/application/posts/handlers/command_handlers.go
package handlers

import (
	"context"

	"github.com/alphaxad9/my-go-backend/post_service/src/posts/application/posts/services"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"

	"github.com/google/uuid"
)

// PostCommandHandler encapsulates the command service and exposes high-level use-case methods.
type PostCommandHandler struct {
	postService services.PostCommandService
}

// NewPostCommandHandler creates a new PostCommandHandler.
func NewPostCommandHandler(postService services.PostCommandService) *PostCommandHandler {
	return &PostCommandHandler{
		postService: postService,
	}
}

// CreatePost creates a new post and returns the resulting aggregate.
func (h *PostCommandHandler) CreatePost(
	ctx context.Context,
	title, content string,
	authorID, communityID uuid.UUID,
	isPublic bool,
) (*domain.PostAggregate, error) {
	return h.postService.CreatePost(ctx, title, content, authorID, communityID, isPublic)
}

// UpdatePost updates an existing post's title and content.
func (h *PostCommandHandler) UpdatePost(
	ctx context.Context,
	postID uuid.UUID,
	newTitle, newContent string,
	requesterID uuid.UUID,
) error {
	return h.postService.UpdatePost(ctx, postID, newTitle, newContent, requesterID)
}

// TogglePostVisibility toggles the visibility of a post.
func (h *PostCommandHandler) TogglePostVisibility(
	ctx context.Context,
	postID uuid.UUID,
	isPublic bool,
	requesterID uuid.UUID,
) error {
	return h.postService.TogglePostVisibility(ctx, postID, isPublic, requesterID)
}

// LikePost increments the like count of a post.
func (h *PostCommandHandler) LikePost(ctx context.Context, postID uuid.UUID) error {
	return h.postService.LikePost(ctx, postID)
}

// UnlikePost decrements the like count of a post.
func (h *PostCommandHandler) UnlikePost(ctx context.Context, postID uuid.UUID) error {
	return h.postService.UnlikePost(ctx, postID)
}

// AddCommentToPost increments the comment count of a post.
func (h *PostCommandHandler) AddCommentToPost(ctx context.Context, postID uuid.UUID) error {
	return h.postService.AddCommentToPost(ctx, postID)
}

// RemoveCommentFromPost decrements the comment count of a post.
func (h *PostCommandHandler) RemoveCommentFromPost(ctx context.Context, postID uuid.UUID) error {
	return h.postService.RemoveCommentFromPost(ctx, postID)
}

// DeletePost deletes a post if the requester is the author.
func (h *PostCommandHandler) DeletePost(
	ctx context.Context,
	postID uuid.UUID,
	requesterID uuid.UUID,
) error {
	return h.postService.DeletePost(ctx, postID, requesterID)
}
