// Package api provides HTTP controllers for post query operations.
// my-go-backend/post_service/src/posts/api/controllers/query_controllers.go
package api

import (
	"net/http"
	"strconv"

	user_services "my-go-backend/post_service/external/services"
	"my-go-backend/post_service/src/posts/application/posts/handlers"
	"my-go-backend/post_service/src/posts/application/posts/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PostQueryController handles HTTP requests for post reads/queries.
type PostQueryController struct {
	handler *handlers.PostQueryHandler
}

// NewPostQueryController creates a new PostQueryController.
func NewPostQueryController(postQueryService services.PostQueryService, userQueryService user_services.UserQueryService) *PostQueryController {
	return &PostQueryController{
		handler: handlers.NewPostQueryHandler(postQueryService, userQueryService),
	}
}

// parsePagination extracts limit and offset from query params with safe defaults.
func parsePagination(c *gin.Context) (limit, offset int) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ = strconv.Atoi(limitStr)
	offset, _ = strconv.Atoi(offsetStr)

	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

// GetPost handles GET /posts/:id
func (ctrl *PostQueryController) GetPost(c *gin.Context) {
	postID, err := parseUUIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post ID"})
		return
	}

	post, err := ctrl.handler.GetPostWithAuthor(c.Request.Context(), postID)
	if err != nil {
		respondError(c, err)
		return
	}

	// You may want to check visibility/authority here in real app
	// (e.g. private post → only author or community member)

	c.JSON(http.StatusOK, post)
}

// GetPostsByAuthor handles GET /users/:userId/posts
func (ctrl *PostQueryController) GetPostsByAuthor(c *gin.Context) {
	authorID, err := parseUUIDParam(c, "userId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	limit, offset := parsePagination(c)

	posts, err := ctrl.handler.GetPostsByAuthorWithAuthors(c.Request.Context(), authorID, limit, offset)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, posts)
}

// GetPostsByCommunity handles GET /communities/:communityId/posts
func (ctrl *PostQueryController) GetPostsByCommunity(c *gin.Context) {
	communityID, err := parseUUIDParam(c, "communityId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid community ID"})
		return
	}

	limit, offset := parsePagination(c)

	requesterID, err := getUserIDFromContext(c)
	var reqIDPtr *uuid.UUID
	if err == nil {
		// Only set pointer if requester ID was successfully extracted
		reqIDPtr = &requesterID
	}
	// If err != nil, reqIDPtr remains nil → handler treats as unauthenticated

	posts, err := ctrl.handler.GetPostsByCommunityWithAuthors(
		c.Request.Context(),
		communityID,
		reqIDPtr,
		limit,
		offset,
	)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, posts)
}

// SearchPosts handles GET /posts/search
func (ctrl *PostQueryController) SearchPosts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing search query parameter 'q'"})
		return
	}

	limit, offset := parsePagination(c)

	results, err := ctrl.handler.SearchPostsEnriched(c.Request.Context(), query, limit, offset)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, results)
}
