# Implementation Plan: Medios — prédicas y podcasts (F9)

**Branch**: `009-medios-predicas-y-podcasts` | **Date**: 2026-10-17 | **Spec**: [spec.md](./spec.md)
(aprobada **v2** por el humano el 2026-10-17 —decisión 27 del roadmap—: importación desde los
feeds a demanda con el botón «Buscar episodios nuevos», **una entrada por sermón** con video y
audio unidos, enlaces externos **sin contenido de terceros**, panel reducido (buscar, ocultar/
mostrar, unir, descripción corta), estado **visible/oculto**, regla del «estante de 15» con
comprobación de disponibilidad (FR-017) y enlaces en pestaña nueva anunciada; 6 US, FR-001…FR-030,
SC-001…SC-017).

**Input**: `specs/009-medios-predicas-y-podcasts/spec.md` (fuente de verdad) ·
`.specify/memory/constitution.md` · `AGENTS.md` · `docs/tecnico/arquitectura.md` (reglas R1–R8 y
convenciones §8.1) · **F2 terminada** (patrón de permisos por módulo, sesión, CSRF, `authz`,
`apperr`, `validate`, `paginate`, auditoría y catálogo con `medios` ya sembrado —`000002`—, que
**no cambia**) · **F3 terminada** (bilingüe `*_es`/`*_en`, `?lang` con fallback, «ocultar secciones
vacías», `media`/`platform/storage`) · **F4 y F6 terminadas** (bloque de portada con endpoint
propio y tope 5, `sort_order`/`ORDER BY` determinista, `IsFilePublished` por ramas, **`/ministerios`
redirigiendo a la portada sin aviso** —decisión 20b—) · **F7 terminada** («**precedente más
directo**»: delta de contrato, 8 códigos `donation.*`, texto editable con estado propio, bloque de
portada tras ministerios, redirección P7-21) · **F10 terminada** (sistema de diseño, modo oscuro,
movimiento con `prefers-reduced-motion`, navegación y `useSiteLinks`: se **aplica**, no se
redefine) · código real de `backend/` y `frontend/` · skills `go-backend`, `postgres-db`,
`react-frontend`.

## Summary

F9 es la sección **«Medios»** del sitio público (`/medios`): las prédicas y los podcasts de la
iglesia, **importados de los feeds que el cliente ya publica** —el feed Atom del canal de YouTube
`@IglesiaSimienteSanta` y el RSS del podcast «Servicio dominical Iglesia Simiente Santa» (Spotify
for Creators/Anchor)— y presentados como **una ficha por sermón** con **carátula, nombre y
descripción corta** y sus **enlaces externos** («Ver en YouTube» y/o «Escuchar en Spotify»): el
visitante **sale del sitio** en una pestaña nueva anunciada y el sitio **no carga ni reproduce nada
de terceros** (Q15, FR-004/FR-005, SC-005). El equipo **no escribe ningún dato**: pulsa
**«Buscar episodios nuevos»** en el panel y el servidor agrega solo lo que falte, **une el video y
el audio del mismo sermón** cuando la correspondencia es segura (título normalizado + mismo día,
FR-013) y deja lo dudoso **marcado y fuera del público** para que el equipo una con un clic
(FR-014). El panel solo ofrece **buscar, revisar, ocultar/mostrar, unir duplicados, ajustar la
descripción corta y resolver marcas** (FR-019): sin alta manual, sin edición de los datos del feed
y sin borrado. La portada cierra con el bloque de los **últimos 5** y «Ver todo» (FR-027).

Técnicamente F9 añade: **tres tablas** —`media_items` (la ficha por sermón, con `visibility`
`visible`/`hidden`, marcas de revisión con resolución y el puntero de unión),
`media_episode_sources` (una fila por elemento de feed, **`UNIQUE (provider, external_id)`** = la
idempotencia de la búsqueda) y `media_sync_runs` (el registro de cada búsqueda, FR-018)—, la
**primera red saliente del backend** —leer los dos feeds, copiar las carátulas a
`platform/storage` y comprobar la disponibilidad de los enlaces que salen de la ventana del feed,
todo **acotado, con allowlist de hosts y regla a prueba de fallos** (research R9-2/R9-8/R9-9)—, un
**parser propio** de Atom/RSS con `encoding/xml` tras una interfaz falsable (**0 dependencias
nuevas**, R9-3), la normalización de títulos con `golang.org/x/text/unicode/norm` (ya en `go.mod`
como indirecta) y la regla de unión de research R9-4, el patrón de autorización
`AdminChain("medios")` (**sin permisos nuevos**: catálogo sigue en 10), **6 códigos de auditoría
`media.*`** (migración `000015`), una rama aditiva en `IsFilePublished` para servir las carátulas
copiadas y la sección pública + bloque de portada + módulo de panel con el sistema de F10.

**Ningún módulo nuevo** ni en backend ni en frontend (research R9-13): `encoding/xml`, `net/http`,
`time` y `golang.org/x/text`/`x/sync` que ya están en `go.mod`.

**Cuatro puntos donde la spec se contradice o no llega** quedan resueltos como **defaults
documentados** y se elevan al humano en «Preguntas para el humano» (al final): (1) qué se publica
con marcas abiertas —FR-009/SC-002 vs. default #2 vs. FR-017—, resuelto por la única lectura que
las cumple (R9-7); (2) Q16 con una ficha de dos enlaces: un enlace confirmado muerto **oculta la
ficha entera** (lectura literal); (3) descripción corta cuando el feed no trae descripción; (4) la
frontera del «mismo día» entre feeds. Además la spec señala ⚠️ el **default #7** (nombre que manda
al unir = el del video).

## Technical Context

**Language/Version**: Go 1.27 (sin cambios) · TypeScript 5.x `strict` sobre Node 22 · React 19.

**Primary Dependencies**: **ninguna nueva** (research R9-13). Parser de feeds con `encoding/xml`
(stdlib, R9-3); HTTP saliente con `net/http` (`http.Client` propio con timeouts); normalización de
títulos con `golang.org/x/text/unicode/norm` (**ya presente** en `backend/go.mod` v0.42.0 como
indirecta de `pgx`; se promueve a directa sin cambiar versiones); paralelismo acotado con
`golang.org/x/sync` (también ya presente) o `sync.WaitGroup`; carátulas con `platform/storage`
(F3/F4/F6); fechas con `time` + `APP_TZ`. En frontend: React Router, TanStack Query, React Hook
Form + Zod, Tailwind, Vitest + Testing Library + MSW y Playwright (R7-16, sin cambios).

**Storage**: **PostgreSQL 16** (tablas `media_items` + `media_episode_sources` + `media_sync_runs`
+ ampliación del `CHECK` de `admin_actions`; migraciones `000014`/`000015`, ver
[data-model.md](./data-model.md)) · Redis 7 **sin tocar** (sesión de F2) · **carátulas copiadas**
en `platform/storage` (JPEG/PNG/WebP por firma binaria, nombres `img_<uuid>.<ext>`, servidas por
`GET /api/v1/media/{fileName}` con la rama nueva de `IsFilePublished`). **Ni audio ni video se
guardan jamás** (FR-001). Capa de datos sqlc (D-A3): consultas en
`queries/media_episodes.sql`, `media_episode_sources.sql`, `media_sync_runs.sql` + rama de
`media.sql`; código generado commiteado; auditoría con `InsertAdminAction` **dentro de la
transacción** de cada mutación. **Sin variables de entorno nuevas** (los dos feeds son constantes
documentadas del dominio —FR-010 fija los orígenes—; los timeouts y topes son constantes
revisables).

**Testing**: estrategia por capa de F1/F2/F3/F4/F6/F7 ampliada: `go test` (unitarias con **fakes**
de repositorio, de `FeedSource`, de probadores y de `storage.Store` —patrón
`fake_repository_test.go`— y **fixtures XML reales** de los dos feeds), `go test -tags=integration`
(repository contra PostgreSQL real —migraciones up/down/up, `UNIQUE` de idempotencia, predicado
público, `IsFilePublished`—), Vitest + Testing Library + MSW y Playwright (e2e pública, de panel,
de accesibilidad y de la portada). **Salvedad hermética declarada** (RG9-7): F9 **no tiene alta
manual** (FR-019), así que los e2e **siembran por SQL** (helper nuevo `e2e/mediosSeed.ts` sobre el
contenedor `db`) y **no** pulsan «Buscar episodios nuevos» contra la red real: la importación se
cubre en las pruebas Go con fakes y en el recorrido manual del [quickstart](./quickstart.md).

**Target Platform**: Docker Compose local + GitHub Actions (CI del kit, sin tocar). Sin despliegue
a producción (fuera de alcance de F9).

**Performance Goals**: `GET /api/v1/medios` < 500 ms con BD sana (una consulta paginada,
`no-store`, tanda de 12) · `GET /api/v1/portada/medios` < 500 ms (`LIMIT 5` en servidor) ·
`GET /api/v1/admin/medios` paginado · **«Buscar episodios nuevos»** dentro del ciclo
petición/respuesta con **presupuesto de 90 s** (2 feeds en paralelo 10 s c/u, ≤ 30 carátulas y
≤ 30 comprobaciones con 4 en vuelo; una búsqueda normal: 5–20 s; la primera importación completa:
~15 carátulas, ~20 s) — research R9-9.

**Constraints**: stack fijo (constitución) · capas `handler → service → repository` (§II) ·
contrato OpenAPI antes que el código (§II) · migraciones versionadas e inmutables (§VI) ·
validación de toda entrada en backend (§IV, CWE-20) · SQL siempre parametrizado (§IV, CWE-89) ·
**el contenido de los feeds son datos, nunca código** (§IV, CWE-79: descripciones a texto plano,
sin `dangerouslySetInnerHTML`, enlaces validados por esquema/host) · **la primera red saliente del
backend se acota** (SSRF: allowlist de hosts, sin IPs privadas, sin redirects fuera de la
allowlist, `io.LimitReader` + firma binaria — R9-8) · authn/authz/CSRF en servidor (§IV, CWE-862)
· **sin proceso en segundo plano** (FR-011: la búsqueda es a demanda y sincrónica) · dependencias
minimizadas (**0 nuevas**) · archivos del kit no editables (regla 10) · **la spec no se reabre**:
los cuatro huecos de arriba son defaults documentados que se elevan, no interpretaciones
silenciosas.

