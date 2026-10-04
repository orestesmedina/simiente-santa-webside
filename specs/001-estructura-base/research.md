# Research — F1 Estructura base

**Fecha**: 2026-09-29 · **Actualizado**: 2026-09-30 (tras la ampliación de alcance y la aprobación de `docs/tecnico/`) · **Rama**: `001-estructura-base`

No había ningún `NEEDS CLARIFICATION` en el Technical Context: el stack es restricción del proyecto (constitución + AGENTS.md) y las convenciones internas están fijadas en las skills `go-backend`, `react-frontend` y `postgres-db`. La investigación consistió en **explorar el repo** (qué resuelve ya el kit) y decidir los puntos que la spec deja deliberadamente al plan. Formato: Decisión / Justificación / Alternativas consideradas.

> **Nota de actualización (2026-09-30)**: las decisiones D-A3…D-A9 de `docs/tecnico/decisiones.md` (aprobadas) **cambian** tres puntos de esta investigación: R3 (router neutralizado tras `Registrar`), R4 (sqlc como capa de datos fijada, ya no diferida) y R5 (el 503 de `/healthz` pasa al sobre de error uniforme). Esas secciones están reescritas; se agregan R16–R20. El resto se mantiene.

## R0. Qué reutilizar del kit instalado

- **Decisión**: construir F1 sobre el kit sin modificarlo: `Makefile` (`up`, `down`, `test`, `lint`, `security`, `ci`, `db-migrate`, `doctor`), `.github/workflows/ci.yml`, `.githooks/`, `.env.example`, `scripts/doctor.sh`, `docs/GUIA-INICIO.md`.
- **Justificación**: `.kit-manifest.json` confirma que esos archivos son del kit (regla 10 de AGENTS.md: no editarlos). El CI ya cumple FR-006/FR-007: se dispara en todo PR y en push a `main`, detecta `backend/go.mod` y `frontend/package.json`, corre gofmt/vet/golangci-lint, migraciones, pruebas con PostgreSQL 16 de servicio, `govulncheck`, ESLint/typecheck/Vitest/build, `npm audit` y gitleaks, con veredicto visible por job. Ojo con la realidad del repo: hoy **no tiene remoto configurado** y su rama por defecto es `master` — la puesta en marcha efectiva del CI es una acción humana (crear remoto y empujar; ver R1 de `plan.md`). Los hooks de git **ya están instalados** en este clon (`core.hooksPath=.githooks`); en clonos nuevos basta `make instalar-hooks` (R3 de `plan.md`).
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

## R3. Router HTTP del backend — *reescrito el 2026-09-30 (D-A4)*

- **Decisión**: **`net/http`** estándar (patrones `GET /ruta`, comodines `{id}` de Go 1.22+) **detrás de la interfaz `httpserver.Registrar`** (`Handle(method, path, h)` y `Group(prefix, mws...)`) y del tipo `type Middleware = func(http.Handler) http.Handler` (alias del stdlib). Los dominios solo ven `Registrar`: F1 implementa el adaptador `muxRegistrar` en `internal/platform/httpserver/` y **chi queda como opción de F2**, sin impacto en los dominios (el cambio futuro se limita a un adaptador de ~15 líneas y a la línea que construye el `Registrar` en `main.go`).
- **Justificación**: F1 expone un único endpoint, pero F2–F9 tendrán grupos de rutas de panel con permisos por módulo; neutralizar el router desde ya evita que el cambio de router toque todos los dominios. Los patrones de `http.ServeMux` de Go 1.22+ cubren el caso actual con cero dependencias de ruteo, y `Middleware` como alias stdlib mantiene compatibilidad total con el ecosistema. Es la decisión D-A4 aprobada; la interfaz completa y su adaptador están en `docs/tecnico/arquitectura.md` §4.
- **Datos medidos (2026-09-30)**: coste de dependencias de los routers evaluados (entradas en `go.mod`, incluidas las de test): **Gin 15 directas / 32 entradas**, **Echo 3/7**, **Fiber 8/13**, **chi 0** (solo stdlib). Gin y Echo imponen su propio tipo de contexto (`gin.Context`, `echo.Context`) que choca con la propagación de `context.Context` de nuestras convenciones; Fiber no implementa `net/http`. Mercado (JetBrains, Go Ecosystem 2025–2026): Gin ~48 %, gorilla/mux 17 % (en declive), Echo 16 %, chi 12 %, Fiber 11 % — la popularidad no decide: decide el coste de dependencias y la compatibilidad con `context.Context` (§II).
- **Alternativas consideradas**: chi desde F1 (pospuesta a F2, no descartada: se reevalúa con los grupos de permisos); Gin/Echo/Fiber (rechazadas por peso y/o contexto propio); dejar los dominios acoplando a `net/http` sin interfaz (rechazada: haría caro el cambio de router); escribir `Registrar` una vez por dominio (rechazada: es duplicación pura — es la única excepción declarada a "la interfaz la define quien la consume", justificada por su neutralidad, `arquitectura.md` §1.3).

