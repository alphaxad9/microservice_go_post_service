// Package outbox defines domain models and errors for the outbox pattern.
package outbox

import (
	"fmt"
	"net/http"

	"my-go-backend/post_service/src/shared"
	"github.com/google/uuid"
)

// Outbox error codes
const (
	ErrorCodeOutboxSaveFailed         shared.ErrorCode = "OUTBOX_SAVE_FAILED"
	ErrorCodeOutboxNotFound           shared.ErrorCode = "OUTBOX_NOT_FOUND"
	ErrorCodeOutboxPublishFailed      shared.ErrorCode = "OUTBOX_PUBLISH_FAILED"
	ErrorCodeOutboxConcurrencyError   shared.ErrorCode = "OUTBOX_CONCURRENCY_ERROR"
	ErrorCodeOutboxMaxRetriesExceeded shared.ErrorCode = "OUTBOX_MAX_RETRIES_EXCEEDED"
)

// Outbox error constructors

func NewOutboxSaveError(eventType string, aggregateID uuid.UUID, reason string) *shared.AppError {
	return &shared.AppError{
		Code:       ErrorCodeOutboxSaveFailed,
		Message:    fmt.Sprintf("Failed to save outbox event '%s' for %s", eventType, aggregateID),
		Details: map[string]interface{}{
			"event_type":   eventType,
			"aggregate_id": aggregateID.String(),
			"reason":       reason,
		},
		HTTPStatus: http.StatusInternalServerError,
	}
}

func NewOutboxNotFoundError(outboxID uuid.UUID) *shared.AppError {
	return &shared.AppError{
		Code:       ErrorCodeOutboxNotFound,
		Message:    fmt.Sprintf("Outbox event with ID %s not found", outboxID),
		Details: map[string]interface{}{
			"outbox_id": outboxID.String(),
		},
		HTTPStatus: http.StatusNotFound,
	}
}

func NewOutboxPublishError(outboxID uuid.UUID, eventType string, brokerError string) *shared.AppError {
	return &shared.AppError{
		Code:       ErrorCodeOutboxPublishFailed,
		Message:    fmt.Sprintf("Failed to publish outbox event %s ('%s')", outboxID, eventType),
		Details: map[string]interface{}{
			"outbox_id":    outboxID.String(),
			"event_type":   eventType,
			"broker_error": brokerError,
		},
		HTTPStatus: http.StatusInternalServerError,
	}
}

func NewOutboxConcurrencyError(outboxID uuid.UUID, message string) *shared.AppError {
	if message == "" {
		message = "Concurrency conflict detected"
	}
	return &shared.AppError{
		Code:       ErrorCodeOutboxConcurrencyError,
		Message:    fmt.Sprintf("%s for outbox ID %s", message, outboxID),
		Details: map[string]interface{}{
			"outbox_id": outboxID.String(),
		},
		HTTPStatus: http.StatusConflict,
	}
}

func NewOutboxMaxRetriesExceededError(outboxID uuid.UUID, maxRetries int, lastError string) *shared.AppError {
	return &shared.AppError{
		Code:       ErrorCodeOutboxMaxRetriesExceeded,
		Message:    fmt.Sprintf("Outbox event %s exceeded max retries (%d)", outboxID, maxRetries),
		Details: map[string]interface{}{
			"outbox_id":   outboxID.String(),
			"max_retries": maxRetries,
			"last_error":  lastError,
		},
		HTTPStatus: http.StatusInternalServerError,
	}
}