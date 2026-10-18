# Tasks: Medios — prédicas y podcasts (F9)

**Branch**: `009-medios-predicas-y-podcasts` | **Date**: 2026-10-17
| **Spec**: [spec.md](./spec.md) (APROBADA **v2**, US1–US6 / FR-001…FR-030 / SC-001…SC-017; modelo
aprobado por el humano el 2026-10-17 —decisión 27 del roadmap— y **cinco precisiones aprobadas en la
puerta 2** el 2026-10-17: «Aplica las 5 y apruebo el plan») · **Plan**: [plan.md](./plan.md)
(P9-1…P9-18, riesgos RG9-1…RG9-12, «Precisiones aprobadas en la puerta 2») · **UX**: [ux.md](./ux.md)
(D-1…D-33; §6 catálogo de componentes —manda en nombres y composición—; §7 textos)

**Input**: plan aprobado (`plan.md`, P9-1…P9-18 + «Propuesta de tareas») · `research.md`
(R9-1…R9-14) · `data-model.md` (migraciones `000014`/`000015`, tablas `media_items` +
`media_episode_sources` + `media_sync_runs`, ampliación del `CHECK` de `admin_actions` con **6
códigos `media.*`**; **sin migración de permisos**: `medios` ya está sembrado por `000002` y el
catálogo sigue en **10**) · `contracts/openapi.yaml` (delta de diseño, inmutable; **7 operaciones**,
`info.version` 0.9.0) · `ux.md` (pantallas, componentes, textos, estados, D-1…D-33) ·
`quickstart.md` (§0…§14, incluida §0.1 captura de fixtures y §5/§8 siembra por SQL) · constitución
(`.specify/memory/constitution.md`) · skills `go-backend`, `postgres-db`, `react-frontend` · F1, F2,
F3, F4, F6, F7 y F10 ya integradas (`specs/001-estructura-base/`, `specs/002-acceso-gestion-usuarios/`
T201…T254, `specs/003-portada-info-general/` T301…T340, `specs/004-eventos-actividades/` T401…T449,
`specs/006-ministerios/` T601…T638, `specs/007-donaciones/` T701…T738, `specs/010-sistema-diseno-movimiento/`).

> **Regla de ejecución**: **un commit por tarea** (Conventional Commits; los hooks del kit lo
> validan). **Test-first** (constitución §III): la prueba de cada tarea se escribe **antes** que su
> código y debe **fallar** antes de implementar; toda tarea de código incluye sus pruebas en el mismo
> commit — backend con `go test` y, donde toca persistencia real, `go test -tags=integration`
> (`//go:build integration`, PostgreSQL real con `testcontainers-go`); frontend con Vitest + Testing
> Library + MSW; flujos críticos con Playwright (`make e2e`). Cobertura exigida **≥ 80 % en
> `internal/medios/service*.go`** (`go test -cover`). Quien escribe no aprueba: cada tarea pasa por
> `qa-tester`, `revisor-codigo` y `seguridad` antes de integrarse (flujo del orquestador). **El CI
> del kit (`.github/workflows/ci.yml`) no se toca** y **ninguna migración aplicada se edita** (§VI).
> **Sin dependencias nuevas** (research R9-13: `encoding/xml`/`net/http`/`time` de la stdlib;
> `golang.org/x/text` y `golang.org/x/sync` ya están en `go.mod` y solo pasan de indirectas a
> directas, sin cambiar versiones). F9 **no toca pruebas de backend de F2** (el catálogo de permisos
> **no** cambia: sigue en 10); solo la actualización **acotada** de las 2–4 pruebas de frontend que
> fijan permisos/accesos (T932, RG9-6). **Los e2e NO pulsan «Buscar episodios nuevos»** (RG9-7): F9
> no tiene alta manual y la importación real se cubre con fakes/fixtures en Go y con el recorrido
> manual del `quickstart.md` §3–§4; los e2e **siembran por SQL** (`e2e/mediosSeed.ts`, patrón nuevo
> declarado en el plan).

## Decisiones del plan (puerta 2, 2026-10-17) que estas tareas respetan (sin reabrir)

- **P9-1 — tres tablas, un dominio**: `media_items` (la ficha por sermón, con `visibility`
  `visible`/`hidden`, marcas con resolución, puntero de unión) + `media_episode_sources` (una fila
  por elemento de feed, **`UNIQUE (provider, external_id)`** = idempotencia) + `media_sync_runs`
  (cada búsqueda) + rama aditiva de `IsFilePublished` (carátulas copiadas); consultas en
  `queries/media_episodes.sql`/`media_episode_sources.sql`/`media_sync_runs.sql` + rama de
  `media.sql`; dominio `internal/medios/`.
- **P9-2 — superficie HTTP**: `GET /api/v1/medios` (sección paginada, `limit` 12 por defecto, tope
  50), `GET /api/v1/portada/medios` (tope 5 **firme en servidor**) y `/api/v1/admin/medios` (listado
  con `lastSync`, `POST buscar-nuevos`, `PATCH {id}`, `POST {id}/unir`, `POST {id}/marcas/resolver`);
  públicos con `Cache-Control: no-store` y `?lang=es|en`. Público y panel se llaman `medios` (Q7/Q13).
- **P9-3 — búsqueda sincrónica** en `POST …/buscar-nuevos` con presupuesto de **90 s** (constante
  revisable), timeouts 10/10/5 s, máximo 4 descargas/comprobaciones en vuelo y topes duros 30/30/30;
  **sin cron, worker ni cola** (FR-011); el resumen parcial ante el agotamiento del presupuesto
  nunca pierde ni oculta nada (R9-9).
- **P9-4 — parser propio** de Atom (YouTube) y RSS 2.0 (podcast) con `encoding/xml`, detrás de la
  interfaz `FeedSource`, con **fixtures XML reales** commiteados (`testdata/`) y fakes en pruebas
  (R9-3/R9-5); 0 dependencias nuevas.
- **P9-5 — unión por título normalizado + mismo día** en `America/Costa_Rica` **sin margen**
  (precisión 4 de la puerta 2); lo dudoso = entradas separadas, adicional oculta, **ambas marcadas**
  «posible duplicado» (FR-014).
- **P9-6 — disponibilidad**: YouTube por **oEmbed** (`200`/`404`/resto) y podcast por **GET del
  enlace del episodio** (`200`/`404`-`410`/resto); **solo un 404/410 confirmado** oculta y marca «no
  disponible»; todo lo demás solo marca y **nunca** oculta (regla a prueba de fallos, FR-017).
- **P9-7 — carátula copiada** con `platform/storage` desde una descarga **acotada** (https, allowlist
  `i.ytimg.com`/`d3t3ozftmdmh3i.cloudfront.net`, sin IPs privadas ni redirects fuera, `LimitReader`
  con `UPLOAD_MAX_BYTES`, firma binaria con `storage.Detect`); fallo → ficha sin carátula (marcador
  neutro), nunca aborta la importación (R9-8).
- **P9-8 — estado `visibility` (`visible`/`hidden`)** + marcas en **dos ejes** con su resolución
  (P9-19); **el único bloqueo público por marca es el par `possible_duplicate` SIN resolver**
  (síntesis R9-7; precisión 1 de la puerta 2); `missing_from_feed` **sigue visible**; una marca
  resuelta no se reabre por la misma causa.
- **P9-9 — instantánea al importar**: la búsqueda solo **agrega** lo que falta y aplica FR-017; no
  refresca título, fecha, carátula, enlaces ni descripción (R9-6).
- **P9-10 — unión por puntero** `merged_into_id` (la absorbida sale de todas las consultas; sus
  `sources` se repuntan; **cero `DELETE`**); base = la entrada del **video** (nombre, carátula,
  fecha) —precisión 5 de la puerta 2— y **manda la descripción editada** si solo una lo está.
- **P9-11 — idempotencia en la BD**: `UNIQUE (provider, external_id)` + `INSERT … ON CONFLICT DO
  NOTHING` (R9-11).
- **P9-12 — sección paginada** (`limit`/`offset`, 12 por tanda, tope 50) + **bloque de portada con
  tope 5 firme** en servidor, mismo esquema de ítem y `items` omitida sin contenido (R9-10).
- **P9-13 — bilingüe** `?lang=es|en` con fallback `en → es` **solo sobre la descripción corta**
  (`description` ya localizada en el servidor); nombre, fecha, carátula y enlaces son un único valor
  (FR-025/FR-026); panel en español (FR-024).
- **P9-14 — auditoría**: 6 códigos `media.*` (migración `000015`), `InsertAdminAction` **dentro de la
  transacción** de cada mutación y `audit.Recorder` del dominio (patrón
  `donaciones/service_audit.go`); la búsqueda completa = **una** fila `media.sync` con extracto en
  `targetLabel` y detalle en `media_sync_runs`.
- **P9-15 — permisos**: `AdminChain("medios")` con el permiso **ya sembrado** por `000002`; **sin
  migración de permisos**, el catálogo sigue en 10; frontend `MEDIOS` en `lib/permissions` y
  `isPermissionAvailable('medios') → true`.
- **P9-16 — frontend**: `features/publico/medios/` (sección + bloque + ficha + enlaces anunciados) y
  `features/medios/` (panel); navegación desde el **documento de la sección** (`usePublicMedios`,
  patrón P7-8/P9-16); **redirección de `/medios`** con la mecánica P7-21 (`SectionSkeleton` →
  `<Navigate to="/" replace />` **solo con éxito y 0 publicables**; **ante error NO se redirige**).
- **P9-17 — descripción corta**: recorte al importar (HTML → texto plano, primeras frases, **≤ 300
  caracteres**, corte en frontera de frase o palabra con «…»); editable (es/en) y protegida por
  `descriptionEdited`; si el feed no trae descripción, el campo queda **ausente** y la ficha omite la
  línea (precisión 3 de la puerta 2).
- **P9-18 — feeds y topes como constantes del dominio** (las dos URLs de FR-010; timeouts, allowlist
  y topes 30/30/30/90 s documentados y revisables); **sin variables de entorno nuevas**.
- **P9-19 — marcas en dos ejes + precedencia (corrección de la fase 5: C1/A3/M2)**: el eje de
  **duplicados** (`duplicate_of_id` recíproco + `duplicate_mark_resolved_at`) y el eje de
  **disponibilidad** (`availability_mark` + `availability_mark_resolved_at` sobre `availability_*`)
  son **independientes** y **conviven** (par abierto + enlace confirmado muerto es un estado legal);
  el resultado de la ficha es **el peor** de sus fuentes comprobadas; **la indisponibilidad
  confirmada manda** (oculta aunque el equipo hubiera resuelto «mantener visible» — precisión
  aprobada en la fase 5, SC-007 —) y **cambiar de clase de marca resetea la resolución** (misma
  clase: no se reabre; clase distinta: la nueva nace abierta).
- **P9-20 — fábrica única de cliente HTTP saliente** (`internal/medios/cliente_http.go`):
  `CheckRedirect` validado contra la allowlist del consumidor, rechazo de IPs privadas tras resolver
  DNS, `io.LimitReader` en **todo** cuerpo y timeouts; la usan los **tres** consumidores (feeds,
  probadores y carátulas) con la **misma tabla de pruebas** (corrección de la fase 5: A1).

**Precisión aprobada en la fase 5 (2026-10-17)**, además de las cinco de la puerta 2: **la
indisponibilidad confirmada manda** (A3) — la comprobación que confirma que un enlace ya no está
oculta la ficha **aunque el equipo hubiera resuelto antes «mantener visible»**; queda marcada y el
equipo puede «Mostrar de nuevo» —. Va fijada por prueba en **T921** (matriz de marcas) y **T941**,
y matiza **SC-007** en `spec.md` (única frase tocada de la spec).

**Precisiones aprobadas en la puerta 2 (2026-10-17) — mandan sobre la redacción previa de la spec**:
(1) solo un «posible duplicado» sin resolver mantiene la ficha fuera del público; una marcada «ya
no aparece en el feed» **sigue visible** (FR-009/FR-017/SC-002); (2) un enlace confirmado no
disponible **oculta la ficha completa**, también la de dos enlaces (FR-017); (3) el episodio sin
descripción en el feed **se importa igual** y la ficha **omite** esa línea (FR-020); (4) el «mismo
día» se calcula en `America/Costa_Rica` **sin margen** y la duda se marca en vez de unir (FR-013);
(5) el nombre que manda al unir es el del **feed del video** (default #7).

## Alineación de nomenclatura (decidida por el arquitecto sobre `ux.md` / `plan.md`; no se reabre)

1. **Rutas SPA**: la sección pública es **`/medios`** (una sola entrada de navegación «Medios», al
   final de la lista de F10 tras «Ofrendas y diezmos», `ux.md` D-1) y el panel es
   **`/panel/medios`** (`ux.md` §3). El bloque de la portada vive en `HomePage`
   **inmediatamente después de `<OfrendasBlock />`** (Q12/FR-027) y es el **único** contenido de F9
   en la portada de F3.
2. **Carpetas de features**: público en **`frontend/src/features/publico/medios/`** y panel en
   **`frontend/src/features/medios/`**. **Nombres de componentes y hooks: los de `ux.md` §6** (el
   plan los declaró alias provisionales —RG9-5/Complexity 7— y aquí se resuelven): público
   (`MediosPage`, `MediosGrid`, `MedioCard`, `CaratulaNeutra`, `MedioLinks`, `CargarMasButton`,
   `MediosBlock`, `MedioItem`, `usePublicMedios`, `usePortadaMedios`); panel (`pages/MediosPage`,
   `BuscarEpisodios`, `ResumenImportacion`, `FiltroMarcas`, `EpisodiosList`, `DescripcionForm`,
   `UnirDuplicadosDialog`, `MarcaDeRevision`, `ResolverMarcaDialog` —este último es el único nombre
   que aporta el plan porque `ux.md` no lo nombra—, `OrigenBadge`, `FuentesList`, `validation.ts`,
   `errors.ts`, `messages.ts`).
   **No existe pieza `ExternalLink`**: los enlaces externos usan `PublicButton` con `external` dentro
   de `MedioLinks`/`MedioItem` (el `target`/`rel` ya los resuelve el componente de F10).
3. **Rutas de API**: las del contrato (`GET /api/v1/medios`, `GET /api/v1/portada/medios`,
   `GET /api/v1/admin/medios`, `POST /api/v1/admin/medios/buscar-nuevos`,
   `PATCH /api/v1/admin/medios/{id}`, `POST /api/v1/admin/medios/{id}/unir`,
   `POST /api/v1/admin/medios/{id}/marcas/resolver`); el snapshot `contracts/openapi.yaml` es
   inmutable tras la aprobación del plan (§8.1.5: la fusión es T904; la **enmienda de la fase 5**
   —C1/M2/B2/A3: marcas por eje, `mark` en `ResolverMarcaInput` y prosa de agregación/`markedCount`—
   está justificada en `plan.md`).
4. **Permiso**: código **`medios`** (el `code` sembrado por `000002`, etiqueta «Medios (prédicas y
   podcasts)»). `lib/permissions.ts` expone `MEDIOS`; el guard es el `RequirePermission` existente y
   el servidor usa `AdminChain("medios")`. **Sin permisos nuevos** (FR-021: el catálogo sigue en 10)
   y **sin renombrar** el permiso (Q13).
5. **Numeración T901…T945** (patrón por funcionalidad: F2 `T2xx`, F3 `T3xx`, F4 `T4xx`, F6 `T6xx`,
   F7 `T7xx`; **F9 empieza en `T901`**).
6. **Auditoría**: 6 códigos `media.*` —`media.sync`, `media.description.update`, `media.hide`,
   `media.show`, `media.merge`, `media.mark.resolve`— (`data-model.md`, P9-14); `target_label`
   `Medios · <title>` (máx. 254) o `Medios · Búsqueda` para `media.sync` (con el extracto del
   resumen). El alta importada por la búsqueda **nunca** emite `media.*` extra (solo `media.sync`).
7. **Registro de rutas del SPA**: se reparte entre las tareas de página —`/medios` en T933 y
   `/panel/medios` en T936— para que **cada commit compile** (mismo reparto que F6/T627·T630 y
   F7/T726·T729).
8. **`SiteLinkId`**: crece con **`'media'`** y la clave i18n es **`nav.sections.media`** (`ux.md`
   §7.2); lo consumen a la vez `PublicNav`, `MobileMenu` y `FooterSiteMap` (grupo **Explorar**, tras
   «Ofrendas y diezmos»: D-1). La entrada existe **solo** si `usePublicMedios` trae `total > 0`
   (P9-16/P7-8: la presencia bebe del **documento de la sección**, nunca del bloque).
9. **Decisiones dirimidas entre `ux.md` y el plan aprobado** (manda el plan/contrato; `ux.md` ajusta
   composición sin tocar la API):
   - **Estado `visibility`** (P9-8/Complexity 1): el panel muestra `StatusPill` «Visible»/«Oculto»;
     **no** se usa `PublishControls` ni `publication_state` (la spec lo excluye: Consecuencia 5).
   - **Filtro de marcas (D-10 ⟲)**: se implementa como **filtro de vista** del listado del panel
     («Todo» / «Solo lo pendiente de revisión», con conteos) sobre los parámetros reales del contrato
     (`visibility`, `reviewMark`); con los dos ejes (P9-19), «pendiente» = **par abierto** o **marca
     de disponibilidad sin resolver** y el parámetro `reviewMark` filtra por la marca de esa clase
     **pendiente** (semántica fijada en T925/`contracts`); `includeMerged` queda solo para revisión
     avanzada (por defecto las absorbidas **no** se listan). Es provisional: retirarlo no toca la API.
   - **Iconos por marca (D-30 ⟲)**: **no** se implementan; la píldora con **palabra** +
     explicador de una línea ya cumple «no solo color» (FR-030). Si el humano los aprueba, es
     presentación en `MarcaDeRevision`.
   - **Fecha omitida en la ficha del bloque y origen (D-32 ⟲)**: la **fecha omitida en el bloque**
     se aplica como **presentación** (el DTO `MedioPublico` sí trae `publishedAt`; el contrato no
     cambia). El **origen va TAMBIÉN en la lista del panel** (`OrigenBadge` + fecha de importación
     `createdAt` en la fila, y `FuentesList` como detalle): FR-019 lo pide literalmente y así lo
     decidió el humano en la fase 5 (**A2**), de modo que la parte de D-32 que lo relegaba a los
     diálogos queda **matizada** (los diálogos siguen mostrándolo).
   - **Unión y resoluciones (D-8/D-26/D-27/D-28)**: ocultar/mostrar directos; «Unir» con **doble
     confirmación** (`UnirDuplicadosDialog` + `ConfirmDialog` danger «Esto no se puede deshacer.»);
     «No son el mismo sermón» directo (resuelve la marca `possible_duplicate` —cada resolución lleva
     su `mark`: los dos ejes pueden estar activos— con `visibility: visible` en **ambas**
     fichas del par); «Mostrar de nuevo» desde «No disponible» con `ConfirmDialog` suave.
   - **Búsqueda (D-16/D-17/D-22)**: sin confirmación previa, botón en loading «Buscando…» y
     acciones de fila deshabilitadas mientras corre; el resumen permanece hasta la próxima búsqueda
     (no es toast) y nombra los servicios como los llama la iglesia («YouTube», «el podcast»).
   - **Bilingüe del bloque**: `usePortadaMedios` consume `GET /api/v1/portada/medios?lang=` con el
     mismo fallback del servidor; el cliente **nunca** reimplementa el fallback (patrón P7-9).

## Leyenda

| Marca | Significado |
|---|---|
| `[backend]` `[frontend]` `[db]` `[infra]` | Capa y subagente responsable (`[backend]`/`[db]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` → `devops`) |
| `[P]` | **Paralelizable** con las otras tareas de su grupo (archivos distintos y sin dependencia entre ellas; ver "Grupos de paralelismo") |
| *(sin `[P]`)* | Secuencial: tiene dependencias previas que deben estar integradas |
| *(FR-…, SC-…, US…)* | Requisitos de `spec.md` que cubre la tarea |

> **Orden de capas**: migraciones (`[db]`) → contrato → `platform/` → dominio puro (título/feeds/
> probadores/carátulas) → repository → service → handler/routes → frontend (api/hooks → i18n →
> permisos → página/navegación/bloque → panel) → e2e → cierre. Cada capa se construye y prueba con lo
> que necesita la siguiente (el service con fakes, el handler con service falso), manteniendo la
> dirección de dependencias **`handler → service → repository`** de §II. La **auditoría** se escribe
> junto a cada mutación porque FR-022 exige que la operación y su registro sean atómicos. Las
> migraciones van **primero** (`000014` → `000015`, ninguna aplicada se edita) y **no hay migración de
> permisos** (P9-15). La **red saliente** (feeds, carátulas, probadores) vive solo en `internal/medios`
> tras interfaces falsables (T908–T910) y la **fábrica única de cliente HTTP** (T907A, P9-20), y se
> revisa con lupa de seguridad (RG9-4).

## Tabla resumen

| ID | Tarea | Capa | Fase | Paralelo | Depende de |
|---|---|---|---|---|---|
| T901 | Migración `000014_create_media_content` (`media_items` + `media_episode_sources` + `media_sync_runs`) | `[db]` | 1 | P1 | — |
| T902 | Migración `000015_extend_admin_actions_for_media` (+6 códigos `media.*`) | `[db]` | 1 | P1 | — |
| T903 | Consultas sqlc `media_episodes.sql` + `media_episode_sources.sql` + `media_sync_runs.sql` + rama de `media.sql` | `[db]` | 1 | — | T901, T902 |
| T904 | Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.8.0 → 0.9.0) | `[backend]` | 2 | P1 | — |
| T905 | Regenerar `frontend/src/api/schema.d.ts` | `[frontend]` | 2 | — | T904 |
| T906 | `platform/audit`: +6 códigos `media.*` | `[backend]` | 3 | P1 | — |
| T907 | `medios`: `titulo.go` (normalización + día en `America/Costa_Rica`) | `[backend]` | 4 | P1 | — |
| T907A | `medios`: `cliente_http.go` — fábrica única del `http.Client` saliente (barreras SSRF) | `[backend]` | 4 | P1 | — |
| T908 | `medios`: `feeds.go` (interfaz `FeedSource`, parser Atom/RSS, `testdata/` con fixtures reales) | `[backend]` | 4 | P1 | T907A |
| T909 | `medios`: `disponibilidad.go` (probadores oEmbed/GET a prueba de fallos) | `[backend]` | 4 | P1 | T907A |
| T910 | `medios`: `caratula.go` (descarga acotada + `platform/storage`) | `[backend]` | 4 | P1 | T907A |
| T911 | `medios`: `model.go` + `repository.go` (pgtype, `withTx`, auditoría sobre el tx) | `[backend]` | 5 | — | T903, T906 |
| T912 | `medios`: `repository_episodios.go` (público/portada/panel/mutaciones) | `[backend]` | 5 | P2 | T911 |
| T913 | `medios`: `repository_fuentes.go` + `repository_sync.go` | `[backend]` | 5 | P2 | T911 |
| T914 | Integración: `000014`/`000015` up→down→up, `CHECK`s, idempotencia, `IsFilePublished` y atomicidad | `[backend]` | 5 | — | T912, T913 |
| T915 | `medios`: `service.go` — `Repository`, constantes, validación e invariantes | `[backend]` | 6 | — | T911 |
| T916 | `medios`: `service_public.go` — `GetMedios` (paginación + `?lang` con fallback) | `[backend]` | 6 | — | T915 |
| T917 | `medios`: `service_public.go` — `GetPortadaMedios` (tope 5 en servidor) | `[backend]` | 6 | — | T916 |
| T918 | `medios`: `service_admin.go` — `List` (con `lastSync`) y `Update` (descripción/visibilidad) | `[backend]` | 6 | — | T915 |
| T919 | `medios`: `service_admin.go` — `Merge` (unión por puntero) y `ResolveMark` | `[backend]` | 6 | — | T918 |
| T920 | `medios`: `service_sync.go` — importación idempotente, recorte de descripción, unión/marcado | `[backend]` | 6 | — | T915, T907, T908 |
| T921 | `medios`: `service_sync.go` — ventana del feed, disponibilidad, carátulas, resumen y `media_sync_runs` | `[backend]` | 6 | — | T920, T909, T910 |
| T922 | `medios`: `service_audit.go` — `audit.Recorder` (denegaciones y rechazos) | `[backend]` | 6 | — | T915 |
| T923 | `handler_public.go` — `GET /api/v1/medios` | `[backend]` | 7 | — | T916, T922 |
| T924 | `handler_public.go` — `GET /api/v1/portada/medios` | `[backend]` | 7 | — | T917, T923 |
| T925 | `handler_admin.go` — listado del panel + `PATCH {id}` | `[backend]` | 7 | — | T918, T923 |
| T926 | `handler_admin.go` — `POST buscar-nuevos` (resumen, nunca texto de red) | `[backend]` | 7 | — | T921, T925 |
| T927 | `handler_admin.go` — `POST {id}/unir` y `POST {id}/marcas/resolver` | `[backend]` | 7 | — | T919, T925 |
| T928 | `routes.go` + `cmd/api/main.go` (`AdminChain("medios")`, DI de feeds/probadores/carátulas) + humo | `[backend]` | 7 | — | T923–T927 |
| T929 | `api/medios.ts` + hooks (público, portada, panel) | `[frontend]` | 8 | P3 | T905 |
| T930 | i18n público (§7.2) + `messages.ts` del panel (§7.1) | `[frontend]` | 8 | P3 | — |
| T931 | Permisos y accesos del panel: `MEDIOS`, sello retirado, `AccesosRapidos`, `ACTION_LABELS` +6 | `[frontend]` | 8 | — | T929 |
| T932 | **Actualización acotada**: pruebas de frontend que fijan permisos/accesos (RG9-6) | `[frontend]` | 8 | — | T931 |
| T933 | `/medios`: `MediosPage`, `MediosGrid`, `MedioCard`, `CaratulaNeutra`, `MedioLinks`, `CargarMasButton` + **redirección** (P9-16) | `[frontend]` | 8 | P4 | T929, T930 |
| T934 | `useSiteLinks`: entrada «Medios» condicional (documento de la sección) | `[frontend]` | 8 | P4 | T929, T930 |
| T935 | `MediosBlock` + `MedioItem` en `HomePage` (tras `OfrendasBlock`) | `[frontend]` | 8 | P4 | T929, T930 |
| T936 | Panel: `MediosPage` + `BuscarEpisodios` + `ResumenImportacion` + `FiltroMarcas` + `EpisodiosList` + ruta `/panel/medios` | `[frontend]` | 8 | — | T929, T931 |
| T937 | Panel: `DescripcionForm` + `validation.ts` + `errors.ts` | `[frontend]` | 8 | — | T936 |
| T938 | Panel: `UnirDuplicadosDialog` + `MarcaDeRevision` y resoluciones (`ResolverMarcaDialog`) | `[frontend]` | 8 | — | T936 |
| T939 | `mediosSeed.ts` (siembra **por SQL**, patrón nuevo RG9-7) | `[frontend]` | 9 | — | T901, T902, T914 |
| T940 | E2E Playwright `medios-publica.spec.ts` (US1, US4, US5; 0 terceros, redirección) | `[frontend]` | 9 | P5 | T933, T934, T939 |
| T941 | E2E Playwright `medios-panel.spec.ts` (US2, US3; unir, marcas, permisos y auditoría) | `[frontend]` | 9 | P5 | T936–T939 |
| T942 | E2E Playwright `medios-accesibilidad.spec.ts` (teclado y foco, SC-008) | `[frontend]` | 9 | P5 | T933, T935, T939 |
| T943 | E2E Playwright `portada-medios.spec.ts` (US6, SC-015) | `[frontend]` | 9 | P5 | T935, T939 |
| T944 | Verificación de cierre: `make ci`, cobertura ≥ 80 %, up→down→up, e2e por tandas, 0 dependencias nuevas | `[infra]` | 10 | — | todas |
| T945 | Documentación y PR (documentador + devops) | `[infra]` | 10 | — | T944 |

