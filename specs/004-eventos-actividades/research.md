# Research — F4 Eventos y actividades

**Fecha**: 2026-10-11 · **Spec**: [spec.md](./spec.md) (aprobada por el humano; Q1–Q6 resueltas el
2026-10-11) · **Plan**: [plan.md](./plan.md)

> **Actualizado el 2026-10-11 con las decisiones firmes de la puerta 2** (humano):
> `APP_TZ = America/Costa_Rica` (R4-2), ruta pública **`/eventos`** (R4-7/R4-14), permiso nuevo
> **`categorias`** para el catálogo (R4-4, cambio de alcance), página vacía con **aviso sin
> redirigir** (R4-13) y defaults convertidos en decisión firme (`endTime` exige `startTime`,
> borrado de categoría en uso → `409`, imagen de elemento vencido no servida).

Este documento resuelve las decisiones técnicas abiertas de F4 con alternativas y motivos. La
numeración es **R4-1…R4-15**, propia de esta funcionalidad, para no colisionar con R3-1…R3-18 de F3,
R1–R23 de F2 ni D1–D23/D-A1…D-A9 de `docs/tecnico/decisiones.md`. Donde se reutiliza una decisión
ya tomada se cita (R3-x, P3-x) y **no se reabre**.

Contexto reutilizado sin cambios: capas `handler → service → repository` y reglas R1–R8 de
`docs/tecnico/arquitectura.md`; sobre de éxito = DTO directo / `{items}` en listados (§8.1.2);
paginación `platform/paginate` (§8.1.3); `ORDER BY` determinista (§8.1.7); validación
`platform/validate`; autorización `middleware.AdminChain`/`AuthzByModule`; auditoría en
`admin_actions` con `InsertAdminAction` transaccional (R3-11); imágenes en disco local con
`platform/storage` (R3-8); bilingüismo por columnas `*_es`/`*_en` con resolución `?lang=` en el
servidor (R3-1/R3-2); `publication_state` por elemento con consultas `…Published` (R3-3).

---

## R4-1. Modelos de datos separados: evento con fechas hijas · actividad con fecha única

**Pregunta.** Decisión 3 y Q1/Q2: el evento es una lista explícita de fechas concretas + texto
opcional de frecuencia; la actividad es **una** fecha con horas opcionales. ¿Cómo se modelan?

**Decisión.** Cuatro tablas (ver [data-model.md](./data-model.md)):

- `agenda_events` (padre del evento: título, frecuencia, detalle, lugar, mapa, imagen, organizador,
  categoría, estado) + **`agenda_event_dates`** (hija: `event_date DATE` + `start_time`/`end_time`
  opcionales por fecha, `UNIQUE (event_id, event_date)`).
- `agenda_activities` (actividad: **todo en una fila**, con `activity_date DATE` +
  `start_time`/`end_time` opcionales dentro del mismo día).
- `agenda_categories` (catálogo de categorías, compartido por ambos conceptos — R4-5).

La periodicidad **no se calcula nunca** (Q1): lo que hay es la lista cargada y el texto de
frecuencia (`frequency_es`/`frequency_en`). La vista pública muestra la **próxima** fecha cargada
(FR-004).

**Alternativas descartadas.**

| Alternativa | Por qué no |
|---|---|
| Una sola tabla `agenda_items` con `kind IN ('event','activity')` | Contradice literalmente la Decisión 3 ("modelos distintos") y obligaría a `CHECK` condicionales (una actividad "sin fechas hijas", un evento "con ≥1 fecha"), a permisos cruzados en un solo endpoint y a que el error de "sin ninguna fecha" se resolviera con dos reglas en una tabla |
| Actividad también con tabla de fechas | Una tabla hija para exactamente una fecha: joins y validaciones extra sin ninguna ganancia; la periodicidad propia del evento queda mejor en su tabla |
| Fechas del evento como `date[]`, JSON o `TEXT` con separadores | Pierden el `CHECK` por fecha (horas coherentes), el `UNIQUE` por fecha, los índices y la consulta sqlc plana; el mapeo `pgtype` se complica sin ganancia |
| Modelo de recurrencia (RRULE, `frequency_code` + excepciones) | La spec lo prohíbe explícitamente (Q1: "sin cálculo automático de recurrencia"; Out of Scope) |

---

## R4-2. "Lo que viene" y retiro automático: **filtro de lectura**, sin cron ni cambios de estado

**Pregunta.** FR-002/FR-006 (Q3/Q4): el público solo ve lo de hoy en adelante, ordenado
cronológicamente, y un elemento cuya fecha pasa **se retira solo** conservando su estado publicado
en el panel. ¿Cómo se implementa?

**Decisión.** El retiro es un **efecto del filtro de lectura**, no un proceso:

