package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"log"
)

func RequireAuth(c *gin.Context) {
	// 1. Get the token from the "Authorization" header
	tokenString := c.GetHeader("Authorization")
	log.Printf("[Auth] Request: %s %s, Token length: %d", c.Request.Method, c.Request.URL.Path, len(tokenString))

	// Check if token format is "Bearer <token>"
	if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing or invalid"})
		return
	}

	// Remove "Bearer " prefix
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// 2. Decode/Validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// MUST match the secret key used in login
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// 3. Check if token is expired
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			return
		}

		// 4. Find the user in Database
		var user models.User
		config.DB.First(&user, claims["user_id"])

		if user.ID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// 5. Attach user to request (so Controllers can use it)
		c.Set("user", user)
		c.Next()
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
	}
}
