# Data Model — F1 Estructura base

**Fecha**: 2026-09-29 · **Rama**: `001-estructura-base`

## Decisión explícita: F1 no tiene modelo de datos de negocio

La spec lo establece sin ambigüedad:

> **Out of Scope**: "Esquema de base de datos de negocio: en F1 la base de datos solo necesita existir y aceptar conexiones; no hay tablas de negocio."
> **Assumptions**: "La base de datos de F1 no contiene tablas de negocio; basta con que acepte conexiones para verificar el estado."

Por tanto:

- **No se crea ninguna tabla, índice ni restricción de negocio en F1.**
- La base de datos (servicio `db` de `docker-compose.yml`, PostgreSQL 16.4-alpine, ya provisto por el kit) solo debe **existir y aceptar conexiones** para que `GET /healthz` pueda verificar la conectividad con `Ping`.
- El primer esquema de negocio llegará con **F2 (Acceso y gestión de usuarios)**, que diseñará su propio `data-model.md` (tabla `users`, roles, permisos…) siguiendo la skill `postgres-db`.

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

## Convenciones que heredarán las tablas futuras (registradas aquí como referencia del patrón)

Cuando F2+ cree tablas, aplicará la skill `postgres-db` (ya vigente, no es decisión de F1): nombres `snake_case` en plural, `id UUID` con `gen_random_uuid()`, `created_at`/`updated_at TIMESTAMPTZ`, claves foráneas e índices explícitos, `NOT NULL` por defecto, una migración `up`/`down` por cambio y jamás editar una migración aplicada (lo vigilan los hooks del kit y el CI).

## Validación

La "corrección" del modelo de F1 se valida indirectamente:

1. `make up` levanta `db` y pasa su healthcheck (`pg_isready`).
2. `make db-migrate` aplica la baseline sin error (`migrate ... up` → versión 1).
3. `GET /healthz` reporta `database: "connected"` con la BD viva y `"disconnected"` con la BD detenida (escenarios en `quickstart.md`).
