---
name: equipo-feature
description: Flujo completo para construir una funcionalidad nueva con el equipo de subagentes, desde la spec hasta el Pull Request, con aprobaciones humanas. Usar cuando el usuario pide construir, agregar o desarrollar una funcionalidad.
---
# Flujo: construir una funcionalidad

Lee `AGENTS.md` y `.specify/memory/constitution.md` antes de empezar.
La funcionalidad es la que describió el usuario en su mensaje.

Los comandos de Spec Kit se invocan como `/speckit.<fase>` (Claude Code, OpenCode) o `$speckit-<fase>` (Codex).
Si tu herramienta no puede lanzar subagentes, asume tú cada rol leyendo `equipo/agentes/<rol>.md`.

## Fase 1 — Especificación
1. Ejecuta la fase `specify` de Spec Kit delegando la redacción en `analista-producto`.
2. Si la spec tiene marcas `[NECESITA ACLARACIÓN]`, ejecuta `clarify` con el mismo subagente y haz las preguntas al humano.
3. **DETENTE.** Muestra un resumen de la spec y pide aprobación explícita ("apruebo la spec").

## Fase 2 — Plan técnico
1. Ejecuta `plan` delegando en `arquitecto`.
2. Si hay interfaz de usuario, pide en paralelo a `disenador-ux` que escriba `ux.md`.
3. **DETENTE.** Resume decisiones clave, modelo de datos, endpoints y riesgos, y pide aprobación explícita ("apruebo el plan").

## Fase 3 — Tareas y coherencia
1. Ejecuta `tasks` delegando en `arquitecto`.
2. Ejecuta `analyze` delegando en `revisor-codigo`. Corrige inconsistencias críticas con el arquitecto antes de seguir.

## Fase 4 — Implementación
1. Ejecuta `implement`, delegando cada tarea según su capa: `[backend]`/`[db]` → `dev-backend`, `[frontend]` → `dev-frontend`, `[infra]` → `devops`.
2. Tareas `[P]` sin dependencias entre sí pueden ir en paralelo.
3. Un commit por tarea terminada (Conventional Commits).

## Fase 5 — Validación
Aplica la skill `equipo-revision`. Si rechaza, envía los bloqueantes al desarrollador correspondiente y repite. Máximo 3 ciclos; luego detente y escala al humano.

## Fase 6 — Convergencia
Ejecuta `converge`. Si agrega tareas, vuelve a la Fase 4 solo para ellas.

## Fase 7 — Entrega
1. `devops`: verifica que CI pase y que `.env.example` y Docker estén al día.
2. `documentador`: actualiza CHANGELOG, README y notas para el cliente.
3. Abre un Pull Request con el resumen, las pruebas y los reportes de validación.
4. **DETENTE.** El merge y el despliegue a producción los aprueba un humano.

Termina con un resumen corto: qué se construyó, estado de las pruebas, riesgos abiertos y enlace al PR.
