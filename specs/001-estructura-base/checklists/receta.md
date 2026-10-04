# Checklist: verificación de la receta de áreas de negocio (SC-007)

**Purpose**: plantilla para registrar la verificación de la receta (US5, FR-014, FR-015, SC-007). Se rellena **durante** el ejercicio de `quickstart.md` §9 y se firma al final (T030). Esta plantilla se entrega **en blanco**: ningún ítem marcado ni evidencia adjunta.
**Created**: 2026-10-03
**Feature**: [spec.md](../spec.md) · [quickstart.md §9](../quickstart.md) · [plan.md D21/R13](../plan.md) · receta en [arquitectura.md §8](../../../docs/tecnico/arquitectura.md)

> **Quién lo rellena**: el humano que ejecuta el ejercicio (confirmación 3), siguiendo la receta **solo** desde `docs/tecnico/arquitectura.md` §8. El *Independent Test* de **US5** pide que sea una persona del equipo que **no** participó en la creación de la receta y sin consultar decisiones de arquitectura; si no hay otra persona disponible se aplica la válvula de escape **R13** de `plan.md` (validación humana explícita) y queda registrado en la sección 0.

---

## 0. Datos de la verificación (evidencia: quién y cuándo)

| Campo | Valor |
|---|---|
| Quién verifica (nombre / rol) | **omedina** (cliente / product owner) |
| ¿Participó en la creación de la receta? (US5 *Independent Test*) | **No** (la escribió el equipo de agentes). No hay otra persona del equipo que sea desarrolladora; se aplica **R13** (sección 9) |
| Fecha de inicio | 2026-10-03 |
| Fecha de cierre | 2026-10-03 |
| Fuente seguida | **Solo** `docs/tecnico/arquitectura.md` §8 + §8.1 (versión tras `f81470f` y `28e4491`) y `quickstart.md` §9 |
| Rama de práctica | `practica/receta-001` (base `001-estructura-base`, porque F1 aún no está fusionada a `main`); **descartada** |
| Área de práctica (dominio / endpoint) | `muestra` → `GET /api/v1/muestra` (tabla `sample_items`) |

> **Nota de adaptación**: el escenario de `quickstart.md` §9 dice «rama desde `main`» y `git diff main --stat`. Como F1 todavía no se ha fusionado a `main`, la rama de práctica se creó desde `001-estructura-base` y el chequeo de aislamiento se hizo con `git diff 001-estructura-base --stat`. Es equivalente una vez fusionado el PR de F1.

## 1. Preparación — rama (`quickstart.md` §9, paso 1)

- [X] Rama de práctica creada desde `001-estructura-base` sin cambios pendientes (`git checkout -b practica/receta-001`).

## 2. Los 10 pasos de la receta (`arquitectura.md` §8) — sin decisiones de arquitectura (FR-014)

Marca cada paso cuando quede hecho **siguiendo solo la receta**; si en algún paso hizo falta decidir algo no indicado, anótalo en la sección 10.

- [X] **1. Contrato primero.** Delta redactado y fusionado en el contrato vivo **antes** de escribir código; tipos del frontend regenerados (`make api-gen`).
- [X] **2. Migración nueva** `backend/migrations/000002_create_sample_items.up.sql` + `.down.sql` completo y `make db-migrate` ejecutado.
- [X] **3. Consultas sqlc** en `internal/db/queries/muestra.sql`, parametrizadas y sin `SELECT *`, y `make sqlc-gen` ejecutado (código generado commiteado).
- [X] **4. `model.go`**: entidad, estados y DTOs de entrada/salida con validaciones y límites espejo del contrato.
- [X] **5. `repository.go`**: implementación sobre `internal/db` + `pgxpool`, errores envueltos con `%w`, conversión `pgtype`→dominio en `mapRow` (§8.1 punto 6).
- [X] **6. `service.go`**: interfaz `Repository` (la define quien consume) + reglas de negocio; sin HTTP ni SQL; acotado `defaultListLimit = 20` (§8.1 punto 3).
- [X] **7. `handler.go`**: interfaz `Service` (la define quien consume) + decodificar, validar, delegar y responder; **todos** los errores por `httpserver.WriteError`.
- [X] **8. `routes.go`**: `RegisterPublic(...)` (solo esa superficie: el área no tiene panel, §8.1 punto 4).
- [X] **9. Cableado en `cmd/api/main.go`**: `NewRepository(pool)` → `NewService(repo)` → `NewHandler(svc, logger)` → `RegisterPublic`.
- [X] **10. Pruebas de las tres capas**: service con fake, handler con `httptest`, repository con `//go:build integration`, prueba de humo del cableado y `make ci` en verde.

## 3. Artefactos regenerados (`quickstart.md` §9, paso 3)

- [X] `make sqlc-gen` ejecutado tras las consultas (`internal/db/` regenerado y commiteado en la rama).
- [X] `make api-gen` ejecutado (el contrato cambió: se añadió la operación).
- [X] Deriva de artefactos descartada: `make sqlc-verify` termina en verde (incluido en `make ci`).

