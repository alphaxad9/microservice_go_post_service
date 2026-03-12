// Package handlers tests the PostQueryHandler implementation.
package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alphaxad9/my-go-backend/post_service/external"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"
	shared "github.com/alphaxad9/my-go-backend/post_service/src/shared"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// -----------------------
// Mocks
// -----------------------

type MockPostQueryService struct {
	mock.Mock
}

func (m *MockPostQueryService) GetPostByID(ctx context.Context, postID uuid.UUID) (*domain.PostView, error) {
	args := m.Called(ctx, postID)
	if view, ok := args.Get(0).(*domain.PostView); ok {
		return view, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostQueryService) GetPostsByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.PostView, error) {
	args := m.Called(ctx, authorID, limit, offset)
	if views, ok := args.Get(0).([]*domain.PostView); ok {
		return views, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostQueryService) GetPostsByCommunity(ctx context.Context, communityID uuid.UUID, requesterID *uuid.UUID, limit, offset int) ([]*domain.PostView, error) {
	args := m.Called(ctx, communityID, requesterID, limit, offset)
	if views, ok := args.Get(0).([]*domain.PostView); ok {
		return views, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostQueryService) SearchPosts(ctx context.Context, query string, limit, offset int) ([]*domain.PostView, error) {
	args := m.Called(ctx, query, limit, offset)
	if views, ok := args.Get(0).([]*domain.PostView); ok {
		return views, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostQueryService) GetPostCountByAuthor(ctx context.Context, authorID uuid.UUID) (int, error) {
	args := m.Called(ctx, authorID)
	return args.Int(0), args.Error(1)
}

func (m *MockPostQueryService) PostExists(ctx context.Context, postID uuid.UUID) (bool, error) {
	args := m.Called(ctx, postID)
	return args.Bool(0), args.Error(1)
}

// -----------------------

// MockUserAPIClient mimics *services.UserAPIClient method signature
type MockUserAPIClient struct {
	mock.Mock
}

func (m *MockUserAPIClient) GetUserByID(userID uuid.UUID) (*external.UserView, error) {
	args := m.Called(userID)
	if user, ok := args.Get(0).(*external.UserView); ok {
		return user, args.Error(1)
	}
	return nil, args.Error(1)
}

// -----------------------
// Helper
// -----------------------

var testTime = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

func newTestPostView(authorID, communityID uuid.UUID) *domain.PostView {
	return &domain.PostView{
		ID:           uuid.New(),
		AuthorID:     authorID,
		CommunityID:  communityID,
		Title:        "Test Post",
		Content:      "This is a test post content.",
		IsPublic:     true,
		LikesCount:   5,
		CommentCount: 2,
		CreatedAt:    testTime,
		UpdatedAt:    testTime,
	}
}

func newTestUserView(userID uuid.UUID) *external.UserView {
	return &external.UserView{
		UserID:    userID,
		Username:  "testuser",
		FirstName: "Test",
		LastName:  "User",
	}
}

// -----------------------
// Tests
// -----------------------

func TestPostQueryHandler_SearchPostsEnriched(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		query           string
		limit           int
		offset          int
		setupMocks      func(postSvc *MockPostQueryService, userClient *MockUserAPIClient)
		expectEmpty     bool
		expectErrorCode shared.ErrorCode
	}{
		{
			name:  "empty query returns empty list",
			query: "",
			limit: 10,
			setupMocks: func(_ *MockPostQueryService, _ *MockUserAPIClient) {
				// no calls expected
			},
			expectEmpty: true,
		},
		{
			name:   "non-empty query success",
			query:  "test",
			limit:  10,
			offset: 0,
			setupMocks: func(postSvc *MockPostQueryService, userClient *MockUserAPIClient) {
				authorID := uuid.New()
				communityID := uuid.New()
				posts := []*domain.PostView{newTestPostView(authorID, communityID)}
				userView := newTestUserView(authorID)

				postSvc.On("SearchPosts", ctx, "test", 10, 0).Return(posts, nil)
				userClient.On("GetUserByID", authorID).Return(userView, nil)
			},
			expectEmpty: false,
		},
		{
			name:  "search error",
			query: "error",
			limit: 5,
			setupMocks: func(postSvc *MockPostQueryService, _ *MockUserAPIClient) {
				// Return wrapped error like the real service would
				postSvc.On("SearchPosts", ctx, "error", 5, 0).
					Return(nil, shared.NewInternalServerError(errors.New("search failed")))
			},
			expectErrorCode: shared.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postSvc := new(MockPostQueryService)
			userClient := new(MockUserAPIClient)
			tt.setupMocks(postSvc, userClient)

			handler := NewPostQueryHandler(postSvc, userClient)
			dto, err := handler.SearchPostsEnriched(ctx, tt.query, tt.limit, tt.offset)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				assert.NotNil(t, dto)
				if tt.expectEmpty {
					assert.Empty(t, dto.Posts)
					assert.Equal(t, 0, dto.TotalCount)
				} else {
					assert.NotEmpty(t, dto.Posts)
					// Note: total count is approximated as offset + len(posts)
					assert.Equal(t, tt.offset+len(dto.Posts), dto.TotalCount)
				}
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
				assert.Nil(t, dto)
			}
		})
	}
}

