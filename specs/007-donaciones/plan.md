# Implementation Plan: Ofrendas y diezmos (F7 · Donaciones)

**Branch**: `007-donaciones` | **Date**: 2026-10-14 | **Spec**: [spec.md](./spec.md)
(aprobada **v2** por el humano el 2026-10-14 con Q1–Q11 resueltas —decisión 22 del roadmap—:
sección pública **«Ofrendas y diezmos»** en **`/ofrendas`** con contenido **agrupado por tipo**
cuentas IBAN y SINPE Móvil, dos textos editables —introductorio y de uso— con estado propio,
bloque de portada **debajo del bloque de ministerios (F6)**, y módulo/permiso del panel intacto
**«donaciones»**; 6 US, 25 FR, 14 SC).

**Input**: `specs/007-donaciones/spec.md` (fuente de verdad) · `.specify/memory/constitution.md`
· `AGENTS.md` · `docs/tecnico/arquitectura.md` (reglas R1–R8 y convenciones §8.1) ·
**F2 terminada** (`specs/002-acceso-gestion-usuarios/`: patrón de permisos por módulo, sesión,
CSRF, `authz`, `apperr`, `validate`, `paginate`, auditoría y catálogo de permisos **con
`donaciones` ya sembrado** — verificado en `000002_create_roles_and_permissions`, línea 25:
`('donaciones', 'Donaciones')`; el catálogo **no cambia** en F7) · **F3 terminada**
(`specs/003-portada-info-general/`: patrón bilingüe `*_es`/`*_en`, publicación por elemento,
singletons `PUT` de la portada, i18n del público, ocultar secciones vacías) · **F4 terminada**
(`specs/004-eventos-actividades/`: módulo de contenido con dominio propio, `AdminChain` por
permiso, bloque de portada con endpoint propio, `sort_order`) · **F6 terminada**
(`specs/006-ministerios/`: **precedente más directo** — bloque de portada, i18n por campo,
`publication_state`, `sort_order`, auditoría y **`/ministerios` redirigiendo a la portada sin
aviso** (decisión 20b) —) · **F10 terminada** (sistema de diseño, modo oscuro, movimiento,
navegación y selector de idioma: se **aplica**, no se redefine) · código real de `backend/` y
`frontend/` · skills `go-backend`, `postgres-db`, `react-frontend`.

## Summary

F7 es la sección **«Ofrendas y diezmos»** del sitio público: cómo apoyar económicamente a la
iglesia. Un **método de donación** es una **cuenta bancaria (IBAN)** o un **número de SINPE Móvil**
—varias de cada una, en **un listado administrable agrupado por tipo** (Q1: primero IBAN y después
SINPE, con **orden manual dentro de cada grupo**)— con **tipo, número y entidad financiera
obligatorios**, **titular opcional** (si se rellena, se publica) y **moneda (colones o dólares)
solo en las cuentas IBAN** (Q2); sin notas, sin imágenes (Q6). Junto a los métodos, **dos textos
libres editables** de uno o varios párrafos: el **introductorio** («cómo donar», Q11) y el **de uso
de las donaciones** (Q5), cada uno con **su propio estado de publicación** (Q10). La sección vive
en **`/ofrendas`** con una única entrada de navegación «Ofrendas y diezmos» (FR-006), bilingüe con
fallback al español (Decisiones 6/8), nacida con el sistema de F10 y **oculta por completo sin
nada publicado** (FR-005/SC-010): la entrada de navegación desaparece y la **URL directa
`/ofrendas` redirige a la portada** sin aviso (Q9, patrón de F6 en `/ministerios`, decisión 20b).
Además la **portada** suma su bloque de ofrendas y diezmos —hasta **5** métodos publicados en el
orden de la sección, con enlace a `/ofrendas`— **inmediatamente debajo del bloque de ministerios
de F6** (Q7, FR-022/SC-014), servido por su propio endpoint público acotado y **oculto por
completo sin métodos publicados** (como el bloque de F6 y **sin** el mensaje del bloque de F4).

Técnicamente F7 añade: **dos tablas** —`donation_methods` (el listado administrable) y
`donation_texts` (los dos textos, un renglón por clave `intro`/`usage` con `UNIQUE`)—, la
**validación y normalización de los números** en el dominio (`internal/donaciones/numero.go`: IBAN
por **estructura y longitud exacta por país** —costarricense e internacional, registro IBAN
embebido, **sin dependencias nuevas**—; SINPE = **celular costarricense de 8 dígitos** con o sin
`+506`, guardado `+506XXXXXXXX`), el patrón de autorización de F2 con
`AdminChain("donaciones")` (**sin permisos nuevos ni migraciones de permisos**: el permiso ya está
sembrado por `000002`), 8 códigos de auditoría `donation.*` (migración `000013`) y **cero cambios
en `media`/`IsFilePublished`** (Q6: F7 no tiene imágenes). El frontend publica **`/ofrendas`**
dentro de F10, añade `/panel/donaciones` bajo `RequirePermission(DONACIONES)`, la entrada de
navegación condicional (fuente única `useSiteLinks`, que bebe del **documento de la sección**) y
el bloque `OfrendasBlock` en `HomePage` **como una inserción** tras `MinisteriosBlock` (el resto
de la portada de F3 no se toca, FR-023).

**Ninguna dependencia nueva** ni en backend ni en frontend.

**Cuatro puntos que la spec deja en el aire o se contradice** quedan resueltos como **defaults
revisables** y se elevan al humano en «Preguntas para el humano» (al final): (1) la **entidad
financiera** es traducible según FR-019/Key Entities y **no** traducible según FR-007/US4/SC-006
—se implementa **no traducible**, única lectura que garantiza SC-006—; (2) qué significa exactamente
"**nada publicado**" cuando hay **solo textos** publicados y 0 métodos; (3) si el IBAN debe además
verificar **dígitos de control** (mod-97), que la spec no pide; (4) si el SINPE debe exigir que el
primer dígito sea 6/7/8 para distinguir **celular** de línea fija.

## Technical Context

**Language/Version**: Go 1.27 (sin cambios) · TypeScript 5.x `strict` sobre Node 22 · React 19.

**Primary Dependencies**: **ninguna nueva**. Se reutiliza todo lo existente: `pgx`/`sqlc`,
`testcontainers-go`, `net/http`; en frontend React Router, TanStack Query, React Hook Form + Zod,
Tailwind, Vitest + Testing Library + MSW y Playwright (R7-16). La validación del IBAN es una
regex + una tabla de longitudes por país (~90 constantes, registro IBAN/ISO 13616) y la del SINPE
una regex + la normalización del prefijo: **sin librerías** (R7-2/R7-3).

**Storage**: **PostgreSQL 16** (tablas `donation_methods` + `donation_texts` + ampliación del
`CHECK` de `admin_actions`; migraciones `000012`/`000013`, ver [data-model.md](./data-model.md)) ·
Redis 7 **sin tocar** (sesión de F2) · **sin almacenamiento de archivos** (Q6/FR-020: F7 no
tiene imágenes y **no toca** `platform/storage`, `internal/media` ni `IsFilePublished`). Capa de
datos sqlc (D-A3): consultas en `backend/internal/db/queries/donation_methods.sql` y
`donation_texts.sql`, código generado commiteado; la auditoría reutiliza `InsertAdminAction` de
`queries/audit.sql` **dentro de la transacción** de cada mutación. **Sin variables de entorno
nuevas**.

**Testing**: estrategia por capa de F1/F2/F3/F4/F6 ampliada (sección propia): `go test` (unitarias),
`go test -tags=integration` (repository contra PostgreSQL real —incluida la migración up/down/up de
`000012`/`000013`, los `CHECK` de formato por tipo, moneda y español por campo, y la regla "0
borradores expuestos"—), Vitest + Testing Library + MSW y Playwright (e2e pública, de panel, de
accesibilidad y de la portada). **Sin excepciones aprobadas**: F7 no toca pruebas de F2 en el
backend (el catálogo de permisos **no** cambia); solo actualiza las **2–3 pruebas de frontend** que
fijan la disponibilidad de permisos y los accesos del panel (RG7-6, mismo gesto que F6/RG6-6).

**Target Platform**: Docker Compose local + GitHub Actions (CI del kit, sin tocar). Sin despliegue a
producción (fuera de alcance de F7).

**Performance Goals**: `GET /api/v1/ofrendas` < 500 ms con BD sana (dos consultas: métodos +
textos, `no-store`, acotada a `maxPublicItems = 200`) · `GET /api/v1/portada/ofrendas` < 500 ms y
**payload acotado** (≤5 ítems, `LIMIT 5` en la misma consulta de la sección) · listados de panel
paginados (`limit`/`offset`, §8.1.3). Sin subidas de archivo.

**Constraints**: stack fijo (constitución) · capas `handler → service → repository` (§II) ·
contrato OpenAPI antes que el código (§II) · migraciones versionadas e inmutables (§VI) ·
validación de toda entrada en backend (§IV, CWE-20) · SQL siempre parametrizado (§IV, CWE-89) ·
nada de contenido ejecutable (§IV, CWE-79: sin `dangerouslySetInnerHTML`) · authn/authz/CSRF en
servidor (§IV, CWE-862) · dependencias minimizadas (0 nuevas) · archivos del kit no editables
(regla 10) · **la spec no se reabre**: Q1–Q11 están cerradas; los cuatro huecos señalados arriba
son defaults revisables que se documentan y se elevan, **no** interpretaciones silenciosas.

**Scale/Scope**: 9 operaciones REST (**2 públicas** · 7 de panel) sobre **2 tablas nuevas** + la
ampliación del `CHECK` de `admin_actions` (8 códigos); 1 dominio backend nuevo (`donaciones`) y
**`media`/`platform` sin cambios de comportamiento** (solo `platform/audit` crece con 8
constantes); 2 migraciones (`000012`, `000013`) y **ninguna de permisos**; frontend: página
pública nueva en `/ofrendas` + **bloque de ofrendas en la portada** + feature de panel
`donaciones` (métodos y textos) + ajustes de router/navegación/permisos/i18n. Sin puertas de
aprobación abiertas más allá de **este plan** (regla 1 de AGENTS.md).

