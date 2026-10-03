# Estado: 001-estructura-base

<!--
Lo mantiene el ORQUESTADOR. Reconstruido el 2026-10-03 al retomar (no existía).
Se actualiza en cada cambio de fase, puerta, ciclo de corrección y al cerrar la sesión.
La fuente de verdad de la ejecución es tasks.md; el alto nivel, docs/producto/roadmap.md.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 001-estructura-base |
| Flujo | equipo-feature |
| Fase | 6/9 · Implementar |
| Ciclo de corrección | 0/3 |
| Próximo paso | Fase 1 de `tasks.md`: T001 (contrato vivo) → T002/T005 en el mismo PR, T003, T004 |
| Bloqueado por | Solo el frontend: falta `npm` en Ubuntu (bloquea T020+). Go ya resuelto |
| Actualizado | 2026-10-03 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano | 2026-09-30 | `spec.md` — "Status: Approved (2026-09-30) … aprobada por el humano tras revisar el delta del 2026-09-30" |
| Plan | aprobado | humano | 2026-09-30 | `tasks.md` — "plan aprobado el 2026-09-30" |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno)

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- **2026-09-30 — Confirmaciones aplicadas a `tasks.md`:** sobre de éxito = DTO directo (sin wrapper `{"data":…}`); el 503 de `/healthz` usa sobre de error `database_unavailable`; la verificación de la receta (SC-007) la hace el humano en rama descartable.
- **2026-10-03 (hecho) — Versión de Go:** el humano decidió instalar Go 1.27; se instaló **Go 1.27.1** en `~/.local/go1.27.1` (sin `sudo`, con checksum verificado) y se puso por delante en el `PATH` (`~/.local/bin/go` y `~/.local/go1.27.1/bin` en `~/.profile` y `~/.bashrc`). `make doctor` reporta Go 1.27.1. Ya no bloquea T002.
- **2026-09-30 (pendiente de confirmar, no se implementa en F1) — Sesiones de F2 (D-A7):** cookie `httpOnly` con sesión en servidor.
- **2026-09-30 (pendiente) — Dudas de contenido de `ux.md` §7.2.1–§7.2.2:** mostrar o no `error.message` como apoyo en el error A, y plegable técnico con `details` del 503 para quien opera.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-03 — Go **1.27.1** instalado en el espacio de usuario (`~/.local/go1.27.1`) y activo en el `PATH`; `make doctor` lo confirma. Queda resuelto el bloqueo de T002.
- 2026-10-03 — Estado reconstruido al retomar (no existía `estado.md`). Fases 1–5 cerradas; fase 6 (implementar) pendiente de arranque (el humano decidió no arrancar aún). No hay código: no existen `backend/` ni `frontend/`. Aprobaciones de spec y plan registradas según consta en `spec.md` y `tasks.md` del 2026-09-30.
- 2026-10-03 — Se ajustó `docs/producto/roadmap.md` (una tabla F1–F9 con columna `Estado`; F1 en curso) y `tasks.md` (tareas como casillas `- [ ]`) para que `make estado` lea roadmap y progreso. Decidido: instalar Go 1.27.
