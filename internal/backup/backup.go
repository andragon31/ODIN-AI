// Package backup provides backup and restore functionality for ODIN
package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupInfo contains information about a backup
type BackupInfo struct {
	Path      string
	Timestamp time.Time
	Size      int64
	Source    string
}

// CreateBackup creates a backup of the specified path
// Returns the backup path on success
func CreateBackup(sourcePath string) (string, error) {
	// Verify source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return "", fmt.Errorf("source path does not exist: %s", sourcePath)
	}

	// Create backup directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	backupRoot := filepath.Join(homeDir, ".odin", "backup")
	if err := os.MkdirAll(backupRoot, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate timestamp-based backup name
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("backup-%s", timestamp)
	backupPath := filepath.Join(backupRoot, backupName)

	// Create the backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy source to backup
	if err := copyDir(sourcePath, backupPath); err != nil {
		// Clean up failed backup
		os.RemoveAll(backupPath)
		return "", fmt.Errorf("failed to copy files: %w", err)
	}

	return backupPath, nil
}

// RestoreBackup restores a backup to its original location
func RestoreBackup(backupPath string) error {
	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup path does not exist: %s", backupPath)
	}

	// Determine original path from backup
	// Backup structure: ~/.odin/backup/backup-TIMESTAMP -> ~/.odin
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	originalPath := filepath.Join(homeDir, ".odin")

	// Remove current .odin if exists
	if _, err := os.Stat(originalPath); err == nil {
		if err := os.RemoveAll(originalPath); err != nil {
			return fmt.Errorf("failed to remove current directory: %w", err)
		}
	}

	// Restore from backup
	if err := copyDir(backupPath, originalPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// ListBackups returns all available backups
func ListBackups() ([]BackupInfo, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	backupRoot := filepath.Join(homeDir, ".odin", "backup")

	// Check if backup directory exists
	if _, err := os.Stat(backupRoot); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	backups := []BackupInfo{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		backupPath := filepath.Join(backupRoot, entry.Name())
		info, err := os.Stat(backupPath)
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Path:      backupPath,
			Timestamp: info.ModTime(),
			Size:      getDirSize(backupPath),
			Source:    ".odin",
		})
	}

	return backups, nil
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Destination path
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			// Create directory
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(path, dstPath); err != nil {
				return err
			}
		}

		return nil
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

// getDirSize calculates the total size of a directory
func getDirSize(path string) int64 {
	var size int64

	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size
}
