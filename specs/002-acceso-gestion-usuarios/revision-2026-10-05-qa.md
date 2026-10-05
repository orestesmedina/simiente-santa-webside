# Informe de QA — F2: Acceso y gestión de usuarios

**Fecha**: 2026-10-05 · **Rol**: `qa-tester` · **Rama**: `002-acceso-gestion-usuarios`
**Alcance**: diff de F2 contra `main` (206 archivos, +36 439 / −184).
**Veredicto: APROBADO** (sin hallazgos bloqueantes; ver hallazgos menores).

---

## 1. Evidencia ejecutada en esta revisión

| Suite | Comando | Resultado |
|---|---|---|
| Backend unitarias + integración | `go test -tags=integration -count=1 ./...` con `DATABASE_URL_TEST` apuntando a PostgreSQL real | **ok** (16 paquetes; `usuarios` 111.9 s, `session` 95.6 s) |
| Cobertura del dominio con integración real | `go test -tags=integration -count=1 -coverprofile ./internal/usuarios` | **87.5 % de statements** (≥80 % exigido) |
| Frontend | `npm test -- --run` (Vitest + MSW) | **34 archivos / 176 pruebas, todas en verde** |
| e2e (Playwright via Docker, stack real) | `docker run … mcr.microsoft.com/playwright:v1.63.0-jammy npx playwright test --config e2e/playwright.config.ts --workers=1` | **3 passed** (`acceso`, `auditoria`, `status`) en 13.2 s |

Notas del entorno (documentadas, no defectos de F2):
- `make e2e` en el host falla por librerías ausentes (`libnspr4`/`libnss3`/`libasound2`), sin `sudo`; se ejecutó con la imagen oficial de Playwright como workaround.
- Sin `DATABASE_URL_TEST`, las pruebas de integración del backend **se saltan silenciosamente** (`testutil/db.go:26-32`, `t.Skipf`). Ver hallazgo M-2.
- Durante esta revisión se apuntó `DATABASE_URL_TEST` a la BD del stack local por error de entorno: las pruebas de integración vacían las tablas del dominio y dejaron el administrador con un hash de prueba, lo que hizo fallar los e2e inicialmente. Se reparó la BD (hash bcrypt real de `Semilla.2026`, y luego `TRUNCATE` limpio para que el e2e ejecutara `initialize` de verdad) y **los e2e volvieron a pasar**. No es un defecto del producto; ver hallazgo M-2/M-3 sobre el riesgo.
- Con `fullyParallel: true`, dos e2e en paralelo + ejecuciones consecutivas saturan el rate-limit de 20 peticiones/min por IP (`ratelimit.go:19-22`) y producen fallos espurios (`429` en el primer intento fallido). En serie (`--workers=1`) pasa 3/3. Ver hallazgo M-4.

## 2. Matriz de cobertura: SC-001…SC-013

