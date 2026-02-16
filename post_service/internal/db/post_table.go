// Package db handles PostgreSQL schema for posts.
package db

import (
	"fmt"
	"log"
)

// CreatePostsTable creates the 'posts' table and necessary indexes in PostgreSQL.
func CreatePostsTable() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Enable pgcrypto for gen_random_uuid()
	if _, err := DB.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`); err != nil {
		return fmt.Errorf("failed to enable pgcrypto extension: %w", err)
	}

	// Create the posts table
	createPostsTable := `
CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL,
    community_id UUID NOT NULL,
    title TEXT NOT NULL CHECK (char_length(title) >= 3 AND char_length(title) <= 150),
    content TEXT NOT NULL CHECK (char_length(content) >= 10 AND char_length(content) <= 5000),
    is_public BOOLEAN NOT NULL DEFAULT true,
    likes_count INTEGER NOT NULL DEFAULT 0 CHECK (likes_count >= 0),
    comment_count INTEGER NOT NULL DEFAULT 0 CHECK (comment_count >= 0),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);`

	if _, err := DB.Exec(createPostsTable); err != nil {
		return fmt.Errorf("failed to create posts table: %w", err)
	}

	// Create performance indexes
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts (author_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_community_id ON posts (community_id)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_is_public ON posts (is_public) WHERE is_public = true`,
		`CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts (created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_community_created ON posts (community_id, created_at DESC)`,
	}

	for _, query := range indexes {
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Optional: full-text search index (non-blocking on failure)
	ftsQuery := `
	CREATE INDEX IF NOT EXISTS idx_posts_search ON posts
	USING gin(to_tsvector('english', title || ' ' || content));
	`
	if _, err := DB.Exec(ftsQuery); err != nil {
		log.Printf("Note: full-text search index not created (may lack extension): %v", err)
	}

	log.Println("Posts table and indexes created successfully")
	return nil
}
