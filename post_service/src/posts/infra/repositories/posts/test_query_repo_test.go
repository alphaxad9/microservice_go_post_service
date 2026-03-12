// Package repositories tests the PostgreSQL implementation of domain query repositories.
package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"
)

// Helper to insert a post directly for query testing
func insertPost(t *testing.T, db *sql.DB, post *domain.PostView) {
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO posts (
			id, author_id, community_id, title, content, is_public,
			likes_count, comment_count, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		post.ID,
		post.AuthorID,
		post.CommunityID,
		post.Title,
		post.Content,
		post.IsPublic,
		post.LikesCount,
		post.CommentCount,
		post.CreatedAt,
		post.UpdatedAt,
	)
	require.NoError(t, err)
}

func TestPostQueryRepository_ByCommunity_PublicOnlyWhenNoRequester(t *testing.T) {
	ResetPosts(t)

	repo := NewPostQueryRepository(GetTestDB())
	ctx := context.Background()

	communityID := uuid.New()
	authorID := uuid.New()
	now := time.Now().UTC()

	publicPost := &domain.PostView{
		ID:           uuid.New(),
		AuthorID:     authorID,
		CommunityID:  communityID,
		Title:        "Public Post",
		Content:      "Visible to all",
		IsPublic:     true,
		LikesCount:   0,
		CommentCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	privatePost := &domain.PostView{
		ID:           uuid.New(),
		AuthorID:     authorID,
		CommunityID:  communityID,
		Title:        "Private Post",
		Content:      "Only for members/authors",
		IsPublic:     false, // 👈 double-check
		LikesCount:   0,
		CommentCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	insertPost(t, GetTestDB(), publicPost)
	insertPost(t, GetTestDB(), privatePost)

	// DEBUG: Check what's actually in the DB
	var count int
	err := GetTestDB().QueryRowContext(ctx, "SELECT COUNT(*) FROM posts").Scan(&count)
	require.NoError(t, err)
	t.Logf("Total posts in DB: %d", count)

	var publicCount int
	err = GetTestDB().QueryRowContext(ctx, "SELECT COUNT(*) FROM posts WHERE is_public = true").Scan(&publicCount)
	require.NoError(t, err)
	t.Logf("Public posts: %d", publicCount)

	// Now run the test
	results, err := repo.ByCommunity(ctx, communityID, nil, 10, 0)
	require.NoError(t, err)
	t.Logf("Returned %d posts", len(results))
	for _, p := range results {
		t.Logf("Post: %s (public: %t)", p.Title, p.IsPublic)
	}

	assert.Len(t, results, 1)
}

func TestPostQueryRepository_ByID(t *testing.T) {
	ResetPosts(t)

	repo := NewPostQueryRepository(GetTestDB())
	ctx := context.Background()

	authorID := uuid.New()
	communityID := uuid.New()
	postID := uuid.New()
	now := time.Now().UTC()

	expected := &domain.PostView{
		ID:           postID,
		AuthorID:     authorID,
		CommunityID:  communityID,
		Title:        "Test Post",
		Content:      "This is a test post.",
		IsPublic:     true,
		LikesCount:   5,
		CommentCount: 2,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	insertPost(t, GetTestDB(), expected)

	actual, err := repo.ByID(ctx, postID)
	require.NoError(t, err)

	// Compare non-time fields exactly
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.AuthorID, actual.AuthorID)
	assert.Equal(t, expected.CommunityID, actual.CommunityID)
	assert.Equal(t, expected.Title, actual.Title)
	assert.Equal(t, expected.Content, actual.Content)
	assert.Equal(t, expected.IsPublic, actual.IsPublic)
	assert.Equal(t, expected.LikesCount, actual.LikesCount)
	assert.Equal(t, expected.CommentCount, actual.CommentCount)

	// Compare time fields with microsecond tolerance (1ms is safe)
	assert.WithinDuration(t, expected.CreatedAt, actual.CreatedAt, time.Millisecond)
	assert.WithinDuration(t, expected.UpdatedAt, actual.UpdatedAt, time.Millisecond)
}

func TestPostQueryRepository_ByID_NotFound(t *testing.T) {
	ResetPosts(t)

	repo := NewPostQueryRepository(GetTestDB())
	ctx := context.Background()

	_, err := repo.ByID(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Post not found")
}

func TestPostQueryRepository_ByAuthor(t *testing.T) {
	ResetPosts(t)

	repo := NewPostQueryRepository(GetTestDB())
	ctx := context.Background()

	authorID := uuid.New()
	otherAuthorID := uuid.New()
	communityID := uuid.New()
	now := time.Now().UTC()

	// Insert 3 posts by author
	posts := []*domain.PostView{
		{
			ID:           uuid.New(),
			AuthorID:     authorID,
			CommunityID:  communityID,
			Title:        "Post 1",
			Content:      "Content 1",
			IsPublic:     true,
			LikesCount:   1,
			CommentCount: 0,
			CreatedAt:    now.Add(-2 * time.Hour),
			UpdatedAt:    now.Add(-2 * time.Hour),
		},
		{
			ID:           uuid.New(),
			AuthorID:     authorID,
			CommunityID:  communityID,
			Title:        "Post 2",
			Content:      "Content 2",
			IsPublic:     false,
			LikesCount:   2,
			CommentCount: 1,
			CreatedAt:    now.Add(-1 * time.Hour),
			UpdatedAt:    now.Add(-1 * time.Hour),
		},
		{
			ID:           uuid.New(),
			AuthorID:     authorID,
			CommunityID:  communityID,
			Title:        "Post 3",
			Content:      "Content 3",
			IsPublic:     true,
			LikesCount:   3,
			CommentCount: 2,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	for _, p := range posts {
		insertPost(t, GetTestDB(), p)
	}

	// Insert one by another author (should not appear)
	otherPost := &domain.PostView{
		ID:           uuid.New(),
		AuthorID:     otherAuthorID,
		CommunityID:  communityID,
		Title:        "Other Author",
		Content:      "Not ours",
		IsPublic:     true,
		LikesCount:   0,
		CommentCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	insertPost(t, GetTestDB(), otherPost)

	// Fetch all (limit 10, offset 0)
	results, err := repo.ByAuthor(ctx, authorID, 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 3)

	// Should be ordered by created_at DESC → Post 3, 2, 1
	assert.Equal(t, "Post 3", results[0].Title)
	assert.Equal(t, "Post 2", results[1].Title)
	assert.Equal(t, "Post 1", results[2].Title)

	// Pagination: limit 1, offset 1 → should get Post 2
	paginated, err := repo.ByAuthor(ctx, authorID, 1, 1)
	require.NoError(t, err)
	assert.Len(t, paginated, 1)
	assert.Equal(t, "Post 2", paginated[0].Title)
}
func TestPostQueryRepository_SearchPosts(t *testing.T) {
	ResetPosts(t)

	repo := NewPostQueryRepository(GetTestDB())
	ctx := context.Background()

	authorID := uuid.New()
	communityID := uuid.New()
	now := time.Now().UTC()

	posts := []*domain.PostView{
		{
			ID:           uuid.New(),
			AuthorID:     authorID,
			CommunityID:  communityID,
			Title:        "Go Programming Guide",
			Content:      "Learn Go from scratch.",
			IsPublic:     true,
			LikesCount:   10,
			CommentCount: 3,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.New(),
			AuthorID:     authorID,
			CommunityID:  communityID,
			Title:        "Python Tips",
			Content:      "Best practices in Python.",
			IsPublic:     true,
			LikesCount:   5,
			CommentCount: 1,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           uuid.New(),
			AuthorID:     authorID,
			CommunityID:  communityID,
			Title:        "Database Design",
			Content:      "How to normalize tables.",
			IsPublic:     true,
			LikesCount:   7,
			CommentCount: 2,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}

	for _, p := range posts {
		insertPost(t, GetTestDB(), p)
	}

	// Search for "Go"
	results, err := repo.SearchPosts(ctx, "Go", 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Go Programming Guide", results[0].Title)

	// Search for "design" (case-insensitive, stemmed)
	// Note: This assumes your SQL uses ILIKE or full-text search.
	// If you're using simple LIKE, you may need to adjust.
	results, err = repo.SearchPosts(ctx, "design", 10, 0)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Database Design", results[0].Title)
}
func TestPostQueryRepository_CountByAuthor(t *testing.T) {
	ResetPosts(t)

	repo := NewPostQueryRepository(GetTestDB())
	ctx := context.Background()

	authorID := uuid.New()
	otherAuthor := uuid.New()
	communityID := uuid.New()
	now := time.Now().UTC()

	// Insert 2 posts by author
	for i := 0; i < 2; i++ {
		p := &domain.PostView{
			ID:           uuid.New(),
			AuthorID:     authorID,
			CommunityID:  communityID,
			Title:        "Post " + string(rune('1'+i)),
			Content:      "Content",
			IsPublic:     true,
			LikesCount:   0,
			CommentCount: 0,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		insertPost(t, GetTestDB(), p)
	}

	// Insert 1 by other author
	other := &domain.PostView{
		ID:           uuid.New(),
		AuthorID:     otherAuthor,
		CommunityID:  communityID,
		Title:        "Other",
		Content:      "Other",
		IsPublic:     true,
		LikesCount:   0,
		CommentCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	insertPost(t, GetTestDB(), other)

	count, err := repo.CountByAuthor(ctx, authorID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	count, err = repo.CountByAuthor(ctx, otherAuthor)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = repo.CountByAuthor(ctx, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
