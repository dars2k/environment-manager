package log

import (
	"context"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles log-related business logic
type Service struct {
	logRepo interfaces.LogRepository
}

// NewService creates a new log service
func NewService(logRepo interfaces.LogRepository) *Service {
	return &Service{
		logRepo: logRepo,
	}
}

// Create creates a new log entry
func (s *Service) Create(ctx context.Context, log *entities.Log) error {
	return s.logRepo.Create(ctx, log)
}

// List retrieves logs with filtering and pagination
func (s *Service) List(ctx context.Context, filter interfaces.LogFilter) ([]*entities.Log, int64, error) {
	return s.logRepo.List(ctx, filter)
}

// GetByID retrieves a log by its ID
func (s *Service) GetByID(ctx context.Context, id string) (*entities.Log, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return s.logRepo.GetByID(ctx, objID)
}

// GetEnvironmentLogs retrieves recent logs for a specific environment
func (s *Service) GetEnvironmentLogs(ctx context.Context, envID string, limit int) ([]*entities.Log, error) {
	objID, err := primitive.ObjectIDFromHex(envID)
	if err != nil {
		return nil, err
	}
	return s.logRepo.GetEnvironmentLogs(ctx, objID, limit)
}

// CleanupOldLogs removes logs older than the specified duration
func (s *Service) CleanupOldLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	return s.logRepo.DeleteOld(ctx, olderThan)
}

// LogEnvironmentAction logs an environment-related action
func (s *Service) LogEnvironmentAction(ctx context.Context, env *entities.Environment, action entities.ActionType, message string, details map[string]interface{}) error {
	log := entities.NewLog(entities.LogTypeAction, entities.LogLevelInfo, message).
		WithEnvironment(env.ID, env.Name).
		WithAction(action).
		WithDetails(details)
	
	return s.Create(ctx, log)
}

// LogHealthCheck logs a health check result
func (s *Service) LogHealthCheck(ctx context.Context, env *entities.Environment, status entities.HealthStatus, message string, details map[string]interface{}) error {
	level := entities.LogLevelSuccess
	if status == entities.HealthStatusUnhealthy {
		level = entities.LogLevelError
	} else if status == entities.HealthStatusUnknown {
		level = entities.LogLevelWarning
	}
	
	log := entities.NewLog(entities.LogTypeHealthCheck, level, message).
		WithEnvironment(env.ID, env.Name).
		WithDetails(details)
	
	return s.Create(ctx, log)
}

// LogAuth logs authentication-related events
func (s *Service) LogAuth(ctx context.Context, userID *primitive.ObjectID, username string, action entities.ActionType, message string, success bool) error {
	level := entities.LogLevelInfo
	if !success {
		level = entities.LogLevelError
	}
	
	log := entities.NewLog(entities.LogTypeAuth, level, message).
		WithAction(action)
	
	if userID != nil && username != "" {
		log = log.WithUser(*userID, username)
	}
	
	return s.Create(ctx, log)
}

// LogError logs system errors
func (s *Service) LogError(ctx context.Context, message string, details map[string]interface{}) error {
	log := entities.NewLog(entities.LogTypeError, entities.LogLevelError, message).
		WithDetails(details)
	
	return s.Create(ctx, log)
}

// LogSystem logs system-level events
func (s *Service) LogSystem(ctx context.Context, message string, details map[string]interface{}) error {
	log := entities.NewLog(entities.LogTypeSystem, entities.LogLevelInfo, message).
		WithDetails(details)
	
	return s.Create(ctx, log)
}

// Count returns the count of logs based on filter
func (s *Service) Count(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
	return s.logRepo.Count(ctx, filter)
}
