Here's the complete Phase 5 implementation prompt ready to use in a new agent session:

---

## Phase 5 Implementation Prompt

### Context

The project `big_data_file_archive_processor` has completed Phases 1-4:
- **Phase 1-3**: Project structure, models, PostgreSQL repository with composite PK `(path, source_id)` for `file_paths`, factory pattern, migration, test infrastructure
- **Phase 4**: Yandex Disk REST API client (`internal/disk/`) with `DiskClient` interface, rate limiting, error classification, factory, mock + real API tests

### What to Build: Phase 5 — Service Layer

Create `internal/service/` package with two services and a transaction manager.

### Architecture Decisions (Pre-Agreed)

| Decision | Choice |
|---|---|
| Service architecture | Domain-driven: separate `ScanService` and `CleanupService` |
| Transaction management | Service-level transaction wrapper using `pgxpool.Pool` |
| Duplicate resolution | Keep the **first** (oldest `createdAt`) `FilePath`, deactivate later duplicates |
| Dry-run mode | First-class feature for CleanupService |
| Config | `ScanPageSize` applies to both Yandex Disk API pagination and DB batch operations |
| Error handling | Retry transient errors (network, 429, 5xx) up to 3 times with backoff; fail on permanent errors (401, 404, validation) |
| Logging | Use `log/slog` with structured JSON, include `source_id` in context |

### Files to Create

#### 1. `internal/service/tx.go` — Transaction Manager

A thin wrapper that provides transactional execution for service operations.

```go
package service

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
)

// TxManager provides transaction coordination for service operations.
// It wraps a pgxpool.Pool and exposes a WithTransaction method.
type TxManager struct {
    pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager

// WithTransaction executes the given function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// The function receives a context and a TxBundle containing transaction-aware repositories.
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx *TxBundle) error) error

// TxBundle holds transaction-aware repository instances.
// These use the same transaction connection for atomicity.
type TxBundle struct {
    FileRepo              *postgres.FileRepository
    FilePathRepo          *postgres.FilePathRepository
    SourceProcessingRepo  *postgres.SourceProcessingRepository
}
```

**Design notes:**
- `TxBundle` reuses existing repository structs but needs a way to pass a transaction handle (`pgx.Tx`) instead of the pool
- Either: (a) add a `WithTx(tx)` method to each repository that returns a copy using the tx, or (b) create a `TxProvider` interface that repositories check
- **Recommendation**: Add `WithTx(tx pgx.Tx) *FileRepository` (etc.) methods to each repository — they return a new instance that uses the tx instead of the pool. This keeps the existing pool-based constructors unchanged.

#### 2. `internal/service/scanner.go` — ScanService

Coordinates `DiskClient` + `Repository` to scan a Yandex Disk source folder.

```go
package service

import (
    "context"
    "log/slog"
    "github.com/iaorekhov-1980/big_data_file_archive_processor/internal/disk"
    "github.com/iaorekhov-1980/big_data_file_archive_processor/internal/repository"
)

type ScanService struct {
    diskClient  disk.DiskClient
    repo        repository.Repository
    txManager   *TxManager
    pageSize    int
    logger      *slog.Logger
}

func NewScanService(diskClient disk.DiskClient, repo repository.Repository, txManager *TxManager, pageSize int, logger *slog.Logger) *ScanService

// ScanResult summarizes the scan operation outcome.
type ScanResult struct {
    SourceID        string
    TotalFilesFound int64
    FilesInserted   int64
    FilesSkipped    int64
    Errors          []string
}

// Scan performs a full scan of the given source folder on Yandex Disk.
// Parameters:
//   - ctx: context for cancellation
//   - sourceID: unique identifier for this source (e.g., "23_03_2026_my_source")
//   - sourceFolder: folder path on Yandex Disk to scan (empty = root)
//   - resumeOffset: if > 0, resume from this offset (from SourceProcessing.CurrentOffset)
//
// Behavior:
//   1. Load or create SourceProcessing record for sourceID
//   2. If scan already completed, return immediately (idempotent)
//   3. If scan was in progress (scanning status), resume from current_offset
//   4. Paginate through ListFiles using pageSize
//   5. For each file:
//      a. Check if FilePath exists by (path, source_id) — skip if exists
//      b. Check if File exists by hash — insert if not
//      c. Insert FilePath record
//   6. Update SourceProcessing on each page (offset tracking)
//   7. On completion, mark scan as completed with total count
//   8. On error, mark scan as failed with error message
func (s *ScanService) Scan(ctx context.Context, sourceID, sourceFolder string, resumeOffset int) (*ScanResult, error)
```

**Key behaviors:**
- Each page of results is processed in a transaction: insert files + file_paths + update source_processing offset
- If a `FilePath` with `(path, source_id)` already exists, skip it (idempotent)
- If a `File` with `hash` already exists, skip the insert (duplicate key is expected)
- Retry transient `DiskError`s (IsRateLimited, 5xx) up to 3 times per page
- Log each page processed with count and offset

#### 3. `internal/service/cleaner.go` — CleanupService

Handles duplicate file path cleanup and (future) empty folder cleanup.

