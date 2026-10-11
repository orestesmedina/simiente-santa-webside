# Implementation Plan: Eventos y actividades (F4)

**Branch**: `004-eventos-actividades` | **Date**: 2026-10-11 | **Spec**: [spec.md](./spec.md)
(aprobada por el humano; Q1–Q6 resueltas el 2026-10-11 y **puerta 2 con las 7 decisiones firmes
aplicadas el 2026-10-11**: `APP_TZ = America/Costa_Rica`, ruta `/eventos`, permiso nuevo
`categorias`, aviso sin redirección, `endTime` exige `startTime`, borrado de categoría en uso →
`409` y imagen de vencido no servida; el cambio de alcance lo recoge `analista-producto` en
`spec.md` en paralelo)

**Input**: `specs/004-eventos-actividades/spec.md` (fuente de verdad) · `.specify/memory/constitution.md`
· `docs/tecnico/arquitectura.md` (reglas R1–R8 y convenciones §8.1) y `docs/tecnico/decisiones.md` ·
**F2 terminada** (`specs/002-acceso-gestion-usuarios/`: patrón de permisos por módulo, sesión,
CSRF, `authz`, `apperr`, `validate`, `paginate`, auditoría y catálogo de permisos **con `eventos` y
`actividades` ya sembrados** — verificado en `000002_create_roles_and_permissions` —; su catálogo
**crece a 10** con `categorias` en F4, ver P4-5) · **F3 terminada**
(`specs/003-portada-info-general/`: patrón bilingüe, publicación por elemento, imágenes en disco
local, i18n del público, auditoría de contenido — se usan como plantilla **y** como código
existente) · **F10 terminada** (`specs/010-sistema-diseno-movimiento/`: sistema de diseño, modo
oscuro, movimiento, navegación y selector de idioma — se **aplica**, no se redefine) · código real
de `backend/` y `frontend/` · skills `go-backend`, `postgres-db`, `react-frontend`.

## Summary

F4 es la **primera sección de contenido dinámico** del sitio público después de la portada: «lo que
viene» —**eventos** (encuentros periódicos: noche de aposentos, vigilia) y **actividades** (acciones
puntuales: repartir comida, predicar en tal lugar)— con título, fecha(s), horario, lugar, detalle,
imagen, organizador y categoría, **bilingüe** (es obligatorio, en opcional con fallback) y con
gestión **por separado** en el panel bajo los permisos **`eventos`** y **`actividades`** que F2 ya
reservó (FR-012: se **activan**, no se crean permisos para los módulos) y el catálogo de categorías
bajo el **permiso nuevo `categorias`** — 10.º del catálogo de F2, **cambio de alcance autorizado
por el humano en la puerta 2** (el único permiso que F4 crea; se siembra en la migración
`000009`)—, estados borrador/publicado y auditoría de solo lectura.

Técnicamente F4 añade: los **primeros modelos de contenido con fechas** (`agenda_events` +
`agenda_event_dates` para la lista explícita de ocurrencias del evento —Q1, sin cálculo de
recurrencia— y `agenda_activities` con fecha única —Decisión 3—), el **primer catálogo
administrable** (`agenda_categories`, Q6, compartido por ambos módulos), el **retiro automático de
la vista pública** resuelto como **filtro de lectura** "publicado y próximo" en el servidor sobre la
zona horaria **`APP_TZ = America/Costa_Rica`** (FR-006/Q4: sin cron y sin cambiar estados; valor
firme de la puerta 2), la **primera superficie de panel con permisos independientes**
(`AdminChain("eventos")`, `AdminChain("actividades")` y `AdminChain("categorias")` — con el
plumbing existente, sin helpers nuevos de middleware —), 15 códigos de auditoría nuevos (migración
`000008`) y la **promoción del servicio de archivos** a un dominio propio `internal/media/` para que
las imágenes de F4–F9 se sirvan con la misma regla de F3 ("solo contenido publicado", extendida a
"y próximo"). El frontend publica **`/eventos`** con los dos bloques dentro del sistema de F10,
oculta la entrada de navegación cuando no hay nada publicado (y la ruta directa muestra un **aviso
de sección vacía sin redirigir**, R4-13) y añade los módulos de panel `/panel/eventos`,
`/panel/actividades` y `/panel/categorias`.

**Ninguna dependencia nueva** ni en backend ni en frontend.

## Technical Context

**Language/Version**: Go 1.27 (sin cambios) · TypeScript 5.x `strict` sobre Node 22 · React 19.

**Primary Dependencies**: **ninguna nueva**. Se reutiliza todo lo existente: `net/http` (multipart
de la subida de imágenes), `net/url`, `time`, `pgx`/`sqlc`, `testcontainers-go`; en frontend React
Router, TanStack Query, React Hook Form + Zod, Tailwind, Vitest + Testing Library + MSW, Playwright
e `Intl.DateTimeFormat` (stdlib del navegador) para las fechas localizadas (research R4-15).

**Storage**: **PostgreSQL 16** (tablas `agenda_*` + ampliación del `CHECK` de `admin_actions` +
sembrado del permiso `categorias` en `permissions`; migraciones `000007`/`000008`/`000009`, ver
[data-model.md](./data-model.md)) · Redis 7 **sin tocar** (sesión
de F2) · **disco local del backend** reutilizado para imágenes (`platform/storage`, volumen
`uploads_data` de F3). Capa de datos sqlc (D-A3): consultas en
`backend/internal/db/queries/agenda.sql` y `media.sql`, código generado commiteado; la auditoría
reutiliza `InsertAdminAction` de `queries/audit.sql` **dentro de la transacción** de cada mutación.

**Testing**: estrategia por capa de F1/F2/F3 ampliada (sección propia): `go test` (unitarias),
`go test -tags=integration` (repository contra PostgreSQL real —incluida la migración up/down/up de
`000008`/`000009` y la regla de "próximo" con fechas fijas—), Vitest + Testing Library + MSW y
Playwright (e2e pública y de panel). **Excepción aprobada**: 3 pruebas de F2 se actualizan
deliberadamente de 9 a 10 permisos (más 2 aserciones y la fixture detectadas en revisión; ver
RG4-10 y la "Propuesta de tareas").

**Target Platform**: Docker Compose local + GitHub Actions (CI del kit, sin tocar). Nueva variable de
entorno **`APP_TZ`** con el valor firme **`America/Costa_Rica`** (zona horaria del lugar para el
corte de "hoy"; P4-3/R4-2). Sin despliegue a producción (fuera de alcance de F4).

**Performance Goals**: `GET /api/v1/proximamente` < 500 ms con BD sana (dos consultas + fechas;
JSON pequeño, `no-store`) · listados de panel paginados (`limit`/`offset`, §8.1.3) · subida de
imagen ≤ 8 MB validada en streaming (reglas de F3 sin cambios) · descarga de imágenes con la misma
política `no-store` de F3.

**Constraints**: stack fijo (constitución) · capas `handler → service → repository` (§II) · contrato
OpenAPI antes que el código (§II) · migraciones versionadas e inmutables (§VI) · validación de toda
entrada en backend (§IV, CWE-20) · SQL siempre parametrizado (§IV, CWE-89) · nada de contenido
ejecutable (§IV, CWE-79: sin `dangerouslySetInnerHTML`, sin SVG en imágenes) · authn/authz/CSRF en
servidor (§IV, CWE-862) · dependencias minimizadas (0 nuevas) · archivos del kit no editables
(regla 10) · **la spec no se reabre**: los huecos (permisos del catálogo, `endTime` sin `startTime`,
borrado de categoría en uso, imagen de elemento vencido, página pública vacía) **quedaron cerrados
como decisiones firmes del humano en la puerta 2** (2026-10-11), documentadas en
[research.md](./research.md); el único cambio de alcance —permiso nuevo `categorias`— está
autorizado y se documenta con su impacto en F2 (RG4-10).

**Scale/Scope**: 17 operaciones REST (1 pública · 16 de panel) sobre 4 tablas nuevas + la ampliación
del `CHECK` de `admin_actions` + la fila nueva del permiso `categorias` en `permissions` (F2); 1
dominio backend nuevo (`agenda`) + 1 dominio promovido (`media`), **`platform/middleware` sin
cambios** (el permiso propio de categorías reutiliza `AdminChain`); 3 migraciones (`000007`,
`000008`, `000009`); frontend: página pública nueva en `/eventos` + feature de panel `eventos` +
ajustes de router/navegación/permisos/i18n. Sin puertas de aprobación abiertas más allá de **este
plan** (regla 1 de AGENTS.md); las 7 consultas de la puerta 2 ya están **resueltas**.

