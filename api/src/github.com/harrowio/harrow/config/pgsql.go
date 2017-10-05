package config

import (
	"net/url"
	"os"
	"time"

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

	// Try and inject os.Args[0] (executable name) into the application_name conn parameter
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("application_name", os.Args[0])
	u.RawQuery = q.Encode()

	db, err := sqlx.Open("postgres", u.String())
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
