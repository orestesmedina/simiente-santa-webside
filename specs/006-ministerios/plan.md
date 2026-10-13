# Implementation Plan: Ministerios (F6)

**Branch**: `006-ministerios` | **Date**: 2026-10-13 | **Spec**: [spec.md](./spec.md)
(aprobada **v2** por el humano el 2026-10-13: «Re-apruebo la spec v2», con Q1–Q9 resueltas —
encargado opcional con contacto opcional y ficha de visibilidad configurable, enlace opcional,
horario y lugar como textos libres, bloque de la portada con tope 5—; 6 US, 29 FR, 17 SC).

**Puerta 2 resuelta por el humano el 2026-10-13** (plan realineado en esa fecha): `leader_visible`
nace **`DEFAULT FALSE`** —«**Oculto por defecto**»: la ficha del encargado no se publica hasta que
el equipo la enciende— (P6-3) y **`/ministerios` sin nada publicado redirige a la portada** (P6-21,
**distinto a propósito** de F4 en `/eventos`, que muestra aviso sin redirigir). El resto del plan
quedó **aprobado tal como está** y los defaults revisables listados al final fueron aceptados.

**Input**: `specs/006-ministerios/spec.md` (fuente de verdad) · `.specify/memory/constitution.md`
· `docs/tecnico/arquitectura.md` (reglas R1–R8 y convenciones §8.1) y `docs/tecnico/decisiones.md` ·
**F2 terminada** (`specs/002-acceso-gestion-usuarios/`: patrón de permisos por módulo, sesión,
CSRF, `authz`, `apperr`, `validate`, `paginate`, auditoría y catálogo de permisos **con
`ministerios` ya sembrado** — verificado en `000002_create_roles_and_permissions`:
`('ministerios', 'Ministerios')`; el catálogo **no cambia** en F6) · **F3 terminada**
(`specs/003-portada-info-general/`: patrón bilingüe, publicación por elemento, imágenes en disco
local, i18n del público, ocultar secciones vacías — se usan como plantilla **y** como código
existente) · **F4 terminada** (`specs/004-eventos-actividades/`: patrón de módulo de contenido con
dominio propio, `AdminChain` por permiso, bloque de portada con endpoint propio, `media/`
promovido y `sort_order`) · **F10 terminada** (`specs/010-sistema-diseno-movimiento/`: sistema de
diseño, modo oscuro, movimiento, navegación y selector de idioma — se **aplica**, no se redefine) ·
código real de `backend/` y `frontend/` · skills `go-backend`, `postgres-db`, `react-frontend`.

## Summary

F6 es la **sección de áreas de servicio** del sitio público: un **ministerio** (alabanza, sonido,
ujieres, acción social, escuela dominical…) con su **encargado opcional** —nombre y **un** medio de
contacto opcional (celular **o** correo), cuya **ficha entera (nombre y contacto)** se publica o no
según un interruptor del equipo (Q1/Q8/Q9: FR-025/FR-026) **que nace apagado —«Oculto por defecto»,
decisión del humano 2026-10-13: privacidad por defecto—**, su **enlace opcional** al grupo de
WhatsApp para unirse (Q2, cualquier `https` válido — Q3), **horario y lugar opcionales como textos
libres** (Q6: sin fechas ni recurrencia — el modelo de F4 no aplica —), **imagen opcional** (Q4),
**orden manual** (`sort_order`, `id` — Q5) y estados borrador/publicado (Decisión 4). La sección
pública es **`/ministerios`** con una única entrada de navegación «Ministerios» (FR-006), bilingüe
con fallback al español (Decisiones 6/8), nacida con el sistema de F10 y **oculta por completo sin
nada publicado** (FR-005/SC-010): la entrada de navegación desaparece y la **URL directa
`/ministerios` redirige a la portada** (P6-21, decisión del cliente distinta a propósito de F4). Además la **portada** suma su bloque de ministerios —hasta **5**
ítems en tarjetas con imagen, en el orden manual, con botón a `/ministerios`— **inmediatamente
después del bloque «Lo que viene» de F4 y antes de «quiénes somos»** (Q7, FR-027…FR-029), servido
por su propio endpoint público acotado y **oculto por completo sin publicados** (a diferencia del
bloque de F4, que muestra un aviso).

Técnicamente F6 añade: **una sola tabla `ministries`** (el encargado son campos del ministerio: no
hay directorio de personas ni fichas compartidas, R6-1), la **garantía de privacidad a nivel de
DTO público y de consulta** —con `leaderVisible` apagado ni el nombre ni el contacto del encargado
salen del servidor hacia ninguna vía pública (FR-026/SC-014: el DTO público no tiene campos para el
dato oculto y las consultas públicas devuelven `NULL` con `CASE WHEN leader_visible`; el bloque de
la portada ni siquiera tiene campos de encargado)—, la validación del enlace con el tag `url`
existente (**cualquier `https` seguro, sin dominios**; FR-015/Q3), el patrón de autorización de F2
con `AdminChain("ministerios")` (**sin permisos nuevos ni migraciones de permisos**: el permiso ya
está sembrado por `000002`), 8 códigos de auditoría `ministry.*` (migración `000011`, incluidos los
dos de la visibilidad de la ficha del encargado) y la **reutilización literal** del servicio de
imágenes de F3/F4 (`platform/storage` + `internal/media`, con `IsFilePublished` ampliado). El
frontend publica **`/ministerios`** dentro de F10, añade `/panel/ministerios` bajo
`RequirePermission(MINISTERIOS)`, la entrada de navegación condicional (fuente única
`useSiteLinks`) y el bloque `MinisteriosBlock` en `HomePage` **como una inserción** (el resto de la
portada de F3 no se toca, FR-021/FR-029).

**Ninguna dependencia nueva** ni en backend ni en frontend.

## Technical Context

**Language/Version**: Go 1.27 (sin cambios) · TypeScript 5.x `strict` sobre Node 22 · React 19.

**Primary Dependencies**: **ninguna nueva**. Se reutiliza todo lo existente: `net/http` (multipart
de la subida de imágenes), `net/url` (validación del enlace), `pgx`/`sqlc`, `testcontainers-go`; en
frontend React Router, TanStack Query, React Hook Form + Zod, Tailwind, Vitest + Testing Library +
MSW y Playwright (research R6-17).

**Storage**: **PostgreSQL 16** (tabla `ministries` + ampliación del `CHECK` de `admin_actions`;
migraciones `000010`/`000011`, ver [data-model.md](./data-model.md)) · Redis 7 **sin tocar** (sesión
de F2) · **disco local del backend** reutilizado para imágenes (`platform/storage`, volumen
`uploads_data` de F3). Capa de datos sqlc (D-A3): consultas en
`backend/internal/db/queries/ministries.sql` y la rama nueva de `media.sql`, código generado
commiteado; la auditoría reutiliza `InsertAdminAction` de `queries/audit.sql` **dentro de la
transacción** de cada mutación. **Sin variables de entorno nuevas** (F4 ya añadió `APP_TZ`; F6 no
calcula fechas).

**Testing**: estrategia por capa de F1/F2/F3/F4 ampliada (sección propia): `go test` (unitarias),
`go test -tags=integration` (repository contra PostgreSQL real —incluida la migración up/down/up de
`000010`/`000011`, los `CHECK` de privacidad/contacto/enlace y la regla "0 apariciones del
encargado oculto"—), Vitest + Testing Library + MSW y Playwright (e2e pública, de panel, de
accesibilidad y de la portada). **Sin excepciones aprobadas**: F6 no toca pruebas de F2 en el
backend (el catálogo de permisos **no** cambia); solo actualiza las 4 pruebas de frontend que fijan
la lista de permisos y accesos del panel (RG6-6, mismo gesto que F4 hizo con las suyas).

**Target Platform**: Docker Compose local + GitHub Actions (CI del kit, sin tocar). Sin despliegue a
producción (fuera de alcance de F6).

**Performance Goals**: `GET /api/v1/ministerios` < 500 ms con BD sana (una consulta, JSON con
descripciones, `no-store`, acotada a `maxPublicItems = 200`) · `GET /api/v1/portada/ministerios` <
500 ms y **payload acotado** (≤5 ítems con `id`/`name`/`image`, `LIMIT 5` en SQL) · listados de
panel paginados (`limit`/`offset`, §8.1.3) · subida de imagen ≤ 8 MB validada en streaming (reglas
de F3 sin cambios) · descarga de imágenes con la misma política `no-store` de F3/F4.

**Constraints**: stack fijo (constitución) · capas `handler → service → repository` (§II) · contrato
OpenAPI antes que el código (§II) · migraciones versionadas e inmutables (§VI) · validación de toda
entrada en backend (§IV, CWE-20) · SQL siempre parametrizado (§IV, CWE-89) · nada de contenido
ejecutable (§IV, CWE-79: sin `dangerouslySetInnerHTML`, sin SVG en imágenes) · authn/authz/CSRF en
servidor (§IV, CWE-862) · **privacidad del encargado verificada en el servidor** (FR-026: la regla
vive en el DTO público y en la consulta, no en la interfaz) · dependencias minimizadas (0 nuevas) ·
archivos del kit no editables (regla 10) · **la spec no se reabre**: Q1–Q9 están cerradas y el
humano confirmó además la lectura de Q6 (horario como texto libre) y el tope 5 del bloque; los dos
puntos que la puerta 2 dejó abiertos quedaron **resueltos por el humano el 2026-10-13** y así se
implementan: **default del interruptor `leaderVisible` = `FALSE`** («Oculto por defecto», R6-3) y
**`/ministerios` sin publicados redirige a la portada** (R6-13/P6-21, distinto a propósito de F4).

**Scale/Scope**: 8 operaciones REST (**2 públicas** · 6 de panel) sobre **1 tabla nueva** + la
ampliación del `CHECK` de `admin_actions` (8 códigos); 1 dominio backend nuevo (`ministerios`) y
**`media`/`platform` sin cambios de comportamiento** (solo la rama de `IsFilePublished`); 2
migraciones (`000010`, `000011`) y **ninguna de permisos**; frontend: página pública nueva en
`/ministerios` + **bloque de ministerios en la portada** + feature de panel `ministerios` + ajustes
de router/navegación/permisos/i18n. Sin puertas de aprobación abiertas más allá de **este plan**
(regla 1 de AGENTS.md).

## Constitution Check

*GATE: debe pasar antes de la investigación y re-verificarse tras el diseño.*

