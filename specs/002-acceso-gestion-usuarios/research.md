# Research — F2 Acceso y gestión de usuarios

**Fecha**: 2026-10-04 · **Rama**: `002-acceso-gestion-usuarios` · **Spec**: `spec.md` (aprobada el
2026-10-04; **cambio de alcance: auditoría** re-aprobado el 2026-10-04 — US8, FR-021…FR-026)

Formato: Decisión / Justificación / Alternativas consideradas. Cierra además las preguntas abiertas
2–5 de `docs/tecnico/decisiones.md` (D-A7, chi, `platform/validate`, tipos UUID). **Numeración**:
las decisiones de este documento son **R1–R23**; la tabla de Riesgos de `plan.md` usa **RG1–RG19**
para que las dos numeraciones no colisionen. **R1–R4 aterrizan
D-A7, CONFIRMADA por el humano el 2026-10-04** con un cambio explícito: **la sesión vive en Redis**
(no en PostgreSQL). Ese día confirmó además los tiempos de sesión (R15: vida absoluta 1 h +
inactividad 30 min) y los datos de la cuenta (nombre, apellidos, correo y teléfono; reflejados en
`data-model.md` y en el contrato). **R21–R23** son del **cambio de alcance: auditoría** (ese mismo
día): Redis **sin persistencia** confirmado, tablas duraderas del registro y sus puntos de escritura.

## R1. Dónde vive la sesión *(D-A7 — CONFIRMADA el 2026-10-04: en Redis)*

- **Decisión**: **la sesión vive en Redis** (decisión explícita del humano el 2026-10-04, con la
  justificación de adoptar/probar Redis como objetivo de aprendizaje). La cookie solo lleva un
  token aleatorio opaco de 32 bytes (`crypto/rand`, base64url); **Redis guarda su SHA-256 como
  parte de la clave, nunca el token en claro**. Revocar es `DEL` de la clave (y de la entrada en el
  índice por usuario). El detalle de claves y TTL está abajo y en `data-model.md`.
- **Diseño de claves (Redis)**:
  - `sess:<sha256(token)>` — string JSON con `userId`, `createdAt`, `lastSeenAt` y
    `absoluteExpiresAt`. **TTL = inactividad (30 min)**, refrescado en cada actividad y acotado al
    tiempo que quede de vida absoluta (`TTL = min(30 min, absoluteExpiresAt - now)`).
  - `user_sessions:<userId>` — `SET` con los hashes de token de las sesiones abiertas de la cuenta
    (TTL ≤ vida absoluta). Permite **revocar todas las sesiones de una cuenta** (FR-012, R17) con
    `SMEMBERS` + `DEL`, sin `SCAN` ni tabla.
  - Los `lastSeenAt` se escriben estrangulados a una vez por minuto (misma idea que el
    `last_seen_at` de la propuesta anterior; solo cambia el almacén).
- **Justificación**:
  - FR-012 exige que desactivar una cuenta corte el acceso **de inmediato, también las sesiones
    abiertas**: con sesión en servidor eso es un `DEL` de sus claves (y, como red de seguridad,
    `authn` revalida que la cuenta siga activa en cada petición). Con un token autocontenido (JWT)
    habría que mantener una lista negra, que es… una sesión en servidor con otro nombre.
  - **Es la decisión explícita del humano de adoptar Redis** (D-A8: dependencia justificada por
    elección humana, con el objetivo declarado de aprender/usar Redis en el proyecto). No se
    argumenta aquí contra ella: la alternativa PostgreSQL queda documentada abajo como descartada
    por esa decisión, y es el plan B si Redis resultara inviable.
  - Redis es la herramienta correcta para este estado: efímero, con **expiración nativa por TTL**
    (sin limpieza manual ni cron) y compartido entre instancias si algún día hay más de una.
  - Guardar el **hash** del token y no el token evita que una lectura indebida de Redis permita
    secuestrar sesiones.
