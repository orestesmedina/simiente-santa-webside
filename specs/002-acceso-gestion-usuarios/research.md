# Research — F2 Acceso y gestión de usuarios

**Fecha**: 2026-10-04 · **Rama**: `002-acceso-gestion-usuarios` · **Spec**: `spec.md` (aprobada el 2026-10-04)

Formato: Decisión / Justificación / Alternativas consideradas. Cierra además las preguntas abiertas
2–5 de `docs/tecnico/decisiones.md` (D-A7, chi, `platform/validate`, tipos UUID). **R1–R4 son la
aterrización de D-A7 y quedan sujetas a confirmación humana** (no se implementan hasta que el
humano las apruebe con el plan).

## R1. Dónde vive la sesión *(D-A7 — PENDIENTE DE CONFIRMACIÓN HUMANA)*

- **Decisión**: **sesión en servidor en la tabla PostgreSQL `sessions`**. La cookie solo lleva un
  token aleatorio opaco; la BD guarda su **SHA-256** (no el token). Revocar es borrar la fila.
  El detalle está en `data-model.md` (`sessions`) y en P1 del plan.
- **Justificación**:
  - FR-012 exige que desactivar una cuenta corte el acceso **de inmediato, también las sesiones
    abiertas**: con sesión en servidor eso es un `DELETE FROM sessions WHERE user_id = …` (y, como
    red de seguridad, `authn` revalida que la cuenta siga activa en cada petición). Con un token
    autocontenido (JWT) habría que mantener una lista negra, que es… una sesión en servidor con
    otro nombre.
  - Cero servicios nuevos: PostgreSQL ya es obligatorio (constitución §II), `database.WithTx` y
    sqlc ya existen y el volumen real son decenas de cuentas con pocas sesiones.
  - Guardar el **hash** del token y no el token evita que una lectura indebida de la BD permita
    secuestrar sesiones; la comparación es por igualdad de un hash (consulta por índice único).
- **Alternativas consideradas**:
  - *Redis / memcached para sesiones*: rechazado — añade un servicio al `docker-compose.yml`, una
    dependencia de cliente y una pieza más que operar para un panel con decenas de usuarios. Se
    reabriría si hubiera varias instancias de backend y la escritura de `last_seen_at` molestara.
  - *JWT autocontenido en cookie*: rechazado (y ya lo estaba en D-A7) — revocación al desactivar
    forzada con lista negra; más superficie (algoritmo, expiración, refresco) para el mismo
    resultado.
  - *Cookie firmada sin estado (clave-valor firmado con `SESSION_SECRET`)*: rechazado — no permite
    revocación individual real ni "cerrar todas las sesiones".
  - *Token en `localStorage`*: rechazado — expuesto a XSS (CWE-79) y prohibido por la skill.
  - *Framework de auth completo (OIDC/Keycloak)*: rechazado — peso desproporcionado para un panel
    interno, y exige operar un IdP.
- **Consecuencias**: tabla `sessions` + `internal/platform/session` (token y cookie, sin SQL) +
  `middleware.Authn` resolviendo la identidad por petición. **Estado: propuesta — pendiente de
  confirmación del humano junto con R2, R3 y R4.**

## R2. La cookie de sesión y sus flags *(D-A7 — pendiente de confirmación)*

- **Decisión**: cookie **`ss_session`** con `HttpOnly`, `SameSite=Lax`, `Path=/`, `Secure` según
  `SESSION_COOKIE_SECURE` (en `false` por defecto en desarrollo; `platform/config` **falla al
  arrancar** si `APP_ENV=production` y está en `false`). Valor: 32 bytes de `crypto/rand` en
  base64url; caduca junto con la sesión (inactividad 30 min; vida absoluta propuesta de 12 h, R15).
  Se emite en la respuesta de `POST /api/v1/auth/login` y se borra en `logout` y en toda
  revocación.
- **Justificación**: `HttpOnly` impide que JavaScript lea la sesión (mitigación de XSS, CWE-79);
  `SameSite=Lax` bloquea el envío de la cookie en peticiones cross-site de métodos no seguros
  (primera barrera de CSRF, que se refuerza con R3); `Secure` impide que viaje en claro cuando hay
  TLS. `Lax` y no `Strict` porque `Strict` rompería el reingreso por enlace de nivel superior a la
  SPA en algunos navegadores y no aporta frente al CSRF que ya cubre R3. El nombre `ss_session` es
  host-only (sin `Domain`): cada host solo ve su cookie.
