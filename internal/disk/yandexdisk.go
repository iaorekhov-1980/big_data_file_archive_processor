package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Default values for Yandex Disk client configuration.
const (
	DefaultBaseURL        = "https://cloud-api.yandex.net/v1/disk"
	DefaultTimeout        = 30 * time.Second
	DefaultRateLimitDelay = 200 * time.Millisecond
	maxIdleConnections    = 10
	idleConnectionTimeout = 90 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	expectContinueTimeout = 1 * time.Second
)

// YandexDiskClient implements the DiskClient interface for Yandex Disk API.
type YandexDiskClient struct {
	baseURL         string
	token           string
	httpClient      *http.Client
	rateLimitDelay  time.Duration
	lastRequestTime time.Time
}

// YandexDiskClientOption defines a functional option for configuring YandexDiskClient.
type YandexDiskClientOption func(*YandexDiskClient)

// WithBaseURL sets a custom base URL for the Yandex Disk API.
func WithBaseURL(baseURL string) YandexDiskClientOption {
	return func(c *YandexDiskClient) {
		c.baseURL = baseURL
	}
}

// WithTimeout sets a custom HTTP client timeout.
func WithTimeout(timeout time.Duration) YandexDiskClientOption {
	return func(c *YandexDiskClient) {
		c.httpClient.Timeout = timeout
	}
}

// WithRateLimitDelay sets a custom delay between API requests.
func WithRateLimitDelay(delay time.Duration) YandexDiskClientOption {
	return func(c *YandexDiskClient) {
		c.rateLimitDelay = delay
	}
}

// NewYandexDiskClient creates a new Yandex Disk API client.
// The token parameter is a Yandex OAuth token required for API authentication.
// Options can be provided to customize the client behavior.
func NewYandexDiskClient(token string, opts ...YandexDiskClientOption) *YandexDiskClient {
	transport := &http.Transport{
		MaxIdleConns:          maxIdleConnections,
		IdleConnTimeout:       idleConnectionTimeout,
		DisableCompression:    false,
		TLSHandshakeTimeout:   tlsHandshakeTimeout,
		ExpectContinueTimeout: expectContinueTimeout,
	}

	client := &YandexDiskClient{
		baseURL:        DefaultBaseURL,
		token:          token,
		rateLimitDelay: DefaultRateLimitDelay,
		httpClient: &http.Client{
			Timeout:   DefaultTimeout,
			Transport: transport,
		},
	}

	// Apply functional options
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// doRequest executes an HTTP request, handles authentication, and decodes the response.
// If responseBody is non-nil, the response JSON is decoded into it.
// If responseBody is nil, the response body is discarded (used for DELETE operations).
func (c *YandexDiskClient) doRequest(req *http.Request, responseBody interface{}) error {
	// Set authentication header
	req.Header.Set("Authorization", "OAuth "+c.token)
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.handleErrorResponse(resp)
	}

	// If no response body expected (e.g., DELETE), return early
	if responseBody == nil {
		return nil
	}

	// Decode JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, responseBody); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// handleErrorResponse processes HTTP error responses and returns appropriate errors.
func (c *YandexDiskClient) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte("unable to read error response body")
	}

	// Try to parse Yandex Disk API error response
	var apiError struct {
		Message     string `json:"message"`
		Description string `json:"description"`
		ErrorCode   string `json:"error"`
	}
	if err := json.Unmarshal(body, &apiError); err == nil && apiError.Message != "" {
		return &DiskError{
			StatusCode:  resp.StatusCode,
			Message:     apiError.Message,
			Description: apiError.Description,
			ErrorCode:   apiError.ErrorCode,
		}
	}

	return &DiskError{
		StatusCode:  resp.StatusCode,
		Message:     fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		Description: string(body),
	}
}

// rateLimitWait enforces a minimum delay between API requests to respect rate limits.
// It blocks until the configured delay has elapsed since the last request.
func (c *YandexDiskClient) rateLimitWait(ctx context.Context) error {
	if c.rateLimitDelay <= 0 {
		return nil
	}

	elapsed := time.Since(c.lastRequestTime)
	if elapsed < c.rateLimitDelay {
		waitDuration := c.rateLimitDelay - elapsed
		select {
		case <-time.After(waitDuration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	c.lastRequestTime = time.Now()
	return nil
}

// buildGetRequest creates a GET request to the Yandex Disk API for the given path.
// This is a helper method used for testing the transport/error handling layer.
func (c *YandexDiskClient) buildGetRequest(path, queryParams string) (*http.Request, error) {
	endpoint := c.baseURL + "/resources?path=" + path
	if queryParams != "" {
		endpoint += "&" + queryParams
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request: %w", err)
	}

	return req, nil
}

// DiskError represents a Yandex Disk API error response.
type DiskError struct {
	StatusCode  int    `json:"status_code"`
	Message     string `json:"message"`
	Description string `json:"description"`
	ErrorCode   string `json:"error_code"`
}

func (e *DiskError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("Yandex Disk API error (HTTP %d): %s - %s", e.StatusCode, e.Message, e.Description)
	}
	return fmt.Sprintf("Yandex Disk API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// IsNotFound returns true if the error indicates a resource was not found (HTTP 404).
func (e *DiskError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsRateLimited returns true if the error indicates rate limiting (HTTP 429).
func (e *DiskError) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsAuthError returns true if the error indicates an authentication issue (HTTP 401).
func (e *DiskError) IsAuthError() bool {
	return e.StatusCode == http.StatusUnauthorized
}
