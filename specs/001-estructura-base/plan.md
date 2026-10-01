# Implementation Plan: Estructura base del proyecto (F1)

**Branch**: `001-estructura-base` | **Date**: 2026-09-30 (reescrito tras la ampliación de alcance) | **Spec**: [spec.md](./spec.md)

**Input**: especificación aprobada el 2026-09-30 (`spec.md`, alcance ampliado con US4–US7 y FR-010…FR-016) · arquitectura y decisiones ya aprobadas (`docs/tecnico/arquitectura.md`, `docs/tecnico/decisiones.md`) · constitución (`.specify/memory/constitution.md`) · skills `go-backend`, `postgres-db`, `react-frontend` · kit del proyecto (Makefile, `docker-compose.yml`, `.github/workflows/ci.yml`, `.env.example`).

## Cambios respecto de la versión anterior del plan (2026-09-30)

**Motivo**: el humano amplió el alcance de F1 (decisión del 2026-09-30, reflejada en `spec.md`) y aprobó la arquitectura de `docs/tecnico/`. Este plan **aterriza** esa arquitectura: la cita, detalla los archivos que se crean y no reabre ninguna de sus decisiones. El delta para revisar es:

1. **Plataforma interna (US4, FR-010/FR-011)** — nuevo **D11/D12**: `internal/platform/{config,logger,apperr,httpserver,middleware,database,testutil}` con lo que F1 necesita (`testutil`, y no `testing`, para no chocar con el paquete `testing` de la stdlib en cada `_test.go`); se **difiere** lo demás (`validate` a F2, `authn`/`authz`/`CSRF` a F2, `paginate` a F2, `migrate` embebido cuando un entorno desplegado lo pida, i18n a F3). Reglas de dependencia `cmd → dominio → platform` (platform no conoce dominios; un dominio no importa a otro).
2. **Receta de áreas de negocio (US5, FR-014/FR-015)** — nuevo **D21**: la receta vive en `docs/tecnico/arquitectura.md` §8 (10 pasos) y F1 la **verifica** con (a) el dominio `status`, primer área real construida sobre ella, y (b) el ejercicio de práctica de `quickstart.md` §9 (SC-007). Este plan no duplica la receta: define cómo se comprueba.
3. **Formato uniforme de respuestas (US6, FR-012/FR-013)** — nuevo **D13**: sobres de éxito y de error definidos una sola vez (`httpserver.WriteJSON` / `httpserver.WriteError`), `ErrorEnvelope` `{"error":{"code","message","details"?}}`, registro de códigos, fallback 404/405 del router también en sobre de error, y el detalle interno **nunca** sale al cliente (va al log con `request_id`). Reflejado en `contracts/openapi.yaml`.
4. **Versiones con soporte de seguridad vigente (US7, FR-016)** — nuevo **D5** (Go **1.27**, que sustituye al "Go 1.23+" del plan anterior: 1.23 está en fin de vida desde 2025-08-12) y tabla de versiones con su estado de soporte (SC-010) en la sección "Métricas del plan".
5. **Capa de datos fijada: sqlc** — nuevo **D6** (antes diferido sin decidir): sqlc lee `backend/migrations/` como esquema, SQL explícito en `backend/internal/db/queries/*.sql`, `sqlc.yaml` en `backend/`, código generado commiteado, CLI como herramienta de desarrollo; stored procedures solo como excepción justificada dentro del `repository`. F1 solo hace el `Ping` de salud con `pgx` (D-A3): el primer código generado llega con la primera consulta de negocio, que es la del ejercicio de práctica.
6. **Router neutralizado** — **D4** reescrito: los dominios dependen de la interfaz `httpserver.Registrar` (`Handle`, `Group`) y del tipo `Middleware`; F1 implementa el adaptador sobre `net/http` (patrones de Go 1.22+). Chi queda como opción de F2, sin impacto en los dominios.
7. **Sin cambios de fondo**: reutilización del kit (D1–D3), contrato único `backend/api/openapi.yaml` (D9), migraciones versionadas con `golang-migrate` (D10), stack del frontend (D14–D16), Playwright local (D17), README (D18), `proyecto.mk` (D19), timeout de 5 s del fetch (D20).
8. **Diferidos con su punto de decisión (D23)**: `validate` (F2), sesiones/authn/authz/permisos (F2 — la propuesta cookie `httpOnly` + sesión en servidor está **pendiente de confirmación del humano**, no decidida), `CSRF`/`rate-limit` (con el primer endpoint público escribible), `paginate` (F2), i18n (F3), métricas/trazas (primer despliegue operado), colas (primera necesidad real de trabajo diferido), almacenamiento de imágenes (F8), generación de handlers desde el contrato (revisión post-MVP).

## Summary

F1 monta el esqueleto técnico que copiarán F2–F9: backend en Go 1.27 con `GET /healthz` que informa en cada consulta el estado real de la conexión a PostgreSQL, frontend en React que muestra ese estado, todo levantable con `make up` desde un clon limpio, validación automática de cada cambio con el pipeline del kit, y —lo nuevo de esta versión— una **plataforma interna** que estandariza lo transversal (configuración, logs, errores, HTTP, acceso a datos, salud) y una **receta verificable** para agregar áreas de negocio nuevas como unidades aisladas. Todo pasa por un **formato de respuesta uniforme** (sobres de éxito y de error) que no expone información interna. El enfoque sigue siendo **reutilizar al máximo el kit instalado** y agregar solo lo que el kit no trae: `backend/`, `frontend/`, los Dockerfiles, el contrato OpenAPI y el README del proyecto. El plan aterriza sin reabrirlas las decisiones de `docs/tecnico/decisiones.md` (D-A1…D-A9).

## Technical Context

**Language/Version**: **Go 1.27** (`go 1.27` en `backend/go.mod`, `golang:1.27` en `backend/Dockerfile`; D5) · TypeScript 5.x `strict` sobre Node 22 · React 19

**Primary Dependencies**: backend runtime → **solo `pgx/v5`** (D22); `sqlc`, `golang-migrate` y `openapi-typescript` son **herramientas de desarrollo** (no se enlazan; sus artefactos generados se commitean). frontend → React 19, Vite, React Router, TanStack Query, Tailwind CSS, Vitest + Testing Library + MSW, Playwright (e2e local), `openapi-typescript` (tipos desde el contrato)

**Storage**: PostgreSQL 16.4-alpine (servicio `db` de `docker-compose.yml`; ver R10 sobre el minor elegido). **Sin tablas de negocio en F1** (ver `data-model.md`). Capa de datos fijada: **sqlc** sobre `backend/migrations/` (D6)

**Testing**: estrategia por capa (sección propia abajo): `go test` (unitarias) + `go test -tags=integration` (repositorio contra PostgreSQL real, `DATABASE_URL_TEST`) + suite del sobre de respuestas; Vitest + Testing Library + MSW (frontend); Playwright (e2e, ejecución local)

