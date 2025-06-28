package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MigrateRemoveEmailField removes the email field from all user documents and drops the email index
func MigrateRemoveEmailField(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	usersCollection := db.Collection("users")

	// Step 1: Drop the email index if it exists
	log.Println("Dropping email index if exists...")
	indexView := usersCollection.Indexes()
	cursor, err := indexView.List(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		return err
	}

	for _, index := range indexes {
		if keys, ok := index["key"].(bson.M); ok {
			if _, hasEmail := keys["email"]; hasEmail {
				if name, ok := index["name"].(string); ok {
					log.Printf("Dropping index: %s", name)
					if _, err := indexView.DropOne(ctx, name); err != nil {
						log.Printf("Error dropping index %s: %v", name, err)
					}
				}
			}
		}
	}

	// Step 2: Remove email field from all user documents
	log.Println("Removing email field from all user documents...")
	update := bson.M{
		"$unset": bson.M{"email": ""},
	}
	result, err := usersCollection.UpdateMany(ctx, bson.M{}, update)
	if err != nil {
		return err
	}
	log.Printf("Updated %d documents to remove email field", result.ModifiedCount)

	// Step 3: List remaining indexes for verification
	log.Println("Current indexes after migration:")
	cursor, err = indexView.List(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var remainingIndexes []bson.M
	if err := cursor.All(ctx, &remainingIndexes); err != nil {
		return err
	}

	for _, index := range remainingIndexes {
		log.Printf("Index: %v", index["name"])
	}

	return nil
}

// RunMigrationOnStartup runs the email migration when the application starts
func RunMigrationOnStartup(client *mongo.Client, dbName string) error {
	db := client.Database(dbName)
	return MigrateRemoveEmailField(db)
}
