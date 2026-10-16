# Tasks: Ofrendas y diezmos (F7 · Donaciones)

**Branch**: `007-donaciones` | **Date**: 2026-10-14
| **Spec**: [spec.md](./spec.md)
(APROBADA v2, US1–US6 / FR-001…FR-025 / SC-001…SC-014 —línea base aprobada—; Q1–Q11 resueltas el
2026-10-14; puerta 2 del 2026-10-14: el plan está **APROBADO** y estas tareas lo implementan sin
reabrirlo) · **Plan**: [plan.md](./plan.md) (P7-1…P7-21, riesgos RG7-1…RG7-12) · **UX**: [ux.md](./ux.md)
(D-1…D-16; pendientes de confirmación humana, ver "Alineación de nomenclatura")

**Input**: plan aprobado (`plan.md`, P7-1…P7-21 + "Propuesta de tareas") · `research.md`
(R7-1…R7-16) · `data-model.md` (migraciones `000012`/`000013`, tablas `donation_methods` +
`donation_texts`, ampliación del `CHECK` de `admin_actions` con 8 códigos `donation.*`; **sin
migración de permisos**: `donaciones` ya está sembrado por `000002`) · `contracts/openapi.yaml`
(delta de diseño, inmutable; **9 operaciones**, `info.version` 0.8.0) · `ux.md` (pantallas,
componentes, textos, estados, D-1…D-16) · `quickstart.md` (§0…§16, verificación) · constitución
(`.specify/memory/constitution.md`) · skills `go-backend`, `postgres-db`, `react-frontend` · F1, F2,
F3, F4, F6 y F10 ya integradas (`specs/001-estructura-base/`, `specs/002-acceso-gestion-usuarios/`
T201…T254, `specs/003-portada-info-general/` T301…T340, `specs/004-eventos-actividades/`
T401…T449, `specs/006-ministerios/` T601…T638, `specs/010-sistema-diseno-movimiento/`).

> **Regla de ejecución**: **un commit por tarea** (Conventional Commits; los hooks del kit lo
> validan). **Test-first** (constitución §III): la prueba de cada tarea se escribe **antes** que su
> código y debe **fallar** antes de implementar; toda tarea de código incluye sus pruebas en el mismo
> commit — backend con `go test` y, donde toca persistencia real, `go test -tags=integration`
> (`//go:build integration`, PostgreSQL real con `testcontainers-go`); frontend con Vitest +
> Testing Library + MSW; flujos críticos con Playwright (`make e2e`). Cobertura exigida **≥ 80 % en
> `internal/donaciones/service*.go`** (`go test -cover`). Quien escribe no aprueba: cada tarea pasa
> por `qa-tester`, `revisor-codigo` y `seguridad` antes de integrarse (flujo del orquestador). **El
> CI del kit (`.github/workflows/ci.yml`) no se toca** y **ninguna migración aplicada se edita**
> (§VI). **Sin excepciones de backend**: F7 no toca pruebas de F2 (el catálogo de permisos **no**
> cambia: sigue en 10); solo la actualización **acotada** de las 2–3 pruebas de frontend que fijan
> permisos/accesos (T725, RG7-6).

## Decisiones del plan (puerta 2, 2026-10-14) que estas tareas respetan (sin reabrir)

- **P7-1 — dos tablas, un dominio**: `donation_methods` (el listado administrable) +
  `donation_texts` (los dos textos, un renglón por clave `intro`/`usage` con `UNIQUE`); consultas en
  `queries/donation_methods.sql`/`donation_texts.sql`; dominio `internal/donaciones/`.
- **P7-2/P7-3/P7-4 — números**: validación y normalización **en el dominio** (`numero.go`, funciones
  puras con tabla de casos, **sin dependencias nuevas**): IBAN por **estructura y longitud exacta por
  país** (registro IBAN embebido, `CR` = 22, **sin mod-97** — default revisable, pregunta al humano
  —) guardado **normalizado a los caracteres del número** (en mayúsculas) y presentado en bloques **en
  el frontend**; SINPE = **celular costarricense de 8 dígitos** (primer dígito 6/7/8 — default
  revisable —) con o sin `+506`, guardado y publicado `+506XXXXXXXX`. `CHECK` por tipo en la BD como
  defensa en profundidad.
- **P7-5 — traducibilidad**: `institution` **NO traducible** (una sola columna), `holder`
  (`holder_es`/`holder_en`) y los `body` de los textos sí; `kind`/`number`/`currency` nunca (SC-006
  se cumple por construcción). *La contradicción de FR-019/Key Entities está señalada como
  **pregunta #1 al humano**; mientras tanto se implementa NO traducible.*
- **P7-6 — permisos**: `AdminChain("donaciones")` con el permiso **ya sembrado** por `000002`;
  **sin migración de permisos**, el catálogo de F2 sigue en 10 y **no se renombra** el permiso (Q8).
- **P7-7 — superficie HTTP**: `GET /api/v1/ofrendas?lang=` (sección) y
  `GET /api/v1/portada/ofrendas?lang=` (bloque, tope 5), ambos `no-store`; panel
  `/api/v1/admin/donaciones…` con `limit`/`offset`, filtro `publicationState` y `PUT` para los
  textos. El público nombra la **sección** (`ofrendas`); el panel, el **módulo** (`donaciones`).
- **P7-8 — presencia de la navegación desde el documento de la SECCIÓN**
  (`usePublicOfrendas` compartido con `useSiteLinks`): la sección puede existir con **solo textos**
  publicados y el bloque solo mira métodos. *Difiere a propósito de F6 y de `ux.md` §3/§10 (que la
  deducen del bloque): ver "Alineación de nomenclatura", punto 9.*
- **P7-9 — bilingüe** `*_es`/`*_en` por campo solo donde hay traducción + `?lang=es|en` resuelto en
  el servidor con fallback `en → es` por campo (nunca un campo vacío, FR-008).
- **P7-10 — duplicado exacto → `409 conflict`** sobre `kind` + `number` **normalizado**
  (`ExistsDonationMethodDuplicate(:kind, :number, :exclude_id)`, que **excluye la fila que se
  edita** con `AND id <> :exclude_id`: una edición que no cambia el número **no** devuelve `409`
  contra sí misma); **sin `UNIQUE`** en la BD (default revisable).
- **P7-11 — auditoría**: 8 códigos `donation.*` (migración `000013`), `InsertAdminAction`
  **transaccional** (fail-closed) y `Recorder` del propio dominio; reordenar = `donation.update`;
  los textos no tienen `create` (`donation.text.update` cubre alta y edición del singleton); el alta
  de un método registra **solo `donation.create`** (también con `publicationState: published`); en
  la edición, datos + estado → **dos filas**.
- **P7-12 — publicación** por método **y por texto** (`publication_state` en las dos tablas), "se
  publica al guardar", consultas `…Published` separadas para el público (SC-002: 0 borradores).
- **P7-13 — "nada publicado" = sin métodos NI textos**: la sección —y su navegación— existe si hay
  ≥1 método **o** ≥1 texto publicado; el bloque solo mira métodos. *Default revisable, **pregunta #2
  al humano**.*
- **P7-14 — bloque de la portada**: endpoint propio `GET /api/v1/portada/ofrendas` (lo registra el
  dominio `donaciones`, Complexity 4), **tope 5 firme en el servidor** (`maxPortadaItems = 5` +
  `LIMIT`), mismo orden y **mismo DTO de ítem** que la sección, **oculto por completo sin métodos
  publicados** (sin mensaje, diferencia consciente con el bloque de F4) y posicionado **tras
  `MinisteriosBlock` y antes de `WhoWeAreSection`** (Q7).
- **P7-15 — validación y límites**: tags existentes (`required`, `max`, `oneof`) + reglas de
  dominio; entidad/titular **160**, textos **2.000**, números por su formato; `400 invalid` con
  `details` por campo y **sin guardado parcial**.
- **P7-16 — sin imágenes** (Q6/FR-020): **cero cambios** en `media`, `platform/storage` e
  `IsFilePublished`; el panel no ofrece carga de imágenes.
- **P7-17 — textos del panel**: `PUT /api/v1/admin/donaciones/textos/{textKey}` (crea o reemplaza el
  singleton) + `GET /api/v1/admin/donaciones/textos`; **sin operación de borrado** (se retiran en
  `draft`).
- **P7-18 — frontend**: `/ofrendas` + `/panel/donaciones` (`RequirePermission(DONACIONES)`),
  `DONACIONES` en `lib/permissions`, tarjeta en `AccesosRapidos`, `ACTION_LABELS` +8; piezas propias
  en `features/publico/ofrendas/` y `features/donaciones/`; **ninguna promoción** al inventario
  único (todo lo necesario ya está promovido por F2/F3/F4/F6/F10) y **`HiddenNotice` no se toca**
  (F7 redirige, no avisa).
- **P7-19 — contrato**: el delta se funde **aditivamente** en `backend/api/openapi.yaml` con
  `info.version` **0.7.0 → 0.8.0**; el registro de `error.code` **no cambia**; el enum de
  `AdminActionItem.action` crece con los 8 códigos `donation.*`.
- **P7-20 — orden**: `ORDER BY CASE kind WHEN 'iban' THEN 0 ELSE 1 END, sort_order, id` en público
  **y** en panel; `sortOrder` se edita por método y manda **dentro de su grupo** (US2 esc. 10).
- **P7-21 — redirección de `/ofrendas` sin parpadeo**: vive en el **frontend** (elemento de ruta):
  `SectionSkeleton` mientras `usePublicOfrendas` está en vuelo y **solo con éxito y sin contenido**
  → `<Navigate to="/" replace />`; **ante error NO se redirige**; la API sigue respondiendo `200`
  con las claves omitidas. **Distinto a propósito de F4 en `/eventos`** (que muestra `HiddenNotice`
  y **no se toca**).

## Alineación de nomenclatura (decidida por el arquitecto sobre `ux.md` / `plan.md`; no se reabre)

1. **Rutas SPA**: la sección pública es **`/ofrendas`** (una sola entrada de navegación «Ofrendas y
   diezmos», `ux.md` D-1) y el panel es **`/panel/donaciones`** (`ux.md` §3). El bloque de la portada
   vive en `HomePage` **inmediatamente después de `<MinisteriosBlock />`** (Q7) y antes de
   `WhoWeAreSection`.
2. **Carpetas de features** (nomenclatura de `ux.md` §6 y del plan P7-18): público en
   **`frontend/src/features/publico/ofrendas/`** y panel en
   **`frontend/src/features/donaciones/`**. Nombres de componentes y hooks: los de `ux.md`
   §6.3/§6.4 (`OfrendasPage`, `IntroText`, `UsageText`, `DonacionesGroup`, `DonacionCard`,
   `NumeroDonacion`, `CopyAccountButton`, `OfrendasBlock`, `OfrendasItem`, `usePublicOfrendas`,
   `usePortadaOfrendas`; panel: `DonacionesPage`, `DonacionesListsByType`, `DonacionForm`,
   `TextoForm`, `NumeroDisplay`, `lib/numero.ts`, `validation.ts`, `errors.ts`, `messages.ts`).
   *El plan P7-18/Complexity 8 declaraba los suyos como **provisionales** hasta que `ux.md`
   existiera; ya existe y manda su catálogo. Las utilidades de presentación viven en
   `features/publico/ofrendas/format.ts` + `features/donaciones/lib/numero.ts` (misma responsabilidad
   que el `format.ts` provisional del plan).*
3. **Rutas de API**: las del contrato (`GET /api/v1/ofrendas`,
   `GET /api/v1/portada/ofrendas`, `GET/POST /api/v1/admin/donaciones`,
   `GET/PATCH/DELETE /api/v1/admin/donaciones/{id}`, `GET /api/v1/admin/donaciones/textos`,
   `PUT /api/v1/admin/donaciones/textos/{textKey}`); el snapshot `contracts/openapi.yaml` es
   inmutable (P7-19).
4. **Permiso**: código **`donaciones`** (el `code` sembrado por `000002`; la etiqueta es
   «Donaciones»). `lib/permissions.ts` expone `DONACIONES`; el guard es el `RequirePermission`
   existente y el servidor usa `AdminChain("donaciones")`.
5. **Numeración T701…T738** (patrón por funcionalidad: F2 `T2xx`, F3 `T3xx`, F4 `T4xx`, F6 `T6xx`;
   **F7 empieza en `T701`**).
6. **Auditoría**: 8 códigos `donation.*` —`donation.create|update|delete|publish|unpublish|
   text.update|text.publish|text.unpublish`— (`data-model.md`, R7-10); `target_label`
   `Donaciones · <institution> · <number>` / `Donaciones · Texto introductorio` /
   `Donaciones · Texto de uso`.
7. **Registro de rutas del SPA**: se reparte entre las tareas de página —`/ofrendas` en T726 y
   `/panel/donaciones` en T729— para que **cada commit compile** (mismo reparto que F6/T627·T630).
8. **`SiteLinkId`**: crece con **`'offerings'`** y la clave i18n es **`nav.sections.offerings`**
   (`ux.md` §7.2); lo consume a la vez `PublicNav`, `MobileMenu` y `FooterSiteMap` (grupo
   **Explorar** del mapa del sitio).
9. **Decisiones dirimidas entre `ux.md` y el plan aprobado** (manda el plan/contrato; `ux.md` ajusta
   composición sin tocar la API):
   - **Presencia de la navegación**: el plan **P7-8** la deduce del **documento de la sección**
     (`usePublicOfrendas`), no del bloque (`ux.md` §3/§6.3/§10). Es la única lectura coherente con
     "la sección vive con ≥1 método **o** ≥1 texto publicado" (P7-13/R7-12): con la regla de `ux.md`,
     un introductorio publicado sin métodos dejaría la sección viva **sin entrada de navegación**.
   - **Endpoint del bloque**: **`GET /api/v1/portada/ofrendas`** (contrato, P7-7/P7-14), no la
     propuesta `GET /api/v1/portada/donaciones` de `ux.md` §4.9/§6.3.
   - **Entidad financiera**: **NO traducible** (P7-5/R7-5). `ux.md` §2.2/§4.4/§8.2 la daba por
     traducible (pestañas es/en del formulario incluían «Entidad financiera»): en la implementación,
     las pestañas de idioma de `DonacionForm` **solo** contienen el **titular**; `institution` es un
     campo no lingüístico único (como `number`/`currency`). Única lectura que garantiza SC-006.
   - **Duplicado**: `409` por **tipo + número normalizado** (P7-10), no por "mismo tipo, misma
     entidad y mismo número" (`ux.md` §2.5, literal de Assumptions). En la edición el chequeo
     **excluye la fila que se edita**: editar sin cambiar el número responde `200`.
   - **SINPE en pantalla**: se muestra en su **formato normalizado** `+506XXXXXXXX` (FR-003/SC-009
     literal, plan P7-4). La agrupación de lectura propuesta en `ux.md` D-4 (`+506 8888 8888`) queda
     como **mejora de presentación pendiente del humano** (D-4 va marcada ⟲): si se aprueba, es una
     línea en `lib/numero.ts` con el invariante de RG7-7 — **quitar los separadores devuelve
     `number`** — y la copia sigue llevando el valor exacto.
10. **Formato de los números en el panel**: la columna Número muestra el valor **normalizado**
    guardado (FR-015: lo que escribió el equipo antes de normalizar no se conserva).

## Leyenda

| Marca | Significado |
|---|---|
| `[backend]` `[frontend]` `[db]` `[infra]` | Capa y subagente responsable (`[backend]`/`[db]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` → `devops`) |
| `[P]` | **Paralelizable** con las otras tareas de su grupo (archivos distintos y sin dependencia entre ellas; ver "Grupos de paralelismo") |
| *(sin `[P]`)* | Secuencial: tiene dependencias previas que deben estar integradas |
| *(FR-…, SC-…, US…)* | Requisitos de `spec.md` que cubre la tarea |

> **Orden de capas**: migraciones (`[db]`) → contrato → `platform/` → números → repository →
> service → handler/routes → frontend (api/hooks → página → navegación/bloque/panel) → e2e → cierre.
> Cada capa se construye y prueba con lo que necesita la siguiente (el service con fakes, el handler
> con service falso), manteniendo la dirección de dependencias **`handler → service → repository`**
> de §II. La **auditoría** se escribe junto a cada mutación porque FR-017 exige que la operación y su
> registro sean atómicos. Las migraciones van **primero** (`000012` → `000013`, ninguna aplicada se
> edita) y **no hay migración de permisos** (P7-6).

