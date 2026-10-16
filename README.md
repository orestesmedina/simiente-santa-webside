# Simiente Santa

Sitio web de la Iglesia Simiente Santa: sitio público (desde F3 la portada con la información general de la iglesia; desde F10 con el sistema visual y de movimiento moderno; desde F4 la sección de eventos y actividades y el bloque «Lo que viene» de la portada; desde F6 la sección de ministerios y su bloque en la portada; desde F7 la sección de ofrendas y diezmos y su bloque en la portada; grupos, noticias y medios por llegar) y panel de administración. El detalle de lo que se construye, fase por fase, está en [`docs/producto/roadmap.md`](docs/producto/roadmap.md).

## Qué es la F1 (Estructura base)

La funcionalidad **F1** es el esqueleto técnico sobre el que se construyen F2–F9:

- **Backend en Go** con `GET /healthz`, que informa en cada consulta el estado **real** de la conexión a PostgreSQL.
- **Frontend en React + TypeScript** con una página inicial que muestra ese estado de forma comprensible — desde F3 vive en `/health`, porque en `/` ahora está la portada pública.
- **PostgreSQL** con migraciones versionadas (`golang-migrate`) — en F1 sin tablas de negocio: solo la baseline `000001` (no-op).
- **Docker Compose** para levantar todo con un solo comando y **CI** (GitHub Actions) que valida cada cambio.
- Contrato vivo en [`backend/api/openapi.yaml`](backend/api/openapi.yaml), del que se generan los tipos TypeScript del frontend.

Rama: `001-estructura-base`. Spec, plan y tareas en [`specs/001-estructura-base/`](specs/001-estructura-base/).

## Qué es F2 (Acceso y gestión de usuarios)

La funcionalidad **F2** abre el **panel de administración** que usarán F3–F9:

- **Inicio de sesión** en <http://localhost:5173/login> con correo y contraseña. Sin sesión, cualquier sección del panel redirige a la pantalla de acceso. La sesión dura como máximo **1 hora** desde el inicio y se cierra a los **30 minutos de inactividad**.
- **Puesta en marcha con una inicialización única**: `POST /api/v1/setup/initialize` con la cabecera `X-Setup-Token` (`BOOTSTRAP_TOKEN`) crea el primer administrador; no se puede repetir (ver «Puesta en marcha del panel» más abajo).
- **Gestión de usuarios**: crear, editar, activar/desactivar y restablecer la contraseña. Quien recibe una contraseña de un administrador debe cambiarla al entrar. Las cuentas **no se eliminan**: solo se desactivan.
- **Roles con permisos por módulo** (sin catálogo fijo de roles), verificados en el servidor en cada operación; los cambios de un rol se reflejan de inmediato en sus cuentas.
- **Auditoría de solo lectura** en `/panel/auditoria`: último acceso de cada cuenta, historial de inicios de sesión (exitosos y fallidos, con IP de origen) e historial de acciones administrativas, con filtros por cuenta y rango de fechas.
- **Redis** (servicio nuevo en compose) guarda la sesión y los contadores de intentos. Corre **sin persistencia a propósito**: un reinicio obliga a volver a entrar y pone a cero los contadores, pero no toca la auditoría, que vive en PostgreSQL.

Rama: `002-acceso-gestion-usuarios`. Spec, plan y quickstart en [`specs/002-acceso-gestion-usuarios/`](specs/002-acceso-gestion-usuarios/).

## Qué es F3 (Portada e información general)

La funcionalidad **F3** es la primera cara pública del sitio: la portada que cualquier persona ve sin cuenta, y el módulo del panel con el que el equipo la mantiene al día.

- **Portada pública en <http://localhost:5173/>**: identidad de la iglesia (nombre oficial, lema/misión/visión, logotipo e imagen de portada), «quiénes somos», horario de servicios, canales de WhatsApp, redes sociales y contacto (dirección, correo y teléfono). Sin sesión ni registro; solo se ve lo que está **publicado**, y una sección sin ningún elemento publicado **se oculta por completo**.
- **Bilingüe (español/inglés)** con selector visible: el inglés es opcional contenido por contenido (lo sin traducir se muestra en español, nunca vacío) y el sitio recuerda la elección en el dispositivo (`localStorage`; la primera visita entra en español).
- **Panel `/panel/informacion`** con las pestañas Identidad, Quiénes somos, Horario de servicios, Canales de WhatsApp, Redes sociales y Contacto; cada elemento tiene estado **Borrador/Publicado** y se publica o retira por separado (lo que ya está publicado es visible al guardarlo). Requiere el permiso **«Portada e información general»** de F2, y toda edición queda registrada en la auditoría de `/panel/auditoria`.
- **Subida de imágenes** (logotipo e imagen de portada): JPEG/PNG/WebP hasta **8 MB**, con texto alternativo obligatorio; se guardan en el volumen `uploads_data` (`UPLOAD_DIR`/`UPLOAD_MAX_BYTES`) y solo se sirven cuando el contenido que las referencia está publicado.
- La página «Estado del sistema» de F1 se traslada a **`/health`** (ruta de la SPA; el endpoint `GET /healthz` no cambia): en `/` empieza la portada.

Rama: `003-portada-info-general`. Spec, plan, UX y quickstart en [`specs/003-portada-info-general/`](specs/003-portada-info-general/).

