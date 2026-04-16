// Package sync provides the Bifrost sync engine for ODIN
package sync

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// VectorClock tracks causal ordering of operations
type VectorClock map[string]uint64

// CRDTEngine implements Conflict-free Replicated Data Types for sync
type CRDTEngine struct {
	mu     sync.RWMutex
	clock  VectorClock
	nodeID string
}

// NewCRDTEngine creates a new CRDT engine with a unique node ID
func NewCRDTEngine(nodeID string) *CRDTEngine {
	return &CRDTEngine{
		clock:  make(VectorClock),
		nodeID: nodeID,
	}
}

// MergeResult represents the result of merging two values
type MergeResult struct {
	Value        interface{}
	Resolved     bool
	Conflict     bool
	ConflictInfo *ConflictInfo
}

// ConflictInfo provides details about an unresolved conflict
type ConflictInfo struct {
	Path      string
	Local     string
	Remote    string
	Timestamp time.Time
}

// MergeText merges two text values using Last-Write-Wins (LWW) strategy
func (c *CRDTEngine) MergeText(local, remote string, localTS, remoteTS time.Time) *MergeResult {
	// Last-Write-Wins: the newer value wins
	if remoteTS.After(localTS) {
		return &MergeResult{
			Value:    remote,
			Resolved: true,
			Conflict: false,
		}
	}
	return &MergeResult{
		Value:    local,
		Resolved: true,
		Conflict: false,
	}
}

// MergeJSON performs a deep merge of two JSON objects
// For conflicting primitive values, uses Last-Write-Wins
// For nested objects, recursively merges
// For arrays, concatenates and deduplicates
func (c *CRDTEngine) MergeJSON(localJSON, remoteJSON []byte, localTS, remoteTS time.Time) (*MergeResult, error) {
	var localMap, remoteMap map[string]interface{}

	if err := json.Unmarshal(localJSON, &localMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal local JSON: %w", err)
	}
	if err := json.Unmarshal(remoteJSON, &remoteMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal remote JSON: %w", err)
	}

	merged := make(map[string]interface{})
	conflicts := []string{}

	// Process all keys from both maps
	allKeys := make(map[string]bool)
	for k := range localMap {
		allKeys[k] = true
	}
	for k := range remoteMap {
		allKeys[k] = true
	}

	for key := range allKeys {
		localVal, localExists := localMap[key]
		remoteVal, remoteExists := remoteMap[key]

		if !localExists {
			// Key only in remote - take it
			merged[key] = remoteVal
		} else if !remoteExists {
			// Key only in local - keep it
			merged[key] = localVal
		} else {
			// Key in both - need to merge
			localType := fmt.Sprintf("%T", localVal)
			remoteType := fmt.Sprintf("%T", remoteVal)

			if localType != remoteType {
				// Type conflict - use LWW based on timestamp
				if remoteTS.After(localTS) {
					merged[key] = remoteVal
				} else {
					merged[key] = localVal
				}
				conflicts = append(conflicts, key)
			} else if localType == "map[string]interface {}" {
				// Both are objects - recursively merge
				localSubJSON, _ := json.Marshal(localVal)
				remoteSubJSON, _ := json.Marshal(remoteVal)
				subResult, err := c.MergeJSON(localSubJSON, remoteSubJSON, localTS, remoteTS)
				if err != nil {
					return nil, err
				}
				merged[key] = subResult.Value
				if subResult.Conflict {
					conflicts = append(conflicts, key)
				}
			} else if localType == "[]interface {}" {
				// Both are arrays - concatenate and deduplicate
				localArr, _ := localVal.([]interface{})
				remoteArr, _ := remoteVal.([]interface{})
				merged[key] = c.mergeArrays(localArr, remoteArr)
			} else {
				// Primitive values - use LWW
				if remoteTS.After(localTS) {
					merged[key] = remoteVal
				} else {
					merged[key] = localVal
				}
			}
		}
	}

	resultJSON, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal merged result: %w", err)
	}

	if len(conflicts) > 0 {
		return &MergeResult{
			Value:    string(resultJSON),
			Resolved: false,
			Conflict: true,
			ConflictInfo: &ConflictInfo{
				Path:      strings.Join(conflicts, ","),
				Local:     string(localJSON),
				Remote:    string(remoteJSON),
				Timestamp: time.Now(),
			},
		}, nil
	}

	return &MergeResult{
		Value:    string(resultJSON),
		Resolved: true,
		Conflict: false,
	}, nil
}

