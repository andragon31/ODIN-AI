# Protocolo de la Tríada Odinista

La Tríada Odinista es la integración de Spec-Driven Development (SDD), Test-Driven Development (TDD) y Behavior-Driven Development (BDD) en un flujo único de ingeniería de alta precisión.

## Los Tres Pilares

1. **SDD (Estructura)**: Define las fases y el flujo de los artefactos. Es el esqueleto.
2. **BDD (Comportamiento)**: Define el contrato de usuario mediante Gherkin. Es la piel.
3. **TDD (Implementación)**: Define la lógica interna mediante ciclos Red-Green-Refactor. Son los músculos.

## Flujo en Modo 'Triad'

| Fase SDD | Acción Triad Obligatoria | Artefacto Generado |
|----------|--------------------------|-------------------|
| **Spec** | Definir Escenarios BDD | `openspec/features/*.feature` |
| **Design**| Plan de Mocks y Stubs | `design.md` (Testing section) |
| **Tasks** | Ciclos Red-Green | `tasks.md` (con tareas `[TDD]`) |
| **Apply** | Implementación TDD | Código + Unit Tests |
| **Verify**| Validación BDD | `verify-report.md` (Compliance Matrix) |

---

## El Ciclo de Implementación (Apply)

En modo Triad, no se escribe código sin un test fallido previo. El proceso es:

1. **Seleccionar un Escenario BDD**: Leer el archivo `.feature`.
2. **Setup del Test (RED)**: Escribir un test unitario o de integración que falle porque la funcionalidad no existe.
3. **Mínima Implementación (GREEN)**: Escribir el código justo para que el test pase.
4. **Refactorización (REFACTOR)**: Limpiar el código manteniendo el test en verde.
5. **Verificación BDD**: Ejecutar el escenario de Gherkin para confirmar que el comportamiento es correcto.

## Definición de "Cumplimiento BDD"

Un requisito se considera cumplido solo si:
- Existe un escenario Gherkin que lo describe.
- Existe al menos un test automatizado (unitario o de integración) vinculado a ese escenario.
- El test ha pasado en el entorno de verificación.

---

> [!TIP]
> La Tríada no busca escribir más código, sino escribir solo el código necesario y que esté validado por contrato. Si un cambio no tiene un escenario Gherkin, no pertenece a la especificación.