**Total: 46 tareas** — 25 `[backend]` · 16 `[frontend]` · 3 `[db]` · 2 `[infra]` · 20 marcadas `[P]`
(**T907A** se añadió en la fase 5 — corrección A1 del `analyze.md` — con letra en el ID para no
renumerar T908–T945; se ejecuta al inicio de la Fase 4, antes de T908–T910).

## Grupos de paralelismo

| Grupo | Tareas | Por qué pueden ir en paralelo |
|---|---|---|
| P1 | T901 · T902 · T904 · T906 · T907 · T907A (oleada 1) y T908 · T909 · T910 (oleada 2) | Dos oleadas de artefactos independientes entre sí: la 1 son dos migraciones (`000014`/`000015`, archivos distintos — el orden numérico solo importa al aplicar con `make db-migrate` —), el contrato (`backend/api/openapi.yaml`), `platform/audit`, las funciones puras de `titulo.go` y la **fábrica única de cliente HTTP** (`cliente_http.go`, corrección A1: sus barreras las consumen los tres de red saliente); la 2 son `feeds.go` (+`testdata/`), `disponibilidad.go` y `caratula.go`, paralelos entre sí pero **dependientes de T907A** (usan su `http.Client` y comparten su tabla de pruebas). `feeds.go`/`disponibilidad.go`/`caratula.go` definen sus interfaces y sus tipos de entrada/salida en su propio archivo (M1: T915 no los redeclara) |
| P2 | T912 · T913 | `repository_episodios.go` y `repository_fuentes.go`/`repository_sync.go` son archivos distintos que parten de T911 |
| P3 | T929 · T930 | El cliente `api/medios.ts` + hooks y el i18n/`messages.ts` no se solapan (T929 depende solo de T905; T930 no depende de nada) |
| P4 | T933 · T934 · T935 | `features/publico/medios/` (página y fichas), `hooks/useSiteLinks.ts` y `MediosBlock`/`MedioItem` + `HomePage` son archivos distintos que parten de T929 **y de T930** (claves i18n tipadas) |
| P5 | T940 · T941 · T942 · T943 | Cuatro specs e2e independientes (`medios-publica`, `medios-panel`, `medios-accesibilidad`, `portada-medios`) sobre la misma base (`mediosSeed.ts`, T939). **T939 ya no depende de T936–T938** (corrección M3 de la fase 5: siembra por SQL, sin componentes de panel), de modo que este grupo arranca al terminar las páginas de T933–T935; solo **T941** necesita además el panel (T936–T938) |

---

## Fase 1 — Migraciones y capa de datos (`[db]`)

> Convenciones de la skill `postgres-db` y de `data-model.md` (fuente de verdad): `id UUID`,
> `created_at`/`updated_at`, índices y `CHECK`/`UNIQUE` **en la base**, `up`/`down` completos, SQL
> **parametrizado** (§IV, CWE-89) y **nunca** se edita una migración aplicada (§VI). Numeración sin
> huecos desde `000001` (F7 terminó en `000013`). **Sin migración de permisos** (P9-15: `medios` ya
> está sembrado por `000002` y el catálogo sigue en 10).

- [X] T901 · Migración `000014_create_media_content` · `[db]` `[P]` (P1)

  - **Archivos**: `backend/migrations/000014_create_media_content.up.sql` y
    `000014_create_media_content.down.sql` (NUEVOS).
  - **Qué hace**: crea **`media_items`**, **`media_episode_sources`** y **`media_sync_runs`** tal
    como `data-model.md`: `media_items` con `title` (1…300, del feed, no editable), `published_at`
    (TIMESTAMPTZ; el día se calcula en `APP_TZ`, FR-006), `description_es`/`description_en`
    (1…300, NULL solo cuando el feed no trae descripción — precisión 3 —), `description_edited`,
    `cover_file` (copia de la carátula, `img_<uuid>.<ext>`), `origin` (`video`/`audio`/`both`),
    **`visibility`** (`visible`/`hidden`) y las marcas en **dos ejes** (P9-19, corrección C1 de la
    fase 5): **duplicados** —`duplicate_of_id` (par dudoso, FK a `media_items`) +
    `duplicate_mark_resolved_at`— y **disponibilidad** —`availability_checked_at`/
    `availability_result` (`available`/`unavailable`/`inconclusive`) + `availability_mark`
    (`unavailable`/`missing_from_feed`) + `availability_mark_resolved_at`—, más `merged_into_id`
    (unión por puntero, FK a `media_items`). **CHECKs**: `description_en` exige `description_es`;
    `duplicate_mark_resolved_at` exige `duplicate_of_id`; `availability_mark_resolved_at` exige
    `availability_mark`; **sin** `CHECK` de simetría entre marca y par (a propósito, C1: la pareja
    recíproca es invariante de service — T914/T920 —); `btrim` en
    `title`/`description_es`/`description_en`; catálogos de `origin`/`visibility`/`availability_mark`/
    `availability_result` y longitudes. `media_episode_sources` con `item_id` FK, `provider`
    (`youtube`/`podcast`), `external_id` (1…120), `feed_title` (procedencia), `url` (1…500) y
    **`UNIQUE (provider, external_id)`** (idempotencia, R9-11) + `CHECK` de forma de URL por
    proveedor (`^https://` + `youtube.com/watch?v=<11>` / `podcasters.spotify.com/`). `media_sync_runs`
    con `actor_user_id` (FK a `users`), `started_at`/`finished_at` (`finished_at >= started_at`),
    conteos `>= 0` (`imported/joined/marked/hidden/availability_checked/availability_inconclusive`;
    `marked_count` cuenta **solo** «posible duplicado» — criterio B2 de la fase 5 —)
    y `youtube_status`/`podcast_status` (`ok`/`failed`, FR-016). **Índices**:
    `media_items_order_idx (published_at, created_at, id)`, `media_items_visibility_idx`,
    `media_items_duplicate_of_idx`, `media_items_availability_mark_idx`,
    `media_items_merged_into_idx`,
    `media_episode_sources_item_idx`, `media_episode_sources_provider_idx (provider, external_id)` y
    `media_sync_runs_started_at_idx (started_at DESC, id)`. El `down` tira las tres tablas.
    *Cubre FR-001, FR-012, FR-013, FR-014, FR-015, FR-017, FR-018, FR-020 y el modelo de FR-006/
    FR-009/FR-023/FR-025 a nivel de datos.*
  - **Pruebas incluidas** (§III): — (SQL; su verificación ejecutable es T914). **Cómo se verifica**:
    `make db-migrate` aplica sin errores sobre la BD de F7 (`000013` → `000014`); `migrate … down`
    revierte sin errores; los `CHECK`/`DEFAULT`/índices de `data-model.md` están literalmente en el
    SQL; ninguna tabla sin `id`/`created_at`/`updated_at`; sin `sort_order` (el orden es por fecha,
    default #11) y sin `deleted_at` (fuera de alcance).
  - **Criterio de terminado**: el esquema coincide con `data-model.md` (lo revisa `revisor-codigo`
    contra la skill `postgres-db`) y los invariantes de idempotencia y de marcas **por eje** están
    garantizados por la BD (la pareja recíproca del par la garantiza el service: T914/T920).
  - **Commit sugerido**: `feat(db): tablas de medios con fuentes por proveedor y registro de busquedas (000014)`

- [X] T902 · Migración `000015_extend_admin_actions_for_media` · `[db]` `[P]` (P1)

  - **Archivos**: `backend/migrations/000015_extend_admin_actions_for_media.up.sql` y
    `.down.sql` (NUEVOS).
  - **Qué hace**: amplía el registro de auditoría de F2 (FR-022, P9-14): sustituye
    `admin_actions_action_check` (nombre propio desde `000006`, verificado en F3/F4/F6/F7)
    añadiendo los **6 códigos** `media.sync`, `media.description.update`, `media.hide`, `media.show`,
    `media.merge` y `media.mark.resolve` (F9 no tiene `create`/`delete`: no hay alta manual ni
    borrado — FR-019 —). El `target_kind` **no cambia** (las fichas son `content`). El `down` elimina
    las filas con códigos F9 y restaura el CHECK de F7 (destructivo a propósito, igual que los `down`
    de `000006`/`000008`/`000011`/`000013`; documentado en `data-model.md`). *Cubre FR-022* a nivel
    de datos.
  - **Pruebas incluidas** (§III): — (SQL; su verificación ejecutable es T914). **Cómo se verifica**:
    `make db-migrate` aplica; los 6 códigos del `CHECK` coinciden uno a uno con la lista de
    `data-model.md` y con `platform/audit` (T906); el registro de acciones queda **cerrado** en la BD.
  - **Criterio de terminado**: registro cerrado garantizado por la BD, sin tocar ninguna migración
    anterior (§VI).
  - **Commit sugerido**: `feat(db): ampliar admin_actions a medios (000015)`

- [X] T903 · Consultas sqlc de `medios` + rama de carátulas en `media.sql` · `[db]`

  - **Archivos**: `backend/internal/db/queries/media_episodes.sql` (NUEVO),
    `media_episode_sources.sql` (NUEVO), `media_sync_runs.sql` (NUEVO), `media.sql` (EDITADO: rama
    aditiva de `IsFilePublished`) y `backend/internal/db/*.go` (GENERADOS y commiteados:
    `make sqlc-gen`). *No se toca `queries/audit.sql` (se reutiliza `InsertAdminAction`, R7-10).*
  - **Qué hace**: escribe las consultas nombradas de `data-model.md`. En `media_episodes.sql`:
    `ListPublishedMediaItems(:limit, :offset)` —**única vía pública** de fichas: predicado completo
    de R9-7 (`visibility='visible' AND merged_into_id IS NULL AND (duplicate_of_id IS NULL OR
    duplicate_mark_resolved_at IS NOT NULL)`), `ORDER BY published_at DESC, created_at
    ASC, id`, `LIMIT/OFFSET`—; `CountPublishedMediaItems`; `ListPortadaMediaItems(:limit)` (mismo
    predicado; el service pasa 5); `ListMediaItems(:limit, :offset, sqlc.narg('visibility'),
    sqlc.narg('review_mark'), sqlc.narg('include_merged'))` y `CountMediaItems` (panel paginado, por
    defecto **sin** absorbidas, mismo orden que el visitante); `GetMediaItemByID`; `InsertMediaItem`/
    `UpdateMediaItem` (el `Update` escribe la fila completa ya fusionada por el service, patrón
    `UpdateDonationMethod` de F7); `RepointMediaItemSources(:from_item_id, :to_item_id)` (unión) y
    `ListMediaItemSourcesByItem(:item_id)`. En `media_episode_sources.sql`:
    `InsertMediaEpisodeSource` con **`ON CONFLICT (provider, external_id) DO NOTHING`** (devuelve la
    fila insertada o vacío si ya existía), `ListMediaEpisodeSourcesByProvider(:provider)` (ventana
    del feed, FR-017) y `GetMediaEpisodeSource(:provider, :external_id)`. En `media_sync_runs.sql`:
    `InsertMediaSyncRun` y `GetLastMediaSyncRun` (`ORDER BY started_at DESC, id LIMIT 1`). En
    `media.sql`, `IsFilePublished` crece con la rama de F9 (la carátula de una ficha **publicable**
    se sirve; oculta/absorbida/con par abierto **no**, sin borrar el archivo). Sin `SELECT *`,
    `ORDER BY` determinista (§8.1.7), todo parametrizado (§IV, CWE-89); **ninguna escritura sobre
    `admin_actions`**. *Cubre FR-002, FR-006, FR-009, FR-012 y FR-027* a nivel de acceso a datos.
  - **Pruebas incluidas** (§III): — (SQL anotado; su ejecución es T912–T914). **Cómo se verifica**:
    `make sqlc-gen` genera sin errores y `make sqlc-verify` no detecta deriva; los tipos generados
    quedan commiteados; el predicado público está **en la consulta** (no solo en el service) y el
    `ORDER BY` es literalmente `published_at DESC, created_at ASC, id`; ninguna consulta de escritura
    sobre `admin_actions`.
  - **Criterio de terminado**: la capa generada compila y es la única fuente de SQL del dominio
    `medios` (T911–T913).
  - **Commit sugerido**: `feat(db): consultas sqlc de medios y rama de caratulas en IsFilePublished`

---

## Fase 2 — Contrato OpenAPI y tipos

- [X] T904 · Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.9.0) · `[backend]` `[P]` (P1)

  - **Archivos**: `backend/api/openapi.yaml` (EDITADO). *No* se toca
    `specs/009-medios-predicas-y-podcasts/contracts/openapi.yaml` (snapshot inmutable).
  - **Qué hace**: funde **aditivamente** el delta aprobado: las **7** operaciones de F9
    (`GET /api/v1/medios?lang=&limit=&offset=`, `GET /api/v1/portada/medios?lang=`, y las 5 de panel
    `GET /api/v1/admin/medios`, `POST …/buscar-nuevos`, `PATCH …/{id}`, `POST …/{id}/unir`,
    `POST …/{id}/marcas/resolver`) y sus schemas (`MediosPublicos`, `MedioPublico`, `PortadaMedios`,
    `MedioAdminList`, `MedioAdmin`, `FuenteAdmin`, `MedioPatch`, `UnirInput`, `ResolverMarcaInput`,
    `ResumenBusqueda`, `SyncRunAdmin`, `Proveedor`, `OrigenMedio`, `Visibilidad`, `MarcaRevision`,
    `MarcaDisponibilidad`,
    `ResultadoDisponibilidad`, `EstadoFeed`, parámetro `MedioItemId`), reutilizando
    `ErrorEnvelope`/`ErrorBody`, las respuestas reutilizables, los parámetros `Limit`/`Offset`
    **reales** del contrato vivo y `sessionCookie`. Aplica los cambios documentados del delta: enum
    `action` de `AdminActionItem` **+ 6 códigos** `media.*`; la prosa del **catálogo de permisos NO
    cambia** (sigue en 10; `medios` ya estaba sembrado por `000002`); `GET /api/v1/media/{fileName}`
    solo crece en prosa (F9 añade la rama de `IsFilePublished`). Documenta el límite por defecto
    **12** y tope **50** de la sección, el **tope 5 del bloque resuelto en servidor** (sin parámetro
    de límite), que `items`/`description`/`coverUrl`/`videoUrl`/`audioUrl` se **omiten** cuando no
    aplican (la redirección de `/medios` es del frontend: la operación sigue siendo `200`), que
    `buscar-nuevos` responde `200` con el resumen **también cuando un feed falló** (no es error) y
    que puede tardar (presupuesto 90 s). Sube `info.version` de **0.8.0 a 0.9.0**. El registro de
    `error.code` **no cambia** (F9 emite `invalid`, `unauthenticated`, `forbidden`, `not_found`,
    `conflict`, `internal`). *Cubre el contrato de todos los FR*; garantiza FR-019 (el `PATCH` solo
    acepta `descriptionEs`/`descriptionEn`/`visibility`, `additionalProperties: false`), FR-021 (el
    permiso `medios` citado en las descripciones de panel) y FR-022 (los 6 códigos).
  - **Pruebas incluidas** (§III): — (artefacto de contrato). **Cómo se verifica**: el YAML parsea
    sin errores y **sin `$ref` sin resolver** (validador de `make api-gen` o búsqueda de `$ref`
    huérfanos — el simple «el YAML parsea» no lo detecta —); `/api/v1/medios` y
    `/api/v1/portada/medios` tienen `security: []`, `Cache-Control: no-store` documentado y los
    schemas públicos **sin** `visibility`/marcas/`sources`; `PortadaMedios.items` con
    `maxItems: 5`; las operaciones de panel citan `sessionCookie` y el permiso `medios`;
    `grep -n "version:" backend/api/openapi.yaml` → `0.9.0`; `make api-gen` completa sin errores
    **solo como validación** (corrección B1 de la fase 5): el generado
    `frontend/src/api/schema.d.ts` **NO se commitea aquí** — es el entregable de T905 —; si `api-gen`
    lo reescribe, se descarta antes del commit de esta tarea (`git checkout -- frontend/src/api/schema.d.ts`),
    de modo que cada commit queda limpio y T905 no queda vacío.
  - **Criterio de terminado**: el contrato vivo contiene F1–F4 + F6 + F7 + F9 y es la única fuente de
    tipos del frontend (T905).
  - **Commit sugerido**: `docs(api): fusionar delta de F9 en el contrato vivo (0.9.0)`

