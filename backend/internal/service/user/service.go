package user

import (
	"context"
	"fmt"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles user business logic
type Service struct {
	userRepo   interfaces.UserRepository
	logService *log.Service
}

// NewService creates a new user service
func NewService(userRepo interfaces.UserRepository, logService *log.Service) *Service {
	return &Service{
		userRepo:   userRepo,
		logService: logService,
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, req entities.CreateUserRequest, createdBy *entities.User) (*entities.User, error) {
	// Check if user can manage users
	if createdBy != nil && !createdBy.CanManageUsers() {
		return nil, fmt.Errorf("unauthorized: insufficient permissions")
	}

	// Check for duplicate username
	if existing, _ := s.userRepo.GetByUsername(ctx, req.Username); existing != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Create user
	user, err := entities.NewUser(req.Username, req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Save to repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Log action
	if createdBy != nil {
		_ = s.logService.LogAuth(ctx, &createdBy.ID, createdBy.Username, entities.ActionTypeCreate,
			fmt.Sprintf("Created user: %s", user.Username), true)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, id string) (*entities.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// ListUsers lists all users
func (s *Service) ListUsers(ctx context.Context, filter interfaces.ListFilter) ([]*entities.User, error) {
	return s.userRepo.List(ctx, filter)
}

// UpdateUser updates a user
func (s *Service) UpdateUser(ctx context.Context, id string, req entities.UpdateUserRequest, updatedBy *entities.User) (*entities.User, error) {
	// Check if user can manage users
	if updatedBy != nil && !updatedBy.CanManageUsers() {
		return nil, fmt.Errorf("unauthorized: insufficient permissions")
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Role != nil {
		user.Role = *req.Role
	}

	if req.Active != nil {
		user.Active = *req.Active
	}

	if req.Metadata != nil {
		user.Metadata = req.Metadata
	}

	// Update in repository
	if err := s.userRepo.Update(ctx, id, user); err != nil {
		return nil, err
	}

	// Log action
	if updatedBy != nil {
		_ = s.logService.LogAuth(ctx, &updatedBy.ID, updatedBy.Username, entities.ActionTypeUpdate,
			fmt.Sprintf("Updated user: %s", user.Username), true)
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *Service) DeleteUser(ctx context.Context, id string, deletedBy *entities.User) error {
	// Check if user can manage users
	if deletedBy != nil && !deletedBy.CanManageUsers() {
		return fmt.Errorf("unauthorized: insufficient permissions")
	}

	// Prevent self-deletion
	if deletedBy != nil && deletedBy.ID.Hex() == id {
		return fmt.Errorf("cannot delete your own account")
	}

	// Get user for logging
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from repository
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Log action
	if deletedBy != nil {
		_ = s.logService.LogAuth(ctx, &deletedBy.ID, deletedBy.Username, entities.ActionTypeDelete,
			fmt.Sprintf("Deleted user: %s", user.Username), true)
	}

	return nil
}

// ChangePassword changes a user's password
func (s *Service) ChangePassword(ctx context.Context, id string, req entities.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify current password
	if !user.CheckPassword(req.CurrentPassword) {
		return fmt.Errorf("current password is incorrect")
	}

	// Update password
	if err := user.UpdatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Save to repository
	if err := s.userRepo.Update(ctx, id, user); err != nil {
		return err
	}

	// Log action
	_ = s.logService.LogAuth(ctx, &user.ID, user.Username, entities.ActionTypeUpdate,
		"Password changed", true)

	return nil
}

// ResetPassword resets a user's password (admin action)
func (s *Service) ResetPassword(ctx context.Context, id string, req entities.ResetPasswordRequest, resetBy *entities.User) error {
	// Check if user can manage users
	if resetBy != nil && !resetBy.CanManageUsers() {
		return fmt.Errorf("unauthorized: insufficient permissions")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update password
	if err := user.UpdatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Save to repository
	if err := s.userRepo.Update(ctx, id, user); err != nil {
		return err
	}

	// Log action
	if resetBy != nil {
		_ = s.logService.LogAuth(ctx, &resetBy.ID, resetBy.Username, entities.ActionTypeUpdate,
			fmt.Sprintf("Reset password for user: %s", user.Username), true)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(ctx context.Context, id string) (*entities.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}
	
	user, err := s.userRepo.GetByID(ctx, objID.Hex())
	if err != nil {
		return nil, err
	}
	
	return user, nil
}
