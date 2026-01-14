# Stationeers Server Mod Updater

Tools to automate updating Stationeers server mods via SteamCMD.

## Overview

This repository provides two implementations for updating Stationeers Workshop mods:

1. **Rust CLI Tool** (`rust/`): A command-line utility for direct server mod management
2. **Go Library** (`go/`): A library for integration into other Go projects (e.g., Stationeers Server UI)

Both implementations provide the same core functionality: parsing `modconfig.xml` files, downloading mods via SteamCMD, and relocating them to the appropriate directories using hardlinks (with fallback to copy).

## Features

- **Automatic mod detection**: Parses `modconfig.xml` to identify installed Workshop mods
- **Selective updates**: Update all mods or specific mods by ID
- **Ignore disabled mods**: Skip mods that are disabled in the configuration (enabled by default)
- **Efficient file operations**: Uses hard links when possible, falls back to copying
- **Clear error reporting**: Detailed error messages with cause chains
- **SteamCMD integration**: Automatically finds SteamCMD in PATH or accepts a custom path

## Implementations

### Rust CLI Tool (`rust/`)

A standalone command-line utility for managing server mods.

**Documentation**: See [rust/README.md](rust/README.md) for detailed Rust CLI documentation

**Quick Start**:
```bash
cd rust
cargo build --release
./target/release/st-update update --config-location /path/to/modconfig.xml
```

### Go Library (`go/`)

A library package for integrating mod update functionality into Go applications.

**Documentation**: See [go/README.md](go/README.md) for API reference and examples

**Quick Start**:
```go
import modupdater "github.com/dbidwell94/stationeers_server_mods_update/go"

opts := modupdater.UpdateOptions{
    ConfigLocation: "/path/to/modconfig.xml",
    IgnoreDisabled: true,
}
result, err := modupdater.UpdateMods(opts)
```

## Prerequisites

