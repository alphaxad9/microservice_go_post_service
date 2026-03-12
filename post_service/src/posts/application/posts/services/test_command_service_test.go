// Package services_test tests the PostCommandService implementation.
package services

import (
	"context"
	"errors"
	"testing"

	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain/events"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain/outbox"
	shared "github.com/alphaxad9/my-go-backend/post_service/src/shared"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// -----------------------
// Mocks
// -----------------------

type MockPostRepo struct {
	mock.Mock
}

func (m *MockPostRepo) Create(ctx context.Context, agg *domain.PostAggregate) error {
	args := m.Called(ctx, agg)
	return args.Error(0)
}

func (m *MockPostRepo) Update(ctx context.Context, agg *domain.PostAggregate) error {
	args := m.Called(ctx, agg)
	return args.Error(0)
}

func (m *MockPostRepo) Delete(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostRepo) GetByID(ctx context.Context, postID uuid.UUID) (*domain.PostAggregate, error) {
	args := m.Called(ctx, postID)
	if agg, ok := args.Get(0).(*domain.PostAggregate); ok {
		return agg, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostRepo) Exists(ctx context.Context, postID uuid.UUID) (bool, error) {
	args := m.Called(ctx, postID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPostRepo) ExistsWithTitleInCommunity(ctx context.Context, communityID uuid.UUID, title string) (bool, error) {
	args := m.Called(ctx, communityID, title)
	return args.Bool(0), args.Error(1)
}

type MockOutboxRepo struct {
	mock.Mock
}

func (m *MockOutboxRepo) Save(ctx context.Context, event *outbox.OutboxEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockOutboxRepo) GetUnpublishedEvents(ctx context.Context, limit int) ([]*outbox.OutboxEvent, error) {
	panic("not implemented")
}

func (m *MockOutboxRepo) MarkAsPublished(ctx context.Context, outboxID uuid.UUID) error {
	panic("not implemented")
}

func (m *MockOutboxRepo) MarkAsFailed(ctx context.Context, outboxID uuid.UUID, errorMsg string) error {
	panic("not implemented")
}

// -----------------------
// Helper
// -----------------------

func newTestPostAggregate() *domain.PostAggregate {
	agg, _ := domain.CreatePost(
		"Test Title",
		"Test content here.",
		uuid.New(),
		uuid.New(),
		true,
	)
	return agg
}

// -----------------------
// Tests
// -----------------------
func TestPostCommandService_UpdatePost(t *testing.T) {
	ctx := context.Background()
	postID := uuid.New()

	tests := []struct {
		name            string
		newTitle        string
		newContent      string
		requesterID     uuid.UUID
		setupMocks      func(*MockPostRepo, *MockOutboxRepo, uuid.UUID)
		expectErrorCode shared.ErrorCode
	}{
		{
			name:        "success",
			newTitle:    "Updated Title",    // ✅ valid
			newContent:  "Updated content.", // ✅ ≥10 chars
			requesterID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepo, outboxRepo *MockOutboxRepo, requesterID uuid.UUID) {
				agg, err := domain.CreatePost("Original Title", "Original content.", requesterID, uuid.New(), true)
				assert.NoError(t, err)
				postRepo.On("GetByID", ctx, postID).Return(agg, nil)
				postRepo.On("Update", ctx, mock.AnythingOfType("*domain.PostAggregate")).Return(nil)
				outboxRepo.On("Save", ctx, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)
			},
			expectErrorCode: "",
		},
		{
			name:        "post not found",
			newTitle:    "Title",
			newContent:  "Content.",
			requesterID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepo, _ *MockOutboxRepo, _ uuid.UUID) {
				postRepo.On("GetByID", ctx, postID).Return(nil, domain.ErrPostNotFound)
			},
			expectErrorCode: domain.ErrorCodePostNotFound,
		},
		{
			name:        "unauthorized update",
			newTitle:    "New Title",
			newContent:  "New content.",
			requesterID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepo, _ *MockOutboxRepo, requesterID uuid.UUID) {
				realAuthor := uuid.New()
				for realAuthor == requesterID {
					realAuthor = uuid.New()
				}
				agg, err := domain.CreatePost("Title", "Valid content here.", realAuthor, uuid.New(), true)
				assert.NoError(t, err)
				postRepo.On("GetByID", ctx, postID).Return(agg, nil)
			},
			expectErrorCode: domain.ErrorCodeUserNotPostAuthor,
		},
		{
			name:        "update validation fails (empty title)",
			newTitle:    "", // ❌ invalid
			newContent:  "Valid content.",
			requesterID: uuid.New(),
			setupMocks: func(postRepo *MockPostRepo, _ *MockOutboxRepo, requesterID uuid.UUID) {
				agg, err := domain.CreatePost("Title", "Original content.", requesterID, uuid.New(), true)
				assert.NoError(t, err)
				postRepo.On("GetByID", ctx, postID).Return(agg, nil)
				// No Update or Save expected — validation fails early
			},
			expectErrorCode: domain.ErrorCodeValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postRepo := new(MockPostRepo)
			outboxRepo := new(MockOutboxRepo)
			tt.setupMocks(postRepo, outboxRepo, tt.requesterID)

			service := NewPostCommandService(postRepo, outboxRepo)
			err := service.UpdatePost(ctx, postID, tt.newTitle, tt.newContent, tt.requesterID)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				postRepo.AssertExpectations(t)
				outboxRepo.AssertExpectations(t)
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
func TestPostCommandService_CreatePost(t *testing.T) {
	ctx := context.Background()
	authorID := uuid.New()
	communityID := uuid.New()

	tests := []struct {
		name            string
		title           string
		content         string
		setupMocks      func(*MockPostRepo, *MockOutboxRepo)
		expectErrorCode shared.ErrorCode
	}{
		{
			name:    "success",
			title:   "Valid Title",
			content: "Valid content with enough characters.", // ✅ >10 chars
			setupMocks: func(postRepo *MockPostRepo, outboxRepo *MockOutboxRepo) {
				postRepo.On("ExistsWithTitleInCommunity", ctx, communityID, "Valid Title").Return(false, nil)
				postRepo.On("Create", ctx, mock.AnythingOfType("*domain.PostAggregate")).Return(nil)
				outboxRepo.On("Save", ctx, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)
			},
			expectErrorCode: "",
		},
		{
			name:    "invalid title (too short)",
			title:   "Hi",                                    // 2 chars
			content: "Valid content with enough characters.", // still valid content
			setupMocks: func(_ *MockPostRepo, _ *MockOutboxRepo) {
				// Fails in domain.CreatePost due to title
			},
			expectErrorCode: domain.ErrorCodeValidationFailed,
		},
		{
			name:    "duplicate title in community",
			title:   "Duplicate Post",
			content: "This is valid content.", // ✅ 24 chars
			setupMocks: func(postRepo *MockPostRepo, _ *MockOutboxRepo) {
				postRepo.On("ExistsWithTitleInCommunity", ctx, communityID, "Duplicate Post").Return(true, nil)
			},
			expectErrorCode: domain.ErrorCodePostAlreadyExists,
		},
		{
			name:    "repo create fails",
			title:   "Title",
			content: "Sufficient content for validation.", // ✅ >10 chars
			setupMocks: func(postRepo *MockPostRepo, _ *MockOutboxRepo) {
				postRepo.On("ExistsWithTitleInCommunity", ctx, communityID, "Title").Return(false, nil)
				postRepo.On("Create", ctx, mock.AnythingOfType("*domain.PostAggregate")).Return(errors.New("db error"))
			},
			expectErrorCode: shared.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postRepo := new(MockPostRepo)
			outboxRepo := new(MockOutboxRepo)
			tt.setupMocks(postRepo, outboxRepo)

			service := NewPostCommandService(postRepo, outboxRepo)
			_, err := service.CreatePost(ctx, tt.title, tt.content, authorID, communityID, true)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				postRepo.AssertExpectations(t)
				outboxRepo.AssertExpectations(t)
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

func TestPostCommandService_LikePost(t *testing.T) {
	ctx := context.Background()
	postID := uuid.New()
	agg := newTestPostAggregate()

	postRepo := new(MockPostRepo)
	outboxRepo := new(MockOutboxRepo)

	postRepo.On("GetByID", ctx, postID).Return(agg, nil)
	postRepo.On("Update", ctx, mock.AnythingOfType("*domain.PostAggregate")).Return(nil)
	outboxRepo.On("Save", ctx, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)

	service := NewPostCommandService(postRepo, outboxRepo)
	err := service.LikePost(ctx, postID)

	assert.NoError(t, err)
	postRepo.AssertExpectations(t)
	outboxRepo.AssertExpectations(t)

	// Verify event type
	outboxRepo.AssertCalled(t, "Save", ctx, mock.MatchedBy(func(e *outbox.OutboxEvent) bool {
		return e.EventType == events.PostEventTypeLiked
	}))
}

func TestPostCommandService_DeletePost(t *testing.T) {
	ctx := context.Background()
	postID := uuid.New()
	authorID := uuid.New()
	agg := newTestPostAggregate()
	// Force author ID for test
	agg = domain.ReconstructPostAggregate(
		postID,
		authorID,
		agg.CommunityID(),
		agg.Title(),
		agg.Content(),
		agg.IsPublic(),
		agg.LikesCount(),
		agg.CommentCount(),
		agg.CreatedAt(),
		agg.UpdatedAt(),
	)

	tests := []struct {
		name          string
		requesterID   uuid.UUID
		setupMocks    func(*MockPostRepo, *MockOutboxRepo)
		expectError   bool
		errorContains string
	}{
		{
			name:        "success",
			requesterID: authorID,
			setupMocks: func(postRepo *MockPostRepo, outboxRepo *MockOutboxRepo) {
				postRepo.On("GetByID", ctx, postID).Return(agg, nil)
				outboxRepo.On("Save", ctx, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)
				postRepo.On("Delete", ctx, postID).Return(nil)
			},
			expectError: false,
		},
		{
			name:        "unauthorized deletion",
			requesterID: uuid.New(), // not author
			setupMocks: func(postRepo *MockPostRepo, _ *MockOutboxRepo) {
				postRepo.On("GetByID", ctx, postID).Return(agg, nil)
			},
			expectError:   true,
			errorContains: "not the author",
		},
		{
			name:        "post not found on load",
			requesterID: authorID,
			setupMocks: func(postRepo *MockPostRepo, _ *MockOutboxRepo) {
				postRepo.On("GetByID", ctx, postID).Return(nil, domain.ErrPostNotFound)
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name:        "post not found on delete",
			requesterID: authorID,
			setupMocks: func(postRepo *MockPostRepo, outboxRepo *MockOutboxRepo) {
				postRepo.On("GetByID", ctx, postID).Return(agg, nil)
				outboxRepo.On("Save", ctx, mock.AnythingOfType("*outbox.OutboxEvent")).Return(nil)
				postRepo.On("Delete", ctx, postID).Return(domain.ErrPostNotFound)
			},
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postRepo := new(MockPostRepo)
			outboxRepo := new(MockOutboxRepo)
			tt.setupMocks(postRepo, outboxRepo)

			service := NewPostCommandService(postRepo, outboxRepo)
			err := service.DeletePost(ctx, postID, tt.requesterID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				postRepo.AssertExpectations(t)
				outboxRepo.AssertExpectations(t)
			}
		})
	}
}
