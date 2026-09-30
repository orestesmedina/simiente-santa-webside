# Research — F1 Estructura base

**Fecha**: 2026-09-29 · **Rama**: `001-estructura-base`

No había ningún `NEEDS CLARIFICATION` en el Technical Context: el stack es restricción del proyecto (constitución + AGENTS.md) y las convenciones internas están fijadas en las skills `go-backend`, `react-frontend` y `postgres-db`. La investigación consistió en **explorar el repo** (qué resuelve ya el kit) y decidir los puntos que la spec deja deliberadamente al plan. Formato: Decisión / Justificación / Alternativas consideradas.

## R0. Qué reutilizar del kit instalado

- **Decisión**: construir F1 sobre el kit sin modificarlo: `Makefile` (`up`, `down`, `test`, `lint`, `security`, `ci`, `db-migrate`, `doctor`), `.github/workflows/ci.yml`, `.githooks/`, `.env.example`, `scripts/doctor.sh`, `docs/GUIA-INICIO.md`.
- **Justificación**: `.kit-manifest.json` confirma que esos archivos son del kit (regla 10 de AGENTS.md: no editarlos). El CI ya cumple FR-006/FR-007: se dispara en todo PR y en push a `main`, detecta `backend/go.mod` y `frontend/package.json`, corre gofmt/vet/golangci-lint, migraciones, pruebas con PostgreSQL 16 de servicio, `govulncheck`, ESLint/typecheck/Vitest/build, `npm audit` y gitleaks, con veredicto visible por job. Ojo con la realidad del repo: hoy **no tiene remoto configurado** y su rama por defecto es `master` — la puesta en marcha efectiva del CI es una acción humana (crear remoto y empujar; ver R1 de `plan.md`).
- **Alternativas consideradas**: pipeline propio del proyecto (rechazada: duplicaría el kit y se perdería en la próxima actualización de este).
- **Verificado en repo**: `docker-compose.yml` y `.env.example` **no** están en `.kit-manifest.json` → son editables por el proyecto. El Makefile incluye `-include proyecto.mk` como vía oficial de extensión.

## R1. Comando único de arranque (FR-001/SC-001)

- **Decisión**: `make up` (`docker compose up -d`). En `docker-compose.yml` se descomentan los servicios `backend` y `frontend` (el propio archivo los trae comentados "para el primer proyecto"), con puertos parametrizados `${DB_PORT:-5432}:5432`, `${HTTP_PORT:-8080}:8080`, `${WEB_PORT:-5173}:80` y **sin** `env_file` obligatorio: toda variable usa `${VAR:-defecto}` con los mismos valores de `.env.example`.
- **Justificación**: compose lee `.env` automáticamente si existe, pero con valores por defecto el arranque funciona en un clon limpio sin ningún paso manual — la lectura estricta de FR-001 ("Docker como único prerequisito"). Los puertos parametrizados resuelven el edge case "puerto ocupado" (basta `DB_PORT=5433 make up`, documentado en README).
- **Alternativas consideradas**: exigir `cp .env.example .env` antes del arranque (válido según la spec, pero añade un paso manual evitable); perfiles de compose dev/prod (complejidad innecesaria en F1).

## R2. URL de la BD dentro de la red de compose

- **Decisión**: el servicio `backend` recibe `DATABASE_URL` construido en compose como `postgres://${POSTGRES_USER:-app}:${POSTGRES_PASSWORD:-app_dev_password}@db:5432/${POSTGRES_DB:-app}?sslmode=disable`.
- **Justificación**: el `DATABASE_URL` de `.env.example` usa `localhost:5432`, correcto para herramientas del host (`make db-migrate`, `go test -tags=integration`) gracias al mapeo de puertos, pero **incorrecto dentro del contenedor** (donde `localhost` es el propio contenedor). Construirlo desde las partes con host `db` evita el error clásico de copiar `.env` y romper el backend.
- **Alternativas consideradas**: `env_file: .env` a secas (rechazada por el fallo anterior); una variable `DATABASE_URL_COMPOSE` extra (innecesaria habiendo partes).

## R3. Router HTTP del backend

