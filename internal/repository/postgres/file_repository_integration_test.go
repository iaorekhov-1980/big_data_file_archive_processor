package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	pool, cleanup := SetupTestDatabase(t, ctx)
	defer cleanup()

	repo := NewFileRepository(pool)

	t.Run("InsertFile and GetFileByHash", func(t *testing.T) {
		// Create test file
		file := createTestFile("testhash123", "testfile.txt")

		// Insert file
		err := repo.InsertFile(ctx, file)
		require.NoError(t, err, "Failed to insert file")

		// Retrieve file
		retrievedFile, err := repo.GetFileByHash(ctx, "testhash123")
		require.NoError(t, err, "Failed to get file by hash")

		// Verify file data
		assert.Equal(t, file.Hash, retrievedFile.Hash)
		assert.Equal(t, file.Name, retrievedFile.Name)
		assert.Equal(t, file.Size, retrievedFile.Size)
		assert.Equal(t, file.MimeType, retrievedFile.MimeType)
		assert.WithinDuration(t, file.CreatedAt, retrievedFile.CreatedAt, time.Second)
		assert.WithinDuration(t, file.UpdatedAt, retrievedFile.UpdatedAt, time.Second)
	})

	t.Run("GetFileByHash_NotFound", func(t *testing.T) {
		// Try to get non-existent file
		file, err := repo.GetFileByHash(ctx, "nonexistenthash")
		assert.Error(t, err, "Expected error for non-existent file")
		assert.Nil(t, file, "Expected nil file for non-existent hash")

		// Verify it's a NotFoundError
		_, ok := err.(interface{ Resource() string })
		assert.True(t, ok, "Expected NotFoundError type")
	})

	t.Run("InsertFile_Duplicate", func(t *testing.T) {
		// Create and insert first file
		file1 := createTestFile("duplicatehash", "file1.txt")
		err := repo.InsertFile(ctx, file1)
		require.NoError(t, err, "Failed to insert first file")

		// Try to insert duplicate file with same hash
		file2 := createTestFile("duplicatehash", "file2.txt")
		err = repo.InsertFile(ctx, file2)
		assert.Error(t, err, "Expected error for duplicate file")

		// Verify it's a DuplicateError
		_, ok := err.(interface{ Resource() string })
		assert.True(t, ok, "Expected DuplicateError type")
	})

	t.Run("InsertFile_MultipleFiles", func(t *testing.T) {
		// Insert multiple files with different hashes
		files := []struct {
			hash string
			name string
		}{
			{"hash1", "file1.txt"},
			{"hash2", "file2.txt"},
			{"hash3", "file3.txt"},
		}

		for _, f := range files {
			file := createTestFile(f.hash, f.name)
			err := repo.InsertFile(ctx, file)
			require.NoError(t, err, "Failed to insert file %s", f.name)

			// Verify each file can be retrieved
			retrieved, err := repo.GetFileByHash(ctx, f.hash)
			require.NoError(t, err, "Failed to get file %s", f.name)
			assert.Equal(t, f.name, retrieved.Name)
		}
	})
}
