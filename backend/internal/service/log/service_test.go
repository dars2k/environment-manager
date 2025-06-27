package log_test

import (
	"context"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/service/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockLogRepository is a mock implementation
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
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)
	assert.NotNil(t, service)
}

func TestService_Create(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	logEntry := &entities.Log{
		ID:        primitive.NewObjectID(),
		Type:      entities.LogTypeSystem,
		Level:     entities.LogLevelInfo,
		Message:   "Test log",
		Timestamp: time.Now(),
	}

	mockRepo.On("Create", ctx, logEntry).Return(nil)

	err := service.Create(ctx, logEntry)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_List(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	filter := interfaces.LogFilter{
		Limit: 10,
		Page:  1,
	}

	logs := []*entities.Log{
		{
			ID:      primitive.NewObjectID(),
			Type:    entities.LogTypeSystem,
			Level:   entities.LogLevelInfo,
			Message: "Test log 1",
		},
		{
			ID:      primitive.NewObjectID(),
			Type:    entities.LogTypeSystem,
			Level:   entities.LogLevelInfo,
			Message: "Test log 2",
		},
	}

	mockRepo.On("List", ctx, filter).Return(logs, int64(2), nil)

	result, count, err := service.List(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), count)
	mockRepo.AssertExpectations(t)
}

