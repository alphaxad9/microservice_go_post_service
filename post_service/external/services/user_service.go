package services

import (
	"github.com/alphaxad9/my-go-backend/post_service/external"

	"github.com/google/uuid"
)

// UserQueryService defines the contract for fetching user data.
type UserQueryService interface {
	GetUserByID(userID uuid.UUID) (*external.UserView, error)
}
