package orchestrator

import (
	"fmt"
	"time"

	"github.com/odin-ai/odin/internal/memory"
	"github.com/odin-ai/odin/pkg/logger"
)

// HeimdallGuardian represents the self-healing monitor
type HeimdallGuardian struct {
	orch *Orchestrator
}

// NewHeimdallGuardian creates a new guardian for the orchestrator
func NewHeimdallGuardian(orch *Orchestrator) *HeimdallGuardian {
	return &HeimdallGuardian{orch: orch}
}

// CheckSession evaluates the current session status and triggers self-healing if needed
func (h *HeimdallGuardian) CheckSession(sessionID string, verifyFailed bool, failureContext string) error {
	session, err := h.orch.store.Load(sessionID)
	if err != nil {
		return err
	}

	if session.Phase == PhaseVerify && verifyFailed {
		logger.Warn("Heimdall detected verification failure. Initiating Self-Healing loop.", "session", sessionID)
		
		// Emit inner monologue
		if h.orch.mimir != nil {
			logger.Think("Heimdall Guardian: Verification failed. Rolling back to 'apply' phase to fix issues.")
		}

		// Transition back to Apply
		session.Phase = PhaseApply
		session.Status = "healing" // Special status for self-healing
		
		if err := h.orch.store.Save(session); err != nil {
			return fmt.Errorf("failed to transition back to apply: %w", err)
		}

		// Record the healing pattern in Mimir Global Memory
		if h.orch.mimir != nil {
			h.orch.mimir.StoreGlobal(&memory.Memory{
				Content: fmt.Sprintf("Self-healing triggered for change %s. Reason: %s", session.ChangeName, failureContext),
				Project: session.ChangeName,
				Tags:    []string{"heimdall", "healing", "learning_pattern"},
				Metadata: map[string]interface{}{
					"session":        sessionID,
					"failure":        failureContext,
					"target_phase":   PhaseApply,
					"recovery_time":  time.Now().Format(time.RFC3339),
				},
			})
		}

		logger.Info("Self-healing triggered: session moved back to 'apply' phase", "id", sessionID)
	}

	return nil
}
