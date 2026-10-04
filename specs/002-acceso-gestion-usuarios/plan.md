# Implementation Plan: Acceso y gestión de usuarios (F2)

**Branch**: `002-acceso-gestion-usuarios` | **Date**: 2026-10-04 | **Spec**: [spec.md](./spec.md) (APROBADA por el humano el 2026-10-04)

**Input**: `specs/002-acceso-gestion-usuarios/spec.md` (fuente de verdad) · `.specify/memory/constitution.md` · `docs/tecnico/decisiones.md` (D-A1…D-A9) y `docs/tecnico/arquitectura.md` (referencia obligada, §8.1 incluida) · plataforma ya construida en F1 (`specs/001-estructura-base/plan.md`, `research.md`, `data-model.md`, `contracts/openapi.yaml`) · skills `go-backend`, `postgres-db`, `react-frontend`.

> ⚠️ **PUERTA DE APROBACIÓN — D-A7 (pendiente de confirmación humana).** Este plan aterriza la
> propuesta D-A7 (sesión en servidor + cookie `httpOnly`, hash bcrypt, permisos por módulo en
> tablas propias) en la decisión **P1** y en `research.md` R1–R4: sesión en **tabla PostgreSQL**
> `sessions`, cookie `ss_session` (`HttpOnly`, `SameSite=Lax`, `Secure` configurable), CSRF
> *double-submit* firmado con `SESSION_SECRET`, hash de contraseñas **bcrypt (cost 12)**. **No está
> aprobada**: se presenta junto con este plan para su confirmación. **No se escribe código de F2
> hasta que el humano apruebe el plan y confirme D-A7.**

## Summary

F2 entrega la puerta del panel de administración y su gestión de accesos: login/logout con sesión
en servidor, inicialización única del administrador (con regla anti-bloqueo), gestión de cuentas
(crear, editar, activar/desactivar, restablecer contraseña — **nunca eliminar**) y gestión de roles
con permisos por módulo (un rol por cuenta, catálogo de permisos fijo = módulos del producto).
Técnicamente es la funcionalidad que **activa los diferidos de F1**: `platform/validate`,
`platform/paginate`, middleware `authn`/`authz`/`CSRF`/`rate-limit`, primeras tablas de negocio
(`users`, `sessions`, `roles`, `permissions`, `role_permissions`, `login_attempts`), primeras
consultas sqlc, primeras dependencias nuevas justificadas (`golang.org/x/crypto`, `github.com/google/uuid`
en backend; `react-hook-form`, `zod`, `@hookform/resolvers` en frontend) y el primer uso de la
superficie pública escribible (con lo que se implementa el `rate-limit` diferido con su punto de
decisión). Todo se construye sobre lo ya hecho: `internal/platform/` (config, logger, apperr,
httpserver con `Registrar`/`Group`/`WriteJSON`/`WriteError`, middleware request-id/recover/logging/CORS,
database con `WithTx`), sqlc sobre `backend/migrations/`, DI manual en `main.go`, sobres uniformes
y `GET /healthz` intacto.

## Technical Context

**Language/Version**: Go 1.27 (D-A5, sin cambios) · TypeScript 5.x `strict` sobre Node 22 · React 19

**Primary Dependencies (nuevas, justificadas en `research.md` R16)**: backend runtime →
`golang.org/x/crypto` (bcrypt; §IV exige bcrypt o argon2 — no hay forma de cumplirlo sin esta
dependencia) y `github.com/google/uuid` (tipos UUID del dominio; §8.1.1 ya lo exige y justifica).
frontend → `react-hook-form` + `zod` + `@hookform/resolvers` (convención de la skill para
formularios). El resto se mantiene: `pgx/v5` único runtime previo; `sqlc`, `golang-migrate`,
`openapi-typescript` como herramientas de desarrollo.

**Storage**: PostgreSQL 16 (servicio `db`). Primeras tablas de negocio (F2 arranca en la migración
`000002`; ver `data-model.md`). Capa de datos sqlc (D-A3): consultas en
`backend/internal/db/queries/`, código generado commiteado.

**Testing**: estrategia por capa de F1 ampliada con las pruebas de sesión/autorización (sección
propia): `go test` (unitarias), `go test -tags=integration` (repositorio contra PostgreSQL real,
incluida la carrera anti-bloqueo), Vitest + Testing Library + MSW, Playwright (e2e del flujo de
acceso y gestión).

**Target Platform**: contenedores Docker en local (Docker Compose) + GitHub Actions (CI del kit,
sin tocar). Sin despliegue a producción (fuera de alcance de F2).

**Performance Goals**: login responde en <2 s con BD sana · el listado de usuarios (≤100 filas) en
<1 s · la resolución de identidad por petición (sesión + permisos) añade ≤2 consultas por petición
de panel, aceptable para el volumen del equipo de la iglesia (decenas de cuentas).

