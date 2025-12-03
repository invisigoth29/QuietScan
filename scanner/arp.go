package scanner

import (
	"os/exec"
	"runtime"
	"strings"
)

func GetARPTable() map[string]string {
	cmd := exec.Command("arp", "-a")
	hideWindow(cmd)
	out, _ := cmd.Output()
	lines := strings.Split(string(out), "\n")

	table := make(map[string]string)

	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}

		// macOS format: ? (192.168.2.15) at 14:cb:19:96:e6:1 on en8 ifscope [ethernet]
		// Windows format: 192.168.1.1           00-11-22-33-44-55     dynamic
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			// macOS/Linux: look for pattern with parentheses
			ipStart := strings.Index(ln, "(")
			ipEnd := strings.Index(ln, ")")
			if ipStart >= 0 && ipEnd > ipStart {
				ip := ln[ipStart+1 : ipEnd]
				// Find MAC address after "at "
				atIdx := strings.Index(ln, " at ")
				if atIdx > 0 {
					macPart := ln[atIdx+4:]
					// MAC address ends at space or end of line
					spaceIdx := strings.Index(macPart, " ")
					if spaceIdx > 0 {
						mac := macPart[:spaceIdx]
						table[ip] = mac
					} else if len(macPart) > 0 {
						mac := macPart
						table[ip] = mac
					}
				}
			}
		} else {
			// Windows format
			fields := strings.Fields(ln)
			if len(fields) >= 2 {
				ip := strings.Trim(fields[0], "()")
				mac := fields[1]
				table[ip] = mac
			}
		}
	}
	return table
}
