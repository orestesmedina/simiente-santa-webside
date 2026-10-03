# QA — Validación de F1 completa (criterios de aceptación y cobertura)

**Fecha:** 2026-10-03 · **Agente:** `qa-tester` · **Alcance:** implementación de F1 (T001–T029). T030–T035 son humanas y quedan fuera de este veredicto.
**Rama:** `001-estructura-base` · **Commits evaluados:** de `6e9ff4b` (inicio de la validación) hasta `d1cbd50`. Nota: durante la sesión otros agentes corrigieron defectos (secuencia `87ed7e2`, `aa60f35`, `1401aa0`, `ef2faca`, `f31c152`, `f20aa8a`, `1c18315`, `234d75b`, `af8a51d`, `d1cbd50`); el sello final de cada hallazgo refleja el árbol **actual** (`d1cbd50`), y para cada uno indico su estado (abierto / corregido en el intervalo).
**Estado del árbol:** yo no modifiqué código de producción. Escribí una prueba auxiliar temporal (`frontend/src/features/status/aux-hora-error.test.tsx`) para reproducir el defecto H1-03, la ejecuté (falla = defecto confirmado) y **la borré**; no queda en el repo. Como acciones de entorno: `ALTER ROLE app PASSWORD 'app_dev_password'` (volumen preexistente con credenciales de una sesión anterior) y `CREATE DATABASE app_test` (para ejecutar las suites de integración contra la BD del kit).

---

## 1. Matriz de cobertura

Leyenda: ✅ verificada y pasa · ⚠️ parcial · ❌ falla. El comando y su salida se citan en la columna de verificación.

### US1 — levantar el entorno con un solo comando

| Criterio | Verificación | Resultado |
|---|---|---|
| Esc. 1: clon limpio → un comando levanta backend+frontend+db | sin `.env`: `mv .env /tmp && make up` — imágenes ya construidas; `docker compose ps`: db (healthy), backend, frontend `Up` (17,4 s) | ✅ |
| Esc. 1 (prerequisito único = Docker) | las imágenes construidas en esta máquina; **sin** `.env` (compose usa defaults documentados) | ✅ |
| Esc. 2: arranque repetible sin estado residual | `docker compose down` (bloqueo de red al recrear) → `docker compose up -d` de nuevo → los tres servicios arriba; los ciclos stop/start de `db` + `make down` final no dejaron residuo | ✅ |
| Esc. 3: persona nueva lo logra solo con la documentación | README (quickstart, puertos, comando, `.env` opcional, tabla de variables) auditable y ejecutable paso a paso; no se simula una persona nueva en esta sesión | ⚠️ |
| Borde: puerto ocupado → fallo identificable + puertos documentados | `HTTP_PORT=8092 docker compose up -d backend` con el puerto ocupado en el host: compose v5.5.1 **respondió exit 0 y "Started"** aunque el binding falló (el puerto quedó sin publicar, `curl` → refused). El error identificable existe (se observó con `db`: `Bind for 0.0.0.0:5432 failed: port is already allocated`), pero en esta prueba el servicio `backend` no lo produjo; los puertos sí están documentados (README) | ⚠️ (ver H1-01) |

### US2 — estado del sistema (SC-002, SC-005)

| Criterio | Verificación | Resultado |
|---|---|---|
| Esc. 1: BD conectada → vivo + conectada | `curl -i http://localhost:8080/healthz` → `HTTP/1.1 200`, `Cache-Control: no-store`, cuerpo `{"status":"ok","database":"connected"}` byte a byte contra contrato/quickstart §1 | ✅ |
| Esc. 2: BD detenida → "no conectada" claramente distinto y el backend sigue respondiendo | `docker compose stop db` && `curl -i` → `HTTP/1.1 503` con `{"error":{"code":"database_unavailable","message":"La base de datos no está conectada","details":{"database":"disconnected"}}}` byte a byte (quickstart §2); 3 repeticiones: 2,0019 / 2,0020 / 2,0037 s | ✅ (nota H1-QT-01) |
| Esc. 3: página inicial muestra el estado sin acciones | `make e2e` (con navegador de la imagen `mcr.microsoft.com/playwright:v1.63.0-jammy`): `la página inicial muestra el estado sin acciones` **pass** (389 ms, aserción con límite 3 s = SC-005) | ✅ (nota H1-QE-01) |
| Esc. 4: estado refleja el presente, no lo memorizado | en vivo: `stop db` → 503; `start db` → 200 **sin reiniciar el backend** (contenedor `StartedAt` 21:28 inalterado). En suite: `TestIntegrationPingReflectsDownAndRecoveryWithoutRestart` (proxy TCP: corta y restaura la BD sin reiniciar) **pass** contra la BD real | ✅ |

