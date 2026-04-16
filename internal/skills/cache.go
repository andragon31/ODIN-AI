package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/odin-ai/odin/pkg/logger"
)

// Cache handles local caching of installed runes
type Cache struct {
	path    string
	index   *RuneIndex
	enabled bool
}

// NewCache creates a new cache instance
func NewCache(path string) (*Cache, error) {
	if path == "" {
		path = DefaultRunesPath()
	}

	cache := &Cache{
		path:    path,
		index:   NewRuneIndex(),
		enabled: true,
	}

	// Try to load existing index
	if err := cache.LoadIndex(); err != nil {
		logger.Debug("No existing cache index found, starting fresh")
	}

	return cache, nil
}

// LoadIndex loads the rune index from disk
func (c *Cache) LoadIndex() error {
	indexPath := filepath.Join(c.path, "index.json")

	data, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

	var index RuneIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("failed to parse index: %w", err)
	}

	c.index = &index
	return nil
}

// SaveIndex saves the rune index to disk
func (c *Cache) SaveIndex() error {
	if err := os.MkdirAll(c.path, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	indexPath := filepath.Join(c.path, "index.json")
	c.index.UpdatedAt = time.Now().Format(time.RFC3339)

	data, err := json.MarshalIndent(c.index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	logger.Debug("Cache index saved", "path", indexPath)
	return nil
}

// Store stores a rune in the cache
func (c *Cache) Store(r *Rune, source string) error {
	if !c.enabled {
		return fmt.Errorf("cache is disabled")
	}

	// Create rune directory
	runePath := filepath.Join(c.path, r.Name, r.Version)
	if err := os.MkdirAll(runePath, 0755); err != nil {
		return fmt.Errorf("failed to create rune directory: %w", err)
	}

	// Save rune data
	runeFile := filepath.Join(runePath, "rune.yaml")
	data, err := yamlMarshal(r)
	if err != nil {
		return fmt.Errorf("failed to marshal rune: %w", err)
	}

	if err := os.WriteFile(runeFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write rune file: %w", err)
	}

	// Update and save index
	r.InstalledAt = time.Now().Format(time.RFC3339)
	c.index.Add(r, source)
	if err := c.SaveIndex(); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	logger.Info("Rune stored in cache", "name", r.Name, "version", r.Version)
	return nil
}

// Load loads a rune from the cache
func (c *Cache) Load(name, version string) (*Rune, error) {
	runePath := filepath.Join(c.path, name, version, "rune.yaml")

	data, err := os.ReadFile(runePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("rune %s@%s not found in cache", name, version)
		}
		return nil, fmt.Errorf("failed to read rune: %w", err)
	}

	rune, result := ValidateYAML(data)
	if !result.Valid {
		return nil, fmt.Errorf("failed to validate rune: %s", result.Errors)
	}

	return rune, nil
}

// Remove removes a rune from the cache
func (c *Cache) Remove(name, version string) error {
	runePath := filepath.Join(c.path, name, version)

	if err := os.RemoveAll(runePath); err != nil {
		return fmt.Errorf("failed to remove rune: %w", err)
	}

	if err := c.index.Remove(name, version); err != nil {
		logger.Warn("Failed to update index", "error", err)
	}

	if err := c.SaveIndex(); err != nil {
		return fmt.Errorf("failed to save index: %w", err)
	}

	logger.Info("Rune removed from cache", "name", name, "version", version)
	return nil
}

// List returns all installed runes
func (c *Cache) List() ([]RuneMetadata, error) {
	var result []RuneMetadata
	for _, versions := range c.index.Runes {
		result = append(result, versions...)
	}
	return result, nil
}

// ListVersions returns all versions of a specific rune
func (c *Cache) ListVersions(name string) ([]RuneMetadata, error) {
	versions, ok := c.index.Get(name)
	if !ok {
		return nil, fmt.Errorf("rune %s not found", name)
	}
	return versions, nil
}

// Search searches for runes matching the query or tags
func (c *Cache) Search(query string, tags []string) ([]RuneMetadata, error) {
	var result []RuneMetadata
	queryLower := strings.ToLower(query)

	for name, versions := range c.index.Runes {
		for _, meta := range versions {
			// Try to load full rune for searching
			rune, err := c.Load(meta.Name, meta.Version)
			if err != nil {
				// If we can't load it, skip it but add metadata
				result = append(result, meta)
				continue
			}

			// Search by query (name or description)
			matchesQuery := query == "" ||
				strings.Contains(strings.ToLower(rune.Name), queryLower) ||
				strings.Contains(strings.ToLower(rune.Description), queryLower)

			// Search by tags
			matchesTags := len(tags) == 0
			if !matchesTags && len(rune.Tags) > 0 {
				for _, tag := range tags {
					for _, runeTag := range rune.Tags {
						if strings.ToLower(tag) == strings.ToLower(runeTag) {
							matchesTags = true
							break
						}
					}
					if matchesTags {
						break
					}
				}
			}

			if matchesQuery && matchesTags {
				result = append(result, meta)
			} else if query != "" && !matchesQuery && matchesTags {
				// Check if name matches even if description doesn't
				if strings.Contains(strings.ToLower(name), queryLower) {
					result = append(result, meta)
				}
			}
		}
	}

	return result, nil
}

// GetSource returns the source URL/path for an installed rune
func (c *Cache) GetSource(name, version string) (string, error) {
	versions, ok := c.index.Get(name)
	if !ok {
		return "", fmt.Errorf("rune %s not found", name)
	}

	for _, meta := range versions {
		if meta.Version == version {
			return meta.Source, nil
		}
	}

	return "", fmt.Errorf("version %s of rune %s not found", version, name)
}

// GetLatestVersion returns the latest version of a rune
func (c *Cache) GetLatestVersion(name string) (string, error) {
	versions, ok := c.index.Get(name)
	if !ok {
		return "", fmt.Errorf("rune %s not found", name)
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no versions found for %s", name)
	}

	// Return the last one (most recently added)
	return versions[len(versions)-1].Version, nil
}

// GetHistory returns the version history of a rune
func (c *Cache) GetHistory(name string) ([]RuneMetadata, error) {
	return c.ListVersions(name)
}

// IsInstalled checks if a rune is installed
func (c *Cache) IsInstalled(name, version string) bool {
	versions, ok := c.index.Get(name)
	if !ok {
		return false
	}

	for _, meta := range versions {
		if meta.Version == version {
			return true
		}
	}

	return false
}

// yamlMarshal is a wrapper around yaml.Marshal that handles marshaling
func yamlMarshal(r *Rune) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// DefaultCache creates a cache with the default path
func DefaultCache() (*Cache, error) {
	return NewCache(DefaultRunesPath())
}
