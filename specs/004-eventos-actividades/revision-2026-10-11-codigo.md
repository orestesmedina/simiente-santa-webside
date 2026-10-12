# Revisión de código — F4 Eventos y actividades

**Fecha**: 2026-10-11 · **Fase**: 7 (validación) · **Revisor**: `revisor-codigo` (solo lectura; este informe es la única escritura)
**Alcance**: `git diff main...HEAD` (rama `004-eventos-actividades`, 163 archivos, +31 487/−1 698) contrastado con `spec.md`, `plan.md`, `tasks.md`, `ux.md`, `data-model.md`, `contracts/openapi.yaml` y `.specify/memory/constitution.md`.

## Veredicto

**APROBADO CON CAMBIOS MENORES** — sin hallazgos CRÍTICOS. La implementación corresponde a lo aprobado (spec v2 con permiso «categorías», plan con las 5 decisiones firmes de la puerta 2), respeta la arquitectura por capas, las migraciones son limpias y reversibles, las pruebas son reales y de test-first, y la excepción RG4-10 se aplicó exactamente como se autorizó. Quedan **1 IMPORTANTE** (rendimiento: consultas N+1 en lecturas de listado) y **5 MENORES**.

## Evidencia verificada en esta revisión

| Comprobación | Resultado |
|---|---|
| `gofmt -l`, `go vet ./...`, `golangci-lint run ./...` (backend) | limpios (0 issues) |
| `go build ./...`, `go test ./internal/...` | en verde |
| `go test -tags=integration ./internal/agenda ./internal/media` | en verde |
| Cobertura de `internal/agenda/service*.go` (79 funcs) | **91,5 %** (≥ 80 % exigido; bitácora T440: 89,4 %) |
| `npx tsc --noEmit`, `npm run lint` (frontend) | limpios |
| `npx vitest run` (frontend) | **505/505 en verde** (74 archivos) |
| Pruebas desactivadas (`skip`/`only`/`xit`) | ninguna |
| `any` injustificado en `frontend/src` (fuera de tests) | ninguno |

## Hallazgos

### IMPORTANTE

**I1 — Consultas N+1 en las lecturas de listado (rendimiento).**
- `backend/internal/agenda/service_public.go:128-139` — `GetProximamente` ejecuta `ListUpcomingEventDates` **una vez por evento** (hasta `MaxBlockItems = 200`, `service.go:63`) sobre un endpoint público con `Cache-Control: no-store` (`handler_public.go:52`): hasta 201 consultas por petición.
- `backend/internal/agenda/service_admin.go:306-313` y `:846-853` — `ListEventos`/`ListActividades` llaman `eventoAdmin`/`actividadAdmin` por fila, y `eventoAdmin` (`service_admin.go:326-336`) consulta `ListEventDates` por evento (hasta 100 filas por página, `AGENDA_PAGE_LIMIT`).
- **Corrección propuesta**: añadir al puerto y a `queries/agenda.sql` una consulta por lote (`ListUpcomingEventDatesForEvents(event_ids uuid[], today, now_time)` con `event_id = ANY(...)` y `GROUP BY`), agrupar por `event_id` en el service y reutilizarla también en los listados del panel (con `ListEventDatesForEvents`). Hoy está acotado y no rompe nada; conviene dejarlo resuelto antes de F5–F9, que heredan el patrón.

### MENOR

**M1 — Código muerto en las pruebas de F3 tras la promoción de `media`.**
`backend/internal/portada/service_test.go:194-200` — el fake sigue implementando `IsHomeFilePublished` (y el campo `publishedFiles`, líneas 33 y 43) pero el puerto `Repository` de portada (`service.go:95`) ya no lo declara desde T415. **Corrección**: borrar el método y el campo del fake.

**M2 — Tope `MaxCategories` por conteo ineficaz y con carrera.**
`backend/internal/agenda/service_admin.go:1117-1126` — `CreateCategoria` carga **todo** el catálogo (`ListCategories`) para contar, y dos altas concurrentes pueden superar las 100 entradas; el envoltorio dice «contar las categorías» y la consulta es de listado. **Corrección**: consulta `CountCategories :one` y comparar con `MaxCategories` (aceptando que el tope es blando si no se serializa).

**M3 — Etiquetas de auditoría poco legibles en los rechazos de edición.**
`backend/internal/agenda/service_admin.go:195`, `:757` y `:1151` — los `fail(...)` de `UpdateEvento`/`UpdateActividad`/`UpdateCategoria` registran `targetLabel` con el **UUID** (`id.String()`) aunque `current` ya está leído y su título/nombre está disponible; el resto de filas usan `Eventos · <título>`. Además `CreateEvento` (`service_admin.go:131`) registra el duplicado sin el título mientras `CreateActividad` (`:745`) sí lo incluye. **Corrección**: pasar `current.TitleEs`/`current.NameEs` a la etiqueta del `fail` de edición y al `fail` de alta cuando el título ya se conoce (coherencia con R4-10 «etiquetas legibles»).

**M4 — Cambio de un test congelado de F3/F10 no registrado como excepción.**
`frontend/src/features/publico/hooks/useSiteLinks.test.ts` → `.tsx` — el test «no inventa enlaces a rutas F4–F9» se sustituye por «las anclas de la portada no cambian…» (sus aserciones de anclas se conservan) porque el producto cambió de forma aprobada (ruta `/eventos`, ⟲ D-1). Es legítimo —no es debilitamiento— pero `estado.md` solo registró el ajuste de `acceso.spec.ts`. **Corrección**: añadirlo a «Decisiones y aclaraciones»/excepciones de `estado.md` y al registro del PR junto a RG4-10.