### US3 — validación automática (SC-003, SC-004)

| Criterio | Verificación | Resultado |
|---|---|---|
| Esc. 1: pruebas + linters backend y frontend en cada cambio | `make ci` **completa en verde hoy** (1 m 55 s): `gofmt` (sin salida) + `go vet` + `golangci-lint run` (**0 issues**; estaba rojo con QF1011, corregido en `87ed7e2`) + vitest 25/25 + `govulncheck` (**limpio**; GO-2026-5970 por `x/text@v0.29` corregido en `aa60f35`) + `npm audit --audit-level=high` (**0**) | ✅ (local; ver ⚠️) |
| Esc. 2/3: veredicto apto/no apto en PR con motivo visible | workflow `ci.yml` del kit auditable estáticamente (jobs `controles`, `backend`, `frontend`, `secretos`; con motivo en logs `::error::`); el disparo solo en push a `main` requiere T031–T033 (humanas) | ⚠️ |
| `npm run build` | pass (dist estático 298 kB gzip 94 kB) | ✅ |

### US4 — base técnica común (SC-006)

| Criterio | Verificación | Resultado |
|---|---|---|
| Capacidades estándar una sola vez, sin reimplementaciones | búsqueda estructural: `os.Getenv` solo en `platform/config` (excepción: helper de pruebas `testutil/db.go`, aceptada); `pgxpool.New` solo en `platform/database`; handlers de `slog` solo en `platform/logger`; ninguna escritura de error fuera de `httpserver.WriteError`; la cadena `request-id → recover → logging → CORS` vive solo en `platform/middleware` (montada en `cmd/api/main.go` y en `testutil.Chain` con el orden aprobado) | ✅ |
| Cambio de configuración aplica a todos | `platform/config` única puerta; los cuatro consumidores leen de esa canónica (5 variables documentadas idénticas en `.env.example`/README) | ✅ |
| Registro de eventos con mismo formato y una sola cadena | logs JSON observados en vivo: arranque, error, cierre; `slog` JSON + logger-por-petición (hijo); suite `middleware_test.go` (orden instrumentado) pass | ✅ |
| Error inesperado registrado a un mismo modo, en cualquier parte | suite del sobre `envelope_test.go`: handler de prueba `/boom` (fecha interna con sintoma `sql:`), fallos `panic` → 500 con sobre y log con `request_id`; **pass** sobre el stack completo | ✅ |

### US5 — receta de áreas (FR-014, FR-015, SC-007)

| Criterio | Verificación | Resultado |
|---|---|---|
| Receta documentada, sin decisiones de arquitectura | `docs/tecnico/arquitectura.md` §8 (10 pasos concretos, herramientas y rutas nombradas); checklist `checklists/receta.md` (plantilla en blanco, 36 casillas) lista para T030 | ✅ (documento) |
| Verificación de la receta de punta a punta sin romper áreas existentes | T030 es **humana** (fuera de alcance QA). Precondición ejecutable ya satisfecha: `make ci` verde hoy; al abrir la validación no lo era (véase H1-02), lo que habría invalidado T030 según la receta | ⚠️ (cierre humano T030–T035) |

### US6 — formato uniforme y sin filtraciones (SC-008, SC-009)

