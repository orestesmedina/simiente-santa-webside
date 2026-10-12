# Tasks: Eventos y actividades (F4)

**Branch**: `004-eventos-actividades` | **Date**: 2026-10-11 | **Spec**: [spec.md](./spec.md)
(APROBADA v2, US1–US6 / FR-001…FR-024; Q1–Q7 resueltas el 2026-10-11) · **Plan**: [plan.md](./plan.md)
(P4-1…P4-17, puerta 2 aplicada) · **UX**: [ux.md](./ux.md) (D-1/D-6/D-13 confirmadas por el humano)

**Input**: plan aprobado (`plan.md`, P4-1…P4-17 + "Propuesta de tareas") · `research.md`
(R4-1…R4-15) · `data-model.md` (migraciones `000007`/`000008`/`000009`, tablas `agenda_*`,
ampliación de `admin_actions` y del catálogo `permissions`) · `contracts/openapi.yaml` (delta de
diseño, **inmutable**) · `ux.md` (pantallas, componentes, textos, estados) · `quickstart.md`
(§0…§14, verificación) · constitución (`.specify/memory/constitution.md`) · skills `go-backend`,
`postgres-db`, `react-frontend` · F1, F2, F3 y F10 ya integradas (`specs/001-estructura-base/`,
`specs/002-acceso-gestion-usuarios/` T201…T254, `specs/003-portada-info-general/` T301…T340,
`specs/010-sistema-diseno-movimiento/`).

> **Regla de ejecución**: **un commit por tarea** (Conventional Commits; los hooks del kit lo
> validan). **Test-first** (constitución §III): la prueba de cada tarea se escribe **antes** que su
> código y debe **fallar** antes de implementar; toda tarea de código incluye sus pruebas en el mismo
> commit — backend con `go test` y, donde toca persistencia real, `go test -tags=integration`
> (`//go:build integration`, PostgreSQL real con `testcontainers-go`); frontend con Vitest +
> Testing Library + MSW; flujos críticos con Playwright (`make e2e`). Cobertura exigida **≥ 80 % en
> `internal/agenda/service*.go`** (`go test -cover`). Quien escribe no aprueba: cada tarea pasa por
> `qa-tester`, `revisor-codigo` y `seguridad` antes de integrarse (flujo del orquestador). **El CI
> del kit (`.github/workflows/ci.yml`) no se toca** y **ninguna migración aplicada se edita** (§VI).

## Decisiones del plan (puerta 2, 2026-10-11) que estas tareas respetan (sin reabrir)

- **P4-3/R4-2 — `APP_TZ = America/Costa_Rica`** (valor firme): el corte de "hoy" del retiro
  automático se calcula en esa zona en el servidor (`today`/`now_time` van **parametrizados** al
  SQL, nunca concatenados); el service valida la zona IANA y **falla con error claro** si es
  inválida (fail-closed). Sin cron, sin jobs y sin cambiar estados: el retiro es un **filtro de
  lectura** "publicado y próximo" (FR-002/FR-006/Q3/Q4).
- **P4-5/R4-4/Q7 — permiso nuevo `categorias`** (10.º del catálogo de F2, **cambio de alcance
  autorizado**): `/api/v1/admin/categorias` se autoriza con `AdminChain("categorias")`; los permisos
  `eventos` y `actividades` **no** dan acceso al catálogo y los tres son independientes (FR-012).
  Se siembra en la migración `000009` (nunca se edita `000002`, ya aplicada). El administrador
  inicial lo recibe solo (el init de F2 lee el catálogo dinámico) y el formulario de roles de F2 lo
  ofrece solo. **Excepción aprobada (RG4-10)**: F4 actualiza deliberadamente las pruebas de F2 que
  fijan el catálogo en 9 → **10** (T405; no es debilitamiento: solo cambia el recuento).
- **P4-14/R4-13 — ruta pública `/eventos` y aviso sin redirigir**: sin nada publicado y próximo,
  `/eventos` muestra la página con el **aviso de sección vacía** («Por ahora no hay eventos ni
  actividades», copia de `ux.md` D-13 `HiddenNotice`) **sin redirigir** a la portada; el ítem
  «Eventos y actividades» (href `/eventos`) sigue **oculto** en barra, menú móvil y mapa del sitio.
  **Lectura de FR-013 «sin contenido»** (constancia, M6 de `analyze.md`; alineada con `ux.md`
  §2.5): la frase de la spec se lee como **«sin nada publicado y próximo»** (la única materia que
  la sección puede mostrar, FR-002/FR-006); con solo borradores o vencidos la sección **también**
  se oculta, porque la propia FR-013/SC-012 prohíben secciones vacías.
- **P4-11/R4-8 — `endTime` exige `startTime`** (decisión firme): un fin sin inicio se rechaza con
  `400 invalid` / `details.endTime`; y `endTime > startTime` cuando ambos existen (Q2).
- **P4-6/R4-5 — borrado de categoría en uso → `409 conflict`** (decisión firme, FK
  `ON DELETE RESTRICT`): no se desliga automáticamente; la interfaz indica "quítala primero"
  (`ux.md` D-16 se alinea a esto).
- **P4-7/R4-6 — imagen de elemento vencido no servida** (decisión firme): la descarga exige
  contenido **publicado y próximo**; la URL deja de responder `200` al vencer y **vuelve** al
  cargar una fecha próxima (el archivo no se borra). Misma regla que F3 para el borrador.
- **P4-1/P4-2/P4-8/P4-9/P4-10/P4-13**: dominios `internal/agenda/` + `internal/media/` (promovido de
  `portada`), modelos separados (`agenda_events` + `agenda_event_dates` / `agenda_activities`),
  un solo documento público `GET /api/v1/proximamente?lang=es|en` (`no-store`, claves omitidas si
  vacías), panel paginado con `platform/paginate` y `publication_state` por elemento con "se publica
  al guardar". **P4-16**: el delta de `contracts/openapi.yaml` se fusiona aditivamente en
  `backend/api/openapi.yaml` (`info.version` 0.4.0 → **0.5.0**); el snapshot de `specs/` no se edita.
- **P4-17 — dependencias nuevas: 0** (backend y frontend).

## Alineación de nomenclatura (decidida en la puerta 2 / `ux.md`; no se reabre)

1. **Rutas SPA**: la sección pública es **`/eventos`** (D-1; no `/eventos-actividades`) y el panel es
   **`/panel/eventos`**, **`/panel/actividades`** y **`/panel/categorias`** (`ux.md` §3).
2. **Carpetas de features** (nomenclatura de `ux.md` §6): público en
   **`frontend/src/features/publico/eventos/`** y panel en **`frontend/src/features/eventos/`**.
   Nombres de componentes y hooks los de `ux.md` §6.3/§6.4 (`EventsPage`, `EventsSection`,
   `ActivitiesSection`, `EventCard`, `ActivityCard`, `CategoryChip`, `DateBadge`, `MapLink`,
   `HiddenNotice`, `useUpcomingEvents`, `formatoFechas.ts`; panel: `EventsPage`, `ActivitiesPage`,
   `CategoriesPage`, `EventForm`, `ActivityForm`, `CategoryForm`, `DatesEditor`, `EventsList`,
   `ActivitiesList`, `messages.ts`).
3. **Rutas de API**: **no cambian** — son las del contrato (`/api/v1/proximamente`,
   `/api/v1/media/{fileName}`, `/api/v1/admin/{eventos,actividades,categorias}`); el snapshot
   `contracts/openapi.yaml` es inmutable (P4-16).
4. **Permisos**: códigos `eventos`, `actividades` y **`categorias`** (sin tilde: es el `code` del
   catálogo; la etiqueta es «Categorías de agenda»). `lib/permissions.ts` expone
   `EVENTOS`/`ACTIVIDADES`/`CATEGORIAS`; el guard es el `RequirePermission` existente (sin
   `RequireAnyPermission`: sobra con permiso propio).
5. **Numeración T401…T442** (T442 añadida al cerrar los hallazgos de `analyze.md`, sin renumerar
   T440/T441): patrón por funcionalidad (F1 `T0xx`, F2 `T2xx`, F3 `T3xx`).
6. **Copia del aviso** de sección vacía: la del humano —«Por ahora no hay eventos ni actividades»—
   (`ux.md` D-13 fija el componente `HiddenNotice`; si su copia difiere, gana la del humano).

## Leyenda

| Marca | Significado |
|---|---|
| `[backend]` `[frontend]` `[db]` `[infra]` | Capa y subagente responsable (`[backend]`/`[db]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` → `devops`) |
| `[P]` | **Paralelizable** con las otras tareas de su grupo (archivos distintos y sin dependencia entre ellas; ver "Grupos de paralelismo") |
| *(sin `[P]`)* | Secuencial: tiene dependencias previas que deben estar integradas |
| *(FR-…, US…)* | Requisitos de `spec.md` que cubre la tarea |

> **Orden de capas**: migraciones (`[db]`) → contrato → `platform/` → repository → service →
> handler/routes → frontend → e2e → cierre. Cada capa se construye y prueba con lo que necesita la
> siguiente (el service con fakes, el handler con service falso), manteniendo la dirección de
> dependencias **`handler → service → repository`** de §II. La **auditoría** se escribe junto a cada
> mutación (Fase 5–7) porque FR-017 exige que edición y registro sean atómicos. Las migraciones
> van **primero** (requisito de F4: `000007` → `000008` → `000009`, ninguna aplicada se edita).

## Tabla resumen