## R4. Capa de datos: sqlc — *reescrito el 2026-09-30 (D-A3)*

- **Decisión**: **sqlc** como capa de datos estándar del proyecto: `backend/sqlc.yaml` con `schema: migrations`, `queries: internal/db/queries`, `out: internal/db`; SQL explícito, revisable y siempre parametrizado en `internal/db/queries/*.sql`; código generado **commiteado**. El CLI de sqlc es herramienta de desarrollo (no dependencia de runtime). Las migraciones siguen mandando con `golang-migrate` (sqlc las lee como esquema, no las sustituye). **Stored procedures / funciones SQL solo como excepción justificada**, invocados siempre dentro del `repository`. En F1 la capa de datos hace solo el `Ping` de salud con `pgx` directo (no hay consultas de negocio): el primer código generado llega con la primera consulta real, que es la del ejercicio de práctica de la receta.
- **Justificación**: tipos derivados del SQL real → un cambio de consulta o de esquema que no acompañe **falla en compilación**; el SQL vive en archivos `.sql` (no en cadenas de Go) y se revisa en el PR; 0 entradas nuevas en `go.mod` (sqlc no se enlaza). Nota operativa registrada: un procedimiento invocado con `CALL` dentro de una transacción no puede controlar transacciones; la transacción la gestiona el código Go (`database.WithTx`).
- **Comparativa medida el 2026-09-30** (entradas que cada librería añadiría a `go.mod`, incluidas sus dependencias de test):

  | Opción | Entradas nuevas en `go.mod` | Validación en compilación | Migraciones | Comentario |
  |---|---|---|---|---|
  | **sqlc (elegida)** | **0 en runtime** (herramienta externa: no se enlaza en el binario) | **Sí** — los tipos se generan del SQL real | Usa nuestras migraciones como esquema; `golang-migrate` sigue mandando | SQL explícito, revisable en el PR |
  | GORM | 5 (+9 del driver oficial) | No (reflection) | `AutoMigrate` impone su mecanismo, **choca** con `golang-migrate` | ORM completo; oculta el SQL |
  | ent | 49 | Sí | Atlas (su mecanismo), **choca** con `golang-migrate` | Framework de entidades completo |
  | sqlx | 3 | No (reflection parcial) | — | Finito, pero sin generación ni tipos derivados del SQL |

- **Alternativas consideradas**: las cuatro filas de la tabla; escribir a mano mappers sobre `pgx` (rechazado: trabajo repetido y propenso a errores, sin validación en compilación); dejar la decisión para F2 (lo que decía el plan anterior — sustituida por D-A3).
- **Pendiente asociado**: el tipo de las columnas `uuid` (sqlc emite `github.com/google/uuid` por defecto; puede forzarse `pgtype.UUID` con un *override*). Es dependencia nueva potencial → decisión del `plan.md` de F2 bajo D-A8 (D22 del plan).

