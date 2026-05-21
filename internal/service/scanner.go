package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/disk"
	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/models"
	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/repository"
)

// ScanService coordinates DiskClient + Repository to scan a Yandex Disk source folder.
type ScanService struct {
	diskClient disk.DiskClient
	repo       repository.Repository
	txManager  *TxManager
	pageSize   int
	logger     *slog.Logger
}

// NewScanService creates a new ScanService instance.
func NewScanService(
	diskClient disk.DiskClient,
	repo repository.Repository,
	txManager *TxManager,
	pageSize int,
	logger *slog.Logger,
) *ScanService {
	return &ScanService{
		diskClient: diskClient,
		repo:       repo,
		txManager:  txManager,
		pageSize:   pageSize,
		logger:     logger,
	}
}

// ScanResult summarizes the scan operation outcome.
type ScanResult struct {
	SourceID        string
	TotalFilesFound int64
	FilesInserted   int64
	FilesSkipped    int64
	Errors          []string
}

// Scan performs a full recursive scan of the given source folder on Yandex Disk.
func (s *ScanService) Scan(ctx context.Context, sourceID, sourceFolder string, resumeOffset int) (*ScanResult, error) {
	logger := s.logger.With("source_id", sourceID)
	logger.Info("starting scan", "source_folder", sourceFolder, "resume_offset", resumeOffset)

	result := &ScanResult{
		SourceID: sourceID,
	}

	// 1. Load or create SourceProcessing record
	sp, err := s.loadOrCreateSourceProcessing(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to load/create source processing: %w", err)
	}

	// 2. If scan already completed, return immediately (idempotent)
	if sp.ScanStatus != nil && *sp.ScanStatus == "completed" {
		totalFound := int64Value(sp.TotalFilesFound)
		logger.Info("scan already completed, skipping",
			"total_files", totalFound)
		result.TotalFilesFound = totalFound
		return result, nil
	}

	// Mark scan as started
	if sp.ScanStatus == nil || *sp.ScanStatus != "scanning" {
		sp.StartScan()
		if err := s.repo.UpsertSourceProcessing(ctx, sp); err != nil {
			return nil, fmt.Errorf("failed to start scan: %w", err)
		}
	}

	// 3. Recursively scan the folder tree
	retryCfg := DefaultRetryConfig()

	err = s.scanFolder(ctx, sourceID, sourceFolder, sourceFolder, result, sp, retryCfg)
	if err != nil {
		return result, err
	}

	// 4. Mark scan as completed
	sp.CompleteScan(result.TotalFilesFound)
	if err := s.repo.UpsertSourceProcessing(ctx, sp); err != nil {
		return nil, fmt.Errorf("failed to complete scan: %w", err)
	}

	logger.Info("scan completed",
		"total_files", result.TotalFilesFound,
		"inserted", result.FilesInserted,
		"skipped", result.FilesSkipped)

	return result, nil
}

// scanFolder recursively scans a single folder and all its subdirectories.
// sourceFolder is the root folder path used to strip prefixes from file paths.
func (s *ScanService) scanFolder(
	ctx context.Context,
	sourceID string,
	sourceFolder string,
	folderPath string,
	result *ScanResult,
	sp *models.SourceProcessing,
	retryCfg RetryConfig,
) error {
	logger := s.logger.With("source_id", sourceID, "folder", folderPath)

	// Fetch folder contents with retry
	var items []disk.Resource
	err := DoWithRetry(ctx, retryCfg, func() error {
		var innerErr error
		items, innerErr = s.diskClient.ListFiles(ctx, folderPath, 0, s.pageSize)
		return innerErr
	})
	if err != nil {
		errMsg := fmt.Sprintf("failed to list folder %s: %v", folderPath, err)
		logger.Error("list folder error", "error", err)
		result.Errors = append(result.Errors, errMsg)
		return errors.New(errMsg)
	}

	// Separate files and subdirectories
	var files []disk.Resource
	var subDirs []string

	for _, item := range items {
		if item.Type == "dir" {
			subDirs = append(subDirs, item.Path)
		} else {
			files = append(files, item)
		}
	}

	// Process files in this folder within a transaction
	if len(files) > 0 {
		pageInserted, pageSkipped, err := s.processFiles(ctx, sourceID, sourceFolder, files, sp)
		if err != nil {
			return fmt.Errorf("failed to process files in %s: %w", folderPath, err)
		}
		result.FilesInserted += pageInserted
		result.FilesSkipped += pageSkipped
		result.TotalFilesFound += pageInserted + pageSkipped

		logger.Info("folder processed",
			"files", len(files),
			"inserted", pageInserted,
			"skipped", pageSkipped)
	}

	// Recursively scan subdirectories
	for _, subDir := range subDirs {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("scan cancelled: %w", err)
		}

		if err := s.scanFolder(ctx, sourceID, sourceFolder, subDir, result, sp, retryCfg); err != nil {
			errMsg := fmt.Sprintf("failed to scan subfolder %s: %v", subDir, err)
			logger.Error("subfolder scan error", "subfolder", subDir, "error", err)
			result.Errors = append(result.Errors, errMsg)
			continue
		}
	}

	return nil
}

