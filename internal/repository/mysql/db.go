package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fangimal/TeamTask/internal/config"
	"github.com/go-sql-driver/mysql"
)

func NewDB(ctx context.Context, config config.DatabaseConfig) (*sql.DB, error) {
	database, err := sql.Open("mysql", DSN(config))
	if err != nil {
		return nil, fmt.Errorf("open mysql connection: %w", err)
	}

	database.SetMaxOpenConns(config.MaxOpenConns)
	database.SetMaxIdleConns(config.MaxIdleConns)
	database.SetConnMaxLifetime(config.ConnMaxLifetime)

	pingContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err = database.PingContext(pingContext); err != nil {
		database.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return database, nil
}

func DSN(config config.DatabaseConfig) string {
	dsnConfig := mysql.Config{
		User:                 config.User,
		Passwd:               config.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", config.Host, config.Port),
		DBName:               config.Name,
		ParseTime:            true,
		AllowNativePasswords: true,
		MultiStatements:      true,
	}

	return dsnConfig.FormatDSN()
}