**Constraints**: stack fijo (constitución) · capas `handler → service → repository` (§II) ·
contrato OpenAPI antes que el código (§II, R8) · migraciones versionadas e inmutables (§VI) ·
authn/authz siempre en servidor (§IV, CWE-862) · contraseñas solo bcrypt/argon2 (§IV, CWE-256) ·
archivos del kit no editables (regla 10 de AGENTS.md) · dependencias minimizadas y justificadas
(D-A8) · **la spec no se reabre**: los huecos se marcan como decisión o pregunta (R3, R15).

**Scale/Scope**: 16 operaciones REST (1 dominio backend `usuarios`, 3 features frontend), 6 tablas,
4 migraciones, 6 middlewares (4 nuevos + CORS ampliado), 1 puerta de aprobación (D-A7).

## Constitution Check

*GATE: debe pasar antes de la investigación y re-verificarse tras el diseño.*

| Principio | Evaluación | Resultado |
|---|---|---|
| §I La spec manda | El plan implementa FR-001…FR-020 sin interpretarlos; cada uno tiene fila en "Cobertura de requisitos". Los dos puntos que la spec deja abiertos están **marcados**, no supuestos: el detalle de D-A7 (que la propia spec declara pendiente, ver *Assumptions*) y el significado de "identificación" (R3, pregunta al humano). Ningún requisito fuera de alcance se construye (sin auto-servicio por correo, sin eliminación de cuentas, sin registro público). | ✅ |
| §II Arquitectura | Monorepo; capas `handler → service → repository` sobre `internal/platform/` (reglas R1–R8 de `arquitectura.md`); API REST JSON documentada en `backend/api/openapi.yaml`, con el delta redactado en esta fase **antes** que el código (`contracts/openapi.yaml`, §8.1.5). Dependencias nuevas: 2 en backend + 3 en frontend, cada una justificada (R16). | ✅ |
| §III Pruebas | Toda tarea lleva sus pruebas (se exigirá en `tasks.md`); cobertura ≥80 % en `service/` (verificable con `go test -cover`); repositorio contra PostgreSQL real con `//go:build integration` (incluida prueba de concurrencia del anti-bloqueo); frontend con Vitest + Testing Library + MSW; flujos críticos con Playwright (local). | ✅ |
| §IV Seguridad | Contraseñas con **bcrypt cost 12** (CWE-256), nunca en texto plano ni en respuestas (FR-003); SQL solo parametrizado vía sqlc (CWE-89); validación de toda entrada en backend (`platform/validate`, CWE-20) aunque el frontend valide; authn/authz **en servidor** en cada operación de panel (CWE-862); secretos (`SESSION_SECRET`, `BOOTSTRAP_TOKEN`) por variables de entorno, `.env.example` sin valores reales; CSRF en todo método inseguro con sesión (P10); sin `dangerouslySetInnerHTML` ni tokens en `localStorage` (skill). `govulncheck`/`npm audit` en el CI del kit. | ✅ |
| §V Calidad | `gofmt`/`go vet`/`golangci-lint` sin errores; TS `strict` sin `any`; errores envueltos con `%w` y traducidos solo en `WriteError`; nombres en inglés dentro del código. | ✅ |
| §VI Base de datos | 4 migraciones versionadas con `up`/`down` completos, numéricas desde `000002`; nunca se edita una aplicada; tablas con `id`, `created_at`, `updated_at`; FK e índices explícitos; `CHECK`/`UNIQUE` en la base (ver `data-model.md`). | ✅ |
| §VII Observabilidad | Logs `log/slog` con `request_id` (incluidos los intentos de acceso fallidos, sin credenciales); `/healthz` intacto y sin sesión (sigue siendo el chequeo operativo); config por variables de entorno (`platform/config` ampliado); todo levantable con `make up`. | ✅ |
| §VIII Gobierno | Este plan **y** la confirmación de D-A7 requieren aprobación humana antes de `/speckit.tasks` y de cualquier código. Sin despliegue a producción (fuera de alcance). Quien escribe no aprueba: todo pasa por `qa-tester`, `revisor-codigo` y `seguridad`. | ✅ |

**Sin violaciones que justificar** → "Complexity Tracking" solo registra desviaciones de
convenciones internas (receta de archivos), no de la constitución.

## Decisiones técnicas (resumen; análisis completo en `research.md`)

