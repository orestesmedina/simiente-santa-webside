# Revisión de seguridad — bug «imágenes de la portada no cargan» (F3)

**Fecha**: 2026-10-10 · **Rama**: `003-portada-info-general` · **Auditor**: `seguridad` (solo
lectura; este reporte es la única escritura).
**Objeto**: commit **`ec2ae76`** — `fix(frontend): F3 resuelve las URLs de media contra la API`
(`mediaUrl()` en `frontend/src/api/client.ts` + su uso como `src` de `<img>` en `IdentityHero`,
`PublicLayout` y `PublicFooter`). Validación del bucle de bug (paso 4/5).

**Verificado que el cambio toca solo 7 archivos de frontend** (`git show ec2ae76 --stat`):
`client.ts`, `client.test.ts`, `IdentityHero.tsx`, `PublicFooter.tsx`, `PublicLayout.tsx`,
`homePage.test.tsx` y `e2e/portada-imagenes.spec.ts`. **Nada de `routes.go`, handlers,
middleware, `platform/storage`, `nginx.conf`, migraciones ni contrato.**

---

## 1. Riesgo del helper `mediaUrl` (pregunta 1 del encargo)

`mediaUrl` (`client.ts:26-31`) prefija `API_BASE_URL` (build time, `client.ts:8`, sin secreto)
a los `path` relativos y deja intactas las URLs que casan con
`ABSOLUTE_URL_PATTERN = /^(?:[a-z][a-z0-9+.-]*:|\/\/)/i` (`client.ts:11`), propagando
`undefined`/`''` (el componente aplica su respaldo). Evaluación por vector:

| Entrada hypothetical | Comportamiento | Riesgo en `<img src>` |
|---|---|---|
| `javascript:alert(1)` | Pasa intacta (casa con el patrón) | **Inerte**: ningún navegador ejecuta `javascript:` en `src` de `<img>`; y el valor llega por atributo React (DOM API, escapado) |
| `data:image/svg+xml,<svg onload=…>` | Pasa intacta | **Inerte**: el SVG cargado como imagen deshabilita scripting por especificación (contexto imagen); otro MIME no descodifica y cae al `onError` |
| `//host/x.png` | Pasa intacta | Solo riesgo de degradado a `http://` si el borde sirviera la SPA sin TLS (nota M3, despliegue) |
| Relativa `/api/v1/media/…` | `API_BASE_URL + path` | El comportamiento buscado (el fix) |

**¿Permitir absolutas es aceptable? Sí**, por tres barreras verificadas:

1. **El backend nunca emite URLs absolutas hoy.** `localizeIdentity`
   (`service_public.go:225-241`) compone `MediaPathPrefix + *LogoFile`/`*CoverImageFile`
   (`/api/v1/media/<fileName>`, `service.go:40`) a partir de columnas que solo admiten nombres
   **generados por el servidor**: `storage.Save` (`storage.go:64-71`) siempre acuña
   `img_<uuid>.<ext>` tras validar la firma binaria (`Detect`, `storage.go:89-100`: solo
   JPEG/PNG/WebP; **SVG y GIF rechazados**), tamaño y escritura atómica. No hay camino donde un
   usuario deposite un `data:`, un `javascript:` ni un host arbitrario en `logoUrl`/`coverImageUrl`.
2. **El consumo es exclusivamente `src` de `<img>`**: `BrandLogo` (renderiza `<img src={src}>`),
   `IdentityHero:22,38`, `PublicFooter:30`, `PublicLayout:51` — grep de `mediaUrl(` en
   `frontend/src`: **solo** esos 3 componentes + pruebas. Grep de
   `dangerouslySetInnerHTML|innerHTML` en `frontend/src`: **0**. Sin `href` nuevo.
3. **El esquema inerte en `img`** (`javascript:`, `vbscript:`, `file:`…) no ejecuta en contexto
   imagen; React además fija atributos vía DOM (sin inyección HTML).

**SSRF: no existe** — la URL la consume el **navegador** (`<img>`); ni el backend ni ningún
servidor del sistema hace `fetch` de URLs de identidad. La subida de imágenes valida bytes
(firma binaria), no URLs.

**Exposición nueva: ninguna** — `GET /api/v1/media/{fileName}` ya era público por diseño
(FR-013/FR-019, `routes.go:30` en `RegisterPublic`) y el fix solo cambia **qué origen** lo pide
(el de la API en vez del de la SPA). Mismos bytes, mismo endpoint, sin cambio de autenticación.
`<img>` no requiere CORS para pintar.

## 2. Endpoint de media re-verificado (pregunta 2 del encargo)

Sin cambios en este commit, pero re-verificado íntegro (política analyze C4/M6):

- **Nombre**: `OpenMedia` (`service_public.go:176-181`) exige `storage.ValidName` — patrón
  exacto `^img_[0-9a-f-]{36}\.(jpg|png|webp)$` (`storage.go:52`): fuera del patrón → **400
  `invalid`** (`routes_test.go:75`, `main_test.go:359`). Path traversal imposible por diseño:
  el nombre jamás lo controla el cliente.
- **Solo publicado**: `IsHomeFilePublished` (`repository_home.go:222`) → no referenciado por
  contenido publicado → **404 `not_found`**. Cubierto por pruebas de unidad
  (`service_public_test.go:171-204`: patrón inválido 400, huérfano 404, publicado 200, físico
  ausente 404), de integración (`repository_home_integration_test.go:211-245`: borrador →
  false, publicado → true, huérfano → false, **retirado → false**) y de handler
  (`handler_media_test.go:38-…`: 200 + `nosniff` + `inline` + `no-store`). **IDOR descartado**:
  el acceso lo gobierna el estado de publicación, no la adivinanza de nombres (UUID aleatorio).
