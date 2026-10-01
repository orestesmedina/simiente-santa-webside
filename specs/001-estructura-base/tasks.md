# Tasks: Estructura base del proyecto (F1)

**Branch**: `001-estructura-base` | **Date**: 2026-09-30 | **Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md)

**Input**: plan aprobado el 2026-09-30 (`plan.md`, D1–D23) · arquitectura y decisiones aprobadas (`docs/tecnico/arquitectura.md`, `docs/tecnico/decisiones.md`) · `data-model.md` · `contracts/openapi.yaml` · `quickstart.md` · constitución (`.specify/memory/constitution.md`) · skills `go-backend`, `postgres-db`, `react-frontend` · kit del proyecto (Makefile, `docker-compose.yml`, `proyecto.mk`, `.env.example`, `.github/workflows/ci.yml`).

> **Regla de ejecución**: **un commit por tarea** (Conventional Commits; los hooks del kit lo validan). Toda tarea de código **incluye sus pruebas** en el mismo commit (constitución §III). Quien escribe no aprueba: cada tarea pasa por `qa-tester`, `revisor-codigo` y `seguridad` antes de integrarse (flujo de entrega del orquestador).

## Confirmaciones del humano aplicadas (2026-09-30)

1. **Sobre de éxito = el DTO directo** (`{"status":"ok","database":"connected"}`), sin wrapper `{"data":…}`. Solo los errores llevan sobre `{"error":{"code","message","details"?}}` (D13).
2. **El 503 de `/healthz`** devuelve sobre de error: `{"error":{"code":"database_unavailable","message":"La base de datos no está conectada","details":{"database":"disconnected"}}}` (D7), no el cuerpo informativo anterior.
3. **La verificación de la receta (SC-007) la hace el humano** siguiendo **solo** `docs/tecnico/arquitectura.md` §8, en una rama aparte y descartable (`practica/receta-001`), sin dejar tablas ni funcionalidad de negocio en `main` (T030).

## Leyenda

| Marca | Significado |
|---|---|
| `[backend]` `[frontend]` `[db]` `[infra]` `[docs]` `[humano]` | Capa y subagente responsable (`[backend]`/`[db]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` → `devops`; `[docs]` → `documentador`; `[humano]` → persona, escala al orquestador) |
| `[P1]`…`[P7]` | **Paralelizable** solo dentro de su grupo (archivos distintos y sin dependencia previa entre ellas; ver "Grupos de paralelismo") |
| *(sin `[P]`)* | Secuencial: tiene dependencias previas que deben estar integradas |

## Tabla resumen

| ID | Tarea | Capa | Fase | Paralelo | Depende de |
|---|---|---|---|---|---|
| T001 | Contrato vivo `backend/api/openapi.yaml` | `[backend]` | 1 | P1 | — |
| T002 | Módulo Go 1.27 (`go.mod`/`go.sum`) y carpetas base | `[backend]` | 1 | P1 | — |
| T003 | `proyecto.mk` con targets propios | `[infra]` | 1 | P1 | — |
| T004 | `.env.example` con variables opcionales de puertos | `[infra]` | 1 | P1 | — |
| T005 | Migración baseline `000001` (no-op) | `[db]` | 2 | P2 | — |
| T006 | `backend/sqlc.yaml` + verificación de R5 | `[db]` | 2 | — | T003, T005 |
| T007 | `platform/config` | `[backend]` | 3 | P3 | T002 |
| T008 | `platform/logger` | `[backend]` | 3 | P3 | T002 |
| T009 | `platform/apperr` | `[backend]` | 3 | P3 | T002 |
| T010 | `platform/database` (pool, `Ping`, `WithTx`) | `[backend]` | 3 | P3 | T002 |
| T011 | `platform/httpserver` (`Registrar`, `WriteJSON`/`WriteError`, servidor) | `[backend]` | 3 | — | T008, T009 |
| T012 | `platform/middleware` (request-id, recover, logging, CORS) | `[backend]` | 3 | — | T011 |
| T013 | `platform/testing` (helpers compartidos) | `[backend]` | 3 | — | T010 |
| T014 | `status`: `model.go` + `service.go` | `[backend]` | 4 | P4 | T009 |
| T015 | `status`: `repository.go` (+ integración) | `[backend]` | 4 | — | T010, T013, T014 |
| T016 | `status`: `handler.go` (200/503 con sobres) | `[backend]` | 4 | — | T011, T014 |
| T017 | `status/routes.go` + `cmd/api/main.go` (cableado y servidor) | `[backend]` | 4 | — | T007, T008, T010, T012, T016 |
| T018 | Suite del sobre de respuestas (SC-008/SC-009) | `[backend]` | 4 | — | T012, T017 |
| T019 | `backend/Dockerfile` multi-stage `golang:1.27` | `[infra]` | 5 | — | T017 |
| T020 | Scaffold frontend (Vite + React + TS strict + Tailwind + pruebas) | `[frontend]` | 6 | P5 | T001 |
| T021 | Router, providers y layout (`src/app/`) | `[frontend]` | 6 | P6 | T020 |
| T022 | Tipos generados `src/api/schema.d.ts` (`npm run api:gen`) | `[frontend]` | 6 | P6 | T001, T020 |
| T023 | `src/api/client.ts` (`apiFetch<T>` + `ApiError`) y `src/api/status.ts` | `[frontend]` | 6 | — | T022 |
| T024 | Feature `status` (página, hook, indicadores) + pruebas | `[frontend]` | 6 | — | T021, T023 |
| T025 | `frontend/nginx.conf` + `frontend/Dockerfile` | `[infra]` | 6 | — | T024 |
| T026 | `docker-compose.yml` (servicios completos, puertos parametrizados) | `[infra]` | 7 | — | T019, T025 |
| T027 | E2E Playwright (`e2e/status.spec.ts`) | `[frontend]` | 7 | — | T024, T026 |
| T028 | `README.md` (quickstart, puertos, comandos, versiones) | `[docs]` | 8 | — | T003, T026 |
| T029 | Checklist de verificación de la receta (SC-007) | `[docs]` | 8 | P7 | — |
| T030 | **[humano]** Verificación de la receta en rama descartable (SC-007) | `[humano]` | 9 | — | T018, T026, T029 |
| T031 | **[humano]** Crear remoto GitHub y empujar (R1) | `[humano]` | 9 | — | — |
| T032 | **[humano]** Renombrar rama por defecto `master` → `main` (R1) | `[humano]` | 9 | — | T031 |
| T033 | **[humano]** Activar protección de rama con checks obligatorios (R1) | `[humano]` | 9 | — | T032 |
| T034 | **[humano]** Decidir imagen de PostgreSQL: `postgres:16-alpine` vs `postgres:16.4-alpine` (R10) | `[humano]` | 9 | — | — (idealmente **antes** de T026) |
| T035 | **[humano]** Registrar vencimiento de Node 22 y proponer actualización al kit (R11) | `[humano]` | 9 | — | — |

**Total: 35 tareas** — 14 `[backend]` · 6 `[frontend]` · 5 `[infra]` · 2 `[db]` · 2 `[docs]` · 6 `[humano]` · 13 marcadas `[P]`.

## Grupos de paralelismo

| Grupo | Tareas | Por qué pueden ir en paralelo |
|---|---|---|
| P1 | T001 · T002 · T003 · T004 | Archivos distintos, ninguna depende de otra |
| P2 | T005 (en paralelo con P1) | Solo necesita el servicio `db` del kit, ya levantable |
| P3 | T007 · T008 · T009 · T010 | Paquetes `platform` distintos, sin dependencias entre sí (solo T002) |
| P4 | T014 (en paralelo con T011–T013) | Solo necesita `apperr` (T009); sus archivos son propios del dominio |
| P5 | T020 (en paralelo con las fases 2–5) | El scaffold solo necesita el contrato (T001); no toca nada del backend |
| P6 | T021 · T022 (entre sí) | `src/app/` y `src/api/schema.d.ts` son archivos distintos; ambas solo dependen de T020 |
| P7 | T029 (en paralelo con casi todo) | Documento de plantilla; solo lee `quickstart.md` §9 y `arquitectura.md` §8 |