## R5. Semántica del endpoint de estado (FR-002/003/004) — *actualizado el 2026-09-30 (sobre de respuesta)*

- **Decisión**: `GET /healthz` responde siempre con cuerpo JSON y cabecera `Cache-Control: no-store`, dentro del formato uniforme (R17):
  - BD accesible → `200 OK` con sobre de éxito `SystemStatus`: `{"status":"ok","database":"connected"}`
  - BD inaccesible → `503 Service Unavailable` con sobre de error: `{"error":{"code":"database_unavailable","message":"La base de datos no está conectada","details":{"database":"disconnected"}}}`
  - El chequeo es un `Ping` con `context.WithTimeout(2s)` ejecutado **en cada petición**; no se guarda estado del arranque.
- **Justificación**: el 503 con cuerpo permite distinguir tres situaciones desde el frontend y desde cualquier monitor: backend vivo + BD viva (200), backend vivo + BD caída (503 con JSON), backend caído (la conexión TCP/fetch falla). Es la convención estándar de health checks y cumple FR-004: el backend *sigue respondiendo* (un 503 es una respuesta). El ping por petición con timeout corto cumple FR-003 sin riesgo de que una BD colgada bloquee el endpoint. El 503 usa el **sobre de error** porque es un fallo **previsible** de la operación (US6 esc. 2: "el formato de error uniforme indica de forma comprensible qué pasa"), lo que mantiene la regla "4xx/5xx = `ErrorEnvelope`" sin excepciones y da al cliente un solo mecanismo.
- **Alternativas consideradas**: 503 con un `SystemStatus` degradado (el cuerpo de 503 de la versión anterior, **descartado** el 2026-09-30 porque obligaría al cliente a manejar dos formas de error y SC-008 perdería el 100 % verificable — ver R17, cuya duda quedó resuelta ese mismo día por confirmación humana: el sobre de éxito es el DTO directo); siempre `200` con un campo `database` (rechazada: pierde la señal estándar para monitoreo y obliga a parsear el cuerpo); sondeo en segundo plano con estado en memoria (rechazada: viola FR-003 literalmente); separar `/healthz` (liveness) y `/readyz` (readiness) (rechazada en F1: la spec pide UN endpoint; la separación puede llegar con el despliegue real).

## R6. Dónde vive el contrato OpenAPI

- **Decisión**: el documento vivo es `backend/api/openapi.yaml` (§II). El archivo de esta fase, `specs/001-estructura-base/contracts/openapi.yaml`, es el **borrador de diseño**: la primera tarea de implementación lo copia verbatim a su ubicación definitiva. Las funcionalidades futuras diseñan su *delta* en su propio `specs/<N>/contracts/` y lo fusionan en el documento vivo al implementar. Los snapshots de `specs/` no se editan después de aprobados.
- **Justificación**: cumple las dos exigencias sin duplicación viva: §II (contrato en `backend/api/openapi.yaml`, escrito antes que el código) y el flujo Spec Kit (contratos como artefacto de diseño por funcionalidad). Una sola fuente viva → imposible que diverjan.
- **Alternativas consideradas**: mantener ambos sincronizados a mano (rechazada: divergencia garantizada); usar `specs/.../contracts/` como fuente y referenciarlo desde backend (rechazada: viola §II y mezcla diseño histórico con estado actual).
- **Versión de OpenAPI**: 3.1.0 (soportada por `openapi-typescript` v7; permite `const` y sin generación de código servidor que limite la versión).

## R7. Base de datos sin esquema de negocio