Numeración **P1…P19** propia de este plan (no confundir con D1–D23 de F1 ni D-A1…D-A9 de
`docs/tecnico/decisiones.md`).

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| **P1** | **D-A7 aterrizada (PENDIENTE DE CONFIRMACIÓN HUMANA)**: sesión en servidor en la **tabla PostgreSQL `sessions`**; cookie `ss_session` con `HttpOnly`, `SameSite=Lax`, `Secure` configurable (`SESSION_COOKIE_SECURE`), `Path=/`; valor = token aleatorio de 32 bytes (`crypto/rand`) y en BD solo su **SHA-256**; permisos por módulo en `permissions`/`role_permissions`; revocación = borrar filas de `sessions` + comprobación por petición de que la cuenta sigue activa | Revocación inmediata al desactivar (FR-012), cero servicios nuevos, `WithTx`/sqlc ya existen, escala de sobra para un equipo de decenas de personas | Redis/memcached (servicio + dependencia para este volumen); JWT autocontenido (revocación forzada con lista negra); cookie firmada sin estado en servidor (sin revocación real); token en `localStorage` (prohibido, CWE-79) |
| P2 | Router: **se mantiene `net/http` tras `Registrar`** (cierra la pregunta 3 de `decisiones.md`/D-A4) | Los grupos con permisos (`Group("/api/v1/admin", authn, authz, CSRF)`) ya se expresan con `Group`; chi solo aportaría sintaxis y cuesta dependencia + adaptador | Adoptar chi ahora (se reevalúa si el ruteo se complica; el cambio sigue limitado a `platform/httpserver`) |
| P3 | `platform/validate` **propio mínimo** (cierra la pregunta 4 de `decisiones.md`): reflexión sobre las etiquetas `validate` de los DTOs (`required`, `omitempty`, `min`, `max`, `email`, `oneof`) → `apperr.Invalid` con `details` por campo | Sin dependencias (D-A8); mensajes en español y por campo; los 5 tags que usamos no justifican una librería | `go-playground/validator` (dependencia + transitivas, mensajes genéricos en inglés, superficie enorme para nuestro uso) |
| P4 | Tipos UUID: `github.com/google/uuid` en el dominio, `pgtype.UUID` solo en `repository.go` (cierra la pregunta 5; §8.1.1) | Mantiene `pgx` fuera de las capas altas (R4); solo tipos, sin transitivas | `pgtype.UUID` en el dominio (arrastra `pgx`); `override` de sqlc (una segunda regla para un tipo) |
| P5 | Hash de contraseñas: **bcrypt cost 12** (`golang.org/x/crypto/bcrypt`); política FR-010 centralizada en `platform/password` y aplicada en los 3 flujos (creación, restablecimiento, cambio propio) | Cumplimiento literal de §IV; bcrypt es el estándar con más revisión para este caso; el hash es autodescriptivo (futura migración a argon2id sin romper nada) | argon2id (más resistente a GPU pero más parámetros que afinar y la misma dependencia; decisión reversible); pbkdf2 de la stdlib (no es ni bcrypt ni argon2, §IV lo exige) |
| P6 | Bloqueo FR-006 en tabla `login_attempts` **por identificador normalizado (correo) exista o no la cuenta**: 5 intentos → 15 min (constantes de código, valores confirmados por el humano) | Concilia FR-003 y FR-006: el mensaje de bloqueo es idéntico para cuentas reales e inexistentes y por tanto **no** revela existencia | Contadores solo en `users` (el bloqueo solo ocurriría en cuentas reales → enumera); contadores en memoria (se pierden al reiniciar); rate-limit por IP como única medida (no frena la prueba masiva contra una cuenta) |
| P7 | Regla anti-bloqueo (FR-008) verificada **dentro de la transacción** de la mutación, con `pg_advisory_xact_lock` sobre una clave fija + recuento **post-mutación** de cuentas activas con el permiso `admin_usuarios_roles`; 0 → `409 conflict` y rollback | Hace imposible el estado prohibido incluso con dos administradores actuando a la vez (edge case de la spec) | Comprobación sin lock (carrera: dos desactivaciones simultáneas dejan 0 administradores); trigger en la BD (lógica de negocio fuera del service, difícil de probar) |
| P8 | Inicialización única (FR-007): `POST /api/v1/setup/initialize` solo con `users` vacío, en transacción con el mismo advisory lock, que crea el rol **"Administrador"** con los 9 permisos y su cuenta; **exige el token `BOOTSTRAP_TOKEN`** (cabecera `X-Setup-Token`) y va rate-limited | El guard de BD impide repetirla (FR-007) y el token impide que un tercero se declare administrador en una instalación recién desplegada ("impedir cualquier uso abusivo", FR-007) | Sin token (ventana de robo del primer administrador en instalaciones públicas); CLI embebida (exige acceso al servidor y no es el producto); token de un solo uso impreso en logs (operación rara y secreto efímero mal resguardado) |
| P9 | Sesión: expiración por inactividad de **30 minutos** (assumption de la spec) medida con `last_seen_at` (escritura estrangulada a 1/min) y **vida absoluta de 12 horas** *(propuesta nueva, se confirma junto con D-A7)*; al desactivar una cuenta o restablecer su contraseña se borran sus sesiones; `authn` comprueba **por petición** que la cuenta sigue activa | FR-005/FR-012/FR-018: los cambios de rol, permisos y estado se reflejan desde la primera acción posterior sin tocar cookies | Expiración solo por inactividad (una cookie robada vive para siempre); refresco de permisos en caché (rompería FR-018/SC-009) |
| P10 | CSRF obligatorio en todo método no seguro con sesión: *double-submit* **firmado** — cookie `csrf_token` (no `HttpOnly`) = `nonce.HMAC-SHA256(SESSION_SECRET, nonce)` + cabecera `X-CSRF-Token` igual, verificado por `middleware.CSRF` en el grupo (cadena de `arquitectura.md` §6) | Sin estado extra ni consultas; el HMAC impide que un atacante que inyecta cookies fabrique un par válido; `SESSION_SECRET` (ya en `.env.example`) por fin tiene uso | Token guardado en la sesión (doble consulta); exigir solo cabecera personalizada (débil); fiarlo todo a `SameSite` (no cubre toda la superficie); frameworks de CSRF (dependencia) |
| P11 | CORS se **amplía sin dependencia** (revisión de D16 cerrada): `Access-Control-Allow-Credentials: true`, eco exacto del `Origin` permitido (nunca `*` con credenciales), cabeceras `Content-Type, X-CSRF-Token, X-Request-ID`, `Vary: Origin` | Las cookies de sesión exigen credenciales; 40 líneas bastan y el middleware ya existe | `rs/cors` (dependencia para lo que ya está escrito) |
| P12 | Modelo: **un rol por cuenta** → `users.role_id` (se **simplifica** el `user_roles` orientativo de F1); catálogo de permisos fijo en tabla `permissions` **sembrada por la migración** + `role_permissions`; sesiones en `sessions`; FR-006 en `login_attempts` (detalle en `data-model.md`) | Refleja literalmente Q4 (un rol), da FK y unicidad reales y permite al panel listar el catálogo con `GET /admin/permisos` | `user_roles` (contradice Q4); permisos como `TEXT[]` en `roles` (sin FK, difícil de validar); permisos solo como constantes de código (el panel no podría listarlos con etiqueta sin duplicar el catálogo) |
| P13 | Normalización (Q5): el correo se guarda `trim`+minúsculas y el nombre de rol `trim`+colapso de espacios (conservando sus mayúsculas de presentación), con `UNIQUE` real y `UNIQUE (lower(name))`; duplicados → `409 conflict` | El mismo dato nunca vive en dos formas y la unicidad la garantiza la BD, no solo el código | Índice funcional sobre el valor crudo (permite guardar duplicados "casi"); comparar solo en el service (carrera entre dos creaciones simultáneas) |
| P14 | `platform/paginate` (diferido de F1) para `GET /admin/usuarios` y `GET /admin/roles`: `limit` 20 por defecto, tope 100, sobre `{items, total, limit, offset}` (§8.1.2/§8.1.3) | Primer listado de panel con parámetros de usuario; convención ya cerrada | Listados sin acotar (rompe §8.1.3); paginación por cursor (innecesaria en listados de decenas de filas) |
| P15 | Un **único dominio `internal/usuarios/`** para todo F2 + `platform/session` (token y cookie, **sin SQL**) + `platform/password`; los archivos de la receta se **dividen por responsabilidad** (`service_auth.go`, `handler_users.go`…) — desviación declarada en "Complexity Tracking" | Un dominio evita que dos paquetes consulten las mismas tablas (R2) y encaja con "transversal en `platform/` + dominio `usuarios`" (F1 `data-model.md`); los archivos por responsabilidad mantienen funciones cortas | Dominios `auth/` + `usuarios/` separados (obliga a consultas compartidas y a una interfaz extra para revocar sesiones); un `handler.go` monolítico de 16 endpoints |
| P16 | `apperr` crece con los kinds ya registrados en el contrato: `Invalid` (400), `Unauthenticated` (401), `Forbidden` (403), `Conflict` (409), `RateLimited` (429, con `Retry-After`) | El registro estaba cerrado y previsto para "crecer bajo demanda con F2+" (`arquitectura.md` §5.11) | Códigos nuevos fuera del registro (rompería el contrato) |
| P17 | `rate-limit` mínimo **en memoria** (cierra el punto de decisión de D23): ventana deslizante por IP sobre los endpoints públicos escribibles (`/auth/login`, `/setup/initialize`), 429 `rate_limited` | El primer endpoint público escribible ha llegado con F2; amortigua la prueba masiva de contraseñas por IP además del bloqueo por cuenta | No hacerlo (deja el diferido sin decidir); Redis/limiter externo (dependencia y servicio para una sola instancia) |
| P18 | Frontend: rutas `/login`, `/cambiar-contrasena`, `/panel`, `/panel/usuarios`, `/panel/roles`; guards `RequireAuth`, `RequirePermission` y `RequirePasswordChange`; sesión vía TanStack Query (`GET /auth/session`); formularios con React Hook Form + Zod; menú filtrado por permisos (US4 esc. 6) | Es la convención de la skill y el patrón que copiarán F3–F9; los permisos se ocultan **y** se deniegan en servidor | Estado global propio (redux/zustand: no hace falta); fetch en componentes (prohibido); validación manual de formularios |
| P19 | Contrato: delta en `specs/002-acceso-gestion-usuarios/contracts/openapi.yaml`, fusionado en `backend/api/openapi.yaml` al implementar (§8.1.5) con `info.version` **0.2.0 → 0.3.0**; esquema de seguridad `sessionCookie`; el registro de `error.code` **no cambia** (F2 emite `invalid`, `unauthenticated`, `forbidden`, `conflict`, `rate_limited`, ya en el enum) | Contrato antes que el código (§II) sin dos copias vivas | Editar el snapshot de `specs/` a posteriori (prohibido); códigos nuevos ad-hoc por endpoint |

