package models

import (
	"time"
)

// File represents a file entity in the database
type File struct {
	Hash      string    `json:"hash" db:"hash"`
	Name      string    `json:"name" db:"name"`
	Size      int64     `json:"size" db:"size"`
	MimeType  string    `json:"mime_type" db:"mime_type"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewFile creates a new File instance with current timestamps
func NewFile(hash, name string, size int64, mimeType string) *File {
	now := time.Now()
	return &File{
		Hash:      hash,
		Name:      name,
		Size:      size,
		MimeType:  mimeType,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// FilePath represents a file path entity in the database
type FilePath struct {
	Path      string     `json:"path" db:"path"`
	SourceID  string     `json:"source_id" db:"source_id"`
	Hash      string     `json:"hash" db:"hash"`
	IsActive  bool       `json:"is_active" db:"is_active"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// NewFilePath creates a new FilePath instance with current timestamps
func NewFilePath(path, sourceID, hash string, isActive bool) *FilePath {
	now := time.Now()
	return &FilePath{
		Path:      path,
		SourceID:  sourceID,
		Hash:      hash,
		IsActive:  isActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// MarkDeleted marks the file path as deleted with the current time
func (fp *FilePath) MarkDeleted() {
	now := time.Now()
	fp.DeletedAt = &now
	fp.IsActive = false
	fp.UpdatedAt = now
}

// SourceProcessing represents source processing status tracking
type SourceProcessing struct {
	SourceID            string     `json:"source_id" db:"source_id"`
	LastProcessedPath   *string    `json:"last_processed_path,omitempty" db:"last_processed_path"`
	LastProcessedHash   *string    `json:"last_processed_hash,omitempty" db:"last_processed_hash"`
	LastProcessedAt     *time.Time `json:"last_processed_at,omitempty" db:"last_processed_at"`
	ProcessingStatus    string     `json:"processing_status" db:"processing_status"`
	TotalFilesProcessed int64      `json:"total_files_processed" db:"total_files_processed"`
	TotalBytesProcessed int64      `json:"total_bytes_processed" db:"total_bytes_processed"`
	LastError           *string    `json:"last_error,omitempty" db:"last_error"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// NewSourceProcessing creates a new SourceProcessing instance with initial status
func NewSourceProcessing(sourceID string) *SourceProcessing {
	now := time.Now()
	return &SourceProcessing{
		SourceID:         sourceID,
		ProcessingStatus: "pending",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// UpdateProcessingStatus updates the processing status and sets the updated timestamp
func (sp *SourceProcessing) UpdateProcessingStatus(status string) {
	sp.ProcessingStatus = status
	sp.UpdatedAt = time.Now()
}

// RecordSuccess updates the source processing record with successful processing details
func (sp *SourceProcessing) RecordSuccess(path, hash string, fileSize int64) {
	now := time.Now()
	sp.LastProcessedPath = &path
	sp.LastProcessedHash = &hash
	sp.LastProcessedAt = &now
	sp.TotalFilesProcessed++
	sp.TotalBytesProcessed += fileSize
	sp.LastError = nil
	sp.UpdatedAt = now
}

// RecordError updates the source processing record with error details
func (sp *SourceProcessing) RecordError(errMsg string) {
	sp.LastError = &errMsg
	sp.UpdatedAt = time.Now()
}
