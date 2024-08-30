// Package postgresql implements and sets up a postgres connection.
package postgresql

import (
	"context"
	"fmt"
	"time"
	log "users/pkg/logger"

	migrate "github.com/golang-migrate/migrate/v4"
	mpg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Interface defines the contract for managing a database connection
// It provides methods to access the database provider from the pool and to close the connection pool
type Interface interface {
	// GetPool returns the db provider methods to query the database
	GetPool() DBProvider

	// Close closes the connection pool
	Close()

	// Ping returns true if pool's ping does not return an error
	Ping(ctx context.Context) bool

	// IsEnabled returns true if postgres instance is enabled
	// will return true in this version
	IsEnabled() bool
}

// postgres represents a PostgreSQL database connection manager with configuration settings
type postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	autoMigrate  bool
	migrationDir string

	Pool *pgxpool.Pool
	l    log.Interface
}

// New handles the postgres initialization, connection and migrations
func New(dsn string, opts ...Option) (Interface, error) {
	pg := &postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	for _, opt := range opts {
		opt(pg)
	}

	if pg.l == nil {
		pg.l = log.New("")
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgresql parse config: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			err = pg.Pool.Ping(context.Background())
			if err == nil {
				break
			}
		}

		pg.l.Warn("Postgres is trying to connect, attempts left: %d", pg.connAttempts)
		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("exceeded attempts: %w", err)
	}

	if pg.autoMigrate {
		if err := pg.migrate(); err != nil {
			return nil, err
		}
	}

	return pg, nil
}

func (p *postgres) migrate() error {
	if p.migrationDir == "" {
		return fmt.Errorf("postgresql-migrate: no migrationDir found")
	}
	driver, err := mpg.WithInstance(stdlib.OpenDBFromPool(p.Pool), &mpg.Config{})
	if err != nil {
		return fmt.Errorf("postgresql-migrate: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", p.migrationDir),
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("postgresql-migrate-New: %w", err)
	}
	defer m.Close()
	if err = m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			p.l.Info("Postgresql-migrate: no new migrations to apply")
		} else {
			return fmt.Errorf("postgresql-migrate: %w", err)
		}
	}
	p.l.Info("Postgresql-migrate: migrations ran successfully")
	return nil
}

// Close closes the connection pool
func (p *postgres) Close() {
	if p.Pool != nil {
		p.l.Debug("Postgresql-Close: %d connections in use, blocking until they are released", p.Pool.Stat().AcquiredConns())
		p.Pool.Close()
	}
	p.l.Debug("Postgresql-Close: connection closed")
}

// GetPool returns the db provider methods to query the database
func (p *postgres) GetPool() DBProvider {
	return p.Pool
}

// IsEnabled returns true if the postgres instance is enabled
// in this version, it returns always true
func (p *postgres) IsEnabled() bool {
	return true
}

// Ping returns true if pool's ping does not return an error
func (p *postgres) Ping(ctx context.Context) bool {
	return p.Pool.Ping(ctx) == nil
}

// Option allows to configure postgresql connection
type Option func(*postgres)

// MaxPoolSize configures the size of the postgresql connection pool
func MaxPoolSize(size int) Option {
	return func(c *postgres) {
		c.maxPoolSize = size
	}
}

// ConnAttempts configures the postgresql connection attempts
func ConnAttempts(attempts int) Option {
	return func(c *postgres) {
		c.connAttempts = attempts
	}
}

// ConnTimeout configures the postgresql connection timeout duration
func ConnTimeout(timeout time.Duration) Option {
	return func(c *postgres) {
		c.connTimeout = timeout
	}
}

// AutoMigrate runs the migration at the client connection
func AutoMigrate(enabled bool, dir string) Option {
	return func(c *postgres) {
		c.autoMigrate = enabled
		c.migrationDir = dir
	}
}

// WithLogger injects the logger dependency
func WithLogger(logger log.Interface) Option {
	return func(c *postgres) {
		c.l = logger
	}
}
