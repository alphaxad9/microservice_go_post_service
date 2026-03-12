package middleware

import (
	"context"
	"net/http"
	"strings"

	auth "github.com/alphaxad9/my-go-backend/post_service/internal/authentication"
	"github.com/alphaxad9/my-go-backend/post_service/internal/contextkeys"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(verifier *auth.Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		if !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}

		token := strings.TrimPrefix(h, "Bearer ")
		claims, err := verifier.Verify(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		ctx := context.WithValue(
			c.Request.Context(),
			contextkeys.UserIDKey,
			claims["user_id"],
		)

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