## Tabla resumen

| ID | Tarea | Capa | Fase | Paralelo | Depende de |
|---|---|---|---|---|---|
| T701 | Migración `000012_create_donation_content` (`donation_methods` + `donation_texts`) | `[db]` | 1 | P1 | — |
| T702 | Migración `000013_extend_admin_actions_for_donations` (+8 códigos `donation.*`) | `[db]` | 1 | P1 | — |
| T703 | Consultas sqlc `donation_methods.sql` + `donation_texts.sql` y generar `internal/db` | `[db]` | 1 | — | T701, T702 |
| T704 | Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.7.0 → 0.8.0) | `[backend]` | 2 | P1 | — |
| T705 | Regenerar `frontend/src/api/schema.d.ts` | `[frontend]` | 2 | — | T704 |
| T706 | `platform/audit`: +8 códigos `donation.*` | `[backend]` | 3 | P1 | — |
| T707 | `donaciones`: `numero.go` (IBAN estructura + longitud por país; SINPE `+506`) | `[backend]` | 4 | P1 | — |
| T708 | `donaciones`: `model.go` + `repository.go` (pgtype, `withTx`, auditoría sobre el tx) | `[backend]` | 4 | — | T703, T706 |
| T709 | `donaciones`: `repository_donaciones.go` (CRUD + `ExistsDonationMethodDuplicate`) | `[backend]` | 4 | P2 | T708 |
| T710 | `donaciones`: `repository_textos.go` + `repository_public.go` (singletons y solo publicado) | `[backend]` | 4 | P2 | T708 |
| T711 | Integración: `000012`/`000013` up→down→up, `CHECK`s de la BD y atomicidad de la auditoría | `[backend]` | 4 | — | T709, T710 |
| T712 | `donaciones`: `service.go` — invariantes, validación y constantes | `[backend]` | 5 | — | T707, T708 |
| T713 | `donaciones`: `service_public.go` — `GetOfrendas` (sección, `?lang`, claves omitidas) | `[backend]` | 5 | — | T712 |
| T714 | `donaciones`: `service_public.go` — `GetPortadaOfrendas` (tope 5 en servidor) | `[backend]` | 5 | — | T713 |
| T715 | `donaciones`: `service_admin.go` — CRUD, textos, publicación y auditoría | `[backend]` | 5 | — | T712 |
| T716 | `donaciones`: `service_audit.go` — `audit.Recorder` (denegaciones y rechazos) | `[backend]` | 5 | — | T712 |
| T717 | `handler_public.go` — `GET /api/v1/ofrendas` | `[backend]` | 6 | — | T713, T716 |
| T718 | `handler_public.go` — `GET /api/v1/portada/ofrendas` | `[backend]` | 6 | — | T714, T717 |
| T719 | `handler_admin.go` — CRUD de `/api/v1/admin/donaciones` | `[backend]` | 6 | — | T715, T717 |
| T720 | `handler_admin.go` — textos (`GET` + `PUT …/textos/{textKey}`) | `[backend]` | 6 | — | T715, T719 |
| T721 | `routes.go` + `cmd/api/main.go` (`AdminChain("donaciones")`) + humo | `[backend]` | 6 | — | T717–T720 |
| T722 | `api/ofrendas.ts` + hooks (público, portada, panel) | `[frontend]` | 7 | P3 | T705 |
| T723 | i18n público (§7.2) + `messages.ts` del panel (§7.1) | `[frontend]` | 7 | P3 | — |
| T724 | Permisos y accesos del panel: `DONACIONES`, sello retirado, `AccesosRapidos`, `ACTION_LABELS` +8 | `[frontend]` | 7 | — | T722 |
| T725 | **Actualización acotada**: pruebas de frontend que fijan permisos/accesos (RG7-6) | `[frontend]` | 7 | — | T724 |
| T726 | `/ofrendas`: `OfrendasPage`, `DonacionCard`, `NumeroDonacion`, `CopyAccountButton` + **redirección** (P7-21) | `[frontend]` | 7 | P4 | T722, T723 |
| T727 | `useSiteLinks`: entrada «Ofrendas y diezmos» condicional (documento de la sección, P7-8) | `[frontend]` | 7 | P4 | T722, T723 |
| T728 | `OfrendasBlock` + `OfrendasItem` en `HomePage` (tras `MinisteriosBlock`) | `[frontend]` | 7 | P4 | T722, T723 |
| T729 | Panel: `DonacionesPage` + `DonacionesListsByType` + ruta `/panel/donaciones` | `[frontend]` | 7 | — | T722, T724 |
| T730 | Panel: `DonacionForm` + `validation.ts` + `errors.ts` (moneda solo IBAN) | `[frontend]` | 7 | — | T729 |
| T731 | Panel: `TextoForm` + zona de textos con estado propio (Q10) | `[frontend]` | 7 | — | T730 |
| T732 | `ofrendasSeed.ts` (fixtures del módulo, patrón `agendaSeed.ts`) | `[frontend]` | 8 | — | T729–T731 |
| T733 | E2E Playwright `ofrendas-publica.spec.ts` (US1, US4, US5; números exactos y redirección) | `[frontend]` | 8 | P5 | T726, T727, T732 |
| T734 | E2E Playwright `ofrendas-panel.spec.ts` (US2, US3; permisos y auditoría) | `[frontend]` | 8 | P5 | T729–T732 |
| T735 | E2E Playwright `ofrendas-accesibilidad.spec.ts` (teclado y foco, SC-008) | `[frontend]` | 8 | P5 | T726, T727, T728, T732 |
| T736 | E2E Playwright `portada-ofrendas.spec.ts` (US6, SC-014) | `[frontend]` | 8 | P5 | T728, T732 |
| T737 | Verificación de cierre: `make ci`, cobertura ≥ 80 %, up→down→up, e2e por tandas | `[infra]` | 9 | — | todas |
| T738 | Documentación y PR (documentador + devops) | `[infra]` | 9 | — | T737 |

**Total: 38 tareas** — 17 `[backend]` · 16 `[frontend]` · 3 `[db]` · 2 `[infra]` · 16 marcadas `[P]`.

## Grupos de paralelismo

| Grupo | Tareas | Por qué pueden ir en paralelo |
|---|---|---|
| P1 | T701 · T702 · T704 · T706 · T707 | Cinco artefactos independientes entre sí: dos migraciones (`000012`/`000013`, archivos distintos — el orden numérico solo importa al aplicar con `make db-migrate` —), el contrato (`backend/api/openapi.yaml`), `platform/audit` y las funciones puras de `numero.go`; nadie comparte archivo ni depende de otro |
| P2 | T709 · T710 | `repository_donaciones.go` y `repository_textos.go`/`repository_public.go` son archivos distintos que parten de T708 |
| P3 | T722 · T723 | El cliente `api/ofrendas.ts` + hooks y el i18n/`messages.ts` no se solapan (T722 depende solo de T705; T723 no depende de nada) |
| P4 | T726 · T727 · T728 | `features/publico/ofrendas/` (página y ficha), `hooks/useSiteLinks.ts` y `OfrendasBlock`/`OfrendasItem` + `HomePage` son archivos distintos que parten de T722 **y de T723** (claves i18n tipadas) |
| P5 | T733 · T734 · T735 · T736 | Cuatro specs e2e independientes (`ofrendas-publica`, `ofrendas-panel`, `ofrendas-accesibilidad`, `portada-ofrendas`) sobre la misma base (`ofrendasSeed`, T732) |

---

## Fase 1 — Migraciones y capa de datos (`[db]`)

> Convenciones de la skill `postgres-db` y de `data-model.md` (fuente de verdad): `id UUID`,
> `created_at`/`updated_at`, índices y `CHECK`/`UNIQUE` **en la base**, `up`/`down` completos, SQL
> **parametrizado** (§IV, CWE-89) y **nunca** se edita una migración aplicada (§VI). Numeración sin
> huecos desde `000001` (F6 terminó en `000011`). **Sin migración de permisos** (P7-6: `donaciones`
> ya está sembrado por `000002` y el catálogo sigue en 10).

- [X] T701 · Migración `000012_create_donation_content` · `[db]` `[P]` (P1)

  - **Archivos**: `backend/migrations/000012_create_donation_content.up.sql` y
    `000012_create_donation_content.down.sql` (NUEVOS).
  - **Qué hace**: crea **`donation_methods`** y **`donation_texts`** tal como `data-model.md`:
    `donation_methods` con `kind` (`'iban'`/`'sinpe'`), `number` (guardado **normalizado**),
    `institution` obligatoria **no traducible** (R7-5), `holder_es`/`holder_en` opcionales
    (traducibles), `currency` (`'CRC'`/`'USD'`) **solo en IBAN**, `publication_state`
    (`'draft'`/`'published'`, default `'draft'`) y `sort_order INTEGER NOT NULL DEFAULT 0`
    (**dentro del grupo de tipo**, Q1); `donation_texts` con `text_key UNIQUE CHECK (text_key IN
    ('intro','usage'))`, `body_es`/`body_en` (≤2.000) y su **propio** `publication_state` (Q10), sin
    `sort_order` y sin filas sembradas. **CHECKs**: formato por tipo (`iban` →
    `'^[A-Z]{2}[0-9]{2}[A-Z0-9]{10,30}$'`, `sinpe` → `'^\+506[0-9]{8}$'`); moneda obligatoria en IBAN
    y prohibida en SINPE; español por campo (`*_en IS NULL OR *_es IS NOT NULL`); longitudes
    (institution/holder 160, body 2.000, number 1–34) y `btrim` en `institution`/`holder_es`/
    `holder_en` (normalización de F2/P13, precedente `000010`). **Índices**:
    `donation_methods_kind_sort_order_id_idx (kind, sort_order, id)` y
    `donation_methods_publication_state_idx (publication_state)`. El `down` tira las dos tablas.
    *Cubre FR-001, FR-004, FR-012, FR-019, FR-021 y el modelo de FR-007/FR-014/FR-015/FR-016 a nivel
    de datos.*
  - **Pruebas incluidas** (§III): — (SQL; su verificación ejecutable es T711). **Cómo se verifica**:
    `make db-migrate` aplica sin errores sobre la BD de F6 (`000011` → `000012`); `migrate … down`
    revierte sin errores; los `CHECK`/`DEFAULT`/índices de `data-model.md` están literalmente en el
    SQL; ninguna tabla sin `id`/`created_at`/`updated_at`; `donation_texts` sin contenido sembrado.
  - **Criterio de terminado**: el esquema coincide con `data-model.md` (lo revisa `revisor-codigo`
    contra la skill `postgres-db`) y los invariantes de Q1/Q2/Q10 están garantizados por la BD.
  - **Commit sugerido**: `feat(db): tablas de ofrendas con numeros normalizados y textos por clave (000012)`

- [X] T702 · Migración `000013_extend_admin_actions_for_donations` · `[db]` `[P]` (P1)

  - **Archivos**: `backend/migrations/000013_extend_admin_actions_for_donations.up.sql` y
    `.down.sql` (NUEVOS).
  - **Qué hace**: amplía el registro de auditoría de F2 (FR-017, R7-10): sustituye
    `admin_actions_action_check` (nombre propio desde `000006`, verificado en F3/F4/F6) añadiendo
    los **8 códigos** `donation.create`, `donation.update`, `donation.delete`, `donation.publish`,
    `donation.unpublish`, `donation.text.update`, `donation.text.publish` y
    `donation.text.unpublish`. El `target_kind` **no cambia** (los métodos y los textos son
    `content`). El `down` elimina las filas con códigos F7 y restaura el CHECK de F6 (destructivo a
    propósito, igual que los `down` de `000006`/`000008`/`000011`; documentado en `data-model.md`).
    *Cubre FR-017* a nivel de datos.
  - **Pruebas incluidas** (§III): — (SQL; su verificación ejecutable es T711). **Cómo se verifica**:
    `make db-migrate` aplica; los 8 códigos del `CHECK` coinciden uno a uno con la lista de
    `data-model.md`; el registro de acciones queda **cerrado** en la BD.
  - **Criterio de terminado**: registro cerrado garantizado por la BD, sin tocar ninguna migración
    anterior (§VI).
  - **Commit sugerido**: `feat(db): ampliar admin_actions a donaciones (000013)`

- [X] T703 · Consultas sqlc `donation_methods.sql` + `donation_texts.sql` y generar · `[db]`

  - **Archivos**: `backend/internal/db/queries/donation_methods.sql` (NUEVO),
    `backend/internal/db/queries/donation_texts.sql` (NUEVO) y `backend/internal/db/*.go`
    (GENERADOS y commiteados: `make sqlc`). *No se toca `queries/audit.sql` (se reutiliza
    `InsertAdminAction`, R7-10) ni `queries/media.sql` (F7 no tiene imágenes, P7-16).*
  - **Qué hace**: escribe las consultas nombradas de `data-model.md`:
    `ListPublishedDonationMethods(:limit)` —**única vía pública** de métodos: solo
    `publication_state = 'published'`, `ORDER BY CASE kind WHEN 'iban' THEN 0 ELSE 1 END, sort_order,
    id`, `LIMIT :limit` (el límite lo pasa el service: `maxPublicItems = 200` en la sección y `5` en
    el bloque, FR-022)—; `ListDonationMethods(:limit, :offset, sqlc.narg('publication_state'))` y
    `CountDonationMethods` (panel paginado, **todo**, **mismo orden que ve el visitante**,
    US2 esc. 10); `GetDonationMethodByID`; `InsertDonationMethod`/`UpdateDonationMethod`/
    `DeleteDonationMethod` (el `Update` escribe **la fila completa ya fusionada** que le pasa el
    service — R7-15) y `ExistsDonationMethodDuplicate(:kind, :number, sqlc.narg('exclude_id'))`
    (mismo tipo y mismo número **normalizado**, R7-9; con `AND (@exclude_id::uuid IS NULL OR id <>
    @exclude_id)` para **excluir la fila que se edita**: `NULL` en el alta, su `id` en la edición).
    En `donation_texts.sql`: `ListPublishedDonationTexts` (solo publicados,
    `ORDER BY text_key`), `ListDonationTexts` (panel, los dos renglones posibles en ambos idiomas),
    `GetDonationTextByKey(:text_key)` y `UpsertDonationText` (`INSERT … ON CONFLICT (text_key) DO
    UPDATE …` + `updated_at = now()`, patrón del singleton de F3). Sin `SELECT *`, `ORDER BY`
    determinista (§8.1.7), todo parametrizado (§IV, CWE-89); **ninguna escritura sobre
    `admin_actions`**. *Cubre FR-002, FR-004, FR-012, FR-021 y FR-022* a nivel de acceso a datos.
  - **Pruebas incluidas** (§III): — (SQL anotado; su ejecución es T709–T711). **Cómo se verifica**:
    `make sqlc` genera sin errores y `make sqlc-verify` no detecta deriva; los tipos generados
    quedan commiteados; ninguna consulta de escritura sobre `admin_actions`; el `ORDER BY` es
    literalmente `CASE kind …, sort_order, id` en público y en panel.
  - **Criterio de terminado**: la capa generada compila y es la única fuente de SQL del dominio
    `donaciones` (T708–T710).
  - **Commit sugerido**: `feat(db): consultas sqlc de metodos y textos de donaciones`

---

## Fase 2 — Contrato OpenAPI y tipos