1. **Ocurrencia "próxima"** (regla exacta, aplicada en SQL con parámetros):

   ```sql
   (d.event_date > $today) OR (d.event_date = $today AND (d.end_time IS NULL OR d.end_time >= $now_time))
   ```

   - `event_date > hoy` → siempre próxima.
   - `event_date = hoy` y **sin hora de fin** → se muestra todo el día y se retira al terminar el día.
   - `event_date = hoy` **con hora de fin** → se retira al terminar su hora (edge case "se retira
     automáticamente al terminar su fecha/hora").
2. **Evento**: aparece en público **si tiene ≥ 1 ocurrencia próxima**; `dates` del DTO público son
   solo las próximas (la primera es "la próxima fecha" de FR-004) y el orden del bloque es por esa
   primera fecha próxima. **Actividad**: aparece si su única fecha cumple la regla.
3. **El estado no cambia**: `publication_state` sigue en `published`; el panel muestra el elemento
   con `isPast`/`nextDate: null` y el aviso "falta cargar la próxima fecha" en eventos (default
   revisable de la spec). Al cargar una fecha futura, el elemento vuelve a aparecer solo.
4. **`$today` / `$now_time` los calcula el servidor** en la zona horaria configurada
   (**`APP_TZ`**, variable de entorno nueva documentada en `.env.example` **con el valor firme
   `America/Costa_Rica`** — decisión del humano en la puerta 2, 2026-10-11) y viajan como
   **parámetros** de la consulta (nunca SQL concatenado). No se usa `CURRENT_DATE`/`now()` de
   PostgreSQL directamente porque el contenedor corre en UTC y "hoy" debe ser el día **de la
   iglesia** (un evento de esta noche no puede retirarse 3 h antes porque en UTC ya es mañana).

Sin cron, sin jobs, sin marcas de "realizado": el sitio muestra "lo que viene" con una sola regla
verificable en las consultas (`…Upcoming`, junto a las `…Published`).

**Alternativas descartadas.**

| Alternativa | Por qué no |
|---|---|
| Job/cron que pase a `past`/borrador los elementos vencidos | Contradice Q4 ("conserva su estado publicado… no se borra ni se convierte en borrador"); añade un proceso con su propio fallo y su propia auditoría para algo que la lectura ya resuelve |
| `event_date >= CURRENT_DATE` de la BD | `CURRENT_DATE` es el día del **servidor** (UTC en Compose): el corte se desplaza horas respecto al día real de la iglesia |
| Resolver "hoy" en el navegador | Cada visitante tendría una vista distinta (el de Tokio ya no vería el culto de hoy) y el filtro dejaría de estar en el servidor, donde vive el resto de la autoridad (R3-3) |
| Marca `realizado` visible para el visitante | Descartada por Q4/Q3: lo pasado no se muestra en el sitio público (su recuento es F8) |

**Valor firme (puerta 2, 2026-10-11)**: `APP_TZ = America/Costa_Rica` (la zona horaria del lugar).
El código no asume zonas: lee `APP_TZ` y **falla con error claro si no es una zona IANA válida**
(fail-closed: mejor un fallo claro al arrancar que un corte silencioso). El valor por defecto documentado en
`.env.example` y en el quickstart es `America/Costa_Rica`.

---

## R4-3. Bilingüismo y formato de fechas/horas

**Pregunta.** FR-007/FR-008/FR-009: contenido es obligatorio / en opcional con fallback, y las
fechas y horarios deben presentarse "de forma clara y comprensible en el idioma elegido… sin
cambiar el día ni la hora reales".

**Decisión.** Sin cambios sobre F3:

- **Guardado**: patrón R3-1 literal (`*_es NOT NULL`, `*_en NULL`, `""`/espacios → `NULL`).
  Traducibles: `title`, `frequency`, `detail`, `place`, `organizer`, `image_alt` y `name` de la
  categoría. **No** traducibles: `map_url`, `image_file`, fechas y horas (R3-5: el dato es el dato).
- **Lectura pública**: `GET /api/v1/proximamente?lang=es|en` resuelve cada campo en el servidor con
  la regla `en si existe, si no es` (R3-2): nunca un campo vacío por falta de traducción (SC-006).
  El cliente público **no** aplica fallback (analyze I2 de F3).
- **Presentación de fechas y horas** (es responsabilidad del cliente, igual que el formato
  a.m./p.m. del horario de F3): el contrato lleva `date` como `YYYY-MM-DD` y `startTime`/`endTime`
  como `HH:MM` 24 h; el frontend los presenta con `Intl.DateTimeFormat` según el **idioma activo**
  del selector ("viernes 17 de octubre, 7:00 p. m." / "Friday, October 17, 7:00 PM"), envueltos en
  `<time datetime="…">`. Se reutiliza `formatTimeRange` (`features/publico/schedule.ts`) para las
  horas y se añade un `formatLongDate(lang)` a juego. El valor nunca se recalcula de zona horaria:
  es la fecha y la hora **del lugar** (FR-009: "sin cambiar el día ni la hora reales").

**Alternativas descartadas.** Guardar la fecha como `TIMESTAMPTZ` (obligaría a conversiones de huso
para mostrar "el 17 de octubre a las 7" a un visitante de otra zona: exactamente lo que FR-009
prohíbe); traducir la fecha en el servidor (el idioma de la fecha es de presentación y el servidor
ya devuelve strings resueltos para los **textos**); formatos numéricos `17/10/2026` (FR-024 pide
"fecha escrita, no solo numérica").

---

## R4-4. Permisos: `eventos` y `actividades` se activan; el catálogo de categorías exige el permiso nuevo `categorias`

**Pregunta.** FR-012: gestión de eventos exige `eventos` y gestión de actividades exige
`actividades`, ambos ya sembrados por F2 (`000002_create_roles_and_permissions`, verificado en el
código: `('eventos','Eventos')`, `('actividades','Actividades')`). ¿Y el catálogo de categorías
(Q6), que es **compartido** por ambos módulos?

**Decisión (actualizada por el humano en la puerta 2, 2026-10-11 — cambio de alcance
autorizado)**: la gestión del catálogo de categorías exige un **permiso propio `categorias`**, el
**10.º del catálogo de F2**. Este permiso **sí se crea** (ampliación aprobada del catálogo fijo de
F2, que hasta ahora decía "sin crear permisos nuevos"): se siembra con una migración nueva que
hace `INSERT` en la tabla `permissions` (ver [data-model.md](./data-model.md), migración
`000009_seed_permission_categorias`; `000002` está aplicada y **nunca se edita**).

- **Eventos**: subgrupo `/api/v1/admin/eventos` con `middleware.AdminChain("eventos", deps)`.
- **Actividades**: subgrupo `/api/v1/admin/actividades` con `middleware.AdminChain("actividades",
  deps)`. Son **independientes** (Decisión 5): tener uno no da acceso al otro; la denegación es
  idéntica a la de F2 (`403 forbidden` genérico, registrada con `result='denied'`).
- **Categorías**: subgrupo `/api/v1/admin/categorias` con `middleware.AdminChain("categorias",
  deps)`. **No** se autoriza con `eventos` ni con `actividades`: son permisos independientes entre
  sí (tener `eventos` no da el catálogo y al revés). Como `AdminChain`/`AuthzByModule` ya existen y
  solo cambia el código de permiso, **no hace falta ningún helper nuevo de middleware** (se
  descarta la variante `AuthzByAnyModule`/`AdminChainAny` que este research proponía antes de la
  puerta 2).
- **Arranque**: el administrador inicial recibe `categorias` **automáticamente** — el init de F2
  (`Initialize`) lee el catálogo dinámicamente y crea el rol «Administrador» con **todo** el
  catálogo (hoy 10 permisos); sin cambios en F2 para ello.
- **Formulario de roles de F2**: lista el catálogo con `GET /api/v1/admin/permisos` (dinámico), así
  que **ofrece solo** el permiso nuevo; solo hay que quitarle el sello «Disponible más adelante»
  (`isPermissionAvailable` de `features/roles/permissions.ts`, mismo gesto que F3 hizo con
  `portada` y que F4 hace con `eventos`/`actividades`).
- Frontend: `EVENTOS`/`ACTIVIDADES`/`CATEGORIAS` en `lib/permissions.ts`; guards
  `RequirePermission` por módulo (`/panel/categorias` con `RequirePermission(CATEGORIAS)`); el
  resto de F2 no se toca (lee el catálogo dinámicamente).

**Impacto deliberado en F2 (funcionalidad cerrada)**: tres pruebas de F2 fijan el catálogo en
**9** y deben actualizarse a **10** como parte de F4 (cambio de producto autorizado, **no** un
debilitamiento): `repository_roles_integration_test.go:52`, `service_init_integration_test.go:57` y
`service_init_test.go:199`. La revisión de código de este plan detecta **dos aserciones más** que
también fijan el catálogo en 9 y deben moverse junto (`service_roles_integration_test.go:69`, que
lee el catálogo real de la BD, y `service_roles_test.go:464`, que cuenta la fixture
`sampleCatalog()` — la fixture debe crecer a 10 para que `service_init_test.go:199` tenga sentido),
más comentarios de texto ("los 9 permisos" en `service_init.go:66` y en los propios tests). La
actualización la hace `dev-backend` dentro de F4 y queda **registrada como excepción aprobada**;
el resto de F2 (servicios, handlers, UI) lee el catálogo dinámicamente y **no** cambia. Riesgo
detalle en RG4-10 de [plan.md](./plan.md).

**Alternativas descartadas.** Catálogo gestionable con **cualquiera** de `eventos`/`actividades`
(el default de este documento antes de la puerta 2; el humano lo descartó: el catálogo es un módulo
propio y su acceso no debe derivarse de otros); reservarlo solo a `eventos` (arbitrario: quien solo
gestiona actividades quedaría sin poder etiquetar); dos catálogos, uno por módulo (el visitante
vería dos vocabularios para la misma idea y la spec pide **un** catálogo); comprobar permisos solo
en el cliente (CWE-862); reutilizar `AuthzByAnyModule` (ya no hace falta: sobra complejidad con un
permiso propio).

---

## R4-5. Categorías: catálogo compartido, sembrado por la migración y administrable

**Pregunta.** Q6: categoría opcional con catálogo administrable (acción social, predicación,
adoración, entrevista, noche de oración, otro). ¿Cómo se administra y qué pasa al borrar una
categoría en uso?

**Decisión.** Tabla `agenda_categories` (`name_es`/`name_en`, `sort_order`) con el **catálogo
inicial sembrado en la migración `000007`** (mismo gesto que F2 con `permissions`), incluidas las
traducciones al inglés (revisables desde el panel; si se vacían, el público muestra el español).
CRUD completo desde el panel (`/api/v1/admin/categorias`):

- nombre duplicado (mismo `name_es` normalizado) → `409 conflict` (`UNIQUE (name_es)`).
- **borrado restringido en uso** (decisión **firme**, puerta 2 del 2026-10-11): la FK de
  `agenda_events.category_id` / `agenda_activities.category_id` es `ON DELETE RESTRICT`; el service
  traduce el error a `409 conflict` ("está en uso por N elementos"). Renombrar sí está permitido
  (el cambio se hace visible en lo publicado de inmediato: es un dato del elemento, como cualquier
  otro).
- La categoría es **opcional** en el elemento (un elemento sin categoría es válido, Q6) y su
  nombre es traducible (es obligatorio, en opcional).

**Alternativas descartadas.** `ON DELETE SET NULL` (borrar una categoría quitaría silenciosamente
la etiqueta de los elementos publicados: un cambio invisible del contenido público); catálogo fijo
en un `CHECK` de la BD como las redes de F3 (Q6 pide **administrable**, no fijo); etiquetas libres
por elemento (descartada por Q6); un catálogo por módulo (R4-4).

---

## R4-6. Imágenes: se reutiliza `platform/storage` y la descarga sube al dominio `internal/media/`

**Pregunta.** FR-020: imagen opcional por elemento con texto alternativo. F3 ya tiene subida
multipart, disco local con volumen `uploads_data` y descarga pública
`GET /api/v1/media/{fileName}` con la regla "solo archivos referenciados por contenido publicado"
(`IsHomeFilePublished`, implementada **dentro** del dominio `portada`). ¿Cómo entra F4 sin romper
esa regla ni duplicar el servicio de archivos?

**Decisión.** Dos piezas:

1. **Subida**: se reutiliza tal cual `platform/storage` (`LocalStore`, firma binaria, sin SVG, 8 MB,
   nombre generado `img_<uuid>.<ext>`) con **un endpoint de subida por módulo** —
   `POST /api/v1/admin/eventos/imagenes` y `POST /api/v1/admin/actividades/imagenes` — cada uno
   tras **su** permiso (una cuenta solo de actividades no sube al área de eventos). Se auditan con
   `event.image.upload` / `activity.image.upload` (fail-closed como en F3, analyze I8).
2. **Descarga**: el servicio de archivos deja de pertenecer a la portada y se promueve a un dominio
   propio **`internal/media/`** (misma ruta `GET /api/v1/media/{fileName}`, mismas cabeceras,
   mismos errores): la consulta `IsFilePublished` (en `internal/db/queries/media.sql`, SQL
   compartido vía consulta nombrada — R2) responde `TRUE` solo si el nombre está referenciado por
   contenido **publicado y visible hoy**:

   - portada (F3): identidad `published` (sin fecha: como hasta ahora);
   - **eventos/actividades (F4)**: elemento `published` **y próximo** según la regla de R4-2
     (decisión **firme**, puerta 2 del 2026-10-11: un elemento retirado de la vista pública por
     fecha pasada **deja de servir** su imagen sin borrarse — no se expone —; al volver a tener una
     fecha próxima, la misma URL vuelve a responder `200` — exactamente el comportamiento que F3 ya
     definió para el borrador, `analyze` C4).

   `Cache-Control: no-store` se mantiene (RG3-8: la descarga depende del estado actual).

**Por qué promover el dominio**: con F4 la descarga ya no es un asunto de la portada; F8 (galería) y
F9 (medios) ampliarán la misma consulta. Dejar `handler_media.go` en `portada` obligaría a que el
dominio de la portada conociera las tablas de todos los módulos (y cada módulo nuevo volvería a
tocarlo). El cambio es de **ubicación**, no de comportamiento: el handler y sus pruebas se mueven
sin reescribirse.

**Alternativas descartadas.** Dejar la descarga en `portada` ampliando `IsHomeFilePublished`
(acoplaría F3 a F4–F9, contra el espíritu de R2); servir **todas** las imágenes subidas sin mirar el
estado (rompe FR-013 "por ninguna vía": una imagen de un borrador sería descubrible); `bytea` en la
BD o S3 (R3-8 ya las descartó); un endpoint de descarga por módulo (varias rutas para el mismo
archivo y el HTML tendría que saber de qué módulo es cada imagen).

---

## R4-7. Superficie HTTP y caché

**Pregunta.** ¿Qué rutas públicas y de panel, y qué política de caché?

**Decisión.** Convención de F3 (R3-13), segmentos en español donde el dominio lo pide (§5.1):

| Superficie | Rutas | Protección |
|---|---|---|
| Pública (visitante) | `GET /api/v1/proximamente?lang=es\|en` | ninguna (solo lectura); **`Cache-Control: no-store`** |
| Pública (archivos) | `GET /api/v1/media/{fileName}` | ninguna; solo archivos de contenido publicado y próximo (R4-6); `no-store` |
| Panel — eventos | `GET/POST /api/v1/admin/eventos`, `GET/PATCH/DELETE /api/v1/admin/eventos/{id}`, `POST /api/v1/admin/eventos/imagenes` | `AdminChain("eventos")` (sesión + guard + CSRF + permiso `eventos`) |
| Panel — actividades | ídem bajo `/api/v1/admin/actividades` | `AdminChain("actividades")` |
| Panel — categorías | `GET/POST /api/v1/admin/categorias`, `PATCH/DELETE /api/v1/admin/categorias/{id}` | `AdminChain("categorias")` (permiso nuevo `categorias`, 10.º del catálogo — R4-4) |

**Ruta del SPA (decisión firme de la puerta 2)**: la página pública vive en **`/eventos`** (la
propuesta de UX, fijada por el humano; no `/eventos-actividades`). El endpoint de la API **no**
cambia de nombre: sigue siendo `GET /api/v1/proximamente` ("lo que viene"), que describe el
contenido, no la sección.

**Un solo documento público** (`GET /api/v1/proximamente`), en el patrón de `GET /api/v1/portada`:

- devuelve `{lang, events?, activities?}`: los dos bloques de la página (Q5/FR-021), **cada clave
  omitida si está vacía** (FR-013/SC-012, igual que las secciones de F3) y, si están presentes,
  sus arrays nunca son `null` (§8.1.2);
- con un solo request el frontend resuelve a la vez la página **y** la visibilidad de la entrada de
  navegación (FR-022: sin contenido próximo, la entrada «Eventos y actividades» no existe en la
  barra, el menú hamburguesa ni el mapa del sitio) — `useSiteLinks` (ya preparado para rutas
  propias, según su docstring) consume este documento con TanStack Query;
- sin paginación ni filtros de usuario (FR-003: el visitante no debe hacer nada para ver lo
  inmediato). Los "filtros públicos" son los del servidor: **publicado** y **próximo** (R4-2).

`no-store` en la respuesta garantiza SC-003 (un cambio guardado se ve en la primera carga
posterior). El registro de `error.code` **no crece**: `invalid`, `unauthenticated`, `forbidden`,
`not_found`, `conflict`, `internal`, todos ya en el enum.

**Alternativas descartadas.** Dos endpoints públicos (`/api/v1/eventos`, `/api/v1/actividades`)
obligarían a **tres** consultas (una por bloque + una de presencia para la navegación) o a que la
navegación se desincronizara del cuerpo; `GET /api/v1/eventos-actividades` (nombre de sección, no de
contenido: el documento reúne "lo que viene"); `ETag`/`immutable` (la respuesta depende de la fecha
actual y del estado: R3-13/RG3-8); filtros públicos por categoría o búsqueda (fuera del MVP: la
spec los descarta como default revisable y se añadirían como extensión aditiva menor si `ux.md` los
pide).

---

## R4-8. Validación de entrada: reutilizar `platform/validate` + reglas de dominio

**Pregunta.** FR-015: obligatorios, formato de fechas y horas coherentes, enlace de mapa válido,
español siempre presente y límites de extensión. ¿Dónde vive cada regla?

**Decisión.** Toda validación en el **backend** (§IV, CWE-20), aunque el formulario también valide:

- **Tags existentes** de `platform/validate` (sin dependencias nuevas): `required`, `omitempty`,
  `max`, `oneof` y el tag **`url`** que F3 ya añadió (https + host, ≤500) para `mapUrl`.
- **Reglas de dominio** en `service` (son de negocio, no de formato):
  - `titleEs`, `detailEs`, `placeEs` obligatorios; `*_En` opcional; `""`/espacios en `*_En` →
    `NULL` (normalización I6 de F3, literal).
  - **Fechas** (`YYYY-MM-DD`): fecha de calendario real (rechaza `2026-13-32`); **horas** `HH:MM`
    24 h; `endTime > startTime` cuando ambos existen; **`endTime` exige `startTime`** (decisión
    **firme**, puerta 2 del 2026-10-11: un fin sin inicio no es un rango; la spec solo describe
    "inicio y fin opcionales" como par del día — Q2).
  - **Evento**: ≥ 1 fecha y ≤ 50 por evento (constante revisable; acota el agregado); fecha
    repetida dentro del mismo evento → `400 invalid` (`UNIQUE (event_id, event_date)` lo refuerza).
    **Las fechas pasadas están permitidas** al crear/editar (edge case de la spec) pero el elemento
    no aparece en público sin fecha próxima.
  - **Imagen**: `imageFile` debe cumplir el patrón de nombre generado; `imageAltEs` **obligatorio
    si hay imagen** (FR-020, `CHECK` en la BD también); quitar la imagen = `imageFile: null`.
  - Límites (constantes del service, revisables): título 160, detalle **2.000** (FR-018/FR-019),
    lugar 200, URL de mapa 500, organizador 160, frecuencia 160, texto alternativo 300, nombre de
    categoría 120.
- Ante datos inválidos: `400 invalid` con `details` por campo y **sin guardado parcial** (la
  mutación completa —incluida la lista de fechas— es una sola transacción).

**Alternativas descartadas.** Validar solo en el formulario (§IV lo prohíbe); librería de fechas o
de URLs (la stdlib + los tags bastan); mover los límites solo a los `CHECK` de la BD (mensajes
peores para el humano); rechazar fechas pasadas (el edge case las permite explícitamente).

---

## R4-9. Duplicados: solo **dentro** del mismo módulo → `409 conflict`

**Pregunta.** Edge case: "un duplicado exacto dentro del mismo módulo (mismo título, misma fecha,
mismo lugar) se rechaza o se señala… un evento y una actividad con datos idénticos no son
duplicados".

