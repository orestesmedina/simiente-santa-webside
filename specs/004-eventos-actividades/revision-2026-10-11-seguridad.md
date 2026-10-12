# Revisión de seguridad — F4 Eventos y actividades

- **Rama**: `004-eventos-actividades` · **Base**: `main` (diff `main...HEAD`, 163 archivos, +31 487/−1 698)
- **Fecha**: 2026-10-11 · **Revisor**: `seguridad` (Fase 7 del flujo `equipo-feature`)
- **Referencias**: `.specify/memory/constitution.md` §IV, `specs/004-eventos-actividades/spec.md` (FR-012/FR-013/FR-015/FR-017, SC-002/SC-005/SC-009/SC-011), `specs/004-eventos-actividades/plan.md`, OWASP Top 10 (2021).
- **Método**: lectura completa del diff (backend `internal/agenda`, `internal/media`, `internal/platform/{storage,validate,paginate,middleware,httpserver,config,session}`, `internal/db/queries`, migraciones 000007–000009, `cmd/api`, frontend `features/eventos`, `features/publico/eventos`, `app/`, `lib/permissions.ts`, e2e de agenda), greps de patrones peligrosos (SQL dinámico, `innerHTML`, secretos), y ejecución de `govulncheck ./...` y `npm audit --audit-level=high`.

## Veredicto: **APROBADO**

Sin hallazgos críticos ni altos. Tres hallazgos bajos (B1–B3) y tres informativos; ninguno bloquea el merge. Las recomendaciones pueden ir en F4 o en deuda técnica documentada.

---

## Hallazgos

### Críticos

Ninguno.

### Altos

Ninguno.

### Medios

Ninguno.

### Bajos

**B1 — El PATCH no revalida la longitud de los campos opcionales; el rechazo llega desde la base como 400 genérico sin `details` por campo (CWE-20, validación de entrada incompleta en la capa de servicio).**
- Archivos: `backend/internal/agenda/service_admin.go:408-491` (`mergeEvento`, campos `TitleEn/FrequencyEs/FrequencyEn/OrganizerEs/OrganizerEn/ImageAltEs/ImageAltEn` vía `Optional` sin tope), `service_admin.go:920-1013` (`mergeActividad`, ídem), `service_admin.go:1240-1269` (`mergeCategoria`, `NameEn`), `backend/internal/agenda/repository.go:231-236` (traducción del SQLSTATE 23514).
- Impacto: **sin consecuencia de seguridad** — el `CHECK` de longitud de la migración 000007 bloquea el guardado y `classify` lo traduce a `400 invalid` (nunca a 500, nunca con detalle interno). Lo que se degrada es el contrato de errores de §8.1.2: el usuario del panel recibe «Algún dato no cumple las reglas de la base de datos» sin saber qué campo recortar, mientras que el `POST` (que sí valida por etiqueta `validate`) responde con el detalle por campo. También es una asimetría POST/PATCH no documentada en `backend/api/openapi.yaml`.
- Corrección sugerida: revalidar los topes (`MaxTitleLength`, etc.) en los `merge*` para los campos presentes del PATCH, o mapear el nombre de la constraint del 23514 a un `details` por campo en `classify`.

**B2 — `imageFile` se acepta solo por patrón, sin comprobar que el archivo exista en el almacén ni que lo subiera este módulo (CWE-20; enlace débil referencia↔archivo).**
- Archivos: `backend/internal/agenda/service.go:411-429` (`validateImage` → `storage.ValidName`), `service_admin.go:480-495` y `1001-1013` (PATCH `imageFile`).
- Impacto: un usuario del panel con `eventos`/`actividades` podría referenciar por nombre un archivo ajeno al módulo (p. ej. una imagen de portada aún en borrador) si conoce el `img_<uuid>.<ext>` exacto, y al publicar su elemento ese binario pasaría a servirse en `/api/v1/media/`. **Explotabilidad ínfima** (UUID v4 no enumerable; el nombre solo lo ve quien tiene el permiso del módulo que la subió) y no expone texto de borradores. Efecto colateral más real: un `imageFile` inexistente deja el elemento publicado con imagen rota (la descarga responde 404), en vez del edge case de la spec «se conserva sin imagen».
- Corrección sugerida: al guardar, comprobar la existencia del archivo en el `Store` (o registrar los archivos subidos en una tabla y validar contra ella). Marcado «a verificar» por el vector improbable.

