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
	"github.com/odin-ai/odin/internal/config"
	"github.com/odin-ai/odin/internal/memory"
	"github.com/odin-ai/odin/pkg/logger"
)

// SDDPhase represents a phase in the Spec-Driven Development lifecycle
type SDDPhase string

const (
	PhaseProposal SDDPhase = "proposal"
	PhaseDomain   SDDPhase = "domain"   // DDD: Domain Modeling
	PhaseContract SDDPhase = "contract" // CFD: Contract Definition
	PhaseSpec     SDDPhase = "spec"     // BDD: Requirements
	PhaseDesign   SDDPhase = "design"   // Architecture
	PhaseTasks    SDDPhase = "tasks"    // Breakdown
	PhaseApply    SDDPhase = "apply"    // TDD Implementation
	PhaseVerify   SDDPhase = "verify"   // Validation
	PhaseDeploy   SDDPhase = "deploy"   // Dvergar: Consultative Deployment
	PhaseArchive  SDDPhase = "archive"  // Completion
)

const (
	MethodologyStandard  = "standard"
	MethodologyTriad     = "triad"
	MethodologyPentakill = "pentakill"
)

// GetPhaseOrder returns the ordered list of phases for a given methodology
func GetPhaseOrder(methodology string) []SDDPhase {
	switch methodology {
	case MethodologyPentakill:
		return []SDDPhase{
			PhaseProposal,
			PhaseDomain,
			PhaseContract,
			PhaseSpec,
			PhaseDesign,
			PhaseTasks,
			PhaseApply,
			PhaseVerify,
			PhaseDeploy,
			PhaseArchive,
		}
	case MethodologyTriad:
		return []SDDPhase{
			PhaseProposal,
			PhaseSpec,
			PhaseDesign,
			PhaseTasks,
			PhaseApply,
			PhaseVerify,
			PhaseArchive,
		}
	default:
		return []SDDPhase{
			PhaseProposal,
			PhaseSpec,
			PhaseDesign,
			PhaseTasks,
			PhaseApply,
			PhaseVerify,
			PhaseArchive,
		}
	}
}

// NextPhase returns the next phase after the current one based on methodology
func NextPhase(current SDDPhase, methodology string) SDDPhase {
	order := GetPhaseOrder(methodology)
	for i, phase := range order {
		if phase == current && i < len(order)-1 {
			return order[i+1]
		}
	}
	return current
}

// IsValidPhase checks if a string is a valid SDD phase for any methodology
func IsValidPhase(phase string) bool {
	allPhases := []SDDPhase{
		PhaseProposal, PhaseDomain, PhaseContract, PhaseSpec,
		PhaseDesign, PhaseTasks, PhaseApply, PhaseVerify, PhaseDeploy, PhaseArchive,
	}
	for _, p := range allPhases {
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
	Methodology string            `json:"methodology"` // "standard", "triad", "pentakill"
	Status      string            `json:"status"`      // "active", "paused", "completed"
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
func (s *SessionStore) Create(changeName, description, methodology string) (*Session, error) {
	if methodology == "" {
		methodology = "standard"
	}
	session := &Session{
		ID:          uuid.New().String(),
		ChangeName:  changeName,
		Phase:       PhaseProposal,
		Methodology: methodology,
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
	mimir          *memory.Store
	guardian       *HeimdallGuardian
	config         *config.Config
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(sessionsPath string) (*Orchestrator, error) {
	store, err := NewSessionStore(sessionsPath)
	if err != nil {
		return nil, err
	}

	mimir, err := memory.NewStore(nil) // Uses default config (with global DB path)
	if err != nil {
		logger.Warn("Failed to initialize Mimir in orchestrator", "error", err)
	}

	o := &Orchestrator{
		sessionsPath: sessionsPath,
		store:        store,
		mimir:        mimir,
	}
	o.guardian = NewHeimdallGuardian(o)
	return o, nil
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
func (o *Orchestrator) CreateSession(changeName, description, methodology string) (*Session, error) {
	session, err := o.store.Create(changeName, description, methodology)
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

	nextPhase := NextPhase(session.Phase, session.Methodology)
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

// CompleteSession marks a session as completed and archives to global memory
func (o *Orchestrator) CompleteSession(id string) error {
	session, err := o.store.Load(id)
	if err != nil {
		return err
	}

	session.Status = "completed"
	if err := o.store.Save(session); err != nil {
		return err
	}

	// Archive to Global Memory (Mimir)
	if o.mimir != nil {
		discovery := &memory.Memory{
			Content: fmt.Sprintf("Completed change: %s. Description: %s", session.ChangeName, session.Description),
			Project: session.ChangeName,
			Tags:    []string{"archive", "discovery", "session_complete"},
			Metadata: map[string]interface{}{
				"id":          session.ID,
				"description": session.Description,
				"methodology": session.Methodology,
				"timestamp":   time.Now().Format(time.RFC3339),
			},
		}
		if err := o.mimir.StoreGlobal(discovery); err != nil {
			logger.Warn("Failed to archive session to global memory", "id", id, "error", err)
		} else {
			logger.Info("Session archived to Mimir Global", "id", id)
		}
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
