package repositories

import (
	"context"
	"errors"
	"time"

	"task-manager/Domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultTimeout = 10 * time.Second

// taskRepository implements domain.TaskRepository against a MongoDB
// collection. It is the only place in the codebase that talks to Mongo for
// task data - all business rules live one layer up, in Usecases.
type taskRepository struct {
	collection *mongo.Collection
}

// NewTaskRepository constructs a domain.TaskRepository bound to the given
// collection.
func NewTaskRepository(collection *mongo.Collection) domain.TaskRepository {
	return &taskRepository{collection: collection}
}

func (r *taskRepository) ensureCollection() error {
	if r.collection == nil {
		return errors.New("task collection not initialized")
	}
	return nil
}

// Create inserts a new task document and stamps CreatedAt/UpdatedAt
// server-side so clients cannot spoof them.
func (r *taskRepository) Create(ctx context.Context, task *domain.Task) error {
	if err := r.ensureCollection(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now

	result, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		task.ID = oid
	}
	return nil
}

// SeedSampleTasks inserts sample tasks if the collection is empty.
func (r *taskRepository) SeedSampleTasks(ctx context.Context) error {
	if err := r.ensureCollection(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()
	tasks := []interface{}{
		domain.Task{
			Title:       "Learn Go",
			Description: "Complete the Go language basics and practice CRUD operations.",
			Status:      "pending",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		domain.Task{
			Title:       "Connect MongoDB Atlas",
			Description: "Connect the Task Manager application to MongoDB Atlas.",
			Status:      "in-progress",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		domain.Task{
			Title:       "Build REST API",
			Description: "Implement and test GET, POST, PUT, and DELETE endpoints.",
			Status:      "completed",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	_, err = r.collection.InsertMany(ctx, tasks)
	return err
}

// GetAll retrieves every task document, sorted by creation date descending.
func (r *taskRepository) GetAll(ctx context.Context) ([]domain.Task, error) {
	if err := r.ensureCollection(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	tasks := make([]domain.Task, 0)
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// Filter retrieves every task document matching the given status.
func (r *taskRepository) Filter(ctx context.Context, status string) ([]domain.Task, error) {
	if err := r.ensureCollection(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"status": status}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	tasks := make([]domain.Task, 0)
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetByID retrieves a single task by its hex ObjectID string.
func (r *taskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	if err := r.ensureCollection(); err != nil {
		return nil, err
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	var task domain.Task
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

// Update applies a partial update to an existing task and returns the
// updated document. Only non-zero fields in the input are applied.
func (r *taskRepository) Update(ctx context.Context, id string, input domain.TaskUpdateInput) (*domain.Task, error) {
	if err := r.ensureCollection(); err != nil {
		return nil, err
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, domain.ErrInvalidID
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	update := bson.M{"updated_at": time.Now().UTC()}
	if input.Title != "" {
		update["title"] = input.Title
	}
	if input.Description != "" {
		update["description"] = input.Description
	}
	if input.Status != "" {
		update["status"] = input.Status
	}
	if input.DueDate != "" {
		update["due_date"] = input.DueDate
	}

	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var task domain.Task
	if err := result.Decode(&task); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

// Delete removes a task by ID. Returns domain.ErrTaskNotFound if no
// document matched (as opposed to a real database/network error).
func (r *taskRepository) Delete(ctx context.Context, id string) error {
	if err := r.ensureCollection(); err != nil {
		return err
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return domain.ErrInvalidID
	}

	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}