**Decisión.** Se **rechaza** con `409 conflict` (coherente con F2/F3), comparando valores
normalizados (`btrim` + minúsculas) y solo dentro del módulo:

- **Actividad**: misma `titleEs` + misma `activityDate` + mismo `placeEs` que otra actividad.
  La BD lo refuerza con `UNIQUE (title_es, activity_date, place_es)` sobre los valores ya
  normalizados (el service responde antes, con mensaje claro).
- **Evento**: mismo `titleEs` + mismo `placeEs` que otro evento **y al menos una fecha en común**.
  No hay índice único simple (las fechas viven en otra tabla): la comprobación es una consulta del
  repository (`ExistsEventDuplicate`) en la misma transacción del guardado.
- **Evento vs actividad**: nunca son duplicados entre sí (Decisión 3), aunque coincidan todo.
- Categoría: mismo `nameEs` → `409` (`UNIQUE (name_es)`).

**Alternativas descartadas.** Duplicado = solo mismo título (más estricto que la spec: dos
repertorios de comida en fechas distintos serían legítimos); no detectar duplicados (el edge case
pide rechazarlos o señalarlos); señalizar solo en el formulario sin rechazar (el "se rechaza" va
primero en la spec y además protege la API forzada).

---

## R4-10. Auditoría: 15 códigos nuevos sobre el registro de F2, transaccional y cerrado

