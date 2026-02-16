package models

import (
	"time"

	"github.com/google/uuid"
)

// PostModel represents the PostgreSQL row structure for a post.
// It maps directly to the 'posts' table and is used only for data transfer.
type PostModel struct {
	ID           uuid.UUID `db:"id"`
	AuthorID     uuid.UUID `db:"author_id"`
	CommunityID  uuid.UUID `db:"community_id"`
	Title        string    `db:"title"`
	Content      string    `db:"content"`
	IsPublic     bool      `db:"is_public"`
	LikesCount   int       `db:"likes_count"`
	CommentCount int       `db:"comment_count"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
