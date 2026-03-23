# Big Data File Archive Processor - ODS Module

A Go-based system for processing file archives on Yandex Disk with PostgreSQL metadata storage.

## Overview

This system provides three main functions:
1. **Scan Source**: Scan Yandex Disk sources, identify duplicates globally
2. **Cleanup Duplicates**: Physically delete duplicate files from Yandex Disk
3. **Cleanup Folders**: Remove empty folders from Yandex Disk

## Architecture

- **Database Abstraction**: PostgreSQL with interface-based design for easy migration to YDB
- **Yandex Disk API**: REST client with rate limiting
- **Modular Commands**: Separate executables for each function

## Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15+
- Yandex Disk account with OAuth token

## Quick Start

### 1. Clone and Setup

```bash
git clone https://github.com/iaorekhov-1980/big_data_file_archive_processor.git
cd big_data_file_archive_processor
```

### 2. Environment Configuration

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your values
# - Set POSTGRES_DSN for your database
# - Set YANDEX_DISK_TOKEN from Yandex OAuth
```

### 3. Database Setup

```bash
# Start PostgreSQL with Docker
docker-compose up -d postgres

# Run migrations
migrate -path migrations -database "$POSTGRES_DSN" up
```

### 4. Build and Run

```bash
# Install dependencies
go mod tidy

# Build all commands
go build ./cmd/scan_source
go build ./cmd/cleanup_duplicates
go build ./cmd/cleanup_folders

# Run scan source
./scan_source --source-id "23_03_2026_my_source"

# Run cleanup duplicates
./cleanup_duplicates --source-id "23_03_2026_my_source"

# Run cleanup folders
./cleanup_folders --source-id "23_03_2026_my_source"
```

## Project Structure

```
big_data_file_archive_processor/
├── cmd/                    # Command executables
│   ├── scan_source/       # Source scanning
│   ├── cleanup_duplicates/# Duplicate cleanup
│   └── cleanup_folders/   # Empty folder cleanup
├── internal/              # Internal packages
│   ├── models/           # Data models
│   ├── repository/       # Database abstraction
│   ├── disk/            # Yandex Disk API client
│   └── config/          # Configuration
├── migrations/           # Database migrations
├── terraform/           # Infrastructure as Code
└── docs/                # Documentation
```

## Database Schema

The system uses three main tables:
- `files`: Global file metadata (hash, size, mime_type)
- `file_paths`: File locations with source tracking
- `source_processing`: Processing status and progress

## Configuration

See `.env.example` for all available environment variables.

## Development

### Adding Dependencies

```bash
go get github.com/jackc/pgx/v5
```

### Running Tests

```bash
go test ./...
```

### Code Style

- Use `go fmt` for formatting
- Follow Go standard library conventions
- Add comments for exported functions and types

## License

[Add your license here]

## Contributing

[Add contribution guidelines]