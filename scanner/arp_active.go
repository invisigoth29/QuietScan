package scanner

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// DetectARPTool determines which ARP scanning tool is available on the system
// Returns: "arp-scan", "arping", "windows", or "" if no tools available
func DetectARPTool() string {
	if runtime.GOOS == "windows" {
		return "windows"
	}

	// Unix-like systems: check for arp-scan first (preferred for batch scanning)
	if executableExists("arp-scan") {
		return "arp-scan"
	}

	// Fall back to arping if available
	if executableExists("arping") {
		return "arping"
	}

	// No active ARP tools available
	return ""
}

// executableExists checks if a command exists in PATH
func executableExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// ARPScanAll performs active ARP scanning on all provided IPs
// Returns a map of IP addresses to MAC addresses
// Falls back to passive ARP cache if active scanning fails or is unavailable
func ARPScanAll(ips []string) map[string]string {
	if len(ips) == 0 {
		return make(map[string]string)
	}

	tool := DetectARPTool()

	// Provide user feedback about tool availability
	switch tool {
	case "arp-scan":
		// Check sudo access on Unix systems
		if !checkSudoAccess() {
			fmt.Fprintf(os.Stderr, "WARNING: Active ARP scanning requires sudo/root privileges.\n")
			fmt.Fprintf(os.Stderr, "         Run with 'sudo quietscan' for accurate MAC addresses.\n")
			fmt.Fprintf(os.Stderr, "         Falling back to passive ARP cache (MAC addresses may be incomplete).\n\n")
			return GetARPTable()
		}
		return arpScanWithArpScan(ips)

	case "arping":
		// Check sudo access on Unix systems
		if !checkSudoAccess() {
			fmt.Fprintf(os.Stderr, "WARNING: Active ARP scanning requires sudo/root privileges.\n")
			fmt.Fprintf(os.Stderr, "         Run with 'sudo quietscan' for accurate MAC addresses.\n")
			fmt.Fprintf(os.Stderr, "         Falling back to passive ARP cache (MAC addresses may be incomplete).\n\n")
			return GetARPTable()
		}
		return arpScanWithArping(ips)

	case "windows":
		return arpScanWindows(ips)

	default:
		fmt.Fprintf(os.Stderr, "INFO: Active ARP tools (arp-scan, arping) not available.\n")
		fmt.Fprintf(os.Stderr, "      Using passive ARP cache (MAC addresses may be incomplete on new subnets).\n")
		fmt.Fprintf(os.Stderr, "      Install arp-scan for better results: brew install arp-scan (macOS) or apt install arp-scan (Linux)\n\n")
		return GetARPTable()
	}
}

// parseArpScanOutput parses the output from arp-scan command
// Expected format: "192.168.1.1    00:11:22:33:44:55    Vendor Name"
func parseArpScanOutput(output string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(output, "\n")

	// Regex to match IP and MAC address
	// Format: IP (whitespace) MAC (whitespace) optional_vendor
	re := regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s+([0-9a-fA-F:]{17})`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 3 {
			ip := matches[1]
			mac := normalizeMAC(matches[2])
			result[ip] = mac
		}
	}

	return result
}

// parseArpingOutput extracts the MAC address from arping command output
// Various formats:
// macOS: "60 bytes from 00:11:22:33:44:55 (192.168.1.1): index=0 time=1.234 msec"
// Linux: "Unicast reply from 192.168.1.1 [00:11:22:33:44:55]  0.645ms"
func parseArpingOutput(output string) string {
	// Regex to match MAC address in format XX:XX:XX:XX:XX:XX
	re := regexp.MustCompile(`([0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2})`)

	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return normalizeMAC(matches[1])
	}

	return ""
}

// normalizeMAC standardizes MAC address format to uppercase with colons
// Handles various formats: colons, dashes, mixed case
// Returns format: AA:BB:CC:DD:EE:FF
func normalizeMAC(mac string) string {
	// Remove common separators
	mac = strings.ReplaceAll(mac, "-", ":")
	mac = strings.ToUpper(mac)

	// Ensure we have exactly 17 characters (XX:XX:XX:XX:XX:XX)
	if len(mac) == 17 && strings.Count(mac, ":") == 5 {
		return mac
	}

	// If format is wrong, return as-is
	return mac
}
