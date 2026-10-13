-- Consultas del dominio `ministerios` (F6): la sección pública, el bloque de la
-- portada y la gestión del panel. Convenciones: skill `postgres-db` y
-- specs/006-ministerios/data-model.md (fuente de verdad).
--
-- Columnas listadas explícitamente (sin `SELECT *`), parámetros nombrados
-- (`sqlc.arg`/`sqlc.narg`) y `ORDER BY` determinista `sort_order, id`
-- (§8.1.7; Q5/FR-004/FR-016/P6-19: el panel ve **el mismo** orden que el
-- visitante).
--
-- Separación público/panel en consultas DISTINTAS (FR-012/FR-026):
--   * `ListPublishedMinistries` es la ÚNICA vía pública: exige
--     `publication_state = 'published'` y, con el interruptor apagado, la fila
--     **no transporta** ningún dato del encargado (`CASE WHEN leader_visible` →
--     `NULL`, R6-2/RG6-1: privacidad garantizada en la consulta, no en la
--     interfaz).
--   * Las del panel (`ListMinistries`, `GetMinistryByID`) son las únicas que
--     devuelven borradores y la ficha del encargado completa, y viven detrás del
--     permiso `ministerios` (FR-011).
--
-- El borrado es físico (sin `deleted_at`) y la imagen **no** borra el archivo:
-- deja de servirse porque `IsFilePublished` (media.sql) ya no lo encuentra
-- (A6-11). La auditoría NO se define aquí: se reutiliza `InsertAdminAction` de
-- `queries/audit.sql` dentro de la transacción de cada mutación (R6-12); no hay
-- ninguna escritura sobre `admin_actions` en este archivo (registro de solo
-- inserción, FR-025 de F2).

-- name: ListPublishedMinistries :many
-- Público (sección y bloque de la portada): solo `published`, en el orden manual
-- del módulo (`sort_order`, `id`) y acotado por `lim` (el service pasa
-- `maxPublicItems = 200` en la sección y `maxPortadaItems = 5` en el bloque: el
-- tope 5 es del SERVIDOR, FR-027).
--
-- Privacidad (FR-026/SC-014, R6-2): los cuatro campos del encargado salen con
-- `CASE WHEN leader_visible THEN … END`, de modo que con el interruptor apagado
-- la fila devuelve `NULL` —ni el nombre ni el celular/correo—. `publication_state`
-- no se selecciona: el DTO público nunca lo expone y la condición ya está en el
-- `WHERE`.
--
-- NOTA para el repository (T607): sqlc no sabe inferir el tipo de un `CASE`
-- (`interface{}` en las cuatro columnas del encargado). El repository las mapea
-- con un helper pequeño (`nil` → sin dato, `string` → dato) y **no** deja salir
-- `interface{}` ni un `pgtype.*` del paquete (`§8.1.6`, R4). Enmascarar con un
-- `UNION ALL` las tiparía, pero R6-2/P6-2 fijan el `CASE` literal.
SELECT id,
       name_es,
       name_en,
       description_es,
       description_en,
       CASE WHEN leader_visible THEN leader_name_es END AS leader_name_es,
       CASE WHEN leader_visible THEN leader_name_en END AS leader_name_en,
       CASE WHEN leader_visible THEN leader_phone END AS leader_phone,
       CASE WHEN leader_visible THEN leader_email END AS leader_email,
       leader_visible,
       whatsapp_url,
       schedule_es,
       schedule_en,
       place_es,
       place_en,
       image_file,
       image_alt_es,
       image_alt_en,
       sort_order,
       created_at,
       updated_at
FROM ministries
WHERE publication_state = 'published'
ORDER BY sort_order, id
LIMIT sqlc.arg('lim');

