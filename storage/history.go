package storage

import (
    "encoding/json"
    "os"

    "quietscan/types"
)

func LoadHistory() []types.ScanResult {
    data, err := os.ReadFile("results.json")
    if err != nil {
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
