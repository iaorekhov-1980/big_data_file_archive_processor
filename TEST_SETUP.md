# Database Test Setup

## Quick Start

1. **Setup configuration**:
   ```bash
   make setup-config
   # Edit test_config.toml with your database credentials
   ```

2. **Test connection**:
   ```bash
   make test-connection
   ```

3. **Run tests**:
   ```bash
   make test-all
   ```

## Detailed Steps

### 1. Configuration Setup

Create configuration file from template:
```bash
# Using make
make setup-config

# Using Go
go run internal/testutils/setup_test_config.go

# Manual
cp test_config.example.toml test_config.toml
```

Edit `test_config.toml`:
```toml
[postgres]
host = "localhost"
port = 5432
database = "ivordb"
username = "gouser"
password = "gopwd"
```

**Note**: `test_config.toml` is in `.gitignore`.

### 2. Database Connection Test

Verify database connection and schema:
```bash
make test-connection
```

Expected output shows connection success and table verification.

### 3. Running Tests

#### All Tests
```bash
make test-all
```

#### Specific Test Categories
```bash
# File operations
make test-file

# File path operations  
make test-filepath

# Source processing operations
make test-sourceprocessing
```

#### Test Options
```bash
# Skip integration tests (unit tests only)
make test-short

# With coverage report
make test-coverage
```

## Alternative Methods

### PowerShell
```powershell
# Setup
.\run_tests.ps1 -Setup

# Run tests
.\run_tests.ps1

# Specific test
.\run_tests.ps1 -TestName TestFileRepository_Integration

# With coverage
.\run_tests.ps1 -Coverage
```

### Batch File
```cmd
# Setup
run_tests.bat --setup

# Run tests
run_tests.bat

# Specific test
run_tests.bat TestFileRepository_Integration
```

### Direct Go Commands
```bash
# Setup
go run internal/testutils/setup_test_config.go

# Test connection
go run internal/testutils/test_db_connection.go

# Run tests
go test ./internal/repository/postgres -v
```

## Test Files

- `internal/repository/postgres/file_repository_integration_test.go` - File operations
- `internal/repository/postgres/filepath_repository_integration_test.go` - File path operations
- `internal/repository/postgres/sourceprocessing_repository_integration_test.go` - Source processing
- `internal/repository/postgres/postgres_repository_integration_test.go` - Complete repository

## Troubleshooting

### Configuration Not Found
```
Error: test_config.toml not found
```
**Solution**: Run `make setup-config`

### Database Connection Failed
```
ERROR: Failed to ping database
```
**Solution**:
1. Check PostgreSQL is running
2. Verify credentials in `test_config.toml`
3. Test with `make test-connection`

### Schema Not Found
```
Table 'files' does not exist
```
**Solution**: Run migration:
```bash
psql -U gouser -d ivordb -f migrations/000001_init.up.sql
```

### Environment Variable Override
Set `TEST_POSTGRES_DSN` to override config file:
```bash
export TEST_POSTGRES_DSN="postgres://user:pass@host:port/db?sslmode=disable"
go test ./internal/repository/postgres -v
```