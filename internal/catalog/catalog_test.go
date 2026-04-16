// Package catalog provides the component catalog system for ODIN
package catalog

import (
	"testing"
)

func TestKnownAgents(t *testing.T) {
	agents := KnownAgents()

	if len(agents) == 0 {
		t.Error("KnownAgents() returned empty slice")
	}

	// Verify all agents have required fields
	for _, agent := range agents {
		if agent.ID == "" {
			t.Error("Agent has empty ID")
		}
		if agent.Name == "" {
			t.Error("Agent has empty Name")
		}
		if agent.Description == "" {
			t.Error("Agent has empty Description")
		}
		if len(agent.SupportedOS) == 0 {
			t.Error("Agent has no supported OS")
		}
	}
}

func TestKnownComponents(t *testing.T) {
	components := KnownComponents()

	if len(components) == 0 {
		t.Error("KnownComponents() returned empty slice")
	}

	// Verify all components have required fields
	for _, comp := range components {
		if comp.ID == "" {
			t.Error("Component has empty ID")
		}
		if comp.Name == "" {
			t.Error("Component has empty Name")
		}
		if comp.Version == "" {
			t.Error("Component has empty Version")
		}
	}
}

func TestAvailableRunes(t *testing.T) {
	runes := AvailableRunes()

	if len(runes) == 0 {
		t.Error("AvailableRunes() returned empty slice")
	}

	// Verify all runes have required fields
	for _, rune := range runes {
		if rune.Name == "" {
			t.Error("Rune has empty Name")
		}
		if rune.Description == "" {
			t.Error("Rune has empty Description")
		}
	}
}

func TestCatalogManagerGetAgent(t *testing.T) {
	manager := NewCatalogManager()

	// Test getting existing agent
	agent := manager.GetAgent(AgentClaudeCode)
	if agent == nil {
		t.Error("GetAgent(claude-code) returned nil")
	}
	if agent.ID != AgentClaudeCode {
		t.Errorf("GetAgent(claude-code) returned wrong agent: %v", agent.ID)
	}

	// Test getting non-existent agent
	nonExistent := manager.GetAgent("non-existent")
	if nonExistent != nil {
		t.Error("GetAgent(non-existent) should return nil")
	}
}

func TestCatalogManagerGetComponent(t *testing.T) {
	manager := NewCatalogManager()

	// Test getting existing component
	comp := manager.GetComponent("sdd")
	if comp == nil {
		t.Error("GetComponent(sdd) returned nil")
	}
	if comp.ID != "sdd" {
		t.Errorf("GetComponent(sdd) returned wrong component: %v", comp.ID)
	}

	// Test getting non-existent component
	nonExistent := manager.GetComponent("non-existent")
	if nonExistent != nil {
		t.Error("GetComponent(non-existent) should return nil")
	}
}

func TestCatalogManagerGetRune(t *testing.T) {
	manager := NewCatalogManager()

	// Test getting existing rune
	rune := manager.GetRune("sdd-propose")
	if rune == nil {
		t.Error("GetRune(sdd-propose) returned nil")
	}
	if rune.Name != "sdd-propose" {
		t.Errorf("GetRune(sdd-propose) returned wrong rune: %v", rune.Name)
	}

	// Test getting non-existent rune
	nonExistent := manager.GetRune("non-existent")
	if nonExistent != nil {
		t.Error("GetRune(non-existent) should return nil")
	}
}

func TestCatalogManagerListByType(t *testing.T) {
	manager := NewCatalogManager()

	agents := manager.ListByType(TypeAgent)
	if agents == nil {
		t.Error("ListByType(agents) returned nil")
	}

	components := manager.ListByType(TypeComponent)
	if components == nil {
		t.Error("ListByType(components) returned nil")
	}

	runes := manager.ListByType(TypeRune)
	if runes == nil {
		t.Error("ListByType(runes) returned nil")
	}

	invalid := manager.ListByType("invalid")
	if invalid != nil {
		t.Error("ListByType(invalid) should return nil")
	}
}

func TestDetectInstalledAgents(t *testing.T) {
	manager := NewCatalogManager()
	installed := manager.DetectInstalledAgents()

	// Should return a slice (possibly empty)
	if installed == nil {
		t.Error("DetectInstalledAgents() returned nil")
	}

	// All returned agents should be valid
	for _, agentID := range installed {
		agent := manager.GetAgent(agentID)
		if agent == nil {
			t.Errorf("DetectInstalledAgents() returned invalid agent ID: %v", agentID)
		}
	}
}

func TestNewCatalogManager(t *testing.T) {
	manager := NewCatalogManager()

	if manager == nil {
		t.Error("NewCatalogManager() returned nil")
	}

	if manager.odinPath == "" {
		t.Error("NewCatalogManager().odinPath is empty")
	}
}

func TestDefaultCatalogManager(t *testing.T) {
	manager := DefaultCatalogManager()

	if manager == nil {
		t.Error("DefaultCatalogManager() returned nil")
	}
}

func TestJoinTags(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"tag1"}, "tag1"},
		{[]string{"tag1", "tag2"}, "tag1, tag2"},
		{[]string{"a", "b", "c"}, "a, b, c"},
	}

	for _, tt := range tests {
		got := joinTags(tt.input)
		if got != tt.expected {
			t.Errorf("joinTags(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestJoinSlice(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{}, "none"},
		{[]string{"item1"}, "item1"},
		{[]string{"item1", "item2"}, "item1, item2"},
	}

	for _, tt := range tests {
		got := joinSlice(tt.input)
		if got != tt.expected {
			t.Errorf("joinSlice(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
