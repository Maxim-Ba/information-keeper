package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var DBinstance *sql.DB

func New(connStr, migrationsPath string) (*sql.DB, error) {

	pool, err := sql.Open("pgx", connStr)

	if err != nil {
		return nil, err
	}
	if err := pool.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database 1: %w", err)
	}
	pool.SetMaxOpenConns(10)

	instance := pool
	DBinstance = pool
	err = applyMigrationsWithNewConnection(connStr, migrationsPath)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database 2: %w", err)
	}
	return instance, nil
}

func applyMigrationsWithNewConnection(connStr, migrationsPath string) error {
    migrationDB, err := sql.Open("pgx", connStr)
    if err != nil {
        return err
    }
    defer migrationDB.Close()

    if err := migrationDB.Ping(); err != nil {
        return fmt.Errorf("failed to ping migration database: %w", err)
    }

    driver, err := postgres.WithInstance(migrationDB, &postgres.Config{})
    if err != nil {
        return err
    }

    m, err := migrate.NewWithDatabaseInstance(
        migrationsPath,
        "postgres", driver)
    if err != nil {
        return err
    }
    defer m.Close()

    err = m.Up()
    if err != nil && err != migrate.ErrNoChange {
        return err
    }

    return nil
}
