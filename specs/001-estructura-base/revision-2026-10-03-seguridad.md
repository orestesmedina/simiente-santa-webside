# Revisión de seguridad — F1 Estructura base (T001–T029)

**Fecha:** 2026-10-03 · **Rama:** `001-estructura-base` · **Auditor:** `seguridad` (solo lectura)
**Alcance:** F1 completa (T001–T029): `backend/`, `frontend/`, `docker-compose.yml`, `.env.example`, Dockerfiles, `ci.yml` (kit), `README.md`, `quickstart.md`, historial de git (60 commits).
**Referentes:** constitución §IV · spec FR-008, FR-013, SC-009, SC-010, Out of Scope (F2 = authn) · plan R10/R11/R12 · arquitectura §5.11, §6 · D-A7 (pendiente de confirmación humana).

**Método:** `gitleaks detect --no-git` (árbol) y `gitleaks detect` (historial, 60 commits) · `govulncheck ./...` (backend) · `npm audit --audit-level=high` (frontend) · greps de secretos/XSS/cookies/SQL · verificación en vivo con `curl` contra el stack levantado · experimento aislado de redacción de pgx (en `/tmp/opencode`, limpiado; no se tocó el repo).

---

## Veredicto

**RECHAZADO** — por **un único hallazgo bloqueante (B1)**: `govulncheck ./...` falla (exit 3) por **GO-2026-5970** (`golang.org/x/text` v0.29.0, vía `pgx`), lo que deja en rojo `make security`/`make ci` y el paso "Vulnerabilidades" del pipeline (constitución §IV, FR-006/FR-007, SC-003/SC-010). **Todo lo demás está conforme**: sin secretos (árbol e historial limpios), sin fugas de información interna (FR-013/SC-009 verificado en código, pruebas y vivo), contenedores y OWASP correctos para el alcance de F1, `npm audit` limpio, y el único endpoint público es exactamente lo que la spec aprobó.

El bloqueante es de **corrección trivial** (bump de una dependencia indirecta, una línea en `go.mod`): corregido B1 y re-ejecutado `make security` en verde, el resto de hallazgos son **menores** y no rechazan (procedería *aprobado con observaciones* sin necesidad de re-auditar lo demás).

---

## Hallazgo bloqueante

### B1 · GO-2026-5970: bucle infinito en `golang.org/x/text` v0.29.0, alcanzable vía pgx — deja `govulncheck`/CI en rojo

- **Dónde:** `backend/internal/platform/database/database.go:32` (`pgxpool.NewWithConfig`) · dependencia indirecta: `backend/go.mod:12` (`golang.org/x/text v0.29.0 // indirect`, traída por `github.com/jackc/pgx/v5 v5.11.0`).
- **CWE:** CWE-835 (bucle con iteración excesiva → denegación de servicio).
- **Evidencia** (2026-10-03, `cd backend && govulncheck ./...`, exit 3):

  ```
  Vulnerability #1: GO-2026-5970 — Infinite loop on invalid input in golang.org/x/text
    Module: golang.org/x/text — Found in: v0.29.0 — Fixed in: v0.39.0
    #1..#3: database.NewPool calls pgxpool.NewWithConfig, which eventually calls norm.Form.Properties/Span/Transform
  Your code is affected by 1 vulnerability from 1 module.
  ```

  La ruta es la autenticación SCRAM de pgx (`pgconn/auth_scram.go` es el único archivo de pgx que importa `x/text`: normaliza usuario/contraseña con NFKC). `pgxpool.ParseConfig/NewWithConfig` y el `Ping` de `/healthz` la alcanzan.
