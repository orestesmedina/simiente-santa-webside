# Verificación de cierre (T339) — F3 Portada e información general

**Fecha**: 2026-10-10 · **Rama**: `003-portada-info-general` · **HEAD al iniciar**: `caf6fcb`
**Responsable**: `devops` · **Veredicto**: **VERDE** ✅ — `make ci` completo EXIT=0 (tras el fix de
lint `dda8120`), e2e 5/5, sin deriva, cobertura ≥ 80 %, smoke ok, quickstart §0–§12 recorrido.
Pendientes pendientes de la fase de validación/entrega: SC-010 (usabilidad con personas) y la parte
manual de SC-008 (checklist WCAG, `qa-tester`); discrepancias del quickstart para `documentador`.

## 1. Resultado por comando (salida real)

### 1.1 `make ci` (lint + test + security) — ✅ EXIT=0 (definitivo)

```
$ make ci
cd backend && gofmt -l . && go vet ./... && golangci-lint run   →  0 issues
cd backend && go test ./... && go test -tags=integration ./...  →  35 paquetes ok (unit + integración, PostgreSQL real)
cd frontend && npm test -- --run                                →  Test Files 45 passed (45) · Tests 269 passed (269)
cd backend && govulncheck ./...                                 →  sin vulnerabilidades llamadas por el código
cd frontend && npm audit --audit-level=high                     →  found 0 vulnerabilities
CI_EXIT=0        (salida completa: /tmp/opencode/t339-make-ci2.log; primera pasada roja: t339-make-ci.log)
```

- **Primera pasada** (2026-10-10 temprano): `golangci-lint run` (v2.14.0, idéntica a la de CI) falló
  con 5 problemas `errcheck`/`unused` en archivos de prueba → corregidas por `dev-backend` en
  commit **`dda8120`** (`test(backend): F3 corregir hallazgos de golangci-lint en pruebas`:
  `local_test.go`, `service_admin_test.go`, `service_public_test.go`).
- **Pasada definitiva**: `make ci` completo → **EXIT=0**: lint 0 issues · 35 paquetes Go ok en
  doble pasada (unit + `-tags=integration` con PostgreSQL real) · Vitest 45/45 archivos y 269/269
  pruebas · govulncheck y `npm audit` sin vulnerabilidades.

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

### 1.3 Deriva de generados — ✅ sin deriva (reverificada tras `dda8120`)

- Los targets `make sqlc-verify` y `make api-gen` **sí existen** (en `proyecto.mk`, incluido por el
  `Makefile` del kit con `-include proyecto.mk`; la orden previa de que «no existen en el Makefile»
  es una discrepancia de la fuente de la tarea, no del repo). Se verificó con los comandos reales,
  **dos veces** (antes y después del fix de lint, que solo tocó archivos de prueba):
  - `cd backend && sqlc generate` → EXIT=0 · `git diff --exit-code -- backend/internal/db backend/sqlc.yaml` → **sin diff** ✅
  - `cd frontend && npm run api:gen` → EXIT=0 (openapi-typescript 7.13.0) ·
    `git diff --exit-code -- frontend/src/api/schema.d.ts` → **sin diff** ✅

### 1.4 Cobertura de `internal/portada/service*.go` — ✅ ≥ 80 % (reverificada tras `dda8120`)

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
| §11 | Pruebas automatizadas completas | ✅ `make ci` EXIT=0 · e2e 5/5 · cobertura 84,7 % |
| §12 | Mapa SC-001…SC-013 confirmado contra lo ejecutado (tabla §3 de abajo) | ✅ (SC-010 y parte manual de SC-008 en validación/entrega) |

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

### Pendientes de la fase de validación/entrega (NO bloquean T339)
1. **SC-010 — prueba de usabilidad con personas** (≥6 personas, 2 por franja 18–35 / 36–59 / 60+,
   protocolo mínimo del `analyze` M7): la coordina el **humano** con `qa-tester` sobre la portada
   real; el informe `pruebas-usabilidad-SC-010.md` se adjunta en la fase de validación/entrega
   (umbral: ≥ 90 % de tareas completadas). No automatizable; fuera de `make ci`/`make e2e`.
2. **SC-008 — parte manual de accesibilidad** (checklist WCAG 2.1 AA de quickstart §10.1–10.6:
   teclado, lector de pantalla, contraste, áreas de 44 px, 200 % de ampliación): la recorre
   `qa-tester` en la validación; el e2e ya cubre la parte automatizable (320/768/1280).
3. **Para `documentador`** (discrepancias del quickstart, sin defecto de código):
   - §2 usa `admin@ejemplo.com` / `Contraseña1!` como ejemplo, pero el administrador real del
     entorno es el sembrado por setup/e2e: **`ana@ejemplo.com` / `Semilla.2026`**
     (`frontend/e2e/helpers.ts`). Alinear texto o aclarar que la cuenta depende del entorno.
   - §8: el orden real de guards en la primera sesión es el de F2: `403
     password_change_required` (mustChangePassword) **antes** que la respuesta de permisos; sin
     sesión → `401 unauthenticated`. Documentar la prioridad para evitar confusión.

### Resueltos durante esta verificación
- **Lint en rojo** (única causa del primer `make ci` rojo): 5 problemas `errcheck`/`unused` en
  archivos de prueba, corregidos por `dev-backend` en **`dda8120`**; `make ci` pasó íntegro
  (EXIT=0) en la pasada definitiva.

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

- ✅ `make ci` **EXIT=0** (lint 0 issues tras `dda8120` · 35 paquetes Go unit+integración · 269
  pruebas frontend · govulncheck/npm audit limpios) · e2e **5/5** · deriva de generados **null**
  · cobertura `service*.go` **84,7 % ≥ 80 %** · smoke `/healthz` + 404 ErrorEnvelope + portada
  pública · quickstart **§0–§12** recorrido con las tablas de errores, la cadena de imágenes C4 y
  el mapa SC-001…SC-013 confirmados.
- **T339 marcada `[X]`** en `tasks.md`: las 40 tareas de F3 quedan completas.
- **Quedan para la fase de validación/entrega** (no forman parte de T339): SC-010 (usabilidad con
  personas, humano + `qa-tester`) y la parte manual de SC-008 (checklist WCAG, `qa-tester`);
  para `documentador`, las dos discrepancias del quickstart señaladas en §3.
- Próximo paso del orquestador: fase de entrega (revisión final `qa-tester`/`revisor-codigo`/
  `seguridad` en paralelo, PR, CHANGELOG/README, cierre de costos con `make costos CERRAR=1`).
