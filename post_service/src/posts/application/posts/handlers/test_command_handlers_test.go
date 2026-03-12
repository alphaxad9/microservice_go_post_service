// Package handlers tests the PostCommandHandler implementation.
package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"
	shared "github.com/alphaxad9/my-go-backend/post_service/src/shared"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// -----------------------
// Mocks
// -----------------------

type MockPostCommandService struct {
	mock.Mock
}

func (m *MockPostCommandService) CreatePost(
	ctx context.Context,
	title, content string,
	authorID, communityID uuid.UUID,
	isPublic bool,
) (*domain.PostAggregate, error) {
	args := m.Called(ctx, title, content, authorID, communityID, isPublic)
	if agg, ok := args.Get(0).(*domain.PostAggregate); ok {
		return agg, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostCommandService) UpdatePost(
	ctx context.Context,
	postID uuid.UUID,
	newTitle, newContent string,
	requesterID uuid.UUID,
) error {
	args := m.Called(ctx, postID, newTitle, newContent, requesterID)
	return args.Error(0)
}

func (m *MockPostCommandService) TogglePostVisibility(
	ctx context.Context,
	postID uuid.UUID,
	isPublic bool,
	requesterID uuid.UUID,
) error {
	args := m.Called(ctx, postID, isPublic, requesterID)
	return args.Error(0)
}

func (m *MockPostCommandService) LikePost(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostCommandService) UnlikePost(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostCommandService) AddCommentToPost(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostCommandService) RemoveCommentFromPost(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostCommandService) DeletePost(
	ctx context.Context,
	postID uuid.UUID,
	requesterID uuid.UUID,
) error {
	args := m.Called(ctx, postID, requesterID)
	return args.Error(0)
}

// -----------------------
// Test Constants
// -----------------------

// -----------------------
// Helper
// -----------------------

func newTestPostAggregate(authorID, communityID uuid.UUID) *domain.PostAggregate {
	id := uuid.New()
	return domain.ReconstructPostAggregate(
		id, authorID, communityID,
		"Test Title", "Test Content", true,
		0, 0,
		testTime, testTime,
	)
}

// -----------------------
// Tests
// -----------------------

func TestPostCommandHandler_TogglePostVisibility(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		isPublic        bool
		setupMocks      func(svc *MockPostCommandService, postID, requesterID uuid.UUID, isPublic bool)
		expectErrorCode shared.ErrorCode
	}{
		{
			name:     "success",
			isPublic: false,
			setupMocks: func(svc *MockPostCommandService, postID, requesterID uuid.UUID, isPublic bool) {
				svc.On("TogglePostVisibility", ctx, postID, isPublic, requesterID).Return(nil)
			},
			expectErrorCode: "",
		},
		{
			name:     "unauthorized",
			isPublic: true,
			setupMocks: func(svc *MockPostCommandService, postID, requesterID uuid.UUID, isPublic bool) {
				err := domain.NewUserNotPostAuthorError(postID.String(), requesterID.String())
				svc.On("TogglePostVisibility", ctx, postID, isPublic, requesterID).Return(err)
			},
			expectErrorCode: domain.ErrorCodeUserNotPostAuthor,
		},
		{
			name:     "post not found",
			isPublic: true,
			setupMocks: func(svc *MockPostCommandService, postID, requesterID uuid.UUID, isPublic bool) {
				svc.On("TogglePostVisibility", ctx, postID, isPublic, requesterID).
					Return(domain.NewPostNotFoundError(postID.String()))
			},
			expectErrorCode: domain.ErrorCodePostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postID := uuid.New()
			requesterID := uuid.New()
			mockSvc := new(MockPostCommandService)
			tt.setupMocks(mockSvc, postID, requesterID, tt.isPublic)

			handler := NewPostCommandHandler(mockSvc)
			err := handler.TogglePostVisibility(ctx, postID, tt.isPublic, requesterID)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
			}
		})
	}
}

func TestPostCommandHandler_LikePost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		setupMocks      func(svc *MockPostCommandService, postID uuid.UUID)
		expectErrorCode shared.ErrorCode
	}{
		{
			name: "success",
			setupMocks: func(svc *MockPostCommandService, postID uuid.UUID) {
				svc.On("LikePost", ctx, postID).Return(nil)
			},
			expectErrorCode: "",
		},
		{
			name: "post not found",
			setupMocks: func(svc *MockPostCommandService, postID uuid.UUID) {
				svc.On("LikePost", ctx, postID).Return(domain.NewPostNotFoundError(postID.String()))
			},
			expectErrorCode: domain.ErrorCodePostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postID := uuid.New()
			mockSvc := new(MockPostCommandService)
			tt.setupMocks(mockSvc, postID)

			handler := NewPostCommandHandler(mockSvc)
			err := handler.LikePost(ctx, postID)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
			}
		})
	}
}

