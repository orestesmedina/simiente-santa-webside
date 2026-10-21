# Constitución: reglas del proyecto

Lo que este proyecto exige además de los principios: su arquitectura, sus mínimos de prueba, sus bibliotecas obligatorias. Son tan obligatorias como los principios.

## Arquitectura
- Monorepo con `backend/` (Go), `frontend/` (React + TypeScript) y las migraciones de PostgreSQL en `backend/migrations/`.
- El backend DEBE seguir arquitectura por capas: `handler → service → repository`. Los handlers no acceden a la base de datos directamente.
- La comunicación frontend-backend DEBE ser una API REST JSON documentada en OpenAPI (`backend/api/openapi.yaml`).

## Pruebas
- Backend: la capa de servicio (los archivos `service*.go` de cada dominio en `backend/internal/`) DEBE tener al menos 80 % de cobertura. Todavía no se mide: ver `docs/tecnico/deudas.md`.
- Backend: las pruebas de repositorio corren contra PostgreSQL real (testcontainers o el contenedor del CI).
- Frontend: los componentes con lógica DEBEN tener pruebas (Vitest + Testing Library). Los flujos críticos DEBERÍAN tener pruebas end-to-end (Playwright).

## Seguridad
- Las consultas se escriben con `sqlc` sobre `pgx`.
- Nada de `dangerouslySetInnerHTML` sin sanitizar.

## Calidad de código
- Go: errores envueltos con contexto (`fmt.Errorf("...: %w", err)`).
- TypeScript en modo `strict`. Sin `any` salvo justificación.

## Base de datos
- Tablas con `id`, `created_at` y `updated_at`. Claves foráneas e índices explícitos.

## Observabilidad y operación
- El backend DEBE usar logs estructurados (`log/slog`) y exponer `/healthz`.
- Todo servicio DEBE poder levantarse con `docker compose up`.
