package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// isDuplicateKeyError checks if the error is a duplicate key violation
func isDuplicateKeyError(err error) bool {
	// PostgreSQL duplicate key error code is 23505
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return false
}