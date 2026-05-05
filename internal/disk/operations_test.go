package disk

import (
	"context"
	"testing"
	"time"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestYandexDiskClient_GetFileInfo_NotFound verifies that querying a non-existent
// file returns a proper 404 error via the interface method.
func TestYandexDiskClient_GetFileInfo_NotFound(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured")
	}

	baseURL := getTestBaseURL(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(10*time.Second),
	)

	ctx := context.Background()

	_, err := client.GetFileInfo(ctx, "/nonexistent_path_12345_test")
	require.Error(t, err)

	var diskErr *DiskError
	if assert.ErrorAs(t, err, &diskErr) {
		assert.True(t, diskErr.IsNotFound(),
			"Expected 404 Not Found error, got HTTP %d: %s",
			diskErr.StatusCode, diskErr.Message)
	}
}

// TestYandexDiskClient_GetFileInfo_Success verifies that querying the root folder
// returns valid resource info.
func TestYandexDiskClient_GetFileInfo_Success(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured")
	}

	baseURL := getTestBaseURL(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(15*time.Second),
	)

	ctx := context.Background()

	resource, err := client.GetFileInfo(ctx, "/")
	require.NoError(t, err, "Failed to get root folder info")

	require.NotNil(t, resource)
	assert.Equal(t, "dir", resource.Type, "Root should be a directory")
	assert.NotEmpty(t, resource.Path, "Root path should not be empty")
	assert.NotEmpty(t, resource.Name, "Root name should not be empty")
	t.Logf("Root folder: name=%s, type=%s, path=%s", resource.Name, resource.Type, resource.Path)
}

// TestYandexDiskClient_GetFolderContents_Root verifies that listing root folder
// contents returns items with valid structure.
func TestYandexDiskClient_GetFolderContents_Root(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured")
	}

	baseURL := getTestBaseURL(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(15*time.Second),
	)

	ctx := context.Background()

	contents, err := client.GetFolderContents(ctx, "/")
	require.NoError(t, err, "Failed to list root folder contents")

	assert.NotNil(t, contents)
	t.Logf("Root folder contains %d items", len(contents))

	for i, item := range contents {
		assert.NotEmpty(t, item.Path, "Item %d has empty path", i)
		assert.NotEmpty(t, item.Name, "Item %d has empty name", i)
		assert.Contains(t, []string{"file", "dir"}, item.Type,
			"Item %d has unexpected type: %s", i, item.Type)
		t.Logf("  [%s] %s (path: %s)", item.Type, item.Name, item.Path)
	}
}

// TestYandexDiskClient_ListFiles verifies that listing files with pagination works.
func TestYandexDiskClient_ListFiles(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured")
	}

	baseURL := getTestBaseURL(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	files, err := client.ListFiles(ctx, "", 0, 10)
	require.NoError(t, err, "Failed to list files")

	assert.NotNil(t, files)
	t.Logf("ListFiles returned %d files (limit=10)", len(files))

	for i, file := range files {
		assert.Equal(t, "file", file.Type, "Item %d is not a file", i)
		assert.NotEmpty(t, file.Path, "File %d has empty path", i)
		t.Logf("  File %d: %s (size: %d bytes)", i, file.Path, file.Size)
	}
}

// getTestFolder retrieves the test folder path from test config.
// Returns empty string if not configured (tests will be skipped).
func getTestFolder(t *testing.T) string {
	config, err := testutils.LoadTestConfig()
	if err != nil {
		t.Logf("No test config found: %v", err)
		return ""
	}
	return config.GetYandexDiskTestFolder()
}

// TestYandexDiskClient_ListFiles_WithSourceFolder verifies listing files
// filtered by a specific source folder.
func TestYandexDiskClient_ListFiles_WithSourceFolder(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured")
	}

	testFolder := getTestFolder(t)
	if testFolder == "" {
		t.Skip("YANDEX_DISK_TEST_FOLDER not set in test_config.toml, skipping source folder test")
	}

	baseURL := getTestBaseURL(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(15*time.Second),
	)

	ctx := context.Background()

	files, err := client.ListFiles(ctx, testFolder, 0, 10)
	require.NoError(t, err, "Failed to list files from source folder")

	assert.NotNil(t, files)
	t.Logf("ListFiles from '%s' returned %d files", testFolder, len(files))

	for i, file := range files {
		assert.Equal(t, "file", file.Type, "Item %d is not a file", i)
		t.Logf("  File %d: %s", i, file.Path)
	}
}

// TestYandexDiskClient_GetFolderContents_Subfolder verifies listing contents
// of a specific subfolder.
func TestYandexDiskClient_GetFolderContents_Subfolder(t *testing.T) {
	token := getTestToken(t)
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured")
	}

	testFolder := getTestFolder(t)
	if testFolder == "" {
		t.Skip("YANDEX_DISK_TEST_FOLDER not set in test_config.toml, skipping subfolder test")
	}

	baseURL := getTestBaseURL(t)

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(15*time.Second),
	)

	ctx := context.Background()

	contents, err := client.GetFolderContents(ctx, testFolder)
	require.NoError(t, err, "Failed to list folder contents")

	assert.NotNil(t, contents)
	t.Logf("Folder '%s' contains %d items", testFolder, len(contents))

	for i, item := range contents {
		assert.NotEmpty(t, item.Path, "Item %d has empty path", i)
		assert.NotEmpty(t, item.Name, "Item %d has empty name", i)
		t.Logf("  [%s] %s", item.Type, item.Name)
	}
}

// TestYandexDiskClient_InterfaceCompliance verifies that YandexDiskClient
// correctly implements the DiskClient interface.
func TestYandexDiskClient_InterfaceCompliance(t *testing.T) {
	var client DiskClient = &YandexDiskClient{}
	_ = client
	assert.True(t, true, "YandexDiskClient implements DiskClient interface")
}
