# Revisión de código — F1 completa (T001–T029)

**Rama**: `001-estructura-base` · **Fecha**: 2026-10-03 · **Revisor**: `revisor-codigo` (fase 5/7 del flujo `equipo-feature`)
**Ámbito**: revisión `analyze` (coherencia spec → plan → tareas → código) + revisión de código de la F1 completa, T001–T029.
**Contexto verificado**: `.specify/memory/constitution.md`, `specs/001-estructura-base/{spec.md,plan.md,tasks.md,ux.md}`, `docs/tecnico/arquitectura.md` (§1.2 R1–R8, §4, §5.11, §8), skills `go-backend` / `react-frontend` / `postgres-db`, `backend/api/openapi.yaml`.
**Solo lectura**: este informe no modifica código, `tasks.md` ni `estado.md`.

---

## Veredicto

## **RECHAZADO** (cambios acotados; sin rediseño)

Un bloqueante (el gate de calidad del propio proyecto está rojo: `make lint` falla) y tres
hallazgos mayores de coherencia con artefactos aprobados (`ux.md`, `arquitectura.md`). El resto de
la F1 está en muy buen estado: arquitectura por capas respetada, contrato idéntico byte a byte,
suite del sobre cubriendo los 7 casos, cobertura del service ≥80 % (100 % en `service.go` y
`handler.go`), pruebas de comportamiento y no de implementación, y las desviaciones declaradas son
casi todas aceptables (dictamen por caso abajo). Corregido el bloqueante y los tres mayores, el
cambio pasa a aprobado sin discusión de fondo.

---

## Evidencia ejecutada (2026-10-03, entorno del equipo: Go 1.27.1, golangci-lint 2.14.0, Node 22.23.3)

| Comando | Resultado |
|---|---|
| `cd backend && gofmt -l .` | ✅ sin salida |
| `cd backend && go vet ./...` | ✅ OK |
| `cd backend && golangci-lint run ./...` | ❌ **1 issue: `QF1011` en `internal/platform/httpserver/server.go:88`** (→ bloqueante B1) |
| `make lint` | ❌ **falla** por lo anterior (`Makefile:72`, target del kit: `gofmt + go vet + golangci-lint` + frontend) |
| `cd backend && go test ./...` | ✅ 9 paquetes en verde |
| `go test -cover ./internal/status/` | ✅ **87,0 %** del paquete; `service.go` **100 %**, `handler.go` **100 %** (umbral §III ≥80 % en `service/` cumplido) |
| `diff -u specs/001-estructura-base/contracts/openapi.yaml backend/api/openapi.yaml` | ✅ idénticos byte a byte (D9) |
| `cd frontend && npm run api:gen` + `git diff` | ✅ sin deriva de `schema.d.ts` (R4/R20) |
| `make sqlc-verify` | ✅ en verde con la omisión documentada (dictamen D6 abajo) |
| `cd frontend && npm run lint / typecheck / test -- --run / build` | ✅ todo en verde (24 pruebas, 6 archivos) |
| `grep -rn "fetch(" frontend/src` fuera de `api/client.ts` | ✅ 0 coincidencias (regla de oro `react-frontend`) |
| `grep` de `: any`, `dangerouslySetInnerHTML`, `localStorage` en `frontend/` | ✅ 0 coincidencias (`strict: true` en `tsconfig.json:15`) |
| `git ls-files` de `.env`, `frontend/dist`, `node_modules` | ✅ ninguno versionado (`.gitignore` del kit los cubre) |

---

## Bloqueantes

### B1 · `make lint` / `make ci` están en rojo: `golangci-lint` reporta `QF1011`

- **Evidencia**: `backend/internal/platform/httpserver/server.go:88` —
  `var handler http.Handler = envelopeFallback(mux, logger)` →
  `QF1011: could omit type http.Handler from declaration (staticcheck)`. Reproducido con
  `make lint` (`Makefile:72`), que es exactamente el gate que usa `make ci` (`Makefile:79`) y el
  paso `golangci-lint` de `.github/workflows/ci.yml:92-93`.