**Pregunta.** FR-017: crear, editar, publicar/retirar y eliminar —incluidos los intentos
fallidos— deben quedar en la auditoría de F2 (solo lectura).

**Decisión.** Se repite literalmente el patrón de F3 (R3-11), que ya está probado en el código:

1. **SQL compartido**: el repository de `agenda` llama `InsertAdminAction` (`internal/db`, F2)
   **dentro de la transacción** de cada mutación (fail-closed: si el registro falla, la operación
   **no se aplica** — edge case "no aplicarse sin registro").
2. **Registro cerrado**: `platform/audit` crece con **15 códigos** y la migración `000008` extiende
   `admin_actions_action_check` (8 de F2 + 15 de F3 + 15 de F4). El `target_kind` **no cambia**: los
   elementos de F4 son `content` (como los de F3: sin FK, `target_label` identifica el elemento).
3. **Regla exacta de qué código corresponde** (idéntica a la de F3, analyze M5):

   | Situación | Código |
   |---|---|
   | Alta (nazca en `draft` o `published`) | `event.create` / `activity.create` / `category.create` |
   | Edición de datos sin tocar el estado | `event.update` / `activity.update` / `category.update` |
   | Solo cambio de estado | `event.publish`/`event.unpublish`, `activity.publish`/`activity.unpublish` |
   | Datos **y** estado en el mismo guardado | **dos filas** (`*.update` + `*.publish`/`*.unpublish`) |
   | Borrado | `event.delete` / `activity.delete` / `category.delete` |
   | Subida de imagen | `event.image.upload` / `activity.image.upload` (`target_label` = `Eventos · Imagen · <fileName>`) |

   `target_label`: `Eventos · <título>`, `Actividades · <título>`, `Categorías · <nombre>` (máx. 254,
   el `CHECK` de F2 ya existe).