## Project Structure

### Documentation (esta funcionalidad)

```text
specs/002-acceso-gestion-usuarios/
├── plan.md              # EDITABLE (este archivo, /speckit.plan)
├── research.md          # EDITABLE (fase 0: R1…R18, incluye D-A7)
├── data-model.md        # EDITABLE (fase 1: tablas, migraciones, consultas)
├── quickstart.md        # EDITABLE (fase 1: cómo probar F2 en local)
├── spec.md              # APROBADA — no se edita
├── ux.md                # DEL disenador-ux (PENDIENTE — riesgo R2; las tareas de frontend lo necesitan)
├── contracts/
│   └── openapi.yaml     # EDITABLE (fase 1): delta de diseño, se fusiona en backend/api/openapi.yaml
├── checklists/          # (existente)
└── tasks.md             # Fase 2 (/speckit.tasks — NO lo crea este comando)
```

### Source Code — backend

```text
backend/
├── cmd/api/
│   ├── main.go                      # EDITADO: DI de F2 + grupos /api/v1/auth, /api/v1/setup y /api/v1/admin
│   └── main_test.go                 # EDITADO: prueba de humo ampliada (rutas nuevas + /healthz intacto)
├── internal/
│   ├── db/
│   │   ├── queries/
│   │   │   ├── users.sql            # NUEVO: cuentas (CRUD + authn + recuento anti-bloqueo)
│   │   │   ├── roles.sql            # NUEVO: roles + role_permissions
│   │   │   ├── permissions.sql      # NUEVO: catálogo
│   │   │   ├── sessions.sql         # NUEVO: sesiones
│   │   │   └── login_attempts.sql   # NUEVO: intentos fallidos (upsert/limpieza)
│   │   └── (generado por sqlc, commiteado)
│   ├── usuarios/                    # NUEVO — dominio F2 (área completa: acceso + gestión)
│   │   ├── model.go                 # entidades, estados y DTOs (etiquetas validate)
│   │   ├── repository.go            # constructor + mapRow/params (pgtype → dominio)
│   │   ├── repository_users.go      # cuentas + sesiones + intentos
│   │   ├── repository_roles.go      # roles + permisos + guard anti-bloqueo
│   │   ├── service.go               # tipos comunes + invariantes compartidos
│   │   ├── service_auth.go          # login, logout, sesión, cambio de contraseña, inicialización
│   │   ├── service_users.go         # CRUD de cuentas, activar/desactivar, restablecer
│   │   ├── service_roles.go         # CRUD de roles y catálogo de permisos
│   │   ├── handler.go               # constructor + helpers HTTP
│   │   ├── handler_auth.go          # /api/v1/auth/* + /api/v1/setup/*
│   │   ├── handler_users.go         # /api/v1/admin/usuarios*
│   │   ├── handler_roles.go         # /api/v1/admin/roles* + /api/v1/admin/permisos
│   │   ├── routes.go                # RegisterPublic (login, setup) + RegisterAdmin (usuarios/roles)
│   │   └── *_test.go                # service con fakes · handler con httptest · repository integration
│   └── platform/
│       ├── apperr/apperr.go         # EDITADO: + Invalid, Unauthenticated, Forbidden, Conflict, RateLimited
│       ├── config/config.go         # EDITADO: + SESSION_SECRET, SESSION_* , SESSION_COOKIE_SECURE, BOOTSTRAP_TOKEN
│       ├── session/                 # NUEVO: token aleatorio + SHA-256, cookie ss_session/csrf_token, tipos Identity y Resolver (sin SQL)
│       ├── password/                # NUEVO: bcrypt (hash/verify) + política FR-010
│       ├── validate/                # NUEVO (P3): validación de DTOs por etiquetas → apperr.Invalid
│       ├── paginate/                # NUEVO (P14): limit/offset con topes
│       └── middleware/
│           ├── authn.go             # NUEVO: cookie → session.Resolver → Identity en el contexto
│           ├── authz.go             # NUEVO: AuthzByModule(código) sobre la Identity
│           ├── csrf.go              # NUEVO: double-submit firmado en métodos no seguros
│           ├── ratelimit.go         # NUEVO (P17): ventana deslizante por IP
│           ├── cors.go              # EDITADO (P11): credenciales + cabeceras nuevas
│           └── chain.go             # EDITADO si hace falta: orden de los grupos
├── migrations/
│   ├── 000002_create_roles_and_permissions.up.sql / .down.sql   # NUEVO (incluye la siembra del catálogo)
│   ├── 000003_create_users.up.sql / .down.sql                   # NUEVO
│   ├── 000004_create_sessions.up.sql / .down.sql                # NUEVO
│   └── 000005_create_login_attempts.up.sql / .down.sql          # NUEVO
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
│   └── schema.d.ts                  # REGENERADO (npm run api:gen)
├── app/
│   ├── router.tsx                   # EDITADO: /login, /cambiar-contrasena, /panel/* con guards
│   ├── guards.tsx                   # NUEVO: RequireAuth, RequirePermission, RequirePasswordChange
│   └── layout.tsx                   # EDITADO: layout del panel con navegación filtrada por permisos
├── features/
│   ├── auth/                        # NUEVO: LoginPage, ChangePasswordPage, hooks useSession/useLogin/…
│   ├── usuarios/                    # NUEVO: UsersPage, UserForm, hooks useUsers/useCreateUser/…
│   └── roles/                       # NUEVO: RolesPage, RoleForm (permisos), hooks useRoles/…
├── components/                      # EDITADO: ConfirmDialog, Field, EmptyState (reutilizables)
└── lib/                             # EDITADO: helpers de permisos y formato
frontend/e2e/acceso.spec.ts          # NUEVO: flujo completo Playwright (ver quickstart §10)

# Raíz
.env.example                         # EDITADO: + BOOTSTRAP_TOKEN, SESSION_SECRET (ya existe), SESSION_* y su documentación
README.md                            # EDITADO por documentador al cerrar: comandos y variables nuevas
docker-compose.yml                   # SIN CAMBIOS de estructura (solo variables nuevas opcionales si hicieran falta)
Makefile / .github/workflows/ci.yml / .githooks/   # SIN CAMBIOS (son del kit)
```