**B3 — Contraseña de prueba hard-coded en los e2e nuevos de F4 (CWE-798, falso positivo probable).**
- Archivo: `frontend/e2e/agenda-panel.spec.ts:450,459` (`'Inicial.2026'`), siguiendo el patrón ya existente de `portada-panel.spec.ts` y `auditoria.spec.ts` (F2/F3, en `main`).
- Impacto: ninguno en producción: son cuentas creadas en caliente contra el stack local de `make up`, con cambio obligatorio de contraseña inmediato (`loginNewAccount`). Se reporta por completitud del grep de secretos; **no** hay secretos reales en el diff (`.env.example` solo placeholders; `SESSION_SECRET`/`BOOTSTRAP_TOKEN` documentados como «generar uno aleatorio»; `APP_TZ=America/Costa_Rica` no es secreto y está validado al arranque, `config.go:208-215`).
- Corrección sugerida: ninguna obligatoria; opcionalmente moverla a `e2e/helpers.ts` como constante única `TEST_INITIAL_PASSWORD` para que el grep futuro no la reporte una y otra vez.

### Informativos (no requieren cambio para aprobar)

- **I1 — `govulncheck`**: 0 vulnerabilidades que afecten al código. Reporta GO-2026-5932 (`golang.org/x/crypto/openpgp` sin mantener) en un módulo requerido **que el código no llama** (`x/crypto@v0.57.0`; se usa `bcrypt` del mismo módulo, no `openpgp`). `npm audit --audit-level=high`: **0 vulnerabilidades**.
- **I2 — Ampliación de la superficie de escritura de auditoría**: cada 403 autenticado en las 18 rutas nuevas de panel inserta una fila `result='denied'` (middleware `authz.go:63-75` → `agenda/service_audit.go:38-49`), sin rate-limit en las rutas de panel (solo login/setup lo tienen). Es el patrón heredado de F2/F3 (no una regresión de F4) y exige sesión válida, pero conviene vigilar el volumen de `admin_actions` en producción.
- **I3 — Lecturas denegadas no auditan**: `auditActionFromRoute` (`service_audit.go:76-121`) solo registra POST/PATCH/DELETE. Coherente con FR-017 (enumera crear/editar/publicar/eliminar) y con F2; se deja constancia para que no se interprete como hueco.

---

## Verificación contra la constitución §IV y OWASP

