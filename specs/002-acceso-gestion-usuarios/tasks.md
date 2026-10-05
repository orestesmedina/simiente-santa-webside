# Tasks: Acceso y gestión de usuarios (F2)

**Branch**: `002-acceso-gestion-usuarios` | **Date**: 2026-10-04 | **Spec**: [spec.md](./spec.md) (APROBADA, US1–US8 / FR-001…FR-026) · **Plan**: [plan.md](./plan.md) (APROBADO, P1–P23)

**Input**: plan aprobado (`plan.md`, P1…P23) · `research.md` (R1…R23) · `data-model.md` (migraciones
`000002`…`000004`, tablas, claves Redis) · `contracts/openapi.yaml` (delta de diseño, inmutable) ·
`ux.md` (pantallas, textos y estados) · `quickstart.md` (§0…§12, verificación) · constitución
(`.specify/memory/constitution.md`) · skills `go-backend`, `postgres-db`, `react-frontend` ·
plataforma de F1 (`specs/001-estructura-base/tasks.md`) ya integrada.

> **Regla de ejecución**: **un commit por tarea** (Conventional Commits; los hooks del kit lo
> validan). Toda tarea de código **incluye sus pruebas** en el mismo commit (constitución §III):
> backend con `go test` y, donde toca persistencia real, `go test -tags=integration`
> (`//go:build integration`, PostgreSQL real y Redis real levantados con `testcontainers-go` — R19);
> frontend con Vitest + Testing Library + MSW; flujos críticos con Playwright (`make e2e`). Quien
> escribe no aprueba: cada tarea pasa por `qa-tester`, `revisor-codigo` y `seguridad` antes de
> integrarse (flujo de entrega del orquestador). **El CI del kit (`.github/workflows/ci.yml`) no se
> toca**: las pruebas de integración llevan su propia infraestructura (R19/R15).

## Decisiones del plan que estas tareas respetan (sin reabrir)

- **P1/P9**: sesión en **Redis** (cookie `ss_session` `HttpOnly`/`SameSite=Lax`/`Secure` configurable,
  token opaco y solo su SHA-256 en Redis; TTL de inactividad **30 min** + vida absoluta **1 h**).
- **P5**: contraseñas con **bcrypt cost 12** y política FR-010 centralizada en `platform/password`.
- **P6**: bloqueo FR-006 con contadores Redis (`login:fail:*`/`login:block:*`, **5 fallos / 15 min**,
  por correo normalizado exista o no la cuenta); **el 5.º fallo responde el `401` genérico y crea el
  bloqueo, y el `429` aplica desde el 6.º intento** (F-03).
- **P7/P8**: regla anti-bloqueo (FR-008) e inicialización única (FR-007) dentro de transacción con
  `pg_advisory_xact_lock`; la inicialización exige además `X-Setup-Token` = `BOOTSTRAP_TOKEN`.
- **P10/P11/P17**: CSRF *double-submit* firmado con `SESSION_SECRET`, CORS con credenciales y
  `rate-limit` en memoria sobre los endpoints públicos escribibles.
- **P12/P13**: un rol por cuenta (`users.role_id`), catálogo de 9 permisos sembrado por migración,
  normalización de correo y nombre de rol con `UNIQUE` en la base.
- **P20–P23**: auditoría duradera en PostgreSQL (`login_events`, `admin_actions`, solo inserción),
  `users.last_login_*`, consulta de solo lectura con el permiso `admin_usuarios_roles`, y Redis
  **sin persistencia** (`docker-compose.yml` sin volumen ni `appendonly`).
- **P19**: el delta de `contracts/openapi.yaml` se fusiona **aditivamente** en
  `backend/api/openapi.yaml` (`info.version` 0.2.0 → **0.3.0**); el snapshot de `specs/` no se edita.

## Discrepancias del `analyze` ya resueltas (orquestador, 2026-10-04; no se reabren)

1. **Nomenclatura de rutas frontend (F-05)**: se usan **las 7 rutas del plan** —`/login`,
   `/cambiar-contrasena`, `/panel`, `/panel/usuarios`, `/panel/roles`, `/panel/auditoria`,
   `/sin-permiso`— en todas las tareas y en los e2e (T244, T251, T252). Los **contenidos, textos y
   estados siguen siendo los de `ux.md`** (§3, §4, §7), que alinea sus rutas (`disenador-ux`).
2. **Pantalla de "Puesta en marcha" (F-06)** **no se construye en F2**: la inicialización única es
   una operación de despliegue vía `POST /api/v1/setup/initialize` con `X-Setup-Token` (P8),
   documentada en `quickstart.md` §1; el e2e la ejecuta **vía API**. La retirada de esa pantalla de
   `ux.md` la hace `disenador-ux`.
3. **Listado de usuarios (F-04)**: se mantiene la **paginación** (`platform/paginate`) y se retira
   del MVP el **buscador de texto libre** (no está en la spec): ninguna tarea lo crea.
4. **Contadores FR-006 (F-13)**: el **mecanismo** Redis (`login:fail:*`/`login:block:*`) vive en el
   plumbing `platform/session` (junto al `Store` de sesiones) y la **semántica y las constantes**
   (5 fallos / 15 min) en el dominio `usuarios`. **Semántica del 5.º intento (F-03)**: el contador
   se incrementa con cada fallo; el **5.º fallo** responde el `401` genérico **y crea el bloqueo**;
   desde el **6.º intento** (dentro de los 15 min) responde `429` con `Retry-After`.
5. **Componentes UI (F-10)**: inventario fijo con **nombres de código en inglés** (`Table`, `Tabs`,
   `Pagination`, `DateRangeFilter`, `StatusPill`, `Notice`, `ModalDialog`, `Field`, `ConfirmDialog`,
   `EmptyState`) en T242; `disenador-ux` da en `ux.md` el mapeo a sus etiquetas en español. Las
   features los **reutilizan sin duplicar markup** (lo revisa `revisor-codigo`).
6. **`platform/testutil` se amplía** con helpers de `testcontainers-go` (Redis y PostgreSQL) para no
   duplicar el arranque de servicios en las pruebas de integración (R19). Extensión del helper de F1,
   sin cambios en su API pública.

## Leyenda

| Marca | Significado |
|---|---|
| `[backend]` `[frontend]` `[db]` `[infra]` | Capa y subagente responsable (`[backend]`/`[db]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` → `devops`) |
| `[P1]`…`[P7]` | **Paralelizable** solo dentro de su grupo (archivos distintos y sin dependencia previa entre ellas; ver "Grupos de paralelismo") |
| *(sin `[P]`)* | Secuencial: tiene dependencias previas que deben estar integradas |
| *(FR-…, US…)* | Requisitos de `spec.md` que cubre la tarea |

## Tabla resumen

| ID | Tarea | Capa | Fase | Paralelo | Depende de |
|---|---|---|---|---|---|
| T201 | Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.3.0) | `[backend]` | 1 | P1 | — |
| T202 | Dependencias Go nuevas justificadas (x/crypto, uuid, go-redis, testcontainers) | `[backend]` | 1 | P1 | — |
| T203 | Dependencias npm nuevas (react-hook-form, zod, @hookform/resolvers) | `[frontend]` | 1 | P1 | — |
| T204 | Regenerar `frontend/src/api/schema.d.ts` desde el contrato fusionado | `[frontend]` | 1 | — | T201, T203 |
| T205 | Servicio `redis` en `docker-compose.yml` (sin persistencia) + variables del backend | `[infra]` | 2 | P2 | — |
| T206 | `.env.example`: variables de sesión y de arranque | `[infra]` | 2 | P2 | — |
| T207 | Migración `000002` roles + permissions (con siembra del catálogo) | `[db]` | 3 | — | — |
| T208 | Migración `000003` users | `[db]` | 3 | — | T207 |
| T209 | Migración `000004` login_events + admin_actions + `users.last_login_*` | `[db]` | 3 | — | T208 |
| T210 | Consultas sqlc `users.sql`, `roles.sql`, `permissions.sql` + generar | `[db]` | 3 | — | T207, T208 |
| T211 | Consultas sqlc `audit.sql` (solo INSERT/SELECT) + generar | `[db]` | 3 | — | T209, T210 |
| T212 | `platform/apperr`: + Invalid, Unauthenticated, Forbidden, Conflict, RateLimited | `[backend]` | 4 | — | — |
| T213 | `platform/validate` (etiquetas de DTO, incl. `phone`) | `[backend]` | 4 | P3 | T212 |
| T214 | `platform/paginate` (limit/offset con topes) | `[backend]` | 4 | P3 | — |
| T215 | `platform/password` (bcrypt 12 + política FR-010) | `[backend]` | 4 | P3 | — |
| T216 | `platform/config` ampliado (Redis, sesión, `BOOTSTRAP_TOKEN`) | `[backend]` | 4 | P3 | — |
| T217 | `platform/audit` (tipos `Event`/`Action` + interfaz `Recorder`) | `[backend]` | 4 | P3 | — |
| T218 | `platform/session`: token, cookies, `Identity`/`Resolver`, `Store` Redis + contadores | `[backend]` | 4 | — | T212, T216 |
| T219 | `platform/testutil`: helpers de testcontainers (Redis y PostgreSQL) | `[backend]` | 4 | P3 | — |
| T220 | `usuarios`: `model.go` + `repository.go` (mapeo pgtype → dominio) | `[backend]` | 5 | — | T210, T212 |
| T221 | `usuarios`: `repository_users.go` + integración | `[backend]` | 5 | P4 | T220 |
| T222 | `usuarios`: `repository_roles.go` + guard anti-bloqueo + integración | `[backend]` | 5 | P4 | T220 |
| T223 | `usuarios`: `repository_audit.go` (inserción y consulta) + integración | `[backend]` | 5 | P4 | T220 |
| T224 | `usuarios`: `service_audit.go` — registro de accesos y acciones (éxito y fallo) | `[backend]` | 5 | — | T223, T217 |
| T225 | `usuarios`: `service_auth.go` — login, logout y sesión | `[backend]` | 5 | — | T221, T224, T218, T215, T213 |
| T226 | `usuarios`: `service_auth.go` — cambio de contraseña propio y `mustChangePassword` | `[backend]` | 5 | — | T225 |
| T227 | `platform/middleware`: authn, authz, csrf, ratelimit, CORS con credenciales, cadena | `[backend]` | 5 | — | T218, T217, T225 |
| T228 | `handler.go` + `handler_auth.go` + `routes.go` + cableado en `cmd/api/main.go` | `[backend]` | 5 | — | T225, T226, T227 |
| T229 | `usuarios`: inicialización única del administrador (FR-007) | `[backend]` | 6 | — | T221, T222, T224 |
| T230 | `handler_auth.go`: `POST /api/v1/setup/initialize` + rate-limit + cableado | `[backend]` | 6 | — | T229, T228 |
| T231 | `usuarios`: `service_users.go` — crear cuenta | `[backend]` | 7 | — | T221, T222, T224 |
| T232 | `usuarios`: `service_users.go` — editar, activar/desactivar (anti-bloqueo, revocación) | `[backend]` | 7 | — | T231 |
| T233 | `usuarios`: `service_users.go` — restablecer contraseña | `[backend]` | 7 | — | T232 |
| T234 | `usuarios`: `handler_users.go` + rutas `/api/v1/admin/usuarios*` + cableado | `[backend]` | 7 | — | T233, T228 |
| T235 | `usuarios`: `service_roles.go` — crear rol y catálogo de permisos | `[backend]` | 8 | — | T222, T224 |
| T236 | `usuarios`: `service_roles.go` — editar y eliminar rol (anti-bloqueo) | `[backend]` | 8 | — | T235 |
| T237 | `usuarios`: `handler_roles.go` + rutas `/api/v1/admin/roles*` y `/permisos` + cableado | `[backend]` | 8 | — | T236, T228 |
| T238 | `usuarios`: `service_audit.go` — consulta con filtros y paginación | `[backend]` | 9 | — | T223, T214 |
| T239 | Puntos de escritura de auditoría restantes: `authz` (denied) y `handler` (JSON inválido) | `[backend]` | 9 | — | T224, T227, T228 |
| T240 | `usuarios`: `handler_audit.go` + `GET /api/v1/admin/auditoria/*` + cableado | `[backend]` | 9 | — | T238, T239, T228 |
| T241 | `api/client.ts` (credenciales + `X-CSRF-Token`) y `api/auth.ts` | `[frontend]` | 10 | P5 | T204 |
| T242 | Componentes compartidos (`Table`, `Tabs`, `Pagination`, `DateRangeFilter`, `StatusPill`, `Notice`, `ModalDialog`, `Field`, `ConfirmDialog`, `EmptyState`) y `lib/` (permisos, formato) | `[frontend]` | 10 | P5 | — |
| T243 | `api/usuarios.ts`, `api/roles.ts`, `api/auditoria.ts` | `[frontend]` | 10 | — | T241 |
| T244 | Guards (`RequireAuth`, `RequirePermission`, `RequirePasswordChange`), router y layout por permisos | `[frontend]` | 10 | — | T241, T242 |
| T245 | `features/auth`: LoginPage y hooks de sesión | `[frontend]` | 10 | — | T244 |
| T246 | `features/auth`: ChangePasswordPage, guard de cambio obligatorio y aviso de sesión | `[frontend]` | 10 | — | T245 |
| T247 | `features/usuarios`: UsersPage (listado con estado, correo, rol y último acceso) | `[frontend]` | 10 | P6 | T243, T244 |
| T248 | `features/usuarios`: UserForm, activar/desactivar y restablecer contraseña | `[frontend]` | 10 | — | T247 |
| T249 | `features/roles`: RolesPage y RoleForm con permisos por módulo | `[frontend]` | 10 | P6 | T243, T244 |
| T250 | `features/auditoria`: AuditPage de solo lectura con filtros y paginación | `[frontend]` | 10 | P6 | T243, T244 |
| T254 | `features/panel`: Inicio del panel (`/panel`) — "Mi cuenta", accesos rápidos y estado "cuenta sin permisos de módulo" *(nueva por el `analyze` F-02)* | `[frontend]` | 10 | P6 | T241, T242, T244, T245 |
| T251 | E2E Playwright `acceso.spec.ts` (recorrido de `quickstart.md` §11) | `[frontend]` | 10 | P7 | T245, T246, T248, T249, T254 |
| T252 | E2E Playwright `auditoria.spec.ts` (recorrido de `quickstart.md` §10) | `[frontend]` | 10 | P7 | T250 |
| T253 | Verificación de cierre: `make ci`, `make e2e`, `quickstart.md` §0–§12 y mapa SC | `[infra]` | 11 | — | todas |

