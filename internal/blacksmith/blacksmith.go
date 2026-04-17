package blacksmith

import (
	"fmt"
	"github.com/odin-ai/odin/pkg/logger"
	"os"
	"path/filepath"
)

// InfrastructureType represents the target deployment environment
type InfrastructureType string

const (
	InfraDockerLocal InfrastructureType = "docker-local"
	InfraCloudRun    InfrastructureType = "cloud-run"
	InfraKubernetes  InfrastructureType = "kubernetes"
	InfraHeroku      InfrastructureType = "heroku"
)

// InterviewResult holds the answers from the technical interview
type InterviewResult struct {
	Target      InfrastructureType
	Stack       string
	UseCompose  bool
	NeedsCI     bool
	Environment map[string]string
}

// Forge handles the generation of infrastructure artifacts
type Forge struct {
	projectPath string
}

// NewForge creates a new Blacksmith forge
func NewForge(projectPath string) *Forge {
	return &Forge{projectPath: projectPath}
}

// Stack represents a detected technology stack
type Stack struct {
	Language string
	Entry    string
}

// DetectStack analyzes the project to identify the technology stack
func (f *Forge) DetectStack() (*Stack, error) {
	logger.Think("Dvergar Blacksmith: Analizando estructura del proyecto para detectar el stack...")

	// Detect Go
	if _, err := os.Stat(filepath.Join(f.projectPath, "go.mod")); err == nil {
		return &Stack{Language: "go", Entry: "cmd/"}, nil
	}

	// Detect Node.js
	if _, err := os.Stat(filepath.Join(f.projectPath, "package.json")); err == nil {
		return &Stack{Language: "node", Entry: "index.js"}, nil
	}

	// Detect Python
	if _, err := os.Stat(filepath.Join(f.projectPath, "requirements.txt")); err == nil {
		return &Stack{Language: "python", Entry: "main.py"}, nil
	}

	return nil, fmt.Errorf("no se pudo detectar un stack soportado automáticamente")
}

// ConductInterview simulates the consultative inquiry (logic to be used by CLI/TUI)
func (f *Forge) ConductInterview() *InterviewResult {
	// In a real CLI, this would prompt the user. 
	// The orchestrator calls this after receiving input from the user.
	logger.Think("Dvergar Blacksmith: Iniciando entrevista técnica para la forja de infraestructura...")
	return &InterviewResult{}
}

// GenerateDockerLocal generates a Dockerfile and docker-compose.yaml for local development
func (f *Forge) GenerateDockerLocal(stack string) error {
	logger.Think(fmt.Sprintf("Dvergar Blacksmith: Forjando contenedor Docker para el stack %s...", stack))
	
	dockerfile := f.getDockerfileTemplate(stack)
	if dockerfile == "" {
		return fmt.Errorf("stack no soportado para generación automática: %s", stack)
	}

	err := os.WriteFile(filepath.Join(f.projectPath, "Dockerfile"), []byte(dockerfile), 0644)
	if err != nil {
		return err
	}

	compose := f.getComposeTemplate(stack)
	err = os.WriteFile(filepath.Join(f.projectPath, "docker-compose.yaml"), []byte(compose), 0644)
	
	logger.Info("Infraestructura local generada exitosamente", "files", "Dockerfile, docker-compose.yaml")
	return err
}

func (f *Forge) getDockerfileTemplate(stack string) string {
	switch stack {
	case "go":
		return `FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/odin
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]`
	case "node":
		return `FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build
CMD ["npm", "start"]`
	default:
		return ""
	}
}

func (f *Forge) getComposeTemplate(stack string) string {
	return `version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=development`
}
