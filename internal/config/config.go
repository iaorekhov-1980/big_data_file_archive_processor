package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	DBType         string
	PostgresDSN    string
	YandexDiskToken string
	LogLevel       string
	ScanPageSize   int
	RateLimitDelayMs int
}

var (
	configInstance *Config
	configOnce     sync.Once
	configErr      error
)

// Load loads configuration from environment variables
func Load() (*Config, error) {
	configOnce.Do(func() {
		// Try to load .env file, but don't fail if it doesn't exist
		_ = godotenv.Load()

		configInstance = &Config{
			DBType:         getEnvWithDefault("DB_TYPE", "postgres"),
			PostgresDSN:    getEnvRequired("POSTGRES_DSN"),
			YandexDiskToken: getEnvRequired("YANDEX_DISK_TOKEN"),
			LogLevel:       getEnvWithDefault("LOG_LEVEL", "info"),
			ScanPageSize:   getEnvIntWithDefault("SCAN_PAGE_SIZE", 100),
			RateLimitDelayMs: getEnvIntWithDefault("RATE_LIMIT_DELAY_MS", 200),
		}

		// Validate configuration
		if err := configInstance.validate(); err != nil {
			configErr = err
			configInstance = nil
		}
	})

	return configInstance, configErr
}

// Get returns the singleton configuration instance
func Get() *Config {
	if configInstance == nil {
		panic("configuration not loaded. Call Load() first")
	}
	return configInstance
}

// validate validates the configuration
func (c *Config) validate() error {
	if c.DBType != "postgres" {
		return fmt.Errorf("unsupported DB_TYPE: %s. Only 'postgres' is supported", c.DBType)
	}

	if c.PostgresDSN == "" {
		return fmt.Errorf("POSTGRES_DSN is required")
	}

	if c.YandexDiskToken == "" {
		return fmt.Errorf("YANDEX_DISK_TOKEN is required")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid LOG_LEVEL: %s. Must be one of: debug, info, warn, error", c.LogLevel)
	}

	if c.ScanPageSize <= 0 {
		return fmt.Errorf("SCAN_PAGE_SIZE must be positive, got: %d", c.ScanPageSize)
	}

	if c.RateLimitDelayMs < 0 {
		return fmt.Errorf("RATE_LIMIT_DELAY_MS must be non-negative, got: %d", c.RateLimitDelayMs)
	}

	return nil
}

// Helper functions for environment variable handling
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return value
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvIntWithDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("environment variable %s must be an integer, got: %s", key, value))
	}

	return intValue
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.LogLevel == "debug"
}

// GetPostgresDSN returns the PostgreSQL DSN
func (c *Config) GetPostgresDSN() string {
	return c.PostgresDSN
}

// GetYandexDiskToken returns the Yandex.Disk OAuth token
func (c *Config) GetYandexDiskToken() string {
	return c.YandexDiskToken
}

// GetScanPageSize returns the page size for scanning operations
func (c *Config) GetScanPageSize() int {
	return c.ScanPageSize
}

// GetRateLimitDelayMs returns the rate limit delay in milliseconds
func (c *Config) GetRateLimitDelayMs() int {
	return c.RateLimitDelayMs
}