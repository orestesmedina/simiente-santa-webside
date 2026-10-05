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
| Ciclo de corrección | 0/3 |
| Próximo paso | Redactar `tasks.md` (`arquitecto`) y luego `analyze` (`revisor-codigo`) |
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

- (ninguno)

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-04 — El humano pidió que la recuperación de contraseña por auto-servicio con correo quede registrada como idea futura/backlog (registrada en `roadmap.md` §5 y §6 decisión 9). En el MVP la contraseña la restablece un administrador.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

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
