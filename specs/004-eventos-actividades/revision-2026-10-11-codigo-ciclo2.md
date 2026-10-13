# Revisión de código — F4 Eventos y actividades · ciclo 2 (re-validación)

**Fecha**: 2026-10-11 · **Fase**: 7 (validación, bucle de corrección 2) · **Revisor**: `revisor-codigo` (solo lectura; este informe es la única escritura)
**Alcance**: commits de corrección `1714e2b`, `3ec0b2d`, `53c7007`, `8b03db3` (diff `d83e74b..8b03db3`: 13 archivos, +598/−96, solo backend) contrastados con los hallazgos I1 y M1–M3 de `revision-2026-10-11-codigo.md`, `spec.md`, `plan.md`, `tasks.md`, `data-model.md`, `contracts/openapi.yaml` y la constitución.

## Veredicto

**APROBADO** — I1, M1, M2 y M3 quedan **cerrados sin regresiones**. Las cuatro correcciones son mínimas, van al hallazgo que les corresponde, conservan el comportamiento observable (orden, filtro «próximo», claves omitidas, fallback de idioma, respuestas de error y de auditoría) y añaden pruebas que fijan el nuevo comportamiento. Quedan solo 3 SUGERENCIAS (deriva menor de documentación y una cobertura de prueba ausente para una mejora coherente con R4-10), ninguna bloqueante.

## Cierre de hallazgos del ciclo 1

### I1 — Consultas N+1 en las lecturas de listado → **CERRADO**

- **Consultas por lote**: `queries/agenda.sql` introduce `ListEventDatesForEvents` y `ListUpcomingEventDatesForEvents`, ambas con `event_id = ANY(sqlc.arg('event_ids')::uuid[])` y `ORDER BY event_id, event_date, start_time NULLS FIRST, id` (el mismo orden por evento que la variante anterior). El corte R4-2 (`today`/`now_time`) sigue parametrizado (§IV/CWE-89) y es **la misma implementación** del filtro (`isUpcoming`, `service.go`; literal SQL idéntico al previo).
- **Sin N+1**: `GetProximamente` (`service_public.go:130`, `upcomingDatesByEvent` en `:167`) hace **1 consulta por bloque** y **0 si no hay eventos**; `ListEventos` (`service_admin.go:327`, `datesByEvent` en `:342`) hace **1 por página** y **0 en página vacía**; `ListActividades` (`service_admin.go:878`) solo pasaba el corte a `actividadAdminFrom` por fila —su fecha vive en su propia fila, nunca hubo consulta por fila— y ahora calcula el corte **una vez**.
- **Sin SQL muerto**: la variante `ListUpcomingEventDates` por evento se retiró de `agenda.sql`, del puerto (`service.go:153-155`) y del repository; verificado que **todas** las consultas nombradas de `agenda.sql` tienen consumidor (barrido automático, 0 huérfanas). `ListEventDates` (por evento) sigue viva y en uso en los caminos de un solo elemento (`eventoAdmin`, `service_admin.go:362`; y `service_admin.go:232` para las fechas efectivas del duplicado).
- **Comportamiento conservado** (verificado leyendo el diff y con pruebas):
  - *Orden*: `groupDatesByEvent` (`service_public.go:181`) agrupa preservando el orden ascendente de la consulta; `sortEvents`/`sortActivities` siguen aplicándose después.
  - *Filtro «próximo» y defensa*: `if len(dates) == 0 { continue }` se conserva (`service_public.go:135-138`); las fechas pasadas siguen sin salir al público.
  - *Claves omitidas*: `out.Events`/`out.Activities` solo se asignan si hay elementos (`service_public.go:153-158`); `eventoAdminFrom` sigue construyendo `Dates: make([]FechaAdmin, 0, …)` → `[]` en JSON (sin regresión a `null` con eventos sin fechas).
  - *Fallback de idioma*: `localizeEvent`/`localizeActivity` y la validación de `lang` no se tocaron.
- **Sin código muerto nuevo**: `eventoAdmin`/`actividadAdmin` se mantienen solo para los caminos de un solo elemento (Create/Update/Get), donde siguen en uso; `mapEventDate` resuelve `EventID` (`repository_eventos.go:363`), que es la clave de agrupación.
- **Pruebas que lo fijan**: unitarias `TestGetProximamenteReadsDatesInOneBatchQuery` (25 eventos → 1 llamada al lote; sin eventos → 0; solo fechas próximas y en orden) y `TestListEventosReadsDatesInOneBatchQuery` (12 eventos → 1; página vacía → 0; `dates` ascendentes, `nextDate`/`isPast` intactos); integración `TestIntegrationListUpcomingEventDatesForEvents` y `TestIntegrationListEventDatesForEvents` contra PostgreSQL real (solo próximas —incluye «hoy sin fin» y «hoy con fin ya terminado»—, vencidos sin fechas, orden ascendente, sin mezclar eventos ajenos, lote vacío). El fake cuenta las llamadas al lote y resuelve `event_id` como el repository real.

### M1 — Código muerto de `IsHomeFilePublished` en el fake de portada → **CERRADO**
`backend/internal/portada/service_test.go` (2+/14−): se retiran únicamente el método muerto del fake, el campo `publishedFiles` y su inicialización. **Ninguna aserción se elimina**; la política de publicación la cubre `internal/media`. Sin cambio de comportamiento.