**Target Platform**: contenedores Docker en máquina local de desarrollo (Linux/macOS/WSL) + GitHub Actions (`ubuntu-latest`)

**Project Type**: web-service (monorepo `backend/` + `frontend/`), arquitectura por capas con plataforma interna (`cmd → dominio → platform`)

**Performance Goals**: `/healthz` responde en ≤2 s aunque la BD esté caída (ping con timeout de 2 s) · página inicial con el estado en <3 s desde la carga (SC-005) · veredicto del pipeline en <10 min (SC-004) · entorno levantado desde un clon limpio en <15 min (SC-001)

**Constraints**: stack fijo por constitución · un solo comando de arranque (`make up`) con Docker como único prerequisito (FR-001) · sin secretos en el repo (FR-008) · dependencias nuevas minimizadas y justificadas (§II, D22) · **archivos del kit no editables** (regla 10 de AGENTS.md; `.kit-manifest.json`) · reglas de dependencia `cmd → dominio → platform` (`docs/tecnico/arquitectura.md` §1.2)

**Scale/Scope**: esqueleto patrón para F2–F9: 1 endpoint, 1 dominio (`status`), plataforma interna de ~7 paquetes en F1, 1 página, 0 tablas de negocio, 1 migración baseline, 1 receta + 1 ejercicio de práctica

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Evaluación | Resultado |
|---|---|---|
| §I La spec manda | El plan implementa FR-001…FR-016 tal cual; cada requisito tiene sección asignada (ver "Cobertura de requisitos"); sin comportamiento inventado. Lo que queda fuera de F1 está dicho explícitamente en la spec (`Out of Scope`) y en D23. | ✅ |
| §II Arquitectura | Monorepo `backend/` + `frontend/` + `backend/migrations/`. Capas `handler → service → repository` sobre `internal/platform/` (D11), con las reglas de dependencia de `arquitectura.md` §1.2. API REST JSON documentada en `backend/api/openapi.yaml`, contrato escrito **antes** que el código (D9). Dependencias minimizadas y justificadas (D22, `research.md`). | ✅ |
| §III Pruebas | Toda tarea de implementación incluirá sus pruebas (se exigirá en `tasks.md`). Cobertura ≥80 % en `service/` (verificable con `go test -cover`; la verificación la hacen QA/revisor). Pruebas de repositorio con tag `integration` contra PostgreSQL real (servicio del CI del kit). Frontend con Vitest + Testing Library; e2e con Playwright (DEBERÍA constitucional → incluido, ejecución local). Suite propia del sobre de respuestas (SC-008/SC-009). | ✅ |
| §IV Seguridad | En F1 la única operación (`GET /healthz`) no tiene entrada de usuario → sin superficie de inyección; cuando haya SQL, solo consultas parametrizadas vía `pgx`/`sqlc` (D6). Los errores inesperados nunca exponen información interna (FR-013, D13). Secretos por variables de entorno; `.env.example` sin valores reales; hooks de git y gitleaks en CI (kit) lo vigilan. `govulncheck` y `npm audit` ya corren en el CI del kit. | ✅ |
| §V Calidad | `gofmt`, `go vet`, `golangci-lint` (targets `lint` del kit + CI). TS `strict`, sin `any`. Errores envueltos con `%w`; traducción a HTTP en un único punto (`WriteError`). Nombres en inglés dentro del código. | ✅ |
| §VI Base de datos | Migraciones versionadas (`NNNNNN_nombre.up.sql`/`.down.sql`) con `golang-migrate`; nunca se edita una aplicada (lo vigilan hooks y CI). F1 crea la baseline `000001` (no-op) que fija la convención (D10, `data-model.md`). | ✅ |
| §VII Observabilidad | Logs estructurados `log/slog` en JSON con logger por petición y `request_id` (D12). Endpoint `/healthz` (nombre exigido). Configuración por variables de entorno. Todo se levanta con `docker compose up` (vía `make up`). | ✅ |
| §VIII Gobierno | Este plan requiere aprobación humana antes de `/speckit.tasks` e implementación. Sin despliegue a producción en F1. La propuesta de sesiones de F2 (D-A7) está **pendiente de confirmación humana** y no se implementa en F1. | ✅ |

**Sin violaciones que justificar** → sección "Complexity Tracking" vacía.

