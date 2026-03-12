// github.com/alphaxad9/my-go-backend/post_service/src/posts/infra/repositories/posts/orm_command_repository.go
package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/infra/models"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/ports"

	"github.com/google/uuid"
)

type PostCommandRepository struct {
	db *sql.DB
}

func NewPostCommandRepository(db *sql.DB) *PostCommandRepository {
	return &PostCommandRepository{db: db}
}

func (r *PostCommandRepository) Create(ctx context.Context, agg *domain.PostAggregate) error {
	model := models.FromDomain(agg)
	_, err := r.db.ExecContext(ctx, ports.QueryInsertPost,
		model.ID,
		model.AuthorID,
		model.CommunityID,
		model.Title,
		model.Content,
		model.IsPublic,
		model.LikesCount,
		model.CommentCount,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert post : %w", err)
	}
	return nil
}

func (r *PostCommandRepository) Update(ctx context.Context, agg *domain.PostAggregate) error {
	model := models.FromDomain(agg)
	result, err := r.db.ExecContext(ctx, ports.QueryUpdatePost,
		model.Title,
		model.Content,
		model.IsPublic,
		model.LikesCount,
		model.CommentCount,
		model.UpdatedAt,
		model.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.NewPostNotFoundError(agg.ID().String())
	}
	return nil
}

func (r *PostCommandRepository) Delete(ctx context.Context, postID uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, ports.QueryDeletePost, postID)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.NewPostNotFoundError(postID.String())
	}
	return nil
}

func (r *PostCommandRepository) GetByID(ctx context.Context, postID uuid.UUID) (*domain.PostAggregate, error) {
	var model models.PostModel
	err := r.db.QueryRowContext(ctx, ports.QuerySelectPostByID, postID).Scan(
		&model.ID,
		&model.AuthorID,
		&model.CommunityID,
		&model.Title,
		&model.Content,
		&model.IsPublic,
		&model.LikesCount,
		&model.CommentCount,
		&model.CreatedAt,
		&model.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.NewPostNotFoundError(postID.String())
		}
		return nil, fmt.Errorf("failed to fetch post by ID: %w", err)
	}

	return domain.ReconstructPostAggregate(
		model.ID,
		model.AuthorID,
		model.CommunityID,
		model.Title,
		model.Content,
		model.IsPublic,
		model.LikesCount,
		model.CommentCount,
		model.CreatedAt,
		model.UpdatedAt,
	), nil
}

func (r *PostCommandRepository) Exists(ctx context.Context, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, ports.QueryExistsPostByID, postID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check post existence: %w", err)
	}
	return exists, nil
}

func (r *PostCommandRepository) ExistsWithTitleInCommunity(
	ctx context.Context,
	communityID uuid.UUID,
	title string,
) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, ports.QueryExistsPostWithTitleInCommunity, communityID, title).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check title uniqueness in community: %w", err)
	}
	return count > 0, nil
}
