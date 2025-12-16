//go:build !windows
// +build !windows

package scanner

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// checkSudoAccess checks if sudo is available without password prompt
// Returns true if passwordless sudo is configured
func checkSudoAccess() bool {
	cmd := exec.Command("sudo", "-n", "true")
	err := cmd.Run()
	return err == nil
}

// arpScanWithArpScan performs batch ARP scanning using the arp-scan tool
// Batches IPs into groups for efficient scanning
func arpScanWithArpScan(ips []string) map[string]string {
	if len(ips) == 0 {
		return make(map[string]string)
	}

	result := make(map[string]string)
	const batchSize = 64 // Scan 64 IPs per batch

	// Process IPs in batches
	for i := 0; i < len(ips); i += batchSize {
		end := i + batchSize
		if end > len(ips) {
			end = len(ips)
		}

		batch := ips[i:end]

		// Build command: sudo arp-scan <ip1> <ip2> ...
		args := append([]string{"arp-scan"}, batch...)
		cmd := exec.Command("sudo", args...)

		// Apply timeout
		timeout := time.Duration(currentTimeoutMs) * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		cmd = exec.CommandContext(ctx, "sudo", args...)

		// Execute command
		output, err := cmd.Output()
		if err != nil {
			// Batch failed - fall back to passive ARP for these IPs
			fmt.Fprintf(os.Stderr, "WARNING: arp-scan batch failed, falling back to passive ARP for this batch\n")
			passive := GetARPTable()
			for _, ip := range batch {
				if mac, exists := passive[ip]; exists {
					result[ip] = mac
				}
			}
			continue
		}

		// Parse output and merge into result
		batchResult := parseArpScanOutput(string(output))
		for ip, mac := range batchResult {
			result[ip] = mac
		}

		// Apply politeness delay between batches (if not last batch)
		if end < len(ips) {
			if currentDelayMs > 0 {
				time.Sleep(time.Duration(currentDelayMs) * time.Millisecond)
			}
			// Add jitter
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
		}
	}

	return result
}

// arpScanWithArping performs per-IP ARP scanning using the arping tool
// Uses worker pool pattern similar to ping scanning
func arpScanWithArping(ips []string) map[string]string {
	if len(ips) == 0 {
		return make(map[string]string)
	}

	result := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use configured concurrency
	workerLimit := currentWorkers
	if workerLimit > len(ips) {
		workerLimit = len(ips)
	}
	semaphore := make(chan struct{}, workerLimit)

	// Determine timeout in seconds for arping
	timeoutSec := currentTimeoutMs/1000 + 1

	// Process each IP concurrently
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

			// Build platform-specific arping command
			var cmd *exec.Cmd
			if runtime.GOOS == "darwin" {
				// macOS: arping -C count ip
				cmd = exec.Command("sudo", "arping", "-C", "1", targetIP)
			} else {
				// Linux: arping -c count -w timeout ip
				cmd = exec.Command("sudo", "arping", "-c", "1", "-w", fmt.Sprintf("%d", timeoutSec), targetIP)
			}

			// Apply timeout context
			timeout := time.Duration(currentTimeoutMs) * time.Millisecond
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			if runtime.GOOS == "darwin" {
				cmd = exec.CommandContext(ctx, "sudo", "arping", "-C", "1", targetIP)
			} else {
				cmd = exec.CommandContext(ctx, "sudo", "arping", "-c", "1", "-w", fmt.Sprintf("%d", timeoutSec), targetIP)
			}

			// Execute command
			output, err := cmd.Output()
			if err != nil {
				// Command failed - no MAC for this IP
				return
			}

			// Parse MAC from output
			mac := parseArpingOutput(string(output))
			if mac != "" {
				mu.Lock()
				result[targetIP] = mac
				mu.Unlock()
			}
		}(ip)
	}

	// Wait for all workers to complete
	wg.Wait()

	return result
}

// arpScanWindows is a stub on Unix platforms (not available)
func arpScanWindows(ips []string) map[string]string {
	return make(map[string]string)
}
