package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// State handles persistence of orchestrator state
type State struct {
	mu        sync.RWMutex
	stateFile string
	currentID string
	sessions  map[string]*Session
	lastSave  time.Time
}

// NewState creates a new State instance
func NewState(statePath string) (*State, error) {
	if err := os.MkdirAll(filepath.Dir(statePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	state := &State{
		stateFile: statePath,
		sessions:  make(map[string]*Session),
		lastSave:  time.Now(),
	}

	// Try to load existing state
	if err := state.Load(); err != nil {
		// State file doesn't exist yet, which is fine
		logger.Debug("No existing state file found, starting fresh")
	}

	return state, nil
}

// Load loads state from disk
func (s *State) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var stateData struct {
		CurrentID string              `json:"current_id"`
		Sessions  map[string]*Session `json:"sessions"`
		LastSave  time.Time           `json:"last_save"`
	}

	if err := json.Unmarshal(data, &stateData); err != nil {
		return fmt.Errorf("failed to unmarshal state: %w", err)
	}

	s.currentID = stateData.CurrentID
	s.sessions = stateData.Sessions
	s.lastSave = stateData.LastSave

	return nil
}

// Save saves state to disk
func (s *State) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stateData := struct {
		CurrentID string              `json:"current_id"`
		Sessions  map[string]*Session `json:"sessions"`
		LastSave  time.Time           `json:"last_save"`
	}{
		CurrentID: s.currentID,
		Sessions:  s.sessions,
		LastSave:  time.Now(),
	}

	data, err := json.MarshalIndent(stateData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(s.stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	s.lastSave = stateData.LastSave
	return nil
}

// GetCurrentID returns the current session ID
func (s *State) GetCurrentID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentID
}

// SetCurrentID sets the current session ID
func (s *State) SetCurrentID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentID = id
}

// GetSession returns a session by ID
func (s *State) GetSession(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	return session, ok
}

// AddSession adds a session to the state
func (s *State) AddSession(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
}

// RemoveSession removes a session from the state
func (s *State) RemoveSession(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

// ListSessions returns all sessions from state
func (s *State) ListSessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}