- **Decisión**: ninguna tabla de negocio en F1. Se crea `backend/migrations/000001_baseline.up.sql` / `.down.sql` con contenido no-op (comentario) para establecer la convención de migraciones y ejercitar el paso "Migraciones" del CI. Quién ejecuta las migraciones en F1: el CLI `golang-migrate` (target `make db-migrate` en desarrollo, paso "Migraciones" del CI); el runner embebido (`platform/migrate`, opt-in con `RUN_MIGRATIONS`) queda diferido hasta que un entorno desplegado necesite auto-migrar sin CLI.
- **Justificación**: la spec lo excluye explícitamente ("la base de datos solo necesita existir y aceptar conexiones"). La baseline deja el toolchain (`golang-migrate`, target `db-migrate`, control de inmutabilidad de hooks/CI) probado antes de que F2 cree las primeras tablas reales (`users`, etc., empezando en `000002`). Ver `data-model.md`.
- **Alternativas consideradas**: carpeta vacía con `.gitkeep` (rechazada: deja el paso de migraciones del CI sin ejercitar de verdad); crear ya la tabla `users` (rechazada: fuera de alcance, es de F2); incluir el runner embebido desde ya (rechazada: configuración muerta — D23).

## R8. Stack del frontend

- **Decisión**: scaffold Vite (plantilla `react-ts`) con React 19 y TypeScript `strict`; React Router; TanStack Query; Tailwind CSS; Vitest + Testing Library + MSW; Playwright; `openapi-typescript`. **shadcn/ui diferido a F3.**
- **Justificación**: es la pila de la skill `react-frontend`; F1 la monta completa porque es el patrón que copiarán F2–F9. La página de estado usa TanStack Query con **consulta al montar + `refetch` a demanda** (botón "Volver a consultar el estado"), siguiendo el flujo manual decidido en `ux.md`: la spec no exige auto-refresco (SC-005 pide ver el estado al abrir la página; FR-003 exige que cada consulta refleje el estado real, no que la página se consulte sola), y el diseño de estados de `ux.md` (botón deshabilitado durante la reconsulta, "Actualizando…", hora de la última consulta) asume ese flujo. **Nota de futuro**: el refresco periódico (polling) puede añadirse en F3+ si el cliente lo pide; no es decisión de F1. Tailwind se instala como base de estilos de todo el proyecto aunque la página de F1 sea funcional. shadcn/ui se defiere porque la spec excluye diseño visual y su configuración (tema, tokens) pertenece a F3.
- **Alternativas consideradas**: `fetch` directo en el componente (rechazada: prohibido por la skill; el hook `useSystemStatus` es el patrón a replicar); incluir shadcn/ui ya (rechazada: dependencias y decisiones de diseño sin necesidad en F1); Redux/Zustand (rechazada: no hay estado global de cliente en F1; TanStack Query cubre el estado de servidor); auto-refresco por sondeo con `refetchInterval` (rechazada en F1: la spec no lo exige y contradiría el diseño de estados de `ux.md`).

## R9. Tipos del contrato en el frontend

- **Decisión**: `openapi-typescript` genera `frontend/src/api/schema.d.ts` desde `backend/api/openapi.yaml` vía script `npm run api:gen`; el archivo generado **se commitea**.
- **Justificación**: la skill exige tipos generados del contrato y prohibe `any`. Commitearlo permite que el CI del kit (`npm ci` → `typecheck` → `test`) funcione sin pasos adicionales no editables en `ci.yml`.
- **Alternativas consideradas**: generar en CI/prebuild (rechazada: exigiría editar el `ci.yml` del kit o añadir pasos opacos); tipos escritos a mano (rechazada: viola la convención y diverge del contrato). El control de deriva de este artefacto generado está en R20.

## R10. CORS

- **Decisión**: middleware mínimo escrito a mano en `internal/platform/middleware`: fija `Access-Control-Allow-Origin` desde `CORS_ALLOWED_ORIGINS` para el `GET` simple de `/healthz`, sin credenciales.
- **Justificación**: una sola ruta `GET` sin cookies ni preflight complejo no justifica una dependencia. F2 introducirá sesión con cookie `HttpOnly` (skill) y su plan reevaluará un middleware completo (`rs/cors` o equivalente) con credenciales.
- **Alternativas consideradas**: `github.com/rs/cors` desde ya (rechazada: dependencia sin necesidad actual); proxy nginx `/api` en el contenedor frontend en lugar de CORS (rechazada: añade configuración de nginx para algo que CORS resuelve con una cabecera; el patrón `VITE_API_URL` + CORS ya está en `.env.example`).

