// Package verify provides Nornir - the verification suite for ODIN
package verify

import (
	"sort"
	"sync"
	"time"
)

// BenchMarker handles latency benchmarking
type BenchMarker struct {
	mu         sync.RWMutex
	Thresholds map[string]float64 // phase -> threshold in ms
	iterations int
}

// NewBenchMarker creates a new benchmark runner
func NewBenchMarker() *BenchMarker {
	return &BenchMarker{
		Thresholds: map[string]float64{
			"init":        100, // 100ms for init phase
			"store":       50,  // 50ms for store operations
			"search":      200, // 200ms for search
			"router":      150, // 150ms for routing decisions
			"auth":        100, // 100ms for auth
			"sync":        500, // 500ms for sync operations
			"config_load": 50,  // 50ms for config loading
		},
		iterations: 100, // Default iterations for statistical significance
	}
}

// SetIterations sets the number of iterations for benchmarks
func (bm *BenchMarker) SetIterations(iterations int) {
	if iterations > 0 {
		bm.iterations = iterations
	}
}

// SetThreshold sets the latency threshold for a phase
func (bm *BenchMarker) SetThreshold(phase string, thresholdMs float64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.Thresholds[phase] = thresholdMs
}

// BenchmarkPhase benchmarks a specific phase
func (bm *BenchMarker) BenchmarkPhase(phase string, fn func() error) BenchmarkResult {
	bm.mu.RLock()
	threshold := bm.Thresholds[phase]
	bm.mu.RUnlock()

	iterations := bm.iterations
	if iterations < 1 {
		iterations = 1
	}

	measurements := make([]float64, 0, iterations)
	var minLatency float64 = -1
	var maxLatency float64 = 0
	var totalLatency float64 = 0

	for i := 0; i < iterations; i++ {
		start := time.Now()
		err := fn()
		elapsed := time.Since(start)

		latencyMs := float64(elapsed.Seconds() * 1000)
		measurements = append(measurements, latencyMs)

		totalLatency += latencyMs
		if minLatency < 0 || latencyMs < minLatency {
			minLatency = latencyMs
		}
		if latencyMs > maxLatency {
			maxLatency = latencyMs
		}

		if err != nil {
			// If the function errors, record but don't continue
			break
		}
	}

	// Sort for percentile calculations
	sort.Float64s(measurements)

	avgLatency := float64(0)
	if len(measurements) > 0 {
		avgLatency = totalLatency / float64(len(measurements))
	}

	p50 := float64(0)
	p95 := float64(0)
	p99 := float64(0)

	if len(measurements) > 0 {
		p50 = measurements[len(measurements)/2]
		p95Index := int(float64(len(measurements)) * 0.95)
		if p95Index >= len(measurements) {
			p95Index = len(measurements) - 1
		}
		p95 = measurements[p95Index]

		p99Index := int(float64(len(measurements)) * 0.99)
		if p99Index >= len(measurements) {
			p99Index = len(measurements) - 1
		}
		p99 = measurements[p99Index]
	}

	return BenchmarkResult{
		Phase:        phase,
		LatencyMs:    avgLatency,
		ThresholdMs:  threshold,
		Passed:       avgLatency <= threshold,
		Iterations:   iterations,
		MinLatencyMs: minLatency,
		MaxLatencyMs: maxLatency,
		AvgLatencyMs: avgLatency,
		P50LatencyMs: p50,
		P95LatencyMs: p95,
		P99LatencyMs: p99,
	}
}

// RunAllBenchmarks runs all built-in benchmarks
func (bm *BenchMarker) RunAllBenchmarks() []BenchmarkResult {
	results := make([]BenchmarkResult, 0, len(bm.Thresholds))

	// Benchmark config load (simulated)
	results = append(results, bm.BenchmarkPhase("config_load", func() error {
		time.Sleep(time.Millisecond * 10) // Simulate work
		return nil
	}))

	// Benchmark store (simulated)
	results = append(results, bm.BenchmarkPhase("store", func() error {
		time.Sleep(time.Millisecond * 20) // Simulate work
		return nil
	}))

	// Benchmark search (simulated)
	results = append(results, bm.BenchmarkPhase("search", func() error {
		time.Sleep(time.Millisecond * 50) // Simulate work
		return nil
	}))

	// Benchmark router (simulated)
	results = append(results, bm.BenchmarkPhase("router", func() error {
		time.Sleep(time.Millisecond * 30) // Simulate work
		return nil
	}))

	return results
}

// BenchmarkComparison compares two implementations
type BenchmarkComparison struct {
	Phase        string
	BaselineMs   float64
	CandidateMs  float64
	DeltaMs      float64
	DeltaPercent float64
	Passed       bool
}

// CompareBenchmarks compares baseline vs candidate benchmarks
func (bm *BenchMarker) CompareBenchmarks(phase string, baseline, candidate func() error) BenchmarkComparison {
	baselineResult := bm.BenchmarkPhase(phase+"_baseline", baseline)
	candidateResult := bm.BenchmarkPhase(phase+"_candidate", candidate)

	delta := candidateResult.LatencyMs - baselineResult.LatencyMs
	deltaPercent := float64(0)
	if baselineResult.LatencyMs > 0 {
		deltaPercent = (delta / baselineResult.LatencyMs) * 100
	}

	return BenchmarkComparison{
		Phase:        phase,
		BaselineMs:   baselineResult.LatencyMs,
		CandidateMs:  candidateResult.LatencyMs,
		DeltaMs:      delta,
		DeltaPercent: deltaPercent,
		Passed:       candidateResult.LatencyMs <= baselineResult.LatencyMs*1.1, // 10% tolerance
	}
}
