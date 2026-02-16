// my-go-backend/post_service/src/posts/infra/repositories/posts/orm_query_repository.go
package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"my-go-backend/post_service/src/posts/domain"
	"my-go-backend/post_service/src/posts/infra/models"
	"my-go-backend/post_service/src/posts/ports"
)

type PostQueryRepository struct {
	db *sql.DB
}

func NewPostQueryRepository(db *sql.DB) *PostQueryRepository {
	return &PostQueryRepository{db: db}
}

func (r *PostQueryRepository) ByID(ctx context.Context, postID uuid.UUID) (*domain.PostView, error) {
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
	return model.ToDomainView(), nil
}

func (r *PostQueryRepository) ByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.PostView, error) {
	rows, err := r.db.QueryContext(ctx, ports.QuerySelectPostsByAuthor, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts by author: %w", err)
	}
	defer rows.Close()

	var views []*domain.PostView
	for rows.Next() {
		var model models.PostModel
		if err := rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}
		views = append(views, model.ToDomainView())
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return views, nil
}

func (r *PostQueryRepository) ByCommunity(
	ctx context.Context,
	communityID uuid.UUID,
	requesterID *uuid.UUID,
	limit, offset int,
) ([]*domain.PostView, error) {
	// Pass requesterID as NULL if not provided
	var requesterOrNil interface{}
	if requesterID != nil {
		requesterOrNil = *requesterID
	} else {
		requesterOrNil = nil // 👈 This is fine
	}

	rows, err := r.db.QueryContext(
		ctx,
		ports.QuerySelectPostsByCommunity,
		communityID,
		requesterOrNil,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts by community: %w", err)
	}
	defer rows.Close()

	var views []*domain.PostView
	for rows.Next() {
		var model models.PostModel
		if err := rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}
		views = append(views, model.ToDomainView())
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return views, nil
}

func (r *PostQueryRepository) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*domain.PostView, error) {
	rows, err := r.db.QueryContext(ctx, ports.QuerySearchPosts, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}
	defer rows.Close()

	var views []*domain.PostView
	for rows.Next() {
		var model models.PostModel
		if err := rows.Scan(
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
		); err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}
		views = append(views, model.ToDomainView())
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return views, nil
}

func (r *PostQueryRepository) CountByAuthor(ctx context.Context, authorID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, ports.QueryCountPostsByAuthor, authorID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // no posts is valid
		}
		return 0, fmt.Errorf("failed to count posts by author: %w", err)
	}
	return count, nil
}

func (r *PostQueryRepository) Exists(ctx context.Context, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, ports.QueryExistsPostByID, postID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check post existence: %w", err)
	}
	return exists, nil
}
