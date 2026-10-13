# Tasks: Ministerios (F6)

**Branch**: `006-ministerios` | **Date**: 2026-10-13
| **Spec**: [spec.md](./spec.md)
(APROBADA v2, US1–US6 / FR-001…FR-029 / SC-001…SC-017 —línea base aprobada—; Q1–Q9 resueltas el
2026-10-13; puerta 2 del 2026-10-13 aplicada: **«Oculto por defecto»** y **`/ministerios` redirige a
la portada, silenciosa y sin aviso**) · **Plan**: [plan.md](./plan.md)
(P6-1…P6-21 —P6-21 nacida en la puerta 2—, riesgos RG6-1…RG6-13) · **UX**: [ux.md](./ux.md)
(D-1…D-14; D-3/D-13/D-14 confirmadas por el humano)

**Input**: plan aprobado (`plan.md`, P6-1…P6-21 + "Propuesta de tareas") · `research.md`
(R6-1…R6-17) · `data-model.md` (migraciones `000010`/`000011`, tabla `ministries` con
`leader_visible DEFAULT FALSE`, ampliación del `CHECK` de `admin_actions` con 8 códigos
`ministry.*`; **sin migración de permisos**: `ministerios` ya está sembrado por `000002`) ·
`contracts/openapi.yaml` (delta de diseño, inmutable; **8 operaciones**, `info.version` 0.7.0) ·
`ux.md` (pantallas, componentes, textos, estados, D-1…D-14) · `quickstart.md` (§0…§16, verificación)
· constitución (`.specify/memory/constitution.md`) · skills `go-backend`, `postgres-db`,
`react-frontend` · F1, F2, F3, F4 y F10 ya integradas (`specs/001-estructura-base/`,
`specs/002-acceso-gestion-usuarios/` T201…T254, `specs/003-portada-info-general/` T301…T340,
`specs/004-eventos-actividades/` T401…T449, `specs/010-sistema-diseno-movimiento/`).

> **Regla de ejecución**: **un commit por tarea** (Conventional Commits; los hooks del kit lo
> validan). **Test-first** (constitución §III): la prueba de cada tarea se escribe **antes** que su
> código y debe **fallar** antes de implementar; toda tarea de código incluye sus pruebas en el mismo
> commit — backend con `go test` y, donde toca persistencia real, `go test -tags=integration`
> (`//go:build integration`, PostgreSQL real con `testcontainers-go`); frontend con Vitest +
> Testing Library + MSW; flujos críticos con Playwright (`make e2e`). Cobertura exigida **≥ 80 % en
> `internal/ministerios/service*.go`** (`go test -cover`). Quien escribe no aprueba: cada tarea pasa
> por `qa-tester`, `revisor-codigo` y `seguridad` antes de integrarse (flujo del orquestador). **El
> CI del kit (`.github/workflows/ci.yml`) no se toca** y **ninguna migración aplicada se edita**
> (§VI). **Sin excepciones de backend**: F6 no toca pruebas de F2 (el catálogo de permisos **no**
> cambia); solo la actualización **acotada** de las 4 pruebas de frontend que fijan permisos/accesos
> (T626, RG6-6).

## Decisiones del plan (puerta 2, 2026-10-13) que estas tareas respetan (sin reabrir)

- **P6-3 / puerta 2-A — `leader_visible DEFAULT FALSE`, «Oculto por defecto»** (privacidad por
  defecto): el ministerio nuevo **nace con la ficha del encargado —nombre y contacto— sin
  publicar**, aunque el formulario lleve encargado; el equipo enciende el interruptor
  (`leaderVisible`) **expresamente, por ministerio**. Se fija por prueba en **las cuatro capas**
  (T611 repository, T613/T615 service, T617 handler, T633/T634 e2e) y en el formulario (T631,
  `ShowContactSwitch` apagado al crear). El interruptor es **propio de F6**: F3/F4 no lo tienen y no
  se copia de ellos ninguna regla de visibilidad.
- **P6-2 / RG6-1 — privacidad del encargado en tres capas** (FR-026/SC-014, **no negociable**):
  (1) el DTO público **no tiene campos** para el dato oculto; (2) las consultas públicas devuelven
  `NULL` con `CASE WHEN leader_visible` (T609); (3) el DTO del bloque de la portada **ni siquiera
  tiene campos de encargado** (T614). Con `leaderVisible` apagado **ni el nombre ni el
  celular/correo salen del servidor** por ninguna vía; con él apagado el sitio público **no muestra
  línea de encargado** (ni vacía ni con solo el nombre). Pruebas «**0 apariciones**» en las cuatro
  capas (T609, T613, T617, T633 — con inspección de tráfico y de HTML —).
- **P6-21 / puerta 2-B — `/ministerios` sin nada publicado redirige a la portada, de forma
  silenciosa** (sin aviso: aclaración de la puerta 2): la decisión vive en el **frontend** (elemento
  de ruta de `/ministerios`); `SectionSkeleton` mientras `usePublicMinistries` está en vuelo y
  **solo con éxito y sin `items`** → `<Navigate to="/" replace />` (**`replace`**: «Atrás» sin bucle);
  **ante error NO se redirige** (estado de error con «Volver a intentar»); la API sigue respondiendo
  **`200` con `items` omitida**. **Distinto a propósito de F4 en `/eventos`** (que muestra
  `HiddenNotice` sin redirigir y **no se toca**): es la decisión literal del cliente, no una
  inconsistencia. Consecuencia de inventario: **`HiddenNotice` no se promueve** (P6-17).
- **P6-4/P6-5 — contacto excluyente y enlace seguro**: `leader_phone` **o** `leader_email` (nunca
  ambos; tags `phone`/`email` existentes; el contacto exige nombre); `whatsapp_url` con tag `url`
  (**cualquier `https` válido, sin restricción de dominios** — Q3) + sin espacios/control + `CHECK`
  en la BD; se **rechaza** `javascript:`/`data:`/`vbscript:`/cualquier esquema distinto de `https` y
  el valor se sirve **tal cual** (sin transformar el `href`).
- **P6-6 — sin permisos nuevos**: `AdminChain("ministerios")` con el permiso **ya sembrado** por
  `000002`; el catálogo de F2 **sigue en 10** y **ninguna migración de F2 se toca** (a diferencia de
  F4, que añadió `categorias`). La prosa del catálogo en el contrato **no cambia**.
- **P6-8/P6-9/P6-16 — superficie pública**: `GET /api/v1/ministerios?lang=` (sección completa,
  `maxPublicItems = 200`) y `GET /api/v1/portada/ministerios?lang=` (bloque, **tope 5 firme en el
  servidor**, sin parámetro de límite), ambos `no-store`; la **presencia de la navegación** se deduce
  del **documento del bloque** (`usePortadaMinisterios`, criterio técnico confirmado por el humano:
  mismo predicado "publicado", 0 peticiones extra en la portada); el bloque se inserta **entre
  `ProximamenteBlock` y `WhoWeAreSection`** (tras el bloque de eventos y antes de «quiénes somos»),
  con **tope 5** y **oculto por completo sin publicados** (FR-028: **sin mensaje**, a diferencia del
  bloque de F4).
- **P6-11/P6-12 — validación y duplicados**: límites (nombre/encargado 160, descripción **2.000**,
  horario/lugar 200, enlace 500, alt 300); español base **por campo** (`*En` exige `*Es`; un campo
  vacío no obliga a nada); imagen exige alt en español; **duplicado exacto → `409 conflict`**
  (normalizado `btrim` + minúsculas) y **enlace de WhatsApp compartido → aviso no bloqueante**
  (`whatsappUrlShared`, recalculado en cada lectura).
- **P6-13/P6-14 — auditoría y publicación**: 8 códigos `ministry.*` (incluidos
  `ministry.leader.show`/`ministry.leader.hide`), `InsertAdminAction` **transaccional** (fail-closed:
  si el registro falla no se aplica la mutación; el alta es **una sola fila** `ministry.create`
  —también con `publicationState: published` y `leaderVisible: true`, A6-10— y en la **edición**
  datos + estado/visibilidad → **dos filas**; reordenar se audita como `ministry.update`);
  `publication_state` por elemento y "se publica al guardar".
- **P6-18/P6-19/P6-20 — contrato y concurrencia**: fusión **aditiva** en `backend/api/openapi.yaml`
  (`info.version` 0.6.0 → **0.7.0**; el registro de `error.code` **no** cambia; `AdminActionItem.action`
  **+8** y prosa de `GET /api/v1/media/{fileName}`); orden público **y** de panel `sort_order, id`;
  última escritura completa gana — el `PATCH` **fusiona los campos presentes** sobre la fila actual
  y la transacción escribe **la fila completa ya fusionada**; los campos ausentes no cambian y
  `null` explícito borra lo opcional (A6-3).
- **P6-7/P6-17 — reutilización literal**: imágenes con `platform/storage` + `internal/media`
  (`IsFilePublished` crece de forma **aditiva** con `ministries`, RG6-5) y **solo `CardImage` se
  promueve** al inventario único (`frontend/src/components/publico/`); `HiddenNotice` **no**.

## Alineación de nomenclatura (decidida en la puerta 2 / `ux.md`; no se reabre)

1. **Rutas SPA**: la sección pública es **`/ministerios`** (una sola entrada de navegación
   «Ministerios», `ux.md` D-1) y el panel es **`/panel/ministerios`** (`ux.md` §3). El bloque de la
   portada vive en `HomePage` entre `ProximamenteBlock` y `WhoWeAreSection`.
2. **Carpetas de features** (nomenclatura de `ux.md` §6): público en
   **`frontend/src/features/publico/ministerios/`** y panel en
   **`frontend/src/features/ministerios/`**. Nombres de componentes y hooks los de `ux.md`
   §6.3/§6.4 (`MinisteriosPage`, `MinistryCard`, `EncargadoLine`, `WhatsappJoinButton`,
   `MinisteriosBlock`, `MinistryItem`, `usePublicMinistries`, `usePortadaMinisterios`; panel:
   `MinistriesPage`, `MinistriesList`, `MinistryForm`, `ContactoEncargado`, `ShowContactSwitch`,
   `useMinistries`…, `validation.ts`, `errors.ts`, `messages.ts`).
3. **Rutas de API**: las del contrato (`GET /api/v1/ministerios`,
   `GET /api/v1/portada/ministerios`, `GET/POST /api/v1/admin/ministerios`,
   `GET/PATCH/DELETE /api/v1/admin/ministerios/{id}`, `POST /api/v1/admin/ministerios/imagenes`);
   el snapshot `contracts/openapi.yaml` es inmutable (P6-18).
4. **Permiso**: código **`ministerios`** (sin tilde: es el `code` sembrado por `000002`; la etiqueta
   es «Ministerios»). `lib/permissions.ts` expone `MINISTERIOS`; el guard es el `RequirePermission`
   existente y el servidor usa `AdminChain("ministerios")`.
5. **Numeración T601…T638** (patrón por funcionalidad: F1 `T0xx`, F2 `T2xx`, F3 `T3xx`, F4 `T4xx`;
   F6 empieza en `T601`).
6. **Auditoría**: 8 códigos `ministry.*` —`ministry.create|update|delete|publish|unpublish|
   leader.show|leader.hide|image.upload`— (`data-model.md`, R6-12); `target_label`
   `Ministerios · <nombre>` / `Ministerios · Imagen · <fileName>`.
7. **Registro de rutas del SPA**: se reparte entre las tareas de página —`/ministerios` en T627 y
   `/panel/ministerios` en T630— para que **cada commit compile** (en F4 iba junto a los permisos;
   aquí el reparto es solo de implementación y no cambia rutas, permisos ni contrato).

## Leyenda

| Marca | Significado |
|---|---|
| `[backend]` `[frontend]` `[db]` `[infra]` | Capa y subagente responsable (`[backend]`/`[db]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` → `devops`) |
| `[P]` | **Paralelizable** con las otras tareas de su grupo (archivos distintos y sin dependencia entre ellas; ver "Grupos de paralelismo") |
| *(sin `[P]`)* | Secuencial: tiene dependencias previas que deben estar integradas |
| *(FR-…, SC-…, US…)* | Requisitos de `spec.md` que cubre la tarea |

> **Orden de capas**: migraciones (`[db]`) → contrato → `platform/` → repository → service →
> handler/routes → frontend (api/hooks → páginas → navegación/panel) → e2e → cierre. Cada capa se
> construye y prueba con lo que necesita la siguiente (el service con fakes, el handler con service
> falso), manteniendo la dirección de dependencias **`handler → service → repository`** de §II. La
> **auditoría** se escribe junto a cada mutación porque FR-017 exige que la operación y su registro
> sean atómicos. Las migraciones van **primero** (`000010` → `000011`, ninguna aplicada se edita) y
> **no hay migración de permisos** (P6-6).

## Tabla resumen