| ID | Tarea | Capa | Fase | Paralelo | Depende de |
|---|---|---|---|---|---|
| T401 | Migración `000007_create_agenda_content` (4 tablas + catálogo de categorías) | `[db]` | 1 | — | — |
| T402 | Migración `000008_extend_admin_actions_for_agenda` (+15 códigos) | `[db]` | 1 | — | T401 |
| T403 | Migración `000009_seed_permission_categorias` (permiso `categorias`) | `[db]` | 1 | — | T402 |
| T404 | Consultas sqlc `agenda.sql` + `media.sql` y generar `internal/db` | `[db]` | 1 | — | T401, T402, T403 |
| T405 | **Excepción aprobada**: pruebas de F2 del catálogo 9 → 10 (5 aserciones + fixture) | `[backend]` | 2 | — | T403 |
| T406 | Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.5.0) | `[backend]` | 3 | — | — |
| T407 | Regenerar `frontend/src/api/schema.d.ts` | `[frontend]` | 3 | — | T406 |
| T408 | `platform/audit`: +15 códigos `event.*`/`activity.*`/`category.*` | `[backend]` | 4 | P1 | — |
| T409 | `platform/config`: `APP_TZ=America/Costa_Rica` (zona IANA, fail-closed) | `[backend]` | 4 | P1 | — |
| T410 | `.env.example`: `APP_TZ=America/Costa_Rica` | `[infra]` | 4 | P1 | — |
| T411 | `agenda`: `model.go` + `repository.go` (pgtype, `withTx`, auditoría sobre el tx) | `[backend]` | 5 | — | T404, T408 |
| T412 | `agenda`: repository de eventos + fechas (`ReplaceEventDates`, duplicados) | `[backend]` | 5 | P2 | T411 |
| T413 | `agenda`: repository de actividades (CRUD + `UNIQUE` de duplicado) | `[backend]` | 5 | P2 | T411 |
| T414 | `agenda`: repository de categorías (CRUD + `CountCategoryUsage`) | `[backend]` | 5 | P2 | T411 |
| T415 | `media`: promoción de la descarga + `IsFilePublished` (regresión de F3) | `[backend]` | 5 | P2 | T404 |
| T416 | Integración: `000008`/`000009` up→down→up, regla "próximo" y atomicidad | `[backend]` | 5 | — | T412, T413, T414, T415 |
| T417 | `agenda`: `service.go` — invariantes, límites, `APP_TZ` y regla "próximo" | `[backend]` | 6 | — | T409, T411 |
| T418 | `agenda`: `service_public.go` — "lo que viene" (publicado+y próximo, idioma, orden) | `[backend]` | 6 | — | T417 |
| T419 | `agenda`: `service_admin.go` — eventos (CRUD + fechas + validación) | `[backend]` | 6 | — | T417 |
| T420 | `agenda`: `service_admin.go` — actividades | `[backend]` | 6 | — | T419 |
| T421 | `agenda`: `service_admin.go` — categorías (`409` en uso y duplicado) | `[backend]` | 6 | — | T420 |
| T422 | `agenda`: `service_audit.go` — `audit.Recorder` (denegaciones y rechazos) | `[backend]` | 6 | — | T417 |
| T423 | `handler_public.go` — `GET /api/v1/proximamente` | `[backend]` | 7 | — | T418, T422 |
| T424 | `handler_admin.go` — eventos | `[backend]` | 7 | — | T419, T423 |
| T425 | `handler_admin.go` — actividades | `[backend]` | 7 | — | T420, T424 |
| T426 | `handler_admin.go` — categorías (permiso `categorias`) | `[backend]` | 7 | — | T421, T425 |
| T427 | `handler_images.go` — subida por módulo (multipart) | `[backend]` | 7 | — | T419, T420, T426 |
| T428 | `routes.go` + `cmd/api/main.go` (3 × `AdminChain`) + humo | `[backend]` | 7 | — | T415, T423–T427 |
| T429 | `api/agenda.ts` + hooks (público, panel, `uploadImage`) | `[frontend]` | 8 | — | T407 |
| T430 | Permisos, rutas y accesos del panel: `CATEGORIAS`, `/eventos`, `/panel/*`, menú, sello retirado, `AccesosRapidos`, `ACTION_LABELS` | `[frontend]` | 8 | — | T429 |
| T431 | Promoción de `ImageUploader`/`PublishControls` al inventario único | `[frontend]` | 8 | P3 | — |
| T432 | `features/publico/eventos/`: `formatoFechas.ts` + claves i18n | `[frontend]` | 8 | P3 | — |
| T433 | `features/publico/eventos/`: `EventsPage`, bloques, tarjetas y `HiddenNotice` | `[frontend]` | 8 | P4 | T429, T432 |
| T434 | `useSiteLinks`: href `/eventos` e ítem oculto en barra/menú/pie | `[frontend]` | 8 | P4 | T429 |
| T435 | Panel: `EventsPage` + `EventsList` + `EventForm` + `DatesEditor` | `[frontend]` | 8 | — | T430, T431 |
| T436 | Panel: `ActivitiesPage` + `ActivitiesList` + `ActivityForm` | `[frontend]` | 8 | — | T435 |
| T437 | Panel: `CategoriesPage` + `CategoryForm` + borrado con `409` | `[frontend]` | 8 | — | T436 |
| T438 | E2E Playwright `agenda-publica.spec.ts` (US1, US5, US6) | `[frontend]` | 9 | P5 | T433, T434 |
| T439 | E2E Playwright `agenda-panel.spec.ts` (US2, US3, US4) | `[frontend]` | 9 | P5 | T435, T436, T437 |
| T442 | E2E Playwright `agenda-accesibilidad.spec.ts` (teclado y foco, SC-008) | `[frontend]` | 9 | P5 | T433, T434 |
| T440 | Verificación de cierre: `make ci`, `make e2e`, quickstart §0–§14, SC-010 | `[infra]` | 10 | — | todas |
| T441 | Documentación y PR (documentador + devops) | `[infra]` | 10 | — | T440 |

**Total: 42 tareas** — 22 `[backend]` · 13 `[frontend]` · 4 `[db]` · 3 `[infra]` · 14 marcadas `[P]`.

> **T442 se añadió** al cerrar los hallazgos de `analyze.md` (M2: el "e2e de teclado" que prometían
> plan y tabla de cobertura no tenía tarea) **sin renumerar T440/T441** (numeración estable): se
> ejecuta en el grupo P5, dentro de la Fase 9.

## Grupos de paralelismo

| Grupo | Tareas | Por qué pueden ir en paralelo |
|---|---|---|
| P1 | T408 · T409 · T410 | Paquetes/archivos distintos (`platform/audit`, `platform/config`, `.env.example`) sin dependencias entre sí |
| P2 | T412 · T413 · T414 · T415 | `repository_eventos.go`, `repository_actividades.go`, `repository_categorias.go` y el dominio `media/` son archivos distintos que parten de T411/T404 |
| P3 | T431 · T432 | Promoción de componentes (`src/components/`) y `features/publico/eventos/formatoFechas.ts` + i18n no se solapan |
| P4 | T433 · T434 | `features/publico/eventos/` y `hooks/useSiteLinks.ts` son archivos distintos (T433 espera además a T432; T434 solo a T429) |
| P5 | T438 · T439 · T442 | Tres specs e2e independientes (`agenda-publica.spec.ts`, `agenda-panel.spec.ts` y `agenda-accesibilidad.spec.ts`) |

---

## Fase 1 — Migraciones y capa de datos (`[db]`)

> Convenciones de la skill `postgres-db` y de `data-model.md` (fuente de verdad): `id UUID`,
> `created_at`/`updated_at`, índices y `CHECK`/`UNIQUE` **en la base**, `up`/`down` completos, SQL
> **parametrizado** (§IV, CWE-89) y **nunca** se edita una migración aplicada (§VI). Numeración sin
> huecos desde `000001` (F3 terminó en `000006`).

- [X] T401 · Migración `000007_create_agenda_content` · `[db]`

  - **Archivos**: `backend/migrations/000007_create_agenda_content.up.sql` y
    `000007_create_agenda_content.down.sql` (NUEVOS).
  - **Qué hace**: crea las **4 tablas** tal como `data-model.md`: `agenda_categories` (catálogo
    compartido: `name_es`/`name_en`, `sort_order`, `UNIQUE (name_es)`, R4-5/Q6),
    `agenda_events` (padre: título/frecuencia/detalle/lugar/mapa/organizador/categoría/imagen+alt/
    `publication_state`) + `agenda_event_dates` (hija: `event_date` + `start_time`/`end_time`
    opcionales, `UNIQUE (event_id, event_date)`, `CHECK (end_time IS NULL OR start_time IS NOT NULL)`
    y `CHECK (end_time IS NULL OR end_time > start_time)`) y `agenda_activities` (todo en una fila
    con `activity_date` + horario opcional, `UNIQUE (title_es, activity_date, place_es)`). Patrón
    bilingüe `*_es NOT NULL` / `*_en NULL` (FR-007, R3-1), FK `category_id … ON DELETE RESTRICT`
    (borrado de categoría en uso protegido, FR-012/Q7), `CHECK` de alt obligatorio con imagen
    (FR-020), límites (detalle ≤ 2.000, FR-018/FR-019) e índices (`created_at DESC, id DESC`,
    `event_date`, `activity_date`, `sort_order, id`). **Siembra el catálogo inicial** de 6 categorías
    con su traducción al inglés (Q6). El `down` tira las 4 tablas en orden inverso. *Cubre FR-004,
    FR-005, FR-018, FR-019 y el modelo de FR-012/FR-013 a nivel de datos.*
  - **Pruebas incluidas** (§III): — (SQL; su verificación ejecutable es T416). **Cómo se verifica**:
    `make db-migrate` aplica sin errores sobre la BD de F3 (`000006` → `000007`); `migrate … down`
    revierte sin errores; los `CHECK`/`UNIQUE` de `data-model.md` están literalmente en el SQL;
    ninguna tabla sin `id`/`created_at`/`updated_at`; las 6 categorías sembradas con `sort_order`
    que deja «Otro» al final.
  - **Criterio de terminado**: el esquema coincide con `data-model.md` (lo revisa `revisor-codigo`
    contra la skill `postgres-db`).
  - **Commit sugerido**: `feat(db): tablas de eventos, actividades y categorías (000007)`

- [X] T402 · Migración `000008_extend_admin_actions_for_agenda` · `[db]`

  - **Archivos**: `backend/migrations/000008_extend_admin_actions_for_agenda.up.sql` y
    `.down.sql` (NUEVOS).
  - **Qué hace**: amplía el registro de auditoría de F2 (FR-017, R4-10): sustituye
    `admin_actions_action_check` (nombre propio desde `000006`, verificado en F3) añadiendo los
    **15 códigos** —6 `event.*`, 6 `activity.*` y **3 `category.*`** (`category.create`,
    `category.update`, `category.delete`): los tres del catálogo de categorías **incluidos**—. El
    `target_kind` **no cambia** (los elementos de F4 son `content`, como los de F3). El `down`
    elimina las filas con códigos F4 y restaura el CHECK de F3 (destructivo a propósito, igual que
    el `down` de `000006`; documentado en `data-model.md`). *Cubre FR-017* a nivel de datos.
  - **Pruebas incluidas** (§III): — (SQL; su verificación ejecutable es T416). **Cómo se verifica**:
    `make db-migrate` aplica; los 15 códigos del `CHECK` coinciden uno a uno con la lista de
    `data-model.md` (los 3 `category.*` presentes); el registro de acciones queda **cerrado** en la
    BD.
  - **Criterio de terminado**: registro cerrado garantizado por la BD, sin tocar ninguna migración
    anterior (§VI).
  - **Commit sugerido**: `feat(db): ampliar admin_actions a eventos, actividades y categorías (000008)`

- [X] T403 · Migración `000009_seed_permission_categorias` · `[db]`

  - **Archivos**: `backend/migrations/000009_seed_permission_categorias.up.sql` y `.down.sql`
    (NUEVOS).
  - **Qué hace**: **siembra el permiso nuevo `categorias`** (cambio de alcance Q7/FR-012 aprobado
    por el humano el 2026-10-11): `INSERT INTO permissions (code, label) VALUES ('categorias',
    'Categorías de agenda')` — el **10.º** del catálogo de F2. Es solo un `INSERT`: **`000002` está
    aplicada y no se toca** (§VI). El `down` retira primero las asignaciones de **este** permiso en
    `role_permissions` (FK `ON DELETE RESTRICT`) y luego la fila (destructivo a propósito, como los
    `down` de `000006`/`000008`). *Cubre FR-012* (autorización del catálogo de categorías) a nivel
    de datos.
  - **Pruebas incluidas** (§III): — (SQL; su verificación ejecutable es T405 y T416). **Cómo se
    verifica**: `make db-migrate` aplica tras `000008`; `SELECT code FROM permissions` devuelve
    **10** filas con `categorias` entre ellas; `down` revierte dejando 9 (y sin filas huérfanas en
    `role_permissions`).
  - **Criterio de terminado**: el catálogo de F2 queda ampliado a 10 permisos por migración, con su
    reversa, sin editar ninguna migración previa.
  - **Commit sugerido**: `feat(db): permiso categorias en el catálogo de F2 (000009)`

- [X] T404 · Consultas sqlc `agenda.sql` + `media.sql` y generar · `[db]`

  - **Archivos**: `backend/internal/db/queries/agenda.sql` y `media.sql` (NUEVOS),
    `backend/internal/db/queries/home.sql` (EDITADO: se retira `IsHomeFilePublished`) y
    `backend/internal/db/*.go` (GENERADOS y commiteados: `make sqlc`).
  - **Qué hace**: escribe las consultas nombradas de `data-model.md`: `ListCategories`,
    `GetCategoryByID`, `Insert/Update/DeleteCategory`, `CountCategoryUsage`; `Insert/Update/Delete/
    GetEventByID`, `ReplaceEventDates` (`DELETE` + `INSERT` de la lista completa, R4-11),
    `ListEventDates`/`ListUpcomingEventDates`, `ExistsEventDuplicate`, `List/CountEvents`,
    `ListUpcomingEvents`; `Insert/Update/Delete/GetActivityByID`, `List/CountActivities`,
    `ListUpcomingActivities`; y en `media.sql` **`IsFilePublished`** (sustituye a
    `IsHomeFilePublished` de F3: `TRUE` solo si el nombre lo referencia contenido **publicado y
    próximo** — `home_identity` `published` o `agenda_*` `published` con fecha próxima, R4-6 —).
    Las variantes `…Upcoming` aplican la regla "próximo" con `sqlc.narg('today')`/`sqlc.narg(
    'now_time')` **parametrizados** (R4-2). Sin `SELECT *`, `ORDER BY` determinista (§8.1.7); la
    auditoría **reutiliza** `InsertAdminAction` de `queries/audit.sql` (sin escrituras sobre
    `admin_actions`, FR-025 de F2). *Cubre FR-001…FR-006, FR-013 y el camino de archivos de
    FR-013/FR-020* a nivel de acceso a datos.
  - **Pruebas incluidas** (§III): — (SQL anotado; su ejecución es T412–T416). **Cómo se verifica**:
    `make sqlc` genera sin errores y `make sqlc-verify` no detecta deriva; los tipos generados
    quedan commiteados; ninguna consulta de escritura sobre `admin_actions`.
  - **Criterio de terminado**: la capa generada compila y es la única fuente de SQL de los dominios
    `agenda` y `media` (T411–T415).
  - **Commit sugerido**: `feat(db): consultas sqlc de agenda y media`

