## Updated Prompt for Phase 3: Database Layer Implementation

**Title:** Phase 3: Database Layer Implementation

**Context:**
We are implementing Phase 3 of the ODS module. Phases 1-2 are complete with:
- Project structure and configuration files
- Data models (`File`, `FilePath`, `SourceProcessing`)
- Repository interface with all required methods
- Configuration loader with environment variables

**Current Status:**
- Repository: `https://github.com/iaorekhov-1980/big_data_file_archive_processor`
- Branch: `main`
- Last commit: "Add name field to File model for version identification"
- Ready for Phase 3 implementation

**Clarifications Received:**
1. **File Name Field**: Keep `name` field in `File` model to track multiple versions of same file. Hash remains primary key.
2. **SourceProcessing Table**: Restore full version with all original fields from prompt.
3. **Schema Alignment**: Update models to match full schema, then create migrations.
4. **Indexes**: Include all indexes from original schema.

**Phase 3 Tasks:**

**Step 3.0: Update Models to Match Full Schema** (PREREQUISITE)
- Update `internal/models/models.go`:
  - Keep `name` field in `File` struct
  - Expand `SourceProcessing` struct with all original fields:
    - `scan_started_at`, `scan_completed_at`, `current_offset`, `total_files_found`, `scan_status`, `scan_error`
    - `processing_started_at`, `processing_completed_at`, `files_processed`, `files_deleted`, `processing_status`, `processing_error`
    - `cleanup_started_at`, `cleanup_completed_at`, `folders_deleted`, `cleanup_status`, `cleanup_error`
  - Update constructor and helper methods accordingly

**Step 3.1: Create PostgreSQL Migration Files**
- Create `migrations/000001_init.up.sql` with table definitions:
  - `files` table: `hash` (PK), `name`, `size`, `mime_type`, `created_at`, `updated_at`
  - `file_paths` table: `path` (PK), `source_id`, `hash` (FK), `is_active`, `deleted_at`, `created_at`, `updated_at`
  - `source_processing` table: All fields from original schema
  - Indexes: `idx_file_paths_hash`, `idx_file_paths_source_active`
- Create `migrations/000001_init.down.sql` for rollback (DROP TABLE statements)

**Step 3.2: Implement PostgreSQL Repository**
- Create `internal/repository/postgres/postgres.go`
- Implement all methods from `Repository` interface using `pgx/v5`
- Struct: `PostgresRepository` with `*pgx.Conn` or connection pool
- Methods to implement:
  - `GetFileByHash`, `InsertFile`
  - `GetFilePath`, `InsertFilePath`, `UpdateFilePathDeletedAt`, `GetFilePathsBySourceAndActive`, `GetFilePathsBySourceAndInactiveNotDeleted`
  - `GetSourceProcessing`, `UpsertSourceProcessing`, `UpdateSourceProcessingStatus`
- Add proper connection handling, error mapping, and transaction support

**Step 3.3: Create Repository Factory**
- Create `internal/repository/factory.go`
- Factory function: `NewRepository(ctx context.Context, config *config.Config) (Repository, error)`
- Based on `config.DBType` returns appropriate implementation
- Currently only PostgreSQL implementation (`NewPostgresRepository`)
- Prepare structure for future YDB support

**Technical Specifications:**

**Database Schema:**
```sql
-- files table
hash VARCHAR(64) PRIMARY KEY,
name TEXT NOT NULL,
size BIGINT NOT NULL,
mime_type TEXT NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

-- file_paths table  
path TEXT PRIMARY KEY,
source_id TEXT NOT NULL,
hash VARCHAR(64) NOT NULL REFERENCES files(hash),
is_active BOOLEAN NOT NULL,
deleted_at TIMESTAMPTZ,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

-- source_processing table (full version)
source_id TEXT PRIMARY KEY,
scan_started_at TIMESTAMPTZ,
scan_completed_at TIMESTAMPTZ,
current_offset BIGINT,
total_files_found BIGINT,
scan_status TEXT,
scan_error TEXT,
processing_started_at TIMESTAMPTZ,
processing_completed_at TIMESTAMPTZ,
files_processed BIGINT,
files_deleted BIGINT,
processing_status TEXT,
processing_error TEXT,
cleanup_started_at TIMESTAMPTZ,
cleanup_completed_at TIMESTAMPTZ,
folders_deleted BIGINT,
cleanup_status TEXT,
cleanup_error TEXT,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

**Indexes:**
- `CREATE INDEX idx_file_paths_hash ON file_paths(hash);`
- `CREATE INDEX idx_file_paths_source_active ON file_paths(source_id, is_active);`

**Requirements:**
- Follow Go best practices and conventions
- Add comprehensive comments
- Include proper error handling (map SQL errors to repository error types)
- Use connection pooling for PostgreSQL
- Commit after each step
- Push to GitHub after each step

**Git Commit Strategy:**
- Step 3.0: "Phase 3.0: Update models to full schema"
- Step 3.1: "Phase 3.1: Create database migration files"
- Step 3.2: "Phase 3.2: Implement PostgreSQL repository"
- Step 3.3: "Phase 3.3: Create repository factory"

**Dependencies to Add:**
- `github.com/jackc/pgx/v5` (already in go.mod)
- Consider adding `github.com/jackc/pgx/v5/pgxpool` for connection pooling

**Ready to proceed with Step 3.0.**

