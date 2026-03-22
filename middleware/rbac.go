package middleware

import (
	"QUICK-READ-BACKEND/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"log"
)

// RequireRole returns a middleware that checks if the authenticated user
// has one of the allowed roles. Must be used AFTER RequireAuth middleware.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		user := userVal.(models.User)
		log.Printf("[RBAC] User: %d, Role: %s, Required: %v", user.ID, user.Role, allowedRoles)
		for _, role := range allowedRoles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Access denied. Required role(s): " + joinRoles(allowedRoles),
		})
	}
}

func joinRoles(roles []string) string {
	result := ""
	for i, r := range roles {
		if i > 0 {
			result += ", "
		}
		result += r
	}
	return result
}