---

## Fase 2 — Excepción aprobada: catálogo de permisos de F2 (9 → 10)

> **⚠️ EXCEPCIÓN APROBADA (RG4-10, decisión del humano el 2026-10-11)**: F4 **edita pruebas de F2
> (funcionalidad cerrada)** porque el catálogo de permisos pasa de 9 a 10 con `categorias`. Es un
> cambio de producto **autorizado**, no un debilitamiento: solo cambia el **recuento** del catálogo;
> ninguna aserción de comportamiento se relaja. La tarea queda **registrada como excepción** en el
> PR y la validan `revisor-codigo` y `seguridad`. El resto de F2 (servicios, handlers, UI) lee el
> catálogo **dinámicamente** y **no** se toca.

- [X] T405 · Pruebas de F2: catálogo de permisos 9 → 10 · `[backend]` *(excepción aprobada RG4-10)*

  - **Archivos**: `backend/internal/usuarios/repository_roles_integration_test.go` (línea 52),
    `service_init_integration_test.go` (línea 57), `service_init_test.go` (línea 199 y fixture
    `sampleCatalog()`, línea 145), `service_roles_integration_test.go` (línea 69) y
    `service_roles_test.go` (línea 464) (EDITADOS); comentarios de texto "los 9 permisos" en
    `backend/internal/usuarios/service_init.go` (línea 66) y en los propios tests (EDITADOS).
    *Nada más de F2 se toca.*
  - **Qué hace**: actualiza **deliberadamente** las **5 aserciones** que fijan el catálogo en 9
    (las 3 señaladas por el humano + las 2 detectadas en la revisión del plan) a **10**, y hace
    crecer la fixture `sampleCatalog()` a 10 entradas (incluida `categorias`) para que las unitarias
    reflejen el catálogo real. Verifica además que el **administrador inicial recibe `categorias`
    solo** (el init de F2 lee el catálogo dinámico: su rol «Administrador» nace con los 10
    permisos). *Cubre FR-012* (el permiso `categorias` existe y el arranque lo otorga).
  - **Pruebas incluidas** (§III): unitarias + integración. **Cómo se verifica**:
    `go test ./internal/usuarios/` y `go test -tags=integration ./internal/usuarios/` en verde con
    el catálogo en **10** (y `categorias` presente en `ListPermissions`); ninguna otra aserción de
    esas pruebas cambió (se revisa el diff: solo recuentos, fixture y comentarios).
  - **Criterio de terminado**: `make ci` pasa con el catálogo de 10; el PR registra la tarea como
    **excepción aprobada** (no debilitamiento) y `revisor-codigo`/`seguridad` la validan.
  - **Commit sugerido**: `test(usuarios): catalogo de permisos 10 con categorias (excepcion aprobada F4)`

---

## Fase 3 — Contrato OpenAPI y tipos

- [X] T406 · Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.5.0) · `[backend]`

  - **Archivos**: `backend/api/openapi.yaml` (EDITADO). *No* se toca
    `specs/004-eventos-actividades/contracts/openapi.yaml` (snapshot inmutable, P4-16).
  - **Qué hace**: funde **aditivamente** el delta aprobado: las **17** operaciones de F4
    (`GET /api/v1/proximamente?lang=`, las 12 de `/api/v1/admin/{eventos,actividades}` —CRUD,
    `PATCH` con reemplazo de `dates` y `POST …/imagenes`— y las 4 de
    `/api/v1/admin/categorias`) y sus schemas (`ProximamentePublica`, `EventoPublico`,
    `ActividadPublica`, `OcurrenciaPublica`, `CategoriaPublica`, `ImagenPublica`, `EventoAdmin*`,
    `ActividadAdmin*`, `CategoriaAdmin*`, `FechaAdmin`/`FechaInput`, `Date`, `ClockTime`),
    reutilizando `ErrorEnvelope`/`ErrorBody`, las respuestas reutilizables y `sessionCookie`.
    Aplica los cambios documentados: enum `action` de `AdminActionItem` **+ 15 códigos**, prosa de
    `GET /api/v1/media/{fileName}` (publicado **y próximo**) y **prosa del catálogo de permisos**
    (`GET /api/v1/admin/permisos` e init de F2 pasan de "9 permisos" a **10** con `categorias`; el
    catálogo **no** se modela como enum: se lee de la BD). Sube `info.version` a **0.5.0**. El
    registro de `error.code` **no cambia** (F4 emite `invalid`, `unauthenticated`, `forbidden`,
    `not_found`, `conflict`, `internal`). El delta incluye los cierres del análisis de coherencia
    (`analyze.md`, 2026-10-11): el tope de **200 elementos por bloque** del listado público (RG4-8)
    y `nextDate` en `required` de `EventoAdmin` (nullable). *Cubre el contrato de todos los FR*; garantiza FR-012
    (los tres permisos en las descripciones), FR-013 (ninguna operación pública devuelve borradores
    ni vencidos) y FR-017 (códigos `event.*`/`activity.*`/`category.*`).
  - **Pruebas incluidas** (§III): — (artefacto de contrato). **Cómo se verifica**: el YAML parsea
    sin errores; `/api/v1/proximamente` tiene `security: []` y las de `/api/v1/admin/*`
    `sessionCookie`; la prosa de `GET /api/v1/proximamente` y de `ProximamentePublica` documenta el
    tope de **200 elementos por bloque** (RG4-8) y `nextDate` figura en `required` de `EventoAdmin`
    (nullable, siempre presente); las operaciones de categorías citan el permiso `categorias`; ningún schema
    público contiene `publicationState` ni campos `*En` sin resolver; los `endTime` documentan que
    **exigen `startTime`** y el `409` de categoría en uso está declarado;
    `grep -n "version:" backend/api/openapi.yaml` → `0.5.0`; `make api-gen` completa sin errores.
  - **Criterio de terminado**: el contrato vivo contiene F1–F3 + F4 y es la única fuente de tipos
    del frontend (T407).
  - **Commit sugerido**: `docs(api): fusionar delta de F4 en el contrato vivo (0.5.0)`

- [X] T407 · Regenerar `frontend/src/api/schema.d.ts` · `[frontend]`

  - **Archivos**: `frontend/src/api/schema.d.ts` (GENERADO y commiteado).
  - **Qué hace**: ejecuta `make api-gen` (o `npm run api-gen`) contra **`backend/api/openapi.yaml`**
    (nunca contra el snapshot de `specs/`, P4-16) y commitea los tipos de F4 (`ProximamentePublica`,
    `EventoAdmin`/`EventoInput`/`EventoPatch`, `ActividadAdmin*`, `CategoriaAdmin*`,
    `FechaAdmin`/`FechaInput`, `ImageUploadResult`, `PublicationState`, `AdminActionItem` con los
    enums ampliados).
  - **Pruebas incluidas** (§III): — (artefacto generado). **Cómo se verifica**: regenerar no produce
    `git diff`; `npm run typecheck` pasa con los tipos nuevos.
  - **Criterio de terminado**: `schema.d.ts` reproducible desde el contrato y listo para
    `api/agenda.ts` (T429; regla anti-deriva de §8.1.5).
  - **Commit sugerido**: `build(frontend): regenerar schema.d.ts con los tipos de F4`

---

## Fase 4 — Núcleo de plataforma (`backend/internal/platform/`)

