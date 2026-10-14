# Revisión de seguridad — F6 Ministerios

- **Rama**: `006-ministerios` · **Base**: `main` (diff `main...HEAD`, 123 archivos, +25 356/−143, 47 commits desde `0797d3d`)
- **Fecha**: 2026-10-13 · **Revisor**: `seguridad` (Fase 5, flujo `equipo-feature`)
- **Referencias**: `.specify/memory/constitution.md` §IV, `specs/006-ministerios/spec.md` (FR-011/FR-012/FR-014/FR-015/FR-017/FR-020/FR-025/**FR-026**, SC-002/SC-005/SC-009/SC-011/**SC-014**/SC-016), `specs/006-ministerios/plan.md`, OWASP Top 10 (2021).
- **Método**: lectura completa del diff (backend `internal/ministerios` completo, `internal/db/queries/{ministries,media}.sql`, migraciones 000010–000011, `cmd/api`, `platform/{audit,middleware}`, frontend `features/{ministerios,publico/ministerios,roles}`, `lib/permissions.ts`, `app/router.tsx`, e2e de F6), greps de patrones peligrosos (SQL dinámico, `innerHTML`, secretos), `govulncheck ./...` + `npm audit --audit-level=high`, y **auditoría en vivo contra el stack local** (`make up` corriendo la rama: backend :8080, frontend :5173, PostgreSQL :5432) con peticiones reales por cada vía de salida del dato del encargado, usando las cuentas de prueba del e2e sembradas en el entorno.

## Veredicto: **APROBADO**

Sin hallazgos críticos ni altos. Dos hallazgos bajos (B1–B2, B2 heredado literal de F4) y cuatro informativos; ninguno bloquea el merge. **La privacidad del encargado (FR-026/SC-014) —el punto más crítico de F6— está cerrada en todas las vías verificadas**, con evidencia en vivo incluida (§Privacidad).

---

## Hallazgos

### Críticos

Ninguno.

### Altos

Ninguno.

### Medios

Ninguno.

### Bajos

**B1 — `imageFile` se acepta solo por patrón, sin comprobar que el archivo exista en el almacén ni que lo subiera este módulo (CWE-20; enlace débil referencia↔archivo). Mismo vector que B2 de F4, heredado del patrón.**
- Archivos: `backend/internal/ministerios/service.go:344-351` (`validateMinistry` → `storage.ValidName`), `service_admin.go:418-421` (PATCH `imageFile`), `platform/storage/storage.go:52` (patrón `^img_[0-9a-f-]{36}\.(jpg|png|webp)$`).
- Impacto: un usuario del panel con `ministerios` podría referenciar por nombre un archivo ajeno al módulo (p. ej., una imagen de portada o de agenda aún en borrador) si conoce el `img_<uuid>.<ext>` exacto, y al **publicar** su ministerio ese binario pasaría a servirse en `/api/v1/media/`. **Explotabilidad ínfima** (UUID v4 no enumerable; el nombre solo lo ve quien tiene el permiso del módulo que la subió) y **no expresa texto de borradores** (solo el binario de una imagen). Efecto colateral más real: un `imageFile` inexistente deja el ministerio publicado con imagen rota (la descarga responde 404) en vez del edge case de la spec «se conserva sin imagen».
- Reproducción (no ejecutada, por no ensuciar el entorno): `PATCH /api/v1/admin/ministerios/{id}` con `{"imageFile":"img_<uuid-de-otro-módulo>.jpg","imageAltEs":"alt"}` → 200 y el archivo ajeno pasa a servirse si el ministerio está publicado.
- Corrección sugerida (dev-backend): al guardar, comprobar la existencia del archivo en el `Store` (o registrar las subidas en una tabla y validar contra ella). Marcado «a verificar» por el vector improbable.

