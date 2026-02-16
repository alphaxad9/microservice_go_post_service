// Package services_test tests the PostQueryService implementation.
package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"my-go-backend/post_service/src/posts/domain"
	shared "my-go-backend/post_service/src/shared"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// -----------------------
// Mocks
// -----------------------

type MockPostQueryRepo struct {
	mock.Mock
}

func (m *MockPostQueryRepo) ByID(ctx context.Context, postID uuid.UUID) (*domain.PostView, error) {
	args := m.Called(ctx, postID)
	if view, ok := args.Get(0).(*domain.PostView); ok {
		return view, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostQueryRepo) ByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.PostView, error) {
	args := m.Called(ctx, authorID, limit, offset)
	if views, ok := args.Get(0).([]*domain.PostView); ok {
		return views, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostQueryRepo) ByCommunity(ctx context.Context, communityID uuid.UUID, requesterID *uuid.UUID, limit, offset int) ([]*domain.PostView, error) {
	args := m.Called(ctx, communityID, requesterID, limit, offset)
	if views, ok := args.Get(0).([]*domain.PostView); ok {
		return views, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostQueryRepo) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*domain.PostView, error) {
	args := m.Called(ctx, query, limit, offset)
	if views, ok := args.Get(0).([]*domain.PostView); ok {
		return views, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostQueryRepo) CountByAuthor(ctx context.Context, authorID uuid.UUID) (int, error) {
	args := m.Called(ctx, authorID)
	return args.Int(0), args.Error(1)
}

func (m *MockPostQueryRepo) Exists(ctx context.Context, postID uuid.UUID) (bool, error) {
	args := m.Called(ctx, postID)
	return args.Bool(0), args.Error(1)
}

// -----------------------
// Helper
// -----------------------

var testTime = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

func newTestPostView() *domain.PostView {
	return &domain.PostView{
		ID:           uuid.New(),
		AuthorID:     uuid.New(),
		CommunityID:  uuid.New(),
		Title:        "Test Post",
		Content:      "This is a test post content.",
		IsPublic:     true,
		LikesCount:   5,
		CommentCount: 2,
		CreatedAt:    testTime,
		UpdatedAt:    testTime,
	}
}

// -----------------------
// Tests
// -----------------------

func TestPostQueryService_GetPostsByAuthor(t *testing.T) {
	ctx := context.Background()
	authorID := uuid.New()
	validViews := []*domain.PostView{newTestPostView(), newTestPostView()}

	tests := []struct {
		name            string
		authorID        uuid.UUID
		limit           int
		offset          int
		setupMocks      func(*MockPostQueryRepo)
		expectErrorCode shared.ErrorCode
	}{
		{
			name:     "success",
			authorID: authorID,
			limit:    10,
			offset:   0,
			setupMocks: func(repo *MockPostQueryRepo) {
				repo.On("ByAuthor", ctx, authorID, 10, 0).Return(validViews, nil)
			},
			expectErrorCode: "",
		},
		{
			name:            "nil author ID",
			authorID:        uuid.Nil,
			limit:           10,
			offset:          0,
			setupMocks:      func(_ *MockPostQueryRepo) {},
			expectErrorCode: domain.ErrorCodeInvalidPostData,
		},
		{
			name:     "repo returns error",
			authorID: authorID,
			limit:    5,
			offset:   20,
			setupMocks: func(repo *MockPostQueryRepo) {
				repo.On("ByAuthor", ctx, authorID, 5, 20).Return(nil, errors.New("query failed"))
			},
			expectErrorCode: shared.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockPostQueryRepo)
			tt.setupMocks(repo)

			service := NewPostQueryService(repo)
			views, err := service.GetPostsByAuthor(ctx, tt.authorID, tt.limit, tt.offset)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				assert.NotNil(t, views)
				repo.AssertExpectations(t)
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

func TestPostQueryService_GetPostCountByAuthor(t *testing.T) {
	ctx := context.Background()
	authorID := uuid.New()

	tests := []struct {
		name            string
		authorID        uuid.UUID
		setupMocks      func(*MockPostQueryRepo)
		expectErrorCode shared.ErrorCode
		expectedCount   int
	}{
		{
			name:            "success",
			authorID:        authorID,
			setupMocks:      func(repo *MockPostQueryRepo) { repo.On("CountByAuthor", ctx, authorID).Return(42, nil) },
			expectErrorCode: "",
			expectedCount:   42,
		},
		{
			name:            "nil author ID",
			authorID:        uuid.Nil,
			setupMocks:      func(_ *MockPostQueryRepo) {},
			expectErrorCode: domain.ErrorCodeInvalidPostData,
			expectedCount:   0,
		},
		{
			name:     "repo error",
			authorID: authorID,
			setupMocks: func(repo *MockPostQueryRepo) {
				repo.On("CountByAuthor", ctx, authorID).Return(0, errors.New("count query failed"))
			},
			expectErrorCode: shared.ErrorCodeInternal,
			expectedCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockPostQueryRepo)
			tt.setupMocks(repo)

			service := NewPostQueryService(repo)
			count, err := service.GetPostCountByAuthor(ctx, tt.authorID)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
				repo.AssertExpectations(t)
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

func TestPostQueryService_PostExists(t *testing.T) {
	ctx := context.Background()
	postID := uuid.New()

	tests := []struct {
		name         string
		postID       uuid.UUID
		setupMocks   func(*MockPostQueryRepo)
		expectExists bool
		expectError  bool
		errorCode    shared.ErrorCode
	}{
		{
			name:         "exists",
			postID:       postID,
			setupMocks:   func(repo *MockPostQueryRepo) { repo.On("Exists", ctx, postID).Return(true, nil) },
			expectExists: true,
			expectError:  false,
		},
		{
			name:         "does not exist",
			postID:       postID,
			setupMocks:   func(repo *MockPostQueryRepo) { repo.On("Exists", ctx, postID).Return(false, nil) },
			expectExists: false,
			expectError:  false,
		},
		{
			name:         "nil post ID returns false, no error",
			postID:       uuid.Nil,
			setupMocks:   func(_ *MockPostQueryRepo) {},
			expectExists: false,
			expectError:  false,
		},
		{
			name:         "repo error",
			postID:       postID,
			setupMocks:   func(repo *MockPostQueryRepo) { repo.On("Exists", ctx, postID).Return(false, errors.New("db error")) },
			expectExists: false,
			expectError:  true,
			errorCode:    shared.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockPostQueryRepo)
			tt.setupMocks(repo)

			service := NewPostQueryService(repo)
			exists, err := service.PostExists(ctx, tt.postID)

			if tt.expectError {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.errorCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectExists, exists)
			}
		})
	}
}

func TestPostQueryService_GetPostsByCommunity(t *testing.T) {
	ctx := context.Background()
	communityID := uuid.New()
	requesterID := uuid.New()
	validViews := []*domain.PostView{newTestPostView()}

	tests := []struct {
		name            string
		communityID     uuid.UUID
		requesterID     *uuid.UUID
		limit           int
		offset          int
		setupMocks      func(*MockPostQueryRepo)
		expectErrorCode shared.ErrorCode
	}{
		{
			name:        "success with requester",
			communityID: communityID,
			requesterID: &requesterID,
			limit:       10,
			offset:      0,
			setupMocks: func(repo *MockPostQueryRepo) {
				repo.On("ByCommunity", ctx, communityID, &requesterID, 10, 0).Return(validViews, nil)
			},
			expectErrorCode: "",
		},
		{
			name:        "success without requester (public only)",
			communityID: communityID,
			requesterID: nil,
			limit:       5,
			offset:      0,
			setupMocks: func(repo *MockPostQueryRepo) {
				repo.On("ByCommunity", ctx, communityID, (*uuid.UUID)(nil), 5, 0).Return(validViews, nil)
			},
			expectErrorCode: "",
		},
		{
			name:            "nil community ID",
			communityID:     uuid.Nil,
			requesterID:     &requesterID,
			limit:           10,
			offset:          0,
			setupMocks:      func(_ *MockPostQueryRepo) {},
			expectErrorCode: domain.ErrorCodeInvalidPostData,
		},
		{
			name:        "repo error",
			communityID: communityID,
			requesterID: &requesterID,
			limit:       10,
			offset:      0,
			setupMocks: func(repo *MockPostQueryRepo) {
				repo.On("ByCommunity", ctx, communityID, &requesterID, 10, 0).Return(nil, errors.New("db timeout"))
			},
			expectErrorCode: shared.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockPostQueryRepo)
			tt.setupMocks(repo)

			service := NewPostQueryService(repo)
			views, err := service.GetPostsByCommunity(ctx, tt.communityID, tt.requesterID, tt.limit, tt.offset)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				assert.NotNil(t, views)
				repo.AssertExpectations(t)
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

func TestPostQueryService_SearchPosts(t *testing.T) {
	ctx := context.Background()
	validViews := []*domain.PostView{newTestPostView()}

	tests := []struct {
		name            string
		query           string
		limit           int
		offset          int
		setupMocks      func(*MockPostQueryRepo)
		expectErrorCode shared.ErrorCode
		expectEmpty     bool
	}{
		{
			name:            "success with non-empty query",
			query:           "test",
			limit:           10,
			offset:          0,
			setupMocks:      func(repo *MockPostQueryRepo) { repo.On("SearchPosts", ctx, "test", 10, 0).Return(validViews, nil) },
			expectErrorCode: "",
			expectEmpty:     false,
		},
		{
			name:            "empty query returns empty slice",
			query:           "",
			limit:           10,
			offset:          0,
			setupMocks:      func(_ *MockPostQueryRepo) {},
			expectErrorCode: "",
			expectEmpty:     true,
		},
		{
			name:   "search error",
			query:  "error",
			limit:  5,
			offset: 0,
			setupMocks: func(repo *MockPostQueryRepo) {
				repo.On("SearchPosts", ctx, "error", 5, 0).Return(nil, errors.New("search backend down"))
			},
			expectErrorCode: shared.ErrorCodeInternal,
			expectEmpty:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockPostQueryRepo)
			tt.setupMocks(repo)

			service := NewPostQueryService(repo)
			views, err := service.SearchPosts(ctx, tt.query, tt.limit, tt.offset)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				if tt.expectEmpty {
					assert.Empty(t, views)
				} else {
					assert.NotEmpty(t, views)
				}
				repo.AssertExpectations(t)
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

func TestPostQueryService_GetPostByID(t *testing.T) {
	ctx := context.Background()
	postID := uuid.New()
	validView := newTestPostView()

	tests := []struct {
		name            string
		postID          uuid.UUID
		setupMocks      func(*MockPostQueryRepo)
		expectErrorCode shared.ErrorCode
		expectNil       bool
	}{
		{
			name:   "success",
			postID: postID,
			setupMocks: func(repo *MockPostQueryRepo) {
				repo.On("ByID", ctx, postID).Return(validView, nil)
			},
			expectErrorCode: "",
			expectNil:       false,
		},
		{
			name:   "post not found",
			postID: postID,
			setupMocks: func(repo *MockPostQueryRepo) {
				repo.On("ByID", ctx, postID).Return(nil, domain.ErrPostNotFound)
			},
			expectErrorCode: domain.ErrorCodePostNotFound,
			expectNil:       true,
		},
		{
			name:            "nil post ID",
			postID:          uuid.Nil,
			setupMocks:      func(_ *MockPostQueryRepo) {},
			expectErrorCode: domain.ErrorCodePostNotFound,
			expectNil:       true,
		},
		{
			name:   "internal error from repo",
			postID: postID,
			setupMocks: func(repo *MockPostQueryRepo) {
				repo.On("ByID", ctx, postID).Return(nil, errors.New("db connection failed"))
			},
			expectErrorCode: shared.ErrorCodeInternal,
			expectNil:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockPostQueryRepo)
			tt.setupMocks(repo)

			service := NewPostQueryService(repo)
			view, err := service.GetPostByID(ctx, tt.postID)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				assert.NotNil(t, view)
				repo.AssertExpectations(t)
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
				if tt.expectNil {
					assert.Nil(t, view)
				}
			}
		})
	}
}
