# PR — F4 Eventos y actividades

> Descripción para el Pull Request de la rama `004-eventos-actividades` (preparada por
> `documentador`, 2026-10-11; T441). **El merge y el despliegue requieren aprobación humana
> explícita** (fase 9/9 del flujo equipo-feature; puertas «PR / merge» y «Despliegue» pendientes
> en `estado.md`). El entorno no tiene `gh`: quien abre el PR es `devops`/el humano con esta
> descripción.

## feat: F4 — Eventos y actividades

### Resumen

Primera sección de contenido dinámico del sitio público: **lo que viene**, visible sin cuenta, y
los tres módulos del panel con los que el equipo lo administra. Cierra las **42 tareas
(T401–T442)** de `tasks.md` (US1–US6 / FR-001…FR-024 de la spec aprobada v2).

**Backend (Go + PostgreSQL, módulos `internal/agenda` e `internal/media`)**
- Documento público **`GET /api/v1/proximamente?lang=es|en`** (`Cache-Control: no-store`): dos
  bloques (`events` / `activities`) con claves **omitidas** cuando no hay contenido, orden
  cronológico por la próxima fecha, fallback `en → es` por campo y `400` para `lang` no soportado.
- Subgrupo de administración `/api/v1/admin/{eventos,actividades,categorias}` con CRUD,
  publicación/retirada, imágenes y catálogo de categorías; contrato vivo **0.4.0 → 0.5.0**
  (`backend/api/openapi.yaml`, fusión aditiva del delta `specs/…/contracts/openapi.yaml`, que no
  se edita) y tipos TS regenerados con `make api-gen`.
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
  (título/nombre actual) también en los rechazos.
- **Permisos**: se activan los reservados `eventos`/`actividades` y se añade **`categorias`**
  («Categorías de agenda», 10.º del catálogo de F2), cada uno con su `AdminChain` y su
  `RequirePermission`; los tres son independientes.
- `internal/media` promovido desde `portada` (misma subida/URL y reglas que F3). **0 dependencias
  nuevas**; cobertura `internal/agenda/service*.go` **89,4 %** en la verificación de cierre
  (≥ 80 % exigido).

**Frontend (React + TypeScript + Vite)**
- **Sección pública `/eventos`** con el sistema de F10: entrada única «Eventos y actividades»
  (barra, menú móvil y pie, **oculta** sin contenido publicado y próximo), dos bloques
  «Próximos eventos» / «Próximas actividades», tarjetas con `DateBadge`/`MapLink`/`CategoryChip`,
  imágenes, **aviso de sección vacía sin redirigir** y 16 claves i18n `es`/`en` con fechas
  localizadas por idioma (`formatoFechas.ts`, `Intl` sin conversión de zona).
- **Panel**: `/panel/eventos`, `/panel/actividades` y `/panel/categorias` (`EventsPage`,
  `ActivitiesPage`, `CategoriesPage`, `EventForm`, `ActivityForm`, `CategoryForm`, `DatesEditor`,
  listados con marca «Pasado», `409` de duplicado y de categoría en uso), permisos filtrados en
  menú, `AccesosRapidos` y rutas; `ImageUploader` y `PublishControls` **promovidos al inventario
  único** `src/components/`; `ACTION_LABELS` con etiqueta humana para los 15 códigos nuevos.
- **Vitest 74 archivos / 505 pruebas**, `lint`, `typecheck` y `build` en verde.

**Infraestructura / Docker (devops)**
- **`APP_TZ=America/Costa_Rica`** documentada en `.env.example` (T410, leída por
  `platform/config`; valor por defecto idéntico, el service valida la zona IANA y falla con
  error claro si es inválida).
- **Sin cambios** en `docker-compose.yml`, Dockerfiles ni `.github/` (árbol limpio, verificado
  en T440).

### Decisiones firmes de la puerta 2 (aplicadas, no reabiertas)

- **`APP_TZ = America/Costa_Rica`**: el corte de «hoy» se calcula en esa zona en el servidor.
- **Ruta pública `/eventos`** (no `/eventos-actividades`) y panel en `/panel/eventos`,
  `/panel/actividades`, `/panel/categorias`.
- **Aviso sin redirigir**: sin nada publicado y próximo, `/eventos` muestra `HiddenNotice` y la
  entrada de navegación queda oculta.
