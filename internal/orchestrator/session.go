package orchestrator

import (
	"time"
)

// SessionState represents the current state of a session
type SessionState struct {
	ID          string        `json:"id"`
	ChangeName  string        `json:"change_name"`
	Phase       SDDPhase      `json:"phase"`
	Methodology string        `json:"methodology"`
	Status      string        `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	ElapsedTime time.Duration `json:"elapsed_time"`
	Progress    float64       `json:"progress"` // 0.0 to 1.0
}

// CalculateProgress calculates the progress percentage for a phase based on methodology
func CalculateProgress(phase SDDPhase, methodology string) float64 {
	order := GetPhaseOrder(methodology)
	for i, p := range order {
		if p == phase {
			return float64(i) / float64(len(order)-1)
		}
	}
	return 0.0
}

// NewSessionState creates a SessionState from a Session
func NewSessionState(session *Session) *SessionState {
	elapsed := time.Since(session.CreatedAt)
	return &SessionState{
		ID:          session.ID,
		ChangeName:  session.ChangeName,
		Phase:       session.Phase,
		Methodology: session.Methodology,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt,
		UpdatedAt:   session.UpdatedAt,
		ElapsedTime: elapsed,
		Progress:    CalculateProgress(session.Phase, session.Methodology),
	}
}
