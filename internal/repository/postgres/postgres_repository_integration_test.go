package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	config := GetTestDBConfig(t)
	
	// Create full PostgresRepository
	repo, err := NewPostgresRepository(ctx, config.DSN)
	require.NoError(t, err, "Failed to create PostgresRepository")
	defer repo.Close()

	t.Run("FullRepositoryLifecycle", func(t *testing.T) {
		// Clean up first
		cleanTestData(ctx, repo.pool)
		
		// Step 1: Create a file
		file := createTestFile("integration-hash", "integration-file.txt")
		err := repo.InsertFile(ctx, file)
		require.NoError(t, err, "Failed to insert file")
		
		// Verify file was created
		retrievedFile, err := repo.GetFileByHash(ctx, "integration-hash")
		require.NoError(t, err)
		assert.Equal(t, "integration-file.txt", retrievedFile.Name)
		
		// Step 2: Create file path
		filePath := createTestFilePath("/integration/path.txt", "integration-source", "integration-hash", true)
		err = repo.InsertFilePath(ctx, filePath)
		require.NoError(t, err, "Failed to insert file path")
		
		// Verify file path was created
		retrievedPath, err := repo.GetFilePathByPathAndSource(ctx, "/integration/path.txt", "integration-source")
		require.NoError(t, err)
		assert.Equal(t, "integration-source", retrievedPath.SourceID)
		assert.Equal(t, "integration-hash", retrievedPath.Hash)
		assert.True(t, retrievedPath.IsActive)
		
		// Step 3: Create source processing
		sourceProcessing := createTestSourceProcessing("integration-source")
		err = repo.UpsertSourceProcessing(ctx, sourceProcessing)
		require.NoError(t, err, "Failed to upsert source processing")
		
		// Verify source processing was created
		retrievedSP, err := repo.GetSourceProcessing(ctx, "integration-source")
		require.NoError(t, err)
		assert.Equal(t, "integration-source", retrievedSP.SourceID)
		assert.Equal(t, "completed", *retrievedSP.ScanStatus)
		
		// Step 4: Get file paths by source and active status
		activePaths, err := repo.GetFilePathsBySourceAndActive(ctx, "integration-source", true)
		require.NoError(t, err)
		assert.Len(t, activePaths, 1)
		assert.Equal(t, "/integration/path.txt", activePaths[0].Path)
		
		// Step 5: Update file path as deleted
		deletedAt := time.Now()
		err = repo.UpdateFilePathDeletedAtByPathAndSource(ctx, "/integration/path.txt", "integration-source", &deletedAt)
		require.NoError(t, err, "Failed to update file path as deleted")
		
		// Verify file path is now inactive
		retrievedPath2, err := repo.GetFilePathByPathAndSource(ctx, "/integration/path.txt", "integration-source")
		require.NoError(t, err)
		assert.False(t, retrievedPath2.IsActive)
		assert.NotNil(t, retrievedPath2.DeletedAt)
		
		// Step 6: Get inactive not deleted file paths (should be empty since we marked it as deleted)
		inactiveNotDeleted, err := repo.GetFilePathsBySourceAndInactiveNotDeleted(ctx, "integration-source")
		require.NoError(t, err)
		assert.Len(t, inactiveNotDeleted, 0)
		
		// Step 7: Update source processing status
		updates := map[string]interface{}{
			"processing_status": "processing",
			"files_processed":   int64(1),
		}
		err = repo.UpdateSourceProcessingStatus(ctx, "integration-source", updates)
		require.NoError(t, err, "Failed to update source processing status")
		
		// Verify update
		retrievedSP2, err := repo.GetSourceProcessing(ctx, "integration-source")
		require.NoError(t, err)
		assert.Equal(t, "processing", *retrievedSP2.ProcessingStatus)
		assert.Equal(t, int64(1), *retrievedSP2.FilesProcessed)
	})

	t.Run("RepositoryCompositionWorks", func(t *testing.T) {
		// Clean up first
		cleanTestData(ctx, repo.pool)
		
		// Test that all embedded repositories work through the main repository
		
		// File repository methods
		file := createTestFile("compose-hash", "compose-file.txt")
		err := repo.InsertFile(ctx, file)
		require.NoError(t, err)
		
		retrievedFile, err := repo.GetFileByHash(ctx, "compose-hash")
		require.NoError(t, err)
		assert.Equal(t, "compose-file.txt", retrievedFile.Name)
		
		// FilePath repository methods
		filePath := createTestFilePath("/compose/path.txt", "compose-source", "compose-hash", true)
		err = repo.InsertFilePath(ctx, filePath)
		require.NoError(t, err)
		
		retrievedPath, err := repo.GetFilePathByPathAndSource(ctx, "/compose/path.txt", "compose-source")
		require.NoError(t, err)
		assert.Equal(t, "compose-source", retrievedPath.SourceID)
		
		// SourceProcessing repository methods
		sourceProcessing := createTestSourceProcessing("compose-source")
		err = repo.UpsertSourceProcessing(ctx, sourceProcessing)
		require.NoError(t, err)
		
		retrievedSP, err := repo.GetSourceProcessing(ctx, "compose-source")
		require.NoError(t, err)
		assert.Equal(t, "compose-source", retrievedSP.SourceID)
	})

	t.Run("TransactionIsolation", func(t *testing.T) {
		// Clean up first
		cleanTestData(ctx, repo.pool)
		
		// This test verifies that operations are isolated properly
		// Insert a file
		file1 := createTestFile("tx-hash1", "tx-file1.txt")
		err := repo.InsertFile(ctx, file1)
		require.NoError(t, err)
		
		// Insert another file
		file2 := createTestFile("tx-hash2", "tx-file2.txt")
		err = repo.InsertFile(ctx, file2)
		require.NoError(t, err)
		
		// Both files should be retrievable
		retrieved1, err := repo.GetFileByHash(ctx, "tx-hash1")
		require.NoError(t, err)
		assert.Equal(t, "tx-file1.txt", retrieved1.Name)
		
		retrieved2, err := repo.GetFileByHash(ctx, "tx-hash2")
		require.NoError(t, err)
		assert.Equal(t, "tx-file2.txt", retrieved2.Name)
		
		// Try to insert duplicate (should fail)
		file3 := createTestFile("tx-hash1", "tx-file3.txt")
		err = repo.InsertFile(ctx, file3)
		assert.Error(t, err, "Expected error for duplicate file")
		
		// Original file should still be retrievable
		retrieved1Again, err := repo.GetFileByHash(ctx, "tx-hash1")
		require.NoError(t, err)
		assert.Equal(t, "tx-file1.txt", retrieved1Again.Name, "Original file should not be affected by failed insert")
	})
}