---

## Fase 1 — Preparación, contrato y herramientas de desarrollo

### T001 · Contrato vivo `backend/api/openapi.yaml` · `[backend]` `[P1]`

- **Archivos**: `backend/api/openapi.yaml` (NUEVO). *No* se toca `specs/001-estructura-base/contracts/openapi.yaml` (snapshot inmutable).
- **Qué hace**: copia **verbatim** el contrato aprobado `specs/001-estructura-base/contracts/openapi.yaml` a `backend/api/openapi.yaml`, que desde este commit es el **documento vivo** (D9, plan §"Contrato OpenAPI"): `/healthz` con `getSystemStatus`, `SystemStatus` (200), `ErrorEnvelope`/`ErrorBody` (404/405/500/503) y `DatabaseUnavailable` con `details.database`. Sin ediciones de contenido.
- **Pruebas incluidas** (§III): — (artefacto de contrato, sin código). Verificación: `diff -u specs/001-estructura-base/contracts/openapi.yaml backend/api/openapi.yaml` sin salida y el YAML parsea sin errores.
- **Criterio de terminado**: los dos archivos son idénticos byte a byte; el contrato queda como única fuente para el código y para `npm run api:gen` (T022).
- **Commit sugerido**: `docs(api): crear backend/api/openapi.yaml como contrato vivo`

### T002 · Módulo Go 1.27 y carpetas base · `[backend]` `[P1]`

- **Archivos**: `backend/go.mod` (NUEVO, `module simiente-santa/backend`, `go 1.27`), `backend/go.sum` (NUEVO), `backend/internal/db/queries/.gitkeep` (NUEVO).
- **Qué hace**: crea el módulo Go con la versión **1.27** (D5/D-A5: el CI del kit lee `backend/go.mod`) y añade la única dependencia de runtime de F1: `github.com/jackc/pgx/v5` (D22/D-A8). Crea la carpeta `internal/db/queries/` (vacía en F1; ver `data-model.md`). El path del módulo es provisional `simiente-santa/backend` hasta que exista el remoto (R1).
- **Pruebas incluidas** (§III): — (sin código todavía). Verificación: `cd backend && go mod verify` en verde; `go list -m all` muestra `pgx/v5`; `head -1 go.mod` declara `go 1.27`.
- **Criterio de terminado**: `go build ./...` no falla (paquete vacío válido) y la versión de Go declarada es 1.27. **Nota de secuenciación**: `go mod tidy` **no** se ejecuta en esta tarea (eliminaría `pgx` sin ningún import todavía); la consolidación de `go.sum` ocurre en T010, donde `platform/database` importa `pgx` por primera vez.
- **Commit sugerido**: `chore(backend): crear módulo Go 1.27 con pgx/v5`

### T003 · `proyecto.mk` con los targets propios · `[infra]` `[P1]`

- **Archivos**: `proyecto.mk` (NUEVO, raíz; incluido por el `Makefile` del kit — **no se edita el Makefile**).
- **Qué hace**: define los targets de extensión previstos en D19: `api-gen` (`cd frontend && npm run api:gen`), `sqlc-gen` (`cd backend && sqlc generate`), `sqlc-verify` (regenera y exige `git diff --exit-code` sobre los artefactos generados — R4/R20), y `e2e` (`cd frontend && npx playwright test --config e2e/playwright.config.ts`, ejecución local, D17). Todos con `.PHONY` y descripción `##` para que aparezcan en `make help`.
- **Pruebas incluidas** (§III): — (scripting de build). Verificación: `make help` lista los cuatro targets; `make sqlc-verify` ejecuta sin error de Make (su contenido se valida en T006).
- **Criterio de terminado**: los cuatro targets existen, `make help` los muestra y ningún archivo del kit fue modificado (`git status` solo muestra `proyecto.mk`). `api-gen` y `e2e` no son ejecutables hasta T022/T027 (target definido antes que su contenido: aceptado y documentado aquí).
- **Commit sugerido**: `build(mk): agregar proyecto.mk con api-gen, sqlc-gen, sqlc-verify y e2e`

### T004 · `.env.example` con variables opcionales de puertos · `[infra]` `[P1]`

- **Archivos**: `.env.example` (EDITABLE; no está en `.kit-manifest.json`).
- **Qué hace**: agrega las variables opcionales de puertos que usará compose (T026): `DB_PORT` (5432), `WEB_PORT` (5173); confirma que ya están documentadas `POSTGRES_*`, `DATABASE_URL`, `DATABASE_URL_TEST`, `HTTP_PORT` (8080), `LOG_LEVEL`, `CORS_ALLOWED_ORIGINS`, `VITE_API_URL`. **Sin secretos reales** (FR-008): solo valores de ejemplo de desarrollo.
- **Pruebas incluidas** (§III): — (documento de configuración). Verificación: ningún valor parece secreto real (lo audita `seguridad`); el archivo sigue siendo un ejemplo copiable (`cp .env.example .env`).
- **Criterio de terminado**: `.env.example` documenta todas las variables que lee `platform/config` (T007) y las que consume compose (T026), con su valor por defecto de desarrollo.
- **Commit sugerido**: `chore(env): documentar variables opcionales de puertos en .env.example`

---

## Fase 2 — Capa de datos: migraciones y sqlc

### T005 · Migración baseline `000001` (no-op) · `[db]` `[P2]`

- **Archivos**: `backend/migrations/000001_baseline.up.sql` (NUEVO), `backend/migrations/000001_baseline.down.sql` (NUEVO).
- **Qué hace**: crea la migración baseline **no-op** (solo un comentario) que fija la convención `NNNNNN_descripcion.{up,down}.sql` de la skill `postgres-db` y ejercita el toolchain de punta a punta (D10, `data-model.md`). Sin tablas de negocio: F1 cierra con cero tablas de negocio. `schema_migrations` es la tabla de control de `golang-migrate`, no cuenta.
- **Pruebas incluidas** (§III): ejecución real del toolchain — `make up` (solo `db`) + `make db-migrate` aplica la versión 1 sin error; `migrate -path backend/migrations -database "$DATABASE_URL" down 1` la revierte sin efecto; el paso "Migraciones" del CI del kit pasa (R8: verificar el comportamiento del paso con una migración no-op y registrarlo en el PR).
- **Criterio de terminado**: `make db-migrate` deja la BD en versión 1, el `down` la deja como estaba y no existe ninguna tabla de negocio (`\dt` solo muestra `schema_migrations`).
- **Commit sugerido**: `feat(db): agregar migración baseline 000001 no-op`

### T006 · `backend/sqlc.yaml` y verificación de R5 · `[db]`

- **Archivos**: `backend/sqlc.yaml` (NUEVO). *(Toca `internal/db/` solo si la verificación de R5 genera código.)*
- **Qué hace**: configura sqlc (D6/D-A3): `schema: migrations` · `queries: internal/db/queries` · `out: internal/db`, `engine: postgresql`. A continuación **verifica el riesgo R5**: ejecuta `make sqlc-gen` con `internal/db/queries/` **vacío** (el caso real de `main` en F1) y registra el comportamiento observado como comentario en el propio `sqlc.yaml`: si el CLI genera artefactos vacíos, se commitean; si exige al menos una consulta, se documenta así y `internal/db/` queda sin generar hasta el primer `sqlc generate` del ejercicio de práctica (flujo normal del paso 3 de la receta).
- **Pruebas incluidas** (§III): — (configuración de herramienta). Verificación: `make sqlc-gen` ejecuta (o falla de forma documentada) y `make sqlc-verify` termina en verde en cualquier caso (`git diff --exit-code` sin deriva).
- **Criterio de terminado**: el comportamiento sin consultas queda **escrito** en `backend/sqlc.yaml` (y en la descripción del PR) y `make sqlc-verify` pasa. La versión de `sqlc` usada se registra para generación reproducible (R4; se documenta en el README, T028).
- **Commit sugerido**: `build(sqlc): configurar sqlc sobre backend/migrations y verificar el caso sin consultas (R5)`

