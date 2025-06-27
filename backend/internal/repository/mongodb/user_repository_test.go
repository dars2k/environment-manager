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

func TestUserRepository_Create(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		user := &entities.User{
			ID:           primitive.NewObjectID(),
			Username:     "testuser",
			PasswordHash: "hashedpassword",
			Role:         entities.UserRoleUser,
			Active:       true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.Create(context.Background(), user)
		assert.NoError(t, err)
	})

	mt.Run("duplicate username", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		user := &entities.User{
			ID:       primitive.NewObjectID(),
			Username: "existing",
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   0,
			Code:    11000,
			Message: "duplicate key error",
		}))

		err := repo.Create(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate key error")
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		userID := primitive.NewObjectID()
		expectedUser := &entities.User{
			ID:       userID,
			Username: "testuser",
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: userID},
			{Key: "username", Value: "testuser"},
		}))

		user, err := repo.GetByID(context.Background(), userID.Hex())
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.Username, user.Username)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.users", mtest.FirstBatch))

		user, err := repo.GetByID(context.Background(), primitive.NewObjectID().Hex())
		assert.Error(t, err)
		assert.Equal(t, interfaces.ErrNotFound, err)
		assert.Nil(t, user)
	})

	mt.Run("invalid id", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		user, err := repo.GetByID(context.Background(), "invalid-id")
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		userID := primitive.NewObjectID()
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: userID},
			{Key: "username", Value: "testuser"},
		}))

		user, err := repo.GetByUsername(context.Background(), "testuser")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "testuser", user.Username)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.users", mtest.FirstBatch))

		user, err := repo.GetByUsername(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.Equal(t, interfaces.ErrNotFound, err)
		assert.Nil(t, user)
	})
}


func TestUserRepository_Update(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		userID := primitive.NewObjectID()
		user := &entities.User{
			ID:       userID,
			Username: "updated",
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		))

		err := repo.Update(context.Background(), userID.Hex(), user)
		assert.NoError(t, err)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		user := &entities.User{
			ID: primitive.NewObjectID(),
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
			bson.E{Key: "nModified", Value: 0},
		))

		err := repo.Update(context.Background(), primitive.NewObjectID().Hex(), user)
		assert.Error(t, err)
		assert.Equal(t, interfaces.ErrNotFound, err)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
		))

		err := repo.Delete(context.Background(), primitive.NewObjectID().Hex())
		assert.NoError(t, err)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
		))

		err := repo.Delete(context.Background(), primitive.NewObjectID().Hex())
		assert.Error(t, err)
		assert.Equal(t, interfaces.ErrNotFound, err)
	})
}

func TestUserRepository_List(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		userID1 := primitive.NewObjectID()
		userID2 := primitive.NewObjectID()
		
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch, 
				bson.D{{Key: "_id", Value: userID1}, {Key: "username", Value: "user1"}},
				bson.D{{Key: "_id", Value: userID2}, {Key: "username", Value: "user2"}},
			),
			mtest.CreateCursorResponse(0, "test.users", mtest.NextBatch),
		)

		filter := interfaces.ListFilter{
			Page:  1,
			Limit: 10,
		}

		users, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
	})

	mt.Run("with search", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.users", mtest.FirstBatch))

		filter := interfaces.ListFilter{
			Page:   1,
			Limit:  10,
			Search: "test",
		}

		users, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, users)
	})

	mt.Run("with active filter", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.users", mtest.FirstBatch))

		active := true
		filter := interfaces.ListFilter{
			Page:   1,
			Limit:  10,
			Active: &active,
		}

		users, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestUserRepository_Count(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch,
			bson.D{{Key: "n", Value: int32(5)}},
		))

		count, err := repo.Count(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "nModified", Value: 1},
		))

		err := repo.UpdateLastLogin(context.Background(), primitive.NewObjectID())
		assert.NoError(t, err)
	})

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
			bson.E{Key: "nModified", Value: 0},
		))

		err := repo.UpdateLastLogin(context.Background(), primitive.NewObjectID())
		assert.Error(t, err)
		assert.Equal(t, interfaces.ErrNotFound, err)
	})
}
