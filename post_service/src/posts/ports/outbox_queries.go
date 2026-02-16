// Package ports defines raw SQL queries used by infrastructure repositories.
package ports

// Outbox SQL queries for PostgreSQL

const (
	// QueryInsertOutboxEvent inserts a new outbox event.
	// Parameters: id, event_type, event_payload (JSONB), aggregate_id, aggregate_type,
	//             aggregate_version, trace_id (NULLable), metadata (JSONB),
	//             created_at, retry_count
	QueryInsertOutboxEvent = `
		INSERT INTO event_outbox (
			id, event_type, event_payload, aggregate_id, aggregate_type,
			aggregate_version, trace_id, metadata, created_at, retry_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	// QuerySelectUnpublishedOutboxEvents fetches unpublished events for processing.
	// Uses SKIP LOCKED for safe concurrent polling.
	// Parameters: limit
	QuerySelectUnpublishedOutboxEvents = `
		SELECT 
			id, event_type, event_payload, aggregate_id, aggregate_type,
			aggregate_version, trace_id, metadata, created_at, retry_count,
			published_at, processed_at, error_message
		FROM event_outbox
		WHERE processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`

	// QueryMarkOutboxAsPublished sets published_at and processed_at to NOW().
	// Parameters: outbox_id
	QueryMarkOutboxAsPublished = `
		UPDATE event_outbox
		SET published_at = NOW(), processed_at = NOW()
		WHERE id = $1
	`

	// QueryMarkOutboxAsFailed increments retry_count and records error_message.
	// Does NOT mark as processed—allows retries.
	// Parameters: error_message, outbox_id
	QueryMarkOutboxAsFailed = `
		UPDATE event_outbox
		SET retry_count = retry_count + 1,
		    error_message = $1
		WHERE id = $2
	`
)
