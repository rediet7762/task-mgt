package usecases

import (
	"context"
	"errors"
	"testing"

	"task-manager/Domain"
)

// --- fakes -------------------------------------------------------------

type fakeUserRepository struct {
	users        map[string]domain.User
	createErr    error
	countAllN    int64
	updateRoleFn func(username, role string) error
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{users: make(map[string]domain.User)}
}

func (f *fakeUserRepository) Create(ctx context.Context, user *domain.User) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.users[user.Username] = *user
	f.countAllN++
	return nil
}

func (f *fakeUserRepository) CountAll(ctx context.Context) (int64, error) {
	return f.countAllN, nil
}

func (f *fakeUserRepository) CountByUsername(ctx context.Context, username string) (int64, error) {
	if _, ok := f.users[username]; ok {
		return 1, nil
	}
	return 0, nil
}

func (f *fakeUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	u, ok := f.users[username]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return &u, nil
}

func (f *fakeUserRepository) UpdateRole(ctx context.Context, username, role string) error {
	if f.updateRoleFn != nil {
		return f.updateRoleFn(username, role)
	}
	u, ok := f.users[username]
	if !ok {
		return domain.ErrUserNotFound
	}
	u.Role = role
	f.users[username] = u
	return nil
}

// fakePasswordService trades real hashing for a prefix marker, which is all
// a unit test needs to verify the usecase calls Hash/Compare correctly.
type fakePasswordService struct{}

func (fakePasswordService) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (fakePasswordService) Compare(hashedPassword, password string) error {
	if hashedPassword != "hashed:"+password {
		return errors.New("mismatch")
	}
	return nil
}

// fakeJWTService returns a deterministic "token" so tests can assert on it
// without depending on a real signing implementation.
type fakeJWTService struct{}

func (fakeJWTService) GenerateToken(user domain.User) (string, error) {
	return "token-for-" + user.Username, nil
}

func (fakeJWTService) ParseToken(tokenStr string) (*domain.Claims, error) {
	return nil, errors.New("not implemented in fake")
}

// --- tests ---------------------------------------------------------------

func TestUserUsecase_Register_FirstUserIsAdmin(t *testing.T) {
	repo := newFakeUserRepository()
	uc := NewUserUsecase(repo, fakePasswordService{}, fakeJWTService{})

	user, err := uc.Register(context.Background(), domain.RegisterRequest{Username: "alice", Password: "password123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Role != domain.RoleAdmin {
		t.Fatalf("expected first user to be admin, got role %q", user.Role)
	}
	if user.Password != "" {
		t.Fatalf("expected password hash to be cleared before returning, got %q", user.Password)
	}
}

func TestUserUsecase_Register_SecondUserIsRegular(t *testing.T) {
	repo := newFakeUserRepository()
	uc := NewUserUsecase(repo, fakePasswordService{}, fakeJWTService{})

	if _, err := uc.Register(context.Background(), domain.RegisterRequest{Username: "alice", Password: "password123"}); err != nil {
		t.Fatalf("unexpected error registering first user: %v", err)
	}
	user, err := uc.Register(context.Background(), domain.RegisterRequest{Username: "bob", Password: "password123"})
	if err != nil {
		t.Fatalf("unexpected error registering second user: %v", err)
	}
	if user.Role != domain.RoleUser {
		t.Fatalf("expected second user to be a regular user, got role %q", user.Role)
	}
}

func TestUserUsecase_Register_DuplicateUsername(t *testing.T) {
	repo := newFakeUserRepository()
	uc := NewUserUsecase(repo, fakePasswordService{}, fakeJWTService{})

	if _, err := uc.Register(context.Background(), domain.RegisterRequest{Username: "alice", Password: "password123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := uc.Register(context.Background(), domain.RegisterRequest{Username: "alice", Password: "different"})
	if !errors.Is(err, domain.ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestUserUsecase_Login_Success(t *testing.T) {
	repo := newFakeUserRepository()
	uc := NewUserUsecase(repo, fakePasswordService{}, fakeJWTService{})

	if _, err := uc.Register(context.Background(), domain.RegisterRequest{Username: "alice", Password: "password123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	token, err := uc.Login(context.Background(), domain.LoginRequest{Username: "alice", Password: "password123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "token-for-alice" {
		t.Fatalf("unexpected token: %q", token)
	}
}

func TestUserUsecase_Login_WrongPassword(t *testing.T) {
	repo := newFakeUserRepository()
	uc := NewUserUsecase(repo, fakePasswordService{}, fakeJWTService{})

	if _, err := uc.Register(context.Background(), domain.RegisterRequest{Username: "alice", Password: "password123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := uc.Login(context.Background(), domain.LoginRequest{Username: "alice", Password: "wrong"})
	if !errors.Is(err, domain.ErrInvalidCreds) {
		t.Fatalf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestUserUsecase_Login_UnknownUsernameDoesNotLeak(t *testing.T) {
	repo := newFakeUserRepository()
	uc := NewUserUsecase(repo, fakePasswordService{}, fakeJWTService{})

	_, err := uc.Login(context.Background(), domain.LoginRequest{Username: "ghost", Password: "whatever"})
	if !errors.Is(err, domain.ErrInvalidCreds) {
		t.Fatalf("expected ErrInvalidCreds (not ErrUserNotFound) for unknown username, got %v", err)
	}
}

func TestUserUsecase_Promote(t *testing.T) {
	repo := newFakeUserRepository()
	uc := NewUserUsecase(repo, fakePasswordService{}, fakeJWTService{})

	if _, err := uc.Register(context.Background(), domain.RegisterRequest{Username: "alice", Password: "password123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := uc.Register(context.Background(), domain.RegisterRequest{Username: "bob", Password: "password123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := uc.Promote(context.Background(), "bob"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.users["bob"].Role != domain.RoleAdmin {
		t.Fatalf("expected bob to be promoted to admin")
	}
}

func TestUserUsecase_Promote_UnknownUser(t *testing.T) {
	repo := newFakeUserRepository()
	uc := NewUserUsecase(repo, fakePasswordService{}, fakeJWTService{})

	err := uc.Promote(context.Background(), "ghost")
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
