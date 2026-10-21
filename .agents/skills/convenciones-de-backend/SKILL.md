---
name: convenciones-de-backend
description: Convenciones propias de la parte backend (nombres, estructura, patrones, base de datos, bibliotecas). Usar siempre que se cree o modifique código en backend/.
---
# Convenciones de backend

Redactadas el 2026-10-20. Estándar de referencia: skills `go-backend` y `postgres-db`; lo que no diga, el estándar. Rutas relativas a `backend/`.

## Nombres
- El código nuevo se nombra en español (decisión del 2026-10-20): `MetodoDonacion` (`donaciones/model.go:97`), `NormalizarIban` (`donaciones/numero.go:59`), `Episodio` (`medios/model.go:182`). Lo que ya está en inglés no se renombra sin una tarea propia: `User` (`usuarios/model.go:31`).
- Cada archivo empieza con `// <archivo>.go <qué hace>`; el doc del paquete va en `model.go` (`donaciones/repository.go:21`, `model.go:1`).
- Archivos por capa: `service*.go`, `repository*.go`, `handler*.go`, `routes.go`, `model.go`, `optional.go`, con sufijo `_public`/`_admin`/`_audit` (`donaciones/service.go:30`).
- Tipos: entidad del dominio, `Input`/`Patch` para entrada, `Publico`/`Admin` para salida (`donaciones/model.go:97`).

## Estructura
- `cmd/api/main.go` solo compone dependencias y arranca, sin estado global (`cmd/api/main.go:337`).
- Un paquete por dominio, capas handler → service → repository; el service no ve `net/http` ni pgx (`donaciones/service.go:27`).
- Lo transversal en `platform/*`, que no conoce dominios (`platform/database/database.go:5`).
- sqlc genera `internal/db` (alias `gendb`) desde `internal/db/queries/*.sql` (`donaciones/repository.go:15`).
- Pruebas en el MISMO paquete; las de integración en `*_integration_test.go` con `//go:build integration` (`donaciones/repository_donaciones_integration_test.go:1`).
- Contrato HTTP en `api/openapi.yaml`, OpenAPI 3.1, `operationId` camelCase (`api/openapi.yaml:490`).

## Patrones
- Interfaces en el paquete que las consume, con `var _ X = (*x)(nil)` (`donaciones/service.go:485`).
- Constructores `NewX(deps XDeps)`; `Logger` nil usa `slog.Default()` (`donaciones/service.go:133`).
- Handler fino: `decodeAndValidate` → delegar → `WriteJSON`/`WriteError`; nunca comprueba permisos (`donaciones/handler_admin.go:202`).
- Errores de dominio `apperr` (mensaje seguro + `details` por campo); `httpserver.WriteError` traduce a HTTP en un solo punto (`platform/httpserver/error.go:85`).
- Mensajes de `apperr.Message` y `details` con mayúscula inicial y punto final (decisión del 2026-10-20): `details["number"] = "Este campo es obligatorio."` (`donaciones/service.go:246`).
- Éxito: DTO directo, sin sobre. Error: `{"error":{code,message,details?}}` (`platform/httpserver/error.go:54`).
- `PATCH` con `Optional[T]` tri-estado (ausente, valor, null): `Kind Optional[string]` (`donaciones/model.go:153`). `OptionalNull()` solo al construir el PATCH en código (`donaciones/optional.go:36`); los punteros `*string` de medios no se repiten.
- Validación en dos capas: etiquetas `validate` del DTO y reglas de dominio sobre lo ya fusionado; la transacción escribe la fila completa (`donaciones/service_admin.go:328`).
- SQLSTATE con `classify`/`wrap`: 23505→409, 23514→400 con detalle por constraint (`donaciones/repository.go:258`).
- Mutaciones con sus filas de auditoría en la MISMA transacción; lo demás, `RecordAction` best-effort (`donaciones/service.go:96`).
- Rutas en `routes.go`: `RegisterPublic` y `RegisterAdmin` con `middleware.AdminChain(PermissionX, …)` (`donaciones/routes.go:18`).
- Textos de entrada normalizados: trim y colapsar espacios; lo opcional vacío se guarda NULL (`donaciones/service.go:149`).

