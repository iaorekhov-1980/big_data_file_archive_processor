package testutils

import (
	"fmt"
)

// SetupTestConfig creates test_config.toml from test_config.example.toml
func SetupTestConfig() error {
	if err := CreateTestConfig(); err != nil {
		return fmt.Errorf("failed to create test config: %w", err)
	}
	return nil
}

// PrintSetupHelp prints usage information for setup command
func PrintSetupHelp() {
	fmt.Println("Usage: go run cmd/testutils/main.go setup")
	fmt.Println("Creates test_config.toml from test_config.example.toml")
}
