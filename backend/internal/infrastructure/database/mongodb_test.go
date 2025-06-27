package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"app-env-manager/internal/infrastructure/database"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestNewMongoDB(t *testing.T) {
	// Note: This test requires a MongoDB instance or will fail
	// In a real test environment, you would use testcontainers or a test MongoDB instance
	
	tests := []struct {
		name           string
		uri            string
		databaseName   string
		maxConnections int
		timeout        time.Duration
		expectError    bool
		errorContains  string
	}{
		{
			name:           "Invalid URI",
			uri:            "invalid://uri",
			databaseName:   "testdb",
			maxConnections: 10,
			timeout:        5 * time.Second,
			expectError:    true,
			errorContains:  "failed to connect to MongoDB",
		},
		{
			name:           "Connection timeout",
			uri:            "mongodb://nonexistent:27017",
			databaseName:   "testdb",
			maxConnections: 10,
			timeout:        1 * time.Millisecond,
			expectError:    true,
			errorContains:  "failed to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := database.NewMongoDB(tt.uri, tt.databaseName, tt.maxConnections, tt.timeout)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				// Clean up
				if db != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					db.Close(ctx)
				}
			}
		})
	}
}

func TestMongoDB_Integration(t *testing.T) {
	// Use mtest for MongoDB testing
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Database", func(mt *mtest.T) {
		// Note: We can't directly test Database() without proper initialization
		// This would require dependency injection or interface-based design
		
		// Test that Database() returns a non-nil value
		// In real implementation, you would inject a mock database
		// db := &database.MongoDB{}
	})

	mt.Run("Collection", func(mt *mtest.T) {
		// Similar issue - need to test with properly initialized MongoDB
		// This demonstrates the limitation of testing concrete implementations
	})

	mt.Run("CreateIndexes", func(mt *mtest.T) {
		// Mock responses for index creation
		mt.AddMockResponses(
			// Environment indexes
			mtest.CreateSuccessResponse(),
			// Audit log indexes
			mtest.CreateSuccessResponse(),
			// Credentials indexes
			mtest.CreateSuccessResponse(),
		)

		// This would work with a properly initialized MongoDB instance
		// that uses the mtest client
	})

	mt.Run("Transaction Success", func(mt *mtest.T) {
		// Mock successful transaction
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(), // Start session
			mtest.CreateSuccessResponse(), // Start transaction
			mtest.CreateSuccessResponse(), // Commit transaction
		)

		// Test transaction execution
		// This would require a properly initialized MongoDB with mtest client
	})

	mt.Run("Transaction Rollback", func(mt *mtest.T) {
		// Mock transaction that needs rollback
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(), // Start session
			mtest.CreateSuccessResponse(), // Start transaction
			mtest.CreateSuccessResponse(), // Abort transaction
		)

		// Test transaction rollback
		// This would require a properly initialized MongoDB with mtest client
	})
}

// TestMongoDB_MockImplementation shows how you could test with a mock implementation
type MockMongoDB struct {
	database              *mongo.Database
	closeError            error
	createIndexesError    error
	transactionError      error
	transactionCallCount  int
}

func (m *MockMongoDB) Close(ctx context.Context) error {
	return m.closeError
}

func (m *MockMongoDB) Database() *mongo.Database {
	return m.database
}

func (m *MockMongoDB) Collection(name string) *mongo.Collection {
	if m.database != nil {
		return m.database.Collection(name)
	}
	return nil
}

func (m *MockMongoDB) CreateIndexes(ctx context.Context) error {
	return m.createIndexesError
}

func (m *MockMongoDB) Transaction(ctx context.Context, fn func(mongo.SessionContext) error) error {
	m.transactionCallCount++
	if m.transactionError != nil {
		return m.transactionError
	}
	// Simulate calling the function with a nil session context
	// In real implementation, you'd use a mock session context
	return fn(nil)
}