- **Permiso nuevo «categorías»** (catálogo de F2: 9 → 10, migración `000009`; `000002` intacto).
- **`endTime` exige `startTime`**, **borrado de categoría en uso → `409`** e **imagen de elemento
  vencido no servida**.

### Cómo probarlo (quickstart)

Recorrido completo con evidencia por sección en [`quickstart.md`](./quickstart.md) (§0–§14).
Resumen:

```bash
make up                 # db, redis, backend, frontend
make db-migrate         # 000007, 000008 y 000009
cd frontend && npm install
# Sección pública (visitante): http://localhost:5173/eventos
curl -s 'http://localhost:8080/api/v1/proximamente?lang=es' | jq
# Panel: login con la cuenta sembrada ana@ejemplo.com / Semilla.2026 (helpers de e2e)
# -> /panel/eventos, /panel/actividades y /panel/categorias con sus permisos
```

Pruebas automatizadas: `make ci` (lint + backend unit/integración + frontend + security) y
`npx playwright test --config e2e/playwright.config.ts --workers=1` desde `frontend/`, **en tres
tandas** (ver abajo).

### Evidencia de pruebas (2026-10-11, verificación de cierre T440)

| Verificación | Resultado |
|---|---|
| `make ci` | **en verde**: `gofmt`/`go vet`/`golangci-lint` sin issues · backend `go test ./...` + `go test -tags=integration ./...` (PostgreSQL real) · frontend Vitest **74 archivos / 505 pruebas** · `govulncheck` **0 vulnerabilidades que afectan al código** · `npm audit --audit-level=high` **0** |
| Backend integración | **20/20 paquetes ok, 0 FAIL** con `-p 1` (la corrida sin `-p 1` puede fallar por contención de `testcontainers` en `internal/platform/session`: infra de pruebas, no de F4) |
| e2e Playwright (3 tandas, `--workers=1`, imagen Docker de Playwright) | **F4 5/5** (`agenda-publica`, `agenda-panel`, `agenda-accesibilidad`) · **F2 2/2** (`acceso`, `auditoria`, congeladas) · **F3/F10 31/31** (`portada-`, `diseno-`, `movimiento`, `navegacion-`, `tema-`, `qa-`, `status`, congeladas) — **38/38** en total |
| Cobertura `internal/agenda/service*.go` | **89,4 % statements** (≥ 80 % exigido); re-validada en los ciclos: 86,6 % (integración) / 91,5 % (unitarias de `service*`) |
| Deriva de generados | `sqlc-verify` y `api-gen` **sin diff** (1 commit de diff: 0) |
| Migraciones | `000007`–`000009` **up → down → up** en verde; BD dev en versión 9 |
| Docker / entorno | 4 servicios arriba · `GET /healthz` 200 · ruta inexistente `404` con `ErrorEnvelope` · `GET /api/v1/proximamente?lang=es` **200 `no-store`** con base vacía (`{"lang":"es"}`, SC-012) · `lang=fr` → `400 invalid` · catálogo de permisos = **10 con `categorias`** · **sin cambios** de F4 en compose, Dockerfiles ni `.github/` |
| Contrato público sin regresiones | `GetProximamente` nunca expone `publicationState`, omite borradores/vencidos, respeta el orden y el fallback `en → es` (pruebas unitarias + integración en verde) |

*Nota*: en una corrida intermedia la tanda 3 falló un caso con `toHaveURL` y salió **31/31** al
re-lanzarla dos veces; queda registrado como **flake aislado** (sin defecto de F4 identificado).

### Validación del equipo (ciclo 2 — APROBADO por los tres roles)

Ciclo 1 halló 1 IMPORTANTE + 3 menores; todos corregidos y re-validados en el ciclo 2 con
**0 bloqueantes** (commits `1714e2b`, `3ec0b2d`, `53c7007`, `8b03db3`):

- **I1 (IMPORTANTE)** — consultas N+1 en `GetProximamente`/`ListEventos`/`ListActividades` →
  consultas por lote (`event_id = ANY($1::uuid[])`): 1 consulta por bloque y 1 por página, sin
  SQL huérfano ni cambios de comportamiento (`1714e2b`), con pruebas que fijan el conteo.
