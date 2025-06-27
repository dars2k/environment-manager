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

func TestEnvironmentRepository_Create(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		env := &entities.Environment{
			ID:   primitive.NewObjectID(),
			Name: "test-env",
			Target: entities.Target{
				Host: "test.example.com",
				Port: 22,
			},
			Credentials: entities.CredentialRef{
				Type:     "password",
				Username: "testuser",
			},
			HealthCheck: entities.HealthCheckConfig{
				Enabled:  true,
				Endpoint: "/health",
				Interval: 60,
			},
			Status: entities.Status{
				Health:    entities.HealthStatusUnknown,
				LastCheck: time.Now(),
			},
			Timestamps: entities.Timestamps{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Create(context.Background(), env)
		assert.NoError(t, err)
		assert.NotEqual(t, primitive.NilObjectID, env.ID)
	})

	mt.Run("duplicate name", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		env := &entities.Environment{
			Name: "existing",
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   0,
			Code:    11000,
			Message: "duplicate key error",
		}))

		err := repo.Create(context.Background(), env)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestEnvironmentRepository_GetByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		envID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.environments", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: envID},
			{Key: "name", Value: "test-env"},
			{Key: "target", Value: bson.D{
				{Key: "host", Value: "test.example.com"},
				{Key: "port", Value: 22},
			}},
		}))

		env, err := repo.GetByID(context.Background(), envID.Hex())
		assert.NoError(t, err)
		assert.NotNil(t, env)
		assert.Equal(t, "test-env", env.Name)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.environments", mtest.FirstBatch))

		env, err := repo.GetByID(context.Background(), primitive.NewObjectID().Hex())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Environment not found")
		assert.Nil(t, env)
	})

	mt.Run("invalid id", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		env, err := repo.GetByID(context.Background(), "invalid-id")
		assert.Error(t, err)
		assert.Nil(t, env)
	})
}

func TestEnvironmentRepository_GetByName(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		envID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.environments", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: envID},
			{Key: "name", Value: "test-env"},
		}))

		env, err := repo.GetByName(context.Background(), "test-env")
		assert.NoError(t, err)
		assert.NotNil(t, env)
		assert.Equal(t, "test-env", env.Name)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.environments", mtest.FirstBatch))

		env, err := repo.GetByName(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Environment not found")
		assert.Nil(t, env)
	})
}

func TestEnvironmentRepository_Update(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		envID := primitive.NewObjectID()
		env := &entities.Environment{
			ID:   envID,
			Name: "updated-env",
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		))

		err := repo.Update(context.Background(), envID.Hex(), env)
		assert.NoError(t, err)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		env := &entities.Environment{
			ID: primitive.NewObjectID(),
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
			bson.E{Key: "nModified", Value: 0},
		))

		err := repo.Update(context.Background(), primitive.NewObjectID().Hex(), env)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Environment not found")
	})

	mt.Run("invalid id", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		env := &entities.Environment{}
		err := repo.Update(context.Background(), "invalid-id", env)
		assert.Error(t, err)
	})
}

func TestEnvironmentRepository_UpdateStatus(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		status := entities.Status{
			Health:    entities.HealthStatusHealthy,
			LastCheck: time.Now(),
			Message:   "Service is healthy",
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		))

		err := repo.UpdateStatus(context.Background(), primitive.NewObjectID().Hex(), status)
		assert.NoError(t, err)
	})

	mt.Run("invalid id", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		status := entities.Status{}
		err := repo.UpdateStatus(context.Background(), "invalid-id", status)
		assert.Error(t, err)
	})
}

func TestEnvironmentRepository_Delete(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
		))

		err := repo.Delete(context.Background(), primitive.NewObjectID().Hex())
		assert.NoError(t, err)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
		))

		err := repo.Delete(context.Background(), primitive.NewObjectID().Hex())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Environment not found")
	})
}

func TestEnvironmentRepository_List(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		envID1 := primitive.NewObjectID()
		envID2 := primitive.NewObjectID()
		
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.environments", mtest.FirstBatch, 
				bson.D{{Key: "_id", Value: envID1}, {Key: "name", Value: "Environment 1"}},
				bson.D{{Key: "_id", Value: envID2}, {Key: "name", Value: "Environment 2"}},
			),
			mtest.CreateCursorResponse(0, "test.environments", mtest.NextBatch),
		)

		filter := interfaces.ListFilter{
			Page:  1,
			Limit: 10,
		}

		envs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, envs, 2)
	})

	mt.Run("with search", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.environments", mtest.FirstBatch))

		filter := interfaces.ListFilter{
			Page:   1,
			Limit:  10,
			Search: "test",
		}

		envs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, envs)
	})

	mt.Run("with status filter", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.environments", mtest.FirstBatch))

		healthStatus := entities.HealthStatusHealthy
		filter := interfaces.ListFilter{
			Page:   1,
			Limit:  10,
			Status: &healthStatus,
		}

		envs, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, envs)
	})
}

func TestEnvironmentRepository_Count(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.environments", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(10)}},
		))

		count, err := repo.Count(context.Background(), interfaces.ListFilter{})
		assert.NoError(t, err)
		assert.Equal(t, int64(10), count)
	})

	mt.Run("with search", func(mt *mtest.T) {
		repo := mongodb.NewEnvironmentRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.environments", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(3)}},
		))

		filter := interfaces.ListFilter{
			Search: "test",
		}

		count, err := repo.Count(context.Background(), filter)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}
