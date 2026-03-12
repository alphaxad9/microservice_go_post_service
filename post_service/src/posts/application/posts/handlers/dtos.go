// github.com/alphaxad9/my-go-backend/post_service/src/posts/application/posts/handlers/dtos.go
package handlers

import (
	"time"

	"github.com/alphaxad9/my-go-backend/post_service/external"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"

	"github.com/google/uuid"
)

// PostResponseDTO represents a post returned in API responses,
// enriched with author details via UserView.
type PostResponseDTO struct {
	ID           uuid.UUID         `json:"id"`
	Author       external.UserView `json:"author"`
	CommunityID  uuid.UUID         `json:"community_id"`
	Title        string            `json:"title"`
	Content      string            `json:"content"`
	IsPublic     bool              `json:"is_public"`
	LikesCount   int               `json:"likes_count"`
	CommentCount int               `json:"comment_count"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// ToPostResponseDTO converts a domain.PostView to PostResponseDTO.
// Requires the associated author as external.UserView.
func ToPostResponseDTO(post *domain.PostView, author external.UserView) PostResponseDTO {
	return PostResponseDTO{
		ID:           post.ID,
		Author:       author,
		CommunityID:  post.CommunityID,
		Title:        post.Title,
		Content:      post.Content,
		IsPublic:     post.IsPublic,
		LikesCount:   post.LikesCount,
		CommentCount: post.CommentCount,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
	}
}

// ToPostListResponseDTO converts a slice of domain.PostView to a paginated response.
// Requires a map of author UserViews keyed by AuthorID.
func ToPostListResponseDTO(
	posts []*domain.PostView,
	authors map[uuid.UUID]external.UserView,
	page, pageSize, totalCount int,
) PostListResponseDTO {
	postDTOs := make([]PostResponseDTO, len(posts))
	for i, post := range posts {
		author, ok := authors[post.AuthorID]
		if !ok {
			// Fallback: create minimal UserView if author data is missing (optional)
			author = external.UserView{UserID: post.AuthorID}
		}
		postDTOs[i] = ToPostResponseDTO(post, author)
	}

	return PostListResponseDTO{
		Posts:      postDTOs,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		HasMore:    (page * pageSize) < totalCount,
	}
}

// PostListResponseDTO represents a paginated list of posts.
type PostListResponseDTO struct {
	Posts      []PostResponseDTO `json:"posts"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	HasMore    bool              `json:"has_more"`
}
