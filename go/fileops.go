package modupdater

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyDirContent copies or hardlinks files from src to dst directory
// It attempts to hardlink first, falling back to copy if hardlinking fails
func CopyDirContent(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			if err := CopyDirContent(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Remove existing file first so hardlink can succeed
			if _, err := os.Stat(dstPath); err == nil {
				if err := os.Remove(dstPath); err != nil {
					return fmt.Errorf("failed to remove existing file %s: %w", dstPath, err)
				}
			}

			// Try to hardlink first
			if err := os.Link(srcPath, dstPath); err != nil {
				// Hardlink failed, fall back to copy
				if err := copyFile(srcPath, dstPath); err != nil {
					return fmt.Errorf("failed to copy file: %w", err)
				}
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Copy file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}