**Total: 54 tareas** — 31 `[backend]` · 15 `[frontend]` · 5 `[db]` · 3 `[infra]` · 22 marcadas `[P]`.

## Grupos de paralelismo

| Grupo | Tareas | Por qué pueden ir en paralelo |
|---|---|---|
| P1 | T201 · T202 · T203 | Contrato, `go.mod` y `package.json` son archivos distintos sin dependencias entre sí |
| P2 | T205 · T206 | `docker-compose.yml` y `.env.example` son archivos distintos |
| P3 | T213 · T214 · T215 · T216 · T217 · T219 | Paquetes `platform` distintos (solo T213 necesita T212 ya integrado) |
| P4 | T221 · T222 · T223 | `repository_users.go`, `repository_roles.go` y `repository_audit.go` son archivos distintos del mismo paquete; todos parten de T220 |
| P5 | T241 · T242 | `src/api/` y `src/components/`+`src/lib/` no se solapan |
| P6 | T247 · T249 · T250 · T254 | Features `usuarios`, `roles`, `auditoria` y `panel` (Inicio) en carpetas distintas; las tres primeras con su módulo de API (T243) y el Inicio solo con los hooks de sesión (T245) |
| P7 | T251 · T252 | Dos specs e2e independientes (`e2e/acceso.spec.ts` y `e2e/auditoria.spec.ts`) |

> **Nota sobre la auditoría (P20)**: el **registro** se construye junto a cada productor
> (`service_auth`, `service_users`, `service_roles`) porque FR-022/FR-023 exigen registrar cada
> desenlace en su misma transacción; la **Fase 9** cierra los puntos de escritura transversales
> (`authz`, `handler`) y toda la **consulta**. Es deliberado: una tarea de "registro" al final
> obligaría a reabrir tareas ya cerradas.

---

## Fase 1 — Contrato OpenAPI, dependencias y tipos

- [ ] T201 · Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.3.0) · `[backend]` `[P1]`

  - **Archivos**: `backend/api/openapi.yaml` (EDITADO). *No* se toca
    `specs/002-acceso-gestion-usuarios/contracts/openapi.yaml` (snapshot inmutable, P19).
  - **Qué hace**: funde **aditivamente** el delta aprobado en el contrato vivo: esquema de seguridad
    `sessionCookie`, las **18** operaciones nuevas (`/api/v1/auth/*`, `/api/v1/setup/initialize`,
    `/api/v1/admin/usuarios*`, `/api/v1/admin/roles*`, `/api/v1/admin/permisos`,
    `/api/v1/admin/auditoria/*`) y sus schemas (`LoginInput`, `SessionUser`, `UserItem`, `RoleItem`,
    `AccessEventItem` con `userName`/`userEmail`, `AdminActionItem` con `actorName`/`actorEmail`…)
    reutilizando `ErrorEnvelope`/`ErrorBody` de F1; sube
    `info.version` a **0.3.0**. El registro de `error.code` **no cambia** (P19): el `404`
    `not_found` de F1 **se reutiliza** para "cuenta/rol inexistente". *Cubre el contrato de
    todos los FR*; es el que garantiza **FR-025** (solo `GET` sobre `/admin/auditoria/*`) y **FR-003**
    (ningún DTO de salida contiene contraseñas ni hashes).
  - **Pruebas incluidas** (§III): — (artefacto de contrato). **Cómo se verifica**: el YAML parsea sin
    errores; búsqueda de métodos sobre `/api/v1/admin/auditoria` → solo `get` (FR-025); ninguna
    propiedad `password`/`passwordHash` en schemas de respuesta (FR-003/FR-026);
    `grep -n "version:" backend/api/openapi.yaml` → `0.3.0`; `make api-gen` completa sin errores.
  - **Criterio de terminado**: el contrato vivo contiene F1 + F2 y es la única fuente de tipos del
    frontend (T204). Los códigos de error emitidos (`invalid`, `unauthenticated`, `forbidden`,
    `conflict`, `rate_limited`) ya estaban en el enum de F1.
  - **Commit sugerido**: `docs(api): fusionar delta de F2 en el contrato vivo (0.3.0)`

- [ ] T202 · Dependencias Go nuevas justificadas · `[backend]` `[P1]`

  - **Archivos**: `backend/go.mod`, `backend/go.sum` (EDITADOS).
  - **Qué hace**: añade las dependencias justificadas en `research.md` R16: runtime →
    `golang.org/x/crypto` (bcrypt; §IV lo exige), `github.com/google/uuid` (tipos UUID del dominio,
    §8.1.1) y `github.com/redis/go-redis/v9` (sesión y contadores, P1/P6); solo pruebas de
    integración → `github.com/testcontainers/testcontainers-go` (R19: el CI del kit no se toca).
  - **Pruebas incluidas** (§III): — (manifiesto de dependencias). **Cómo se verifica**:
    `cd backend && go mod tidy && go mod verify && go build ./...` en verde;
    `govulncheck ./...` sin vulnerabilidades; `go list -m all` muestra los cuatro módulos.
  - **Criterio de terminado**: `go.mod`/`go.sum` consistentes y solo con las dependencias
    justificadas (D-A8); ningún archivo del kit modificado.
  - **Commit sugerido**: `chore(backend): agregar dependencias de F2 (x/crypto, uuid, go-redis, testcontainers)`

- [ ] T203 · Dependencias npm nuevas · `[frontend]` `[P1]`

  - **Archivos**: `frontend/package.json`, `frontend/package-lock.json` (EDITADOS).
  - **Qué hace**: añade `react-hook-form`, `zod` y `@hookform/resolvers` (convención de la skill para
    formularios con validación tipada; justificación en `research.md` R16). Nada más: no se instala
    ninguna otra librería (P18: sin redux/zustand; el estado es TanStack Query).
  - **Pruebas incluidas** (§III): — (manifiesto). **Cómo se verifica**: `npm ci && npm run lint &&
    npm run typecheck && npm test -- --run && npm run build` en verde desde `frontend/`;
    `npm audit` sin altas ni críticas (§IV).
  - **Criterio de terminado**: solo las tres dependencias nuevas; el árbol sigue compilando.
  - **Commit sugerido**: `chore(frontend): agregar react-hook-form, zod y @hookform/resolvers`

- [ ] T204 · Regenerar `frontend/src/api/schema.d.ts` · `[frontend]`

  - **Archivos**: `frontend/src/api/schema.d.ts` (GENERADO y commiteado).
  - **Qué hace**: ejecuta `make api-gen` (o `npm run api:gen`) contra **`backend/api/openapi.yaml`**
    (nunca contra el snapshot de `specs/`, P19) y commitea los tipos de F2 (`SessionUser`, `UserItem`,
    `UserList`, `RoleItem`, `AccessEventItem`, `AdminActionItem`, los `Input`, etc.).
  - **Pruebas incluidas** (§III): — (artefacto generado). **Cómo se verifica**: regenerar no produce
    `git diff`; `npm run typecheck` pasa con los tipos nuevos.
  - **Criterio de terminado**: `schema.d.ts` reproducible desde el contrato y listo para el cliente de
    la Fase 10 (regla anti-deriva: quien cambia el contrato regenera en el mismo PR).
  - **Commit sugerido**: `build(frontend): regenerar schema.d.ts con los tipos de F2`

---

## Fase 2 — Infraestructura local (`[infra]`, la aplica `devops`)

- [ ] T205 · Servicio `redis` en `docker-compose.yml` · `[infra]` `[P2]`

  - **Archivos**: `docker-compose.yml` (EDITABLE: no está en `.kit-manifest.json`). *No se toca*
    `.github/workflows/ci.yml` (archivo del kit, R19).
  - **Qué hace**: añade el servicio **`redis`** (`redis:7-alpine`, puerto parametrizado
    `${REDIS_PORT:-6379}:6379`, `healthcheck` con `redis-cli ping`) **sin volumen, sin `appendonly`
    y sin `save`** (P23: persistencia desactivada a propósito). El servicio `backend` pasa a
    `depends_on: {db: service_healthy, redis: service_healthy}` y recibe `REDIS_URL`
    (apuntando al host `redis`, nunca `localhost`), `SESSION_SECRET`, `SESSION_COOKIE_SECURE`,
    `SESSION_IDLE_TTL_MINUTES`, `SESSION_ABSOLUTE_TTL_MINUTES` y `BOOTSTRAP_TOKEN`. *Soporta
    FR-005/FR-006* (sesión y contadores en Redis) y §VII (`make up` levanta todo).
  - **Pruebas incluidas** (§III): — (orquestación). **Cómo se verifica**: `make up` desde un clon
    limpio levanta `db`, `redis`, `backend` y `frontend`; `docker compose exec redis redis-cli ping`
    → `PONG`; `docker compose restart redis` borra el estado (claves `sess:*` desaparecen, P23);
    `curl -i http://localhost:8080/healthz` responde igual que en F1 (§8.1.9);
    `docker compose config` no declara volúmenes ni `appendonly` en `redis`.
  - **Criterio de terminado**: los cuatro servicios levantan con `make up`; los errores de Redis
    caído se traducen a `503`/`500` con sobre uniforme (R1); ningún archivo del kit modificado.
  - **Commit sugerido**: `build(compose): servicio redis sin persistencia y variables de sesión`

- [ ] T206 · `.env.example`: variables de sesión y de arranque · `[infra]` `[P2]`

  - **Archivos**: `.env.example` (EDITABLE).
  - **Qué hace**: documenta las variables nuevas con valores de ejemplo (nunca secretos reales, §IV):
    `REDIS_URL` (`redis://localhost:6379/0`), `SESSION_SECRET`, `BOOTSTRAP_TOKEN`,
    `SESSION_COOKIE_SECURE` (`false` en desarrollo), `SESSION_IDLE_TTL_MINUTES` (`30`) y
    `SESSION_ABSOLUTE_TTL_MINUTES` (`60`) — las mismas que usan `quickstart.md` §0 y `platform/config`
    (T216). Cada variable se documenta con su consumidor.
  - **Pruebas incluidas** (§III): — (documento de configuración). **Cómo se verifica**: ningún valor
    parece secreto real (lo audita `seguridad`); `cp .env.example .env` es usable tal cual en
    desarrollo; ninguna variable sin consumidor y ningún consumidor sin su variable.
  - **Criterio de terminado**: la lista documentada coincide exactamente con la que lee
    `platform/config` (T216) y con la que inyecta compose (T205).
  - **Commit sugerido**: `chore(env): documentar REDIS_URL, sesión y BOOTSTRAP_TOKEN en .env.example`

---

## Fase 3 — Migraciones y capa de datos (`[db]`)

> Convenciones de la skill `postgres-db` y de `data-model.md`: `id UUID`, `created_at`/`updated_at`,
> FK e índices explícitos, `CHECK`/`UNIQUE` **en la base**, `up`/`down` completos, y **nunca** se
> edita una migración aplicada. Numeración sin huecos desde `000001` (F1).

- [ ] T207 · Migración `000002_create_roles_and_permissions` · `[db]`

  - **Archivos**: `backend/migrations/000002_create_roles_and_permissions.up.sql` y
    `.down.sql` (NUEVOS).
  - **Qué hace**: crea `permissions` (con la **siembra del catálogo fijo de 9 permisos** de FR-015:
    portada, eventos, actividades, grupos, ministerios, donaciones, noticias, medios y
    `admin_usuarios_roles`), `roles` (con `UNIQUE` normalizado `roles_name_lower_idx`) y
    `role_permissions` (`UNIQUE (role_id, permission_id)`, `ON DELETE CASCADE` en `role_id`,
    `ON DELETE RESTRICT` en `permission_id`). *Cubre FR-014 (parte de datos), FR-015, FR-017*.
  - **Pruebas incluidas** (§III): ejecución real del toolchain y aserciones SQL. **Cómo se verifica**:
    `make up && make db-migrate` aplica `000002`; el `down` la revierte sin dejar tablas; el catálogo
    sembrado tiene exactamente los 9 códigos de `data-model.md`; insertar dos roles que solo difieren
    en mayúsculas → violación de `roles_name_lower_idx`; duplicar `(role_id, permission_id)` → error.
  - **Criterio de terminado**: versión 2 aplicada y reversible; el catálogo de permisos existe en la
    base (no solo en código).
  - **Commit sugerido**: `feat(db): migración 000002 de roles y permisos con catálogo sembrado`

