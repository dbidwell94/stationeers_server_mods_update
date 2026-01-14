# Quick Start Guide - Go Library

## Installation

In your Go project, add the module:

```bash
go get github.com/dbidwell94/stationeers_server_mods_update/go
```

## Basic Usage

```go
package main

import (
    "fmt"
    "log"
    
    modupdater "github.com/dbidwell94/stationeers_server_mods_update/go"
)

func main() {
    // Configure update options
    opts := modupdater.UpdateOptions{
        ConfigLocation: "/path/to/modconfig.xml",
        IgnoreDisabled: true,  // Skip disabled mods
    }
    
    // Update mods
    result, err := modupdater.UpdateMods(opts)
    if err != nil {
        log.Fatalf("Failed to update mods: %v", err)
    }
    
    // Display results
    fmt.Printf("Successfully updated %d mods\n", len(result.UpdatedMods))
    for _, mod := range result.UpdatedMods {
        fmt.Printf("  ✓ Mod %d: %s\n", mod.WorkshopID, mod.InstallPath)
    }
}
```

## Advanced Usage

### Update Specific Mods Only

```go
opts := modupdater.UpdateOptions{
    ConfigLocation: "/path/to/modconfig.xml",
    ModIDs:         []uint64{3576112002, 3575689739}, // Specific mods
    IgnoreDisabled: false, // Update even if disabled in config
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

### Handle Partial Failures

```go
result, err := modupdater.UpdateMods(opts)
if err != nil {
    // Some mods may have succeeded even if there was an error
    if result != nil {
        fmt.Printf("Partially updated %d mods\n", len(result.UpdatedMods))
        
        // Log individual errors
        for _, e := range result.Errors {
            log.Printf("Error: %v", e)
        }
    }
    return err
}
```

## Integration Example (Web Server)

```go
package main

import (
    "encoding/json"
    "net/http"
    
    modupdater "github.com/dbidwell94/stationeers_server_mods_update/go"
)

type UpdateRequest struct {
    ConfigPath string   `json:"config_path"`
    ModIDs     []uint64 `json:"mod_ids,omitempty"`
}

func handleUpdateMods(w http.ResponseWriter, r *http.Request) {
    var req UpdateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    opts := modupdater.UpdateOptions{
        ConfigLocation: req.ConfigPath,
        ModIDs:         req.ModIDs,
        IgnoreDisabled: len(req.ModIDs) == 0, // Ignore disabled if updating all
    }
    
    result, err := modupdater.UpdateMods(opts)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "updated_count": len(result.UpdatedMods),
        "mods":          result.UpdatedMods,
    })
}
```

## API Reference

### UpdateOptions

| Field | Type | Description |
|-------|------|-------------|
| `ConfigLocation` | `string` | **Required**. Path to modconfig.xml file |
| `SteamCmdPath` | `string` | Optional. Custom steamcmd path (uses PATH if empty) |
| `ModIDs` | `[]uint64` | Optional. Specific mod IDs to update (updates all enabled if empty) |
| `IgnoreDisabled` | `bool` | Whether to skip disabled mods (default: true) |

### UpdateResult

| Field | Type | Description |
|-------|------|-------------|
| `UpdatedMods` | `[]ModUpdateInfo` | List of successfully updated mods |
| `Errors` | `[]error` | Any errors that occurred during update |

### ModUpdateInfo

| Field | Type | Description |
|-------|------|-------------|
| `WorkshopID` | `uint64` | Steam Workshop mod ID |
| `InstallPath` | `string` | Where the mod was installed |
| `SourcePath` | `string` | Where steamcmd downloaded the mod |

## Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `steamcmd not found` | SteamCMD not in PATH | Install SteamCMD or provide custom path |
| `could not read config file` | Config file doesn't exist | Check the ConfigLocation path |
| `failed to parse config XML` | Invalid XML syntax | Validate modconfig.xml format |
| `mod X was not successfully downloaded` | SteamCMD failed to download | Check mod exists, network connection |

## Testing

Run the included tests:

```bash
cd go
go test -v
```

See `examples/example.go` for more usage examples.
