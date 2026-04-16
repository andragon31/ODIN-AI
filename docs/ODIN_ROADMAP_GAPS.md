# ODIN — Plan de Paridad con gentle-ai + Metodología TDD·BDD·SDD

> **Versión**: 1.0.0  
> **Fecha**: Abril 2026  
> **Autor**: Andragon  
> **Base**: Análisis de gaps ODIN vs gentle-ai (verificado contra código real)  
> **Estado**: Draft — Listo para ejecutar

---

## Contexto: ¿De dónde venimos?

Después de revisar el código real de ambos proyectos:

| | gentle-ai | ODIN |
|--|-----------|------|
| **Rol** | Instalador + configurador de agentes AI | Runtime enterprise local-first |
| **Fortaleza** | AgentBuilder, Catalog, Pipeline, Multi-editor | Router, Mimir, Heimdall, Bifrost, WASM |
| **Debilidad** | Sin memory, sin security, sin router | Runes vacíos, sin catalog, sin agentbuilder |
| **Stack Go** | bubbletea + lipgloss + bubbles | cobra + viper + bubbletea + sqlite + go-git |

**El objetivo no es clonar gentle-ai — es absorber lo que le falta a ODIN sin perder sus vitaminas.**

---

## Mapa de Paridad: ¿Qué construir?

```
gentle-ai tiene          ODIN necesita implementar
─────────────────────    ──────────────────────────
agentbuilder/          → internal/runeforge/       (Rune Generator via Router local)
catalog/               → internal/catalog/         (Catálogo instalable)
pipeline/              → internal/pipeline/        (Pipeline con rollback real)
installcmd/            → internal/installcmd/      (Instalador de componentes)
backup/                → internal/backup/          (Backup pre-cambio)
update/                → internal/update/          (Auto-update del binario)
agents/ (multi-editor) → internal/agents/          (Config installers)
skills/ (contenido)    → runes/ (contenido real)   (15+ runes funcionales)
```

---

## Milestone 1 — Runes Registry con Contenido Real
**Prioridad**: CRÍTICO  
**Tiempo estimado**: 1 sprint (1 semana)  
**Por qué primero**: Sin contenido, toda la infraestructura de Runes es una caja vacía.

### 1.1 Poblar `runes/` con el Core Set

Crear la siguiente estructura de directorios y archivos:

```
runes/
├── _shared/                      # Assets compartidos entre runes
│   ├── engram-convention.md
│   └── persistence-contract.md
│
├── branch-pr/                    # Portado de gentle-ai skills/branch-pr
│   ├── RUNE.md                   # Instrucciones del rune (= SKILL.md)
│   └── rune.yaml                 # Metadata + schema CUE
│
├── issue-creation/               # Portado de gentle-ai
│   ├── RUNE.md
│   └── rune.yaml
│
├── sdd-explore/
│   ├── RUNE.md
│   └── rune.yaml
│
├── sdd-propose/
├── sdd-spec/
├── sdd-design/
├── sdd-tasks/
├── sdd-apply/
├── sdd-verify/
├── sdd-archive/
│
├── go-testing/
│   ├── RUNE.md
│   └── rune.yaml
│
└── skill-creator/
    ├── RUNE.md
    └── rune.yaml
```

**Formato `rune.yaml`** (schema unificado):
```yaml
name: branch-pr
version: 1.0.0
description: "PR creation workflow following issue-first enforcement"
author: Andragon
tags: [git, pr, workflow, sdd]

triggers:
  commands: [pr, branch, merge]
  context: [git, development]

execution:
  type: prompt
  sandbox: false

outputs:
  console: "PR created with conventional commit format"
```

### 1.2 Archivos a crear

| Archivo | Acción |
|---------|--------|
| `runes/branch-pr/RUNE.md` | NUEVO — portar de gentle-ai skills/branch-pr |
| `runes/branch-pr/rune.yaml` | NUEVO |
| `runes/issue-creation/RUNE.md` | NUEVO — portar de gentle-ai |
| `runes/issue-creation/rune.yaml` | NUEVO |
| `runes/sdd-{phase}/RUNE.md` (×8) | NUEVO — portar de antigravity skills/ |
| `runes/sdd-{phase}/rune.yaml` (×8) | NUEVO |
| `runes/go-testing/RUNE.md` | NUEVO |
| `runes/skill-creator/RUNE.md` | NUEVO |

