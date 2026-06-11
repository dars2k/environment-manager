package mongodb_test

// Additional tests to push repository coverage above 90%.
// These tests cover error paths and filter combinations not exercised elsewhere.

import (
	"context"
	"testing"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/repository/mongodb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// ---- LogRepository additional error paths ----

func TestLogRepository_List_CountError(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("count error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		_, _, err := repo.List(context.Background(), interfaces.LogFilter{Limit: 10})
		assert.Error(t, err)
	})
}

func TestLogRepository_List_WithSearchFilter(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("with search filter", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(0)}},
		))
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.logs", mtest.FirstBatch))

		filter := interfaces.LogFilter{
			Search: "error message",
			Limit:  10,
		}
		logs, total, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
		assert.Equal(t, int64(0), total)
	})
}

func TestLogRepository_List_WithAllFilters(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("all filters combined", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)

		envID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(0)}},
		))
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.logs", mtest.FirstBatch))

		now := time.Now()
		filter := interfaces.LogFilter{
			EnvironmentID: &envID,
			UserID:        &userID,
			Type:          entities.LogTypeAction,
			Level:         entities.LogLevelError,
			Action:        entities.ActionTypeRestart,
			StartTime:     now.Add(-1 * time.Hour),
			EndTime:       now,
			Search:        "test",
			Limit:         5,
			Page:          1,
		}

		_, _, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
	})
}

func TestLogRepository_GetEnvironmentLogs_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("find error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		envID := primitive.NewObjectID()
		logs, err := repo.GetEnvironmentLogs(context.Background(), envID, 10)
		assert.Error(t, err)
		assert.Nil(t, logs)
	})
}

func TestLogRepository_Count_WithAllFilters(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("count with all filters", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)

		envID := primitive.NewObjectID()
		userID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(3)}},
		))

		now := time.Now()
		filter := interfaces.LogFilter{
			EnvironmentID: &envID,
			UserID:        &userID,
			Type:          entities.LogTypeAction,
			Level:         entities.LogLevelError,
			Action:        entities.ActionTypeCreate,
			StartTime:     now.Add(-24 * time.Hour),
			EndTime:       now,
			Search:        "search term",
		}

		count, err := repo.Count(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}

func TestLogRepository_Count_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("count error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		_, err := repo.Count(context.Background(), interfaces.LogFilter{})
		assert.Error(t, err)
	})
}

func TestLogRepository_DeleteOld_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("delete error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		_, err := repo.DeleteOld(context.Background(), 30*24*time.Hour)
		assert.Error(t, err)
	})
}

// ---- EnvironmentRepository additional error paths ----

func TestEnvironmentRepository_List_WithPagination(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("with pagination", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)

		envID := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.environments", mtest.FirstBatch,
				bson.D{{Key: "_id", Value: envID}, {Key: "name", Value: "env1"}},
			),
			mtest.CreateCursorResponse(0, "test.environments", mtest.NextBatch),
		)

		filter := interfaces.ListFilter{
			Pagination: &interfaces.Pagination{Page: 2, Limit: 5},
		}
		envs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, envs, 1)
	})
}

func TestEnvironmentRepository_List_WithStatusFilter(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("with healthy status", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.environments", mtest.FirstBatch))

		healthStatus := entities.HealthStatusHealthy
		filter := interfaces.ListFilter{
			Status: &healthStatus,
		}
		envs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, envs)
	})
}

func TestEnvironmentRepository_Count_WithStatusFilter(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("count with status filter", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.environments", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(4)}},
		))

		healthStatus := entities.HealthStatusUnhealthy
		filter := interfaces.ListFilter{
			Status: &healthStatus,
		}
		count, err := repo.Count(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(4), count)
	})
}

func TestEnvironmentRepository_Count_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("count error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		_, err := repo.Count(context.Background(), interfaces.ListFilter{})
		assert.Error(t, err)
	})
}

func TestEnvironmentRepository_GetByName_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("query error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		env, err := repo.GetByName(context.Background(), "test-env")
		assert.Error(t, err)
		assert.Nil(t, env)
	})
}

func TestEnvironmentRepository_Update_DuplicateName(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("duplicate key error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   0,
			Code:    11000,
			Message: "duplicate key error",
		}))

		env := &entities.Environment{Name: "existing"}
		err := repo.Update(context.Background(), primitive.NewObjectID().Hex(), env)
		assert.Error(t, err)
	})
}

func TestEnvironmentRepository_UpdateStatus_Healthy(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("update to healthy sets lastHealthyAt", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		))

		status := entities.Status{
			Health:    entities.HealthStatusHealthy,
			LastCheck: time.Now(),
		}
		err := repo.UpdateStatus(context.Background(), primitive.NewObjectID().Hex(), status)
		assert.NoError(t, err)
	})
}

func TestEnvironmentRepository_UpdateStatus_NotFound(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
			bson.E{Key: "nModified", Value: 0},
		))

		status := entities.Status{Health: entities.HealthStatusUnknown}
		err := repo.UpdateStatus(context.Background(), primitive.NewObjectID().Hex(), status)
		assert.Error(t, err)
	})
}

func TestEnvironmentRepository_Delete_InvalidID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("invalid id", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)

		err := repo.Delete(context.Background(), "invalid-id")
		assert.Error(t, err)
	})
}

// ---- AuditLogRepository additional error paths ----

func TestAuditLogRepository_Create_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("insert error", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		auditLog := &entities.AuditLog{
			ID:        primitive.NewObjectID(),
			Timestamp: time.Now(),
		}
		err := repo.Create(context.Background(), auditLog)
		assert.Error(t, err)
	})
}

func TestAuditLogRepository_List_WithSeverityFilter(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("with severity filter", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.audit_log", mtest.FirstBatch))

		severity := entities.SeverityError
		filter := interfaces.AuditLogFilter{
			Severity: &severity,
			Pagination: &interfaces.Pagination{
				Page:  1,
				Limit: 10,
			},
		}

		logs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
	})
}

func TestAuditLogRepository_Count_WithAllFilters(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("count with all filters", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.audit_log", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(7)}},
		))

		eventType := entities.EventTypeRestart
		severity := entities.SeverityInfo
		now := time.Now()
		start := now.Add(-24 * time.Hour)
		filter := interfaces.AuditLogFilter{
			EnvironmentID: primitive.NewObjectID().Hex(),
			ActorID:       "actor123",
			Type:          &eventType,
			Severity:      &severity,
			StartDate:     &start,
			EndDate:       &now,
		}

		count, err := repo.Count(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(7), count)
	})
}

func TestAuditLogRepository_Count_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("count error", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		_, err := repo.Count(context.Background(), interfaces.AuditLogFilter{})
		assert.Error(t, err)
	})
}

func TestAuditLogRepository_DeleteOlderThan_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("delete error", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		_, err := repo.DeleteOlderThan(context.Background(), time.Now())
		assert.Error(t, err)
	})
}
