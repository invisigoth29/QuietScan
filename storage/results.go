package storage

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    "quietscan/types"
)

// getResultsFilePath returns the path to the results.json file
// Uses OS-appropriate application data directory for security
func getResultsFilePath() string {
    // Try to get OS-specific config directory
    configDir, err := os.UserConfigDir()
    if err != nil {
        // Fallback to current directory if config dir is unavailable
        return "results.json"
    }

    // Create QuietScan subdirectory in config dir
    appDir := filepath.Join(configDir, "QuietScan")

    // SECURITY: Ensure directory exists with restricted permissions (owner only)
    if err := os.MkdirAll(appDir, 0700); err != nil {
        // Fallback to current directory if mkdir fails
        fmt.Fprintf(os.Stderr, "Warning: Could not create app data directory, using current directory: %v\n", err)
        return "results.json"
    }

    return filepath.Join(appDir, "results.json")
}

func SaveScanResult(r *types.ScanResult) {
    FileLock.Lock()
    defer FileLock.Unlock()

    h := LoadHistory()
    if r != nil {
        h = append([]types.ScanResult{*r}, h...)
        if len(h) > 5 {
            h = h[:5]
        }
    } else {
        h = []types.ScanResult{}
    }

    data, err := json.MarshalIndent(map[string][]types.ScanResult{"history": h}, "", "  ")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error marshaling scan results: %v\n", err)
        return
    }

    resultsPath := getResultsFilePath()
    // SECURITY: Use 0600 permissions (owner read/write only) instead of 0644
    if err := os.WriteFile(resultsPath, data, 0600); err != nil {
        fmt.Fprintf(os.Stderr, "Error writing scan results to %s: %v\n", resultsPath, err)
    }
}

func LoadLatestResults() *types.ScanResult {
    h := LoadHistory()
    if len(h) == 0 {
        return nil
    }
    return &h[0]
}
