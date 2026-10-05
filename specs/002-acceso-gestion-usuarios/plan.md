# Implementation Plan: Acceso y gestión de usuarios (F2)

**Branch**: `002-acceso-gestion-usuarios` | **Date**: 2026-10-04 (actualizado el 2026-10-04 con las
decisiones del humano: sesión en Redis, tiempos de sesión y datos de la cuenta; y ese mismo día con
el **cambio de alcance: auditoría** — US8, FR-021…FR-026, spec re-aprobada) | **Spec**: [spec.md](./spec.md) (APROBADA por el humano el 2026-10-04)

**Input**: `specs/002-acceso-gestion-usuarios/spec.md` (fuente de verdad) · `.specify/memory/constitution.md` · `docs/tecnico/decisiones.md` (D-A1…D-A9) y `docs/tecnico/arquitectura.md` (referencia obligada, §8.1 incluida) · plataforma ya construida en F1 (`specs/001-estructura-base/plan.md`, `research.md`, `data-model.md`, `contracts/openapi.yaml`) · skills `go-backend`, `postgres-db`, `react-frontend`.

> ✅ **D-A7 CONFIRMADA por el humano el 2026-10-04** — con un cambio explícito: **la sesión vive en
> Redis** (no en PostgreSQL; el humano la adopta/proba como objetivo de aprendizaje). Se mantiene
> lo confirmado: cookie `httpOnly`/`SameSite=Lax`/`Secure` configurable, hash de contraseñas
> **bcrypt (cost 12)**, permisos por módulo en tablas propias y CSRF *double-submit* firmado con
> `SESSION_SECRET`. Confirmó además los **tiempos de sesión: vida absoluta de 1 hora + inactividad
> de 30 minutos** (P9/R15) y los **datos de la cuenta: nombre, apellidos, correo y teléfono**
> (P12/R20). Todo ello está aterrizado en **P1/P5/P6/P9/P12** y en `research.md` R1–R5, R15, R19–R20.
>
> 🆕 **CAMBIO DE ALCANCE: AUDITORÍA (aprobado por el humano el 2026-10-04; spec re-aprobada)** —
> F2 incorpora la auditoría de su ámbito: **último acceso exitoso por cuenta** (FR-021),
> **historial de intentos de acceso** (FR-022), **historial de acciones administrativas** (FR-023)
> y una **sección de solo lectura** con filtros por cuenta y rango de fechas y paginación (FR-024…
> FR-026, US8/P3). Este plan lo incorpora en **P20–P23**, las migraciones **000004**, las tablas
> `login_events`/`admin_actions` (`data-model.md`), los endpoints `GET /api/v1/admin/auditoria/…`
> (`contracts/openapi.yaml`) y la sección de prueba §10 (`quickstart.md`). **Redis queda confirmado
> sin persistencia** (sin `appendonly`, sin volumen): un reinicio solo obliga a re-loguearse y
> reinicia contadores; **la auditoría no se ve afectada** porque vive en PostgreSQL (P23, R21).
>
> **Queda la puerta de aprobación del plan**: por ser un cambio de alcance, el plan actualizado
> **vuelve a necesitar aprobación humana** antes de `/speckit.tasks` y de cualquier código de F2
> (regla 1 de AGENTS.md).

## Summary

F2 entrega la puerta del panel de administración y su gestión de accesos: login/logout con sesión
en servidor, inicialización única del administrador (con regla anti-bloqueo), gestión de cuentas
(crear, editar, activar/desactivar, restablecer contraseña — **nunca eliminar**) y gestión de roles
con permisos por módulo (un rol por cuenta, catálogo de permisos fijo = módulos del producto). Con
el **cambio de alcance: auditoría** (aprobado el 2026-10-04) entrega además el registro duradero de
su ámbito: **historial de intentos de acceso** (FR-022) y **de acciones administrativas** (FR-023)
en PostgreSQL, el **último acceso exitoso** en la ficha de cada cuenta (FR-021) y una **sección de
solo lectura** con filtros por cuenta y rango de fechas y paginación (FR-024…FR-026, P20–P22).
Técnicamente es la funcionalidad que **activa los diferidos de F1**: `platform/validate`,
`platform/paginate`, middleware `authn`/`authz`/`CSRF`/`rate-limit`, primeras tablas de negocio
(`users`, `roles`, `permissions`, `role_permissions`, `login_events`, `admin_actions`), primeras
consultas sqlc, el **primer uso de Redis** (sesión y contadores de acceso — estado efímero **sin
persistencia**, confirmado; la auditoría duradera vive en PostgreSQL), primeras dependencias
nuevas justificadas (`golang.org/x/crypto`, `github.com/google/uuid`, `github.com/redis/go-redis/v9`
en backend, `github.com/testcontainers/testcontainers-go` solo en pruebas de integración;
`react-hook-form`, `zod`, `@hookform/resolvers` en frontend) y el primer uso de la
superficie pública escribible (con lo que se implementa el `rate-limit` diferido con su punto de
decisión). Todo se construye sobre lo ya hecho: `internal/platform/` (config, logger, apperr,
httpserver con `Registrar`/`Group`/`WriteJSON`/`WriteError`, middleware request-id/recover/logging/CORS,
database con `WithTx`), sqlc sobre `backend/migrations/`, DI manual en `main.go`, sobres uniformes
y `GET /healthz` intacto.

## Technical Context

**Language/Version**: Go 1.27 (D-A5, sin cambios) · TypeScript 5.x `strict` sobre Node 22 · React 19

**Primary Dependencies (nuevas, justificadas en `research.md` R16)**: backend runtime →
`golang.org/x/crypto` (bcrypt; §IV exige bcrypt o argon2 — no hay forma de cumplirlo sin esta
dependencia), `github.com/google/uuid` (tipos UUID del dominio; §8.1.1 ya lo exige y justifica) y
`github.com/redis/go-redis/v9` (cliente de Redis para la sesión y los contadores de acceso;
**justificado bajo D-A8 por decisión explícita del humano el 2026-10-04 de adoptar Redis**).
Solo pruebas de integración → `github.com/testcontainers/testcontainers-go` (levanta Redis y, si
hace falta, PostgreSQL desde el propio test: el CI del kit no se puede editar — R19).
frontend → `react-hook-form` + `zod` + `@hookform/resolvers` (convención de la skill para
formularios). El resto se mantiene: `pgx/v5` único runtime previo; `sqlc`, `golang-migrate`,
`openapi-typescript` como herramientas de desarrollo.

**Storage**: **PostgreSQL 16** (servicio `db`) para las tablas de negocio y para la **auditoría
duradera** (F2 crea las migraciones `000002`, `000003` y `000004`; ver `data-model.md`) + **Redis 7**
(servicio `redis`, nuevo, **sin persistencia** — confirmado) para la sesión y los contadores de
intentos de acceso (claves `sess:*`, `user_sessions:*`, `login:fail:*`, `login:block:*` con TTL;
D-A7 confirmada). Capa de datos sqlc (D-A3) solo para PostgreSQL: consultas en
`backend/internal/db/queries/`, código generado commiteado; Redis se accede por la interfaz
`session.Store` (`internal/platform/session`).

**Testing**: estrategia por capa de F1 ampliada con las pruebas de sesión/autorización (sección
propia): `go test` (unitarias), `go test -tags=integration` (repositorio contra PostgreSQL real,
incluida la carrera anti-bloqueo, y store de sesiones/contadores contra Redis real — ambos
servicios levantados con `testcontainers-go`), Vitest + Testing Library + MSW, Playwright (e2e del
flujo de acceso y gestión).

**Target Platform**: contenedores Docker en local (Docker Compose) + GitHub Actions (CI del kit,
sin tocar). Sin despliegue a producción (fuera de alcance de F2).

**Performance Goals**: login responde en <2 s con BD y Redis sanos · el listado de usuarios (≤100
filas) en <1 s · la resolución de identidad por petición añade **1 lectura Redis (sesión) + 1
consulta SQL (cuenta + rol + permisos)** por petición de panel, aceptable para el volumen del
equipo de la iglesia (decenas de cuentas) · la consulta de auditoría pagina y filtra por cuenta y
rango de fechas sobre índices propios (<1 s incluso con decenas de miles de registros; el volumen
real de este panel son decenas de intentos y acciones al día).

**Constraints**: stack fijo (constitución) · capas `handler → service → repository` (§II) ·
contrato OpenAPI antes que el código (§II, R8) · migraciones versionadas e inmutables (§VI) ·
authn/authz siempre en servidor (§IV, CWE-862) · contraseñas solo bcrypt/argon2 (§IV, CWE-256) ·
archivos del kit no editables (regla 10 de AGENTS.md; en particular `.github/workflows/ci.yml`, que
F2 resuelve con `testcontainers-go` en las pruebas — R19) · dependencias minimizadas y justificadas
(D-A8) · **la spec no se reabre**: los huecos se marcan como decisión o pregunta (R3, R15).