- **Impacto:** (a) **DoS por CPU** (bucle infinito) si entra UTF-8 inválido a `norm.Iter`; la base de datos de la vulnerabilidad no asigna rating, clase práctica **moderada**; en F1 la entrada (usuario/contraseña de `DATABASE_URL`) la controla quien opera, no el cliente HTTP → **explotabilidad remota baja en F1**. (b) **Objetiva e inmediata**: el paso `Vulnerabilidades` del `ci.yml` (kit, no editable) y `make security`/`make ci` fallan con exit 3 → todo PR queda "no apto" (FR-006/FR-007, SC-003) y la constitución §IV ("DEBEN pasar `govulncheck`…") queda incumplida. F1 no puede cerrar en verde con esto.
- **Corrección concreta** (dev-backend): `cd backend && go get golang.org/x/text@v0.39.0 && go mod tidy && go test ./... && go vet ./... && govulncheck ./...` — bump de la indirecta, no toca código propio (la API de `x/text` no cambia). Re-ejecutar `make security` y confirmar exit 0. No hay riesgo de regresión para F1 (pgx solo usa `norm`).

---

## Hallazgos menores

> Ninguno rechaza. M1 y M4–M9 son endurecimiento barato; M1 requiere quedar **registrado como condición previa a cualquier despliegue**.

### M1 · El `nginx` del frontend corre el proceso master como root — **dictamen: asumible con nota en F1; condición previa a cualquier despliegue**

- **Dónde:** `frontend/Dockerfile:33-40` (sin directiva `USER`; imagen `nginx:1.30.5-alpine`) · `frontend/nginx.conf`.
- **CWE:** CWE-250 (ejecución con privilegios innecesarios).
- **Evidencia:** el Dockerfile final no declara `USER`; la imagen oficial arranca `nginx -g daemon off;` con master root y workers como usuario `nginx` (puerto 80 privilegiado). El backend, en cambio, sí es no-root: `backend/Dockerfile:46` (`USER nonroot:nonroot`, distroless, uid 65532) ✓.
- **Impacto:** una RCE en nginx daría **root dentro del contenedor** (mejor punto de partida para escalada/escape que un worker sin privilegios). En el alcance de F1 (entorno local de desarrollo; la spec excluye explícitamente el despliegue a producción; nginx 1.30.5 sirve **solo estáticos** con versión actual), la probabilidad de explotación es baja y el riesgo es auto-infligido.
- **Dictamen (lo que pide la tarea):** **no bloqueante para F1 — asumible con nota.** Dos razones para no dejarlo suelto: (1) el puerto del host ya es alto (`${WEB_PORT:-5173}:80`), y (2) estas imágenes son el **patrón que copiarán F2–F9**, así que la nota debe quedar registrada (mismo mecanismo que R10/T034). Debe resolverse **antes de cualquier despliegue** a un entorno visible fuera de la máquina de desarrollo.
- **Corrección concreta (para la funcionalidad de despliegue):** usar la imagen `nginxinc/nginx-unprivileged` (worker-only, puerto 8080), o bien en esta imagen: `listen 8080;` + `USER nginx` + en `nginx.conf`: `pid /tmp/nginx.pid;` y `error_log /dev/stderr;` con `client_body_temp_path`/`proxy_temp_path`/`fastcgi_temp_path`/`uwsgi_temp_path`/`scgi_temp_path` bajo `/tmp/`, y publicar como `${WEB_PORT:-5173}:8080`. Añadir al README/pendientes el ítem igual que R10.

### M2 · Imagen distroless referenciada con `:latest` (y el comentario del Dockerfile dice lo contrario)

- **Dónde:** `backend/Dockerfile:29` (`FROM gcr.io/distroless/static-debian12:latest`) vs. `backend/Dockerfile:8-10` ("La etiqueta va fijada (nada de `:latest`) para builds reproducibles").
- **CWE:** sin CWE directo — riesgo de cadena de suministro/reproducibilidad (etiqueta mutable).
- **Impacto:** `:latest` es mutable: dos builds del mismo commit pueden producir imágenes distintas y un repositorio o registry comprometido/repoblado entregaría contenido diferente sin señal en el repo. Además el comentario es inexacto (revisión futura confiará en un pinning que no existe). `golang:1.27`, `node:22.22-alpine` y `nginx:1.30.5-alpine` también son etiquetas flotantes (menor: al menos fijan minor).
- **Corrección concreta:** fijar por digest la etapa final (`gcr.io/distroless/static-debian12@sha256:…`, actualizándolo con criterio) o a una etiqueta versionada concreta; idealmente también las de build. Ajustar el comentario para que refleje la realidad.

