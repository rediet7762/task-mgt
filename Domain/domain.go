// Package domain contains the core business entities, DTOs, and the
// interfaces that define the boundaries between architectural layers.
//
// This package sits at the center of the Clean Architecture dependency
// graph: it depends on nothing else in this codebase, and every other layer
// (Repositories, Usecases, Infrastructure, Delivery) depends on it instead
// of on one another directly. That inversion is what lets, e.g., MongoDB be
// swapped out for another database, or bcrypt for another hashing library,
// without touching business logic.
//
// The only third-party import here is the MongoDB ObjectID type, kept
// because it is the identifier format used throughout the API's JSON
// contract. Everything else - entities, request/response shapes, and all
// interfaces - is plain Go.
package domain

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ---------------------------------------------------------------------------
// Sentinel errors
//
// Shared across layers so controllers can map failures to the right HTTP
// status without inspecting error strings or depending on a repository's
// concrete error type.
// ---------------------------------------------------------------------------

var (
	// Task errors
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidID    = errors.New("invalid task id")

	// User / auth errors
	ErrUsernameTaken = errors.New("username already taken")
	ErrInvalidCreds  = errors.New("invalid username or password")
	ErrUserNotFound  = errors.New("user not found")
)

// Role constants used throughout the application.
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// ---------------------------------------------------------------------------
// Entities
// ---------------------------------------------------------------------------

// Task represents a task in the system - the core business entity this
// application manages.
type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	DueDate     string             `bson:"due_date" json:"due_date"`
	Status      string             `bson:"status" json:"status"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// User represents a registered account in the system.
//
// Password is only ever populated long enough to hash it on registration or
// compare it on login; the `json:"-"` tag ensures the hash is never
// serialized back to a client.
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username  string             `bson:"username" json:"username"`
	Password  string             `bson:"password" json:"-"`
	Role      string             `bson:"role" json:"role,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// Claims is the business representation of an authenticated session. It is
// deliberately framework-agnostic: the Infrastructure layer's JWT service is
// responsible for mapping this to and from whatever claims type the
// underlying JWT library requires.
type Claims struct {
	UserID    string
	Username  string
	Role      string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// ---------------------------------------------------------------------------
// Request / response DTOs
// ---------------------------------------------------------------------------

// CreateTaskRequest represents the request body for creating a new task.
type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	DueDate     string `json:"due_date" binding:"required"`
	Status      string `json:"status" binding:"required,oneof=pending in-progress completed"`
}

// TaskUpdateInput represents the fields that can be updated on a task. Zero
// values are treated as "not supplied" (partial update semantics).
type TaskUpdateInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
	Status      string `json:"status" binding:"omitempty,oneof=pending in-progress completed"`
}

// UpdateTaskRequest represents the request body for updating a task.
type UpdateTaskRequest = TaskUpdateInput

// RegisterRequest represents the request body for POST /register.
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the request body for POST /login.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ---------------------------------------------------------------------------
// Repository interfaces (implemented in Repositories/, consumed by Usecases/)
// ---------------------------------------------------------------------------

// TaskRepository abstracts persistence for tasks. Implementations live in
// the Repositories layer and may talk to any storage engine; the rest of
// the application only ever depends on this interface.
type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	SeedSampleTasks(ctx context.Context) error
	GetAll(ctx context.Context) ([]Task, error)
	Filter(ctx context.Context, status string) ([]Task, error)
	GetByID(ctx context.Context, id string) (*Task, error)
	Update(ctx context.Context, id string, input TaskUpdateInput) (*Task, error)
	Delete(ctx context.Context, id string) error
}

// UserRepository abstracts persistence for user accounts.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	CountAll(ctx context.Context) (int64, error)
	CountByUsername(ctx context.Context, username string) (int64, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	UpdateRole(ctx context.Context, username, role string) error
}

// ---------------------------------------------------------------------------
// Infrastructure service interfaces (implemented in Infrastructure/,
// consumed by Usecases/)
// ---------------------------------------------------------------------------

// PasswordService abstracts password hashing/verification so use cases
// don't depend on bcrypt (or any specific library) directly.
type PasswordService interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

// JWTService abstracts token generation/parsing so use cases don't depend
// on a specific JWT library directly.
type JWTService interface {
	GenerateToken(user User) (string, error)
	ParseToken(tokenStr string) (*Claims, error)
}

// ---------------------------------------------------------------------------
// Usecase interfaces (implemented in Usecases/, consumed by Delivery/controllers)
// ---------------------------------------------------------------------------

// TaskUsecase encapsulates the application's business rules for tasks.
type TaskUsecase interface {
	GetTasks(ctx context.Context, status string) ([]Task, error)
	GetTask(ctx context.Context, id string) (*Task, error)
	CreateTask(ctx context.Context, req CreateTaskRequest) (*Task, error)
	UpdateTask(ctx context.Context, id string, req UpdateTaskRequest) (*Task, error)
	DeleteTask(ctx context.Context, id string) error
}

// UserUsecase encapsulates the application's business rules for accounts
// and authentication.
type UserUsecase interface {
	Register(ctx context.Context, req RegisterRequest) (*User, error)
	Login(ctx context.Context, req LoginRequest) (string, error)
	Promote(ctx context.Context, username string) error
}
