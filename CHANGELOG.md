# Changelog

Todos los cambios notables de este proyecto se documentan en este archivo.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es/1.1.0/) y el proyecto usa [Versionado Semántico](https://semver.org/lang/es/).

## [0.3.0] — 2026-10-10

Tercera entrega: **F3 — Portada e información general** (rama `003-portada-info-general`). Spec aprobada: [`specs/003-portada-info-general/spec.md`](specs/003-portada-info-general/spec.md). Verificación de cierre en verde ([`cierre.md`](specs/003-portada-info-general/cierre.md): `make ci` EXIT=0, e2e 7/7, cobertura 84,7 % en `internal/portada/service*.go`, sin deriva de generados). Validación del 2026-10-10 en dos ciclos: el primero rechazó un bloqueante (la semántica `null` de los `PATCH`) y el segundo quedó **sin hallazgos bloqueantes** — [QA](specs/003-portada-info-general/revision-2026-10-10-qa-ciclo2.md), [revisión de código](specs/003-portada-info-general/revision-2026-10-10-codigo-ciclo2.md) y [seguridad](specs/003-portada-info-general/revision-2026-10-10-seguridad-ciclo2.md) aprobados, con deuda menor aceptada (ver [`estado.md`](specs/003-portada-info-general/estado.md)).

### Agregado

- **Portada pública bilingüe en `/`** (FR-001…FR-007): la página de inicio del sitio es ahora la portada de la iglesia, visible **sin cuenta ni registro**. Reúne la identidad (nombre oficial, lema/misión/visión, logotipo e imagen de portada), «quiénes somos» (texto plano, ≤ 1.000 caracteres), el horario de servicios (día, hora, nombre y lugar por servicio), los canales de WhatsApp (mensaje directo y/o enlace de grupo, varios canales), las redes sociales (catálogo fijo Facebook, Instagram, YouTube, TikTok y Spotify, un enlace por red) y el contacto (dirección, correo y teléfono). El diseño respeta el Manual de Identidad del cliente (`resources/MANUAL DE MARCA.pdf`), incluidas las reglas de uso del logotipo; la portada se adapta a 320 px en adelante, navega con teclado y lector de pantalla, y todo contenido se trata como dato —nunca como código—.
- **Idiomas español e inglés** (Decisión 6 y 8): selector visible en la cabecera; el cambio aplica de inmediato y la elección **se recuerda en el dispositivo** (`localStorage`; la primera visita sin preferencia entra en español). Cada contenido puede llevar su versión en inglés, que es **opcional**: lo sin traducir se muestra íntegramente en español (nunca un campo vacío ni traducción automática, FR-009).
- **Control de publicación por elemento** (Decisión 4): cada pieza de la portada tiene estado **borrador/publicado** y se publica o retira por separado, incluidos los elementos únicos (identidad, «quiénes somos» y contacto); solo lo publicado llega al visitante y **una sección sin elementos publicados se oculta por completo**, sin secciones vacías (SC-012). Editar algo ya publicado y guardarlo lo hace visible de inmediato (FR-014).
- **Panel `/panel/informacion`** con pestañas **Identidad · Quiénes somos · Horario de servicios · Canales de WhatsApp · Redes sociales · Contacto**, cada elemento con su píldora de estado y sus acciones de publicar/retirar. Todo bajo el permiso **«Portada e información general»** (`portada`), ya reservado por F2 (FR-015 de F2) y ahora activo: verificado en el servidor en cada operación, con la entrada del menú y la ruta protegidas (`RequirePermission`) y la denegación clara «No tienes permiso para acceder a este módulo» (FR-012).
- **Subida de imágenes** desde el panel (logotipo e imagen de portada): JPEG/PNG/WebP **por firma binaria** (sin SVG/GIF), ≤ 8 MB (`UPLOAD_MAX_BYTES`), nombre generado por el servidor (`img_<uuid>`), texto alternativo obligatorio y descarga `GET /api/v1/media/{fileName}` **solo** para archivos referenciados por contenido publicado, con `nosniff`, `inline` y `no-store`; retirar la identidad deja de servir su imagen sin borrarla (FR-019, `analyze` C4/M6).
- **Auditoría ampliada al módulo** (FR-017): toda edición, alta, publicación, retirada, subida de imagen, rechazo y denegación queda registrada en el registro de solo lectura de F2 con **15 nuevos códigos `home.*`** y etiquetas «Portada · `<sección>` · `<elemento>`»; el registro es **atómico** con la mutación (si falla el registro, la edición no se aplica).
- **API y base de datos**: contrato vivo a **0.4.0** ([`backend/api/openapi.yaml`](backend/api/openapi.yaml)) con `GET /api/v1/portada`, `GET /api/v1/media/{fileName}` y el subgrupo `/api/v1/admin/portada/*`; migraciones `000005` (tablas `home_*`, singletons con `upsert` y catálogo de redes) y `000006` (ampliación de `admin_actions`), ambas con `up`/`down`; tipos TypeScript regenerados con `make api-gen`.
- **Entorno**: volumen Docker `uploads_data` (como `pgdata`, sobrevive a los rebuilds) y variables nuevas documentadas en `.env.example`: `UPLOAD_DIR` (carpeta de imágenes; `/var/lib/simiente/uploads` en el contenedor) y `UPLOAD_MAX_BYTES` (8 MB).
- **Pruebas**: backend unitarias e integración contra PostgreSQL real (incluida la migración `000006` up→down→up y la atomicidad mutación+auditoría; cobertura 84,7 % en `internal/portada/service*.go`), frontend Vitest + Testing Library + MSW (más de 270 pruebas) y **e2e Playwright 7/7** (`acceso`, `auditoria`, `portada-accesibilidad`, `portada-panel`, `portada-patch-null`, `portada-publica`, `status`), recorriendo publicación, borradores, idioma es/en con fallback, responsividad y denegación con auditoría.

### Cambiado

- **`/` deja de ser la página de estado**: la portada pública pasa a ser la página de inicio (FR-001, decisión del humano del 2026-10-09) y la pantalla «Estado del sistema» de F1 se traslada a **`/health`** (ruta de la SPA, distinta del endpoint `GET /healthz`, que **no cambia**); el e2e de estado de F1 se actualizó a la nueva ruta.
- **El permiso `portada` de F2 deja de estar reservado**: «Portada e información general» se activa para los roles que lo necesiten; el sello «Disponible más adelante» desaparece de ese módulo del panel.
- **Auditoría de F2 ampliada**: el enum de acciones de `AdminActionItem` incorpora los códigos `home.*` y el `targetKind` `content`, sin romper los registros existentes.

### Corregido

- **Permisos de escritura del volumen de subidas**: el backend (usuario no-root) no podía escribir en `uploads_data` y toda subida de imagen respondía `500` (fix `a93a233`; defecto hallado al correr los e2e).
- **Selector de idioma**: cabecera y pie compartían el mismo `name` y los radios no se marcaban entre sí; ahora cada selector usa `useId()`.
- **Navegación a 320 px**: la `nav` de la portada provocaba desplazamiento horizontal en pantallas estrechas (`min-w-0`), incumpliendo la promesa de responsividad.
- **`PATCH` con `null` no vaciaba los campos opcionales** (BLOQUEANTE del primer ciclo de validación): los `PATCH` del panel ignoraban `null`; ahora se distingue «vaciar el campo» de «no enviado» con el tipo `Optional[T]` (fix `d5d4610`, con e2e de regresión `portada-patch-null`).
- **Formato de la hora del horario y área de toque del selector**: el horario se mostraba en formato 24 h (se muestra en a.m./p.m. localizado, con «m.» solo en el mediodía exacto: `fc96185`, `db83a61`) y el área táctil del selector de idioma pasó de 20 px a 44 px (`2b72138`).

### No entra en esta versión

- **Sin las demás secciones del sitio público**: eventos, actividades (F4), grupos (F5), ministerios (F6), donaciones (F7), noticias y galería (F8) y medios (F9).
- **Sin formularios** (contacto, inscripción) ni mapa incrustado: la comunicación es por WhatsApp, redes y los datos publicados; la dirección se muestra como texto.
- **Sin traducción automática**: el equipo ingresa el inglés a mano, contenido por contenido.
- **Sin historial de versiones, publicación programada ni verificación de la vigencia de los enlaces** (borrar es borrado físico; para ocultar sin perder se retira el elemento).
- **Pendientes de validación con personas** (no de código): la prueba de usabilidad SC-010 (≥ 6 personas, protocolo del `analyze` M7) y la parte manual de accesibilidad SC-008 (checklist WCAG 2.1 AA) se cierran en la fase de validación.
- **Sin despliegue a producción**: la entrega se completa con el merge humano del PR de la rama `003-portada-info-general`.

## [0.2.0] — 2026-10-05

Segunda entrega: **F2 — Acceso y gestión de usuarios** (rama `002-acceso-gestion-usuarios`). Spec aprobada: [`specs/002-acceso-gestion-usuarios/spec.md`](specs/002-acceso-gestion-usuarios/spec.md). Validación del 2026-10-05 **sin hallazgos bloqueantes**: [QA](specs/002-acceso-gestion-usuarios/revision-2026-10-05-qa.md) (aprobado), [revisión de código](specs/002-acceso-gestion-usuarios/revision-2026-10-05-codigo.md) y [seguridad](specs/002-acceso-gestion-usuarios/revision-2026-10-05-seguridad.md) (aprobados con observaciones).

### Agregado

- **Panel de administración con inicio de sesión**: rutas `/login`, `/cambiar-contrasena`, `/panel`, `/panel/usuarios`, `/panel/roles`, `/panel/auditoria` y `/sin-permiso`. Sin sesión, cualquier sección redirige a la pantalla de acceso (FR-001); el fallo de acceso responde **siempre el mismo mensaje genérico** y nunca revela si la cuenta existe (FR-003); tras el **5.º intento fallido** el acceso queda bloqueado **15 minutos** (FR-006); cerrar sesión en cualquier momento termina la sesión de verdad (FR-004).
- **Sesión en Redis**: **30 minutos de inactividad** y **1 hora máxima** desde el inicio, aunque haya actividad continua (FR-005); cookie `ss_session` `HttpOnly`/`SameSite=Lax`, CSRF *double-submit* firmado y revocación inmediata de todas las sesiones de una cuenta al desactivarla o restablecerle la contraseña (FR-012). Redis corre **sin persistencia** por decisión humana (2026-10-04): reiniciarlo obliga a volver a entrar y pone a cero los contadores de intentos, pero **no toca la auditoría**, que vive en PostgreSQL.
- **Inicialización única de la puesta en marcha**: `POST /api/v1/setup/initialize` con la cabecera `X-Setup-Token` (`BOOTSTRAP_TOKEN`) crea el rol «Administrador» con los 9 permisos; solo funciona con el sistema sin cuentas y **no puede repetirse** (FR-007, US2). Ninguna operación puede dejar el panel sin al menos un administrador activo (FR-008).
- **Gestión de usuarios**: crear, editar, activar/desactivar y restablecer la contraseña de las cuentas (FR-009, FR-011). Nombre, apellidos, correo y teléfono obligatorios y validados (teléfono con ≥ 7 dígitos); correos y nombres de rol comparados **normalizados**, de modo que uno que solo difiere en mayúsculas o espacios se trata como duplicado (Q5). Política de contraseñas: **8 a 64 caracteres**, combinando mayúsculas, minúsculas, números y caracteres especiales, y distintas del nombre, los apellidos y el correo (FR-010); quien recibe una contraseña de un administrador **debe cambiarla al entrar**. Las cuentas **nunca se eliminan**: solo se desactivan y sus datos se conservan siempre (FR-012/FR-013).
- **Roles con permisos por módulo** y **sin catálogo fijo de roles** (Decisión 5 del roadmap): creación, edición de nombre y permisos, y eliminación **solo** cuando el rol no tiene cuentas asignadas; un rol siempre conserva al menos un permiso (FR-014, FR-017). Catálogo de 9 permisos sembrado por la migración (`portada`, `eventos`, `actividades`, `grupos`, `ministerios`, `donaciones`, `noticias`, `medios` y `admin_usuarios_roles`), verificado **en el servidor** en cada operación (FR-015/FR-016); un cambio de permisos se refleja de inmediato en el acceso de las cuentas del rol (FR-018). Los permisos de los módulos que aún no existen quedan **reservados** para F3–F9.
- **Auditoría de solo lectura** (cambio de alcance aprobado por el humano el 2026-10-04): **último acceso** en la ficha y en el listado de cada cuenta, sin fechas inventadas para las que nunca entraron (FR-021); **historial de accesos** con fecha y hora, resultado e IP de origen, incluidos los intentos contra correos que no existen —registrados sin asociar a ninguna cuenta y sin guardar el correo probado— (FR-022); **historial de acciones administrativas** con quién la hizo, qué hizo, sobre qué, cuándo y con qué resultado, también cuando falla o se deniega (FR-023). La sección `/panel/auditoria` filtra por cuenta y rango de fechas y pagina los resultados (FR-024); los registros **no se pueden editar ni borrar** por ninguna vía y **nunca contienen contraseñas** (FR-025/FR-026).
- **Base de datos**: migraciones `000002`–`000004` (`permissions`/`roles`/`role_permissions`, `users`, `login_events`/`admin_actions`), sin huecos y con `up`/`down`.
- **Contrato vivo a 0.3.0** ([`backend/api/openapi.yaml`](backend/api/openapi.yaml)): 18 operaciones nuevas (sesión, inicialización, administración de usuarios, roles y permisos, y auditoría) con el esquema de seguridad `sessionCookie`; los tipos TypeScript del frontend se regeneran con `make api-gen`.
- **Entorno y operación**: `make up` levanta ahora **`db` + `redis` + `backend` + `frontend`**; servicio nuevo `redis` (`redis:7-alpine`, sin volumen ni persistencia) en `docker-compose.yml`. Variables nuevas documentadas en `.env.example`: `REDIS_URL`, `SESSION_COOKIE_SECURE`, `SESSION_IDLE_TTL_MINUTES` (30), `SESSION_ABSOLUTE_TTL_MINUTES` (60) y `BOOTSTRAP_TOKEN`; `SESSION_SECRET` deja de estar solo documentada y el backend la usa para firmar la cookie CSRF.
- **Pruebas**: unitarias e integración del backend contra **PostgreSQL y Redis reales** (`testcontainers-go`; cobertura 87,5 % en el dominio `usuarios`), frontend con Vitest + Testing Library + MSW (34 archivos / 176 pruebas) y e2e con Playwright (`acceso`, `auditoria`, `status`: 3/3 en verde).

### Cambiado

- **`SESSION_SECRET` deja de ser opcional en producción**: con `APP_ENV=production` el arranque falla si está vacía o por debajo de 32 caracteres, y también si `SESSION_COOKIE_SECURE` sigue en `false` (en desarrollo se mantiene un valor por defecto claramente de desarrollo).
- **`docker-compose.yml`** añade el servicio `redis` (puerto `REDIS_PORT`, 6379) y las nuevas variables del backend; el backend ya no arranca sin `REDIS_URL`.

### Corregido

- **Rutas con parámetros**: `platform/httpserver` no rellenaba `PathValue`, por lo que las operaciones sobre `/api/v1/admin/usuarios/{id}` y `/api/v1/admin/roles/{id}` no veían el id (bug heredado de F1).
- **`Retry-After` en `429`**: `httpserver.WriteError` no emitía la cabecera que el contrato documenta para las respuestas limitadas (bug heredado de F1).
- **CI en rojo por la stdlib de Go**: `backend/go.mod` declaraba `go 1.27` sin parche y el CI —que instala Go con `go-version-file: backend/go.mod`— resolvía go1.27.1, con 9 vulnerabilidades de la stdlib reportadas por `govulncheck` y corregidas en go1.27.2 (`GO-2026-6603, -6605, -6607, -6608, -6610, -6611, -6612, -6613 y -6617`: `net/http`, HTTP/2, `net/textproto`, `crypto/tls`). Se fijan `go 1.27.2` en `go.mod` y `golang:1.27.2` en el `Dockerfile`, de modo que tanto el CI como el binario de producción quedan con la stdlib parcheada; `govulncheck` v1.8.0 sale en verde.

### No entra en esta versión

- **Sin recuperación de contraseña por auto-servicio con correo**: en el MVP la restablece un administrador desde el panel y la persona debe cambiarla al entrar (decisión Q2; queda como idea futura en el [roadmap](docs/producto/roadmap.md)).
- **Sin eliminación de cuentas ni registro público de usuarios**: solo un administrador crea cuentas, y retiras el acceso desactivándolas.
- **Auditoría limitada al ámbito de F2** (accesos al panel y gestión de cuentas y roles), sin exportación, alertas ni purga; la de los módulos de contenido llegará con F3–F9.
- **Sin despliegue a producción**: la entrega se completa con el merge humano del PR de la rama `002-acceso-gestion-usuarios`.

## [0.1.0] — 2026-10-03

Primera entrega: **F1 — Estructura base** (rama `001-estructura-base`, PR #1).

### Agregado

- **Backend en Go 1.27** (`backend/`) con el endpoint `GET /healthz`, que informa en cada consulta el estado **real** de la conexión a PostgreSQL: `200` con `{"status":"ok","database":"connected"}` cuando todo está vivo y `503` con sobre de error `database_unavailable` cuando la base no responde (respuesta en ≤ 2 s; FR-004).
- **Formato de respuesta uniforme** para éxito y error en toda la API (SC-008), sin exponer información interna del sistema ante errores inesperados (SC-009).
- **Plataforma interna del backend**: configuración por variables de entorno, logger, manejo de errores, conexión a PostgreSQL, servidor HTTP y middleware (CORS y logging por petición).
- **Frontend en React 19 + TypeScript + Vite** (`frontend/`): página de estado que consulta `/healthz` y muestra de forma comprensible si el backend y la base de datos están vivos. Los tipos TypeScript se generan desde el contrato de la API (`make api-gen`).
- **PostgreSQL 16** con migraciones versionadas (`golang-migrate`): baseline `000001` (no-op). En F1 **no** hay tablas de negocio.
- **Docker Compose**: `make up` levanta base de datos, backend y frontend con un único comando y Docker como único prerequisito (FR-001); `make down` detiene el entorno conservando los datos.
- **CI en GitHub Actions con 6 jobs**: kit y configuración de agentes, migraciones y constitución, detección de proyectos, backend (Go), frontend (React) y búsqueda de secretos. En verde (6/6) en el PR #1.
- **Contrato vivo** [`backend/api/openapi.yaml`](backend/api/openapi.yaml) (OpenAPI 3.1): única fuente de la API y origen de los tipos del frontend.
- **Toolchain fijada**: Go 1.27, Node.js 24 (LTS) y PostgreSQL 16; además `golang-migrate` v4.20.1, `sqlc` v1.31.1, `golangci-lint` v2.14.0 y `govulncheck` v1.8.0. `make doctor` verifica el entorno y todas las versiones en uso tienen soporte de seguridad vigente (SC-010, comprobado a 2026-10-03).
- **Receta para agregar un área de negocio nueva** (`docs/tecnico/arquitectura.md` §8): verificada antes de cerrar F1 (SC-007) con tres corridas de un agente fresco; la tercera añadió un área de práctica (`GET /api/v1/muestra`) sin tomar decisiones de arquitectura y sin modificar las áreas existentes, con `make ci` en verde. Evidencias en `specs/001-estructura-base/checklists/receta.md`.
- **Comandos `make`** para el flujo diario (`up`, `down`, `test`, `lint`, `security`, `ci`, `db-migrate`, `e2e`, `api-gen`, `sqlc-gen`, `doctor`, …) y hooks de git con búsqueda de secretos (`gitleaks`).
- **Pruebas automáticas**: backend (unitarias y de integración con `-tags=integration`), frontend (Vitest) y end-to-end (Playwright).

### No entra en esta versión

- **Sin tablas de negocio ni endpoints además de `/healthz`**: F1 monta el terreno (Go + React + PostgreSQL + Docker + CI) pero no añade contenido visible al sitio.
- **Sin autenticación ni sesiones**: llegan con **F2** (acceso y gestión de usuarios), que introducirá también el uso de `SESSION_SECRET` (hoy documentada en `.env.example` pero sin consumo en F1).
- **Sin despliegue a producción**: la entrega se completa con el merge humano del PR #1.
