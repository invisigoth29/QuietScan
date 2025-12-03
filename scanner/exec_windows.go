//go:build windows
// +build windows

package scanner

import (
	"os/exec"
	"syscall"
)

// hideWindow sets the HideWindow attribute on Windows to suppress console windows
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

