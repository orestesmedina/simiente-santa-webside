# Revisión de seguridad — F2 Acceso y gestión de usuarios (2026-10-05)

**Rama**: `002-acceso-gestion-usuarios` · **Alcance**: `git diff main...HEAD` (F2: autenticación, sesión, autorización, gestión de usuarios/roles y auditoría)
**Auditor**: agente `seguridad` (solo lectura; OWASP Top 10 + checklist del plan)

## Veredicto: **APROBADO CON OBSERVACIONES**

Sin hallazgos críticos ni altos. Dos hallazgos medios —configuración fail-open de `SESSION_SECRET` y causa de errores tragada en `authn`— no son explotables por sí mismos y tienen remediación puntual. El resto son observaciones bajas y trade-offs documentados. Todos los controles exigidos por la spec (FR-002…FR-026) están implementados y respaldados por pruebas.

---

## Hallazgos

### Críticos

Ninguno.

### Altos

Ninguno.

### Medios

**M-1 · `SESSION_SECRET` vacío aceptado al arrancar: la firma del CSRF se degrada en silencio** (CWE-347/CWE-1188)

- Dónde: `backend/internal/platform/config/config.go:139` (asignación sin validar), `config.go:162-167` (la única validación de producción es `SESSION_COOKIE_SECURE`), `docker-compose.yml:69` (`SESSION_SECRET: ${SESSION_SECRET:-}` → vacío por defecto), consumido en `backend/cmd/api/main.go:81,128,140`.
- Impacto: a diferencia de `REDIS_URL` (obligatoria y validada, `config.go:132`), un `SESSION_SECRET` vacío **no impide el arranque**. Con clave HMAC vacía (conocida), cualquiera puede calcular firmas válidas de cookies `csrf_token`: el *double-submit firmado* (P10) se degrada a double-submit simple. `.env.example:19-20` dice «Vacío no funciona», pero sí funciona —degradado—. Un `make up` en un clon limpio corre hoy con la capa de firma anulada.
- Atenuación existente (por eso es Media y no Alta): `SameSite=Lax` bloquea las cookies en POST cross-site y un atacante cross-origin no puede poner la cabecera `X-CSRF-Token` en el navegador de la víctima; explotarlo requeriría inyección de cookie en subdominio o XSS.
- Remediación: rechazar el arranque si `SESSION_SECRET` está vacío (igual que `REDIS_URL`), o como mínimo exigirlo en producción junto a `SESSION_COOKIE_SECURE`; corregir el comentario de `.env.example` («vacío no funciona» → hoy sí funciona degradado).

**M-2 · `authn` convierte cualquier error del Store/Resolver en el mismo 401 sin registrar la causa** (CWE-754, riesgo evaluado a petición)

- Dónde: `backend/internal/platform/middleware/authn.go:47-57` (ni el error de `store.Resolve` ni el de `resolver.Resolve` se registran), `backend/internal/platform/httpserver/error.go:134-159` (`WriteError` solo deja `code=unauthenticated` a nivel Warn, sin causa).
- Impacto: la confluencia en un 401 idéntico es **fail-closed y sin fuga de enumeración** (correcto para FR-003/FR-012: «sin cookie», «expirada» y «desactivada» son indistinguibles). Pero un **fallo transitorio de PostgreSQL o Redis** produce lo mismo: todo el panel recibe 401, los guards del frontend (`frontend/src/app/guards.tsx:29-33`) expulsan a todo el mundo a `/login`, y **la causa raíz no queda en el log** — no se puede distinguir una caída de infraestructura de un incidente de acceso, y la detección del problema queda a ciegas para operaciones.
- Remediación: registrar la causa en el log del servidor (p. ej. `logger.Warn(..., "error", err)` en las dos ramas de error de `authn`) manteniendo el 401 uniforme al cliente; valorar responder 503 ante errores de infraestructura identificables (Redis/PG caídos), que son uniformes para todas las cuentas y no ayudan a la enumeración.

### Bajos

**B-1 · `BOOTSTRAP_TOKEN` vacío desactiva la inicialización en silencio** — `config.go:140`, `usuarios/handler_auth.go:123-134`. Fail-closed (`ConstantTimeCompare` con esperado vacío nunca autoriza: el 401 de cabecera ausente llega antes), pero tampoco avisa: un despliegue sin la variable deja el panel ininizializable sin ningún error de arranque. Remediación: warning al arrancar (o no publicar la ruta si el token está vacío, o error de arranque en producción).