- [X] T408 · `platform/audit`: 15 códigos `event.*`/`activity.*`/`category.*` · `[backend]` `[P]` (P1)

  - **Archivos**: `backend/internal/platform/audit/audit.go` y `audit_test.go` (EDITADOS).
  - **Qué hace**: amplía el registro cerrado del plumbing de F2 (R4-10) con los **15 códigos**
    `event.create|update|delete|publish|unpublish|image.upload`,
    `activity.create|update|delete|publish|unpublish|image.upload` y
    `category.create|update|delete` (los 3 del catálogo incluidos) en `ActionCodes` y
    `ValidActionCode`. `TargetKindContent = "content"` **no cambia** (ya lo añadió F3). *Cubre
    FR-017*.
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/platform/audit/` en verde: cada código `event.*`/`activity.*`/`category.*`
    pasa `Validate` con objetivo `content` y `TargetLabel`; los códigos de F2/F3 siguen intactos
    (regresión); un código fuera de la lista falla.
  - **Criterio de terminado**: la lista cerrada coincide literalmente con el `CHECK` de la migración
    `000008` (T402) — lo comprueba la integración de T416.
  - **Commit sugerido**: `feat(audit): codigos event.*/activity.*/category.* para F4`

- [X] T409 · `platform/config`: `APP_TZ=America/Costa_Rica` · `[backend]` `[P]` (P1)

  - **Archivos**: `backend/internal/platform/config/config.go` y `config_test.go` (EDITADOS).
  - **Qué hace**: lee `APP_TZ` con el valor por defecto **`America/Costa_Rica`** (la zona del lugar,
    valor firme de la puerta 2 — R4-2/RG4-1) y la **valida como zona IANA** (`time.LoadLocation`)
    al arrancar: una zona inválida produce un **error claro** (fail-closed: mejor un fallo al
    arrancar que un corte silencioso del "hoy"). Expone la zona para que el service calcule
    `today`/`now_time` (T417). *Soporta FR-002/FR-006* (el corte del retiro automático) y §VII.
  - **Pruebas incluidas** (§III): unitarias. **Cómo se verifica**:
    `go test ./internal/platform/config/` en verde: por defecto carga `America/Costa_Rica`; una
    variable de entorno la sobrescribe; `APP_TZ=Marte/Central` (zona inválida) → error descriptivo
    al construir la config; `America/Costa_Rica` carga sin error (offset −06:00).
  - **Criterio de terminado**: ninguna zona horaria vive "decente" en el código; la lista coincide
    con `.env.example` (T410) y con `quickstart.md` §0.
  - **Commit sugerido**: `feat(config): APP_TZ=America/Costa_Rica para el corte de hoy`

- [ ] T410 · `.env.example`: `APP_TZ` · `[infra]` `[P]` (P1)

  - **Archivos**: `.env.example` (EDITABLE). *No se toca* `.github/workflows/ci.yml` ni ningún
    archivo del kit.
  - **Qué hace**: documenta **`APP_TZ=America/Costa_Rica`** (zona horaria del lugar para el corte de
    "hoy" del retiro automático; research R4-2) con su consumidor (`platform/config`, T409) y su uso
    en `quickstart.md` §0. Sin secretos (§IV): `APP_TZ` no lo es.
  - **Pruebas incluidas** (§III): — (documento de configuración). **Cómo se verifica**: ninguna
    variable sin consumidor ni consumidor sin variable; `cp .env.example .env` es usable tal cual;
    lo audita `seguridad`.
  - **Criterio de terminado**: el valor documentado es exactamente `America/Costa_Rica` (decisión
    firme, no un ejemplo).
  - **Commit sugerido**: `chore(env): documentar APP_TZ=America/Costa_Rica`

---

## Fase 5 — Dominios `agenda` y `media`: repository

- [X] T411 · `agenda`: `model.go` + `repository.go` · `[backend]`

  - **Archivos**: `backend/internal/agenda/model.go`, `repository.go`, `repository_test.go`
    (NUEVOS).
  - **Qué hace**: entidades del dominio (`Event`, `EventDate`, `Activity`, `Category` +
    `PublicationState`), DTOs con etiquetas `validate` y sufijos `Es`/`En` (R3-1), y la base del
    repository: constructor sobre `*pgxpool.Pool`, mapeos `pgtype` → dominio (§8.1.6), `withTx`
    (patrón de F2) y la **inserción de `admin_actions` dentro del `tx`** vía `InsertAdminAction`
    (mapeo `audit.Action` → `InsertAdminActionParams`; R4-10: sin que un dominio importe a otro,
    R2). *Soporta FR-007 (patrón bilingüe) y FR-017 (registro atómico)*.
  - **Pruebas incluidas** (§III): unitarias de los mapeos (`pgtype` ↔ dominio, nil/`Valid`) y del
    mapeo de acciones. **Cómo se verifica**: `go test ./internal/agenda/` en verde; `gofmt`/`go vet`
    limpios; ningún `pgtype` sale de `repository.go` (R4).
  - **Criterio de terminado**: el paquete compila con sus tipos y el tx de escritura queda listo para
    T412–T414.
  - **Commit sugerido**: `feat(agenda): modelo y base del repository con auditoria transaccional`

- [X] T412 · `agenda`: repository de eventos + fechas · `[backend]` `[P]` (P2)

  - **Archivos**: `backend/internal/agenda/repository_eventos.go` y
    `repository_eventos_integration_test.go` (NUEVOS; `//go:build integration`).
  - **Qué hace**: `InsertEvent`/`UpdateEvent`/`DeleteEvent`/`GetEventByID`,
    `ReplaceEventDates` (reemplazo **completo** de la lista en la transacción del guardado, R4-11),
    `ListEventDates`/`ListUpcomingEventDates` (las públicas solo próximas, R4-2),
    `ExistsEventDuplicate` (título+lugar normalizados + fecha en común, R4-9), `List/CountEvents`
    (panel paginado, `ORDER BY created_at DESC, id DESC`) y `ListUpcomingEvents` (`published` **y**
    `EXISTS` una fecha próxima; orden por la primera fecha próxima con desempate
    `title_es, id`). Borrado físico con `ON DELETE CASCADE` de las fechas. *Cubre FR-004, FR-010,
    FR-013, FR-018*.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/agenda/` en verde: `UNIQUE (event_id, event_date)` rechaza
    fecha repetida; `ReplaceEventDates` deja exactamente la lista pedida y borra las anteriores;
    `ExistsEventDuplicate` detecta el duplicado del módulo (y no el de otro módulo);
    `ListUpcomingEvents` **nunca** devuelve borradores ni eventos sin fecha próxima (fechas fijas de
    prueba); el borrado del evento deja las fechas huérfanas fuera.
  - **Criterio de terminado**: persistencia real de eventos con sus invariantes y su única vía
    pública (`…Upcoming`).
  - **Commit sugerido**: `feat(agenda): persistencia de eventos con lista de fechas`

- [X] T413 · `agenda`: repository de actividades · `[backend]` `[P]` (P2)

  - **Archivos**: `backend/internal/agenda/repository_actividades.go` y
    `repository_actividades_integration_test.go` (NUEVOS).
  - **Qué hace**: `InsertActivity`/`UpdateActivity`/`DeleteActivity`/`GetActivityByID`,
    `List/CountActivities` (panel paginado) y `ListUpcomingActivities` (`published` **y** fecha
    próxima; `ORDER BY activity_date, start_time NULLS FIRST, title_es, id`), con la traducción de
    restricciones a `apperr` (`conflict` para `UNIQUE (title_es, activity_date, place_es)`, R4-9).
    *Cubre FR-005, FR-011, FR-013, FR-019*.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/agenda/` en verde: el duplicado exacto del módulo falla con
    `conflict` (y una actividad idéntica a un **evento** sí se guarda — Decisión 3 —); los `CHECK`
    de horario (`end_time` exige `start_time` y es posterior) se disparan; `ListUpcomingActivities`
    sin borradores ni vencidos (fechas fijas).
  - **Criterio de terminado**: persistencia real de actividades con sus invariantes.
  - **Commit sugerido**: `feat(agenda): persistencia de actividades`

