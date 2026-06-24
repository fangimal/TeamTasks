package usecase

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/fangimal/TeamTask/internal/domain"
	"github.com/fangimal/TeamTask/pkg/password"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister_Success(t *testing.T) {
	mockUsers := new(MockUserRepository)
	authUseCase := NewAuthUseCase(mockUsers, "test-secret", 24*time.Hour)

	mockUsers.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, sql.ErrNoRows)
	mockUsers.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
		return user.Email == "test@example.com" && user.PasswordHash != ""
	})).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*domain.User)
		user.ID = 1
	})

	authUser, err := authUseCase.Register(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), authUser.ID)
	assert.Equal(t, "test@example.com", authUser.Email)
	mockUsers.AssertExpectations(t)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	mockUsers := new(MockUserRepository)
	authUseCase := NewAuthUseCase(mockUsers, "test-secret", 24*time.Hour)

	existingUser := &domain.User{ID: 1, Email: "existing@example.com"}
	mockUsers.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)

	authUser, err := authUseCase.Register(context.Background(), RegisterInput{
		Email:    "existing@example.com",
		Password: "password123",
	})

	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
	assert.Nil(t, authUser)
	mockUsers.AssertExpectations(t)
}

func TestRegister_InvalidEmail(t *testing.T) {
	mockUsers := new(MockUserRepository)
	authUseCase := NewAuthUseCase(mockUsers, "test-secret", 24*time.Hour)

	authUser, err := authUseCase.Register(context.Background(), RegisterInput{
		Email:    "not-an-email",
		Password: "password123",
	})

	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Nil(t, authUser)
}

func TestRegister_ShortPassword(t *testing.T) {
	mockUsers := new(MockUserRepository)
	authUseCase := NewAuthUseCase(mockUsers, "test-secret", 24*time.Hour)

	authUser, err := authUseCase.Register(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "ab",
	})

	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	assert.Nil(t, authUser)
}

func TestLogin_Success(t *testing.T) {
	mockUsers := new(MockUserRepository)
	authUseCase := NewAuthUseCase(mockUsers, "test-secret", 24*time.Hour)

	passwordHash, err := password.Hash("password123")
	assert.NoError(t, err)

	storedUser := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
	}
	mockUsers.On("GetByEmail", mock.Anything, "test@example.com").Return(storedUser, nil)

	result, err := authUseCase.Login(context.Background(), LoginInput{
		Email:    "Test@Example.com",
		Password: "password123",
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, int64(1), result.User.ID)
	assert.Equal(t, "test@example.com", result.User.Email)
	mockUsers.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockUsers := new(MockUserRepository)
	authUseCase := NewAuthUseCase(mockUsers, "test-secret", 24*time.Hour)

	passwordHash, err := password.Hash("correct-password")
	assert.NoError(t, err)

	storedUser := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: passwordHash,
	}
	mockUsers.On("GetByEmail", mock.Anything, "test@example.com").Return(storedUser, nil)

	result, err := authUseCase.Login(context.Background(), LoginInput{
		Email:    "test@example.com",
		Password: "wrong-password",
	})

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	assert.Nil(t, result)
	mockUsers.AssertExpectations(t)
}

func TestRegister_DatabaseError(t *testing.T) {
	mockUsers := new(MockUserRepository)
	authUseCase := NewAuthUseCase(mockUsers, "test-secret", 24*time.Hour)

	mockUsers.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, assert.AnError)

	authUser, err := authUseCase.Register(context.Background(), RegisterInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Nil(t, authUser)
	mockUsers.AssertExpectations(t)
}

func TestLogin_DatabaseError(t *testing.T) {
	mockUsers := new(MockUserRepository)
	authUseCase := NewAuthUseCase(mockUsers, "test-secret", 24*time.Hour)

	mockUsers.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, assert.AnError)

	result, err := authUseCase.Login(context.Background(), LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	})

	assert.Error(t, err)
	assert.Nil(t, result)
	mockUsers.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockUsers := new(MockUserRepository)
	authUseCase := NewAuthUseCase(mockUsers, "test-secret", 24*time.Hour)

	mockUsers.On("GetByEmail", mock.Anything, "unknown@example.com").Return(nil, sql.ErrNoRows)

	result, err := authUseCase.Login(context.Background(), LoginInput{
		Email:    "unknown@example.com",
		Password: "password123",
	})

	assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
	assert.Nil(t, result)
	mockUsers.AssertExpectations(t)
}