## Constitution Check

*GATE: debe pasar antes de la investigación y re-verificarse tras el diseño.*

| Principio | Evaluación | Resultado |
|---|---|---|
| §I La spec manda | El plan implementa FR-001…FR-024 e US1–US6 sin interpretarlos; cada uno tiene fila en "Cobertura de requisitos". Los huecos se cerraron como **decisiones firmes del humano** (puerta 2, 2026-10-11) en `research.md` (R4-4 permiso `categorias`, R4-5 borrado de categoría `409`, R4-8 `endTime` exige `startTime`, R4-6 imagen de elemento vencido, R4-13 aviso sin redirección). No se construye nada de Out of Scope (sin inscripciones, recordatorios, calendarios, recurrencia calculada, bloque en la portada, traducción automática, historial, verificación de enlaces, bilingüe del panel ni rediseño F10B). | ✅ |
| §II Arquitectura | Monorepo; capas `handler → service → repository` sobre `internal/platform/` (R1–R8); API REST JSON documentada en `backend/api/openapi.yaml` con el delta redactado **antes** que el código (`contracts/openapi.yaml`, §8.1.5); **0 dependencias nuevas** (R4-15). | ✅ |
| §III Pruebas | Toda tarea llevará sus pruebas (se exigirá en `tasks.md`); cobertura ≥80 % en `service/`; repository contra PostgreSQL real (`//go:build integration`), incluida la migración `000008`/`000009` (up/down/up), la regla de "próximo" con fechas fijas y la atomicidad mutación+auditoría; frontend con Vitest + Testing Library + MSW; flujos críticos con Playwright (página pública, retiro automático, publicar/retirar, permisos cruzados). **Excepción aprobada**: 3 pruebas de F2 (9 → 10 permisos) + 2 aserciones y fixture detectadas en revisión, actualizadas deliberadamente en F4 (RG4-10). | ✅ |
| §IV Seguridad | SQL solo parametrizado vía sqlc (CWE-89); validación de toda entrada en backend (CWE-20); texto plano sin ejecución (CWE-79: sin `dangerouslySetInnerHTML`); imágenes con firma binaria, tipos cerrados, sin SVG, nombre generado y descarga solo de contenido publicado y próximo (CWE-79/434, R4-6); operaciones de panel con sesión + guard + CSRF + permiso **del módulo** verificados en servidor (CWE-862; `eventos`, `actividades` y `categorias` independientes —SC-005/SC-011—); secretos: ninguno nuevo (`APP_TZ` no es secreto); `govulncheck`/`npm audit` en el CI del kit. Borradores y vencidos **nunca** hacia el público: consultas `…Upcoming` separadas (FR-013/SC-002/SC-009). | ✅ |
| §V Calidad | `gofmt`/`go vet`/`golangci-lint` sin errores; TS `strict` sin `any`; errores envueltos con `%w` y traducidos solo en `WriteError`; nombres de código en inglés (dominios `agenda`/`media` siguen la pauta ya establecida: carpeta de dominio + tablas en inglés con prefijo de dominio, como `home_*` de F3). | ✅ |
| §VI Base de datos | 3 migraciones versionadas con `up`/`down` completos (`000007`, `000008`, `000009` — el permiso `categorias` se siembra con `INSERT` en `permissions` sin tocar `000002`, que está aplicada); ninguna migración existente se modifica; tablas con `id`, `created_at`, `updated_at`; FK e índices explícitos; `CHECK`/`UNIQUE` en la base (bilingüe, horarios coherentes, alt con imagen, duplicados, registro cerrado de auditoría). | ✅ |
| §VII Observabilidad | Logs `log/slog` con `request_id` (incluidos los fallos de validación, sin cuerpos de petición); auditoría duradera ampliada a F4 en `admin_actions` (FR-017); `/healthz` intacto y sin sesión; config por variables de entorno (`APP_TZ=America/Costa_Rica` documentada en `.env.example`); `docker compose up` levanta todo (sin servicios nuevos). | ✅ |
| §VIII Gobierno | **Este plan requiere aprobación humana** antes de `/speckit.tasks` y de cualquier código (regla 1), incluidos los puntos señalados al final. Sin despliegue a producción. Quien escribe no aprueba: todo pasa por `qa-tester`, `revisor-codigo` y `seguridad`. | ✅ |

**Sin violaciones que justificar** → "Complexity Tracking" solo registra desviaciones de
convenciones internas (naming, excepción de paginación y el mapeo de auditoría repetido).

## Decisiones técnicas (resumen; análisis completo en `research.md`)