- [X] T414 · `agenda`: repository de categorías · `[backend]` `[P]` (P2)

  - **Archivos**: `backend/internal/agenda/repository_categorias.go` y
    `repository_categorias_integration_test.go` (NUEVOS).
  - **Qué hace**: `ListCategories`/`GetCategoryByID` (`ORDER BY sort_order, id`), `InsertCategory`,
    `UpdateCategory`, `DeleteCategory` y `CountCategoryUsage`. El `DELETE` falla con
    `FK RESTRICT` si está en uso → error traducido a **`conflict`** (con el recuento para el mensaje
    "está en uso por N elementos", R4-5/P4-6: borrado restringido, decisión firme); `UNIQUE
    (name_es)` → `conflict` en el duplicado. *Cubre FR-012/FR-018/FR-019* (categoría opcional del
    catálogo).
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/agenda/` en verde: borrar una categoría usada por un
    evento o una actividad → error `conflict` (y la fila sigue ahí); con `category_id IS NULL` en
    todos sus usos el borrado procede; `CountCategoryUsage` cuenta ambos módulos; nombre duplicado
    → `conflict`.
  - **Criterio de terminado**: catálogo protegido en uso por la BD y traducido a `409` en el
    dominio.
  - **Commit sugerido**: `feat(agenda): persistencia del catalogo de categorias`

- [X] T415 · `media`: promoción de la descarga + `IsFilePublished` · `[backend]` `[P]` (P2)

  - **Archivos**: `backend/internal/media/handler_media.go`, `repository.go`, `routes.go`,
    `*_test.go` (NUEVOS, movidos desde `backend/internal/portada/`); `backend/internal/portada/`
    (EDITADO: se retiran `handler_media.go` y la consulta `IsHomeFilePublished`).
  - **Qué hace**: el servicio de archivos deja de pertenecer a la portada y se promueve a
    **`internal/media/`** (P4-1/R4-6): mismo `GET /api/v1/media/{fileName}`, mismas cabeceras
    (`nosniff`, `inline`, `no-store`) y mismos errores (`400` si el nombre no cumple el patrón,
    `404` si es válido pero no existe o **no está referenciado por contenido publicado y próximo**).
    La consulta **`IsFilePublished`** (T404) cubre `home_identity` publicada (F3, sin cambio de
    comportamiento) **y** `agenda_events`/`agenda_activities` publicados **con fecha próxima** (R4-2:
    un vencido deja de servir su imagen sin borrarse y la URL vuelve al haber fecha próxima —
    decisión firme de la puerta 2). El handler y sus pruebas se **mueven sin reescribirse** (cambio
    de ubicación, no de comportamiento). *Cubre FR-013 (camino de archivos) y FR-020*.
  - **Pruebas incluidas** (§III): las pruebas de descarga de F3 **movidas** + casos nuevos.
    **Cómo se verifica**: `go test ./internal/media/ ./internal/portada/` y
    `go test -tags=integration ./internal/media/` en verde — **regresión de F3**: logo/imagen de
    portada publicada se sigue sirviendo igual y un borrador de portada sigue dando `404` (mismos
    casos que antes de la movida); **F4**: imagen de evento/actividad publicado **y próximo** →
    `200`; el mismo elemento con la fecha pasada → `404` sin borrar el archivo; al cargarle una
    fecha futura vuelve a `200` (RG4-2: los e2e de F3 `portada-imagenes` siguen en verde como
    evidencia).
  - **Criterio de terminado**: la descarga vive en `media/`, la regla "publicado y próximo" está
    verificada en ambos sentidos y F3 no regresiona.
  - **Commit sugerido**: `refactor(media): promover la descarga de archivos a internal/media`

- [X] T416 · Integración: migraciones, regla "próximo" y atomicidad · `[backend]`

  - **Archivos**: `backend/internal/agenda/repository_migrations_integration_test.go` y
    `repository_audit_integration_test.go` (NUEVOS; `//go:build integration`).
  - **Qué hace**: la verificación ejecutable de `data-model.md` §"Validación del modelo": (1) las
    migraciones `000008` y `000009` aplican **up → down → up** sin errores (restauración del
    `CHECK` de `admin_actions` y de la fila `categorias` con sus asignaciones); (2) el catálogo de
    `permissions` devuelve **10** con `categorias`; (3) la **regla "próximo"** con fechas fijas
    cubre los tres casos de R4-2 en las consultas `…Upcoming` e `IsFilePublished` (futuro siempre;
    hoy sin hora de fin: todo el día; hoy **con** fin ya terminado: fuera) y un evento solo aparece
    con ≥ 1 fecha próxima; (4) **mutación + `admin_actions` atómicas**: al forzar un fallo en el
    registro, la mutación **no se aplica** (edge case FR-017), y los 15 códigos del `CHECK` de la
    BD coinciden con `platform/audit` (T408). *Cubre FR-002, FR-006, FR-013 y FR-017*.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/agenda/` en verde con los cuatro bloques; sin cron ni
    procesos: el retiro es solo de lectura (el `publication_state` del elemento vencido sigue
    `published`, FR-006).
  - **Criterio de terminado**: la regla "próximo" (requisito 5 del plan) y la atomicidad quedan
    verificadas en la base; las migraciones son reversibles.
  - **Commit sugerido**: `test(agenda): integracion de migraciones, regla proximo y auditoria atomica`

---

## Fase 6 — Dominio `agenda`: servicios

- [X] T417 · `agenda`: `service.go` — invariantes, límites y regla "próximo" · `[backend]`

  - **Archivos**: `backend/internal/agenda/service.go` y `service_test.go` (NUEVOS).
  - **Qué hace**: interfaz `Repository` (definida por quien la consume, R3), constantes revisables
    (título 160, detalle **2.000**, lugar 200, mapa 500, organizador 160, frecuencia 160, alt 300,
    categoría 120, **≤ 50 fechas** por evento, ≤ 100 categorías y **200 elementos por bloque**
    `maxBlockItems` — tope de seguridad del listado público, RG4-8: lo aplica `GetProximamente`
    (T418) y **lo documenta el contrato** —), normalización (trim; `""`/espacios
    en `*_En` → `NULL`, `analyze` I6), y la **regla de "próximo"** (R4-2) como una sola función de
    dominio que calcula `today`/`now_time` con **`APP_TZ=America/Costa_Rica`** (T409) y resuelve los
    tres casos: fecha futura → próxima; hoy sin hora de fin → próxima todo el día; hoy con fin ya
    terminado → **no** próxima. Sin cron ni cambios de estado (FR-006/Q4). Validación de fechas
    reales (`YYYY-MM-DD`), horas `HH:MM`, `endTime > startTime` y **`endTime` exige `startTime`**
    (decisión firme: `400` con `details.endTime`). *Soporta FR-002, FR-005, FR-006, FR-015*.
  - **Pruebas incluidas** (§III): unitarias con fechas fijas (tabla de casos). **Cómo se verifica**:
    `go test ./internal/agenda/` en verde: los tres casos de la regla R4-2 exactos (sin depender de
    la hora real); `APP_TZ` inválida → error claro; `{"endTime":"21:00"}` sin inicio → `invalid`
    con `details.endTime`; `2026-13-32` → `invalid`; `""`/`"  "` en un `*En` → `NULL`;
    `maxBlockItems` es **200** (su prueba de comportamiento de truncado vive en T418).
  - **Criterio de terminado**: la regla "próximo" tiene **una** implementación y sus pruebas; los
    límites son el único punto de verdad (los usan T419–T421 y los cita el contrato).
  - **Commit sugerido**: `feat(agenda): invariantes, limites y regla de proximidad con APP_TZ`

- [X] T418 · `agenda`: `service_public.go` — "lo que viene" · `[backend]`

  - **Archivos**: `backend/internal/agenda/service_public.go` y `service_public_test.go` (NUEVOS).
  - **Qué hace**: `GetProximamente(lang)` (FR-001…FR-005): devuelve `{lang, events?, activities?}`
    con **solo** contenido `published` **y próximo** (consultas `…Upcoming`, R4-2: ninguna vía
    pública a borradores ni vencidos, FR-013/SC-002/SC-009), cada campo traducible resuelto con
    fallback **`en` si existe, si no `es`** (FR-008/SC-006; el cliente **no** reimplementa el
    fallback), categoría resuelta, `dates` del evento solo con sus fechas **próximas** (la primera
    es "la próxima fecha", FR-004), orden cronológico por la próxima fecha con desempate estable
    (`title`, `id`; FR-003 y el edge case "dos elementos en la misma fecha"), **claves de bloque
    omitidas si están vacías** (SC-012) y **cada bloque acotado a `maxBlockItems` (200)** — tope de
    seguridad de T417 (RG4-8): se sirven los 200 primeros ya ordenados y el contrato lo documenta —.
    `lang` fuera de `es|en` → `apperr.Invalid`. Evento sin
    fecha próxima: no aparece (hasta que se cargue una). *Cubre FR-001…FR-008*.
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos). **Cómo se verifica**:
    `go test ./internal/agenda/` en verde: con todo en borrador o vencido la respuesta solo lleva
    `lang`; los tres casos de la regla "próximo" filtrados; fallback por campo (nunca hueco);
    orden cronológico con empate estable; `dates` sin pasadas; con más de 200 elementos próximos en
    un bloque devuelve **200** (los primeros en orden cronológico, RG4-8); `lang=fr` → `invalid`;
    el DTO público
    **nunca** contiene `publicationState`.
  - **Criterio de terminado**: SC-002/SC-006/SC-009 verificables en el service; cobertura medida
    (`go test -cover`, ≥ 80 % al cierre).
  - **Commit sugerido**: `feat(agenda): lo que viene publicado y proximo con fallback en->es`

- [X] T419 · `agenda`: `service_admin.go` — eventos · `[backend]`

  - **Archivos**: `backend/internal/agenda/service_admin.go` y `service_admin_test.go` (NUEVOS).
  - **Qué hace**: `CreateEvento`/`UpdateEvento`/`DeleteEvento`/`ListEventos`/`GetEvento` (FR-010):
    validación FR-015 completa (obligatorios en español; `dates` de **1 a 50** sin repetir; cada
    fecha real con horario coherente — `endTime` exige `startTime` y es posterior —; `mapUrl` con
    tag `url`; `imageAltEs` obligatorio si hay imagen, FR-020; `categoryId` desconocido → `400`),
    normalización `*En` → `NULL`, duplicado del módulo → **`409`** (R4-9), fechas pasadas permitidas
    al crear/editar (edge case) pero el elemento no es público sin fecha próxima, `PATCH` con
    `dates` **reemplaza la lista completa** (R4-11), `publicationState` por elemento (FR-013) y
    "se publica al guardar" (FR-014), y **registro transaccional** `event.create|update|delete` con
    la regla exacta de códigos (R4-10): el alta usa `event.create` aunque nazca publicada;
    `event.publish`/`event.unpublish` **solo** para cambios de estado; datos y estado a la vez →
    **dos filas** (fail-closed). El panel marca `isPast`/`nextDate` sin tocar el estado (FR-006).
    *Cubre FR-004, FR-010, FR-013, FR-014, FR-015, FR-017, FR-018*.
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos). **Cómo se verifica**:
    `go test ./internal/agenda/` en verde: cada campo inválido → `400` con `details` y **sin
    guardar nada** (FR-015); `endTime` sin `startTime` → `details.endTime` (decisión firme); fecha
    repetida → `details.dates`; duplicado → `conflict`; `PATCH` publicado deja el cambio visible
    (FR-014); una edición que cambia datos y estado deja **dos filas** (R4-10); si el registro
    falla no se guarda; `""`/espacios en `*En` → `NULL`; tildes/`ñ`/`¿¡`/emojis intactos.
  - **Criterio de terminado**: US2 y FR-010/FR-015/FR-017 cubiertos en el service.
  - **Commit sugerido**: `feat(agenda): gestion de eventos con fechas, validacion y auditoria`

- [X] T420 · `agenda`: `service_admin.go` — actividades · `[backend]`

  - **Archivos**: `backend/internal/agenda/service_admin.go` (EDITADO) y `service_admin_test.go`
    (AMPLIADO).
  - **Qué hace**: `CreateActividad`/`UpdateActividad`/`DeleteActividad`/`ListActividades`/
    `GetActividad` (FR-011): mismas reglas que los eventos con el modelo de **fecha única**
    (Decisión 3/FR-005: una `date` con `startTime`/`endTime` opcionales, sin periodicidad ni lista),
    duplicado exacto del módulo (título+fecha+lugar) → **`409`** y un evento idéntico **no** cuenta
    como duplicado, publicación por elemento y registro transaccional `activity.*` con la regla de
    códigos (R4-10). *Cubre FR-005, FR-011, FR-013, FR-014, FR-015, FR-017, FR-019*.
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos). **Cómo se verifica**:
    `go test ./internal/agenda/` en verde: los mismos casos de validación que T419 sobre el modelo
    de fecha única; `endTime` sin inicio → `invalid`; duplicado → `conflict`; publicar/retirar por
    elemento sin tocar eventos; auditoría `activity.create|update|delete|publish|unpublish` con la
    regla de dos filas.
  - **Criterio de terminado**: US3 y FR-011 cubiertas en el service.
  - **Commit sugerido**: `feat(agenda): gestion de actividades con fecha unica y auditoria`

- [X] T421 · `agenda`: `service_admin.go` — categorías · `[backend]`

  - **Archivos**: `backend/internal/agenda/service_admin.go` (EDITADO) y `service_admin_test.go`
    (AMPLIADO).
  - **Qué hace**: `CreateCategoria`/`UpdateCategoria`/`DeleteCategoria`/`ListCategorias`
    (Q6/Q7): nombre en español obligatorio (≤ 120), en inglés opcional, `sortOrder` opcional,
    nombre duplicado → **`409`** (`UNIQUE (name_es)`) y **borrado en uso → `409`** con el mensaje
    "está en uso por N elementos" (P4-6: decisión firme, `CountCategoryUsage`; la interfaz dirá
    "quítala primero", `ux.md` D-16). Renombrar está permitido (el cambio se ve en lo publicado de
    inmediato). Registro transaccional `category.create|update|delete` (los 3 códigos, R4-10). El
    catálogo **no** tiene `publication_state` (son etiquetas, no contenido publicable). *Cubre
    FR-012/FR-018/FR-019* (categoría del catálogo).
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos). **Cómo se verifica**:
    `go test ./internal/agenda/` en verde: duplicado → `conflict`; borrado de categoría en uso →
    `conflict` **sin desligarla** de los elementos (decisión firme); borrado sin uso → `204` y
    auditoría `category.delete`; `""` en `nameEn` → `NULL`.
  - **Criterio de terminado**: Q6/Q7 y el `409` firme cubiertos en el service.
  - **Commit sugerido**: `feat(agenda): gestion del catalogo de categorias con 409 en uso`

- [X] T422 · `agenda`: `service_audit.go` — `audit.Recorder` · `[backend]`

  - **Archivos**: `backend/internal/agenda/service_audit.go` y `service_audit_test.go` (NUEVOS).
  - **Qué hace**: implementa `audit.Recorder` para el dominio (R4-10.4): `RecordDenied` resuelve
    **método+ruta → acción** de las rutas de F4 (tabla del dominio: `/api/v1/admin/eventos…` →
    `event.*`, `…/actividades…` → `activity.*`, `…/categorias…` → `category.*`, las subidas →
    `*.image.upload`) y la persiste con `result='denied'` best-effort; el helper de rechazo hace lo
    propio con `result='failure'`. Nunca transporta el cuerpo de la petición (FR-026 de F2).
    `target_label`: `Eventos · <título>`, `Actividades · <título>`, `Categorías · <nombre>`,
    `… · Imagen · <fileName>` (máx. 254). *Cubre FR-017 (y el patrón de denegación de FR-012)*.
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/agenda/` en verde: cada ruta del módulo resuelve a su acción y
    `targetLabel`; una ruta desconocida no deja fila; un fallo de registro no cambia la respuesta
    (best-effort).
  - **Criterio de terminado**: las denegaciones sobre las rutas de F4 quedan registradas igual que
    las de F2/F3 (SC-013).
  - **Commit sugerido**: `feat(agenda): registro de denegaciones y rechazos del modulo`

---

## Fase 7 — Handlers, rutas y cableado

- [X] T423 · `handler_public.go` — `GET /api/v1/proximamente` · `[backend]`

  - **Archivos**: `backend/internal/agenda/handler.go`, `handler_public.go`,
    `handler_public_test.go` (NUEVOS).
  - **Qué hace**: publica la operación del contrato: valida `lang` (`es|en`, por defecto `es`),
    llama a `GetProximamente` (T418) y responde `200` con `ProximamentePublica` y **`Cache-Control:
    no-store`** (SC-003). Sin autenticación (FR-001). El DTO público **nunca** contiene
    `publicationState`, borradores ni fechas pasadas; las claves de bloque vacías se omiten
    (SC-012). Helpers HTTP del dominio (decodificar/validar/responder con `apperr` + `WriteError`,
    patrón de F2). *Cubre FR-001…FR-003, FR-013*.
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/agenda/` en verde: `200` con el sobre correcto y `no-store`; `lang=fr` →
    `400 invalid` con `details.lang`; bloques ausentes cuando el service no los devuelve; error del
    service → sobre `ErrorEnvelope` genérico.
  - **Criterio de terminado**: `quickstart.md` §1 es ejecutable tal cual (US1).
  - **Commit sugerido**: `feat(agenda): endpoint publico de lo que viene`

- [X] T424 · `handler_admin.go` — eventos · `[backend]`

  - **Archivos**: `backend/internal/agenda/handler_admin.go`, `handler_admin_test.go` (NUEVOS).
  - **Qué hace**: `GET/POST /api/v1/admin/eventos`, `GET/PATCH/DELETE /api/v1/admin/eventos/{id}`
    sobre el service de T419: listado paginado (`limit`/`offset` con `platform/paginate`, filtro
    opcional `publicationState`), `201`/`200`/`204` según el contrato, `400` con `details` por
    campo, `404` por id inexistente y `409` en duplicado. El listado del panel trae `dates` (cada
    una con `isPast`), `nextDate` **siempre presente** (`null` si no hay fecha próxima; está en
    `required` del contrato) e `isPast` del elemento (FR-006/FR-010). *Cubre FR-010, FR-013, FR-015*.
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/agenda/` en verde: cada operación con sus códigos 200/201/204/400/404/409 y
    sobres uniformes; `details` por campo en los `400`; el DTO del panel incluye `isPast` y
    `nextDate` **siempre presente** (nullable: `null` sin fecha próxima, nunca campo ausente).
  - **Criterio de terminado**: `quickstart.md` §3/§7 ejecutables vía API (US2).
  - **Commit sugerido**: `feat(agenda): endpoints de panel de eventos`

