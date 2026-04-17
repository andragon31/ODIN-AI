# ODIN — Roadmap de Paridad con gentle-ai + Metodología TDD·BDD·SDD

> **Versión**: 5.1.0 — Roadmap 100% completado + Instalador remoto  
> **Fecha**: Abril 2026  
> **Estado**: 🏆 ROADMAP 100% COMPLETADO — Sin gaps restantes.

---

## Estado General — v5.1

```
RESUMEN EJECUTIVO — AUDITORÍA EXHAUSTIVA FINAL
─────────────────────────────────────────────────────────────────────
✅ Implementado y verificado (funcional)               : 36 items
⚠️  Implementado con gap menor (no bloquea nada)       : 0 items
❌ No implementado                                     : 0 items
─────────────────────────────────────────────────────────────────────
Progreso del roadmap: 100%
Delta desde v5.0: +1% (migrate_test.go 18/18 PASS,
                        TestMimirSearch_SemanticRelevance corregido y PASS)
```

---

## Mapa de Estado v5 — Inventario Completo Verificado

### Paquetes Core

| Paquete | Archivos | Estado |
|---------|----------|--------|
| `internal/catalog/` | catalog.go, catalog_test.go, commands.go | ✅ Completo |
| `internal/pipeline/` | orchestrator.go (514 ln) + commands.go + pipeline_test.go | ✅ Completo |
| `internal/pipeline/copyRuneFiles()` | en orchestrator.go ln 462 | ✅ Copia RUNE.md + rune.yaml reales |
| `internal/backup/` | backup.go + backup_test.go + commands.go | ✅ Completo |
| `internal/agents/` | 6 agentes + detector + agents_test.go + commands.go | ✅ Completo |
| `internal/runeforge/` | forge.go, engine.go, parser.go, prompt.go, forge_test.go, **integration_test.go**, commands.go | ✅ Completo |
| `internal/memory/` | embedder.go, embedder_test.go, memory.go, store.go, graph.go, prune.go, encrypt.go, sync.go, **memory_test.go (977 líneas)**, commands.go | ✅ Completo |
| `internal/update/` | update.go, verify.go (extractTarGz + extractZip reales), update_test.go, commands.go | ✅ Completo |
| `internal/migrate/` | commands.go (migrateConfigFiles, migrateMemories, migrateSkills, convertRulesToRego) | ✅ Completo |
| `internal/orchestrator/` | state.go | ✅ Presente |
| `internal/cli/` | cli.go | ✅ Completo |

### Guardian (Heimdall) — Paquete completo

| Archivo | Contenido | Estado |
|---------|-----------|--------|
| `guardian/commands.go` | 661 líneas — `odin heimdall check`, `runes`, `hook-install`, `report`, `status` | ✅ **COMPLETO** |
| `guardian/guardian.go` | Core Guardian struct + Check() | ✅ |
| `guardian/policy.go` | OPA policy loading | ✅ |
| `guardian/saast.go` | gosec + semgrep integration | ✅ |
| `guardian/hooks.go` | pre-commit hook install/uninstall | ✅ |
| `guardian/reporter.go` | JSON/text report generation | ✅ |
| `guardian/exec.go` | External tool execution | ✅ |
| `guardian/guardian_test.go` | Unit tests | ✅ |

> **Nota**: `odin heimdall runes` evalúa runes contra `security.rego` (sandbox, WASM, script, semver, `rm -rf`, `sudo`). El gap "odin guardian check" del roadmap anterior **ESTÁ IMPLEMENTADO** — solo que el comando se llama `odin heimdall check`, no `odin guardian check`.

### TUI (Völva) — Dashboard completo

| Archivo | Contenido | Estado |
|---------|-----------|--------|
| `tui/dashboard.go` | 343 líneas — `EcosystemDashboard`, `GatherComponentStatuses()`, `RenderDashboard()`, `GetOverallHealth()`, `GetDashboardJSON()` | ✅ **COMPLETO** |
| `tui/components.go` | 7529 bytes | ✅ |
| `tui/styles.go` | 4676 bytes | ✅ |
| `tui/theme.go` | 4141 bytes | ✅ |
| `tui/tui_test.go` | Tests | ✅ |
| `tui/commands.go` | CLI commands | ✅ |

> El dashboard muestra: Odin Core, Mimir (Memory), Heimdall (Security), Bifrost (Sync), Runes (count), Router (agents), Pipeline, Nornir. Incluye `GetOverallHealth()` → HEALTHY / DEGRADED / UNHEALTHY y `GetDashboardJSON()` para `odin status --json`.

### Testing

| Archivo | Contenido | Estado |
|---------|-----------|--------|
| `e2e/main_test.go` | 1353 líneas — `TestMain` + `godog.TestSuite` + `InitializeScenario` (60+ step defs) | ✅ BDD real |
| `e2e/catalog_test.go` | e2e adicionales | ✅ |
| `e2e/pipeline_test.go` | e2e adicionales | ✅ |
| `e2e/runeforge_test.go` | e2e adicionales | ✅ |
| `e2e/agents_test.go` | e2e adicionales | ✅ |
| `internal/memory/memory_test.go` | 977 líneas — Store, Search, Tags, Delete, Graph, Prune, Encrypt, Sync, Pruner, SmartPruner | ✅ Exhaustivo |
| `internal/runeforge/integration_test.go` | ParseRune, ValidateRune via skills, ForgeResult, ParsePartialRune | ✅ Con MockRouter |
| `go.mod` | `github.com/cucumber/godog v0.14.1` + gherkin + messages | ✅ |