---

## Fase 3 — Plataforma interna (`backend/internal/platform/`)

> Reglas de `docs/tecnico/arquitectura.md` §1.2: `cmd → dominio → platform`; `platform` **no conoce dominios**; sin estado global ni `init()` con lógica (R6). Toda la fase cumple §V (`gofmt`, `go vet`, `golangci-lint`, errores con `%w`, nombres en inglés en código).

### T007 · `platform/config` — variables de entorno con validación al arrancar · `[backend]` `[P3]`

- **Archivos**: `backend/internal/platform/config/config.go`, `backend/internal/platform/config/config_test.go` (NUEVOS).
- **Qué hace**: `Load() (Config, error)` con `os.Getenv` + validación al arrancar y valores por defecto de desarrollo (D12/D-A6, sin Viper): `HTTP_ADDR`/`HTTP_PORT`, `DATABASE_URL`, `LOG_LEVEL`, `CORS_ALLOWED_ORIGINS`. Falla rápido con mensaje claro si algo obligatorio falta o es inválido (puerto no numérico, `LOG_LEVEL` desconocido). Ningún secreto en el código (FR-008).
- **Pruebas incluidas** (§III): tabla de casos con `t.Setenv`: valores por defecto en clon limpio (sin `.env`); carga de cada variable; errores por variable obligatoria ausente o inválida, con mensaje que identifica la variable.
- **Criterio de terminado**: `go test ./internal/platform/config/` en verde; un entorno sin `.env` produce una `Config` válida de desarrollo (requisito de `make up` en clon limpio, FR-001).
- **Commit sugerido**: `feat(platform): config por variables de entorno con validación al arrancar`

### T008 · `platform/logger` — `slog` JSON y logger por petición · `[backend]` `[P3]`

- **Archivos**: `backend/internal/platform/logger/logger.go`, `backend/internal/platform/logger/logger_test.go` (NUEVOS).
- **Qué hace**: constructor de `*slog.Logger` con handler **JSON** y nivel desde `LOG_LEVEL`; helper para el **logger por petición** (hij con `request_id`, método y ruta) que usará `middleware/request-id` (D12). Mismo formato y severidad en todo el sistema (US4 esc. 3).
- **Pruebas incluidas** (§III): tabla de casos: nivel aplicado por `LOG_LEVEL`; salida JSON válida con los campos estándar; el logger por petición agrega `request_id`/`método`/`ruta` sin perder los campos padre (usando un handler en memoria que captura registros).
- **Criterio de terminado**: `go test ./internal/platform/logger/` en verde; toda línea emitida es JSON parseable y lleva su nivel.
- **Commit sugerido**: `feat(platform): logger slog JSON con logger por petición`

### T009 · `platform/apperr` — errores de dominio tipados · `[backend]` `[P3]`

- **Archivos**: `backend/internal/platform/apperr/apperr.go`, `backend/internal/platform/apperr/apperr_test.go` (NUEVOS).
- **Qué hace**: tipos `Invalid` (400), `Unauthenticated` (401), `Forbidden` (403), `NotFound` (404), `Conflict` (409), `Internal` (500) y **`DatabaseUnavailable` (503, code `database_unavailable`)** — el kind que usa `/healthz` (D11: "los kinds que usa `/healthz` + `Internal`; crecen bajo demanda"). Cada error lleva `Message` (seguro para el cliente, en español), detalle opcional (`Details`) y el error interno **envuelto con `%w`**, que nunca se serializa (FR-013). Incluye la tabla de traducción de §5.11 de `arquitectura.md` (kind → status HTTP → `error.code`) que consumirá `WriteError`.
- **Pruebas incluidas** (§III): tabla de casos: constructores y `code` asociados; `errors.Is/As` y `Unwrap` con error envuelto `%w`; `Message` nunca contiene el error interno; `Internal` produce mensaje genérico; mapeo kind → status/code correcto para los siete kinds.
- **Criterio de terminado**: `go test ./internal/platform/apperr/` en verde; ningún método de `apperr` expone el error interno (lo verifica también `seguridad`).
- **Commit sugerido**: `feat(platform): apperr con kinds de dominio y traducción a HTTP`

### T010 · `platform/database` — pool `pgx`, `Ping` con timeout y `WithTx` · `[backend]` `[P3]`

- **Archivos**: `backend/internal/platform/database/database.go`, `backend/internal/platform/database/database_test.go`, `backend/internal/platform/database/database_integration_test.go` (`//go:build integration`) (NUEVOS). *Consolida `go.mod`/`go.sum`.*
- **Qué hace**: construcción del `*pgxpool.Pool` desde `DATABASE_URL` (pool **perezoso**: no bloquea ni falla si la BD tarda — R7), helper de salud `Ping(ctx)` con `context.WithTimeout` de **2 s** (D8: estado real por petición, `/healthz` responde en ≤2 s aunque la BD esté caída) y `WithTx(ctx, pool, fn)` para transacciones (D-A3). Primer import real de `pgx/v5` → aquí se ejecuta `go mod tidy` y se consolida `go.sum` (ver T002).
- **Pruebas incluidas** (§III): unitarias (configuración del pool, propagación de `ctx`, timeout de 2 s aplicado) + **integración** con `//go:build integration` contra `DATABASE_URL_TEST` (con *skip* automático si no hay BD): `NewPool` + `Ping` con la BD viva; `Ping` con host inaccesible devuelve error en ≤2 s sin panic; `WithTx` confirma y revierte según el resultado de `fn`.
- **Criterio de terminado**: `go test ./...` y `go test -tags=integration ./...` (con la BD del kit) en verde; `go mod tidy` deja `go.sum` consistente sin perder `pgx/v5`.
- **Commit sugerido**: `feat(platform): pool pgx con Ping por petición y WithTx`

### T011 · `platform/httpserver` — `Registrar`, sobres de respuesta y servidor · `[backend]`

- **Archivos**: `backend/internal/platform/httpserver/registrar.go`, `error.go`, `server.go` y sus `*_test.go` (NUEVOS).
- **Qué hace**: (a) la interfaz `Registrar` (`Handle(method, path, h)`, `Group(prefix, mws...)`) y el tipo `Middleware = func(http.Handler) http.Handler` tal como están en `arquitectura.md` §4, con el adaptador `muxRegistrar` sobre `net/http` (patrones `GET /ruta` de Go 1.22+; D4/D-A4) incluido el fallback 404/405 del router convertido a **sobre de error** (SC-008: la stdlib respondería texto); (b) `WriteJSON` (sobre de éxito = DTO directo, confirmación 1) y `WriteError` (**único** punto de traducción de `apperr` a HTTP, §5.11; `internal` responde mensaje genérico y el detalle interno solo va al log con `request_id`); (c) `New`/`Run`: servidor con `ReadHeaderTimeout`/`ReadTimeout`/`WriteTimeout`/`IdleTimeout` y **apagado ordenado** (`signal.NotifyContext` + `Shutdown` con periodo de gracia).
- **Pruebas incluidas** (§III): tabla de casos + `httptest`: `Handle` publica `method path`; `Group` acumula prefijo y middlewares con el **primero de la lista más externo** (incluye `joinPath`); método no documentado → 405 sobre de error; ruta no documentada → 404 sobre de error; `WriteJSON` con status/`Content-Type`/cuerpo correctos; `WriteError` por cada kind (status + `error.code` + `message` + `details`); `internal` → `"Error interno del servidor"` sin el error interno; `Run` con servidor en puerto efímero y cancelación de contexto → `Shutdown` limpio.
- **Criterio de terminado**: `go test ./internal/platform/httpserver/` en verde; ningún handler puede escribir errores fuera de `WriteError` (revisión de `revisor-codigo`).
- **Commit sugerido**: `feat(platform): httpserver con Registrar, WriteJSON/WriteError y apagado ordenado`

