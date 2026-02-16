package models

import (
	domain "my-go-backend/post_service/src/posts/domain"
)

// FromDomain converts a PostAggregate to a PostModel for persistence.
// This is used when saving an aggregate to the database.
func FromDomain(agg *domain.PostAggregate) *PostModel {
	return &PostModel{
		ID:           agg.ID(),
		AuthorID:     agg.AuthorID(),
		CommunityID:  agg.CommunityID(),
		Title:        agg.Title(),
		Content:      agg.Content(),
		IsPublic:     agg.IsPublic(),
		LikesCount:   agg.LikesCount(),
		CommentCount: agg.CommentCount(),
		CreatedAt:    agg.CreatedAt(),
		UpdatedAt:    agg.UpdatedAt(),
	}
}

// ToDomainView converts a PostModel to a PostView for read operations.
// This is used when querying data to return to API clients.
func (m *PostModel) ToDomainView() *domain.PostView {
	return &domain.PostView{
		ID:           m.ID,
		AuthorID:     m.AuthorID,
		CommunityID:  m.CommunityID,
		Title:        m.Title,
		Content:      m.Content,
		IsPublic:     m.IsPublic,
		LikesCount:   m.LikesCount,
		CommentCount: m.CommentCount,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
