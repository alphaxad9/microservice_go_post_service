// Package repositories contains PostgreSQL repository implementations.
// github.com/alphaxad9/my-go-backend/post_service/src/posts/infra/repositories/db_test.go
package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const (
	testPostsSchema = `
CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY,
    author_id UUID NOT NULL,
    community_id UUID NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT true,
    likes_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);`

	testOutboxSchema = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS event_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    event_payload JSONB NOT NULL,
    aggregate_id UUID NOT NULL,
    aggregate_type TEXT NOT NULL,
    aggregate_version INTEGER NOT NULL CHECK (aggregate_version >= 0),
    trace_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    retry_count SMALLINT NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
    error_message TEXT
);`
)

var (
	testDB *sql.DB
)

func TestMain(m *testing.M) {
	// Connect to externally managed test DB
	dsn := "host=localhost port=5433 user=testuser password=testpass dbname=testdb sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("failed to connect to test DB: %w", err))
	}

	testDB = db

	// Create schemas
	ctx := context.Background()
	if _, err := testDB.ExecContext(ctx, testPostsSchema); err != nil {
		panic(fmt.Errorf("failed to create posts schema: %w", err))
	}
	if _, err := testDB.ExecContext(ctx, testOutboxSchema); err != nil {
		panic(fmt.Errorf("failed to create outbox schema: %w", err))
	}

	// Reset before all tests
	ResetPosts(nil) // pass nil since we don't have *testing.T here
	ResetOutbox(nil)

	code := m.Run()
	os.Exit(code)
}

// Allow nil *testing.T for TestMain usage
func ResetPosts(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "TRUNCATE TABLE posts RESTART IDENTITY CASCADE")
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}
}

func ResetOutbox(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "TRUNCATE TABLE event_outbox RESTART IDENTITY CASCADE")
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}
}

// GetTestDB returns the shared test database.
func GetTestDB() *sql.DB {
	return testDB
}
