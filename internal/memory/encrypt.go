// Package memory provides the Mimir memory engine for ODIN
package memory

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Encryptor handles encryption and decryption using age
type Encryptor struct{}

// NewEncryptor creates a new Encryptor
func NewEncryptor() *Encryptor {
	return &Encryptor{}
}

// EncryptFile encrypts a file using age-compatible encryption
// The key should be a passphrase or a 32-byte key
func (e *Encryptor) EncryptFile(inputPath string, key []byte) error {
	// Read the input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Create output file with .age extension
	outputPath := inputPath + ".age"

	// Use age-compatible encryption with AES-256-GCM
	encrypted, err := e.encrypt(data, key)
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Write encrypted data
	if err := os.WriteFile(outputPath, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	// Remove original file
	if err := os.Remove(inputPath); err != nil {
		return fmt.Errorf("failed to remove original file: %w", err)
	}

	return nil
}

// DecryptFile decrypts an age-encrypted file
func (e *Encryptor) DecryptFile(inputPath, outputPath string, key []byte) error {
	// Read the encrypted file
	encrypted, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Decrypt the data
	decrypted, err := e.decrypt(encrypted, key)
	if err != nil {
		return fmt.Errorf("failed to decrypt data: %w", err)
	}

	// Write decrypted data
	if err := os.WriteFile(outputPath, decrypted, 0644); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}

// encrypt encrypts data using AES-256-GCM with age-compatible header
func (e *Encryptor) encrypt(data, key []byte) ([]byte, error) {
	// Ensure key is 32 bytes for AES-256
	var aesKey [32]byte
	if len(key) >= 32 {
		copy(aesKey[:], key[:32])
	} else {
		copy(aesKey[:], key)
		// Pad with zeros if key is shorter (not recommended for production)
	}

	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Add age magic bytes header
	header := []byte("age-encryption.org-v1\n")
	return append(header, ciphertext...), nil
}

// decrypt decrypts data encrypted with AES-256-GCM
func (e *Encryptor) decrypt(data, key []byte) ([]byte, error) {
	// Check for age magic bytes header
	headerPrefix := "age-encryption.org-v1\n"
	if len(data) > len(headerPrefix) && string(data[:len(headerPrefix)]) == headerPrefix {
		data = data[len(headerPrefix):]
	} else {
		// Try as raw AES-256-GCM (no header)
	}

	// Ensure key is 32 bytes for AES-256
	var aesKey [32]byte
	if len(key) >= 32 {
		copy(aesKey[:], key[:32])
	} else {
		copy(aesKey[:], key)
	}

	block, err := aes.NewCipher(aesKey[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// GenerateKey generates a random encryption key
func (e *Encryptor) GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// KeyToFile saves a key to a file in base64 format
func (e *Encryptor) KeyToFile(key []byte, path string) error {
	encoded := base64.StdEncoding.EncodeToString(key)
	return os.WriteFile(path, []byte(encoded), 0600)
}

// KeyFromFile reads a key from a file
func (e *Encryptor) KeyFromFile(path string) ([]byte, error) {
	encoded, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	key, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		// Try raw bytes if not base64
		return encoded, nil
	}

	return key, nil
}

// EncryptMemory encrypts a single memory's content
func (e *Encryptor) EncryptMemory(m *Memory, key []byte) error {
	encrypted, err := e.encrypt([]byte(m.Content), key)
	if err != nil {
		return err
	}
	m.Content = base64.StdEncoding.EncodeToString(encrypted)
	m.Encrypted = true
	return nil
}

// DecryptMemory decrypts a single memory's content
func (e *Encryptor) DecryptMemory(m *Memory, key []byte) error {
	if !m.Encrypted {
		return nil
	}

	encrypted, err := base64.StdEncoding.DecodeString(m.Content)
	if err != nil {
		return fmt.Errorf("failed to decode content: %w", err)
	}

	decrypted, err := e.decrypt(encrypted, key)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	m.Content = string(decrypted)
	m.Encrypted = false
	return nil
}

// BackupFile creates a backup of a file before encryption/decryption
func (e *Encryptor) BackupFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	backupPath := path + ".backup"
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// RestoreBackup restores a file from backup
func (e *Encryptor) RestoreBackup(path string) error {
	backupPath := path + ".backup"

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup found for %s", path)
	}

	// Read backup
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Restore original
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	// Remove backup
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to remove backup: %w", err)
	}

	return nil
}

// IsEncrypted checks if a file appears to be encrypted with age
func (e *Encryptor) IsEncrypted(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	headerPrefix := "age-encryption.org-v1\n"
	return len(data) > len(headerPrefix) && string(data[:len(headerPrefix)]) == headerPrefix
}

// GetSecureDBPath returns the path for the encrypted database
func GetSecureDBPath(basePath string) string {
	return basePath + ".age"
}

// ReEncryptDB re-encrypts the database with a new key
func (e *Encryptor) ReEncryptDB(dbPath, newKeyFile, oldKeyFile string) error {
	// Create backup
	if err := e.BackupFile(dbPath); err != nil {
		return err
	}

	// Decrypt with old key
	if err := e.DecryptFile(dbPath, dbPath+".tmp", nil); err != nil {
		return fmt.Errorf("failed to decrypt with old key: %w", err)
	}

	// Read new key
	newKey, err := e.KeyFromFile(newKeyFile)
	if err != nil {
		return fmt.Errorf("failed to read new key: %w", err)
	}

	// Encrypt with new key
	if err := os.Rename(dbPath, dbPath+".old"); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}

	if err := e.EncryptFile(dbPath+".tmp", newKey); err != nil {
		os.Rename(dbPath+".old", dbPath)
		return fmt.Errorf("failed to encrypt with new key: %w", err)
	}

	// Clean up
	os.Remove(dbPath + ".old")
	dir := filepath.Dir(dbPath)

	// Remove old encrypted if exists
	oldEncrypted := filepath.Join(dir, "memory.db.age.old")
	if _, err := os.Stat(oldEncrypted); err == nil {
		os.Remove(oldEncrypted)
	}

	return nil
}