- [X] T704 · Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.8.0) · `[backend]` `[P]` (P1)

  - **Archivos**: `backend/api/openapi.yaml` (EDITADO). *No* se toca
    `specs/007-donaciones/contracts/openapi.yaml` (snapshot inmutable, P7-19).
  - **Qué hace**: funde **aditivamente** el delta aprobado: las **9** operaciones de F7
    (`GET /api/v1/ofrendas?lang=`, `GET /api/v1/portada/ofrendas?lang=`, las 4 de panel
    `GET/POST /api/v1/admin/donaciones` y `GET/PATCH/DELETE …/{id}`, y las 2 de textos
    `GET /api/v1/admin/donaciones/textos` y `PUT …/textos/{textKey}`) y sus schemas
    (`OfrendasPublicas`, `MetodoDonacionPublico`, `TextoPublico`, `PortadaOfrendas`,
    `DonacionAdminList`, `DonacionAdmin`, `DonacionInput`, `DonacionPatch`, `DonacionTextosAdmin`,
    `DonacionTextoAdmin`, `DonacionTextoInput`, `MetodoKind`, `Moneda`, `NumeroNormalizado`,
    parámetros `DonacionItemId`/`DonacionTextKey`), reutilizando `ErrorEnvelope`/`ErrorBody`,
    `PublicationState`, las respuestas reutilizables, los parámetros `Limit`/`Offset` **reales** del
    contrato vivo (no inventar `ListLimit`/`ListOffset`) y `sessionCookie`. Aplica los cambios
    documentados del delta: enum `action` de `AdminActionItem` **+ 8 códigos** `donation.*`; la
    prosa del **catálogo de permisos NO cambia** (sigue en 10; `donaciones` ya estaba sembrado — a
    diferencia de F4 —); **no se toca** `GET /api/v1/media/{fileName}` (F7 no tiene imágenes,
    P7-16). Documenta el tope de la sección (`maxPublicItems = 200`), el **tope 5 del bloque
    resuelto en servidor** (sin parámetro de límite), que las claves `items`/`intro`/`usage` se
    **omiten** sin contenido publicado (la redirección de `/ofrendas` es del frontend: la operación
    sigue siendo `200`) y que `number` viaja **normalizado**. Sube `info.version` de **0.7.0 a
    0.8.0**. El registro de `error.code` **no cambia** (F7 emite `invalid`, `unauthenticated`,
    `forbidden`, `not_found`, `conflict`, `internal`). *Cubre el contrato de todos los FR*; garantiza
    FR-011 (el permiso `donaciones` citado en las descripciones de panel), FR-012 (ninguna operación
    pública devuelve borradores), FR-015/FR-016 (la validación de números documentada) y FR-017
    (los 8 códigos).
  - **Pruebas incluidas** (§III): — (artefacto de contrato). **Cómo se verifica**: el YAML parsea
    sin errores y **sin `$ref` sin resolver** (validador de `make api-gen` o búsqueda de `$ref`
    huérfanos — el simple «el YAML parsea» no lo detecta —); `/api/v1/ofrendas` y
    `/api/v1/portada/ofrendas` tienen `security: []`, `Cache-Control: no-store` documentado y los
    schemas públicos **sin** `publicationState`; `MetodoDonacionPublico` **sin** `currency` cuando
    `kind: sinpe` (documentado) y `holder` opcional; las operaciones de panel citan `sessionCookie`
    y el permiso `donaciones`; `grep -n "version:" backend/api/openapi.yaml` → `0.8.0`; `make api-gen`
    completa sin errores.
  - **Criterio de terminado**: el contrato vivo contiene F1–F4 + F6 + F7 y es la única fuente de
    tipos del frontend (T705).
  - **Commit sugerido**: `docs(api): fusionar delta de F7 en el contrato vivo (0.8.0)`

- [X] T705 · Regenerar `frontend/src/api/schema.d.ts` · `[frontend]`

  - **Archivos**: `frontend/src/api/schema.d.ts` (GENERADO y commiteado).
  - **Qué hace**: ejecuta `make api-gen` (o `npm run api:gen`) contra **`backend/api/openapi.yaml`**
    (nunca contra el snapshot de `specs/`, P7-19) y commitea los tipos de F7 (`OfrendasPublicas`,
    `MetodoDonacionPublico`, `TextoPublico`, `PortadaOfrendas`, `DonacionAdmin`/`DonacionInput`/
    `DonacionPatch`, `DonacionTextoAdmin`/`DonacionTextoInput`, `PublicationState`,
    `AdminActionItem` con los enums ampliados a los 8 códigos `donation.*`).
  - **Pruebas incluidas** (§III): — (artefacto generado). **Cómo se verifica**: regenerar no produce
    `git diff`; `npm run typecheck` pasa con los tipos nuevos.
  - **Criterio de terminado**: `schema.d.ts` reproducible desde el contrato y listo para
    `api/ofrendas.ts` (T722; regla anti-deriva de §8.1.5).
  - **Commit sugerido**: `build(frontend): regenerar schema.d.ts con los tipos de F7`

---

## Fase 3 — Núcleo de plataforma (`backend/internal/platform/`)

- [X] T706 · `platform/audit`: 8 códigos `donation.*` · `[backend]` `[P]` (P1)

  - **Archivos**: `backend/internal/platform/audit/audit.go` y `audit_test.go` (EDITADOS).
  - **Qué hace**: amplía el registro cerrado del plumbing de F2 (R7-10) con los **8 códigos**
    `donation.create|update|delete|publish|unpublish|text.update|text.publish|text.unpublish` en
    `ActionCodes` y `ValidActionCode`. `TargetKindContent = "content"` **no cambia** (ya lo añadió
    F3). *Cubre FR-017*.
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/platform/audit/` en verde: cada código `donation.*` pasa `Validate` con
    objetivo `content` y `TargetLabel`; los códigos de F2/F3/F4/F6 siguen intactos (regresión); un
    código fuera de la lista falla.
  - **Criterio de terminado**: la lista cerrada coincide literalmente con el `CHECK` de la migración
    `000013` (T702) — lo comprueba la integración de T711.
  - **Commit sugerido**: `feat(audit): codigos donation.* para F7`

---

## Fase 4 — Dominio `donaciones`: números y repository

- [X] T707 · `donaciones`: `numero.go` — normalización y validación de IBAN y SINPE · `[backend]` `[P]` (P1)

  - **Archivos**: `backend/internal/donaciones/numero.go` y `numero_test.go` (NUEVOS).
  - **Qué hace**: funciones **puras** del dominio (R7-2/R7-3, P7-2/P7-3/P7-4):
    `Normalizar/ValidarIban` — quita espacios y guiones, pasa a mayúsculas, valida **estructura**
    (`^[A-Z]{2}[0-9]{2}[A-Z0-9]+$`) y **longitud exacta por país** con la tabla del registro IBAN
    embebida (`CR` = 22, `ES` = 24, `DE` = 22, `GB` = 22 …, **sin mod-97** — default revisable,
    pregunta #3 al humano —) y devuelve el número **normalizado a los caracteres del número** (el
    dato normalizado es el que se guarda: FR-015); `Normalizar/ValidarSinpe` — acepta con o sin
    `+506` y con espacios/guiones, exige **8 dígitos con primer dígito 6/7/8** (celular, no línea
    fija — default revisable, RG7-3 —) y devuelve `+506XXXXXXXX` (FR-016). Sin librerías nuevas
    (R7-16). *Cubre FR-015, FR-016 y SC-009* (capa de dominio).
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/donaciones/` en verde. **IBAN aceptados** (escritos con espacios/guiones y
    en minúsculas, que se normalizan; el valor devuelto es el normalizado y **no cambia** al
    guardarlo): `CR05 0152-0200-1234-5678-90` → `CR05015202001234567890` (**22** = `CR` + 20
    dígitos), `ES91 2100 0418 4502 0005 1332` → `ES9121000418450200051332` (ES = 24),
    `DE89 3704 0044 0532 0130 00` → `DE89370400440532013000` (DE = 22),
    `GB29 NWBK 6016 1331 9268 19` → `GB29NWBK60161331926819` (GB = 22). **IBAN rechazados**, cada
    uno por la razón que se indica: `CR05 0152 0200 1234 5678 9` (**21**: le falta uno a `CR`, que
    exige 22) y `CR05 0152 0200 1234 5678 901` (**23**: le sobra uno) — los dos casos explícitos
    "CR de 21 y de 23" —, `CR05 0152-0200-1234-5678-90-1` (sufijo tras un guion: normaliza a 23),
    `ZZ05 0152 0200 1234 5678 90` (22 pero `ZZ` está **fuera del registro** IBAN),
    `CRAB 0152 0200 1234 5678 90` (letras donde van los dos dígitos de control: fuera de la
    estructura), `CR05 0152-0200-1234-5678-9😀` (emoji fuera de la estructura), `hola` y la cadena
    vacía (ni estructura ni longitud). **SINPE aceptados**: `88888888`, `8888-8888`, `+506 8888 8888`,
    `+50688888888` → todos `+50688888888`; **rechazados** 7/9 dígitos, `22222222` (fija), letras,
    `+507…`, sufijo extra (SC-009).
  - **Criterio de terminado**: las reglas de número tienen **una** implementación pura y sus casos
    de prueba; sin imports fuera de la stdlib; `golangci-lint` sin comentarios.
  - **Commit sugerido**: `feat(donaciones): validacion y normalizacion de IBAN y SINPE Movil`

- [X] T708 · `donaciones`: `model.go` + `repository.go` · `[backend]`

  - **Archivos**: `backend/internal/donaciones/model.go`, `repository.go`, `repository_test.go`
    (NUEVOS).
  - **Qué hace**: entidades del dominio (`MetodoDonacion` con `Kind`, `Number`, `Institution`,
    `HolderEs`/`HolderEn`, `Currency`, `SortOrder`, `PublicationState`; `TextoDonacion` con
    `TextKey`, `BodyEs`/`BodyEn`, `PublicationState`), DTOs con etiquetas `validate` y sufijos
    `Es`/`En` solo donde hay traducción (R3-1/R7-5: `institution` **no** tiene par de idiomas) y los
    **DTOs públicos** (`MetodoDonacionPublico`: `currency` solo en IBAN y `holder` solo cuando
    existe — el mapeo público se compone en `service_public.go`, un solo lugar —). Base del
    repository: constructor sobre `*pgxpool.Pool`, mapeos `pgtype` → dominio (§8.1.6), `withTx`
    (patrón de F2) y la **inserción de `admin_actions` dentro del `tx`** vía `InsertAdminAction`
    (mapeo `audit.Action` → `InsertAdminActionParams`; sin que un dominio importe a otro, R2).
    *Soporta FR-007 (patrón bilingüe) y FR-017 (registro atómico)*.
  - **Pruebas incluidas** (§III): unitarias de los mapeos (`pgtype` ↔ dominio, nil/`Valid`) y del
    mapeo de acciones (los 8 códigos `donation.*`, incluidos `text.update`/`text.publish`/
    `text.unpublish`). **Cómo se verifica**: `go test ./internal/donaciones/` en verde;
    `gofmt`/`go vet` limpios; ningún `pgtype` sale de `repository.go` (R4).
  - **Criterio de terminado**: el paquete compila con sus tipos y el tx de escritura queda listo para
    T709/T710.
  - **Commit sugerido**: `feat(donaciones): modelo y base del repository con auditoria transaccional`

- [X] T709 · `donaciones`: `repository_donaciones.go` (CRUD y duplicados) · `[backend]` `[P]` (P2)

  - **Archivos**: `backend/internal/donaciones/repository_donaciones.go` y
    `repository_donaciones_integration_test.go` (NUEVOS; `//go:build integration`).
  - **Qué hace**: `InsertDonationMethod`/`UpdateDonationMethod`/`DeleteDonationMethod`/
    `GetDonationMethodByID`, `ListDonationMethods`/`CountDonationMethods` (panel paginado, `ORDER BY
    CASE kind …, sort_order, id` — el mismo orden que ve el visitante, P7-20 —, filtro opcional
    `publication_state`) y `ExistsDonationMethodDuplicate(:kind, :number, :exclude_id)` (mismo
    `kind` + `number` **normalizado**, R7-9; `:exclude_id` **opcional** — `NULL` en el alta — y la
    consulta lleva `AND (@exclude_id::uuid IS NULL OR id <> @exclude_id)` para **excluir la fila
    que se edita**). Traducción de restricciones a `apperr`: duplicado → `conflict`, violaciones de `CHECK`
    (formato por tipo, moneda por tipo, `*En` sin `*Es`, `btrim`) → `invalid` con `details`. Borrado
    físico (sin `deleted_at`, default revisable). *Cubre FR-010, FR-014, FR-015, FR-016, FR-019*.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/donaciones/` en verde: un IBAN con espacios
    (`'CR05 0152…'`) o en minúsculas falla en la BD y el normalizado procede; un SINPE sin `+506`
    falla y `'+50688888888'` procede; una cuenta `iban` **sin** `currency` falla y una `sinpe`
    **con** `currency` falla; `holder_en` sin `holder_es` falla; `institution`/`holder_*` con
    espacios sobrantes fallan (`btrim`); `ExistsDonationMethodDuplicate` detecta `'cr05 0152-…'` vs
    `'CR050152…'` (tras normalizar) y **no** dispara por el mismo número en otro tipo; con
    `:exclude_id` = el `id` de la propia fila devuelve `false` (**"editar un método sin cambiar el
    número no devuelve `409`"**) y con `:exclude_id` de otra fila devuelve `true` cuando el número
    coincide con el de **otro** método (el `PATCH` que cambia el número al de otro método sí debe
    chocar); el `Update` deja exactamente la fila completa que recibe del service (ya **fusionada**:
    campos ausentes intactos, `null` explícito borrando lo opcional — R7-15).
  - **Criterio de terminado**: persistencia real de métodos con sus invariantes protegidos por la BD
    y traducidos a errores del dominio.
  - **Commit sugerido**: `feat(donaciones): persistencia de metodos con duplicados por tipo y numero`

- [X] T710 · `donaciones`: `repository_textos.go` + `repository_public.go` · `[backend]` `[P]` (P2)

  - **Archivos**: `backend/internal/donaciones/repository_textos.go`, `repository_public.go`,
    `repository_textos_integration_test.go` y `repository_public_integration_test.go` (NUEVOS;
    `//go:build integration`).
  - **Qué hace**: `ListDonationTexts`/`GetDonationTextByKey`/`UpsertDonationText` (crea o reemplaza
    el singleton por clave `intro`/`usage`; sin borrado — P7-17) y la **única vía pública**:
    `ListPublishedDonationMethods(:limit)` y `ListPublishedDonationTexts` (solo `publication_state =
    'published'`; los textos viajan empaquetados por `text_key`). *Cubre FR-002, FR-012, FR-021 y
    FR-022* (capa de consulta).
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real** (las más importantes del
    módulo junto a T711). **Cómo se verifica**:
    `go test -tags=integration ./internal/donaciones/` en verde: `ListPublished*` **nunca** devuelve
    borradores —ni métodos ni textos (SC-002)—; el orden es `CASE kind …, sort_order, id` (iban
    primero) y el `LIMIT` se respeta (5 en el bloque, 200 en la sección); `UpsertDonationText`
    actualiza en vez de duplicar (`UNIQUE (text_key)`); `text_key` fuera del catálogo (`'otro'`)
    falla; `body_en` con `body_es` presente procede.
  - **Criterio de terminado**: la regla "0 borradores expuestos" está garantizada **en la consulta**,
    no solo en el servicio (FR-012: "el sistema NUNCA debe exponer al público un borrador").
  - **Commit sugerido**: `feat(donaciones): textos por clave y lectura publica sin borradores`

