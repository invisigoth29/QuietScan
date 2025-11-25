package scanner

import (
    "net"
    "os/exec"
    "strings"
)

func ResolveHostname(ip string) string {
    ptr, err := net.LookupAddr(ip)
    if err == nil && len(ptr) > 0 {
        return strings.Trim(ptr[0], ".")
    }
    out, _ := exec.Command("nbtstat", "-A", ip).Output()
    return parseNBTName(string(out))
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