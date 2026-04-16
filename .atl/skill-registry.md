# ODIN AI - Skill Registry

## Project Overview

**ODIN AI** es el ecosistema nórdico local-first, evolucionando el ecosistema Gentleman AI para funcionar 100% offline con capacidades enterprise.

## Skills

### Project-Level Skills

| Skill | Trigger Context | Purpose |
|-------|-----------------|---------|
| `go-testing` | `*.go`, `go test`, `testing` | Go testing patterns with teatest |
| `sdd-init` | `sdd init`, project bootstrap | Initialize SDD context |
| `sdd-apply` | `sdd-apply`, implementation | Implement code from tasks |
| `sdd-verify` | `sdd-verify`, validation | Validate implementation against specs |

### Skill Triggers

| Context | Skill |
|---------|-------|
| Writing Go tests | `go-testing` |
| Initializing SDD | `sdd-init` |
| Implementing tasks | `sdd-apply` |
| Verifying implementation | `sdd-verify` |

## Stack

- **Language**: Go 1.24+
- **CLI**: cobra + viper
- **TUI**: bubbletea + lipgloss
- **Testing**: go test, testcontainers-go
- **Linting**: golangci-lint
- **Formatter**: gofmt, goimports

## Conventions

### Git Commits
- Follow Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`
- **NO** Co-Authored-By
- **NO** AI attribution

### Code Style
- Format: `gofmt` + `goimports`
- Lint: `golangci-lint`
- Tests: `go test -race -cover`

## File Structure

```
odin-ecosystem/
├── cmd/odin/           # CLI entrypoint
├── internal/
│   ├── cli/            # Command-line interface
│   ├── config/         # Configuration management
│   ├── orchestrator/   # SDD orchestration
│   ├── router/         # Model routing
│   ├── guardian/       # Heimdall security
│   ├── memory/         # Mimir persistence
│   ├── sync/           # Bifrost sync
│   ├── skills/         # Runes registry
│   ├── plugins/        # WASM runtime
│   └── verify/         # Nornir testing
├── pkg/
│   └── logger/         # Structured logging
├── deploy/             # Dvergar
├── docs/               # Documentation
├── runes/              # Skills base
├── e2e/                # E2E tests
├── themes/             # Völva themes
├── rules/              # Heimdall OPA policies
├── go.mod
├── go.work
├── Makefile
└── AGENTS.md
```

## Components

| Norse God | Component | Status |
|-----------|-----------|--------|
| Odin | Core | ✅ Base CLI created |
| Heimdall | Security | 🔲 Not started |
| Mimir | Memory | 🔲 Not started |
| Runes | Skills | 🔲 Not started |
| Bifrost | Sync | 🔲 Not started |
| Völva | UI | 🔲 Not started |
| Nornir | Testing | 🔲 Not started |
| Dvergar | Deploy | 🔲 Not started |

## Notes

- Last updated: 2026-04-12
- Project is in Sprint 1 (Odin Core)
- Testing capabilities pending (project is new)
