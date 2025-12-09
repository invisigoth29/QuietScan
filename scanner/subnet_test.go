package scanner

import (
	"testing"
)

func TestCountTargets_SingleIP(t *testing.T) {
	count, err := CountTargets("192.168.1.1")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func TestCountTargets_CIDR_24(t *testing.T) {
	count, err := CountTargets("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 256 {
		t.Errorf("Expected count 256, got %d", count)
	}
}

func TestCountTargets_CIDR_25(t *testing.T) {
	count, err := CountTargets("192.168.1.0/25")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 128 {
		t.Errorf("Expected count 128, got %d", count)
	}
}

func TestCountTargets_CIDR_26(t *testing.T) {
	count, err := CountTargets("192.168.1.0/26")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 64 {
		t.Errorf("Expected count 64, got %d", count)
	}
}

func TestCountTargets_CIDR_23(t *testing.T) {
	count, err := CountTargets("192.168.0.0/23")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 512 {
		t.Errorf("Expected count 512, got %d", count)
	}
}

func TestCountTargets_Range_LastOctet(t *testing.T) {
	count, err := CountTargets("192.168.1.1-254")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 254 {
		t.Errorf("Expected count 254, got %d", count)
	}
}

func TestCountTargets_Range_FullIP(t *testing.T) {
	count, err := CountTargets("192.168.1.1-192.168.1.10")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 10 {
		t.Errorf("Expected count 10, got %d", count)
	}
}

func TestValidateTargets_DefaultLimit(t *testing.T) {
	// Reset flag state
	SetAllowLargeRanges(false)

	// Test with exactly 256 (should pass)
	ips := make([]string, 256)
	err := ValidateTargets(ips)
	if err != nil {
		t.Errorf("Expected no error for 256 IPs, got: %v", err)
	}

	// Test with 257 (should fail without override)
	ips = make([]string, 257)
	err = ValidateTargets(ips)
	if err == nil {
		t.Error("Expected error for 257 IPs without override, got nil")
	}
	if err != nil && !contains(err.Error(), "default limit") {
		t.Errorf("Expected error about default limit, got: %v", err)
	}
}

func TestValidateTargets_WithOverride(t *testing.T) {
	// Enable override
	SetAllowLargeRanges(true)

	// Test with 500 IPs (should pass with override)
	ips := make([]string, 500)
	err := ValidateTargets(ips)
	if err != nil {
		t.Errorf("Expected no error for 500 IPs with override, got: %v", err)
	}

	// Reset flag
	SetAllowLargeRanges(false)
}

func TestValidateTargets_HardLimit(t *testing.T) {
	// Test with 1025 IPs (should fail even with override)
	SetAllowLargeRanges(true)
	ips := make([]string, 1025)
	err := ValidateTargets(ips)
	if err == nil {
		t.Error("Expected error for 1025 IPs (hard limit), got nil")
	}
	if err != nil && !contains(err.Error(), "hard safety limit") {
		t.Errorf("Expected error about hard limit, got: %v", err)
	}

	SetAllowLargeRanges(false)
}

func TestValidateCIDRPrefix_24(t *testing.T) {
	SetAllowLargeRanges(false)
	err := ValidateCIDRPrefix("192.168.1.0/24")
	if err != nil {
		t.Errorf("Expected no error for /24, got: %v", err)
	}
}

func TestValidateCIDRPrefix_25(t *testing.T) {
	SetAllowLargeRanges(false)
	err := ValidateCIDRPrefix("192.168.1.0/25")
	if err != nil {
		t.Errorf("Expected no error for /25, got: %v", err)
	}
}

func TestValidateCIDRPrefix_26(t *testing.T) {
	SetAllowLargeRanges(false)
	err := ValidateCIDRPrefix("192.168.1.0/26")
	if err != nil {
		t.Errorf("Expected no error for /26, got: %v", err)
	}
}

func TestValidateCIDRPrefix_23_WithoutOverride(t *testing.T) {
	SetAllowLargeRanges(false)
	err := ValidateCIDRPrefix("192.168.0.0/23")
	if err == nil {
		t.Error("Expected error for /23 without override, got nil")
	}
	if err != nil && !contains(err.Error(), "up to /24") {
		t.Errorf("Expected error about /24 limit, got: %v", err)
	}
}

func TestValidateCIDRPrefix_23_WithOverride(t *testing.T) {
	SetAllowLargeRanges(true)
	err := ValidateCIDRPrefix("192.168.0.0/23")
	if err != nil {
		t.Errorf("Expected no error for /23 with override, got: %v", err)
	}
	SetAllowLargeRanges(false)
}

func TestValidateCIDRPrefix_16_WithOverride(t *testing.T) {
	SetAllowLargeRanges(true)
	err := ValidateCIDRPrefix("192.168.0.0/16")
	if err == nil {
		t.Error("Expected error for /16 (exceeds hard limit), got nil")
	}
	if err != nil && !contains(err.Error(), "hard safety limit") {
		t.Errorf("Expected error about hard limit, got: %v", err)
	}
	SetAllowLargeRanges(false)
}

func TestParseSubnetInput_24_Allowed(t *testing.T) {
	SetAllowLargeRanges(false)
	ips, err := ParseSubnetInput("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(ips) != 256 {
		t.Errorf("Expected 256 IPs, got %d", len(ips))
	}
}

func TestParseSubnetInput_23_Blocked(t *testing.T) {
	SetAllowLargeRanges(false)
	_, err := ParseSubnetInput("192.168.0.0/23")
	if err == nil {
		t.Error("Expected error for /23 without override, got nil")
	}
	if err != nil && !contains(err.Error(), "up to /24") {
		t.Errorf("Expected error about /24 limit, got: %v", err)
	}
}

func TestParseSubnetInput_23_AllowedWithOverride(t *testing.T) {
	SetAllowLargeRanges(true)
	ips, err := ParseSubnetInput("192.168.0.0/23")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(ips) != 512 {
		t.Errorf("Expected 512 IPs, got %d", len(ips))
	}
	SetAllowLargeRanges(false)
}

func TestParseSubnetInput_Range_257_Blocked(t *testing.T) {
	SetAllowLargeRanges(false)
	// Use a valid range that exceeds 256 hosts
	_, err := ParseSubnetInput("192.168.1.1-192.168.2.10")
	if err == nil {
		t.Error("Expected error for range >256 without override, got nil")
	}
	if err != nil && !contains(err.Error(), "default limit") {
		t.Errorf("Expected error about default limit, got: %v", err)
	}
}

func TestParseSubnetInput_Range_257_AllowedWithOverride(t *testing.T) {
	SetAllowLargeRanges(true)
	ips, err := ParseSubnetInput("192.168.1.1-192.168.1.257")
	if err != nil {
		// This might fail due to invalid IP, but if it parses, should be allowed
		if !contains(err.Error(), "invalid") {
			t.Fatalf("Unexpected error: %v", err)
		}
	}
	// Test with valid range that's >256
	ips, err = ParseSubnetInput("192.168.1.1-192.168.2.10")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(ips) <= 256 {
		t.Errorf("Expected >256 IPs, got %d", len(ips))
	}
	SetAllowLargeRanges(false)
}

func TestParseSubnetInput_HardLimit(t *testing.T) {
	SetAllowLargeRanges(true)
	// Try to parse a /16 which exceeds hard limit
	_, err := ParseSubnetInput("192.168.0.0/16")
	if err == nil {
		t.Error("Expected error for /16 (exceeds hard limit), got nil")
	}
	if err != nil && !contains(err.Error(), "hard safety limit") {
		t.Errorf("Expected error about hard limit, got: %v", err)
	}
	SetAllowLargeRanges(false)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}


