package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/config"
)

// --- Helpers ---

// mockYandexDiskServer creates a test HTTP server that mimics Yandex Disk API responses.
func mockYandexDiskServer(statusCode int, responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fmt.Fprint(w, responseBody)
	}))
}

// newTestClient creates a YandexDiskClient pointing to a test server with no rate limiting.
func newTestClient(serverURL string) *YandexDiskClient {
	return NewYandexDiskClient("test-token",
		WithBaseURL(serverURL),
		WithTimeout(5*time.Second),
		WithRateLimitDelay(0),
	)
}

// --- Group 1: Error scenarios ---

func TestDiskError_IsRateLimited(t *testing.T) {
	server := mockYandexDiskServer(http.StatusTooManyRequests, `{
		"message": "Too many requests",
		"description": "Rate limit exceeded",
		"error": "too_many_requests"
	}`)
	defer server.Close()

	client := newTestClient(server.URL)
	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "/test")
	require.Error(t, err)

	var diskErr *DiskError
	if assert.ErrorAs(t, err, &diskErr) {
		assert.True(t, diskErr.IsRateLimited())
		assert.Equal(t, http.StatusTooManyRequests, diskErr.StatusCode)
		assert.Contains(t, diskErr.Message, "Too many requests")
	}
}

func TestDiskError_InternalServerError(t *testing.T) {
	server := mockYandexDiskServer(http.StatusInternalServerError, `{
		"message": "Internal server error",
		"description": "Something went wrong"
	}`)
	defer server.Close()

	client := newTestClient(server.URL)
	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "/test")
	require.Error(t, err)

	var diskErr *DiskError
	if assert.ErrorAs(t, err, &diskErr) {
		assert.Equal(t, http.StatusInternalServerError, diskErr.StatusCode)
		assert.False(t, diskErr.IsNotFound())
		assert.False(t, diskErr.IsAuthError())
		assert.False(t, diskErr.IsRateLimited())
	}
}

func TestDiskError_NonJSONErrorBody(t *testing.T) {
	server := mockYandexDiskServer(http.StatusBadGateway, "upstream connect error")
	defer server.Close()

	client := newTestClient(server.URL)
	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "/test")
	require.Error(t, err)

	var diskErr *DiskError
	if assert.ErrorAs(t, err, &diskErr) {
		assert.Equal(t, http.StatusBadGateway, diskErr.StatusCode)
		// Should contain the raw body as description
		assert.Contains(t, diskErr.Description, "upstream connect error")
	}
}

func TestDiskError_EmptyErrorBody(t *testing.T) {
	server := mockYandexDiskServer(http.StatusServiceUnavailable, "")
	defer server.Close()

	client := newTestClient(server.URL)
	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "/test")
	require.Error(t, err)

	var diskErr *DiskError
	if assert.ErrorAs(t, err, &diskErr) {
		assert.Equal(t, http.StatusServiceUnavailable, diskErr.StatusCode)
	}
}

func TestNetworkError(t *testing.T) {
	// Use a server that immediately closes connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Skip("server does not support hijacking")
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer server.Close()

	client := newTestClient(server.URL)
	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "/test")
	require.Error(t, err)
	// Should be a generic network error, not a DiskError
	assert.False(t, strings.Contains(err.Error(), "Yandex Disk API error"),
		"Network errors should not be wrapped as DiskError")
}

// --- Group 4: Factory ---

func TestNewDiskClient_ValidConfig(t *testing.T) {
	cfg := &config.Config{
		YandexDiskToken:            "test-token-123",
		YandexDiskBaseURL:          "https://cloud-api.yandex.net/v1/disk",
		YandexDiskTimeout:          30,
		YandexDiskRateLimitDelayMs: 200,
	}

	client, err := NewDiskClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Verify it's a YandexDiskClient with correct settings
	ydClient, ok := client.(*YandexDiskClient)
	require.True(t, ok, "NewDiskClient should return a YandexDiskClient")

	assert.Equal(t, "test-token-123", ydClient.token)
	assert.Equal(t, "https://cloud-api.yandex.net/v1/disk", ydClient.baseURL)
	assert.Equal(t, 30*time.Second, ydClient.httpClient.Timeout)
	assert.Equal(t, 200*time.Millisecond, ydClient.rateLimitDelay)
}

func TestNewDiskClient_NilConfig(t *testing.T) {
	client, err := NewDiskClient(nil)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "config is required")
}

func TestNewDiskClient_EmptyToken(t *testing.T) {
	cfg := &config.Config{
		YandexDiskToken:            "",
		YandexDiskBaseURL:          "https://cloud-api.yandex.net/v1/disk",
		YandexDiskTimeout:          30,
		YandexDiskRateLimitDelayMs: 200,
	}

	client, err := NewDiskClient(cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "token is required")
}

func TestNewDiskClient_InterfaceCompliance(t *testing.T) {
	cfg := &config.Config{
		YandexDiskToken:            "test-token",
		YandexDiskBaseURL:          "https://cloud-api.yandex.net/v1/disk",
		YandexDiskTimeout:          30,
		YandexDiskRateLimitDelayMs: 200,
	}

	client, err := NewDiskClient(cfg)
	require.NoError(t, err)

	// Verify it satisfies the DiskClient interface
	var diskClient DiskClient = client
	_ = diskClient
}

// --- Group 5: doRequest edge cases ---

func TestDoRequest_NilResponseBody(t *testing.T) {
	// Simulate a DELETE-like request where no response body is expected
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, server.URL+"/resource", nil)
	require.NoError(t, err)

	// Pass nil responseBody — should succeed without trying to decode
	err = client.doRequest(req, nil)
	assert.NoError(t, err)
}

func TestDoRequest_Non2xxNoJSONBody(t *testing.T) {
	server := mockYandexDiskServer(http.StatusForbidden, "Forbidden")
	defer server.Close()

	client := newTestClient(server.URL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/resource", nil)
	require.NoError(t, err)

	err = client.doRequest(req, nil)
	require.Error(t, err)

	var diskErr *DiskError
	if assert.ErrorAs(t, err, &diskErr) {
		assert.Equal(t, http.StatusForbidden, diskErr.StatusCode)
		// Non-JSON body should be stored as description
		assert.Contains(t, diskErr.Description, "Forbidden")
	}
}

func TestDoRequest_MalformedJSON(t *testing.T) {
	server := mockYandexDiskServer(http.StatusOK, `{"name": "test", "type": "dir", broken json`)
	defer server.Close()

	client := newTestClient(server.URL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/resource", nil)
	require.NoError(t, err)

	var result map[string]interface{}
	err = client.doRequest(req, &result)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestDoRequest_AuthHeaderSet(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"name": "root", "type": "dir"})
	}))
	defer server.Close()

	client := newTestClient(server.URL)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/resource", nil)
	require.NoError(t, err)

	var result map[string]string
	err = client.doRequest(req, &result)
	require.NoError(t, err)

	assert.Equal(t, "OAuth test-token", authHeader)
	assert.Equal(t, "root", result["name"])
}
