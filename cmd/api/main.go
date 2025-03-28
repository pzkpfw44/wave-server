package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pzkpfw44/wave-server/internal/api"
	"github.com/pzkpfw44/wave-server/internal/api/handlers"
	"github.com/pzkpfw44/wave-server/internal/api/middleware"
	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/internal/repository"
	"github.com/pzkpfw44/wave-server/internal/service"
	"github.com/pzkpfw44/wave-server/pkg/health"
	"github.com/pzkpfw44/wave-server/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	log, err := logger.New(cfg.LogLevel, cfg.IsDevelopment())
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync(log)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	db, err := repository.New(ctx, cfg, log)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(ctx); err != nil {
		log.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Create handlers
	h := handlers.NewHandler(db, cfg, log)

	// Create services for middleware and authentication
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	authService := service.NewAuthService(userRepo, tokenRepo, cfg, log)

	// Setup health checker
	healthChecker := health.New(db.Pool, log)

	// Configure middleware
	middleware.SetupMiddleware(e, cfg, log, authService)

	// Configure routes
	api.SetupRoutes(e, h, cfg, authService, healthChecker, log)

	// Start server
	go func() {
		address := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Info("Starting server", zap.String("address", address))
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Gracefully shutdown
	log.Info("Shutting down server...")
	ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown error", zap.Error(err))
	}

	log.Info("Server stopped")
}