// processFiles processes a batch of files within a single transaction.
// sourceFolder is the root folder path; it is stripped from file paths to store relative paths.
func (s *ScanService) processFiles(
	ctx context.Context,
	sourceID string,
	sourceFolder string,
	files []disk.Resource,
	sp *models.SourceProcessing,
) (int64, int64, error) {
	var pageInserted, pageSkipped int64

	err := s.txManager.WithTransaction(ctx, func(txCtx context.Context, tx *TxBundle) error {
		for _, file := range files {
			hash := file.MD5
			if hash == "" {
				s.logger.Warn("file has no MD5 hash, skipping", "path", file.Path)
				pageSkipped++
				continue
			}

			// Strip source folder prefix from path to store relative path
			relativePath := stripPrefix(file.Path, sourceFolder)

			existingFP, fpErr := tx.FilePathRepo.GetFilePathByPathAndSource(txCtx, relativePath, sourceID)
			if fpErr == nil && existingFP != nil {
				s.logger.Debug("file path already exists, skipping", "path", relativePath)
				pageSkipped++
				continue
			}

			_, fileErr := tx.FileRepo.GetFileByHash(txCtx, hash)
			if fileErr != nil {
				newFile := models.NewFile(hash, file.Name, file.Size, file.MimeType)
				if insErr := tx.FileRepo.InsertFile(txCtx, newFile); insErr != nil {
					var dupErr *repository.DuplicateError
					if !errorsAs(insErr, &dupErr) {
						return fmt.Errorf("failed to insert file %s: %w", hash, insErr)
					}
				}
			}

			newFilePath := models.NewFilePath(relativePath, sourceID, hash, true)
			if insErr := tx.FilePathRepo.InsertFilePath(txCtx, newFilePath); insErr != nil {
				return fmt.Errorf("failed to insert file path %s: %w", relativePath, insErr)
			}
			pageInserted++
		}

		if err := tx.SourceProcessingRepo.UpsertSourceProcessing(txCtx, sp); err != nil {
			return fmt.Errorf("failed to update source processing: %w", err)
		}

		return nil
	})

	return pageInserted, pageSkipped, err
}

// stripPrefix removes the sourceFolder prefix from a full Yandex Disk path.
// E.g. "disk:/ods/folder/subdir/file.txt" with sourceFolder "/ods/folder" -> "subdir/file.txt"
func stripPrefix(fullPath, sourceFolder string) string {
	// Remove "disk:" prefix if present
	path := fullPath
	if len(path) > 5 && path[:5] == "disk:" {
		path = path[5:]
	}

	// Remove sourceFolder prefix
	if len(path) > len(sourceFolder) && path[:len(sourceFolder)] == sourceFolder {
		path = path[len(sourceFolder):]
	}

	// Strip leading slash
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	return path
}

// loadOrCreateSourceProcessing loads an existing SourceProcessing record or creates a new one.
func (s *ScanService) loadOrCreateSourceProcessing(ctx context.Context, sourceID string) (*models.SourceProcessing, error) {
	sp, err := s.repo.GetSourceProcessing(ctx, sourceID)
	if err != nil {
		var notFound *repository.NotFoundError
		if errorsAs(err, &notFound) {
			sp = models.NewSourceProcessing(sourceID)
			if err := s.repo.UpsertSourceProcessing(ctx, sp); err != nil {
				return nil, fmt.Errorf("failed to create source processing: %w", err)
			}
			return sp, nil
		}
		return nil, err
	}
	return sp, nil
}

// markScanFailed updates the source processing record with a failed status.
func (s *ScanService) markScanFailed(ctx context.Context, sourceID, errorMsg string) {
	sp, err := s.repo.GetSourceProcessing(ctx, sourceID)
	if err != nil {
		s.logger.Error("failed to load source processing for error marking",
			"source_id", sourceID, "error", err)
		return
	}
	sp.FailScan(errorMsg)
	if err := s.repo.UpsertSourceProcessing(ctx, sp); err != nil {
		s.logger.Error("failed to mark scan as failed",
			"source_id", sourceID, "error", err)
	}
}

// int64Value safely dereferences an *int64 pointer.
func int64Value(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// errorsAs is a wrapper for errors.As to avoid import cycle issues in tests.
func errorsAs(err error, target interface{}) bool {
	return errors.As(err, target)
}
