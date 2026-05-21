package disk

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/testutils"
)

// loadTestConfig loads test configuration. Skips the test if config is not found.
func loadTestConfig(t *testing.T) *testutils.TestConfig {
	config, err := testutils.LoadTestConfig()
	if err != nil {
		t.Skipf("test_config.toml not found: %v", err)
	}
	return config
}

// getTestToken retrieves the Yandex Disk token from test config.
func getTestToken(t *testing.T) string {
	return loadTestConfig(t).GetYandexDiskToken()
}

// getTestBaseURL retrieves the Yandex Disk API base URL from test config.
func getTestBaseURL(t *testing.T) string {
	return loadTestConfig(t).GetYandexDiskBaseURL()
}

// getTestTimeout retrieves the HTTP client timeout from test config.
func getTestTimeout(t *testing.T) time.Duration {
	timeout := loadTestConfig(t).GetYandexDiskTimeout()
	return time.Duration(timeout) * time.Second
}

// getTestRateLimitDelay retrieves the rate limit delay from test config.
func getTestRateLimitDelay(t *testing.T) time.Duration {
	delayMs := loadTestConfig(t).GetYandexDiskRateLimitDelayMs()
	return time.Duration(delayMs) * time.Millisecond
}

// TestYandexDiskClient_Creation verifies that the client is created correctly
// with configuration from test config.
func TestYandexDiskClient_Creation(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured in test_config.toml")
	}

	baseURL := getTestBaseURL(t)
	timeout := getTestTimeout(t)
	rateLimitDelay := getTestRateLimitDelay(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(timeout),
		WithRateLimitDelay(rateLimitDelay),
	)

	require.NotNil(t, client)
	assert.Equal(t, token, client.token)
	assert.Equal(t, baseURL, client.baseURL)
	assert.Equal(t, timeout, client.httpClient.Timeout)
	assert.Equal(t, rateLimitDelay, client.rateLimitDelay)
}

// TestYandexDiskClient_doRequest_Unauthorized verifies that an invalid token
// returns a proper 401 error via the doRequest method.
func TestYandexDiskClient_doRequest_Unauthorized(t *testing.T) {
	baseURL := getTestBaseURL(t)
	timeout := getTestTimeout(t)

	client := NewYandexDiskClient("invalid-token-for-testing",
		WithBaseURL(baseURL),
		WithTimeout(timeout),
	)

	req, err := client.buildGetRequest("/", "")
	if err != nil {
		t.Skipf("Failed to build request: %v", err)
	}

	err = client.doRequest(req, nil)
	require.Error(t, err)

	var diskErr *DiskError
	if assert.ErrorAs(t, err, &diskErr) {
		assert.True(t, diskErr.IsAuthError(),
			"Expected 401 Unauthorized error, got HTTP %d: %s",
			diskErr.StatusCode, diskErr.Message)
	}
}

// TestYandexDiskClient_doRequest_NotFound verifies that querying a non-existent
// path returns a proper 404 error via the doRequest method.
func TestYandexDiskClient_doRequest_NotFound(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured")
	}

	baseURL := getTestBaseURL(t)
	timeout := getTestTimeout(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(timeout),
	)

	req, err := client.buildGetRequest("/nonexistent_path_12345_test", "")
	if err != nil {
		t.Skipf("Failed to build request: %v", err)
	}

	err = client.doRequest(req, nil)
	require.Error(t, err)

	var diskErr *DiskError
	if assert.ErrorAs(t, err, &diskErr) {
		assert.True(t, diskErr.IsNotFound(),
			"Expected 404 Not Found error, got HTTP %d: %s",
			diskErr.StatusCode, diskErr.Message)
	}
}

// TestYandexDiskClient_doRequest_Success verifies that a valid request to the
// root folder succeeds via the doRequest method.
func TestYandexDiskClient_doRequest_Success(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured")
	}

	baseURL := getTestBaseURL(t)
	timeout := getTestTimeout(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(timeout),
	)

	req, err := client.buildGetRequest("/", "")
	if err != nil {
		t.Skipf("Failed to build request: %v", err)
	}

	var response struct {
		Type string `json:"type"`
		Path string `json:"path"`
		Name string `json:"name"`
	}

	err = client.doRequest(req, &response)
	require.NoError(t, err, "Failed to get root folder info")

	assert.Equal(t, "dir", response.Type, "Root should be a directory")
	assert.NotEmpty(t, response.Path, "Root path should not be empty")
	t.Logf("Root folder: name=%s, type=%s, path=%s", response.Name, response.Type, response.Path)
}

// TestYandexDiskClient_RateLimiting verifies that rate limiting works correctly
// by measuring delays between requests via doRequest.
func TestYandexDiskClient_RateLimiting(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured")
	}

	baseURL := getTestBaseURL(t)
	timeout := getTestTimeout(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(timeout),
		WithRateLimitDelay(100*time.Millisecond),
	)

	start := time.Now()
	requestCount := 3

	for i := 0; i < requestCount; i++ {
		req, err := client.buildGetRequest("/", "")
		if err != nil {
			t.Skipf("Failed to build request: %v", err)
		}

		err = client.doRequest(req, nil)
		require.NoError(t, err, "Request %d failed", i)
	}

	elapsed := time.Since(start)

	expectedMinDelay := time.Duration(requestCount-1) * 100 * time.Millisecond
	assert.GreaterOrEqual(t, elapsed, expectedMinDelay,
		"Rate limiting should enforce delays between requests")
	t.Logf("Made %d requests in %v (expected min: %v)",
		requestCount, elapsed, expectedMinDelay)
}
