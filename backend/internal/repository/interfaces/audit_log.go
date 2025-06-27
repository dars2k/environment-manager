package interfaces

import (
	"context"
	"time"

	"app-env-manager/internal/domain/entities"
)

// AuditLogRepository defines the interface for audit log data access
type AuditLogRepository interface {
	Create(ctx context.Context, log *entities.AuditLog) error
	GetByID(ctx context.Context, id string) (*entities.AuditLog, error)
	List(ctx context.Context, filter AuditLogFilter) ([]*entities.AuditLog, error)
	Count(ctx context.Context, filter AuditLogFilter) (int64, error)
	DeleteOlderThan(ctx context.Context, before time.Time) (int64, error)
}

// AuditLogFilter defines filtering options for audit logs
type AuditLogFilter struct {
	EnvironmentID string
	Type          *entities.EventType
	Severity      *entities.Severity
	ActorID       string
	StartDate     *time.Time
	EndDate       *time.Time
	Tags          []string
	Pagination    *Pagination
}