**A01 — Control de acceso (CWE-862). ✅**
- Los tres grupos de panel montan la cadena aprobada **por permiso propio** —authn → PasswordGuard → `AuthzByModule(<permiso>)` → CSRF—: `agenda/routes.go:59-87` con `PermissionEventos`/`PermissionActividades`/`PermissionCategorias`, `middleware/chain.go:51-57`. La autorización es **en el servidor**, por rol resuelto en cada petición (`session.Resolver`; los cambios de rol aplican de inmediato), con 401 de red de seguridad si no hay identidad (`middleware/authz.go:41-44`, además `handler.go:184-191`).
- **Cruce entre permisos**: cada grupo exige SU código; `HasPermission` es coincidencia exacta (`session/identity.go:30-38`). Tener «eventos» no abre `actividades` ni `categorias`, ni al revés; ni siquiera «admin_usuarios_roles» (dueño del catálogo) hereda los nuevos. Verificado en código y en e2e (`agenda-panel.spec.ts:408-488`: cuentas con un solo permiso, navegación filtrada, **API forzada → 403** en los módulos ajenos).
- **Público vs panel**: solo `GET /api/v1/proximamente` (agenda) y `GET /api/v1/media/{fileName}` (media) se registran sin middlewares de sesión (`routes.go:32-38`, `media/routes.go:14-18`); ambas son de solo lectura. No existe ruta pública por id de evento/actividad (imposible pedir un borrador por su UUID).
- **CSRF**: double-submit firmado (HMAC) en todo método no seguro de la cadena (`middleware/csrf.go:39-45`, `session/cookies.go:173`); el cliente envía `X-CSRF-Token` con `credentials: 'include'` (`api/client.ts:37-40,152-162`). La sesión es cookie `HttpOnly, SameSite=Lax, Secure` (según `SESSION_COOKIE_SECURE`, obligatorio `true` en producción, `config.go:216-226`).
- **Borradores y vencido nunca públicos (FR-013/SC-002/SC-009). ✅** Las **únicas** consultas que alimentan la superficie pública (`ListUpcomingEvents`, `ListUpcomingEventDates`, `ListUpcomingActivities`, `IsFilePublished`) exigen `publication_state='published'` **y** la regla de próximo (`fecha > today` OR `hoy con end_time >= now`) en la misma consulta, con `today/now_time` calculados en `APP_TZ` (`queries/agenda.sql:145-176,206-222,312-332`; `queries/media.sql`). El DTO público no incluye `publicationState`. El e2e público comprueba texto de borradores/vencidos ausente y la **imagen de un borrador responde 404** (`agenda-publica.spec.ts:173-181`). La descarga `media` valida el nombre contra el patrón del servidor (sin path traversal) y sirve con `X-Content-Type-Options: nosniff`, `Content-Disposition: inline`, `Cache-Control: no-store` y `Content-Type` por extensión (`media/handler_media.go:76-79`, `storage.go:88-101`).

**A03 — Inyección (CWE-89). ✅** Todo el SQL del dominio pasa por sqlc con parámetros nombrados; **cero** `pool.Query/Exec` o `Sprintf` con SQL en `internal/agenda` y `internal/media` (grep vacío). Los `{id}` de ruta se validan con `uuid.Parse` (`handler.go:286-295`); `lang` es `es|en` por lista cerrada; la paginación se acota en el servidor (`MaxLimit=100`, `paginate.go`); el tope público `MaxBlockItems=200` es constante del servidor, no de la petición. Validación de entrada (CWE-20): `DisallowUnknownFields` + un único objeto JSON + `MaxBytesReader` 1 MiB (`handler.go:125-146`); etiquetas `validate` por DTO, fechas reales (`time.Parse` + ida y vuelta), horas `HH:MM` 24 h, `end > start` y `end` exige `start`; `mapUrl` solo `https` con host (`validate.go:275-282` — rechaza `javascript:`, `data:`, `http:`); `categoryId` se resuelve contra el catálogo. Residual: hallazgo B1.

**A03 — XSS (CWE-79). ✅** Cero `dangerouslySetInnerHTML`/`innerHTML`/`eval` en todo el frontend (grep). El contenido de eventos/actividades se renderiza como texto plano de React (`pre-line` para el detalle). El único `href` dinámico (`MapLink`) viaja validado a https por el backend y abre con `rel="noopener noreferrer"` (verificado en e2e). Sin `target="_blank"` sin `noopener`.

**Subida de archivos. ✅** Solo JPEG/PNG/WebP **por firma binaria** (`http.DetectContentType`; SVG, GIF y HTML renombrado caen en `ErrInvalidFormat`); tope aplicado en stream (`MaxBytesReader` con margen multipart) y revalidado en el `Store`; nombre generado por el servidor `img_<uuid>.<ext>` (el del cliente jamás toca el disco); patrón estricto anti-traversal en Open/Delete; escritura atómica temp+rename; directorio `0750` (`storage/{storage,local}.go`). Cada endpoint de subida va tras el permiso de SU módulo. La subida se audita `fail-closed`: si el registro de auditoría falla, el archivo se elimina y la subida se rechaza (`service_admin.go:553-595`). La descarga pública solo sirve archivos referenciados por contenido **publicado y próximo** (ver A01).

