package scanner

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"quietscan/types"
	"quietscan/vendordb"
)

// Concurrency limits for scan workers
const (
	DefaultConcurrency = 32    // Default number of concurrent workers
	MaxConcurrency     = 128   // Hard maximum, even with user override
	DefaultTimeoutMs   = 2000  // Default network timeout: 2 seconds
	MaxTimeoutMs       = 30000 // Maximum network timeout: 30 seconds
)

var (
	ouiDB            vendordb.OUIDatabase
	currentWorkers   int = DefaultConcurrency // Current concurrency setting
	currentDelayMs   int = 0                  // Per-target delay in milliseconds
	currentTimeoutMs int = DefaultTimeoutMs   // Current network timeout in milliseconds
)

// SetConcurrency sets the number of concurrent workers (called from main)
// Clamps to MaxConcurrency and returns true if clamping occurred
func SetConcurrency(workers int) bool {
	if workers > MaxConcurrency {
		currentWorkers = MaxConcurrency
		return true
	}
	if workers < 1 {
		currentWorkers = 1
		return false
	}
	currentWorkers = workers
	return false
}

// SetDelayMs sets the per-target delay in milliseconds (called from main)
func SetDelayMs(delayMs int) {
	if delayMs < 0 {
		delayMs = 0
	}
	currentDelayMs = delayMs
}

// GetConcurrency returns the current concurrency setting
func GetConcurrency() int {
	return currentWorkers
}

// GetDelayMs returns the current delay setting
func GetDelayMs() int {
	return currentDelayMs
}

func init() {
	// Load OUI database on startup
	db, err := vendordb.LoadLocalOUI()
	if err == nil {
		ouiDB = db
	}
}

func GetLocalCIDR() string {
	return GetLocalCIDRForAdapter("")
}

func GetLocalCIDRForAdapter(adapterName string) string {
	if adapterName == "" {
		_, subnet := GetActiveAdapter()
		return subnet
	}
	return GetAdapterSubnet(adapterName)
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

// SetTimeoutMs sets the network timeout in milliseconds
// Clamps to MaxTimeoutMs and returns true if clamping occurred
func SetTimeoutMs(timeoutMs int) bool {
	if timeoutMs > MaxTimeoutMs {
		currentTimeoutMs = MaxTimeoutMs
		return true
	}
	if timeoutMs < 0 {
		currentTimeoutMs = DefaultTimeoutMs
		return false
	}
	currentTimeoutMs = timeoutMs
	return false
}

// GetTimeoutMs returns the current network timeout in milliseconds
func GetTimeoutMs() int {
	return currentTimeoutMs
}

func PingHost(ip string) bool {
	// SECURITY: Validate IP address before using in command to prevent command injection
	if err := ValidateIP(ip); err != nil {
		// Invalid IP address - return false instead of attempting ping
		return false
	}

	timeoutMs := currentTimeoutMs
	// Clamp ping timeout to reasonable values (ping command has its own limits)
	if timeoutMs > 5000 {
		timeoutMs = 5000 // ping command max is typically around 5 seconds
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: -w is timeout in milliseconds
		cmd = exec.Command("ping", "-n", "1", "-w", fmt.Sprintf("%d", timeoutMs), ip)
	} else {
		// macOS/Linux: -c for count, -W for timeout (milliseconds)
		cmd = exec.Command("ping", "-c", "1", "-W", fmt.Sprintf("%d", timeoutMs), ip)
	}
	hideWindow(cmd)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(currentTimeoutMs)*time.Millisecond)
	defer cancel()

	// Attach context to command
	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	hideWindow(cmd)

	out, err := cmd.Output()
	if err != nil {
		// Timeout or other error - treat as not reachable
		return false
	}
	output := string(out)
	// Check for both TTL= (Windows) and ttl= (macOS/Linux)
	return strings.Contains(output, "TTL=") || strings.Contains(output, "ttl=")
}

func EnumerateSubnet() []string {
	return EnumerateSubnetForAdapter("")
}

func EnumerateSubnetForAdapter(adapterName string) []string {
	var ips []string

	cidr := GetLocalCIDRForAdapter(adapterName)

	// Validate the auto-detected CIDR
	if err := ValidateCIDR(cidr); err != nil {
		// If validation fails, return empty list - UI will handle the error
		return []string{}
	}

	if err := ValidateCIDRPrefix(cidr); err != nil {
		// If validation fails, return empty list - UI will handle the error
		return []string{}
	}

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Should not happen if ValidateCIDR passed, but handle gracefully
		return []string{}
	}

	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		ips = append(ips, ip.String())
	}

	// Validate the generated IP list
	if err := ValidateTargets(ips); err != nil {
		// If validation fails, return empty list - UI will handle the error
		return []string{}
	}

	// Re-validate each IP to ensure they're all valid
	validIPs, _, _ := ValidateIPList(ips)
	return validIPs
}

// ProgressCallback is called during scanning to report progress
// current: number of hosts scanned so far
// total: total number of hosts to scan
type ProgressCallback func(current, total int)

