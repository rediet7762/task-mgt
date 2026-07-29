package usecases

import (
	"context"
	"errors"
	"time"

	"task-manager/Domain"
)

// userUsecase implements domain.UserUsecase. It orchestrates
// account/authentication business logic and depends only on domain
// interfaces (repository + infrastructure services), never on concrete
// implementations like MongoDB, bcrypt, or a specific JWT library.
type userUsecase struct {
	userRepo    domain.UserRepository
	passwordSvc domain.PasswordService
	jwtSvc      domain.JWTService
}

// NewUserUsecase constructs a domain.UserUsecase backed by the given
// repository and infrastructure services.
func NewUserUsecase(userRepo domain.UserRepository, passwordSvc domain.PasswordService, jwtSvc domain.JWTService) domain.UserUsecase {
	return &userUsecase{
		userRepo:    userRepo,
		passwordSvc: passwordSvc,
		jwtSvc:      jwtSvc,
	}
}

// Register enforces the account-creation business rules: usernames must be
// unique, and the very first user ever created (i.e. the repository is
// empty at registration time) is automatically granted the admin role;
// everyone after that starts out as a regular user.
func (u *userUsecase) Register(ctx context.Context, req domain.RegisterRequest) (*domain.User, error) {
	existing, err := u.userRepo.CountByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if existing > 0 {
		return nil, domain.ErrUsernameTaken
	}

	total, err := u.userRepo.CountAll(ctx)
	if err != nil {
		return nil, err
	}

	role := domain.RoleUser
	if total == 0 {
		role = domain.RoleAdmin
	}

	hashed, err := u.passwordSvc.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Username:  req.Username,
		Password:  hashed,
		Role:      role,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	user.Password = "" // never hand the hash back to the caller
	return user, nil
}

// Login verifies the supplied credentials and, if valid, returns a signed
// JWT the client can use on subsequent requests.
func (u *userUsecase) Login(ctx context.Context, req domain.LoginRequest) (string, error) {
	user, err := u.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Don't reveal whether the username exists.
			return "", domain.ErrInvalidCreds
		}
		return "", err
	}

	if err := u.passwordSvc.Compare(user.Password, req.Password); err != nil {
		return "", domain.ErrInvalidCreds
	}

	return u.jwtSvc.GenerateToken(*user)
}

// Promote grants the admin role to the given username. Authorization (only
// an existing admin may call this) is enforced by AdminOnly middleware at
// the delivery layer, not here.
func (u *userUsecase) Promote(ctx context.Context, username string) error {
	return u.userRepo.UpdateRole(ctx, username, domain.RoleAdmin)
}
