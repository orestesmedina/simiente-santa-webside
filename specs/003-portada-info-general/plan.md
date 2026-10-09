# Implementation Plan: Portada e información general (F3)

**Branch**: `003-portada-info-general` | **Date**: 2026-10-09 | **Spec**: [spec.md](./spec.md)
(aprobada por el humano el 2026-10-09, con las 10 aclaraciones resueltas)

**Input**: `specs/003-portada-info-general/spec.md` (fuente de verdad) · `.specify/memory/constitution.md`
· `docs/tecnico/arquitectura.md` (reglas R1–R8 y convenciones §8.1) y `docs/tecnico/decisiones.md` ·
**F2 terminada** (`specs/002-acceso-gestion-usuarios/`: `plan.md`, `research.md`, `data-model.md`,
`contracts/openapi.yaml`, `quickstart.md` se usan como plantilla **y** como código existente a
reutilizar: sesión, CSRF, `authz`, `apperr`, `validate`, `paginate`, auditoría y catálogo de
permisos) · código real de `backend/` y `frontend/` · skills `go-backend`, `postgres-db`,
`react-frontend` · decisiones 4, 6 y 8 del roadmap · Manual de Identidad del cliente
(`resources/MANUAL DE MARCA.pdf`, logo `resources/simiente.jpeg`).

## Summary

F3 es la **primera cara pública del sitio**: la portada que cualquier visitante ve sin cuenta, con la
identidad de la iglesia, «quiénes somos», el horario de servicios, los canales de WhatsApp, las redes
sociales y los datos de contacto — **bilingüe** (es base, en opcional con fallback, Decisión 6/8) — y
con todo ese contenido **editable y publicable por elemento desde el panel** (Decisión 4, FR-013/FR-014)
bajo el permiso **«portada e información general»** (`portada`) que F2 ya sembró en su catálogo
(`000002_create_roles_and_permissions`: verificado en el código), con **toda edición auditada** en el
registro de F2 (FR-017).

Técnicamente F3 añade: el **primer contenido bilingüe** (patrón de columnas `*_es`/`*_en`
reutilizable por F4–F9, [research.md](./research.md) R3-1), el **primer estado de publicación
borrador/publicado** (R3-3), la **primera superficie pública de lectura** (`GET /api/v1/portada`,
resuelta por idioma en el servidor, R3-2), el **primer almacenamiento de archivos** (logo e imagen de
portada en disco local con volumen Docker, `platform/storage`, R3-8), el **primer i18n del frontend**
(diccionarios tipados es/en, R3-9) y la **primera ampliación del registro cerrado de auditoría**
(14 códigos `home.*` y `target_kind='content'`, migración `000006`, R3-11). Todo se construye con lo
que F1/F2 ya dejó: capas `handler → service → repository`, `platform/` (config, logger, apperr,
httpserver, middleware, validate, paginate, session, audit), sqlc sobre `internal/db/`, DI manual en
`main.go`, sobres uniformes, `/healthz` intacto y la cadena de panel `authn → guard → authz → CSRF`.

**Ninguna dependencia de runtime nueva en backend**; en frontend solo las 3 tipografías del Manual de
Identidad (`@fontsource/…`, R3-16).

## Technical Context

**Language/Version**: Go 1.27 (sin cambios) · TypeScript 5.x `strict` sobre Node 22 · React 19.

**Primary Dependencies**: **ninguna nueva de runtime en backend** (`net/http` para multipart y
descarga de archivos, `net/url`, `http.DetectContentType`, `os`, `pgx`/`sqlc` ya existen; pruebas con
`testcontainers-go` de F2). Frontend: `@fontsource/bebas-neue`, `@fontsource/poppins`,
`@fontsource/playfair-display` (**solo assets de build**, R3-12/R3-16). Se reutilizan React Router,
TanStack Query, React Hook Form + Zod, Tailwind, Vitest + Testing Library + MSW y Playwright.

**Storage**: **PostgreSQL 16** (tablas `home_*` + ampliación de `admin_actions`; migraciones
`000005`/`000006`, ver [data-model.md](./data-model.md)) · Redis 7 **sin tocar** (sesión de F2) ·
**disco local del backend** para imágenes (`UPLOAD_DIR`, volumen Docker `uploads_data`; R3-8). Capa de
datos sqlc (D-A3): consultas en `backend/internal/db/queries/home.sql`, código generado commiteado;
la auditoría reutiliza `InsertAdminAction` de `queries/audit.sql` **dentro de la transacción** de cada
mutación (R3-11).

**Testing**: estrategia por capa de F1/F2 ampliada (sección propia): `go test` (unitarias),
`go test -tags=integration` (repository contra PostgreSQL real — incluida la migración up/down/up de
`000006` — y almacenamiento de archivos con `t.TempDir()`), Vitest + Testing Library + MSW, Playwright
(e2e de la portada pública y del panel).

**Target Platform**: Docker Compose local (nuevo volumen `uploads_data`) + GitHub Actions (CI del kit,
sin tocar). Sin despliegue a producción (fuera de alcance de F3).

**Performance Goals**: `GET /api/v1/portada` < 500 ms con BD sana (una consulta por colección; JSON
pequeño, `no-store`) · panel con el agregado completo en < 1 s · subida de imagen ≤ 8 MB validada en
streaming sin cargarla toda en memoria · imágenes servidas con caché `immutable` (un nombre por
subida) para que la portada cargue bien en móvil.

**Constraints**: stack fijo (constitución) · capas `handler → service → repository` (§II) · contrato
OpenAPI antes que el código (§II) · migraciones versionadas e inmutables (§VI) · validación de toda
entrada en backend (§IV, CWE-20) · nada de contenido ejecutable (§IV, CWE-79: sin
`dangerouslySetInnerHTML`, sin SVG en imágenes) · authn/authz/CSRF en servidor (§IV, CWE-862) ·
dependencias minimizadas y justificadas (D-A8) · archivos del kit no editables (regla 10) · **la spec
no se reabre**: los huecos se cierran como decisión documentada en `research.md`.

**Scale/Scope**: 16 operaciones REST (2 públicas: portada y media · 14 de panel) sobre 6 tablas
nuevas + la ampliación de `admin_actions`; 1 dominio backend nuevo (`portada`) + 1 plumbing nuevo
(`platform/storage`) + extensión de `platform/audit` y `platform/validate`; 2 migraciones
(`000005`, `000006`); frontend: 2 features nuevas (`public`, `portada`) + módulo `i18n/` + ajustes de
router/permisos; **0 dependencias de runtime nuevas en backend**. Sin puertas de aprobación abiertas
más allá de **este plan** (regla 1 de AGENTS.md).

## Constitution Check

*GATE: debe pasar antes de la investigación y re-verificarse tras el diseño.*

