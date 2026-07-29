package repositories

import (
	"context"
	"errors"

	"task-manager/Domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// userRepository implements domain.UserRepository against a MongoDB
// collection. It is the only place in the codebase that talks to Mongo for
// user data - all business rules (uniqueness policy, first-user-is-admin,
// credential checking) live one layer up, in Usecases.
type userRepository struct {
	collection *mongo.Collection
}

// NewUserRepository constructs a domain.UserRepository bound to the given
// collection.
func NewUserRepository(collection *mongo.Collection) domain.UserRepository {
	return &userRepository{collection: collection}
}

func (r *userRepository) ensureCollection() error {
	if r.collection == nil {
		return errors.New("user collection not initialized")
	}
	return nil
}

// Create inserts a new user document. The caller is responsible for
// populating a pre-hashed password and any business-decided fields (e.g.
// role) before calling this.
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.ensureCollection(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid
	}
	return nil
}

// CountAll returns the total number of user documents.
func (r *userRepository) CountAll(ctx context.Context) (int64, error) {
	if err := r.ensureCollection(); err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{})
}

// CountByUsername returns how many user documents match the given username
// (0 or 1, since usernames are unique by application policy).
func (r *userRepository) CountByUsername(ctx context.Context, username string) (int64, error) {
	if err := r.ensureCollection(); err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	return r.collection.CountDocuments(ctx, bson.M{"username": username})
}

// FindByUsername retrieves a single user by username.
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	if err := r.ensureCollection(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// UpdateRole sets the role field for the given username. Returns
// domain.ErrUserNotFound if no document matched.
func (r *userRepository) UpdateRole(ctx context.Context, username, role string) error {
	if err := r.ensureCollection(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"username": username},
		bson.M{"$set": bson.M{"role": role}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
