// Package ports defines request/response structs for post command controllers.
package ports

import (
	"time"

	"github.com/google/uuid"
)

// CreatePostRequest represents the input to create a new post.
type CreatePostRequest struct {
	Title       string    `json:"title" binding:"required,min=3,max=150"`
	Content     string    `json:"content" binding:"required,min=10,max=5000"`
	CommunityID uuid.UUID `json:"community_id" binding:"required"`
	IsPublic    bool      `json:"is_public" binding:"required"`
}

// UpdatePostRequest represents the input to update an existing post.
type UpdatePostRequest struct {
	Title   string `json:"title" binding:"omitempty,min=3,max=150"`
	Content string `json:"content" binding:"omitempty,min=10,max=5000"`
}

// TogglePostVisibilityRequest toggles a post's visibility.
type TogglePostVisibilityRequest struct {
	IsPublic bool `json:"is_public"`
}

// LikePostRequest – no fields needed; post ID from URL, user from context
type LikePostRequest struct{}

// UnlikePostRequest – same
type UnlikePostRequest struct{}

// AddCommentToPostRequest – same
type AddCommentToPostRequest struct{}

// RemoveCommentFromPostRequest – same
type RemoveCommentFromPostRequest struct{}

// DeletePostRequest – no fields; everything from URL + context
type DeletePostRequest struct{}

type PostResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	AuthorID    string    `json:"author_id"`
	CommunityID string    `json:"community_id"`
	IsPublic    bool      `json:"is_public"`
	Likes       int       `json:"likes"`
	Comments    int       `json:"comments"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
