// Package memory provides the Mimir memory engine for ODIN
package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/odin-ai/odin/pkg/logger"
)

// Pruner handles automatic pruning of old memories
type Pruner struct {
	db              *DB
	keepTags        []string
	interval        time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.Mutex
	isRunning       bool
	lastPrune       time.Time
	memoriesDeleted int
}

// NewPruner creates a new Pruner
func NewPruner(db *DB, keepTags []string, interval time.Duration) *Pruner {
	if keepTags == nil {
		keepTags = []string{"arch", "spec", "security"}
	}
	if interval <= 0 {
		interval = 24 * time.Hour
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Pruner{
		db:              db,
		keepTags:        keepTags,
		interval:        interval,
		ctx:             ctx,
		cancel:          cancel,
		isRunning:       false,
		memoriesDeleted: 0,
	}
}

// Start begins the automatic pruning background goroutine
func (p *Pruner) Start() {
	p.mu.Lock()
	if p.isRunning {
		p.mu.Unlock()
		return
	}
	p.isRunning = true
	p.mu.Unlock()

	p.wg.Add(1)
	go p.run()

	logger.Info("Pruner started", "interval", p.interval, "keep_tags", p.keepTags)
}

// Stop stops the automatic pruning background goroutine
func (p *Pruner) Stop() {
	p.mu.Lock()
	if !p.isRunning {
		p.mu.Unlock()
		return
	}
	p.isRunning = false
	p.mu.Unlock()

	p.cancel()
	p.wg.Wait()

	logger.Info("Pruner stopped", "total_deleted", p.memoriesDeleted)
}

// run is the main pruning loop
func (p *Pruner) run() {
	defer p.wg.Done()

	// Run immediately on start
	p.prune()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.prune()
		}
	}
}

// prune performs the actual pruning
func (p *Pruner) prune() {
	p.mu.Lock()
	p.lastPrune = time.Now()
	p.mu.Unlock()

	count, err := p.db.PruneMemories(p.keepTags)
	if err != nil {
		logger.Error("Pruning failed", "error", err)
		return
	}

	if count > 0 {
		p.mu.Lock()
		p.memoriesDeleted += count
		p.mu.Unlock()
		logger.Info("Pruning completed", "deleted", count, "keep_tags", p.keepTags)
	}
}

// PruneNow performs pruning immediately and returns the count
func (p *Pruner) PruneNow() (int, error) {
	count, err := p.db.PruneMemories(p.keepTags)
	if err != nil {
		return 0, fmt.Errorf("failed to prune: %w", err)
	}

	if count > 0 {
		p.mu.Lock()
		p.memoriesDeleted += count
		p.mu.Unlock()
	}

	return count, nil
}

// SetKeepTags updates the tags to keep during pruning
func (p *Pruner) SetKeepTags(tags []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.keepTags = tags
	logger.Info("Pruner keep tags updated", "keep_tags", tags)
}

// SetInterval updates the pruning interval
func (p *Pruner) SetInterval(interval time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.interval = interval
	logger.Info("Pruner interval updated", "interval", interval)
}

// GetStats returns pruning statistics
func (p *Pruner) GetStats() *PrunerStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return &PrunerStats{
		IsRunning:    p.isRunning,
		Interval:     p.interval,
		LastPrune:    p.lastPrune,
		TotalDeleted: p.memoriesDeleted,
		KeepTags:     p.keepTags,
		NextPrune:    p.lastPrune.Add(p.interval),
	}
}

// PrunerStats holds statistics about the pruner
type PrunerStats struct {
	IsRunning    bool
	Interval     time.Duration
	LastPrune    time.Time
	TotalDeleted int
	KeepTags     []string
	NextPrune    time.Time
}

// SmartPruner implements intelligent pruning based on usage patterns
type SmartPruner struct {
	db       *DB
	baseTags []string
	ageLimit time.Duration
	minScore float64
}

