# PR — F3 Portada e información general

> Descripción para el Pull Request de la rama `003-portada-info-general` (preparada por `devops`,
> 2026-10-10). **El merge y el despliegue requieren aprobación humana explícita** (fases 9/9 del
> flujo equipo-feature; puertas «PR / merge» y «Despliegue» pendientes en `estado.md`).

## feat: F3 — Portada e información general

### Resumen

La primera cara pública del sistema: portada web editable con identidad, contenido y
auditoría, sobre la base de F1/F2 (Go + PostgreSQL + React). Cierra las
**40 tareas (T301–T340)** de `tasks.md`.

**Backend (Go + PostgreSQL, módulo `internal/portada`)**
- Contrato OpenAPI extendido a **0.4.0** (`backend/api/openapi.yaml` + `specs/…/contracts/openapi.yaml`)
  con API pública (`GET /api/v1/portada`, media) y de administración (identidad, quiénes somos,
  horario, WhatsApp, redes, contacto, publicación e imágenes).
- Migraciones **000005** (`home_content`) y **000006** (extensión de `admin_actions` para contenido
  de portada); queries y código **sqlc** generado (`home.sql`, `repository_home.go`).
- Servicios con 85.3 % de cobertura en `service*.go`: localización `es`/`en` con fallback (FR-009),
  estados `draft/published`, publicación por sección/elemento (FR-013), `PATCH` que distingue
  `null` de campo ausente (tipo `Optional[T]`), validación de URLs (catálogo de redes, dominio
  oficial, `https://wa.me/…`/grupos), normalización de teléfonos/WhatsApp con código país.
- Imágenes: almacenamiento local en `UPLOAD_DIR` (contenedor: `/var/lib/simiente/uploads`,
  volumen `uploads_data`), límite `UPLOAD_MAX_BYTES` (8 MB), validación por **firma** (no por
  extensión), descarga solo si la imagen es referencida por contenido **publicado**, `Cache-Control:
  no-store` + `nosniff` en `/api/v1/media`.
- Auditoría completa (FR-017, SC-013): `home.*.update/create/delete/publish/unpublish/image.upload`
  con `targetLabel «Portada · <sección> · <elemento>»`, resultado `success/failure/denied`,
  registro solo-lectura; denegaciones del permiso `portada` también quedan registradas.
- Permiso de módulo `portada` reutilizado del catálogo de F2 (migración `000002`, sin datos nuevos).

**Frontend (React + TypeScript + Vite)**
- Marca según el Manual de Identidad del cliente (`resources/` versionado: paleta
  `#1a2b4a`/`#00c9a7`/…, tipografías Bebas Neue/Poppins/Playfair Display, logo `simiente.jpeg`).
- i18n `es`/`en` con memoria en `localStorage` (primera visita en español, FR-010).
- **Portada pública en `/`** (estados: sin contenido, cargando, con secciones publicadas; las
  secciones sin elementos publicados desaparecen por completo, SC-012); la página «Estado del
  sistema» de F1 se reubica en **`/health`**; horario presentado en a.m./p.m. localizado («10:00
  a. m. − 12:00 m.» / «10:00 AM − 12:00 PM», dato del contrato sigue en 24 h).
- **Panel de información**: editor por secciones, publica/retira por elemento, edición visible al
  guardar (FR-014), subida de imágenes con texto alternativo obligatorio, selector de idioma con
  objetivo táctil ≥ 44 px.
- 273 pruebas unitarias (45 archivos) con Vitest.

**Infraestructura / Docker (devops)**
- `docker-compose.yml`: volumen nombrado **`uploads_data`** montado en
  `UPLOAD_DIR=/var/lib/simiente/uploads` del backend (usuario no-root con permisos de escritura
  verificados por e2e) y `UPLOAD_MAX_BYTES=8388608`; **`.env.example` documenta ambas variables**
  (por defecto local `./uploads`, en el contenedor la fija el compose).
- Sin cambios en `.github/workflows/` ni en archivos del kit (verificado vs merge-base `0b2f525`).

### Cómo probarlo (quickstart)

Recorrido completo con evidencia por sección en [`quickstart.md`](./quickstart.md) (§0–§12).
Resumen:

```bash
make up                 # db, redis, backend, frontend
make db-migrate         # 000005 y 000006
cd frontend && npm install
# Portada pública (visitante): http://localhost:5173/  ·  estado: /health
curl -s http://localhost:8080/api/v1/portada?lang=es | jq
# Panel: login con la cuenta sembrada ana@ejemplo.com / Semilla.2026 (helpers de e2e)
# -> sección "Información" con el permiso `portada`
```

Pruebas automatizadas: `make ci` (lint + backend unit/integración + frontend + security) y
`make e2e` / `npx playwright test --config e2e/playwright.config.ts --workers=1` desde `frontend/`.

### Evidencia de pruebas (2026-10-10, HEAD `583d3c1`)