| Principio | Evaluación | Resultado |
|---|---|---|
| §I La spec manda | El plan implementa FR-001…FR-029 e US1–US6 sin interpretarlos; cada uno tiene fila en "Cobertura de requisitos". Q1–Q9 están cerradas; la única lectura señalada por la propia spec (Q6 como texto libre) la **confirmó el humano** el 2026-10-13 y se implementa así. Nada de Out of Scope se construye (sin F5, sin fechas/calendario, sin inscripciones, sin verificación de enlaces, sin directorio de encargados, sin traducción automática, sin historial ni papelera, sin bilingüismo del panel, sin rediseñar la portada de F3 ni F10B) y el permiso `ministerios` solo se **activa** (FR-011). Los dos puntos que se elevaban a la puerta 2 quedaron **resueltos por el humano el 2026-10-13** e implementados literalmente: `leaderVisible` **«Oculto por defecto»** (P6-3) y **`/ministerios` sin publicados redirige a la portada** (P6-21, distinto a propósito de F4). | ✅ |
| §II Arquitectura | Monorepo; capas `handler → service → repository` sobre `internal/platform/` (R1–R8); API REST JSON documentada en `backend/api/openapi.yaml` con el delta redactado **antes** que el código (`contracts/openapi.yaml`, §8.1.5); **0 dependencias nuevas** (R6-17). | ✅ |
| §III Pruebas | Toda tarea llevará sus pruebas (se exigirá en `tasks.md`); cobertura ≥80 % en `service/`; repository contra PostgreSQL real (`//go:build integration`), incluidas las migraciones `000010`/`000011` (up/down/up), los `CHECK` de contacto/enlace/español por campo, **el default «Oculto por defecto»** (`leader_visible DEFAULT FALSE`) y **la regla de privacidad** (`leader_visible = FALSE` → la fila pública sale con `NULL`); frontend con Vitest + Testing Library + MSW; flujos críticos con Playwright (sección pública, ficha oculta, sin enlace, publicar/retirar, permisos, bloque de la portada y **redirección a la portada sin publicados**). Las 4 pruebas de frontend que fijan permisos/accesos se actualizan de forma acotada y documentada (RG6-6). | ✅ |
| §IV Seguridad | SQL solo parametrizado vía sqlc (CWE-89); validación de toda entrada en backend (CWE-20); texto plano sin ejecución (CWE-79: sin `dangerouslySetInnerHTML`, sin SVG); enlace solo `https` (FR-015: sin `javascript:`/`data:`/`vbscript:`, con `CHECK` en la BD); imágenes con firma binaria, tipos cerrados y descarga solo de contenido publicado (CWE-79/434); operaciones de panel con sesión + guard + CSRF + permiso **`ministerios`** verificados en servidor (CWE-862); **privacidad del encargado aplicada en el servidor** (FR-026/SC-014: DTO público y consulta pública sin el dato oculto, inventario de vías de salida para `seguridad`); secretos: ninguno nuevo; `govulncheck`/`npm audit` en el CI del kit. Borradores **nunca** hacia el público (consultas `…Published`). | ✅ |
| §V Calidad | `gofmt`/`go vet`/`golangci-lint` sin errores; TS `strict` sin `any`; errores envueltos con `%w` y traducidos solo en `WriteError`; nombres de código en inglés (dominio `ministerios` sigue la pauta de F2/F3/F4: carpeta de dominio + tablas en inglés). | ✅ |
| §VI Base de datos | 2 migraciones versionadas con `up`/`down` completos (`000010`, `000011`); ninguna migración existente se modifica (**ni `000002`**: no hay permisos nuevos); tabla con `id`, `created_at`, `updated_at`; índices explícitos; `CHECK`/`UNIQUE` en la base (español por campo, contacto excluyente, alt con imagen, enlace `https`, longitudes, estado). | ✅ |
| §VII Observabilidad | Logs `log/slog` con `request_id` (incluidos los fallos de validación, sin cuerpos de petición); auditoría duradera ampliada a F6 en `admin_actions` (FR-017, 8 códigos); `/healthz` intacto y sin sesión; config por variables de entorno (**sin variables nuevas**); `docker compose up` levanta todo (sin servicios nuevos). | ✅ |
| §VIII Gobierno | **Este plan requiere aprobación humana** antes de `/speckit.tasks` y de cualquier código (regla 1), incluidos los puntos señalados al final. Sin despliegue a producción. Quien escribe no aprueba: todo pasa por `qa-tester`, `revisor-codigo` y `seguridad`. | ✅ |

**Sin violaciones que justificar** → "Complexity Tracking" solo registra desviaciones de
convenciones internas (naming de la tabla, presencia de navegación desde el bloque y el mapeo de
auditoría repetido).

## Decisiones técnicas (resumen; análisis completo en `research.md`)

Numeración **P6-1…P6-21**, propia de este plan (no confundir con P4-1…P4-20 de F4, P3-1…P3-19 de
F3, R6-1…R6-17 de `research.md` ni RG6-1… de "Riesgos"). **P6-21 es nueva** (siguiente número
libre): se añadió al realinear el plan con las decisiones de la puerta 2 del 2026-10-13; P6-1…P6-20
conservan su identificador y solo P6-3 y P6-15 cambiaron de contenido (las dos decisiones del
humano).

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| **P6-1** | **Un dominio `internal/ministerios/`** con **una tabla `ministries`** (encargado como columnas) y consultas en `queries/ministries.sql` | Key Entities manda: el encargado es "un campo del ministerio… sin ficha compartida"; un solo agregado no justifica más tablas (R6-1) | Tabla `leaders` con FK (compartiría fichas entre ministerios); EAV/traducciones (rompe R3-1); JSON (pierde `CHECK` y sqlc) |
| **P6-2** | **Privacidad en DTO + consulta** (R6-2): `MinisterioPublico` solo tiene `leader` si `leaderVisible`; las consultas públicas devuelven `NULL` con `CASE WHEN leader_visible`; `ItemMinisterioPortada` no tiene campos de encargado | FR-026 exige que "no salga del servidor": una garantía verificable por inspección de red/HTML y testeable como "0 apariciones" | Ocultar solo en React (FR-026 lo prohíbe literalmente); borrar el dato en el handler con un `if` (frágil ante cualquier DTO nuevo); dos endpoints públicos (más superficie que filtrar) |
| **P6-3** | **Interruptor `leaderVisible`** (columna `leader_visible`), equivalente al `showContact` de la spec; gobierna **solo la publicación** (dato guardado ≠ dato publicado), audita con `ministry.leader.show`/`hide` y **nace apagado —«Oculto por defecto»—** (`leader_visible DEFAULT FALSE`, `default: false` en `MinisterioInput`, interruptor apagado al crear) | Q9: oculta **nombre y contacto**; un nombre honesto evita leer "solo contacto" (FR-025). El default **`FALSE`** lo decidió el humano el 2026-10-13: **privacidad por defecto** — publicar un celular/correo personal exige un acto afirmativo —, aunque la recomendación de `research.md` R6-3 era la contraria | Mantener `showContact` (engañoso desde Q9); **`DEFAULT TRUE`** (recomendación (a) de R6-3, **descartada por el humano**); sin default / elección obligatoria (fricción sin ganancia); borrar el contacto al ocultar (contradice US2 esc. 11: "vuelve a publicarse sin reescribir nada") |
| **P6-4** | **Contacto = `leader_phone` **o** `leader_email`** (mutuamente excluyentes, `CHECK` + service), validados con los tags `phone`/`email` **existentes**; el contacto exige nombre | Q8 literal y **cero código de validación nuevo** (R6-4) | Un par `kind`/`value` (pierde los tags); ambos a la vez (contradice Q8); librería de teléfono/correo |
| **P6-5** | **Enlace**: tag **`url`** de `platform/validate` (https + host, **sin dominios**) + sin espacios/control + `CHECK '^https://[^[:space:]]+$'`; el valor se sirve **tal cual** | FR-015/Q3 piden exactamente esta regla, que F3 ya implementó y probó (R6-5) | Lista blanca de dominios de WhatsApp (descartada por Q3); librería de URLs; normalizar/reescribir el `href` (FR-015 lo prohíbe); validar solo en el frontend |
| **P6-6** | **Permisos**: `AdminChain("ministerios")` con el permiso **ya sembrado** en `000002`; **sin migración de permisos** ni cambios en F2 (el catálogo sigue en 10) | FR-011 literal ("F6 NO crea permisos nuevos"); `AdminChain`/`AuthzByModule` bastan (mismo gesto que F3/F4) | Permiso nuevo por sub-acción; derivar el acceso de `portada`/`eventos`; migración de siembra innecesaria (F4 la necesitó; F6 no) |
| **P6-7** | **Imágenes**: `platform/storage` + `internal/media` reutilizados; subida `POST /api/v1/admin/ministerios/imagenes` tras su permiso; **`IsFilePublished` crece** con `ministries` (el borrador no sirve su imagen) | Un solo servicio de archivos para F3–F9 con la regla "solo contenido publicado" intacta (R6-7) | Servicio de archivos nuevo; servir todo lo subido (rompe FR-012 "por ninguna vía"); endpoint de descarga por módulo |
| **P6-8** | **Superficie HTTP**: `GET /api/v1/ministerios?lang=` (sección completa) + `GET /api/v1/portada/ministerios?lang=` (bloque, tope 5), ambos `no-store`; panel `/api/v1/admin/ministerios` con `limit`/`offset` y filtro `publicationState` | FR-027 exige el tope **resuelto en el servidor**, que no es lo que necesita la sección: dos documentos, un solo predicado "publicado" (R6-8) | Un solo endpoint para ambos (o el tope queda en el cliente, o la sección se queda sin el resto); derivar los 5 en el navegador; añadir el bloque a `GET /api/v1/portada` de F3 (tocaría su contrato) |
| **P6-9** | **Presencia de la navegación desde el documento del bloque** (`usePortadaMinisterios`, compartido con `MinisteriosBlock`): entrada «Ministerios» solo si `items` trae algo | Es **equivalente** a mirar la sección (mismo predicado "publicado") y **0 peticiones extra** en la portada, la página más visitada (R6-8) | Presencia desde la lista completa (cada página descargaría descripciones de 2.000 caracteres para pintar un enlace); un endpoint "presencia" (una ruta pública más para un booleano) |
| **P6-10** | **Bilingüe** `*_es`/`*_en` por campo + `?lang=es\|en` resuelto en el servidor con fallback `en → es` (patrón R3-1/R3-2 literal); traducibles: nombre, descripción, encargado, horario, lugar y alt; **no** traducible: el enlace | FR-007/FR-008/SC-006 garantizados en una capa reutilizable; Q6/FR-024 tratan horario y lugar como textos normales | Tabla de traducciones EAV; fallback en el cliente (analyze I2 de F3); `Accept-Language` |
| **P6-11** | **Validación**: tags existentes (`required`, `max`, `email`, `phone`, `url`) + reglas de dominio y límites como constantes del service (nombre/encargado 160, descripción **2.000**, horario/lugar 200, enlace 500, alt 300) | FR-014 completo sin dependencias nuevas; los mensajes salen del service con `details` por campo | Librerías de validación; límites solo en `CHECK` (mensajes peores); exigir imagen u horario (son opcionales desde Q4/Q6) |
| **P6-12** | **Duplicado exacto → `409 conflict`** (nombre+descripción+encargado+enlace normalizados); **enlace de WhatsApp compartido → aviso** (`whatsappUrlShared` en el DTO del panel) sin bloquear | Los dos defaults revisables de la spec, aplicados como los describe: "se rechaza" el duplicado y "se señala como aviso pero se permite" el enlace compartido (R6-11) | `409` también para el enlace compartido (la spec lo permite); solo señalizar en el formulario (la API forzada quedaría sin señal); `UNIQUE (whatsapp_url)` (prohibiría el grupo compartido) |
| **P6-13** | **Auditoría**: 8 códigos `ministry.*` (migración `000011`), `InsertAdminAction` transaccional y `Recorder` del propio dominio; visibilidad de la ficha con código propio; reordenar = `ministry.update` | FR-017 enumera la visibilidad como operación aparte y es la decisión sensible del módulo (R6-12) | Registro propio de F6 (FR-017 pide el de F2); un solo código `ministry.update` para todo; registrar tras el commit |
| **P6-14** | **Publicación** por elemento (`publication_state`) y "se publica al guardar"; consultas `…Published` separadas para el público | FR-012/FR-013 literales; patrón probado en F3/F4 (R6-13) | Estado por módulo; paso intermedio de borrador al editar (contradice FR-013) |
| **P6-15** | **Sección sin contenido**: entrada oculta (barra, hamburguesa y pie) y **`/ministerios` redirige a la portada** (mecánica en **P6-21**): decisión del humano del 2026-10-13, **distinta a propósito** del patrón de F4/R4-13 (aviso sin redirigir), que **se conserva en `/eventos`** y no se unifica | SC-010 exige 0 enlaces rotos **y** 0 pantallas sin explicación: la redirección lleva a una **página completa con toda la navegación**, de modo que un enlace guardado no muere en un callejón (R6-13) | **Aviso de sección vacía sin redirigir** (`HiddenNotice`, patrón firme de F4 — era la propuesta de este plan, **descartada por el cliente para F6**—); página vacía sin mensaje; `404` (un error para algo normal) |
| **P6-16** | **Bloque de la portada**: endpoint propio `GET /api/v1/portada/ministerios` (registrado por el dominio `ministerios`, mismo gesto que F4 con `proximamente`), **tope 5 firme** (`maxPortadaItems = 5` + `LIMIT`), orden `sort_order, id`, DTO `id`+`name`+`schedule`/`place`+`image` (la tarjeta compacta de `ux.md` §4.9), **oculto por completo sin publicados** y posicionado **tras `ProximamenteBlock` y antes de `WhoWeAreSection`** | FR-027/FR-028 literales, incluida la diferencia consciente con el bloque de F4 (aquí no hay mensaje) y la posición acordada (R6-14) | Bloque dentro de `GET /api/v1/portada` (mezcla dominios); tope configurable (Out of Scope); mensaje de sección vacía como F4 (contradice FR-028); carrusel/paginación (Out of Scope) |
| **P6-17** | **Frontend**: `/ministerios` + `/panel/ministerios` (`RequirePermission(MINISTERIOS)`), `MINISTERIOS` en `lib/permissions`, tarjeta en `AccesosRapidos`, `ACTION_LABELS` +8, y **solo `CardImage`** promovido al inventario público (`HiddenNotice` **no**: F6 redirige en vez de avisar, P6-21); nombres de pieza según `ux.md` §6.3/§6.4 | Una sola entrada de navegación (FR-006) y un módulo de panel con el patrón de F4; la regla del inventario único (F2/T242) manda promover lo compartido | Copiar `CardImage`/`ImageUploader` en la feature; promover `HiddenNotice` sin uso (quedaría muerta fuera de F4); `RequireAnyPermission` (sobra: permiso propio); una sola página de panel para público y gestión |
| **P6-18** | **Contrato**: delta en `specs/006-ministerios/contracts/openapi.yaml` → fusión aditiva en `backend/api/openapi.yaml` con `info.version` **0.6.0 → 0.7.0**; el registro de `error.code` **no cambia**; en la fusión se actualiza el enum de `AdminActionItem.action` (+8) y la prosa de `GET /api/v1/media/{fileName}` | Contrato antes que el código (§II) sin dos copias vivas ni códigos de error ad-hoc (§8.1.10: minor al añadir operaciones) | Prefijo nuevo de rutas; `error.code` para "enlace compartido" (es un aviso, no un error); enum de permisos en el contrato (el catálogo vive en la BD) |
| **P6-19** | **Orden `sort_order, id`** en público **y** en panel (el contrato documenta el criterio) | US2 esc. 9: "el nuevo orden es exactamente el que ve el visitante"; el panel debe mostrar el orden real (Q5) | `created_at DESC` en el panel (el default de §8.1.7 para listados sin criterio: aquí **hay** criterio) |
| **P6-20** | **Concurrencia**: última escritura completa gana (`PATCH` escribe el DTO entero en una transacción con su auditoría); **0 dependencias nuevas** | Defaults ya aprobados de F3/F4 (R6-16/R6-17) | Control optimista `version`/`If-Match` (fuera de alcance; si se pide, es una columna + `409`) |
| **P6-21** | **Redirección de `/ministerios` sin parpadeo** *(nueva, siguió a la puerta 2 del 2026-10-13)*: la decisión vive en el **frontend**, en el **elemento de ruta** de `/ministerios` (no en el backend ni en el contrato); `SectionSkeleton` mientras `usePublicMinistries` está en vuelo y **solo con éxito y sin `items`** → `<Navigate to="/" replace />`; **ante error NO se redirige** (estado de error con «Volver a intentar», patrón de `EventsPage`; con caché, contenido + aviso); la API sigue respondiendo **`200` con `items` omitida** | La redirección es una decisión de **presentación del SPA**: `replace` evita el bucle con «Atrás», el esqueleto impide el destello del estado vacío (estado de carga → redirección limpia) y un `3xx` del API rompería el sobre `{lang, items?}` y a sus consumidores (navegación y pruebas) | Redirigir desde el backend o el contrato (rompe el contrato JSON); redirigir también ante **error** de red (una caída transitoria expulsaría al visitante); pintar el aviso vacío antes de redirigir (destello); aviso sin redirigir (patrón F4, descartado por el cliente para F6) |

