# Verificación de cierre (T1036) — F10 Sistema de diseño y movimiento

**Fecha**: 2026-10-11 · **Rama**: `010-sistema-diseno-movimiento` · **HEAD al iniciar**: `7b84c1c`
**Responsable**: `devops` · **Veredicto**: **VERDE** ✅ — F10 es un cambio **exclusivamente de frontend**
+ specs: **sin deriva de generados, sin backend, sin migraciones nuevas, sin variables de entorno
nuevas y sin cambios en infra/CI**. Stack levantado y sano, build de producción ok, unitarias
65/65 archivos y 402/402 pruebas en verde (idénticas a las de T1034: sin deriva de pruebas).
Pendiente de la fase de entrega: **aprobación humana** (SC-002/SC-013/SC-014 no se auto-aprueban).

## 1. Qué se verificó y resultado

### 1.1 Sin deriva en el contrato de API — ✅

```
$ cd frontend && npm run api:gen        → openapi-typescript 7.13.0 · EXIT=0
$ git diff --exit-code -- frontend/src/api/schema.d.ts → SIN DIFF ✅
```

F10 no cambia la API; el OpenAPI (`backend/api/openapi.yaml`) y el contrato tipado quedan intactos.

### 1.2 Sin cambios en backend, migraciones ni infra — ✅

Comparado contra la **base de la rama** (`17094ce`, merge-base `main...HEAD`):

| Alcance | `git diff 17094ce..HEAD` | Resultado |
|---|---|---|
| `backend/` (incl. `migrations/`) | vacío | ✅ **0 líneas** — no hay migraciones nuevas (la última sigue siendo `000006_*`) ni código de servidor |
| `.env.example` | vacío | ✅ sin variables nuevas |
| `docker-compose.yml`, `Dockerfile*`, `.github/` | vacío | ✅ sin cambios de infra ni CI |
| Total del branch | 81 archivos | solo `frontend/` (59 src + 7 e2e + index.html), `specs/010-*` (13) y `docs/producto` (1) |

### 1.3 Stack levantado y sano — ✅

```
$ make up   → db / redis / backend / frontend Up
docker compose ps → db healthy · redis healthy · backend running (Up 14 h) · frontend running
$ curl /healthz → HTTP 200   $ curl http://localhost:5173/ → HTTP 200 (SPA)
```

### 1.4 Build de producción del frontend — ✅

```
$ cd frontend && npm run build → ✓ built in 3.01s
dist: index-D9rn8pMN.js 548.26 kB (gzip 162.91 kB) · index-CcvwiLah.css 46.84 kB
```

Nota (preexistente, no bloqueante): aviso de chunk > 500 kB tras minificar. Ya existía antes de
F10; si se quiere atacar (code-splitting), es tarea de mejora aparte, no del cierre.

### 1.5 Unitarias de frontend (sin deriva de pruebas) — ✅

```
$ cd frontend && npm run lint      → 0 problemas      $ npm run typecheck → ok
$ npm test -- --run                → Test Files 65 passed (65) · Tests 402 passed (402)
```

Coincide exactamente con la evidencia registrada en T1034 (`quickstart.md` §8.1): ninguna prueba
de F2/F3 fue editada para pasar (SC-008 / RG10-6).

### 1.6 `.env.example` correcto — ✅ (sin cambios)

F10 **no añade variables nuevas**. Uso de entorno en el código: solo `import.meta.env.VITE_API_URL`
(ya documentada). El diff de la rama sobre `.env.example` es vacío y el archivo mantiene sus 57
líneas de F2/F3 (PostgreSQL, backend/sesión, BOOTSTRAP_TOKEN, imágenes, Redis, puertos, VITE_API_URL).

### 1.7 CI coherente con F10 — ✅ (sin cambios necesarios)

`.github/workflows/ci.yml` (único workflow del repo) ya cubre todo lo que F10 produce:

| Job | Qué corre para F10 |
|---|---|
| `agentes` | verifica kit + `sincronizar.py --verificar` (no tocado por F10) |
| `controles` | PR no modifica migraciones existentes (F10 no toca `backend/migrations/`) ✅ |
| `frontend` | `npm ci` · `lint` · `typecheck` · `test -- --run` · **`build`** · `npm audit` — pasa |
| `backend` | corre por detección, pero el diff del branch en `backend/` es vacío: *takedown a lo que no cambió* |
| `secretos` | gitleaks sobre todo el diff ✅ (F10 no introduce secretos; tokens solo *design* en CSS/TSX) |