func TestPostCommandHandler_UnlikePost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		setupMocks      func(svc *MockPostCommandService, postID uuid.UUID)
		expectErrorCode shared.ErrorCode
	}{
		{
			name: "success",
			setupMocks: func(svc *MockPostCommandService, postID uuid.UUID) {
				svc.On("UnlikePost", ctx, postID).Return(nil)
			},
			expectErrorCode: "",
		},
		{
			name: "underflow error",
			setupMocks: func(svc *MockPostCommandService, postID uuid.UUID) {
				err := domain.NewValidationFailed(map[string]string{"likes_count": "Cannot decrement likes below zero"})
				svc.On("UnlikePost", ctx, postID).Return(err)
			},
			expectErrorCode: domain.ErrorCodeValidationFailed,
		},
		{
			name: "post not found",
			setupMocks: func(svc *MockPostCommandService, postID uuid.UUID) {
				svc.On("UnlikePost", ctx, postID).Return(domain.NewPostNotFoundError(postID.String()))
			},
			expectErrorCode: domain.ErrorCodePostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postID := uuid.New()
			mockSvc := new(MockPostCommandService)
			tt.setupMocks(mockSvc, postID)

			handler := NewPostCommandHandler(mockSvc)
			err := handler.UnlikePost(ctx, postID)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
			}
		})
	}
}

func TestPostCommandHandler_AddCommentToPost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		setupMocks      func(svc *MockPostCommandService, postID uuid.UUID)
		expectErrorCode shared.ErrorCode
	}{
		{
			name: "success",
			setupMocks: func(svc *MockPostCommandService, postID uuid.UUID) {
				svc.On("AddCommentToPost", ctx, postID).Return(nil)
			},
			expectErrorCode: "",
		},
		{
			name: "post not found",
			setupMocks: func(svc *MockPostCommandService, postID uuid.UUID) {
				svc.On("AddCommentToPost", ctx, postID).Return(domain.NewPostNotFoundError(postID.String()))
			},
			expectErrorCode: domain.ErrorCodePostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postID := uuid.New()
			mockSvc := new(MockPostCommandService)
			tt.setupMocks(mockSvc, postID)

			handler := NewPostCommandHandler(mockSvc)
			err := handler.AddCommentToPost(ctx, postID)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
			}
		})
	}
}

func TestPostCommandHandler_RemoveCommentFromPost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		setupMocks      func(svc *MockPostCommandService, postID uuid.UUID)
		expectErrorCode shared.ErrorCode
	}{
		{
			name: "success",
			setupMocks: func(svc *MockPostCommandService, postID uuid.UUID) {
				svc.On("RemoveCommentFromPost", ctx, postID).Return(nil)
			},
			expectErrorCode: "",
		},
		{
			name: "underflow error",
			setupMocks: func(svc *MockPostCommandService, postID uuid.UUID) {
				err := domain.NewValidationFailed(map[string]string{"comment_count": "Cannot decrement comment count below zero"})
				svc.On("RemoveCommentFromPost", ctx, postID).Return(err)
			},
			expectErrorCode: domain.ErrorCodeValidationFailed,
		},
		{
			name: "post not found",
			setupMocks: func(svc *MockPostCommandService, postID uuid.UUID) {
				svc.On("RemoveCommentFromPost", ctx, postID).Return(domain.NewPostNotFoundError(postID.String()))
			},
			expectErrorCode: domain.ErrorCodePostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postID := uuid.New()
			mockSvc := new(MockPostCommandService)
			tt.setupMocks(mockSvc, postID)

			handler := NewPostCommandHandler(mockSvc)
			err := handler.RemoveCommentFromPost(ctx, postID)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
			}
		})
	}
}

