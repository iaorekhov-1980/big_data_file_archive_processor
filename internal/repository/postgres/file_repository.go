package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/models"
	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/repository"
)

// FileRepository handles file-related database operations
type FileRepository struct {
	q Querier
}

// NewFileRepository creates a new FileRepository instance
func NewFileRepository(pool *pgxpool.Pool) *FileRepository {
	return &FileRepository{q: pool}
}

// WithTx returns a new FileRepository that uses the given transaction.
func (r *FileRepository) WithTx(tx pgx.Tx) *FileRepository {
	return &FileRepository{q: tx}
}

// GetFileByHash retrieves a file by its hash
func (r *FileRepository) GetFileByHash(ctx context.Context, hash string) (*models.File, error) {
	query := `
		SELECT hash, name, size, mime_type, created_at, updated_at
		FROM files
		WHERE hash = $1
	`

	var file models.File
	err := r.q.QueryRow(ctx, query, hash).Scan(
		&file.Hash,
		&file.Name,
		&file.Size,
		&file.MimeType,
		&file.CreatedAt,
		&file.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.NewNotFoundError("file", hash)
		}
		return nil, repository.NewRepositoryError("GetFileByHash", err)
	}

	return &file, nil
}

// InsertFile inserts a new file record
func (r *FileRepository) InsertFile(ctx context.Context, file *models.File) error {
	query := `
		INSERT INTO files (hash, name, size, mime_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.q.Exec(ctx, query,
		file.Hash,
		file.Name,
		file.Size,
		file.MimeType,
		file.CreatedAt,
		file.UpdatedAt,
	)

	if err != nil {
		// Check for duplicate key error
		if isDuplicateKeyError(err) {
			return repository.NewDuplicateError("file", file.Hash)
		}
		return repository.NewRepositoryError("InsertFile", err)
	}

	return nil
}