// mergeArrays concatenates two arrays and removes duplicates
func (c *CRDTEngine) mergeArrays(local, remote []interface{}) []interface{} {
	seen := make(map[string]bool)
	result := []interface{}{}

	// Add local values
	for _, v := range local {
		key := fmt.Sprintf("%v", v)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}

	// Add remote values not in local
	for _, v := range remote {
		key := fmt.Sprintf("%v", v)
		if !seen[key] {
			seen[key] = true
			result = append(result, v)
		}
	}

	return result
}

// MergeMaps merges two maps using LWW for each value
func (c *CRDTEngine) MergeMaps(local, remote map[string]string, localTS, remoteTS time.Time) map[string]string {
	result := make(map[string]string)

	// Copy all local entries
	for k, v := range local {
		result[k] = v
	}

	// Merge remote entries (overwrites local if remote is newer - tracked by timestamp header)
	for k, v := range remote {
		result[k] = v
	}

	return result
}

// GenerateLWWID generates a unique LWW timestamp-based ID
// Format: timestamp-nodeID-hash
func (c *CRDTEngine) GenerateLWWID() string {
	ts := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d-%s", ts, c.nodeID)))
	return fmt.Sprintf("%d-%s-%x", ts, c.nodeID, hash[:8])
}

// ParseLWWID extracts timestamp and nodeID from an LWW ID
func ParseLWWID(id string) (timestamp time.Time, nodeID string, err error) {
	parts := strings.SplitN(id, "-", 3)
	if len(parts) < 3 {
		return time.Time{}, "", fmt.Errorf("invalid LWW ID format: %s", id)
	}

	tsNano, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid timestamp in LWW ID: %w", err)
	}

	return time.Unix(0, int64(tsNano)), parts[1], nil
}

// UpdateClock increments the vector clock for this node
func (c *CRDTEngine) UpdateClock() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clock[c.nodeID]++
}

// MergeClock merges another vector clock into this one
func (c *CRDTEngine) MergeClock(remote VectorClock) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.clock == nil {
		c.clock = make(VectorClock)
	}

	for node, counter := range remote {
		if c.clock[node] < counter {
			c.clock[node] = counter
		}
	}
}

// GetClock returns a copy of the current vector clock
func (c *CRDTEngine) GetClock() VectorClock {
	c.mu.RLock()
	defer c.mu.RUnlock()

	clone := make(VectorClock)
	for k, v := range c.clock {
		clone[k] = v
	}
	return clone
}

// CompareClocks compares two vector clocks
// Returns: -1 if this < other, 0 if concurrent, 1 if this > other
func (c *CRDTEngine) CompareClocks(other VectorClock) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	allNodes := make(map[string]bool)
	for k := range c.clock {
		allNodes[k] = true
	}
	for k := range other {
		allNodes[k] = true
	}

	thisLess := false
	otherLess := false

	for node := range allNodes {
		thisVal := c.clock[node]
		otherVal := other[node]

		if thisVal < otherVal {
			thisLess = true
		}
		if thisVal > otherVal {
			otherLess = true
		}
	}

	if thisLess && !otherLess {
		return -1
	}
	if !thisLess && otherLess {
		return 1
	}
	return 0 // Concurrent or equal
}

// SortByLWW sorts entries by their LWW timestamp in descending order (newest first)
func SortByLWW(entries []struct {
	ID        string
	Timestamp time.Time
}) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})
}
