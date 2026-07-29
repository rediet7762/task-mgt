package infrastructure

import "golang.org/x/crypto/bcrypt"

// PasswordService implements domain.PasswordService using bcrypt.
type PasswordService struct{}

// NewPasswordService constructs a PasswordService.
func NewPasswordService() *PasswordService {
	return &PasswordService{}
}

// Hash returns the bcrypt hash of password using the library's default cost.
func (p *PasswordService) Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// Compare reports whether password matches hashedPassword, returning a
// non-nil error if it does not.
func (p *PasswordService) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