- [ ] T208 · Migración `000003_create_users` · `[db]`

  - **Archivos**: `backend/migrations/000003_create_users.up.sql` y `.down.sql` (NUEVOS).
  - **Qué hace**: crea `users` con `email` (`UNIQUE` + `CHECK (email = lower(btrim(email)))`),
    `first_name`, `last_name`, `phone` (obligatorio, `CHECK (phone = btrim(phone))`),
    `password_hash` (solo hash bcrypt, §IV), `must_change_password`, `is_active` y **`role_id` NOT
    NULL** con `ON DELETE RESTRICT` (un rol por cuenta, Q4; un rol en uso no se elimina, FR-017).
    Índices: `users_role_id_idx`, `users_created_at_id_idx` (orden por defecto, §8.1.7) y el parcial
    `users_is_active_idx` para el recuento del guard anti-bloqueo (FR-008). *Cubre FR-009 (datos),
    FR-012/FR-013 (sin borrado: no hay `DELETE`), FR-017, FR-019*.
  - **Pruebas incluidas** (§III): ejecución del toolchain + aserciones SQL. **Cómo se verifica**:
    `make db-migrate` aplica `000003` y el `down` la revierte; crear un usuario sin `role_id` → error
    por `NOT NULL`; eliminar un rol con usuarios → error por `RESTRICT` (FR-017); insertar
    `"  Ana@Ejemplo.com "` → error por el `CHECK` de normalización (Q5).
  - **Criterio de terminado**: versión 3 aplicada y reversible; las invariantes de datos de
    `data-model.md` están **en la base**, no solo en el código.
  - **Commit sugerido**: `feat(db): migración 000003 de cuentas del panel`

- [ ] T209 · Migración `000004_create_login_events_and_admin_actions` · `[db]`

  - **Archivos**: `backend/migrations/000004_create_login_events_and_admin_actions.up.sql` y
    `.down.sql` (NUEVOS).
  - **Qué hace**: añade `users.last_login_at`/`last_login_ip` (proyección del último acceso exitoso,
    FR-021), `login_events` (FR-022: `result IN ('success','failure')`, `ip`, `user_id` anulable) y
    `admin_actions` (FR-023: `action` en los 8 códigos, `target_kind`, `target_user_id`/`target_role_id`
    (`ON DELETE SET NULL` en el rol, para que el registro sobreviva a la eliminación de un rol,
    FR-025), `target_label`, `result IN ('success','failure','denied')`, `CHECK` de actor solo ausente
    en la inicialización). Índices por `created_at DESC, id DESC` y por cuenta + fecha (FR-024).
    **Sin columnas de credenciales** (FR-026). *Cubre FR-021…FR-026 (parte de datos)*.
  - **Pruebas incluidas** (§III): ejecución del toolchain + aserciones SQL. **Cómo se verificar**:
    `make db-migrate` aplica `000004` y el `down` la revierte **incluidas** las columnas
    `last_login_*`; un `admin_actions` con `target_kind='user'` y `target_role_id` relleno → error
    (CHECK); un `admin_actions` sin actor y acción `role.create` → error; `\d login_events` no muestra
    ninguna columna de contraseña ni de correo de intento (FR-026).
  - **Criterio de terminado**: versión 4 aplicada y reversible; las tablas de auditoría son de solo
    inserción por diseño (sin `UPDATE`/`DELETE` posibles desde el contrato, FR-025) y no contienen
    credenciales.
  - **Commit sugerido**: `feat(db): migración 000004 de auditoría y último acceso`

- [ ] T210 · Consultas sqlc `users.sql`, `roles.sql`, `permissions.sql` · `[db]`

  - **Archivos**: `backend/internal/db/queries/users.sql`, `roles.sql`, `permissions.sql` (NUEVOS) y
    el código generado `backend/internal/db/` (GENERADO y commiteado).
  - **Qué hace**: escribe las consultas parametrizadas de `data-model.md` (sin `SELECT *`, con
    `ORDER BY` determinista y `LIMIT`/`OFFSET`): `users.sql` → `InsertUser`, `GetUserByID`,
    `GetUserByEmail`, `GetUserAuthByEmail` (correo + hash + estado + rol + permisos, para
    login/`authn`), `ListUsers` + `CountUsers`, `UpdateUser`, `UpdateUserPassword`,
    `SetUserMustChangePassword`, `UpdateUserLastLogin`, `CountActiveAdmins`, `CountUsersByRole`,
    `CountUsers`; `roles.sql` → `InsertRole`, `GetRoleByID`, `GetRoleByNameLower`, `ListRoles` +
    `CountRoles`, `UpdateRoleName`, `DeleteRolePermissions`, `InsertRolePermission`, `DeleteRole`,
    `CountRoleUsers`, `LockAdminGuard` (`pg_advisory_xact_lock`, P7); `permissions.sql` →
    `ListPermissions`, `GetPermissionIDsByCodes`. **Sin consultas de sesión ni de contadores** (viven
    en Redis, P1/P6). Ejecuta `make sqlc-gen`.
  - **Pruebas incluidas** (§III): — (SQL de consultas + código generado). **Cómo se verifica**:
    `make sqlc-gen` compila el SQL contra las migraciones; `make sqlc-verify` sin deriva;
    `go build ./...` compila con el código generado; cada consulta usa parámetros `$1…$n`.
  - **Criterio de terminado**: el código generado está commiteado y acompaña al SQL (regla de
    revisión: un PR que toca `queries/` toca `internal/db/`).
  - **Commit sugerido**: `feat(db): consultas sqlc de cuentas, roles y permisos`

- [ ] T211 · Consultas sqlc `audit.sql` · `[db]`

  - **Archivos**: `backend/internal/db/queries/audit.sql` (NUEVO) + código generado (GENERADO).
  - **Qué hace**: `InsertLoginEvent`, `ListLoginEvents` + `CountLoginEvents` (filtros `userId`,
    `from`, `to` — semirango `[from, to)`), `InsertAdminAction`, `ListAdminActions` +
    `CountAdminActions` (filtro por cuenta **involucrada**: `actor OR target`; rango de fechas);
    orden `created_at DESC, id DESC` con `LIMIT`/`OFFSET`. **Solo `INSERT` y `SELECT`**: sin
    `UPDATE` ni `DELETE` sobre el registro (FR-025). *Cubre FR-022…FR-025 (capa de datos)*.
  - **Pruebas incluidas** (§III): — (SQL + generado). **Cómo se verifica**: `make sqlc-gen &&
    make sqlc-verify` en verde; `grep -E "UPDATE|DELETE" backend/internal/db/queries/audit.sql` → 0
    coincidencias (FR-025); `go build ./...` compila.
  - **Criterio de terminado**: la consulta de auditoría es filtrable y paginable desde el SQL, y el
    registro es estructuralmente de solo inserción.
  - **Commit sugerido**: `feat(db): consultas sqlc de auditoría (solo inserción y consulta)`

---

## Fase 4 — Núcleo de plataforma (`backend/internal/platform/`)

> `platform` **no conoce dominios** (arq. R1); los tipos de plumbing (`Identity`, `Resolver`,
> `Recorder`) los implementa el dominio `usuarios` (excepciones declaradas en el plan, P15/P20).

- [ ] T212 · `platform/apperr`: kinds de F2 · `[backend]`

  - **Archivos**: `backend/internal/platform/apperr/apperr.go` y `apperr_test.go` (EDITADOS).
  - **Qué hace**: añade al registro cerrado de F1 los kinds que F2 emite (P16, tabla de §5.11 de
    `arquitectura.md`): `Invalid` (400), `Unauthenticated` (401), `Forbidden` (403), `Conflict` (409)
    y `RateLimited` (429, con `Retry-After` y `details.retryAfterSeconds`), cada uno con su
    constructor, `Message` en español seguro para el cliente y error interno envuelto con `%w` que
    nunca se serializa.
  - **Pruebas incluidas** (§III): tabla de casos. **Cómo se verifica**:
    `go test ./internal/platform/apperr/` en verde: cada kind → status y `error.code` correctos;
    `errors.Is/As` y `Unwrap` con error envuelto; `Message` nunca contiene el error interno.
  - **Criterio de terminado**: los cinco kinds nuevos traducen a HTTP exactamente como el contrato
    (T201) documenta; ningún código fuera del registro.
  - **Commit sugerido**: `feat(platform): apperr con Invalid, Unauthenticated, Forbidden, Conflict y RateLimited`

- [ ] T213 · `platform/validate` — validación de DTOs por etiquetas · `[backend]` `[P3]`

  - **Archivos**: `backend/internal/platform/validate/validate.go`, `validate_test.go` (NUEVOS).
  - **Qué hace**: validación propia mínima por reflexión sobre etiquetas `validate` de los DTOs (P3):
    `required`, `omitempty`, `min`, `max`, `email`, `oneof` y **`phone`** (FR-009: dígitos con
    espacios/guiones/paréntesis, prefijo internacional opcional y **≥ 7 dígitos**) →
    `apperr.Invalid` con `details` por campo en español (CWE-20: toda entrada se valida en backend
    aunque el frontend ya valide). Sin dependencias (D-A8).
  - **Pruebas incluidas** (§III): tabla de casos por etiqueta. **Cómo se verifica**:
    `go test ./internal/platform/validate/` en verde: cada etiqueta acepta lo válido y rechaza lo
    inválido con el campo exacto en `details`; casos `phone` (`"+34 612 345 678"` ok, `"12"` y
    `"no es un teléfono"` → `details.phone`); un DTO válido produce `nil`.
  - **Criterio de terminado**: las **7** etiquetas cubiertas (`required`, `omitempty`, `min`, `max`,
    `email`, `oneof`, `phone`); el mensaje de cada campo es
    comprensible para personas no técnicas (spec, "Errores esperados").
  - **Commit sugerido**: `feat(platform): validate por etiquetas con details por campo`

- [ ] T214 · `platform/paginate` — paginación con topes · `[backend]` `[P3]`

  - **Archivos**: `backend/internal/platform/paginate/paginate.go`, `paginate_test.go` (NUEVOS).
  - **Qué hace**: parseo y normalización de `limit`/`offset` para los listados (P14, §8.1.2/§8.1.3):
    defecto **20**, tope **100**, `offset ≥ 0`, entradas no numéricas o negativas → `apperr.Invalid`.
    El sobre de salida `{items, total, limit, offset}` lo construye cada handler.
  - **Pruebas incluidas** (§III): tabla de casos. **Cómo se verifica**:
    `go test ./internal/platform/paginate/` en verde: sin parámetros → 20/0; `limit=1000` → 100;
    `limit=0`, `limit=abc`, `offset=-1` → `apperr.Invalid` con `details`.
  - **Criterio de terminado**: ningún listado de F2 puede desbordar el tope (§8.1.3).
  - **Commit sugerido**: `feat(platform): paginate con defecto 20 y tope 100`

- [ ] T215 · `platform/password` — bcrypt 12 y política FR-010 · `[backend]` `[P3]`

  - **Archivos**: `backend/internal/platform/password/password.go`, `password_test.go` (NUEVOS).
  - **Qué hace**: `Hash`/`Verify` con **bcrypt cost 12** (`golang.org/x/crypto/bcrypt`, P5) y la
    política FR-010 centralizada (se aplica en los 3 flujos: creación, restablecimiento y cambio
    propio): **entre 8 y 64 caracteres** (tope `maxLength: 64` del contrato en todos los campos de
    contraseña), combinación de
    **mayúsculas, minúsculas, números y caracteres especiales**, y **distinta de** —igualdad con
    comparación normalizada (trim + minúsculas), **no de contenido**: puede contener esos datos— el
    nombre, los apellidos y el correo (que recibe como contexto). El límite de 72 bytes de bcrypt se
    mantiene como comprobación técnica (riesgo RG6). La contraseña nunca se loguea ni se devuelve.
  - **Pruebas incluidas** (§III): tabla de casos de la política + criptografía. **Cómo se verifica**:
    `go test ./internal/platform/password/` en verde: cada requisito incumplido produce el error que
    nombra el requisito (spec, "Errores esperados"); hash distinto por sal; `Verify` correcto/incorrecto;
    cost observado = 12; más de 64 caracteres → rechazada; y una contraseña **igual** al nombre, a los
    apellidos o al correo (normalizada) rechazada mientras que una que los **contenga** se acepta.
  - **Criterio de terminado**: los tres flujos de contraseña del dominio pueden compartir una única
    implementación de la política (§IV, CWE-256).
  - **Commit sugerido**: `feat(platform): password con bcrypt cost 12 y política FR-010`

- [ ] T216 · `platform/config` ampliado · `[backend]` `[P3]`

  - **Archivos**: `backend/internal/platform/config/config.go`, `config_test.go` (EDITADOS).
  - **Qué hace**: añade a la lista canónica las variables de F2 (coherentes con T205/T206):
    `REDIS_URL`, `SESSION_SECRET`, `SESSION_COOKIE_SECURE`, `SESSION_IDLE_TTL_MINUTES` (30),
    `SESSION_ABSOLUTE_TTL_MINUTES` (60) y `BOOTSTRAP_TOKEN`, con valores por defecto de desarrollo y
    **validación al arrancar** (§VII): `REDIS_URL` obligatoria y parseable (R1); si
    `APP_ENV=production` y `SESSION_COOKIE_SECURE=false` → **falla al arrancar** (R11); los minutos
    deben ser positivos. Ningún secreto en el código (§IV).
  - **Pruebas incluidas** (§III): tabla de casos con `t.Setenv`. **Cómo se verifica**:
    `go test ./internal/platform/config/` en verde: cada variable nueva se carga con su defecto;
    producción con cookie insegura → error que nombra la variable; `REDIS_URL` ausente → error claro.
  - **Criterio de terminado**: la lista de `platform/config` = la documentada en `.env.example`
    (T206) = la inyectada por compose (T205).
  - **Commit sugerido**: `feat(platform): config de Redis, sesión y BOOTSTRAP_TOKEN`