**B2 — Contraseñas de prueba hard-coded en los e2e nuevos de F6 (CWE-798, falso positivo probable; mismo patrón que B3 de F4).**
- Archivos: `frontend/e2e/ministerios-panel.spec.ts:436,443,458,465` y `ministeriosSeed.ts` (`'Inicial.2026'` / `'Definitiva.2026'`), siguiendo el patrón de `portada-panel.spec.ts`/`agenda-panel.spec.ts` (F2/F3/F4, ya en `main`).
- Impacto: ninguno en producción: son cuentas creadas en caliente contra el stack local de `make up`, con cambio obligatorio de contraseña inmediato (`loginNewAccount`). Se reporta por completitud del grep de secretos; **no** hay secretos reales en el diff (verificado: `.env.example` sin cambios en el diff, sin valores reales).
- Corrección sugerida (dev-frontend): ninguna obligatoria; opcionalmente moverlas a `e2e/helpers.ts` como constantes únicas (`TEST_INITIAL_PASSWORD`/`TEST_FINAL_PASSWORD`) para que el grep futuro no las reporte una y otra vez.

### Informativos (no requieren cambio para aprobar)

- **I1 — `govulncheck`**: 0 vulnerabilidades que afecten al código. Reporta GO-2026-5932 (`golang.org/x/crypto/openpgp` sin mantener, fixed-in N/A) en un módulo requerido **que el código no llama** — idéntico al I1 de F4. `npm audit --audit-level=high`: **0 vulnerabilidades**.
- **I2 — Ampliación de la superficie de escritura de auditoría**: cada 403 autenticado en las 7 rutas nuevas de panel inserta una fila `result='denied'` (middleware `authz.go:60-72` → `ministerios/service_audit.go:44-55`), sin rate-limit en las rutas de panel (solo login/setup lo tienen). Es el patrón heredado de F2/F3/F4 (no una regresión de F6) y exige sesión válida; conviene vigilar el volumen de `admin_actions` en producción (verificado en vivo: mis denegaciones de prueba dejaron filas `ministry.create/delete/update/image.upload · denied`).
- **I3 — `target_label` de auditoría incluye el nombre del ministerio** (`Ministerios · <nombre>`, `service.go:440-455`): visible para cuentas con permiso de auditoría. **Nunca transporta datos del encargado** (verificado en vivo, §Privacidad v7) — se deja constancia para que no se interprete como superficie de fuga.
- **I4 — `sortOrder` sin cota explícita** (`MinisterioInput.SortOrder int`, `model.go:120`): sin impacto de seguridad (solo decide el orden de presentación; un valor fuera de rango int64 es rechazado al decodificar el JSON). Se deja constancia; una cota (`min/max`) sería solo calidad de contrato.

---

## Auditoría de privacidad del encargado (FR-026/SC-014) — evidencia por vía

F6 guarda **nombre + celular/correo de una persona real** y un interruptor (`leader_visible`, `DEFAULT FALSE`, «oculto por defecto») decide si se publica. Se auditaron **todas** las vías de salida; el dato existe en la base del entorno local (**fila sembrada por el e2e**: `M1 Confidencial`, `leader_name_es='Roberto Teléfono'`, `leader_phone='+506 8888-8888'`, `leader_visible=f`, `publication_state=published`), de modo que cada prueba se hizo **contra un dato real guardado**.

### Defensa en profundidad (código)

1. **SQL (primera barrera)**: `ListPublishedMinistries` es la **única** consulta que alimenta la superficie pública: exige `publication_state='published'` **y** enmascara los cuatro campos del encargado con `CASE WHEN leader_visible THEN … END` → con el interruptor apagado **la fila no transporta el dato** (`backend/internal/db/queries/ministries.sql:44-68`). Las consultas del panel (`ListMinistries`, `GetMinistryByID`) son las únicas con la ficha íntegra y viven tras el permiso.
2. **Entidad/DTO (segunda barrera)**: `MinisterioPublico` **carece de campos** para nombre/contacto del encargado fuera de `leader`, y `leader` solo se construye en `localizeMinistry` si `LeaderVisible && LeaderNameEs != nil` (`repository_public.go:49-74`, `service_public.go:62-86`). El extracto de la portada (`ItemMinisterioPortada`) **no tiene campos de encargado ni de enlace**: la privacidad es estructural en esa ruta.
3. **Borradores**: `WHERE publication_state='published'` (misma consulta); no existe ruta pública por id (imposible pedir un borrador por su UUID); la imagen de un borrador no se sirve (`media.sql` rama `ministries`, `repository_media_integration_test.go:250-292` `TestIntegrationIsFilePublishedFollowsMinistryState`).

### Evidencia en vivo (stack local de la rama, 2026-10-13)

