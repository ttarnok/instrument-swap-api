// Package testhelpers provides functionality for integration testing.
package testhelpers

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// migrateDatabase handles database migration.
// In case of project restructure, check the declaration of pathToMigrationFiles.
func migrateDatabase(databaseURL string) error {
	// get location of test
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get path")
	}
	pathToMigrationFiles := filepath.Join(filepath.Dir(path), "..", "..", "migrations")

	finalpathToMigrateFiles := fmt.Sprintf("file:%s", pathToMigrationFiles)

	m, err := migrate.New(finalpathToMigrateFiles, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// PostgresContainer represents a Postgres testing container and its connection string.
type PostgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

// CreatePostgresContainer creates a new PostgresConatiner.
func CreatePostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithInitScripts(filepath.Join("..", "..", "init-scripts", "init.sql")),
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(25*time.Second)),
	)
	if err != nil {
		return nil, err
	}
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	// do the database migration
	err = migrateDatabase(connStr)
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}, nil
}

// RedisContainer represents a redis test conatainer and its connection string.
type RedisContainer struct {
	*redis.RedisContainer
	ConnectionString string
}

// CreateRedisContainer creates a new RedisContainer.
func CreateRedisContainer(ctx context.Context) (*RedisContainer, error) {
	redisContainer, err := redis.Run(ctx,
		"redis:alpine",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
	)
	if err != nil {
		return nil, err
	}
	connStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	return &RedisContainer{
		RedisContainer:   redisContainer,
		ConnectionString: connStr,
	}, nil
}