## Base de datos
- Tablas en inglés, plural, snake_case; PK `id UUID DEFAULT gen_random_uuid()`; FK `<tabla singular>_id` con `ON DELETE` decidido y su índice (`migrations/000003_create_users.up.sql:17`).
- Texto: `TEXT` + `CHECK (char_length(…) BETWEEN 1 AND n)`; catálogos y estados: `TEXT NOT NULL DEFAULT '…' CHECK (col IN (…))` (`migrations/000012_create_donation_content.up.sql:31`).
- `created_at`/`updated_at` `TIMESTAMPTZ NOT NULL DEFAULT now()`; el `UPDATE` fija `updated_at = now()` (`db/queries/donation_methods.sql:129`).
- Índices `<tabla>_<columnas>_idx` en el orden del `ORDER BY`; parciales cuando aplican (`migrations/000003_create_users.up.sql:32`).
- Cambios en parejas `migrations/NNNNNN_snake_case.up.sql`/`.down.sql`; migrada, NUNCA se edita (`migrations/000012_create_donation_content.up.sql:1`).
- Consultas solo en `internal/db/queries/*.sql`: columnas explícitas, `sqlc.arg`/`sqlc.narg`, `:one`/`:many`/`:execrows` (`db/queries/donation_methods.sql:68`).
- Filtro opcional `(sqlc.narg('x')::text IS NULL OR col = sqlc.narg('x'))`; fechas en semirango `>=`/`<` (`db/queries/audit.sql:43`).
- Cada listado lleva su `Count…` con el MISMO filtro; `ORDER BY` determinista con desempate por `id` (`db/queries/donation_methods.sql:72`).
- Las públicas solo `publication_state = 'published'` y con `LIMIT` del servidor (`db/queries/donation_methods.sql:45`).

## Bibliotecas y herramientas
- `pgx/v5` + `pgxpool`: pool compartido y perezoso (`platform/database/database.go:26`).
- `github.com/google/uuid` para ids: el dominio usa `uuid.UUID` y el repository traduce `pgtype.*` (`donaciones/repository.go:98`).
- `go-redis/v9` para sesión y contadores (`cmd/api/main.go:60`); `bcrypt` con coste 12 (`platform/password/password.go:29`).
- `log/slog` estructurado, con logger por petición y `request_id` del contexto (`donaciones/service.go:468`).
- `net/http` de la stdlib (`ServeMux` con `/{id}`), sin framework; validación propia (`platform/httpserver/registrar.go:49`).
- Pruebas: `testing` + `httptest` + fakes por dominio; integración con `testcontainers-go` (`platform/testutil/db.go:40`).

## Dónde se aparta del estándar
| El estándar dice | Aquí se hace | En código nuevo |
|---|---|---|
| postgres-db: listados grandes paginan por la clave | `LIMIT/OFFSET` acotados y `count(*)` pareja (`db/queries/donation_methods.sql:68`) | Se sigue el proyecto |
| go-backend: SQL escrito en el paquete | SQL solo en `queries/*.sql` (sqlc) (`donaciones/repository.go:43`) | Se sigue el proyecto |
| go-backend: el paquete ya es parte del nombre | `*service` concreto; interfaz `<Dominio>Service` (`donaciones/handler.go:32`) | Se sigue el proyecto |

## Deuda
- Seguridad: el estándar pide un rol de aplicación sin SUPERUSER; hoy `DATABASE_URL` usa `POSTGRES_USER` (`../docker-compose.yml:58`). Decide una persona.
- Seguridad: `?sslmode=disable` en desarrollo y CI; el estándar pide scram y nada de `trust` por red (`../docker-compose.yml:58`).
- Seguridad: credenciales de desarrollo en `docker-compose.yml` (`:10`, `:72`); `.env` sí está fuera de git.
- `EXPLAIN (ANALYZE, BUFFERS)` en cada listado nuevo: sin constancia en el repo.