## Project Structure

### Documentation (esta funcionalidad)

```text
specs/006-ministerios/
├── plan.md              # EDITABLE (este archivo, /speckit.plan)
├── research.md          # EDITABLE (fase 0: R6-1…R6-17)
├── data-model.md        # EDITABLE (fase 1: tabla ministries, migraciones 000010–000011, consultas)
├── quickstart.md        # EDITABLE (fase 1: cómo probar F6 en local)
├── spec.md              # APROBADA (v2) — no se edita
├── contracts/
│   └── openapi.yaml     # EDITABLE (fase 1): delta de diseño, se fusiona en backend/api/openapi.yaml
├── ux.md                # DEL disenador-ux (ya producido en paralelo): composición visual de la
│                        #   ficha, del listado y del bloque (D-1…D-12, §9). Este plan fija rutas,
│                        #   DTOs, reglas de contenido y posiciones; sus nombres de pieza son la
│                        #   referencia de tasks.md (como en F4, RG6-4)
└── tasks.md             # Fase 3 (/speckit.tasks — NO lo crea este comando)
```

### Source Code — backend

```text
backend/
├── cmd/api/
│   ├── main.go                      # EDITADO: DI del dominio ministerios + registro de
│   │                                #   /api/v1/ministerios, /api/v1/portada/ministerios y
│   │                                #   /api/v1/admin/ministerios
│   └── main_test.go                 # EDITADO: prueba de humo ampliada (rutas nuevas + /healthz intacto)
├── internal/
│   ├── db/
│   │   ├── queries/
│   │   │   ├── ministries.sql       # NUEVO: ListPublishedMinistries (público, con
│   │   │                            #   CASE WHEN leader_visible), ListMinistries/CountMinistries,
│   │   │                            #   GetMinistryByID (+whatsapp_url_shared), Insert/Update/Delete,
│   │   │                            #   ExistsMinistryDuplicate
│   │   │   ├── media.sql            # EDITADO: IsFilePublished crece con ministries (R6-7)
│   │   │   └── audit.sql            # SIN CAMBIOS: se reutiliza InsertAdminAction (R6-12)
│   │   └── (generado por sqlc, commiteado)
│   ├── ministerios/                 # NUEVO — dominio F6 (receta de arq. §8)
│   │   ├── model.go                 # entidad Ministry (leader como campos) y DTOs (etiquetas
│   │   │                            #   validate, Es/En, LeaderVisible, SortOrder)
│   │   ├── repository.go            # constructor + mapRow/params (pgtype → dominio) + withTx
│   │   │                            #   + inserción de admin_actions sobre el tx
│   │   ├── repository_ministerios.go# CRUD + ExistsMinistryDuplicate + whatsapp_url_shared
│   │   ├── repository_public.go     # ListPublishedMinistries (lectura pública: solo published y
│   │   │                            #   sin datos ocultos del encargado — R6-2)
│   │   ├── service.go               # tipos comunes, invariantes, constantes (límites,
│   │   │                            #   maxPublicItems = 200, maxPortadaItems = 5)
│   │   ├── service_public.go        # GetMinisterios (sección, ?lang + fallback) y
│   │   │                            #   GetPortadaMinisterios (tope 5 en servidor)
│   │   ├── service_admin.go         # CRUD, validaciones de dominio (contacto excluyente,
│   │   │                            #   enlace https, español por campo, imagen+alt),
│   │   │                            #   publicar/retirar, visibilidad del encargado, duplicados
│   │   │                            #   y auditoría transaccional
│   │   ├── service_audit.go         # audit.Recorder del dominio: método+ruta → ministry.* (denegaciones)
│   │   ├── handler.go               # constructor + helpers HTTP (decodificar/validar/responder)
│   │   ├── handler_public.go        # GET /api/v1/ministerios y GET /api/v1/portada/ministerios
│   │   ├── handler_admin.go         # CRUD de /api/v1/admin/ministerios
│   │   ├── handler_images.go        # POST /api/v1/admin/ministerios/imagenes (multipart)
│   │   ├── routes.go                # RegisterPublic (2 rutas) y RegisterAdmin
│   │   │                            #   (AdminChain("ministerios"))
│   │   └── *_test.go                # service con fakes · handler con httptest · repository integration
│   ├── media/                       # EDITADO (solo la consulta): IsFilePublished cubre ministries
│   └── platform/
│       ├── audit/audit.go           # EDITADO: + 8 códigos ministry.* (y su prueba del registro cerrado)
│       ├── middleware/chain.go      # SIN CAMBIOS: AdminChain("ministerios") reutiliza AuthzByModule
│       ├── validate/validate.go     # SIN CAMBIOS (los tags url/phone/email de F2/F3 bastan)
│       └── config/config.go         # SIN CAMBIOS (sin variables de entorno nuevas)
├── migrations/
│   ├── 000010_create_ministries.up.sql / .down.sql              # NUEVO: tabla ministries + índices
│   └── 000011_extend_admin_actions_for_ministries.up.sql / .down.sql # NUEVO: CHECK de auditoría +8 códigos
└── api/openapi.yaml                 # EDITADO: se fusiona el delta (info.version → 0.7.0) + make api-gen
```

