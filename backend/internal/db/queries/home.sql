-- Consultas del contenido de la portada (F3): identidad, «quiénes somos»,
-- contacto, horario de servicios, canales de WhatsApp y redes sociales.
-- Convenciones: skill `postgres-db` y
-- specs/003-portada-info-general/data-model.md (fuente de verdad).
--
-- Columnas listadas explícitamente (sin `SELECT *`), parámetros nombrados
-- (`sqlc.arg`) y `ORDER BY` determinista: `sort_order, id` en las colecciones
-- ordenables y `network` en redes (§8.1.7).
--
-- Variantes `…Published`: son la ÚNICA vía pública (P3-4/FR-013); la variante
-- sin sufijo es la del panel (exige el permiso `portada`) y es la única que
-- devuelve borradores. Los singletons llevan `singleton BOOLEAN` + `UNIQUE` y
-- el upsert va en transacción con `pg_advisory_xact_lock` (R3-4, lo toma el
-- repository de T314, no esta consulta).
--
-- La auditoría NO se define aquí: se reutiliza `InsertAdminAction` de
-- `queries/audit.sql` dentro de la transacción de cada mutación (R3-11).

-- ------------------------------------------------------------------ identidad

-- name: UpsertHomeIdentity :one
-- Crea o reemplaza el singleton de identidad (UNIQUE (singleton)).
INSERT INTO home_identity (
    name_es, name_en, tagline_es, tagline_en, mission_es, mission_en,
    vision_es, vision_en, logo_file, logo_alt_es, logo_alt_en,
    cover_image_file, cover_image_alt_es, cover_image_alt_en, publication_state
) VALUES (
    sqlc.arg('name_es'), sqlc.arg('name_en'), sqlc.arg('tagline_es'), sqlc.arg('tagline_en'),
    sqlc.arg('mission_es'), sqlc.arg('mission_en'), sqlc.arg('vision_es'), sqlc.arg('vision_en'),
    sqlc.arg('logo_file'), sqlc.arg('logo_alt_es'), sqlc.arg('logo_alt_en'),
    sqlc.arg('cover_image_file'), sqlc.arg('cover_image_alt_es'), sqlc.arg('cover_image_alt_en'),
    sqlc.arg('publication_state')
)
ON CONFLICT (singleton) DO UPDATE SET
    name_es            = EXCLUDED.name_es,
    name_en            = EXCLUDED.name_en,
    tagline_es         = EXCLUDED.tagline_es,
    tagline_en         = EXCLUDED.tagline_en,
    mission_es         = EXCLUDED.mission_es,
    mission_en         = EXCLUDED.mission_en,
    vision_es          = EXCLUDED.vision_es,
    vision_en          = EXCLUDED.vision_en,
    logo_file          = EXCLUDED.logo_file,
    logo_alt_es        = EXCLUDED.logo_alt_es,
    logo_alt_en        = EXCLUDED.logo_alt_en,
    cover_image_file   = EXCLUDED.cover_image_file,
    cover_image_alt_es = EXCLUDED.cover_image_alt_es,
    cover_image_alt_en = EXCLUDED.cover_image_alt_en,
    publication_state  = EXCLUDED.publication_state,
    updated_at         = now()
RETURNING id, singleton, name_es, name_en, tagline_es, tagline_en,
          mission_es, mission_en, vision_es, vision_en,
          logo_file, logo_alt_es, logo_alt_en,
          cover_image_file, cover_image_alt_es, cover_image_alt_en,
          publication_state, created_at, updated_at;

-- name: GetHomeIdentity :one
-- Panel: devuelve el singleton aunque esté en borrador.
SELECT id, singleton, name_es, name_en, tagline_es, tagline_en,
       mission_es, mission_en, vision_es, vision_en,
       logo_file, logo_alt_es, logo_alt_en,
       cover_image_file, cover_image_alt_es, cover_image_alt_en,
       publication_state, created_at, updated_at
FROM home_identity
WHERE singleton
LIMIT 1;

-- name: GetHomeIdentityPublished :one
-- Público: solo el singleton publicado (FR-013); en borrador no devuelve fila.
SELECT id, singleton, name_es, name_en, tagline_es, tagline_en,
       mission_es, mission_en, vision_es, vision_en,
       logo_file, logo_alt_es, logo_alt_en,
       cover_image_file, cover_image_alt_es, cover_image_alt_en,
       publication_state, created_at, updated_at
FROM home_identity
WHERE singleton AND publication_state = 'published'
LIMIT 1;