## Decisiones técnicas (resumen; análisis completo en `research.md`)

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| D1 | Reutilizar el kit tal cual: `make up` como comando único, CI de `.github/workflows/ci.yml` sin tocar, hooks, `.env.example`, `doctor.sh` | El kit ya resuelve FR-001, FR-006, FR-007 y FR-008; duplicarlo violaría la regla 10 de AGENTS.md | Crear Makefile/CI propios del proyecto |
| D2 | `docker-compose.yml` **sí es editable** (no está en `.kit-manifest.json`): descomentar servicios `backend`/`frontend`, parametrizar puertos (`${VAR:-defecto}`) y **no** exigir `env_file` | Permite `make up` en clon limpio sin pasos manuales (FR-001 estricto) y resuelve el edge case de puertos ocupados | Exigir `cp .env.example .env` como paso previo obligatorio |
| D3 | El `DATABASE_URL` del contenedor backend se construye en compose desde `POSTGRES_*` apuntando al host `db` | Evita que un `DATABASE_URL=...@localhost:5432` válido en el host se filtre al contenedor, donde `localhost` no es la BD | Usar `env_file: .env` a secas |
| D4 | **Router neutralizado**: los dominios dependen de `httpserver.Registrar` (`Handle(method, path, h)`, `Group(prefix, mws...)`) y de `type Middleware = func(http.Handler) http.Handler`; F1 implementa el adaptador sobre `net/http` (patrones `GET /ruta` de Go 1.22+) | Los dominios no conocen el router: adoptar chi (u otro) después toca un solo archivo de `platform/httpserver` y una línea de `main.go` (D-A4) | `chi` desde ya (innecesario hoy; se evalúa en F2); Gin/Echo/Fiber (pesadas y/o con contexto propio incompatible con `context.Context`); acoplar los dominios a `net/http` (haría caro el cambio) |
| D5 | **Go 1.27** (`golang:1.27` en el Dockerfile, `go 1.27` en `go.mod`) | Go 1.23 está en fin de vida desde 2025-08-12; solo 1.26 y 1.27 reciben parches de seguridad (FR-016). El CI del kit lee `backend/go.mod`, así que la versión la decide el proyecto (D-A5) | Go 1.23 (sin soporte); Go 1.26 (ventana de soporte más corta); `golang:latest` (builds no reproducibles) |
| D6 | **Capa de datos: sqlc** — `sqlc.yaml` en `backend/` (`schema: migrations`, `queries: internal/db/queries`, `out: internal/db`); SQL explícito y parametrizado en `.sql`; código generado **commiteado**; el CLI es herramienta de desarrollo. SP/funciones SQL solo como excepción justificada dentro del `repository` | Tipos derivados del SQL real: un cambio de consulta o esquema que no acompañe **falla en compilación**; 0 entradas nuevas en `go.mod` (D-A3) | GORM (5+9 entradas, reflection, `AutoMigrate` choca con `golang-migrate`); ent (49 entradas); sqlx (3, sin generación); mappers a mano sobre `pgx` |
| D7 | Contrato `/healthz`: **200** con `SystemStatus` `{"status":"ok","database":"connected"}` · **503** con `ErrorEnvelope` `database_unavailable` y `details.database:"disconnected"` · `Cache-Control: no-store` en ambos | El 503 con cuerpo distingue "BD caída" (503 con JSON) de "backend caído" (sin respuesta) y mantiene la regla uniforme de que todo 4xx/5xx es sobre de error (FR-012, US6 esc. 2) | 503 con `SystemStatus` (rompería el sobre único y daría al cliente dos mecanismos); siempre 200 con un campo (pierde la señal estándar para monitores); `/healthz` + `/readyz` (la spec pide un endpoint) |
| D8 | El estado de la BD se consulta con `Ping` en **cada** petición, con `context.WithTimeout` de 2 s, sobre un pool perezoso de `pgx` | FR-003 (estado real, no memorizado) y FR-004 (responde aunque la BD esté caída) sin que una BD colgada bloquee el endpoint | Caché/sondeo en segundo plano (violaría FR-003) |
| D9 | Contrato fuente único: el diseño vive en `specs/001-estructura-base/contracts/openapi.yaml` (snapshot inmutable) y la primera tarea de implementación lo copia verbatim a `backend/api/openapi.yaml`, que desde entonces es el documento vivo | Cumple §II (contrato antes que el código, ubicación fija) sin dos copias vivas que diverjan | Mantener ambos sincronizados a mano (divergencia garantizada) |
| D10 | Migraciones con **`golang-migrate`** (el kit ya lo usa en CI), versionadas e inmutables; baseline `000001_baseline` no-op que fija la convención. El runner embebido (`platform/migrate`) se **difiere** hasta que un entorno desplegado necesite auto-migrar | Ejercita el toolchain desde F1 (target `db-migrate`, paso "Migraciones" del CI, control de inmutabilidad) sin inventar mecanismos nuevos | Carpeta vacía con `.gitkeep` (toolchain sin probar); `AutoMigrate` de GORM/Atlas (choca con `golang-migrate`); runner embebido ya (configuración muerta) |
| D11 | **Plataforma interna** `internal/platform/` (D-A2): F1 implementa `config`, `logger`, `apperr`, `httpserver`, `middleware` (request-id, recover, logging, CORS), `database`, `testutil` (helpers de prueba; se llama `testutil` y no `testing` para no chocar con la stdlib); **difiere** `validate` (F2), `paginate` (F2), `migrate` (D10), `authn`/`authz`/`CSRF`/`rate-limit` (F2), i18n (F3). Reglas: `cmd → dominio → platform`; platform no conoce dominios; un dominio no importa a otro | Lo transversal existe una sola vez (SC-006) y ninguna área lo reimplementa (FR-011); ~10 paquetes pequeños en vez de un framework | Framework de aplicación de terceros; módulo compartido externo (no hay segundo consumidor aún); plantillas copiadas por proyecto (divergen) |
| D12 | Config por variables de entorno con stdlib (`os.Getenv` + validación al arrancar, sin Viper); logs `log/slog` JSON con logger por petición; errores stdlib + `apperr` traducidos a HTTP **solo** en `httpserver.WriteError`; DI manual en `main.go` (sin wire/fx) | `go.mod` casi vacío, `main.go` legible como mapa de dependencias, fallos de configuración detectados al arrancar (D-A6) | Viper/envconfig (17 entradas / 1 dependencia para 20 líneas); zap/zerolog (`slog` está en la stdlib); `pkg/errors` (`%w` basta); wire/fx (runtime para un grafo de ~4 nodos) |
| D13 | **Formato uniforme de respuestas** (FR-012/FR-013): sobre de éxito = DTO documentado de la operación (objeto JSON, `camelCase`, `additionalProperties: false`) escrito con `WriteJSON`; sobre de error = `{"error":{"code","message","details"?}}` escrito **solo** con `WriteError`; el fallback 404/405 del router y el `recover` también escriben sobre de error; `internal` responde mensaje genérico y el detalle va al log con `request_id` | Quien consume la API implementa el manejo de respuestas **una sola vez** (`apiFetch<T>` + `ApiError`); SC-008 es verificable al 100 % y SC-009 no deja salir información interna | Wrapper `{"data":…}` (contradice `arquitectura.md` §5.1/§5.7 aprobada; **resuelto por confirmación humana del 2026-09-30: el sobre de éxito es el DTO directo**, sin wrapper — ver `research.md` R17); formato libre por operación (rompe FR-012); dejar los 404/405 de texto de la stdlib (rompen SC-008); exponer el detalle interno en la respuesta (rompe FR-013) |
| D14 | Frontend: React Router + TanStack Query (**consulta al montar + `refetch` a demanda** con el botón "Volver a consultar el estado"; sin auto-refresco en F1) + Tailwind CSS; **shadcn/ui se difiere a F3** | Router/Query/Tailwind son la convención de la skill y el patrón que copiarán F2–F9; la spec no exige refresco automático (SC-005 = ver el estado al abrir la página) y el diseño de `ux.md` es manual; shadcn/ui sin diseño definido sería peso muerto | Sondeo periódico (`refetchInterval`, nota de futuro para F3+ si el cliente lo pide); incluir shadcn/ui ya; fetch directo en el componente (prohibido por la skill); Redux/Zustand (no hay estado global de cliente) |
| D15 | Tipos de API generados con `openapi-typescript` desde `backend/api/openapi.yaml`, **commiteados** en `frontend/src/api/schema.d.ts` | El CI corre `npm ci && typecheck` sin pasos extra; se regeneran con `npm run api:gen` cuando cambia el contrato | Generarlos en CI (exigiría tocar el `ci.yml` del kit); escribirlos a mano (divergen del contrato) |
| D16 | CORS mínimo escrito a mano (una cabecera para un `GET` simple sin credenciales) | Evita una dependencia para un solo header; F2 (sesiones) reevaluará el middleware completo | Dependencia `rs/cors` desde ya; proxy `/api` en nginx |
| D17 | Playwright configurado con un e2e del flujo de estado, ejecución **local** (`make e2e`), no en CI | §III dice DEBERÍA para flujos críticos y monta la infraestructura que F2–F9 heredan; el `ci.yml` es del kit y no se edita (regla 10) | No montar Playwright (aplazar a F2); editar el `ci.yml` (prohibido) |
| D18 | `README.md` en la raíz (no existe; no es del kit) con quickstart, puertos y comandos | FR-009: documentación para levantar el entorno sin ayuda; `docs/GUIA-INICIO.md` es del kit y genérica | Documentar solo en `docs/` |
| D19 | Targets propios en `proyecto.mk` (incluido por el Makefile del kit): `e2e`, `api-gen`, `sqlc-gen`, `sqlc-verify` | El Makefile es del kit y no se edita; `proyecto.mk` es el mecanismo previsto para extensiones | Editar el Makefile (prohibido) |
| D20 | El fetch del frontend a `/healthz` usa `AbortController` con timeout de **5 s**; timeout, error de red o respuesta que no cumple el sobre mapean al estado `inaccesible` ("No se pudo consultar" de `ux.md`) | 5 s > 2 s del ping de BD (D8): una BD lenta o caída produce el 503 legible (error A de `ux.md`) antes de que el frontend se rinda | Timeout por defecto del navegador; timeout ≤ 2 s (falsos `inaccesible`); reintentos automáticos (el botón manual ya es el reintento) |
| D21 | **Receta verificable** (US5): la receta vive en `docs/tecnico/arquitectura.md` §8; F1 la verifica con (a) el dominio `status`, área real construida con ella, y (b) el ejercicio de práctica de `quickstart.md` §9 en rama aparte (área pública de solo lectura), ejecutado por **el humano** (confirmación 3, siguiendo solo §8) con checklist de SC-007 rellena y `make ci` en verde. La exigencia de *persona ajena a la receta* es el **Independent Test de US5** (si no hay otra persona disponible, válvula de escape: R13) | Cumple SC-007 literal ("se sigue de principio a fin… sin modificar las áreas existentes") sin dejar en `main` tablas de práctica ni funcionalidad visible (Out of Scope de la spec) | Receta como documento aparte en `specs/` (duplicaría `arquitectura.md`); validar solo con `status` (no cumple SC-007: falta un área nueva; ni el Independent Test de US5: falta la persona ajena); ejercicio con escritura o panel (arrastraría `rate-limit`/`CSRF`/`authn` diferidos a F2); dejar el ejercicio fusionado en `main` (ensucia F1 con una tabla no de negocio) |
| D22 | **Dependencias mínimas** (D-A8): en F1 solo `pgx/v5` entra al binario; `sqlc`, `golang-migrate` y `openapi-typescript` son herramientas de desarrollo. Toda dependencia nueva se justifica en el `plan.md` que la introduce | §II: cada dependencia es un coste que se ve antes de escribir código; el binario solo enlaza lo que usa | Instalar el toolchain completo desde F1 "por si acaso" (configuración muerta) |
| D23 | **Diferidos con punto de decisión** (D-A9): `validate` (F2), sesiones/authn/authz/permisos (F2 — **propuesta pendiente de confirmación humana**), `CSRF`/`rate-limit` (primer endpoint público escribible), `paginate` (F2), i18n (F3), métricas/trazas (primer despliegue operado), colas (primera necesidad de trabajo diferido), imágenes (F8), generación de handlers desde el contrato (revisión post-MVP) | Nada de esto lo necesita F1; cada diferido tiene anclado el momento en que se decide, para no implementarlo "de paso" | Adoptar todo desde F1 (dependencias sin uso); aplazar sin decidir dónde se decide |