| ID | Tarea | Capa | Fase | Paralelo | Depende de |
|---|---|---|---|---|---|
| T601 | Migración `000010_create_ministries` (tabla + índices, `leader_visible DEFAULT FALSE`) | `[db]` | 1 | P1 | — |
| T602 | Migración `000011_extend_admin_actions_for_ministries` (+8 códigos `ministry.*`) | `[db]` | 1 | P1 | — |
| T603 | Consultas sqlc `ministries.sql` + rama de `media.sql` y generar `internal/db` | `[db]` | 1 | — | T601, T602 |
| T604 | Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.6.0 → 0.7.0) | `[backend]` | 2 | P1 | — |
| T605 | Regenerar `frontend/src/api/schema.d.ts` | `[frontend]` | 2 | — | T604 |
| T606 | `platform/audit`: +8 códigos `ministry.*` | `[backend]` | 3 | P1 | — |
| T607 | `ministerios`: `model.go` + `repository.go` (pgtype, `withTx`, auditoría sobre el tx) | `[backend]` | 4 | — | T603, T606 |
| T608 | `ministerios`: `repository_ministerios.go` (CRUD, duplicados, `whatsapp_url_shared`) | `[backend]` | 4 | P2 | T607 |
| T609 | `ministerios`: `repository_public.go` (privacidad `CASE WHEN leader_visible`, default `FALSE`) | `[backend]` | 4 | P2 | T607 |
| T610 | `media`: `IsFilePublished` ampliado a `ministries` (regresión de F3/F4) · RG6-5 | `[backend]` | 4 | P2 | T603 |
| T611 | Integración: `000010`/`000011` up→down→up, «oculto por defecto» y atomicidad | `[backend]` | 4 | — | T608, T609, T610 |
| T612 | `ministerios`: `service.go` — invariantes, validación y constantes | `[backend]` | 5 | — | T607 |
| T613 | `ministerios`: `service_public.go` — `GetMinisterios` (sección, privacidad, fallback) | `[backend]` | 5 | — | T612 |
| T614 | `ministerios`: `service_public.go` — `GetPortadaMinisterios` (tope 5, `items` omitida) | `[backend]` | 5 | — | T613 |
| T615 | `ministerios`: `service_admin.go` — CRUD, publicación, visibilidad del encargado, auditoría | `[backend]` | 5 | — | T612 |
| T616 | `ministerios`: `service_audit.go` — `audit.Recorder` (denegaciones y rechazos) | `[backend]` | 5 | — | T612 |
| T617 | `handler_public.go` — `GET /api/v1/ministerios` | `[backend]` | 6 | — | T613, T616 |
| T618 | `handler_public.go` — `GET /api/v1/portada/ministerios` | `[backend]` | 6 | — | T614, T617 |
| T619 | `handler_admin.go` — CRUD de `/api/v1/admin/ministerios` | `[backend]` | 6 | — | T615, T617 |
| T620 | `handler_images.go` — `POST /api/v1/admin/ministerios/imagenes` (multipart) | `[backend]` | 6 | — | T615, T619 |
| T621 | `routes.go` + `cmd/api/main.go` (`AdminChain("ministerios")`) + humo | `[backend]` | 6 | — | T610, T617–T620 |
| T622 | `api/ministerios.ts` + hooks (público, portada, panel, `uploadMinistryImage`) | `[frontend]` | 7 | — | T605 |
| T623 | Promoción de `CardImage` al inventario único (`components/publico/`) | `[frontend]` | 7 | P3 | — |
| T624 | i18n público (§7.2) + `messages.ts` del panel (§7.1) | `[frontend]` | 7 | P3 | — |
| T625 | Permisos y accesos del panel: `MINISTERIOS`, sello retirado, `AccesosRapidos`, `ACTION_LABELS` +8 | `[frontend]` | 7 | — | T622 |
| T626 | **Actualización acotada**: 4 pruebas de frontend que fijan permisos/accesos (RG6-6) | `[frontend]` | 7 | — | T625 |
| T627 | `/ministerios`: `MinisteriosPage`, `MinistryCard`, `EncargadoLine`, `WhatsappJoinButton` + **redirección** (P6-21) | `[frontend]` | 7 | P4 | T622, T623, T624 |
| T628 | `useSiteLinks`: entrada «Ministerios» condicional en barra, menú y pie | `[frontend]` | 7 | P4 | T622, T624 |
| T629 | `MinisteriosBlock` + `MinistryItem` en `HomePage` (tras eventos, antes de «quiénes somos») | `[frontend]` | 7 | P4 | T622, T623, T624 |
| T630 | Panel: `MinistriesPage` + `MinistriesList` + ruta `/panel/ministerios` (`RequirePermission`) | `[frontend]` | 7 | — | T622, T625 |
| T631 | Panel: `MinistryForm` + `ContactoEncargado` + `ShowContactSwitch` (apagado al crear) | `[frontend]` | 7 | — | T630 |
| T632 | `ministeriosSeed.ts` (fixtures del módulo, patrón `agendaSeed.ts`) | `[frontend]` | 8 | — | T630, T631 |
| T633 | E2E Playwright `ministerios-publica.spec.ts` (US1, US4, US5; ficha oculta y redirección) | `[frontend]` | 8 | P5 | T627, T628, T632 |
| T634 | E2E Playwright `ministerios-panel.spec.ts` (US2, US3; permisos y privacidad) | `[frontend]` | 8 | P5 | T630, T631, T632 |
| T635 | E2E Playwright `ministerios-accesibilidad.spec.ts` (teclado y foco, SC-008) | `[frontend]` | 8 | P5 | T627, T628, T629, T632 |
| T636 | E2E Playwright `portada-ministerios.spec.ts` (US6, SC-016) | `[frontend]` | 8 | P5 | T629, T632 |
| T637 | Verificación de cierre: `make ci`, cobertura ≥ 80 %, up→down→up, deriva, e2e por tandas | `[infra]` | 9 | — | todas |
| T638 | Documentación y PR (documentador + devops) | `[infra]` | 9 | — | T637 |

**Total: 38 tareas** — 17 `[backend]` · 16 `[frontend]` · 3 `[db]` · 2 `[infra]` · 16 marcadas `[P]`.

## Grupos de paralelismo

| Grupo | Tareas | Por qué pueden ir en paralelo |
|---|---|---|
| P1 | T601 · T602 · T604 · T606 | Cuatro artefactos independientes entre sí: dos migraciones (`000010`/`000011`, archivos distintos — el orden numérico solo importa al aplicar con `make db-migrate` —), el contrato (`backend/api/openapi.yaml`) y `platform/audit`; nadie comparte archivo ni depende de otro |
| P2 | T608 · T609 · T610 | `repository_ministerios.go`, `repository_public.go` y el dominio `media/` son archivos distintos que parten de T607/T603 |
| P3 | T623 · T624 | Promoción de `CardImage` (`src/components/publico/`) e i18n/`messages.ts` no se solapan |
| P4 | T627 · T628 · T629 | `features/publico/ministerios/` (página y ficha), `hooks/useSiteLinks.ts` y `MinisteriosBlock`/`MinistryItem` + `HomePage` son archivos distintos que parten de T622 **y de T624** (claves i18n tipadas; A6-5: T628 depende también de T624) |
| P5 | T633 · T634 · T635 · T636 | Cuatro specs e2e independientes (`ministerios-publica`, `ministerios-panel`, `ministerios-accesibilidad`, `portada-ministerios`) sobre la misma base (`ministeriosSeed`, T632) |

---

## Fase 1 — Migraciones y capa de datos (`[db]`)

> Convenciones de la skill `postgres-db` y de `data-model.md` (fuente de verdad): `id UUID`,
> `created_at`/`updated_at`, índices y `CHECK`/`UNIQUE` **en la base**, `up`/`down` completos, SQL
> **parametrizado** (§IV, CWE-89) y **nunca** se edita una migración aplicada (§VI). Numeración sin
> huecos desde `000001` (F4 terminó en `000009`). **Sin migración de permisos** (P6-6: `ministerios`
> ya está sembrado por `000002` y el catálogo sigue en 10).

- [X] T601 · Migración `000010_create_ministries` · `[db]` `[P]` (P1)

  - **Archivos**: `backend/migrations/000010_create_ministries.up.sql` y
    `000010_create_ministries.down.sql` (NUEVOS).
  - **Qué hace**: crea la tabla **`ministries`** tal como `data-model.md`: nombre/descripción
    obligatorios en español (`name_es`/`description_es`, `NOT NULL`) con versión en inglés opcional;
    **encargado opcional** (`leader_name_es`/`leader_name_en`, `leader_phone`, `leader_email` y
    **`leader_visible BOOLEAN NOT NULL DEFAULT FALSE`** —«Oculto por defecto», decisión de la puerta
    2: la ficha nace sin publicar y el `DEFAULT` cubre también los `INSERT` que omitan la columna—);
    `whatsapp_url` opcional (no traducible); `schedule_*`/`place_*` opcionales como **textos**
    (`TEXT`, sin columnas de fecha: Q6/FR-024 no aplican el modelo de F4); `image_file` +
    `image_alt_es`/`image_alt_en` opcionales; `publication_state` (`'draft'`/`'published'`, default
    `'draft'`) y `sort_order INTEGER NOT NULL DEFAULT 0`; `id`/`created_at`/`updated_at`. **CHECKs**:
    celular **o** correo, nunca ambos (Q8); el contacto exige nombre del encargado; español por campo
    (`*_en IS NULL OR *_es IS NOT NULL` en cada par); imagen exige `image_alt_es` (FR-020); enlace
    `^https://[^[:space:]]+$` (FR-015); longitudes (nombre/encargado 160, descripción 2.000,
    horario/lugar 200, enlace 500, alt 300, celular 7–32, correo 3–254, archivo 1–120) y `btrim` en
    los textos —el dato normalizado es el que se guarda: **A6-9**, `CHECK` de `btrim` también en
    `leader_phone`/`leader_email` con `IS NULL OR …`, como el precedente `000005`—. **Índices**: `ministries_sort_order_id_idx (sort_order, id)` y
    `ministries_publication_state_idx (publication_state)`. El `down` tira la tabla. *Cubre FR-004,
    FR-016, FR-019, FR-024, FR-025 y el modelo de FR-007/FR-012/FR-015/FR-020 a nivel de datos.*
  - **Pruebas incluidas** (§III): — (SQL; su verificación ejecutable es T611). **Cómo se verifica**:
    `make db-migrate` aplica sin errores sobre la BD de F4 (`000009` → `000010`); `migrate … down`
    revierte sin errores; los `CHECK`/`DEFAULT`/índices de `data-model.md` están literalmente en el
    SQL; ninguna tabla sin `id`/`created_at`/`updated_at`; `leader_visible DEFAULT FALSE` presente.
  - **Criterio de terminado**: el esquema coincide con `data-model.md` (lo revisa `revisor-codigo`
    contra la skill `postgres-db`) y el «oculto por defecto» está garantizado por la BD.
  - **Commit sugerido**: `feat(db): tabla ministries con ficha del encargado oculta por defecto (000010)`

- [X] T602 · Migración `000011_extend_admin_actions_for_ministries` · `[db]` `[P]` (P1)

  - **Archivos**: `backend/migrations/000011_extend_admin_actions_for_ministries.up.sql` y
    `.down.sql` (NUEVOS).
  - **Qué hace**: amplía el registro de auditoría de F2 (FR-017, R6-12): sustituye
    `admin_actions_action_check` (nombre propio desde `000006`, verificado en F3/F4) añadiendo los
    **8 códigos** `ministry.create`, `ministry.update`, `ministry.delete`, `ministry.publish`,
    `ministry.unpublish`, `ministry.leader.show`, `ministry.leader.hide` y `ministry.image.upload`
    (los dos de la visibilidad de la ficha del encargado, FR-017 los enumera como operación aparte).
    El `target_kind` **no cambia** (los ministerios son `content`, como los de F3/F4). El `down`
    elimina las filas con códigos F6 y restaura el CHECK de F4 (destructivo a propósito, igual que
    los `down` de `000006`/`000008`; documentado en `data-model.md`). *Cubre FR-017* a nivel de datos.
  - **Pruebas incluidas** (§III): — (SQL; su verificación ejecutable es T611). **Cómo se verifica**:
    `make db-migrate` aplica; los 8 códigos del `CHECK` coinciden uno a uno con la lista de
    `data-model.md` (incluidos `leader.show`/`leader.hide`); el registro de acciones queda
    **cerrado** en la BD.
  - **Criterio de terminado**: registro cerrado garantizado por la BD, sin tocar ninguna migración
    anterior (§VI).
  - **Commit sugerido**: `feat(db): ampliar admin_actions a ministerios (000011)`

- [X] T603 · Consultas sqlc `ministries.sql` + rama de `media.sql` y generar · `[db]`

  - **Archivos**: `backend/internal/db/queries/ministries.sql` (NUEVO),
    `backend/internal/db/queries/media.sql` (EDITADO: `IsFilePublished` crece con `ministries`) y
    `backend/internal/db/*.go` (GENERADOS y commiteados: `make sqlc`).
  - **Qué hace**: escribe las consultas nombradas de `data-model.md`:
    `ListPublishedMinistries(:limit)` —**única vía pública**: solo `publication_state = 'published'`,
    `ORDER BY sort_order, id`, `LIMIT :limit`, y los campos del encargado salen con **`CASE WHEN
    leader_visible THEN … END`** (con el interruptor apagado la fila **no transporta** el dato:
    privacidad en la consulta, R6-2/RG6-1)—; `ListMinistries(:limit, :offset, sqlc.narg(
    'publication_state'))` y `CountMinistries` (panel paginado, **todo**, mismo orden
    `sort_order, id`); `GetMinistryByID` (todas las columnas + `whatsapp_url_shared BOOLEAN` con
    `EXISTS` sobre otra fila con el mismo `whatsapp_url`, R6-11); `InsertMinistry`/`UpdateMinistry`/
    `DeleteMinistry` (el `Update` escribe **la fila completa ya fusionada** que le pasa el service:
    campos presentes del `PATCH` sobre la fila actual, ausentes intactos y `null` explícito borrando
    lo opcional — A6-3/P6-20) y `ExistsMinistryDuplicate` (nombre +
    descripción + encargado + enlace **normalizados** `btrim` + minúsculas, con
    `IS NOT DISTINCT FROM` en los opcionales). En `media.sql`, **`IsFilePublished`** crece con una
    tercera rama `OR` (`ministries.image_file` con `publication_state = 'published'`): el borrador
    no sirve su imagen (R6-7). Sin `SELECT *`, `ORDER BY` determinista (§8.1.7), todo parametrizado
    (§IV, CWE-89); la auditoría **reutiliza** `InsertAdminAction` de `queries/audit.sql` (sin
    escrituras sobre `admin_actions`, FR-025 de F2). *Cubre FR-002, FR-004, FR-012, FR-026 y el
    camino de archivos de FR-012/FR-020* a nivel de acceso a datos.
  - **Pruebas incluidas** (§III): — (SQL anotado; su ejecución es T608–T611). **Cómo se verifica**:
    `make sqlc` genera sin errores y `make sqlc-verify` no detecta deriva; los tipos generados
    quedan commiteados; ninguna consulta de escritura sobre `admin_actions`; la rama de `ministries`
    en `IsFilePublished` es **aditiva** (el resto de la consulta no cambia, RG6-5).
  - **Criterio de terminado**: la capa generada compila y es la única fuente de SQL del dominio
    `ministerios` y de la descarga de `media` (T607–T610).
  - **Commit sugerido**: `feat(db): consultas sqlc de ministerios y rama de media`

---

## Fase 2 — Contrato OpenAPI y tipos

