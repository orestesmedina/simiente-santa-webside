-- Consultas del dominio `media` (F4, R4-6): la descarga pública de archivos
-- (`GET /api/v1/media/{fileName}`) deja de ser un asunto de la portada y pasa a
-- un dominio propio, compartido por F3–F9.
-- Convenciones: skill `postgres-db` y
-- specs/004-eventos-actividades/data-model.md (fuente de verdad).
--
-- `IsFilePublished` sustituye a `IsHomeFilePublished` (de `home.sql`): responde
-- TRUE solo si el nombre está referenciado por contenido **publicado y visible
-- hoy** (decisión firme de la puerta 2, R4-6; F6 añade la tercera rama, R6-7):
--   * `home_identity` publicada (F3: logo e imagen de portada), sin fecha;
--   * `agenda_events` / `agenda_activities` publicados **y próximos** según la
--     regla R4-2 (`today`/`now_time` que el service calcula en `APP_TZ` y pasa
--     parametrizados); y
--   * `ministries` publicado (F6: la imagen del ministerio, sin fecha).
-- Un elemento vencido deja de servir su imagen **sin borrar el archivo** y la
-- misma URL vuelve a responder 200 al cargarle una fecha próxima; igual en F6: el
-- ministerio en borrador no sirve su imagen y la URL vuelve al publicarlo.

-- name: IsFilePublished :one
-- ¿El nombre de archivo lo referencia contenido publicado y próximo? `TRUE`
-- solo en ese caso; `FALSE` cuando no hay referencia, el elemento está en
-- borrador o su fecha ya pasó (el archivo sigue en disco: la consulta no lo
-- toca). Sin `SELECT *` y todo parametrizado (§IV).
SELECT EXISTS (
    SELECT 1
    FROM (
        SELECT hi.id
        FROM home_identity hi
        WHERE hi.publication_state = 'published'
          AND sqlc.arg('file_name')::text IN (hi.logo_file, hi.cover_image_file)

        UNION ALL

        SELECT e.id
        FROM agenda_events e
        WHERE e.publication_state = 'published'
          AND e.image_file = sqlc.arg('file_name')::text
          AND EXISTS (
              SELECT 1
              FROM agenda_event_dates d
              WHERE d.event_id = e.id
                AND (
                    d.event_date > sqlc.arg('today')::date
                    OR (d.event_date = sqlc.arg('today')::date
                        AND (d.end_time IS NULL
                             OR d.end_time::time >= sqlc.arg('now_time')::time))
                )
          )

        UNION ALL

        SELECT a.id
        FROM agenda_activities a
        WHERE a.publication_state = 'published'
          AND a.image_file = sqlc.arg('file_name')::text
          AND (
              a.activity_date > sqlc.arg('today')::date
              OR (a.activity_date = sqlc.arg('today')::date
                  AND (a.end_time IS NULL
                       OR a.end_time::time >= sqlc.arg('now_time')::time))
          )

        UNION ALL

        -- F6 (R6-7): la imagen de un ministerio PUBLICADO. Rama aditiva: el
        -- ministerio en borrador no sirve su imagen y el archivo no se borra
        -- (al publicarlo, la misma URL vuelve a responder 200).
        SELECT m.id
        FROM ministries m
        WHERE m.publication_state = 'published'
          AND m.image_file = sqlc.arg('file_name')::text
    ) AS referenced
) AS published;
