package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/models"
	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/testutils"
)

func TestSourceProcessingRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDatabase(t, ctx)
	defer cleanup()

	repo := NewSourceProcessingRepository(pool)

	t.Run("UpsertSourceProcessing and GetSourceProcessing", func(t *testing.T) {
		// Create test source processing
		sourceProcessing := testutils.CreateTestSourceProcessing("test-source-1")

		// Upsert source processing
		err := repo.UpsertSourceProcessing(ctx, sourceProcessing)
		require.NoError(t, err, "Failed to upsert source processing")

		// Retrieve source processing
		retrieved, err := repo.GetSourceProcessing(ctx, "test-source-1")
		require.NoError(t, err, "Failed to get source processing")

		// Verify data
		assert.Equal(t, sourceProcessing.SourceID, retrieved.SourceID)
		assert.WithinDuration(t, *sourceProcessing.ScanStartedAt, *retrieved.ScanStartedAt, time.Second)
		assert.WithinDuration(t, *sourceProcessing.ScanCompletedAt, *retrieved.ScanCompletedAt, time.Second)
		assert.Equal(t, *sourceProcessing.TotalFilesFound, *retrieved.TotalFilesFound)
		assert.Equal(t, *sourceProcessing.ScanStatus, *retrieved.ScanStatus)
		assert.WithinDuration(t, sourceProcessing.CreatedAt, retrieved.CreatedAt, time.Second)
		assert.WithinDuration(t, sourceProcessing.UpdatedAt, retrieved.UpdatedAt, time.Second)
	})

	t.Run("GetSourceProcessing_NotFound", func(t *testing.T) {
		// Try to get non-existent source processing
		sp, err := repo.GetSourceProcessing(ctx, "nonexistent-source")
		assert.Error(t, err, "Expected error for non-existent source processing")
		assert.Nil(t, sp, "Expected nil for non-existent source processing")
	})

	t.Run("UpsertSourceProcessing_UpdateExisting", func(t *testing.T) {
		// First insert
		sp1 := testutils.CreateTestSourceProcessing("test-source-2")
		sp1.ScanStatus = stringPtr("scanning")
		err := repo.UpsertSourceProcessing(ctx, sp1)
		require.NoError(t, err, "Failed to insert source processing")

		// Update with new data
		sp2 := testutils.CreateTestSourceProcessing("test-source-2")
		sp2.ScanStatus = stringPtr("completed")
		sp2.TotalFilesFound = int64Ptr(20)
		err = repo.UpsertSourceProcessing(ctx, sp2)
		require.NoError(t, err, "Failed to update source processing")

		// Retrieve and verify update
		retrieved, err := repo.GetSourceProcessing(ctx, "test-source-2")
		require.NoError(t, err)
		assert.Equal(t, "completed", *retrieved.ScanStatus)
		assert.Equal(t, int64(20), *retrieved.TotalFilesFound)
	})

	t.Run("UpdateSourceProcessingStatus", func(t *testing.T) {
		// First insert
		sp := testutils.CreateTestSourceProcessing("test-source-3")
		err := repo.UpsertSourceProcessing(ctx, sp)
		require.NoError(t, err, "Failed to insert source processing")

		// Update specific fields
		updates := map[string]interface{}{
			"scan_status":       "failed",
			"error":             "scan failed due to timeout",
			"total_files_found": int64(5),
		}

		err = repo.UpdateSourceProcessingStatus(ctx, "test-source-3", updates)
		require.NoError(t, err, "Failed to update source processing status")

		// Retrieve and verify updates
		retrieved, err := repo.GetSourceProcessing(ctx, "test-source-3")
		require.NoError(t, err)
		assert.Equal(t, "failed", *retrieved.ScanStatus)
		assert.Equal(t, "scan failed due to timeout", *retrieved.Error)
		assert.Equal(t, int64(5), *retrieved.TotalFilesFound)
	})

	t.Run("UpdateSourceProcessingStatus_NotFound", func(t *testing.T) {
		// Try to update non-existent source processing
		updates := map[string]interface{}{
			"scan_status": "failed",
		}

		err := repo.UpdateSourceProcessingStatus(ctx, "nonexistent-source", updates)
		assert.Error(t, err, "Expected error for non-existent source processing")
	})

	t.Run("UpdateSourceProcessingStatus_EmptyUpdates", func(t *testing.T) {
		// Empty updates should not cause error
		err := repo.UpdateSourceProcessingStatus(ctx, "any-source", map[string]interface{}{})
		require.NoError(t, err, "Empty updates should not cause error")
	})

	t.Run("MultipleSourceProcessings", func(t *testing.T) {
		// Insert multiple source processings
		sources := []string{"source-a", "source-b", "source-c"}

		for _, sourceID := range sources {
			sp := testutils.CreateTestSourceProcessing(sourceID)
			sp.TotalFilesFound = int64Ptr(int64(len(sourceID) * 10)) // Different value for each
			err := repo.UpsertSourceProcessing(ctx, sp)
			require.NoError(t, err, "Failed to insert source processing for %s", sourceID)

			// Verify each can be retrieved
			retrieved, err := repo.GetSourceProcessing(ctx, sourceID)
			require.NoError(t, err, "Failed to get source processing for %s", sourceID)
			assert.Equal(t, sourceID, retrieved.SourceID)
			assert.Equal(t, int64(len(sourceID)*10), *retrieved.TotalFilesFound)
		}
	})

	t.Run("FullSourceProcessingLifecycle", func(t *testing.T) {
		sourceID := "full-lifecycle-source"

		// Phase 1: Start scan
		sp1 := &models.SourceProcessing{
			SourceID:      sourceID,
			ScanStartedAt: timePtr(time.Now()),
			ScanStatus:    stringPtr("scanning"),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		err := repo.UpsertSourceProcessing(ctx, sp1)
		require.NoError(t, err)

		retrieved1, err := repo.GetSourceProcessing(ctx, sourceID)
		require.NoError(t, err)
		assert.Equal(t, "scanning", *retrieved1.ScanStatus)
		assert.NotNil(t, retrieved1.ScanStartedAt)
		assert.Nil(t, retrieved1.ScanCompletedAt)

		// Phase 2: Complete scan
		updates1 := map[string]interface{}{
			"scan_status":       "completed",
			"scan_completed_at": time.Now(),
			"total_files_found": int64(100),
			"processing_status": "pending",
		}
		err = repo.UpdateSourceProcessingStatus(ctx, sourceID, updates1)
		require.NoError(t, err)

		retrieved2, err := repo.GetSourceProcessing(ctx, sourceID)
		require.NoError(t, err)
		assert.Equal(t, "completed", *retrieved2.ScanStatus)
		assert.Equal(t, int64(100), *retrieved2.TotalFilesFound)
		assert.Equal(t, "pending", *retrieved2.ProcessingStatus)

		// Phase 3: Start processing
		updates2 := map[string]interface{}{
			"processing_status":     "processing",
			"processing_started_at": time.Now(),
			"files_processed":       int64(50),
		}
		err = repo.UpdateSourceProcessingStatus(ctx, sourceID, updates2)
		require.NoError(t, err)

		retrieved3, err := repo.GetSourceProcessing(ctx, sourceID)
		require.NoError(t, err)
		assert.Equal(t, "processing", *retrieved3.ProcessingStatus)
		assert.Equal(t, int64(50), *retrieved3.FilesProcessed)

		// Phase 4: Complete processing with error
		updates3 := map[string]interface{}{
			"processing_status":       "failed",
			"processing_completed_at": time.Now(),
			"error":                   "processing failed: disk full",
		}
		err = repo.UpdateSourceProcessingStatus(ctx, sourceID, updates3)
		require.NoError(t, err)

		retrieved4, err := repo.GetSourceProcessing(ctx, sourceID)
		require.NoError(t, err)
		assert.Equal(t, "failed", *retrieved4.ProcessingStatus)
		assert.Equal(t, "processing failed: disk full", *retrieved4.Error)
		assert.NotNil(t, retrieved4.ProcessingCompletedAt)
	})
}

// Helper functions for pointer creation
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