| Principio | Evaluación | Resultado |
|---|---|---|
| §I La spec manda | El plan implementa FR-001…FR-019 e US1–US5 sin interpretarlos; cada uno tiene fila en "Cobertura de requisitos". Los huecos de la spec ya están cerrados con decisiones del humano (Q1–Q10 del 2026-10-09) o como decisión de diseño revisable en `research.md` (memoria del idioma R3-9, duplicado de WhatsApp R3-6, chrome sin identidad R3-18, catálogo de redes R3-7). No se construye nada de Out of Scope (sin F4–F9, sin formularios, sin traducción automática, sin historial de versiones, sin verificación de enlaces, sin bilingüismo del panel). | ✅ |
| §II Arquitectura | Monorepo; capas `handler → service → repository` sobre `internal/platform/` (R1–R8); API REST JSON documentada en `backend/api/openapi.yaml` con el delta redactado **antes** que el código (`contracts/openapi.yaml`, §8.1.5); dependencias nuevas: 0 en backend, 3 de assets tipográficos en frontend, justificadas (R3-16). | ✅ |
| §III Pruebas | Toda tarea llevará sus pruebas (se exigirá en `tasks.md`); cobertura ≥80 % en `service/`; repository contra PostgreSQL real (`//go:build integration`), incluida la migración `000006` (up/down/up) y la atomicidad mutación+auditoría; frontend con Vitest + Testing Library + MSW; flujos críticos con Playwright (portada pública, idioma, publicar/retirar). | ✅ |
| §IV Seguridad | SQL solo parametrizado vía sqlc (CWE-89); validación de toda entrada en backend (`platform/validate` + reglas de dominio, CWE-20); texto plano sin ejecución (FR-003/FR-015): sin `dangerouslySetInnerHTML`; imágenes con firma binaria, tipos cerrados, **sin SVG** (XSS almacenado), nombre generado por el servidor (sin path traversal), `nosniff` y `Content-Disposition: inline` (CWE-79/CWE-434/R3-8); operaciones de panel con sesión + CSRF + permiso `portada` verificados en servidor (CWE-862); secretos: ninguno nuevo (`UPLOAD_DIR`/`UPLOAD_MAX_BYTES` no son secretos); `govulncheck`/`npm audit` en el CI del kit. Borradores **nunca** hacia el público: consultas `…Published` separadas (FR-013/SC-002). | ✅ |
| §V Calidad | `gofmt`/`go vet`/`golangci-lint` sin errores; TS `strict` sin `any`; errores envueltos con `%w` y traducidos solo en `WriteError`; nombres de código en inglés (las carpetas de dominio siguen la convención ya establecida: `usuarios` → `portada`, con tablas en inglés `home_*`). | ✅ |
| §VI Base de datos | 2 migraciones versionadas con `up`/`down` completos (`000005`, `000006`); ninguna migración existente se modifica; tablas con `id`, `created_at`, `updated_at`; FK e índices explícitos (los únicos FK nuevos son los de F2 en `admin_actions`, ya existentes; el contenido no referencia usuarios); `CHECK`/`UNIQUE` en la base (bilingüe, publicación, catálogo de redes, duplicados de WhatsApp, singleton). | ✅ |
| §VII Observabilidad | Logs `log/slog` con `request_id` (incluidos los fallos de validación y de subida, sin cuerpos de petición); auditoría duradera ampliada a F3 en `admin_actions` (FR-017); `/healthz` intacto y sin sesión; config por variables de entorno (`UPLOAD_DIR`, `UPLOAD_MAX_BYTES`); `docker compose up` levanta todo (nuevo volumen `uploads_data`). | ✅ |
| §VIII Gobierno | **Este plan requiere aprobación humana** antes de `/speckit.tasks` y de cualquier código (regla 1). Sin despliegue a producción. Quien escribe no aprueba: todo pasa por `qa-tester`, `revisor-codigo` y `seguridad`. | ✅ |

**Sin violaciones que justificar** → "Complexity Tracking" solo registra desviaciones de convenciones
internas (receta y nombres de archivos/tablas).

## Decisiones técnicas (resumen; análisis completo en `research.md`)

