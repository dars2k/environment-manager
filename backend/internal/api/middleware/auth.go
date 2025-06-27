package middleware

import (
	"net/http"
	"strings"

	"app-env-manager/internal/service/auth"
	"app-env-manager/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(authService *auth.Service, userService *user.Service, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Validate token
		token := parts[1]
		claims, err := authService.ValidateToken(token)
		if err != nil {
			logger.WithError(err).Warn("Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Get user from database
		user, err := userService.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil {
			logger.WithError(err).Error("Failed to get user from token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// Check if user is active
		if !user.Active {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Account is inactive",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("userID", user.ID.Hex())
		c.Set("user", user)
		c.Set("userRole", user.Role)

		c.Next()
	}
}

// RequireRole ensures the user has the required role
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			c.Abort()
			return
		}

		role := string(userRole.(string))
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
		c.Abort()
	}
}

// OptionalAuth middleware that doesn't require authentication but sets user context if token is provided
func OptionalAuth(authService *auth.Service, userService *user.Service, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		// Validate token
		token := parts[1]
		claims, err := authService.ValidateToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Get user from database
		user, err := userService.GetUserByID(c.Request.Context(), claims.UserID)
		if err != nil || !user.Active {
			c.Next()
			return
		}

		// Set user context
		c.Set("userID", user.ID.Hex())
		c.Set("user", user)
		c.Set("userRole", user.Role)

		c.Next()
	}
}