## 4. Validaciones automáticas y aislamiento (SC-007, FR-015 — `quickstart.md` §9, pasos 4–5)

- [X] **`make ci` completo en verde** al terminar la rama → evidencia **A** (sección 6).
- [X] **`git diff 001-estructura-base --stat`** muestra solo archivos nuevos del área de práctica, el delta aditivo del contrato, los artefactos regenerados y el cableado en `cmd/api/main.go` → evidencia **B**.
- [X] **Ninguna área existente modificada ni con cambio de comportamiento**: `git diff 001-estructura-base -- backend/internal/status backend/internal/platform` sale **vacío** → evidencia **B**.
- [X] El verde de `make ci` se logró **sin ajustar las áreas existentes** (US5 esc. 3).

## 5. Verificación funcional (`quickstart.md` §9, paso 6)

- [X] `docker compose up -d --build` levanta el entorno con la rama de práctica.
- [X] **`GET /api/v1/muestra`** responde `200` con el **sobre de éxito** (`{"items":[]}`; nunca `null`) → evidencia **C**.
- [X] **`/healthz` sigue funcionando igual** (`200 {"status":"ok","database":"connected"}`) → evidencia **C**.
- [X] El ejercicio incluye sus **pasos de comprobación** (ruta nueva, `/healthz` y una ruta inexistente `404` con `ErrorEnvelope`) (US5 esc. 4).

## 6. Evidencias (adjuntar la salida; dejar el bloque vacío si aún no se ejecutó)

**Evidencia A — salida de `make ci`** (ejecutado por el orquestador sobre la corrida final, 2026-10-03):

```text
gofmt -l .            → sin salida
go vet ./...          → ok
golangci-lint run     → 0 issues.
go test ./...         → ok en todos los paquetes (incluye internal/muestra y cmd/api)
go test -tags=integration ./...  → ok simiente-santa/backend/internal/muestra (PostgreSQL real; app_test)
frontend: eslint ok · tsc --noEmit ok · vitest --run → 6 archivos / 25 pruebas OK
govulncheck ./...     → No vulnerabilities found.
npm audit --audit-level=high → found 0 vulnerabilities
cobertura service.go  → 100 % (requisito ≥80 % en service/)
make sqlc-verify     → sin deriva
```

**Evidencia B — `git diff 001-estructura-base --stat`** (salida literal, sin editar):

```text
 backend/api/openapi.yaml                           | 70 +++++++++++++++-
 backend/cmd/api/main.go                            | 15 ++--
 backend/cmd/api/main_test.go                       | 81 +++++++++++++++++-
 backend/go.mod                                     |  5 +-
 backend/go.sum                                     |  2 +
 backend/internal/db/db.go                          | 32 ++++++++
 backend/internal/db/models.go                      | 17 ++++
 backend/internal/db/muestra.sql.go                 | 52 ++++++++++++++++
 backend/internal/db/queries/muestra.sql            |  9 ++
 backend/internal/muestra/handler.go                | 58 +++++++++++++
 backend/internal/muestra/handler_test.go           | 96 ++++++++++++++++++++++
 backend/internal/muestra/model.go                  | 43 +++++++++++++
 backend/internal/muestra/repository.go             | 55 +++++++++++++
 backend/internal/muestra/repository_test.go        | 82 +++++++++++++++++++++
 backend/internal/muestra/routes.go                 | 18 ++++
 backend/internal/muestra/service.go                | 59 +++++++++++++
 backend/internal/muestra/service_test.go           | 95 +++++++++++++++++++++
 backend/migrations/000002_create_sample_items.down.sql | 2 +
 backend/migrations/000002_create_sample_items.up.sql   | 14 ++++
 frontend/src/api/schema.d.ts                       | 75 +++++++++++++++++
 20 files changed, 869 insertions(+), 11 deletions(-)

# Aislamiento (debe estar vacío): git diff 001-estructura-base -- backend/internal/status backend/internal/platform
(vacío)
```

**Evidencia C — resultado funcional** (código y cuerpo de cada respuesta):

```text
GET /api/v1/muestra        → HTTP 200 · {"items":[]}
GET /healthz               → HTTP 200 · {"status":"ok","database":"connected"}
GET /ruta-que-no-existe    → HTTP 404 · {"error":{"code":"not_found","message":"Recurso no encontrado"}}
POST /api/v1/muestra       → HTTP 405 · {"error":{"code":"method_not_allowed","message":"Método no permitido"}}
```

## 7. Decisión sobre la rama (`quickstart.md` §9, paso 8)

- [X] Decisión registrada: **descartar** (D21: la tabla de práctica nunca llega a `main`).
- [X] Ejecutado `git checkout 001-estructura-base && git branch -D practica/receta-001`; `git branch --list 'practica/*'` sin resultados.
- [X] `main`/`001-estructura-base` comprobados **sin tablas de práctica ni código de ejercicio**; además se limpió la **BD local** (`app` y `app_test` de vuelta a la versión 1, solo `schema_migrations`). F1 cierra con 0 tablas de negocio.

## 8. Cobertura de US5, FR-014, FR-015 y SC-007

