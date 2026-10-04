# Quickstart — Cómo probar F2 (acceso y gestión de usuarios) en local

**Fecha**: 2026-10-04 · **Rama**: `002-acceso-gestion-usuarios` · **Spec**: `spec.md` · **Plan**: `plan.md`

Guía de validación manual + automatizada de F2 una vez implementada. Cada sección indica qué
criterio/SC verifica. Prerrequisitos heredados de F1 (`specs/001-estructura-base/quickstart.md`):
Docker, `make up` levantado, `make db-migrate` aplicado, hooks con `make instalar-hooks`.

## 0. Preparar el entorno

```bash
make up                      # levanta db, backend y frontend
make db-migrate              # aplica 000001…000005 (F2: tablas de usuarios/sesiones/roles)
docker compose ps            # db debe estar "healthy"
```

Variables nuevas de F2 (documentadas en `.env.example`; con sus defectos basta en local):

| Variable | Defecto en desarrollo | Para qué |
|---|---|---|
| `SESSION_SECRET` | valor de ejemplo | Firma del token CSRF (P10) |
| `BOOTSTRAP_TOKEN` | valor de ejemplo | Cabecera `X-Setup-Token` de la inicialización única (P8) |
| `SESSION_COOKIE_SECURE` | `false` | `true` solo con HTTPS; `platform/config` no arranca en `production` con `false` |
| `SESSION_IDLE_TTL_MINUTES` | `30` | Inactividad (FR-005) |
| `SESSION_ABSOLUTE_TTL_HOURS` | `12` | Vida absoluta (R15 — solo si se confirma) |

Si cambió el código tras el último `make up`: `docker compose up -d --build` (§8.1.8: `make up`
reutiliza la imagen en caché y es la causa más frecuente de "el cambio no aparece").

Regenerar artefactos si tocó el contrato o las consultas: `make api-gen` y `make sqlc-gen`
(`make sqlc-verify` debe quedar en verde).

## 1. Inicialización única del administrador (FR-007, US2, SC-003)

```bash
curl -i -X POST http://localhost:8080/api/v1/setup/initialize \
  -H "Content-Type: application/json" \
  -H "X-Setup-Token: $BOOTSTRAP_TOKEN" \
  -d '{"fullName":"Ana Responsable","email":"ana@ejemplo.com","password":"Semilla.2026"}'
```

Esperado: `201` con la cuenta creada (rol "Administrador", todos los permisos). **Repetir el mismo
comando** → `409` con *"La inicialización ya se hizo y no puede repetirse"* (US2 esc. 2). Sin la
cabecera o con token erróneo → `401`/`403`. ✅ Verifica SC-003.

## 2. Iniciar sesión (FR-001…FR-003, US1, SC-002)

```bash
curl -i -c /tmp/f2-cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ana@ejemplo.com","password":"Semilla.2026"}'
```

Esperado: `200` con `SessionUser` (permisos incluidos) y dos cookies (`ss_session` `HttpOnly`,
`csrf_token`). Con contraseña errónea **y** con un correo inexistente → el **mismo** `401`
*"Correo o contraseña incorrectos"* (FR-003/SC-008). Abrir `http://localhost:5173/login` en el
navegador y entrar debe llevar al panel en menos de 30 s (SC-002). Sin sesión, cualquier sección del
panel redirige a `/login` (SC-001).

## 3. Sesión actual y cierre (FR-004, FR-005)

```bash
curl -i -b /tmp/f2-cookies.txt http://localhost:8080/api/v1/auth/session   # 200 con la identidad
curl -i -b /tmp/f2-cookies.txt -X POST http://localhost:8080/api/v1/auth/logout \
  -H "X-CSRF-Token: $CSRF_TOKEN"                                            # 200 {"loggedOut":true}
curl -i -b /tmp/f2-cookies.txt http://localhost:8080/api/v1/auth/session   # 401
```

Expiración por inactividad: deja una sesión sin usar `SESSION_IDLE_TTL_MINUTES` (30 min por
defecto; se puede bajar a `1` en `.env` para la prueba) y repite `GET /auth/session` → `401`
(FR-005).

