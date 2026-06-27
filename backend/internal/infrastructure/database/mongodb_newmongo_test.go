package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestNewMongoDB_SuccessPath(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("returns populated MongoDB struct on success", func(mt *mtest.T) {
		// Add mock response for the Ping call inside NewMongoDB
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		orig := mongoConnect
		defer func() { mongoConnect = orig }()

		mongoConnect = func(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
			return mt.Client, nil
		}

		db, err := NewMongoDB("mongodb://localhost:27017", "testdb", 10, time.Second)
		require.NoError(t, err)
		assert.NotNil(t, db)
		if db != nil {
			assert.NotNil(t, db.Database())
			assert.NotNil(t, db.Collection("test"))
		}
	})
}