| Criterio | Verificación | Resultado |
|---|---|---|
| Éxito = DTO directo documentado | en vivo + `envelope_test.go::TestEnvelopeSuccessIsDirectDTO` (byte a byte; `Content-Type` JSON; `Cache-Control: no-store`) | ✅ |
| Error previsto con mensaje comprensible | 503 con sobre exacto (detalles limitados a `database: disconnected`, sin filtraciones); `/ruta-que-no-existe` → 404 y `POST /healthz` → 405 también con sobre exacto; ninguno con el texto plano de la stdlib | ✅ |
| Error inesperado genérico, sin filtraciones | `TestEnvelopeUnexpectedErrorHidesInternals` (cadena sensible **con leitmotiv** `sql:`/`db-interna`/`.go`/`goroutine`/`stack` y trazas) y `TestEnvelopePanicIsRecovered` (panic → 500 con sobre + proceso vivo) — pass | ✅ |
| Dos operaciones distintas comparten el mismo formato de respuestas | 200/503/404/405/500 **producidas en vivo** contra el stack real, todas con el mismo `ErrorEnvelope` o el DTO directo | ✅ |
| Detalle de errores inesperados registrado con `request_id` | en vivo: con `db` caída, log del backend: `{"level":"WARN","msg":"error de dominio","request_id":"00aeb9…","method":"GET","path":"/healthz","code":"database_unavailable","error":"La base de datos no está conectada: ping database: context deadline exceeded"}` (la causa solo en el log, **no** en la respuesta; corregido en `1c18315`) | ✅ |

### US7 — versiones con soporte vigente (FR-016, SC-010)

| Criterio | Verificación | Resultado |
|---|---|---|
| Tecnologías en rango de soporte | runtime verificable: Go 1.27.1 (toolchain local instancia `go.mod`), Node 22.23.3, PostgreSQL 16.4-alpine (`docker compose images`), distroless static, nginx 1.30.5-alpine, pgx v5 | ✅ con registros de **final de ventana** |
| Pendientes de actualización identificados (esc. 2) | README registra los 4 pendientes (PostgreSQL 16.4 con CVEs corregidos en minors posteriores; Node 22 → abril 2027; del kit: la imagen postgres del `ci.yml` y el «Go 1.23+» de `docs/GUIA-INICIO.md`); T034/T035 humanas | ✅ (constancias) + ⚠️ (cierre humano) |
| Dependencias sin vulnerabilidades altas/críticas (constit. §IV) | `govulncheck ./...`: **sin vulnerabilidades** (se corrigió `x/text@v0.39.0`); `npm audit --audit-level=high`: **0** | ✅ |

### Criterios globales SC-001…SC-010

| ID | Criterio | Estado |
|---|---|---|
| SC-001 | persona nueva, entorno en <15 min solo con docs | ⚠️ — procedimiento cubierto; medición incompleta (imágenes ya compiladas en el arranque probado: 17,4 s; sin `.env`; el coste de build en frío no se repite exactamente en esta máquina — WSL sobre `/mnt/d` con warning de `make doctor`) |
| SC-002 | distingue conectada / no conectada 100 % | ✅ (dos direcciones + recuperación; bd real + suites) |
| SC-003 / SC-004 | veredicto apto/no apto, <10 min | ⚠️ local sí (`make ci` = 1 m 55 s, verde); en GitHub depende de T031–T033 |
| SC-005 | página muestra el estado sin acciones en <3 s | ✅ (e2e 389 ms; aserción explícita de límite) |
| SC-006 | capacidades transversales una sola vez | ✅ (búsqueda estructural + suite del sobre) |
| SC-007 | receta seguida de punta a punta con validaciones verdes | ⚠️ humano (T030); hoy `make ci` ya es verde (prerequisito cumplido) |
| SC-008 | 100 % de respuestas en formato uniforme | ✅ (en vivo y suites; sobre en 200/503/404/405/500) |
| SC-009 | 0 filtraciones en errores inesperados, detalle en log | ✅ (en vivo y suites; causa interna registrada con `request_id` en el log del backend) |
| SC-010 | 100 % versiones dentro de soporte | ⚠️ (verificados en rango; los pendientes están documentados; cierre en T034/T035) |

### Criterios no verificables localmente (y por qué)

