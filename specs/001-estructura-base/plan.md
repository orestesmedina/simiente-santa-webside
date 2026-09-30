# Implementation Plan: Estructura base del proyecto (F1)

**Branch**: `001-estructura-base` | **Date**: 2026-09-29 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-estructura-base/spec.md` (aprobada)

## Summary

F1 monta el esqueleto técnico que copiarán F2–F9: backend en Go con `GET /healthz` que informa en cada consulta el estado real de la conexión a PostgreSQL (sin estado memorizado), frontend en React que muestra ese estado, todo levantable con `make up` (Docker Compose) desde un clon limpio, y validación automática de cada cambio vía el pipeline de CI que **el kit ya provee**. El enfoque es **reutilizar al máximo el kit instalado** (Makefile, `docker-compose.yml`, `.github/workflows/ci.yml`, hooks de git, `.env.example`, `doctor.sh`) y agregar únicamente lo que el kit no trae: el código de `backend/` y `frontend/`, los Dockerfiles, el contrato OpenAPI y el README del proyecto.

## Technical Context

**Language/Version**: Go 1.23+ (backend, `go.mod`); TypeScript 5.x `strict` sobre Node 22 (frontend); React 19

**Primary Dependencies**: backend → `pgx/v5` (única dependencia externa de runtime; router `net/http` estándar); frontend → React 19, Vite, React Router, TanStack Query, Tailwind CSS, Vitest + Testing Library + MSW, Playwright (e2e local), `openapi-typescript` (tipos desde el contrato)

**Storage**: PostgreSQL 16.4-alpine (servicio `db` ya definido en `docker-compose.yml` del kit). **Sin tablas de negocio en F1** (ver `data-model.md`)

**Testing**: `go test` (unitarias) + `go test -tags=integration` (repositorio contra PostgreSQL real, `DATABASE_URL_TEST`); Vitest + Testing Library + MSW (frontend); Playwright (e2e, ejecución local)

**Target Platform**: contenedores Docker en máquina local de desarrollo (Linux/macOS/WSL) + GitHub Actions (`ubuntu-latest`)

**Project Type**: web-service (monorepo `backend/` + `frontend/`)

**Performance Goals**: `/healthz` responde aunque la BD esté caída (ping con timeout de 2 s); página inicial muestra el estado en <3 s desde la carga (SC-005); pipeline <10 min (SC-004)

**Constraints**: stack fijo por constitución; un solo comando de arranque (`make up`) con Docker como único prerequisito (FR-001/SC-001); sin secretos en el repo (FR-008); mínimo de dependencias nuevas, cada una justificada (§II)

**Scale/Scope**: esqueleto patrón para F2–F9: 1 endpoint, 1 página, 0 tablas de negocio, 1 migración trivial de baseline

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principio | Evaluación | Resultado |
|---|---|---|
| §I La spec manda | El plan implementa FR-001…FR-009 tal cual; sin comportamiento inventado. | ✅ |
| §II Arquitectura | Monorepo `backend/` + `frontend/` + `backend/migrations/`. Capas `handler → service → repository` con interfaces (dominio `internal/status/`). API REST documentada en `backend/api/openapi.yaml`, contrato escrito **antes** que el código (ver "Contrato OpenAPI" abajo). Dependencias minimizadas y justificadas en `research.md`. | ✅ |
| §III Pruebas | Toda tarea de implementación incluye pruebas (se exigirá en `tasks.md`). Cobertura ≥80 % en `service/` (verificable con `go test -cover`; la exigencia de umbral la aplican QA/revisor, ver Riesgos). Pruebas de repositorio con tag `integration` contra PostgreSQL real (servicio `postgres` del CI del kit). Frontend con Vitest + Testing Library; e2e con Playwright (DEBERÍA constitucional → incluido, ejecución local). | ✅ |
| §IV Seguridad | Sin entrada de usuario en F1 (único endpoint `GET` sin parámetros) → sin superficie de inyección; cuando la haya, solo consultas parametrizadas vía `pgx`. Secretos por variables de entorno; `.env.example` sin valores reales; hooks de git y gitleaks en CI (kit) ya lo vigilan. `govulncheck` y `npm audit` ya corren en el CI del kit. | ✅ |
| §V Calidad | `gofmt`, `go vet`, `golangci-lint` (targets `lint` del kit + CI). TS `strict`, sin `any`. Errores envueltos con `%w`. | ✅ |
| §VI Base de datos | Migraciones versionadas con `golang-migrate` (convención `NNNNNN_nombre.up.sql`/`.down.sql`). F1 solo crea la migración baseline `000001` (no-op) para establecer la convención; los hooks del kit y el CI ya impiden modificar migraciones aplicadas. | ✅ |
| §VII Observabilidad | Logs estructurados `log/slog` en JSON (nivel por `LOG_LEVEL`). Endpoint `/healthz` (nombre exigido). Configuración por variables de entorno. Todo el sistema se levanta con `docker compose up` (vía `make up`). | ✅ |
| §VIII Gobierno | Este plan requiere aprobación humana antes de `/speckit.tasks` e implementación. Sin despliegue a producción en F1. | ✅ |

**Sin violaciones que justificar** → sección "Complexity Tracking" vacía.

## Decisiones técnicas (resumen; análisis completo en `research.md`)

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| D1 | Reutilizar el kit tal cual: `make up` como comando único, CI de `.github/workflows/ci.yml` sin tocar, hooks, `.env.example`, `doctor.sh` | El kit ya resuelve FR-001, FR-006, FR-007 y FR-008; duplicarlo violaría la regla 10 de AGENTS.md | Crear Makefile/CI propios del proyecto |
| D2 | `docker-compose.yml` **sí es editable** (no está en `.kit-manifest.json`): descomentar servicios `backend`/`frontend`, parametrizar puertos (`${VAR:-defecto}`) y **no** exigir `env_file` | Permite `make up` en clon limpio sin pasos manuales (FR-001 estricto) y resuelve el edge case de puertos ocupados | Exigir `cp .env.example .env` como paso previo obligatorio |
| D3 | El `DATABASE_URL` del contenedor backend se construye en compose desde `POSTGRES_*` apuntando al host `db` | Evita el error clásico de que `DATABASE_URL=...@localhost:5432` (válido en el host) se filtre al contenedor, donde `localhost` no es la BD | Usar `env_file: .env` a secas |
| D4 | Router `net/http` estándar (patrones `GET /healthz` de Go 1.22+), sin `chi` | Un solo endpoint; la stdlib basta y la skill go-backend lo permite por defecto | `chi` (innecesario hoy; se puede introducir en F2 si el ruteo crece) |
| D5 | `pgx/v5` directo con `pgxpool.Ping(ctx)`; **sin `sqlc` en F1** | No hay consultas SQL de negocio; `sqlc` sin queries es configuración muerta. Se introduce en F2 con las primeras tablas | Configurar `sqlc` ya "por si acaso" |
| D6 | Contrato `/healthz`: `200 {"status":"ok","database":"connected"}` / `503 {"status":"degraded","database":"disconnected"}`, siempre con cuerpo JSON y `Cache-Control: no-store` | El 503 con cuerpo distingue "BD caída" de "backend caído" (fetch rechaza) y sigue la convención de health checks; FR-004 se cumple: el backend responde | Siempre 200 con un campo (pierde la semántica estándar para monitoreo futuro) |
| D7 | El estado de la BD se consulta con `Ping` en **cada** petición, con `context.WithTimeout` de 2 s | FR-003 (estado real, no memorizado) sin riesgo de que una BD colgada bloquee el endpoint | Caché/sondeo en segundo plano (violaría FR-003) |
| D8 | Contrato fuente único: `backend/api/openapi.yaml` (exigido por §II). `specs/001-estructura-base/contracts/openapi.yaml` es el **borrador de diseño** de esta fase; la primera tarea de implementación lo copia verbatim a `backend/api/openapi.yaml`, que desde entonces es el documento vivo | Cumple §II (contrato antes que el código, ubicación fija) sin mantener dos copias vivas que diverjan: la de `specs/` es snapshot inmutable de diseño, como `plan.md` | Mantener ambos archivos sincronizados a mano (divergencia garantizada) |
| D9 | Migración baseline `000001_baseline` no-op (solo comentario) con su `.down.sql` | Establece la convención de `backend/migrations/` con archivos reales y ejercita el paso "Migraciones" del CI desde F1; el esquema de negocio llega en F2 | Carpeta vacía con `.gitkeep` (deja el toolchain de migraciones sin probar) |
| D10 | Frontend: React Router + TanStack Query (**consulta al montar + `refetch` a demanda** con el botón "Volver a consultar el estado"; **sin auto-refresco en F1**) + Tailwind CSS; **shadcn/ui se difiere a F3** | Router/Query/Tailwind son la convención de la skill y el patrón que copiarán F2–F9; la spec no exige refresco automático (SC-005 = ver el estado al abrir la página; FR-003 = que cada consulta refleje el estado real, no que la página se consulte sola) y el diseño de estados de `ux.md` es manual; shadcn/ui sin diseño definido sería peso muerto (la spec excluye diseño visual) | Sondeo periódico (`refetchInterval`): **nota de futuro**, puede añadirse en F3+ si el cliente lo pide — no es decisión de F1; incluir shadcn/ui ya; no usar Query para un solo fetch (rompería el patrón) |
| D11 | Tipos de API generados con `openapi-typescript` desde `backend/api/openapi.yaml`, **commiteados** en `frontend/src/api/schema.d.ts` | CI corre `npm ci && typecheck` sin pasos extra; se regeneran con `npm run api:gen` cuando cambie el contrato | Generarlos en CI (más pasos, mismas garantías) |
| D12 | CORS mínimo escrito a mano (una cabecera para un `GET` simple sin credenciales) | Evita una dependencia para un solo header; F2 introducirá cookies de sesión y reevaluará middleware completo (`rs/cors`) | Dependencia `rs/cors` desde ya |
| D13 | Playwright configurado con un e2e del flujo de estado, ejecución **local** (`npx playwright test`), no en CI | §III dice DEBERÍA para flujos críticos y monta la infraestructura que F2–F9 necesitarán; el `ci.yml` es del kit y no se puede editar (regla 10) — se propone al humano llevar un job e2e al kit | No montar Playwright (retrasaría la infra a F2) |
| D14 | `README.md` en la raíz (no existe; no es del kit) con quickstart, puertos y comandos | FR-009 exige documentación para levantar el entorno sin ayuda; `docs/GUIA-INICIO.md` es del kit y genérica (entorno/herramientas), no del proyecto | Documentar solo en `docs/` (el README es lo primero que lee una persona nueva) |
| D15 | Targets propios en `proyecto.mk` (incluido por el Makefile del kit): `e2e`, `api-gen` | El Makefile es del kit y no se edita; `proyecto.mk` es el mecanismo previsto para extensiones | Editar el Makefile (prohibido) |
| D16 | El fetch del frontend a `/healthz` usa `AbortController` con timeout de **5 s**; timeout, error de red o respuesta ininterpretable mapean al estado `inaccesible` ("No se pudo consultar" de `ux.md`) | 5 s > 2 s del ping de BD (D7): una BD lenta o caída produce un 503 `degraded` legible (error A de `ux.md`) antes de que el frontend se rinda, en vez de un falso "sin respuesta"; resuelve el hueco 2 reportado por `ux.md`. No requiere cambios en el contrato: es comportamiento del cliente | Timeout por defecto del navegador (decenas de segundos con la página en "Consultando…"); timeout ≤ 2 s (falsos `inaccesible` cuando la BD solo está lenta); reintentos automáticos (el botón manual ya es el reintento, según `ux.md`) |

## Project Structure

### Documentation (this feature)

```text
specs/001-estructura-base/
├── plan.md              # Este archivo (/speckit.plan)
├── research.md          # Fase 0: decisiones con justificación y alternativas
├── data-model.md        # Fase 1: decisión explícita de "sin tablas de negocio"
├── quickstart.md        # Fase 1: guía de validación end-to-end
├── contracts/
│   └── openapi.yaml     # Fase 1: borrador de diseño del contrato (fuente viva: backend/api/openapi.yaml)
├── checklists/          # (existente, de fases anteriores)
└── tasks.md             # Fase 2 (/speckit.tasks — NO lo crea este comando)
```

### Source Code (repository root)

```text
backend/
├── cmd/
│   └── api/
│       └── main.go              # arranque: config, pool de BD, router, servidor con apagado ordenado
├── internal/
│   ├── config/
│   │   ├── config.go            # carga de variables de entorno con valores por defecto
│   │   └── config_test.go
│   ├── status/                  # dominio "estado del sistema" (patrón que copiarán los dominios F2+)
│   │   ├── handler.go           # GET /healthz: decodifica nada, llama al service, responde 200/503
│   │   ├── service.go           # lógica: interpreta el resultado del repositorio
│   │   ├── repository.go        # interfaz Pinger + implementación pgxpool (Ping con timeout)
│   │   ├── model.go             # DTO SystemStatus
│   │   └── *_test.go            # service con repo falso; handler con service falso (httptest)
│   └── platform/
│       ├── database/            # construcción del pgxpool.Pool desde DATABASE_URL
│       └── httpserver/          # servidor con timeouts, middleware CORS mínimo, logging slog
├── migrations/
│   ├── 000001_baseline.up.sql   # no-op (comentario): establece la convención
│   └── 000001_baseline.down.sql
├── api/
│   └── openapi.yaml             # CONTRATO VIVO (se crea copiando contracts/openapi.yaml)
├── go.mod / go.sum
└── Dockerfile                   # multi-stage: golang:1.23 build → imagen mínima (devops)