**A02 — Secretos (CWE-798). ✅** Sin secretos en el diff (grep de `password|secret|token|api_key` sobre `main...HEAD`: solo campos de formulario, nombres de variables y la contraseña de e2e del hallazgo B3). `.env.example` documenta placeholders y cómo generar `SESSION_SECRET`/`BOOTSTRAP_TOKEN`; `SESSION_SECRET` con longitud mínima 32 obligatoria en producción (fail-closed, `config.go:242-260`); `APP_TZ` documentado y validado como zona IANA al arranque. bcrypt cost 12 (F2, sin cambios).

**A06 — Dependencias. ✅** `govulncheck ./...`: **0 vulnerabilidades que afecten al código** (1 informativa no llamada, ver I1). `npm audit --audit-level=high`: **0 vulnerabilidades**. Ejecutados ambos el 2026-10-11 sobre la rama.

**A09 — Manejo de errores (CWE-209). ✅** `httpserver.WriteError` nunca envía la causa ni los `details` de un 500 al cliente (`error.go:85-116`); los 401/403/404 son mensajes genéricos en español («No tienes permiso para acceder a este módulo», sin revelar qué permiso falta); el 400 del CHECK de base es genérico (B1 es solo calidad del detalle, no fuga). El SQLSTATE se queda en el log con `request_id`.

**Permiso nuevo «categorías» (cambio de alcance aprobado). ✅ Sin vía de escalada.**
- `000009` es solo un `INSERT` al catálogo (`permissions`): **no** otorga el permiso a rol alguno existente; el rol «Administrador» solo lo recibe al (re)crearse en la inicialización única, que lee el catálogo dinámicamente (`service_init.go:106-110`). El `down` retira primero las asignaciones de este permiso y luego la fila (evita el `RESTRICT` de `role_permissions`).
- `000008` amplía el `CHECK` cerrado de `admin_actions` con los 15 códigos `event.*`/`activity.*`/`category.*` — la lista sigue siendo cerrada y garantizada por la BD; el dominio no puede inventar códigos.
- Un rol con solo `categorias` administra el catálogo (nombres/orden) y nada más: no ve eventos ni actividades (403), y el 409 de borrado solo revela el recuento de uso, no contenido.
- F5–F9 siguen «reservados»: `isPermissionAvailable` los marca como no construidos y el e2e de F2 ajustado (`acceso.spec.ts`) conserva el escenario con módulos aún no activos.

**Auditoría (FR-017/SC-013). ✅** Toda mutación inserta su fila(s) en la **misma transacción** (`withTx` + `insertActions`; o el cambio y su registro, o nada); las denegaciones y rechazos de DTO son best-effort, sin transportar el cuerpo de la petición, y las lecturas denegadas no dejan fila (I3). El registro conserva el carácter de solo lectura de F2.

**CORS. ✅** Echo exacto del origen permitido (nunca `*` con credenciales), `Vary: Origin`, `Allow-Credentials` solo para orígenes de la lista (`middleware/cors.go:44-77`); `CORS_ALLOWED_ORIGINS` por entorno.

---

## Recomendaciones (deuda, no bloqueantes)

1. **B1**: revalidar topes de longitud en los `merge*` del PATCH (o mapear constraint→campo en `classify`) para que el 400 del PATCH traiga `details` por campo como el POST.
2. **B2**: comprobar la existencia de `imageFile` en el `Store` al guardar el elemento (evita imágenes públicas rotas y cierra el enlace débil entre módulos).
3. Cuando exista despliegue con TLS: añadir CSP (y HSTS) al frontend público; `SESSION_COOKIE_SECURE=true` ya es obligatorio en producción por configuración.
4. Vigilar el volumen de `admin_actions` (I2); si crece, considerar agregación o muestreo de denegaciones repetidas.

## Pruebas ejecutadas por el revisor

- `govulncheck ./...` (backend): 0 alcanzables. · `npm audit --audit-level=high` (frontend): 0.
- `go vet` + `go test ./internal/agenda/... ./internal/media/...`: ok (los de integración requieren `-tags=integration` + PostgreSQL; ya verificados por T440 según `estado.md`).
- Greps: SQL dinámico (0), `dangerouslySetInnerHTML`/`innerHTML`/`eval` (0), secretos (solo B3), `target=_blank` sin `noopener` (0).