### 1.3 Criterios de Aceptación

```
[ ] odin rune list           → lista 12+ runes instalados
[ ] odin rune validate <name> → schema CUE válido
[ ] odin rune test branch-pr  → pass en sandbox
[ ] odin rune search --tags sdd → encuentra los 8 runes SDD
[ ] odin rune install --from-registry sdd-explore → instala en ~/.odin/runes/
```

---

## Milestone 2 — Catalog System + Installation Pipeline
**Prioridad**: CRÍTICO  
**Tiempo estimado**: 1 sprint (1 semana)

### 2.1 `internal/catalog/` — Catálogo de componentes

**Archivos a crear:**

```
internal/catalog/
├── agents.go        # Agentes AI conocidos (claude, opencode, gemini, codex)
├── components.go    # Componentes instalables (sdd, engram, guardian, bifrost)
├── runes.go         # Runes disponibles en el registry
├── catalog.go       # Catalog manager core
└── catalog_test.go  # Tests unitarios
```

**Tipos core:**
```go
type AgentID string

const (
    AgentClaudeCode AgentID = "claude-code"
    AgentGeminiCLI  AgentID = "gemini-cli"
    AgentOpenCode   AgentID = "opencode"
    AgentCodex      AgentID = "codex"
    AgentCursor     AgentID = "cursor"
    AgentWindsurf   AgentID = "windsurf"
)

type Component struct {
    ID          string
    Name        string
    Description string
    DependsOn   []string
    Runes       []string
}
```

**Comandos CLI a implementar:**
```bash
odin catalog list
odin catalog list --type agents
odin catalog list --type components
odin catalog install sdd
odin catalog install heimdall
odin catalog install mimir
odin catalog info sdd
```

### 2.2 `internal/pipeline/` — Pipeline con Rollback

Adaptado de gentle-ai/internal/pipeline:

```
internal/pipeline/
├── orchestrator.go
├── stages.go
├── runner.go
├── rollback.go
├── result.go
└── pipeline_test.go
```

**Flujo del pipeline:**
```
odin catalog install <component>
          │
[Stage 1: Detect]  → OS, arquitectura, agentes instalados
          │
[Stage 2: Backup]  → ~/.odin/backup/<timestamp>/
          │
[Stage 3: Install] → runes, configs, hooks
          │
[Stage 4: Verify]  → Nornir verifica
          │
     [Commit]  ←→  [Rollback] si Verify falla
```

### 2.3 `internal/backup/` — Backup pre-cambio

```
internal/backup/
├── backup.go
└── restore.go
```

### 2.4 Archivos a crear/modificar

| Archivo | Acción |
|---------|--------|
| `internal/catalog/catalog.go` | NUEVO |
| `internal/catalog/agents.go` | NUEVO |
| `internal/catalog/components.go` | NUEVO |
| `internal/catalog/runes.go` | NUEVO |
| `internal/catalog/catalog_test.go` | NUEVO |
| `internal/pipeline/orchestrator.go` | NUEVO |
| `internal/pipeline/stages.go` | NUEVO |
| `internal/pipeline/runner.go` | NUEVO |
| `internal/pipeline/rollback.go` | NUEVO |
| `internal/pipeline/result.go` | NUEVO |
| `internal/pipeline/pipeline_test.go` | NUEVO |
| `internal/backup/backup.go` | NUEVO |
| `internal/backup/restore.go` | NUEVO |
| `internal/cli/cli.go` | MODIFICAR — añadir `catalog` command |

### 2.5 Criterios de Aceptación

```
[ ] odin catalog list         → tabla con agentes, components, runes
[ ] odin catalog install sdd  → instala en <2 min con rollback si falla
[ ] Stage "backup" crea ~/.odin/backup/<timestamp>/ antes de sobrescribir
[ ] Rollback restaura estado previo en <=5s
[ ] Pipeline cancelable con Ctrl+C (context.WithCancel)
```

---

## Milestone 3 — Multi-Agent Config Installer + RuneForge
**Prioridad**: ALTA  
**Tiempo estimado**: 1 sprint (1 semana)

### 3.1 `internal/agents/` — Config Installers

