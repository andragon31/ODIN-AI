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

// AllStages returns all pipeline stages in execution order
func AllStages() []Stage {
	return []Stage{StageDetect, StageBackup, StageInstall, StageVerify, StageCommit}
}

// StageResult represents the result of a stage execution
type StageResult struct {
	Stage     Stage
	Success   bool
	Output    string
	Error     error
	Timestamp time.Time
	Duration  time.Duration
}

// NewStageResult creates a new stage result
func NewStageResult(stage Stage, success bool, output string, err error, duration time.Duration) StageResult {
	return StageResult{
		Stage:     stage,
		Success:   success,
		Output:    output,
		Error:     err,
		Timestamp: time.Now(),
		Duration:  duration,
	}
}

// IsSuccess returns true if the stage completed successfully
func (r StageResult) IsSuccess() bool {
	return r.Success && r.Error == nil
}

// Summary returns a one-line summary of the result
func (r StageResult) Summary() string {
	if r.Success {
		return fmt.Sprintf("[%s] ✓ %s (%.2fs)", r.Stage, r.Output, r.Duration.Seconds())
	}
	return fmt.Sprintf("[%s] ✗ %s: %v (%.2fs)", r.Stage, r.Output, r.Error, r.Duration.Seconds())
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

// NewSystemDetection creates a new system detection
func NewSystemDetection() *SystemDetection {
	return &SystemDetection{
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		HomeDir: os.Getenv("HOME"),
	}
}

// GetUser populates user information in the detection
func (d *SystemDetection) GetUser() error {
	if usr, err := user.Current(); err == nil {
		d.User = usr.Username
	}
	return nil
}

// GetAgents detects installed agents
func (d *SystemDetection) GetAgents() {
	catalogManager := catalog.DefaultCatalogManager()
	detected := catalogManager.DetectInstalledAgents()
	d.Agents = make([]catalog.AgentID, len(detected))
	for i, agent := range detected {
		d.Agents[i] = agent.ID
	}
}

// GetComponents populates detected components from .odin directory
func (d *SystemDetection) GetComponents() error {
	if d.HomeDir == "" {
		var err error
		d.HomeDir, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	}

	odinPath := filepath.Join(d.HomeDir, ".odin")
	if entries, err := os.ReadDir(odinPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				d.Components = append(d.Components, entry.Name())
			}
		}
	}
	return nil
}

// GetRunes populates detected runes from .odin/runes directory
func (d *SystemDetection) GetRunes() error {
	if d.HomeDir == "" {
		var err error
		d.HomeDir, err = os.UserHomeDir()
		if err != nil {
			return err
		}
	}

	runesPath := filepath.Join(d.HomeDir, ".odin", "runes")
	if entries, err := os.ReadDir(runesPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				d.Runes = append(d.Runes, entry.Name())
			}
		}
	}
	return nil
}

// SetCanInstall sets the CanInstall flag
func (d *SystemDetection) SetCanInstall(can bool) {
	d.CanInstall = can
}

// Summary returns a human-readable summary of the detection
func (d *SystemDetection) Summary() string {
	return fmt.Sprintf("OS: %s, Arch: %s, Agents: %d, Components: %d, Runes: %d",
		d.OS, d.Arch, len(d.Agents), len(d.Components), len(d.Runes))
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

// NewPipeline creates a new pipeline for installing a component
func NewPipeline(componentID string) *Pipeline {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pipeline{
		componentID: componentID,
		ctx:         ctx,
		cancel:      cancel,
		stages:      AllStages(),
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

	return NewStageResult(stage, err == nil, output, err, duration)
}

// runDetect performs system detection
func (p *Pipeline) runDetect() (string, error) {
	detection := NewSystemDetection()

	if err := detection.GetUser(); err != nil {
		logger.Warn("Failed to get user info", "error", err)
	}

	detection.GetAgents()
	detection.GetComponents()
	detection.GetRunes()
	detection.SetCanInstall(true)

	p.detected = detection

	return detection.Summary(), nil
}

// runBackup creates a backup before installation
func (p *Pipeline) runBackup() (string, error) {
	if p.detected == nil {
		return "", fmt.Errorf("system detection not completed")
	}

	backupPath, err := backup.CreateBackup(filepath.Join(p.detected.HomeDir, ".odin"))
	if err != nil {
		return "", fmt.Errorf("backup failed: %w", err)
	}

	p.backupPath = backupPath
	return fmt.Sprintf("Backup created at: %s", backupPath), nil
}

// runInstall performs the actual installation
func (p *Pipeline) runInstall() (string, error) {
	if p.detected == nil {
		return "", fmt.Errorf("system detection not completed")
	}

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

	// Create a marker file indicating installation
	markerPath := filepath.Join(compPath, ".installed")
	if err := os.WriteFile(markerPath, []byte(fmt.Sprintf("Installed at %s", time.Now().Format(time.RFC3339))), 0644); err != nil {
		logger.Warn("Failed to create marker file", "error", err)
	}

	// Install runes if specified
	runesInstalled := 0
	for _, runeName := range comp.Runes {
		runePath := filepath.Join(odinPath, "runes", runeName)
		if err := os.MkdirAll(runePath, 0755); err != nil {
			logger.Warn("Failed to create rune directory", "rune", runeName, "error", err)
		} else {
			// Copy RUNE.md and rune.yaml from source repo
			if err := p.copyRuneFiles(runeName, runePath); err != nil {
				logger.Warn("Failed to copy rune files", "rune", runeName, "error", err)
			}
			runesInstalled++
		}
	}

	return fmt.Sprintf("Installed %s with %d runes", comp.Name, runesInstalled), nil
}

// runVerify verifies the installation
func (p *Pipeline) runVerify() (string, error) {
	if p.detected == nil {
		return "", fmt.Errorf("system detection not completed")
	}

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
	if p.detected == nil {
		return fmt.Errorf("system detection not completed")
	}

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

// copyRuneFiles copies RUNE.md and rune.yaml from source to target
func (p *Pipeline) copyRuneFiles(runeName string, targetPath string) error {
	// Try to find source rune directory in the repo
	// The source could be in the ODIN_REPO/runes/{runeName} directory
	possibleSources := []string{
		filepath.Join(".", "runes", runeName),
		filepath.Join("..", "runes", runeName),
		filepath.Join(os.Getenv("ODIN_ROOT"), "runes", runeName),
	}

	foundRune := false
	for _, srcPath := range possibleSources {
		runeMdPath := filepath.Join(srcPath, "RUNE.md")
		runeYamlPath := filepath.Join(srcPath, "rune.yaml")

		if _, err := os.Stat(runeMdPath); err == nil {
			// RUNE.md exists in this source
			if err := copyFile(runeMdPath, filepath.Join(targetPath, "RUNE.md")); err == nil {
				foundRune = true
				logger.Debug("Copied RUNE.md", "source", runeMdPath, "target", targetPath)
			}
		}

		if _, err := os.Stat(runeYamlPath); err == nil {
			// rune.yaml exists
			if err := copyFile(runeYamlPath, filepath.Join(targetPath, "rune.yaml")); err == nil {
				logger.Debug("Copied rune.yaml", "source", runeYamlPath, "target", targetPath)
			}
		}

		if foundRune {
			return nil
		}
	}

	if !foundRune {
		// Create minimal RUNE.md if source not found
		defaultContent := "# " + runeName + "\n\n## Purpose\nRune installed via pipeline.\n"
		os.WriteFile(filepath.Join(targetPath, "RUNE.md"), []byte(defaultContent), 0644)
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
