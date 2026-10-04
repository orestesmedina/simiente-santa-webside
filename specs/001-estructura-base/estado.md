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
| Fase | 7/9 · Validar (validación aprobada; kit actualizado a 1.6.4; falta empujar la rama y el merge humano) |
| Ciclo de corrección | 1/3 (cerrado en verde) |
| Bloqueado por | **Push al remoto no posible desde este entorno:** no hay `gh`, ni credential helper, ni `~/.git-credentials`, ni `~/.ssh`. El kit ya está corregido (**1.6.4, `2957bb0`**, commit local `74b01c1`), así que **g1 está resuelto** (`golangci-lint-action@v9` + `v2.14.0`) y el CI debería pasar al re-ejecutarse. Falta que el humano empuje la rama. |
| Próximo paso | El humano empuja `001-estructura-base` (3 commits locales: kit 1.6.4, costos y estado) → el CI del PR [#1](https://github.com/orestesmedina/simiente-santa-webside/pull/1) re-corre → si queda verde, **aprobación humana del merge** → fusionar y cerrar (`make costos CERRAR=1`). Pendientes: T030 (receta SC-007, humano) y T033 (protección de rama, depende de checks verdes). Pasos manuales de la actualización: instalar Node 24 en local y alinear el README (Node 22→24; «Go 1.23+»→1.26+). |
| Actualizado | 2026-10-03 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano | 2026-09-30 | `spec.md` — "Status: Approved (2026-09-30) … aprobada por el humano tras revisar el delta del 2026-09-30" |
| Plan | aprobado | humano | 2026-09-30 | `tasks.md` — "plan aprobado el 2026-09-30" |
| PR / merge | **bloqueado** | humano | 2026-10-03 | PR [#1](https://github.com/orestesmedina/simiente-santa-webside/pull/1) abierto. El bloqueo por el kit (g1) ya está resuelto con la actualización a 1.6.4; falta empujar la rama y re-ejecutar el CI. No se fusiona sin aprobación explícita del humano. |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno bloqueante) Menores abiertos tras el ciclo 1, registrados para F2/despliegue:
  - **Despliegue:** nginx del frontend corre master como root (M1-seg); imágenes con etiquetas flotantes/`:latest` en distroless (N4/M2-seg); sin cabeceras de seguridad ni `server_tokens off` (M3/M4-seg); puertos de compose publicados en `0.0.0.0` (M9-seg).
  - **F2:** `APP_ENV != development` debería exigir configuración explícita (M7-seg); límites de tamaño de cuerpo con el primer endpoint escribible; idioma de identificadores TS (N6); `WithTx` con `ctx` cancelado (N5); `NotFound`/`MethodNotAllowed` sin referenciar en el path de `/healthz` (N7).
  - **Operación local:** `make up` no avisa si un puerto host está ocupado (H1-01); el healthcheck del `db` pasa aunque las credenciales del volumen estén desalineadas (H1-04); `app_test` no se crea en local (H1-05); e2e requiere librerías del navegador (H1-QE-01).

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- **2026-09-30 — Confirmaciones aplicadas a `tasks.md`:** sobre de éxito = DTO directo (sin wrapper `{"data":…}`); el 503 de `/healthz` usa sobre de error `database_unavailable`; la verificación de la receta (SC-007) la hace el humano en rama descartable.
- **2026-10-03 (hecho) — Versión de Go:** el humano decidió instalar Go 1.27; se instaló **Go 1.27.1** en `~/.local/go1.27.1` (sin `sudo`, con checksum verificado) y se puso por delante en el `PATH` (`~/.local/bin/go` y `~/.local/go1.27.1/bin` en `~/.profile` y `~/.bashrc`). `make doctor` reporta Go 1.27.1. Ya no bloquea T002.
- **2026-10-03 (hecho) — Node/npm en WSL:** `npm` resolvía al de Windows. Se instaló **Node 22.23.3** (npm 10.9.9) en `~/.local/node-v22.23.3` (tarball oficial con checksum verificado, sin `sudo`) y se puso por delante en el `PATH` (`~/.local/bin/{node,npm,npx,corepack}` y `~/.local/node-v22.23.3/bin` en `~/.profile` y `~/.bashrc`). `make doctor`: "Las herramientas de desarrollo son las de Ubuntu". El `nodejs` de `apt` sigue instalado pero queda eclipsado.
- **2026-10-03 (hecho) — Herramientas de calidad:** instaladas en `~/.local/bin` (sin `sudo`): `golangci-lint` 2.14.0, `govulncheck`, `golang-migrate` (con `-tags postgres`, para `make db-migrate` de T005) y `gitleaks` `v8.27.2` (ruta antigua `zricethezav`, porque los releases nuevos no resuelven por módulo; su comando `protect` que usa el hook funciona). `make doctor`: las cuatro ✓.
- **2026-10-03 (hecho) — Cabeceras de C:** el humano instaló `build-essential`/`libc6-dev` (`sudo apt install -y build-essential`). Verificado: `go build` con cgo y `go build -race` compilan. Entorno listo para paridad con el CI.
- **2026-10-03 (hecho) — Kit actualizado a 1.6.4 (`2957bb0`):** `make actualizar-kit` trajo la corrección de g1 y más. Instalado y verificado (`make verificar-kit` OK; `make doctor`: kit al día, config de agentes al día). Commit local `74b01c1` (aún sin empujar, por falta de credenciales en este entorno). Resuelve los puntos P1–P3 de la propuesta T035: `golangci-lint-action@v9` + `v2.14.0`, `migrate`/`govulncheck` fijados, actions a Node 24, `postgres:16-alpine` en el CI y `make help` con dígitos. Trae además `make costos` (commit `7b645d1` con `costos.json`) y `make novedades`.
- **2026-09-30 (pendiente de confirmar, no se implementa en F1) — Sesiones de F2 (D-A7):** cookie `httpOnly` con sesión en servidor.
- **2026-09-30 (pendiente) — Dudas de contenido de `ux.md` §7.2.1–§7.2.2:** mostrar o no `error.message` como apoyo en el error A, y plegable técnico con `details` del 503 para quien opera.
- **2026-10-03 (nota técnica, T002) — Orden de `backend/go.mod`:** para cumplir la verificación literal de T002 (`head -1 go.mod` declara `go 1.27`), la directiva `go` va en la primera línea y `module` después. Verificado que el toolchain Go 1.27.1 lo preserva tras `go mod download`/`build`/`verify` (y que `go mod tidy` no lo reordena). Revisar en T010 si conviene normalizarlo al orden convencional (`module` primero) y ajustar la verificación de T002 en consecuencia.
- **2026-10-03 (propuesta al kit, T035) — `make help` oculta `e2e`:** el regex del recetario `help` del Makefile del kit (`^[a-zA-Z_-]+:.*?## `) no incluye dígitos, así que el target `e2e` de `proyecto.mk` funciona pero no aparece en `make help`. Es item del mismo tipo que el (f) de T035: proponer al repo del kit ampliar la clase a `[a-zA-Z0-9_-]`. No se toca el Makefile (regla 10).

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-03 — **Kit actualizado a 1.6.4 (`2957bb0`) — g1 resuelto, listo para re-ejecutar CI.** `make verificar-kit` OK y `make doctor` sin problemas de kit. Commiteado en `74b01c1` (22 archivos del kit: `golangci-lint-action@v9` + `v2.14.0`, `migrate`/`govulncheck` fijados, Node 24, `postgres:16-alpine`, `make costos`/`novedades`, hook de `costos.json` cerrado, `make help` con dígitos). `costos.json` de F1 en `7b645d1` (`make costos`, abierto, $6.76 equivalentes). **Push imposible desde este entorno** (sin `gh` ni credenciales git/SSH): queda en manos del humano empujar los 3 commits y re-lanzar el CI del PR #1. Pasos manuales pendientes: Node 24 local y alinear README.
- 2026-10-03 — **Merge de F1 bloqueado por el kit (decisión humana).** PR [#1](https://github.com/orestesmedina/simiente-santa-webside/pull/1) abierto; el job `Backend (Go)` sale rojo porque `golangci/golangci-lint-action@v6` (sin `version:` en el `ci.yml` del kit) instala v1.64.8, compilada con go1.24, y no puede analizar un módulo `go 1.27` (run [37161716733](https://github.com/orestesmedina/simiente-santa-webside/actions/runs/37161716733)). No es defecto de F1: `make ci` local verde con 2.14.0 y los otros 5 jobs verdes. El humano eligió **opción A: no fusionar hasta que el kit se corrija** y se le entregó el texto de propuesta al kit (10 puntos, P1–P3). T031/T032 hechas; T033 pendiente de checks verdes; T035 es la vía de desbloqueo.
- 2026-10-03 — **T034 decidido y aplicado:** el humano eligió `postgres:16-alpine` para desarrollo; `85ba674` actualiza `docker-compose.yml` y el README. `make up` + `/healthz` 200 verificados; `make down` al terminar. Queda como pendiente del kit la imagen del `ci.yml` (T035). Incidente: deriva de credencial en el volumen `pgdata` resuelta con `ALTER ROLE app PASSWORD` sin borrar el volumen.
- 2026-10-03 — **Fase 7 (validar) cerrada — validación APROBADA (ciclo 1/3).** Veredictos iniciales: `revisor-codigo` RECHAZADO (B1 QF1011; M1/M2/M3) y `seguridad` RECHAZADO (B1 GO-2026-5970 en `x/text`), `qa-tester` sin bloqueantes. Ciclo 1 aplicado con aprobación del humano (B1 + menores baratos): `aa60f35` x/text→v0.39 (`govulncheck` exit 0), `87ed7e2` QF1011, `f20aa8a` M2 (logger por petición con productor), `ef2faca`+`1401aa0` M6, `f31c152` M5, `1c18315` M8, `af8a51d` N2, `d1cbd50` N3, `234d75b` M1 (frontend), `3e09619` M3 (`ux.md`). Revalidado: **`make ci` en verde verificado por el orquestador** (lint 0 issues, 25/25 frontend, `govulncheck` sin vulnerabilidades, `npm audit` 0). Reportes: `revision-2026-10-03-{qa,codigo,seguridad}.md`. Próximo: aprobaciones humanas de T030–T035 y entrega (PR).
- 2026-10-03 — **Backend de F1 completo (T006–T018):** sqlc (R5: el CLI exige ≥1 consulta → `internal/db/` sin generar en F1; v1.31.1), `platform` (config, logger, apperr, database, httpserver, middleware, testutil), dominio `status` + `GET /healthz` y suite del sobre (SC-008/SC-009). Commits T006 `5542b2a`, T007 `e2ba57b`, T008 `c1ef24f`, T009 `d4fda46`, T010 `aebbace`, T011 `a032e74`, T012 `7e5ab61`, T013 `dc962a7`, T014 `8628ec1`, T015 `ec24b76`, T016 `8bb6322`, T017 `4057959`, T018 `1262ba1`, más el fix `c43f000` (timeout de `Ping` → 503 en vez de 500, hallado por el humo de T017). `go test ./...` e integración en verde; service `status` 100 %. Próximo: T019 (Dockerfile backend) y frontend.
- 2026-10-03 — **Fase 1 de `tasks.md` completada (T001–T005):** T003+T004 (`91c3cc3`, `28ff85b`), T001 (`5042cab`), T005 (`e1a40b4`), T002 (`4d45cc8`). Contrato vivo idéntico byte a byte; módulo Go 1.27 con pgx v5.11.0 (`go mod verify` OK, sin `go mod tidy`); migración baseline aplicada/revertida con el toolchain real (`\dt` solo `schema_migrations`). Próximo: T006 (sqlc) y grupo P3.
- 2026-10-03 — Entorno completo: `build-essential`/`libc6-dev` instalados por el humano; verificados `go build` con cgo y `-race`. Ya no hay bloqueos para la fase 6.
- 2026-10-03 — Herramientas de calidad instaladas en `~/.local/bin` (golangci-lint, govulncheck, golang-migrate con `postgres`, gitleaks v8.27.2). `make doctor` las da por buenas. Detectado: faltan cabeceras de C (`libc6-dev`) para cgo/`-race`.
- 2026-10-03 — Node 22.23.3 + npm 10.9.9 instalados en el espacio de usuario y activos en el `PATH`; resuelto el `npm` de Windows. `make doctor`: herramientas de Ubuntu ✓.
- 2026-10-03 — Go **1.27.1** instalado en el espacio de usuario (`~/.local/go1.27.1`) y activo en el `PATH`; `make doctor` lo confirma. Queda resuelto el bloqueo de T002.
- 2026-10-03 — Estado reconstruido al retomar (no existía `estado.md`). Fases 1–5 cerradas; fase 6 (implementar) pendiente de arranque (el humano decidió no arrancar aún). No hay código: no existen `backend/` ni `frontend/`. Aprobaciones de spec y plan registradas según consta en `spec.md` y `tasks.md` del 2026-09-30.
- 2026-10-03 — Se ajustó `docs/producto/roadmap.md` (una tabla F1–F9 con columna `Estado`; F1 en curso) y `tasks.md` (tareas como casillas `- [ ]`) para que `make estado` lea roadmap y progreso. Decidido: instalar Go 1.27.
