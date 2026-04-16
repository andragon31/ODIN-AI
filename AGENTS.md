# ODIN AI - Agents Configuration

## Project Overview

**ODIN AI** es el ecosistema nórdico local-first, evolucionando el ecosistema Gentleman AI para funcionar 100% offline con capacidades enterprise.

## Architecture

```
odin (CLI)
├── cmd/odin/           # Entry point
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
└── pkg/
    └── logger/         # Structured logging
```

## SDD Phases

ODIN sigue Spec-Driven Development con 9 fases:

1. **explore** — Investigar ideas antes de comprometerse
2. **propose** — Crear propuesta de cambio
3. **spec** — Escribir especificaciones con requisitos
4. **design** — Documento técnico de arquitectura
5. **tasks** — Desglose en tareas de implementación
6. **apply** — Implementar código siguiendo specs
7. **verify** — Validar que implementación matchea specs
8. **archive** — Archivar cambios completados

## Supported Platforms

- ✅ Ubuntu 20.04+ (amd64, arm64)
- ✅ macOS 12+ (Intel, Apple Silicon)
- ✅ Windows 10+ (WSL2, native)
- ✅ Arch Linux
- ✅ Fedora 36+
- ✅ Docker

## Licensing

- **Code**: MIT
- **Infrastructure/Plugins**: Apache 2.0
- **Documentation**: CC-BY-SA 4.0

## Quick Start

```bash
# Build
make build

# Run
./build/odin init
./build/odin status

# Test
make test-unit
```

## Development Conventions

### Git Commits

Follow Conventional Commits:
- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation
- `refactor:` — Code refactoring
- `test:` — Adding tests
- `chore:` — Maintenance

**NO** Co-Authored-By, **NO** AI attribution in commits.

### Code Style

- Format: `gofmt` + `goimports`
- Lint: `golangci-lint`
- Vet: `go vet`
- Tests: `go test -race -cover`

### Pull Requests

1. Create feature branch: `feat/feature-name`
2. Run tests: `make test`
3. Run linter: `make lint`
4. Submit PR with description
5. Require 1 approval before merge

## Components Map

| Norse God | Component | Function |
|-----------|-----------|----------|
| Odin | Core | Orchestrator, router |
| Heimdall | Security | Guardian, SAST, OPA |
| Mimir | Memory | Vector search, persistence |
| Runes | Skills | Registry, validation |
| Bifrost | Sync | CRDT, Git-backed |
| Völva | UI | TUI, themes, a11y |
| Nornir | Testing | E2E, benchmarks |
| Dvergar | Deploy | Install, upgrade, rollback |
