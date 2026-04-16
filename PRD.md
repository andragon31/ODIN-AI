# PRD: ODIN AI Ecosystem

> **Nórdico Local-First · 100% OSS · $0 Costo Infra · Para la Banda**

**Versión**: 1.0.0-draft  
**Autor**: Andragon  
**Base**: Gentleman AI Ecosystem  
**Licencia Objetivo**: MIT (código), Apache 2.0 (infra/plugins), CC-BY-SA 4.0 (docs)  
**Fecha**: Abril 2026  
**Estado**: Draft — Listo para revisión

---

## Tabla de Contenidos

1. [Resumen Ejecutivo](#1-resumen-ejecutivo)
2. [Declaración del Problema](#2-declaración-del-problema)
3. [Visión y Principios](#3-visión-y-principios)
4. [Mapa de Homólogos Nórdicos](#4-mapa-de-homólogos-nórdicos)
5. [Análisis Detallado por Componente](#5-análisis-detallado-por-componente)
   - [5.1 Odin AI (Core Orchestrator)](#51-odin-ai-core-orchestrator)
   - [5.2 Heimdall (Guardian Security)](#52-heimdall-guardian-security)
   - [5.3 Mimir (Memory Engine)](#53-mimir-memory-engine)
   - [5.4 Runes (Skills Registry)](#54-runes-skills-registry)
   - [5.5 Bifrost (Sync Engine)](#55-bifrost-sync-engine)
   - [5.6 Völva (Interface Engine)](#56-völva-interface-engine)
   - [5.7 Nornir (Verification Suite)](#57-nornir-verification-suite)
   - [5.8 Dvergar (Forge/Deploy)](#58-dvergar-forgedeploy)
6. [Arquitectura del Sistema](#6-arquitectura-del-sistema)
7. [Flujo de Datos Local-First](#7-flujo-de-datos-local-first)
8. [Modelo de Datos](#8-modelo-de-datos)
9. [API & Interfaces](#9-api--interfaces)
10. [Seguridad, Privacidad y Licencias](#10-seguridad-privacidad-y-licencias)
11. [Roadmap por Sprint](#11-roadmap-por-sprint)
12. [Criterios de Aceptación](#12-criterios-de-aceptación)
13. [Glosario](#13-glosario)
14. [Apéndice: Migración desde Gentle AI](#14-apéndice-migración-desde-gentle-ai)

---

## 1. Resumen Ejecutivo

### 1.1 ¿Qué es ODIN AI?

**ODIN AI** es la evolución **local-first** del ecosistema Gentleman AI. Donde Gentle AI configura agentes de IA en la nube, ODIN AI lleva toda esa inteligencia a tu máquina — sin dependencias de nube, sin costos de infraestructura, sin telemetría oculta.

```
Gentle AI: "Tu agente tiene superpoderes, pero necesita la nube."
ODIN AI:   "Tus superpoderes ahora viven en tu máquina. Offline. Libre."
```

### 1.2 ¿Qué Problema Resuelve?

En 2026, cada desarrollador usa al menos un agente de IA. Pero:

- **Gentle AI** requiere configuración en la nube yAPI keys externas
- **Los agentes de IA** sin configuración son como autos deportivos sin afinar — funcionan, pero no rinden
- **La memoria, skills y workflow** son difíciles de configurar y reproducir entre máquinas

ODIN AI resuelve esto con:

| Problema | Solución ODIN |
|----------|--------------|
| Dependencia de API keys externas | Modelos locales via Ollama como fallback primario |
| Sin memoria persistente | Mimir con búsqueda vectorial sqlite-vss |
| Skills estáticos sin validación | Runes con schema CUE + sandbox testing |
| Sincronización frágil | Bifrost con CRDT y Git-backed config |
| Sin verificación de seguridad | Heimdall con OPA + SAST local |
| Instalación compleja | Dvergar: `curl \| sh` con verificación criptográfica |

### 1.3 Pilares Fundamentales

```
🏔️ LOCAL-FIRST    Todo funciona offline. La nube es opcional.
🔓 100% OSS        Zero licencias propietarias.
📦 $0 COSTO        Sin costos de infraestructura.
🔐 PRIVACY-FIRST   Cifrado en reposo, sandbox plugins, sin telemetría.
🤝 GENTLE COMPAT   Migración automática desde Gentle AI.
```

---

## 2. Declaración del Problema

### 2.1 Contexto del Ecosistema Gentleman AI

Gentleman AI resolvió el problema de configurar agentes de IA correctamente. Su ecosistema incluye:

- **gentle-ai**: CLI + orquestador SDD para configuración
- **Engram**: Memoria persistente SQLite con FTS5
- **SDD**: Workflow de 9 fases (Spec-Driven Development)
- **GGA (Guardian Angel)**: Pre-commit + AI review
- **Skills Library**: Workflows en Markdown estáticos
- **MCP Servers**: Context7, Notion, Jira

### 2.2 Limitaciones de Gentle AI

| Componente | Limitación |
|------------|------------|
| **gentle-ai** | Routing atado a un agente, sin fallback automático, estado volátil entre fases |
| **Engram** | Solo keywords, sin embeddings vectoriales, sin cifrado |
| **Skills** | Estáticos, sin validación runtime, errores rompen pipeline |
| **Config Sync** | Conflictos silenciosos, paths incompatibles cross-platform |
| **GGA** | Reglas hardcodeadas, dependencia de proveedor externo |
| **TUI** | Tema fijo, sin accesibilidad, no extensible |

### 2.3 ¿Por Qué ODIN AI?

ODIN AI no es un fork — es una **reinterpretación local-first** que:

1. **Potencia cada componente** con capacidades locales (vector search, CRDT, WASM sandbox)
2. **Elimina dependencias de nube** donde sea posible
3. **Añade capas de seguridad** (cifrado, OPA, SAST)
4. **Mantiene compatibilidad** con Gentle AI via comando de migración

---

## 3. Visión y Principios

### 3.1 Declaración de Visión

> *"Evolucionar cada herramienta del ecosistema Gentleman AI en un homólogo nórdico local-first, open-source y sin dependencias de nube. Mantener la simplicidad de instalación de Gentle, pero elevar la arquitectura a nivel enterprise mediante motores vectoriales locales, políticas como código, sincronización CRDT y runtime seguro WASM."*

### 3.2 Principios Fundamentales

| Principio | Implicación Técnica | Justificación |
|-----------|---------------------|---------------|
| **🏔️ Local-First** | Todo funciona offline. La nube es opcional (sync/observabilidad) | Privacidad, costo, disponibilidad |
| **🔓 100% OSS** | Stack validado por licencia (MIT/Apache/BSD/MPL) | Transparencia, libertad, auditabilidad |
| **📦 Instalación Progresiva** | `local` (1 binario) → `docker` (stack completo) → `cluster` (K8s) | Escalabilidad sin fricción |
| **🛡️ Privacidad por Defecto** | Cifrado en reposo, sandbox plugins, sin telemetría, keys locales | Confianza, compliance, seguridad |
| **🔄 Compatibilidad Gentle** | `odyn migrate --from gentle` sin breaking changes en v1.0 | Migración suave, comunidad |
| **🤝 Para la Banda** | Docs claros, $0 costo, gobernanza abierta | Accesibilidad, comunidad |

### 3.3 Non-Goals (Lo Que NO Es)

- ❌ No es un agente de IA nuevo — es un orquestador de ecosistema
- ❌ No reemplaza a Gentle AI — lo complementa para casos local-first
- ❌ No requiere GPU — funciona en CPU con Ollama (GPU es opcional)
- ❌ No es solo para desarrolladores avanzados — para la banda significa simpleza

---

## 4. Mapa de Homólogos Nórdicos

### 4.1 Tabla de Homología

| Componente Gentle | Función Actual | Homólogo Nórdico | Rol en ODIN |
|-------------------|----------------|------------------|-------------|
| `gentle-ai` | CLI + Orquestador SDD | **Odin AI** 🐺 | Core orchestrator, router universal, state manager |
| `GGA` (Guardian) | Pre-commit + AI review | **Heimdall** 🌉 | Guardian security, policy-as-code, local SAST |
| `Engram` | Memoria persistente SQLite | **Mimir** 🧠 | Memory engine, vector search, knowledge graph |
| `Skills Library` | Workflows `.md` estáticos | **Runes** 🔮 | Skills registry, validation, sandbox testing |
| `Config Sync` | Sincronización cross-agent | **Bifrost** 🌈 | CRDT sync engine, conflict resolution, Git-backed |
| `TUI` | Interfaz Bubbletea | **Völva** 🕳️ | Interface engine, themes, accessibility, plugin UI |
| `E2E Pipeline` | Testing en contenedores | **Nornir** 📜 | Verification suite, flaky detection, local benchmarks |
| `Installers` | Scripts de despliegue | **Dvergar** ⚒️ | Progressive forge, atomic upgrades, rollback, cosign |

### 4.2 Metáforas Nórdicas

| Dios/Ser Mitológico | Significado | Aplicación en ODIN |
|--------------------|-------------|--------------------|
| **Odin** | Dios de la sabiduría, la guerra y la poesía | Orquestador central, router de modelos |
| **Heimdall** | Guardián del Bifröst, ve a lo lejos | Security guardian, SAST, policy enforcement |
| **Mímir** | Ser de gran sabiduría, fuente de conocimiento | Motor de memoria, búsqueda vectorial |
| **Runes** | Alfabeto mágico con poder intrínseco | Skills con validación semántica |
| **Bifröst** | Puente arcoíris entre mundos | Sync engine, CRDT, multi-device |
| **Völva** | Chamán femenino con poderes de visión | Interface engine, TUI, accesibilidad |
| **Nornir** | Tres hermanas que tejen el destino | Verification suite, testing, fate (benchmarks) |
| **Dvergar** | Enanos maestros herreros | Installers, forge, atomic upgrades |

---

## 5. Análisis Detallado por Componente

---

### 5.1 Odin AI (Core Orchestrator)

#### Descripción

El núcleo del ecosistema ODIN. Maneja el enrutamiento de modelos, gestión de estado entre fases SDD, y la ejecución de plugins WASM.

#### Análisis Comparativo (Gentle → ODIN)

| Aspecto | Gentle (`gentle-ai`) | ODIN (`odin`) |
|---------|---------------------|---------------|
| **Routing** | Atado a un agente | Tabla dinámica por fase + fallback chain |
| **Estado** | Volátil entre fases | Persistente en `~/.odin/sessions/` con snapshots Git |
| **Extensibilidad** | Sin runtime de extensiones | WASM plugin runtime (`wasmtime-go`) |
| **Fallback** | Sin fallback automático | Ollama → económico → premium |
| **Instalación** | Multi-step | 1 binario estático, <50MB |
| **Modo Local** | No existe | `local` puro sin red |

#### Stack Tecnológico

```
✅ cobra           (MIT)  — CLI framework
✅ viper           (MIT)  — Configuration management
✅ wasmtime-go     (Apache 2.0) — WASM runtime
✅ go-crdt         (MIT)  — CRDT operations
✅ ollama/ollama   (MIT)  — Local model inference
✅ go-git          (Apache 2.0) — Git operations
```

#### Criterios de Aceptación

```markdown
- [ ] `odin status` retorna JSON/TUI válido en <800ms
- [ ] Fallback automático en ≤3 intentos si proveedor falla
- [ ] Plugin WASM ejecuta sin red/FS por defecto (sandbox)
- [ ] `odin migrate --from gentle` copia configs sin romper
- [ ] Estado persiste entre fases SDD en `~/.odin/sessions/`
- [ ] Modo local puro: 1 binario, <50MB, sin red requerida
```

#### Comandos CLI

```bash
odin init                    # Inicializar entorno ODIN
odin status                  # Mostrar salud del ecosistema
odin session list            # Listar sesiones SDD activas
odin session resume <id>     # Reanudar sesión desde snapshot
odin migrate --from gentle   # Migrar desde Gentle AI
odin plugin install <path>   # Instalar plugin WASM
odin router set <provider>   # Configurar router de modelos
odin router fallback add <provider>  # Añadir fallback
```

---

### 5.2 Heimdall (Guardian Security)

#### Descripción

Guardian de seguridad que reemplaza GGA. Ejecuta análisis SAST local con OPA (Rego) y herramientas OSS como semgrep y gosec.

#### Análisis Comparativo (Gentle → ODIN)

| Aspecto | Gentle (`GGA`) | ODIN (`Heimdall`) |
|---------|----------------|-------------------|
| **Reglas** | Hardcodeadas | Policy-as-Code con OPA (Rego) |
| **Motor SAST** | No existe | gosec + semgrep OSS |
| **Revisión IA** | Depende de proveedor externo | Multi-model con fallback local |
| **Bloqueo** | No bloquea commits inseguros | Hook pre-commit bloquea vulnerabilidades críticas |
| **CI/CD** | Reporte básico | Reporte JSON para pipelines |

#### Stack Tecnológico

```
✅ opa            (Apache 2.0) — Policy engine (Rego)
✅ semgrep        (LGPLv2.1) — Static analysis OSS
✅ gosec          (Apache 2.0) — Go security checker
✅ pre-commit     (MIT) — Git hooks framework
✅ ollama         (MIT) — Local model fallback
```

#### Criterios de Aceptación

```markdown
- [ ] `heimdall check` analiza diff + aplica reglas Rego en <2s
- [ ] Fallback a modelo local si API externa falla
- [ ] Hook pre-commit bloquea commits con vulnerabilidades críticas
- [ ] Reporte JSON para CI/CD pipelines (exit 0/1/2)
- [ ] Reglas OPA customizables en `~/.odin/rules/`
- [ ] Soporte para reglas custom por proyecto
```

#### Flujo de Ejecución

```
git commit → pre-commit hook → Heimdall check
                                       │
                    ┌──────────────────┼──────────────────┐
                    ▼                  ▼                  ▼
              [PASSED]           [WARNING]          [BLOCKED]
              (commit ok)      (commit + log)    (commit rejected)
```

---

### 5.3 Mimir (Memory Engine)

#### Descripción

Motor de memoria persistente que evoluciona Engram con búsqueda semántica vectorial y cifrado.

#### Análisis Comparativo (Gentle → ODIN)

| Aspecto | Gentle (`Engram`) | ODIN (`Mimir`) |
|---------|-------------------|----------------|
| **Búsqueda** | Keywords (FTS5) | Semántica (sqlite-vss embeddings) |
| **Contexto** | Aislado por proyecto | Cross-project con tags |
| **Cifrado** | No | age encryption en reposo |
| **Motor Vectorial** | No existe | sqlite-vss (local) o pgvector (docker) |
| **Knowledge Graph** | No | Aristas entre recuerdos |
| **Retención** | Manual | Pruning automático inteligente |

#### Stack Tecnológico

```
✅ go-sqlite3     (MIT)  — SQLite Go bindings
✅ sqlite-vss    (MIT)  — Vector search extension
✅ pgvector       (PostgreSQL) — Vector search (docker opcional)
✅ age           (BSD 3) — Encryption
✅ go-crdt       (MIT)  — CRDT sync
```

#### Modelo de Datos

```sql
-- Memories table
CREATE TABLE memories (
    id          TEXT PRIMARY KEY,
    content     TEXT NOT NULL,
    embedding   BLOB,                    -- Vector embedding (sqlite-vss)
    project     TEXT,
    tags        TEXT[],                  -- ['arch', 'spec', 'security', ...]
    created_at  DATETIME DEFAULT NOW,
    updated_at  DATETIME DEFAULT NOW,
    accessed_at  DATETIME DEFAULT NOW,
    encrypted   BOOLEAN DEFAULT FALSE
);

-- Knowledge graph edges
CREATE TABLE memory_edges (
    from_id     TEXT REFERENCES memories(id),
    to_id       TEXT REFERENCES memories(id),
    relation    TEXT,                    -- 'inspired_by', 'depends_on', 'contradicts'
    weight      FLOAT,
    PRIMARY KEY (from_id, to_id, relation)
);

-- Full-text search
CREATE VIRTUAL TABLE memories_fts USING fts5(content, content='memories');
```

#### Criterios de Aceptación

```markdown
- [ ] Búsqueda semántica <200ms en 10k registros
- [ ] `mimir encrypt/decrypt` cifra/restaura sin pérdida
- [ ] `mimir sync --remote` mergea sin conflictos (CRDT)
- [ ] Pruning automático en background (conserva tags: arch, spec, security)
- [ ] Knowledge graph query: `mimir graph --from <memory-id>`
- [ ] Fallback a FTS5 si sqlite-vss no disponible
```

#### Comandos CLI

```bash
mimir store <content>              # Almacenar con embedding automático
mimir search --query <text>       # Búsqueda semántica
mimir search --query <text> --limit 10  # Con límite
mimir recall --id <memory-id>      # Recuperar por ID
mimir tags list                    # Listar tags disponibles
mimir tags add <memory-id> <tag>  # Añadir tag
mimir prune --keep-tags arch,spec,security  # Pruning inteligente
mimir encrypt --key <key-file>    # Cifrar base de datos
mimir decrypt --key <key-file>    # Descifrar base de datos
mimir sync push                    # Sync a remote
mimir sync pull                   # Pull desde remote
mimir graph --from <memory-id>    # Query knowledge graph
```

---

### 5.4 Runes (Skills Registry)

#### Descripción

Sistema de skills que evoluciona la Skills Library con validación schema CUE y sandbox testing.

#### Análisis Comparativo (Gentle → ODIN)

| Aspecto | Gentle (Skills) | ODIN (Runes) |
|---------|-----------------|--------------|
| **Formato** | Markdown estático | Markdown + Schema CUE |
| **Validación** | Ninguna runtime | CUE validation antes de cargar |
| **Descubrimiento** | Manual | `odin rune search` |
| **Testing** | No | Sandbox con output diff |
| **Versioning** | No | Semver con rollback |
| **Cache** | No | Local para evitar red |

#### Stack Tecnológico

```
✅ cue            (Apache 2.0) — Schema validation
✅ go-yaml        (MIT) — YAML parsing
✅ wasmtime       (Apache 2.0) — Sandboxed execution
```

#### Schema CUE para Skills

```cue
# RuneSkill: schema para validación de skills
# Validación: cue validate rune.cue

# Skill metadata
name:        string @regex(#^[a-z][a-z0-9-]+$#)
version:     string @regex(#^\d+\.\d+\.\d+$#)
description: string
author?:     string
tags:        [...string]

# Trigger conditions (when is this skill relevant?)
triggers: {
    filePatterns?: [...string]  // e.g., ["*.go", "go.mod"]
    commands?:    [...string]   // e.g., ["test", "build"]
    context?:     [...string]   // e.g., ["testing", "CI"]
}

# Execution
execution: {
    type:   "prompt" | "script" | "wasm"
    prompt?: string
    script?: string
    wasm?:   string  // path to .wasm file
    sandbox: bool @default(true)
}

# Expected outputs
outputs?: {
    files?:   [...string]
    console?: string
    errors?:  [...string]
}
```

#### Criterios de Aceptación

```markdown
- [ ] `odin rune install` valida schema CUE antes de escribir
- [ ] Skill malformado no bloquea SDD, solo warning logueado
- [ ] `odin rune test` retorna pass/fail con diff de salida
- [ ] Cache local evita red para skills frecuentes
- [ ] `odin rune search --tags testing,CI` filtra por tags
- [ ] `odin rune rollback <name>@<version>` revierte a versión anterior
```

#### Comandos CLI

```bash
odin rune search <query>          # Buscar skills
odin rune search --tags testing   # Filtrar por tags
odin rune install <path|url>      # Instalar desde path o URL
odin rune list                    # Listar skills instalados
odin rune test <name>             # Testear skill en sandbox
odin rune validate <path>         # Validar schema CUE
odin rune remove <name>           # Desinstalar skill
odin rune rollback <name>@<version>  # Rollback a versión
```

---

### 5.5 Bifrost (Sync Engine)

#### Descripción

Motor de sincronización que reemplaza Config Sync con CRDT para resolución de conflictos.

#### Análisis Comparativo (Gentle → ODIN)

| Aspecto | Gentle (Config Sync) | ODIN (Bifrost) |
|---------|---------------------|----------------|
| **Conflictos** | Silenciosos (sobrescribe) | CRDT merge automático |
| **Historial** | No | Git completo |
| **Ramas** | No | Soporte multi-rama |
| **Paths** | Incompatible WSL/macOS | Cross-platform paths |
| **Firma** | No | GPG opcional para commits |

#### Stack Tecnológico

```
✅ go-git          (Apache 2.0) — Git operations
✅ go-crdt         (MIT) — CRDT implementation
✅ golang.org/x/crypto/ssh  (BSD) — SSH for Git
```

#### Criterios de Aceptación

```markdown
- [ ] `odin sync init` crea repo local en `~/.odin/config/`
- [ ] Conflictos resueltos automáticamente o marcados para revisión
- [ ] `odin sync diff` muestra cambios antes de aplicar
- [ ] Funciona en WSL2, macOS ARM, Linux x86 sin paths hardcodeados
- [ ] `odin sync log` muestra historial de cambios
- [ ] GPG signing opcional para commits de config
```

#### Comandos CLI

```bash
odin sync init                     # Inicializar repo local
odin sync push                     # Push a remote
odin sync pull                     # Pull desde remote
odin sync status                   # Estado actual
odin sync diff                     # Diff antes de aplicar
odin sync log                      # Historial de cambios
odin sync branch list              # Listar ramas
odin sync branch create <name>     # Crear rama
odin sync merge <branch>           # Mergear rama
odin sync sign on                  # Activar GPG signing
```

---

### 5.6 Völva (Interface Engine)

#### Descripción

Motor de interfaz TUI que evoluciona la TUI de Gentle con temas, accesibilidad y modos de salida.

#### Análisis Comparativo (Gentle → ODIN)

| Aspecto | Gentle (TUI) | ODIN (Völva) |
|---------|--------------|--------------|
| **Tema** | Fijo (Rose Pine) | Engine de temas intercambiables |
| **Accesibilidad** | No | Alto contraste + screen reader |
| **Modos** | Solo interactivo | `--json`, `--quiet`, `--interactive` |
| **Extensibilidad** | No | Plugin puede inyectar componentes via WASM |
| **Navegación** | Básico | vim-style (j/k, Enter, Esc) |

#### Stack Tecnológico

```
✅ bubbletea      (MIT) — TUI framework (Charm)
✅ lipgloss      (MIT) — Styling (Charm)
✅ huh           (MIT) — Forms/inputs (Charm)
```

#### Temas Soportados

```yaml
# ~/.odin/themes/rose-pine.toml
[theme]
name = "Rose Pine"
bg = "#1a1b26"
card = "#232136"
text = "#e0def4"
accent = "#c4a7e7"

# ~/.odin/themes/nord.toml
[theme]
name = "Nord"
bg = "#2e3440"
card = "#3b4252"
text = "#eceff4"
accent = "#88c0d0"
```

#### Criterios de Aceptación

```markdown
- [ ] `völva theme set rose-pine|nord|custom.toml` cambia tema sin reiniciar
- [ ] `--json` produce salida serializable válida
- [ ] Soporte básico screen reader + navegación por teclado
- [ ] Plugin puede inyectar TUI component via WASM hook
- [ ] `--quiet` silencia output excepto errores
- [ ] `--interactive` forza modo TUI incluso en pipe
```

#### Comandos CLI

```bash
völva theme list                    # Listar temas disponibles
völva theme set <name>              # Cambiar tema
völva theme preview                 # Preview todos los temas
odin --json                         # Output JSON
odin --quiet                        # Modo silencioso
odin --interactive                  # Forzar TUI
```

---

### 5.7 Nornir (Verification Suite)

#### Descripción

Suite de verificación que evoluciona E2E Pipeline con detección de tests flaky y benchmarks.

#### Análisis Comparativo (Gentle → ODIN)

| Aspecto | Gentle (E2E) | ODIN (Nornir) |
|---------|--------------|---------------|
| **Contenedores** | Pesados | testcontainers-go efímeros + cleanup |
| **Flaky Tests** | No detecta | Re-ejecución + umbral de consistencia |
| **Benchmarks** | No | k6 + pprof locales |
| **Reportes** | Básico | JSON para dashboard local |
| **Matrix** | Docker | Multi-OS testing |

#### Stack Tecnológico

```
✅ testcontainers-go  (MIT) — Ephemeral containers
✅ k6               (Apache 2.0) — Load testing
✅ go-pprof         (BSD) — Profiling
```

#### Criterios de Aceptación

```markdown
- [ ] `make test-e2e` corre en <15 min en CI
- [ ] Tests flaky marcados con `@flaky` y reportados
- [ ] Benchmarks de latencia <2s para fases críticas
- [ ] Cleanup automático de contenedores post-run
- [ ] `nornir report --format json|html` genera reporte
- [ ] Matrix testing: Ubuntu, Arch, macOS, WSL2
```

---

### 5.8 Dvergar (Forge/Deploy)

#### Descripción

Sistema de instalación que evoluciona Installers con verificación criptográfica y rollback atómico.

#### Análisis Comparativo (Gentle → ODIN)

| Aspecto | Gentle (Installers) | ODIN (Dvergar) |
|---------|---------------------|----------------|
| **Verificación** | Frágil (sha256) | cosign + sigstore |
| **Rollback** | No | Automático en ≤3s si falla |
| **Backup** | No | Atómico pre-upgrade |
| **Detección SO** | Básica | Inteligente (SO, arch, Docker/K8s) |
| **Instalación** | Multi-step | One-liner con detección automática |

#### Stack Tecnológico

```
✅ cosign        (Apache 2.0) — Binary verification
✅ sigstore      (Apache 2.0) — Transparency log
✅ bash          (GPL) — Installer script
✅ tar/gzip      (GPL) — Package format
```

#### Criterios de Aceptación

```markdown
- [ ] Instalación limpia en <2 min en Ubuntu/macOS/WSL2
- [ ] `cosign verify` pasa antes de ejecutar
- [ ] `odin upgrade` backup + migrate + rollback automático
- [ ] `odin status` post-install muestra stack saludable
- [ ] Detección automática: SO, arquitectura, Docker/Podman/K8s
- [ ] Log de instalación en `~/.odin/logs/`
```

#### One-Liner de Instalación

```bash
curl -fsSL https://get.odin.ai/install.sh | bash -s -- --mode auto
```

---

## 6. Arquitectura del Sistema

### 6.1 Estructura de Repositorio (Monorepo Inicial)

```
odin-ecosystem/
├── cmd/
│   └── odin/                  # CLI entrypoint (Odin + Völva)
├── internal/
│   ├── orchestrator/          # SDD + State + Git snapshots
│   ├── router/                # Universal model routing
│   ├── guardian/              # Heimdall (OPA + SAST)
│   ├── memory/                # Mimir (sqlite-vss + age)
│   ├── sync/                 # Bifrost (go-git + CRDT)
│   ├── skills/               # Runes (CUE + sandbox)
│   ├── plugins/               # WASM runtime + hooks
│   └── verify/               # Nornir (testcontainers + k6)
├── deploy/                    # Dvergar (install.sh + cosign)
├── docs/                      # Docusaurus + playgrounds
├── runes/                     # Skills registry base
├── e2e/                       # Matrix tests
├── themes/                    # Völva themes
├── rules/                     # Heimdall OPA policies
├── go.work
└── Makefile
```

### 6.2 Configuración Unificada

**Ubicación**: `~/.odin/config.yaml`

```yaml
version: "1.0"
mode: "local"                    # local | docker | cluster

memory:
  engine: "sqlite-vss"           # sqlite-vss | pgvector
  path: "~/.odin/memory.db"
  encryption: true
  pruning:
    keep_tags: ["arch", "spec", "security"]
    interval: "24h"

sync:
  backend: "git"
  remote: "git@github.com:user/odin-config.git"
  auto_push: false
  gpg_sign: false

guardian:
  policy_engine: "opa"
  rules_path: "~/.odin/rules/"
  saast:
    enabled: true
    tools: ["gosec", "semgrep"]
  block_on_critical: true

router:
  default: "ollama-local"
  fallback:
    - "openrouter"
    - "anthropic"
  cost_cap_daily: 0.0            # 0 = unlimited local

observability:
  metrics_port: 9090
  log_level: "info"
  log_path: "~/.odin/logs/"

plugins:
  sandbox: true
  auto_update: false
  allowed_paths: ["~/.odin/plugins/"]

themes:
  current: "rose-pine"
  path: "~/.odin/themes/"

session:
  path: "~/.odin/sessions/"
  snapshot_interval: "5m"
  max_sessions: 10
```

### 6.3 Arquitectura de Componentes

```
┌─────────────────────────────────────────────────────────────────┐
│                         USUARIO / TERMINAL                       │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Völva (TUI/JSON/CLI)                       │
│                   ┌─────────────────────────┐                   │
│                   │ --interactive (default) │                   │
│                   │ --json (scripts)        │                   │
│                   │ --quiet (CI)            │                   │
│                   └─────────────────────────┘                   │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Odin AI (Orchestrator)                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │ SDD State    │  │ Router       │  │ Plugin Host  │         │
│  │ Manager      │  │ (fallback)   │  │ (WASM)       │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└───────────────────────────────┬─────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│   Heimdall    │     │    Mimir      │     │   Bifrost     │
│   (Security)   │     │   (Memory)    │     │    (Sync)     │
│  ┌───────────┐ │     │  ┌─────────┐ │     │  ┌─────────┐ │
│  │ OPA       │ │     │  │sqlite-  │ │     │  │ go-git  │ │
│  │ semgrep  │ │     │  │vss      │ │     │  │ +CRDT   │ │
│  │ gosec    │ │     │  └─────────┘ │     │  └─────────┘ │
│  └───────────┘ │     └─────────────┘     └───────────────┘
└───────────────┘
        │
        ▼
┌───────────────┐     ┌───────────────┐
│    Runes      │     │    Nornir     │
│  (Skills)     │     │  (Testing)    │
│  ┌─────────┐ │     │  ┌─────────┐ │
│  │ CUE     │ │     │  │testcont.│ │
│  │ validation│     │  │k6       │ │
│  └─────────┘ │     │  └─────────┘ │
└───────────────┘     └───────────────┘
```

---

## 7. Flujo de Datos Local-First

### 7.1 Flujo Principal

```
[Usuario/Terminal] 
        │
        ▼
   ┌─────────┐
   │ Völva   │ ◄── TUI (interactive) / JSON (scripts) / quiet (CI)
   └────┬────┘
        │
        ▼
   ┌─────────────────────┐
   │    Odin (Router)    │ ◄── Tabla dinámica por fase SDD
   │  ┌───────────────┐  │
   │  │ollama-local   │  │──► Fallback: openrouter → anthropic
   │  │(default)      │  │
   │  └───────────────┘  │
   └────────┬────────────┘
            │
    ┌───────┴───────┐
    ▼               ▼
┌────────┐     ┌──────────┐
│Heimdall│     │  Mimir   │
│ (valida│     │ (indexa/ │
│ politi-│     │  consulta│     ◄──► Bifrost
│ cas)   │     │ memoria) │         (sync config)
└───┬────┘     └────┬─────┘
    │               │
    │         ┌─────┴─────┐
    │         ▼           ▼
    │    ┌────────┐  ┌──────────┐
    │    │ Runes  │  │ Plugins  │
    │    │(skills)│  │  (WASM)  │
    │    └────────┘  └──────────┘
    │               │
    └───────┬───────┘
            │
            ▼
    ┌──────────────┐
    │   Nornir     │
    │ (verifica    │
    │  E2E +       │
    │  benchmarks) │
    └──────┬───────┘
           │
           ▼
    ┌──────────────┐
    │   Dvergar    │
    │ (backup/     │
    │  upgrade     │
    │  atómico)    │
    └──────────────┘
```

### 7.2 Flujo de Sesión SDD

```
odyn new <change-name>
        │
        ▼
┌─────────────────────────────────────────┐
│          SDD Phase Orchestration         │
│                                          │
│  1. proposal ──► 2. spec ──► 3. design  │
│         │              │              │
│         ▼              ▼              │
│  4. tasks ──► 5. apply ──► 6. verify  │
│                                      │
│              7. archive               │
└─────────────────────────────────────────┘
        │
        ▼
   Session Snapshot
   (Git commit en ~/.odin/sessions/)
        │
        ▼
   Mimir store
   (decisions, artifacts refs)
```

---

## 8. Modelo de Datos

### 8.1 Entidades Principales

#### Session

```go
type Session struct {
    ID          string    `json:"id"`
    ChangeName  string    `json:"change_name"`
    Phase       SDDPhase  `json:"phase"`
    Status      string    `json:"status"` // active, paused, completed
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    SnapshotRef string    `json:"snapshot_ref"` // Git commit SHA
}
```

#### Memory

```go
type Memory struct {
    ID        string   `json:"id"`
    Content   string   `json:"content"`
    Embedding []float32 `json:"embedding,omitempty"`
    Project   string   `json:"project,omitempty"`
    Tags      []string `json:"tags"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Encrypted bool     `json:"encrypted"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

#### Skill (Rune)

```go
type Rune struct {
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    Description string            `json:"description"`
    Author      string            `json:"author,omitempty"`
    Tags        []string          `json:"tags"`
    Triggers    SkillTriggers     `json:"triggers"`
    Execution   SkillExecution    `json:"execution"`
    Outputs     SkillOutputs      `json:"outputs,omitempty"`
    Schema      cue.Value         `json:"-"` // CUE schema for validation
}
```

#### Policy (Heimdall)

```go
type Policy struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Rego        string   `json:"rego"` // OPA Rego code
    Severity    string   `json:"severity"` // critical, high, medium, low
    Enabled     bool     `json:"enabled"`
    Scope       string   `json:"scope"` // pre-commit, CI, manual
}
```

---

## 9. API & Interfaces

### 9.1 CLI Commands (Nivel Superior)

```bash
odin <command> [subcommand] [flags]

Commands:
  init          Inicializar entorno ODIN
  status        Mostrar salud del ecosistema
  session       Gestionar sesiones SDD
  migrate       Migrar desde Gentle AI
  plugin        Gestionar plugins WASM
  router        Configurar enrutamiento de modelos
  
  # Sub-sistemas
  mimir         Motor de memoria
  heimdall      Guardian de seguridad
  bifrost       Motor de sincronización
  runes         Registro de skills
  nornir        Suite de verificación
  dvergar       Instalación y deploy
  völva         Interface engine

Flags:
  --json        Output JSON
  --quiet       Modo silencioso
  --debug       Modo debug
```

### 9.2 Plugin Interface (WASM)

```go
// internal/plugins/plugin.go
type Plugin interface {
    // Metadata returns plugin info
    Metadata() PluginMetadata
    
    // Init initializes the plugin with config
    Init(ctx context.Context, config json.RawMessage) error
    
    // Execute runs the plugin's main logic
    Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
    
    // Health checks plugin status
    Health(ctx context.Context) error
}

type PluginMetadata struct {
    Name        string
    Version     string
    Author      string
    Description string
    Permissions []string // "fs:readonly", "net:disallow", etc.
}
```

### 9.3 Provider Interface (Router)

```go
// internal/router/provider.go
type Provider interface {
    Name() string
    Supports(model string) bool
    Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
    Embed(ctx context.Context, texts []string) ([]Embedding, error)
    CostPerToken(model string) float64
    IsAvailable(ctx context.Context) bool
}

// Built-in providers
type OllamaProvider struct{ ... }
type OpenRouterProvider struct{ ... }
type AnthropicProvider struct{ ... }
```

---

## 10. Seguridad, Privacidad y Licencias

### 10.1 Matriz de Licencias Permitidas

| Capa | ✅ Permitidas | ❌ Rechazadas |
|------|---------------|---------------|
| **Código Core** | MIT, Apache 2.0, BSD 2/3, ISC, MPL 2.0 | GPL v2/v3, SSPL, BSL, Propietarias |
| **Infraestructura** | AGPLv3 (self-hosted), LGPLv2.1 (semgrep) | SaaS-only, Vendor-locked |
| **Documentación** | CC-BY-SA 4.0, CC0 | NDAs, Restrictivas |
| **Binarios** | MIT, Apache 2.0 | Confidencial |

### 10.2 Políticas de Seguridad Local-First

| Política | Implementación |
|----------|---------------|
| **Zero telemetría** | Todo logueado localmente en `~/.odin/logs/` |
| **API keys seguras** | `~/.odin/providers.yaml` con `chmod 600` |
| **Plugins sandbox** | WASM sandbox (sin red/FS por defecto) |
| **Firmas криптографические** | `cosign verify` en every release |
| **Dependencias escaneadas** | `dependabot` + `trivy` en CI |
| **Cifrado en reposo** | `age` para base de datos Mimir |
| **Auditoría local** | Git-backed audit log en Bifrost |

### 10.3 Permisos de Plugins (WASM)

```yaml
# Plugin manifest (plugin.yaml)
name: "my-plugin"
version: "1.0.0"
permissions:
  fs: "readonly"        # Read-only filesystem access
  net: "disallow"       # No network access
  env: ["HOME", "PATH"]  # Allowed env vars only
sandbox: true           # Force sandbox mode
```

---

## 11. Roadmap por Sprint

### 11.1 Vista General

```
Sprint 1-2  │ S1: Odin Core (CLI base, config, TUI flags)
            │ S2: Mimir (sqlite-vss, search, encrypt)
────────────┼─────────────────────────────────────────────
Sprint 3-4  │ S3: Router (provider loading, fallback chain)
            │ S4: Heimdall (OPA, semgrep, pre-commit)
────────────┼─────────────────────────────────────────────
Sprint 5-6  │ S5: Bifrost (git repo, CRDT, push/pull)
            │ S6: Runes (CUE validation, install/test)
────────────┼─────────────────────────────────────────────
Sprint 7-8  │ S7: Dvergar (install.sh, cosign, rollback)
            │ S8: Nornir + Völva (testcontainers, themes)
────────────┼─────────────────────────────────────────────
Sprint 9+   │ v1.0 Release + Documentation + Community
```

### 11.2 Detalle por Sprint

#### Sprint 1: Odin Core

**Homólogo**: Odin AI  
**Entregables**:
- CLI base con cobra + viper
- Config parser para `~/.odin/config.yaml`
- Comandos `odin init` y `odin status`
- Flags `--quiet/--json/--interactive`
- TUI básica con bubbletea

**Criterios de Aceptación**:
```markdown
- `odin init` configura entorno en <10s (modo auto)
- `odin status` JSON válido en <800ms
- `--quiet/--json` funcional
- `odin --help` muestra todos los comandos
```

**Prompt OpenCode**:
```
"Genera cmd/odin/, internal/cli/, go.work, Makefile. 
Cobra + viper para CLI. Implementa init/status con TUI + JSON. 
Añade flags --quiet/--json/--interactive. 
Valida con make test-unit."
```

#### Sprint 2: Mimir

**Homólogo**: Mimir  
**Entregables**:
- Motor sqlite-vss con embeddings
- `mimir store/search/prune`
- Cifrado age
- Pruning automático

**Criterios de Aceptación**:
```markdown
- Búsqueda semántica <200ms en 10k registros
- `mimir encrypt/decrypt` funciona sin pérdida
- Pruning conserva tags: arch, spec, security
- Fallback a FTS5 si sqlite-vss no disponible
```

#### Sprint 3: Router

**Homólogo**: Odin (interno)  
**Entregables**:
- Provider loading dinámico
- Fallback chain: ollama → openrouter → anthropic
- Cost tracking local
- Métricas exportables

**Criterios de Aceptación**:
```markdown
- Fallback automático en ≤3 intentos
- Métricas registradas (latencia, costo)
- Compatibilidad openai-go
- `odin router set` y `odin router fallback add`
```

#### Sprint 4: Heimdall

**Homólogo**: Heimdall  
**Entregables**:
- OPA integration con Rego
- semgrep + gosec para SAST
- Pre-commit hook
- Reporte CI (exit 0/1/2)

**Criterios de Aceptación**:
```markdown
- `heimdall check` en <2s
- Reglas custom en `~/.odin/rules/`
- Bloqueo en vulnerabilidades críticas
- Reporte JSON para CI/CD
```

#### Sprint 5: Bifrost

**Homólogo**: Bifrost  
**Entregables**:
- Git repo local en `~/.odin/config/`
- CRDT merge para conflictos
- `push/pull/status/diff`

**Criterios de Aceptación**:
```markdown
- Conflictos resueltos automáticamente o marcados
- `odin sync diff` antes de aplicar
- Funciona en WSL2, macOS, Linux
- Historial de cambios
```

#### Sprint 6: Runes

**Homólogo**: Runes  
**Entregables**:
- Schema validation CUE
- `install/test/search` commands
- Sandbox execution
- Cache local

**Criterios de Aceptación**:
```markdown
- Validación CUE antes de cargar
- Skill malformado = warning, no block
- `odin rune test` con diff output
- Cache evita red para skills frecuentes
```

#### Sprint 7: Dvergar

**Homólogo**: Dvergar  
**Entregables**:
- `install.sh` con detección SO
- `cosign verify` antes de ejecutar
- Backup + rollback automático

**Criterios de Aceptación**:
```markdown
- Instalación en <2 min
- `cosign verify` pasa
- Rollback automático si falla
- Log en `~/.odin/logs/`
```

#### Sprint 8: Nornir + Völva

**Homólogos**: Nornir, Völva  
**Entregables**:
- testcontainers-go para E2E
- Matrix testing (Ubuntu, Arch, macOS, WSL2)
- Flaky detection
- Theme engine para Völva

**Criterios de Aceptación**:
```markdown
- E2E en <15 min
- Flaky tests marcados
- Benchmarks <2s para fases críticas
- `völva theme set` sin reiniciar
```

---

## 12. Criterios de Aceptación

### 12.1 Checklist de Entrega Final (v1.0)

```markdown
## Core
- [ ] `odin init` configura entorno en <10s (modo auto)
- [ ] `odin status` muestra salud de Mimir, Heimdall, Bifrost, Runes
- [ ] `odin migrate --from gentle` funciona sin breaking changes

## Mimir
- [ ] Memoria semántica <200ms + cifrado opcional age
- [ ] `mimir encrypt/decrypt` sin pérdida
- [ ] Pruning automático con tags configurables

## Router
- [ ] Router multi-model con fallback chain
- [ ] Cost tracking local ($0 para modo local)
- [ ] Métricas exportables en JSON

## Heimdall
- [ ] OPA + semgrep + gosec integrados
- [ ] Pre-commit hook bloquea vulnerabilidades críticas
- [ ] Reporte CI con exit codes correctos

## Bifrost
- [ ] CRDT sin conflictos + diff pre-apply
- [ ] Multi-device sync funcional
- [ ] GPG signing opcional

## Runes
- [ ] CUE validation + sandbox test
- [ ] `odin rune search/install/test/remove`
- [ ] Cache local + rollback de versiones

## Dvergar
- [ ] `curl | sh` con cosign verify
- [ ] Upgrade + rollback automático
- [ ] Detección SO/arch automatica

## Nornir
- [ ] Matrix <15 min + flaky detection
- [ ] Benchmarks de latencia
- [ ] Cleanup automático

## Völva
- [ ] Temas intercambiables
- [ ] --json/--quiet/--interactive
- [ ] Accesibilidad básica

## General
- [ ] 100% OSS, $0 costo infraestructura
- [ ] Local-first garantizado (funciona sin red)
- [ ] Docs completas + ejemplos
```

---

## 13. Glosario

| Término | Definición |
|---------|------------|
| **CRDT** | Conflict-free Replicated Data Type. Estructura de datos que permite sincronización sin conflictos. |
| **CUE** | Lenguaje de schema/validation. Valida configuration con tipos y constraints. |
| **FTS5** | Full-Text Search 5. Motor de búsqueda por keywords en SQLite. |
| **Local-First** | Arquitectura donde la aplicación funciona principalmente offline, usando la nube solo como sync opcional. |
| **OPA** | Open Policy Agent. Motor de políticas Rego para control de acceso. |
| **SAST** | Static Application Security Testing. Análisis de código sin ejecución. |
| **SDD** | Spec-Driven Development. Metodología de desarrollo basada en especificaciones formales. |
| **sqlite-vss** | Extensión de SQLite para búsqueda vectorial con embeddings. |
| **WASM** | WebAssembly. Runtime sandbox para plugins seguros. |
| **Fallback Chain** | Lista ordenada de providers a intentar si el primario falla. |

---

## 14. Apéndice: Migración desde Gentle AI

### 14.1 Comando de Migración

```bash
odyn migrate --from gentle [flags]

Flags:
  --dry-run              Simular migración sin aplicar cambios
  --backup               Crear backup antes de migrar
  --configs string       Path a configs de Gentle (default "~/.gentle-ai")
  --overwrite            Sobrescribir si existe
```

### 14.2 Mapa de Correspondencia

| Gentle AI | ODIN AI | Notas |
|-----------|---------|-------|
| `gentle-ai` | `odin` | CLI principal |
| `~/.gentle-ai/` | `~/.odin/` | Directorio de config |
| `gentle-ai agents` | `odin status` | Estado del sistema |
| `gentle-ai update` | `odin upgrade` | Actualización |
| `engram` | `mimir` | Motor de memoria |
| `AGENTS.md` | `~/.odin/rules/` | Reglas de Heimdall |
| `skills/` | `runes/` | Skills registry |
| `~/.config/gentle/` | `~/.odin/config/` | Sync config |

### 14.3 Lo Que Se Migra

```markdown
✅ Configuración de agentes (Claude Code, OpenCode, etc.)
✅ Preferences y settings
✅ Reglas AGENTS.md → policies Rego
✅ Skills instalados
✅ MCP server configs
✅ Engram memories → Mimir (conversión de schema)

⚠️ Skills custom (requiere validación CUE)
⚠️ Configuraciones de provider (requiere re-ingreso de API keys)

❌ Persona (no aplica en ODIN, es local-first)
❌ Telemetry settings (no existe en ODIN)
```

---

## Footer

> *"Gentle AI fue la semilla. El ecosistema Odyn es el bosque. Local. Libre. Para la banda. Sin excusas."*
>
> **Autor**: Andragon  
> **Inspirado por**: [Gentleman-Programming/gentle-ai](https://github.com/Gentleman-Programming/gentle-ai)  
> **Fecha**: Abril 2026  
> **Estado**: Listo para desarrollo en OpenCode
