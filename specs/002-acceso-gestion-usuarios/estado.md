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
| Fase | 1/9 · Especificar |
| Ciclo de corrección | 0/3 |
| Próximo paso | Aprobación humana de la spec (Q1–Q5 ya resueltas) |
| Bloqueado por | — |
| Actualizado | 2026-10-04 |

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

- 2026-10-04 — El humano pidió que la recuperación de contraseña por auto-servicio con correo quede registrada como idea futura/backlog (registrada en `roadmap.md` §5 y §6 decisión 9). En el MVP la contraseña la restablece un administrador.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-04 — El humano fijó la política de contraseñas (mín. 8, mayúsculas+minúsculas+números+especiales, distinta del nombre y del correo) y el límite de 5 intentos fallidos; `analista-producto` los incorporó a la spec.
- 2026-10-04 — Aclaraciones Q1–Q5 resueltas con el humano; `analista-producto` actualizó la spec (cero marcas pendientes).
- 2026-10-04 — Inicio de F2. Spec redactada por `analista-producto` con 5 aclaraciones abiertas (Q1–Q5); pendiente de aprobación humana.
