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

## Prerrequisito

**Docker** (con Docker Compose v2) en ejecución. Es el único requisito (FR-001): el clon limpio levanta **sin `.env`** y sin tener Go ni Node instalados.

Una vez por clon, activa los hooks de git (plan R3):

```bash
make instalar-hooks
```

## Quickstart

```bash
git clone <repo>
cd simiente_santa
make instalar-hooks   # solo la primera vez por clon
make up               # docker compose up -d: db + backend + frontend
```

Comprobar:

```bash
docker compose ps                      # db (healthy), backend y frontend en ejecución
curl -i http://localhost:8080/healthz  # 200 → {"status":"ok","database":"connected"}
```

Abrir <http://localhost:5173>: la página muestra el estado del sistema sin ninguna acción adicional.

Detener: `make down` (los datos se conservan en el volumen `pgdata`; `make down && make up` es repetible).

### Puertos

| Servicio | Puerto en el host | Puerto dentro del contenedor | Variable |
|---|---|---|---|
| Frontend (nginx) | **5173** | 80 | `WEB_PORT` |
| Backend (API) | **8080** | 8080 (**fijo**) | `HTTP_PORT` |
| PostgreSQL | **5432** | 5432 | `DB_PORT` |

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

Todas las variables tienen valor por defecto de desarrollo; ninguna es obligatoria para levantar (FR-001).

| Variable | Quién la lee | Por defecto |
|---|---|---|
| `APP_ENV` | backend (`platform/config`) | `development` |
| `HTTP_PORT` | backend (local) / publicación del puerto en el host | `8080` |
| `DATABASE_URL` | backend y `make db-migrate` (desde el host apunta a `localhost`) | `postgres://app:app_dev_password@localhost:5432/app?sslmode=disable` |
| `LOG_LEVEL` | backend (`debug`/`info`/`warn`/`error`) | `info` |
| `CORS_ALLOWED_ORIGINS` | backend (middleware CORS) | `http://localhost:5173` |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | servicio `db` de compose (y con ellos compose construye el `DATABASE_URL` **del contenedor**, apuntando al host `db`) | `app` / `app_dev_password` / `app` |
| `DB_PORT` / `WEB_PORT` | compose (puertos del host) | `5432` / `5173` |
| `DATABASE_URL_TEST` | pruebas de integración del backend (`-tags=integration`); sin ella, esos tests se omiten | la trae `.env.example` apuntando a `app_test` en `localhost:5432` |
| `VITE_API_URL` | build de la imagen del frontend (URL de la API que usa el navegador) | `http://localhost:8080` |
| `SESSION_SECRET` | **F2** (sesiones): el backend de F1 no la consume | — |

## Comandos

| Comando | Qué hace |
|---|---|
| `make up` | Levanta `db` + `backend` + `frontend` (`docker compose up -d`, con build de las imágenes si faltan) |
| `make down` | Detiene el entorno conservando los datos (`pgdata`) |
| `make test` | Pruebas de backend (`go test ./...` + `-tags=integration`) y frontend (Vitest) |
| `make lint` | `gofmt` + `go vet` + `golangci-lint` (backend) y ESLint + `tsc` (frontend) |
| `make security` | `govulncheck ./...` (Go) y `npm audit --audit-level=high` (frontend) |
| `make ci` | `lint` + `test` + `security`: lo mismo que corre el CI |
| `make db-migrate` | Aplica las migraciones de `backend/migrations` con `golang-migrate` |
| `make doctor` | Verifica que el entorno tenga todo lo necesario (Docker, Go, Node, hooks, kit…) |
| `make e2e` | Pruebas end-to-end con Playwright (requiere `make up` levantado) |
| `make api-gen` | Regenera `frontend/src/api/schema.d.ts` desde `backend/api/openapi.yaml` |
| `make sqlc-gen` | Regenera el código Go de consultas de `backend/internal/db/` |
| `make sqlc-verify` | Regenera y exige `git diff --exit-code` (sin deriva de artefactos) |
| `make instalar-hooks` | **Una vez por clon**: activa los hooks de git (`.githooks`) |
| `make help` | Lista los comandos disponibles (ver nota sobre `e2e`) |
| `make estado` | Por dónde va el proyecto: roadmap, fase, aprobaciones y próximo paso |

Notas:

- **`make db-migrate` usa `$DATABASE_URL` del shell** (no lee `.env` por sí mismo). Con `.env`: `set -a; source .env; set +a` antes de llamarlo; sin `.env`: `DATABASE_URL='postgres://app:app_dev_password@localhost:5432/app?sslmode=disable' make db-migrate`.
- **`make help` no lista `e2e`**: el recetario `help` del `Makefile` del kit usa un regex cuya clase (`[a-zA-Z_-]`) no incluye dígitos. Es un pendiente del **kit** (no editable aquí: regla 10, T035); el target **funciona igual**.
- En F1 `make sqlc-verify` termina en verde avisando que no hay consultas: `backend/internal/db/queries/` está vacío hasta la primera consulta de negocio (R5 del plan).

## Herramientas de desarrollo (opcionales)

Solo hace falta instalarlas para trabajar **fuera** de Docker (tests, lint, generación). Verificables con `make doctor`. Versiones fijadas para generación reproducible (plan R4):