| # | Criterio (literal) | Se cubre en | Verificado |
|---|---|---|---|
| 1 | **US5 esc. 1**: se sigue todo el proceso sin decisiones de arquitectura ni consultar a otra persona; cada paso está indicado | Sección 2 (los 10 pasos) + sección 10 | [X] (tras corregir la receta; ver sección 10) |
| 2 | **US5 esc. 2**: las áreas existentes no se modifican ni cambian de comportamiento; todo lo agregado pertenece a la nueva unidad | Sección 4 + evidencia **B** | [X] |
| 3 | **US5 esc. 3**: las validaciones automáticas pasan sin ajustar las áreas existentes | Sección 4 + evidencia **A** | [X] |
| 4 | **US5 esc. 4**: la receta incluye los pasos para comprobar por uno mismo que el área quedó integrada y las demás están intactas | Sección 5 + evidencia **C** | [X] |
| 5 | **US5 · Independent Test**: una persona **ajena a la creación de la receta** la sigue de principio a fin, sin consultar decisiones de arquitectura | Sección 0 + sección 9 (R13) | [X] (R13: validación humana explícita; no había persona desarrolladora ajena disponible) |
| 6 | **FR-014**: receta documentada, seguible de principio a fin sin decisiones de arquitectura, con pasos de verificación | Secciones 2 y 5 (fuente: `arquitectura.md` §8/§8.1) | [X] |
| 7 | **FR-015**: el área nueva es una unidad aislada; nunca modifica las áreas existentes ni su comportamiento | Sección 4 + evidencia **B** | [X] |
| 8 | **SC-007**: aplicada al menos una vez antes del cierre de F1, sin modificar áreas existentes y con las validaciones en verde | Secciones 0–7 completas y firmadas | [X] |

## 9. Válvula de escape R13 (solo si aplica — `plan.md`)

- [ ] **No aplica**: el ejercicio lo hizo una persona del equipo que no participó en la creación de la receta.
- [X] **Aplica**: no había otra persona disponible; se deja constancia de la **validación humana explícita** del ejercicio.
  - Validado por: **omedina** (cliente / product owner) · Fecha: **2026-10-03** · Notas: el ejercicio lo ejecutó un agente de desarrollo **fresco**, limitado a `arquitectura.md` §8/§8.1 (sin leer los documentos de planificación de F1), en tres corridas: la 1ª destapó 10 decisiones de arquitectura no escritas (sección 10); se completó la receta (`f81470f`); la 2ª destapó un defecto de esa corrección (el tipo UUID real de sqlc, `pgtype.UUID`; corregido en `28e4491`); la **3ª corrida**, con la receta corregida, no requirió ninguna decisión de arquitectura. El humano validó el resultado tras revisar las evidencias.

## 10. Incidencias y decisiones tomadas durante el ejercicio

Hallazgos de las corridas (su resolución es parte de T030; ninguno queda abierto):

- **1ª corrida (receta original) — 10 decisiones no escritas**, resueltas como convenciones en `docs/tecnico/arquitectura.md` §8.1 (`f81470f`): clave primaria y tipos UUID; sobre de éxito de un listado; paginación por defecto; `RegisterAdmin` cuando no hay panel; ubicación del delta del contrato; mapeo `pgtype`→dominio; orden por defecto; operativa del entorno (BD arriba antes de migrar, `--build`, `DATABASE_URL_TEST`); prueba del cableado; metadatos del contrato.
- **2ª corrida — defecto en la corrección**: §8.1/§5.3 afirmaban que sqlc emite `uuid.UUID`; con `sql_package: pgx/v5` (sqlc v1.31.1) emite **`pgtype.UUID`**. Corregido en `28e4491` (la conversión `uuid.UUID(row.ID.Bytes)` va en `repository.go`).
- **3ª corrida (receta corregida) — ninguna decisión de arquitectura**. Restaron solo elecciones normales de desarrollo, que la receta no tiene por qué fijar: las columnas de `sample_items`, los campos del DTO de salida y el nombre del tipo de parámetros interno del repository. No se consideran incumplimiento de FR-014.
- **Residuo de BD entre corridas**: las corridas repetidas dejaron `sample_items` en la BD local; se limpió (`down`/`force`) para dejar el entorno en la versión 1. No afecta al repositorio.

---

## Firma de cierre

| Campo | Valor |
|---|---|
| Checklist completada por | orquestador, a partir de la corrida ejecutada por un agente fresco y de las evidencias verificadas |
| Fecha | 2026-10-03 |
| SC-007 registrada como cumplida | [X] |

**Notas de uso**

- Plantilla **sin evidencias precargadas**: los bloques de evidencia dicen «pendiente» hasta que el ejecutor los sustituye con la salida real.
- **Límites declarados** (plan D21): el área de práctica es pública y de solo lectura — sin entradas de usuario (no arrastra `platform/validate`, `rate-limit` ni CSRF, diferidos a F2) y sin rutas de panel (dependen de `authn`/`authz`, F2).
- **F1 no se cierra sin este registro**: SC-007 queda cumplida solo con esta checklist rellena y firmada, antes del cierre de F1.
