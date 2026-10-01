# Data Model — F1 Estructura base

**Fecha**: 2026-09-29 · **Actualizado**: 2026-09-30 (ampliación de alcance: convención de migraciones y criterio de asignación de tablas) · **Rama**: `001-estructura-base`

## Decisión explícita: F1 no tiene modelo de datos de negocio

La spec lo establece sin ambigüedad:

> **Out of Scope**: "Esquema de base de datos de negocio: en F1 la base de datos solo necesita existir y aceptar conexiones; no hay tablas de negocio."
> **Assumptions**: "La base de datos de F1 no contiene tablas de negocio; basta con que acepte conexiones para verificar el estado."

Por tanto:

- **No se crea ninguna tabla, índice ni restricción de negocio en F1.**
- La base de datos (servicio `db` de `docker-compose.yml`, PostgreSQL 16.4-alpine, ya provisto por el kit) solo debe **existir y aceptar conexiones** para que `GET /healthz` pueda verificar la conectividad con `Ping`.
- El primer esquema de negocio llegará con **F2 (Acceso y gestión de usuarios)**, que diseñará su propio `data-model.md` (tabla `users`, `sessions`, roles, permisos…) siguiendo la skill `postgres-db`.
- El **ejercicio de práctica de la receta** (`quickstart.md` §9) sí crea una tabla (`sample_items`, ejemplo sugerido), pero vive en una rama de práctica que **se descarta**: nunca llega a `main`, que cierra F1 con cero tablas.

## Lo único que toca la base de datos en F1

### Migración baseline `000001` (no-op)

```text
backend/migrations/
├── 000001_baseline.up.sql    -- contenido: solo un comentario (no-op)
└── 000001_baseline.down.sql  -- contenido: solo un comentario (no-op)
```

- **Propósito**: (1) crear la carpeta `backend/migrations/` con archivos reales que fijan la convención de nomenclatura `NNNNNN_descripcion.{up,down}.sql` de la skill `postgres-db`; (2) ejercitar el toolchain de migraciones de punta a punta desde F1 (target `make db-migrate`, paso "Migraciones" del CI, control de inmutabilidad de los hooks de git).
- **Efecto en la BD**: ninguno, más allá de que `golang-migrate` registra la versión `1` en su tabla interna `schema_migrations` (tabla de control de la herramienta, no de negocio).
- **Reversible**: el `.down.sql` también es no-op; `migrate down 1` deja la BD como estaba.

### Conexión de verificación de estado

- El backend abre un pool `pgxpool` contra `DATABASE_URL` y ejecuta `Ping(ctx)` (equivalente a una conexión corta, sin consultas de negocio) en cada llamada a `/healthz`.
- No hay entidades, campos, relaciones, validaciones ni transiciones de estado que modelar.

## Convención de `backend/migrations/` (vigente desde F1)

| Regla | Detalle |
|---|---|
| Nomenclatura | `NNNNNN_descripcion_snake_case.up.sql` + su `.down.sql` (numeración secuencial, 6 dígitos; F1 usa `000001`, F2 arranca en `000002`) |
| Completitud | Cada `up` tiene su `down` que lo revierte por completo (skill `postgres-db`) |
| Inmutabilidad | **Nunca** se edita una migración aplicada en ningún entorno; se crea una nueva (lo vigilan los hooks del kit y el paso "Migraciones" del CI) |
| Contenido | SQL válido y autónomo por archivo: sqlc usa estas migraciones como **esquema** (D6 del plan), así que no deben depender de objetos creados fuera de ellas |
| Convenciones de tablas (para las futuras) | Nombres en inglés, `snake_case`, plural; `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`; `created_at`/`updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`; `NOT NULL` por defecto; `CHECK`/`UNIQUE` en la base, no solo en el código; claves foráneas explícitas con `ON DELETE` decidido; índices en FK y columnas de `WHERE`/`ORDER BY`; `TEXT` con `CHECK` de longitud; dinero en `NUMERIC(12,2)` o centavos `BIGINT`, nunca `FLOAT` |

## Quién ejecuta las migraciones (y el runner diferido)

