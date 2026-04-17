package deploy

import (
	"fmt"
	"github.com/odin-ai/odin/internal/blacksmith"
	"github.com/odin-ai/odin/pkg/logger"
)

// ForgeManager coordinates the blacksmith forge for user projects
type ForgeManager struct {
	forge *blacksmith.Forge
}

// NewForgeManager creates a new forge manager
func NewForgeManager(projectPath string) *ForgeManager {
	return &ForgeManager{
		forge: blacksmith.NewForge(projectPath),
	}
}

// RunForge starts the infrastructure generation process
func (fm *ForgeManager) RunForge() error {
	logger.Think("Dvergar: Iniciando el proceso de forja de infraestructura...")

	// 1. Detect Stack
	stack, err := fm.forge.DetectStack()
	if err != nil {
		logger.Warn("Dvergar: No se pudo detectar el stack automáticamente. Entrando en modo manual.")
	} else {
		logger.Info("Stack detectado", "language", stack.Language, "entry", stack.Entry)
	}

	// 2. Interview (Simulated for now, would be handled by CLI prompts)
	// In v6.0, the orquestrador will use the Think stream to guide this
	
	// 3. Generate Docker Local as default for Fase 2 validation
	targetStack := "go"
	if stack != nil {
		targetStack = stack.Language
	}

	err = fm.forge.GenerateDockerLocal(targetStack)
	if err != nil {
		return fmt.Errorf("error al forjar infraestructura: %w", err)
	}

	return nil
}
