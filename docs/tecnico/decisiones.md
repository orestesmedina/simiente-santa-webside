# Registro de decisiones de arquitectura

**Fecha:** 2026-09-30 · **Formato:** ADR corto (contexto · decisión · alternativas · consecuencias)
· **Documento técnico asociado:** [arquitectura.md](./arquitectura.md)

Decisiones tomadas por el humano el 2026-09-30 sobre la arquitectura técnica del proyecto
*Sitio web de la Iglesia Simiente Santa*. Son la referencia para `plan.md`, `tasks.md` y el código
de F2…F9. Donde difieren del plan vigente de F1 (`specs/001-estructura-base/plan.md`), **manda este
registro**; las tres diferencias conocidas están en D-A3 (sqlc como capa de datos fijada), D-A4
(interfaz `Registrar` que neutraliza el router) y D-A5 (Go 1.27 en lugar de 1.23).

## Tabla resumen

| ID | Decisión (una línea) | Estado | Fecha |
|---|---|---|---|
| D-A1 | Monorepo `backend/` + `frontend/`, API REST JSON, contrato OpenAPI como única frontera | Aceptada | 2026-09-30 |
| D-A2 | Arquitectura por capas + plataforma interna propia (`internal/platform/`) en vez de framework de terceros | Aceptada | 2026-09-30 |
| D-A3 | Capa de datos: **sqlc** sobre PostgreSQL; stored procedures solo como excepción justificada dentro del `repository` | Aceptada | 2026-09-30 |
| D-A4 | Router: `net/http` en F1 tras la interfaz `Registrar`; chi queda como opción de F2 | Aceptada (revisión en F2) | 2026-09-30 |
| D-A5 | Versión del lenguaje: **Go 1.27** | Aceptada | 2026-09-30 |
| D-A6 | Config por variables de entorno (stdlib), logs `log/slog` JSON, errores stdlib + errores de dominio tipados, inyección de dependencias manual | Aceptada | 2026-09-30 |
| D-A7 | Sesiones de F2: cookie `httpOnly` con sesión en servidor **en Redis**, hash bcrypt, permisos por módulo en tablas propias | **Aceptada — confirmada por el humano el 2026-10-04 (sesión en Redis)** | 2026-09-30 · confirmada 2026-10-04 |
| D-A8 | Dependencias mínimas: en F1 solo `pgx` entra al binario; toda dependencia nueva se justifica en el `plan.md` que la introduce | Aceptada | 2026-09-30 |
| D-A9 | Diferidos a propósito: métricas/trazas, colas de trabajo, almacenamiento de imágenes, i18n y generación de handlers desde el contrato | Aceptada (diferimientos) | 2026-09-30 |

---

## D-A1 · Monorepo con API REST JSON y contrato OpenAPI como única frontera

- **Contexto.** Un solo producto, un equipo pequeño y agentes que trabajan por capas. La
  constitución (§II) ya fija el monorepo `backend/` (Go) + `frontend/` (React + TypeScript) +
  `backend/migrations/` (PostgreSQL) y exige que la comunicación sea una API REST JSON
  documentada en OpenAPI, escrita **antes** que el código.
- **Decisión.** Un único repositorio con ambos extremos; la **única** frontera entre frontend y
  backend es el contrato `backend/api/openapi.yaml`. Cada funcionalidad diseña su *delta* en
  `specs/<N>-<feature>/contracts/` y lo fusiona en el documento vivo al implementar (reglas de
  `specs/001-estructura-base/plan.md`, sección "Contrato OpenAPI"). El frontend consume el
  contrato, nunca el código Go.
- **Alternativas consideradas.**
  - *Microservicios o repositorios separados*: rechazado — sin necesidad de escalado ni equipos
    independientes; multiplica CI, despliegue y coordinación.
  - *GraphQL*: rechazado — el modelo de contenido del roadmap (F3…F9) es simple y REST basta;
    GraphQL añade servidor, tooling y complejidad sin ventaja hoy.
  - *tRPC / RPC tipado Go↔TS*: rechazado — exigiría un puente de generación propio y alejaría el
    contrato del estándar que la constitución pide.
  - *Tipos compartidos "a mano" en ambos lados*: rechazado — divergen del contrato garantizado.