**Scale/Scope**: 7 operaciones REST (**2 públicas** · 5 de panel) sobre **3 tablas nuevas** + la
ampliación del `CHECK` de `admin_actions` (6 códigos) + una rama de `IsFilePublished`; 1 dominio
backend nuevo (`medios`) y `media`/`platform` sin cambios de comportamiento (solo `platform/audit`
crece con 6 constantes); 2 migraciones (`000014`, `000015`) y **ninguna de permisos**; frontend:
página pública nueva en `/medios` + **bloque de medios al final de la portada** (tras el bloque de
ofrendas de F7) + feature de panel `medios` + ajustes de router/navegación/permisos/i18n.

## Constitution Check

*GATE: debe pasar antes de la investigación y re-verificarse tras el diseño.*

| Principio | Evaluación | Resultado |
|---|---|---|
| §I La spec manda | El plan implementa FR-001…FR-030 e US1–US6 sin interpretarlos; cada uno tiene fila en «Cobertura de requisitos». El modelo aprobado (decisión 27) se implementa literal: importación a demanda, una entrada por sermón, enlaces externos, panel reducido, visible/oculto y regla del «estante de 15» (FR-017). Nada de Out of Scope se construye: **sin** reproducción dentro del sitio, **sin** cron/worker/cola, **sin** alta manual ni edición de datos del feed ni borrado, **sin** filtros/búsqueda en la sección, **sin** descargas ni transcripciones, **sin** estadísticas, **sin** traducción del nombre, **sin** permisos nuevos (catálogo sigue en 10), **sin** tocar el resto de la portada ni F10B. Los cuatro huecos señalados arriba son defaults documentados que se elevan al humano («Preguntas para el humano»), no reglas inventadas. | ✅ |
| §II Arquitectura | Monorepo; capas `handler → service → repository` sobre `internal/platform/` (R1–R8); API REST JSON documentada en `backend/api/openapi.yaml` con el delta redactado **antes** que el código (`contracts/openapi.yaml`, §8.1.5); **0 dependencias nuevas** (research R9-13); la única superficie nueva —red saliente— vive en el dominio tras interfaces (`FeedSource`, `AvailabilityProber`, `CoverFetcher`) falsables en pruebas. | ✅ |
| §III Pruebas | Toda tarea llevará sus pruebas (se exigirá en `tasks.md`); cobertura ≥80 % en `service/`; repository contra PostgreSQL real (`//go:build integration`, con `testutil.SerializeMigrations` —decisión 25—); importación/unión/marcas/disponibilidad probadas con **fakes y fixtures** sin red; frontend con Vitest + Testing Library + MSW; flujos críticos con Playwright (sección, «Cargar más», ocultar/mostrar, unir, marcas, permisos, bloque de portada y **redirección a la portada sin nada publicable**). La salvedad hermética de los e2e (siembra por SQL, sin botón de búsqueda en e2e) está declarada en RG9-7 y cubierta por el quickstart manual. | ✅ |
| §IV Seguridad | SQL solo parametrizado vía sqlc (CWE-89); validación de toda entrada en backend (CWE-20: descripción límites + español por campo + `visibility` catálogo; los datos del feed se validan igual —URL por esquema/host, títulos recortados—); **el contenido externo es datos, nunca código** (CWE-79: descripciones del feed a **texto plano** sin interpretar HTML, sin `dangerouslySetInnerHTML`, emojis/tildes intactos); **SSRF acotado** en la descarga de carátulas y los probadores (https + allowlist de hosts + sin IPs privadas + sin redirects fuera + `LimitReader` + firma binaria en `storage.Detect`); operaciones de panel con sesión + guard + CSRF + permiso **`medios`** verificado **en el servidor** (CWE-862); solo lo publicable sale al público (SC-002/SC-014: ni ocultas, ni absorbidas, ni duplicados abiertos); secretos: ninguno nuevo (`govulncheck`/`npm audit` en el CI del kit). | ✅ |
| §V Calidad | `gofmt`/`go vet`/`golangci-lint` sin errores; TS `strict` sin `any`; errores envueltos con `%w` y traducidos solo en `WriteError`; nombres en inglés dentro del código (dominio `medios` sigue la pauta: carpeta de dominio en español + tablas en inglés). | ✅ |
| §VI Base de datos | 2 migraciones versionadas con `up`/`down` completos (`000014`, `000015`); ninguna migración existente se modifica (**ni `000002`**: no hay permisos nuevos); tablas con `id`, `created_at`, `updated_at`; FK e índices explícitos; `CHECK`/`UNIQUE` en la base (catálogos de `provider`/`origin`/`visibility`/`review_mark`, forma de URL por proveedor, español por campo, `UNIQUE (provider, external_id)`, `btrim`). | ✅ |
| §VII Observabilidad | Logs `log/slog` con `request_id` (incluidos los fallos de feeds/carátulas/probadores, **sin cuerpos de petición**); auditoría duradera ampliada a F9 en `admin_actions` (FR-022, 6 códigos) + `media_sync_runs` (FR-018); `/healthz` intacto y sin sesión; config por variables de entorno (**sin variables nuevas**); `docker compose up` levanta todo (sin servicios nuevos). | ✅ |
| §VIII Gobierno | **Este plan requiere aprobación humana** antes de `/speckit.tasks` y de cualquier código (regla 1), incluidos los puntos de «Preguntas para el humano». Sin despliegue a producción. Quien escribe no aprueba: todo pasa por `qa-tester`, `revisor-codigo` y `seguridad`. | ✅ |

**Sin violaciones que justificar** → «Complexity Tracking» solo registra desviaciones de
convenciones internas (nombrar `visibility` en vez de `publication_state`, siembra e2e por SQL,
`coverUrl` sin `alt` editable, `media.sync` como un solo código de auditoría por búsqueda y
`/api/v1/portada/medios` registrado por el dominio `medios`).

## Decisiones técnicas (resumen; análisis completo en `research.md`)

