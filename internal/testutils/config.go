package testutils

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// TestConfig holds test database configuration
type TestConfig struct {
	Postgres struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		Database string `toml:"database"`
		Username string `toml:"username"`
		Password string `toml:"password"`
	} `toml:"postgres"`

	// YandexDisk holds optional Yandex Disk API test configuration
	YandexDisk struct {
		Token            string `toml:"token"`
		BaseURL          string `toml:"base_url"`
		TestFolder       string `toml:"test_folder"`
		Timeout          int    `toml:"timeout"`
		RateLimitDelayMs int    `toml:"rate_limit_delay_ms"`
	} `toml:"yandex_disk"`
}

// LoadTestConfig loads test configuration from file
func LoadTestConfig() (*TestConfig, error) {
	// Look for config file in current directory and parent directories
	configPaths := []string{
		"test_config.toml",
		"../../test_config.toml",
		"../../../test_config.toml",
	}

	var configFile string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	if configFile == "" {
		return nil, fmt.Errorf("test_config.toml not found. Please create it from test_config.example.toml")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	var config TestConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	return &config, nil
}

// GetPostgresDSN returns PostgreSQL DSN from config
func (c *TestConfig) GetPostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.Postgres.Username,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.Database)
}

// GetYandexDiskToken returns the Yandex Disk OAuth token from config
func (c *TestConfig) GetYandexDiskToken() string {
	return c.YandexDisk.Token
}

// GetYandexDiskTestFolder returns the test folder path for Yandex Disk tests
func (c *TestConfig) GetYandexDiskTestFolder() string {
	return c.YandexDisk.TestFolder
}

// GetYandexDiskBaseURL returns the Yandex Disk API base URL from config
func (c *TestConfig) GetYandexDiskBaseURL() string {
	if c.YandexDisk.BaseURL == "" {
		return "https://cloud-api.yandex.net/v1/disk"
	}
	return c.YandexDisk.BaseURL
}

// GetYandexDiskTimeout returns the HTTP client timeout in seconds from config
func (c *TestConfig) GetYandexDiskTimeout() int {
	if c.YandexDisk.Timeout <= 0 {
		return 30
	}
	return c.YandexDisk.Timeout
}

// GetYandexDiskRateLimitDelayMs returns the rate limit delay in milliseconds from config
func (c *TestConfig) GetYandexDiskRateLimitDelayMs() int {
	if c.YandexDisk.RateLimitDelayMs < 0 {
		return 200
	}
	return c.YandexDisk.RateLimitDelayMs
}

// CreateTestConfig creates a test config file from example
func CreateTestConfig() error {
	exampleFile := "test_config.example.toml"
	configFile := "test_config.toml"

	// Check if example exists
	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		return fmt.Errorf("example config file %s not found", exampleFile)
	}

	// Check if config already exists
	if _, err := os.Stat(configFile); err == nil {
		return fmt.Errorf("config file %s already exists", configFile)
	}

	// Read example
	data, err := os.ReadFile(exampleFile)
	if err != nil {
		return fmt.Errorf("failed to read example config: %w", err)
	}

	// Write config
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Created %s from %s\n", configFile, exampleFile)
	fmt.Println("Please edit test_config.toml with your database credentials")
	return nil
}
