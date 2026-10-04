# Quickstart de validación — F1 Estructura base

Guía ejecutable para validar F1 de punta a punta. Cada escenario referencia los criterios de aceptación de `spec.md` que verifica. Los detalles de implementación viven en `plan.md` / `research.md`; el contrato en `contracts/openapi.yaml` (documento vivo: `backend/api/openapi.yaml`); la arquitectura y la receta en `docs/tecnico/arquitectura.md`.

## Prerequisitos

- Máquina con **Docker** (y Docker Compose v2) en ejecución. Es el único prerequisito (FR-001).
- Clon del repositorio en la rama `001-estructura-base` (o en `main` tras integrar; la rama por defecto se renombra `master` → `main` en T032).
- En clonos nuevos, una vez: `make instalar-hooks` (activa los hooks de git; ver R3 de `plan.md`).
- Opcional (desarrollo fuera de Docker): **Go 1.27**, Node 22, y las herramientas `migrate`, `sqlc`, `openapi-typescript` — verificables con `make doctor`. Las versiones exactas están en el `README.md`.

## 0. Arranque en un solo comando (US1, FR-001, SC-001)

```bash
git clone <repo> && cd simiente_santa
make up          # docker compose up -d: db + backend + frontend
```

Sin `.env` ni pasos manuales: compose usa los valores por defecto documentados en `.env.example`. Para valores propios: `cp .env.example .env`, editar y repetir `make up`.

**Resultado esperado**: `docker compose ps` muestra `db` (healthy), `backend` y `frontend` en ejecución. Puertos: **5173** (frontend), **8080** (backend), **5432** (PostgreSQL). Si un puerto está ocupado, el arranque falla con error identificable; se cambia con variables (`DB_PORT=5433 HTTP_PORT=8081 WEB_PORT=5174 make up`), tal como documenta el README.

**Arranque repetible** (escenario 2 de US1): `make down && make up` → el entorno vuelve igual, sin estado residual que lo impida (los datos de la BD se conservan en el volumen `pgdata`).

**SC-001 (medible)**: una persona nueva, siguiendo solo este documento, debe tener el entorno levantado en **menos de 15 minutos**.

## 1. Estado del sistema: BD conectada (US2 esc. 1, FR-002, SC-002)

```bash
curl -i http://localhost:8080/healthz
```

**Esperado**: `HTTP/1.1 200 OK`, cabecera `Cache-Control: no-store` y sobre de éxito:

```json
{"status":"ok","database":"connected"}
```

## 2. Estado del sistema: BD no conectada (US2 esc. 2, FR-004, SC-002)

```bash
docker compose stop db
curl -i http://localhost:8080/healthz
```

**Esperado**: `HTTP/1.1 503 Service Unavailable` con **sobre de error** (fallo previsible, US6 esc. 2):

```json
{"error":{"code":"database_unavailable","message":"La base de datos no está conectada","details":{"database":"disconnected"}}}
```

**El backend sigue respondiendo** (no cuelga ni reinicia; la respuesta llega en ~2 s máximo por el timeout del ping).

## 3. Recuperación sin reiniciar (US2 esc. 4, FR-003, edge cases)

```bash
docker compose start db
# esperar unos segundos a que db esté healthy
curl -s http://localhost:8080/healthz
```

**Esperado**: vuelve al sobre de éxito `{"status":"ok","database":"connected"}` **sin reiniciar el backend**: cada consulta refleja el estado real, no el del arranque. El caso inverso (BD que tarda más que el backend en el primer arranque) se observa consultando `/healthz` inmediatamente tras `make up`: responde `503 database_unavailable` hasta que la BD acepta conexiones, y luego `200`.

## 4. Página inicial del frontend (US2 esc. 3, FR-005, SC-005)

Abrir `http://localhost:5173` en el navegador.

**Esperado**: sin ninguna acción, en menos de 3 segundos la página muestra de forma comprensible el estado del backend y de la base de datos (conectado / no conectado). Con `docker compose stop db`, al pulsar el botón **«Volver a consultar el estado»** la página muestra «no conectada» (cada consulta refleja el estado real, FR-003); tras `docker compose start db` y una nueva consulta con el botón, vuelve a «conectada». Con el backend detenido (`docker compose stop backend`), la página muestra «No se pudo consultar el estado del sistema» (error de red o timeout de 5 s de la consulta), claramente distinguido del caso «BD no conectada».

## 5. Formato uniforme de respuestas (US6, FR-012, FR-013, SC-008, SC-009)

Comprobación manual de que **toda** respuesta sigue los sobres del contrato (`contracts/openapi.yaml`):

```bash
# Ruta no documentada → sobre de error (no el texto de la stdlib)
curl -i http://localhost:8080/ruta-que-no-existe
# Esperado: 404 con {"error":{"code":"not_found","message":"…"}}

# Método no documentado → sobre de error
curl -i -X POST http://localhost:8080/healthz
# Esperado: 405 con {"error":{"code":"method_not_allowed","message":"…"}}
```

**Errores inesperados** (US6 esc. 3, SC-009): en F1 no se pueden provocar contra la API viva (la única operación no tiene entrada de usuario — ver R12 de `plan.md`); se verifican con la **suite del sobre de respuestas** (`internal/platform/httpserver/`), que provoca un error interno y un `panic` sobre el stack completo y comprueba que (a) la respuesta es `{"error":{"code":"internal","message":"Error interno del servidor"}}` sin nombres internos, trazas, SQL ni datos de infraestructura, y (b) el detalle interno queda en el log estructurado con su `request_id`. Ejecutarla con:

```bash
cd backend && go test ./internal/platform/...
```