- **US3 esc. 2–3 / FR-007 / SC-003–SC-004 en GitHub**: no existe remoto (T031 humana) ni protección de rama (T033). Solo se audita el workflow declarado.
- **US5 esc. 1–4 / FR-015 comportamiento / SC-007 de hecho**: la verificación de la receta es ejercicio humano T030. Verifico únicamente que la receta y la checklist existen y están completas.
- **SC-001 end-to-end en máquina limpia**: imposible replicar un clon nuevo + build frío sin Docker en el mismo WSL; sustituto: auditoría crítica de la documentación y arranque real sin `.env`.
- **`make e2e` en el host de esta sesión**: `chrome-headless-shell` recae en `libnspr4.so` ausente; lo ejecuto bajo la imagen oficial Playwright (pass). README ya documenta `npx playwright install chromium` + `sudo npx playwright install-deps chromium` como requisito del host (H1-QE-01).

---

## 2. Hallazgos

Severidades: bloqueante / mayor / menor. Los hallazgos con estado "corregido" fueron cerrados por otros agentes durante la sesión (ref. commit).

### H1-01 (mayor) — `make up` no falla visiblemente cuando un puerto del host está ocupado (servicio `backend`)

- **Pasos:** proceso bloqueando el puerto TCP 8092 del host; `HTTP_PORT=8092 docker compose up -d backend` (Docker Compose v5.5.1 sobre WSL).
- **Esperado (borde US1):** que el arranque falle con un error identificable y que la documentación indique los puertos que usa el proyecto (la documentación sí existe).
- **Obtenido:** exit 0 + estado "Started"; publicación silenciosamente incompleta (curl → refused). El mismo edge en el servicio `db` sí produce el error `Bind for 0.0.0.0:5432 failed: port is already allocated`.
- **Impacto:** un humano puede creer que el backend arrancó cuando el puerto no se publica; es detectable a los segundos con `curl`/`make doctor`… pero el fallo no aparece en el propio comando.
- **Origen:** comportamiento de compose/WSL, no del código del repo. Propuesta: añadir al README un "smoke check" de una línea (`curl -fsS localhost:8080/healthz` tras `make up`). No bloqueante para el circuito F1 (la documentación ya indica los puertos).

### H1-02 (mayor → corregido durante la sesión) — `make ci` rojo (lint + govulncheck)

- **Al inicio de la validación:** `golangci-lint run` fallaba con `QF1011` (`server.go:88`) y `govulncheck` fallaba con **GO-2026-5970** (`golang.org/x/text@v0.29.0`, reachable desde `pgxpool.NewWithConfig`; fix `v0.39.0`).
- **Estado final:** `87ed7e2` (lint: 0 issues) y `aa60f35` (`x/text v0.39.0`; `govulncheck` limpio). **`make ci` completo pasa hoy en 1 m 55 s** (SC-004 local, <10 min).
- **Moraleja operativa:** el CI ejecutaría con las mismas puertas; cualquier PR habría quedado «no apto» hasta corregir ambos puntos. El ciclo demostró su utilidad: con el estado inicial, T030 (SC-007) habría sido imposible; correcto haberse detectado antes de la entrega.

### H1-03 (mayor → corregido durante la sesión) — «hora del intento» ausente en los estados de error

- **Pasos:** con la primera consulta fallida (error B), la página no mostraba la línea «Última consulta: [hora del intento]» que `ux.md` §3.3 exige para el estado *No se pudo consultar* (y por extensión en el error A); en las reconsultas fallidas, la hora mostrada era la del último éxito (FR-003: «nunca un estado memorizado»).
- **Repro:** prueba auxiliar (borrada tras su uso): `expect(screen.getByText(/Última consulta:/))` en el primer intento fallido → **falla** en el árbol previo al arreglo.
- **Corregido en:** `234d75b` (uso de `errorUpdatedAt` para la marca del intento) + aserciones nuevas en `status.test.tsx` (error A y las dos variantes del error B). Suite actual: **25/25 en verde**.

### H1-04 (menor, operativo) — el healthcheck del contenedor `db` no detecta credencial incorrecta; ej. al reutilizar un volumen con otra contraseña