func TestService_LogAuth(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	userID := primitive.NewObjectID()
	username := "testuser"
	action := entities.ActionTypeLogin
	details := "Login successful"
	success := true

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogAuth(ctx, &userID, username, action, details, success)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_LogAuth_Failed(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	username := "testuser"
	action := entities.ActionTypeLogin
	details := "Invalid password"
	success := false

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogAuth(ctx, nil, username, action, details, success)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_LogSystem(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	message := "System initialized"
	metadata := map[string]interface{}{
		"version": "1.0.0",
		"module":  "auth",
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogSystem(ctx, message, metadata)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_LogEnvironmentAction(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
	}
	actionType := entities.ActionTypeCreate
	message := "Environment created"
	metadata := map[string]interface{}{
		"user": "admin",
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogEnvironmentAction(ctx, env, actionType, message, metadata)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_LogHealthCheck(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
	}
	status := entities.HealthStatusHealthy
	message := "Health check passed"
	metadata := map[string]interface{}{
		"responseTime": 150,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogHealthCheck(ctx, env, status, message, metadata)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_LogHealthCheck_Failed(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
	}
	status := entities.HealthStatusUnhealthy
	message := "Health check failed"
	metadata := map[string]interface{}{
		"error": "timeout",
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogHealthCheck(ctx, env, status, message, metadata)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetEnvironmentLogs(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	envID := primitive.NewObjectID()
	limit := 50

	logs := []*entities.Log{
		{
			ID:            primitive.NewObjectID(),
			Type:          entities.LogTypeAction,
			Level:         entities.LogLevelInfo,
			Message:       "Test log 1",
			EnvironmentID: &envID,
		},
	}

	mockRepo.On("GetEnvironmentLogs", ctx, envID, limit).Return(logs, nil)

	result, err := service.GetEnvironmentLogs(ctx, envID.Hex(), limit)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	mockRepo.AssertExpectations(t)
}

func TestService_CleanupOldLogs(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	retention := 30 * 24 * time.Hour

	mockRepo.On("DeleteOld", ctx, retention).Return(int64(100), nil)

	deleted, err := service.CleanupOldLogs(ctx, retention)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), deleted)
	mockRepo.AssertExpectations(t)
}

func TestService_GetByID(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	logID := primitive.NewObjectID()
	expectedLog := &entities.Log{
		ID:      logID,
		Type:    entities.LogTypeSystem,
		Level:   entities.LogLevelInfo,
		Message: "Test log",
	}

	mockRepo.On("GetByID", ctx, logID).Return(expectedLog, nil)

	result, err := service.GetByID(ctx, logID.Hex())
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedLog.Message, result.Message)
	mockRepo.AssertExpectations(t)
}

func TestService_GetByID_InvalidID(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	invalidID := "invalid-id"

	result, err := service.GetByID(ctx, invalidID)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestService_GetByID_NotFound(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	logID := primitive.NewObjectID()

	mockRepo.On("GetByID", ctx, logID).Return(nil, assert.AnError)

	result, err := service.GetByID(ctx, logID.Hex())
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestService_LogError(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	message := "System error occurred"
	metadata := map[string]interface{}{
		"error":     "database connection failed",
		"component": "auth",
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogError(ctx, message, metadata)
	assert.NoError(t, err)
	
	// Verify the log was created with correct type and level
	mockRepo.AssertCalled(t, "Create", ctx, mock.MatchedBy(func(log *entities.Log) bool {
		return log.Type == entities.LogTypeError &&
			log.Level == entities.LogLevelError &&
			log.Message == message
	}))
	mockRepo.AssertExpectations(t)
}

func TestService_Count(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	filter := interfaces.LogFilter{
		Type:  entities.LogTypeAuth,
		Level: entities.LogLevelInfo,
	}

	mockRepo.On("Count", ctx, filter).Return(int64(42), nil)

	count, err := service.Count(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), count)
	mockRepo.AssertExpectations(t)
}

func TestService_Count_Error(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	filter := interfaces.LogFilter{
		Type: entities.LogTypeSystem,
	}

	mockRepo.On("Count", ctx, filter).Return(int64(0), assert.AnError)

	count, err := service.Count(ctx, filter)
	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
	mockRepo.AssertExpectations(t)
}

func TestService_Create_Error(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	logEntry := &entities.Log{
		ID:        primitive.NewObjectID(),
		Type:      entities.LogTypeSystem,
		Level:     entities.LogLevelInfo,
		Message:   "Test log",
		Timestamp: time.Now(),
	}

	mockRepo.On("Create", ctx, logEntry).Return(assert.AnError)

	err := service.Create(ctx, logEntry)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_List_Error(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	filter := interfaces.LogFilter{
		Limit: 10,
		Page:  1,
	}

	mockRepo.On("List", ctx, filter).Return(nil, int64(0), assert.AnError)

	result, count, err := service.List(ctx, filter)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), count)
	mockRepo.AssertExpectations(t)
}

func TestService_GetEnvironmentLogs_InvalidID(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	invalidEnvID := "invalid-id"
	limit := 50

	result, err := service.GetEnvironmentLogs(ctx, invalidEnvID, limit)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestService_GetEnvironmentLogs_Error(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	envID := primitive.NewObjectID()
	limit := 50

	mockRepo.On("GetEnvironmentLogs", ctx, envID, limit).Return(nil, assert.AnError)

	result, err := service.GetEnvironmentLogs(ctx, envID.Hex(), limit)
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestService_CleanupOldLogs_Error(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	retention := 30 * 24 * time.Hour

	mockRepo.On("DeleteOld", ctx, retention).Return(int64(0), assert.AnError)

	deleted, err := service.CleanupOldLogs(ctx, retention)
	assert.Error(t, err)
	assert.Equal(t, int64(0), deleted)
	mockRepo.AssertExpectations(t)
}

func TestService_LogAuth_WithoutUserID(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	username := ""
	action := entities.ActionTypeLogin
	details := "Anonymous action"
	success := true

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogAuth(ctx, nil, username, action, details, success)
	assert.NoError(t, err)
	
	// Verify the log was created without user information
	mockRepo.AssertCalled(t, "Create", ctx, mock.MatchedBy(func(log *entities.Log) bool {
		return log.Type == entities.LogTypeAuth &&
			log.UserID == nil &&
			log.Username == ""
	}))
	mockRepo.AssertExpectations(t)
}

func TestService_LogHealthCheck_Unknown(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
	}
	status := entities.HealthStatusUnknown
	message := "Health check status unknown"
	metadata := map[string]interface{}{
		"reason": "timeout",
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogHealthCheck(ctx, env, status, message, metadata)
	assert.NoError(t, err)
	
	// Verify warning level for unknown status
	mockRepo.AssertCalled(t, "Create", ctx, mock.MatchedBy(func(log *entities.Log) bool {
		return log.Type == entities.LogTypeHealthCheck &&
			log.Level == entities.LogLevelWarning
	}))
	mockRepo.AssertExpectations(t)
}

func TestService_LogHealthCheck_Healthy(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	env := &entities.Environment{
		ID:   primitive.NewObjectID(),
		Name: "Test Environment",
	}
	status := entities.HealthStatusHealthy
	message := "Health check passed"
	metadata := map[string]interface{}{
		"responseTime": 100,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogHealthCheck(ctx, env, status, message, metadata)
	assert.NoError(t, err)
	
	// Verify success level for healthy status
	mockRepo.AssertCalled(t, "Create", ctx, mock.MatchedBy(func(log *entities.Log) bool {
		return log.Type == entities.LogTypeHealthCheck &&
			log.Level == entities.LogLevelSuccess
	}))
	mockRepo.AssertExpectations(t)
}

func TestService_LogAuth_WithUserDetails(t *testing.T) {
	mockRepo := new(MockLogRepository)
	service := log.NewService(mockRepo)

	ctx := context.Background()
	userID := primitive.NewObjectID()
	username := "testuser"
	action := entities.ActionTypeUpdate
	details := "User profile updated"
	success := true

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Log")).Return(nil)

	err := service.LogAuth(ctx, &userID, username, action, details, success)
	assert.NoError(t, err)
	
	// Verify the log was created with user information
	mockRepo.AssertCalled(t, "Create", ctx, mock.MatchedBy(func(log *entities.Log) bool {
		return log.Type == entities.LogTypeAuth &&
			log.UserID != nil &&
			*log.UserID == userID &&
			log.Username == username
	}))
	mockRepo.AssertExpectations(t)
}