- [X] T711 · Integración: migraciones up→down→up, `CHECK`s de la BD y atomicidad · `[backend]`

  - **Archivos**: `backend/internal/donaciones/repository_migrations_integration_test.go` y
    `repository_audit_integration_test.go` (NUEVOS; `//go:build integration`).
  - **Qué hace**: la verificación ejecutable de `data-model.md` §"Validación del modelo":
    (1) las migraciones `000012` y `000013` aplican **up → down → up** sin errores (restauración del
    `CHECK` de `admin_actions`); (2) el catálogo de `permissions` **sigue en 10** filas con
    `donaciones` entre ellas (F7 no lo toca; el `down` no lo roza); (3) los `CHECK`s de formato por
    tipo, moneda por tipo, español por campo, `btrim`, `text_key` y `UNIQUE (text_key)` se disparan
    en la BD; (4) **mutación + `admin_actions` atómicas**: al forzar un fallo en el registro, la
    mutación **no se aplica**, y los 8 códigos del `CHECK` de la BD coinciden con `platform/audit`
    (T706) y caben en el registro (y un código fuera del registro se rechaza); (5) la inserción de
    `admin_actions` con `target_kind='content'`, `target_label` correcto y máximo 254. *Cubre
    FR-014, FR-015, FR-016, FR-017 y SC-011.*
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/donaciones/` en verde con los cinco bloques.
  - **Criterio de terminado**: las migraciones son reversibles y la atomicidad de la auditoría queda
    verificada en la base (edge case "ninguna operación se aplica sin registro").
  - **Commit sugerido**: `test(donaciones): integracion de migraciones y auditoria atomica`

---

## Fase 5 — Dominio `donaciones`: servicios

- [X] T712 · `donaciones`: `service.go` — invariantes, validación y constantes · `[backend]`

  - **Archivos**: `backend/internal/donaciones/service.go` y `service_test.go` (NUEVOS).
  - **Qué hace**: interfaz `Repository` (definida por quien la consume, R3), constantes revisables
    (institution/holder **160**, textos **2.000**, número por su formato — máx. 60 en el input —,
    **`maxPublicItems = 200`** y **`maxPortadaItems = 5`** — tope fijo de la spec, resuelto en
    servidor: los aplican T713/T714 y los cita el contrato —), normalización (trim; `""`/espacios en
    `*En` y opcionales → `NULL`, I6) y la **validación de dominio** (FR-014): obligatorios **solo**
    `kind`, `number`, `institution` (y `currency` si es `iban`; Q2), **titular opcional** (un campo
    vacío no obliga a nada), **moneda obligatoria en IBAN y rechazada en SINPE**, **español por
    campo** (`holderEn` exige `holderEs`; `textEn` exige `textEs`), números validados y normalizados
    con `numero.go` (T707) **antes de persistir**, límites por campo y `lang` ∈ `es|en`. Todo
    contenido se trata como datos y nunca como código (§IV). *Soporta FR-014, FR-015, FR-016,
    FR-019 y FR-021*.
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/donaciones/` en verde: cada campo inválido → `apperr.Invalid` con `details`
    por campo (`kind`/`number`/`institution` vacíos; IBAN mal formado → `details.number`; SINPE no
    celular → `details.number`; `currency` faltante en IBAN → `details.currency`; `currency` presente
    en SINPE → `details.currency`; `holderEn` sin `holderEs` → `details.holderEs`; textos de 2.001
    caracteres → `details.textEs`/`details.textEn`); `""`/`"  "` → `NULL`; tildes/`ñ`/`¿¡`/emojis
    intactos en textos y **fuera de lugar en números** (rechazados); `maxPublicItems` es **200** y
    `maxPortadaItems` es **5** (fijados por prueba).
  - **Criterio de terminado**: las reglas de validación tienen **una** implementación y sus pruebas;
    los límites son el único punto de verdad (los usan T713–T715 y los cita el contrato).
  - **Commit sugerido**: `feat(donaciones): invariantes y validacion de dominio`

- [X] T713 · `donaciones`: `service_public.go` — `GetOfrendas` (sección) · `[backend]`

  - **Archivos**: `backend/internal/donaciones/service_public.go` y `service_public_test.go`
    (NUEVOS).
  - **Qué hace**: `GetOfrendas(lang)` (FR-002…FR-006): devuelve `{lang, items?, intro?, usage?}` con
    **solo** contenido `published` (métodos con tope `maxPublicItems = 200` y textos por clave) en
    orden `CASE kind …, sort_order, id` (iban primero, FR-004/Q1), cada campo traducible (`holder`,
    `text`) resuelto con fallback **`en` si existe, si no `es`** (FR-008; el cliente **no**
    reimplementa el fallback) y **cada clave omitida cuando no hay nada publicado de esa clave**
    (SC-010/US3 esc. 7: sin huecos; nunca `null`). `kind`, `number`, `institution` y `currency` no
    cambian con el idioma (FR-007/SC-006). `lang` fuera de `es|en` → `apperr.Invalid`. El DTO público
    **nunca** contiene `publicationState`; `currency` **ausente** en los `sinpe` y `holder` **ausente**
    sin titular (el frontend decide con este documento: redirección, navegación y grupos). *Cubre
    FR-002, FR-003, FR-004, FR-005, FR-007, FR-008, FR-012, FR-021* (capa de servicio).
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos; **las más importantes del
    módulo**). **Cómo se verifica**: `go test ./internal/donaciones/` en verde: solo `published`
    **y por texto** (un texto en borrador no sale aunque haya métodos — US3 esc. 7); orden
    `iban → sinpe` y `sortOrder, id` dentro de cada grupo; fallback por campo (nunca hueco); con 0
    métodos y 0 textos la respuesta es **solo `{lang}`** (todas las claves omitidas — P7-13: la
    redirección la decide el frontend —); con solo textos publicados **sí** sale `intro`/`usage`
    (la sección vive: pregunta #2 fijada por prueba); con > 200 publicados salen **200** ya
    ordenados; los DTOs públicos **no** traen `publicationState`, `currency` en SINPE ni `holder`
    vacío; `lang=fr` → `invalid`.
  - **Criterio de terminado**: SC-002 (0 borradores) y SC-010 (claves omitidas, sin huecos)
    verificables en el service; cobertura medida (`go test -cover`, ≥ 80 % al cierre).
  - **Commit sugerido**: `feat(donaciones): seccion publica con textos, grupos y fallback`

- [X] T714 · `donaciones`: `service_public.go` — `GetPortadaOfrendas` (tope 5) · `[backend]`

  - **Archivos**: `backend/internal/donaciones/service_public.go` (EDITADO) y
    `service_public_test.go` (AMPLIADO).
  - **Qué hace**: `GetPortadaOfrendas(lang)` (FR-022/US6/SC-014): mismo predicado "publicado" y
    mismo fallback `en → es` que la sección, con **tope `maxPortadaItems = 5` resuelto en el
    servidor** (`LIMIT 5`; el contrato **no** expone parámetro de límite) y el **mismo orden de la
    sección** (`CASE kind …, sort_order, id`): los 5 **primeros**. Devuelve `PortadaOfrendas` con el
    **mismo esquema de ítem** `MetodoDonacionPublico` que la sección (FR-022: "el bloque es un
    extracto de los mismos métodos, no un modelo nuevo"; la tarjeta compacta recorta campos en la
    composición, no en el contrato). Sin métodos publicados → **`items` omitida** (el bloque no se
    dibuja — **sin mensaje**, diferencia consciente con el bloque de F4). **Este bloque solo mira
    métodos**: con solo textos publicados la respuesta es solo `{lang}`. *Cubre FR-022 y SC-014*.
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos). **Cómo se verifica**:
    `go test ./internal/donaciones/` en verde: con 8 publicados salen **exactamente los 5 primeros**
    en el orden de la sección (iban → sinpe, `sortOrder, id`; los otros 3 no aparecen); con menos,
    exactos (sin huecos); con 0 métodos publicados → `items` **omitida** (aunque haya textos
    publicados); 0 borradores; fallback `en → es`; `lang` inválido → `invalid`; el tope **5** queda
    fijado por prueba (constante: no crece con el contenido).
  - **Criterio de terminado**: US6/SC-014 resueltos a nivel de dominio y cobertura de
    `service*.go` ≥ 80 % intacta.
  - **Commit sugerido**: `feat(donaciones): extracto de la portada con tope 5 en el servidor`

- [X] T715 · `donaciones`: `service_admin.go` — CRUD, textos, publicación y auditoría · `[backend]`

  - **Archivos**: `backend/internal/donaciones/service_admin.go` y `service_admin_test.go` (NUEVOS).
  - **Qué hace**: `CreateDonacion`/`UpdateDonacion`/`DeleteDonacion`/`ListDonaciones`/
    `GetDonacion` y `GetTextos`/`PutTexto` (FR-010/FR-021): validación completa (T712) con `details`
    por campo y **sin guardar nada a medias**; **duplicado exacto → `409 conflict`** (P7-10) con
    `ExistsDonationMethodDuplicate(:kind, :number, :exclude_id)` **excluyendo la fila que se
    edita** (una edición que no cambia el número **no** devuelve `409` contra sí misma);
    normalización del número antes de persistir (IBAN a los caracteres del número; SINPE
    `+506XXXXXXXX`); `publication_state` por elemento **y por texto** (Q10) y **"se publica al
    guardar"** (FR-013: la edición de algo publicado es visible en la primera carga); `sortOrder`
    editable **dentro del grupo** (US2 esc. 10); panel paginado (`limit`/`offset`,
    `platform/paginate`, filtro `publicationState`) en el orden del visitante; los textos se guardan
    con `UpsertDonationText` (crea o reemplaza el singleton, **sin borrado** — P7-17 — y con su
    estado propio); **registro transaccional** con la regla exacta de códigos (R7-10): el alta deja
    **solo `donation.create`** (una sola fila, también con `publicationState: published`);
    `donation.publish`/`donation.unpublish` **solo** para cambios de estado de un método;
    `donation.text.update` para el guardado de un texto (alta o edición) y
    `donation.text.publish`/`unpublish` **solo** para cambios de estado de un texto; `donation.update`
    para cualquier otro cambio de método (incluido `sortOrder`); en la edición, datos + estado →
    **dos filas**; **fail-closed** (si el registro falla no se aplica nada). El `UpdateDonacion`
    **fusiona los campos presentes** sobre la fila actual y escribe la fila completa ya fusionada
    (R7-15). *Cubre US2, US3 y FR-010, FR-012, FR-013, FR-014, FR-017, FR-019, FR-021*.
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos). **Cómo se verifica**:
    `go test ./internal/donaciones/` en verde: cada campo inválido → `400` con `details` y **sin
    guardar nada**; IBAN con espacios → se guarda el normalizado; SINPE sin `+506` → normalizado;
    duplicado → `conflict` (alta repetida); **"editar un método sin cambiar el número no da
    `409`"** → `200` (el chequeo excluye la fila que se edita; también con `PATCH` que solo toca
    `institution`/`sortOrder`/`publicationState`) y un `PATCH` que cambia el número al de **otro**
    método → `conflict`; moneda obligatoria en IBAN y rechazada en SINPE; `holderEn` sin
    `holderEs` → `invalid`; crear con `publicationState: published` deja **solo**
    `donation.create` (**cero** filas `donation.publish`); una edición que cambia datos y estado deja
    **dos filas**; reordenar deja `donation.update`; `PUT` de texto deja `donation.text.update` y su
    cambio de estado `donation.text.publish`/`unpublish`; si el registro falla no se guarda;
    `""`/espacios en `*En` → `NULL`; tildes/`ñ`/`¿¡`/emojis intactos; eliminar deja su fila
    `donation.delete`. **Casos de fusión del `PATCH`**: actualizar **un solo campo** → el resto
    queda **intacto**; `null` explícito → **borrado opcional** (p. ej. `holderEs: null`);
    `number` en el `PATCH` se revalida sobre el conjunto fusionado **sin contar la fila propia**
    como duplicado (R7-9).
  - **Criterio de terminado**: US2/US3 y FR-010/FR-012/FR-013/FR-017 cubiertos en el service.
  - **Commit sugerido**: `feat(donaciones): gestion del panel con textos, publicacion y auditoria`

- [X] T716 · `donaciones`: `service_audit.go` — `audit.Recorder` · `[backend]`

  - **Archivos**: `backend/internal/donaciones/service_audit.go` y `service_audit_test.go` (NUEVOS).
  - **Qué hace**: implementa `audit.Recorder` para el dominio (patrón de F4/F6, R7-10):
    `RecordDenied` resuelve **método+ruta → acción** de las rutas del módulo
    (`/api/v1/admin/donaciones…` → `donation.*`; los textos → `donation.text.*`) y la persiste con
    `result='denied'` best-effort; el helper de rechazo hace lo propio con `result='failure'`. Nunca
    transporta el cuerpo de la petición (FR-026 de F2). `target_label`: `Donaciones · <institution> ·
    <number>`, `Donaciones · Texto introductorio`, `Donaciones · Texto de uso` (máx. 254). *Cubre
    FR-017 (y el patrón de denegación de FR-011)*.
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/donaciones/` en verde: cada ruta del módulo resuelve a su acción y
    `targetLabel`; una ruta desconocida no deja fila; un fallo de registro no cambia la respuesta
    (best-effort).
  - **Criterio de terminado**: las denegaciones sobre las rutas de F7 quedan registradas igual que las
    de F2/F3/F4/F6 (SC-011).
  - **Commit sugerido**: `feat(donaciones): registro de denegaciones y rechazos del modulo`

---

## Fase 6 — Handlers, rutas y cableado

- [X] T717 · `handler_public.go` — `GET /api/v1/ofrendas` · `[backend]`

  - **Archivos**: `backend/internal/donaciones/handler.go`, `handler_public.go`,
    `handler_public_test.go` (NUEVOS).
  - **Qué hace**: publica la operación del contrato: valida `lang` (`es|en`, por defecto `es`),
    llama a `GetOfrendas` (T713) y responde `200` con `OfrendasPublicas` y **`Cache-Control:
    no-store`** (SC-003/FR-013). Sin autenticación (FR-002). Helpers HTTP del dominio
    (decodificar/validar/responder con `apperr` + `WriteError`, patrón de F2/F4). El cuerpo JSON
    **nunca** contiene `publicationState` ni borradores; las claves `items`/`intro`/`usage` se
    **omiten** sin contenido. *Cubre FR-002, FR-005, FR-008, FR-012* (vía de salida pública).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/donaciones/` en verde: `200` con el sobre correcto y `no-store`; `lang=fr`
    → `400 invalid` con `details.lang`; claves omitidas cuando el service no las devuelve (0
    publicados → solo `{"lang":…}`, que es lo que dispara la redirección del frontend); el cuerpo
    JSON serializado **no contiene `publicationState`** (aserción sobre el JSON crudo); error del
    service → sobre `ErrorEnvelope` genérico.
  - **Criterio de terminado**: `quickstart.md` §1 es ejecutable tal cual por API (US1).
  - **Commit sugerido**: `feat(donaciones): endpoint publico de ofrendas y diezmos`

- [X] T718 · `handler_public.go` — `GET /api/v1/portada/ofrendas` · `[backend]`

  - **Archivos**: `backend/internal/donaciones/handler_public.go` (EDITADO) y
    `handler_public_test.go` (AMPLIADO).
  - **Qué hace**: publica la operación del contrato (P7-14): valida `lang`, llama a
    `GetPortadaOfrendas` (T714) y responde `200` con `PortadaOfrendas` y **`Cache-Control:
    no-store`**; sin parámetro de límite (el **tope 5 es del servidor**); `items` **omitida** con 0
    métodos publicados (nunca `null`: el bloque no se dibuja, FR-022). El DTO **nunca** contiene
    `publicationState`. La ruta va en `RegisterPublic` **sin middlewares** (pública, solo lectura) —
    anotando que el segmento `portada` es el agregado de F3 pero el recurso lo sirve `donaciones`
    (Complexity 4 del plan). *Cubre FR-022*.
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/donaciones/` en verde: `200` con el sobre correcto y `no-store`; `items`
    omitida con 0 métodos; `lang=fr` → `400 invalid`; error del service → sobre genérico.
  - **Criterio de terminado**: `quickstart.md` §14 es ejecutable tal cual por API (US6).
  - **Commit sugerido**: `feat(donaciones): endpoint publico del bloque de ofrendas de la portada`

- [X] T719 · `handler_admin.go` — CRUD de `/api/v1/admin/donaciones` · `[backend]`

  - **Archivos**: `backend/internal/donaciones/handler_admin.go` y `handler_admin_test.go`
    (NUEVOS).
  - **Qué hace**: `GET/POST /api/v1/admin/donaciones` y `GET/PATCH/DELETE
    /api/v1/admin/donaciones/{id}` sobre el service de T715: listado paginado (`limit`/`offset` con
    `platform/paginate`, filtro opcional `publicationState`), `201`/`200`/`204` según el contrato,
    `400` con `details` por campo, `404` por id inexistente y `409` en duplicado exacto (**nunca**
    contra la propia fila: editar un método sin cambiar su número responde `200`); el `PATCH`
    **fusiona los campos presentes** sobre la fila actual (última escritura completa gana, R7-15):
    los campos ausentes no cambian y `null` explícito borra lo opcional. El input acepta el número
    **tal como lo escribe el equipo** (espacios/guiones; SINPE con o sin `+506`) y el service lo
    normaliza y valida (FR-015/FR-016). Las respuestas del panel (`DonacionAdmin`) incluyen los dos
    idiomas del titular y el estado. *Cubre FR-010, FR-014, FR-015, FR-016, FR-019* (vía de panel).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/donaciones/` en verde: cada operación con sus códigos
    200/201/204/400/404/409 y sobres uniformes; `details` por campo en los `400` (incluido
    `details.number` para IBAN/SINPE mal formados); **casos del `PATCH`**: petición con **un solo
    campo** → el resto **intacto** y `null` explícito → **borrado opcional**; el `403` de permiso se
    prueba en T721 (cada grupo con su `AdminChain`).
  - **Criterio de terminado**: `quickstart.md` §3/§4/§5/§6 ejecutables vía API (US2).
  - **Commit sugerido**: `feat(donaciones): endpoints del panel de donaciones`

