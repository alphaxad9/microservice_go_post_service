// github.com/alphaxad9/my-go-backend/post_service/src/posts/domain/aggregate.go
package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// PostAggregate encapsulates the Post entity and enforces business rules.
// It is the transactional boundary for all post-related operations.
type PostAggregate struct {
	// Immutable identity
	id          uuid.UUID
	authorID    uuid.UUID
	communityID uuid.UUID

	// Mutable state
	title        string
	content      string
	isPublic     bool
	likesCount   int
	commentCount int

	// Metadata
	createdAt time.Time
	updatedAt time.Time
}

// ID returns the aggregate's unique identifier.
func (a *PostAggregate) ID() uuid.UUID { return a.id }

// AuthorID returns the author's ID.
func (a *PostAggregate) AuthorID() uuid.UUID { return a.authorID }

// CommunityID returns the community ID this post belongs to.
func (a *PostAggregate) CommunityID() uuid.UUID { return a.communityID }

// Title returns the current title.
func (a *PostAggregate) Title() string { return a.title }

// Content returns the current content.
func (a *PostAggregate) Content() string { return a.content }

// IsPublic returns visibility status.
func (a *PostAggregate) IsPublic() bool { return a.isPublic }

// LikesCount returns the number of likes.
func (a *PostAggregate) LikesCount() int { return a.likesCount }

// CommentCount returns the number of comments.
func (a *PostAggregate) CommentCount() int { return a.commentCount }

// CreatedAt returns creation timestamp.
func (a *PostAggregate) CreatedAt() time.Time { return a.createdAt }

// UpdatedAt returns last update timestamp.
func (a *PostAggregate) UpdatedAt() time.Time { return a.updatedAt }

// ---------- FACTORY METHOD ----------

// CreatePost creates a new PostAggregate with validation.
func CreatePost(title, content string, authorID, communityID uuid.UUID, isPublic bool) (*PostAggregate, error) {
	if err := validatePostInput(title, content, authorID, communityID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &PostAggregate{
		id:           uuid.New(),
		authorID:     authorID,
		communityID:  communityID,
		title:        strings.TrimSpace(title),
		content:      strings.TrimSpace(content),
		isPublic:     isPublic,
		likesCount:   0,
		commentCount: 0,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ---------- BUSINESS OPERATIONS (Commands) ----------

// UpdateContent modifies the post's title and content.
// Only the author may update the post.
func (a *PostAggregate) UpdateContent(newTitle, newContent string, requesterID uuid.UUID) error {
	if !a.canUserModify(requesterID) {
		return ErrUserNotPostAuthor
	}

	newTitle = strings.TrimSpace(newTitle)
	newContent = strings.TrimSpace(newContent)

	validationErrors := make(map[string]string)

	if newTitle == "" {
		validationErrors["title"] = "Post title is required"
	} else if len(newTitle) < 3 {
		validationErrors["title"] = "Post title must be at least 3 characters"
	} else if len(newTitle) > 150 {
		validationErrors["title"] = "Post title must be less than 150 characters"
	}

	if newContent == "" {
		validationErrors["content"] = "Post content is required"
	} else if len(newContent) < 10 {
		validationErrors["content"] = "Post content must be at least 10 characters"
	} else if len(newContent) > 5000 {
		validationErrors["content"] = "Post content must be less than 5000 characters"
	}

	if len(validationErrors) > 0 {
		return NewValidationFailed(validationErrors)
	}

	a.title = newTitle
	a.content = newContent
	a.updatedAt = time.Now().UTC()
	return nil
}

// ToggleVisibility changes the post's public/private status.
// Only the author may change visibility.
func (a *PostAggregate) ToggleVisibility(isPublic bool, requesterID uuid.UUID) error {
	if !a.canUserModify(requesterID) {
		return ErrUserNotPostAuthor
	}
	a.isPublic = isPublic
	a.updatedAt = time.Now().UTC()
	return nil
}

// AddLike increments the like count.
func (a *PostAggregate) AddLike() {
	a.likesCount++
	a.updatedAt = time.Now().UTC()
}

// RemoveLike decrements the like count (with underflow protection).
func (a *PostAggregate) RemoveLike() error {
	if a.likesCount <= 0 {
		return NewValidationFailed(map[string]string{
			"likes_count": "Cannot decrement likes below zero",
		})
	}
	a.likesCount--
	a.updatedAt = time.Now().UTC()
	return nil
}

// AddComment increments the comment count.
func (a *PostAggregate) AddComment() {
	a.commentCount++
	a.updatedAt = time.Now().UTC()
}

// RemoveComment decrements the comment count (with underflow protection).
func (a *PostAggregate) RemoveComment() error {
	if a.commentCount <= 0 {
		return NewValidationFailed(map[string]string{
			"comment_count": "Cannot decrement comment count below zero",
		})
	}
	a.commentCount--
	a.updatedAt = time.Now().UTC()
	return nil
}

// Delete marks the post as deleted (soft delete via state change).
// In a full implementation, you might add a `status` field or use a tombstone event.
// For now, deletion is handled externally (e.g., by removing from repo).
// This method is provided as a placeholder for future extension.
func (a *PostAggregate) Delete(requesterID uuid.UUID) error {
	if !a.canUserModify(requesterID) {
		return ErrUserNotPostAuthor
	}
	// In practice, you'd emit a PostDeletedEvent or set internal state
	// Since we don't have a 'deleted' field yet, we just validate permissions.
	return nil
}

// ---------- HELPER METHODS ----------

func (a *PostAggregate) canUserModify(userID uuid.UUID) bool {
	return a.authorID == userID
}

// ---------- READ MODEL PROJECTION (Optional View) ----------

// ToPost converts the aggregate to a read-safe Post struct.
// Useful for persistence or API responses.
func (a *PostAggregate) ToPost() *Post {
	return &Post{
		ID:           a.id,
		AuthorID:     a.authorID,
		CommunityID:  a.communityID,
		Title:        a.title,
		Content:      a.content,
		IsPublic:     a.isPublic,
		LikesCount:   a.likesCount,
		CommentCount: a.commentCount,
		CreatedAt:    a.createdAt,
		UpdatedAt:    a.updatedAt,
	}
}

// ReconstructPostAggregate rebuilds an aggregate from stored state.
// Used only by infrastructure for rehydration. Skips validation.
// Not for application code.
func ReconstructPostAggregate(
	id, authorID, communityID uuid.UUID,
	title, content string,
	isPublic bool,
	likesCount, commentCount int,
	createdAt, updatedAt time.Time,
) *PostAggregate {
	return &PostAggregate{
		id:           id,
		authorID:     authorID,
		communityID:  communityID,
		title:        title,
		content:      content,
		isPublic:     isPublic,
		likesCount:   likesCount,
		commentCount: commentCount,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}
