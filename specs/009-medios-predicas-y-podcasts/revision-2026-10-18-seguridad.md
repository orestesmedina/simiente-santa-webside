# Revisión de seguridad — F9 Medios (prédicas y podcasts)

- **Rama**: `009-medios-predicas-y-podcasts` · **Base**: `main` (`ee02ab1`, diff `main...HEAD`, 160 archivos, +38 852/−65, 70 commits)
- **Fecha**: 2026-10-18 · **Revisor**: `seguridad` (Fase 7, flujo `equipo-feature`)
- **Referencias**: `.specify/memory/constitution.md` §IV, `specs/009…/spec.md` (FR-004/FR-009/FR-015/FR-017/FR-019…FR-023, SC-002/SC-005/SC-009/SC-010/SC-014), `specs/009…/plan.md` (P9-7/P9-14/P9-15/P9-18/**P9-19/P9-20**, RG9-1/RG9-4, «Resolución de los hallazgos del analyze»: correcciones A1/A3/C1/M2), `specs/009…/data-model.md`, `specs/009…/analyze.md` (A1 resuelto con la fábrica única), OWASP Top 10 (2021).
- **Método**: lectura completa del backend de F9 —`internal/medios` íntegro (fábrica de red, feeds, disponibilidad, carátulas, service/handlers/routes, repositorio, migraciones `000014`/`000015`), `platform/{storage,audit}`, dominio `media` (servicio de archivos) y `internal/db/queries/media.sql`, cableado `cmd/api/main.go` y su humo—; frontend —`api/medios.ts`, `features/publico/medios/*`, `features/medios/*`, permisos, router, e2e (seeds y specs)—; greps de patrones peligrosos (SQL dinámico, `innerHTML`/`dangerouslySetInnerHTML`, `eval`, secretos, `href` dinámicos, `http.Client` fuera de la fábrica, `InsecureSkipVerify`); `govulncheck ./...` + `npm audit --audit-level=high` + `go vet` + `go test ./internal/medios/... ./cmd/api/...` (verde); **verificación en vivo contra el stack local** (públicos, cabeceras, 401/403, política de archivos con intentos de traversal, casos borde de paginación) y un **experimento local de Go** que confirma que `encoding/xml` rechaza entidades externas (XXE) y no conserva marcado anidado en campos de texto.

## Veredicto: **APRUEBA**

Sin hallazgos bloqueantes, altos ni medios. Tres bajos (B2 y B3 de defensa en profundidad/falso positivo; **B1** nuevo en superficie pública pero de robustez, sin confidencialidad ni integridad en juego) y seis informativos; ninguno bloquea el merge. **El punto más sensible de F9 —la primera red saliente del backend— está cerrada por construcción**: la fábrica única de `cliente_http.go` es el único punto de todo el backend que construye un `http.Client` (grep global: cero `http.Get`/`http.Post`/dialers fuera de ella), aplica las cuatro barreras a los tres consumidores con la misma tabla de pruebas, y conecta a la IP ya validada (sin ventana de rebinding entre validar y conectar).

---

## Hallazgos

### Bloqueantes

Ninguno.

### Altos

Ninguno.

### Medios

Ninguno.

### Bajos

**B1 — `offset` sin tope superior en la primera superficie pública paginada: un visitante sin sesión puede provocar un 500 (CWE-20 / OWASP A03:2021).**
- Dónde: `backend/internal/medios/service.go:238-258` (`normalizarPaginacionPublica` valida `limit < 0` y `offset < 0`, pero no acota `offset` hacia arriba — `limit` sí se acota a `maxPage=50`); `backend/internal/medios/repository_episodios.go:55-56` y `:102-103` (`int32(params.Offset)` trunca/wrappea el enter de 64 bits); espejo del panel: `backend/internal/platform/paginate/paginate.go:75-84` (tampoco acota `offset`).
- Qué: `GET /api/v1/medios?offset=2147483648` → wrap de `int32` → `OFFSET` negativo → error de Postgres → **500** (verificado en vivo). `?offset=9999999999999` responde 200 con `items: []` y `total` correcto; `?offset=-5` → 400 correcto.
- Impacto: **robustez/availabilidad menor, sin confidencialidad ni integridad**: el 500 es genérico («Error interno del servidor», sin causa ni SQLSTATE — verificado en vivo); sin inyección (todo parametrizado por sqlc); el visitante sin sesión solo fuerza un error y ruido de log. El patrón `int32(params.Offset)` es **heredado** (listados de panel de F4/F6/F7, tras sesión+permiso); F9 lo estrena en una ruta **pública**.
- Corrección propuesta (dev-backend, no bloqueante): acotar `offset` en `normalizarPaginacionPublica` como se acota `limit` (clamp a un tope absoluto o `MaxInt32`) o rechazarlo con 400 `details.offset`; de paso, el mismo clamp en `platform/paginate` para todos los listados.

**B2 — la allowlist de la red saliente fija el host pero no el puerto (CWE-918 parcialmente mitigado / OWASP A10:2021 — SSRF; defensa en profundidad).**
- Dónde: `backend/internal/medios/cliente_http.go:264` (`verificarURL` compara `u.Hostname()`, que descarta el puerto) y `:181-199` (`dialSeguro` valida el host de `net.SplitHostPort` y conecta con el puerto que traiga la petición).
- Qué: una URL de **carátula** del feed tipo `https://i.ytimg.com:8443/x.jpg` supera las barreras en lo que toca al host y conectaría al puerto 8443 de `i.ytimg.com`. Los **enlaces de episodio** no pueden llevar puerto (regex ancladas `feeds.go:66-67` + `CHECK` por proveedor en `migrations/000014:142-144`), el oEmbed lo construye el servidor con constantes (`disponibilidad.go:112-113`) y las URLs de los feeds son constantes (`feeds.go:50-51`): el único vector es la carátula del feed, y el destino sigue clavado a los hosts de la allowlist exacta.
- Impacto: residual y mínimo: exige que el contenido del feed esté bajo control del atacante (compromiso de las cuentas de YouTube/Anchor del cliente o de esos CDNs) y el alcance máximo es conectar a **puertos arbitrarios de 5 hosts fijos de terceros**. Nada de red interna ni servicio de metadatos: eso lo corta la barrera 2 aunque el «host» vaya como IP.
- Corrección propuesta (dev-backend, no bloqueante): exigir en `verificarURL` que `u.Port()` sea vacío o `443` (una línea); el dialer puede reconstruir `net.JoinHostPort(ip, "443")`.

**B3 — Contraseñas de e2e hard-coded (CWE-798, falso positivo probable; mismo patrón que B1 de F4/F6 y B1 de F7).**
- Dónde: `frontend/e2e/mediosSeed.ts:1001` (`'Inicial.2026'`), `frontend/e2e/medios-panel.spec.ts:578,624,642` (`loginNewAccount(…, 'Inicial.2026', 'Definitiva.2026')`).
- Impacto: ninguno en producción: cuentas creadas en caliente contra el stack local de `make up`, con cambio obligatorio de contraseña inmediato. Se reporta por completitud del grep de secretos; **no** hay secretos reales en el diff (`.env.example` sin cambios, sin `.env` versionado — verificado con `git ls-files`).
- Corrección sugerida (dev-frontend): ninguna obligatoria; opcionalmente centralizarlas en `e2e/helpers.ts` (deuda heredada de F4/F6/F7).

### Informativos (no requieren cambio para aprobar)

- **I1 — `govulncheck ./...`**: **0 vulnerabilidades que afecten al código** (1 informativa en módulo requerido que el código no llama — idéntica a la de F4–F7). `npm audit --audit-level=high`: **0**. Coincide con lo declarado en el cierre.
- **I2 — Dependencias exactas**: la única diferencia del diff es `golang.org/x/text v0.42.0` pasando de indirecta a directa (misma versión, ya presente en `go.sum`; la usa `titulo.go` para NFKD). `package.json`, `package-lock.json` y `.env.example` **sin cambios**; sin variables de entorno nuevas (P9-18: allowlists/topes/tiempos son constantes revisables del dominio — decisión documentada frente a §VII). El «0 dependencias nuevas» del informe de cierre es exacto.
- **I3 — `target_label` de auditoría transporta el título del episodio** (dato del feed, recortado a 254 runas en `etiquetaContenido`): visible solo para cuentas con permiso de auditoría; es un dato destinado a publicarse. Patrón heredado de F4/F6/F7. El resumen de la búsqueda solo lleva conteos y estado por feed (tests que grepean el JSON crudo buscando internos: 0).
- **I4 — Superficie de panel sin rate-limit (heredado de F2–F7)**: `POST /buscar-nuevos` sostiene una petición hasta 90 s, con todo acotado por constantes del dominio (30 elementos/30 comprobaciones/30 carátulas por búsqueda, 4 en vuelo, 10/5 s por petición saliente, cuerpos ≤2 MiB/64 KiB/8 MiB). Exige sesión + permiso `medios`; conviene vigilar el volumen en producción.
- **I5 — `Truncado` conservador**: un cuerpo exactamente del tamaño del tope se marca como truncado (`cliente_http.go:303`) y feeds/carátulas lo rechazan. A propósito (mejor rechazar que aceptar a ciegas); sin impacto.
- **I6 — RG9-1 documentado** (el feed de YouTube responde 404 desde direcciones de servidor, `estado.md`): es un riesgo de **disponibilidad de la búsqueda**, no de seguridad, y ya está documentado; la regla a prueba de fallos (FR-016/FR-017) garantiza que un feed caído no oculta ni marca nada — se cumple por diseño y por pruebas.

---

## Verificación contra la constitución §IV y OWASP

**A10 — SSRF / red saliente (CWE-918). ✅ (residual B2, defensa en profundidad)**
- **Fábrica única real**: `cliente_http.go` es el **único** punto de todo el backend que construye un `http.Client` (grep global: ninguna otra aparición en código de producción, cero `http.Get`/`http.Post`/dialers/resolvers fuera de ella); la única puerta exportada es `NewRedSaliente` (`red_saliente.go:41-55`), que construye los tres consumidores, y `cmd/api/main.go:197-210` solo puede cablear por ahí. Un consumidor nuevo no puede montarse sin barreras (los constructores fallan sin allowlist/tope/timeout, y `NewRedSaliente` es la única vía exportada).
- **Barrera 1 (https + allowlist por consumidor)**: feeds `www.youtube.com`/`anchor.fm` (2 MiB, 10 s); probador `www.youtube.com`/`podcasters.spotify.com` (64 KiB, 5 s); carátulas `i.ytimg.com`/`d3t3ozftmdmh3i.cloudfront.net` (tope `UPLOAD_MAX_BYTES` = 8 MiB, 10 s). Allowlist **exacta por host** (mapa, sin comodines) y **por consumidor** (test: el host de otro consumidor se rechaza). Las URLs de los feeds **no son configurables**: constantes del dominio.
- **Barrera 2 (IPs privadas tras DNS)**: `dialSeguro` resuelve con `LookupNetIP`, exige que **todas** las direcciones sean públicas (`ipPublica`: unicast global y ni loopback/privada/link-local/multicast/unspecified; `Unmap()` contra IPv4-mapped; tabla de pruebas con `127.0.0.1`, `10.x`, `192.168.x`, `172.16.x`, **`169.254.169.254`**, `::1`, `fd00::1`, `fe80::1`, `0.0.0.0`) y **conecta a la IP ya validada, no al nombre** — sin ventana TOCTOU/DNS-rebinding entre validar y conectar. Se aplica en cada conexión, también tras redirects.
- **Barrera 3 (redirects)**: `CheckRedirect` valida cada `Location` contra la allowlist del consumidor (https + host) con tope de 10 saltos; el redirect además vuelve a pasar por el dialer (barreras 1 y 2 re-aplicadas en la conexión — test directo). `Proxy: nil`: un proxy del entorno no puede saltarse las barreras.
- **Barrera 4 (cuerpo)**: `io.LimitReader` en **todo** cuerpo (el test usa un `cuerpoInfinito` que **falla** si alguien intenta leer más allá del tope); feeds y carátulas rechazan el cuerpo truncado (nunca se parsea ni se guarda un archivo cortado).
- Sin credenciales salientes: solo `User-Agent` propio fijo; ni cookies ni `Authorization` viajan jamás a internet. Sin `InsecureSkipVerify` ni TLS debilitado (grep global: 0).

**A03 — Contenido no confiable (CWE-79/CWE-91/CWE-89). ✅**
- **XXE imposible**: parseo con `encoding/xml` estándar (0 dependencias nuevas). **Verificado con un experimento local**: el parser de Go **rechaza** entidades externas (`<!ENTITY xxe SYSTEM …>` → «invalid character entity») y no resuelve DTDs externos.
- **Descripciones = texto plano en todas las capas**: entidades decodificadas y etiquetas fuera en el parseo (fixtures real y patológico; probado que `strings.ContainsAny(desc, "<>")` no da); la descripción completa del feed **no** se publica (solo el recorte ≤300 con corte de frase, `descripcionCorta`); se guarda como dato y se pinta como **texto de React**.
- **XSS**: **0** `dangerouslySetInnerHTML`/`innerHTML`/`document.write`/`eval` en todo el frontend (grep; las 12 apariciones de `outerHTML` son **lecturas** de los e2e para grepear el HTML renderizado). Enlaces externos con `target="_blank" rel="noopener noreferrer"` (`PublicButton.tsx:77`) y anunciados (FR-005).
- **Enlaces como datos validados**: los `href` de ficha y panel (`MedioLinks.tsx:38,52`, `FuentesList.tsx:33`) vienen del servidor, que valida la forma por proveedor al importar (`feeds.go:66-67`: `^https://(www\.)?youtube\.com/watch\?v=[A-Za-z0-9_-]{11}$` **anclado con `$`** y `^https://podcasters\.spotify\.com/`) y la BD revalida con `CHECK` por proveedor (`migrations/000014:142-144`; test de integración con `javascript:alert(1)` rechazado). **Imposible `javascript:` en un href.**
- **SC-005 (0 terceros en el navegador)**: e2e que intercepta todas las peticiones del contexto y grepea el DOM buscando `youtube|ytimg|spotify|anchor|cloudfront` → 0; carátulas servidas desde `/api/v1/media/<file>` (copias locales). Se cumple el riesgo asumido y documentado (el visitante sale del sitio; nada se incrusta).
- **SQL (CWE-89)**: todo por sqlc parametrizado; grep de SQL construido por concatenación/`Sprintf` en producción: **0** (solo tests de integración con identificadores fijos).

**A04/A05 — Descarga y servicio de archivos (CWE-434/CWE-22). ✅**
- **Descarga de carátulas**: por la fábrica única (allowlist de 2 CDNs, https, sin IPs privadas, redirects validados, cuerpo acotado a `UPLOAD_MAX_BYTES`, rechazo de truncados). **No hay endpoint de subida en F9**: la única entrada de imágenes es la descarga del feed.
- **Guardado** (`platform/storage`): valida **firma binaria** (`http.DetectContentType`: solo JPEG/PNG/WebP — SVG, GIF y HTML renombrado se rechazan), tope de tamaño y **nombre generado por el servidor** (`img_<uuid>.<ext>`); escritura atómica (temp+rename).
- **Servicio** (`GET /api/v1/media/{fileName}`, dominio `media`): patrón **exacto** de nombre antes de tocar el disco (`^img_[0-9a-f-]{36}\.(jpg|png|webp)$`) → 400 (verificado en vivo con `../` codificado y con `img_index.html`); predicado **publicable** — la rama F9 de `IsFilePublished` usa el **mismo predicado público** de la sección (visible + no absorbida + sin par abierto; `missing_from_feed` sin resolver **sí** se sirve, como exige FR-017) → 404 para la carátula de una ficha oculta (test de integración + e2e); cabeceras `X-Content-Type-Options: nosniff`, `Content-Disposition: inline`, `Cache-Control: no-store` y `Content-Type` solo del mapa jpg/png/webp (nunca `text/html`). **Path traversal imposible por diseño** (nombre generado + patrón estricto + `filepath.Join` sobre nombre ya validado).

**A01 — Autorización y sesión (CWE-862/CWE-352). ✅**
- Los **5 endpoints de panel** cuelgan de `root.Group("/api/v1/admin/medios", middleware.AdminChain("medios")…)` (`routes.go:63`): authn → PasswordGuard → AuthzByModule("medios") → CSRF, **en el servidor**, por petición. Humo de `cmd/api` (verde): **401** en las 5 sin sesión; **403** genérico desde `eventos|actividades|categorias|ministerios|donaciones` con fila `media.*` `result='denied'`, y cruce inverso (solo `medios` → 403 en los demás). Los e2e fuerzan las 7 operaciones sin permiso → 403 todas. **Sin permisos nuevos** (catálogo en 10, permiso `medios` sembrado por F2; el frontend solo lo activa en `MODULE_PERMISSIONS`, cosmético).
- Los **2 públicos** son de solo lectura, sin middlewares de sesión, `Cache-Control: no-store` y `lang` cerrado es|en (verificados en vivo: `lang=fr` → 400). El tope 5 del bloque es **firme del servidor** (la ruta no admite `limit`; verificado en vivo: `?limit=100` ignorado).
- **CSRF**: mutaciones por el double-submit firmado de F2 (cadena) y el cliente manda `X-CSRF-Token` en todo no-GET (`client.ts:152-155`); verificado en vivo (POST sin sesión → 401). Sesión/cookies/bcrypt de F2 sin cambios.
- **Sin IDOR públicos**: no existe ruta pública por id; los `{id}` del panel están tras el permiso, por módulo (mismo diseño que F3–F7).
- **A09 (CWE-209) — errores**: `WriteError` no envía la causa de un 500 (verificado en vivo); el 403 no revela el permiso que falta; los 400 llevan `details` por campo; el resumen de la búsqueda son conteos y estado por feed (tests que grepean el JSON crudo: 0 internos).

**A02/A06 — Secretos y dependencias (CWE-798). ✅** Grep de `password|secret|token|api_key` sobre el diff: solo B3, nombres de variables y `CSRFSecret:"secret"` en tests. `.env.example` sin cambios; sin `.env` versionado. `govulncheck`/`npm audit`: ver I1/I2. Las URLs de los feeds son **públicas y sin claves** (documentado y verificado: constantes, sin configuración).

**A09 — Auditoría y datos personales. ✅**
- **6 códigos `media.*`** (sync, description.update, hide, show, merge, mark.resolve) con la regla exacta de P9-14: toda mutación del panel inserta su(s) fila(s) **en la misma transacción** (tests de integración de rollback); el `CHECK` cerrado de `admin_actions` revalida en BD (000015, con `up`/`down` completos: el `down` borra las filas F9 y restaura el CHECK de F7). El alta importada por la búsqueda se registra **solo** con `media.sync` + `media_sync_runs` (decisión documentada).
- **Denegaciones** (`result='denied'`, Recorder = el service de F9 en la cadena) y **rechazos** de validación (`result='failure'` best-effort) **sin el cuerpo de la petición** (`Rejection` lleva actor+método+ruta).
- **Datos personales**: el parser **no lee autores** del Atom (el struct no tiene campo `author`); lo guardado de los feeds es título, id externo, enlace, fecha, carátula y descripción — datos públicos del feed. `media_sync_runs` guarda el actor (uuid) tras el permiso del panel. I3 como constancia.

**Riesgos asumidos y documentados — verificados, no contados como hallazgo**: (a) el visitante **sale del sitio** y no se incrusta nada de terceros (FR-004/SC-005: cumplido y probado); (b) RG9-1 (feed de YouTube 404 desde IPs de servidor): documentado en `estado.md`; la regla a prueba de fallos cumple; (c) allowlist/topes/tiempos como **constantes revisables** del dominio frente a §VII: decisión documentada en el analyze/plan, sin variables de entorno nuevas.

---

## Recomendaciones (deuda, no bloqueantes)

1. **B1 (dev-backend)**: acotar `offset` en `normalizarPaginacionPublica` (clamp o 400) y, de paso, en `platform/paginate` para todos los listados (el mismo wrap existe tras sesión en F4/F6/F7).
2. **B2 (dev-backend)**: exigir puerto vacío o `443` en `verificarURL` (una línea; el dialer ya valida el host).
3. **B3 (dev-frontend)**: opcionalmente centralizar las contraseñas de e2e en `e2e/helpers.ts` (deuda heredada).
4. Vigilar el volumen de `admin_actions` y de `buscar-nuevos` en producción (I4); si crece, considerar rate-limit o agregación de denegaciones repetidas.
5. Cuando exista despliegue con TLS: añadir CSP (y HSTS) al frontend público (recomendación heredada de F4–F7, pendiente a nivel de proyecto).

## Pruebas ejecutadas por el revisor

- `govulncheck ./...` (backend): **0 alcanzables** (1 informativa no llamada, I1). · `npm audit --audit-level=high` (frontend): **0**. · `go vet ./internal/medios/... ./cmd/api/...`: ok. · `go test ./internal/medios/... ./cmd/api/...`: **ok** (incluye el humo de composición —401 en las 5 rutas de panel sin sesión, públicas 200 + `no-store`, `lang` inválido 400, 404 en rutas no registradas— y el de permisos cruzados —403 con fila `denied` en ambas direcciones—).
- **En vivo contra el stack local** (solo lecturas y un POST sin sesión): públicos 200 con `no-store`; `lang=fr` → 400; `POST /buscar-nuevos` sin sesión → 401; `GET /api/v1/media/..%2f..%2fetc%2fpasswd` → 400; `img_<uuid>.png` inexistente → 404; `img_index.html` → 400; `?limit=100` en el bloque de portada ignorado (tope 5 del servidor); `?offset=2147483648` → **500 genérico** y `?offset=-5` → 400 (hallazgo B1, verificado).
- Greps: SQL dinámico en producción (**0**), `dangerouslySetInnerHTML`/`innerHTML`/`eval`/`document.write` (**0** en src; las 12 de e2e son lecturas `outerHTML`), secretos (solo B3), `http.Client`/dialers/resolvers fuera de la fábrica (**0**), `InsecureSkipVerify` (**0**), `href` dinámicos sin validar (solo los validados por el servidor).
- **Experimento local de Go** (fuera del repo): `encoding/xml` rechaza `<!ENTITY … SYSTEM>` (XXE imposible) y no conserva marcado anidado en campos de texto.
- Las pruebas de integración (`-tags=integration` + PostgreSQL) y los e2e completos de F9 **no fueron reejecutados por el revisor** (qa-tester los corre en paralelo sobre el mismo stack; `estado.md` los reporta en verde; su cobertura crítica está citada por test concreto arriba).

## Nota de alcance

- El revisor **no modificó código**: el único archivo escrito por el revisor es este informe.
- No se disparó `POST /buscar-nuevos` con sesión desde la auditoría (habría escrito en la base y salido a internet mutando el estado del stack compartido que el resto del equipo está usando): las barreras de red están fijadas por la tabla compartida de pruebas sobre los tres consumidores con dobles, y `estado.md` documenta la verificación manual contra los feeds reales con el mismo código cableado en `cmd/api` (10 episodios importados, 8 unidos, bloque de portada con exactamente 5).
- No se pudo comprobar el comportamiento del feed de YouTube desde la IP de este servidor (RG9-1, documentado): la ruta a prueba de fallos está cubierta por diseño y por pruebas.