- [X] T720 · `handler_admin.go` — textos (`GET` + `PUT …/textos/{textKey}`) · `[backend]`

  - **Archivos**: `backend/internal/donaciones/handler_admin.go` (EDITADO) y
    `handler_admin_test.go` (AMPLIADO).
  - **Qué hace**: `GET /api/v1/admin/donaciones/textos` (los dos singletons, `null` hasta el primer
    guardado, patrón F3) y `PUT /api/v1/admin/donaciones/textos/{textKey}` (crea o reemplaza;
    `textEs` obligatorio, `textEn` opcional con español base por campo, `publicationState` por
    defecto `draft`; `textKey` fuera de `intro|usage` → `400 invalid`). Texto **plano** de uno o
    varios párrafos, ≤2.000 caracteres por idioma, guardado íntegro sin interpretar (§IV; tildes,
    `ñ`, `¿¡` y emojis intactos). **Sin operación de borrado** (P7-17: un texto se retira en
    `draft`). *Cubre FR-010, FR-021 y Q5/Q10/Q11*.
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/donaciones/` en verde: `GET` devuelve `{intro: …|null, usage: …|null}`;
    `PUT` crea y luego reemplaza; `textEn` sin `textEs` → `400` con `details.textEs`; 2.001
    caracteres → `400`; `textKey` inválido → `400`; respuestas con los dos idiomas y el estado;
    `401`/`403` según el contrato.
  - **Criterio de terminado**: `quickstart.md` §7 ejecutable vía API (US2, FR-021).
  - **Commit sugerido**: `feat(donaciones): endpoints de los textos de la seccion`

- [X] T721 · `routes.go` + `cmd/api/main.go` (`AdminChain("donaciones")`) + humo · `[backend]`

  - **Archivos**: `backend/internal/donaciones/routes.go` y `routes_test.go` (NUEVOS),
    `backend/cmd/api/main.go` y `main_test.go` (EDITADOS: DI del dominio `donaciones` + registro de
    `/api/v1/ofrendas`, `/api/v1/portada/ofrendas` y `/api/v1/admin/donaciones`).
  - **Qué hace**: cablea el grupo de panel con la cadena aprobada de F2 —
    `authn → guard de cambio de contraseña → authz("donaciones") → CSRF` — **sin helpers nuevos de
    middleware**: `AdminChain("donaciones")` con el permiso **ya sembrado** (P7-6/FR-011: verificado
    **en servidor**; independiente de los demás permisos — Decisión 5 —). Las 2 rutas públicas
    quedan sin middlewares (solo lectura, sin rate-limit: no hay superficie pública escribible) y
    `/healthz` intacto (§8.1.9). *Cubre FR-011 (el servidor es la única autoridad) y FR-017* (el
    `Recorder` de T716 recibe cada cadena).
  - **Pruebas incluidas** (§III): humo de composición + de permisos. **Cómo se verifica**:
    `go test ./cmd/api/` y `go test ./internal/donaciones/` en verde: rutas nuevas con su sobre;
    `/healthz` sin sesión y `404` `not_found` en rutas inexistentes (§8.1.9); una cuenta con
    `donaciones` pasa y una cuenta con `eventos`/`ministerios`/`portada` recibe **`403`** en
    `/api/v1/admin/donaciones` (y al revés: una cuenta solo `donaciones` no accede a otros módulos,
    US2 esc. 9); el `403` genérico queda registrado como `denied`; `go test ./...` completo en verde.
  - **Criterio de terminado**: el permiso queda verificado en servidor y auditable de un vistazo en
    `routes.go` (SC-005).
  - **Commit sugerido**: `feat(donaciones): rutas con AdminChain("donaciones")`

---

## Fase 7 — Frontend

> Se **aplica** el sistema de F10 y los patrones de F2/F3/F4/F6 (no se redefinen). Nombres de pieza y
> composición según `ux.md` §6 (y §4 para pantallas/estados); i18n solo en el sitio público (FR-018:
> el panel va en español). Ningún `fetch` fuera del cliente de API (regla F2); sin imágenes
> (Q6/FR-020: 0 `img` y 0 `ImageUploader` en F7). **`HiddenNotice` no se usa** (F7 redirige, P7-21).

- [X] T722 · `api/ofrendas.ts` + hooks · `[frontend]` `[P]` (P3)

  - **Archivos**: `frontend/src/api/ofrendas.ts` (NUEVO),
    `frontend/src/features/publico/ofrendas/hooks/usePublicOfrendas.ts`,
    `hooks/usePortadaOfrendas.ts`, `frontend/src/features/donaciones/hooks/useDonaciones.ts`,
    `useCreateDonacion.ts`, `useUpdateDonacion.ts`, `useDeleteDonacion.ts`,
    `useDonacionTextos.ts`, `usePutDonacionTexto.ts` (NUEVOS; queryKeys en `hooks/queryKeys.ts`).
  - **Qué hace**: cliente tipado desde `schema.d.ts` (T705): `getOfrendas(lang)` y
    `getPortadaOfrendas(lang)` (públicos) y panel (list/create/update/delete de métodos + getTextos/
    putTexto); hooks TanStack Query (`usePublicOfrendas` con key `['ofrendas', lang]`,
    `usePortadaOfrendas` con `['portada-ofrendas', lang]` — `placeholderData: keepPreviousData`
    y retención de la última respuesta válida: sin parpadeos al cambiar de idioma — y los del panel
    con invalidación tras mutaciones para que el cambio sea visible en la primera carga, SC-003). El
    cliente **no filtra ni reimplementa** el fallback ni la regla de borradores: el servidor ya
    resolvió todo. *Soporta US1–US4 y US6*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: cada hook contra un handler MSW (éxito, `400` con `details`,
    `401`, `403`, `404`, `409`); respuesta solo con `lang` (claves omitidas); respuesta del bloque
    sin `items`; error con datos previos retenidos; invalidación de caché tras crear/editar/
    eliminar/guardar un texto.
  - **Criterio de terminado**: ningún componente hace `fetch` directo (regla F2) y los tipos salen
    solo del contrato vivo (0 `any`).
  - **Commit sugerido**: `feat(frontend): cliente y hooks de ofrendas`

- [X] T723 · i18n público y textos del panel · `[frontend]` `[P]` (P3)

  - **Archivos**: `frontend/src/features/publico/i18n/{messages,es,en}.ts` (EDITADOS: claves
    tipadas de `ux.md` §7.2 — `nav.sections.offerings`, `offerings.page.*`, `offerings.iban.heading`,
    `offerings.sinpe.heading`, `offerings.holder.label`, `offerings.currency.*`, `offerings.copy`,
    `offerings.copied`, `offerings.copy.hint`, `offerings.usage.heading`, `offerings.loading`,
    `offerings.error.*`, `a11y.offerings.*`, `home.offerings.*`, `a11y.home.offerings.heading` —) y
    `frontend/src/features/donaciones/messages.ts` (NUEVO: catálogo del panel de `ux.md` §7.1).
    Sin la clave, TypeScript **no compila** (garantía de SC-006).
  - **Qué hace**: completa el diccionario bilingüe del sitio público y los textos en español del
    panel (FR-018: el panel no pasa por `t()`). Traducciones fieles a `ux.md` §7 (incluidos los
    mensajes de error de §2.5: IBAN no válido, SINPE no válido, moneda sin elegir, inglés sin
    español, límite de los textos, duplicado, sección sin publicados, error del sistema "sin perder
    lo escrito"). **La redirección no añade cadena** (no hay aviso que mostrar). *Cubre FR-006
    (etiqueta de navegación), FR-008 y FR-018*.
  - **Pruebas incluidas** (§III): Vitest (tabla de casos). **Cómo se verifica**:
    `npm test -- --run` en verde: las claves `es`/`en` son simétricas (sin huecos); los textos del
    panel son cadenas en español (0 claves de i18n en `messages.ts` del panel).
  - **Criterio de terminado**: i18n completo del sitio público (SC-006) y catálogo de textos del
    panel listo para T726–T731.
  - **Commit sugerido**: `feat(frontend): textos de ofrendas (i18n publico y panel)`

- [X] T724 · Permisos, accesos del panel y `ACTION_LABELS` · `[frontend]`

  - **Archivos**: `frontend/src/lib/permissions.ts` (EDITADO: `DONACIONES = 'donaciones'` y su
    entrada en `MODULE_PERMISSIONS`), `frontend/src/features/roles/permissions.ts` (EDITADO: el sello
    «Disponible más adelante» deja de aplicarse a `donaciones` — el catálogo se lee dinámico de la
    API y **no** cambia —), `frontend/src/features/panel/components/AccesosRapidos.tsx` y
    `frontend/src/features/panel/pages/InicioPage.tsx` (EDITADOS: tarjeta **«Gestión de donaciones»**
    **solo** con el permiso, `ux.md` §4.2), `frontend/src/app/layout.tsx` (EDITADO: entrada del menú
    del panel filtrada por permiso) y `frontend/src/features/auditoria/messages.ts` (EDITADO:
    `ACTION_LABELS` cubre los **8 códigos** `donation.*`, `ux.md` §4.8/§7.1: «Creó/Modificó/Eliminó
    la cuenta "…"», «el número de SINPE "…"», «Publicó/Retiró …», «Modificó/Creó el texto
    introductorio», «el texto de uso»…). *El registro de rutas del SPA va con las páginas (T726
    pública, T729 panel) para que cada commit compile (nota 7 de la alineación).*
  - **Qué hace**: activa en el frontend el permiso existente `donaciones` (FR-011: **sin permisos
    nuevos** y **sin renombrarlo**, Q8): el formulario de roles de F2 lo ofrece (sin sello), el menú
    y los accesos rápidos se filtran por permiso (una cuenta solo `donaciones` no ve «Eventos» ni
    «Ministerios») y la auditoría muestra etiquetas legibles para los 8 códigos (FR-017). *Cubre
    FR-011 (visibilidad del acceso en el panel), FR-017 (etiquetas) y FR-018.*
  - **Pruebas incluidas** (§III): Vitest + Testing Library (las del propio cambio; las que fijan
    listas de permisos van en T725). **Cómo se verifica**: `npm test -- --run` en verde (con T725
    integrada): `isPermissionAvailable('donaciones') === true` y el sello fuera; el menú renderiza la
    entrada solo con el permiso; la tarjeta de `AccesosRapidos` se muestra solo con `donaciones` (y
    ninguna con la cuenta sin permisos de F7); los **8 códigos** `donation.*` tienen etiqueta humana
    en `ACTION_LABELS`.
  - **Criterio de terminado**: el acceso «Gestión de donaciones» es visible **solo** con el permiso
    `donaciones` y la auditoría no muestra códigos crudos de F7 (`ux.md` §4.8).
  - **Commit sugerido**: `feat(frontend): permiso donaciones, accesos del panel y etiquetas de auditoria`

- [X] T725 · Actualización **acotada** de las pruebas que fijan permisos/accesos · `[frontend]` *(cambio acotado documentado, RG7-6)*

  - **Archivos** (SOLO los que fijen la lista de permisos/accesos; al menos los 3 de RG7-6):
    `frontend/src/features/roles/permissions.test.ts`,
    `frontend/src/features/roles/roles.test.tsx`, `frontend/src/features/panel/panel.test.tsx` y, si
    fija `MODULE_PERMISSIONS`, `frontend/src/lib/permissions.test.ts` (EDITADOS). *Nada más se toca.*
  - **⚠️ CAMBIO ACOTADO Y DOCUMENTADO (RG7-6)**: F7 edita pruebas de F2/F4/F6 (funcionalidades
    cerradas) porque el catálogo de accesos del frontend activa `donaciones`. Es un cambio de
    producto **autorizado por el plan**, no un debilitamiento: solo cambian **recuentos, listas y
    sellos** (la aserción `isPermissionAvailable('donaciones') === false` pasa a `true`, el sello
    «Disponible más adelante» deja de aplicarse y la tarjeta nueva de accesos rápidos se cuenta);
    ninguna aserción de comportamiento se relaja. El backend **no** se toca (el catálogo de permisos
    de F2 sigue en 10). La tarea queda **registrada como excepción** en el PR y la validan
    `revisor-codigo` y `seguridad`.
  - **Qué hace**: ajusta las aserciones que fijan la lista de permisos y accesos para incluir
    `donaciones`. *Cubre FR-011.*
  - **Pruebas incluidas** (§III): son las propias pruebas actualizadas. **Cómo se verifica**:
    `npm test -- --run` en verde; el diff toca **solo** recuentos/listas/comentarios de esos
    archivos (se revisa el diff: ninguna aserción de comportamiento cambia); el PR lo registra como
    **cambio acotado** (no debilitamiento).
  - **Criterio de terminado**: `make ci` pasa con `donaciones` activo en el catálogo de accesos y la
    excepción queda documentada en la evidencia.
  - **Commit sugerido**: `test(frontend): permisos y accesos con donaciones (cambio acotado RG7-6)`

- [X] T726 · `/ofrendas`: `OfrendasPage`, `DonacionCard`, `NumeroDonacion`, `CopyAccountButton` y **redirección** · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/ofrendas/OfrendasPage.tsx`, `IntroText.tsx`,
    `UsageText.tsx`, `DonacionesGroup.tsx`, `DonacionCard.tsx`, `NumeroDonacion.tsx`,
    `CopyAccountButton.tsx`, `format.ts` (presentación de los números) y `*.test.tsx` (NUEVOS),
    `frontend/src/app/router.tsx` (EDITADO: ruta pública **`/ofrendas`** dentro de `PublicLayout`).
  - **Qué hace**: la página con H1 «Ofrendas y diezmos» + subtítulo fijo (D-2, cadena de interfaz),
    **texto introductorio** solo si viene (`IntroText`, columna de lectura ~65ch, saltos
    preservados), grupo **«Cuentas IBAN»** (`DonacionesGroup` con `SectionHeading` kicker 01 +
    `RevealGroup`) y grupo **«Números de SINPE Móvil»** (kicker 02) en el orden del API
    (`iban → sinpe`, `sortOrder, id`; el cliente **no** reordena, FR-004) y **texto de uso** solo si
    viene (`UsageText` con el rótulo de interfaz D-3). **Cada elemento solo existe si tiene dato**
    (SC-010): sin titular no hay línea; en SINPE no hay línea de moneda; un grupo vacío no dibuja su
    rótulo; un texto en borrador no deja hueco. `DonacionCard` (`Card solid`, **sin zona de imagen**,
    Q6): entidad (h3), `NumeroDonacion` — **IBAN agrupado en bloques de 4** solo como presentación
    (`format.ts`, RG7-7: quitar los separadores devuelve `number`), **SINPE en su formato
    normalizado** (alineación §9; D-4 pendiente del humano) y `aria-label` con el número completo —,
    `CopyAccountButton` (icono **+ texto** «Copiar número», ≥44 px; copia al portapapeles el número
    **exacto normalizado**, confirmación `role="status"` «Número copiado.» y ayuda de transcripción
    siempre visible — D-5 —), moneda **solo en IBAN** («Moneda: colones»/«dólares», D-6) y titular
    **solo si existe**. **Redirección silenciosa (P7-21, RG7-12)**: `SectionSkeleton` mientras
    `usePublicOfrendas` está en vuelo y **solo con éxito y sin contenido** (sin `items`, `intro` ni
    `usage`) → `<Navigate to="/" replace />` —**sin aviso**, sin pintar jamás la grilla vacía ni
    `HiddenNotice`—; **ante error NO se redirige** (H1 + «No pudimos cargar la información de
    ofrendas…» + «Volver a intentar»; con caché, contenido + aviso discreto — patrón de F3/`ux.md`
    §4.1). Estados de `ux.md` §4.1 y `prefers-reduced-motion` respetado. *Cubre US1, US4, US5 y
    FR-001, FR-002, FR-003, FR-005, FR-006, FR-008, FR-021, FR-023, FR-024 (UI).*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: dos grupos en el orden del API y encabezados correctos; IBAN
    **agrupado en bloques** en pantalla y el valor del portapapeles = `number` exacto (SC-009);
    SINPE en su formato normalizado; **sin titular no se pinta línea** y **sin moneda en SINPE**
    (0 huecos, SC-010/US1 esc. 5); texto en borrador → sin hueco (US3 esc. 7); fallback `en → es`
    visible (nunca campo vacío) y **número/tipo/entidad/moneda idénticos entre idiomas** (SC-006);
    **redirección sin parpadeo**: con la consulta en caché redirige en el primer render sin pintar H1
    ni fichas; sin caché solo se ve el esqueleto mientras se decide; `<Navigate replace>`; y **ante
    error NO redirige** (se muestra el estado de error); texto largo con varios párrafos y emojis
    íntegro.
  - **Criterio de terminado**: `quickstart.md` §1/§4/§5/§7/§9/§10 ejecutables en navegador
    (US1/US4) y la redirección silenciosa sin destello verificada (RG7-12).
  - **Commit sugerido**: `feat(frontend): pagina /ofrendas con grupos, copia exacta y redireccion sin publicados`

