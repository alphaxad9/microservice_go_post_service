// Package api tests the PostCommandController HTTP handlers.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alphaxad9/my-go-backend/post_service/internal/contextkeys"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/domain"
	"github.com/alphaxad9/my-go-backend/post_service/src/posts/ports"
	shared "github.com/alphaxad9/my-go-backend/post_service/src/shared"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// -----------------------
// Mock Service (CORRECTLY IMPLEMENTED)
// -----------------------

type mockPostCommandService struct {
	mock.Mock
}

func (m *mockPostCommandService) CreatePost(
	ctx context.Context,
	title, content string,
	authorID, communityID uuid.UUID,
	isPublic bool,
) (*domain.PostAggregate, error) {
	args := m.Called(ctx, title, content, authorID, communityID, isPublic)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PostAggregate), args.Error(1)
}

func (m *mockPostCommandService) UpdatePost(
	ctx context.Context,
	postID uuid.UUID,
	newTitle, newContent string,
	requesterID uuid.UUID,
) error {
	args := m.Called(ctx, postID, newTitle, newContent, requesterID)
	return args.Error(0)
}

func (m *mockPostCommandService) TogglePostVisibility(
	ctx context.Context,
	postID uuid.UUID,
	isPublic bool,
	requesterID uuid.UUID,
) error {
	args := m.Called(ctx, postID, isPublic, requesterID)
	return args.Error(0)
}

func (m *mockPostCommandService) LikePost(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *mockPostCommandService) UnlikePost(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *mockPostCommandService) AddCommentToPost(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *mockPostCommandService) RemoveCommentFromPost(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *mockPostCommandService) DeletePost(
	ctx context.Context,
	postID uuid.UUID,
	requesterID uuid.UUID,
) error {
	args := m.Called(ctx, postID, requesterID)
	return args.Error(0)
}

// -----------------------
// Test Helpers
// -----------------------

var testTime = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

func newTestPostAggregate(authorID, communityID uuid.UUID) *domain.PostAggregate {
	id := uuid.New()
	return domain.ReconstructPostAggregate(
		id,
		authorID,
		communityID,
		"Test Title",
		"Test Content",
		true,
		0,
		0,
		testTime,
		testTime,
	)
}

// -----------------------
// Tests
// -----------------------

func TestPostCommandController_DeletePost(t *testing.T) {
	postID := uuid.New()

	tests := []struct {
		name            string
		path            string
		postIDParam     string
		userID          string
		setupMocks      func(svc *mockPostCommandService, postID, requesterID uuid.UUID)
		expectStatus    int
		expectErrorCode shared.ErrorCode
	}{
		{
			name:        "success",
			path:        "/posts/" + postID.String(),
			postIDParam: postID.String(),
			userID:      uuid.New().String(),
			setupMocks: func(svc *mockPostCommandService, pid, rid uuid.UUID) {
				svc.On("DeletePost", mock.Anything, pid, rid).Return(nil)
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "no auth",
			path:         "/posts/" + postID.String(),
			postIDParam:  postID.String(),
			userID:       "",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:        "unauthorized",
			path:        "/posts/" + postID.String(),
			postIDParam: postID.String(),
			userID:      uuid.New().String(),
			setupMocks: func(svc *mockPostCommandService, pid, rid uuid.UUID) {
				err := &shared.AppError{
					Message:    "user is not the author of this post",
					Code:       domain.ErrorCodeUserNotPostAuthor,
					HTTPStatus: http.StatusForbidden,
					Details:    map[string]string{"post_id": pid.String(), "user_id": rid.String()},
				}
				svc.On("DeletePost", mock.Anything, pid, rid).Return(err)
			},
			expectStatus:    http.StatusForbidden,
			expectErrorCode: domain.ErrorCodeUserNotPostAuthor,
		},
		{
			name:         "invalid UUID",
			path:         "/posts/not-uuid",
			postIDParam:  "not-uuid",
			userID:       uuid.New().String(),
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodDelete, tt.path, nil)

			if tt.userID != "" {
				req = req.WithContext(
					context.WithValue(req.Context(), contextkeys.UserIDKey, tt.userID),
				)
			}

			ctx.Request = req
			ctx.Params = []gin.Param{{Key: "id", Value: tt.postIDParam}}

			mockSvc := new(mockPostCommandService)

			if tt.setupMocks != nil && tt.userID != "" {
				requesterID, _ := uuid.Parse(tt.userID)
				if actualPostID, err := uuid.Parse(tt.postIDParam); err == nil {
					tt.setupMocks(mockSvc, actualPostID, requesterID)
				}
			}

			ctrl := NewPostCommandController(mockSvc)
			ctrl.DeletePost(ctx)

			assert.Equal(t, tt.expectStatus, w.Code)

			if tt.expectErrorCode != "" {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, string(tt.expectErrorCode), resp["code"])
			}
		})
	}
}