- **Alternativas consideradas**:
  - *Tabla PostgreSQL `sessions`* (la propuesta original de D-A7): **descartada por la decisión
    explícita del humano el 2026-10-04 de vivir en Redis**. Sus ventajas quedan registradas para el
    plan B: cero servicios nuevos, transaccional con `users`, durabilidad y revisión por SQL. Es la
    opción a la que se volvería si Redis no fuera viable (ver riesgo RG15/RG16 del plan).
  - *JWT autocontenido en cookie*: rechazado (y ya lo estaba en D-A7) — revocación al desactivar
    forzada con lista negra; más superficie (algoritmo, expiración, refresco) para el mismo
    resultado.
  - *Cookie firmada sin estado (clave-valor firmado con `SESSION_SECRET`)*: rechazado — no permite
    revocación individual real ni "cerrar todas las sesiones".
  - *Token en `localStorage`*: rechazado — expuesto a XSS (CWE-79) y prohibido por la skill.
  - *Framework de auth completo (OIDC/Keycloak)*: rechazado — peso desproporcionado para un panel
    interno, y exige operar un IdP.
- **Consecuencias**: servicio `redis` en `docker-compose.yml`, dependencia de runtime
  `github.com/redis/go-redis/v9` (justificada en R16), `internal/platform/session` con la interfaz
  `Store` y su implementación Redis (sin SQL), **sin** tabla `sessions` ni `sessions.sql`, y las
  pruebas de integración de sesión con `testcontainers-go` (R19). `middleware.Authn` resuelve la
  identidad por petición igual que antes. **Estado: confirmada por el humano el 2026-10-04.**

## R2. La cookie de sesión y sus flags *(D-A7 — confirmada el 2026-10-04)*

- **Decisión**: cookie **`ss_session`** con `HttpOnly`, `SameSite=Lax`, `Path=/`, `Secure` según
  `SESSION_COOKIE_SECURE` (en `false` por defecto en desarrollo; `platform/config` **falla al
  arrancar** si `APP_ENV=production` y está en `false`). Valor: 32 bytes de `crypto/rand` en
  base64url; caduca junto con la sesión (**inactividad 30 min + vida absoluta 1 h**, R15) y su
  `Max-Age` se fija a la **vida absoluta (3600 s)**, de modo que el navegador la descarta en el
  mismo instante en que el servidor la daría por vencida. Se emite en la respuesta de
  `POST /api/v1/auth/login` y se borra en `logout` y en toda revocación.
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

## R3. CSRF en las rutas con sesión *(D-A7 — confirmada el 2026-10-04)*

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

## R4. Hash de contraseñas: bcrypt *(D-A7 — confirmada el 2026-10-04)*

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
- **Política FR-010 (implementación)**: longitud **entre 8 y 64 caracteres** (el tope del contrato
  es `maxLength: 64` en **todos** los campos de contraseña), al menos una mayúscula, una minúscula,
  un número y un carácter especial, y **distinta de** —igualdad con comparación normalizada
  (trim + minúsculas), **no de contenido**: la contraseña puede contener esos datos; lo que se
  rechaza es que sea **igual** a cualquiera de ellos— el nombre, los apellidos y el correo. El error
  identifica el requisito incumplido (`details.newPassword`). El límite duro de bcrypt (72 bytes) se
  mantiene solo como **comprobación técnica** (solo puede chocar con contraseñas de 64 caracteres
  *multi-byte*) y se rechaza con mensaje claro; no forma parte de la política que ve la persona ni
  del contrato.
- **Alternativas consideradas**:
  - *argon2id*: válida por §IV y más resistente a GPU/ASIC, pero añade parámetros que afinar
    (memoria, iteraciones, paralelismo) sin una necesidad hoy; además usa **la misma** dependencia.
    Queda como evolución natural (migración transparente por el prefijo del hash).
  - *pbkdf2/scrypt de la stdlib*: rechazado — §IV dice literalmente bcrypt o argon2.
  - *Un hash "propio" (SHA-256 + sal)*: rechazado — viola §IV y es criptográficamente insuficiente
    (rápido de fuerza bruta).
- **Constantes de código**: `bcryptCost = 12`, longitud **8–64 caracteres** (el límite de 72 bytes
  de bcrypt solo como comprobación técnica). No se pasan por
  variables de entorno: son exigencias de la política confirmada por el humano (2026-10-04), no
  ajustes operativos.