- [ ] T217 · `platform/audit` — plumbing del registro · `[backend]` `[P3]`

  - **Archivos**: `backend/internal/platform/audit/audit.go`, `audit_test.go` (NUEVOS).
  - **Qué hace**: los tipos de plumbing `Event` (intento de acceso: resultado, IP, usuario opcional) y
    `Action` (acción administrativa: actor, código de acción, objetivo, resultado) y la interfaz
    `Recorder` que consume `middleware/authz` para registrar denegaciones **sin conocer dominios**
    (P20; misma excepción declarada que `session.Identity`/`Resolver`, plan §Complexity Tracking 4).
    La implementa el dominio `usuarios` (T224). Sin SQL, sin Redis, sin HTTP.
  - **Pruebas incluidas** (§III): unitarias mínimas de los tipos (constructores, campos, validación de
    códigos de acción). **Cómo se verifica**: `go test ./internal/platform/audit/` en verde; el
    paquete no importa ningún dominio (revisión de `revisor-codigo`, arq. R1).
  - **Criterio de terminado**: `authz` puede registrar una denegación con solo este paquete.
  - **Commit sugerido**: `feat(platform): tipos e interfaz Recorder de auditoría`

- [ ] T219 · `platform/testutil`: helpers de testcontainers · `[backend]` `[P3]`

  - **Archivos**: `backend/internal/platform/testutil/containers.go`, `containers_test.go`
    (NUEVOS; extiende el `testutil` de F1 sin cambiar su API).
  - **Qué hace**: helpers de `testcontainers-go` para las pruebas de integración (R19: el CI del kit
    no se toca y los runners tienen Docker): levantar **Redis** (`redis:7-alpine`) y **PostgreSQL**
    (o reutilizar `DATABASE_URL_TEST` si existe) dentro del propio test, con limpieza al terminar y
    *skip* documentado si no hay Docker. Evita duplicar ese arranque en las suites de sesión y de
    repositorio.
  - **Pruebas incluidas** (§III): pruebas de los propios helpers (`//go:build integration`): levantan
    cada contenedor, devuelven una URL usable y lo liberan; sin Docker → *skip* con mensaje claro.
  - **Criterio de terminado**: las suites de T218/T221/T222/T223 pueden usar estos helpers; el CI del
    kit ejecuta `go test -tags=integration ./...` sin cambios.
  - **Commit sugerido**: `test(platform): helpers de testcontainers para Redis y PostgreSQL`

- [ ] T218 · `platform/session` — sesión en Redis y contadores de acceso · `[backend]`

  - **Archivos**: `backend/internal/platform/session/` (`token.go`, `cookies.go`, `identity.go`,
    `store.go`, `store_redis.go`, `throttle.go` y sus `*_test.go`; NUEVOS).
  - **Qué hace**: (a) token aleatorio de 32 bytes (`crypto/rand`) y su **SHA-256** como clave Redis
    (el token en claro nunca se guarda, P1); (b) cookies `ss_session` (`HttpOnly`, `SameSite=Lax`,
    `Secure` configurable, `Path=/`, `Max-Age` = vida absoluta) y `csrf_token` (no `HttpOnly`) con
    `nonce.HMAC-SHA256(SESSION_SECRET, nonce)` para el *double-submit* firmado (P10); (c) tipos
    `Identity`/`Resolver` (plumbing que implementa el dominio, R3); (d) `Store` sobre Redis
    (`github.com/redis/go-redis/v9`): claves `sess:<sha256>` con JSON `userId`/`createdAt`/
    `lastSeenAt`/`absoluteExpiresAt`, **TTL de inactividad 30 min** refrescado por actividad
    (estrangulado a 1/min) y **acotado a la vida absoluta de 1 h** (`absoluteExpiresAt` inmóvil,
    R15), índice `user_sessions:<userId>` para revocar por cuenta (FR-012); (e) **mecanismo de los
    contadores de FR-006** (P6, reparto F-13): `login:fail:<identificador>` (`INCR`/`DEL`, TTL 15 min)
    y `login:block:<identificador>` (bandera con TTL 15 min creada por el 5.º fallo) con el resto de
    `Retry-After`. La **semántica y las constantes** (5 fallos / 15 min, y que el 5.º fallo aún
    responde `401`) las define el dominio `usuarios` (T225): este paquete solo aporta el mecanismo.
    *Cubre FR-001,
    FR-004, FR-005, FR-006 (mecanismo), FR-012 (revocación)*.
  - **Pruebas incluidas** (§III): unitarias (token distinto por llamada, SHA-256 estable, atributos de
    cada cookie, verificación del HMAC CSRF) + **integración** `//go:build integration` contra Redis
    real (T219): crear/resolver sesión; el TTL se refresca con la actividad pero el corte a los
    60 min se produce aunque haya actividad; revocación por cuenta borra `sess:*` y `user_sessions:*`;
    contadores: el contador crece con cada fallo y el **5.º fallo** crea la bandera de bloqueo (sin
    responder aún `429`: eso lo decide el dominio), `Retry-After` = TTL restante, `DEL` al entrar bien.
  - **Criterio de terminado**: `go test ./internal/platform/session/` y
    `go test -tags=integration ./internal/platform/session/` en verde; ningún valor de contraseña
    pasa por este paquete.
  - **Commit sugerido**: `feat(platform): sesión en Redis (30 min + 1 h) y contadores de intentos`

---

## Fase 5 — Autenticación y sesión (dominio `internal/usuarios/`)

> Un único dominio `usuarios` para todo F2 (P15), con archivos por responsabilidad
> (`service_auth.go`, `handler_users.go`… — desviación declarada del plan). Capas
> `handler → service → repository` (§II): los handlers no tocan SQL.

- [ ] T220 · `usuarios`: modelo y base del repositorio · `[backend]`

  - **Archivos**: `backend/internal/usuarios/model.go`, `repository.go` y
    `model_test.go`/`repository_test.go` (NUEVOS).
  - **Qué hace**: entidades del dominio (`User`, `Role`, `Permission`, `LoginEvent`, `AdminAction`,
    estados, códigos de acción) con tipos `github.com/google/uuid` (P4) y los DTOs del contrato con
    etiquetas `validate` (T213); `repository.go` con el constructor sobre el pool y el mapeo
    `pgtype.*` → dominio **solo aquí** (§8.1.6, sin `overrides` en sqlc). Sin lógica de negocio.
  - **Pruebas incluidas** (§III): unitarias del mapeo (nulos → punteros/`time.Time` correctos,
    `last_login_*` nulos, `target_*` nulos). **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde; `golangci-lint` sin `pgx` fuera de `repository.go`
    (arq. R4).
  - **Criterio de terminado**: las tres tareas `[P4]` pueden escribir sus repositorios sin duplicar
    mapeo.
  - **Commit sugerido**: `feat(usuarios): modelo del dominio y base del repositorio`

- [ ] T221 · `usuarios`: `repository_users.go` · `[backend]` `[P4]`

  - **Archivos**: `backend/internal/usuarios/repository_users.go`,
    `repository_users_integration_test.go` (`//go:build integration`; NUEVOS).
  - **Qué hace**: implementa sobre sqlc las consultas de cuentas: autenticación
    (`GetUserAuthByEmail`: correo + hash + estado + rol + permisos), lectura (`GetUserByID`,
    `ListUsers`/`CountUsers`), inserción (`InsertUser`), edición (`UpdateUser`),
    contraseña (`UpdateUserPassword`, `SetUserMustChangePassword`), último acceso
    (`UpdateUserLastLogin`, **solo** para el login exitoso, FR-021) y recuentos (`CountUsers`,
    `CountUsersByRole`). **Sin `DELETE` de cuentas en ninguna consulta** (FR-013). Errores con `%w`.
  - **Pruebas incluidas** (§III): **integración** contra PostgreSQL real (T219): `InsertUser` +
    `GetUserAuthByEmail` con permisos del rol; duplicado de correo normalizado → error de `UNIQUE`
    (Q5); `UpdateUserLastLogin` escribe las dos columnas y ningún DTO de entrada las toca;
    `UpdateUser` con rol inexistente → error de FK; inexistente → `pgx.ErrNoRows` mapeable a 404.
  - **Criterio de terminado**: `go test -tags=integration ./internal/usuarios/` en verde; el
    repositorio no contiene reglas de negocio (solo SQL y mapeo).
  - **Commit sugerido**: `feat(usuarios): repositorio de cuentas sobre sqlc`

- [ ] T222 · `usuarios`: `repository_roles.go` y guard anti-bloqueo · `[backend]` `[P4]`

  - **Archivos**: `backend/internal/usuarios/repository_roles.go`,
    `repository_roles_integration_test.go` (`//go:build integration`; NUEVOS).
  - **Qué hace**: roles y permisos (`InsertRole`, `GetRoleByID`, `GetRoleByNameLower`, `ListRoles`/
    `CountRoles`, `UpdateRoleName`, `DeleteRolePermissions`+`InsertRolePermission` en transacción,
    `DeleteRole`, `CountRoleUsers`, `ListPermissions`, `GetPermissionIDsByCodes`) y el **guard
    anti-bloqueo FR-008** (P7): dentro de `database.WithTx`, `pg_advisory_xact_lock` (clave fija,
    `LockAdminGuard`) + recuento **post-mutación** de cuentas activas con `admin_usuarios_roles`
    (`CountActiveAdmins`); 0 → `apperr.Conflict` con `details.reason="admin_required"` y rollback.
  - **Pruebas incluidas** (§III): **integración** (T219): `GetRoleByNameLower` distingue nombres que
    solo difieren en espacios pero no en mayúsculas (Q5); `DeleteRole` con usuarios → error por
    `RESTRICT` (FR-017); **carrera del guard**: dos transacciones concurrentes que intentan dejar 0
    administradores → exactamente una completa y queda ≥ 1 administrador activo (SC-004).
  - **Criterio de terminado**: el estado prohibido "panel sin administración" es imposible incluso
    con dos administradores actuando a la vez (prueba concurrente en verde).
  - **Commit sugerido**: `feat(usuarios): repositorio de roles con guard anti-bloqueo transaccional`

- [ ] T223 · `usuarios`: `repository_audit.go` · `[backend]` `[P4]`

  - **Archivos**: `backend/internal/usuarios/repository_audit.go`,
    `repository_audit_integration_test.go` (`//go:build integration`; NUEVOS).
  - **Qué hace**: inserción y consulta del registro: `InsertLoginEvent`, `InsertAdminAction`,
    `ListLoginEvents`/`CountLoginEvents` y `ListAdminActions`/`CountAdminActions` con los filtros
    `userId`/`from`/`to` (semirango `[from, to)`) y `LIMIT`/`OFFSET` (P22); los listados llevan
    `LEFT JOIN users` para **derivar** `userEmail`/`userName` y `actorEmail`/`actorName` (F-01: no se
    guardan como dato duplicado). **Solo `INSERT`/`SELECT`**
    (FR-025). El correo de un intento no identificado **no se guarda** (mínimo dato, FR-026).
  - **Pruebas incluidas** (§III): **integración** (T219): cada desenlace inserta su fila (incluido el
    correo inexistente con `user_id` NULL y **sin crear nada**, US8 esc. 7); los filtros por cuenta y
    rango de fechas y la paginación devuelven lo esperado; en `/acciones` el filtro por cuenta
    incluye actor **y** objetivo; eliminar un rol deja `target_role_id` NULL y `target_label` intacto
    (FR-025); `grep -E "UPDATE|DELETE"` sobre el archivo → 0; coherencia FR-021: tras un login, el
    `last_login_*` de `users` coincide con la última fila `success` de `login_events`.
  - **Criterio de terminado**: el registro queda duradero, filtrable y estructuralmente de solo
    inserción; `go test -tags=integration ./internal/usuarios/` en verde.
  - **Commit sugerido**: `feat(usuarios): repositorio de auditoría (inserción y consulta)`

- [ ] T224 · `usuarios`: `service_audit.go` — registro de accesos y acciones · `[backend]`

  - **Archivos**: `backend/internal/usuarios/service_audit.go`, `service_audit_test.go` (NUEVOS).
  - **Qué hace**: el servicio de registro que consumen el resto de servicios y el middleware:
    `RecordLoginEvent` (todo intento de acceso: éxito, fallo, cuenta inactiva, intento durante el
    bloqueo — FR-022) y `RecordAction` (cada acción administrativa sensible con quién/qué/sobre
    qué/cuándo/resultado — FR-023, **también cuando falla o se deniega**), e implementación de la
    interfaz `audit.Recorder` (T217) para que `authz` registre denegaciones. **Éxito de una mutación
    y su registro van en la misma transacción** (o ambos o ninguno); el registro de un intento que ya
    va a fallar es *best-effort* con error en log (`request_id`) y **nunca cambia la respuesta**
    (P20, `research.md` R23). **Nunca se registra una contraseña ni credenciales** (FR-026): de un
    restablecimiento, solo actor/objetivo/fecha.
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: cada desenlace de login deja su fila con el `result`
    correcto y la IP; una acción sensible deja `admin_actions` con éxito **y** con fallo **y** con
    `denied`; un fallo del repositorio de registro no altera el error devuelto al cliente (se loguea);
    ninguna prueba observa una contraseña en el registro (FR-026).
  - **Criterio de terminado**: los tres puntos de escritura de P20 tienen su mecanismo (el de
    `service_users`/`service_roles` se conecta en T231–T236; el de `authz`/`handler` en T239).
  - **Commit sugerido**: `feat(usuarios): registro de accesos y acciones administrativas`