- **SteamCMD**: Must be installed and accessible. Download from [Valve's SteamCMD wiki](https://developer.valvesoftware.com/wiki/SteamCMD)

For the Rust CLI:
- **Rust toolchain** (for building from source): Install from [rustup.rs](https://rustup.rs/)

For the Go library:
- **Go**: 1.21 or later recommended

## How It Works

### 1. Configuration Parsing

The tool parses the `modconfig.xml` file to extract:
- Workshop mod IDs from path values (e.g., `/app/mods/Workshop_3576112002`)
- Enabled/disabled status of each mod
- Installation paths

### 2. Mod Filtering

Based on the provided arguments:
- Filters out disabled mods if `--ignore-disabled` is `true` (default)
- Filters to specific mod IDs if `--mod-id` arguments are provided

### 3. SteamCMD Execution

Constructs and executes a SteamCMD command:
```bash
steamcmd +login anonymous \
  +workshop_download_item 544550 <MOD_ID_1> validate \
  +workshop_download_item 544550 <MOD_ID_2> validate \
  ... \
  +logoff \
  +quit
```

Where `544550` is the Stationeers app ID.

### 4. Download Verification

- Parses SteamCMD output using regex to verify successful downloads
- Pattern: `Downloaded item (\d+) to "([^"]*)"`
- Reports success for each downloaded mod
- Fails if any requested mod was not successfully downloaded

### 5. File Relocation

For each downloaded mod:
1. Creates the destination directory if it doesn't exist
2. Attempts to create hard links (for efficiency)
3. Falls back to copying files if hard linking fails
4. Recursively processes subdirectories

## Error Handling

The tool provides detailed error messages for various failure scenarios:

### Common Errors

**`SteamCMD is missing`**
- **Cause**: SteamCMD binary not found in PATH or specified location
- **Solution**: Install SteamCMD or provide the correct path via `--steam-cmd-path`

**`The specified ModID was not downloaded: <ID>`**
- **Cause**: SteamCMD failed to download a specific mod
- **Possible reasons**:
  - Mod doesn't exist or was removed from Workshop
  - Network connectivity issues
  - SteamCMD authentication problems
- **Solution**: Verify the mod ID exists on Steam Workshop, check network connection

**`An error occurred with steamcmd: Non-0 exit status: <CODE>`**
- **Cause**: SteamCMD exited with an error
- **Solution**: Run SteamCMD manually to diagnose the issue

**`Could not open config file at <PATH>`**
- **Cause**: The specified `modconfig.xml` file doesn't exist or is inaccessible
- **Solution**: Verify the path and file permissions

**Parse errors**
- **Cause**: Invalid XML in `modconfig.xml`
- **Solution**: Validate the XML structure and fix any syntax errors

**File operation errors**
- **Cause**: Permission issues, disk full, or I/O errors during file copying
- **Solution**: Check disk space and file/directory permissions

## Repository Structure

```
stationeers_server_mods_update/
├── rust/                # Rust CLI implementation
│   ├── src/
│   │   ├── main.rs          # CLI argument parsing and entry point
│   │   ├── update.rs        # Core update logic and SteamCMD integration
│   │   ├── modconfig.rs     # XML parsing for modconfig.xml
│   │   └── utils/
│   │       ├── mod.rs       # Utilities module
│   │       └── fs.rs        # File system operations (copying/linking)
│   └── Cargo.toml           # Project dependencies and metadata
├── go/                  # Go library implementation
│   ├── modupdater.go        # Main library interface
│   ├── modconfig.go         # XML parsing
│   ├── steamcmd.go          # SteamCMD execution
│   ├── fileops.go           # File operations
│   ├── modupdater_test.go   # Tests
│   ├── go.mod               # Go module file
│   ├── README.md            # Go library documentation
│   └── examples/            # Usage examples
└── modconfig.xml        # Example configuration file
```

## Configuration File Format

The `modconfig.xml` file follows the Stationeers server format:

```xml
<?xml version="1.0" encoding="utf-8"?>
<ModConfig xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
           xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <Core Enabled="true">
    <Path />
  </Core>
  <Local Enabled="true">
    <Path Value="/app/mods/Workshop_3576112002" />
  </Local>
  <Local Enabled="false">
    <Path Value="/app/mods/Workshop_3575689739" />
  </Local>
</ModConfig>
```

- `Enabled` attribute determines if a mod is active
- `Path Value` must contain the pattern `Workshop_<ID>` where `<ID>` is the Steam Workshop item ID

## Technical Details

### Dependencies

- **clap**: Command-line argument parsing with derive macros and environment variable support
- **tokio**: Async runtime for process execution
- **quick-xml**: XML parsing and deserialization
- **serde**: Serialization/deserialization framework
- **regex**: Pattern matching for Workshop IDs and SteamCMD output
- **which**: Binary location detection
- **colored**: Terminal color output
- **anyhow**: Flexible error handling with context
- **thiserror**: Custom error type derivation

### Performance Optimizations

- **Hard linking**: Attempts to create hard links instead of copying files to save disk space and time
- **Fallback mechanism**: Automatically falls back to file copying if hard linking is not supported by the filesystem
- **Async process execution**: Uses tokio for non-blocking SteamCMD execution

## Limitations and Known Issues

1. **Backup functionality**: The `--backup-updated` flag is accepted but not fully implemented in the current version
2. **Steam authentication**: Only supports anonymous login (sufficient for public Workshop items)
3. **Windows support**: File system operations may behave differently on Windows (hard linking limitations)
4. **Error recovery**: If a single mod fails to download, the entire operation fails

## Contributing

Contributions are welcome! Please ensure:

For Rust code:
- Code follows Rust formatting standards (`cargo fmt`)
- No new warnings are introduced (`cargo clippy`)

For Go code:
- Code follows Go formatting standards (`go fmt`)
- Passes `go vet` checks
- Includes tests for new functionality

## License

Please refer to the repository license file for licensing information.

## Support

For issues, feature requests, or questions, please use the GitHub issue tracker at:
https://github.com/dbidwell94/stationeers_server_mods_update/issues