## Constitution Check

*GATE: debe pasar antes de la investigación y re-verificarse tras el diseño.*

| Principio | Evaluación | Resultado |
|---|---|---|
| §I La spec manda | El plan implementa FR-001…FR-025 e US1–US6 sin interpretarlos; cada uno tiene fila en "Cobertura de requisitos". Q1–Q11 están cerradas y no se reabren; nada de Out of Scope se construye (sin pasarelas de pago, sin montos/metas/estadísticas, sin registro de donantes, **sin imágenes** (Q6), sin campo de notas, sin QR ni enlaces de cobro, sin historial ni papelera, sin bilingüismo del panel, sin rediseñar la portada de F3 ni F10B) y el permiso `donaciones` solo se **activa** (FR-011). Los cuatro huecos de la spec (entidad traducible, "nada publicado" con solo textos, mod-97 del IBAN, prefijo del celular) se resuelven como **defaults revisables documentados** y se elevan al humano: no son reglas de negocio inventadas sino lecturas literales de FR-007/FR-015/FR-005/Q4. | ✅ |
| §II Arquitectura | Monorepo; capas `handler → service → repository` sobre `internal/platform/` (R1–R8); API REST JSON documentada en `backend/api/openapi.yaml` con el delta redactado **antes** que el código (`contracts/openapi.yaml`, §8.1.5); **0 dependencias nuevas** (R7-16). | ✅ |
| §III Pruebas | Toda tarea llevará sus pruebas (se exigirá en `tasks.md`); cobertura ≥80 % en `service/`; repository contra PostgreSQL real (`//go:build integration`), incluidas las migraciones `000012`/`000013` (up/down/up), los `CHECK` de formato por tipo/moneda/español por campo y la regla "0 borradores expuestos"; frontend con Vitest + Testing Library + MSW; flujos críticos con Playwright (sección pública, números exactos, publicar/retirar de métodos y textos, permisos, bloque de la portada y **redirección a la portada sin nada publicado**). Las 2–3 pruebas de frontend que fijan permisos/accesos se actualizan de forma acotada y documentada (RG7-6). | ✅ |
| §IV Seguridad | SQL solo parametrizado vía sqlc (CWE-89); validación de toda entrada en backend (CWE-20: los números se validan y normalizan **en el servidor**, el formulario solo acompaña); texto plano sin ejecución (CWE-79: sin `dangerouslySetInnerHTML`; en los números no caben emojis ni letras fuera de estructura); operaciones de panel con sesión + guard + CSRF + permiso **`donaciones`** verificado en servidor (CWE-862); borradores **nunca** hacia el público (consultas `…Published`; SC-002) y nada de imágenes que servir (Q6); secretos: ninguno nuevo; `govulncheck`/`npm audit` en el CI del kit. | ✅ |
| §V Calidad | `gofmt`/`go vet`/`golangci-lint` sin errores; TS `strict` sin `any`; errores envueltos con `%w` y traducidos solo en `WriteError`; nombres de código en inglés (dominio `donaciones` sigue la pauta de F2/F3/F4/F6: carpeta de dominio en español + tablas en inglés). | ✅ |
| §VI Base de datos | 2 migraciones versionadas con `up`/`down` completos (`000012`, `000013`); ninguna migración existente se modifica (**ni `000002`**: no hay permisos nuevos); tablas con `id`, `created_at`, `updated_at`; índices explícitos; `CHECK`/`UNIQUE` en la base (formato por tipo, moneda por tipo, español por campo, longitudes, catálogo de `text_key`, estado). | ✅ |
| §VII Observabilidad | Logs `log/slog` con `request_id` (incluidos los fallos de validación, sin cuerpos de petición); auditoría duradera ampliada a F7 en `admin_actions` (FR-017, 8 códigos); `/healthz` intacto y sin sesión; config por variables de entorno (**sin variables nuevas**); `docker compose up` levanta todo (sin servicios nuevos). | ✅ |
| §VIII Gobierno | **Este plan requiere aprobación humana** antes de `/speckit.tasks` y de cualquier código (regla 1), incluidos los cuatro puntos de «Preguntas para el humano». Sin despliegue a producción. Quien escribe no aprueba: todo pasa por `qa-tester`, `revisor-codigo` y `seguridad`. | ✅ |

**Sin violaciones que justificar** → "Complexity Tracking" solo registra desviaciones de
convenciones internas (presencia de la navegación desde el documento de la sección, naming
público/panel distinto a propósito, y el mapeo de auditoría repetido).

## Decisiones técnicas (resumen; análisis completo en `research.md`)

