package mongodb

import (
	"context"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type logRepository struct {
	collection *mongo.Collection
}

// NewLogRepository creates a new MongoDB log repository
func NewLogRepository(db *mongo.Database) interfaces.LogRepository {
	return &logRepository{
		collection: db.Collection("logs"),
	}
}

// Create inserts a new log entry
func (r *logRepository) Create(ctx context.Context, log *entities.Log) error {
	log.Timestamp = time.Now()
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

// List retrieves logs with filtering and pagination
func (r *logRepository) List(ctx context.Context, filter interfaces.LogFilter) ([]*entities.Log, int64, error) {
	// Build query
	query := bson.M{}
	
	if filter.EnvironmentID != nil {
		query["environmentId"] = filter.EnvironmentID
	}
	
	if filter.UserID != nil {
		query["userId"] = filter.UserID
	}
	
	if filter.Type != "" {
		query["type"] = filter.Type
	}
	
	if filter.Level != "" {
		query["level"] = filter.Level
	}
	
	if filter.Action != "" {
		query["action"] = filter.Action
	}
	
	if !filter.StartTime.IsZero() || !filter.EndTime.IsZero() {
		timeFilter := bson.M{}
		if !filter.StartTime.IsZero() {
			timeFilter["$gte"] = filter.StartTime
		}
		if !filter.EndTime.IsZero() {
			timeFilter["$lte"] = filter.EndTime
		}
		query["timestamp"] = timeFilter
	}
	
	if filter.Search != "" {
		query["$or"] = []bson.M{
			{"message": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"environmentName": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"username": bson.M{"$regex": filter.Search, "$options": "i"}},
		}
	}
	
	// Count total documents
	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	
	// Find documents with pagination and sorting
	findOptions := options.Find()
	if filter.Limit > 0 {
		findOptions.SetLimit(int64(filter.Limit))
		if filter.Page > 0 {
			findOptions.SetSkip(int64((filter.Page - 1) * filter.Limit))
		}
	}
	
	// Sort by timestamp descending by default
	findOptions.SetSort(bson.D{{Key: "timestamp", Value: -1}})
	
	cursor, err := r.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	
	var logs []*entities.Log
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}
	
	return logs, total, nil
}

// GetByID retrieves a log by its ID
func (r *logRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*entities.Log, error) {
	var log entities.Log
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&log)
	if err == mongo.ErrNoDocuments {
		return nil, interfaces.ErrNotFound
	}
	return &log, err
}

// DeleteOld removes logs older than the specified duration
func (r *logRepository) DeleteOld(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)
	result, err := r.collection.DeleteMany(ctx, bson.M{
		"timestamp": bson.M{"$lt": cutoffTime},
	})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// GetEnvironmentLogs retrieves recent logs for a specific environment
func (r *logRepository) GetEnvironmentLogs(ctx context.Context, envID primitive.ObjectID, limit int) ([]*entities.Log, error) {
	findOptions := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit))
	
	cursor, err := r.collection.Find(ctx, bson.M{"environmentId": envID}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var logs []*entities.Log
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, err
	}
	
	return logs, nil
}

// Count returns the count of logs based on filter
func (r *logRepository) Count(ctx context.Context, filter interfaces.LogFilter) (int64, error) {
	// Build query
	query := bson.M{}
	
	if filter.EnvironmentID != nil {
		query["environmentId"] = filter.EnvironmentID
	}
	
	if filter.UserID != nil {
		query["userId"] = filter.UserID
	}
	
	if filter.Type != "" {
		query["type"] = filter.Type
	}
	
	if filter.Level != "" {
		query["level"] = filter.Level
	}
	
	if filter.Action != "" {
		query["action"] = filter.Action
	}
	
	if !filter.StartTime.IsZero() || !filter.EndTime.IsZero() {
		timeFilter := bson.M{}
		if !filter.StartTime.IsZero() {
			timeFilter["$gte"] = filter.StartTime
		}
		if !filter.EndTime.IsZero() {
			timeFilter["$lte"] = filter.EndTime
		}
		query["timestamp"] = timeFilter
	}
	
	if filter.Search != "" {
		query["$or"] = []bson.M{
			{"message": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"environmentName": bson.M{"$regex": filter.Search, "$options": "i"}},
			{"username": bson.M{"$regex": filter.Search, "$options": "i"}},
		}
	}
	
	// Count documents
	return r.collection.CountDocuments(ctx, query)
}