```go
package service

type CleanupService struct {
    repo      repository.Repository
    txManager *TxManager
    logger    *slog.Logger
}

func NewCleanupService(repo repository.Repository, txManager *TxManager, logger *slog.Logger) *CleanupService

// CleanupResult summarizes the cleanup operation.
type CleanupResult struct {
    SourceID          string
    DuplicatesFound   int64
    DuplicatesRemoved int64
    Errors            []string
}

// CleanupDuplicates finds and deactivates duplicate file paths for a source.
// A duplicate is defined as two FilePath records with the same hash but different paths.
// The record with the earliest CreatedAt is kept active; all others are marked inactive.
//
// Parameters:
//   - ctx: context for cancellation
//   - sourceID: the source to clean up
//   - dryRun: if true, only log what would be done, make no DB changes
//
// Algorithm:
//   1. Fetch all active FilePath records for the source
//   2. Group by hash
//   3. For groups with count > 1:
//      a. Sort by CreatedAt ascending
//      b. Keep the first (oldest) as active
//      c. Mark all others as inactive (set is_active=false, deleted_at=NOW())
//   4. Update SourceProcessing cleanup phase status
func (s *CleanupService) CleanupDuplicates(ctx context.Context, sourceID string, dryRun bool) (*CleanupResult, error)
```

**Key behaviors:**
- In dry-run mode: log each duplicate that *would* be removed, return counts, make no DB changes
- In live mode: each group of duplicates is processed in a single transaction
- Update `SourceProcessing` cleanup status on completion

#### 4. `internal/service/retry.go` — Retry Utility

```go
package service

// RetryConfig defines retry behavior for transient errors.
type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
}

// DefaultRetryConfig returns sensible defaults: 3 attempts, 100ms base, 2s max.
func DefaultRetryConfig() RetryConfig

// DoWithRetry executes the given function, retrying on transient errors.
// A transient error is one that implements:
//   - disk.DiskError with IsRateLimited() == true, or
//   - any error wrapped in a RetryableError
// Uses exponential backoff: delay = min(baseDelay * 2^attempt, maxDelay)
func DoWithRetry(ctx context.Context, config RetryConfig, operation func() error) error
```

### Test Files

#### 5. `internal/service/scanner_test.go`

- Mock `DiskClient` using a simple struct implementing the interface
- Test scenarios:
  - Successful scan of empty folder
  - Successful scan with multiple pages
  - Resume from offset (simulate interrupted scan)
  - Skip already-imported files (duplicate FilePath)
  - Skip already-known File hashes (duplicate hash)
  - Transient API error with retry success
  - Permanent API error (401) — fail immediately
  - Context cancellation mid-scan
  - Verify SourceProcessing status updates at each stage

#### 6. `internal/service/cleaner_test.go`

- Test scenarios:
  - No duplicates — nothing to clean
  - Single duplicate group — keep oldest, deactivate newest
  - Multiple duplicate groups
  - Dry-run mode — no DB changes, correct counts reported
  - Source with no records — no-op
  - Verify SourceProcessing cleanup status update

### Files to Modify

#### 7. `internal/repository/postgres/file_repository.go` — Add `WithTx` method

```go
// WithTx returns a new FileRepository that uses the given transaction.
func (r *FileRepository) WithTx(tx pgx.Tx) *FileRepository {
    return &FileRepository{pool: &pgxpool.Pool{}} // or use a TxQuerier interface
}
```

**Note**: The actual implementation needs careful design. The repositories currently use `*pgxpool.Pool` directly. Options:
- **Option A**: Change repository methods to accept a `Querier` interface (pgxpool.Pool and pgx.Tx both implement it). This is the cleanest approach but requires refactoring all repository methods.
- **Option B**: Add `WithTx()` methods that return a copy using a `pgx.Tx`. Requires a `Querier` interface internally.
- **Recommendation**: Option A — define a `Querier` interface in the postgres package with `QueryRow`, `Exec`, `Query` methods. Both `*pgxpool.Pool` and `pgx.Tx` satisfy it. Change repository structs to hold a `Querier` instead of `*pgxpool.Pool`. This is a minimal refactor and enables clean transaction support.

#### 8. `internal/repository/postgres/postgres.go` — Minor update if Querier pattern is used

### What NOT to Build

- **Orchestrator** — deferred to later phase
- **Folder cleanup** — deferred (CleanupService only handles duplicates for now)
- **Circuit breaker** — not needed yet
- **Metrics/monitoring** — not needed yet
- **Command binaries** (`cmd/scan_source/`, etc.) — will be Phase 6

### Implementation Order

1. Add `Querier` interface to `internal/repository/postgres/` + refactor repositories to use it
2. Add `WithTx()` methods to each repository
3. Create `internal/service/tx.go` — TxManager + TxBundle
4. Create `internal/service/retry.go` — retry utility
5. Create `internal/service/scanner.go` — ScanService
6. Create `internal/service/scanner_test.go` — tests
7. Create `internal/service/cleaner.go` — CleanupService
8. Create `internal/service/cleaner_test.go` — tests

### Verification

- All existing tests continue to pass (`go test ./...`)
- New service tests pass with mocked dependencies
- Integration tests can be run manually against real DB + Yandex Disk API