Numeración **P3-1…P3-19**, propia de este plan (no confundir con P1–P23 de F2, R3-1…R3-18 de
`research.md` ni RG3-1… de "Riesgos").

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| **P3-1** | **Un dominio `internal/portada/`** para todo el módulo + plumbing **`platform/storage/`** (interfaz `Store` con `Save`/`Open`/`Delete` e implementación `LocalStore`) | Un dominio evita dos paquetes sobre las mismas tablas (R2) y encaja con la receta de `arquitectura.md` §8; el disco es plumbing como lo es PostgreSQL en `platform/database` | Dominios `contenido/` + `media/` separados (interfaces extra para un solo consumidor); guardar archivos dentro del repository (mezclaría `os` con SQL); librería de storage (dependencia) |
| **P3-2** | **Patrón bilingüe por columnas** `*_es NOT NULL` / `*_en NULL` por campo traducible (R3-1) | Decisión 8 fija 2 idiomas con inglés opcional: la regla es una propiedad de esquema, sin joins y con fallback trivial; reutilizable en F4–F9 | Tabla de traducciones EAV (tipos y `CHECK` perdidos); tabla hija por idioma (N tablas y joins para 2 idiomas); JSON `{"es","en"}` (sin validación por idioma) |
| **P3-3** | **`GET /api/v1/portada?lang=es\|en`** resuelve el idioma **en el servidor** con fallback por campo `en → es` (R3-2) | FR-009/SC-006 quedan garantizados en una sola capa (el service) y F4–F9 heredan la regla; el selector vuelve a pedir con el otro `lang` | Devolver ambos idiomas y resolver en cliente (regla duplicada por consumidor y huecos posibles); `Accept-Language` (opaco y no cacheable); un endpoint por idioma |
| **P3-4** | **`publication_state` por elemento** —identidad, «quiénes somos» y contacto incluidos: cada singleton es un elemento publicable **por separado**— + consultas **`…Published` separadas** para el público + secciones vacías omitidas; se publica al guardar (FR-014) | FR-013/Q7/Q10 literales: un solo camino público hace imposible filtrar un borrador (SC-002) y el frontend solo renderiza lo presente (SC-012) | Estado por sección/agrupado (contradice Q7 y FR-013: cada elemento, también los singletons, tiene su propio estado); flag en la respuesta pública (filtra borradores); estado resuelto en el frontend; paso intermedio de borrador al editar (contradice FR-014) |
| **P3-5** | **Singletons** (`home_identity`, `home_about`, `home_contact`) con `singleton BOOLEAN CHECK (singleton)` + `UNIQUE (singleton)` y **upsert** (`ON CONFLICT (singleton) DO UPDATE`) en transacción con `pg_advisory_xact_lock` (R3-4) | ≤1 fila garantizado por la BD sin relajar los `NOT NULL`/`CHECK` de negocio; el advisory lock evita la carrera de dos primeras escrituras (mismo patrón que F2/P7) | Sembrar la fila vacía en la migración (relajaría las restricciones); índice parcial `UNIQUE ((true))` (conflict target críptico); guardia solo en el service (carrera) |
| **P3-6** | **Horario estructurado**: `day_of_week SMALLINT (0=domingo…6)` + `start_time TEXT "HH:MM"` con `CHECK` + `sort_order` (R3-5) | Día y hora son valores ordenables y localizables por el cliente (el nombre del día lo pone el i18n); solo nombre/descripción y lugar son traducibles | Día como texto libre (imposible de ordenar ni traducir sin tocar datos); `TIMESTAMPTZ` (mezcla fecha y hora de un horario semanal); tipo `TIME` (conversión `pgtype` sin ganancia: el valor se muestra) |
| **P3-7** | **WhatsApp**: `kind direct\|group` + `destination` normalizada; `direct` valida teléfono (tag `phone` de F2) y arma `https://wa.me/<dígitos>`; `group` valida URL https de `chat.whatsapp.com`/`wa.me`; duplicado **exacto** (kind+destino+nombre normalizados) → `409` (R3-6) | Q4 pide ambos tipos con un único destino y el edge case define el duplicado; la normalización hace estable el `UNIQUE` de la BD | Dos columnas mutuamente excluyentes (CHECK cruzados); guardar solo la URL final (se pierde la validación del número); verificar el destino contra WhatsApp (fuera de alcance); duplicado = mismo destino sin mirar nombre (más estricto que la spec — anotado como default revisable) |
| **P3-8** | **Redes**: catálogo fijo en el `CHECK` de la BD (`facebook, instagram, youtube, tiktok, spotify`) + `UNIQUE (network)` + validación de URL https **del dominio oficial de la red** (R3-7) | Q5 pide catálogo fijo y un enlace por red; el host oficial evita el error frecuente de "enlace de Facebook" que apunta a otro sitio | Catálogo solo como constantes de Go (la BD no lo impondría); tabla `social_networks` (una tabla para 5 valores); cualquier URL válida (el edge case pide rechazar redes repetidas o fuera de catálogo) |
| **P3-9** | **Imágenes en disco local** (`UPLOAD_DIR`) + volumen Docker `uploads_data`; subida multipart (`net/http`) con **firma binaria** (JPEG/PNG/WebP; sin SVG ni GIF), ≤ 8 MB, nombre generado `img_<uuid>.<ext>`; descarga `GET /api/v1/media/{fileName}` con regex de nombre, `nosniff`, `inline` y caché `immutable` (R3-8) | Sin dependencias ni credenciales para el MVP en Compose; el volumen sobrevive a rebuilds como `pgdata`; el nombre generado elimina path traversal y el SVG el XSS almacenado | S3/Cloudinary (dependencia + red + secretos); bytes en `bytea` (bloat de la BD); servirlo con nginx (pierde control de cabeceras/tipos); aceptar cualquier formato que el navegador reproduzca (CWE-434) |
| **P3-10** | **i18n propio** en `frontend/src/i18n/`: `LanguageProvider` + `useLanguage()` + diccionarios `es.ts`/`en.ts` con la **misma clave tipada**; memoria en **`localStorage`** (entre visitas) (R3-9) | SC-006 se garantiza en compilación (falta una clave → no compila) sin dependencias; `localStorage` conserva la elección del dispositivo entre visitas y la **primera visita sin preferencia se muestra en español** (decisión del humano del 2026-10-09, que ajustó FR-010 para permitirlo; se refleja en la spec) | `react-i18next` (2 dependencias sin la garantía tipada); JSON por red (una petición más); `sessionStorage` (memoria solo por visita: era la primera elección, sustituida por decisión del humano); `navigator.language` (descartado por la spec: no se auto-detecta el idioma del navegador) |
| **P3-11** | **Se activa el permiso `portada`** ya sembrado por F2: subgrupo `/api/v1/admin/portada` con `middleware.AdminChain("portada", …)` (authn → guard de contraseña → authz → CSRF) + frontend `PORTADA` en `lib/permissions.ts`, `RequirePermission` y `isPermissionAvailable` (R3-10) | FR-012 literal y Decisión 5 (catálogo fijo de F2 sin cambios); la denegación es idéntica a la de F2 y se registra en `admin_actions` | Permiso nuevo (rompería el catálogo y la Decisión 5); middleware de autorización propio (ya existe `AuthzByModule`); comprobar permisos solo en el cliente (CWE-862) |
| **P3-12** | **Auditoría ampliada sin dominios cruzados** (R3-11): el repository de `portada` inserta `admin_actions` con la consulta compartida `InsertAdminAction` de `internal/db` **dentro de la transacción** de cada mutación (fail-closed); `platform/audit` crece con 14 códigos `home.*` y `TargetKindContent`; la migración `000006` extiende los `CHECK` (registro cerrado en la BD); las denegaciones de las rutas de F3 las resuelve el propio dominio `portada` (implementa `audit.Recorder` con su tabla método+ruta→acción) | FR-017 + el edge case exigen atomicidad ("no aplicarse sin registro"); R2 permite SQL compartido vía consulta nombrada; el registro cerrado se mantiene en la BD y en `platform/audit` a la vez | Que `usuarios` conozca las rutas de F3 en `actionFromRoute` (acoplaría F2 a F3–F9); SQL de auditoría duplicado a mano en otro paquete con otra query (dos fuentes de la misma tabla); registrar tras el commit (rompe el edge case); un registro nuevo para contenido (FR-017 pide el de F2) |
| **P3-13** | **Marca versionada en el repo**: tokens como variables CSS + tema Tailwind (`--color-azul #1a2b4a` 70 %, `--color-turquesa #00c9a7`, `--color-blanco`, `--color-crema #F5F2EC`, `--color-naranja #ff6b3d`, `--color-verde #217638`; fuentes `Bebas Neue`/`Poppins`/`Playfair Display`), tipografías **autoalojadas** con `@fontsource/…` y logo `frontend/public/brand/simiente-logo.jpeg` con sus reglas de uso en el componente `BrandLogo` (R3-12) | FR-002 exige respetar el Manual; los tokens nombrados evitan una paleta paralela y el autoalojamiento evita depender de un CDN (público con conexiones lentas) | Google Fonts por `<link>` (tercero en cada carga); tokens solo en el PDF (el código no lee PDFs); no versionar el logo |
| **P3-14** | **Contrato**: delta en `specs/003-portada-info-general/contracts/openapi.yaml` → fusión aditiva en `backend/api/openapi.yaml` con `info.version` **0.3.0 → 0.4.0**; rutas `/api/v1/portada`, `/api/v1/media/{fileName}` y `/api/v1/admin/portada/*`; el registro de `error.code` **no cambia**; la ampliación de `AdminActionItem` (enum `action` + `targetKind: content`) se aplica en la fusión (documentada en el delta) | Contrato antes que el código (§II) sin dos copias vivas ni códigos de error ad-hoc (una subida inválida es `400 invalid`) | Prefijo `/api/v1/public/…` (no usado por F1/F2: lo público es lo que no está bajo `/admin`); nuevo `error.code` para archivos pesados (no hace falta) |
| **P3-15** | **Validación**: tag **`url`** nuevo en `platform/validate` (https + host) + reglas de dominio (hosts de WhatsApp/redes) + límites como constantes del service («quiénes somos» 1.000; colecciones ≤50 servicios y ≤20 canales) (R3-14) | FR-015 completo con el patrón de tags de F2; los límites acotan el agregado del panel (que no es un listado paginado) | Librería de URLs (dependencia); validar solo en frontend (§IV); límites solo en BD (mensajes peores para el humano) |
| **P3-16** | **Rutas SPA**: `/` pasa a ser la portada pública (`PublicLayout`) y la página «Estado del sistema» (`StatusPage`) se mantiene en **`/health`** (ruta de la SPA, distinta del endpoint de backend `/healthz`, que no se toca) (R3-15) | FR-001: la portada es la página de inicio; la pantalla de estado sigue disponible para diagnóstico/QA bajo el vocabulario operativo del proyecto | Borrar `StatusPage` (F1 la certificó y es útil); portada en `/inicio` y `/` sin nada (contradice FR-001); `/estado` (primera elección, sustituida por decisión del humano del 2026-10-09) |
| **P3-17** | **Concurrencia**: última escritura completa gana; cada guardado escribe el elemento entero en una transacción (PUT de singletons = reemplazo completo; PATCH = campos presentes en una fila) (R3-17) | Es el default de la spec; la transacción evita la "mezcla incoherente" del edge case | `version`/`If-Match` optimista (fuera de alcance; queda como ampliación si el humano lo pide) |
| **P3-18** | **Chrome del sitio público** (encabezado/pie): identidad publicada → su nombre y logo; si no, fallback **estático** (nombre del diccionario i18n + logo del repo) (R3-18) | FR-013 prohíbe exponer borradores y una portada sin marca es un error de producto; el fallback es un asset del producto, no contenido del CMS | Mostrar la identidad aunque esté en borrador (viola FR-013/SC-002); encabezado sin nombre ni logo |
| **P3-19** | **Dependencias**: 0 nuevas de runtime en backend; 3 `@fontsource/…` en frontend (R3-16) | §II: minimizadas y justificadas; las tipografías son requisito del Manual (FR-002) | `sharp`/procesamiento de imágenes (no hay derivados en el MVP); cliente S3; `react-i18next` |

