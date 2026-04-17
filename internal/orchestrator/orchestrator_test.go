package orchestrator

import (
	"os"
	"testing"
)

func TestOrchestrator_TriadMethodology(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "odin-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	orch, err := NewOrchestrator(tempDir)
	if err != nil {
		t.Fatalf("Failed to create orchestrator: %v", err)
	}

	// Test 1: Create standard session
	s1, err := orch.CreateSession("fix-bug", "Description 1", "standard")
	if err != nil {
		t.Fatalf("Failed to create standard session: %v", err)
	}
	if s1.Methodology != "standard" {
		t.Errorf("Expected standard methodology, got %s", s1.Methodology)
	}

	// Test 2: Create triad session
	s2, err := orch.CreateSession("new-feature", "Description 2", "triad")
	if err != nil {
		t.Fatalf("Failed to create triad session: %v", err)
	}
	if s2.Methodology != "triad" {
		t.Errorf("Expected triad methodology, got %s", s2.Methodology)
	}

	// Test 3: Default methodology
	s3, err := orch.CreateSession("default-test", "Description 3", "")
	if err != nil {
		t.Fatalf("Failed to create session with empty methodology: %v", err)
	}
	if s3.Methodology != "standard" {
		t.Errorf("Expected default standard methodology, got %s", s3.Methodology)
	}

	// Test 4: Persistence
	loaded, err := orch.GetSession(s2.ID)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}
	if loaded.Methodology != "triad" {
		t.Errorf("Persisted methodology mismatch: expected triad, got %s", loaded.Methodology)
	}
	// Test 5: Pentakill methodology
	s4, err := orch.CreateSession("complex-domain", "Description 4", MethodologyPentakill)
	if err != nil {
		t.Fatalf("Failed to create pentakill session: %v", err)
	}
	if s4.Methodology != MethodologyPentakill {
		t.Errorf("Expected pentakill methodology, got %s", s4.Methodology)
	}

	// Test 6: Phase advancement in Pentakill
	if s4.Phase != PhaseProposal {
		t.Errorf("Expected start phase proposal, got %s", s4.Phase)
	}
	s4, err = orch.AdvancePhase(s4.ID)
	if err != nil {
		t.Fatalf("Failed to advance phase: %v", err)
	}
	if s4.Phase != PhaseDomain {
		t.Errorf("Expected next phase domain, got %s", s4.Phase)
	}
}