### T012 · `platform/middleware` — cadena transversal en orden aprobado · `[backend]`

- **Archivos**: `backend/internal/platform/middleware/{requestid,recover,logging,cors}.go` + `middleware_test.go` (NUEVOS).
- **Qué hace**: los cuatro middlewares de F1 (D11/D12): `request-id` (toma `X-Request-ID` del cliente o genera uno; lo guarda en el `context` y crea el logger por petición), `recover` (todo `panic` → 500 vía `WriteError`; el proceso nunca cae), `logging` (al terminar: método, ruta, status, duración, `request_id`) y `CORS` mínimo escrito a mano (una cabecera para un `GET` simple sin credenciales; responde los preflight `OPTIONS` y los corta ahí; orígenes desde `CORS_ALLOWED_ORIGINS`) — D16, sin dependencias. Orden de montaje en `httpserver.New`: `request-id → recover → logging → CORS → handler` (F2+ añadirá `rate-limit → [authn → authz → CSRF]`; **no** se implementan ahora, ver "Fuera de F1").
- **Pruebas incluidas** (§III): tabla de casos + `httptest`: la cadena ejecuta en el orden aprobado (middlewares instrumentados que registran su orden); `X-Request-ID` del cliente se respeta y el generado llega al contexto/logger; `recover` convierte un `panic` provocado en 500 con sobre de error y la cadena sigue sirviendo; `logging` emite los cinco campos; `CORS` responde preflight con las cabeceras del origen permitido y sin ellas para un origen no permitido.
- **Criterio de terminado**: `go test ./internal/platform/middleware/` en verde; la cadena de F1 queda montada exactamente en el orden documentado (SC-006: las capacidades transversales existen una sola vez, aquí).
- **Commit sugerido**: `feat(platform): middleware request-id, recover, logging y CORS`

### T013 · `platform/testing` — helpers compartidos de prueba · `[backend]`

- **Archivos**: `backend/internal/platform/testing/{db,http,log}.go` + `testing_test.go` (NUEVOS).
- **Qué hace**: helpers que usarán todas las pruebas (D11): conexión a `DATABASE_URL_TEST` con *skip* automático si no hay BD (y creación del esquema de prueba cuando aplique), utilidades `httptest` (recorders, servidor con la cadena completa) y captura de logs (handler `slog` en memoria) para las aserciones de SC-009.
- **Pruebas incluidas** (§III): pruebas de los propios helpers: el helper de BD hace *skip* sin `DATABASE_URL_TEST`; la captura de logs devuelve las líneas con sus campos; el helper HTTP devuelve respuestas observables.
- **Criterio de terminado**: `go test ./internal/platform/testing/` en verde; las suites de T010, T012 y T018 pueden usar estos helpers sin duplicar código (SC-006).
- **Commit sugerido**: `test(platform): helpers compartidos de testing (BD, httptest, logs)`

---

## Fase 4 — Dominio `status` y `GET /healthz`

> Es el **primer área real construida con la receta** de `docs/tecnico/arquitectura.md` §8 (D21): mismos archivos y convenciones que copiarán F2–F9. Contrato ya fijado en `backend/api/openapi.yaml` (T001) — regla R8.

### T014 · `status`: modelo y servicio · `[backend]` `[P4]`

- **Archivos**: `backend/internal/status/model.go`, `service.go`, `service_test.go` (NUEVOS).
- **Qué hace**: `model.go` con el DTO `SystemStatus` (`{"status":"ok","database":"connected"}`, espejo del contrato, `camelCase`, `additionalProperties: false`); `service.go` con la interfaz `Repository` (**la define quien la consume**, R3: `Ping(ctx) error` con timeout ya aplicado en `platform/database`) y las reglas: conexión viva → `SystemStatus` ok; error de conexión → `apperr.DatabaseUnavailable` con `Details{"database":"disconnected"}` (confirmación 2: el 503 es sobre de error). Sin HTTP, sin SQL, sin `net/http` ni `pgx` (R4).
- **Pruebas incluidas** (§III): tabla de casos con **fake** de `Repository`: BD conectada → `{"status":"ok","database":"connected"}`; BD caída → `apperr` con code `database_unavailable` y `details.database="disconnected"`; error inesperado del repo → error envuelto con `%w` (llegará como `internal`); cancelación de `ctx` respetada.
- **Criterio de terminado**: `go test ./internal/status/` en verde con cobertura del service **≥80 %** (`go test -cover`, umbral de §III).
- **Commit sugerido**: `feat(status): modelo y servicio del estado del sistema`

### T015 · `status`: repositorio sobre `pgxpool` · `[backend]`

- **Archivos**: `backend/internal/status/repository.go`, `repository_test.go` (`//go:build integration`) (NUEVOS).
- **Qué hace**: implementación concreta de la interfaz `Repository` (definida en `service.go`): `Ping` por petición con timeout de 2 s sobre el `*pgxpool.Pool` compartido (D8: estado **real** en cada consulta, no memorizado — FR-003). Errores envueltos con `fmt.Errorf("ping database: %w", err)` (R7). Sin SQL de negocio: F1 solo hace `Ping` (D6); el primer código generado por sqlc llega con el ejercicio de práctica.
- **Pruebas incluidas** (§III): **integración** con `//go:build integration` contra `DATABASE_URL_TEST` (helpers de T013, *skip* sin BD): `Ping` con la BD viva devuelve nil; con la BD detenida devuelve error en ≤2 s; dos consultas seguidas reflejan el estado **actual** (caída y recuperación sin reiniciar — escenario 4 de US2).
- **Criterio de terminado**: `go test -tags=integration ./internal/status/` en verde contra el servicio `db` del kit; el comportamiento "caída y recuperación sin reinicio" queda probado.
- **Commit sugerido**: `feat(status): repositorio de estado sobre pgxpool`

### T016 · `status`: handler HTTP de `/healthz` · `[backend]`

- **Archivos**: `backend/internal/status/handler.go`, `handler_test.go` (NUEVOS).
- **Qué hace**: la interfaz `Service` (quien la consume) + solo HTTP: `GetSystemStatus` responde **200** con el DTO `SystemStatus` (sobre de éxito = DTO directo, confirmación 1) o **503** con `WriteError(apperr.DatabaseUnavailable…)`, y fija `Cache-Control: no-store` en ambos (D7, contrato). Todos los errores salen por `httpserver.WriteError`; ningún código de estado escrito a mano (R7).
- **Pruebas incluidas** (§III): `httptest.NewRecorder` + **fake** de `Service`: 200 → cuerpo exacto `{"status":"ok","database":"connected"}` sin wrapper; 503 → `{"error":{"code":"database_unavailable","message":"La base de datos no está conectada","details":{"database":"disconnected"}}}`; `Cache-Control: no-store` en ambos; error inesperado del service → 500 `internal` genérico.
- **Criterio de terminado**: `go test ./internal/status/` en verde; la respuesta observada coincide **byte a byte** con los ejemplos de `contracts/openapi.yaml` y de `quickstart.md` §1–2.
- **Commit sugerido**: `feat(status): handler de /healthz con sobres de respuesta 200/503`

### T017 · Rutas del dominio y cableado en `cmd/api/main.go` · `[backend]`