### Source Code — frontend y raíz

```text
frontend/src/
├── api/
│   ├── ministerios.ts               # NUEVO: getMinisterios(lang), getPortadaMinisterios(lang),
│   │                                #   panel (list/create/update/delete) + uploadMinistryImage
│   └── schema.d.ts                  # REGENERADO (npm run api-gen; contrato 0.7.0)
├── app/
│   ├── router.tsx                   # EDITADO: '/ministerios' (pública) y '/panel/ministerios'
│   │                                #   con RequirePermission(MINISTERIOS)
│   ├── layout.tsx                   # EDITADO: entrada del menú del panel filtrada por permiso
│   └── guards.tsx                   # SIN CAMBIOS (RequirePermission de F2 basta)
├── features/
│   ├── publico/                     # EDITADO: sitio público (F10 aplicado, no redefinido)
│   │   ├── ministerios/             # NUEVO (nombres de pieza según ux.md §6.3):
│   │   │   ├── MinisteriosPage.tsx  #   H1 + grilla de fichas; sin publicados → REDIRIGE a la
│   │   │   │                        #   portada (P6-21: SectionSkeleton durante la carga y
│   │   │   │                        #   <Navigate to="/" replace /> solo con éxito y sin items;
│   │   │   │                        #   nunca aviso de sección vacía; ante error → estado de
│   │   │   │                        #   error con «Volver a intentar», sin redirigir)
│   │   │   ├── MinistryCard.tsx     #   ficha: imagen (CardImage), nombre, descripción completa,
│   │   │   │                        #   líneas de horario/lugar solo con dato, EncargadoLine SOLO
│   │   │   │                        #   si viene `leader` (FR-026) y WhatsappJoinButton SOLO si
│   │   │   │                        #   hay whatsappUrl (FR-003)
│   │   │   ├── EncargadoLine.tsx / WhatsappJoinButton.tsx  # piezas de la ficha (ux.md §6.3)
│   │   │   ├── MinisteriosBlock.tsx #   bloque de la portada: fila/grilla de hasta 5 MinistryItem;
│   │   │   │                        #   sin publicados → null: NO se renderiza (FR-028)
│   │   │   ├── MinistryItem.tsx     #   tarjeta compacta: imagen, chip «Ministerio», título-enlace
│   │   │   │                        #   a /ministerios, horario/lugar si existen (sin encargado
│   │   │   │                        #   ni botón de unión; ux.md §4.9)
│   │   │   ├── hooks/usePublicMinistries.ts     # TanStack Query ['ministerios', lang]
│   │   │   ├── hooks/usePortadaMinisterios.ts   # ['portada-ministerios', lang] (bloque + navegación)
│   │   │   └── *.test.tsx
│   │   ├── hooks/useSiteLinks.ts    # EDITADO: SiteLinkId + 'ministries'; entrada «Ministerios»
│   │   │                            #   (href /ministerios) solo si el bloque trae ítems
│   │   ├── pages/HomePage.tsx       # EDITADO: + <MinisteriosBlock /> inmediatamente después de
│   │   │                            #   <ProximamenteBlock /> y antes de <WhoWeAreSection />
│   │   │                            #   (una inserción; el resto de la portada de F3 intacto)
│   │   └── i18n/{messages,es,en}.ts # EDITADO: claves tipadas nuevas (catálogo de ux.md §7.2)
│   ├── ministerios/                 # NUEVO: gestión en el panel (en español, FR-018; ux.md §6.4)
│   │   ├── pages/MinistriesPage.tsx #   listado con StatusPill, sortOrder y acciones (ux.md §4.3)
│   │   ├── components/              # MinistryForm (Dialog, pestañas es/en), MinistriesList,
│   │   │                            #   ContactoEncargado (Select «Sin contacto/Celular/Correo» +
│   │   │                            #   un solo Field), ShowContactSwitch (role="switch" con
│   │   │                            #   leyendas de privacidad, ux.md §4.5; NACE APAGADO al
│   │   │                            #   crear —«Oculto por defecto», P6-3— y precarga su última
│   │   │                            #   decisión al editar)
│   │   ├── hooks/                   # useMinistries, useCreateMinistry, useUpdateMinistry,
│   │   │                            #   useDeleteMinistry, useUploadMinistryImage, queryKeys
│   │   ├── validation.ts            # Zod (espejo de las reglas del server)
│   │   ├── errors.ts                # details → errores por campo
│   │   ├── messages.ts              # textos del panel (catálogo de ux.md §7.1)
│   │   └── *.test.tsx
│   ├── roles/permissions.ts         # EDITADO: isPermissionAvailable deja de marcar «ministerios»
│   │                                #   como «Disponible más adelante» (y sus tests)
│   ├── panel/components/AccesosRapidos.tsx  # EDITADO: tarjeta «Gestión de ministerios» filtrada
│   │                                #   por permiso (camino feliz de US2)
│   ├── auditoria/messages.ts        # EDITADO: ACTION_LABELS cubre los 8 códigos ministry.*
│   ├── eventos/components/          # MOVIDO → src/components/publico/: la pieza CardImage de
│   │                                #   ItemCardParts.tsx (inventario único). HiddenNotice SE
│   │                                #   QUEDA en eventos: F6 ya no pinta aviso de sección
│   │                                #   vacía (redirige, P6-21) y F4 lo sigue usando
│   └── (resto SIN CAMBIOS)
├── lib/permissions.ts               # EDITADO: + MINISTERIOS = 'ministerios' (+ MODULE_PERMISSIONS)
├── components/publico/              # EDITADO: + CardImage (promovido; HiddenNotice NO se
│                                    #   promueve: F6 redirige en vez de avisar — P6-21)
└── (index.css, tailwind.config)     # SIN CAMBIOS (tokens de F10/F3)
frontend/e2e/
├── ministerios-publica.spec.ts      # NUEVO: visitante (US1, US4, US5) — incluye ficha oculta,
│                                    #   ministerio sin enlace, redirección a la portada sin
│                                    #   publicados (P6-21) y alta sin leaderVisible (nace oculto)
├── ministerios-panel.spec.ts        # NUEVO: gestión, permisos y privacidad (US2, US3)
├── ministerios-accesibilidad.spec.ts# NUEVO: recorrido por teclado y foco (US5, SC-008), incluida
│                                    #   la redirección (el foco no queda atrapado)
├── portada-ministerios.spec.ts      # NUEVO: bloque de la portada (US6, SC-016)
└── ministeriosSeed.ts               # NUEVO: fixtures del módulo (patrón de agendaSeed.ts)

# Raíz
README.md                            # EDITADO por documentador al cerrar
.env.example                         # SIN CAMBIOS (F6 no añade variables)
docker-compose.yml / Makefile / .github/workflows/ci.yml / .githooks/  # SIN CAMBIOS (son del kit)
```

### Inventario único de componentes UI

F6 **promueve una pieza** al inventario único, de `features/publico/eventos/` porque F6 la
comparte (regla de F2/T242, mismo gesto que F4 hizo con `ImageUploader`/`PublishControls`):
**`CardImage`** (la imagen de la tarjeta, hoy dentro de `ItemCardParts.tsx` — decisión D-3 de
`ux.md`, que deja la elección al arquitecto: se promueve en lugar de importarla por feature). Va a
`src/components/publico/`, **sin cambios de comportamiento** (solo se mueve junto con sus pruebas).
**`HiddenNotice` NO se promueve** (cambio sobre la primera versión de este plan, tras la decisión
de la puerta 2): F6 ya **no** pinta un aviso de sección vacía —la URL directa `/ministerios`
redirige a la portada, P6-21— y la pieza se queda en `features/publico/eventos/`, donde F4 la
sigue usando para `/eventos` (su patrón no se toca). El resto **se reutiliza sin tocar**:
`Field`, `TextArea` (de `features/informacion/`), `Select`, `Button`, `Notice`, `ConfirmDialog`,
`Dialog`, `EmptyState`, `Table`, `Tabs`, `StatusPill`, `ImageUploader`, `PublishControls` y el
sistema de F10 (`Card`, `SectionHeading`, `PublicButton`, `Reveal`/`RevealGroup`, `SectionSkeleton`,
`PublicLayout`, `PublicNav`, `MobileMenu`, `FooterSiteMap`, `LanguageSwitcher`, `ThemeToggle`).
Los componentes propios de F6 (`MinistryCard`, `EncargadoLine`, `WhatsappJoinButton`,
`MinistryItem`, `MinistryForm`, `ContactoEncargado`, `ShowContactSwitch`, `MinistriesList`) viven en
sus carpetas de feature y **solo se promueven** si otra feature los necesita más adelante.

## Cadena de middleware y grupos (orden; el primero es el más externo)

```text
Global:   request-id → recover → logging → CORS(con credenciales) → handler
          /api/v1/ministerios               → (sin middlewares)  → handler    ← público, solo lectura
          /api/v1/portada/ministerios       → (sin middlewares)  → handler    ← público, solo lectura (lo registra
                                                                               `ministerios.RegisterPublic`, aunque el
                                                                               segmento `portada` sea el agregado de F3)
          /api/v1/media/{fileName}          → (sin middlewares)  → handler    ← público, solo lectura (F4; ahora con
                                                                               imágenes de ministerios publicados)
          /api/v1/admin/ministerios         → authn → guard de cambio de contraseña → authz("ministerios") → CSRF → handler
```

- La superficie pública de F6 es de **solo lectura**: sin `authn`, `CSRF` ni `rate-limit` (F2 solo
  lo monta sobre superficies públicas escribibles; aquí no hay ninguna).
- El grupo de panel monta la **cadena aprobada de F2** con `AuthzByModule` tal cual (solo cambia el
  código de permiso: `ministerios`, ya sembrado). La denegación responde `403 forbidden` genérico
  y **queda registrada** en `admin_actions` (`result='denied'`); el `Recorder` de la cadena es el
  del dominio `ministerios` (resuelve sus rutas → acción `ministry.*`, R6-12).
- `/healthz` **no cambia**: sigue público, sin sesión y con `Cache-Control: no-store`.

## Contrato OpenAPI: dónde vive y cómo se mantiene

