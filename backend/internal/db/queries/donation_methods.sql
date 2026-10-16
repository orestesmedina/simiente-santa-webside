-- Consultas del dominio `donaciones` (F7): los métodos de donación (cuentas IBAN
-- y números de SINPE Móvil). Convenciones: skill `postgres-db` y
-- specs/007-donaciones/data-model.md (fuente de verdad).
--
-- Columnas listadas explícitamente (sin `SELECT *`), parámetros nombrados
-- (`sqlc.arg`/`sqlc.narg`) y `ORDER BY` determinista
-- `CASE kind WHEN 'iban' THEN 0 ELSE 1 END, sort_order, id` (§8.1.7;
-- Q1/FR-004/SC-014/P7-20: iban primero, sinpe después y, dentro de cada grupo,
-- `sort_order` con desempate por `id`; el panel ve **el mismo** orden que el
-- visitante, US2 esc. 10).
--
-- Separación público/panel en consultas DISTINTAS (FR-012/SC-002):
--   * `ListPublishedDonationMethods` es la ÚNICA vía pública: exige
--     `publication_state = 'published'` y acota con `lim` (el service pasa
--     `maxPublicItems = 200` en la sección y `maxPortadaItems = 5` en el bloque:
--     el tope 5 es del SERVIDOR, FR-022).
--   * Las del panel (`ListDonationMethods`, `GetDonationMethodByID`) son las
--     únicas que devuelven borradores y viven detrás del permiso `donaciones`
--     (FR-011).
--
-- El borrado es físico (sin `deleted_at`, default revisable). La auditoría NO se
-- define aquí: se reutiliza `InsertAdminAction` de `queries/audit.sql` dentro de
-- la transacción de cada mutación (R7-10); no hay ninguna escritura sobre
-- `admin_actions` en este archivo (registro de solo inserción).

-- name: ListPublishedDonationMethods :many
-- Público (sección y bloque de la portada): solo `published`, en el orden de la
-- sección (`CASE kind` → iban primero, luego `sort_order`, `id`) y acotado por
-- `lim` (el service pasa `maxPublicItems = 200` en la sección y
-- `maxPortadaItems = 5` en el bloque: el tope 5 es del SERVIDOR, FR-014/FR-022).
--
-- `publication_state` no se selecciona: el DTO público nunca lo expone y la
-- condición ya está en el `WHERE`.
SELECT id,
       kind,
       number,
       institution,
       holder_es,
       holder_en,
       currency,
       sort_order,
       created_at,
       updated_at
FROM donation_methods
WHERE publication_state = 'published'
ORDER BY CASE kind WHEN 'iban' THEN 0 ELSE 1 END, sort_order, id
LIMIT sqlc.arg('lim');

-- name: ListDonationMethods :many
-- Panel paginado: **todo** (borradores incluidos) y con el **mismo** orden que ve
-- el visitante (`CASE kind`, `sort_order`, `id`, P7-20). `publication_state`
-- opcional (NULL = sin filtro).
SELECT m.id,
       m.kind,
       m.number,
       m.institution,
       m.holder_es,
       m.holder_en,
       m.currency,
       m.publication_state,
       m.sort_order,
       m.created_at,
       m.updated_at
FROM donation_methods m
WHERE (sqlc.narg('publication_state')::text IS NULL
       OR m.publication_state = sqlc.narg('publication_state')::text)
ORDER BY CASE m.kind WHEN 'iban' THEN 0 ELSE 1 END, m.sort_order, m.id
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountDonationMethods :one
-- Mismo filtro que `ListDonationMethods` para que `total` corresponda a la página.
SELECT count(*)::bigint AS total
FROM donation_methods m
WHERE (sqlc.narg('publication_state')::text IS NULL
       OR m.publication_state = sqlc.narg('publication_state')::text);

-- name: GetDonationMethodByID :one
-- Panel (detalle y respuestas de escritura): todas las columnas, incluido el
-- estado (el panel sí ve borradores).
SELECT id,
       kind,
       number,
       institution,
       holder_es,
       holder_en,
       currency,
       publication_state,
       sort_order,
       created_at,
       updated_at
FROM donation_methods
WHERE id = sqlc.arg('id');

-- name: InsertDonationMethod :one
-- Alta (FR-010): el `number` llega ya NORMALIZADO por el service (IBAN a los
-- caracteres del número; SINPE `+506XXXXXXXX`, FR-015/FR-016) y lo opcional va
-- como `narg` (NULL cuando no existe). `currency` solo se rellena en IBAN (lo
-- garantiza el CHECK doble de la tabla, Q2/FR-014).
INSERT INTO donation_methods (
    kind, number, institution, holder_es, holder_en, currency,
    publication_state, sort_order
) VALUES (
    sqlc.arg('kind'),
    sqlc.arg('number'),
    sqlc.arg('institution'),
    sqlc.narg('holder_es'),
    sqlc.narg('holder_en'),
    sqlc.narg('currency'),
    sqlc.arg('publication_state'),
    sqlc.arg('sort_order')
)
RETURNING id, kind, number, institution, holder_es, holder_en, currency,
          publication_state, sort_order, created_at, updated_at;

-- name: UpdateDonationMethod :one
-- El service **fusiona** los campos presentes del `PATCH` sobre la fila actual y
-- esta consulta escribe **la fila completa ya fusionada** (los campos ausentes no
-- cambian y un `null` explícito borra lo opcional): última escritura completa gana
-- (P7-15/R7-15).
UPDATE donation_methods
SET kind              = sqlc.arg('kind'),
    number            = sqlc.arg('number'),
    institution       = sqlc.arg('institution'),
    holder_es         = sqlc.narg('holder_es'),
    holder_en         = sqlc.narg('holder_en'),
    currency          = sqlc.narg('currency'),
    publication_state = sqlc.arg('publication_state'),
    sort_order        = sqlc.arg('sort_order'),
    updated_at        = now()
WHERE id = sqlc.arg('id')
RETURNING id, kind, number, institution, holder_es, holder_en, currency,
          publication_state, sort_order, created_at, updated_at;

-- name: DeleteDonationMethod :execrows
-- Borrado físico (sin `deleted_at`, default revisable igual que F3/F4/F6); el
-- número de filas afectadas distingue el `404` del éxito.
DELETE FROM donation_methods
WHERE id = sqlc.arg('id');

-- name: ExistsDonationMethodDuplicate :one
-- Duplicado exacto del módulo (R7-9/P7-10): otro método con el MISMO `kind` y el
-- MISMO `number` **normalizado** (la igualdad es simple porque el número se
-- guarda normalizado). `exclude_id` permite editar sin detectarse a sí mismo
-- (`NULL` en el alta; su `id` en la edición): `AND (@exclude_id::uuid IS NULL OR
-- id <> @exclude_id)` garantiza que una edición que no cambia el número **no**
-- devuelva `409` contra sí misma, y sí lo devuelva si el número coincide con el
-- de **otro** método.
SELECT EXISTS (
    SELECT 1
    FROM donation_methods m
    WHERE m.kind = sqlc.arg('kind')
      AND m.number = sqlc.arg('number')
      AND (sqlc.narg('exclude_id')::uuid IS NULL
           OR m.id <> sqlc.narg('exclude_id')::uuid)
) AS duplicate;
