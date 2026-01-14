package modupdater

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var workshopRegex = regexp.MustCompile(`(?i)workshop_(\d+)`)

// ModConfig represents the root XML structure
type ModConfig struct {
	XMLName xml.Name     `xml:"ModConfig"`
	Core    ConfigItem   `xml:"Core"`
	Locals  []ConfigItem `xml:"Local"`
}

// ConfigItem represents a mod entry in the config
type ConfigItem struct {
	Enabled bool     `xml:"Enabled,attr"`
	Path    PathItem `xml:"Path"`
}

// PathItem represents the path element
type PathItem struct {
	Value string `xml:"Value,attr"`
}

// WorkshopID extracts the workshop ID from the path value
func (p *PathItem) WorkshopID() (uint64, bool) {
	matches := workshopRegex.FindStringSubmatch(p.Value)
	if len(matches) < 2 {
		return 0, false
	}
	id, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}

// ParseModConfig reads and parses the modconfig.xml file
func ParseModConfig(configPath string) (*ModConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file at %s: %w", configPath, err)
	}

	var config ModConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config XML: %w", err)
	}

	return &config, nil
}

// ModToUpdate represents a mod that needs updating
type ModToUpdate struct {
	WorkshopID  uint64
	InstallPath string
}

// GetModsToUpdate filters and returns mods that should be updated
func (c *ModConfig) GetModsToUpdate(configPath string, ignoreDisabled bool, specificModIDs []uint64) ([]ModToUpdate, error) {
	modsDir := filepath.Join(filepath.Dir(configPath), "mods")

	var mods []ModToUpdate
	specificModMap := make(map[uint64]bool)
	for _, id := range specificModIDs {
		specificModMap[id] = true
	}

	for _, local := range c.Locals {
		// Skip disabled mods if requested
		if ignoreDisabled && !local.Enabled {
			continue
		}

		// Extract workshop ID
		workshopID, ok := local.Path.WorkshopID()
		if !ok {
			continue
		}

		// If specific mod IDs were provided, only include those
		if len(specificModIDs) > 0 && !specificModMap[workshopID] {
			continue
		}

		// Determine install path
		folderName := filepath.Base(local.Path.Value)
		installPath := filepath.Join(modsDir, folderName)

		mods = append(mods, ModToUpdate{
			WorkshopID:  workshopID,
			InstallPath: installPath,
		})
	}

	return mods, nil
}