frontend/
├── src/
│   ├── app/                     # router (React Router), providers (QueryClientProvider), layout
│   ├── api/
│   │   ├── client.ts            # fetch tipado base (URL desde import.meta.env.VITE_API_URL)
│   │   └── schema.d.ts          # generado con openapi-typescript (commiteado)
│   ├── components/              # UI genérica reutilizable (vacío o casi en F1)
│   ├── features/
│   │   └── status/
│   │       ├── pages/StatusPage.tsx        # página inicial (la "PaginaEstado" de ux.md): estado del backend y la BD
│   │       ├── hooks/useSystemStatus.ts    # TanStack Query sobre GET /healthz (consulta al montar + refetch manual, sin auto-refresco)
│   │       ├── components/                 # indicadores de estado (conectado / no conectado)
│   │       └── *.test.tsx                  # estados: cargando, ok, BD caída, backend inalcanzable
│   ├── lib/                     # utilidades
│   ├── main.tsx
│   └── index.css                # Tailwind
├── e2e/
│   ├── status.spec.ts           # Playwright: la página muestra el estado del sistema
│   └── playwright.config.ts
├── package.json / package-lock.json
├── vite.config.ts / tsconfig.json / tailwind.config.* / eslint / prettier
├── nginx.conf                   # SPA fallback para la imagen Docker
└── Dockerfile                   # multi-stage: node:22 build → nginx:alpine (devops)