## Qué es F10 (Sistema de diseño y movimiento)

La funcionalidad **F10** es el rediseño moderno del sitio público (propuesta visual **B «Marea Turquesa»**, elegida por el cliente), aplicado a la portada de F3 y dejado como base para las secciones F4–F9:

- **Sistema visual moderno** construido sobre el Manual de Identidad y los tokens de F3: tipografías de marca, proporción de color 70/20/10, componentes reutilizables y más aire entre secciones (tokens `--space-section*`).
- **Movimiento marcado/elegante con revelado al hacer scroll**: las secciones se revelan al entrar en pantalla y cada elemento interactivo responde en hover, foco y activación, **respetando por completo `prefers-reduced-motion`** (sin animaciones y con todo el contenido disponible).
- **Navegación por dispositivo**: barra fija visible en escritorio y **menú hamburguesa accesible en móvil** (≤ 860 px), más el **mapa del sitio en el pie** (grupos Explorar / Contacto / Síguenos).
- **Selector de idioma compacto** (mismos textos completos, menos espacio) y **modo oscuro del sitio público** con interruptor manual: por defecto sigue la preferencia del dispositivo y la elección se recuerda en el mismo dispositivo.
- **Sin cambios de comportamiento ni de datos**: el contenido, el bilingüismo, la publicación por elemento, los permisos, la auditoría y el panel de F2/F3 siguen intactos. El rediseño del panel queda para **F10B**.

Rama: `010-sistema-diseno-movimiento`. Spec, plan, UX y quickstart en [`specs/010-sistema-diseno-movimiento/`](specs/010-sistema-diseno-movimiento/).

## Qué es F4 (Eventos y actividades)

La funcionalidad **F4** es la primera sección de contenido dinámico del sitio: **lo que viene**, visible sin cuenta, y los tres módulos del panel con los que el equipo lo mantiene al día.

- **Sección pública en <http://localhost:5173/eventos>**: dos bloques separados, **Próximos eventos** (encuentros periódicos con fechas concretas y texto opcional de frecuencia) y **Próximas actividades** (acciones puntuales con fecha única y hora de inicio/fin opcional), con título, fecha, hora, lugar, detalle e imagen opcional, en **orden cronológico**. Entrada de navegación propia («Eventos y actividades») en la barra, el menú móvil y el mapa del pie, **oculta** mientras no haya nada publicado y próximo; sin contenido, la página muestra un aviso **sin redirigir**.
- **Bloque «Lo que viene» en la portada** (cambio de alcance Q8, **primera sección tras el hero y antes de «quiénes somos»**, `ux.md` D-24 confirmada el 2026-10-12): hasta **5 ítems** de la agenda —eventos y actividades juntos, los más cercanos, con imagen— en una fila de tarjetas compactas, **prefiriendo los próximos 2 meses con respaldo** a los siguientes próximos; con nada publicado muestra el mensaje «Por el momento no hay eventos ni actividades» **sin botón** (única excepción a la regla de secciones vacías), y el botón **«Ver más eventos»** lleva a `/eventos`. Se sirve con **`GET /api/v1/portada/proximamente?lang=es|en`** (`no-store`, mismo filtro «publicado y próximo» que la sección).
- **Bilingüe (español/inglés)** con el mismo patrón de F3: el inglés es opcional contenido por contenido (fallback `en → es`, nunca campos vacíos) y las fechas/horas se localizan por idioma sin alterar el valor.
- **Solo lo publicado y lo próximo** (retiro automático al pasar la fecha): la regla se aplica como filtro de lectura en el servidor, en la zona horaria **`APP_TZ`** (`America/Costa_Rica`), sin cron y sin cambiar estados; en el panel el elemento sigue publicado y se marca como pasado.
- **Panel con tres rutas y tres permisos independientes**: `/panel/eventos`, `/panel/actividades` y `/panel/categorias` —permisos «Eventos», «Actividades» y **«Categorías de agenda»** (el nuevo `categorias`, **10.º del catálogo de F2**, sembrado por la migración `000009`)—, cada elemento con estado **Borrador/Publicado** (se publica al guardar), subida de imagen, catálogo de categorías opcional (borrar una categoría en uso → `409`) y todo registrado en `/panel/auditoria` con **15 códigos nuevos** `event.*`/`activity.*`/`category.*`.
- **API y datos**: `GET /api/v1/proximamente?lang=es|en` (`no-store`, sección completa) y `GET /api/v1/portada/proximamente?lang=es|en` (extracto de hasta 5 ítems del bloque de la portada), más el subgrupo `/api/v1/admin/{eventos,actividades,categorias}`; contrato vivo a **0.6.0**; migraciones `000007`–`000009`; sin dependencias nuevas.

Rama: `004-eventos-actividades`. Spec, plan, UX y quickstart en [`specs/004-eventos-actividades/`](specs/004-eventos-actividades/).

## Qué es F6 (Ministerios)

La funcionalidad **F6** es la segunda sección de contenido dinámico del sitio: **las áreas de servicio de la iglesia**, visibles sin cuenta, con el módulo del panel que las administra y su bloque en la portada.