## Project Structure

### Documentation (esta funcionalidad)

```text
specs/003-portada-info-general/
├── plan.md              # EDITABLE (este archivo, /speckit.plan)
├── research.md          # EDITABLE (fase 0: R3-1…R3-18)
├── data-model.md        # EDITABLE (fase 1: tablas, migraciones 000005/000006, consultas)
├── quickstart.md        # EDITABLE (fase 1: cómo probar F3 en local)
├── spec.md              # APROBADA — no se edita
├── contracts/
│   └── openapi.yaml     # EDITABLE (fase 1): delta de diseño, se fusiona en backend/api/openapi.yaml
├── ux.md                # DEL disenador-ux (corre en paralelo; este plan fija rutas, DTOs y tokens)
└── tasks.md             # Fase 2 (/speckit.tasks — NO lo crea este comando)
```

### Source Code — backend

```text
backend/
├── cmd/api/
│   ├── main.go                      # EDITADO: DI del dominio portada (repo, storage, audit de F3)
│   │                                #   + registro de /api/v1/portada, /api/v1/media y
│   │                                #   /api/v1/admin/portada (AdminChain con módulo "portada")
│   └── main_test.go                 # EDITADO: prueba de humo ampliada (rutas nuevas + /healthz intacto)
├── internal/
│   ├── db/
│   │   ├── queries/
│   │   │   ├── home.sql             # NUEVO: consultas de home_* (upsert singletons, CRUD listas,
│   │   │                            #   variantes …Published para el público)
│   │   │   └── audit.sql            # SIN CAMBIOS: se reutiliza InsertAdminAction (R3-11)
│   │   └── (generado por sqlc, commiteado)
│   ├── portada/                     # NUEVO — dominio F3 (receta de arq. §8 con archivos por responsabilidad)
│   │   ├── model.go                 # entidades (Identity, About, Contact, Service, WhatsappChannel,
│   │   │                            #   SocialLink + PublicationState) y DTOs (etiquetas validate, Es/En)
│   │   ├── repository.go            # constructor + mapRow/params (pgtype → dominio) + withTx
│   │   │                            #   + inserción de admin_actions sobre el tx (audit.Action → params)
│   │   ├── repository_home.go       # home_identity/about/contact (upsert) + services/whatsapp/socials
│   │   ├── service.go               # tipos comunes, invariantes, constantes de límite
│   │   ├── service_public.go        # portada pública: resolver idioma + fallback en→es + omitir vacías
│   │   ├── service_admin.go         # CRUD del panel, validaciones de dominio (URLs/WhatsApp/redes),
│   │   │                            #   publicar/retirar y registro de auditoría (transaccional)
│   │   ├── service_audit.go         # audit.Recorder del dominio: método+ruta → acción home.* (denegaciones)
│   │   ├── handler.go               # constructor + helpers HTTP (decodificar/validar/responder)
│   │   ├── handler_public.go        # GET /api/v1/portada
│   │   ├── handler_admin.go         # /api/v1/admin/portada/* (identidad, quiénes somos, contacto,
│   │   │                            #   horario, whatsapp, redes)
│   │   ├── handler_media.go         # GET /api/v1/media/{fileName} (descarga segura)
│   │   ├── handler_images.go        # POST /api/v1/admin/portada/imagenes (multipart)
│   │   ├── routes.go                # RegisterPublic (/api/v1/portada + /api/v1/media) y
│   │   │                            #   RegisterAdmin (/api/v1/admin/portada, AdminChain "portada")
│   │   └── *_test.go                # service con fakes · handler con httptest · repository integration
│   └── platform/
│       ├── storage/                 # NUEVO (plumbing): interfaz Store + LocalStore (Save/Open/Delete,
│       │                            #   nombres validados, sin path traversal) + *_test.go
│       ├── audit/audit.go           # EDITADO: + 14 códigos home.*, TargetKindContent y su validación
│       ├── validate/validate.go     # EDITADO: + etiqueta `url` (https + host, ≤500)
│       └── config/config.go         # EDITADO: + UPLOAD_DIR, UPLOAD_MAX_BYTES
├── migrations/
│   ├── 000005_create_home_content.up.sql / .down.sql                  # NUEVO: 6 tablas home_*
│   └── 000006_extend_admin_actions_for_home_content.up.sql / .down.sql # NUEVO: CHECK de auditoría
└── api/openapi.yaml                 # EDITADO: se fusiona el delta (info.version → 0.4.0) + make api-gen
```

