# Checklist: verificación de la receta de áreas de negocio (SC-007)

**Purpose**: plantilla para registrar la verificación de la receta (US5, FR-014, FR-015, SC-007). Se rellena **durante** el ejercicio de `quickstart.md` §9 y se firma al final (T030). Esta plantilla se entrega **en blanco**: ningún ítem marcado ni evidencia adjunta.
**Created**: 2026-10-03
**Feature**: [spec.md](../spec.md) · [quickstart.md §9](../quickstart.md) · [plan.md D21/R13](../plan.md) · receta en [arquitectura.md §8](../../../docs/tecnico/arquitectura.md)

> **Quién lo rellena**: el humano que ejecuta el ejercicio (confirmación 3), siguiendo la receta **solo** desde `docs/tecnico/arquitectura.md` §8. El *Independent Test* de **US5** pide que sea una persona del equipo que **no** participó en la creación de la receta y sin consultar decisiones de arquitectura; si no hay otra persona disponible se aplica la válvula de escape **R13** de `plan.md` (validación humana explícita) y queda registrado en la sección 0.

---

## 0. Datos de la verificación (evidencia: quién y cuándo)

| Campo | Valor |
|---|---|
| Quién verifica (nombre / rol) | _(por rellenar)_ |
| ¿Participó en la creación de la receta? (US5 *Independent Test*) | _(sí / no — si es «sí», justificar en la sección 9, R13)_ |
| Fecha de inicio | _(por rellenar)_ |
| Fecha de cierre | _(por rellenar)_ |
| Fuente seguida | _(por rellenar; debe ser solo `docs/tecnico/arquitectura.md` §8)_ |
| Rama de práctica | _(por rellenar; sugerida: `practica/receta-001`)_ |
| Área de práctica (dominio / endpoint) | _(por rellenar; ejemplo sugerido: `muestra` → `GET /api/v1/muestra`)_ |

## 1. Preparación — rama (`quickstart.md` §9, paso 1)

- [ ] Rama de práctica creada desde `main` sin cambios pendientes (`git checkout -b practica/receta-001` o la indicada arriba).

## 2. Los 10 pasos de la receta (`arquitectura.md` §8) — sin decisiones de arquitectura (FR-014)

Marca cada paso cuando quede hecho **siguiendo solo la receta**; si en algún paso hizo falta decidir algo no indicado, anótalo en la sección 10.

- [ ] **1. Contrato primero.** Delta redactado en `specs/…/contracts/` y fusionado en `backend/api/openapi.yaml` **antes** de escribir código; tipos del frontend regenerados (`npm run api:gen`) si el contrato cambió.
- [ ] **2. Migración nueva** `backend/migrations/00000N_create_<tabla>.up.sql` + `.down.sql` completo (nunca se edita una migración aplicada) y `make db-migrate` ejecutado.
- [ ] **3. Consultas sqlc** en `internal/db/queries/<dominio>.sql`, parametrizadas y sin `SELECT *`, y `make sqlc-gen` ejecutado (código generado commiteado).
- [ ] **4. `model.go`**: entidad, estados y DTOs de entrada/salida con validaciones y límites espejo del contrato.
- [ ] **5. `repository.go`**: implementación sobre `internal/db` + `pgxpool`, errores envueltos con `%w`, integridad conocida → `apperr.Conflict`.
- [ ] **6. `service.go`**: interfaz `Repository` (la define quien consume) + reglas de negocio; sin HTTP ni SQL; reloj inyectado si hay fechas.
- [ ] **7. `handler.go`**: interfaz `Service` (la define quien consume) + decodificar, validar, delegar y responder; **todos** los errores por `httpserver.WriteError`, ningún código escrito a mano.
- [ ] **8. `routes.go`**: `RegisterPublic(...)` usado (la variante `RegisterAdmin` **no aplica** en este ejercicio: rutas de panel dependen de `authn`/`authz`, que llegan en F2 — límite declarado D21).
- [ ] **9. Cableado en `cmd/api/main.go`**: `NewRepository(pool)` → `NewService(repo)` → `NewHandler(svc, logger)` → `RegisterPublic`.
- [ ] **10. Pruebas de las tres capas**: service con fake, handler con `httptest`, repository con `//go:build integration` (sin UI en este ejercicio) y `make ci` en verde al terminar.

## 3. Artefactos regenerados (`quickstart.md` §9, paso 3)

- [ ] `make sqlc-gen` ejecutado tras las consultas (`internal/db/` regenerado y commiteado en la rama).
- [ ] `make api-gen` ejecutado si el contrato cambió _(no aplica / aplicado — tachar lo que no corresponda)_.
- [ ] Deriva de artefactos descartada: `make sqlc-verify` termina en verde (incluido en `make ci`).

## 4. Validaciones automáticas y aislamiento (SC-007, FR-015 — `quickstart.md` §9, pasos 4–5)