- **Sección pública en <http://localhost:5173/ministerios>**: una ficha por ministerio con **nombre, descripción, horario y lugar** (textos libres, sin fechas calculadas), **imagen opcional**, **encargado opcional** (nombre y, si se quiere, **un** medio de contacto: celular o correo) y el botón **«Unirse al grupo de WhatsApp»** cuando hay enlace. La entrada **«Ministerios»** de la barra, el menú móvil y el pie aparece **solo cuando hay ministerios publicados**; **sin ninguno publicado, `/ministerios` redirige a la portada** de forma **silenciosa y sin aviso** —a propósito distinto de `/eventos`, que muestra un aviso sin salir— (decisión del cliente, puerta 2-B).
- **Privacidad del encargado, «oculto por defecto»**: el interruptor **«Mostrar la ficha del encargado en el sitio público»** nace **apagado**; con él apagado **ningún dato del encargado —ni el nombre— sale del servidor** hacia la sección, el bloque de la portada ni los registros, mientras que en el panel el dato **sigue guardado íntegro**. Publicar nombre o contacto exige encender el interruptor de forma expresa.
- **Bloque «Ministerios» en la portada**, colocado **después del bloque «Lo que viene» y antes de «Quiénes somos»**: hasta **5 fichas** por el orden manual (imagen, chip «Ministerio», nombre y horario/lugar; **nunca** encargado ni enlace) con el botón **«Ver más ministerios»**; sin ningún ministerio publicado el bloque **no aparece** (otra diferencia deliberada con el bloque de F4, que sí muestra un mensaje).
- **Bilingüe (español/inglés)** con el mismo patrón de F3/F4: el inglés es opcional contenido por contenido (fallback `en → es`, nunca campos vacíos) y la entrada del menú se recuerda en el dispositivo.
- **Panel `/panel/ministerios`** bajo el permiso **«Ministerios»** de F2 —**ya existente: F6 no añade permisos ni migración de permisos** y el catálogo sigue en **10**—: alta, edición, publicación/retirada por elemento, eliminación con confirmación, interruptor de visibilidad del encargado, **reordenación manual** y subida de imagen; todo queda en `/panel/auditoria` con **8 códigos nuevos** `ministry.*` (incluidas las denegaciones).
- **API y datos**: `GET /api/v1/ministerios?lang=es|en` (sección) y `GET /api/v1/portada/ministerios?lang=es|en` (bloque, tope 5 resuelto en el servidor, `no-store`), más el subgrupo `/api/v1/admin/ministerios` (lista, alta, `GET`/`PATCH`/`DELETE` por id e imagen); contrato vivo a **0.7.0**; migraciones **`000010`** (tabla `ministries`) y **`000011`** (los 8 códigos de auditoría); sin dependencias nuevas ni variables de entorno nuevas.

Rama: `006-ministerios`. Spec, plan, UX y quickstart en [`specs/006-ministerios/`](specs/006-ministerios/).

## Qué es F7 (Ofrendas y diezmos)

La funcionalidad **F7** es la tercera sección de contenido dinámico del sitio: **cómo apoyar a la iglesia**, con el módulo del panel que la administra y su bloque en la portada. La sección pública se llama **«Ofrendas y diezmos»** y vive en **`/ofrendas`**; el permiso y el módulo del panel siguen llamándose **«donaciones»**.

- **Sección pública en <http://localhost:5173/ofrendas>**: visible **sin cuenta ni registro**, con el **texto introductorio** («cómo donar»), el grupo de **cuentas IBAN** y el grupo de **números de SINPE Móvil** (agrupado por tipo, con **orden manual** dentro de cada grupo) y el **texto de uso de las donaciones**. Cada ficha muestra **tipo, número, entidad financiera y —si existe— titular**, y la **moneda (colones o dólares) solo en las cuentas IBAN**. El **IBAN se muestra agrupado en bloques** y el **SINPE en su formato normalizado**, con un botón **«Copiar número»** que copia el **número exacto guardado**. **Sin imágenes** y **bilingüe es/en** con fallback al español (los números y datos de la cuenta **no** se traducen).
- **Números validados**: el **IBAN** por estructura y longitud (costarricense e internacional); el **SINPE Móvil** como **celular costarricense de 8 dígitos**, con o sin `+506`, normalizado al guardar. Un número mal formado se rechaza con un aviso claro y **no se guarda nada a medias**.
- **Aparece solo cuando hay contenido**: mientras no haya nada publicado, la entrada **«Ofrendas y diezmos»** no se muestra y **`/ofrendas` redirige a la portada** de forma silenciosa y sin aviso (a propósito distinto de la sección de eventos, que muestra un aviso sin salir).
- **Bloque «Ofrendas y diezmos» en la portada**, colocado **justo después del bloque de ministerios**: hasta **5 métodos** (los primeros según el orden de la sección) con el botón **«Ver formas de ofrendar»**; **si no hay ningún método publicado, el bloque no aparece**.
- **Panel `/panel/donaciones`** bajo el permiso **«Donaciones»** de F2 —**ya existente: F7 no añade permisos ni migración de permisos** y el catálogo sigue en **10**—: listado agrupado por tipo con orden numérico, alta/edición/eliminación, publicación y retirada **por cada cuenta o número** y **por cada texto**, y validación completa en el servidor (duplicado por tipo+número → `409`). Todo queda en `/panel/auditoria` con **8 códigos** `donation.*` (incluidas las denegaciones). Acceso directo **«Gestión de donaciones»** en Inicio.
- **API y datos**: `GET /api/v1/ofrendas?lang=es|en` (sección) y `GET /api/v1/portada/ofrendas?lang=es|en` (bloque, tope 5 resuelto en el servidor, `no-store`), más el subgrupo `/api/v1/admin/donaciones` (lista, alta, `GET`/`PATCH`/`DELETE` por id, `GET /textos` y `PUT /textos/{textKey}`); contrato vivo a **0.8.0**; migraciones **`000012`** (`donation_methods` y `donation_texts`) y **`000013`** (los 8 códigos de auditoría); sin dependencias nuevas ni variables de entorno nuevas.