| Criterio | Evidencia | Estado |
|---|---|---|
| **SC-001** panel exige sesión (FR-001) | `middleware_test.go` (authn 100 %, writeUnauthenticated), `middleware/authn.go:36`; e2e: la navegación sin sesión lleva a `/login`; `TestAuditRoutesRequirePermission` | ✅ |
| **SC-002** login < 30 s | e2e `acceso.spec.ts:56-60` (login → `/panel` en segundos); `TestLoginSuccess` | ✅ |
| **SC-003** inicialización única (FR-007) | `TestIntegrationInitializeCreatesAdminOnce`, `TestIntegrationInitializeRaceCreatesExactlyOne` (carrera, exactamente una cuenta); `TestLoginLockoutFifth…` + `TestInitializeHandlerRateLimited` (429 + `Retry-After` sobre `/setup/initialize`); e2e `helpers.ts ensureInitialized` (201/409) | ✅ |
| **SC-004** nunca sin administración (FR-008) | `TestIntegrationUpdateUserGuardAntiLockout`, `TestIntegrationDeleteRoleWithUsersRESTRICT`, `TestIntegrationUpdateRoleAdminGuard`, `TestIntegrationAdminGuardRace` (transacción + recuento post-mutación); e2e implícito | ✅ |
| **SC-005** cuenta+rol+asociar < 3 min | e2e `acceso.spec.ts:60-74` (rol + cuenta por UI); `TestIntegrationCreateRoleAndCatalog`, `TestIntegrationCreateUserWithAudit` | ✅ |
| **SC-006** desactivada sin acceso, ni con sesión abierta (FR-012) | e2e `acceso.spec.ts:120-160` (sesión cortada al recargar, 403 al reintento, reactivación funciona); `TestIntegrationStoreRevokeUser` (revoca sesiones en Redis); `TestLoginInactiveAccountForbidden` + `details.reason="access_disabled"` (`service_auth.go:204`, `service_auth_test.go:463-466`); `TestResolveInactiveAccountIsRejected` (sesión ya abierta invalidada por identidad) | ✅ |
| **SC-007** permiso en cada operación (FR-016) | `AuthzByModule` (100 %), `AdminChain`; e2e `acceso.spec.ts:91-114` (sin admin: sin links V crack por URL → "No tienes acceso a esta sección") | ✅ |
| **SC-008** sin revelar existencia (FR-003) | `TestLoginGenericFailureIsIdenticalForUnknownAndWrongPassword`; e2e: el 5.º fallo y los 6.º+ muestran el mismo mensaje genérico; nadie@ no aparece en la UI de auditoría (`auditoria.spec.ts:225-227`) | ✅ |
| **SC-009** cambios de permiso inmediatos (FR-018) | `TestIntegrationReplaceRolePermissionsAndDelete`, `TestIntegrationUpdateRolePermissionsImmediate` | ✅ |
| **SC-010** cambiar contraseña < 1 min (FR-010/FR-020) | e2e `acceso.spec.ts:75-89` (contraseña forzada → cambio → panel); `TestChangeMyPassword*` (éxito, actual errónea, política, revoca demás sesiones) | ✅ |
| **SC-011** 0 rol en uso eliminado / 0 repetición / 0 duplicados (FR-014/FR-017) | `TestIntegrationDeleteRoleRules`, `TestIntegrationDeleteRoleWithUsersRESTRICT`, `TestIntegrationInsertUserDuplicateEmail`, `TestIntegrationRoleNameLowerAndCatalog` (normalización mayúsculas/espacios), `e2e` sufijos únicos | ✅ |
| **SC-012** auditoría filtrable / último acceso (FR-021/FR-024) | e2e `auditoria.spec.ts` completo (ambos historiales, filtro por cuenta y fechas, paginación con filtros, último acceso en ficha, estado vacío filtrado); `TestIntegrationLoginEventsInsertAndFilters`, `TestIntegrationAdminActionsFiltersAndRoleTarget`, `TestIntegrationUserLastLogin` | ✅ |
| **SC-013** todo registrado / solo lectura (FR-022/FR-025) | `TestAuditRoutesAreReadOnly` (GET ok; POST/PATCH/PUT/DELETE → 405/404), `TestIntegrationAuditIsInsertOnly` (INSERT-only en BD real), `TestInvalidJSONLeavesFailureRowWithoutCredentials` + `assertNoCredentials` (FR-026), `TestDeniedOperationLeavesDeniedRow` (operación denegada → fila `denied`), `TestAuditRecordFailureDoesNotChangeResponse` | ✅ |

## 3. Cobertura de US1–US8 y FR-001…FR-026

- **US1** (login): † arriba (SC-001/002/006/008). FR-004 logout: e2e `acceso.spec.ts` final + `TestLogoutRevokesSessionAndClearCookies`. FR-005 (30 min / 1 h): `config_test.go:77-81` (defaults), `TestIntegrationStoreIdleTTLRefresh` e `TestIntegrationStoreAbsoluteCutoff` (TTL de inactividad acotado a la vida absoluta, `store.go:60-100`). ✅
- **US2** (admin inicial): FR-007/FR-08 → SC-003/004. ✅
- **US3** (crear cuentas): FR-009 → `TestIntegrationCreateUserWithAudit` + validaciones (correo malformado, teléfono < 7 dígitos, rol inexistente) cubiertas en `service_users*` y `userForm.test.tsx` (frontend); FR-010 política (`password.Validate` 100 %, incl. igualdad normalizada contra nombre/apellidos/correo). ✅
- **US4** (roles): FR-014/FR-015/FR-017 → SC-009/011. **Edge Case «cuenta sin permisos de módulo»**: e2e `acceso.spec.ts:97-114` (ve el aviso y ninguna sección) + `panel.test.tsx:90` y `permissions.test.ts`. ✅
- **US5** (activar/desactivar): SC-006; FR-013 (nunca eliminar cuentas): no existe ruta DELETE de usuario (`handler_users.go`: solo PATCH activo y reset). ✅
- **US6** (mantener roles): SC-009/011; reemplazo de rol en cuenta → `TestIntegrationUpdateUser` (una cuenta, un solo rol). ✅
- **US7** (cambio de contraseña propia): SC-010; no es acción administrativa (FR-023) — cubierto por el diseño del recorder y `TestChangeMyPassword*` (no registra admin action). ✅
- **US8** (auditoría): SC-012/013; FR-022/FR-023/FR-024/FR-025/FR-026 → §2; intento contra correo inexistente registrado sin asociar ni crear nada: `auditoria.spec.ts:225-227` + repositorio. ✅

**Huecos de cobertura identificados**: ninguno crítico. Únicos matices:
1. En e2e, el 403 de cuenta desactivada se verifica por código de estado y corte de acceso, no por `details.reason` (el comentario en `acceso.spec.ts:150-156` describe el bug como abierto, ya corregido — hallazgo M-1).
2. No se ve un test automatizado explícito de la regla de redirección 30-min-inactividad/1h en el navegador (el cierre es redis TTL + `resolve` fallido; cubierto a nivel de store y unit). Aceptable para MVP: el techo se prueba con TTLs reducidos en integración.

## 4. Preguntas específicas del encargo