- [X] T604 · Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.7.0) · `[backend]` `[P]` (P1)

  - **Archivos**: `backend/api/openapi.yaml` (EDITADO). *No* se toca
    `specs/006-ministerios/contracts/openapi.yaml` (snapshot inmutable, P6-18).
  - **Qué hace**: funde **aditivamente** el delta aprobado: las **8** operaciones de F6
    (`GET /api/v1/ministerios?lang=`, `GET /api/v1/portada/ministerios?lang=`, las 5 de panel
    `GET/POST /api/v1/admin/ministerios` y `GET/PATCH/DELETE …/{id}` y `POST
    /api/v1/admin/ministerios/imagenes`) y sus schemas (`MinisteriosPublicos`, `MinisterioPublico`,
    `LiderPublico`, `PortadaMinisterios`, `ItemMinisterioPortada`, `MinisterioAdminList`,
    `MinisterioAdmin`, `MinisterioInput`, `MinisterioPatch`, `WhatsappUrl`, `MinisterioItemId`),
    reutilizando `ErrorEnvelope`/`ErrorBody`, las respuestas reutilizables y `sessionCookie`.
    **Regla de fusión de parámetros (A6-2)**: los `$ref` `ListLimit`/`ListOffset` del delta se
    fusionan como **`Limit`/`Offset`**, los parámetros reutilizables reales del contrato vivo
    (`backend/api/openapi.yaml`); el resto del delta es literal. (El snapshot ya quedó normalizado a
    `Limit`/`Offset` en el `/speckit.analyze`; la instrucción se deja escrita para que la fusión no
    invente `ListLimit`/`ListOffset`, que **no existen** en el contrato vivo.) Aplica
    los cambios documentados del delta: enum `action` de `AdminActionItem` **+ 8 códigos**
    `ministry.*`; prosa de `GET /api/v1/media/{fileName}` (ahora también sirve imágenes de
    ministerios **publicados**); la prosa del **catálogo de permisos NO cambia** (sigue en 10;
    `ministerios` ya estaba sembrado — a diferencia de F4 —). Documenta el tope de la sección
    (`maxPublicItems = 200`), el **tope 5 del bloque resuelto en servidor** (sin parámetro de
    límite), `leaderVisible` con `default: false` («oculto por defecto»), `whatsappUrlShared` como
    **aviso** no bloqueante y que el enlace se sirve **tal cual**. Sube `info.version` de **0.6.0 a
    0.7.0**. El registro de `error.code` **no cambia** (F6 emite `invalid`, `unauthenticated`,
    `forbidden`, `not_found`, `conflict`, `internal`). *Cubre el contrato de todos los FR*; garantiza
    FR-011 (el permiso `ministerios` citado en las descripciones de panel), FR-012 (ninguna
    operación pública devuelve borradores), FR-015 (la validación del enlace documentada) y FR-017
    (los 8 códigos).
  - **Pruebas incluidas** (§III): — (artefacto de contrato). **Cómo se verifica**: el YAML parsea
    sin errores y **sin `$ref` sin resolver** (validador de `make api-gen` o búsqueda de `$ref`
    huérfanos — el simple «el YAML parsea» no lo detecta —);
    `/api/v1/ministerios` y `/api/v1/portada/ministerios` tienen `security: []`,
    `Cache-Control: no-store` documentado y **sin** campos de encargado oculto en los schemas
    públicos (`ItemMinisterioPortada` **sin** campos de encargado en ningún caso); las operaciones de
    panel citan `sessionCookie` y el permiso `ministerios`; `MinisterioInput.leaderVisible` declara
    `default: false`; `grep -n "version:" backend/api/openapi.yaml` → `0.7.0`; `make api-gen`
    completa sin errores.
  - **Criterio de terminado**: el contrato vivo contiene F1–F4 + F6 y es la única fuente de tipos del
    frontend (T605).
  - **Commit sugerido**: `docs(api): fusionar delta de F6 en el contrato vivo (0.7.0)`

- [ ] T605 · Regenerar `frontend/src/api/schema.d.ts` · `[frontend]`

  - **Archivos**: `frontend/src/api/schema.d.ts` (GENERADO y commiteado).
  - **Qué hace**: ejecuta `make api-gen` (o `npm run api:gen`) contra **`backend/api/openapi.yaml`**
    (nunca contra el snapshot de `specs/`, P6-18) y commitea los tipos de F6 (`MinisteriosPublicos`,
    `MinisterioPublico`/`LiderPublico`, `PortadaMinisterios`/`ItemMinisterioPortada`,
    `MinisterioAdmin`/`MinisterioInput`/`MinisterioPatch`, `ImageUploadResult`, `PublicationState`,
    `AdminActionItem` con los enums ampliados a los 8 códigos `ministry.*`).
  - **Pruebas incluidas** (§III): — (artefacto generado). **Cómo se verifica**: regenerar no produce
    `git diff`; `npm run typecheck` pasa con los tipos nuevos.
  - **Criterio de terminado**: `schema.d.ts` reproducible desde el contrato y listo para
    `api/ministerios.ts` (T622; regla anti-deriva de §8.1.5).
  - **Commit sugerido**: `build(frontend): regenerar schema.d.ts con los tipos de F6`

---

## Fase 3 — Núcleo de plataforma (`backend/internal/platform/`)

- [X] T606 · `platform/audit`: 8 códigos `ministry.*` · `[backend]` `[P]` (P1)

  - **Archivos**: `backend/internal/platform/audit/audit.go` y `audit_test.go` (EDITADOS).
  - **Qué hace**: amplía el registro cerrado del plumbing de F2 (R6-12) con los **8 códigos**
    `ministry.create|update|delete|publish|unpublish|leader.show|leader.hide|image.upload` en
    `ActionCodes` y `ValidActionCode`. `TargetKindContent = "content"` **no cambia** (ya lo añadió
    F3). *Cubre FR-017*.
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/platform/audit/` en verde: cada código `ministry.*` pasa `Validate` con
    objetivo `content` y `TargetLabel`; los códigos de F2/F3/F4 siguen intactos (regresión); un
    código fuera de la lista falla.
  - **Criterio de terminado**: la lista cerrada coincide literalmente con el `CHECK` de la migración
    `000011` (T602) — lo comprueba la integración de T611.
  - **Commit sugerido**: `feat(audit): codigos ministry.* para F6`

---

## Fase 4 — Dominio `ministerios`: repository

- [X] T607 · `ministerios`: `model.go` + `repository.go` · `[backend]`

  - **Archivos**: `backend/internal/ministerios/model.go`, `repository.go`, `repository_test.go`
    (NUEVOS).
  - **Qué hace**: entidades del dominio (`Ministry` con el encargado como **campos** —`LeaderNameEs`,
    `LeaderNameEn`, `LeaderPhone`, `LeaderEmail`, `LeaderVisible`—, `SortOrder`, `PublicationState`),
    DTOs con etiquetas `validate` y sufijos `Es`/`En` (R3-1) y el **DTO público** (`MinisterioPublico`
    con `leader` **opcional**: el mapeo público se compone en `service_public.go`, un solo lugar,
    P6-2/RG6-1), y la base del repository: constructor sobre `*pgxpool.Pool`, mapeos `pgtype` →
    dominio (§8.1.6), `withTx` (patrón de F2) y la **inserción de `admin_actions` dentro del `tx`**
    vía `InsertAdminAction` (mapeo `audit.Action` → `InsertAdminActionParams`; sin que un dominio
    importe a otro, R2). *Soporta FR-007 (patrón bilingüe) y FR-017 (registro atómico)*.
  - **Pruebas incluidas** (§III): unitarias de los mapeos (`pgtype` ↔ dominio, nil/`Valid`) y del
    mapeo de acciones (incluidos `ministry.leader.show`/`hide`). **Cómo se verifica**:
    `go test ./internal/ministerios/` en verde; `gofmt`/`go vet` limpios; ningún `pgtype` sale de
    `repository.go` (R4).
  - **Criterio de terminado**: el paquete compila con sus tipos y el tx de escritura queda listo para
    T608/T609.
  - **Commit sugerido**: `feat(ministerios): modelo y base del repository con auditoria transaccional`

- [X] T608 · `ministerios`: `repository_ministerios.go` (CRUD, duplicados y aviso de enlace) · `[backend]` `[P]` (P2)

  - **Archivos**: `backend/internal/ministerios/repository_ministerios.go` y
    `repository_ministerios_integration_test.go` (NUEVOS; `//go:build integration`).
  - **Qué hace**: `InsertMinistry`/`UpdateMinistry`/`DeleteMinistry`/`GetMinistryByID`,
    `ListMinistries`/`CountMinistries` (panel paginado, `ORDER BY sort_order, id` — el mismo orden
    que ve el visitante, P6-19 —, filtro opcional `publication_state`), `ExistsMinistryDuplicate`
    (normalizado) y `whatsapp_url_shared` (recalculado en cada lectura, nunca persistido, RG6-11).
    Traducción de restricciones a `apperr`: `UNIQUE`/duplicado → `conflict`, violaciones de `CHECK`
    (contacto excluyente, contacto sin nombre, `*En` sin `*Es`, imagen sin alt, enlace que no es
    `https`) → `invalid` con `details`. Borrado físico (sin `deleted_at`, default revisable).
    *Cubre FR-010, FR-014, FR-015, FR-016, FR-019*.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/ministerios/` en verde: una fila con `leader_phone` **y**
    `leader_email` falla (y con solo uno procede); `leader_phone` sin `leader_name_es` falla;
    `schedule_en` sin `schedule_es` falla (ídem el resto de pares); `image_file` sin `image_alt_es`
    falla; `http://…`, `javascript:alert(1)` y un valor con espacio fallan en la BD mientras
    `https://chat.whatsapp.com/…` **y** `https://ejemplo.com/grupo` (dominio no de WhatsApp, Q3)
    pasan; `ExistsMinistryDuplicate` detecta "Alabanza" / " alabanza " y **no** dispara por un
    enlace compartido; `whatsapp_url_shared` es `TRUE` solo con otra fila igual; el `Update` deja
    exactamente la fila completa que recibe del service (ya **fusionada**: los campos ausentes del
    `PATCH` conservan su valor y un `null` explícito borra lo opcional — última escritura completa
    gana, P6-20/A6-3); **ciclo de vida del archivo (A6-11)**: `DeleteMinistry` **no borra** el
    archivo de la imagen (sigue en disco y deja de servirse por `IsFilePublished`, T610) y
    reemplazar o quitar la imagen limpia el archivo anterior **best-effort** (como en `portada`; un
    fallo de limpieza no tira la operación).
  - **Criterio de terminado**: persistencia real de ministerios con sus invariantes protegidos por la
    BD y traducidos a errores del dominio.
  - **Commit sugerido**: `feat(ministerios): persistencia de ministerios con duplicados y aviso de enlace`

- [X] T609 · `ministerios`: `repository_public.go` — privacidad del encargado en la consulta · `[backend]` `[P]` (P2)

  - **Archivos**: `backend/internal/ministerios/repository_public.go` y
    `repository_public_integration_test.go` (NUEVOS; `//go:build integration`).
  - **Qué hace**: `ListPublishedMinistries` (única vía pública): solo `publication_state =
    'published'`, `ORDER BY sort_order, id`, `LIMIT :limit`, con los campos del encargado
    **enmascarados con `CASE WHEN leader_visible`** (R6-2: con el interruptor apagado la fila sale
    con `NULL`, ni nombre ni contacto). *Cubre FR-002, FR-004, FR-012 y FR-026* (capa de consulta,
    **RG6-1**).
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real** (las más importantes del
    módulo). **Cómo se verifica**:
    `go test -tags=integration ./internal/ministerios/` en verde: la consulta **nunca** devuelve
    borradores; con `leader_visible = FALSE` y nombre/celular guardados, las columnas del encargado
    salen **`NULL`** (prueba «0 apariciones» de la capa de datos); con `leader_visible = TRUE` salen
    los datos (y solo el nombre si no hay contacto); **un `INSERT` que omite `leader_visible` guarda
    `FALSE`** (prueba del `DEFAULT` de `000010`, «oculto por defecto», P6-3/RG6-10); el orden es
    `sort_order, id` y el `LIMIT` se respeta (5 en el bloque, 200 en la sección).
  - **Criterio de terminado**: la privacidad del encargado está garantizada **en la base**, no solo
    en la interfaz (FR-026: "no basta con ocultarlo en la interfaz").
  - **Commit sugerido**: `feat(ministerios): lectura publica sin datos del encargado oculto`

- [X] T610 · `media`: `IsFilePublished` con `ministries` (regresión de F3/F4) · `[backend]` `[P]` (P2) *(RG6-5)*

  - **Archivos**: `backend/internal/media/repository.go`, `repository_media_integration_test.go`
    (EDITADOS) — la consulta ya creció en T603. *El resto de `internal/media/` y de F3/F4 no se
    toca.*
  - **Qué hace**: la descarga `GET /api/v1/media/{fileName}` acepta también archivos referenciados
    por **`ministries.image_file` con `publication_state = 'published'`** (tercera rama `OR`, rama
    **aditiva**): un borrador **no** sirve su imagen, la URL deja de responder `200` al retirar y
    **vuelve** al publicar (el archivo no se borra) — misma regla que F3/F4 (R6-7). *Cubre FR-012
    (camino de archivos) y FR-020.*
  - **Pruebas incluidas** (§III): las pruebas de `internal/media` existentes **se conservan** +
    casos nuevos. **Cómo se verifica**:
    `go test ./internal/media/` y `go test -tags=integration ./internal/media/` en verde —
    **regresión de F3/F4 (RG6-5)**: logo/imagen de portada publicada se sigue sirviendo igual, un
    borrador de portada sigue dando `404` y las imágenes de agenda publicada siguen `200` (mismos
    casos que antes); **F6**: imagen de ministerio publicado → `200`; el mismo ministerio en
    borrador → `404` sin borrar el archivo; al publicarlo vuelve a `200`. Los e2e de F3/F4
    (`portada-imagenes`, `agenda-*`) siguen en verde como evidencia.
  - **Criterio de terminado**: la regla "solo contenido publicado" cubre ministerios sin regresión de
    las imágenes de F3/F4.
  - **Commit sugerido**: `refactor(media): IsFilePublished cubre imagenes de ministerios`

