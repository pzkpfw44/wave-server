package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/pzkpfw44/wave-server/internal/api"
	"github.com/pzkpfw44/wave-server/internal/api/handlers"
	"github.com/pzkpfw44/wave-server/internal/config"
	"github.com/pzkpfw44/wave-server/internal/repository"
	"github.com/pzkpfw44/wave-server/internal/service"
	"github.com/pzkpfw44/wave-server/pkg/health"
)

// TestEnvironment holds test dependencies
type TestEnvironment struct {
	Logger   *zap.Logger
	Config   *config.Config
	DB       *repository.Database
	UserRepo *repository.UserRepository
	Security *security
	Services *services
}

type security struct {
	HashUsername func(string) string
}

type services struct {
	Auth    *service.AuthService
	User    *service.UserService
	Message *service.MessageService
	Contact *service.ContactService
	Account *service.AccountService
}

var testEnv *TestEnvironment

// setupTestServer sets up a test server for integration tests
func setupTestServer(t *testing.T) (*httptest.Server, func()) {
	if testEnv == nil {
		// Create test environment only once
		logger := zaptest.NewLogger(t)

		// Load test config
		cfg := &config.Config{}
		cfg.Server.Port = 0 // Use any available port
		cfg.Database.Host = "localhost"
		cfg.Database.Port = 5433
		cfg.Database.User = "yugabyte"
		cfg.Database.Password = "yugabyte"
		cfg.Database.Name = "wave_test"
		cfg.Auth.JWTSecret = "test-secret-key"
		cfg.Auth.TokenExpiry = 15 * 60 // 15 minutes

		// Connect to database
		ctx := context.Background()
		db, err := repository.New(ctx, cfg, logger)
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}

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

		testEnv = &TestEnvironment{
			Logger:   logger,
			Config:   cfg,
			DB:       db,
			UserRepo: userRepo,
			Security: &security{
				HashUsername: security.HashUsername,
			},
			Services: &services{
				Auth:    authService,
				User:    userService,
				Message: messageService,
				Contact: contactService,
				Account: accountService,
			},
		}
	}

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Create handler
	h := &handlers.Handler{
		Auth: &handlers.AuthHandler{
			AuthService: testEnv.Services.Auth,
			userService: testEnv.Services.User,
			config:      testEnv.Config,
			logger:      testEnv.Logger,
		},
		Message: handlers.NewMessageHandler(
			testEnv.Services.Message,
			testEnv.Services.User,
			testEnv.Logger,
		),
		Contact: handlers.NewContactHandler(
			testEnv.Services.Contact,
			testEnv.Logger,
		),
		Key: handlers.NewKeyHandler(
			testEnv.Services.User,
			testEnv.Logger,
		),
		Account: handlers.NewAccountHandler(
			testEnv.Services.Account,
			testEnv.Services.Auth,
			testEnv.Logger,
		),
		logger: testEnv.Logger,
	}

	// Create health checker
	healthChecker := health.New(testEnv.DB.Pool, testEnv.Logger)

	// Configure routes
	api.SetupRoutes(e, h, testEnv.Config, testEnv.Services.Auth, healthChecker, testEnv.Logger)

	// Start server
	ts := httptest.NewServer(e)

	// Return server and cleanup function
	return ts, func() {
		ts.Close()
	}
}

// NewJSONBody creates a new JSON request body
func NewJSONBody(data interface{}) io.Reader {
	b, _ := json.Marshal(data)
	return bytes.NewReader(b)
}

// ReadJSONBody reads and parses a JSON response body
func ReadJSONBody(t *testing.T, body io.Reader, v interface{}) {
	b, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	err = json.Unmarshal(b, v)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}
}
