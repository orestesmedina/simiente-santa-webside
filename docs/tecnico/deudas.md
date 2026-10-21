# Deudas técnicas abiertas

Lo que el proyecto necesita y todavía no tiene. A diferencia de `docs/producto/roadmap.md` (el rumbo del
producto), aquí va lo que afecta a cómo se construye y se comprueba. Cada deuda nombra el rol que la trabaja.

| Deuda | Qué falta | Rol | Origen |
|---|---|---|---|
| **`make db-migrate` no existe** | El README (`## Comandos`), `docs/tecnico/arquitectura.md` y las guías de entrega lo usan para aplicar las migraciones con `golang-migrate`, pero no hay target en `Makefile` ni en `proyecto.mk`. Hoy hay que invocar `migrate -path backend/migrations -database "$DATABASE_URL" up` a mano. | `devops` | Redacción del perfil (`equipo/perfil.json`), 2026-10-20 |
| **Cobertura sin verbo** | Las reglas del proyecto exigen **80% en `service/`**, pero nadie lo comprueba: el CI genera `coverage.out` de Go sin umbral y el frontend no tiene `@vitest/coverage-v8`. Por eso `cobertura` queda en `null` en las dos partes y `make ci` lo informa como «sin definir». Hace falta una herramienta (p. ej. `tools/cobertura.sh` y el umbral de Vitest) y después declararla en el perfil. | `dev-backend`, `dev-frontend` (con `devops` para el CI) | Redacción del perfil (`equipo/perfil.json`), 2026-10-20 |

Al resolver una, se borra su fila y se actualiza el perfil en el mismo commit.
