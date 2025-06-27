package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	// Test creating a new user
	user, err := NewUser("testuser", "password123")
	
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, UserRoleAdmin, user.Role) // Default role is admin
	assert.True(t, user.Active)
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, "password123", user.PasswordHash) // Password should be hashed
}

func TestUser_CheckPassword(t *testing.T) {
	user, _ := NewUser("testuser", "password123")
	
	// Test correct password
	assert.True(t, user.CheckPassword("password123"))
	
	// Test incorrect password
	assert.False(t, user.CheckPassword("wrongpassword"))
}

func TestUser_UpdatePassword(t *testing.T) {
	user, _ := NewUser("testuser", "password123")
	oldHash := user.PasswordHash
	
	// Update password
	err := user.UpdatePassword("newpassword456")
	assert.NoError(t, err)
	
	// Check that hash changed
	assert.NotEqual(t, oldHash, user.PasswordHash)
	
	// Check new password works
	assert.True(t, user.CheckPassword("newpassword456"))
	assert.False(t, user.CheckPassword("password123"))
}

func TestUser_Permissions(t *testing.T) {
	tests := []struct {
		name               string
		role               UserRole
		canManageUsers     bool
		canEditEnvironments bool
		canViewOnly        bool
	}{
		{
			name:               "Admin user",
			role:               UserRoleAdmin,
			canManageUsers:     true,
			canEditEnvironments: true,
			canViewOnly:        false,
		},
		{
			name:               "Regular user",
			role:               UserRoleUser,
			canManageUsers:     false,
			canEditEnvironments: true,
			canViewOnly:        false,
		},
		{
			name:               "Viewer user",
			role:               UserRoleViewer,
			canManageUsers:     false,
			canEditEnvironments: false,
			canViewOnly:        true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, _ := NewUser("testuser", "password123")
			user.Role = tt.role // Set the role for testing
			
			assert.Equal(t, tt.canManageUsers, user.CanManageUsers())
			assert.Equal(t, tt.canEditEnvironments, user.CanEditEnvironments())
			assert.Equal(t, tt.canViewOnly, user.CanViewOnly())
		})
	}
}