- **Archivos**: `backend/internal/status/routes.go`, `backend/cmd/api/main.go`, `backend/cmd/api/main_test.go` (NUEVOS).
- **Qué hace**: `RegisterPublic(r httpserver.Registrar, h *Handler)` publica `GET /healthz` (sin autenticación; la superficie pública queda auditable de un vistazo). `main.go` = composición **manual** de dependencias (D12/D-A6, sin wire/fx): `config.Load` → `logger.New` → `database.NewPool` → `status.NewRepository → NewService → NewHandler` → `RegisterPublic` → `httpserver.New(mux, Options{…}, logger, middleware.RequestID, middleware.Recover, middleware.Logging, middleware.CORS)` con la cadena en el orden aprobado (T012) → `srv.Run(ctx)` con apagado ordenado. Se extrae una función `nuevaApp(...)`/`nuevoMux(...)` testeable para no probar `main` directamente. Sin estado global (R6).
- **Pruebas incluidas** (§III): prueba de humo con `httptest.NewServer` sobre el mux ensamblado: `GET /healthz` responde con el sobre (200 o 503 según el repo falso) y la cadena de middlewares está montada (cabecera `X-Request-ID` presente); `go build ./cmd/api` compila.
- **Criterio de terminado**: `go test ./...` en verde; `go run ./cmd/api` levanta y responde `/healthz` en `localhost:8080`; el proceso se apaga ordenadamente con SIGTERM (log de cierre presente).
- **Commit sugerido**: `feat(api): cablear cmd/api/main.go y publicar GET /healthz`

### T018 · Suite del sobre de respuestas (SC-008 / SC-009) · `[backend]`

- **Archivos**: `backend/internal/platform/httpserver/envelope_test.go` (NUEVO; suite de aceptación del sobre).
- **Qué hace**: suite sobre el **stack completo** (router `Registrar` + cadena de middlewares + handlers, incluidos los reales de `status`) que verifica la cobertura del 100 % del formato uniforme (D13, SC-008) y la ausencia de filtraciones (SC-009). Provoca: éxito; error previsto; **error inesperado** (handler de prueba que devuelve un error interno con datos sensibles tipo `sql: conexión a db-interna falló`) y **`panic`** (recuperado por `recover`). *Limitación conocida (R12)*: en F1 no pueden provocarse errores inesperados contra la API viva porque `/healthz` no tiene entrada de usuario; la provocación es vía handlers de prueba sobre el mismo stack.
- **Pruebas incluidas** (§III) — son el objeto de la tarea: (1) 200 → DTO directo; (2) 503 → sobre `database_unavailable` con `details`; (3) ruta no documentada → 404 sobre `not_found` (nunca texto de la stdlib); (4) `POST /healthz` → 405 sobre `method_not_allowed`; (5) error inesperado → 500 `{"error":{"code":"internal","message":"Error interno del servidor"}}` y el cuerpo **no** contiene el detalle (`"sql"`, `"db-interna"`, trazas, nombres de archivo, datos de infraestructura); (6) `panic` → 500 con sobre, proceso vivo; (7) el log capturado (helpers de T013) contiene el detalle interno **y** el `request_id`, que coincide con la `X-Request-ID` de la respuesta.
- **Criterio de terminado**: `go test ./internal/platform/... ./internal/status/...` en verde; SC-008 y SC-009 quedan **evidenciados** con estas pruebas (quedan como referencia de F2–F9: toda nueva operación debe pasar la misma regla).
- **Commit sugerido**: `test(httpserver): suite del sobre de respuestas (SC-008, SC-009)`

---

## Fase 5 — Imagen Docker del backend

### T019 · `backend/Dockerfile` multi-stage · `[infra]`

- **Archivos**: `backend/Dockerfile` (NUEVO), `backend/.dockerignore` (NUEVO).
- **Qué hace**: build multi-stage: `FROM golang:1.27` (D5) como etapa de compilación (`go mod download` con `go.mod`/`go.sum` copiados antes del resto para cachear, `CGO_ENABLED=0 go build ./cmd/api`) → imagen **mínima** de ejecución (alpine mínima o distroless; la elección concreta la justifica `devops` en el PR) con usuario no-root, `EXPOSE 8080` y solo el binario + certificados. `.dockerignore` evita copiar `.git`, `node_modules`, etc.
- **Pruebas incluidas** (§III): — (build de imagen). Verificación ejecutable: `docker build -f backend/Dockerfile backend` en verde; `docker run` + `curl localhost:8080/healthz` responde; la imagen final **no** contiene toolchain de Go ni fuentes (`docker run … ls` / inspección del tamaño).
- **Criterio de terminado**: la imagen arranca, sirve `/healthz` y su tamaño es de imagen mínima (no de imagen de build); `golang:1.27` es el único `FROM` de compilación.
- **Commit sugerido**: `build(backend): Dockerfile multi-stage sobre golang:1.27`

---

## Fase 6 — Frontend

### T020 · Scaffold: Vite + React + TypeScript `strict` · `[frontend]` `[P5]`

- **Archivos**: `frontend/package.json`, `frontend/package-lock.json`, `frontend/vite.config.ts`, `frontend/tsconfig.json`, `frontend/tailwind.config.*`, configuración de ESLint + Prettier, `frontend/src/main.tsx`, `frontend/src/index.css` (Tailwind), carpetas `frontend/src/{app,api,components,features,lib}/` (NUEVOS).
- **Qué hace**: monta el proyecto según la skill `react-frontend`: React 19, TypeScript con `strict: true` (prohibido `any`), Vite, Tailwind CSS, Vitest + Testing Library + MSW, y los scripts `dev`, `build`, `lint`, `typecheck`, `test`, **`api:gen`** (`openapi-typescript ../backend/api/openapi.yaml --output src/api/schema.d.ts`) y `e2e` (consumido por `make e2e`, T003). Dependencias justificadas en el PR (D22): React, React Router, TanStack Query, Tailwind, Vitest, Testing Library, MSW, Playwright (T027), `openapi-typescript` (herramienta de desarrollo, D-A8). **No** se instala shadcn/ui (diferido a F3, D14).
- **Pruebas incluidas** (§III): prueba de humo con Vitest + Testing Library (render de la app en su estado inicial) y `npm run lint` / `npm run typecheck` / `npm test -- --run` en verde.
- **Criterio de terminado**: `npm ci && npm run lint && npm run typecheck && npm test -- --run && npm run build` pasan desde `frontend/`; el árbol de carpetas coincide con el de `plan.md` (Project Structure).
- **Commit sugerido**: `feat(frontend): scaffold Vite + React + TypeScript strict con Tailwind y pruebas`

### T021 · Router, providers y layout · `[frontend]` `[P6]`

- **Archivos**: `frontend/src/app/router.tsx`, `frontend/src/app/providers.tsx`, `frontend/src/app/layout.tsx` + tests (NUEVOS).
- **Qué hace**: React Router con la ruta inicial `/` (destino: `StatusPage`, T024), `QueryClientProvider` de TanStack Query como único proveedor de datos del servidor (D14) y un layout mínimo con `main`/`nav` semánticos y foco visible (accesibilidad de la skill). Los datos se consultan **al montar + `refetch` a demanda** (sin auto-refresco en F1).
- **Pruebas incluidas** (§III): Vitest + Testing Library: la ruta `/` renderiza el outlet esperado; los providers envuelven la app (una consulta de prueba pasa por el `QueryClient`); markup semántico accesible (`getByRole`).
- **Criterio de terminado**: `npm test -- --run` y `npm run typecheck` en verde; ningún componente hace `fetch` directo (regla de oro de la skill; lo revisa `revisor-codigo`).
- **Commit sugerido**: `feat(frontend): router, providers y layout de la aplicación`

### T022 · Tipos generados desde el contrato · `[frontend]` `[P6]`

- **Archivos**: `frontend/src/api/schema.d.ts` (GENERADO y commiteado).
- **Qué hace**: ejecuta `npm run api:gen` (`openapi-typescript`) contra **`backend/api/openapi.yaml`** (nunca contra el snapshot de `specs/`, D15) y commitea el resultado. Establece la regla anti-deriva (R4/R20): quien cambia el contrato regenera `schema.d.ts` en el mismo PR.
- **Pruebas incluidas** (§III): — (artefacto generado). Verificación: regenerar (`npm run api:gen`) no produce `git diff`; `npm run typecheck` pasa con los tipos generados (`SystemStatus`, `ErrorEnvelope`, `ErrorBody`, respuestas 200/404/405/500/503).
- **Criterio de terminado**: `schema.d.ts` commiteado, reproducible desde el contrato y usado como única fuente de tipos de la API en el frontend.
- **Commit sugerido**: `build(frontend): generar tipos de la API desde backend/api/openapi.yaml`

