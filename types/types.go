package types

import "time"

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