- [X] T425 · `handler_admin.go` — actividades · `[backend]`

  - **Archivos**: `backend/internal/agenda/handler_admin.go` (EDITADO) y `handler_admin_test.go`
    (AMPLIADO).
  - **Qué hace**: `GET/POST /api/v1/admin/actividades`, `GET/PATCH/DELETE …/{id}` sobre el service
    de T420, con las mismas convenciones de T424 (paginación, `publicationState`, códigos y sobres).
    *Cubre FR-011, FR-013, FR-015*.
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/agenda/` en verde con los mismos casos de T424 aplicados al módulo.
  - **Criterio de terminado**: `quickstart.md` §4 ejecutable vía API (US3).
  - **Commit sugerido**: `feat(agenda): endpoints de panel de actividades`

- [X] T426 · `handler_admin.go` — categorías (permiso `categorias`) · `[backend]`

  - **Archivos**: `backend/internal/agenda/handler_admin.go` (EDITADO) y `handler_admin_test.go`
    (AMPLIADO).
  - **Qué hace**: `GET/POST /api/v1/admin/categorias`, `PATCH/DELETE …/{id}` sobre el service de
    T421: catálogo acotado (≤ 100, `ORDER BY sortOrder, id`, sin paginación —§8.1.3, excepción
    documentada—), `409` en duplicado **y en borrado en uso** (decisión firme) con mensaje
    ("está en uso por N elementos"). La operación queda tras el permiso **`categorias`** (Q7/FR-012:
    **no** `eventos` ni `actividades`; T428 cablea `AdminChain("categorias")`). *Cubre FR-012,
    FR-018/FR-019*.
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/agenda/` en verde: códigos 200/201/204/400/404/409; el `409` de borrado en
    uso con su mensaje; el `403` de permiso se prueba en T428 (cada grupo con su `AdminChain`).
  - **Criterio de terminado**: `quickstart.md` §5 ejecutable vía API con la cuenta `categorias`.
  - **Commit sugerido**: `feat(agenda): endpoints del catalogo de categorias`

- [X] T427 · `handler_images.go` — subida por módulo · `[backend]`

  - **Archivos**: `backend/internal/agenda/handler_images.go` y `handler_images_test.go` (NUEVOS).
  - **Qué hace**: `POST /api/v1/admin/eventos/imagenes` y `POST /api/v1/admin/actividades/imagenes`
    (multipart del contrato, R3-8 de F3): tope con `http.MaxBytesReader` (≤ 8 MB), tipo por
    **firma binaria** (solo JPEG/PNG/WebP — sin SVG ni GIF), nombre generado `img_<uuid>.<ext>` con
    `platform/storage` (reutilizado), **registrada con `event.image.upload`/`activity.image.upload`**
    (fail-closed: si el registro falla se aborta y se borra el archivo) y respuesta `201` con
    `ImageUploadResult`. Cada endpoint va tras **su** permiso de módulo (una cuenta solo de
    actividades no sube al área de eventos). *Cubre FR-020 y el edge case "archivo inválido o carga
    fallida"*.
  - **Pruebas incluidas** (§III): `httptest` con `Store` falso y `t.TempDir()`. **Cómo se verifica**:
    `go test ./internal/agenda/` en verde: sin archivo o campo erróneo → `400` con `details.file`;
    archivo > 8 MB → `400` con `details.fileSize`; `.svg`/HTML renombrado → `400` (la firma manda);
    cada subida deja su fila `*.image.upload` y un fallo de registro aborta sin dejar archivo; la
    subida **no** cambia contenido visible.
  - **Criterio de terminado**: `quickstart.md` §6 ejecutable y `seguridad` aprueba las reglas de
    subida (idénticas a las de F3).
  - **Commit sugerido**: `feat(agenda): subida de imagenes por modulo con auditoria`

- [X] T428 · `routes.go` + `cmd/api/main.go` (3 × `AdminChain`) + humo · `[backend]`

  - **Archivos**: `backend/internal/agenda/routes.go` (NUEVO), `backend/cmd/api/main.go` (EDITADO:
    DI de `agenda` y `media` + registro de `/api/v1/proximamente`, `/api/v1/media` y
    `/api/v1/admin/{eventos,actividades,categorias}`), `backend/cmd/api/main_test.go` (EDITADO).
  - **Qué hace**: cablea los tres grupos de panel con la cadena aprobada de F2 —
    `authn → guard de cambio de contraseña → authz(código) → CSRF` — **sin helpers nuevos de
    middleware**: `AdminChain("eventos")`, `AdminChain("actividades")` y **`AdminChain("categorias")`**
    (Q7/FR-012: el permiso `categorias` se verifica **en servidor**; los tres independientes entre
    sí, SC-005/SC-011). Las rutas públicas quedan sin middlewares (solo lectura) y `/healthz`
    intacto (§8.1.9). *Cubre FR-012, FR-013 (el servidor es la única autoridad) y FR-017*
    (el `Recorder` de T422 recibe cada cadena).
  - **Pruebas incluidas** (§III): humo de composición + de permisos. **Cómo se verifica**:
    `go test ./cmd/api/` en verde: rutas nuevas con su sobre; `/healthz` sin sesión y `404`
    `not_found` en rutas inexistentes (§8.1.9); una cuenta con `eventos` recibe `403` en
    `/api/v1/admin/categorias` y viceversa (403 genérico registrado como `denied`);
    `go test ./...` completo en verde.
  - **Criterio de terminado**: los tres permisos quedan verificados en servidor y auditables de un
    vistazo en `routes.go`.
  - **Commit sugerido**: `feat(agenda): rutas con AdminChain por permiso (eventos/actividades/categorias)`

---

## Fase 8 — Frontend

> Se **aplica** el sistema de F10 y los patrones de F3 (no se redefinen). Nombres de pieza y
> composición según `ux.md` §6; las pantallas y textos siguen `ux.md`. Ningún `fetch` fuera del
> cliente de API (regla F2); i18n solo en el sitio público (FR-016: el panel va en español).

- [X] T429 · `api/agenda.ts` + hooks · `[frontend]`

  - **Archivos**: `frontend/src/api/agenda.ts` (NUEVO) y `frontend/src/features/publico/eventos/hooks/useUpcomingEvents.ts`,
    `frontend/src/features/eventos/hooks/*` (NUEVOS).
  - **Qué hace**: cliente tipado desde `schema.d.ts` (T407): `getProximamente(lang)` (público),
    panel de eventos/actividades (list/create/update/delete) y categorías, y `uploadImage` por
    módulo; hooks TanStack Query (`useUpcomingEvents` con key `['proximamente', lang]`,
    `useEvents`/`useCreateEvent`/`useUpdateEvent`/`useDeleteEvent`, `useActivities…`,
    `useEventCategories`…, `useUploadAgendaImage`) siguiendo el patrón de F2/F3 (invalidación tras
    mutaciones para que el cambio sea visible en la primera carga, SC-003). *Soporta US1–US4*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: cada hook contra un handler MSW (éxito, `400` con `details`,
    `403`, `404`, `409`); la invalidación de caché tras crear/editar/eliminar.
  - **Criterio de terminado**: ningún componente hace `fetch` directo (regla F2) y los tipos salen
    solo del contrato vivo.
  - **Commit sugerido**: `feat(frontend): cliente de agenda y hooks de consulta`

- [ ] T430 · Permisos, rutas y accesos del panel: `CATEGORIAS`, `/eventos`, `/panel/*`,
  `AccesosRapidos`, `ACTION_LABELS` · `[frontend]`

  - **Archivos**: `frontend/src/lib/permissions.ts`, `frontend/src/features/roles/permissions.ts` y
    `permissions.test.ts`/`roles.test.tsx` (EDITADOS), `frontend/src/app/router.tsx`,
    `frontend/src/app/layout.tsx` (EDITADOS), `frontend/src/features/panel/components/
    AccesosRapidos.tsx`, `frontend/src/features/panel/pages/InicioPage.tsx` y
    `frontend/src/features/panel/panel.test.tsx` (EDITADOS), `frontend/src/features/auditoria/
    messages.ts` y su test (EDITADOS).
  - **Qué hace**: añade `EVENTOS`/`ACTIVIDADES`/`CATEGORIAS` (`'categorias'`) a `lib/permissions.ts`;
  retira el sello «Disponible más adelante» de los tres en `isPermissionAvailable` (y actualiza sus
    pruebas de F2 de frontend) para que el **formulario de roles de F2 ofrezca el permiso nuevo
    solo** (el catálogo lo lee dinámico de la API); registra las rutas **`/eventos`** (pública) y
    `/panel/eventos`, `/panel/actividades`, `/panel/categorias`, cada una con `RequirePermission`
    (**EVENTOS / ACTIVIDADES / CATEGORIAS**; sin `RequireAnyPermission`, sobra con permiso propio);
    las entradas del menú del panel se filtran por permiso («Categorías» solo con `categorias`;
    una cuenta solo `eventos` no la ve).
  - Amplía **`AccesosRapidos` de `InicioPage`** con las tres tarjetas de `ux.md` §2.3 paso 1 y §4.2
    —«Gestión de eventos» / «Gestión de actividades» / «Categorías», patrón ícono + texto
    existente—, cada una **solo** con su permiso (camino feliz de US2): con un solo permiso se ve
    solo su tarjeta y sin permisos de F4 no se ve ninguna.
  - Amplía **`ACTION_LABELS`** (`features/auditoria/messages.ts`, módulo de auditoría de F2) con
    etiqueta humana para los **15 códigos** `event.*`/`activity.*`/`category.*` (`ux.md` §4.9:
    «Creó/Modificó/Publicó/Retiró/Eliminó el {evento/actividad} "…"»), para que el panel no muestre
    códigos crudos. Los `home.*` de F3 siguen sin etiqueta: **deuda heredada**, fuera del alcance
    de F4. *Cubre FR-012 (visibilidad del acceso en el panel), FR-016, FR-017 (etiquetas legibles
    del registro de auditoría) y el inicio de US2*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library. **Cómo se verifica**:
    `npm test -- --run` en verde: `isPermissionAvailable('categorias') === true` y el sello fuera de
    los tres; el menú renderiza cada entrada solo con su permiso; una ruta forzada sin permiso cae
    en `/sin-permiso` de F2 (texto intacto); **cada tarjeta de `AccesosRapidos` se muestra solo con
    su permiso** (y ninguna con la cuenta sin permisos); los **15 códigos** `event.*`/`activity.*`/
    `category.*` tienen etiqueta humana en `ACTION_LABELS`.
  - **Criterio de terminado**: el acceso «Categorías» del panel es visible **solo** con el permiso
    `categorias` (requisito 3 del plan), la ruta pública es `/eventos`, el inicio del panel ofrece
    los tres accesos filtrados por permiso (`ux.md` §4.2) y la auditoría muestra etiquetas legibles
    para los 15 códigos de F4 (`ux.md` §4.9).
  - **Commit sugerido**: `feat(frontend): permisos categorias, rutas y accesos del panel`

- [X] T431 · Promoción de `ImageUploader`/`PublishControls` al inventario único · `[frontend]` `[P]` (P3)

  - **Archivos**: `frontend/src/components/ImageUploader.tsx`, `PublishControls.tsx` y sus tests
    (MOVIDOS desde `frontend/src/features/informacion/components/`); imports ajustados en
    `features/informacion/` (EDITADOS).
  - **Qué hace**: promueve los dos componentes compartidos al inventario único `src/components/`
    (regla de F2/T242: lo compartido se promueve en lugar de copiarlo) para que F4 los reutilice en
    eventos y actividades. **Sin cambios de comportamiento**: mismos props, mismos textos, mismas
    pruebas (se **mueven** con el componente). *Soporta FR-013/FR-014 (publicar/retirar) y FR-020
    (imagen + alt) en los módulos nuevos.*
  - **Pruebas incluidas** (§III): las pruebas existentes de los componentes **movidas** +
    ajuste de imports. **Cómo se verifica**: `npm test -- --run` en verde (regresión de F3: los
    formularios de `features/informacion/` siguen funcionando con los componentes ya promovidos).
  - **Criterio de terminado**: una sola copia de cada componente en el inventario y F3 sin
    regresión.
  - **Commit sugerido**: `refactor(frontend): promover ImageUploader y PublishControls al inventario`

