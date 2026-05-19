package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoDB_Methods(t *testing.T) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB not available for testing")
		return
	}
	defer client.Disconnect(ctx)

	database := client.Database("test_db")
	
	mongodb := &MongoDB{
		client:   client,
		database: database,
	}

	t.Run("Database", func(t *testing.T) {
		db := mongodb.Database()
		assert.NotNil(t, db)
		assert.Equal(t, database, db)
	})

	t.Run("Collection", func(t *testing.T) {
		collection := mongodb.Collection("test_collection")
		assert.NotNil(t, collection)
		assert.Equal(t, "test_collection", collection.Name())
	})

	t.Run("Close", func(t *testing.T) {
		testClient, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
		testDB := &MongoDB{
			client:   testClient,
			database: testClient.Database("test_db"),
		}
		
		err := testDB.Close(ctx)
		assert.NoError(t, err)
	})
}
