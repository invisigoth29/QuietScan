package scanner

import (
	"testing"
)

func TestValidateIP_Valid(t *testing.T) {
	testCases := []string{
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
		"127.0.0.1",
		"0.0.0.0",
		"255.255.255.255",
	}

	for _, ip := range testCases {
		err := ValidateIP(ip)
		if err != nil {
			t.Errorf("Expected no error for valid IP %s, got: %v", ip, err)
		}
	}
}

func TestValidateIP_Invalid(t *testing.T) {
	testCases := []struct {
		ip          string
		expectError bool
		errorMsg    string
	}{
		{"", true, "empty"},
		{"invalid", true, "not a valid"},
		{"192.168.1", true, "not a valid"},
		{"192.168.1.256", true, "not a valid"},
		{"192.168.1.1.1", true, "not a valid"},
		{"::1", true, "only IPv4"},
		{"2001:db8::1", true, "only IPv4"},
		{"  192.168.1.1  ", false, ""}, // Should trim and be valid
	}

	for _, tc := range testCases {
		err := ValidateIP(tc.ip)
		if tc.expectError {
			if err == nil {
				t.Errorf("Expected error for invalid IP %s, got nil", tc.ip)
			} else if !contains(err.Error(), tc.errorMsg) {
				t.Errorf("Expected error message containing '%s' for IP %s, got: %v", tc.errorMsg, tc.ip, err)
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error for IP %s, got: %v", tc.ip, err)
			}
		}
	}
}

func TestValidateCIDR_Valid(t *testing.T) {
	testCases := []string{
		"192.168.1.0/24",
		"10.0.0.0/8",
		"172.16.0.0/16",
		"192.168.1.0/25",
		"192.168.1.0/26",
		"0.0.0.0/0",
	}

	for _, cidr := range testCases {
		err := ValidateCIDR(cidr)
		if err != nil {
			t.Errorf("Expected no error for valid CIDR %s, got: %v", cidr, err)
		}
	}
}

func TestValidateCIDR_Invalid(t *testing.T) {
	testCases := []struct {
		cidr        string
		expectError bool
		errorMsg    string
	}{
		{"", true, "empty"},
		{"192.168.1.0", true, "must contain"},
		{"192.168.1.0/", true, "invalid CIDR"},
		{"/24", true, "invalid CIDR"},
		{"192.168.1.256/24", true, "invalid CIDR"},
		{"192.168.1.0/33", true, "invalid CIDR"},
		{"2001:db8::/32", true, "only IPv4"},
		{"invalid/24", true, "invalid CIDR"},
	}

	for _, tc := range testCases {
		err := ValidateCIDR(tc.cidr)
		if tc.expectError {
			if err == nil {
				t.Errorf("Expected error for invalid CIDR %s, got nil", tc.cidr)
			} else if !contains(err.Error(), tc.errorMsg) {
				t.Errorf("Expected error message containing '%s' for CIDR %s, got: %v", tc.errorMsg, tc.cidr, err)
			}
		} else {
			if err != nil {
				t.Errorf("Expected no error for CIDR %s, got: %v", tc.cidr, err)
			}
		}
	}
}

func TestValidateIPList_Empty(t *testing.T) {
	validIPs, invalidCount, err := ValidateIPList([]string{})
	if err == nil {
		t.Error("Expected error for empty IP list, got nil")
	}
	if len(validIPs) != 0 {
		t.Errorf("Expected 0 valid IPs, got %d", len(validIPs))
	}
	if invalidCount != 0 {
		t.Errorf("Expected 0 invalid IPs, got %d", invalidCount)
	}
	if !contains(err.Error(), "no valid scan targets") {
		t.Errorf("Expected error about no valid targets, got: %v", err)
	}
}

func TestValidateIPList_AllValid(t *testing.T) {
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
	validIPs, invalidCount, err := ValidateIPList(ips)
	if err != nil {
		t.Errorf("Expected no error for valid IP list, got: %v", err)
	}
	if len(validIPs) != 3 {
		t.Errorf("Expected 3 valid IPs, got %d", len(validIPs))
	}
	if invalidCount != 0 {
		t.Errorf("Expected 0 invalid IPs, got %d", invalidCount)
	}
}

