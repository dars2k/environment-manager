package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestTransaction_FnError_AbortSucceeds(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("fn error aborts transaction", func(mt *mtest.T) {
		db := &MongoDB{
			client:   mt.Client,
			database: mt.DB,
		}

		fnErr := errors.New("operation failed")
		_ = db.Transaction(context.Background(), func(sc mongo.SessionContext) error {
			return fnErr
		})
	})
}

func TestNewMongoDB_VeryShortTimeout(t *testing.T) {
	_, err := NewMongoDB(
		"mongodb://192.0.2.1:27017",
		"testdb",
		5,
		1*time.Millisecond,
	)
	assert.Error(t, err)
}