func TestPostQueryHandler_GetPostWithAuthor(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		setupMocks      func(postID uuid.UUID, postSvc *MockPostQueryService, userClient *MockUserAPIClient) (*domain.PostView, *external.UserView)
		expectErrorCode shared.ErrorCode
	}{
		{
			name: "success with user data",
			setupMocks: func(postID uuid.UUID, postSvc *MockPostQueryService, userClient *MockUserAPIClient) (*domain.PostView, *external.UserView) {
				authorID := uuid.New()
				postView := &domain.PostView{
					ID:           postID, // ← Use the same postID!
					AuthorID:     authorID,
					CommunityID:  uuid.New(),
					Title:        "Test Post",
					Content:      "This is a test post content.",
					IsPublic:     true,
					LikesCount:   5,
					CommentCount: 2,
					CreatedAt:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				}
				userView := &external.UserView{
					UserID:    authorID,
					Username:  "testuser",
					FirstName: "Test",
					LastName:  "User",
				}

				postSvc.On("GetPostByID", ctx, postID).Return(postView, nil)
				userClient.On("GetUserByID", authorID).Return(userView, nil)

				return postView, userView
			},
			expectErrorCode: "",
		},
		{
			name: "success with user service failure (fallback)",
			setupMocks: func(postID uuid.UUID, postSvc *MockPostQueryService, userClient *MockUserAPIClient) (*domain.PostView, *external.UserView) {
				authorID := uuid.New()
				postView := &domain.PostView{
					ID:           postID,
					AuthorID:     authorID,
					CommunityID:  uuid.New(),
					Title:        "Test Post",
					Content:      "This is a test post content.",
					IsPublic:     true,
					LikesCount:   5,
					CommentCount: 2,
					CreatedAt:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				}

				postSvc.On("GetPostByID", ctx, postID).Return(postView, nil)
				userClient.On("GetUserByID", authorID).Return(nil, errors.New("user service down"))

				return postView, nil
			},
			expectErrorCode: "",
		},
		{
			name: "post not found",
			setupMocks: func(postID uuid.UUID, postSvc *MockPostQueryService, userClient *MockUserAPIClient) (*domain.PostView, *external.UserView) {
				postSvc.On("GetPostByID", ctx, postID).Return(nil, domain.ErrPostNotFound)
				return nil, nil
			},
			expectErrorCode: domain.ErrorCodePostNotFound,
		},
		{
			name: "internal error from post service",
			setupMocks: func(postID uuid.UUID, postSvc *MockPostQueryService, userClient *MockUserAPIClient) (*domain.PostView, *external.UserView) {
				postSvc.On("GetPostByID", ctx, postID).Return(nil, shared.NewInternalServerError(errors.New("db error")))
				return nil, nil
			},
			expectErrorCode: shared.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postID := uuid.New() // ← Generate ONCE per test case
			postSvc := new(MockPostQueryService)
			userClient := new(MockUserAPIClient)

			postView, userView := tt.setupMocks(postID, postSvc, userClient)

			handler := NewPostQueryHandler(postSvc, userClient)
			dto, err := handler.GetPostWithAuthor(ctx, postID) // ← Use the SAME postID

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				assert.NotNil(t, dto)
				assert.Equal(t, postID, dto.ID)
				if postView != nil {
					assert.Equal(t, postView.AuthorID, dto.Author.UserID)
					if userView != nil {
						assert.Equal(t, "testuser", dto.Author.Username)
					} else {
						assert.Empty(t, dto.Author.Username)
						assert.Equal(t, postView.AuthorID, dto.Author.UserID)
					}
				}
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
				assert.Nil(t, dto)
			}
		})
	}
}
func TestPostQueryHandler_GetPostsByCommunityWithAuthors(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		setup           func() (communityID uuid.UUID, requesterID *uuid.UUID, limit, offset int)
		setupMocks      func(postSvc *MockPostQueryService, userClient *MockUserAPIClient, communityID uuid.UUID, requesterID *uuid.UUID, limit, offset int)
		expectErrorCode shared.ErrorCode
		expectFallback  bool
	}{
		{
			name: "success with all authors fetched",
			setup: func() (uuid.UUID, *uuid.UUID, int, int) {
				return uuid.New(), &[]uuid.UUID{uuid.New()}[0], 10, 0
			},
			setupMocks: func(postSvc *MockPostQueryService, userClient *MockUserAPIClient, communityID uuid.UUID, requesterID *uuid.UUID, limit, offset int) {
				author1 := uuid.New()
				author2 := uuid.New()
				post1 := newTestPostView(author1, communityID)
				post2 := newTestPostView(author2, communityID)
				posts := []*domain.PostView{post1, post2}

				postSvc.On("GetPostsByCommunity", ctx, communityID, requesterID, limit, offset).Return(posts, nil)
				userClient.On("GetUserByID", author1).Return(newTestUserView(author1), nil)
				userClient.On("GetUserByID", author2).Return(newTestUserView(author2), nil)
			},
			expectErrorCode: "",
		},
		{
			name: "partial user fetch failure (graceful fallback)",
			setup: func() (uuid.UUID, *uuid.UUID, int, int) {
				return uuid.New(), nil, 5, 0
			},
			setupMocks: func(postSvc *MockPostQueryService, userClient *MockUserAPIClient, communityID uuid.UUID, requesterID *uuid.UUID, limit, offset int) {
				author1 := uuid.New()
				author2 := uuid.New()
				post1 := newTestPostView(author1, communityID)
				post2 := newTestPostView(author2, communityID)
				posts := []*domain.PostView{post1, post2}

				postSvc.On("GetPostsByCommunity", ctx, communityID, requesterID, limit, offset).Return(posts, nil)
				userClient.On("GetUserByID", author1).Return(nil, errors.New("fail"))
				userClient.On("GetUserByID", author2).Return(newTestUserView(author2), nil)
			},
			expectErrorCode: "",
			expectFallback:  true,
		},
		{
			name: "nil community ID",
			setup: func() (uuid.UUID, *uuid.UUID, int, int) {
				requesterID := uuid.New()
				return uuid.Nil, &requesterID, 10, 0
			},
			setupMocks: func(postSvc *MockPostQueryService, userClient *MockUserAPIClient, communityID uuid.UUID, requesterID *uuid.UUID, limit, offset int) {
				// Mock the service returning an error for uuid.Nil
				postSvc.On("GetPostsByCommunity", ctx, communityID, requesterID, limit, offset).
					Return(nil, domain.NewInvalidPostDataError("community_id is required"))
			},
			expectErrorCode: domain.ErrorCodeInvalidPostData,
		},
		{
			name: "post service error",
			setup: func() (uuid.UUID, *uuid.UUID, int, int) {
				return uuid.New(), &[]uuid.UUID{uuid.New()}[0], 10, 0
			},
			setupMocks: func(postSvc *MockPostQueryService, userClient *MockUserAPIClient, communityID uuid.UUID, requesterID *uuid.UUID, limit, offset int) {
				postSvc.On("GetPostsByCommunity", ctx, communityID, requesterID, limit, offset).
					Return(nil, shared.NewInternalServerError(errors.New("db error")))
			},
			expectErrorCode: shared.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			communityID, requesterID, limit, offset := tt.setup()
			postSvc := new(MockPostQueryService)
			userClient := new(MockUserAPIClient)

			tt.setupMocks(postSvc, userClient, communityID, requesterID, limit, offset)

			handler := NewPostQueryHandler(postSvc, userClient)
			dto, err := handler.GetPostsByCommunityWithAuthors(ctx, communityID, requesterID, limit, offset)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				assert.NotNil(t, dto)
				assert.Equal(t, offset+len(dto.Posts), dto.TotalCount) // approximated
				assert.Equal(t, 1, dto.Page)

				// For success cases, we can check author data if needed
				if !tt.expectFallback {
					for _, post := range dto.Posts {
						assert.Equal(t, "testuser", post.Author.Username)
					}
				} else {
					// At least one author should have empty username (fallback)
					hasFallback := false
					for _, post := range dto.Posts {
						if post.Author.Username == "" {
							hasFallback = true
							break
						}
					}
					assert.True(t, hasFallback, "expected at least one fallback author")
				}
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
				assert.Nil(t, dto)
			}
		})
	}
}
func TestPostQueryHandler_GetPostsByAuthorWithAuthors(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		setup           func() (authorID uuid.UUID, limit, offset int)
		setupMocks      func(postSvc *MockPostQueryService, userClient *MockUserAPIClient, authorID uuid.UUID, limit, offset int)
		expectErrorCode shared.ErrorCode
		expectFallback  bool
	}{
		{
			name: "success with user data",
			setup: func() (uuid.UUID, int, int) {
				return uuid.New(), 10, 0
			},
			setupMocks: func(postSvc *MockPostQueryService, userClient *MockUserAPIClient, authorID uuid.UUID, limit, offset int) {
				communityID := uuid.New()
				post1 := newTestPostView(authorID, communityID)
				post2 := newTestPostView(authorID, communityID)
				posts := []*domain.PostView{post1, post2}

				postSvc.On("GetPostsByAuthor", ctx, authorID, limit, offset).Return(posts, nil)
				postSvc.On("GetPostCountByAuthor", ctx, authorID).Return(2, nil)
				userClient.On("GetUserByID", authorID).Return(newTestUserView(authorID), nil)
			},
			expectErrorCode: "",
		},
		{
			name: "success with user service failure (fallback)",
			setup: func() (uuid.UUID, int, int) {
				return uuid.New(), 5, 0
			},
			setupMocks: func(postSvc *MockPostQueryService, userClient *MockUserAPIClient, authorID uuid.UUID, limit, offset int) {
				communityID := uuid.New()
				post1 := newTestPostView(authorID, communityID)
				post2 := newTestPostView(authorID, communityID)
				posts := []*domain.PostView{post1, post2}

				postSvc.On("GetPostsByAuthor", ctx, authorID, limit, offset).Return(posts, nil)
				postSvc.On("GetPostCountByAuthor", ctx, authorID).Return(2, nil)
				userClient.On("GetUserByID", authorID).Return(nil, errors.New("timeout"))
			},
			expectErrorCode: "",
			expectFallback:  true,
		},
		{
			name: "nil author ID",
			setup: func() (uuid.UUID, int, int) {
				return uuid.Nil, 10, 0
			},
			setupMocks: func(postSvc *MockPostQueryService, userClient *MockUserAPIClient, authorID uuid.UUID, limit, offset int) {
				// Mock the service returning an error for uuid.Nil
				postSvc.On("GetPostsByAuthor", ctx, authorID, limit, offset).
					Return(nil, domain.NewInvalidPostDataError("author_id is required"))
				// Count method should also return error, but handler doesn't call it if GetPostsByAuthor fails
			},
			expectErrorCode: domain.ErrorCodeInvalidPostData,
		},
		{
			name: "post service error",
			setup: func() (uuid.UUID, int, int) {
				return uuid.New(), 10, 0
			},
			setupMocks: func(postSvc *MockPostQueryService, userClient *MockUserAPIClient, authorID uuid.UUID, limit, offset int) {
				postSvc.On("GetPostsByAuthor", ctx, authorID, limit, offset).
					Return(nil, shared.NewInternalServerError(errors.New("query failed")))
			},
			expectErrorCode: shared.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authorID, limit, offset := tt.setup()
			postSvc := new(MockPostQueryService)
			userClient := new(MockUserAPIClient)

			tt.setupMocks(postSvc, userClient, authorID, limit, offset)

			handler := NewPostQueryHandler(postSvc, userClient)
			dto, err := handler.GetPostsByAuthorWithAuthors(ctx, authorID, limit, offset)

			if tt.expectErrorCode == "" {
				assert.NoError(t, err)
				assert.NotNil(t, dto)
				assert.Equal(t, 2, dto.TotalCount)
				assert.Equal(t, 1, dto.Page)

				for _, post := range dto.Posts {
					assert.Equal(t, authorID, post.Author.UserID)
					if tt.expectFallback {
						assert.Empty(t, post.Author.Username)
					} else {
						assert.Equal(t, "testuser", post.Author.Username)
					}
				}
			} else {
				assert.Error(t, err)
				var appErr *shared.AppError
				if errors.As(err, &appErr) {
					assert.Equal(t, tt.expectErrorCode, appErr.Code)
				} else {
					t.Fatalf("expected *shared.AppError, got %T", err)
				}
				assert.Nil(t, dto)
			}
		})
	}
}
