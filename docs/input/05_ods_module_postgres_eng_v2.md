Below is the revised English prompt. It removes any Go code specifics, adds the table schema, and leaves the implementation details to the agent.

---

**Prompt for Agent (English) – ODS Module with PostgreSQL Abstraction**

---

### Context
We are building a file archive processing system on Yandex Disk (Go). The first module is **ODS**: scanning sources, identifying duplicates (global), physically deleting duplicates, and removing empty folders. To avoid early costs, we use **local PostgreSQL** (Docker) for metadata storage; later we will add YDB support. Code must be abstracted so switching DB is easy.

The GitHub repository `big_data_file_archive_processor` is currently empty. You will generate the complete project structure, code, migrations, and documentation.

### Environment
- VS Code with Go, YAML, Terraform, Continue (DeepSeek) plugins.
- Docker running PostgreSQL container.
- Yandex CLI installed (not used yet).

---

### Requirements

#### 1. Project Structure (create from scratch)
```
big_data_file_archive_processor/
├── cmd/
│   ├── scan_source/
│   │   └── main.go
│   ├── cleanup_duplicates/
│   │   └── main.go
│   └── cleanup_folders/
│       └── main.go
├── internal/
│   ├── models/
│   │   └── models.go
│   ├── repository/
│   │   ├── interface.go
│   │   ├── postgres/
│   │   │   └── postgres.go
│   │   └── factory.go
│   ├── disk/
│   │   └── disk.go
│   └── config/
│       └── config.go
├── migrations/
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
├── terraform/ (empty for now)
├── docs/
├── go.mod
├── .gitignore
├── README.md
└── .env.example
```

#### 2. Database Abstraction
Define a `Repository` interface (in `internal/repository/interface.go`) with methods for:
- `GetFileByHash(ctx, hash) (*models.File, error)`
- `InsertFile(ctx, *models.File) error`
- `GetFilePath(ctx, path) (*models.FilePath, error)`
- `InsertFilePath(ctx, *models.FilePath) error`
- `UpdateFilePathDeletedAt(ctx, path, deletedAt) error`
- `GetFilePathsBySourceAndActive(ctx, sourceID, isActive) ([]*models.FilePath, error)`
- `GetFilePathsBySourceAndInactiveNotDeleted(ctx, sourceID) ([]*models.FilePath, error)`
- `GetSourceProcessing(ctx, sourceID) (*models.SourceProcessing, error)`
- `UpsertSourceProcessing(ctx, *models.SourceProcessing) error`
- `UpdateSourceProcessingStatus(ctx, sourceID, updates map[string]interface{}) error`

Implement PostgreSQL version using `pgx` in `internal/repository/postgres/postgres.go`. Use environment variable `DB_TYPE=postgres` and `POSTGRES_DSN` for connection.

#### 3. Database Schema (PostgreSQL)
Provide the following table definitions in `migrations/000001_init.up.sql`:

```sql
-- Table: files
CREATE TABLE files (
    hash VARCHAR(64) PRIMARY KEY,
    size BIGINT NOT NULL,
    mime_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: file_paths
CREATE TABLE file_paths (
    path TEXT PRIMARY KEY,
    source_id TEXT NOT NULL,
    hash VARCHAR(64) NOT NULL REFERENCES files(hash),
    is_active BOOLEAN NOT NULL,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_file_paths_hash ON file_paths(hash);
CREATE INDEX idx_file_paths_source_active ON file_paths(source_id, is_active);

-- Table: source_processing
CREATE TABLE source_processing (
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
);
```

#### 4. Models (`internal/models/models.go`)
Define Go structs corresponding to the tables above. Use appropriate types (`string`, `int64`, `time.Time`, `bool`). Include JSON tags if needed.

#### 5. Yandex Disk API Client (`internal/disk/disk.go`)
Implement client with methods:
- `ListFiles(sourceFolder, offset, limit) ([]Resource, error)` – calls `GET /resources/files` with pagination.
- `DeleteFile(path) error` – `DELETE /resources?path=...&permanently=true`.
- `GetFolderContents(path) ([]Resource, error)` – `GET /resources?path=...` to check folder emptiness.

Use OAuth token from env `YANDEX_DISK_TOKEN`. Respect rate limits (e.g., add delays). Use `net/http`.

#### 6. Configuration (`internal/config/config.go`)
Load env vars:
- `DB_TYPE` (default "postgres")
- `POSTGRES_DSN`
- `YANDEX_DISK_TOKEN`
- `LOG_LEVEL` (optional)

#### 7. ODS Functions (cmd/*)
Each executable:
- Loads config, initializes repository and disk client.
- Implements logic:

**`scan_source`** (takes `source_id` and optional `resume_offset` via command-line flags)
- Updates `source_processing` record.
- Iterates over files using `ListFiles` with pagination.
- For each file, gets MD5 from API, checks existence in `files` table.
- If not exists → inserts into `files` and into `file_paths` with `is_active=true`.
- If exists → inserts into `file_paths` with `is_active=false`.
- Updates `current_offset` in `source_processing` after each page.
- Sets status `completed` when done.

**`cleanup_duplicates`** (takes `source_id`)
- Selects `file_paths` where `source_id = ?` and `is_active = false` and `deleted_at IS NULL`.
- For each, calls `DeleteFile` via API, then updates `deleted_at`.
- Updates `source_processing` counters.

**`cleanup_folders`** (takes `source_id`)
- Recursively finds empty folders (using `GetFolderContents`).
- Deletes empty folders via API.
- Updates `source_processing` counters.

All functions must log progress and errors. Use `log.Printf`.

#### 8. Supporting Files
- `.gitignore`: exclude binaries, .env, terraform state.
- `README.md`: setup instructions (clone, install Go, run PostgreSQL with Docker, apply migrations, set env, run commands).
- `.env.example`: template for env variables.
- `go.mod`: module initialized with `github.com/yourusername/big_data_file_archive_processor` (replace with actual username). Use Go 1.21.

#### 9. Migration Tool
Use `golang-migrate/migrate` CLI. Provide instructions in README to run:
```bash
migrate -path migrations -database "$POSTGRES_DSN" up
```

---

### Deliverables
Generate all files described above with full code. Ensure the code compiles and follows Go best practices. Provide clear comments. The agent should propose the implementation details; we will review and test locally.

If anything is unclear, ask for clarification before coding.

