# Estado: 006-ministerios

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 006-ministerios |
| Flujo | equipo-feature |
| Fase | 1/9 · Especificar |
| Ciclo de corrección | 0/3 |
| Próximo paso | Redactar la spec con `analista-producto` (F6 — Ministerios) |
| Bloqueado por | — |
| Actualizado | 2026-10-13 |

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

- 2026-10-13 — **Arranque de F6 (decisión del humano en sesión).** El humano pidió «continuemos con el f6» en lugar de F5 (el siguiente pendiente del roadmap). F4, F5, F6 y F7 son independientes entre sí según el roadmap §3, así que se arranca F6 — Ministerios por el flujo `equipo-feature`. Numeración de la rama: **`006-ministerios`**, alineada con el identificador de roadmap (precedente: F10 → `010-sistema-diseno-movimiento`), dejando el número 005 libre para F5 (grupos de conexión).

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-13 — Inicio de la funcionalidad F6 (ministerios): rama `006-ministerios` creada desde `main` (`0797d3d`), `spec.md` desde la plantilla de Spec Kit. Fase 1 (especificar). Roadmap: F6 → **en curso**.