### Source Code — frontend y raíz

```text
frontend/src/
├── api/
│   ├── portada.ts                   # NUEVO: getPortada(lang) + panel (get/update/CRUD) + uploadImage
│   └── schema.d.ts                  # REGENERADO (npm run api-gen)
├── i18n/                            # NUEVO (R3-9): i18n del sitio público (el panel sigue en español)
│   ├── LanguageProvider.tsx         # context: lang (es|en), setLang, t()
│   ├── messages.ts                  # tipo PublicMessageKey (claves cerradas)
│   ├── es.ts / en.ts                # diccionarios tipados (Record<PublicMessageKey, string>)
│   └── i18n.test.tsx                # exhaustividad de claves y cambio de idioma
├── app/
│   ├── router.tsx                   # EDITADO: '/' → PublicLayout + PortadaPage; StatusPage → '/health';
│   │                                #   '/panel/portada' con RequirePermission code={PORTADA}
│   ├── layout.tsx                   # EDITADO: + PublicLayout (header con BrandLogo y LanguageSwitcher,
│   │                                #   main, footer) y entrada "Portada" en el menú del panel
│   └── guards.tsx                   # SIN CAMBIOS (RequirePermission reutilizado)
├── features/
│   ├── public/                      # NUEVO: sitio público
│   │   ├── pages/PortadaPage.tsx    # ensambla las secciones presentes (SC-012: sin secciones vacías)
│   │   ├── components/              # BrandLogo, LanguageSwitcher, IdentitySection, AboutSection,
│   │   │                            #   ScheduleSection, WhatsappSection, SocialsSection, ContactSection
│   │   ├── hooks/usePortada.ts      # TanStack Query por lang (queryKey ['portada', lang])
│   │   └── *.test.tsx
│   ├── portada/                     # NUEVO: gestión del módulo en el panel (en español, FR-016)
│   │   ├── pages/PortadaAdminPage.tsx
│   │   ├── components/              # IdentityForm, AboutForm, ContactForm, ScheduleItems,
│   │   │                            #   WhatsappChannels, SocialLinks, ItemForm (compartido),
│   │   │                            #   ImageUploadField, PublicationToggle
│   │   ├── hooks/                   # usePortadaAdmin, useSaveIdentity, useSaveAbout, useSaveContact,
│   │   │                            #   useScheduleItems, useWhatsappChannels, useSocialLinks, useUploadImage
│   │   └── *.test.tsx
│   ├── status/                      # EDITADO: la página se mantiene; ahora vive en '/health'
│   └── roles/
│       └── permissions.ts           # EDITADO: isPermissionAvailable incluye 'portada' (quita el sello
│                                    #   "Disponible más adelante" del módulo)
├── lib/
│   └── permissions.ts               # EDITADO: + PORTADA = 'portada'
├── components/                      # SIN CAMBIOS salvo reutilización (inventario único de F2)
└── (index.css, tailwind.config)     # EDITADO: tokens de marca del Manual (R3-12)
frontend/e2e/
├── portada-publica.spec.ts          # NUEVO: visitante (US1/US3/US4/US5)
└── portada-panel.spec.ts            # NUEVO: edición y publicación desde el panel (US2/US3)

# Raíz
frontend/public/brand/simiente-logo.jpeg   # NUEVO: logo del Manual (recurso de marca, R3-12)
frontend/package.json                      # EDITADO: + @fontsource/bebas-neue, poppins, playfair-display
.env.example                               # EDITADO: + UPLOAD_DIR, UPLOAD_MAX_BYTES (documentados)
docker-compose.yml                         # EDITADO: volumen uploads_data + UPLOAD_DIR en el backend
README.md                                  # EDITADO por documentador al cerrar
Makefile / .github/workflows/ci.yml / .githooks/   # SIN CAMBIOS (son del kit)
```

### Inventario único de componentes UI

F3 **no añade componentes** al inventario único de F2 (los 13 componentes siguen siendo los
compartidos: `Field`, `Select`, `Button`, `Notice`, `ConfirmDialog`, `Dialog`, `EmptyState`, `Table`,
`Tabs`, `Pagination`, `StatusPill`…). Los componentes propios de F3 viven en `features/public/` y
`features/portada/` (son de producto, no de catálogo) y **reutilizan** el inventario sin duplicar
markup. Cualquier componente que acabe compartido con F4–F9 se propone primero al inventario
(regla de F2/T242).

## Cadena de middleware y grupos (orden; el primero es el más externo)

```text
Global:   request-id → recover → logging → CORS(con credenciales) → handler
          /api/v1/portada            → (sin middlewares)  → handler    ← público, solo lectura
          /api/v1/media/{fileName}   → (sin middlewares)  → handler    ← público, solo lectura
          /api/v1/admin/portada      → authn → guard de cambio de contraseña → authz("portada") → CSRF → handler
```

- La superficie **pública de F3 es de solo lectura**: no necesita `authn`, `CSRF` ni `rate-limit`
  (F2 solo lo monta sobre superficies públicas **escribibles**; aquí no hay ninguna).
- El grupo `/api/v1/admin/portada` monta **`middleware.AdminChain("portada", deps)`** — la misma
  cadena aprobada de F2 con su módulo —, de modo que la denegación responde `403 forbidden`
  genérico, **queda registrada** en `admin_actions` (`result='denied'`) y una cuenta con
  `mustChangePassword` no ve nada del módulo (guard de contraseña). El `Recorder` que recibe la
  cadena es el **del dominio `portada`** (resuelve sus propias rutas → acción `home.*`, R3-11); el de
  `usuarios` sigue atendiendo el grupo de F2 sin tocarlo.
- `/healthz` **no cambia**: sigue público, sin sesión y con `Cache-Control: no-store`.

## Contrato OpenAPI: dónde vive y cómo se mantiene

Igual que F1/F2 (reglas §8.1.5 y §8.1.10):

1. **Diseño (esta fase)**: `specs/003-portada-info-general/contracts/openapi.yaml` — delta inmutable
   una vez aprobado el plan.
2. **Fusión**: la primera tarea de implementación lo funde en `backend/api/openapi.yaml` de forma
   **aditiva**, aplica los dos cambios documentados del delta (enum `action` + `targetKind:
   content` de `AdminActionItem`) y sube `info.version` a **0.4.0**; luego `make api-gen`.
3. **Consumo**: el frontend genera tipos solo desde `backend/api/openapi.yaml`.
4. El snapshot de `specs/` no se vuelve a editar.

## Estrategia de pruebas

Cobertura exigida: **80 % en `service/`** (§III). Comandos: `go test ./...` ·
`go test -tags=integration ./...` · `npm test -- --run` · `make e2e` · `make ci`.

