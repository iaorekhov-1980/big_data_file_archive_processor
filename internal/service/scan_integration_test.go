package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/disk"
	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/repository/postgres"
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

func TestScanService_Integration_RealYandexDisk(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := loadTestConfig(t)

	// Check required config
	token := config.GetYandexDiskToken()
	if token == "" {
		t.Skip("YANDEX_DISK_TOKEN not configured in test_config.toml")
	}

	testFolder := config.GetYandexDiskTestFolder()
	if testFolder == "" {
		t.Skip("test_folder not configured in test_config.toml [yandex_disk] section")
	}

	ctx := context.Background()

	// Setup database
	pool, _ := testutils.SetupTestDatabase(t, ctx)
	// cleanup disabled to inspect DB data after test

	// Clean DB before test
	err := testutils.CleanTestData(ctx, pool)
	require.NoError(t, err, "Failed to clean test data")

	// Create repository
	repo := postgres.NewPostgresRepositoryFromPool(pool)

	// Create Yandex Disk client
	baseURL := config.GetYandexDiskBaseURL()
	timeout := time.Duration(config.GetYandexDiskTimeout()) * time.Second
	rateLimitDelay := time.Duration(config.GetYandexDiskRateLimitDelayMs()) * time.Millisecond

	diskClient := disk.NewYandexDiskClient(token,
		disk.WithBaseURL(baseURL),
		disk.WithTimeout(timeout),
		disk.WithRateLimitDelay(rateLimitDelay),
	)

	// Create TxManager
	txManager := NewTxManager(pool)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create ScanService
	svc := NewScanService(diskClient, repo, txManager, 100, logger)

	// source_id is the folder name; full Yandex Disk path is /ods/<folder>
	sourceID := testFolder
	yandexDiskPath := "/ods/" + testFolder

	// Run scan
	result, err := svc.Scan(ctx, sourceID, yandexDiskPath, 0)
	require.NoError(t, err, "Scan should succeed")

	// Verify results
	t.Logf("Scan result: TotalFilesFound=%d, FilesInserted=%d, FilesSkipped=%d",
		result.TotalFilesFound, result.FilesInserted, result.FilesSkipped)
	t.Logf("Errors: %v", result.Errors)

	assert.Greater(t, result.TotalFilesFound, int64(0), "Should find at least one file")
	assert.Equal(t, result.TotalFilesFound, result.FilesInserted+result.FilesSkipped,
		"Total should equal inserted + skipped")

	// Verify data in DB
	sp, err := repo.GetSourceProcessing(ctx, sourceID)
	require.NoError(t, err, "Should find source processing record")
	require.NotNil(t, sp.ScanStatus)
	assert.Equal(t, "completed", *sp.ScanStatus)
	assert.Equal(t, result.TotalFilesFound, *sp.TotalFilesFound)

	// Verify file paths were inserted
	filePaths, err := repo.GetFilePathsBySourceAndActive(ctx, sourceID, true)
	require.NoError(t, err)
	assert.Len(t, filePaths, int(result.FilesInserted),
		"Active file paths should match inserted count")

	// Log what was found
	for _, fp := range filePaths {
		t.Logf("  FilePath: path=%s, hash=%s", fp.Path, fp.Hash)
	}

	// Verify idempotency: running again should return immediately
	result2, err := svc.Scan(ctx, sourceID, yandexDiskPath, 0)
	require.NoError(t, err)
	assert.Equal(t, result.TotalFilesFound, result2.TotalFilesFound,
		"Second scan should report same total")
	assert.Equal(t, int64(0), result2.FilesInserted,
		"Second scan should insert 0 files")

	// Verify no duplicate file_paths were created
	filePathsAfterSecondScan, err := repo.GetFilePathsBySourceAndActive(ctx, sourceID, true)
	require.NoError(t, err)
	assert.Len(t, filePathsAfterSecondScan, len(filePaths),
		"No new file paths should be added after second scan")
}
