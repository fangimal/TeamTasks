package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Logger   LoggerConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type LoggerConfig struct {
	Level  string
	Format string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

func Load(path string) (*Config, error) {
	config := defaultConfig()

	values, err := readYAML(path)
	if err != nil {
		return nil, err
	}

	if err = applyYAML(config, values); err != nil {
		return nil, err
	}

	if err = applyEnv(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 5 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            "mysql",
			Port:            3306,
			User:            "teamtasks",
			Password:        "teamtasks",
			Name:            "teamtasks",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Redis: RedisConfig{
			Host: "redis",
			Port: 6379,
			DB:   0,
		},
		Logger: LoggerConfig{
			Level:  "info",
			Format: "json",
		},
		JWT: JWTConfig{
			Secret:     "change-me-in-env",
			Expiration: 24 * time.Hour,
		},
	}
}

func readYAML(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer file.Close()

	values := make(map[string]string)
	var section string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("invalid config line: %q", line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if value == "" {
			section = key
			continue
		}

		if section != "" {
			key = section + "." + key
		}
		values[key] = strings.Trim(value, `"'`)
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan config file: %w", err)
	}

	return values, nil
}

func applyYAML(config *Config, values map[string]string) error {
	return applyValues(config, values)
}

func applyEnv(config *Config) error {
	values := make(map[string]string)

	mapEnv(values, "server.host", "SERVER_HOST")
	mapEnv(values, "server.port", "SERVER_PORT")
	mapEnv(values, "server.read_timeout", "SERVER_READ_TIMEOUT")
	mapEnv(values, "server.write_timeout", "SERVER_WRITE_TIMEOUT")
	mapEnv(values, "server.idle_timeout", "SERVER_IDLE_TIMEOUT")
	mapEnv(values, "server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT")
	mapEnv(values, "database.host", "DATABASE_HOST")
	mapEnv(values, "database.port", "DATABASE_PORT")
	mapEnv(values, "database.user", "DATABASE_USER")
	mapEnv(values, "database.password", "DATABASE_PASSWORD")
	mapEnv(values, "database.name", "DATABASE_NAME")
	mapEnv(values, "database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	mapEnv(values, "database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	mapEnv(values, "database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")
	mapEnv(values, "redis.host", "REDIS_HOST")
	mapEnv(values, "redis.port", "REDIS_PORT")
	mapEnv(values, "redis.password", "REDIS_PASSWORD")
	mapEnv(values, "redis.db", "REDIS_DB")
	mapEnv(values, "logger.level", "LOGGER_LEVEL")
	mapEnv(values, "logger.format", "LOGGER_FORMAT")
	mapEnv(values, "jwt.secret", "JWT_SECRET")
	mapEnv(values, "jwt.expiration", "JWT_EXPIRATION")

	return applyValues(config, values)
}

func mapEnv(values map[string]string, configKey string, envKey string) {
	if value, ok := os.LookupEnv(envKey); ok {
		values[configKey] = value
	}
}

func applyValues(config *Config, values map[string]string) error {
	var errs []error

	for key, value := range values {
		switch key {
		case "server.host":
			config.Server.Host = value
		case "server.port":
			assignInt(&config.Server.Port, key, value, &errs)
		case "server.read_timeout":
			assignDuration(&config.Server.ReadTimeout, key, value, &errs)
		case "server.write_timeout":
			assignDuration(&config.Server.WriteTimeout, key, value, &errs)
		case "server.idle_timeout":
			assignDuration(&config.Server.IdleTimeout, key, value, &errs)
		case "server.shutdown_timeout":
			assignDuration(&config.Server.ShutdownTimeout, key, value, &errs)
		case "database.host":
			config.Database.Host = value
		case "database.port":
			assignInt(&config.Database.Port, key, value, &errs)
		case "database.user":
			config.Database.User = value
		case "database.password":
			config.Database.Password = value
		case "database.name":
			config.Database.Name = value
		case "database.max_open_conns":
			assignInt(&config.Database.MaxOpenConns, key, value, &errs)
		case "database.max_idle_conns":
			assignInt(&config.Database.MaxIdleConns, key, value, &errs)
		case "database.conn_max_lifetime":
			assignDuration(&config.Database.ConnMaxLifetime, key, value, &errs)
		case "redis.host":
			config.Redis.Host = value
		case "redis.port":
			assignInt(&config.Redis.Port, key, value, &errs)
		case "redis.password":
			config.Redis.Password = value
		case "redis.db":
			assignInt(&config.Redis.DB, key, value, &errs)
		case "logger.level":
			config.Logger.Level = strings.ToLower(value)
		case "logger.format":
			config.Logger.Format = strings.ToLower(value)
		case "jwt.secret":
			config.JWT.Secret = value
		case "jwt.expiration":
			assignDuration(&config.JWT.Expiration, key, value, &errs)
		default:
			return fmt.Errorf("unknown config key %q", key)
		}
	}

	return errors.Join(errs...)
}

func assignInt(target *int, key string, value string, errs *[]error) {
	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		*errs = append(*errs, fmt.Errorf("parse %s as int: %w", key, err))
		return
	}

	*target = parsedValue
}

func assignDuration(target *time.Duration, key string, value string, errs *[]error) {
	parsedValue, err := time.ParseDuration(value)
	if err != nil {
		*errs = append(*errs, fmt.Errorf("parse %s as duration: %w", key, err))
		return
	}

	*target = parsedValue
}
