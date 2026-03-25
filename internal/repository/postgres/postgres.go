package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/repository"
)

// PostgresRepository implements the Repository interface using PostgreSQL
// It composes separate repositories for each entity type
type PostgresRepository struct {
	pool *pgxpool.Pool
	
	// Embedded repositories
	*FileRepository
	*FilePathRepository
	*SourceProcessingRepository
}

// NewPostgresRepository creates a new PostgreSQL repository instance
func NewPostgresRepository(ctx context.Context, dsn string) (*PostgresRepository, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create separate repositories
	fileRepo := NewFileRepository(pool)
	filePathRepo := NewFilePathRepository(pool)
	sourceProcessingRepo := NewSourceProcessingRepository(pool)

	return &PostgresRepository{
		pool:                        pool,
		FileRepository:              fileRepo,
		FilePathRepository:          filePathRepo,
		SourceProcessingRepository:  sourceProcessingRepo,
	}, nil
}

// Close closes the database connection pool
func (r *PostgresRepository) Close() {
	if r.pool != nil {
		r.pool.Close()
	}
}

// Verify that PostgresRepository implements the Repository interface
var _ repository.Repository = (*PostgresRepository)(nil)