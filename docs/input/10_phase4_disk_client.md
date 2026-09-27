## Updated Prompt for Phase 4: Yandex Disk Client Implementation

**Title:** Phase 4: Yandex Disk API Client Implementation

**Context:**
Phase 3 (Database Layer) is complete with:
- Full database schema with composite primary key `(path, source_id)` for `file_paths` ✓
- PostgreSQL repository implementation ✓
- Comprehensive test infrastructure ✓
- Repository factory pattern ✓
- Repository method names updated to reflect composite key structure ✓

**Clarifications Received:**
1. ✅ **Composite Primary Key**: Confirmed - `file_paths` uses `(path, source_id)` as composite PK
2. ✅ **Method Names**: Updated to reflect composite key (e.g., `GetFilePathByPathAndSource`)
3. ✅ **Phase Focus**: Phase 4 = Yandex Disk client only
4. ✅ **Phase 5**: Business logic layer (services)
5. ✅ **Phase 6**: Main command implementations

**Current Project Structure Status:**
```
big_data_file_archive_processor/
├── cmd/
│   └── test/              # Test command (exists)
│       └── main.go
├── internal/
│   ├── config/           # Configuration (exists)
│   ├── models/           # Data models (exists)
│   ├── repository/       # Database abstraction (exists)
│   │   ├── factory/      # Repository factory (exists)
│   │   └── postgres/     # PostgreSQL implementation (exists)
│   └── testutils/        # Test utilities (exists)
├── migrations/           # Database migrations (exists)
└── (missing: internal/disk/, cmd/* main commands)
```

**Phase 4 Tasks: Yandex Disk Client Implementation**

**Step 4.1: Create Disk Client Interface**
- Create `internal/disk/interface.go`
- Define `DiskClient` interface with methods:
  - `ListFiles(ctx context.Context, sourceFolder string, offset, limit int) ([]Resource, error)`
  - `DeleteFile(ctx context.Context, path string) error`
  - `GetFolderContents(ctx context.Context, path string) ([]Resource, error)`
  - `GetFileInfo(ctx context.Context, path string) (*Resource, error)` (optional helper)
- Define `Resource` struct matching Yandex Disk API response

**Step 4.2: Implement Yandex Disk Client**
- Create `internal/disk/yandexdisk.go`
- Implement `YandexDiskClient` struct with:
  - HTTP client with timeout
  - OAuth token
  - Base URL (`https://cloud-api.yandex.net/v1/disk`)
  - Rate limiting mechanism
- Implement all interface methods using Yandex Disk REST API:
  - `ListFiles`: Calls `GET /resources/files` with pagination
  - `DeleteFile`: Calls `DELETE /resources` with `permanently=true`
  - `GetFolderContents`: Calls `GET /resources` for folder listing
- Add proper error handling for API responses

**Step 4.3: Add Rate Limiting**
- Implement configurable delay between requests (default 200ms)
- Use `time.Sleep` or more sophisticated rate limiter
- Respect Yandex Disk API rate limits

**Step 4.4: Create Configuration Integration**
- Update `internal/config/config.go` to include disk client settings:
  - `YandexDiskToken` (already exists)
  - `YandexDiskBaseURL` (default: `https://cloud-api.yandex.net/v1/disk`)
  - `YandexDiskTimeout` (default: 30 seconds)
  - `YandexDiskRateLimitDelayMs` (default: 200)
- Add validation for token presence

**Step 4.5: Add Client Factory**
- Create `internal/disk/factory.go`
- Factory function: `NewDiskClient(config *config.Config) (DiskClient, error)`
- Currently only Yandex Disk implementation
- Prepare for potential future cloud storage providers

**Step 4.6: Create Unit Tests**
- Create `internal/disk/yandexdisk_test.go`
- Mock HTTP responses for testing
- Test error scenarios
- Test rate limiting behavior

**Step 4.7: Update Documentation**
- Update README.md with Yandex Disk setup instructions
- Add API token acquisition guide
- Document rate limits and quotas

**Technical Specifications:**

**Yandex Disk API Endpoints:**
- List files: `GET /resources/files?offset={offset}&limit={limit}`
- Delete file: `DELETE /resources?path={path}&permanently=true`
- Get folder: `GET /resources?path={path}&limit={limit}`

**Resource Structure (based on Yandex Disk API):**
```go
type Resource struct {
    Path        string    `json:"path"`
    Name        string    `json:"name"`
    Type        string    `json:"type"` // "file" or "dir"
    Size        int64     `json:"size,omitempty"`
    MD5         string    `json:"md5,omitempty"`
    SHA256      string    `json:"sha256,omitempty"`
    Modified    time.Time `json:"modified,omitempty"`
    Created     time.Time `json:"created,omitempty"`
    MimeType    string    `json:"mime_type,omitempty"`
    MediaType   string    `json:"media_type,omitempty"`
}
```

**Error Handling:**
- Map HTTP status codes to appropriate errors
- Handle token expiration/authentication errors
- Handle rate limit exceeded (429) responses
- Handle network timeouts

**Dependencies to Add:**
- Standard library: `net/http`, `time`, `context`, `encoding/json`
- Consider adding `github.com/stretchr/testify` for testing (already in go.mod)

**Requirements:**
- Follow Go best practices and conventions
- Add comprehensive comments
- Include proper error handling and logging
- Use context for cancellation and timeouts
- Implement connection pooling for HTTP client
- Commit after each step
- Push to GitHub after each step

**Git Commit Strategy:**
- Step 4.1: "Phase 4.1: Create disk client interface"
- Step 4.2: "Phase 4.2: Implement Yandex Disk client"
- Step 4.3: "Phase 4.3: Add rate limiting"
- Step 4.4: "Phase 4.4: Update configuration"
- Step 4.5: "Phase 4.5: Create client factory"
- Step 4.6: "Phase 4.6: Add unit tests"
- Step 4.7: "Phase 4.7: Update documentation"

**Ready to proceed with Step 4.1.**

