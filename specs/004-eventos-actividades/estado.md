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
| Fase | 4/9 · Implementar (42 tareas; en curso) |
| Ciclo de corrección | 0/3 |
| Próximo paso | T401 `[db]` Migración `000007_create_agenda_content` → `dev-backend` |
| Bloqueado por | — |
| Actualizado | 2026-10-11 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano | 2026-10-11 | v1: "apruebo la spec"; v2 (permiso «categorías»): "Re-apruebo la spec" |
| Plan | aprobado | humano | 2026-10-11 | "Apruebo el plan" (con las decisiones firmes: APP_TZ=America/Costa_Rica, ruta /eventos, aviso sin redirigir, permiso «categorías», endTime exige startTime, borrado de categoría en uso 409) |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno)

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-11 — **Arranque de F4 (decisión del humano en sesión).** El humano indicó «continuemos con F4», el siguiente pendiente del roadmap (F1, F2, F3 y F10 ya terminadas). Se arranca el flujo `equipo-feature` en la rama `004-eventos-actividades`.
- 2026-10-11 — **Aclaraciones de F4 resueltas por el humano** (registradas en `spec.md`, sección «Aclaraciones resueltas (2026-10-11)»; también decisión 16 del roadmap): las 6 preguntas se resolvieron por la opción recomendada — **Q1** fechas concretas + texto de frecuencia; **Q2** hora de inicio/fin opcionales; **Q3** solo lo próximo; **Q4** retiro automático de la vista pública al pasar la fecha; **Q5** una entrada «Eventos y actividades» con dos bloques; **Q6** categoría opcional con catálogo administrable.
- 2026-10-11 — **Decisiones de la puerta 2 (plan) y cambio de alcance** (registradas en `plan.md` §"Puntos resueltos" y decisión 17 del roadmap): **`APP_TZ = America/Costa_Rica`**; **ruta pública `/eventos`**; **página con aviso sin redirigir** si no hay nada publicado; **permiso nuevo «categorías»** para el catálogo (amplía el catálogo de F2 de 9→10; el humano asumió el impacto: migración que siembra el permiso y actualización de las pruebas de F2 que fijan 9); `endTime` exige `startTime`; borrar categoría en uso → 409; imagen de elemento vencido no se sirve.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-11 — **Puertas 1 (re-aprobación) y 2 superadas**: el humano **re-aprobó la spec** ("Re-apruebo la spec") y **aprobó el plan** ("Apruebo el plan"). Arranca la Fase 3 (tareas + coherencia).
- 2026-10-11 — **Fase 2 (Planificar)**: `plan.md`, `research.md`, `data-model.md`, `contracts/openapi.yaml` y `quickstart.md` redactados por `arquitecto`; `ux.md` por `disenador-ux`. 4 tablas `agenda_*`, 3 migraciones (`000007` contenido, `000008` auditoría, `000009` permiso «categorías»), 0 dependencias nuevas, reutiliza el sistema de F10 y la subida de imágenes de F3. **Cambio de alcance aprobado** en la puerta 2 (permiso «categorías»); spec actualizada → **re-aprobación pendiente**. **Puerta 2 pendiente** de aprobación humana del plan.
- 2026-10-11 — **Puerta 1 superada**: el humano **aprobó la spec** ("apruebo la spec"). Arranca la Fase 2 (plan técnico).
- 2026-10-11 — **Fase 1 (Especificar)**: `spec.md` redactada por `analista-producto` y **aclarada** (Q1–Q6 resueltas por el humano, todas por la opción recomendada). 6 user stories, FR-001–FR-024, SC-001–SC-013; sin marcas `[NECESITA ACLARACIÓN]`. En la **puerta 1**, pendiente la aprobación humana de la spec. Rama `004-eventos-actividades` creada; roadmap: F4 → **en curso**.
- 2026-10-11 — Inicio de la funcionalidad F4 (eventos y actividades): rama `004-eventos-actividades` creada, `spec.md` desde plantilla. Fase 1 (especificar).
