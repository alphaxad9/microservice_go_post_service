// Package repositories tests the PostgreSQL implementation of the outbox repository.
package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"my-go-backend/post_service/src/posts/domain/outbox"
	"my-go-backend/post_service/src/shared"
)

const (
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

// Global test database connection
var (
	testDB *sql.DB
)

// TestMain connects to the externally managed test database (e.g., docker run postgres on port 5433).
func TestMain(m *testing.M) {
	dsn := "host=localhost port=5433 user=testuser password=testpass dbname=testdb sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("failed to connect to test DB: %w", err))
	}

	testDB = db

	// Create outbox schema
	ctx := context.Background()
	if _, err := testDB.ExecContext(ctx, testOutboxSchema); err != nil {
		panic(fmt.Errorf("failed to create outbox schema: %w", err))
	}

	// Reset outbox table before running tests
	resetOutbox(nil)

	code := m.Run()

	os.Exit(code)
}

// resetOutbox truncates the event_outbox table for test isolation.
// Accepts nil *testing.T for use in TestMain.
func resetOutbox(t *testing.T) {
	ctx := context.Background()
	_, err := testDB.ExecContext(ctx, "TRUNCATE TABLE event_outbox RESTART IDENTITY CASCADE")
	if t != nil {
		require.NoError(t, err)
	} else if err != nil {
		panic(err)
	}
}

// getTestDB returns the shared test database connection.
func getTestDB() *sql.DB {
	return testDB
}

// --- Original tests below (unchanged logic, just faster) ---

func TestOutboxRepository_MarkAsPublished(t *testing.T) {
	resetOutbox(t)
	repo := NewOutboxRepository(getTestDB())
	ctx := context.Background()

	event, err := outbox.NewOutboxEvent("TestEvent", map[string]string{"key": "value"}, uuid.New(), "Test", nil, nil)
	require.NoError(t, err)
	err = repo.Save(ctx, event)
	require.NoError(t, err)

	err = repo.MarkAsPublished(ctx, event.ID)
	require.NoError(t, err)

	events, err := repo.GetUnpublishedEvents(ctx, 10)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestOutboxRepository_MarkAsPublished_NotFound(t *testing.T) {
	resetOutbox(t)
	repo := NewOutboxRepository(getTestDB())
	ctx := context.Background()

	err := repo.MarkAsPublished(ctx, uuid.New())
	require.Error(t, err)
	assert.IsType(t, &shared.AppError{}, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestOutboxRepository_MarkAsFailed(t *testing.T) {
	resetOutbox(t)
	repo := NewOutboxRepository(getTestDB())
	ctx := context.Background()

	event, err := outbox.NewOutboxEvent("FailingEvent", map[string]string{"test": "data"}, uuid.New(), "Test", nil, nil)
	require.NoError(t, err)
	err = repo.Save(ctx, event)
	require.NoError(t, err)

	err = repo.MarkAsFailed(ctx, event.ID, "connection timeout")
	require.NoError(t, err)

	events, err := repo.GetUnpublishedEvents(ctx, 10)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, 1, events[0].RetryCount)
}

func TestOutboxRepository_MarkAsFailed_NotFound(t *testing.T) {
	resetOutbox(t)
	repo := NewOutboxRepository(getTestDB())
	ctx := context.Background()

	err := repo.MarkAsFailed(ctx, uuid.New(), "simulated error")
	require.Error(t, err)
	assert.IsType(t, &shared.AppError{}, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestOutboxRepository_ConcurrentPolling(t *testing.T) {
	resetOutbox(t)
	ctx := context.Background()

	repo := NewOutboxRepository(getTestDB())
	for i := 0; i < 3; i++ {
		event, err := outbox.NewOutboxEvent(
			"ConcurrentEvent",
			map[string]int{"id": i},
			uuid.New(),
			"Test",
			nil,
			nil,
		)
		require.NoError(t, err)
		err = repo.Save(ctx, event)
		require.NoError(t, err)
	}

	// Simulate two concurrent consumers with separate transactions
	tx1, err := testDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx1.Rollback()

	tx2, err := testDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx2.Rollback()

	repo1 := NewOutboxRepositoryTx(tx1)
	events1, err := repo1.GetUnpublishedEvents(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, events1, 2)

	repo2 := NewOutboxRepositoryTx(tx2)
	events2, err := repo2.GetUnpublishedEvents(ctx, 2)
	require.NoError(t, err)
	assert.Len(t, events2, 1)

	idMap := make(map[uuid.UUID]bool)
	for _, e := range events1 {
		idMap[e.ID] = true
	}
	for _, e := range events2 {
		assert.NotContains(t, idMap, e.ID, "event was processed twice")
	}

	require.NoError(t, tx1.Commit())
	require.NoError(t, tx2.Commit())
}

func TestOutboxRepository_Save(t *testing.T) {
	resetOutbox(t)
	repo := NewOutboxRepository(getTestDB())
	ctx := context.Background()

	payload := map[string]interface{}{
		"post_id":   "123e4567-e89b-12d3-a456-426614174000",
		"title":     "Hello World",
		"author_id": "987e4567-e89b-12d3-a456-426614174001",
	}
	metadata := map[string]interface{}{
		"source":  "post_service",
		"version": "1.0",
	}
	traceID := uuid.New()

	event, err := outbox.NewOutboxEvent(
		"PostCreated",
		payload,
		uuid.New(),
		"Post",
		&traceID,
		metadata,
	)
	require.NoError(t, err)

	err = repo.Save(ctx, event)
	require.NoError(t, err)

	events, err := repo.GetUnpublishedEvents(ctx, 10)
	require.NoError(t, err)
	require.Len(t, events, 1)

	fetched := events[0]
	assert.Equal(t, event.ID, fetched.ID)
	assert.Equal(t, event.EventType, fetched.EventType)
	assert.Equal(t, event.AggregateID, fetched.AggregateID)
	assert.Equal(t, event.AggregateType, fetched.AggregateType)
	assert.Equal(t, event.TraceID, fetched.TraceID)
	assert.Equal(t, event.RetryCount, fetched.RetryCount)
	assert.WithinDuration(t, event.CreatedAt, fetched.CreatedAt, time.Second)

	var fetchedPayload map[string]interface{}
	err = json.Unmarshal(fetched.EventPayload, &fetchedPayload)
	require.NoError(t, err)
	assert.Equal(t, payload["title"], fetchedPayload["title"])

	var fetchedMetadata map[string]interface{}
	err = json.Unmarshal(fetched.Metadata, &fetchedMetadata)
	require.NoError(t, err)
	assert.Equal(t, metadata["source"], fetchedMetadata["source"])
}

func TestOutboxRepository_GetUnpublishedEvents(t *testing.T) {
	resetOutbox(t)
	repo := NewOutboxRepository(getTestDB())
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		event, err := outbox.NewOutboxEvent(
			fmt.Sprintf("TestEvent%d", i),
			map[string]string{"data": fmt.Sprintf("value%d", i)},
			uuid.New(),
			"TestAggregate",
			nil,
			nil,
		)
		require.NoError(t, err)
		err = repo.Save(ctx, event)
		require.NoError(t, err)
	}

	events, err := repo.GetUnpublishedEvents(ctx, 1)
	require.NoError(t, err)
	require.Len(t, events, 1)
	err = repo.MarkAsPublished(ctx, events[0].ID)
	require.NoError(t, err)

	unpublished, err := repo.GetUnpublishedEvents(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, unpublished, 1)
	assert.NotEqual(t, events[0].ID, unpublished[0].ID)
}
