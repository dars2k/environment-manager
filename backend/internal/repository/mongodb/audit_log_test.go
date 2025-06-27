package mongodb_test

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

func TestAuditLogRepository_Create(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		auditLog := &entities.AuditLog{
			ID:              primitive.NewObjectID(),
			Timestamp:       time.Now(),
			EnvironmentID:   primitive.NewObjectID(),
			EnvironmentName: "test-env",
			Type:            entities.EventTypeRestart,
			Severity:        entities.SeverityInfo,
			Actor: entities.Actor{
				Type: "user",
				ID:   "user123",
				Name: "Test User",
			},
			Action: entities.Action{
				Operation: "create_environment",
				Status:    "completed",
			},
			Payload: entities.Payload{
				Metadata: map[string]interface{}{
					"key": "value",
				},
			},
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Create(context.Background(), auditLog)
		assert.NoError(t, err)
		assert.NotEqual(t, primitive.NilObjectID, auditLog.ID)
	})
}

func TestAuditLogRepository_GetByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		auditLogID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.audit_logs", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: auditLogID},
			{Key: "environmentName", Value: "test-env"},
			{Key: "type", Value: "create"},
		}))

		auditLog, err := repo.GetByID(context.Background(), auditLogID.Hex())
		assert.NoError(t, err)
		assert.NotNil(t, auditLog)
		assert.Equal(t, "test-env", auditLog.EnvironmentName)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.audit_logs", mtest.FirstBatch))

		auditLog, err := repo.GetByID(context.Background(), primitive.NewObjectID().Hex())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "audit log not found")
		assert.Nil(t, auditLog)
	})

	mt.Run("invalid id", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		auditLog, err := repo.GetByID(context.Background(), "invalid-id")
		assert.Error(t, err)
		assert.Nil(t, auditLog)
	})
}

func TestAuditLogRepository_List(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		logID1 := primitive.NewObjectID()
		logID2 := primitive.NewObjectID()
		
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.audit_logs", mtest.FirstBatch, 
				bson.D{{Key: "_id", Value: logID1}, {Key: "environmentName", Value: "env1"}},
				bson.D{{Key: "_id", Value: logID2}, {Key: "environmentName", Value: "env2"}},
			),
			mtest.CreateCursorResponse(0, "test.audit_logs", mtest.NextBatch),
		)

		filter := interfaces.AuditLogFilter{
			Pagination: &interfaces.Pagination{
				Page:  1,
				Limit: 10,
			},
		}

		logs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, logs, 2)
	})

	mt.Run("with environment filter", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.audit_logs", mtest.FirstBatch))

		filter := interfaces.AuditLogFilter{
			EnvironmentID: primitive.NewObjectID().Hex(),
			Pagination: &interfaces.Pagination{
				Page:  1,
				Limit: 10,
			},
		}

		logs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
	})

	mt.Run("with actor filter", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.audit_logs", mtest.FirstBatch))

		filter := interfaces.AuditLogFilter{
			ActorID: "user123",
			Pagination: &interfaces.Pagination{
				Page:  1,
				Limit: 10,
			},
		}

		logs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
	})

	mt.Run("with event type filter", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.audit_logs", mtest.FirstBatch))

		eventType := entities.EventTypeRestart
		filter := interfaces.AuditLogFilter{
			Type: &eventType,
			Pagination: &interfaces.Pagination{
				Page:  1,
				Limit: 10,
			},
		}

		logs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
	})

	mt.Run("with date range", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.audit_logs", mtest.FirstBatch))

		now := time.Now()
		startDate := now.Add(-24 * time.Hour)
		endDate := now
		filter := interfaces.AuditLogFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
			Pagination: &interfaces.Pagination{
				Page:  1,
				Limit: 10,
			},
		}

		logs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
	})

	mt.Run("with pagination", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.audit_logs", mtest.FirstBatch))

		filter := interfaces.AuditLogFilter{
			Pagination: &interfaces.Pagination{
				Page:  2,
				Limit: 10,
			},
		}

		logs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
	})
}

func TestAuditLogRepository_Count(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.audit_logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(25)}},
		))

		count, err := repo.Count(context.Background(), interfaces.AuditLogFilter{})
		assert.NoError(t, err)
		assert.Equal(t, int64(25), count)
	})

	mt.Run("with filters", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.audit_logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(5)}},
		))

		filter := interfaces.AuditLogFilter{
			EnvironmentID: primitive.NewObjectID().Hex(),
		}

		count, err := repo.Count(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})
}

func TestAuditLogRepository_DeleteOlderThan(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewAuditLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 50},
		))

		before := time.Now().Add(-90 * 24 * time.Hour)
		count, err := repo.DeleteOlderThan(context.Background(), before)
		assert.NoError(t, err)
		assert.Equal(t, int64(50), count)
	})
}
