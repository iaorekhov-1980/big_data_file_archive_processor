package disk

import (
	"context"
	"time"
)

// Resource represents a Yandex Disk API resource (file or folder).
// Fields map directly to the Yandex Disk REST API response format.
type Resource struct {
	// Path is the full path to the resource on Yandex Disk.
	Path string `json:"path"`

	// Name is the display name of the resource.
	Name string `json:"name"`

	// Type indicates whether the resource is a "file" or "dir" (directory).
	Type string `json:"type"`

	// Size is the file size in bytes. Omitted for directories.
	Size int64 `json:"size,omitempty"`

	// MD5 is the MD5 hash of the file. Omitted for directories.
	MD5 string `json:"md5,omitempty"`

	// SHA256 is the SHA-256 hash of the file. Omitted for directories.
	SHA256 string `json:"sha256,omitempty"`

	// Modified is the last modification timestamp.
	Modified time.Time `json:"modified,omitempty"`

	// Created is the creation timestamp.
	Created time.Time `json:"created,omitempty"`

	// MimeType is the MIME type of the file (e.g., "image/png").
	MimeType string `json:"mime_type,omitempty"`

	// MediaType is the media type classification (e.g., "image", "video").
	MediaType string `json:"media_type,omitempty"`
}

// DiskClient defines the interface for cloud disk operations.
// Currently implemented for Yandex Disk, but designed to be extensible
// for other cloud storage providers in the future.
type DiskClient interface {
	// ListFiles retrieves a paginated list of all files from the specified source folder.
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//   - sourceFolder: the folder path to scan (empty string for root)
	//   - offset: pagination offset (0-based)
	//   - limit: maximum number of items to return
	// Returns a slice of Resource items and any error encountered.
	ListFiles(ctx context.Context, sourceFolder string, offset, limit int) ([]Resource, error)

	// GetFolderContents retrieves the contents of a folder at the specified path.
	// Returns a list of Resource items representing files and subdirectories.
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//   - path: the folder path to list
	// Returns a slice of Resource items and any error encountered.
	GetFolderContents(ctx context.Context, path string) ([]Resource, error)

	// GetFileInfo retrieves metadata for a single file or folder.
	// This is a helper method for checking resource existence and details.
	// Parameters:
	//   - ctx: context for cancellation and timeouts
	//   - path: the full path to the resource
	// Returns a pointer to Resource and any error encountered.
	// Returns nil and an error if the resource does not exist.
	GetFileInfo(ctx context.Context, path string) (*Resource, error)
}
