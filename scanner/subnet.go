package scanner

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Safety limits for IP range scanning
const (
	MaxTotalIPsDefault = 256  // Default safety limit (/24 subnet)
	MaxTotalIPsHard    = 1024 // Hard ceiling, even with overrides
	MinCIDRPrefix      = 24   // Minimum allowed CIDR prefix length (/24 or smaller)
)

var (
	allowLargeRanges bool // Set via SetAllowLargeRanges() from CLI flag
)

// SetAllowLargeRanges sets whether large ranges are allowed (called from main)
func SetAllowLargeRanges(allow bool) {
	allowLargeRanges = allow
}

func GetActiveAdapter() (string, string) {
	ifaces, _ := net.Interfaces()

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && ipNet.IP.To4() != nil {
				return iface.Name, ipNet.String()
			}
		}
	}

	return "Unknown", "0.0.0.0/0"
}

// GetAllAdapters returns all non-loopback network interface names
func GetAllAdapters() []string {
	ifaces, _ := net.Interfaces()
	var adapters []string

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Check if interface has at least one IPv4 address
		addrs, _ := iface.Addrs()
		hasIPv4 := false
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && ipNet.IP.To4() != nil {
				hasIPv4 = true
				break
			}
		}

		if hasIPv4 {
			adapters = append(adapters, iface.Name)
		}
	}

	return adapters
}

// GetAdapterSubnet returns the subnet for a specific adapter name
func GetAdapterSubnet(adapterName string) string {
	ifaces, _ := net.Interfaces()

	for _, iface := range ifaces {
		if iface.Name != adapterName {
			continue
		}

		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && ipNet.IP.To4() != nil {
				return ipNet.String()
			}
		}
	}

	return "0.0.0.0/0"
}

// CountTargets counts the number of IP addresses that would be generated from an input string
// Returns the count and any validation errors
func CountTargets(input string) (int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, &ValidationError{
			Field:   "input",
			Message: "input cannot be empty",
		}
	}

	// Check if it's CIDR format (contains /)
	if strings.Contains(input, "/") {
		return countCIDRTargets(input)
	}

	// Check if it's range format (contains -)
	if strings.Contains(input, "-") {
		return countRangeTargets(input)
	}

	// Single IP address - validate it
	if err := ValidateIP(input); err != nil {
		return 0, err
	}
	return 1, nil
}

// ValidateTargets validates that a list of target IPs meets safety limits
func ValidateTargets(ips []string) error {
	totalHosts := len(ips)

	// Hard limit - always enforced
	if totalHosts > MaxTotalIPsHard {
		return fmt.Errorf("requested scan covers %d hosts, which exceeds the hard safety limit of %d. Reduce your scope", totalHosts, MaxTotalIPsHard)
	}

	// Default limit - requires override flag
	if totalHosts > MaxTotalIPsDefault && !allowLargeRanges {
		return fmt.Errorf("requested scan covers %d hosts. QuietScan's default limit is %d to avoid noisy scans. Reduce your scope or run with --allow-large-ranges to override", totalHosts, MaxTotalIPsDefault)
	}

	// Warn if large range is being scanned
	if totalHosts > MaxTotalIPsDefault && allowLargeRanges {
		fmt.Fprintf(os.Stderr, "WARNING: Large scan requested (>%d hosts). This may generate noticeable network traffic and trigger monitoring systems.\n", MaxTotalIPsDefault)
	}

	return nil
}

// ValidateCIDRPrefix validates that a CIDR prefix meets safety requirements
func ValidateCIDRPrefix(cidr string) error {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	ones, _ := ipnet.Mask.Size()
	if ones < MinCIDRPrefix {
		if !allowLargeRanges {
			return fmt.Errorf("QuietScan is designed for networks up to /%d. You requested %s. Narrow your range or run with --allow-large-ranges if you understand the impact", MinCIDRPrefix, cidr)
		}
		// With override, check total host count instead
		count, err := countCIDRTargets(cidr)
		if err != nil {
			return err
		}
		return ValidateTargets(make([]string, count))
	}

	return nil
}

// countCIDRTargets counts IPs in a CIDR without generating the full list
func countCIDRTargets(cidr string) (int, error) {
	if err := ValidateCIDR(cidr); err != nil {
		return 0, err
	}

	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Should not happen if ValidateCIDR passed, but handle gracefully
		return 0, &ValidationError{
			Field:   "CIDR",
			Value:   cidr,
			Message: fmt.Sprintf("failed to parse CIDR: %v", err),
		}
	}

	ones, bits := ipnet.Mask.Size()
	if bits != 32 {
		return 0, &ValidationError{
			Field:   "CIDR",
			Value:   cidr,
			Message: "invalid IPv4 CIDR format",
		}
	}

	// Calculate number of hosts: 2^(32-ones)
	hostBits := 32 - ones
	if hostBits > 31 {
		return 0, &ValidationError{
			Field:   "CIDR",
			Value:   cidr,
			Message: "CIDR prefix too small (must be /0 to /31)",
		}
	}

	count := 1 << hostBits
	return count, nil
}

