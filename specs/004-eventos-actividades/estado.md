# Estado: 004-eventos-actividades

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 004-eventos-actividades |
| Flujo | equipo-feature |
| Fase | 1/9 · Especificar (spec redactada y aclarada; en la puerta 1) |
| Ciclo de corrección | 0/3 |
| Próximo paso | **Puerta 1**: el humano aprueba la spec → pasa a Fase 2 (plan técnico) |
| Bloqueado por | — |
| Actualizado | 2026-10-11 |

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

- 2026-10-11 — **Arranque de F4 (decisión del humano en sesión).** El humano indicó «continuemos con F4», el siguiente pendiente del roadmap (F1, F2, F3 y F10 ya terminadas). Se arranca el flujo `equipo-feature` en la rama `004-eventos-actividades`.
- 2026-10-11 — **Aclaraciones de F4 resueltas por el humano** (registradas en `spec.md`, sección «Aclaraciones resueltas (2026-10-11)»; también decisión 16 del roadmap): las 6 preguntas se resolvieron por la opción recomendada — **Q1** fechas concretas + texto de frecuencia; **Q2** hora de inicio/fin opcionales; **Q3** solo lo próximo; **Q4** retiro automático de la vista pública al pasar la fecha; **Q5** una entrada «Eventos y actividades» con dos bloques; **Q6** categoría opcional con catálogo administrable.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-11 — **Fase 1 (Especificar)**: `spec.md` redactada por `analista-producto` y **aclarada** (Q1–Q6 resueltas por el humano, todas por la opción recomendada). 6 user stories, FR-001–FR-024, SC-001–SC-013; sin marcas `[NECESITA ACLARACIÓN]`. En la **puerta 1**, pendiente la aprobación humana de la spec. Rama `004-eventos-actividades` creada; roadmap: F4 → **en curso**.
- 2026-10-11 — Inicio de la funcionalidad F4 (eventos y actividades): rama `004-eventos-actividades` creada, `spec.md` desde plantilla. Fase 1 (especificar).