```
internal/agents/
├── agent.go          # Interface AgentInstaller
├── claude.go         # Claude Code → genera CLAUDE.md / AGENTS.md
├── gemini.go         # Gemini CLI → .gemini/GEMINI.md
├── opencode.go       # OpenCode → .opencode/config
├── cursor.go         # Cursor → .cursor/rules
├── windsurf.go       # Windsurf → .windsurf/rules
├── codex.go          # Codex config
├── detector.go       # Detecta qué agentes están instalados
└── agents_test.go
```

**Interface core:**
```go
type AgentInstaller interface {
    ID() AgentID
    Name() string
    Available() bool
    Install(cfg *AgentConfig) error
    Verify() error
    Uninstall() error
}
```

**Comandos:**
```bash
odin install --agent claude-code
odin install --agent gemini-cli
odin install --all-agents
odin install --detect
```

**DIFERENCIA vs gentle-ai**: Si el agente CLI no está instalado, ODIN usa su
Router local (Ollama) para generar el contenido del config — sin dependencias externas.

### 3.2 `internal/runeforge/` — RuneForge (= AgentBuilder de ODIN)

Donde gentle-ai usa `claude exec prompt`, ODIN usa su propio Router:

```
internal/runeforge/
├── forge.go          # RuneForge — genera runes via Router
├── engine.go         # Interface Engine
├── parser.go         # Parsea output del modelo → Rune struct
├── prompt.go         # Gestiona prompts de generación
├── forge_test.go
└── integration_test.go
```

**Flujo de generación:**
```go
type RuneForge struct {
    router  *router.Router   // usa el Router existente (Ollama/OpenRouter/Anthropic)
    parser  *Parser
}

func (f *RuneForge) Generate(ctx context.Context, req ForgeRequest) (*skills.Rune, error) {
    prompt := f.buildPrompt(req)
    response, _ := f.router.Generate(ctx, GenerateRequest{Model: req.Model, Prompt: prompt})
    return f.parser.ParseRune(response.Content)
}
```

**Comandos:**
```bash
odin rune forge "branch-pr" \
    --description "PR workflow siguiendo issue-first" \
    --model ollama:deepseek-coder \
    --tags git,pr,workflow

odin rune forge --from-example gentle-ai/skills/branch-pr \
    --adapt-for odin
```

### 3.3 Criterios de Aceptación

```
[ ] odin install --detect → lista agentes disponibles en <1s
[ ] odin install --agent claude-code → genera CLAUDE.md sin necesitar claude CLI
[ ] odin rune forge "test-rune" --model ollama:deepseek-coder → rune válido
[ ] RuneForge funciona offline con Ollama
[ ] Rune generado pasa validación CUE antes de instalarse
```

---

## Milestone 4 — Mimir con Embeddings Reales
**Prioridad**: ALTA  
**Tiempo estimado**: 1 sprint (1 semana)

### 4.1 Reemplazar SimpleVectorSearch con OllamaEmbedder

`SimpleVectorSearch.GenerateEmbedding()` es actualmente un placeholder TF-IDF
que no sirve para búsqueda semántica real.

```go
// NUEVO: internal/memory/embedder.go
type OllamaEmbedder struct {
    endpoint string   // "http://localhost:11434"
    model    string   // "nomic-embed-text"
}

func (e *OllamaEmbedder) GenerateEmbedding(text string) ([]float32, error) {
    // POST /api/embeddings {"model": "nomic-embed-text", "prompt": text}
    // Vector real de 768 dimensiones
}
```

**Fallback chain:**
```
1. OllamaEmbedder (nomic-embed-text) — local, 768 dims
2. OpenRouterEmbedder               — si Ollama no disponible
3. SimpleVectorSearch               — último recurso (TF-IDF)
```

### 4.2 Archivos afectados

| Archivo | Acción |
|---------|--------|
| `internal/memory/memory.go` | MODIFICAR — NewStore() usa OllamaEmbedder |
| `internal/memory/embedder.go` | NUEVO — OllamaEmbedder |
| `internal/memory/embedder_test.go` | NUEVO |
| `internal/memory/store.go` | MODIFICAR — VectorSearch con embeddings reales |

### 4.3 Criterios de Aceptación

