package disk

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/testutils"
)

// getTestToken retrieves the Yandex Disk token from environment or test config.
// Returns empty string if no token is configured (tests will be skipped).
func getTestToken(t *testing.T) string {
	// First try environment variable
	if token := os.Getenv("YANDEX_DISK_TOKEN"); token != "" {
		return token
	}

	// Try test config file
	config, err := testutils.LoadTestConfig()
	if err != nil {
		t.Logf("No test config found: %v", err)
		return ""
	}

	return config.GetYandexDiskToken()
}

// getTestBaseURL retrieves the Yandex Disk API base URL from test config.
func getTestBaseURL(t *testing.T) string {
	config, err := testutils.LoadTestConfig()
	if err != nil {
		return DefaultBaseURL
	}

	return config.GetYandexDiskBaseURL()
}

// TestYandexDiskClient_Creation verifies that the client is created correctly
// with configuration from test config.
func TestYandexDiskClient_Creation(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured. Set it in test_config.toml or YANDEX_DISK_TOKEN environment variable")
	}

	baseURL := getTestBaseURL(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(10*time.Second),
		WithRateLimitDelay(500*time.Millisecond),
	)

	require.NotNil(t, client)
	assert.Equal(t, token, client.token)
	assert.Equal(t, baseURL, client.baseURL)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
	assert.Equal(t, 500*time.Millisecond, client.rateLimitDelay)
}

// TestYandexDiskClient_doRequest_Unauthorized verifies that an invalid token
// returns a proper 401 error via the doRequest method.
func TestYandexDiskClient_doRequest_Unauthorized(t *testing.T) {
	baseURL := getTestBaseURL(t)

	client := NewYandexDiskClient("invalid-token-for-testing",
		WithBaseURL(baseURL),
		WithTimeout(10*time.Second),
	)

	// Use doRequest directly to test the transport/error handling layer
	// without requiring the interface methods (GetFileInfo, etc.)
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

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(10*time.Second),
	)

	// Query a path that definitely doesn't exist
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

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(15*time.Second),
	)

	// Request root folder info - this should succeed with a valid token
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

	// Use a short delay to keep the test fast
	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(15*time.Second),
		WithRateLimitDelay(100*time.Millisecond),
	)

	// Make multiple requests and measure total time
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

	// With 100ms delay between 3 requests, minimum time should be ~200ms
	// (first request has no delay, then 2 delays of 100ms each)
	expectedMinDelay := time.Duration(requestCount-1) * 100 * time.Millisecond
	assert.GreaterOrEqual(t, elapsed, expectedMinDelay,
		"Rate limiting should enforce delays between requests")
	t.Logf("Made %d requests in %v (expected min: %v)",
		requestCount, elapsed, expectedMinDelay)
}