| Capa | Tipo | Qué verifica F3 | Dónde |
|---|---|---|---|
| `platform/storage` | Unitaria | `Save` con nombre generado único, `Open` solo con nombres que cumplen la regex (rechaza `../`, absolutos y nombres de cliente), `Delete` idempotente, errores envueltos | `internal/platform/storage/*_test.go` |
| `platform/validate` | Unitaria (tabla de casos) | Nueva etiqueta `url`: https + host, límite 500, rechaza `http://`, `javascript:`, URLs sin host | `internal/platform/validate/*_test.go` |
| `platform/audit` | Unitaria | Los 14 códigos `home.*` en el registro cerrado; `TargetKindContent` exige `TargetLabel` y ambas FK en nil | `internal/platform/audit/*_test.go` |
| `service` (dominio) | Unitaria con **fakes** | **Público**: fallback `en → es` por campo (nunca vacío, SC-006), solo `published`, secciones vacías omitidas (SC-012), `lang` inválido → `invalid`, orden de listas. **Panel**: validación completa de FR-015 (obligatorios, español presente, 1.000 caracteres, correo/teléfono, URLs de WhatsApp y de red con host correcto, red fuera de catálogo, duplicados → `conflict`), límites de colección, publicar/retirar por elemento, "se publica al guardar" (FR-014), normas de imagen (alt obligatorio con imagen, quitar imagen). **Auditoría**: cada mutación registra su acción con el código y `targetLabel` correctos, en la misma transacción que la mutación (si el registro falla, la mutación no se aplica); denegaciones y rechazos registrados best-effort | `internal/portada/service_*_test.go` |
| `handler` | Unitaria con `httptest` + service falso | Cada operación: decodificación, `details` por campo en 400, códigos 200/201/204/400/401/403/404/409, sobres uniformes, **el DTO público nunca contiene `publicationState` ni campos `En` sin resolver**; multipart: sin archivo, archivo gigante (400 con `details.file`), extensión no permitida | `internal/portada/handler_*_test.go` |
| `repository` | **Integración** (PostgreSQL real, testcontainers si no hay `DATABASE_URL_TEST`) | SQL real: upsert de singletons (dos concurrentes → una fila), `UNIQUE (network)` → conflict, `UNIQUE (kind, destination, name_es)` solo en el duplicado exacto, `CHECK` de `publication_state`/día/hora/longitudes, consultas `…Published` sin borradores, **mutación + `admin_actions` atómicas** (y rollback si falla el registro), `admin_actions` con `target_kind='content'` y su `CHECK` de coherencia, migración `000006` up→down→up | `internal/portada/repository_*_test.go` (`//go:build integration`) |
| `cmd/api` | Humo de composición | Rutas nuevas publicadas con su sobre, `/healthz` intacto y `404` `not_found` en rutas inexistentes (§8.1.9) | `cmd/api/main_test.go` |
| Frontend | Unitaria (Vitest + MSW) | Portada: secciones presentes/ausentes, fallback en→es visible, cambio de idioma inmediato y persistente entre visitas (`localStorage`, primera visita en español), enlaces de WhatsApp/red con un clic, `alt` en imágenes; panel: formularios con errores por campo, subida de imagen con error de tipo/tamaño, publicar/retirar por elemento (incluidos los singletons), entrada de menú y ruta solo con permiso `portada`, sello "Disponible más adelante" retirado del módulo; i18n: claves completas en ambos idiomas | `frontend/src/features/*/*.test.tsx`, `frontend/src/i18n/i18n.test.tsx` |
| E2E pública | Playwright (local) | Recorrido de `quickstart.md`: visitante sin sesión ve todo lo publicado, nada de borradores, cambia a inglés (interfaz + contenidos con fallback), vuelve a español, abre WhatsApp/red, recorre con teclado y en 3 anchos de pantalla (SC-001, SC-002, SC-006, SC-007, SC-009, SC-012) | `frontend/e2e/portada-publica.spec.ts` |
| E2E panel | Playwright (local) | Recorrido de `quickstart.md`: login con permiso, editar identidad/quiénes somos/contacto/horario/WhatsApp/red, subir logo, publicar/retirar por elemento, ver el cambio en la portada pública **desde la primera carga** (SC-003, SC-004), denegación total con cuenta sin permiso (SC-005, SC-011) y registro de todo en auditoría (SC-013) | `frontend/e2e/portada-panel.spec.ts` |

**Pruebas que no faltan** (las que la spec hace críticas): **0 borradores expuestos** por ninguna
respuesta pública (SC-002: además del DTO, ninguna ruta pública devuelve `publicationState` ni
textos `En` sin resolver); **publicar/retirar por elemento** y sección vacía oculta (SC-012);
**cambio guardado visible en la primera carga** (SC-003: `no-store` + e2e); **fallback en→es sin
campos vacíos** (SC-006: tabla de casos por campo en el service + e2e en inglés); **toda operación
del módulo verifica el permiso** también por API forzada (SC-005/SC-011, igual que F2); **toda edición
queda registrada** —éxito, fallo y denegación— en `admin_actions` y el registro sigue siendo de solo
lectura (SC-013); **texto con tildes, `ñ`, `¿¡` y emojis intacto** (edge case); **archivo inválido no
rompe la portada** (edge case de imágenes).

## Cobertura de requisitos (US1–US5 y FR-001…FR-019)

| Historia | Dónde se cubre | Verificación |
|---|---|---|
| **US1** — Ver la portada con la información general (P1) | `GET /api/v1/portada` (P3-3) + `features/public` (P3-10/P3-13/P3-16) + `/api/v1/media/{fileName}` (P3-9) | quickstart §1 · e2e pública (SC-001, SC-009) |
| **US2** — Editar la información general desde el panel (P1) | `/api/v1/admin/portada/*` (P3-11/P3-14) + `features/portada` con validación de FR-015 (P3-15) + imágenes (P3-9) | quickstart §2–§6 · e2e panel (SC-004, SC-005) |
| **US3** — Controlar la publicación por elemento (P2) | `publication_state` + consultas `…Published` (P3-4) + `PublicationToggle`/`PATCH` con `publicationState` | quickstart §5 · e2e panel (SC-002, SC-003, SC-012) |
| **US4** — Ver la portada en inglés (P2) | `?lang=` con fallback en el servidor (P3-3) + `i18n/` con diccionarios tipados (P3-10) | quickstart §7 · e2e pública (SC-006) |
| **US5** — Navegar en cualquier dispositivo y de forma accesible (P2) | Layout público responsivo + reglas de accesibilidad de la skill `react-frontend` (teclado, `alt`, foco, ≥44 px, contraste de tokens del Manual) | quickstart §10 (checklist manual por dispositivo/teclado/lector) · e2e (SC-007, SC-008) |

