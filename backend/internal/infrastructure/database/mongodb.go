package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB represents a MongoDB connection
type MongoDB struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDB creates a new MongoDB connection
func NewMongoDB(uri, databaseName string, maxConnections int, timeout time.Duration) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Configure client options
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(uint64(maxConnections)).
		SetMaxConnIdleTime(5 * time.Minute).
		SetServerSelectionTimeout(timeout)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get database
	database := client.Database(databaseName)

	return &MongoDB{
		client:   client,
		database: database,
	}, nil
}

// Close closes the MongoDB connection
func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// Database returns the database instance
func (m *MongoDB) Database() *mongo.Database {
	return m.database
}

// Collection returns a collection from the database
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.database.Collection(name)
}

// CreateIndexes creates indexes for all collections
func (m *MongoDB) CreateIndexes(ctx context.Context) error {
	// Environment indexes
	envCollection := m.Collection("environments")
	envIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"name": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"status.health": 1},
		},
		{
			Keys: map[string]interface{}{"timestamps.lastCheck": 1},
		},
	}
	if _, err := envCollection.Indexes().CreateMany(ctx, envIndexes); err != nil {
		return fmt.Errorf("failed to create environment indexes: %w", err)
	}

	// Audit log indexes
	auditCollection := m.Collection("audit_log")
	auditIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"timestamp":     -1,
				"environmentId": 1,
			},
		},
		{
			Keys: map[string]interface{}{"type": 1},
		},
		{
			Keys: map[string]interface{}{"actor.id": 1},
		},
		{
			Keys: map[string]interface{}{"tags": 1},
		},
	}
	if _, err := auditCollection.Indexes().CreateMany(ctx, auditIndexes); err != nil {
		return fmt.Errorf("failed to create audit log indexes: %w", err)
	}

	// Credentials indexes
	credCollection := m.Collection("credentials")
	credIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"name": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"usage.environmentId": 1},
		},
		{
			Keys: map[string]interface{}{"timestamps.expiresAt": 1},
		},
	}
	if _, err := credCollection.Indexes().CreateMany(ctx, credIndexes); err != nil {
		return fmt.Errorf("failed to create credentials indexes: %w", err)
	}

	return nil
}

// Transaction executes a function within a transaction
func (m *MongoDB) Transaction(ctx context.Context, fn func(sessionContext mongo.SessionContext) error) error {
	session, err := m.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	return mongo.WithSession(ctx, session, func(sessionContext mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		if err := fn(sessionContext); err != nil {
			if abortErr := session.AbortTransaction(sessionContext); abortErr != nil {
				return fmt.Errorf("failed to abort transaction: %v (original error: %w)", abortErr, err)
			}
			return err
		}

		if err := session.CommitTransaction(sessionContext); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})
}