Igual que F1/F2/F3/F4 (reglas §8.1.5 y §8.1.10):

1. **Diseño (esta fase)**: `specs/006-ministerios/contracts/openapi.yaml` — delta inmutable una vez
   aprobado el plan.
2. **Fusión**: la primera tarea de implementación lo funde en `backend/api/openapi.yaml` de forma
   **aditiva**, aplica los cambios documentados del delta (enum `action` de `AdminActionItem` **+8
   códigos** `ministry.*` y prosa de `GET /api/v1/media/{fileName}`: ahora también sirve imágenes de
   ministerios publicados) y sube `info.version` de **0.6.0 a 0.7.0** (minor: operaciones y
   esquemas nuevos; el registro de `error.code` **no** cambia). Luego `make api-gen`.
3. **Consumo**: el frontend genera tipos solo desde `backend/api/openapi.yaml` (`schema.d.ts`).
4. El snapshot de `specs/` no se vuelve a editar.
5. **La prosa del catálogo de permisos NO cambia**: F2 sigue teniendo 10 permisos; `ministerios`
   ya estaba sembrado por `000002` (a diferencia de F4, que sí tuvo que ampliar el catálogo).

## Estrategia de pruebas

Cobertura exigida: **80 % en `service/`** (§III). Comandos: `go test ./...` ·
`go test -tags=integration ./...` · `npm test -- --run` · `make e2e` · `make ci`.

| Capa | Tipo | Qué verifica F6 | Dónde |
|---|---|---|---|
| `platform/audit` | Unitaria | Los 8 códigos nuevos en el registro cerrado; objetivo `content` con `targetLabel` | `internal/platform/audit/*_test.go` |
| `service` (dominio `ministerios`) | Unitaria con **fakes** | **Público**: solo `published`; **`leader` presente solo si `leaderVisible` y hay nombre** (la prueba más importante: con el interruptor apagado el DTO público **no contiene** el nombre ni el contacto, ni como campo ni como cadena, en `es` y `en`); **«Oculto por defecto»**: un alta **sin** `leaderVisible` nace con `leaderVisible: false` y su DTO público **no** trae `leader` aunque haya nombre y contacto guardados (P6-3); fallback `en → es` por campo (nunca huecos); `lang` inválido → `invalid`; orden `sortOrder, id`; `items` omitida sin publicados; **tope `maxPublicItems = 200`** y **`maxPortadaItems = 5`** fijados por prueba (el bloque: 5 exactos con 8 publicados, exactos con menos, sin `items` con 0). **Panel**: validación de FR-014 (obligatorios, `*En` exige su `*Es`, celular **o** correo — nunca ambos —, contacto exige nombre, `whatsappUrl` https seguro — tabla de casos con `javascript:`/`data:`/`vbscript:`/`http://`/sin host/espacios/501 chars, y dominio no de WhatsApp **aceptado** (Q3) —, alt con imagen, límites, `""` → `NULL`), duplicado exacto → `conflict`, **enlace compartido → permitido con `whatsappUrlShared`**, publicar/retirar por elemento, "se publica al guardar" (FR-013), visibilidad del encargado sin perder datos (US2 esc. 10–11). **Auditoría**: cada mutación registra su acción con el código y `targetLabel` correctos en la misma transacción (fail-closed; datos+estado/visibilidad → dos filas); denegaciones y rechazos registrados best-effort | `internal/ministerios/service_*_test.go` |
| `handler` | Unitaria con `httptest` + service falso | Cada operación: decodificación, `details` por campo en 400, códigos 200/201/204/400/401/403/404/409, sobres uniformes; **el DTO público serializado nunca contiene datos del encargado oculto** (aserción sobre el cuerpo JSON) ni `publicationState`; `GET /api/v1/portada/ministerios`: `200` con `no-store`, `items` omitida con 0 publicados; multipart: sin archivo, archivo gigante (400 con `details.file`), extensión no permitida | `internal/ministerios/handler_*_test.go` |
| `repository` | **Integración** (PostgreSQL real, `//go:build integration`) | SQL real: `CHECK` de contacto excluyente, "contacto exige nombre", español por campo, alt con imagen y enlace `^https://[^[:space:]]+$` (con `javascript:` rechazado en la BD); `ListPublishedMinistries` **nunca** devuelve borradores y con `leader_visible = FALSE` las columnas del encargado salen **`NULL`** (R6-2); **el `DEFAULT` de `leader_visible` es `FALSE`** (un `INSERT` que omite la columna nace oculto — P6-3 —); `whatsapp_url_shared` solo con otra fila igual; `ExistsMinistryDuplicate` con normalización (y sin disparar por enlace compartido); `IsFilePublished` con ministerio publicado / borrador; **mutación + `admin_actions` atómicas** (y rollback si falla el registro); migración `000010`/`000011` up→down→up | `internal/ministerios/repository_*_test.go`, `internal/media/*_test.go` |
| `cmd/api` | Humo de composición | Rutas nuevas con su sobre y permiso, `/healthz` intacto y `404` `not_found` en rutas inexistentes (§8.1.9) | `cmd/api/main_test.go` |
| Frontend | Unitaria (Vitest + MSW) | Pública: ficha completa con horario/lugar/imagen; **sin `leader` no se pinta línea de encargado**; **sin `whatsappUrl` no se pinta botón de unión** (SC-009/SC-017); fallback en→es visible; entrada de navegación y bloque ocultos sin publicados; **sin publicados la página redirige a la portada sin destello** (esqueleto `SectionSkeleton` mientras carga; nunca se pinta la grilla vacía ni aviso alguno; `<Navigate replace>`; y **ante error NO redirige**: estado de error con «Volver a intentar»); **el interruptor `leaderVisible` nace apagado al crear** («Oculto por defecto», P6-3); `alt` en imágenes; bloque con ≤5 `MinistryItem` (imagen, chip «Ministerio», título-enlace, horario/lugar si existen) y sin huecos. Panel: formulario con errores por campo (incluidos contacto y enlace), interruptor `leaderVisible` con datos conservados, `whatsappUrlShared` como `Notice`, publicar/retirar, `RequirePermission(MINISTERIOS)`, accesos rápidos, **`ACTION_LABELS` con etiqueta humana para los 8 códigos**, subida de imagen con error de tipo/tamaño; y las 4 pruebas actualizadas (permisos/accesos, RG6-6) | `frontend/src/features/*/*.test.tsx` |
| E2E pública | Playwright (local) | Recorrido de `quickstart.md`: visitante sin sesión ve los ministerios publicados en orden `sortOrder, id`, con encargado/horario/lugar/enlace cuando existen; **caso ficha oculta**: con `leaderVisible: false` y datos reales, **0 apariciones** del nombre ni del celular/correo en la respuesta de red y en el HTML (SC-014); **caso sin enlace**: sin botón de unión ni huecos (SC-009/SC-017); borrador invisible; **sin publicados, la visita directa a `/ministerios` redirige a la portada** (URL `/`, contenido de la portada, **sin destello** de sección vacía —ni grilla ni aviso en ningún instante— y «Atrás» sin bucles por el `replace`) y no hay entrada en nav/pie; **un alta sin `leaderVisible` nace oculto** («Oculto por defecto», P6-3: 0 apariciones del encargado); inglés con fallback y vuelta a español; 3 anchos (SC-001, SC-002, SC-006, SC-007, SC-010) | `frontend/e2e/ministerios-publica.spec.ts` |
| E2E panel | Playwright (local) | Recorrido de `quickstart.md`: login con permiso, crear+publicar un ministerio en < 2 min (SC-004), crear uno solo con nombre y descripción, **crear uno con encargado sin tocar el interruptor —nace oculto, «Oculto por defecto» (P6-3)—**, apagar y encender el interruptor sin perder datos (US2 esc. 10–11), enlace `javascript:` rechazado sin guardar (SC-009), celular **y** correo juntos rechazados (Q8), reordenar y ver el orden en público (SC-003), eliminar con confirmación, denegación cruzada (cuenta sin `ministerios`: 403 en API forzada) y registro de todo en auditoría (SC-005/SC-011) | `frontend/e2e/ministerios-panel.spec.ts` |
| E2E accesibilidad | Playwright (local) | Recorrido por **teclado** de `/ministerios` y del bloque con foco visible en cada paso (ficha y botones activables con `Enter`, selector de idioma operable, sin trampas de foco), en claro y oscuro (SC-008; precedente `portada-accesibilidad.spec.ts`); **la redirección sin publicados no atrapa el foco** (el visitante termina en la portada navegable, P6-21) | `frontend/e2e/ministerios-accesibilidad.spec.ts` |
| E2E portada | Playwright (local) | Bloque en `/`: exactamente los 5 primeros en orden `sortOrder, id` (con 8 publicados), exactos con menos, **sin huecos**; 0 publicados → **el bloque no existe** (sin mensaje, a diferencia de «Lo que viene»); posición inmediatamente después del bloque de eventos y antes de «Quiénes somos»; botón/tarjetas hacia `/ministerios`; 320 px y teclado (US6, SC-016) | `frontend/e2e/portada-ministerios.spec.ts` |

**Pruebas que no faltan** (las que la spec hace críticas): **0 borradores expuestos** por ninguna
respuesta pública ni por la descarga de imágenes (SC-002); **0 apariciones del encargado oculto**
en tráfico y HTML (SC-014, el e2e más exigente del módulo); **0 enlaces no seguros** aceptados ni
renderizados y **0 botones de unión vacíos** (SC-009); **cambio guardado visible en la primera
carga** (SC-003); **fallback en→es sin campos vacíos** (SC-006); **permisos verificados en
servidor** con cruce contra otros módulos (SC-005); **toda operación queda registrada** —éxito,
fallo y denegación— y el registro es de solo lectura (SC-011); **sin publicados la URL directa
`/ministerios` redirige a la portada sin destello de contenido vacío** (SC-010, P6-21);
**«Oculto por defecto» fijado por prueba en las cuatro capas** (P6-3: un alta sin `leaderVisible`
nace oculto); **bloque con exactamente 5 y oculto
sin publicados** (SC-016); **texto con tildes, `ñ`, `¿¡` y emojis intacto** (edge case);
**archivo inválido no rompe la sección** (edge case de imágenes); **horario de texto libre largo**
(«Una vez al mes, fecha variable») sin romper la presentación (SC-015).

## Cobertura de requisitos (US1–US6 y FR-001…FR-029)

