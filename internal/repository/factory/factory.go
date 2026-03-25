package factory

import (
	"context"
	"fmt"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/config"
	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/repository"
	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/repository/postgres"
)

// NewRepository creates a new repository instance based on configuration
func NewRepository(ctx context.Context, config *config.Config) (repository.Repository, error) {
	switch config.DBType {
	case "postgres":
		return postgres.NewPostgresRepository(ctx, config.GetPostgresDSN())
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.DBType)
	}
}