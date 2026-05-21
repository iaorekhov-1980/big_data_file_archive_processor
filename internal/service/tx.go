package service

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/repository/postgres"
)

// TxManager provides transaction coordination for service operations.
// It wraps a pgxpool.Pool and exposes a WithTransaction method.
type TxManager struct {
	pool *pgxpool.Pool
}

// NewTxManager creates a new TxManager instance.
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithTransaction executes the given function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// The function receives a context and a TxBundle containing transaction-aware repositories.
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx *TxBundle) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	bundle := &TxBundle{
		FileRepo:             postgres.NewFileRepository(tm.pool).WithTx(tx),
		FilePathRepo:         postgres.NewFilePathRepository(tm.pool).WithTx(tx),
		SourceProcessingRepo: postgres.NewSourceProcessingRepository(tm.pool).WithTx(tx),
	}

	if err := fn(ctx, bundle); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("transaction error: %v; rollback error: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// TxBundle holds transaction-aware repository instances.
// These use the same transaction connection for atomicity.
type TxBundle struct {
	FileRepo             *postgres.FileRepository
	FilePathRepo         *postgres.FilePathRepository
	SourceProcessingRepo *postgres.SourceProcessingRepository
}
