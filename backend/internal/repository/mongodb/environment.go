package mongodb

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"app-env-manager/internal/domain/entities"
	"app-env-manager/internal/domain/errors"
	"app-env-manager/internal/repository/interfaces"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// sanitizeString removes special characters that could be used for NoSQL injection
func sanitizeString(input string) string {
	// Allow only alphanumeric characters and basic punctuation
	safeChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_ "
	var sanitized bytes.Buffer
	for _, char := range input {
		if strings.ContainsRune(safeChars, char) {
			sanitized.WriteRune(char)
		}
	}
	return sanitized.String()
}

// EnvironmentRepository implements the environment repository interface for MongoDB
type EnvironmentRepository struct {
	collection *mongo.Collection
}

// NewEnvironmentRepository creates a new environment repository
func NewEnvironmentRepository(db *mongo.Database) *EnvironmentRepository {
	return &EnvironmentRepository{
		collection: db.Collection("environments"),
	}
}

// validateStringInput ensures the input is a valid string and sanitizes it to prevent NoSQL injection
func validateStringInput(input interface{}) (string, error) {
	// Ensure input is a string type
	str, ok := input.(string)
	if !ok {
		return "", fmt.Errorf("input must be a string")
	}
	
	// Additional validation: ensure it's not empty
	if str == "" {
		return "", fmt.Errorf("input cannot be empty")
	}
	
	// Sanitize input by removing special characters that could be used for NoSQL injection
	sanitizedStr := sanitizeString(str)
	
	return sanitizedStr, nil
}

// Create creates a new environment
func (r *EnvironmentRepository) Create(ctx context.Context, env *entities.Environment) error {
	env.Timestamps.CreatedAt = time.Now()
	env.Timestamps.UpdatedAt = time.Now()
	
	result, err := r.collection.InsertOne(ctx, env)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.ErrEnvironmentAlreadyExists
		}
		return fmt.Errorf("failed to create environment: %w", err)
	}
	
	env.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetByID retrieves an environment by ID
func (r *EnvironmentRepository) GetByID(ctx context.Context, id string) (*entities.Environment, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.NewValidationError("id", "invalid object ID")
	}
	
	var env entities.Environment
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&env)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.ErrEnvironmentNotFound
		}
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}
	
	return &env, nil
}

// GetByName retrieves an environment by name
func (r *EnvironmentRepository) GetByName(ctx context.Context, name string) (*entities.Environment, error) {
	// Validate and sanitize name to prevent NoSQL injection
	validatedName, err := validateStringInput(name)
	if err != nil {
		return nil, errors.NewValidationError("name", "invalid environment name")
	}
	
	var env entities.Environment
	err = r.collection.FindOne(ctx, bson.M{"name": validatedName}).Decode(&env)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.ErrEnvironmentNotFound
		}
		return nil, fmt.Errorf("failed to get environment by name: %w", err)
	}
	
	return &env, nil
}

// List retrieves a list of environments with optional filtering
func (r *EnvironmentRepository) List(ctx context.Context, filter interfaces.ListFilter) ([]*entities.Environment, error) {
	query := bson.M{}
	
	// Apply status filter if provided
	if filter.Status != nil {
		// Convert HealthStatus to string and validate to prevent NoSQL injection
		statusStr := string(*filter.Status)
		validatedStatus, err := validateStringInput(statusStr)
		if err != nil {
			return nil, errors.NewValidationError("status", "invalid status filter")
		}
		query["status.health"] = validatedStatus
	}
	
	// Set up find options
	findOptions := options.Find()
	if filter.Pagination != nil {
		findOptions.SetSkip(int64(filter.Pagination.GetOffset()))
		findOptions.SetLimit(int64(filter.Pagination.GetLimit()))
	}
	findOptions.SetSort(bson.M{"name": 1})
	
	cursor, err := r.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}
	defer cursor.Close(ctx)
	
	var environments []*entities.Environment
	if err := cursor.All(ctx, &environments); err != nil {
		return nil, fmt.Errorf("failed to decode environments: %w", err)
	}
	
	return environments, nil
}

// Update updates an existing environment
func (r *EnvironmentRepository) Update(ctx context.Context, id string, env *entities.Environment) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.NewValidationError("id", "invalid object ID")
	}
	
	env.Timestamps.UpdatedAt = time.Now()
	
	update := bson.M{
		"$set": bson.M{
			"name":           env.Name,
			"description":    env.Description,
			"environmentURL": env.EnvironmentURL,
			"target":         env.Target,
			"credentials":    env.Credentials,
			"healthCheck":    env.HealthCheck,
			"commands":       env.Commands,
			"upgradeConfig":  env.UpgradeConfig,
			"systemInfo":     env.SystemInfo,
			"metadata":       env.Metadata,
			"timestamps":     env.Timestamps,
		},
	}
	
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.ErrEnvironmentAlreadyExists
		}
		return fmt.Errorf("failed to update environment: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return errors.ErrEnvironmentNotFound
	}
	
	return nil
}

// UpdateStatus updates only the status of an environment
func (r *EnvironmentRepository) UpdateStatus(ctx context.Context, id string, status entities.Status) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.NewValidationError("id", "invalid object ID")
	}
	
	update := bson.M{
		"$set": bson.M{
			"status": status,
			"timestamps.updatedAt": time.Now(),
		},
	}
	
	// Update lastHealthyAt if status is healthy
	if status.Health == entities.HealthStatusHealthy {
		now := time.Now()
		update["$set"].(bson.M)["timestamps.lastHealthyAt"] = &now
	}
	
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return fmt.Errorf("failed to update environment status: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return errors.ErrEnvironmentNotFound
	}
	
	return nil
}

// Delete deletes an environment
func (r *EnvironmentRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.NewValidationError("id", "invalid object ID")
	}
	
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}
	
	if result.DeletedCount == 0 {
		return errors.ErrEnvironmentNotFound
	}
	
	return nil
}

// Count counts environments matching the filter
func (r *EnvironmentRepository) Count(ctx context.Context, filter interfaces.ListFilter) (int64, error) {
	query := bson.M{}
	
	// Apply status filter if provided
	if filter.Status != nil {
		// Convert HealthStatus to string and validate to prevent NoSQL injection
		statusStr := string(*filter.Status)
		validatedStatus, err := validateStringInput(statusStr)
		if err != nil {
			return 0, errors.NewValidationError("status", "invalid status filter")
		}
		query["status.health"] = validatedStatus
	}
	
	count, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count environments: %w", err)
	}
	
	return count, nil
}