Numeración **P4-1…P4-17**, propia de este plan (no confundir con P3-1…P3-19 de F3, P1–P23 de F2,
R4-1…R4-15 de `research.md` ni RG4-1… de "Riesgos").

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| **P4-1** | **Un dominio `internal/agenda/`** para los tres conceptos de F4 (eventos, actividades y categorías) + dominio **`internal/media/`** promovido desde `portada` (descarga de archivos) | La categoría es compartida por los dos módulos y la descarga de imágenes ya es transversal (F3–F9): un dominio evita dos paquetes sobre tablas que se referencian (R2) | Un dominio por módulo (dos paquetes con el catálogo compartido y el registro de auditoría duplicado); dejar la descarga en `portada` (acoplaría la portada a las tablas de F4–F9); media en `platform/` (platform no hace SQL) |
| **P4-2** | **Modelos separados**: `agenda_events` + `agenda_event_dates` (lista explícita de fechas con horario propio) y `agenda_activities` con fecha única inline | Decisión 3 literal ("modelos distintos") y Q1/Q2: las fechas del evento son filas con su horario; la actividad es una acción de un día | Una tabla con `kind` (CHECKs condicionales, permisos cruzados); tabla de fechas también para la actividad (una hija para una fecha); `date[]`/JSON (pierden CHECK, índices y sqlc plano) |
| **P4-3** | **Retiro automático = filtro de lectura** "publicado y próximo" en SQL, con `today`/`now_time` del servidor en **`APP_TZ = America/Costa_Rica`** (valor firme de la puerta 2; parámetros, nunca concatenados) | FR-006/Q4: se retira solo sin cambiar el estado y sin procesos; el huso del lugar evita el corte UTC del contenedor | Cron que cambie estados (contradice Q4 y añade un proceso); `CURRENT_DATE` de la BD (UTC: corte desplazado); resolver "hoy" en el navegador (cada visitante vería otra cosa); dejar `APP_TZ` sin valor (resuelto: el humano fijó la zona de la iglesia) |
| **P4-4** | **Bilingüe** `*_es`/`*_en` + `?lang=es\|en` resuelto en el servidor con fallback por campo (patrón R3-1/R3-2 de F3, literal) | FR-008/SC-006 garantizados en una capa y reutilizables en F5–F9; el cliente público no reimplementa fallback | Tabla de traducciones EAV; resolver en cliente; `Accept-Language` |
| **P4-5** | **Permisos**: `AdminChain("eventos")` y `AdminChain("actividades")` independientes + **`AdminChain("categorias")`** para `/api/v1/admin/categorias`, con el **permiso nuevo `categorias`** (10.º del catálogo de F2, sembrado en `000009`; **cambio de alcance autorizado por el humano en la puerta 2**) | FR-012 y Decisión 5 siguen: `eventos`/`actividades` se **activan** sin crearse (verificado en `000002`); el catálogo de categorías es un módulo propio y su permiso no se deriva de otros. Con permiso propio, `AdminChain`/`AuthzByModule` bastan: **sin helpers nuevos de middleware** | Catálogo con cualquiera de `eventos`/`actividades` (el default anterior, **descartado por el humano**: el acceso al catálogo no debe derivarse de otros permisos); permiso único para los tres módulos (mezclaría áreas independientes); `AuthzByAnyModule`/`AdminChainAny` (sobra complejidad con permiso propio); comprobar permisos solo en el cliente (CWE-862) |
| **P4-6** | **Categorías**: `agenda_categories` con catálogo inicial **sembrado en `000007`** (Q6), CRUD en el panel (permiso `categorias`, P4-5), borrado **restringido en uso** (`FK RESTRICT` → `409`, decisión **firme** de la puerta 2) | Q6 pide un catálogo administrable; proteger el uso evita quitar etiquetas silenciosamente a contenido publicado | `ON DELETE SET NULL` (cambios invisibles en lo publicado); catálogo fijo en `CHECK` (Q6 es administrable); etiquetas libres (descartadas por Q6); borrar desligando la categoría de los elementos (descartado: el humano fijó `409`) |
| **P4-7** | **Imágenes**: `platform/storage` reutilizado (F3) + subida por módulo (`POST …/{eventos\|actividades}/imagenes` tras su permiso) + descarga `GET /api/v1/media/{fileName}` ahora en `internal/media/` con `IsFilePublished` (publicado **y** próximo — decisión **firme**: el vencido deja de servirse y vuelve al haber fecha próxima) | Un solo servicio de archivos para F3–F9 con la regla de F3 intacta ("por ninguna vía" un borrador —y ahora un vencido— expone su imagen) | Subida genérica con permiso cualquiera (difumina autorización); descarga sin comprobar estado (rompe FR-013); servir también imágenes de vencidos (descartado por el humano en la puerta 2: el vencido no se expone; la URL vuelve a responder `200` al cargar una fecha próxima) |
| **P4-8** | **Un documento público** `GET /api/v1/proximamente?lang=es\|en` con los dos bloques (`events`, `activities`; claves omitidas si vacías) y `Cache-Control: no-store` | Un solo request resuelve la página **y** la visibilidad de la navegación (FR-022); patrón idéntico al `GET /api/v1/portada` de F3 | Dos endpoints públicos (tres consultas para la navegación o menú desincronizado del cuerpo); paginación/filtros públicos (FR-003); `ETag` (la respuesta depende de la fecha actual) |
| **P4-9** | **Panel paginado** (`limit`/`offset` con `platform/paginate`, filtro opcional `publicationState`) con `GET/POST /admin/{eventos\|actividades}`, `GET/PATCH/DELETE …/{id}`, `POST …/imagenes`; el `PATCH` de `dates` **reemplaza la lista completa** | Los listados crecen (§8.1.3); el reemplazo de fechas es atómico y evita CRUD por fecha en la interfaz (R4-11) | CRUD por fecha (`/fechas/{id}`: 3 endpoints más sin ganancia para el equipo); agregado sin paginación como el de F3 (F3 era un documento acotado; aquí es un listado abierto) |
| **P4-10** | **Publicación** por elemento con `publicationState` (eventos y actividades; **no** las categorías) y "se publica al guardar" (R4-12) | FR-013/FR-014 literales; patrón probado en F3 | Estado por módulo; paso intermedio de borrador al editar (contradice FR-014); estado de publicación en categorías (son etiquetas, no contenido) |
| **P4-11** | **Validación**: tags existentes (`required`, `max`, `oneof`, `url`) + reglas de dominio (fecha de calendario real, `HH:MM`, `endTime > startTime`, **`endTime` exige `startTime`** — decisión **firme** de la puerta 2 —, ≥1 fecha por evento, ≤50, alt obligatorio con imagen) y límites como constantes del service (detalle **2.000**) | FR-015 completo sin dependencias nuevas; los mensajes salen del service con `details` por campo | Librerías de fechas/URLs; validar solo en frontend (§IV); rechazar fechas pasadas (el edge case las permite); aceptar un fin sin inicio (descartado por el humano: no es un rango) |
| **P4-12** | **Duplicados → `409 conflict`** dentro del módulo (actividad: título+fecha+lugar; evento: título+lugar+fecha en común; categoría: nombre) sobre valores normalizados | Edge case de la spec ("se rechaza o se señala"): se rechaza, coherente con F2/F3 | Solo señalizar en el formulario (la API forzada quedaría sin regla); duplicado entre módulos (prohibido por Decisión 3) |
| **P4-13** | **Auditoría**: 15 códigos `event.*`/`activity.*`/`category.*` (migración `000008` sobre `admin_actions_action_check`), `target_kind='content'`, `InsertAdminAction` transaccional y `Recorder` del propio dominio `agenda` | FR-017 pide el registro de F2; el patrón de F3 (R3-11) ya está probado y mantiene el registro cerrado en la BD | Registro nuevo (no es el de F2); que `usuarios` resuelva las rutas de F4 (acoplaría F2 a cada módulo); registrar tras el commit (rompe "no aplicarse sin registro") |
| **P4-14** | **SPA**: ruta pública **`/eventos`** (fijada por el humano en la puerta 2; la propuso `ux.md` D-1) con dos bloques dentro de `PublicLayout`; entrada «Eventos y actividades» (href `/eventos`) en `useSiteLinks` **solo si** el documento público trae algún bloque (barra, hamburguesa y pie a la vez); página sin contenido → **aviso de sección vacía** (`HiddenNotice` de `ux.md`), **sin redirigir** | Q5/FR-021/FR-022/SC-012: una entrada, dos bloques, cero enlaces rotos y cero pantallas sin explicación; el hook ya estaba preparado para rutas propias; el aviso explica la ausencia de contenido sin robarle la URL al visitante | Redirigir a `/` (descartado por el humano: URL sin explicación); dos entradas de navegación (contradice Q5); mostrar la entrada siempre (enlace a una página vacía); tocar la portada de F3 (prohibido); `/eventos-actividades` (descartado: la ruta la fija el humano) |
| **P4-15** | **Frontend de panel**: feature `features/eventos/` (inventario de `ux.md` §6.4) con `EventsPage`, `ActivitiesPage`, `CategoriesPage` bajo `RequirePermission` (**EVENTOS/ACTIVIDADES/CATEGORIAS**, uno por ruta); `ImageUploader` y `PublishControls` **se promueven al inventario único** `src/components/` | Tres módulos con permisos distintos e independientes; la regla de F2/T242 manda promover lo compartido en lugar de copiarlo; el guard `RequirePermission` existente basta (sin `RequireAnyPermission`) | Copiar los componentes de F3 (duplicado); una sola página de panel para los permisos (mezclaría módulos independientes); `RequireAnyPermission` (era para el catálogo con "cualquiera de los dos": sobra con permiso propio) |
| **P4-16** | **Contrato**: delta en `specs/004-eventos-actividades/contracts/openapi.yaml` → fusión aditiva en `backend/api/openapi.yaml` con `info.version` **0.4.0 → 0.5.0**; el registro de `error.code` **no cambia**; en la fusión se actualiza la prosa de `/api/v1/media/{fileName}` y la del catálogo de permisos (`GET /api/v1/admin/permisos` e init: 9 → **10** permisos con `categorias`) | Contrato antes que el código (§II) sin dos copias vivas ni códigos de error ad-hoc | Prefijo nuevo de rutas; `error.code` para "fecha pasada" (no es un error: simplemente no se publica); enum de permisos en el contrato (el catálogo vive en la BD y se lee dinámico) |
| **P4-17** | **Dependencias: 0 nuevas** (backend y frontend) | §II: minimizadas; la stdlib cubre todo (fechas localizadas con `Intl`) | Librería de recurrencia (prohibida por Q1), `date-fns`/`dayjs`, `react-big-calendar` (la UX es lista cronológica) |

## Project Structure

### Documentation (esta funcionalidad)

```text
specs/004-eventos-actividades/
├── plan.md              # EDITABLE (este archivo, /speckit.plan)
├── research.md          # EDITABLE (fase 0: R4-1…R4-15)
├── data-model.md        # EDITABLE (fase 1: tablas, migraciones 000007–000009, consultas)
├── quickstart.md        # EDITABLE (fase 1: cómo probar F4 en local)
├── spec.md              # APROBADA — no se edita
├── contracts/
│   └── openapi.yaml     # EDITABLE (fase 1): delta de diseño, se fusiona en backend/api/openapi.yaml
├── ux.md                # DEL disenador-ux (ya producido; sus decisiones D-1/D-6/D-13 confirmadas
│                        #   por el humano están aplicadas en este plan; la composición final de
│                        #   UI la fija ux.md y las rutas/DTOs las fija este plan)
└── tasks.md             # Fase 2 (/speckit.tasks — NO lo crea este comando)
```

### Source Code — backend

