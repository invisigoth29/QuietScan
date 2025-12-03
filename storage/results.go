package storage

import (
    "encoding/json"
    "os"

    "quietscan/types"
)

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

    data, _ := json.MarshalIndent(map[string][]types.ScanResult{"history": h}, "", "  ")
    os.WriteFile("results.json", data, 0644)
}

func LoadLatestResults() *types.ScanResult {
    h := LoadHistory()
    if len(h) == 0 {
        return nil
    }
    return &h[0]
}