- [X] T432 · `features/publico/eventos/`: `formatoFechas.ts` + claves i18n · `[frontend]` `[P]` (P3)

  - **Archivos**: `frontend/src/features/publico/eventos/formatoFechas.ts` (NUEVO),
    `frontend/src/features/publico/i18n/{messages,es,en}.ts` (EDITADOS).
  - **Qué hace**: `formatLongDate(lang)` y el reuso de `formatTimeRange` (`features/publico/
    schedule.ts`) para presentar fechas y horas **localizadas** con `Intl.DateTimeFormat` («viernes
    17 de octubre, 7:00 p. m.» / «Friday, October 17, 7:00 PM») **sin alterar el valor** de la API
    (`YYYY-MM-DD`/`HH:MM`; FR-009/R4-3), envueltas en `<time datetime>`. Añade las claves tipadas
    nuevas de `ux.md` (nav, títulos de bloque, etiquetas de fecha/categoría, estados vacíos,
    aviso de sección vacía): sin la clave, TypeScript **no compila** (garantía de SC-006).
    *Cubre FR-007, FR-009*.
  - **Pruebas incluidas** (§III): Vitest (tabla de casos por idioma). **Cómo se verifica**:
    `npm test -- --run` en verde: misma fecha/hora de entrada produce los dos formatos según idioma
    y **el valor real no cambia**; las claves `es`/`en` son simétricas (sin huecos).
  - **Criterio de terminado**: fechas legibles "escritas, no solo numéricas" (FR-024) y i18n
    completo.
  - **Commit sugerido**: `feat(frontend): formato de fechas localizado y claves i18n de agenda`

- [ ] T433 · `features/publico/eventos/`: `EventsPage`, bloques, tarjetas y `HiddenNotice` · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/eventos/EventsPage.tsx`, `EventsSection.tsx`,
    `ActivitiesSection.tsx`, `EventCard.tsx`, `ActivityCard.tsx`, `CategoryChip.tsx`, `DateBadge.tsx`,
    `MapLink.tsx`, `HiddenNotice.tsx` y `*.test.tsx` (NUEVOS).
  - **Qué hace**: la página **`/eventos`** dentro de `PublicLayout` con H1 «Eventos y actividades» y
    los dos bloques («Próximos eventos» / «Próximas actividades», FR-021/Q5) compuestos con el
    inventario de F10 (`Card`, `SectionHeading`, `Reveal`…): cada tarjeta con imagen+`alt`
    (FR-020), `CategoryChip`, `DateBadge` (fecha/hora localizadas con `<time>`, T432), lugar +
    `MapLink`, detalle, y en eventos su frecuencia y su lista de fechas próximas (FR-004). Solo se
    dibujan los bloques **con** elementos (SC-012) y **sin nada publicado y próximo se muestra
    `HiddenNotice`** — página con el aviso «Por ahora no hay eventos ni actividades» — **sin
    redirigir** a `/` (decisión firme de la puerta 2: `useNavigate`/`<Navigate>` **no** se usan
    aquí). Estados de carga (`SectionSkeleton`) y error (aviso con última versión válida, `ux.md`
    §4.1). *Cubre US1, US5, US6 y FR-001…FR-005, FR-008, FR-009, FR-013, FR-020, FR-021, FR-023,
    FR-024*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: dos bloques con contenido en orden cronológico (empate estable);
    fallback `en → es` visible (nunca campo vacío); con ambos bloques vacíos se pinta `HiddenNotice`
    con el aviso y **sin navegación/redirect** (se comprueba que la URL no cambia); `alt` en las
    imágenes; `<time datetime>` con el valor real; un elemento sin imagen se ve completo.
  - **Criterio de terminado**: `quickstart.md` §1/§8 ejecutables en navegador (US1/US5) y el aviso
    sin redirigir verificado (requisito 6 del plan).
  - **Commit sugerido**: `feat(frontend): pagina /eventos con bloques y aviso sin redireccion`

- [ ] T434 · `useSiteLinks`: href `/eventos` e ítem oculto sin contenido · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/hooks/useSiteLinks.ts` y su test (EDITADOS).
  - **Qué hace**: `SiteLink` crece con `href` de **ruta** (el hook ya estaba preparado) y añade la
    entrada única **«Eventos y actividades»** con href **`/eventos`** **solo si** el documento
    público (`useUpcomingEvents`, T429) trae algún bloque (FR-022): la consumen a la vez
    `PublicNav`, `MobileMenu` y `FooterSiteMap` (fuente única: barra, menú hamburguesa y mapa del
    sitio del pie **a la vez**, sin huérfanos ni enlaces rotos). Sin contenido próximo, el ítem
    **no existe** en ninguna de las tres superficies. *Cubre FR-021, FR-022 y el edge case
    "navegación sin enlaces rotos"*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: con contenido el ítem aparece en las tres superficies con
    `href="/eventos"`; sin contenido, en ninguna; el resto de enlaces de la portada (anclas
    `/#…`) no cambia.
  - **Criterio de terminado**: ocultación coherente del ítem (requisito 6 del plan) verificada en
    las tres superficies.
  - **Commit sugerido**: `feat(frontend): navegacion con /eventos condicional en barra, menu y pie`

- [ ] T435 · Panel: `EventsPage` + `EventsList` + `EventForm` + `DatesEditor` · `[frontend]`

  - **Archivos**: `frontend/src/features/eventos/pages/EventsPage.tsx`,
    `components/EventsList.tsx`, `components/EventForm.tsx`, `components/DatesEditor.tsx`,
    `messages.ts` y tests (NUEVOS).
  - **Qué hace**: gestión de eventos en el panel **en español** (FR-016) con la composición de
    `ux.md` §4.3/§5: listado (tabla en desktop / tarjetas en móvil) con próxima fecha, frecuencia,
    estado (`StatusPill` borrador/publicado) y marca «Pasado»; `EventForm` en `Dialog` (dos pestañas
    de idioma, `Select` de categoría, `ImageUploader` y `PublishControls` **promovidos** de T431)
    y `DatesEditor` para la lista de fechas (agregar/quitar filas, inputs nativos, validación por
    fila: fecha real, `HH:MM`, **fin exige inicio** y fin > inicio — decisión firme —, ≥ 1 fecha,
    ≤ 50, sin repetir). Errores junto al campo con foco en el primero (FR-015); guardar sin tocar
    el idioma inglés es válido (Decisión 8); confirmación para eliminar (sin papelera). Estados
    vacíos con el aviso de "nada publicado" de `ux.md` §4.7. *Cubre US2, US4 y FR-010, FR-013,
    FR-014, FR-015, FR-016, FR-018*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: alta con fechas válidas; `endTime` sin `startTime` muestra el
    error del campo (decisión firme); fecha repetida rechazada; edición publicada con aviso "ya
    visibles en la sección" (FR-014); publicar/retirar por elemento; eliminar con confirmación;
    errores del servidor (`409` duplicado) mostrados sin perder lo escrito.
  - **Criterio de terminado**: US2 ejecutable en < 2 min según `quickstart.md` §3 (SC-004).
  - **Commit sugerido**: `feat(frontend): panel de eventos con editor de fechas`

- [ ] T436 · Panel: `ActivitiesPage` + `ActivitiesList` + `ActivityForm` · `[frontend]`

  - **Archivos**: `frontend/src/features/eventos/pages/ActivitiesPage.tsx`,
    `components/ActivitiesList.tsx`, `components/ActivityForm.tsx` y tests (NUEVOS).
  - **Qué hace**: lo mismo para actividades (`ux.md` §4.4) con el modelo de **fecha única**
    (Decisión 3): `date` + `startTime`/`endTime` opcionales (mismas reglas de horario que T435),
    sin frecuencia ni lista de fechas; reutiliza `ImageUploader`/`PublishControls` y el patrón de
    formulario de T435 (`ItemFormFields` compartido). *Cubre US3, US4 y FR-011, FR-013, FR-014,
    FR-015, FR-016, FR-019*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde con los mismos casos de T435 aplicados al modelo de fecha única;
    duplicado exacto → mensaje de `409`; publicar/retirar solo toca la actividad.
  - **Criterio de terminado**: US3 ejecutable según `quickstart.md` §4 (SC-004).
  - **Commit sugerido**: `feat(frontend): panel de actividades con fecha unica`

- [ ] T437 · Panel: `CategoriesPage` + `CategoryForm` + borrado con `409` · `[frontend]`

  - **Archivos**: `frontend/src/features/eventos/pages/CategoriesPage.tsx`,
    `components/CategoryForm.tsx`, `components/DeleteConfirm.tsx` y tests (NUEVOS).
  - **Qué hace**: catálogo de categorías (`ux.md` §4.6) tras `RequirePermission(CATEGORIAS)`
    (T430): listado (≤ 100) con nombre es/en y orden, alta/edición en `Dialog` (nombre en español
    obligatorio, en opcional) y borrado con `ConfirmDialog`; si la categoría está en uso el servidor
    responde `409` y el diálogo **lo explica y no ofrece desligarla** — «Está en uso por N
    elementos: quítala primero» — (`ux.md` D-16 alineado a la decisión firme: la categoría en uso
    **no** se desliga automáticamente). *Cubre Q6/Q7, FR-012 y FR-018/FR-019* (categoría del
    catálogo).
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: crear/editar categoría; nombre duplicado → mensaje de `409`;
    borrado en uso → el diálogo indica "quítala primero" **sin** modificar los elementos; borrado
    válido con confirmación; sin permiso `categorias` la página no es accesible (T430).
  - **Criterio de terminado**: `quickstart.md` §5 ejecutable en el panel con la cuenta `categorias`
    y el `409` firme reflejado en la interfaz.
  - **Commit sugerido**: `feat(frontend): panel del catalogo de categorias con 409 en uso`

---

## Fase 9 — E2E

- [ ] T438 · E2E Playwright `agenda-publica.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/agenda-publica.spec.ts` (NUEVO).
  - **Qué hace**: recorrido del visitante (US1, US5, US6) según `quickstart.md` §1/§8/§9/§10: ve lo
    próximo con categoría, detalle e imagen en orden cronológico sin cuenta; un elemento con fecha
    de ayer **no** aparece (SC-009) y el evento con todas sus fechas pasadas tampoco; con
    `endTime` de hoy ya terminado desaparece solo; cambia a inglés (interfaz + contenidos con
    fallback) y vuelve; y **`/eventos` sin contenido muestra el aviso sin redirigir** (la URL se
    conserva; el ítem está oculto en barra, menú móvil y pie). Tres anchos (320/768/1280,
    SC-007/SC-012). *Cubre SC-001, SC-002, SC-006, SC-007, SC-009, SC-012*.
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde sobre la
    base con fixtures publicados, vencidos y en borrador (los borradores y sus imágenes **no**
    aparecen por ninguna vía, SC-002).
  - **Criterio de terminado**: US1/US5/US6 cubiertas de extremo a extremo y el aviso sin redirigir
    verificado (requisito 6 del plan).
  - **Commit sugerido**: `test(e2e): seccion publica /eventos (lo que viene, idioma y aviso)`

- [ ] T439 · E2E Playwright `agenda-panel.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/agenda-panel.spec.ts` (NUEVO).
  - **Qué hace**: recorrido del panel (US2, US3, US4) según `quickstart.md` §2–§7/§11/§12: login con
    permiso, crear+publicar un evento con varias fechas y una actividad en **< 2 min** (SC-004),
    editar y ver el cambio en público desde la primera carga (SC-003), retirar/eliminar con
    confirmación, gestión de categorías con la cuenta `categorias`, y **denegaciones cruzadas**
    (cuenta solo `eventos` frente a actividades **y** frente a categorías, y al revés: SC-005,
    SC-011 — la API forzada responde `403`) con todo registrado en la auditoría de F2 (SC-013).
    *Cubre SC-003, SC-004, SC-005, SC-011, SC-013*.
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde con
    cuentas de los tres permisos y la cuenta sin permisos; la auditoría muestra los códigos
    `event.*`/`activity.*`/`category.*` esperados **con su etiqueta humana** (`ACTION_LABELS` de
    T430, `ux.md` §4.9).
  - **Criterio de terminado**: US2/US3/US4 cubiertas de extremo a extremo con los tres permisos
    independientes.
  - **Commit sugerido**: `test(e2e): gestion de agenda en el panel y permisos cruzados`

