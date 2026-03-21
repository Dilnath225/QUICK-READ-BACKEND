package middleware

import (
	"QUICK-READ-SYSTEM/models"
	"net/http"

	"github.com/gin-gonic/gin"
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
