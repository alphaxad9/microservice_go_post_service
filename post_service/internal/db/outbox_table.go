// Package db handles PostgreSQL schema management.
package db

import (
	"fmt"
	"log"
)

// CreateOutboxTable creates the 'event_outbox' table and necessary indexes for the transactional outbox pattern.
func CreateOutboxTable() error {
	if DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Ensure pgcrypto is available for gen_random_uuid()
	if _, err := DB.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`); err != nil {
		return fmt.Errorf("failed to enable pgcrypto extension: %w", err)
	}

	// Create the event_outbox table
	createOutboxTable := `
CREATE TABLE IF NOT EXISTS event_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Core event data
    event_type TEXT NOT NULL,
    event_payload JSONB NOT NULL,
    
    -- DDD / Tracing
    aggregate_id UUID NOT NULL,
    aggregate_type TEXT NOT NULL,
    aggregate_version INTEGER NOT NULL CHECK (aggregate_version >= 0),
    
    -- Observability
    trace_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}',
    
    -- Timing
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    
    -- Resilience
    retry_count SMALLINT NOT NULL DEFAULT 0 CHECK (retry_count >= 0),
    error_message TEXT
);`

	if _, err := DB.Exec(createOutboxTable); err != nil {
		return fmt.Errorf("failed to create event_outbox table: %w", err)
	}

	// Add unique constraint: (aggregate_id, aggregate_version)
	_, err := DB.Exec(`
		DO $$ 
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'unique_aggregate_version'
			) THEN
				ALTER TABLE event_outbox 
				ADD CONSTRAINT unique_aggregate_version 
				UNIQUE (aggregate_id, aggregate_version);
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add unique_aggregate_version constraint: %w", err)
	}

	// Create essential indexes for polling and querying
	indexes := []string{
		// Index for polling unpublished events (core for outbox processor)
		`CREATE INDEX IF NOT EXISTS idx_outbox_unpublished ON event_outbox (created_at) 
		 WHERE processed_at IS NULL`,

		// Composite index for efficient range polling
		`CREATE INDEX IF NOT EXISTS idx_outbox_polling ON event_outbox (processed_at, created_at) 
		 WHERE processed_at IS NULL`,

		// Lookup by aggregate
		`CREATE INDEX IF NOT EXISTS idx_outbox_aggregate ON event_outbox (aggregate_id, aggregate_type)`,

		// Traceability
		`CREATE INDEX IF NOT EXISTS idx_outbox_trace_id ON event_outbox (trace_id) WHERE trace_id IS NOT NULL`,

		// Event type filtering
		`CREATE INDEX IF NOT EXISTS idx_outbox_event_type ON event_outbox (event_type)`,

		// Retry monitoring
		`CREATE INDEX IF NOT EXISTS idx_outbox_retry_count ON event_outbox (retry_count) 
		 WHERE processed_at IS NULL`,
	}

	for _, query := range indexes {
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf("failed to create outbox index: %w", err)
		}
	}

	log.Println("Event outbox table and indexes created successfully")
	return nil
}