- [ ] T905 · Regenerar `frontend/src/api/schema.d.ts` · `[frontend]`

  - **Archivos**: `frontend/src/api/schema.d.ts` (GENERADO y commiteado).
  - **Qué hace**: ejecuta `make api-gen` contra **`backend/api/openapi.yaml`** (nunca contra el
    snapshot de `specs/`) y commitea los tipos de F9 (`MediosPublicos`, `MedioPublico`,
    `PortadaMedios`, `MedioAdmin`/`MedioPatch`/`UnirInput`/`ResolverMarcaInput`, `ResumenBusqueda`,
    `SyncRunAdmin`, `Visibilidad`, `MarcaRevision`, `MarcaDisponibilidad`, `ResultadoDisponibilidad`,
    y `AdminActionItem` con el enum ampliado a los 6 códigos `media.*`). **Aquí** es donde se
    commitea el generado (corrección B1 de la fase 5: T904 lo valida pero no lo pisa), así que esta
    tarea es además la **comprobación de deriva** del contrato.
  - **Pruebas incluidas** (§III): — (artefacto generado). **Cómo se verifica**: regenerar no produce
    `git diff`; `npm run typecheck` pasa con los tipos nuevos.
  - **Criterio de terminado**: `schema.d.ts` reproducible desde el contrato y listo para
    `api/medios.ts` (T929; regla anti-deriva de §8.1.5).
  - **Commit sugerido**: `build(frontend): regenerar schema.d.ts con los tipos de F9`

---

## Fase 3 — Núcleo de plataforma (`backend/internal/platform/`)

- [X] T906 · `platform/audit`: 6 códigos `media.*` · `[backend]` `[P]` (P1)

  - **Archivos**: `backend/internal/platform/audit/audit.go` y `audit_test.go` (EDITADOS).
  - **Qué hace**: amplía el registro cerrado del plumbing de F2 (P9-14) con los **6 códigos**
    `media.sync|description.update|hide|show|merge|mark.resolve` en `ActionCodes` y
    `ValidActionCode`. `TargetKindContent = "content"` **no cambia** (ya lo añadió F3). *Cubre
    FR-022.*
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/platform/audit/` en verde: cada código `media.*` pasa `Validate` con objetivo
    `content` y `TargetLabel`; los códigos de F2/F3/F4/F6/F7 siguen intactos (regresión); un código
    fuera de la lista falla.
  - **Criterio de terminado**: la lista cerrada coincide literalmente con el `CHECK` de la migración
    `000015` (T902) — lo comprueba la integración de T914.
  - **Commit sugerido**: `feat(audit): codigos media.* para F9`

---

## Fase 4 — Dominio `medios`: título, feeds y red saliente acotada

> Piezas **puras y falsables** (research R9-3/R9-4/R9-8): sin HTTP real en las pruebas (fakes de
> `http.RoundTripper` y fixtures XML), sin dependencias nuevas. Es la superficie sensible de F9:
> **toda la red saliente del backend vive aquí**, detrás de interfaces.

- [X] T907 · `medios`: `titulo.go` — normalización de títulos y día en `America/Costa_Rica` · `[backend]` `[P]` (P1)

  - **Archivos**: `backend/internal/medios/titulo.go` y `titulo_test.go` (NUEVOS).
  - **Qué hace**: funciones **puras** (R9-4, P9-5): `normalizarTitulo` — NFKD + quitar marcas
    diacríticas con `golang.org/x/text/unicode/norm` (**sin dependencia nueva**: `x/text v0.42.0` ya
    está en `go.mod` como indirecta de `pgx`; se promueve a directa sin cambiar versiones),
    mayúsculas Unicode, puntuación y símbolos → espacio (`unicode.IsPunct`/`IsSymbol`, incluidos
    `¿ ¡ – — …`), colapsar espacios y recortar — y `diaEnLugar` — fecha → día calendario en la zona
    `APP_TZ` (`America/Costa_Rica`): la llave de «mismo día» de la unión (FR-013). El texto original
    **nunca se altera** (se publica el título del feed tal cual, FR-003): la normalización es solo la
    llave de comparación. *Cubre FR-013 (mitad de la regla) y FR-006 (día en la zona de la iglesia)*.
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/medios/` en verde: el **par real** de títulos normaliza igual
    («Potencial sin Resultados La Lección de la Higuera» / «Potencial sin Resultados: La Lección de
    la Higuera» → `POTENCIAL SIN RESULTADOS LA LECCION DE LA HIGUERA`); «Lección» = «Leccion»;
    tildes/`ñ`/`¿¡`/emojis íntegros en el original y neutros en la llave; espacios dobles y
    puntuación colapsados; frontera de día real `Fri, 03 Jul 2026 01:22:48 GMT` → **2026-07-02** en
    `America/Costa_Rica` (y distinto si se calculara en UTC: la prueba fija por qué).
  - **Criterio de terminado**: la regla de comparación tiene **una** implementación pura y sus casos
    de prueba; `golangci-lint` sin comentarios.
  - **Commit sugerido**: `feat(medios): normalizacion de titulos y dia en la zona de la iglesia`

- [X] T907A · `medios`: `cliente_http.go` — fábrica única del `http.Client` saliente · `[backend]` `[P]` (P1) *(red saliente)*

  - **Archivos**: `backend/internal/medios/cliente_http.go` y `cliente_http_test.go` (NUEVOS).
  - **⚠️ ID con letra**: tarea **añadida en la fase 5** (corrección **A1** del `analyze.md`); la
    letra evita renumerar T908–T945. Es la **primera** de la Fase 4 y de ella dependen T908–T910.
  - **Qué hace**: la **fábrica única** del `http.Client` con el que el dominio sale a internet
    (P9-20/R9-8: primera red saliente del backend, RG9-4). Cuatro barreras aplicadas **por
    construcción** a todo consumidor: (1) solo **`https`** y host de la **allowlist del consumidor**
    (feeds: los hosts de FR-010; probador: el host de oEmbed y el del propio enlace ya validado;
    carátula: `i.ytimg.com`/`d3t3ozftmdmh3i.cloudfront.net`); (2) **sin IPs privadas/loopback/
    link-local** —incluido `169.254.169.254`—: el dialer resuelve DNS y rechaza el destino en
    **cada** conexión, también tras cada redirect; (3) **`CheckRedirect`** que valida cada
    `Location` contra la allowlist del consumidor (fuera de ella → error; además `net/http` no sigue
    más de 10); (4) **`io.LimitReader` en todo cuerpo leído** con tope por consumidor (feeds y
    probador: constantes revisables del dominio; carátula: `UPLOAD_MAX_BYTES`) — nunca se lee un
    cuerpo entero sin tope —. Además: **`User-Agent` propio** en todas las peticiones y timeouts de
    conexión/respuesta coherentes con los 10/10/5 s de P9-3 (más el `context` con presupuesto de la
    búsqueda). *Cubre la mitigación SSRF de RG9-4 sobre FR-010/FR-017 y es la base de T908–T910*.
  - **Pruebas incluidas** (§III): unitarias con fake `http.RoundTripper` (**sin red**). **Cómo se
    verifica**: `go test ./internal/medios/` en verde con la **tabla compartida de barreras** (la
    misma que T908/T909/T910 ejecutan sobre cada consumidor): redirect a host **ajeno** rechazado;
    redirect a host **de la allowlist** seguido; `http://` rechazado; IP privada/loopback/metadata
    (`169.254.169.254`) rechazada tras resolver DNS; cuerpo por encima del **tope** cortado sin
    leerse entero; `User-Agent` presente; petición que no responde → error con timeout (nunca
    cuelga).
  - **Criterio de terminado**: las cuatro barreras viven en **un** sitio (imposible de olvidar en un
    consumidor nuevo) y su tabla de pruebas está lista para reutilizarse en T908/T909/T910.
  - **Commit sugerido**: `feat(medios): fabrica unica del cliente HTTP saliente (barreras SSRF)`

- [X] T908 · `medios`: `feeds.go` — `FeedSource` + parser Atom/RSS + fixtures reales · `[backend]` `[P]` (P1) *(red saliente)*

  - **Archivos**: `backend/internal/medios/feeds.go`, `feeds_test.go`,
    `fake_feed_source_test.go` (NUEVO) y `backend/internal/medios/testdata/` (NUEVO: los XML
    capturados en `quickstart.md` §0.1 — `youtube_feed.xml`, `podcast_feed.xml` — más casos
    patológicos).
  - **Qué hace**: la interfaz del dominio **`FeedSource`** (`FetchYouTube(ctx)` y
    `FetchPodcast(ctx)`; doble falso para los dos orígenes) — **declarada aquí, con sus tipos
    (`EntradaFeed`); T915 no la redeclara (M1)** — y su implementación: el **cliente de la fábrica
    única** (T907A/P9-20: https + allowlist de hosts, sin IPs privadas, redirects validados,
    `LimitReader`, `User-Agent` propio) contra las **dos URLs fijas de FR-010** (constantes del
    dominio, P9-18) y **parser propio con `encoding/xml`** (R9-3) — Atom de YouTube
    (`yt:videoId`, `title`, `link`, `published` RFC3339, `media:thumbnail`, `media:description`) y
    RSS 2.0 del podcast (`title`, `description` con **HTML dentro de CDATA**, `link`, `guid`,
    `pubDate` RFC1123 con `GMT`, `itunes:image`) → `[]EntradaFeed` (id externo, título, enlace
    **validado contra el host esperado del proveedor**, fecha, URL de carátula, descripción **a texto
    plano**). Una fecha imposible de parsear **descarta el item con aviso** (nunca inventa fecha); un
    enlace con host ajeno **descarta el item con aviso**; XML roto → **error claro** (nunca items a
    medias). La captura de fixtures (curl de `quickstart.md` §0.1) se hace desde un entorno donde los
    feeds respondan — el de YouTube puede devolver 404 desde IPs de datacenter (RG9-1) — y se
    commitean como datos de prueba. *Cubre FR-010 y FR-023 (texto plano, datos nunca código)*.
  - **Pruebas incluidas** (§III): unitarias sobre **fixtures**. **Cómo se verifica**:
    `go test ./internal/medios/` en verde: parseo de los dos fixtures reales (9 y 9 elementos con el
    par de títulos que la unión debe resolver); campos ausentes → defaults documentados (sin carátula
    → `""`; sin descripción → ausente); CDATA y HTML de la descripción → texto plano sin etiquetas y
    con entidades decodificadas; tildes/`ñ`/`¿¡`/emojis intactos; fechas `GMT` y `+0000` parseadas;
    enlace de host ajeno → item rechazado con aviso; XML roto → error claro sin items parciales.
    **Sin red en ninguna prueba** (el `http.Client` se inyecta) y la **tabla de barreras de T907A**
    aplicada al lector de feeds (corrección A1: redirect a host ajeno, IP privada, cuerpo > tope).
  - **Criterio de terminado**: dos formatos fijos parseados con stdlib, detrás de interfaz falsable,
    con los fixtures reales commiteados; el lector de feeds pasa la tabla de barreras de la red
    saliente (A1); 0 dependencias nuevas.
  - **Commit sugerido**: `feat(medios): lector de feeds Atom/RSS con fixtures reales`

- [X] T909 · `medios`: `disponibilidad.go` — probadores a prueba de fallos · `[backend]` `[P]` (P1) *(red saliente)*

  - **Archivos**: `backend/internal/medios/disponibilidad.go` y `disponibilidad_test.go` (NUEVOS).
  - **Qué hace**: la interfaz **`AvailabilityProber`** con sus tipos de resultado (**declarada aquí;
    T915 no la redeclara, M1**) y dos implementaciones gratuitas y sin
    claves (R9-2, P9-6), ambas sobre el **cliente de la fábrica única** (T907A/P9-20 — el enlace del
    episodio lo trae el feed, así que las barreras de red son obligatorias: corrección A1 —):
    **YouTube** → `GET https://www.youtube.com/oembed?url=<video>&format=json`
    (`200` → `available`; **`404` → `unavailable` confirmado**; `400`/`401`/`403`/`429`/`5xx`/
    timeout/red → `inconclusive`) y **Podcast** → `GET` del **enlace del episodio** devuelto por el
    feed (`200` → `available`; **`404`/`410` → `unavailable` confirmado**; resto → `inconclusive`;
    el oEmbed de Spotify se descarta: 504 verificado con nuestros enlaces). Timeout 5 s por
    petición. La regla es **fail-safe**: **solo el confirmado dice `unavailable`**; `inconclusive`
    **nunca** oculta nada (FR-017/SC-014). *Cubre FR-017 y SC-014* (capa de red saliente).
  - **Pruebas incluidas** (§III): unitarias con fake `http.RoundTripper`. **Cómo se verifica**:
    `go test ./internal/medios/` en verde con la **tabla de casos por probador**: YouTube
    `200`→available, `404`→unavailable, `400`/`500`/timeout→inconclusive; podcast `200`→available,
    `404`/`410`→unavailable, `504`/timeout→inconclusive; error de red → inconclusive. Se comprueba
    además que **solo** `404`/`410` producen `unavailable` (la tabla lo fija: nada más oculta) y que
    el probador pasa la **tabla de barreras de T907A** (corrección A1: redirect a host ajeno, IP
    privada, cuerpo > tope).
  - **Criterio de terminado**: la regla a prueba de fallos de FR-017 está fijada por prueba, el
    probador no hace I/O real en tests y sus peticiones salen **solo** por la fábrica única (A1).
  - **Commit sugerido**: `feat(medios): probadores de disponibilidad a prueba de fallos`

- [X] T910 · `medios`: `caratula.go` — descarga acotada de carátulas con `platform/storage` · `[backend]` `[P]` (P1) *(red saliente)*

  - **Archivos**: `backend/internal/medios/caratula.go` y `caratula_test.go` (NUEVOS).
  - **Qué hace**: la interfaz **`CoverFetcher`** (**declarada aquí; T915 no la redeclara, M1**) con
    la descarga **acotada** de R9-8/P9-7 (red saliente del backend, RG9-4): las cuatro barreras las
    aplica el **cliente de la fábrica única** (T907A/P9-20, corrección A1) —solo **`https`** y host
    de **allowlist** (`i.ytimg.com`, `d3t3ozftmdmh3i.cloudfront.net`); rechazo de **IPs
    privadas/loopback** tras resolver DNS; **sin seguir redirecciones** fuera de la allowlist (cada
    `Location` validado); tamaño acotado con `io.LimitReader` (**`UPLOAD_MAX_BYTES`**,
    configuración existente)— y esta tarea añade la **firma binaria** con
    `storage.Detect` (solo JPEG/PNG/WebP: SVG, GIF, HTML renombrado → rechazados) dentro de
    `Store.Save` (nombres `img_<uuid>.<ext>` generados por el servidor, escritura atómica); timeout
    10 s y máximo 30 descargas por búsqueda (constantes revisables). **Un fallo de descarga nunca
    tira la importación**: el elemento entra **sin carátula** (marcador neutro, default #4). *Cubre
    FR-004/SC-005 (la carátula se sirve del sitio) y la mitigación SSRF de RG9-4*.
  - **Pruebas incluidas** (§III): unitarias con fake `http.RoundTripper` + `storage.Store` falso.
    **Cómo se verifica**: `go test ./internal/medios/` en verde: host **no permitido** rechazado;
    `http://` rechazado; redirect a host ajeno rechazado; IP privada (`169.254.169.254`, loopback)
    rechazada; HTML/SVG renombrado rechazado por `storage.Detect`; más de `UPLOAD_MAX_BYTES`
    rechazado (sin leer todo el cuerpo); descarga correcta → `Save` con **bytes intactos** y nombre
    generado; fallo de descarga → ficha sin carátula (nunca aborta). La **tabla de barreras de
    T907A** se ejecuta también sobre la descarga de carátulas (corrección A1: la misma tabla en los
    tres consumidores).
  - **Criterio de terminado**: la descarga de carátulas sale **solo** por la fábrica única (sus
    cuatro barreras, P9-20) y la firma binaria tiene su prueba; cada barrera queda cubierta.
  - **Commit sugerido**: `feat(medios): descarga acotada de caratulas con platform/storage`

---

## Fase 5 — Dominio `medios`: repository (integración con `000014`/`000015`)

- [X] T911 · `medios`: `model.go` + `repository.go` · `[backend]`

  - **Archivos**: `backend/internal/medios/model.go`, `repository.go` y `repository_test.go`
    (NUEVOS).
  - **Qué hace**: entidades del dominio (`Episodio`, `FuenteEpisodio`, `SyncRun` con sus enums
    `Visibilidad`/`MarcaDisponibilidad`/`ResultadoDisponibilidad`/`Proveedor`/`OrigenMedio`) y los **DTOs**
    del dominio (etiquetas `validate`; sufijos `Es`/`En` **solo** en la descripción — FR-025/FR-026
    —; el DTO público se compone en `service_public.go`, un solo lugar, sin `visibility`/
    `reviewMark`/`sources`; las marcas viajan en sus campos por eje solo en el DTO de panel). Base del repository: constructor sobre `*pgxpool.Pool`, mapRow/params
    `pgtype` → dominio (§8.1.6), `withTx` (patrón de F2) y la **inserción de `admin_actions` dentro
    del `tx`** vía `InsertAdminAction` (mapeo `audit.Action` → `InsertAdminActionParams`; sin que un
    dominio importe a otro, R2). *Soporta FR-001/FR-003 (procedencia) y FR-022 (registro atómico)*.
  - **Pruebas incluidas** (§III): unitarias de los mapeos (`pgtype` ↔ dominio, nil/`Valid`) y del
    mapeo de acciones (los 6 códigos `media.*`). **Cómo se verifica**: `go test ./internal/medios/`
    en verde; `gofmt`/`go vet` limpios; ningún `pgtype` sale de `repository.go` (R4).
  - **Criterio de terminado**: el paquete compila con sus tipos y el tx de escritura queda listo para
    T912/T913.
  - **Commit sugerido**: `feat(medios): modelo y base del repository con auditoria transaccional`

- [X] T912 · `medios`: `repository_episodios.go` (fichas: público, portada, panel y mutaciones) · `[backend]` `[P]` (P2)

  - **Archivos**: `backend/internal/medios/repository_episodios.go` y
    `repository_episodios_integration_test.go` (NUEVOS; `//go:build integration`).
  - **Qué hace**: `ListPublishedMediaItems`/`CountPublishedMediaItems` (única vía pública: predicado
    de R9-7 — incluida la síntesis de la precisión 1: **`missing_from_feed` sale**,
    **el par abierto** no —), `ListPortadaMediaItems` (LIMIT 5 por el service),
    `ListMediaItems`/`CountMediaItems` (panel paginado con filtros `visibility`/`reviewMark`/
    `includeMerged`, por defecto **sin** absorbidas, mismo orden que el visitante),
    `GetMediaItemByID`, `InsertMediaItem`/`UpdateMediaItem` (fila completa ya fusionada) y
    `RepointMediaItemSources`. Traducción de restricciones a `apperr`. *Cubre FR-002, FR-006,
    FR-009, FR-015, FR-019*.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/medios/` en verde: `ListPublishedMediaItems` **nunca**
    devuelve `hidden`, absorbidas ni un **par abierto**; **sí** devuelve
    `missing_from_feed` sin resolver (FR-017) y el par **resuelto**; el orden es
    `published_at DESC, created_at ASC, id` con fechas iguales y el `LIMIT/OFFSET` de las tandas
    funciona (12 por defecto); `ListPortadaMediaItems` devuelve exactamente 5 con 8 publicables y
    exactos con menos; `RepointMediaItemSources` mueve las fuentes y la resultante queda con las dos;
    el `Update` deja exactamente la fila completa que recibe del service.
  - **Criterio de terminado**: la regla «0 no publicables expuestos» está garantizada **en la
    consulta**, no solo en el servicio (SC-002).
  - **Commit sugerido**: `feat(medios): persistencia de fichas con el predicado publico completo`

- [X] T913 · `medios`: `repository_fuentes.go` + `repository_sync.go` · `[backend]` `[P]` (P2)

  - **Archivos**: `backend/internal/medios/repository_fuentes.go`, `repository_sync.go`,
    `repository_fuentes_integration_test.go` y `repository_sync_integration_test.go` (NUEVOS;
    `//go:build integration`).
  - **Qué hace**: `InsertMediaEpisodeSource` (**`ON CONFLICT (provider, external_id) DO NOTHING`**:
    devuelve fila o vacío sin actualizar nada — instantánea R9-6 —), `ListMediaEpisodeSourcesByProvider`
    (detección de «ya no está en la ventana del feed», FR-017), `GetMediaEpisodeSource`,
    `ListMediaItemSourcesByItem` (procedencia en el panel y en la unión) y, en `repository_sync.go`,
    `InsertMediaSyncRun`/`GetLastMediaSyncRun` (FR-018). *Cubre FR-012, FR-017 y FR-018* a nivel de
    persistencia.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/medios/` en verde: el `UNIQUE (provider, external_id)`
    rechaza el doble insert y el `ON CONFLICT` no falla ni duplica (búsqueda repetida y dos
    búsquedas simultáneas → 0 filas nuevas); el `CHECK` de URL por proveedor rechaza
    `javascript:`, host ajeno y `watch?v=` sin 11 caracteres, y una URL de podcast que no sea
    `podcasters.spotify.com`; `feed_title`/`external_id` con espacios sobrantes fallan (`btrim`);
    dos fuentes del **mismo** proveedor no caben en la misma ficha (regla de service + prueba);
    `InsertMediaSyncRun` persiste los conteos y `GetLastMediaSyncRun` devuelve la más reciente.
  - **Criterio de terminado**: la idempotencia de la búsqueda está garantizada **por la base de
    datos** (R9-11), incluso con concurrencia.
  - **Commit sugerido**: `feat(medios): fuentes por proveedor idempotentes y registro de busquedas`

- [X] T914 · Integración: migraciones up→down→up, `CHECK`s, `IsFilePublished` y atomicidad · `[backend]`

  - **Archivos**: `backend/internal/medios/repository_migrations_integration_test.go` y
    `repository_audit_integration_test.go` (NUEVOS; `//go:build integration`).
  - **Qué hace**: la verificación ejecutable de `data-model.md` §"Validación del modelo":
    (1) las migraciones `000014` y `000015` aplican **up → down → up** sin errores (restauración del
    `CHECK` de `admin_actions`; con `testutil.SerializeMigrations`, decisión 25); (2) el catálogo de
    `permissions` **sigue en 10** filas con `medios` entre ellas (F9 no lo toca); (3) los `CHECK`s
    del modelo se disparan en la BD: `description_en` sin `description_es`, `duplicate_mark_resolved_at`
    sin `duplicate_of_id`, `availability_mark_resolved_at` sin `availability_mark`,
    `availability_mark`/`origin`/`visibility`/`availability_result` fuera de catálogo, textos con
    espacios sobrantes; **no hay `CHECK` de simetría marca-par** (a propósito, C1) y la **pareja
    recíproca** se verifica por prueba (A→B y B→A); (3bis) la **matriz de marcas
    «par abierto/resuelto × enlace vivo/muerto/inconcluso»** (correcciones C1/A3): los seis estados
    son **legales** —los dos ejes conviven—, escribir `availability_mark` sobre un par **no rompe**
    `duplicate_of_id` (ni su recíproco), y el predicado público bloquea solo el par abierto; (4)
    **mutación + `admin_actions` atómicas**: al forzar un fallo en el registro la mutación **no se
    aplica**, los 6 códigos caben en el `CHECK` ampliado (y un código fuera se rechaza), con
    `target_kind='content'` y `target_label` ≤ 254; (5) `IsFilePublished` con la rama de F9: la
    carátula de una ficha **publicable** responde 200 y la de una oculta/absorbida/con par abierto
    **no** (y el archivo no se borra). *Cubre FR-012, FR-014, FR-017, FR-021, FR-022 y SC-010*.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/medios/` en verde con los cinco bloques.
  - **Criterio de terminado**: las migraciones son reversibles, el catálogo de permisos queda fijado
    en 10 y la atomicidad de la auditoría está verificada en la base (edge case «ninguna operación se
    aplica sin registro»).
  - **Commit sugerido**: `test(medios): integracion de migraciones, checks y auditoria atomica`

---

## Fase 6 — Dominio `medios`: servicios

- [X] T915 · `medios`: `service.go` — interfaces, constantes, validación e invariantes · `[backend]`

  - **Archivos**: `backend/internal/medios/service.go` y `service_test.go` (NUEVOS).
  - **Qué hace**: **solo** la interfaz `Repository` que consume el service, más el uso de
    `storage.Store` (R3) —`FeedSource` (T908), `AvailabilityProber` (T909) y `CoverFetcher` (T910)
    **ya están declaradas en su tarea, con sus tipos** y aquí **no se redeclaran** (corrección M1)— y
    las **constantes revisables**
    (P9-18/R9-9): las dos URLs de FR-010, timeouts 10/10/5 s
    (las allowlists de hosts, los topes de cuerpo y el resto de la política de red viven con la
    fábrica de T907A/P9-20),
    topes duros **30** elementos / **30** carátulas / **30** comprobaciones, **máx. 4 en vuelo**,
    presupuesto de la búsqueda **90 s**, `pageDefault = 12`, `maxPage = 50`, `maxPortada = 5`,
    `maxTitulo = 300`, `maxDescripcion = 300`. **Normalización** (trim; `""`/espacios en
    `descriptionEn` → `NULL`, patrón I6) y **validación de dominio** (FR-023): descripción 1…300 por
    idioma con **español por campo** (`descriptionEn` exige `descriptionEs` → `apperr.Invalid` con
    `details.descriptionEs`), `visibility` dentro del catálogo, `lang` ∈ `es|en`, `otherId` ≠ `id`
    y límites de paginación; todo contenido como **datos** (§IV). *Soporta FR-019, FR-020, FR-023,
    FR-025*.
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/medios/` en verde: cada entrada inválida → `apperr.Invalid` con `details`
    por campo (`descriptionEs` vacía o de 301 caracteres, `descriptionEn` sin `descriptionEs`,
    `visibility: 'draft'`, `lang=fr`, `otherId` igual al `id`); `""`/`"  "` en `descriptionEn` →
    `NULL`; tildes/`ñ`/`¿¡`/emojis intactos; las constantes quedan **fijadas por prueba**
    (`pageDefault` 12, `maxPage` 50, `maxPortada` 5, `maxDescripcion` 300, presupuesto 90 s).
  - **Criterio de terminado**: las reglas de validación y los topes tienen **una** implementación y
    sus pruebas; los límites son el único punto de verdad (los usan T916–T921 y los cita el contrato).
  - **Commit sugerido**: `feat(medios): invariantes, validacion y constantes del dominio`

- [X] T916 · `medios`: `service_public.go` — `GetMedios` (sección con «Cargar más») · `[backend]`

  - **Archivos**: `backend/internal/medios/service_public.go` y `service_public_test.go` (NUEVOS).
  - **Qué hace**: `GetMedios(lang, limit, offset)` (FR-002…FR-006): solo fichas **publicables**
    (predicado de R9-7), orden `published_at DESC, created_at ASC, id` (FR-006/default #11),
    tanda por `limit`/`offset` (default **12**, tope **50**; R9-10) y sobre
    `{lang, items, total, limit, offset}`. La **descripción** se resuelve con fallback
    **`en` si existe, si no `es`** en el servidor (FR-025; el cliente no reimplementa el fallback) y
    las claves **se omiten** cuando no aplican (`description` sin descripción, `coverUrl` sin
    carátula, `videoUrl`/`audioUrl` según los enlaces que la importación unió — SC-010: sin huecos,
    nunca `null` —). `title`, `publishedAt`, `coverUrl`, `videoUrl` y `audioUrl` son **idénticos en
    ambos idiomas** (FR-026). El DTO público **nunca** contiene `visibility`, marcas de ningún eje
    (`duplicateOfId`/`duplicateMarkResolvedAt`/`availability*`), `mergedIntoId` ni `sources`. *Cubre FR-002, FR-003, FR-005,
    FR-006, FR-009, FR-025, FR-026 y SC-002/SC-011* (capa de servicio).
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos; **las más importantes del
    módulo**). **Cómo se verifica**: `go test ./internal/medios/` en verde: solo publicables —ni
    `hidden`, ni absorbidas, ni par abierto— pero **sí** `missing_from_feed`
    (FR-017: sigue visible) y el par **resuelto**; orden con fechas iguales por
    `created_at ASC, id`; con 13 publicables y `limit=12` salen 12 y `total=13`; `limit=0`/`>50` →
    `invalid` o tope aplicado según el contrato; fallback por campo (nunca campo vacío); con
    `descriptionEn` existe → en inglés; `title`/fechas/enlaces idénticos entre `lang=es` y `lang=en`;
    `lang=fr` → `invalid`; el DTO serializado **no** contiene las claves internas.
  - **Criterio de terminado**: SC-002 (0 no publicables) y SC-011 (0 huecos, datos idénticos)
    verificables en el service; cobertura medida (`go test -cover`, ≥ 80 % al cierre).
  - **Commit sugerido**: `feat(medios): seccion publica paginada con fallback de idioma`

- [X] T917 · `medios`: `service_public.go` — `GetPortadaMedios` (tope 5) · `[backend]`

  - **Archivos**: `backend/internal/medios/service_public.go` (EDITADO) y `service_public_test.go`
    (AMPLIADO).
  - **Qué hace**: `GetPortadaMedios(lang)` (FR-027/US6/SC-015): mismo predicado publicable y mismo
    fallback que la sección, con **tope `maxPortada = 5` resuelto en el servidor** (`LIMIT 5`; el
    contrato **no** expone parámetro de límite) y el **mismo orden de la sección**: los 5 primeros.
    Devuelve `PortadaMedios` con el **mismo esquema de ítem** `MedioPublico` («el bloque es un
    extracto de los mismos episodios, no un modelo nuevo»). Sin publicables → **`items` omitida**
    (el bloque no se dibuja — sin mensaje —). *Cubre FR-027 y SC-015*.
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos). **Cómo se verifica**:
    `go test ./internal/medios/` en verde: con 8 publicables salen **exactamente los 5 primeros** en
    el orden de la sección (los otros 3 no aparecen); con menos, exactos (sin huecos); con 0 →
    `items` **omitida** (nunca `null`); 0 no publicables; fallback `en → es`; `lang` inválido →
    `invalid`; el tope **5** queda fijado por prueba (constante: no crece con el contenido).
  - **Criterio de terminado**: US6/FR-027/SC-015 resueltos a nivel de dominio.
  - **Commit sugerido**: `feat(medios): extracto de la portada con tope 5 en el servidor`

