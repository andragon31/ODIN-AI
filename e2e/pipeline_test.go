// Package e2e provides end-to-end tests for ODIN components
//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/odin-ai/odin/internal/pipeline"
)

// TestPipelineE2E tests the pipeline orchestration end-to-end
func TestPipelineE2E(t *testing.T) {
	t.Run("CreatePipeline", func(t *testing.T) {
		p := pipeline.NewPipeline("test-component")

		if p == nil {
			t.Fatal("Expected non-nil pipeline")
		}

		results := p.GetResults()
		if results == nil {
			t.Error("Expected non-nil results slice")
		}
	})

	t.Run("SystemDetection", func(t *testing.T) {
		detection := pipeline.NewSystemDetection()

		if detection == nil {
			t.Fatal("Expected non-nil system detection")
		}

		if detection.OS == "" {
			t.Error("Expected OS to be set")
		}

		t.Logf("Detected OS: %s, Arch: %s", detection.OS, detection.Arch)
	})

	t.Run("StageResults", func(t *testing.T) {
		result := pipeline.NewStageResult(
			pipeline.StageDetect,
			true,
			"Detection successful",
			nil,
			0,
		)

		if !result.IsSuccess() {
			t.Error("Expected success result")
		}

		summary := result.Summary()
		if summary == "" {
			t.Error("Expected non-empty summary")
		}

		t.Logf("Result summary: %s", summary)
	})

	t.Run("AllStages", func(t *testing.T) {
		stages := pipeline.AllStages()

		if len(stages) != 5 {
			t.Errorf("Expected 5 stages, got %d", len(stages))
		}

		expected := []pipeline.Stage{
			pipeline.StageDetect,
			pipeline.StageBackup,
			pipeline.StageInstall,
			pipeline.StageVerify,
			pipeline.StageCommit,
		}

		for i, expectedStage := range expected {
			if stages[i] != expectedStage {
				t.Errorf("Expected stage %d to be %v, got %v", i, expectedStage, stages[i])
			}
		}
	})

	t.Run("PipelineCancellation", func(t *testing.T) {
		p := pipeline.NewPipeline("test-component")
		p.Cancel()

		// Cancellation should not panic
		t.Log("Pipeline cancelled successfully")
	})
}

// TestPipelineStages tests individual stage behavior
func TestPipelineStages(t *testing.T) {
	t.Run("StageConstants", func(t *testing.T) {
		if pipeline.StageDetect != "detect" {
			t.Errorf("Expected StageDetect to be 'detect', got '%s'", pipeline.StageDetect)
		}
		if pipeline.StageBackup != "backup" {
			t.Errorf("Expected StageBackup to be 'backup', got '%s'", pipeline.StageBackup)
		}
		if pipeline.StageInstall != "install" {
			t.Errorf("Expected StageInstall to be 'install', got '%s'", pipeline.StageInstall)
		}
		if pipeline.StageVerify != "verify" {
			t.Errorf("Expected StageVerify to be 'verify', got '%s'", pipeline.StageVerify)
		}
		if pipeline.StageCommit != "commit" {
			t.Errorf("Expected StageCommit to be 'commit', got '%s'", pipeline.StageCommit)
		}
	})
}