## Project Structure

### Documentation (this feature y documentos fundacionales)

```text
specs/001-estructura-base/
├── plan.md              # EDITABLE (este archivo, /speckit.plan — reescrito 2026-09-30)
├── research.md          # EDITABLE (fase 0: decisiones con justificación y alternativas)
├── data-model.md        # EDITABLE (fase 1: sin tablas de negocio + convención de migraciones)
├── quickstart.md        # EDITABLE (fase 1: guía de validación end-to-end, incluye la receta)
├── spec.md              # APROBADA — no se edita
├── ux.md                # Del disenador-ux — no lo edita el arquitecto (ajustes ya aplicados: ver sección "Ajustes en ux.md")
├── contracts/
│   └── openapi.yaml     # EDITABLE (fase 1): borrador de diseño del contrato (fuente viva: backend/api/openapi.yaml)
├── checklists/          # (existente, de fases anteriores)
└── tasks.md             # Fase 2 (/speckit.tasks — NO lo crea este comando)

docs/tecnico/            # Fundacionales y ya aprobados — este plan los aterriza, no los toca
├── arquitectura.md      # Reglas de dependencia, árbol de carpetas, Registrar, ejemplo vertical, §8 = receta
└── decisiones.md        # D-A1…D-A9 (referencia obligada de este plan)

docs/producto/roadmap.md # F1…F9 (fuente de alcance)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── api/
│       └── main.go                  # NUEVO — composición manual (única raíz): config → logger → pool → dominios → servidor
├── internal/
│   ├── db/                          # NUEVO — CÓDIGO GENERADO por sqlc (vacío en F1: sin consultas de negocio)
│   │   └── queries/                 # NUEVO — SQL anotado (vacío en F1; la primera consulta llega con la primera área con tablas)
│   ├── status/                      # NUEVO — dominio "estado del sistema" (patrón de la receta, D21)
│   │   ├── model.go                 # DTO SystemStatus
│   │   ├── repository.go            # implementación de la interfaz del service (pgxpool.Ping con timeout)
│   │   ├── service.go               # interfaz Repository + reglas (sin HTTP, sin SQL)
│   │   ├── handler.go               # interfaz Service + HTTP (responde 200/503 por WriteJSON/WriteError)
│   │   ├── routes.go                # RegisterPublic(r httpserver.Registrar, h)
│   │   └── *_test.go                # service con repo falso; handler con service falso (httptest)
│   └── platform/                    # NUEVO — plataforma interna (D11)
│       ├── config/                  # variables de entorno + validación al arrancar
│       ├── logger/                  # slog JSON + logger por petición
│       ├── apperr/                  # errores tipados de dominio (Invalid, NotFound, …, Internal)
│       ├── httpserver/              # servidor con timeouts + apagado ordenado; Registrar + Middleware; WriteJSON/WriteError
│       ├── middleware/              # request-id, recover, logging, CORS mínimo (rate-limit/CSRF/authn/authz: F2)
│       ├── database/                # pool pgx desde DATABASE_URL + Ping con timeout + WithTx
│       └── testutil/                # helpers compartidos (conexión a DATABASE_URL_TEST con skip, httptest); nombre que evita el choque con el paquete `testing` de la stdlib
│       # diferidos (no en F1): validate/ (F2) · paginate/ (F2) · migrate/ (cuando un entorno desplegado lo pida)
├── migrations/
│   ├── 000001_baseline.up.sql       # NUEVO — no-op (comentario): fija la convención
│   └── 000001_baseline.down.sql     # NUEVO — no-op
├── api/
│   └── openapi.yaml                 # NUEVO — CONTRATO VIVO (se crea copiando contracts/openapi.yaml)
├── sqlc.yaml                        # NUEVO — schema: migrations · queries: internal/db/queries · out: internal/db
├── go.mod / go.sum                  # NUEVO — módulo provisional `simiente-santa/backend` (path definitivo con el remoto, R1)
└── Dockerfile                       # NUEVO — multi-stage: golang:1.27 build → imagen mínima (lo crea devops)

frontend/
├── src/
│   ├── app/                         # NUEVO — router (React Router), providers (QueryClientProvider), layout
│   ├── api/
│   │   ├── client.ts                # NUEVO — fetch tipado base (URL desde import.meta.env.VITE_API_URL)
│   │   ├── status.ts                # NUEVO — función tipada getSystemStatus() (único lugar que hace fetch del estado)
│   │   └── schema.d.ts              # GENERADO con openapi-typescript (commiteado; npm run api:gen)
│   ├── components/                  # NUEVO — UI genérica reutilizable (vacía o casi en F1)
│   ├── features/
│   │   └── status/
│   │       ├── pages/StatusPage.tsx        # NUEVO — página inicial ("PaginaEstado" de ux.md)
│   │       ├── hooks/useSystemStatus.ts    # NUEVO — TanStack Query sobre getSystemStatus (consulta al montar + refetch manual)
│   │       ├── components/                 # NUEVO — indicadores de estado (conectado / no conectado)
│   │       └── *.test.tsx                  # NUEVO — estados: cargando, ok, BD caída, backend inalcanzable
│   ├── lib/                         # NUEVO — utilidades
│   ├── main.tsx                     # NUEVO
│   └── index.css                    # NUEVO — Tailwind
├── e2e/
│   ├── status.spec.ts               # NUEVO — Playwright: la página muestra el estado del sistema
│   └── playwright.config.ts         # NUEVO
├── package.json / package-lock.json # NUEVO
├── vite.config.ts / tsconfig.json / tailwind.config.* / eslint / prettier   # NUEVO
├── nginx.conf                       # NUEVO — SPA fallback para la imagen Docker
└── Dockerfile                       # NUEVO — multi-stage: node:22 build → nginx:alpine (lo crea devops)

# Raíz (estado de cada archivo)
docker-compose.yml               # EDITABLE (no está en .kit-manifest.json): descomentar backend/frontend, puertos parametrizados (D2, D3)
.env.example                     # EDITABLE: agregar DB_PORT y WEB_PORT opcionales; sin secretos reales (FR-008)
proyecto.mk                      # NUEVO: targets propios (e2e, api-gen, sqlc-gen, sqlc-verify) incluidos por el Makefile del kit
README.md                        # NUEVO: quickstart del proyecto, puertos, comandos, versiones (FR-009)
.github/workflows/ci.yml         # SIN CAMBIOS (kit): ya valida backend/frontend/migraciones/secretos
Makefile                         # SIN CAMBIOS (kit): up/down/test/lint/security/ci/db-migrate/doctor
.githooks/                       # SIN CAMBIOS (kit): pre-commit (formato/secretos) y commit-msg (Conventional Commits); YA instalados en este clon (R3)
```

