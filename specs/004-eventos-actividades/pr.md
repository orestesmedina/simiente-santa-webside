# PR — F4 Eventos y actividades

> Descripción para el Pull Request de la rama `004-eventos-actividades` (preparada por
> `documentador`, 2026-10-11; T441 — **actualizada el 2026-10-12** con las 3 correcciones de bugs,
> la enmienda de alcance Q8, su re-validación y la **tanda 2** del mismo día (fondo negro de las
> pantallas de acceso + reorden del bloque). **El merge y el despliegue requieren aprobación
> humana explícita** (fase 9/9 del flujo equipo-feature; puertas «PR / merge» y «Despliegue»
> pendientes en `estado.md`). El entorno no tiene `gh`: quien abre el PR es `devops`/el humano con
> esta descripción.

## feat: F4 — Eventos y actividades (sección, panel y bloque «Lo que viene»)

### Resumen

Primera sección de contenido dinámico del sitio público: **lo que viene**, visible sin cuenta, el
**extracto de agenda en la portada** y los tres módulos del panel con los que el equipo lo
administra. Cubre las **49 tareas (T401–T449)** de `tasks.md` (US1–US7 / FR-001…FR-027 de la
spec; Q1–Q7 resueltos el 2026-10-11 y **enmienda de alcance Q8** el 2026-10-12).

**Backend (Go + PostgreSQL, módulos `internal/agenda` e `internal/media`)**
- Documento público **`GET /api/v1/proximamente?lang=es|en`** (`Cache-Control: no-store`): dos
  bloques (`events` / `activities`) con claves **omitidas** cuando no hay contenido, orden
  cronológico por la próxima fecha, fallback `en → es` por campo y `400` para `lang` no soportado.
- **Documento público nuevo `GET /api/v1/portada/proximamente?lang=es|en`** (enmienda Q8,
  T443–T446): extracto de **hasta 5 ítems** con la consulta sqlc `ListNextUpcomingItems`
  (`LIMIT 5` en SQL, ventana `in_window = EXISTS(…)` de **2 meses de calendario** con **respaldo**
  a los siguientes próximos —las dos ramas nunca se mezclan—), `service.GetPortadaProximamente`
  (`maxPortadaItems = 5`, `portadaWindowMonths = 2`), handler público en `RegisterPublic` (sin
  sesión), `lang` cerrado a `es|en`, `Cache-Control: no-store`, clave `items` **omitida** si está
  vacía y DTO mínimo (`kind`, `id`, `title`, `date`, `startTime`, `endTime`, `place`, `image`) sin
  `publicationState` ni detalle. Misma regla R4-2 (publicado y próximo, corte en `APP_TZ`
  parametrizado al SQL) que la sección completa.
- Subgrupo de administración `/api/v1/admin/{eventos,actividades,categorias}` con CRUD,
  publicación/retirada, imágenes y catálogo de categorías; contrato vivo **0.4.0 → 0.6.0**
  (`backend/api/openapi.yaml`: 0.5.0 con la primera tanda de F4 y 0.6.0 con la enmienda Q8;
  fusión aditiva del delta `specs/…/contracts/openapi.yaml`, que no se edita) y tipos TS
  regenerados con `make api-gen`.
- Regla **«publicado y próximo»** como filtro de lectura (sin cron, sin jobs, sin cambiar
  estados), con el corte calculado en **`APP_TZ`** y pasado al SQL como parámetros (`today`/
  `now_time`), nunca concatenados (§IV/CWE-89); imagen de elemento vencido no servida y
  restituida al haber fecha próxima.
- Modelos separados (Decisión 3): `agenda_events` + `agenda_event_dates` (fechas concretas +
  texto de frecuencia) y `agenda_activities` (fecha única); validaciones completas —
  `endTime` exige `startTime` y debe ser mayor, duplicado exacto → `409`, **borrar categoría en
  uso → `409`** (FK `ON DELETE RESTRICT`), tope del catálogo por conteo (`CountCategories`).
- Migraciones **`000007`** (contenido `agenda_*`), **`000008`** (CHECK de `admin_actions` + 15
  códigos `event.*`/`activity.*`/`category.*`) y **`000009`** (siembra del permiso `categorias`),
  con `up`/`down`; auditoría atómica con la mutación, solo lectura y con etiquetas legibles
  (título/nombre actual) también en los rechazos. **La enmienda Q8 no añade migraciones.**
- **Permisos**: se activan los reservados `eventos`/`actividades` y se añade **`categorias`**
  («Categorías de agenda», 10.º del catálogo de F2), cada uno con su `AdminChain` y su
  `RequirePermission`; los tres son independientes.
- `internal/media` promovido desde `portada` (misma subida/URL y reglas que F3). **0 dependencias
  nuevas**; cobertura de `internal/agenda` **89,4 %** en la verificación de cierre del 2026-10-11 y
  **87,0 %** en la re-validación del 2026-10-12 (≥ 80 % exigido).

**Frontend (React + TypeScript + Vite)**
- **Sección pública `/eventos`** con el sistema de F10: entrada única «Eventos y actividades»
  (barra, menú móvil y pie, **oculta** sin contenido publicado y próximo), dos bloques
  «Próximos eventos» / «Próximas actividades», tarjetas con `DateBadge`/`MapLink`/`CategoryChip`,
  imágenes, **aviso de sección vacía sin redirigir** y 16 claves i18n `es`/`en` con fechas
  localizadas por idioma (`formatoFechas.ts`, `Intl` sin conversión de zona).
- **Bloque «Lo que viene» en la portada** (enmienda Q8, T447–T449): cliente `api/agenda.ts` +
  hook `usePortadaProximamente` y componente `ProximamenteBlock` con `UpcomingItem`, insertado
  **inmediatamente después del hero y antes de «quiénes somos»** (`ux.md` D-24 confirmada por el
  humano el 2026-10-12; reorden en `9aef868` — `HomePage` no se rediseña y el resto de secciones
  solo se recorre hacia abajo), **fila horizontal de tarjetas compactas con imagen** (`CardImage` + `alt`),
  etiquetas «Evento»/«Actividad», tope de 5 sin huecos ni marcadores, **ventana de 2 meses con
  respaldo**, mensaje `role="status"` **«Por el momento no hay eventos ni actividades»** con el
  botón oculto cuando no hay ítems (excepción de Q8 a SC-012), botón **«Ver más eventos»** y
  enlaces a `/eventos#eventos` / `/eventos#actividades`, i18n `es`/`en` con fallback, responsive
  (móvil en columna, sin scroll horizontal a 320 px), tema oscuro y foco visible.
- **Superficie clara fuera del panel** (`PanelSurfaceLayout`, `2a70591`): `/login`,
  `/cambiar-contrasena`, `/sin-permiso` y el 404 se envuelven en fondo claro del panel **sin depender
  del tema** y sin fondo global en `body`/`html` (evita el destello en la portada oscura); ver el
  bug 4 de la tabla siguiente.
- **Panel**: `/panel/eventos`, `/panel/actividades` y `/panel/categorias` (`EventsPage`,
  `ActivitiesPage`, `CategoriesPage`, `EventForm`, `ActivityForm`, `CategoryForm`, `DatesEditor`,
  listados con marca «Pasado», `409` de duplicado y de categoría en uso), permisos filtrados en
  menú, `AccesosRapidos` y rutas; `ImageUploader` y `PublishControls` **promovidos al inventario
  único** `src/components/`; `ACTION_LABELS` con etiqueta humana para los 15 códigos nuevos.
- **Vitest 78 archivos / 530 pruebas**, `lint`, `typecheck` y `build` en verde (74/505 antes de la
  enmienda Q8, 528 con ella: +4 archivos y +23 pruebas de alcance; +2 del `PanelSurfaceLayout`
  en la tanda 2).

**Infraestructura / Docker (devops)**
- **`APP_TZ=America/Costa_Rica`** documentada en `.env.example` (T410, leída por
  `platform/config`; valor por defecto idéntico, el service valida la zona IANA y falla con
  error claro si es inválida).
- **Sin cambios** en `docker-compose.yml`, Dockerfiles ni `.github/` (árbol limpio, verificado
  en T440).

### Bugs corregidos (equipo-bug + tanda 2 del 2026-10-12)

| # | Problema | Arreglo | Pruebas |
|---|---|---|---|
| 1 | Desde `/eventos` el menú no navegaba a las anclas de la portada (`href` relativos) y el ancla no se resolvía al llegar | enlaces **absolutos `/#ancla`** (`useSiteLinks`) + resolución del ancla al cargar la portada (`useHashScrollOnArrival`, con reintento) — `ef93c77`/`e1f6fed`/`4348875` | `qa-b1-siteLinks-relativas.test.tsx`, e2e `qa-bugs-004` caso 1, `navegacion-movil` (anclas absolutas) |
| 2 | El menú público no marcaba la sección actual ni respondía al *hover* | *hover* con par acento/azul marino **certificado AA** + `aria-current="page"` en la ruta activa — `72665e2` | `qa-b2-PublicNav-activo.test.tsx`, e2e caso 2 (repintado real) |
| 3 | El modal de «Agregar evento» era angosto y con doble scroll horizontal | `Dialog size="wide"` + `overflow-x-hidden`, `minmax(0,1fr)` en `DatesEditor`, imagen `max-w-full` — `66c8472` | e2e caso 3 (`width ≥ 640`, `scrollWidth ≤ clientWidth`), sin tocar `Dialog.test.tsx` |
| 4 | **Fondo negro** en `/login`, `/cambiar-contrasena`, `/sin-permiso` y 404 con el dispositivo en tema oscuro (texto ilegible): **hueco previo de F10**, no regresión de F4 (el anti-FOUC fija `color-scheme: dark` en `<html>` y esas rutas no declaraban fondo) | esas 4 rutas se envuelven en **`PanelSurfaceLayout`** (`bg-white` sin depender del tema) sin fondo global en `body`/`html` — `2a70591` | e2e de regresión `qa-bug-fondo-oscuro-fuera-layouts.spec.ts` con `colorScheme: 'dark'` (3 rutas, `f179daa`) + 2 unitarias nuevas en `layout.test.tsx` para `/cambiar-contrasena` (`310cd1c`) |
| — | Colateral: el enlace-título de las tarjetas del bloque nuevo no llegaba a 44 px | `min-h-11` (`e74d086`, FR-023/FR-024) | e2e `portada-accesibilidad` (mide todos los enlaces de `main`) |

**Ajuste (no bug)**: el bloque «Lo que viene» pasa a la **primera sección tras el hero, antes de
«quiénes somos»** (`9aef868`, D-24 confirmada por el humano el 2026-10-12).

### Decisiones firmes de la puerta 2 (aplicadas, no reabiertas)

- **`APP_TZ = America/Costa_Rica`**: el corte de «hoy» se calcula en esa zona en el servidor.
- **Ruta pública `/eventos`** (no `/eventos-actividades`) y panel en `/panel/eventos`,
  `/panel/actividades`, `/panel/categorias`.
- **Aviso sin redirigir**: sin nada publicado y próximo, `/eventos` muestra `HiddenNotice` y la
  entrada de navegación queda oculta.
- **Permiso nuevo «categorías»** (catálogo de F2: 9 → 10, migración `000009`; `000002` intacto).
- **`endTime` exige `startTime`**, **borrado de categoría en uso → `409`** e **imagen de elemento
  vencido no servida**.

### Cambio de alcance Q8 (2026-10-12) — bloque «Lo que viene»

Decisiones firmes del humano (versión definitiva del mismo día):

- **Hasta 5 ítems** de la agenda en la portada, **eventos y actividades juntos** y etiquetados,
  los más cercanos por fecha; **ventana preferente de los próximos 2 meses con respaldo** a los
  siguientes próximos.
- **Fila horizontal de tarjetas compactas con imagen**, título, fecha y lugar; **botón «Ver más
  eventos»** → `/eventos` en un paso (no es paginación).
- Sin nada publicado y próximo, el bloque **sí se muestra** con el mensaje «Por el momento no hay
  eventos ni actividades» **sin botón** (excepción explícita a SC-012; el resto de secciones
  sigue ocultándose vacías).
- **Único contenido de F4 en la portada**, colocado **inmediatamente después del hero y antes de
  «quiénes somos»** (`ux.md` **D-24 confirmada por el humano el 2026-10-12**; reorden `9aef868`,
  con `plan.md` P4-19 realineado). El resto de la portada no cambia; sin ítems pasados, sin
  paginación, sin más de 5 ítems y sin configuración del bloque desde el panel.

### Cómo probarlo (quickstart)

Recorrido completo con evidencia por sección en [`quickstart.md`](./quickstart.md) (§0–§15).
Resumen:

```bash
make up                 # db, redis, backend, frontend
make db-migrate         # 000007, 000008 y 000009
cd frontend && npm install
# Portada (visitante): http://localhost:5173/  -> bloque «Lo que viene»
curl -s 'http://localhost:8080/api/v1/portada/proximamente?lang=es' | jq
# Sección completa:    http://localhost:5173/eventos
curl -s 'http://localhost:8080/api/v1/proximamente?lang=es' | jq
# Panel: login con la cuenta sembrada ana@ejemplo.com / Semilla.2026 (helpers de e2e)
# -> /panel/eventos, /panel/actividades y /panel/categorias con sus permisos
```

Pruebas automatizadas: `make ci` (lint + backend unit/integración + frontend + security) y
`npx playwright test --config e2e/playwright.config.ts --workers=1` desde `frontend/`, **en cuatro
tandas** (ver abajo).

### Evidencia de pruebas

**Verificación de cierre (2026-10-11, T440)**

| Verificación | Resultado |
|---|---|
| `make ci` | **en verde**: `gofmt`/`go vet`/`golangci-lint` sin issues · backend `go test ./...` + `go test -tags=integration ./...` (PostgreSQL real) · frontend Vitest **74 archivos / 505 pruebas** · `govulncheck` **0 vulnerabilidades que afectan al código** · `npm audit --audit-level=high` **0** |
| Backend integración | **20/20 paquetes ok, 0 FAIL** con `-p 1` (la corrida sin `-p 1` puede fallar por contención de `testcontainers` en `internal/platform/session`: infra de pruebas, no de F4) |
| e2e Playwright (3 tandas, `--workers=1`, imagen Docker de Playwright) | **F4 5/5** (`agenda-publica`, `agenda-panel`, `agenda-accesibilidad`) · **F2 2/2** (`acceso`, `auditoria`, congeladas) · **F3/F10 31/31** — **38/38** en total |
| Cobertura `internal/agenda/service*.go` | **89,4 % statements** (≥ 80 % exigido); re-validada en los ciclos: 86,6 % (integración) / 91,5 % (unitarias de `service*`) |
| Deriva de generados | `sqlc-verify` y `api-gen` **sin diff** (1 commit de diff: 0) |
| Migraciones | `000007`–`000009` **up → down → up** en verde; BD dev en versión 9 |
| Docker / entorno | 4 servicios arriba · `GET /healthz` 200 · ruta inexistente `404` con `ErrorEnvelope` · `GET /api/v1/proximamente?lang=es` **200 `no-store`** con base vacía (`{"lang":"es"}`, SC-012) · `lang=fr` → `400 invalid` · catálogo de permisos = **10 con `categorias`** · **sin cambios** de F4 en compose, Dockerfiles ni `.github/` |

*Nota*: en una corrida intermedia la tanda 3 falló un caso con `toHaveURL` y salió **31/31** al
re-lanzarla dos veces; queda registrado como **flake aislado** (sin defecto de F4 identificado).

**Re-validación (2026-10-12, tras los 3 bugs y la enmienda Q8)**

| Verificación | Resultado |
|---|---|
| Backend unitarias | **20 paquetes ok** |
| Backend integración (`-p 1`) | **21/21 paquetes ok** — `agenda` con **139 subpruebas** sin ningún `SKIP`; `portada` en verde |
| Cobertura `internal/agenda` | **87,0 %** (≥ 80 %) |
| Frontend | **Vitest 78 archivos / 528 pruebas** · `lint` y `typecheck` sin errores |
| e2e Playwright (3 tandas, sin `429`) | **43/43**: tanda A F3/F10 + bloque **33/33** (incluidos los 4 de `portada-proximamente.spec.ts`), tanda B F4 + bugs **8/8** (`agenda-` + `qa-bugs-004`), tanda C F2 **2/2** |
| Bugs | cada uno con su prueba de reproducción **en verde** (unidad y/o e2e), sin debilitar ninguna prueba existente |
| Bloque (US7, FR-025…FR-027, SC-014/SC-015) | cubierto en las 3 capas (cliente API, hook, componente + service/repository sobre PostgreSQL): tope 5, ventana de 2 meses con respaldo, mensaje sin botón, botón a `/eventos`, fallback `en → es`, 320 px y teclado |
| Dependencias / secretos | `govulncheck` **0** que afecten al código · `npm audit` **0** · sin secretos nuevos |

**Re-validación tanda 2 (2026-10-12 — fondo negro + reorden del bloque)**

| Verificación | Resultado |
|---|---|
| e2e nuevo de regresión (fondo) | `qa-bug-fondo-oscuro-fuera-layouts.spec.ts` **3/3 (7,5 s)** con `colorScheme: 'dark'` y repintado real (fondo `rgb(255,255,255)` + título legible) en `/login`, `/sin-permiso` y 404 |
| e2e (4 tandas, `--workers=1`, sin `429`) | **46/46**: fondo oscuro **3/3** · tanda A F3/F10 + bloque **33/33** (incluidos los 4 de `portada-proximamente.spec.ts`) · tanda B F4 + bugs **8/8** · tanda C F2 **2/2** |
| Frontend | **Vitest 78 archivos / 530 pruebas** (+2 de `PanelSurfaceLayout`) · `lint` y `typecheck` sin errores |
| Backend | unit **20 paquetes ok** · integración **21/21** · cobertura `internal/agenda` **87,0 %** |
| Reorden | aserción de orden en `homePage.test.tsx` (`compareDocumentPosition`: hero → bloque → «quiénes somos» → horario) y tanda A íntegra; el resto de F3 conserva contenido, anclas y direcciones de entrada |
| Defectos de este alcance | **ninguno** |

### Validación del equipo

**Tanda 2 del 2026-10-12 (fondo negro + reorden) — APROBADO por los tres roles**
([consolidado](./revision-2026-10-12-tanda2.md)):

| Informe | Veredicto | Nota |
|---|---|---|
| [`revision-2026-10-12-qa-tanda2.md`](./revision-2026-10-12-qa-tanda2.md) | **APROBADO** | 46 e2e en 4 tandas, 530 unitarias frontend y suites Go (unit + integración, cobertura 87,0 %) en verde; 0 defectos |
| [`revision-2026-10-12-codigo-tanda2.md`](./revision-2026-10-12-codigo-tanda2.md) | **APROBADO** | arreglo del login correcto y **sin tocar el sitio público**; el reorden cumple D-24; 1 IMPORTANTE (documentación) + 5 sugerencias |
| [`revision-2026-10-12-seguridad-tanda2.md`](./revision-2026-10-12-seguridad-tanda2.md) | **APROBADO** | solo presentación/orden de componentes; autorización intacta; **sin secretos ni hallazgos** |

**Re-validación del 2026-10-12 — APROBADO por los tres roles, sin bloqueantes**
([consolidado](./revision-2026-10-12.md)):

| Informe | Veredicto | Nota |
|---|---|---|
| [`revision-2026-10-12-qa.md`](./revision-2026-10-12-qa.md) | **APROBADO** | 3 bugs con prueba de reproducción en verde; bloque cubre US7/FR-025–027/SC-014/015 en 3 capas; Go integración 21/21, cobertura 87,0 %, Vitest 528, e2e 43 en 3 tandas; 0 defectos de este alcance |
| [`revision-2026-10-12-codigo.md`](./revision-2026-10-12-codigo.md) | **APROBADO** | sin bloqueantes; 2 hallazgos **IMPORTANTE solo de documentación** (`plan.md` P4-19 y RG4-11 frente a `ux.md`/código) + 3 sugerencias menores |
| [`revision-2026-10-12-seguridad.md`](./revision-2026-10-12-seguridad.md) | **APROBADO** | endpoint solo publicado+próximo, SQL parametrizado, DTO mínimo, `no-store`; **0 críticos/altos/medios/bajos**; 2 informativos (uno preexistente) |

**Ciclo 2 (2026-10-11) — APROBADO por los tres roles**: el ciclo 1 halló 1 IMPORTANTE + 3
menores, todos corregidos (`1714e2b`, `3ec0b2d`, `53c7007`, `8b03db3`) y re-validados con
**0 bloqueantes** — I1 (consultas N+1 → por lote), M1 (código muerto), M2 (tope por conteo) y M3
(etiquetas de auditoría legibles).

| Informe | Veredicto | Nota |
|---|---|---|
| [`revision-2026-10-11-ciclo2.md`](./revision-2026-10-11-ciclo2.md) (consolidado) | **APROBADO** | los tres roles aprueban |
| [`revision-2026-10-11-qa-ciclo2.md`](./revision-2026-10-11-qa-ciclo2.md) | **APROBADO** | Go 20/20, Vitest **505/505**, e2e 5/5 F4 + 2/2 F2 + 31/31 F3/F10 |
| [`revision-2026-10-11-codigo-ciclo2.md`](./revision-2026-10-11-codigo-ciclo2.md) | **APROBADO** | I1/M1–M3 cerrados; sin pruebas congeladas tocadas ni `skip`/`only`/`xit` |
| [`revision-2026-10-11-seguridad-ciclo2.md`](./revision-2026-10-11-seguridad-ciclo2.md) | **APROBADO** | SQL parametrizado, filtro público intacto, sin secretos |

El ciclo 1 queda como registro en [`revision-2026-10-11.md`](./revision-2026-10-11.md).

### Excepciones aprobadas (cambio de producto, no debilitamiento)

- **RG4-10 / T405 — pruebas de F2: catálogo de permisos 9 → 10** (decisión del humano en la
  puerta 2, 2026-10-11): F4 actualiza deliberadamente las pruebas de F2 que fijaban el recuento
  (5 aserciones + fixture `sampleCatalog()` + comentarios «los 9 permisos»). **Solo cambia el
  recuento**: ninguna aserción de comportamiento se relaja y el init de F2 sigue leyendo el
  catálogo dinámico. Verificada línea a línea por `revisor-codigo` y `seguridad`.
- **`acceso.spec.ts` (e2e congelado de F2)**: el fixture creaba un rol con «Eventos» +
  «Actividades» para probar «cuenta con módulos **reservados** → solo Inicio y el aviso». Con F4
  esos dos permisos ya son módulos reales, así que el fixture pasa a **dos módulos de F5–F9**
  («Noticias y galería» y «Donaciones»): escenario y aserciones intactos. Misma naturaleza que
  RG4-10; registrado en `estado.md` (bitácora 2026-10-11). Relacionados y del mismo calibre:
  `permissions.test.ts` / `roles.test.tsx` (fixtures `eventos` → `noticias`/`donaciones`).
- **`useSiteLinks.test.tsx` (F3/F10)**: renombrado de `.ts` a `.tsx` porque el test «no inventa
  enlaces a rutas F4–F9» pasa a comprobar que las anclas de la portada no cambian (sus aserciones
  se conservan), dado que `/eventos` es ahora una ruta real (D-1). Hallazgo **M4** del ciclo 1.
  Con la corrección del bug 1 esas mismas aserciones pasan a afirmar el ancla **absoluta**
  (`/#ancla`) con idéntica forma (verificado en `revision-2026-10-12-codigo.md`): misma intención,
  valor actualizado — no es debilitamiento.

### Deuda aceptada (registrada; no bloquea el PR)

- **Documentación (resuelta en la tanda 2)** — los 2 hallazgos **IMPORTANTE** de documentación del
  revisor del 2026-10-12 quedaron **cerrados**: `plan.md` **P4-19** está realineado con `ux.md`
  **D-24** (bloque tras el hero, antes de «quiénes somos», confirmada por el humano) y **RG4-11**
  ya describe que «el bloque **nunca** se omite» (mensaje + `Notice` de error con «Volver a
  intentar»). No queda deuda documental pendiente.
- **Seguridad (bajos, ciclo 1)** — **B1**: el `PATCH` no revalida la longitud de los campos
  opcionales (la BD corta con `400` genérico); **B2**: validación de `imageFile` por patrón
  (UUID no enumerable); **B3**: contraseña de prueba fija en un e2e. Ninguno es crítico/alto/medio.
- **Revisor (menor)** — **M5**: tipo `Optional[T]` duplicado entre `portada` y `agenda`.
- **QA (manual, fuera de `make ci`)** — **SC-010/SC-015** (usabilidad con ≥ 6 personas, protocolo
  en `quickstart.md` §10.7, ampliado con una tarea sobre la portada) y la **lectura con lector de
  pantalla físico** (NVDA/VoiceOver) del checklist WCAG 2.1 AA de `/eventos` **y del bloque de la
  portada**: verificación **manual con el humano**; lo automatizable (teclado, foco, jerarquía,
  `<time>`, `alt`, contraste, 200 %, 320 px, `aria-current`) ya está en verde.
- **Sugerencias de la tanda 2 (deuda menor, no bloquean)** — aserción de posición de WhatsApp
  retirada en `homePage.test.tsx` al cambiar el orden (se propone añadir el orden completo de
  secciones); selector muerto `#root > div` en `layout.test.tsx`; el docstring del e2e de fondo
  promete 4 pantallas y cubre 3 (`/cambiar-contrasena` requiere sesión); el token
  `--color-panel-ink` queda **fuera del sistema F10** y aplicado a medias junto a
  `text-slate-900/700` (convendría usarlo en las 4 pantallas o documentarlo como excepción en
  `ux.md`); y un posible destello previo al montaje en `/login` con tema oscuro (el arreglo es
  post-montaje de React — no urgente).
- **Sugerencias de código (2026-10-12, no bloqueantes)** — `useLocation` en `try/catch` (mejor
  `MemoryRouter` en las pruebas), lógica «retained» duplicada entre hooks (`useRetainedQuery`) y
  `handled` de `useHashScrollOnArrival` que no re-enrola en la misma ruta+hash.
- **Sugerencias del ciclo 2 (no bloqueantes)** — **S1** deriva documental (`data-model.md`/
  `tasks.md` aún nombran `ListUpcomingEventDates`, retirada), **S2** sin prueba que fije la rama
  de conflicto de `UpdateEvento`, **S3** estilo de un comentario.
- **QA (menores conocidos)** — contención de `testcontainers` sin `-p 1` (heredada de F3), el
  flake aislado de un `toHaveURL` en la tanda del 2026-10-11, y la nota de entorno de 2026-10-12:
  sin `DATABASE_URL_TEST` exportada las pruebas de repositorio se **omiten en silencio** (`t.Skip`);
  conviene que `make test`/`make ci` la exporte o falle explícitamente (otro ciclo).

### Notas de entrega

- Estados: **47 de 49 tareas completas** (T401–T439 y T442–T449; T443–T449 añadidas por la
  enmienda Q8). Quedan abiertas solo **T440** (verificación de cierre: sus suites corrieron en la
  re-validación del 2026-10-12) y **T441** (documentación y PR, esta misma entrega), que se
  cierran con el merge. Esta descripción del PR es la parte **T441** de la fase 9 (entrega), junto
  con la documentación. Documentación actualizada por `documentador`: `CHANGELOG.md` (**0.5.0**, con «Agregado» del
  bloque, «Cambiado» con el reorden y «Corregido» con los **4** bugs), `README.md` (bloque de la
  portada, endpoint `/api/v1/portada/proximamente`, contrato 0.6.0 y e2e del 2026-10-12),
  `docs/entrega/F4-eventos-y-actividades.md` (nota para el cliente, con «Correcciones incluidas»
  y «Ajuste de posición del bloque») y este `pr.md`.
- El contrato `backend/api/openapi.yaml` está en **0.6.0** y `make api-gen`/`make sqlc-verify`
  quedaron sin deriva.
- **Spec v3 y plan re-aprobados por el humano el 2026-10-12**: el cambio de alcance Q8 está
  aplicado en código y aprobado en ambas puertas; **no queda nada pendiente de re-aprobación**.
- **Merge y despliegue**: solo con aprobación humana explícita (puertas pendientes en
  `estado.md`); después, `make costos CERRAR=1`.