**B-2 · Rate-limit e IP de auditoría tras un proxy inverso** — `middleware/ratelimit.go:194-206` y `usuarios/handler.go:214-227` usan `RemoteAddr` sin soporte de proxy de confianza (decisión documentada, RG5). En producción detrás de un proxy TLS, todas las peticiones comparten la IP del proxy: el umbral de 20/min de `login`/`setup` se vuelve **global** (denegaciones accidentales masivas) y `login_events.last_login_ip`/`ip` registran la IP del proxy. Remediación: al desplegar en producción, configurar un `ClientIP` de confianza (solo con un proxy confiable) o mover el rate-limit al proxy; anotarlo como requisito del despliegue.

**B-3 · `POST /api/v1/auth/password` sin rate-limit ni contador de fallos** — `routes.go:61-64`, `main.go:126-129` (el grupo de sesión solo monta authn+CSRF; `DefaultRateLimitPaths` solo cubre `login` y `setup/initialize`). Un atacante con una sesión válida (robada) puede probar contraseñas actuales a ~coste bcrypt (≈100 ms) sin tope. Exige sesión+CSRF, por lo que el riesgo es bajo. Remediación: extender el rate-limit a esta ruta (por identidad).

**B-4 · Cabeceras de seguridad ausentes** — ni el backend (`httpserver/error.go:103-116`, `WriteJSON`) ni el nginx de la SPA (`frontend/nginx.conf`) fijan `X-Content-Type-Options: nosniff`, `X-Frame-Options`/`frame-ancestors`, `Content-Security-Policy` o HSTS. Riesgo bajo (API JSON + SPA estática). Remediación: añadir `nosniff` en el middleware transversal y CSP/HSTS cuando haya HTTPS real.

**B-5 · Bloqueo por identificador permite DoS de cuenta conocida (trade-off aceptado por la spec)** — `service_auth.go:160-175`, `session/throttle.go` (claves `login:fail:<correo>` exista o no la cuenta, a propósito para FR-003). Cualquiera puede bloquear 15 min un correo conocido con 5 fallos. Es exactamente el comportamiento confirmado por el humano (FR-006, 2026-10-04): se deja constancia del trade-off; no requiere cambio.

**Nota sobre secretos de desarrollo**: `frontend/e2e/helpers.ts:21-22` fija `dev-bootstrap-token` como valor por defecto. Es un valor de desarrollo documentado (quickstart), no un secreto real; el riesgo solo existe si alguien despliega a producción reutilizando el valor del ejemplo. Se anota para que el despliegue genere siempre un token propio.

---

## Confirmación de controles correctos (verificados en código y pruebas)

**Autenticación (OWASP A07)**
- **bcrypt cost 12**: `password/password.go:29` (`bcryptCost = 12`); el hash dummy de correo inexistente es también cost 12 (`service_auth.go:45`), y `password.MaxBytes` respeta el límite de 72 bytes de bcrypt (`password.go:68`).
- **Verificación dummy (R18)**: correo inexistente → `password.Verify(dummyPasswordHash, …)` antes del 401 (`service_auth.go:179-185`); mismo coste temporal, sin enumeración. Prueba: `TestLoginGenericFailureIsIdenticalForUnknownAndWrongPassword`.
- **Enumeración de cuentas (FR-003/SC-008)**: 401 genérico idéntico (`loginFailureMessage`) para «contraseña incorrecta» y «correo inexistente»; el 403 de cuenta inactiva exige credenciales **correctas** (US1 esc. 3, spec-mandated) y lleva `details.reason=access_disabled` solo para la UI; el 429 de bloqueo es idéntico exista o no la cuenta. El login durante bloqueo no consulta la cuenta (`service_auth.go:167-175`) → tampoco filtra por ahí.
- **Bloqueo FR-006 (5/15 min) y `Retry-After`**: constantes en el dominio (`MaxFailedAttempts=5`, `LockoutDuration=15m`, `service_auth.go:34-39`); el 5.º fallo responde 401 genérico y crea la bandera con TTL (`session/throttle.go:69-86`); desde el 6.º → 429 con `Retry-After` y `details.retryAfterSeconds` (`error.go:94-96` fija la cabecera; d76a45c). Prueba: `TestLoginLockoutFifthRespondsGenericAndSixthRateLimited`. Los intentos durante el bloqueo se registran (Edge Case FR-006/FR-022).
- **Inicialización única (FR-007)**: `X-Setup-Token` comparado en tiempo constante (`subtle.ConstantTimeCompare`, `handler_auth.go:131`); sin cabecera → 401, valor distinto → 403, mensajes que no revelan valor esperado ni estado (`handler_auth.go:25-28`); `CountUsers=0` **dentro** de la transacción con advisory lock compartido con el guard (`service_init.go:90-97`, `repository_roles.go:225-235`) → dos inicializaciones simultáneas no crean dos administradores; repetida → 409; rate-limit por IP.

