package router

import (
	"fmt"
	"sync"
	"time"
)

// ProviderMetrics holds metrics for a single provider
type ProviderMetrics struct {
	RequestCount  int64         `json:"request_count"`
	SuccessCount  int64         `json:"success_count"`
	ErrorCount    int64         `json:"error_count"`
	TotalLatency  time.Duration `json:"total_latency"`
	TotalTokens   int64         `json:"total_tokens"`
	TotalCost     float64       `json:"total_cost"`
	LastUsed      time.Time     `json:"last_used"`
	LastError     string        `json:"last_error,omitempty"`
	LastErrorTime time.Time     `json:"last_error_time,omitempty"`
	mu            sync.RWMutex
}

// Metrics holds router-wide metrics
type Metrics struct {
	providers map[string]*ProviderMetrics
	mu        sync.RWMutex
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		providers: make(map[string]*ProviderMetrics),
	}
}

// GetProviderMetrics returns metrics for a specific provider
func (m *Metrics) GetProviderMetrics(providerName string) *ProviderMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if metrics, ok := m.providers[providerName]; ok {
		return metrics
	}
	return nil
}

// GetAllMetrics returns all metrics
func (m *Metrics) GetAllMetrics() map[string]*ProviderMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*ProviderMetrics)
	for k, v := range m.providers {
		v.mu.RLock()
		result[k] = &ProviderMetrics{
			RequestCount:  v.RequestCount,
			SuccessCount:  v.SuccessCount,
			ErrorCount:    v.ErrorCount,
			TotalLatency:  v.TotalLatency,
			TotalTokens:   v.TotalTokens,
			TotalCost:     v.TotalCost,
			LastUsed:      v.LastUsed,
			LastError:     v.LastError,
			LastErrorTime: v.LastErrorTime,
		}
		v.mu.RUnlock()
	}
	return result
}

// RecordSuccess records a successful request
func (m *Metrics) RecordSuccess(providerName string, latency time.Duration, tokens int) {
	m.mu.Lock()
	metrics, ok := m.providers[providerName]
	if !ok {
		metrics = &ProviderMetrics{}
		m.providers[providerName] = metrics
	}
	metrics.RequestCount++
	metrics.SuccessCount++
	metrics.TotalLatency += latency
	metrics.TotalTokens += int64(tokens)
	metrics.LastUsed = time.Now()
	m.mu.Unlock()
}

// RecordError records a failed request
func (m *Metrics) RecordError(providerName string) {
	m.mu.Lock()
	metrics, ok := m.providers[providerName]
	if !ok {
		metrics = &ProviderMetrics{}
		m.providers[providerName] = metrics
	}
	metrics.RequestCount++
	metrics.ErrorCount++
	m.mu.Unlock()
}

// RecordCost records the cost of a request
func (m *Metrics) RecordCost(providerName string, cost float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	metrics, ok := m.providers[providerName]
	if !ok {
		metrics = &ProviderMetrics{}
		m.providers[providerName] = metrics
	}
	metrics.TotalCost += cost
}

// RecordLastError records the last error for a provider
func (m *Metrics) RecordLastError(providerName string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	metrics, ok := m.providers[providerName]
	if !ok {
		metrics = &ProviderMetrics{}
		m.providers[providerName] = metrics
	}
	metrics.LastError = err.Error()
	metrics.LastErrorTime = time.Now()
}

// SuccessRate returns the success rate for a provider (0-100)
func (m *Metrics) SuccessRate(providerName string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics, ok := m.providers[providerName]
	if !ok || metrics.RequestCount == 0 {
		return 0
	}
	return float64(metrics.SuccessCount) / float64(metrics.RequestCount) * 100
}

// AverageLatency returns the average latency for a provider
func (m *Metrics) AverageLatency(providerName string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	metrics, ok := m.providers[providerName]
	if !ok || metrics.RequestCount == 0 {
		return 0
	}
	return metrics.TotalLatency / time.Duration(metrics.RequestCount)
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers = make(map[string]*ProviderMetrics)
}

// FormatMetricsTable formats metrics as a table string
func (m *Metrics) FormatMetricsTable() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.providers) == 0 {
		return "No metrics available"
	}

	header := fmt.Sprintf("%-15s %10s %10s %10s %15s %12s %10s\n",
		"Provider", "Requests", "Success", "Errors", "Avg Latency", "Success %", "Total Cost")
	separator := fmt.Sprintf("%s\n", "-------------------------------------------------------------------")

	result := header + separator

	for name, metrics := range m.providers {
		metrics.mu.RLock()
		avgLatency := time.Duration(0)
		if metrics.RequestCount > 0 {
			avgLatency = metrics.TotalLatency / time.Duration(metrics.RequestCount)
		}
		successRate := float64(0)
		if metrics.RequestCount > 0 {
			successRate = float64(metrics.SuccessCount) / float64(metrics.RequestCount) * 100
		}
		result += fmt.Sprintf("%-15s %10d %10d %10d %15s %11.1f%% %9.4f\n",
			name, metrics.RequestCount, metrics.SuccessCount, metrics.ErrorCount,
			avgLatency.String(), successRate, metrics.TotalCost)
		metrics.mu.RUnlock()
	}

	return result
}