- **Consecuencias.**
  - El contrato es artefacto de diseño **y** documento vivo: cambios de API se deciden primero en
    el YAML y luego en código.
  - Los tipos TypeScript se **generan** (`openapi-typescript` → `frontend/src/api/schema.d.ts`),
    con lo que un cambio de contrato rompe la compilación del frontend si no se regenera.
  - No hay tipos compartidos directamente Go↔TS: la pérdida de "un solo idioma" se compensa con
    validación en ambos extremos (constitución §IV: validar en backend aunque el frontend valide).

## D-A2 · Arquitectura por capas + plataforma interna propia (el "casi framework")

- **Contexto.** F2…F9 repetirán el mismo patrón (handler → service → repository, errores, HTTP,
  config, pruebas) y quien escribe el código son agentes que necesitan convenciones fijas y
  verificables. La constitución (§II) exige capas y minimizar dependencias.
- **Decisión.** Arquitectura por capas (`handler → service → repository`) en cada dominio, más una
  **plataforma interna propia** en `internal/platform/` (`config`, `logger`, `apperr`,
  `httpserver`, `middleware`, `database`, `migrate`, `validate`, `paginate`, `testutil`) que
  concentra lo transversal. Es el "casi framework" del proyecto: pequeño, escrito aquí, sin
  dependencias de terceros. Reglas completas y árbol de carpetas en
  [arquitectura.md](./arquitectura.md) §1–§2.
- **Alternativas consideradas.**
  - *Framework de aplicación de terceros (ver D-A4 para el router)*: rechazado — fija estilo,
    acopla el código a su ciclo de vida y añade dependencias que §II pide minimizar.
  - *Kit/módulo compartido externo (un "starter" de la empresa)*: pospuesto — hoy solo existe este
    proyecto; extraer `platform` a un módulo reutilizable solo tiene sentido cuando aparezca un
    segundo consumidor (regla de tres).
  - *Copiar plantillas por proyecto*: rechazado — divergen y no se corrigen en todos los sitios.
- **Consecuencias.**
  - Cero dependencias de framework: todo el código transversal es nuestro y se revisa aquí.
  - Somos los dueños del mantenimiento de `platform`: sus paquetes se tratan como **contrato
    interno** (cambios con revisión de `revisor-codigo` y actualización de este registro).
  - Coste de entrada mínimo para nuevos agentes: aprenden ~10 paquetes pequeños y una regla de
    dependencias, no un framework entero.

## D-A3 · Capa de datos: sqlc sobre PostgreSQL (stored procedures solo como excepción)

