package scanner

import (
	"math/rand"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"quietscan/types"
	"quietscan/vendordb"
)

var ouiDB vendordb.OUIDatabase

func init() {
	// Load OUI database on startup
	db, err := vendordb.LoadLocalOUI()
	if err == nil {
		ouiDB = db
	}
}

func GetLocalCIDR() string {
	_, subnet := GetActiveAdapter()
	return subnet
}

func LookupVendor(mac string) string {
	if mac == "" {
		return "Unknown Vendor"
	}

	// Normalize MAC address
	mac = strings.ToUpper(strings.ReplaceAll(mac, "-", ":"))

	// Extract OUI (first 3 octets)
	parts := strings.Split(mac, ":")
	if len(parts) < 3 {
		return "Unknown Vendor"
	}

	prefix := strings.Join(parts[:3], ":")

	if ouiDB == nil {
		db, err := vendordb.LoadLocalOUI()
		if err == nil {
			ouiDB = db
		} else {
			return "Unknown Vendor"
		}
	}

	if vendor, ok := ouiDB[prefix]; ok {
		return vendor
	}

	return "Unknown Vendor"
}

func PingHost(ip string) bool {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "500", ip)
	} else {
		// macOS/Linux: -c for count, -W for timeout (milliseconds)
		cmd = exec.Command("ping", "-c", "1", "-W", "500", ip)
	}
	hideWindow(cmd)

	out, _ := cmd.Output()
	output := string(out)
	// Check for both TTL= (Windows) and ttl= (macOS/Linux)
	return strings.Contains(output, "TTL=") || strings.Contains(output, "ttl=")
}

func EnumerateSubnet() []string {
	var ips []string

	_, ipnet, _ := net.ParseCIDR(GetLocalCIDR())
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		ips = append(ips, ip.String())
	}

	return ips
}

// ProgressCallback is called during scanning to report progress
// current: number of hosts scanned so far
// total: total number of hosts to scan
type ProgressCallback func(current, total int)

func RunPoliteScan() *types.ScanResult {
	return RunPoliteScanWithProgress(nil)
}

func RunPoliteScanWithProgress(progressCallback ProgressCallback) *types.ScanResult {
	arp := GetARPTable()
	ips := EnumerateSubnet()
	total := len(ips)

	var devices []types.DeviceEntry

	for i, ip := range ips {
		if progressCallback != nil {
			progressCallback(i, total)
		}

		time.Sleep(time.Duration(rand.Intn(300)+400) * time.Millisecond)

		reachable := PingHost(ip)
		if !reachable {
			continue
		}

		mac := arp[ip]
		vendor := LookupVendor(mac)
		host := ResolveHostname(ip)

		devices = append(devices, types.DeviceEntry{
			IP:       ip,
			MAC:      mac,
			Vendor:   vendor,
			Hostname: host,
		})
	}

	if progressCallback != nil {
		progressCallback(total, total)
	}

	adapter, subnet := GetActiveAdapter()
	return &types.ScanResult{
		Timestamp: time.Now(),
		Adapter:   adapter,
		Subnet:    subnet,
		Devices:   devices,
	}
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
