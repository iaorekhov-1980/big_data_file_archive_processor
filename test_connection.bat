@echo off
REM Test database connection
REM Usage: test_connection.bat

setlocal

echo Testing PostgreSQL connection...
echo Configuration will be loaded from test_config.toml
echo.

echo Testing Go database connection with test_config.toml...
echo.

REM Run the Go test connection program


go run cmd/test/main.go test-connection
if %ERRORLEVEL% EQU 0 (
    echo.
    echo SUCCESS: Database connection and schema are ready for testing!
) else (
    echo.
    echo ERROR: Database connection test failed
)

endlocal