Rama: `007-donaciones`. Spec, plan, UX y quickstart en [`specs/007-donaciones/`](specs/007-donaciones/).

## Prerrequisito

**Docker** (con Docker Compose v2) en ejecución. Es el único requisito (FR-001 de la spec de F1): el clon limpio levanta **sin `.env`** y sin tener Go ni Node instalados.

Una vez por clon, activa los hooks de git (plan R3):

```bash
make instalar-hooks
```

## Quickstart

```bash
git clone <repo>
cd simiente_santa
make instalar-hooks   # solo la primera vez por clon
make up               # docker compose up -d: db + redis + backend + frontend
```

Comprobar:

```bash
docker compose ps                      # db y redis (healthy), backend y frontend en ejecución
curl -i http://localhost:8080/healthz  # 200 → {"status":"ok","database":"connected"}
```

Abrir <http://localhost:5173>: desde F3 esa es la **portada pública** de la iglesia; con la base recién migrada se muestra sin secciones, porque no hay contenido publicado todavía (el equipo lo carga desde el panel). Desde F4 esa misma portada trae el bloque **«Lo que viene»** (justo después de la imagen de bienvenida y antes de «Quiénes somos»), con hasta 5 eventos y actividades próximos y el botón **«Ver más eventos»**; si no hay nada publicado y próximo, el bloque muestra un mensaje y la entrada de navegación sigue oculta. Desde F10 esa portada se navega con **barra fija en escritorio** y con el **botón «Menú» en móvil**, y trae el **selector de idioma compacto**, el **interruptor de tema claro/oscuro** y el **mapa del sitio en el pie**. También existe la sección **<http://localhost:5173/eventos>** —«Eventos y actividades»—, que aparece en la barra, en el menú y en el pie **solo cuando hay contenido publicado y próximo**; sin nada publicado muestra un aviso sin salir de la página. Desde F6 existe también **<http://localhost:5173/ministerios>** —«Ministerios»— con la misma entrada condicional en la navegación, pero **sin ningún ministerio publicado la dirección redirige a la portada** (sin aviso y sin parpadeo), y la portada trae el bloque **«Ministerios»** entre «Lo que viene» y «Quiénes somos» (oculto por completo si no hay nada publicado). La pantalla «Estado del sistema» de F1, que antes ocupaba el inicio, está ahora en <http://localhost:5173/health>. El **panel de administración** está en <http://localhost:5173/login> y todavía no tiene cuentas: hay que inicializarlo (siguiente sección).

Detener: `make down` (los datos se conservan en el volumen `pgdata`; `make down && make up` es repetible).

### Puesta en marcha del panel (solo la primera vez)

El primer administrador se crea con una **inicialización única** por API: solo funciona si el sistema no tiene ninguna cuenta y solo con el token `BOOTSTRAP_TOKEN` (FR-007).

1. Crea el `.env` y pon ahí un token (genera uno propio con `openssl rand -base64 32`; en local sirve un valor de desarrollo como el que usan las pruebas e2e, `dev-bootstrap-token`):

   ```bash
   cp .env.example .env      # y editar BOOTSTRAP_TOKEN
   make up                   # compose relee el .env y recrea el backend
   ```

2. Ejecuta la inicialización (datos de ejemplo; la contraseña cumple la política):

   ```bash
   curl -i -X POST http://localhost:8080/api/v1/setup/initialize \
     -H "Content-Type: application/json" \
     -H "X-Setup-Token: $BOOTSTRAP_TOKEN" \
     -d '{"firstName":"Ana","lastName":"Responsable","email":"ana@ejemplo.com","phone":"+34 612 345 678","password":"Semilla.2026"}'
   ```

   Esperado: `201` con la cuenta creada (rol «Administrador», los **10** permisos del catálogo —desde F4 se añade «Categorías de agenda»—). Repetirlo → `409` «La inicialización ya se hizo y no puede repetirse»; sin la cabecera o con el token equivocado → `401`/`403`. **Sin `BOOTSTRAP_TOKEN` la inicialización queda desactivada** (siempre `401`/`403`).

3. Entra en <http://localhost:5173/login> con ese correo y contraseña. La validación paso a paso está en [`specs/002-acceso-gestion-usuarios/quickstart.md`](specs/002-acceso-gestion-usuarios/quickstart.md) §1–§3.

### Puertos