// NewSmartPruner creates a new SmartPruner
func NewSmartPruner(db *DB, baseTags []string, ageLimit time.Duration) *SmartPruner {
	if baseTags == nil {
		baseTags = []string{"arch", "spec", "security"}
	}
	if ageLimit <= 0 {
		ageLimit = 30 * 24 * time.Hour // 30 days default
	}

	return &SmartPruner{
		db:       db,
		baseTags: baseTags,
		ageLimit: ageLimit,
		minScore: 0.1,
	}
}

// ShouldPrune determines if a memory should be pruned based on multiple factors
func (sp *SmartPruner) ShouldPrune(m *Memory) (bool, string) {
	// Always keep if has any protected tag
	for _, tag := range m.Tags {
		for _, protected := range sp.baseTags {
			if strings.EqualFold(strings.TrimSpace(tag), protected) {
				return false, fmt.Sprintf("has protected tag: %s", tag)
			}
		}
	}

	// Check age limit
	if time.Since(m.AccessedAt) < sp.ageLimit {
		return false, "recently accessed"
	}

	// Check if encrypted (encrypted memories are kept)
	if m.Encrypted {
		return false, "is encrypted"
	}

	// Check metadata for importance score
	if m.Metadata != nil {
		if score, ok := m.Metadata["importance"].(float64); ok {
			if score > sp.minScore {
				return false, fmt.Sprintf("high importance score: %.2f", score)
			}
		}
	}

	// Default: can be pruned
	return true, "no protected tags, old, low importance"
}

// PruneSmart performs intelligent pruning
func (sp *SmartPruner) PruneSmart() (int, error) {
	memories, err := sp.db.ListMemories("")
	if err != nil {
		return 0, fmt.Errorf("failed to list memories: %w", err)
	}

	toDelete := make([]string, 0)
	for _, m := range memories {
		shouldPrune, reason := sp.ShouldPrune(m)
		if shouldPrune {
			toDelete = append(toDelete, m.ID)
			logger.Debug("Smart prune candidate", "id", m.ID, "reason", reason)
		}
	}

	// Delete candidates
	count := 0
	for _, id := range toDelete {
		if err := sp.db.DeleteMemory(id); err != nil {
			logger.Warn("Failed to delete memory during smart prune", "id", id, "error", err)
			continue
		}
		count++
	}

	return count, nil
}

// PruningReport contains detailed information about pruning results
type PruningReport struct {
	TotalScanned      int
	TotalPruned       int
	ProtectedMemories []MemorySummary
	PrunedMemories    []MemorySummary
	Duration          time.Duration
	Timestamp         time.Time
}

// MemorySummary contains basic memory info for reports
type MemorySummary struct {
	ID        string
	Tags      []string
	Age       time.Duration
	Accessed  time.Time
	Protected bool
	Reason    string
}

// GenerateReport generates a detailed pruning report
func (sp *SmartPruner) GenerateReport() (*PruningReport, error) {
	memories, err := sp.db.ListMemories("")
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}

	report := &PruningReport{
		TotalScanned:      len(memories),
		ProtectedMemories: make([]MemorySummary, 0),
		PrunedMemories:    make([]MemorySummary, 0),
		Timestamp:         time.Now(),
	}

	for _, m := range memories {
		shouldPrune, reason := sp.ShouldPrune(m)

		summary := MemorySummary{
			ID:       m.ID,
			Tags:     m.Tags,
			Age:      time.Since(m.CreatedAt),
			Accessed: m.AccessedAt,
			Reason:   reason,
		}

		if shouldPrune {
			summary.Protected = false
			report.PrunedMemories = append(report.PrunedMemories, summary)
		} else {
			summary.Protected = true
			report.ProtectedMemories = append(report.ProtectedMemories, summary)
		}
	}

	report.TotalPruned = len(report.PrunedMemories)
	return report, nil
}

// PruningDryRun performs a dry run of pruning without deleting
func (sp *SmartPruner) PruningDryRun() (*PruningReport, error) {
	report, err := sp.GenerateReport()
	if err != nil {
		return nil, err
	}

	logger.Info("Pruning dry run",
		"would_prune", report.TotalPruned,
		"would_keep", len(report.ProtectedMemories))

	return report, nil
}