| Historia | Dónde se cubre | Verificación |
|---|---|---|
| **US1** — Conocer los ministerios y unirse (P1) | `GET /api/v1/ministerios` (P6-8/P6-10) + `MinisteriosPage`/`MinistryCard` en `/ministerios` (P6-17) | quickstart §1 · e2e pública (SC-001, SC-009, SC-017) |
| **US2** — Gestionar los ministerios (P1) | `/api/v1/admin/ministerios` (P6-6/P6-11/P6-12/P6-13) + `features/ministerios/` (P6-17) + accesos de `InicioPage` | quickstart §2–§6 · e2e panel (SC-004, SC-005, SC-011) |
| **US3** — Controlar la publicación (P2) | `publication_state` + consultas `…Published` (P6-14) + `PublishControls` (promovido) + ocultación de sección/navegación (P6-9/P6-15) | quickstart §8 · e2e panel/pública (SC-002, SC-003, SC-010) |
| **US4** — Ver en inglés (P2) | `?lang=` con fallback en el servidor (P6-10) + i18n tipado | quickstart §9 · e2e pública (SC-006) |
| **US5** — Cualquier dispositivo y accesible (P2) | Sistema de F10 aplicado (P6-17) + reglas `react-frontend` + `alt` obligatorio | quickstart §11 (checklist) · e2e en 3 anchos (SC-007) · e2e de teclado (SC-008) |
| **US6** — Ver los ministerios desde la portada (P2) | `GET /api/v1/portada/ministerios` con tope 5 (P6-16) + `MinisteriosBlock` tras `ProximamenteBlock` (P6-16) + visibilidad coherente (P6-9) | quickstart §14 · e2e `portada-ministerios.spec.ts` (SC-016) |

| FR | Dónde se resuelve en este plan | Verificación |
|---|---|---|
| FR-001 | P6-1 (una tabla `ministries` sin fechas ni categorías: horario/lugar son `TEXT`) + P6-10 | quickstart §1 · pruebas de modelo (SC-013: nada de F4/F5 en la sección) |
| FR-002 | P6-8 (`GET /api/v1/ministerios`, solo publicado) + P6-2 (ficha `leader` condicional) | quickstart §1 y §7 · pruebas de service/handler |
| FR-003 | P6-5 (`whatsappUrl` opcional; la ficha solo pinta botón si hay enlace) | quickstart §5 y §7 · e2e pública (SC-009, SC-017) |
| FR-004 | P6-19 (`ORDER BY sort_order, id` en sección y bloque) | quickstart §1 y §6 · pruebas de orden |
| FR-005 | P6-9 (entrada oculta sin publicados) + P6-15/P6-21 (la URL directa `/ministerios` **redirige a la portada**) | quickstart §1 y §8 · e2e (SC-010) |
| FR-006 | P6-17 (ruta `/ministerios` y entrada única «Ministerios» en barra, hamburguesa y pie) | quickstart §1 · e2e pública (SC-010) |
| FR-007 | P6-10 (columnas `*_es`/`*_en` + regla por campo) | quickstart §4 y §9 · pruebas de validación |
| FR-008 | P6-10 (fallback `en → es` en el servidor; el enlace no es traducible) | quickstart §9 · tabla de casos del service (SC-006) |
| FR-009 | `LanguageSwitcher` de F10 (sin cambios; patrón F3) | quickstart §9 · e2e pública (cambio de idioma persistente) |
| FR-010 | `/api/v1/admin/ministerios` (P6-8) + `MinistriesPage`/`MinistryForm` (P6-17) | quickstart §2–§6 · e2e panel (SC-004) |
| FR-011 | P6-6 (`AdminChain("ministerios")`, permiso ya sembrado en `000002`; sin permisos nuevos) | quickstart §2 y §12 · pruebas de handler (403 por API forzada, cruce de permisos) · e2e (SC-005) |
| FR-012 | P6-14 (`publication_state` por elemento) + consultas `…Published` + `IsFilePublished` (P6-7) | quickstart §8 · pruebas "0 borradores expuestos" (SC-002) |
| FR-013 | P6-14 (guardado de elemento publicado visible en la primera carga; `no-store`) | quickstart §8 · e2e panel (SC-003) |
| FR-014 | P6-11 (obligatorios, español por campo, contacto excluyente, enlace, imagen+alt, límites) | quickstart §3–§6 (tablas de errores) · pruebas de service/handler |
| FR-015 | P6-5 (tag `url` + endurecimiento + `CHECK`; sin dominios; sin transformar) | quickstart §5 · tabla de casos + e2e (SC-009) |
| FR-016 | P6-19 (`sortOrder` editable en el panel; el visitante ve ese orden) | quickstart §6 · e2e panel (SC-003) |
| FR-017 | P6-13 (8 códigos `ministry.*`, transaccional, denegaciones registradas, solo lectura) | quickstart §13 · pruebas de service/repository (SC-011) |
| FR-018 | Panel en español (convención de F2/F3/F4; sin `i18n/` en la gestión) | revisión de `revisor-codigo` (0 cadenas de panel pasadas por `t()`) |
| FR-019 | P6-1 (`ministries`: nombre, descripción, encargado, enlace, horario, lugar, imagen) | quickstart §3–§6 · pruebas de persistencia |
| FR-020 | P6-7 (imagen opcional, `CHECK` de alt obligatorio, subida validada) | quickstart §10 · pruebas de service/handler |
| FR-021 | P6-16/P6-17 (F10 aplicado; el único contenido de F6 en la portada es el bloque) | revisión de `revisor-codigo` sobre `HomePage` (una inserción) y `PublicLayout` sin cambios |
| FR-022 | P6-17 (componentes de F10, títulos claros, `toc-targets` ≥ 44 px) | quickstart §11 · e2e en 320/768/1280 (SC-007) |
| FR-023 | Accesibilidad: teclado, semántica, foco, contraste claro/oscuro, zoom 200 %, `alt` (P6-7/P6-17) | quickstart §11 (checklist WCAG 2.1 AA) · e2e de teclado (SC-008) |
| FR-024 | P6-1/P6-10 (horario y lugar `TEXT` traducibles, sin fechas; presentación de texto libre) | quickstart §4 · pruebas de presentación (SC-015: 0 cálculos de recurrencia) |
| FR-025 | P6-3 (`leaderVisible`, dato guardado ≠ publicado, **«Oculto por defecto»**) + P6-4 (contacto opcional) | quickstart §3 y §7 · e2e panel (US2 esc. 10–11) |
| FR-026 | **P6-2** (DTO público + consulta con `CASE WHEN leader_visible` + bloque sin campos de encargado + pruebas "0 apariciones") | quickstart §7 (inspección de red y HTML) · pruebas de service/handler/repository · e2e (SC-014) |
| FR-027 | P6-16 (endpoint propio, tope 5 en servidor, orden manual, tarjetas con imagen, botón a `/ministerios`, posición tras el bloque de eventos) | quickstart §14 · pruebas del service/repository (SC-016) · e2e de portada |
| FR-028 | P6-16 (sin publicados: `items` omitida y **el bloque no se renderiza**, sin mensaje) | quickstart §14 · e2e de portada (SC-010, SC-016) |
| FR-029 | P6-10 (bilingüe con fallback) + P6-16/P6-17 (F10 aplicado al bloque; el resto de la portada intacto) | quickstart §9/§11/§14 · e2e en 320 px y por teclado (SC-006, SC-007, SC-008) |

## Métricas del plan (coherentes con los SC de la spec)

| SC | Qué se mide | Cómo se mide en F6 | Umbral |
|---|---|---|---|
| SC-001 | Todo visitante ve los ministerios publicados sin cuenta (nombre/descripción siempre; encargado solo si se publica; horario/lugar/imagen/enlace cuando existen) y encuentra cómo unirse en < 30 s | e2e pública: la sección muestra las fichas con sus datos en la primera carga y el botón de unión cuando hay enlace | 100 % |
| SC-002 | 0 borradores visibles al público (tampoco en imágenes) | Pruebas "0 borradores expuestos" (service + handler + repository + `IsFilePublished`) y e2e con borradores en BD | 0 |
| SC-003 | Cambio guardado visible desde la primera carga posterior | e2e panel: editar/reordenar → recargar → ver el dato (`no-store` de respaldo) | 100 % |
| SC-004 | Crear y publicar un ministerio en < 2 min desde el panel | e2e/observación (quickstart §3) | < 2 min |
| SC-005 | Toda operación verifica el permiso `ministerios` en el servidor (incluidos intentos forzados y el cruce con otros permisos) | Pruebas de handler por operación (403) + e2e con cuenta sin `ministerios` y con otros permisos de módulo | 0 operaciones sin permiso / 100 % denegadas |
| SC-006 | Con `lang=en`: 100 % de la interfaz traducida y 0 campos vacíos (fallback a es) | Pruebas de i18n (claves tipadas) + tabla de casos de fallback + e2e en inglés | 100 % / 0 |
| SC-007 | Usable en teléfono, tableta y computadora sin desplazamiento horizontal (320 px) — sección **y** bloque | e2e en 3 anchos (320/768/1280) + e2e de portada en 320 px + checklist (quickstart §11/§14) | 0 pérdida / 0 scroll |
| SC-008 | Accesibilidad: 0 hallazgos críticos (WCAG 2.1 AA, claro y oscuro) — sección **y** bloque | Checklist de quickstart §11 (teclado, lector, contraste, zoom) revisada por QA + e2e de teclado | 0 críticos |
| SC-009 | 100 % de enlaces publicados se abren en un paso y **0 enlaces no seguros** aceptados ni renderizados; sin enlace → 0 botones vacíos | Tabla de casos del service (https sí / `javascript:`…) + e2e del panel (rechazo) y de la pública (`href` = valor guardado) | 100 % / 0 |
| SC-010 | 0 secciones vacías ni campos en blanco; navegación sin enlaces rotos **y** el bloque de la portada oculto sin publicados | Bloques/entrada omitidos sin contenido (pruebas de contrato y de `buildSiteLinks`) + e2e + la URL directa **redirige a la portada** sin destello de contenido vacío (P6-21) | 0 |
| SC-011 | 100 % de operaciones registradas (quién, qué, sobre qué, cuándo); 0 registros editables/borrables | quickstart §13 + pruebas transaccionales + contrato sin escritura sobre el registro | 100 % / 0 |
| SC-012 | ≥ 90 % de personas de distintas edades identifican qué ministerio es y cómo participar sin ayuda | Prueba manual con personas (protocolo en `quickstart.md` §11, estilo F3/F4/M7): ≥ 6 personas (2 por franja 18–35/36–59/60+), guion de 3 tareas, informe `pruebas-usabilidad-SC-012.md` | ≥ 90 % |
| SC-013 | 0 grupos de conexión (F5) ni eventos (F4) en la sección ni en el bloque; horario como texto libre sin fechas | Revisión de contenido e2e (solo ministerios) + pruebas de modelo (sin columnas de fecha) | 0 |
| SC-014 | **Con la ficha del encargado no publicada: 0 apariciones de su nombre y de su celular/correo** en cualquier respuesta pública (API, HTML, metadatos) y 0 líneas de encargado en la interfaz; los datos se conservan en el panel | Pruebas "0 apariciones" (service: el DTO no los contiene · handler: el cuerpo JSON no los contiene · repository: la fila sale con `NULL`) + **e2e con inspección de tráfico y de HTML** (quickstart §7), incluido el caso por defecto —**un alta sin `leaderVisible` nace oculto** («Oculto por defecto», P6-3)— | 0 (100 % de las pruebas) |
| SC-015 | Los ministerios con horario/lugar publicados los muestran tal como se escribieron (con fallback); 0 líneas vacías y 0 cálculos de recurrencia/fechas | Pruebas de presentación (ficha con texto libre largo) + revisión de `revisor-codigo` (sin fechas derivadas) | 0 |
| SC-016 | El bloque muestra exactamente los primeros 5 publicados (`sortOrder`, `id`) o todos si hay menos; 0 borradores, 0 huecos; botón hacia `/ministerios` con ítems; **el bloque no aparece** sin publicados | Pruebas del service/repository (tope 5, orden, `items` omitida) + e2e `portada-ministerios.spec.ts` (quickstart §14) | 100 % / 0 |
| SC-017 | Un ministerio **sin enlace y sin encargado mostrable** se presenta completo y coherente (sección y bloque) | e2e con esa fixture (quickstart §7) + pruebas de la ficha (sin botones vacíos, sin líneas vacías, sin huecos) | 0 / 100 % |