## R5. Bloqueo por intentos fallidos (FR-006) sin enumerar cuentas (FR-003) *(contadores en Redis)*

- **Decisión**: los contadores viven en **Redis**, con una entrada **por identificador normalizado
  (el correo), exista o no la cuenta**. Dos claves: `login:fail:<identificador>` (contador que se
  incrementa con `INCR` en cada fallo, TTL 15 min) y `login:block:<identificador>` (bandera de
  bloqueo con TTL 15 min, creada por el **5.º** fallo). **Semántica del 5.º intento (FR-006)**: el
  contador se incrementa con **cada** fallo; el **5.º fallo** responde el error genérico `401
  unauthenticated` **y crea el bloqueo**; desde el **6.º intento** —y durante esos 15 minutos— cada
  intento responde `429 rate_limited` con el mensaje de bloqueo y `Retry-After` = TTL restante.
  Fuera de bloqueo, la respuesta es siempre el mismo `401` genérico. Un inicio de sesión correcto
  borra las dos claves; el bloqueo vencido lo borra el propio TTL (sin limpieza manual). Los
  valores **5** y **15 minutos** siguen siendo **constantes de código** (`maxFailedAttempts`,
  `lockoutDuration`) porque el humano los confirmó el 2026-10-04.
- **Justificación**: FR-006 pide bloquear tras 5 intentos y FR-003/SC-008 piden que ningún mensaje
  revele si la cuenta existe. Si el contador viviera solo en `users`, el mensaje de bloqueo
  **solo** aparecería para cuentas reales → enumeración garantizada. Contando por identificador
  (haya cuenta o no) el comportamiento —mensaje, código y tiempos— es **idéntico** en ambos casos y
  las dos exigencias se cumplen a la vez. Con Redis el estado es efímero por definición (la tabla
  propuesta antes se auto-borraba igual al resolverse: no tenía valor duradero), su TTL es
  exactamente el mecanismo de expiración que la regla necesita y, como la sesión ya vive en Redis
  (R1), no se añade ninguna pieza nueva.
- **Alternativas consideradas**:
  - *Tabla PostgreSQL `login_attempts`* (la propuesta anterior): descartada al confirmarse la
    sesión en Redis — exigiría una migración, consultas sqlc y limpieza manual para un estado que
    no debe durar. Queda como plan B junto con la tabla `sessions` si Redis no fuera viable.
  - *Contador solo en `users`* (enumeración); *contador solo en memoria del proceso* (se pierde al
    reiniciar y no sobrevive a varias instancias — ahora que Redis existe, no hay razón); *rate-limit
    por IP como única medida* (no frena la prueba masiva de contraseñas contra **una** cuenta desde
    una IP); *bloqueo permanente hasta intervención de un administrador* (no lo pide la spec y
    bloquea al usuario legítimo).
- **Mensajes** (siempre en español, sin datos internos): intentos 1–5 → *"Correo o contraseña
  incorrectos"* (genérico; el 5.º crea el bloqueo pero recibe este mismo error); desde el 6.º
  intento y durante el bloqueo → *"Demasiados intentos fallidos. El acceso queda bloqueado
  temporalmente durante 15 minutos"*, con `Retry-After` en segundos.
- **Riesgo registrado**: si Redis se reinicia sin persistencia, se pierden los contadores (y las
  sesiones, que solo obligan a volver a entrar). Un atacante no puede forzar ese reinicio, así que
  la pérdida es aceptable; **la persistencia de Redis queda descartada a propósito** (sin
  `appendonly`, sin volumen — confirmado por el humano el 2026-10-04, R21) y **el registro duradero
  no depende de Redis**: cada intento deja fila en PostgreSQL (`login_events`, R22), de modo que un
  reinicio no borra el historial.

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
  `quickstart.md` §1. Si el humano **no** quiere el token (RG13 del plan), se retira esta única
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

