// This example demonstrates how to use the modupdater library
// in a Go application (like the Stationeers Server UI)
package main

import (
	"fmt"
	"log"

	modupdater "github.com/dbidwell94/stationeers_server_mods_update/go"
)

func main() {
	// Example 1: Update all enabled mods
	fmt.Println("=== Example 1: Update all enabled mods ===")
	updateAllEnabledMods()

	// Example 2: Update specific mods
	fmt.Println("\n=== Example 2: Update specific mods ===")
	updateSpecificMods()

	// Example 3: Update all mods (including disabled)
	fmt.Println("\n=== Example 3: Update all mods (including disabled) ===")
	updateAllMods()

	// Example 4: Error handling
	fmt.Println("\n=== Example 4: Error handling ===")
	handleErrors()
}

func updateAllEnabledMods() {
	opts := modupdater.UpdateOptions{
		ConfigLocation: "/path/to/modconfig.xml",
		IgnoreDisabled: true,
	}

	result, err := modupdater.UpdateMods(opts)
	if err != nil {
		log.Printf("Failed to update mods: %v", err)
		return
	}

	fmt.Printf("Successfully updated %d mods\n", len(result.UpdatedMods))
	for _, mod := range result.UpdatedMods {
		fmt.Printf("  Mod %d: %s\n", mod.WorkshopID, mod.InstallPath)
	}
}

func updateSpecificMods() {
	opts := modupdater.UpdateOptions{
		ConfigLocation: "/path/to/modconfig.xml",
		ModIDs:         []uint64{3576112002, 3575689739},
		IgnoreDisabled: false, // Update even if disabled
	}

	result, err := modupdater.UpdateMods(opts)
	if err != nil {
		log.Printf("Failed to update mods: %v", err)
		return
	}

	fmt.Printf("Successfully updated %d specific mods\n", len(result.UpdatedMods))
	for _, mod := range result.UpdatedMods {
		fmt.Printf("  Mod %d: %s -> %s\n", mod.WorkshopID, mod.SourcePath, mod.InstallPath)
	}
}

func updateAllMods() {
	opts := modupdater.UpdateOptions{
		ConfigLocation: "/path/to/modconfig.xml",
		IgnoreDisabled: false,                      // Include disabled mods
		SteamCmdPath:   "/custom/path/to/steamcmd", // Optional custom path
	}

	result, err := modupdater.UpdateMods(opts)
	if err != nil {
		log.Printf("Failed to update mods: %v", err)
		return
	}

	fmt.Printf("Successfully updated %d mods (including disabled)\n", len(result.UpdatedMods))
}

func handleErrors() {
	opts := modupdater.UpdateOptions{
		ConfigLocation: "/path/to/modconfig.xml",
		IgnoreDisabled: true,
	}

	result, err := modupdater.UpdateMods(opts)
	if err != nil {
		log.Printf("Update operation encountered errors: %v", err)

		// Even if there was an error, some mods might have been updated
		if result != nil {
			fmt.Printf("Partial success: %d mods updated\n", len(result.UpdatedMods))

			// Print individual errors
			if len(result.Errors) > 0 {
				fmt.Println("Individual errors:")
				for i, e := range result.Errors {
					fmt.Printf("  %d. %v\n", i+1, e)
				}
			}
		}
		return
	}

	fmt.Println("All mods updated successfully!")
}

// Example integration in Stationeers Server UI
type StationeersServerUI struct {
	configPath string
}

func (s *StationeersServerUI) UpdateServerMods() error {
	opts := modupdater.UpdateOptions{
		ConfigLocation: s.configPath,
		IgnoreDisabled: true,
	}

	result, err := modupdater.UpdateMods(opts)
	if err != nil {
		return fmt.Errorf("mod update failed: %w", err)
	}

	// Log or display results to UI
	fmt.Printf("Updated %d mods\n", len(result.UpdatedMods))
	for _, mod := range result.UpdatedMods {
		fmt.Printf("  ✓ Mod %d updated\n", mod.WorkshopID)
	}

	return nil
}
