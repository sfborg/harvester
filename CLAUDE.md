# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Harvester is a Go CLI tool that converts biodiversity datasets from various idiosyncratic formats into SFGA (Species File Group Archive) format. It downloads, extracts, and transforms taxonomic data from external sources like GRIN, ION, IOC Birds, PaleoDB, and World Plants.

## Commands

This project uses [just](https://github.com/casey/just) as a command runner
(modern alternative to make). Run `just --list` to see all available recipes.

### Building and Installing
```bash
just build         # Build binary with debug info (development)
just build-release # Build release binary (no debug info, hardcoded version)
just install       # Build and install to ~/go/bin
just clean         # Remove generated files
```

### Testing
```bash
just test        # Run all tests
go test ./...    # Run tests directly with go

# Run a single test
go test -run TestFunctionName ./path/to/package
```

### Development Tools
```bash
just fmt         # Format all Go code
just lint        # Run linter (requires golangci-lint)
just tidy        # Tidy dependencies
just verify      # Format, tidy, test, and build (full verification)
```

### Release
```bash
just release     # Build releases for all platforms (Linux, Mac, Windows)
```

### Application Usage
```bash
harvester list              # List all supported datasets
harvester list -v           # List with detailed information
harvester get <label>       # Convert dataset to SFGA format
harvester get <label> -s    # Skip download step (use cached data)
harvester get <label> -z    # Output compressed zip file
```

## Architecture

### Core Components

**Main Entry Point** (`main.go` → `cmd/root.go`)
- Uses Cobra for CLI interface
- Sets up structured logging with `tint` handler
- Commands: `list` and `get`

**Harvester Interface** (`pkg/harvester.go`, `pkg/interface.go`)
- `Harvester` interface provides two methods:
  - `List()` - returns map of all registered data sources
  - `Get(label, outPath)` - converts a specific dataset to SFGA format
- Core workflow: Download → Extract → InitSfga → ToSfga → Export

**Data Convertor Interface** (`pkg/data/interface.go`)
- All data sources implement the `Convertor` interface
- Split into `Accessor` (metadata) and `Processor` (actions)
- Key methods each source must implement:
  - `Download()` - fetch data from URL or local file
  - `Extract()` - unpack archives
  - `InitSfga()` - create SFGA archive structure
  - `ToSfga()` - transform data into SFGA format

**Base Implementation** (`internal/base/base.go`)
- Provides default implementations for common Convertor methods
- Handles automatic archive format detection (zip, tar.gz, tar.bz2, tar.xz)
- Integrates gnparser for scientific name parsing
- Sources can embed this and override specific methods

**Source Registration** (`internal/list/list.go`)
- `GetDataSets()` function returns map of all available sources
- Each source is registered with a unique label (map key)
- Currently supports: grin, ion, ioc birds, paleodb, wfwp

### Adding a New Data Source

1. Create package in `internal/sources/<source>/`
2. Implement `data.Convertor` interface (typically by embedding `base.Convertor`)
3. Create `New()` function that returns `data.Convertor`
4. Implement source-specific logic:
   - Override `Extract()` if special handling needed
   - Implement `ToSfga()` for data transformation (required)
   - Create helper files (e.g., `sfga.go`, `meta.go`, `name_usage.go`) as patterns
5. Register in `internal/list/list.go`

### Configuration

**Config Struct** (`pkg/config/config.go`)
- Manages cache directories (download, extract, sfga)
- Controls behavior: skip download, zip output, verbose mode
- CSV/TSV parsing options: delimiters, quotes, bad row handling
- Archive metadata: issued date, version
- Uses functional options pattern (`Option` type)

### Key Dependencies

- **sflib** - Core SFGA library for Species File Group Archive format
- **gnparser** - Scientific name parsing (handles nomenclatural codes)
- **gnsys** - System utilities (download, archive extraction)
- **cobra** - CLI framework
- **modernc.org/sqlite** - Embedded SQLite database (used by SFGA)

### File Organization Pattern

Sources in `internal/sources/<name>/` typically follow this structure:
- `<name>.go` - Main implementation, `New()` constructor
- `sfga.go` - SFGA-specific conversion logic
- `meta.go` - Metadata handling
- `name_usage.go` - Name/taxon processing

## Resources

- [SFGA schema](https://github.com/sfborg/sfga/blob/main/schema.sql)

## Important Notes

- Version is embedded at build time via ldflags from git tags
- Cache directory: `~/.cache/sfborg/harvester/` (or OS temp dir)
- CGO is disabled (`CGO_ENABLED=0`) for static compilation
- Tests use `-shuffle=on` to catch order dependencies
- Some sources require manual steps (check `ManualSteps` field)

## Code Style

- Follow standard Go formatting (enforced by `go fmt`)
- Code should unless absolutely necessary fit into 80 columns
- Use golangci-lint for linting
- Run `go mod tidy` before finalizing tasks to avoid lint warnings
- Comments use full sentences with periods
- Exported types/functions have doc comments

## Human Development Oriented Style

- Functions implemented with human reader in mind 
- Code must be easy to understand
- Documentation aims to be concise and clear
