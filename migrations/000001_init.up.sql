-- Migration: 000001_init.up.sql
-- Creates the initial database schema for the file archive processor

-- Create files table
CREATE TABLE files (
    hash VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    size BIGINT NOT NULL,
    mime_type VARCHAR(127) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create file_paths table
CREATE TABLE file_paths (
    path VARCHAR(1024) NOT NULL,
    source_id VARCHAR(255) NOT NULL,
    hash VARCHAR(64) NOT NULL REFERENCES files(hash) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (path, source_id)
);

-- Create source_processing table
CREATE TABLE source_processing (
    source_id VARCHAR(255) PRIMARY KEY,
    
    -- Scan phase fields
    scan_started_at TIMESTAMPTZ,
    scan_completed_at TIMESTAMPTZ,
    current_offset BIGINT,
    total_files_found BIGINT,
    scan_status VARCHAR(31),
    
    -- Processing phase fields
    processing_started_at TIMESTAMPTZ,
    processing_completed_at TIMESTAMPTZ,
    files_processed BIGINT,
    files_deleted BIGINT,
    processing_status VARCHAR(31),
    
    -- Cleanup phase fields
    cleanup_started_at TIMESTAMPTZ,
    cleanup_completed_at TIMESTAMPTZ,
    folders_deleted BIGINT,
    cleanup_status VARCHAR(31),
    
    -- Common error field for all phases
    error VARCHAR(511),
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_file_paths_hash ON file_paths(hash);
CREATE INDEX idx_file_paths_source_active ON file_paths(source_id, is_active);

-- Create indexes for source_processing table
CREATE INDEX idx_source_processing_scan_status ON source_processing(scan_status);
CREATE INDEX idx_source_processing_processing_status ON source_processing(processing_status);
CREATE INDEX idx_source_processing_cleanup_status ON source_processing(cleanup_status);