- [X] T918 · `medios`: `service_admin.go` — `List` (con `lastSync`) y `Update` · `[backend]`

  - **Archivos**: `backend/internal/medios/service_admin.go` y `service_admin_test.go` (NUEVOS).
  - **Qué hace**: `List` (panel paginado con `platform/paginate`, filtros `visibility`/`reviewMark`/
    `includeMerged`, por defecto sin absorbidas, mismo orden del visitante, con `lastSync` =
    `GetLastMediaSyncRun` cuando ya hubo alguna búsqueda, FR-018; `reviewMark` filtra por la marca de
    esa clase **pendiente** — semántica fijada en la fase 5: par abierto o marca de disponibilidad
    sin resolver —) y `Update` (FR-019/FR-020):
    **solo** `descriptionEs`, `descriptionEn` y `visibility`; validación completa (T915) con
    `details` por campo y **sin guardar nada a medias**; la descripción guardada fija
    `descriptionEdited = true` (nunca se regenera ni se pisa: FR-015/SC-007); **«se publica al
    guardar»** (FR-015: el cambio de algo visible es inmediato). **Registro transaccional con la
    regla exacta de códigos** (P9-14): `media.description.update` por los datos (incluye el primer
    ajuste); `media.hide`/`media.show` **solo** por cambios de `visibility`; un guardado que cambia
    datos **y** visibilidad deja **dos filas**; **fail-closed** (si el registro falla no se aplica
    nada). *Cubre US3, FR-015, FR-019, FR-020, FR-022, FR-023*.
  - **Pruebas incluidas** (§III): unitarias con fakes (repo, auditoría). **Cómo se verifica**:
    `go test ./internal/medios/` en verde: cada campo inválido → `400` con `details` y **sin
    guardar nada**; `descriptionEn` sin `descriptionEs` → `details.descriptionEs`; `""`/espacios en
    `descriptionEn` → `NULL`; 301 caracteres → `details.descriptionEn`; tildes/`ñ`/`¿¡`/emojis
    intactos; guardar la descripción deja `descriptionEdited=true` y la siguiente búsqueda (falso de
    sync) no la toca; una edición que cambia datos y visibilidad deja **dos** filas de auditoría y
    una que solo cambia datos deja **una**; si el registro falla **no se guarda**; `List` filtra por
    marca/visibilidad, excluye absorbidas por defecto y trae `lastSync` solo cuando existe.
  - **Criterio de terminado**: US3 (descripción y visibilidad) y la regla de auditoría de FR-022
    cubiertas en el service con sus pruebas.
  - **Commit sugerido**: `feat(medios): gestion del panel con descripcion, visibilidad y auditoria`

- [X] T919 · `medios`: `service_admin.go` — `Merge` (unión por puntero) y `ResolveMark` · `[backend]`

  - **Archivos**: `backend/internal/medios/service_admin.go` (EDITADO) y `service_admin_test.go`
    (AMPLIADO).
  - **Qué hace**: `Merge(id, otherId)` (FR-013/FR-014/FR-019, P9-10): `{id}` es la ficha que se
    **conserva** (base: nombre, carátula y fecha de la entrada del **video** — precisión 5 —;
    adopta el enlace que faltaba) y `otherId` la que queda **absorbida**: `merged_into_id`,
    `visibility='hidden'`, sus `sources` se **repuntan** a la resultante (que queda `origin='both'`,
    con los dos enlaces y ambas procedencias), el **eje de duplicados** de ambas se **despeja**
    (`duplicate_of_id`/`duplicate_mark_resolved_at`; el eje de disponibilidad **no se toca**: es
    independiente — P9-19 —), y si solo una
    tenía la descripción ajustada por el equipo **manda esa** (FR-015/SC-007). La carátula propia de
    la absorbida, si quedó sin uso, se borra del almacén en mejor esfuerzo (`storage.Delete` es
    idempotente). **Cero `DELETE`** sobre las fichas (FR-019). Auditoría `media.merge` sobre la
    resultante (la absorbida se cita en `targetLabel`). `ResolveMark(id, mark, visibility)`
    (FR-017/US3 esc. 2 y 6) resuelve **la marca indicada** —los dos ejes pueden estar activos a la
    vez (P9-19), así que el cuerpo trae `mark`: `possible_duplicate` resuelve el eje de duplicados y
    `unavailable`/`missing_from_feed` el de disponibilidad—: fija la resolución **de ese eje**
    (`duplicate_mark_resolved_at` / `availability_mark_resolved_at`), aplica la visibilidad elegida,
    es **idempotente** sobre una marca ya resuelta, responde `conflict` si esa marca no está activa y
    **no se reabre por la misma causa**
    (el sync de T921 lo respeta; la excepción A3 la fija T921); auditoría `media.mark.resolve` con la
    marca y la visibilidad elegidas. En «no son el mismo sermón» (`mark='possible_duplicate'`) se
    resuelven **ambas** fichas del par.
    Errores: `otherId` igual a `{id}` → `invalid`; `otherId` inexistente o ya absorbida →
    `not_found`; estado inconsistente por escritura simultánea → `conflict`. *Cubre US3, FR-013,
    FR-014, FR-017, FR-019, FR-022*.
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/medios/` en verde: la unión deja **una** ficha con los dos enlaces,
    `origin='both'`, las dos fuentes repuntadas, sin marca de duplicado y **visible** si lo era la
    base (si tenía una marca de disponibilidad abierta, **sigue**: no se despeja);
    la
    absorbida queda `merged_into_id` + `hidden` y **fuera de todas las consultas** (el fake la
    excluye como el predicado real); **0 `DELETE`** (se comprueba que el repository no recibe
    borrados); descripción editada que sobrevive a la unión (ambos órdenes); eje de duplicados
    despejado en ambas; auditoría `media.merge`; `ResolveMark` con las tres marcas (por eje),
    idempotente, `conflict` sin la marca indicada (incluido el caso de dos marcas activas y resolver
    la otra), «no son el mismo sermón» deja **ambas** visibles y con el par resuelto; errores
    `invalid`/`not_found`/`conflict` según el contrato.
  - **Criterio de terminado**: la única operación irreversible del módulo está cubierta campo a campo
    y la no-reapertura de marcas queda fijada por prueba.
  - **Commit sugerido**: `feat(medios): union por puntero y resolucion de marcas`

- [X] T920 · `medios`: `service_sync.go` — importación idempotente, recorte y unión/marcado · `[backend]`

  - **Archivos**: `backend/internal/medios/service_sync.go` y `service_sync_test.go` (NUEVOS).
  - **Qué hace**: la primera mitad de `BuscarNuevos` (FR-011…FR-016, P9-3/P9-5/P9-9/P9-11/P9-17),
    orquestada con **dobles falsos de `FeedSource`** (los dos orígenes): (1) lee los dos feeds **en
    paralelo**; un feed que falla → resumen `failed` para ese proveedor, se procesa el otro y **0
    ocultas/marcadas por esa causa** (FR-016); (2) **alta idempotente** de lo que falta por
    `(provider, external_id)` (`InsertMediaEpisodeSource` con `ON CONFLICT DO NOTHING`; nunca dos
    veces, nunca borra, nunca modifica lo que el equipo decidió — FR-012/FR-015 —); (3)
    **instantánea**: los datos del feed se toman una sola vez (título, fecha, enlaces, carátula y la
    descripción corta generada); búsquedas posteriores **no los refrescan** (R9-6); (4)
    **descripción corta** recortada al importar (HTML → texto plano, primeras frases, **≤ 300
    caracteres**, corte en frontera de frase o palabra con «…»); si el feed **no trae descripción**
    el campo queda **ausente** y la ficha omitirá la línea (precisión 3) — nunca se inventa texto;
    protegida por `descriptionEdited` (jamás se pisa una descripción del equipo); (5) **unión
    automática** cuando título normalizado igual **y** mismo día en `America/Costa_Rica` **sin
    margen** (T907) → una ficha con los dos enlaces, nombre/carátula/fecha del **video** (precisión
    5); candidato sin ambas condiciones → **dudoso** (FR-014): entradas separadas, la adicional
    **oculta**, **ambas** con la marca «posible duplicado» **abierta** (`duplicate_of_id` recíproco;
    solo el eje de duplicados — P9-19 —) (la primera
    conserva su flag `visible`, R9-7); sin candidatura → entradas independientes sin marca. Todo
    dentro de transacciones cortas por elemento (nada a medias, FR-023). *Cubre US2, FR-011, FR-012,
    FR-013, FR-014, FR-016, FR-020 y SC-003/SC-004/SC-007*.
  - **Pruebas incluidas** (§III): unitarias con fakes (repo, `FeedSource` doble). **Cómo se
    verifica**: `go test ./internal/medios/` en verde: **alta idempotente** (segunda búsqueda = 0
    cambios; doble búsqueda simultánea simulada → el `UNIQUE` decide); **instantánea** (un feed con
    título/fecha/carátula cambiados **no** altera lo importado); descripción recortada (≤300, corte
    en frase/palabra con «…», HTML del podcast → texto plano) y **protegida** (`descriptionEdited`:
    el sync no la escribe); **sin descripción en el feed → campo ausente** (precisión 3 fijada por
    prueba); unión automática con el **par real** de títulos (8 unidas del set real) y nombre del
    video; **dudoso** (mismo día con títulos distintos, o mismo título con días distintos) → adicional
    oculta y **ambas** marcadas con `duplicate_of_id` recíproco; sin candidatura → independientes;
    **feed caído** → resumen `failed`, se importa lo del otro y **0 ocultas/marcadas**; un episodio
    **oculto** sigue oculto tras el sync (SC-007) y una unión confirmada **no se deshace**.
  - **Criterio de terminado**: US2 «publicar sin escribir ningún dato» (SC-004) y las reglas de
    FR-012…FR-016 están fijadas por prueba con los dos feeds falsos (sin red).
  - **Commit sugerido**: `feat(medios): importacion idempotente con union automatica y marcado`

- [X] T921 · `medios`: `service_sync.go` — ventana del feed, disponibilidad, carátulas y resumen · `[backend]`

  - **Archivos**: `backend/internal/medios/service_sync.go` (EDITADO) y `service_sync_test.go`
    (AMPLIADO).
  - **Qué hace**: la segunda mitad de `BuscarNuevos` (FR-017/FR-018, P9-3/P9-6/P9-7/P9-19): (1)
    **ventana del feed**: para cada proveedor **que respondió**, compara los identificadores
    importados (`ListMediaEpisodeSourcesByProvider`) con los que trajo el feed; lo que ya no aparece
    **solo** se **comprueba**, nunca se oculta por esa razón (FR-017/SC-014). **Cola de
    comprobación** (corrección de la fase 5): se comprueba lo que salió de la ventana y pertenece a
    fichas **`visibility='visible'` no absorbidas** — una ficha oculta por el equipo no se
    re-comprueba y «Mostrar de nuevo» la devuelve a la cola (D-28: si sigue muerta, la próxima
    búsqueda la vuelve a ocultar) —; (2) **disponibilidad por fuente** (≤ 30, máx. 4 en vuelo,
    timeout 5 s) con la regla **fail-safe** de T909 y **agregación por ficha** (corrección M2: el
    resultado es **el peor** de sus fuentes comprobadas — `unavailable` peor que `inconclusive`,
    peor que `available` —; un enlace confirmado muerto oculta **la ficha entera** aunque el otro
    responda 200: precisión 2). Escritura por ficha: **404/410 confirmado** (el peor es
    `unavailable`) → `availability_result='unavailable'` + `visibility='hidden'` + marca
    **`unavailable` abierta** (+`hiddenCount`) — **excepción aprobada en la fase 5 (A3)**: la
    indisponibilidad confirmada **manda siempre**, aunque hubiera resolución previa «mantener
    visible» o «Mostrar de nuevo»; se escribe marca **nueva abierta** porque **cambiar de clase
    resetea la resolución** —; sin confirmación de muerte y con **todas** las fuentes de la ficha
    fuera de sus ventanas → marca `missing_from_feed`, **sigue visible** (si **alguna** fuente sigue
    en su ventana: **sin marca** — M2 —); inconcluso → cuenta `availabilityInconclusiveCount` y
    **nunca** oculta (precisión 1/2); una marca ya **resuelta** por el equipo **no se reabre por la
    misma causa** (default #6: misma clase y sin confirmación nueva); (3) **carátulas** de lo importado con `CoverFetcher`
    (≤ 30, máx. 4 en vuelo, timeout 10 s; fallo → ficha sin carátula, nunca aborta); (4)
    **presupuesto de 90 s** sobre el `context` de la petición: al agotarse se cierra **con lo hecho**
    y el resumen lo dice (conteos parciales; cada elemento ya quedó guardado en su transacción);
    (5) **resumen** `ResumenBusqueda` (imported/joined/marked/hidden/availability* + estado de cada
    feed; `markedCount` cuenta **solo** «posible duplicado», criterio B2 de la fase 5) +
    `InsertMediaSyncRun` y **una** fila de auditoría `media.sync` con extracto en
    `targetLabel` (el detalle numérico vive en `media_sync_runs`, P9-14). *Cubre FR-017, FR-018,
    FR-022 y SC-007/SC-010/SC-014*.
  - **Pruebas incluidas** (§III): unitarias con fakes (repo, probadores, store). **Cómo se
    verifica**: `go test ./internal/medios/` en verde con la tabla de FR-017: sale del feed +
    `404` → **oculta** + «no disponible» + `hiddenCount`; sale + `200` → **visible** + «ya no
    aparece en el feed» **solo si TODAS sus fuentes** están fuera de ventana (M2; con una sola
    fuera → sin marca); sale + inconcluso → **visible** + marca + `availabilityInconclusiveCount`;
    **nunca se oculta por salir de la ventana**; una ficha de dos enlaces con un enlace confirmado
    muerto → **ficha entera oculta** (precisión 2, el peor gana); **matriz «par abierto/resuelto ×
    enlace vivo/muerto/inconcluso»** (los seis casos dejan el estado que fija P9-19: los dos ejes
    conviven, el par **nunca** se rompe al marcar disponibilidad y «Revisar duplicado» sigue
    disponible — C1 —); caso **A3** fijado: «marca de disponibilidad resuelta + ficha visible +
    `404` confirmado → **oculta** + marca nueva abierta» (y su recíproco: misma clase sin
    confirmación nueva → **no** se reabre); la cola solo mira fichas visibles (oculta por el equipo
    → 0 comprobaciones); `markedCount` = solo «posible duplicado» (B2); feed caído → **0**
    comprobaciones ni ocultaciones por esa rama (y su fuente no cuenta como «fuera de ventana»);
    carátula que falla → ficha sin carátula y el resumen lo cuenta; presupuesto agotado (context
    cancelado) → resumen parcial **sin pérdidas** (lo hecho persiste, nada queda a medias);
    `InsertMediaSyncRun` con los conteos y **una** fila `media.sync` (nunca una por elemento).
  - **Criterio de terminado**: la regla del «estante de 15» (SC-014), la precedencia de la fase 5
    (A3), la agregación (M2) y el resumen de FR-018 están fijados por prueba; la búsqueda cabe en su
    presupuesto y nunca vacía el sitio.
  - **Commit sugerido**: `feat(medios): disponibilidad a prueba de fallos, caratulas y resumen de la busqueda`

- [X] T922 · `medios`: `service_audit.go` — `audit.Recorder` del dominio · `[backend]`

  - **Archivos**: `backend/internal/medios/service_audit.go` y `service_audit_test.go` (NUEVOS).
  - **Qué hace**: implementa `audit.Recorder` para el dominio (patrón de F4/F6/F7,
    `donaciones/service_audit.go`, P9-14): `RecordDenied` resuelve **método+ruta → acción** de las
    rutas del módulo (`/api/v1/admin/medios…` → `media.*`) y la persiste con `result='denied'`
    best-effort; el helper de rechazo hace lo propio con `result='failure'`. Nunca transporta el
    cuerpo de la petición (regla de F2). `target_label`: `Medios · <title>` (máx. 254) y
    `Medios · Búsqueda` (con el extracto del resumen) para `media.sync`. *Cubre FR-022 (y el patrón
    de denegación de FR-021) y SC-009/SC-010*.
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/medios/` en verde: cada ruta del módulo resuelve a su acción y
    `targetLabel`; una ruta desconocida no deja fila; un fallo de registro no cambia la respuesta
    (best-effort); el `target_label` nunca supera 254.
  - **Criterio de terminado**: las denegaciones sobre las rutas de F9 quedan registradas igual que las
    de F2–F7 (SC-010).
  - **Commit sugerido**: `feat(medios): registro de denegaciones y rechazos del modulo`