```text
backend/
├── cmd/api/
│   ├── main.go                      # EDITADO: DI de los dominios agenda y media
│   │                                #   + registro de /api/v1/proximamente, /api/v1/media y
│   │                                #   /api/v1/admin/{eventos,actividades,categorias}
│   └── main_test.go                 # EDITADO: prueba de humo ampliada (rutas nuevas + /healthz intacto)
├── internal/
│   ├── db/
│   │   ├── queries/
│   │   │   ├── agenda.sql           # NUEVO: consultas de agenda_* (CRUD, fechas, duplicados,
│   │   │                            #   variantes …Upcoming para el público)
│   │   │   ├── media.sql            # NUEVO: IsFilePublished (home_* + agenda_*; sustituye a
│   │   │                            #   IsHomeFilePublished de home.sql)
│   │   │   ├── home.sql             # EDITADO: se retira IsHomeFilePublished (pasa a media.sql)
│   │   │   └── audit.sql            # SIN CAMBIOS: se reutiliza InsertAdminAction (R4-10)
│   │   └── (generado por sqlc, commiteado)
│   ├── agenda/                      # NUEVO — dominio F4 (receta de arq. §8)
│   │   ├── model.go                 # entidades (Event, EventDate, Activity, Category) y DTOs
│   │   │                            #   (etiquetas validate, Es/En, Fechas)
│   │   ├── repository.go            # constructor + mapRow/params (pgtype → dominio) + withTx
│   │   │                            #   + inserción de admin_actions sobre el tx (audit.Action → params)
│   │   ├── repository_eventos.go    # eventos + fechas (ReplaceEventDates, ExistsEventDuplicate)
│   │   ├── repository_actividades.go# actividades
│   │   ├── repository_categorias.go # catálogo + CountCategoryUsage
│   │   ├── service.go               # tipos comunes, invariantes, constantes (límites, ≤50 fechas),
│   │   │                            #   cálculo de today/now_time con APP_TZ=America/Costa_Rica y
│   │   │                            #   regla "próximo"
│   │   ├── service_public.go        # "lo que viene": solo published+próximo, resolución de idioma,
│   │   │                            #   orden cronológico, claves omitidas si vacías
│   │   ├── service_admin.go         # CRUD de eventos/actividades/categorías, validaciones de
│   │   │                            #   dominio (fechas, horas, duplicados, imagen+alt), publicar/
│   │   │                            #   retirar y auditoría transaccional
│   │   ├── service_audit.go         # audit.Recorder del dominio: método+ruta → acción
│   │   │                            #   event.*/activity.*/category.* (denegaciones)
│   │   ├── handler.go               # constructor + helpers HTTP (decodificar/validar/responder)
│   │   ├── handler_public.go        # GET /api/v1/proximamente
│   │   ├── handler_admin.go         # CRUD de /api/v1/admin/{eventos,actividades,categorias}
│   │   ├── handler_images.go        # POST /api/v1/admin/{eventos,actividades}/imagenes (multipart)
│   │   ├── routes.go                # RegisterPublic (proximamente) y RegisterAdmin (3 grupos:
│   │   │                            #   AdminChain("eventos"), AdminChain("actividades"),
│   │   │                            #   AdminChain("categorias"))
│   │   └── *_test.go                # service con fakes · handler con httptest · repository integration
│   ├── media/                       # NUEVO (promovido de portada, R4-6): descarga pública
│   │   ├── handler_media.go         # GET /api/v1/media/{fileName} (regex, nosniff, inline, no-store)
│   │   ├── repository.go            # IsFilePublished (queries/media.sql)
│   │   ├── routes.go                # RegisterPublic (/api/v1/media)
│   │   └── *_test.go                # se mueven/ajustan las pruebas de descarga de F3
│   ├── portada/                     # EDITADO: se retiran handler_media.go y la consulta
│   │                                #   IsHomeFilePublished (pasan a media/; comportamiento igual)
│   ├── usuarios/                    # EDITADO (excepción aprobada, RG4-10): 3 pruebas de F2 pasan
│   │                                #   de 9 a 10 permisos (repository_roles_integration_test.go:52,
│   │                                #   service_init_integration_test.go:57, service_init_test.go:199)
│   │                                #   + 2 aserciones más (service_roles_integration_test.go:69 y
│   │                                #   service_roles_test.go:464) + fixture sampleCatalog() a 10 y
│   │                                #   comentarios ("los 9 permisos"); NADA más de F2 cambia
│   │                                #   (el catálogo se lee dinámico)
│   └── platform/
│       ├── middleware/chain.go      # SIN CAMBIOS: AdminChain("categorias") reutiliza
│       │                            #   AuthzByModule existente (sin AuthzByAnyModule)
│       ├── audit/audit.go           # EDITADO: + 15 códigos event.*/activity.*/category.*
│       ├── validate/validate.go     # SIN CAMBIOS (el tag `url` de F3 basta)
│       └── config/config.go         # EDITADO: + APP_TZ (zona del corte de "hoy"; valor firme
│                                    #   America/Costa_Rica, validada como zona IANA)
├── migrations/
│   ├── 000007_create_agenda_content.up.sql / .down.sql          # NUEVO: 4 tablas agenda_* + catálogo
│   │                                                            #   de categorías
│   ├── 000008_extend_admin_actions_for_agenda.up.sql / .down.sql # NUEVO: CHECK de auditoría +15 códigos
│   │                                                            #   (incluidos los 3 category.*)
│   └── 000009_seed_permission_categorias.up.sql / .down.sql      # NUEVO: INSERT del permiso
│                                                                #   `categorias` en permissions (F2)
└── api/openapi.yaml                 # EDITADO: se fusiona el delta (info.version → 0.5.0) + make api-gen
```

### Source Code — frontend y raíz

```text
frontend/src/
├── api/
│   ├── agenda.ts                    # NUEVO: getProximamente(lang) + panel (list/create/update/delete,
│   │                                #   categorías) + uploadImage
│   └── schema.d.ts                  # REGENERADO (npm run api-gen)
├── app/
│   ├── router.tsx                   # EDITADO: '/eventos' (pública) y '/panel/eventos',
│   │                                #   '/panel/actividades', '/panel/categorias', cada una con
│   │                                #   RequirePermission(EVENTOS | ACTIVIDADES | CATEGORIAS)
│   ├── layout.tsx                   # EDITADO: entradas del menú del panel filtradas por permiso
│   └── guards.tsx                   # SIN CAMBIOS (RequirePermission de F2 basta; ya no hace falta
│                                    #   RequireAnyPermission)
├── features/
│   ├── publico/                     # EDITADO: sitio público (F10 aplicado, no redefinido)
│   │   ├── eventos/                 # NUEVO (inventario de ux.md §6.3):
│   │   │   ├── EventsPage.tsx       #   H1 + dos bloques; sin contenido → HiddenNotice (aviso,
│   │   │                            #   SIN redirigir, R4-13)
│   │   │   ├── EventsSection.tsx / ActivitiesSection.tsx / EventCard.tsx / ActivityCard.tsx
│   │   │   ├── CategoryChip.tsx / DateBadge.tsx (<time datetime>, Intl) / MapLink.tsx
│   │   │   ├── HiddenNotice.tsx     #   aviso de sección vacía (página mínima, sin datos)
│   │   │   ├── hooks/useUpcomingEvents.ts # TanStack Query ['proximamente', lang]
│   │   │   └── formatoFechas.ts     #   fecha/hora localizada por idioma (Intl)
│   │   ├── hooks/useSiteLinks.ts    # EDITADO: SiteLink con href de ruta; entrada
│   │   │                            #   «Eventos y actividades» (href /eventos) solo si hay bloque
│   │   ├── i18n/{messages,es,en}.ts # EDITADO: claves tipadas nuevas (nav, bloques, fechas, estados)
│   │   └── schedule.ts / format     # EDITADO: formatLongDate(lang) + reutiliza formatTimeRange
│   ├── eventos/                     # NUEVO: gestión en el panel (en español, FR-016; ux.md §6.4)
│   │   ├── pages/EventsPage.tsx / ActivitiesPage.tsx / CategoriesPage.tsx
│   │   ├── components/              # EventForm, ActivityForm (comparten ItemFormFields),
│   │   │                            #   DatesEditor (lista de fechas del evento), CategoryForm,
│   │   │                            #   EventsList, ActivitiesList, EstadoBadge, DeleteConfirm
│   │   ├── hooks/                   # useEvents, useCreateEvent, useUpdateEvent, useDeleteEvent,
│   │   │                            #   useActivities…, useEventCategories…, useUploadAgendaImage
│   │   └── messages.ts / *.test.tsx
│   ├── informacion/
│   │   └── components/              # EDITADO: ImageUploader y PublishControls se promueven a
│   │                                #   src/components/ (inventario único, regla T242)
│   ├── roles/permissions.ts         # EDITADO: isPermissionAvailable incluye 'eventos',
│   │                                #   'actividades' y 'categorias' (se retira el sello
│   │                                #   «Disponible más adelante»); permissions.test.ts y
│   │                                #   roles.test.tsx se actualizan en consecuencia
│   │                                #   (el formulario de roles ofrece el permiso nuevo solo:
│   │                                #   el catálogo lo lee dinámico de la API)
│   └── (resto SIN CAMBIOS)
├── lib/permissions.ts               # EDITADO: + EVENTOS = 'eventos', ACTIVIDADES = 'actividades',
│                                    #   CATEGORIAS = 'categorias'
├── components/                      # EDITADO: + ImageUploader, PublishControls (promovidos)
└── (index.css, tailwind.config)     # SIN CAMBIOS (tokens de F10/F3)
frontend/e2e/
├── agenda-publica.spec.ts           # NUEVO: visitante (US1, US5, US6)
└── agenda-panel.spec.ts             # NUEVO: gestión y permisos (US2, US3, US4)

# Raíz
.env.example                         # EDITADO: + APP_TZ=America/Costa_Rica (documentada)
README.md                            # EDITADO por documentador al cerrar
docker-compose.yml                   # SIN CAMBIOS (el volumen uploads_data de F3 ya existe)
Makefile / .github/workflows/ci.yml / .githooks/   # SIN CAMBIOS (son del kit)
```