## R11. E2E con Playwright fuera del CI

- **Decisión**: Playwright queda instalado y configurado con una prueba del flujo de estado (`frontend/e2e/`), ejecutable con `npx playwright test` (target `make e2e` en `proyecto.mk`). **No corre en CI** porque `.github/workflows/ci.yml` es del kit y no se edita.
- **Justificación**: §III marca los e2e como DEBERÍA para flujos críticos; montar la infraestructura ahora (con el flujo más simple que existirá jamás) sale casi gratis y F2–F9 la heredan. La ejecución en CI se propone al humano como mejora del repositorio del kit (regla 10).
- **Alternativas consideradas**: no montar Playwright (rechazada: aplaza la infraestructura a F2, que ya tendrá presión de alcance); editar el `ci.yml` (prohibida).

## R12. Imágenes Docker de desarrollo

- **Decisión**: `backend/Dockerfile` multi-stage (**`golang:1.27`** compila → imagen mínima con el binario) y `frontend/Dockerfile` multi-stage (`node:22` compila con Vite → `nginx:alpine` sirve estáticos en el puerto 80, con `nginx.conf` de fallback SPA). Compose publica `5173:80` como ya anticipa la plantilla comentada del kit.
- **Justificación**: imágenes autocontenidas y reproducibles (SC-001/SC-002, arranque repetible); el desarrollo diario con hot-reload sigue siendo posible fuera de Docker (`npm run dev`, `go run ./cmd/api`) contra la BD de compose — se documenta en README/quickstart. La versión de Go es decisión D-A5 (R16); se fija en el `FROM` y en `go.mod`, los dos únicos puntos de actualización.
- **Alternativas consideradas**: contenedores de desarrollo con volúmenes y hot-reload (rechazada en F1: más frágil entre SO — especialmente WSL — y la spec pide entorno reproducible, no hot-reload); servir el frontend desde el backend Go (rechazada: rompe la separación del monorepo y el patrón de la skill).

## R13. Variables de entorno

- **Decisión**: el backend lee `APP_ENV`, `HTTP_PORT`, `DATABASE_URL`, `LOG_LEVEL`, `CORS_ALLOWED_ORIGINS` (todas con valores por defecto de desarrollo, ya documentadas en `.env.example`). `SESSION_SECRET` permanece en `.env.example` pero **el backend de F1 no la consume** (es de F2). Se agregan a `.env.example` las opcionales `DB_PORT` y `WEB_PORT`.
- **Justificación**: FR-008 y §VII; ninguna variable es obligatoria para arrancar en local (valores por defecto), ningún valor real es secreto en desarrollo. La carga y su validación al arrancar viven en `internal/platform/config` (R19).
- **Alternativas consideradas**: librería de configuración (`viper`, `envconfig`) (rechazada: `os.Getenv` con un pequeño paquete `internal/platform/config` basta y es testeable sin dependencias).

## R14. Documentación del proyecto (FR-009)

- **Decisión**: crear `README.md` en la raíz: prerequisito (Docker), arranque (`make up`), verificación (`/healthz`, página inicial), puertos usados y cómo cambiarlos, comandos (`make help`, `make test`, `make ci`, `make e2e`, `make sqlc-verify`), versiones del toolchain (Go 1.27, Node 22, sqlc) y cómo detener (`make down`).
- **Justificación**: no existe README y no es archivo del kit; es lo primero que lee una persona nueva (SC-001: entorno en <15 min siguiendo solo la documentación). `docs/GUIA-INICIO.md` (kit) cubre la instalación de herramientas del entorno, no el arranque de este proyecto.
- **Alternativas consideradas**: ampliar `docs/GUIA-INICIO.md` (rechazada: es archivo del kit, no editable).

## R15. Timeout de la consulta del frontend

