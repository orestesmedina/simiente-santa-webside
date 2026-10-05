# Simiente Santa

Sitio web de la Iglesia Simiente Santa: sitio público (eventos, grupos, ministerios, donaciones) y panel de administración. El detalle de lo que se construye, fase por fase, está en [`docs/producto/roadmap.md`](docs/producto/roadmap.md).

## Qué es la F1 (Estructura base)

La funcionalidad **F1** es el esqueleto técnico sobre el que se construyen F2–F9:

- **Backend en Go** con `GET /healthz`, que informa en cada consulta el estado **real** de la conexión a PostgreSQL.
- **Frontend en React + TypeScript** con una página inicial que muestra ese estado de forma comprensible.
- **PostgreSQL** con migraciones versionadas (`golang-migrate`) — en F1 sin tablas de negocio: solo la baseline `000001` (no-op).
- **Docker Compose** para levantar todo con un solo comando y **CI** (GitHub Actions) que valida cada cambio.
- Contrato vivo en [`backend/api/openapi.yaml`](backend/api/openapi.yaml), del que se generan los tipos TypeScript del frontend.

Rama: `001-estructura-base`. Spec, plan y tareas en [`specs/001-estructura-base/`](specs/001-estructura-base/).

## Qué es F2 (Acceso y gestión de usuarios)

La funcionalidad **F2** abre el **panel de administración** que usarán F3–F9:

- **Inicio de sesión** en <http://localhost:5173/login> con correo y contraseña. Sin sesión, cualquier sección del panel redirige a la pantalla de acceso. La sesión dura como máximo **1 hora** desde el inicio y se cierra a los **30 minutos de inactividad**.
- **Puesta en marcha con una inicialización única**: `POST /api/v1/setup/initialize` con la cabecera `X-Setup-Token` (`BOOTSTRAP_TOKEN`) crea el primer administrador; no se puede repetir (ver «Puesta en marcha del panel» más abajo).
- **Gestión de usuarios**: crear, editar, activar/desactivar y restablecer la contraseña. Quien recibe una contraseña de un administrador debe cambiarla al entrar. Las cuentas **no se eliminan**: solo se desactivan.
- **Roles con permisos por módulo** (sin catálogo fijo de roles), verificados en el servidor en cada operación; los cambios de un rol se reflejan de inmediato en sus cuentas.
- **Auditoría de solo lectura** en `/panel/auditoria`: último acceso de cada cuenta, historial de inicios de sesión (exitosos y fallidos, con IP de origen) e historial de acciones administrativas, con filtros por cuenta y rango de fechas.
- **Redis** (servicio nuevo en compose) guarda la sesión y los contadores de intentos. Corre **sin persistencia a propósito**: un reinicio obliga a volver a entrar y pone a cero los contadores, pero no toca la auditoría, que vive en PostgreSQL.

Rama: `002-acceso-gestion-usuarios`. Spec, plan y quickstart en [`specs/002-acceso-gestion-usuarios/`](specs/002-acceso-gestion-usuarios/).

## Prerrequisito

**Docker** (con Docker Compose v2) en ejecución. Es el único requisito (FR-001 de la spec de F1): el clon limpio levanta **sin `.env`** y sin tener Go ni Node instalados.

Una vez por clon, activa los hooks de git (plan R3):

```bash
make instalar-hooks
```

## Quickstart

```bash
git clone <repo>
cd simiente_santa
make instalar-hooks   # solo la primera vez por clon
make up               # docker compose up -d: db + redis + backend + frontend
```

Comprobar:

```bash
docker compose ps                      # db y redis (healthy), backend y frontend en ejecución
curl -i http://localhost:8080/healthz  # 200 → {"status":"ok","database":"connected"}
```

Abrir <http://localhost:5173>: la página muestra el estado del sistema sin ninguna acción adicional. El **panel de administración** está en <http://localhost:5173/login> y todavía no tiene cuentas: hay que inicializarlo (siguiente sección).

Detener: `make down` (los datos se conservan en el volumen `pgdata`; `make down && make up` es repetible).

