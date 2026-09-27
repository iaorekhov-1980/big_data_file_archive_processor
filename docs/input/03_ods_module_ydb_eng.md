**Prompt for ODS Module Development (Scanning, Duplicate Cleanup, Empty Folder Removal)**

---

### **1. Project Overview**
We are building a system to manage a personal file archive stored on **Yandex Disk** (2 TB). The current task is to implement the **ODS (raw data) layer processing module**.  

**Goal:** For each new source (a folder manually uploaded to `/ods/<source_name>`), we need to:
- Scan all files in the source folder, retrieving metadata and MD5 hash via Yandex Disk API.
- Identify **global duplicates** (by hash) across all previously processed sources and mark them in the database.
- Later, **physically delete** duplicate files from the source folder.
- **Remove empty folders** left after deletion.

All operations are triggered **manually step by step** via HTTP-triggered Cloud Functions (Yandex Cloud Functions).  
Metadata is stored in **Yandex Database (YDB)**.  
Code is written in **Go**, infrastructure is described with **Terraform**.

---

### **2. YDB Table Schema**

#### **Table `files`** (global unique file catalog)
| Field        | Type      | Description                              |
|--------------|-----------|------------------------------------------|
| `hash`       | Utf8      | MD5 hash (as returned by Yandex Disk API), primary key |
| `size`       | Uint64    | File size in bytes                       |
| `mime_type`  | Utf8      | MIME type                                 |
| `created_at` | Timestamp | When this record was first created        |
| `updated_at` | Timestamp | Last update time                          |

#### **Table `file_paths`** (all file paths with status)
| Field        | Type      | Description                                                                 |
|--------------|-----------|-----------------------------------------------------------------------------|
| `path`       | Utf8      | Full ODS path (e.g., `/ods/2025-03-20_phone/photo.jpg`), primary key      |
| `source_id`  | Utf8      | Source folder name (e.g., `2025-03-20_phone`)                              |
| `hash`       | Utf8      | Reference to `files.hash`                                                   |
| `is_active`  | Bool      | `true` = file physically exists on disk and is unique (not a duplicate)    |
| `deleted_at` | Timestamp?| When the file was physically deleted (if `is_active=false` and deleted)    |
| `created_at` | Timestamp | When this path record was created                                           |
| `updated_at` | Timestamp | Last update                                                                 |

**Index:** `hash_idx` GLOBAL ON (`hash`) – for fast lookup by hash.

#### **Table `source_processing`** (monitoring source processing)
| Field                   | Type       | Description                                  |
|-------------------------|------------|----------------------------------------------|
| `source_id`             | Utf8       | Source folder name, primary key              |
| `scan_started_at`       | Timestamp? | When scanning started                         |
| `scan_completed_at`     | Timestamp? | When scanning completed                       |
| `current_offset`        | Uint64?    | For resuming interrupted scanning             |
| `total_files_found`     | Uint64?    | Total number of files in the source           |
| `scan_status`           | Utf8?      | `in_progress`, `completed`, `failed`          |
| `scan_error`            | Utf8?      | Error details if any                          |
| `processing_started_at` | Timestamp? | When duplicate cleanup started                 |
| `processing_completed_at`| Timestamp?| When duplicate cleanup completed              |
| `files_processed`       | Uint64?    | Number of duplicate files processed            |
| `files_deleted`         | Uint64?    | Number of files actually deleted               |
| `processing_status`     | Utf8?      | `in_progress`, `completed`, `failed`          |
| `processing_error`      | Utf8?      | Error details if any                          |
| `cleanup_started_at`    | Timestamp? | When empty folder cleanup started              |
| `cleanup_completed_at`  | Timestamp? | When empty folder cleanup completed            |
| `folders_deleted`       | Uint64?    | Number of empty folders deleted                |
| `cleanup_status`        | Utf8?      | `in_progress`, `completed`, `failed`          |
| `cleanup_error`         | Utf8?      | Error details if any                          |
| `created_at`            | Timestamp  | Record creation time                           |
| `updated_at`            | Timestamp  | Last update                                    |

---

### **3. Yandex Disk API Interaction**

- **Base URL:** `https://cloud-api.yandex.net/v1/disk/`
- **Authorization:** OAuth token (store in Yandex Lockbox; function retrieves it via SDK).
- **List files in a folder (flat list, recursive):**
  ```
  GET /resources/files?path=/ods/{source_id}&limit=100&offset={offset}
  ```
  Response contains `items` array with fields: `path`, `md5`, `size`, `mime_type`, `created`, `modified`.  
  Pagination: while `items` length equals `limit`, there are more pages.
- **Delete a file:**
  ```
  DELETE /resources?path={path}&permanently=true
  ```
  Success: `204 No Content`.
- **Check folder contents (for empty folder cleanup):**
  ```
  GET /resources?path={folder_path}
  ```
  Response `_embedded.items` lists files and subfolders.

**Important:** Respect API rate limits (max 40 requests per second). Add delays between requests.

---

### **4. Functions to Implement**

#### **4.1 `scan_source` (HTTP trigger)**

- **Input (JSON):**
  ```json
  {
    "source_id": "2025-03-20_phone",
    "resume_offset": 0   // optional, to resume interrupted scan
  }
  ```
