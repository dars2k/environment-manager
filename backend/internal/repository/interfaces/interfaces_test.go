package interfaces_test

import (
	"context"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{
			name: "ErrNotFound",
			err:  interfaces.ErrNotFound,
			msg:  "resource not found",
		},
		{
			name: "ErrUserNotFound",
			err:  interfaces.ErrUserNotFound,
			msg:  "user not found",
		},
		{
			name: "ErrInvalidID",
			err:  interfaces.ErrInvalidID,
			msg:  "invalid ID format",
		},
		{
			name: "ErrDuplicate",
			err:  interfaces.ErrDuplicate,
			msg:  "resource already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.err)
			assert.Equal(t, tt.msg, tt.err.Error())
		})
	}
}

func TestPagination(t *testing.T) {
	tests := []struct {
		name          string
		pagination    interfaces.Pagination
		expectedLimit int
		expectedSkip  int
	}{
		{
			name: "Default limit",
			pagination: interfaces.Pagination{
				Page:  1,
				Limit: 0,
			},
			expectedLimit: 20, // Default limit is 20, not 10
			expectedSkip:  0,
		},
		{
			name: "Custom limit",
			pagination: interfaces.Pagination{
				Page:  2,
				Limit: 20,
			},
			expectedLimit: 20,
			expectedSkip:  20,
		},
		{
			name: "Page 3 with limit 15",
			pagination: interfaces.Pagination{
				Page:  3,
				Limit: 15,
			},
			expectedLimit: 15,
			expectedSkip:  30,
		},
		{
			name: "Zero page defaults to 1",
			pagination: interfaces.Pagination{
				Page:  0,
				Limit: 10,
			},
			expectedLimit: 10,
			expectedSkip:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedLimit, tt.pagination.GetLimit())
			assert.Equal(t, tt.expectedSkip, tt.pagination.GetOffset())
		})
	}
}

func TestListFilter(t *testing.T) {
	filter := interfaces.ListFilter{
		Search: "test",
		Pagination: &interfaces.Pagination{
			Page:  2,
			Limit: 25,
		},
	}

	assert.Equal(t, "test", filter.Search)
	assert.NotNil(t, filter.Pagination)
	assert.Equal(t, 2, filter.Pagination.Page)
	assert.Equal(t, 25, filter.Pagination.Limit)
}

func TestLogFilter(t *testing.T) {
	now := time.Now()
	userID := primitive.NewObjectID()
	envID := primitive.NewObjectID()

	filter := interfaces.LogFilter{
		UserID:        &userID,
		EnvironmentID: &envID,
		StartTime:     now,
		EndTime:       now,
		Limit:         10,
		Page:          1,
		Type:          entities.LogTypeAction,
		Level:         entities.LogLevelInfo,
		Action:        entities.ActionTypeCreate,
		Search:        "test",
	}

	assert.Equal(t, userID, *filter.UserID)
	assert.Equal(t, envID, *filter.EnvironmentID)
	assert.Equal(t, now, filter.StartTime)
	assert.Equal(t, now, filter.EndTime)
	assert.Equal(t, 10, filter.Limit)
	assert.Equal(t, 1, filter.Page)
	assert.Equal(t, entities.LogTypeAction, filter.Type)
	assert.Equal(t, entities.LogLevelInfo, filter.Level)
	assert.Equal(t, entities.ActionTypeCreate, filter.Action)
	assert.Equal(t, "test", filter.Search)
}

func TestAuditLogFilter(t *testing.T) {
	now := time.Now()
	
	filter := interfaces.AuditLogFilter{
		EnvironmentID: "env123",
		ActorID:       "user456",
		StartDate:     &now,
		EndDate:       &now,
		Tags:          []string{"deploy", "production"},
		Pagination: &interfaces.Pagination{
			Page:  1,
			Limit: 20,
		},
	}

	assert.Equal(t, "env123", filter.EnvironmentID)
	assert.Equal(t, "user456", filter.ActorID)
	assert.NotNil(t, filter.StartDate)
	assert.NotNil(t, filter.EndDate)
	assert.Len(t, filter.Tags, 2)
	assert.Contains(t, filter.Tags, "deploy")
	assert.Contains(t, filter.Tags, "production")
}

func TestListFilter_GetOffset(t *testing.T) {
	tests := []struct {
		name         string
		filter       interfaces.ListFilter
		expectedOffset int
	}{
		{
			name: "First page",
			filter: interfaces.ListFilter{
				Page:  1,
				Limit: 20,
			},
			expectedOffset: 0,
		},
		{
			name: "Second page",
			filter: interfaces.ListFilter{
				Page:  2,
				Limit: 20,
			},
			expectedOffset: 20,
		},
		{
			name: "Third page with custom limit",
			filter: interfaces.ListFilter{
				Page:  3,
				Limit: 15,
			},
			expectedOffset: 30,
		},
		{
			name: "Zero page defaults to 1",
			filter: interfaces.ListFilter{
				Page:  0,
				Limit: 10,
			},
			expectedOffset: 0,
		},
		{
			name: "Negative page defaults to 1",
			filter: interfaces.ListFilter{
				Page:  -5,
				Limit: 10,
			},
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedOffset, tt.filter.GetOffset())
		})
	}
}