### Inventario único de componentes UI

F4 **promueve dos componentes** al inventario único (`ImageUploader` y `PublishControls`, hoy en
`features/informacion/`, porque los módulos de F4 los usan — regla de F2/T242) y **no añade
más**: los componentes propios de F4 viven en `features/publico/eventos/` y
`features/eventos/components/` (nombres del inventario de `ux.md` §6.3/§6.4) y **reutilizan** el
inventario (`Field`, `Select`, `Button`, `Notice`, `ConfirmDialog`, `Dialog`, `EmptyState`, `Table`,
`Tabs`, `StatusPill`) y el
sistema de F10 (`Card`, `SectionHeading`, `PublicButton`, `Reveal`, `PublicLayout`, `PublicNav`,
`MobileMenu`, `FooterSiteMap`, `LanguageSwitcher`, `ThemeToggle`).

## Cadena de middleware y grupos (orden; el primero es el más externo)

```text
Global:   request-id → recover → logging → CORS(con credenciales) → handler
          /api/v1/proximamente              → (sin middlewares)  → handler    ← público, solo lectura
          /api/v1/media/{fileName}          → (sin middlewares)  → handler    ← público, solo lectura
          /api/v1/admin/eventos             → authn → guard de cambio de contraseña → authz("eventos") → CSRF → handler
          /api/v1/admin/actividades         → authn → guard de cambio de contraseña → authz("actividades") → CSRF → handler
          /api/v1/admin/categorias          → authn → guard de cambio de contraseña → authz("categorias") → CSRF → handler
```

- La superficie pública de F4 es de **solo lectura**: sin `authn`, `CSRF` ni `rate-limit` (F2 solo
  lo monta sobre superficies públicas escribibles; aquí no hay ninguna).
- Los tres grupos de panel montan la **cadena aprobada de F2**, con `AuthzByModule` tal cual (la
  única diferencia es el código de permiso; con el permiso propio `categorias` **no hace falta**
  ningún helper nuevo — se descarta `AuthzByAnyModule`, R4-4). La denegación responde
  `403 forbidden` genérico y **queda registrada** en `admin_actions` (`result='denied'`); el
  `Recorder` que recibe cada cadena es el **del dominio `agenda`** (resuelve sus rutas → acción
  `event.*`/`activity.*`/`category.*`, R4-10).
- `/healthz` **no cambia**: sigue público, sin sesión y con `Cache-Control: no-store`.

## Contrato OpenAPI: dónde vive y cómo se mantiene

Igual que F1/F2/F3 (reglas §8.1.5 y §8.1.10):

1. **Diseño (esta fase)**: `specs/004-eventos-actividades/contracts/openapi.yaml` — delta inmutable
   una vez aprobado el plan.
2. **Fusión**: la primera tarea de implementación lo funde en `backend/api/openapi.yaml` de forma
   **aditiva**, aplica los cambios documentados del delta (enum `action` de `AdminActionItem` +15
   códigos, prosa de `/api/v1/media/{fileName}` y prosa del catálogo de permisos: `GET
   /api/v1/admin/permisos` e init de F2 pasan de "9 permisos" a **10** con el código `categorias`) y
   sube `info.version` a **0.5.0**; luego `make api-gen`.
3. **Consumo**: el frontend genera tipos solo desde `backend/api/openapi.yaml`.
4. El snapshot de `specs/` no se vuelve a editar.

## Estrategia de pruebas

Cobertura exigida: **80 % en `service/`** (§III). Comandos: `go test ./...` ·
`go test -tags=integration ./...` · `npm test -- --run` · `make e2e` · `make ci`.

| Capa | Tipo | Qué verifica F4 | Dónde |
|---|---|---|---|
| `platform/config` | Unitaria | `APP_TZ = America/Costa_Rica` → `today`/`now_time` del corte de "hoy"; zona IANA inválida → error claro al arrancar (fail-closed) | `internal/platform/config/*_test.go` |
| `platform/audit` | Unitaria | Los 15 códigos nuevos en el registro cerrado (incluidos los 3 `category.*`); objetivo `content` con `targetLabel` | `internal/platform/audit/*_test.go` |
| `service` (dominio `agenda`) | Unitaria con **fakes** | **Público**: solo `published` **y próximo** (tres casos de la regla R4-2: futuro, hoy sin fin, hoy con fin ya terminado), orden cronológico con desempate estable, `dates` sin pasadas, evento sin fecha próxima no aparece, claves omitidas si vacías, fallback `en → es` por campo (nunca huecos), `lang` inválido → `invalid`. **Panel**: validación completa de FR-015 (obligatorios, español presente, 2.000 caracteres, fechas reales, `HH:MM`, `endTime > startTime`, `endTime` exige `startTime`, ≥1 fecha y ≤50, fechas repetidas, URL de mapa, categoryId desconocido, alt obligatorio con imagen, normalización `""` → `NULL`), duplicados → `conflict`, publicar/retirar por elemento, "se publica al guardar" (FR-014), categoría en uso → `conflict` al borrar, señal de "falta la próxima fecha" (`nextDate: null`). **Auditoría**: cada mutación registra su acción con el código y `targetLabel` correctos en la misma transacción (fail-closed); denegaciones y rechazos registrados best-effort | `internal/agenda/service_*_test.go` |
| `handler` | Unitaria con `httptest` + service falso | Cada operación: decodificación, `details` por campo en 400, códigos 200/201/204/400/401/403/404/409, sobres uniformes, **el DTO público nunca contiene `publicationState` ni fechas pasadas**; multipart: sin archivo, archivo gigante (400 con `details.file`), extensión no permitida | `internal/agenda/handler_*_test.go` |
| `repository` | **Integración** (PostgreSQL real, `//go:build integration`) | SQL real: `UNIQUE (event_id, event_date)`, `CHECK` de horas/longitudes/estado, `UNIQUE (title_es, activity_date, place_es)`, `ExistsEventDuplicate`, `FK RESTRICT` de categoría → error traducido a `conflict`, `ReplaceEventDates` (reemplazo completo), lecturas `…Upcoming` sin borradores ni vencidos (fechas fijas), `IsFilePublished` (publicado **y** próximo; al pasar la fecha → `FALSE` sin borrar), **mutación + `admin_actions` atómicas** (y rollback si falla el registro), migración `000008`/`000009` up→down→up | `internal/agenda/repository_*_test.go`, `internal/media/*_test.go` |
| `cmd/api` | Humo de composición | Rutas nuevas con su sobre y permiso, `/healthz` intacto y `404` `not_found` en rutas inexistentes (§8.1.9) | `cmd/api/main_test.go` |
| Frontend | Unitaria (Vitest + MSW) | Pública: dos bloques, orden cronológico, fechas/horas localizadas por idioma (`<time>`), fallback en→es visible, bloque/entrada de navegación ocultos sin contenido, **aviso de sección vacía (`HiddenNotice`) sin redirigir**, `alt` en imágenes, enlace de mapa. Panel: formularios con errores por campo (incluidos `endTime` y fechas), editor de fechas del evento, publicar/retirar por elemento, sello "Disponible más adelante" retirado de `eventos`/`actividades`/`categorias` (y sus tests de F2 actualizados), entradas y rutas filtradas por permiso (`RequirePermission(CATEGORIAS)` en categorías), subida de imagen con error de tipo/tamaño | `frontend/src/features/*/*.test.tsx` |
| `usuarios` (F2 — **excepción aprobada**, RG4-10) | Unitaria + integración | El catálogo devuelve **10** permisos (actualización deliberada de las aserciones que fijaban 9: las 3 señaladas por el humano + `service_roles_integration_test.go:69` y `service_roles_test.go:464`, más la fixture `sampleCatalog()`); el init sigue creando el rol «Administrador» con **todo** el catálogo (ahora incluye `categorias`) | `backend/internal/usuarios/*_test.go` |
| E2E pública | Playwright (local) | Recorrido de `quickstart.md`: visitante sin sesión ve lo próximo con categoría y detalle, en orden cronológico; un elemento con fecha de ayer **no** aparece; sin contenido, `/eventos` muestra el **aviso** (sin redirigir); cambia a inglés (interfaz + contenidos con fallback) y vuelve; 3 anchos de pantalla (SC-001, SC-002, SC-006, SC-007, SC-009, SC-012) | `frontend/e2e/agenda-publica.spec.ts` |
| E2E panel | Playwright (local) | Recorrido de `quickstart.md`: login con permiso, crear+publicar un evento con fechas y una actividad en <2 min, editar y ver el cambio en público desde la primera carga (SC-003/SC-004), retirar, eliminar con confirmación, denegación cruzada (cuenta solo `eventos` frente a actividades: SC-005/SC-011) y registro de todo en auditoría (SC-013) | `frontend/e2e/agenda-panel.spec.ts` |

