// Package modupdater provides functionality to update Stationeers server mods via SteamCMD.
// This package is designed to be used as a library in other Go projects.
package modupdater

import (
	"fmt"
)

// UpdateOptions contains configuration for updating mods
type UpdateOptions struct {
	// ConfigLocation is the path to the modconfig.xml file
	ConfigLocation string

	// SteamCmdPath is the path to the steamcmd executable (optional, will search PATH if empty)
	SteamCmdPath string

	// ModIDs is a list of specific mod IDs to update (if empty, updates all enabled mods)
	ModIDs []uint64

	// IgnoreDisabled determines whether to skip disabled mods (default: true)
	IgnoreDisabled bool
}

// UpdateResult contains information about the update operation
type UpdateResult struct {
	// UpdatedMods is the list of mods that were successfully updated
	UpdatedMods []ModUpdateInfo

	// Errors contains any errors that occurred during the update
	Errors []error
}

// ModUpdateInfo contains information about a single mod update
type ModUpdateInfo struct {
	WorkshopID  uint64
	InstallPath string
	SourcePath  string
}

// UpdateMods is the main entry point for updating Stationeers mods.
// It parses the modconfig.xml, downloads mods via SteamCMD, and relocates them.
func UpdateMods(opts UpdateOptions) (*UpdateResult, error) {
	result := &UpdateResult{
		UpdatedMods: make([]ModUpdateInfo, 0),
		Errors:      make([]error, 0),
	}

	// Parse the modconfig.xml file
	config, err := ParseModConfig(opts.ConfigLocation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse mod config: %w", err)
	}

	// Get the list of mods to update
	mods, err := config.GetModsToUpdate(opts.ConfigLocation, opts.IgnoreDisabled, opts.ModIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get mods to update: %w", err)
	}

	if len(mods) == 0 {
		// No mods to update
		return result, nil
	}

	// Find steamcmd
	steamCmdPath, err := FindSteamCmd(opts.SteamCmdPath)
	if err != nil {
		return nil, fmt.Errorf("steamcmd not found: %w", err)
	}

	// Download mods via steamcmd
	downloaded, err := ExecuteSteamCmd(steamCmdPath, mods)
	if err != nil {
		return nil, fmt.Errorf("steamcmd execution failed: %w", err)
	}

	// Create a map of downloaded mods for quick lookup
	downloadedMap := make(map[uint64]string)
	for _, dm := range downloaded {
		downloadedMap[dm.ModID] = dm.Path
	}

	// Relocate mods to their install paths
	for _, mod := range mods {
		sourcePath, found := downloadedMap[mod.WorkshopID]
		if !found {
			err := fmt.Errorf("mod %d was not found in downloaded mods", mod.WorkshopID)
			result.Errors = append(result.Errors, err)
			continue
		}

		if err := CopyDirContent(sourcePath, mod.InstallPath); err != nil {
			err := fmt.Errorf("failed to relocate mod %d: %w", mod.WorkshopID, err)
			result.Errors = append(result.Errors, err)
			continue
		}

		result.UpdatedMods = append(result.UpdatedMods, ModUpdateInfo{
			WorkshopID:  mod.WorkshopID,
			InstallPath: mod.InstallPath,
			SourcePath:  sourcePath,
		})
	}

	// If any errors occurred, return them
	if len(result.Errors) > 0 {
		return result, fmt.Errorf("some mods failed to update (%d errors)", len(result.Errors))
	}

	return result, nil
}
