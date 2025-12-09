package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateFilePath validates a file path before writing
// Returns an error if the path is invalid, unwritable, or exists without overwrite permission
func ValidateFilePath(filePath string, allowOverwrite bool) error {
	// Check if path is empty
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("cannot resolve file path: %v", err)
	}

	// Check if path refers to a directory
	info, err := os.Stat(absPath)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("path refers to a directory, not a file: %s", absPath)
		}
		// File exists - check overwrite permission
		if !allowOverwrite {
			return fmt.Errorf("file already exists (use --overwrite to allow): %s", absPath)
		}
		// Check if file is writable
		if info.Mode().Perm()&0200 == 0 {
			return fmt.Errorf("file exists but is not writable: %s", absPath)
		}
	}

	// Check if parent directory exists and is writable
	dir := filepath.Dir(absPath)
	dirInfo, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("parent directory does not exist: %s", dir)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("parent path is not a directory: %s", dir)
	}

	// Check if directory is writable
	if dirInfo.Mode().Perm()&0200 == 0 {
		return fmt.Errorf("parent directory is not writable: %s", dir)
	}

	return nil
}

// CheckFileExists returns true if the file exists
func CheckFileExists(filePath string) bool {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	info, err := os.Stat(absPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}


