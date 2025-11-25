package storage

import (
    "encoding/json"
    "time"

    "quietscan/scanner"
)

type DeviceEntry struct {
    IP       string `json:"ip"`
    MAC      string `json:"mac"`
    Vendor   string `json:"vendor"`
    Hostname string `json:"hostname"`
}

type ScanResult struct {
    Timestamp time.Time     `json:"timestamp"`
    Adapter   string        `json:"adapter"`
    Subnet    string        `json:"subnet"`
    Devices   []DeviceEntry `json:"devices"`
}

func NewScanResult(devs []DeviceEntry) *ScanResult {
    adapter, subnet := scanner.GetActiveAdapter()
    return &ScanResult{
        Timestamp: time.Now(),
        Adapter:   adapter,
        Subnet:    subnet,
        Devices:   devs,
    }
}

func SaveScanResult(r *ScanResult) {
    FileLock.Lock()
    defer FileLock.Unlock()

    h := LoadHistory()
    if r != nil {
        h = append([]ScanResult{*r}, h...)
        if len(h) > 5 {
            h = h[:5]
        }
    } else {
        h = []ScanResult{}
    }

    data, _ := json.MarshalIndent(map[string][]ScanResult{"history": h}, "", "  ")
    os.WriteFile("results.json", data, 0644)
}

func LoadLatestResults() *ScanResult {
    h := LoadHistory()
    if len(h) == 0 {
        return nil
    }
    return &h[0]
}