**Structure Decision**: opción "Web application" del monorepo constitucional, siguiendo las estructuras de las skills `go-backend` y `react-frontend` y el árbol de `docs/tecnico/arquitectura.md` §2. El dominio `internal/status/` y el feature `features/status/` son el **patrón de referencia** que replicarán los dominios de negocio en F2–F9 (misma disposición de archivos, mismas convenciones de pruebas, mismo cableado en `main.go`).

## Reutilización del kit (qué NO se construye)

| Necesidad de la spec | Ya resuelto por | Qué falta en F1 |
|---|---|---|
| FR-001 comando único | `Makefile` (`up`/`down`) + `docker-compose.yml` (servicio `db` con healthcheck) | Descomentar servicios `backend`/`frontend` y crear sus Dockerfiles |
| FR-006/FR-007 pipeline con veredicto | `.github/workflows/ci.yml`: detecta `backend/go.mod` y `frontend/package.json`, corre lint, pruebas con PostgreSQL de servicio, migraciones, `govulncheck`, `npm audit`, gitleaks, control de migraciones inmutables | Nada de pipeline: solo que el código pase. *La protección de rama es ajuste humano de GitHub (R1)* |
| FR-008 config por entorno | `.env.example` (ya documenta `POSTGRES_*`, `DATABASE_URL`, `HTTP_PORT`, `LOG_LEVEL`, `CORS_ALLOWED_ORIGINS`, `VITE_API_URL`), `.gitignore` y hook pre-commit anti-secretos | Agregar vars opcionales de puertos; el backend lee las suyas en `platform/config` |
| FR-009 documentación | `docs/GUIA-INICIO.md` (kit: instalación de herramientas) + `make doctor` | `README.md` del proyecto con el quickstart concreto |
| Migraciones | Target `db-migrate` + paso "Migraciones" del CI + hooks de inmutabilidad | Carpeta `backend/migrations/` con la baseline `000001` |
| Calidad | Targets `lint`/`test`/`security`/`ci`, hooks pre-commit (gofmt, prettier) y commit-msg (Conventional Commits) | Que el código nuevo los pase |
| Herramientas de generación | `proyecto.mk` (vía oficial de extensión del Makefile) | Targets `api-gen`, `sqlc-gen`, `sqlc-verify`, `e2e` |

## Cobertura de requisitos (FR-001…FR-016)

Ningún requisito queda fuera de F1: los FR-001…FR-016 son el alcance de esta funcionalidad. Lo que sí queda fuera son las áreas de negocio (F2–F9), la autenticación (F2) y el diseño visual/i18n (F3), tal como fija la spec.

| FR | Dónde se resuelve en este plan | Verificación |
|---|---|---|
| FR-001 | D1, D2, D3 + `research.md` R1/R2 | `quickstart.md` §0 (SC-001) |
| FR-002 | D7 + `contracts/openapi.yaml` (`/healthz`) | `quickstart.md` §1–2 + pruebas del handler |
| FR-003 | D8 (ping por petición) | `quickstart.md` §3 + pruebas del service |
| FR-004 | D8 (pool perezoso + timeout) + `recover` (D11) | `quickstart.md` §2 |
| FR-005 | D14 + estructura `frontend/features/status/` + `ux.md` | `quickstart.md` §4 |
| FR-006 | D1 (CI del kit, sin tocar) | `quickstart.md` §6 |
| FR-007 | D1 + R1 (branch protection, acción humana) | `quickstart.md` §6 |
| FR-008 | D12 (`platform/config`) + `.env.example` + hooks/gitleaks | `quickstart.md` §0 + auditoría de `seguridad` |
| FR-009 | D18 (`README.md`) + `quickstart.md` + `docs/GUIA-INICIO.md` (kit) | `quickstart.md` §0 (persona nueva, SC-001) |
| FR-010 | D11 (plataforma interna) + `arquitectura.md` §2.1 | SC-006 (revisión de `revisor-codigo` + búsqueda de reimplantaciones) |
| FR-011 | D11 + reglas de dependencia (`arquitectura.md` §1.2) | SC-006 (revisión) |
| FR-012 | D13 (sobres de respuesta) + `contracts/openapi.yaml` | SC-008 (suite del sobre, sección de pruebas) |
| FR-013 | D13 (`WriteError` único, `internal` genérico, detalle solo en log con `request_id`) | SC-009 (suite con error provocado) |
| FR-014 | D21 (receta en `arquitectura.md` §8) + `quickstart.md` §9 | SC-007 (checklist de receta) |
| FR-015 | D21 (unidad aislada: solo archivos nuevos + una línea de cableado) | SC-007 (`git diff` del ejercicio) |
| FR-016 | Tabla de versiones (sección "Métricas del plan") + R10/R11 + política de actualización | `quickstart.md` §10 (SC-010) |

