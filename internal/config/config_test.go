package config

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetSingleton resets the config singleton for testing.
func resetSingleton() {
	configOnce = sync.Once{}
	configInstance = nil
	configErr = nil
}

// setRequiredEnv sets the minimum required env vars for Load() to succeed.
func setRequiredEnv(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "postgres://test:test@localhost:5432/testdb")
	t.Setenv("YANDEX_DISK_TOKEN", "test-token-12345")
}

// --- Load tests ---

func TestLoad_Defaults(t *testing.T) {
	setRequiredEnv(t)
	resetSingleton()

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "postgres", cfg.DBType)
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb", cfg.PostgresDSN)
	assert.Equal(t, "test-token-12345", cfg.YandexDiskToken)
	assert.Equal(t, "https://cloud-api.yandex.net/v1/disk", cfg.YandexDiskBaseURL)
	assert.Equal(t, 30, cfg.YandexDiskTimeout)
	assert.Equal(t, 200, cfg.YandexDiskRateLimitDelayMs)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, 100, cfg.ScanPageSize)
	assert.Equal(t, 200, cfg.RateLimitDelayMs)
}

func TestLoad_CustomValues(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("DB_TYPE", "postgres")
	t.Setenv("YANDEX_DISK_BASE_URL", "https://custom-api.example.com/v1/disk")
	t.Setenv("YANDEX_DISK_TIMEOUT", "60")
	t.Setenv("YANDEX_DISK_RATE_LIMIT_DELAY_MS", "500")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("SCAN_PAGE_SIZE", "50")
	t.Setenv("RATE_LIMIT_DELAY_MS", "1000")
	resetSingleton()

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "postgres", cfg.DBType)
	assert.Equal(t, "https://custom-api.example.com/v1/disk", cfg.YandexDiskBaseURL)
	assert.Equal(t, 60, cfg.YandexDiskTimeout)
	assert.Equal(t, 500, cfg.YandexDiskRateLimitDelayMs)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, 50, cfg.ScanPageSize)
	assert.Equal(t, 1000, cfg.RateLimitDelayMs)
}

func TestLoad_MissingPostgresDSN(t *testing.T) {
	resetSingleton()
	os.Unsetenv("POSTGRES_DSN")
	t.Setenv("YANDEX_DISK_TOKEN", "test-token")

	require.Panics(t, func() {
		Load()
	}, "Load should panic when POSTGRES_DSN is missing")
}

func TestLoad_MissingYandexDiskToken(t *testing.T) {
	resetSingleton()
	t.Setenv("POSTGRES_DSN", "postgres://test:test@localhost:5432/testdb")
	os.Unsetenv("YANDEX_DISK_TOKEN")

	require.Panics(t, func() {
		Load()
	}, "Load should panic when YANDEX_DISK_TOKEN is missing")
}

func TestLoad_InvalidScanPageSize(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("SCAN_PAGE_SIZE", "0")
	resetSingleton()

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "SCAN_PAGE_SIZE")
}

func TestLoad_InvalidRateLimitDelayMs(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("RATE_LIMIT_DELAY_MS", "-1")
	resetSingleton()

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "RATE_LIMIT_DELAY_MS")
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("LOG_LEVEL", "trace")
	resetSingleton()

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "LOG_LEVEL")
}

func TestLoad_InvalidDBType(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("DB_TYPE", "mysql")
	resetSingleton()

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "DB_TYPE")
}

// --- Yandex Disk specific validation tests ---

func TestLoad_InvalidYandexDiskBaseURL_Empty(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("YANDEX_DISK_BASE_URL", "")
	resetSingleton()

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "YANDEX_DISK_BASE_URL")
}

func TestLoad_InvalidYandexDiskTimeout_Zero(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("YANDEX_DISK_TIMEOUT", "0")
	resetSingleton()

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "YANDEX_DISK_TIMEOUT")
}

func TestLoad_InvalidYandexDiskTimeout_Negative(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("YANDEX_DISK_TIMEOUT", "-5")
	resetSingleton()

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "YANDEX_DISK_TIMEOUT")
}

func TestLoad_InvalidYandexDiskRateLimitDelayMs_Negative(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("YANDEX_DISK_RATE_LIMIT_DELAY_MS", "-1")
	resetSingleton()

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "YANDEX_DISK_RATE_LIMIT_DELAY_MS")
}

func TestLoad_YandexDiskRateLimitDelayMs_Zero(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("YANDEX_DISK_RATE_LIMIT_DELAY_MS", "0")
	resetSingleton()

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, 0, cfg.YandexDiskRateLimitDelayMs,
		"Zero rate limit delay should be valid (disables rate limiting)")
}

// --- Getter tests ---

func TestGetters(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("YANDEX_DISK_BASE_URL", "https://custom.api.com/v1/disk")
	t.Setenv("YANDEX_DISK_TIMEOUT", "45")
	t.Setenv("YANDEX_DISK_RATE_LIMIT_DELAY_MS", "300")
	t.Setenv("SCAN_PAGE_SIZE", "25")
	t.Setenv("RATE_LIMIT_DELAY_MS", "150")
	resetSingleton()

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "https://custom.api.com/v1/disk", cfg.GetYandexDiskBaseURL())
	assert.Equal(t, 45, cfg.GetYandexDiskTimeout())
	assert.Equal(t, 300, cfg.GetYandexDiskRateLimitDelayMs())
	assert.Equal(t, 25, cfg.GetScanPageSize())
	assert.Equal(t, 150, cfg.GetRateLimitDelayMs())
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb", cfg.GetPostgresDSN())
	assert.Equal(t, "test-token-12345", cfg.GetYandexDiskToken())
}

// --- Singleton behavior tests ---

func TestGet_PanicsBeforeLoad(t *testing.T) {
	resetSingleton()

	require.Panics(t, func() {
		Get()
	}, "Get() should panic if Load() was not called")
}

func TestLoad_Singleton(t *testing.T) {
	setRequiredEnv(t)
	resetSingleton()

	cfg1, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg1)

	cfg2, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg2)

	assert.Same(t, cfg1, cfg2, "Load() should return the same singleton instance")
}

func TestGet_AfterLoad(t *testing.T) {
	setRequiredEnv(t)
	resetSingleton()

	cfg, err := Load()
	require.NoError(t, err)

	got := Get()
	assert.Same(t, cfg, got, "Get() should return the same instance as Load()")
}

// --- IsDevelopment tests ---

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		want     bool
	}{
		{"debug level", "debug", true},
		{"info level", "info", false},
		{"warn level", "warn", false},
		{"error level", "error", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{LogLevel: tt.logLevel}
			assert.Equal(t, tt.want, cfg.IsDevelopment())
		})
	}
}