- [X] T611 · Integración: migraciones up→down→up, «oculto por defecto» y atomicidad · `[backend]`

  - **Archivos**: `backend/internal/ministerios/repository_migrations_integration_test.go` y
    `repository_audit_integration_test.go` (NUEVOS; `//go:build integration`).
  - **Qué hace**: la verificación ejecutable de `data-model.md` §"Validación del modelo":
    (1) las migraciones `000010` y `000011` aplican **up → down → up** sin errores (restauración del
    `CHECK` de `admin_actions`); (2) el catálogo de `permissions` **sigue en 10** filas con
    `ministerios` entre ellas (F6 no lo toca; el `down` no lo roza); (3) los **CHECKs** de contacto
    excluyente, contacto exige nombre, español por campo, imagen+alt y enlace `https` se disparan en
    la BD; (4) **«Oculto por defecto»**: un `INSERT` que omite `leader_visible` guarda `FALSE`
    (P6-3/RG6-10) y una fila con datos de encargado en su default sale **sin** datos del encargado en
    la lectura pública; (5) **mutación + `admin_actions` atómicas**: al forzar un fallo en el
    registro, la mutación **no se aplica**, y los 8 códigos del `CHECK` de la BD coinciden con
    `platform/audit` (T606) y caben en el registro (y un código fuera del registro se rechaza).
    *Cubre FR-014, FR-015, FR-017, FR-025 y FR-026.*
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/ministerios/` en verde con los cinco bloques.
  - **Criterio de terminado**: las migraciones son reversibles, el default del interruptor y la
    atomicidad de la auditoría quedan verificados en la base.
  - **Commit sugerido**: `test(ministerios): integracion de migraciones, oculto por defecto y auditoria atomica`

---

## Fase 5 — Dominio `ministerios`: servicios

- [X] T612 · `ministerios`: `service.go` — invariantes, validación y constantes · `[backend]`

  - **Archivos**: `backend/internal/ministerios/service.go` y `service_test.go` (NUEVOS).
  - **Qué hace**: interfaz `Repository` (definida por quien la consume, R3), constantes revisables
    (nombre/encargado **160**, descripción **2.000**, horario/lugar **200**, enlace **500**, alt
    **300**, celular 7–32, correo 3–254, **`maxPublicItems = 200`** y **`maxPortadaItems = 5`** —
    tope fijo de la spec, resuelto en servidor: los aplican T613/T614 y los cita el contrato —),
    normalización (trim; `""`/espacios en `*_En` y opcionales → `NULL`, I6) y la **validación de
    dominio** (FR-014/FR-015): obligatorios en español (nombre y descripción), **español por campo**
    (`*En` exige su `*Es`; un campo vacío no obliga a nada, Q6), **contacto excluyente** (celular
    **o** correo, nunca ambos — Q8 —; tags `phone`/`email` existentes; el contacto exige nombre del
    encargado), **enlace `https` seguro** (tag `url` + sin espacios/caracteres de control; se
    rechaza `javascript:`/`data:`/`vbscript:`/`http://`/sin host/501 chars y se acepta **cualquier
    dominio**, incluido uno que no sea de WhatsApp — Q3 —; el valor **no se transforma**),
    imagen exige `imageAltEs` (FR-020) y límites por campo. `lang` ∈ `es|en`. *Soporta FR-014,
    FR-015, FR-019, FR-020, FR-024 y FR-025*.
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/ministerios/` en verde: cada campo inválido → `apperr.Invalid` con
    `details` por campo; `{"leaderPhone": …, "leaderEmail": …}` → `invalid` (Q8);
    `whatsappUrl: "javascript:alert(1)"` → `invalid` con `details.whatsappUrl` (ídem `data:`,
    `vbscript:`, `http://…`, sin host, con espacio, 501 chars) y un dominio no de WhatsApp **pasa**;
    `{"descriptionEn": "x"}` sin `descriptionEs` → `invalid`; `""`/`"  "` → `NULL`; tildes/`ñ`/`¿¡`/
    emojis intactos; `maxPublicItems` es **200** y `maxPortadaItems` es **5** (fijados por prueba).
  - **Criterio de terminado**: las reglas de validación tienen **una** implementación y sus pruebas;
    los límites son el único punto de verdad (los usan T613–T615 y los cita el contrato).
  - **Commit sugerido**: `feat(ministerios): invariantes y validacion de dominio`

- [X] T613 · `ministerios`: `service_public.go` — `GetMinisterios` (sección) · `[backend]`

  - **Archivos**: `backend/internal/ministerios/service_public.go` y `service_public_test.go`
    (NUEVOS).
  - **Qué hace**: `GetMinisterios(lang)` (FR-002, FR-005): devuelve `{lang, items?}` con **solo**
    contenido `published` (`ListPublishedMinistries`, tope `maxPublicItems = 200`) en orden
    `sort_order, id` (FR-004), cada campo traducible resuelto con fallback **`en` si existe, si no
    `es`** (FR-008; el cliente **no** reimplementa el fallback; el enlace **no** se traduce) y
    **`items` omitida sin publicados** (el frontend decide: redirección o bloque oculto). **La
    privacidad del encargado se compone aquí, en un solo lugar** (P6-2/RG6-1): `leader` (nombre y,
    si existe, el **único** medio de contacto) entra en el DTO **solo si `leaderVisible` y hay
    nombre**; con el interruptor apagado el DTO **no tiene ni campo ni cadena** del encargado; con
    `leaderVisible` y sin contacto → solo nombre. `lang` fuera de `es|en` → `apperr.Invalid`. El DTO
    público **nunca** contiene `publicationState`. Horario y lugar viajan como **texto** (FR-024: sin
    fechas derivadas). *Cubre FR-002, FR-007, FR-008, FR-012, FR-024 y FR-026* (capa de servicio).
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos; **las más importantes del
    módulo**). **Cómo se verifica**: `go test ./internal/ministerios/` en verde: con
    `leaderVisible: false` y nombre/celular guardados, el DTO serializado **no contiene** el nombre
    ni el contacto **ni como campo ni como cadena**, en `es` y en `en` («0 apariciones»); **un alta
    sin `leaderVisible` devuelve `leaderVisible: false`** y su DTO público no trae `leader` aunque
    haya datos guardados («oculto por defecto», P6-3); con `leaderVisible: true` y sin contacto →
    solo nombre, sin línea de contacto vacía; fallback por campo (nunca hueco); orden
    `sortOrder, id`; con más de 200 publicados devuelve **200** (los primeros ya ordenados);
    `items` omitida con 0 publicados; `lang=fr` → `invalid`; el DTO nunca contiene
    `publicationState`.
  - **Criterio de terminado**: SC-014 (0 apariciones) y SC-002 (0 borradores) verificables en el
    service; cobertura medida (`go test -cover`, ≥ 80 % al cierre).
  - **Commit sugerido**: `feat(ministerios): seccion publica con privacidad del encargado`

- [X] T614 · `ministerios`: `service_public.go` — `GetPortadaMinisterios` (tope 5) · `[backend]`

  - **Archivos**: `backend/internal/ministerios/service_public.go` (EDITADO) y
    `service_public_test.go` (AMPLIADO).
  - **Qué hace**: `GetPortadaMinisterios(lang)` (FR-027/FR-028): mismo predicado "publicado" y mismo
    fallback `en → es` que la sección, con **tope `maxPortadaItems = 5` resuelto en el servidor**
    (`LIMIT 5`; el contrato **no** expone parámetro de límite) y orden `sort_order, id` (los 5
    **primeros** según el orden manual del módulo). Devuelve `PortadaMinisterios` con
    `ItemMinisterioPortada` (**`id`, `name`, `schedule`/`place`, `image`+`alt`** — la tarjeta compacta
    de `ux.md` §4.9) y **sin campos de encargado en ningún caso** (P6-2: ni con el interruptor
    encendido — el extracto no publica encargado, D-12) y **sin botón de unión** (D-12). Sin
    publicados → **`items` omitida** (FR-028: el bloque no se dibuja — **sin mensaje**, diferencia
    consciente con el bloque de F4). *Cubre FR-027, FR-028 y los aspectos de extracto de FR-002/
    FR-012/FR-026*.
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos). **Cómo se verifica**:
    `go test ./internal/ministerios/` en verde: con 8 publicados salen **exactamente los 5 primeros**
    en `sortOrder, id` (los otros 3 no aparecen ni con paginación); con menos, exactos (sin huecos);
    con 0 publicados → `items` **omitida**; 0 borradores; el DTO del bloque **jamás** contiene el
    nombre ni el contacto del encargado aunque estén publicados (prueba «0 apariciones» del extracto);
    `image`/`alt` solo cuando el elemento tiene imagen; fallback `en → es` por campo; `lang`
    inválido → `invalid`; el tope **5** queda fijado por prueba (constante: no crece con el
    contenido).
  - **Criterio de terminado**: US6/SC-016 resueltos a nivel de dominio y cobertura de
    `service*.go` ≥ 80 % intacta.
  - **Commit sugerido**: `feat(ministerios): extracto de la portada con tope 5 y sin encargado`

- [X] T615 · `ministerios`: `service_admin.go` — CRUD, publicación y visibilidad del encargado · `[backend]`

  - **Archivos**: `backend/internal/ministerios/service_admin.go` y `service_admin_test.go`
    (NUEVOS).
  - **Qué hace**: `CreateMinisterio`/`UpdateMinisterio`/`DeleteMinisterio`/`ListMinisterios`/
    `GetMinisterio` (FR-010): validación completa (T612) con `details` por campo y **sin guardar
    nada a medias**; **duplicado exacto → `409 conflict`** (P6-12) y **enlace compartido permitido**
    con `whatsappUrlShared` recalculado en cada lectura (aviso, nunca bloqueo, RG6-11);
    **`leaderVisible` con `default: false`** — el service aplica `false` si falta en el input
    («oculto por defecto», P6-3) — y gobierna **solo la publicación** (dato guardado ≠ dato
    publicado: apagar y encender el interruptor **conserva** los datos y se puede repetir sin
    reescribir, US2 esc. 10–11); `publication_state` por elemento y **"se publica al guardar"**
    (FR-013: la edición de un ministerio publicado es visible en la primera carga); panel paginado
    (`limit`/`offset`, `platform/paginate`, filtro `publicationState`) en orden `sort_order, id`
    (P6-19) con `isPast` **ausente** (nada vence: sin fechas) y marca de ficha oculta derivada de
    `leaderVisible`; **registro transaccional** con la regla exacta de códigos (R6-12, precisión
    A6-10): el alta registra **solo `ministry.create`** —una sola fila, también con
    `publicationState: published` y `leaderVisible: true`—; `ministry.publish`/`unpublish` **solo**
    para cambios de estado y `ministry.leader.show`/`hide` **solo** para cambios del interruptor,
    ambos **sobre un ministerio existente** (operaciones de `PATCH`); `ministry.update` para
    cualquier otro cambio (incluido `sortOrder`); en la edición, datos + estado/visibilidad a la vez
    → **dos filas** (fail-closed: si el registro falla no se aplica nada). El `UpdateMinisterio`
    **fusiona los campos presentes** sobre la fila actual y escribe la fila completa ya fusionada
    (A6-3/P6-20). *Cubre FR-010, FR-013, FR-014,
    FR-016, FR-017, FR-025 y US2/US3*.
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos). **Cómo se verificar**:
    `go test ./internal/ministerios/` en verde: cada campo inválido → `400` con `details` y **sin
    guardar nada**; ambos contactos → `details` de contacto; enlace `javascript:` → `invalid`;
    límite de 500 caracteres exacto y 501 → `invalid`; duplicado → `conflict`; enlace compartido → guardado + `whatsappUrlShared: true`;
    **crear con encargado sin `leaderVisible` → `leaderVisible: false` y el DTO público no trae
    `leader`** («oculto por defecto»); apagar/encender el interruptor deja `ministry.leader.hide`/
    `show` y **conserva** nombre y contacto; una edición que cambia datos y estado deja **dos
    filas**; si el registro falla no se guarda; `""`/espacios en `*En` → `NULL`; tildes/`ñ`/`¿¡`/
    emojis intactos; eliminar deja su fila `ministry.delete`. **Casos A6-3 (merge del `PATCH`)**:
    actualizar **un solo campo** → el resto queda **intacto** (la fusión es sobre la fila actual) y
    `null` explícito → **opcional borrado** (p. ej. `whatsappUrl: null`, `scheduleEs: null`).
    **Caso A6-10 (auditoría del alta)**: el alta deja **solo `ministry.create`** —también con
    `publicationState: published` y `leaderVisible: true`: **cero** filas `ministry.publish`/
    `unpublish`/`leader.show`/`leader.hide`—. **Caso A6-11 (archivos)**: `DeleteMinisterio` **no
    borra** el archivo de la imagen (queda en disco y deja de servirse) y reemplazar o quitar la
    imagen limpia el archivo anterior **best-effort** (como en `portada`; un fallo de limpieza no
    cambia la respuesta).
  - **Criterio de terminado**: US2/US3 y FR-010/FR-014/FR-017/FR-025 cubiertos en el service.
  - **Commit sugerido**: `feat(ministerios): gestion del panel con visibilidad del encargado y auditoria`

- [X] T616 · `ministerios`: `service_audit.go` — `audit.Recorder` · `[backend]`

  - **Archivos**: `backend/internal/ministerios/service_audit.go` y `service_audit_test.go`
    (NUEVOS).
  - **Qué hace**: implementa `audit.Recorder` para el dominio (patrón de F4, R6-12): `RecordDenied`
    resuelve **método+ruta → acción** de las rutas del módulo (`/api/v1/admin/ministerios…` →
    `ministry.*`; las subidas → `ministry.image.upload`) y la persiste con `result='denied'`
    best-effort; el helper de rechazo hace lo propio con `result='failure'`. Nunca transporta el
    cuerpo de la petición (FR-026 de F2). `target_label`: `Ministerios · <nombre>`,
    `Ministerios · Imagen · <fileName>` (máx. 254). *Cubre FR-017 (y el patrón de denegación de
    FR-011)*.
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/ministerios/` en verde: cada ruta del módulo resuelve a su acción y
    `targetLabel`; una ruta desconocida no deja fila; un fallo de registro no cambia la respuesta
    (best-effort).
  - **Criterio de terminado**: las denegaciones sobre las rutas de F6 quedan registradas igual que las
    de F2/F3/F4 (SC-011).
  - **Commit sugerido**: `feat(ministerios): registro de denegaciones y rechazos del modulo`

---

## Fase 6 — Handlers, rutas y cableado

- [X] T617 · `handler_public.go` — `GET /api/v1/ministerios` · `[backend]`

  - **Archivos**: `backend/internal/ministerios/handler.go`, `handler_public.go`,
    `handler_public_test.go` (NUEVOS).
  - **Qué hace**: publica la operación del contrato: valida `lang` (`es|en`, por defecto `es`),
    llama a `GetMinisterios` (T613) y responde `200` con `MinisteriosPublicos` y **`Cache-Control:
    no-store`** (SC-003/FR-013). Sin autenticación (FR-002). Helpers HTTP del dominio
    (decodificar/validar/responder con `apperr` + `WriteError`, patrón de F2/F4). El cuerpo JSON
    **nunca** contiene `publicationState` ni datos del encargado oculto. *Cubre FR-002, FR-008,
    FR-012 y FR-026* (vía de salida pública).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/ministerios/` en verde: `200` con el sobre correcto y `no-store`; `lang=fr`
    → `400 invalid` con `details.lang`; `items` omitida cuando el service no la devuelve; **el
    cuerpo JSON serializado no contiene el nombre ni el contacto del encargado cuando el service no
    los entrega** (aserción sobre el JSON crudo: «0 apariciones», **RG6-1**); error del service →
    sobre `ErrorEnvelope` genérico.
  - **Criterio de terminado**: `quickstart.md` §1 es ejecutable tal cual por API (US1) y la privacidad
    está verificada también en la capa HTTP.
  - **Commit sugerido**: `feat(ministerios): endpoint publico de la seccion`