```
[ ] mimir store "texto" → embedding de 768 dims con Ollama
[ ] mimir search "texto similar" → resultados semánticamente relevantes
[ ] Búsqueda semántica < 200ms en 10k registros
[ ] Fallback a SimpleVectorSearch si Ollama no disponible (no crash)
[ ] odin status → muestra "Mimir: sqlite-vss + nomic-embed-text"
```

---

## Milestone 5 — Auto-Update + Estado del Sistema
**Prioridad**: MEDIA  
**Tiempo estimado**: Medio sprint (3-4 días)

### 5.1 `internal/update/` — Auto-Update

```
internal/update/
├── update.go
├── verify.go         # Verifica firma cosign del binario
└── update_test.go
```

```bash
odin self-update                    # Última versión estable
odin self-update --channel beta
odin self-update --check            # Solo verifica
```

### 5.2 `odin status` — Estado completo mejorado

```
odin status

╭─ ODIN AI Ecosystem Status ────────────────────────────────────────╮
│                                                                     │
│  Odin Core      v0.1.0              OK Running                     │
│  Mimir          sqlite-vss + ollama  OK 10k memories               │
│  Heimdall       OPA + gosec          OK 3 policies active          │
│  Bifrost        go-git + CRDT        OK synced                     │
│  Runes          15 runes             OK all valid                  │
│  Nornir         0 flaky              OK all benchmarks pass        │
│                                                                     │
│  Router:                                                            │
│    ollama-local    OK (deepseek-coder, nomic-embed-text)           │
│    openrouter      OK                                               │
│    anthropic       OK (claude-3-5-sonnet)                          │
│                                                                     │
│  Agents: claude-code OK  gemini-cli OK  cursor WARN               │
│                                                                     │
╰──────────────────────────────────────── v0.1.0 · local-first ────╯
```

---

## Resumen de Cambios por Milestone

| Milestone | Archivos nuevos | Archivos mod | Sprint |
|-----------|----------------|-------------|--------|
| M1: Runes Content | 26 `.md` + `.yaml` | 0 | S1 |
| M2: Catalog + Pipeline | 13 `.go` | 1 | S2 |
| M3: Agents + RuneForge | 10 `.go` | 1 | S3 |
| M4: Mimir Embeddings | 3 `.go` | 2 | S4 |
| M5: Auto-Update + Status | 3 `.go` | 2 | S5 (parcial) |

---

---

# PARTE II — Metodología Unificada: SDD + TDD + BDD

## El Problema que Resuelve

ODIN implementa **SDD** (Spec-Driven Development) con 9 fases — excelente para
*planear* qué construir. Pero falta la capa que garantiza *cómo* se comporta
y *cómo* sabemos que funciona correctamente.

```
SDD solo:    Plan detallado → Código → "Parece que funciona" → Ship
SDD+TDD+BDD: Plan → Behavior → Código → Tests → Verificación → Ship
```

---

## La Triada: Cada metodología en su capa

```
┌──────────────────────────────────────────────────────────────────┐
│                        CICLO DE DESARROLLO                        │
│                                                                    │
│  SDD ─────────────────────────────────────────────────────────── │
│  "¿QUÉ construimos y POR QUÉ?"                                    │
│  proposal → spec → design → tasks → apply → verify → archive      │
│                │                                                   │
│                ▼ (al entrar en spec)                              │
│                                                                    │
│  BDD ─────────────────────────────────────────────────────────── │
│  "¿CÓMO debería COMPORTARSE?"                                     │
│  Feature files Gherkin → Given/When/Then → contrato público       │
│                │                                                   │
│                ▼ (al entrar en apply)                             │
│                                                                    │
│  TDD ─────────────────────────────────────────────────────────── │
│  "¿CÓMO lo IMPLEMENTAMOS correctamente?"                          │
│  Red → Green → Refactor → Red → Green → Refactor...              │
│                                                                    │
└──────────────────────────────────────────────────────────────────┘
```

**Regla de oro**:
- **SDD** opera a nivel de *cambio* (change)
- **BDD** opera a nivel de *feature/comportamiento*
- **TDD** opera a nivel de *unidad/integración*

---

## El Flujo Unificado SDD·BDD·TDD

### Fase `sdd-spec` → genera Features BDD

Como parte del entregable de spec, se crean Feature files Gherkin:

```gherkin
# openspec/features/catalog.feature

Feature: Catalog System
  Como desarrollador usando ODIN
  Quiero ver y instalar componentes desde un catálogo
  Para configurar mi entorno AI sin buscar manualmente

  Background:
    Given que ODIN está inicializado en mi sistema
    And el catálogo base está disponible

  Scenario: Listar catálogo completo
    When ejecuto "odin catalog list"
    Then veo una tabla con agentes, componentes y runes disponibles
    And el tiempo de respuesta es menor a 800ms

  Scenario: Instalar componente SDD
    When ejecuto "odin catalog install sdd"
    Then se instalan los 8 runes SDD en ~/.odin/runes/
    And se crea un backup previo en ~/.odin/backup/
    And si falla, el rollback restaura el estado anterior

  Scenario: Instalar componente sin internet
    Given que no hay conexión a internet
    When ejecuto "odin catalog install sdd"
    Then la instalación procede usando runes en caché local
    And no se bloquea esperando red
```

### Fase `sdd-design` → define contratos de interface

```go
// Contrato definido en design, implementado en apply
type CatalogManager interface {
    List(filter CatalogFilter) (*CatalogEntry, error)
    Install(componentID string, opts InstallOptions) (*InstallResult, error)
    Info(componentID string) (*Component, error)
}
```

### Fase `sdd-tasks` → tests TDD antes del código

Cada task incluye obligatoriamente test Red + implementación Green:

```markdown
## Task 3.2: Implementar CatalogManager.Install()

**Test primero (Red):**
```go
func TestCatalogInstall_SDD_HappyPath(t *testing.T) {
    catalog := NewCatalogManager(testConfig)
    result, err := catalog.Install("sdd", InstallOptions{DryRun: false})
    assert.NoError(t, err)
    assert.Equal(t, 8, result.RunesInstalled)
    assert.True(t, result.BackupCreated)
}

func TestCatalogInstall_Rollback_OnVerifyFail(t *testing.T) {
    catalog := NewCatalogManager(testConfig)
    catalog.SetVerifier(AlwaysFailVerifier{})
    result, err := catalog.Install("sdd", InstallOptions{})
    assert.Error(t, err)
    assert.True(t, result.RolledBack)
}
```
**Implementación (Green):** catalog.go → Install()
**Criterio de aceptación BDD:** Scenario "Instalar componente SDD" pasa
```

---

## Stack Técnico BDD para Go

**Herramienta**: `godog` (Cucumber para Go)

```go
// e2e/catalog_test.go
package e2e

import "github.com/cucumber/godog"

func TestCatalogFeatures(t *testing.T) {
    suite := godog.TestSuite{
        ScenarioInitializer: InitializeCatalogScenarios,
        Options: &godog.Options{
            Format:   "pretty",
            Paths:    []string{"../openspec/features"},
            TestingT: t,
        },
    }
    suite.Run()
}

func InitializeCatalogScenarios(ctx *godog.ScenarioContext) {
    steps := &CatalogSteps{odin: newTestODIN()}
    ctx.Step(`^ejecuto "([^"]*)"$`, steps.executeCommand)
    ctx.Step(`^veo una tabla con (.*)$`, steps.seeTableWith)
    ctx.Step(`^el tiempo de respuesta es menor a (\d+)ms$`, steps.responseTimeUnder)
}
```

**Patrón TDD obligatorio para cada paquete:**
```go
func TestCatalogManager_List(t *testing.T) {
    tests := []struct {
        name    string
        filter  CatalogFilter
        want    int
        wantErr bool
    }{
        {"sin filtro retorna todo", CatalogFilter{}, 20, false},
        {"filtro agents", CatalogFilter{Type: "agents"}, 6, false},
        {"filtro runes", CatalogFilter{Type: "runes"}, 12, false},
        {"filtro inválido", CatalogFilter{Type: "unknown"}, 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mgr := NewCatalogManager(testConfig)
            got, err := mgr.List(tt.filter)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Len(t, got.All(), tt.want)
        })
    }
}
```

---

## Estructura de Directorios con TDD·BDD·SDD

```
odin/
├── openspec/
│   ├── changes/                     # SDD — artifacts por cambio
│   │   └── catalog-system/
│   │       ├── proposal.md
│   │       ├── spec.md
│   │       ├── design.md
│   │       └── tasks.md
│   └── features/                    # BDD — Feature files Gherkin
│       ├── catalog.feature
│       ├── runeforge.feature
│       └── mimir-search.feature
│
├── internal/
│   └── catalog/
│       ├── catalog.go               # Implementación (resultado de TDD)
│       ├── catalog_test.go          # TDD — unit tests
│       └── testdata/
│           └── catalog.json
│
├── e2e/                             # BDD — step definitions godog
│   ├── catalog_test.go
│   ├── runeforge_test.go
│   └── helpers_test.go
│
└── Makefile
```

**Targets de Make:**
```makefile
test-unit:      go test -race -cover ./internal/...    # TDD
test-bdd:       go test -race ./e2e/...                # BDD (godog)
test-e2e:       go test -tags=e2e ./e2e/...            # Nornir matrix
test:           test-unit test-bdd
test-coverage:  go test -coverprofile=coverage.out ./...
```

---

## Actualización del rune `sdd-spec`

Como parte de este plan, el rune `sdd-spec` se actualiza para incluir Feature
files como entregable obligatorio:

```markdown
## Entregables de sdd-spec (actualizado)