4. **Éxito** en la transacción de la mutación; **`failure`** (DTO inválido, duplicado) y **`denied`**
   (sin permiso) best-effort, sin cambiar la respuesta (mismo criterio que P20 de F2). La resolución
   método+ruta → acción vive en el dominio `agenda` (implementa `audit.Recorder` con su tabla de
   rutas, igual que `portada`): `usuarios` no conoce las rutas de F4.
5. El registro sigue siendo de **solo lectura**: ninguna consulta `UPDATE`/`DELETE` sobre
   `admin_actions` (FR-025 de F2).

**Alternativas descartadas.** Un registro de auditoría nuevo para F4 (FR-017 pide **el de F2**);
que `usuarios` resuelva las rutas de F4 (acoplaría F2 a cada módulo); registrar tras el commit
(rompe el edge case de atomicidad); `target_kind` nuevo por módulo (`content` ya significa
"elemento de contenido"; añadiría un `CHECK` sin información nueva).

---

## R4-11. Concurrencia: gana la última escritura completa

Igual que F3 (R3-17): cada guardado escribe el elemento entero —incluida la **lista completa de
fechas** de un evento (reemplazo, no diferencias)— en una transacción; dos guardados simultáneos se
serializan en la BD y queda el conjunto completo de un mismo formulario. Sin control optimista
(`version`/`If-Match`) ni historial (fuera de alcance; si el humano lo pide, es una columna + `409`).

