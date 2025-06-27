package handlers

import (
	"fmt"
	"net/http"

	"app-env-manager/internal/api/dto"
	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *user.Service
	logger      *logrus.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *user.Service, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// List retrieves all users
func (h *UserHandler) List(c *gin.Context) {
	filter := interfaces.ListFilter{
		Page:  1,
		Limit: 100,
	}

	// Parse query parameters
	if page := c.Query("page"); page != "" {
		var p int
		if _, err := fmt.Sscanf(page, "%d", &p); err == nil && p > 0 {
			filter.Page = p
		}
	}

	if limit := c.Query("limit"); limit != "" {
		var l int
		if _, err := fmt.Sscanf(limit, "%d", &l); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}

	users, err := h.userService.ListUsers(c.Request.Context(), filter)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list users")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve users",
		})
		return
	}

	// Convert to DTOs to filter out sensitive data
	userResponses := dto.ToUserResponses(users)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"users": userResponses,
		},
	})
}

// Get retrieves a specific user
func (h *UserHandler) Get(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		if err == interfaces.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user",
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

// Create creates a new user
func (h *UserHandler) Create(c *gin.Context) {
	var req entities.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get current user from context
	currentUser := h.getCurrentUser(c)

	user, err := h.userService.CreateUser(c.Request.Context(), req, currentUser)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Convert to DTO to filter out sensitive data
	userResponse := dto.ToUserResponse(user)

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"user": userResponse,
		},
	})
}

// Update updates a user
func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("id")
	
	var req entities.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get current user from context
	currentUser := h.getCurrentUser(c)

	user, err := h.userService.UpdateUser(c.Request.Context(), userID, req, currentUser)
	if err != nil {
		if err == interfaces.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to update user")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
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

// Delete deletes a user
func (h *UserHandler) Delete(c *gin.Context) {
	userID := c.Param("id")

	// Get current user from context
	currentUser := h.getCurrentUser(c)

	err := h.userService.DeleteUser(c.Request.Context(), userID, currentUser)
	if err != nil {
		if err == interfaces.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to delete user")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message": "User deleted successfully",
		},
	})
}

// ChangePassword changes the current user's password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("userID")
	
	var req entities.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID.(string), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to change password")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message": "Password changed successfully",
		},
	})
}

// ResetPassword resets a user's password (admin only)
func (h *UserHandler) ResetPassword(c *gin.Context) {
	userID := c.Param("id")
	
	var req entities.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get current user from context
	currentUser := h.getCurrentUser(c)

	err := h.userService.ResetPassword(c.Request.Context(), userID, req, currentUser)
	if err != nil {
		if err == interfaces.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		h.logger.WithError(err).Error("Failed to reset password")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"message": "Password reset successfully",
		},
	})
}

// getCurrentUser gets the current authenticated user from context
func (h *UserHandler) getCurrentUser(c *gin.Context) *entities.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*entities.User)
}
