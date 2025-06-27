package user_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/log"
	"app-env-manager/internal/service/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockUserRepository is a mock implementation
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}


func (m *MockUserRepository) Update(ctx context.Context, id string, user *entities.User) error {
	args := m.Called(ctx, id, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, filter interfaces.ListFilter) ([]*entities.User, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockLogRepository is a mock implementation for log repository
type MockLogRepository struct {
	mock.Mock
}

func (m *MockLogRepository) Create(ctx context.Context, log *entities.Log) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockLogRepository) List(ctx context.Context, filter interfaces.LogFilter) ([]*entities.Log, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*entities.Log), args.Get(1).(int64), args.Error(2)
}

func (m *MockLogRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entities.Log, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Log), args.Error(1)
}

func (m *MockLogRepository) DeleteOld(ctx context.Context, olderThan time.Duration) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockLogRepository) GetEnvironmentLogs(ctx context.Context, envID primitive.ObjectID, limit int) ([]*entities.Log, error) {
	args := m.Called(ctx, envID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Log), args.Error(1)
}

func (m *MockLogRepository) Count(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func TestNewService(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)
	assert.NotNil(t, service)
}

func TestService_CreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	// Create a real log service with a mocked repository to avoid nil pointer
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	service := user.NewService(mockRepo, logService)

	ctx := context.Background()
	req := entities.CreateUserRequest{
		Username: "testuser",
		Password: "password123",
	}
	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	// Check username doesn't exist
	mockRepo.On("GetByUsername", ctx, req.Username).Return(nil, interfaces.ErrUserNotFound)
	// Create user
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.User")).Return(nil)
	// Mock log service repository call
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	createdUser, err := service.CreateUser(ctx, req, adminUser)
	assert.NoError(t, err)
	assert.NotNil(t, createdUser)
	assert.Equal(t, req.Username, createdUser.Username)
	assert.Equal(t, entities.UserRoleAdmin, createdUser.Role) // Default role is admin
	assert.True(t, createdUser.Active)
	mockRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestService_CreateUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	req := entities.CreateUserRequest{
		Username: "testuser",
		Password: "password123",
	}

	mockRepo.On("GetByUsername", ctx, req.Username).Return(nil, interfaces.ErrUserNotFound)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.User")).Return(fmt.Errorf("database error"))

	createdUser, err := service.CreateUser(ctx, req, nil)
	assert.Error(t, err)
	assert.Nil(t, createdUser)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestService_CreateUser_DuplicateUsername(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	req := entities.CreateUserRequest{
		Username: "testuser",
		Password: "password123",
	}

	existingUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: req.Username,
	}

	mockRepo.On("GetByUsername", ctx, req.Username).Return(existingUser, nil)

	createdUser, err := service.CreateUser(ctx, req, nil)
	assert.Error(t, err)
	assert.Nil(t, createdUser)
	assert.Contains(t, err.Error(), "username already exists")
	mockRepo.AssertExpectations(t)
}


func TestService_GetUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	expectedUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
	}

	mockRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

	user, err := service.GetUser(ctx, userID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Username, user.Username)
	mockRepo.AssertExpectations(t)
}

func TestService_GetUser_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()

	mockRepo.On("GetByID", ctx, userID).Return(nil, interfaces.ErrUserNotFound)

	user, err := service.GetUser(ctx, userID)
	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}


func TestService_UpdateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	service := user.NewService(mockRepo, logService)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	role := entities.UserRoleAdmin
	active := false

	req := entities.UpdateUserRequest{
		Role:   &role,
		Active: &active,
	}

	existingUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
		Role:     entities.UserRoleUser,
		Active:   true,
	}

	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, userID, mock.AnythingOfType("*entities.User")).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	updatedUser, err := service.UpdateUser(ctx, userID, req, adminUser)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, role, updatedUser.Role)
	assert.Equal(t, active, updatedUser.Active)
	mockRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestService_UpdateUser_NonAdmin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	role := entities.UserRoleAdmin

	req := entities.UpdateUserRequest{
		Role: &role,
	}

	nonAdminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleUser,
	}

	updatedUser, err := service.UpdateUser(ctx, userID, req, nonAdminUser)
	assert.Error(t, err)
	assert.Nil(t, updatedUser)
	assert.Contains(t, err.Error(), "unauthorized")
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateUser_NoPermissions(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	active := false

	req := entities.UpdateUserRequest{
		Active: &active,
	}

	existingUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
		Active:   true,
	}

	nonAdminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleUser,
	}

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil).Maybe()

	updatedUser, err := service.UpdateUser(ctx, userID, req, nonAdminUser)
	assert.Error(t, err)
	assert.Nil(t, updatedUser)
	assert.Contains(t, err.Error(), "unauthorized")
}


