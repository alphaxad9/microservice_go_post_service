// github.com/alphaxad9/my-go-backend/post_service/src/posts/domain/repos/command_repo.go
package repos

import (
	"context"

	post_aggregate "github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"

	"github.com/google/uuid"
)

// PostCommandRepository handles persistence of PostAggregate.
// Used exclusively by command handlers and application services.
// Enforces transactional consistency.
type PostCommandRepository interface {
	// Create persists a new post aggregate.
	Create(ctx context.Context, agg *post_aggregate.PostAggregate) error

	// Update persists changes to an existing post aggregate.
	Update(ctx context.Context, agg *post_aggregate.PostAggregate) error

	// Delete removes a post (soft or hard delete).
	Delete(ctx context.Context, postID uuid.UUID) error

	// GetByID loads a post as an aggregate for modification.
	GetByID(ctx context.Context, postID uuid.UUID) (*post_aggregate.PostAggregate, error)

	// Exists checks existence before loading (for validation).
	Exists(ctx context.Context, postID uuid.UUID) (bool, error)

	// ExistsWithTitleInCommunity checks uniqueness of title in a community (if needed).
	ExistsWithTitleInCommunity(ctx context.Context, communityID uuid.UUID, title string) (bool, error)
}
