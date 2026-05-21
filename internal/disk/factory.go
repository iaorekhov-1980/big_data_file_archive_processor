package disk

import (
	"fmt"
	"time"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/config"
)

// NewDiskClient creates a new DiskClient based on the provided configuration.
// Currently only Yandex Disk is supported as a cloud storage provider.
// The function is designed to be extensible for future providers by checking
// configuration fields that would indicate which provider to use.
func NewDiskClient(cfg *config.Config) (DiskClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	token := cfg.GetYandexDiskToken()
	if token == "" {
		return nil, fmt.Errorf("Yandex Disk token is required")
	}

	baseURL := cfg.GetYandexDiskBaseURL()
	timeout := time.Duration(cfg.GetYandexDiskTimeout()) * time.Second
	rateLimitDelay := time.Duration(cfg.GetYandexDiskRateLimitDelayMs()) * time.Millisecond

	client := NewYandexDiskClient(token,
		WithBaseURL(baseURL),
		WithTimeout(timeout),
		WithRateLimitDelay(rateLimitDelay),
	)

	return client, nil
}