## Riesgos

> Numeración **RG6-1…RG6-13**, propia de este plan (para no colisionar con RG4-1…RG4-14 de F4,
> RG3-1… de F3 ni R6-1…R6-17 de `research.md`). **RG6-13 es nuevo** (siguiente número libre, nacido
> con P6-21 al realinear el plan con la puerta 2); **RG6-10 quedó resuelto** por el humano el
> 2026-10-13 y se conserva su identificador como registro.

| Riesgo | Impacto | Mitigación |
|---|---|---|
| **RG6-1 — Fuga de datos del encargado (FR-026/SC-014, crítico)** | El requisito no negociable de F6 falla: un celular/correo personal se publica sin permiso; `seguridad` rechaza el PR | Garantía en **tres capas** (P6-2/R6-2): DTO público sin campos para el dato oculto, consultas públicas con `CASE WHEN leader_visible` (la fila sale con `NULL`) y bloque de portada sin campos de encargado; pruebas "0 apariciones" en las **cuatro** capas (service/handler/repository/e2e con inspección de red y HTML); inventario de vías de salida público/privadas para la auditoría de `seguridad`; el mapeo público se compone en `service_public.go` (un solo lugar) |
| **RG6-2 — Incoherencia entre sección, bloque y navegación** (un "publicado" para unos y otro para otros) | SC-010/SC-016 fallan: entrada de navegación hacia una sección vacía, o bloque con huecos | **Un mismo predicado** "publicado" y la misma consulta (`ListPublishedMinistries`, solo con `LIMIT` distinto) alimentan sección y bloque; la navegación bebe del documento del bloque (P6-9); pruebas unitarias con los mismos fixtures en las tres superficies y e2e que comprueba "entrada visible ⇔ bloque pintado ⇔ hay ítems" |
| **RG6-3 — Enlace no seguro por alguna vía** (`javascript:`/`data:` aceptado o renderizado) | SC-009 y constitución §IV (CWE-79) fallan | Validación en el service (tag `url` + sin espacios/control) + `CHECK` en la BD + render como atributo `href` (sin HTML interpretado) + tabla de casos con esquemas hostiles y e2e de rechazo; el valor nunca se transforma |
| **RG6-4 — `ux.md` se produce en paralelo** (composición de ficha, listado y bloque) | Tareas de frontend que componen pantallas que no encajan con los DTOs | Este plan fija rutas, DTOs, reglas de contenido y posiciones: el diseño puede variar la composición **sin cambiar la API** (mismo encaje que F4, RG4-5); los nombres de pieza propuestos son provisionales hasta que `ux.md` los confirme. La puerta 2 (2026-10-13) ya **cerró las dos divergencias de fondo**: la presencia de la navegación se deduce del documento del **bloque** (criterio técnico de este plan, confirmado) y el interruptor nace **apagado**; la redirección de `/ministerios` sustituye al aviso `HiddenNotice` (ver P6-21) |
| **RG6-5 — `IsFilePublished` toca código de F3/F4** (rama nueva para `ministries`) | Regresión en la descarga del logo, la portada o las imágenes de agenda | La consulta crece de forma **aditiva** (una tercera rama `OR`); las pruebas de `internal/media` existentes se **conservan** y solo se añaden casos de ministerio publicado/borrador; los e2e de F3/F4 (`portada-imagenes`, `agenda-*`) siguen corriendo como evidencia |
| **RG6-6 — Pruebas de frontend que fijan la lista de permisos y accesos** (F2/F4, funcionalidades cerradas) | `make ci` falla al añadir `MINISTERIOS`, o se lee como "debilitar" pruebas ajenas | Cambio **acotado y documentado** (mismo gesto que F4 hizo con los suyos): `lib/permissions.test.ts` (lista de `MODULE_PERMISSIONS`), `features/roles/permissions.test.ts` y `roles.test.tsx` (el sello «Disponible más adelante» deja de aplicarse a `ministerios`) y `features/panel/panel.test.tsx` (tarjeta nueva en accesos rápidos); **sin** cambios de aserciones de comportamiento en backend (el catálogo de permisos no cambia) |
| **RG6-7 — Formato del celular** (el tag `phone` acepta `+` y separadores; el equipo usa formatos locales) | Un celular legítimo se rechaza (o uno malformado pasa) | Regla de formato, no de existencia (default revisable de la spec); límites 7–32 iguales a los de F2; el error es por campo y con mensaje claro; si el equipo necesita otro formato, se ajusta el tag en `platform/validate` (cambio menor y testeable) |
| **RG6-8 — Duplicados con variantes de redacción** ("Alabanza" / "alabanza " / " Ministerio de alabanza") | Falsos negativos (duplicados que pasan) o falsos positivos (bloqueos indebidos) | Normalización documentada (`btrim` + minúsculas) igual que F2/F3/F4; el `409` del duplicado exacto es un default revisable y su regla está en el contrato; el enlace compartido **nunca** bloquea (solo avisa, R6-11) |
| **RG6-9 — Accesibilidad y usabilidad (SC-008/SC-012) no cubiertas por automatización** | Hallazgos críticos en la revisión manual | Checklist explícita en `quickstart.md` §11 ejecutada por `qa-tester`; componentes semánticos de F10 (ya auditados); `alt` obligatorio en la BD; protocolo de usabilidad con personas (SC-012) como en F3/F4 |
| **RG6-10 — Default del interruptor `leaderVisible`** *(RIESGO RESUELTO por el humano el 2026-10-13; se conserva el identificador)* | — *(era: un ministerio nuevo publicaba u ocultaba la ficha del encargado sin que el equipo lo hubiera elegido)* | **Decisión explícita del humano: `FALSE`, «Oculto por defecto»** (R6-3; la recomendación de este plan —`TRUE`— quedó **descartada**). Mitigación vigente: `leader_visible DEFAULT FALSE` en `000010` (también para `INSERT` sin la columna), `default: false` en `MinisterioInput` (el service aplica `false` si falta), interruptor apagado al crear (Q1) y **prueba del default en las cuatro capas** (service/handler/repository/e2e) |
| **RG6-11 — `whatsappUrlShared` obsoleto tras un guardado de otro ministerio** | El aviso "este enlace ya lo usa otro" no se muestra en el momento exacto | El flag se **recalcula en cada lectura** (`EXISTS` sobre la tabla), no se persiste: siempre refleja el estado actual; es un aviso, no una regla (R6-11) |
| **RG6-12 — Muchos ministerios publicados** (sin límite artificial, default revisable) | Respuestas grandes en móvil | Constante de seguridad `maxPublicItems = 200` en el service (documentada en el contrato, como `maxBlockItems` de F4) y orden estable; si el volumen real lo exige, agrupar o "ver más" es el ajuste menor que la propia spec deja acordado con el humano |
| **RG6-13 — Destello, callejón o bucle en la redirección de `/ministerios`** *(riesgo **nuevo**, RG6-13 = siguiente número libre, nacido con P6-21 tras la puerta 2)* | SC-010/SC-008: un visitante con un enlace guardado ve un destello de sección vacía, queda sin explicación, o «Atrás» lo atrapa en un bucle de redirección | Mecánica fijada en **P6-21**: `SectionSkeleton` mientras la consulta está en vuelo (**nunca** el estado vacío), redirección **solo** con éxito y sin `items`, con `<Navigate to="/" replace />` (`replace` = «Atrás» sin bucles); **ante error NO se redirige** (estado de error con «Volver a intentar»); e2e que comprueba que en ningún instante se pinta la grilla vacía ni un aviso, que la URL termina en `/`, que la portada carga completa y que «Atrás» no devuelve a `/ministerios`; el destino tiene toda la navegación (el enlace guardado no muere en un callejón) |

## Complexity Tracking

> Sin violaciones de la constitución. Desviaciones de **convenciones internas**, declaradas:

1. **Nombre de la tabla** (P6-1): la pauta de F3/F4 es prefijo de dominio (`home_*`, `agenda_*`)
   porque son agregados de varias tablas; F6 tiene **una** tabla y sigue la pauta de F2
   (`users`, `roles`): **`ministries`**, en inglés y plural, con la carpeta de dominio en español
   (`internal/ministerios/`, como `usuarios`/`portada`/`agenda`). Si `revisor-codigo` prefiere
   `ministry_*`, es un renombrado sin consecuencias (todo el cableado está en `main.go`).
2. **La presencia de la navegación bebe del documento del bloque** (P6-9): `useSiteLinks` consume
   `GET /api/v1/portada/ministerios` en lugar del documento completo de la sección. Es una
   desviación menor de "la presencia se deduce de la misma respuesta que pinta la sección" (D-23 de
   F4), justificada porque ambas miran el **mismo predicado** y así la portada no descarga la lista
   completa para pintar un enlace (documentado en R6-8 y en RG6-2). **Criterio confirmado por el
   humano en la puerta 2 (2026-10-13)** frente a la propuesta de `ux.md` §3.
3. **Mapeo `audit.Action` → `InsertAdminActionParams` repetido** (R6-12): ~10 líneas en el
   repository de `ministerios` y otras tantas en `portada`/`agenda`/`usuarios`, todas sobre la
   **misma** consulta nombrada de `internal/db` (misma decisión que F3/F4; subirlo a `platform/`
   arrastraría `pgtype` fuera del diagrama de dependencias).
4. **`/api/v1/portada/ministerios` lo registra el dominio `ministerios`** (P6-16): el segmento
   `portada` pertenece al agregado de F3 pero el recurso es de este módulo (mismo gesto que la
   Complexity 8 de F4 con `proximamente`); queda anotado en la cadena de rutas y en el contrato.
