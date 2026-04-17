// Package e2e provides end-to-end tests for ODIN components
//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/odin-ai/odin/internal/catalog"
)

// TestCatalogE2E tests the catalog system end-to-end
func TestCatalogE2E(t *testing.T) {
	t.Run("ListComponents", func(t *testing.T) {
		cm := catalog.DefaultCatalogManager()
		components := cm.ListComponents()

		if len(components) == 0 {
			t.Skip("No components in catalog")
		}

		for _, comp := range components {
			if comp.Name == "" {
				t.Error("Component name should not be empty")
			}
		}
	})

	t.Run("SearchByTag", func(t *testing.T) {
		cm := catalog.DefaultCatalogManager()
		results := cm.SearchByTag("security")

		// Should return components with security tag (if any)
		for _, r := range results {
			t.Logf("Found: %s", r)
		}
	})

	t.Run("GetComponent", func(t *testing.T) {
		cm := catalog.DefaultCatalogManager()

		// Try to get a known component
		comp := cm.GetComponent("heimdall")
		if comp != nil {
			t.Logf("Found component: %s", comp.Name)
		}
	})

	t.Run("DetectInstalledAgents", func(t *testing.T) {
		cm := catalog.DefaultCatalogManager()
		agents := cm.DetectInstalledAgents()

		t.Logf("Detected %d agents", len(agents))
		for _, agent := range agents {
			t.Logf("Agent: %s", agent)
		}
	})

	t.Run("GetRune", func(t *testing.T) {
		cm := catalog.DefaultCatalogManager()
		rune := cm.GetRune("sdd-spec")

		if rune != nil {
			t.Logf("Found rune: %s", rune.Name)
		}
	})
}

// TestCatalogUpdateCheck tests catalog update functionality
func TestCatalogUpdateCheck(t *testing.T) {
	t.Run("CheckForUpdates", func(t *testing.T) {
		cm := catalog.DefaultCatalogManager()

		// This would check remote catalog for updates
		// In real scenario, this would hit a remote endpoint
		version := cm.GetCatalogVersion()
		t.Logf("Catalog version: %s", version)
	})
}
