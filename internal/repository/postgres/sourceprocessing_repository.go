package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/models"
	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/repository"
)

// SourceProcessingRepository handles source processing-related database operations
type SourceProcessingRepository struct {
	q Querier
}

// NewSourceProcessingRepository creates a new SourceProcessingRepository instance
func NewSourceProcessingRepository(pool *pgxpool.Pool) *SourceProcessingRepository {
	return &SourceProcessingRepository{q: pool}
}

// WithTx returns a new SourceProcessingRepository that uses the given transaction.
func (r *SourceProcessingRepository) WithTx(tx pgx.Tx) *SourceProcessingRepository {
	return &SourceProcessingRepository{q: tx}
}

// GetSourceProcessing retrieves source processing status by source ID
func (r *SourceProcessingRepository) GetSourceProcessing(ctx context.Context, sourceID string) (*models.SourceProcessing, error) {
	query := `
		SELECT 
			source_id,
			scan_started_at, scan_completed_at, current_offset, total_files_found, scan_status,
			processing_started_at, processing_completed_at, files_processed, files_deleted, processing_status,
			cleanup_started_at, cleanup_completed_at, folders_deleted, cleanup_status,
			error,
			created_at, updated_at
		FROM source_processing
		WHERE source_id = $1
	`

	var sp models.SourceProcessing
	err := r.q.QueryRow(ctx, query, sourceID).Scan(
		&sp.SourceID,
		&sp.ScanStartedAt, &sp.ScanCompletedAt, &sp.CurrentOffset, &sp.TotalFilesFound, &sp.ScanStatus,
		&sp.ProcessingStartedAt, &sp.ProcessingCompletedAt, &sp.FilesProcessed, &sp.FilesDeleted, &sp.ProcessingStatus,
		&sp.CleanupStartedAt, &sp.CleanupCompletedAt, &sp.FoldersDeleted, &sp.CleanupStatus,
		&sp.Error,
		&sp.CreatedAt, &sp.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.NewNotFoundError("source_processing", sourceID)
		}
		return nil, repository.NewRepositoryError("GetSourceProcessing", err)
	}

	return &sp, nil
}

// UpsertSourceProcessing inserts or updates a source processing record
func (r *SourceProcessingRepository) UpsertSourceProcessing(ctx context.Context, sourceProcessing *models.SourceProcessing) error {
	query := `
		INSERT INTO source_processing (
			source_id,
			scan_started_at, scan_completed_at, current_offset, total_files_found, scan_status,
			processing_started_at, processing_completed_at, files_processed, files_deleted, processing_status,
			cleanup_started_at, cleanup_completed_at, folders_deleted, cleanup_status,
			error,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
		ON CONFLICT (source_id) DO UPDATE SET
			scan_started_at = EXCLUDED.scan_started_at,
			scan_completed_at = EXCLUDED.scan_completed_at,
			current_offset = EXCLUDED.current_offset,
			total_files_found = EXCLUDED.total_files_found,
			scan_status = EXCLUDED.scan_status,
			processing_started_at = EXCLUDED.processing_started_at,
			processing_completed_at = EXCLUDED.processing_completed_at,
			files_processed = EXCLUDED.files_processed,
			files_deleted = EXCLUDED.files_deleted,
			processing_status = EXCLUDED.processing_status,
			cleanup_started_at = EXCLUDED.cleanup_started_at,
			cleanup_completed_at = EXCLUDED.cleanup_completed_at,
			folders_deleted = EXCLUDED.folders_deleted,
			cleanup_status = EXCLUDED.cleanup_status,
			error = EXCLUDED.error,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.q.Exec(ctx, query,
		sourceProcessing.SourceID,
		sourceProcessing.ScanStartedAt, sourceProcessing.ScanCompletedAt, sourceProcessing.CurrentOffset, sourceProcessing.TotalFilesFound, sourceProcessing.ScanStatus,
		sourceProcessing.ProcessingStartedAt, sourceProcessing.ProcessingCompletedAt, sourceProcessing.FilesProcessed, sourceProcessing.FilesDeleted, sourceProcessing.ProcessingStatus,
		sourceProcessing.CleanupStartedAt, sourceProcessing.CleanupCompletedAt, sourceProcessing.FoldersDeleted, sourceProcessing.CleanupStatus,
		sourceProcessing.Error,
		sourceProcessing.CreatedAt, sourceProcessing.UpdatedAt,
	)

	if err != nil {
		return repository.NewRepositoryError("UpsertSourceProcessing", err)
	}

	return nil
}

// UpdateSourceProcessingStatus updates specific fields of a source processing record
func (r *SourceProcessingRepository) UpdateSourceProcessingStatus(ctx context.Context, sourceID string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic update query
	query := "UPDATE source_processing SET "
	args := []interface{}{}
	argIndex := 1

	for field, value := range updates {
		query += fmt.Sprintf("%s = $%d, ", field, argIndex)
		args = append(args, value)
		argIndex++
	}

	// Add updated_at and WHERE clause
	query += "updated_at = NOW() WHERE source_id = $" + fmt.Sprintf("%d", argIndex)
	args = append(args, sourceID)

	result, err := r.q.Exec(ctx, query, args...)
	if err != nil {
		return repository.NewRepositoryError("UpdateSourceProcessingStatus", err)
	}

	if result.RowsAffected() == 0 {
		return repository.NewNotFoundError("source_processing", sourceID)
	}

	return nil
}
