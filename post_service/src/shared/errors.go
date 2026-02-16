// shared/errors.go
package shared

import (
	"fmt"
	"net/http"
)

type ErrorCode string

type AppError struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	HTTPStatus int         `json:"-"`
	Cause      error       `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

const (
	ErrorCodeInternal ErrorCode = "INTERNAL_SERVER_ERROR"
)

// Exported function
func NewInternalServerError(err error) *AppError {
	return &AppError{
		Code:       ErrorCodeInternal,
		Message:    "An internal server error occurred",
		Details:    map[string]interface{}{"cause": err.Error()},
		HTTPStatus: http.StatusInternalServerError,
		Cause:      err,
	}
}
