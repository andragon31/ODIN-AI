# Protocolo PENTAKILL (Odinista v2)

El protocolo **PENTAKILL** es la evolución final de la ingeniería agentica en ODIN. Combina cinco metodologías fundamentales para garantizar que el software sea conceptualmente sólido, técnicamente robusto y funcionalmente impecable.

## Los 5 Pilares del PENTAKILL

1. **DDD (Domain-Driven Design)**: Define el "qué" y el "quién". Establece el Lenguaje Ubicuo y los modelos tácticos (Entidades, Agregados).
2. **SDD (Spec-Driven Development)**: Define el flujo y la estructura. Es el orquestador de las fases.
3. **CFD (Contract-First Development)**: Define el "cómo" se comunican las piezas. Establece interfaces técnicas (gRPC, OpenAPI, Go Interfaces) antes de la lógica.
4. **TDD (Test-Driven Development)**: Define la verdad de la lógica. Ciclos Red-Green-Refactor para asegurar robustez interna.
5. **BDD (Behavior-Driven Development)**: Define la verdad del usuario. Escenarios de Gherkin que validan el comportamiento final.
6. **DVERGAR (Consultative Deployment)**: Define la verdad del entorno. Forja y despliegue consultivos de la infraestructura.

## Flujo del Pentáculo de Poder

| Orden | Fase | Acción Requerida | Artefacto |
|-------|------|------------------|------------|
| 1 | **Domain** | Extracción de términos y modelado | `openspec/domain/domain.md` |
| 2 | **Contract**| Definición de esquema técnico | `openspec/contracts/*.proto|yaml|go` |
| 3 | **Spec** | Requisitos BDD en Lenguaje Ubicuo | `openspec/features/*.feature` |
| 4 | **Design** | Plan de arquitectura y testing | `design.md` |
| 5 | **Tasks** | Desglose en ciclos TDD | `tasks.md` |
| 6 | **Apply** | Implementación estricta TDD | Código + Unit Tests |
| 7 | **Verify** | Cumplimiento Contractual y BDD | `verify-report.md` |
| 8 | **Deploy** | Forja consultiva de infraestructura | `infra_design.md` + IAcC |

---

## Reglas de Oro de PENTAKILL (El Código del Honor)

### I. Indagación Proactiva (La Regla del Arquitecto)
**ODIN NO SUPONE.** Si el alcance, el dominio o el contrato son ambiguos, el agente DEBE detenerse e iniciar una **Conversación Dinámica** con el usuario. Es preferible preguntar diez veces que reconstruir una vez.

### I. Primacía del Dominio
Si un término no existe en el `domain.md`, no puede usarse en el código ni en los escenarios BDD. El agente DEBE forzar el uso del Lenguaje Ubicuo.

### II. El Contrato es la Ley
Ninguna fase de implementación (`apply`) puede comenzar si el archivo de contrato técnica (`openspec/contracts/`) no ha sido validado. El contrato actúa como el "Single Source of Truth" para la comunicación entre componentes.

### III. Prohibición de Código Huérfano
Todo código escrito debe:
1. Pertenecer a un Agregado del Dominio.
2. Cumplir con un Contrato definido.
3. Tener un test fallido previo (TDD).
4. Satisfacer un escenario de usuario (BDD).

---

> [!IMPORTANT]
> **PENTAKILL no es negociable.** En el momento en que se activa este modo, ODIN se comporta como un arquitecto senior extremadamente riguroso. No hay atajos para la calidad.
