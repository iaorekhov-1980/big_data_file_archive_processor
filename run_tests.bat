@echo off
REM Test runner for PostgreSQL repository tests
REM Usage: run_tests.bat [test_name] [--setup] [--help]

setlocal

REM Check for setup flag
if "%1"=="--setup" (
    echo Setting up test configuration...
    go run cmd/test/main.go setup
    goto :end
)

if "%1"=="--help" (
    echo PostgreSQL Repository Test Runner
    echo ==================================
    echo.
    echo Usage: run_tests.bat [options]
    echo.
    echo Options:
    echo   test_name          Run specific test
    echo   --setup            Setup test configuration
    echo   --help             Show this help message
    echo.
    echo Examples:
    echo   run_tests.bat                     Run all tests
    echo   run_tests.bat TestFileRepository_Integration
    echo   run_tests.bat --setup             Setup test configuration
    goto :end
)

echo Running PostgreSQL repository tests...
echo Configuration will be loaded from test_config.toml
echo.

REM Check if test name is provided
if "%1"=="" (
    echo Running all database tests...
    set TEST_CMD=go test ./internal/repository/postgres -v
) else (
    echo Running test: %1
    set TEST_CMD=go test ./internal/repository/postgres -v -run %1
)

REM Run the test
echo Executing: %TEST_CMD%
echo.
%TEST_CMD%

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Tests passed successfully!
) else (
    echo.
    echo Tests failed with error code: %ERRORLEVEL%
)

:end
endlocal