package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/fangimal/TeamTask/internal/config"
	mysqlrepo "github.com/fangimal/TeamTask/internal/repository/mysql"
)

const (
	configPath     = "configs/config.yaml"
	migrationsPath = "file://migrations"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: teamtasks-migrate up|down")
		os.Exit(1)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	switch os.Args[1] {
	case "up":
		err = mysqlrepo.RunMigrations(cfg.Database, migrationsPath)
	case "down":
		err = mysqlrepo.RollbackMigrations(cfg.Database, migrationsPath)
	default:
		fmt.Fprintln(os.Stderr, "usage: teamtasks-migrate up|down")
		os.Exit(1)
	}

	if err != nil {
		slog.New(slog.NewJSONHandler(os.Stderr, nil)).Error("migration failed", slog.Any("error", err))
		os.Exit(1)
	}
}