| Servicio | Puerto en el host | Puerto dentro del contenedor | Variable |
|---|---|---|---|
| Frontend (nginx) | **5173** | 80 | `WEB_PORT` |
| Backend (API) | **8080** | 8080 (**fijo**) | `HTTP_PORT` |
| PostgreSQL | **5432** | 5432 | `DB_PORT` |
| Redis (sesión y contadores) | **6379** | 6379 | `REDIS_PORT` |

El backend **siempre escucha en el 8080 dentro de su contenedor**; `HTTP_PORT` solo parametriza el puerto publicado en el host (`"${HTTP_PORT:-8080}:8080"` en `docker-compose.yml`).

Con puertos propios (útil si alguno está ocupado):

```bash
DB_PORT=5433 HTTP_PORT=8081 WEB_PORT=5174 make up
```

**Al cambiar puertos hay que reconstruir** (ver «Notas operativas»: `VITE_API_URL` se hornea en build):

```bash
HTTP_PORT=8081 WEB_PORT=5174 CORS_ALLOWED_ORIGINS=http://localhost:5174 docker compose up -d --build
```

Dos detalles al cambiar puertos:

- **`WEB_PORT`**: hay que añadir el origen nuevo a `CORS_ALLOWED_ORIGINS` (por defecto `http://localhost:5173`); si no, el navegador bloquea las llamadas del frontend a la API.
- **`HTTP_PORT`**: compose deriva de él el `VITE_API_URL` con el que se compila la imagen del frontend; sin `--build` esa imagen sigue apuntando al puerto anterior.

### `.env` opcional

```bash
cp .env.example .env   # editar valores; .env está en .gitignore
make up
```

Todas las variables tienen valor por defecto de desarrollo; ninguna es obligatoria para levantar (FR-001 de la spec de F1). Lo único que exige tocar `.env` es **inicializar el panel**, que necesita `BOOTSTRAP_TOKEN` (ver «Puesta en marcha del panel»).

| Variable | Quién la lee | Por defecto |
|---|---|---|
| `APP_ENV` | backend (`platform/config`) | `development` |
| `HTTP_PORT` | backend (local) / publicación del puerto en el host | `8080` |
| `DATABASE_URL` | backend y `make db-migrate` (desde el host apunta a `localhost`) | `postgres://app:app_dev_password@localhost:5432/app?sslmode=disable` |
| `LOG_LEVEL` | backend (`debug`/`info`/`warn`/`error`) | `info` |
| `CORS_ALLOWED_ORIGINS` | backend (middleware CORS) | `http://localhost:5173` |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | servicio `db` de compose (y con ellos compose construye el `DATABASE_URL` **del contenedor**, apuntando al host `db`) | `app` / `app_dev_password` / `app` |
| `DB_PORT` / `REDIS_PORT` / `WEB_PORT` | compose (puertos del host) | `5432` / `6379` / `5173` |
| `DATABASE_URL_TEST` | pruebas de integración del backend (`-tags=integration`); sin ella, esos tests **se omiten en silencio** | la trae `.env.example` apuntando a `app_test` en `localhost:5432`; que apunte **solo a una BD efímera**: esas pruebas vacían las tablas del dominio y arruinarían el stack local |
| `VITE_API_URL` | build de la imagen del frontend (URL de la API que usa el navegador) | `http://localhost:8080` |
| `REDIS_URL` | backend (sesión en Redis y contadores de intentos); compose la fija a `redis://redis:6379/0` dentro de la red | `redis://localhost:6379/0` (solo fuera de Docker: tests y scripts) |
| `SESSION_SECRET` | backend (`platform/config`): firma HMAC de la cookie CSRF | valor de ejemplo (≥ 32 caracteres); **obligatoria en producción** (el arranque falla si está vacía o corta) |
| `SESSION_COOKIE_SECURE` | backend (`Secure` en la cookie de sesión) | `false` (con HTTPS → `true`; en producción el arranque falla con `false`) |
| `SESSION_IDLE_TTL_MINUTES` / `SESSION_ABSOLUTE_TTL_MINUTES` | backend (TTL de la sesión en Redis) | `30` (inactividad) / `60` (vida absoluta desde el login) |
| `BOOTSTRAP_TOKEN` | backend: cabecera `X-Setup-Token` de `POST /api/v1/setup/initialize` (inicialización única) | vacío → la inicialización queda desactivada (la ruta siempre responde `401`/`403` y el backend avisa al arrancar) |
| `UPLOAD_DIR` | backend (carpeta donde guarda y lee las imágenes de la portada: logotipo e imagen); compose la fija a `/var/lib/simiente/uploads` con el volumen `uploads_data` | `./uploads` (fuera de Docker) |
| `UPLOAD_MAX_BYTES` | backend (tope de tamaño de cada imagen subida, comprobado en stream) | `8388608` (**8 MB**) |
| `APP_TZ` | backend (`platform/config`): zona horaria del lugar con la que el servidor calcula el corte «hoy» de la vista pública (regla «publicado **y** próximo» de `/eventos` y de las imágenes); se valida como zona IANA al arrancar y un valor inválido detiene el arranque | `America/Costa_Rica` |

## Comandos