-- name: ListMinistries :many
-- Panel paginado: **todo** (incluidos borradores y fichas del encargado no
-- publicadas: dato guardado ≠ dato publicado) y con el **mismo** orden que ve el
-- visitante (`sort_order`, `id`, P6-19). `publication_state` opcional (NULL =
-- sin filtro). `whatsapp_url_shared` se recalcula en cada lectura (nunca se
-- persiste, R6-11): es el aviso no bloqueante de grupo compartido.
SELECT m.id,
       m.name_es,
       m.name_en,
       m.description_es,
       m.description_en,
       m.leader_name_es,
       m.leader_name_en,
       m.leader_phone,
       m.leader_email,
       m.leader_visible,
       m.whatsapp_url,
       m.schedule_es,
       m.schedule_en,
       m.place_es,
       m.place_en,
       m.image_file,
       m.image_alt_es,
       m.image_alt_en,
       m.publication_state,
       m.sort_order,
       m.created_at,
       m.updated_at,
       (
           m.whatsapp_url IS NOT NULL
           AND EXISTS (
               SELECT 1
               FROM ministries other
               WHERE other.whatsapp_url = m.whatsapp_url
                 AND other.id <> m.id
           )
       ) AS whatsapp_url_shared
FROM ministries m
WHERE (sqlc.narg('publication_state')::text IS NULL
       OR m.publication_state = sqlc.narg('publication_state')::text)
ORDER BY m.sort_order, m.id
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountMinistries :one
-- Mismo filtro que `ListMinistries` para que `total` corresponda a la página.
SELECT count(*)::bigint AS total
FROM ministries m
WHERE (sqlc.narg('publication_state')::text IS NULL
       OR m.publication_state = sqlc.narg('publication_state')::text);

-- name: GetMinistryByID :one
-- Panel (detalle y respuestas de escritura): todas las columnas —incluida la
-- ficha del encargado íntegra aunque no esté publicada— más el aviso de enlace
-- compartido (R6-11), recalculado aquí.
SELECT m.id,
       m.name_es,
       m.name_en,
       m.description_es,
       m.description_en,
       m.leader_name_es,
       m.leader_name_en,
       m.leader_phone,
       m.leader_email,
       m.leader_visible,
       m.whatsapp_url,
       m.schedule_es,
       m.schedule_en,
       m.place_es,
       m.place_en,
       m.image_file,
       m.image_alt_es,
       m.image_alt_en,
       m.publication_state,
       m.sort_order,
       m.created_at,
       m.updated_at,
       (
           m.whatsapp_url IS NOT NULL
           AND EXISTS (
               SELECT 1
               FROM ministries other
               WHERE other.whatsapp_url = m.whatsapp_url
                 AND other.id <> m.id
           )
       ) AS whatsapp_url_shared
FROM ministries m
WHERE m.id = sqlc.arg('id');

-- name: InsertMinistry :one
-- Alta (FR-010): todo lo opcional va como `narg` (NULL cuando no existe) y
-- `leader_visible` llega SIEMPRE explícito desde el service —que aplica `false`
-- si el DTO no trae la clave, «oculto por defecto» (P6-3)—; el `DEFAULT FALSE`
-- de la tabla cubre los INSERT que omitan la columna (T611).
INSERT INTO ministries (
    name_es, name_en, description_es, description_en,
    leader_name_es, leader_name_en, leader_phone, leader_email, leader_visible,
    whatsapp_url,
    schedule_es, schedule_en, place_es, place_en,
    image_file, image_alt_es, image_alt_en,
    publication_state, sort_order
) VALUES (
    sqlc.arg('name_es'), sqlc.narg('name_en'),
    sqlc.arg('description_es'), sqlc.narg('description_en'),
    sqlc.narg('leader_name_es'), sqlc.narg('leader_name_en'),
    sqlc.narg('leader_phone'), sqlc.narg('leader_email'), sqlc.arg('leader_visible'),
    sqlc.narg('whatsapp_url'),
    sqlc.narg('schedule_es'), sqlc.narg('schedule_en'),
    sqlc.narg('place_es'), sqlc.narg('place_en'),
    sqlc.narg('image_file'), sqlc.narg('image_alt_es'), sqlc.narg('image_alt_en'),
    sqlc.arg('publication_state'), sqlc.arg('sort_order')
)
RETURNING id, name_es, name_en, description_es, description_en,
          leader_name_es, leader_name_en, leader_phone, leader_email, leader_visible,
          whatsapp_url, schedule_es, schedule_en, place_es, place_en,
          image_file, image_alt_es, image_alt_en,
          publication_state, sort_order, created_at, updated_at;