- **Regla rota**: constitución §V ("`gofmt`, `go vet` y `golangci-lint` sin errores", DEBE);
  plan §"Estrategia de pruebas" (`make ci` = el veredicto completo); `tasks.md` Fase 3 ("toda la
  fase cumple §V"). Con esto en rojo, F1 no puede cerrar: su propia US3/SC-003 exige que el
  veredicto automático de cada cambio sea "apto", y hoy el veredicto local es "no apto".
- **Propuesta (una línea)**: `var handler = envelopeFallback(mux, logger)` en `server.go:88`
  (el tipo se infiere; `registrar.go:52` sí necesita la anotación porque convierte
  `http.HandlerFunc` a `http.Handler` y ahí no dispara el check). Alternativa equivalente:
  `var handler http.Handler = http.HandlerFunc(...)` no aplica; basta quitar el tipo.
- **Nota adicional**: el CI del kit usa `golangci/golangci-lint-action@v6` sin versión fijada
  mientras localmente hay 2.14.0; si las versiones divergen, el mismo código puede dar distinto
  veredicto local y en CI. Merece una línea en la propuesta al kit (plan R2/R4): fijar la versión
  del linter.

---

## Mayores

### M1 · La "hora de la última consulta" no aparece en los estados de error (contradice `ux.md` aprobado)

- **Evidencia**: `frontend/src/features/status/hooks/useSystemStatus.ts:36-37` usa
  `query.dataUpdatedAt > 0` para calcular `fechaConsulta`. En TanStack Query v5,
  `dataUpdatedAt` solo se actualiza en el dispatch `success` (`query-core/build/modern/query.js`,
  caso `"success"` → `successState(...)`); en el dispatch `"error"` solo se actualiza
  `errorUpdatedAt`. Consecuencias observables:
  1. **Primer consulta fallida (error A o error B)**: `dataUpdatedAt` queda en 0 →
     `fechaConsulta` es `undefined` → `StatusResult.tsx:31/52/78` no renderiza "Última consulta".
  2. **Éxito y luego reconsulta fallida**: `dataUpdatedAt` conserva la hora del último **éxito** y
     se muestra como "Última consulta" del resultado de error — una hora **anticuada** como si
     fuera la del intento (contradice "nunca un estado memorizado", FR-003/escenario 2.4).
- **Regla rota**: `ux.md` §2 (`ux.md:60`): "Hora de la última consulta: **se muestra siempre que
  hay un resultado (éxito o error)**, con la hora exacta"; `ux.md` §6 manda literalmente
  "Última consulta: [hora exacta]" en el error A (`ux.md:179`) y "Última consulta: [hora exacta
  del intento]" en las dos variantes del error B (`ux.md:189`, `ux.md:197`). Además el doc-comment
  del hook afirma "Fecha de la última consulta con resultado (**éxito o error**)"
  (`useSystemStatus.ts:10`), que hoy es falso. Constitución §I (el código implementa lo pedido,
  no una interpretación).
- **Propuesta**: calcular la marca de tiempo con
  `query.dataUpdatedAt || query.errorUpdatedAt` (o `errorUpdatedAt` cuando `query.isError`), y
  añadir en `frontend/src/features/status/status.test.tsx` aserciones de `Última consulta:` en los
  casos de error A (línea 57) y error B (líneas 84 y 99) — hoy solo se comprueba en el caso de
  éxito (`status.test.tsx:54`), por eso la brecha pasó los tests.

### M2 · El "logger por petición" nunca se crea en el contexto: `WriteError` no puede usarlo (código muerto en producción)

- **Evidencia**: `backend/internal/platform/httpserver/error.go:36-46` define
  `ContextWithRequestLogger` / `RequestLoggerFromContext`, y `logError` (`error.go:125-132`) las
  consume como primera opción. Pero **nadie las llama en código de producción** (grep: los únicos
  usos son `error.go` y la prueba `error_test.go:230`): `middleware.RequestID` solo guarda el ID
  (`middleware/requestid.go:34`) y `middleware.Logging` construye un logger hijo **local** con
  `logger.Request(...)` (`middleware/logging.go:33`) sin publicarlo en el contexto.
- **Regla rota**: `arquitectura.md` §2.1 (fila `logger/`: "logger por petición (hij con
  `request_id`, ruta y método)") y §6 paso 2 (`arquitectura.md:1033-1034`: "`request-id`: … crea
  el **logger por petición** (hij de `slog` con `request_id`, método y ruta)"), que D12 del plan
  aterriza igual. También FR-010/US4: la capacidad transversal "tratamiento de cada petición"
  existe a medias: la trazabilidad llega al log de error solo por el fallback de `request_id`
  (`error.go:128-131`), y los registros de `WriteError` pierden método/ruta.