- **Alternativas consideradas**: `SameSite=Strict` (ver arriba); `SameSite=None; Secure` (obliga a
  TLS y abre la superficie cross-site sin necesidad); cookie con `Domain` de la organización
  (innecesario: frontend y API comparten *site*); exponer el token de sesión en el cuerpo de la
  respuesta (invita a guardarlo en JS; innecesario porque el navegador gestiona la cookie).
- **Nota de entorno local**: el frontend (`.env.example` → `VITE_API_URL=http://localhost:8080`)
  corre en otro puerto pero el mismo *site* (`localhost`), de modo que `SameSite=Lax` **sí** se
  envía en las peticiones `fetch` con `credentials: "include"`. Con `Secure=true` sobre `http://`
  algunos navegadores rechazan la cookie: por eso el desarrollo va con `false`.

## R3. CSRF en las rutas con sesión *(D-A7 — pendiente de confirmación)*

- **Decisión**: **double-submit firmado**. Cookie `csrf_token` (no `HttpOnly`) cuyo valor es
  `nonce.HMAC-SHA256(SESSION_SECRET, nonce)`; todo método no seguro dentro de un grupo con sesión
  exige la cabecera `X-CSRF-Token` con **el mismo valor**, y `middleware.CSRF` comprueba (a) que la
  cookie pasa el HMAC y (b) que cookie y cabecera coinciden. El token se emite junto a la sesión y
  se renueva en cada login. No toca la base de datos.
- **Justificación**: con cookies de sesión, todo endpoint escribible autenticado es vulnerable a
  CSRF (D-A7 ya lo anuncia como consecuencia). El patrón double-submit es el estándar para SPAs con
  API en otro origen; firmarlo con `SESSION_SECRET` (ya documentada en `.env.example`, hoy sin uso)
  cierra la variante en la que un atacante que puede escribir cookies (subdominio comprometido)
  fija un par cookie/cabecera de su cosecha. Es ~50 líneas y cero dependencias.
- **Alternativas consideradas**:
  - *Token anti-CSRF guardado en la fila de sesión*: rechazado — exige leer la sesión dos veces o
    propagar el token por el contexto; mismo resultado con más estado. Queda como variante si
    double-submit resultara incómodo.
  - *Exigir solo una cabecera personalizada (p. ej. `X-Requested-With`)*: rechazado — barrera
    débil y sin valor criptográfico.
  - *Confiar solo en `SameSite`*: rechazado — `Lax` no cubre toda la superficie (navegación de
    nivel superior con GET no es el problema, pero integraciones futuras y navegadores viejos sí).
  - *Sincronizer token de un framework (gorilla/csrf)*: rechazado — dependencia para lo que ya está
    decidido que sea propio (D-A8).
- **Alcance**: solo rutas con sesión y método no seguro. `POST /api/v1/auth/login` y
  `POST /api/v1/setup/initialize` **no** llevan sesión (no hay token que secuestrar todavía) y se
  cubren con `rate-limit` + el bloqueo por intentos (R5/R12). El login CSRF (que un atacante
  "entre" a la víctima en la cuenta del atacante) queda documentado como riesgo aceptado: sin datos
  de la víctima y sin persistencia, su impacto es despreciable en este panel.

## R4. Hash de contraseñas: bcrypt *(D-A7 — pendiente de confirmación)*

- **Decisión**: **bcrypt con cost 12** (`golang.org/x/crypto/bcrypt`) en un paquete propio
  `internal/platform/password` que además implementa la política FR-010 (validación) y expone
  `Hash`, `Verify` y `Validate`. La política se aplica **idéntica** en los tres flujos: contraseña
  inicial al crear la cuenta, restablecimiento por administrador y cambio propio (FR-010).
- **Justificación**: §IV exige "bcrypt o argon2" (CWE-256) — **no hay forma de cumplirlo sin una
  dependencia nueva**, y `golang.org/x/crypto` es el subrepositorio oficial de criptografía de Go.
  bcrypt es el estándar más revisado para credenciales de usuario, su cost es ajustable y el hash
  es **autodescriptivo** (`$2a$…`), de modo que un futuro paso a argon2id no rompe nada: se
  re-hashea al verificar/autenticar. Cost 12 ≈ 250 ms por verificación: cómodo para el usuario y
  caro para la prueba masiva de contraseñas (se combina con el bloqueo de R5 y el rate-limit de R12).
