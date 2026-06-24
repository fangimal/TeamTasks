package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	mysqlDriver "github.com/go-sql-driver/mysql"
	tc "github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

type testContainer struct {
	database *sql.DB
	cleanup  func()
}

func setupTestContainer(t *testing.T) *testContainer {
	t.Helper()

	ctx := context.Background()

	mysqlContainer, err := tcmysql.RunContainer(ctx,
		tc.WithImage("mysql:8.0"),
		tcmysql.WithDatabase("teamtasks_test"),
		tcmysql.WithUsername("test"),
		tcmysql.WithPassword("test"),
	)
	if err != nil {
		t.Fatalf("failed to start mysql container: %v", err)
	}

	host, err := mysqlContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := mysqlContainer.MappedPort(ctx, "3306/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	dsnConfig := mysqlDriver.Config{
		User:            "test",
		Passwd:          "test",
		Net:             "tcp",
		Addr:            fmt.Sprintf("%s:%s", host, port.Port()),
		DBName:          "teamtasks_test",
		ParseTime:       true,
		MultiStatements: true,
	}

	database, err := sql.Open("mysql", dsnConfig.FormatDSN())
	if err != nil {
		mysqlContainer.Terminate(ctx)
		t.Fatalf("failed to open database: %v", err)
	}

	if err = database.PingContext(ctx); err != nil {
		database.Close()
		mysqlContainer.Terminate(ctx)
		t.Fatalf("failed to ping database: %v", err)
	}

	if err = applyMigrations(database); err != nil {
		database.Close()
		mysqlContainer.Terminate(ctx)
		t.Fatalf("failed to apply migrations: %v", err)
	}

	cleanup := func() {
		database.Close()
		mysqlContainer.Terminate(ctx)
	}

	return &testContainer{
		database: database,
		cleanup:  cleanup,
	}
}

func applyMigrations(database *sql.DB) error {
	migrationPath := filepath.Join("..", "..", "..", "migrations", "000001_init_schema.up.sql")
	if _, err := os.Stat(migrationPath); err != nil {
		migrationPath = filepath.Join("migrations", "000001_init_schema.up.sql")
		if _, err := os.Stat(migrationPath); err != nil {
			return fmt.Errorf("migration file not found: %w", err)
		}
	}

	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	if _, err = database.Exec(string(sqlBytes)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}
	return nil
}
