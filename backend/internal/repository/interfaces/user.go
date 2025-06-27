package interfaces

import (
	"context"

	"app-env-manager/internal/domain/entities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the interface for user storage operations
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id string) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	List(ctx context.Context, filter ListFilter) ([]*entities.User, error)
	Update(ctx context.Context, id string, user *entities.User) error
	UpdateLastLogin(ctx context.Context, id primitive.ObjectID) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}
