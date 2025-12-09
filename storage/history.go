package storage

import (
    "encoding/json"
    "os"

    "quietscan/types"
)

func LoadHistory() []types.ScanResult {
    // Use the secure results file path
    resultsPath := getResultsFilePath()
    data, err := os.ReadFile(resultsPath)
    if err != nil {
        // File doesn't exist or can't be read - return empty history
        return nil
    }
    var out struct {
        History []types.ScanResult `json:"history"`
    }
    if json.Unmarshal(data, &out) != nil {
        return nil
    }
    return out.History
}