### T023 · Cliente HTTP único: `apiFetch<T>` + `ApiError` · `[frontend]`

- **Archivos**: `frontend/src/api/client.ts`, `frontend/src/api/status.ts`, `frontend/src/api/client.test.ts` (NUEVOS).
- **Qué hace**: el **único** mecanismo de manejo de respuestas (D13, FR-012): `apiFetch<T>()` resuelve 2xx con el DTO directo (tipado con `schema.d.ts`) y convierte todo 4xx/5xx en `ApiError` con el `ErrorEnvelope`; timeout de **5 s** con `AbortController` (D20); error de red, timeout, JSON inválido o respuesta que no cumple el sobre → error de tipo *inaccesible*. `status.ts` expone `getSystemStatus()` (único lugar que consulta el estado). URL base desde `import.meta.env.VITE_API_URL`.
- **Pruebas incluidas** (§III): Vitest + MSW: 200 → DTO `SystemStatus`; 503 → `ApiError` con `code:"database_unavailable"` y `details.database`; 500 → `ApiError` `internal`; 404 sobre → `not_found`; respuesta malformada y timeout de 5 s → error *inaccesible*; el `ApiError` nunca pierde `code`/`message`/`details`.
- **Criterio de terminado**: `npm test -- --run` en verde; no existe ningún otro `fetch` en el código fuente (búsqueda de `fetch(` fuera de `client.ts` = 0 coincidencias).
- **Commit sugerido**: `feat(frontend): apiFetch con manejo único del sobre y timeout de 5 s`

### T024 · Feature `status`: página, hook e indicadores · `[frontend]`

- **Archivos**: `frontend/src/features/status/pages/StatusPage.tsx`, `frontend/src/features/status/hooks/useSystemStatus.ts`, `frontend/src/features/status/components/` (indicadores de estado), `frontend/src/features/status/status.test.tsx` (NUEVOS).
- **Qué hace**: la página inicial («PaginaEstado» de `ux.md`): consulta el estado al montar vía `useSystemStatus` (TanStack Query sobre `getSystemStatus`, con botón **«Volver a consultar el estado»** que dispara `refetch`) y presenta los estados de forma comprensible (FR-005, SC-005). **Mapeo confirmado** (plan §"Ajustes pendientes en ux.md" 1): `200` + `SystemStatus` → *conectado*; `503` + `error.code = "database_unavailable"` → *bd-no-conectada* (error A, texto de `ux.md` §6); timeout/red/respuesta malformada → *inaccesible* (error B, «No se pudo consultar el estado del sistema»). `error.message` del sobre puede usarse como explicación de apoyo del error A; los literales de `ux.md` mandan si difieren. Estados cubiertos: cargando, conectado, error A, error B.
- **Pruebas incluidas** (§III): Vitest + Testing Library + MSW: los cuatro estados visibles (cargando → conectado; 503 → error A; red caída/timeout → error B distinguido de error A); el botón refresca y refleja el estado **actual** (US2 esc. 4 con MSW cambiando la respuesta); accesibilidad básica (`getByRole("button")`, `getByRole("status")` o equivalente).
- **Criterio de terminado**: `npm test -- --run` y `npm run typecheck` en verde; la página muestra el estado sin acciones adicionales en <3 s (SC-005, se comprueba en e2e T027 y manualmente en `quickstart.md` §4); los textos respetan `ux.md`.
- **Commit sugerido**: `feat(frontend): feature status con página, hook y estados de la interfaz`

### T025 · `frontend/nginx.conf` y `frontend/Dockerfile` · `[infra]`

- **Archivos**: `frontend/nginx.conf` (NUEVO), `frontend/Dockerfile` (NUEVO), `frontend/.dockerignore` (NUEVO).
- **Qué hace**: imagen multi-stage `FROM node:22` (versión que fija el `ci.yml` del kit — R11) con `npm ci && npm run build` → `FROM nginx:alpine` sirviendo `dist/` con **SPA fallback** (`try_files $uri /index.html`) y cabeceras de caché razonables para `index.html` (sin caché) y assets con hash. `VITE_API_URL` se hornea en build time vía `ARG` (R6, por defecto `http://localhost:8080`): documentado en el README (T028).
- **Pruebas incluidas** (§III): — (build de imagen). Verificación ejecutable: `docker build ./frontend` en verde; servir la imagen y cargar `/` y una ruta profunda de la SPA (fallback a `index.html`) en el navegador o con `curl`.
- **Criterio de terminado**: la imagen sirve la SPA con fallback y sin toolchain de Node en la etapa final; el valor de `VITE_API_URL` por defecto apunta al backend local.
- **Commit sugerido**: `build(frontend): Dockerfile nginx con SPA fallback`

---

## Fase 7 — Docker Compose y pruebas end-to-end

### T026 · `docker-compose.yml` completo · `[infra]`

- **Archivos**: `docker-compose.yml` (EDITABLE: no está en `.kit-manifest.json`).
- **Qué hace**: descomenta y configura los servicios `backend` y `frontend` (D2/D3): **sin `env_file` obligatorio** (el entorno local funciona en clon limpio con los valores por defecto, FR-001); puertos **parametrizados** `${DB_PORT:-5432}:5432`, `${HTTP_PORT:-8080}:8080`, `${WEB_PORT:-5173}:80`; `DATABASE_URL` del contenedor backend **construido desde `POSTGRES_*` apuntando al host `db`** (nunca el `localhost` del host — R9); `depends_on: db: condition: service_healthy` y `frontend` depende de `backend` (R7); volúmenes y healthcheck del `db` del kit intactos. **Imagen de PostgreSQL**: queda el punto de decisión R10 (T034) — si el humano ya decidió `postgres:16-alpine`, se aplica aquí; si no, se mantiene `postgres:16.4-alpine` con la nota de "pendiente de actualización" (US7 esc. 2).
- **Pruebas incluidas** (§III): — (orquestación). Verificación ejecutable: `make up` desde un clon limpio **sin `.env`** levanta los tres servicios (SC-001); `docker compose ps` muestra `db` healthy + `backend` + `frontend`; `curl -i http://localhost:8080/healthz` y `http://localhost:5173` responden; edge case de puerto ocupado → fallo identificable y `DB_PORT=5433 HTTP_PORT=8081 WEB_PORT=5174 make up` funciona; `make down && make up` es repetible sin estado residual (US1 esc. 2).
- **Criterio de terminado**: los tres servicios levantan con un solo comando desde un clon limpio (FR-001, SC-001 < 15 min medidos con `quickstart.md` §0); ningún archivo del kit fue modificado.
- **Commit sugerido**: `build(compose): servicios backend y frontend con puertos parametrizados`

### T027 · Pruebas end-to-end con Playwright · `[frontend]`

- **Archivos**: `frontend/e2e/playwright.config.ts`, `frontend/e2e/status.spec.ts`, `frontend/package.json` (devDependency `@playwright/test`) (NUEVOS).
- **Qué hace**: configura Playwright para ejecución **local** (D17; el `ci.yml` del kit no se edita) con `baseURL` `http://localhost:5173` y el e2e del flujo de estado: abrir la página inicial y verificar que muestra el estado del sistema (conectado) sin acciones adicionales y en <3 s (SC-005). Los escenarios de error A/B se validan manualmente con `quickstart.md` §4 (detener `db` / `backend`): se documentan como pasos manuales del quickstart, no como e2e que manipule contenedores.
- **Pruebas incluidas** (§III): la propia prueba e2e (`e2e/status.spec.ts`) más su ejecución: `make up && make e2e` en verde.
- **Criterio de terminado**: `make e2e` pasa contra el stack levantado; el flujo crítico de F1 queda cubierto para heredar en F2–F9. *(R2: el CI no corre e2e; propuesta al kit como nota para el humano, ver sección final.)*
- **Commit sugerido**: `test(e2e): flujo de estado de la página inicial con Playwright`