## Contrato OpenAPI: dónde vive y cómo se mantiene

1. **Diseño (esta fase)**: el contrato se redacta en `specs/001-estructura-base/contracts/openapi.yaml` — artefacto de diseño inmutable una vez aprobado el plan, igual que `plan.md`.
2. **Documento vivo**: la primera tarea de implementación backend lo copia verbatim a `backend/api/openapi.yaml` (ubicación exigida por §II). A partir de ahí, **ese** es el único contrato fuente: se edita antes de tocar código en cada funcionalidad futura.
3. **Funcionalidades futuras**: cada `specs/<N>-<feature>/contracts/` contiene el *delta* de diseño (nuevos paths/esquemas/códigos de error); la implementación lo fusiona en `backend/api/openapi.yaml`. Nunca se edita el snapshot de `specs/` a posteriori.
4. **Consumo**: el frontend genera sus tipos desde `backend/api/openapi.yaml` (`npm run api:gen`), nunca desde el snapshot de `specs/`.

## Formato de respuesta uniforme (FR-012, FR-013)

Dos sobres, definidos en el contrato y escritos **únicamente** por `internal/platform/httpserver` (`WriteJSON` para éxito, `WriteError` para error). Ningún handler escribe códigos de estado ni cuerpos de error a mano.

1. **Sobre de éxito (2xx)** — el DTO documentado de la operación:

   ```json
   {"status": "ok", "database": "connected"}
   ```

   Convenciones comunes a toda la API: `Content-Type: application/json`; objeto raíz; campos en `camelCase`; `additionalProperties: false`; fechas ISO-8601 en UTC; ningún tipo ni mensaje interno de Go. El consumidor lo maneja con un solo mecanismo: `apiFetch<T>()` parsea la respuesta como el tipo `T` generado del contrato.

2. **Sobre de error (4xx/5xx)** — siempre `ErrorEnvelope`:

   ```json
   {"error": {"code": "database_unavailable", "message": "La base de datos no está conectada", "details": {"database": "disconnected"}}}
   ```

   - `code`: registro cerrado en `snake_case` — `invalid`, `unauthenticated`, `forbidden`, `not_found`, `method_not_allowed`, `conflict`, `rate_limited`, `database_unavailable`, `internal`. **F1 puede emitir**: `not_found` (ruta no documentada), `method_not_allowed` (método no documentado), `database_unavailable` (`/healthz` con la BD caída), `internal` (error inesperado o `panic` recuperado). El resto nace con F2+ al ritmo de su contrato.
   - `message`: texto seguro y comprensible en español. Nunca contiene nombres internos, trazas, SQL ni datos de infraestructura (FR-013).
   - `details` (opcional): detalle por campo en validaciones (F2+) o datos de diagnóstico seguros (`{"database": "disconnected"}`).
   - Caso `internal`: `message` genérico (`"Error interno del servidor"`) y el detalle interno **solo** en el log estructurado, con su `request_id` (SC-009).

3. **Cobertura del 100 % (SC-008)**: el adaptador del router convierte también los 404/405 de la stdlib (que responden texto) en `ErrorEnvelope`, y el middleware `recover` convierte cualquier `panic` en un 500 con sobre de error. Así, **toda** respuesta de la API —documentada o no— cumple el formato.

## Cadena de middleware (orden)

La monta `httpserver.New`; el primero de la lista es el más externo (`docs/tecnico/arquitectura.md` §6):

```text
F1:      request-id → recover → logging → CORS → handler
F2+:     request-id → recover → logging → CORS → rate-limit → [authn → authz → CSRF] → handler
                                        (los de grupo los añade Registrar.Group; el dominio no los menciona)
```

- `request-id`: toma `X-Request-ID` del cliente o genera uno; lo guarda en el `context` y crea el **logger por petición** (hij de `slog` con `request_id`, método y ruta). Va el primero para que todo lo demás loguee con trazabilidad.
- `recover`: captura cualquier `panic` de la cadena interna y responde 500 vía `WriteError` (el proceso nunca cae por una petición).
- `logging`: al terminar registra método, ruta, status, duración y `request_id`.
- `CORS`: responde los preflight `OPTIONS` y fija las cabeceras desde `CORS_ALLOWED_ORIGINS`; corta ahí los preflight.
- `rate-limit` (diferido): con el primer endpoint público escribible. `[authn → authz → CSRF]` (F2): solo en grupos, vía `Group(...)`.

## Estrategia de pruebas por capa

| Capa | Tipo | Herramienta | Qué verifica | Dónde |
|---|---|---|---|---|
| `service` | Unitaria | `testing` + **fake** de la interfaz `Repository` | Reglas con tabla de casos (`tests := []struct{...}` + `t.Run`), errores `apperr`, `%w` envuelto | `internal/<dominio>/service_test.go` |
| `handler` | Unitaria | `httptest.NewRecorder` + **fake** de `Service` | Decodificación, códigos de estado, sobre de éxito y de error, contenido de la respuesta | `internal/<dominio>/handler_test.go` |
| `repository` | **Integración** | PostgreSQL real (`DATABASE_URL_TEST`), `//go:build integration` | SQL real, restricciones, traducción de errores | `internal/<dominio>/repository_test.go` |
| `platform` | Unitaria | tabla de casos + `httptest` | `apperr`, cadena de middlewares, `Registrar` (incluido `Group`), timeouts/apagado del servidor | `internal/platform/*/**_test.go` |
| **Sobre de respuestas** (SC-008/SC-009) | Unitaria + aceptación | `httptest` sobre el stack completo (router + cadena + handlers) provocando éxito, error previsto y **error inesperado** | 100 % de respuestas en sobre; `internal` sin información interna; detalle en log con `request_id` | `internal/platform/httpserver/` |
| Frontend | Unitaria | Vitest + Testing Library + **MSW** | Estados cargando/éxito/BD-caída/inaccesible; mapeo de respuestas del sobre a los estados de `ux.md` | `frontend/src/features/<feature>/*.test.tsx` |
| E2E | End-to-end | **Playwright** (local) | Flujo de estado contra el stack levantado (`make up`) | `frontend/e2e/*.spec.ts` |

Comandos: `go test ./...` (unitarias) · `go test -tags=integration ./...` (integración; *skip* si no hay `DATABASE_URL_TEST`) · `npm test -- --run` · `make e2e` · `make ci` (el veredicto completo: lint, pruebas, migraciones, `govulncheck`, `npm audit`).