### Puesta en marcha del panel (solo la primera vez)

El primer administrador se crea con una **inicialización única** por API: solo funciona si el sistema no tiene ninguna cuenta y solo con el token `BOOTSTRAP_TOKEN` (FR-007).

1. Crea el `.env` y pon ahí un token (genera uno propio con `openssl rand -base64 32`; en local sirve un valor de desarrollo como el que usan las pruebas e2e, `dev-bootstrap-token`):

   ```bash
   cp .env.example .env      # y editar BOOTSTRAP_TOKEN
   make up                   # compose relee el .env y recrea el backend
   ```

2. Ejecuta la inicialización (datos de ejemplo; la contraseña cumple la política):

   ```bash
   curl -i -X POST http://localhost:8080/api/v1/setup/initialize \
     -H "Content-Type: application/json" \
     -H "X-Setup-Token: $BOOTSTRAP_TOKEN" \
     -d '{"firstName":"Ana","lastName":"Responsable","email":"ana@ejemplo.com","phone":"+34 612 345 678","password":"Semilla.2026"}'
   ```

   Esperado: `201` con la cuenta creada (rol «Administrador», los 9 permisos). Repetirlo → `409` «La inicialización ya se hizo y no puede repetirse»; sin la cabecera o con el token equivocado → `401`/`403`. **Sin `BOOTSTRAP_TOKEN` la inicialización queda desactivada** (siempre `401`/`403`).

3. Entra en <http://localhost:5173/login> con ese correo y contraseña. La validación paso a paso está en [`specs/002-acceso-gestion-usuarios/quickstart.md`](specs/002-acceso-gestion-usuarios/quickstart.md) §1–§3.

### Puertos

| Servicio | Puerto en el host | Puerto dentro del contenedor | Variable |
|---|---|---|---|
| Frontend (nginx) | **5173** | 80 | `WEB_PORT` |
| Backend (API) | **8080** | 8080 (**fijo**) | `HTTP_PORT` |
| PostgreSQL | **5432** | 5432 | `DB_PORT` |
| Redis (sesión y contadores) | **6379** | 6379 | `REDIS_PORT` |

El backend **siempre escucha en el 8080 dentro de su contenedor**; `HTTP_PORT` solo parametriza el puerto publicado en el host (`"${HTTP_PORT:-8080}:8080"` en `docker-compose.yml`).

Con puertos propios (útil si alguno está ocupado):

```bash
DB_PORT=5433 HTTP_PORT=8081 WEB_PORT=5174 make up
```

**Al cambiar puertos hay que reconstruir** (ver «Notas operativas»: `VITE_API_URL` se hornea en build):

```bash
HTTP_PORT=8081 WEB_PORT=5174 CORS_ALLOWED_ORIGINS=http://localhost:5174 docker compose up -d --build
```

Dos detalles al cambiar puertos:

- **`WEB_PORT`**: hay que añadir el origen nuevo a `CORS_ALLOWED_ORIGINS` (por defecto `http://localhost:5173`); si no, el navegador bloquea las llamadas del frontend a la API.
- **`HTTP_PORT`**: compose deriva de él el `VITE_API_URL` con el que se compila la imagen del frontend; sin `--build` esa imagen sigue apuntando al puerto anterior.

### `.env` opcional

```bash
cp .env.example .env   # editar valores; .env está en .gitignore
make up
```

Todas las variables tienen valor por defecto de desarrollo; ninguna es obligatoria para levantar (FR-001 de la spec de F1). Lo único que exige tocar `.env` es **inicializar el panel**, que necesita `BOOTSTRAP_TOKEN` (ver «Puesta en marcha del panel»).