- **Decisión**: el fetch del frontend a `/healthz` se hace con `AbortController` y timeout de **5 segundos**. El mapeo a los estados de `ux.md` queda: `200` con `SystemStatus` válido → `conectado`; `503` con `error.code = "database_unavailable"` → `bd-no-conectada` (error A); timeout, error de red o respuesta que no cumple el sobre → `inaccesible` ("No se pudo consultar", error B).
- **Justificación**: resuelve el hueco 2 reportado por `ux.md` (la spec no cubre "el backend cuelga sin responder"). 5 s > 2 s del timeout del ping a la BD (R5): si la BD está lenta o caída, el backend responde `503` **antes** de que el frontend se rinda, y la persona ve el diagnóstico correcto ("base de datos no conectada") en lugar de un falso "sin respuesta". Es comportamiento del cliente: el contrato `contracts/openapi.yaml` no requiere cambios adicionales (sus respuestas posibles siguen siendo 200 y 503; el caso "sin respuesta" nunca toca el contrato).
- **Alternativas consideradas**: timeout por defecto del navegador (rechazada: decenas de segundos con la página en "Consultando…", inaceptable para el caso de backend colgado); timeout ≤ 2 s (rechazada: falsos `inaccesible` cuando la BD solo está lenta); reintentos automáticos (rechazada en F1: el botón "Volver a consultar el estado" ya es el reintento, según `ux.md`).

## R16. Versión de Go y periodo de soporte (FR-016) — *nuevo 2026-09-30 (D-A5)*

- **Decisión**: **Go 1.27** en todo el proyecto: `go 1.27` en `backend/go.mod` y `golang:1.27` en `backend/Dockerfile`. El CI del kit lee `backend/go.mod` (`go-version-file`), así que la versión la decide el proyecto, no el kit.
- **Justificación**: el plan anterior fijaba "Go 1.23+", pero **Go 1.23 alcanzó su fin de vida el 2025-08-12**: hoy solo las ramas **1.26 y 1.27** reciben parches de seguridad (FR-016/SC-010 exigen soporte vigente). La skill `go-backend` pide "Go 1.23 o superior": un mínimo, no un techo. Go 1.27 además trae los patrones de `http.ServeMux` de Go 1.22+ en los que se apoya `Registrar` (R3) sin reservas. Al actualizar Go en el futuro se tocan dos puntos (`go.mod` y el `FROM` del Dockerfile) y se re-ejecuta `make ci`.
- **Alternativas consideradas**: quedarse en Go 1.23 (rechazada: sin soporte de seguridad desde 2025-08-12 — exactamente lo que US7 quiere evitar); Go 1.26 (viable, pero 1.27 tiene la ventana de soporte más larga y no hay dependencias sensibles a la versión en F1); `golang:latest` (rechazada: builds no reproducibles).

## R17. Formato uniforme de respuestas (FR-012/FR-013) — *nuevo 2026-09-30 (D13)*

- **Decisión**: dos sobres, definidos en el contrato y escritos solo por `internal/platform/httpserver`:
  - **Éxito (2xx)**: el DTO documentado de la operación (objeto JSON, campos `camelCase`, `additionalProperties: false`, sin tipos internos). Es el formato que ya fija `docs/tecnico/arquitectura.md` §5.1/§5.7 (`WriteJSON(w, 201, ContactCreatedResponse{…})`).
  - **Error (4xx/5xx)**: siempre `ErrorEnvelope` `{"error":{"code","message","details"?}}` escrito con `WriteError`, único punto de traducción de errores a HTTP (§5.11). Registro de códigos cerrado (`invalid`, `unauthenticated`, `forbidden`, `not_found`, `method_not_allowed`, `conflict`, `rate_limited`, `database_unavailable`, `internal`); F1 emite `not_found`, `method_not_allowed`, `database_unavailable` e `internal`.
  - **Cobertura del 100 % (SC-008)**: el fallback del router convierte los 404/405 de texto de la stdlib en `ErrorEnvelope`, y `recover` convierte los `panic` en 500 con sobre de error. El error `internal` responde mensaje genérico y deja su detalle **solo** en el log estructurado con `request_id` (FR-013/SC-009).
