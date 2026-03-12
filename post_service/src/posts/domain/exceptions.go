package domain

import (
	"net/http"

	"github.com/alphaxad9/my-go-backend/post_service/src/shared"
)

// Post domain error codes
const (
	ErrorCodePostNotFound           shared.ErrorCode = "POST_NOT_FOUND"
	ErrorCodePostAlreadyExists      shared.ErrorCode = "POST_ALREADY_EXISTS"
	ErrorCodeInvalidPostData        shared.ErrorCode = "INVALID_POST_DATA"
	ErrorCodeUnauthorizedPostAccess shared.ErrorCode = "UNAUTHORIZED_POST_ACCESS"
	ErrorCodePostTitleTooShort      shared.ErrorCode = "POST_TITLE_TOO_SHORT"
	ErrorCodePostTitleTooLong       shared.ErrorCode = "POST_TITLE_TOO_LONG"
	ErrorCodePostContentTooShort    shared.ErrorCode = "POST_CONTENT_TOO_SHORT"
	ErrorCodePostContentTooLong     shared.ErrorCode = "POST_CONTENT_TOO_LONG"
	ErrorCodeUserNotPostAuthor      shared.ErrorCode = "USER_NOT_POST_AUTHOR"
	ErrorCodeValidationFailed       shared.ErrorCode = "VALIDATION_FAILED"
	ErrorCodeDomainError            shared.ErrorCode = "DOMAIN_ERROR"
)

// Post error variables that implement the error interface
var (
	ErrPostNotFound           = NewPostNotFoundError("")
	ErrPostAlreadyExists      = NewPostAlreadyExistsError("")
	ErrInvalidPostData        = NewInvalidPostDataError("")
	ErrUnauthorizedPostAccess = NewUnauthorizedPostAccessError("")
	ErrUserNotPostAuthor      = NewUserNotPostAuthorError("", "")
)

// Post error constructors

func NewPostNotFoundError(postID string) *shared.AppError {
	details := map[string]interface{}{}
	if postID != "" {
		details["post_id"] = postID
	}
	return &shared.AppError{
		Code:       ErrorCodePostNotFound,
		Message:    "Post not found",
		Details:    details,
		HTTPStatus: http.StatusNotFound,
	}
}

func NewPostAlreadyExistsError(postID string) *shared.AppError {
	details := map[string]interface{}{}
	if postID != "" {
		details["post_id"] = postID
	}
	return &shared.AppError{
		Code:       ErrorCodePostAlreadyExists,
		Message:    "Post already exists",
		Details:    details,
		HTTPStatus: http.StatusConflict,
	}
}

func NewInvalidPostDataError(reason string) *shared.AppError {
	details := map[string]interface{}{}
	if reason != "" {
		details["reason"] = reason
	}
	return &shared.AppError{
		Code:       ErrorCodeInvalidPostData,
		Message:    "Invalid post data",
		Details:    details,
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewUnauthorizedPostAccessError(resource string) *shared.AppError {
	details := map[string]interface{}{}
	if resource != "" {
		details["resource"] = resource
	}
	return &shared.AppError{
		Code:       ErrorCodeUnauthorizedPostAccess,
		Message:    "Unauthorized access to post",
		Details:    details,
		HTTPStatus: http.StatusForbidden,
	}
}

func NewUserNotPostAuthorError(postID, userID string) *shared.AppError {
	details := map[string]interface{}{}
	if postID != "" {
		details["post_id"] = postID
	}
	if userID != "" {
		details["user_id"] = userID
	}
	return &shared.AppError{
		Code:       ErrorCodeUserNotPostAuthor,
		Message:    "User is not the author of this post",
		Details:    details,
		HTTPStatus: http.StatusForbidden,
	}
}

func NewPostTitleTooShortError() *shared.AppError {
	return &shared.AppError{
		Code:       ErrorCodePostTitleTooShort,
		Message:    "Post title must be at least 3 characters",
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewPostTitleTooLongError() *shared.AppError {
	return &shared.AppError{
		Code:       ErrorCodePostTitleTooLong,
		Message:    "Post title must be less than 150 characters",
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewPostContentTooShortError() *shared.AppError {
	return &shared.AppError{
		Code:       ErrorCodePostContentTooShort,
		Message:    "Post content must be at least 10 characters",
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewPostContentTooLongError() *shared.AppError {
	return &shared.AppError{
		Code:       ErrorCodePostContentTooLong,
		Message:    "Post content must be less than 5000 characters",
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewValidationFailed(details interface{}) *shared.AppError {
	return &shared.AppError{
		Code:       ErrorCodeValidationFailed,
		Message:    "Validation failed",
		Details:    details,
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewDomainError(message string, err error) *shared.AppError {
	details := map[string]interface{}{}
	if err != nil {
		details["cause"] = err.Error()
	}
	return &shared.AppError{
		Code:       ErrorCodeDomainError,
		Message:    message,
		Details:    details,
		HTTPStatus: http.StatusInternalServerError,
	}
}
