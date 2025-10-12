package postgres

import "database/sql"

type PostgresInstance struct {
	conn *sql.DB
}

// New создает новый экземпляр PostgresInstance и возвращает его.
func New() (*PostgresInstance, error) {
	if conn, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres"); err != nil {
		return nil, err
	} else {
		return &PostgresInstance{conn}, nil
	}
}
