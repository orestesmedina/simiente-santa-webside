# Revisión de seguridad — F3 Portada e información general

**Fecha**: 2026-10-10 · **Rama**: `003-portada-info-general` · **Auditor**: `seguridad` (solo
lectura; este reporte es la única escritura). **Alcance**: `git diff main...HEAD` del código de
F3 (`backend/internal/portada/*`, `platform/storage`, `platform/validate`, `platform/audit`,
`platform/middleware`, migraciones `000005`/`000006`, `backend/api/openapi.yaml` 0.4.0,
`frontend/src/features/publico|informacion`, `frontend/src/api/portada.ts`, `docker-compose.yml`,
`.env.example`), contrastado con la constitución §IV, `spec.md` (FR-012/FR-013/FR-015/FR-017),
`plan.md` (RG3-6…RG3-9), `data-model.md`, `research.md` (R3-8/R3-11) y el OWASP Top 10 2021.

**Método**: auditoría **estática** (lectura de código, greps, consultas sqlc, migraciones) +
**dinámica** (el stack local estaba operativo: `docker ps` con backend/frontend/db/redis sanos,
`/healthz` 200; se ejercitó la API real con `curl` sobre http://localhost:8080). No se modificó
ningún archivo del repositorio; las mutaciones de prueba se listan en §7.

## 1. Verificaciones ejecutadas

| Comando | Resultado |
|---|---|
| `make security` → `cd backend && govulncheck ./...` | **0 vulnerabilidades que afecten al código** (1 en módulos requeridos no llamada por el código — fuera del criterio §IV) ✓ |
| `make security` → `cd frontend && npm audit --audit-level=high` | **0 vulnerabilidades** ✓ |
| `go test ./internal/portada/... ./internal/platform/storage/... ./internal/platform/validate/... ./internal/platform/audit/... ./cmd/api/...` | todo `ok` ✓ |
| `grep dangerouslySetInnerHTML\|innerHTML frontend/src` | **0 coincidencias** ✓ |
| `grep` SQL dinámico (`Sprintf`+`SELECT/INSERT/UPDATE/DELETE`) en `backend/` | **0 coincidencias** ✓ |
| `git ls-files \| grep -E '^\.env$|\.pem$|\.key$|secrets/'` | **0 secretos versionados** ✓ |
| `grep (password\|secret\|api_key\|token)\s*[:=]\s*"..."` en `backend/ frontend/ docker-compose.yml` | **0 secretos reales** (solo placeholders documentados) ✓ |
| `grep -rniE "(UPDATE\|DELETE) admin_actions" backend/internal/db/queries/` | **0** — el registro de auditoría es de solo lectura en SQL ✓ |
| Pruebas dinámicas (curl, §2) | matriz de acceso, CSRF, media, subidas, borradores, validación de URLs ✓ |

## 2. Pruebas dinámicas (contra el stack real, 2026-10-10)

| # | Prueba | Resultado |
|---|---|---|
| D1 | Las 14 operaciones del panel sin sesión (`GET/PUT/PATCH/POST/DELETE /api/v1/admin/portada*`) | **401 `unauthenticated`** en las 14 ✓ |
| D2 | Login ADMIN → cookies | `ss_session` **HttpOnly; SameSite=Lax**; `csrf_token` firmada (`valor.firma`), no HttpOnly por diseño del double-submit ✓ |
| D3 | `PUT /quienes-somos` con sesión pero **sin** `X-CSRF-Token` y con **token falsificado** | **403 `forbidden`** («El token de seguridad no es válido…») en ambos ✓ |
| D4 | `PUT /quienes-somos` con CSRF válido (re-escritura idéntica) | **200** + fila `home.about.update` `result=success` en auditoría + contenido público idéntico ✓ |
| D5 | Cuenta **sin permiso `portada`** (rol con solo `eventos`), tras resolver el cambio obligatorio de contraseña: las 14 operaciones | **403 genérico** «No tienes permiso para acceder a este módulo» en las **14/14**; sin detalles internos ✓ |
| D6 | Denegaciones de D5 en `/api/v1/admin/auditoria/acciones` | **13 filas `result=denied`** (las 13 mutaciones; la lectura del agregado no tiene acción asociada por diseño R3-11) ✓ |
| D7 | `GET /api/v1/portada` (es/en) vs. estado del panel | 5 elementos en borrador (horario): **0 IDs y 0 textos de borrador** en la respuesta pública; **0 apariciones de `publicationState`** en el DTO público; recuentos publicados = públicos (16/16, 16/16, 1/1) ✓ |
| D8 | `GET /api/v1/media/{archivo publicado}` | **200** con `X-Content-Type-Options: nosniff`, `Content-Disposition: inline`, `Cache-Control: no-store`, `Content-Type: image/jpeg` (JPEG real verificado) ✓ |
| D9 | `GET /api/v1/media/{nombre inválido}` (`foo.txt`, `img_no.txt`, `img_00000000-….exe`) | **400 `invalid`** ✓ |
| D10 | Path traversal en media: `..%2f..%2fetc%2fpasswd`, `img_..%2Fsecret.jpg`, y ruta cruda `../../../../etc/passwd` | **400 `invalid`** (patrón) y **404** (el router no resuelve el crudo); **nunca se sirve el archivo** ✓ |
| D11 | `GET /api/v1/media/{nombre válido inexistente}` | **404 `not_found`** ✓ |
| D12 | Subida (como ADMIN) de **SVG con `<script>`**, de **HTML renombrado .jpg** (Content-Type `image/jpeg` mentiroso) | **400** «solo se admiten imágenes JPEG, PNG o WebP» — la firma binaria manda sobre el Content-Type del cliente ✓ |
| D13 | Subida de **9 MB** | **400** «La imagen es demasiado grande» (tope 8 MB, `MaxBytesReader` en stream) ✓ |
| D14 | Subida de JPEG válido | **201** con nombre generado por el servidor `img_490ccc33-….jpg` ✓ |
| D15 | Descarga del archivo recién subido (no referenciado por contenido publicado) | **404** «La imagen no está disponible» — no hay vía pública a archivos no publicados ✓ |
| D16 | Validación de URLs hostiles (panel): `javascript:alert(1)`, `http://…`, `https://evil.com`, red fuera de catálogo, grupo de WhatsApp en host ajeno, teléfono inválido | **400** con `details` por campo (`url`, `network`, `destination`) + filas `result=failure` en auditoría ✓ |
| D17 | `GET /api/v1/portada?lang=<script>` | **400** «solo se admite es o en» ✓ |
| D18 | Cobertura dinámica previa (e2e de F3, corridos en verde el 2026-10-09 en este entorno, `cierre.md` §1.2): identidad en borrador → media **404**; republicada → media **200** (`portada-publica.spec.ts:331-353`) | ✓ |

## 3. Análisis por área (OWASP 2021 + constitución §IV)

**A01 — Control de acceso (CWE-862).** El grupo `/api/v1/admin/portada` monta
`middleware.AdminChain("portada", …)` (`routes.go:49`, `chain.go:51-57`):
`authn → guard de contraseña → authz(portada) → CSRF`, igual que F2 pero con el módulo de F3;
el permiso `portada` es del catálogo cerrado de F2 (migración `000002`, sin cambios). Verificado
en servidor: 401 sin sesión y 403 sin permiso en **las 14 operaciones** (D1/D5); el mensaje 403
es genérico por diseño (`authz.go:18`). **Borradores**: la única vía pública son las consultas
`…Published` (`home.sql`: `WHERE publication_state='published'`), separadas de las del panel
(P3-4); el DTO público no tiene `publicationState` ni pares `Es/En` crudos (D7); la descarga de
imágenes comprueba `IsHomeFilePublished` antes de abrir el archivo (`service_public.go:176-201`,
`home.sql:288-298`) — archivo de borrador/no referenciado → 404 (D15 + e2e D18). IDOR: los `id`
son UUID y toda mutación exige el permiso del módulo (no hay operaciones por objeto sin
autorización); los handlers re-verifican identidad del contexto como red de seguridad
(`handler.go:189-196`).

**A03 — Inyección SQL (CWE-89).** Toda consulta vive en `queries/home.sql` con parámetros
`sqlc.arg` (código generado, commiteado); el repository solo llama a `gendb.Queries`
(`repository.go`, `repository_home.go`); 0 construcción dinámica de SQL (grep). Restricciones
`CHECK`/`UNIQUE` refuerzan en la base lo que valida el service.

**A03 — XSS (CWE-79).** 0 `dangerouslySetInnerHTML`/`innerHTML` en `frontend/src` (grep); React
escapa todo el contenido (los textos viajan como datos: FR-003/FR-015). Las imágenes se suben
con **firma binaria** (JPEG/PNG/WebP; SVG y GIF rechazados — D12) y se sirven con `nosniff`,
`inline` y Content-Type deducido de la extensión de un nombre generado por el servidor
(`handler_media.go:38-41`, `storage.go:105-116`); el nombre se valida con la regex
`^img_[0-9a-f-]{36}\.(jpg|png|webp)$` antes de tocar el disco (`storage.go:52`, `local.go:64`).
Los enlaces externos (`wa.me`, redes) se validan `https` + host oficial en el backend (D16) y el
frontend los abre con `rel="noopener noreferrer"` (`WhatsAppSection.tsx:49-53`,
`SocialSection.tsx:29-33`, `PublicFooter.tsx:36-39`). El idioma `lang` se valida contra
`es|en` en el servidor (D17). Se verificó además que un elemento de prueba con nombre
`<script>alert(1)</script>` (creado por una validación anterior) se trató como dato: hoy no
queda rastro en la base y React lo habría renderizado escapado.

**A04/A05 — Subida de archivos y configuración (CWE-434/CWE-400/CWE-22).** Superficie de subida
solo tras `authn+authz(portada)+CSRF` (D1/D5); tope `UPLOAD_MAX_BYTES` (8 MB) aplicado **en
stream** con `http.MaxBytesReader` + margen multipart (`handler_images.go:39-69`) y re-comprobado
exacto; tipos cerrados por firma (D12); tamaño (D13); **nombre generado por el servidor**
`img_<uuid>.<ext>` (D14) — el nombre del cliente jamás toca el disco; escritura atómica
(temp+rename, `local.go:98-117`); path traversal imposible: la regex se aplica **antes** de
componer la ruta en `Open`/`Delete` y de nuevo en la política pública (D9/D10);
`Content-Type` spoofing inútil (firma binaria, D12). Auditoría de subida **fail-closed**: si el
registro `home.image.upload` falla, el archivo se borra y la subida falla
(`service_admin.go:351-356`, prueba `TestUploadImageHandlerAuditFailureIsClosed`). Config:
`UPLOAD_DIR`/`UPLOAD_MAX_BYTES` desde entorno con validación de arranque
(`config.go:181-188`); `SESSION_SECRET` obligatoria ≥32 en producción (error de arranque) y
aviso en desarrollo; `SESSION_COOKIE_SECURE=true` forzado en producción (`config.go:192-197`);
`BOOTSTRAP_TOKEN` vacío deshabilita la inicialización (aviso, `config.go:205-210`).

**A02/A05 — Secretos (CWE-798).** Ningún secreto versionado (`.env` ignorado por git y con
placeholders de desarrollo; `.env.example` documenta sin valores reales; 0 claves/tokens en
código y en compose — los defaults de desarrollo están marcados como tales y la producción los
rechaza). La sesión viaja en cookie `HttpOnly` (no en `localStorage`; el frontend solo usa
`localStorage['ss.lang']` para el idioma, D2 + `LanguageProvider.tsx`).

**Autenticación/sesión/CSRF (CWE-614/CWE-352).** Cookie `ss_session`: `HttpOnly`, `SameSite=Lax`,
`Secure` según entorno, TTL de inactividad 30 min + absoluta 60 min con comprobación defensiva
de `ExpiredAt` en `authn` (`authn.go:59-62`). CSRF double-submit **firmado con HMAC**
(`csrf.go:34-52`): sin cabecera → 403; token falsificado → 403 (D3); el cliente lee la cookie
firmada y la replica en `X-CSRF-Token` (`client.ts`). El `PasswordGuard` bloquea el panel a
cuentas con cambio pendiente **antes** de authz (verificado: el 403 con
`reason=password_change_required` hasta resolverlo). CORS: echo exacto del Origin contra
allowlist, `Allow-Credentials` sin `*`, `Vary: Origin` (`cors.go:46-73`).

**A06 — Dependencias (CWE-1104).** `govulncheck`: 0 vulnerabilidades que afecten al código (1
en módulos requeridos no llamada — informativa); `npm audit --audit-level=high`: 0 (§IV ✓).
0 dependencias nuevas de runtime en backend; 3 `@fontsource` (assets de build) en frontend,
justificadas en el plan (P3-19/RG3-12).

**Auditoría (FR-017/SC-013).** Cada mutación inserta `admin_actions` **dentro de la misma
transacción** (`repository.go:41-59`, `withTx` + `InsertAdminAction` compartida); si el registro
falla, la mutación no se aplica (prueba de integración
`TestIntegrationContentMutationAndAuditAreAtomic` + `TestSaveIdentityRollsBackOnAuditFailure`).
Registro cerrado por la BD (CHECK de `000006`) y por `platform/audit` (15 códigos `home.*`);
**sin `UPDATE`/`DELETE` sobre `admin_actions`** en las consultas (grep) y las rutas del módulo
de auditoría son GET-only (`handler_audit.go:100-106`). Denegaciones y rechazos best-effort
verificados dinámicamente (D6: 13 filas `denied`; D16: filas `failure`).

**Higiene de errores (CWE-209).** `WriteError` es el único punto de traducción: los 500 llevan
mensaje genérico al cliente y el detalle interno solo al log con `request_id`
(`httpserver/error.go:85-150`); los 401/403 son uniformes y genéricos (D1/D5). El logging no
registra cuerpos ni tokens (`logging.go`, `authn.go:85-90`).

## 4. Hallazgos

### BLOQUEANTE

**Ninguno.**

### IMPORTANTE

**Ninguno.**

### MENOR

**S1 — `logoFile`/`coverImageFile` se aceptan sin el patrón de nombre generado por el servidor
(CWE-20, validación de entrada incompleta en una vía de panel).**
- **Dónde**: `backend/internal/portada/service_admin.go:906-924` (`identityImageRules` solo
  comprueba longitud ≤ 120 y presencia de `logoAltEs`/`coverImageAltEs`) y
  `model.go:150-153` (`validate:"omitempty,max=120"`).
- **Riesgo**: una cuenta **con permiso** puede guardar como referencia de imagen cualquier
  cadena ≤120 (p. ej. `../x.jpg` o una nunca subida). No hay path traversal (la descarga valida
  el patrón antes de tocar disco, D9/D10) ni exposición (un archivo no referenciado por
  contenido publicado responde 404, D15): el único efecto es una URL pública rota (400/404) y
  una referencia huérfana. Es higiene de datos dentro del perímetro autorizado.
- **Recomendación**: en `identityImageRules`, exigir `storage.ValidName(...) == true` cuando el
  campo venga relleno (y devolver `details.logoFile`/`details.coverImageFile` si no). Prueba
  unitaria con `../`, nombre de cliente y nombre inexistente.

**S2 — Las subidas rechazadas no dejan fila `result='failure'` en la auditoría (CWE-778,
traza incompleta; coincide con M1 de `revision-2026-10-10-codigo.md`).**
- **Dónde**: `service_admin.go:322-349` (`UploadImage`): archivo vacío, firma no permitida y
  tamaño excedido devuelven 400 sin `recordContentFailure`, mientras el resto de mutaciones del
  módulo sí registra su rechazo (verificado dinámicamente: mis rechazos de URL de D16 dejaron
  filas `failure`, los de subida de D12/D13 no).
- **Riesgo**: pierde trazabilidad de intentos abusivos de subida (la superficie está protegida
  por authz+CSRF, así que es observabilidad, no bypass). No incumple SC-013 (que trata de
  ediciones aplicadas), pero rompe la uniformidad del patrón de auditoría de F3.
- **Recomendación**: `recordContentFailure(ctx, actorID, audit.ActionHomeImageUpload, …)` en los
  tres retornos de 400, o documentar la excepción en R3-11 si se decide a propósito.

**S3 — La semántica `null` de los PATCH no distingue «campo ausente» de «vaciar campo» (lente de
seguridad de B1 de `revision-2026-10-10-codigo.md`).**
- **Dónde**: `model.go:152-182` (punteros `*string` sin presencia real), `service_admin.go:732-777/822-825`
  (`merge*` solo actúan con `!= nil`); detalle completo en el reporte de código (B1).
- **Riesgo (seguridad)**: la intención explícita y autorizada de **retirar** un dato opcional
  (quitar la hora de fin, vaciar una traducción al inglés) se ignora en silencio con un
  «Guardado» y el valor previo permanece. Los campos afectados son contenido público en ambos
  idiomas y horas de un servicio: **no cruza ningún perímetro de confidencialidad ni
  autorización** (quien escribe ya tiene el permiso, y el dato viejo ya era público en su
  versión base). Es un defecto de integridad de la intención del usuario dentro del perímetro,
  no una vulnerabilidad explotable.
- **Recomendación**: la de B1 (presencia real en el backend o `""` como valor de limpieza en el
  contrato + frontend). **El merge sigue bloqueado por B1 desde la revisión de código**; este
  hallazgo no añade un bloqueo de seguridad, pero debe resolverse en el mismo PR.

**S4 — `X-Content-Type-Options: nosniff` solo en la descarga de imágenes (CWE-693, hardening).**
- **Dónde**: `handler_media.go:38` la fija; el resto de respuestas del backend (JSON con
  `Content-Type: application/json` explícito) y el nginx de la SPA (`frontend/nginx.conf`) no la
  envían.
- **Riesgo**: muy bajo: las respuestas JSON declaran su tipo y la SPA sirve estáticos propios.
  Es profundidad de defensa estándar.
- **Recomendación**: añadir `nosniff` (y opcionalmente una CSP mínima para la SPA) en el punto
  único de salida (`httpserver`) y en nginx cuando se toque infra; no urgente para F3.

**S5 — Archivos huérfanos acumulables en `uploads_data` (limpieza incompleta; coincide con M3
de `revision-2026-10-10-codigo.md` y RG3-1 del plan).**
- **Dónde**: `service_admin.go:1022-1033` (`deleteReplacedImage` best-effort, leído fuera de la
  transacción); huérfanos de subidas sin guardar (como el de mi prueba D14).
- **Riesgo**: acumulación de disco; **sin exposición** (verificado: un huérfano responde 404
  en la vía pública, D15). Diseño aceptado (R3-8 best-effort).
- **Recomendación**: rutina de limpieza de huérfanos en una funcionalidad futura (F4+),
  ya anotada en el plan.

## 5. Cobertura de los requisitos de seguridad de la spec

| Requisito | Estado |
|---|---|
| FR-012 — permiso `portada` verificado en servidor en toda operación | ✓ (D1/D5: 14/14; SC-005) |
| FR-013/SC-002 — borradores nunca expuestos por ninguna vía | ✓ (D7 consultas `…Published`; D15 `IsHomeFilePublished` → 404; D18 e2e borrador→404) |
| FR-015 — validación de toda entrada en backend | ✓ (etiquetas + dominio: URLs/correo/teléfono/catálogo/duplicados/límites; D16/D17; S1 es la única brecha, menor y de panel) |
| FR-017/SC-013 — auditoría transaccional, fail-closed, solo lectura | ✓ (misma transacción, rollback si falla, sin UPDATE/DELETE, GET-only; D4/D6/D16) |
| RG3-6 — contenido como vector XSS | ✓ mitigado (texto plano + React + firma binaria sin SVG + nosniff) |
| RG3-7 — subidas abusivas | ✓ (8 MB en stream, tipos cerrados, superficie solo autorizada) |
| RG3-8 — borrador filtrado por vía nueva (archivos) | ✓ cerrado (analyze C4) |
| §IV constitución — SQLi/XSS/secretos/authz/dependencias | ✓ (§1, §3) |

## 6. Pendientes (ninguno por el entorno; notas de despliegue)

La auditoría dinámica pudo ejecutarse completa (stack operativo); **no quedan pruebas
pendientes por el entorno**. Como referencia para el despliegue futuro (fuera del alcance de
F3, sin producción en esta fase): `SESSION_COOKIE_SECURE=true` con TLS (ya forzado por config
si `APP_ENV=production`), TLS también para PostgreSQL, Redis con autenticación, y el
rate-limit tras proxy inverso necesita un `ClientIP` de confianza (limitación documentada en
`ratelimit.go:66-70`).

## 7. Residuos de las pruebas dinámicas (datos añadidos a la base compartida)

- Rol **«Auditoria seguridad sin portada»** (permiso `eventos`) y cuenta
  **`sara-auditoria-seguridad@ejemplo.com`** (contraseña `Nueva.Aud.2026`), sin permiso
  `portada` — mismo patrón que dejan los e2e.
- Una imagen huérfana válida en el volumen: **`img_490ccc33-5a2d-4785-b70e-3f1edeb49528.jpg`**
  (subida D14; **no descargable**: 404 al no estar referenciada, D15).
- Filas de auditoría resultantes: 13 `denied` (matriz D5), varias `failure` (D16),
  1 `home.about.update` success (D4, re-escritura idéntica) y 1 `home.image.upload` success (D14).
- El contenido público quedó **idéntico** al inicio de la auditoría (verificado tras D4/D16).

## 8. Veredicto

**APROBADO** — 0 hallazgos BLOQUEANTES y 0 IMPORTANTE de seguridad; 5 MENORES (S1–S5), dos de
ellos ya registrados por `revision-2026-10-10-codigo.md` (S2≡M1, S5≡M3) y S3 es la lectura de
seguridad de su B1.

**Nota para el orquestador**: este APROBADO es el de la dimensión de seguridad (CWE/OWASP
§IV). El merge de la rama sigue **bloqueado** por el veredicto RECHAZADO de la revisión de
código (B1, `revision-2026-10-10-codigo.md`), que este reporte no revoce: B1 debe corregirse
antes del PR; S1 y S2 se recomienda cerrarlas en el mismo PR (son pequeñas), S4 y S5 pueden
quedar como deuda registrada.