---

## Fase 8 — Documentación

### T028 · `README.md` del proyecto · `[docs]`

- **Archivos**: `README.md` (NUEVO, raíz).
- **Qué hace**: la documentación para levantar el entorno sin ayuda (FR-009, D18): qué es la F1; prerequisito único (Docker); **quickstart** (`make up`, puertos 5173/8080/5432 y cómo cambiarlos con `DB_PORT`/`HTTP_PORT`/`WEB_PORT`, `.env` opcional); comandos (`make up/down/test/lint/security/ci/db-migrate/doctor/e2e/api-gen/sqlc-gen/sqlc-verify`, `make instalar-hooks` una vez por clon — R3); herramientas de desarrollo opcionales con **versiones fijadas** (Go 1.27, Node 22, `golang-migrate`, `sqlc`, `openapi-typescript` — R4: generación reproducible); **tabla de versiones y soporte de seguridad** (FR-016, SC-010) con los pendientes ya conocidos: PostgreSQL `16.4` acumula CVEs corregidos en minors posteriores (R10) y Node 22 deja de recibir soporte en abril de 2027 (R11); notas operativas: `VITE_API_URL` se hornea en build (R6) y regeneración de artefactos generados (`make api-gen` / `make sqlc-gen`, regla de revisión R4). Enlaces a `specs/001-estructura-base/quickstart.md`, `docs/tecnico/arquitectura.md` (§8 = receta) y `docs/GUIA-INICIO.md` (kit).
- **Pruebas incluidas** (§III): — (documento). Verificación: una persona nueva sigue el README/quickstart y levanta el entorno en <15 min (SC-001; parte de la validación de cierre, `quickstart.md` §0).
- **Criterio de terminado**: el quickstart es ejecutable de punta a punta siguiendo **solo** el README; toda variable de entorno usada por el código aparece documentada.
- **Commit sugerido**: `docs(readme): quickstart, puertos, comandos y versiones del proyecto`

### T029 · Checklist de verificación de la receta (SC-007) · `[docs]` `[P7]`

- **Archivos**: `specs/001-estructura-base/checklists/receta.md` (NUEVO).
- **Qué hace**: plantilla de checklist alineada con `quickstart.md` §9 y `docs/tecnico/arquitectura.md` §8: casillas para cada paso (rama creada, 10 pasos de la receta, artefactos regenerados, `make ci` en verde, `git diff main --stat` sin cambios en áreas existentes, verificación funcional de `GET /api/v1/muestra` e integridad de `/healthz`, decisión de descarte de la rama) con campos de evidencia (quién verifica y fecha, salida adjunta de `make ci`, diff, resultado funcional).
- **Pruebas incluidas** (§III): — (plantilla). Verificación: la plantilla cubre literalmente los escenarios de US5 y los criterios de SC-007/FR-014/FR-015.
- **Criterio de terminado**: la plantilla existe, está lista para rellenar y no contiene evidencias precargadas.
- **Commit sugerido**: `docs(checklist): plantilla de verificación de la receta (SC-007)`

---

## Fase 9 — Verificación de la receta y cierre

> El CHANGELOG, el PR de entrega y las notas para el cliente los cierra el `documentador` en la fase de entrega del orquestador (no son tareas de este documento). La validación técnica (`qa-tester`, `revisor-codigo`, `seguridad` en paralelo) se ejecuta sobre cada tarea y sobre el conjunto antes del cierre.

### T030 · **[humano]** Verificación de la receta en rama descartable (SC-007) · `[humano]`

- **Archivos**: ninguno en `main`. Rama aparte `practica/receta-001` (**descartable**); evidencias en `specs/001-estructura-base/checklists/receta.md` (rellenada).
- **Qué hace**: una persona que **no** participó en la creación de la receta (SC-007, R13 — si no hay otra persona disponible, escala al humano) sigue **solo** `docs/tecnico/arquitectura.md` §8 (10 pasos) y `quickstart.md` §9 para agregar un área de práctica **pública y de solo lectura** (ejemplo sugerido: dominio `muestra`, migración `000002_create_sample_items`, `internal/db/queries/muestra.sql`, `internal/muestra/{model,repository,service,handler,routes}.go`, `RegisterPublic` + una línea de cableado en `main.go`; `make sqlc-gen` regenera `internal/db/`). **Límites declarados** (plan D21): sin entradas de usuario (así no arrastra `validate`, `rate-limit` ni `CSRF`, diferidos) y sin rutas de panel (dependen de `authn`/`authz`, que llegan en F2). Al terminar: `make ci` en verde, `git diff main --stat` muestra **solo archivos nuevos** del área + la línea de cableado (FR-015), `GET /api/v1/muestra` responde con el sobre de éxito y `/healthz` sigue intacto; se rellena la checklist con las evidencias. Finalmente **se descarta la rama** (`git checkout main && git branch -D practica/receta-001`): la tabla de práctica **nunca** llega a `main`, que cierra F1 con 0 tablas de negocio.
- **Pruebas incluidas** (§III): las del propio ejercicio (paso 10 de la receta: service/handler/repository con sus pruebas) — **pero viven solo en la rama descartada**; la evidencia conservada es la checklist.
- **Criterio de terminado**: checklist de SC-007 rellena y firmada; `main` sin tablas de negocio ni código de práctica; SC-007 queda **registrada** como cumplida antes del cierre de F1.
- **Commit sugerido**: *(en la rama de práctica)* `feat(muestra): ejercicio de práctica de la receta de áreas de negocio` — **se descarta con la rama**.

### T031 · **[humano]** Crear el repositorio remoto y empujar · `[humano]` *(bloqueante R1)*

- **Archivos**: ninguno (acción en GitHub).
- **Qué hace**: crear el repositorio remoto en GitHub y empujar las ramas existentes. Sin remoto, el workflow `CI` del kit no corre en ningún sitio y FR-006/FR-007 no pueden verificarse.
- **Pruebas incluidas** (§III): — . Verificación: el repositorio es accesible y las ramas están en el remoto.
- **Criterio de terminado**: remoto configurado en este clon (`git remote -v`) y ramas empujadas. **Nota**: el path definitivo del módulo Go (`backend/go.mod`, hoy `simiente-santa/backend`) se fija cuando exista el remoto (R1, "Preguntas abiertas" 6 de `decisiones.md`).
- **Commit sugerido**: — (sin commit)

### T032 · **[humano]** Renombrar la rama por defecto `master` → `main` · `[humano]` *(bloqueante R1)*

- **Archivos**: ninguno (acción en GitHub).
- **Qué hace**: renombrar la rama por defecto del repositorio a `main`. El `ci.yml` del kit solo se dispara en push a `main` (y en PR): con `master` como rama por defecto, el CI no corre (riesgo R1). Alternativa (si no se renombra): proponer el ajuste al repositorio del kit — pero renombrar es la opción recomendada.
- **Pruebas incluidas** (§III): — . Verificación: al abrir un PR contra la rama por defecto, el workflow `CI` se dispara.
- **Criterio de terminado**: la rama por defecto es `main` y un push a ella dispara el CI.
- **Commit sugerido**: — (sin commit)

### T033 · **[humano]** Activar la protección de la rama principal · `[humano]` *(bloqueante R1)*

- **Archivos**: ninguno (acción en GitHub).
- **Qué hace**: activar branch protection sobre `main` con los checks del workflow `CI` **obligatorios** (jobs: `agentes`, `controles`, `backend`, `frontend`, `secretos` según detecte el kit) y, al menos, "requiere PR antes de integrar". Es la garantía real de FR-007/SC-003 (los hooks locales son solo red de seguridad temprana, R3).
- **Pruebas incluidas** (§III): — . Verificación: un PR con una prueba rota a propósito queda marcado **no apto** y no puede integrarse; uno limpio queda **apto** en <10 min (SC-003, SC-004 — `quickstart.md` §6).
- **Criterio de terminado**: protección activa y verificados los dos escenarios (apto / no apto) con PRs reales.
- **Commit sugerido**: — (sin commit)

