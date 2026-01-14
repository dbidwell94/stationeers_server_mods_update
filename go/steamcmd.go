package modupdater

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

const stationeersAppID = 544550

var successCheckRegex = regexp.MustCompile(`(?i)downloaded item (\d+) to "([^"]*)"`)

// SteamCmdDownloadResult contains information about downloaded mods
type SteamCmdDownloadResult struct {
	ModID uint64
	Path  string
}

// ExecuteSteamCmd runs steamcmd to download the specified mods
func ExecuteSteamCmd(steamCmdPath string, mods []ModToUpdate) ([]SteamCmdDownloadResult, error) {
	// Build the command arguments
	args := []string{"+login anonymous"}

	for _, mod := range mods {
		args = append(args, fmt.Sprintf("+workshop_download_item %d %d validate", stationeersAppID, mod.WorkshopID))
	}

	args = append(args, "+logoff", "+quit")

	// Execute steamcmd
	cmd := exec.Command(steamCmdPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("steamcmd execution failed: %w (stderr: %s)", err, stderr.String())
	}

	// Check exit code
	if cmd.ProcessState.ExitCode() != 0 {
		return nil, fmt.Errorf("steamcmd exited with non-zero status: %d", cmd.ProcessState.ExitCode())
	}

	// Parse the output to verify downloads
	output := stdout.String()
	downloadedMods := parseDownloadedMods(output)

	// Verify all requested mods were downloaded
	downloadedMap := make(map[uint64]string)
	for _, dm := range downloadedMods {
		downloadedMap[dm.ModID] = dm.Path
	}

	for _, mod := range mods {
		if _, found := downloadedMap[mod.WorkshopID]; !found {
			return nil, fmt.Errorf("mod %d was not successfully downloaded", mod.WorkshopID)
		}
	}

	return downloadedMods, nil
}

// parseDownloadedMods extracts download information from steamcmd output
func parseDownloadedMods(output string) []SteamCmdDownloadResult {
	var results []SteamCmdDownloadResult

	matches := successCheckRegex.FindAllStringSubmatch(output, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			modID, err := strconv.ParseUint(match[1], 10, 64)
			if err != nil {
				continue
			}
			results = append(results, SteamCmdDownloadResult{
				ModID: modID,
				Path:  match[2],
			})
		}
	}

	return results
}

// FindSteamCmd locates the steamcmd executable
func FindSteamCmd(customPath string) (string, error) {
	if customPath != "" {
		return customPath, nil
	}

	path, err := exec.LookPath("steamcmd")
	if err != nil {
		return "", fmt.Errorf("steamcmd not found in PATH: %w", err)
	}

	return path, nil
}
