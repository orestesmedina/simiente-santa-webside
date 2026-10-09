# Estado: 002-acceso-gestion-usuarios

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 002-acceso-gestion-usuarios |
| Flujo | equipo-feature |
| Fase | 9/9 · Entrega — PR #4 abierto. Bug de CI (Go 1.27.2) corregido y validado; pendiente de push |
| Ciclo de corrección | 3/3 (F2: 5 correcciones aplicadas + 2 documentales) · bug de CI: 1 ciclo, 0 bloqueantes |
| Próximo paso | Push de los 6 commits del arreglo de CI en la rama (~5 archivos: infra, docs y estado) → CI del PR #4 en verde → aprobación humana del merge → `make costos CERRAR=1`. Luego F3 (Portada e información general). |
| Bloqueado por | — (a la espera del push y de la aprobación del merge del PR #4) |
| Actualizado | 2026-10-08 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano | 2026-10-04 | "apruebo la spec actualizada" (con auditoría, US8 + FR-021…FR-026) |
| Plan | aprobado | humano | 2026-10-04 | "apruebo el plan" |
| PR / merge | pendiente | | | (PR #4 abierto; CI en rojo por el bug de Go 1.27.1, corregido el 2026-10-08) |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno) — el ciclo 1 del `analyze` (F-01…F-15) se cerró en verde en el re-análisis; N-1 y N-2 cerrados también. La spec está aprobada y el plan, aprobado.

### Bug de CI del 2026-10-08 (Go 1.27.2)

- Revisión del arreglo en [`revision-2026-10-08-ci-go1272.md`](revision-2026-10-08-ci-go1272.md): **APROBADO** por los tres roles, **0 bloqueantes** (el único «Importante» documental quedó cerrado en `a2ab81a`).
- Deuda no bloqueante: GO-2026-5932 (`golang.org/x/crypto@v0.57.0`, `openpgp`, sin fix) y el `:latest` de la imagen distroless (preexistente de F1).

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-04 — El humano pidió que la recuperación de contraseña por auto-servicio con correo quede registrada como idea futura/backlog (registrada en `roadmap.md` §5 y §6 decisión 9). En el MVP la contraseña la restablece un administrador.
- 2026-10-04 — **Rutas del panel (coherencia `analyze` F-05):** se adoptan las del plan (`/login`, `/cambiar-contrasena`, `/panel`, `/panel/usuarios`, `/panel/roles`, `/panel/auditoria`, `/sin-permiso`); se alinea `ux.md`, que usaba `/entrar` y `/panel/cuenta`.
- 2026-10-04 — **Pantalla de "Puesta en marcha" (coherencia `analyze` F-06):** se retira de `ux.md`; la inicialización única es por API con `BOOTSTRAP_TOKEN` y no debe exponer el secreto en el navegador.
- 2026-10-04 — **T202 (dependencias Go, implementación):** se anclan con `backend/internal/tools/tools.go` (`//go:build tools`) para que `go mod tidy` no las borre antes de sus primeros imports reales (T215/T218/T219/T220); el archivo se retira cuando se usen. Se sube la transitiva `moby/go-archive` a v0.3.0 por `GO-2026-6253` (govulncheck, §IV).
- 2026-10-04 — **T205 (Redis sin persistencia, implementación):** en Redis 7 la imagen trae `save 3600 1` y `VOLUME /data`; "sin volumen" en Compose no bastaba (el estado sobrevivía al `restart`). Se desactiva explícitamente con `--save "" --appendonly no --dir /tmp` (P23).
- 2026-10-05 — **T225 (Retry-After):** `httpserver.WriteError` no emitía la cabecera `Retry-After` que el contrato documenta; se corrigió (bug de F1).
- 2026-10-05 — **Discrepancia pendiente de validar (T226):** `ux.md` §3.8 dice que el cambio obligatorio de contraseña no pide la actual, pero el contrato (aprobado) marca `currentPassword` como `required` y T226 dice "exige la contraseña actual". Se implementó lo del contrato; lo revisa `revisor-codigo`/`disenador-ux` en la fase 7. **No se reabre la spec sin aprobación.**
- 2026-10-05 — **T227 (`authn`):** cualquier error del `Resolver` (incluido un fallo transitorio de BD) se traduce a `401 unauthenticated` (fail-closed). Se deja señalado para la revisión de seguridad/arquitectura.
- 2026-10-05 — **Decisión (validación): contraseña actual en el cambio obligatorio.** La spec (US7 esc. 2) y el contrato (`currentPassword: required`) la exigen; `ux.md` §3.8 decía lo contrario y se **corrigió el documento** (no se reabre spec ni plan). Exigir la actual evita que quien robe la sesión fije su propia contraseña.
- 2026-10-05 — **Hallazgos menores aceptados:** casilla nativa (`<input type="checkbox">`) en el formulario de roles (M7), inmutabilidad del registro solo por convención+prueba (S1), timeout de mutaciones de 5 s (S2), `normalizeConflict` por restricción pendiente (S3) y e2e con `--workers=1` por el rate-limit (M-4 de QA). Se registran como deuda para F3.
- 2026-10-08 — **Bug de CI (Go 1.27.2).** El CI del PR #4 falló en `govulncheck` (exit 3) porque `backend/go.mod` fijaba `go 1.27` **sin parche** y `actions/setup-go` (que lee `go-version-file`) instalaba go1.27.1, con 9 vulnerabilidades de stdlib corregidas en go1.27.2. Se fija `go 1.27.2` en `backend/go.mod` y `golang:1.27.2` en `backend/Dockerfile`. **No se toca el CI** (archivo del kit); la versión la decide el proyecto vía `go.mod`, que es la palanca que el propio kit documenta. Sin cambio de alcance: cae dentro de FR-016 («versión en soporte de seguridad vigente»).
- 2026-10-08 — **Precisión documental (I1 de la revisión).** La nota de D-A5 citaba el rango `GO-2026-6603…6617`, que incluye IDs inexistentes (6606, 6614) y de OpenTelemetry (6615, 6616); se sustituye por los 9 IDs exactos del escaneo (`a2ab81a`).

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-08 — **Bug de CI corregido (Go 1.27.2).** El humano hizo push y abrió el **PR #4**; el CI falló en el paso «Vulnerabilidades» (`govulncheck`, exit 3) por 9 vulnerabilidades de la stdlib de **go1.27.1** (corregidas en **go1.27.2**), causadas por `go 1.27` sin parche en `backend/go.mod`. Diagnóstico → arreglo mínimo (`go 1.27.2` + `golang:1.27.2`) → validación en paralelo (**QA APROBADO**, **seguridad APROBADO**, **revisor APROBADO**; 1 hallazgo documental «Importante» cerrado en `a2ab81a`). Commits locales sin push: `b0cefda` (fix), `a2ab81a` (nota D-A5), `bf999e7` y `8336967` (CHANGELOG). Reporte: [`revision-2026-10-08-ci-go1272.md`](revision-2026-10-08-ci-go1272.md).
- 2026-10-08 — **Sesión retomada brevemente y pausada de nuevo por decisión del humano.** Sin cambios de código; F2 sigue validada y en local, pendiente de revisión local → push + PR.
- 2026-10-05 — **Sesión pausada por decisión del humano: sin push todavía.** F2 completa y validada, documentación de entrega lista, todo en local. Próximo paso: revisión local → push + PR → `make costos CERRAR=1` al aprobar. Costo abierto: **$5.99**.
- 2026-10-05 — **Validación (fase 7) y correcciones.** QA **APROBADO**; revisor **APROBADO CON OBSERVACIONES**; seguridad **APROBADO CON OBSERVACIONES** (0 bloqueantes). Reportes: `revision-2026-10-05-{qa,codigo,seguridad}.md`. Correcciones aplicadas: M1 (auditoría no pierde filas con objetivo inexistente, FR-023), M2 (SESSION_SECRET vacío/corto no arranca en producción), M3 (cambio de contraseña atómico), M4 (TTL atómico del contador de fallos), M5 (`retryAfterSeconds` consistente), M6 (`details.roleId` en la FK de rol) y M9 (log de causa en `authn`). Documental: `ux.md` §3.8 (la contraseña actual se pide también en el cambio obligatorio) y el e2e de acceso desactivado. Menor aceptada: casilla nativa en el formulario de roles (M7). `make ci`, `sqlc-verify` y `api-gen` sin deriva en verde; e2e 3/3 (vía imagen Docker de Playwright; en el host faltan librerías de Chromium).
- 2026-10-05 — **Backend completo (40/54).** Fases 6–9: inicialización única (T229–T230), gestión de cuentas (T231–T234), roles y permisos (T235–T237) y auditoría de solo lectura (T238–T240). Corregido otro bug de F1: `httpserver` no rellenaba `PathValue` en rutas con parámetros (`{id}`). `go test ./...` y `go test -tags=integration ./...` en verde.
- 2026-10-05 — **Implementación (fase 6/9): 28/54 tareas.** Completadas las fases 3–5: migraciones `000002`–`000004` y consultas sqlc (T207–T211); núcleo `platform` (T212–T219, incluida la sesión en Redis); dominio `usuarios` (T220–T228: repositorios, auditoría, login/logout/Resolve, cambio de contraseña, middleware y handlers de acceso). Corregido un ciclo de imports de pruebas de integración (`session`→`testutil`→`middleware`→`session`) extrayendo los helpers de contenedores a `testutil/containers`. Próximo: T229/T230 (inicialización única).
- 2026-10-04 — **Implementación iniciada (fase 6/9).** Fases 1–2 completadas y commiteadas: T201 (contrato 0.3.0), T202 (deps Go), T203 (deps npm), T204 (regenerar `schema.d.ts`) y T205/T206 (Redis en Compose + `.env.example`). 6/54 tareas. Próximo: T207 (migración `000002`). Nota T201: el snapshot `contracts/openapi.yaml` no era YAML válido (escalares con `:`) y la corrección se aplicó **solo** al contrato vivo (el snapshot no se edita, P19).
- 2026-10-04 — **Sesión cerrada.** F2 lista para implementar: spec y plan aprobados, coherencia en verde, 53 tareas. Próximo paso: **T201** (fusionar el contrato OpenAPI 0.3.0) y en paralelo **T202/T203** (dependencias) y **T205/T206** (Redis en Compose y `.env.example`). Costo de la tarea (abierto): **$1.52**.
- 2026-10-04 — Re-`analyze`: **APROBADO** (F-01…F-15 cerrados, sin regresiones). N-1 (inventario de componentes) y N-2 (texto del guard) cerrados. Fase 5/9 (coherencia) cerrada en verde.
- 2026-10-04 — Ciclo de corrección 1: aplicados F-01…F-15 (`arquitecto`, `disenador-ux`, `analista-producto`). Pendiente re-analizar.
- 2026-10-04 — `analyze` (`revisor-codigo`): **RECHAZADO en coherencia documental** (F-01 crítico; F-02–F-06 mayores; F-07–F-15 menores). Ciclo de corrección 1/3; se corrigen con `arquitecto`, `disenador-ux` y `analista-producto`.
- 2026-10-04 — **`tasks.md` redactado** por `arquitecto` (53 tareas, T201–T253). Pendiente `analyze`.
- 2026-10-04 — **Plan de F2 aprobado por el humano** ("apruebo el plan"). Inicio de la fase 3 (tareas y coherencia).
- 2026-10-04 — Plan y `ux.md` actualizados con la **auditoría** (tablas `login_events`/`admin_actions`, endpoints de consulta, pantalla `/panel/auditoria`) y Redis sin persistencia. Pendiente la aprobación del plan.
- 2026-10-04 — **Spec actualizada re-aprobada por el humano** ("apruebo la spec actualizada") con la auditoría incluida. Se actualiza el plan.
- 2026-10-04 — **Cambio de alcance aprobado por el humano: auditoría en F2** (último acceso, historial de inicios de sesión y acciones administrativas, visible en el panel). `analista-producto` añadió US8 y FR-021…FR-026; Redis **sin persistencia**. Requiere re-aprobar la spec.
- 2026-10-04 — El humano resolvió los puntos del plan: **D-A7 confirmada con Redis** (sesiones en Redis por objetivo de aprendizaje), sesión **1 h máxima + 30 min de inactividad**, cuenta con **nombre, apellidos, correo y teléfono**, y token `BOOTSTRAP_TOKEN` mantenido. Spec, plan, ux.md y `docs/tecnico/decisiones.md` (D-A7) actualizados.
- 2026-10-04 — **Plan técnico y `ux.md` redactados** (`arquitecto`, `disenador-ux`); pendiente la aprobación humana del plan y la confirmación de D-A7.
- 2026-10-04 — **Spec de F2 aprobada por el humano** ("apruebo la spec"). Inicio de la fase 2 (plan técnico).
- 2026-10-04 — El humano fijó el bloqueo por intentos fallidos en **15 minutos** (FR-006); incorporado a la spec.
- 2026-10-04 — El humano fijó la política de contraseñas (mín. 8, mayúsculas+minúsculas+números+especiales, distinta del nombre y del correo) y el límite de 5 intentos fallidos; `analista-producto` los incorporó a la spec.
- 2026-10-04 — Aclaraciones Q1–Q5 resueltas con el humano; `analista-producto` actualizó la spec (cero marcas pendientes).
- 2026-10-04 — Inicio de F2. Spec redactada por `analista-producto` con 5 aclaraciones abiertas (Q1–Q5); pendiente de aprobación humana.
