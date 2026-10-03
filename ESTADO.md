# Estado del proyecto — Sitio web de la Iglesia Simiente Santa

**Última actualización:** 2026-10-03 · **Rama de trabajo:** `001-estructura-base` · **Funcionalidad en curso:** F1 — Estructura base (MVP núcleo, `docs/producto/roadmap.md` §3)

Este documento es el **punto de entrada para retomar el trabajo**. Cualquier sesión nueva (persona o agente) debe leerlo antes de tocar nada.

---

## 1. Cómo continuar en una sesión nueva

Abre el agente en la carpeta del proyecto y dile literalmente:

> **«Lee ESTADO.md y continúa.»**

El orquestador leerá su manual (`equipo/orquestador.md`), este documento y los artefactos de la funcionalidad en curso (`specs/001-estructura-base/`), y retomará por el punto que dice la sección 3.

Si prefieres ser explícito: *«Retoma F1 en la fase 6 (implementación), empezando por la Fase 1 de `tasks.md`»*.

**Antes de retomar**, comprueba el entorno con `make doctor` (sección 5).

---

## 2. Dónde estamos (fases del proceso)

| Fase | Estado | Artefacto |
|---|---|---|
| 1 · Especificar | ✅ Aprobada por el humano | `specs/001-estructura-base/spec.md` |
| 2 · Aclarar | ✅ Sin marcas pendientes | `spec.md` (Assumptions) |
| 3 · Planificar | ✅ Aprobado por el humano (2026-09-30) | `specs/001-estructura-base/plan.md` |
| 4 · Tareas | ✅ 35 tareas (T001–T035) | `specs/001-estructura-base/tasks.md` |
| 5 · Coherencia | ✅ Verificada y 16 hallazgos corregidos | `tasks.md`, `plan.md`, `research.md`, `docs/tecnico/` |
| **6 · Implementar** | ⬜ **Punto de continuación** | `tasks.md` § «Estado de ejecución» |
| 7 · Validar (QA + revisor + seguridad) | ⬜ Pendiente | — |
| 8 · Converger | ⬜ Pendiente | — |
| 9 · Entregar (PR) | ⬜ Pendiente | — |

**Nada de código está escrito todavía.** No existen las carpetas `backend/` ni `frontend/`.

---

## 3. Punto exacto de continuación

**Fase 6 — implementación**, empezando por la **Fase 1 de `tasks.md`**: `T001–T005` (contrato vivo, módulo Go, `proyecto.mk`, `.env.example` y migración baseline), con la salvedad de secuenciación ya documentada: **T002 y T005 viajan en el mismo PR** (el CI del kit falla si existe `backend/go.mod` sin `backend/migrations/`).

Reglas de ejecución que aplican desde la primera tarea (están en `tasks.md` y en la constitución):

- **Un commit por tarea**, con Conventional Commits.
- **Toda tarea de código incluye sus pruebas** en el mismo commit (constitución §III).
- **Quien escribe no aprueba**: QA, revisión y seguridad validan antes de integrar.
- **Ningún archivo del kit se edita** (regla 10; ver `tasks.md` § «Regla 10»).

---

## 4. Lo aprobado (no reabrir)

- **Alcance de F1** (`spec.md`, aprobada 2026-09-30): entorno con un solo comando, estado del sistema (`/healthz`), validación automática de cada cambio, **plataforma interna** de capacidades transversales, **receta** para agregar áreas de negocio, **formato uniforme de respuestas** y versiones con soporte de seguridad vigente.
- **Arquitectura** (`docs/tecnico/arquitectura.md`): capas `handler → service → repository`; `cmd → dominio → platform`; `platform` no conoce dominios; interfaces definidas por quien las consume; composición solo en `main.go`. Incluye el ejemplo vertical completo del dominio `contacto` y la **receta de 10 pasos**.
- **Decisiones** (`docs/tecnico/decisiones.md`, D-A1…D-A9): monorepo `backend/` + `frontend/` con API REST JSON y contrato OpenAPI como frontera; plataforma interna propia en lugar de un framework; **capa de datos `sqlc`** (SP/funciones solo como excepción justificada); router tras la interfaz `httpserver.Registrar` con `net/http` en F1 (chi, opción de F2, sin impacto en los dominios); **Go 1.27**; config/logs/errores con stdlib y DI manual; **solo `pgx` entra al binario**.
- **Confirmaciones del humano (2026-09-30)**: el sobre de éxito es el **DTO directo** (sin wrapper `{"data":…}`); el 503 de `/healthz` usa el **sobre de error** `database_unavailable`; la **verificación de la receta la hace el humano**.

