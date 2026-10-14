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
| Fase | 1/9 · Especificar (spec redactada; pendiente de `clarify` y de la puerta 1) |
| Ciclo de corrección | 0/3 |
| Próximo paso | Resolver con el humano las 11 preguntas abiertas (Q1–Q11) y actualizar la spec; luego pedir aprobación de la spec (puerta 1) |
| Bloqueado por | — |
| Actualizado | 2026-10-14 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | pendiente | | | |
| Plan | pendiente | | | |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno)

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-14 (humano): se construye **F7 (donaciones) antes que F5 (grupos de conexión)**. F5 queda pendiente; ambas son independientes (roadmap §3).

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-14 — Fase 1: rama `007-donaciones` creada y `spec.md` redactada por `analista-producto` (25 FR, 13 SC, 5 historias). 11 preguntas abiertas (Q1–Q11) para `clarify`.
- 2026-10-14 — Inicio de la funcionalidad.
