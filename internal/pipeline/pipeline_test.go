// Package pipeline provides the installation pipeline with staged execution and rollback
package pipeline

import (
	"testing"

	"github.com/odin-ai/odin/internal/catalog"
)

func TestNewPipeline(t *testing.T) {
	p := NewPipeline("sdd")

	if p == nil {
		t.Fatal("NewPipeline() returned nil")
	}

	if p.componentID != "sdd" {
		t.Errorf("NewPipeline().componentID = %v, want 'sdd'", p.componentID)
	}

	if len(p.stages) != 5 {
		t.Errorf("NewPipeline().stages has %d stages, want 5", len(p.stages))
	}

	if p.ctx == nil {
		t.Error("NewPipeline().ctx is nil")
	}

	if p.cancel == nil {
		t.Error("NewPipeline().cancel is nil")
	}
}

func TestPipelineStages(t *testing.T) {
	p := NewPipeline("test")

	expectedStages := []Stage{StageDetect, StageBackup, StageInstall, StageVerify, StageCommit}

	if len(p.stages) != len(expectedStages) {
		t.Errorf("Pipeline has %d stages, want %d", len(p.stages), len(expectedStages))
	}

	for i, stage := range p.stages {
		if stage != expectedStages[i] {
			t.Errorf("Pipeline stage[%d] = %v, want %v", i, stage, expectedStages[i])
		}
	}
}

func TestPipelineCancel(t *testing.T) {
	p := NewPipeline("test")

	// Cancel should not panic
	p.Cancel()
}

func TestPipelineGetResults(t *testing.T) {
	p := NewPipeline("test")

	results := p.GetResults()
	if results == nil {
		t.Error("GetResults() returned nil")
	}

	if len(results) != 0 {
		t.Errorf("GetResults() returned %d results, want 0", len(results))
	}
}

func TestPipelineGetBackupPath(t *testing.T) {
	p := NewPipeline("test")

	backupPath := p.GetBackupPath()
	if backupPath != "" {
		t.Errorf("GetBackupPath() = %v, want empty string", backupPath)
	}
}

func TestSystemDetection(t *testing.T) {
	detection := &SystemDetection{
		OS:         "linux",
		Arch:       "amd64",
		Container:  "docker",
		User:       "testuser",
		HomeDir:    "/home/testuser",
		Agents:     []catalog.AgentID{catalog.AgentClaudeCode},
		CanInstall: true,
	}

	if detection.OS != "linux" {
		t.Errorf("Detection.OS = %v, want 'linux'", detection.OS)
	}

	if detection.Arch != "amd64" {
		t.Errorf("Detection.Arch = %v, want 'amd64'", detection.Arch)
	}

	if len(detection.Agents) != 1 {
		t.Errorf("Detection.Agents has %d agents, want 1", len(detection.Agents))
	}

	if !detection.CanInstall {
		t.Error("Detection.CanInstall = false, want true")
	}
}

func TestStageResult(t *testing.T) {
	result := StageResult{
		Stage:   StageDetect,
		Success: true,
		Output:  "test output",
	}

	if result.Stage != StageDetect {
		t.Errorf("StageResult.Stage = %v, want StageDetect", result.Stage)
	}

	if !result.Success {
		t.Error("StageResult.Success = false, want true")
	}

	if result.Output != "test output" {
		t.Errorf("StageResult.Output = %v, want 'test output'", result.Output)
	}
}

func TestHasRune(t *testing.T) {
	// sdd-propose should exist in catalog
	if !HasRune("sdd-propose") {
		t.Error("HasRune(sdd-propose) = false, want true")
	}

	// non-existent rune
	if HasRune("non-existent-rune") {
		t.Error("HasRune(non-existent-rune) = true, want false")
	}
}

func TestHasComponent(t *testing.T) {
	// sdd should exist in catalog
	if !HasComponent("sdd") {
		t.Error("HasComponent(sdd) = false, want true")
	}

	// non-existent component
	if HasComponent("non-existent-component") {
		t.Error("HasComponent(non-existent-component) = true, want false")
	}
}

func TestDetectContainer(t *testing.T) {
	container := detectContainer()

	// Should return one of the known container types or empty string
	validContainers := map[string]bool{
		"":           true,
		"docker":     true,
		"podman":     true,
		"kubernetes": true,
	}

	if !validContainers[container] {
		t.Errorf("detectContainer() = %v, want a valid container type", container)
	}
}

func TestPipelineRunWithCancel(t *testing.T) {
	p := NewPipeline("non-existent-component")

	// Cancel immediately
	p.Cancel()

	// Run should handle cancellation gracefully
	err := p.Run()
	if err == nil {
		t.Error("Run() should return error after cancel")
	}
}
