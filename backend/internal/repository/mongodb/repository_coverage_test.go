package mongodb_test

import (
	"context"
	"testing"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/repository/mongodb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestEnvironmentRepository_Coverage(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Delete invalid id", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		err := repo.Delete(context.Background(), "invalid-id")
		assert.Error(t, err)
	})

	mt.Run("List error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "find error",
		}))

		filter := interfaces.ListFilter{Page: 1, Limit: 10}
		envs, err := repo.List(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, envs)
	})

	mt.Run("List decode error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.environments", mtest.FirstBatch, bson.D{{Key: "_id", Value: "invalid"}}))

		filter := interfaces.ListFilter{Page: 1, Limit: 10}
		envs, err := repo.List(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, envs)
	})

	mt.Run("Count error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "count error",
		}))

		count, err := repo.Count(context.Background(), interfaces.ListFilter{})
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	mt.Run("Update duplicate name", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		envID := primitive.NewObjectID()
		env := &entities.Environment{Name: "existing"}
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   0,
			Code:    11000,
			Message: "duplicate key error",
		}))
		err := repo.Update(context.Background(), envID.Hex(), env)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	mt.Run("Update database error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		envID := primitive.NewObjectID()
		env := &entities.Environment{Name: "new"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "update error",
		}))
		err := repo.Update(context.Background(), envID.Hex(), env)
		assert.Error(t, err)
	})

	mt.Run("UpdateStatus database error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		envID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "update status error",
		}))
		err := repo.UpdateStatus(context.Background(), envID.Hex(), entities.Status{})
		assert.Error(t, err)
	})

	mt.Run("UpdateStatus healthy", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		envID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))
		status := entities.Status{Health: entities.HealthStatusHealthy}
		err := repo.UpdateStatus(context.Background(), envID.Hex(), status)
		assert.NoError(t, err)
	})

	mt.Run("GetByName invalid input", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		env, err := repo.GetByName(context.Background(), "")
		assert.Error(t, err)
		assert.Nil(t, env)
	})

	mt.Run("GetByName database error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "find one error",
		}))
		env, err := repo.GetByName(context.Background(), "test")
		assert.Error(t, err)
		assert.Nil(t, env)
	})

	mt.Run("Count status filter error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		status := entities.HealthStatus("")
		filter := interfaces.ListFilter{Status: &status}
		count, err := repo.Count(context.Background(), filter)
		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
	})

	mt.Run("List status filter error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		status := entities.HealthStatus("")
		filter := interfaces.ListFilter{Status: &status}
		envs, err := repo.List(context.Background(), filter)
		assert.Error(t, err)
		assert.Nil(t, envs)
	})

	mt.Run("Create error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		env := &entities.Environment{Name: "test"}
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "insert error",
		}))
		err := repo.Create(context.Background(), env)
		assert.Error(t, err)
	})

	mt.Run("GetByID database error", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    1,
			Message: "find one error",
		}))
		env, err := repo.GetByID(context.Background(), primitive.NewObjectID().Hex())
		assert.Error(t, err)
		assert.Nil(t, env)
	})
}