func TestService_UpdateUser_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	role := entities.UserRoleUser

	req := entities.UpdateUserRequest{
		Role: &role,
	}

	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	mockRepo.On("GetByID", ctx, userID).Return(nil, interfaces.ErrUserNotFound)

	updatedUser, err := service.UpdateUser(ctx, userID, req, adminUser)
	assert.Error(t, err)
	assert.Nil(t, updatedUser)
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	role := entities.UserRoleAdmin

	req := entities.UpdateUserRequest{
		Role: &role,
	}

	existingUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
		Role:     entities.UserRoleUser,
	}

	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, userID, mock.AnythingOfType("*entities.User")).Return(fmt.Errorf("database error"))

	updatedUser, err := service.UpdateUser(ctx, userID, req, adminUser)
	assert.Error(t, err)
	assert.Nil(t, updatedUser)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestService_ChangePassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	service := user.NewService(mockRepo, logService)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	req := entities.ChangePasswordRequest{
		CurrentPassword: "oldpass",
		NewPassword:     "newpass",
	}

	existingUser, _ := entities.NewUser("testuser", req.CurrentPassword)
	existingUser.ID = primitive.NewObjectID()

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, userID, mock.AnythingOfType("*entities.User")).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.ChangePassword(ctx, userID, req)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestService_ChangePassword_WrongCurrent(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	req := entities.ChangePasswordRequest{
		CurrentPassword: "wrongpass",
		NewPassword:     "newpass",
	}

	existingUser, _ := entities.NewUser("testuser", "correctpass")
	existingUser.ID = primitive.NewObjectID()

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)

	err := service.ChangePassword(ctx, userID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "current password is incorrect")
	mockRepo.AssertExpectations(t)
}

func TestService_ChangePassword_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	req := entities.ChangePasswordRequest{
		CurrentPassword: "oldpass",
		NewPassword:     "newpass",
	}

	mockRepo.On("GetByID", ctx, userID).Return(nil, interfaces.ErrUserNotFound)

	err := service.ChangePassword(ctx, userID, req)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_ChangePassword_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	req := entities.ChangePasswordRequest{
		CurrentPassword: "oldpass",
		NewPassword:     "newpass",
	}

	existingUser, _ := entities.NewUser("testuser", req.CurrentPassword)
	existingUser.ID = primitive.NewObjectID()

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, userID, mock.AnythingOfType("*entities.User")).Return(fmt.Errorf("database error"))

	err := service.ChangePassword(ctx, userID, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestService_ResetPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	service := user.NewService(mockRepo, logService)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	req := entities.ResetPasswordRequest{
		NewPassword: "newpass",
	}

	existingUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
	}

	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, userID, mock.AnythingOfType("*entities.User")).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.ResetPassword(ctx, userID, req, adminUser)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestService_ResetPassword_NonAdmin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	req := entities.ResetPasswordRequest{
		NewPassword: "newpass",
	}

	nonAdminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleUser,
	}

	err := service.ResetPassword(ctx, userID, req, nonAdminUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	mockRepo.AssertExpectations(t)
}

func TestService_ResetPassword_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	req := entities.ResetPasswordRequest{
		NewPassword: "newpass",
	}

	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	mockRepo.On("GetByID", ctx, userID).Return(nil, interfaces.ErrUserNotFound)

	err := service.ResetPassword(ctx, userID, req, adminUser)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	service := user.NewService(mockRepo, logService)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()

	existingUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
	}

	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Delete", ctx, userID).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.DeleteUser(ctx, userID, adminUser)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}