- [ ] T225 · `usuarios`: `service_auth.go` — login, logout y sesión · `[backend]`

  - **Archivos**: `backend/internal/usuarios/service_auth.go`, `service_auth_test.go` (NUEVOS).
  - **Qué hace**: `Login` (FR-002/FR-003/US1): normaliza el correo (Q5), comprueba el bloqueo FR-006
    (**5 fallos / 15 min** por identificador normalizado exista o no la cuenta, P6 — mensaje
    idéntico en ambos casos; **semántica del 5.º intento, F-03**: el contador se incrementa con cada
    fallo, el **5.º fallo** responde el `401` genérico **y crea el bloqueo**, y el `429` aplica
    **desde el 6.º** intento dentro de los 15 min), verifica con bcrypt **con verificación dummy para tiempos uniformes**
    (R18), rechaza cuentas inactivas con el mensaje de "acceso desactivado" (US1 esc. 3) y, en éxito,
    limpia el contador, escribe `login_events` + `last_login_*` (FR-021/FR-022; **sin registro, sin
    acceso**), crea la sesión en Redis (P1/P9) y emite las cookies; `Logout` (FR-004): borra la sesión
    y las cookies; `Resolve` (implementa `session.Resolver`): resuelve `Identity` con permisos del rol
    **por petición** (FR-018/SC-009) y exige cuenta activa (FR-012). Todo desenlace queda registrado
    vía `service_audit` (FR-022). Códigos: credenciales incorrectas o correo inexistente → el
    **mismo** `401` genérico (FR-003/SC-008), **también en el 5.º fallo** (que crea el bloqueo);
    desde el 6.º intento y durante el bloqueo → `429` con `Retry-After`.
  - **Pruebas incluidas** (§III): unitarias con fakes de repositorio, `session.Store` y contadores.
    **Cómo se verifica**: `go test ./internal/usuarios/` en verde con cobertura del service **≥ 80 %**
    (§III): éxito; credenciales incorrectas y correo inexistente producen idéntico mensaje y código;
    cuenta inactiva → `Forbidden` con su mensaje; el **5.º fallo** responde `401` genérico y crea el
    bloqueo, el **6.º** y los siguientes (dentro de los 15 min) → `429` con `Retry-After`, y pasados
    los 15 min se permite
    de nuevo; éxito limpia el contador; cada desenlace deja `login_events` con la IP; el mensaje de
    error nunca revela existencia (SC-008) y ninguna respuesta contiene la contraseña.
  - **Criterio de terminado**: US1 y FR-002…FR-006, FR-022 cubiertos en el service; `Logout` deja la
    sesión inutilizable en el store.
  - **Commit sugerido**: `feat(usuarios): login, logout y resolución de identidad con bloqueo FR-006`

- [ ] T226 · `usuarios`: cambio de contraseña propio y `mustChangePassword` · `[backend]`

  - **Archivos**: `backend/internal/usuarios/service_auth.go` (EDITADO) y `service_auth_test.go`.
  - **Qué hace**: `ChangeMyPassword` (FR-020/US7): exige la contraseña actual, aplica la política
    FR-010 con `platform/password` (también "distinta del nombre/apellidos/correo"), guarda el hash
    nuevo, pone `mustChangePassword=false`, **revoca las demás sesiones** de la cuenta (R17) y **no**
    se registra en `admin_actions` (no es acción administrativa, FR-023). El **guard de cambio
    obligatorio** (FR-010/US7 esc. 4–5): si `mustChangePassword`, solo se autorizan `/auth/session`,
    `/auth/logout` y `/auth/password`; el resto → `403` con `details.reason="password_change_required"`.
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: cambio válido aplica y permite entrar con la nueva;
    contraseña actual incorrecta → error claro; política incumplida (cada requisito) → `400` con el
    requisito nombrado; cuenta con `mustChangePassword` → el guard deniega lo demás con su `details`;
    **0 filas nuevas en `admin_actions`** por el cambio propio.
  - **Criterio de terminado**: FR-010 (flujos de la persona) y FR-020 cubiertos; las sesiones
    revocadas dejan de resolver en el store.
  - **Commit sugerido**: `feat(usuarios): cambio de contraseña propio con política y guard obligatorio`