- **Pasos:** volumen `pgdata` inicializado con credenciales de una sesión previa; `make up` → `db` **healthy** (`pg_isready` no autentica); `/healthz` devolvía 503 **indefinidamente** y las credenciales no se autoreparan (`POSTGRES_PASSWORD` solo aplica al primer init).
- **Reproducido en esta sesión** con las credenciales documentadas del repo; riesgo operativo real que se materializó (diagnóstico demorado ~4 min hasta acudir a `psql` TCP).
- **Esperado:** un error identificable en el arranque o la protección del healthcheck — la documentación de recuperación (`docker compose down -v` o `ALTER ROLE`) no existe.
- **Atenuantes:** tras `1c18315` el log del backend **sí** muestra la causa en vivo (`context deadline exceeded`), lo que acorta el diagnóstico; en clon limpio (escenario SC-001) el fallo no se manifiesta.
- **Propuesta:** añadir al README una nota de operación "la BD me dice no conectada tras un cambio de `POSTGRES_PASSWORD`: recrear el volumen o reset con `ALTER ROLE`". No bloquea gates; cae en categoría operativa.

### H1-05 (menor) — `app_test` local no se crea nunca: `make test` con `.env` del ejemplo falla en las pruebas de integración

- **Pasos:** copiar `.env.example` → `.env`; `make test`.
- **Esperado:** verde sin pasos manuales (como sí ocurre en CI: `ci.yml` crea `app_test` con `POSTGRES_DB: app_test`).
- **Obtenido:** `FATAL: database "app_test" does not exist (SQLSTATE 3D000)` (la BD `app` sí existe pero no `app_test`).
- **Mitigación de diseño aceptada:** sin `DATABASE_URL_TEST` las pruebas de integración se **saltan silenciosamente** (documentado en README:63 «sin ella, esos tests se omiten»). Verificado **que con la BD del kit corren de verdad**: `DATABASE_URL_TEST='postgres://app:app_dev_password@localhost:5432/app_test' go test -tags=integration -count=1 ./...` → **9/9 paquetes `ok` en 2026-10-03** (14 pruebas, cero `SKIP`), incluidos el proxy TCP de caída/recuperación y `WithTx` commit/rollback. Riesgo residual: un entorno local que nunca exporta la variable no ejercita los repositories (constitución §III exige repos sobre PostgreSQL real) — **propuesta menor**: que `make test-backend` advierta (no falle) cuando `DATABASE_URL_TEST` esté vacía.

### H1-06 (menor) — «Inicio» (nav) contradiere la página único sin navegación de `ux.md` §2

- **Obtenido:** `frontend/src/app/layout.tsx` incluye un `header/nav` con enlace `Inicio`; `ux.md` dice «Acciones disponibles: exactamente una (– volver a consultar). No hay navegación, enlaces…». `tasks.md` T021 pide un «layout mínimo con `main`/`nav` semánticos» — conflicto entre himnos aprobados.
- **Impacto:** no bloquea; el enlace apunta a `/` (misma página). El diseño visual real llega con F3; recomiendo resolver el conflicto de fuentes (a favor de `tasks.md` o de `ux.md`) antes de que F3 copie el layout.

### H1-QT-01 (menor, observación) — el 503 llega a 2,001–2,004 s (D8 dice «≤2 s»)

- Medición en vivo: 2,0019 / 2,0020 / 2,0037 s (`PingTimeout` = 2 s exactos + coste de serialización de la respuesta). En la redacción del quickstart («~2 s máximo») se cumple; si se exigiera el literal «≤2 s», bastaría reducir el timeout a ~1,9 s en `platform/database` o medir con margen. No lo considero bloqueante ni mayor.

### H1-QE-01 (menor) — `make e2e` en host depende de librerías del sistema del navegador

- `chrome-headless-shell` no arranca en esta máquina sin `sudo npx playwright install-deps chromium` (pasa bajo la imagen oficial `mcr.microsoft.com/playwright:v1.63.0-jammy`); el README ya documenta ese paso (líneas 137–142). No rompe FR-001 (e2e no es parte del arranque único). Sugerencia: copiar esa nota al §7 del quickstart para que no se omita en una máquina nueva.

### Notas sin severidad (informativas)

