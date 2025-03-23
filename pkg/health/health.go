package health

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Checker performs health checks
type Checker struct {
	dbPool     *pgxpool.Pool
	logger     *zap.Logger
	lastStatus Status
	mu         sync.RWMutex
}

// Status represents the health status
type Status struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
	Details struct {
		Database bool `json:"database"`
	} `json:"details"`
}

// New creates a new health checker
func New(dbPool *pgxpool.Pool, logger *zap.Logger) *Checker {
	return &Checker{
		dbPool: dbPool,
		logger: logger,
	}
}

// RegisterHandlers registers health check handlers with Echo
func (c *Checker) RegisterHandlers(e *echo.Echo) {
	e.GET("/health", c.HandleHealth)
	e.GET("/health/liveness", c.HandleLiveness)
	e.GET("/health/readiness", c.HandleReadiness)
}

// HandleHealth is a general health handler
func (c *Checker) HandleHealth(ctx echo.Context) error {
	return ctx.JSON(200, map[string]bool{"healthy": true})
}

// HandleLiveness checks if the service is running
func (c *Checker) HandleLiveness(ctx echo.Context) error {
	// Liveness just checks if the service is running
	return ctx.JSON(200, map[string]bool{"alive": true})
}

// HandleReadiness checks if the service is ready to serve requests
func (c *Checker) HandleReadiness(ctx echo.Context) error {
	status := c.Check(ctx.Request().Context())

	if !status.Healthy {
		return ctx.JSON(503, status)
	}
	return ctx.JSON(200, status)
}

// Check performs a health check
func (c *Checker) Check(ctx context.Context) Status {
	c.mu.RLock()
	cached := c.lastStatus
	c.mu.RUnlock()

	// Don't perform health checks too frequently
	if time.Now().Unix()%10 != 0 {
		return cached
	}

	status := Status{}
	status.Details.Database = c.checkDatabase(ctx)
	status.Healthy = status.Details.Database

	if !status.Healthy {
		status.Message = "One or more services are unhealthy"
	}

	c.mu.Lock()
	c.lastStatus = status
	c.mu.Unlock()

	return status
}

// checkDatabase checks database connectivity
func (c *Checker) checkDatabase(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := c.dbPool.Ping(ctx); err != nil {
		c.logger.Error("Database health check failed", zap.Error(err))
		return false
	}
	return true
}