### M3 · nginx expone su versión (`Server: nginx/1.30.5`) y la página 404 por defecto

- **Dónde:** `frontend/nginx.conf` (sin `server_tokens off;`).
- **CWE:** CWE-200 (exposición de información).
- **Evidencia (en vivo):** `curl -sI http://localhost:5173/` → `Server: nginx/1.30.5`; `curl -si http://localhost:5173/assets/no-existe.js` → 404 HTML por defecto de nginx con la versión.
- **Impacto:** fingerprinting exacto de versión para mapear CVEs conocidos. Bajo con versión al día; se arrastra al patrón de despliegue.
- **Corrección concreta:** `server_tokens off;` en `server {}` de `nginx.conf` (archivo del proyecto, editable) y, para despliegues, una página de error propia. (El backend Go no expone versión ni framework en las respuestas: verificado en vivo ✓.)

### M4 · Sin cabeceras de seguridad ni en nginx ni en el backend

- **Dónde:** `frontend/nginx.conf` y `backend/internal/platform/httpserver/{error.go,server.go}` (ninguno las fija).
- **CWE:** CWE-693 (mecanismo de protección no desplegado) — leve en F1.
- **Impacto:** falta `X-Content-Type-Options: nosniff` (API JSON + SPA), `Referrer-Policy`, `X-Frame-Options`/`frame-ancestors` (anti-clickjacking para el panel futuro). En F1 (página de estado, sin sesión) el riesgo es marginal.
- **Corrección concreta:** añadir `X-Content-Type-Options: nosniff` (y `Referrer-Policy: strict-origin-when-cross-origin`) en `nginx.conf` para el HTML; valorar CSP mínima con F3 (diseño). Puede esperar a F3 o al pre-despliegue; registrar como pendiente de despliegue junto a M1.

### M5 · CORS: `Allow-Methods`/`Allow-Headers` estáticos y más anchos que la superficie real de F1

- **Dónde:** `backend/internal/platform/middleware/cors.go:20-21` (`"GET, POST, PUT, PATCH, DELETE, OPTIONS"`, `"Content-Type, X-Request-ID"`).
- **CWE:** CWE-942 (política entre dominios permisiva) — en grado menor.
- **Evidencia (en vivo):** preflight permitido → `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS` cuando F1 solo expone `GET`. Orígenes no permitidos **no** reciben cabeceras CORS (verificado con `Origin: http://evil.example`: GET sin `Access-Control-Allow-Origin`, preflight 204 sin cabeceras) ✓; no hay `Access-Control-Allow-Credentials` ✓; `Vary: Origin` presente ✓.
- **Impacto:** bajo (CORS no protege al servidor; los métodos no registrados siguen recibiendo 405), pero contradice el "mínimo" de D16 y pre-concede verbos de escritura para todo origin permitido.
- **Corrección concreta:** estrechar la lista a lo registrado (`"GET, OPTIONS"` hoy) o derivarla de la configuración cuando lleguen endpoints escribibles (F2). Toque de un carácter en `cors.go`; puede ir con B1 en el mismo ciclo.

### M6 · `X-Request-ID` del cliente aceptado sin límite ni validación (se refleja y se loguea tal cual)

- **Dónde:** `backend/internal/platform/middleware/requestid.go:29-34` (solo `TrimSpace`).
- **CWE:** CWE-20 (validación de entrada insuficiente).
- **Impacto:** un cliente puede inyectar un valor arbitrariamente largo (hasta el tope de cabeceras del servidor, ~1 MB por defecto de Go) o con contenido extraño; se refleja en la respuesta y entra entero en los logs. El log es JSON de `slog` (escapa comillas/controles → no hay *log forging* real), pero ensucia la trazabilidad y permite inflar los logs.
- **Corrección concreta:** aceptar el valor del cliente solo si cumple `^[A-Za-z0-9-]{1,128}$` (si no, generar uno propio). Una función pequeña en `requestid.go`; candidato para F2 si no entra en este ciclo.

