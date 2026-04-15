package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/testutils"
)

func TestFilePathRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDatabase(t, ctx)
	defer cleanup()

	fileRepo := NewFileRepository(pool)
	repo := NewFilePathRepository(pool)

	// Setup: Insert a test file first
	testFile := testutils.CreateTestFile("testfilehash", "testfile.txt")
	err := fileRepo.InsertFile(ctx, testFile)
	require.NoError(t, err, "Failed to insert test file")

	t.Run("InsertFilePath and GetFilePathByPathAndSource", func(t *testing.T) {
		// Create test file path
		filePath := testutils.CreateTestFilePath("/path/to/file.txt", "source1", "testfilehash", true)

		// Insert file path
		err := repo.InsertFilePath(ctx, filePath)
		require.NoError(t, err, "Failed to insert file path")

		// Retrieve file path
		retrievedPath, err := repo.GetFilePathByPathAndSource(ctx, "/path/to/file.txt", "source1")
		require.NoError(t, err, "Failed to get file path")

		// Verify file path data
		assert.Equal(t, filePath.Path, retrievedPath.Path)
		assert.Equal(t, filePath.SourceID, retrievedPath.SourceID)
		assert.Equal(t, filePath.Hash, retrievedPath.Hash)
		assert.Equal(t, filePath.IsActive, retrievedPath.IsActive)
		assert.Nil(t, retrievedPath.DeletedAt)
		assert.WithinDuration(t, filePath.CreatedAt, retrievedPath.CreatedAt, time.Second)
		assert.WithinDuration(t, filePath.UpdatedAt, retrievedPath.UpdatedAt, time.Second)
	})

	t.Run("GetFilePathByPathAndSource_NotFound", func(t *testing.T) {
		// Try to get non-existent file path
		filePath, err := repo.GetFilePathByPathAndSource(ctx, "/nonexistent/path", "source1")
		assert.Error(t, err, "Expected error for non-existent file path")
		assert.Nil(t, filePath, "Expected nil for non-existent file path")
	})

	t.Run("InsertFilePath_Duplicate", func(t *testing.T) {
		// Create and insert first file path
		path1 := testutils.CreateTestFilePath("/duplicate/path", "source1", "testfilehash", true)
		err := repo.InsertFilePath(ctx, path1)
		require.NoError(t, err, "Failed to insert first file path")

		// Try to insert duplicate file path with same path and source
		path2 := testutils.CreateTestFilePath("/duplicate/path", "source1", "testfilehash", false)
		err = repo.InsertFilePath(ctx, path2)
		assert.Error(t, err, "Expected error for duplicate file path")
	})

	t.Run("InsertFilePath_SamePathDifferentSources", func(t *testing.T) {
		// Same path, different sources should be allowed
		path1 := testutils.CreateTestFilePath("/same/path", "source1", "testfilehash", true)
		err := repo.InsertFilePath(ctx, path1)
		require.NoError(t, err, "Failed to insert file path for source1")

		path2 := testutils.CreateTestFilePath("/same/path", "source2", "testfilehash", true)
		err = repo.InsertFilePath(ctx, path2)
		require.NoError(t, err, "Failed to insert file path for source2")

		// Both should be retrievable
		retrieved1, err := repo.GetFilePathByPathAndSource(ctx, "/same/path", "source1")
		require.NoError(t, err)
		assert.Equal(t, "source1", retrieved1.SourceID)

		retrieved2, err := repo.GetFilePathByPathAndSource(ctx, "/same/path", "source2")
		require.NoError(t, err)
		assert.Equal(t, "source2", retrieved2.SourceID)
	})

	t.Run("UpdateFilePathDeletedAtByPathAndSource", func(t *testing.T) {
		// Insert a file path
		filePath := testutils.CreateTestFilePath("/path/to/delete.txt", "source1", "testfilehash", true)
		err := repo.InsertFilePath(ctx, filePath)
		require.NoError(t, err, "Failed to insert file path")

		// Update deleted_at
		deletedAt := time.Now()
		err = repo.UpdateFilePathDeletedAtByPathAndSource(ctx, "/path/to/delete.txt", "source1", &deletedAt)
		require.NoError(t, err, "Failed to update deleted_at")

		// Retrieve and verify
		retrieved, err := repo.GetFilePathByPathAndSource(ctx, "/path/to/delete.txt", "source1")
		require.NoError(t, err)
		assert.NotNil(t, retrieved.DeletedAt)
		assert.WithinDuration(t, deletedAt, *retrieved.DeletedAt, time.Second)
		assert.False(t, retrieved.IsActive, "File path should be inactive after deletion")
	})

	t.Run("GetFilePathsBySourceAndActive", func(t *testing.T) {
		// Clean up first
		testutils.CleanTestData(ctx, pool)

		// Insert test file again
		err := fileRepo.InsertFile(ctx, testFile)
		require.NoError(t, err)

		// Insert multiple file paths for same source
		paths := []struct {
			path     string
			isActive bool
		}{
			{"/active/path1", true},
			{"/active/path2", true},
			{"/inactive/path1", false},
			{"/inactive/path2", false},
		}

		for _, p := range paths {
			fp := testutils.CreateTestFilePath(p.path, "test-source", "testfilehash", p.isActive)
			err := repo.InsertFilePath(ctx, fp)
			require.NoError(t, err, "Failed to insert file path %s", p.path)
		}

		// Get active file paths
		activePaths, err := repo.GetFilePathsBySourceAndActive(ctx, "test-source", true)
		require.NoError(t, err)
		assert.Len(t, activePaths, 2, "Expected 2 active file paths")

		// Get inactive file paths
		inactivePaths, err := repo.GetFilePathsBySourceAndActive(ctx, "test-source", false)
		require.NoError(t, err)
		assert.Len(t, inactivePaths, 2, "Expected 2 inactive file paths")

		// Verify active paths
		for _, path := range activePaths {
			assert.True(t, path.IsActive, "Path should be active")
			assert.Contains(t, []string{"/active/path1", "/active/path2"}, path.Path)
		}
	})

	t.Run("GetFilePathsBySourceAndInactiveNotDeleted", func(t *testing.T) {
		// Clean up first
		testutils.CleanTestData(ctx, pool)

		// Insert test file again
		err := fileRepo.InsertFile(ctx, testFile)
		require.NoError(t, err)

		// Insert test file paths
		now := time.Now()

		// Path 1: Inactive, not deleted
		fp1 := testutils.CreateTestFilePath("/inactive/notdeleted", "test-source", "testfilehash", false)
		err = repo.InsertFilePath(ctx, fp1)
		require.NoError(t, err)

		// Path 2: Inactive, deleted
		fp2 := testutils.CreateTestFilePath("/inactive/deleted", "test-source", "testfilehash", false)
		fp2.DeletedAt = &now
		err = repo.InsertFilePath(ctx, fp2)
		require.NoError(t, err)

		// Path 3: Active, not deleted
		fp3 := testutils.CreateTestFilePath("/active/notdeleted", "test-source", "testfilehash", true)
		err = repo.InsertFilePath(ctx, fp3)
		require.NoError(t, err)

		// Get inactive not deleted paths
		paths, err := repo.GetFilePathsBySourceAndInactiveNotDeleted(ctx, "test-source")
		require.NoError(t, err)

		// Should only get path 1
		assert.Len(t, paths, 1, "Expected 1 inactive not deleted file path")
		assert.Equal(t, "/inactive/notdeleted", paths[0].Path)
		assert.False(t, paths[0].IsActive)
		assert.Nil(t, paths[0].DeletedAt)
	})
}