- [X] T727 · `useSiteLinks`: entrada «Ofrendas y diezmos» condicional · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/hooks/useSiteLinks.ts` y su test (EDITADOS).
  - **Qué hace**: `SiteLinkId` crece con **`'offerings'`** (alineación §9.8) y la entrada única
    **«Ofrendas y diezmos»** con href **`/ofrendas`** existe **solo si** `usePublicOfrendas` trae
    `items`, `intro` **o** `usage` (P7-8: la presencia bebe del **documento de la sección**, no del
    bloque — ver alineación §9 —; la consulta se comparte por `queryClient` con la página, 0
    peticiones extra). **Depende de T722 y T723**: usa la clave i18n tipada
    **`nav.sections.offerings`** que crea T723. La consumen a la vez `PublicNav`, `MobileMenu` y
    `FooterSiteMap` (grupo **Explorar**, después de «Ministerios»: D-1; fuente única: barra, menú
    hamburguesa y pie **a la vez**, sin huérfanos ni enlaces rotos, FR-005). Sin nada publicado, el
    ítem **no existe** en ninguna de las tres superficies. *Cubre FR-005, FR-006 y el edge case
    "navegación sin enlaces rotos"* (SC-010).
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: con `items` el ítem aparece en las tres superficies con
    `href="/ofrendas"`; **con solo `intro`/`usage` publicados también aparece** (P7-13: la sección
    vive con solo textos — pregunta #2 fijada por prueba —); con el documento solo `lang`, en
    ninguna; el resto de enlaces (anclas de portada, `/eventos`, `/ministerios`) no cambia.
  - **Criterio de terminado**: ocultación coherente del ítem (FR-005) verificada en las tres
    superficies (SC-010) con la fuente de P7-8.
  - **Commit sugerido**: `feat(frontend): navegacion con /ofrendas condicional en barra, menu y pie`

- [X] T728 · `OfrendasBlock` + `OfrendasItem` en la portada · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/ofrendas/OfrendasBlock.tsx`, `OfrendasItem.tsx` y
    `*.test.tsx` (NUEVOS), `frontend/src/features/publico/pages/HomePage.tsx` (EDITADO: **una
    inserción** de `<OfrendasBlock />` **inmediatamente después de `<MinisteriosBlock />`** y antes
    de la siguiente sección de F3) y `homePage.test.tsx` (AMPLIADO).
  - **Qué hace**: bloque «Ofrendas y diezmos» con el patrón de los bloques de F4/F6 (P7-14):
    `SectionHeading` + fila/grilla de **hasta 5 `OfrendasItem`** en el orden de la sección —tarjeta
    compacta de texto **sin imagen** (Q6) y **sin botón de copiar ni enlace por ficha** (D-14: la
    copia vive en `/ofrendas`): chip de tipo («Cuenta IBAN»/«SINPE Móvil», D-12 — necesario porque
    en la portada no hay grupos rotulados —), entidad (h3, **no** enlace), número en presentación de
    lectura **sin elipsis nunca** (si no cabe, la ficha crece; RG7-7), moneda solo en IBAN y titular
    solo si existe— y **botón «Ver formas de ofrendar»** a `/ofrendas` (solo con ítems; ≥44 px). El
    bloque se alimenta de `usePortadaOfrendas` (`GET /api/v1/portada/ofrendas`, tope 5 del servidor;
    el cliente no recorta). **Sin métodos publicados el bloque devuelve `null`: NO se renderiza** —
    sin título, sin mensaje, sin hueco— (FR-022: **diferencia consciente** con `ProximamenteBlock`,
    que sí muestra aviso). Esqueleto propio de altura acotada mientras carga; **error sin caché → no
    se dibuja**; error con caché → contenido + aviso discreto. Sistema de F10 (`Card`,
    `SectionHeading`, `Reveal` con la dirección de F6 para no romper la alternancia — D-15 —),
    responsive desde 320 px (1 col móvil, 2 tableta, 5 escritorio; **sin carrusel ni scroll
    lateral**), foco visible, `section` con `aria-labelledby` al h2, ítems como `li` con `h3`.
    **La portada de F3 no se toca** más allá de esta inserción (FR-023). *Cubre US6, FR-022, FR-023,
    FR-024.*
  - **Pruebas incluidas** (§III): Vitest + Testing Library. **Cómo se verifica**:
    `npm test -- --run` en verde: **5 exactos** con 8 publicados (los 5 primeros del orden de la
    sección); exactos con menos, **sin huecos ni marcadores**; **con 0 métodos publicados el bloque no
    se renderiza** (sin título ni mensaje) y `homePage.test.tsx` confirma que **el bloque de
    ministerios y el resto de la portada se renderizan igual que antes** y que el bloque queda
    **inmediatamente después de `MinisteriosBlock`**; el botón **ausente** sin ítems; cada ficha
    muestra su chip de tipo y **sin botón de copiar**; en inglés traducido con fallback; número
    nunca recortado (SC-009).
  - **Criterio de terminado**: US6/FR-022/SC-014 en pantalla, en la posición acordada (tras el
    bloque de ministerios), y `quickstart.md` §14 ejecutable en navegador.
  - **Commit sugerido**: `feat(frontend): bloque de ofrendas y diezmos en la portada`

- [X] T729 · Panel: `DonacionesPage` + `DonacionesListsByType` + ruta `/panel/donaciones` · `[frontend]`

  - **Archivos**: `frontend/src/features/donaciones/pages/DonacionesPage.tsx`,
    `components/DonacionesListsByType.tsx` y tests (NUEVOS), `frontend/src/app/router.tsx` (EDITADO:
    **`/panel/donaciones`** con `RequirePermission(DONACIONES)`).
  - **Qué hace**: gestión en el panel **en español** (FR-018) con la composición de `ux.md` §4.3/§5:
    título «Gestión de donaciones» + banda de ayuda (D-7); **zona «Textos de la sección»** arriba
    (dos tarjetas —introductorio y uso— con fragmento del contenido, `StatusPill` de **estado
    propio** (Q10) y botón Editar; un texto **nunca escrito** muestra «Todavía sin escribir.» con su
    acción — D-8, no es un error —) y debajo el **listado de métodos agrupado por tipo** (Q1):
    bloque **«Cuentas IBAN»** con botón «Agregar cuenta» y bloque **«Números de SINPE Móvil»** con
    «Agregar número»; cada bloque es una `Table` (desktop) / tarjetas (móvil) con columnas **Orden**
    (el valor numérico dentro del grupo, D-10), **Número** (normalizado, con `NumeroDisplay` y su
    copia exacta — D-9 —), **Entidad**, **Titular** («—» si vacío: aquí sí se usa «—», D-16),
    **Moneda** (solo la tabla IBAN: la columna **no existe** en SINPE, Q2), **Estado** (`StatusPill`)
    y **Acciones** (`Editar` · `Publicar`/`Retirar de la sección` · `Eliminar`). `ConfirmDialog` suave
    para retirar y danger para eliminar («se eliminará para siempre… Esto no se puede deshacer.»);
    estados de `ux.md` §5 (cargando `EmptyState` por bloque, vacío con botón por bloque,
    **«Nada está publicado ahora: el público no ve esta sección ni el bloque de la portada, y
    `/ofrendas` redirige a la portada…»** cuando todo está en borrador, error con «Reintentar»,
    éxito con `Notice` de §7.1) y sin permiso la ruta cae en `/sin-permiso` de F2. El listado va
    ordenado como el del visitante (iban primero; `sortOrder, id` en cada grupo). *Cubre US2, US3 y
    FR-010, FR-012 (estado visible), FR-018.*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: los dos bloques con sus columnas (sin columna Moneda en SINPE);
    Orden/Estado visibles; publicar/retirar/eliminar con sus confirmaciones y mensajes (§7.1); el
    estado vacío por bloque y el aviso «Nada está publicado ahora…»; «Todavía sin escribir.» en la
    zona de textos; sin permiso `donaciones` la ruta no es accesible (texto de `/sin-permiso`
    intacto).
  - **Criterio de terminado**: `quickstart.md` §2/§8/§13 ejecutables en el panel (US3).
  - **Commit sugerido**: `feat(frontend): panel de donaciones con listados agrupados y textos`

- [X] T730 · Panel: `DonacionForm` + `validation.ts` + `errors.ts` · `[frontend]`

  - **Archivos**: `frontend/src/features/donaciones/components/DonacionForm.tsx`,
    `NumeroDisplay.tsx`, `validation.ts`, `errors.ts` y tests (NUEVOS).
  - **Qué hace**: alta/edición en `Dialog` con el patrón F3/F4/F6 (`ux.md` §4.4): pestañas de idioma
    **Español** / **English (opcional)** que contienen **solo el titular** (D-9 de la alineación §9:
    `institution` **no es traducible** y `number`/`kind`/`currency` no aparecen en las pestañas);
    **Entidad financiera** (`Field`, obligatorio, 160, ayuda §7.1) y **Titular** (opcional) como
    campos no lingüísticos/según pestaña; **Número** (`Field`, obligatorio, etiqueta dinámica por
    tipo — el **tipo es etiqueta fija** al editar y viene fijado por el bloque de origen al crear
    (D-9): sin selector —, ayuda «Se guardará sin espacios ni guiones…» para IBAN y «acepta +506 o
    no» para SINPE); **Moneda** (`Select` «Colones»/«Dólares», **obligatoria y solo en el formulario
    de IBAN**; en SINPE **el campo no existe**, Q2); **Orden** (`Field` numérico, default 0, ayuda
    §7.1; sin drag & drop); **`PublishControls`** con el estado del elemento. `validation.ts` espejo
    de las reglas del server (obligatorios, IBAN por estructura/longitud con espacios/guiones
    tolerados al escribir, SINPE 8 dígitos ± `+506` con primer dígito 6/7/8, moneda por tipo,
    español por campo, límites) y `errors.ts` traduce `details` por campo; errores junto al campo
    con resumen arriba y foco al primero (FR-014); `Esc` no descarta lo escrito sin aviso (D-18 de
    F4); el **duplicado (`409`)** se avisa sin perder lo escrito y sin guardar; error del sistema →
    mensaje genérico **sin perder lo escrito** (§2.5). *Cubre US2, FR-010, FR-014, FR-015, FR-016,
    FR-019 (UI).*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: moneda obligatoria en IBAN y **campo ausente** en SINPE (Q2);
    IBAN con espacios aceptado por el espejo Zod y enviado tal cual (la normalización es del
    servidor); IBAN mal formado y SINPE no celular rechazados con el error del campo (SC-009);
    `holderEn` sin `holderEs` → error que indica qué falta; titular vacío guarda sin problemas
    (US2 esc. 8); límites con sus contadores; duplicado (`409`) mostrado sin perder lo escrito; el
    formulario de edición **no ofrece selector de tipo** (D-9) y el guardado conserva el tipo;
    `Currency` primera opción no vale como vacía.
  - **Criterio de terminado**: US2 ejecutable en < 2 min según `quickstart.md` §3 (SC-004) con las
    reglas de Q2/Q3/Q4 reflejadas en la interfaz.
  - **Commit sugerido**: `feat(frontend): formulario de metodo de donacion con moneda por tipo`

- [X] T731 · Panel: `TextoForm` + zona de textos con estado propio · `[frontend]`

  - **Archivos**: `frontend/src/features/donaciones/components/TextoForm.tsx` y tests (NUEVOS);
    `DonacionesPage.tsx` (EDITADO: la zona de textos de `ux.md` §4.5/§5.a queda completa).
  - **Qué hace**: un solo componente parametrizado para los dos textos («texto introductorio» /
    «texto de uso de las donaciones») en `Dialog`: pestañas **Español** (obligatorio si se guarda
    algo) / **English (opcional)** con la ayuda de F3; **`TextArea`** grande con contador vivo
    `{n}/2.000` (default revisable), saltos preservados, tildes/`ñ`/`¿¡`/emojis íntegros y ayuda
    «Uno o varios párrafos. Aparecerá tal como lo escribas…»; **el estado NO se cambia en el
    formulario**: se cambia con `PublishControls` en su tarjeta (D-11: «editar» y «publicar» son dos
    decisiones separadas, igual que para un método); guardar con el `PUT` del contrato (crea o
    reemplaza el singleton) y `Notice` según el estado: «Cambios guardados: ya visibles en la
    sección.» (publicado) o «Cambios guardados.» (borrador). Sin operación de borrado (P7-17).
    *Cubre US2, US3 (esc. 6–7), FR-010, FR-012, FR-021 y Q5/Q10/Q11.*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: el mismo formulario sirve para `intro` y `usage`; `textEn` sin
    `textEs` bloquea el guardado indicando qué falta; 2.001 caracteres lo indica con el límite y lo
    que falta recortar; el guardado **no** cambia el estado del texto (Q10: se publica desde su
    tarjeta); un texto sin escribir se puede crear desde «Todavía sin escribir.»; emojis y saltos
    intactos de ida y vuelta.
  - **Criterio de terminado**: `quickstart.md` §7 ejecutable en el panel y US3 cubierta para los dos
    textos con su estado propio.
  - **Commit sugerido**: `feat(frontend): editor de los textos introductorio y de uso`

---

## Fase 8 — E2E

> Toda la Fase 8 exige el backend completo (T721) integrado y la base de F7 migrada. Las specs se
> ejecutan con `make e2e` **por tandas** respetando el **límite de intentos de login** de F2 (las
> specs de panel hacen login; ver T737).

