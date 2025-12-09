//go:build !windows
// +build !windows

package scanner

import "os/exec"

// hideWindow is a no-op on non-Windows platforms
func hideWindow(cmd *exec.Cmd) {
	// No-op on Unix-like systems
}








