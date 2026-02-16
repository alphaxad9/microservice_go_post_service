// Package api provides HTTP controllers for post command operations.
// my-go-backend/post_service/src/posts/api/controllers/command_controllers.go
package api

import (
	"errors"
	"net/http"

	"my-go-backend/post_service/internal/contextkeys"
	"my-go-backend/post_service/src/posts/application/posts/handlers"
	"my-go-backend/post_service/src/posts/application/posts/services"
	"my-go-backend/post_service/src/posts/ports"
	"my-go-backend/post_service/src/shared"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PostCommandController handles HTTP requests for post mutations.
type PostCommandController struct {
	handler *handlers.PostCommandHandler
}

// NewPostCommandController creates a new PostCommandController.
func NewPostCommandController(postService services.PostCommandService) *PostCommandController {
	return &PostCommandController{
		handler: handlers.NewPostCommandHandler(postService),
	}
}

// In your domain or a new file, e.g., ports/post_dto.go

// respondError writes a standardized error response using domain.AppError or generic message.
func respondError(c *gin.Context, err error) {
	var appErr *shared.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, gin.H{
			"error":   appErr.Message,
			"code":    appErr.Code,
			"details": appErr.Details,
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

// getUserIDFromContext retrieves the authenticated user ID from Gin context.
// Assumes auth middleware has set context.Value(contextkeys.UserIDKey) as string (from JWT claims).
// We convert it to uuid.UUID here.
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	val := c.Request.Context().Value(contextkeys.UserIDKey)
	if val == nil {
		return uuid.Nil, errors.New("user not authenticated")
	}

	userIDStr, ok := val.(string)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID type")
	}

	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID format in token")
	}

	return id, nil
}

// parseUUIDParam extracts and parses a UUID path parameter.
func parseUUIDParam(c *gin.Context, paramName string) (uuid.UUID, error) {
	idStr := c.Param(paramName)
	if idStr == "" {
		return uuid.Nil, errors.New("missing ID parameter")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid UUID format")
	}
	return id, nil
}

// CreatePost handles POST /posts
func (ctrl *PostCommandController) CreatePost(c *gin.Context) {
	var req ports.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	authorID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	post, err := ctrl.handler.CreatePost(c.Request.Context(), req.Title, req.Content, authorID, req.CommunityID, req.IsPublic)
	if err != nil {
		respondError(c, err)
		return
	}
	dto := ports.PostResponse{
		ID:          post.ID().String(),
		Title:       post.Title(),
		Content:     post.Content(),
		AuthorID:    post.AuthorID().String(),
		CommunityID: post.CommunityID().String(),
		IsPublic:    post.IsPublic(),
		Likes:       post.LikesCount(),
		CreatedAt:   post.CreatedAt(),
		UpdatedAt:   post.UpdatedAt(),
	}
	c.JSON(http.StatusCreated, dto)
}

// UpdatePost handles PUT /posts/:id
func (ctrl *PostCommandController) UpdatePost(c *gin.Context) {
	postID, err := parseUUIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	var req ports.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	requesterID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	err = ctrl.handler.UpdatePost(c.Request.Context(), postID, req.Title, req.Content, requesterID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post updated successfully"})
}

// TogglePostVisibility handles PATCH /posts/:id/visibility
func (ctrl *PostCommandController) TogglePostVisibility(c *gin.Context) {
	postID, err := parseUUIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	var req ports.TogglePostVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	requesterID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	err = ctrl.handler.TogglePostVisibility(c.Request.Context(), postID, req.IsPublic, requesterID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post visibility updated"})
}

// LikePost handles POST /posts/:id/like
func (ctrl *PostCommandController) LikePost(c *gin.Context) {
	postID, err := parseUUIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	err = ctrl.handler.LikePost(c.Request.Context(), postID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post liked"})
}

// UnlikePost handles POST /posts/:id/unlike
func (ctrl *PostCommandController) UnlikePost(c *gin.Context) {
	postID, err := parseUUIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	err = ctrl.handler.UnlikePost(c.Request.Context(), postID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post unliked"})
}

// AddCommentToPost handles POST /posts/:id/comment
func (ctrl *PostCommandController) AddCommentToPost(c *gin.Context) {
	postID, err := parseUUIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	err = ctrl.handler.AddCommentToPost(c.Request.Context(), postID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment count incremented"})
}

// RemoveCommentFromPost handles DELETE /posts/:id/comment
func (ctrl *PostCommandController) RemoveCommentFromPost(c *gin.Context) {
	postID, err := parseUUIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	err = ctrl.handler.RemoveCommentFromPost(c.Request.Context(), postID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment count decremented"})
}

// DeletePost handles DELETE /posts/:id
func (ctrl *PostCommandController) DeletePost(c *gin.Context) {
	postID, err := parseUUIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	requesterID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	err = ctrl.handler.DeletePost(c.Request.Context(), postID, requesterID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "post deleted successfully"})
}