5. **Reordenar se registra como `ministry.update`** (R6-12): la spec enumera "reordenar" como
   operación de auditoría y lo es (queda registrada, quién/qué/cuándo), pero sin código propio,
   igual que en F3/F4. Si el humano prefiere `ministry.reorder`, es un código más en la misma
   lista cerrada.
6. **Nombres de pieza de UI según `ux.md`** (RG6-4): `MinisteriosPage`, `MinistryCard`,
   `WhatsappJoinButton`, `EncargadoLine`, `MinisteriosBlock`, `MinistryItem`, `MinistryForm`,
   `ContactoEncargado`, `ShowContactSwitch`, `MinistriesList`… son los del inventario de `ux.md`
   §6.3/§6.4 (ya producido en paralelo); este plan fija rutas, DTOs y reglas y toma esos nombres
   como referencia de `tasks.md`, igual que F4 hizo con los suyos. Cualquier ajuste de composición
   posterior se aplica en las tareas `[frontend]` **sin** tocar la API.
7. **F6 se aparta a propósito del patrón de F4 en la sección sin contenido** (P6-15/P6-21): F4
   muestra un aviso (`HiddenNotice`) en `/eventos` **sin redirigir** (R4-13) y F6 **redirige**
   `/ministerios` a la portada. No es una inconsistencia accidental sino la decisión literal del
   humano del 2026-10-13: F4 **no** se toca y la divergencia queda anotada en R6-13, P6-15, P6-21,
   el contrato y el quickstart. Consecuencia de inventario: `HiddenNotice` **no** se promueve.

## Propuesta de tareas (insumo para `/speckit.tasks`; **no** es `tasks.md`)

Orden típico migración → contrato → repositorio → servicio → handler → frontend → e2e. Las capas se
reparten `[db]`/`[backend]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` → `devops`;
las marcadas `[P]` son independientes entre sí y pueden ejecutarse en paralelo.

1. `[db]` Migración `000010` (tabla `ministries` + índices, con `leader_visible DEFAULT FALSE`
   —«Oculto por defecto», decisión de la puerta 2 ya aplicada, P6-3—) con su `down` y prueba
   up/down/up. `[P]`
2. `[db]` Migración `000011` (`admin_actions_action_check` + 8 códigos `ministry.*`) con su `down`
   y prueba up/down/up. `[P]`
3. `[backend]` Fusión del delta en `backend/api/openapi.yaml` (`info.version` → **0.7.0**; enum
   `AdminActionItem.action` +8; prosa de `GET /api/v1/media/{fileName}`) + `make api-gen`. `[P]`
4. `[backend]` `platform/audit`: 8 códigos + pruebas del registro cerrado. `[P]`
5. `[backend]` Consultas `ministries.sql` + rama nueva de `media.sql` (`make sqlc`) y repository de
   `ministerios` con pruebas de integración (incluidos los `CHECK` de contacto/enlace/español por
   campo, `ListPublishedMinistries` con `CASE WHEN leader_visible`, `whatsapp_url_shared`,
   `ExistsMinistryDuplicate`, `IsFilePublished` con ministerios y la atomicidad con auditoría).
   *Depende de 1, 2 y 4.*
6. `[backend]` Service público de `ministerios` (sección + bloque con `maxPublicItems`/`maxPortadaItems`,
   idioma con fallback, **regla de privacidad del encargado**) + pruebas. *Depende de 5.*
7. `[backend]` Service de panel (CRUD, validaciones —contacto excluyente, enlace https, español por
   campo, imagen+alt—, duplicados `409`, `whatsappUrlShared`, publicar/retirar, visibilidad del
   encargado, auditoría) + pruebas. *Depende de 5.*
8. `[backend]` Handlers y `routes.go` (`RegisterPublic` con las 2 rutas; `RegisterAdmin` con
   `AdminChain("ministerios")`) + pruebas + humo en `main_test.go`. *Depende de 6 y 7.*
9. `[frontend]` Cliente `api/ministerios.ts` + `schema.d.ts` regenerado + hooks. *Depende de 3.*
   `[P]` respecto a 5–8 (avanza contra el contrato).
10. `[frontend]` Página pública `/ministerios` (`MinisteriosPage`, `MinistryCard` con
    `CardImage`/`EncargadoLine`/`WhatsappJoinButton`, **redirección a la portada sin publicados**
    —P6-21: `SectionSkeleton` → `<Navigate to="/" replace />`, sin aviso vacío y sin redirigir ante
    error—, i18n de `ux.md` §7.2) con la ficha completa, **sin botón sin enlace** y **sin línea de
    encargado sin `leader`**. *Depende de 9.*
11. `[frontend]` `useSiteLinks` con la entrada «Ministerios» condicional + `MinisteriosBlock` con
    sus `MinistryItem` en `HomePage` (tras `ProximamenteBlock`, antes de `WhoWeAreSection`) + i18n.
    *Depende de 9.*
12. `[frontend]` Panel `features/ministerios/` (`MinistriesPage`, `MinistryForm` con
    `ContactoEncargado` y `ShowContactSwitch` —**apagado al crear**, «Oculto por defecto»—, aviso
    `whatsappUrlShared`, `MinistriesList` con
    `StatusPill`), ruta `/panel/ministerios` con `RequirePermission(MINISTERIOS)`,
    `lib/permissions.ts`, `features/roles/permissions.ts`, `AccesosRapidos` y `ACTION_LABELS`
    (+8) + actualización acotada de las 4 pruebas que fijan permisos/accesos (RG6-6). *Depende de 9.*
13. `[…]` e2e `ministerios-publica.spec.ts` (incluye **ficha oculta**, **ministerio sin enlace**,
    **redirección a la portada sin publicados** y **alta sin `leaderVisible` nace oculto**),
    `ministerios-panel.spec.ts` (incluye cruce de permisos y rechazo de enlaces), `ministerios-
    accesibilidad.spec.ts` y `portada-ministerios.spec.ts` + `ministeriosSeed.ts` + checklist de
    quickstart. *Depende de 10–12.*
14. `[…]` Validación (qa/revisor/seguridad — con foco en SC-014 —), documentación (documentador) y
    PR. *Depende de 13.*

## Decisiones del humano en la puerta 2 (2026-10-13) — **resueltas**

Los cuatro puntos que este plan señaló para la puerta 2 quedaron **resueltos por el humano el
2026-10-13**; el plan está realineado a ellos y **el resto quedó aprobado tal como está**.

1. **Default del interruptor `leaderVisible` → `FALSE`, «Oculto por defecto»** (R6-3/RG6-10/P6-3):
   el ministerio nuevo **nace con la ficha del encargado no publicada** —nombre y contacto— y nada
   sale al público hasta que el equipo encienda el interruptor (**privacidad por defecto**: publicar
   un celular/correo personal exige un acto afirmativo). **La recomendación de este plan era la
   contraria (`TRUE`, opción (a) de R6-3) y quedó descartada.** Aplicado en: `000010`
   (`leader_visible DEFAULT FALSE`), `MinisterioInput` (`default: false`; el service aplica `false`
   si falta), formulario del panel (interruptor **apagado** al crear) y **todas las pruebas
   previstas** que asumían el default visible (service/handler/repository/e2e). El interruptor es
   **propio de F6** (F3/F4 no lo tienen).
2. **`/ministerios` sin nada publicado → redirige a la portada** (R6-13/P6-15/**P6-21**): decisión
   del cliente **distinta a propósito** de F4 en `/eventos` (que muestra aviso sin redirigir y **no
   se toca**). Resuelta en el plan: la decisión vive en el **frontend** (elemento de ruta de
   `/ministerios`), la mecánica evita el parpadeo (`SectionSkeleton` mientras carga →
   `<Navigate to="/" replace />` **solo** con éxito y sin `items`; **sin redirección ante error**),
   la API sigue respondiendo **`200` con `items` omitida** (contrato coherente, sin `3xx`), la
   entrada de navegación **sigue oculta** (barra, hamburguesa y pie) y la accesibilidad queda
   cubierta porque el destino es una página completa con toda la navegación.
3. **Lectura de Q6 (horario y lugar como textos libres, sin recurrencia ni calendario)**: ya
   **confirmada por el humano el 2026-10-13** (estado.md); se implementa así (sin cambios).
4. **`ux.md` y sus decisiones D-1…D-12 (§9)**: la composición visual la fija `disenador-ux` sin
   cambiar rutas ni DTOs. Sus propuestas de privacidad in-UI (§4.5: `ShowContactSwitch` con
   leyendas y `Notice` warning al ocultar la ficha) son coherentes con este plan. Los puntos donde
   `ux.md` y este plan decían cosas distintas quedaron **dirimidos** (y `ux.md` se actualiza en
   paralelo):
   - **Presencia de la navegación** (`ux.md` §3, D-23): este plan la deduce del documento del
     **bloque** (`GET /api/v1/portada/ministerios`, P6-9/R6-8) —**criterio técnico confirmado por
     el humano**— en lugar del documento completo de la sección: mismas garantías (mismo predicado
     "publicado") y **0 peticiones extra** en la portada. Cambiar de una a otra es un hook; no toca
     la API.
   - **Default del interruptor** (`ux.md` §4.5/§5.b lo proponía **encendido** al crear): queda
     **apagado** («Oculto por defecto», punto 1).
   - **Sección oculta** (`ux.md` §3/§4.1/§6.2/§6.3 mostraban la página mínima de `HiddenNotice`
     sin redirigir): queda **redirección a la portada** (punto 2) y `HiddenNotice` **no** se
     promueve al inventario único.

**Defaults revisables ya aplicados por el arquitecto** (documentados en `research.md` y en el
contrato) **y aceptados por el humano el 2026-10-13**: duplicado exacto → `409`; enlace de
WhatsApp compartido → aviso no bloqueante (`whatsappUrlShared`); reordenar se audita como
`ministry.update`; límites de extensión (nombre/encargado 160, descripción 2.000, horario/lugar
200, enlace 500, alt 300); `maxPublicItems = 200` en la sección; eliminación definitiva con
confirmación y sin papelera; validación del contacto solo de formato; nombre del interruptor
`leaderVisible` (equivale al `showContact` de la spec).

**Requisito que la spec aún no cubre** (señalado al `analista-producto`, **no** implementado sin su
línea en la spec): un anuncio breve accesible (`aria-live`) al redirigir `/ministerios` → portada
ayudaría a quien usa lector de pantalla a entender el salto. Tocaría superficies de F3/F10 y
FR-021 limita el contenido de F6 en la portada al bloque; sin él la accesibilidad ya queda
razonablemente cubierta (el destino es una página completa navegable), por lo que **no** entra en
este plan.