**Pruebas que no faltan** (las que la spec hace críticas): **0 borradores ni vencidos expuestos** por
ninguna respuesta pública ni por la descarga de imágenes (SC-002/SC-009); **retiro automático**
observado sin intervención (SC-009: fixtures con fecha pasada y con fin de hora ya terminado);
**publicar/retirar por elemento** y sección/navegación ocultas sin contenido (SC-012); **cambio
guardado visible en la primera carga** (SC-003); **fallback en→es sin campos vacíos** (SC-006);
**permisos independientes verificados en servidor**, incluido el cruce `eventos`↔`actividades`
(SC-005/SC-011); **toda operación queda registrada** —éxito, fallo y denegación— y el registro sigue
siendo de solo lectura (SC-013); **texto con tildes, `ñ`, `¿¡` y emojis intacto** (edge case);
**archivo inválido no rompe la sección** (edge case de imágenes); **dos elementos en la misma fecha**
en orden estable y reproducible (edge case).

## Cobertura de requisitos (US1–US6 y FR-001…FR-024)

| Historia | Dónde se cubre | Verificación |
|---|---|---|
| **US1** — Descubrir lo que viene (P1) | `GET /api/v1/proximamente` (P4-8/P4-3) + `EventsPage` en `/eventos` con dos bloques (P4-14) | quickstart §1 · e2e pública (SC-001, SC-009) |
| **US2** — Gestionar eventos (P1) | `/api/v1/admin/eventos` (P4-5/P4-9/P4-11/P4-13) + `features/eventos/EventsPage` (P4-15) | quickstart §3–§5 · e2e panel (SC-004, SC-005) |
| **US3** — Gestionar actividades (P1) | `/api/v1/admin/actividades` (P4-5/P4-9) + `ActivitiesPage` (P4-15) | quickstart §4–§6 · e2e panel (SC-004, SC-005) |
| **US4** — Controlar la publicación (P2) | `publicationState` + consultas `…Upcoming` (P4-10/P4-3) + `PublishControls` (promovido) | quickstart §7 · e2e panel (SC-002, SC-003, SC-012) |
| **US5** — Ver en inglés (P2) | `?lang=` con fallback en el servidor (P4-4) + i18n tipado + fechas/horas localizadas (R4-3) | quickstart §8 · e2e pública (SC-006) |
| **US6** — Cualquier dispositivo y accesible (P2) | Sistema de F10 aplicado (P4-14) + reglas de accesibilidad de la skill `react-frontend` + `alt` obligatorio | quickstart §10 (checklist) · e2e en 3 anchos (SC-007, SC-008) |

| FR | Dónde se resuelve en este plan | Verificación |
|---|---|---|
| FR-001 | P4-8 (documento público con los dos conceptos diferenciados) + P4-14 (una entrada, dos bloques) | quickstart §1 · e2e pública (SC-001) |
| FR-002 | P4-3 (filtro "publicado y próximo" en el servidor) | quickstart §2 y §9 · pruebas de service/repository (SC-009) |
| FR-003 | P4-8 (orden por próxima fecha con desempate estable; sin filtros ni paginación) | quickstart §1 · pruebas de orden del service |
| FR-004 | P4-2 (`agenda_events` + `agenda_event_dates`; `dates[0]` es la próxima fecha; `frequency` opcional) | quickstart §3 · pruebas de service/repository |
| FR-005 | P4-2 (`agenda_activities` con fecha única + `start_time`/`end_time`; mismo criterio en cada fecha del evento) | quickstart §4 · pruebas de validación (Q2) |
| FR-006 | P4-3 (retiro por filtro de lectura; el panel marca `isPast`/`nextDate: null` sin cambiar el estado) | quickstart §9 · e2e con fecha pasada (SC-009) |
| FR-007 | P4-4 (columnas `*_es`/`*_en` en todos los textos) | quickstart §5 y §8 |
| FR-008 | P4-4 (fallback `en → es` por campo en el service) | quickstart §8 · tabla de casos del service (SC-006) |
| FR-009 | P4-4/R4-3 (`LanguageSwitcher` existente; fechas y horas localizadas con `Intl` y `formatTimeRange`, sin alterar el valor) | quickstart §8 · e2e pública (cambio de idioma) |
| FR-010 | `/api/v1/admin/eventos` (P4-9) + `EventsPage` (P4-15) | quickstart §3–§5 · e2e panel (SC-004) |
| FR-011 | `/api/v1/admin/actividades` (P4-9) + `ActivitiesPage` (P4-15) | quickstart §4–§6 · e2e panel (SC-004) |
| FR-012 | P4-5 (`AdminChain("eventos")`/`AdminChain("actividades")`/`AdminChain("categorias")` en servidor + guards `RequirePermission`; permiso nuevo `categorias` para el catálogo) | quickstart §11 · pruebas de handler (403 por API forzada, cruce de permisos) · e2e (SC-005, SC-011) |
| FR-013 | P4-10 (`publication_state` por elemento) + P4-3 (consultas `…Upcoming`) + P4-14 (sección oculta sin contenido) | quickstart §7 y §9 · pruebas "0 borradores/vencidos expuestos" · e2e (SC-002, SC-012) |
| FR-014 | P4-10/P4-3 (guardado de elemento publicado visible en la primera carga; `no-store`) | quickstart §7 · e2e panel (SC-003) |
| FR-015 | P4-11 (obligatorios, fechas reales, horas coherentes, URL de mapa, español presente, límites, imagen+alt) | quickstart §3–§6 (tablas de errores) · pruebas de service/handler |
| FR-016 | Panel en español (convención de F2/F3; solo el sitio público lleva `i18n/`) | revisión de `revisor-codigo` (0 cadenas de panel pasadas por `t()`) |
| FR-017 | P4-13 (`admin_actions` transaccional + 15 códigos + denegaciones registradas; solo lectura) | quickstart §12 · pruebas de service/repository (SC-013) |
| FR-018 | `agenda_events` + `agenda_event_dates` + `agenda_categories` (P4-2/P4-6) | quickstart §3 · pruebas de persistencia |
| FR-019 | `agenda_activities` (P4-2) | quickstart §4 · pruebas de persistencia |
| FR-020 | Imagen opcional + `CHECK` de alt obligatorio + subida validada (P4-7) | quickstart §10 · pruebas de service/handler (SC-008 parcial) |
| FR-021 | P4-14 (una entrada «Eventos y actividades» con dos bloques, sistema de F10 aplicado, portada intacta) | quickstart §1 · revisión de `revisor-codigo` sobre `PublicLayout`/`HomePage` sin cambios |
| FR-022 | P4-14 (`useSiteLinks`: entrada solo con contenido; barra, hamburguesa y pie a la vez) | quickstart §9 · e2e (SC-012) |
| FR-023 | Sistema de F10 + reglas `react-frontend` (toc-targets ≥44 px, títulos claros) | quickstart §10 · e2e en 320/768/1280 (SC-007) |
| FR-024 | Accesibilidad: teclado, semántica, foco, contraste claro/oscuro, zoom 200 %, `alt` obligatorio (P4-7/P4-14), fechas legibles (R4-3) | quickstart §10 (checklist WCAG 2.1 AA) · e2e de teclado (SC-008) |

## Métricas del plan (coherentes con los SC de la spec)

