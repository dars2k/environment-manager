package entities

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleUser   UserRole = "user"
	UserRoleViewer UserRole = "viewer"
)

// User represents a system user
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	Role         UserRole           `bson:"role" json:"role"`
	Active       bool               `bson:"active" json:"active"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
	LastLoginAt  *time.Time         `bson:"lastLoginAt,omitempty" json:"lastLoginAt,omitempty"`
	Metadata     map[string]any     `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// NewUser creates a new user
func NewUser(username, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           primitive.NewObjectID(),
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         UserRoleAdmin, // Default to admin role
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Metadata:     make(map[string]any),
	}, nil
}

// CheckPassword verifies the user's password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdatePassword updates the user's password
func (u *User) UpdatePassword(newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	u.UpdatedAt = time.Now()
	return nil
}

// CanManageUsers checks if user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == UserRoleAdmin
}

// CanEditEnvironments checks if user can edit environments
func (u *User) CanEditEnvironments() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleUser
}

// CanViewOnly checks if user has view-only access
func (u *User) CanViewOnly() bool {
	return u.Role == UserRoleViewer
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	User      *User     `json:"user"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Username string   `json:"username" validate:"required,alphanum,min=3,max=50"`
	Password string   `json:"password" validate:"required,min=8"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Role     *UserRole `json:"role,omitempty" validate:"omitempty,oneof=admin user viewer"`
	Active   *bool     `json:"active,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}
