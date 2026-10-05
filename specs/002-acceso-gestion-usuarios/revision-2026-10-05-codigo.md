# Revisión de código — F2 (acceso y gestión de usuarios)

**Rama:** `002-acceso-gestion-usuarios` · **Fecha:** 2026-10-05 · **Revisor:** `revisor-codigo`
**Persistido por:** orquestador (la sesión del revisor no tenía herramienta de shell/escritura).
**Alcance:** implementación de F2 (backend `usuarios` + `platform`, frontend, migraciones `000002`–`000004`, contrato 0.3.0) frente a `plan.md`, `.specify/memory/constitution.md`, skills `go-backend`/`react-frontend`/`postgres-db` y `contracts/openapi.yaml`.

## Veredicto: APROBADO CON OBSERVACIONES

No hay hallazgos bloqueantes. La implementación es fiel al plan, respeta las capas, cumple el contrato 0.3.0 y viene con pruebas de comportamiento (incluidas carreras anti-bloqueo e inicialización única).

## Puntos correctos (contrastados)

- **Capas**: `pgx`/`pgtype` solo en `internal/db` (sqlc), `repository*.go`, `platform/database` y `testutil`; `service*.go`/`handler*.go`/`model.go` sin pgx. `platform` no importa dominios.
- **UI**: los 13 componentes del inventario existen y las features los reutilizan; sin markup de tabla/formulario duplicado (único `<input>` crudo: M7).
- **Cliente HTTP**: `fetch` solo en `src/api/client.ts`, con `credentials: 'include'`, `X-CSRF-Token` y `ApiError` con `code/message/details`; sin `localStorage`.
- **Contrato 0.3.0**: DTOs, códigos y `details` coinciden con los schemas.
- **Bugs de F1 corregidos con regresión**: `PathValue` (`httpserver/server.go`), `Retry-After` (`httpserver/error.go`).
- **Seguridad funcional**: verificación dummy bcrypt cost 12, 401 genérico, bloqueo 5×15 min, guard anti-bloqueo transaccional, sesión Redis con token opaco/SHA-256, CSRF double-submit firmado, cookies `HttpOnly`/`SameSite=Lax`.
- **Constitución**: migraciones sin huecos con `up`/`down` e invariantes en la base; pruebas por tarea; sin `any`/`@ts-ignore`/`dangerouslySetInnerHTML`.

## Hallazgos

### Bloqueantes
Ninguno.

### Mayores

- **M1 — Se pierden filas `denied`/`failure` cuando el objetivo no existe (FR-023).**
  `backend/internal/usuarios/service_audit.go:237-276` (`actionFromRoute`) fija `TargetUserID`/`TargetRoleID` desde el `{id}` de la ruta sin comprobar existencia; el INSERT choca con la FK (`migrations/000004...up.sql:35,40`) y la fila se pierde (best-effort → solo log). El contrato dice que `targetId` es `null` "cuando no llegó a existir". **Propuesta:** si el id de objetivo no existe (o tras error de FK), reintentar la inserción con objetivo `nil` conservando `target_kind`, o resolver la existencia antes de insertar.
- **M2 — `SESSION_SECRET` puede quedar vacío y el sistema arranca igual.**
  `config/config.go:139-140` acepta secreto vacío; el HMAC del CSRF queda con clave vacía (`session/cookies.go:144-168`). **Propuesta:** exigir `SESSION_SECRET` no vacío (con longitud mínima) cuando `APP_ENV=production`, con error de arranque; avisar también si `BOOTSTRAP_TOKEN` está vacío.

### Menores

- **M3 — `ChangeMyPassword` sin transacción.** `service_auth.go:311-316` hace `UpdateUserPassword` y `SetUserMustChangePassword` por separado. Un `UPDATE`/`WithTx` lo haría atómico.
- **M4 — `INCR`/`EXPIRE` no atómicos en el contador de fallos.** `platform/session/throttle.go:69-85`: si el `EXPIRE` del primer fallo falla, la clave queda sin TTL. Usar `SET NX EX`+`INCR`, scripting o reponer TTL en cada fallo.
- **M5 — Dos `retryAfterSeconds` con semántica distinta.** `usuarios/service_auth.go:389-396` (0) vs `middleware/ratelimit.go:208-219` (mín. 1). Unificar.
- **M6 — FK de rol en carrera responde 400 sin `details.roleId`.** `repository.go:383-408`. Traducir esa FK en los flujos de cuenta a `roleNotFoundError()`.
- **M7 — `<input type="checkbox">` crudo en `RoleForm`.** `frontend/src/features/roles/components/RoleForm.tsx:177-183` es el único markup fuera del inventario de 13.
- **M8 — `estado.md` desactualizado.** Decía 40/54 cuando `tasks.md` solo dejaba T253 (corregido por el orquestador).
- **M9 — `authn` traduce toda caída de infraestructura a 401** (`middleware/authn.go:47-57`); fail-closed correcto, pero conviene log a nivel `Error`/métrica.

### Sugerencias

- **S1** — Inmutabilidad del registro solo por convención + test; una red a nivel de BD (trigger o permisos) la cerraría estructuralmente.
- **S2** — Timeout de 5 s en mutaciones (`api/client.ts`) puede reportar fallo de un POST ya aplicado.
- **S3** — `normalizeConflict` convierte todo 409 no-`admin_required` en "correo/nombre en uso"; filtrar por `pgErr.ConstraintName` lo haría explícito.

## Discrepancia `ux.md` §3.8 ↔ contrato (`currentPassword`)

**Recomendación: mantener lo implementado (contraseña actual obligatoria también en el cambio forzado) y corregir `ux.md` §3.8 como cambio documental.** Fundamento: la spec (`US7 esc. 2`) y el contrato (`currentPassword: required`) son la fuente de verdad; `ux.md` no lo es. Exigir la actual evita que quien robe la sesión fije su propia contraseña. Acción: `disenador-ux`/`documentador` retiran el paréntesis de `ux.md:130` (fase 9), sin tocar `spec.md` ni el contrato.

## Recomendaciones de cierre

1. Corregir **M1** antes del merge (roza un "DEBE" de FR-023).
2. **M2** en la misma pasada de configuración.
3. M3–M6 como deuda menor o `fix:` dedicado; M7–M8 documentales.
4. Registrar M9 en la revisión de `seguridad` y en `estado.md`.
