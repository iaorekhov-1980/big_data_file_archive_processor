# Database Integration Tests

## Setup

1. Ensure PostgreSQL is running
2. Create database schema using migrations:
   ```bash
   psql -d your_database -f migrations/000001_init.up.sql
   ```
3. Set environment variable:
   ```bash
   export TEST_POSTGRES_DSN="postgres://user:pass@localhost:5432/your_database?sslmode=disable"
   ```

## Run Tests

```bash
# Run all tests
go test ./internal/repository/postgres -v

# Run specific test
go test ./internal/repository/postgres -v -run TestFileRepository_Integration

# Skip integration tests
go test ./internal/repository/postgres -short
```

## Test Files

- `file_repository_integration_test.go` - File operations
- `filepath_repository_integration_test.go` - File path operations  
- `sourceprocessing_repository_integration_test.go` - Source processing operations
- `postgres_repository_integration_test.go` - Complete repository tests

## Notes

- Tests clean up data after each run
- Schema must exist before running tests
- Uses composite primary key (path, source_id) for file_paths table