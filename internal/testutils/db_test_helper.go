package testutils

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/models"
)

// TestDBConfig holds test database configuration
type TestDBConfig struct {
	DSN string
}

// GetTestDBConfig returns test database configuration from config file
func GetTestDBConfig(t *testing.T) *TestDBConfig {
	// First try environment variable (for backward compatibility)
	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		return &TestDBConfig{DSN: dsn}
	}

	// Try to load from config file
	config, err := LoadTestConfig()
	if err != nil {
		t.Skipf("Failed to load test config: %v", err)
	}

	return &TestDBConfig{DSN: config.GetPostgresDSN()}
}

// SetupTestDatabase creates a test database connection
// Assumes database schema is already created
func SetupTestDatabase(t *testing.T, ctx context.Context) (*pgxpool.Pool, func()) {
	config := GetTestDBConfig(t)

	// Create connection pool
	pool, err := pgxpool.New(ctx, config.DSN)
	require.NoError(t, err, "Failed to create connection pool")

	// Test connection
	err = pool.Ping(ctx)
	require.NoError(t, err, "Failed to ping database")

	// Verify required tables exist
	err = verifySchema(ctx, pool)
	require.NoError(t, err, "Required database schema not found")

	// Cleanup function
	cleanup := func() {
		// Clean up test data
		err := CleanTestData(ctx, pool)
		require.NoError(t, err, "Failed to clean test data")

		pool.Close()
	}

	return pool, cleanup
}

// verifySchema checks that required tables exist
func verifySchema(ctx context.Context, pool *pgxpool.Pool) error {
	requiredTables := []string{"files", "file_paths", "source_processing"}

	for _, table := range requiredTables {
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`

		var exists bool
		err := pool.QueryRow(ctx, query, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check if table %s exists: %w", table, err)
		}

		if !exists {
			return fmt.Errorf("required table %s does not exist in database", table)
		}
	}

	return nil
}

// CleanTestData cleans up all test data from the database
func CleanTestData(ctx context.Context, pool *pgxpool.Pool) error {
	// Delete all data in reverse order of dependencies
	queries := []string{
		"DELETE FROM source_processing",
		"DELETE FROM file_paths",
		"DELETE FROM files",
	}

	for _, query := range queries {
		_, err := pool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to clean data with query '%s': %w", query, err)
		}
	}

	return nil
}

// CreateTestFile creates a test file model
func CreateTestFile(hash, name string) *models.File {
	return &models.File{
		Hash:      hash,
		Name:      name,
		Size:      1024,
		MimeType:  "text/plain",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestFilePath creates a test file path model
func CreateTestFilePath(path, sourceID, hash string, isActive bool) *models.FilePath {
	return &models.FilePath{
		Path:      path,
		SourceID:  sourceID,
		Hash:      hash,
		IsActive:  isActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestSourceProcessing creates a test source processing model
func CreateTestSourceProcessing(sourceID string) *models.SourceProcessing {
	now := time.Now()
	scanStatus := "completed"
	totalFilesFound := int64(10)

	return &models.SourceProcessing{
		SourceID:        sourceID,
		ScanStartedAt:   &now,
		ScanCompletedAt: &now,
		TotalFilesFound: &totalFilesFound,
		ScanStatus:      &scanStatus,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
