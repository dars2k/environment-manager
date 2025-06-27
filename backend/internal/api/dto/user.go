package dto

import (
	"time"

	"app-env-manager/internal/domain/entities"
)

// UserResponse represents a safe user response without sensitive data
type UserResponse struct {
	ID           string                  `json:"id"`
	Username     string                  `json:"username"`
	Role         entities.UserRole       `json:"role"`
	Active       bool                    `json:"active"`
	CreatedAt    time.Time               `json:"createdAt"`
	UpdatedAt    time.Time               `json:"updatedAt"`
	LastLoginAt  *time.Time              `json:"lastLoginAt,omitempty"`
	Metadata     map[string]interface{}  `json:"metadata,omitempty"`
}

// ToUserResponse converts a User entity to a safe UserResponse DTO
func ToUserResponse(user *entities.User) *UserResponse {
	if user == nil {
		return nil
	}

	return &UserResponse{
		ID:          user.ID.Hex(),
		Username:    user.Username,
		Role:        user.Role,
		Active:      user.Active,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
		Metadata:    user.Metadata,
	}
}

// ToUserResponses converts a slice of User entities to UserResponse DTOs
func ToUserResponses(users []*entities.User) []*UserResponse {
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToUserResponse(user)
	}
	return responses
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users []*UserResponse `json:"users"`
}

// SingleUserResponse represents the response for a single user
type SingleUserResponse struct {
	User *UserResponse `json:"user"`
}
