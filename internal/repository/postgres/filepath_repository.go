package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/models"
	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/repository"
)

// FilePathRepository handles file path-related database operations
type FilePathRepository struct {
	pool *pgxpool.Pool
}

// NewFilePathRepository creates a new FilePathRepository instance
func NewFilePathRepository(pool *pgxpool.Pool) *FilePathRepository {
	return &FilePathRepository{pool: pool}
}

// GetFilePathByPathAndSource retrieves a file path by its path and source ID
func (r *FilePathRepository) GetFilePathByPathAndSource(ctx context.Context, path, sourceID string) (*models.FilePath, error) {
	query := `
		SELECT path, source_id, hash, is_active, deleted_at, created_at, updated_at
		FROM file_paths
		WHERE path = $1 AND source_id = $2
	`

	var filePath models.FilePath
	var deletedAt *time.Time
	err := r.pool.QueryRow(ctx, query, path, sourceID).Scan(
		&filePath.Path,
		&filePath.SourceID,
		&filePath.Hash,
		&filePath.IsActive,
		&deletedAt,
		&filePath.CreatedAt,
		&filePath.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.NewNotFoundError("file_path", path+"@"+sourceID)
		}
		return nil, repository.NewRepositoryError("GetFilePathByPathAndSource", err)
	}

	filePath.DeletedAt = deletedAt
	return &filePath, nil
}

// InsertFilePath inserts a new file path record
func (r *FilePathRepository) InsertFilePath(ctx context.Context, filePath *models.FilePath) error {
	query := `
		INSERT INTO file_paths (path, source_id, hash, is_active, deleted_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		filePath.Path,
		filePath.SourceID,
		filePath.Hash,
		filePath.IsActive,
		filePath.DeletedAt,
		filePath.CreatedAt,
		filePath.UpdatedAt,
	)

	if err != nil {
		// Check for duplicate key error
		if isDuplicateKeyError(err) {
			return repository.NewDuplicateError("file_path", filePath.Path+"@"+filePath.SourceID)
		}
		return repository.NewRepositoryError("InsertFilePath", err)
	}

	return nil
}

// UpdateFilePathDeletedAtByPathAndSource updates the deleted_at timestamp for a file path
func (r *FilePathRepository) UpdateFilePathDeletedAtByPathAndSource(ctx context.Context, path, sourceID string, deletedAt *time.Time) error {
	query := `
		UPDATE file_paths
		SET deleted_at = $1, updated_at = NOW(), is_active = false
		WHERE path = $2 AND source_id = $3
	`

	result, err := r.pool.Exec(ctx, query, deletedAt, path, sourceID)
	if err != nil {
		return repository.NewRepositoryError("UpdateFilePathDeletedAtByPathAndSource", err)
	}

	if result.RowsAffected() == 0 {
		return repository.NewNotFoundError("file_path", path+"@"+sourceID)
	}

	return nil
}

// GetFilePathsBySourceAndActive retrieves file paths by source ID and active status
func (r *FilePathRepository) GetFilePathsBySourceAndActive(ctx context.Context, sourceID string, isActive bool) ([]*models.FilePath, error) {
	query := `
		SELECT path, source_id, hash, is_active, deleted_at, created_at, updated_at
		FROM file_paths
		WHERE source_id = $1 AND is_active = $2
		ORDER BY path
	`

	rows, err := r.pool.Query(ctx, query, sourceID, isActive)
	if err != nil {
		return nil, repository.NewRepositoryError("GetFilePathsBySourceAndActive", err)
	}
	defer rows.Close()

	var filePaths []*models.FilePath
	for rows.Next() {
		var fp models.FilePath
		var deletedAt *time.Time
		err := rows.Scan(
			&fp.Path,
			&fp.SourceID,
			&fp.Hash,
			&fp.IsActive,
			&deletedAt,
			&fp.CreatedAt,
			&fp.UpdatedAt,
		)
		if err != nil {
			return nil, repository.NewRepositoryError("GetFilePathsBySourceAndActive", err)
		}
		fp.DeletedAt = deletedAt
		filePaths = append(filePaths, &fp)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.NewRepositoryError("GetFilePathsBySourceAndActive", err)
	}

	return filePaths, nil
}

// GetFilePathsBySourceAndInactiveNotDeleted retrieves inactive file paths that are not marked as deleted
func (r *FilePathRepository) GetFilePathsBySourceAndInactiveNotDeleted(ctx context.Context, sourceID string) ([]*models.FilePath, error) {
	query := `
		SELECT path, source_id, hash, is_active, deleted_at, created_at, updated_at
		FROM file_paths
		WHERE source_id = $1 AND is_active = false AND deleted_at IS NULL
		ORDER BY path
	`

	rows, err := r.pool.Query(ctx, query, sourceID)
	if err != nil {
		return nil, repository.NewRepositoryError("GetFilePathsBySourceAndInactiveNotDeleted", err)
	}
	defer rows.Close()

	var filePaths []*models.FilePath
	for rows.Next() {
		var fp models.FilePath
		var deletedAt *time.Time
		err := rows.Scan(
			&fp.Path,
			&fp.SourceID,
			&fp.Hash,
			&fp.IsActive,
			&deletedAt,
			&fp.CreatedAt,
			&fp.UpdatedAt,
		)
		if err != nil {
			return nil, repository.NewRepositoryError("GetFilePathsBySourceAndInactiveNotDeleted", err)
		}
		fp.DeletedAt = deletedAt
		filePaths = append(filePaths, &fp)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.NewRepositoryError("GetFilePathsBySourceAndInactiveNotDeleted", err)
	}

	return filePaths, nil
}
