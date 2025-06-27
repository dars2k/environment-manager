package mongodb

import (
	"context"
	"fmt"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/repository/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AuditLogRepository implements the audit log repository interface for MongoDB
type AuditLogRepository struct {
	collection *mongo.Collection
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *mongo.Database) *AuditLogRepository {
	return &AuditLogRepository{
		collection: db.Collection("audit_log"),
	}
}

// Create creates a new audit log entry
func (r *AuditLogRepository) Create(ctx context.Context, log *entities.AuditLog) error {
	result, err := r.collection.InsertOne(ctx, log)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	
	log.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves an audit log entry by ID
func (r *AuditLogRepository) GetByID(ctx context.Context, id string) (*entities.AuditLog, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid object ID: %w", err)
	}
	
	var log entities.AuditLog
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&log)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("audit log not found")
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}
	
	return &log, nil
}

// List retrieves a list of audit logs with filtering
func (r *AuditLogRepository) List(ctx context.Context, filter interfaces.AuditLogFilter) ([]*entities.AuditLog, error) {
	query := bson.M{}
	
	// Apply filters
	if filter.EnvironmentID != "" {
		objectID, err := primitive.ObjectIDFromHex(filter.EnvironmentID)
		if err == nil {
			query["environmentId"] = objectID
		}
	}
	
	if filter.Type != nil {
		query["type"] = *filter.Type
	}
	
	if filter.Severity != nil {
		query["severity"] = *filter.Severity
	}
	
	if filter.ActorID != "" {
		query["actor.id"] = filter.ActorID
	}
	
	if len(filter.Tags) > 0 {
		query["tags"] = bson.M{"$in": filter.Tags}
	}
	
	// Date range filter
	if filter.StartDate != nil || filter.EndDate != nil {
		dateFilter := bson.M{}
		if filter.StartDate != nil {
			dateFilter["$gte"] = *filter.StartDate
		}
		if filter.EndDate != nil {
			dateFilter["$lte"] = *filter.EndDate
		}
		query["timestamp"] = dateFilter
	}
	
	// Set up find options
	findOptions := options.Find()
	if filter.Pagination != nil {
		findOptions.SetSkip(int64(filter.Pagination.GetOffset()))
		findOptions.SetLimit(int64(filter.Pagination.GetLimit()))
	}
	findOptions.SetSort(bson.M{"timestamp": -1}) // Sort by timestamp descending
	
	cursor, err := r.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer cursor.Close(ctx)
	
	var logs []*entities.AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, fmt.Errorf("failed to decode audit logs: %w", err)
	}
	
	return logs, nil
}

// Count counts audit logs matching the filter
func (r *AuditLogRepository) Count(ctx context.Context, filter interfaces.AuditLogFilter) (int64, error) {
	query := bson.M{}
	
	// Apply same filters as List
	if filter.EnvironmentID != "" {
		objectID, err := primitive.ObjectIDFromHex(filter.EnvironmentID)
		if err == nil {
			query["environmentId"] = objectID
		}
	}
	
	if filter.Type != nil {
		query["type"] = *filter.Type
	}
	
	if filter.Severity != nil {
		query["severity"] = *filter.Severity
	}
	
	if filter.ActorID != "" {
		query["actor.id"] = filter.ActorID
	}
	
	if len(filter.Tags) > 0 {
		query["tags"] = bson.M{"$in": filter.Tags}
	}
	
	// Date range filter
	if filter.StartDate != nil || filter.EndDate != nil {
		dateFilter := bson.M{}
		if filter.StartDate != nil {
			dateFilter["$gte"] = *filter.StartDate
		}
		if filter.EndDate != nil {
			dateFilter["$lte"] = *filter.EndDate
		}
		query["timestamp"] = dateFilter
	}
	
	count, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit logs: %w", err)
	}
	
	return count, nil
}

// DeleteOlderThan deletes audit logs older than the specified time
func (r *AuditLogRepository) DeleteOlderThan(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.collection.DeleteMany(ctx, bson.M{
		"timestamp": bson.M{"$lt": before},
	})
	if err != nil {
		return 0, fmt.Errorf("failed to delete old audit logs: %w", err)
	}
	
	return result.DeletedCount, nil
}
