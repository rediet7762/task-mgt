package infrastructure

import (
	"errors"
	"os"
	"time"

	"task-manager/Domain"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 24 * time.Hour

// jwtClaims is the library-specific claims shape golang-jwt requires
// (it must embed jwt.RegisteredClaims / implement jwt.Claims). Keeping this
// type private to the Infrastructure package means Domain and Usecases
// never need to import the JWT library themselves.
type jwtClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService implements domain.JWTService using the golang-jwt/jwt library
// and HMAC-signed tokens.
type JWTService struct {
	secret []byte
}

// NewJWTService constructs a JWTService. If secret is empty, a development
// fallback is used (see Env behavior below) - real deployments must set
// JWT_SECRET.
func NewJWTService(secret string) *JWTService {
	if secret == "" {
		secret = devSecretFallback()
	}
	return &JWTService{secret: []byte(secret)}
}

// devSecretFallback returns a fixed development-only secret. Anyone who
// knows this value could forge admin tokens, so real deployments must set
// JWT_SECRET in the environment.
func devSecretFallback() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	return "dev-secret-change-me"
}

// GenerateToken creates a signed JWT for the given user.
func (s *JWTService) GenerateToken(user domain.User) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ParseToken validates a raw JWT string and returns its claims translated
// into the framework-agnostic domain.Claims type.
func (s *JWTService) ParseToken(tokenStr string) (*domain.Claims, error) {
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	result := &domain.Claims{
		UserID:   claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
	}
	if claims.IssuedAt != nil {
		result.IssuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		result.ExpiresAt = claims.ExpiresAt.Time
	}
	return result, nil
}