### M7 · `APP_ENV=production` no exige configuración explícita: los defaults de desarrollo siguen aplicando

- **Dónde:** `backend/internal/platform/config/config.go:69-94` (`Load` aplica defaults siempre; `EnvProduction` existe como constante pero no cambia el comportamiento) y `config.go:39-44` (defaults de desarrollo, incluida la contraseña del compose).
- **CWE:** CWE-453 (inicialización insegura por defecto).
- **Impacto:** en F1 no hay producción (Out of Scope de la spec), pero el interruptor ya existe: un despliegue futuro que olvide `DATABASE_URL` conectaría (fallaría) en silencio contra `localhost` con la contraseña de desarrollo, en vez de fallar al arrancar con un mensaje claro (que es el diseño declarado de `platform/config`).
- **Corrección concreta:** en F2 (o al definir despliegue): si `APP_ENV != development` (y != `test`), exigir `DATABASE_URL` y el resto de variables sin default y abortar al arrancar. Nota positiva verificada: pgx **redacta la contraseña** en errores de parseo (experimento: `cannot parse "postgres://app:xxxxx@…"` → `xxxxx`), así que ese fallo no filtraría el secreto al log.

### M8 · El 503 `database_unavailable` se loguea solo con el `code`, sin la causa interna

- **Dónde:** `backend/internal/platform/httpserver/error.go:137-143` (`logError` registra `code` a nivel `warn` para todo kind no-`Internal`; la causa envuelta con `%w` — p. ej. "connection refused" — se descarta).
- **CWE:** ninguno — observabilidad (no es fuga: la causa **no** debe salir al cliente, y no sale).
- **Evidencia (en vivo):** con la BD detenida (escenario de validación de QA), el log registra `{"level":"WARN","msg":"error de dominio","code":"database_unavailable"}` y nada más: desde el log no se distingue DNS/timeout/credenciales. Lo comprobé en primera persona durante esta auditoría.
- **Impacto:** diagnóstico difícil del fallo de conexión (FR-013/SC-009 exigen detalle en el log solo para errores **inesperados** (500), y §5.11 marca el 503 con "—" → **conforme al diseño aprobado**; esto es una mejora, no un incumplimiento).
- **Corrección concreta:** en `logError`, añadir para `KindDatabaseUnavailable` el atributo `error` con `domainErr.Error()` (la causa) a nivel `warn` — solo al log, nunca a la respuesta. Compatible con FR-013 (la causa nunca se serializa).

### M9 · Puertos de compose publicados en `0.0.0.0` (todos los interfaces), incluida la BD con contraseña de desarrollo

- **Dónde:** `docker-compose.yml:12` (`"${DB_PORT:-5432}:5432"`), `:38` y `:53` (backend/frontend).
- **CWE:** CWE-668 (recurso expuesto a la esfera incorrecta).
- **Impacto:** en una red compartida (oficina, LAN, WSL con reenvío de Windows), PostgreSQL queda alcanzable desde fuera de la máquina con la contraseña por defecto `app_dev_password` (visible en el repo). El auth es `scram-sha-256` para conexiones externas (la imagen oficial), pero el secreto es público. Solo afecta al entorno local de desarrollo.
- **Corrección concreta:** publicar en loopback: `"127.0.0.1:${DB_PORT:-5432}:5432"` (idem 8080/5173). Sigue sirviendo para `make db-migrate` y navegadores locales; reversible si algún día se necesita acceso externo. Decisión de `devops`/humano; puede ir con M1 en los pendientes.

---

## Notas para F2+ (sin hallazgo en F1; que no se pierdan al crecer)