- **Justificación**: US6 pide que quien consume la API maneje cualquier respuesta con un solo mecanismo. El mecanismo es `apiFetch<T>()` (éxito → tipo generado del contrato) + `ApiError` (error → `ErrorEnvelope`): exactamente el patrón de `arquitectura.md` §5.10. Unificar en sobre de error **todo** 4xx/5xx (incluido el 503 de `/healthz`, R5) elimina las excepciones y hace verificable el "100 %" de SC-008. FR-013 se cumple por construcción: solo `WriteError` serializa errores y su tabla de traducción ya distingue `message` (seguro) de error interno (solo log).
- **Alternativas consideradas**:
  - *Wrapper genérico `{"data": …}` para éxito* (lo que la palabra "sobre" sugiere a primera lectura): **descartado** porque contradice los ejemplos aprobados de `docs/tecnico/arquitectura.md` §5.1/§5.7/§5.10 (el cuerpo de éxito es el DTO de la operación y `apiFetch<ContactCreated>` lo parsea tal cual) y no aporta al mecanismo único: el consumidor ya conoce el tipo de cada operación porque vive generado del contrato. **Resuelto por confirmación humana del 2026-09-30**: el sobre de éxito es el **DTO directo**, sin wrapper `{"data": …}` — la duda queda cerrada y la decisión no cambia.
  - *Formato libre por operación* (cada endpoint responde a su manera): rechazada — rompe FR-012 y multiplica el manejo en el cliente.
  - *Dejar los 404/405 de la stdlib*: rechazada — responden `text/plain`, rompen SC-008 ("100 % de las respuestas").
  - *Exponer el detalle del error interno en la respuesta para depurar*: rechazada — viola FR-013 (nombres internos, trazas, SQL) y abre la puerta a ataques; el diagnóstico es el log con `request_id`.

## R18. La receta de áreas de negocio y cómo se verifica (US5, SC-007) — *nuevo 2026-09-30 (D21)*

- **Decisión**: la receta (10 pasos) vive en `docs/tecnico/arquitectura.md` §8 y **no se duplica** en `specs/`. F1 la verifica con dos cosas: (a) el dominio `status` de F1 es el **primer área real construida siguiendo la receta** (sus cinco archivos, su `routes.go`, su cableado en `main.go` y sus pruebas son el patrón, no una hipótesis de documentación); (b) el **ejercicio de práctica** de `quickstart.md` §9: lo ejecuta **el humano** (confirmación 3), siguiendo la receta **solo** desde `arquitectura.md` §8, y agrega un área nueva de práctica (pública y de solo lectura, ejemplo sugerido: dominio `muestra`, tabla `sample_items`, `GET /api/v1/muestra`) en una rama aparte, de principio a fin, con `make ci` en verde y sin modificar las áreas existentes (solo archivos nuevos + la línea de cableado). La rama se descarta: F1 cierra con 0 tablas de negocio.
- **Justificación**: SC-007 exige que la receta se aplique "al menos una vez antes del cierre de F1… por una persona del equipo", sin tomar decisiones de arquitectura, sin modificar áreas existentes y con las validaciones en verde; el **Independent Test de US5** es quien añade que esa persona **no** haya participado en la creación de la receta (y la siga sin consultar decisiones de arquitectura), con la válvula de escape de R13 de `plan.md` si no hay otra persona disponible. El ejercicio de práctica lo cumple sin dejar funcionalidad visible ni tablas de negocio en `main` (Out of Scope de la spec). Que el área sea **pública y de solo lectura** evita arrastrar a F1 lo que está diferido (validación de entradas, `rate-limit`, `CSRF`, `authn`/`authz` — D23): la receta se sigue completa en su variante pública.
- **Alternativas consideradas**: redactar la receta como documento aparte en `specs/` (rechazada: duplicaría `arquitectura.md` y las dos copias divergerían); validar solo con `status` (rechazada: no cumple SC-007 — falta un área nueva — ni el Independent Test de US5 — falta la persona ajena a la receta); ejercicio con formulario o rutas de panel (rechazada: arrastraría middlewares y `platform/validate` diferidos a F2); dejar el ejercicio fusionado en `main` como ejemplo permanente (rechazada: dejaría una tabla de práctica en la BD y una área ficticia que F2 tendría que borrar; alternativa registrada como duda menor para el humano); verificar la variante de panel en F1 (imposible: `RegisterAdmin` con permisos es F2 — el límite queda declarado en el plan).