## 4. Crear rol y cuenta (FR-009, FR-014, FR-019, SC-005)

Con la sesión de administrador (guarda `csrf_token` de la cookie para la cabecera):

```bash
# Rol con dos permisos combinados libremente (US4 esc. 2)
curl -i -b /tmp/f2-cookies.txt -X POST http://localhost:8080/api/v1/admin/roles \
  -H "Content-Type: application/json" -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"Contenido","permissions":["eventos","actividades"]}'

# Repetir el nombre con otras mayúsculas/espacios → 409 (Q5)
# {"name":"  contenido  ","permissions":["eventos"]} → 409

# Rol sin permisos → 400 "un rol debe tener al menos un permiso" (FR-014)

# Cuenta con ese rol y contraseña inicial
curl -i -b /tmp/f2-cookies.txt -X POST http://localhost:8080/api/v1/admin/usuarios \
  -H "Content-Type: application/json" -H "X-CSRF-Token: $CSRF" \
  -d '{"fullName":"Carlos Ayudante","email":"carlos@ejemplo.com","roleId":"<id-del-rol>","password":"Cambio.2026"}'

# Correo repetido (aunque se escriba "  Carlos@Ejemplo.com ") → 409 (Q5)
# roleId inexistente → 400 con details.roleId (US3 esc. 5)

curl -i -b /tmp/f2-cookies.txt "http://localhost:8080/api/v1/admin/usuarios?limit=20&offset=0"
# → items con estado, correo y rol (FR-019)
```

Crear cuenta + rol + asociar debe llevar menos de 3 minutos (SC-005) y no debe quedar ningún correo
ni nombre de rol casi duplicado (SC-011).

## 5. Cambio de contraseña (FR-010, FR-020, US7, SC-010)

1. Entra como `carlos@ejemplo.com` con `Cambio.2026`: el panel **exige** cambiarla antes de usar
   nada (US7 esc. 4; el servidor responde `403` con `details.reason=password_change_required` a
   cualquier otra ruta).
2. `POST /api/v1/auth/password` con `currentPassword` y `newPassword`. Prueba los rechazos: menos
   de 8 caracteres, sin mayúscula, sin número, sin especial, igual al correo o al nombre → `400`
   con el requisito incumplido (FR-010/US7 esc. 3).
3. Con una contraseña válida → `200`; sal y vuelve a entrar con la nueva (SC-010: < 1 min).
4. Como administrador, `POST /api/v1/admin/usuarios/{id}/password` con una contraseña nueva: el
   titular debe cambiarla al entrar (US7 esc. 5) y **sus sesiones abiertas quedan revocadas**.

## 6. Permisos y regla anti-bloqueo (FR-008, FR-016…FR-018, SC-004, SC-007, SC-009)

- **Sin permiso no se puede forzar**: con la cuenta de Carlos (rol "Contenido"),
  `POST /api/v1/admin/usuarios` → `403` aunque la UI no muestre los botones (SC-007); en el panel
  no le aparecen las secciones ajenas (US4 esc. 6).
- **Cambio efectivo de inmediato**: quita `eventos` al rol "Contenido" (o añade `medios`) y, sin
  volver a entrar, la próxima acción de Carlos lo refleja (FR-018/SC-009).
- **Anti-bloqueo (US2 esc. 4–6)**:
  - Desactiva a un administrador cuando es el único → `409` con el mensaje de que debe quedar al
    menos un administrador activo.
  - Cámbiale el rol a uno sin `admin_usuarios_roles` siendo el único → `409` por la misma regla.
  - Con dos administradores, desactiva uno → se completa con normalidad (US2 esc. 5).
- **Eliminar roles (US6)**: `DELETE /api/v1/admin/roles/{id}` de un rol sin cuentas → `200`; de un
  rol con cuentas → `409` explicando que primero hay que reasignarlas (SC-011).

## 7. Desactivar cuenta: acceso cortado de inmediato (FR-012, FR-013, US5, SC-006)