**M5 — `Optional[T]` duplicado entre dominios (desviación declarada).**
`backend/internal/agenda/optional.go:23` repite el tipo de `internal/portada` (nota en `optional.go:15-17`). **Aceptable hoy** (regla de tres: dos usos); se deja como recordatorio de promoverlo a `platform/` cuando lo pida un tercer dominio.

### Sin hallazgos (verificado y correcto)

- **Capas** (§II): handlers sin SQL ni reglas de negocio (`handler*.go`); service sin `net/http`/`pgx`; repository es el único que conoce pgx y traduce errores (`repository.go:219-239`). `media` incorpora su `service.go`, lo que **refuerza** §II.
- **Seguridad**: SQL 100 % parametrizado (`queries/agenda.sql`); autorización por módulo en servidor (`routes.go:59-88`, `AdminChain("eventos"|"actividades"|"categorias")`); validación de toda entrada en backend (`model.go` + `service.go`/`service_admin.go`); subidas por firma binaria JPEG/PNG/WebP sin SVG (`uploadImage`, `service_admin.go:553`), nombre generado por el servidor, fail-closed si la auditoría falla; sin secretos ni `dangerouslySetInnerHTML`.
- **Migraciones** (§VI): `000007`/`000008`/`000009` con `up`/`down`, sin tocar `000002` aplicada (el permiso `categorias` es un `INSERT`), FK explícitas (`ON DELETE RESTRICT/CASCADE`) e índices explícitos en todas las tablas nuevas. `down` destructivos documentados como en F2/F3.
- **Regla de «próximo» (R4-2)**: una sola implementación (`isUpcoming`, `service.go:284`), coherente con el literal SQL y con `media` (`IsFilePublished`); corte en `APP_TZ` validado fail-closed (`service.go:266`), `config.go` con `time/tzdata` y default firme `America/Costa_Rica`.
- **Pruebas**: test-first (cada tarea con su prueba), unitarias + integración contra PostgreSQL real (contenedor), e2e de las tres superficies, y las pruebas congeladas de F2/F3/F10 intactas salvo lo autorizado (ver dictámenes abajo).
- **Trazabilidad**: `backend/api/openapi.yaml` v0.5.0 contiene las 17 operaciones de F4 (rutas verificadas contra `routes.go`), `schema.d.ts` regenerado sin deriva, `data-model.md` realineado con `ReplaceEventDates` = `DELETE`+`INSERT`, `tasks.md` T401–T441 completas (41/42).

## Dictamen sobre las desviaciones reportadas

| Desviación | Dictamen |
|---|---|
| `media` con `service.go` (el plan describía handler+repository) | **Aceptable**: añade la capa que exige §II; sin cambio de comportamiento (mismas rutas, cabeceras y errores que F3). |
| `home.sql` sin retirar en T404 | **Aceptable, ya cerrada**: se retiró en T415 (commit `3909a56`) y `IsFilePublished` de `media.sql` cubre el comportamiento con pruebas de integración equivalentes (`repository_media_integration_test.go`). Solo queda el residuo M1. |
| `ReplaceEventDates` compuesto (`DeleteEventDates`+`InsertEventDates`) | **Aceptable**: razón técnica válida (el `UNIQUE (event_id, event_date)` se comprueba contra la instantánea previa; ver `agenda.sql:173-183`), documentada en código y realineada en `data-model.md:328`. |
| `T430` ejecutada al final | **Aceptable**: el orden de ejecución no altera el resultado; dependencias respetadas, tests añadidos (`panel.test.tsx`, `permissions.test.ts`) y `tasks.md` al día. |
| `Optional[T]` repetido | **Aceptable con condición** (M5): promover a `platform/` en el tercer uso. |

## Excepción aprobada RG4-10 y ajustes a pruebas congeladas

- **RG4-10 (backend F2, 9→10 permisos)**: el diff se limita a las **5 aserciones de recuento** (9→10), la fixture `sampleCatalog()` (+`categorias`) y comentarios. **Ninguna aserción de comportamiento se relajó**; el init de F2 sigue verificando el catálogo dinámico. **Correcto.**
- **Frontend F2 (`permissions.test.ts`, `roles.test.tsx`)**: sustitución de fixtures (`eventos`→`noticias`/`donaciones`) porque F4 convirtió esos permisos en módulos reales; escenarios y aserciones conservados y se añaden pruebas nuevas. Misma naturaleza que RG4-10: **aceptable** (recomendación M4: registrarlo en el PR).
- **`frontend/e2e/acceso.spec.ts`**: el fixture pasa a dos módulos de F5–F9 («Noticias y galería», «Donaciones») y el escenario «cuenta con módulos reservados → solo Inicio y el aviso» queda **intacto**. **Correcto** (ya registrado en `estado.md`).

## Condiciones para el merge (cambios menores)

1. I1 (N+1) puede resolverse en este PR o abrirse como deuda explícita antes de F5–F9; conviene decidirlo antes del PR.
2. M1–M4 se corrigen/registran en el mismo PR o quedan como hallazgos abiertos en `estado.md`.
