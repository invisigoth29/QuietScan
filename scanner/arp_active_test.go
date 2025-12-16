package scanner

import (
	"runtime"
	"testing"
)

func TestDetectARPTool(t *testing.T) {
	tool := DetectARPTool()

	if runtime.GOOS == "windows" {
		if tool != "windows" {
			t.Errorf("Expected 'windows' on Windows platform, got '%s'", tool)
		}
	} else {
		// On Unix, should return arp-scan, arping, or empty string
		validTools := map[string]bool{
			"arp-scan": true,
			"arping":   true,
			"":         true,
		}
		if !validTools[tool] {
			t.Errorf("Expected 'arp-scan', 'arping', or '', got '%s'", tool)
		}
	}
}

func TestExecutableExists(t *testing.T) {
	// Test with a command that should always exist
	if !executableExists("ping") {
		t.Error("Expected 'ping' to exist in PATH")
	}

	// Test with a command that should not exist
	if executableExists("this-command-definitely-does-not-exist-12345") {
		t.Error("Expected non-existent command to return false")
	}
}

func TestParseArpScanOutput(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name: "Valid arp-scan output",
			input: `Interface: en0, type: EN10MB, MAC: aa:bb:cc:dd:ee:ff, IPv4: 192.168.1.100
Starting arp-scan 1.9.7 with 256 hosts
192.168.1.1	00:11:22:33:44:55	Vendor Name Inc.
192.168.1.2	aa:bb:cc:dd:ee:ff	Another Vendor

2 packets received by filter
Ending arp-scan 1.9.7: 256 hosts scanned in 2.000 seconds`,
			expected: map[string]string{
				"192.168.1.1": "00:11:22:33:44:55",
				"192.168.1.2": "AA:BB:CC:DD:EE:FF",
			},
		},
		{
			name:     "Empty output",
			input:    "",
			expected: map[string]string{},
		},
		{
			name: "Output with header only",
			input: `Interface: en0, type: EN10MB
Starting arp-scan 1.9.7`,
			expected: map[string]string{},
		},
		{
			name: "Mixed valid and invalid lines",
			input: `192.168.1.1	00:11:22:33:44:55	Vendor
This is an invalid line
192.168.1.2	ff:ee:dd:cc:bb:aa	Another Vendor`,
			expected: map[string]string{
				"192.168.1.1": "00:11:22:33:44:55",
				"192.168.1.2": "FF:EE:DD:CC:BB:AA",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseArpScanOutput(tc.input)

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d entries, got %d", len(tc.expected), len(result))
			}

			for ip, expectedMAC := range tc.expected {
				if actualMAC, exists := result[ip]; !exists {
					t.Errorf("Expected IP %s not found in result", ip)
				} else if actualMAC != expectedMAC {
					t.Errorf("For IP %s: expected MAC %s, got %s", ip, expectedMAC, actualMAC)
				}
			}
		})
	}
}

func TestParseArpingOutput(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "macOS format",
			input:    "60 bytes from 00:11:22:33:44:55 (192.168.1.1): index=0 time=1.234 msec",
			expected: "00:11:22:33:44:55",
		},
		{
			name:     "Linux format",
			input:    "Unicast reply from 192.168.1.1 [aa:bb:cc:dd:ee:ff]  0.645ms",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "No MAC in output",
			input:    "No reply received",
			expected: "",
		},
		{
			name:     "Empty output",
			input:    "",
			expected: "",
		},
		{
			name:     "Lowercase MAC",
			input:    "Reply from ff:ee:dd:cc:bb:aa",
			expected: "FF:EE:DD:CC:BB:AA",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseArpingOutput(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestNormalizeMAC(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Colon-separated uppercase",
			input:    "AA:BB:CC:DD:EE:FF",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "Colon-separated lowercase",
			input:    "aa:bb:cc:dd:ee:ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "Dash-separated",
			input:    "aa-bb-cc-dd-ee-ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "Mixed case with dashes",
			input:    "Aa-Bb-Cc-Dd-Ee-Ff",
			expected: "AA:BB:CC:DD:EE:FF",
		},
		{
			name:     "Already normalized",
			input:    "00:11:22:33:44:55",
			expected: "00:11:22:33:44:55",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeMAC(tc.input)
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestARPScanAll_EmptyList(t *testing.T) {
	result := ARPScanAll([]string{})
	if len(result) != 0 {
		t.Errorf("Expected empty map for empty IP list, got %d entries", len(result))
	}
}

func TestARPScanAll_NilList(t *testing.T) {
	result := ARPScanAll(nil)
	if result == nil {
		t.Error("Expected non-nil map even for nil input")
	}
	if len(result) != 0 {
		t.Errorf("Expected empty map for nil IP list, got %d entries", len(result))
	}
}
