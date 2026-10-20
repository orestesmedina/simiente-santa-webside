# Constitución: reglas del proyecto

Lo que este proyecto exige además de los principios: su arquitectura, sus mínimos de prueba, sus bibliotecas obligatorias. Son tan obligatorias como los principios.

Estas reglas venían en la constitución que el kit instalaba antes (versión 1.0.0). Desde ahora son de este proyecto: conserva las que le sirven y borra las que no.

## Arquitectura
- Monorepo con `backend/` (Go), `frontend/` (React + TypeScript) y `backend/migrations/` (PostgreSQL).
- El backend DEBE seguir arquitectura por capas: `handler → service → repository`. Los handlers no acceden a la base de datos directamente.
- La comunicación frontend-backend DEBE ser una API REST JSON documentada en OpenAPI (`backend/api/openapi.yaml`). El contrato se escribe antes que el código.

## Pruebas
- Backend: cobertura mínima DEBE ser 80% en `service/`. Pruebas de repositorio contra PostgreSQL real (testcontainers o el contenedor de CI).
- Frontend: componentes con lógica DEBEN tener pruebas (Vitest + Testing Library). Flujos críticos DEBERÍAN tener pruebas end-to-end (Playwright).

## Seguridad
- Las consultas parametrizadas se escriben con `pgx` / `sqlc`.
- Nada de `dangerouslySetInnerHTML` sin sanitizar.
- Dependencias DEBEN pasar `govulncheck` y `npm audit` sin vulnerabilidades altas o críticas.

## Calidad de código
- Go: `gofmt`, `go vet` y `golangci-lint` sin errores. Errores envueltos con contexto (`fmt.Errorf("...: %w", err)`).
- TypeScript en modo `strict`. Sin `any` salvo justificación.
- Nombres en inglés dentro del código.

## Base de datos
- Todo cambio de esquema DEBE hacerse con migraciones versionadas (up y down).
- Tablas con `id`, `created_at`, `updated_at`. Claves foráneas e índices explícitos.

## Observabilidad y operación
- El backend DEBE usar logs estructurados (`log/slog`) y exponer `/healthz`.
- Todo servicio DEBE poder levantarse con `docker compose up`.
