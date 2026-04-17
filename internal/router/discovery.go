package router

import (
	"fmt"
	"sync"

	"github.com/odin-ai/odin/internal/config"
)

// Detector is the interface that all tool-specific detectors must implement
type Detector interface {
	Name() string
	Detect() (*config.DiscoveryResult, error)
}

// DiscoveryService manages multiple detectors and aggregates results
type DiscoveryService struct {
	detectors []Detector
}

// NewDiscoveryService creates a new discovery service with default detectors
func NewDiscoveryService() *DiscoveryService {
	return &DiscoveryService{
		detectors: []Detector{
			&CursorDetector{},
			&VSCodeDetector{},
			&OpenCodeDetector{},
			&WindsurfDetector{},
		},
	}
}

// DiscoverAll executes all registered detectors and returns the results
func (s *DiscoveryService) DiscoverAll() ([]*config.DiscoveryResult, []error) {
	var results []*config.DiscoveryResult
	var errors []error
	var mu sync.Mutex

	var wg sync.WaitGroup
	for _, d := range s.detectors {
		wg.Add(1)
		go func(d Detector) {
			defer wg.Done()
			res, err := d.Detect()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errors = append(errors, fmt.Errorf("%s: %w", d.Name(), err))
				return
			}
			if res != nil {
				results = append(results, res)
			}
		}(d)
	}
	wg.Wait()

	return results, errors
}

