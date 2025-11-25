package scanner

import (
    "math/rand"
    "net"
    "os/exec"
    "strings"
    "time"

    "quietscan/storage"
)

func PingHost(ip string) bool {
    out, _ := exec.Command("ping", "-n", "1", "-w", "500", ip).Output()
    return strings.Contains(string(out), "TTL=")
}

func EnumerateSubnet() []string {
    var ips []string

    _, ipnet, _ := net.ParseCIDR(GetLocalCIDR())
    for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
        ips = append(ips, ip.String())
    }

    return ips
}

func RunPoliteScan() *storage.ScanResult {
    arp := GetARPTable()
    ips := EnumerateSubnet()

    var devices []storage.DeviceEntry

    for _, ip := range ips {
        time.Sleep(time.Duration(rand.Intn(300)+400) * time.Millisecond)

        reachable := PingHost(ip)
        if !reachable {
            continue
        }

        mac := arp[ip]
        vendor := LookupVendor(mac)
        host := ResolveHostname(ip)

        devices = append(devices, storage.DeviceEntry{
            IP:       ip,
            MAC:      mac,
            Vendor:   vendor,
            Hostname: host,
        })
    }

    return storage.NewScanResult(devices)
}

func incIP(ip net.IP) {
    for j := len(ip) - 1; j >= 0; j-- {
        ip[j]++
        if ip[j] > 0 {
            break
        }
    }
}