Numeración **P7-1…P7-21**, propia de este plan (no confundir con P6-1…P6-21 de F6, P4-1…P4-20 de
F4, R7-1…R7-16 de `research.md` ni RG7-1… de "Riesgos").

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| **P7-1** | **Un dominio `internal/donaciones/`** con **dos tablas** —`donation_methods` (listado administrable) y `donation_texts` (singleton por clave `intro`/`usage`)— y consultas en `queries/donation_methods.sql`/`donation_texts.sql` | Key Entities define tres cosas editables con ciclos de vida distintos (lista ordenada + dos singletons); los dos textos tienen la misma forma y comparten tabla (R7-1) | Dos tablas singleton idénticas (`home_about` ×2); una tabla para todo con columnas anulables por uso; EAV/traducciones (rompe R3-1); JSON (pierde `CHECK` y sqlc) |
| **P7-2** | **Números: validación y normalización en el dominio** (`internal/donaciones/numero.go`, funciones puras con tabla de casos) + `CHECK` por tipo en la BD; **sin tags nuevos** en `platform/validate` | IBAN y SINPE son formatos de negocio de este módulo (los tags de `platform` son genéricos: R7-2/R7-3); la BD refuerza la forma | Tags `iban`/`sinpe` en `platform/validate` (solo los usaría este módulo); librerías de IBAN/teléfonos (dependencias nuevas sin justificación); validar solo en el frontend (§IV) |
| **P7-3** | **IBAN: estructura + longitud exacta por país** (registro IBAN embebido, `CR` = 22), guardado **normalizado a los caracteres del número**, presentación agrupada en bloques **en el frontend** (`formatIbanDisplay`), botón de copiar = valor exacto | Q3/FR-015 piden literalmente "estructura y longitud", guardado normalizado y presentación en bloques sin alterar el valor (SC-009) | Mod-97 (regla no pedida; **pregunta al humano**); solo rango 15–34 (aceptaría `CR` con 30 caracteres); guardar con espacios (rompe normalización y duplicados); agrupar en el servidor (FR-015 prohíbe transformar el valor publicado) |
| **P7-4** | **SINPE: celular CR de 8 dígitos** (primer dígito 6/7/8) con o sin `+506`; guardado y publicación en **`+506XXXXXXXX`** | Q4 pide "celular costarricense de 8 dígitos" y "prefijo normalizado al guardar"; el prefijo presente es la forma autónoma para quien dona desde el extranjero (US4) | Guardar solo 8 dígitos (el cliente tendría que inventar el prefijo); guardar "tal como se escribió" (rompe normalización); aceptar cualquier 8 dígitos (**pregunta al humano**); librería de teléfonos |
| **P7-5** | **`institution` NO traducible** (una sola columna) y **`holder` traducible** (`holder_es`/`holder_en`) | Única lectura que garantiza SC-006 "0 alteraciones por cambio de idioma" (FR-007/US4 esc. 5); la contradicción de FR-019/Key Entities queda señalada (**pregunta #1**) | `institution_es`/`institution_en` (SC-006 se volvería incumplible si alguien rellena el inglés); traducir todo (contradice FR-007) |
| **P7-6** | **Permisos**: `AdminChain("donaciones")` con el permiso **ya sembrado** en `000002`; **sin migración de permisos** ni cambios en F2 (el catálogo sigue en 10) | FR-011 literal ("F7 NO crea permisos nuevos"); `AdminChain`/`AuthzByModule` bastan (mismo gesto que F3/F4/F6) | Permiso nuevo por sub-acción; renombrar el permiso a `ofrendas` (Q8/FR-015 de F2 lo prohíben); derivar el acceso de `portada`/`ministerios` |
| **P7-7** | **Superficie HTTP**: `GET /api/v1/ofrendas?lang=` (sección) + `GET /api/v1/portada/ofrendas?lang=` (bloque, tope 5), ambos `no-store`; panel `/api/v1/admin/donaciones…` con `limit`/`offset`, filtro `publicationState` y `PUT` para los textos | Q8 separa los nombres a propósito: el público nombra la **sección** (`ofrendas`, espejo de `/ofrendas`) y el panel el **módulo** (`donaciones`, que es como se llama el permiso) (R7-7) | Nombrar todo `donaciones` (el público prometería un nombre que Q8 reserva al panel); nombrar todo `ofrendas` (el permiso/módulo es `donaciones`); un solo endpoint para sección y bloque (el tope 5 es de servidor) |
| **P7-8** | **Presencia de la navegación desde el documento de la sección** (`usePublicOfrendas` compartido con `useSiteLinks`), **no** del bloque | La sección puede existir con **solo textos** publicados (R7-12) y el bloque solo mira métodos (FR-022): solo el documento de la sección decide sin contradicción (patrón F4/D-23) | Presencia desde el bloque (P6-9 de F6: aquí dejaría la navegación huérfana con solo textos); un endpoint "presencia" (una ruta pública más por un booleano) |
| **P7-9** | **Bilingüe** `*_es`/`*_en` por campo solo donde hay traducción (titular y textos) + `?lang=es\|en` resuelto en el servidor con fallback `en → es` | FR-007/FR-008/SC-006 garantizados en una capa reutilizable; `institution`/`kind`/`number`/`currency` son un único valor (R7-5) | Tabla de traducciones EAV; fallback en el cliente (analyze I2 de F3); `Accept-Language` |
| **P7-10** | **Duplicado exacto → `409 conflict`** sobre `kind` + `number` **normalizado** (`ExistsDonationMethodDuplicate`); **sin `UNIQUE`** en la BD | Default revisable de la spec ("se rechaza o se señala"), regla más fuerte que la de Assumptions y cubre el edge case (R7-9); sin índice = el default "se señala" sigue siendo una línea | `409` solo con tipo+entidad+número (pasaría la misma cuenta con el banco escrito distinto); solo señalizar en el formulario (la API forzada quedaría sin señal); `UNIQUE (kind, number)` desde ya (bloquearía el default "se señala"; queda documentado como adición posible) |
| **P7-11** | **Auditoría**: 8 códigos `donation.*` (migración `000013`), `InsertAdminAction` transaccional y `Recorder` del propio dominio; reordenar = `donation.update`; textos sin `create` (singleton) | FR-017 pide el registro de F2 y distinguir quién **retiró** un número de cuenta del público (R7-10) | Registro propio de F7; un solo código `donation.update` para todo; registrar tras el commit |
| **P7-12** | **Publicación** por método **y por texto** (`publication_state` en las dos tablas), "se publica al guardar", consultas `…Published` separadas para el público | FR-012/FR-013/Q10 literales; patrón probado en F3/F4/F6 (R7-11) | Estado por módulo; un estado común para los dos textos (Q10 exige el suyo a cada uno); paso intermedio de borrador al editar (contradice FR-013) |
| **P7-13** | **"Nada publicado" = sin métodos NI textos** (default sobre un hueco de la spec): la sección —y su navegación— existe si hay ≥1 método **o** ≥1 texto publicado; el bloque solo mira métodos | Lectura literal de FR-005 ("todo está en borrador —métodos y textos incluidos—"); **se eleva como pregunta** (R7-12) | Exigir ≥1 método publicado (igual de defendible: es un predicado, `hasOfrendasContent`); decidirlo por la presencia del bloque (iría contra FR-022) |
| **P7-14** | **Bloque de la portada**: endpoint propio `GET /api/v1/portada/ofrendas` (registrado por el dominio `donaciones`), **tope 5 firme** (`maxPortadaItems = 5` + `LIMIT`), mismo orden y **mismo DTO de ítem** que la sección, **oculto por completo sin métodos publicados** y posicionado **tras `MinisteriosBlock` y antes de `WhoWeAreSection`** | FR-022/SC-014 literales, incluida la diferencia consciente con el bloque de F4 (aquí no hay mensaje) y la posición acordada; "el bloque es un extracto de los mismos métodos, no un modelo nuevo" (R7-13) | DTO de bloque más pequeño (duplicaría esquema); bloque dentro de `GET /api/v1/portada` de F3 (mezcla dominios); tope configurable (Out of Scope); carrusel/paginación (Out of Scope) |
| **P7-15** | **Validación y límites**: tags existentes (`required`, `max`, `oneof`) + reglas de dominio; entidad/titular 160, textos **2.000**, números por su formato; `400 invalid` con `details` por campo y **sin guardado parcial** | FR-014 completo sin dependencias nuevas; los límites son el default revisable de la spec (R7-14) | Librerías de validación; límites solo en `CHECK` (mensajes peores); notas/campos extra (Q2 los prohíbe) |
| **P7-16** | **Sin imágenes**: **cero cambios** en `media`, `platform/storage` e `IsFilePublished`; el panel no ofrece carga | Q6/FR-020 literales (decisión firme del 2026-10-14) | Reutilizar el servicio de subida "por si acaso" (contradice Q6 y añadiría superficie muerta) |
| **P7-17** | **Textos del panel**: `PUT /api/v1/admin/donaciones/textos/{textKey}` (crea o reemplaza el singleton) + `GET /api/v1/admin/donaciones/textos`; **sin operación de borrado** (se retiran en `draft`) | FR-010 solo habla de "mantener" los textos; es literal el patrón `PUT /admin/portada/quienes-somos` de F3 (R7-1/R7-20) | `POST`/`PATCH` por texto (dos rutas de más para un singleton); `DELETE` de textos (no está en la spec: **se pregunta** si hace falta) |
| **P7-18** | **Frontend**: `/ofrendas` + `/panel/donaciones` (`RequirePermission(DONACIONES)`), `DONACIONES` en `lib/permissions`, tarjeta en `AccesosRapidos`, `ACTION_LABELS` +8; piezas propias en `features/publico/ofrendas/` y `features/donaciones/`; **ninguna promoción** de inventario | Una sola entrada de navegación (FR-006) y un módulo de panel con el patrón de F6; F7 no comparte piezas nuevas con otras features (todo lo necesario ya está en el inventario: F2/F3/F4/F6/F10) | Copiar `Field`/`Table`/`PublishControls` en la feature (duplicado, contra la regla del inventario); promover `HiddenNotice` (F7 **redirige**, no avisa: P7-21); una sola página de panel para público y gestión |
| **P7-19** | **Contrato**: delta en `specs/007-donaciones/contracts/openapi.yaml` → fusión aditiva en `backend/api/openapi.yaml` con `info.version` **0.7.0 → 0.8.0**; el registro de `error.code` **no cambia**; en la fusión se actualiza el enum de `AdminActionItem.action` (+8) | Contrato antes que el código (§II) sin dos copias vivas ni códigos de error ad-hoc (§8.1.10: minor al añadir operaciones) | Prefijo nuevo de rutas; `error.code` para "duplicado" (ya existe `conflict`); enum de permisos en el contrato (el catálogo vive en la BD) |
| **P7-20** | **Orden**: `ORDER BY CASE kind WHEN 'iban' THEN 0 ELSE 1 END, sort_order, id` en público **y** en panel; `sortOrder` se edita por método y manda **dentro de su grupo** | FR-004/US2 esc. 10: "el nuevo orden es exactamente el que ve el visitante dentro de ese grupo"; el panel muestra el orden real | `ORDER BY kind` a secas (coincidencia alfabética frágil); `created_at` (§8.1.7: hay criterio); listas separadas por tipo en el API (el sobre `{lang, items?}` de F3/F4/F6 se conserva) |
| **P7-21** | **Redirección de `/ofrendas` sin parpadeo**: la decisión vive en el **frontend**, en el **elemento de ruta** de `/ofrendas` (mecánica literal de P6-21 de F6): `SectionSkeleton` mientras `usePublicOfrendas` está en vuelo y **solo con éxito y sin contenido** → `<Navigate to="/" replace />`; **ante error NO se redirige**; la API sigue respondiendo **`200` con las claves omitidas** | Q9 es la decisión del cliente (patrón F6, **distinto a propósito** de F4 en `/eventos`, que no se toca); `replace` evita bucles con «Atrás», el esqueleto impide el destello y un `3xx`/`404` del API rompería el sobre JSON | Redirigir desde el backend o con `302` en el contrato; `404` (un error para algo normal); redirigir también ante **error** de red (una caída transitoria expulsaría al visitante); aviso sin redirigir (`HiddenNotice`, patrón F4: descartado por Q9) |

## Project Structure

### Documentation (esta funcionalidad)

```text
specs/007-donaciones/
├── plan.md              # EDITABLE (este archivo, /speckit.plan)
├── research.md          # EDITABLE (fase 0: R7-1…R7-16)
├── data-model.md        # EDITABLE (fase 1: tablas donation_methods/donation_texts,
│                        #   migraciones 000012–000013, consultas)
├── quickstart.md        # EDITABLE (fase 1: cómo probar F7 en local)
├── spec.md              # APROBADA (v2) — no se edita
├── contracts/
│   └── openapi.yaml     # EDITABLE (fase 1): delta de diseño, se fusiona en backend/api/openapi.yaml
└── tasks.md             # Fase 3 (/speckit.tasks — NO lo crea este comando)
```

> `ux.md` **aún no existe** para F7: a diferencia de F6, el diseño de esta sección está por
> producirse (`disenador-ux`, en paralelo con la implementación). Este plan fija rutas, DTOs,
> reglas de contenido y posiciones; la composición visual la define `ux.md` sin cambiar la API
> (mismo encaje que F4/F6, RG7-5). Los nombres de pieza de UI de este plan son **provisionales**
> hasta que `ux.md` los confirme.

### Source Code — backend

```text
backend/
├── cmd/api/
│   ├── main.go                      # EDITADO: DI del dominio donaciones + registro de
│   │                                #   /api/v1/ofrendas, /api/v1/portada/ofrendas y
│   │                                #   /api/v1/admin/donaciones
│   └── main_test.go                 # EDITADO: prueba de humo ampliada (rutas nuevas + /healthz intacto)
├── internal/
│   ├── db/
│   │   ├── queries/
│   │   │   ├── donation_methods.sql # NUEVO: ListPublishedDonationMethods (:limit),
│   │   │   │                        #   ListDonationMethods/CountDonationMethods (filtro
│   │   │   │                        #   publication_state), GetDonationMethodByID,
│   │   │   │                        #   Insert/Update/Delete, ExistsDonationMethodDuplicate
│   │   │   ├── donation_texts.sql   # NUEVO: ListPublishedDonationTexts, ListDonationTexts,
│   │   │   │                        #   GetDonationTextByKey, UpsertDonationText
│   │   │   └── audit.sql            # SIN CAMBIOS: se reutiliza InsertAdminAction (R7-10)
│   │   └── (generado por sqlc, commiteado)
│   ├── donaciones/                  # NUEVO — dominio F7 (receta de arq. §8)
│   │   ├── numero.go                # normalización y validación de IBAN (estructura + longitud
│   │   │                            #   por país) y de SINPE (+506, 8 dígitos) — funciones puras
│   │   ├── numero_test.go           # tabla de casos: CR/ES/DE/GB válidos, espacios y guiones,
│   │   │                            #   longitud errónea, país desconocido, SINPE con/sin +506,
│   │   │                            #   fija 2xxx, 7 dígitos, emojis/letras en el número
│   │   ├── model.go                 # entidad MetodoDonacion/TextoDonacion y DTOs (etiquetas
│   │   │                            #   validate, Es/En solo donde hay traducción, SortOrder)
│   │   ├── repository.go            # constructor + mapRow/params (pgtype → dominio) + withTx
│   │   │                            #   + inserción de admin_actions sobre el tx
│   │   ├── repository_donaciones.go # CRUD + ExistsDonationMethodDuplicate
│   │   ├── repository_textos.go     # List/Get/Upsert de los dos textos (singleton por clave)
│   │   ├── repository_public.go     # ListPublishedDonationMethods + ListPublishedDonationTexts
│   │   │                            #   (única vía pública: solo published)
│   │   ├── service.go               # tipos comunes, invariantes, constantes (límites,
│   │   │                            #   maxPublicItems = 200, maxPortadaItems = 5)
│   │   ├── service_public.go        # GetOfrendas (sección: items + intro + usage, ?lang con
│   │   │                            #   fallback) y GetPortadaOfrendas (tope 5 en servidor)
│   │   ├── service_admin.go         # CRUD de métodos (normalización + validación de números,
│   │   │                            #   moneda por tipo, español por campo, duplicados 409),
│   │   │                            #   guardado de textos, publicar/retirar y auditoría
│   │   ├── service_audit.go         # audit.Recorder del dominio: método+ruta → donation.*
│   │   ├── handler.go               # constructor + helpers HTTP (decodificar/validar/responder)
│   │   ├── handler_public.go        # GET /api/v1/ofrendas y GET /api/v1/portada/ofrendas
│   │   ├── handler_admin.go         # CRUD de /api/v1/admin/donaciones + textos
│   │   ├── routes.go                # RegisterPublic (2 rutas) y RegisterAdmin
│   │   │                            #   (AdminChain("donaciones"))
│   │   └── *_test.go                # service con fakes · handler con httptest · repository integration
│   ├── media/                       # SIN CAMBIOS (F7 no tiene imágenes: Q6/FR-020)
│   └── platform/
│       ├── audit/audit.go           # EDITADO: + 8 códigos donation.* (y su prueba del registro cerrado)
│       ├── middleware/chain.go      # SIN CAMBIOS: AdminChain("donaciones") reutiliza AuthzByModule
│       ├── validate/validate.go     # SIN CAMBIOS (los tags existentes bastan; las reglas de
│       │                            #   número viven en el dominio, R7-2/R7-3)
│       └── config/config.go         # SIN CAMBIOS (sin variables de entorno nuevas)
├── migrations/
│   ├── 000012_create_donation_content.up.sql / .down.sql               # NUEVO: donation_methods + donation_texts
│   └── 000013_extend_admin_actions_for_donations.up.sql / .down.sql    # NUEVO: CHECK de auditoría +8 códigos
└── api/openapi.yaml                 # EDITADO: se fusiona el delta (info.version → 0.8.0) + make api-gen
```

### Source Code — frontend y raíz

```text
frontend/src/
├── api/
│   ├── ofrendas.ts                  # NUEVO: getOfrendas(lang), getPortadaOfrendas(lang) y panel
│   │                                #   (list/create/update/delete de métodos, getTextos,
│   │                                #   putTexto)
│   └── schema.d.ts                  # REGENERADO (npm run api:gen; contrato 0.8.0)
├── app/
│   ├── router.tsx                   # EDITADO: '/ofrendas' (pública) y '/panel/donaciones'
│   │                                #   con RequirePermission(DONACIONES)
│   ├── layout.tsx                   # EDITADO: entrada del menú del panel filtrada por permiso
│   └── guards.tsx                   # SIN CAMBIOS (RequirePermission de F2 basta)
├── features/
│   ├── publico/                     # EDITADO: sitio público (F10 aplicado, no redefinido)
│   │   ├── ofrendas/                # NUEVO (nombres de pieza provisionales; ver nota de ux.md):
│   │   │   ├── OfrendasPage.tsx     #   H1 + texto introductorio (si viene) + dos grupos
│   │   │   │                        #   («Cuentas bancarias (IBAN)» y «SINPE Móvil») + texto de
│   │   │   │                        #   uso (si viene); sin publicados → REDIRIGE a la portada
│   │   │   │                        #   (P7-21: SectionSkeleton durante la carga y
│   │   │   │                        #   <Navigate to="/" replace /> solo con éxito y sin
│   │   │   │                        #   contenido; nunca aviso de sección vacía; ante error →
│   │   │   │                        #   estado de error con «Volver a intentar», sin redirigir)
│   │   │   ├── MetodoCard.tsx       #   ficha: tipo (chip), número (IBAN agrupado en bloques /
│   │   │   │                        #   SINPE normalizado), entidad, moneda SOLO si IBAN y
│   │   │   │                        #   titular SOLO si existe (SC-010: sin líneas vacías) +
│   │   │   │                        #   botón de copiar (copia `number` exacto)
│   │   │   ├── OfrendasBlock.tsx    #   bloque de la portada: hasta 5 MetodoItem con enlace a
│   │   │   │                        #   /ofrendas; sin métodos publicados → null (NO se
│   │   │   │                        #   renderiza; FR-022)
│   │   │   ├── MetodoItem.tsx       #   tarjeta compacta del bloque (tipo + número + entidad;
│   │   │   │                        #   el detalle vive en /ofrendas)
│   │   │   ├── format.ts            #   formatIbanDisplay (agrupación en bloques de 4, solo
│   │   │   │                        #   presentación) con sus pruebas unitarias
│   │   │   ├── hooks/usePublicOfrendas.ts    # TanStack Query ['ofrendas', lang]
│   │   │   ├── hooks/usePortadaOfrendas.ts   # ['portada-ofrendas', lang] (bloque)
│   │   │   └── *.test.tsx
│   │   ├── hooks/useSiteLinks.ts    # EDITADO: SiteLinkId + 'ofrendas'; entrada «Ofrendas y
│   │   │                            #   diezmos» (href /ofrendas) solo si el documento de la
│   │   │                            #   SECCIÓN trae items/intro/usage (P7-8)
│   │   ├── pages/HomePage.tsx       # EDITADO: + <OfrendasBlock /> inmediatamente después de
│   │   │                            #   <MinisteriosBlock /> (una inserción; el resto de la
│   │   │                            #   portada de F3 intacto)
│   │   └── i18n/{messages,es,en}.ts # EDITADO: claves tipadas nuevas (nav, sección, errores)
│   ├── donaciones/                  # NUEVO: gestión en el panel (en español, FR-018)
│   │   ├── pages/DonacionesPage.tsx #   pestañas o secciones: métodos (listado con StatusPill,
│   │   │                            #   sortOrder y acciones) y textos (introductorio/uso)
│   │   ├── components/              # DonationMethodForm (Dialog, kind + número + entidad +
│   │   │                            #   titular es/en + moneda SOLO si IBAN + PublishControls),
│   │   │                            #   DonationMethodsList, TextosForm (dos editores con su
│   │   │                            #   estado), CopyableNumber (previsualización normalizada)
│   │   ├── hooks/                   # useDonaciones, useCreateDonacion, useUpdateDonacion,
│   │   │                            #   useDeleteDonacion, useDonacionTextos, usePutDonacionTexto,
│   │   │                            #   queryKeys
│   │   ├── validation.ts            # Zod (espejo de las reglas del server: obligatorios, moneda
│   │   │                            #   por tipo, español por campo, límites)
│   │   ├── errors.ts                # details → errores por campo
│   │   ├── messages.ts              # textos del panel
│   │   └── *.test.tsx
│   ├── roles/permissions.ts         # EDITADO: isPermissionAvailable deja de marcar «donaciones»
│   │                                #   como «Disponible más adelante» (y sus tests)
│   ├── panel/components/AccesosRapidos.tsx  # EDITADO: tarjeta «Gestión de donaciones» filtrada
│   │                                #   por permiso (camino feliz de US2)
│   ├── auditoria/messages.ts        # EDITADO: ACTION_LABELS cubre los 8 códigos donation.*
│   └── (resto SIN CAMBIOS)
├── lib/permissions.ts               # EDITADO: + DONACIONES = 'donaciones' (+ MODULE_PERMISSIONS)
└── (index.css, tailwind.config)     # SIN CAMBIOS (tokens de F10/F3)
frontend/e2e/
├── ofrendas-publica.spec.ts         # NUEVO: visitante (US1, US4, US5) — incluye ficha sin
│                                    #   titular, SINPE sin moneda, números exactos y redirección
│                                    #   a la portada sin nada publicado (P7-21)
├── ofrendas-panel.spec.ts           # NUEVO: gestión de métodos y textos, permisos y auditoría
│                                    #   (US2, US3)
├── ofrendas-accesibilidad.spec.ts   # NUEVO: recorrido por teclado y foco (US5, SC-008), incluida
│                                    #   la redirección (el foco no queda atrapado)
├── portada-ofrendas.spec.ts         # NUEVO: bloque de la portada (US6, SC-014)
└── ofrendasSeed.ts                  # NUEVO: fixtures del módulo (patrón de agendaSeed.ts /
                                     #   ministeriosSeed.ts)

# Raíz
README.md                            # EDITADO por documentador al cerrar
.env.example                         # SIN CAMBIOS (F7 no añade variables)
docker-compose.yml / Makefile / .github/workflows/ci.yml / .githooks/  # SIN CAMBIOS (son del kit)
```

### Inventario único de componentes UI

F7 **no promueve ninguna pieza** al inventario único (regla de F2/T242): todo lo que necesita ya
está promovido por F3/F4/F6/F10 y se **reutiliza sin tocar**: `Field`, `TextArea`, `Select`,
`Button`, `Notice`, `ConfirmDialog`, `Dialog`, `EmptyState`, `Table`, `Tabs`, `StatusPill`,
`PublishControls` y el sistema de F10 (`Card`, `SectionHeading`, `PublicButton`, `Reveal`/
`RevealGroup`, `SectionSkeleton`, `PublicLayout`, `PublicNav`, `MobileMenu`, `FooterSiteMap`,
`LanguageSwitcher`, `ThemeToggle`), más `CardImage`… que F7 **ni siquiera usa** (Q6: sin
imágenes). **`HiddenNotice` no se toca**: F7 **redirige** en lugar de avisar (P7-21) y la pieza se
queda en `features/publico/eventos/`, donde F4 la sigue usando para `/eventos`. Las piezas propias
de F7 (`OfrendasPage`, `MetodoCard`, `MetodoItem`, `OfrendasBlock`, `CopyableNumber`,
`DonationMethodForm`, `DonationMethodsList`, `TextosForm`) viven en sus carpetas de feature y solo
se promueven si otra feature las necesita más adelante.

## Cadena de middleware y grupos (orden; el primero es el más externo)

```text
Global:   request-id → recover → logging → CORS(con credenciales) → handler
          /api/v1/ofrendas                 → (sin middlewares)  → handler    ← público, solo lectura
          /api/v1/portada/ofrendas         → (sin middlewares)  → handler    ← público, solo lectura (lo registra
                                                                                `donaciones.RegisterPublic`, aunque el
                                                                                segmento `portada` sea el agregado de F3)
          /api/v1/admin/donaciones         → authn → guard de cambio de contraseña → authz("donaciones") → CSRF → handler
          /api/v1/admin/donaciones/{id}    → (misma cadena)                          → handler
          /api/v1/admin/donaciones/textos  → (misma cadena)                          → handler
```

- La superficie pública de F7 es de **solo lectura**: sin `authn`, `CSRF` ni `rate-limit` (F2 solo
  lo monta sobre superficies públicas escribibles; aquí no hay ninguna).
- El grupo de panel monta la **cadena aprobada de F2** con `AuthzByModule` tal cual (solo cambia el
  código de permiso: `donaciones`, ya sembrado). La denegación responde `403 forbidden` genérico y
  **queda registrada** en `admin_actions` (`result='denied'`); el `Recorder` de la cadena es el
  del dominio `donaciones` (resuelve sus rutas → acción `donation.*`, R7-10).
- `/healthz` **no cambia**: sigue público, sin sesión y con `Cache-Control: no-store`.

## Contrato OpenAPI: dónde vive y cómo se mantiene

Igual que F1/F2/F3/F4/F6 (reglas §8.1.5 y §8.1.10):

1. **Diseño (esta fase)**: `specs/007-donaciones/contracts/openapi.yaml` — delta inmutable una vez
   aprobado el plan.
2. **Fusión**: la primera tarea de implementación lo funde en `backend/api/openapi.yaml` de forma
   **aditiva**, aplica los cambios documentados del delta (enum `action` de `AdminActionItem`
   **+8 códigos** `donation.*`) y sube `info.version` de **0.7.0 a 0.8.0** (minor: operaciones y
   esquemas nuevos; el registro de `error.code` **no** cambia; `GET /api/v1/media/{fileName}` **no
   se toca**). Luego `make api-gen`.
3. **Consumo**: el frontend genera tipos solo desde `backend/api/openapi.yaml` (`schema.d.ts`).
4. El snapshot de `specs/` no se vuelve a editar.
5. **La prosa del catálogo de permisos NO cambia**: F2 sigue teniendo 10 permisos; `donaciones`
   ya estaba sembrado por `000002` (como `ministerios` en F6; a diferencia de F4, que sí tuvo que
   ampliar el catálogo).

## Estrategia de pruebas

Cobertura exigida: **80 % en `service/`** (§III). Comandos: `go test ./...` ·
`go test -tags=integration ./...` · `npm test -- --run` · `make e2e` · `make ci`.

| Capa | Tipo | Qué verifica F7 | Dónde |
|---|---|---|---|
| `platform/audit` | Unitaria | Los 8 códigos nuevos en el registro cerrado; objetivo `content` con `targetLabel` | `internal/platform/audit/*_test.go` |
| `donaciones/numero` | Unitaria (tabla de casos) | **IBAN**: aceptados `CR` 22, `ES` 24, `DE` 22, `GB` 22 (con espacios/guiones y en minúsculas, que se normalizan), rechazados longitud errónea por país (`CR` con 21/23), país fuera del registro, letras fuera de estructura, emojis/números con guion interno; el valor guardado es el normalizado y **no** cambia. **SINPE**: aceptados `88888888`, `8888-8888`, `+506 8888 8888`, `+50688888888` → todos `+50688888888`; rechazados 7/9 dígitos, `2xxx` (fija), letras, `+507`, sufijo extra (SC-009) | `internal/donaciones/numero_test.go` |
| `service` (dominio `donaciones`) | Unitaria con **fakes** | **Público**: solo `published` **y por texto** (un texto en borrador no sale aunque haya métodos); orden `iban → sinpe` y `sortOrder, id` dentro de cada grupo; fallback `en → es` por campo (**nunca** huecos); `lang` inválido → `invalid`; claves omitidas sin contenido; **tope `maxPublicItems = 200`** y **`maxPortadaItems = 5`** fijados por prueba (el bloque: 5 exactos con 8 publicados, exactos con menos, sin `items` con 0; los 5 son los **primeros del orden de la sección**, SC-014); los DTOs públicos **no** traen `publicationState` ni campos inexistentes (`currency` ausente en SINPE, `holder` ausente sin titular). **Panel**: validación de FR-014 (obligatorios; `currency` obligatoria en IBAN y **rechazada** en SINPE; `holderEn` exige `holderEs`; números mal formados → `400` con `details.number` y **sin guardado parcial**; límites; `""` → `NULL`), normalización aplicada antes de persistir, duplicado exacto → `conflict`, publicar/retirar por método **y por texto**, "se publica al guardar" (FR-013), `PATCH` que fusiona y `PUT` de textos que reemplaza. **Auditoría**: cada mutación registra su acción con el código y `targetLabel` correctos en la misma transacción (fail-closed; datos + estado → dos filas; alta = solo `donation.create`); denegaciones y rechazos registrados best-effort | `internal/donaciones/service_*_test.go` |
| `handler` | Unitaria con `httptest` + service falso | Cada operación: decodificación, `details` por campo en 400, códigos 200/201/204/400/401/403/404/409, sobres uniformes; **el DTO público serializado nunca contiene `publicationState`** ni borradores; `GET /api/v1/ofrendas` y `/api/v1/portada/ofrendas`: `200` con `no-store`, claves omitidas con 0 publicados | `internal/donaciones/handler_*_test.go` |
| `repository` | **Integración** (PostgreSQL real, `//go:build integration`) | SQL real: `CHECK` de formato por tipo (IBAN con espacios o en minúsculas rechazado en la BD; SINPE sin `+506` rechazado), `CHECK` de moneda (IBAN sin moneda falla; SINPE con moneda falla), español por campo (`holder_en` sin `holder_es`), `btrim` en entidad/titular, `text_key` fuera de catálogo y `UNIQUE (text_key)`; `ListPublished*` **nunca** devuelve borradores (ni métodos ni textos); orden `CASE kind …, sort_order, id` y `LIMIT` (5 en el bloque); `ExistsDonationMethodDuplicate` con normalización; `UpsertDonationText` crea y actualiza sin duplicar; **mutación + `admin_actions` atómicas** (y rollback si falla el registro); migración `000012`/`000013` up→down→up; el catálogo de `permissions` sigue en 10 | `internal/donaciones/repository_*_test.go` |
| `cmd/api` | Humo de composición | Rutas nuevas con su sobre y permiso, `/healthz` intacto y `404` `not_found` en rutas inexistentes (§8.1.9) | `cmd/api/main_test.go` |
| Frontend | Unitaria (Vitest + MSW) | Pública: dos grupos con sus encabezados y en el orden del API; IBAN **agrupado en bloques** en pantalla y botón de copiar con el **valor exacto** (SC-009); SINPE en su formato normalizado; **sin titular no se pinta línea** y **sin moneda en SINPE** (SC-010); fallback en→es visible; entrada de navegación y bloque ocultos sin publicados; **sin publicados la página redirige a la portada sin destello** (esqueleto `SectionSkeleton`; nunca la grilla vacía ni aviso; `<Navigate replace>`; **ante error NO redirige**). Panel: formulario con errores por campo (números y moneda), moneda oculta en SINPE, textos con su `StatusPill`, publicar/retirar, `RequirePermission(DONACIONES)`, accesos rápidos, **`ACTION_LABELS` con etiqueta humana para los 8 códigos**, error inesperado con mensaje genérico **sin perder lo escrito** (tabla de errores de la spec); y las 2–3 pruebas actualizadas (permisos/accesos, RG7-6) | `frontend/src/features/*/*.test.tsx` |
| E2E pública | Playwright (local) | Recorrido de `quickstart.md`: visitante sin sesión ve los dos grupos en el orden exacto (`iban` → `sinpe`, `sortOrder`), los números **íntegros** (IBAN en bloques en pantalla; SINPE `+506…`), entidad, moneda solo en IBAN y titular solo cuando existe; borrador invisible (ni método ni texto); **sin nada publicado, la visita directa a `/ofrendas` redirige a la portada** (URL `/`, contenido de la portada, **sin destello** de sección vacía y «Atrás» sin bucles por el `replace`) y no hay entrada en nav/pie; inglés con fallback y vuelta a español (**números, tipo, entidad y moneda idénticos**: SC-006); 3 anchos (SC-001, SC-007, SC-009, SC-010) | `frontend/e2e/ofrendas-publica.spec.ts` |
| E2E panel | Playwright (local) | Recorrido de `quickstart.md`: login con permiso, crear+publicar un método en < 2 min (SC-004), IBAN con espacios guardado normalizado, IBAN mal formado rechazado sin guardar (SC-009), SINPE con y sin `+506`, SINPE inválido rechazado, moneda obligatoria en IBAN y rechazada en SINPE, crear un método sin titular (válido), reordenar dentro del grupo y ver el orden en público (SC-003), editar y publicar los dos textos por separado, eliminar con confirmación, denegación cruzada (cuenta sin `donaciones`: 403 en API forzada) y registro de todo en auditoría (SC-005/SC-011) | `frontend/e2e/ofrendas-panel.spec.ts` |
| E2E accesibilidad | Playwright (local) | Recorrido por **teclado** de `/ofrendas` y del bloque con foco visible en cada paso (fichas y botón de copiar activables, selector de idioma operable, sin trampas de foco), en claro y oscuro (SC-008; precedente `ministerios-accesibilidad.spec.ts`); **la redirección sin nada publicado no atrapa el foco** (el visitante termina en la portada navegable) | `frontend/e2e/ofrendas-accesibilidad.spec.ts` |
| E2E portada | Playwright (local) | Bloque en `/`: exactamente los **5** primeros del orden de la sección (con 8 publicados), exactos con menos, **sin huecos**; 0 métodos publicados → **el bloque no existe** (sin mensaje); posición inmediatamente después del bloque de ministerios y antes de «Quiénes somos»; enlace/botón hacia `/ofrendas`; 320 px y teclado (US6, SC-014) | `frontend/e2e/portada-ofrendas.spec.ts` |

**Pruebas que no faltan** (las que la spec hace críticas): **0 borradores expuestos** por ninguna
respuesta pública (SC-002: ni métodos ni textos); **0 números mal formados** aceptados ni
renderizados y **0 caracteres alterados** en los números publicados (SC-009: IBAN en bloques, SINPE
normalizado, copia = valor exacto); **cambio guardado visible en la primera carga** (SC-003);
**fallback en→es sin campos vacíos** y **números/tipo/entidad/moneda idénticos entre idiomas**
(SC-006); **permisos verificados en servidor** con cruce contra otros módulos (SC-005); **toda
operación queda registrada** —éxito, fallo y denegación— y el registro es de solo lectura (SC-011);
**sin nada publicado la URL directa `/ofrendas` redirige a la portada sin destello** (SC-010,
P7-21); **bloque con exactamente 5 y oculto sin métodos publicados** (SC-014); **texto con tildes,
`ñ`, `¿¡` y emojis intacto** y **números sin emojis ni letras fuera de estructura** (edge cases);
**ficha sin titular y SINPE sin moneda** presentadas sin huecos (US1 esc. 5, SC-013); **los dos
textos con su propio estado** (texto en borrador con métodos publicados no deja hueco: US3 esc. 7).

## Cobertura de requisitos (US1–US6 y FR-001…FR-025)

| Historia | Dónde se cubre | Verificación |
|---|---|---|
| **US1** — Ver cómo apoyar a la iglesia (P1) | `GET /api/v1/ofrendas` (P7-7/P7-9/P7-13) + `OfrendasPage`/`MetodoCard` en `/ofrendas` (P7-18) | quickstart §1 · e2e pública (SC-001, SC-009, SC-012) |
| **US2** — Gestionar la información desde el panel (P1) | `/api/v1/admin/donaciones…` (P7-6/P7-15/P7-10/P7-11) + `features/donaciones/` (P7-18) + accesos de `InicioPage` | quickstart §2–§7 · e2e panel (SC-004, SC-005, SC-011) |
| **US3** — Controlar la publicación (P2) | `publication_state` por método **y por texto** + consultas `…Published` (P7-12) + `PublishControls` + ocultación de sección/navegación (P7-13/P7-21) | quickstart §8 · e2e panel/pública (SC-002, SC-003, SC-010) |
| **US4** — Ver la información en inglés (P2) | `?lang=` con fallback en el servidor (P7-9) + i18n tipado; `institution`/`kind`/`number`/`currency` idénticos (P7-5/P7-4) | quickstart §9 · e2e pública (SC-006) |
| **US5** — Cualquier dispositivo y accesible (P2) | Sistema de F10 aplicado (P7-18) + reglas `react-frontend` + números legibles/copiables (P7-3/P7-4) | quickstart §11 (checklist) · e2e en 3 anchos (SC-007) · e2e de teclado (SC-008) |
| **US6** — Ver las ofrendas y diezmos desde la portada (P2) | `GET /api/v1/portada/ofrendas` con tope 5 (P7-14) + `OfrendasBlock` tras `MinisteriosBlock` (P7-14) + visibilidad coherente (P7-13) | quickstart §14 · e2e `portada-ofrendas.spec.ts` (SC-014) |

| FR | Dónde se resuelve en este plan | Verificación |
|---|---|---|
| FR-001 | P7-1 (dos tablas: métodos + los dos textos; sin montos, sin donantes, sin imágenes) | quickstart §1 · pruebas de modelo (SC-013: nada de F4/F5/F6 en la sección) |
| FR-002 | P7-7 (`GET /api/v1/ofrendas`, solo publicado) + P7-13 (los dos textos en su sitio) | quickstart §1 y §7 · pruebas de service/handler (SC-001) |
| FR-003 | P7-2/P7-3/P7-4 (tipo, número exacto, entidad, titular solo si existe, moneda solo IBAN; IBAN en bloques, SINPE normalizado) | quickstart §3–§5 · tabla de casos + e2e (SC-009) |
| FR-004 | P7-20 (`CASE kind …, sort_order, id`; grupos iban→sinpe) | quickstart §6 · pruebas de orden (SC-014) |
| FR-005 | P7-13 (alcance de "nada publicado") + P7-8 (entrada oculta) + P7-21 (`/ofrendas` **redirige a la portada**) | quickstart §10 · e2e (SC-010) |
| FR-006 | P7-18 (ruta `/ofrendas` y entrada única «Ofrendas y diezmos» en barra, hamburguesa y pie) | quickstart §1 · e2e pública (SC-010) |
| FR-007 | P7-5/P7-9 (`holder` y textos traducibles; `institution`/`kind`/`number`/`currency` no) | quickstart §4 y §9 · pruebas de validación (SC-006) |
| FR-008 | P7-9 (fallback `en → es` en el servidor; los datos no traducibles no cambian) | quickstart §9 · tabla de casos del service (SC-006) |
| FR-009 | `LanguageSwitcher` de F10 (sin cambios; patrón F3) | quickstart §9 · e2e pública (cambio de idioma persistente) |
| FR-010 | P7-7/P7-17 (`/api/v1/admin/donaciones…`, CRUD de métodos + `PUT` de textos) + `features/donaciones/` (P7-18) | quickstart §2–§7 · e2e panel (SC-004) |
| FR-011 | P7-6 (`AdminChain("donaciones")`, permiso ya sembrado en `000002`; sin permisos nuevos) | quickstart §2 y §12 · pruebas de handler (403 por API forzada, cruce de permisos) · e2e (SC-005) |
| FR-012 | P7-12 (`publication_state` por método y por texto) + consultas `…Published` | quickstart §8 · pruebas "0 borradores expuestos" (SC-002) |
| FR-013 | P7-12 (guardado de elemento publicado visible en la primera carga; `no-store`) | quickstart §8 · e2e panel (SC-003) |
| FR-014 | P7-15 (obligatorios, moneda por tipo, español por campo, números válidos, límites) | quickstart §3–§5 (tablas de errores) · pruebas de service/handler |
| FR-015 | P7-2/P7-3 (estructura + longitud por país, guardado normalizado, presentación en bloques sin alterar el valor) | quickstart §4 · tabla de casos + e2e (SC-009) |
| FR-016 | P7-2/P7-4 (celular CR 8 dígitos, con/sin `+506`, prefijo normalizado) | quickstart §5 · tabla de casos + e2e (SC-009) |
| FR-017 | P7-11 (8 códigos `donation.*`, transaccional, denegaciones registradas, solo lectura) | quickstart §13 · pruebas de service/repository (SC-011) |
| FR-018 | Panel en español (convención de F2/F3/F4/F6; sin `i18n/` en la gestión) | revisión de `revisor-codigo` (0 cadenas de panel pasadas por `t()`) |
| FR-019 | P7-1/P7-5 (`donation_methods`: tipo, número, entidad obligatorios; titular opcional; moneda solo IBAN; sin notas) | quickstart §3–§5 · pruebas de persistencia |
| FR-020 | P7-16 (sin imágenes: sin carga, sin columnas de archivo, `media` intacto) | quickstart (sin § de imágenes, a propósito) · revisión de `revisor-codigo` (0 rutas de subida nuevas) |
| FR-021 | P7-1/P7-17 (`donation_texts` con `intro`/`usage`, cada uno con su estado y traducibles) | quickstart §7 · e2e panel (US2) y pública (US3 esc. 7) |
| FR-022 | P7-14 (endpoint propio, tope 5 en servidor, mismo DTO de ítem, botón a `/ofrendas`, posición tras el bloque de ministerios, oculto sin métodos) | quickstart §14 · pruebas del service/repository (SC-014) · e2e de portada |
| FR-023 | P7-18/P7-14 (F10 aplicado a sección y bloque; el único contenido de F7 en la portada es el bloque) | revisión de `revisor-codigo` sobre `HomePage` (una inserción) y `PublicLayout` sin cambios |
| FR-024 | P7-18 (componentes de F10, títulos claros, `toc-targets` ≥ 44 px) + números legibles en 320 px (P7-3/P7-4) | quickstart §11 · e2e en 320/768/1280 (SC-007) |
| FR-025 | Accesibilidad: teclado, semántica, foco, contraste claro/oscuro, zoom 200 % (P7-18) | quickstart §11 (checklist WCAG 2.1 AA) · e2e de teclado (SC-008) |

## Métricas del plan (coherentes con los SC de la spec)

| SC | Qué se mide | Cómo se mide en F7 | Umbral |
|---|---|---|---|
| SC-001 | Todo visitante ve la información publicada sin cuenta (texto introductorio, grupos IBAN y SINPE, texto de uso) y encuentra cómo donar en < 30 s | e2e pública: la sección muestra los dos grupos con sus datos en la primera carga (**el e2e no cronometra**; el umbral < 30 s se cronometra en el protocolo de usabilidad de SC-012, como en F3/F4/F6) | 100 % · < 30 s (cronometrado en SC-012) |
| SC-002 | 0 borradores visibles al público (ni métodos ni textos) | Pruebas "0 borradores expuestos" (service + handler + repository) y e2e con borradores en BD | 0 |
| SC-003 | Cambio guardado visible desde la primera carga posterior | e2e panel: editar/reordenar → recargar → ver el dato (`no-store` de respaldo) | 100 % |
| SC-004 | Crear y publicar un método en < 2 min desde el panel | e2e/observación (quickstart §3) | < 2 min |
| SC-005 | Toda operación verifica el permiso `donaciones` en el servidor (incluidos intentos forzados y el cruce con otros permisos) | Pruebas de handler por operación (403) + e2e con cuenta sin `donaciones` y con otros permisos de módulo | 0 operaciones sin permiso / 100 % denegadas |
| SC-006 | Con `lang=en`: 100 % de la interfaz traducida y 0 campos vacíos (fallback a es); **números, tipo, entidad y moneda idénticos en ambos idiomas** | Pruebas de i18n (claves tipadas) + tabla de casos de fallback (el DTO no tiene `institutionEn`: idéntico por construcción) + e2e en inglés que compara los valores | 100 % / 0 alteraciones |
| SC-007 | Usable en teléfono, tableta y computadora sin desplazamiento horizontal (320 px) — sección **y** bloque | e2e en 3 anchos (320/768/1280) + e2e de portada en 320 px + checklist (quickstart §11/§14) | 0 pérdida / 0 scroll |
| SC-008 | Accesibilidad: 0 hallazgos críticos (WCAG 2.1 AA, claro y oscuro) — sección **y** bloque | Checklist de quickstart §11 revisada por QA + e2e de teclado | 0 críticos |
| SC-009 | 100 % de los números publicados íntegros, exactos y copiables (IBAN en bloques, SINPE normalizado) y **0 números mal formados** aceptados ni renderizados | Tabla de casos de `numero.go` (aceptados/rechazados por tipo) + e2e del panel (rechazo) y de la pública (valor en pantalla agrupado = `number` + separadores; copia = `number`) | 100 % / 0 |
| SC-010 | 0 secciones vacías ni campos en blanco; navegación sin enlaces rotos y el bloque oculto sin métodos; `/ofrendas` **redirige a la portada** sin destello ni avisos | Bloques/entrada omitidos sin contenido (pruebas de contrato y de `buildSiteLinks`) + e2e (la URL termina en `/`, sin grilla vacía ni aviso en ningún instante, «Atrás» sin bucles) | 0 |
| SC-011 | 100 % de operaciones registradas (quién, qué, sobre qué, cuándo); 0 registros editables/borrables | quickstart §13 + pruebas transaccionales + contrato sin escritura sobre el registro | 100 % / 0 |
| SC-012 | ≥ 90 % de personas de distintas edades identifican cómo donar y en qué se usan las donaciones sin ayuda | Prueba manual con personas (protocolo en `quickstart.md` §11, estilo F3/F4/F6): ≥ 6 personas (2 por franja 18–35/36–59/60+), guion de 3 tareas, informe `pruebas-usabilidad-SC-012.md` | ≥ 90 % |
| SC-013 | 0 eventos, actividades, grupos de conexión ni ministerios en la sección ni en el bloque | Revisión de contenido e2e (solo métodos y textos) + pruebas de modelo (sin columnas ajenas) | 0 |
| SC-014 | El bloque muestra exactamente los primeros **5** métodos publicados (o todos si hay menos) según el orden de la sección, **debajo del bloque de ministerios**, con **0 borradores**, **0 huecos**, enlace hacia `/ofrendas` con ítems y **sin bloque** sin métodos publicados | Pruebas del service/repository (tope 5, orden, `items` omitida) + e2e `portada-ofrendas.spec.ts` (quickstart §14) | 100 % / 0 |

## Riesgos

> Numeración **RG7-1…RG7-12**, propia de este plan (para no colisionar con RG6-1…RG6-13 de F6,
> RG4-1… de F4 ni R7-1… de `research.md`).

| Riesgo | Impacto | Mitigación |
|---|---|---|
| **RG7-1 — La spec se contradice sobre si la entidad financiera es traducible** (FR-019/Key Entities dicen que sí; FR-007/US4 esc. 5/SC-006 dicen que no) | El modelo cambia de una a dos columnas y SC-006 ("0 alteraciones por cambio de idioma") pasa a ser incumplible si alguien rellena el inglés | Default documentado (**no traducible**, P7-5/R7-5): única lectura que garantiza SC-006 por construcción; **pregunta #1 al humano** con el impacto exacto del cambio (migración + contrato + tests); mientras tanto, `institution` es un único valor y el e2e de idioma lo comprueba |
| **RG7-2 — La tabla de longitudes IBAN queda desactualizada** (países nuevos o correcciones del registro) | Un IBAN legítimo se rechaza (falso negativo) o uno malo pasa | La tabla es una constante documentada con su fuente (registro IBAN/ISO 13616) y `CR` explícito; tabla de pruebas con los países habituales; el error es por campo y con mensaje claro; **actualizar la constante es un cambio menor** sin migración; alternativa "solo rango 15–34" documentada y descartada |
| **RG7-3 — El SINPE exige que el primer dígito sea 6/7/8** (default aplicado para distinguir celular de línea fija) | Si el equipo tiene un número que empieza por otro dígito, se rechaza | Default revisable documentado (**pregunta al humano**): quitar la condición es una línea + sus casos de prueba; el mensaje de error explica la regla; sin impacto en el modelo |
| **RG7-4 — Incoherencia entre sección, bloque y navegación** (predicados distintos: la sección vive con métodos **o** textos; el bloque solo con métodos) | SC-010/SC-014 fallan: entrada hacia una sección vacía, bloque pintado con menos de 5, o página viva sin entrada | Un **único documento** (`GET /api/v1/ofrendas`) decide presencia de la navegación **y** la redirección (P7-8/P7-21) y el bloque mira el mismo predicado "publicado" sobre métodos con la misma consulta (`ListPublishedDonationMethods`, solo con `LIMIT` distinto); pruebas con los mismos fixtures en las tres superficies y e2e que comprueba "entrada visible ⇔ sección viva ⇔ (hay métodos o textos)" y "bloque pintado ⇔ hay métodos" |
| **RG7-5 — `ux.md` aún no existe** (composición de ficha, grupos, textos y bloque) | Tareas de frontend que compongan pantallas que no encajen con los DTOs | Este plan fija rutas, DTOs, reglas de contenido y posiciones: el diseño puede variar la composición **sin cambiar la API** (mismo encaje que F4/RG4-5 y F6/RG6-4); los nombres de pieza son provisionales hasta que `ux.md` los confirme; el bloque reutiliza el DTO de ítem de la sección, así que la tarjeta compacta puede recortar campos sin tocar el contrato |
| **RG7-6 — Pruebas de frontend que fijan la lista de permisos y accesos** (F2/F4/F6, funcionalidades cerradas) | `make ci` falla al añadir `DONACIONES`, o se lee como "debilitar" pruebas ajenas | Cambio **acotado y documentado** (mismo gesto que F6/RG6-6): `features/roles/permissions.test.ts` (la aserción `isPermissionAvailable('donaciones') === false` pasa a `true`), `features/roles/roles.test.tsx` (el sello «Disponible más adelante» deja de aplicarse) y `features/panel/panel.test.tsx` (tarjeta nueva en accesos rápidos); **sin** cambios de aserciones de comportamiento en backend (el catálogo de permisos no cambia) |
| **RG7-7 — Exactitud de los números** (SC-009: "0 caracteres alterados" entre el dato y lo que ve y copia el visitante) | Un aporte perdido por un dígito y un SC crítico en rojo | `number` es siempre el valor exacto guardado (FR-015 prohíbe transformarlo); la agrupación del IBAN vive en un helper puro con pruebas unitarias y se comprueba en e2e que **quitar los separadores devuelve `number`**; el botón de copiar copia `number` (aserción sobre el portapapeles en el e2e); la validación rechaza cualquier forma que no sea número |
| **RG7-8 — Duplicados con variantes de escritura** (`CR05 0152-…` vs `CR050152…`) | Falsos negativos (duplicados que pasan) o falsos positivos | La comparación es sobre el número **ya normalizado** (la normalización es determinista y estructural: R7-2/R7-3), no sobre el texto del formulario; prueba del repository con las variantes; el `409` del duplicado es un default revisable y su regla está en el contrato |
| **RG7-9 — Accesibilidad y usabilidad (SC-008/SC-012) no cubiertas por automatización** | Hallazgos críticos en la revisión manual | Checklist explícita en `quickstart.md` §11 ejecutada por `qa-tester`; componentes semánticos de F10 (ya auditados); números con tamaño de fuente y contraste suficientes (decisión de `ux.md`, coherente con US5); protocolo de usabilidad con personas (SC-012) como en F3/F4/F6 |
| **RG7-10 — El caso "solo textos publicados"** (0 métodos) no está en ningún escenario de aceptación | La sección puede quedar viva sin cuentas que mostrar (o muerta con textos publicados, según la lectura) | Default documentado (**la sección vive**: R7-12/P7-13) con la alternativa lista (exigir ≥1 método es un predicado); **pregunta #2 al humano**; el e2e cubre el caso explícitamente para que la decisión quede fijada por prueba |
| **RG7-11 — Muchos métodos publicados** (sin límite artificial, default revisable) y textos de 2.000 caracteres | Respuestas grandes en móvil; textos que rompen la composición | Constante de seguridad `maxPublicItems = 200` en el service (documentada en el contrato, como en F6) y orden estable; los textos tienen límite por `CHECK` y la presentación debe acomodar varios párrafos (US5/edge case: el límite es default revisable) |
| **RG7-12 — Destello, callejón o bucle en la redirección de `/ofrendas`** | SC-010/SC-008: un visitante con un enlace guardado ve un destello de sección vacía, queda sin explicación o «Atrás» lo atrapa en bucles | Mecánica fijada en **P7-21** (idéntica a la probada en F6/P6-21): `SectionSkeleton` mientras la consulta está en vuelo (**nunca** el estado vacío), redirección **solo** con éxito y sin contenido, con `<Navigate to="/" replace />`; **ante error NO se redirige** (estado de error con «Volver a intentar»); e2e que comprueba que en ningún instante se pinta la sección vacía ni un aviso, que la URL termina en `/`, que la portada carga completa y que «Atrás» no devuelve a `/ofrendas` |

## Complexity Tracking

> Sin violaciones de la constitución. Desviaciones de **convenciones internas**, declaradas:

1. **La presencia de la navegación bebe del documento de la sección** (P7-8), no del bloque como
   en F6/P6-9: es el patrón de F4/D-23 y está justificado porque los textos solo viven en la
   sección (R7-7/R7-12). Coste declarado: una petición pública más por página (documento acotado,
   compartido por `queryClient` con la página `/ofrendas`).
2. **El segmento público y el de panel se nombran distinto** (`ofrendas` vs `donaciones`, P7-7):
   no es una inconsistencia sino la distinción literal de Q8 (sección pública vs. módulo/permiso
   del panel). Queda anotada en el contrato y en el quickstart; renombrar un segmento es un
   refactor acotado si `revisor-codigo` prefiere unificar.
3. **Mapeo `audit.Action` → `InsertAdminActionParams` repetido** (R7-10): ~10 líneas en el
   repository de `donaciones` y otras tantas en `portada`/`agenda`/`ministerios`/`usuarios`,
   todas sobre la **misma** consulta nombrada de `internal/db` (misma decisión que F3/F4/F6;
   subirlo a `platform/` arrastraría `pgtype` fuera del diagrama de dependencias).
4. **`/api/v1/portada/ofrendas` lo registra el dominio `donaciones`** (P7-14): el segmento
   `portada` pertenece al agregado de F3 pero el recurso es de este módulo (mismo gesto que la
   Complexity 8 de F4 con `proximamente` y la 4 de F6 con `ministerios`).
5. **Reordenar se registra como `donation.update`** (R7-10): la spec enumera "reordenar" como
   operación de auditoría y lo es (queda registrada), pero sin código propio, igual que en
   F3/F4/F6. Si el humano prefiere `donation.reorder`, es un código más en la misma lista cerrada.
6. **Los dos textos comparten tabla con `text_key`** (P7-1/R7-1): F3 usó una tabla singleton por
   contenido (`home_about`…); aquí los dos textos tienen la misma forma y el catálogo cerrado
   `text_key` + `UNIQUE` da las mismas garantías con la mitad de superficie. Si `revisor-codigo`
   prefiere dos tablas, es un refactor de migración + consultas sin cambio de contrato.
7. **F7 se aparta a propósito del patrón de F4 en la sección sin contenido** (P7-21): F4 muestra
   un aviso (`HiddenNotice`) en `/eventos` **sin redirigir** y F7 **redirige** `/ofrendas` a la
   portada (Q9). No es una inconsistencia accidental: F4 **no** se toca y la divergencia queda
   anotada en R7-12, P7-21, el contrato y el quickstart (igual que F6 la anotó en la suya).
8. **Nombres de pieza de UI provisionales** (RG7-5): `OfrendasPage`, `MetodoCard`, `MetodoItem`,
   `OfrendasBlock`, `CopyableNumber`, `DonationMethodForm`, `DonationMethodsList`, `TextosForm`
   son propuestos aquí porque `ux.md` aún no existe para F7; `tasks.md` los tomará como referencia
   y cualquier ajuste de composición posterior se aplica en las tareas `[frontend]` **sin** tocar
   la API.

## Propuesta de tareas (insumo para `/speckit.tasks`; **no** es `tasks.md`)

Orden típico migración → contrato → repositorio → servicio → handler → frontend → e2e. Las capas se
reparten `[db]`/`[backend]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` → `devops`;
las marcadas `[P]` son independientes entre sí y pueden ejecutarse en paralelo.

1. `[db]` Migración `000012` (`donation_methods` + `donation_texts`, con sus `CHECK` de formato
   por tipo, moneda por tipo, español por campo, `text_key` y `btrim`) con su `down` y prueba
   up/down/up. `[P]`
2. `[db]` Migración `000013` (`admin_actions_action_check` + 8 códigos `donation.*`) con su `down`
   y prueba up/down/up. `[P]`
3. `[backend]` Fusión del delta en `backend/api/openapi.yaml` (`info.version` → **0.8.0**; enum
   `AdminActionItem.action` +8) + `make api-gen`. `[P]`
4. `[backend]` `platform/audit`: 8 códigos + pruebas del registro cerrado. `[P]`
5. `[backend]` `numero.go` + `numero_test.go` (normalización y validación de IBAN por estructura y
   longitud por país, y de SINPE con/sin `+506`) — **sin dependencias**. `[P]`
6. `[backend]` Consultas `donation_methods.sql` + `donation_texts.sql` (`make sqlc`) y repository
   de `donaciones` con pruebas de integración (los `CHECK` de la BD, `ListPublished*`, orden
   `CASE kind`, `ExistsDonationMethodDuplicate`, `UpsertDonationText` y la atomicidad con
   auditoría). *Depende de 1, 2, 4 y 5.*
7. `[backend]` Service público de `donaciones` (sección con `items`+`intro`+`usage` y bloque con
   `maxPublicItems`/`maxPortadaItems`, idioma con fallback, claves omitidas sin contenido) +
   pruebas. *Depende de 6.*
8. `[backend]` Service de panel (CRUD de métodos con normalización/validación de números y moneda
   por tipo, duplicados `409`, publicar/retirar por método y por texto, guardado de textos,
   auditoría) + pruebas. *Depende de 6.*
9. `[backend]` Handlers y `routes.go` (`RegisterPublic` con las 2 rutas; `RegisterAdmin` con
   `AdminChain("donaciones")`) + pruebas + humo en `main_test.go`. *Depende de 7 y 8.*
10. `[frontend]` Cliente `api/ofrendas.ts` + `schema.d.ts` regenerado + hooks. *Depende de 3.*
    `[P]` respecto a 5–9 (avanza contra el contrato).
11. `[frontend]` Página pública `/ofrendas` (`OfrendasPage` con los dos grupos en el orden del
    API, `MetodoCard` con `formatIbanDisplay` y botón de copiar, titular solo si existe, moneda
    solo en IBAN, textos solo si vienen, **redirección a la portada sin nada publicado** —P7-21—,
    i18n). *Depende de 10.*
12. `[frontend]` `useSiteLinks` con la entrada «Ofrendas y diezmos» condicional (fuente: documento
    de la sección) + `OfrendasBlock` con sus `MetodoItem` en `HomePage` (tras `MinisteriosBlock`)
    + i18n. *Depende de 10.*
13. `[frontend]` Panel `features/donaciones/` (`DonacionesPage` con `DonationMethodsList` y
    `TextosForm`, `DonationMethodForm` con moneda solo si IBAN y validación Zod espejo),
    ruta `/panel/donaciones` con `RequirePermission(DONACIONES)`, `lib/permissions.ts`,
    `features/roles/permissions.ts`, `AccesosRapidos` y `ACTION_LABELS` (+8) + actualización
    acotada de las 2–3 pruebas que fijan permisos/accesos (RG7-6). *Depende de 10.*
14. `[…]` e2e `ofrendas-publica.spec.ts` (incluye **ficha sin titular**, **SINPE sin moneda**,
    **números exactos** y **redirección a la portada sin nada publicado**), `ofrendas-panel.spec.ts`
    (incluye cruce de permisos, rechazo de números y textos), `ofrendas-accesibilidad.spec.ts` y
    `portada-ofrendas.spec.ts` + `ofrendasSeed.ts` + checklist de quickstart. *Depende de 11–13.*
15. `[…]` Validación (qa/revisor/seguridad), documentación (documentador) y PR. *Depende de 14.*

## Preguntas para el humano (defaults aplicados; nada se implementa como regla cerrada sin esto)

Los cuatro primeros puntos son huecos o contradicciones de la **spec**; se aplican los defaults
indicados —los más literales posibles— y quedan fijados por prueba, de modo que cambiarlos es un
cambio acotado. **No son reglas de negocio inventadas**: en cada caso se cita el texto que manda.

1. **¿La entidad financiera (nombre del banco) es traducible al inglés?** *(contradicción interna
   de la spec: FR-019/Key Entities dicen que sí; FR-007/US4 esc. 5/SC-006 dicen que no)* —
   **Default aplicado: NO traducible** (una sola columna `institution`), porque es la única
   lectura que permite cumplir SC-006 ("la entidad financiera es idéntica en ambos idiomas, 0
   alteraciones"). Si decides que sí: `institution_es`/`institution_en` + `CHECK` de español por
   campo + fallback, y habría que ajustar la redacción de SC-006/US4 esc. 5 (hoy imposibles de
   cumplir con traducción).
2. **¿La sección «Ofrendas y diezmos» existe con solo textos publicados y 0 métodos?** *(la spec
   no tiene ningún escenario para este caso)* — **Default aplicado: sí** (lectura literal de
   FR-005: "no tiene nada publicado (o todo está en borrador —métodos y textos incluidos—)":
   con un texto publicado "hay algo publicado"), con la navegación viva y **el bloque de portada
   oculto** (FR-022 solo mira métodos). Alternativa: exigir ≥1 método publicado para que la sección
   y su entrada existan (es un predicado, `hasOfrendasContent`).
3. **¿El IBAN debe validar además los dígitos de control (mod-97)?** *(Q3/FR-015 piden solo
   "estructura y longitud")* — **Default aplicado: no** (solo estructura + longitud exacta por
   país), para no rechazar lo que la spec acepta. Si quieres el control mod-97 (recomendable como
   endurecimiento, no como regla de negocio): una función más en `numero.go` y sus casos de prueba.
4. **¿El número de SINPE debe empezar por 6, 7 u 8?** *(Q4 dice "celular costarricense de 8
   dígitos"; distinguir celular de línea fija exige mirar el primer dígito)* — **Default aplicado:
   sí** (se rechazan los `2xxx` de línea fija). Si prefieres aceptar cualquier 8 dígitos, se quita
   esa condición.

Y dos confirmaciones menores (defaults ya aplicados por precedentes de F3/F4/F6, revisables):

5. **Duplicados**: se **rechazan** (`409`) cuando coinciden **tipo y número normalizado** (más
   fuerte que el "mismo tipo, misma entidad y mismo número" de Assumptions y cubre "una cuenta
   repetida" del edge case) y **sin `UNIQUE`** en la BD (patrón F6). Si prefieres la regla literal
   de Assumptions o un índice único, es un cambio acotado.
6. **Textos sin borrado**: los textos introductorio y de uso **no se pueden eliminar**, solo
   retirar a borrador (FR-010 solo habla de "mantenerlos"; F3/F4/F6 tampoco tienen borrado de
   singletons). Si el equipo necesita borrarlos, añadimos `DELETE /api/v1/admin/donaciones/textos/{textKey}`
   en el contrato y en las tareas.
