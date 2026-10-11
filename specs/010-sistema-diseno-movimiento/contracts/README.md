# Contracts: F10 — Sistema de diseño y movimiento

**Feature**: F10 · **Branch**: `010-sistema-diseno-movimiento` · **Fecha**: 2026-10-10

## Sin cambios en la API

**F10 no modifica ni añade ningún endpoint.** No hay delta de OpenAPI para esta funcionalidad:

- `backend/api/openapi.yaml` **no cambia** (ni rutas, ni esquemas, ni `info.version`).
- No hay tareas de fusión de contrato ni de `make api-gen`: los tipos de `frontend/src/api/schema.d.ts`
  se quedan como están.
- El frontend sigue consumiendo exactamente los mismos endpoints de F3
  (`GET /api/v1/portada?lang=…`, `GET /api/v1/media/{fileName}`) con los mismos DTOs.

**Por qué** (ver `data-model.md` y la spec): F10 es de presentación (FR-020, FR-021). Cambia cómo se
ve, se mueve y se tematiza el sitio público —cómo se navega y cómo se muestra el selector de idioma y
el control de tema—, nunca qué datos se piden ni cómo se publican. El control de tema es un control
de presentación del propio navegador (persistencia local), sin llamadas al servidor.

## Frontera con futuras funcionalidades

- **F10B (panel)**: cuando se rediseñe el panel, tampoco debería necesitar cambios de contrato (es
  visual); si los necesitara, se documentarían allí.
- **F4–F9**: sus secciones sí añadirán endpoints propios; este documento de F10 no los anticipa.

## Comportamiento preservado (referencia para pruebas de no-regresión)

Los contratos de comportamiento que F10 debe conservar intactos (SC-008) son los de F3: rutas `/` y
`/health`, bilingüismo con fallback `en → es` resuelto por el servidor, borrador/publicado, permisos
y auditoría del panel. Ninguno se renegocia aquí.