-- name: UpdateMinistry :one
-- El service **fusiona** los campos presentes del `PATCH` sobre la fila actual y
-- esta consulta escribe **la fila completa ya fusionada** (los campos ausentes no
-- cambian y un `null` explícito borra lo opcional): última escritura completa
-- gana (P6-20/R6-16/A6-3).
UPDATE ministries
SET name_es           = sqlc.arg('name_es'),
    name_en           = sqlc.narg('name_en'),
    description_es    = sqlc.arg('description_es'),
    description_en    = sqlc.narg('description_en'),
    leader_name_es    = sqlc.narg('leader_name_es'),
    leader_name_en    = sqlc.narg('leader_name_en'),
    leader_phone      = sqlc.narg('leader_phone'),
    leader_email      = sqlc.narg('leader_email'),
    leader_visible    = sqlc.arg('leader_visible'),
    whatsapp_url      = sqlc.narg('whatsapp_url'),
    schedule_es       = sqlc.narg('schedule_es'),
    schedule_en       = sqlc.narg('schedule_en'),
    place_es          = sqlc.narg('place_es'),
    place_en          = sqlc.narg('place_en'),
    image_file        = sqlc.narg('image_file'),
    image_alt_es      = sqlc.narg('image_alt_es'),
    image_alt_en      = sqlc.narg('image_alt_en'),
    publication_state = sqlc.arg('publication_state'),
    sort_order        = sqlc.arg('sort_order'),
    updated_at        = now()
WHERE id = sqlc.arg('id')
RETURNING id, name_es, name_en, description_es, description_en,
          leader_name_es, leader_name_en, leader_phone, leader_email, leader_visible,
          whatsapp_url, schedule_es, schedule_en, place_es, place_en,
          image_file, image_alt_es, image_alt_en,
          publication_state, sort_order, created_at, updated_at;

-- name: DeleteMinistry :execrows
-- Borrado físico (sin `deleted_at`, default revisable igual que F3/F4); el
-- número de filas afectadas distingue el `404` del éxito. **No** borra el
-- archivo de la imagen: deja de servirse por `IsFilePublished` y puede volver a
-- referenciarse (A6-11).
DELETE FROM ministries
WHERE id = sqlc.arg('id');

-- name: ExistsMinistryDuplicate :one
-- Duplicado exacto del módulo (R6-11): mismo nombre, descripción, encargado y
-- enlace **normalizados** (`btrim` + minúsculas) —los opcionales con
-- `IS NOT DISTINCT FROM`, de modo que dos ministerios sin encargado y sin enlace
-- con el mismo nombre y descripción SÍ son duplicados—. `exclude_id` permite
-- editar sin detectarse a sí mismo. El enlace **compartido** con datos distintos
-- no es duplicado (se señala como aviso).
SELECT EXISTS (
    SELECT 1
    FROM ministries m
    WHERE lower(btrim(m.name_es)) = lower(btrim(sqlc.arg('name_es')::text))
      AND lower(btrim(m.description_es)) = lower(btrim(sqlc.arg('description_es')::text))
      AND lower(btrim(m.leader_name_es))
          IS NOT DISTINCT FROM lower(btrim(sqlc.narg('leader_name_es')::text))
      AND lower(btrim(m.whatsapp_url))
          IS NOT DISTINCT FROM lower(btrim(sqlc.narg('whatsapp_url')::text))
      AND (sqlc.narg('exclude_id')::uuid IS NULL
           OR m.id <> sqlc.narg('exclude_id')::uuid)
) AS duplicate;