1. `openspec/changes/{change}/spec.md`       — Spec de requisitos (existente)
2. `openspec/features/{change}.feature`      — Feature file BDD (NUEVO)
3. Interface contracts Go en el spec          — (mover de design a spec)

## Criterio de "Done" para sdd-apply:
- [ ] Todos los Scenarios del Feature file pasan (make test-bdd)
- [ ] Cobertura de unit tests >= 80% (make test-coverage)
- [ ] make test-race sin data races
```

---

## La Mejora Real: ¿Por qué vale la pena?

### Sin la triada (estado actual de ODIN):
```
SDD spec: "mimir debe buscar semánticamente"
→ Implementación: SimpleVectorSearch con TF-IDF placeholder
→ Resultado: Nadie verifica que sea semántico
→ Bug en producción: búsqueda retorna resultados irrelevantes
```

### Con SDD·BDD·TDD:
```
SDD spec: "mimir debe buscar semánticamente"
→ BDD Feature:
    Scenario: Búsqueda semántica encuentra conceptos relacionados
      Given 100 memorias sobre arquitectura Go
      When busco "diseño modular"
      Then top-3 contiene "hexagonal architecture"
      And score > 0.85
→ TDD:
    Red: falla porque SimpleVectorSearch no es semántico
    Green: implementar OllamaEmbedder
    Refactor: abstraer Embedder interface
→ Resultado: Mimir busca REALMENTE semántico, verificable, CI detecta regresiones
```

### Beneficios cuantificables:

| Métrica | Sin triada | Con SDD·BDD·TDD |
|---------|-----------|-----------------|
| Tiempo en encontrar regresiones | Días (manual) | Minutos (CI) |
| Documentación viva | README desactualizado | Feature files = spec ejecutable |
| Confianza para refactorizar | Baja | Alta |
| Onboarding | "Lee el PRD" | `make test-bdd` muestra comportamiento |
| Detección de placeholders | Alta deuda técnica | TDD Red los detecta |

---

## Resumen Ejecutivo

### ¿Qué construimos y en qué orden?

```
Sprint 1: Runes content (15 runes funcionales)    → ODIN tiene contenido real
Sprint 2: Catalog + Pipeline + Backup             → ODIN puede instalar como gentle-ai
Sprint 3: RuneForge + Multi-agent installer        → ODIN genera runes offline via Ollama
Sprint 4: Mimir con embeddings reales              → Búsqueda semántica funcional
Sprint 5: Auto-update + Status mejorado            → Production-ready
```

### ¿Cómo trabajamos?

```
Cada cambio sigue: SDD spec → BDD features → TDD red/green → verify
No se hace apply sin feat files.
No se hace verify sin tests pasando.
```

### ¿Cuál es el diferencial ODIN vs gentle-ai al final?

```
gentle-ai: Necesitas claude/opencode/gemini instalados → ellos generan tus skills
ODIN:      Usas Ollama local → ODIN genera tus runes offline
           Y además tienes: memory semántica, security OPA, CRDT sync, WASM plugins
```

> ODIN no es gentle-ai con lipstick. Es gentle-ai con cortex.

---

*Documento generado: Abril 2026*  
*Autor: Andragon*  
*Próxima revisión: Al completar Sprint 1 (runes content)*
