# Test runner for PostgreSQL repository tests
# Usage: .\run_tests.ps1 [-TestName "TestName"] [-Short] [-Coverage] [-Setup]

param(
    [string]$TestName,
    [switch]$Short,
    [switch]$Coverage,
    [switch]$Setup,
    [switch]$Help
)

if ($Help) {
    Write-Host "PostgreSQL Repository Test Runner"
    Write-Host "=================================="
    Write-Host ""
    Write-Host "Usage: .\run_tests.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -TestName <name>    Run specific test (e.g., 'TestFileRepository_Integration')"
    Write-Host "  -Short              Run only unit tests (skip integration tests)"
    Write-Host "  -Coverage           Generate code coverage report"
    Write-Host "  -Setup              Setup test configuration"
    Write-Host "  -Help               Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\run_tests.ps1                          # Run all tests"
    Write-Host "  .\run_tests.ps1 -Short                   # Run only unit tests"
    Write-Host "  .\run_tests.ps1 -TestName TestFileRepository_Integration"
    Write-Host "  .\run_tests.ps1 -Coverage                # Run with coverage"
    Write-Host "  .\run_tests.ps1 -Setup                   # Setup test configuration"
    Write-Host ""
    exit 0
}

if ($Setup) {
    Write-Host "Setting up test configuration..." -ForegroundColor Cyan
    go run cmd/test/main.go setup
    exit $LASTEXITCODE
}

Write-Host "Running PostgreSQL repository tests..." -ForegroundColor Cyan
Write-Host "Configuration will be loaded from test_config.toml"
Write-Host ""

# Build test command
$testCmd = "go test ./internal/repository/postgres"

if ($Short) {
    $testCmd += " -short"
    Write-Host "Running in short mode (skipping integration tests)" -ForegroundColor Yellow
}

if ($TestName) {
    $testCmd += " -run `"$TestName`""
    Write-Host "Running specific test: $TestName" -ForegroundColor Yellow
}

if ($Coverage) {
    $testCmd += " -cover -coverprofile=coverage.out"
    Write-Host "Generating coverage report" -ForegroundColor Yellow
}

$testCmd += " -v"

Write-Host "Executing: $testCmd" -ForegroundColor Green
Write-Host ""

# Run the test
Invoke-Expression $testCmd

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Tests passed successfully!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "Tests failed with error code: $LASTEXITCODE" -ForegroundColor Red
}

# Generate coverage report if requested
if ($Coverage -and (Test-Path "coverage.out")) {
    Write-Host ""
    Write-Host "Generating coverage report..." -ForegroundColor Cyan
    go tool cover -html=coverage.out -o coverage.html
    if (Test-Path "coverage.html") {
        Write-Host "Coverage report generated: coverage.html" -ForegroundColor Green
    }
}

