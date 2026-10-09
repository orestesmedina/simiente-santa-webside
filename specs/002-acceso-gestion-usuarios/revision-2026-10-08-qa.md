# QA · Revisión 2026-10-08 (commit `b0cefda`)

**Rol:** qa-tester · **Alcance:** solo el commit `b0cefda8ab8f1b9679076b0af38d48cbb8b9dc30`
("fix(infra): fija Go 1.27.2 para cerrar las vulnerabilidades de la stdlib").
La F2 completa ya se validó el 2026-10-05; no se re-valida aquí.
**Ámbito del commit:** 4 archivos, solo infraestructura/documentación — `backend/go.mod`
(`go 1.27`→`go 1.27.2`), `backend/Dockerfile` (`golang:1.27`→`golang:1.27.2` + comentario),
`README.md` (2 líneas de tablas de versiones), `docs/tecnico/decisiones.md`
(nota D-A5 del 2026-10-08). Sin cambios de código de producción ni de API.

## Matriz de cobertura

| Criterio de aceptación del arreglo | Prueba / comando | Resultado |
|---|---|---|
| `govulncheck ./...` con toolchain 1.27.2 en verde (el check que fallaba en CI) | `GOTOOLCHAIN=auto go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...` en `backend/` | ✅ exit 0 — "No vulnerabilities found. Your code is affected by 0 vulnerabilities." |
| Toolchain efectiva = `go1.27.2` (regresión de la causa raíz) | `go env GOVERSION` en `backend/` | ✅ `go1.27.2` |
| El bump no rompe compilación / análisis estático | `go build ./...`, `go vet ./...`, `gofmt -l .` | ✅ exit 0 en los tres; `gofmt -l` sin salida |
| Suite unitaria sin regresiones | `go test -race ./...` | ✅ exit 0 — 15 paquetes `ok` (incl. `internal/usuarios` y `internal/platform/middleware`) |
| Suite de integración sin regresiones | `go test -tags=integration ./...` con `DATABASE_URL_TEST` y `REDIS_URL_TEST` apuntando al `docker-compose` local | ✅ exit 0 — todos los paquetes `ok`, incl. `internal/usuarios` 79.8 s, `containers` 17.5 s, `database` 4.8 s |
| Criterios de producto existentes (F2) no afectados | suites unitarias + integración completas en verde | ✅ cobertura intacta (el commit no toca código) |

## Resultados de las validaciones

Con la toolchain efectiva `go1.27.2`, `govulncheck` v1.8.0 ya no reporta ninguna de las 9
vulnerabilidades objetivo (GO-2026-6603, -6605, -6607, -6608, -6610, -6611, -6612, -6613,
-6617): salen en verde, corregidas en 1.27.2. El único hallazgo residual del escaneo
(con `-show verbose`) es:

- **GO-2026-5932** · `golang.org/x/crypto/openpgp` no mantenido — hallado en
  `golang.org/x/crypto@v0.57.0` (corrección N/A). Nivel: **sólo módulos requeridos**;
  `govulncheck` confirma que el código no llama a los símbolos afectados
  ("your code doesn't appear to call these vulnerabilities"). No bloqueante.

## Hallazgos

- **Bloqueantes:** ninguno.
- **Mayores:** ninguno.
- **Menores / observaciones:**
  - (Residual, no bloqueante) GO-2026-5932 en `golang.org/x/crypto@v0.57.0`, sin llamadas
    al símbolo afectado. Es un hecho de dependencia, no de este commit; queda pendiente
    para la F3 (seguimiento / revisión semestral de dependencias).
  - Nota operativa (no defecto de este commit, apunte de QA): las pruebas de integración
    requieren las variables `DATABASE_URL_TEST`/`REDIS_URL_TEST` apuntando a las
    credenciales reales del `docker-compose` del entorno (documentadas en `.env.example`,
    que usa valores de ejemplo distintivos). Sin ese contexto mandado a mano, las pruebas
    fallan por autenticación (SASL 28P01); los defaults de `docker-compose.yml`
    (`app/app_dev_password`) no reflejan el `.env` real del entorno.

## Reproductor del bug (el check de CI que fallaba)

- `govulncheck ./...` con toolchain 1.27.2: **exit 0** (en CI fallaba con exit 3 y las
  9 vulns de stdlib de `go1.27.1`).
- `go env GOVERSION` en `backend/` → `go1.27.2`.
- `.github/workflows/ci.yml:97`: `actions/setup-go` usa `go-version-file: backend/go.mod`;
  en las líneas 115–116 instala y ejecuta `govulncheck ./...`. El `go 1.27.2` de `go.mod`
  corrige directamente la causa raíz: el CI instalará `go1.27.2`, no `1.27.1`.

## Veredicto

# ✅ APROBADO

El commit `b0cefda` resuelve la causa raíz, cierra las 9 vulnerabilidades de la stdlib
(GO-2026-6603…6617) y no introduce regresiones: build, vet, `gofmt`, suite unitaria con
`-race` y suite de integración contra PostgreSQL/Redis reales, todas en verde.
Sin cambios de producto ni de API, así que no hay criterios de aceptación nuevos.
Queda una sola observación residual (GO-2026-5932, no bloqueante, sin llamadas en el
código) para el seguimiento de dependencias de la F3.
