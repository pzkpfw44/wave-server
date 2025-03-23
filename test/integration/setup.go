package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"
	"unsafe"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	Server         *httptest.Server
	Echo           *echo.Echo
	Config         *config.Config
	Logger         *zap.Logger
	UserRepo       *repository.UserRepository
	TokenRepo      *repository.TokenRepository
	Security       *security.SecurityTestHelper
	AuthService    *service.AuthService
	UserService    *service.UserService
	MessageService *service.MessageService
	ContactService *service.ContactService
	AccountService *service.AccountService
}

var testEnv *TestEnv

// mockDBPool is a simplified mock for pgxpool.Pool
type mockDBPool struct {
	t *testing.T
}

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
		Server:         server,
		Echo:           e,
		Config:         cfg,
		Logger:         logger,
		UserRepo:       userRepo,
		TokenRepo:      tokenRepo,
		Security:       &security.SecurityTestHelper{},
		AuthService:    authService,
		UserService:    userService,
		MessageService: messageService,
		ContactService: contactService,
		AccountService: accountService,
	}

	// Return cleanup function
	cleanup := func() {
		server.Close()
	}

	return server, cleanup
}

// setupTestDatabase creates an in-memory test database
func setupTestDatabase(t *testing.T) *repository.Database {
	// Create logger
	logger := zaptest.NewLogger(t)

	// Create config
	cfg := &config.Config{}

	// Create mock pool
	mockPool := &mockDBPool{t: t}

	// Create database with mock pool
	db := &repository.Database{
		Pool:   forceCastToPoolPtr(mockPool),
		Logger: logger,
		Config: cfg,
	}

	return db
}

// forceCastToPoolPtr performs an unsafe cast for testing purposes
func forceCastToPoolPtr(mock *mockDBPool) *pgxpool.Pool {
	return (*pgxpool.Pool)(unsafe.Pointer(mock))
}

// Close implements the Pool.Close method
func (m *mockDBPool) Close() {}

// Ping implements the Pool.Ping method
func (m *mockDBPool) Ping(ctx context.Context) error {
	return nil
}

// Exec implements the Pool.Exec method
func (m *mockDBPool) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

// Query implements the Pool.Query method
func (m *mockDBPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return &MockRows{}, nil
}

// QueryRow implements the Pool.QueryRow method
func (m *mockDBPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return &MockRow{}
}

// Begin implements the Pool.Begin method
func (m *mockDBPool) Begin(ctx context.Context) (pgx.Tx, error) {
	return nil, nil
}

// BeginTx implements the Pool.BeginTx method
func (m *mockDBPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	return nil, nil
}

// Acquire implements the Pool.Acquire method
func (m *mockDBPool) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	return nil, nil
}

// AcquireFunc implements the Pool.AcquireFunc method
func (m *mockDBPool) AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error {
	return nil
}

// AcquireAllIdle implements the Pool.AcquireAllIdle method
func (m *mockDBPool) AcquireAllIdle(ctx context.Context) []*pgxpool.Conn {
	return nil
}

// Config implements the Pool.Config method
func (m *mockDBPool) Config() *pgxpool.Config {
	return nil
}

// Stat implements the Pool.Stat method
func (m *mockDBPool) Stat() *pgxpool.Stat {
	return nil
}

// Reset implements the Pool.Reset method
func (m *mockDBPool) Reset() {}

// MockRow implements pgx.Row for testing
type MockRow struct{}

// Scan implements the pgx.Row.Scan method
func (m *MockRow) Scan(dest ...interface{}) error {
	return pgx.ErrNoRows
}

// MockRows implements pgx.Rows for testing
type MockRows struct {
	closed bool
}

// Close implements the pgx.Rows.Close method
func (m *MockRows) Close() {}

// Err implements the pgx.Rows.Err method
func (m *MockRows) Err() error {
	return nil
}

// CommandTag implements the pgx.Rows.CommandTag method
func (m *MockRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

// FieldDescriptions implements the pgx.Rows.FieldDescriptions method
func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

// Next implements the pgx.Rows.Next method
func (m *MockRows) Next() bool {
	return false
}

// Scan implements the pgx.Rows.Scan method
func (m *MockRows) Scan(dest ...interface{}) error {
	return nil
}

// Values implements the pgx.Rows.Values method
func (m *MockRows) Values() ([]interface{}, error) {
	return nil, nil
}

// RawValues implements the pgx.Rows.RawValues method
func (m *MockRows) RawValues() [][]byte {
	return nil
}

// Conn implements the pgx.Rows.Conn method
func (m *MockRows) Conn() *pgx.Conn {
	return nil
}

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
