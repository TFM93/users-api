package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"users/config"
	"users/internal/app"
	"users/internal/controller/grpc"
	"users/internal/controller/http"
	"users/internal/domain"
	"users/internal/infra/notification"
	"users/internal/infra/outbox"
	repo "users/internal/infra/postgresql"
	"users/pkg/grpcserver"
	"users/pkg/httpserver"
	"users/pkg/logger"
	postgres "users/pkg/postgresql"
	"users/pkg/pubsub"
)

func main() {
	// -------------------------------------------------------------------------
	// Configuration

	configPath := flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.NewConfig(*configPath)
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}
	l := logger.New(cfg.LogLevel)

	if err := run(cfg, l); err != nil {
		l.Error("Run error: %s", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config, l logger.Interface) error {

	// -------------------------------------------------------------------------
	// Setup Infra

	pg, err := postgres.New(
		cfg.PG.DSN,
		postgres.MaxPoolSize(cfg.PG.PoolMax),
		postgres.AutoMigrate(true, "../../migrations/postgresql"),
		postgres.WithLogger(l))

	if err != nil {
		return fmt.Errorf("postgres.New: %w", err)
	}
	defer pg.Close()

	txSupplier := repo.NewTransactionSupplier(pg)
	outboxRepoCommands := repo.NewOutboxCommandsRepo(pg, l)

	pubsubClient, err := pubsub.New(cfg.PubSub.Enabled, cfg.PubSub.ProjectID, pubsub.WithLogger(l))
	if err != nil {
		return fmt.Errorf("pubsub.New: %w", err)
	}

	var notificationService domain.NotificationService
	if cfg.PubSub.Enabled {
		notificationService = notification.NewPubSubNotifierService(pubsubClient, l, cfg.PubSub.UsersTopic)
	} else {
		notificationService = notification.NewLoggerNotifierService(l)
	}

	outboxProcessor := outbox.NewProcessor(l, txSupplier, outboxRepoCommands, notificationService)
	interval := time.Duration(cfg.Notifications.Interval) * time.Second
	go outboxProcessor.StartScheduleProcess(context.Background(), interval, cfg.Notifications.MaxBatchSize)

	// -------------------------------------------------------------------------
	// Setup Service Layer

	healthCheckQueries := app.NewHealthCheckQueries(pg, pubsubClient)
	userServiceCommands := app.NewUserServiceCommands(l, txSupplier, repo.NewUserCommandsRepo(pg, l), outboxRepoCommands)
	userServiceQueries := app.NewUserServiceQueries(l, repo.NewUserQueriesRepo(pg, l))

	// -------------------------------------------------------------------------
	// Setup Controller Layer

	httpEngine, err := http.Setup(l, cfg.GRPC.Port, healthCheckQueries)
	if err != nil {
		return fmt.Errorf("httpServer.Setup: %w", err)
	}

	settedUpServer, err := grpc.Setup(l, userServiceCommands, userServiceQueries)
	if err != nil {
		return fmt.Errorf("grpcServer.Setup: %w", err)
	}

	// -------------------------------------------------------------------------
	// Start API Servers

	grpcServer := grpcserver.New(settedUpServer, grpcserver.Port(cfg.GRPC.Port), grpcserver.WithLogger(l))
	httpServer := httpserver.New(httpEngine, httpserver.Port(cfg.HTTP.Port), httpserver.WithLogger(l))

	// -------------------------------------------------------------------------
	// Shutdown

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		l.Info("received run signal: " + s.String())
	case err = <-grpcServer.Notify():
		l.Error(fmt.Errorf("run - grpcServer.Notify: %w", err))
	case err = <-httpServer.Notify():
		l.Error(fmt.Errorf("run - httpServer.Notify: %w", err))
	}

	err = httpServer.Shutdown()
	if err != nil {
		return fmt.Errorf("httpServer.Shutdown: %w", err)
	}
	grpcServer.GracefulStop()
	outboxProcessor.GracefulStop()
	return nil
}