- [X] T732 · `ofrendasSeed.ts` (fixtures del módulo) · `[frontend]`

  - **Archivos**: `frontend/e2e/ofrendasSeed.ts` (NUEVO; patrón de `frontend/e2e/agendaSeed.ts` /
    `ministeriosSeed.ts`).
  - **Qué hace**: fixtures reutilizables por las 4 specs e2e (patrón de F4/F6): métodos publicados en
    orden conocido (una cuenta IBAN costarricense **con titular y moneda**, una IBAN
    **internacional** larga de 24+ caracteres — el caso de 320 px y de «nunca elipsis», SC-009 —,
    una cuenta IBAN **sin titular** (US1 esc. 5), números SINPE con y sin titular — **nunca con
    moneda** (Q2) —, textos introductorio y de uso publicados con tildes/`ñ`/`¿¡`/emojis y varios
    párrafos), borradores de método **y de texto** (SC-002), **8 publicados** para el tope 5
    (SC-014), un escenario **solo textos publicados y 0 métodos** (P7-13: la sección vive y el bloque
    no — pregunta #2 —), el escenario **0 publicados** para la redirección (SC-010) y cuentas de
    panel: una con `donaciones`, una con otros permisos de módulo y una sin permisos (SC-005).
    *Soporta US1–US6.*
  - **Pruebas incluidas** (§III): — (fixture e2e). **Cómo se verifica**: las 4 specs la consumen sin
    declarar datos propios; los datos se crean/limpian de forma determinista entre tandas.
  - **Criterio de terminado**: fixtures listas y documentadas (qué caso cubre cada una).
  - **Commit sugerido**: `test(e2e): fixtures de ofrendas`

- [X] T733 · E2E Playwright `ofrendas-publica.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/ofrendas-publica.spec.ts` (NUEVO; usa `helpers.ts` y
    `ofrendasSeed.ts`).
  - **Qué hace**: recorrido del visitante (US1, US4, US5) según `quickstart.md`
    §1/§4/§5/§7/§9/§10/§11: sin sesión ve los dos grupos en el orden exacto (`iban` → `sinpe`,
    `sortOrder, id`) con el texto introductorio arriba y el de uso abajo; los números **íntegros**
    (IBAN en bloques en pantalla; SINPE `+506…`), entidad, **moneda solo en IBAN** y **titular solo
    cuando existe** (0 líneas vacías, SC-010); **botón de copiar**: el portapapeles recibe el
    **número normalizado** y **quitar los separadores de la presentación devuelve ese valor**
    (SC-009, RG7-7); borradores invisibles —ni método ni texto— (SC-002); **sin nada publicado la
    visita directa a `/ofrendas` redirige a la portada** — URL `/`, contenido de la portada, **sin
    destello** de sección vacía ni aviso en ningún instante, y «Atrás» sin bucles por el `replace`
    (RG7-12)— y no hay entrada en nav/pie; con **solo textos publicados** la sección **sí** existe y
    el bloque de portada no (P7-13); inglés con fallback (contenido sin inglés se muestra en español)
    y vuelta a español (**número, tipo, entidad y moneda idénticos**: SC-006); 3 anchos
    (320/768/1280, SC-007) con el IBAN internacional largo legible sin desplazamiento horizontal.
    *Cubre SC-001, SC-002, SC-006, SC-007, SC-009, SC-010, SC-013.*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde sobre la
    base con las fixtures de T732; el resto de e2e (F3/F4/F6) sigue pasando sin tocar (y `/eventos`
    conserva su `HiddenNotice`).
  - **Criterio de terminado**: US1/US4/US5 y la exactitud de los números (SC-009) cubiertas de
    extremo a extremo.
  - **Commit sugerido**: `test(e2e): seccion publica de ofrendas (numeros exactos y redireccion)`

- [X] T734 · E2E Playwright `ofrendas-panel.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/ofrendas-panel.spec.ts` (NUEVO).
  - **Qué hace**: recorrido del panel (US2, US3) según `quickstart.md` §2–§8/§12/§13: login con
    permiso `donaciones`, **crear+publicar un método en < 2 min** (SC-004), IBAN con espacios
    guardado normalizado (SC-003), **IBAN mal formado rechazado sin guardar** (SC-009), SINPE con y
    sin `+506` y **SINPE inválido rechazado** (línea fija `2xxx`, 7/9 dígitos), moneda obligatoria en
    IBAN y **rechazada/ausente en SINPE** (Q2), crear un método **sin titular** (válido), crear un
    método con `holderEn` sin `holderEs` → error que indica qué falta, **reordenar** dentro del grupo
    con el campo numérico y ver el orden en público (SC-003/FR-004), editar y publicar los **dos
    textos por separado** (Q10), retirar con confirmación y eliminar con confirmación explícita
    (SC-002/SC-011), **denegación cruzada** (cuenta sin `donaciones` —incluida una con otros
    permisos de módulo—: `403` en la API forzada y mensaje claro en la interfaz, SC-005) y **todo
    registrado en la auditoría de F2** con los códigos `donation.*` y sus etiquetas legibles
    (SC-011). *Los logins se ejecutan por tandas respetando el límite de intentos de F2 (T737).*
    *Cubre SC-002, SC-003, SC-004, SC-005, SC-011.*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde con la
    cuenta `donaciones` y con las cuentas de denegación de T732; la auditoría muestra los códigos
    esperados con su etiqueta humana (`ACTION_LABELS` de T724).
  - **Criterio de terminado**: US2/US3 cubiertas de extremo a extremo con el permiso verificado en
    servidor y las reglas de número probadas en el flujo real.
  - **Commit sugerido**: `test(e2e): gestion de donaciones en el panel, permisos y auditoria`

- [X] T735 · E2E Playwright `ofrendas-accesibilidad.spec.ts` (teclado y foco) · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/ofrendas-accesibilidad.spec.ts` (NUEVO; patrón de
    `ministerios-accesibilidad.spec.ts` / `portada-accesibilidad.spec.ts`).
  - **Qué hace**: recorrido por **teclado** de `/ofrendas` y del bloque de la portada (FR-025,
    SC-008, quickstart §11): el primer `Tab` enfoca el enlace de salto al contenido; la travesía con
    `Tab` alcanza navegación, textos, fichas, **`CopyAccountButton`** (activable con `Enter`/
    `Espacio`, con confirmación anunciada por `role="status"`), selector de idioma y el botón del
    bloque en orden comprensible; **foco visible en cada paso** (anillo de F10, también sobre el
    número y la copia); **sin trampas de foco** (`Shift+Tab` sale de cualquier panel); el selector de
    idioma se opera con teclado sin perder el foco; y **la redirección de `/ofrendas` sin publicados
    no atrapa el foco** (el visitante termina en una portada navegable, P7-21/RG7-12). Se comprueba
    en **tema claro y oscuro** (F10) y con `prefers-reduced-motion`. *Cubre SC-008 y FR-025
    (teclado y foco; la checklist WCAG completa la cierra T737).*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde sobre la
    base con fixtures publicadas y, con 0 publicados, la redirección deja el foco en una página
    operable; 0 elementos `<img>` en F7 (Q6) como aserción extra.
  - **Criterio de terminado**: el "e2e de teclado" de la sección y del bloque existe y corre en
    `make e2e` junto a las otras specs.
  - **Commit sugerido**: `test(e2e): accesibilidad por teclado de ofrendas (SC-008)`

- [X] T736 · E2E Playwright `portada-ofrendas.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/portada-ofrendas.spec.ts` (NUEVO).
  - **Qué hace**: bloque de la portada (US6, SC-014) según `quickstart.md` §14: con 8 publicados
    muestra **exactamente los 5 primeros** en el orden de la sección (`iban` → `sinpe`,
    `sortOrder, id`); con menos, exactos, **sin huecos ni marcadores**; con **0 métodos publicados el
    bloque no existe** (sin título, sin mensaje — a diferencia de «Lo que viene» de F4 —) y la
    portada queda coherente; **posición**: inmediatamente después del bloque de ministerios (F6) y
    antes de la siguiente sección de F3; cada ficha lleva su chip de tipo y **no es clicable** (el
    único interactivo del bloque es el botón); el botón **«Ver formas de ofrendar»** lleva a
    `/ofrendas` en un solo paso; 0 borradores (SC-002); 320 px sin desplazamiento horizontal
    (SC-007) y recorrido por teclado con foco visible (SC-008). *Cubre US6, SC-014 y los aspectos de
    portada de SC-010/SC-013.*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**:
    `npx playwright test e2e/portada-ofrendas.spec.ts` en verde; los e2e de F3/F4/F6 de la portada
    (`portada-publica`, `portada-proximamente`, `portada-ministerios`) siguen pasando sin tocar.
  - **Criterio de terminado**: US6/SC-014 cubiertos de extremo a extremo en la posición acordada.
  - **Commit sugerido**: `test(e2e): bloque de ofrendas de la portada`

---

## Fase 9 — Cierre

- [X] T737 · Verificación de cierre · `[infra]`

  - **Archivos**: ninguno (evidencia en el PR). *No se toca* `.github/workflows/ci.yml` ni ningún
    archivo del kit. `.env.example`, `docker-compose.yml` y `Makefile`: **sin cambios** (F7 no añade
    variables ni servicios — P7: sin dependencias nuevas —; solo se verifica que ninguna hiciera
    falta).
  - **Qué hace**: ejecuta la verificación completa de F7 y la registra como evidencia: `make ci`
    (**lint + test + security**; `make db-migrate` al preparar el entorno, §0); **cobertura ≥ 80 %
    en `internal/donaciones/service*.go`** (`go test -cover`); migraciones `000012`/`000013`
    **up → down → up** (y `permissions` sigue en 10 con `donaciones`); `make sqlc-verify` y
    `make api-gen` **sin deriva**; `make e2e` **por tandas** respetando el **límite de intentos de
    login** de F2 (las specs del panel de F7 y de F2/F4/F6 hacen login: se ejecutan por tandas —o con
    las cuentas y ventanas que no agoten el límite— y la salida de cada tanda se adjunta);
    `.env.example` sin cambios confirmado; recorrido manual de `quickstart.md` §0–§16 (mapa §16
    (SC-001…SC-014) y títulos de §1–§14 (US/FR)); `GET /healthz` responde como en F1 y una ruta
    inexistente devuelve `404` con `ErrorEnvelope` (§8.1.9). La **checklist de accesibilidad**
    (quickstart §11, WCAG 2.1 AA, claro y oscuro) la recorre `qa-tester` (SC-008, RG7-9).
    **Prueba de usabilidad SC-012** (protocolo mínimo, estilo F3/F4/F6): **≥ 6 personas** (2 por
    franja 18–35 / 36–59 / 60+, incluidas con poca experiencia en internet), guion de **3 tareas sin
    ayuda** (identificar a qué cuenta o número donar; encontrar en qué se usan las donaciones;
    copiar o transcribir un número de cuenta sin error), sesión breve observada por `qa-tester` con
    el humano; resultado documentado en **`specs/007-donaciones/pruebas-usabilidad-SC-012.md`**
    (umbral: ≥ 90 % de tareas completadas). **El umbral de SC-001 (< 30 s) se cronometra aquí** (el
    e2e no puede medirlo): la **tarea 1** se ejecuta con **cronómetro** y el tiempo de cada
    participante se registra en el informe. Es una prueba **manual con personas** (no automatizable)
    y queda fuera de `make ci`/`make e2e`.
  - **Pruebas incluidas** (§III): — (es la ejecución de todas las anteriores). **Cómo se verifica**:
    cada comando anterior en verde y su salida adjunta al PR; el mapa de criterios → secciones queda
    recorrido completo (mapa §16 para SC y los títulos de §1–§14 para US/FR) y el informe de
    usabilidad de SC-012 adjunto **con los tiempos cronometrados de SC-001**.
  - **Criterio de terminado**: `make ci` y `make e2e` (por tandas) en verde; SC-001…SC-014
    evidenciados (con foco explícito en SC-009 — números exactos — y SC-010 — redirección sin
    destello —); el **cambio acotado RG7-6** de T725 queda registrado en la evidencia.
  - **Commit sugerido**: — (sin cambios de código)

- [ ] T738 · Documentación y PR · `[infra]`

  - **Archivos**: `README.md`, `CHANGELOG.md` (o notas de entrega) y `docs/` (EDITADOS por
    `documentador`); PR en GitHub (creado por `devops`).
  - **Qué hace**: cierra la entrega con `documentador` (README del proyecto y notas de entrega: qué
    agrega F7, la sección `/ofrendas` «Ofrendas y diezmos», el permiso existente `donaciones` **sin
    migración de permisos** y **sin cambio de nombre** (Q8), las tablas `donation_methods`/
    `donation_texts` y las migraciones `000012`/`000013`, la **validación y normalización de IBAN y
    SINPE** (FR-015/FR-016), la **redirección silenciosa** de `/ofrendas` sin publicados (distinta a
    propósito de `/eventos` de F4), el bloque de la portada (`GET /api/v1/portada/ofrendas`, tope 5)
    y la documentación de API actualizada tras la fusión del contrato **0.8.0**) y con `devops`
    (rama → PR con descripción, tareas cubiertas y los cambios documentados explícitamente:
    **RG7-6** — las pruebas de frontend de permisos/accesos — y los **defaults revisables**
    aplicados —entidad no traducible, «nada publicado» con solo textos, IBAN sin mod-97, SINPE con
    primer dígito 6/7/8, duplicado `409` por tipo+número —la edición **se excluye a sí misma**:
    editar sin cambiar el número no da `409`—, textos sin borrado—). *Cubre el cierre de
    US1–US6/FR-001…FR-025 como entregable.*
  - **Pruebas incluidas** (§III): — (documentación y proceso). **Cómo se verifica**: la documentación
    refleja el comportamiento real (la prueba es el `quickstart.md` ejecutado en T737); el PR
    referencia este `tasks.md` con las tareas marcadas y pasa por la revisión final (`qa-tester`,
    `revisor-codigo`, `seguridad`).
  - **Criterio de terminado**: PR abierto y **aprobado por el humano** (regla 1: merge y producción
    requieren aprobación explícita). El merge lo ejecuta el orquestador con `devops`.
  - **Commit sugerido**: `docs: notas de entrega de F7 (ofrendas y diezmos)`

---

## Estrategia de pruebas (resumen ejecutable)

Cobertura exigida: **80 % en `service/`** (§III). Comandos: `go test ./...` ·
`go test -tags=integration ./...` · `npm test -- --run` · `make e2e` · `make ci`.

| Capa | Tipo | Qué verifica F7 | Dónde |
|---|---|---|---|
| `platform/audit` | Unitaria | Los 8 códigos nuevos en el registro cerrado; objetivo `content` con `targetLabel` | `internal/platform/audit/*_test.go` (T706) |
| `donaciones/numero` | Unitaria (tabla de casos) | IBAN aceptados `CR` 22, `ES` 24, `DE` 22, `GB` 22 (con espacios/guiones y en minúsculas) → normalizados; rechazados longitud errónea por país, país fuera del registro, letras/emojis; SINPE aceptados con y sin `+506` → `+506XXXXXXXX`; rechazados 7/9 dígitos, `2xxx`, letras, `+507`, sufijo extra | `internal/donaciones/numero_test.go` (T707) |
| `service` | Unitaria con fakes | Solo `published` **y por texto**; orden `iban → sinpe` y `sortOrder, id`; fallback `en → es` por campo (nunca huecos); claves omitidas sin contenido; **tope 200 y tope 5 fijados por prueba**; DTOs sin `publicationState`, sin `currency` en SINPE, sin `holder` vacío; validación completa (obligatorios, moneda por tipo, español por campo, números, límites) con `details` y **sin guardado parcial**; normalización antes de persistir; duplicado → `conflict` (alta repetida y `PATCH` hacia el número de **otro** método) y **editar sin cambiar el número no lo dispara**; publicar/retirar por método **y por texto**; "se publica al guardar"; `PATCH` que fusiona y `PUT` que reemplaza; auditoría con la regla exacta de códigos (fail-closed; datos + estado → dos filas; alta = solo `donation.create`) | `internal/donaciones/service_*_test.go` (T712–T715) |
| `handler` | Unitaria `httptest` | Cada operación: decodificación, `details` en 400, códigos 200/201/204/400/401/403/404/409, sobres uniformes; **el JSON público jamás contiene `publicationState`**; `GET /api/v1/ofrendas` y `/portada/ofrendas`: `200` con `no-store`, claves omitidas con 0 publicados | `internal/donaciones/handler_*_test.go` (T717–T720) |
| `repository` | **Integración** (PostgreSQL real) | `CHECK`s de formato por tipo, moneda por tipo, español por campo, `btrim`, `text_key` y `UNIQUE (text_key)`; `ListPublished*` **nunca** devuelve borradores; orden `CASE kind …, sort_order, id` y `LIMIT` (5/200); `ExistsDonationMethodDuplicate` con normalización y **exclusión de la fila que se edita** (`:exclude_id`: "editar sin cambiar el número no da `409`"); `UpsertDonationText` sin duplicar; **mutación + `admin_actions` atómicas** (y rollback si falla el registro); migraciones up→down→up; `permissions` sigue en 10 | `internal/donaciones/repository_*_test.go` (T709–T711) |
| `cmd/api` | Humo de composición | Rutas nuevas con su sobre y permiso, `/healthz` intacto, `404` `not_found` en rutas inexistentes, **403 cruzado** por permiso | `cmd/api/main_test.go` (T721) |
| Frontend | Unitaria (Vitest + MSW) | Pública: dos grupos en el orden del API; IBAN en bloques en pantalla y copia = `number` (SC-009); SINPE normalizado; sin titular → 0 líneas; sin moneda en SINPE; fallback visible; navegación y bloque ocultos sin publicados; **redirección sin destello** (esqueleto; `<Navigate replace>`; ante error NO redirige); bloque con 1…5 y **ausente con 0**; panel: errores por campo, moneda oculta en SINPE, textos con su `StatusPill`, publicar/retirar, `RequirePermission(DONACIONES)`, `ACTION_LABELS` +8, error sin perder lo escrito; y las pruebas actualizadas (RG7-6) | `frontend/src/features/*/*.test.tsx` (T722–T731) |
| E2E pública | Playwright | Recorrido de `quickstart.md`: dos grupos en orden exacto, números íntegros, moneda solo IBAN, titular solo si existe, borradores invisibles, **`/ofrendas` redirige sin nada publicado** (sin destello, «Atrás» sin bucles), inglés con fallback y datos idénticos, 3 anchos | `frontend/e2e/ofrendas-publica.spec.ts` (T733) |
| E2E panel | Playwright | Crear+publicar < 2 min, IBAN con espacios normalizado, IBAN/SINPE inválidos rechazados, moneda por tipo, sin titular, reordenar y ver el orden en público, dos textos por separado, eliminar con confirmación, denegación cruzada (403 forzado) y auditoría | `frontend/e2e/ofrendas-panel.spec.ts` (T734) |
| E2E accesibilidad | Playwright | Teclado y foco en `/ofrendas` y el bloque, claro y oscuro, copia activable, **la redirección no atrapa el foco** | `frontend/e2e/ofrendas-accesibilidad.spec.ts` (T735) |
| E2E portada | Playwright | Exactamente 5 con 8 publicados, exactos con menos, **sin bloque con 0**, posición tras ministerios, botón a `/ofrendas`, 320 px y teclado | `frontend/e2e/portada-ofrendas.spec.ts` (T736) |

**Pruebas que no faltan** (las que la spec hace críticas): **0 borradores expuestos** por ninguna
respuesta pública (SC-002: ni métodos ni textos); **0 números mal formados** aceptados ni
renderizados y **0 caracteres alterados** en los números publicados (SC-009: IBAN en bloques, SINPE
normalizado, copia = valor exacto); **cambio guardado visible en la primera carga** (SC-003);
**fallback `en → es` sin campos vacíos** y **número/tipo/entidad/moneda idénticos entre idiomas**
(SC-006); **permisos verificados en servidor** con cruce contra otros módulos (SC-005); **toda
operación queda registrada** —éxito, fallo y denegación— y el registro es de solo lectura (SC-011);
**sin nada publicado la URL directa `/ofrendas` redirige a la portada sin destello** (SC-010,
P7-21); **bloque con exactamente 5 y oculto sin métodos publicados** (SC-014); **texto con tildes,
`ñ`, `¿¡` y emojis intacto** y **números sin emojis ni letras fuera de estructura** (edge cases);
**ficha sin titular y SINPE sin moneda** presentadas sin huecos (US1 esc. 5, SC-010); **los dos
textos con su propio estado** (texto en borrador con métodos publicados no deja hueco: US3 esc. 7).

---

## Cobertura de requisitos (US1–US6, FR-001…FR-025, SC-001…SC-014)

| Historia | Tareas | Verificación principal |
|---|---|---|
| **US1** Ver cómo apoyar a la iglesia | T703, T710, T712, T713, T717, T722, T723, T726, T727, T733 | quickstart §1/§4/§5/§7 · e2e T733 (SC-001, SC-009, SC-013) |
| **US2** Gestionar la información desde el panel | T709, T715, T719, T720, T721, T724, T725, T729, T730, T731, T734 | quickstart §2–§7 · e2e T734 (SC-004, SC-005, SC-011) |
| **US3** Controlar la publicación | T713, T714, T715, T726, T728, T729, T731, T733, T734 | quickstart §8 · e2e T733/T734 (SC-002, SC-003, SC-010) |
| **US4** Ver la información en inglés | T713, T714, T723, T726, T728, T733 | quickstart §9 · e2e T733 (SC-006) |
| **US5** Cualquier dispositivo y accesible | T726, T728, T733, T735, T736, T737 | quickstart §11 · e2e en 3 anchos (SC-007) · teclado T735 (SC-008) |
| **US6** Ver las ofrendas desde la portada | T714, T718, T728, T736 | quickstart §14 · e2e T736 (SC-014) |

| FR | Tareas | Verificación principal |
|---|---|---|
| FR-001 (solo datos de donación; nada de F4/F5/F6; sin montos ni donantes) | T701, T712, T726, T728 | quickstart §1 · pruebas de modelo (sin columnas ajenas) · SC-013 |
| FR-002 (sección pública: intro + grupos + uso) | T703, T710, T713, T717, T726 | quickstart §1/§7 · pruebas de service/handler (SC-001) |
| FR-003 (tipo, número exacto, entidad, titular si existe, moneda solo IBAN) | T707, T712, T713, T726, T730 | quickstart §3–§5 · tabla de casos + e2e (SC-009) |
| FR-004 (agrupado iban→sinpe; `sort_order, id` por grupo) | T701, T703, T713, T714, T715, T726, T728, T729 | quickstart §1/§6 · pruebas de orden (SC-014) |
| FR-005 (ocultar todo sin publicados + **redirigir** la URL directa) | T713, T726, T727, T733 | quickstart §10 · e2e T733 (SC-010, RG7-12) |
| FR-006 (ruta `/ofrendas` y entrada única «Ofrendas y diezmos») | T723, T726, T727 | quickstart §1 · e2e T733 (SC-010) |
| FR-007 (bilingüe por campo; español base; datos no traducibles) | T701, T708, T712, T723, T726, T730 | quickstart §4/§9 · pruebas de validación (SC-006) |
| FR-008 (fallback `en → es`; nunca campo vacío) | T713, T714, T726, T728 | quickstart §9 · tabla de casos del service (SC-006) |
| FR-009 (selector de idioma existente) | — (F10 sin cambios), T733 | quickstart §9 · e2e T733 |
| FR-010 (CRUD del panel + textos) | T709, T715, T719, T720, T729, T730, T731 | quickstart §2–§7 · e2e T734 (SC-004) |
| FR-011 (permiso `donaciones` existente, verificado en servidor, independiente) | T721, T724, T725, T729, T734 | quickstart §2/§12 · humo T721 (403 cruzado) · e2e T734 (SC-005) |
| FR-012 (borrador/publicado por método **y por texto**; 0 borradores por ninguna vía) | T703, T710, T713, T714, T715, T726, T728, T729, T731, T733 | quickstart §8 · pruebas «0 borradores» · e2e T733 (SC-002) |
| FR-013 (se publica al guardar; `no-store`) | T715, T717, T718, T722, T734 | quickstart §8 · e2e T734 (SC-003) |
| FR-014 (validación completa; sin cambios parciales) | T712, T715, T719, T720, T730 | quickstart §3–§5 (tablas de errores) · pruebas de service/handler |
| FR-015 (IBAN por estructura y longitud; guardado normalizado; presentación en bloques) | T701, T707, T712, T715, T726, T730, T733, T734 | quickstart §4 · tabla de casos + e2e (SC-009) |
| FR-016 (SINPE celular CR 8 dígitos ± `+506`; prefijo normalizado) | T701, T707, T712, T715, T726, T730, T733, T734 | quickstart §5 · tabla de casos + e2e (SC-009) |
| FR-017 (auditoría atómica, 8 códigos, solo lectura) | T702, T706, T708, T711, T715, T716, T724, T734 | quickstart §13 · integración (mutación+registro) · e2e T734 (SC-011) |
| FR-018 (panel en español) | T723, T729, T730, T731 | revisión de `revisor-codigo` (0 cadenas del panel por `t()`) |
| FR-019 (tipo/número/entidad obligatorios; titular opcional; moneda solo IBAN; sin notas) | T701, T709, T712, T715, T730 | quickstart §3/§4 · pruebas de persistencia |
| FR-020 (sin imágenes) | T704, T726, T728, T735 | quickstart (sin § de imágenes, a propósito) · revisión de `revisor-codigo` (0 rutas de subida nuevas, `media` intacto) |
| FR-021 (intro y uso: textos libres con estado propio y traducibles) | T701, T703, T710, T715, T720, T726, T731 | quickstart §7 · e2e T734/T733 (US3 esc. 7) |
| FR-022 (bloque: tope 5 en servidor, orden de la sección, botón, posición, oculto sin métodos) | T714, T718, T728, T736 | quickstart §14 · pruebas del service (SC-014) · e2e T736 |
| FR-023 (F10 aplicado; solo el bloque en la portada de F3) | T726, T728 | revisión de `revisor-codigo` sobre `HomePage` (una inserción) y `PublicLayout` sin cambios |
| FR-024 (responsividad, navegación simple; números legibles) | T726, T728, T733, T736 | quickstart §11 · e2e en 320/768/1280 (SC-007) |
| FR-025 (accesibilidad WCAG 2.1 AA, claro y oscuro) | T726, T728, T735, T737 | checklist de quickstart §11 · e2e de teclado T735 (SC-008) |

| SC | Tareas | Verificación principal |
|---|---|---|
| SC-001 (ve la información y encuentra cómo donar en < 30 s) | T726, T733, T737 | quickstart §1 · e2e T733 (contenido; **el e2e no cronometra**) · el umbral **< 30 s se cronometra** en el protocolo de usabilidad de SC-012 (**T737**, tarea 1 con cronómetro) |
| SC-002 (0 borradores visibles: ni métodos ni textos) | T710, T713, T714, T726, T728, T733 | pruebas «0 borradores» (service/repository) · e2e T733 |
| SC-003 (cambio guardado visible en la primera carga) | T715, T722, T729, T734 | e2e T734 · invalidación de caché (T722) · `no-store` (T717/T718) |
| SC-004 (crear y publicar un método en < 2 min) | T730, T734 | quickstart §3 · e2e T734 |
| SC-005 (permiso verificado en servidor; 0 operaciones sin permiso) | T721, T724, T729, T734 | humo T721 · e2e T734 (cruce de permisos) |
| SC-006 (traducción 100 %, 0 campos vacíos, datos idénticos entre idiomas) | T713, T714, T723, T726, T728, T733 | pruebas de i18n y fallback · e2e T733 (comparación de valores) |
| SC-007 (usable en teléfono/tableta/PC; 0 scroll horizontal) | T726, T728, T733, T736 | e2e en 320/768/1280 (T733) y bloque a 320 px (T736) |
| SC-008 (accesibilidad; teclado y lector; claro y oscuro) | T735, T737 | e2e de teclado T735 · checklist WCAG (T737) |
| SC-009 (números íntegros, exactos y copiables; 0 mal formados) | T707, T712, T726, T730, T733, T734 | tabla de casos de `numero.go` · e2e T734 (rechazo) y T733 (pantalla/copia = `number`) |
| SC-010 (0 secciones vacías ni campos en blanco; navegación sin huérfanos; redirección limpia) | T713, T726, T727, T728, T733 | pruebas de `useSiteLinks`/página/bloque · e2e T733 (RG7-12) |
| SC-011 (100 % de operaciones registradas; registro de solo lectura) | T711, T715, T716, T734 | quickstart §13 · integración transaccional · e2e T734 |
| SC-012 (≥ 90 % de personas identifican cómo donar y en qué se usan las donaciones) | T737 | prueba manual con personas (`pruebas-usabilidad-SC-012.md`) |
| SC-013 (0 eventos/actividades/grupos/ministerios en la sección ni el bloque) | T701, T726, T728, T733, T736 | revisión de contenido e2e · pruebas de modelo (sin columnas ajenas) |
| SC-014 (bloque: exactamente 5 o todos; 0 huecos; debajo de ministerios; oculto sin métodos) | T714, T728, T736 | pruebas del service (tope 5, orden, `items` omitida) · e2e T736 |

También quedan cubiertos los **Edge Cases** de la spec sin FR propio: **borradores filtrados por
cualquier vía** (T710, T713, T714, T733) · **dato publicado que se edita** (visible en la primera
carga: T715, T734) · **IBAN mal formado o de longitud errónea** (rechazado sin guardar: T707, T712,
T730, T734) · **IBAN con espacios/guiones** (válido y normalizado: T707, T715, T726, T734) ·
**SINPE mal formado** (T707, T712, T730, T734) · **SINPE con o sin `+506`** (normalizado: T707, T715,
T734) · **IBAN sin titular** (ficha completa sin línea: T713, T726, T733) · **SINPE con moneda**
(imposible: campo ausente y `CHECK`; T701, T712, T730) · **duplicado exacto → `409`** (T709, T715,
T730) y **editar un método sin cambiar el número no da `409`** (T709, T715, T719) · **cambio de
cuenta bancaria** (edición normal sin historial: T715, T734; se documenta) ·
**textos largos con varios párrafos** (límite 2.000, saltos intactos: T712, T726, T731) · **texto en
borrador con métodos publicados** (sin hueco: T713, T726, T731, T733) · **tildes, `ñ`, `¿¡` y emojis
intactos** y **emojis/letras en números rechazados** (T707, T712, T731) · **inglés sin español por
campo** (T712, T715, T720, T730, T731) · **números idénticos entre idiomas** (T713, T733) ·
**operación sin permiso por cualquier vía** (T721, T724, T729, T734) · **operación no completada sin
registro y registro sin operación** (T711, T715) · **ediciones simultáneas** (última escritura
completa gana: T709, T719) · **eliminación definitiva con confirmación y registrada** (T715, T729) ·
**muchos métodos publicados** (tope 200 en la sección; sin límite artificial: T713) · **más de 5 en el
bloque** (solo los 5 primeros: T714, T736) · **sección sin nada publicado** (redirección sin
destello: T726, T727, T733).

**Cambios acotados y documentados registrados**: **T725** actualiza las pruebas de frontend que fijan
la lista de permisos y accesos (`features/roles/permissions.test.ts`, `features/roles/roles.test.tsx`,
`features/panel/panel.test.tsx` y, si aplica, `lib/permissions.test.ts`) al **activar** `donaciones`
(**RG7-6**: cambio de producto autorizado, **no** debilitamiento — solo recuentos/listas/sellos —).
El backend **no** toca pruebas de F2: el catálogo de permisos sigue en **10** (a diferencia de F4,
que necesitó la excepción T405). F7 **no** promueve piezas al inventario único y **no** toca
`media`/`IsFilePublished` (Q6/P7-16).

---

## Defaults revisables aplicados (preguntas abiertas para el humano; nada se implementa como regla cerrada sin esto)

> Son los cuatro huecos/contradicciones que el plan señaló y sus dos confirmaciones menores. Las
> tareas implementan los **defaults** indicados y los fijan con pruebas, de modo que cambiarlos es un
> cambio acotado (ver plan.md «Preguntas para el humano» y RG7-1…RG7-12).

1. **Entidad financiera traducible** (contradicción FR-019/Key Entities vs FR-007/US4 esc. 5/SC-006):
   default **NO traducible** (T701/T708/T730; única lectura que garantiza SC-006). Si el humano
   decide que sí: `institution_es`/`institution_en` + `CHECK` de español por campo + ajuste de
   SC-006/US4 esc. 5 (migración nueva y contrato; nada más se rompe).
2. **«Nada publicado» con solo textos publicados y 0 métodos** (la spec no tiene escenario): default
   **la sección vive** (T713/T727: navegación viva, redirección **no** se dispara) y **el bloque no
   existe** (T714/T728). Fijado por prueba en T727/T733. Alternativa: exigir ≥1 método
   (`hasOfrendasContent`).
3. **Mod-97 del IBAN**: default **no** (solo estructura + longitud por país, T707) para no rechazar
   lo que la spec acepta. Endurecimiento posible: una función más en `numero.go` y sus casos.
4. **Primer dígito del SINPE 6/7/8** (celular vs línea fija, RG7-3): default **sí** (T707: se
   rechazan los `2xxx`). Quitar la condición es una línea + sus casos de prueba.
5. **Duplicados** (P7-10): se **rechazan** con `409` por **tipo + número normalizado**, **sin
   `UNIQUE`** en la BD y **excluyendo la fila que se edita** (editar sin cambiar el número no da
   `409`). Si se prefiere la regla literal de Assumptions (tipo + entidad + número) o un
   índice único, es un cambio acotado (T709/T715).
6. **Textos sin borrado** (P7-17): los textos solo se retiran en `draft`; no hay `DELETE`. Si el
   equipo lo necesita, se añade `DELETE /api/v1/admin/donaciones/textos/{textKey}` al contrato y a
   T720/T731.
7. **Presentación del SINPE** (`ux.md` D-4, marcada ⟲): se muestra en su formato normalizado; la
   agrupación de lectura (`+506 8888 8888`) espera confirmación humana y, si se aprueba, es solo
   presentación (T726/`lib/numero.ts`) con el invariante de RG7-7.
