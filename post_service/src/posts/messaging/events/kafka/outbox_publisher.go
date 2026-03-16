// github.com/alphaxad9/my-go-backend/post_service/src/posts/messaging/events/kafka/outbox_publisher.go
package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/alphaxad9/my-go-backend/post_service/internal/db"
	outboxrepos "github.com/alphaxad9/my-go-backend/post_service/src/posts/infra/repositories/outbox"
)

type OutboxPublisher struct {
	repo     *outboxrepos.OutboxRepositoryImpl
	producer Producer
	logger   *slog.Logger
	cfg      struct {
		PollInterval time.Duration
		BatchSize    int
		MaxRetries   int
	}
}

func NewOutboxPublisher(
	repo *outboxrepos.OutboxRepositoryImpl,
	producer Producer,
	pollInterval time.Duration,
	batchSize int,
	maxRetries int,
) *OutboxPublisher {
	return &OutboxPublisher{
		repo:     repo,
		producer: producer,
		logger:   slog.With("component", "outbox_publisher"),
		cfg: struct {
			PollInterval time.Duration
			BatchSize    int
			MaxRetries   int
		}{
			PollInterval: pollInterval,
			BatchSize:    batchSize,
			MaxRetries:   maxRetries,
		},
	}
}

func (op *OutboxPublisher) Start(ctx context.Context) {
	op.logger.Info("Starting outbox publisher", "interval_sec", op.cfg.PollInterval.Seconds())
	ticker := time.NewTicker(op.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			op.logger.Info("Shutting down outbox publisher")
			return
		case <-ticker.C:
			if err := op.publishBatch(ctx); err != nil {
				op.logger.Error("Failed to publish batch", "error", err)
			}
		}
	}
}

func (op *OutboxPublisher) publishBatch(ctx context.Context) error {
	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Track if we've committed successfully
	committed := false
	defer func() {
		if !committed {
			if rbErr := tx.Rollback(); rbErr != nil {
				// Log rollback error but don't overshadow any existing error
				op.logger.Error("Failed to rollback transaction", "rollback_error", rbErr)
			}
		}
	}()

	repoTx := outboxrepos.NewOutboxRepositoryTx(tx)

	events, err := repoTx.GetUnpublishedEvents(ctx, op.cfg.BatchSize)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		committed = true // Nothing to commit, but we're done
		return nil
	}

	op.logger.Debug("Processing outbox batch", "count", len(events))

	for _, event := range events {
		payloadBytes, err := json.Marshal(event.EventPayload)
		if err != nil {
			op.logger.Error("Failed to serialize event payload", "event_id", event.ID, "error", err)
			continue
		}

		headers := map[string]string{
			"event_type": event.EventType,
		}

		err = op.producer.Produce(ctx, []byte(event.ID.String()), payloadBytes, headers)
		if err != nil {
			op.logger.Error("Failed to produce Kafka message", "event_id", event.ID, "error", err)
			if err := repoTx.MarkAsFailed(ctx, event.ID, err.Error()); err != nil {
				op.logger.Error("Failed to mark event as failed", "event_id", event.ID, "error", err)
			}
			continue
		}

		if err := repoTx.MarkAsPublished(ctx, event.ID); err != nil {
			op.logger.Error("Failed to mark event as published", "event_id", event.ID, "error", err)
			continue
		}

		op.logger.Info("Published event to Kafka",
			"event_id", event.ID,
			"event_type", event.EventType,
			"aggregate_id", event.AggregateID,
		)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	committed = true
	return nil
}
