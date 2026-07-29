package usecases

import (
	"context"

	"task-manager/Domain"
)

// taskUsecase implements domain.TaskUsecase. It orchestrates task-related
// business logic and depends only on the domain.TaskRepository interface,
// never on a concrete database implementation.
type taskUsecase struct {
	taskRepo domain.TaskRepository
}

// NewTaskUsecase constructs a domain.TaskUsecase backed by the given
// repository.
func NewTaskUsecase(taskRepo domain.TaskRepository) domain.TaskUsecase {
	return &taskUsecase{taskRepo: taskRepo}
}

// GetTasks returns all tasks, optionally filtered by status.
func (u *taskUsecase) GetTasks(ctx context.Context, status string) ([]domain.Task, error) {
	if status != "" {
		return u.taskRepo.Filter(ctx, status)
	}
	return u.taskRepo.GetAll(ctx)
}

// GetTask returns a single task by ID.
func (u *taskUsecase) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	return u.taskRepo.GetByID(ctx, id)
}

// CreateTask builds a new Task entity from the request and persists it.
// Status is validated as one of pending/in-progress/completed at the
// delivery layer via binding tags, so no further business validation is
// needed here beyond mapping the request into the entity.
func (u *taskUsecase) CreateTask(ctx context.Context, req domain.CreateTaskRequest) (*domain.Task, error) {
	task := &domain.Task{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
		Status:      req.Status,
	}
	if err := u.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// UpdateTask applies a partial update to an existing task.
func (u *taskUsecase) UpdateTask(ctx context.Context, id string, req domain.UpdateTaskRequest) (*domain.Task, error) {
	return u.taskRepo.Update(ctx, id, req)
}

// DeleteTask removes a task by ID.
func (u *taskUsecase) DeleteTask(ctx context.Context, id string) error {
	return u.taskRepo.Delete(ctx, id)
}