| Variable | Quién la lee | Por defecto |
|---|---|---|
| `APP_ENV` | backend (`platform/config`) | `development` |
| `HTTP_PORT` | backend (local) / publicación del puerto en el host | `8080` |
| `DATABASE_URL` | backend y `make db-migrate` (desde el host apunta a `localhost`) | `postgres://app:app_dev_password@localhost:5432/app?sslmode=disable` |
| `LOG_LEVEL` | backend (`debug`/`info`/`warn`/`error`) | `info` |
| `CORS_ALLOWED_ORIGINS` | backend (middleware CORS) | `http://localhost:5173` |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | servicio `db` de compose (y con ellos compose construye el `DATABASE_URL` **del contenedor**, apuntando al host `db`) | `app` / `app_dev_password` / `app` |
| `DB_PORT` / `REDIS_PORT` / `WEB_PORT` | compose (puertos del host) | `5432` / `6379` / `5173` |
| `DATABASE_URL_TEST` | pruebas de integración del backend (`-tags=integration`); sin ella, esos tests **se omiten en silencio** | la trae `.env.example` apuntando a `app_test` en `localhost:5432`; que apunte **solo a una BD efímera**: esas pruebas vacían las tablas del dominio y arruinarían el stack local |
| `VITE_API_URL` | build de la imagen del frontend (URL de la API que usa el navegador) | `http://localhost:8080` |
| `REDIS_URL` | backend (sesión en Redis y contadores de intentos); compose la fija a `redis://redis:6379/0` dentro de la red | `redis://localhost:6379/0` (solo fuera de Docker: tests y scripts) |
| `SESSION_SECRET` | backend (`platform/config`): firma HMAC de la cookie CSRF | valor de ejemplo (≥ 32 caracteres); **obligatoria en producción** (el arranque falla si está vacía o corta) |
| `SESSION_COOKIE_SECURE` | backend (`Secure` en la cookie de sesión) | `false` (con HTTPS → `true`; en producción el arranque falla con `false`) |
| `SESSION_IDLE_TTL_MINUTES` / `SESSION_ABSOLUTE_TTL_MINUTES` | backend (TTL de la sesión en Redis) | `30` (inactividad) / `60` (vida absoluta desde el login) |
| `BOOTSTRAP_TOKEN` | backend: cabecera `X-Setup-Token` de `POST /api/v1/setup/initialize` (inicialización única) | vacío → la inicialización queda desactivada (la ruta siempre responde `401`/`403` y el backend avisa al arrancar) |

## Comandos

| Comando | Qué hace |
|---|---|
| `make up` | Levanta `db` + `redis` + `backend` + `frontend` (`docker compose up -d`, con build de las imágenes si faltan; también recrea el backend si cambió `.env`) |
| `make down` | Detiene el entorno conservando los datos (`pgdata`) |
| `make test` | Pruebas de backend (`go test ./...` + `-tags=integration`) y frontend (Vitest) |
| `make lint` | `gofmt` + `go vet` + `golangci-lint` (backend) y ESLint + `tsc` (frontend) |
| `make security` | `govulncheck ./...` (Go) y `npm audit --audit-level=high` (frontend) |
| `make ci` | `lint` + `test` + `security`: lo mismo que corre el CI |
| `make db-migrate` | Aplica las migraciones de `backend/migrations` con `golang-migrate` (F2 añade `000002`–`000004`) |
| `make doctor` | Verifica que el entorno tenga todo lo necesario (Docker, Go, Node, hooks, kit…) |
| `make e2e` | Pruebas end-to-end con Playwright (requiere `make up` levantado) |
| `make api-gen` | Regenera `frontend/src/api/schema.d.ts` desde `backend/api/openapi.yaml` |
| `make sqlc-gen` | Regenera el código Go de consultas de `backend/internal/db/` |
| `make sqlc-verify` | Regenera y exige `git diff --exit-code` (sin deriva de artefactos) |
| `make instalar-hooks` | **Una vez por clon**: activa los hooks de git (`.githooks`) |
| `make help` | Lista los comandos disponibles |
| `make estado` | Por dónde va el proyecto: roadmap, fase, aprobaciones y próximo paso |
| `make costos` | Costo de IA de la tarea actual (`CERRAR=1` lo congela al aprobar el PR; `TODO=1` resume el proyecto) |
| `make novedades` | Historial de cambios del kit (`DESDE=1.4.0` para ver desde una versión) |

Notas:

- **`make db-migrate` usa `$DATABASE_URL` del shell** (no lee `.env` por sí mismo). Con `.env`: `set -a; source .env; set +a` antes de llamarlo; sin `.env`: `DATABASE_URL='postgres://app:app_dev_password@localhost:5432/app?sslmode=disable' make db-migrate`.
- **`make sqlc-verify` comprueba que no hay deriva** entre `backend/internal/db/queries/` y el código generado: F2 añadió las consultas `users`, `roles`, `permissions` y `audit` (en F1 estaba vacío y el comando terminaba en verde sin comprobar nada).

## Herramientas de desarrollo (opcionales)

Solo hace falta instalarlas para trabajar **fuera** de Docker (tests, lint, generación). Verificables con `make doctor`. Versiones fijadas para generación reproducible (plan R4):

| Herramienta | Versión fijada | Dónde se fija / cómo se obtiene |
|---|---|---|
| Go | **1.27** | `go 1.27` en `backend/go.mod`, `golang:1.27` en `backend/Dockerfile`; el CI lee `backend/go.mod` |
| Node.js | **24** | `NODE_VERSION: '24'` en `.github/workflows/ci.yml` (kit 1.6.4); `node:24.21-alpine` en `frontend/Dockerfile` |
| `golang-migrate` | **v4.20.1** | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1` (mismo pin que el CI del kit 1.6.4) |
| `sqlc` | **v1.31.1** | instalado con `go install …@v1.31.1`; el binario queda en `$(go env GOPATH)/bin`, que debe estar en el `PATH` |
| `openapi-typescript` | `^7.13.0` | `devDependency` de `frontend/package.json`, fijada en `package-lock.json` (`npm ci`) |
| Playwright | `@playwright/test` `^1.63.0` | `devDependency` de `frontend/package.json` |

Otras herramientas del flujo de calidad: `golangci-lint` **v2.14.0** y `govulncheck` **v1.8.0** (los usan `make lint` y `make security`; el CI del kit 1.6.4 fija esas mismas versiones) y `gitleaks` (hooks de git y el paso de secretos del CI).

Para `make e2e` hace falta además el **navegador** de Playwright:

```bash
cd frontend
npx playwright install chromium
sudo npx playwright install-deps chromium   # solo Linux: librerías del sistema
```

## Versiones y soporte de seguridad (FR-016 / SC-010)

Comprobado a **2026-10-05** (Redis se añadió en F2; el resto de la tabla, a 2026-10-03):

| Tecnología | Versión en uso | Estado de soporte |
|---|---|---|
| Go | 1.27 | Vigente (solo 1.26 y 1.27 reciben parches; 1.23 está en fin de vida desde 2025-08-12) |
| Node.js | 24 (LTS «Krypton») | **LTS activa** (mantenimiento desde 2026-10-20; fin de soporte 2028-04-30); el kit 1.6.4 fija esta línea mayor en el CI |
| PostgreSQL | 16 (imagen `postgres:16-alpine` en desarrollo — **decisión del 2026-10-03**, T034/plan R10) | Rama 16 en soporte hasta noviembre de 2028 y recibe parches de seguridad; fijar un minor/digest exacto para reproducibilidad queda pendiente para builds/despliegue |
| Redis | 7 (imagen `redis:7-alpine`, servicio `redis` en `docker-compose.yml`; añadida en F2, T205 — decisión P23 del plan) | Series 7.2 y 7.4 de Redis OSS en soporte extendido hasta **2029-12-01** (ciclo oficial de versiones de Redis, consultado el 2026-10-05); fijar un minor/digest exacto queda pendiente, igual que con `postgres:16-alpine` |
| React / TypeScript / Vite | 19 / 5.x / actual | Mantenidas; vulnerabilidades vía `npm audit` |
| `pgx/v5`, `sqlc`, `golang-migrate`, `openapi-typescript` | fijadas en `go.sum`, `package-lock.json` y este README | Herramientas de desarrollo; `govulncheck` y `npm audit` sin altas/críticas |

**Pendientes de actualización conocidos** (US7 esc. 2; identificados, propuestos y gestionados por el humano):

- **Resueltos por el kit 1.6.4** (`2957bb0`, actualizado el 2026-10-03): `postgres:16-alpine` en el CI (R10/T034), Node 24 en el CI (R11), `golangci-lint` v2 en el CI (g1) y «Go 1.26+» en `docs/GUIA-INICIO.md`. Ya no son pendientes del proyecto.
- **Sin pendientes abiertos** por el momento. Queda para builds/despliegue fijar un **minor/digest exacto** de las imágenes (FR-016).

Los archivos del kit no se editan aquí (regla 10). **Política (FR-016)**: toda funcionalidad que fije o actualice una versión revisa esta tabla; lo que deje de estar en soporte se registra como pendiente antes de seguir construyendo.

## Notas operativas

- **`VITE_API_URL` se hornea en el build** (plan R6): la URL de la API queda dentro del bundle del frontend. Por defecto `http://localhost:8080` y compose la deriva de `HTTP_PORT`. Por eso **cambiar puertos exige `docker compose up -d --build`** (o borrar la imagen del frontend), y al cambiar `WEB_PORT` hay que ajustar también `CORS_ALLOWED_ORIGINS`.
- **Artefactos generados (se commitean)**: `frontend/src/api/schema.d.ts` se regenera con `make api-gen` cuando cambia `backend/api/openapi.yaml`; el código de sqlc con `make sqlc-gen` cuando cambian `backend/migrations/` o `backend/internal/db/queries/`. Regla de revisión (plan R4): un PR que toca migraciones o consultas debe regenerar `internal/db/`, y uno que toca el contrato debe regenerar `schema.d.ts`. `make sqlc-verify` comprueba que no hay deriva.
- **E2E**: `make e2e` necesita el stack levantado (`make up`) y el navegador de Playwright instalado (arriba). El CI del kit **no** corre e2e. Dos notas de F2: si en tu entorno faltan las librerías de sistema de Chromium (`libnspr4`, `libnss3`, `libasound2`) y no puedes instalarlas, las suites se pueden correr con la imagen oficial de Playwright (`mcr.microsoft.com/playwright:v1.63.0-jammy`) contra el stack levantado — así se validó F2 (3/3); y conviene ejecutarlas **en serie** (`--workers=1`), porque en paralelo saturan el rate-limit de 20 peticiones/min por IP a `/api/v1/auth/login` y producen `429` espurios.

