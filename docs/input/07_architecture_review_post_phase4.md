## Complete Architecture Discussion Prompt for Next Plan Session

**Title:** Architecture Review & Planning Session - Post Phase 4 (Yandex Disk Client)

**Context:**
We are planning the next phases of the big_data_file_archive_processor project. Phase 4 (Yandex Disk Client Implementation) is currently being implemented in a separate agent session.

**Current Project Status Summary:**

### ✅ **Completed Phases:**

**Phase 1: Project Structure Setup** ✓
- Directory structure created
- Go module initialized (`github.com/iaorekhov-1980/big_data_file_archive_processor`)
- Supporting files: `.gitignore`, `README.md`, `.env.example`, `go.mod`

**Phase 2: Core Models and Interfaces** ✓
1. `internal/models/models.go` - Full schema with composite PK support
2. `internal/repository/interface.go` - Repository interface with composite key methods
3. `internal/config/config.go` - Configuration loader with validation

**Phase 3: Database Layer** ✓
1. Migration files with composite PK `(path, source_id)` for `file_paths`
2. PostgreSQL repository implementation (separate files per entity)
3. Repository factory pattern
4. Comprehensive test infrastructure with utilities

### 🔄 **Currently Implementing:**

**Phase 4: Yandex Disk Client Implementation** (in progress)
1. Disk client interface definition
2. Yandex Disk REST API implementation
3. Rate limiting and error handling
4. Configuration integration
5. Client factory and unit tests

### 📋 **Remaining Phases (To Be Planned):**

**Phase 5: Business Logic Layer**
- Service layer to coordinate repository and disk client
- Business rules and validation
- Workflow orchestration

**Phase 6: Main Command Implementations**
- `cmd/scan_source/main.go` - Source scanning
- `cmd/cleanup_duplicates/main.go` - Duplicate cleanup
- `cmd/cleanup_folders/main.go` - Empty folder cleanup

**Phase 7: Advanced Features & Optimization**
- Performance optimizations
- Advanced error recovery
- Monitoring and metrics
- Additional cloud storage providers

---

## Architecture Discussion Points for Next Session:

### 1. **Service Layer Design (Phase 5)**
**Questions:**
- Should we have separate services (`ScanService`, `CleanupService`) or a unified `ArchiveService`?
- How should services handle transaction management?
- What business rules need to be implemented?
- How to handle partial failures and retries?

**Proposed Structure:**
```
internal/service/
├── scanner.go      # ScanService: Source scanning logic
├── cleaner.go      # CleanupService: Duplicate and folder cleanup
├── orchestrator.go # Workflow orchestration
└── validator.go    # Business rule validation
```

### 2. **Command Layer Design (Phase 6)**
**Questions:**
- Command-line argument structure for each command?
- How to handle configuration loading in commands?
- Logging strategy for production use?
- Exit codes and error reporting?

**Proposed Command Structure:**
```bash
# Scan source
./scan_source --source-id "23_03_2026_my_source" [--resume-offset 0]

# Cleanup duplicates
./cleanup_duplicates --source-id "23_03_2026_my_source" [--dry-run]

# Cleanup folders  
./cleanup_folders --source-id "23_03_2026_my_source" [--recursive]
```

### 3. **Error Handling Strategy**
**Questions:**
- Global error handling approach?
- Retry logic for transient failures?
- Circuit breaker pattern for API calls?
- Error reporting and monitoring?

### 4. **Performance Considerations**
**Questions:**
- Batch operations for database inserts?
- Parallel processing for file operations?
- Memory management for large file lists?
- Database connection pooling optimization?

### 5. **Testing Strategy**
**Questions:**
- Integration test coverage for services?
- Mock strategies for Yandex Disk API?
- End-to-end testing approach?
- Performance/load testing needs?

### 6. **Deployment & Operations**
**Questions:**
- Docker containerization strategy?
- Configuration management in production?
- Log aggregation and monitoring?
- Health checks and readiness probes?

### 7. **Future Extensibility**
**Questions:**
- Adding new cloud storage providers?
- Supporting different database backends (YDB)?
- Plugin architecture for custom processors?
- Web API/REST interface potential?

---

## Technical Decisions Needed:

### **A. Service Layer Architecture**
**Option 1:** Domain-driven services (separate for scanning, cleanup)
**Option 2:** Unified service with internal specialization
**Option 3:** Pipeline/processor pattern with middleware

### **B. Transaction Management**
**Option 1:** Repository-level transactions
**Option 2:** Service-level transaction coordination  
**Option 3:** Saga pattern for distributed transactions

### **C. Configuration Management**
**Option 1:** Environment variables only
**Option 2:** Config files + environment variables
**Option 3:** Configuration service/API

### **D. Logging Strategy**
**Option 1:** Structured logging (JSON)
**Option 2:** Simple text logging with levels
**Option 3:** Contextual logging with request IDs

### **E. Monitoring & Observability**
**Option 1:** Basic logging only
**Option 2:** Metrics collection (Prometheus)
**Option 3:** Distributed tracing (OpenTelemetry)

---

## Preparation for Next Session:

**Please review:**
1. Current implementation of Phase 4 (Yandex Disk client)
2. The proposed architecture questions above
3. Any additional requirements or constraints

**Expected Deliverables from Next Session:**
1. Detailed design for Phase 5 (Business Logic Layer)
2. Implementation plan for Phase 6 (Main Commands)
3. Technical decisions on architecture questions
4. Updated project roadmap with timelines

**Ready to discuss architecture decisions once Phase 4 implementation is complete.**

