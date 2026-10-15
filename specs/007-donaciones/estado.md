# Estado: 007-donaciones

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 007-donaciones |
| Flujo | equipo-feature |
| Fase | 2/9 · Planificar (plan, research, data-model, contratos y ux listos; pendiente la puerta 2) |
| Ciclo de corrección | 0/3 |
| Próximo paso | Aprobación humana del plan (puerta 2). Después: fase 3 (`tasks` con `arquitecto` + `analyze` con `revisor-codigo`) |
| Bloqueado por | — |
| Actualizado | 2026-10-14 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano (cliente) | 2026-10-14 | "si" (aprueba la spec v2 de F7) |
| Plan | pendiente | | | |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno)

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-14 (humano): se construye **F7 (donaciones) antes que F5 (grupos de conexión)**. F5 queda pendiente; ambas son independientes (roadmap §3).
- 2026-10-14 (humano): aclaraciones de F7 (Q1–Q11) resueltas; registradas como **decisión 22 del roadmap**. Cambios de alcance: sección «Ofrendas y diezmos» en `/ofrendas` (Q8) y **bloque en la portada debajo del de ministerios** (Q7).

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-14 — Fase 2: `arquitecto` produjo plan/research/data-model/contracts/quickstart y `disenador-ux` produjo `ux.md`. Corregida una contradicción interna de la spec (entidad financiera no traducible). Pendiente la puerta 2 (aprobación del plan).
- 2026-10-14 — Puerta 1: **spec v2 aprobada** por el humano. Fase 2 (planificar) iniciada.
- 2026-10-14 — Fase 1 (clarify): las 11 preguntas resueltas y aplicadas; `spec.md` pasa a v2 sin marcas abiertas. Pendiente la puerta 1 (aprobación de la spec).
- 2026-10-14 — Fase 1: rama `007-donaciones` creada y `spec.md` redactada por `analista-producto` (25 FR, 13 SC, 5 historias). 11 preguntas abiertas (Q1–Q11) para `clarify`.
- 2026-10-14 — Inicio de la funcionalidad.
