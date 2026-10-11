# Data Model: F10 — Sistema de diseño y movimiento

**Feature**: F10 · **Branch**: `010-sistema-diseno-movimiento` · **Fecha**: 2026-10-10

## Sin cambios de datos — y por qué

**F10 no crea ni modifica ninguna entidad, tabla, columna, índice, restricción ni migración.**
No hay archivos en `backend/migrations/` para esta funcionalidad. La razón es de alcance:

- F10 es una funcionalidad **de presentación** del sitio público (FR-021): «no introduce datos nuevos
  ni cambia los existentes; no agrega entidades, campos ni reglas de publicación».
- Toda la información que muestra el sitio sigue siendo la de F3 (identidad, «quiénes somos», horario
  de servicios, canales de WhatsApp, redes sociales, contacto, estados de publicación), gestionada
  igual desde el panel, que **no se modifica** (FR-020).
- El **tema** (claro/oscuro) y la **preferencia de movimiento** son estado **del dispositivo del
  visitante**, no datos del sistema: viven en `localStorage['ss.theme']` y en
  `prefers-color-scheme`/`prefers-reduced-motion` del navegador. El idioma sigue en
  `localStorage['ss.lang']` (F3, sin cambios). Ninguno de ellos pasa por la API ni se persiste en el
  servidor.
- El panel (F2/F3) y sus entidades —cuentas, roles, permisos, auditoría— **no se tocan en absoluto**
  (el rediseño del panel es F10B, fuera de F10).

## Entidades existentes (referencia, sin cambios)

Las entidades de F3 se conservan tal cual —ver
`specs/003-portada-info-general/data-model.md` como fuente vigente—:

| Entidad (F3) | Tablas | F10 |
|---|---|---|
| Identidad (singleton) | `home_identity` | Sin cambios; solo cambia su presentación |
| «Quiénes somos» (singleton) | `home_about` | Sin cambios; solo cambia su presentación |
| Contacto (singleton) | `home_contact` | Sin cambios; solo cambia su presentación |
| Horario de servicios | `home_schedule_items` | Sin cambios; solo cambia su presentación |
| Canales de WhatsApp | `home_whatsapp_channels` | Sin cambios; solo cambia su presentación |
| Redes sociales | `home_social_links` | Sin cambios; solo cambia su presentación |
| Publicación/auditoría | columnas `publication_state`, `admin_actions` | Sin cambios |

## Artefactos que F10 crea (conceptuales, no persistidos)

Los «datos» de F10 son de diseño y viven en el repositorio de código/documentación, no en la BD
(detalle en `plan.md`):

1. **Sistema visual**: design tokens (claro y oscuro) en `frontend/src/index.css` y primitivas de
   componente en `frontend/src/components/publico/`.
2. **Temas claro y oscuro**: valores por tema de los tokens semánticos + resolución
   sistema→manual en `frontend/src/features/publico/theme/`.
3. **Patrones de movimiento**: utilidades CSS + `Reveal`/hooks en `frontend/src/features/publico/motion/`.
4. **Patrón de navegación**: cabecera fija y modelo de anclas (`SiteHeader`, `PublicNav`, `sections.ts`).
5. **Formato del selector de idioma**: presentación compacta de `LanguageSwitcher` (comportamiento
   idéntico a F3).

## Migraciones

**Ninguna.** Si una tarea de F10 pidiera una migración, sería una señal de que el alcance se está
violando (FR-021) y debe escalarse antes de implementar.