| FR | Dónde se resuelve en este plan | Verificación |
|---|---|---|
| FR-001 | P3-16 (`/` = portada pública) + `GET /api/v1/portada` (P3-3/P3-4) | quickstart §1 · e2e pública (SC-001) |
| FR-002 | P3-13 (tokens + `BrandLogo` con reglas del logo) + identidad gestionable (P3-5, `PUT …/identidad`) | quickstart §3 y §6 · revisión de `BrandLogo`/tema por `revisor-codigo` + `ux.md` |
| FR-003 | `home_about.text_es` ≤1.000 (P3-2/P3-15) + texto plano sin ejecución (§IV) | quickstart §3 · pruebas del service/handler (límite, tildes/emojis) |
| FR-004 | `home_services` (P3-6) + `ScheduleItem*` del contrato | quickstart §4 · pruebas de service/repository |
| FR-005 | `home_whatsapp_channels` (P3-7) + `WhatsappChannelPublic.url` listo para abrir | quickstart §4 · e2e pública (SC-009) |
| FR-006 | `home_social_links` con catálogo + `UNIQUE (network)` (P3-8) | quickstart §4 · pruebas de service/repository (SC-009) |
| FR-007 | `home_contact` (P3-5) + `ContactPublic` | quickstart §1 y §3 |
| FR-008 | P3-2 (columnas `*_es`/`*_en`) + `i18n/` (P3-10) | quickstart §7 · pruebas de i18n (SC-006) |
| FR-009 | P3-3 (fallback `en → es` por campo en el service) | quickstart §7 · tabla de casos del service + e2e en inglés (SC-006) |
| FR-010 | P3-10 (`LanguageSwitcher`, `localStorage`, primera visita sin preferencia → español) | quickstart §7 · e2e pública (cambio inmediato + persistencia entre visitas) |
| FR-011 | `GET/PUT/PATCH/POST/DELETE /api/v1/admin/portada/*` (P3-14) + `features/portada` | quickstart §2–§6 · e2e panel (SC-004) |
| FR-012 | P3-11 (`AdminChain("portada")` en servidor + guard `RequirePermission`) | quickstart §8 · pruebas de handler (403 por API forzada) · e2e (SC-005, SC-011) |
| FR-013 | P3-4 (`publication_state` por elemento —también cada singleton: identidad, «quiénes somos» y contacto publican por separado—, consultas `…Published`, secciones vacías omitidas) | quickstart §5 · pruebas de service/repository ("0 borradores expuestos") · e2e (SC-002, SC-012) |
| FR-014 | P3-4/P3-17 (el guardado de un elemento publicado es visible en la primera carga; `no-store`) | quickstart §5 · e2e panel (SC-003) |
| FR-015 | P3-15 (validación completa: obligatorios, es obligatorio, 1.000, correo/teléfono F2, URLs, catálogo y duplicados de red, enlaces de WhatsApp) + `platform/validate` `url` | quickstart §3 y §4 · pruebas de service/handler (tabla de casos por campo, `details`) |
| FR-016 | Panel en español (convención de F2; solo el sitio público lleva `i18n/`) | revisión de `revisor-codigo` (0 cadenas de panel pasadas por `t()`) |
| FR-017 | P3-12 (`admin_actions` transaccional + 14 códigos `home.*` + `target_kind='content'` + denegaciones registradas) | quickstart §9 · pruebas de service/repository/integración (SC-013) |
| FR-018 | `PublicLayout` responsivo (P3-13/P3-16): móvil/tableta/escritorio, títulos claros, toc-targets ≥44 px | quickstart §10 · e2e en 3 anchos (SC-007) |
| FR-019 | Accesibilidad: teclado, semántica, foco visible, contraste de la paleta del Manual, `logoAlt`/`coverImageAlt` obligatorios con imagen (P3-9/`data-model.md`) | quickstart §6 y §10 (checklist WCAG 2.1 AA) · e2e de navegación por teclado (SC-008) |

## Métricas del plan (coherentes con los SC de la spec)

| SC | Qué se mide | Cómo se mide en F3 | Umbral |
|---|---|---|---|
| SC-001 | Todo visitante ve la información general sin cuenta y la encuentra en <30 s | e2e pública: la portada muestra identidad, quiénes somos, horario, WhatsApp, redes y contacto en la primera carga | 100 % |
| SC-002 | 0 borradores visibles al público | Pruebas "0 borradores expuestos" (service + handler + repository) y e2e con elementos en draft | 0 |
| SC-003 | Cambio guardado visible desde la primera carga posterior | e2e panel: editar → recargar portada → ver el dato (respaldo: `Cache-Control: no-store`) | 100 % |
| SC-004 | Actualizar y publicar un dato en <2 min desde el panel | e2e/observación (quickstart §3–§5) | < 2 min |
| SC-005 | Toda operación del módulo verifica el permiso | Pruebas de handler por operación (403) + e2e con cuenta sin permiso | 0 operaciones sin permiso |
| SC-006 | Con `lang=en`: 100 % de la interfaz traducida y 0 campos vacíos (fallback a es) | Pruebas de i18n (claves tipadas) + tabla de casos de fallback + e2e en inglés | 100 % / 0 |
| SC-007 | Usable en teléfono, tableta y computadora sin desplazamiento horizontal | e2e en 3 anchos (360/768/1280) + checklist manual (quickstart §10) | 0 pérdida / 0 scroll horizontal |
| SC-008 | Accesibilidad: 0 hallazgos críticos (referencia WCAG 2.1 AA) | Checklist manual de quickstart §10 (teclado, lector, contraste, zoom) revisado por QA | 0 críticos |
| SC-009 | WhatsApp y redes se abren con un clic al destino correcto | e2e pública (hrefs construidos: `wa.me`/URL de grupo/red) | 100 % |
| SC-010 | Personas de distintas edades encuentran horario y contacto sin ayuda | Prueba con personas (fuera de la automatización; se documenta el resultado en el cierre) | ≥ 90 % |
| SC-011 | Quien no tiene permiso no ve ni modifica nada | e2e con cuenta sin permiso (UI oculta + API forzada 403) | 100 % de intentos denegados |
| SC-012 | 0 secciones vacías ni campos en blanco | El público solo recibe secciones con elementos publicados (pruebas de contrato) + e2e | 0 |
| SC-013 | 100 % de ediciones registradas en la auditoría de F2; 0 registros editables/borrables | quickstart §9 + pruebas transaccionales + revisión del contrato (sin escritura sobre el registro) | 100 % / 0 |

## Riesgos

> Numeración **RG3-1…RG3-12**, propia de este plan (para no colisionar con RG1–RG19 de F2 ni con
> R3-1…R3-18 de `research.md`).

