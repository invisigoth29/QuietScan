package storage

import (
    "encoding/json"
    "os"
)

func LoadHistory() []ScanResult {
    data, err := os.ReadFile("results.json")
    if err != nil {
        return nil
    }
    var out struct {
        History []ScanResult `json:"history"`
    }
    if json.Unmarshal(data, &out) != nil {
        return nil
    }
    return out.History
}