package scanner

import (
	"net"
	"os/exec"
	"runtime"
	"strings"
)

func ResolveHostname(ip string) string {
	// Try reverse DNS lookup first
	ptr, err := net.LookupAddr(ip)
	if err == nil && len(ptr) > 0 {
		return strings.Trim(ptr[0], ".")
	}

	// On Windows, try nbtstat for NetBIOS names
	if runtime.GOOS == "windows" {
		cmd := exec.Command("nbtstat", "-A", ip)
		hideWindow(cmd)
		out, _ := cmd.Output()
		if name := parseNBTName(string(out)); name != "" {
			return name
		}
	}

	return ""
}

func parseNBTName(out string) string {
	lines := strings.Split(out, "\n")
	for _, ln := range lines {
		if strings.Contains(ln, "<00>") {
			parts := strings.Fields(ln)
			if len(parts) > 0 {
				return parts[0]
			}
		}
	}
	return ""
}
