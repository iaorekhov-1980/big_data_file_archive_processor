-- Migration: 000001_init.down.sql
-- Drops the database schema for the file archive processor

-- Drop indexes first (in reverse order of creation)
DROP INDEX IF EXISTS idx_source_processing_cleanup_status;
DROP INDEX IF EXISTS idx_source_processing_processing_status;
DROP INDEX IF EXISTS idx_source_processing_scan_status;
DROP INDEX IF EXISTS idx_file_paths_source_active;
DROP INDEX IF EXISTS idx_file_paths_hash;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS source_processing;
DROP TABLE IF EXISTS file_paths;
DROP TABLE IF EXISTS files;