- `HEAD /healthz` responde 200 (Go auto-registra HEAD en rutas GET): sin cuerpo, sin sobre — coherente con el contrato (solo `GET`). No es hallazgo: comportamiento estándar de `net/http`.
- `//healthz` responde 307 (canonización del `ServeMux`): el cuerpo del redirect es texto de la stdlib, no un sobre; tras seguir el redirect la respuesta es 200 con sobre correcto. Cohallazgo menor para F2 si se audita el 100 % de las respuestas (los redirects no son ErrorEnvelope).
- Las 36 casillas de `checklists/receta.md` están vacías (T029 bien hecho: plantilla sin evidencias preconfeccionadas).
- El contrato vivo y el snapshot son idénticos byte a byte (`diff` sin salida) y `npm run api:gen` regenera sin deriva (regla anti-deriva plan R4) — verificado.

---

## 3. Cobertura de pruebas existentes (calidad, sin arrastre)

- `envelope_test.go` (T018, SC-008/SC-009): cubre exactamente lo pedido (DTO directo, 503, fallback 404, `POST /healthz` 405, error inesperado sin filtraciones con cadena sensible realista, panic con proceso vivo y log con `request_id` coincidente con la `X-Request-ID` de la respuesta). Sin huecos relevantes: los agujeros imaginables (rutas no documentadas, `POST /healthz`, panic, filtraciones) están todos cubiertos.
- `status/handler_test.go`: cuerpos esperados byte a byte (idénticos a `contracts/openapi.yaml`/`quickstart.md`), no acoplados a la implementación (usa fakes de Service, no del router).
- `status/service_test.go`: con fake de Repository cubre BD viva, BD caída → apperr correcto, error inesperado envuelto con `%w` y cancelación de ctx. Cobertura del paquete: **87 %** (≥ 80 % exigido por constitución §III). Resto: httpserver 91,9 %; middleware 92,9 %.
- `status/repository_test.go` y `database/database_integration_test.go`: **corren de verdad contra la BD del kit** (no saltan) con `DATABASE_URL_TEST` apuntada; la prueba de caída/recuperación usa un proxy TCP para no tocar el contenedor — diseño fiel al escenario US2 esc. 4 y tolerante a una BD lenta (waitPing de 10 s); los timeouts de Ping (2 s) se verifican con pruebas adicionales (`TestPingTimeoutIsTwoSeconds` + `TestIntegrationPingUnreachableHost`; en vivo: ~2,00 s).
- Nada de pruebas acomodadas a la implementación: los cuerpos del sobre provienen del contrato aprobado y del quickstart (no de lo que produce el código); no detecté contradicciones con lo que `ux.md` hubiera pedido. Al contrario: detecté un desvío real (H1-03) con una prueba auxiliar (ejecutada y borrada; se explica arriba), corregida después por el desarrollador en `234d75b` con el test apropiado.

## 4. Veredicto global: **APROBADO CON OBSERVACIONES**

**Resumen:** F1 cumple los criterios de aceptación verificados en local: los sobres de respuesta (éxito = DTO, errores uniformes sin filtraciones), el estado real de la BD (caída/recuperación sin reinicio, ≤2 s), el estado en página en <3 s, la infraestructura común única y transparente (0 reimplementaciones), las pruebas realistas de integración contra PostgreSQL y las puertas de calidad (gofmt/vet/golangci/govulncheck/npm audit/compilación/typecheck/tests unitarios e integración/e2e en verde hoy). El ciclo de corrección (10 commits) cerró los 2 problemas mayores detectados al abrirse la validación (H1-02, H1-03), demostrando que el bucle de revisión funciona.

**Condiciones para el cierre** (no bloquean este QA, pero obligan antes de integrar):
1. Confirmar la resolución humana de T030–T035 (hay que crear el remoto y la protección de rama: sin eso, el pipeline no protege nada — H1-04/H1-05 quedan sin runbook).
2. Encargar al orquestador: resolver el conflicto `ux.md` ↔ `tasks.md` del layout (H1-06) y añadir al README las dos notas de operación (H1-01: smoke check tras `make up`; H1-04: recuperación de credenciales de BD).
3. Evidenciar en la checklist de cierre que `make test` local se corre con `DATABASE_URL_TEST` activa (sello de esta validación: 9/9 paquetes `ok`, 0 `SKIP`).