## R19. Configuración, logs y inyección de dependencias (D-A6) — *nuevo 2026-09-30 (D12)*

- **Decisión**:
  - **Config**: `os.Getenv` + validación al arrancar en `internal/platform/config` (`Load() (Config, error)`), con valores por defecto de desarrollo para que `make up` funcione en un clon limpio. Falla rápido y con mensaje claro si algo obligatorio falta o es inválido.
  - **Logs**: `log/slog` con handler **JSON**, nivel desde `LOG_LEVEL`, y **logger por petición** (hij con `request_id`, método y ruta, creado por `middleware/request-id`).
  - **Errores**: stdlib (`errors.Is/As`, `fmt.Errorf("…: %w", err)`) más los errores tipados de `internal/platform/apperr`, traducidos a HTTP solo en `httpserver.WriteError` (R17).
  - **DI**: composición **manual** desde `cmd/api/main.go` (repository → service → handler → rutas), sin `wire` ni `fx`; sin estado global ni `init()` con lógica.
- **Justificación**: §II (minimizar dependencias), §V (errores con contexto) y §VII (logs estructurados, config por entorno). `main.go` queda como mapa de dependencias del sistema: se lee entero y se audita de un vistazo. Los fallos de configuración se detectan al arrancar, no en caliente.
- **Alternativas consideradas**: Viper (17 entradas en `go.mod`) o `envconfig` (una dependencia) para leer variables (rechazadas: son ~20 líneas de `os.Getenv` + validación); `zerolog`/`zap` para logs (rechazadas: `slog` está en la stdlib desde Go 1.21 y es la que pide la constitución); librerías de errores ricos tipo `pkg/errors` (rechazadas: `%w` basta); `wire`/`fx` para DI (rechazadas: generación de código y runtime para un grafo de ~4 nodos); `log.Println` sin estructura (rechazada: §VII).

## R20. Control de deriva de los artefactos generados — *nuevo 2026-09-30 (D6/D15 del plan)*

- **Decisión**: dos artefactos generados se **commitean** y se regeneran a mano con herramientas cuya versión se fija en el README: `backend/internal/db/` (`make sqlc-gen` → `sqlc generate`) y `frontend/src/api/schema.d.ts` (`make api-gen` → `npm run api:gen`). Para controlar la deriva: `make sqlc-verify` regenera y exige `git diff --exit-code`; regla de revisión (un PR que toca `migrations/` o `queries/` debe tocar `internal/db/`; uno que toca el contrato debe tocar `schema.d.ts`); y **propuesta al repo del kit** (regla 10: no se edita `ci.yml` aquí) de un paso de CI que verifique ambos.
- **Justificación**: el `ci.yml` del kit no corre estas herramientas y no se puede editar; commitear el código generado mantiene el CI sin pasos extra y hace que un cambio de SQL sin regenerar **falle en compilación**. El riesgo real (alguien edita una consulta y olvida `sqlc generate`) se cubre con el target de verificación, la regla de revisión y la versión fijada (generación reproducible entre máquinas).
- **Alternativas consideradas**: generar en CI (rechazada: exigiría editar el `ci.yml` del kit); no commitear el código generado (rechazada: el CI no tendría las herramientas y cada build las instalaría); git hook propio de verificación (rechazada: `.githooks/` es del kit y no se edita); confiar solo en la disciplina del equipo (rechazada: es exactamente la deriva que SC-006/§III no pueden permitirse).
