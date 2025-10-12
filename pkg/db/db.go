// Package db предоставляет функциональность для работы с базой данных,
// включая подключение и применение миграций.
package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DBinstance глобальная переменная для хранения подключения к базе данных.
var DBinstance *sql.DB

// New создает новое подключение к базе данных и применяет миграции.
// connStr - строка подключения к базе данных.
// migrationsPath - путь к директории с файлами миграций.
// Возвращает подключение к базе данных или ошибку.
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

// applyMigrationsWithNewConnection применяет миграции к базе данных,
// используя отдельное подключение.
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
