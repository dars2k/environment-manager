package handlers

import (
	"net/http"

	"app-env-manager/internal/api/dto"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/service/auth"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *auth.Service
	logger      *logrus.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req entities.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).WithField("ip", c.ClientIP()).Warn("Invalid login request format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Log the login attempt with client IP
	h.logger.WithFields(map[string]interface{}{
		"username": req.Username,
		"ip":       c.ClientIP(),
	}).Info("Login attempt")

	response, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"username": req.Username,
			"ip":       c.ClientIP(),
		}).Warn("Login failed")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Log successful login
	h.logger.WithFields(map[string]interface{}{
		"username": req.Username,
		"userId":   response.User.ID.Hex(),
		"ip":       c.ClientIP(),
	}).Info("Login successful")

	// Convert to safe DTO
	safeResponse := &dto.LoginResponseDTO{
		Token:     response.Token,
		User:      dto.ToUserResponse(response.User),
		ExpiresAt: response.ExpiresAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": safeResponse,
	})
}

// GetCurrentUser gets the authenticated user's information
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	user, err := h.authService.GetUserFromContext(c.Request.Context(), userID.(string))
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user information",
		})
		return
	}

	// Convert to DTO to filter out sensitive data
	userResponse := dto.ToUserResponse(user)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user": userResponse,
		},
	})
}

// Logout handles user logout (placeholder for frontend to clear token)
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message": "Logged out successfully",
		},
	})
}