| Riesgo | Impacto | Mitigación |
|---|---|---|
| **RG3-1 — Imágenes en disco local sin política de copia** | Un `docker compose down -v` o una pérdida del host borra logos/portadas (el texto está en PostgreSQL) | Volumen con nombre `uploads_data` (como `pgdata`); `UPLOAD_DIR` documentado en `.env.example` y en el quickstart (cómo copiar los archivos); el plan B (S3/CDN) está cerrado en R3-8 detrás de la interfaz `storage.Store`. Los archivos huérfanos se limpian al reemplazar/borrar (best-effort con log) |
| **RG3-2 — `000006` referencia restricciones con nombre generado** (`admin_actions_check1`) | Si el nombre real difiere, la migración falla (fail-closed: nada queda a medias) | Los nombres son deterministas (migraciones inmutables; orden de definición de 000004) y la prueba de integración hace **up → down → up** de `000006`; alternativa documentada (bloque `DO` que localiza la restricción por definición) si la prueba demostrara lo contrario |
| **RG3-3 — Falta una traducción de la interfaz** (SC-006) | Texto en español dentro del sitio en inglés o hueco visible | Diccionarios con **la misma clave tipada** (`Record<PublicMessageKey, string>`): sin la clave, TypeScript no compila; la prueba de i18n recorre ambas listas; los contenidos con fallback los garantiza el servidor (P3-3) |
| **RG3-4 — `ux.md` del disenador-ux se produce en paralelo** | Las tareas de frontend podrían diseñar pantallas que no encajan con los DTOs/rutas de este plan | Este plan fija rutas, DTOs, tokens de marca, secciones y reglas de contenido: el diseño puede variar la composición **sin cambiar la API**; las tareas `[frontend]` de `tasks.md` quedarán marcadas como dependientes de `ux.md`, igual que en F2 (RG2) |
| **RG3-5 — Patrón de columnas `es/en` si algún día entra un tercer idioma** | Migración a tabla hija por idioma en todas las tablas de contenido | Decisión asumida y documentada (R3-1): hoy son exactamente 2 idiomas (Decisión 8) y la API pública ya esconde el esquema (resuelve idioma en el servidor), de modo que el cambio no rompe a los visitantes |
| **RG3-6 — Contenido como vector XSS** (textos con HTML, imágenes SVG) | Ejecución de script en el visitante (CWE-79) | Texto plano: se renderiza como texto (sin `dangerouslySetInnerHTML` en ninguna vista de F3) y el backend no lo transforma; imágenes solo JPEG/PNG/WebP por firma binaria (sin SVG) con `nosniff`; validación en backend aunque el cliente valide; revisión de `seguridad` sobre las vistas y la descarga |
| **RG3-7 — Subidas abusivas** (archivos gigantes o masivos) | Disco lleno / CPU | Tope `UPLOAD_MAX_BYTES` (8 MB) con `MaxBytesReader` (rechaza en stream), tipos cerrados, superficie de subida **solo autenticada con permiso** (no hay subida pública), huérfanos limpiados; `rate-limit` no aplica (superficie de panel ya protegida) — se reabre si el volumen lo pide |
| **RG3-8 — Un borrador se filtra por una vía nueva** (una consulta olvidada, un DTO futuro) | Incumplimiento directo de FR-013/SC-002, el requisito más sensible de la spec | Consultas **`…Published` separadas** (la única función que ve borradores es la del panel, P3-4), DTO público **sin** `publicationState` ni campos sin resolver, pruebas dedicadas "0 borradores expuestos" en las tres capas y e2e con borradores presentes en BD |
| **RG3-9 — Caché del navegador al retrasar un cambio publicado** (SC-003) | El visitante ve datos viejos tras un guardado | `Cache-Control: no-store` en `GET /api/v1/portada` (P3-14) y prueba e2e de "editar → recargar → ver el cambio"; las imágenes sí se cachean pero su nombre cambia en cada subida (immutable seguro) |
| **RG3-10 — Accesibilidad (SC-008) no cubierta por automatización** | Hallazgos críticos en la revisión manual | Checklist explícita en `quickstart.md` §10 (teclado, lector de pantalla, contraste, zoom, toc-targets) ejecutada por `qa-tester` en la validación; componentes semánticos y `alt` obligatorio con imagen en la BD (`CHECK`); WCAG 2.1 AA como referencia |
| **RG3-11 — Mover `StatusPage` de `/` a `/health` rompe referencias** (e2e de F2, enlaces) | Pruebas e2e rojas o enlaces muertos | La tarea de rutas actualiza los e2e existentes y `router.test.tsx`; quickstart y documentación citan `/health` (ruta de la SPA; no confundir con el endpoint `/healthz`, que sigue siendo el chequeo operativo y no depende de la SPA) |
| **RG3-12 — Nuevas dependencias de frontend (`@fontsource` ×3)** | Superficie de `npm audit` | Solo assets de build (sin runtime ni llamadas de red); `npm audit` sin altas/críticas en `make ci` (§IV); versiones fijadas en `package.json` |

## Complexity Tracking

> Sin violaciones de la constitución. Desviaciones de **convenciones internas**, declaradas:

1. **Nombres de dominio y tablas** (P3-1): la receta usa carpetas de dominio en español
   (`status`, `usuarios`) y tablas en inglés (`users`, `admin_actions`); F3 sigue la misma pauta:
   dominio `internal/portada/` y tablas `home_*`. Si `revisor-codigo` prefiere `internal/home/`, es un
   renombrado sin consecuencias (todo el cableado está en `main.go`).
2. **Tablas singleton en singular** (`home_identity`, `home_about`, `home_contact`): la skill
   `postgres-db` pide nombres en plural, pero describen **una** fila única (patrón R3-4); las
   colecciones sí van en plural (`home_services`, `home_whatsapp_channels`, `home_social_links`).
3. **El agregado `GET /api/v1/admin/portada` no pagina** (§8.1.3): es un **documento** de estado del
   módulo (singleton + colecciones), no un listado de crecimiento abierto; sus colecciones van
   acotadas por constantes del service (≤50 servicios, ≤20 canales; redes ≤ tamaño del catálogo) en
   sobres `{items}` (§8.1.2). Si alguna colección creciera más allá, se le añade `limit`/`offset` con
   `platform/paginate` en esa misma operación.
4. **Mapeo `audit.Action` → `InsertAdminActionParams` repetido** (R3-11): ~10 líneas en el repository
   de `portada` y otras tantas en el de `usuarios`, ambas sobre la **misma** consulta nombrada de
   `internal/db`. Alternativas descartadas: subir el mapeo a `platform/` (arrastraría `pgtype` a
   `platform`, fuera del diagrama de dependencias) o que un dominio importe a otro (prohibido por R2).
5. **Respuesta binaria en `GET /api/v1/media/{fileName}`** (R3-13): excepción documentada al sobre
   uniforme "2xx → DTO"; los errores de esa ruta siguen usando `ErrorEnvelope`.
6. **Cadena de panel con módulo por subgrupo**: F2 monta `AdminChain` una vez para su módulo; F3 crea
   su **propio** grupo `/api/v1/admin/portada` con `AdminChain("portada", …)` (F4–F9 repetirán el
   patrón). Ninguna regla de orden cambia: sigue `authn → guard → authz → CSRF`.