- **Logic:**
  1. Check `source_processing` for given `source_id`. If not exists, create record with `scan_status='in_progress'`, `scan_started_at=now()`, `current_offset=resume_offset`. If exists and status `completed`, either return error or reset (depending on requirement; here resetting allowed).
  2. Loop until last page:
     - Call Yandex Disk API to fetch files.
     - For each file:
       - Check `files` table for hash (`SELECT hash FROM files WHERE hash = $md5`).
       - If hash **not found**:
         - Insert into `files` (hash, size, mime_type, created_at=now()).
         - Insert into `file_paths` (path, source_id, hash, is_active=true, created_at=now()).
       - If hash **found**:
         - Insert into `file_paths` (path, source_id, hash, is_active=false, created_at=now()).
         - (Do not modify `files`.)
     - After processing the page, update `source_processing`: increment `current_offset` by number of files processed, set `updated_at`.
  3. When received items count < `limit`, scanning completed:
     - Set `scan_completed_at=now()`, `scan_status='completed'`, `total_files_found = current_offset`.
  4. On error (network, API, timeout), record `scan_error`, set `scan_status='failed'`, and exit. Offset preserved for resumption.
- **Idempotency:** Use `UPSERT` with conflict handling or check existence before insert to avoid duplicate key errors.

#### **4.2 `cleanup_duplicates` (HTTP trigger)**

- **Input:**
  ```json
  { "source_id": "2025-03-20_phone" }
  ```
- **Precondition:** Source must have `scan_status = 'completed'`; else return error.
- **Logic:**
  1. Update `source_processing`: `processing_started_at=now()`, `processing_status='in_progress'`, reset counters.
  2. Select from `file_paths` all records for this source where `is_active=false` and `deleted_at IS NULL` (i.e., marked as duplicates not yet deleted).
  3. For each record:
     - Call Yandex Disk API to delete the file at `path`.
     - On success (`204`), update `file_paths`: set `deleted_at=now()`, `updated_at=now()`.
     - On error (e.g., 404 – file already gone), consider it deleted, update `deleted_at` (log warning).
     - Increment `files_processed` and `files_deleted` (latter only on success or 404).
     - Add a small delay (e.g., 50-100 ms) between calls to avoid rate limits.
  4. After processing all records, set `processing_completed_at=now()`, `processing_status='completed'`, `files_processed`, `files_deleted`.
  5. On error (e.g., 500), record `processing_error`, set status `failed`, and abort. On retry, function will reprocess records with `is_active=false` and `deleted_at IS NULL` (idempotent).

#### **4.3 `cleanup_folders` (HTTP trigger)**

- **Input:**
  ```json
  { "source_id": "2025-03-20_phone" }
  ```
- **Precondition:** (optional) duplicate cleanup completed, but not strictly required.
- **Logic:**
  1. Update `source_processing`: `cleanup_started_at=now()`, `cleanup_status='in_progress'`.
  2. Obtain list of all folders inside the source folder. Approach:
     - Extract all directory paths from `file_paths` for this source (by parsing each `path` and collecting all parent prefixes). This gives folders that ever contained files.
     - Sort folders by depth descending (deepest first).
     - For each folder, call API `GET /resources?path={folder_path}` and inspect `_embedded.items`. If empty (no files and no subfolders), the folder is empty.
  3. For each empty folder:
     - Call API `DELETE /resources?path={folder_path}&permanently=true`.
     - Increment `folders_deleted`.
     - After deletion, check parent folder recursively (may become empty now) and delete if empty.
  4. Update `cleanup_completed_at`, `cleanup_status='completed'`, `folders_deleted`.

---

### **5. Implementation Requirements (Go)**

- **Language:** Go 1.21+.
- **Project Structure:**
  ```
  /cmd
    /scan_source          // main for scan_source function
    /cleanup_duplicates   // main for cleanup_duplicates
    /cleanup_folders      // main for cleanup_folders
  /internal
    /ydb                  // YDB client and queries
    /disk                 // Yandex Disk API client
    /models               // data structures
    /config               // environment config loader
  /terraform              // .tf files
  ```
- **Error Handling:** Log errors with `log.Printf`. For transient API errors (5xx), retry with exponential backoff (max 3 attempts). Use context for timeouts.
- **Configuration:** Environment variables: `YDB_ENDPOINT`, `YDB_DATABASE`, `LOCKBOX_SECRET_ID` (containing OAuth token), `IAM_TOKEN_URL` (for metadata token). Functions will retrieve secrets via Yandex Cloud SDK.
- **Logging:** Write to stdout (captured by Yandex Cloud Logging). Include start/end of processing, counts, and errors.
- **Timeouts:** Cloud Functions have max 10 minutes. Assume source size fits within that; if not, later we can add chunking, but for now ignore.

---

### **6. Terraform Artifacts**

Create Terraform configuration to deploy:
- Three Cloud Functions (`scan_source`, `cleanup_duplicates`, `cleanup_folders`) with HTTP triggers.
- YDB database and tables (as defined).
- Service account with roles:
  - `ydb.editor` – for YDB access.
  - `lockbox.payloadViewer` – to read OAuth token from Lockbox.
  - `functions.functionInvoker` – if needed for cross-invocation.
- Lockbox secret to store the OAuth token (value to be filled manually after creation).
- Environment variables for functions (e.g., `YDB_ENDPOINT`, `YDB_DATABASE`, `LOCKBOX_SECRET_ID`).

Also provide output URLs for each function.

---

### **7. Expected Output**

The agent should deliver:
1. **Complete Go source code** for all three functions, including internal packages (disk client, YDB client, models, config).
2. **Terraform files** (`.tf`) that deploy the infrastructure.
3. **Brief instructions** on deployment (how to apply Terraform, how to obtain OAuth token, how to invoke functions).

Code must be well-structured, commented (English or Russian), with error handling and logging.

---

### **8. Additional Notes**

- Do not generate code now; this prompt is for context. When I say "start", you may begin writing code and Terraform manifests.
- Adhere strictly to the described architecture. If any details are ambiguous, ask clarifying questions before coding.
- Use idiomatic Go, follow best practices (e.g., separate concerns, avoid global state, use interfaces for testability).

---

**This prompt is ready to be passed to the agent in VS Code (Continue + DeepSeek).**