---

## 5. Pendiente de tu parte (acciones humanas)

| # | Pendiente | Detalle | Cuándo |
|---|---|---|---|
| 1 | **`npm` de Windows en WSL** | `make doctor` lo marca ✗; bloquea todo el frontend. Ver `docs/GUIA-INICIO.md` §3.1.2 | Antes de la Fase 6 de `tasks.md` |
| 2 | **Versión de Go local** | La máquina tiene **1.26.8** y el plan fija **1.27**. Opciones: instalar Go 1.27, bajar el plan a 1.26 (ambas líneas tienen soporte), o confiar en `GOTOOLCHAIN=auto` | Antes de T002 |
| 3 | **Herramientas de calidad locales** | Faltan `golangci-lint`, `govulncheck`, `golang-migrate` y `gitleaks` (el CI sí los corre) | Antes de la implementación de backend |
| 4 | **`gh auth login`** | Sin GitHub CLI autenticado los agentes no pueden abrir el PR | Antes de la fase 9 |
| 5 | **Remoto y rama `main`** | Crear el repositorio en GitHub, renombrar la rama por defecto `master` → `main` y **activar la protección de rama** con los checks del CI (bloqueante para SC-003/SC-004) | T031–T033, antes del PR de cierre |
| 6 | **Imagen de PostgreSQL** | `postgres:16.4-alpine` acumula CVEs de minors posteriores; propuesta `postgres:16-alpine` | T034, antes de T026 |
| 7 | **Vencimientos del kit** | Node 22 (abril de 2027) y el «Go 1.23+» de `docs/GUIA-INICIO.md` del kit: proponer las mejoras al repositorio del kit | T035 |
| 8 | **Verificación de la receta (SC-007)** | Seguir solo `docs/tecnico/arquitectura.md` §8 en una rama descartable | T030, al cerrar F1 |

---

## 6. Decisiones abiertas

1. **Arrancar la fase 6**: quedó pendiente de responder al cerrar la sesión del 2026-09-30.
2. **Go local**: instalar 1.27, bajar el plan a 1.26 o usar la descarga automática del toolchain (sección 5, punto 2).
3. **Sesiones de F2** (D-A7): la propuesta es cookie `httpOnly` con sesión en servidor; **pendiente de confirmación** y no se implementa en F1.
4. **Dudas de contenido de UX** (`specs/001-estructura-base/ux.md` §7.2.1–§7.2.2): mostrar o no `error.message` como apoyo en el error A, y un plegable técnico con `details` del 503 para quien opera.

---

## 7. Estado técnico del repositorio

- **Árbol limpio**; todos los cambios commiteados (`git log --oneline` cuenta la historia completa con Conventional Commits).
- **Hooks de git activos** (`make instalar-hooks` ejecutado). Son por clon: un clon nuevo debe ejecutarlo.
- **Kit** en `ca93b38`, `make verificar-kit` en OK; configuración de agentes sincronizada.
- **`make doctor`**: 16 correctos, 6 avisos y 1 problema (el `npm` de Windows, sección 5).

---

## 8. Dónde está cada cosa

| Qué | Dónde |
|---|---|
| Idea y roadmap (negocio) | `docs/producto/idea.md`, `docs/producto/roadmap.md` |
| Spec, plan, tareas y diseño de F1 | `specs/001-estructura-base/` |
| Progreso de las tareas | `specs/001-estructura-base/tasks.md` § «Estado de ejecución» |
| Arquitectura y decisiones técnicas | `docs/tecnico/arquitectura.md`, `docs/tecnico/decisiones.md` |
| Proceso del equipo y reglas | `AGENTS.md`, `equipo/orquestador.md`, `docs/GUIA-INICIO.md` |
| Reglas no negociables del código | `.specify/memory/constitution.md` |
| Convenciones del stack | `.agents/skills/` (`go-backend`, `postgres-db`, `react-frontend`) |

---

## 9. Registro de cambios de este documento

| Fecha | Cambio |
|---|---|
| 2026-10-03 | Creación: fases 1–5 cerradas, fase 6 como punto de continuación, acciones humanas y decisiones abiertas registradas |