**Sesión (OWASP A02/A07)**
- Token opaco de 32 bytes `crypto/rand` (`session/token.go:25-31`); **Redis guarda solo SHA-256** (`sess:<hash>`, `store_redis.go:50-52`); el token en claro solo existe en la cookie.
- Cookies: `ss_session` con `HttpOnly`, `SameSite=Lax`, `Path=/`, `Secure` configurable y **obligatorio `true` en producción (fail-fast, `config.go:162-167`)**, `Max-Age = vida absoluta` (`session/cookies.go:61-71`); borrado con `Max-Age<0` en logout/revocación. La cookie CSRF no es `HttpOnly` a propósito (la SPA la lee); no es una credencial.
- **TTL de inactividad 30 min + vida absoluta 1 h**: `StoreConfig` validada; `Resolve` refresca el TTL de inactividad **acotado** a la vida absoluta (`nextTTL`, `store.go:95-101`) y `authn` comprueba `ExpiredAt` de forma defensiva (`authn.go:48`); `lastSeenAt` estrangulado a 1 escritura/min.
- **Revocación por cuenta**: índice `user_sessions:<uuid>` permite `RevokeUser` (desactivación/restablecimiento, FR-012/R17) y `RevokeUserExcept` (cambio propio conservando la sesión actual, R17) sin `SCAN`.
- **Nada de tokens en `localStorage`**: grep limpio en `frontend/src` (solo `document.cookie` para la CSRF, `client.ts:104-116`); `fetch` con `credentials: 'include'` (`client.ts:139`). Prueba de `TestLoginResponseNeverContainsPassword`.

**CSRF (A01)** — Double-submit **firmado**: cookie `csrf_token` + cabecera `X-CSRF-Token` deben coincidir (`hmac.Equal`) y la cookie debe llevar firma HMAC-SHA256(SESSION_SECRET, nonce) válida (`session/cookies.go:173-181`, middleware `csrf.go:34-52`); exención solo de métodos seguros (GET/HEAD/OPTIONS/TRACE); aplica en `/api/v1/auth` (logout, password) y en todo `/api/v1/admin` tras authn. La CSRF del login no es necesaria (no hay sesión previa y el token de sesión lo genera siempre el servidor — sin fijación de sesión). *(La capa de firma depende de M-1.)*

**CORS (A05)** — Eco **exacto** del `Origin` de la lista permitida, **nunca `*`**; `Allow-Credentials: true` solo con origen permitido; `Vary: Origin` (anti-confusión de caché); preflight real (OPTIONS con `Access-Control-Request-Method`) → 204 con `Allow-Methods/Headers/Max-Age` acotados al contrato (`middleware/cors.go:38-79`). Orígenes vacíos o no permitidos no reciben cabeceras CORS.

**Autorización (A01/A08)** — Todo `/api/v1/admin` pasa por la cadena aprobada `authn → PasswordGuard → authz(admin_usuarios_roles) → CSRF` (`middleware/chain.go:51-58`, `routes.go:79-92`, cableada en `main.go:134-146`): permiso FR-016 verificado **en servidor** en cada endpoint (los guards de React son solo UX). La denegación responde 403 genérico y se registra como `result='denied'` en `admin_actions` (best-effort, fail-closed aunque el registro falle, `authz.go:35-53`). Sin identidad → 401 de red de seguridad. El guard de cambio obligatorio responde 403 con `details.reason=password_change_required` **solo** en `/admin`; `/auth/session`, `/auth/logout` y `/auth/password` quedan libres para resolverlo (FR-010/US7 esc. 4-5). Permisos resueltos **por petición** desde el rol vigente (FR-018/SC-009) y solo para códigos del catálogo sembrado (FR-015; `resolvePermissionIDs` rechaza códigos desconocidos).

