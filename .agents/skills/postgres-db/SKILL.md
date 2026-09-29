---
name: postgres-db
description: Convenciones de la empresa para diseñar esquemas, migraciones y consultas en PostgreSQL. Usar al crear tablas, migraciones, índices o consultas SQL.
---
# PostgreSQL — convenciones

## Versión
PostgreSQL 16 o superior. Migraciones con `golang-migrate`; consultas tipadas con `sqlc`.

## Migraciones
- Archivos en `backend/migrations/` con numeración secuencial: `000003_add_orders.up.sql` y `000003_add_orders.down.sql`.
- Cada `up` tiene su `down` que lo revierte por completo.
- **Nunca** edites una migración ya aplicada en algún entorno; crea una nueva.
- Cambios peligrosos en tablas grandes (agregar columna `NOT NULL`, renombrar) se hacen en varios pasos: agregar nullable → rellenar → restringir.
- Crea índices en tablas con datos usando `CREATE INDEX CONCURRENTLY` (en una migración separada, sin transacción).

## Diseño de tablas
- Nombres en inglés, `snake_case`, tablas en plural (`users`, `order_items`).
- Clave primaria: `id UUID PRIMARY KEY DEFAULT gen_random_uuid()` (o `BIGINT GENERATED ALWAYS AS IDENTITY` si el plan lo justifica).
- Siempre: `created_at TIMESTAMPTZ NOT NULL DEFAULT now()` y `updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`.
- Fechas siempre `TIMESTAMPTZ`. Dinero en `NUMERIC(12,2)` o en centavos `BIGINT`, nunca `FLOAT`.
- Texto: `TEXT` con `CHECK` de longitud si hay límite de negocio.
- Claves foráneas explícitas con `ON DELETE` decidido conscientemente.
- `NOT NULL` por defecto; nullable solo si el negocio lo requiere.
- Índice en cada clave foránea y en columnas usadas en `WHERE` u `ORDER BY` frecuentes.
- Restricciones `UNIQUE` y `CHECK` en la base de datos, no solo en el código.
- Borrado lógico (`deleted_at`) solo si la spec lo pide.

## Consultas
- Siempre parametrizadas (`$1`, `$2`). Nunca concatenar strings.
- Nada de `SELECT *`; lista las columnas.
- Paginación por cursor (`WHERE id > $1 ORDER BY id LIMIT $2`) para listas grandes.
- Operaciones que modifican varias tablas van en una transacción.
- Revisa con `EXPLAIN ANALYZE` las consultas de listados o reportes.

## Ejemplo
```sql
-- 000001_create_users.up.sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE CHECK (char_length(email) <= 254),
    password_hash TEXT NOT NULL,
    full_name     TEXT NOT NULL CHECK (char_length(full_name) BETWEEN 1 AND 120),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 000001_create_users.down.sql
DROP TABLE IF EXISTS users;
```
