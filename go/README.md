# Stationeers Mod Updater - Go Library

A Go library for updating Stationeers server mods via SteamCMD. This package provides a simple function-based API that can be integrated into other Go projects.

## Overview

This library provides functionality to:
- Parse `modconfig.xml` files
- Extract Workshop mod IDs
- Execute SteamCMD to download mods
- Verify successful downloads
- Relocate mods using hardlinks (with fallback to copy)

## Installation

```bash
go get github.com/dbidwell94/stationeers_server_mods_update/go
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "log"
    
    modupdater "github.com/dbidwell94/stationeers_server_mods_update/go"
)

func main() {
    opts := modupdater.UpdateOptions{
        ConfigLocation: "/path/to/modconfig.xml",
        IgnoreDisabled: true,
    }
    
    result, err := modupdater.UpdateMods(opts)
    if err != nil {
        log.Fatalf("Failed to update mods: %v", err)
    }
    
    fmt.Printf("Successfully updated %d mods\n", len(result.UpdatedMods))
    for _, mod := range result.UpdatedMods {
        fmt.Printf("  Mod %d: %s\n", mod.WorkshopID, mod.InstallPath)
    }
}
```

### Update Specific Mods

```go
opts := modupdater.UpdateOptions{
    ConfigLocation: "/path/to/modconfig.xml",
    ModIDs:         []uint64{3576112002, 3575689739},
    IgnoreDisabled: false,
}

result, err := modupdater.UpdateMods(opts)
```

### Custom SteamCMD Path

```go
opts := modupdater.UpdateOptions{
    ConfigLocation: "/path/to/modconfig.xml",
    SteamCmdPath:   "/custom/path/to/steamcmd",
    IgnoreDisabled: true,
}

result, err := modupdater.UpdateMods(opts)
```

### Error Handling

```go
result, err := modupdater.UpdateMods(opts)
if err != nil {
    log.Printf("Update failed: %v", err)
    
    // Check partial results
    if result != nil && len(result.Errors) > 0 {
        fmt.Println("Individual errors:")
        for _, e := range result.Errors {
            fmt.Printf("  - %v\n", e)
        }
    }
}
```

## API Reference

### UpdateOptions

Configuration structure for updating mods.

```go
type UpdateOptions struct {
    // ConfigLocation is the path to the modconfig.xml file (required)
    ConfigLocation string
    
    // SteamCmdPath is the path to the steamcmd executable (optional)
    // If empty, the function will search for steamcmd in PATH
    SteamCmdPath string
    
    // ModIDs is a list of specific mod IDs to update (optional)
    // If empty, updates all enabled mods
    ModIDs []uint64
    
    // IgnoreDisabled determines whether to skip disabled mods
    // Default: true
    IgnoreDisabled bool
}
```

### UpdateMods

Main function for updating mods.

```go
func UpdateMods(opts UpdateOptions) (*UpdateResult, error)
```

**Parameters:**
- `opts`: Configuration options for the update operation

**Returns:**
- `*UpdateResult`: Information about updated mods and any errors
- `error`: Critical error if the operation failed completely

### UpdateResult

Result structure containing update information.

```go
type UpdateResult struct {
    // UpdatedMods is the list of mods that were successfully updated
    UpdatedMods []ModUpdateInfo
    
    // Errors contains any errors that occurred during the update
    Errors []error
}
```

### ModUpdateInfo

Information about a single mod update.

```go
type ModUpdateInfo struct {
    WorkshopID  uint64
    InstallPath string
    SourcePath  string
}
```

## Requirements

- **SteamCMD**: Must be installed and accessible
  - Either in system PATH
  - Or provide custom path via `UpdateOptions.SteamCmdPath`
- **Go**: 1.21 or later recommended

## How It Works

1. **Parse Configuration**: Reads and parses the `modconfig.xml` file
2. **Filter Mods**: Applies filters based on enabled/disabled status and specific mod IDs
3. **Download**: Executes SteamCMD to download the mods
4. **Verify**: Parses SteamCMD output to verify successful downloads
5. **Relocate**: Copies or hardlinks mod files to their destination directories

## File Operations

The library uses an efficient file relocation strategy:
1. Attempts to create hardlinks (saves disk space and is faster)
2. Falls back to file copying if hardlinking fails
3. Recursively processes subdirectories
4. Preserves file permissions

## Error Handling

The library provides detailed error information:
- Critical errors return immediately via the `error` return value
- Per-mod errors are collected in `UpdateResult.Errors`
- All errors include context about what failed and why

## Differences from Rust Version

This Go port maintains the same functionality as the Rust version but:
- Provides a library API instead of a CLI
- Uses standard Go patterns and idioms
- Returns structured results instead of printing to console
- Designed for integration into other Go projects (like Stationeers Server UI)

## License

Please refer to the repository license file for licensing information.