**Scale/Scope**: 18 operaciones REST (1 dominio backend `usuarios`, 5 features frontend — añade la
sección de auditoría y el Inicio del panel), 6 tablas PostgreSQL (`users`, `roles`, `permissions`, `role_permissions`,
`login_events`, `admin_actions`) + 4 familias de claves Redis, 3 migraciones (`000002`…`000004`,
sin huecos), 6 middlewares (4 nuevos + CORS ampliado), 1 puerta de aprobación (D-A7 —
**confirmada** el 2026-10-04; **el plan vuelve a la puerta** tras el cambio de alcance de la
auditoría, que añade las operaciones de consulta `GET /api/v1/admin/auditoria/…`).

## Constitution Check

*GATE: debe pasar antes de la investigación y re-verificarse tras el diseño.*

| Principio | Evaluación | Resultado |
|---|---|---|
| §I La spec manda | El plan implementa FR-001…FR-026 sin interpretarlos; cada uno tiene fila en "Cobertura de requisitos". Los puntos que la spec dejaba abiertos están **cerrados con decisión humana**, no supuestos: D-A7 (confirmada el 2026-10-04, con sesión en **Redis**), los tiempos de sesión (1 h absoluta + 30 min de inactividad), los datos de la cuenta (nombre, apellidos, correo, teléfono — R20) y el significado de "identificación" (R3, resuelto). Los requisitos de **auditoría (FR-021…FR-026)** vienen del **cambio de alcance aprobado por el humano el 2026-10-04** (US8), no de interpretación propia, y se implementan tal cual —incluido que la consulta use el permiso de administrar usuarios y roles—. Ningún requisito fuera de alcance se construye (sin auto-servicio por correo, sin eliminación de cuentas, sin registro público, sin purga/exportación del registro). | ✅ |
| §II Arquitectura | Monorepo; capas `handler → service → repository` sobre `internal/platform/` (reglas R1–R8 de `arquitectura.md`); API REST JSON documentada en `backend/api/openapi.yaml`, con el delta redactado en esta fase **antes** que el código (`contracts/openapi.yaml`, §8.1.5). Dependencias nuevas: 3 de runtime en backend + 1 solo de pruebas de integración + 3 en frontend, cada una justificada (R16). | ✅ |
| §III Pruebas | Toda tarea lleva sus pruebas (se exigirá en `tasks.md`); cobertura ≥80 % en `service/` (verificable con `go test -cover`); repositorio contra PostgreSQL real con `//go:build integration` (incluida prueba de concurrencia del anti-bloqueo) y sesión/contadores contra Redis real (ambos servicios con `testcontainers-go`, R19); frontend con Vitest + Testing Library + MSW; flujos críticos con Playwright (local). | ✅ |
| §IV Seguridad | Contraseñas con **bcrypt cost 12** (CWE-256), nunca en texto plano ni en respuestas (FR-003); SQL solo parametrizado vía sqlc (CWE-89); validación de toda entrada en backend (`platform/validate`, CWE-20) aunque el frontend valide; authn/authz **en servidor** en cada operación de panel (CWE-862); secretos (`SESSION_SECRET`, `BOOTSTRAP_TOKEN`) por variables de entorno, `.env.example` sin valores reales; CSRF en todo método inseguro con sesión (P10); sin `dangerouslySetInnerHTML` ni tokens en `localStorage` (skill). **Privacidad de los datos de auditoría (FR-026)**: ningún registro contiene contraseñas ni credenciales —de un login solo su resultado y de un restablecimiento quién/sobre qué/cuándo— y se aplica el **mínimo dato necesario** (sin *user-agent* ni el correo de un intento no identificado); el registro es de **solo lectura** (FR-025: sin endpoints de escritura) y su consulta exige el mismo permiso de administrar usuarios y roles (FR-024). `govulncheck`/`npm audit` en el CI del kit. | ✅ |
| §V Calidad | `gofmt`/`go vet`/`golangci-lint` sin errores; TS `strict` sin `any`; errores envueltos con `%w` y traducidos solo en `WriteError`; nombres en inglés dentro del código. | ✅ |
| §VI Base de datos | 3 migraciones versionadas con `up`/`down` completos (`000002`…`000004`, sin huecos); nunca se edita una aplicada; tablas con `id`, `created_at`, `updated_at`; FK e índices explícitos; `CHECK`/`UNIQUE` en la base (ver `data-model.md`). La auditoría duradera son las tablas `login_events`/`admin_actions` (solo inserción, FR-025) más la proyección `users.last_login_*`; la sesión y los contadores de acceso viven en Redis por decisión humana (no son tablas; su contrato de claves/TTL está documentado en `data-model.md`, junto con la relación Redis-efímero ↔ PostgreSQL-duradero). | ✅ |
| §VII Observabilidad | Logs `log/slog` con `request_id` (incluidos los intentos de acceso fallidos, sin credenciales); la auditoría añade su **fuente duradera** en PostgreSQL (`login_events`/`admin_actions`, FR-022/FR-023) además del log, con su consulta en el panel (US8); `/healthz` intacto y sin sesión (sigue siendo el chequeo operativo); config por variables de entorno (`platform/config` ampliado); todo levantable con `make up` (ahora incluye el servicio `redis`, sin persistencia). | ✅ |
| §VIII Gobierno | Este plan requiere aprobación humana antes de `/speckit.tasks` y de cualquier código (D-A7 ya está confirmada; **el plan vuelve a la puerta** tras el cambio de alcance de la auditoría, aprobado por el humano el 2026-10-04 y ya reflejado en la spec). Sin despliegue a producción (fuera de alcance). Quien escribe no aprueba: todo pasa por `qa-tester`, `revisor-codigo` y `seguridad`. | ✅ |

**Sin violaciones que justificar** → "Complexity Tracking" solo registra desviaciones de
convenciones internas (receta de archivos), no de la constitución.

## Decisiones técnicas (resumen; análisis completo en `research.md`)

