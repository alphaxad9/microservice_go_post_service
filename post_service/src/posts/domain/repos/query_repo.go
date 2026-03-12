// github.com/alphaxad9/my-go-backend/post_service/src/posts/domain/repos/query_repo.go
package repos

import (
	"context"

	post_domain "github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"

	"github.com/google/uuid"
)

// PostQueryRepository defines read-only operations returning *post_domain.PostView.
// Used by query services, APIs, or reporting logic.
// Implementations may use different data sources (e.g., read replica, cache).
type PostQueryRepository interface {
	// ByID retrieves a post by its unique ID.
	ByID(ctx context.Context, postID uuid.UUID) (*post_domain.PostView, error)

	// ByAuthor retrieves all posts by a given author (paginated).
	ByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*post_domain.PostView, error)

	// ByCommunity retrieves posts in a community (paginated, public only if requester not member).
	ByCommunity(ctx context.Context, communityID uuid.UUID, requesterID *uuid.UUID, limit, offset int) ([]*post_domain.PostView, error)

	// SearchPosts performs full-text or keyword search (optional).
	SearchPosts(ctx context.Context, query string, limit, offset int) ([]*post_domain.PostView, error)

	// CountByAuthor returns total number of posts by an author.
	CountByAuthor(ctx context.Context, authorID uuid.UUID) (int, error)

	// Exists checks if a post exists.
	Exists(ctx context.Context, postID uuid.UUID) (bool, error)
}
