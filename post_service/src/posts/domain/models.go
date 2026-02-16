package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Post represents a user-created post within a community or feed.
type Post struct {
	ID           uuid.UUID `json:"id"`
	AuthorID     uuid.UUID `json:"author_id"`
	CommunityID  uuid.UUID `json:"community_id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	IsPublic     bool      `json:"is_public"`
	LikesCount   int       `json:"likes_count"`
	CommentCount int       `json:"comment_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type PostView struct {
	ID           uuid.UUID `json:"id"`
	AuthorID     uuid.UUID `json:"author_id"`
	CommunityID  uuid.UUID `json:"community_id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	IsPublic     bool      `json:"is_public"`
	LikesCount   int       `json:"likes_count"`
	CommentCount int       `json:"comment_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewPost creates a new post with validation.
func NewPost(title, content string, authorID, communityID uuid.UUID, isPublic bool) (*Post, error) {
	if err := validatePostInput(title, content, authorID, communityID); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Post{
		ID:           uuid.New(),
		AuthorID:     authorID,
		CommunityID:  communityID,
		Title:        strings.TrimSpace(title),
		Content:      strings.TrimSpace(content),
		IsPublic:     isPublic,
		LikesCount:   0,
		CommentCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// UpdatePost modifies the title and content of the post.
func (p *Post) UpdatePost(title, content string) error {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)

	validationErrors := make(map[string]string)

	if title == "" {
		validationErrors["title"] = "Post title is required"
	} else if len(title) < 3 {
		validationErrors["title"] = "Post title must be at least 3 characters"
	} else if len(title) > 150 {
		validationErrors["title"] = "Post title must be less than 150 characters"
	}

	if content == "" {
		validationErrors["content"] = "Post content is required"
	} else if len(content) < 10 {
		validationErrors["content"] = "Post content must be at least 10 characters"
	} else if len(content) > 5000 {
		validationErrors["content"] = "Post content must be less than 5000 characters"
	}

	if len(validationErrors) > 0 {
		return NewValidationFailed(validationErrors)
	}

	p.Title = title
	p.Content = content
	p.UpdatedAt = time.Now()
	return nil
}

// IncrementLikes safely increases the like count.
func (p *Post) IncrementLikes() {
	p.LikesCount++
	p.UpdatedAt = time.Now()
}

// DecrementLikes decreases the like count, ensuring it doesn't go below zero.
func (p *Post) DecrementLikes() error {
	if p.LikesCount <= 0 {
		return NewValidationFailed(map[string]string{
			"likes_count": "Cannot decrement likes below zero",
		})
	}
	p.LikesCount--
	p.UpdatedAt = time.Now()
	return nil
}

// IncrementComments increases the comment count.
func (p *Post) IncrementComments() {
	p.CommentCount++
	p.UpdatedAt = time.Now()
}

// DecrementComments decreases the comment count, with underflow protection.
func (p *Post) DecrementComments() error {
	if p.CommentCount <= 0 {
		return NewValidationFailed(map[string]string{
			"comment_count": "Cannot decrement comment count below zero",
		})
	}
	p.CommentCount--
	p.UpdatedAt = time.Now()
	return nil
}

// validatePostInput ensures all fields meet domain rules.
func validatePostInput(title, content string, authorID, communityID uuid.UUID) error {
	validationErrors := make(map[string]string)

	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)

	if title == "" {
		validationErrors["title"] = "Post title is required"
	} else if len(title) < 3 {
		validationErrors["title"] = "Post title must be at least 3 characters"
	} else if len(title) > 150 {
		validationErrors["title"] = "Post title must be less than 150 characters"
	}

	if content == "" {
		validationErrors["content"] = "Post content is required"
	} else if len(content) < 10 {
		validationErrors["content"] = "Post content must be at least 10 characters"
	} else if len(content) > 5000 {
		validationErrors["content"] = "Post content must be less than 5000 characters"
	}

	// Validate UUIDs: zero value check is sufficient for google/uuid
	if authorID == uuid.Nil {
		validationErrors["author_id"] = "Valid author ID is required"
	}

	if communityID == uuid.Nil {
		validationErrors["community_id"] = "Valid community ID is required"
	}

	if len(validationErrors) > 0 {
		return NewValidationFailed(validationErrors)
	}
	return nil
}