func TestPostCommandHandler_DeletePost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		setupMocks      func(svc *MockPostCommandService, postID, requesterID uuid.UUID)
		expectErrorCode shared.ErrorCode
	}{
		{
			name: "success",
			setupMocks: func(svc *MockPostCommandService, postID, requesterID uuid.UUID) {
				svc.On("DeletePost", ctx, postID, requesterID).Return(nil)
			},
			expectErrorCode: "",
		},
		{
			name: "unauthorized",
			setupMocks: func(svc *MockPostCommandService, postID, requesterID uuid.UUID) {
				err := domain.NewUserNotPostAuthorError(postID.String(), requesterID.String())
				svc.On("DeletePost", ctx, postID, requesterID).Return(err)
			},
			expectErrorCode: domain.ErrorCodeUserNotPostAuthor,
		},
		{
			name: "post not found",
			setupMocks: func(svc *MockPostCommandService, postID, requesterID uuid.UUID) {
				svc.On("DeletePost", ctx, postID, requesterID).Return(domain.NewPostNotFoundError(postID.String()))
			},
			expectErrorCode: domain.ErrorCodePostNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postID := uuid.New()
			requesterID := uuid.New()
			mockSvc := new(MockPostCommandService)
			tt.setupMocks(mockSvc, postID, requesterID)

			handler := NewPostCommandHandler(mockSvc)
			err := handler.DeletePost(ctx, postID, requesterID)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
			}
		})
	}
}
func TestPostCommandHandler_CreatePost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		title           string
		content         string
		isPublic        bool
		setupMocks      func(svc *MockPostCommandService, authorID, communityID uuid.UUID, title, content string, isPublic bool)
		expectErrorCode shared.ErrorCode
	}{
		{
			name:     "success",
			title:    "Valid Title",
			content:  "Valid content here.",
			isPublic: true,
			setupMocks: func(svc *MockPostCommandService, authorID, communityID uuid.UUID, title, content string, isPublic bool) {
				agg := newTestPostAggregate(authorID, communityID)
				svc.On("CreatePost", ctx, title, content, authorID, communityID, isPublic).
					Return(agg, nil)
			},
			expectErrorCode: "",
		},
		{
			name:     "domain validation error",
			title:    "Hi",
			content:  "ok",
			isPublic: true,
			setupMocks: func(svc *MockPostCommandService, authorID, communityID uuid.UUID, title, content string, isPublic bool) {
				err := domain.NewValidationFailed(map[string]string{"title": "too short"})
				svc.On("CreatePost", ctx, title, content, authorID, communityID, isPublic).
					Return(nil, err)
			},
			expectErrorCode: domain.ErrorCodeValidationFailed,
		},
		{
			name:     "post already exists",
			title:    "Duplicate",
			content:  "content",
			isPublic: true,
			setupMocks: func(svc *MockPostCommandService, authorID, communityID uuid.UUID, title, content string, isPublic bool) {
				err := domain.NewPostAlreadyExistsError(uuid.New().String())
				svc.On("CreatePost", ctx, title, content, authorID, communityID, isPublic).
					Return(nil, err)
			},
			expectErrorCode: domain.ErrorCodePostAlreadyExists,
		},
		{
			name:     "internal error",
			title:    "OK",
			content:  "OK",
			isPublic: true,
			setupMocks: func(svc *MockPostCommandService, authorID, communityID uuid.UUID, title, content string, isPublic bool) {
				svc.On("CreatePost", ctx, title, content, authorID, communityID, isPublic).
					Return(nil, shared.NewInternalServerError(errors.New("db down")))
			},
			expectErrorCode: shared.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorID := uuid.New()
			communityID := uuid.New()
			mockSvc := new(MockPostCommandService)
			tt.setupMocks(mockSvc, authorID, communityID, tt.title, tt.content, tt.isPublic)

			handler := NewPostCommandHandler(mockSvc)
			agg, err := handler.CreatePost(ctx, tt.title, tt.content, authorID, communityID, tt.isPublic)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				assert.NotNil(t, agg)
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
				assert.Nil(t, agg)
			}
		})
	}
}
func TestPostCommandHandler_UpdatePost(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		newTitle        string
		newContent      string
		setupMocks      func(svc *MockPostCommandService, postID, requesterID uuid.UUID, newTitle, newContent string)
		expectErrorCode shared.ErrorCode
	}{
		{
			name:       "success",
			newTitle:   "New Title",
			newContent: "New content.",
			setupMocks: func(svc *MockPostCommandService, postID, requesterID uuid.UUID, newTitle, newContent string) {
				svc.On("UpdatePost", ctx, postID, newTitle, newContent, requesterID).Return(nil)
			},
			expectErrorCode: "",
		},
		{
			name:       "unauthorized",
			newTitle:   "New",
			newContent: "New",
			setupMocks: func(svc *MockPostCommandService, postID, requesterID uuid.UUID, newTitle, newContent string) {
				err := domain.NewUserNotPostAuthorError(postID.String(), requesterID.String())
				svc.On("UpdatePost", ctx, postID, newTitle, newContent, requesterID).Return(err)
			},
			expectErrorCode: domain.ErrorCodeUserNotPostAuthor,
		},
		{
			name:       "post not found",
			newTitle:   "New",
			newContent: "New",
			setupMocks: func(svc *MockPostCommandService, postID, requesterID uuid.UUID, newTitle, newContent string) {
				svc.On("UpdatePost", ctx, postID, newTitle, newContent, requesterID).
					Return(domain.NewPostNotFoundError(postID.String()))
			},
			expectErrorCode: domain.ErrorCodePostNotFound,
		},
		{
			name:       "validation error",
			newTitle:   "OK",
			newContent: "hi",
			setupMocks: func(svc *MockPostCommandService, postID, requesterID uuid.UUID, newTitle, newContent string) {
				err := domain.NewValidationFailed(map[string]string{"content": "too short"})
				svc.On("UpdatePost", ctx, postID, newTitle, newContent, requesterID).Return(err)
			},
			expectErrorCode: domain.ErrorCodeValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postID := uuid.New()
			requesterID := uuid.New()
			mockSvc := new(MockPostCommandService)
			tt.setupMocks(mockSvc, postID, requesterID, tt.newTitle, tt.newContent)

			handler := NewPostCommandHandler(mockSvc)
			err := handler.UpdatePost(ctx, postID, tt.newTitle, tt.newContent, requesterID)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
			}
		})
	}
}
