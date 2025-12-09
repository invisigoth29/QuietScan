package scanner

import (
	"context"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func ResolveHostname(ip string) string {
	// Try reverse DNS lookup first with timeout
	timeout := time.Duration(currentTimeoutMs) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	r := &net.Resolver{}
	ptr, err := r.LookupAddr(ctx, ip)
	if err == nil && len(ptr) > 0 {
		return strings.Trim(ptr[0], ".")
	}

	// On Windows, try nbtstat for NetBIOS names with timeout
	if runtime.GOOS == "windows" {
		cmd := exec.Command("nbtstat", "-A", ip)
		hideWindow(cmd)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
		hideWindow(cmd)

		out, err := cmd.Output()
		if err == nil {
			if name := parseNBTName(string(out)); name != "" {
				return name
			}
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