- **Dictamen sobre la desviación declarada** ("`RequestID` que no crea el logger (lo hace
  `Logging`)"): **aceptable como decisión, inaceptable como está implementada** — quién lo crea es
  indiferente, pero hoy *nadie* lo crea en el contexto, así que la desviación declarada no describe
  el código y la API de contexto queda sin productor.
- **Propuesta (una de las dos, ambas válidas)**:
  a. *Cablear*: en `middleware.RequestID` (o al inicio de `Logging`), hacer
     `r = r.WithContext(httpserver.ContextWithRequestLogger(r.Context(), applogger.Request(parent, id, method, path)))`
     y que `Logging`/`WriteError` usen ese logger; o
  b. *Simplificar*: eliminar `ContextWithRequestLogger`/`RequestLoggerFromContext` y el branch
     `reqLogger` de `logError`, dejar el fallback `request_id` como mecanismo único y actualizar
     `arquitectura.md` §2.1/§6 para que describan la cadena real (el `documentador` lo hace al
     cerrar F1).

### M3 · `Recover`/`WriteError` conviven con una promesa de `ux.md` incumplida en el layout: `<nav>` con enlace (desviación declarada que no quedó reflejada en el diseño)

- **Evidencia**: `frontend/src/app/layout.tsx:12-19` renderiza un `<nav aria-label="Navegación
  principal">` con el enlace "Inicio".