- **v1 — `GET /api/v1/ministerios` (sección pública, sin sesión)**: `M1 Confidencial` aparece con nombre/descripción/horario/lugar/enlace, **sin clave `leader`**. **0 apariciones** de «Roberto Teléfono» y «+506 8888-8888» en el cuerpo crudo.
- **v2 — `GET /api/v1/portada/ministerios` (bloque, sin sesión)**: tarjeta compacta sin encargado; **0 apariciones** del contacto en el JSON crudo.
- **v3 — Caso contrario (interruptor ENCENDIDO, prueba no vacua)**: `M2 Visible` (`leader_visible=t`) devuelve `leader: {name:"Lucía Correo", email:"lucia@ejemplo.com"}` — el mecanismo **sí publica** cuando el equipo lo enciende. Ciclo completo verificado además con un ministerio de prueba propio: alta con encargado+teléfono e interruptor apagado → público sin `leader`; `PATCH leaderVisible=true` → público con `leader:{name,phone}`; `DELETE` → desaparece.
- **v4 — Borrador con ficha visible**: `M3 Borrador Secreto` (`draft`, `leader_visible=t`, `Persona Oculta / oculto@ejemplo.com` guardados): **no aparece** en sección ni en portada (solo los publicados). Su imagen tampoco se sirve (test de integración + e2e `ministerios-publica.spec.ts` con `draftImageFile` → 404).
- **v5 — HTML/JS servido**: `GET /` y `GET /ministerios` del frontend: **0 apariciones** del contacto (grep sobre el HTML servido real).
- **v6 — Logs del servidor** (`docker logs` durante las peticiones): el log estructurado solo registra `request_id/method/path/code/status/duration_ms` — **0 apariciones** del contacto. `recordRejected` y `RecordDenied` transportan solo method+path (`handler.go:158-179`, `service_audit.go:31-68`).
- **v7 — Registro de auditoría** (`admin_actions`): `target_label` es siempre `Ministerios · <nombre del ministerio>` o `Ministerios · <uuid>` o `Ministerios · Imagen · <fileName>` — **nunca** nombre ni contacto del encargado (verificado en la tabla tras las operaciones de prueba). Rechazos (`failure`) y denegaciones (`denied`) sin cuerpo de petición.
- **v8 — Panel con permiso (dato guardado ≠ dato publicado)**: `GET /api/v1/admin/ministerios` con una cuenta con `ministerios` devuelve la ficha íntegra de `M1 Confidencial` (`Roberto Teléfono / +506 8888-8888`) — FR-026 exige que el dato siga disponible en el panel: **sí lo está**, y solo ahí.
- **v9 — Respuestas de error**: 500 sin causa interna (test `TestGetMinisteriosHandlerServiceError`: «boom» no sale del servidor), 401/403 genéricos sin revelar el permiso que falta, 400 con `details` por campo (nada interno). El PATCH revalida la fila fusionada completa (a diferencia del B1 de F4, **no** hay asimetría POST/PATCH en F6: `mergeMinisterio` → `validateMinistry`).
- **v10 — Servicios vecinos**: `GET /api/v1/portada/proximamente` (F4) no incluye ministerios ni datos del encargado; no hay ninguna otra consulta pública que toque `ministries` (grep sobre `queries/*.sql`: la única vía pública es `ListPublishedMinistries`).
- **v11 — e2e de privacidad**: `ministerios-publica.spec.ts:220-299` comprueba SC-014 sobre tráfico de red, HTML de `/ministerios`, HTML de la portada y respuestas forzadas de la API — **0 apariciones** del nombre y del celular del encargado oculto; `ministerios-panel.spec.ts` comprueba el ciclo del interruptor en la interfaz.

**Conclusión de privacidad**: con el interruptor apagado, el nombre y el celular/correo del encargado **no salen del servidor** por ninguna vía verificada (API pública, portada, HTML, JS, logs, auditoría, media, errores, servicios vecinos). Un borrador no se filtra por ninguna vía (sección, portada, imagen por URL directa).

---

## Verificación contra la constitución §IV y OWASP

