package orchestrator

import (
	"github.com/odin-ai/odin/pkg/logger"
)

// Monologue handles the inner thoughts of the orchestrator
type Monologue struct{}

// NewMonologue creates a new monologue instance
func NewMonologue() *Monologue {
	return &Monologue{}
}

// Think emits an inner thought to the terminal
func (m *Monologue) Think(msg string) {
	logger.Think(msg)
}

// GlobalThink is a convenience function for global monologue emitting
func GlobalThink(msg string) {
	logger.Think(msg)
}