func TestListFilter_GetLimit(t *testing.T) {
	tests := []struct {
		name         string
		filter       interfaces.ListFilter
		expectedLimit int
	}{
		{
			name: "Default limit when zero",
			filter: interfaces.ListFilter{
				Limit: 0,
			},
			expectedLimit: 20,
		},
		{
			name: "Default limit when negative",
			filter: interfaces.ListFilter{
				Limit: -10,
			},
			expectedLimit: 20,
		},
		{
			name: "Custom limit",
			filter: interfaces.ListFilter{
				Limit: 50,
			},
			expectedLimit: 50,
		},
		{
			name: "Max limit enforcement",
			filter: interfaces.ListFilter{
				Limit: 200,
			},
			expectedLimit: 100,
		},
		{
			name: "Edge case: exactly max limit",
			filter: interfaces.ListFilter{
				Limit: 100,
			},
			expectedLimit: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedLimit, tt.filter.GetLimit())
		})
	}
}

func TestListFilter_CompleteScenario(t *testing.T) {
	// Test ListFilter with all fields
	active := true
	healthStatus := entities.HealthStatusHealthy
	
	filter := interfaces.ListFilter{
		Page:   5,
		Limit:  25,
		Active: &active,
		Search: "production",
		Status: &healthStatus,
	}

	// Verify all fields
	assert.Equal(t, 5, filter.Page)
	assert.Equal(t, 25, filter.Limit)
	assert.NotNil(t, filter.Active)
	assert.True(t, *filter.Active)
	assert.Equal(t, "production", filter.Search)
	assert.NotNil(t, filter.Status)
	assert.Equal(t, entities.HealthStatusHealthy, *filter.Status)
	
	// Verify pagination calculation
	assert.Equal(t, 100, filter.GetOffset()) // (5-1) * 25
	assert.Equal(t, 25, filter.GetLimit())
}

func TestPagination_EdgeCases(t *testing.T) {
	// Test max limit enforcement in Pagination
	p := interfaces.Pagination{
		Page:  1,
		Limit: 500, // Exceeds max
	}
	
	assert.Equal(t, 100, p.GetLimit()) // Should cap at 100
	assert.Equal(t, 0, p.GetOffset())   // First page
	
	// Test with very high page number
	p2 := interfaces.Pagination{
		Page:  1000,
		Limit: 10,
	}
	
	assert.Equal(t, 9990, p2.GetOffset()) // (1000-1) * 10
}

func TestRepositoryInterfaceCompliance(t *testing.T) {
	// This test ensures that the interface definitions are properly structured
	// and can be used in type assertions
	
	t.Run("UserRepository interface", func(t *testing.T) {
		var _ interfaces.UserRepository = (*mockUserRepo)(nil)
	})
	
	t.Run("EnvironmentRepository interface", func(t *testing.T) {
		var _ interfaces.EnvironmentRepository = (*mockEnvRepo)(nil)
	})
	
	t.Run("LogRepository interface", func(t *testing.T) {
		var _ interfaces.LogRepository = (*mockLogRepo)(nil)
	})
	
	t.Run("AuditLogRepository interface", func(t *testing.T) {
		var _ interfaces.AuditLogRepository = (*mockAuditRepo)(nil)
	})
}

// Mock implementations for interface compliance testing
type mockUserRepo struct{}

func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) error { return nil }
func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*entities.User, error) { return nil, nil }
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*entities.User, error) { return nil, nil }
func (m *mockUserRepo) List(ctx context.Context, filter interfaces.ListFilter) ([]*entities.User, error) { return nil, nil }
func (m *mockUserRepo) Update(ctx context.Context, id string, user *entities.User) error { return nil }
func (m *mockUserRepo) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *mockUserRepo) Count(ctx context.Context) (int64, error) { return 0, nil }

type mockEnvRepo struct{}

func (m *mockEnvRepo) Create(ctx context.Context, env *entities.Environment) error { return nil }
func (m *mockEnvRepo) GetByID(ctx context.Context, id string) (*entities.Environment, error) { return nil, nil }
func (m *mockEnvRepo) GetByName(ctx context.Context, name string) (*entities.Environment, error) { return nil, nil }
func (m *mockEnvRepo) List(ctx context.Context, filter interfaces.ListFilter) ([]*entities.Environment, error) { return nil, nil }
func (m *mockEnvRepo) Update(ctx context.Context, id string, env *entities.Environment) error { return nil }
func (m *mockEnvRepo) UpdateStatus(ctx context.Context, id string, status entities.Status) error { return nil }
func (m *mockEnvRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *mockEnvRepo) Count(ctx context.Context, filter interfaces.ListFilter) (int64, error) { return 0, nil }

type mockLogRepo struct{}

func (m *mockLogRepo) Create(ctx context.Context, log *entities.Log) error { return nil }
func (m *mockLogRepo) GetByID(ctx context.Context, id primitive.ObjectID) (*entities.Log, error) { return nil, nil }
func (m *mockLogRepo) List(ctx context.Context, filter interfaces.LogFilter) ([]*entities.Log, int64, error) { return nil, 0, nil }
func (m *mockLogRepo) GetEnvironmentLogs(ctx context.Context, envID primitive.ObjectID, limit int) ([]*entities.Log, error) { return nil, nil }
func (m *mockLogRepo) Count(ctx context.Context, filter interfaces.LogFilter) (int64, error) { return 0, nil }
func (m *mockLogRepo) DeleteOld(ctx context.Context, olderThan time.Duration) (int64, error) { return 0, nil }

type mockAuditRepo struct{}

func (m *mockAuditRepo) Create(ctx context.Context, log *entities.AuditLog) error { return nil }
func (m *mockAuditRepo) GetByID(ctx context.Context, id string) (*entities.AuditLog, error) { return nil, nil }
func (m *mockAuditRepo) List(ctx context.Context, filter interfaces.AuditLogFilter) ([]*entities.AuditLog, error) { return nil, nil }
func (m *mockAuditRepo) Count(ctx context.Context, filter interfaces.AuditLogFilter) (int64, error) { return 0, nil }
func (m *mockAuditRepo) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) { return 0, nil }
