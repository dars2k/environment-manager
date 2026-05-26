package mongodb_test

import (
	"context"
	"testing"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"app-env-manager/internal/repository/mongodb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestEnvironmentRepository_Coverage(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("GetByID_InvalidID", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		env, err := repo.GetByID(context.Background(), "invalid-id")
		assert.Error(t, err)
		assert.Nil(t, env)
	})

	mt.Run("Update_InvalidID", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		err := repo.Update(context.Background(), "invalid-id", &entities.Environment{})
		assert.Error(t, err)
	})

	mt.Run("UpdateStatus_InvalidID", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		err := repo.UpdateStatus(context.Background(), "invalid-id", entities.Status{})
		assert.Error(t, err)
	})

	mt.Run("Delete_InvalidID", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		err := repo.Delete(context.Background(), "invalid-id")
		assert.Error(t, err)
	})

	mt.Run("List_WithFilters", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.environments", mtest.FirstBatch))

		active := true
		status := entities.HealthStatusHealthy
		filter := interfaces.ListFilter{
			Page:   1,
			Limit:  10,
			Active: &active,
			Status: &status,
			Search: "test",
		}
		envs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, envs)
	})

	mt.Run("Count_WithFilters", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.environments", mtest.FirstBatch, primitive.D{{Key: "n", Value: 0}}))

		active := false
		status := entities.HealthStatusHealthy
		filter := interfaces.ListFilter{
			Active: &active,
			Status: &status,
			Search: "foo",
		}
		count, err := repo.Count(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}