| Comando | Qué hace |
|---|---|
| `make up` | Levanta `db` + `redis` + `backend` + `frontend` (`docker compose up -d`, con build de las imágenes si faltan; también recrea el backend si cambió `.env`) |
| `make down` | Detiene el entorno conservando los datos (`pgdata`) |
| `make test` | Pruebas de backend (`go test ./...` + `-tags=integration`) y frontend (Vitest) |
| `make lint` | `gofmt` + `go vet` + `golangci-lint` (backend) y ESLint + `tsc` (frontend) |
| `make security` | `govulncheck ./...` (Go) y `npm audit --audit-level=high` (frontend) |
| `make ci` | `lint` + `test` + `security`: lo mismo que corre el CI |
| `make db-migrate` | Aplica las migraciones de `backend/migrations` con `golang-migrate` (F2 añadió `000002`–`000004`; F3, `000005`–`000006`; F10 no añade ninguna; F4, `000007`–`000009`; F6, `000010`–`000011`) |
| `make doctor` | Verifica que el entorno tenga todo lo necesario (Docker, Go, Node, hooks, kit…) |
| `make e2e` | Pruebas end-to-end con Playwright (requiere `make up` levantado) |
| `make api-gen` | Regenera `frontend/src/api/schema.d.ts` desde `backend/api/openapi.yaml` |
| `make sqlc-gen` | Regenera el código Go de consultas de `backend/internal/db/` |
| `make sqlc-verify` | Regenera y exige `git diff --exit-code` (sin deriva de artefactos) |
| `make instalar-hooks` | **Una vez por clon**: activa los hooks de git (`.githooks`) |
| `make help` | Lista los comandos disponibles |
| `make estado` | Por dónde va el proyecto: roadmap, fase, aprobaciones y próximo paso |
| `make costos` | Costo de IA de la tarea actual (`CERRAR=1` lo congela al aprobar el PR; `TODO=1` resume el proyecto) |
| `make novedades` | Historial de cambios del kit (`DESDE=1.4.0` para ver desde una versión) |

Notas:

- **`make db-migrate` usa `$DATABASE_URL` del shell** (no lee `.env` por sí mismo). Con `.env`: `set -a; source .env; set +a` antes de llamarlo; sin `.env`: `DATABASE_URL='postgres://app:app_dev_password@localhost:5432/app?sslmode=disable' make db-migrate`.
- **`make sqlc-verify` comprueba que no hay deriva** entre `backend/internal/db/queries/` y el código generado: F2 añadió las consultas `users`, `roles`, `permissions` y `audit`, F3 las del dominio `portada` (`home.sql`), F4 las de `agenda` y `media` (`agenda.sql`, `media.sql`) y F6 las de `ministries` (`ministries.sql`, además de ampliar `media.sql` con el estado de publicación de los ministerios) (en F1 estaba vacío y el comando terminaba en verde sin comprobar nada).

## Herramientas de desarrollo (opcionales)

Solo hace falta instalarlas para trabajar **fuera** de Docker (tests, lint, generación). Verificables con `make doctor`. Versiones fijadas para generación reproducible (plan R4):

| Herramienta | Versión fijada | Dónde se fija / cómo se obtiene |
|---|---|---|
| Go | **1.27.2** | `go 1.27.2` en `backend/go.mod`, `golang:1.27.2` en `backend/Dockerfile`; el CI lee `backend/go.mod` |
| Node.js | **24** | `NODE_VERSION: '24'` en `.github/workflows/ci.yml` (kit 1.6.4); `node:24.21-alpine` en `frontend/Dockerfile` |
| `golang-migrate` | **v4.20.1** | `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.20.1` (mismo pin que el CI del kit 1.6.4) |
| `sqlc` | **v1.31.1** | instalado con `go install …@v1.31.1`; el binario queda en `$(go env GOPATH)/bin`, que debe estar en el `PATH` |
| `openapi-typescript` | `^7.13.0` | `devDependency` de `frontend/package.json`, fijada en `package-lock.json` (`npm ci`) |
| Playwright | `@playwright/test` `^1.63.0` | `devDependency` de `frontend/package.json` |

Otras herramientas del flujo de calidad: `golangci-lint` **v2.14.0** y `govulncheck` **v1.8.0** (los usan `make lint` y `make security`; el CI del kit 1.6.4 fija esas mismas versiones) y `gitleaks` (hooks de git y el paso de secretos del CI).

Para `make e2e` hace falta además el **navegador** de Playwright:

```bash
cd frontend
npx playwright install chromium
sudo npx playwright install-deps chromium   # solo Linux: librerías del sistema
```

## Versiones y soporte de seguridad (FR-016 / SC-010)

Comprobado a **2026-10-05** (Redis se añadió en F2; el resto de la tabla, a 2026-10-03):