- **Decisión**: en PostgreSQL solo quedan `users`, `roles`, `permissions` y `role_permissions`.
  **Se elimina `user_roles`** (la referencia orientativa de F1 `data-model.md` §"F2" lo incluía):
  una cuenta tiene **un solo rol** (decisión Q4), así que la relación vive en `users.role_id`.
  **Se añade `permissions`** como catálogo fijo sembrado por la migración (los 9 módulos de
  FR-015, incluidos los reservados de F3–F9). **`sessions` y `login_attempts` no son tablas**: la
  sesión vive en Redis (R1) y los contadores de intentos de acceso también (R5), por decisión
  confirmada del humano el 2026-10-04. *(Actualizado por el cambio de alcance: la auditoría añade
  además `login_events` y `admin_actions` y las columnas `users.last_login_*` — R22—; lo que sigue
  en pie aquí es la simplificación de `user_roles` y el catálogo de permisos.)* Detalle completo en
  `data-model.md` (tablas y claves Redis).
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
  de inicialización (DOS a bcrypt, RG14 del plan).
- **Limitación declarada**: en memoria no comparte umbrales entre instancias (RG5 del plan). Hoy hay
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
  declarada en el plan). `internal/platform/session` contiene el **plumbing** de token/cookie, los
  tipos `Identity`/`Resolver`, la interfaz `Store` de sesiones y su implementación sobre Redis
  (sin SQL ni conocimiento del dominio: las claves solo guardan el `userId` y marcas de tiempo);
  el dominio resuelve la identidad (cuenta → rol → permisos) y revoca sesiones por cuenta a través
  de `session.Store`.
- **Justificación**: la spec de F2 es **una** área ("Acceso y gestión de usuarios") y sus tablas se
  consultan juntas a cada paso (login = `users` + `roles` + `permissions`; gestión = las mismas).
  Con dos dominios, R2 ("un dominio no importa a otro") obligaría a consultas
  compartidas en `internal/db/queries/` y a una interfaz extra para revocar sesiones desde la
  gestión de cuentas: reglas adicionales para ningún beneficio hoy. Además `data-model.md` de F1 ya
  preveía estas tablas transversales "en `internal/platform/` + el dominio `usuarios`".
- **Alternativas consideradas**: `auth/` + `usuarios/` (ver arriba; se reabre si F3–F9 necesitan
  autenticación de usuarios públicos con otro modelo); tres dominios `auth/`, `usuarios/`, `roles/`
  (fragmentación sin aislamiento real: comparten tablas); todo en `platform/` (el dominio no es
  plumbing: tiene reglas de negocio de la spec).

## R15. Vida absoluta de la sesión *(CONFIRMADA el 2026-10-04: 1 hora)*

- **Decisión**: además de la inactividad de **30 minutos** (assumption de la spec), la sesión tiene
  una **vida absoluta de 1 hora desde el inicio de sesión** (ambos valores confirmados por el
  humano el 2026-10-04). Superada cualquiera de las dos, hay que volver a iniciar sesión.
- **Cómo se implementa sobre Redis (R1)**: una sola clave por sesión con **TTL de inactividad de
  30 min** que se refresca en cada actividad, y dentro del valor un `absoluteExpiresAt = createdAt +
  1 h` que nunca se mueve. En cada petición: si la clave no existe → sesión caducada por
  inactividad; si existe pero `now >= absoluteExpiresAt` → caducada por vida absoluta (se borra la
  clave); si sigue viva → se refresca el TTL **acotado** al tiempo que quede de vida absoluta
  (`TTL = min(30 min, absoluteExpiresAt - now)`), de modo que Redis expira la clave como muy tarde
  a la hora de nacer. La cookie se emite con `Max-Age` de 1 h (R2) y el cliente puede mostrar la
  cuenta atrás con las mismas marcas.
- **Justificación**: la spec solo exige expiración por inactividad (FR-005), pero sin vida absoluta
  una cookie robada seguiría viva para siempre mientras haya actividad mínima. La vida absoluta de
  1 h es corta a propósito (panel interno, re-login barato) y el mecanismo —TTL de Redis + fecha
  límite inmóvil— es simple de explicar y de probar. Los dos valores se leen de
  `SESSION_IDLE_TTL_MINUTES` (30) y `SESSION_ABSOLUTE_TTL_MINUTES` (60) con esos defectos.