Cobertura mínima exigida: **80 % en `service/`** (constitución §III); la verifican `qa-tester` y `revisor-codigo`. Ningún cambio se integra con pruebas fallando.

## La receta de áreas de negocio y su verificación (US5, SC-007)

- **Dónde vive la receta**: `docs/tecnico/arquitectura.md` §8 (10 pasos: contrato → migración → consultas sqlc → `model.go` → `repository.go` → `service.go` → `handler.go` → `routes.go` → cableado en `main.go` → pruebas). Este plan **no la duplica**: la hace seguible y la verifica.
- **Por qué es seguible en F1**: la plataforma interna (D11) ya provee todo lo que la receta pide usar (`config`, `logger`, `apperr`, `httpserver.WriteJSON/WriteError`, `Registrar`, `database`, `testutil`) y las convenciones de carpetas están fijadas. El dominio `status` es el **primer área real construida con ella** (su ejemplo, no una hipótesis de documentación).
- **Cómo se verifica (SC-007)** — `quickstart.md` §9: **el humano** (confirmación 3) sigue los 10 pasos **solo** desde `arquitectura.md` §8 en una rama de práctica (`practica/receta-001`) para agregar un área nueva **pública y de solo lectura** (ejemplo sugerido: dominio `muestra`, tabla `sample_items`, `GET /api/v1/muestra`). El *Independent Test* de **US5** es quien pide que la receta la siga una persona que **no** participó en su creación (y sin consultar decisiones de arquitectura); si no hay otra persona disponible se aplica la válvula de escape de R13 (validación humana explícita del ejercicio) y así queda registrado. Al terminar: `make ci` en verde, `git diff` sin cambios en áreas existentes (solo archivos nuevos + la línea de cableado en `main.go`), y checklist de SC-007 rellena. La rama se **descarta**: F1 cierra con 0 tablas de negocio (la tabla de práctica nunca llega a `main`).
- **Límites declarados**: la variante de panel (`RegisterAdmin` con `authn`/`authz`) se verifica en F2, cuando existan esos middlewares; el ejercicio de práctica no tiene entradas de usuario (así evita arrastrar `platform/validate`, `rate-limit` y `CSRF`, diferidos — D23). Si SC-007 se quisiera verificar con la receta completa incluido panel, sería criterio de aceptación de F2.

## Métricas del plan (coherentes con los SC aprobados)

| Criterio | Qué se mide | Cómo se mide en F1 | Umbral |
|---|---|---|---|
| SC-001 | Tiempo de arranque desde clon limpio | Persona nueva siguiendo `quickstart.md` §0 (cronómetro) | < 15 min |
| — | Arranque repetible (US1 esc. 2) | `make down && make up` sin estado residual | 100 % |
| SC-002 | Distinción "conectada" / "no conectada" | `quickstart.md` §1–3 (ambos escenarios + recuperación) + pruebas automatizadas | 100 % |
| SC-003 | Cambios con veredicto automático | PRs con el workflow `CI` + branch protection (R1) | 100 % |
| SC-004 | Latencia del veredicto | Duración del workflow `CI` en los primeros PRs | < 10 min |
| SC-005 | Página de estado legible al cargar | Navegador / e2e Playwright | < 3 s |
| SC-006 | Capacidades transversales una sola vez | Revisión de `revisor-codigo` + búsqueda de reimplantaciones fuera de `platform/` | 0 reimplantaciones |
| SC-007 | Receta aplicada de principio a fin | Checklist de receta (`quickstart.md` §9) rellena por el humano que la ejecuta (confirmación 3; Independent Test de US5 con persona ajena a la receta si la hay — R13) | 1 aplicación en verde |
| SC-008 | Formato uniforme de respuestas | Suite del sobre de respuestas + respuestas observadas en las pruebas de aceptación | 100 % |
| SC-009 | Sin información interna ante errores inesperados | Suite que provoca error inesperado: respuesta `internal` genérica + log con detalle y `request_id` | 100 % |
| SC-010 | Versiones con soporte de seguridad vigente | Tabla de versiones (abajo) + `govulncheck`/`npm audit` de `make ci` | 100 % |

**Versiones en uso y su soporte (SC-010), comprobado a 2026-09-30**:

| Tecnología | Versión en uso | Estado de soporte | Nota |
|---|---|---|---|
| Go | 1.27 | Vigente (solo 1.26 y 1.27 reciben parches; 1.23 está en fin de vida desde 2025-08-12) | D5; se actualiza tocando `go.mod` y el `FROM` del Dockerfile |
| Node.js | 22 | En mantenimiento hasta abril de 2027 | Lo fija el `ci.yml` del kit (no editable) → R11: proponer la actualización al kit antes de esa fecha |
| PostgreSQL | 16 (imagen `postgres:16.4-alpine`) | La rama 16 está en soporte hasta noviembre de 2028; **el minor 16.4 acumula CVEs corregidos en minors posteriores** | R10: queda identificado como **pendiente de actualización** (US7 esc. 2); propuesta `postgres:16-alpine` en el compose editable. La misma imagen `postgres:16.4-alpine` la usa el **servicio `postgres` del `ci.yml` del kit** (no editable aquí): pendiente aparte, ver registro siguiente |
| React / TypeScript / Vite | 19 / 5.x / actual | Mantenidas; vulnerabilidades vía `npm audit` | Sin versiones fijas por constitución, solo `strict` |
| `pgx/v5`, `sqlc`, `golang-migrate`, `openapi-typescript` | versiones fijadas en `go.sum` / `package-lock.json` / README | `govulncheck` y `npm audit` sin altas/críticas (§IV) | `sqlc`/`migrate`/`openapi-typescript` son herramientas de desarrollo |

**Registro de pendientes de actualización (SC-010, US7 esc. 2), a 2026-09-30** — cada uno identificado, propuesto y gestionado por el humano (mismo mecanismo que R11):

1. `postgres:16.4-alpine` del servicio `db` de `docker-compose.yml` (R10; decisión en T034): propuesta `postgres:16-alpine`.
2. `postgres:16.4-alpine` del **servicio `postgres` del `ci.yml` del kit** (archivo no editable, regla 10): acumula los mismos CVEs que el compose → propuesta al repositorio del kit (T035).
3. **Node 22**, fijado por el `ci.yml` del kit (R11; T035): soporte de mantenimiento hasta abril de 2027 → propuesta de actualización a una LTS vigente.
4. **«Go 1.23+»** que aún recomienda `docs/GUIA-INICIO.md` del kit: una rama en fin de vida desde 2025-08-12 (el proyecto usa Go 1.27, D5) → propuesta de actualizar el texto al repositorio del kit (T035).

**Política (FR-016)**: toda funcionalidad futura que fije o actualice una versión revisa esta tabla; lo que deje de estar en soporte se registra como pendiente de actualización antes de seguir construyendo sobre ello.

## Riesgos