## Cadena de middleware y grupos (orden; el primero es el más externo)

```text
Global:   request-id → recover → logging → CORS(con credenciales, P11) → handler
Grupos:   /api/v1/setup                    → rate-limit (P17)                       → handler
          /api/v1/auth  (público)  POST /login → rate-limit (P17)                   → handler
          /api/v1/auth  (sesión)            → authn → CSRF                          → handler
          /api/v1/admin                     → authn → guard de cambio de contraseña → CSRF → (subgrupo) authz(módulo) → handler
```

- `authn` (P1/P9): lee `ss_session`, resuelve la identidad vía `session.Resolver` (interfaz
  definida en `platform/session`, implementada por `usuarios.Service` — R3), comprueba que la
  cuenta siga activa y la deja en el contexto. Sin sesión válida → `401 unauthenticated`.
- **Guard de cambio de contraseña**: si `mustChangePassword`, solo se autorizan `/auth/session`,
  `/auth/logout` y `/auth/password`; el resto → `403 forbidden` con `details.reason =
  "password_change_required"` (la UI redirige al formulario).
- `authz(módulo)` (P16/FR-016): exige el permiso del módulo (p. ej. `admin_usuarios_roles`) en la
  identidad resuelta; sin permiso → `403 forbidden` con mensaje claro. F3–F9 crearán sus subgrupos
  con su módulo sin tocar este dominio.
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
`go test -tags=integration ./...` · `npm test -- --run` · `make e2e` · `make ci`.

