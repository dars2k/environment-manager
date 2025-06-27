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

func TestLogRepository_Create(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		envID := primitive.NewObjectID()
		log := &entities.Log{
			ID:              primitive.NewObjectID(),
			Timestamp:       time.Now(),
			Type:            entities.LogTypeAction,
			Level:           entities.LogLevelInfo,
			Message:         "Test log message",
			EnvironmentID:   &envID,
			EnvironmentName: "test-env",
			Action:          entities.ActionTypeCreate,
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Create(context.Background(), log)
		assert.NoError(t, err)
		assert.NotEqual(t, primitive.NilObjectID, log.ID)
	})
}

func TestLogRepository_List(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		logID1 := primitive.NewObjectID()
		logID2 := primitive.NewObjectID()
		
		// Mock count response
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(2)}},
		))
		
		// Mock find response
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch, 
				bson.D{{Key: "_id", Value: logID1}, {Key: "message", Value: "log1"}},
				bson.D{{Key: "_id", Value: logID2}, {Key: "message", Value: "log2"}},
			),
			mtest.CreateCursorResponse(0, "test.logs", mtest.NextBatch),
		)

		filter := interfaces.LogFilter{
			Limit: 10,
		}

		logs, total, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, logs, 2)
		assert.Equal(t, int64(2), total)
	})

	mt.Run("with environment filter", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		// Mock count response
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(0)}},
		))
		
		// Mock find response
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.logs", mtest.FirstBatch))

		envID := primitive.NewObjectID()
		filter := interfaces.LogFilter{
			EnvironmentID: &envID,
			Limit:        10,
		}

		logs, total, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
		assert.Equal(t, int64(0), total)
	})

	mt.Run("with level filter", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		// Mock count response
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(0)}},
		))
		
		// Mock find response
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.logs", mtest.FirstBatch))

		filter := interfaces.LogFilter{
			Level: entities.LogLevelError,
			Limit: 10,
		}

		logs, total, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
		assert.Equal(t, int64(0), total)
	})

	mt.Run("with date range", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		// Mock count response
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(0)}},
		))
		
		// Mock find response
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.logs", mtest.FirstBatch))

		now := time.Now()
		filter := interfaces.LogFilter{
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
			Limit:     10,
		}

		logs, total, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
		assert.Equal(t, int64(0), total)
	})

	mt.Run("with pagination", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		// Mock count response
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(0)}},
		))
		
		// Mock find response
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.logs", mtest.FirstBatch))

		filter := interfaces.LogFilter{
			Page:  2,
			Limit: 10,
		}

		logs, total, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, logs)
		assert.Equal(t, int64(0), total)
	})
}

func TestLogRepository_GetByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		logID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: logID},
			{Key: "message", Value: "Test log"},
		}))

		log, err := repo.GetByID(context.Background(), logID)
		assert.NoError(t, err)
		assert.NotNil(t, log)
		assert.Equal(t, "Test log", log.Message)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.logs", mtest.FirstBatch))

		log, err := repo.GetByID(context.Background(), primitive.NewObjectID())
		assert.Error(t, err)
		assert.Equal(t, interfaces.ErrNotFound, err)
		assert.Nil(t, log)
	})
}

func TestLogRepository_DeleteOld(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 100},
		))

		count, err := repo.DeleteOld(context.Background(), 30*24*time.Hour)
		assert.NoError(t, err)
		assert.Equal(t, int64(100), count)
	})
}

func TestLogRepository_GetEnvironmentLogs(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		envID := primitive.NewObjectID()
		logID := primitive.NewObjectID()
		
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch, 
				bson.D{{Key: "_id", Value: logID}, {Key: "message", Value: "env log"}},
			),
			mtest.CreateCursorResponse(0, "test.logs", mtest.NextBatch),
		)

		logs, err := repo.GetEnvironmentLogs(context.Background(), envID, 10)
		assert.NoError(t, err)
		assert.Len(t, logs, 1)
	})
}

func TestLogRepository_Count(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(50)}},
		))

		count, err := repo.Count(context.Background(), interfaces.LogFilter{})
		assert.NoError(t, err)
		assert.Equal(t, int64(50), count)
	})
}