Numeración **P1…P23** propia de este plan (no confundir con D1–D23 de F1 ni D-A1…D-A9 de
`docs/tecnico/decisiones.md`). **P20–P23** son del **cambio de alcance: auditoría** (aprobado el
2026-10-04).

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| **P1** | **D-A7 aterrizada (CONFIRMADA el 2026-10-04, con Redis)**: sesión en servidor **en Redis**; cookie `ss_session` con `HttpOnly`, `SameSite=Lax`, `Secure` configurable (`SESSION_COOKIE_SECURE`), `Path=/`, `Max-Age` = vida absoluta (1 h); valor = token aleatorio de 32 bytes (`crypto/rand`) y en Redis solo su **SHA-256** (clave `sess:<sha256>`; índice `user_sessions:<userId>` para revocar por cuenta); permisos por módulo en `permissions`/`role_permissions`; revocación = `DEL` de las claves + comprobación por petición de que la cuenta sigue activa | Es la decisión explícita del humano (adoptar/probar Redis como objetivo de aprendizaje); revocación inmediata al desactivar (FR-012), expiración por TTL nativa y un único almacén de estado efímero compartido con los intentos de acceso | Tabla PostgreSQL `sessions` (la propuesta original: sin servicios nuevos y transaccional con `users`; **descartada por la decisión humana** y conservada como plan B); JWT autocontenido (revocación forzada con lista negra); cookie firmada sin estado en servidor (sin revocación real); token en `localStorage` (prohibido, CWE-79) |
| P2 | Router: **se mantiene `net/http` tras `Registrar`** (cierra la pregunta 3 de `decisiones.md`/D-A4) | Los grupos con permisos (`Group("/api/v1/admin", authn, authz, CSRF)`) ya se expresan con `Group`; chi solo aportaría sintaxis y cuesta dependencia + adaptador | Adoptar chi ahora (se reevalúa si el ruteo se complica; el cambio sigue limitado a `platform/httpserver`) |
| P3 | `platform/validate` **propio mínimo** (cierra la pregunta 4 de `decisiones.md`): reflexión sobre las etiquetas `validate` de los DTOs (`required`, `omitempty`, `min`, `max`, `email`, `oneof`) → `apperr.Invalid` con `details` por campo | Sin dependencias (D-A8); mensajes en español y por campo; los 5 tags que usamos no justifican una librería | `go-playground/validator` (dependencia + transitivas, mensajes genéricos en inglés, superficie enorme para nuestro uso) |
| P4 | Tipos UUID: `github.com/google/uuid` en el dominio, `pgtype.UUID` solo en `repository.go` (cierra la pregunta 5; §8.1.1) | Mantiene `pgx` fuera de las capas altas (R4); solo tipos, sin transitivas | `pgtype.UUID` en el dominio (arrastra `pgx`); `override` de sqlc (una segunda regla para un tipo) |
| P5 | Hash de contraseñas: **bcrypt cost 12** (`golang.org/x/crypto/bcrypt`); política FR-010 centralizada en `platform/password` y aplicada en los 3 flujos (creación, restablecimiento, cambio propio) | Cumplimiento literal de §IV; bcrypt es el estándar con más revisión para este caso; el hash es autodescriptivo (futura migración a argon2id sin romper nada) | argon2id (más resistente a GPU pero más parámetros que afinar y la misma dependencia; decisión reversible); pbkdf2 de la stdlib (no es ni bcrypt ni argon2, §IV lo exige) |
| P6 | Bloqueo FR-006 con **contadores en Redis** (`login:fail:<correo>` / `login:block:<correo>`, TTL 15 min) **por identificador normalizado (correo) exista o no la cuenta**: 5 fallos → bloqueo de 15 min (constantes de código, valores confirmados por el humano). **Semántica del 5.º intento**: el contador se incrementa con cada fallo; el **5.º fallo** responde el error genérico `401` **y crea el bloqueo**; desde el **6.º intento** (y durante los 15 min) la respuesta es `429` con `Retry-After` | Concilia FR-003 y FR-006: el mensaje de bloqueo es idéntico para cuentas reales e inexistentes y por tanto **no** revela existencia; con Redis ya presente por la sesión (P1), el TTL sustituye a la limpieza manual y el estado se comparte entre instancias | Contadores solo en `users` (el bloqueo solo ocurriría en cuentas reales → enumera); tabla `login_attempts` en PostgreSQL (una migración y limpieza para un estado que no debe durar); contadores en memoria (se pierden al reiniciar); rate-limit por IP como única medida (no frena la prueba masiva contra una cuenta); responder `429` en el propio 5.º intento (contradice el Edge Case de la spec: el 5.º recibe el error genérico) |
| P7 | Regla anti-bloqueo (FR-008) verificada **dentro de la transacción** de la mutación, con `pg_advisory_xact_lock` sobre una clave fija + recuento **post-mutación** de cuentas activas con el permiso `admin_usuarios_roles`; 0 → `409 conflict` y rollback | Hace imposible el estado prohibido incluso con dos administradores actuando a la vez (edge case de la spec) | Comprobación sin lock (carrera: dos desactivaciones simultáneas dejan 0 administradores); trigger en la BD (lógica de negocio fuera del service, difícil de probar) |
| P8 | Inicialización única (FR-007): `POST /api/v1/setup/initialize` solo con `users` vacío, en transacción con el mismo advisory lock, que crea el rol **"Administrador"** con los 9 permisos y su cuenta; **exige el token `BOOTSTRAP_TOKEN`** (cabecera `X-Setup-Token`) y va rate-limited | El guard de BD impide repetirla (FR-007) y el token impide que un tercero se declare administrador en una instalación recién desplegada ("impedir cualquier uso abusivo", FR-007) | Sin token (ventana de robo del primer administrador en instalaciones públicas); CLI embebida (exige acceso al servidor y no es el producto); token de un solo uso impreso en logs (operación rara y secreto efímero mal resguardado) |
| P9 | Sesión (**confirmada el 2026-10-04**): expiración por inactividad de **30 minutos** (assumption de la spec) medida con el TTL de la clave Redis (refrescado en cada actividad y con `last_seen_at` estrangulado a 1/min) y **vida absoluta de 1 hora** desde el login (fijada como `absoluteExpiresAt` inmóvil en la clave; el TTL se acota a ella — R15); al desactivar una cuenta o restablecer su contraseña se revocan sus sesiones (`DEL` vía `user_sessions:<userId>`); `authn` comprueba **por petición** que la cuenta sigue activa | FR-005/FR-012/FR-018: los cambios de rol, permisos y estado se reflejan desde la primera acción posterior sin tocar cookies; 1 h de vida absoluta corta el alcance de una cookie robada | Expiración solo por inactividad (una cookie robada vive para siempre); vida absoluta de 12 h (propuesta inicial: demasiado larga para un panel interno); refresco de permisos en caché (rompería FR-018/SC-009) |
| P10 | CSRF obligatorio en todo método no seguro con sesión: *double-submit* **firmado** — cookie `csrf_token` (no `HttpOnly`) = `nonce.HMAC-SHA256(SESSION_SECRET, nonce)` + cabecera `X-CSRF-Token` igual, verificado por `middleware.CSRF` en el grupo (cadena de `arquitectura.md` §6) | Sin estado extra ni consultas; el HMAC impide que un atacante que inyecta cookies fabrique un par válido; `SESSION_SECRET` (ya en `.env.example`) por fin tiene uso | Token guardado en la sesión (doble consulta); exigir solo cabecera personalizada (débil); fiarlo todo a `SameSite` (no cubre toda la superficie); frameworks de CSRF (dependencia) |
| P11 | CORS se **amplía sin dependencia** (revisión de D16 cerrada): `Access-Control-Allow-Credentials: true`, eco exacto del `Origin` permitido (nunca `*` con credenciales), cabeceras `Content-Type, X-CSRF-Token, X-Request-ID`, `Vary: Origin` | Las cookies de sesión exigen credenciales; 40 líneas bastan y el middleware ya existe | `rs/cors` (dependencia para lo que ya está escrito) |
| P12 | Modelo: **un rol por cuenta** → `users.role_id` (se **simplifica** el `user_roles` orientativo de F1) con **nombre, apellidos, correo y teléfono** (`first_name`, `last_name`, `email`, `phone` — datos confirmados el 2026-10-04, R20); catálogo de permisos fijo en tabla `permissions` **sembrada por la migración** + `role_permissions`; sesión y FR-006 **en Redis**, no en tablas (detalle en `data-model.md`) | Refleja literalmente Q4 (un rol) y los datos de la cuenta de la spec, da FK y unicidad reales y permite al panel listar el catálogo con `GET /admin/permisos` | `user_roles` (contradice Q4); permisos como `TEXT[]` en `roles` (sin FK, difícil de validar); permisos solo como constantes de código (el panel no podría listarlos con etiqueta sin duplicar el catálogo); `full_name` único (la spec pide nombre y apellidos separados) |
| P13 | Normalización (Q5): el correo se guarda `trim`+minúsculas y el nombre de rol `trim`+colapso de espacios (conservando sus mayúsculas de presentación), con `UNIQUE` real y `UNIQUE (lower(name))`; duplicados → `409 conflict` | El mismo dato nunca vive en dos formas y la unicidad la garantiza la BD, no solo el código | Índice funcional sobre el valor crudo (permite guardar duplicados "casi"); comparar solo en el service (carrera entre dos creaciones simultáneas) |
| P14 | `platform/paginate` (diferido de F1) para `GET /admin/usuarios` y `GET /admin/roles`: `limit` 20 por defecto, tope 100, sobre `{items, total, limit, offset}` (§8.1.2/§8.1.3). **Sin buscador de texto libre** en el listado de usuarios en el MVP (decisión del `analyze` F-04: no está en la spec; solo paginación) | Primer listado de panel con parámetros de usuario; convención ya cerrada | Listados sin acotar (rompe §8.1.3); paginación por cursor (innecesaria en listados de decenas de filas); buscador de texto libre en cliente (fuera de la spec: se retira del MVP y vuelve solo con requisito) |
| P15 | Un **único dominio `internal/usuarios/`** para todo F2 + `platform/session` (token y cookie, tipos `Identity`/`Resolver` y el `Store` de sesiones **sobre Redis**, sin SQL ni dominio) + `platform/password`; los archivos de la receta se **dividen por responsabilidad** (`service_auth.go`, `handler_users.go`…) — desviación declarada en "Complexity Tracking" | Un dominio evita que dos paquetes consulten las mismas tablas (R2) y encaja con "transversal en `platform/` + dominio `usuarios`" (F1 `data-model.md`); el acceso a Redis es plumbing, como el de PostgreSQL en `platform/database`; los archivos por responsabilidad mantienen funciones cortas | Dominios `auth/` + `usuarios/` separados (obliga a consultas compartidas y a una interfaz extra para revocar sesiones); la persistencia de sesiones dentro del repository del dominio (mezclaría Redis con sqlc); un `handler.go` monolítico de 18 endpoints |
| P16 | `apperr` crece con los kinds ya registrados en el contrato: `Invalid` (400), `Unauthenticated` (401), `Forbidden` (403), `Conflict` (409), `RateLimited` (429, con `Retry-After`) | El registro estaba cerrado y previsto para "crecer bajo demanda con F2+" (`arquitectura.md` §5.11) | Códigos nuevos fuera del registro (rompería el contrato) |
| P17 | `rate-limit` mínimo **en memoria** (cierra el punto de decisión de D23): ventana deslizante por IP sobre los endpoints públicos escribibles (`/auth/login`, `/setup/initialize`), 429 `rate_limited` | El primer endpoint público escribible ha llegado con F2; amortigua la prueba masiva de contraseñas por IP además del bloqueo por cuenta | No hacerlo (deja el diferido sin decidir); Redis/limiter externo (hoy Redis ya existe por la sesión, P1, pero el umbral por IP no necesita compartirse con una sola instancia: se moverá a Redis si llega el escalado) |
| P18 | Frontend: las **7 rutas** `/login`, `/cambiar-contrasena`, `/panel`, `/panel/usuarios`, `/panel/roles`, `/panel/auditoria` y `/sin-permiso`; guards `RequireAuth`, `RequirePermission` y `RequirePasswordChange`; sesión vía TanStack Query (`GET /auth/session`); formularios con React Hook Form + Zod; menú filtrado por permisos (US4 esc. 6). **Inicio del panel (`/panel`)**: tarjeta "Mi cuenta", accesos rápidos a las secciones autorizadas y el estado **"cuenta sin permisos de módulo"** (Edge Case de la spec: entra pero no ve ninguna sección de gestión, y eso no rompe el panel) | Es la convención de la skill y el patrón que copiarán F3–F9; los permisos se ocultan **y** se deniegan en servidor; el Inicio es la única pantalla que toda cuenta ve y da cobertura al Edge Case de la cuenta sin permisos | Estado global propio (redux/zustand: no hace falta); fetch en componentes (prohibido); validación manual de formularios; rutas propias de `ux.md` (`/entrar`, `/panel/cuenta`): **descartadas**, se usan las 7 rutas del plan |
| P19 | Contrato: delta en `specs/002-acceso-gestion-usuarios/contracts/openapi.yaml`, fusionado en `backend/api/openapi.yaml` al implementar (§8.1.5) con `info.version` **0.2.0 → 0.3.0**; esquema de seguridad `sessionCookie`; el registro de `error.code` **no cambia** (F2 emite `invalid`, `unauthenticated`, `forbidden`, `conflict`, `rate_limited`, ya en el enum, y **reutiliza `not_found`** de F1 para "cuenta/rol inexistente") | Contrato antes que el código (§II) sin dos copias vivas | Editar el snapshot de `specs/` a posteriori (prohibido); códigos nuevos ad-hoc por endpoint |
| **P20** | **Auditoría duradera (cambio de alcance)**: tablas `login_events` (cada intento de acceso, FR-022) y `admin_actions` (cada acción administrativa, FR-023) en PostgreSQL, de **solo inserción**; el registro se escribe desde el dominio (`service_audit.go` + `repository_audit.go`) y las denegaciones de permiso pasan por la interfaz de plumbing `audit.Recorder` (`internal/platform/audit`, mismo precedente que `session.Identity`/`Resolver`), que implementa el dominio. Puntos de escritura: (1) `service_auth.go` — todo intento de login (éxito, fallo, cuenta inactiva, intento durante el bloqueo); (2) los services de gestión — cada operación sensible, éxito **y** fallo (en éxito, `INSERT` en la misma transacción que la mutación); (3) el helper común de decodificación/validación del `handler.go` (JSON inválido o DTO no válido) y `authz` al denegar (403 → `result='denied'`) | FR-022/FR-023 piden registrar **cada** intento y **cada** acción sensible, también las que fallan o se deniegan; con la capa de servicio como dueña del registro, el objetivo (id y etiqueta) se conoce de forma fiable y el éxito es transaccional con su mutación | Registro solo en el handler (no ve lo que falla dentro del service ni el id recién creado); registro en un middleware único que lee la respuesta (acoplado a la forma de los DTO y no distingue el objetivo de una creación); triggers de BD (lógica de auditoría fuera del service, difícil de probar y de traducir a `apperr`); solo éxito (contradice el Edge Case de la spec) |
| **P21** | **Último acceso por cuenta** (FR-021): columnas `users.last_login_at` / `users.last_login_ip`, escritas **solo** por el login exitoso junto a su fila en `login_events`; anulables y no editables desde la API | La ficha y el listado muestran el último acceso con la fila de `users`, sin consultar una tabla que crece sin límite, y siguen mostrándolo aunque en el futuro haya retención del registro (riesgo RG17) | Derivarlo de `login_events` (`WHERE user_id=$1 AND result='success' ORDER BY created_at DESC LIMIT 1`): consulta extra por cuenta y la ficha depende del crecimiento/retención del historial (descartado; queda como fuente de verdad y prueba de coherencia) |
| **P22** | **Consulta de la auditoría (solo lectura)**: `GET /api/v1/admin/auditoria/accesos` y `GET /api/v1/admin/auditoria/acciones`, con filtros `userId` y rango `from`/`to` (semirango `[from, to)`) y paginación `platform/paginate` (§8.1.3, defecto 20 / tope 100, orden `createdAt DESC`); **mismo permiso** `admin_usuarios_roles` que la gestión (decisión de la spec) y **ninguna** operación de escritura en el contrato (FR-025). En `/acciones`, `userId` filtra por la cuenta **involucrada** (la que hizo la acción o la sobre la que se hizo). Los DTOs exponen `userName`/`userEmail` y `actorName`/`actorEmail` **derivados por `JOIN` con `users`** (no se guardan como dato duplicado); un intento sin cuenta asociada va con esos campos en `null` y **sin guardar ni mostrar correo alguno** (la UI indica "Intento sin cuenta asociada") | Cumple literalmente FR-024 (filtros por cuenta y fechas + paginación) reutilizando el permiso existente sin ensanchar el catálogo (Decisión 5/FR-015) y deja fuera toda vía de edición del registro | Permiso propio de "auditoría" (la spec lo descarta: la auditoría no es un módulo del producto); un solo endpoint con `type` (dos historiales con DTOs distintos quedan mejor separados); filtros `actorId`/`targetId` por separado (la spec pide "filtrar por cuenta", uno solo); paginación por cursor (innecesaria para el volumen; P14) |
| **P23** | **Redis sin persistencia (confirmado por el humano el 2026-10-04)**: el servicio `redis` corre **sin `appendonly`, sin `save` y sin volumen**; un reinicio solo obliga a re-loguearse y reinicia los contadores de intentos, y **no toca la auditoría** (que vive en PostgreSQL) | Es la decisión confirmada del humano; el estado de Redis (sesión y contadores) es efímero por definición y la dureza que importa —el registro— va a PostgreSQL | AOF/RDB y volumen en Redis (descartado por decisión humana: nada de lo que guarda Redis es duradero; se reabriría solo si se exige dureza de sesión); guardar el registro también en Redis (la auditoría debe durar y sobrevivir reinicios) |

