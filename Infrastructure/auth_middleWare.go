package infrastructure

import (
	"net/http"
	"strings"

	"task-manager/Domain"

	"github.com/gin-gonic/gin"
)

// Context keys set by AuthMiddleware and read by AdminOnly / handlers.
const (
	ContextUserID   = "user_id"
	ContextUsername = "username"
	ContextRole     = "role"
)

// AuthMiddleware validates the JWT sent in the Authorization header
// ("Authorization: Bearer <token>"). Requests without a valid, unexpired
// token are rejected with 401 before they ever reach a handler.
//
// It depends only on domain.JWTService (an interface), not on the concrete
// JWTService in this package, so it can be tested or swapped independently.
func AuthMiddleware(jwtService domain.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header must be in the form 'Bearer <token>'"})
			return
		}

		claims, err := jwtService.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

// AdminOnly rejects the request with 403 unless AuthMiddleware has already
// established that the caller has the admin role. It must be chained after
// AuthMiddleware.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextRole)
		if !exists || role != domain.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin privileges are required for this action"})
			return
		}
		c.Next()
	}
}