- **F1**: el CLI **`golang-migrate`** — en desarrollo con `make db-migrate` (`migrate -path backend/migrations -database "$DATABASE_URL" up`) y en CI con el paso "Migraciones" (el mismo del kit, contra el PostgreSQL de servicio). Es la decisión D10 del plan.
- **Diferido** (con punto de decisión): `internal/platform/migrate`, un runner **embebido** (`//go:embed` de `backend/migrations/`) opt-in por entorno (`RUN_MIGRATIONS=true`), para cuando exista un entorno desplegado que deba auto-migrar sin CLI. Se decide en el plan de esa funcionalidad; **no** se implementa en F1 (D23 del plan).
- `schema_migrations` es la tabla de control de `golang-migrate`; no se modela, no se edita a mano y no cuenta como tabla de negocio.

## Capa de datos: sqlc sobre estas migraciones (D6)

- `backend/sqlc.yaml`: `schema: migrations` · `queries: internal/db/queries` · `out: internal/db`.
- Las consultas de negocio viven en `backend/internal/db/queries/*.sql`, siempre **parametrizadas** y sin `SELECT *`; el código generado se commitea (`make sqlc-gen`) y su deriva se controla con `make sqlc-verify` (riesgo R4 de `plan.md`; R20 de `research.md`).
- En F1 `internal/db/queries/` está **vacío** (la única operación es un `Ping`); la primera consulta nace con el ejercicio de práctica de la receta y la primera real con F2.
- **Stored procedures / funciones SQL**: solo como excepción justificada, invocados dentro del `repository` (nunca en el service); recuerda que un `CALL` dentro de una transacción no puede controlar transacciones — la gestiona `database.WithTx` (D-A3).

## Criterio: qué tabla va a cada funcionalidad futura

1. **Cada tabla nace con la funcionalidad del roadmap que la necesita**: nunca se anticipa el esquema "por si acaso" (mismo criterio que D22/D23 del plan: sin configuración muerta).
2. **Una tabla pertenece a un dominio**: la crea su migración y la consulta únicamente el `repository` de ese dominio. Si dos dominios necesitan el mismo SQL, esa consulta **sube a `internal/db/queries/`** como consulta nombrada compartida; los dominios no se importan entre sí (`arquitectura.md` §1.2).
3. **Tablas transversales** (usuarios, sesiones, permisos) las crea la funcionalidad que introduce esa capacidad (F2) y viven en `internal/platform/` + el dominio `usuarios`; las áreas de negocio las consumen vía middlewares (`authn`/`authz`), sin tocarlas.

Asignación prevista (orientativa; cada funcionalidad refina su modelo en su propio `data-model.md`):

| Funcionalidad | Tablas previstas (nombre orientativo) |
|---|---|
| F2 — Acceso y gestión de usuarios | `users`, `sessions`, `roles`, `user_roles`, permisos por módulo (p. ej. `role_permissions`) — ver D-A7, **pendiente de confirmación del humano** |
| F3 — Portada e información general | contenido de la portada/información con estado borrador/publicado y versión por idioma (español base, inglés opcional — decisiones 6 y 8 del roadmap) |
| F4 — Eventos y actividades | `events` y `activities` (modelos **separados**, decisión 3 del roadmap) |
| F5 — Grupos de conexión | `groups` (catálogo con horario, lugar, contacto del encargado) |
| F6 — Ministerios | `ministries` (encargado, enlace de WhatsApp) |
| F7 — Donaciones | `donation_accounts` (IBAN/SINPE) e información de uso |
| F8 — Noticias y galería | `posts`/`news`, `media_items` (imágenes/videos; aquí entra el almacenamiento de objetos, diferido en D23) |
| F9 — Medios (prédicas y podcasts) | `series`/`episodes` con enlaces a YouTube/Spotify |

## Validación

La "corrección" del modelo de F1 se valida indirectamente:

1. `make up` levanta `db` y pasa su healthcheck (`pg_isready`).
2. `make db-migrate` aplica la baseline sin error (`migrate ... up` → versión 1) y `migrate down 1` la revierte sin efecto.
3. `GET /healthz` reporta `database: "connected"` con la BD viva y `503 database_unavailable` con la BD detenida (escenarios en `quickstart.md` §1–3).
4. El CI del kit aplica las migraciones a su PostgreSQL de servicio y el paso "Migraciones existentes no se modifican" protege la inmutabilidad.