**A01 — Control de acceso (CWE-862). ✅**
- Las 7 rutas del panel montan la cadena aprobada **por el permiso propio** —authn → PasswordGuard → `AuthzByModule("ministerios")` → CSRF (`routes.go:52-66`, `middleware/chain.go:51-57`), autorización **en el servidor** por rol resuelto en cada petición. **En vivo**: sin sesión → **401** en las 5 operaciones (GET/POST lista, GET/PATCH/DELETE `{id}`, POST `/imagenes`); cuenta con permisos `eventos,noticias` → **403** en ministerios y **200** en eventos (su módulo); cuenta `solo-ministerios` → **200** en su módulo y **403** en `eventos`, `actividades` y `usuarios` (cruce en ambas direcciones); 403 genérico («No tienes permiso para acceder a este módulo», sin revelar el permiso). **CSRF**: POST con sesión válida y **sin** token → 403 («El token de seguridad no es válido o ha caducado»).
- Cobertura de pruebas: `TestMinisteriosPermissionsSmoke` (`cmd/api/main_test.go:943`, cruce con los 3 permisos de F4), `middleware_test.go:645-706`, e2e `ministerios-panel.spec.ts:408-527` (incluida la subida de imágenes → 403 y la navegación que oculta el módulo).
- **Permisos (FR-011/Decisión 9)**: F6 **no crea ni amplía permisos**: el diff solo añade las migraciones 000010/000011 (la tabla y el `CHECK` cerrado de `admin_actions` con 8 códigos `ministry.*`); `permissions`/`roles`/`role_permissions` de F2 intactos (verificado en el diff de `backend/migrations/`), el catálogo sigue en 10 con `('ministerios','Ministerios')` sembrado por 000002, y `TestIntegrationPermissionsCatalogUnchanged` lo fija. El frontend **activa** `ministerios` en `MODULE_PERMISSIONS` (`lib/permissions.ts`) sin tocar el modelo.
- **Público vs panel**: solo `GET /api/v1/ministerios` y `GET /api/v1/portada/ministerios` se registran sin middlewares de sesión (`routes.go:27-36`), ambas de solo lectura y solo publicadas.

**A03 — Inyección SQL (CWE-89). ✅** Todo el SQL del dominio pasa por sqlc con parámetros nombrados (`sqlc.arg`/`sqlc.narg`); **cero** `pool.Query/Exec` o `Sprintf` con SQL en `internal/ministerios` y `internal/media` (grep vacío). Los `{id}` de ruta se validan con `uuid.Parse` (→ 400 `details.id`); `lang` es `es|en` por lista cerrada (→ 400 `details.lang`); `limit/offset` acotados en el servidor (`paginate.MaxLimit=100`); el tope 5 del bloque y el `maxPublicItems=200` de la sección son constantes del servidor, no de la petición; el filtro `publicationState` del panel es lista cerrada.

**A03 — XSS (CWE-79). ✅** Cero `dangerouslySetInnerHTML`/`innerHTML`/`eval`/`document.write` en todo el frontend (grep, incluidos los archivos nuevos de F6). Todo el contenido se renderiza como texto de React (`whitespace-pre-line` en la descripción). **En vivo**: un `PATCH` con `leaderNameEs:"<script>alert(1)</script>"` se guardó como texto (dato guardado) y con el interruptor apagado **no salió** por la API pública; al publicarse saldría escapado como texto de React (sin `innerHTML`, sin vector). El contacto del encargado se muestra como **texto**, nunca como enlace (`ContactoEncargado.tsx`/`EncargadoLine.tsx` sin `href`/`mailto:`/`tel:`).