1. **Límites de tamaño de cuerpo:** F1 no lee cuerpos (solo `GET`), así que no hay superficie de bodies enormes; con el **primer endpoint escribible** (F2), el handler debe usar `http.MaxBytesReader` (añadirlo a la receta §8 de `arquitectura.md` y a `platform/validate`).
2. **gzip + sesiones:** `nginx.conf` activa gzip. Cuando existan cookies de sesión y respuestas que reflejen entrada del cliente (BREACH/CRIME-style), revisar compresión por endpoint bajo TLS. Hoy no hay cookies ni reflejo ✓.
3. **Cache-Control en 404/405/500:** solo `/healthz` fija `no-store` (200 y 503), exactamente como documenta el contrato (D7) ✓. Los errores del fallback no lo llevan; valorar extenderlo cuando haya respuestas sensibles al paso.
4. **TLS:** todo va en claro (local). Con el primer despliegue: terminación TLS y `SESSION_SECRET`/cookies `Secure` (D-A7).
5. **`postgres:16.4-alpine`** (compose y `ci.yml` del kit) y **Node 22**: **pendientes ya registrados** por el plan/README (R10/R11, T034/T035) — confirmados por esta auditoría; no se reabren aquí.
6. **`app_ci_password` en `ci.yml`:** credencial efímera de un servicio de CI que se destruye en cada ejecución; práctica estándar, no es un secreto operativo (el archivo es del kit, no editable).

---

## Verificado sin hallazgos (evidencia de lo que SÍ está bien)