## Documentación relacionada

- [`CHANGELOG.md`](CHANGELOG.md) — cambios notables por versión (Keep a Changelog + SemVer): la 0.1.0 corresponde a F1 y la 0.2.0 a F2.
- [`docs/entrega/F1-estructura-base.md`](docs/entrega/F1-estructura-base.md) — notas de entrega de F1 para el cliente (lenguaje no técnico).
- [`docs/entrega/F2-acceso-y-gestion-de-usuarios.md`](docs/entrega/F2-acceso-y-gestion-de-usuarios.md) — notas de entrega de F2 para el cliente (lenguaje no técnico).
- [`specs/001-estructura-base/quickstart.md`](specs/001-estructura-base/quickstart.md) — validación ejecutable de F1 de punta a punta (qué se espera en cada escenario).
- [`specs/002-acceso-gestion-usuarios/quickstart.md`](specs/002-acceso-gestion-usuarios/quickstart.md) — validación ejecutable de F2 (inicialización, sesión, usuarios, roles y auditoría).
- [`docs/tecnico/arquitectura.md`](docs/tecnico/arquitectura.md) — arquitectura del sistema; **§8 es la receta** de 10 pasos para agregar un área de negocio nueva.
- [`docs/tecnico/decisiones.md`](docs/tecnico/decisiones.md) — decisiones de arquitectura y su justificación.
- [`docs/GUIA-INICIO.md`](docs/GUIA-INICIO.md) — guía del kit: agentes de IA, entorno WSL y problemas comunes.
- [`docs/producto/roadmap.md`](docs/producto/roadmap.md) — funcionalidades F1–F9 y su estado.
- [`.specify/memory/constitution.md`](.specify/memory/constitution.md) — reglas no negociables del proyecto.
