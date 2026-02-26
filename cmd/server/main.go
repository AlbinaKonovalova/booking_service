package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	httpserver "github.com/AlbinaKonovalova/booking_service/internal/adapters/http"
	"github.com/AlbinaKonovalova/booking_service/internal/adapters/http/handlers"
	"github.com/AlbinaKonovalova/booking_service/internal/adapters/repository/postgres"
	"github.com/AlbinaKonovalova/booking_service/internal/adapters/scheduler"
	"github.com/AlbinaKonovalova/booking_service/internal/application"
	"github.com/AlbinaKonovalova/booking_service/internal/config"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config file")
	flag.Parse()

	logger := setupLogger("info")
	logger.Info("starting application")

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	logger = setupLogger(cfg.Log.Level)

	db, err := setupDatabase(cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("connected to database")

	// Hotel timezone
	hotelTZ, err := time.LoadLocation(cfg.Hotel.Timezone)
	if err != nil {
		logger.Error("failed to load hotel timezone", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("hotel timezone loaded", slog.String("timezone", cfg.Hotel.Timezone))

	// Repositories
	resourceRepo := postgres.NewResourceRepository(db)
	bookingRepo := postgres.NewBookingRepository(db)
	txManager := postgres.NewTxManager(db)

	// Application services
	resourceService := application.NewResourceService(resourceRepo, bookingRepo, txManager)
	bookingService := application.NewBookingService(bookingRepo, resourceRepo, txManager, hotelTZ)

	// HTTP handlers
	resourceHandler := handlers.NewResourceHandler(resourceService)
	bookingHandler := handlers.NewBookingHandler(bookingService)

	expirationService := application.NewExpirationService(bookingRepo, logger)
	sched, err := scheduler.NewScheduler(expirationService, cfg.Scheduler.ExpirationInterval, cfg.Scheduler.CompletionTime, logger)
	if err != nil {
		logger.Error("failed to create scheduler", slog.Any("error", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sched.Start(ctx)

	server := httpserver.NewServer(
		cfg.Server.Port,
		cfg.Server.ReadTimeout,
		cfg.Server.WriteTimeout,
		logger,
		resourceHandler,
		bookingHandler,
	)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", slog.Any("error", err))
	}
	
	logger.Info("server stopped")
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})

	return slog.New(handler)
}

func setupDatabase(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