| Herramienta | Versión fijada | Dónde se fija / cómo se obtiene |
|---|---|---|
| Go | **1.27** | `go 1.27` en `backend/go.mod`, `golang:1.27` en `backend/Dockerfile`; el CI lee `backend/go.mod` |
| Node.js | **22** | `node-version: 22` en `.github/workflows/ci.yml` (kit); `node:22.22-alpine` en `frontend/Dockerfile` |
| `golang-migrate` | CLI con `-tags postgres` | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest` (mismo comando que el CI); **el proyecto no fija su versión** |
| `sqlc` | **v1.31.1** | instalado con `go install …@v1.31.1`; el binario queda en `$(go env GOPATH)/bin`, que debe estar en el `PATH` |
| `openapi-typescript` | `^7.13.0` | `devDependency` de `frontend/package.json`, fijada en `package-lock.json` (`npm ci`) |
| Playwright | `@playwright/test` `^1.63.0` | `devDependency` de `frontend/package.json` |

Otras herramientas del flujo de calidad: `golangci-lint` y `govulncheck` (los usan `make lint` y `make security`) y `gitleaks` (hooks de git y el paso de secretos del CI).

Para `make e2e` hace falta además el **navegador** de Playwright:

```bash
cd frontend
npx playwright install chromium
sudo npx playwright install-deps chromium   # solo Linux: librerías del sistema
```

## Versiones y soporte de seguridad (FR-016 / SC-010)

Comprobado a **2026-09-30**:

| Tecnología | Versión en uso | Estado de soporte |
|---|---|---|
| Go | 1.27 | Vigente (solo 1.26 y 1.27 reciben parches; 1.23 está en fin de vida desde 2025-08-12) |
| Node.js | 22 | En mantenimiento hasta **abril de 2027** |
| PostgreSQL | 16 (imagen `postgres:16.4-alpine`) | Rama 16 en soporte hasta noviembre de 2028; **el minor 16.4 acumula CVEs corregidos en minors posteriores** |
| React / TypeScript / Vite | 19 / 5.x / actual | Mantenidas; vulnerabilidades vía `npm audit` |
| `pgx/v5`, `sqlc`, `golang-migrate`, `openapi-typescript` | fijadas en `go.sum`, `package-lock.json` y este README | Herramientas de desarrollo; `govulncheck` y `npm audit` sin altas/críticas |

**Pendientes de actualización conocidos** (US7 esc. 2; identificados, propuestos y gestionados por el humano, plan R10/R11 → T034/T035):

1. **`postgres:16.4-alpine` del servicio `db`** en `docker-compose.yml` (R10, decisión en T034): propuesta `postgres:16-alpine`.
2. **`postgres:16.4-alpine` del servicio `postgres` del `ci.yml` del kit**: mismo caso; pertenece al kit, se propone a su repositorio (T035).
3. **Node 22**, fijado por el `ci.yml` del kit (R11): soporte hasta abril de 2027 → proponer al kit una LTS vigente antes de esa fecha (T035).
4. **«Go 1.23+»** que aún recomienda `docs/GUIA-INICIO.md` del kit (Go 1.23, fin de vida desde 2025-08-12): el proyecto usa Go 1.27 → propuesta al repo del kit (T035).

Los archivos del kit no se editan aquí (regla 10). **Política (FR-016)**: toda funcionalidad que fije o actualice una versión revisa esta tabla; lo que deje de estar en soporte se registra como pendiente antes de seguir construyendo.

## Notas operativas

- **`VITE_API_URL` se hornea en el build** (plan R6): la URL de la API queda dentro del bundle del frontend. Por defecto `http://localhost:8080` y compose la deriva de `HTTP_PORT`. Por eso **cambiar puertos exige `docker compose up -d --build`** (o borrar la imagen del frontend), y al cambiar `WEB_PORT` hay que ajustar también `CORS_ALLOWED_ORIGINS`.
- **Artefactos generados (se commitean)**: `frontend/src/api/schema.d.ts` se regenera con `make api-gen` cuando cambia `backend/api/openapi.yaml`; el código de sqlc con `make sqlc-gen` cuando cambian `backend/migrations/` o `backend/internal/db/queries/`. Regla de revisión (plan R4): un PR que toca migraciones o consultas debe regenerar `internal/db/`, y uno que toca el contrato debe regenerar `schema.d.ts`. `make sqlc-verify` comprueba que no hay deriva.
- **E2E**: `make e2e` necesita el stack levantado (`make up`) y el navegador de Playwright instalado (arriba). El CI del kit **no** corre e2e.
- **`make e2e` no aparece en `make help`** (bug del regex del Makefile del kit, pendiente T035): ejecútalo directamente.

## Documentación relacionada

- [`specs/001-estructura-base/quickstart.md`](specs/001-estructura-base/quickstart.md) — validación ejecutable de F1 de punta a punta (qué se espera en cada escenario).
- [`docs/tecnico/arquitectura.md`](docs/tecnico/arquitectura.md) — arquitectura del sistema; **§8 es la receta** de 10 pasos para agregar un área de negocio nueva.
- [`docs/tecnico/decisiones.md`](docs/tecnico/decisiones.md) — decisiones de arquitectura y su justificación.
- [`docs/GUIA-INICIO.md`](docs/GUIA-INICIO.md) — guía del kit: agentes de IA, entorno WSL y problemas comunes.
- [`docs/producto/roadmap.md`](docs/producto/roadmap.md) — funcionalidades F1–F9 y su estado.
- [`.specify/memory/constitution.md`](.specify/memory/constitution.md) — reglas no negociables del proyecto.
