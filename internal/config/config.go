package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds the application configuration
type Config struct {
	Server struct {
		Port           int           `envconfig:"PORT" default:"8080"`
		Timeout        time.Duration `envconfig:"SERVER_TIMEOUT" default:"30s"`
		AllowedOrigins []string      `envconfig:"ALLOWED_ORIGINS" default:"*"`
	}

	Database struct {
		Host     string `envconfig:"DB_HOST" required:"true"`
		Port     int    `envconfig:"DB_PORT" default:"5433"`
		User     string `envconfig:"DB_USER" required:"true"`
		Password string `envconfig:"DB_PASSWORD" required:"true"`
		Name     string `envconfig:"DB_NAME" default:"wave"`
		PoolSize int    `envconfig:"DB_POOL_SIZE" default:"10"`
	}

	Auth struct {
		JWTSecret     string        `envconfig:"JWT_SECRET" required:"true"`
		TokenExpiry   time.Duration `envconfig:"TOKEN_EXPIRY" default:"24h"`
		RefreshExpiry time.Duration `envconfig:"REFRESH_EXPIRY" default:"720h"`
	}

	Environment string `envconfig:"ENVIRONMENT" default:"production"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	return &cfg, nil
}

// IsDevelopment checks if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// GetDSN returns the database connection string
func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
	)
}