## Project Structure

### Documentation (esta funcionalidad)

```text
specs/002-acceso-gestion-usuarios/
├── plan.md              # EDITABLE (este archivo, /speckit.plan)
├── research.md          # EDITABLE (fase 0: R1…R23, incluye D-A7 confirmada con Redis, Redis sin persistencia y auditoría)
├── data-model.md        # EDITABLE (fase 1: tablas, migraciones, consultas)
├── quickstart.md        # EDITABLE (fase 1: cómo probar F2 en local)
├── spec.md              # APROBADA — no se edita
├── ux.md                # DEL disenador-ux (entregado; pendiente alinear sus campos de usuario con los confirmados el 2026-10-04 y añadir la sección de auditoría — riesgo RG2)
├── contracts/
│   └── openapi.yaml     # EDITABLE (fase 1): delta de diseño, se fusiona en backend/api/openapi.yaml
├── checklists/          # (existente)
└── tasks.md             # Fase 2 (/speckit.tasks — NO lo crea este comando)
```

### Source Code — backend

```text
backend/
├── cmd/api/
│   ├── main.go                      # EDITADO: DI de F2 (incl. cliente Redis) + grupos /api/v1/auth, /api/v1/setup y /api/v1/admin
│   └── main_test.go                 # EDITADO: prueba de humo ampliada (rutas nuevas + /healthz intacto)
├── internal/
│   ├── db/
│   │   ├── queries/
│   │   │   ├── users.sql            # NUEVO: cuentas (CRUD + authn + recuento anti-bloqueo + último acceso)
│   │   │   ├── roles.sql            # NUEVO: roles + role_permissions
│   │   │   ├── permissions.sql      # NUEVO: catálogo
│   │   │   └── audit.sql            # NUEVO (cambio de alcance): login_events + admin_actions (solo INSERT/SELECT)
│   │   └── (generado por sqlc, commiteado)   # sin sessions.sql ni contadores de FR-006: viven en Redis
│   ├── usuarios/                    # NUEVO — dominio F2 (área completa: acceso + gestión + auditoría)
│   │   ├── model.go                 # entidades, estados y DTOs (etiquetas validate)
│   │   ├── repository.go            # constructor + mapRow/params (pgtype → dominio)
│   │   ├── repository_users.go      # cuentas (PostgreSQL)
│   │   ├── repository_roles.go      # roles + permisos + guard anti-bloqueo
│   │   ├── repository_audit.go      # NUEVO (auditoría): inserción y consulta de login_events/admin_actions
│   │   ├── service.go               # tipos comunes + invariantes compartidos
│   │   ├── service_auth.go          # login, logout, sesión, cambio de contraseña, inicialización (usa session.Store y los contadores Redis) + registro de cada intento en login_events
│   │   ├── service_users.go         # CRUD de cuentas, activar/desactivar, restablecer
│   │   ├── service_roles.go         # CRUD de roles y catálogo de permisos
│   │   ├── service_audit.go         # NUEVO (auditoría): registro de acciones (éxito y fallo) + consulta con filtros/paginación
│   │   ├── handler.go               # constructor + helpers HTTP (el de decodificar/validar registra el intento si lo rechaza)
│   │   ├── handler_auth.go          # /api/v1/auth/* + /api/v1/setup/*
│   │   ├── handler_users.go         # /api/v1/admin/usuarios*
│   │   ├── handler_roles.go         # /api/v1/admin/roles* + /api/v1/admin/permisos
│   │   ├── handler_audit.go         # NUEVO (auditoría): /api/v1/admin/auditoria/* (solo GET)
│   │   ├── routes.go                # RegisterPublic (login, setup) + RegisterAdmin (usuarios/roles/auditoría)
│   │   └── *_test.go                # service con fakes · handler con httptest · repository integration (testcontainers)
│   └── platform/
│       ├── apperr/apperr.go         # EDITADO: + Invalid, Unauthenticated, Forbidden, Conflict, RateLimited
│       ├── config/config.go         # EDITADO: + REDIS_URL, SESSION_SECRET, SESSION_* , SESSION_COOKIE_SECURE, BOOTSTRAP_TOKEN
│       ├── session/                 # NUEVO: token aleatorio + SHA-256, cookies ss_session/csrf_token, tipos Identity y Resolver, interfaz Store + implementación Redis (claves sess:* y user_sessions:*)
│       ├── audit/                   # NUEVO (auditoría): plumbing — tipos Event/Action y la interfaz Recorder que implementa el dominio (para que authz registre denegaciones sin conocerlo)
│       ├── password/                # NUEVO: bcrypt (hash/verify) + política FR-010
│       ├── validate/                # NUEVO (P3): validación de DTOs por etiquetas (incl. `phone`) → apperr.Invalid
│       ├── paginate/                # NUEVO (P14): limit/offset con topes
│       └── middleware/
│           ├── authn.go             # NUEVO: cookie → session.Store (Redis) → session.Resolver → Identity en el contexto
│           ├── authz.go             # NUEVO: AuthzByModule(código) sobre la Identity
│           ├── csrf.go              # NUEVO: double-submit firmado en métodos no seguros
│           ├── ratelimit.go         # NUEVO (P17): ventana deslizante por IP
│           ├── cors.go              # EDITADO (P11): credenciales + cabeceras nuevas
│           └── chain.go             # EDITADO si hace falta: orden de los grupos
├── migrations/
│   ├── 000002_create_roles_and_permissions.up.sql / .down.sql   # NUEVO (incluye la siembra del catálogo)
│   ├── 000003_create_users.up.sql / .down.sql                   # NUEVO (first_name, last_name, email, phone)
│   └── 000004_create_login_events_and_admin_actions.up.sql / .down.sql   # NUEVO (cambio de alcance: auditoría + users.last_login_*)
└── api/openapi.yaml                 # EDITADO: se fusiona el delta (info.version → 0.3.0)
```

