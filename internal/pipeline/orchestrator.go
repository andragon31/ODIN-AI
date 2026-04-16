// Package pipeline provides the installation pipeline with staged execution and rollback
package pipeline

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"

	"github.com/odin-ai/odin/internal/backup"
	"github.com/odin-ai/odin/internal/catalog"
	"github.com/odin-ai/odin/pkg/logger"
)

// Stage represents a pipeline stage
type Stage string

const (
	StageDetect  Stage = "detect"
	StageBackup  Stage = "backup"
	StageInstall Stage = "install"
	StageVerify  Stage = "verify"
	StageCommit  Stage = "commit"
)

// StageResult represents the result of a stage execution
type StageResult struct {
	Stage     Stage
	Success   bool
	Output    string
	Error     error
	Timestamp time.Time
	Duration  time.Duration
}

// Pipeline represents the installation pipeline
type Pipeline struct {
	componentID string
	ctx         context.Context
	cancel      context.CancelFunc
	stages      []Stage
	results     []StageResult
	backupPath  string
	detected    *SystemDetection
}

// SystemDetection contains detected system information
type SystemDetection struct {
	OS         string
	Arch       string
	Container  string
	User       string
	HomeDir    string
	Agents     []catalog.AgentID
	Components []string
	Runes      []string
	CanInstall bool
	Issues     []string
}

// NewPipeline creates a new pipeline for installing a component
func NewPipeline(componentID string) *Pipeline {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pipeline{
		componentID: componentID,
		ctx:         ctx,
		cancel:      cancel,
		stages:      []Stage{StageDetect, StageBackup, StageInstall, StageVerify, StageCommit},
		results:     []StageResult{},
	}
}

// Cancel cancels the pipeline execution
func (p *Pipeline) Cancel() {
	if p.cancel != nil {
		p.cancel()
		logger.Info("Pipeline cancelled by user")
	}
}

// Run executes the pipeline
func (p *Pipeline) Run() error {
	logger.Info("Starting pipeline", "component", p.componentID)

	for _, stage := range p.stages {
		select {
		case <-p.ctx.Done():
			logger.Info("Pipeline interrupted", "stage", stage)
			return p.handleInterrupt()
		default:
		}

		result := p.executeStage(stage)
		p.results = append(p.results, result)

		if !result.Success {
			logger.Error("Stage failed", "stage", stage, "error", result.Error)
			return p.rollback(stage, result.Error)
		}

		logger.Info("Stage completed", "stage", stage, "duration", result.Duration)
	}

	logger.Info("Pipeline completed successfully", "component", p.componentID)
	return nil
}

// executeStage runs a single stage
func (p *Pipeline) executeStage(stage Stage) StageResult {
	start := time.Now()

	var output string
	var err error

	switch stage {
	case StageDetect:
		output, err = p.runDetect()
	case StageBackup:
		output, err = p.runBackup()
	case StageInstall:
		output, err = p.runInstall()
	case StageVerify:
		output, err = p.runVerify()
	case StageCommit:
		output, err = p.runCommit()
	default:
		err = fmt.Errorf("unknown stage: %s", stage)
	}

	duration := time.Since(start)

	return StageResult{
		Stage:     stage,
		Success:   err == nil,
		Output:    output,
		Error:     err,
		Timestamp: start,
		Duration:  duration,
	}
}

// runDetect performs system detection
func (p *Pipeline) runDetect() (string, error) {
	detection := &SystemDetection{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Container: detectContainer(),
		HomeDir:   os.Getenv("HOME"),
	}

	if detection.HomeDir == "" {
		detection.HomeDir, _ = os.UserHomeDir()
	}

	// Get user
	if usr, err := user.Current(); err == nil {
		detection.User = usr.Username
	}

	// Detect installed agents
	catalogManager := catalog.DefaultCatalogManager()
	detection.Agents = catalogManager.DetectInstalledAgents()

	// Check existing components
	odinPath := filepath.Join(detection.HomeDir, ".odin")
	if entries, err := os.ReadDir(odinPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				detection.Components = append(detection.Components, entry.Name())
			}
		}
	}

	// Check installed runes
	runesPath := filepath.Join(odinPath, "runes")
	if entries, err := os.ReadDir(runesPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				detection.Runes = append(detection.Runes, entry.Name())
			}
		}
	}

	detection.CanInstall = true

	p.detected = detection

	output := fmt.Sprintf("OS: %s, Arch: %s, Agents: %d, Components: %d, Runes: %d",
		detection.OS, detection.Arch, len(detection.Agents), len(detection.Components), len(detection.Runes))

	return output, nil
}