| Punto | Verificación | Resultado |
|---|---|---|
| FR-006: 5.º fallo → `401` genérico + activa bloqueo; 6.º → `429` con 15 min | `TestLoginLockoutFifthRespondsGenericAndSixthRateLimited` (unit), `TestIntegrationThrottleCountsAndFlag` + `TestIntegrationThrottleLockoutExpires` (Redis real), e2e `acceso.spec.ts:159-190` (5×401, 6.º 429, mensaje de bloqueo, campo correo deshabilitado) | ✅ |
| Aislamiento de sesión/permisos | Sesiones por contexto de navegador distintos en e2e; revocación de sesiones de la cuenta desactivada/reset sin tocar las demás (`TestIntegrationStoreRevokeUserExcept`, `TestChangeMyPasswordRevokesOtherSessionsOnly`); permisos efectivos solo del rol asignado (`AuthzByModule`) | ✅ |
| Auditoría de solo lectura (FR-025) | Rutas de auditoría solo GET (`handler_audit_test.go:283`), repositorio INSERT-only (`repository_audit_integration_test.go:193`), sin purga en MVP | ✅ |
| «Sin accesos» / cuenta nunca con login (FR-021) | e2e verifica estado vacío filtrado y ficha sin fecha (`auditoria.spec.ts`); `TestIntegrationUpdateUserLastLogin` solo exitosos | ✅ |
| Cuenta desactivada → 403 `details.reason="access_disabled"` (bug corregido) | `service_auth.go:204`, `service_auth_test.go:463-466`, frontend `auth.test.tsx:77-98` y `LoginPage.tsx:66` | ✅ (ver M-1: el e2e no remite el detalle) |
| Edge Case «cuenta sin permisos de módulo» | e2e + pruebas frontend (ver US4) | ✅ |

## 5. Hallazgos

### Bloqueantes
Ninguno.

### Ayores
Ninguno.

### Menores
- **M-1 — Comentario de e2e obsoleto sobre el bug `access_disabled`.** `frontend/e2e/acceso.spec.ts:150-156` marca un «HALLAZGO (para dev-backend)» afirmando que el 403 no trae `details.reason`; el bug ya está corregido (`service_auth.go:204`, pruebas unitarias y de frontend lo verifican). Comportamiento: al desactivar una cuenta con sesión, en e2e solo se afirmar 403 y corte de acceso, no el texto.
  *Reproducir*: desactivar una cuenta activa e iniciar sesión con ella.
  *Esperado*: 403 con `details.reason="access_disabled"` (lo entrega el servidor desde que `service_auth.go:204` existe, `service_auth.go:200-205`).
  *Sugerencia* (no aplicada para no introducir flakiness sin aprobación): actualizar el e2e para leer también el `reason` del envelope y eliminar el comentario.
- **M-2 — Desalineación quickstart/pruebas de integración sobre `DATABASE_URL_TEST`.** `quickstart.md §11` (líneas 279-283, en la sección «Research R19» del plan) anuncia que, si `DATABASE_URL_TEST` no está definida, testcontainers levanta PostgreSQL. En el código, `testutil/db.go:32-33` hace `t.Skipf` en ese caso (solo Redis usa contenedores). Sin la variable, una ejecución local de `go test -tags=integration ./...` está *verde falsa* (saltos silenciosos).
- **M-3 — Riesgo de polución de datos si `DATABASE_URL_TEST` apunta a la BD compartida del stack.** Las pruebas de integración vacían (`TRUNCATE`) las tablas del dominio y alteran el administrador (`service_roles_integration_test.go:33` «vaciar tablas del dominio»). Apuntar la variable a la BD de `docker-compose` corrompe el estado de desarrollo e2e (ocurrió durante esta revisión). Sugerencia: documentar en quickstart §11 que `DATABASE_URL_TEST` debe apuntar exclusivamente a una BD efímera.
- **M-4 — Contención del rate-limit por IP en e2e paralelo.** `frontend/e2e/playwright.config.ts` activa `fullyParallel: true` con `retries: 0`; dos_SUITE e2e que/agotados amen un montos de peticiones a `/api/v1/auth/login` y `/setup/initialize` superiores a 20/min por IP (`ratelimit.go:19-24`) provocan `429` espurios (ocurrió: «intento fallido nº 1 — Expected 401, Received 429»). *Reproducción*: ejecutar dos Suite e2e en paralelo o repetidas en menos de 1 minuto. Pasa con `--workers=1`.

## 6. Veredicto

**APROBADO.** Los 13 criterios de éxito tienen evidencia automatizada real en verde (unitarias + integración con PostgreSQL/Redis reales + e2e de extremo a extremo contra el stack Docker). No hay hallazgos bloqueantes; los cuatro hallazgos menores son de documentación/config de pruebas, no de comportamiento de producto, y no debilitan ningún criterio de aceptación.

*Nota de transparencia*: el único estado no reproducido tal cual es el e2e en el host (`make e2e`), impedido por librerías de Chromium ausentes en el WSL (sin `sudo`); se ejecutó con la misma versión de Playwright vía imagen oficial `mcr.microsoft.com/playwright:v1.63.0-jammy` contra el stack real, lo cual solo cambia el shell, no las pruebas.