func RunPoliteScan() *types.ScanResult {
	return RunPoliteScanWithProgress(nil)
}

func RunPoliteScanWithProgress(progressCallback ProgressCallback) *types.ScanResult {
	return RunPoliteScanWithProgressForAdapter("", progressCallback)
}

func RunPoliteScanWithProgressForAdapter(adapterName string, progressCallback ProgressCallback) *types.ScanResult {
	ips := EnumerateSubnetForAdapter(adapterName)
	var adapter, subnet string
	if adapterName == "" {
		adapter, subnet = GetActiveAdapter()
	} else {
		adapter = adapterName
		subnet = GetAdapterSubnet(adapterName)
	}
	return RunPoliteScanWithIPs(ips, adapter, subnet, progressCallback)
}

// CheckHighIntensityScan checks if scan settings are high-intensity and prints warning if needed
// Warns if: totalHosts > 256 OR concurrency > 64 OR (no delay AND large host count)
func CheckHighIntensityScan(totalHosts int) {
	shouldWarn := false
	reasons := []string{}

	if totalHosts > 256 {
		shouldWarn = true
		reasons = append(reasons, fmt.Sprintf("large host count (%d)", totalHosts))
	}

	if currentWorkers > 64 {
		shouldWarn = true
		reasons = append(reasons, fmt.Sprintf("high concurrency (%d workers)", currentWorkers))
	}

	// Warn if scanning many hosts with no delay configured
	if currentDelayMs == 0 && totalHosts > 256 {
		shouldWarn = true
		reasons = append(reasons, "no delay configured")
	}

	if shouldWarn {
		fmt.Fprintf(os.Stderr, "WARNING: High-intensity scan settings (%s) may generate noticeable network traffic and trigger monitoring systems.\n", strings.Join(reasons, ", "))
	}
}

// RunPoliteScanWithIPs performs a scan with a custom list of IP addresses
// Uses concurrent workers for faster scanning while maintaining progress reporting
// Validates all IPs before scanning to ensure no invalid addresses are processed
func RunPoliteScanWithIPs(ips []string, adapter, subnet string, progressCallback ProgressCallback) *types.ScanResult {
	// Validate IP list before proceeding
	validIPs, invalidCount, err := ValidateIPList(ips)
	if err != nil {
		// Log validation errors but don't fail completely if some IPs are valid
		if invalidCount > 0 && len(validIPs) == 0 {
			// All IPs invalid - this is a fatal error
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return &types.ScanResult{
				Timestamp: time.Now(),
				Adapter:   adapter,
				Subnet:    subnet,
				Devices:   []types.DeviceEntry{},
			}
		}
		// Some IPs invalid but some valid - log warning and continue
		if invalidCount > 0 {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}

	// Use validated IPs
	ips = validIPs
	if len(ips) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No valid scan targets found.\n")
		return &types.ScanResult{
			Timestamp: time.Now(),
			Adapter:   adapter,
			Subnet:    subnet,
			Devices:   []types.DeviceEntry{},
		}
	}

	arp := GetARPTable()
	total := len(ips)

	// Check for high-intensity scan settings and warn if needed
	CheckHighIntensityScan(total)

	// Use worker pool pattern for concurrent scanning
	var wg sync.WaitGroup
	var mu sync.Mutex
	var devices []types.DeviceEntry

	// Progress tracking with atomic operations for thread safety
	var completed int64

	// Use configured concurrency, but don't exceed total IPs
	workerLimit := currentWorkers
	if workerLimit > total {
		workerLimit = total
	}
	semaphore := make(chan struct{}, workerLimit)

	// Process each IP concurrently
	for _, ip := range ips {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(targetIP string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			// Apply configured delay (if any) plus small random jitter for politeness
			if currentDelayMs > 0 {
				time.Sleep(time.Duration(currentDelayMs) * time.Millisecond)
			}
			// Always add small random jitter (50-150ms) to avoid synchronized bursts
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)

			reachable := PingHost(targetIP)
			if !reachable {
				// Update progress even for unreachable hosts
				current := atomic.AddInt64(&completed, 1)
				if progressCallback != nil {
					progressCallback(int(current), total)
				}
				return
			}

			// Get device information (ARP table is read-only, safe for concurrent reads)
			mac := arp[targetIP]
			vendor := LookupVendor(mac)
			host := ResolveHostname(targetIP)

			// Thread-safe append to devices slice
			mu.Lock()
			devices = append(devices, types.DeviceEntry{
				IP:       targetIP,
				MAC:      mac,
				Vendor:   vendor,
				Hostname: host,
			})
			mu.Unlock()

			// Update progress
			current := atomic.AddInt64(&completed, 1)
			if progressCallback != nil {
				progressCallback(int(current), total)
			}
		}(ip)
	}

	// Wait for all workers to complete
	wg.Wait()

	// Final progress update
	if progressCallback != nil {
		progressCallback(total, total)
	}

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