| Capa | Tipo | Qué verifica F2 además de lo de F1 | Dónde |
|---|---|---|---|
| `platform/password` | Unitaria (tabla de casos) | Política FR-010: longitud 8–64 y ≤72 bytes, mayúsculas+minúsculas+números+especiales, distinta del nombre y del correo; hash/verify bcrypt; límite de bytes | `internal/platform/password/*_test.go` |
| `platform/session` | Unitaria | Token aleatorio distinto por llamada, SHA-256 estable, cookies con `HttpOnly`/`SameSite`/`Secure`/`Max-Age`, lectura y borrado | `internal/platform/session/*_test.go` |
| `platform/validate` | Unitaria (tabla de casos) | Cada etiqueta y el `details` por campo que produce | `internal/platform/validate/*_test.go` |
| `middleware` | Unitaria + `httptest` | `authn` (sin cookie, cookie inválida, sesión expirada, cuenta inactiva), `authz` (con/sin permiso), `csrf` (método seguro pasa, inseguro sin/desalineado con token), `ratelimit` (429 tras el umbral) | `internal/platform/middleware/*_test.go` |
| `service` (dominio) | Unitaria con **fakes** | Login (éxito, credenciales genéricas, cuenta inactiva, bloqueo 5/15 min, limpieza de intentos), cambio/restablecimiento de contraseña con `mustChangePassword`, anti-bloqueo (desactivar, cambiar rol, quitar permiso al rol), reglas de roles (≥1 permiso, duplicados normalizados, eliminación solo sin uso), un rol por cuenta | `internal/usuarios/service_*_test.go` |
| `handler` | Unitaria con `httptest` + service falso | Cada operación: decodificación, validación con `details`, códigos 200/201/400/401/403/404/409/429, sobres de éxito y de error, **que la contraseña nunca aparece en la respuesta** | `internal/usuarios/handler_*_test.go` |
| `repository` | **Integración** (PostgreSQL real) | SQL real, `UNIQUE` de correos/nombres normalizados, FK y `ON DELETE`, `upsert` de intentos, expiración de sesiones, **carrera anti-bloqueo** (dos transacciones concurrentes no dejan 0 administradores) e **inicialización única** (dos `initialize` simultáneos → uno solo crea) | `internal/usuarios/repository_*_test.go` (`//go:build integration`) |
| `cmd/api` | Humo de composición | Rutas nuevas publicadas con su sobre y `/healthz` intacto (§8.1.9) | `cmd/api/main_test.go` |
| Frontend | Unitaria (Vitest + MSW) | Login (errores genéricos, bloqueo, cuenta desactivada), guard de cambio de contraseña, navegación filtrada por permisos, formulario de cuenta (duplicado, rol inexistente, política), formulario de rol (sin permisos → error), eliminar rol en uso | `frontend/src/features/*/*.test.tsx` |
| E2E | Playwright (local) | Recorrido completo de `quickstart.md` §10 (valida SC-002, SC-005, SC-006, SC-007, SC-010, SC-011) | `frontend/e2e/acceso.spec.ts` |