// countRangeTargets counts IPs in a range without generating the full list
func countRangeTargets(rangeStr string) (int, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return 0, &ValidationError{
			Field:   "IP range",
			Value:   rangeStr,
			Message: "invalid range format: expected 'start-end' (e.g., 192.168.1.1-254)",
		}
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	// Validate start IP
	if err := ValidateIP(startStr); err != nil {
		return 0, &ValidationError{
			Field:   "IP range start",
			Value:   startStr,
			Message: fmt.Sprintf("invalid start IP: %v", err),
		}
	}

	startIP := net.ParseIP(startStr)
	if startIP == nil || startIP.To4() == nil {
		return 0, &ValidationError{
			Field:   "IP range start",
			Value:   startStr,
			Message: "failed to parse start IP address",
		}
	}

	// Ensure startIP is IPv4 format
	startIP4 := startIP.To4()
	if startIP4 == nil {
		return 0, fmt.Errorf("invalid start IP format")
	}

	var endIP net.IP
	if strings.Contains(endStr, ".") {
		// Full IP address - validate it
		if err := ValidateIP(endStr); err != nil {
			return 0, &ValidationError{
				Field:   "IP range end",
				Value:   endStr,
				Message: fmt.Sprintf("invalid end IP: %v", err),
			}
		}
		endIP = net.ParseIP(endStr)
		if endIP == nil || endIP.To4() == nil {
			return 0, &ValidationError{
				Field:   "IP range end",
				Value:   endStr,
				Message: "failed to parse end IP address",
			}
		}
		endIP = endIP.To4()
	} else {
		// Just the last octet
		endOctet, err := strconv.Atoi(endStr)
		if err != nil || endOctet < 0 || endOctet > 255 {
			return 0, &ValidationError{
				Field:   "IP range end",
				Value:   endStr,
				Message: fmt.Sprintf("invalid end octet (must be 0-255): %v", err),
			}
		}

		// Create endIP from startIP with replaced last octet
		endIP = make(net.IP, 4)
		copy(endIP, startIP4)
		endIP[3] = byte(endOctet)
	}

	// Calculate range size
	start := ipToInt(startIP4)
	end := ipToInt(endIP)
	if end < start {
		return 0, &ValidationError{
			Field:   "IP range",
			Value:   rangeStr,
			Message: fmt.Sprintf("start IP (%s) must be less than or equal to end IP (%s)", startIP4.String(), endIP.String()),
		}
	}
	count := int(end - start + 1)

	return count, nil
}

// ipToInt converts an IPv4 address to a 32-bit integer
func ipToInt(ip net.IP) uint32 {
	if len(ip) < 4 {
		return 0
	}
	// Ensure we have exactly 4 bytes
	ip4 := ip.To4()
	if ip4 == nil {
		return 0
	}
	return uint32(ip4[0])<<24 | uint32(ip4[1])<<16 | uint32(ip4[2])<<8 | uint32(ip4[3])
}

// ParseSubnetInput parses subnet input in CIDR or range format and returns a list of IP addresses
// CIDR format: "192.168.1.0/24"
// Range format: "192.168.1.1-254" or "192.168.1.1-192.168.1.254"
// Validates against safety limits before parsing
// Returns clear, user-friendly errors for invalid input
func ParseSubnetInput(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, &ValidationError{
			Field:   "subnet input",
			Message: "input cannot be empty. Expected CIDR (e.g., 192.168.1.0/24) or range (e.g., 192.168.1.1-254)",
		}
	}

	// Validate CIDR prefix if it's a CIDR
	if strings.Contains(input, "/") {
		if err := ValidateCIDR(input); err != nil {
			return nil, err
		}
		if err := ValidateCIDRPrefix(input); err != nil {
			return nil, err
		}
		return parseCIDR(input)
	}

	// For ranges and single IPs, count first then validate
	count, err := CountTargets(input)
	if err != nil {
		return nil, err
	}

	// Create dummy slice for validation
	dummyIPs := make([]string, count)
	if err := ValidateTargets(dummyIPs); err != nil {
		return nil, err
	}

	// Check if it's range format (contains -)
	if strings.Contains(input, "-") {
		return parseRange(input)
	}

	// Single IP address - validate it
	if err := ValidateIP(input); err != nil {
		return nil, err
	}
	return []string{input}, nil
}

