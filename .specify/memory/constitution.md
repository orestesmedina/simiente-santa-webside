# Constitución de la empresa

Principios no negociables para todos los proyectos. Todos los agentes deben cumplirlos.
Niveles: **DEBE** (obligatorio), **DEBERÍA** (recomendado), **PUEDE** (opcional).

## I. La especificación manda
- La spec aprobada es la fuente de verdad. El código DEBE implementar la spec, no interpretarla.
- Todo cambio de comportamiento DEBE reflejarse primero en `spec.md`.
- Cada requisito DEBE tener criterios de aceptación verificables.

## II. Arquitectura
- Monorepo con `backend/` (Go), `frontend/` (React + TypeScript) y `backend/migrations/` (PostgreSQL).
- El backend DEBE seguir arquitectura por capas: `handler → service → repository`. Los handlers no acceden a la base de datos directamente.
- La comunicación frontend-backend DEBE ser una API REST JSON documentada en OpenAPI (`backend/api/openapi.yaml`). El contrato se escribe antes que el código.
- Las dependencias externas DEBERÍAN minimizarse; toda dependencia nueva requiere justificación en `plan.md`.

## III. Pruebas (test-first)
- Toda tarea de implementación DEBE incluir sus pruebas.
- Backend: cobertura mínima DEBE ser 80% en `service/`. Pruebas de repositorio contra PostgreSQL real (testcontainers o el contenedor de CI).
- Frontend: componentes con lógica DEBEN tener pruebas (Vitest + Testing Library). Flujos críticos DEBERÍAN tener pruebas end-to-end (Playwright).
- Ningún cambio se integra con pruebas fallando.

## IV. Seguridad
- DEBE prevenirse inyección SQL: solo consultas parametrizadas (`pgx` / `sqlc`), nunca concatenación de strings. (CWE-89)
- DEBE validarse toda entrada en el backend, aunque el frontend ya valide. (CWE-20)
- DEBE evitarse XSS: nada de `dangerouslySetInnerHTML` sin sanitizar. (CWE-79)
- Contraseñas DEBEN guardarse con bcrypt o argon2. Nunca en texto plano. (CWE-256)
- Secretos DEBEN venir de variables de entorno; nunca en el repositorio. (CWE-798)
- Endpoints protegidos DEBEN verificar autenticación y autorización en el servidor. (CWE-862)
- Dependencias DEBEN pasar `govulncheck` y `npm audit` sin vulnerabilidades altas o críticas.

## V. Calidad de código
- Go: `gofmt`, `go vet` y `golangci-lint` sin errores. Errores envueltos con contexto (`fmt.Errorf("...: %w", err)`).
- TypeScript en modo `strict`. Sin `any` salvo justificación.
- Funciones cortas y con una sola responsabilidad. Nombres descriptivos en inglés dentro del código.

## VI. Base de datos
- Todo cambio de esquema DEBE hacerse con migraciones versionadas (up y down).
- Nunca se modifica una migración ya aplicada; se crea una nueva.
- Tablas con `id`, `created_at`, `updated_at`. Claves foráneas e índices explícitos.

## VII. Observabilidad y operación
- El backend DEBE usar logs estructurados (`log/slog`) y exponer `/healthz`.
- La configuración DEBE venir de variables de entorno.
- Todo servicio DEBE poder levantarse con `docker compose up`.

## VIII. Gobierno
- La spec y el plan requieren aprobación humana antes de implementar.
- Los despliegues a producción requieren aprobación humana.
- Esta constitución solo se modifica con aprobación de la dirección técnica.

**Versión:** 1.0.0