| SC | Qué se mide | Cómo se mide en F4 | Umbral |
|---|---|---|---|
| SC-001 | Todo visitante ve lo que viene (título, fecha, lugar, detalle) sin cuenta y lo encuentra en <30 s | e2e pública: la sección muestra ambos bloques con sus datos en la primera carga | 100 % |
| SC-002 | 0 borradores visibles al público (tampoco en imágenes) | Pruebas "0 borradores expuestos" (service + handler + repository + `IsFilePublished`) y e2e con borradores en BD | 0 |
| SC-003 | Cambio guardado visible desde la primera carga posterior | e2e panel: editar → recargar → ver el dato (`no-store` de respaldo) | 100 % |
| SC-004 | Crear y publicar un evento o actividad en <2 min desde el panel | e2e/observación (quickstart §3–§6) | < 2 min |
| SC-005 | Toda operación verifica el permiso del módulo, **incluido el cruce** (`eventos`/`actividades`/`categorias`) | Pruebas de handler por operación (403) + e2e con cuenta solo `eventos` frente a actividades y frente al catálogo de categorías | 0 operaciones sin permiso |
| SC-006 | Con `lang=en`: 100 % de la interfaz traducida y 0 campos vacíos (fallback a es) | Pruebas de i18n (claves tipadas) + tabla de casos de fallback + e2e en inglés | 100 % / 0 |
| SC-007 | Usable en teléfono, tableta y computadora sin desplazamiento horizontal (320 px mínimo) | e2e en 3 anchos (320/768/1280) + checklist manual (quickstart §10) | 0 pérdida / 0 scroll |
| SC-008 | Accesibilidad: 0 hallazgos críticos (referencia WCAG 2.1 AA, claro y oscuro) | Checklist manual de quickstart §10 (teclado, lector, contraste, zoom) revisada por QA | 0 críticos |
| SC-009 | 0 elementos con fecha pasada entre "lo que viene"; retiro automático en el 100 % de los casos | Pruebas de la regla R4-2 (3 casos por elemento) + e2e con fixtures vencidos + evento sin fecha próxima nunca vacío | 0 / 100 % |
| SC-010 | ≥90 % de personas de distintas edades identifican qué viene, cuándo y dónde sin ayuda | Prueba manual con personas (protocolo en `quickstart.md` §10, estilo F3/M7): ≥6 personas (2 por franja 18–35/36–59/60+), guion de 3 tareas, informe `pruebas-usabilidad-SC-010.md` | ≥ 90 % |
| SC-011 | Sin el permiso del módulo no se ve ni modifica nada (100 % de intentos) | e2e con cuenta sin permiso (UI oculta + API forzada 403) + cruce de permisos | 100 % denegados |
| SC-012 | 0 secciones vacías ni campos en blanco; navegación sin enlaces rotos | Bloques/entrada omitidos sin contenido (pruebas de contrato y de `buildSiteLinks`) + e2e + la URL escrita a mano muestra el **aviso de sección vacía** (sin redirigir, R4-13) | 0 |
| SC-013 | 100 % de operaciones registradas (quién, qué, sobre qué, cuándo); 0 registros editables/borrables | quickstart §12 + pruebas transaccionales + contrato sin escritura sobre el registro | 100 % / 0 |

## Riesgos

> Numeración **RG4-1…RG4-10**, propia de este plan (para no colisionar con RG3-1…RG3-12 de F3 ni
> R4-1…R4-15 de `research.md`).

| Riesgo | Impacto | Mitigación |
|---|---|---|
| **RG4-1 — Zona horaria (`APP_TZ`) mal configurada** | El corte de "hoy" se desplaza y un elemento se retira pronto/tarde (SC-009 se ve afectado) | El valor está **fijado por el humano**: `APP_TZ = America/Costa_Rica` (puerta 2); documentado en `.env.example` y en el quickstart; el service valida que sea una zona IANA y falla con error claro si no (fail-closed, mejor que un corte silencioso); pruebas con fechas fijas cubren la regla sin depender de la hora real |
| **RG4-2 — La regla "próximo" vive en varias consultas** (`…Upcoming` + `IsFilePublished`) y puede divergir | Un vencido seguiría visible por una vía (incumple FR-006/SC-009) | Una sola función Go de dominio calcula `today`/`now_time` y las consultas comparten el mismo predicado (copiado en SQL, probado por separado en cada camino); pruebas dedicadas en listing y descarga con los mismos fixtures |
| **RG4-3 — Promover la descarga a `internal/media/` toca código de F3** | Regresión en la descarga del logo/imagen de portada | El handler y sus pruebas se **mueven** sin reescribirse (mismo comportamiento, mismos casos); los e2e de F3 (`portada-imagenes`) siguen corriendo como evidencia; la consulta nueva cubre `home_identity` con los mismos criterios que `IsHomeFilePublished` |
| **RG4-4 — Tres permisos en un solo feature** (módulos cruzados; el catálogo de categorías es compartido pero tiene permiso propio) | Un hueco de autorización daría acceso cruzado (SC-005/SC-011) | Cada grupo de rutas monta su `AdminChain` en `routes.go` (auditable de un vistazo: `eventos`, `actividades`, `categorias`); pruebas de handler por operación con cuentas de cada permiso; sin `AuthzByAnyModule` (superficie de autorización mínima); auditoría de `seguridad` específica sobre el cruce |
| **RG4-5 — `ux.md` del disenador-ux se produce en paralelo** | Las tareas de frontend podrían diseñar pantallas que no encajan con los DTOs | Este plan fija rutas, DTOs, reglas de contenido y bloques: el diseño puede variar la composición **sin cambiar la API**; `ux.md` ya existe y sus nombres de pieza son la referencia de `tasks.md`; divergencias de copia/interfaz detectadas (aviso D-13, modal de borrado D-16) se alinean en tareas `[frontend]` sin tocar la API |
| **RG4-6 — Lista de fechas reemplazada en bloque** (ediciones simultáneas, R4-11) | La última escritura completa gana y puede pisar fechas recién cargadas por otra persona | Es el default aprobado de la spec; la escritura es atómica (nunca una mezcla); el panel muestra las fechas actuales al abrir el formulario; si el humano lo pide más adelante: columna `version` + `409` (cambio acotado) |
| **RG4-7 — Duplicados "título + lugar + fecha" con variantes de redacción** ("Vigilia" / "vigilia ") | Falsos negativos (duplicados que pasan) o falsos positivos (bloqueos indebidos) | Normalización documentada (`btrim` + minúsculas) igual que F2/F3; el `409` es un default revisable y su regla está en el contrato; si el equipo prefiere "solo señalar", es un cambio menor del service |
| **RG4-8 — Muchos elementos próximos** (sin límite artificial, default revisable) | Respuestas grandes en móvil | El service acota con una constante de seguridad (200 por bloque) y ordena cronológicamente; si el volumen real lo exige, "ver más"/paginación es el ajuste menor que la propia spec deja acordado con el humano |
| **RG4-9 — Accesibilidad y usabilidad (SC-008/SC-010) no cubiertas por automatización** | Hallazgos críticos en la revisión manual | Checklist explícita en `quickstart.md` §10 ejecutada por `qa-tester`; componentes semánticos de F10 (ya auditados en F10); `alt` obligatorio en la BD; protocolo de usabilidad con personas (SC-010) como en F3 |
| **RG4-10 — F4 edita pruebas de F2 (funcionalidad cerrada) y amplía el catálogo de la spec de F2** (9 → 10 permisos con `categorias`) | Regresión inadvertida en F2, o lectura como "debilitamiento" de pruebas; la spec de F2 queda ampliada por un cambio de producto | Cambio **autorizado por el humano** (puerta 2) y registrado como **excepción aprobada**: la tarea es explícita y acotada (las 3 pruebas fijadas por el humano + 2 aserciones detectadas en revisión + fixture `sampleCatalog()` + comentarios "los 9 permisos"); solo cambia el **recuento** del catálogo (ninguna aserción de comportamiento se relaja); la hace `dev-backend` con sus pruebas y la validan `revisor-codigo`/`seguridad`; el resto de F2 lee el catálogo **dinámicamente** (init, `ListPermissions`, formulario de roles) y no se toca; `000002` nunca se edita (el permiso se siembra en `000009`) |

## Complexity Tracking

> Sin violaciones de la constitución. Desviaciones de **convenciones internas**, declaradas:

1. **Nombres de dominio y tablas** (P4-1): la pauta establecida (F3) es carpeta de dominio + tablas
   en inglés con prefijo de dominio (`portada` → `home_*`); F4 la sigue: dominios `agenda` y `media`
   con tablas `agenda_*`. Si `revisor-codigo` prefiere `internal/eventos/`, es un renombrado sin
   consecuencias (todo el cableado está en `main.go`).
