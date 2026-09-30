# Quickstart de validación — F1 Estructura base

Guía ejecutable para validar F1 de punta a punta. Cada escenario referencia los criterios de aceptación de `spec.md` que verifica. Los detalles de implementación viven en `plan.md` / `research.md`; el contrato en `contracts/openapi.yaml` (documento vivo: `backend/api/openapi.yaml`).

## Prerequisitos

- Máquina con **Docker** (y Docker Compose v2) en ejecución. Es el único prerequisito (FR-001).
- Clon del repositorio en la rama `001-estructura-base` (o `master` tras integrar).
- Opcional (desarrollo fuera de Docker): Go 1.23+, Node 22 — verificables con `make doctor`.

## 0. Arranque en un solo comando (US1, FR-001, SC-001)

```bash
git clone <repo> && cd simiente_santa
make up          # docker compose up -d: db + backend + frontend
```

Sin `.env` ni pasos manuales: compose usa los valores por defecto documentados en `.env.example`. Para valores propios: `cp .env.example .env`, editar y repetir `make up`.

**Resultado esperado**: `docker compose ps` muestra `db` (healthy), `backend` y `frontend` en ejecución. Puertos: **5173** (frontend), **8080** (backend), **5432** (PostgreSQL). Si un puerto está ocupado, el arranque falla con error identificable; se cambia con variables (`DB_PORT=5433 HTTP_PORT=8081 WEB_PORT=5174 make up`), tal como documenta el README.

**Arranque repetible** (escenario 2 de US1): `make down && make up` → el entorno vuelve igual, sin estado residual que lo impida (los datos de la BD se conservan en el volumen `pgdata`).

## 1. Estado del sistema: BD conectada (US2 esc. 1, FR-002, SC-002)

```bash
curl -i http://localhost:8080/healthz
```

**Esperado**: `HTTP/1.1 200 OK`, cabecera `Cache-Control: no-store` y cuerpo:

```json
{"status":"ok","database":"connected"}
```

## 2. Estado del sistema: BD no conectada (US2 esc. 2, FR-004, SC-002)

```bash
docker compose stop db
curl -i http://localhost:8080/healthz
```

**Esperado**: `HTTP/1.1 503 Service Unavailable` con cuerpo `{"status":"degraded","database":"disconnected"}`. **El backend sigue respondiendo** (no cuelga ni reinicia; la respuesta llega en ~2 s máximo por el timeout del ping).

## 3. Recuperación sin reiniciar (US2 esc. 4, FR-003, edge cases)

```bash
docker compose start db
# esperar unos segundos a que db esté healthy
curl -s http://localhost:8080/healthz
```

**Esperado**: vuelve a `{"status":"ok","database":"connected"}` **sin reiniciar el backend**: cada consulta refleja el estado real, no el del arranque. El caso inverso (BD que tarda más que el backend en el primer arranque) se observa consultando `/healthz` inmediatamente tras `make up`: responde `disconnected` hasta que la BD acepta conexiones, y luego `connected`.

## 4. Página inicial del frontend (US2 esc. 3, FR-005, SC-005)

Abrir `http://localhost:5173` en el navegador.

**Esperado**: sin ninguna acción, en menos de 3 segundos la página muestra de forma comprensible el estado del backend y de la base de datos (conectado / no conectado). Con `docker compose stop db`, al pulsar el botón **«Volver a consultar el estado»** la página muestra «no conectada» (cada consulta refleja el estado real, FR-003); tras `docker compose start db` y una nueva consulta con el botón, vuelve a «conectada». Con el backend detenido (`docker compose stop backend`), la página muestra «No se pudo consultar el estado del sistema» (error de red o timeout de 5 s de la consulta), claramente distinguido del caso «BD no conectada».

## 5. Validación automática de cambios (US3, FR-006, FR-007, SC-003, SC-004)

1. Abrir un PR con un cambio que **pasa** validaciones → el workflow `CI` de GitHub Actions se dispara solo y todos los jobs quedan en verde (backend: formato, vet, golangci-lint, migraciones, pruebas unitarias e integración contra PostgreSQL real, govulncheck; frontend: lint, typecheck, pruebas, build, npm audit; secretos: gitleaks). Veredicto: **apto**, en menos de 10 minutos.
2. Abrir un PR con una prueba rota a propósito → el job correspondiente falla con el motivo visible en el log. Veredicto: **no apto**.
3. (Garantía de integración completa: requiere dos ajustes humanos de GitHub — (a) crear el repositorio remoto y empujar las ramas, sin lo cual el CI no corre en ningún sitio; (b) activar la protección de la rama principal (`master`) con estos checks obligatorios — ver Riesgo R1 de `plan.md`.)

Localmente, lo mismo que CI corre con:

```bash
make ci        # lint + test + security (backend y frontend)
```

## 6. Pruebas end-to-end (local)

```bash
make up
make e2e       # proyecto.mk → cd frontend && npx playwright test
```

**Esperado**: la prueba `e2e/status.spec.ts` abre la página inicial y verifica que el estado del sistema se muestra.

## 7. Migraciones (convención establecida)

```bash
make db-migrate   # aplica 000001_baseline (no-op) → esquema en versión 1
```

**Esperado**: `migrate` reporta la aplicación de la versión 1 sin errores; la BD sigue sin tablas de negocio (ver `data-model.md`).
