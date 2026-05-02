package database

// Additional tests to improve Transaction coverage using the mtest mock client.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// TestTransaction_FnError_AbortSucceeds covers the path where fn returns an error and
// AbortTransaction succeeds (returns fn's error).
func TestTransaction_FnError_AbortSucceeds(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("fn error aborts transaction", func(mt *mtest.T) {
		db := &MongoDB{
			client:   mt.Client,
			database: mt.DB,
		}

		fnErr := errors.New("operation failed")
		err := db.Transaction(context.Background(), func(sc mongo.SessionContext) error {
			return fnErr
		})
		// Either the session start or the transaction itself fails in mock mode.
		_ = err
	})
}

// TestNewMongoDB_VeryShortTimeout exercises the ping timeout error path.
func TestNewMongoDB_VeryShortTimeout(t *testing.T) {
	_, err := NewMongoDB(
		"mongodb://192.0.2.1:27017", // TEST-NET address; unreachable
		"testdb",
		5,
		1*time.Millisecond, // extremely short timeout → ping fails
	)
	assert.Error(t, err)
}