### BDD Feature Files + Spec

| Archivo | Estado |
|---------|--------|
| `openspec/features/catalog.feature` | ✅ 6 scenarios |
| `openspec/features/pipeline.feature` | ✅ Completo |
| `openspec/features/runeforge.feature` | ✅ Completo |
| `openspec/features/agents.feature` | ✅ Completo |
| `openspec/features/memory.feature` | ✅ Completo |
| `runes/sdd-spec/RUNE.md` | ✅ **Incluye sección BDD** (líneas 137-176) — `## BDD Feature File Template (Required Deliverable)` con cuándo generarlo y dónde guardarlo |

### OPA + Runes

| Archivo | Estado |
|---------|--------|
| `rules/security.rego` | ✅ 89 líneas — sandbox, WASM, script, semver, `rm -rf`, `sudo` |
| `rules/rune-validation.rego` | ✅ Valida RUNE.md estructura, semver, descripción |
| `rules/data.json` | ✅ |
| `runes/` (13) | ✅ Completos incluyendo sdd-propose |

---

## Milestones — Estado Final Definitivo

| Milestone | Estado | Completitud |
|-----------|--------|-------------|
| M1 — Runes Registry | ✅ | 100% |
| M2 — Catalog + Pipeline | ✅ | 100% |
| M3 — RuneForge + Multi-Agent | ✅ | 100% |
| M4 — Mimir Embeddings | ✅ | 100% |
| M5 — Auto-Update + TUI Status | ✅ | 100% |
| M6 — BDD + e2e godog | ✅ | 100% |
| M7 — OPA Rules + Heimdall | ✅ | 100% |
| M8 — Makefile Targets | ✅ | 100% |
| M9 — Migrate | ✅ | 100% |

---



## Todo completado ✅

El ecosistema ODIN ha alcanzado paridad total con gentle-ai y la metodología TDD·BDD·SDD está completamente operativa. Todos los tests pasan en verde.

```
internal/migrate  — 18 tests PASS
internal/memory   — 37 tests PASS (incluye TestMimirSearch_SemanticRelevance + BenchmarkSemanticSearch)
e2e/              — godog BDD PASS contra 5 feature files
internal/guardian — unit tests PASS
```


## Instalador Remoto

```powershell
irm https://raw.githubusercontent.com/andragon31/ODIN-AI/main/scripts/install.ps1 | iex
```

El instalador soporta:
- **Method auto** (default): descarga binario pre-compilado desde GitHub Releases
- **Method go**: `go install github.com/andragon31/odin-ai/cmd/odin@latest`
- **Method binary**: fuerza descarga directa
- Verificación SHA256 via checksums.txt
- Detección de plataforma (amd64, arm64)
- Agregado automático al PATH

---

## Lo que ODIN tiene que gentle-ai no tiene

| Capacidad | gentle-ai | ODIN |
|-----------|-----------|------|
| Testing | Sin metodología específica | TDD + BDD (godog) + SDD integrados |
| Seguridad | AGENTS.md rules (texto) | OPA Rego policies (código evaluable) + SAST (gosec + semgrep) |
| Pre-commit hook | No tiene | `odin heimdall hook-install` |
| Memoria semántica | FTS5 únicamente | sqlite-vss + OllamaEmbedder 768 dims + fallback TF-IDF |
| Rune generation | Depende de agente externo (Claude/OpenCode) | RuneForge offline con Ollama |
| Multi-agente | Agente único | 6 agentes + detector automático en PATH + fallback chain |
| Sync | No tiene | Bifrost (go-git + CRDT) |
| Dashboard | No tiene | `odin status` con estado de 8 componentes del ecosistema (TUI + JSON) |
| CI/CD reports | No tiene | `odin heimdall report generate` → JSON para pipelines |
| Migración desde gentle-ai | N/A | `odin migrate --from gentle` importa config, memories, skills, AGENTS.md → Rego |

---

## Comandos disponibles (verificados)

```bash
# Testing
make test-unit          # Go tests estándar con -race y cobertura
make test-bdd           # godog contra openspec/features/ (BDD real)
make test-race          # Solo race detector
make test-all           # unit + integration + race
make test-coverage      # Genera coverage.html
make validate-coverage  # Falla si cobertura < 80%

# Guardian (Heimdall)
odin heimdall check [files...]          # SAST + OPA en archivos/staged
odin heimdall runes [paths...]          # Evalúa runes contra security.rego
odin heimdall hook-install              # Instala pre-commit hook
odin heimdall hook-uninstall            # Remueve pre-commit hook
odin heimdall report generate           # JSON para CI/CD
odin heimdall status                    # Estado del guardian

# Status del ecosistema
odin status                             # Dashboard TUI completo
odin status --json                      # JSON de todos los componentes

# Migración
odin migrate --from gentle              # Importa desde gentle-ai
odin migrate --from gentle --dry-run    # Preview sin cambios
```

---

## Criterio de Done Universal (Metodología TDD·BDD·SDD)

```
[ ] SDD: spec + design + tasks completados en Engram
[ ] BDD: Feature file en openspec/features/{change}.feature
[ ] TDD: make test-unit → Green (cobertura ≥ 80%)
[ ] BDD: make test-bdd → todos los godog scenarios pasan
[ ] Race: make test-race → sin data races
[ ] SDD: sdd-verify → sin CRITICAL ni WARNING
[ ] Security: odin heimdall runes → todos los runes pasan
```

---

*Versión 5.1.0 — Roadmap 100% completado + Instalador remoto, Abril 2026*
