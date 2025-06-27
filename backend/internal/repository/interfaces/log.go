package interfaces

import (
	"context"
	"time"

	"app-env-manager/internal/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LogFilter represents filter options for listing logs
type LogFilter struct {
	EnvironmentID *primitive.ObjectID
	UserID        *primitive.ObjectID
	Type          entities.LogType
	Level         entities.LogLevel
	Action        entities.ActionType
	StartTime     time.Time
	EndTime       time.Time
	Search        string
	Page          int
	Limit         int
}

// LogRepository defines the interface for log storage operations
type LogRepository interface {
	Create(ctx context.Context, log *entities.Log) error
	List(ctx context.Context, filter LogFilter) ([]*entities.Log, int64, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*entities.Log, error)
	DeleteOld(ctx context.Context, olderThan time.Duration) (int64, error)
	GetEnvironmentLogs(ctx context.Context, envID primitive.ObjectID, limit int) ([]*entities.Log, error)
	Count(ctx context.Context, filter LogFilter) (int64, error)
}
