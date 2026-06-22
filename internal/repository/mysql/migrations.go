package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/fangimal/TeamTask/internal/config"
	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(config config.DatabaseConfig, sourceURL string) error {
	migration, err := newMigration(config, sourceURL)
	if err != nil {
		return err
	}
	defer migration.Close()

	if err = migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func RollbackMigrations(config config.DatabaseConfig, sourceURL string) error {
	migration, err := newMigration(config, sourceURL)
	if err != nil {
		return err
	}
	defer migration.Close()

	if err = migration.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback migrations: %w", err)
	}

	return nil
}

func newMigration(config config.DatabaseConfig, sourceURL string) (*migrate.Migrate, error) {
	database, err := sql.Open("mysql", DSN(config))
	if err != nil {
		return nil, fmt.Errorf("open mysql for migrations: %w", err)
	}

	driver, err := migratemysql.WithInstance(database, &migratemysql.Config{})
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("create mysql migration driver: %w", err)
	}

	migration, err := migrate.NewWithDatabaseInstance(sourceURL, config.Name, driver)
	if err != nil {
		database.Close()
		return nil, fmt.Errorf("create migration instance: %w", err)
	}

	return migration, nil
}
