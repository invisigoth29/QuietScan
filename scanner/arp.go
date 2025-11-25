package scanner

import (
    "os/exec"
    "strings"
)

func GetARPTable() map[string]string {
    out, _ := exec.Command("arp", "-a").Output()
    lines := strings.Split(string(out), "\n")

    table := make(map[string]string)

    for _, ln := range lines {
        fields := strings.Fields(ln)
        if len(fields) >= 4 {
            ip := strings.Trim(fields[1], "()")
            mac := fields[3]
            table[ip] = mac
        }
    }
    return table
}