- **Alternativas consideradas**: solo inactividad (una cookie robada viviría lo que durara la
  actividad); dos claves Redis por sesión (una por expiración: dos viajes y dos TTL que
  sincronizar); TTL absoluto de 1 h sin refresco de inactividad (mataría sesiones legítimas activas
  a la hora justa aunque haya actividad); refresco de vida absoluta por actividad (es un "sliding
  window" total: vuelve a ser indefinida).
- **Estado**: **confirmada por el humano el 2026-10-04** (1 h absoluta + 30 min de inactividad,
  valores que sustituyen a los 12 h propuestos inicialmente).

## R16. Dependencias nuevas (D-A8: justificación por dependencia)

| Dependencia | Dónde | Entradas | Por qué | Alternativa |
|---|---|---|---|---|
| `golang.org/x/crypto` | backend (runtime) | 1 directa (+ `golang.org/x/sys` y/o `x/term` indirectas) | bcrypt: §IV **exige** bcrypt o argon2 y la stdlib no lo trae | argon2id (misma dependencia); hash propio (viola §IV) |
| `github.com/google/uuid` | backend (runtime) | 1 directa, sin transitivas | Tipos UUID del dominio (§8.1.1; R11) | `pgtype.UUID` en el dominio (arrastra `pgx`) |
| `github.com/redis/go-redis/v9` | backend (runtime) | 1 directa (+ `cespare/xxhash` y `dgryski/go-rendezvous` indirectas) | Cliente de Redis para la sesión y los contadores de acceso (R1/R5). **Justificado bajo D-A8 por decisión explícita del humano el 2026-10-04: adoptar Redis como objetivo de aprendizaje**; go-redis es el cliente de referencia, mantenido y con soporte de TTL/pipelines | La alternativa técnica era **no** usar Redis (sesión y contadores en PostgreSQL); queda descartada por esa decisión humana y documentada como plan B. Otros clientes (`rueidis`, cliente RESP propio) sin ventaja para este uso |
| `github.com/testcontainers/testcontainers-go` | backend (**solo** pruebas `//go:build integration`) | 1 directa (+ transitivas del SDK de Docker) | Levanta Redis (y PostgreSQL si hace falta) **dentro de las pruebas de integración**: `.github/workflows/ci.yml` es del kit y **no se puede editar**, y F2 necesita Redis en esas pruebas (R19) | Editar `ci.yml` con un *service* de Redis (**prohibido**: archivo del kit); `miniredis` (reimplementación en memoria: no valida el cliente real ni los TTL reales); exigir Redis manual en cada entorno (no reproducible) |
| `react-hook-form` | frontend | 1 | Convención de la skill para formularios (estado, errores junto al campo) | Estado manual en React (más código y más errores) |
| `zod` | frontend | 1 | Esquemas de validación de formularios **espejo del contrato** (skill) | Validación a mano (duplica y diverge del contrato) |
| `@hookform/resolvers` | frontend | 1 | Puente RHF↔Zod (oficial) | Integración manual |

Ninguna otra dependencia entra en F2. `sqlc`, `golang-migrate` y `openapi-typescript` siguen siendo
herramientas de desarrollo; `testcontainers-go` se enlaza **solo** en el build de integración (no
llega al binario de producción). Todas pasan `govulncheck`/`npm audit` (§IV) en `make ci`.

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

## R19. Redis en las pruebas de integración sin tocar el CI del kit

- **Decisión**: las pruebas de integración que necesitan Redis (sesiones, contadores de acceso) y
  las de repositorio levantan sus servicios con **`testcontainers-go` dentro del propio proceso de
  prueba** (`//go:build integration`): `redis:7-alpine` **siempre**, y PostgreSQL desde
  `DATABASE_URL_TEST` si la variable existe (el CI del kit ya lo aporta como *service*) o también
  con testcontainers si no. El flujo local puede usar el servicio `redis` de `docker-compose.yml`
  vía `REDIS_URL`; el CI **no cambia**.
- **Justificación**: `.github/workflows/ci.yml` está en `.kit-manifest.json` y **no se puede
  editar** (regla 10 de AGENTS.md), pero su job de backend ya ejecuta
  `go test -tags=integration ./...` y los runners `ubuntu-latest` de GitHub Actions tienen Docker
  disponible. testcontainers gestiona el ciclo de vida (arranque, puerto, limpieza) desde el
  propio test, así que la integración con Redis es reproducible en cualquier entorno sin tocar el
  kit. `testcontainers-go` queda como dependencia **de desarrollo** justificada bajo D-A8 (R16).
- **Alternativas consideradas**: editar `ci.yml` añadiendo un *service* de Redis (prohibido: archivo
  del kit; y su hash está registrado en `.kit-manifest.json`); proponer el cambio al repositorio del
  kit (vía válida para el futuro, pero bloquearía F2 y no es decisión de esta fase);
  `miniredis` como Redis embebido (reimplementación en memoria: sirve para unit tests, pero no
  valida el cliente real ni los TTLs de verdad); exigir un Redis manual en cada entorno (no
  reproducible y frágil).
- **Consecuencias**: los tests de integración exigen Docker (ya lo exigía `make up`); el primer
  arranque descarga `redis:7-alpine` (~15 MB); si el runner no tuviera Docker o fallara la descarga
  de imágenes, la integración fallaría **sin** poder remediarse desde el CI — registrado como
  riesgo RG15 del plan.

## R20. Datos de la cuenta: nombre, apellidos, correo y teléfono *(confirmados el 2026-10-04)*

- **Decisión**: la cuenta del panel tiene **nombre** (`users.first_name`), **apellidos**
  (`users.last_name`), **correo** (`users.email`, identificación de acceso) y **teléfono**
  (`users.phone`); los cuatro obligatorios, con nombre y apellidos como campos separados. La spec
  ya está actualizada (FR-009, FR-011 y el glosario) y esto cierra el riesgo RG3 del plan
  ("identificación" ambigua) y la duda sobre un eventual documento de identidad: **no hace falta**.
- **Justificación**: son los datos que el equipo de la iglesia necesita para identificar a cada
  persona y contactarla; el teléfono se valida con un formato telefónico razonable (FR-009:
  dígitos con espacios, guiones o paréntesis, prefijo internacional opcional y **al menos 7
  dígitos**) en `platform/validate` (etiqueta `phone`), sin normalizar su forma de presentación
  (no participa en unicidad). La política de contraseñas se compara contra nombre, apellidos y
  correo (FR-010).
- **Alternativas consideradas**: un solo campo `full_name` (la spec pide nombre y apellidos
  separados); documento de identidad aparte (no lo pide la spec; quedaría como columna nueva si
  algún día aparece); teléfono opcional (la spec lo hace obligatorio); guardar solo dígitos o
  normalizar a E.164 (se pierde la forma de presentación sin ganar unicidad ni búsquedas).

## R21. Redis sin persistencia *(CONFIRMADO por el humano el 2026-10-04)*

- **Decisión**: el servicio `redis` de `docker-compose.yml` corre **sin persistencia**: **sin
  `appendonly` (AOF), sin `save` (RDB) y sin volumen**. Es estado efímero por definición: un
  reinicio de Redis (o de su contenedor) borra las sesiones abiertas y los contadores de intentos,
  y **nada más**. Quien tenía sesión vuelve a iniciar sesión; el bloqueo de FR-006 empieza de cero.
- **Justificación**: es la decisión confirmada del humano (2026-10-04), coherente con lo que Redis
  guarda en F2: sesiones (ya caducables a las 2 h como mucho) y contadores con TTL de 15 min. Ninguno
  de los dos merece durar un reinicio; persistirlos solo añadiría operativa (volúmenes, AOF,
  copias) para datos que se reconstruyen con un re-login. Lo que sí debe durar —el registro de
  auditoría (R22)— vive en **PostgreSQL**, que ya es el almacén duradero del proyecto.
- **Qué pasa exactamente al reiniciar Redis**: (1) todas las sesiones desaparecen → hay que volver
  a entrar (la cookie deja de resolver); (2) los contadores `login:fail:*`/`login:block:*` se
  borran → la ventana de 5 intentos / 15 min se reinicia; (3) el **historial de accesos, el de
  acciones administrativas y el último acceso de cada cuenta siguen intactos**, porque están en
  PostgreSQL. Se comprueba en `quickstart.md` §10 (`docker compose restart redis`).
- **Alternativas consideradas**: *AOF (`appendonly yes`)* y *RDB (snapshots `save`)* con volumen
  —**descartados por decisión humana**: nada de lo que guarda Redis tiene valor duradero y la dureza
  que importa va a PostgreSQL—; guardar el registro de auditoría también en Redis (contradice
  FR-025: el registro debe sobrevivir reinicios y purgas de caché); un Redis con `maxmemory` y
  política de evicción (innecesario con este volumen; se reabre si crece).
- **Estado**: confirmado por el humano el 2026-10-04. Aterrizado en el plan (P23, riesgo RG16),
  `data-model.md` (sección Redis ↔ PostgreSQL) y `quickstart.md` §3 y §10.

## R22. Auditoría duradera: `login_events`, `admin_actions` y el último acceso por cuenta

- **Decisión**: el registro de la auditoría de F2 son **dos tablas de PostgreSQL** de solo inserción
  (detalle completo —columnas, `CHECK`, índices, `up`/`down`— en `data-model.md`, migración
  `000004`):
  - `login_events` — una fila por **intento de inicio de sesión** (FR-022): fecha y hora
    (`created_at`), resultado (`success`/`failure`), IP de origen y `user_id` **anulable** (solo
    cuando la cuenta existe; un correo inexistente deja fila sin asociación y sin crear nada,
    FR-003).
  - `admin_actions` — una fila por **acción administrativa sensible** (FR-023): quién
    (`actor_user_id`, anulable solo en la inicialización), qué (`action`, los ocho códigos de la
    spec), sobre qué (`target_kind` + FK a `users`/`roles` + `target_label` con la etiqueta del
    momento), cuándo (`created_at`) y resultado (`success`/`failure`/`denied`).
  - El **último acceso exitoso** (FR-021) son dos columnas en `users` (`last_login_at`,
    `last_login_ip`) escritas solo por el login exitoso: la ficha y el listado las muestran sin
    consultar el historial y siguen funcionando aunque en el futuro haya retención del registro.
- **Justificación**: la spec (cambio de alcance) pide que el registro **dure** y se conserve aunque
  la cuenta se desactive o se edite (FR-025), así que no puede vivir en Redis (R21) ni ser un log
  que se rote. PostgreSQL da dureza, consultas con filtros (FR-024) y revisión por SQL. La
  proyección del último acceso en `users` evita una subconsulta por cuenta en cada listado y
  desacopla la ficha del crecimiento del historial. `target_label` (y no solo la FK) hace que el
  "sobre qué" del registro siga respondiéndose cuando un rol se elimina (FR-017): la FK es
  `ON DELETE SET NULL` y la etiqueta capturada se conserva.
- **Mínimo dato necesario (FR-026, constitución §IV)**: el registro no contiene contraseñas ni
  credenciales —de un login solo su resultado y de un restablecimiento solo quién/sobre qué/cuándo—
  y **tampoco** el correo de un intento no identificado (evita "cuentas fantasma" y datos de
  terceros), ni *user-agent*, ni la IP de las acciones administrativas (no la pide la spec). La IP
  se toma de `RemoteAddr` (`net.SplitHostPort`); `X-Forwarded-For` queda fuera hasta que haya un
  proxy documentado (decisión futura). Los nombres y correos que muestra el contrato
  (`userName`/`userEmail` en los accesos, `actorName`/`actorEmail` en las acciones) se **derivan por
  `JOIN` con `users`** al consultar —no se guardan como dato duplicado— y son `null` cuando no hay
  cuenta asociada: en ese caso **no se guarda ni se muestra el correo probado** y la interfaz
  muestra **"Intento sin cuenta asociada"**.
- **Alternativas consideradas**: *una sola tabla `audit_log` polivalente* (campos según el tipo de
  evento: casi todo anulable y con `CHECK` frágiles; la spec distingue dos historiales con datos
  distintos); *log estructurado en archivos* (sin consultas ni filtros, difícil de paginar y de
  proteger con permisos); *Redis Streams / lista con TTL* (se purgan solas: contradicen FR-025);
  *derivar el último acceso del historial* (ver arriba: consulta extra y dependencia de la
  retención); *guardar el correo de todo intento* (PII de terceros y "cuentas fantasma");
  *snapshot completo de los datos cambiados en cada acción* (JSONB con antes/después) — la spec lo
  marca como "deseable pero revisable" y no requisito del MVP, así que queda como ampliación.

## R23. Dónde se escribe el registro (y qué pasa si falla)

- **Decisión**: el registro se escribe desde el dominio `internal/usuarios` (`service_audit.go` +
  `repository_audit.go`), que es el único dueño de las tablas. Hay **tres puntos de escritura**,
  porque FR-022/FR-023 piden registrar **todo** intento y **toda** acción, también las que fallan o
  se deniegan:
  1. `service_auth.go` — cada intento de login deja su fila en `login_events` en los cuatro
     desenlaces: éxito (que además actualiza `last_login_*`), credenciales incorrectas, cuenta
     inactiva e intento durante un bloqueo temporal (el Edge Case de la spec lo exige).
  2. Los services de gestión (`service_users.go`, `service_roles.go`, `service_auth.go` en la
     inicialización) — cada operación sensible deja su fila en `admin_actions` **también al
     fallar** (duplicado, rol en uso, regla anti-bloqueo, no encontrado, política de contraseña);
     en el éxito, el `INSERT` va en la **misma transacción** que la mutación.
  3. Lo que muere antes del service: el helper común de decodificación/validación del `handler.go`
     registra el JSON inválido o el DTO no válido, y `middleware.AuthzByModule` registra la
     denegación por falta de permiso (`result='denied'`) vía la interfaz de plumbing
     `audit.Recorder` (`internal/platform/audit`, mismo precedente que `session.Identity`/`Resolver`
     de R14): la acción y el objetivo se resuelven desde `method`+`path` con una tabla del dominio,
     de modo que `platform` sigue sin conocer dominios (R1).
- **Si el registro falla**: si la operación **iba a completarse**, su registro es requisito —misma
  transacción: o ambos o ninguno, y el login exitoso deja su fila antes de emitir la sesión—. Si
  falla el registro de un **intento que ya va a fallar** (un login con mala contraseña, una
  operación denegada), se loguea el error con `request_id` y se devuelve el error original: el
  registro nunca cambia la respuesta que ve la persona.
- **Límite declarado**: lo que se rechaza antes de identificar una operación de gestión —CSRF
  inválido, sesión inexistente o expirada (401), `rate-limit`, guard de cambio de contraseña— **no**
  entra en `admin_actions`: no son acciones administrativas con actor y objeto identificables y
  quedan en el log de la aplicación con su `request_id`. Es un default revisable (si se quisiera un
  registro de seguridad más allá de la auditoría de gestión, sería una ampliación).
- **Alternativas consideradas**: *registrar solo desde el handler con un decorador por ruta* (no ve
  el id recién creado ni distingue el objetivo de una creación, y las denegaciones ocurren fuera);
  *un middleware único que mire el código de estado y lea la respuesta* (acoplado a la forma de los
  DTO y a los cuerpos de respuesta); *registrar solo los éxitos* (contradice el Edge Case de la
  spec); *triggers de BD sobre las tablas de negocio* (la lógica de qué es una acción sensible vive
  fuera del service, es difícil de probar y de traducir a `apperr`, y no cubre los fallos de
  validación que ocurren antes de tocar la base).
- **Consecuencia para F3–F9**: toda operación sensible nueva de otros módulos deberá añadir su
  código a la tabla de acciones y su punto de registro (hoy fuera de alcance: la auditoría cubre
  solo el ámbito de F2, Out of Scope de la spec). Queda anotado para `revisor-codigo`.
