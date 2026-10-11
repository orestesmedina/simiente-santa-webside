# Estado: 010-sistema-diseno-movimiento

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 010-sistema-diseno-movimiento |
| Flujo | equipo-feature |
| Fase | 2/9 · Planificar (entregables listos; en la puerta 2) |
| Ciclo de corrección | 0/3 |
| Próximo paso | **Puerta 2**: aprobar el plan y elegir una de las 3 propuestas (A/B/C); luego Fase 3 (tareas + coherencia) |
| Bloqueado por | — |
| Actualizado | 2026-10-10 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano | 2026-10-10 | "Apruebo la spec" (con ajuste de alcance: F10 = sitio público + modo oscuro con interruptor manual; panel → F10B) |
| Plan | pendiente | | | |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno)

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-10 — **Arranque de F10 (decisión del humano en sesión).** Ante la pregunta «¿qué funcionalidad arrancamos?», el humano eligió **F10 — Sistema de diseño y movimiento** antes que F4, coherente con la Decisión 11 del roadmap (construirlo antes de F4 para que las secciones futuras nazcan ya con el sistema). Se arranca el flujo `equipo-feature` en la rama `010-sistema-diseno-movimiento`.
- 2026-10-10 — **Aclaraciones de F10 resueltas por el cliente** (registradas en `spec.md`, sección «Aclaraciones resueltas (2026-10-10)»; también decisión 12 del roadmap): **Q1** F10 incluye también el **panel de administración** (no solo el sitio público); **Q2** el «scroll incómodo» eran **ambos** casos (enlaces que se desplazan de lado en móvil y navegación que se pierde al bajar); **Q3** navegación en **barra fija siempre visible**; **Q4** animación **moderada** (respetando `prefers-reduced-motion`); **Q5** el selector **solo compacta los textos actuales** (sin píldoras ni desplegables, comportamiento de F3 intacto); **Q6** **se permite crear iconografía/ilustraciones nuevas** dentro de la marca; **Q7** **se incluye el modo oscuro** (sitio y panel); **Q8** referencias de inspiración (no copiar) `iglesiaoasis.com` y `river-arena.org`, y el cliente quiere **3 propuestas de diseño distintas e innovadoras** antes de implementar.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-10 — **Fase 2 completada**: `plan.md`, `research.md`, `data-model.md` (sin cambios de datos), `contracts/README.md` (sin cambios de API) y `quickstart.md` (`arquitecto`); `ux.md` con **3 propuestas** —A «Raíz Viva», B «Cielo Abierto», C «Luz Serena»— más maquetas HTML navegables en `propuestas/` (`disenador-ux`). Sin dependencias nuevas (movimiento con CSS puro); tema con `data-theme` + variables semánticas. Pendiente: **puerta 2** (aprobar el plan y elegir una propuesta).
- 2026-10-10 — **Puerta 1 superada:** el humano aprobó la spec (**«Apruebo la spec»**) con dos ajustes de alcance: **F10 = sitio público + modo oscuro** (con **interruptor manual claro/oscuro**) y el **panel se separa a F10B**. Aprobación registrada en «Aprobaciones». Roadmap: F10 ajustada y **F10B** añadida (decisión 13). Arranca la Fase 2 (plan + 3 propuestas de diseño).
- 2026-10-10 — **8 aclaraciones resueltas con el cliente** (Q1–Q8): F10 se amplía a **panel de administración + modo oscuro**, navegación en **barra fija siempre visible**, animación **moderada**, selector de idioma **solo compactado**, **iconografía/ilustraciones nuevas permitidas**, referencias de inspiración y **3 propuestas de diseño a presentar antes de implementar**. Spec actualizada por `analista-producto`. Pendiente: aprobación de la spec (puerta 1).
- 2026-10-10 — Inicio de F10: rama `010-sistema-diseno-movimiento` y spec redactada por `analista-producto` (5 historias US1–US5, 25 escenarios, 22 FR, 11 SC, 13 Edge Cases). Quedan **8 preguntas abiertas (Q1–Q8)** que cambian alcance o experiencia; se aclaran con el humano antes de aprobar la spec.
