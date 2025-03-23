package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/pzkpfw44/wave-server/internal/api"
	"github.com/pzkpfw44/wave-server/internal/api/handlers"
	"github.com/pzkpfw44/wave-server/internal/api/middleware"
	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/internal/repository"
	"github.com/pzkpfw44/wave-server/internal/security"
	"github.com/pzkpfw44/wave-server/internal/service"
	"github.com/pzkpfw44/wave-server/pkg/health"
)

// TestEnv represents the test environment
type TestEnv struct {
	Server      *httptest.Server
	Echo        *echo.Echo
	Config      *config.Config
	Logger      *zap.Logger
	UserRepo    *repository.UserRepository
	TokenRepo   *repository.TokenRepository
	Security    *security.SecurityTestHelper
	AuthService *service.AuthService
}

var testEnv *TestEnv

// setupTestServer sets up a test server and returns it along with a cleanup function
func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	// If test environment already exists, reuse it
	if testEnv != nil {
		return testEnv.Server, func() {}
	}

	// Create test config
	cfg := &config.Config{}
	cfg.Server.Port = 8081
	cfg.Auth.JWTSecret = "test_secret_key"
	cfg.Auth.TokenExpiry = 24 * time.Hour

	// Create logger
	logger := zaptest.NewLogger(t)

	// Create an in-memory database for testing
	db := setupTestDatabase(t)

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	contactRepo := repository.NewContactRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// Create services
	userService := service.NewUserService(userRepo, logger)
	authService := service.NewAuthService(userRepo, tokenRepo, cfg, logger)
	messageService := service.NewMessageService(messageRepo, userRepo, logger)
	contactService := service.NewContactService(contactRepo, logger)
	accountService := service.NewAccountService(userRepo, contactRepo, messageRepo, tokenRepo, logger)

	// Create handlers
	h := handlers.NewHandler(db, cfg, logger)

	// Create Echo instance
	e := echo.New()
	e.Validator = middleware.NewValidationMiddleware(logger).GetValidator()

	// Configure routes
	healthChecker := health.New(db.Pool, logger)
	api.SetupRoutes(e, h, cfg, authService, healthChecker, logger)

	// Create test server
	server := httptest.NewServer(e)

	// Store test environment
	testEnv = &TestEnv{
		Server:      server,
		Echo:        e,
		Config:      cfg,
		Logger:      logger,
		UserRepo:    userRepo,
		TokenRepo:   tokenRepo,
		Security:    &security.SecurityTestHelper{}, // Mock security helper for tests
		AuthService: authService,
	}

	// Return cleanup function
	cleanup := func() {
		server.Close()
	}

	return server, cleanup
}

// setupTestDatabase creates an in-memory test database
func setupTestDatabase(t *testing.T) *repository.Database {
	// This is a simplified version - in a real implementation,
	// you would use an actual in-memory database like SQLite

	// Here we're just returning a mock DB object for testing
	pool := &mockDBPool{t: t}

	return &repository.Database{
		Pool:   pool,
		Logger: zaptest.NewLogger(t),
		Config: &config.Config{},
	}
}

// mockDBPool is a simplified mock for pgxpool.Pool
type mockDBPool struct {
	t *testing.T
}

// Close implements the Pool.Close method
func (m *mockDBPool) Close() {}

// Ping implements the Pool.Ping method
func (m *mockDBPool) Ping(ctx context.Context) error {
	return nil
}

// Exec, QueryRow, and Query would be implemented here for a complete mock

// Helper functions for tests

// NewJSONBody creates a new JSON request body
func NewJSONBody(body interface{}) io.Reader {
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(b)
}

// ReadJSONBody reads a JSON response body
func ReadJSONBody(t *testing.T, body io.Reader, v interface{}) {
	b, err := io.ReadAll(body)
	require.NoError(t, err)

	err = json.Unmarshal(b, v)
	require.NoError(t, err)
}
