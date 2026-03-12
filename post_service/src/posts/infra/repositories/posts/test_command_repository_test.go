// Package repositories tests the PostgreSQL implementation of domain query repositories.
package repositories

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"
)

func TestPostCommandRepository_ExistsWithTitleInCommunity(t *testing.T) {
	ResetPosts(t)

	repo := NewPostCommandRepository(GetTestDB())
	ctx := context.Background()

	communityID := uuid.New()
	authorID := uuid.New()

	// Create first post with VALID content (>=10 chars)
	agg1, err := domain.CreatePost(
		"Unique Title",
		"This is valid content for post one.",
		authorID,
		communityID,
		true,
	)
	require.NoError(t, err)

	err = repo.Create(ctx, agg1)
	require.NoError(t, err)

	// Check existence (case-insensitive match)
	exists, err := repo.ExistsWithTitleInCommunity(ctx, communityID, "unique title")
	require.NoError(t, err)
	assert.True(t, exists)

	// Different community → should not exist
	otherCommunity := uuid.New()
	exists, err = repo.ExistsWithTitleInCommunity(ctx, otherCommunity, "Unique Title")
	require.NoError(t, err)
	assert.False(t, exists)

	// New title in same community → should not exist
	exists, err = repo.ExistsWithTitleInCommunity(ctx, communityID, "Another Title")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestPostCommandRepository_Create(t *testing.T) {
	ResetPosts(t)

	repo := NewPostCommandRepository(GetTestDB())
	ctx := context.Background()

	authorID := uuid.New()
	communityID := uuid.New()

	agg, err := domain.CreatePost(
		"My First Post",
		"This is the content of my first post.",
		authorID,
		communityID,
		true,
	)
	require.NoError(t, err)

	err = repo.Create(ctx, agg)
	require.NoError(t, err)

	exists, err := repo.Exists(ctx, agg.ID())
	require.NoError(t, err)
	assert.True(t, exists)

	fetched, err := repo.GetByID(ctx, agg.ID())
	require.NoError(t, err)
	assert.Equal(t, agg.ID(), fetched.ID())
	assert.Equal(t, authorID, fetched.AuthorID())
	assert.Equal(t, communityID, fetched.CommunityID())
	assert.Equal(t, "My First Post", fetched.Title())
	assert.Equal(t, "This is the content of my first post.", fetched.Content())
	assert.True(t, fetched.IsPublic())
	assert.Equal(t, 0, fetched.LikesCount())
	assert.Equal(t, 0, fetched.CommentCount())
}

func TestPostCommandRepository_Update(t *testing.T) {
	ResetPosts(t) // ✅ Fixed name

	repo := NewPostCommandRepository(GetTestDB()) // ✅ Fixed name
	ctx := context.Background()

	authorID := uuid.New()
	communityID := uuid.New()

	agg, err := domain.CreatePost("Original Title", "Original content", authorID, communityID, true)
	require.NoError(t, err)

	err = repo.Create(ctx, agg)
	require.NoError(t, err)

	// Update
	err = agg.UpdateContent("Updated Title", "Updated content body.", authorID)
	require.NoError(t, err)

	err = repo.Update(ctx, agg)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, agg.ID())
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title())
	assert.Equal(t, "Updated content body.", updated.Content())
	assert.NotEqual(t, agg.CreatedAt(), updated.UpdatedAt())
}

func TestPostCommandRepository_Delete(t *testing.T) {
	ResetPosts(t) // ✅

	repo := NewPostCommandRepository(GetTestDB()) // ✅
	ctx := context.Background()

	authorID := uuid.New()
	communityID := uuid.New()

	agg, err := domain.CreatePost("To Delete", "Delete me.", authorID, communityID, true)
	require.NoError(t, err)

	err = repo.Create(ctx, agg)
	require.NoError(t, err)

	// Delete
	err = repo.Delete(ctx, agg.ID())
	require.NoError(t, err)

	// Should not exist
	exists, err := repo.Exists(ctx, agg.ID())
	require.NoError(t, err)
	assert.False(t, exists)

	// GetByID should fail
	_, err = repo.GetByID(ctx, agg.ID())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Post not found")
}

func TestPostCommandRepository_Delete_NonExistent(t *testing.T) {
	ResetPosts(t) // ✅

	repo := NewPostCommandRepository(GetTestDB()) // ✅
	ctx := context.Background()

	err := repo.Delete(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Post not found")
}
func TestPostCommandRepository_GetByID_NotFound(t *testing.T) {
	ResetPosts(t) // ✅

	repo := NewPostCommandRepository(GetTestDB()) // ✅
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Post not found")
}

func TestPostCommandRepository_Update_NonExistent(t *testing.T) {
	ResetPosts(t)

	repo := NewPostCommandRepository(GetTestDB())
	ctx := context.Background()

	// Use VALID inputs that satisfy domain rules
	title := "A Valid Post Title"
	content := "This content has more than ten characters."
	authorID := uuid.New()
	communityID := uuid.New()

	agg, err := domain.CreatePost(title, content, authorID, communityID, true)
	require.NoError(t, err) // This will now pass

	// Attempt update without inserting
	err = repo.Update(ctx, agg)
	require.Error(t, err)

	// Assert expected "not found" error
	assert.Contains(t, err.Error(), "Post not found")
}

func TestPostCommandRepository_Exists(t *testing.T) {
	ResetPosts(t)

	repo := NewPostCommandRepository(GetTestDB())
	ctx := context.Background()

	// Check non-existent ID
	id := uuid.New()
	exists, err := repo.Exists(ctx, id)
	require.NoError(t, err)
	assert.False(t, exists)

	// Create a VALID post (content >= 10 chars)
	agg, err := domain.CreatePost(
		"Exists?",
		"This is valid content with enough characters.",
		uuid.New(),
		uuid.New(),
		true,
	)
	require.NoError(t, err)

	// Persist it
	err = repo.Create(ctx, agg)
	require.NoError(t, err)

	// Now check existence by its real ID
	exists, err = repo.Exists(ctx, agg.ID())
	require.NoError(t, err)
	assert.True(t, exists)
}
