// Package services implements application services for post domain queries.
// github.com/alphaxad9/my-go-backend/post_service/src/posts/application/posts/services/query_service.go
package services

import (
	"context"

	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain/repos"
	"github.com/alphaxad9/my-go-backend/post_service/src/shared"

	"github.com/google/uuid"
)

// PostQueryService defines the application-level interface for reading posts.
type PostQueryService interface {
	GetPostByID(ctx context.Context, postID uuid.UUID) (*domain.PostView, error)
	GetPostsByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.PostView, error)
	GetPostsByCommunity(ctx context.Context, communityID uuid.UUID, requesterID *uuid.UUID, limit, offset int) ([]*domain.PostView, error)
	SearchPosts(ctx context.Context, query string, limit, offset int) ([]*domain.PostView, error)
	GetPostCountByAuthor(ctx context.Context, authorID uuid.UUID) (int, error)
	PostExists(ctx context.Context, postID uuid.UUID) (bool, error)
}

var _ PostQueryService = (*PostQueryServiceImpl)(nil)

type PostQueryServiceImpl struct {
	postRepo repos.PostQueryRepository
}

func NewPostQueryService(postRepo repos.PostQueryRepository) PostQueryService {
	return &PostQueryServiceImpl{
		postRepo: postRepo,
	}
}

// GetPostByID retrieves a single post by its ID.
func (s *PostQueryServiceImpl) GetPostByID(ctx context.Context, postID uuid.UUID) (*domain.PostView, error) {
	if postID == uuid.Nil {
		return nil, domain.NewPostNotFoundError("")
	}

	postView, err := s.postRepo.ByID(ctx, postID)
	if err != nil {
		if err == domain.ErrPostNotFound {
			return nil, domain.NewPostNotFoundError(postID.String())
		}
		return nil, shared.NewInternalServerError(err)
	}

	return postView, nil
}

// GetPostsByAuthor retrieves all posts authored by a given user (paginated).
func (s *PostQueryServiceImpl) GetPostsByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.PostView, error) {
	if authorID == uuid.Nil {
		return nil, domain.NewInvalidPostDataError("author_id is required")
	}

	posts, err := s.postRepo.ByAuthor(ctx, authorID, limit, offset)
	if err != nil {
		return nil, shared.NewInternalServerError(err)
	}

	return posts, nil
}

// GetPostsByCommunity retrieves posts in a community.
// If requesterID is provided and matches a member of the community (assumed via authz layer),
// private posts may be included—this logic is delegated to the repository implementation.
func (s *PostQueryServiceImpl) GetPostsByCommunity(
	ctx context.Context,
	communityID uuid.UUID,
	requesterID *uuid.UUID,
	limit, offset int,
) ([]*domain.PostView, error) {
	if communityID == uuid.Nil {
		return nil, domain.NewInvalidPostDataError("community_id is required")
	}

	posts, err := s.postRepo.ByCommunity(ctx, communityID, requesterID, limit, offset)
	if err != nil {
		return nil, shared.NewInternalServerError(err)
	}

	return posts, nil
}

// SearchPosts performs a keyword-based or full-text search across posts.
func (s *PostQueryServiceImpl) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*domain.PostView, error) {
	if query == "" {
		return []*domain.PostView{}, nil // empty query → no results
	}

	posts, err := s.postRepo.SearchPosts(ctx, query, limit, offset)
	if err != nil {
		return nil, shared.NewInternalServerError(err)
	}

	return posts, nil
}

// GetPostCountByAuthor returns the total number of posts by an author.
func (s *PostQueryServiceImpl) GetPostCountByAuthor(ctx context.Context, authorID uuid.UUID) (int, error) {
	if authorID == uuid.Nil {
		return 0, domain.NewInvalidPostDataError("author_id is required")
	}

	count, err := s.postRepo.CountByAuthor(ctx, authorID)
	if err != nil {
		return 0, shared.NewInternalServerError(err)
	}

	return count, nil
}

// PostExists checks whether a post with the given ID exists.
func (s *PostQueryServiceImpl) PostExists(ctx context.Context, postID uuid.UUID) (bool, error) {
	if postID == uuid.Nil {
		return false, nil
	}

	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		return false, shared.NewInternalServerError(err)
	}

	return exists, nil
}
