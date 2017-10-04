package config

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func (c *Config) PgDataSourceName() (string, error) {
	return getEnvWithDefault("HAR_PGSQL_DSN", "postgres://harrow-dev-test:harrow-dev-test@localhost:5432/harrow-dev-test?sslmode=disable"), nil
}

func (c *Config) DB() (*sqlx.DB, error) {

	dsn, err := c.PgDataSourceName()
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)

	return db, nil
}