- [X] T618 · `handler_public.go` — `GET /api/v1/portada/ministerios` · `[backend]`

  - **Archivos**: `backend/internal/ministerios/handler_public.go` (EDITADO),
    `handler_public_test.go` (AMPLIADO).
  - **Qué hace**: publica la operación del contrato (P6-16): valida `lang` (`es|en`, por defecto
    `es`), llama a `GetPortadaMinisterios` (T614) y responde `200` con `PortadaMinisterios` y
    **`Cache-Control: no-store`**; sin parámetro de límite (el **tope 5 es del servidor**); `items`
    **omitida** con 0 publicados (nunca `null`: el bloque no se dibuja, FR-028). El DTO **nunca**
    contiene campos de encargado, ni `publicationState`. La ruta va en `RegisterPublic` **sin
    middlewares** (pública, solo lectura) — anotando que el segmento `portada` es el agregado de F3
    pero el recurso lo sirve `ministerios` (Complexity 4 del plan). *Cubre FR-027, FR-028*.
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/ministerios/` en verde: `200` con el sobre correcto y `no-store`; `items`
    omitida con 0 publicados; `lang=fr` → `400 invalid` con `details.lang`; el cuerpo JSON jamás
    contiene el nombre/correo del encargado («0 apariciones» del extracto); error del service →
    sobre genérico.
  - **Criterio de terminado**: `quickstart.md` §14 es ejecutable tal cual por API (US6).
  - **Commit sugerido**: `feat(ministerios): endpoint publico del bloque de la portada`

- [X] T619 · `handler_admin.go` — CRUD de `/api/v1/admin/ministerios` · `[backend]`

  - **Archivos**: `backend/internal/ministerios/handler_admin.go` y `handler_admin_test.go`
    (NUEVOS).
  - **Qué hace**: `GET/POST /api/v1/admin/ministerios` y `GET/PATCH/DELETE
    /api/v1/admin/ministerios/{id}` sobre el service de T615: listado paginado (`limit`/`offset`
    con `platform/paginate`, filtro opcional `publicationState`), `201`/`200`/`204` según el
    contrato, `400` con `details` por campo, `404` por id inexistente y `409` en duplicado exacto;
    el `PATCH` **fusiona los campos presentes** sobre la fila actual y la transacción escribe **la
    fila completa ya fusionada** (última escritura completa gana, P6-20/A6-3): los campos ausentes
    no cambian y `null` explícito borra lo opcional. Las respuestas del panel
    (`MinisterioAdmin`) **sí** incluyen los datos guardados del encargado con `leaderVisible` apagado
    (dato guardado ≠ dato publicado) y `whatsappUrlShared` como booleano (aviso). *Cubre FR-010,
    FR-014, FR-016, FR-025* (vía de panel).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/ministerios/` en verde: cada operación con sus códigos
    200/201/204/400/404/409 y sobres uniformes; `details` por campo en los `400`; la respuesta de
    panel incluye `leaderVisible` y `whatsappUrlShared` siempre presentes; **casos A6-3 del
    `PATCH`**: petición con **un solo campo** → el resto **intacto** (el handler transmite el DTO
    parcial tal cual al service, que fusiona) y `null` explícito → **opcional borrado**; el `403` de
    permiso se
    prueba en T621 (cada grupo con su `AdminChain`).
  - **Criterio de terminado**: `quickstart.md` §3/§4/§6 ejecutables vía API (US2).
  - **Commit sugerido**: `feat(ministerios): endpoints del panel de ministerios`

- [X] T620 · `handler_images.go` — `POST /api/v1/admin/ministerios/imagenes` · `[backend]`

  - **Archivos**: `backend/internal/ministerios/handler_images.go` y `handler_images_test.go`
    (NUEVOS).
  - **Qué hace**: subida multipart del contrato (patrón literal de F3/F4, R6-7): tope con
    `http.MaxBytesReader` (≤ 8 MB), tipo por **firma binaria** (solo JPEG/PNG/WebP — sin SVG ni
    GIF), nombre generado `img_<uuid>.<ext>` con `platform/storage` (reutilizado), **registrada con
    `ministry.image.upload`** (fail-closed: si el registro falla se aborta y se borra el archivo) y
    respuesta `201` con `ImageUploadResult`. Tras `AdminChain("ministerios")` (T621). La subida **no**
    cambia contenido visible. *Cubre FR-020 y el edge case "archivo inválido o carga fallida"*.
  - **Pruebas incluidas** (§III): `httptest` con `Store` falso y `t.TempDir()`. **Cómo se verifica**:
    `go test ./internal/ministerios/` en verde: sin archivo o campo erróneo → `400` con
    `details.file`; archivo > 8 MB → `400` con `details.fileSize`; `.svg`/HTML renombrado → `400`
    (la firma manda); cada subida deja su fila `ministry.image.upload` y un fallo de registro aborta
    sin dejar archivo.
  - **Criterio de terminado**: `quickstart.md` §10 ejecutable y `seguridad` aprueba las reglas de
    subida (idénticas a las de F3/F4).
  - **Commit sugerido**: `feat(ministerios): subida de imagenes con auditoria`

- [X] T621 · `routes.go` + `cmd/api/main.go` (`AdminChain("ministerios")`) + humo · `[backend]`

  - **Archivos**: `backend/internal/ministerios/routes.go` y `routes_test.go` (NUEVOS),
    `backend/cmd/api/main.go` y `main_test.go` (EDITADOS: DI del dominio `ministerios` + registro de
    `/api/v1/ministerios`, `/api/v1/portada/ministerios` y `/api/v1/admin/ministerios`).
  - **Qué hace**: cablea el grupo de panel con la cadena aprobada de F2 —
    `authn → guard de cambio de contraseña → authz("ministerios") → CSRF` — **sin helpers nuevos de
    middleware**: `AdminChain("ministerios")` con el permiso **ya sembrado** (P6-6/FR-011: verificado
    **en servidor**; independiente de los demás permisos). Las 2 rutas públicas quedan sin
    middlewares (solo lectura) y `/healthz` intacto (§8.1.9). *Cubre FR-011 (el servidor es la única
    autoridad) y FR-017* (el `Recorder` de T616 recibe cada cadena).
  - **Pruebas incluidas** (§III): humo de composición + de permisos. **Cómo se verifica**:
    `go test ./cmd/api/` y `go test ./internal/ministerios/` en verde: rutas nuevas con su sobre;
    `/healthz` sin sesión y `404` `not_found` en rutas inexistentes (§8.1.9); una cuenta con
    `ministerios` pasa y una cuenta con `eventos`/`actividades`/`categorias` recibe **`403`** en
    `/api/v1/admin/ministerios` (y al revés: una cuenta sin `ministerios` no accede a otros módulos);
    el `403` genérico queda registrado como `denied`; `go test ./...` completo en verde.
  - **Criterio de terminado**: el permiso queda verificado en servidor y auditable de un vistazo en
    `routes.go` (SC-005).
  - **Commit sugerido**: `feat(ministerios): rutas con AdminChain("ministerios")`

---

## Fase 7 — Frontend

> Se **aplica** el sistema de F10 y los patrones de F2/F3/F4 (no se redefinen). Nombres de pieza y
> composición según `ux.md` §6; las pantallas y textos siguen `ux.md`. Ningún `fetch` fuera del
> cliente de API (regla F2); i18n solo en el sitio público (FR-018: el panel va en español).

- [ ] T622 · `api/ministerios.ts` + hooks · `[frontend]`

  - **Archivos**: `frontend/src/api/ministerios.ts` (NUEVO),
    `frontend/src/features/publico/ministerios/hooks/usePublicMinistries.ts`,
    `hooks/usePortadaMinisterios.ts`, `frontend/src/features/ministerios/hooks/useMinistries.ts`,
    `useCreateMinistry.ts`, `useUpdateMinistry.ts`, `useDeleteMinistry.ts`,
    `useUploadMinistryImage.ts` (NUEVOS).
  - **Qué hace**: cliente tipado desde `schema.d.ts` (T605): `getMinisterios(lang)` y
    `getPortadaMinisterios(lang)` (públicos) y panel (list/create/update/delete) +
    `uploadMinistryImage`; hooks TanStack Query (`usePublicMinistries` con key `['ministerios', lang]`,
    `usePortadaMinisterios` con `['portada-ministerios', lang]` —`placeholderData: keepPreviousData`
    y retención de la última respuesta válida, patrón de F4/D-23: sin parpadeos al cambiar de
    idioma— y los del panel con invalidación tras mutaciones para que el cambio sea visible en la
    primera carga, SC-003). El cliente **no filtra ni reimplementa** el fallback ni la privacidad: el
    servidor ya resolvió todo. *Soporta US1–US3 y US6*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: cada hook contra un handler MSW (éxito, `400` con `details`,
    `403`, `404`, `409`); respuesta sin `items` (clave omitida); error con datos previos retenidos;
    invalidación de caché tras crear/editar/eliminar.
  - **Criterio de terminado**: ningún componente hace `fetch` directo (regla F2) y los tipos salen
    solo del contrato vivo (0 `any`).
  - **Commit sugerido**: `feat(frontend): cliente y hooks de ministerios`

- [ ] T623 · Promoción de `CardImage` al inventario único · `[frontend]` `[P]` (P3)

  - **Archivos**: `frontend/src/components/publico/CardImage.tsx` y su test (MOVIDOS desde
    `frontend/src/features/publico/eventos/ItemCardParts.tsx`); imports ajustados en
    `features/publico/eventos/` (`EventCard`, `ActivityCard`, `UpcomingItem` — EDITADOS).
    **`HiddenNotice` NO se promueve** (P6-17/P6-21: F6 redirige en vez de avisar y F4 lo sigue
    usando en `/eventos`).
  - **Qué hace**: promueve la pieza de imagen de la tarjeta al inventario único
    `src/components/publico/` (regla de F2/T242: lo compartido se promueve en lugar de copiarlo)
    para que F6 la reutilice en fichas y extracto. **Sin cambios de comportamiento**: mismos props,
    mismos textos, mismas pruebas (se **mueven** con el componente; `onError` retira la imagen sin
    romper el layout). *Soporta FR-020 (imagen opcional sin hueco) en las superficies nuevas.*
  - **Pruebas incluidas** (§III): las pruebas existentes del componente **movidas** + ajuste de
    imports. **Cómo se verifica**: `npm test -- --run` en verde (**regresión de F4**: `EventCard`,
    `ActivityCard` y `UpcomingItem` siguen renderizando su imagen con `alt`).
  - **Criterio de terminado**: una sola copia de `CardImage` en el inventario y F4 sin regresión.
  - **Commit sugerido**: `refactor(frontend): promover CardImage al inventario publico`

- [ ] T624 · i18n público y textos del panel · `[frontend]` `[P]` (P3)

  - **Archivos**: `frontend/src/features/publico/i18n/{messages,es,en}.ts` (EDITADOS: claves
    tipadas de `ux.md` §7.2 — entrada de navegación «Ministerios»/«Ministries», H1 y subtítulo de la
    sección (D-2), etiquetas «Horario»/«Lugar»/«Encargado»/contacto, botón «Unirse al grupo de
    WhatsApp», chip «Ministerio», botón «Ver más ministerios» y textos del bloque —) y
    `frontend/src/features/ministerios/messages.ts` (NUEVO: catálogo del panel de `ux.md` §7.1).
    Sin la clave, TypeScript **no compila** (garantía de SC-006).
  - **Qué hace**: completa el diccionario bilingüe del sitio público y los textos en español del
    panel (FR-018: el panel no pasa por `t()`). Traducciones fieles a `ux.md` §7 (incluida la ayuda
    del interruptor de `ux.md` §4.5).
  - **Pruebas incluidas** (§III): Vitest (tabla de casos). **Cómo se verifica**:
    `npm test -- --run` en verde: las claves `es`/`en` son simétricas (sin huecos); los textos del
    panel son cadenas en español (0 claves de i18n en `messages.ts` del panel).
  - **Criterio de terminado**: i18n completo del sitio público (SC-006) y catálogo de textos del
    panel listo para T630/T631.
  - **Commit sugerido**: `feat(frontend): textos de ministerios (i18n publico y panel)`

- [ ] T625 · Permisos, accesos del panel y `ACTION_LABELS` · `[frontend]`

  - **Archivos**: `frontend/src/lib/permissions.ts` (EDITADO: `MINISTERIOS = 'ministerios'` y su
    entrada en `MODULE_PERMISSIONS`), `frontend/src/features/roles/permissions.ts` (EDITADO: el sello
    «Disponible más adelante» deja de aplicarse a `ministerios` — el catálogo se lee dinámico de la
    API y **no** cambia —), `frontend/src/features/panel/components/AccesosRapidos.tsx` y
    `frontend/src/features/panel/pages/InicioPage.tsx` (EDITADOS: tarjeta «Gestión de ministerios»
    **solo** con el permiso), `frontend/src/app/layout.tsx` (EDITADO: entrada del menú del panel
    filtrada por permiso) y `frontend/src/features/auditoria/messages.ts` (EDITADO: `ACTION_LABELS`
    cubre los **8 códigos** `ministry.*`, `ux.md` §4.8/§7.1: «Creó/Modificó/Eliminó el ministerio
    "…"», «Publicó/Retiró el ministerio "…"», «Publicó/Ocultó la ficha del encargado del ministerio
    "…"»). *El registro de rutas del SPA va con las páginas (T627 pública, T630 panel) para que cada
    commit compile (nota 7 de la alineación de nomenclatura).*
  - **Qué hace**: activa en el frontend el permiso existente `ministerios` (FR-011: **sin permisos
    nuevos**): el formulario de roles de F2 lo ofrece (sin sello), el menú y los accesos rápidos se
    filtran por permiso (una cuenta solo `ministerios` no ve «Eventos» ni «Información») y la
    auditoría muestra etiquetas legibles para los 8 códigos (FR-017). *Cubre FR-011 (visibilidad del
    acceso en el panel), FR-017 (etiquetas) y FR-018.*
  - **Pruebas incluidas** (§III): Vitest + Testing Library (las del propio cambio; las 4 que fijan
    listas de permisos van en T626). **Cómo se verifica**: `npm test -- --run` en verde (con T626
    integrada): `isPermissionAvailable('ministerios') === true` y el sello fuera; el menú renderiza
    la entrada solo con el permiso; **la tarjeta de `AccesosRapidos` se muestra solo con
    `ministerios`** (y ninguna con la cuenta sin permisos de F6); los **8 códigos** `ministry.*`
    tienen etiqueta humana en `ACTION_LABELS`.
  - **Criterio de terminado**: el acceso «Gestión de ministerios» es visible **solo** con el permiso
    `ministerios` y la auditoría no muestra códigos crudos de F6 (`ux.md` §4.8).
  - **Commit sugerido**: `feat(frontend): permiso ministerios, accesos del panel y etiquetas de auditoria`