- [ ] T227 · `platform/middleware`: authn, authz, CSRF, rate-limit y cadena · `[backend]`

  - **Archivos**: `backend/internal/platform/middleware/authn.go`, `authz.go`, `csrf.go`,
    `ratelimit.go`, `passwordguard.go`, `cors.go` (EDITADO), `chain.go` (EDITADO si aplica) y
    `middleware_test.go` (NUEVOS/EDITADOS).
  - **Qué hace**: (a) `authn` (P1/P9): lee `ss_session`, resuelve la sesión en Redis comprobando
    inactividad y vida absoluta, resuelve `Identity` vía `session.Resolver`, comprueba que la cuenta
    siga **activa** (FR-012, red de seguridad por petición) y la deja en el contexto; sin sesión →
    `401 unauthenticated`. (b) `passwordguard`: la cadena de cambio obligatorio (T226); se monta
    **solo en el grupo `/api/v1/admin`** (entre `authn` y `authz`), y las rutas
    `/api/v1/auth/session`, `/auth/logout` y `/auth/password` quedan **blanqueadas** (viven en el
    grupo `/api/v1/auth`, que no lo monta). (c)
    `AuthzByModule(código)` (P16/FR-016): exige el permiso del módulo; sin permiso → `403` con
    mensaje claro y **registro del intento** vía `audit.Recorder` (`result='denied'`, P20). (d)
    `CSRF` (P10): solo métodos no seguros; exige `X-CSRF-Token` = cookie `csrf_token` firmada. (e)
    `rate-limit` (P17): ventana deslizante **en memoria** por IP sobre `/auth/login` y
    `/setup/initialize` → `429 rate_limited`. (f) `CORS` ampliado (P11): `Allow-Credentials`, eco
    exacto del `Origin` permitido (nunca `*`), cabeceras `Content-Type, X-CSRF-Token, X-Request-ID`,
    `Vary: Origin`. Cadena del grupo de panel: `authn → passwordguard → authz → CSRF` (plan §"Cadena
    de middleware"). `/healthz` **no cambia**.
  - **Pruebas incluidas** (§III): unitarias + `httptest`: `authn` (sin cookie, cookie inválida, sesión
    expirada, cuenta inactiva → `401`); `authz` con/sin permiso (`403` + registro `denied`);
    `csrf` (método seguro pasa; inseguro sin token o desalineado → `403`); `ratelimit` (tras el
    umbral → `429`); CORS (preflight con credenciales y eco de origen; origen no permitido sin
    cabeceras); el orden de la cadena es el aprobado.
  - **Criterio de terminado**: toda operación de panel exige sesión + CSRF + permiso **en el servidor**
    (§IV, CWE-862); SC-007 verificable por estas pruebas.
  - **Commit sugerido**: `feat(platform): middleware authn, authz, CSRF, rate-limit y CORS con credenciales`

- [ ] T228 · `usuarios`: handlers de acceso y cableado · `[backend]`

  - **Archivos**: `backend/internal/usuarios/handler.go`, `handler_auth.go`, `routes.go`
    (`RegisterPublic`), `handler_auth_test.go` (NUEVOS) y `backend/cmd/api/main.go`,
    `main_test.go` (EDITADOS).
  - **Qué hace**: `handler.go` con los helpers HTTP (decodificación + `platform/validate`; el helper
    que rechaza un JSON/DTO inválido **registra el intento**, P20 — se conecta del todo en T239);
    `handler_auth.go` con `POST /api/v1/auth/login`, `POST /api/v1/auth/logout`,
    `GET /api/v1/auth/session` y `POST /api/v1/auth/password` según el contrato (T201); `routes.go`
    publica `/api/v1/auth` (login con `rate-limit`; el resto con `authn → CSRF`, **sin guard**: son
    las rutas blanqueadas `/auth/session`, `/auth/logout` y `/auth/password` que necesita una cuenta
    con `mustChangePassword`) y deja listo
    `RegisterAdmin` **con el guard de cambio obligatorio montado solo en el grupo `/api/v1/admin`**
    (entre `authn` y `authz`; F-15: ahí y no en `/auth/*`). `main.go`: DI manual con el **cliente
    Redis**, `session.Store`, el resolver del
    dominio, `audit.Recorder` y los grupos de la cadena aprobada; `main_test.go` amplía el humo
    (rutas nuevas con su sobre y `/healthz` intacto, §8.1.9).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/usuarios/ ./cmd/api/` en verde: cada operación con sus códigos
    200/201/400/401/403/429 y su sobre (éxito = DTO directo; error = `ErrorEnvelope`); `details` por
    campo en validaciones; **ninguna respuesta contiene `password` ni `passwordHash`** (FR-003);
    una cuenta con `mustChangePassword` recibe `403` con
    `details.reason="password_change_required"` en `/api/v1/admin/*` y en cambio `/auth/session`,
    `/auth/logout` y `/auth/password` le responden (rutas blanqueadas);
    logout borra cookies; el humo de `cmd/api` publica las rutas y `/healthz` responde igual.
  - **Criterio de terminado**: US1 es usable de punta a punta por API (`quickstart.md` §2–§3); la
    superficie pública queda auditable de un vistazo en `routes.go`.
  - **Commit sugerido**: `feat(usuarios): handlers de acceso y cableado con Redis en cmd/api`

---

## Fase 6 — Inicialización única del administrador (US2, FR-007)

- [ ] T229 · `usuarios`: inicialización única · `[backend]`

  - **Archivos**: `backend/internal/usuarios/service_auth.go` (EDITADO) o `service_init.go` (NUEVO),
    `service_init_test.go` y `repository_*_integration_test.go` (ampliados).
  - **Qué hace**: `Initialize` (FR-007/US2, P8): en **una transacción** con `LockAdminGuard` (advisory
    lock) comprueba que **no exista ninguna cuenta** (`CountUsers = 0`), crea el rol **"Administrador"**
    con los **9 permisos** del catálogo y la cuenta inicial (nombre, apellidos, correo, teléfono y
    contraseña con política FR-010; `mustChangePassword=false`: ya la eligió quien inicializa) y deja
    el registro `admin_actions` **`user.create` sin actor** (único caso permitido, FR-023). Repetirla →
    `409` ("ya se hizo y no puede repetirse"). La validación de la cabecera `X-Setup-Token` =
    `BOOTSTRAP_TOKEN` la hace el handler (T230).
  - **Pruebas incluidas** (§III): unitarias (con `users` vacío crea todo; con cuentas existentes →
    `409`; datos inválidos → `400` sin crear nada) + **integración** (T219): **dos `Initialize`
    simultáneos → exactamente uno crea** (SC-003); la cuenta creada puede crear usuarios y roles de
    inmediato (rol con los 9 permisos); el registro `user.create` sin actor queda insertado.
  - **Criterio de terminado**: SC-003 verificado (100 % de instalaciones nuevas inicializadas una
    sola vez; repeticiones bloqueadas).
  - **Commit sugerido**: `feat(usuarios): inicialización única del administrador inicial`

- [ ] T230 · `usuarios`: `POST /api/v1/setup/initialize` y rate-limit · `[backend]`

  - **Archivos**: `backend/internal/usuarios/handler_auth.go` (EDITADO), `routes.go` (EDITADO),
    `handler_auth_test.go` (ampliado) y `backend/cmd/api/main.go` (EDITADO si aplica).
  - **Qué hace**: publica `POST /api/v1/setup/initialize` según el contrato: exige la cabecera
    `X-Setup-Token` (distinta o ausente → `401`/`403` sin revelar el valor esperado), valida el DTO
    (`InitializeInput`), llama al servicio (T229) y responde `201` con `UserItem`. Monta el grupo
    `/api/v1/setup` y `POST /api/v1/auth/login` con **`rate-limit`** (P17).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: `201` con el sobre correcto; sin token o con token
    erróneo → `401`/`403`; repetida → `409` con el mensaje de la spec; DTO inválido → `400` con
    `details`; tras el umbral de la IP → `429`; ninguna respuesta devuelve la contraseña (FR-003).
  - **Criterio de terminado**: `quickstart.md` §1 es ejecutable tal cual (incluida la repetición
    bloqueada, SC-003) y la superficie pública escribible va rate-limited (P17).
  - **Commit sugerido**: `feat(usuarios): endpoint de inicialización única con X-Setup-Token y rate-limit`

---

## Fase 7 — Gestión de cuentas (US3, US5, US7 parcial)

- [ ] T231 · `usuarios`: `service_users.go` — crear cuenta · `[backend]`

  - **Archivos**: `backend/internal/usuarios/service_users.go`, `service_users_test.go` (NUEVOS).
  - **Qué hace**: `CreateUser` (FR-009/US3): valida con `platform/validate` (nombre, apellidos, correo
    y teléfono **obligatorios**; teléfono con formato telefónico razonable → `details.phone`), exige
    un **rol existente** (`details.roleId`, US3 esc. 5), normaliza el correo (trim + minúsculas, Q5)
    y rechaza duplicados con `409` ("ese correo ya está en uso"), aplica la política FR-010 a la
    contraseña inicial definida por el administrador (quedará `mustChangePassword=true`), crea la
    cuenta **activa** (US3 esc. 1) y registra `admin_actions` **`user.create`** con actor y objetivo
    (FR-023) en la misma transacción. Sin permiso → lo deniega `authz` (T227).
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: cuenta válida creada activa con `mustChangePassword`;
    correo duplicado **también con otras mayúsculas o espacios** → `409` sin crear (SC-011); cada
    campo inválido/incompleto → `400` con `details` del campo; rol inexistente → `400`
    `details.roleId`; contraseña que incumple la política → `400` con el requisito; **una fila
    `user.create`** en el registro con actor, objetivo y `result='success'`; una creación fallida deja
    su fila con `result='failure'`.
  - **Criterio de terminado**: US3 cubierta en el service; nunca se devuelve la contraseña (FR-003).
  - **Commit sugerido**: `feat(usuarios): creación de cuentas con validación y registro`

- [ ] T232 · `usuarios`: `service_users.go` — editar y activar/desactivar · `[backend]`

  - **Archivos**: `backend/internal/usuarios/service_users.go` (EDITADO) y `service_users_test.go`.
  - **Qué hace**: `UpdateUser` (FR-011/US5): edita nombre, apellidos, correo, teléfono, **rol**
    (el nuevo **reemplaza** al anterior: un solo rol, Q4) y **estado** `isActive`, con las mismas
    validaciones que al crear. Al **desactivar** (FR-012/US5 esc. 1–2): revoca todas las sesiones de
    la cuenta (`user_sessions:*`, P9) para cortar el acceso **de inmediato**, conserva todos los datos
    (FR-013: **no existe operación de borrado**) y registra `user.deactivate`; reactivar registra
    `user.activate`; la edición registra `user.update` (FR-023). Toda mutación aplica el **guard
    anti-bloqueo** (FR-008, P7): desactivar o cambiar el rol de forma que el panel se quede sin una
    cuenta activa con `admin_usuarios_roles` → `409` con `details.reason="admin_required"` (US2
    esc. 4–6).
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: desactivar revoca las sesiones (el `store` falso lo
    observa) y conserva los datos; reactivar restaura el acceso; cambiar de rol reemplaza (nunca
    suma); la regla del último administrador bloquea desactivación **y** cambio de rol con `409`
    explicativo (SC-004); cada desenlace deja su fila en `admin_actions` (también los fallidos).
  - **Criterio de terminado**: US5 y FR-011/FR-012/FR-013 cubiertos; SC-006 verificable (cuenta
    desactivada sin acceso ni con sesión abierta).
  - **Commit sugerido**: `feat(usuarios): edición y activación/desactivación con guard anti-bloqueo`

- [ ] T233 · `usuarios`: `service_users.go` — restablecer contraseña · `[backend]`

  - **Archivos**: `backend/internal/usuarios/service_users.go` (EDITADO) y `service_users_test.go`.
  - **Qué hace**: `ResetUserPassword` (FR-010/US7 esc. 5): el administrador define una contraseña
    nueva que **debe cumplir la política** (aplicada también a las restablecidas), la guarda con
    bcrypt, marca `mustChangePassword=true` (su titular debe cambiarla al entrar) y **revoca todas
    las sesiones** de la cuenta (R17). Registra **`user.password_reset`** con actor, objetivo y
    fecha — **nunca la contraseña** (FR-026).
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: restablecimiento válido → hash nuevo + flag + sesiones
    revocadas; política incumplida → `400` con el requisito; cuenta inexistente → `404`; el registro
    contiene `user.password_reset` **sin ningún valor de contraseña** (FR-026); una operación
    fallida también se registra (FR-023).
  - **Criterio de terminado**: la recuperación de credenciales del MVP (decisión Q2) queda cubierta
    sin servicio de correo (Out of Scope respetado).
  - **Commit sugerido**: `feat(usuarios): restablecimiento de contraseña con registro sin credenciales`

- [ ] T234 · `usuarios`: `handler_users.go` y rutas de cuentas · `[backend]`

  - **Archivos**: `backend/internal/usuarios/handler_users.go`, `handler_users_test.go` (NUEVOS),
    `routes.go` (EDITADO: `RegisterAdmin`) y `backend/cmd/api/main.go` (EDITADO).
  - **Qué hace**: `GET /api/v1/admin/usuarios` (listado paginado con `platform/paginate`, P14:
    `UserItem` con estado, correo, rol **y `lastLoginAt`/`lastLoginIp`**, FR-019/FR-021),
    `POST /api/v1/admin/usuarios`, `GET /api/v1/admin/usuarios/{id}` (ficha con último acceso;
    cuenta que nunca entró → `lastLoginAt: null`, US8 esc. 6), `PATCH /api/v1/admin/usuarios/{id}` y
    `POST /api/v1/admin/usuarios/{id}/password`. Todo bajo `authn → passwordguard → authz(admin_usuarios_roles)
    → CSRF` (T227). **No se registra `DELETE /usuarios/{id}` en ninguna parte** (FR-013).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: cada operación con sus códigos
    200/201/400/401/403/404/409 y su sobre; listado con `limit`/`offset` acotados (tope 100);
    `UserItem` incluye `lastLoginAt`/`lastLoginIp` anulables; ninguna respuesta contiene credenciales;
    **0 rutas de borrado** de cuentas (búsqueda en `routes.go`); sin permiso → `403` con mensaje claro
    (SC-007).
  - **Criterio de terminado**: `quickstart.md` §4, §6 y §7 son ejecutables por API (SC-005, SC-006).
  - **Commit sugerido**: `feat(usuarios): handlers de gestión de cuentas sobre /api/v1/admin/usuarios`

---

## Fase 8 — Gestión de roles y permisos (US4, US6)

- [ ] T235 · `usuarios`: `service_roles.go` — crear rol y catálogo · `[backend]`

  - **Archivos**: `backend/internal/usuarios/service_roles.go`, `service_roles_test.go` (NUEVOS).
  - **Qué hace**: `CreateRole` (FR-014/US4): nombre único **normalizado** (trim + colapso de espacios;
    unicidad sin distinguir mayúsculas → duplicado "casi igual" = `409`, Q5/SC-011) y **al menos un
    permiso** del catálogo (`400` "un rol debe tener al menos un permiso", US4 esc. 4); los permisos
    se combinan **libremente** (sin catálogo fijo de roles, US4 esc. 2) y deben existir en
    `permissions` (FR-015). `ListPermissions` expone el catálogo con etiquetas (los módulos de
    F3–F9 aparecen reservados). Registra **`role.create`** (FR-023) en la misma transacción.
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: rol con dos permisos combinados libremente se crea;
    nombre duplicado con otras mayúsculas/espacios → `409` sin crear (SC-011); sin permisos → `400`;
    permiso inexistente → `400` con `details`; `ListPermissions` devuelve los 9 códigos con su
    etiqueta; el registro `role.create` queda con actor/objetivo/resultado.
  - **Criterio de terminado**: US4 cubierta en el service; el catálogo de FR-015 se lee de la base
    (T207), no se duplica en código.
  - **Commit sugerido**: `feat(usuarios): creación de roles con permisos por módulo y catálogo`

- [ ] T236 · `usuarios`: `service_roles.go` — editar y eliminar rol · `[backend]`

  - **Archivos**: `backend/internal/usuarios/service_roles.go` (EDITADO) y `service_roles_test.go`.
  - **Qué hace**: `UpdateRole` (FR-017/FR-018/US6): cambia nombre (mismas reglas de unicidad) y/o
    permisos (`DELETE`+`INSERT` de `role_permissions` en transacción; **nunca sin permisos**, US6
    esc. 6); los cambios se reflejan **de inmediato** en las cuentas con ese rol porque los permisos
    se resuelven por petición (FR-018/SC-009). `DeleteRole` (FR-017/US6 esc. 4–5): solo cuando
    **ninguna cuenta** lo tiene (`CountRoleUsers = 0`); si hay cuentas → `409` explicando que primero
    deben reasignarse. Toda mutación aplica el guard anti-bloqueo (FR-008: quitar
    `admin_usuarios_roles` al único rol que lo tiene → `409`). Registra `role.update`/`role.delete`
    (FR-023); al eliminar, `target_role_id` queda NULL y `target_label` conserva el nombre (FR-025).
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: quitar/agregar un permiso cambia los permisos efectivos
    de las cuentas (identidad resuelta por petición); renombrar conserva cuentas y permisos; rol en
    uso → `409` con el mensaje que pide la spec (SC-011); dejar el rol sin permisos → `400`; regla
    anti-bloqueo aplicada a la edición del rol; los registros `role.update`/`role.delete` quedan con
    su resultado.
  - **Criterio de terminado**: US6 y FR-014/FR-017/FR-018 cubiertos; 0 eliminaciones de roles en uso
    posibles.
  - **Commit sugerido**: `feat(usuarios): edición y eliminación de roles con reglas de uso`

- [ ] T237 · `usuarios`: `handler_roles.go` y rutas de roles · `[backend]`

  - **Archivos**: `backend/internal/usuarios/handler_roles.go`, `handler_roles_test.go` (NUEVOS),
    `routes.go` (EDITADO) y `backend/cmd/api/main.go` (EDITADO).
  - **Qué hace**: `GET /api/v1/admin/roles` (paginado; `RoleItem` con `userCount` para saber si se
    puede eliminar), `POST /api/v1/admin/roles`, `GET /api/v1/admin/roles/{id}`,
    `PATCH /api/v1/admin/roles/{id}`, `DELETE /api/v1/admin/roles/{id}` y
    `GET /api/v1/admin/permisos` (catálogo, FR-015). Bajo el grupo de panel (sesión + CSRF + permiso
    `admin_usuarios_roles`, FR-016).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: cada operación con sus códigos
    200/201/400/401/403/404/409 y su sobre; `RoleItem` incluye `permissions` y `userCount`;
    `DELETE` sobre rol en uso → `409` con `details`; sin permiso → `403` (SC-007); CSRF ausente en
    `POST`/`PATCH`/`DELETE` → `403`.
  - **Criterio de terminado**: `quickstart.md` §4 y §6 son ejecutables por API (SC-005, SC-009).
  - **Commit sugerido**: `feat(usuarios): handlers de roles y catálogo de permisos`

---

## Fase 9 — Auditoría: puntos de escritura restantes y consulta (US8, FR-021…FR-026)

> El registro de cada desenlace de login y de cada operación sensible ya se escribe desde su
> productor (T224–T236, P20). Esta fase cierra los dos puntos transversales restantes y toda la
> consulta de solo lectura.

- [ ] T238 · `usuarios`: `service_audit.go` — consulta con filtros y paginación · `[backend]`

  - **Archivos**: `backend/internal/usuarios/service_audit.go` (EDITADO) y `service_audit_test.go`.
  - **Qué hace**: `ListAccessEvents` y `ListAdminActions` (FR-024/US8, P22): filtros `userId` (en
    `/acciones`, la cuenta **involucrada**: actor o objetivo) y rango `from`/`to` (**semirango
    `[from, to)`**), paginación `platform/paginate` (defecto 20, tope 100, orden `createdAt DESC`);
    `from > to` o fechas mal formadas → `400 invalid`. Los registros sin cuenta asociada solo
    aparecen cuando **no** se filtra por `userId`. **Sin ninguna operación de escritura** (FR-025).
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: filtros combinados devuelven solo lo esperado; `from`
    posterior a `to` → `400`; la paginación respeta `limit`/`offset` y devuelve `total`; la página
    vacía es `items: []` (nunca `null`); ningún método del service modifica ni borra registros.
  - **Criterio de terminado**: SC-012 verificable (encontrar accesos y acciones de una cuenta con
    filtros por cuenta y fechas).
  - **Commit sugerido**: `feat(usuarios): consulta de auditoría con filtros y paginación`

- [ ] T239 · Auditoría: denegaciones y datos inválidos registrados · `[backend]`

  - **Archivos**: `backend/internal/platform/middleware/authz.go` (EDITADO),
    `backend/internal/usuarios/handler.go` (EDITADO), `handler_test.go`/`middleware_test.go`
    (ampliados) y `backend/internal/usuarios/service_audit.go` (EDITADO si aplica).
  - **Qué hace**: conecta los dos puntos de escritura restantes de P20: (1) `authz` registra toda
    denegación (`result='denied'`) resolviendo acción y objetivo desde `method`+`path` con la tabla
    del dominio vía `audit.Recorder`; (2) el helper común de decodificación/validación del `handler`
    registra el intento rechazado (`result='failure'`, JSON inválido o DTO no válido) con el objetivo
    que se pueda identificar. En ambos casos el registro es *best-effort*: su fallo no cambia la
    respuesta (R23) y **nunca se guarda el cuerpo con credenciales** (FR-026).
  - **Pruebas incluidas** (§III): `httptest` + fakes. **Cómo se verifica**:
    `go test ./internal/platform/middleware/ ./internal/usuarios/` en verde: una operación denegada
    deja su fila `denied` con actor y objetivo; un JSON inválido deja su fila `failure`; un fallo del
    registro no altera el `403`/`400` devuelto (solo log con `request_id`); ninguna fila contiene el
    cuerpo de la petición ni contraseña alguna (FR-026).
  - **Criterio de terminado**: los tres puntos de escritura de P20 están cubiertos y probados (R19:
    "un punto olvidado dejaría un intento sin registrar").
  - **Commit sugerido**: `feat(usuarios): registrar denegaciones y datos inválidos en admin_actions`

- [ ] T240 · `usuarios`: `handler_audit.go` y rutas de auditoría · `[backend]`

  - **Archivos**: `backend/internal/usuarios/handler_audit.go`, `handler_audit_test.go` (NUEVOS),
    `routes.go` (EDITADO) y `backend/cmd/api/main.go` (EDITADO).
  - **Qué hace**: **solo** `GET /api/v1/admin/auditoria/accesos` y
    `GET /api/v1/admin/auditoria/acciones` (FR-024/FR-025, P22) bajo el permiso
    `admin_usuarios_roles` (misma decisión que la gestión: sin él → `403` con mensaje claro, US8
    esc. 5), con parámetros `userId`/`from`/`to`/`limit`/`offset`, `AccessEventItem` (con `userId`,
    `userEmail` y `userName` anulables y **derivados por `JOIN` con `users`**: un intento contra un
    correo inexistente queda **sin asociación** y **sin guardar ni mostrar correo** —la UI muestra
    "Intento sin cuenta asociada"—, US8 esc. 7, F-01) y `AdminActionItem` (actor con
    `actorId`/`actorEmail`/`actorName` derivados igualmente, acción, objetivo con `targetLabel`
    preservado, resultado).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/usuarios/` en verde: ambos historiales responden `200` con
    `{items,total,limit,offset}`; filtros aplicados; `from > to` → `400`; sin permiso → `403`
    (SC-012); **0 operaciones de escritura** sobre `/admin/auditoria` en `routes.go` y en el contrato
    (FR-025/SC-013); ningún DTO de salida contiene credenciales (FR-026).
  - **Criterio de terminado**: `quickstart.md` §10 es ejecutable por API; el registro es de solo
    lectura por construcción.
  - **Commit sugerido**: `feat(usuarios): handlers de solo lectura del registro de auditoría`

---

## Fase 10 — Frontend (pantallas y guards)

> Convenciones de la skill `react-frontend` y de `ux.md` (textos, estados y accesibilidad WCAG 2.1
> AA). Nomenclatura de rutas según `plan.md` P18 + `quickstart.md` (ver "Huecos…", punto 1). Estado
> del servidor vía TanStack Query; formularios con React Hook Form + Zod; sin `fetch` fuera de
> `src/api/client.ts`; sin tokens en `localStorage` (CWE-79).

- [ ] T241 · `api/client.ts` con credenciales/CSRF y `api/auth.ts` · `[frontend]` `[P5]`

  - **Archivos**: `frontend/src/api/client.ts` (EDITADO), `frontend/src/api/auth.ts`,
    `client.test.ts`/`auth.test.ts` (NUEVOS/EDITADOS).
  - **Qué hace**: amplía `apiFetch<T>` con `credentials: "include"` y la cabecera `X-CSRF-Token`
    leída de la cookie `csrf_token` en todo método no seguro (P10); `ApiError` conserva
    `code`/`message`/`details` (incluidos `retryAfterSeconds`, `reason`). `auth.ts` expone
    `login()`, `logout()`, `getSession()` y `changeMyPassword()` tipados con `schema.d.ts` (T204).
    *Cubre FR-001, FR-003, FR-004, FR-020 (capa de cliente)*.
  - **Pruebas incluidas** (§III): Vitest + MSW. **Cómo se verifica**: `npm test -- --run` en verde:
    200 → DTO; 401/403/409/429 → `ApiError` con `code` y `details` intactos; toda petición no segura
    lleva `credentials` y `X-CSRF-Token`; búsqueda de `fetch(` fuera de `client.ts` → 0 coincidencias.
  - **Criterio de terminado**: único mecanismo de acceso a la API (regla de oro de la skill);
    `npm run typecheck` en verde.
  - **Commit sugerido**: `feat(frontend): cliente con credenciales y CSRF, y API de autenticación`

- [ ] T242 · Componentes compartidos y helpers de `lib/` · `[frontend]` `[P5]`

  - **Archivos**: `frontend/src/components/` (NUEVOS, **nombres de código en inglés**):
    `Table.tsx`, `Tabs.tsx`, `Pagination.tsx`, `DateRangeFilter.tsx`, `StatusPill.tsx`, `Notice.tsx`,
    `ModalDialog.tsx`, `Field.tsx`, `ConfirmDialog.tsx`, `EmptyState.tsx` (+ tests) y
    `frontend/src/lib/permissions.ts`, `format.ts` (+ tests) (NUEVOS).
  - **Qué hace**: el inventario de componentes genéricos que necesita F2 (F-10), sin duplicar markup
    en las features: **`Table`** (tabla accesible),
    **`Tabs`** (los dos historiales de auditoría), **`Pagination`** (`limit`/`offset`, conserva los
    filtros), **`DateRangeFilter`** (`from`/`to`), **`StatusPill`** (estado de cuenta y resultado de
    registro: texto además de color), **`Notice`** (avisos con `aria-live`), **`Field`** (campo con
    etiqueta, error y `autocomplete`), **`ModalDialog`** (diálogo accesible),
    **`ConfirmDialog`** (confirmación accesible) y **`EmptyState`** (estado vacío comprensible —
    p. ej. "no hay registros que coincidan con esos filtros"). Los nombres de código son en inglés;
    **`ux.md` (disenador-ux) da el mapeo** a sus etiquetas en español y los textos. Helpers:
    `hasPermission(session, code)` y formato de fecha/hora en español (para `lastLoginAt` y los
    historiales). Accesibilidad: `getByRole`, foco visible, `aria-live` para avisos (§6).
  - **Pruebas incluidas** (§III): Vitest + Testing Library. **Cómo se verifica**:
    `npm test -- --run` en verde: cada componente renderiza sus estados (con/sin error, vacío,
    paginación, pestañas) y los diálogos exponen roles accesibles; `hasPermission` cubre con/sin
    permiso y cuenta sin permisos de módulo (Edge Case válido); `format` formatea fechas y el caso
    "sin accesos".
  - **Criterio de terminado**: las features de la Fase 10 (T244–T250, T254) reutilizan estos
    componentes (sin duplicar markup; lo revisa `revisor-codigo`) y `ux.md` documenta su mapeo de
    nombres.
  - **Commit sugerido**: `feat(frontend): componentes compartidos y helpers de permisos y formato`

- [ ] T243 · `api/usuarios.ts`, `api/roles.ts`, `api/auditoria.ts` · `[frontend]`

  - **Archivos**: `frontend/src/api/usuarios.ts`, `roles.ts`, `auditoria.ts` y sus tests (NUEVOS).
  - **Qué hace**: las funciones tipadas de cada dominio según el contrato: usuarios (`listUsers`,
    `getUser`, `createUser`, `updateUser`, `resetUserPassword`), roles (`listRoles`, `getRole`,
    `createRole`, `updateRole`, `deleteRole`, `listPermissions`) y auditoría (`listAccessEvents`,
    `listAdminActions` con `userId`/`from`/`to`/`limit`/`offset`). **Solo lectura en auditoría**
    (FR-025): no se escribe ninguna función de escritura sobre el registro.
  - **Pruebas incluidas** (§III): Vitest + MSW. **Cómo se verifica**: `npm test -- --run` en verde:
    cada función construye la URL con sus parámetros, tipa su DTO de salida y propaga `ApiError`;
    búsqueda de métodos `POST`/`PATCH`/`DELETE` en `auditoria.ts` → 0 (FR-025).
  - **Criterio de terminado**: las features consumen estas funciones (sin `fetch` propio).
  - **Commit sugerido**: `feat(frontend): APIs tipadas de usuarios, roles y auditoría`

- [ ] T244 · Guards, router y layout con navegación por permisos · `[frontend]`

  - **Archivos**: `frontend/src/app/router.tsx` (EDITADO), `frontend/src/app/guards.tsx`,
    `frontend/src/app/layout.tsx` (EDITADO) y sus tests (NUEVOS/EDITADOS).
  - **Qué hace**: rutas `/login`, `/cambiar-contrasena`, `/sin-permiso`, `/panel`, `/panel/usuarios`,
    `/panel/roles`, `/panel/auditoria` (P18 + `ux.md` §2; ver "Huecos…", punto 1) con los guards
    `RequireAuth` (sin sesión o `401` → `/login` con `destino`, FR-001/SC-001),
    `RequirePermission(code)` (sin permiso → `/sin-permiso`, sin adivinar nada) y
    `RequirePasswordChange` (`mustChangePassword` → cambio obligatorio). Layout del panel (ux §3.3):
    navegación **filtrada por permisos** (Usuarios/Roles/Auditoría solo con
    `admin_usuarios_roles`; una cuenta sin permisos de módulo ve solo Inicio), "Salir" y menú móvil.
    Catch-all → "Página no encontrada".
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: sin sesión, cualquier ruta del panel redirige a `/login`
    (SC-001); con permiso, las secciones aparecen; sin permiso, no aparecen y `/panel/usuarios`
    muestra `/sin-permiso`; `mustChangePassword` redirige al cambio obligatorio; ningún componente
    hace `fetch` directo.
  - **Criterio de terminado**: el panel es 100 % inaccesible sin sesión y la navegación solo muestra
    lo autorizado (US4 esc. 6; la autoridad real sigue siendo el servidor).
  - **Commit sugerido**: `feat(frontend): guards de sesión, permiso y cambio de contraseña con navegación filtrada`

- [ ] T245 · `features/auth`: pantalla de acceso y hooks · `[frontend]`

  - **Archivos**: `frontend/src/features/auth/pages/LoginPage.tsx`, `hooks/useSession.ts`,
    `useLogin.ts`, `useLogout.ts` y `auth.test.tsx` (NUEVOS).
  - **Qué hace**: la pantalla de acceso de `ux.md` §3.1 (correo y contraseña con mostrar/ocultar,
    `autocomplete` correcto, sin "regístrate" ni "olvidé mi contraseña" — Out of Scope) y los hooks
    de sesión vía TanStack Query. Muestra los mensajes de la spec: credenciales incorrectas → error
    **genérico** (FR-003/SC-008); cuenta desactivada → "ese acceso está desactivado" (US1 esc. 3);
    bloqueo por intentos → el mensaje de 15 minutos (FR-006); sesión expirada/cerrada → aviso que
    pide iniciar sesión de nuevo (FR-005). Tras entrar, redirige al `destino` o a `/panel` (SC-002).
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: éxito → navegación al panel; `401` → mensaje genérico idéntico para
    "correo inexistente" y "contraseña errónea" (SC-008); `403` de cuenta desactivada → su mensaje;
    `429` → mensaje de bloqueo con los minutos; accesibilidad (`getByRole("button")`, labels
    asociados).
  - **Criterio de terminado**: US1 usable en el navegador; los textos respetan `ux.md` §7.
  - **Commit sugerido**: `feat(frontend): pantalla de acceso con los estados de error de la spec`

- [ ] T246 · `features/auth`: cambio de contraseña y aviso de sesión · `[frontend]`

  - **Archivos**: `frontend/src/features/auth/pages/ChangePasswordPage.tsx`,
    `hooks/useChangePassword.ts`, `components/SessionWarning.tsx` y tests (NUEVOS).
  - **Qué hace**: la pantalla de cambio de contraseña (ux §3.9) en dos modos: **voluntario** (desde
    "Mi cuenta") y **obligatorio** (sin navegación, "Por seguridad, cambia esta contraseña antes de
    continuar", US7 esc. 4), con checklist de la política FR-010 en el formulario y mensajes del
    requisito incumplido (US7 esc. 3) y de contraseña actual incorrecta (US7 esc. 2). El
    `SessionWarning` global (ux §3.12) avisa 2 minutos antes del cierre por inactividad con "Sigo
    aquí" (FR-005) y, si la sesión termina, redirige a `/login` con su aviso.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: cambio válido → aviso de éxito y salida del modo obligatorio;
    `400` con `details` → el requisito se muestra; `401`/`403` → mensajes correctos; el modo
    obligatorio no muestra navegación; el aviso de sesión aparece en el umbral y "Sigo aquí" lo
    descarta.
  - **Criterio de terminado**: US7 y FR-005/FR-010/FR-020 cubiertos en la interfaz (SC-010).
  - **Commit sugerido**: `feat(frontend): cambio de contraseña con política y aviso de sesión`

- [ ] T247 · `features/usuarios`: listado de cuentas · `[frontend]` `[P6]`

  - **Archivos**: `frontend/src/features/usuarios/pages/UsersPage.tsx`, `hooks/useUsers.ts` y
    `usuarios.test.tsx` (NUEVOS).
  - **Qué hace**: el listado de cuentas (ux §3.5) con **estado (activo/inactivo), correo y rol**
    (FR-019) y el **último acceso en cada fila** (`lastLoginAt`/`lastLoginIp`, FR-021; la **ficha**
    de la cuenta es su detalle/edición y va en T248): la cuenta que
    nunca ha iniciado sesión indica "sin accesos" sin inventar fechas (US8 esc. 6). Tarjetas en
    móvil y tabla desde tableta (ux §4.b) con los componentes `Table`/`Pagination`/`StatusPill`/
    `EmptyState` (T242, sin duplicar markup); **paginación** con `limit`/`offset` y **sin buscador
    de texto libre** (F-04: fuera del MVP, no está en la spec); estado de carga y estado
    vacío.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: el listado muestra estado, correo, rol y último acceso en cada
    fila; la cuenta
    sin accesos muestra la indicación correcta; la paginación pasa de página conservando el estado
    y **no existe** ningún buscador de texto libre (F-04);
    accesibilidad de la tabla (`getByRole("table")`).
  - **Criterio de terminado**: FR-019 y FR-021 visibles en la interfaz.
  - **Commit sugerido**: `feat(frontend): listado de cuentas con estado, rol y último acceso`

- [ ] T248 · `features/usuarios`: formulario de cuenta y acciones · `[frontend]`

  - **Archivos**: `frontend/src/features/usuarios/components/UserForm.tsx`,
    `ResetPasswordDialog.tsx`, `hooks/useCreateUser.ts`, `useUpdateUser.ts`, `useResetPassword.ts`,
    `useSetUserActive.ts` y tests (NUEVOS).
  - **Qué hace**: la **ficha = detalle/edición** de la cuenta y sus acciones (ux §3.6) con React
    Hook Form + Zod: nombre, apellidos, correo,
    teléfono, rol y contraseña inicial (solo al crear); en la ficha se muestran también
    `lastLoginAt`/`lastLoginIp` (FR-021; la cuenta que nunca entró indica "aún no ha iniciado
    sesión"/"sin accesos", US8 esc. 6). Validación en cliente **y** manejo de los
    `details` del servidor (correo duplicado → "ya está en uso", teléfono → `details.phone`, rol
    inexistente → `details.roleId`, FR-009/US3 esc. 2–5). Acciones: **activar/desactivar** con
    `ConfirmDialog` (US5; el texto explica que los datos se conservan, FR-012/FR-013) y
    **restablecer contraseña** con la política (FR-010/US7 esc. 5). Tras crear, aviso "Cuenta creada"
    (ux §3.2).
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: envío válido llama a la API y refresca el listado; cada error del
    servidor se muestra en su campo (SC-011: duplicado "casi igual" no crea nada); la ficha muestra
    el último acceso y el caso "sin accesos"; desactivar pide
    confirmación y refleja el estado nuevo; restablecer exige la política; ningún campo de
    contraseña se reenvía ni se muestra después de guardar (FR-003/FR-026).
  - **Criterio de terminado**: US3/US5 y FR-009…FR-013 cubiertos en la interfaz.
  - **Commit sugerido**: `feat(frontend): alta, edición, activación y restablecimiento de cuentas`

- [ ] T249 · `features/roles`: listado y formulario de roles · `[frontend]` `[P6]`

  - **Archivos**: `frontend/src/features/roles/pages/RolesPage.tsx`, `components/RoleForm.tsx`,
    `hooks/useRoles.ts`, `useCreateRole.ts`, `useUpdateRole.ts`, `useDeleteRole.ts`,
    `usePermissions.ts` y `roles.test.tsx` (NUEVOS).
  - **Qué hace**: el listado y formulario de roles (ux §3.7/§3.8): nombre único, permisos por módulo
    con casillas (los de F3–F9 marcados "disponible más adelante", §3.3), **al menos un permiso**
    (FR-014/US4 esc. 4), combinaciones libres (US4 esc. 2), y acciones: editar (nombre y permisos),
    **eliminar** solo cuando `userCount === 0` (FR-017/US6 esc. 4–5: si está en uso, el mensaje pide
    reasignar antes) y visibilidad del `userCount`. Los cambios de permisos surten efecto sin pasos
    adicionales (FR-018/SC-009).
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: crear rol con dos permisos; sin permisos → error "al menos un
    permiso"; nombre duplicado normalizado → mensaje de duplicado (SC-011); eliminar rol sin uso →
    confirmación y desaparece; eliminar rol en uso → mensaje de reasignar y **no** se llama al
    `DELETE`; los permisos del rol editado se reflejan en el listado.
  - **Criterio de terminado**: US4/US6 y FR-014…FR-018 cubiertos en la interfaz.
  - **Commit sugerido**: `feat(frontend): gestión de roles con permisos por módulo`

- [ ] T250 · `features/auditoria`: sección de registro · `[frontend]` `[P6]`

  - **Archivos**: `frontend/src/features/auditoria/pages/AuditPage.tsx`, `hooks/useAccessEvents.ts`,
    `useAdminActions.ts` y `auditoria.test.tsx` (NUEVOS).
  - **Qué hace**: la sección de registro (ux §3.11, US8) con **dos historiales** (accesos y acciones
    administrativas), filtros por **cuenta** y **rango de fechas** (`from`/`to`), **paginación** que
    conserva los filtros (FR-024/US8 esc. 2–3), resultado e IP de origen en los accesos y
    actor/objetivo/resultado en las acciones (FR-022/FR-023; la fila de un intento sin cuenta
    asociada muestra **"Intento sin cuenta asociada"**, sin mostrar ningún correo — F-01), estado
    vacío con el texto de
    "no hay registros que coincidan con esos filtros" y **sin ningún control de edición ni borrado**
    (FR-025/US8 esc. 4). Solo visible con `admin_usuarios_roles` (FR-024/US8 esc. 5).
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: ambos historiales con sus columnas (incluido `userName`/
    `actorName`); el intento sin cuenta asociada se muestra como "Intento sin cuenta asociada" y sin
    correo (F-01); los filtros por cuenta y fechas
    se envían a la API y la paginación los conserva; la página vacía muestra el estado vacío; la
    pantalla **no contiene** botones ni acciones de edición/borrado (aserción explícita, SC-013);
    sin permiso → `403` mostrado con mensaje claro.
  - **Criterio de terminado**: US8 y FR-021…FR-025 cubiertos en la interfaz (SC-012, SC-013).
  - **Commit sugerido**: `feat(frontend): sección de auditoría de solo lectura con filtros`

- [ ] T254 · `features/panel`: Inicio del panel (`/panel`) · `[frontend]` `[P6]`
  *(añadida por el `analyze` F-02)*

  - **Archivos**: `frontend/src/features/panel/pages/InicioPage.tsx`,
    `components/MiCuentaCard.tsx`, `components/AccesosRapidos.tsx` y `panel.test.tsx` (NUEVOS).
  - **Qué hace**: la página de **Inicio del panel** (`/panel`, ruta 4 de las 7 de P18) que toda
    cuenta autenticada ve al entrar: (a) tarjeta **"Mi cuenta"** con los datos de la sesión (nombre,
    apellidos, correo y rol) y enlace a `/cambiar-contrasena`; (b) **accesos rápidos** solo a las
    secciones autorizadas por permiso (`/panel/usuarios`, `/panel/roles`, `/panel/auditoria` con
    `admin_usuarios_roles`; `hasPermission` de `lib/permissions`); (c) el estado **"cuenta sin
    permisos de módulo"** (Edge Case de la spec: "entra pero no ve ninguna sección de gestión; esto
    es válido y no debe romper el panel"): aviso comprensible que no rompe la página ni deja
    secciones vacías a la vista. Reutiliza `Notice`, `EmptyState` y `StatusPill` de T242 (sin
    duplicar markup) y `useSession` de T245.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: la tarjeta "Mi cuenta" muestra los datos de la sesión y el enlace
    de cambio de contraseña; los accesos rápidos muestran **solo** las secciones autorizadas; una
    cuenta **sin ningún permiso de módulo** ve el estado "cuenta sin permisos de módulo" y ningún
    enlace a secciones (Edge Case, caso incluido de forma explícita); ningún componente hace `fetch`
    directo.
  - **Criterio de terminado**: `/panel` es usable para toda cuenta autenticada y el Edge Case "cuenta
    sin permisos de módulo" queda cubierto con su prueba (F-02).
  - **Commit sugerido**: `feat(frontend): inicio del panel con Mi cuenta, accesos rápidos y estado sin permisos`

- [ ] T251 · E2E `acceso.spec.ts` · `[frontend]` `[P7]`

  - **Archivos**: `frontend/e2e/acceso.spec.ts` (NUEVO; config existente de F1 sin cambios).
  - **Qué hace**: el recorrido completo de `quickstart.md` §11 con Playwright (ejecución **local**;
    el CI del kit no corre e2e y no se toca): inicializar (vía API, ver "Huecos…", punto 2) → login →
    crear rol → crear cuenta → entrar con ella (contraseña forzada → cambio) → ver solo sus módulos →
    desactivarla → acceso cortado → intentos fallidos (el 5.º crea el bloqueo, el 6.º recibe `429`) → logout. Cubre **SC-002, SC-005,
    SC-006, SC-007, SC-010, SC-011**.
  - **Pruebas incluidas** (§III): la propia prueba e2e **es** el objeto de la tarea. **Cómo se
    verifica**: `make up && make e2e` en verde contra el stack levantado (con `BOOTSTRAP_TOKEN` de
    `.env`); `make e2e` es repetible (reinicia el estado con `make down && make up` o datos propios).
  - **Criterio de terminado**: el flujo crítico de acceso y gestión queda cubierto de punta a punta.
  - **Commit sugerido**: `test(e2e): recorrido de acceso y gestión con Playwright`

- [ ] T252 · E2E `auditoria.spec.ts` · `[frontend]` `[P7]`

  - **Archivos**: `frontend/e2e/auditoria.spec.ts` (NUEVO).
  - **Qué hace**: el recorrido de `quickstart.md` §10 con Playwright: generar accesos (exitosos y
    fallidos, incluido un correo inexistente) y acciones de gestión, abrir la sección de registro y
    verificar ambos historiales con fecha/hora, resultado e IP; filtros por cuenta y rango de fechas
    con paginación; el último acceso en la ficha de la cuenta (y "sin accesos" en una cuenta nueva);
    el intento fallido sin cuenta asociada; y que **no existe** ninguna forma de editar ni borrar
    registros. Cubre **SC-012, SC-013**.
  - **Pruebas incluidas** (§III): la propia prueba e2e. **Cómo se verifica**: `make up && make e2e`
    en verde (incluye `acceso.spec.ts` y este).
  - **Criterio de terminado**: US8 verificable de punta a punta.
  - **Commit sugerido**: `test(e2e): recorrido de la auditoría con Playwright`

---

## Fase 11 — Cierre

- [ ] T253 · Verificación de cierre · `[infra]`

  - **Archivos**: ninguno (evidencia en el PR). *No se toca* `.github/workflows/ci.yml` ni ningún
    archivo del kit.
  - **Qué hace**: ejecuta la verificación completa de F2 y la registra como evidencia:
    `make ci` (**lint + test + security**: `security` son `govulncheck` y `npm audit`; `test` ya
    incluye `go test ./...` y `go test -tags=integration ./...`. **`make db-migrate` es un comando
    aparte** y no forma parte de `make ci` — se ejecuta al preparar el entorno, §0), `make e2e`,
    `make sqlc-verify` y `make api-gen` sin deriva, y el
    recorrido manual de `quickstart.md` §0–§12 (mapa §12: SC-001…SC-013). Comprueba además que
    `GET /healthz` responde igual que en F1 y que una ruta inexistente responde `404` con
    `ErrorEnvelope` (§8.1.9), y que la cobertura del service es ≥ 80 % (`go test -cover`).
  - **Pruebas incluidas** (§III): — (es la ejecución de todas las anteriores). **Cómo se verifica**:
    cada comando anterior en verde y su salida adjunta al PR; el mapa de criterios → secciones de
    `quickstart.md` §12 queda recorrido completo.
  - **Criterio de terminado**: `make ci` y `make e2e` en verde; SC-001…SC-013 evidenciados. La
    revisión final (`qa-tester`, `revisor-codigo`, `seguridad` en paralelo) y el PR/CHANGELOG/README
    los cierra el orquestador con `documentador` y `devops` en la fase de entrega (no son tareas de
    este documento).
  - **Commit sugerido**: — (sin cambios de código)

---

## Cobertura de requisitos (FR-001…FR-026)

| FR | Tareas | Verificación principal |
|---|---|---|
| FR-001 (panel exige sesión) | T227 (authn), T244 (guards) | e2e T251 · quickstart §2 (SC-001) |
| FR-002 (autenticar, inactivas fuera) | T225 | pruebas del service · quickstart §2 |
| FR-003 (mensaje genérico, sin credenciales) | T225, T228, T230 | suite de handlers · quickstart §8 (SC-008) |
| FR-004 (cerrar sesión) | T225, T241, T245 | quickstart §3 |
| FR-005 (1 h absoluta + 30 min inactividad) | T218, T225, T246 | integración Redis · quickstart §3 |
| FR-006 (5 fallos / 15 min; `429` desde el 6.º intento) | T218, T225 | pruebas del service + integración · quickstart §8 |
| FR-007 (inicialización única) | T229, T230 | integración concurrente · quickstart §1 (SC-003) |
| FR-008 (nunca sin administración) | T222, T232, T236 | integración concurrente · quickstart §6 (SC-004) |
| FR-009 (crear cuenta con datos y teléfono) | T213, T231 | pruebas de service/handler · quickstart §4 |
| FR-010 (política de contraseñas) | T215, T226, T233 | tabla de casos de la política · quickstart §5 |
| FR-011 (editar cuenta) | T232, T234 | pruebas de service/handler · quickstart §4 |
| FR-012 (desactivar corta el acceso) | T218, T232 | e2e T251 · quickstart §7 (SC-006) |
| FR-013 (nunca se eliminan cuentas) | T208, T221, T234 | 0 rutas/consultas de borrado · revisión |
| FR-014 (roles con ≥1 permiso, nombre único) | T207, T235, T236 | pruebas de service · quickstart §4 |
| FR-015 (catálogo = módulos del producto) | T207, T237 | `GET /admin/permisos` · data-model |
| FR-016 (permiso por operación) | T227, T234, T237, T240 (+ T244/T254: la UI solo muestra lo autorizado) | pruebas de middleware/handler (SC-007) |
| FR-017 (un rol por cuenta; eliminar solo sin uso) | T208, T232, T236 | integración (`RESTRICT`) · quickstart §4/§6 |
| FR-018 (cambios de rol inmediatos) | T225, T236 | e2e T251 · quickstart §6 (SC-009) |
| FR-019 (listado con estado, correo, rol) | T234, T247 | quickstart §4 |
| FR-020 (cambiar la propia contraseña) | T226, T246 | quickstart §5 (SC-010) |
| FR-021 (último acceso en la ficha) | T209, T225, T234, T247 (fila del listado), T248 (**ficha = detalle/edición**, con `lastLoginAt`/`lastLoginIp`) | integración de coherencia · e2e T252 (SC-012) |
| FR-022 (historial de accesos) | T209, T223, T224, T225, T240 | pruebas de service/handler/integración · e2e T252 (SC-013) |
| FR-023 (acciones administrativas) | T209, T223, T224, T231–T236, T239 | pruebas de service/handler · quickstart §10 |
| FR-024 (sección con filtros y paginación) | T238, T240, T250 | quickstart §10 · e2e T252 (SC-012) |
| FR-025 (registro de solo lectura) | T211, T223, T240, T250 | sin `UPDATE`/`DELETE` ni rutas de escritura · SC-013 |
| FR-026 (sin credenciales en el registro) | T224, T239 + DTOs | suite de handlers · auditoría de `seguridad` |

También quedan cubiertos los Edge Cases de la spec sin FR propio: el **"cuenta sin ningún permiso
de módulo"** (entra al panel, no ve secciones de gestión y no se rompe nada) lo cubre **T254**
(Inicio del panel) con **T244** (navegación filtrada), y el **"intento sin cuenta asociada"** (sin
guardar ni mostrar correo) lo cubren T223/T240/T250 (F-01).