func TestService_DeleteUser_NonAdmin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()

	nonAdminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleUser,
	}

	err := service.DeleteUser(ctx, userID, nonAdminUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteUser_CannotDeleteSelf(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID()

	adminUser := &entities.User{
		ID:   userID, // Same ID as the user to delete
		Role: entities.UserRoleAdmin,
	}

	err := service.DeleteUser(ctx, userID.Hex(), adminUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete your own account")
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteUser_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()

	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	mockRepo.On("GetByID", ctx, userID).Return(nil, interfaces.ErrUserNotFound)

	err := service.DeleteUser(ctx, userID, adminUser)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteUser_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()

	existingUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
	}

	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Delete", ctx, userID).Return(fmt.Errorf("database error"))

	err := service.DeleteUser(ctx, userID, adminUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestService_ListUsers(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	filter := interfaces.ListFilter{
		Pagination: &interfaces.Pagination{
			Page:  1,
			Limit: 10,
		},
	}

	users := []*entities.User{
		{
			ID:       primitive.NewObjectID(),
			Username: "user1",
		},
		{
			ID:       primitive.NewObjectID(),
			Username: "user2",
		},
	}

	mockRepo.On("List", ctx, filter).Return(users, nil)

	result, err := service.ListUsers(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestService_ListUsers_Error(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	filter := interfaces.ListFilter{
		Pagination: &interfaces.Pagination{
			Page:  1,
			Limit: 10,
		},
	}

	mockRepo.On("List", ctx, filter).Return(nil, fmt.Errorf("database error"))

	result, err := service.ListUsers(ctx, filter)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestService_GetUserByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID()
	expectedUser := &entities.User{
		ID:       userID,
		Username: "testuser",
	}

	mockRepo.On("GetByID", ctx, userID.Hex()).Return(expectedUser, nil)

	user, err := service.GetUserByID(ctx, userID.Hex())
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Username, user.Username)
	mockRepo.AssertExpectations(t)
}

func TestService_GetUserByID_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()

	mockRepo.On("GetByID", ctx, userID).Return(nil, interfaces.ErrUserNotFound)

	user, err := service.GetUserByID(ctx, userID)
	assert.Error(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}


func TestService_ResetPassword_RepositoryError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	req := entities.ResetPasswordRequest{
		NewPassword: "newpass",
	}

	existingUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
	}

	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, userID, mock.AnythingOfType("*entities.User")).Return(fmt.Errorf("database error"))

	err := service.ResetPassword(ctx, userID, req, adminUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
	mockRepo.AssertExpectations(t)
}

func TestService_GetUserByID_InvalidID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	invalidID := "invalid-id"

	user, err := service.GetUserByID(ctx, invalidID)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid user ID")
}

func TestService_CreateUser_NoPermissions(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := user.NewService(mockRepo, nil)

	ctx := context.Background()
	req := entities.CreateUserRequest{
		Username: "testuser",
		Password: "password123",
	}

	nonAdminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleUser,
	}

	createdUser, err := service.CreateUser(ctx, req, nonAdminUser)
	assert.Error(t, err)
	assert.Nil(t, createdUser)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestService_UpdateUser_WithMetadata(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockLogRepo := new(MockLogRepository)
	logService := log.NewService(mockLogRepo)
	service := user.NewService(mockRepo, logService)

	ctx := context.Background()
	userID := primitive.NewObjectID().Hex()
	metadata := map[string]interface{}{
		"department": "Engineering",
		"location":   "Remote",
	}

	req := entities.UpdateUserRequest{
		Metadata: metadata,
	}

	existingUser := &entities.User{
		ID:       primitive.NewObjectID(),
		Username: "testuser",
		Role:     entities.UserRoleUser,
	}

	adminUser := &entities.User{
		ID:   primitive.NewObjectID(),
		Role: entities.UserRoleAdmin,
	}

	mockRepo.On("GetByID", ctx, userID).Return(existingUser, nil)
	mockRepo.On("Update", ctx, userID, mock.AnythingOfType("*entities.User")).Return(nil)
	mockLogRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	updatedUser, err := service.UpdateUser(ctx, userID, req, adminUser)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, metadata, updatedUser.Metadata)
	mockRepo.AssertExpectations(t)
	mockLogRepo.AssertExpectations(t)
}