// parseCIDR parses CIDR notation and returns all IPs in the range
// Assumes CIDR has already been validated by ValidateCIDR
func parseCIDR(cidr string) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		// This should not happen if ValidateCIDR was called, but handle it gracefully
		return nil, &ValidationError{
			Field:   "CIDR",
			Value:   cidr,
			Message: fmt.Sprintf("failed to parse CIDR: %v", err),
		}
	}

	if ipnet.IP.To4() == nil {
		return nil, &ValidationError{
			Field:   "CIDR",
			Value:   cidr,
			Message: "only IPv4 addresses are supported",
		}
	}

	var ips []string
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) == 0 {
		return nil, &ValidationError{
			Field:   "CIDR",
			Value:   cidr,
			Message: "CIDR range contains no valid IP addresses",
		}
	}

	return ips, nil
}

// parseRange parses range notation and returns all IPs in the range
// Supports formats: "192.168.1.1-254" or "192.168.1.1-192.168.1.254"
func parseRange(rangeStr string) ([]string, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, &ValidationError{
			Field:   "IP range",
			Value:   rangeStr,
			Message: "invalid range format: expected 'start-end' (e.g., 192.168.1.1-254)",
		}
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	// Validate start IP
	if err := ValidateIP(startStr); err != nil {
		return nil, &ValidationError{
			Field:   "IP range start",
			Value:   startStr,
			Message: fmt.Sprintf("invalid start IP: %v", err),
		}
	}

	// Parse start IP (already validated, so this should not fail)
	startIP := net.ParseIP(startStr)
	if startIP == nil || startIP.To4() == nil {
		return nil, &ValidationError{
			Field:   "IP range start",
			Value:   startStr,
			Message: "failed to parse start IP address",
		}
	}

	// Check if endStr is just a number (last octet) or full IP
	var endIP net.IP
	if strings.Contains(endStr, ".") {
		// Full IP address - validate it
		if err := ValidateIP(endStr); err != nil {
			return nil, &ValidationError{
				Field:   "IP range end",
				Value:   endStr,
				Message: fmt.Sprintf("invalid end IP: %v", err),
			}
		}
		endIP = net.ParseIP(endStr)
		if endIP == nil || endIP.To4() == nil {
			return nil, &ValidationError{
				Field:   "IP range end",
				Value:   endStr,
				Message: "failed to parse end IP address",
			}
		}
	} else {
		// Just the last octet
		endOctet, err := strconv.Atoi(endStr)
		if err != nil || endOctet < 0 || endOctet > 255 {
			return nil, &ValidationError{
				Field:   "IP range end",
				Value:   endStr,
				Message: fmt.Sprintf("invalid end octet (must be 0-255): %v", err),
			}
		}

		// Build end IP from start IP with replaced last octet
		startIP4 := startIP.To4()
		if startIP4 == nil {
			return nil, &ValidationError{
				Field:   "IP range start",
				Value:   startStr,
				Message: "invalid start IP format",
			}
		}
		endIP = make(net.IP, 4)
		copy(endIP, startIP4)
		endIP[3] = byte(endOctet)
	}

	// Ensure start <= end
	if compareIPs(startIP, endIP) > 0 {
		return nil, &ValidationError{
			Field:   "IP range",
			Value:   rangeStr,
			Message: fmt.Sprintf("start IP (%s) must be less than or equal to end IP (%s)", startIP.String(), endIP.String()),
		}
	}

	// Generate IP list
	var ips []string
	startIP4 := startIP.To4()
	if startIP4 == nil {
		return nil, &ValidationError{
			Field:   "IP range start",
			Value:   startStr,
			Message: "invalid start IP format",
		}
	}
	current := make(net.IP, 4)
	copy(current, startIP4)

	for {
		ips = append(ips, current.String())
		if current.Equal(endIP) {
			break
		}
		incIP(current)
		if compareIPs(current, endIP) > 0 {
			break
		}
	}

	if len(ips) == 0 {
		return nil, &ValidationError{
			Field:   "IP range",
			Value:   rangeStr,
			Message: "IP range contains no valid IP addresses",
		}
	}

	return ips, nil
}

// compareIPs compares two IP addresses
// Returns: -1 if ip1 < ip2, 0 if ip1 == ip2, 1 if ip1 > ip2
func compareIPs(ip1, ip2 net.IP) int {
	ip1 = ip1.To4()
	ip2 = ip2.To4()
	if ip1 == nil || ip2 == nil {
		return 0
	}

	for i := 0; i < 4; i++ {
		if ip1[i] < ip2[i] {
			return -1
		}
		if ip1[i] > ip2[i] {
			return 1
		}
	}
	return 0
}
