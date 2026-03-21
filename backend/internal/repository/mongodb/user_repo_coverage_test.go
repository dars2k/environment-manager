package mongodb_test

// Additional user repository tests for error paths and filter combinations.

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

func TestUserRepository_List_WithActiveFilter(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("active filter true", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)

		userID := primitive.NewObjectID()
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "test.users", mtest.FirstBatch,
				bson.D{{Key: "_id", Value: userID}, {Key: "username", Value: "u1"}, {Key: "active", Value: true}},
			),
			mtest.CreateCursorResponse(0, "test.users", mtest.NextBatch),
		)

		active := true
		filter := interfaces.ListFilter{Active: &active, Limit: 10, Page: 1}
		users, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
	})

	mt.Run("with pagination", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.users", mtest.FirstBatch))

		filter := interfaces.ListFilter{Limit: 5, Page: 2}
		users, err := repo.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestUserRepository_List_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("find error", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		_, err := repo.List(context.Background(), interfaces.ListFilter{})
		assert.Error(t, err)
	})
}

func TestUserRepository_Update_NotFound(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("not found", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.E{Key: "n", Value: 0},
			bson.E{Key: "nModified", Value: 0},
		))

		user := &entities.User{Username: "testuser"}
		err := repo.Update(context.Background(), primitive.NewObjectID().Hex(), user)
		assert.Error(t, err)
		assert.Equal(t, interfaces.ErrNotFound, err)
	})
}

func TestUserRepository_Update_InvalidID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("invalid id", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)

		user := &entities.User{}
		err := repo.Update(context.Background(), "invalid-id", user)
		assert.Error(t, err)
		assert.Equal(t, interfaces.ErrInvalidID, err)
	})
}

func TestUserRepository_Update_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("update error", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		user := &entities.User{}
		err := repo.Update(context.Background(), primitive.NewObjectID().Hex(), user)
		assert.Error(t, err)
	})
}

func TestUserRepository_UpdateLastLogin_NotFound(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

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

func TestUserRepository_Delete_Error(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("delete error", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    2,
			Message: "command failed",
		}))

		err := repo.Delete(context.Background(), primitive.NewObjectID().Hex())
		assert.Error(t, err)
	})
}

func TestUserRepository_Delete_InvalidID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("invalid id", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)

		err := repo.Delete(context.Background(), "bad-id")
		assert.Error(t, err)
		assert.Equal(t, interfaces.ErrInvalidID, err)
	})
}

func TestUserRepository_Delete_NotFound(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

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
