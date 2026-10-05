# Quickstart — Cómo probar F2 (acceso y gestión de usuarios) en local

**Fecha**: 2026-10-04 · **Rama**: `002-acceso-gestion-usuarios` · **Spec**: `spec.md` · **Plan**: `plan.md`

Guía de validación manual + automatizada de F2 una vez implementada. Cada sección indica qué
criterio/SC verifica. Prerrequisitos heredados de F1 (`specs/001-estructura-base/quickstart.md`):
Docker, `make up` levantado, `make db-migrate` aplicado, hooks con `make instalar-hooks`.

## 0. Preparar el entorno

```bash
make up                      # levanta db, redis, backend y frontend
make db-migrate              # aplica 000001…000004 (F2: usuarios/roles/permisos + auditoría)
docker compose ps            # db y redis deben estar "healthy"
```

Variables nuevas de F2 (documentadas en `.env.example`; con sus defectos basta en local):

| Variable | Defecto en desarrollo | Para qué |
|---|---|---|
| `REDIS_URL` | `redis://localhost:6379/0` | Sesión del panel y contadores de intentos (`research.md` R1/R5) |
| `SESSION_SECRET` | valor de ejemplo | Firma del token CSRF (P10) |
| `BOOTSTRAP_TOKEN` | valor de ejemplo | Cabecera `X-Setup-Token` de la inicialización única (P8) |
| `SESSION_COOKIE_SECURE` | `false` | `true` solo con HTTPS; `platform/config` no arranca en `production` con `false` |
| `SESSION_IDLE_TTL_MINUTES` | `30` | Inactividad (FR-005, confirmada el 2026-10-04) |
| `SESSION_ABSOLUTE_TTL_MINUTES` | `60` | Vida absoluta de la sesión: **1 hora** desde el inicio (R15, confirmada el 2026-10-04) |

Si cambió el código tras el último `make up`: `docker compose up -d --build` (§8.1.8: `make up`
reutiliza la imagen en caché y es la causa más frecuente de "el cambio no aparece").

Regenerar artefactos si tocó el contrato o las consultas: `make api-gen` y `make sqlc-gen`
(`make sqlc-verify` debe quedar en verde).

## 1. Inicialización única del administrador (FR-007, US2, SC-003)

```bash
curl -i -X POST http://localhost:8080/api/v1/setup/initialize \
  -H "Content-Type: application/json" \
  -H "X-Setup-Token: $BOOTSTRAP_TOKEN" \
  -d '{"firstName":"Ana","lastName":"Responsable","email":"ana@ejemplo.com","phone":"+34 612 345 678","password":"Semilla.2026"}'
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

La sesión vive en **Redis** (D-A7 confirmada el 2026-10-04): puedes verla mientras existe.
**Redis corre sin persistencia** (confirmado: sin `appendonly`, sin `save` y sin volumen — P23/R21):
`docker compose restart redis` borra las sesiones (hay que volver a entrar) y los contadores de
intentos (el bloqueo empieza de cero), pero **no toca el registro de auditoría**, que vive en
PostgreSQL (se comprueba en §10).

```bash
docker compose exec redis redis-cli --scan --pattern 'sess:*'      # hay 1 clave con la sesión abierta
docker compose exec redis redis-cli --scan --pattern 'user_sessions:*'

curl -i -b /tmp/f2-cookies.txt http://localhost:8080/api/v1/auth/session   # 200 con la identidad
curl -i -b /tmp/f2-cookies.txt -X POST http://localhost:8080/api/v1/auth/logout \
  -H "X-CSRF-Token: $CSRF_TOKEN"                                            # 200 {"loggedOut":true}
