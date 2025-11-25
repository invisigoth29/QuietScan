package scanner

import (
    "net"
)

func GetActiveAdapter() (string, string) {
    ifaces, _ := net.Interfaces()

    for _, iface := range ifaces {
        if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
            continue
        }

        addrs, _ := iface.Addrs()
        for _, addr := range addrs {
            ipNet, ok := addr.(*net.IPNet)
            if ok && ipNet.IP.To4() != nil {
                return iface.Name, ipNet.String()
            }
        }
    }

    return "Unknown", "0.0.0.0/0"
}