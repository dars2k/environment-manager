package mongodb_test

import (
	"context"
	"testing"

	"app-env-manager/internal/repository/mongodb"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestUserRepository_Coverage(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("GetByUsername database error", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 1, Message: "find error"}))
		_, err := repo.GetByUsername(context.Background(), "test")
		assert.Error(t, err)
	})

	mt.Run("UpdateLastLogin database error", func(mt *mtest.T) {
		repo := mongodb.NewUserRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 1, Message: "update error"}))
		err := repo.UpdateLastLogin(context.Background(), primitive.NewObjectID())
		assert.Error(t, err)
	})
}