**Contraseñas y auditoría (A09, FR-003/FR-026)** — `UserItem`/`SessionUser` sin hash ni contraseña (`model.go:266-293`); el restablecimiento registra actor+objetivo+fecha, **nunca** la contraseña (`service_users.go:332-341`); `audit.Event` solo lleva `UserID/Result/IP` — del intento con correo inexistente no se guarda el correo probado (`audit/audit.go:85-89`, LEFT JOIN null). `decodeAndValidate`/`recordRejected` nunca transportan el cuerpo (`handler.go:130-191`). Grep limpio de `console.*` con contraseñas en el frontend; los formularios limpian la credencial tras el envío (`UserForm.tsx:273-275`). **Auditoría de solo lectura (FR-025)**: solo GET en `/auditoria/*` (`handler_audit.go:100-106`); `audit.sql` sin una sola sentencia UPDATE/DELETE (solo INSERT/SELECT, verificado); `admin_actions` con FKs `ON DELETE RESTRICT` y catálogos cerrados por CHECK (migración 000004).

**Invariantes de negocio** — Guard anti-bloqueo FR-008: advisory lock transaccional + recuento de admins activos **después** de la mutación, 409 `admin_required` con rollback (`repository_roles.go:225-259`); cuentas **sin** operación de borrado (FR-013, verificado en rutas y SQL); un rol en uso no se elimina (FK `RESTRICT` + `CountRoleUsers`); rol siempre con ≥1 permiso; unicidad normalizada de correos y nombres de rol (Q5) con índice `lower(name)` y `UNIQUE(email)`.

**Inyección y validación (A03)** — 100 % de las consultas generadas por sqlc con `sqlc.arg/narg` (verificado: 26/27/18 parámetros en audit/users/roles); cero `fmt.Sprintf` SQL, cero comandos de sistema. Entrada: `http.MaxBytesReader` 1 MiB (CWE-400), `DisallowUnknownFields`, rechazo de segundo valor JSON, etiquetas `min/max/email/phone` + UUID en `{id}` + RFC3339 en filtros + paginación con tope 100.

**XSS (A03)** — Sin `dangerouslySetInnerHTML` ni URLs dinámicas sin validar en `src/`; React escapa por defecto.

**Secretos (A02/A07)** — `.env` gitignored y **no versionado** (verificado con `git ls-files`/`check-ignore`); `.env.example` solo placeholders (`cambiar_esto`, `generar_un_valor_aleatorio_largo`); `SESSION_SECRET`/`BOOTSTRAP_TOKEN` sin defecto en código. **gitleaks sobre el diff `main...HEAD`**: 1 único hallazgo, falso positivo (hash `h1:` del módulo `moby/moby/api` en `go.sum`, regla `generic-api-key`).

**Configuración (A05)** — Errores con sobre uniforme; la causa interna **solo** al log estructurado con `request_id` (`error.go:80-159`); `details` suprimidos en los 500; mensajes siempre en español, sin internos. Timeouts HTTP presentes (Slowloris mitigado, `server.go:22-27`). Redis sin persistencia a propósito (P23): sesión y contadores son efímeros.

**Dependencias (A06)** — `govulncheck ./...`: **0 vulnerabilidades en código llamado**; único aviso a nivel de módulo: GO-2026-5932 (`golang.org/x/crypto/openpgp`, paquete sin uso — el proyecto solo usa `x/crypto/bcrypt`). `npm audit --audit-level=high`: **0**. La transitiva `github.com/moby/go-archive v0.3.0` (vía testcontainers) está bumpeada y solo afecta a pruebas; `internal/tools/tools.go` con `//go:build tools` es correcto: nunca entra en un build real y solo justifica el `go.mod` para `go mod tidy`.

---

## Resumen de remediación (orden sugerido)

1. **M-1**: validar `SESSION_SECRET` no vacío al arrancar (o al menos en producción) — `config.go`.
2. **M-2**: registrar la causa de los errores de `store.Resolve`/`resolver.Resolve` en el log de `authn` (401 uniforme intacto al cliente) — `middleware/authn.go`.
3. B-1…B-4 según criterio; B-5 queda como trade-off documentado.