- **M1** — código muerto `IsHomeFilePublished` retirado del fake de portada (`3ec0b2d`).
- **M2** — tope de categorías por conteo (`CountCategories`, 1 conteo / 0 listados) (`53c7007`).
- **M3** — etiquetas de auditoría con el título/nombre actual en los `fail` (`8b03db3`).

Informe consolidado: **[`revision-2026-10-11-ciclo2.md`](./revision-2026-10-11-ciclo2.md)** (veredicto
global **APROBADO**). Por rol (ciclo 2; el ciclo 1 queda como registro en `revision-2026-10-11.md`):

| Informe | Veredicto | Nota |
|---|---|---|
| [`revision-2026-10-11-qa-ciclo2.md`](./revision-2026-10-11-qa-ciclo2.md) | **APROBADO** | Go unit + integración 20/20 (`-p 1`), Vitest **505/505**, e2e **5/5** F4 (+ 2/2 F2 y 31/31 F3/F10), vista pública intacta, matriz US1–US6 / SC-001…SC-013 sin pruebas eliminadas ni debilitadas |
| [`revision-2026-10-11-codigo-ciclo2.md`](./revision-2026-10-11-codigo-ciclo2.md) | **APROBADO** | I1 y M1–M3 cerrados; sin pruebas congeladas tocadas ni `skip`/`only`/`xit`; contrato 0.5.0 sin cambios; cobertura 91,5 % |
| [`revision-2026-10-11-seguridad-ciclo2.md`](./revision-2026-10-11-seguridad-ciclo2.md) | **APROBADO** | SQL parametrizado, filtro público y separación público/panel intactos, `govulncheck`/`npm audit` sin altas/críticas, **sin secretos**; 2 informativos heredados |

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
  se conservan), dado que `/eventos` es ahora una ruta real (D-1). Hallazgo **M4** del ciclo 1:
  legítimo, pero no estaba registrado; queda registrado aquí y en `estado.md`.

### Deuda aceptada (registrada; no bloquea el PR)

- **Seguridad (bajos, ciclo 1)** — **B1**: el `PATCH` no revalida la longitud de los campos
  opcionales (la BD corta con `400` genérico); **B2**: validación de `imageFile` por patrón
  (UUID no enumerable); **B3**: contraseña de prueba fija en un e2e. Ninguno es crítico/alto/medio.
- **Revisor (menor)** — **M5**: tipo `Optional[T]` duplicado entre `portada` y `agenda`.
- **QA (manual, fuera de `make ci`)** — **SC-010** (usabilidad con ≥ 6 personas, protocolo en
  `quickstart.md` §10.7) y la **lectura con lector de pantalla físico** (NVDA/VoiceOver) del
  checklist WCAG 2.1 AA de `/eventos`: verificación **manual con el humano**; lo automatizable
  (teclado, foco, jerarquía, `<time>`, `alt`, contraste, 200 %, 320 px) ya está en verde.
- **Sugerencias del ciclo 2 (no bloqueantes)** — **S1** deriva documental (`data-model.md`/
  `tasks.md` aún nombran `ListUpcomingEventDates`, retirada), **S2** sin prueba que fije la rama
  de conflicto de `UpdateEvento`, **S3** estilo de un comentario.
- **QA (menores conocidos)** — contención de `testcontainers` sin `-p 1` (heredada de F3) y el
  flake aislado de un `toHaveURL` en la tanda 3.

### Notas de entrega

- Estados: **42/42 tareas** completas; esta descripción del PR es la parte **T441** de la fase 9 (entrega), junto con la documentación. La deriva documental **S1** queda registrada como sugerencia no bloqueante (ver «Deuda aceptada»). Documentación actualizada por `documentador`: `CHANGELOG.md`
  (**0.5.0**), `README.md` (sección F4, `/eventos`, `APP_TZ`, 10 permisos, migraciones y e2e en
  3 tandas), `docs/entrega/F4-eventos-y-actividades.md` (nota para el cliente) y este `pr.md`.
- El contrato `backend/api/openapi.yaml` está ya en **0.5.0** y `make api-gen`/`make sqlc-verify`
  quedaron sin deriva.
- **Merge y despliegue**: solo con aprobación humana explícita (puertas pendientes en
  `estado.md`); después, `make costos CERRAR=1`.