---

## R4-12. Publicación: `publication_state` por elemento y "se publica al guardar"

Patrón R3-3 sin cambios: `publication_state TEXT CHECK IN ('draft','published')` en eventos y
actividades (**las categorías no lo tienen**: son etiquetas del catálogo, no contenido publicable);
consultas `…Published` separadas para el público (la única función que ve borradores es la del
panel); un `PATCH` sobre un elemento ya publicado guarda y **conserva** el estado (FR-014: los
cambios se ven desde la primera carga posterior); el paso por borrador solo existe al crear. La
sección pública entera se oculta cuando no queda nada publicado **y próximo** (FR-013/FR-022): la
entrada de navegación desaparece de barra, menú hamburguesa y mapa del sitio (SC-012), y la ruta
directa a la página muestra el **aviso de sección vacía sin redirigir** (R4-13).

---

## R4-13. La página pública sin contenido: aviso de sección vacía, **sin redirigir**

**Pregunta.** Si alguien teclea `/eventos` y no hay nada publicado y próximo, ¿qué se muestra?
(FR-013: "el sitio NUNCA DEBE mostrar secciones vacías").

**Decisión (firme, puerta 2 del 2026-10-11; sustituye a la redirección propuesta antes)**: se
muestra **la página con un aviso de sección vacía** — mensaje tipo «Por ahora no hay eventos ni
actividades» (copia literal final según `ux.md` D-13, `HiddenNotice`: H1 + aviso estándar de
interfaz, sin datos)— y **no se redirige** a la portada: la URL escrita a mano o guardada en
favoritos se queda donde está, con el sitio completo alrededor (nav, pie, selector de idioma). La
navegación **sigue sin ofrecer el enlace** (R4-7/R4-12: el ítem está oculto en la barra, el menú
móvil y el mapa del sitio); el aviso es la red de seguridad para la URL directa. Nunca se
renderizan borradores ni vencidos en esa página (FR-013 "por ninguna vía").