- [ ] T626 · Actualización **acotada** de las 4 pruebas que fijan permisos/accesos · `[frontend]` *(cambio acotado documentado, RG6-6)*

  - **Archivos** (SOLO estos 4): `frontend/src/lib/permissions.test.ts`,
    `frontend/src/features/roles/permissions.test.ts`, `frontend/src/features/roles/roles.test.tsx`
    y `frontend/src/features/panel/panel.test.tsx` (EDITADOS). *Nada más se toca.*
  - **⚠️ CAMBIO ACOTADO Y DOCUMENTADO (RG6-6)**: F6 edita pruebas de F2/F4 (funcionalidades
    cerradas) porque el catálogo de accesos del frontend crece con `ministerios`. Es un cambio de
    producto **autorizado por el plan**, no un debilitamiento: solo cambian **recuentos, listas y
    sellos**; ninguna aserción de comportamiento se relaja. El backend **no** se toca (el catálogo de
    permisos de F2 sigue en 10). La tarea queda **registrada como excepción** en el PR y la validan
    `revisor-codigo` y `seguridad`.
  - **Qué hace**: ajusta las aserciones que fijan la lista de permisos y accesos para incluir
    `ministerios`: la lista de `MODULE_PERMISSIONS` (`lib/permissions.test.ts`), el sello «Disponible
    más adelante» que deja de aplicarse a `ministerios` (`features/roles/permissions.test.ts` y
    `roles.test.tsx`) y la tarjeta nueva de accesos rápidos (`features/panel/panel.test.tsx`).
    Verifica además que el formulario de roles ofrece el permiso **solo** (sin sello). *Cubre FR-011.*
  - **Pruebas incluidas** (§III): son las propias pruebas actualizadas. **Cómo se verifica**:
    `npm test -- --run` en verde; el diff toca **solo** recuentos/listas/comentarios de esos 4
    archivos (se revisa el diff: ninguna aserción de comportamiento cambia); el PR lo registra como
    **cambio acotado** (no debilitamiento).
  - **Criterio de terminado**: `make ci` pasa con `ministerios` en el catálogo de accesos y la
    excepción queda documentada en la evidencia.
  - **Commit sugerido**: `test(frontend): permisos y accesos con ministerios (cambio acotado RG6-6)`

- [ ] T627 · `/ministerios`: `MinisteriosPage`, `MinistryCard`, `EncargadoLine`, `WhatsappJoinButton` y **redirección** · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/ministerios/MinisteriosPage.tsx`, `MinistryCard.tsx`,
    `EncargadoLine.tsx`, `WhatsappJoinButton.tsx` y `*.test.tsx` (NUEVOS),
    `frontend/src/app/router.tsx` (EDITADO: ruta pública **`/ministerios`** dentro de
    `PublicLayout`).
  - **Qué hace**: la página con H1 «Ministerios» + subtítulo fijo (D-2) y **una grilla única** de
    fichas (`RevealGroup`, todos los publicados en orden `sortOrder, id`, sin paginar) compuesta con
    el inventario de F10 (`Card solid`, `Reveal`…). `MinistryCard` dibuja **cada elemento solo si
    tiene dato**: `CardImage` opcional (sin hueco; `onError` la retira), nombre (h3), descripción
    **completa** (sin «Ver más»), línea de **horario** y de **lugar** solo con dato — **texto tal
    cual**, con icono, sin `<time>`, sin `DateBadge` y sin calcular nada (FR-024/SC-015) —,
    `EncargadoLine` **solo si viene `leader`** (nombre; y contacto —📱 celular o ✉️ correo— solo si
    existe; nunca línea vacía ni nombre huérfano: FR-026/SC-017) y `WhatsappJoinButton` **solo si
    hay `whatsappUrl`** (variante verde de `PublicButton`, icono + texto «Unirse al grupo de
    WhatsApp», `target="_blank" rel="noopener"`, nombre accesible «Unirse al grupo de WhatsApp de
    {nombre}», `href` = valor guardado **sin transformar**, FR-003/FR-015). **Redirección silenciosa
    (P6-21, RG6-13)**: `SectionSkeleton` mientras `usePublicMinistries` está en vuelo y **solo con
    éxito y sin `items`** → `<Navigate to="/" replace />` —**sin aviso** (aclaración de la puerta 2),
    sin pintar jamás la grilla vacía ni `HiddenNotice`—; **ante error NO se redirige** (H1 + mensaje
    «No pudimos cargar los ministerios…» + «Volver a intentar»; con caché, contenido + aviso
    discreto — patrón de F3/`ux.md` §4.1). Estados de carga/error/imagen rota de `ux.md` §4.1 y
    `prefers-reduced-motion` respetado. *Cubre US1, US4, US5 y FR-002, FR-003, FR-005, FR-006,
    FR-008, FR-022, FR-023, FR-024, FR-026 (UI).*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: grilla en orden `sortOrder, id`; cada caso «sin dato» (sin imagen,
    sin horario/lugar, sin enlace → **sin botón**, con ficha oculta → **sin línea de encargado**:
    0 huecos, SC-017); con `leader` oculto **el DOM público no contiene el nombre ni el contacto**
    (SC-014 en frontend); **redirección sin parpadeo**: con la consulta en caché redirige en el
    primer render sin pintar H1 ni grilla; sin caché solo se ve el esqueleto mientras se decide;
    `<Navigate replace>`; y **ante error NO redirige** (se muestra el estado de error); `href` =
    valor guardado; `alt` en las imágenes; fallback `en → es` visible (nunca campo vacío); el texto
    libre largo («Una vez al mes, fecha variable») se acomoda sin romperse.
  - **Criterio de terminado**: `quickstart.md` §1/§4/§5/§7 ejecutables en navegador (US1/US4) y la
    redirección silenciosa sin destello verificada (requisito de la puerta 2, RG6-13).
  - **Commit sugerido**: `feat(frontend): pagina /ministerios con redireccion silenciosa sin publicados`

- [ ] T628 · `useSiteLinks`: entrada «Ministerios» condicional · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/hooks/useSiteLinks.ts` y su test (EDITADOS).
  - **Qué hace**: `SiteLinkId` crece con **`'ministries'`** y la entrada única **«Ministerios»** con
    href **`/ministerios`** existe **solo si** `usePortadaMinisterios` trae `items` (P6-9: la
    presencia se deduce del **documento del bloque** — mismo predicado "publicado", 0 peticiones
    extra en la portada —, criterio técnico confirmado por el humano). **Depende de T622 y T624**
    (A6-5): usa la clave i18n tipada **`nav.sections.ministries`** que crea T624 (sin ella este
    commit no compila, igual que T627/T629). La consumen a la vez
    `PublicNav`, `MobileMenu` y `FooterSiteMap` (grupo **Explorar** del mapa del sitio; fuente única:
    barra, menú hamburguesa y pie **a la vez**, sin huérfanos ni enlaces rotos, FR-005). Sin
    publicados, el ítem **no existe** en ninguna de las tres superficies. *Cubre FR-005, FR-006 y el
    edge case "navegación sin enlaces rotos"* (SC-010).
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: con ítems el ítem aparece en las tres superficies con
    `href="/ministerios"`; sin `items` (clave omitida), en ninguna; el resto de enlaces (anclas de
    portada, `/eventos`) no cambia.
  - **Criterio de terminado**: ocultación coherente del ítem (FR-005) verificada en las tres
    superficies (SC-010).
  - **Commit sugerido**: `feat(frontend): navegacion con /ministerios condicional en barra, menu y pie`

- [ ] T629 · `MinisteriosBlock` + `MinistryItem` en la portada · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/ministerios/MinisteriosBlock.tsx`, `MinistryItem.tsx`
    y `*.test.tsx` (NUEVOS), `frontend/src/features/publico/pages/HomePage.tsx` (EDITADO: **una
    inserción** de `<MinisteriosBlock />` **inmediatamente después de `<ProximamenteBlock />` y antes
    de `<WhoWeAreSection />`**) y `homePage.test.tsx` (AMPLIADO).
  - **Qué hace**: bloque «Ministerios» con el patrón del bloque «Lo que viene» de F4 (P6-16):
    `SectionHeading` + fila/grilla de **hasta 5 `MinistryItem`** en el orden manual (`sortOrder,
    id`) —tarjeta compacta con `CardImage` opcional (sin hueco al faltar), chip neutro «Ministerio»
    (D-11), **título que es el enlace** a `/ministerios` (nombre accesible «Ministerio: {nombre}»,
    máx. 2 líneas con elipsis), líneas de horario/lugar **solo con dato** (texto truncado con
    `title`; sin `<time>`) y **sin encargado ni botón de unión** (D-12)— y **botón «Ver más
    ministerios»** a `/ministerios` (solo con ítems). **Sin publicados el bloque devuelve `null`: NO
    se renderiza** —sin título, sin mensaje, sin hueco— (FR-028: regla general de F3, **diferencia
    consciente** con `ProximamenteBlock`, que sí muestra aviso). Esqueleto propio mientras carga
    (D-9); **error sin caché → no se dibuja** (D-10); error con caché → contenido + aviso discreto.
    Sistema de F10 (`Card`, `SectionHeading`, `Reveal`), responsive desde 320 px (1 col móvil, 2
    tableta, 5 escritorio; **sin carrusel ni scroll lateral**), foco visible, `section` con
    `aria-labelledby` al h2, ítems como `li` con `h3`. **La portada de F3 no se toca** más allá de
    esta inserción (FR-021/FR-029). *Cubre US6, FR-027, FR-028, FR-029.*
  - **Pruebas incluidas** (§III): Vitest + Testing Library. **Cómo se verifica**:
    `npm test -- --run` en verde: **5 exactos** con 8 publicados (los 5 primeros por `sortOrder,
    id`); exactos con menos, sin huecos ni marcadores; **con 0 publicados el bloque no se renderiza**
    (sin título ni mensaje) y `homePage.test.tsx` confirma que **las demás secciones de la portada se
    renderizan igual que antes**; el botón **ausente** sin ítems; las tarjetas **no** muestran
    encargado ni botón de unión (D-12); en inglés traducido con fallback; sin imagen la tarjeta
    conserva tamaño.
  - **Criterio de terminado**: US6/FR-027/FR-028 en pantalla, en la posición acordada (tras el bloque
    de eventos y antes de «quiénes somos»), y `quickstart.md` §14 ejecutable en navegador.
  - **Commit sugerido**: `feat(frontend): bloque de ministerios en la portada`

- [ ] T630 · Panel: `MinistriesPage` + `MinistriesList` + ruta `/panel/ministerios` · `[frontend]`

  - **Archivos**: `frontend/src/features/ministerios/pages/MinistriesPage.tsx`,
    `components/MinistriesList.tsx` y tests (NUEVOS), `frontend/src/app/router.tsx` (EDITADO:
    **`/panel/ministerios`** con `RequirePermission(MINISTERIOS)`).
  - **Qué hace**: gestión en el panel **en español** (FR-018) con la composición de `ux.md` §4.3/§5.a:
    título + banda de ayuda; listado (tabla en desktop / tarjetas en móvil, patrón `EventsList`) con
    columnas **Orden** (el valor numérico, D-6), Nombre, Horario/Lugar («—» si no hay), Encargado
    (el nombre guardado **o** «—», con la etiqueta **«Ficha oculta»** cuando `leaderVisible` está
    apagado — el panel **siempre** ve el dato guardado —), Estado (`StatusPill` borrador/publicado;
    **sin** marca «Pasado»: nada vence) y Acciones (`Editar` · `Publicar`/`Retirar de la sección` ·
    `Eliminar`); `ConfirmDialog` suave para retirar y danger para eliminar («se eliminará para
    siempre…»); estados de `ux.md` §5.a (cargando `EmptyState`, vacío con botón, **«Nada está
    publicado ahora: … /ministerios redirige a la portada»** cuando todo está en borrador, error con
    «Reintentar», éxito con `Notice` de `ux.md` §7.1) y sin permiso la ruta cae en `/sin-permiso` de
    F2. El listado va ordenado `sortOrder, id` (el mismo orden que consume el público, US2 esc. 9).
    *Cubre US2, US3 y FR-010, FR-012 (estado visible), FR-016, FR-018.*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: el listado muestra Orden/Estado/«Ficha oculta» solo con el
    interruptor apagado; publicar/retirar/eliminar con sus confirmaciones y mensajes; el estado
    vacío y el aviso «Nada está publicado ahora…»; sin permiso `ministerios` la ruta no es accesible
    (texto de `/sin-permiso` intacto).
  - **Criterio de terminado**: `quickstart.md` §2/§8 ejecutables en el panel (US3).
  - **Commit sugerido**: `feat(frontend): panel de ministerios con listado y acciones`

- [ ] T631 · Panel: `MinistryForm`, `ContactoEncargado` y `ShowContactSwitch` · `[frontend]`

  - **Archivos**: `frontend/src/features/ministerios/components/MinistryForm.tsx`,
    `ContactoEncargado.tsx`, `ShowContactSwitch.tsx`, `validation.ts`, `errors.ts` y tests (NUEVOS).
  - **Qué hace**: alta/edición en `Dialog` con el patrón F3/F4 (`ux.md` §4.4/§4.5): pestañas
    **Español** (obligatorio) / **English (opcional)** con la regla **por campo** —un campo opcional
    rellenado en inglés sin su versión española bloquea el guardado indicando qué falta; **un campo
    vacío no obliga a nada** (Q6)—; Nombre (`Field`, contador, 160) y Descripción (`TextArea`,
    `{n}/2.000`, saltos preservados); Horario y Lugar opcionales con la ayuda de `ux.md` («Se muestra
    tal como lo escribas — no se calculan fechas»); **bloque Encargado (§4.5)**: nombre es/en,
    **`ContactoEncargado`** (Select «— Sin contacto —»/«Celular»/«Correo» + un único `Field`
    dinámico que cambia etiqueta y validación; al cambiar de tipo se borra el anterior; **nunca dos
    medios**, Q8) y **`ShowContactSwitch`** (`role="switch"`, etiqueta «Mostrar la ficha del
    encargado en el sitio público», **APAGADO por defecto al crear** — D-13, «Oculto por defecto»,
    P6-3 — y precarga **su última decisión** al editar; leyendas de `ux.md` §4.5/D-8: ayuda fija,
    `Notice` warning claramente visible con el interruptor apagado y encargado con datos, leyenda
    verde solo al encender); campos no lingüísticos: **Enlace** (Zod estricto `https`: rechaza
    `javascript:`/`data:`/`vbscript:`/`http://`; ayuda «Cualquier https válido»), **Orden** numérico
    (D-6) e **Imagen** (`ImageUploader`, alt obligatorio en español); **`whatsappUrlShared`** como
    `Notice` **no bloqueante** tras guardar (D-7); errores junto al campo con foco en el primero
    (FR-014), `Esc` no descarta lo escrito sin aviso, `validation.ts` espejo de las reglas del server
    y `errors.ts` traduce `details` por campo. *Cubre US2, FR-010, FR-014, FR-015, FR-025 (UI).*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: **crear con encargado sin tocar el interruptor → se guarda con
    `leaderVisible: false`** («oculto por defecto»: el interruptor nace apagado, D-13); encender y
    apagar el interruptor **sin perder datos** (US2 esc. 10–11); celular **y** correo juntos
    rechazados con el error del campo (Q8); enlace `javascript:` rechazado sin guardar (SC-009);
    `descriptionEn` sin `descriptionEs` → error que indica qué falta; límites con sus contadores;
    duplicado (`409`) mostrado sin perder lo escrito; aviso de enlace compartido no bloquea el
    guardado; editar precarga la última decisión del interruptor; subida de imagen con error de
    tipo/tamaño no rompe el formulario.
  - **Criterio de terminado**: US2 ejecutable en < 2 min según `quickstart.md` §3 (SC-004) con el
    «oculto por defecto» reflejado en la interfaz (D-13).
  - **Commit sugerido**: `feat(frontend): formulario de ministerio con privacidad del encargado`

