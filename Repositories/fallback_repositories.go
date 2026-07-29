package repositories

import (
	"context"
	"sort"
	"time"

	"task-manager/Domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// inMemoryTaskRepository provides a simple local fallback for development when
// MongoDB is unavailable. It implements the same domain.TaskRepository
// interface so the rest of the application can run unchanged.
type inMemoryTaskRepository struct {
	tasks map[string]domain.Task
}

func NewInMemoryTaskRepository() domain.TaskRepository {
	return &inMemoryTaskRepository{tasks: make(map[string]domain.Task)}
}

func (r *inMemoryTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	if task.ID.IsZero() {
		task.ID = primitive.NewObjectID()
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now().UTC()
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = task.CreatedAt
	}

	r.tasks[task.ID.Hex()] = *task
	return nil
}

func (r *inMemoryTaskRepository) SeedSampleTasks(ctx context.Context) error {
	if len(r.tasks) > 0 {
		return nil
	}

	now := time.Now().UTC()
	tasks := []domain.Task{
		{
			ID:          primitive.NewObjectID(),
			Title:       "Learn Go",
			Description: "Complete the Go language basics and practice CRUD operations.",
			Status:      "pending",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Title:       "Connect MongoDB Atlas",
			Description: "Connect the Task Manager application to MongoDB Atlas.",
			Status:      "in-progress",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          primitive.NewObjectID(),
			Title:       "Build REST API",
			Description: "Implement and test GET, POST, PUT, and DELETE endpoints.",
			Status:      "completed",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	for _, task := range tasks {
		r.tasks[task.ID.Hex()] = task
	}
	return nil
}

func (r *inMemoryTaskRepository) GetAll(ctx context.Context) ([]domain.Task, error) {
	tasks := make([]domain.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})
	return tasks, nil
}

func (r *inMemoryTaskRepository) Filter(ctx context.Context, status string) ([]domain.Task, error) {
	tasks := make([]domain.Task, 0)
	for _, task := range r.tasks {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})
	return tasks, nil
}

func (r *inMemoryTaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	task, ok := r.tasks[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}
	return &task, nil
}

func (r *inMemoryTaskRepository) Update(ctx context.Context, id string, input domain.TaskUpdateInput) (*domain.Task, error) {
	task, ok := r.tasks[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}

	if input.Title != "" {
		task.Title = input.Title
	}
	if input.Description != "" {
		task.Description = input.Description
	}
	if input.Status != "" {
		task.Status = input.Status
	}
	if input.DueDate != "" {
		task.DueDate = input.DueDate
	}
	task.UpdatedAt = time.Now().UTC()

	r.tasks[id] = task
	return &task, nil
}

func (r *inMemoryTaskRepository) Delete(ctx context.Context, id string) error {
	if _, ok := r.tasks[id]; !ok {
		return domain.ErrTaskNotFound
	}
	delete(r.tasks, id)
	return nil
}

// inMemoryUserRepository provides a simple local fallback for authentication
// state when MongoDB is unavailable.
type inMemoryUserRepository struct {
	users map[string]domain.User
}

func NewInMemoryUserRepository() domain.UserRepository {
	return &inMemoryUserRepository{users: make(map[string]domain.User)}
}

func (r *inMemoryUserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID.IsZero() {
		user.ID = primitive.NewObjectID()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	r.users[user.Username] = *user
	return nil
}

func (r *inMemoryUserRepository) CountAll(ctx context.Context) (int64, error) {
	return int64(len(r.users)), nil
}

func (r *inMemoryUserRepository) CountByUsername(ctx context.Context, username string) (int64, error) {
	if _, ok := r.users[username]; ok {
		return 1, nil
	}
	return 0, nil
}

func (r *inMemoryUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, ok := r.users[username]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return &user, nil
}

func (r *inMemoryUserRepository) UpdateRole(ctx context.Context, username, role string) error {
	user, ok := r.users[username]
	if !ok {
		return domain.ErrUserNotFound
	}
	user.Role = role
	r.users[username] = user
	return nil
}

var (
	_ domain.TaskRepository = (*inMemoryTaskRepository)(nil)
	_ domain.UserRepository = (*inMemoryUserRepository)(nil)
)
