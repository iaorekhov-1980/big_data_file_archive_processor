package testutils

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDBConnection tests PostgreSQL connection and verifies schema
func TestDBConnection() error {
	// Load configuration
	config, err := LoadTestConfig()
	if err != nil {
		return fmt.Errorf("failed to load test config: %w", err)
	}

	fmt.Printf("Testing PostgreSQL connection...\n")
	fmt.Printf("Database: %s\n", config.Postgres.Database)
	fmt.Printf("Host: %s:%d\n", config.Postgres.Host, config.Postgres.Port)
	fmt.Printf("User: %s\n", config.Postgres.Username)
	fmt.Println()

	dsn := config.GetPostgresDSN()
	fmt.Printf("DSN: %s\n", dsn)
	fmt.Println()

	// Test connection
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("SUCCESS: Connected to PostgreSQL database!")
	fmt.Println()

	// Test schema
	fmt.Println("Checking database schema...")
	requiredTables := []string{"files", "file_paths", "source_processing"}
	allTablesExist := true

	for _, table := range requiredTables {
		query := `SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`

		var exists bool
		err := pool.QueryRow(ctx, query, table).Scan(&exists)
		if err != nil {
			fmt.Printf("ERROR: Failed to check table %s: %v\n", table, err)
			allTablesExist = false
			continue
		}

		if exists {
			fmt.Printf("  ✓ Table '%s' exists\n", table)
		} else {
			fmt.Printf("  ✗ Table '%s' does not exist\n", table)
			allTablesExist = false
		}
	}

	fmt.Println()
	if allTablesExist {
		fmt.Println("SUCCESS: All required tables exist!")
		fmt.Println("Database is ready for testing.")
		return nil
	} else {
		return fmt.Errorf("some required tables are missing. Please run: psql -U gouser -d ivordb -f migrations/000001_init.up.sql")
	}
}

// PrintConnectionHelp prints usage information for connection test command
func PrintConnectionHelp() {
	fmt.Println("Usage: go run cmd/testutils/main.go test-connection")
	fmt.Println("Tests PostgreSQL connection and verifies schema")
}
