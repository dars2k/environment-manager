package mongodb_test

import (
	"context"
	"testing"
	"time"

	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/repository/mongodb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestLogRepository_Coverage(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("List complex filter", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		envID := primitive.NewObjectID()
		userID := primitive.NewObjectID()
		filter := interfaces.LogFilter{
			EnvironmentID: &envID,
			UserID:        &userID,
			Type:          "system",
			Level:         "error",
			Action:        "update",
			StartTime:     time.Now().Add(-time.Hour),
			EndTime:       time.Now(),
			Search:        "test.io",
			Limit:         10,
			Page:          1,
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(1)}}),
			mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch, bson.D{{Key: "_id", Value: primitive.NewObjectID()}}),
			mtest.CreateCursorResponse(0, "test.logs", mtest.NextBatch),
		)

		logs, total, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, logs, 1)
	})

	mt.Run("List count error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 1, Message: "count error"}))
		logs, total, err := repo.List(context.Background(), interfaces.LogFilter{})
		assert.Error(t, err)
		assert.Equal(t, int64(0), total)
		assert.Nil(t, logs)
	})

	mt.Run("List find error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(1)}}),
			mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 1, Message: "find error"}),
		)
		logs, _, err := repo.List(context.Background(), interfaces.LogFilter{})
		assert.Error(t, err)
		assert.Nil(t, logs)
	})

	mt.Run("List decode error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(1)}}),
			mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch, bson.D{{Key: "_id", Value: "invalid"}}),
		)
		logs, _, err := repo.List(context.Background(), interfaces.LogFilter{})
		assert.Error(t, err)
		assert.Nil(t, logs)
	})

	mt.Run("DeleteOld error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 1, Message: "delete error"}))
		count, err := repo.DeleteOld(context.Background(), time.Hour)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	mt.Run("GetEnvironmentLogs error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 1, Message: "find error"}))
		logs, err := repo.GetEnvironmentLogs(context.Background(), primitive.NewObjectID(), 10)
		assert.Error(t, err)
		assert.Nil(t, logs)
	})

	mt.Run("GetEnvironmentLogs decode error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch, bson.D{{Key: "_id", Value: "invalid"}}))
		logs, err := repo.GetEnvironmentLogs(context.Background(), primitive.NewObjectID(), 10)
		assert.Error(t, err)
		assert.Nil(t, logs)
	})

	mt.Run("Count error", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 1, Message: "count error"}))
		count, err := repo.Count(context.Background(), interfaces.LogFilter{})
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	mt.Run("Count complex filter", func(mt *mtest.T) {
		repo := mongodb.NewLogRepository(mt.DB)
		envID := primitive.NewObjectID()
		userID := primitive.NewObjectID()
		filter := interfaces.LogFilter{
			EnvironmentID: &envID,
			UserID:        &userID,
			Type:          "system",
			Level:         "error",
			Action:        "update",
			StartTime:     time.Now().Add(-time.Hour),
			EndTime:       time.Now(),
			Search:        "test",
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.logs", mtest.FirstBatch, bson.D{{Key: "n", Value: int32(5)}}))
		count, err := repo.Count(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})
}
