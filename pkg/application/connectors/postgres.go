package connectors

import (
	"context"
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log/slog"
	"sync"
)

type Postgres struct {
	db   *sql.DB
	DSN  string
	init sync.Once
}

func (p *Postgres) Client(ctx context.Context) *sql.DB {
	var err error
	p.init.Do(func() {
		p.db, err = sql.Open("postgres", p.DSN)
		if err != nil {
			logger(ctx).Error("Error connecting to Postgres", slog.String("error", err.Error()))
			panic(err)
		}
		err = p.db.PingContext(ctx)
		if err != nil {
			logger(ctx).Error("Error connecting to Postgres", slog.String("error", err.Error()))
			panic(err)
		}
		logger(ctx).Info("Successfully connected to Postgres database")
	})
	return p.db
}

func (p *Postgres) Close(ctx context.Context) {
	if err := p.db.Close(); err != nil {
		logger(ctx).Error("postgresClient.Close", slog.String("error", err.Error()))
	}

	logger(ctx).Info(
		"postgres disconnected",
		slog.String("database", p.DSN),
	)
}

func (p *Postgres) RunMigrations(ctx context.Context) error {
	m, err := migrate.New("file://db/migrations", p.DSN)
	if err != nil {
		logger(ctx).Error("Error running migrations",
			slog.String("error", err.Error()))
		return err
	}
	defer m.Close()
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger(ctx).Error("Error running migrations",
			slog.String("error", err.Error()))
		return err
	}
	logger(ctx).Info("Successfully migrated migrations")
	return nil
}