### Source Code — frontend y raíz

```text
frontend/src/
├── api/
│   ├── client.ts                    # EDITADO: credentials include + cabecera X-CSRF-Token + ApiError
│   ├── auth.ts                      # NUEVO: login, logout, getSession, changeMyPassword
│   ├── usuarios.ts                  # NUEVO: CRUD de cuentas + reset
│   ├── roles.ts                     # NUEVO: CRUD de roles + catálogo de permisos
│   ├── auditoria.ts                 # NUEVO (auditoría): historial de accesos y de acciones (solo lectura, con filtros y paginación)
│   └── schema.d.ts                  # REGENERADO (npm run api:gen)
├── app/
│   ├── router.tsx                   # EDITADO: /login, /cambiar-contrasena, /panel, /panel/usuarios, /panel/roles, /panel/auditoria y /sin-permiso (las 7 rutas de P18) con guards
│   ├── guards.tsx                   # NUEVO: RequireAuth, RequirePermission, RequirePasswordChange
│   └── layout.tsx                   # EDITADO: layout del panel con navegación filtrada por permisos
├── features/
│   ├── auth/                        # NUEVO: LoginPage, ChangePasswordPage, hooks useSession/useLogin/…
│   ├── panel/                       # NUEVO: InicioPage (tarjeta "Mi cuenta", accesos rápidos y estado "cuenta sin permisos de módulo", Edge Case de la spec)
│   ├── usuarios/                    # NUEVO: UsersPage, UserForm (con último acceso en la ficha = detalle/edición, FR-021), hooks useUsers/useCreateUser/…
│   ├── roles/                       # NUEVO: RolesPage, RoleForm (permisos), hooks useRoles/…
│   └── auditoria/                   # NUEVO (auditoría): AuditPage (dos historiales, filtros por cuenta y fechas, paginación; sin acciones de edición), hooks useAccessEvents/useAdminActions
├── components/                      # EDITADO: inventario único de 13 componentes compartidos (ver "Inventario único de componentes UI" más abajo): Field, PasswordField, Select, Button, Notice, ConfirmDialog, Dialog, EmptyState, Table, Tabs, Pagination, DateRangeFilter, StatusPill
└── lib/                             # EDITADO: helpers de permisos y formato
frontend/e2e/acceso.spec.ts          # NUEVO: flujo completo Playwright (ver quickstart §11)
frontend/e2e/auditoria.spec.ts       # NUEVO (auditoría): sección de registro, filtros y solo lectura (ver quickstart §10)

# Raíz
.env.example                         # EDITADO: + REDIS_URL, BOOTSTRAP_TOKEN, SESSION_SECRET (ya existe), SESSION_* y su documentación
README.md                            # EDITADO por documentador al cerrar: comandos y variables nuevas
docker-compose.yml                   # EDITADO: servicio `redis` (redis:7-alpine, healthcheck, puerto — SIN volumen ni appendonly: persistencia desactivada a propósito, P23) + variables del backend (REDIS_URL, SESSION_*, BOOTSTRAP_TOKEN)
Makefile / .github/workflows/ci.yml / .githooks/   # SIN CAMBIOS (son del kit; ci.yml NO se toca: las pruebas de integración levantan Redis con testcontainers-go, R19)
```

### Inventario único de componentes UI (F-10, el que cita `ux.md`)

Es el **único** inventario de componentes compartidos de F2 (T242): **13 componentes** con
**nombres de código en inglés**. `ux.md` (`disenador-ux`) lo **cita** y da el mapeo a sus etiquetas
en español, sus props y sus textos; este plan fija los nombres de código. Ninguno sobra ni falta
para las pantallas de F2: los estados de **carga y error** también los cubre `EmptyState` (según
`ux.md`), por lo que no hace falta un componente aparte.

| Componente (código) | Responsabilidad |
|---|---|
| `Field` | campo de formulario con etiqueta, error y `autocomplete` |
| `PasswordField` | campo de contraseña: mostrar/ocultar y checklist en vivo de la política FR-010 |
| `Select` | selector (rol en el formulario de usuario, entre otros) |
| `Button` | acciones y envío de formularios (variantes, `loading`, mínimo 44 px) |
| `Notice` | avisos de sistema con `aria-live` |
| `ConfirmDialog` | confirmación accesible (desactivar cuenta, eliminar rol) |
| `Dialog` | diálogo accesible con los formularios en modal (crear/editar usuario/rol) — **sustituye a `ModalDialog`** de las rondas previas |
| `EmptyState` | estados vacío, cargando y error |
| `Table` | tabla accesible (en móvil, lista de tarjetas) |
| `Tabs` | los dos historiales de la auditoría |
| `Pagination` | `limit`/`offset`, conservando los filtros |
| `DateRangeFilter` | rango de fechas `from`/`to` de la auditoría |
| `StatusPill` | estado de cuenta y resultado de registro (texto además de color) |

Las features los **reutilizan sin duplicar markup** (lo revisa `revisor-codigo`). Un componente
fuera de esta lista exige **primero** actualizarla aquí y en T242; cualquier cambio de nombre se
propaga a `ux.md` para que la cita siga siendo inequívoca.

## Cadena de middleware y grupos (orden; el primero es el más externo)

```text
Global:   request-id → recover → logging → CORS(con credenciales, P11) → handler
Grupos:   /api/v1/setup                    → rate-limit (P17)                       → handler
          /api/v1/auth  (público)  POST /login → rate-limit (P17)                   → handler
          /api/v1/auth  (sesión)            → authn → CSRF                          → handler   ← SIN guard (rutas blanqueadas)
          /api/v1/admin                     → authn → guard de cambio de contraseña → CSRF → (subgrupo) authz(módulo) → handler
```

- `authn` (P1/P9): lee `ss_session`, resuelve la sesión en **Redis** vía `session.Store`
  (comprobando inactividad y vida absoluta), resuelve la identidad vía `session.Resolver` (interfaz
  definida en `platform/session`, implementada por `usuarios.Service` — R3), comprueba que la
  cuenta siga activa y la deja en el contexto. Sin sesión válida → `401 unauthenticated`.