- **Regla rota**: `ux.md` §1 (`ux.md:24`: "Página única, pública, **sin navegación**") y §2
  (`ux.md:62`: "**Acciones disponibles**: exactamente una — volver a consultar. **No hay
  navegación, enlaces**, formularios ni autenticación"). Constitución §I: el comportamiento se
  refleja primero en el documento que manda; aquí el código contradice el diseño aprobado sin que
  el diseño lo recoja.
- **Dictamen sobre la desviación declarada** ("layout con `<nav>` que `ux.md` decía evitar"):
  técnicamente **aceptable** (es un shell de aplicación semántico, accesible y barato para F2+; la
  skill `react-frontend` pide semántica `nav`), pero **queda mal documentada**: dos textos
  normativos de `ux.md` afirman que no existe navegación.
- **Propuesta**: no hace falta quitar el `<nav>`; sí dejar rastro. Añadir en `ux.md` (§1/§2) el
  mismo tipo de "Ajuste" que el plan ya usó ("Ajustes en ux.md"): "en F1 existe un encabezado con
  navegación mínima (enlace Inicio) fuera del área de estado; las acciones del área de estado
  siguen siendo exactamente una". Si producto prefiere el literal, alternativa: quitar el enlace y
  dejar el `<header>` sin `<nav>`.

---

## Menores

### N1 · `var RequestID httpserver.Middleware = func(...)` es una variable de paquete mutable

- **Evidencia**: `backend/internal/platform/middleware/requestid.go:27`.
- **Regla**: `arquitectura.md` R6 ("Sin estado global — ni **variables de paquete mutables**…").
  Nadie la reasigna, pero la letra de R6 se cumple declarándola como función.
- **Propuesta**: `func RequestID(next http.Handler) http.Handler { … }` — sigue pasándose como
  `httpserver.Middleware` por asignabilidad de funciones. (Es la única variable de paquete mutable
  no-test de todo `backend/`: verificado con `grep "^var "`.)

### N2 · La cadena de middlewares está definida dos veces

- **Evidencia**: `backend/cmd/api/main.go:68-75` (`middlewareChain`) y
  `backend/internal/platform/testutil/http.go:17-24` (`Chain`) — misma lista y mismo orden.
- **Regla**: SC-006/FR-011 (espíritu): las capacidades transversales viven una sola vez. Si F2
  añade `rate-limit`, hay que recordar dos sitios; el de `testutil` roto enmascararía fallos de la
  cadena real en las suites.
- **Propuesta**: extraer la composición a un único constructor (p. ej.
  `middleware.Chain(logger *slog.Logger, origins []string) []httpserver.Middleware`) y usarlo desde
  `main.go` y `testutil`. `platform` puede definirlo sin conocer dominios (R1 se respeta).

### N3 · `NewService` devuelve `*service` en vez de la interfaz `Service` del consumidor

- **Evidencia**: `backend/internal/status/service.go:39` — `func NewService(repo Repository) *service`.
- **Regla**: R3 se cumple (la interfaz `Service` la define `handler.go`, su consumidor), pero el
  patrón de referencia de `arquitectura.md` §5.5 (`NewService(repo Repository) Service`) devuelve
  la interfaz; devolver el tipo no exportado choca con la convención de godoc/golint y con el
  "patrón que copiarán F2–F9" (plan, *Structure Decision*).
- **Propuesta**: `func NewService(repo Repository) Service` (o mover la interfaz al sitio que se
  quiera como canónico y actualizar §5.5; lo importante es que el patrón de referencia y el código
  coincidan).

### N4 · `backend/Dockerfile:29` usa `:latest` mientras su propio comentario prohíbe `:latest`

- **Evidencia**: `backend/Dockerfile:9-10` ("La etiqueta va fijada (nada de `:latest`) para builds
  reproducibles") frente a `FROM gcr.io/distroless/static-debian12:latest` (línea 29). El frontend
  sí fija (`node:22.22-alpine`, `nginx:1.30.5-alpine`, `frontend/Dockerfile:64,85`).
- **Regla**: reproducibilidad (plan D5/"Métricas del plan" SC-010: toda versión en uso debe poder
  identificarse y rastrearse; un tag flotante no se puede listar en la tabla de versiones del
  README).
- **Propuesta**: fijar por digest o por tag con fecha (`gcr.io/distroless/static-debian12:nonroot-YYYYMMDD`
  o `@sha256:…`) y añadir la imagen a la tabla de versiones del `README.md`.

### N5 · `WithTx` revierte con un `ctx` posiblemente cancelado

- **Evidencia**: `backend/internal/platform/database/database.go:70` —
  `defer func() { _ = tx.Rollback(ctx) }()`.
- **Regla**: R5 (respetar la cancelación) leído en sentido inverso: si `fn` falló *por*
  cancelación, el `Rollback` con ese mismo `ctx` no llega a mandarse y la transacción se resuelve
  por cierre de conexión, no por diseño.
- **Propuesta**: `defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()` (Go ≥1.21) o un
  `context.Background()` con timeout corto para el rollback.

### N6 · Identificadores en español en el frontend frente a constitución §V ("nombres en inglés")

- **Evidencia**: `frontend/src/features/status/estado.ts:6` (`EstadoSistema`),
  `toEstadoSistema.ts:24` (`toEstadoSistema`), `useSystemStatus.ts:53-58` (`estado`,
  `fechaConsulta`, `consultando`), archivos `estado.ts` / `toEstadoSistema.ts`.
- **Regla**: constitución §V ("Nombres descriptivos **en inglés** dentro del código", y la regla
  de revisión "nombres en inglés en el código"). El backend sí cumple al 100 %.
- **Matiz**: `arquitectura.md` §5.10 (aprobado) muestra nombres en español en el ejemplo de
  frontend (`ContactoForm`, `useEnviarContacto`, `AdminContactosPage`), así que la práctica actual
  es coherente con el documento fundacional y **no lo trato como bloqueante**.
- **Propuesta**: zanjar la ambigüedad en un solo lugar (una línea en
  `arquitectura.md` §2.2: "identificadores TS en inglés salvo los nombres de dominio público" o
  "identificadores TS en español/inglés según `ux.md`"), y aplicar lo decidido a partir de F2. No
  renombrar nada en F1 salvo que la dirección técnica diga lo contrario.

### N7 · Menores de coherencia documental (sin impacto funcional)

- `backend/api/openapi.yaml`: las respuestas `NotFound`/`MethodNotAllowed` (`components/responses`,
  líneas ~170-183) están definidas pero **no referenciadas** desde el path item de `/healthz`, que
  es justo donde el fallback puede emitirlas (solo aparecen en la tabla de la descripción). Para
  deltas futuros, referenciarlas en cada operación.
- `.env.example` documenta `DATABASE_URL` con la contraseña `cambiar_esto`, mientras el default de
  desarrollo real es `app_dev_password` (`config.go:42`, `docker-compose.yml:133`). Es un nudge
  razonable, pero T004 pedía documentar "el valor por defecto de desarrollo"; una nota al lado lo
  resuelve.
- `go.mod` con `go 1.27` antes que `module` (desviación declarada, `estado.md`): ver dictamen.

---

## Desviaciones declaradas por los desarrolladores — dictamen

| # | Desviación | Dictamen | Fundamento |
|---|---|---|---|
| 1 | Orden `go`/`module` en `go.mod` | ✅ **Aceptable** | Los toolchains de Go no lo exigen ni lo reordenan de forma conflictiva; `go mod verify/build/test` en verde. Aun así, normalizarlo (`module` primero) en la próxima pasada de `go.mod` cuesta cero y elimina la nota. |
| 2 | `Config` sin variables obligatorias | ✅ **Aceptable** | Es literalmente lo que pide la spec: FR-001 (clon limpio sin `.env`) y el caso límite "Falta una variable de entorno: el entorno local funciona con los valores de ejemplo". Los valores inválidos sí fallan rápido con mensaje que identifica la variable (`config.go:109,126,141`) — T007 cubierto. |
| 3 | `RequestID` no crea el logger (lo hace `Logging`) | ⚠️ **Aceptable la decisión, no el estado actual** | Hoy *nadie* publica el logger en el contexto → M2. |
| 4 | `Recover` registrando el stack | ✅ **Aceptable** | FR-013 prohíbe la información interna **en la respuesta**, no en el log; `envelope_test.go:177-197` prueba que el cuerpo no filtra y que el proceso sigue vivo. |
| 5 | `sqlc-verify` omite la regeneración sin consultas | ✅ **Aceptable** | Comportamiento verificado y documentado en `sqlc.yaml:1-18` y `proyecto.mk:16-22`; en cuanto exista `*.sql` regenera y exige `git diff --exit-code` (R4/R20 se cumple desde el primer uso real). |
| 6 | Tailwind v4 y TypeScript 5.9 | ✅ **Aceptable** | El plan fija "TypeScript 5.x" y no fija versión de Tailwind; son las versiones mayores vigentes y `lint/typecheck/test/build` en verde. |
| 7 | `.prettierignore` en la raíz | ✅ **Aceptable** | Cubre rutas del frontend y protege `schema.d.ts` del formateo; existe además `frontend/.prettierignore`. Riesgo nulo; si acaso, consolidar en uno solo en F2. |
| 8 | Layout con `<nav>` que `ux.md` evitaba | ⚠️ **Aceptable, con documentación pendiente** | Ver M3: el diseño aprobado dice textualmente que no hay navegación; el ajuste debe reflejarse en `ux.md`. |
| 9 | Traducción de estados en la feature (no en `getSystemStatus`) | ✅ **Aceptable** | El invariante de `ux.md` §3.1 se conserva: la página (`StatusPage`) solo ve `EstadoSistema`; nunca códigos HTTP ni `error.code`. El único punto de entrada sigue siendo `api/` + el hook. Sugiero que `documentador` ajuste la frase de `ux.md` §3.1/§4 que atribuye la traducción a `getSystemStatus`, para que el diseño describa el código. |
| 10 | `nginx` como root en la imagen del frontend | ✅ **Aceptable en F1** | Comportamiento estándar de `nginx:alpine`, alcance local, y el backend sí corre no-root (distroless `nonroot`, `backend/Dockerfile:46`). Deja propuesta para endurecimiento futuro (nginx unprivileged en 8080 + `user` directive) cuando haya entorno desplegado. |
| 11 | Commit extra `c43f000` (`fix(status)`) | ✅ **Aceptable** | Fix correcto y necesario (timeout del ping → 503 `database_unavailable`, no 500 interno), con prueba que lo cubre (`service_test.go:54-59`, "el ping agota su timeout"). Conventional Commits respetado; un commit adicional de fix no rompe la regla "un commit por tarea". |
| 12 | `RequestID` como `var` (no detectado por los desarrolladores, lo añado) | ⚠️ **Menor** | Ver N1. |

---

## Coherencia con `spec.md` (fase `analyze`)

- **Cobertura FR-001…FR-016**: completa y sin invención de alcance. Cada FR tiene dueño en
  `plan.md` ("Cobertura de requisitos") y se verifica en el código: FR-001/009 (compose + README),
  FR-002/003/004 (dominio `status`, ping por petición con timeout — `database.go:20`,
  `status/repository.go:28`), FR-005 (página + hook), FR-006/007 (CI del kit intacto — verificado
  que `.github/workflows/ci.yml` y `Makefile` no aparecen en el diff de la rama), FR-008
  (`.env.example` sin secretos; `.env` sin versionar), FR-010/011 (`internal/platform/` completo;
  `grep` de reimplantaciones: config/logging/errores/acceso a datos viven solo en `platform/`),
  FR-012/013 (sobres; suite del sobre), FR-014/015 (receta `arquitectura.md` §8 + checklist; la
  verificación SC-007 es T030, humana y **pendiente por diseño**), FR-016 (tabla de versiones y
  registro de pendientes en README/plan).
- **SC-008/SC-009 — los 7 casos de la suite**: cubiertos de verdad en
  `backend/internal/platform/httpserver/envelope_test.go` sobre el stack completo (router + cadena
  + handler real de `status`): (1) 200 DTO directo (`:69`), (2) 503 con sobre + `details` (`:86`),
  (3) 404 sobre, no texto stdlib (`:104`), (4) `POST /healthz` → 405 sobre (`:122`), (5) error
  inesperado → 500 genérico sin fugas (chequea `"sql"`, `"db-interna"`, `"repository.go"`, `".go"`,
  `"goroutine"`, `"panic"`, `"stack"` en el cuerpo — `:138-155`), (6) `panic` → 500 + proceso vivo
  (`:177`), (7) log con el detalle interno **y** `request_id` == `X-Request-ID` de la respuesta
  (`:157-172`).
- **Alcance no inventado**: la única pieza que la spec no pedía literalmente —el párrafo "sitio en
  construcción"— ya venía aprobado en `ux.md` §2/§6, y el `<nav>` está tratado en M3. No hay
  tablas de negocio, ni `validate`/`paginate`/`migrate`/auth (diferidos D23), ni dependencias
  nuevas sin justificar (`go.mod` = solo `pgx/v5` en runtime; `package.json` = las del plan D14–D17).
- **Arquitectura (R1–R8)**: `cmd → dominio → platform` verificado por imports; `platform` no
  importa dominios salvo en el paquete de prueba externo `httpserver_test`
  (`envelope_test.go:22`, con justificación escrita y acordada en plan §"Estrategia de pruebas");
  sin `init()` con lógica; interfaces definidas por el consumidor (`handler.go:13` `Service`,
  `service.go:25` `Repository`, con el matiz N3); `status/service.go` no importa `net/http` ni
  `pgx` (R4); errores con `%w` (`service.go:51,56`, `database.go:29,34,56,65,77`);
  traducción a HTTP **solo** en `WriteError` — único `http.Status*` fuera de `platform` es el 200
  de éxito de `handler.go:45` vía `WriteJSON`, tal como el patrón §5.7; `Registrar` según §4 con
  su excepción declarada §1.3.

---

## Resumen para el ciclo de corrección

| ID | Hallazgo | Archivo:línea | Arreglable en |
|---|---|---|---|
| B1 | `make lint`/`make ci` en rojo (QF1011) | `backend/internal/platform/httpserver/server.go:88` | 1 minuto |
| M1 | "Última consulta" ausente/anticuada en errores | `frontend/src/features/status/hooks/useSystemStatus.ts:36-37` + tests | < 1 hora |
| M2 | Logger por petición sin productor (código muerto) | `backend/internal/platform/httpserver/error.go:36-46` | < 1 hora |
| M3 | `<nav>` sin reflejar en `ux.md` | `frontend/src/app/layout.tsx:12-19` + `ux.md` §1/§2 | < 30 min (documental) |
| N1–N7 | Ver sección "Menores" | — | A criterio; N1/N2/N3 conviene hacerlas ahora (son el patrón de F2–F9) |

Ciclo de corrección **1/3**. Con B1 + M1 + M2 + M3 resueltos (y N1–N3 si se quiere dejar el patrón
impecable para F2), el veredicto pasa a **APROBADO** sin necesidad de reabrir decisiones de diseño.
