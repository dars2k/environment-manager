package interfaces

import (
	"context"

	"app-env-manager/internal/domain/entities"
)

// EnvironmentRepository defines the interface for environment data access
type EnvironmentRepository interface {
	Create(ctx context.Context, env *entities.Environment) error
	GetByID(ctx context.Context, id string) (*entities.Environment, error)
	GetByName(ctx context.Context, name string) (*entities.Environment, error)
	List(ctx context.Context, filter ListFilter) ([]*entities.Environment, error)
	Update(ctx context.Context, id string, env *entities.Environment) error
	UpdateStatus(ctx context.Context, id string, status entities.Status) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, filter ListFilter) (int64, error)
}

// ListFilter defines filtering options for listing resources
type ListFilter struct {
	// Common filters
	Page   int
	Limit  int
	Active *bool
	Search string
	
	// Environment specific
	Status     *entities.HealthStatus
	
	// Deprecated - for backward compatibility
	Pagination *Pagination
}

// GetOffset calculates the offset for pagination
func (f *ListFilter) GetOffset() int {
	if f.Page <= 0 {
		f.Page = 1
	}
	return (f.Page - 1) * f.Limit
}

// GetLimit returns the limit with a default value
func (f *ListFilter) GetLimit() int {
	if f.Limit <= 0 {
		return 20 // default limit
	}
	if f.Limit > 100 {
		return 100 // max limit
	}
	return f.Limit
}

// Pagination defines pagination options (deprecated, use ListFilter directly)
type Pagination struct {
	Page  int
	Limit int
}

// GetOffset calculates the offset for pagination
func (p *Pagination) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	return (p.Page - 1) * p.Limit
}

// GetLimit returns the limit with a default value
func (p *Pagination) GetLimit() int {
	if p.Limit <= 0 {
		return 20 // default limit
	}
	if p.Limit > 100 {
		return 100 // max limit
	}
	return p.Limit
}