---

## Fase 8 — E2E

> Toda la Fase 8 exige el backend completo (T621) integrado y la base de F6 migrada. Las specs se
> ejecutan con `make e2e` **por tandas** respetando el **límite de intentos de login** de F2 (las
> specs de panel hacen login; ver T637).

- [ ] T632 · `ministeriosSeed.ts` (fixtures del módulo) · `[frontend]`

  - **Archivos**: `frontend/e2e/ministeriosSeed.ts` (NUEVO; patrón de `frontend/e2e/agendaSeed.ts`).
  - **Qué hace**: fixtures reutilizables por las 4 specs e2e (patrón de F4): ministerios publicados
    en orden conocido (con horario/lugar, con imagen+alt, con enlace de WhatsApp, con encargado
    **visible** y con encargado **oculto pero con nombre y celular/correo reales** — el caso SC-014
    —), uno **sin enlace y sin encargado mostrable** (SC-017), uno solo con nombre y descripción,
    borradores (SC-002), **8 publicados** para el tope 5 (SC-016), dos ministerios con el mismo
    enlace (aviso D-7), textos largos y con tildes/`ñ`/`¿¡`/emojis, y el escenario **0 publicados**
    para la redirección (SC-010). *Soporta US1–US6.*
  - **Pruebas incluidas** (§III): — (fixture e2e). **Cómo se verifica**: las 4 specs la consumen sin
    declarar datos propios; los datos se crean/limpian de forma determinista entre tandas.
  - **Criterio de terminado**: fixtures listas y documentadas (qué caso cubre cada una).
  - **Commit sugerido**: `test(e2e): fixtures de ministerios`

- [ ] T633 · E2E Playwright `ministerios-publica.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/ministerios-publica.spec.ts` (NUEVO; usa `helpers.ts` y
    `ministeriosSeed.ts`).
  - **Qué hace**: recorrido del visitante (US1, US4, US5) según `quickstart.md`
    §1/§4/§5/§7/§8/§9/§11: sin sesión ve los ministerios publicados en orden `sortOrder, id` con
    descripción, horario/lugar, imagen y enlace cuando existen; **caso ficha oculta (SC-014,
    RG6-1)**: con `leaderVisible: false` y nombre/celular reales, **0 apariciones** del nombre ni del
    contacto en la **respuesta de red** (todas las peticiones, incluida la API forzada sin sesión) y
    en el **HTML** (`document.documentElement.outerHTML`), y 0 líneas de encargado en la interfaz;
    **caso sin enlace**: sin botón de unión ni huecos (SC-009/SC-017); borradores invisibles
    (SC-002); **«oculto por defecto»**: un alta hecha por la API sin `leaderVisible` nace oculto y el
    público no recibe nada del encargado (P6-3); **sin publicados: la visita directa a `/ministerios`
    redirige a la portada** — URL termina en `/`, la portada carga completa y navegable, **sin
    destello**: en ningún instante se pinta la grilla vacía, el H1 ni aviso alguno, y «Atrás» no
    devuelve a `/ministerios` (RG6-13, `replace`)— y la entrada sigue oculta en barra/menú/pie;
    inglés con fallback visible y vuelta a español; `href` del botón = valor guardado; 3 anchos
    (320/768/1280, SC-007). *Cubre SC-001, SC-002, SC-006, SC-007, SC-009, SC-010, SC-014, SC-017.*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde sobre la
    base con las fixtures de T632; el resto de e2e (F3/F4) sigue pasando sin tocar.
  - **Criterio de terminado**: US1/US4/US5 y la privacidad del encargado (SC-014) cubiertos de
    extremo a extremo con inspección de tráfico y de HTML.
  - **Commit sugerido**: `test(e2e): seccion publica de ministerios (privacidad, enlace y redireccion)`

- [ ] T634 · E2E Playwright `ministerios-panel.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/ministerios-panel.spec.ts` (NUEVO).
  - **Qué hace**: recorrido del panel (US2, US3) según `quickstart.md` §2–§6/§8/§12/§13: login con
    permiso `ministerios`, **crear+publicar un ministerio en < 2 min** (SC-004), crear uno solo con
    nombre y descripción, **crear uno con encargado sin tocar el interruptor — nace oculto
    («Oculto por defecto», P6-3) —**, apagar y encender el interruptor sin perder datos (US2 esc.
    10–11), enlace `javascript:` rechazado sin guardar y celular **y** correo juntos rechazados
    (Q8/SC-009), reordenar con el campo numérico y ver el orden en público (SC-003/FR-016),
    publicar/retirar/eliminar con confirmación (SC-002/SC-003), **denegación cruzada** (cuenta sin
    `ministerios` —incluida una con otros permisos de módulo—: `403` en la API forzada y mensaje
    claro en la interfaz, SC-005) y **todo registrado en la auditoría de F2** con los códigos
    `ministry.*` y sus etiquetas legibles (SC-011). *Los logins se ejecutan por tandas respetando el
    límite de intentos de F2 (ver T637).* *Cubre SC-003, SC-004, SC-005, SC-011.*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde con la
    cuenta `ministerios` y con las cuentas de denegación de T632; la auditoría muestra los códigos
    esperados con su etiqueta humana (`ACTION_LABELS` de T625).
  - **Criterio de terminado**: US2/US3 cubiertas de extremo a extremo con el permiso verificado en
    servidor y la privacidad del interruptor probada en el flujo real.
  - **Commit sugerido**: `test(e2e): gestion de ministerios en el panel y privacidad del encargado`

