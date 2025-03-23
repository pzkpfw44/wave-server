package integration

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/yourusername/wave-server/internal/api"
	"github.com/yourusername/wave-server/internal/api/handlers"
	"github.com/yourusername/wave-server/internal/api/middleware"
	"github.com/yourusername/wave-server/internal/config"
	"github.com/yourusername/wave-server/internal/repository"
	"github.com/yourusername/wave-server/internal/security"
	"github.com/yourusername/wave-server/internal/service"
	"github.com/yourusername/wave-server/pkg/health"
)

// TestEnv represents a test environment
type TestEnv struct {
	Echo        *echo.Echo
	Config      *config.Config
	Logger      *zap.Logger
	DB          *repository.Database
	UserService *service.UserService
	AuthService *service.AuthService
	Handler     *handlers.Handler
	Server      *httptest.Server
}

// SetupTest sets up a test environment
func SetupTest(t *testing.T) *TestEnv {
	// Setup logger
	logger := zaptest.NewLogger(t)

	// Load test config
	cfg := &config.Config{
		Server: struct {
			Port           int           `envconfig:"PORT" default:"8080"`
			Timeout        time.Duration `envconfig:"SERVER_TIMEOUT" default:"30s"`
			AllowedOrigins []string      `envconfig:"ALLOWED_ORIGINS" default:"*"`
		}{
			Port:           8080,
			Timeout:        30 * time.Second,
			AllowedOrigins: []string{"*"},
		},
		Database: struct {
			Host     string `envconfig:"DB_HOST" required:"true"`
			Port     int    `envconfig:"DB_PORT" default:"5433"`
			User     string `envconfig:"DB_USER" required:"true"`
			Password string `envconfig:"DB_PASSWORD" required:"true"`
			Name     string `envconfig:"DB_NAME" default:"wave"`
			PoolSize int    `envconfig:"DB_POOL_SIZE" default:"10"`
		}{
			Host:     "localhost",
			Port:     5433,
			User:     "yugabyte",
			Password: "yugabyte",
			Name:     "wave_test",
			PoolSize: 10,
		},
		Auth: struct {
			JWTSecret     string        `envconfig:"JWT_SECRET" required:"true"`
			TokenExpiry   time.Duration `envconfig:"TOKEN_EXPIRY" default:"24h"`
			RefreshExpiry time.Duration `envconfig:"REFRESH_EXPIRY" default:"720h"`
		}{
			JWTSecret:     "test_secret",
			TokenExpiry:   24 * time.Hour,
			RefreshExpiry: 720 * time.Hour,
		},
		Environment: "test",
		LogLevel:    "info",
	}

	// Check if we're running in CI
	if os.Getenv("CI") == "true" {
		cfg.Database.Host = os.Getenv("DB_HOST")
	}

	// Create database connection
	ctx := context.Background()
	db, err := repository.New(ctx, cfg, logger)
	require.NoError(t, err, "Failed to connect to database")

	// Run migrations
	err = db.RunMigrations(ctx)
	require.NoError(t, err, "Failed to run migrations")

	// Create repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)

	// Create services
	userService := service.NewUserService(userRepo, logger)
	authService := service.NewAuthService(userRepo, tokenRepo, cfg, logger)

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Create handlers
	handler := &handlers.Handler{
		Auth: &handlers.AuthHandler{
			AuthService: authService,
			UserService: userService,
			Config:      cfg,
			Logger:      logger,
		},
		logger: logger,
	}

	// Setup health checker
	healthChecker := health.New(db.Pool, logger)

	// Configure middleware
	middleware.SetupMiddleware(e, cfg, logger, authService)

	// Configure routes
	api.SetupRoutes(e, handler, cfg, authService, healthChecker, logger)

	// Create test server
	server := httptest.NewServer(e)

	return &TestEnv{
		Echo:        e,
		Config:      cfg,
		Logger:      logger,
		DB:          db,
		UserService: userService,
		AuthService: authService,
		Handler:     handler,
		Server:      server,
	}
}

// TearDown cleans up the test environment
func (env *TestEnv) TearDown() {
	if env.Server != nil {
		env.Server.Close()
	}
	if env.DB != nil {
		env.DB.Close()
	}
}

// CreateTestUser creates a test user
func (env *TestEnv) CreateTestUser(t *testing.T, username string) (string, string) {
	// Generate a test key pair
	publicKey := make([]byte, 800)            // Dummy public key
	encryptedPrivateKey := make([]byte, 1200) // Dummy encrypted private key
	salt := make([]byte, 16)                  // Dummy salt

	// Register user
	userID := security.HashUsername(username)
	token, err := env.AuthService.Login(context.Background(), username)
	require.NoError(t, err, "Failed to login")

	return userID, token
}

// GetAuthHeader returns an authentication header
func (env *TestEnv) GetAuthHeader(token string) http.Header {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)
	return header
}

// GetBase64 encodes a byte slice as URL-safe base64
func GetBase64(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}