- **Guard de cambio de contraseña**: montado **solo en el grupo `/api/v1/admin`** (entre `authn` y
  `authz`; lo cablea T228 en `routes.go`/`main.go`). Las rutas **blanqueadas** —`/api/v1/auth/session`,
  `/api/v1/auth/logout` y `/api/v1/auth/password`— viven en el grupo `/api/v1/auth` (sesión), que
  **no monta el guard**, para que una cuenta con `mustChangePassword` pueda ver su sesión, salir y
  cambiar la contraseña; `/api/v1/auth/login` y `/api/v1/setup/initialize` son públicos. Con
  `mustChangePassword`, todo lo demás del grupo `/api/v1/admin` → `403 forbidden` con
  `details.reason = "password_change_required"` (la UI redirige al formulario). Los grupos de
  F3–F9 montarán el mismo guard en su subgrupo de panel.
- `authz(módulo)` (P16/FR-016): exige el permiso del módulo (p. ej. `admin_usuarios_roles`) en la
  identidad resuelta; sin permiso → `403 forbidden` con mensaje claro y **el intento queda
  registrado** en `admin_actions` con `result='denied'` vía `audit.Recorder` (P20: la acción y el
  objetivo se resuelven desde `method`+`path` con la tabla del dominio). F3–F9 crearán sus
  subgrupos con su módulo sin tocar este dominio.
- `CSRF` (P10): solo métodos no seguros; exige `X-CSRF-Token` = cookie `csrf_token` firmada.
- `rate-limit` (P17): solo en las rutas públicas escribibles (`/auth/login`, `/setup/initialize`).
- **Orden interno del grupo de panel**: `authn → guard → authz → CSRF` (dentro de la cadena
  `[authn → authz → CSRF]` de `arquitectura.md` §6; el guard se añade después de `authn`).
- `/healthz` **no cambia**: sigue público, sin sesión y con su `Cache-Control: no-store`.

## Contrato OpenAPI: dónde vive y cómo se mantiene

Igual que F1 (reglas de "Contrato OpenAPI" de `specs/001-estructura-base/plan.md` y §8.1.5):

1. **Diseño (esta fase)**: `specs/002-acceso-gestion-usuarios/contracts/openapi.yaml` — delta
   inmutable una vez aprobado el plan.
2. **Fusión**: la primera tarea de implementación lo funde en `backend/api/openapi.yaml` (documento
   vivo) de forma **aditiva** y sube `info.version` a **0.3.0**; luego `make api-gen`.
3. **Consumo**: el frontend genera tipos solo desde `backend/api/openapi.yaml`.
4. El snapshot de `specs/` no se vuelve a editar.

## Estrategia de pruebas (incluye sesión y autorización)

Cobertura exigida: **80 % en `service/`** (§III). Comandos: `go test ./...` ·
`go test -tags=integration ./...` · `npm test -- --run` · `make e2e` · `make ci` (**lint + test +
security**; `make db-migrate` es un comando aparte y no forma parte de `make ci`).

| Capa | Tipo | Qué verifica F2 además de lo de F1 | Dónde |
|---|---|---|---|
| `platform/password` | Unitaria (tabla de casos) | Política FR-010: longitud 8–64 caracteres, mayúsculas+minúsculas+números+especiales, **distinta de** (igualdad normalizada, no de contenido) el nombre, los apellidos y el correo; hash/verify bcrypt; comprobación técnica de bytes (límite de bcrypt) | `internal/platform/password/*_test.go` |
| `platform/session` | Unitaria | Token aleatorio distinto por llamada, SHA-256 estable, cookies con `HttpOnly`/`SameSite`/`Secure`/`Max-Age` (1 h), lectura y borrado; `Store` con fakes | `internal/platform/session/*_test.go` |
| `session.Store` (Redis) | **Integración** (Redis real, testcontainers) | Crear/resolver sesión, TTL de inactividad y su refresco acotado a la vida absoluta, corte a los 60 min aunque haya actividad, revocación por cuenta (`user_sessions:*`), contadores `login:fail`/`login:block` y `Retry-After` | `internal/platform/session/*_test.go` (`//go:build integration`) |
| `platform/validate` | Unitaria (tabla de casos) | Cada etiqueta y el `details` por campo que produce | `internal/platform/validate/*_test.go` |
| `middleware` | Unitaria + `httptest` | `authn` (sin cookie, cookie inválida, sesión expirada, cuenta inactiva), `authz` (con/sin permiso), `csrf` (método seguro pasa, inseguro sin/desalineado con token), `ratelimit` (429 tras el umbral) | `internal/platform/middleware/*_test.go` |
| `service` (dominio) | Unitaria con **fakes** | Login (éxito, credenciales genéricas, cuenta inactiva, bloqueo 5/15 min, limpieza de intentos), cambio/restablecimiento de contraseña con `mustChangePassword`, anti-bloqueo (desactivar, cambiar rol, quitar permiso al rol), reglas de roles (≥1 permiso, duplicados normalizados, eliminación solo sin uso), un rol por cuenta, **auditoría** (cada desenlace de login deja su fila en `login_events` y el éxito actualiza el último acceso; cada operación sensible deja su `admin_actions` también al fallar; nunca se registra una contraseña) | `internal/usuarios/service_*_test.go` |
| `handler` | Unitaria con `httptest` + service falso | Cada operación: decodificación, validación con `details`, códigos 200/201/400/401/403/404/409/429, sobres de éxito y de error, **que la contraseña nunca aparece en la respuesta**; endpoints de auditoría (filtros `userId`/`from`/`to`, paginación, `400` con `from > to`, `403` sin permiso, JSON inválido registrado como intento fallido) | `internal/usuarios/handler_*_test.go` |
| `repository` | **Integración** (PostgreSQL real, testcontainers si no hay `DATABASE_URL_TEST`) | SQL real, `UNIQUE` de correos/nombres normalizados, FK y `ON DELETE`, **carrera anti-bloqueo** (dos transacciones concurrentes no dejan 0 administradores) e **inicialización única** (dos `initialize` simultáneos → uno solo crea); tablas de auditoría (inserción con los `CHECK`, filtros por cuenta y fechas, paginación, `target_role_id` en NULL y `target_label` intacto al eliminar un rol, coherencia `last_login_*` ↔ última fila `success`, y que no hay consultas de `UPDATE`/`DELETE` sobre el registro) | `internal/usuarios/repository_*_test.go` (`//go:build integration`) |
| `cmd/api` | Humo de composición | Rutas nuevas publicadas con su sobre y `/healthz` intacto (§8.1.9) | `cmd/api/main_test.go` |
| Frontend | Unitaria (Vitest + MSW) | Login (errores genéricos, bloqueo, cuenta desactivada), guard de cambio de contraseña, navegación filtrada por permisos, **Inicio del panel** (tarjeta "Mi cuenta", accesos rápidos y estado "cuenta sin permisos de módulo"), formulario de cuenta (duplicado, rol inexistente, política), formulario de rol (sin permisos → error), eliminar rol en uso, sección de auditoría (dos historiales, filtros por cuenta y fechas, paginación, estado vacío, último acceso en la ficha, `403` sin permiso) | `frontend/src/features/*/*.test.tsx` |
| E2E | Playwright (local) | Recorrido completo de `quickstart.md` §11 (valida SC-002, SC-005, SC-006, SC-007, SC-010, SC-011) | `frontend/e2e/acceso.spec.ts` |
| E2E auditoría | Playwright (local) | Recorrido de `quickstart.md` §10: generar accesos y acciones, ver ambos historiales con filtros y paginación, último acceso en la ficha, intento fallido sin cuenta asociada, registro sin controles de edición (valida SC-012, SC-013) | `frontend/e2e/auditoria.spec.ts` |

**Pruebas de sesión/autorización que no faltan** (las que la spec hace críticas): desactivar una
cuenta con sesión abierta corta el acceso en la primera acción posterior (FR-012); quitar un permiso
a un rol se refleja en la siguiente petición de sus cuentas (FR-018); sin permiso no se puede forzar
la operación por API aunque la UI oculte el botón (FR-016/SC-007); los mensajes de error de login no
revelan existencia (FR-003/SC-008, incluido el caso de cuenta inactiva — ver R4); el bloqueo por
intentos es idéntico para cuentas reales e inexistentes (FR-006); **el registro de auditoría es de
solo lectura** (0 endpoints de escritura sobre él en el contrato y en las pruebas, FR-025) y **no
contiene credenciales** (FR-026); todo intento de acceso queda registrado —también el contra un
correo inexistente, sin asociarse a nada (FR-022/US8 esc. 7)— y toda acción sensible queda registrada
**también cuando falla o se deniega** (FR-023).

## Cobertura de requisitos (FR-001…FR-026)

