package services

import (
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	HashPassword(password string) (string, error)
	CheckPassword(hash, password string) error
	GenerateToken(userID uuid.UUID, email, role string) (string, error)
}

type authService struct{}

func NewAuthService() AuthService {
	return &authService{}
}

func (s *authService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func (s *authService) CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (s *authService) GenerateToken(userID uuid.UUID, email, role string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET environment variable is not set")
	}

	// Read expiry from env, default 24 hours
	expireHours := 24
	if h := os.Getenv("JWT_EXPIRE_HOURS"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil {
			expireHours = parsed
		}
	}

	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * time.Duration(expireHours)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