---

## Fase 7 — Handlers, rutas y cableado

- [X] T923 · `handler.go` + `handler_public.go` — `GET /api/v1/medios` · `[backend]`

  - **Archivos**: `backend/internal/medios/handler.go`, `handler_public.go` y
    `handler_public_test.go` (NUEVOS).
  - **Qué hace**: helpers HTTP del dominio (decodificar/validar/responder con `apperr` +
    `WriteError`, patrón de F2/F4) y la operación del contrato: valida `lang` (`es|en`, por defecto
    `es`) y `limit`/`offset`, llama a `GetMedios` (T916) y responde `200` con `MediosPublicos` y
    **`Cache-Control: no-store`** (SC-003/FR-015). Sin autenticación (FR-002). El cuerpo JSON
    **nunca** contiene `visibility`, marcas, `sources` ni fichas no publicables; las claves
    opcionales viajan **omitidas** (nunca `null`). *Cubre FR-002, FR-005, FR-006, FR-008, FR-009*
    (vía de salida pública).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/medios/` en verde: `200` con el sobre correcto y `no-store`; `lang=fr` →
    `400 invalid` con `details.lang`; `limit=51`/`offset=-1` → `400 invalid`; claves omitidas cuando
    el service no las devuelve; el JSON serializado **no contiene** `visibility`/marcas/
    `sources` (aserción sobre el JSON crudo); error del service → sobre `ErrorEnvelope` genérico
    (nunca texto interno).
  - **Criterio de terminado**: `quickstart.md` §1 es ejecutable tal cual por API (US1).
  - **Commit sugerido**: `feat(medios): endpoint publico de la seccion de medios`

- [X] T924 · `handler_public.go` — `GET /api/v1/portada/medios` · `[backend]`

  - **Archivos**: `backend/internal/medios/handler_public.go` (EDITADO) y `handler_public_test.go`
    (AMPLIADO).
  - **Qué hace**: publica la operación del contrato (P9-2/P9-12): valida `lang`, llama a
    `GetPortadaMedios` (T917) y responde `200` con `PortadaMedios` y **`Cache-Control: no-store`**;
    **sin parámetro de límite** (el tope 5 es del servidor); `items` **omitida** con 0 publicables
    (nunca `null`: el bloque no se dibuja, FR-027). La ruta va en `RegisterPublic` **sin
    middlewares** (pública, solo lectura) — anotando que el segmento `portada` es el agregado de F3
    pero el recurso lo sirve `medios` (Complexity 5 del plan). *Cubre FR-027*.
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/medios/` en verde: `200` con el sobre correcto y `no-store`; `items` omitida
    con 0 publicables; `lang=fr` → `400 invalid`; error del service → sobre genérico.
  - **Criterio de terminado**: `quickstart.md` §10 es ejecutable tal cual por API (US6).
  - **Commit sugerido**: `feat(medios): endpoint publico del bloque de medios de la portada`

- [X] T925 · `handler_admin.go` — listado del panel + `PATCH {id}` · `[backend]`

  - **Archivos**: `backend/internal/medios/handler_admin.go` y `handler_admin_test.go` (NUEVOS).
  - **Qué hace**: `GET /api/v1/admin/medios` (listado paginado con `limit`/`offset` vía
    `platform/paginate`, filtros `visibility`/`reviewMark`/`includeMerged` — `reviewMark` = marca de
    esa clase **pendiente**, semántica de la fase 5 —, `lastSync` cuando
    existe) y `PATCH /api/v1/admin/medios/{id}` (FR-019/FR-020): **solo** `descriptionEs`,
    `descriptionEn` y `visibility` —cualquier otra clave del cuerpo → `400 invalid` (el DTO es
    `additionalProperties: false`)—; 1…300 caracteres por idioma, texto plano guardado íntegro
    (tildes, `ñ`, `¿¡`, emojis; §IV); `descriptionEn` sin `descriptionEs` → `400` con
    `details.descriptionEs`; respuesta `200` con `MedioAdmin` (los dos idiomas, `descriptionEdited`,
    `sources` y las marcas **por eje** — `duplicateOfId`/`duplicateMarkResolvedAt` y
    `availabilityMark`/`availabilityMarkResolvedAt`: una fila puede mostrar dos píldoras —). Códigos 200/400/401/403/404/500 y sobres uniformes. *Cubre FR-019, FR-020, FR-023*
    (vías de panel).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/medios/` en verde: cada operación con sus códigos y sobres; `details` por
    campo en los `400` (incluida una **clave desconocida** del `PATCH`, p. ej. `title`); el
    `PATCH` con un solo campo no toca los demás (lo garantiza el service; aquí se comprueba que el
    body se transfiere completo); `403` de permiso se prueba en T928 (cada grupo con su
    `AdminChain`).
  - **Criterio de terminado**: `quickstart.md` §6/§7 ejecutables vía API (US3).
  - **Commit sugerido**: `feat(medios): endpoints del panel (listado y edicion de descripcion)`

- [X] T926 · `handler_admin.go` — `POST buscar-nuevos` · `[backend]`

  - **Archivos**: `backend/internal/medios/handler_admin.go` (EDITADO) y `handler_admin_test.go`
    (AMPLIADO).
  - **Qué hace**: publica la única automatización de F9 (FR-011): invoca `BuscarNuevos` (T921) con
    el **context de la petición** (presupuesto 90 s) y responde `200` con `ResumenBusqueda`
    **también cuando un feed falló o el presupuesto se agotó** (el resumen lo dice: **no es error**,
    FR-016/R9-9); `500` solo ante fallo interno (lo ya importado queda intacto). El resumen **nunca
    expone texto técnico de red** (nombres de host internos, errores de Go…: FR-018). Códigos
    200/401/403/500. *Cubre FR-011, FR-016, FR-018*.
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/medios/` en verde: `200` con el sobre `ResumenBusqueda` completo
    (`importedCount`…`youtube`/`podcast`); un resumen con `youtube: "failed"` sigue siendo `200`; el
    JSON del resumen **no contiene** textos de error ni URLs internas (aserción sobre el JSON crudo);
    `401`/`403` según el contrato; error del service → `500` genérico.
  - **Criterio de terminado**: `quickstart.md` §3 es ejecutable vía API (US2) y el resumen cumple
    FR-018 sin filtrar información interna.
  - **Commit sugerido**: `feat(medios): endpoint de buscar episodios nuevos con resumen`