| Verificación | Resultado |
|---|---|
| `make ci` | **EXIT=0** · `gofmt`/`go vet`/`golangci-lint` **0 issues** · backend `go test ./...` + `go test -tags=integration ./...` (PostgreSQL real) **ok** · frontend Vitest **45 archivos / 273 pruebas pasadas** · `govulncheck` **0 vulnerabilidades que afecten al código** · `npm audit --audit-level=high` **0 vulnerabilidades** |
| e2e Playwright (`--workers=1`, stack Docker real) | **7/7 passed (25.6 s)**: acceso, auditoría, portada-accesibilidad (WCAG automatizable), portada-panel, portada-patch-null (regresión B1), portada-pública (idioma/fallback/responsiva 320/768/1280), status `/health` |
| Cobertura `internal/portada/service*.go` | **85.3 % statements (529/620)** — `service.go` 89.1 %, `service_admin.go` 83.4 %, `service_audit.go` 95.5 %, `service_public.go` 87.7 % (≥ 80 % exigido) |
| Deriva de generados | `sqlc generate` + `git diff -- backend/internal/db sqlc.yaml` → **sin diff** · `npm run api:gen` + `git diff -- frontend/src/api/schema.d.ts` → **sin diff** |
| Docker / entorno | `docker compose config` **válido** · `.env.example` documenta `UPLOAD_DIR` y `UPLOAD_MAX_BYTES` · `UPLOAD_DIR=/var/lib/simiente/uploads` coherente con el volumen `uploads_data` (`:uploads_data:/var/lib/simiente/uploads`) |
| Smoke operativo | `GET /healthz` → 200 ok · `GET /api/v1/portada?lang=es` → 200 sin `publicationState` (quickstart §0–§12 recorrido en `cierre.md`) |

### Validación del equipo (ciclo 2 — APROBADO por los tres roles)

Ciclo 1 halló 4 hallazgos; todos corregidos y re-validados en el ciclo 2 con **0 bloqueantes**:

- **B1 (BLOQUEANTE)** — los `PATCH` no distinguían `null` de campo ausente y la limpieza no
  persistía: `d5d4610` (tipo `Optional[T]`), 3 pruebas unitarias de regresión + e2e en el stack
  real `portada-patch-null.spec.ts` (`d98df37`) con auditoría viva. Cerrado en las tres capas.
- **I1 (IMPORTANTE)** — horario en 24 h -> presentación a.m./p.m. localizada (`fc96185`), con
  pruebas `schedule.test.ts` y `homePage.test.tsx`.
- **A1 (MENOR)** — objetivo táctil del selector de idioma 20 px -> 44 px (`2b72138`).
- **N1 (MENOR)** — «m.» solo para el mediodía exacto, «12:45 p. m.» en la franja (`db83a61`).

Informes (ciclo 2, veredicto **APROBADO** los tres; los del ciclo 1 quedan como registro):

| Informe | Veredicto | Hallazgos |
|---|---|---|
| [`revision-2026-10-10-codigo-ciclo2.md`](./revision-2026-10-10-codigo-ciclo2.md) | **APROBADO** | 0 bloqueantes · B1/I1 cerrados con prueba de cable real (JSON, `DisallowUnknownFields` intacto) |
| [`revision-2026-10-10-qa-ciclo2.md`](./revision-2026-10-10-qa-ciclo2.md) | **APROBADO** | 0 bloqueantes · suites completas en verde (backend + 273 unitarias + 7/7 e2e) |
| [`revision-2026-10-10-seguridad-ciclo2.md`](./revision-2026-10-10-seguridad-ciclo2.md) | **APROBADO** | 0 bloqueantes · 14 pruebas dinámicas (P1–P14): tipos, `null`, CSRF, permisos, auditoría y vía pública sin fugas |

Pendientes coordinables por el humano (así se deja asentado):

- **SC-010 (usabilidad, pendiente de coordinación humana)** — prueba con ≥ 6 personas (2 por
  franja 18–35 / 36–59 / 60+), protocolo en `quickstart.md` §10.7, informe
  `pruebas-usabilidad-SC-010.md` (umbral ≥ 90 % de tareas completadas). No automatizable; queda
  **pendiente de coordinación humana** junto con `qa-tester`.
- **SC-008 (accesibilidad, pendiente de coordinación humana)** — el e2e cubre la parte
  automatizable (estructura, teclado, foco, toque ≥ 44 px, contraste, texto al 200 %); la **pasada
  manual con lector de pantalla real** (NVDA/VoiceOver) del checklist WCAG 2.1 AA queda pendiente
  de coordinación humana (`qa-tester`).

### Deuda aceptada (registrada, no bloquea el PR)

- **M-a11y (QA, menor)** — reflow al 200 % de zoom sobre un teléfono de 320 px en las tarjetas de
  contacto (viewport efectivo 160 px); WCAG 1.4.4/1.4.10 siguen cubiertos por el e2e.
- **Código (menores)** — M1 (reservas de error de `UploadImage` sin `recordContentFailure`),
  M2 (fila de auditoría al guardar sin cambios), M3 (lectura de `previous` fuera de transacción y
  borrado de imagen reemplazada best-effort), N2 (`Optional[T]` sin `MarshalJSON`, latente),
  N3 (longitud de campos PATCH validada solo por `CHECK` de BD -> 400 genérico).
- **Seguridad (menores, higiene dentro del perímetro autorizado)** — S1 (patrón de nombre de
  imagen subida), S2 (auditoría de subidas rechazadas), S4 (`nosniff` solo en media), S5
  (huérfanos acumulables en `uploads_data`), S6 (nota de auditoría de `null` sobre valor ya
  vacío). S3 quedó resuelta por el fix de B1.
- Todas van anotadas para una funcionalidad futura; ninguna incumple un criterio de la spec.

### Notas de entrega

- Estados: 40/40 tareas, **Converged** (fase 8); esta preparación del PR cierra la parte `devops`
  de la fase 9.
- Tras el merge: `documentador` (CHANGELOG/README/quickstart — tiene 3 discrepancias anotadas en
  `cierre.md` §3) y cierre de costos con `make costos CERRAR=1`.
