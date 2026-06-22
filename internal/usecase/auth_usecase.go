package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/fangimal/TeamTask/internal/domain"
	jwtpkg "github.com/fangimal/TeamTask/pkg/jwt"
	"github.com/fangimal/TeamTask/pkg/password"
)

const minPasswordLength = 6

type AuthUseCase struct {
	users         domain.UserRepository
	jwtSecret     string
	jwtExpiration time.Duration
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type LoginResult struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

func NewAuthUseCase(users domain.UserRepository, jwtSecret string, jwtExpiration time.Duration) *AuthUseCase {
	return &AuthUseCase{
		users:         users,
		jwtSecret:     jwtSecret,
		jwtExpiration: jwtExpiration,
	}
}

func (useCase *AuthUseCase) Register(ctx context.Context, input RegisterInput) (*AuthUser, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, err
	}

	if err = validatePassword(input.Password); err != nil {
		return nil, err
	}

	existingUser, err := useCase.users.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	passwordHash, err := password.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: passwordHash,
	}
	if err = useCase.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return &AuthUser{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

func (useCase *AuthUseCase) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := useCase.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}

		return nil, fmt.Errorf("get user by email: %w", err)
	}

	if !password.Compare(user.PasswordHash, input.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := jwtpkg.Generate(user.ID, user.Email, useCase.jwtSecret, useCase.jwtExpiration)
	if err != nil {
		return nil, fmt.Errorf("generate jwt: %w", err)
	}

	return &LoginResult{
		Token: token,
		User: AuthUser{
			ID:    user.ID,
			Email: user.Email,
		},
	}, nil
}

func normalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return "", domain.ErrInvalidInput
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return "", domain.ErrInvalidInput
	}

	return email, nil
}

func validatePassword(plainPassword string) error {
	if len(plainPassword) < minPasswordLength {
		return domain.ErrInvalidInput
	}

	return nil
}