- [X] T927 · `handler_admin.go` — `POST {id}/unir` y `POST {id}/marcas/resolver` · `[backend]`

  - **Archivos**: `backend/internal/medios/handler_admin.go` (EDITADO) y `handler_admin_test.go`
    (AMPLIADO).
  - **Qué hace**: `POST /api/v1/admin/medios/{id}/unir` (cuerpo `UnirInput` con `otherId`: UUID,
    distinto de `{id}`) y `POST /api/v1/admin/medios/{id}/marcas/resolver` (cuerpo
    `ResolverMarcaInput` con **`mark`** —qué marca se resuelve: los dos ejes pueden estar activos a
    la vez, P9-19/C1— y `visibility` del catálogo) sobre el service de T919; respuestas `200`
    con `MedioAdmin`; códigos 200/400/401/403/404/409/500 según el contrato (`otherId` mal formado o
    igual a `{id}` → `400`; `mark`/`visibility` fuera de catálogo → `400` con `details`;
    inexistente/absorbida → `404`; inconsistente → `409`; sin marca **de esa clase** que resolver →
    `409`). *Cubre FR-014, FR-017, FR-019* (vías de panel).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/medios/` en verde: cada operación con sus códigos y sobres; `details` por
    campo en los `400` (`otherId`, `mark`, `visibility`); resolución **por eje** (ficha con las dos
    marcas activas: resolver una no resuelve la otra) y `409` cuando `mark` no corresponde a la
    marca activa; `404`/`409` trasladados del service sin filtrar
    textos internos; el `403` de permiso se prueba en T928.
  - **Criterio de terminado**: `quickstart.md` §5 es ejecutable vía API (US3, SC-003).
  - **Commit sugerido**: `feat(medios): endpoints de union de duplicados y resolucion de marcas`

- [X] T928 · `routes.go` + `cmd/api/main.go` (`AdminChain("medios")`) + humo · `[backend]`

  - **Archivos**: `backend/internal/medios/routes.go` y `routes_test.go` (NUEVOS),
    `backend/cmd/api/main.go` y `main_test.go` (EDITADOS: DI del dominio `medios` — repository,
    `FeedSource` real, probadores, `CoverFetcher`, `storage.Store` — + registro de
    `/api/v1/medios`, `/api/v1/portada/medios` y `/api/v1/admin/medios`).
  - **Qué hace**: cablea el grupo de panel con la cadena aprobada de F2 —`authn → guard de cambio de
    contraseña → authz("medios") → CSRF`— **sin helpers nuevos de middleware**: `AdminChain("medios")`
    con el permiso **ya sembrado** (P9-15/FR-021: verificado **en el servidor**; independiente de los
    demás permisos — Decisión 5 —). Las 2 rutas públicas quedan **sin middlewares** (solo lectura,
    sin rate-limit: no hay superficie pública escribible) y `/healthz` intacto (§8.1.9). *Cubre
    FR-021 (el servidor es la única autoridad) y FR-022* (el `Recorder` de T922 recibe cada cadena).
  - **Pruebas incluidas** (§III): humo de composición + de permisos. **Cómo se verifica**:
    `go test ./cmd/api/` y `go test ./internal/medios/` en verde: rutas nuevas con su sobre y su
    `Cache-Control`; `/healthz` sin sesión y `404` `not_found` en rutas inexistentes (§8.1.9); una
    cuenta con `medios` pasa y una cuenta con `eventos`/`ministerios`/`donaciones` recibe **`403`**
    en `/api/v1/admin/medios` (y al revés: una cuenta solo `medios` no accede a otros módulos, US2
    esc. 9); el `403` genérico queda registrado como `denied`; `go test ./...` completo en verde.
  - **Criterio de terminado**: el permiso queda verificado en servidor y auditable de un vistazo en
    `routes.go` (SC-009).
  - **Commit sugerido**: `feat(medios): rutas con AdminChain("medios") y cableado del dominio`

- [ ] T946 · `medios`: permitir la descripción vacía (contrato + validación) · `[backend]` · **decisión posterior (2026-10-18)**

  - **Archivos**: `specs/009-medios-predicas-y-podcasts/contracts/openapi.yaml`, `backend/api/openapi.yaml`,
    `backend/internal/medios/service_admin.go`, `service.go`, `handler_admin.go` y sus pruebas.
  - **Qué hace**: **decisión del humano del 2026-10-18** — *la descripción corta puede estar vacía*.
    (a) El contrato deja de exigir un mínimo de un carácter en la descripción en español (se retira el
    `minLength: 1`), tanto en el delta como en el **contrato vivo** (sigue en **0.9.0**: F9 no se ha
    publicado todavía); (b) el `PATCH` del panel **puede vaciarla** (`""` pasa a significar «sin
    descripción», con el mismo tratamiento que la versión en inglés, patrón I6); (c) el servicio deja
    de responder `400` en ese caso y **la auditoría sigue registrando el cambio**; (d) el DTO público
    **omite** la descripción cuando está vacía. *Cubre FR-020 (ampliado) y la decisión del 2026-10-18.*
    **Nota**: en el frontend, T937 ya contempla **no pintar** la línea de la descripción cuando está vacía.
  - **Pruebas incluidas** (§III): unitarias de handler y de servicio (vaciar la descripción → `200`, el
    DTO público sale sin descripción, la auditoría queda registrada) e integración del `PATCH` contra
    PostgreSQL real. **Cómo se verifica**: `go test ./internal/medios/...` y
    `go test -tags=integration ./internal/medios/...` en verde; el contrato sigue válido y
    `make api-gen` sin deriva.
  - **Criterio de terminado**: el equipo puede dejar un episodio sin descripción desde el panel y la
    ficha pública no pinta esa línea.
  - **Commit sugerido**: `feat(medios): permitir dejar la descripcion vacia desde el panel`

---

## Fase 8 — Frontend

> Se **aplica** el sistema de F10 y los patrones de F2/F3/F4/F6/F7 (no se redefinen). Nombres de
> pieza y composición según `ux.md` §6 (y §4 para pantallas/estados); i18n solo en el sitio público
> (FR-024: el panel va en español). Ningún `fetch` fuera del cliente de API (regla F2); cero
> `dangerouslySetInnerHTML` (el contenido del feed es dato, §IV); **`HiddenNotice` no se usa** (F9
> redirige, P9-16). Los enlaces externos usan `PublicButton` con `external` (el `target`/`rel` ya los
> resuelve F10) y **anuncian** servicio, salida del sitio y pestaña nueva (FR-005).

- [ ] T929 · `api/medios.ts` + hooks (público, portada, panel) · `[frontend]` `[P]` (P3)

  - **Archivos**: `frontend/src/api/medios.ts` (NUEVO),
    `frontend/src/features/publico/medios/hooks/usePublicMedios.ts`, `hooks/usePortadaMedios.ts`
    (NUEVOS), `frontend/src/features/medios/hooks/useMedios.ts`, `useBuscarEpisodios.ts`,
    `useActualizarDescripcion.ts`, `useOcultarMostrarEpisodio.ts`, `useUnirDuplicados.ts`,
    `useResolverMarca.ts`, `hooks/queryKeys.ts` (NUEVOS).
  - **Qué hace**: cliente tipado desde `schema.d.ts` (T905): `getMedios(lang, limit, offset)` y
    `getPortadaMedios(lang)` (públicos) y panel (`listMedios`, `syncMedios`, `updateMedio`,
    `mergeMedio`, `resolveMedioMark`). Hooks TanStack Query: `usePublicMedios` con key
    `['medios', lang, limit, offset]` y **acumulación de tandas** para «Cargar más» (el hook expone
    `hasMore` a partir de `total`); `usePortadaMedios` con `['portada-medios', lang]`
    (`placeholderData: keepPreviousData` y retención de la última respuesta válida: sin parpadeos al
    cambiar de idioma); los del panel (`useMedios` con filtros, `useBuscarEpisodios` como mutation
    que **deshabilita la UI**, devuelve el resumen y **invalida** catálogo + documento público +
    portada; las demás mutations invalidan igual para que el cambio sea visible en la primera carga,
    SC-003/FR-015). El cliente **no filtra ni reimplementa** el fallback de idioma ni la regla de
    publicables: el servidor ya resolvió todo. *Soporta US1–US4 y US6*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: cada hook contra un handler MSW (éxito, `400` con `details`,
    `401`, `403`, `404`, `409`, `500`); «Cargar más» acumula la segunda tanda y `hasMore` se agota
    con `total`; respuesta del bloque sin `items`; error con datos previos retenidos; la mutation de
    búsqueda desactiva el estado «en vuelo» y refresca las tres consultas; invalidación tras
    descripción/ocultar/unir/resolver.
  - **Criterio de terminado**: ningún componente hace `fetch` directo (regla F2) y los tipos salen
    solo del contrato vivo (0 `any`).
  - **Commit sugerido**: `feat(frontend): cliente y hooks de medios`

- [ ] T930 · i18n público (§7.2) + `messages.ts` del panel (§7.1) · `[frontend]` `[P]` (P3)

  - **Archivos**: `frontend/src/features/publico/i18n/{messages,es,en}.ts` (EDITADOS: claves
    tipadas de `ux.md` §7.2 — `nav.sections.media`, `media.page.title`, `media.page.subtitle`,
    `media.loadMore`, `media.loading`, `media.error.load`, `media.error.retry`,
    `media.error.partial`, `media.link.youtube`, `media.link.spotify`, `media.link.newTab`,
    `a11y.media.link.youtube`, `a11y.media.link.spotify`, `a11y.media.image`,
    `a11y.media.card.heading`, `a11y.media.card.newTab`, `home.media.title`, `home.media.subtitle`,
    `home.media.seeAll`, `a11y.home.media.heading` —) y `frontend/src/features/medios/messages.ts`
    (NUEVO: catálogo del panel de `ux.md` §7.1). Sin la clave, TypeScript **no compila** (garantía
    de SC-006/SC-011).
  - **Qué hace**: completa el diccionario bilingüe del sitio público (FR-025) y los textos en
    español del panel (FR-024: el panel **no** pasa por `t()`): banda de ayuda, zona de búsqueda,
    los seis casos del resumen (§4.4 de `ux.md`: sin segmentos en 0 — D-19 — y nombra los servicios
    como los llama la iglesia — D-21 —), estados vacíos, avisos de ocultar/mostrar, límite de la
    descripción, «inglés sin español», textos de la unión (vista previa, «Esto no se puede
    deshacer.»), explicadores de las tres marcas y la no-reapertura. **La redirección no añade
    cadena** (no hay aviso que mostrar). *Cubre FR-024, FR-025 y SC-006/SC-011*.
  - **Pruebas incluidas** (§III): Vitest (tabla de casos). **Cómo se verifica**:
    `npm test -- --run` en verde: las claves `es`/`en` son simétricas (0 huecos); los textos del
    panel son cadenas en español (0 claves de i18n en `messages.ts` del panel); el anuncio de
    `a11y.media.link.*` contiene servicio, «pestaña nueva» y «sales del sitio» (y su equivalente en
    inglés).
  - **Criterio de terminado**: i18n completo del sitio público y catálogo del panel listo para
    T933–T938.
  - **Commit sugerido**: `feat(frontend): textos de medios (i18n publico y panel)`

- [ ] T931 · Permisos, accesos del panel y `ACTION_LABELS` · `[frontend]`

  - **Archivos**: `frontend/src/lib/permissions.ts` (EDITADO: `MEDIOS = 'medios'` y su entrada en
    `MODULE_PERMISSIONS`), `frontend/src/features/roles/permissions.ts` (EDITADO: el sello
    «Disponible más adelante» deja de aplicarse a `medios` — el catálogo se lee dinámico de la API y
    **no** cambia —), `frontend/src/features/panel/components/AccesosRapidos.tsx` y
    `frontend/src/features/panel/pages/InicioPage.tsx` (EDITADOS: tarjeta **«Gestión de medios»**
    **solo** con el permiso, `ux.md` §4.2), `frontend/src/app/layout.tsx` (EDITADO: entrada del menú
    del panel filtrada por permiso) y `frontend/src/features/auditoria/messages.ts` (EDITADO:
    `ACTION_LABELS` cubre los **6 códigos** `media.*` con etiqueta humana, `ux.md` §7.1: «Buscó
    episodios nuevos», «Modificó la descripción de …», «Ocultó/Mostró …», «Unió …», «Resolvió la
    marca de …»). *El registro de rutas del SPA va con las páginas (T933 pública, T936 panel) para
    que cada commit compile (alineación §7).*
  - **Qué hace**: activa en el frontend el permiso existente `medios` (FR-021: **sin permisos
    nuevos** y **sin renombrarlo**, Q13): el formulario de roles de F2 lo ofrece sin sello, el menú
    y los accesos rápidos se filtran por permiso (una cuenta solo `medios` no ve «Eventos» ni
    «Ministerios») y la auditoría muestra etiquetas legibles para los 6 códigos (FR-022). *Cubre
    FR-021 (visibilidad del acceso en el panel) y FR-022 (etiquetas).*
  - **Pruebas incluidas** (§III): Vitest + Testing Library (las del propio cambio; las que fijan
    listas de permisos van en T932). **Cómo se verifica**: `npm test -- --run` en verde (con T932
    integrada): el menú renderiza la entrada solo con el permiso; la tarjeta de `AccesosRapidos` se
    muestra solo con `medios` (y ninguna con la cuenta sin permisos de F9); los **6 códigos**
    `media.*` tienen etiqueta humana en `ACTION_LABELS`.
  - **Criterio de terminado**: el acceso «Gestión de medios» es visible **solo** con el permiso
    `medios` y la auditoría no muestra códigos crudos de F9.
  - **Commit sugerido**: `feat(frontend): permiso medios, accesos del panel y etiquetas de auditoria`

- [ ] T932 · Actualización **acotada** de las pruebas que fijan permisos/accesos · `[frontend]` *(cambio acotado documentado, RG9-6)*

  - **Archivos** (SOLO los que fijen la lista de permisos/accesos; al menos los 4 de RG9-6):
    `frontend/src/features/roles/permissions.test.ts` (hoy fija «el resto del catálogo (F5, F8,
    F9…)» como «Disponible más adelante»: `medios` pasa a `true` y **`noticias` de F8 sigue en
    `false`**), `frontend/src/features/roles/roles.test.tsx` (el sello deja de aplicarse a `medios`),
    `frontend/src/lib/permissions.test.ts` (lista `MODULE_PERMISSIONS`) y
    `frontend/src/features/panel/panel.test.tsx` (tarjeta nueva). *Nada más se toca.*
  - **⚠️ CAMBIO ACOTADO Y DOCUMENTADO (RG9-6)**: F9 edita pruebas de F2/F4/F6/F7 (funcionalidades
    cerradas) porque el catálogo de accesos del frontend activa `medios`. Es un cambio de producto
    **autorizado por el plan**, no un debilitamiento: solo cambian **recuentos, listas y sellos**;
    ninguna aserción de comportamiento se relaja. El backend **no** se toca (el catálogo de permisos
    de F2 sigue en 10). La tarea queda **registrada como excepción** en el PR y la validan
    `revisor-codigo` y `seguridad`.
  - **Qué hace**: ajusta las aserciones que fijan la lista de permisos y accesos para incluir
    `medios`. *Cubre FR-021.*
  - **Pruebas incluidas** (§III): son las propias pruebas actualizadas. **Cómo se verifica**:
    `npm test -- --run` en verde; el diff toca **solo** recuentos/listas/comentarios de esos
    archivos (se revisa el diff: ninguna aserción de comportamiento cambia); el PR lo registra como
    **cambio acotado** (no debilitamiento).
  - **Criterio de terminado**: `make ci` pasa con `medios` activo en el catálogo de accesos y la
    excepción queda documentada en la evidencia.
  - **Commit sugerido**: `test(frontend): permisos y accesos con medios (cambio acotado RG9-6)`

- [ ] T933 · `/medios`: sección pública con fichas, enlaces anunciados y **redirección** · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/medios/MediosPage.tsx`, `MediosGrid.tsx`,
    `MedioCard.tsx`, `CaratulaNeutra.tsx`, `MedioLinks.tsx`, `CargarMasButton.tsx` y `*.test.tsx`
    (NUEVOS), `frontend/src/app/router.tsx` (EDITADO: ruta pública **`/medios`** dentro de
    `PublicLayout`).
  - **Qué hace**: la página con H1 «Medios» + subtítulo fijo que **anuncia la salida del sitio**
    (D-2, cadena de interfaz §7.2) y lista única (`ux.md` §4.1): `MediosGrid` con `<ul>` semántica,
    1 columna móvil / 2 tableta-escritorio (D-13), `RevealGroup` con stagger; `MedioCard`
    (`Card solid as="li"`) con **carátula** `CardImage` 16:9 (o `CaratulaNeutra` — identidad de la
    iglesia, default #4 —), **nombre** h3 **no enlace** (D-5), **fecha** con `formatLongDate`
    (dato único entre idiomas), **descripción corta** con `line-clamp` y **texto completo en el
    DOM** (D-6) **omitida si no existe** (precisión 3) y `MedioLinks` con «Ver en YouTube» y/o
    «Escuchar en Spotify» (`PublicButton external`, icono ↗ + microtexto «pestaña nueva» visible —
    D-7 — y `aria-label` completo con servicio, salida del sitio y pestaña nueva, §7.2); un solo
    formato → **solo su botón** (sin líneas vacías). `CargarMasButton` agrega la siguiente tanda
    (12) y **desaparece** al agotarse (D-3). **Redirección silenciosa (P9-16/RG9-12)**:
    `SectionSkeleton` mientras `usePublicMedios` está en vuelo y **solo con éxito y 0 publicables**
    → `<Navigate to="/" replace />` —**sin aviso**, sin pintar jamás la grilla vacía ni
    `HiddenNotice`—; **ante error NO se redirige** (H1 + «No pudimos cargar los medios…» + «Volver a
    intentar»; con caché, contenido + aviso discreto — `ux.md` §4.1/D-12). Estados de `ux.md` §4.1 y
    `prefers-reduced-motion` respetado; 0 recursos de terceros (SC-005). *Cubre US1, US4, US5 y
    FR-002…FR-008, FR-025, FR-026, FR-028, FR-029, FR-030 (UI).*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: ficha con 1 y 2 enlaces (0 líneas vacías); sin carátula →
    marcador neutro; sin descripción → **sin esa línea**; orden del más reciente al más antiguo con
    desempate por importación (el cliente **no** reordena); «Cargar más» acumula tandas y se agota
    con `total`; fallback `en → es` visible y **`title`/fecha/enlaces idénticos** entre idiomas
    (SC-011); etiquetas de enlace anunciadas (`target="_blank"`, `rel="noopener noreferrer"`,
    `aria-label` con servicio + «pestaña nueva» + «sales del sitio»); **0 iframe y 0 `src` de
    terceros** (SC-005); **redirección sin destello** (con caché redirige en el primer render sin
    pintar H1; sin caché solo el esqueleto; `<Navigate replace>`; **ante error NO redirige**);
    `prefers-reduced-motion` sin revelados.
  - **Criterio de terminado**: `quickstart.md` §1/§4/§9/§11 ejecutables en navegador (US1/US4) y la
    redirección silenciosa verificada (SC-016).
  - **Commit sugerido**: `feat(frontend): pagina /medios con fichas, enlaces anunciados y redireccion`

- [ ] T934 · `useSiteLinks`: entrada «Medios» condicional · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/hooks/useSiteLinks.ts` y su test (EDITADOS).
  - **Qué hace**: `SiteLinkId` crece con **`'media'`** (alineación §8) y la entrada única
    **«Medios»** con href **`/medios`** existe **solo si** `usePublicMedios` trae `total > 0`
    (P9-16: la presencia bebe del **documento de la sección**; la consulta se comparte por
    `queryClient` con la página, 0 peticiones extra). **Depende de T929 y T930** (usa la clave
    i18n tipada `nav.sections.media`). La consumen a la vez `PublicNav`, `MobileMenu` y
    `FooterSiteMap` (grupo **Explorar**, después de «Ofrendas y diezmos»: D-1; fuente única: barra,
    menú hamburguesa y pie **a la vez**, sin huérfanos ni enlaces rotos, FR-007/FR-008). Sin nada
    publicable, el ítem **no existe** en ninguna de las tres superficies. `usePortadaMedios` **no**
    es fuente de navegación (separación P7-8). *Cubre FR-007, FR-008 y SC-016*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: con `total > 0` el ítem aparece en las **tres** superficies con
    `href="/medios"` y su clave traducida; con el documento en `total: 0`, **en ninguna**; el resto
    de enlaces (anclas de portada, `/eventos`, `/ministerios`, `/ofrendas`) no cambia.
  - **Criterio de terminado**: ocultación coherente del ítem (FR-008) verificada en las tres
    superficies (SC-016) con la fuente de P9-16.
  - **Commit sugerido**: `feat(frontend): navegacion con /medios condicional en barra, menu y pie`

- [ ] T935 · `MediosBlock` + `MedioItem` en la portada · `[frontend]` `[P]` (P4)

  - **Archivos**: `frontend/src/features/publico/medios/MediosBlock.tsx`, `MedioItem.tsx` y
    `*.test.tsx` (NUEVOS), `frontend/src/features/publico/pages/HomePage.tsx` (EDITADO: **una
    inserción** de `<MediosBlock />` **inmediatamente después de `<OfrendasBlock />`**) y
    `homePage.test.tsx` (AMPLIADO).
  - **Qué hace**: bloque «Medios» con el patrón de los bloques de F4/F6/F7 (FR-027): `SectionHeading`
    + fila de **hasta 5 `MedioItem`** en el orden de la sección —ficha compacta con carátula 16:9,
    nombre (h3, máx. 2 líneas visibles, completo en el DOM), descripción corta (2 líneas visibles,
    completa en el DOM; **omitida si no existe**) y **enlaces externos por ficha** (pastillas con el
    mismo anuncio completo que en la sección); **fecha omitida** (D-32: presentación, el DTO sí la
    trae)— y **botón «Ver todo»** a `/medios` (solo con ítems; ≥ 44 px). El bloque se alimenta de
    `usePortadaMedios` (`GET /api/v1/portada/medios`, tope 5 del servidor; el cliente no recorta).
    **Sin publicables el bloque devuelve `null`: NO se renderiza** — sin título, sin mensaje, sin
    hueco (patrón `MinisteriosBlock`/`OfrendasBlock`, **diferencia consciente** con
    `ProximamenteBlock` de F4). Esqueleto propio con áreas de carátula mientras carga; **error sin
    caché → no se dibuja**; error con caché → contenido + aviso discreto. Sistema de F10 (`Card`,
    `SectionHeading`, `Reveal` con la dirección que conserve la alternancia — D-14 —), responsive
    1 col móvil / 2 tableta / 5 escritorio (**sin carrusel ni scroll lateral**), foco visible,
    `section` con `aria-labelledby` al h2, ítems como `li` con `h3`. **La portada de F3 no se toca**
    más allá de esta inserción (FR-027/Out of Scope). *Cubre US6, FR-027, FR-028, FR-029, FR-030.*
  - **Pruebas incluidas** (§III): Vitest + Testing Library. **Cómo se verifica**:
    `npm test -- --run` en verde: **5 exactos** con 8 publicables (los 5 primeros del orden de la
    sección); exactos con menos, **sin huecos ni marcadores**; **con 0 publicables el bloque no se
    renderiza** (sin título ni mensaje) y `homePage.test.tsx` confirma que **el bloque de ofrendas y
    el resto de la portada se renderizan igual que antes** y que el bloque queda **inmediatamente
    después de `OfrendasBlock`**; el botón «Ver todo» ausente sin ítems; cada ficha muestra sus
    enlaces anunciados (0 recursos de terceros); en inglés traducido con fallback; descripción nunca
    vacía.
  - **Criterio de terminado**: US6/FR-027/SC-015 en pantalla, en la posición acordada (tras ofrendas),
    y `quickstart.md` §10 ejecutable en navegador.
  - **Commit sugerido**: `feat(frontend): bloque de medios al final de la portada`

- [ ] T936 · Panel: `MediosPage` + búsqueda + resumen + filtros + listado · `[frontend]`

  - **Archivos**: `frontend/src/features/medios/pages/MediosPage.tsx`,
    `components/BuscarEpisodios.tsx`, `components/ResumenImportacion.tsx`,
    `components/FiltroMarcas.tsx`, `components/EpisodiosList.tsx`, `components/MarcaDeRevision.tsx` y
    tests (NUEVOS), `frontend/src/app/router.tsx` (EDITADO: **`/panel/medios`** con
    `RequirePermission(MEDIOS)`).
  - **Qué hace**: gestión en el panel **en español** (FR-024) con la composición de `ux.md` §4.3/§5:
    título «Medios» + banda de ayuda (§7.1); **zona «Buscar episodios nuevos»** (`BuscarEpisodios`:
    título «Desde YouTube y el podcast», botón primario, línea de apoyo que **dice la idempotencia**
    — D-9 — y `role="status"` «Buscando episodios nuevos en YouTube y en el podcast…» mientras corre;
    botón **`Button loading` deshabilitado** y **acciones de fila deshabilitadas** durante la
    búsqueda — D-16 —); **`ResumenImportacion`** (`Notice` **permanente** hasta la próxima búsqueda,
    D-22, con los seis casos de §4.4 y **sin segmentos en 0**, D-19); **`FiltroMarcas`** (D-10 ⟲:
    «Todo» / «Solo lo pendiente de revisión» con conteos; filtro de vista sobre
    `visibility`/`reviewMark`); **`EpisodiosList`** (`Table` en desktop / tarjetas en móvil, patrón
    F2/F4/F6/F7): carátula miniatura 48 px (marcador neutro si no hay), nombre (no editable) +
    fecha, chips de formato «Video»/«Audio» (D-15), **`OrigenBadge`** —«De YouTube» / «Del podcast»
    / «De ambos»— con la **fecha de importación** (`createdAt`) como línea secundaria y
    **`FuentesList`** como detalle de la fila (proveedor, título del feed y enlace de cada `source`;
    **recuperado en la fase 5 — A2 —**: FR-019 pide «el origen de cada entrada» en la lista; D-32
    queda matizado), `StatusPill` «Visible»/«Oculto»,
    `MarcaDeRevision` (píldora con **palabra** + explicador de una línea por tipo, §4.7 —sin iconos,
    D-30 sin implementar—; **una píldora por eje**, P9-19: una fila puede mostrar dos) y acciones por
    fila: **«Editar descripción» · «Ocultar»/«Mostrar» ·
    «Revisar duplicado»** (solo con el par **abierto**) · resolución de marca según su tipo (los
    diálogos los aportan T937/T938; aquí se dejan los ganchos). Estados de `ux.md` §5 (cargando,
    vacío remitiendo **a la zona de importación** — D-33, sin duplicar el botón —, error con
    «Reintentar») y sin permiso la ruta cae en `/sin-permiso` de F2. El listado va en el orden del
    visitante (FR-006: sin orden manual, D-11). *Cubre US2, US3 (listado) y FR-018, FR-019, FR-021,
    FR-024.*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: la zona de búsqueda está disponible antes que el listado; el botón
    se deshabilita y muestra «Buscando…» con `role="status"`; el resumen de los seis casos (con y
    sin segmentos en 0); filtros con conteos y su estado vacío; la fila muestra carátula, nombre,
    fecha, chips, **origen y fecha de importación** (`OrigenBadge`/`FuentesList`, corrección A2 de la
    fase 5), estado y marca(s) con explicador (dos píldoras cuando conviven las dos marcas); «Ocultar» es directo y su `Notice` dice el efecto
    real (§7.1); sin permiso `medios` la ruta no es accesible (texto de `/sin-permiso` intacto).
  - **Criterio de terminado**: `quickstart.md` §2/§3/§6 ejecutables en el panel (US2/US3).
  - **Commit sugerido**: `feat(frontend): panel de medios con busqueda, resumen y listado`

- [ ] T937 · Panel: `DescripcionForm` + `validation.ts` + `errors.ts` · `[frontend]`

  - **Archivos**: `frontend/src/features/medios/components/DescripcionForm.tsx`, `validation.ts`,
    `errors.ts` y tests (NUEVOS).
  - **Qué hace**: edición de la descripción corta en `Dialog` (patrón F3/F6/F7, `ux.md` §4.5):
  **miniatura de contexto** arriba (carátula + nombre + fecha en solo lectura — D-23 —),
    `Tabs` **Español** (obligatorio **si el campo se rellena**, FR-023) / **English (opcional)** con
    la ayuda de F3; `TextArea` grande con contador vivo `{n}/{L}` (**L = 300**, default #1), valor
    vigente **precargado**, texto plano íntegro (tildes/`ñ`/`¿¡`/emojis), ayuda «Se muestra tal como
    la escribas… La descripción completa del feed no se publica.»; **la visibilidad NO se cambia
    aquí** (D-24: solo en la fila); límite excedido → contador en error y «Guardar» deshabilitado
    («{n} de {L} caracteres. Recorta {m} para poder guardar.»); `descriptionEn` sin `descriptionEs`
    → «El contenido en español es obligatorio…» (FR-025); `Button` loading y `Esc` sin descartar lo
    escrito sin aviso (criterio D-18 de F4); al guardar: cierre + `Notice` «Cambios guardados: ya
    visibles en la sección.» (o «Cambios guardados.» si la ficha está oculta). `validation.ts` espejo
    de las reglas del server (Zod) y `errors.ts` traduce `details` por campo; errores junto al campo
    con resumen arriba y foco al primero (FR-023). **No valida datos del feed** (nunca pasan por un
    formulario). *Cubre US3, FR-020, FR-023, FR-025 (UI).*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: el valor vigente se precarga (incluida la generada por el feed);
    `descriptionEn` sin `descriptionEs` bloquea el guardado indicando qué falta; 301 caracteres lo
    indica con el límite y lo que falta recortar; el guardado **no** cambia la visibilidad (D-24);
    `400` del servidor con `details` se pinta junto al campo; emojis y tildes intactos de ida y
    vuelta; `Esc` no pierde lo escrito sin aviso.
  - **Criterio de terminado**: `quickstart.md` §7 ejecutable en el panel y FR-020/FR-023 cubiertos
    en la interfaz.
  - **Commit sugerido**: `feat(frontend): editor de la descripcion corta con es/en`

- [ ] T938 · Panel: `UnirDuplicadosDialog` + resolución de marcas · `[frontend]`

  - **Archivos**: `frontend/src/features/medios/components/UnirDuplicadosDialog.tsx`,
    `ResolverMarcaDialog.tsx` y tests (NUEVOS); `MediosPage.tsx`/`EpisodiosList.tsx` (EDITADOS: las
    acciones de fila quedan completas).
  - **Qué hace**: el flujo de unión (`ux.md` §4.6) y las resoluciones (§4.7). **Unión con decisión
    antes que sorpresa**: «Revisar duplicado» abre un `Dialog` con **las dos candidatas lado a lado**
    (carátula, nombre, fecha, enlace y origen «De YouTube»/«Del podcast», con su `StatusPill`
    actual) y la zona **«Así quedaría la ficha si las unes»** (vista previa del resultado: nombre del
    **video** con la frase declaratoria «Al unir, el nombre publicado es el de YouTube.» — D-25,
    precisión 5 —, carátula, fecha y los dos enlaces). Tres acciones: **«Unir»** → `ConfirmDialog`
    **danger** «Se unirán las dos entradas… **Esto no se puede deshacer.**» («Unir definitivamente» /
    «Cancelar», D-26) → `Notice` «Unidas en una sola ficha, visible en “Medios”.»;
    **«No son el mismo sermón»** (D-27, sin confirmación extra) → resuelve la marca
    `possible_duplicate` (el cuerpo lleva **`mark`**) con
    `visibility: visible` en **ambas** fichas → «Listo: quedaron las dos como episodios distintos.»;
    **«Cancelar»** cierra sin cambios con devolución del foco. **Sin un par abierto no existe
    el botón ni el diálogo** (D-28: unir solo opera sobre pares que el sistema marcó). Resoluciones
    (`ResolverMarcaDialog` + acciones de fila, **cada una con su `mark`** — los dos ejes pueden estar
    activos a la vez y se resuelven por separado, P9-19/C1 —): «No disponible» → **«Mostrar de
    nuevo»** con
    `ConfirmDialog` **suave** («Si el enlace de “{nombre}” sigue sin funcionando, volverá a quedar
    oculto…» — D-28; T921 lo cumple: una confirmación nueva de 404/410 la vuelve a ocultar —) o
    «Mantener oculto»; «Ya no aparece en el feed» → «Mantener visible» u
    «Ocultar» (directas, D-29); toda resolución cierra con «La marca se cerró y no se vuelve a abrir
    por esta misma causa.» (§7.1), **matizado por la excepción aprobada en la fase 5 (A3)**: una
    indisponibilidad **confirmada** posterior sí oculta y marca de nuevo. *Cubre US3 (esc. 1, 2 y 6),
    FR-013, FR-014, FR-017 (UI).*
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: el diálogo muestra las dos candidatas y la vista previa con el
    nombre de YouTube; **`mergeMedio` no se llama sin confirmar** el `ConfirmDialog` danger; «No son
    el mismo sermón» llama la resolución (`mark: possible_duplicate`) en las dos fichas sin
    confirmación extra; «Mostrar de nuevo» exige su `ConfirmDialog` y «Mantener oculto/visible» es
    directo (cada acción envía **su `mark`**); con las dos marcas activas la fila ofrece las dos
    resoluciones y cada una resuelve **solo su eje**; el foco vuelve a la fila
    al cerrar; con 0 pares **abiertos** ninguna fila ofrece «Revisar duplicado».
  - **Criterio de terminado**: `quickstart.md` §5 ejecutable en el navegador **sin depender de la
    red** (con el par sembrado) y US3 completa.
  - **Commit sugerido**: `feat(frontend): union de duplicados con vista previa y resolucion de marcas`

---

## Fase 9 — E2E

> Toda la Fase 9 exige el backend completo (T928) integrado y la base de F9 migrada. Los datos
> **se siembran por SQL** (`mediosSeed.ts`, patrón nuevo declarado en RG9-7: F9 no tiene alta
> manual) y **ningún e2e pulsa «Buscar episodios nuevos»** (quedaría conmutado por la red real y
> dispararía probadores contra datos sembrados): la importación se cubre en Go con fakes/fixtures
> (T920/T921) y en el recorrido manual de `quickstart.md` §3–§4. Las specs de panel hacen login: se
> ejecutan **por tandas** respetando el **límite de intentos de login** de F2 (ver T944).

- [ ] T939 · `mediosSeed.ts` — siembra **por SQL** (patrón nuevo, RG9-7) · `[frontend]`

  - **Archivos**: `frontend/e2e/mediosSeed.ts` (NUEVO; usa el contenedor `db` —
    `docker compose exec -T db psql -U app -d app` — con el mismo SQL documentado en
    `quickstart.md` §5/§8).
  - **Qué hace**: fixtures reutilizables por las 4 specs e2e, **sembrados directamente en
    PostgreSQL** (F9 no puede crear fichas por API salvo importando de la red real): (a) par
    **«posible duplicado»** (base `visible` + adicional `hidden`, ambas con `duplicate_of_id`
    recíproco — marca **abierta** — y sus `media_episode_sources`; quickstart §5); (b) ficha con **los
    dos enlaces** (`origin='both'`, dos fuentes) y ficha con **un solo enlace** (video y audio por
    separado); (c) ficha **sin carátula** (marcador neutro) y con descripción **sin** versión en
    inglés (fallback); (d) ficha **oculta** por el equipo (SC-002/SC-007); (e) ficha marcada
    **«ya no aparece en el feed» que sigue VISIBLE** (`availability_mark='missing_from_feed'`
    abierta, `availability_result='available'`) y otra
    con `inconclusive` (SC-014: nunca oculta por falta de certeza); (f) **8 publicables** para el
    tope 5 del bloque (SC-015) y el escenario **0 publicables** (redirección, SC-016); (g) cuentas
    de panel: una con `medios`, una con otros permisos de módulo y una sin permisos (SC-009);
    (h) **estado resultante de la precedencia A3** (fase 5): ficha que resolvió «mantener visible» y
    que después quedó **oculta** con `availability_mark='unavailable'` **abierta**
    (`availability_result='unavailable'`) — para que T941 verifique la presentación y «Mostrar de
    nuevo»; la precedencia en sí la fija T921 —.
    Creación/limpieza determinista entre tandas. *Soporta US1–US6 y SC-002/SC-003/SC-009/SC-014/
    SC-015/SC-016.*
  - **Pruebas incluidas** (§III): — (fixture e2e). **Cómo se verifica**: las 4 specs la consumen sin
    declarar datos propios; los datos se crean/limpian de forma determinista entre tandas; el SQL
    coincide con el de `quickstart.md` §5/§8 (fuente única de los escenarios) y **respeta los
    `CHECK`** del modelo de marcas (los comprueba T914).
  - **Criterio de terminado**: fixtures listas y documentadas (qué caso cubre cada una) y el patrón
    de siembra por SQL declarado en la evidencia del PR (Complexity 2). **Desacoplada (corrección
    M3 de la fase 5)**: depende **solo** del esquema y sus `CHECK` (**T901/T902/T914**) y del SQL de
    `quickstart.md` §5/§8 — **no** de los componentes del panel (T936–T938) —, de modo que el grupo
    P5 puede arrancar antes.
  - **Commit sugerido**: `test(e2e): fixtures de medios sembrados por SQL`

- [ ] T940 · E2E Playwright `medios-publica.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/medios-publica.spec.ts` (NUEVO; usa `helpers.ts` y `mediosSeed.ts`).
  - **Qué hace**: recorrido del visitante (US1, US4, US5) según `quickstart.md`
    §1/§4/§6/§7/§9/§11: sin sesión ve las fichas con carátula (local), nombre, fecha y descripción
    corta del más reciente al más antiguo; fichas con **dos enlaces** y con **uno solo** (sin líneas
    vacías); **sin descripción → sin esa línea**; **«Cargar más»** acumula tandas y desaparece al
    final; ocultas/inestables **invisibles** (SC-002: ni ocultas, ni absorbidas, ni «posible
    duplicado» abierto —pero sí «ya no aparece en el feed» visible, FR-017—); **0 peticiones a
    terceros** (interceptor de red: 0 a `youtube.com`/`spotify.com`/`anchor.fm`/`i.ytimg.com`/
    `cloudfront.net`; las carátulas son `/api/v1/media/…`, SC-005); **enlaces anunciados** (texto
    visible + `target="_blank"` + `rel="noopener noreferrer"` + `aria-label` con servicio/salida/
    pestaña nueva, SC-006) que abren en pestaña nueva; **redirección a `/` con 0 publicables** (URL
    `/`, contenido de la portada, **sin destello** de sección vacía ni aviso en ningún instante, y
    «Atrás» sin bucles por el `replace`, SC-016); inglés con fallback (contenido sin inglés en
    español) y vuelta a español (**`title`/fecha/enlaces idénticos**: SC-011); 3 anchos
    (320/768/1280, SC-017) sin desplazamiento horizontal; solo prédicas y podcasts (SC-013).
    *Cubre SC-001, SC-002, SC-005, SC-006, SC-011, SC-013, SC-016, SC-017.*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde sobre la
    base con las fixtures de T939; el resto de e2e (F3/F4/F6/F7) sigue pasando sin tocar (y
    `/eventos` conserva su `HiddenNotice`).
  - **Criterio de terminado**: US1/US4/US5 y la privacidad/enlaces (SC-005/SC-006) cubiertos de
    extremo a extremo.
  - **Commit sugerido**: `test(e2e): seccion publica de medios (enlaces anunciados, 0 terceros y redireccion)`

- [ ] T941 · E2E Playwright `medios-panel.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/medios-panel.spec.ts` (NUEVO).
  - **Qué hace**: recorrido del panel (US2, US3) sobre datos **sembrados por SQL** (RG9-7) según
    `quickstart.md` §2/§5/§6/§7/§8/§12: login con permiso `medios`; **ocultar** → desaparece de
    `/medios` y del bloque → **mostrar** → reaparece en su lugar por fecha; **ajustar la descripción
    (es/en)** y verla publicada (y **sobrevivir** a la siguiente búsqueda: SC-007 — la sincronización
    real no se ejecuta en e2e; la garantía la fija T920/T921 y aquí se comprueba el ciclo de
    edición —); **unir** un par marcado con **vista previa y doble confirmación** → una ficha con los
    dos enlaces, eje de duplicados despejado y absorbida fuera del listado; **resolver «no son
    duplicados»** →
    ambas visibles y sin marca activa; marcas de disponibilidad con sus explicadores, **resoluciones
    por eje** (cada acción envía su `mark`; con las dos marcas activas se resuelven por separado) y
    **no reapertura** de una marca resuelta (la resolución se registra y la fila queda «Resuelta»);
    el **estado de la precedencia A3** (fixture (h) de T939: resuelta «mantener visible» y después
    oculta con «no disponible» **abierta**) se presenta con su marca y su **«Mostrar de nuevo»**
    funciona — la precedencia en sí la fija **T921** (el sync no corre en e2e) —;
    denegación cruzada (cuenta sin `medios` —incluida una con otros permisos de módulo—: `403`
    forzado por API y mensaje claro en la interfaz, SC-009) y **todo registrado en la auditoría** de
    F2 con los 6 códigos `media.*` y sus etiquetas legibles (SC-010). **Nunca pulsa «Buscar episodios
    nuevos»** (RG9-7). *Cubre SC-003, SC-007, SC-009, SC-010.*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde con las
    cuentas de T939; la auditoría muestra los códigos esperados con su etiqueta humana
    (`ACTION_LABELS` de T931); los logins se ejecutan por tandas (T944).
  - **Criterio de terminado**: US2/US3 cubiertas de extremo a extremo con el permiso verificado en
    servidor y la unión/resoluciones probadas en el flujo real.
  - **Commit sugerido**: `test(e2e): gestion de medios en el panel, marcas, permisos y auditoria`

- [ ] T942 · E2E Playwright `medios-accesibilidad.spec.ts` (teclado y foco) · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/medios-accesibilidad.spec.ts` (NUEVO; patrón de
    `ministerios-accesibilidad.spec.ts` / `ofrendas-accesibilidad.spec.ts`).
  - **Qué hace**: recorrido por **teclado** de `/medios`, sus fichas, los **enlaces externos** y el
    bloque de la portada (FR-030, SC-008, quickstart §13): el primer `Tab` enfoca el enlace de salto
    al contenido; la travesía alcanza navegación, fichas, botones de enlace (activables con `Enter`/
    `Espacio`), «Cargar más», «Ver todo», selector de idioma y tema en orden comprensible; **foco
    visible en cada paso** (anillo de F10); **sin trampas de foco** (`Shift+Tab` sale de cualquier
    panel); cada **enlace externo** anuncia servicio, salida del sitio y pestaña nueva (visible y
    `aria-label`); **la redirección de `/medios` sin publicables no atrapa el foco** (el visitante
    termina en una portada navegable, P9-16/RG9-12). Se comprueba en **tema claro y oscuro** (F10),
    con `prefers-reduced-motion` y con **zoom 200 %** (US5 esc. 4). *Cubre SC-008 y FR-028…FR-030
    (teclado y foco; la checklist WCAG completa la cierra T944).*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde sobre la
    base con fixtures publicadas y, con 0 publicables, la redirección deja el foco en una página
    operable; el `aria-label` de cada enlace contiene servicio + «pestaña nueva» + «sales del sitio»
    (y su equivalente inglés).
  - **Criterio de terminado**: el e2e de teclado de la sección, las fichas, los enlaces y el bloque
    existe y corre en `make e2e` junto a las otras specs.
  - **Commit sugerido**: `test(e2e): accesibilidad por teclado de medios (SC-008)`

- [ ] T943 · E2E Playwright `portada-medios.spec.ts` · `[frontend]` `[P]` (P5)

  - **Archivos**: `frontend/e2e/portada-medios.spec.ts` (NUEVO).
  - **Qué hace**: bloque de la portada (US6, SC-015) según `quickstart.md` §10: con 8 publicables
    muestra **exactamente los 5 más recientes** en el orden de la sección; con menos, exactos, **sin
    huecos ni marcadores vacíos**; con **0 publicables el bloque no existe** (sin título, sin
    mensaje) y la portada queda coherente; **posición**: inmediatamente después del bloque de
    ofrendas y diezmos (Q12) y antes de la siguiente sección de F3; cada ficha compacta lleva sus
    **enlaces externos anunciados** (salen del sitio igual que en la sección, FR-004/FR-005); el
    botón **«Ver todo»** lleva a `/medios` en un solo paso; 0 ocultos ni duplicados abiertos
    (SC-002); 320 px sin desplazamiento horizontal (SC-017) y recorrido por teclado con foco
    visible. *Cubre US6, SC-015 y los aspectos de portada de SC-013/SC-016/SC-017.*
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**:
    `npx playwright test e2e/portada-medios.spec.ts` en verde; los e2e de la portada de F3/F4/F6/F7
    (`portada-publica`, `portada-proximamente`, `portada-ministerios`, `portada-ofrendas`) siguen
    pasando sin tocar.
  - **Criterio de terminado**: US6/SC-015 cubiertos de extremo a extremo en la posición acordada.
  - **Commit sugerido**: `test(e2e): bloque de medios de la portada`

---

## Fase 10 — Cierre

- [ ] T944 · Verificación de cierre · `[infra]`

  - **Archivos**: ninguno (evidencia en el PR). *No se toca* `.github/workflows/ci.yml` ni ningún
    archivo del kit. `docker-compose.yml`, `Makefile` y `proyecto.mk`: **sin cambios** (F9 no añade
    servicios). `.env.example`: **sin cambios** (F9 no añade variables — P9-18 —; solo se verifica
    que ninguna hiciera falta).
  - **Qué hace**: ejecuta la verificación completa de F9 y la registra como evidencia: `make ci`
    (**lint + test + security**; `make db-migrate` al preparar el entorno, quickstart §0); **cobertura
    ≥ 80 % en `internal/medios/service*.go`** (`go test -cover`); migraciones `000014`/`000015`
    **up → down → up** (y `permissions` sigue en **10** con `medios`); `make sqlc-verify` y
    `make api-gen` **sin deriva**; `make e2e` **por tandas** respetando el **límite de intentos de
    login** de F2 (las specs del panel de F9 y de F2/F4/F6/F7 hacen login: se ejecutan por tandas —o
    con las cuentas y ventanas que no agoten el límite— y la salida de cada tanda se adjunta);
    **dependencias nuevas: 0** (verificado: `go.mod` solo promueve `golang.org/x/text` y
    `golang.org/x/sync` de indirectas a directas **sin cambiar versiones**; `package.json` intacto;
    `govulncheck`/`npm audit` del CI en verde); `.env.example` sin cambios confirmado; recorrido
    manual de `quickstart.md` §0–§14 (mapa §14 (SC-001…SC-017) y títulos de §1–§13 (US/FR)),
    incluidos §0.1 (captura de fixtures) y §3 (primera búsqueda con los feeds reales: el dato de
    referencia es **10 fichas** —8 unidas y 2 con un solo enlace—, comprobación no regla) y §8 (la
    regla del «estante de 15» con la siembra determinista); `GET /healthz` responde como en F1 y una
    ruta inexistente devuelve `404` con `ErrorEnvelope` (§8.1.9) — y el **matiz del disparador de la
    comprobación** (corrección B4 de la fase 5): la indisponibilidad solo se detecta en lo que **sale
    de la ventana del feed**; un enlace borrado que **aún** está en la ventana se publica hasta que
    salga de ella (con los feeds congelados puede tardar indefinidamente) y **eso no es un fallo** —.
    **SC-008 y SC-012 son verificaciones manuales POST-DESPLIEGUE** (corrección B3 de la fase 5: la
    spec los sitúa «después del despliegue» y **aquí no se dan por hechos**): esta tarea **prepara el
    protocolo y la plantilla del informe** (`pruebas-usabilidad-SC-012.md`, estilo F6/F7: guion de
    **3 tareas sin ayuda** —encontrar una prédica; saber que el enlace sale del sitio hacia
    YouTube/Spotify en pestaña nueva **antes** de activarlo; abrir la grabación y volver al sitio—,
    **≥ 6 personas** —2 por franja 18–35 / 36–59 / 60+, incluidas con poca experiencia en internet—,
    checklist WCAG 2.1 AA de quickstart §13 y **cronómetro de SC-001** en la tarea 1) y adjunta al PR
    el protocolo con la **pendencia explícita**: la ejecución con personas y con **lector de pantalla
    real** (SC-008) se cierra **tras el despliegue**, observada por `qa-tester` con el humano y
    **firmada por el humano** (el informe se registra después en
    `specs/009-medios-predicas-y-podcasts/pruebas-usabilidad-SC-012.md`). Lo automatizable (e2e de
    teclado, checklist con `qa-tester` sobre el entorno) sí se ejecuta aquí. Las dos pruebas quedan
    fuera de `make ci`/`make e2e`.
  - **Pruebas incluidas** (§III): — (es la ejecución de todas las anteriores). **Cómo se verifica**:
    cada comando anterior en verde y su salida adjunta al PR; el mapa de criterios → secciones queda
    recorrido completo (mapa §14 para SC y los títulos de §1–§13 para US/FR) y el **protocolo** de
    SC-008/SC-012 adjunto con su **pendencia post-despliegue explícita** (el informe con los tiempos
    cronometrados de SC-001 se cierra tras el despliegue, con firma del humano).
  - **Criterio de terminado**: `make ci` y `make e2e` (por tandas) en verde; SC-001…SC-017
    evidenciados (con foco explícito en SC-005 —0 terceros—, SC-007 —la sincronización no pisa lo del
    equipo, **salvo la ocultación por indisponibilidad confirmada** (precisión de la fase 5)—,
    SC-014 —regla del «estante de 15»— y SC-016 —redirección sin destello—); **SC-008/SC-012**
    quedan como **protocolo preparado + ejecución post-despliegue** (no se dan por hechos); el **cambio
    acotado RG9-6** de T932 y el **patrón de siembra por SQL** de T939 quedan registrados en la
    evidencia.
  - **Commit sugerido**: — (sin cambios de código)

- [ ] T945 · Documentación y PR · `[infra]`

  - **Archivos**: `README.md`, `CHANGELOG.md` (o notas de entrega) y `docs/` (EDITADOS por
    `documentador`); PR en GitHub (creado por `devops`).
  - **Qué hace**: cierra la entrega con `documentador` (README del proyecto y notas de entrega: qué
    agrega F9, la sección `/medios` «Medios», la importación **a demanda** desde los dos feeds con el
    botón «Buscar episodios nuevos» (sin cron/worker/cola), una entrada por sermón con **unión por
    título normalizado + mismo día**, el permiso existente `medios` **sin migración de permisos** y
    **sin cambio de nombre** (el catálogo sigue en 10), las tablas `media_items`/
    `media_episode_sources`/`media_sync_runs` y las migraciones `000014`/`000015`, el contrato
    **0.9.0**, la **carátula copiada al sitio** (0 recursos de terceros; y el dato verificado de
    RG9-12: el podcast trae la **misma carátula** para todos sus episodios — arte del programa —, que
    se informa al cliente), la **regla del «estante de 15»** (FR-017: salir del feed no oculta; solo
    un 404/410 confirmado oculta y la comprobación fallida nunca oculta), la **redirección
    silenciosa** de `/medios` sin nada publicable (patrón de `/ministerios` y `/ofrendas`, distinto
    de `/eventos`), el bloque de la portada (`GET /api/v1/portada/medios`, tope 5, tras ofrendas), la
    **siembra e2e por SQL** (RG9-7) y las **precisiones de la puerta 2** y defaults aplicados
    (marcas: solo «posible duplicado» sin resolver bloquea; enlace muerto de dos → ficha entera
    oculta; sin descripción → línea omitida; mismo día en `America/Costa_Rica` sin margen; nombre del
    video al unir). *Cubre el cierre de US1–US6/FR-001…FR-030 como entregable.*
  - **Pruebas incluidas** (§III): — (documentación y proceso). **Cómo se verifica**: la documentación
    refleja el comportamiento real (la prueba es el `quickstart.md` ejecutado en T944); el PR
    referencia este `tasks.md` con las tareas marcadas, registra los cambios acotados (T932/T939) y
    pasa por la revisión final (`qa-tester`, `revisor-codigo`, `seguridad`).
  - **Criterio de terminado**: PR abierto y **aprobado por el humano** (regla 1: merge y producción
    requieren aprobación explícita). El merge lo ejecuta el orquestador con `devops`.
  - **Commit sugerido**: `docs: notas de entrega de F9 (medios: predicas y podcasts)`

---

## Estrategia de pruebas (resumen ejecutable)

Cobertura exigida: **80 % en `service/`** (§III). Comandos: `go test ./...` ·
`go test -tags=integration ./...` · `npm test -- --run` · `make e2e` · `make ci`.

| Capa | Tipo | Qué verifica F9 | Dónde |
|---|---|---|---|
| `platform/audit` | Unitaria | Los 6 códigos `media.*` en el registro cerrado; objetivo `content` con `targetLabel` | `internal/platform/audit/*_test.go` (T906) |
| `medios/titulo` | Unitaria (tabla de casos) | El **par real** de títulos normaliza igual; tildes/`ñ`/`¿¡`/emojis; espacios y puntuación; «Lección»=`Leccion`; `diaEnLugar` con la frontera real `Fri, 03 Jul 2026 01:22:48 GMT` → **2026-07-02** en `America/Costa_Rica` | `internal/medios/titulo_test.go` (T907) |
| `medios/feeds` | Unitaria sobre **fixtures** | Atom de YouTube (`videoId`, título, enlace, fecha RFC3339, carátula, descripción larga) y RSS del podcast (`guid`, `pubDate` `GMT`, `itunes:image`, HTML de la descripción → texto plano); campos ausentes → defaults; enlace de host ajeno → item rechazado con aviso; XML roto → error claro (nunca item a medias); caracteres especiales intactos | `internal/medios/feeds_test.go` + `testdata/` (T908) |
| `medios/cliente_http` | Unitaria (fake `RoundTripper`) | La **fábrica única** de la red saliente (corrección A1, T907A): redirect a host fuera de la allowlist rechazado; IP privada/loopback/metadata (`169.254.169.254`) rechazada tras resolver DNS; cuerpo **> tope** cortado (`LimitReader`, nunca se lee entero); `http://` rechazado; `User-Agent` propio y timeouts. **La misma tabla se ejecuta sobre los tres consumidores** (feeds, probadores y carátulas) | `internal/medios/cliente_http_test.go` (T907A) + los `*_test.go` de T908–T910 |
| `medios/disponibilidad` | Unitaria (fake `RoundTripper`) | YouTube: `200`→disponible, `404`→no disponible, `400`/`500`/timeout→**inconcluso**; podcast: `200`→disponible, `404`/`410`→no disponible, `504`/timeout→inconcluso; **solo el confirmado oculta** | `internal/medios/disponibilidad_test.go` (T909) |
| `medios/caratula` | Unitaria | Host no permitido, `http`, redirect a host ajeno, IP privada, HTML/SVG renombrado y `> UPLOAD_MAX_BYTES` rechazados; descarga correcta → `Save` con bytes intactos; fallo → ficha sin carátula (nunca aborta) | `internal/medios/caratula_test.go` (T910) |
| `service` (dominio `medios`) | Unitaria con **fakes** (repo/feeds/probadores/store) | **Importación**: alta idempotente (segunda búsqueda = 0 cambios), instantánea (no refresca nada), descripción recortada (≤300, corte en frase/palabra, HTML→texto plano; **ausente** si el feed no trae — precisión 3) y protegida por `descriptionEdited`; **unión automática** (título normalizado + mismo día sin margen → 1 ficha `both` con los dos enlaces, nombre/carátula del video — precisión 5 —); **dudoso** (mismo día con títulos distintos, o mismo título con días distintos) → adicional oculta y **ambas** con `duplicate_of_id` recíproco (marca «posible duplicado» abierta); **sin candidatura** → independientes; **FR-016**: feed caído → 0 ocultas/marcadas por esa causa y resumen con `failed`; **FR-017**: sale del feed + `404` → oculta + «no disponible» (**ficha de dos enlaces: entera**, precisión 2 — el peor de sus fuentes: M2 —); sale + `200` → **visible** + «ya no aparece en el feed» **solo si TODAS sus fuentes** salieron de la ventana (M2); sale + inconcluso → **visible** + marca, contado en `availabilityInconclusive`; marca resuelta **no se reabre por la misma causa** (pero clase distinta o **confirmación nueva** manda: A3); **matriz par abierto/resuelto × enlace vivo/muerto/inconcluso** (los dos ejes conviven); `markedCount` = solo «posible duplicado» (B2); presupuesto agotado → resumen parcial sin pérdidas; **panel**: descripción válida/inválida (`descriptionEn` sin `Es` → `400` con `details`), ocultar/mostrar, unión (base=video, descripción editada manda, eje de duplicados despejado, repunte de fuentes, `origin='both'`, absorbida fuera de consultas, **0 `DELETE`**), resolver marca **por eje** (`mark` + visibilidad elegida, idempotente, `409` sin marca de esa clase); **público**: solo publicables (ni `hidden`, ni absorbidas, ni par abierto; **sí** `missing_from_feed`), orden `published_at DESC, created_at ASC, id`, tanda 12/tope 50, fallback `en→es` sin huecos, `title`/fechas/enlaces idénticos entre idiomas, tope 5 del bloque y `items` omitida con 0 | `internal/medios/service_*_test.go` (T915–T922) |
| `handler` | Unitaria con `httptest` + service falso | Cada operación: decodificación, `details` por campo en 400 (claves desconocidas del `PATCH` también), códigos 200/400/401/403/404/409/500, sobres uniformes; el DTO público **nunca** contiene `visibility`, marcas, `sources` ni fichas no publicables; `Cache-Control: no-store` en las públicas; el resumen de `buscar-nuevos` **nunca** expone texto técnico de red y es `200` aunque un feed falle | `internal/medios/handler_*_test.go` (T923–T927) |
| `repository` | **Integración** (PostgreSQL real) | `UNIQUE (provider, external_id)` + `ON CONFLICT DO NOTHING` (doble búsqueda y doble insert); `CHECK` de URL por proveedor (host ajeno, `javascript:`); `CHECK` de marcas (resolución sin marca/par; **sin** `CHECK` de simetría — C1 — y recíproco del par verificado por prueba); predicado público completo (incluida la síntesis R9-7/precisión 1: `missing_from_feed` sale, par abierto no); orden con fechas iguales y `LIMIT/OFFSET`; `RepointMediaItemSources` en la unión; `InsertMediaSyncRun`/`GetLastMediaSyncRun`; mutación + `admin_actions` atómicas (y rollback si falla el registro); migraciones `000014`/`000015` up→down→up (con `testutil.SerializeMigrations`); el catálogo de `permissions` sigue en **10**; `IsFilePublished` con la rama de F9 (carátula publicable sí; oculta/absorbida/par abierto no) | `internal/medios/repository_*_test.go` (T912–T914) |
| `cmd/api` | Humo de composición | Rutas nuevas con su sobre y permiso, `/healthz` intacto, `404` `not_found` en rutas inexistentes y **403 cruzado** por permiso | `cmd/api/main_test.go` (T928) |
| Frontend | Unitaria (Vitest + MSW) | Pública: ficha con 1 y 2 enlaces (sin líneas vacías), marcador neutro sin carátula, **sin descripción sin esa línea**, etiquetas de enlace anunciadas, «Cargar más» acumula tandas y se agota con `total`, fallback `en→es` visible, `title`/fecha/enlaces idénticos en ambos idiomas, navegación y bloque ocultos sin publicables, **redirección de `/medios` sin destello** (esqueleto; `<Navigate replace>`; **ante error NO redirige**), 0 recursos de terceros. Panel: botón desactivado durante la búsqueda + `role="status"` + resumen (incluidos feeds fallidos y sin segmentos en 0), filtros de marca con conteos, descripción con errores por campo (`descriptionEn` sin `Es`), `UnirDuplicadosDialog` con vista previa y confirmación (no llama sin confirmar), resoluciones de marca («Mostrar de nuevo» con `ConfirmDialog`), `RequirePermission(MEDIOS)`, accesos rápidos, **`ACTION_LABELS`** con etiqueta humana para los 6 códigos; y las **2–4 pruebas actualizadas** de permisos/accesos (RG9-6) | `frontend/src/features/*/*.test.tsx` (T929–T938) |
| E2E pública | Playwright (local) | Recorrido de `quickstart.md` §1/§11: fichas con dos enlaces y con uno solo, descripción corta y su línea omitida, orden del más reciente al más antiguo, «Cargar más», ocultas/inestables invisibles (pero «ya no aparece en el feed» visible), inglés con fallback y vuelta a español, **0 peticiones a terceros**, enlaces anunciados y en pestaña nueva, **redirección a `/` sin nada publicable** (sin destello, «Atrás» sin bucles), 320/768/1280 | `frontend/e2e/medios-publica.spec.ts` (T940) |
| E2E panel | Playwright (local) | Sobre datos **sembrados por SQL** (RG9-7): ocultar → desaparece de la sección y del bloque → mostrar; ajustar descripción (es/en) y verla publicada; **unir** un par marcado con vista previa y doble confirmación → una ficha con los dos enlaces, marca despejada, absorbida fuera del listado; **resolver «no son duplicados»** → ambas visibles y sin marca activa; marcas de disponibilidad y no-reapertura; denegación cruzada (cuenta sin `medios`: 403 forzado por API) y auditoría de todo (SC-009/SC-010) | `frontend/e2e/medios-panel.spec.ts` (T941) |
| E2E accesibilidad | Playwright (local) | Teclado por `/medios`, fichas y **enlaces externos** (foco visible, sin trampas, etiqueta que anuncia servicio + salida del sitio + pestaña nueva), bloque, modo claro/oscuro, zoom 200 % y redirección sin atrapar el foco (SC-008; precedente `ministerios-accesibilidad.spec.ts`) | `frontend/e2e/medios-accesibilidad.spec.ts` (T942) |
| E2E portada | Playwright (local) | Bloque **al final, después del bloque de ofrendas** (SC-015): exactamente 5 con más publicables, exactos con menos, **sin huecos**, oculto con 0, «Ver todo» → `/medios`, enlaces que salen del sitio, 320 px | `frontend/e2e/portada-medios.spec.ts` (T943) |

**Pruebas que no faltan** (las que la spec hace críticas): **0 fichas no publicables expuestas**
(SC-002: ni ocultas, ni absorbidas, ni «posible duplicado» abierto —pero sí «ya no aparece en el
feed» visible, precisión 1/FR-017—); **0 duplicados visibles** (SC-003: el par real se une; el
dudoso se marca y la adicional no sale); **0 campos completados a mano en la importación** (SC-004);
**0 recursos de terceros** (SC-005: la carátula viaja por `/api/v1/media/…`; e2e que comprueba que la
página no pide nada a youtube/spotify/anchor/i.ytimg/cloudfront); **enlaces anunciados y en pestaña
nueva** (SC-006); **la sincronización no pisa lo del equipo** (SC-007: visibilidad, descripción
ajustada y uniones fijadas antes/después de sincronizar; **salvo la ocultación por indisponibilidad
confirmada** — precisión aprobada en la fase 5, A3 —); **fallback `en→es` sin huecos** y datos
idénticos entre idiomas (SC-011); **0 enlaces no disponibles mostrados como disponibles** y **0
ocultos por salir de la ventana del feed** (SC-014: tabla de casos del probador + e2e con la marca
«ya no aparece en el feed» **sigue visible**); **bloque de 5 al final** (SC-015); **redirección
limpia sin nada publicable** (SC-016); **320 px sin scroll horizontal** (SC-017); texto con tildes,
`ñ`, `¿¡` y emojis intacto; feed caído sin efectos colaterales (FR-016); permisos verificados en
servidor con cruce contra otros módulos (SC-009); toda operación registrada —éxito, fallo y
denegación— (SC-010).

---

## Cobertura de requisitos (US1–US6, FR-001…FR-030, SC-001…SC-017)

| Historia | Tareas | Verificación principal |
|---|---|---|
| **US1** Ver las grabaciones y abrirlas donde están | T916, T923, T929, T930, T933, T934, T940 | quickstart §1/§4 · e2e T940 (SC-001, SC-002, SC-005, SC-006) |
| **US2** Publicar un sermón sin escribir datos | T908, T920, T921, T926, T936, T941, T944 | quickstart §2–§4 (con los feeds reales) · pruebas Go del sync (SC-004) |
| **US3** Mantener el catálogo (ocultar, unir, describir) | T918, T919, T927, T936, T937, T938, T941 | quickstart §5–§8 · e2e T941 (SC-003, SC-007, SC-010) |
| **US4** Ver el contenido en inglés | T916, T917, T930, T933, T940 | quickstart §9 · e2e T940 (SC-011) |
| **US5** Cualquier dispositivo y accesible | T933, T935, T940, T942, T943, T944 | quickstart §13 (checklist) · e2e en 3 anchos y de teclado (SC-008, SC-017) |
| **US6** Últimos episodios en la portada | T917, T924, T935, T943 | quickstart §10 · e2e `portada-medios.spec.ts` (SC-015) |

| FR | Tareas | Verificación principal |
|---|---|---|
| FR-001 (una entrada por sermón; sin alojar la grabación) | T901, T911, T913, T919, T920, T933, T935 | quickstart §4 · pruebas de modelo (SC-013: nada de F4/F5/F6/F8 en la sección) |
| FR-002 (sección pública `/medios` con ficha y enlaces) | T903, T912, T916, T923, T933 | quickstart §1 · e2e T940 (SC-001) |
| FR-003 (procedencia: solo descripción y visibilidad ajustables) | T901, T915, T918, T920, T933, T936 | quickstart §1/§7 · pruebas de instantánea (SC-007) |
| FR-004 (0 terceros; carátula copiada al sitio) | T910, T912, T923, T933, T935, T940 | quickstart §1 · e2e con interceptor de red (SC-005) |
| FR-005 (enlaces anunciados, pestaña nueva) | T930, T933, T935, T940, T942 | quickstart §1/§13 · e2e de enlaces y accesibilidad (SC-006) |
| FR-006 (orden por fecha + desempate de importación; «Cargar más») | T901, T903, T916, T923, T933 | quickstart §1 · pruebas de orden y tandas |
| FR-007 (entrada única de navegación) | T930, T934 | quickstart §1 · e2e T940 (SC-016) |
| FR-008 (sin nada visible: nav/bloque ocultos y redirección) | T933, T934, T935, T940 | quickstart §11 · e2e (SC-016: sin destello, sin avisos, sin enlaces rotos) |
| FR-009 (solo lo publicable sale al visitante) | T903, T912, T916, T917, T923, T924, T940, T941 | quickstart §1/§5 · pruebas «0 no publicables expuestos» (SC-002) |
| FR-010 (los dos feeds, públicos y sin claves) | T907A, T908, T915, T920 | quickstart §3 · fixtures de los dos feeds |
| FR-011 (botón sincrónico; sin cron/worker/cola) | T915, T920, T921, T926, T936, T944 | quickstart §3 · pruebas del service (SC-004) |
| FR-012 (idempotencia y protección de lo ya hecho) | T901, T903, T913, T920, T941 | quickstart §3 (repetir la búsqueda) · pruebas de idempotencia y carrera |
| FR-013 (unión automática: título normalizado + mismo día) | T907, T919, T920, T938, T941 | quickstart §4/§5 · tabla de casos con el par real (SC-003) |
| FR-014 (duplicado dudoso: separado, adicional oculta, ambas marcadas) | T901, T919, T920, T938, T941 | quickstart §5 · e2e T941 (SC-003) |
| FR-015 (visible/oculto; la sync no cambia decisiones del equipo) | T901, T918, T919, T920, T936, T941 | quickstart §6–§7 · pruebas SC-007 |
| FR-016 (feed caído: resumen y 0 cambios) | T920, T926, T936 | quickstart §8 · pruebas con `FeedSource` que falla (SC-007) |
| FR-017 (regla del «estante de 15» + disponibilidad fail-safe) | T901, T907A, T909, T918, T919, T921, T936, T938, T941 | quickstart §8 · tabla de casos + pruebas (SC-014) |
| FR-018 (resumen de la búsqueda y `lastSync`) | T901, T913, T921, T926, T936 | quickstart §3 · pruebas del resumen |
| FR-019 (panel reducido: sin alta, sin edición del feed, sin borrado) | T912, T918, T919, T925, T927, T936, T937, T938 | revisión de `revisor-codigo` (0 rutas de alta/edición de feed/borrado) · e2e T941 |
| FR-020 (descripción corta recortada, editable y protegida) | T901, T918, T920, T937 | quickstart §7 · pruebas de recorte y de protección (SC-007) |
| FR-021 (permiso `medios` existente, verificado en servidor; catálogo en 10) | T914, T928, T931, T932, T936, T941 | quickstart §2 · humo T928 (403 cruzado) · e2e T941 (SC-009) |
| FR-022 (auditoría: 6 códigos, transaccional, solo lectura) | T902, T906, T911, T914, T918, T919, T921, T922, T931, T941 | quickstart §12 · pruebas transaccionales · e2e T941 (SC-010) |
| FR-023 (validación completa; datos nunca código) | T908, T915, T918, T920, T925, T937 | quickstart §7 · pruebas de validación y de contenido hostil |
| FR-024 (panel en español) | T930, T936, T937, T938 | revisión de `revisor-codigo` (0 cadenas del panel por `t()`) |
| FR-025 (bilingüe con fallback `en→es` solo en descripción) | T901, T915, T916, T917, T930, T933, T940 | quickstart §9 · tabla de casos del service (SC-011) |
| FR-026 (carátula/nombre/fecha/enlaces idénticos entre idiomas) | T916, T933, T940 | quickstart §9 · e2e que compara valores entre idiomas (SC-011) |
| FR-027 (bloque de portada: tope 5, posición, oculto sin nada) | T903, T917, T924, T935, T943 | quickstart §10 · pruebas del service (SC-015) · e2e de portada |
| FR-028 (F10 aplicado a sección y bloque) | T933, T935 | revisión de `revisor-codigo` sobre `HomePage` (una inserción) y `PublicLayout` sin cambios |
| FR-029 (responsivo desde 320 px, navegación simple) | T933, T935, T940, T943 | quickstart §13 · e2e en 320/768/1280 (SC-017) |
| FR-030 (accesibilidad WCAG 2.1 AA, claro y oscuro) | T933, T935, T942, T944 | quickstart §13 (checklist) · e2e de teclado T942 (SC-008 manual tras despliegue) |

| SC | Tareas | Verificación principal |
|---|---|---|
| SC-001 (ve los episodios y llega a YouTube en < 30 s) | T933, T940, T944 | quickstart §1 · e2e T940 (contenido; **el e2e no cronometra**) · el umbral **< 30 s se cronometra** en el protocolo de usabilidad de SC-012 (**T944**, tarea 1 con cronómetro) |
| SC-002 (0 ocultos y 0 «posible duplicado» abiertos visibles; «ya no aparece» sí) | T912, T916, T917, T923, T924, T939, T940 | pruebas «0 no publicables expuestos» (service/handler/repository) · e2e T940 |
| SC-003 (0 repeticiones del mismo sermón; dudosos marcados) | T907, T919, T920, T938, T941 | tabla de casos de unión con el par real · e2e T941 |
| SC-004 (publicar un sermón sin escribir ningún dato) | T920, T936, T944 | quickstart §3 (una pulsación) · pruebas del sync (0 campos escritos) |
| SC-005 (0 recursos de terceros al abrir `/medios` y la portada) | T910, T933, T935, T940 | e2e con interceptor de red (0 peticiones a youtube/spotify/anchor/i.ytimg/cloudfront) |
| SC-006 (enlaces anunciados y en pestaña nueva; 0 rotos) | T930, T933, T935, T940, T942 | e2e comprueba etiqueta + `target`/`rel`/`aria-label`; enlaces validados por host al importar |
| SC-007 (la sincronización no pisa lo del equipo; **salvo la ocultación por indisponibilidad confirmada** — precisión de la fase 5, A3 —) | T918, T919, T920, T921, T941 | pruebas que fijan visibilidad/descripción/unión antes y después; el caso A3 («resuelta visible + 404 confirmado → oculta + marca nueva») en **T921**; feed caído → 0 ocultas/marcadas |
| SC-008 (accesibilidad WCAG 2.1 AA con lector de pantalla real) | T942, T944 | e2e de teclado T942 · checklist manual (T944, tras el despliegue, protocolo F6/F7) |
| SC-009 (100 % de operaciones verifican `medios` en el servidor) | T922, T928, T931, T941 | humo T928 (403 cruzado) · e2e T941 (cuentas cruzadas) |
| SC-010 (100 % de operaciones registradas; 0 registros editables) | T911, T914, T918, T919, T921, T922, T941 | quickstart §12 · integración transaccional · e2e T941 |
| SC-011 (inglés 100 % traducido, 0 huecos; datos idénticos) | T916, T917, T930, T933, T940 | pruebas de i18n y fallback · e2e T940 (comparación de valores) |
| SC-012 (≥ 90 % de personas encuentran una prédica y acceden; identifican la salida) | T944 | prueba manual con personas **post-despliegue** (protocolo en `pruebas-usabilidad-SC-012.md` preparado en T944, con cronómetro de SC-001; ejecución y firma del humano tras el despliegue — B3 —) |
| SC-013 (0 eventos/actividades/noticias/grupos/ministerios en la sección ni el bloque) | T933, T935, T940, T943 | revisión de contenido e2e · pruebas de modelo (sin columnas ajenas) |
| SC-014 (0 enlaces no disponibles mostrados; 0 ocultos por salir del feed) | T909, T921, T939, T941 | tabla de casos del probador (solo 404/410 confirman) · e2e con «ya no aparece en el feed» **visible** |
| SC-015 (bloque con exactamente los últimos 5 al final, sin huecos, oculto con 0) | T917, T935, T943 | pruebas del service (tope 5, orden) · e2e `portada-medios.spec.ts` |
| SC-016 (0 secciones vacías; navegación sin rotos; `/medios` redirige limpio) | T933, T934, T935, T940, T943 | pruebas de `useSiteLinks`/página/bloque · e2e (URL `/`, sin destello, «Atrás» sin bucles) |
| SC-017 (usable desde 320 px en teléfono/tableta/computadora) | T933, T935, T940, T943 | e2e en 3 anchos (320/768/1280) + checklist |

También quedan cubiertos los **Edge Cases** de la spec sin FR propio: **títulos que no coinciden
carácter por carácter** (T907, T920) · **mismo elemento ya importado** (T913, T920) · **feed que no
responde** (T920, T926, T936) · **ventana del feed** (T921, T939, T941; **matiz B4 de la fase 5**: la
comprobación solo mira lo que **sale** de la ventana — un enlace muerto aún dentro de ella se publica
hasta que salga: no es un fallo —) · **enlace confirmado muerto, también con dos enlaces** (T909,
T921 — precisión 2, y el **peor** de sus fuentes gana: M2 —) · **par abierto/resuelto × enlace
vivo/muerto/inconcluso** (T914, T921: los dos ejes de marcas **conviven** sin romper el par —
C1/A3 —) · **marca resuelta que podría reabrirse** (T921: misma clase sin confirmación nueva no
reabre; clase distinta o confirmación de 404/410 **manda** — A3 —) · **comprobación fallida** (T909,
T921, T941) · **episodio con un solo formato** (T916, T933, T940) · **sin carátula → marcador
neutro** (T910, T933, T935) · **descripción del feed muy larga** (T920: nunca se publica completa) ·
**descripción ajustada que manda** (T918, T920, T941) · **oculto que sigue en el feed** (T920) ·
**texto solo en español navegando en inglés** (T916, T933, T940) · **tildes/`ñ`/`¿¡`/emojis
íntegros** (T908, T915, T918, T937) · **fecha en el límite del día** (T907, T916) · **dos elementos
con la misma fecha** (T916: desempate FIFO, default #11) · **nada visible** (T933, T934, T935) ·
**muchos episodios** (T916: tandas de 12, tope 50, sin límite artificial) · **dos personas en el
panel a la vez** (T913, T920: `UNIQUE` + última escritura completa) · **operación sin permiso**
(T928, T936, T941) · **operación no completada sin registro y registro sin operación** (T914, T918,
T919) · **búsqueda interrumpida** (T921: resumen parcial sin pérdidas) · **datos hostiles del feed**
(T908, T920: HTML→texto plano, títulos gigantes recortados, fechas inválidas → item descartado con
aviso).

**Cambios acotados y documentados registrados**: **T932** actualiza las pruebas de frontend que
fijan la lista de permisos y accesos (`features/roles/permissions.test.ts`,
`features/roles/roles.test.tsx`, `lib/permissions.test.ts`, `features/panel/panel.test.tsx`) al
**activar** `medios` (**RG9-6**: cambio de producto autorizado, **no** debilitamiento — solo
recuentos/listas/sellos —). **T939** introduce el patrón de **siembra e2e por SQL**
(`e2e/mediosSeed.ts`, **RG9-7/Complexity 2**: F9 no tiene alta manual; si `revisor-código` prefiere
un script SQL versionado en `backend/`, es un refactor acotado). El backend **no** toca pruebas de
F2: el catálogo de permisos sigue en **10**. F9 **no** promueve piezas al inventario único y solo
**amplía** `IsFilePublished` con una rama aditiva (T903).

---

## Precisiones de la puerta 2 aplicadas y puntos abiertos menores

**Las cinco precisiones aprobadas (2026-10-17) van fijadas por prueba** (cambiarlas es un cambio
acotado): (1) solo «posible duplicado» sin resolver bloquea el público; «ya no aparece en el feed»
sigue visible (T912/T916/T921/T940); (2) un enlace confirmado muerto oculta la **ficha entera**
(T921); (3) sin descripción en el feed → campo ausente y línea omitida (T920/T916/T933); (4) mismo
día en `America/Costa_Rica` **sin margen**, y la duda se marca en vez de unir (T907/T920); (5) al
unir manda el nombre del **video** (T919/T920/T938).

**Y una precisión aprobada en la fase 5 (2026-10-17)**, tras el `analyze.md` del `revisor-código`:
**la indisponibilidad confirmada manda** (A3) — la comprobación que confirma que un enlace ya no está
oculta la ficha **aunque el equipo hubiera resuelto antes «mantener visible»** (o «Mostrar de
nuevo»); queda **marcada** y el equipo puede volver a mostrarla —. Va fijada por prueba en **T921**
(caso «resuelta visible + 404 confirmado → oculta + marca nueva» y matriz de marcas) y presentada en
**T941**; **SC-007** de `spec.md` la recoge en su única frase tocada.

Puntos que quedan **a la vista del humano** (no bloquean; son revisables y acotados):

1. **D-30 ⟲ (iconos por marca)**: no se implementan; la píldora con palabra + explicador cumple
   «no solo color» (FR-030). Si se aprueban, es presentación en `MarcaDeRevision` (T936).
2. **D-10 ⟲ (filtro «Solo lo pendiente de revisión»)**: se implementa como filtro de vista sobre
   los parámetros del contrato (T936). Retirarlo no toca la API ni el servidor.
3. **RG9-2 (videos «no listados» en el feed de YouTube)**: incógnita no cerrable desde el equipo
   (evidencia 3:1 a favor de que el feed solo publica públicos, pero una fuente discrepa). La
   mitigación es de coste cero y ya está en el alcance: revisión del equipo tras cada búsqueda +
   «ocultar» con un clic + no-reimportación. Se informa en el cierre (T945).
4. **RG9-12 (carátula del podcast compartida)**: los 9 episodios de solo audio repiten la carátula
   del programa porque es lo que trae el feed (FR-003). Se publica así y se informa al cliente; si
   quiere arte por episodio, lo publica en Spotify (fuera de F9).
5. **Orden de ejecución que conviene ver**: la fusión del contrato (T904) va en el primer grupo
   paralelo —antes que la mayor parte del backend— para que el frontend (T905→T929) avance al mismo
   tiempo; y las dos tareas de `service_sync.go` (T920/T921) y las tres de `handler_admin.go`
   (T925–T927) son **secuenciales por compartir archivo**, aunque sus grupos de pruebas sean
   independientes: se reparten así para dejar cada tarea por debajo de 2 h y cada commit compila.

---

## Correcciones del analyze aplicadas (fase 5, 2026-10-17)

Resumen operativo (la narrativa completa está en `plan.md`, sección «Resolución de los hallazgos del
analyze (2026-10-17)»): **C1** → marcas en **dos ejes** (P9-19; T901, T914, T919–T921; sin `CHECK`
de simetría; matriz «par abierto/resuelto × enlace vivo/muerto/inconcluso» probada) · **A1** →
**fábrica única** de cliente HTTP (P9-20; **T907A**) y la misma tabla de barreras en
T908/T909/T910 · **A2** → `OrigenBadge`/`FuentesList` **recuperados** en T936 (D-32 matizado) ·
**A3** → precedencia fijada en T921/T941 y **SC-007 matizado en `spec.md`** (única frase tocada) ·
**M1** → interfaces declaradas **una sola vez** (T908–T910; T915 solo `Repository`) · **M2** →
regla de agregación por ficha (**el peor** de sus fuentes gana; `missing_from_feed` solo con
**todas** fuera de ventana) · **M3** → T939 **desacoplada** de T936–T938 (depende de
T901/T902/T914) · **B1** → T904 **no commitea** `schema.d.ts` (lo hace T905, con comprobación de
deriva) · **B2** → `markedCount` = **solo** «posible duplicado» · **B3** → SC-008/SC-012: protocolo
en T944 y ejecución **post-despliegue** con firma del humano · **B4** → matiz del disparador (solo
lo que sale de la ventana del feed) en `quickstart.md` §8 y en el recorrido de T944.