docker compose exec redis redis-cli --scan --pattern 'sess:*'      # ya no hay claves: logout las borra
curl -i -b /tmp/f2-cookies.txt http://localhost:8080/api/v1/auth/session   # 401
```

**Expiración por inactividad (30 min, FR-005)**: deja una sesión sin usar
`SESSION_IDLE_TTL_MINUTES` (bájalo a `1` en `.env` para la prueba) y repite
`GET /auth/session` → `401`. Comprueba el TTL de la clave: `docker compose exec redis redis-cli
TTL sess:<sha256>` baja de 30 min (1800 s) y se refresca con cada petición.

**Vida absoluta (1 hora desde el login, R15)**: pon `SESSION_ABSOLUTE_TTL_MINUTES=2` en `.env`,
reinicia el backend (`docker compose up -d backend`), entra y sigue usándolo cada pocos segundos
(para que la inactividad no sea la que corte): pasados los 2 minutos la sesión caduca igualmente
→ `401`. En la clave se ve `absoluteExpiresAt` fijo desde el login, sin moverse.

**Revocación (FR-012)**: con una sesión abierta de Carlos, desactiva su cuenta desde el panel (o
restablece su contraseña): las claves `sess:*` de Carlos y su `user_sessions:*` desaparecen de
Redis y su cookie deja de servir al instante (§7).

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

# Cuenta con ese rol y contraseña inicial (nombre, apellidos, correo y teléfono, FR-009)
curl -i -b /tmp/f2-cookies.txt -X POST http://localhost:8080/api/v1/admin/usuarios \
  -H "Content-Type: application/json" -H "X-CSRF-Token: $CSRF" \
  -d '{"firstName":"Carlos","lastName":"Ayudante","email":"carlos@ejemplo.com","phone":"612 345 678","roleId":"<id-del-rol>","password":"Cambio.2026"}'

# Correo repetido (aunque se escriba "  Carlos@Ejemplo.com ") → 409 (Q5)
# roleId inexistente → 400 con details.roleId (US3 esc. 5)
# teléfono "12" o "no es un teléfono" → 400 con details.phone (FR-009/US3 esc. 3)
# nombre, apellidos, correo o teléfono vacíos → 400 con el campo que corregir

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
   de 8 caracteres, sin mayúscula, sin número, sin especial, igual al nombre, a los apellidos o al
   correo → `400` con el requisito incumplido (FR-010/US7 esc. 3).
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
  intentar (FR-006). Los contadores viven en Redis (R5): `docker compose exec redis redis-cli
  KEYS 'login:*'` muestra `login:fail:<correo>` (contador) y `login:block:<correo>` (bandera con
  TTL de 900 s); un login correcto las borra.
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

## 10. Auditoría: registro de accesos y de acciones (FR-021…FR-026, US8, SC-012, SC-013)

Genera primero actividad con las cuentas de las secciones anteriores: un login correcto de Ana, un
login de Carlos con contraseña incorrecta, un login contra `nadie@ejemplo.com`, los 5 intentos de §8
(bloqueo) y varias acciones de gestión (crear una cuenta, editarla, desactivarla, reactivarla,
restablecer su contraseña, crear un rol, editar sus permisos, eliminar un rol sin uso).

1. **Historial de accesos** (solo lectura):

   ```bash
   curl -i -b /tmp/f2-cookies.txt "http://localhost:8080/api/v1/admin/auditoria/accesos?limit=20&offset=0"
   # → items con createdAt, result (success/failure), ip, userId y userEmail

   curl -i -b /tmp/f2-cookies.txt "http://localhost:8080/api/v1/admin/auditoria/accesos?userId=<id-de-carlos>&from=2026-10-01T00:00:00Z&to=2026-10-05T00:00:00Z"
   # → solo los intentos de Carlos en el rango [from, to)
   ```

   Los intentos con contraseña incorrecta y los hechos durante el bloqueo aparecen como `failure`;
   el intento contra `nadie@ejemplo.com` aparece con `userId: null` y `userEmail: null`, **sin
   asociarse a ninguna cuenta** y sin crear nada (US8 esc. 7): comprueba en `GET /admin/usuarios`
   que no hay ninguna cuenta "fantasma".

2. **Historial de acciones administrativas**:

   ```bash
   curl -i -b /tmp/f2-cookies.txt "http://localhost:8080/api/v1/admin/auditoria/acciones?limit=20&offset=0"
   ```

   Cada fila dice quién (`actorId`/`actorEmail`), qué (`action`), sobre qué (`targetKind`,
   `targetId`, `targetLabel`), cuándo (`createdAt`) y con qué resultado (US8 esc. 8). Comprueba:

   - la **inicialización** de §1 aparece como `user.create` **sin actor** (FR-007: cuenta como
     creación de cuenta);
   - el restablecimiento de contraseña aparece como `user.password_reset` con quién, sobre qué
     cuenta y cuándo, y **ninguna contraseña en ninguna parte** (FR-026);
   - las operaciones que fallaron (correo duplicado, eliminar un rol en uso, el `409` anti-bloqueo
     de §6) también quedaron registradas, con `result: failure`;
   - el intento de Carlos (sin permiso) sobre `/admin/usuarios` quedó con `result: denied`;
   - el cambio de la propia contraseña (§5) **no** aparece: no es acción administrativa.

3. **Filtros y paginación** (FR-024, SC-012): filtra por `userId` y por `from`/`to` en los dos
   historiales → solo esos registros; `limit`/`offset` pagina **sin perder los filtros** (`total`
   cuadra con lo que se ve); un rango sin resultados → `items: []` y la UI indica que no hay
   registros que coincidan. En `/acciones`, `userId` devuelve lo que esa cuenta hizo **y** lo que se
   hizo sobre ella. Encontrar los accesos y las acciones de una cuenta concreta debe llevar menos de
   1 minuto (SC-012).

4. **Último acceso** (FR-021, US8 esc. 6): `GET /api/v1/admin/usuarios/{id}` de Ana →
   `lastLoginAt`/`lastLoginIp` de su último acceso **exitoso**; el de una cuenta creada y nunca
   usada → `null` (la ficha indica "aún no ha iniciado sesión", sin inventar ninguna fecha). Un
   login fallido no lo mueve.

5. **Solo lectura** (FR-025, SC-013): el contrato solo publica `GET` sobre
   `/api/v1/admin/auditoria/…` —`POST`/`PATCH`/`DELETE` sobre esas rutas no existen (`405`/`404`
   con `ErrorEnvelope`)— y la sección del panel no ofrece editar ni borrar registros. Los registros
   se conservan tras desactivar la cuenta de Carlos (§7) y tras editar sus datos.

6. **Sin permiso** (US8 esc. 5): con la cuenta de Carlos (rol "Contenido"),
   `GET /api/v1/admin/auditoria/accesos` → `403`, y en el panel la sección no le aparece (SC-007).

7. **Auditoría y Redis** (R21/R22): `docker compose restart redis` → vuelve a entrar (las sesiones
   se perdieron) y comprueba que los **dos historiales siguen intactos** y que el último acceso de
   la ficha no cambió: la auditoría vive en PostgreSQL, no en Redis.

## 11. Pruebas automatizadas y veredicto completo

```bash
go test ./...                          # unitarias (service, handler, platform, middleware)
go test -tags=integration ./...        # integración: PostgreSQL real + Redis (testcontainers, ver abajo)
npm test -- --run                      # frontend (Vitest + Testing Library + MSW)
make e2e                               # Playwright: frontend/e2e/acceso.spec.ts + auditoria.spec.ts
make ci                                # lint + pruebas + migraciones + govulncheck + npm audit
```

Las pruebas de integración necesitan **Docker** (que ya lo requiere `make up`): levantan Redis
(`redis:7-alpine`) y, si no hay `DATABASE_URL_TEST`, también PostgreSQL, con `testcontainers-go`
**dentro del propio test** (`research.md` R19). Esto es deliberado: `.github/workflows/ci.yml` es
un archivo del kit y **no se puede editar**, así que el CI los recibe tal cual y los runners de
GitHub Actions tienen Docker disponible.

El e2e recorre el camino completo: inicializar → login → crear rol → crear cuenta → entrar con ella
(contraseña forzada) → ver solo sus módulos → cambiar contraseña → desactivarla → acceso cortado →
5 intentos fallidos → bloqueo. Cubre SC-002, SC-005, SC-006, SC-007, SC-010 y SC-011. El e2e de
auditoría (`auditoria.spec.ts`) genera accesos y acciones y recorre la sección de registro: ambos
historiales, filtros por cuenta y rango de fechas, paginación, último acceso en la ficha, intento
fallido sin cuenta asociada y registro sin controles de edición (cubre SC-012 y SC-013, ver §10).

Comprobación funcional mínima (§8.1.9): `curl -i http://localhost:8080/healthz` sigue respondiendo
`200`/`503` igual que en F1 y una ruta inexistente responde `404` con `ErrorEnvelope`.

## 12. Mapa de criterios → secciones

| Criterio | Sección |
|---|---|
| SC-001 (panel exige sesión) | §2 |
| SC-002 (login < 30 s) | §2, §11 |
| SC-003 (inicialización única) | §1 |
| SC-004 (nunca sin administración) | §6 |
| SC-005 (cuenta+rol+asociar < 3 min) | §4, §11 |
| SC-006 (desactivada sin acceso) | §7, §11 |
| SC-007 (permiso en cada operación) | §6, §9, §11 |
| SC-008 (sin revelar existencia) | §2, §8 |
| SC-009 (cambios de permisos inmediatos) | §6 |
| SC-010 (cambio de contraseña < 1 min) | §5, §11 |
| SC-011 (0 rol en uso eliminado / 0 repeticiones / 0 duplicados) | §1, §4, §6, §8 |
| SC-012 (auditoría: encontrar accesos y acciones de una cuenta < 1 min) | §10, §11 |
| SC-013 (auditoría: todo registrado; 0 registros editables/borrables) | §10, §11 |