- [ ] **`make ci` completo en verde** al terminar la rama → evidencia **A** (sección 6).
- [ ] **`git diff main --stat`** muestra **solo archivos nuevos** del área de práctica y la línea de cableado en `cmd/api/main.go` → evidencia **B** (sección 6).
- [ ] **Ninguna área existente modificada ni con cambio de comportamiento** (`status/`, `platform/` intactos; cero archivos de esas áreas en el diff) → evidencia **B**.
- [ ] El verde de `make ci` se logró **sin ajustar las áreas existentes** (US5 esc. 3).

## 5. Verificación funcional (`quickstart.md` §9, paso 6)

- [ ] `make up` levanta el entorno con la rama de práctica.
- [ ] **`GET /api/v1/muestra`** responde con el **sobre de éxito** → evidencia **C**.
- [ ] **`/healthz` sigue funcionando igual** (áreas existentes intactas) → evidencia **C**.
- [ ] El ejercicio incluye sus **pasos de comprobación** para verificar por uno mismo que el área quedó integrada y las demás siguen intactas (US5 esc. 4).

## 6. Evidencias (adjuntar la salida; dejar el bloque vacío si aún no se ejecutó)

**Evidencia A — salida de `make ci`** (mínimo: final del log con la lista de targets en verde, y fecha):

```text
(pendiente de adjuntar)
```

**Evidencia B — `git diff main --stat`** (salida literal, sin editar):

```text
(pendiente de adjuntar)
```

**Evidencia C — resultado funcional** (código y cuerpo de cada respuesta):

```text
GET /api/v1/muestra → (pendiente)
GET /healthz        → (pendiente)
```

## 7. Decisión sobre la rama (`quickstart.md` §9, paso 8)

- [ ] Decisión registrada: **_(descartar / conservar — por rellenar)_**. Por defecto, D21 manda **descartar** (la tabla de práctica nunca llega a `main`).
- [ ] Si se descarta: ejecutado `git checkout main && git branch -D practica/receta-001` y comprobado con `git branch --list 'practica/*'` sin resultados.
- [ ] `main` comprobado **sin tablas de práctica ni código de ejercicio** (F1 cierra con 0 tablas de negocio).

## 8. Cobertura de US5, FR-014, FR-015 y SC-007

| # | Criterio (literal) | Se cubre en | Verificado |
|---|---|---|---|
| 1 | **US5 esc. 1**: se sigue todo el proceso sin decisiones de arquitectura ni consultar a otra persona; cada paso está indicado | Sección 2 (los 10 pasos) + sección 10 | [ ] |
| 2 | **US5 esc. 2**: las áreas existentes no se modifican ni cambian de comportamiento; todo lo agregado pertenece a la nueva unidad | Sección 4 + evidencia **B** | [ ] |
| 3 | **US5 esc. 3**: las validaciones automáticas pasan sin ajustar las áreas existentes | Sección 4 + evidencia **A** | [ ] |
| 4 | **US5 esc. 4**: la receta incluye los pasos para comprobar por uno mismo que el área quedó integrada y las demás están intactas | Sección 5 + evidencia **C** | [ ] |
| 5 | **US5 · Independent Test**: una persona **ajena a la creación de la receta** la sigue de principio a fin, sin consultar decisiones de arquitectura | Sección 0 + sección 9 (R13) | [ ] |
| 6 | **FR-014**: receta documentada, seguible de principio a fin sin decisiones de arquitectura, con pasos de verificación | Secciones 2 y 5 (la fuente es `arquitectura.md` §8) | [ ] |
| 7 | **FR-015**: el área nueva es una unidad aislada; nunca modifica las áreas existentes ni su comportamiento | Sección 4 + evidencia **B** | [ ] |
| 8 | **SC-007**: aplicada al menos una vez antes del cierre de F1, sin modificar áreas existentes y con las validaciones en verde | Secciones 0–7 completas y firmadas | [ ] |

## 9. Válvula de escape R13 (solo si aplica — `plan.md`)

- [ ] **No aplica**: el ejercicio lo hizo una persona del equipo que no participó en la creación de la receta.
- [ ] **Aplica**: no había otra persona disponible; se deja constancia de la **validación humana explícita** del ejercicio.
  - Validado por: _(por rellenar)_ · Fecha: _(por rellenar)_ · Notas: _(por rellenar)_

## 10. Incidencias y decisiones tomadas durante el ejercicio

_(por rellenar; «ninguna» es una respuesta válida. Si hubo alguna, la checklist no se cierra hasta justificarla: F1 exige seguir la receta sin decisiones de arquitectura)_

---

## Firma de cierre

| Campo | Valor |
|---|---|
| Checklist completada por | _(por rellenar)_ |
| Fecha | _(por rellenar)_ |
| SC-007 registrada como cumplida | [ ] |

**Notas de uso**

- Plantilla **sin evidencias precargadas**: los bloques de evidencia dicen «pendiente» hasta que el ejecutor los sustituye con la salida real.
- **Límites declarados** (plan D21): el área de práctica es pública y de solo lectura — sin entradas de usuario (no arrastra `platform/validate`, `rate-limit` ni CSRF, diferidos a F2) y sin rutas de panel (dependen de `authn`/`authz`, F2).
- **F1 no se cierra sin este registro**: SC-007 queda cumplida solo con esta checklist rellena y firmada, antes del cierre de F1.
