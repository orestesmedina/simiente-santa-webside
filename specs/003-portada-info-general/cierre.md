# Verificación de cierre (T339) — F3 Portada e información general

**Fecha**: 2026-10-10 · **Rama**: `003-portada-info-general` · **HEAD al iniciar**: `caf6fcb`
**Responsable**: `devops` · **Veredicto**: **ROJO** (`make ci` falla en *lint*; e2e, deriva, cobertura, smoke y quickstart §0–§12 en verde)

## 1. Resultado por comando (salida real)

### 1.1 `make ci` (lint + test + security) — ❌ ROJO

```
$ make ci          # salida completa en /tmp/opencode/t339-make-ci.log
cd backend && gofmt -l . && go vet ./... && golangci-lint run
internal/platform/storage/local_test.go:93:20: Error return value of `reader.Close` is not checked (errcheck)
	defer reader.Close()
internal/portada/service_admin_test.go:90:13:  Error return value is not checked (errcheck)
internal/portada/service_admin_test.go:161:13: Error return value is not checked (errcheck)
internal/portada/service_admin_test.go:205:13: Error return value is not checked (errcheck)
internal/portada/service_admin_test.go:361:6:  func intptr is unused (unused)
5 issues:
* errcheck: 4
* unused: 1
make: *** [Makefile:80: lint] Error 1      → EXIT=2
```

- `gofmt -l` y `go vet` pasaron; `golangci-lint run` (v2.14.0 idéntica a la de CI:
  `GOLANGCI_LINT_VERSION` por defecto del kit) falla con **5 problemas**, todos en **archivos de
  prueba**. Falla-rápido: `make test` y `make security` no llegaron a ejecutarse dentro de `make ci`.
- Bloques restantes verificados por separado (falla-rápido los saltó dentro de `make ci`):
  - `make test` (backend unitarias + **integración con PostgreSQL real**, y frontend): ✅
    `EXIT=0` — 35 paquetes Go `ok` (dos pasadas: `go test ./...` y `go test -tags=integration ./...`,
    sin FAIL) y Vitest **45 archivos / 269 pruebas passed**.
    *(salida: /tmp/opencode/t339-make-test.log)*
  - `make security` (govulncheck + `npm audit --audit-level=high`): ✅
    `EXIT=0` — govulncheck: «code doesn't appear to call these vulnerabilities» ·
    `npm audit`: **found 0 vulnerabilities**. *(salida: /tmp/opencode/t339-make-sec.log)*

### 1.2 e2e Playwright (T337+T338 + specs de F1/F2) — ✅ 5 passed

```
$ cd frontend && LD_LIBRARY_PATH=/tmp/opencode/pwlibs/usr/lib/x86_64-linux-gnu \
  npx playwright test --config e2e/playwright.config.ts --workers=1
  ✓ acceso.spec.ts … inicializar → login → rol → cuenta → contraseña → permisos → desactivar → bloqueo
  ✓ auditoria.spec.ts … accesos y acciones quedan registrados y no son editables ni borrables
  ✓ portada-panel.spec.ts … editor con permiso gestiona y publica; sin permiso queda denegado y auditado
  ✓ portada-publica.spec.ts … secciones publicadas, idioma es/en con fallback, sin borradores y responsiva
  ✓ status.spec.ts … la página de estado (/health) muestra el estado del sistema
  5 passed (26.7s)      → EXIT=0
```

### 1.3 Deriva de generados — ✅ sin deriva

- Los targets `make sqlc-verify` y `make api-gen` **sí existen** (en `proyecto.mk`, incluido por el
  `Makefile` del kit con `-include proyecto.mk`; la orden previa de que «no existen en el Makefile»
  es una discrepancia de la fuente de la tarea, no del repo). Se verificó con los comandos reales:
  - `cd backend && sqlc generate` → EXIT=0 · `git diff --exit-code -- backend/internal/db backend/sqlc.yaml` → **sin diff** ✅
  - `cd frontend && npm run api:gen` → EXIT=0 (openapi-typescript 7.13.0, 133 ms) ·
    `git diff --exit-code -- frontend/src/api/schema.d.ts` → **sin diff** ✅

### 1.4 Cobertura de `internal/portada/service*.go` — ✅ ≥ 80 %

```
$ cd backend && go test -coverprofile=cover.out ./internal/portada/
service.go:        89,1 % (57/64 decisiones)
service_admin.go:  82,5 % (362/439)
service_audit.go:  95,5 % (42/44)
service_public.go: 87,7 % (64/73)
── TOTAL service*.go: 84,7 % statements (≥ 80 % exigido por §III/T339) ✅
Nota: la métrica "package: 62.7%" que imprime go test es del paquete completo (incluye
handlers/handlers de imágenes aún sin ejecutar en unitarias); el criterio formal es sobre
service*.go y se cumple.
```