- **Contexto.** Hace falta acceso a datos tipado y seguro (constitución §IV: solo consultas
  parametrizadas; §VI: migraciones versionadas). El plan de F1 posponía la decisión ("sqlc en F2
  sin decidirla"); aquí queda **decidida** como capa de datos estándar del proyecto.
- **Decisión.** **sqlc** sobre PostgreSQL, leyendo **nuestras migraciones** (`backend/migrations/`)
  como esquema, con el SQL explícito y revisable en `internal/db/queries/*.sql` y el código
  generado commiteado en `internal/db/`. Las migraciones siguen mandando con `golang-migrate`:
  sqlc las lee, no las sustituye. **Stored procedures / funciones SQL solo como excepción
  justificada**, y siempre invocados dentro del `repository` (nunca en el service, nunca fuera de
  una capa de datos). Nota operativa: un procedimiento invocado con `CALL` dentro de un bloque de
  transacción **no** puede ejecutar control de transacciones (no vale `BEGIN`/`COMMIT` dentro de
  él); la transacción la gestiona el código Go (`database.WithTx`).
  Su entrada al código llega con la **primera consulta de negocio** (F1 solo hace un `Ping` de
  salud con `pgx` directo; ver ejemplo completo en [arquitectura.md](./arquitectura.md) §5.3).
- **Comparativa medida el 2026-09-30** (entradas que cada librería añadiría a `go.mod`, incluidas
  sus dependencias de test):

  | Opción | Entradas nuevas en `go.mod` | Validación en compilación | Migraciones | Comentario |
  |---|---|---|---|---|
  | **sqlc (elegida)** | **0 en runtime** (es una herramienta externa: no se enlaza en el binario) | **Sí** — los tipos se generan del SQL real | Usa nuestras migraciones como esquema; `golang-migrate` sigue mandando | SQL explícito, revisable en el PR |
  | GORM | 5 (+9 del driver oficial) | No (reflection) | `AutoMigrate` impone su mecanismo, **choca** con `golang-migrate` | ORM completo; oculta el SQL |
  | ent | 49 | Sí | Atlas (su mecanismo), **choca** con `golang-migrate` | Framework de entidades completo |
  | sqlx | 3 | No (reflection parcial) | — | Finito, pero sin generación ni tipos derivados del SQL |

- **Alternativas consideradas.** Las cuatro filas de la tabla; además, *escribir a mano mappers
  sobre `pgx`* (rechazado: trabajo repetido y propenso a errores, sin validación en compilación).
- **Consecuencias.**
  - Un cambio de consulta o de esquema que no acompañe **falla en compilación** (seguridad de
    tipos desde el SQL real).
  - El SQL vive en archivos `.sql` revisables; no hay cadenas SQL incrustadas en Go.
  - Dos artefactos generados se commitean y se regeneran a mano (`sqlc generate`,
    `npm run api:gen`): mismas garantías que generarlo en CI, sin tocar el `ci.yml` del kit.
  - `go.mod` limpio: la única entrada nueva prevista es la de los tipos UUID si se acepta
    `github.com/google/uuid` (ver "Preguntas abiertas").

## D-A4 · Router: `net/http` en F1 tras la interfaz `Registrar`; chi como opción de F2

- **Contexto.** F1 expone un endpoint; F2 agrupará rutas de panel con permisos por módulo. La
  constitución (§II) exige minimizar dependencias y justificarlas.
- **Decisión.** **`net/http`** estándar (patrones `GET /ruta`, comodines `{id}` de Go 1.22+)
  detrás de la interfaz `httpserver.Registrar` (ver [arquitectura.md](./arquitectura.md) §4), de
  modo que los dominios no conocen el router. **chi queda como opción de F2**: se decide en el
  plan de esa funcionalidad, cuando aparezcan los grupos de rutas con permisos por módulo que
  motivan su evaluación. Si se adopta, se implementa un adaptador de ~15 líneas y **ningún dominio
  cambia**.
- **Datos y motivos de no adoptar un framework de ruteo desde ya.**
  - Coste de dependencias medido el 2026-09-30 (entradas en `go.mod`, incluidas las de test):
    **Gin 15 directas / 32 entradas**, **Echo 3/7**, **Fiber 8/13**, **chi 0** (solo la stdlib).
  - Gin y Echo imponen su propio tipo de contexto (`gin.Context`, `echo.Context`) que **choca con
    la propagación de `context.Context`** exigida por nuestras convenciones
    (`.agents/skills/go-backend/SKILL.md`); Fiber no implementa `net/http` de la stdlib.
  - Go 1.22 llevó *method matching* y comodines a `http.ServeMux`, que era la razón principal
    para usar un router externo.
  - Mercado (JetBrains, Go Ecosystem 2025–2026): Gin ~48 %, gorilla/mux 17 % (en declive),
    Echo 16 %, chi 12 %, Fiber 11 %. La popularidad no decide por nosotros: decide el coste de
    dependencias y la compatibilidad con `context.Context` (§II).
- **Alternativas consideradas.** Adoptar chi desde F1 (pospuesta, no descartada); Gin/Echo/Fiber
  (rechazadas: dependencias pesadas y/o contexto propio incompatible); quedarse sin interfaz y
  acoplar los dominios a `net/http` (rechazada: haría caro el cambio futuro de router).
- **Consecuencias.**
  - F1: cero dependencias de ruteo y un único adaptador pequeño en `platform/httpserver`.
  - El día que se adopte chi (u otro), el diff se limita a `platform/httpserver/` y a la línea que
    construye el `Registrar` en `main.go`.
  - El criterio de decisión para F2 queda explícito: si los grupos de rutas con permisos por
    módulo se vuelven incómodos con `net/http`, se adopta chi con justificación en ese `plan.md`.

## D-A5 · Versión del lenguaje: Go 1.27

- **Contexto.** El plan vigente de F1 fijaba Go 1.23+. Go 1.23 alcanzó su **fin de vida el
  2025-08-12**: hoy solo las ramas **1.26 y 1.27** reciben parches de seguridad. La skill
  `go-backend` pide "Go 1.23 o superior", es decir, un mínimo, no un techo.
- **Decisión.** **Go 1.27** en todo el proyecto: imagen `golang:1.27` en `backend/Dockerfile` y
  `go 1.27` en `backend/go.mod`. El CI del kit lee `backend/go.mod`, así que la versión la decide
  el proyecto, no el kit.
- **Alternativas consideradas.**
  - *Go 1.23 (lo que decía el plan)*: rechazado — sin soporte de seguridad desde 2025-08-12.
  - *Go 1.26*: viable, pero 1.27 es la rama con ventana de soporte más larga y no hay coste en
    adoptarla (sin dependencias sensibles a la versión en F1).
  - *Flotar con `golang:latest`*: rechazado — builds no reproducibles (SC-002 del plan de F1).
- **Consecuencias.**
  - Imágenes reproducibles y en soporte de seguridad.
  - Podemos usar sin reservas los patrones `net/http` de Go 1.22+ en los que se apoya D-A4.
  - Al actualizar Go en el futuro se tocan dos puntos (`go.mod` y el `FROM` del Dockerfile) y se
    re-ejecuta `make ci`; la política "versión siempre en soporte" queda como criterio.

## D-A6 · Config por entorno (stdlib), logs `slog` JSON, errores stdlib + `apperr`, DI manual

- **Contexto.** La constitución (§V, §VII) exige config por variables de entorno, logs
  estructurados, errores con contexto y calidad de Go estándar. La dependencia externa es un coste
  que hay que minimizar (§II).
- **Decisión.**
  - **Config**: `os.Getenv` + validación al arrancar en `internal/platform/config` (sin Viper:
    sus 17 entradas en `go.mod` no se justifican para leer variables). Valores por defecto de
    desarrollo para que `make up` funcione en un clon limpio.
  - **Logs**: `log/slog` con handler **JSON**, nivel desde `LOG_LEVEL`, logger por petición con
    `request_id`.
  - **Errores**: la stdlib (`errors.Is/As`, `fmt.Errorf("...: %w", err)`) más los errores de
    dominio tipados de `internal/platform/apperr`, traducidos a HTTP solo en
    `httpserver.WriteError`.
  - **Inyección de dependencias**: composición **manual** desde `cmd/api/main.go` (constructores
    por capa: repository → service → handler), sin `wire` ni `fx`. Sin estado global ni `init()`
    con lógica.
- **Alternativas consideradas.** Viper o `envconfig` para config (rechazadas: 17 entradas / una
  dependencia para lo que son 20 líneas de `os.Getenv` + validación); `zerolog`/`zap` para logs
  (rechazadas: `slog` está en la stdlib desde Go 1.21 y es la que pide la constitución); librerías
  de errores ricos tipo `pkg/errors` (rechazadas: `%w` basta); `wire`/`fx` para DI (rechazadas:
  generación de código y runtime para un grafo de ~4 nodos que se lee entero en `main.go`).
- **Consecuencias.**
  - `go.mod` casi vacío: la única dependencia de runtime de F1 es `pgx` (D-A8).
  - `main.go` es el mapa de dependencias del sistema: se lee entero y se audita de un vistazo.
  - Configurar mal el entorno falla **al arrancar** con mensaje claro, no en caliente.

## D-A7 · Sesiones de F2 *(confirmada por el humano el 2026-10-04 — sesión en Redis)*

- **Contexto.** F2 (acceso y gestión de usuarios) necesita sesión para el panel, y la skill
  `react-frontend` prohíbe tokens en `localStorage` (constitución §IV: contraseñas con
  bcrypt/argon2; §IV: los endpoints protegidos deben verificar authn/authz en el servidor).
- **Decisión (confirmada por el humano el 2026-10-04).**
  - **Sesión en servidor, en Redis** (cambio respecto de la propuesta original: el humano eligió
    explícitamente Redis, con el objetivo de adoptar/probar la tecnología). La cookie `httpOnly`
    (y `Secure`/`SameSite` en entornos con TLS) solo lleva el identificador de sesión (un token
    opaco del que en Redis se guarda su **SHA-256**, nunca el token en claro); revocar es borrar la
    clave, con lo que **desactivar un usuario surte efecto inmediato** (imposible con un JWT
    autocontenido). Tiempos confirmados: **vida absoluta de 1 hora + inactividad de 30 minutos**.
  - **Hash de contraseñas con bcrypt** (constitución §IV; CWE-256). Nunca en texto
    plano ni reversible.
  - **Permisos por módulo en tablas propias** (módulos: `contactos`, `eventos`, `ministerios`…),
    agrupados en roles creados por el administrador (decisión 5 de
    `docs/producto/roadmap.md`). La autorización se verifica **en el servidor** por módulo
    (`middleware.AuthzByModule`).
- **Alternativas consideradas.**
  - *Tabla `sessions` en PostgreSQL*: era la recomendación de la propuesta; **descartada por la
    decisión explícita del humano del 2026-10-04**. Queda documentada como plan B (cero servicios
    nuevos, transaccional con `users`, durabilidad) en
    `specs/002-acceso-gestion-usuarios/research.md` R1.
  - *JWT autocontenido en cookie*: rechazado por defecto — la revocación al desactivar un usuario
    es inmediata con sesión en servidor y forzada con listas negras en JWT.
  - *Token en `localStorage`*: rechazado — expuesto a XSS (CWE-79) y prohibido por la skill.
  - *Framework de auth completo (p. ej. OIDC/Keycloak)*: rechazado — peso desproporcionado para
    un panel con pocos usuarios internos.
- **Consecuencias.**
  - Servicio **`redis`** en `docker-compose.yml`, variable `REDIS_URL` y dependencia de runtime
    `github.com/redis/go-redis/v9` (justificada bajo D-A8 por esta decisión humana).
  - **Sin** tabla `sessions` en PostgreSQL: las tablas de F2 son `users`, `roles`, `permissions` y
    `role_permissions`; los intentos de acceso (bloqueo de 5 intentos / 15 min) también viven en
    Redis con TTL.
  - Un middleware de `authn` en `platform/middleware` y un `session.Store` en `platform/session`.
  - Las pruebas de integración necesitan Redis y lo levantan con `testcontainers-go` dentro del
    propio test: `.github/workflows/ci.yml` es del kit y **no se puede editar**.
  - CSRF pasa a ser obligatorio en rutas con sesión (por eso está en la cadena de grupos de
    [arquitectura.md](./arquitectura.md) §6).
  - **Estado: aceptada.** Confirmada por el humano el 2026-10-04; el detalle de implementación
    (claves, TTL y tiempos) está en `specs/002-acceso-gestion-usuarios/plan.md` (P1/P9) y en su
    `data-model.md`.

## D-A8 · Dependencias mínimas

- **Contexto.** Constitución §II: "las dependencias externas DEBERÍAN minimizarse; toda
  dependencia nueva requiere justificación en `plan.md`".
- **Decisión.** En F1 **solo `pgx/v5` entra al binario**. El CLI de `golang-migrate` se usa en
  desarrollo y CI (target `make db-migrate`, paso "Migraciones"), **no se enlaza**; lo mismo
  aplica a `sqlc` y a `openapi-typescript` (herramientas de desarrollo, sus artefactos generados
  se commitean). Toda dependencia nueva — runtime o de desarrollo con impacto en el build — se
  justifica en el `plan.md` de la funcionalidad que la introduce, con su cuenta de entradas en
  `go.mod`/`package.json` y las alternativas descartadas. Las dependencias deben pasar
  `govulncheck` y `npm audit` sin vulnerabilidades altas o críticas (constitución §IV).
- **Alternativas consideradas.** Instalar el toolchain completo desde F1 "por si acaso"
  (rechazada: configuración muerta y dependencias sin uso, como ya se decidió para sqlc en
  `specs/001-estructura-base/research.md` R4).
- **Consecuencias.**
  - `go.mod` de F1 con una sola dependencia directa; el Dockerfile de producción solo enlaza lo
    que se usa.
  - El coste de cada dependencia se ve **antes** de escribir código, en la fase de plan.
  - Ejemplos conocidos que pasarán por este filtro en F2: tipos UUID (`google/uuid` vs.
    `pgtype.UUID`, ver "Preguntas abiertas"), la librería de validación de DTOs y el adaptador
    chi si se adopta.

## D-A9 · Diferido a propósito (con su punto de decisión)

| Qué se difiere | Por qué ahora no | Dónde se decide |
|---|---|---|
| **Métricas y trazas** (métricas estilo Prometheus, OpenTelemetry) | No hay entorno desplegado ni operación que las consuma; añadirlas ahora sería configuración muerta | La funcionalidad que traiga el **primer despliegue operado** (probablemente junto al trabajo de `devops` para staging/producción) |
| **Colas de trabajo** (procesamiento asíncrono: correos, transcodificación de medios) | Todo F1–F7 es síncrono y de bajo volumen; no hay tareas en segundo plano reales | La primera funcionalidad que **necesite trabajo diferido** (candidata natural: F8/F9 si hay notificaciones o procesamiento de medios) |
| **Almacenamiento de imágenes** (objeto/S3, subida de archivos) | F1–F7 no suben archivos; las imágenes de F8 llegarán con su spec | **F8** (noticias y galería), o antes si el cliente lo pide |
| **i18n** (español/inglés) | El sitio bilingüe es requisito del roadmap (decisiones 6 y 8), pero F1–F2 son el panel y la infraestructura: la primera UI pública bilingüe es F3 | **F3** (portada e información general), que fija el mecanismo (catálogo de textos, selector de idioma, contenido en 1 o 2 idiomas) |
| **Generación de handlers desde el contrato OpenAPI** (codegen servidor) | Hoy el contrato es fuente de verdad y los handlers se escriben a mano con validación explícita; el generador añadiría una herramienta y rígidez sin dolor actual | Cuando el dolor de mantener handlers a mano se mida de verdad (revisión posterior al MVP); no es una promesa |

- **Alternativas consideradas.** Adoptar todo desde F1 (rechazado: viola D-A8 y §II —
  dependencias sin uso) y aplazar sin decidir dónde se decide (rechazado: este registro fija el
  punto de decisión para no olvidarlo).
- **Consecuencias.** El roadmap tiene anclados los momentos de revisión; nada de esto se implementa
  "de paso" en otra funcionalidad sin actualizar este registro y justificarlo en su `plan.md`.

---

## Preguntas abiertas / pendientes

1. **¿Cómo encaja la ampliación en el proceso?** Las decisiones D-A3…D-A9 amplían lo previsto en
   F1 (sqlc fijado, `Registrar`, Go 1.27, la plataforma interna). Falta decidir si esto se
   incorpora **ampliando F1** (la spec de F1 se está actualizando en paralelo) o como una
   **funcionalidad nueva tipo F1.5** con su propia spec. Decisión del orquestador/humano; estos
   documentos no dependen de ella (describen el **cómo**, no el alcance de F1).
2. **Confirmación de D-A7 (sesiones de F2) — RESUELTA el 2026-10-04.** El humano confirmó la
   propuesta (cookie `httpOnly` + sesión en servidor, hash bcrypt, permisos por módulo en tablas
   propias) y decidió el detalle asociado: **la sesión vive en Redis** (no en una tabla de
   PostgreSQL), con vida absoluta de 1 hora e inactividad de 30 minutos. Ver D-A7 y
   `specs/002-acceso-gestion-usuarios/plan.md` (P1/P9) y su `research.md` (R1/R15).
3. **Trigger y decisión de chi en F2.** D-A4 deja chi como opción. El plan de F2 debe cerrarla
   explícitamente: seguir con `net/http` o adoptar chi con adaptador, según el peso que tengan los
   grupos de rutas con permisos por módulo.
4. **Mecanismo de `platform/validate` (F2).** Las etiquetas de los DTOs están decididas; falta
   elegir la implementación: librería (`go-playground/validator`, con su entrada en `go.mod` bajo
   D-A8) o validador propio mínimo (sin dependencia). Decisión del `plan.md` de F2.
5. **Tipos UUID en Go (F2).** `github.com/google/uuid` (una dependencia nueva justificable) vs.
   `pgtype.UUID` de `pgx` (cero dependencias nuevas, tipos más torpes). Decisión del `plan.md` de
   F2; ver §5.3 de [arquitectura.md](./arquitectura.md).
6. **Path definitivo del módulo Go.** Los ejemplos usan `simiente-santa/backend`; se fija cuando
   exista el remoto (acción humana pendiente ya registrada en
   `specs/001-estructura-base/plan.md`, riesgo R1).

Todo lo que aparezca nuevo se agrega aquí antes de suponer nada.
