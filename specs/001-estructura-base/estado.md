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
| Bloqueado por | Pendiente (sudo): `build-essential`/`libc6-dev` para compilar con cgo (`go test -race` de CI). Mientras tanto, usar `CGO_ENABLED=0` |
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
- **2026-10-03 (hecho) — Node/npm en WSL:** `npm` resolvía al de Windows. Se instaló **Node 22.23.3** (npm 10.9.9) en `~/.local/node-v22.23.3` (tarball oficial con checksum verificado, sin `sudo`) y se puso por delante en el `PATH` (`~/.local/bin/{node,npm,npx,corepack}` y `~/.local/node-v22.23.3/bin` en `~/.profile` y `~/.bashrc`). `make doctor`: "Las herramientas de desarrollo son las de Ubuntu". El `nodejs` de `apt` sigue instalado pero queda eclipsado.
- **2026-10-03 (hecho) — Herramientas de calidad:** instaladas en `~/.local/bin` (sin `sudo`): `golangci-lint` 2.14.0, `govulncheck`, `golang-migrate` (con `-tags postgres`, para `make db-migrate` de T005) y `gitleaks` `v8.27.2` (ruta antigua `zricethezav`, porque los releases nuevos no resuelven por módulo; su comando `protect` que usa el hook funciona). `make doctor`: las cuatro ✓.
- **2026-10-03 (pendiente, requiere `sudo`) — Cabeceras de C:** falta `libc6-dev`/`build-essential`; gcc está pero sin `grp.h`/`pthread.h`, así que toda build con cgo falla (incluido `go test -race`, que corre el CI). Comando: `sudo apt install -y build-essential`. Alternativa sin `sudo`: `CGO_ENABLED=0` (pierde `-race`).
- **2026-09-30 (pendiente de confirmar, no se implementa en F1) — Sesiones de F2 (D-A7):** cookie `httpOnly` con sesión en servidor.
- **2026-09-30 (pendiente) — Dudas de contenido de `ux.md` §7.2.1–§7.2.2:** mostrar o no `error.message` como apoyo en el error A, y plegable técnico con `details` del 503 para quien opera.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-03 — Herramientas de calidad instaladas en `~/.local/bin` (golangci-lint, govulncheck, golang-migrate con `postgres`, gitleaks v8.27.2). `make doctor` las da por buenas. Detectado: faltan cabeceras de C (`libc6-dev`) para cgo/`-race`.
- 2026-10-03 — Node 22.23.3 + npm 10.9.9 instalados en el espacio de usuario y activos en el `PATH`; resuelto el `npm` de Windows. `make doctor`: herramientas de Ubuntu ✓.
- 2026-10-03 — Go **1.27.1** instalado en el espacio de usuario (`~/.local/go1.27.1`) y activo en el `PATH`; `make doctor` lo confirma. Queda resuelto el bloqueo de T002.
- 2026-10-03 — Estado reconstruido al retomar (no existía `estado.md`). Fases 1–5 cerradas; fase 6 (implementar) pendiente de arranque (el humano decidió no arrancar aún). No hay código: no existen `backend/` ni `frontend/`. Aprobaciones de spec y plan registradas según consta en `spec.md` y `tasks.md` del 2026-09-30.
- 2026-10-03 — Se ajustó `docs/producto/roadmap.md` (una tabla F1–F9 con columna `Estado`; F1 en curso) y `tasks.md` (tareas como casillas `- [ ]`) para que `make estado` lea roadmap y progreso. Decidido: instalar Go 1.27.