- **Decisión**: `net/http` estándar con patrones de ruta de Go 1.22+ (`mux.HandleFunc("GET /healthz", ...)`).
- **Justificación**: la skill `go-backend` lo da como opción por defecto y F1 expone un único endpoint `GET` sin parámetros. Cero dependencias de ruteo.
- **Alternativas consideradas**: `chi` (rechazada por ahora: dependencia sin necesidad actual; si F2+ requiere middleware de ruteo avanzado, el plan de esa funcionalidad lo justificará).

## R4. Acceso a datos: pgx directo, sqlc diferido

- **Decisión**: `github.com/jackc/pgx/v5` con `pgxpool` como **única dependencia externa de runtime** del backend. El repositorio de estado envuelve `pool.Ping(ctx)`. `sqlc` se introduce en F2, cuando existan las primeras consultas de negocio.
- **Justificación**: la constitución (§IV) exige consultas parametrizadas vía `pgx`/`sqlc`; en F1 no hay ninguna consulta de negocio, solo el ping de salud. Configurar `sqlc` sin archivos `.sql` sería herramienta muerta. `pgx` es dependencia obligada en cualquier caso y su pool perezoso (`pgxpool.New` no conecta de inmediato) es exactamente lo que FR-004 necesita: el proceso arranca y responde aunque la BD no exista.
- **Alternativas consideradas**: `database/sql` + driver `lib/pq` (rechazada: la skill fija `pgx/v5`); instalar `sqlc` ya (rechazada: configuración sin uso).

## R5. Semántica del endpoint de estado (FR-002/003/004)

- **Decisión**: `GET /healthz` responde siempre con cuerpo JSON `SystemStatus{status, database}`:
  - BD accesible → `200 OK` con `{"status":"ok","database":"connected"}`
  - BD inaccesible → `503 Service Unavailable` con `{"status":"degraded","database":"disconnected"}`
  - Cabecera `Cache-Control: no-store` en ambas. El chequeo es un `Ping` con `context.WithTimeout(2s)` ejecutado **en cada petición**; no se guarda estado del arranque.
- **Justificación**: el 503 con cuerpo permite distinguir tres situaciones desde el frontend y desde cualquier monitor: backend vivo + BD viva (200), backend vivo + BD caída (503 con JSON), backend caído (la conexión TCP/fetch falla). Es la convención estándar de health checks (facilita readiness/liveness futuros) y cumple FR-004: el backend *sigue respondiendo* (un 503 es una respuesta). El ping por petición con timeout corto cumple FR-003 (estado real en cada consulta) sin riesgo de que una BD colgada bloquee el endpoint.
- **Alternativas consideradas**: siempre `200` con un campo `database` (rechazada: pierde la señal estándar para monitoreo y obliga a todo cliente a parsear el cuerpo); sondeo en segundo plano con estado en memoria (rechazada: viola FR-003 literalmente); endpoint separado `/healthz` (liveness) + `/readyz` (readiness) (rechazada en F1: la spec pide UN endpoint de estado; la separación puede llegar con el despliegue real).

## R6. Dónde vive el contrato OpenAPI

- **Decisión**: el documento vivo es `backend/api/openapi.yaml` (§II). El archivo de esta fase, `specs/001-estructura-base/contracts/openapi.yaml`, es el **borrador de diseño**: la primera tarea de implementación lo copia verbatim a su ubicación definitiva. Las funcionalidades futuras diseñan su *delta* en su propio `specs/<N>/contracts/` y lo fusionan en el documento vivo al implementar. Los snapshots de `specs/` no se editan después de aprobados.
- **Justificación**: cumple las dos exigencias sin duplicación viva: §II (contrato en `backend/api/openapi.yaml`, escrito antes que el código) y el flujo Spec Kit (contratos como artefacto de diseño por funcionalidad). Una sola fuente viva → imposible que diverjan.
- **Alternativas consideradas**: mantener ambos sincronizados a mano (rechazada: divergencia garantizada); usar `specs/.../contracts/` como fuente y referenciarlo desde backend (rechazada: viola §II y mezcla diseño histórico con estado actual).
- **Versión de OpenAPI**: 3.1.0 (soportada por `openapi-typescript` v7; sin generación de código servidor que limite la versión).

## R7. Base de datos sin esquema de negocio