Numeración **P9-1…P9-18**, propia de este plan (no confundir con P7-1… de F7, P6-1… de F6,
R9-1…R9-14 de `research.md` ni RG9-… de «Riesgos»).

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| **P9-1** | **Dominio `internal/medios/`** con **tres tablas** —`media_items` (ficha por sermón), `media_episode_sources` (una fila por elemento de feed) y `media_sync_runs` (registro de búsquedas)— + rama aditiva de `IsFilePublished` | Key Entities define ficha, procedencia por feed, marcas, unión y resumen de búsqueda con ciclos de vida distintos; la fuente por proveedor da unicidad y procedencia (R9-11) | Una tabla con columnas `youtube_id`/`podcast_id` (la unión reescribiría la ficha y perdería el título de cada feed); JSON de fuentes (pierde `UNIQUE` y sqlc); tabla de marcas aparte (un estado por ficha basta, R9-7) |
| **P9-2** | **Superficie HTTP**: `GET /api/v1/medios` (sección paginada), `GET /api/v1/portada/medios` (tope 5) y `/api/v1/admin/medios` (listado, `POST buscar-nuevos`, `PATCH {id}`, `POST {id}/unir`, `POST {id}/marcas/resolver`), públicos con `Cache-Control: no-store` y `?lang=es\|en` | Sigue el patrón literal de `portada/proximamente` y `portada/ofrendas` (F4/F7) y el panel reducido de FR-019; público y panel se llaman `medios` (Q7/Q13: aquí **sí** coinciden, a diferencia de F7) | `GET /api/v1/medios` sin paginación (recortaría el catálogo: R9-10); `POST` por acción con verbo en la raíz (`/ocultar`, `/mostrar`) en vez de `PATCH visibility` (F7 resuelve visibilidad en el `PATCH` y registra dos auditorías); endpoint público de detalle (la ficha no tiene página propia: la lista es la sección) |
| **P9-3** | **Búsqueda sincrónica** en la petición `POST …/buscar-nuevos` con presupuesto de **90 s**, timeouts por petición saliente (10/10/5 s), 4 descargas/comprobaciones en vuelo y topes duros (30/30/30) | FR-011 y constitución: **sin** cron, worker ni cola; el resumen parcial ante el agotamiento del presupuesto nunca pierde ni oculta nada (R9-9) | Proceso en segundo plano/cola (prohibido); presupuestos cortos que abortarían la primera importación; hacer la búsqueda en `GET` (muta el catálogo: es `POST`) |
| **P9-4** | **Parser propio** de Atom (YouTube) y RSS 2.0 (podcast) con `encoding/xml`, detrás de la interfaz `FeedSource`, con **fixtures XML reales** commiteados y fake en pruebas | Dos formatos fijos y verificados (R9-5); **0 dependencias nuevas**; control total de HTML→texto plano, hosts de enlaces y fechas `GMT`/RFC3339 (R9-3) | `gofeed` (arrastra `goquery`/`x/net` y da universalidad que no se usa); regex sobre el XML; parsear el HTML de las páginas de YouTube/Spotify |
| **P9-5** | **Unión por título normalizado + mismo día** en `America/Costa_Rica` (NFKD + sin diacríticos + mayúsculas + sin puntuación + espacios colapsados, con `x/text/unicode/norm`); candidato sin ambas condiciones = **dudoso** (par marcado, adicional oculta) | Literal FR-013/FR-014 y verificado con el par real de títulos (R9-4); conservador en la dirección correcta (SC-003: el visitante nunca ve un sermón dos veces) | Título exacto (los reales no coinciden); similaridad difusa (umbral arbitrario); margen ±1 día (la spec dice «mismo día»; queda como pregunta menor); unir por descripción |
| **P9-6** | **Disponibilidad**: YouTube por **oEmbed** (`200`/`404`/resto), podcast por **GET del enlace del episodio** (`200`/`404`-`410`/resto); solo un 404/410 **confirmado** oculta y marca «no disponible»; todo lo demás solo marca y **nunca** oculta | Gratuitos, sin claves ni cuota, verificados con recursos reales e inexistentes (R9-2); oEmbed de Spotify **no** resuelve nuestros enlaces (504 verificado) | oEmbed de Spotify para el audio (504); `HEAD` de `watch?v=` (YouTube responde 200 con error); YouTube Data API (clave + cuota); ocultar al salir del feed (vaciaría el sitio: FR-017 lo prohíbe) |
| **P9-7** | **Carátula copiada** al importar con `platform/storage` (`storage.Save`: firma JPEG/PNG/WebP + nombre generado) desde una descarga **acotada**: https, hosts permitidos (`i.ytimg.com`, `d3t3ozftmdmh3i.cloudfront.net`), sin IPs privadas ni redirects fuera, `LimitReader` con `UPLOAD_MAX_BYTES` | SC-005 (0 recursos de terceros) sin abrir SSRF con la primera red saliente del backend (R9-8); si la descarga falla: ficha sin carátula + marcador neutro | Enlazar la carátula a `i.ytimg.com` (viola SC-005); `http.Get` sin allowlist (SSRF); servir lo que el feed traiga (contenido activo) |
| **P9-8** | **Estado `visibility` (`visible`/`hidden`)** + `review_mark`/`mark_resolved_at`; **el único bloqueo público por marca es `possible_duplicate` sin resolver**; `missing_from_feed` sigue visible | Es la única lectura que cumple a la vez FR-009/SC-002, el default #2 y FR-017 (contradicción de la spec resuelta y elevada, R9-7); la no-reapertura de marcas resueltas es un índice, no una tabla | `publication_state` de F3/F4/F6 (la spec lo excluye); que toda marca bloquee (contradice FR-017); borrar la marca al resolver (se pierde la no-reapertura) |
| **P9-9** | **Instantánea al importar**: la búsqueda solo **agrega** lo que falta y aplica las reglas de FR-017; no refresca título, fecha, carátula, enlaces ni descripción | FR-012 literal («agregar solo lo que falte») + default #10 (carátula congelada) + SC-007 verificable campo a campo (R9-6); el default #1 de «regeneración» queda sin efecto, documentado | Refrescar lo no editado (títulos publicados que cambian sin aviso); guardar la descripción completa para regenerar (FR-020 prohíbe publicarla) |
| **P9-10** | **Unión por puntero** `merged_into_id` (la absorbida sale de todas las consultas; sus `sources` se repuntan a la resultante; **cero `DELETE`**), base = la entrada del **video** (nombre, carátula, fecha) y **manda la descripción editada** si solo una lo está | FR-019 literal («no hay borrado»), procedencia conservada y SC-007 (las descripciones del equipo no se pierden); el diálogo confirma el resultado (default #8: unión definitiva) | `DELETE` de la fila sobrante (incumple «no hay borrado»); copiar el audio y dejar la otra huérfana sin puntero; unión reversible con historial (fuera de alcance) |
| **P9-11** | **Idempotencia en la BD**: `UNIQUE (provider, external_id)` + `INSERT … ON CONFLICT DO NOTHING` | FR-012 «nunca dos veces» garantizado incluso con dos búsquedas simultáneas (default #12) y con la unión (los identificadores se repuntan) (R9-11) | `SELECT` previo (carrera); unicidad por título/fecha (los títulos cambian); memoria/JSON |
| **P9-12** | **Sección paginada** (`limit`/`offset`, 12 por tanda, tope 50) + **bloque de portada con tope 5 firme** en servidor, mismo esquema de ítem y `items` omitida sin contenido | «Cargar más» de Q11 + §8.1.3 (todo listado se acota) + sin límite artificial de episodios (R9-10); patrón exacto de los bloques de F6/F7 (SC-015) | Todo en una respuesta con tope 200 (recorta el catálogo); carrusel/paginación en el bloque (Out of Scope); tope configurable (Out of Scope) |
| **P9-13** | **Bilingüe** `?lang=es\|en` con fallback `en → es` **solo sobre la descripción corta** (`descriptionEs`/`descriptionEn`); nombre, fecha, carátula y enlaces son un único valor; panel en español | FR-025/FR-026/Decisión 8 y Q4 acotado; «0 campos vacíos» (SC-011) garantizado por el fallback en el servidor (patrón P7-9) | Traducir el nombre (FR-026 lo prohíbe); fallback en el cliente (analyze I2 de F3); `Accept-Language` |
| **P9-14** | **Auditoría**: 6 códigos `media.*` (migración `000015`), `InsertAdminAction` transaccional y `audit.Recorder` del dominio (patrón `donaciones/service_audit.go`); la búsqueda completa = **una** fila `media.sync` con extracto en `targetLabel` y detalle en `media_sync_runs` | FR-022 con el patrón probado de F2/F3/F4/F6/F7; distinguir quién **ocultó**, quién **unió** y quién **resolvió** una marca | Registro propio de F9; un código por elemento importado (ruido: la operación pedida por el equipo es la búsqueda); registrar tras el commit |
| **P9-15** | **Permisos**: `AdminChain("medios")` con el permiso **ya sembrado** en `000002`; **sin migración de permisos** ni cambios en F2 (el catálogo sigue en 10); frontend `MEDIOS` en `lib/permissions` y `isPermissionAvailable('medios') → true` | FR-021 literal («F9 NO crea permisos nuevos»); mismo gesto que F6/F7 | Permiso por sub-acción; renombrar el permiso (F2 lo prohíbe); derivar el acceso de `donaciones`/`ministerios` |
| **P9-16** | **Frontend**: `features/publico/medios/` (sección + bloque + ficha + `ExternalLink` anunciado) y `features/medios/` (panel); navegación desde el **documento de la sección** (`useSiteLinks`, patrón P7-8); **redirección de `/medios`** con la mecánica P7-21 (esqueleto → `<Navigate to="/" replace />` solo con éxito y 0 publicables; **ante error no se redirige**) | Una sola entrada de navegación sin enlaces rotos (FR-007/FR-008/SC-016) y el patrón de redirección ya probado en F6/F7 | Presencia desde el bloque (dejaría la navegación huérfana); `HiddenNotice` sin redirigir (patrón F4: Q10 pide redirigir); redirigir también ante error (una caída transitoria expulsaría al visitante) |
| **P9-17** | **Descripción corta**: al importar se recorta de la descripción del feed —HTML → texto plano, primeras frases, **≤ 300 caracteres** (default revisable #1/#13), corte en frontera de frase o palabra con «…»—; editable (es/en) y protegida por `descriptionEdited`; si el feed no trae descripción, el campo queda **ausente** y la ficha omite la línea (hueco documentado) | FR-020 + los ~1.500 caracteres reales de las descripciones (R9-5); «sin líneas vacías» (SC-010); el texto plano cumple §IV | Guardar la descripción completa (FR-020: no se publica); `dangerouslySetInnerHTML` (prohibido); descripción vacía como `""` (rompe los `CHECK` de 1…300 y la regla de omitidos) |
| **P9-18** | **Feeds y topes como constantes del dominio** (las dos URLs de FR-010; timeouts, allowlist de hosts y topes 30/30/30/90 s documentados y revisables); **sin variables de entorno nuevas** | FR-010 fija los orígenes; mínima superficie operativa (F7 también salió sin variables) | URLs por configuración (superficie que nadie pidió); claves de API (FR-010: sin claves ni pagos) |

## Project Structure

### Documentation (esta funcionalidad)

```text
specs/009-medios-predicas-y-podcasts/
├── plan.md              # EDITABLE (este archivo, /speckit.plan)
├── research.md          # EDITABLE (fase 0: R9-1…R9-14, con evidencia)
├── data-model.md        # EDITABLE (fase 1: media_items/media_episode_sources/
│                        #   media_sync_runs, migraciones 000014–000015, consultas)
├── quickstart.md        # EDITABLE (fase 1: cómo probar F9 en local, incluido
│                        #   cómo forzar una búsqueda y unir sin depender de la red)
├── spec.md              # APROBADA (v2) — no se edita
├── contracts/
│   └── openapi.yaml     # EDITABLE (fase 1): delta de diseño, se fusiona en backend/api/openapi.yaml
└── tasks.md             # Fase 4 (/speckit.tasks — NO lo crea este comando)
```

> **`ux.md` ya existe** (producido en paralelo por `disenador-ux` durante esta fase): fija la
> composición visual de las fichas, del bloque y del panel **sin cambiar la API ni los datos**
> (FR-003) y sus piezas coinciden con este plan (`MediosPage`, `MediosBlock`, `MedioCard`,
> `MedioItem`, `MedioLinks`, `CaratulaNeutra`, `CargarMasButton`, `BuscarEpisodios`,
> `ResumenImportacion`, `EpisodiosList`, `usePublicMedios`, `usePortadaMedios`). Los nombres de
> pieza que usa este plan son **alias provisionales**; **manda `ux.md`** para composición y
> nombres, y cualquier discrepancia se resuelve a favor de `ux.md` sin tocar el contrato (RG9-5).

### Source Code — backend

```text
backend/
├── cmd/api/
│   ├── main.go                      # EDITADO: DI del dominio medios (repository, feeds, probadores,
│   │                                #   storage.Store) + registro de /api/v1/medios,
│   │                                #   /api/v1/portada/medios y /api/v1/admin/medios
│   └── main_test.go                 # EDITADO: prueba de humo (rutas nuevas + /healthz intacto)
├── internal/
│   ├── db/
│   │   ├── queries/
│   │   │   ├── media_episodes.sql   # NUEVO: ListPublishedMediaItems/CountPublishedMediaItems
│   │   │   │                        #   (predicado público completo), ListPortadaMediaItems (LIMIT 5),
│   │   │   │                        #   ListMediaItems/CountMediaItems (panel con filtros),
│   │   │   │                        #   GetMediaItemByID, InsertMediaItem, UpdateMediaItem,
│   │   │   │                        #   RepointMediaItemSources (unión)
│   │   │   ├── media_episode_sources.sql # NUEVO: InsertMediaEpisodeSource (ON CONFLICT DO NOTHING),
│   │   │   │                        #   ListMediaEpisodeSourcesByProvider, GetMediaEpisodeSource,
│   │   │   │                        #   ListMediaItemSourcesByItem
│   │   │   ├── media_sync_runs.sql  # NUEVO: InsertMediaSyncRun, GetLastMediaSyncRun
│   │   │   ├── media.sql            # EDITADO: rama aditiva de IsFilePublished (carátula de una
│   │   │   │                        #   ficha publicable; R9-8)
│   │   │   └── audit.sql            # SIN CAMBIOS: se reutiliza InsertAdminAction (R7-10)
│   │   └── (generado por sqlc, commiteado)
│   ├── medios/                      # NUEVO — dominio F9 (receta de arq. §8)
│   │   ├── model.go                 # Episodio/Fuente/SyncRun y DTOs (etiquetas validate, Es/En solo
│   │   │                            #   en descripción, origin/visibility/marks)
│   │   ├── titulo.go                # normalizarTitulo (NFKD + sin diacríticos + mayúsculas + sin
│   │   │                            #   puntuación + espacios) y diaEnLugar (fecha → día en APP_TZ)
│   │   ├── titulo_test.go           # tabla de casos: par real de títulos, tildes/ñ/¿¡, emojis,
│   │   │                            #   espacios dobles, frontera de día 03 Jul 01:22 GMT → 02 Jul CR
│   │   ├── feeds.go                 # interfaz FeedSource + implementación (net/http con timeouts,
│   │   │                            #   User-Agent propio) y parseo Atom/RSS con encoding/xml →
│   │   │                            #   []EntradaFeed (id externo, título, enlace validado por host,
│   │   │                            #   fecha, carátula URL, descripción a texto plano)
│   │   ├── testdata/                # FIXTURES XML reales de los dos feeds (capturados en quickstart §0)
│   │   │                            #   + casos patológicos (CDATA, campos ausentes, fechas GMT/+0000,
│   │   │                            #   XML roto)
│   │   ├── feeds_test.go            # parseo de fixtures; XML roto → error claro (nunca item a medias)
│   │   ├── disponibilidad.go        # interfaz AvailabilityProber + YouTube (oEmbed 200/404/resto) y
│   │   │                            #   Podcast (GET del enlace 200/404-410/resto) → disponible /
│   │   │                            #   no_disponible / inconcluso (fail-safe)
│   │   ├── disponibilidad_test.go   # tabla de casos por probador (fake http.RoundTripper)
│   │   ├── caratula.go              # CoverFetcher: https + allowlist de hosts + sin IPs privadas +
│   │   │                            #   sin redirects fuera + LimitReader + storage.Detect/Save
│   │   ├── caratula_test.go         # host no permitido, redirect malicioso, HTML renombrado, > límite
│   │   ├── repository.go            # constructor + mapRow/params (pgtype → dominio) + withTx
│   │   │                            #   + inserción de admin_actions sobre el tx
│   │   ├── repository_episodios.go  # consultas de fichas (público/panel/mutaciones)
│   │   ├── repository_fuentes.go    # fuentes (ON CONFLICT, repunte de unión)
│   │   ├── repository_sync.go       # media_sync_runs
│   │   ├── service.go               # tipos comunes, invariantes, constantes (topes, timeouts,
│   │   │                            #   allowlist, maxPage=50, pageDefault=12, maxPortada=5)
│   │   ├── service_public.go        # GetMedios (paginación + ?lang con fallback) y GetPortadaMedios
│   │   │                            #   (tope 5 en servidor)
│   │   ├── service_admin.go         # List (con lastSync), Update (descripción/visibilidad),
│   │   │                            #   Merge (unión por puntero + repunte + despeje de marcas),
│   │   │                            #   ResolveMark (marca resuelta + visibilidad; no-reapertura)
│   │   ├── service_sync.go          # BuscarNuevos: feeds en paralelo → alta idempotente → unión
│   │   │                            #   automática/marcado dudoso → carátulas → disponibilidad de lo
│   │   │                            #   que salió del feed → resumen + media_sync_runs (presupuesto 90 s)
│   │   ├── service_audit.go         # audit.Recorder del dominio: método+ruta → media.*
│   │   ├── handler.go               # constructor + helpers HTTP (decodificar/validar/responder)
│   │   ├── handler_public.go        # GET /api/v1/medios y GET /api/v1/portada/medios
│   │   ├── handler_admin.go         # listado, buscar-nuevos, PATCH, unir, resolver marca
│   │   ├── routes.go                # RegisterPublic (2 rutas) y RegisterAdmin (AdminChain("medios"))
│   │   └── *_test.go                # service con fakes (repo/feeds/probadores/store) · handler con
│   │                                #   httptest · repository integration (//go:build integration)
│   ├── media/                       # SIN CAMBIOS de código (solo crece queries/media.sql)
│   └── platform/
│       ├── audit/audit.go           # EDITADO: + 6 códigos media.* (y su prueba del registro cerrado)
│       ├── middleware/chain.go      # SIN CAMBIOS: AdminChain("medios") reutiliza AuthzByModule
│       ├── validate/validate.go     # SIN CAMBIOS (los tags existentes bastan)
│       ├── storage/                 # SIN CAMBIOS (Save/Open/Delete/Detect tal cual)
│       └── config/config.go         # SIN CAMBIOS (sin variables de entorno nuevas)
├── migrations/
│   ├── 000014_create_media_content.up.sql / .down.sql              # NUEVO: media_items +
│   │                                                               #   media_episode_sources + media_sync_runs
│   └── 000015_extend_admin_actions_for_media.up.sql / .down.sql    # NUEVO: CHECK de auditoría +6 códigos
└── api/openapi.yaml                 # EDITADO: se fusiona el delta (info.version → 0.9.0) + make api-gen
```

### Source Code — frontend y raíz

```text
frontend/src/
├── api/
│   ├── medios.ts                    # NUEVO: getMedios(lang, limit, offset), getPortadaMedios(lang),
│   │                                #   admin: listMedios, syncMedios, updateMedio, mergeMedio,
│   │                                #   resolveMedioMark
│   └── schema.d.ts                  # REGENERADO (make api-gen; contrato 0.9.0)
├── app/
│   ├── router.tsx                   # EDITADO: '/medios' (pública) y '/panel/medios' con
│   │                                #   RequirePermission(MEDIOS)
│   ├── layout.tsx                   # EDITADO: entrada del menú del panel filtrada por permiso
│   └── guards.tsx                   # SIN CAMBIOS (RequirePermission de F2 basta)
├── features/
│   ├── publico/                     # EDITADO: sitio público (F10 aplicado, no redefinido)
│   │   ├── medios/                  # NUEVO (alias de este plan; nombres finales en ux.md §6):
│   │   │   ├── MediosPage.tsx       #   H1 + lista con «Cargar más» (tandas de 12); 0 publicables →
│   │   │   │                        #   REDIRIGE a la portada (P7-21: SectionSkeleton durante la
│   │   │   │                        #   carga y <Navigate to="/" replace /> solo con éxito y 0;
│   │   │   │                        #   ante error → estado de error con «Volver a intentar», sin redirigir)
│   │   │   ├── EpisodioCard.tsx     #   ficha: carátula (o marcador neutro), nombre, descripción
│   │   │   │                        #   corta (solo si existe) y los enlaces que existen
│   │   │   ├── ExternalLink.tsx     #   enlace externo con target="_blank" rel="noopener noreferrer"
│   │   │   │                        #   y etiqueta anunciada («Ver en YouTube · se abre en YouTube,
│   │   │   │                        #   en una pestaña nueva»; FR-005/FR-030) + texto sr-only
│   │   │   ├── MediosBlock.tsx      #   bloque de la portada: hasta 5 fichas + «Ver todo» → /medios;
│   │   │   │                        #   sin publicables → null (FR-027)
│   │   │   ├── MediosItem.tsx       #   tarjeta compacta del bloque
│   │   │   ├── hooks/usePublicMedios.ts   # TanStack Query ['medios', lang, limit, offset] (acumula
│   │   │   │                        #   tandas para «Cargar más»)
│   │   │   ├── hooks/usePortadaMedios.ts  # ['portada-medios', lang]
│   │   │   └── *.test.tsx
│   │   ├── hooks/useSiteLinks.ts    # EDITADO: SiteLinkId + 'media'; entrada «Medios» (href /medios)
│   │   │                            #   solo si el documento de la SECCIÓN trae total > 0 (P7-8/P9-16)
│   │   ├── pages/HomePage.tsx       # EDITADO: + <MediosBlock /> AL FINAL, después de <OfrendasBlock />
│   │   │                            #   (una inserción; el resto de la portada de F3/F4/F6/F7 intacto)
│   │   └── i18n/{messages,es,en}.ts # EDITADO: claves tipadas nuevas (nav, sección, enlaces, errores)
│   ├── medios/                      # NUEVO: gestión en el panel (en español, FR-024)
│   │   ├── pages/MediosPage.tsx     #   botón «Buscar episodios nuevos» + resumen + listado con
│   │   │                            #   filtros (visibilidad, marca), StatusPill de marcas y acciones
│   │   ├── components/              # BuscarNuevosButton (desactivado mientras corre), ResumenBusqueda,
│   │   │                            #   MediosList, DescripcionForm (es/en, Dialog), UnirDialog
│   │   │                            #   (confirmación explícita con el resultado), ResolverMarcaDialog,
│   │   │                            #   OrigenBadge (video/audio/ambos) y FuentesList (procedencia)
│   │   ├── hooks/                   # useMediosAdmin, useSyncMedios, useUpdateMedio, useMergeMedio,
│   │   │                            #   useResolveMark, queryKeys
│   │   ├── validation.ts            # Zod (espejo del server: 1…300 por idioma, descriptionEn exige Es)
│   │   ├── errors.ts                # details → errores por campo
│   │   ├── messages.ts              # textos del panel
│   │   └── *.test.tsx
│   ├── roles/permissions.ts         # EDITADO: isPermissionAvailable deja de marcar «medios» como
│   │                                #   «Disponible más adelante» (y sus tests, RG9-6)
│   ├── panel/components/AccesosRapidos.tsx  # EDITADO: tarjeta «Gestión de medios» filtrada por permiso
│   ├── auditoria/messages.ts        # EDITADO: ACTION_LABELS cubre los 6 códigos media.*
│   └── (resto SIN CAMBIOS)
├── lib/permissions.ts               # EDITADO: + MEDIOS = 'medios' (+ MODULE_PERMISSIONS)
└── (index.css, tailwind.config)     # SIN CAMBIOS (tokens de F10/F3)
frontend/e2e/
├── medios-publica.spec.ts           # NUEVO: visitante (US1, US4, US5) — ficha con uno/dos enlaces,
│                                    #   sin carátula → marcador neutro, «Cargar más», redirección a la
│                                    #   portada sin nada publicable, 320 px
├── medios-panel.spec.ts             # NUEVO: ocultar/mostrar, descripción, unir con confirmación,
│                                    #   resolver marcas, permisos y auditoría (US2, US3)
├── medios-accesibilidad.spec.ts     # NUEVO: teclado y foco (US5, SC-008), incluidos los enlaces
│                                    #   externos anunciados
├── portada-medios.spec.ts           # NUEVO: bloque al final de la portada (US6, SC-015)
├── mediosSeed.ts                    # NUEVO: siembra POR SQL (docker compose exec -T db psql …;
│                                    #   patrón nuevo declarado en RG9-7: F9 no tiene alta manual)
└── (helpers.ts                      # SIN CAMBIOS)

# Raíz
README.md                            # EDITADO por documentador al cerrar
.env.example                         # SIN CAMBIOS (F9 no añade variables)
docker-compose.yml / Makefile / proyecto.mk / .github/workflows/ci.yml / .githooks/  # SIN CAMBIOS (kit)
```

### Inventario único de componentes UI

F9 **no promueve ninguna pieza** al inventario único: reutiliza sin tocar `Field`, `TextArea`,
`Button`, `Notice`, `ConfirmDialog`, `Dialog`, `EmptyState`, `Table`, `Tabs`, `StatusPill`,
`Pagination` y el sistema de F10 (`Card`, `CardImage`, `SectionHeading`, `PublicButton`,
`Reveal`/`RevealGroup`, `SectionSkeleton`, `PublicLayout`, `PublicNav`, `MobileMenu`,
`FooterSiteMap`, `LanguageSwitcher`, `ThemeToggle`, `BrandLogo` —este último para el **marcador
neutro** de la ficha sin carátula —`CaratulaNeutra` en `ux.md`—, default revisable #4—). Piezas
propias de F9 (`MediosPage`, `MedioCard`, `MedioLinks`/`ExternalLink`, `MedioItem`,
`MediosBlock`, `CargarMasButton`, `BuscarEpisodios`, `ResumenImportacion`, `UnirDialog`,
`DescripcionForm`, `ResolverMarcaDialog`) viven en sus carpetas de feature —**nombres definitivos
en `ux.md` §6**—; se promueven solo si otra feature las necesita después. `HiddenNotice` **no se
toca** (F4 la sigue usando en `/eventos`; F9 **redirige**, P9-16).

## Cadena de middleware y grupos (orden; el primero es el más externo)

```text
Global:   request-id → recover → logging → CORS(con credenciales) → handler
          /api/v1/medios                     → (sin middlewares)  → handler    ← público, solo lectura
          /api/v1/portada/medios             → (sin middlewares)  → handler    ← público, solo lectura (lo registra
                                                                               `medios.RegisterPublic`, aunque el
                                                                               segmento `portada` sea el agregado de F3)
          /api/v1/admin/medios               → authn → guard de cambio de contraseña → authz("medios") → CSRF → handler
          /api/v1/admin/medios/{id}          → (misma cadena)                          → handler
          /api/v1/admin/medios/{id}/unir     → (misma cadena)                          → handler
          /api/v1/admin/medios/{id}/marcas/resolver → (misma cadena)                   → handler
```

- La superficie pública es de **solo lectura**: sin `authn`, `CSRF` ni `rate-limit` (F2 solo lo
  monta sobre superficies públicas escribibles; aquí no hay ninguna).
- El grupo de panel monta la **cadena aprobada de F2** con `AuthzByModule` tal cual (solo cambia
  el código de permiso: `medios`, ya sembrado). La denegación responde `403 forbidden` genérico y
  **queda registrada** en `admin_actions` (`result='denied'`); el `Recorder` de la cadena es el del
  dominio `medios` (resuelve sus rutas → acción `media.*`, R7-10).
- `/healthz` **no cambia**: sigue público, sin sesión y con `Cache-Control: no-store`.

## Contrato OpenAPI: dónde vive y cómo se mantiene

Igual que F1/F2/F3/F4/F6/F7 (reglas §8.1.5 y §8.1.10):

1. **Diseño (esta fase)**: `specs/009-medios-predicas-y-podcasts/contracts/openapi.yaml` — delta
   inmutable una vez aprobado el plan.
2. **Fusión**: la primera tarea de implementación lo funde en `backend/api/openapi.yaml` de forma
   **aditiva**, aplica los cambios documentados del delta (enum `action` de `AdminActionItem`
   **+6 códigos** `media.*`) y sube `info.version` de **0.8.0 a 0.9.0** (minor: operaciones y
   esquemas nuevos; el registro de `error.code` **no** cambia; `GET /api/v1/media/{fileName}` solo
   crece en prosa). Luego `make api-gen`.
3. **Consumo**: el frontend genera tipos solo desde `backend/api/openapi.yaml` (`schema.d.ts`).
4. El snapshot de `specs/` no se vuelve a editar.
5. **La prosa del catálogo de permisos NO cambia**: F2 sigue teniendo 10 permisos; `medios` ya
   estaba sembrado por `000002` (como `ministerios` en F6 y `donaciones` en F7).

## Estrategia de pruebas

Cobertura exigida: **80 % en `service/`** (§III). Comandos: `go test ./...` ·
`go test -tags=integration ./...` · `npm test -- --run` · `make e2e` · `make ci`.

| Capa | Tipo | Qué verifica F9 | Dónde |
|---|---|---|---|
| `platform/audit` | Unitaria | Los 6 códigos nuevos en el registro cerrado; objetivo `content` con `targetLabel` | `internal/platform/audit/*_test.go` |
| `medios/titulo` | Unitaria (tabla de casos) | El **par real** de títulos normaliza igual; tildes/`ñ`/`¿¡`/emojis; espacios dobles y puntuación; mayúsculas; `Lección`=`Leccion`; `diaEnLugar` con la frontera real `Fri, 03 Jul 2026 01:22:48 GMT` → **2026-07-02** en `America/Costa_Rica` | `internal/medios/titulo_test.go` |
| `medios/feeds` | Unitaria sobre **fixtures** | Atom de YouTube (videoId, título, enlace, fecha RFC3339, carátula, descripción larga) y RSS del podcast (guid, pubDate `GMT`, `itunes:image`, descripción con HTML → texto plano); campos ausentes → defaults documentados; enlace con host ajeno → item **rechazado con aviso**; XML roto → error claro (nunca item a medias); caracteres especiales intactos | `internal/medios/feeds_test.go` + `testdata/` |
| `medios/disponibilidad` | Unitaria (fake `RoundTripper`) | YouTube: `200`→disponible, `404`→no disponible, `400`/`500`/timeout→**inconcluso**; podcast: `200`→disponible, `404`/`410`→no disponible, `504`/timeout→inconcluso; **solo el confirmado oculta** (tabla de casos) | `internal/medios/disponibilidad_test.go` |
| `medios/caratula` | Unitaria | Host no permitido rechazado; `http` rechazado; redirect a host ajeno rechazado; IP privada rechazada; HTML/SVG renombrado rechazado por `storage.Detect`; > `UPLOAD_MAX_BYTES` rechazado; descarga correcta → `Save` con bytes intactos; fallo de descarga → ficha sin carátula (nunca aborta la importación) | `internal/medios/caratula_test.go` |
| `service` (dominio `medios`) | Unitaria con **fakes** (repo/feeds/probadores/store) | **Importación**: alta idempotente (segunda búsqueda = 0 cambios), instantánea (no refresca nada), descripción recortada (≤300, corte en frase/palabra, HTML→texto plano) y protegida por `descriptionEdited`; **unión automática** (título normalizado + mismo día → 1 ficha `both` con los dos enlaces, nombre/carátula del video); **dudoso** (mismo día con títulos distintos, o mismo título con días distintos) → adicional oculta y **ambas marcadas** con `duplicate_of_id`; **sin candidatura** → independientes; **FR-016**: feed caído → 0 ocultas/marcadas por esa causa y resumen con `failed`; **FR-017**: sale del feed + `404` → oculta + «no disponible»; sale + `200` → **visible** + «ya no aparece en el feed»; sale + inconcluso → **visible** + marca, contado en `availabilityInconclusive`; marca resuelta **no se reabre**; presupuesto agotado → resumen parcial sin pérdidas; **panel**: descripción válida/inválida (`descriptionEn` sin `Es` → `400` con `details`), ocultar/mostrar, unión (base=video, descripción editada manda, marcas despejadas, repunte de fuentes, `origin='both'`, absorbida fuera de consultas), resolver marca (visibilidad elegida, idempotente, `409` sin marca); **público**: solo publicables (ni `hidden`, ni absorbidas, ni `possible_duplicate` abierto; **sí** `missing_from_feed`), orden `published_at DESC, created_at ASC, id`, tanda 12/tope 50, fallback `en→es` sin huecos, `title`/fechas/enlaces idénticos entre idiomas, tope 5 del bloque y `items` omitida con 0 | `internal/medios/service_*_test.go` |
| `handler` | Unitaria con `httptest` + service falso | Cada operación: decodificación, `details` por campo en 400 (claves desconocidas del `PATCH` también), códigos 200/400/401/403/404/409/500, sobres uniformes; el DTO público **nunca** contiene `visibility`, `reviewMark`, `sources` ni fichas no publicables; `Cache-Control: no-store` en las públicas; el resumen de `buscar-nuevos` nunca expone texto técnico de red | `internal/medios/handler_*_test.go` |
| `repository` | **Integración** (PostgreSQL real) | `UNIQUE (provider, external_id)` + `ON CONFLICT DO NOTHING` (doble búsqueda y doble insert); `CHECK` de URL por proveedor (host ajeno, `javascript:`); `CHECK` de marcas (resuelta sin marca, `possible_duplicate` sin par); predicado público completo (incluida la síntesis R9-7: `missing_from_feed` sale, `possible_duplicate` abierto no); orden con fechas iguales y `LIMIT/OFFSET`; `RepointMediaItemSources` en la unión; `InsertMediaSyncRun`/`GetLastMediaSyncRun`; mutación + `admin_actions` atómicas (y rollback si falla el registro); migraciones `000014`/`000015` up→down→up (con `testutil.SerializeMigrations`); el catálogo de `permissions` sigue en **10**; `IsFilePublished` con la rama de F9 (carátula publicable sí; oculta/absorbida/duplicado abierto no) | `internal/medios/repository_*_test.go` |
| `cmd/api` | Humo de composición | Rutas nuevas con su sobre y permiso, `/healthz` intacto y `404 not_found` en rutas inexistentes | `cmd/api/main_test.go` |
| Frontend | Unitaria (Vitest + MSW) | Pública: ficha con 1 y 2 enlaces (sin líneas vacías), marcador neutro sin carátula, etiquetas de enlace anunciadas, «Cargar más» acumula tandas y se agota con `total`, fallback en→es visible, `title`/fecha/enlaces idénticos en ambos idiomas, navegación y bloque ocultos sin publicables, **redirección de `/medios` sin destello** (esqueleto; `<Navigate replace>`; **ante error NO redirige**). Panel: botón desactivado durante la búsqueda + resumen (incluidos feeds fallidos), filtros de marca, descripción con errores por campo (`descriptionEn` sin `Es`), `UnirDialog` con confirmación (no llama sin confirmar), `ResolverMarcaDialog`, `RequirePermission(MEDIOS)`, accesos rápidos, **`ACTION_LABELS`** con etiqueta humana para los 6 códigos; y las **2–3 pruebas actualizadas** de permisos/accesos (RG9-6) | `frontend/src/features/*/*.test.tsx` |
| E2E pública | Playwright (local) | Recorrido de `quickstart.md` §1: fichas con dos enlaces y con uno solo, descripción corta, orden del más reciente al más antiguo, «Cargar más», ocultas/inestables invisibles, inglés con fallback y vuelta a español, **redirección a `/` sin nada publicable** (sin destello, «Atrás» sin bucles), 320/768/1280 | `frontend/e2e/medios-publica.spec.ts` |
| E2E panel | Playwright (local) | Sobre datos **sembrados por SQL** (RG9-7): ocultar → desaparece de la sección y del bloque → mostrar; ajustar descripción (es/en) y verla publicada; **unir** un par marcado con confirmación → una ficha con los dos enlaces, marca despejada, absorbida fuera del listado; **resolver** «no son duplicados» → ambas visibles y sin marca activa; denegación cruzada (cuenta sin `medios`: 403 forzado por API) y auditoría de todo (SC-009/SC-010) | `frontend/e2e/medios-panel.spec.ts` |
| E2E accesibilidad | Playwright (local) | Teclado por `/medios`, fichas y **enlaces externos** (foco visible, sin trampas, etiqueta que anuncia servicio + pestaña nueva), modo oscuro y redirección sin atrapar el foco (SC-008; precedente `ministerios-accesibilidad.spec.ts`) | `frontend/e2e/medios-accesibilidad.spec.ts` |
| E2E portada | Playwright (local) | Bloque **al final, después del bloque de ofrendas** (SC-015): exactamente 5 con más publicables, exactos con menos, **sin huecos**, oculto con 0, «Ver todo» → `/medios`, enlaces que salen del sitio, 320 px | `frontend/e2e/portada-medios.spec.ts` |

**Pruebas que no faltan** (las que la spec hace críticas): **0 fichas no publicables expuestas**
(SC-002: ni ocultas, ni absorbidas, ni «posible duplicado» abierto —pero sí «ya no aparece en el
feed» visible, FR-017—); **0 duplicados visibles** (SC-003: el par real se une; el dudoso se marca
y la adicional no sale); **0 campos completados a mano en la importación** (SC-004); **0 recursos
de terceros** (SC-005: la carátula viaja por `/api/v1/media/…`; e2e que comprueba que la página no
pide nada a youtube/spotify/anchor); **enlaces anunciados y en pestaña nueva** (SC-006);
**la sincronización no pisa lo del equipo** (SC-007: visibilidad, descripción ajustada y uniones
fijadas antes/después de sincronizar); **fallback en→es sin huecos** y datos idénticos entre
idiomas (SC-011); **0 enlaces no disponibles mostrados como disponibles** y **0 ocultos por salir
de la ventana del feed** (SC-014: tabla de casos del probador + e2e); **bloque de 5 al final**
(SC-015); **redirección limpia sin nada publicable** (SC-016); **320 px sin scroll horizontal**
(SC-017); texto con tildes, `ñ`, `¿¡` y emojis intacto; feed caído sin efectos colaterales
(FR-016); permisos verificados en servidor con cruce contra otros módulos (SC-009); toda operación
registrada —éxito, fallo y denegación— (SC-010).

## Cobertura de requisitos (US1–US6 y FR-001…FR-030)

| Historia | Dónde se cubre | Verificación |
|---|---|---|
| **US1** — Ver las grabaciones y abrirlas donde están (P1) | `GET /api/v1/medios` (P9-2/P9-12/P9-13) + `MediosPage`/`EpisodioCard`/`ExternalLink` en `/medios` (P9-16) | quickstart §1 · e2e pública (SC-001, SC-002, SC-005, SC-006) |
| **US2** — Publicar un sermón sin escribir datos (P1) | `POST /api/v1/admin/medios/buscar-nuevos` + `service_sync.go` (P9-3/P9-4/P9-5/P9-11) + panel `features/medios/` | quickstart §2–§4 (con los feeds reales) · pruebas Go del sync (SC-004) |
| **US3** — Mantener el catálogo (P2) | `PATCH {id}`, `POST {id}/unir`, `POST {id}/marcas/resolver` (P9-8/P9-10) + `UnirDialog`/`DescripcionForm`/`ResolverMarcaDialog` | quickstart §5–§8 · e2e panel (SC-003, SC-007, SC-010) |
| **US4** — Ver el contenido en inglés (P2) | `?lang` con fallback `en→es` solo en descripción (P9-13) + i18n tipado | quickstart §9 · e2e pública (SC-011) |
| **US5** — Cualquier dispositivo y accesible (P2) | Sistema de F10 aplicado + `ExternalLink` anunciado + reglas `react-frontend` (P9-16) | quickstart §11 (checklist) · e2e en 3 anchos y de teclado (SC-008, SC-017) |
| **US6** — Últimos episodios en la portada (P2) | `GET /api/v1/portada/medios` con tope 5 (P9-12) + `MediosBlock` al final de `HomePage` (P9-16) | quickstart §10 · e2e `portada-medios.spec.ts` (SC-015) |

| FR | Dónde se resuelve en este plan | Verificación |
|---|---|---|
| FR-001 | P9-1 (ficha = `media_items` + 1–2 `media_episode_sources`; sin audio/video alojado) + P9-10 (una entrada por sermón) | quickstart §4 · pruebas de modelo (SC-013: nada de F4/F5/F6/F8 en la sección) |
| FR-002 | P9-2/P9-12 (`GET /api/v1/medios`, solo publicables) + P9-16 (`MediosPage`/`EpisodioCard`) | quickstart §1 · e2e pública (SC-001) |
| FR-003 | P9-9 (procedencia: todo del feed salvo descripción y visibilidad) + P9-17 (descripción corta) | quickstart §1 y §7 · pruebas de instantánea (SC-007) |
| FR-004 | P9-7 (carátula copiada, `/api/v1/media/…`) + P9-16 (sin reproductores; enlaces que salen) | quickstart §1 · e2e que comprueba 0 peticiones a terceros (SC-005) |
| FR-005 | P9-16 (`ExternalLink`: `target="_blank" rel="noopener noreferrer"` + etiqueta anunciada + sr-only) | quickstart §11 · e2e de enlaces y accesibilidad (SC-006) |
| FR-006 | P9-12 (orden `published_at DESC, created_at ASC, id` en `America/Costa_Rica`, tandas de 12) | quickstart §1 · pruebas de orden y «Cargar más» |
| FR-007 | P9-16 (`useSiteLinks` + entrada única «Medios» en barra, hamburguesa y pie) | quickstart §1 · e2e pública (SC-016) |
| FR-008 | P9-16 (redirección `/medios` con la mecánica P7-21) + P9-12 (bloque oculto con 0) | quickstart §10 · e2e (SC-016: sin destello, sin avisos, sin enlaces rotos) |
| FR-009 | P9-8 (predicado público: `visible` + no absorbida + sin `possible_duplicate` abierto) | quickstart §1 y §6 · pruebas «0 no publicables expuestas» (SC-002) |
| FR-010 | P9-4/P9-18 (los dos feeds de la spec, públicos y sin claves; F9 no administra cuentas) | quickstart §3 · fixtures de los dos feeds |
| FR-011 | P9-3 (botón, sincrónico, sin cron/worker/cola/despliegue previo, repetible) | quickstart §3 · pruebas del service (SC-004) |
| FR-012 | P9-11 (`UNIQUE (provider, external_id)` + `ON CONFLICT DO NOTHING`) + P9-9 (instantánea) | quickstart §3 (repetir la búsqueda) · pruebas de idempotencia y carrera |
| FR-013 | P9-5 (unión automática por título normalizado + mismo día) | quickstart §4 · tabla de casos con el par real (SC-003) |
| FR-014 | P9-5/P9-8/P9-10 (dudoso → adicional oculta, ambas marcadas, unión con un clic) | quickstart §5 · e2e panel (SC-003) |
| FR-015 | P9-8 (visibility que la sync no toca) + P9-9/P9-17 (`descriptionEdited`) | quickstart §6–§7 · pruebas SC-007 |
| FR-016 | P9-3/P9-6 (feed caído → solo resumen `failed`; 0 ocultas/marcadas; se recupera al repetir) | quickstart §8 · pruebas con `FeedSource` que falla (SC-007) |
| FR-017 | P9-6 (probadores 200/404/resto) + P9-8 (marcas y no-reapertura) + P9-2 (`marcas/resolver`) | quickstart §8 · tabla de casos + pruebas (SC-014) |
| FR-018 | P9-2 (`ResumenBusqueda`) + P9-14 (`media_sync_runs`/`lastSync`) | quickstart §3 · pruebas del resumen |
| FR-019 | P9-2 (solo buscar/listar/ocultar/mostrar/unir/descripción/marcas) + P9-10 (sin DELETE) | revisión de `revisor-codigo` (0 rutas de alta/edición de feed/borrado) · e2e panel |
| FR-020 | P9-17 (recorte ≤300, editable, protegido) | quickstart §7 · pruebas de recorte y de protección (SC-007) |
| FR-021 | P9-15 (`AdminChain("medios")`, permiso ya sembrado; catálogo sigue en 10) | quickstart §2 · pruebas de handler (403 forzado, cruce de permisos) · e2e (SC-009) |
| FR-022 | P9-14 (6 códigos `media.*`, transaccional, denegaciones registradas, solo lectura) | quickstart §12 · pruebas de service/repository (SC-010) |
| FR-023 | P9-17 (límites, español por campo) + §IV (texto plano, sin ejecución) | quickstart §7 · pruebas de validación y de contenido hostil |
| FR-024 | Panel en español (convención de F2/F3/F4/F6/F7; sin `i18n/` en la gestión) | revisión de `revisor-codigo` (0 cadenas de panel pasadas por `t()`) |
| FR-025 | P9-13 (fallback `en→es` solo en descripción; nunca campos vacíos) | quickstart §9 · tabla de casos del service (SC-011) |
| FR-026 | P9-13 (carátula/nombre/fecha/enlaces/formatos idénticos en ambos idiomas) | quickstart §9 · e2e que compara valores entre idiomas (SC-011) |
| FR-027 | P9-12 (endpoint propio, tope 5 firme, mismo DTO de ítem, «Ver todo», posición al final tras ofrendas, oculto con 0) | quickstart §10 · pruebas del service (SC-015) · e2e de portada |
| FR-028 | P9-16 (F10 aplicado a sección y bloque; navegación de F10) | revisión de `revisor-codigo` sobre `HomePage` (una inserción) y `PublicLayout` sin cambios |
| FR-029 | P9-16 (componentes de F10, `toc-targets` ≥ 44 px, ficha legible en 320 px) | quickstart §11 · e2e en 320/768/1280 (SC-017) |
| FR-030 | Accesibilidad: teclado, foco, contraste claro/oscuro, zoom 200 %, lector de pantalla, enlaces anunciados (P9-16) | quickstart §11 (checklist WCAG 2.1 AA) · e2e de teclado (SC-008 manual tras despliegue) |

## Métricas del plan (coherentes con los SC de la spec)

| SC | Qué se mide | Cómo se mide en F9 | Umbral |
|---|---|---|---|
| SC-001 | Todo visitante ve los episodios sin cuenta y llega a YouTube en < 30 s | e2e pública: fichas completas en la primera carga + enlace directo (**el e2e no cronometra**; el umbral se cronometra en el protocolo de SC-012, como en F3/F4/F6/F7) | 100 % · < 30 s (SC-012) |
| SC-002 | 0 ocultos y 0 marcas «posible duplicado» abiertas visibles | Pruebas «0 no publicables expuestos» (service + handler + repository) y e2e con fichas ocultas/absorbidas/marcadas en BD | 0 |
| SC-003 | 0 repeticiones del mismo sermón; los dudosos quedan marcados | Tabla de casos de unión con el par real (8/2 de los feeds) + e2e de unión y de resolución | 100 % / 0 |
| SC-004 | Publicar un sermón sin escribir ningún dato | quickstart §3 (una pulsación) + pruebas del sync (0 campos escritos por el equipo) | 0 campos |
| SC-005 | 0 recursos de terceros al abrir `/medios` y la portada | e2e que intercepta la red del navegador y comprueba **0 peticiones** a youtube.com/spotify.com/anchor.fm/i.ytimg.com/cloudfront (la carátula es `/api/v1/media/…`) | 0 |
| SC-006 | Enlaces anunciados (servicio, salida del sitio, pestaña nueva) y 0 rotos | `ExternalLink` con etiqueta + `target="_blank"`; e2e comprueba la etiqueta y el atributo; los enlaces vienen validados del feed (host por proveedor) | 100 % / 0 |
| SC-007 | La sincronización no pisa lo del equipo | Pruebas que fijan visibilidad/descripción/union antes de sincronizar y comparan después; feed caído → 0 ocultas/marcadas | 0 cambios |
| SC-008 | Accesibilidad (WCAG 2.1 AA, claro y oscuro) con lector de pantalla real | Checklist de quickstart §11 + e2e de teclado; **verificación manual tras el despliegue** (protocolo equivalente al de F6/F7) | 0 críticos |
| SC-009 | 100 % de operaciones verifican `medios` en el servidor | Pruebas de handler por operación (403) + e2e con cuentas cruzadas (otros permisos de módulo, sin `medios`) | 0 sin permiso / 100 % denegadas |
| SC-010 | 100 % de operaciones registradas; 0 registros editables | quickstart §12 + pruebas transaccionales + contrato sin escritura sobre `admin_actions` | 100 % / 0 |
| SC-011 | Interfaz en inglés 100 % traducida, 0 campos vacíos; datos no traducibles idénticos | Pruebas de i18n (claves tipadas) + tabla de casos de fallback + e2e en inglés que compara `title`/fecha/enlaces | 100 % / 0 alteraciones |
| SC-012 | ≥ 90 % de personas encuentran una prédica y acceden; ≥ 90 % identifican la salida del sitio | Prueba manual con personas (protocolo en `quickstart.md` §11, estilo F6/F7): ≥ 6 personas, guion de 3 tareas, informe `pruebas-usabilidad-SC-012.md` | ≥ 90 % |
| SC-013 | 0 eventos/actividades/noticias/grupos/ministerios en la sección ni en el bloque | Revisión de contenido e2e (solo fichas de medios) + pruebas de modelo (sin columnas ajenas) | 0 |
| SC-014 | 0 enlaces no disponibles mostrados; 0 ocultos por salir de la ventana del feed | Tabla de casos del probador (solo 404/410 confirman) + e2e con marca «ya no aparece en el feed» que **sigue visible** + pruebas de la regla a prueba de fallos | 0 / 0 |
| SC-015 | Bloque con exactamente los últimos 5 al final, 0 ocultos, sin huecos, «Ver todo», oculto con 0 | Pruebas del service/repository (tope 5, orden) + e2e `portada-medios.spec.ts` (posición tras ofrendas) | 100 % / 0 |
| SC-016 | 0 secciones vacías; navegación sin enlaces rotos; `/medios` redirige limpio | Bloques/entrada omitidos sin contenido + e2e (URL termina en `/`, sin destello ni aviso, «Atrás» sin bucles) | 0 |
| SC-017 | Usable en teléfono/tableta/computadora desde 320 px | e2e en 3 anchos (320/768/1280) + checklist | 0 pérdida / 0 scroll |

## Riesgos

> Numeración **RG9-1…RG9-12**, propia de este plan (para no colisionar con RG7-1… de F7, RG6-1…
> de F6 ni R9-1… de `research.md`).

| Riesgo | Impacto | Mitigación |
|---|---|---|
| **RG9-1 — El feed de YouTube responde 404/500 de forma intermitente desde IPs de servidor** (verificado el 2026-10-17 desde este entorno: 3×404; WPRSS documenta el comportamiento para datacenter IPs) | El equipo pulsa «Buscar episodios nuevos» y el resumen dice que YouTube falló; si se leyera como «desaparecieron los videos», se ocultaría contenido vivo | Es literalmente FR-016: el feed caído **no oculta ni marca nada** y el resumen lo dice; retry manual cuando responda; documentado en quickstart §8 con el texto exacto que verá el equipo; timeout propio para no arrastrar la búsqueda |
| **RG9-2 — Videos «no listados» en el feed de YouTube** (3 fuentes dicen que no; 1 discrepa y no se pudo refutar con el feed real desde aquí — R9-1) | Un video con circulación restringida aparecería en el sitio | Decisión documentada (el feed es la fuente de verdad) + mitigación cero-coste: revisión del equipo tras cada búsqueda, «ocultar» con un clic y no-reimportación; **señalado a la puerta 2** como incógnita no cerrable |
| **RG9-3 — Falsos positivos/negativos en la unión** (dos sermones distintos el mismo día; frontera de día entre feeds; títulos muy distintos del mismo sermón) | El visitante ve un sermón dos veces (SC-003 en rojo) o un sermón real queda marcado | Regla literal de FR-013/FR-014 (R9-4): lo dudoso **nunca** se une y queda **fuera del público** hasta que el equipo resuelva (US3 esc. 2 lo cubre); el cálculo del día es en `America/Costa_Rica` (el caso real `03 Jul 01:22 GMT` → `02 Jul` lo prueba); tabla de casos con los 9+9 elementos reales |
| **RG9-4 — SSRF/contenido activo vía la primera red saliente del backend** (URLs de carátula y de probador controladas por el feed) | Acceso a red interna o a servicios cloud; contenido ejecutable en el sitio | P9-7: https + allowlist de hosts + rechazo de IPs privadas + sin redirects fuera + `LimitReader` + firma binaria (`storage.Detect`, sin SVG/GIF); enlaces de ficha validados por esquema/host **al importar** (y `CHECK` en la BD); descripciones a texto plano (sin `dangerouslySetInnerHTML`); revisión de `seguridad` antes del PR |
| **RG9-5 — Discrepancias residuales con `ux.md`** (producido en paralelo durante esta fase: composición de fichas, bloque y panel ya definidas) | Ajustes de nombres/composición en las tareas de frontend que no encajen del todo con los DTOs | `ux.md` **manda** en composición y nombres (§6) y este plan fija rutas, DTOs, reglas de contenido y posiciones: cualquier ajuste posterior se aplica en las tareas `[frontend]` **sin cambiar la API** (encaje F4/F7; el bloque reutiliza el DTO de ítem de la sección, y la fecha que `ux.md` omite en el bloque es presentación, no contrato) |
| **RG9-6 — Pruebas de frontend que fijan la lista de permisos y accesos** (F2/F4/F6/F7 cerradas) | `make ci` falla al añadir `MEDIOS`, o se lee como «debilitar» pruebas ajenas | Cambio **acotado y documentado** (mismo gesto que F6/RG6-6 y F7/RG7-6): `features/roles/permissions.test.ts` (`isPermissionAvailable('medios')` pasa a `true`), `features/roles/roles.test.tsx` (el sello «Disponible más adelante» deja de aplicarse a `medios`) y `features/panel/panel.test.tsx` (tarjeta nueva); **sin** cambios de aserciones en backend (el catálogo sigue en 10) |
| **RG9-7 — Hermeticidad de los e2e sin alta manual** (F9 no puede crear fichas por API salvo importando de la red real) | e2e dependientes de internet (flaky, lentos, inaceptables en CI) o sin datos que mostrar | **Siembra por SQL** (helper nuevo `e2e/mediosSeed.ts` sobre el contenedor `db`, patrón declarado) y **el e2e no pulsa «Buscar episodios nuevos»** (quedaría conmutado por la red real y dispararía probadores contra datos sembrados): la importación se cubre con **fakes y fixtures** en Go y con el **recorrido manual** del quickstart §3–§4. Coste declarado: un patrón de siembra nuevo en `frontend/e2e/` |
| **RG9-8 — La búsqueda dura más de lo que aguanta la petición** (primera importación con ~15 carátulas + probadores) | El botón parece colgado; timeouts del proxy o del navegador | Presupuesto de 90 s con topes y paralelismo (R9-9); el botón se desactiva y muestra «Buscando…»; el resumen parcial es válido (nada se pierde: lo hecho está guardado); tiempos medidos en quickstart §3 |
| **RG9-9 — Contradicciones de la spec sobre marcas y visibilidad** (FR-009/SC-002 ↔ default #2 ↔ FR-017) | Implementar la lectura «equivocada» y fallar un SC en aceptación | Síntesis única documentada (R9-7/P9-8) que los cumple a los tres, fijada por pruebas; **pregunta #1 al humano** con el impacto exacto de cada alternativa |
| **RG9-10 — Un enlace muerto de una ficha de dos enlaces** (Q16 literal: oculta el episodio entero) | Un sermón con el video borrado pero el audio vivo desaparece del público | Lectura literal aplicada (pregunta #2 al humano); la marca y «mostrar» permiten al equipo reactivarlo; si el humano prefiere quitar solo el enlace roto, es un cambio acotado en el service (y tocaría FR-003/SC-014) |
| **RG9-11 — Datos hostiles o imprevisibles de los feeds** (HTML en descripciones, títulos gigantes, emojis, fechas inválidas, XML roto, enlaces a hosts ajenos) | XSS, fichas rotas, importaciones a medias | Texto plano forzado (R9-3), límites 300 (título/descripción con recorte documentado), fechas no parseables → item descartado con aviso en el resumen, enlaces fuera de host → item descartado con aviso, XML roto → error claro sin tocar el catálogo; pruebas con casos patológicos en `testdata/` |
| **RG9-12 — Carátula del podcast compartida por todos los episodios de solo audio** (verificado: el `itunes:image` es el arte del programa) | Las fichas de solo audio se ven repetidas | Es un dato del feed (FR-003) y así se publica; se informa al cliente en el cierre; si quiere arte por episodio, lo publica en Spotify (fuera de F9). El marcador neutro cubre el caso sin carátula |

## Complexity Tracking

> Sin violaciones de la constitución. Desviaciones de **convenciones internas**, declaradas:

1. **`visibility ('visible','hidden')` en lugar de `publication_state ('draft','published')`**
   (P9-8): F9 rompe a propósito el patrón de F3/F4/F6/F7 porque la spec lo excluye (Consecuencia
   5: el contenido entra solo y no hay borrador). Los `CHECK`, el orden y el predicado público son
   igual de estrictos; queda anotada en el data-model y en el contrato.
2. **Siembra e2e por SQL** (RG9-7): `e2e/mediosSeed.ts` ejecuta `INSERT` contra el contenedor `db`
   porque F9 no tiene alta manual. Es un patrón nuevo en `frontend/e2e/`; si `revisor-código`
   prefiere otra vía (p. ej. un script SQL versionado en `backend/`), es un refactor acotado.
3. **`coverUrl` sin `alt` editable** (P9-7/FR-003): a diferencia de F6 (`image {url, alt}` con
   `imageAltEs` obligatorio), la carátula de F9 es importada y su metadato no es editable
   (FR-019): el texto alternativo lo deriva el cliente del título (presentación localizada). Si el
   humano quiere `alt` por ficha, habría que hacerlo editable —cambio de alcance sobre FR-003—.
4. **`media.sync` = una fila de auditoría por búsqueda** (P9-14): FR-022 pide registrar la
   operación «Buscar episodios nuevos»; se registra la operación (no cada elemento importado, que
   no es una acción del equipo). El detalle numérico vive en `media_sync_runs`. Si el humano
   prefiere una fila por elemento, es más ruido en `admin_actions` con el mismo modelo.
5. **`/api/v1/portada/medios` lo registra el dominio `medios`** (P9-2): el segmento `portada`
   pertenece al agregado de F3 pero el recurso es de este módulo (mismo gesto que F4 con
   `proximamente`, F6 con `ministerios` y F7 con `ofrendas`).
6. **La búsqueda es una operación larga dentro del ciclo petición/respuesta** (P9-3/RG9-8): la
   constitución prohíbe procesos en segundo plano y la spec exige que quepa en la petición; el
   presupuesto de 90 s es el coste declarado (con topes, paralelismo y resumen parcial). Si el
   volumen real hiciera falta más, el camino es un incremento futuro (la «sincronización de
   episodios» del roadmap §5), no de F9.
7. **Nombres de pieza de UI: alias de este plan** (RG9-5): `ux.md` se produjo en paralelo durante
   esta fase y **manda** en composición y nombres (`MedioCard`, `MedioItem`, `MedioLinks`,
   `CaratulaNeutra`, `CargarMasButton`, `BuscarEpisodios`, `ResumenImportacion`,
   `EpisodiosList`…). Los nombres de este plan (`EpisodioCard`, `ExternalLink`, `MediosItem`,
   `BuscarNuevosButton`…) son alias que `tasks.md` resolverá a los de `ux.md`; cualquier ajuste de
   composición posterior se aplica en las tareas `[frontend]` **sin** tocar la API.

## Propuesta de tareas (insumo para `/speckit.tasks`; **no** es `tasks.md`)

Orden típico migración → contrato → repositorio → servicio → handler → frontend → e2e. Las capas
se reparten `[db]`/`[backend]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` →
`devops`; las marcadas `[P]` son independientes entre sí y pueden ejecutarse en paralelo.

1. `[db]` Migración `000014` (`media_items` + `media_episode_sources` + `media_sync_runs`, con
   sus `CHECK`) con su `down` y prueba up/down/up. `[P]`
2. `[db]` Migración `000015` (`admin_actions_action_check` + 6 códigos `media.*`) con su `down` y
   prueba up/down/up. `[P]`
3. `[backend]` Fusión del delta en `backend/api/openapi.yaml` (`info.version` → **0.9.0**; enum
   `AdminActionItem.action` +6) + `make api-gen`. `[P]`
4. `[backend]` `platform/audit`: 6 códigos + pruebas del registro cerrado. `[P]`
5. `[backend]` `titulo.go` (normalización + día en `APP_TZ`) + **captura de fixtures** de los dos
   feeds en `testdata/` (quickstart §0) + `titulo_test.go`/`feeds_test.go`. `[P]`
6. `[backend]` Consultas `media_episodes.sql`/`media_episode_sources.sql`/`media_sync_runs.sql` +
   rama de `media.sql` (`make sqlc-gen`) y repository de `medios` con pruebas de integración.
   *Depende de 1, 2, 4 y 5.*
7. `[backend]` `disponibilidad.go` y `caratula.go` (probadores y descarga acotada) + sus pruebas
   (fake `RoundTripper`). `[P]` respecto a 6.
8. `[backend]` Service público (`GetMedios` paginado + `GetPortadaMedios`, idioma con fallback,
   predicado público) + pruebas. *Depende de 6.*
9. `[backend]` Service de panel (listado + `lastSync`, descripción/visibilidad, unión por puntero,
   resolver marcas) + pruebas. *Depende de 6.*
10. `[backend]` `service_sync.go` (`BuscarNuevos`: fakes de feeds/probadores/store, idempotencia,
    unión automática y marcado, FR-016/FR-017, resumen y `media_sync_runs`) + pruebas. *Depende de
    6 y 7.*
11. `[backend]` Handlers y `routes.go` (`RegisterPublic` 2 rutas; `RegisterAdmin` con
    `AdminChain("medios")`) + `service_audit.go` + pruebas + humo en `main_test.go`. *Depende de 8,
    9 y 10.*
12. `[frontend]` Cliente `api/medios.ts` + `schema.d.ts` regenerado + hooks. *Depende de 3.* `[P]`
    respecto a 5–11 (avanza contra el contrato).
13. `[frontend]` Sección pública `/medios` (`MediosPage` con «Cargar más», `EpisodioCard`,
    `ExternalLink` anunciado, marcador neutro, **redirección** —P9-16—, i18n). *Depende de 12.*
14. `[frontend]` `useSiteLinks` con la entrada «Medios» condicional + `MediosBlock` al final de
    `HomePage` (tras `OfrendasBlock`) + i18n. *Depende de 12.*
15. `[frontend]` Panel `features/medios/` (`MediosPage`, `BuscarNuevosButton` + `ResumenBusqueda`,
    filtros, `DescripcionForm`, `UnirDialog`, `ResolverMarcaDialog`), ruta `/panel/medios` con
    `RequirePermission(MEDIOS)`, `lib/permissions.ts`, `features/roles/permissions.ts`,
    `AccesosRapidos` y `ACTION_LABELS` (+6) + actualización acotada de las 2–3 pruebas de
    permisos/accesos (RG9-6). *Depende de 12.*
16. `[…]` e2e `medios-publica.spec.ts`, `medios-panel.spec.ts`, `medios-accesibilidad.spec.ts`,
    `portada-medios.spec.ts` + `mediosSeed.ts` (siembra por SQL, RG9-7) + checklist de quickstart.
    *Depende de 13–15.*
17. `[…]` Validación (qa/revisor/seguridad), documentación (documentador) y PR. *Depende de 16.*

## Preguntas para el humano (defaults aplicados; nada se implementa como regla cerrada sin esto)

Los cuatro primeros puntos son huecos o contradicciones de la **spec**; se aplican los defaults
indicados —los más literales posibles— y quedan fijados por prueba, de modo que cambiarlos es un
cambio acotado. **No son reglas de negocio inventadas**: en cada caso se cita el texto que manda.

1. **¿Qué fichas con marcas abiertas ve el visitante?** *(contradicción: FR-009/SC-002 dicen que
   ninguna entrada con marca sin resolver se publica; el default revisable #2 dice que de un par
   «solo la primera entrada importada queda visible»; FR-017 dice que «ya no aparece en el feed»
   **sigue visible**)* — **Default aplicado (R9-7)**: el **único** bloqueo por marca es
   `possible_duplicate` **sin resolver** (el par entero queda fuera del público; la «primera»
   conserva su flag `visible` para que, al resolver «no son duplicados», ambas se publiquen —US3
   esc. 2—); `missing_from_feed` **sí** se publica (FR-017 literal) y `unavailable` ya está oculta.
   Alternativa A (toda marca bloquea): incumple FR-017. Alternativa B (el par marcado se publica
   parcialmente): incumple SC-002/FR-009.
2. **¿Un enlace confirmado muerto oculta la ficha entera aunque el otro enlace siga vivo?**
   *(Q16/FR-017 dicen literalmente «el episodio desaparece de la página pública»)* — **Default
   aplicado: sí** (ficha entera oculta + marca «no disponible»; el equipo puede reactivarla tras
   resolver). Si prefieres quitar **solo** el enlace roto y publicar la ficha con el otro: cambia
   el service (1 regla + pruebas) y hay que ajustar FR-003/FR-017/SC-014, que hoy asumen la
   ocultación de la ficha.
3. **¿Qué descripción corta se publica si el feed no trae descripción?** *(FR-020: «cada episodio
   DEBE tener descripción corta»; pero la fuente puede venir vacía)* — **Default aplicado**: el
   campo queda **ausente** y la ficha **omite la línea** (SC-010: sin líneas vacías ni marcadores);
   el equipo puede escribirla después (es editable). Con los feeds reales no ocurre (todos los
   items traen descripción: verificado). Alternativa: exigir descripción (habría que inventar
   texto) o guardar `""` (rompe los `CHECK`).
4. **Frontera del «mismo día» entre feeds** *(FR-013: «fueron publicados el mismo día»; los feeds
   publican en zonas distintas y el caso real `03 Jul 01:22 GMT` cae en día distinto según el
   cálculo)* — **Default aplicado**: mismo día **en `America/Costa_Rica`** y **sin margen**: si los
   días difieren, el par entra como dudoso (marcado, fuera del público, reversible con un clic).
   Si prefieres margen ±1 día para la unión automática, es una constante y sus casos de prueba.

Y dos confirmaciones menores (defaults ya señalados por la spec o por precedentes, revisables):

5. **Nombre que manda al unir** (⚠️ señalado por la spec, default revisable #7): la ficha unida
   publica **el nombre del feed del video** (el mismo que trae la carátula y la fecha); el título
   del podcast se conserva como dato de origen en `sources`. Si prefieres el del podcast o el
   primero importado, cambia solo esa regla (P9-10).
6. **Desempate del orden con la misma fecha** (default revisable #11): literal «orden de
   importación» = **FIFO** (`created_at ASC`, luego `id`) dentro del orden general
   `published_at DESC`. Invertir la dirección del desempate es una línea + sus pruebas.
