-- 000007_create_agenda_content (down) — revierte 000007 por completo (§VI).
-- Las tablas hijas se tiran primero (agenda_event_dates referencia
-- agenda_events; eventos y actividades referencian agenda_categories). El
-- borrado del catálogo sembrado es implícito al tirar la tabla.
DROP TABLE IF EXISTS agenda_activities;
DROP TABLE IF EXISTS agenda_event_dates;
DROP TABLE IF EXISTS agenda_events;
DROP TABLE IF EXISTS agenda_categories;
