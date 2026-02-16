// post_service/internal/http/router.go
package router

import (
	"net/http"
	"time"

	auth "my-go-backend/post_service/internal/authentication"
	"my-go-backend/post_service/internal/config"
	"my-go-backend/post_service/internal/contextkeys"
	"my-go-backend/post_service/internal/http/middleware"
	postapi "my-go-backend/post_service/src/posts/api/controllers"

	"github.com/gin-gonic/gin"
)

type Router struct {
	postCommandController *postapi.PostCommandController
	postQueryController   *postapi.PostQueryController
}

func NewRouter(
	postCommandController *postapi.PostCommandController,
	postQueryController *postapi.PostQueryController,
) *Router {
	return &Router{
		postCommandController: postCommandController,
		postQueryController:   postQueryController,
	}
}

func SetupRouter(cfg *config.Config, r *Router) *gin.Engine {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(SetupCORS(cfg.FrontendURL))
	verifier := auth.NewVerifier(
		cfg.AuthPublicKeyURL,
		time.Duration(cfg.AuthPublicKeyTTL)*time.Second,
	)

	api := router.Group("/api/v1")
	// Keep existing auth test routes
	api.GET("/auth/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api.GET(
		"/auth/test",
		middleware.AuthMiddleware(verifier),
		func(c *gin.Context) {
			userID := c.Request.Context().Value(contextkeys.UserIDKey)
			c.JSON(http.StatusOK, gin.H{
				"user_id_from_jwt": userID,
			})
		},
	)

	// Public post routes (read-only)
	publicPosts := api.Group("/posts")
	{
		publicPosts.GET("/:id", r.postQueryController.GetPost)
		publicPosts.GET("/search", r.postQueryController.SearchPosts)
	}

	// User-scoped public routes
	api.GET("/users/:userId/posts", r.postQueryController.GetPostsByAuthor)

	// Community-scoped public routes
	api.GET("/communities/:communityId/posts", r.postQueryController.GetPostsByCommunity)

	// Protected post mutation routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(verifier))
	{
		protected.POST("/posts", r.postCommandController.CreatePost)
		protected.PUT("/posts/:id", r.postCommandController.UpdatePost)
		protected.PATCH("/posts/:id/visibility", r.postCommandController.TogglePostVisibility)
		protected.POST("/posts/:id/like", r.postCommandController.LikePost)
		protected.POST("/posts/:id/unlike", r.postCommandController.UnlikePost)
		protected.POST("/posts/:id/comment", r.postCommandController.AddCommentToPost)
		protected.DELETE("/posts/:id/comment", r.postCommandController.RemoveCommentFromPost)
		protected.DELETE("/posts/:id", r.postCommandController.DeletePost)
	}

	return router
}
