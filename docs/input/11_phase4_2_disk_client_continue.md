Here's the prompt for continuing Phase 4 in a new chat:

---

**Phase 4: Yandex Disk API Client Implementation — Remaining Steps**

**Completed so far:**
- Step 4.1: `internal/disk/interface.go` — `DiskClient` interface with `Resource` struct
- Step 4.2a: `internal/disk/yandexdisk.go` — Transport, HTTP client, error handling (`DiskError`), rate limiting, `buildGetRequest` helper
- Step 4.2b: `internal/disk/operations.go` — Interface methods: `ListFiles`, `GetFolderContents`, `GetFileInfo` (no `DeleteFile` yet)
- Connectivity tests (`connectivity_test.go`) and operations tests (`operations_test.go`) — all passing with real Yandex Disk API
- Config integration: `testutils/config.go` updated with `YandexDisk` section (Token, BaseURL, TestFolder)
- Committed and pushed to GitHub

**Remaining steps:**

**Step 4.4: Update Configuration**
- Update `internal/config/config.go` to add:
  - `YandexDiskBaseURL` (default: `https://cloud-api.yandex.net/v1/disk`)
  - `YandexDiskTimeout` (default: 30 seconds)
  - `YandexDiskRateLimitDelayMs` (default: 200) — already exists as `RateLimitDelayMs`
- Add validation for new fields

**Step 4.5: Create Client Factory**
- Create `internal/disk/factory.go`
- Factory function: `NewDiskClient(config *config.Config) (DiskClient, error)`
- Currently only Yandex Disk implementation
- Prepare for potential future cloud storage providers

**Step 4.6: Add Unit Tests (mock-based)**
- Create `internal/disk/yandexdisk_test.go`
- Mock HTTP responses for testing
- Test error scenarios without real API calls
- Test rate limiting behavior

**Step 4.7: Update Documentation**
- Update README.md with Yandex Disk setup instructions
- Add API token acquisition guide
- Document rate limits and quotas

**Current project structure:**
```
internal/
├── config/config.go          # Needs update (Step 4.4)
├── disk/
│   ├── interface.go          # Done
│   ├── yandexdisk.go         # Done (transport, errors, rate limiting)
│   ├── operations.go         # Done (ListFiles, GetFolderContents, GetFileInfo)
│   ├── connectivity_test.go  # Done (real API tests)
│   ├── operations_test.go    # Done (real API tests)
│   └── factory.go            # Missing (Step 4.5)
├── models/models.go          # Done
├── repository/               # Done
├── testutils/config.go       # Updated with YandexDisk section
```

**Key details:**
- Yandex Disk API base URL: `https://cloud-api.yandex.net/v1/disk`
- Auth: `OAuth` token in `Authorization` header
- Rate limiting: configurable delay between requests (default 200ms)
- `DiskError` type with `IsNotFound()`, `IsAuthError()`, `IsRateLimited()` helpers
- `buildGetRequest(path, queryParams)` helper exists for testing
- `doRequest(req, responseBody)` handles auth, JSON decode, error mapping
- Test config uses TOML with `[yandex_disk]` section (token, base_url, test_folder)
- `getTestToken(t)`, `getTestBaseURL(t)`, `getTestFolder(t)` helpers available in tests


