package scanner

import (
	"fmt"
	"net"
	"strings"
)

// ValidationError represents a validation error with a clear message
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("Invalid %s: %s - %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("Invalid %s: %s", e.Field, e.Message)
}

// ValidateIP validates a single IP address and returns a clear error if invalid
func ValidateIP(ipStr string) error {
	ipStr = strings.TrimSpace(ipStr)
	if ipStr == "" {
		return &ValidationError{
			Field:   "IP address",
			Value:   ipStr,
			Message: "IP address cannot be empty",
		}
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return &ValidationError{
			Field:   "IP address",
			Value:   ipStr,
			Message: "not a valid IP address format",
		}
	}

	if ip.To4() == nil {
		return &ValidationError{
			Field:   "IP address",
			Value:   ipStr,
			Message: "only IPv4 addresses are supported",
		}
	}

	return nil
}

// ValidateCIDR validates a CIDR notation string and returns a clear error if invalid
func ValidateCIDR(cidrStr string) error {
	cidrStr = strings.TrimSpace(cidrStr)
	if cidrStr == "" {
		return &ValidationError{
			Field:   "CIDR",
			Value:   cidrStr,
			Message: "CIDR cannot be empty",
		}
	}

	if !strings.Contains(cidrStr, "/") {
		return &ValidationError{
			Field:   "CIDR",
			Value:   cidrStr,
			Message: "CIDR must contain '/' (e.g., 192.168.1.0/24)",
		}
	}

	_, ipnet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return &ValidationError{
			Field:   "CIDR",
			Value:   cidrStr,
			Message: fmt.Sprintf("invalid CIDR format: %v", err),
		}
	}

	if ipnet.IP.To4() == nil {
		return &ValidationError{
			Field:   "CIDR",
			Value:   cidrStr,
			Message: "only IPv4 addresses are supported",
		}
	}

	return nil
}

// ValidateIPList validates a list of IP addresses and returns clear errors
// Returns the number of valid IPs and any validation errors
func ValidateIPList(ips []string) (validIPs []string, invalidCount int, err error) {
	if len(ips) == 0 {
		return nil, 0, &ValidationError{
			Field:   "target list",
			Message: "no valid scan targets found",
		}
	}

	validIPs = make([]string, 0, len(ips))
	invalidIPs := make([]string, 0)
	invalidCount = 0

	for _, ipStr := range ips {
		ipStr = strings.TrimSpace(ipStr)
		// Skip empty lines
		if ipStr == "" {
			continue
		}
		// Skip comment lines (starting with #)
		if strings.HasPrefix(ipStr, "#") {
			continue
		}

		if err := ValidateIP(ipStr); err != nil {
			invalidIPs = append(invalidIPs, ipStr)
			invalidCount++
			continue
		}

		validIPs = append(validIPs, ipStr)
	}

	if len(validIPs) == 0 {
		return nil, invalidCount, &ValidationError{
			Field:   "target list",
			Message: fmt.Sprintf("no valid scan targets found after filtering (found %d invalid entries)", invalidCount),
		}
	}

	if invalidCount > 0 {
		// Return warning but still proceed with valid IPs
		// The error will be logged but not fatal
		return validIPs, invalidCount, fmt.Errorf("filtered out %d invalid IP addresses: %s", invalidCount, strings.Join(invalidIPs[:min(5, len(invalidIPs))], ", "))
	}

	return validIPs, 0, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


