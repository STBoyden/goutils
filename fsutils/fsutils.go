// Package fsutils provides simple file-system-based utility functions
package fsutils

import (
	"errors"
	"io/fs"
	"os"
)

// PathExists checks if path exists on the system filesystem.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, fs.ErrNotExist)
}

// PathExistsOnFS checks if a given path exists on a fs.FS
func PathExistsOnFS(filesystem fs.FS, path string) bool {
	_, err := filesystem.Open(path)
	return err == nil
}