**Enlace de WhatsApp (FR-015/SC-009). ✅** Validación en el servidor con `platform/validate` etiqueta `url` (`validate.go:277-282`: **solo `https` con host**, rechaza `javascript:`, `data:`, `vbscript:`, `http:`, rutas relativas) + rechazo de espacios/caracteres de control (`service.go:334-340`); el `CHECK` de la tabla (`whatsapp_url ~ '^https://[^[:space:]]+$'`) es la última línea. **En vivo**: `javascript:alert(1)`, `data:text/html,<script>` y `http://…` → **400** con `details.whatsappUrl` («Escribe un enlace válido que empiece por https://»); un `https` válido se acepta y se sirve **tal cual** (FR-015: sin transformaciones). El único `href` dinámico (`WhatsappJoinButton`) viaja validado a https por el backend y abre con `target="_blank" rel="noopener noreferrer"` (`PublicButton` con `external`, `WhatsappJoinButton.tsx:46-52`); sin `noopener` faltante en ningún `target=_blank` (grep). El botón solo se monta si hay enlace (0 botones vacíos/rotos, SC-009).

**A04 — Subida de archivos (FR-020). ✅** Solo JPEG/PNG/WebP **por firma binaria** (`http.DetectContentType`; SVG, GIF y HTML renombrado caen en `ErrInvalidFormat`); tope aplicado en stream (`MaxBytesReader` con margen multipart) y revalidado; nombre generado por el servidor `img_<uuid>.<ext>` (el del cliente jamás toca el disco); patrón estricto anti-traversal (`^img_[0-9a-f-]{36}\.(jpg|png|webp)$`) en `ValidName`/Open/Delete (**en vivo**: `imageFile:"../../etc/passwd"` → 400 «sube la imagen desde el formulario»); escritura atómica temp+rename (F3, sin cambios); `imageAltEs` obligatorio si hay imagen (service + CHECK de BD). La subida se audita **fail-closed** (`service_admin.go:340-345` + `TestUploadMinisterioImageHandlerAuditFailureIsClosed`): sin registro no hay subida. La descarga pública sirve con `X-Content-Type-Options: nosniff`, `Content-Disposition: inline`, `Cache-Control: no-store` y `Content-Type` por extensión, y **solo archivos referenciados por contenido publicado** (`IsFilePublished`, rama `ministries` añadida): la imagen de un borrador responde 404 y vuelve a 200 al publicarlo (test de integración). Residual: B1.

**A02 — Secretos (CWE-798). ✅** Sin secretos en el diff (grep de `password|secret|token|api_key` sobre `main...HEAD`: solo nombres de variables, campos de formulario, `CSRFSecret:"secret"` de un test y las contraseñas efímeras de e2e de B2). `.env.example` **sin cambios** (0 variables nuevas, solo placeholders documentados). La sesión sigue siendo cookie `HttpOnly, SameSite=Lax, Secure` y CSRF double-submit firmado (F2, sin cambios); bcrypt sin cambios.

**A05 — Configuración. ✅** CORS: echo exacto del origen permitido — **en vivo**, con `Origin: https://evil.example` la respuesta no lleva `Access-Control-Allow-Origin` (solo `Vary: Origin`); `CORS_ALLOWED_ORIGINS` por entorno (F2, sin cambios). Cabeceras de seguridad de media intactas; `Cache-Control: no-store` en las dos respuestas públicas nuevas (verificado en test y handler).

**A06 — Dependencias. ✅** `govulncheck ./...`: **0 vulnerabilidades que afecten al código** (1 informativa no llamada, I1). `npm audit --audit-level=high`: **0 vulnerabilidades**. **0 dependencias nuevas**: `go.mod`, `go.sum`, `package.json` y `package-lock.json` **sin cambios** en el diff (F6 reusa sqlc/pgx/storage/zod existentes; `CardImage` es una promoción literal dentro del inventario único, sin props nuevas).

**A09 — Manejo de errores (CWE-209). ✅** `httpserver.WriteError` nunca envía la causa ni los `details` de un 500 al cliente (verificado en test y en vivo); los 401/403/404 son mensajes genéricos en español; el 400 lleva `details` por campo con las claves del contrato, sin nombres de constraints ni SQLSTATE (el 23514 se traduce a `details` por campo con `checkDetails`, `repository.go:267`). El SQLSTATE queda en el log con `request_id`.

**Validación de entrada (CWE-20). ✅** `MaxBytesReader` 1 MiB + `DisallowUnknownFields` + un único objeto JSON (**en vivo**: campo desconocido → 400). Etiquetas `validate` en el DTO + validación de dominio completa en el service sobre la fila **normalizada y fusionada** (`validateMinistry`: obligatorios, longitudes por campo, español base por campo, contacto excluyente celular-o-correo — **en vivo**: ambos → 400 —, contacto exige nombre, enlace https sin espacios/control, imagen con alt, `imageFile` con patrón del servidor). La BD revalida con 12 `CHECK` (000010) y `checkDetails` traduce la violación a 400 con `details` (nunca 500 con detalle interno). El PATCH **revalida la fila fusionada completa** (mejora respecto al B1 de F4).

**Auditoría (FR-017/SC-011). ✅** Toda mutación inserta su fila(s) `ministry.*` en la **misma transacción** (`withTx` + `insertActions`, o el cambio y su registro, o nada), con la regla de códigos de A6-10 (`ministry.create/update/delete/publish/unpublish/leader.show/leader.hide/image.upload`); las denegaciones y rechazos son best-effort **sin transportar el cuerpo** (verificado en vivo: filas `denied`/`failure` con `target_label` sin datos del encargado). El registro de códigos sigue cerrado por el `CHECK` de la BD (000011). Las lecturas denegadas no dejan fila (coherente con F2/F3/F4).

**Bilingüe y de superficie (FR-007/FR-008). ✅** `lang` resuelto en el servidor con fallback `en → es` por campo; el DTO público omite claves sin dato (0 líneas vacías). La redirección silenciosa de `/ministerios` sin publicados es decisión del frontend sobre su propia consulta (`MinisteriosPage.tsx:47-50`, `<Navigate to="/" replace />`), sin exponer nada: la API pública responde 200 con solo `lang` (patrón SC-010, test `TestGetMinisteriosHandlerOmitsItemsWhenEmpty`).

---

## Recomendaciones (deuda, no bloqueantes)

1. **B1 (dev-backend)**: comprobar la existencia de `imageFile` en el `Store` al guardar (cierra el enlace débil entre módulos y evita imágenes públicas rotas). Podría subir como issue de deuda técnica compartido con F4 (mismo patrón).
2. **B2 (dev-frontend)**: opcionalmente centralizar las contraseñas de e2e como constantes únicas en `e2e/helpers.ts`.
3. Vigilar el volumen de `admin_actions` (I2); si crece, considerar agregación o muestreo de denegaciones repetidas.
4. Cuando exista despliegue con TLS: añadir CSP (y HSTS) al frontend público (recomendación heredada de F4, aún pendiente a nivel de proyecto).

## Pruebas ejecutadas por el revisor

- `govulncheck ./...` (backend): **0 alcanzables** (1 informativa no llamada, I1). · `npm audit --audit-level=high` (frontend): **0**.
- `go vet ./...` + `go test ./internal/ministerios/...`: **ok**. Las de integración (`-tags=integration` + PostgreSQL) ya ejecutadas en verde por T637 según `estado.md`; no reejecutadas por el revisor para no tocar la base compartida del stack local (el comportamiento está además verificado en vivo, §Privacidad).
- Greps: SQL dinámico (**0**), `dangerouslySetInnerHTML`/`innerHTML`/`eval`/`document.write` (**0**), secretos (solo B2), `target=_blank` sin `noopener` (**0**), `href` dinámicos sin validar (solo `WhatsappJoinButton`, https validado en servidor).
- **Auditoría en vivo** (stack `make up` de la rama): 401 en las 5 rutas de panel sin sesión; 403/200 por cruce de permisos en ambas direcciones (cuentas `sin-ministerios`/`solo-ministerios` del e2e); CSRF 403 sin token; ciclo completo de privacidad del encargado (alta oculta → draft invisible → publicado sin `leader` → interruptor encendido → `leader` visible → borrado); borrador con ficha visible invisible; validaciones 400 de `javascript:`/`data:`/`http:`/traversal/doble contacto/campo desconocido; logs y `admin_actions` sin el contacto; HTML del frontend sin el contacto; CORS sin ACAO para origen no permitido; media inexistente 404. Los datos de prueba del revisor se limpiaron (el ministerio de prueba se eliminó con DELETE y el fixture modificado se restauró a su valor original; las filas de auditoría `ministry.*` resultantes son el comportamiento esperado del registro en un entorno local de pruebas).

## Nota de alcance

- El flaky `frontend/e2e/qa-bugs-004.spec.ts` es preexistente de F3/F4 y **no bloquea F6** (defecto conocido y ajeno a este diff).
- El revisor **no modificó código**: el único archivo escrito por el revisor es este informe.

---

## Ciclo 1 — re-validación (2026-10-13)

Tras este informe se aplicaron tres correcciones sobre superficies auditadas aquí: **1e0d9e5** (contrato: `whatsappUrl` anulable + `minProperties: 1` en `MinisterioPatch`), **6328cff** (frontend: el formulario manda `null`, nunca `""`) y **2d75367** (backend: `sortOrder` acotado a int32). Se re-auditó el diff de las tres y se re-verificó en vivo lo esencial contra el stack reconstruido con las correcciones (backend :8080, frontend :5173, PostgreSQL :5432). **Confirmado que el binario corriendo lleva el fix**: `sortOrder: 4294967296` → 400 con `details.sortOrder`, comportamiento que solo existe tras 2d75367.

### Análisis de los tres diffs

**1e0d9e5 — `whatsappUrl: [string, "null"]` + `minProperties: 1` (contrato). No abre vía nueva.**

- El inline (`type: [string,"null"]` + `pattern: ^https://[^\s]+$` + `maxLength: 500`) es **equivalente o más estricto** que el `$ref: WhatsappUrl` que sustituye en `MinisterioInput`/`MinisterioPatch`/`MinisterioAdmin`: mismo patrón, misma longitud, y el `null` solo representa «sin enlace». La **superficie pública no se toca**: el DTO público (`MinisteriosPublicos`) sigue con el `$ref` estricto y la clave **se omite** cuando no hay enlace (`omitempty` sobre `*string`) — el visitante jamás recibe un `null` ni un enlace sin validar.
- **`null` no puede transportar un enlace inseguro**: `null` no es cadena. En el alta, `encoding/json` decodifica `null` en el `string` del DTO como no-op (→ `""` → `normalizeOptional` → nil, «sin enlace»); en el PATCH, `Optional.UnmarshalJSON` marca set+null → `normalizeOptional("")` → nil (borra). El único camino hacia la columna es una **cadena** que debe pasar tres barreras intactas: el tag `url` del DTO (solo `https` con host), la regla 4 de `validateMinistry` (sobre la fila normalizada o **fusionada** tras el PATCH) y el `CHECK` de la BD (`whatsapp_url ~ '^https://[^[:space:]]+$'`, que deja pasar `NULL` — semántica SQL correcta para «sin enlace»). **Verificado en vivo**: `javascript:alert(1)`, `data:text/html,<script>`, `http://ejemplo.com`, `vbscript:`, `https://` sin host y enlace con espacio → **400** `details.whatsappUrl` en el alta; `javascript:` **por PATCH** → **400** (la revalidación de la fila fusionada sigue viva); `whatsappUrl: null` por alta → **201** sin enlace, y por PATCH → **200** borra el enlace.
- **`minProperties: 1` no permite evadir validaciones**: es un mínimo (prohíbe el PATCH `{}`, no habilita nada). El servidor nunca confió en el contrato para esto: ya rechazaba el cuerpo vacío con 400 «No hay cambios que guardar» (`hasChanges()`, `service_admin.go:106-112,181-183`; **verificado en vivo**: PATCH `{}` → 400). Todo PATCH que pasa se fusiona y **revalida completo** (A6-3), así que no existe combinación «un solo campo» que deje fuera una regla. Nota de coherencia: los campos no anulables (`sortOrder`, `leaderVisible`) siguen como `*T`: un `null` explícito en ellos cuenta como ausente (→ 400 si va solo), consistente con el contrato que no los declara anulables.
- `go.mod`: `gopkg.in/yaml.v3 v3.0.1` pasa de indirecta a directa — misma versión, ya estaba en el módulo; sin módulos ni versiones nuevos (`govulncheck` re-ejecutado: 0 alcanzables). El nuevo `contract_test.go` es código de prueba que solo lee el YAML del contrato.

**6328cff — el formulario manda `null` (nunca `""`). No abre vía nueva.**

- `nullable()` recorta y manda `null` cuando no hay valor — espejo exacto del `normalizeOptional` del servidor (antes mandaba `""`, que el servidor también normalizaba a nil: equivalente en persistencia). El valor **no vacío** viaja como cadena y sigue validado igual (tag `url` del DTO + regla 4 del service + `CHECK` de BD + `isSecureUrl` en el cliente, espejo del tag del servidor). Con el servidor reconstruido, el payload exacto del formulario se probó en vivo: POST `whatsappUrl: null` → **201** con `"whatsappUrl":null` (sin enlace); PATCH `{"whatsappUrl":null}` → **200** y el enlace queda borrado. Las pruebas nuevas (`ministryForm.test.tsx`, Vitest+MSW) no añaden credenciales (B2 intacto).

**2d75367 — cota de `sortOrder` (int32). Restricción pura; cierra I4.**

- `validateMinistry` añade la regla 6 (`MinSortOrder`/`MaxSortOrder` = rango int32), que cubre el alta (`ministryFromInput`) y el PATCH (fila fusionada) — antes un valor fuera de rango se **truncaba en silencio** al guardar (defecto de integridad de datos, no un hueco de seguridad: nunca se guardó fuera del rango de la columna). El contrato acota `minimum`/`maximum` espejo. **Verificado en vivo**: alta con `4294967296` → **400** `details.sortOrder`; PATCH con `9999999999` → **400**. La columna y la migración 000010 no cambian: cero superficie nueva.

### Re-verificación en vivo (stack reconstruido, 2026-10-13)

- **Permisos (A01)**: sin sesión → **401** en GET/POST `/admin/ministerios`, PATCH/DELETE `/admin/ministerios/{id}` y POST `/admin/ministerios/imagenes`. Cuenta con permiso ajeno (`actividades`) → **403** en GET/POST de ministerios («No tienes permiso para acceder a este módulo», genérico); cuenta `solo-ministerios` → **403** en `/admin/eventos` (cruce en ambas direcciones). **CSRF**: POST con sesión válida y sin `X-CSRF-Token` → **403**.
- **Privacidad del encargado (FR-026/SC-014), con dato real sembrado por el revisor** (publicado, `leaderVisible=false`, `Roberta Telefono C1` / `+506 7777-7777`, con enlace https válido):
  - `GET /api/v1/ministerios` (sin sesión): el ministerio aparece **sin clave `leader`**; **0 apariciones** del nombre y del celular en el cuerpo crudo; el enlace https se sirve tal cual.
  - `GET /api/v1/portada/ministerios` (sin sesión): **0 apariciones** del contacto.
  - `GET /` y `GET /ministerios` del frontend (:5173): **0 apariciones** en el HTML servido.
  - Interruptor ON → `leader:{name,phone}` sí sale (mecanismo vivo, prueba no vacua); OFF → 0 de nuevo.
  - Logs del backend durante las pruebas (627 líneas): **0 apariciones** del contacto (solo `request_id/method/path/code/status/duration_ms`).
  - `admin_actions` de las operaciones de prueba: `target_label` siempre `Ministerios · <nombre del ministerio>`, **nunca** datos del encargado; los rechazos quedan como `ministry.update · failure` **sin cuerpo** (I3 re-verificado con filas propias del revisor).
  - Limpieza: DELETE del ministerio de prueba → 204; 0 apariciones públicas y 0 filas en la BD tras la limpieza.
- **Dependencias (A06)**: `govulncheck ./...` → **0 alcanzables** (1 informativa en módulo requerido no llamado, idéntica a I1); `npm audit --audit-level=high` → **0**. `go test ./internal/ministerios/...` → ok (incluidas las pruebas nuevas de contrato y de rango).

### Estado de los hallazgos tras el ciclo 1

- **B1 (enlace débil `imageFile`)**: sin cambios — los tres commits no tocan esa vía (la regla 5 de `validateMinistry` y `storage.ValidName` quedaron intactos). Sigue **Baja, «a verificar», no bloqueante**; la recomendación de deuda se mantiene.
- **B2 (contraseñas efímeras de e2e)**: sin cambios; `ministryForm.test.tsx` no añade credenciales. Sigue **Baja (falso positivo probable)**.
- **I1/I2/I3**: sin cambios (govulncheck re-ejecutado con el mismo resultado; las operaciones de prueba del revisor insertaron las filas de auditoría esperadas, sin datos del encargado).
- **I4 (`sortOrder` sin cota)**: **cerrado** por 2d75367 — el servidor ahora rechaza el fuera de rango con 400 `details.sortOrder` y el contrato acota `minimum`/`maximum`. No cambia de severidad (ya era informativo): queda **resuelto**.

### Veredicto del ciclo 1: **APROBADO (ratificado)**

Las tres correcciones son de endurecimiento o pura coherencia de contrato — **ninguna relaja una barrera del servidor**, y todas las barreras auditadas (control de acceso, CSRF, https del enlace, validación de entrada, privacidad del encargado, auditoría, manejo de errores, dependencias) se re-verificaron en vivo intactas. B1/B2 siguen siendo bajos no bloqueantes. Sin hallazgos nuevos.

*El revisor no modificó código: el único archivo escrito en este ciclo es esta sección del informe.*
