# Propuestas al repositorio del kit

**Origen:** proyecto Simiente Santa (funcionalidad F1, `specs/001-estructura-base/`) · **Fecha:** 2026-10-03 · **Estado:** pendientes de llevar al repositorio del kit

Este documento reúne las mejoras que **no se pueden aplicar aquí** porque tocan archivos del kit (regla 10 de `AGENTS.md`: los archivos listados en `.kit-manifest.json` no se editan en el proyecto). Cada propuesta indica qué, por qué, dónde tocaría y cómo se verificaría; están pensadas para que el equipo del kit decida con criterio, no para aplicarlas a la fuerza.

Índice: [1. Estado derivado y verificable](#1-estado-del-proyecto-derivado-y-verificable) · [2. ID de tarea en el `commit-msg`](#2-el-hook-commit-msg-exige-el-id-de-la-tarea) · [3. Job e2e en el CI](#3-job-de-pruebas-end-to-end-en-el-ci) · [4. Umbral de cobertura](#4-umbral-de-cobertura-mínimo-en-el-ci) · [5. Artefactos generados](#5-verificación-de-artefactos-generados-en-el-ci) · [6. Versiones](#6-versiones-del-kit-en-fin-de-vida-o-con-cves) · [7. Lectura del estado](#7-leer-el-estado-al-iniciar-sesión)

---

## 1. Estado del proyecto derivado y verificable

**Qué.** Añadir al kit un `make estado` que **genere** un `ESTADO.md` del proyecto y un `make estado-verificar` que falle si el archivo está desactualizado. La idea no es nueva en el kit: es el mismo patrón que ya usa para otras cosas.

| Fuente escrita a mano | Archivo generado | Verificador que ya existe |
|---|---|---|
| `equipo/agentes/*.md` | `.claude/`, `.codex/`, `.opencode/` | `make verificar-agentes` |
| instalación del kit | `.kit-manifest.json` | `make verificar-kit` |
| **`equipo/estado/pendientes.yaml`** | **`ESTADO.md`** | **`make estado-verificar`** ← propuesta |

**Por qué.** Hoy el estado del proyecto es lo único que se escribe a mano, y por tanto lo único que puede quedar mintiendo: un dato derivado **no puede desviarse**, y el verificador convierte el olvido en un error visible en vez de en silencio. Con 9 funcionalidades por delante y varias sesiones por funcionalidad, el coste de no tenerlo se multiplica.

**Qué se deriva (y de dónde).**

| Dato | Fuente |
|---|---|
| Rama, árbol limpio, últimos commits | `git` |
| Fase del proceso (spec → plan → tareas → código → validación → PR) | Existencia y campo `Status` de los artefactos + existencia de `backend/` y `frontend/` |
| Progreso por tarea | IDs `T0NN` en `git log` cruzados con `specs/*/tasks.md` |
| Validaciones ejecutadas y su veredicto | `specs/*/revision-*.md` |
| Entorno (WSL, herramientas, Docker, `gh`) | Salida de `scripts/doctor.sh` |
| Kit al día · archivos del kit intactos | `instalar_kit.py --verificar` |
| Configuración de agentes al día | `sincronizar.py --verificar` |
| Hooks activos | `git config core.hooksPath` |
| Versiones en soporte (Go, Node, PostgreSQL…) | Tabla de fechas de fin de soporte mantenida en el kit |

**Qué NO se puede derivar (y por tanto se escribe).** Las **decisiones abiertas** y los **pendientes humanos** exigen juicio: ningún generador los deduce. Se registran, cuando nacen, en `equipo/estado/pendientes.yaml`:

```yaml
pendientes_humanos:
  - id: npm-wsl
    descripcion: npm de Windows dentro de WSL; bloquea el frontend
    bloquea: Fase 6 de tasks.md
    referencia: docs/GUIA-INICIO.md §3.1.2
decisiones_abiertas:
  - id: go-local-1.27
    descripcion: instalar Go 1.27, bajar el plan a 1.26 o usar GOTOOLCHAIN=auto
    decide: humano
```

**Dónde tocaría.** `scripts/estado.py` (nuevo), targets `estado` y `estado-verificar` en `Makefile`, plantilla `equipo/estado/pendientes.yaml`, hook `.githooks/pre-commit` (regenerar y exigir `git diff --exit-code` cuando el commit toca `specs/`, `docs/` o `equipo/`), `equipo/orquestador.md` (leer el estado al iniciar sesión; regenerarlo al cerrar cada fase; **ninguna puerta se cierra con el estado sin commitear**) y `docs/GUIA-INICIO.md` (sección «cómo retomar el trabajo»).

**Cómo se verificaría.** `make estado` dos veces seguidas no deja diff; un commit que cambia artefactos sin regenerar el estado es rechazado por el hook; en CI, `make estado-verificar` en el mismo espíritu que `make verificar-agentes`.

**Límite declarado.** La automatización cubre los hechos (≈80 %); el juicio se registra a mano **en el momento en que aparece**, no al final. Prometer más sería vender humo.

---

## 2. El hook `commit-msg` exige el ID de la tarea

**Qué.** Que el hook `commit-msg` valide que el asunto del commit —además de Conventional Commits— termine con el identificador de la tarea, por ejemplo `feat(backend): añade platform/config con validación al arranque (T007)`, y que ese ID exista en el `tasks.md` de la funcionalidad en curso.

**Por qué.** Es la pieza que hace posible el punto 1: sin el ID en el historial, el progreso hay que marcarlo a mano; con él, se **deriva del `git log`**. Además deja trazabilidad directa entre la tarea aprobada y el código que la implementa (útil para QA, para el revisor y para el propio cliente). En este proyecto ya está adoptado como regla de ejecución (`specs/001-estructura-base/tasks.md`), aunque hoy nadie la valide automáticamente.

**Dónde tocaría.** `.githooks/commit-msg`.

**Cómo se verificaría.** Un commit sin ID es rechazado; uno con un ID inexistente en `tasks.md`, también.

---

## 3. Job de pruebas end-to-end en el CI

**Qué.** Añadir al `ci.yml` un job que ejecute Playwright sobre el entorno levantado.

**Por qué.** La constitución (§III) marca las pruebas e2e de flujos críticos como *DEBERÍA*; hoy el kit **monta** la infraestructura (el proyecto la configura, D17 del plan) pero **no la ejecuta nunca en CI**, así que un flujo crítico roto puede integrarse en verde. Todo lo que no corre en CI acaba envejeciendo.

**Dónde tocaría.** `.github/workflows/ci.yml`, reutilizando el servicio de PostgreSQL que ya define.

**Cómo se verificaría.** El job falla si un e2e falla; en un proyecto sin e2e, se salta (como ya hace el CI con `backend/` y `frontend/` inexistentes).

---

## 4. Umbral de cobertura mínimo en el CI

**Qué.** Hacer que el paso de pruebas del backend exija un umbral de cobertura (el proyecto asume ≥80 % en `service/`).

**Por qué.** Hoy el proyecto declara el umbral en el plan y lo verifican a mano `qa-tester` y `revisor-codigo`; sin un umbral en CI, la cobertura depende de que alguien mire, y la deuda entra sola.

**Dónde tocaría.** `.github/workflows/ci.yml` (`go test -cover` + umbral, y `vitest --coverage` en el frontend).

**Cómo se verificaría.** Un PR que baje del umbral queda rojo.

---

## 5. Verificación de artefactos generados en el CI

**Qué.** Un paso que compruebe que el código generado está commiteado y al día: `sqlc generate` y `npm run api:gen` regenerando sin dejar `git diff`.

**Por qué.** El CI del kit no puede saber que el proyecto usa generadores; pero la deriva entre el esquema SQL y el código generado (o entre el contrato OpenAPI y los tipos del frontend) es un fallo silencioso clásico. El proyecto lo mitiga con `make sqlc-verify` y una regla de revisión (plan R4), pero la garantía fuerte es un paso de CI.

**Dónde tocaría.** `.github/workflows/ci.yml`, ejecutando esos targets **si existen** en `proyecto.mk` (así el kit no impone ninguna herramienta).

**Cómo se verificaría.** Un PR con SQL o contrato cambiados y sin regenerar queda rojo.

---

## 6. Versiones del kit en fin de vida o con CVEs

| Dónde | Qué pasa | Propuesta |
|---|---|---|
| `docs/GUIA-INICIO.md` | Dice «Go 1.23 o superior»; esa versión alcanzó **fin de vida el 2025-08-12** y ya no recibe parches de seguridad | Subir a la línea soportada (hoy 1.26/1.27) |
| `.github/workflows/ci.yml` | El servicio de PostgreSQL fija `postgres:16.4-alpine`, un *minor* con vulnerabilidades corregidas en minors posteriores | Usar la etiqueta de rama (`postgres:16-alpine`) o revisar el minor periódicamente |
| `.github/workflows/ci.yml` | Fija **Node 22**, cuyo mantenimiento termina en **abril de 2027** | Anticipar la actualización a la siguiente LTS |
| `.agents/skills/go-backend/SKILL.md` | Repite «Go 1.23 o superior» | Alinear con la línea soportada |

**Por qué.** El propio proceso exige que las versiones en uso tengan soporte de seguridad vigente (requisito del proyecto, `spec.md` FR-016); si el kit las fija, el proyecto no puede cumplirlo solo.

---

## 7. Leer el estado al iniciar sesión

**Qué.** Que el manual del orquestador (`equipo/orquestador.md`) incluya `ESTADO.md` en la lectura obligatoria al iniciar cada sesión, junto a `AGENTS.md`.

**Por qué.** De nada sirve un estado derivado y verificado si el agente que retoma el trabajo no lo lee. Es la mitad del valor del punto 1.

**Dónde tocaría.** `equipo/orquestador.md` (sección «Documentos que mandan» y arranque de sesión) y `docs/GUIA-INICIO.md`.

---

## Seguimiento

| # | Propuesta | Estado en el proyecto | Dueño |
|---|---|---|---|
| 1 | Estado derivado y verificable | `ESTADO.md` se mantiene a mano mientras tanto | Equipo del kit |
| 2 | ID de tarea en `commit-msg` | **Adoptado** como regla de ejecución (2026-10-03); sin validación automática | Equipo del kit |
| 3 | Job e2e | Registrado en el plan (R2) y en `tasks.md` T035 | Equipo del kit |
| 4 | Umbral de cobertura | Registrado en el plan (R2) | Equipo del kit |
| 5 | Artefactos generados | Mitigado con `sqlc-verify` + regla de revisión (plan R4 / research R20) | Equipo del kit |
| 6 | Versiones (Go, Node, PostgreSQL) | Registrado en el plan (R10, R11) y en `tasks.md` T034–T035 | Equipo del kit |
| 7 | Lectura del estado | Pendiente, depende de la propuesta 1 | Equipo del kit |