**Alternativas descartadas.** Redirigir a `/` (la propuesta anterior de este documento: el humano
la descartó — una URL que "no lleva a ninguna parte" sin explicación confunde, sobre todo a
adultos mayores); página vacía sin mensaje (contradice SC-012: 0 secciones vacías y 0 pantallas
sin explicación); `404` (un error para algo que es normal mientras el equipo carga contenido);
renderizar borradores para previsualizar (prohibido: FR-013 "por ninguna vía").

---

## R4-14. Frontend: se reutiliza el sistema de F10 y los patrones de F3; el inventario compartido crece

**Decisión.** F4 **aplica** F10 (tokens, modo oscuro, movimiento con `prefers-reduced-motion`,
`PublicLayout`, `PublicNav`, `MobileMenu`, `FooterSiteMap`, `LanguageSwitcher`) sin redefinirlo:

- **Página pública `/eventos`** (ruta firme de la puerta 2; la propuso `ux.md` D-1 y el humano la
  fijó) dentro de `PublicLayout`, con dos bloques (`Próximos eventos`, `Próximas actividades`)
  compuestos con el inventario público de F10 (`Card`, `SectionHeading`, `PublicButton`, `Reveal`).
  La composición y los nombres de pieza los fija `ux.md` §6.3 (`features/publico/eventos/`:
  `EventsPage`, `EventsSection`, `ActivitiesSection`, `EventCard`, `ActivityCard`, `CategoryChip`,
  `DateBadge`, `MapLink`, `HiddenNotice`, hook `useUpcomingEvents`, `formatoFechas.ts`). Sin nada
  publicado y próximo, `EventsPage` muestra el aviso de `HiddenNotice` (R4-13, **sin redirigir**).
- **Navegación**: `SiteLink` crece con `href` de **ruta** (el hook ya estaba preparado para ello);
  la entrada «Eventos y actividades» (href `/eventos`) existe solo si el documento público trae algún
  bloque (FR-022), y la consumen a la vez `PublicNav`, `MobileMenu` y `FooterSiteMap` (fuente única).
- **i18n**: nuevas claves tipadas en `features/publico/i18n/{messages,es,en}.ts` (nav, títulos de
  bloque, etiquetas de fecha/categoría, estados vacíos): sin la clave, TypeScript no compila
  (garantía de SC-006).
- **Panel** (en español, FR-016): nueva feature `features/eventos/` (inventario de `ux.md` §6.4:
  `EventsPage`, `ActivitiesPage`, `CategoriesPage`, `EventForm`, `ActivityForm`, `CategoryForm`,
  `DatesEditor`, `EventsList`, `ActivitiesList`, `messages.ts` y hooks TanStack Query), rutas
  `/panel/eventos` (`RequirePermission(EVENTOS)`), `/panel/actividades`
  (`RequirePermission(ACTIVIDADES)`) y `/panel/categorias` (`RequirePermission(CATEGORIAS)`: el
  permiso nuevo, R4-4 — **no** `RequireAnyPermission`, ya no hace falta) y entradas de menú
  filtradas por permiso.
- **Permisos de F2 en el frontend**: `lib/permissions.ts` crece con `EVENTOS`, `ACTIVIDADES` y
  `CATEGORIAS`; `features/roles/permissions.ts` deja de marcar los tres como «Disponible más
  adelante» (y su prueba unitaria se actualiza en consecuencia) para que el formulario de roles de
  F2 ofrezca el permiso nuevo.
- **Componentes compartidos**: los que F3 dejó en `features/informacion/` y F4 necesita también
  (`ImageUploader`, `PublishControls`) **se promueven al inventario único** `src/components/` (regla
  de F2/T242: "cualquier componente compartido con F4–F9 se propone primero al inventario"). El
  panel de F4 comparte además `Field`, `Select`, `Button`, `Notice`, `ConfirmDialog`, `Table`,
  `StatusPill`, `EmptyState`, `Tabs` (la paginación de `ux.md` D-17 se decide en `tasks.md`: el
  contrato conserva `limit`/`offset`, §8.1.3).

**Alternativas descartadas.** Copiar `ImageUploader`/`PublishControls` en la feature nueva (duplica
markup y pruebas); crear un segundo `LanguageProvider` o un layout público nuevo (F10 ya los
entregó); resolver la visibilidad de la navegación en cada componente (fuente única: `useSiteLinks`).

---

## R4-15. Dependencias nuevas: **ninguna**