### 1.5 Smoke de operación — ✅

| Verificación | Resultado |
|---|---|
| `GET /healthz` | `200` · `{"status":"ok","database":"connected"}` · `Cache-Control: no-store` (como F1) |
| `GET /api/v1/no-existe` | `404` · `{"error":{"code":"not_found","message":"Recurso no encontrado"}}` (ErrorEnvelope §8.1.9) |
| `GET /api/v1/portada?lang=es` | `200` · JSON completo con identidad/horario/whatsapp/redes/contacto · sin `publicationState` |

## 2. Recorrido de `quickstart.md` §0–§12 (los automatizables)

| § | Qué se ejecutó | Resultado |
|---|---|---|
| §0 | `docker ps` (db/redis healthy, backend/frontend up) · `healthz` 200 | ✅ |
| §1 | `GET /api/v1/portada?lang=es` sin cuenta · SPA `/` y `/health` (título «Estado del sistema») · `/api/v1/media/…` con `nosniff`/`inline`/`no-store` | ✅ |
| §2 | Login admin (`ana@ejemplo.com`) · rol con permiso `portada` · cuenta editora · cambio de contraseña en primer acceso | ✅ (nota C del quickstart, abajo) |
| §3 | PUT identidad (es+en, tildes/ñ/¿¡/emoji 🙌), quiénes somos, contacto · inválidos: `nameEs` vacío, email mal, `publicationState` inválido → `400 invalid` con `details` y **sin guardar** · `textEn:""` → guardado como `null` (I6) | ✅ |
| §4 | POST horario (dom 10:00–12:00), WhatsApp direct (normalizado a `+50688888888`) y group (`chat.whatsapp.com`), red | ✅ |
| §4 errores | whatsapp `direct` `abc` → 400 · group no-wa → 400 · duplicado exacto → **409** · red fuera de catálogo → 400 · host no oficial (facebook) → 400 · `25:99` → 400 · `endTime ≤ startTime` → 400 | ✅ (tabla §4 completa en verde) |
| §5 | PATCH elemento por elemento: publicado→draft **desaparece** del público · draft→published reaparece · editar publicado → visible de inmediato (FR-014) · segunda red `instagram` → 409 | ✅ |
| §6 | Subir JPEG (29 KB): `201` con `img_<uuid>.jpg` · SVG renombrado `.jpg` → **400 por firma** (no confía en extensión) · `logoAltEs` obligatorio con `logoFile` → 400 · **C4**: sin referenciar → 404, referenciado y publicado → 200, identidad en draft → 404, republicada → 200 · patrón de nombre: `otro.jpg` → **400**, `img_…` inexistente → **404** (M6) · normalizada/red `instagram` repetida → 409 | ✅ |
| §7 | `lang=en` con textos solo en español → **texto en español, nunca vacío** (FR-009) · `lang=fr` → 400 `details.lang` | ✅ |
| §8 | Cuenta sin permiso (rol solo `eventos`): PUT contacto → `403 forbidden` «No tienes permiso para acceder a este módulo» · sin sesión → `401 unauthenticated` · denegación **registrada** `result='denied'` | ✅ (nota B, abajo) |
| §9 | Auditoría: `home.identity.update/about.update/…` con `targetLabel` «Portada · <sección> · <elemento>» · **regla M5**: `schedule.create` aunque nazca publicado; `home.publish`/`home.unpublish` solo en cambios de estado · `failure` de rechazos · **`home.image.upload`** con «Portada · Imagen · <file>» (I8) · solo lectura | ✅ |
| §10 | Responsividad 320/768/1280 cubierta por e2e `portada-publica` («sin borradores y responsiva»); checklist WCAG (teclado, lector de pantalla, contraste, 44 px) | ✅ automatizado · ⏳ checklist manual → `qa-tester` |
| §11 | Pruebas automatizadas completas | ✅ e2e 5/5 · `make test` ✅ · `make security` ✅ · ❌ lint (§1.1) |
| §12 | Mapa SC-001…SC-013 confirmado contra lo ejecutado (tabla §3 de abajo) | ✅ (SC-010+y SC-008-manual pendientes) |

### Mapa de criterios SC

