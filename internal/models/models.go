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

// SourceProcessing represents source processing status tracking with full schema
type SourceProcessing struct {
	SourceID string `json:"source_id" db:"source_id"`

	// Scan phase fields
	ScanStartedAt   *time.Time `json:"scan_started_at,omitempty" db:"scan_started_at"`
	ScanCompletedAt *time.Time `json:"scan_completed_at,omitempty" db:"scan_completed_at"`
	CurrentOffset   *int64     `json:"current_offset,omitempty" db:"current_offset"`
	TotalFilesFound *int64     `json:"total_files_found,omitempty" db:"total_files_found"`
	ScanStatus      *string    `json:"scan_status,omitempty" db:"scan_status"`

	// Processing phase fields
	ProcessingStartedAt   *time.Time `json:"processing_started_at,omitempty" db:"processing_started_at"`
	ProcessingCompletedAt *time.Time `json:"processing_completed_at,omitempty" db:"processing_completed_at"`
	FilesProcessed        *int64     `json:"files_processed,omitempty" db:"files_processed"`
	FilesDeleted          *int64     `json:"files_deleted,omitempty" db:"files_deleted"`
	ProcessingStatus      *string    `json:"processing_status,omitempty" db:"processing_status"`

	// Cleanup phase fields
	CleanupStartedAt   *time.Time `json:"cleanup_started_at,omitempty" db:"cleanup_started_at"`
	CleanupCompletedAt *time.Time `json:"cleanup_completed_at,omitempty" db:"cleanup_completed_at"`
	FoldersDeleted     *int64     `json:"folders_deleted,omitempty" db:"folders_deleted"`
	CleanupStatus      *string    `json:"cleanup_status,omitempty" db:"cleanup_status"`

	// Common error field for all phases
	Error *string `json:"error,omitempty" db:"error"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewSourceProcessing creates a new SourceProcessing instance with initial timestamps
func NewSourceProcessing(sourceID string) *SourceProcessing {
	now := time.Now()
	return &SourceProcessing{
		SourceID:  sourceID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// StartScan marks the beginning of the scan phase
func (sp *SourceProcessing) StartScan() {
	now := time.Now()
	sp.ScanStartedAt = &now
	status := "scanning"
	sp.ScanStatus = &status
	sp.UpdatedAt = now
}

// CompleteScan marks the completion of the scan phase
func (sp *SourceProcessing) CompleteScan(totalFilesFound int64) {
	now := time.Now()
	sp.ScanCompletedAt = &now
	status := "completed"
	sp.ScanStatus = &status
	sp.TotalFilesFound = &totalFilesFound
	sp.UpdatedAt = now
}

// FailScan marks the scan phase as failed with an error
func (sp *SourceProcessing) FailScan(errorMsg string) {
	now := time.Now()
	status := "failed"
	sp.ScanStatus = &status
	sp.Error = &errorMsg
	sp.UpdatedAt = now
}

// StartProcessing marks the beginning of the processing phase
func (sp *SourceProcessing) StartProcessing() {
	now := time.Now()
	sp.ProcessingStartedAt = &now
	status := "processing"
	sp.ProcessingStatus = &status
	sp.UpdatedAt = now
}

// CompleteProcessing marks the completion of the processing phase
func (sp *SourceProcessing) CompleteProcessing(filesProcessed, filesDeleted int64) {
	now := time.Now()
	sp.ProcessingCompletedAt = &now
	status := "completed"
	sp.ProcessingStatus = &status
	sp.FilesProcessed = &filesProcessed
	sp.FilesDeleted = &filesDeleted
	sp.UpdatedAt = now
}

// FailProcessing marks the processing phase as failed with an error
func (sp *SourceProcessing) FailProcessing(errorMsg string) {
	now := time.Now()
	status := "failed"
	sp.ProcessingStatus = &status
	sp.Error = &errorMsg
	sp.UpdatedAt = now
}

// StartCleanup marks the beginning of the cleanup phase
func (sp *SourceProcessing) StartCleanup() {
	now := time.Now()
	sp.CleanupStartedAt = &now
	status := "cleaning"
	sp.CleanupStatus = &status
	sp.UpdatedAt = now
}

// CompleteCleanup marks the completion of the cleanup phase
func (sp *SourceProcessing) CompleteCleanup(foldersDeleted int64) {
	now := time.Now()
	sp.CleanupCompletedAt = &now
	status := "completed"
	sp.CleanupStatus = &status
	sp.FoldersDeleted = &foldersDeleted
	sp.UpdatedAt = now
}

// FailCleanup marks the cleanup phase as failed with an error
func (sp *SourceProcessing) FailCleanup(errorMsg string) {
	now := time.Now()
	status := "failed"
	sp.CleanupStatus = &status
	sp.Error = &errorMsg
	sp.UpdatedAt = now
}

// UpdateCurrentOffset updates the current offset during scanning
func (sp *SourceProcessing) UpdateCurrentOffset(offset int64) {
	sp.CurrentOffset = &offset
	sp.UpdatedAt = time.Now()
}

// UpdateProcessingStatus updates the processing status and sets the updated timestamp
func (sp *SourceProcessing) UpdateProcessingStatus(status string) {
	sp.ProcessingStatus = &status
	sp.UpdatedAt = time.Now()
}

// RecordSuccess updates the source processing record with successful processing details
func (sp *SourceProcessing) RecordSuccess(filesProcessed, filesDeleted int64) {
	now := time.Now()
	sp.FilesProcessed = &filesProcessed
	sp.FilesDeleted = &filesDeleted
	sp.UpdatedAt = now
}

// RecordError updates the source processing record with error details
func (sp *SourceProcessing) RecordError(errorMsg string) {
	now := time.Now()
	sp.Error = &errorMsg
	sp.UpdatedAt = now
}