-- ------------------------------------------------------------- quiénes somos

-- name: UpsertHomeAbout :one
INSERT INTO home_about (text_es, text_en, publication_state)
VALUES (sqlc.arg('text_es'), sqlc.arg('text_en'), sqlc.arg('publication_state'))
ON CONFLICT (singleton) DO UPDATE SET
    text_es           = EXCLUDED.text_es,
    text_en           = EXCLUDED.text_en,
    publication_state = EXCLUDED.publication_state,
    updated_at        = now()
RETURNING id, singleton, text_es, text_en, publication_state, created_at, updated_at;

-- name: GetHomeAbout :one
SELECT id, singleton, text_es, text_en, publication_state, created_at, updated_at
FROM home_about
WHERE singleton
LIMIT 1;

-- name: GetHomeAboutPublished :one
SELECT id, singleton, text_es, text_en, publication_state, created_at, updated_at
FROM home_about
WHERE singleton AND publication_state = 'published'
LIMIT 1;

-- -------------------------------------------------------------------- contacto

-- name: UpsertHomeContact :one
INSERT INTO home_contact (address_es, address_en, email, phone, publication_state)
VALUES (
    sqlc.arg('address_es'), sqlc.arg('address_en'), sqlc.arg('email'),
    sqlc.arg('phone'), sqlc.arg('publication_state')
)
ON CONFLICT (singleton) DO UPDATE SET
    address_es        = EXCLUDED.address_es,
    address_en        = EXCLUDED.address_en,
    email             = EXCLUDED.email,
    phone             = EXCLUDED.phone,
    publication_state = EXCLUDED.publication_state,
    updated_at        = now()
RETURNING id, singleton, address_es, address_en, email, phone,
          publication_state, created_at, updated_at;

-- name: GetHomeContact :one
SELECT id, singleton, address_es, address_en, email, phone,
       publication_state, created_at, updated_at
FROM home_contact
WHERE singleton
LIMIT 1;

-- name: GetHomeContactPublished :one
SELECT id, singleton, address_es, address_en, email, phone,
       publication_state, created_at, updated_at
FROM home_contact
WHERE singleton AND publication_state = 'published'
LIMIT 1;

-- --------------------------------------------------------- horario (servicios)

-- name: InsertHomeService :one
INSERT INTO home_services (
    day_of_week, start_time, end_time, name_es, name_en,
    description_es, description_en, place_es, place_en,
    publication_state, sort_order
) VALUES (
    sqlc.arg('day_of_week'), sqlc.arg('start_time'), sqlc.arg('end_time'),
    sqlc.arg('name_es'), sqlc.arg('name_en'),
    sqlc.arg('description_es'), sqlc.arg('description_en'),
    sqlc.arg('place_es'), sqlc.arg('place_en'),
    sqlc.arg('publication_state'), sqlc.arg('sort_order')
)
RETURNING id, day_of_week, start_time, end_time, name_es, name_en,
          description_es, description_en, place_es, place_en,
          publication_state, sort_order, created_at, updated_at;

-- name: UpdateHomeService :one
-- Reemplazo completo de los campos mutables (el service fusiona el PATCH y
-- valida antes de llamar).
UPDATE home_services
SET day_of_week       = sqlc.arg('day_of_week'),
    start_time        = sqlc.arg('start_time'),
    end_time          = sqlc.arg('end_time'),
    name_es           = sqlc.arg('name_es'),
    name_en           = sqlc.arg('name_en'),
    description_es    = sqlc.arg('description_es'),
    description_en    = sqlc.arg('description_en'),
    place_es          = sqlc.arg('place_es'),
    place_en          = sqlc.arg('place_en'),
    publication_state = sqlc.arg('publication_state'),
    sort_order        = sqlc.arg('sort_order'),
    updated_at        = now()
WHERE id = sqlc.arg('id')
RETURNING id, day_of_week, start_time, end_time, name_es, name_en,
          description_es, description_en, place_es, place_en,
          publication_state, sort_order, created_at, updated_at;

-- name: DeleteHomeService :execrows
DELETE FROM home_services
WHERE id = sqlc.arg('id');

-- name: GetHomeServiceByID :one
SELECT id, day_of_week, start_time, end_time, name_es, name_en,
       description_es, description_en, place_es, place_en,
       publication_state, sort_order, created_at, updated_at
FROM home_services
WHERE id = sqlc.arg('id');