| SC | Evidencia | Estado |
|---|---|---|
| SC-001 | §1/§2: portada 200 sin cuenta, texto visible; e2e `portada-publica` | ✅ |
| SC-002 | §6 C4 (imágenes), §5, pruebas del service («solo publicado») + e2e | ✅ |
| SC-003 | `no-store` verificado en portada y media; cambios visibles tras recarga (§5) | ✅ |
| SC-004 | e2e `portada-panel` (8,1 s para el flujo completo, margen < 2 min) | ✅ |
| SC-005 / SC-011 | §8: 403 forbidden por API y SPA (e2e) | ✅ |
| SC-006 | §7: `lang=en` con fallback al español, nunca vacío; e2e | ✅ |
| SC-007 | e2e en 320/768/1280 («sin desplazamiento horizontal» implícito en el spec de Playwright) | ✅ (automatizable) |
| SC-008 | e2e responsabilidad WCAG parcial; checklist manual | ⏳ pendiente manual (`qa-tester`) |
| SC-009 | §4: URLs `https://wa.me/…` y de grupo con un clic (e2e) | ✅ |
| SC-010 | Prueba de usabilidad (≥6 personas, protocolo M7) | ⏳ **fuera de alcance de devops: la coordina el humano con `qa-tester`** · informe: `pruebas-usabilidad-SC-010.md` (aún no existe) |
| SC-012 | §1/§5: sección sin publicados desaparece por completo | ✅ |
| SC-013 | §9: todas las mutaciones registradas (`success`/`failure`/`denied`), registro solo lectura | ✅ |

## 3. Pendientes y hallazgos

### Pendientes de cierre de F3
1. **SC-010**: prueba de usabilidad con ≥6 personas (2 por franja 18–35 / 36–59 / 60+, protocolo
   `analyze` M7), observada por `qa-tester` con el humano; el informe
   `pruebas-usabilidad-SC-010.md` queda **pendiente** (fuera del alcance de devops, no automatizable).
2. **SC-008** (checklist manual de accesibilidad WCAG de quickstart §10.1–10.6): la recorre
   `qa-tester` en la validación; el e2e solo cubre la parte automatizable (3 anchos).
3. **Lint en rojo** (única causa del fallo de `make ci`): 5 problemas en archivos de prueba, todos
   triviales, para `dev-backend` la primera vuelta del bucle de corrección:
   - `backend/internal/platform/storage/local_test.go:93` → `defer reader.Close()` sin error; usar
     `defer func() { _ = reader.Close() }()`.
   - `backend/internal/portada/service_admin_test.go:90/161/205` → `requireKind` (definido en
     `service_test.go:568`) devuelve `*apperr.Error` y las llamadas descartan; escribir `_ =` o
     volverla void.
   - `backend/internal/portada/service_admin_test.go:361` → helper `intptr` sin usar; eliminarlo.

### Discrepancias de la fuente de la tarea (sin cambios en el repo)
- **A.** La instrucción «los targets make sqlc-verify / api-gen no existen en el Makefile» es
  **imprecisa**: existen en `proyecto.mk` (incluido por `Makefile` con `-include proyecto.mk`) y
  funcionan; la verificación se hizo igualmente con los comandos directos que equivalen a esos targets.
- **B.** quickstart §8 dice «Sin sesión → 401 unauthenticated; con mustChangePassword → 403 con la
  razón del guard (F2)»: el orden real observado es **guard de contraseña primero** (403
  `password_change_required` en la primera sesión sin excepto sesión válida), y `401
  unauthenticated` sin sesión. Documentado para `documentador` (la prioridad de guards es la de F2,
  no es un defecto).
- **C.** quickstart §2 usa credenciales de ejemplo (`admin@ejemplo.com` / `Contraseña1!`) que **no**
  coinciden con el administrador sembrado por el e2e/setup (`ana@ejemplo.com` /
  `Semilla.2026`, helpers de `frontend/e2e/helpers.ts`). Sugerencia para `documentador`: alinear el
  texto del quickstart al admin real o indicar que la cuenta depende del entorno.

## 4. Conclusión

- ✅ e2e (5)**·** deriva de generados**·** cobertura service*.go 84,7% ≥ 80%**·** smoke
  `/healthz` + 404 ErrorEnvelope + portada pública**·** quickstart §0–§9 y §12**.
- ❌ `make ci`: **lint** con 5 hallazgos en archivos de prueba (los bloques `test` y `security`,
  verificados aparte, están en verde).
- T339 queda **NO marcada**: se cumplirá la condición «`make ci` y e2e en verde» cuando
  `dev-backend` corrija los 5 hallazgos de lint y `make ci` pase íntegro. Los pendientes manuales
  (SC-010, SC-008) se señalan para `qa-tester` y el humano.
