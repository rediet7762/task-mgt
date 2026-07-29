package usecases

import (
	"context"
	"errors"
	"testing"

	"task-manager/Domain"
)

// fakeTaskRepository is a minimal in-memory stand-in for domain.TaskRepository.
// Because Usecases depends only on the interface, tests never need a real
// database or a mocking framework.
type fakeTaskRepository struct {
	tasks       map[string]domain.Task
	nextID      int
	createErr   error
	getByIDErr  error
	updateErr   error
	deleteErr   error
	filterCalls []string
}

func newFakeTaskRepository() *fakeTaskRepository {
	return &fakeTaskRepository{tasks: make(map[string]domain.Task)}
}

func (f *fakeTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.nextID++
	id := string(rune('a' + f.nextID))
	f.tasks[id] = *task
	return nil
}

func (f *fakeTaskRepository) SeedSampleTasks(ctx context.Context) error { return nil }

func (f *fakeTaskRepository) GetAll(ctx context.Context) ([]domain.Task, error) {
	out := make([]domain.Task, 0, len(f.tasks))
	for _, t := range f.tasks {
		out = append(out, t)
	}
	return out, nil
}

func (f *fakeTaskRepository) Filter(ctx context.Context, status string) ([]domain.Task, error) {
	f.filterCalls = append(f.filterCalls, status)
	out := make([]domain.Task, 0)
	for _, t := range f.tasks {
		if t.Status == status {
			out = append(out, t)
		}
	}
	return out, nil
}

func (f *fakeTaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}
	t, ok := f.tasks[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}
	return &t, nil
}

func (f *fakeTaskRepository) Update(ctx context.Context, id string, input domain.TaskUpdateInput) (*domain.Task, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	t, ok := f.tasks[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}
	if input.Title != "" {
		t.Title = input.Title
	}
	if input.Status != "" {
		t.Status = input.Status
	}
	f.tasks[id] = t
	return &t, nil
}

func (f *fakeTaskRepository) Delete(ctx context.Context, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if _, ok := f.tasks[id]; !ok {
		return domain.ErrTaskNotFound
	}
	delete(f.tasks, id)
	return nil
}

func TestTaskUsecase_GetTasks_NoFilter(t *testing.T) {
	repo := newFakeTaskRepository()
	repo.tasks["a"] = domain.Task{Title: "one", Status: "pending"}
	repo.tasks["b"] = domain.Task{Title: "two", Status: "completed"}
	uc := NewTaskUsecase(repo)

	tasks, err := uc.GetTasks(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if len(repo.filterCalls) != 0 {
		t.Fatalf("expected Filter not to be called when status is empty, got calls: %v", repo.filterCalls)
	}
}

func TestTaskUsecase_GetTasks_WithFilter(t *testing.T) {
	repo := newFakeTaskRepository()
	repo.tasks["a"] = domain.Task{Title: "one", Status: "pending"}
	repo.tasks["b"] = domain.Task{Title: "two", Status: "completed"}
	uc := NewTaskUsecase(repo)

	tasks, err := uc.GetTasks(context.Background(), "completed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Title != "two" {
		t.Fatalf("expected exactly the completed task, got %+v", tasks)
	}
	if len(repo.filterCalls) != 1 || repo.filterCalls[0] != "completed" {
		t.Fatalf("expected Filter to be called with 'completed', got calls: %v", repo.filterCalls)
	}
}

func TestTaskUsecase_GetTask_NotFound(t *testing.T) {
	repo := newFakeTaskRepository()
	uc := NewTaskUsecase(repo)

	_, err := uc.GetTask(context.Background(), "missing")
	if !errors.Is(err, domain.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskUsecase_CreateTask(t *testing.T) {
	repo := newFakeTaskRepository()
	uc := NewTaskUsecase(repo)

	task, err := uc.CreateTask(context.Background(), domain.CreateTaskRequest{
		Title:       "New task",
		Description: "desc",
		DueDate:     "2026-08-01",
		Status:      "pending",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "New task" || task.Status != "pending" {
		t.Fatalf("unexpected task returned: %+v", task)
	}
	if len(repo.tasks) != 1 {
		t.Fatalf("expected task to be persisted via repository, got %d tasks", len(repo.tasks))
	}
}

func TestTaskUsecase_UpdateTask_NotFound(t *testing.T) {
	repo := newFakeTaskRepository()
	uc := NewTaskUsecase(repo)

	_, err := uc.UpdateTask(context.Background(), "missing", domain.UpdateTaskRequest{Title: "x"})
	if !errors.Is(err, domain.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskUsecase_DeleteTask(t *testing.T) {
	repo := newFakeTaskRepository()
	repo.tasks["a"] = domain.Task{Title: "one"}
	uc := NewTaskUsecase(repo)

	if err := uc.DeleteTask(context.Background(), "a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := repo.tasks["a"]; ok {
		t.Fatalf("expected task to be removed")
	}

	err := uc.DeleteTask(context.Background(), "a")
	if !errors.Is(err, domain.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound on second delete, got %v", err)
	}
}
