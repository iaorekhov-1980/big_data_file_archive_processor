# Makefile for PostgreSQL repository tests

# PostgreSQL connection details
PG_HOST ?= localhost
PG_PORT ?= 5432
PG_DB ?= ivordb
PG_USER ?= gouser
PG_PASS ?= gopwd

# Build DSN string
TEST_POSTGRES_DSN = postgres://$(PG_USER):$(PG_PASS)@$(PG_HOST):$(PG_PORT)/$(PG_DB)?sslmode=disable

.PHONY: help test test-all test-short test-file test-filepath test-sourceprocessing test-coverage test-connection setup-config clean

help:
	@echo "PostgreSQL Repository Test Runner"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  test-all           Run all database tests (default)"
	@echo "  test-short         Run only unit tests (skip integration)"
	@echo "  test-file          Run file repository tests"
	@echo "  test-filepath      Run file path repository tests"
	@echo "  test-sourceprocessing Run source processing repository tests"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  test-connection    Test database connection and schema"
	@echo "  setup-config       Setup test configuration"
	@echo "  clean              Clean up generated files"
	@echo ""
	@echo "Configuration:"
	@echo "  Create test_config.toml from test_config.example.toml"
	@echo "  or set TEST_POSTGRES_DSN environment variable"
	@echo ""
test-all: export TEST_POSTGRES_DSN = $(TEST_POSTGRES_DSN)
test-all:
	@echo "Running all database tests..."
	@go test ./internal/repository/postgres -v

test-short: export TEST_POSTGRES_DSN = $(TEST_POSTGRES_DSN)
test-short:
	@echo "Running unit tests (skipping integration)..."
	@go test ./internal/repository/postgres -short -v

test-file: export TEST_POSTGRES_DSN = $(TEST_POSTGRES_DSN)
test-file:
	@echo "Running file repository tests..."
	@go test ./internal/repository/postgres -v -run TestFileRepository_

test-filepath: export TEST_POSTGRES_DSN = $(TEST_POSTGRES_DSN)
test-filepath:
	@echo "Running file path repository tests..."
	@go test ./internal/repository/postgres -v -run TestFilePathRepository_

test-sourceprocessing: export TEST_POSTGRES_DSN = $(TEST_POSTGRES_DSN)
test-sourceprocessing:
	@echo "Running source processing repository tests..."
	@go test ./internal/repository/postgres -v -run TestSourceProcessingRepository_

test-coverage: export TEST_POSTGRES_DSN = $(TEST_POSTGRES_DSN)
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./internal/repository/postgres -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-connection: export TEST_POSTGRES_DSN = $(TEST_POSTGRES_DSN)
test-connection:
	@echo "Testing database connection and schema..."
	@go run internal/testutils/test_db_connection.go

setup-config:
	@echo "Setting up test configuration..."
	@go run internal/testutils/setup_test_config.go

clean:
	@rm -f coverage.out coverage.html 2>/dev/null || true
	@echo "Cleaned up generated files"

# Default target
.DEFAULT_GOAL := test-all