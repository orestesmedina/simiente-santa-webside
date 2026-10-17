# Estado: 009-medios-predicas-y-podcasts

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 009-medios-predicas-y-podcasts |
| Flujo | equipo-feature |
| Fase | 1/9 · Especificar (spec **v1 borrador** con 24 preguntas abiertas) |
| Ciclo de corrección | 0/3 |
| Próximo paso | El cliente responde `docs/producto/cuestionario-F9-medios.md`; con las respuestas, fase 2 (`clarify`) y **puerta 1** (aprobación de la spec) |
| Bloqueado por | **Respuestas del cliente** al cuestionario (24 marcas `[NECESITA ACLARACIÓN]`; la spec no se puede aprobar con marcas abiertas) |
| Actualizado | 2026-10-17 |

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

- (ninguno: todavía no hay validación)

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-17 (humano): se construye **F9 antes que F5 y F8** (F9 depende solo de F2 y F3, ambas terminadas). Registrado como **decisión 26 del roadmap**.
- 2026-10-17 (humano, alcance asumido): F9 se especifica con la **Opción B** (pegar la dirección + relleno automático de título/autor/imagen + reproducción **dentro** del sitio), **pendiente de confirmación del cliente** en el cuestionario. La **sincronización automática (Opción C) queda fuera de F9** (fase 3, por feeds RSS/Atom; no por las interfaces oficiales: Spotify exige pago desde el 9-mar-2026).
- 2026-10-17 (humano): el cuestionario se le entrega al cliente **antes** de aprobar la spec, para no inventar reglas de negocio.
- 2026-10-17 (humano): autoriza **cerrar el costo del bug del date picker desde su rama muerta** y traer ese único cambio a `main` con un **cherry-pick** (el PR ya estaba aprobado y fusionado).

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-17 — **Cierre de sesión.** Costo del bug del date picker **cerrado** en **68,8 M ≈ $1,41** (desde su rama muerta) y traído a `main` con **cherry-pick** (`ee02ab1`); la rama de F9 quedó **rebasada sobre `main`** y el informe del proyecto ya muestra ese bug como **cerrado**. **`main` quedó 1 commit por delante de `origin`**: el `git push` **falló por falta de credenciales** en el entorno del agente (no hay `gh`, ni llave SSH, ni `credential.helper`, ni token) — **lo empuja el humano**. El mismo problema bloqueará **abrir el PR** en la fase 9 si no se resuelve antes.
- 2026-10-17 — **Inicio de F9**: rama `009-medios-predicas-y-podcasts` creada desde `main`; **spec v1** redactada por `analista-producto` (7 historias, FR-001…FR-028, SC-001…SC-016, criterios manuales SC-008 y SC-012); **cuestionario para el cliente** escrito en `docs/producto/cuestionario-F9-medios.md`; roadmap → F9 **en curso** y decisión 26 registrada. **Puerta 1 pendiente** de las respuestas del cliente (24 preguntas abiertas).
- 2026-10-17 — Contexto previo del mismo día: cerrados y fusionados el **PR #11** (selector de fecha/hora del panel) y el **PR #12** (flaky del CI); el **defecto conocido 23** queda **resuelto** (decisión 25 del roadmap), así que F9 ya no arrastra ese riesgo.