| Riesgo | Impacto | Mitigación |
|---|---|---|
| **R1** — El repositorio **no tiene remoto configurado** y su rama por defecto actual es `master`, mientras que el `ci.yml` del kit se dispara en push solo a `main` | Sin remoto en GitHub, el CI de FR-006/FR-007 no corre en ningún sitio; y sin protección de rama, SC-003 no queda garantizado | **Dos acciones humanas pendientes**: (a) crear el repositorio remoto y empujar, renombrando la rama por defecto a `main` al hacerlo (o proponer el ajuste al repo del kit); (b) activar la branch protection de la rama principal con los checks del CI obligatorios. Se reporta al orquestador |
| **R2** — El `ci.yml` del kit no ejecuta Playwright ni exige umbral de cobertura ≥80 % | Un flujo crítico roto o cobertura baja podrían integrarse en verde | Playwright corre local (documentado en `quickstart.md` §7); cobertura la verifican `qa-tester`/`revisor-codigo`. **Propuesta para el repo del kit** (regla 10): job e2e y chequeo de umbral |
| **R3** — Los hooks de git **ya están instalados en este clon** (`core.hooksPath=.githooks`), pero son por clon y se pueden saltar (`--no-verify`) | Un clon nuevo sin `make instalar-hooks` pierde el control local; en ningún caso los hooks son la garantía de integración | README documenta `make instalar-hooks` como paso una vez por clon; la garantía real es el CI + branch protection (R1). Los hooks son la red de seguridad temprana, no el control |
| **R4** — **Deriva de los artefactos generados**: `sqlc generate` y `npm run api:gen` son pasos manuales de desarrollo (el `ci.yml` del kit no se edita) | Consultas/esquema cambiados sin regenerar `internal/db/`, o contrato cambiado sin regenerar `schema.d.ts` | Código generado **commiteado** (un SQL que no acompaña falla en compilación); `make sqlc-verify` (regenera y exige `git diff --exit-code`) y `make api-gen` en `proyecto.mk`; regla de revisión: un PR que toca `migrations/` o `queries/` debe tocar `internal/db/`, y uno que toca el contrato debe tocar `schema.d.ts`; versión de `sqlc` fijada en el README para generación reproducible. **Propuesta al kit**: paso de CI que verifique ambos |
| **R5** — `sqlc generate` sin archivos de consulta (caso de `main` en F1) | Comportamiento incierto del CLI con `internal/db/queries/` vacío | Verificarlo en la primera tarea de la capa de datos; si el CLI exige al menos una consulta, la receta ya lo ejecuta tras escribir la primera (flujo normal del paso 3) y `main` queda sin `internal/db/` generado hasta entonces |
| **R6** — `VITE_API_URL` se hornea en build time; la imagen del frontend se construye con el valor por defecto `http://localhost:8080` | Si alguien cambia el puerto del backend, la imagen del frontend debe reconstruirse | Documentado en README; aceptable para entorno local (único alcance de F1) |
| **R7** — Primer arranque: la BD tarda más que el backend (edge case de la spec) | `/healthz` debe decir "no conectada" sin caerse | El backend nunca falla por BD ausente: pool perezoso de `pgx` + `Ping` por petición (D8). Además compose arranca `db` con `service_healthy` antes que backend |
| **R8** — `migrate up` sobre una migración no-op | El paso "Migraciones" del CI podría comportarse distinto de lo esperado | La baseline es SQL trivial (comentario), `golang-migrate` la aplica sin efecto; se verifica en el primer PR |
| **R9** — Sustitución de variables de compose: si el usuario copia `.env.example` a `.env`, `DATABASE_URL` apunta a `localhost` | El contenedor backend no alcanzaría la BD si heredara ese valor | D3: compose construye el `DATABASE_URL` del contenedor desde `POSTGRES_*` con host `db`; el `localhost` solo aplica a herramientas del host (`make db-migrate`) |
| **R10** — Versión de PostgreSQL: el compose fija `postgres:16.4-alpine`, un minor con vulnerabilidades corregidas en minors posteriores | SC-010/FR-016: la imagen en uso no está en su última corrección de seguridad | Queda **identificado como pendiente de actualización** (US7 esc. 2): propuesta de usar `postgres:16-alpine` (rama en soporte con parches) en el compose editable — decisión de `devops`/humano al implementar. Separar decisión de reproducibilidad (Dockerfile) de la base de datos de desarrollo |
| **R11** — Node 22 lo fija el `ci.yml` del kit; su soporte de mantenimiento termina en **abril de 2027** | A partir de esa fecha, SC-010 dejaría de cumplirse sin que el proyecto pueda remediarlo solo | Registrar el vencimiento en la tabla de versiones; proponer al repo del kit la actualización a una versión LTS vigente con antelación (regla 10: no se edita `ci.yml` aquí) |
| **R12** — US6 esc. 3 y SC-009 (errores inesperados) no pueden provocarse contra `GET /healthz` (una sola operación sin entrada) | La verificación literal de "pruebas de aceptación" contra la API viva sería incompleta en F1 | La suite del sobre de respuestas provoca errores previstos e inesperados sobre el stack completo (router + cadena + handlers) con `httptest` y verifica respuesta genérica + log con detalle; queda documentado como limitación de F1 y se refuerza con el ejercicio de práctica y con las primeras operaciones con entrada en F2 |
| **R13** — SC-007 exige que la receta la siga **una persona que no participó en su creación** | Si no hay otra persona disponible, la verificación quedaría sesgada | Escala al humano/orquestador: o bien una persona del equipo distinta de quien escribió la receta, o bien validación humana explícita del ejercicio; no se cierra F1 sin el registro del SC-007 |

## Complexity Tracking

> Sin violaciones de la constitución que justificar.

## Ajustes en ux.md (aplicados)

Los tres ajustes que este plan pedía a `disenador-ux` ya están **aplicados** en `ux.md` (2026-09-30; sin cambios de decisión):

1. **Mapeo de la respuesta al estado** — aplicado en `ux.md` §3.1 (tipo `EstadoSistema` en §4): `200` + `SystemStatus` válido → *conectado*; `503` + `error.code = "database_unavailable"` → *bd-no-conectada* (error A); cualquier otra cosa (timeout de 5 s, error de red, respuesta que no cumple el sobre) → *inaccesible* (error B, con motivo *sin respuesta* / *respuesta inesperada*).
2. **El mensaje del sobre de error** (`error.message`, p. ej. "La base de datos no está conectada") — aplicado en `ux.md` §7.1.6: no se muestra; los literales de `ux.md` §6 mandan.
3. **`details` del 503** (`{"database": "disconnected"}`) — aplicado en `ux.md` §7.1.7: no se muestra en F1 (decisión de contenido reversible).

Quedan **solo** como dudas de contenido al humano las de `ux.md` §7.2.1 (usar `error.message` como explicación de apoyo en el error A) y §7.2.2 (visibilidad de `details` para quien opera). El resto de `ux.md` §7.2 (§7.2.3–§7.2.6) son notas de diseño que no requieren decisión humana en F1.
