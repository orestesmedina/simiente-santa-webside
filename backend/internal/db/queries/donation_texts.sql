-- Consultas del dominio `donaciones` (F7): los dos textos libres de la sección
-- «Ofrendas y diezmos» (uno por clave, `text_key UNIQUE`). Convenciones: skill
-- `postgres-db` y specs/007-donaciones/data-model.md (fuente de verdad).
--
-- Patrón «singleton por clave»: `UNIQUE (text_key)` garantiza ≤1 fila por clave
-- (`intro`/`usage`) y las filas se crean en el primer guardado (`UpsertDonationText`
-- del `PUT`), como los singletons de F3. La migración NO siembra contenido y no
-- hay operación de borrado (un texto se retira con `publication_state = 'draft'`,
-- P7-17).
--
-- Separación público/panel en consultas DISTINTAS (FR-012/SC-002):
--   * `ListPublishedDonationTexts` es la ÚNICA vía pública: exige
--     `publication_state = 'published'`.
--   * `ListDonationTexts`/`GetDonationTextByKey` son las del panel (borradores
--     incluidos, detrás del permiso `donaciones`, FR-011).
--
-- La auditoría NO se define aquí: se reutiliza `InsertAdminAction` de
-- `queries/audit.sql` dentro de la transacción de cada mutación (R7-10).
--
-- `ORDER BY text_key` es determinista (`intro` < `usage`; la composición de la
-- sección la decide el service, §8.1.7).

-- name: ListPublishedDonationTexts :many
-- Público: solo los textos publicados, en orden de clave. El service los
-- empaqueta como `intro`/`usage` (P7-12).
SELECT id,
       text_key,
       body_es,
       body_en,
       publication_state,
       created_at,
       updated_at
FROM donation_texts
WHERE publication_state = 'published'
ORDER BY text_key;

-- name: ListDonationTexts :many
-- Panel: las dos filas posibles (existentes o no) en ambos idiomas con su estado
-- de publicación propio (Q10).
SELECT id,
       text_key,
       body_es,
       body_en,
       publication_state,
       created_at,
       updated_at
FROM donation_texts
ORDER BY text_key;

-- name: GetDonationTextByKey :one
-- Panel (detalle y respuesta del `PUT`): la fila de la clave.
SELECT id,
       text_key,
       body_es,
       body_en,
       publication_state,
       created_at,
       updated_at
FROM donation_texts
WHERE text_key = sqlc.arg('text_key');

-- name: UpsertDonationText :one
-- Crea o reemplaza el singleton de la clave (`INSERT … ON CONFLICT (text_key) DO
-- UPDATE …` + `updated_at = now()`, patrón del singleton de F3). El `body_es` es
-- obligatorio y `body_en` opcional (el `CHECK` de español por campo lo garantiza).
-- El estado de publicación es propio del texto (Q10).
INSERT INTO donation_texts (
    text_key, body_es, body_en, publication_state
) VALUES (
    sqlc.arg('text_key'),
    sqlc.arg('body_es'),
    sqlc.narg('body_en'),
    sqlc.arg('publication_state')
)
ON CONFLICT (text_key) DO UPDATE SET
    body_es           = EXCLUDED.body_es,
    body_en           = EXCLUDED.body_en,
    publication_state = EXCLUDED.publication_state,
    updated_at        = now()
RETURNING id, text_key, body_es, body_en, publication_state, created_at, updated_at;