// runBackup creates a backup before installation
func (p *Pipeline) runBackup() (string, error) {
	backupPath, err := backup.CreateBackup(filepath.Join(p.detected.HomeDir, ".odin"))
	if err != nil {
		return "", fmt.Errorf("backup failed: %w", err)
	}

	p.backupPath = backupPath
	return fmt.Sprintf("Backup created at: %s", backupPath), nil
}

// runInstall performs the actual installation
func (p *Pipeline) runInstall() (string, error) {
	comp := catalog.DefaultCatalogManager().GetComponent(p.componentID)
	if comp == nil {
		return "", fmt.Errorf("component %s not found", p.componentID)
	}

	odinPath := filepath.Join(p.detected.HomeDir, ".odin")
	compPath := filepath.Join(odinPath, p.componentID)

	// Create component directory
	if err := os.MkdirAll(compPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create component directory: %w", err)
	}

	// Install runes if specified
	for _, runeName := range comp.Runes {
		runePath := filepath.Join(odinPath, "runes", runeName)
		if err := os.MkdirAll(runePath, 0755); err != nil {
			logger.Warn("Failed to create rune directory", "rune", runeName, "error", err)
		}
	}

	return fmt.Sprintf("Installed %s with %d runes", comp.Name, len(comp.Runes)), nil
}

// runVerify verifies the installation
func (p *Pipeline) runVerify() (string, error) {
	odinPath := filepath.Join(p.detected.HomeDir, ".odin")
	compPath := filepath.Join(odinPath, p.componentID)

	// Check component directory exists
	if _, err := os.Stat(compPath); err != nil {
		return "", fmt.Errorf("component directory not found: %w", err)
	}

	return fmt.Sprintf("Verification passed for %s", p.componentID), nil
}

// runCommit finalizes the installation
func (p *Pipeline) runCommit() (string, error) {
	return fmt.Sprintf("Installation committed for %s", p.componentID), nil
}

// rollback performs rollback on failure
func (p *Pipeline) rollback(failedStage Stage, originalError error) error {
	logger.Info("Starting rollback", "failed_stage", failedStage)

	// Find the index of the failed stage
	failedIdx := -1
	for i, stage := range p.stages {
		if stage == failedStage {
			failedIdx = i
			break
		}
	}

	// Rollback in reverse order, only for stages that completed
	for i := failedIdx - 1; i >= 0; i-- {
		stage := p.stages[i]
		logger.Info("Rolling back stage", "stage", stage)

		switch stage {
		case StageInstall:
			if err := p.rollbackInstall(); err != nil {
				logger.Error("Rollback install failed", "error", err)
			}
		case StageBackup:
			if p.backupPath != "" {
				if err := backup.RestoreBackup(p.backupPath); err != nil {
					logger.Error("Rollback backup failed", "error", err)
				}
			}
		}
	}

	return fmt.Errorf("pipeline failed at %s: %w", failedStage, originalError)
}

// rollbackInstall reverses the install stage
func (p *Pipeline) rollbackInstall() error {
	odinPath := filepath.Join(p.detected.HomeDir, ".odin")
	compPath := filepath.Join(odinPath, p.componentID)

	// Remove component directory
	if err := os.RemoveAll(compPath); err != nil {
		return fmt.Errorf("failed to remove component directory: %w", err)
	}

	return nil
}

// handleInterrupt handles pipeline interruption (Ctrl+C)
func (p *Pipeline) handleInterrupt() error {
	logger.Info("Pipeline interrupted, initiating rollback")

	// Rollback any completed stages
	for i := len(p.results) - 1; i >= 0; i-- {
		result := p.results[i]
		if !result.Success {
			break
		}

		switch result.Stage {
		case StageInstall:
			p.rollbackInstall()
		case StageBackup:
			if p.backupPath != "" {
				backup.RestoreBackup(p.backupPath)
			}
		}
	}

	return fmt.Errorf("pipeline interrupted")
}

// GetResults returns all stage results
func (p *Pipeline) GetResults() []StageResult {
	return p.results
}

// GetBackupPath returns the backup path
func (p *Pipeline) GetBackupPath() string {
	return p.backupPath
}

// GetSystemDetection returns the system detection data
func (p *Pipeline) GetSystemDetection() *SystemDetection {
	return p.detected
}

// detectContainer detects if running in a container
func detectContainer() string {
	// Check for common container indicators
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "docker"
	}

	// Check for container environment variables
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return "kubernetes"
	}

	if os.Getenv("DOCKER_CONTAINER") == "true" {
		return "docker"
	}

	return ""
}

// HasRune checks if a rune is available in the catalog
func HasRune(name string) bool {
	return catalog.DefaultCatalogManager().GetRune(name) != nil
}

// HasComponent checks if a component is available in the catalog
func HasComponent(name string) bool {
	return catalog.DefaultCatalogManager().GetComponent(name) != nil
}
