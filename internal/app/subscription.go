package app

import (
	"context"
	"effectiveMobile_test/internal/config"
	"effectiveMobile_test/internal/domain/service"
	"effectiveMobile_test/internal/infrastructure/persistent"
	"effectiveMobile_test/internal/server"
	"effectiveMobile_test/pkg/application/connectors"
	"effectiveMobile_test/pkg/contextx"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	logger := contextx.DefaultLogger
	ctx = contextx.WithLogger(ctx, logger)

	// Создание зависимостей с обработкой ошибок
	psql := connectors.Postgres{DSN: cfg.Postgres.DSN()}
	err = psql.RunMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	db := psql.Client(ctx)
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database", "error", err)
		}
	}()

	// Проверка соединения с БД
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Создание слоёв
	repo := persistent.NewPostgresRepo(db)
	service := service.NewSubscriptionService(repo)
	router := server.NewRouter(service)

	// Настройка HTTP сервера
	httpServer := &http.Server{
		Addr:         cfg.HTTP.Addr,
		Handler:      router,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting HTTP server", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		logger.Info("shutting down server gracefully...")
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	logger.Info("server stopped successfully")
	return nil
}