1. Con Carlos dentro del panel (sesión abierta), desactiva su cuenta desde el panel de
   administración.
2. Cualquier acción de Carlos en el panel → `401` de inmediato (también con la cookie que tenía).
3. Intenta entrar con sus credenciales correctas → `403` *"Este acceso está desactivado"*.
4. Los datos de la cuenta siguen en el listado (con `isActive: false`) y **no existe** operación de
   eliminación (FR-013): comprueba que el contrato no tiene `DELETE /admin/usuarios/{id}`.
5. Reactívala → recupera el acceso con normalidad (US5 esc. 3).

## 8. Bloqueo por intentos y superficie pública (FR-006, FR-003, SC-008, SC-011)

```bash
for i in 1 2 3 4 5; do
  curl -s -o /dev/null -w "%{http_code}\n" -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"carlos@ejemplo.com","password":"incorrecta"}'
done
# 401 401 401 401 429   ← el 5.º intento ya bloquea
curl -i -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"carlos@ejemplo.com","password":"correcta_incluso"}'   # 429 durante 15 min
```

- El mensaje de bloqueo es **el mismo** para un correo que no existe (repite el bucle con
  `nadie@ejemplo.com`): no revela existencia (SC-008). Pasados 15 minutos, se puede volver a
  intentar (FR-006).
- `rate-limit` por IP: más de 20 peticiones por minuto a `/auth/login` → `429` (P17).
- Los mensajes de error son comprensibles y sin información interna (tabla de "Errores esperados"
  de la spec): verifica cada uno de los listados en `plan.md` §Cobertura.

## 9. CSRF (P10)

Con la cookie de sesión pero **sin** la cabecera `X-CSRF-Token`:

```bash
curl -i -b /tmp/f2-cookies.txt -X POST http://localhost:8080/api/v1/admin/roles \
  -H "Content-Type: application/json" \
  -d '{"name":"SinCsrf","permissions":["eventos"]}'      # 403 forbidden
```

Lo mismo con la cabecera igual al valor de la cookie `csrf_token` → pasa al handler (SC-007: el
servidor verifica todo).

## 10. Pruebas automatizadas y veredicto completo

```bash
go test ./...                          # unitarias (service, handler, platform, middleware)
go test -tags=integration ./...        # repository contra PostgreSQL real (crea app_test y migra antes)
npm test -- --run                      # frontend (Vitest + Testing Library + MSW)
make e2e                               # Playwright: frontend/e2e/acceso.spec.ts
make ci                                # lint + pruebas + migraciones + govulncheck + npm audit
```

El e2e recorre el camino completo: inicializar → login → crear rol → crear cuenta → entrar con ella
(contraseña forzada) → ver solo sus módulos → cambiar contraseña → desactivarla → acceso cortado →
5 intentos fallidos → bloqueo. Cubre SC-002, SC-005, SC-006, SC-007, SC-010 y SC-011.

Comprobación funcional mínima (§8.1.9): `curl -i http://localhost:8080/healthz` sigue respondiendo
`200`/`503` igual que en F1 y una ruta inexistente responde `404` con `ErrorEnvelope`.

## 11. Mapa de criterios → secciones

| Criterio | Sección |
|---|---|
| SC-001 (panel exige sesión) | §2 |
| SC-002 (login < 30 s) | §2, §10 |
| SC-003 (inicialización única) | §1 |
| SC-004 (nunca sin administración) | §6 |
| SC-005 (cuenta+rol+asociar < 3 min) | §4, §10 |
| SC-006 (desactivada sin acceso) | §7, §10 |
| SC-007 (permiso en cada operación) | §6, §9, §10 |
| SC-008 (sin revelar existencia) | §2, §8 |
| SC-009 (cambios de permisos inmediatos) | §6 |
| SC-010 (cambio de contraseña < 1 min) | §5, §10 |
| SC-011 (0 rol en uso eliminado / 0 repeticiones / 0 duplicados) | §1, §4, §6, §8 |