func TestPostCommandController_AddCommentToPost(t *testing.T) {
	postID := uuid.New()

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	ctx.Request, _ = http.NewRequest(http.MethodPost, "/posts/"+postID.String()+"/comment", nil)
	ctx.Params = []gin.Param{{Key: "id", Value: postID.String()}}

	mockSvc := new(mockPostCommandService)
	mockSvc.On("AddCommentToPost", mock.Anything, postID).Return(nil)

	ctrl := NewPostCommandController(mockSvc)
	ctrl.AddCommentToPost(ctx)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"message":"comment count incremented"`)
}

func TestPostCommandController_CreatePost(t *testing.T) {
	tests := []struct {
		name               string
		reqBody            interface{}
		setupMocks         func(svc *mockPostCommandService, authorID, communityID uuid.UUID)
		userIDInContext    string
		expectStatus       int
		expectErrorCode    shared.ErrorCode
		expectResponseBody bool
	}{
		{
			name: "success",
			reqBody: ports.CreatePostRequest{
				Title:       "Valid Title",
				Content:     "Valid content with enough length.",
				CommunityID: uuid.New(),
				IsPublic:    true,
			},
			setupMocks: func(svc *mockPostCommandService, authorID, communityID uuid.UUID) {
				agg := newTestPostAggregate(authorID, communityID)
				svc.On("CreatePost", mock.Anything, "Valid Title", "Valid content with enough length.", authorID, communityID, true).
					Return(agg, nil)
			},
			userIDInContext:    uuid.New().String(),
			expectStatus:       http.StatusCreated,
			expectResponseBody: true,
		},
		{
			name:            "invalid JSON",
			reqBody:         "{ invalid json }",
			userIDInContext: uuid.New().String(),
			expectStatus:    http.StatusBadRequest,
		},
		{
			name: "validation error from domain",
			reqBody: ports.CreatePostRequest{
				Title:       "Short",
				Content:     "This is valid content over ten chars.",
				CommunityID: uuid.New(),
				IsPublic:    true,
			},
			setupMocks: func(svc *mockPostCommandService, authorID, communityID uuid.UUID) {
				err := &shared.AppError{
					Message:    "validation failed",
					Code:       domain.ErrorCodeValidationFailed,
					HTTPStatus: http.StatusBadRequest,
					Details:    map[string]string{"title": "too short"},
				}
				svc.On("CreatePost", mock.Anything, "Short", "This is valid content over ten chars.", authorID, communityID, true).
					Return((*domain.PostAggregate)(nil), err)
			},
			userIDInContext: uuid.New().String(),
			expectStatus:    http.StatusBadRequest,
			expectErrorCode: domain.ErrorCodeValidationFailed,
		},
		{
			name: "missing auth",
			reqBody: ports.CreatePostRequest{
				Title:       "OK Title",
				Content:     "This is valid content over ten chars.",
				CommunityID: uuid.New(),
				IsPublic:    true,
			},
			userIDInContext: "",
			expectStatus:    http.StatusUnauthorized,
		},
		{
			name: "internal error",
			reqBody: ports.CreatePostRequest{
				Title:       "OK Title",
				Content:     "This is valid content over ten chars.",
				CommunityID: uuid.New(),
				IsPublic:    true,
			},
			setupMocks: func(svc *mockPostCommandService, authorID, communityID uuid.UUID) {
				svc.On("CreatePost", mock.Anything, "OK Title", "This is valid content over ten chars.", authorID, communityID, true).
					Return((*domain.PostAggregate)(nil), errors.New("db down"))
			},
			userIDInContext: uuid.New().String(),
			expectStatus:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			var bodyBytes []byte
			if str, ok := tt.reqBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				bodyBytes, _ = json.Marshal(tt.reqBody)
			}
			req, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			if tt.userIDInContext != "" {
				req = req.WithContext(
					context.WithValue(req.Context(), contextkeys.UserIDKey, tt.userIDInContext),
				)
			}

			ctx.Request = req

			mockSvc := new(mockPostCommandService)
			if tt.setupMocks != nil && tt.userIDInContext != "" {
				authorID, _ := uuid.Parse(tt.userIDInContext)
				var communityID uuid.UUID
				if reqTyped, ok := tt.reqBody.(ports.CreatePostRequest); ok {
					communityID = reqTyped.CommunityID
				} else {
					communityID = uuid.New()
				}
				tt.setupMocks(mockSvc, authorID, communityID)
			}

			ctrl := NewPostCommandController(mockSvc)
			ctrl.CreatePost(ctx)

			assert.Equal(t, tt.expectStatus, w.Code)

			if tt.expectErrorCode != "" {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, string(tt.expectErrorCode), resp["code"])
			}

			if tt.expectResponseBody {
				assert.Contains(t, w.Body.String(), `"title":"Test Title"`)
			}
		})
	}
}

func TestPostCommandController_UpdatePost(t *testing.T) {
	postID := uuid.New()
	validReq := ports.UpdatePostRequest{
		Title:   "Updated Valid Title",
		Content: "This is updated content with sufficient length.",
	}

	tests := []struct {
		name            string
		reqBody         interface{}
		path            string
		postIDParam     string
		userID          string
		setupMocks      func(svc *mockPostCommandService, postID, requesterID uuid.UUID)
		expectStatus    int
		expectErrorCode shared.ErrorCode
	}{
		{
			name:        "success",
			reqBody:     validReq,
			path:        "/posts/" + postID.String(),
			postIDParam: postID.String(),
			userID:      uuid.New().String(),
			setupMocks: func(svc *mockPostCommandService, pid, rid uuid.UUID) {
				svc.On("UpdatePost", mock.Anything, pid, "Updated Valid Title", "This is updated content with sufficient length.", rid).Return(nil)
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "invalid UUID in path",
			reqBody:      validReq,
			path:         "/posts/invalid-uuid",
			postIDParam:  "invalid-uuid",
			userID:       uuid.New().String(),
			expectStatus: http.StatusBadRequest,
		},
		{
			name:        "unauthorized",
			reqBody:     validReq,
			path:        "/posts/" + postID.String(),
			postIDParam: postID.String(),
			userID:      uuid.New().String(),
			setupMocks: func(svc *mockPostCommandService, pid, rid uuid.UUID) {
				err := &shared.AppError{
					Message:    "user is not the author of this post",
					Code:       domain.ErrorCodeUserNotPostAuthor,
					HTTPStatus: http.StatusForbidden,
					Details:    map[string]string{"post_id": pid.String(), "user_id": rid.String()},
				}
				svc.On("UpdatePost", mock.Anything, pid, "Updated Valid Title", "This is updated content with sufficient length.", rid).Return(err)
			},
			expectStatus:    http.StatusForbidden,
			expectErrorCode: domain.ErrorCodeUserNotPostAuthor,
		},
		{
			name:         "missing auth",
			reqBody:      validReq,
			path:         "/posts/" + postID.String(),
			postIDParam:  postID.String(),
			userID:       "",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:        "post not found",
			reqBody:     validReq,
			path:        "/posts/" + postID.String(),
			postIDParam: postID.String(),
			userID:      uuid.New().String(),
			setupMocks: func(svc *mockPostCommandService, pid, rid uuid.UUID) {
				err := &shared.AppError{
					Message:    "post not found",
					Code:       domain.ErrorCodePostNotFound,
					HTTPStatus: http.StatusNotFound,
					Details:    map[string]string{"post_id": pid.String()},
				}
				svc.On("UpdatePost", mock.Anything, pid, "Updated Valid Title", "This is updated content with sufficient length.", rid).Return(err)
			},
			expectStatus:    http.StatusNotFound,
			expectErrorCode: domain.ErrorCodePostNotFound,
		},
		{
			name: "binding validation error - title too short",
			reqBody: ports.UpdatePostRequest{
				Title:   "Hi",
				Content: "Valid content here.",
			},
			path:         "/posts/" + postID.String(),
			postIDParam:  postID.String(),
			userID:       uuid.New().String(),
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			var body []byte
			if str, ok := tt.reqBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.reqBody)
			}

			req, _ := http.NewRequest(http.MethodPut, tt.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				req = req.WithContext(
					context.WithValue(req.Context(), contextkeys.UserIDKey, tt.userID),
				)
			}

			ctx.Request = req
			ctx.Params = []gin.Param{{Key: "id", Value: tt.postIDParam}}

			mockSvc := new(mockPostCommandService)

			if tt.setupMocks != nil && tt.userID != "" {
				requesterID, _ := uuid.Parse(tt.userID)
				if actualPostID, err := uuid.Parse(tt.postIDParam); err == nil {
					tt.setupMocks(mockSvc, actualPostID, requesterID)
				}
			}

			ctrl := NewPostCommandController(mockSvc)
			ctrl.UpdatePost(ctx)

			assert.Equal(t, tt.expectStatus, w.Code)

			if tt.expectErrorCode != "" {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, string(tt.expectErrorCode), resp["code"])
			}
		})
	}
}

func TestPostCommandController_TogglePostVisibility(t *testing.T) {
	postID := uuid.New()

	tests := []struct {
		name            string
		requestBody     string
		path            string
		postIDParam     string
		userID          string
		setupMocks      func(svc *mockPostCommandService, postID, requesterID uuid.UUID)
		expectStatus    int
		expectErrorCode shared.ErrorCode
	}{
		{
			name:        "success",
			requestBody: `{"is_public": false}`,
			path:        "/posts/" + postID.String() + "/visibility",
			postIDParam: postID.String(),
			userID:      uuid.New().String(),
			setupMocks: func(svc *mockPostCommandService, pid, rid uuid.UUID) {
				svc.On("TogglePostVisibility", mock.Anything, pid, false, rid).Return(nil)
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "invalid path UUID",
			requestBody:  `{"is_public": true}`,
			path:         "/posts/invalid-uuid/visibility",
			postIDParam:  "invalid-uuid",
			userID:       uuid.New().String(),
			expectStatus: http.StatusBadRequest,
		},
		{
			name:        "unauthorized",
			requestBody: `{"is_public": true}`,
			path:        "/posts/" + postID.String() + "/visibility",
			postIDParam: postID.String(),
			userID:      uuid.New().String(),
			setupMocks: func(svc *mockPostCommandService, pid, rid uuid.UUID) {
				err := &shared.AppError{
					Message:    "user is not the author of this post",
					Code:       domain.ErrorCodeUserNotPostAuthor,
					HTTPStatus: http.StatusForbidden,
					Details:    map[string]string{"post_id": pid.String(), "user_id": rid.String()},
				}
				svc.On("TogglePostVisibility", mock.Anything, pid, true, rid).Return(err)
			},
			expectStatus:    http.StatusForbidden,
			expectErrorCode: domain.ErrorCodeUserNotPostAuthor,
		},
		{
			name:         "no auth",
			requestBody:  `{"is_public": false}`,
			path:         "/posts/" + postID.String() + "/visibility",
			postIDParam:  postID.String(),
			userID:       "",
			expectStatus: http.StatusUnauthorized,
		},
		{
			name:         "invalid JSON",
			requestBody:  `{ invalid json }`,
			path:         "/posts/" + postID.String() + "/visibility",
			postIDParam:  postID.String(),
			userID:       uuid.New().String(),
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodPatch, tt.path, bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")

			if tt.userID != "" {
				req = req.WithContext(
					context.WithValue(req.Context(), contextkeys.UserIDKey, tt.userID),
				)
			}

			ctx.Request = req
			ctx.Params = []gin.Param{{Key: "id", Value: tt.postIDParam}}

			mockSvc := new(mockPostCommandService)

			if tt.setupMocks != nil && tt.userID != "" {
				requesterID, _ := uuid.Parse(tt.userID)
				if actualPostID, err := uuid.Parse(tt.postIDParam); err == nil {
					tt.setupMocks(mockSvc, actualPostID, requesterID)
				}
			}

			ctrl := NewPostCommandController(mockSvc)
			ctrl.TogglePostVisibility(ctx)

			assert.Equal(t, tt.expectStatus, w.Code)

			if tt.expectErrorCode != "" {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, string(tt.expectErrorCode), resp["code"])
			}
		})
	}
}

func TestPostCommandController_LikePost(t *testing.T) {
	postID := uuid.New()

	tests := []struct {
		name            string
		path            string
		postIDParam     string
		expectStatus    int
		expectErrorCode shared.ErrorCode
		setupMocks      func(svc *mockPostCommandService)
	}{
		{
			name:         "success",
			path:         "/posts/" + postID.String() + "/like",
			postIDParam:  postID.String(),
			expectStatus: http.StatusOK,
			setupMocks: func(svc *mockPostCommandService) {
				svc.On("LikePost", mock.Anything, postID).Return(nil)
			},
		},
		{
			name:            "post not found",
			path:            "/posts/" + postID.String() + "/like",
			postIDParam:     postID.String(),
			expectStatus:    http.StatusNotFound,
			expectErrorCode: domain.ErrorCodePostNotFound,
			setupMocks: func(svc *mockPostCommandService) {
				err := &shared.AppError{
					Message:    "post not found",
					Code:       domain.ErrorCodePostNotFound,
					HTTPStatus: http.StatusNotFound,
					Details:    map[string]string{"post_id": postID.String()},
				}
				svc.On("LikePost", mock.Anything, postID).Return(err)
			},
		},
		{
			name:         "invalid UUID",
			path:         "/posts/invalid/like",
			postIDParam:  "invalid",
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			ctx.Request, _ = http.NewRequest(http.MethodPost, tt.path, nil)
			ctx.Params = []gin.Param{{Key: "id", Value: tt.postIDParam}}

			mockSvc := new(mockPostCommandService)
			if tt.setupMocks != nil {
				tt.setupMocks(mockSvc)
			}

			ctrl := NewPostCommandController(mockSvc)
			ctrl.LikePost(ctx)

			assert.Equal(t, tt.expectStatus, w.Code)

			if tt.expectErrorCode != "" {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.Equal(t, string(tt.expectErrorCode), resp["code"])
			}
		})
	}
}
