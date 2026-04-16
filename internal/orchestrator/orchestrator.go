// Package orchestrator provides SDD orchestration for ODIN
// Völva is the Norse sorceress who guides the journey through the SDD phases
package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/odin-ai/odin/pkg/logger"
)

// SDDPhase represents a phase in the Spec-Driven Development lifecycle
type SDDPhase string

const (
	PhaseProposal SDDPhase = "proposal"
	PhaseSpec     SDDPhase = "spec"
	PhaseDesign   SDDPhase = "design"
	PhaseTasks    SDDPhase = "tasks"
	PhaseApply    SDDPhase = "apply"
	PhaseVerify   SDDPhase = "verify"
	PhaseArchive  SDDPhase = "archive"
)

// PhaseOrder defines the ordering of SDD phases
var PhaseOrder = []SDDPhase{
	PhaseProposal,
	PhaseSpec,
	PhaseDesign,
	PhaseTasks,
	PhaseApply,
	PhaseVerify,
	PhaseArchive,
}

// NextPhase returns the next phase after the current one
func NextPhase(current SDDPhase) SDDPhase {
	for i, phase := range PhaseOrder {
		if phase == current && i < len(PhaseOrder)-1 {
			return PhaseOrder[i+1]
		}
	}
	return current
}

// IsValidPhase checks if a string is a valid SDD phase
func IsValidPhase(phase string) bool {
	for _, p := range PhaseOrder {
		if string(p) == phase {
			return true
		}
	}
	return false
}

// Session represents an active SDD session
type Session struct {
	ID          string            `json:"id"`
	ChangeName  string            `json:"change_name"`
	Phase       SDDPhase          `json:"phase"`
	Status      string            `json:"status"` // "active", "paused", "completed"
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	SnapshotRef string            `json:"snapshot_ref"` // Git commit SHA
	Description string            `json:"description"`
	Artifacts   map[string]string `json:"artifacts"` // Paths to artifacts
}

// SessionStore handles session persistence
type SessionStore struct {
	sessionsPath string
}

// NewSessionStore creates a new session store
func NewSessionStore(sessionsPath string) (*SessionStore, error) {
	if err := os.MkdirAll(sessionsPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sessions directory: %w", err)
	}
	return &SessionStore{sessionsPath: sessionsPath}, nil
}

// Create creates a new session
func (s *SessionStore) Create(changeName, description string) (*Session, error) {
	session := &Session{
		ID:          uuid.New().String(),
		ChangeName:  changeName,
		Phase:       PhaseProposal,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Description: description,
		Artifacts:   make(map[string]string),
	}

	if err := s.Save(session); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	logger.Info("Session created", "id", session.ID, "change", changeName)
	return session, nil
}

// Save saves a session to disk
func (s *SessionStore) Save(session *Session) error {
	session.UpdatedAt = time.Now()

	sessionFile := filepath.Join(s.sessionsPath, session.ID+".json")
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(sessionFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Load loads a session by ID
func (s *SessionStore) Load(id string) (*Session, error) {
	sessionFile := filepath.Join(s.sessionsPath, id+".json")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s", id)
		}
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// List returns all sessions
func (s *SessionStore) List() ([]*Session, error) {
	entries, err := os.ReadDir(s.sessionsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Session{}, nil
		}
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []*Session
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-5] // Remove .json
		session, err := s.Load(id)
		if err != nil {
			logger.Warn("Failed to load session", "id", id, "error", err)
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Delete deletes a session
func (s *SessionStore) Delete(id string) error {
	sessionFile := filepath.Join(s.sessionsPath, id+".json")
	if err := os.Remove(sessionFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session not found: %s", id)
		}
		return fmt.Errorf("failed to delete session: %w", err)
	}

	logger.Info("Session deleted", "id", id)
	return nil
}

// Orchestrator manages SDD sessions and phase transitions
type Orchestrator struct {
	sessionsPath   string
	currentSession *Session
	store          *SessionStore
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(sessionsPath string) (*Orchestrator, error) {
	store, err := NewSessionStore(sessionsPath)
	if err != nil {
		return nil, err
	}

	return &Orchestrator{
		sessionsPath: sessionsPath,
		store:        store,
	}, nil
}

// CurrentSession returns the current active session
func (o *Orchestrator) CurrentSession() *Session {
	return o.currentSession
}

// SetCurrentSession sets the current active session
func (o *Orchestrator) SetCurrentSession(session *Session) {
	o.currentSession = session
}

// CreateSession creates a new SDD session
func (o *Orchestrator) CreateSession(changeName, description string) (*Session, error) {
	session, err := o.store.Create(changeName, description)
	if err != nil {
		return nil, err
	}

	o.currentSession = session
	return session, nil
}

// ResumeSession resumes a paused session
func (o *Orchestrator) ResumeSession(id string) (*Session, error) {
	session, err := o.store.Load(id)
	if err != nil {
		return nil, err
	}

	if session.Status == "completed" {
		return nil, fmt.Errorf("cannot resume completed session")
	}

	session.Status = "active"
	if err := o.store.Save(session); err != nil {
		return nil, err
	}

	o.currentSession = session
	logger.Info("Session resumed", "id", id, "phase", session.Phase)
	return session, nil
}

// PauseSession pauses the current session
func (o *Orchestrator) PauseSession(id string) error {
	session, err := o.store.Load(id)
	if err != nil {
		return err
	}

	if session.Status != "active" {
		return fmt.Errorf("session is not active")
	}

	session.Status = "paused"
	if err := o.store.Save(session); err != nil {
		return err
	}

	if o.currentSession != nil && o.currentSession.ID == id {
		o.currentSession = nil
	}

	logger.Info("Session paused", "id", id)
	return nil
}

// AdvancePhase advances the current session to the next phase
func (o *Orchestrator) AdvancePhase(id string) (*Session, error) {
	session, err := o.store.Load(id)
	if err != nil {
		return nil, err
	}

	if session.Status != "active" {
		return nil, fmt.Errorf("session is not active")
	}

	nextPhase := NextPhase(session.Phase)
	if nextPhase == session.Phase {
		return nil, fmt.Errorf("already at final phase")
	}

	session.Phase = nextPhase
	if err := o.store.Save(session); err != nil {
		return nil, err
	}

	if o.currentSession != nil && o.currentSession.ID == id {
		o.currentSession = session
	}

	logger.Info("Session advanced", "id", id, "new_phase", nextPhase)
	return session, nil
}

// CompleteSession marks a session as completed
func (o *Orchestrator) CompleteSession(id string) error {
	session, err := o.store.Load(id)
	if err != nil {
		return err
	}

	session.Status = "completed"
	if err := o.store.Save(session); err != nil {
		return err
	}

	if o.currentSession != nil && o.currentSession.ID == id {
		o.currentSession = nil
	}

	logger.Info("Session completed", "id", id)
	return nil
}

// ListSessions returns all sessions
func (o *Orchestrator) ListSessions() ([]*Session, error) {
	return o.store.List()
}

// GetSession returns a session by ID
func (o *Orchestrator) GetSession(id string) (*Session, error) {
	return o.store.Load(id)
}
