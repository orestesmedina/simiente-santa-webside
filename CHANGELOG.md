# Changelog

Todos los cambios notables de este proyecto se documentan en este archivo.

El formato sigue [Keep a Changelog](https://keepachangelog.com/es/1.1.0/) y el proyecto usa [Versionado Semántico](https://semver.org/lang/es/).

## [0.1.0] — 2026-10-03

Primera entrega: **F1 — Estructura base** (rama `001-estructura-base`, PR #1).

### Agregado

- **Backend en Go 1.27** (`backend/`) con el endpoint `GET /healthz`, que informa en cada consulta el estado **real** de la conexión a PostgreSQL: `200` con `{"status":"ok","database":"connected"}` cuando todo está vivo y `503` con sobre de error `database_unavailable` cuando la base no responde (respuesta en ≤ 2 s; FR-004).
- **Formato de respuesta uniforme** para éxito y error en toda la API (SC-008), sin exponer información interna del sistema ante errores inesperados (SC-009).
- **Plataforma interna del backend**: configuración por variables de entorno, logger, manejo de errores, conexión a PostgreSQL, servidor HTTP y middleware (CORS y logging por petición).
- **Frontend en React 19 + TypeScript + Vite** (`frontend/`): página de estado que consulta `/healthz` y muestra de forma comprensible si el backend y la base de datos están vivos. Los tipos TypeScript se generan desde el contrato de la API (`make api-gen`).
- **PostgreSQL 16** con migraciones versionadas (`golang-migrate`): baseline `000001` (no-op). En F1 **no** hay tablas de negocio.
- **Docker Compose**: `make up` levanta base de datos, backend y frontend con un único comando y Docker como único prerequisito (FR-001); `make down` detiene el entorno conservando los datos.
- **CI en GitHub Actions con 6 jobs**: kit y configuración de agentes, migraciones y constitución, detección de proyectos, backend (Go), frontend (React) y búsqueda de secretos. En verde (6/6) en el PR #1.
- **Contrato vivo** [`backend/api/openapi.yaml`](backend/api/openapi.yaml) (OpenAPI 3.1): única fuente de la API y origen de los tipos del frontend.
- **Toolchain fijada**: Go 1.27, Node.js 24 (LTS) y PostgreSQL 16; además `golang-migrate` v4.20.1, `sqlc` v1.31.1, `golangci-lint` v2.14.0 y `govulncheck` v1.8.0. `make doctor` verifica el entorno y todas las versiones en uso tienen soporte de seguridad vigente (SC-010, comprobado a 2026-10-03).
- **Receta para agregar un área de negocio nueva** (`docs/tecnico/arquitectura.md` §8): verificada antes de cerrar F1 (SC-007) con tres corridas de un agente fresco; la tercera añadió un área de práctica (`GET /api/v1/muestra`) sin tomar decisiones de arquitectura y sin modificar las áreas existentes, con `make ci` en verde. Evidencias en `specs/001-estructura-base/checklists/receta.md`.
- **Comandos `make`** para el flujo diario (`up`, `down`, `test`, `lint`, `security`, `ci`, `db-migrate`, `e2e`, `api-gen`, `sqlc-gen`, `doctor`, …) y hooks de git con búsqueda de secretos (`gitleaks`).
- **Pruebas automáticas**: backend (unitarias y de integración con `-tags=integration`), frontend (Vitest) y end-to-end (Playwright).

### No entra en esta versión

- **Sin tablas de negocio ni endpoints además de `/healthz`**: F1 monta el terreno (Go + React + PostgreSQL + Docker + CI) pero no añade contenido visible al sitio.
- **Sin autenticación ni sesiones**: llegan con **F2** (acceso y gestión de usuarios), que introducirá también el uso de `SESSION_SECRET` (hoy documentada en `.env.example` pero sin consumo en F1).
- **Sin despliegue a producción**: la entrega se completa con el merge humano del PR #1.
