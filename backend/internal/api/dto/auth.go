package dto

import (
	"time"
)

// LoginResponseDTO represents a safe login response without sensitive user data
type LoginResponseDTO struct {
	Token     string        `json:"token"`
	User      *UserResponse `json:"user"`
	ExpiresAt time.Time     `json:"expiresAt"`
}