| Tecnología | Versión en uso | Estado de soporte |
|---|---|---|
| Go | 1.27.2 (parche de la rama 1.27) | Vigente (solo 1.26 y 1.27 reciben parches; 1.23 está en fin de vida desde 2025-08-12) |
| Node.js | 24 (LTS «Krypton») | **LTS activa** (mantenimiento desde 2026-10-20; fin de soporte 2028-04-30); el kit 1.6.4 fija esta línea mayor en el CI |
| PostgreSQL | 16 (imagen `postgres:16-alpine` en desarrollo — **decisión del 2026-10-03**, T034/plan R10) | Rama 16 en soporte hasta noviembre de 2028 y recibe parches de seguridad; fijar un minor/digest exacto para reproducibilidad queda pendiente para builds/despliegue |
| Redis | 7 (imagen `redis:7-alpine`, servicio `redis` en `docker-compose.yml`; añadida en F2, T205 — decisión P23 del plan) | Series 7.2 y 7.4 de Redis OSS en soporte extendido hasta **2029-12-01** (ciclo oficial de versiones de Redis, consultado el 2026-10-05); fijar un minor/digest exacto queda pendiente, igual que con `postgres:16-alpine` |
| React / TypeScript / Vite | 19 / 5.x / actual | Mantenidas; vulnerabilidades vía `npm audit` |
| `pgx/v5`, `sqlc`, `golang-migrate`, `openapi-typescript` | fijadas en `go.sum`, `package-lock.json` y este README | Herramientas de desarrollo; `govulncheck` y `npm audit` sin altas/críticas |

**Pendientes de actualización conocidos** (US7 esc. 2; identificados, propuestos y gestionados por el humano):

- **Resueltos por el kit 1.6.4** (`2957bb0`, actualizado el 2026-10-03): `postgres:16-alpine` en el CI (R10/T034), Node 24 en el CI (R11), `golangci-lint` v2 en el CI (g1) y «Go 1.26+» en `docs/GUIA-INICIO.md`. Ya no son pendientes del proyecto.
- **Sin pendientes abiertos** por el momento. Queda para builds/despliegue fijar un **minor/digest exacto** de las imágenes (FR-016).

Los archivos del kit no se editan aquí (regla 10). **Política (FR-016)**: toda funcionalidad que fije o actualice una versión revisa esta tabla; lo que deje de estar en soporte se registra como pendiente antes de seguir construyendo.

## Notas operativas

- **`VITE_API_URL` se hornea en el build** (plan R6): la URL de la API queda dentro del bundle del frontend. Por defecto `http://localhost:8080` y compose la deriva de `HTTP_PORT`. Por eso **cambiar puertos exige `docker compose up -d --build`** (o borrar la imagen del frontend), y al cambiar `WEB_PORT` hay que ajustar también `CORS_ALLOWED_ORIGINS`.
- **Artefactos generados (se commitean)**: `frontend/src/api/schema.d.ts` se regenera con `make api-gen` cuando cambia `backend/api/openapi.yaml`; el código de sqlc con `make sqlc-gen` cuando cambian `backend/migrations/` o `backend/internal/db/queries/`. Regla de revisión (plan R4): un PR que toca migraciones o consultas debe regenerar `internal/db/`, y uno que toca el contrato debe regenerar `schema.d.ts`. `make sqlc-verify` comprueba que no hay deriva.
- **E2E**: `make e2e` necesita el stack levantado (`make up`) y el navegador de Playwright instalado (arriba). El CI del kit **no** corre e2e. Notas de F2/F3: si en tu entorno faltan las librerías de sistema de Chromium (`libnspr4`, `libnss3`, `libasound2`) y no puedes instalarlas, las suites se pueden correr con la imagen oficial de Playwright (`mcr.microsoft.com/playwright:v1.63.0-jammy`) contra el stack levantado; y conviene ejecutarlas **en serie** (`--workers=1`), porque en paralelo saturan el rate-limit de 20 peticiones/min por IP a `/api/v1/auth/login` y producen `429` espurios. Así se validaron F2 (**3/3**) y F3 (**7/7**: `acceso`, `auditoria`, `portada-accesibilidad`, `portada-panel`, `portada-patch-null`, `portada-publica` y `status`). En F10 (2026-10-11) se corrieron en **dos tandas** —las suites congeladas de F2/F3 (**8/8**, incluida `portada-imagenes`) y las nuevas de F10 (**25/25**: `diseno-movimiento`, `navegacion-movil`, `tema-oscuro`, `movimiento`, `movimiento-reducido` y `qa-medidas-f10`)—. En F4 (2026-10-11) se corrieron en **tres tandas** —F2 (**2/2**: `acceso`, `auditoria`), F4 (**5/5**: `agenda-publica`, `agenda-panel`, `agenda-accesibilidad`) y F3/F10 (**31/31**: `portada-`, `diseno-`, `movimiento`, `navegacion-`, `tema-`, `qa-` y `status`)—, porque el rate-limit de login también afecta a estas suites; y en la **re-validación del 2026-10-12** (3 bugs corregidos + bloque «Lo que viene» de la portada) en **tres tandas** con **43 casos**: A F3/F10 + bloque (**33/33**, incluidos los 4 de `portada-proximamente`), B F4 + bugs (**8/8**, con `qa-bugs-004`) y C F2 (**2/2**); la **tanda 2** del mismo día (arreglo del fondo negro de `/login`, `/cambiar-contrasena`, `/sin-permiso` y 404 en tema oscuro + reorden del bloque de la portada) se corrió en **cuatro tandas** con **46 casos**: nueva tanda de regresión `qa-bug-fondo-oscuro-fuera-layouts` **3/3**, A **33/33**, B **8/8** y C **2/2**; la imagen Docker de Playwright sigue siendo válida donde falten las librerías del navegador.