- **Política FR-010 (implementación)**: longitud **8–64 caracteres y ≤ 72 bytes** (límite duro de
  bcrypt: rechaza entradas mayores de 72 bytes — por eso la validación lo advierte antes), al menos
  una mayúscula, una minúscula, un número y un carácter especial, y **distinta del nombre y del
  correo** comparada de forma normalizada (trim + minúsculas). El error identifica el requisito
  incumplido (`details.newPassword`).
- **Alternativas consideradas**:
  - *argon2id*: válida por §IV y más resistente a GPU/ASIC, pero añade parámetros que afinar
    (memoria, iteraciones, paralelismo) sin una necesidad hoy; además usa **la misma** dependencia.
    Queda como evolución natural (migración transparente por el prefijo del hash).
  - *pbkdf2/scrypt de la stdlib*: rechazado — §IV dice literalmente bcrypt o argon2.
  - *Un hash "propio" (SHA-256 + sal)*: rechazado — viola §IV y es criptográficamente insuficiente
    (rápido de fuerza bruta).
- **Constantes de código**: `bcryptCost = 12`, longitud 8–64 caracteres, 72 bytes. No se pasan por
  variables de entorno: son exigencias de la política confirmada por el humano (2026-10-04), no
  ajustes operativos.

## R5. Bloqueo por intentos fallidos (FR-006) sin enumerar cuentas (FR-003)

- **Decisión**: tabla `login_attempts` con fila **por identificador normalizado (el correo), exista
  o no la cuenta**. Al fallar: `upsert` que incrementa `failed_count` y fija `last_failed_at`; al
  llegar a **5** se fija `blocked_until = now() + 15 minutos`. Si hay bloqueo vigente, la respuesta
  es `429 rate_limited` con el mensaje de bloqueo; si no, la respuesta es siempre el mismo
  `401 unauthenticated` genérico. Un inicio de sesión correcto borra la fila; el bloqueo expirado
  se resetea al siguiente intento. Los valores **5** y **15 minutos** son **constantes de código**
  (`maxFailedAttempts`, `lockoutDuration`) porque el humano los confirmó el 2026-10-04.
- **Justificación**: FR-006 pide bloquear tras 5 intentos y FR-003/SC-008 piden que ningún mensaje
  revele si la cuenta existe. Si el contador viviera solo en `users`, el mensaje de bloqueo
  **solo** aparecería para cuentas reales → enumeración garantizada. Contando por identificador
  (haya cuenta o no) el comportamiento —mensaje, código y tiempos— es **idéntico** en ambos casos y
  las dos exigencias se cumplen a la vez. La fila se limpia con el éxito y con la expiración, así
  que no crece sin control.
- **Alternativas consideradas**: contador solo en `users` (enumeración); contador solo en memoria
  (se pierde al reiniciar el proceso y no sobrevive a varias instancias); rate-limit por IP como
  única medida (no frena la prueba masiva de contraseñas contra **una** cuenta desde una IP); bloqueo
  permanente hasta intervención de un administrador (no lo pide la spec y bloquea al usuario
  legítimo).
- **Mensajes** (siempre en español, sin datos internos): intentos 1–4 → *"Correo o contraseña
  incorrectos"* (genérico); desde el 5.º y durante el bloqueo → *"Demasiados intentos fallidos. El
  acceso queda bloqueado temporalmente durante 15 minutos"*, con `Retry-After` en segundos.

## R6. Inicialización única del administrador (FR-007)

