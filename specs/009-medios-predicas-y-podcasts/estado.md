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
| Fase | 6/9 · Implementar (bloque `[db]` hecho: **T901–T903**; **3/46** tareas) |
| Ciclo de corrección | 0/3 |
| Próximo paso | Fase 6, oleada P1: **T904** (contrato), **T906** (auditoría), **T907/T907A** (cliente HTTP) y **T908–T910** (feeds, disponibilidad, carátulas); después T911+ |
| Bloqueado por | — |
| Actualizado | 2026-10-17 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano (cliente) | 2026-10-17 | "Apruebo la spec v2" |
| Plan | aprobado | humano (cliente) | 2026-10-17 | "Aplica las 5 y apruebo el plan" |
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
- 2026-10-17 (humano, **puerta 2**): aprueba el plan con **5 precisiones** («Aplica las 5 y apruebo el plan»): (1) solo un «posible duplicado» sin resolver mantiene la ficha fuera del público (una marcada «ya no aparece en el feed» sigue visible); (2) un enlace confirmado como no disponible oculta la **ficha completa**; (3) un episodio sin descripción en el feed se importa igual y la ficha omite esa línea; (4) el «mismo día» de la unión se calcula en `America/Costa_Rica` sin margen y la duda se marca en vez de unir; (5) al unir manda el nombre del feed del video (default #7). **Ninguna cambia el alcance**: quedaron firmadas en `spec.md` y anotadas en `plan.md` (§ «Precisiones aprobadas en la puerta 2»).
- 2026-10-17 (humano): el **mismo sermón en YouTube y Spotify** se muestra como **una sola entrada con video y audio** (Q18).

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-17 — **Fase 6 arrancada: bloque `[db]` terminado (T901, T902, T903).** Commits `a112ed9`, `b187028`, `9e43394`. Verificado contra **PostgreSQL real**: ciclo up→down→up de las migraciones `000014`/`000015` (incluido el `down` destructivo de `admin_actions`), **38 + 13 comprobaciones de aserciones en verde**, `sqlc-verify` sin deriva, `make lint` limpio y `make test-backend` completo en verde. El catálogo de permisos sigue en **10**. El `dev-backend` detectó y esquivó un **error de transcripción en `data-model.md`** (declaraba dos veces las mismas columnas de disponibilidad): corregido el documento, la migración es la correcta.
- 2026-10-17 — **Fases 4 y 5 cerradas.** `tasks.md` con **46 tareas** (T901–T945 + T907A: 25 backend, 20 frontend, 3 db, 2 infra; 20 paralelizables) y `analyze.md` con **1 crítico, 3 altos, 3 medios y 4 bajos**, **todos resueltos** por el arquitecto: el crítico (el modelo de marcas se contradecía con su propia restricción) quedó partido en **dos ejes independientes** (duplicados y disponibilidad) con su precedencia y sus pruebas; las barreras de red se **centralizaron** en un cliente HTTP único del dominio que usan los tres consumidores (feeds, disponibilidad y carátulas); se **devolvió el origen** a la lista del panel (FR-019); y se aplicó la decisión del humano para A3 (**manda la indisponibilidad confirmada**, con `SC-007` matizado en la spec y el caso fijado por prueba). Cobertura verificada: **30/30 FR y 17/17 SC con tarea**.
- 2026-10-17 — ⚠️ **Incidente de entorno (reparado).** Un reinicio del servidor a mitad de las ediciones dejó **corrupto el archivo de referencia de la rama** (`.git/refs/heads/009-medios-predicas-y-podcasts`, 41 bytes nulos), lo que hizo que git viera el repositorio entero como «todo nuevo» y `HEAD` como roto. **Daño acotado a ese único archivo**: objetos, reflog y todos los archivos del proyecto (verificados byte a byte) estaban intactos. Reparado reescribiendo la referencia al último commit (`037321a`) y limpiando los nulos de los reflogs; la historia y las ediciones se recuperaron completas. **Causa probable**: escritura fallida en el disco de Windows (`/mnt/d/…`), justo lo que advierte `make doctor` («muévelo a Ubuntu para evitar lentitud y problemas de permisos»). **Recomendación**: mover el proyecto a `~/proyectos/` en Ubuntu.
- 2026-10-17 — **Puerta 2: plan APROBADO** por el humano («Aplica las 5 y apruebo el plan»), con las 5 precisiones aplicadas a `spec.md` y anotadas en `plan.md`. Fase → **4/9 · Tareas**.
- 2026-10-17 — **Fase 3 entregada**: `plan.md` (18 decisiones técnicas, 12 riesgos), `research.md` (14 incógnitas con evidencia real: XML del feed, pruebas HTTP de disponibilidad), `data-model.md` (3 tablas + 2 migraciones), `contracts/openapi.yaml` (delta aditivo, 0.8.0 → **0.9.0**, 7 endpoints) y `quickstart.md`; `ux.md` con 7 pantallas y los flujos de «Buscar episodios nuevos» y unir duplicados. **Puerta 2 pendiente** con **5 puntos abiertos** señalados por el arquitecto.
- 2026-10-17 — **Puerta 1: spec v2 APROBADA** por el humano («Apruebo la spec v2»). Fase → **3/9 · Planificar**. Se arranca el plan técnico (`arquitecto`) y `ux.md` (`disenador-ux`).
- 2026-10-17 — **Ajustes aprobados por el humano antes de la puerta 1** (aplicados a la v2): (1) **regla del «estante de 15»** — un episodio que sale de la ventana del feed **no se oculta**: se comprueba si su enlace sigue disponible y solo se oculta cuando la comprobación **confirma** que ya no lo está, con regla a prueba de fallos (nunca se oculta por falta de certeza); (2) los enlaces externos se abren en **pestaña nueva anunciada**. Queda señalado para la puerta 1 el **default #7** (al unir video y audio manda el nombre del feed de YouTube).
- 2026-10-17 — **Spec v2** redactada tras las respuestas del cliente: **6 historias de usuario**, **FR-001…FR-030**, **SC-001…SC-017** (criterios manuales SC-008 lector de pantalla y SC-012 usabilidad), **10 consecuencias asumidas**, **16 defaults revisables** y **0 preguntas abiertas**. Limpiadas las referencias técnicas a migraciones que arrastraba la v1. Roadmap: **decisión 27** y corrección de la 26. **Puerta 1 pendiente.**
- 2026-10-17 — **Cierre de sesión.** Costo del bug del date picker **cerrado** en **68,8 M ≈ $1,41** (desde su rama muerta) y traído a `main` con **cherry-pick** (`ee02ab1`); la rama de F9 quedó **rebasada sobre `main`** y el informe del proyecto ya muestra ese bug como **cerrado**. **`main` quedó 1 commit por delante de `origin`**: el `git push` **falló por falta de credenciales** en el entorno del agente (no hay `gh`, ni llave SSH, ni `credential.helper`, ni token) — **lo empuja el humano**. El mismo problema bloqueará **abrir el PR** en la fase 9 si no se resuelve antes.
- 2026-10-17 — **Inicio de F9**: rama `009-medios-predicas-y-podcasts` creada desde `main`; **spec v1** redactada por `analista-producto` (7 historias, FR-001…FR-028, SC-001…SC-016, criterios manuales SC-008 y SC-012); **cuestionario para el cliente** escrito en `docs/producto/cuestionario-F9-medios.md`; roadmap → F9 **en curso** y decisión 26 registrada. **Puerta 1 pendiente** de las respuestas del cliente (24 preguntas abiertas).
- 2026-10-17 — Contexto previo del mismo día: cerrados y fusionados el **PR #11** (selector de fecha/hora del panel) y el **PR #12** (flaky del CI); el **defecto conocido 23** queda **resuelto** (decisión 25 del roadmap), así que F9 ya no arrastra ese riesgo.
