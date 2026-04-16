package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/odin-ai/odin/pkg/logger"
)

// Commands returns the orchestrator CLI commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orchestrator",
		Short: "Manage SDD orchestration",
		Long: `Manage Spec-Driven Development sessions and phase transitions.
Völva guides you through the SDD lifecycle.`,
	}

	cmd.AddCommand(
		newNextCmd(),
		newStatusCmd(),
		newSessionListCmd(),
		newSessionResumeCmd(),
		newSessionPauseCmd(),
	)

	return cmd
}

func newNextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Advance to next SDD phase",
		Long: `Advance the current session to the next phase in the SDD lifecycle.
Only works if there is an active session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNext(cmd)
		},
	}
}

func runNext(cmd *cobra.Command) error {
	// Get session ID
	sessionID, _ := cmd.Flags().GetString("session")
	if sessionID == "" {
		// Try to get current session from state
		state, err := loadState()
		if err != nil {
			return fmt.Errorf("failed to load state: %w", err)
		}
		sessionID = state.GetCurrentID()
		if sessionID == "" {
			return fmt.Errorf("no active session. Use 'odin session resume <id>' first")
		}
	}

	// Create orchestrator
	cfg := DefaultSessionConfig()
	orch, err := NewOrchestrator(cfg.Path)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	// Advance phase
	session, err := orch.AdvancePhase(sessionID)
	if err != nil {
		return fmt.Errorf("failed to advance phase: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]interface{}{
			"status": "success",
			"id":     session.ID,
			"phase":  session.Phase,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Printf("Session advanced to phase: %s\n", session.Phase)
	return nil
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current phase status",
		Long:  `Display the current session status including phase and progress.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd)
		},
	}
}

func runStatus(cmd *cobra.Command) error {
	// Get session ID
	sessionID, _ := cmd.Flags().GetString("session")
	if sessionID == "" {
		// Try to get current session from state
		state, err := loadState()
		if err != nil {
			return fmt.Errorf("failed to load state: %w", err)
		}
		sessionID = state.GetCurrentID()
	}

	if sessionID == "" {
		fmt.Println("No active session")
		return nil
	}

	// Create orchestrator
	cfg := DefaultSessionConfig()
	orch, err := NewOrchestrator(cfg.Path)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	session, err := orch.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	state := NewSessionState(session)

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(state)
	}

	// Display status
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║           SDD Session Status                   ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Printf("║  Change:    %-35s║\n", session.ChangeName)
	fmt.Printf("║  Phase:     %-35s║\n", session.Phase)
	fmt.Printf("║  Status:    %-35s║\n", session.Status)
	fmt.Printf("║  Progress:  %-35s║\n", fmt.Sprintf("%.0f%%", state.Progress*100))
	fmt.Printf("║  Started:   %-35s║\n", session.CreatedAt.Format(time.RFC822))
	fmt.Println("╚══════════════════════════════════════════════════╝")

	// Show phase progress
	fmt.Println()
	fmt.Println("Phase Progress:")
	for i, phase := range PhaseOrder {
		prefix := "[ ]"
		if phase == session.Phase {
			prefix = "[*]"
		} else if i < len(PhaseOrder) {
			// Check if phase is completed
			sessionState := NewSessionState(session)
			if sessionState.Progress >= float64(i)/float64(len(PhaseOrder)-1) {
				prefix = "[x]"
			}
		}
		fmt.Printf("  %s %s\n", prefix, phase)
	}

	return nil
}

func newSessionListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active sessions",
		Long:  `Display all SDD sessions with their status and phase.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSessionList(cmd)
		},
	}
}

func runSessionList(cmd *cobra.Command) error {
	cfg := DefaultSessionConfig()
	orch, err := NewOrchestrator(cfg.Path)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	sessions, err := orch.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(sessions)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCHANGE\tPHASE\tSTATUS\tCREATED")

	for _, s := range sessions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			s.ID[:8], // Short ID
			s.ChangeName,
			s.Phase,
			s.Status,
			s.CreatedAt.Format("2006-01-02"),
		)
	}
	w.Flush()

	return nil
}

func newSessionResumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <id>",
		Short: "Resume a session",
		Long: `Resume a paused or existing session by its ID.
Use 'odin session list' to see available sessions.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSessionResume(cmd, args[0])
		},
	}

	cmd.Flags().Bool("create", false, "Create a new session if not found")
	return cmd
}

func runSessionResume(cmd *cobra.Command, id string) error {
	cfg := DefaultSessionConfig()
	orch, err := NewOrchestrator(cfg.Path)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	session, err := orch.ResumeSession(id)
	if err != nil {
		// Check if we should create a new session
		if os.IsNotExist(err) {
			create, _ := cmd.Flags().GetBool("create")
			if create {
				// Create new session with id as name
				session, err = orch.CreateSession(id, "")
				if err != nil {
					return fmt.Errorf("failed to create session: %w", err)
				}
				logger.Info("Created new session", "id", session.ID)
			} else {
				return fmt.Errorf("session not found: %s (use --create to create new)", id)
			}
		} else {
			return fmt.Errorf("failed to resume session: %w", err)
		}
	}

	// Save current session ID to state
	state, err := loadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}
	state.SetCurrentID(session.ID)
	if err := state.Save(); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]interface{}{
			"status": "success",
			"id":     session.ID,
			"phase":  session.Phase,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Printf("Session resumed: %s (phase: %s)\n", session.ID[:8], session.Phase)
	return nil
}

func newSessionPauseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pause <id>",
		Short: "Pause current session",
		Long:  `Pause an active SDD session. Use 'odin session resume <id>' to continue.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSessionPause(cmd, args[0])
		},
	}
}

func runSessionPause(cmd *cobra.Command, id string) error {
	cfg := DefaultSessionConfig()
	orch, err := NewOrchestrator(cfg.Path)
	if err != nil {
		return fmt.Errorf("failed to create orchestrator: %w", err)
	}

	if err := orch.PauseSession(id); err != nil {
		return fmt.Errorf("failed to pause session: %w", err)
	}

	// Clear current session from state if it was this one
	state, err := loadState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}
	if state.GetCurrentID() == id {
		state.SetCurrentID("")
		if err := state.Save(); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		data := map[string]string{
			"status": "success",
			"id":     id,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	fmt.Printf("Session paused: %s\n", id[:8])
	return nil
}

// SessionConfig holds session configuration
type SessionConfig struct {
	Path             string
	SnapshotInterval time.Duration
	MaxSessions      int
}

// DefaultSessionConfig returns the default session configuration
func DefaultSessionConfig() *SessionConfig {
	homeDir, _ := os.UserHomeDir()
	return &SessionConfig{
		Path:             filepath.Join(homeDir, ".odin", "sessions"),
		SnapshotInterval: 5 * time.Minute,
		MaxSessions:      10,
	}
}

// loadState loads the orchestrator state
func loadState() (*State, error) {
	homeDir, _ := os.UserHomeDir()
	statePath := filepath.Join(homeDir, ".odin", "orchestrator-state.json")
	return NewState(statePath)
}