2. **El público no pagina** (§8.1.3): `GET /api/v1/proximamente` sirve **todos** los próximos porque
   FR-003 prohíbe exigir paginación al visitante; se acota con una constante de seguridad del
   service (200 por bloque, RG4-8) y el contrato lo documenta. Los listados de panel **sí** paginan
   con `platform/paginate`. El catálogo de categorías es una colección acotada (≤100) sin
   paginación, como las de F3.
3. **Mapeo `audit.Action` → `InsertAdminActionParams` repetido** (R4-10): ~10 líneas en el repository
   de `agenda` y otras tantas en `portada`/`usuarios`, todas sobre la **misma** consulta nombrada
   de `internal/db` (misma decisión que Complexity 4 de F3; subirlo a `platform/` arrastraría
   `pgtype` fuera del diagrama de dependencias).
4. **Pruebas de F2 editadas por F4 (9 → 10 permisos)** (RG4-10): desviación de la convención "quien
   escribe no toca lo cerrado", **autorizada por el humano** como parte del cambio de alcance del
   catálogo. Se acota a aserciones de recuento (3 señaladas + 2 detectadas en revisión), la fixture
   `sampleCatalog()` y comentarios; queda registrada como **excepción aprobada** y validada por los
   revisores. Sin ella, `make ci` fallaría por una decisión de producto ya tomada.
5. **`endTime` exige `startTime`** (R4-8): la spec solo dice "hora de inicio y hora de fin
   opcionales"; se interpretó el par como rango y la puerta 2 lo convirtió en **decisión firme**:
   un fin sin inicio se rechaza con `details.endTime`.
6. **Regla de la imagen de un elemento vencido** (P4-7): la descarga exige "publicado **y** próximo"
   (una excepción *más* estricta que F3, que solo mira "publicado"). Coherente con FR-006 ("se
   retira de la vista pública") y con `analyze` C4 de F3; **decisión firme** de la puerta 2: el
   vencido no se expone y la URL vuelve a responder `200` al haber fecha próxima.
7. **Nombres de pieza de UI según `ux.md`** (R4-14): `features/publico/eventos/` y
   `features/eventos/` con los nombres del inventario de `ux.md` §6.3/§6.4 (`EventsPage`,
   `useUpcomingEvents`, `HiddenNotice`…). Es la alineación de nomenclatura que este plan dejaba
   abierta (punto 7 de aprobación): rutas `/eventos` y `/panel/{eventos,actividades,categorias}`
   fijadas por el humano; dominios backend `agenda`/`media` y tablas `agenda_*` sin cambios.

## Propuesta de tareas (insumo para `/speckit.tasks`; **no** es `tasks.md`)

Orden típico migración → contrato → repositorio → servicio → handler → frontend → e2e. Las capas se
reparten `[db]`/`[backend]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` → `devops`.

1. `[db]` Migración `000007` (tablas `agenda_*` + catálogo de categorías sembrado) con su `down` y
   prueba up/down/up.
2. `[db]` Migración `000008` (`admin_actions_action_check` +15 códigos, **incluidos los 3
   `category.*`**) con su `down` y prueba.
3. `[db]` Migración `000009` (sembrado del permiso **`categorias`**: `INSERT` en `permissions` de
   F2; `down` que retira sus asignaciones y la fila) con prueba up/down/up. **Nunca** se edita
   `000002` (migración ya aplicada).
4. `[backend]` **Excepción aprobada (RG4-10)**: actualización deliberada **9 → 10** de las pruebas
   de F2 que fijan el catálogo (las 3 señaladas por el humano + `service_roles_integration_test.go:69`
   y `service_roles_test.go:464` detectadas en revisión, fixture `sampleCatalog()` y comentarios),
   registrada como excepción en la entrega. *Depende de 3.*
5. `[backend]` Fusión del delta en `backend/api/openapi.yaml` (`info.version` → 0.5.0; prosa de
   permisos 9 → 10 y de `media`) + `make api-gen`.
6. `[backend]` `platform/audit`: 15 códigos + pruebas del registro cerrado.
7. `[backend]` `config`: `APP_TZ` (validación de zona IANA; valor `America/Costa_Rica`) +
   `.env.example`.
8. `[backend]` Consultas `agenda.sql` + `media.sql` (`make sqlc`) y repository de `agenda` con
   pruebas de integración (incluida la regla "próximo", la atomicidad con auditoría y las
   migraciones `000008`/`000009` up/down/up).
9. `[backend]` Service público de `agenda` (lo que viene, idioma, orden) + pruebas.
10. `[backend]` Service de panel (CRUD, validaciones —`endTime` exige `startTime`—, duplicados,
    publicación, `409` de categoría en uso, auditoría) + pruebas.
11. `[backend]` Handlers y `routes.go` (tres grupos: `AdminChain("eventos"|"actividades"|
    "categorias")`) + pruebas + humo en `main_test.go`.
12. `[backend]` Dominio `media` (promoción de la descarga + `IsFilePublished`: vencido no servido,
    vuelve con fecha próxima) + pruebas.
13. `[frontend]` Cliente `api/schema.d.ts` + `api/agenda.ts` + hooks (conforme a `ux.md`).
14. `[frontend]` Página pública **`/eventos`** (`EventsPage` + dos bloques + fechas localizadas +
    i18n) y `HiddenNotice` (**aviso de sección vacía sin redirigir**).
15. `[frontend]` `useSiteLinks` con href `/eventos` + entrada condicional en barra, menú móvil y pie.
16. `[frontend]` Panel: `EventsPage`, `ActivitiesPage`, `CategoriesPage` y rutas
    `/panel/{eventos,actividades,categorias}` con `RequirePermission` (EVENTOS/ACTIVIDADES/
    **CATEGORIAS**); `lib/permissions.ts` y `features/roles/permissions.ts` (+ tests: el sello
    «Disponible más adelante» se retira de los tres).
17. `[frontend]` Promoción de `ImageUploader`/`PublishControls` al inventario único.
18. `[…]` e2e `agenda-publica.spec.ts` (incluye `/eventos` con aviso sin redirigir) y
    `agenda-panel.spec.ts` (incluye el cruce con el permiso `categorias`) + checklist de quickstart.
19. `[…]` Validación (qa/revisor/seguridad), documentación (documentador) y PR.

## Puntos que necesitaban aprobación humana — **RESUELTOS en la puerta 2 (2026-10-11)**

Los 7 puntos quedaron cerrados por el humano; se implementan tal cual y **no se reabren**:

1. **Permisos del catálogo de categorías** (R4-4) → **RESUELTO**: permiso nuevo **`categorias`**
   (10.º del catálogo de F2, migración `000009`); el catálogo no se gestiona con `eventos` ni con
   `actividades`. Implica la actualización deliberada de pruebas de F2 (RG4-10, excepción aprobada).
2. **Valor de `APP_TZ`** (R4-2/RG4-1) → **RESUELTO**: **`America/Costa_Rica`** (valor firme
   documentado en `.env.example` y quickstart).
3. **`endTime` exige `startTime`** (R4-8) → **RESUELTO**: decisión firme (un fin sin inicio se
   rechaza con `details.endTime`).
4. **Borrado de categoría en uso** (R4-5) → **RESUELTO**: **`409 conflict`** (no se desliga
   automáticamente de los elementos).
5. **Imagen de un elemento publicado pero vencido** (P4-7) → **RESUELTO**: **deja de servirse** (no
   se expone) y vuelve a servirse cuando el elemento tiene de nuevo una fecha próxima.
6. **Página pública sin contenido** (R4-13) → **RESUELTO**: se muestra la página con **aviso**
   («Por ahora no hay eventos ni actividades») **sin redirigir**; el ítem sigue oculto en barra,
   menú móvil y mapa del sitio.
7. **Naming** (R4-14/Complexity 7) → **RESUELTO**: ruta pública **`/eventos`** (fijada por el
   humano) y panel `/panel/{eventos,actividades,categorias}`; nombres de pieza de UI según `ux.md`;
   dominios `internal/agenda/`/`internal/media/` y tablas `agenda_*` según la pauta de F3.

**Pendiente menor (no bloquea el plan)**: dos detalles de interfaz en los que `ux.md` y la puerta 2
dicen cosas ligeramente distintos, a alinear en las tareas `[frontend]` (sin efecto en la API ni en
los datos): (a) la **copia literal del aviso** (humano: «Por ahora no hay eventos ni actividades» ·
`ux.md` D-13: «Ahora mismo no hay nada publicado en esta sección. Vuelve pronto.»); (b) el **modal
de borrado de categoría** de `ux.md` D-16, que describe desligar la categoría de los elementos — la
decisión firme es `409`, así que la UI debe indicar "quítala primero".