- **Decisión**: ninguna tabla en F1. Se crea `backend/migrations/000001_baseline.up.sql` / `.down.sql` con contenido no-op (comentario) para establecer la convención de migraciones y ejercitar el paso "Migraciones" del CI.
- **Justificación**: la spec lo excluye explícitamente ("la base de datos solo necesita existir y aceptar conexiones"). La baseline deja el toolchain (`golang-migrate`, target `db-migrate`, control de inmutabilidad de hooks/CI) probado antes de que F2 cree las primeras tablas reales (`users`, etc., empezando en `000002`).
- **Alternativas consideradas**: carpeta vacía con `.gitkeep` (rechazada: deja el paso de migraciones del CI sin ejercitar de verdad); crear ya la tabla `users` (rechazada: fuera de alcance, es de F2).
- Ver `data-model.md`.

## R8. Stack del frontend

- **Decisión**: scaffold Vite (plantilla `react-ts`) con React 19 y TypeScript `strict`; React Router; TanStack Query; Tailwind CSS; Vitest + Testing Library + MSW; Playwright; `openapi-typescript`. **shadcn/ui diferido a F3.**
- **Justificación**: es la pila de la skill `react-frontend`; F1 la monta completa porque es el patrón que copiarán F2–F9. La página de estado usa TanStack Query con **consulta al montar + `refetch` a demanda** (botón "Volver a consultar el estado"), siguiendo el flujo manual decidido en `ux.md`: la spec no exige auto-refresco (SC-005 pide ver el estado al abrir la página; FR-003 exige que cada consulta refleje el estado real, no que la página se consulte sola), y el diseño de estados de `ux.md` (botón deshabilitado durante la reconsulta, "Actualizando…", hora de la última consulta) asume ese flujo. **Nota de futuro**: el refresco periódico (polling) puede añadirse en F3+ si el cliente lo pide; no es decisión de F1. Tailwind se instala como base de estilos de todo el proyecto aunque la página de F1 sea funcional. shadcn/ui se defiere porque la spec excluye diseño visual y su configuración (tema, tokens) pertenece a F3.
- **Alternativas consideradas**: `fetch` directo en el componente (rechazada: prohibido por la skill; el hook `useSystemStatus` es el patrón a replicar); incluir shadcn/ui ya (rechazada: dependencias y decisiones de diseño sin necesidad en F1); Redux/Zustand (rechazada: no hay estado global de cliente en F1; TanStack Query cubre el estado de servidor); auto-refresco por sondeo con `refetchInterval` (rechazada en F1: la spec no lo exige y contradiría el diseño de estados de `ux.md`).

## R9. Tipos del contrato en el frontend

- **Decisión**: `openapi-typescript` genera `frontend/src/api/schema.d.ts` desde `backend/api/openapi.yaml` vía script `npm run api:gen`; el archivo generado **se commitea**.
- **Justificación**: la skill exige tipos generados del contrato y prohibe `any`. Commitearlo permite que el CI del kit (`npm ci` → `typecheck` → `test`) funcione sin pasos adicionales no editables en `ci.yml`.
- **Alternativas consideradas**: generar en CI/prebuild (rechazada: exigiría editar el `ci.yml` del kit o añadir pasos opacos); tipos escritos a mano (rechazada: viola la convención y diverge del contrato).

## R10. CORS

- **Decisión**: middleware mínimo escrito a mano en `internal/platform/httpserver`: fija `Access-Control-Allow-Origin` desde `CORS_ALLOWED_ORIGINS` para el `GET` simple de `/healthz`, sin credenciales.
- **Justificación**: una sola ruta `GET` sin cookies ni preflight complejo no justifica una dependencia. F2 introducirá sesión con cookie `HttpOnly` (skill) y su plan reevaluará un middleware completo (`rs/cors` o equivalente) con credenciales.
- **Alternativas consideradas**: `github.com/rs/cors` desde ya (rechazada: dependencia sin necesidad actual); proxy nginx `/api` en el contenedor frontend en lugar de CORS (rechazada: añade configuración de nginx para algo que CORS resuelve con una cabecera; el patrón `VITE_API_URL` + CORS ya está en `.env.example`).

## R11. E2E con Playwright fuera del CI

- **Decisión**: Playwright queda instalado y configurado con una prueba del flujo de estado (`frontend/e2e/`), ejecutable con `npx playwright test` (target `make e2e` en `proyecto.mk`). **No corre en CI** porque `.github/workflows/ci.yml` es del kit y no se edita.
- **Justificación**: §III marca los e2e como DEBERÍA para flujos críticos; montar la infraestructura ahora (con el flujo más simple que existirá jamás) sale casi gratis y F2–F9 la heredan. La ejecución en CI se propone al humano como mejora del repositorio del kit (regla 10).
- **Alternativas consideradas**: no montar Playwright (rechazada: aplaza la infraestructura a F2, que ya tendrá presión de alcance); editar el `ci.yml` (prohibida).