**Pruebas de sesión/autorización que no faltan** (las que la spec hace críticas): desactivar una
cuenta con sesión abierta corta el acceso en la primera acción posterior (FR-012); quitar un permiso
a un rol se refleja en la siguiente petición de sus cuentas (FR-018); sin permiso no se puede forzar
la operación por API aunque la UI oculte el botón (FR-016/SC-007); los mensajes de error de login no
revelan existencia (FR-003/SC-008, incluido el caso de cuenta inactiva — ver R4); el bloqueo por
intentos es idéntico para cuentas reales e inexistentes (FR-006).

## Cobertura de requisitos (FR-001…FR-020)

| FR | Dónde se resuelve en este plan | Verificación |
|---|---|---|
| FR-001 | P18 (guards `RequireAuth` en el router) + `middleware/authn` (P1) | e2e + quickstart §2 (SC-001) |
| FR-002 | P1 + P5 + `service_auth.go` (login: correo+contraseña, cuenta inactiva rechazada) | quickstart §2 y §7 · pruebas del service |
| FR-003 | P6 + P20 *(en `research.md` R18: verificación dummy para uniformidad de tiempos)* + contrato (ninguna respuesta contiene `password` ni `passwordHash`) | quickstart §8 · suite de handlers (SC-008) |
| FR-004 | `POST /api/v1/auth/logout` (P1: borra la sesión y la cookie) | quickstart §3 |
| FR-005 | P9 (inactividad 30 min con `last_seen_at`) | quickstart §3 · pruebas de `platform/session` |
| FR-006 | P6 (`login_attempts`, 5 intentos / 15 min como constantes) | quickstart §8 · pruebas del service + integración |
| FR-007 | P8 (`POST /api/v1/setup/initialize` + guard de BD + `BOOTSTRAP_TOKEN`) | quickstart §1 y §8 · integración (doble inicialización) |
| FR-008 | P7 (recuento post-mutación en transacción con advisory lock) | quickstart §6 · integración concurrente (SC-004) |
| FR-009 | P3 + P13 + `POST /api/v1/admin/usuarios` (rol inexistente → 400 con `details.roleId`) | quickstart §4 · pruebas de service/handler |
| FR-010 | P5 (política en `platform/password`) + P9 (`mustChangePassword`) + `POST /admin/usuarios/{id}/password` | quickstart §5 · tabla de casos de la política |
| FR-011 | `PATCH /api/v1/admin/usuarios/{id}` (fullName, email, roleId, isActive) | quickstart §4 y §7 |
| FR-012 | P1 + P9 (borrado de sesiones al desactivar + comprobación por petición) | quickstart §7 · e2e (SC-006) |
| FR-013 | **No existe operación de borrado** de cuentas en el contrato ni en el servicio (solo `is_active`) | revisión del contrato + `revisor-codigo` |
| FR-014 | P12 + P13 + `POST/PATCH /api/v1/admin/roles` (≥1 permiso, nombre único normalizado) | quickstart §4 · pruebas de service |
| FR-015 | P12 (catálogo de 9 permisos sembrado en `000002`) | `data-model.md` · `GET /admin/permisos` |
| FR-016 | `middleware.AuthzByModule` (P1/P16) en cada operación de panel | quickstart §6 · pruebas de middleware (SC-007) |
| FR-017 | P12 (`users.role_id`, un solo rol) + `PATCH/DELETE /api/v1/admin/roles/{id}` (eliminar solo sin uso → 409) | quickstart §4 y §6 |
| FR-018 | P9 (permisos resueltos **por petición**, sin caché) | quickstart §6 · e2e (SC-009) |
| FR-019 | `GET /api/v1/admin/usuarios` (`UserItem`: estado, correo, rol) | quickstart §4 |
| FR-020 | `POST /api/v1/auth/password` (cambio propio con contraseña actual) | quickstart §5 (SC-010) |

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

## Riesgos

