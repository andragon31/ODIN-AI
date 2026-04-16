// Package backup provides backup and restore functionality for ODIN
package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateBackup(t *testing.T) {
	// Create a temp directory to backup
	tmpDir := os.TempDir()
	testDir := filepath.Join(tmpDir, "odin-test-backup-source")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(testDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create backup
	backupPath, err := CreateBackup(testDir)
	if err != nil {
		t.Fatalf("CreateBackup() error = %v", err)
	}

	if backupPath == "" {
		t.Error("CreateBackup() returned empty path")
	}

	// Verify backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("Backup directory does not exist")
	}

	// Verify backup contains the test file
	backupFile := filepath.Join(backupPath, "test.txt")
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Backup file does not exist")
	}

	// Cleanup
	os.RemoveAll(testDir)
	os.RemoveAll(backupPath)
}

func TestCreateBackupNonExistent(t *testing.T) {
	_, err := CreateBackup("/non/existent/path")
	if err == nil {
		t.Error("CreateBackup() should return error for non-existent path")
	}
}

func TestListBackups(t *testing.T) {
	backups, err := ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}

	if backups == nil {
		t.Error("ListBackups() returned nil")
	}
}

func TestBackupInfo(t *testing.T) {
	info := BackupInfo{
		Path:   "/test/path",
		Size:   1024,
		Source: ".odin",
	}

	if info.Path != "/test/path" {
		t.Errorf("BackupInfo.Path = %v, want '/test/path'", info.Path)
	}

	if info.Size != 1024 {
		t.Errorf("BackupInfo.Size = %v, want 1024", info.Size)
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := os.TempDir()
	src := filepath.Join(tmpDir, "test-src.txt")
	dst := filepath.Join(tmpDir, "test-dst.txt")

	// Write source file
	if err := os.WriteFile(src, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// Copy file
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	// Verify copy
	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(data) != "test content" {
		t.Errorf("Copied content = %v, want 'test content'", string(data))
	}

	// Cleanup
	os.Remove(src)
	os.Remove(dst)
}

func TestGetDirSize(t *testing.T) {
	tmpDir := os.TempDir()
	testDir := filepath.Join(tmpDir, "odin-test-size")

	// Create test directory
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test files
	for i := 0; i < 3; i++ {
		f, err := os.Create(filepath.Join(testDir, "file"+string(rune('0'+i))+".txt"))
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		f.WriteString("test content")
		f.Close()
	}

	// Get directory size
	size := getDirSize(testDir)
	if size == 0 {
		t.Error("getDirSize() returned 0, expected non-zero")
	}

	// Cleanup
	os.RemoveAll(testDir)
}