| FR | Dónde se resuelve en este plan | Verificación |
|---|---|---|
| FR-001 | P18 (guards `RequireAuth` en el router) + `middleware/authn` (P1) | e2e + quickstart §2 (SC-001) |
| FR-002 | P1 + P5 + `service_auth.go` (login: correo+contraseña, cuenta inactiva rechazada) | quickstart §2 y §7 · pruebas del service |
| FR-003 | P6 + **R18** (verificación dummy para uniformidad de tiempos) + contrato (ninguna respuesta contiene `password` ni `passwordHash`) | quickstart §8 · suite de handlers (SC-008) |
| FR-004 | `POST /api/v1/auth/logout` (P1: `DEL` de la sesión en Redis y borrado de las cookies) | quickstart §3 |
| FR-005 | P9 (inactividad 30 min por TTL + vida absoluta de 1 h con `absoluteExpiresAt`, R15) | quickstart §3 · pruebas de `platform/session` (unitarias + Redis) |
| FR-006 | P6 (contadores `login:fail:*`/`login:block:*` en Redis, 5 fallos / 15 min como constantes; el **5.º fallo** crea el bloqueo y aún responde `401`, el `429` desde el **6.º** intento) | quickstart §8 · pruebas del service + integración |
| FR-007 | P8 (`POST /api/v1/setup/initialize` + guard de BD + `BOOTSTRAP_TOKEN`) | quickstart §1 y §8 · integración (doble inicialización) |
| FR-008 | P7 (recuento post-mutación en transacción con advisory lock) | quickstart §6 · integración concurrente (SC-004) |
| FR-009 | P3 + P13 + P12 + `POST /api/v1/admin/usuarios` (nombre, apellidos, correo, teléfono, rol, contraseña; teléfono mal formado → 400 con `details.phone`; rol inexistente → 400 con `details.roleId`) | quickstart §4 · pruebas de service/handler |
| FR-010 | P5 (política en `platform/password`: 8–64 caracteres, 4 clases de caracteres, **distinta de** —igualdad normalizada, no de contenido— el nombre, los apellidos y el correo) + P9 (`mustChangePassword`) + `POST /admin/usuarios/{id}/password` | quickstart §5 · tabla de casos de la política |
| FR-011 | `PATCH /api/v1/admin/usuarios/{id}` (firstName, lastName, email, phone, roleId, isActive) | quickstart §4 y §7 |
| FR-012 | P1 + P9 (revocación de claves de sesión al desactivar + comprobación por petición) | quickstart §3 y §7 · e2e (SC-006) |
| FR-013 | **No existe operación de borrado** de cuentas en el contrato ni en el servicio (solo `is_active`) | revisión del contrato + `revisor-codigo` |
| FR-014 | P12 + P13 + `POST/PATCH /api/v1/admin/roles` (≥1 permiso, nombre único normalizado) | quickstart §4 · pruebas de service |
| FR-015 | P12 (catálogo de 9 permisos sembrado en `000002`) | `data-model.md` · `GET /admin/permisos` |
| FR-016 | `middleware.AuthzByModule` (P1/P16) en cada operación de panel | quickstart §6 · pruebas de middleware (SC-007) |
| FR-017 | P12 (`users.role_id`, un solo rol) + `PATCH/DELETE /api/v1/admin/roles/{id}` (eliminar solo sin uso → 409) | quickstart §4 y §6 |
| FR-018 | P9 (permisos resueltos **por petición**, sin caché) | quickstart §6 · e2e (SC-009) |
| FR-019 | `GET /api/v1/admin/usuarios` (`UserItem`: estado, correo, rol) | quickstart §4 |
| FR-020 | `POST /api/v1/auth/password` (cambio propio con contraseña actual) | quickstart §5 (SC-010) |
| FR-021 *(auditoría)* | P21 (`users.last_login_at`/`last_login_ip`, solo el login exitoso) + `UserItem` del contrato (`lastLoginAt`/`lastLoginIp`, anulables) en la fila del listado y en la **ficha = detalle/edición** de la cuenta | quickstart §10 · listado y ficha de la cuenta en el e2e de auditoría (SC-012) |
| FR-022 *(auditoría)* | P20 (tabla `login_events`: cada intento con resultado e IP, `user_id` solo si la cuenta existe) + `GET /admin/auditoria/accesos` (P22) | quickstart §10 · pruebas de service/handler/integración (SC-013) |
| FR-023 *(auditoría)* | P20 (tabla `admin_actions`: quién/qué/sobre qué/cuándo/resultado, también al fallar —service, handler y `authz`—; inicialización como `user.create` sin actor) + `GET /admin/auditoria/acciones` (P22) | quickstart §10 · pruebas de service/handler/integración (SC-013) |
| FR-024 *(auditoría)* | P22 (los dos `GET /admin/auditoria/…` con filtros `userId`/`from`/`to` y paginación `platform/paginate`, bajo el permiso `admin_usuarios_roles`) + `features/auditoria` en el panel | quickstart §10 (filtros, paginación, `403` sin permiso) · e2e de auditoría (SC-012) |
| FR-025 *(auditoría)* | Solo inserción: `audit.sql` sin `UPDATE`/`DELETE`, contrato **solo con `GET`** sobre `/admin/auditoria/*`, y sin controles de edición en la UI (P20/P22) | revisión del contrato + `revisor-codigo` + quickstart §10 (SC-013) |
| FR-026 *(auditoría)* | Sin columnas de credenciales en las tablas; de un login solo `result`, de un restablecimiento solo `user.password_reset` con actor/objetivo (P20; mínimo dato: sin *user-agent* ni correos de intentos no identificados) | suite de handlers (ningún DTO de salida las incluye) + `seguridad` |

## Métricas del plan (coherentes con los SC de la spec)

| Criterio | Qué se mide | Cómo se mide en F2 | Umbral |
|---|---|---|---|
| SC-001 | Panel 100 % detrás de sesión | e2e: cada sección sin sesión redirige a `/login`; 0 accesos sin sesión | 100 % |
| SC-002 | Login en <30 s | e2e con cronómetro / observación manual (quickstart §2) | < 30 s |
| SC-003 | Inicialización única efectiva | quickstart §1 + integración con dos peticiones simultáneas | 100 % instalaciones; repeticiones bloqueadas |
| SC-004 | 0 escenarios sin administración | quickstart §6 + integración concurrente del guard anti-bloqueo | 0 |
| SC-005 | Crear cuenta + rol + asociar en <3 min | e2e / observación (quickstart §4) | < 3 min |
| SC-006 | Cuenta desactivada sin acceso (ni con sesión) | quickstart §7 + e2e | 100 % |
| SC-007 | Toda operación de gestión verifica permiso | pruebas de `authz` por operación + quickstart §6 | 0 operaciones sin permiso |
| SC-008 | Ningún error de login revela existencia | suite de handlers + quickstart §8 (mensajes observados) | 100 % |
| SC-009 | Cambio de permisos efectivo desde la primera acción | quickstart §6 + e2e | 100 % |
| SC-010 | Cambio de contraseña y reingreso <1 min | quickstart §5 | < 1 min |
| SC-011 | 0 eliminaciones de rol en uso, 0 repeticiones de inicialización, 0 duplicados casi idénticos | quickstart §4 y §8 + integración de `UNIQUE` | 0 |
| SC-012 *(auditoría)* | Encontrar los accesos y las acciones de una cuenta en <1 min | quickstart §10 (filtros por cuenta y rango de fechas) + e2e de auditoría | < 1 min |
| SC-013 *(auditoría)* | 100 % de intentos y acciones registrados; 0 registros editables/borrables | quickstart §10 + pruebas de service/handler/integración + revisión del contrato | 100 % / 0 |

## Riesgos

> Numeración **RG1–RG19**, propia de este plan: se usa `RG` para **no colisionar con R1–R23 de
> `research.md`**. Dentro de esta tabla, las referencias `Rn` (p. ej. `R20`, `R18`) siguen siendo de
> `research.md` (o de F1 cuando se dice).