| # | Qué se auditó | Cómo se verificó | Resultado |
|---|---|---|---|
| 1 | **Secretos** (FR-008) | `gitleaks detect --no-git` (1,64 MB, árbol) y `gitleaks detect` (historial, 60 commits) → **0 hallazgos** ambos · `git ls-files` → `.env` **no trackeado** (`.gitignore:3-5`: `.env`, `.env.*`, `!.env.example`) · grep de patrones `password\|secret\|token\|api_key` en todo el repo → solo placeholders/defaults documentados · hook `pre-commit` bloquea `.env`/`*.pem` y corre `gitleaks protect` · job `secretos` del CI (gitleaks-action) | ✅ Sin secretos reales en árbol, historial, `docker-compose.yml`, `README.md`, reportes ni CI. Defaults de desarrollo **claramente marcados**: `config.go:36-44` ("valor de ejemplo, no un secreto real"), `docker-compose.yml:2-5`, `README.md:81` ("Todas las variables tienen valor por defecto de desarrollo"). `SESSION_SECRET` es placeholder (`generar_un_valor_aleatorio_largo`) y el README declara que F1 no la consume. `.env` local existe con los valores de ejemplo, sin trackear |
| 2 | **Fugas de información** (FR-013/SC-009) | Código: `WriteError` (`error.go:79-90`) responde mensaje genérico (`apperr.go:31,142-143`), **suprime `details` para `Internal`** (`error.go:84-86`) y loguea la causa con `request_id` (`error.go:124-143`) · suite `envelope_test.go:135-196` provoca 500 y panic y comprueba que el cuerpo no contiene `sql`/`db-interna`/`repository.go`/`goroutine`/`stack` y que el log sí lleva el detalle con `request_id` = `X-Request-ID` · en vivo: 200/404/405/503 todos en sobre · experimento pgx: la contraseña se redacta (`xxxxx`) en errores de parseo | ✅ El error interno no sale al cliente; el detalle queda en el log con `request_id`; README/quickstart/estado publican solo defaults de desarrollo, nunca credenciales operativas |
| 3 | **Servidor OWASP** | Código: timeouts en `server.go:22-28` + `withDefaults` (`main.go` solo pasa `Addr` → **aplican los defaults**: `ReadHeaderTimeout` 5 s, Read 10, Write 15, Idle 60 — mitiga slowloris) · en vivo: `POST /healthz` → **405** con sobre y `Allow: GET, HEAD`; ruta desconocida → **404** con sobre; `GET /healthz` → `Cache-Control: no-store` en 200 **y 503** · CORS en vivo: origen permitido → eco exacto + `Vary: Origin`; `evil.example` → **sin** `Access-Control-Allow-Origin`; preflight solo se concede a orígenes permitidos; **sin credenciales** | ✅ Sin `*` con credenciales; preflight no abre de más (más allá de M5); método/routing correcto; `no-store` donde el contrato lo exige |
| 4 | **Inyección/SQL** | F1 no ejecuta SQL de negocio (solo `pgxpool.Ping`; `internal/db/queries/` vacío); grep de concatenación SQL → nada | ✅ Sin superficie (la capa sqlc llega con F2, parametrizada por diseño) |
| 5 | **XSS y cliente** | grep `dangerouslySetInnerHTML\|innerHTML\|href=\|window.open\|eval\|new Function` en `frontend/src` → **0 coincidencias**; la UI renderiza literales fijos y estados de enums; sin cookies/`localStorage`/`sessionStorage` (`grep Set-Cookie\|document.cookie` → 0) | ✅ Sin superficie XSS en F1; sesión en storage ni cookies no existe (correcto: F2) |
| 6 | **Contenedores** | `backend/Dockerfile`: multi-stage, `CGO_ENABLED=0` estático, `-trimflags -s -w`, **distroless sin shell**, `USER nonroot:nonroot` (uid 65532) ✅ · `.dockerignore` de ambos excluye `.env`, docs, tests del contexto de build ✅ · `frontend/Dockerfile` → M1 (arriba) · `frontend/dist/` gitignored y sin trackear (`git ls-files` → vacío) | ✅ Backend correcto; frontend con nota M1 |
| 7 | **Dependencias** | `cd backend && govulncheck ./...` → **exit 3, GO-2026-5970** (→ B1) · `cd frontend && npm audit --audit-level=high` → **0 vulnerabilidades (exit 0)** | ⚠️ Solo B1; frontend limpio |
| 8 | **Control de acceso** | `backend/internal/status/routes.go:13-15` publica **solo** `GET /healthz` público; grep `RegisterAdmin\|authn\|authz` en `backend/` → solo comentarios de F2 (0 implementaciones); sin cookies ni grupos admin | ✅ Correcto frente a la spec: F1 declara el endpoint y la página **públicos** (Assumptions, US2) y difiere authn/authz a F2 (Out of Scope). **No existe ninguna superficie que requiera autorización** (única operación: estado del sistema, sin datos sensibles). D-A7 sigue pendiente de confirmación humana y no se ha adelantado nada — correcto. `security: []` documentado en el contrato |
| 9 | **Comportamiento en vivo (FR-003/FR-004)** | Durante la auditoría, con la BD detenida (escenario de validación de QA) `/healthz` → 503 en ~5 ms con sobre; al volver la BD → 200 | ✅ El estado es real por consulta y el backend sigue vivo con la BD caída |

---

## Registro de pendientes de seguridad (mismo mecanismo que R10/R11 → T034/T035; para que `documentador`/orquestador los lleven al README)

1. **B1** (bloqueante, inmediato): bump `golang.org/x/text@v0.39.0` → `make security` en verde.
2. **M1** (condición previa a despliegue): nginx no-root (imagen unprivileged o `USER nginx` + puerto 8080).
3. **M3/M4** (pre-despliegue): `server_tokens off` + cabeceras `nosniff`/`Referrer-Policy`.
4. **M2** (pre-despliegue): fijar imágenes por digest.
5. **M9** (local, opcional): puertos de compose en `127.0.0.1`.
6. Ya registrados por el plan (confirmados): `postgres:16.4-alpine` (R10/T034) y Node 22 (R11/T035).

---

*Auditoría de solo lectura: no se modificó código, `tasks.md` ni `estado.md`. Los contenedores se inspeccionaron con `docker compose ps/logs/exec` (lectura) mientras QA ejecutaba el escenario "BD caída" del quickstart; la cronología de logs confirma que los 503 observados eran ese escenario, no un fallo propio.*