- [ ] T635 · E2E Playwright `ministerios-accesibilidad.spec.ts` (teclado y foco) · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/ministerios-accesibilidad.spec.ts` (NUEVO; patrón de
    `portada-accesibilidad.spec.ts`).
  - **Qué hace**: recorrido por **teclado** de `/ministerios` y del bloque de la portada (FR-023,
    SC-008, quickstart §11): el primer `Tab` enfoca el enlace de salto al contenido; la travesía con
    `Tab` alcanza navegación, fichas, `WhatsappJoinButton` (activable con `Enter`/`Espacio`),
    selector de idioma y bloque de la portada en orden comprensible; **foco visible en cada paso**;
    **sin trampas de foco** (`Shift+Tab` sale de cualquier panel); el selector de idioma se opera con
    teclado sin perder el foco; y **la redirección de `/ministerios` sin publicados no atrapa el
    foco** (el visitante termina en una portada navegable, P6-21/RG6-13). Se comprueba en **tema
    claro y oscuro** (F10). *Cubre SC-008 y FR-023 (teclado y foco; la checklist WCAG completa la
    cierra T637).*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde sobre la
    base con fixtures publicadas y, con 0 publicados, la redirección deja el foco en una página
    operable.
  - **Criterio de terminado**: el "e2e de teclado" de la sección y del bloque existe y corre en
    `make e2e` junto a las otras specs.
  - **Commit sugerido**: `test(e2e): accesibilidad por teclado de ministerios (SC-008)`

- [ ] T636 · E2E Playwright `portada-ministerios.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/portada-ministerios.spec.ts` (NUEVO).
  - **Qué hace**: bloque de la portada (US6, SC-016) según `quickstart.md` §14: con 8 publicados
    muestra **exactamente los 5 primeros** en orden `sortOrder, id` (con imagen, chip «Ministerio» y
    horario/lugar cuando existen); con menos, exactos, **sin huecos ni marcadores**; con **0
    publicados el bloque no existe** (sin título, sin mensaje — a diferencia de «Lo que viene» de F4
    —) y la portada queda coherente; **posición**: inmediatamente después del bloque de eventos/agenda
    y **antes de «Quiénes somos»**; el título-enlace de cada tarjeta y el botón «Ver más ministerios»
    llevan a `/ministerios` en un solo paso; 0 borradores (SC-002); 320 px sin desplazamiento
    horizontal (SC-007) y recorrido por teclado con foco visible (SC-008). *Cubre US6, SC-016 y los
    aspectos de portada de SC-010/SC-013.*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**:
    `npx playwright test e2e/portada-ministerios.spec.ts` en verde; los e2e de F3/F4 de la portada
    (`portada-publica`, `portada-proximamente`) siguen pasando sin tocar.
  - **Criterio de terminado**: US6/SC-016 cubiertos de extremo a extremo en la posición acordada.
  - **Commit sugerido**: `test(e2e): bloque de ministerios de la portada`

---

## Fase 9 — Cierre

- [ ] T637 · Verificación de cierre · `[infra]`

  - **Archivos**: ninguno (evidencia en el PR). *No se toca* `.github/workflows/ci.yml` ni ningún
    archivo del kit. `.env.example`: **sin cambios** (F6 no añade variables — P6: `APP_TZ` es de F4
    —; solo se verifica que ninguna hiciera falta).
  - **Qué hace**: ejecuta la verificación completa de F6 y la registra como evidencia: `make ci`
    (**lint + test + security**; `make db-migrate` al preparar el entorno, §0); **cobertura ≥ 80 %
    en `internal/ministerios/service*.go`** (`go test -cover`); migraciones `000010`/`000011`
    **up → down → up** (y `permissions` sigue en 10 con `ministerios`); `make sqlc-verify` y
    `make api-gen` **sin deriva**; `make e2e` **por tandas** respetando el **límite de intentos de
    login** de F2 (las specs del panel de F6 y de F2/F4 hacen login: se ejecutan por tandas —o con
    las cuentas y ventanas que no agoten el límite— y la salida de cada tanda se adjunta);
    `.env.example` sin cambios confirmado; recorrido manual de `quickstart.md` §0–§16 (mapa §16
    (SC-001…SC-017) y títulos de §1–§14 (US/FR)); `GET /healthz` responde como en F1 y una ruta
    inexistente devuelve `404` con `ErrorEnvelope` (§8.1.9). La **checklist de accesibilidad**
    (quickstart §11, WCAG 2.1 AA, claro y oscuro) la recorre `qa-tester`. **Prueba de usabilidad
    SC-012** (protocolo mínimo, estilo F3/F4/M7): **≥ 6 personas** (2 por franja 18–35 / 36–59 /
    60+, incluidas con poca experiencia en internet), guion de **3 tareas sin ayuda** (identificar de
    qué trata un ministerio y cómo participar; encontrar el horario/lugar de un ministerio concreto;
    unirse por el enlace o acercarse al encargado cuando su contacto está publicado), sesión breve
    observada por `qa-tester` con el humano; resultado documentado en
    **`specs/006-ministerios/pruebas-usabilidad-SC-012.md`** (umbral: ≥ 90 % de tareas completadas).
    **El umbral de SC-001 (< 30 s) se cronometra aquí** (el e2e no puede medirlo): la **tarea 1** —
    ver los ministerios sin cuenta y encontrar cómo unirse — se ejecuta con **cronómetro** y el
    tiempo de cada participante se registra en el informe (se da por cumplida con < 30 s por
    participante; umbral marcado como cronometrado en la fila de SC-001 de la cobertura).
    Es una prueba **manual con personas** (no automatizable) y queda fuera de `make ci`/`make e2e`.
  - **Pruebas incluidas** (§III): — (es la ejecución de todas las anteriores). **Cómo se verifica**:
    cada comando anterior en verde y su salida adjunta al PR; el mapa de criterios → secciones queda
    recorrido completo (mapa §16 para SC y los títulos de §1–§14 para US/FR) y el informe de
    usabilidad de SC-012 adjunto **con los tiempos cronometrados de SC-001**.
  - **Criterio de terminado**: `make ci` y `make e2e` (por tandas) en verde; SC-001…SC-017
    evidenciados (con foco explícito en SC-014: 0 apariciones del encargado oculto); el **cambio
    acotado RG6-6** de T626 y la **regresión de imágenes RG6-5** de T610 quedan registrados en la
    evidencia.
  - **Commit sugerido**: — (sin cambios de código)

- [ ] T638 · Documentación y PR · `[infra]`

  - **Archivos**: `README.md` y `docs/` (EDITADOS por `documentador`); PR en GitHub (creado por
    `devops`).
  - **Qué hace**: cierra la entrega con `documentador` (README del proyecto y notas de entrega: qué
    agrega F6, la sección `/ministerios`, el permiso existente `ministerios` **sin migración de
    permisos**, la tabla `ministries` y las migraciones `000010`/`000011`, la **privacidad del
    encargado con «oculto por defecto»** (FR-026/SC-014), la **redirección silenciosa** de
    `/ministerios` sin publicados (distinta a propósito de `/eventos` de F4), el bloque de la portada
    (`GET /api/v1/portada/ministerios`, tope 5) y la documentación de API actualizada tras la fusión
    del contrato 0.7.0) y con `devops` (rama → PR con descripción, tareas cubiertas y los dos
    cambios documentados explícitamente: **RG6-6** — las 4 pruebas de frontend de permisos/accesos —
    y **RG6-5** — la regresión de imágenes de F3/F4 verificada —). *Cubre el cierre de
    US1–US6/FR-001…FR-029 como entregable.*
  - **Pruebas incluidas** (§III): — (documentación y proceso). **Cómo se verifica**: la documentación
    refleja el comportamiento real (la prueba es el `quickstart.md` ejecutado en T637); el PR
    referencia este `tasks.md` con las tareas marcadas y pasa por la revisión final (`qa-tester`,
    `revisor-codigo`, `seguridad`).
  - **Criterio de terminado**: PR abierto y **aprobado por el humano** (regla 1: merge y producción
    requieren aprobación explícita). El merge lo ejecuta el orquestador con `devops`.
  - **Commit sugerido**: `docs: notas de entrega de F6 (ministerios)`

---

## Cobertura de requisitos (US1–US6, FR-001…FR-029, SC-001…SC-017)

| Historia | Tareas | Verificación principal |
|---|---|---|
| **US1** Conocer los ministerios y unirse | T603, T609, T612, T613, T617, T622, T624, T627, T628, T633 | quickstart §1/§5/§7 · e2e T633 (SC-001, SC-009, SC-017) |
| **US2** Gestionar los ministerios desde el panel | T608, T615, T619, T620, T621, T625, T626, T630, T631, T634 | quickstart §2–§6 · e2e T634 (SC-004, SC-005, SC-011) |
| **US3** Controlar la publicación | T613, T614, T615, T627, T629, T630, T633, T634 | quickstart §8 · e2e T633/T634 (SC-002, SC-003, SC-010) |
| **US4** Ver los ministerios en inglés | T613, T614, T624, T627, T629, T633 | quickstart §9 · e2e T633 (SC-006) |
| **US5** Cualquier dispositivo y accesible | T627, T629, T633, T635, T636, T637 | quickstart §11 · e2e en 3 anchos (SC-007) · teclado T635 (SC-008) |
| **US6** Ver los ministerios desde la portada | T614, T618, T629, T636 | quickstart §14 · e2e T636 (SC-016) |

| FR | Tareas | Verificación principal |
|---|---|---|
| FR-001 (ministerio ≠ F5/F4; horario texto libre, sin recurrencia) | T601, T612, T627, T629 | quickstart §1 · pruebas de modelo (sin columnas de fecha) · SC-013 |
| FR-002 (sección pública con ficha completa) | T603, T609, T613, T617, T627 | quickstart §1/§7 · pruebas de service/handler |
| FR-003 (enlace en un paso; sin enlace sin botón) | T612, T627, T631, T633 | quickstart §5 · e2e T633 (SC-009, SC-017) |
| FR-004 (orden `sort_order, id`) | T601, T603, T609, T613, T614, T627, T629, T630 | quickstart §1/§6 · pruebas de orden |
| FR-005 (ocultar sección sin publicados + **redirigir** la URL directa) | T613, T627, T628, T633 | quickstart §1/§8 · e2e T633 (SC-010, RG6-13) |
| FR-006 (ruta `/ministerios` y entrada única de navegación) | T627, T628 | quickstart §1 · e2e T633 (SC-010) |
| FR-007 (bilingüe por campo; español base) | T601, T607, T612, T624, T627 | quickstart §4/§9 · pruebas de validación |
| FR-008 (fallback `en → es`; enlace no traducible) | T613, T614, T627, T629 | quickstart §9 · tabla de casos del service (SC-006) |
| FR-009 (selector de idioma existente) | — (F10 sin cambios), T633 | quickstart §9 · e2e T633 |
| FR-010 (CRUD del panel) | T608, T615, T619, T630, T631 | quickstart §2–§6 · e2e T634 (SC-004) |
| FR-011 (permiso `ministerios` existente, verificado en servidor, independiente) | T621, T625, T626, T630, T634 | quickstart §2/§12 · humo T621 (403 cruzado) · e2e T634 (SC-005) |
| FR-012 (borrador/publicado; 0 borradores por ninguna vía) | T603, T609, T610, T613, T614, T627, T629, T633 | quickstart §8 · pruebas «0 borradores» · e2e T633 (SC-002) |
| FR-013 (se publica al guardar; `no-store`) | T615, T617, T618, T630, T634 | quickstart §8 · e2e T634 (SC-003) |
| FR-014 (validación completa; sin cambios parciales) | T612, T615, T619, T631 | quickstart §3–§5 (tablas de errores) · pruebas de service/handler |
| FR-015 (enlace `https` seguro, sin dominios; se sirve tal cual) | T601, T612, T615, T631, T633, T634 | quickstart §5 · tabla de casos + e2e T634 (SC-009) |
| FR-016 (orden manual editable y respetado) | T615, T630, T631, T634 | quickstart §6 · e2e T634 (SC-003) |
| FR-017 (auditoría atómica, 8 códigos, solo lectura) | T602, T606, T607, T611, T615, T616, T625, T634 | quickstart §13 · integración (mutación+registro) · e2e T634 (SC-011) |
| FR-018 (panel en español) | T624, T630, T631 | revisión de `revisor-codigo` (0 cadenas del panel por `t()`) |
| FR-019 (contenido del ministerio, encargado y contacto opcionales) | T601, T608, T615, T631 | quickstart §3/§4 · pruebas de persistencia |
| FR-020 (imagen opcional con alt) | T601, T610, T620, T623, T627, T631 | quickstart §10 · pruebas de subida/descarga |
| FR-021 (F10 aplicado; solo el bloque en la portada de F3) | T627, T629 | revisión de `revisor-codigo` sobre `HomePage` (una inserción) y `PublicLayout` sin cambios |
| FR-022 (responsividad, navegación simple) | T627, T629, T633, T636 | quickstart §11 · e2e en 320/768/1280 (SC-007) |
| FR-023 (accesibilidad WCAG 2.1 AA) | T627, T629, T635, T637 | checklist de quickstart §11 · e2e de teclado T635 (SC-008) |
| FR-024 (horario/lugar como textos libres, sin cálculos) | T601, T612, T613, T627, T629 | quickstart §4 · pruebas de presentación (SC-015: 0 recurrencias) |
| FR-025 (encargado opcional + interruptor `leaderVisible`, **apagado por defecto**) | T601, T611, T615, T631, T634 | quickstart §3/§4/§7 · pruebas del default en 4 capas (P6-3, RG6-10) |
| FR-026 (**privacidad**: 0 apariciones del encargado no publicado) | T603, T609, T613, T614, T617, T618, T627, T633, T634 | quickstart §7 (inspección de red y HTML) · pruebas «0 apariciones» · e2e T633 (SC-014) |
| FR-027 (bloque de portada: tope 5 en servidor, orden manual, botón, posición) | T614, T618, T629, T636 | quickstart §14 · pruebas del service (SC-016) · e2e T636 |
| FR-028 (bloque oculto por completo sin publicados) | T614, T618, T629, T636 | quickstart §14 · e2e T636 (SC-010, SC-016) |
| FR-029 (bloque bilingüe, responsive y accesible; portada intacta) | T614, T629, T635, T636 | quickstart §9/§11/§14 · e2e T635/T636 (SC-006, SC-007, SC-008) |

| SC | Tareas | Verificación principal |
|---|---|---|
| SC-001 (ve los ministerios y encuentra cómo unirse en < 30 s) | T627, T633, T637 | quickstart §1 · e2e T633 (contenido y botón de unión; **el e2e no cronometra**) · el umbral **< 30 s se cronometra** en el protocolo de usabilidad de SC-012 (**T637**, tarea 1 con cronómetro, tiempo por participante en el informe) |
| SC-002 (0 borradores visibles) | T609, T613, T614, T627, T629, T633 | pruebas «0 borradores» (service/repository) · e2e T633 |
| SC-003 (cambio guardado visible en la primera carga) | T615, T622, T630, T634 | e2e T634 · invalidación de caché (T622) |
| SC-004 (crear y publicar en < 2 min) | T631, T634 | quickstart §3 · e2e T634 |
| SC-005 (permiso verificado en servidor; 0 operaciones sin permiso) | T621, T625, T630, T634 | humo T621 · e2e T634 (cruce de permisos) |
| SC-006 (0 campos vacíos por falta de traducción) | T613, T614, T624, T627, T629, T633 | pruebas de i18n y fallback · e2e T633 |
| SC-007 (usable en teléfono/tableta/PC; 0 scroll horizontal) | T627, T629, T633, T636 | e2e en 320/768/1280 (T633) y bloque a 320 px (T636) |
| SC-008 (accesibilidad; teclado y lector) | T635, T637 | e2e de teclado T635 · checklist WCAG (T637) |
| SC-009 (enlaces seguros; 0 botones vacíos) | T612, T627, T631, T633, T634 | tabla de casos del service · e2e T634 (rechazo) y T633 (`href` = guardado) |
| SC-010 (0 secciones vacías; navegación sin enlaces rotos; redirección limpia) | T627, T628, T629, T633, T636 | pruebas de `useSiteLinks`/bloque · e2e T633 (RG6-13) |
| SC-011 (100 % de operaciones registradas; registro de solo lectura) | T611, T615, T616, T634 | quickstart §13 · integración transaccional · e2e T634 |
| SC-012 (≥ 90 % de personas identifican qué es y cómo participar) | T637 | prueba manual con personas (`pruebas-usabilidad-SC-012.md`) |
| SC-013 (0 F5/F4 en la sección; horario sin fechas) | T601, T627, T629, T633, T636 | revisión de contenido e2e · pruebas de modelo (sin columnas de fecha) |
| SC-014 (**0 apariciones** del encargado no publicado) | T609, T613, T614, T617, T618, T627, T633, T634 | pruebas «0 apariciones» en 4 capas · e2e T633 con inspección de tráfico y HTML |
| SC-015 (horario/lugar tal como se escribieron; 0 cálculos) | T613, T627, T629 | pruebas de presentación · revisión de `revisor-codigo` (sin fechas derivadas) |
| SC-016 (bloque: exactamente 5 o todos; 0 huecos; oculto sin publicados) | T614, T629, T636 | pruebas del service · e2e T636 |
| SC-017 (ministerio sin enlace y sin encargado mostrable: completo y coherente) | T627, T629, T633 | **Lectura fijada (A6-4)**: en el bloque, «completo y coherente» = imagen + nombre + horario/lugar cuando existen + chip, **sin descripción** (Q7/D-12: el detalle vive en `/ministerios`); la descripción es exigible solo a la ficha de la sección. Pruebas de la ficha (0 botones vacíos, 0 líneas vacías) · e2e T633 |

También quedan cubiertos los **Edge Cases** de la spec sin FR propio: **borradores filtrados por
cualquier vía** (T609, T613, T614, T633) · **ministerio publicado que se edita** (visible en la
primera carga: T615, T634) · **sin imagen / imagen rota** (sin huecos, `onError` la retira: T623,
T627, T629) · **sin enlace** (sin botón: T627, T629, T633) · **sin encargado o con ficha oculta**
(sin línea de encargado: T613, T627, T633) · **creado con encargado nace oculto** (P6-3: T601, T611,
T615, T631, T633, T634) · **ficha visible sin contacto** (solo nombre, sin línea vacía: T613, T627)
· **sin horario/lugar** (sin líneas vacías: T627, T629) · **sin enlace y sin encargado** (SC-017:
T627, T633) · **enlace mal formado o no seguro** (rechazado sin guardar, nunca renderizado: T612,
T631, T633) · **enlace caducado** (no se verifica vigencia: fuera de alcance; se documenta) ·
**contacto con formato inválido** (T612, T631) · **horario de texto libre largo** («Una vez al mes,
fecha variable» sin romperse ni calcularse: T613, T627, T629) · **varios ministerios con el mismo
encargado** (sin ficha compartida: T607, T613) · **duplicado exacto → `409`** (T608, T615, T631) ·
**dos ministerios con el mismo enlace** (aviso no bloqueante: T608, T615, T631) · **caracteres
especiales y emojis intactos** (T615) · **descripción sobre el límite** (T612, T631) · **archivo
inválido no rompe la sección** (T620, T627) · **inglés sin español por campo** (T612, T615, T631) ·
**ediciones simultáneas** (última escritura completa gana: T608, T619) · **eliminación definitiva con
confirmación y registrada** (T615, T630) · **muchos ministerios publicados** (tope 200 en sección;
sin límite artificial: T613) · **más de 5 en el bloque** (solo los 5 primeros: T614, T636) ·
**operación sin permiso por cualquier vía** (T621, T625, T630, T634) · **nada se aplica sin
registro** (T611, T615).

**Cambios acotados y documentados registrados**: **T626** actualiza las **4 pruebas de frontend** que
fijan la lista de permisos y accesos (`lib/permissions.test.ts`, `features/roles/permissions.test.ts`,
`features/roles/roles.test.tsx`, `features/panel/panel.test.tsx`) al añadir `ministerios`
(**RG6-6**: cambio de producto autorizado, **no** debilitamiento — solo recuentos/listas/sellos —);
**T610** amplía `IsFilePublished` de forma **aditiva** y conserva las pruebas de `media` de F3/F4
(**RG6-5**: regresión verificada, e2e `portada-imagenes`/`agenda-*` en verde). El backend **no** toca
pruebas de F2: el catálogo de permisos sigue en **10** (a diferencia de F4, que necesitó la excepción
T405).

---

## Aclaraciones de la puerta 2 (2026-10-13) — notas de cierre

> Las aclaraciones **Q1–Q9** de la spec están cerradas y no se reabren; estas tareas las implementan
> literalmente (encargado opcional con contacto opcional —celular **o** correo—, enlace opcional con
> cualquier `https` seguro, imagen opcional, orden manual `sort_order, id`, horario y lugar como
> textos libres sin recurrencia — lectura de Q6 **firme**, confirmada por el humano —, bloque de la
> portada con tope 5, e interruptor que oculta **nombre y contacto** de la ficha del encargado).

1. **Puerta 2-A — «Oculto por defecto»** (`leader_visible DEFAULT FALSE`, P6-3/RG6-10): el
   interruptor de visibilidad de la ficha del encargado **nace apagado**; la recomendación original
   del plan (`TRUE`) quedó **descartada por el humano**. Se fija por prueba en las cuatro capas
   (T611, T613/T615, T617 y T633/T634) y en el formulario (T631).
2. **Puerta 2-B — `/ministerios` sin nada publicado redirige a la portada** (P6-15/P6-21/RG6-13) y
   la precisión posterior de la misma fecha: la redirección es **silenciosa, sin aviso** (la opción
   recomendada). **No se añade anuncio accesible para lectores de pantalla al redirigir**: quedó
   **fuera de alcance** en la última aclaración (el plan lo señalaba como requisito que la spec no
   cubría; sin él la accesibilidad queda razonablemente cubierta porque el destino es una portada
   completa y navegable — ver T627/T635).
3. **La redirección es deliberadamente distinta a la de F4**: en `/eventos` (F4) la sección sin
   contenido muestra `HiddenNotice` **sin redirigir** y ese comportamiento **no se toca**; en
   `/ministerios` (F6) la visita directa **redirige a la portada** (con `replace`, sin parpadeo y
   sin aviso). Es la decisión literal del cliente, **no una inconsistencia** entre secciones; por eso
   `HiddenNotice` **no** se promueve al inventario único (P6-17) y no hay tarea que unifique ambos
   comportamientos.
4. **Decisiones dirimidas entre `ux.md` y el plan** (aplicadas en estas tareas): la presencia de la
   navegación se deduce del **documento del bloque** (P6-9 → T628, criterio técnico confirmado por
   el humano); el interruptor nace **apagado** (D-13 → T631); la sección sin publicados **redirige**
   (D-14 → T627) en lugar del aviso propuesto al principio por `ux.md`.
5. **Nota de implementación (sin efecto en el contrato)**: el registro de las rutas del SPA se
   reparte entre T627 (`/ministerios`) y T630 (`/panel/ministerios`) — en lugar de concentrarse en
   una sola tarea como en F4 — para que cada commit compile; las rutas, los permisos y el contrato
   son exactamente los aprobados.