- **Cabeceras seguras** (`handler_media.go:38-41`): `X-Content-Type-Options: nosniff`,
  `Content-Disposition: inline`, `Cache-Control: no-store` (la descarga depende del estado de
  publicación, RG3-8) y `Content-Type` derivado de la extensión permitida (`ContentTypeFor`,
  `storage.go:105-116`): nunca se sirve como HTML/ejecutable.

## 3. `make security` (pregunta 3 del encargo) — 2026-10-10, verde

- `govulncheck ./...`: **0 vulnerabilidades que afecten al código**. 1 aviso en el grafo de
  módulos (ver M5), no llamado.
- `npm audit --audit-level=high`: **0 vulnerabilidades**.

## 4. Hallazgos

### BLOQUEANTE

Ninguno.

### IMPORTANTE

Ninguno.

### MENOR

- **M1 · Observación de integridad del entorno (a verificar con el humano) · sin CWE (no es
  vulnerabilidad)**. Durante la ventana de auditoría, un proceso concurrente dejó **staged un
  revert exacto del fix** en la working tree y luego lo restauró. Evidencia observada
  directamente: a las ~14:47–14:49 `git status --short` mostraba `M ` (staged) en los 4 archivos
  del fix, `git diff --cached` mostraba la eliminación de `mediaUrl` (−42/+5) y el contenido de
  `client.ts`/componentes era el pre-fix; mtimes de los archivos **14:46:39** y del índice
  **14:47:56** (posteriores al commit, 14:44:07); después (14:48:36–38) los archivos y el índice
  volvieron al estado del fix. **Estado final verificado al cerrar (14:49:58): `git status`
  limpio, worktree = index = HEAD = `ec2ae76`, `mediaUrl` presente en los 4 archivos.** Ningún
  commit ni reset en el reflog. Impacto de seguridad: **ninguno** (el estado revertido es
  equivalente en seguridad: mismo `<img>` escapado por React; el revert solo reintroduciría el
  bug funcional, y lo detectaría el e2e de regresión). Acción: confirmar con el orquestador que
  ningún escritor debía estar activo durante la validación, y antes del merge verificar que el
  tip sigue siendo `ec2ae76` y el árbol está limpio.
- **M2 · Passthrough de cualquier esquema, sin allowlist · CWE-79 (latente, hoy inerte)**.
  `ABSOLUTE_URL_PATTERN` acepta **cualquier** esquema `[a-z][a-z0-9+.-]*:`, no una lista
  explícita (`https?:`, `data:`). Hoy es inerte (uso exclusivo `img` + backend que solo emite
  rutas relativas, §1) y lo califico **aceptable**; se volvería relevante solo si `mediaUrl` se
  reutilizara algún día en un `href`/`src` de script. Acción (futuro, no bloqueante): restringir
  el passthrough a `https?://`|`data:` y documentar en el helper el contrato «solo para `<img>`».
- **M3 · Degradado de protocolo en despliegue (a verificar en producción)**. Las URLs
  protocolo-relativas (`//host`) heredan el esquema de la página, y `API_BASE_URL` por defecto
  es `http://localhost:8080` (dev). El contenedor de la SPA escucha en 80 (`nginx.conf`):
  garantizar terminación TLS en el borde y hornear `VITE_API_URL` con `https://` en las builds
  de producción para no pedir las imágenes por HTTP. No lo introduce este cambio (también
  aplicaba a `apiFetch`), es nota de despliegue.
- **M4 · Sin CSP en la SPA (preexistente, defensa en profundidad)**. `nginx.conf` no declara
  `Content-Security-Policy`. Una política con `img-src 'self' <origen-api> data:` acotaría por
  construcción cualquier uso futuro indebido del passthrough de absolutas (y de `blob:`/`file:`
  no intencionales). Preexistente a este commit; no lo agrava el cambio (las URLs de media ya
  cruzan al origen de la API por diseño).
- **M5 · GO-2026-5932 (preexistente, sin acción posible)**. `golang.org/x/crypto@v0.57.0`
  arrastra el aviso del subpaquete `openpgp` (desmantenido); el proyecto **no lo importa**
  (usa `x/crypto` para bcrypt) y no existe fix (`Fixed in: N/A`). No lo introduce este commit
  (es frontend-only). Seguimiento upstream.

## 5. Cobertura del encargo

| Encargo | Resultado |
|---|---|
| 1. Riesgo del helper en `<img src>` | Sin vector: esquemas inertos en `img`, React escapa atributos, 0 `dangerouslySetInnerHTML`; permitir absolutas es **aceptable** (§1) |
| 2. Sin XSS/SSRF/exposición nueva; media solo contenido publicado | Verificado (§1-§2): 400 patrón inválido, 404 no publicado/retirado/huérfano, 200 solo publicado, cabeceras `nosniff`/`inline`/`no-store` |
| 3. `make security` | Verde: govulncheck 0 en código, npm audit 0 (§3; M5 solo aviso no llamado) |

## 6. Veredicto

**APROBADO** — `ec2ae76` no introduce ninguna vulnerabilidad (0 BLOQUEANTES, 0 IMPORTANTES,
5 MENORES: 1 observación de entorno ya auto-resuelta y verificada al cierre, 2 latentes/de
hardening sin efecto hoy, 1 nota de despliegue, 1 aviso preexistente de dependencia). El fix es
una resolución pura de URL del lado del navegador sobre un endpoint público que ya existía, y
el backend ya garantiza que las URLs de identidad solo pueden ser rutas relativas a archivos
raster generados por el servidor y referenciados por contenido publicado. Condición para el
merge (de M1): confirmar que el tip es `ec2ae76` con el árbol limpio.
