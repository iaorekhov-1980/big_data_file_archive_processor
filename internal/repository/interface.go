package repository

import (
	"context"
	"time"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/models"
)

// Repository defines the interface for data access operations
type Repository interface {
	// File operations
	GetFileByHash(ctx context.Context, hash string) (*models.File, error)
	InsertFile(ctx context.Context, file *models.File) error

	// FilePath operations
	GetFilePath(ctx context.Context, path string) (*models.FilePath, error)
	InsertFilePath(ctx context.Context, filePath *models.FilePath) error
	UpdateFilePathDeletedAt(ctx context.Context, path string, deletedAt *time.Time) error
	GetFilePathsBySourceAndActive(ctx context.Context, sourceID string, isActive bool) ([]*models.FilePath, error)
	GetFilePathsBySourceAndInactiveNotDeleted(ctx context.Context, sourceID string) ([]*models.FilePath, error)

	// SourceProcessing operations
	GetSourceProcessing(ctx context.Context, sourceID string) (*models.SourceProcessing, error)
	UpsertSourceProcessing(ctx context.Context, sourceProcessing *models.SourceProcessing) error
	UpdateSourceProcessingStatus(ctx context.Context, sourceID string, updates map[string]interface{}) error
}

// Error types for repository operations
type RepositoryError struct {
	Operation string
	Err       error
}

func (e *RepositoryError) Error() string {
	return "repository error during " + e.Operation + ": " + e.Err.Error()
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new RepositoryError
func NewRepositoryError(operation string, err error) *RepositoryError {
	return &RepositoryError{
		Operation: operation,
		Err:       err,
	}
}

// NotFoundError represents when a resource is not found
type NotFoundError struct {
	Resource string
	Key      string
}

func (e *NotFoundError) Error() string {
	return e.Resource + " with key '" + e.Key + "' not found"
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource, key string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		Key:      key,
	}
}

// DuplicateError represents when a duplicate resource is being inserted
type DuplicateError struct {
	Resource string
	Key      string
}

func (e *DuplicateError) Error() string {
	return e.Resource + " with key '" + e.Key + "' already exists"
}

// NewDuplicateError creates a new DuplicateError
func NewDuplicateError(resource, key string) *DuplicateError {
	return &DuplicateError{
		Resource: resource,
		Key:      key,
	}
}