# Estado: 002-acceso-gestion-usuarios

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 002-acceso-gestion-usuarios |
| Flujo | equipo-feature |
| Fase | 3/9 · Tareas y coherencia |
| Ciclo de corrección | 1/3 (`analyze`: RECHAZADO en coherencia documental) |
| Próximo paso | Re-analizar para verificar el cierre de F-01…F-15 |
| Bloqueado por | — |
| Actualizado | 2026-10-04 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano | 2026-10-04 | "apruebo la spec actualizada" (con auditoría, US8 + FR-021…FR-026) |
| Plan | aprobado | humano | 2026-10-04 | "apruebo el plan" |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- **F-01 (crítico)** — la pantalla de auditoría de `ux.md` no es construible con el contrato: pide nombre/correo que los DTOs no exponen y afirma que se guarda el correo de intentos sin cuenta (no se guarda). Resolución: añadir `userName`/`actorName` al DTO (derivado cuando la cuenta se identifica) y mostrar "Intento sin cuenta asociada" en el resto.
- **F-02 (mayor)** — falta tarea para la página de Inicio del panel (`/panel`) y el caso "cuenta sin permisos de módulo".
- **F-03 (mayor)** — semántica contradictoria del 5.º intento fallido (FR-006 vs R5 vs quickstart vs data-model). Resolución: el 5.º responde genérico y crea el bloqueo; el 429 aplica desde el 6.º intento.
- **F-04 (mayor)** — `ux.md` dice "sin paginar" con buscador en cliente; plan/contrato/tareas usan paginación. Resolución: mantener paginación; retirar el buscador de texto libre del MVP.
- **F-05 (mayor)** — rutas divergentes. Resolución: las del plan (`/login`, `/cambiar-contrasena`, `/panel`, `/panel/usuarios`, `/panel/roles`, `/panel/auditoria`, `/sin-permiso`).
- **F-06 (mayor)** — pantalla de "Puesta en marcha" en `ux.md`. Resolución: retirarla (la inicialización es por API con `BOOTSTRAP_TOKEN`).
- **F-07…F-15 (menores)** — cobertura FR-021 (T247), filtro de auditoría (inicialización), política de contraseñas (igual vs contener, tope de longitud), inventario de componentes UI, código de permiso, número de operaciones, reparto de contadores, "F3–F9", ámbito del guard de contraseña y descripción de `make ci`.

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-04 — El humano pidió que la recuperación de contraseña por auto-servicio con correo quede registrada como idea futura/backlog (registrada en `roadmap.md` §5 y §6 decisión 9). En el MVP la contraseña la restablece un administrador.
- 2026-10-04 — **Rutas del panel (coherencia `analyze` F-05):** se adoptan las del plan (`/login`, `/cambiar-contrasena`, `/panel`, `/panel/usuarios`, `/panel/roles`, `/panel/auditoria`, `/sin-permiso`); se alinea `ux.md`, que usaba `/entrar` y `/panel/cuenta`.
- 2026-10-04 — **Pantalla de "Puesta en marcha" (coherencia `analyze` F-06):** se retira de `ux.md`; la inicialización única es por API con `BOOTSTRAP_TOKEN` y no debe exponer el secreto en el navegador.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-04 — Ciclo de corrección 1: aplicados F-01…F-15 (`arquitecto`, `disenador-ux`, `analista-producto`). Pendiente re-analizar.
- 2026-10-04 — `analyze` (`revisor-codigo`): **RECHAZADO en coherencia documental** (F-01 crítico; F-02–F-06 mayores; F-07–F-15 menores). Ciclo de corrección 1/3; se corrigen con `arquitecto`, `disenador-ux` y `analista-producto`.
- 2026-10-04 — **`tasks.md` redactado** por `arquitecto` (53 tareas, T201–T253). Pendiente `analyze`.
- 2026-10-04 — **Plan de F2 aprobado por el humano** ("apruebo el plan"). Inicio de la fase 3 (tareas y coherencia).
- 2026-10-04 — Plan y `ux.md` actualizados con la **auditoría** (tablas `login_events`/`admin_actions`, endpoints de consulta, pantalla `/panel/auditoria`) y Redis sin persistencia. Pendiente la aprobación del plan.
- 2026-10-04 — **Spec actualizada re-aprobada por el humano** ("apruebo la spec actualizada") con la auditoría incluida. Se actualiza el plan.
- 2026-10-04 — **Cambio de alcance aprobado por el humano: auditoría en F2** (último acceso, historial de inicios de sesión y acciones administrativas, visible en el panel). `analista-producto` añadió US8 y FR-021…FR-026; Redis **sin persistencia**. Requiere re-aprobar la spec.
- 2026-10-04 — El humano resolvió los puntos del plan: **D-A7 confirmada con Redis** (sesiones en Redis por objetivo de aprendizaje), sesión **1 h máxima + 30 min de inactividad**, cuenta con **nombre, apellidos, correo y teléfono**, y token `BOOTSTRAP_TOKEN` mantenido. Spec, plan, ux.md y `docs/tecnico/decisiones.md` (D-A7) actualizados.
- 2026-10-04 — **Plan técnico y `ux.md` redactados** (`arquitecto`, `disenador-ux`); pendiente la aprobación humana del plan y la confirmación de D-A7.
- 2026-10-04 — **Spec de F2 aprobada por el humano** ("apruebo la spec"). Inicio de la fase 2 (plan técnico).
- 2026-10-04 — El humano fijó el bloqueo por intentos fallidos en **15 minutos** (FR-006); incorporado a la spec.
- 2026-10-04 — El humano fijó la política de contraseñas (mín. 8, mayúsculas+minúsculas+números+especiales, distinta del nombre y del correo) y el límite de 5 intentos fallidos; `analista-producto` los incorporó a la spec.
- 2026-10-04 — Aclaraciones Q1–Q5 resueltas con el humano; `analista-producto` actualizó la spec (cero marcas pendientes).
- 2026-10-04 — Inicio de F2. Spec redactada por `analista-producto` con 5 aclaraciones abiertas (Q1–Q5); pendiente de aprobación humana.