- **Decisión**: `POST /api/v1/setup/initialize`, público y **rate-limited**, que exige la cabecera
  `X-Setup-Token` con el valor de la variable `BOOTSTRAP_TOKEN` y solo funciona mientras la tabla
  `users` esté **vacía**. Todo ocurre en una transacción con `pg_advisory_xact_lock` (misma clave
  que el guard anti-bloqueo, R7): comprueba `users` vacío, crea el rol **"Administrador"** con los
  9 permisos del catálogo y la cuenta inicial con `mustChangePassword = false` (su contraseña ya es
  una elegida por quien inicializa). Repetirla → `409 conflict` (*"La inicialización ya se hizo y
  no puede repetirse"*).
- **Justificación**: US2 esc. 1–2 exigen la acción única y que no se repita; el guard de BD lo
  garantiza incluso con dos peticiones simultáneas. FR-007 añade "impedir cualquier uso abusivo":
  sin más, cualquiera que alcance una instalación recién desplegada antes que el equipo se autonombra
  administrador. El token de despliegue cierra esa ventana con **una sola** variable de entorno (no
  es un secreto de usuario: se configura al desplegar, como `DATABASE_URL`).
- **¿Por qué se crea también un rol?** El modelo es "una cuenta = un rol" (Q4) y `users.role_id` es
  `NOT NULL`: el administrador inicial necesita un rol con todos los permisos. Ese rol nace con el
  nombre "Administrador" (editable después como cualquier otro) y es el que satisface la regla
  anti-bloqueo desde el primer segundo.
- **Alternativas consideradas**: inicialización sin token (rechazada: ventana de robo del primer
  administrador); script/CLI de arranque (rechazada: exige acceso al servidor y no es el producto);
  crear el administrador en una migración con contraseña fija (rechazada: secreto en el repo,
  §IV CWE-798); primera cuenta auto-registrable con un código por correo (rechazada: no hay servicio
  de correo en el MVP y la spec prohíbe el auto-servicio).
- **Operativa**: `BOOTSTRAP_TOKEN` se documenta en `.env.example` (valor de ejemplo) y en
  `quickstart.md` §1. Si el humano **no** quiere el token (R13 del plan), se retira esta única
  validación y queda el guard de `users` vacío + rate-limit.

## R7. Regla anti-bloqueo (FR-008) y sus carreras

- **Decisión**: toda operación que podría dejar el panel sin administración (desactivar una cuenta,
  cambiar su rol, editar los permisos de un rol) se ejecuta en una transacción que (1) toma
  `pg_advisory_xact_lock(hashtext('f2_admin_guard'))` —serializa estas operaciones, que son raras—,
  (2) aplica la mutación y (3) **cuenta después** las cuentas activas cuyo rol incluye el permiso
  `admin_usuarios_roles`; si el resultado es 0 → `apperr.Conflict` y rollback. La eliminación de
  roles ya está bloqueada por otra regla si hay cuentas asignadas (FR-017), y aun así conserva el
  mismo recuento como red de seguridad.
- **Justificación**: la spec (Edge Cases) pide que **ninguna operación** pueda dejar el panel sin
  administración. Con dos administradores desactivándose mutuamente a la vez, una comprobación
  "previa" sin serialización deja el sistema en el estado prohibido; el advisory lock lo hace
  imposible sin depender de niveles de aislamiento especiales (READ COMMITTED basta). El recuento es
  una consulta que ya existe para el listado (filtro por permiso), así que no añade lógica nueva.
- **Alternativas consideradas**: comprobación previa sin lock (rechazada: carrera demostrable);
  `SERIALIZABLE` en estas transacciones (funciona pero obliga a manejar serializability failures);
  trigger o restricción de BD (rechazada: lógica de negocio fuera del service, difícil de probar y
  de traducir a `apperr`); impedir desactivar administradores en absoluto (contradice US2 esc. 5,
  que pide poder desactivar uno si queda otro).

## R8. Modelo de datos: simplificación de la referencia de F1

- **Decisión**: `users`, `sessions`, `roles`, `permissions`, `role_permissions`, `login_attempts`.
  **Se elimina `user_roles`** (la referencia orientativa de F1 `data-model.md` §"F2" lo incluía):
  una cuenta tiene **un solo rol** (decisión Q4), así que la relación vive en `users.role_id`.
  **Se añade `permissions`** como catálogo fijo sembrado por la migración (los 9 módulos de
  FR-015, incluidos los reservados de F3–F9) y **`login_attempts`** (R5). Detalle completo en
  `data-model.md`.
- **Justificación**: `user_roles` permitiría varios roles por cuenta, exactamente lo que Q4 prohíbe;
  conservarlo "por si acaso" añadiría una tabla y reglas de "rol efectivo" que la spec no quiere. La
  tabla `permissions` da FK y unicidad reales a `role_permissions` y permite al panel listar el
  catálogo con etiquetas (`GET /api/v1/admin/permisos`) sin duplicarlo en código.
- **Alternativas consideradas**: mantener `user_roles` (contradice Q4); permisos como `TEXT[]` en
  `roles` (sin FK, validación de elementos incómoda en SQL, sin catálogo listable); permisos solo
  como constantes de Go (el catálogo viviría en dos sitios: BD para `role_permissions` y código para
  la UI); permisos dinámicos creados por el administrador (la spec fija el catálogo = módulos del
  producto; "sin catálogo fijo" se refiere a **roles**, no a permisos).

## R9. Normalización y unicidad (decisión Q5)

- **Decisión**: el correo se guarda **normalizado** (`trim` + minúsculas) en `users.email` con
  `UNIQUE` (y `CHECK` que lo garantiza); el nombre de rol se guarda con `trim` + colapso de espacios
  conservando sus mayúsculas de presentación, con `UNIQUE (lower(name))`. Un duplicado "casi igual"
  (otras mayúsculas o espacios sobrantes) se rechaza con `409 conflict` y un mensaje claro, tanto en
  la comprobación previa del service como en la traducción del `unique_violation` (carrera entre dos
  creaciones simultáneas, Edge Case de la spec).
- **Justificación**: la unicidad la garantiza la **base**, no solo el código; guardar ya normalizado
  evita que "Ana@Ejemplo.com" y "ana@ejemplo.com " convivan como datos distintos. El nombre de rol
  conserva su forma de presentación porque la spec pide que "aparezca en todo el panel con el nuevo
  nombre" (US6 esc. 3).
- **Alternativas consideradas**: índice funcional sobre el valor crudo (permite guardar la variante
  "sucia"); comparar solo en el service (pierde ante dos peticiones simultáneas); guardar el correo
  tal cual con `UNIQUE (lower(email))` (aceptable, pero deja a la BD dos formas del mismo dato y
  complica los mensajes de duplicado).

## R10. chi vs. `net/http` (cierra la pregunta 3 de `decisiones.md`, D-A4)

- **Decisión**: **se mantiene `net/http`** tras la interfaz `httpserver.Registrar`. No se adopta chi
  en F2.
- **Justificación**: el caso que motivaba reevaluar chi (grupos de rutas de panel con permisos) ya
  está resuelto por `Registrar.Group(prefix, mws...)`: `/api/v1/admin` con `authn → guard → authz →
  CSRF` y subgrupos por módulo son 3 líneas en `main.go`. Los comodines `{id}` de Go 1.22+ cubren
  `/admin/usuarios/{id}` sin más. chi aportaría 0 dependencias directas pero sí un adaptador y una
  segunda pieza que mantener, sin dolor actual.
- **Criterio de reapertura**: si aparecen rutas que `http.ServeMux` no exprese con claridad
  (grupos anidados profundos, middleware por ruta, wildcards compuestos), se adopta chi con su
  adaptador de ~15 líneas y **ningún dominio cambia** (la neutralidad de `Registrar` lo garantiza).

## R11. `platform/validate` propio y tipos UUID (cierra las preguntas 4 y 5)

- **Decisión (validación)**: implementación **propia mínima** en `internal/platform/validate` con
  reflexión sobre las etiquetas `validate` ya usadas en los DTOs de `arquitectura.md` §5.4
  (`required`, `omitempty`, `min`, `max`, `email`, `oneof`). Produce `apperr.Invalid` con
  `details` por campo (el formato que ya documenta el sobre de error de F1).
- **Justificación**: son seis reglas para decenas de campos; una librería (`go-playground/validator`)
  añade una dependencia directa con transitivas, mensajes genéricos en inglés que habría que
  traducir igualmente, y una superficie de reglas que no usaremos (D-A8). La validación de negocio
  (política de contraseñas, normalización) vive en el service: la frontera HTTP solo hace la
  validación de forma/formato (§IV: validar en el backend aunque el frontend valide).
- **Decisión (UUID)**: `github.com/google/uuid` en el dominio (`uuid.UUID` en entidades y
  parámetros; en JSON viaja como `string`), `pgtype.UUID` solo en `repository.go` (conversión por
  `.Bytes`). Es **la primera funcionalidad con tablas UUID**, así que se repite aquí la
  justificación que exige §8.1.1: dependencia nueva **justificada** bajo D-A8 (solo tipos, sin
  dependencias transitivas; mantiene `pgx` fuera del dominio, R4). `sqlc.yaml` **no** usa
  `overrides`.
- **Alternativas consideradas**: `go-playground/validator` (ver arriba); validar a mano en cada
  handler (repetición propensa a errores y sin formato de `details` uniforme); `pgtype.UUID` en el
  dominio (arrastra `pgx` a las capas altas); `override` de sqlc para emitir `uuid.UUID` (segunda
  regla para un tipo, contraria a §8.1.6).

## R12. `rate-limit` mínimo (cierra el punto de decisión de D23)

- **Decisión**: middleware propio en memoria, ventana deslizante por IP, aplicado **solo** a los
  endpoints públicos escribibles (`POST /api/v1/auth/login`, `POST /api/v1/setup/initialize`):
  umbral de **20 peticiones por minuto por IP** (constante), respuesta `429 rate_limited` con
  `Retry-After`. Marca de tiempo y contador por IP en un mapa con limpieza oportunista.
- **Justificación**: D23/D-A9 fijaban este punto de decisión en "el primer endpoint público
  escribible" — ese momento ha llegado con F2 (login). El bloqueo por cuenta (R5) protege una
  cuenta; el rate-limit por IP amortigua la prueba masiva de contraseñas y el bombardeo del endpoint
  de inicialización (DOS a bcrypt, R14 del plan).
- **Limitación declarada**: en memoria no comparte umbrales entre instancias (R5 del plan). Hoy hay
  una instancia; si se escalase, la decisión se reabre en esa funcionalidad.
- **Alternativas consideradas**: no hacerlo (dejaría el diferido sin decidir y el login sin
  amortiguación por IP); `golang.org/x/time/rate` (dependencia para ~40 líneas); Redis (servicio
  extra); rate-limit global en todas las rutas (molestaría al panel legítimo sin aportar).

## R13. CORS con credenciales (revisión de D16, cerrada)

- **Decisión**: se **amplía** el middleware CORS actual (sin dependencia): `Access-Control-Allow-
  Credentials: true`, eco exacto del `Origin` si está en `CORS_ALLOWED_ORIGINS` (nunca `*` cuando
  hay credenciales), `Access-Control-Allow-Headers: Content-Type, X-CSRF-Token, X-Request-ID`,
  `Access-Control-Expose-Headers: X-Request-ID`, `Vary: Origin` y los preflight como hasta ahora.
- **Justificación**: la sesión viaja en cookie, así que las llamadas del navegador van con
  `credentials: "include"` y CORS debe declararlo. El middleware ya existe, resuelve bien el eco de
  `Origin` y su ampliación son unas pocas líneas.
- **Alternativas consideradas**: `github.com/rs/cors` (dependencia para lo ya escrito); proxy
  `/api` en nginx del frontend (esconde CORS pero complica el desarrollo con `npm run dev` y el
  patrón `VITE_API_URL` ya está fijado); `Access-Control-Allow-Origin: *` (imposible con
  credenciales).

## R14. Un dominio `usuarios` (y no `auth` + `usuarios`)

- **Decisión**: todo F2 vive en **un paquete de dominio**, `internal/usuarios/`, con los archivos de
  la receta divididos por responsabilidad (`service_auth.go`, `handler_users.go`… — desviación
  declarada en el plan). `internal/platform/session` solo contiene el **plumbing** de token/cookie y
  los tipos `Identity`/`Resolver` (sin SQL); la persistencia de sesiones es parte del repository del
  dominio.
- **Justificación**: la spec de F2 es **una** área ("Acceso y gestión de usuarios") y sus tablas se
  consultan juntas a cada paso (login = `users` + `roles` + `permissions`; gestión = las mismas +
  `sessions`). Con dos dominios, R2 ("un dominio no importa a otro") obligaría a consultas
  compartidas en `internal/db/queries/` y a una interfaz extra para revocar sesiones desde la
  gestión de cuentas: reglas adicionales para ningún beneficio hoy. Además `data-model.md` de F1 ya
  preveía estas tablas transversales "en `internal/platform/` + el dominio `usuarios`".
- **Alternativas consideradas**: `auth/` + `usuarios/` (ver arriba; se reabre si F3–F9 necesitan
  autenticación de usuarios públicos con otro modelo); tres dominios `auth/`, `usuarios/`, `roles/`
  (fragmentación sin aislamiento real: comparten tablas); todo en `platform/` (el dominio no es
  plumbing: tiene reglas de negocio de la spec).

## R15. Vida absoluta de la sesión *(propuesta nueva — se confirma junto con D-A7)*

- **Decisión propuesta**: además de la inactividad de **30 minutos** (assumption de la spec,
  ajustable por el cliente), la sesión tiene una **vida absoluta de 12 horas** (`expires_at` fijado
  al iniciar sesión). Superada cualquiera de las dos, hay que volver a iniciar sesión.
- **Justificación**: la spec solo exige expiración por inactividad (FR-005), pero sin vida absoluta
  una cookie robada seguiría viva para siempre mientras haya actividad mínima, y los permisos de
  una sesión "vieja" se revalidan igualmente (P9). Es una medida de seguridad estándar y de coste
  nulo.
- **Estado**: **no está en la spec**; se propone como complemento de seguridad y se marca para
  confirmación humana junto con D-A7 (si no se confirma, se implementa solo la inactividad de 30 min
  y `expires_at` pasa a ser derivado). El valor (12 h) y los 30 min serían configurables por
  variables de entorno con esos defectos.

## R16. Dependencias nuevas (D-A8: justificación por dependencia)

| Dependencia | Dónde | Entradas | Por qué | Alternativa |
|---|---|---|---|---|
| `golang.org/x/crypto` | backend (runtime) | 1 directa (+ `golang.org/x/sys` y/o `x/term` indirectas) | bcrypt: §IV **exige** bcrypt o argon2 y la stdlib no lo trae | argon2id (misma dependencia); hash propio (viola §IV) |
| `github.com/google/uuid` | backend (runtime) | 1 directa, sin transitivas | Tipos UUID del dominio (§8.1.1; R11) | `pgtype.UUID` en el dominio (arrastra `pgx`) |
| `react-hook-form` | frontend | 1 | Convención de la skill para formularios (estado, errores junto al campo) | Estado manual en React (más código y más errores) |
| `zod` | frontend | 1 | Esquemas de validación de formularios **espejo del contrato** (skill) | Validación a mano (duplica y diverge del contrato) |
| `@hookform/resolvers` | frontend | 1 | Puente RHF↔Zod (oficial) | Integración manual |

Ninguna otra dependencia entra en F2. `sqlc`, `golang-migrate` y `openapi-typescript` siguen siendo
herramientas de desarrollo. Todas pasan `govulncheck`/`npm audit` (§IV) en `make ci`.

## R17. `mustChangePassword` y revocación de sesiones

- **Decisión**: `users.must_change_password` es `true` cuando un administrador define o restablece
  la contraseña, y `false` cuando la persona la cambia por sí misma. Mientras sea `true`, la cuenta
  solo puede usar `/auth/session`, `/auth/logout` y `/auth/password` (guard de la cadena del panel);
  la UI redirige al formulario de cambio. **Toda** contraseña definida o restablecida por un
  administrador **revoca las sesiones abiertas** de esa cuenta (y el cambio propio revoca las demás
  sesiones, dejando viva la actual).
- **Justificación**: US3 esc. 6 y US7 esc. 4–5 exigen el cambio inmediato al entrar; el guard lo
  hace verificable en servidor (no solo en la UI). Revocar sesiones en cambios de contraseña es la
  práctica estándar: un tercero con una sesión robada no conserva el acceso tras el restablecimiento
  (complementa FR-012).
- **Alternativas consideradas**: forzar el cambio solo en la UI (FR-016 lo impide: el servidor debe
  verificarlo); campo `password_changed_at` con comparación de fechas (más estado para el mismo
  resultado); no revocar sesiones en el restablecimiento (dejaría abierta la puerta a la sesión
  robada que motivó el restablecimiento).

## R18. Tiempos de respuesta uniformes en el login

- **Decisión**: cuando el correo no existe, el servicio ejecuta igualmente una verificación bcrypt
  contra un hash **ficticio** antes de responder el mismo error genérico. Nada de lo que distingue
  los casos (¿existe la cuenta?, ¿está inactiva?) cambia el tiempo de respuesta de forma medible.
- **Justificación**: FR-003/SC-008 exigen no revelar existencia; la comparación de contraseñas con
  bcrypt cuesta ~250 ms solo cuando la cuenta existe, y esa diferencia **sí** la puede medir un
  atacante (usuario + cronómetro = enumeración). El coste es un hash de más en el caso de error.
- **Alternativas consideradas**: no hacerlo (enumeración por timing, hallazgo seguro para
  `seguridad`); retardar artificialmente todas las respuestas a un fijo (peor: alarga también los
  éxitos y no elimina la señal).