### T034 · **[humano]** Decidir la imagen de PostgreSQL (R10) · `[humano]` *(bloqueante para cerrar SC-010)*

- **Archivos**: `docker-compose.yml` (solo si la decisión cambia la imagen — mínima tarea de seguimiento `[infra]`).
- **Qué hace**: decidir entre mantener `postgres:16.4-alpine` (reproducible pero con CVEs corregidos en minors posteriores) y pasar a `postgres:16-alpine` (rama 16 en soporte hasta noviembre de 2028, con parches de seguridad). La recomendación del plan (R10) es `postgres:16-alpine` para el entorno de desarrollo, separando reproducibilidad del build de la base de datos de desarrollo.
- **Pruebas incluidas** (§III): — . Verificación: `docker compose images` refleja la imagen decidida y la tabla de versiones del README (T028) queda consistente (US7 esc. 2: lo que no esté en su última corrección queda identificado como pendiente de actualización).
- **Criterio de terminado**: decisión **registrada** (en el PR/registro de decisiones) y `docker-compose.yml` + README coinciden con ella. **Secuenciación**: idealmente decidir **antes** de T026 para no reabrir el compose.
- **Commit sugerido**: *(si cambia la imagen)* `build(compose): usar postgres:16-alpine en el entorno de desarrollo`

### T035 · **[humano]** Registrar el vencimiento de Node 22 y proponer la actualización al kit (R11) · `[humano]`

- **Archivos**: ninguno directo (la constancia vive en el README, T028; la propuesta es una issue/comunicación al repositorio del kit).
- **Qué hace**: registrar que Node 22 (fijado por el `ci.yml` del kit, **no editable** aquí — regla 10) deja de recibir soporte de mantenimiento en **abril de 2027**, y proponer al repositorio del kit su actualización a una versión LTS vigente con antelación. Cumple FR-016/SC-010: lo que vaya a perder soporte queda identificado como pendiente de actualización.
- **Pruebas incluidas** (§III): — . Verificación: la tabla de versiones del README contiene el vencimiento con fecha y el estado "propuesto al kit".
- **Criterio de terminado**: constancia registrada y propuesta emitida al repo del kit.
- **Commit sugerido**: — (sin commit; el registro va en T028)

---

## Lo que NO se hace en F1 (diferidos de D23 — no implementar "de paso")

| Diferido | Punto de decisión |
|---|---|
| `platform/validate` (validación de DTOs con detalle por campo) | **F2** (primeras entradas de usuario) |
| Sesiones / `authn` / `authz` / permisos por módulo (propuesta D-A7: cookie `httpOnly` + sesión en servidor, bcrypt/argon2id — **pendiente de confirmación del humano**) | **F2** |
| `CSRF` y `rate-limit` | Con el **primer endpoint público escribible** (F2) |
| `platform/paginate` (listados paginados) | **F2** (primer listado del panel) |
| `platform/migrate` (runner embebido `//go:embed`, opt-in `RUN_MIGRATIONS`) | Cuando un **entorno desplegado** necesite auto-migrar sin CLI |
| i18n (español/inglés) | **F3** (primer sitio público bilingüe) |
| Métricas y trazas (Prometheus/OpenTelemetry) | El **primer despliegue operado** |
| Colas de trabajo (procesamiento asíncrono) | Primera necesidad real (candidata: F8/F9) |
| Almacenamiento de imágenes / subida de archivos | **F8** |
| Generación de handlers desde el contrato (codegen servidor) | Revisión **post-MVP** |
| shadcn/ui y diseño visual de la página | **F3** (F1 es meramente funcional) |
| Auto-refresco / sondeo del estado (`refetchInterval`) | F3+ si el cliente lo pide (F1: consulta al montar + botón) |
| Router chi (o cualquier otro) | Opción de **F2** (D-A4); hoy `net/http` tras `Registrar` |
| Tablas de negocio (usuarios, sesiones, roles, contenido…) | **F2** en adelante (una tabla por funcionalidad, `data-model.md`) |
| Wrapper `{"data":…}` en respuestas de éxito | **Descartado** por confirmación humana del 2026-09-30: sobre de éxito = DTO directo |
| Despliegue a producción u otros entornos | Fuera de alcance de F1 (spec, *Out of Scope*) |

## Regla 10: archivos del kit — NO se editan

`Makefile`, `.github/workflows/ci.yml`, `.githooks/`, `AGENTS.md`, `docs/GUIA-INICIO.md`, `equipo/` y el resto de `.kit-manifest.json` **no se tocan**. Los targets propios van en `proyecto.mk` (T003); `docker-compose.yml` y `.env.example` **sí** son editables (no están en el manifiesto del kit).

**Nota para el humano — propuesta de mejora al repositorio del kit (no es tarea de implementación aquí)**: (a) job de **e2e** (Playwright) en el `ci.yml`; (b) chequeo del **umbral de cobertura ≥80 %** en `service/`; (c) paso de CI que verifique la **ausencia de deriva** de artefactos generados (`make sqlc-verify`, `npm run api:gen` sin diff); (d) actualización de **Node 22** a una LTS vigente antes de abril de 2027 (R11, T035). Cualquier cambio en el kit se propone a su repositorio y llega por `make actualizar-kit`.

## Notas y riesgos de secuenciación

1. **Orden docker/compose**: el orden sugerido era «… status → docker/compose → frontend …», pero `docker-compose.yml` (T026) referencia `frontend/Dockerfile` (T025): la fase de compose queda **después** del frontend, y el Dockerfile del backend (T019) se ejecuta antes, tras el backend. Sin esto, `make up` no puede levantar los tres servicios en un solo comando (FR-001).
2. **R10 antes de T026**: la decisión de imagen de PostgreSQL (T034) debería cerrarse antes de T026; si se decide después, hay que reabrir `docker-compose.yml` (diff mínimo, pero evitable).
3. **R5 (sqlc sin consultas)**: el comportamiento del CLI con `internal/db/queries/` vacío es incierto por definición del riesgo; T006 lo deja **documentado** y `sqlc-verify` debe pasar en cualquiera de los dos desenlaces (genera artefactos vacíos / exige al menos una consulta).
4. **`go mod tidy` diferido**: T002 no puede tidyar (eliminaría `pgx` sin importar nada); la consolidación ocurre en T010. No ejecutar `go mod tidy` entre T002 y T010.
5. **`make api-gen` y `make e2e` definidos antes que su contenido** (T003 vs. T022/T027): los targets existen pero fallan hasta que llega su fase; aceptado y documentado en T003.
6. **SC-007 exige una persona ajena a la receta** (R13): si no hay otra persona disponible, la verificación (T030) escala al humano con validación explícita; **F1 no se cierra sin el registro de SC-007**.
7. **R12 / SC-009**: los errores inesperados no pueden provocarse contra la API viva en F1 (una sola operación sin entrada); la suite de T018 los provoca sobre el stack completo y la limitación queda documentada. Se refuerza con las primeras operaciones con entrada en F2.
8. **`ux.md` se está ajustando en paralelo** (agente `disenador-ux`): T024 usa el mapeo de estados ya fijado en `plan.md`; si `ux.md` cambia, solo se actualizan literales de T024 (el mapeo y los tipos no cambian).
9. **T031–T033 (remoto, rama `main`, protección)** son bloqueantes para SC-003/SC-004 y para que el CI corra; pueden ejecutarse en cuanto el código pase `make ci`, pero **deben estar completas antes del PR de cierre**.
10. **Deriva de artefactos generados** (R4/R20): un PR que toque `migrations/` o `internal/db/queries/` debe tocar `internal/db/`; uno que toque `backend/api/openapi.yaml` debe tocar `frontend/src/api/schema.d.ts`. Es criterio de revisión de `revisor-codigo`.
