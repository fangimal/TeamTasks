package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fangimal/TeamTask/internal/domain"
	"github.com/go-sql-driver/mysql"
)

const duplicateEntryErrorNumber uint16 = 1062

type UserRepository struct {
	database *sql.DB
}

func NewUserRepository(database *sql.DB) *UserRepository {
	return &UserRepository{
		database: database,
	}
}

func (repository *UserRepository) Ping(ctx context.Context) error {
	if err := repository.database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	return nil
}

func (repository *UserRepository) Create(ctx context.Context, user *domain.User) error {
	result, err := repository.database.ExecContext(
		ctx,
		`INSERT INTO users (email, password_hash) VALUES (?, ?)`,
		user.Email,
		user.PasswordHash,
	)
	if err != nil {
		if isDuplicateEntryError(err) {
			return domain.ErrUserAlreadyExists
		}

		return fmt.Errorf("insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get inserted user id: %w", err)
	}

	user.ID = id

	return nil
}

func (repository *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}

	err := repository.database.QueryRowContext(
		ctx,
		`SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, fmt.Errorf("select user by email: %w", err)
	}

	return user, nil
}

func isDuplicateEntryError(err error) bool {
	var mysqlError *mysql.MySQLError
	return errors.As(err, &mysqlError) && mysqlError.Number == duplicateEntryErrorNumber
}