Constitución §II: dependencias minimizadas y justificadas. F4 **no añade ninguna dependencia** ni
en backend (`net/http`, `net/url`, `time`, `pgx`/`sqlc`, `testcontainers-go` ya están) ni en frontend
(React 19, React Router, TanStack Query, React Hook Form + Zod, Tailwind, Vitest + Testing Library +
MSW, Playwright, `Intl` de la stdlib del navegador para las fechas localizadas).

**Descartado explícitamente.** Librería de RRULE/recurrencia (la spec la prohíbe: Q1); librería de
fechas tipo `dayjs`/`date-fns` (para presentar una fecha localizada basta `Intl.DateTimeFormat`;
el cálculo de "hoy" es del servidor); librería de calendario/selector de fechas (el `input[type=date]`
del navegador es accesible y ya se usa en F2/F3 si hace falta); `react-big-calendar` (la UX es una
lista cronológica, no un calendario — fuera de alcance).

---

## Resumen de decisiones (puerta 2: las 7 consultas quedaron **resueltas** el 2026-10-11)

| # | Decisión | Un vistazo |
|---|---|---|
| R4-1 | Modelos separados | `agenda_events` + `agenda_event_dates` (fechas hijas) · `agenda_activities` (fecha única) · `agenda_categories` (catálogo compartido) |
| R4-2 | Lo que viene / retiro automático | Filtro de lectura (SQL parametrizado con `today`/`now_time` del servidor en **`APP_TZ = America/Costa_Rica`** — valor firme); sin cron y sin cambiar estados; el panel marca `isPast` |
| R4-3 | Bilingüe y fechas | Columnas `*_es`/`*_en` (R3-1) + `?lang` resuelto en servidor (R3-2); `YYYY-MM-DD`/`HH:MM` en la API y presentación localizada con `Intl` + `formatTimeRange` |
| R4-4 | Permisos | `eventos` y `actividades` se activan (ya sembrados por F2) e independientes; **el catálogo de categorías exige el permiso nuevo `categorias`** (10.º del catálogo de F2, sembrado en la migración `000009`; cambio de alcance aprobado). Sin helpers nuevos de middleware |
| R4-5 | Categorías | Catálogo sembrado en `000007`, CRUD en el panel (permiso `categorias`), borrado restringido en uso (**`409` firme**), nombre es obligatorio / en opcional |
| R4-6 | Imágenes | `platform/storage` reutilizado; subida por módulo (`…/imagenes`); descarga promovida a `internal/media/` con `IsFilePublished` (publicado **y** próximo — **firme**: el vencido no se sirve y vuelve al haber fecha próxima) |
| R4-7 | Superficie HTTP | `GET /api/v1/proximamente?lang=` (documento de dos bloques, `no-store`); panel `/api/v1/admin/{eventos,actividades,categorias}` con `AdminChain` por módulo; **ruta del SPA `/eventos`**; sin filtros ni paginación públicos |
| R4-8 | Validación | Tags existentes + reglas de dominio (fecha real, `HH:MM`, `endTime > startTime`, **`endTime` exige `startTime` — firme**, ≥1 fecha por evento, ≤50, alt con imagen, límites) |
| R4-9 | Duplicados | `409 conflict` dentro del módulo (actividad: título+fecha+lugar; evento: título+lugar+fecha en común); evento vs actividad nunca |
| R4-10 | Auditoría | 15 códigos `event.*`/`activity.*`/`category.*` (migración `000008`, incluidos los 3 `category.*`); `target_kind='content'`; transaccional fail-closed; Recorder del dominio |
| R4-11 | Concurrencia | Última escritura completa gana (reemplazo de la lista de fechas en una transacción) |
| R4-12 | Publicación | `publication_state` por elemento (sin categorías); consultas `…Published`/`…Upcoming`; se publica al guardar |
| R4-13 | Página pública vacía | **Aviso de sección vacía en `/eventos`, sin redirigir** (firme); el ítem sigue oculto en barra, menú móvil y mapa del sitio |
| R4-14 | Frontend | F10 aplicado sin cambios; **`/eventos`** con dos bloques (`features/publico/eventos/` de `ux.md`); `SiteLink` con ruta; panel en `features/eventos/` con `RequirePermission` (EVENTOS/ACTIVIDADES/**CATEGORIAS**); `ImageUploader`/`PublishControls` al inventario |
| R4-15 | Dependencias | Ninguna nueva en backend ni frontend |

**Riesgo nuevo del cambio de alcance (RG4-10 en `plan.md`)**: F4 **edita pruebas de F2 (funcionalidad
cerrada)** — el catálogo pasa de 9 a 10 permisos — y amplía la spec de F2 (catálogo "fijo"). Es una
decisión de producto autorizada; se documenta como excepción aprobada y se limita a las aserciones
de recuento (3 fijadas por el humano + 2 detectadas en revisión), la fixture `sampleCatalog()` y
comentarios; el resto de F2 lee el catálogo dinámicamente.

**Pendiente menor (no bloquea)**: la copia literal del aviso de R4-13 (el humano: «Por ahora no hay
eventos ni actividades»; `ux.md` D-13: «Ahora mismo no hay nada publicado en esta sección. Vuelve
pronto.») y el modal de borrado de categoría de `ux.md` D-16 (describe desligar la categoría; la
decisión firme es `409` y la UI debe indicar "quítala primero") — se alinean en `tasks.md`/UI, sin
efecto sobre la API ni los datos.
