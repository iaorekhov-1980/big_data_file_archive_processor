package main

import (
	"fmt"
	"os"

	"github.com/iaorekhov-1980/big_data_file_archive_processor/internal/testutils"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "setup":
		if len(os.Args) > 2 && os.Args[2] == "--help" {
			testutils.PrintSetupHelp()
			return
		}
		if err := testutils.SetupTestConfig(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "test-connection":
		if len(os.Args) > 2 && os.Args[2] == "--help" {
			testutils.PrintConnectionHelp()
			return
		}
		if err := testutils.TestDBConnection(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Test Utilities for Database Integration")
	fmt.Println()
	fmt.Println("Usage: go run cmd/test/main.go <command> [--help]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  setup           - Create test_config.toml from test_config.example.toml")
	fmt.Println("  test-connection - Test PostgreSQL connection and verify schema")
	fmt.Println("  help            - Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/test/main.go setup")
	fmt.Println("  go run cmd/test/main.go test-connection")
	fmt.Println("  go run cmd/test/main.go setup --help")
}