| Riesgo | Impacto | Mitigación |
|---|---|---|
| **R1 — D-A7 sin confirmar** | El núcleo de F2 (sesión, CSRF, hash) no puede implementarse | **Bloqueante explícito**: la puerta de aprobación de este plan incluye confirmar D-A7 tal como está aterrizada en P1/P5/P10/P9. Si el humano cambia la opción (p. ej. argon2id o Redis), solo se reescriben P1/P5 y `research.md` R1/R4 antes de `tasks` |
| **R2 — `ux.md` pendiente** (`disenador-ux`) | Las tareas de frontend no tienen pantallas/estados definitivos | Se pide `ux.md` en paralelo a la aprobación; las tareas `[frontend]` de `tasks.md` quedarán marcadas como dependientes de él. El plan define rutas, guards y contenido mínimo (P18) para que el diseño no cambie la API |
| **R3 — "Identificación" de la cuenta es ambigua** en la spec (¿nombre completo o documento de identidad?) | Modelado de `users` y formulario | Se modela como `full_name` (nombre completo) porque la política de contraseñas habla de "nombre" (FR-010). **Pregunta al humano**; si exige además un documento (cédula), es una columna nueva y un campo más de los DTOs — cambio menor y localizado |
| **R4 — Enumeración de cuentas** | US1 esc. 3 pide un mensaje de "acceso desactivado" y el bloqueo pide el suyo, ambos distintos del genérico | Es exigido por la spec, así que se implementa **tal cual** pero se limita la superficie: credenciales incorrectas y correos inexistentes comparten mensaje y tiempo de respuesta (P6/R18) y el mensaje de bloqueo es idéntico exista o no la cuenta. Queda registrado para `seguridad` |
| **R5 — `rate-limit` en memoria** | Con varias instancias el umbral se multiplica | Limitación declarada (P17): hoy hay una instancia; si llega el escalado, se decide un almacén compartido en esa funcionalidad |
| **R6 — bcrypt limita a 72 bytes** | Contraseñas muy largas (multi-byte) | La validación exige ≤72 bytes y ≤64 caracteres y lo dice el mensaje (P5) |
| **R7 — Escritura de `last_seen_at` por petición** | Ruido de escritura en el panel | Estrangulado a una actualización por minuto por sesión (P9); se revisa si el volumen crece |
| **R8 — Carreras** (anti-bloqueo, inicialización, duplicados simultáneos) | Estados prohibidos o duplicados | Advisory lock + transacción (P7/P8) y `UNIQUE` en la base (P13); pruebas de integración **concurrentes** que lo demuestran |
| **R9 — Deriva de artefactos generados** (sqlc, `schema.d.ts`) | Compilación contra SQL o tipos viejos | Heredado de F1 (R4): código generado commiteado, `make sqlc-verify`, regla de revisión (un PR que toca `migrations/`/`queries/` debe tocar `internal/db/`) |
| **R10 — Latencia por resolver identidad en cada petición** | Panel lento con muchas peticiones | 2 consultas por petición (sesión + identidad) en un panel de decenas de usuarios; si se nota, caché corta con invalidación por evento — decisión futura, no de F2 (rompería FR-018 si se hace mal) |
| **R11 — Cookie `Secure` en local (http)** | Sesión que "no se guarda" en desarrollo | `SESSION_COOKIE_SECURE=false` por defecto en desarrollo; `platform/config` **falla al arrancar** si `APP_ENV=production` y está en `false` (P1) |
| **R12 — Dependencias nuevas** (2 backend + 3 frontend) | Superficie de vulnerabilidades | Justificadas en R16; `govulncheck` y `npm audit` en `make ci` sin altas/críticas (§IV) |
| **R13 — `BOOTSTRAP_TOKEN` puede parecerle excesivo al humano** | FR-007 "impedir cualquier uso abusivo" | Va **en la puerta de aprobación**: si se rechaza, se retira una única validación del endpoint (queda el guard de "solo con `users` vacío" + rate-limit) y se anota en la spec/plan como decisión humana |
| **R14 — Denegación de servicio contra bcrypt** (login público) | CPU agotada por intentos masivos | `rate-limit` por IP (P17) + bloqueo por cuenta (P6) + `govulncheck`; registrado para `seguridad` |

## Complexity Tracking

> Sin violaciones de la constitución que justificar. Desviaciones de **convenciones internas**,
> declaradas y aceptadas en este plan:

1. **Archivos del dominio por responsabilidad** (P15): la receta de `arquitectura.md` §8 fija
   `model.go`/`repository.go`/`service.go`/`handler.go`/`routes.go`; F2 usa esos nombres y añade
   sufijos (`service_auth.go`, `handler_users.go`…) dentro del **mismo paquete**. Motivo: 16
   operaciones en un solo archivo violan "funciones cortas y con una sola responsabilidad" (§V) en
   forma de archivos ilegibles. Ninguna regla de capas ni de dependencias cambia. Si
   `revisor-codigo` prefiere los cinco archivos literales, es un renombrado sin consecuencias.
2. **`platform/session` define `Identity`/`Resolver`** (P1): tipo de plumbing (igual que
   `Registrar`, excepción declarada de R3): la implementa el dominio `usuarios`, de modo que
   `platform` sigue sin conocer dominios (R1).
3. **El guard de cambio de contraseña se añade a la cadena de grupos** `[authn → authz → CSRF]`
   de `arquitectura.md` §6 (queda como `[authn → guard → authz → CSRF]`); el `documentador`
   actualizará §6 al cerrar la funcionalidad.