## R12. Imágenes Docker de desarrollo

- **Decisión**: `backend/Dockerfile` multi-stage (`golang:1.23` compila → imagen mínima con el binario) y `frontend/Dockerfile` multi-stage (`node:22` compila con Vite → `nginx:alpine` sirve estáticos en el puerto 80, con `nginx.conf` de fallback SPA). Compose publica `5173:80` como ya anticipa la plantilla comentada del kit.
- **Justificación**: imágenes autocontenidas y reproducibles (SC-001/SC-002, arranque repetible); el desarrollo diario con hot-reload sigue siendo posible fuera de Docker (`npm run dev`, `go run ./cmd/api`) contra la BD de compose — se documenta en README/quickstart.
- **Alternativas consideradas**: contenedores de desarrollo con volúmenes y hot-reload (rechazada en F1: más frágil entre SO — especialmente WSL — y la spec pide entorno reproducible, no hot-reload); servir el frontend desde el backend Go (rechazada: rompe la separación del monorepo y el patrón de la skill).

## R13. Variables de entorno

- **Decisión**: el backend lee `APP_ENV`, `HTTP_PORT`, `DATABASE_URL`, `LOG_LEVEL`, `CORS_ALLOWED_ORIGINS` (todas con valores por defecto de desarrollo, ya documentadas en `.env.example`). `SESSION_SECRET` permanece en `.env.example` pero **el backend de F1 no la consume** (es de F2). Se agregan a `.env.example` las opcionales `DB_PORT` y `WEB_PORT`.
- **Justificación**: FR-008 y §VII; ninguna variable es obligatoria para arrancar en local (valores por defecto), ningún valor real es secreto en desarrollo.
- **Alternativas consideradas**: librería de configuración (`viper`, `envconfig`) (rechazada: `os.Getenv` con un pequeño paquete `internal/config` basta y es testeable sin dependencias).

## R14. Documentación del proyecto (FR-009)

- **Decisión**: crear `README.md` en la raíz: prerequisito (Docker), arranque (`make up`), verificación (`/healthz`, página inicial), puertos usados y cómo cambiarlos, comandos (`make help`, `make test`, `make ci`, `make e2e`), y cómo detener (`make down`).
- **Justificación**: no existe README y no es archivo del kit; es lo primero que lee una persona nueva (SC-001: entorno en <15 min siguiendo solo la documentación). `docs/GUIA-INICIO.md` (kit) cubre la instalación de herramientas del entorno, no el arranque de este proyecto.
- **Alternativas consideradas**: ampliar `docs/GUIA-INICIO.md` (rechazada: es archivo del kit, no editable).

## R15. Timeout de la consulta del frontend

- **Decisión**: el fetch del frontend a `/healthz` se hace con `AbortController` y timeout de **5 segundos**. El mapeo a los estados de `ux.md` queda: `200` con cuerpo válido → `conectado`; `503` con cuerpo válido → `bd-no-conectada` (error A); timeout, error de red o respuesta ininterpretable → `inaccesible` ("No se pudo consultar", error B).
- **Justificación**: resuelve el hueco 2 reportado por `ux.md` (la spec no cubre "el backend cuelga sin responder"). 5 s > 2 s del timeout del ping a la BD (R5/D7): si la BD está lenta o caída, el backend responde `503 degraded` **antes** de que el frontend se rinda, y la persona ve el diagnóstico correcto ("base de datos no conectada") en lugar de un falso "sin respuesta". Es comportamiento del cliente: el contrato `contracts/openapi.yaml` no requiere cambios (sus dos respuestas posibles, 200 y 503, siguen siendo las únicas; el caso "sin respuesta" nunca toca el contrato).
- **Alternativas consideradas**: timeout por defecto del navegador (rechazada: decenas de segundos con la página en "Consultando…", inaceptable para el caso de backend colgado); timeout ≤ 2 s (rechazada: falsos `inaccesible` cuando la BD solo está lenta); reintentos automáticos (rechazada en F1: el botón "Volver a consultar el estado" ya es el reintento, según `ux.md`).
