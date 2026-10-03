package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"Be-Crustea/internal/model"
	"Be-Crustea/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrJWTSecretMissing   = errors.New("JWT secret is not configured")
)

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, *model.User, error)
	Register(ctx context.Context, name, email, password string) (*model.User, error)
	GetCurrentUser(ctx context.Context, userID int64) (*model.User, error)
}

type authService struct {
	userRepository repository.UserRepository
	jwtSecret      string
}

func NewAuthService(
	userRepository repository.UserRepository,
	jwtSecret string,
) AuthService {
	return &authService{
		userRepository: userRepository,
		jwtSecret:      jwtSecret,
	}
}

func (s *authService) Register(
	ctx context.Context,
	name, email, password string,
) (*model.User, error) {
	existingUser, err := s.userRepository.GetByEmail(ctx, email)

	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.userRepository.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.Password = ""

	return user, nil
}

func (s *authService) Login(
	ctx context.Context,
	email, password string,
) (string, *model.User, error) {
	if s.jwtSecret == "" {
		return "", nil, ErrJWTSecretMissing
	}

	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(password),
	); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	user.Password = ""

	return token, user, nil
}

func (s *authService) GetCurrentUser(
	ctx context.Context,
	userID int64,
) (*model.User, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}

		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	user.Password = ""

	return user, nil
}

func (s *authService) generateToken(user *model.User) (string, error) {
	now := time.Now()
	expiration := now.Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"name":  user.Name,
		"iat":   now.Unix(),
		"exp":   expiration.Unix(),
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(s.jwtSecret))
}