Observaciones (reportes, no cambios):
- **A**: los **e2e de Playwright no corren en CI** (ningún workflow los ejecuta). Es la situación de
  F1–F3 y F10 no la empeora; si el equipo quiere subirlos, es una mejora de pipeline separada.
- **B**: no existe workflow de **CD/despliegue**. Coherente con el estado del proyecto: no hay
  entorno de producción aprobado todavía; no se agrega aquí.

## 2. No-regresión de F2/F3 — verificada ✅

Según la evidencia congelada de **`quickstart.md` §8.1 (T1034, 2026-10-11)** y re-ejecutadas hoy por
esta verificación en su parte automatizable local:

| Comprobación | Resultado |
|---|---|
| Unitarias + componentes (`npm test -- --run`) | **402/402 en verde** (65 archivos) — re-ejecutado hoy, idéntico a §8.1 |
| Estilo y tipos (`lint` / `typecheck`) | en verde, 0 avisos |
| E2E suites congeladas de F2/F3 (8 specs: `acceso`, `auditoria`, `portada-accesibilidad`, `portada-imagenes`, `portada-panel`, `portada-patch-null`, `portada-publica`, `status`) | **8/8 en verde** (31,4 s) — §8.1 |
| E2E suites nuevas de F10 (5 specs: `diseno-movimiento`, `navegacion-movil`, `tema-oscuro`, `movimiento`, `movimiento-reducido`) | **22/22 en verde** (24,0 s) — §8.1 |
| Suites de F2/F3 editadas | **0 líneas eliminadas**; solo adiciones permitidas (D12) en `homePage.test.tsx` e `i18n.test.tsx` |
| Contrato semántico congelado (R10-9) | anclas, `ss.lang`, claves `nav.label`/`a11y.skip`/`lang.*`/`brand.name`/`footer.welcome`, `tabCycle < 30` a 320 px — intactos (§8.1) |

## 3. Notas de entorno

- **Rate-limit de login (M-4, F2)**: el backend limita `POST /api/v1/auth/login` y
  `/setup/initialize` a **20 peticiones/min por IP** (`ratelimit.go`). Con las 13 specs e2e, un run
  único acumula suficientes logins para que alguna spec reciba `429`; **ejecutar los e2e en dos
  tandas** (congeladas / nuevas), como se hizo en T1034 y así quedó registrado en §8.1.
- **Chromium en WSL**: en este equipo Chromium no arranca por falta de librerías del sistema
  (`libnspr4`/`libnss3`, sin `sudo`). Los e2e se ejecutan con la **imagen oficial Docker de
  Playwright** (`mcr.microsoft.com/playwright:v1.63.0-jammy`) contra el stack real levantado con
  `docker compose up -d --build frontend`, igual que en F1–F3.
- **Repetibilidad**: `cd frontend && npm run api:gen && git diff --exit-code -- src/api/schema.d.ts`
  es el oro para detectar deriva de contrato; `make ci` para pipeline local; `make e2e` requiere
  stack levantado.

## 4. Conclusión

- ✅ Sin deriva de contracto (`api:gen` idempotente), sin backend, sin migraciones nuevas, sin
  variables de entorno nuevas, sin cambios en compose/Docker/CI.
- ✅ Stack levantado (`healthz` + SPA 200), `npm run build` ok, lint/typecheck/unit 402/402 en verde.
- ✅ No-regresión F2/F3 certificada: unit 402 + e2e congeladas 8/8 + nuevas 22/22 (`quickstart.md`
  §8.1), con las notas de entorno de M-4 (rate-limit login) y la imagen de Playwright.
- ✅ CI existente ya cubre el pipeline de F10; observaciones A (e2e fuera de CI) y B (sin CD)
  reportadas, sin cambios de pipeline en esta feature.
- **T1036 marcada `[X]`** en `tasks.md`: las 36 tareas de F10 (Fase 1–9, incluidas las 3 excepciones
  §III declaradas en N4) quedan completas.
- **Queda para la fase de entrega (no es de devops)**: la **aprobación humana** de SC-002
  («moderno, acorde a la marca»), SC-013 (0 elementos copiados) y SC-014 (movimiento perceptible y
  no mareante) — la checklist de `quickstart.md` §9 está lista para la revisión del cliente, junto
  con las validaciones de `qa-tester`/`revisor-codigo`/`seguridad` y la apertura del PR.