- [ ] T442 · E2E Playwright `agenda-accesibilidad.spec.ts` (teclado y foco) · `[frontend]` `[P]` (P5)

  > *Tarea añadida al cerrar los hallazgos de `analyze.md` (M2: el "e2e de teclado" que prometían
  > plan y tabla de cobertura no tenía dueño). Se numera al final para no renumerar T440/T441.*

  - **Archivos**: `frontend/e2e/agenda-accesibilidad.spec.ts` (NUEVO).
  - **Qué hace**: recorrido por **teclado** de la sección pública `/eventos` (FR-024, SC-008,
    quickstart §10.3) con el patrón del precedente de F3 (`portada-accesibilidad.spec.ts`): el
    primer `Tab` enfoca el enlace de salto al contenido; la travesía con `Tab` alcanza navegación,
    tarjetas de cada bloque, `MapLink` y selector de idioma en orden comprensible; **foco visible
    en cada paso**; activación con `Enter`/`Espacio`; **sin trampas de foco** (`Shift+Tab` sale de
    cualquier panel); el selector de idioma se opera con teclado (radios con flechas) sin perder el
    foco. Se comprueba en **tema claro y oscuro** (F10). *Cubre SC-008 y FR-024 (teclado y foco; la
    checklist WCAG completa la cierra T440)*.
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde sobre la
    base con fixtures publicados y, con la sección vacía, el aviso `HiddenNotice` también es
    alcanzable y operable con teclado (sin redirigir, T433).
  - **Criterio de terminado**: el "e2e de teclado" que prometen `plan.md` (FR-024) y la tabla de
    cobertura existe y corre en `make e2e` junto a las otras specs.
  - **Commit sugerido**: `test(e2e): accesibilidad por teclado de /eventos (SC-008)`

---

## Fase 10 — Cierre

- [ ] T440 · Verificación de cierre · `[infra]`

  - **Archivos**: ninguno (evidencia en el PR). *No se toca* `.github/workflows/ci.yml` ni ningún
    archivo del kit.
  - **Qué hace**: ejecuta la verificación completa de F4 y la registra como evidencia: `make ci`
    (**lint + test + security**; `make db-migrate` se ejecuta al preparar el entorno, §0),
    `make e2e`, `make sqlc-verify` y `make api-gen` sin deriva, y el recorrido manual de
    `quickstart.md` §0–§14 (mapa §14: SC-001…SC-013 y FR-015/FR-018–FR-024). Comprueba además que
    `GET /healthz` responde como en F1, que una ruta inexistente responde `404` con `ErrorEnvelope`
    (§8.1.9), que la cobertura de `internal/agenda/service*.go` es **≥ 80 %** (`go test -cover`),
    que las migraciones `000007`–`000009` aplican y revierten (up/down/up) y que el catálogo de
    permisos tiene **10** entradas con `categorias`. La **checklist de accesibilidad** (quickstart
    §10, WCAG 2.1 AA, claro y oscuro) la recorre `qa-tester`. **Prueba de usabilidad SC-010**
    (protocolo mínimo, estilo F3/M7): **≥ 6 personas** (2 por franja 18–35 / 36–59 / 60+, incluidas
    con poca experiencia en internet), guion de **3 tareas sin ayuda** sobre la sección real
    (decir qué eventos vienen y cuándo es el próximo; encontrar dónde y a qué hora es una actividad
    concreta; cambiar a inglés y volver), sesión breve observada por `qa-tester` con el humano;
    resultado documentado en **`specs/004-eventos-actividades/pruebas-usabilidad-SC-010.md`**
    (umbral: ≥ 90 % de tareas completadas). Es una prueba **manual con personas** (no
    automatizable) y queda fuera de `make ci`/`make e2e`.
  - **Pruebas incluidas** (§III): — (es la ejecución de todas las anteriores). **Cómo se verifica**:
    cada comando anterior en verde y su salida adjunta al PR; el mapa de criterios → secciones queda
    recorrido completo y el informe de usabilidad de SC-010 adjunto.
  - **Criterio de terminado**: `make ci` y `make e2e` en verde; SC-001…SC-013 evidenciados; la
    **excepción aprobada** de T405 (pruebas de F2 9 → 10) queda registrada en la evidencia.
  - **Commit sugerido**: — (sin cambios de código)

- [ ] T441 · Documentación y PR · `[infra]`

  - **Archivos**: `README.md` y `docs/` (EDITADOS por `documentador`); PR en GitHub (creado por
    `devops`).
  - **Qué hace**: cierra la entrega con `documentador` (README del proyecto y notas de entrega:
    qué agrega F4, el permiso nuevo `categorias` y su migración `000009`, la variable
    `APP_TZ=America/Costa_Rica`, la ruta pública `/eventos` y la documentación de API actualizada
    tras la fusión del contrato) y con `devops` (rama → PR con descripción, tareas cubiertas y la
    **excepción aprobada de F4 sobre pruebas de F2** documentada explícitamente). *Cubre el cierre
    de US1–US6/FR-001…FR-024 como entregable.*
  - **Pruebas incluidas** (§III): — (documentación y proceso). **Cómo se verifica**: la documentación
    refleja el comportamiento real (la prueba es el `quickstart.md` ejecutado en T440); el PR
    referencia este `tasks.md` con las tareas marcadas y pasa por la revisión final (`qa-tester`,
    `revisor-codigo`, `seguridad`).
  - **Criterio de terminado**: PR abierto y **aprobado por el humano** (regla 1: merge y producción
    requieren aprobación explícita). El merge lo ejecuta el orquestador con `devops`.
  - **Commit sugerido**: `docs: notas de entrega de F4 (eventos y actividades)`

---

## Cobertura de requisitos (US1–US6, FR-001…FR-024)

| Requisito | Tareas | Verificación principal |
|---|---|---|
| **US1** Descubrir lo que viene | T404, T417, T418, T423, T429, T432, T433, T434, T438 | quickstart §1 · e2e T438 (SC-001, SC-009) |
| **US2** Gestionar eventos | T412, T419, T424, T427, T430, T435, T439 | quickstart §3 · e2e T439 (SC-004, SC-005) |
| **US3** Gestionar actividades | T413, T420, T425, T436, T439 | quickstart §4 · e2e T439 (SC-004, SC-005) |
| **US4** Controlar la publicación | T418–T420, T431, T435–T437, T439 | quickstart §7 · e2e T439 (SC-002, SC-003, SC-012) |
| **US5** Ver en inglés | T418, T432, T433, T438 | quickstart §8 · e2e T438 (SC-006) |
| **US6** Dispositivo y accesible | T433, T434, T438, T442, T440 | quickstart §10 · e2e en 3 anchos (SC-007) · e2e de teclado T442 (SC-008) |
| FR-001 (sección pública, dos conceptos) | T404, T418, T423, T433 | quickstart §1 · e2e T438 |
| FR-002 (solo lo próximo, orden cronológico) | T404, T417, T418, T423 | pruebas de la regla R4-2 (SC-009) |
| FR-003 (orden por próxima fecha, sin filtros) | T418, T423, T433 | pruebas de orden del service |
| FR-004 (evento con fechas + frecuencia) | T401, T412, T417, T418, T433, T435 | quickstart §3 · pruebas de service/repository |
| FR-005 (actividad fecha única + horario) | T401, T413, T417, T436 | quickstart §4 · pruebas de validación (Q2) |
| FR-006 (retiro automático sin cambiar estado) | T404, T416, T417, T418, T433, T438 | quickstart §9 · e2e con fechas pasadas (SC-009) |
| FR-007 (contenido es/en) | T401, T411, T432, T433 | quickstart §5 y §8 |
| FR-008 (fallback `en → es`, nunca vacío) | T418, T433 | tabla de casos del service (SC-006) |
| FR-009 (fechas/horas localizadas sin alterar el valor) | T432, T433 | quickstart §8 · e2e T438 |
| FR-010 (gestión de eventos) | T412, T419, T424, T435 | quickstart §3–§5 · e2e T439 (SC-004) |
| FR-011 (gestión de actividades) | T413, T420, T425, T436 | quickstart §4–§6 · e2e T439 (SC-004) |
| FR-012 (permisos `eventos`/`actividades`/`categorias` independientes) | T403, T405, T426, T428, T430, T437, T439 | quickstart §2/§11 · pruebas de handler (403 cruzado) · e2e (SC-005, SC-011) |
| FR-013 (borrador/publicado; nada de borradores ni vencidos al público; «sin contenido» = «sin nada publicado **y próximo**») | T404, T415, T416, T418–T420, T423, T433, T434, T435–T437 | pruebas "0 borradores/vencidos expuestos" · e2e T438 (SC-002, SC-012) |
| FR-014 (se publica al guardar) | T419, T420, T435, T436 | e2e T439 (SC-003) |
| FR-015 (validación completa; sin cambios parciales) | T417, T419–T421, T424–T427 | tablas de casos (service/handler) · quickstart §3/§4 |
| FR-016 (panel en español) | T435–T437 | revisión de `revisor-codigo` (0 cadenas del panel por `t()`) |
| FR-017 (auditoría de F2, atómica, solo lectura) | T402, T408, T411, T416, T419–T422, T427, T430 (etiquetas legibles) | integración (mutación+registro) · quickstart §12 (SC-013) |
| FR-018 (contenido del evento + categoría opcional) | T401, T412, T414, T419, T421, T424, T435 | quickstart §3/§5 · pruebas de persistencia |
| FR-019 (contenido de la actividad + categoría opcional) | T401, T413, T414, T420, T421, T425, T436 | quickstart §4/§5 · pruebas de persistencia |
| FR-020 (imagen opcional con alt) | T401, T415, T427, T431, T433, T435/T436 | quickstart §6 · pruebas de subida/descarga (SC-008 parcial) |
| FR-021 (F10 aplicado, una entrada, dos bloques, portada intacta) | T433, T434 | quickstart §1 · revisión de `revisor-codigo` sobre `PublicLayout`/`HomePage` sin cambios |
| FR-022 (navegación oculta sin contenido) | T429, T433, T434, T438 | quickstart §7 · e2e T438 (SC-012) |
| FR-023 (responsividad y navegación simple) | T433, T434, T438 | quickstart §10 · e2e en 320/768/1280 (SC-007) |
| FR-024 (accesibilidad y fechas legibles) | T432, T433, T442, T440 | checklist WCAG de quickstart §10 · e2e de teclado T442 (SC-008) |

También quedan cubiertos los **Edge Cases** de la spec sin FR propio: **página sin contenido con
aviso, sin redirigir** (T433, T438; y el ítem oculto en nav/menú/pie: T434) · **imagen de elemento
vencido no servida y restituida al haber fecha próxima** (T415, T416) · **`endTime` exige
`startTime`** (T417, T419, T420, T435) · **borrado de categoría en uso → `409`** (T414, T421, T437)
· **duplicado exacto dentro del módulo → `409`** (T412, T413, T421) · **dos elementos en la misma
fecha** en orden estable (T418) · **texto con tildes/`ñ`/`¿¡`/emojis intacto** (T419–T421) ·
**archivo inválido no rompe la sección** (T427, T433) · **fechas pasadas permitidas al cargar**
(T419) · **ediciones simultáneas** (última escritura completa gana: T412) · **sin permiso por
cualquier vía, incluido el cruce y el catálogo de categorías** (T426, T428, T430, T439) · **nada se
aplica sin registro** (T416).

**Excepción aprobada registrada**: T405 actualiza las 5 pruebas de F2 que fijaban el catálogo de
permisos en 9 a **10** (más la fixture `sampleCatalog()`), por el cambio de alcance Q7/FR-012
autorizado por el humano el 2026-10-11 (RG4-10: cambio de producto, **no** debilitamiento).