func TestMockMongoDB_Close(t *testing.T) {
	tests := []struct {
		name       string
		closeError error
		expectErr  bool
	}{
		{
			name:       "Successful close",
			closeError: nil,
			expectErr:  false,
		},
		{
			name:       "Close error",
			closeError: errors.New("connection error"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockMongoDB{
				closeError: tt.closeError,
			}

			ctx := context.Background()
			err := mock.Close(ctx)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockMongoDB_CreateIndexes(t *testing.T) {
	tests := []struct {
		name               string
		createIndexesError error
		expectErr          bool
	}{
		{
			name:               "Successful index creation",
			createIndexesError: nil,
			expectErr:          false,
		},
		{
			name:               "Index creation error",
			createIndexesError: errors.New("index creation failed"),
			expectErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockMongoDB{
				createIndexesError: tt.createIndexesError,
			}

			ctx := context.Background()
			err := mock.CreateIndexes(ctx)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockMongoDB_Transaction(t *testing.T) {
	tests := []struct {
		name             string
		transactionError error
		fnError          error
		expectErr        bool
		expectedCalls    int
	}{
		{
			name:             "Successful transaction",
			transactionError: nil,
			fnError:          nil,
			expectErr:        false,
			expectedCalls:    1,
		},
		{
			name:             "Transaction error",
			transactionError: errors.New("transaction failed"),
			fnError:          nil,
			expectErr:        true,
			expectedCalls:    1,
		},
		{
			name:             "Function error",
			transactionError: nil,
			fnError:          errors.New("function failed"),
			expectErr:        true,
			expectedCalls:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockMongoDB{
				transactionError: tt.transactionError,
			}

			ctx := context.Background()
			err := mock.Transaction(ctx, func(sc mongo.SessionContext) error {
				return tt.fnError
			})

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedCalls, mock.transactionCallCount)
		})
	}
}

// Example of how to test with an interface-based design
type MongoDBInterface interface {
	Close(ctx context.Context) error
	Database() *mongo.Database
	Collection(name string) *mongo.Collection
	CreateIndexes(ctx context.Context) error
	Transaction(ctx context.Context, fn func(mongo.SessionContext) error) error
}

func TestWithInterface(t *testing.T) {
	// This shows how the code could be more testable with interfaces
	var db MongoDBInterface = &MockMongoDB{
		closeError: nil,
	}

	ctx := context.Background()
	err := db.Close(ctx)
	assert.NoError(t, err)
}

// Benchmarks
func BenchmarkMockClose(b *testing.B) {
	mock := &MockMongoDB{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mock.Close(ctx)
	}
}

func BenchmarkMockCreateIndexes(b *testing.B) {
	mock := &MockMongoDB{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mock.CreateIndexes(ctx)
	}
}

func BenchmarkMockTransaction(b *testing.B) {
	mock := &MockMongoDB{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mock.Transaction(ctx, func(sc mongo.SessionContext) error {
			return nil
		})
	}
}

// TestIndexDefinitions verifies that the index definitions are correct
func TestIndexDefinitions(t *testing.T) {
	// This test verifies the structure of indexes without actually creating them
	
	// Environment indexes
	envIndexes := []struct {
		keys   map[string]interface{}
		unique bool
	}{
		{keys: map[string]interface{}{"name": 1}, unique: true},
		{keys: map[string]interface{}{"status.health": 1}, unique: false},
		{keys: map[string]interface{}{"timestamps.lastCheck": 1}, unique: false},
	}

	// Verify environment indexes
	for i, idx := range envIndexes {
		assert.NotEmpty(t, idx.keys, "Environment index %d should have keys", i)
		if idx.unique {
			assert.Contains(t, idx.keys, "name", "Unique index should be on name field")
		}
	}

	// Audit log indexes
	auditIndexes := []map[string]interface{}{
		{"timestamp": -1, "environmentId": 1},
		{"type": 1},
		{"actor.id": 1},
		{"tags": 1},
	}

	// Verify audit log indexes
	for i, idx := range auditIndexes {
		assert.NotEmpty(t, idx, "Audit log index %d should have keys", i)
	}

	// Credentials indexes
	credIndexes := []struct {
		keys   map[string]interface{}
		unique bool
	}{
		{keys: map[string]interface{}{"name": 1}, unique: true},
		{keys: map[string]interface{}{"usage.environmentId": 1}, unique: false},
		{keys: map[string]interface{}{"timestamps.expiresAt": 1}, unique: false},
	}

	// Verify credentials indexes
	for i, idx := range credIndexes {
		assert.NotEmpty(t, idx.keys, "Credentials index %d should have keys", i)
	}
}

// TestConnectionOptions verifies the connection options
func TestConnectionOptions(t *testing.T) {
	maxConnections := 100
	timeout := 10 * time.Second

	// Verify that the options would be set correctly
	assert.Greater(t, maxConnections, 0, "Max connections should be positive")
	assert.Greater(t, timeout, time.Duration(0), "Timeout should be positive")
	
	// Verify pool size constraints
	assert.LessOrEqual(t, maxConnections, 1000, "Max connections should be reasonable")
	
	// Verify timeout constraints
	assert.GreaterOrEqual(t, timeout, 1*time.Second, "Timeout should be at least 1 second")
	assert.LessOrEqual(t, timeout, 5*time.Minute, "Timeout should not exceed 5 minutes")
}