| Riesgo | Impacto | Mitigación |
|---|---|---|
| **RG1 — Redis como nueva pieza operativa** (D-A7 confirmada el 2026-10-04 con sesión en Redis) | Si Redis cae o no arranca, nadie puede entrar ni mantener la sesión (y se pierden los contadores de intentos) | Servicio `redis` con `healthcheck` y `depends_on: service_healthy` en `docker-compose.yml`; `platform/config` valida `REDIS_URL` al arrancar (falla con mensaje claro, §VII); los errores de Redis se traducen a `503`/`500` con sobre uniforme y log con `request_id`. El plan B (sesión y contadores en PostgreSQL) está documentado en `research.md` R1/R5 por si Redis no fuera viable |
| **RG2 — `ux.md` por alinear con los datos de la cuenta y con la auditoría** (`disenador-ux`) | Las tareas de frontend podrían reproducir los campos antiguos (`nombre`/`correo`) en vez de **nombre, apellidos, correo y teléfono**, o no definir la sección de registro | `ux.md` está entregado pero hay que actualizarlo con los campos confirmados el 2026-10-04 (R20) **y** con la sección de auditoría del cambio de alcance (dos historiales, filtros por cuenta y fechas, paginación, último acceso en la ficha; P20–P22); las tareas `[frontend]` de `tasks.md` quedarán marcadas como dependientes de ese ajuste. El plan define rutas, guards, DTOs y contenido mínimo (P18 + contrato) para que el diseño no cambie la API |
| **RG3 — Datos de la cuenta (cerrado el 2026-10-04)** | — | La spec (ya actualizada) y el humano fijaron **nombre, apellidos, correo y teléfono** como datos de la cuenta, todos obligatorios (`research.md` R20): `users.first_name`/`last_name`/`phone` y los DTOs del contrato. Sin documento de identidad. Si algún día hiciera falta, sería una columna nueva y un campo más — cambio menor y localizado |
| **RG4 — Enumeración de cuentas** | US1 esc. 3 pide un mensaje de "acceso desactivado" y el bloqueo pide el suyo, ambos distintos del genérico | Es exigido por la spec, así que se implementa **tal cual** pero se limita la superficie: credenciales incorrectas y correos inexistentes comparten mensaje y tiempo de respuesta (P6/R18) y el mensaje de bloqueo es idéntico exista o no la cuenta. Queda registrado para `seguridad` |
| **RG5 — `rate-limit` en memoria** | Con varias instancias el umbral se multiplica | Limitación declarada (P17): hoy hay una instancia; si llega el escalado, se decide un almacén compartido en esa funcionalidad (Redis ya está disponible por P1, aunque el umbral por IP no necesita compartirse hoy) |
| **RG6 — bcrypt limita a 72 bytes** | Contraseñas de 64 caracteres con caracteres *multi-byte* | La política fija **entre 8 y 64 caracteres** (tope `maxLength: 64` del contrato, el único límite que ve la persona); la validación rechaza además lo que exceda 72 bytes (límite técnico de bcrypt, solo alcanzable con *multi-byte*) con mensaje claro (P5) |
| **RG7 — Escritura de `lastSeenAt` por petición** | Ruido de escritura (ahora en Redis) | Estrangulado a una actualización por minuto por sesión (P9); se revisa si el volumen crece |
| **RG8 — Carreras** (anti-bloqueo, inicialización, duplicados simultáneos) | Estados prohibidos o duplicados | Advisory lock + transacción (P7/P8) y `UNIQUE` en la base (P13); pruebas de integración **concurrentes** que lo demuestran |
| **RG9 — Deriva de artefactos generados** (sqlc, `schema.d.ts`) | Compilación contra SQL o tipos viejos | Heredado de F1 (R4): código generado commiteado, `make sqlc-verify`, regla de revisión (un PR que toca `migrations/`/`queries/` debe tocar `internal/db/`) |
| **RG10 — Latencia por resolver identidad en cada petición** | Panel lento con muchas peticiones | 1 lectura Redis (sesión) + 1 consulta SQL (identidad y permisos) por petición en un panel de decenas de usuarios; si se nota, caché corta con invalidación por evento — decisión futura, no de F2 (rompería FR-018 si se hace mal) |
| **RG11 — Cookie `Secure` en local (http)** | Sesión que "no se guarda" en desarrollo | `SESSION_COOKIE_SECURE=false` por defecto en desarrollo; `platform/config` **falla al arrancar** si `APP_ENV=production` y está en `false` (P1) |
| **RG12 — Dependencias nuevas** (2 backend + 3 frontend) | Superficie de vulnerabilidades | Justificadas en R16; `govulncheck` y `npm audit` en `make ci` sin altas/críticas (§IV) |
| **RG13 — `BOOTSTRAP_TOKEN` puede parecerle excesivo al humano** | FR-007 "impedir cualquier uso abusivo" | Va **en la puerta de aprobación**: si se rechaza, se retira una única validación del endpoint (queda el guard de "solo con `users` vacío" + rate-limit) y se anota en la spec/plan como decisión humana |
| **RG14 — Denegación de servicio contra bcrypt** (login público) | CPU agotada por intentos masivos | `rate-limit` por IP (P17) + bloqueo por cuenta (P6) + `govulncheck`; registrado para `seguridad` |
| **RG15 — Redis en las pruebas sin poder tocar el CI del kit** | Si el runner no tiene Docker o falla la descarga de imágenes (`redis:7-alpine`), la integración falla y **no** se puede remediar desde `ci.yml` (archivo del kit, `.kit-manifest.json`) | Las pruebas de integración levantan Redis (y, si no hay `DATABASE_URL_TEST`, PostgreSQL) con `testcontainers-go` dentro del propio test (`research.md` R19); los runners `ubuntu-latest` tienen Docker y el CI del kit ya ejecuta `go test -tags=integration ./...` sin cambios. Imágenes pequeñas y reutilización entre tests. Si aun así fuera inviable, el plan B es la sesión/contadores en PostgreSQL (R1) y se comunicaría al humano para proponer el cambio en el repositorio del kit |
| **RG16 — Redis sin persistencia (estado efímero; CONFIRMADO por el humano el 2026-10-04)** | Un reinicio de Redis borra sesiones (hay que volver a entrar) y contadores de intentos (se reabre la ventana de prueba de contraseñas) | Decisión tomada y documentada (`research.md` R21, P23): el servicio `redis` corre **sin `appendonly`, sin `save` y sin volumen**. Un tercero no puede forzar el reinicio y las sesiones no guardan datos que no estén en PostgreSQL. **La auditoría no se ve afectada**: `login_events`/`admin_actions` viven en PostgreSQL (quickstart §10 lo comprueba). Si algún día se exige dureza de sesión, se reabre la decisión (AOF/RDB) |
| **RG17 — Crecimiento de las tablas de auditoría y retención** *(auditoría)* | `login_events`/`admin_actions` crecen sin límite (cada intento y cada acción deja fila) y la spec deja **fuera de alcance** toda purga, retención y exportación | Volumen real de este panel: decenas de filas al día, riesgo de medio plazo; mitigación técnica: índices por fecha y cuenta, paginación con topes (§8.1.3) y consultas acotadas. La política de retención/purga queda como **ampliación futura** ya anotada en Out of Scope (un `DELETE` por fecha en una tarea programada); no se hace en F2 por decisión de la spec. Registrado para que `revisor-codigo`/`seguridad` lo tengan presente |
| **RG18 — Datos personales en el registro (correos e IPs)** *(auditoría)* | La auditoría concentra PII y, en malas manos, es un listado de quién usa el panel y desde dónde | **Mínimo dato necesario** (FR-026, constitución §IV): sin contraseñas ni credenciales, sin *user-agent*, sin el correo de los intentos no identificados; consulta restringida al permiso `admin_usuarios_roles` (FR-024) verificado **en servidor**; registro de **solo lectura** sin exportación (FR-025, Out of Scope). `seguridad` audita el acceso a estas rutas y los DTOs de salida |
| **RG19 — El registro se escribe en tres puntos (P20)** *(auditoría)* | Un punto de escritura olvidado (p. ej. una operación nueva de F3–F9) dejaría un intento sin registrar; y una caída de escritura de la BD puede bloquear un login (fail-closed) | Los tres puntos están enumerados en P20 y cada uno lleva su prueba (service, handler, `authz`); `tasks.md` exige la etiqueta de acción de toda operación sensible y `revisor-codigo` lo comprueba al revisar rutas nuevas. Fail-closed **solo** cuando la operación iba a completarse (misma transacción: o ambos o ninguno); el registro de un intento que ya falla es *best-effort* con error en el log (`request_id`), de modo que un fallo del registro nunca cambia la respuesta que ve la persona |

## Complexity Tracking

> Sin violaciones de la constitución que justificar. Desviaciones de **convenciones internas**,
> declaradas y aceptadas en este plan:

1. **Archivos del dominio por responsabilidad** (P15): la receta de `arquitectura.md` §8 fija
   `model.go`/`repository.go`/`service.go`/`handler.go`/`routes.go`; F2 usa esos nombres y añade
   sufijos (`service_auth.go`, `handler_users.go`…) dentro del **mismo paquete**. Motivo: 18
   operaciones en un solo archivo violan "funciones cortas y con una sola responsabilidad" (§V) en
   forma de archivos ilegibles. Ninguna regla de capas ni de dependencias cambia. Si
   `revisor-codigo` prefiere los cinco archivos literales, es un renombrado sin consecuencias.
2. **`platform/session` define `Identity`/`Resolver`** (P1): tipo de plumbing (igual que
   `Registrar`, excepción declarada de R3): la implementa el dominio `usuarios`, de modo que
   `platform` sigue sin conocer dominios (R1).
3. **El guard de cambio de contraseña se añade a la cadena de grupos** `[authn → authz → CSRF]`
   de `arquitectura.md` §6 (queda como `[authn → guard → authz → CSRF]`); el `documentador`
   actualizará §6 al cerrar la funcionalidad.
4. **`platform/audit` define `Event`/`Action` y la interfaz `Recorder`** *(auditoría, P20)*: mismo
   tipo de excepción declarada que el punto 2 (plumbing que consume el middleware `authz` para
   registrar denegaciones sin conocer dominios); la implementa el dominio `usuarios`, con lo que
   `platform` sigue sin conocer dominios (R1). Además, las tablas de auditoría son **de solo
   inserción** y conservan `created_at`/`updated_at` por convención de esquema, aunque
   `updated_at` nunca cambie (se documenta en `data-model.md`).
