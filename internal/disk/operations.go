package disk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ListFiles retrieves a paginated list of all files from the specified source folder.
// Implements the DiskClient interface.
func (c *YandexDiskClient) ListFiles(ctx context.Context, sourceFolder string, offset, limit int) ([]Resource, error) {
	if err := c.rateLimitWait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	params := url.Values{}
	if sourceFolder != "" {
		params.Set("path", sourceFolder)
	}
	params.Set("offset", fmt.Sprintf("%d", offset))
	params.Set("limit", fmt.Sprintf("%d", limit))

	endpoint := c.baseURL + "/resources/files?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list files request: %w", err)
	}

	var response struct {
		Items  []Resource `json:"items"`
		Total  int        `json:"total"`
		Offset int        `json:"offset"`
		Limit  int        `json:"limit"`
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, fmt.Errorf("list files request failed: %w", err)
	}

	return response.Items, nil
}

// GetFolderContents retrieves the contents of a folder at the specified path.
// Implements the DiskClient interface.
func (c *YandexDiskClient) GetFolderContents(ctx context.Context, path string) ([]Resource, error) {
	if err := c.rateLimitWait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	params := url.Values{}
	params.Set("path", path)
	params.Set("limit", "100") // Maximum items per page

	endpoint := c.baseURL + "/resources?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get folder contents request: %w", err)
	}

	var response struct {
		Embedded struct {
			Items []Resource `json:"items"`
			Total int        `json:"total"`
		} `json:"_embedded"`
	}

	if err := c.doRequest(req, &response); err != nil {
		return nil, fmt.Errorf("get folder contents request failed for path '%s': %w", path, err)
	}

	return response.Embedded.Items, nil
}

// GetFileInfo retrieves metadata for a single file or folder.
// Implements the DiskClient interface.
func (c *YandexDiskClient) GetFileInfo(ctx context.Context, path string) (*Resource, error) {
	if err := c.rateLimitWait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	params := url.Values{}
	params.Set("path", path)

	endpoint := c.baseURL + "/resources?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get file info request: %w", err)
	}

	var resource Resource
	if err := c.doRequest(req, &resource); err != nil {
		return nil, fmt.Errorf("get file info request failed for path '%s': %w", path, err)
	}

	return &resource, nil
}
