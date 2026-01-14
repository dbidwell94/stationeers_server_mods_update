# Stationeers Server Mod Updater

A command-line tool to automate updating Stationeers server mods via SteamCMD.

## Overview

`st-update` is a Rust-based utility that simplifies the process of updating Stationeers Workshop mods on a server. It reads a `modconfig.xml` file, downloads updated mods using SteamCMD, and relocates them to the appropriate directories.

## Features

- **Automatic mod detection**: Parses `modconfig.xml` to identify installed Workshop mods
- **Selective updates**: Update all mods or specific mods by ID
- **Ignore disabled mods**: Skip mods that are disabled in the configuration (enabled by default)
- **Efficient file operations**: Uses hard links when possible, falls back to copying
- **Clear error reporting**: Detailed error messages with cause chains
- **SteamCMD integration**: Automatically finds SteamCMD in PATH or accepts a custom path

## Installation

### Prerequisites

- **SteamCMD**: Must be installed and accessible. Download from [Valve's SteamCMD wiki](https://developer.valvesoftware.com/wiki/SteamCMD)
- **Rust toolchain** (for building from source): Install from [rustup.rs](https://rustup.rs/)

### Building from Source

```bash
git clone https://github.com/dbidwell94/stationeers_server_mods_update.git
cd stationeers_server_mods_update
cargo build --release
```

The binary will be available at `target/release/st-update`.

## Usage

### Basic Command Structure

```bash
st-update update --config-location <PATH_TO_MODCONFIG_XML> [OPTIONS]
```

### Required Arguments

- `-c, --config-location <CONFIG_LOCATION>`: Path to the `modconfig.xml` file
  - Can also be set via the `CONFIG_LOCATION` environment variable

### Optional Arguments

- `-m, --mod-id <MOD_ID>`: Specific Workshop mod ID(s) to update
  - Can be specified multiple times to update multiple specific mods
  - If omitted, updates all mods in the config
  
- `-d, --ignore-disabled`: Skip updating disabled mods (default: `true`)
  - Set to `false` to update disabled mods as well

- `-b, --backup-updated`: Backup mods before updating (default: `false`)
  - **Note**: Currently implemented as a parameter but not fully utilized in the codebase

- `-s, --steam-cmd-path <STEAM_CMD_PATH>`: Custom path to SteamCMD binary
  - Can also be set via the `STEAM_CMD_PATH` environment variable
  - If omitted, the tool searches for `steamcmd` in your system's PATH

### Examples

#### Update all enabled mods

```bash
st-update update --config-location /path/to/modconfig.xml
```

#### Update specific mods by Workshop ID

```bash
st-update update --config-location /path/to/modconfig.xml \
  --mod-id 3576112002 --mod-id 3575689739
```

#### Update all mods (including disabled)

```bash
st-update update --config-location /path/to/modconfig.xml --ignore-disabled false
```

#### Use custom SteamCMD path

```bash
st-update update --config-location /path/to/modconfig.xml \
  --steam-cmd-path /custom/path/to/steamcmd
```

#### Using environment variables

```bash
export CONFIG_LOCATION=/path/to/modconfig.xml
export STEAM_CMD_PATH=/custom/path/to/steamcmd
st-update update
```

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

## File Structure

```
stationeers_server_mods_update/
├── src/
│   ├── main.rs          # CLI argument parsing and entry point
│   ├── update.rs        # Core update logic and SteamCMD integration
│   ├── modconfig.rs     # XML parsing for modconfig.xml
│   └── utils/
│       ├── mod.rs       # Utilities module
│       └── fs.rs        # File system operations (copying/linking)
├── Cargo.toml           # Project dependencies and metadata
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
- Code follows Rust formatting standards (`cargo fmt`)
- No new warnings are introduced (`cargo clippy`)
- Existing functionality is not broken

## License

Please refer to the repository license file for licensing information.

## Support

For issues, feature requests, or questions, please use the GitHub issue tracker at:
https://github.com/dbidwell94/stationeers_server_mods_update/issues
