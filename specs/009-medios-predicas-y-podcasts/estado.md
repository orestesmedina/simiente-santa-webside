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
| Fase | 1/9 · Especificar (**spec v2** lista y sin preguntas abiertas; **puerta 1 pendiente**) |
| Ciclo de corrección | 0/3 |
| Próximo paso | **Puerta 1: aprobación humana de la spec v2** (checklist de `docs/GUIA-INICIO.md` §5) |
| Bloqueado por | — (nada: la spec v2 no tiene marcas `[NECESITA ACLARACIÓN]`) |
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
- 2026-10-17 (humano, alcance asumido): F9 se especifica con la **Opción B** (pegar la dirección + relleno automático de título/autor/imagen + reproducción **dentro** del sitio), **pendiente de confirmación del cliente** en el cuestionario. La **sincronización automática (Opción C) queda fuera de F9** (fase 3, por feeds RSS/Atom; no por las interfaces oficiales: Spotify exige pago desde el 9-mar-2026). *(Superado por el giro del 2026-10-17, más abajo: la importación desde los feeds **entra** en F9, a demanda.)*
- 2026-10-17 (humano): el cuestionario se le entrega al cliente **antes** de aprobar la spec, para no inventar reglas de negocio.
- 2026-10-17 (humano): autoriza **cerrar el costo del bug del date picker desde su rama muerta** y traer ese único cambio a `main` con un **cherry-pick** (el PR ya estaba aprobado y fusionado).
- 2026-10-17 (cliente, por medio del humano): **respuestas al cuestionario** (15 de 17 coinciden con lo recomendado) y **giro de alcance**: F9 se construye con **importación desde los feeds** (YouTube + podcast) y **enlaces que salen del sitio**, con un botón **«Buscar episodios nuevos»** en el panel. Registrado como **decisión 27 del roadmap**; corrige el punto de la decisión 26 que dejaba la sincronización fuera de F9.
- 2026-10-17 (humano): el **mismo sermón en YouTube y Spotify** se muestra como **una sola entrada con video y audio** (Q18).

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-17 — **Spec v2** redactada tras las respuestas del cliente: **6 historias de usuario**, **FR-001…FR-030**, **SC-001…SC-017** (criterios manuales SC-008 lector de pantalla y SC-012 usabilidad), **10 consecuencias asumidas**, **16 defaults revisables** y **0 preguntas abiertas**. Limpiadas las referencias técnicas a migraciones que arrastraba la v1. Roadmap: **decisión 27** y corrección de la 26. **Puerta 1 pendiente.**
- 2026-10-17 — **Cierre de sesión.** Costo del bug del date picker **cerrado** en **68,8 M ≈ $1,41** (desde su rama muerta) y traído a `main` con **cherry-pick** (`ee02ab1`); la rama de F9 quedó **rebasada sobre `main`** y el informe del proyecto ya muestra ese bug como **cerrado**. **`main` quedó 1 commit por delante de `origin`**: el `git push` **falló por falta de credenciales** en el entorno del agente (no hay `gh`, ni llave SSH, ni `credential.helper`, ni token) — **lo empuja el humano**. El mismo problema bloqueará **abrir el PR** en la fase 9 si no se resuelve antes.
- 2026-10-17 — **Inicio de F9**: rama `009-medios-predicas-y-podcasts` creada desde `main`; **spec v1** redactada por `analista-producto` (7 historias, FR-001…FR-028, SC-001…SC-016, criterios manuales SC-008 y SC-012); **cuestionario para el cliente** escrito en `docs/producto/cuestionario-F9-medios.md`; roadmap → F9 **en curso** y decisión 26 registrada. **Puerta 1 pendiente** de las respuestas del cliente (24 preguntas abiertas).
- 2026-10-17 — Contexto previo del mismo día: cerrados y fusionados el **PR #11** (selector de fecha/hora del panel) y el **PR #12** (flaky del CI); el **defecto conocido 23** queda **resuelto** (decisión 25 del roadmap), así que F9 ya no arrastra ese riesgo.