- **E2E de F6 (2026-10-13)**: la suite completa se corre en **cuatro tandas** con `--workers=1` (el límite de 20 intentos de `/api/v1/auth/login` por minuto y por IP lo impone): **1** F2 (`acceso`, `auditoria`), **2** F4 (`agenda-`), **3** F3/F10 (`portada- diseno- movimiento navegacion- tema- qa- status`; el filtro `portada-` arrastra también `portada-ministerios`) y **4** F6 (`ministerios- portada-ministerios`). La tanda 4 de F6 salió **19/19**; los comandos exactos están documentados en la cabecera de `frontend/e2e/ministerios-publica.spec.ts`.

- **E2E de F7 (2026-10-15)**: la suite completa se corre en **cinco tandas** con `--workers=1` (el mismo límite de login): **1** F2 (`acceso`, `auditoria`), **2** F4 (`agenda-`), **3** F3/F10 (`portada- diseno- movimiento navegacion- tema- qa- status`; el filtro `portada-` arrastra también `portada-ministerios` y `portada-ofrendas`), **4** F6 (`ministerios-`) y **5** F7 (`ofrendas- portada-ofrendas`). La tanda 5 de F7 salió **23/23**; los comandos exactos están documentados en la cabecera de `frontend/e2e/ofrendas-publica.spec.ts`.

## Documentación relacionada

- [`CHANGELOG.md`](CHANGELOG.md) — cambios notables por versión (Keep a Changelog + SemVer): la 0.1.0 corresponde a F1, la 0.2.0 a F2, la 0.3.0 a F3, la 0.4.0 a F10, la 0.5.0 a F4, la 0.6.0 a F6 y la 0.7.0 a F7.
- [`docs/entrega/F1-estructura-base.md`](docs/entrega/F1-estructura-base.md) — notas de entrega de F1 para el cliente (lenguaje no técnico).
- [`docs/entrega/F2-acceso-y-gestion-de-usuarios.md`](docs/entrega/F2-acceso-y-gestion-de-usuarios.md) — notas de entrega de F2 para el cliente (lenguaje no técnico).
- [`docs/entrega/F3-portada-e-informacion-general.md`](docs/entrega/F3-portada-e-informacion-general.md) — notas de entrega de F3 para el cliente (lenguaje no técnico).
- [`docs/entrega/F4-eventos-y-actividades.md`](docs/entrega/F4-eventos-y-actividades.md) — notas de entrega de F4 para el cliente (lenguaje no técnico).
- [`docs/entrega/F10-sistema-de-diseno-y-movimiento.md`](docs/entrega/F10-sistema-de-diseno-y-movimiento.md) — notas de entrega de F10 para el cliente (lenguaje no técnico).
- [`docs/entrega/F6-ministerios.md`](docs/entrega/F6-ministerios.md) — notas de entrega de F6 para el cliente (lenguaje no técnico).
- [`docs/entrega/F7-ofrendas-y-diezmos.md`](docs/entrega/F7-ofrendas-y-diezmos.md) — notas de entrega de F7 para el cliente (lenguaje no técnico).
- [`specs/001-estructura-base/quickstart.md`](specs/001-estructura-base/quickstart.md) — validación ejecutable de F1 de punta a punta (qué se espera en cada escenario).
- [`specs/002-acceso-gestion-usuarios/quickstart.md`](specs/002-acceso-gestion-usuarios/quickstart.md) — validación ejecutable de F2 (inicialización, sesión, usuarios, roles y auditoría).
- [`specs/003-portada-info-general/quickstart.md`](specs/003-portada-info-general/quickstart.md) — validación ejecutable de F3 (portada pública, edición desde el panel, publicación por elemento, imágenes, idioma y auditoría).
- [`specs/004-eventos-actividades/quickstart.md`](specs/004-eventos-actividades/quickstart.md) — validación ejecutable de F4 (sección `/eventos`, permisos, panel de eventos/actividades/categorías, retiro automático, idioma, auditoría y bloque «Lo que viene» de la portada en §15).
- [`specs/010-sistema-diseno-movimiento/quickstart.md`](specs/010-sistema-diseno-movimiento/quickstart.md) — validación ejecutable de F10 (sistema visual, tema claro/oscuro, movimiento y movimiento reducido, navegación por dispositivo, mapa del sitio e idioma compacto, con checklist de aceptación en §9).
- [`specs/006-ministerios/quickstart.md`](specs/006-ministerios/quickstart.md) — validación ejecutable de F6 (sección `/ministerios`, permisos, panel, privacidad del encargado, idioma, auditoría y bloque de la portada en §14, con mapa de criterios en §16).
- [`docs/tecnico/arquitectura.md`](docs/tecnico/arquitectura.md) — arquitectura del sistema; **§8 es la receta** de 10 pasos para agregar un área de negocio nueva.
- [`docs/tecnico/decisiones.md`](docs/tecnico/decisiones.md) — decisiones de arquitectura y su justificación.
- [`docs/GUIA-INICIO.md`](docs/GUIA-INICIO.md) — guía del kit: agentes de IA, entorno WSL y problemas comunes.
- [`docs/producto/roadmap.md`](docs/producto/roadmap.md) — funcionalidades F1–F10 (y F10B) y su estado.
- [`.specify/memory/constitution.md`](.specify/memory/constitution.md) — reglas no negociables del proyecto.