func TestValidateIPList_AllInvalid(t *testing.T) {
	ips := []string{"invalid", "192.168.1.256", "not.an.ip"}
	validIPs, invalidCount, err := ValidateIPList(ips)
	if err == nil {
		t.Error("Expected error for all invalid IP list, got nil")
	}
	if len(validIPs) != 0 {
		t.Errorf("Expected 0 valid IPs, got %d", len(validIPs))
	}
	if invalidCount != 3 {
		t.Errorf("Expected 3 invalid IPs, got %d", invalidCount)
	}
	if !contains(err.Error(), "no valid scan targets") {
		t.Errorf("Expected error about no valid targets, got: %v", err)
	}
}

func TestValidateIPList_Mixed(t *testing.T) {
	ips := []string{"192.168.1.1", "invalid", "192.168.1.2", "192.168.1.256", "10.0.0.1"}
	validIPs, invalidCount, err := ValidateIPList(ips)
	// Should return warning but continue with valid IPs
	if len(validIPs) != 3 {
		t.Errorf("Expected 3 valid IPs, got %d", len(validIPs))
	}
	if invalidCount != 2 {
		t.Errorf("Expected 2 invalid IPs, got %d", invalidCount)
	}
	// Should have error/warning about invalid IPs
	if err == nil {
		t.Error("Expected warning about invalid IPs, got nil")
	}
}

func TestValidateIPList_WithComments(t *testing.T) {
	ips := []string{"# Comment line", "192.168.1.1", "  # Another comment", "192.168.1.2", ""}
	validIPs, invalidCount, err := ValidateIPList(ips)
	if err != nil {
		t.Errorf("Expected no error for IP list with comments, got: %v", err)
	}
	if len(validIPs) != 2 {
		t.Errorf("Expected 2 valid IPs, got %d", len(validIPs))
	}
	if invalidCount != 0 {
		t.Errorf("Expected 0 invalid IPs, got %d", invalidCount)
	}
}

func TestParseSubnetInput_InvalidCIDR(t *testing.T) {
	testCases := []string{
		"192.168.1.256/24",
		"invalid/24",
		"192.168.1.0/33",
		"/24",
		"192.168.1.0/",
	}

	for _, input := range testCases {
		_, err := ParseSubnetInput(input)
		if err == nil {
			t.Errorf("Expected error for invalid CIDR %s, got nil", input)
		}
		if !contains(err.Error(), "Invalid CIDR") && !contains(err.Error(), "invalid") {
			t.Errorf("Expected error about invalid CIDR for %s, got: %v", input, err)
		}
	}
}

func TestParseSubnetInput_InvalidIP(t *testing.T) {
	testCases := []string{
		"192.168.1.256",
		"invalid",
		"192.168.1",
		"::1",
	}

	for _, input := range testCases {
		_, err := ParseSubnetInput(input)
		if err == nil {
			t.Errorf("Expected error for invalid IP %s, got nil", input)
		}
		if !contains(err.Error(), "Invalid") && !contains(err.Error(), "invalid") {
			t.Errorf("Expected error about invalid IP for %s, got: %v", input, err)
		}
	}
}

func TestParseSubnetInput_InvalidRange(t *testing.T) {
	testCases := []string{
		"192.168.1.256-254",
		"invalid-254",
		"192.168.1.1-256",
		"192.168.1.10-5", // start > end
		"-254",
		"192.168.1.1-",
	}

	for _, input := range testCases {
		_, err := ParseSubnetInput(input)
		if err == nil {
			t.Errorf("Expected error for invalid range %s, got nil", input)
		}
		if !contains(err.Error(), "Invalid") && !contains(err.Error(), "invalid") {
			t.Errorf("Expected error about invalid range for %s, got: %v", input, err)
		}
	}
}

func TestParseSubnetInput_Empty(t *testing.T) {
	_, err := ParseSubnetInput("")
	if err == nil {
		t.Error("Expected error for empty input, got nil")
	}
	if !contains(err.Error(), "empty") && !contains(err.Error(), "cannot be empty") {
		t.Errorf("Expected error about empty input, got: %v", err)
	}
}