# Raíz (cambios sobre lo existente)
docker-compose.yml               # EDITABLE (no es del kit): descomentar backend/frontend, puertos parametrizados
.env.example                     # editable: agregar vars opcionales de puertos (DB_PORT, WEB_PORT)
proyecto.mk                      # NUEVO: targets propios (e2e, api-gen) incluidos por el Makefile del kit
README.md                        # NUEVO: quickstart del proyecto (FR-009)
.github/workflows/ci.yml         # SIN CAMBIOS (kit): ya valida backend/frontend/migraciones/secretos
Makefile                         # SIN CAMBIOS (kit): up/down/test/lint/security/ci/db-migrate/doctor
```

**Structure Decision**: opción "Web application" del monorepo constitucional, siguiendo las estructuras de las skills `go-backend` y `react-frontend`. El dominio `internal/status/` y el feature `features/status/` son el **patrón de referencia** que replicarán los dominios de negocio en F2–F9 (misma disposición de archivos, mismas convenciones de pruebas).

## Contrato OpenAPI: dónde vive y cómo se mantiene

1. **Diseño (esta fase)**: el contrato se redacta en `specs/001-estructura-base/contracts/openapi.yaml` — es artefacto de diseño inmutable una vez aprobado el plan, igual que `plan.md`.
2. **Documento vivo**: la primera tarea de implementación backend lo copia verbatim a `backend/api/openapi.yaml` (ubicación exigida por §II). A partir de ahí, **ese** es el único contrato fuente: se edita antes de tocar código en cada funcionalidad futura.
3. **Funcionalidades futuras**: cada `specs/<N>-<feature>/contracts/` contiene el *delta* de diseño (nuevos paths/esquemas); la implementación lo fusiona en `backend/api/openapi.yaml`. Nunca se edita el snapshot de `specs/` a posteriori.
4. **Consumo**: el frontend genera sus tipos desde `backend/api/openapi.yaml` (`npm run api:gen`), nunca desde el snapshot de `specs/`.

## Reutilización del kit (qué NO se construye)

| Necesidad de la spec | Ya resuelto por | Qué falta en F1 |
|---|---|---|
| FR-001 comando único | `Makefile` (`up`/`down`) + `docker-compose.yml` (servicio `db` con healthcheck) | Descomentar servicios `backend`/`frontend` y crear sus Dockerfiles |
| FR-006/FR-007 pipeline con veredicto | `.github/workflows/ci.yml`: detecta `backend/go.mod` y `frontend/package.json`, corre lint, pruebas con PostgreSQL de servicio, migraciones, `govulncheck`, `npm audit`, gitleaks, control de migraciones inmutables | Nada de pipeline: solo que el código pase. *La protección de rama es ajuste humano de GitHub (ver Riesgos)* |
| FR-008 config por entorno | `.env.example` (ya documenta `POSTGRES_*`, `DATABASE_URL`, `HTTP_PORT`, `LOG_LEVEL`, `CORS_ALLOWED_ORIGINS`, `VITE_API_URL`), `.gitignore` y hook pre-commit anti-secretos | Agregar vars opcionales de puertos; el backend lee las suyas |
| FR-009 documentación | `docs/GUIA-INICIO.md` (kit: instalación de herramientas) + `make doctor` | `README.md` del proyecto con el quickstart concreto |
| Migraciones | Target `db-migrate` + paso "Migraciones" del CI + hooks de inmutabilidad | Carpeta `backend/migrations/` con la baseline `000001` |
| Calidad | Targets `lint`/`test`/`security`/`ci`, hooks pre-commit (gofmt, prettier) y commit-msg (Conventional Commits) | Que el código nuevo los pase |

## Riesgos

| Riesgo | Impacto | Mitigación |
|---|---|---|
| **R1** — El repositorio **no tiene remoto configurado** y su rama por defecto actual es `master` | Sin remoto en GitHub, el CI de FR-006/FR-007 no corre en ningún sitio (`ci.yml` existe solo como archivo); y sin protección de rama, SC-003 ("ningún cambio con validaciones fallidas queda integrado") no queda garantizado | **Dos acciones humanas pendientes**: (a) crear el repositorio remoto en GitHub y empujar las ramas —ojo: el `ci.yml` del kit se dispara en todo PR pero en push solo a `main`, así que conviene renombrar la rama por defecto a `main` al crear el remoto (o proponer el ajuste al repo del kit); (b) tras empujar, activar la branch protection de la rama principal con los checks del CI obligatorios. Se reporta al orquestador |
| **R2** — El `ci.yml` del kit no ejecuta Playwright ni exige umbral de cobertura ≥80 % | Un flujo crítico roto o cobertura baja podrían integrarse en verde | Playwright corre local (documentado en quickstart); cobertura la verifican `qa-tester`/`revisor-codigo`. **Propuesta para el repo del kit** (regla 10): job e2e y chequeo de umbral |
| **R3** — `VITE_API_URL` se hornea en build time; la imagen del frontend se construye con el valor por defecto `http://localhost:8080` | Si alguien cambia el puerto del backend, la imagen del frontend debe reconstruirse | Documentado en README; aceptable para entorno local (único alcance de F1) |
| **R4** — Primer arranque: la BD tarda más que el backend (edge case de la spec) | `/healthz` debe decir "no conectada" sin caerse | El backend nunca falla por BD ausente: pool perezoso de `pgx` + `Ping` por petición (D7). Además compose arranca `db` con `service_healthy` antes que backend |
| **R5** — `migrate up` sobre una migración no-op | El paso "Migraciones" del CI podría comportarse distinto de lo esperado | La baseline es SQL trivial (comentario), `golang-migrate` la aplica sin efecto; se verifica en el primer PR |
| **R6** — Sustitución de variables de compose: si el usuario copia `.env.example` a `.env`, `DATABASE_URL` apunta a `localhost` | El contenedor backend no alcanzaría la BD si heredara ese valor | D3: compose construye el `DATABASE_URL` del contenedor desde `POSTGRES_*` con host `db`; el `localhost` solo aplica a herramientas del host (`make db-migrate`) |

## Complexity Tracking

> Sin violaciones de la constitución que justificar.
