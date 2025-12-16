//go:build windows
// +build windows

package scanner

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"sync"
	"time"
)

// checkSudoAccess is a no-op on Windows (arp command works without elevation for reading)
// Returns true since Windows doesn't require sudo
func checkSudoAccess() bool {
	return true
}

// arpScanWindows performs ARP scanning on Windows using arp -d + ping method
// Clears ARP cache, pings each IP to force ARP, then reads the cache
func arpScanWindows(ips []string) map[string]string {
	if len(ips) == 0 {
		return make(map[string]string)
	}

	// Clear entire ARP cache once at the start
	clearCmd := exec.Command("arp", "-d", "*")
	hideWindow(clearCmd)
	clearCmd.Run() // Ignore errors - cache might already be empty

	// Use worker pool to ping each IP (forces ARP)
	var wg sync.WaitGroup

	// Use configured concurrency
	workerLimit := currentWorkers
	if workerLimit > len(ips) {
		workerLimit = len(ips)
	}
	semaphore := make(chan struct{}, workerLimit)

	// Ping each IP to populate ARP cache
	for _, ip := range ips {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore

		go func(targetIP string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore

			// Apply politeness delay plus jitter
			if currentDelayMs > 0 {
				time.Sleep(time.Duration(currentDelayMs) * time.Millisecond)
			}
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)

			// Ping to force ARP (timeout in milliseconds for Windows)
			timeoutMs := currentTimeoutMs
			if timeoutMs > 5000 {
				timeoutMs = 5000 // Clamp to 5 seconds for ping
			}

			pingCmd := exec.Command("ping", "-n", "1", "-w", fmt.Sprintf("%d", timeoutMs), targetIP)
			hideWindow(pingCmd)

			// Apply timeout context
			timeout := time.Duration(currentTimeoutMs) * time.Millisecond
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			pingCmd = exec.CommandContext(ctx, "ping", "-n", "1", "-w", fmt.Sprintf("%d", timeoutMs), targetIP)
			hideWindow(pingCmd)

			pingCmd.Run() // Ignore errors - we just want to populate ARP
		}(ip)
	}

	// Wait for all pings to complete
	wg.Wait()

	// Small delay to ensure ARP cache is updated
	time.Sleep(100 * time.Millisecond)

	// Read the populated ARP cache
	// Use existing GetARPTable() which already handles Windows parsing
	return GetARPTable()
}

// arpScanWithArpScan is a stub on Windows (not available)
func arpScanWithArpScan(ips []string) map[string]string {
	return make(map[string]string)
}

// arpScanWithArping is a stub on Windows (not available)
func arpScanWithArping(ips []string) map[string]string {
	return make(map[string]string)
}