## 6. Validación automática de cambios (US3, FR-006, FR-007, SC-003, SC-004)

1. Abrir un PR con un cambio que **pasa** validaciones → el workflow `CI` de GitHub Actions se dispara solo y todos los jobs quedan en verde (backend: formato, vet, golangci-lint, migraciones, pruebas unitarias e integración contra PostgreSQL real, govulncheck; frontend: lint, typecheck, pruebas, build, npm audit; secretos: gitleaks). Veredicto: **apto**, en menos de 10 minutos.
2. Abrir un PR con una prueba rota a propósito → el job correspondiente falla con el motivo visible en el log. Veredicto: **no apto**.
3. (Garantía de integración completa: requiere dos ajustes humanos de GitHub — (a) crear el repositorio remoto y empujar las ramas, sin lo cual el CI no corre en ningún sitio; (b) activar la protección de la rama principal con estos checks obligatorios — ver Riesgo R1 de `plan.md`. Atención: el `ci.yml` del kit se dispara en push solo a `main`, y la rama por defecto de este repo es `master`.)

Localmente, lo mismo que CI corre con:

```bash
make ci        # lint + test + security (backend y frontend)
```

## 7. Pruebas end-to-end (local)

```bash
make up
make e2e       # proyecto.mk → cd frontend && npx playwright test
```

**Esperado**: la prueba `e2e/status.spec.ts` abre la página inicial y verifica que el estado del sistema se muestra.

## 8. Migraciones (convención establecida)

```bash
make db-migrate   # aplica 000001_baseline (no-op) → esquema en versión 1
```

**Esperado**: `migrate` reporta la aplicación de la versión 1 sin errores; la BD sigue sin tablas de negocio (ver `data-model.md`). Comprobar también que `migrate -path backend/migrations -database "$DATABASE_URL" down 1` la revierte sin efecto.

## 9. Verificación de la receta (US5, FR-014, FR-015, SC-007)

> **Quién lo hace**: **el humano** (confirmación 3), siguiendo la receta **solo** desde `docs/tecnico/arquitectura.md` §8. El *Independent Test* de **US5** añade la exigencia de que sea una persona del equipo que **no** participó en la creación de la receta (y sin consultar decisiones de arquitectura); si no hay otra persona disponible se aplica la válvula de escape de R13 de `plan.md` (validación humana explícita del ejercicio) y así queda registrado en la checklist.

La receta está en `docs/tecnico/arquitectura.md` §8 (10 pasos). Este ejercicio la sigue de principio a fin para agregar un **área de práctica** nueva:

1. Crear una rama aparte: `git checkout -b practica/receta-001`.
2. Seguir los 10 pasos de la receta con un área **pública y de solo lectura** (ejemplo sugerido: dominio `muestra` — migración `000002_create_sample_items`, consultas en `internal/db/queries/muestra.sql`, `internal/muestra/{model,repository,service,handler,routes}.go`, `RegisterPublic` y cableado en `cmd/api/main.go`). Los pasos que aplican íntegros son todos: el área no tiene entradas de usuario (así no depende de `validate`, `rate-limit` ni `CSRF`, diferidos) ni rutas de panel (esas dependen de `authn`/`authz`, que llegan en F2 — límite declarado en `plan.md`).
3. Regenerar los artefactos generados que toque el ejercicio: `make sqlc-gen` (y `make api-gen` si el contrato cambió).
4. **Dejarlo en verde**: `make ci` pasa completo.
5. **Comprobar el aislamiento** (FR-015): `git diff main --stat` muestra **solo archivos nuevos** del área de práctica y la línea de cableado en `cmd/api/main.go`; ninguna área existente (`status`, `platform/`) cambia ni de comportamiento.
6. **Verificación funcional**: `make up` y comprobar que `GET /api/v1/muestra` responde con el sobre de éxito y que `/healthz` sigue funcionando igual (las áreas existentes están intactas).
7. Rellenar la checklist de SC-007 con las evidencias (salida de `make ci`, `git diff --stat`, resultado del paso 6).
8. **Descartar la rama**: `git checkout main && git branch -D practica/receta-001` (la tabla de práctica nunca llega a `main`: F1 cierra con 0 tablas de negocio).

**Esperado (SC-007)**: el proceso se completa sin tomar decisiones de arquitectura (cada paso está indicado), sin modificar las áreas existentes y con las validaciones automáticas en verde.

## 10. Versiones y soporte de seguridad (US7, FR-016, SC-010)

```bash
go version          # (dentro de backend/ o con Go instalado) → go1.27.x
node --version      # → v22.x (lo fija el ci.yml del kit)
docker compose images   # → postgres:16.4-alpine (ver nota)
make ci                 # incluye govulncheck y npm audit sin altas/críticas
```

**Esperado**: todas las tecnologías de la tabla de versiones de `plan.md` ("Métricas del plan") dentro de su periodo de soporte de seguridad vigente. Cuatro apuntes ya registrados como **pendientes de actualización** (US7 esc. 2; el registro completo está en `plan.md`): el minor `postgres:16.4` acumula CVEs corregidos en minors posteriores (propuesta: `postgres:16-alpine` — plan R10/T034); Node 22 deja de recibir soporte en abril de 2027 (plan R11/T035); la imagen `postgres:16.4-alpine` del servicio del `ci.yml` del kit acumula los mismos CVEs; y `docs/GUIA-INICIO.md` del kit aún recomienda «Go 1.23+», una rama en fin de vida desde 2025-08-12. Los dos últimos pertenecen al kit (no editable aquí): se tratan como plan R11 — identificados, propuestos al repositorio del kit y gestionados por el humano (T035).