### M2 — Tope `MaxCategories` con conteo → **CERRADO**
`CountCategories :one` (`agenda.sql:66-71`, `count(*)::bigint`), repository (`repository_categorias.go:129`), puerto (`service.go:145`) y comparación `total >= MaxCategories` (`service_admin.go:1169`). El caso «límite del catálogo» se **refuerza**: ahora exige 1 conteo y **0 listados** del catálogo. Mismo `400 invalid` con `details.categorias`. La naturaleza de tope blando ante altas concurrentes queda documentada en la consulta, tal como aceptó el ciclo 1.

### M3 — Etiquetas de auditoría legibles → **CERRADO**
La etiqueta de cada `fail` arranca en el UUID (solo mientras el elemento no está leído: id inexistente o petición sin cambios) y toma `current.TitleEs`/`current.NameEs` en cuanto se conoce (`service_admin.go:218`, `:811`, `:1217`); las filas de éxito siguen con el título/nombre guardado (`:250`, `:821`, `:1227`). Las ramas de conflicto de la base usan la identidad **actual** (`:264`, `:834`, `:1233`), con la razón documentada (nombrar por el nombre nuevo apuntaría a otro elemento). `CreateEvento` registra el duplicado con el título ya normalizado (`:173`), como `CreateActividad`. `TestRejectionLabelsUseTheTitle` cubre 5 casos con `result='failure'` (datos inválidos → título actual; id inexistente → UUID; actividad; categoría duplicada → nombre actual; alta duplicada → título).

## Pruebas congeladas y coherencia

- **Ninguna prueba congelada tocada y ningún debilitamiento**: el diff de pruebas se limita a archivos de F4 (`fake_repository_test.go`, `service_admin_test.go` 165+/0−, `service_public_test.go` 46+/0− — solo adiciones —, `repository_eventos_integration_test.go`), al fake de portada (M1, solo código muerto) y a los archivos de integración de F4. El renombrado de `TestIntegrationListUpcomingEventDatesOnlyUpcoming` → `TestIntegrationListUpcomingEventDatesForEvents` es la adaptación obligada al retirar la consulta por evento: sus aserciones (solo próximas, «hoy sin fin», panel con todas las fechas) se conservan **y se amplían** (múltiples eventos, vencidos, lote vacío, no mezclar ajenos). Sin `skip`/`only`/`xit` en ningún archivo tocado.
- **Contrato**: `openapi.yaml` (v0.5.0) no cambia — las correcciones son internas (consultas, auditoría, tests); las respuestas HTTP son idénticas.
- **`tasks.md`**: las tareas siguen completas y coherentes en alcance; solo deriva nominal menor (S1).

## Hallazgos nuevos (ciclo 2)

### SUGERENCIA

**S1 — Deriva de documentación por el renombrado de consultas.** `data-model.md:329` todavía nombra `ListUpcomingEventDates` (ya retirada) y no lista las variantes por lote (`ListEventDatesForEvents`, `ListUpcomingEventDatesForEvents`) ni `CountCategories` (M2); `tasks.md:247` y `tasks.md:426` también citan el nombre antiguo. **Corrección**: actualizar la fila de `data-model.md` a `ListEventDates` / `ListEventDatesForEvents` / `ListUpcomingEventDatesForEvents`, añadir `CountCategories` a la tabla y, opcionalmente, una nota en T411/T412 (el registro histórico puede quedarse como está).

**S2 — Rama de conflicto de `UpdateEvento` sin prueba que la fije.** `service_admin.go:263-266` (nueva en `8b03db3`) registra un rechazo `failure` cuando la base detecta conflicto en `UpdateEvent`, alineando `UpdateEvento` con `UpdateActividad`/`UpdateCategoria` y con el espíritu de R4-10 («los intentos fallidos también quedan en la auditoría», `service.go:486-489`). Es una mejora coherente, pero `TestRejectionLabelsUseTheTitle` no cubre esa rama (los conflictos de edición probados son de `ExistsEventDuplicate`/`UpdateCategory`). **Corrección**: un caso unitario con el fake devolviendo conflicto desde `UpdateEvent`, que fije la etiqueta `Eventos · <título actual>` y `result='failure'`.

**S3 — Comentario de sección pegado.** `service_public.go:203` (tras `eventIDs`) deja el comentario `// --- Categorías del bloque ---` sin línea en blanco después; pura legibilidad. *(Solo si el formateo no lo resuelve; no toca comportamiento.)*

## Evidencia verificada en esta re-validación

| Comprobación | Resultado |
|---|---|
| `gofmt -l`, `go vet ./...`, `golangci-lint run ./...` (backend) | limpios (0 issues) |
| `go test ./internal/...` (unitarias) | en verde |
| `go test -tags=integration ./internal/agenda ./internal/media` con `DATABASE_URL_TEST` (base `app_test`, migraciones 000001–000009) | en verde (agenda 9,7 s; media 0,1 s) — ejecutadas de verdad, no omitidas (las 2 pruebas por lote en `--- PASS`) |
| Cobertura de `internal/agenda/service*.go` (83 funcs) | **91,5 %** (≥ 80 % exigido; sin cambios respecto al ciclo 1) |
| Consultas huérfanas en `db/queries/agenda.sql` | ninguna |
| Frontend (sin cambios en los 4 commits): `tsc --noEmit`, `npm run lint`, `vitest` | limpios; **505/505 en verde** |
| Pruebas desactivadas (`skip`/`only`/`xit`) | ninguna |
| Pruebas congeladas de F1–F3/F10 | intactas (solo el código muerto aprobado en M1) |

## Condiciones para el merge

Ninguna. I1/M1–M3 quedan cerrados; S1–S3 pueden resolverse en este PR (S1 es un cambio de dos líneas de documentación) o quedar como hallazgos abiertos menores registrados en `estado.md`.