-- name: ListHomeServices :many
-- Panel: todos los servicios (borradores incluidos), orden estable.
SELECT id, day_of_week, start_time, end_time, name_es, name_en,
       description_es, description_en, place_es, place_en,
       publication_state, sort_order, created_at, updated_at
FROM home_services
ORDER BY sort_order, id;

-- name: ListHomeServicesPublished :many
-- Público: solo servicios publicados (FR-013).
SELECT id, day_of_week, start_time, end_time, name_es, name_en,
       description_es, description_en, place_es, place_en,
       publication_state, sort_order, created_at, updated_at
FROM home_services
WHERE publication_state = 'published'
ORDER BY sort_order, id;

-- ------------------------------------------------- canales de WhatsApp

-- name: InsertHomeWhatsappChannel :one
INSERT INTO home_whatsapp_channels (
    name_es, name_en, kind, destination, publication_state, sort_order
) VALUES (
    sqlc.arg('name_es'), sqlc.arg('name_en'), sqlc.arg('kind'),
    sqlc.arg('destination'), sqlc.arg('publication_state'), sqlc.arg('sort_order')
)
RETURNING id, kind, destination, name_es, name_en, publication_state, sort_order,
          created_at, updated_at;

-- name: UpdateHomeWhatsappChannel :one
UPDATE home_whatsapp_channels
SET name_es           = sqlc.arg('name_es'),
    name_en           = sqlc.arg('name_en'),
    kind              = sqlc.arg('kind'),
    destination       = sqlc.arg('destination'),
    publication_state = sqlc.arg('publication_state'),
    sort_order        = sqlc.arg('sort_order'),
    updated_at        = now()
WHERE id = sqlc.arg('id')
RETURNING id, kind, destination, name_es, name_en, publication_state, sort_order,
          created_at, updated_at;

-- name: DeleteHomeWhatsappChannel :execrows
DELETE FROM home_whatsapp_channels
WHERE id = sqlc.arg('id');

-- name: GetHomeWhatsappChannelByID :one
SELECT id, kind, destination, name_es, name_en, publication_state, sort_order,
       created_at, updated_at
FROM home_whatsapp_channels
WHERE id = sqlc.arg('id');

-- name: ListHomeWhatsappChannels :many
SELECT id, kind, destination, name_es, name_en, publication_state, sort_order,
       created_at, updated_at
FROM home_whatsapp_channels
ORDER BY sort_order, id;

-- name: ListHomeWhatsappChannelsPublished :many
SELECT id, kind, destination, name_es, name_en, publication_state, sort_order,
       created_at, updated_at
FROM home_whatsapp_channels
WHERE publication_state = 'published'
ORDER BY sort_order, id;

-- ----------------------------------------------------------------- redes

-- name: InsertHomeSocialLink :one
INSERT INTO home_social_links (network, url, publication_state)
VALUES (sqlc.arg('network'), sqlc.arg('url'), sqlc.arg('publication_state'))
RETURNING id, network, url, publication_state, created_at, updated_at;

-- name: UpdateHomeSocialLink :one
UPDATE home_social_links
SET network           = sqlc.arg('network'),
    url               = sqlc.arg('url'),
    publication_state = sqlc.arg('publication_state'),
    updated_at        = now()
WHERE id = sqlc.arg('id')
RETURNING id, network, url, publication_state, created_at, updated_at;

-- name: DeleteHomeSocialLink :execrows
DELETE FROM home_social_links
WHERE id = sqlc.arg('id');

-- name: GetHomeSocialLinkByID :one
SELECT id, network, url, publication_state, created_at, updated_at
FROM home_social_links
WHERE id = sqlc.arg('id');

-- name: ListHomeSocialLinks :many
SELECT id, network, url, publication_state, created_at, updated_at
FROM home_social_links
ORDER BY network;

-- name: ListHomeSocialLinksPublished :many
SELECT id, network, url, publication_state, created_at, updated_at
FROM home_social_links
WHERE publication_state = 'published'
ORDER BY network;

-- ---------------------------------------------------- archivos de imagen

-- name: IsHomeFilePublished :one
-- ¿El nombre de archivo está referenciado por contenido PUBLICADO? (analyze C4)
-- `TRUE` solo si la identidad publicada lo usa como logo o imagen de portada;
-- `FALSE` cuando no existe referencia, la identidad está en borrador o el
-- archivo es huérfano. F4–F9 amplían esta consulta con sus tablas.
SELECT EXISTS (
    SELECT 1
    FROM home_identity
    WHERE publication_state = 'published'
      AND sqlc.arg('file_name') IN (logo_file, cover_image_file)
) AS published;
