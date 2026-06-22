package mysql

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	database *sql.DB
}

func NewRepository(database *sql.DB) *Repository {
	return &Repository{
		database: database,
	}
}

func (repository *Repository) Ping(ctx context.Context) error {
	if err := repository.database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql: %w", err)
	}

	return nil
}
