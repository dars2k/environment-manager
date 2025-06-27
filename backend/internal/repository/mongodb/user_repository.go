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

type userRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new MongoDB user repository
func NewUserRepository(db *mongo.Database) interfaces.UserRepository {
	collection := db.Collection("users")
	
	// Create unique indexes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Username unique index
	_, _ = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	
	return &userRepository{
		collection: collection,
	}
}

// Create inserts a new user
func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, interfaces.ErrInvalidID
	}

	var user entities.User
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, interfaces.ErrNotFound
	}
	return &user, err
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	var user entities.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, interfaces.ErrNotFound
	}
	return &user, err
}


// List retrieves all users
func (r *userRepository) List(ctx context.Context, filter interfaces.ListFilter) ([]*entities.User, error) {
	query := bson.M{}
	if filter.Active != nil {
		query["active"] = *filter.Active
	}

	findOptions := options.Find()
	if filter.Limit > 0 {
		findOptions.SetLimit(int64(filter.Limit))
		if filter.Page > 0 {
			findOptions.SetSkip(int64((filter.Page - 1) * filter.Limit))
		}
	}
	findOptions.SetSort(bson.D{{Key: "username", Value: 1}})

	cursor, err := r.collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*entities.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, id string, user *entities.User) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return interfaces.ErrInvalidID
	}

	user.UpdatedAt = time.Now()
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": user}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return interfaces.ErrNotFound
	}

	return nil
}

// UpdateLastLogin updates the user's last login time
func (r *userRepository) UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"lastLoginAt": now,
			"updatedAt":   now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return interfaces.ErrNotFound
	}

	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return interfaces.ErrInvalidID
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return interfaces.ErrNotFound
	}

	return nil
}

// Count returns the total number of users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
