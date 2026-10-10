# Revisión de seguridad — F3 Portada e información general · Ciclo 2 (re-validación)

**Fecha**: 2026-10-10 · **Rama**: `003-portada-info-general` · **Auditor**: `seguridad` (solo
lectura; este reporte es la única escritura). **Objeto**: re-validación del bucle de corrección 2
contra mi informe anterior (`revision-2026-10-10-seguridad.md`, veredicto **APROBADO** con 5 MENORES
S1–S5). El informe del ciclo 1 se commiteó junto a los fixes en `1e317f1`; el estado de código que
audité era el anterior a `d5d4610`, así que el diff efectivo desde mi auditoría es el que evalúa
este ciclo.

**Cambios evaluados** (diff completo `062fcf0..HEAD`, verificado que no hay nada más):

| Commit | Cambio | Contexto |
|---|---|---|
| `d5d4610` | **B1**: nuevo `backend/internal/portada/optional.go` (tipo `Optional[T]` con `UnmarshalJSON`) + DTOs/`merge*`/`hasChanges` del PATCH en `service_admin.go` (+ pruebas) | Cerraba el B1 de código del ciclo 1, que yo registré como **S3** |
| `fc96185` | **I1**: frontend, presentación del horario en a.m./p.m. localizado (`schedule.ts`, i18n, `ScheduleSection.tsx`) | Solo presentación; dato del contrato en 24 h intacto |
| `2b72138` | **A1**: frontend, objetivo táctil del selector de idioma (clases CSS en `LanguageSwitcher.tsx`) | WCAG 2.5.8 |
| `d98df37` | e2e permanente `portada-patch-null.spec.ts` (traza viva de QA versionada) | Prueba; crea y borra sus datos |

Nada de `routes.go`, `middleware/`, `platform/storage`, `platform/validate`, `platform/audit`,
migraciones, `docker-compose.yml` ni `.env.example` cambió desde mi auditoría del ciclo 1
(`git diff 062fcf0..HEAD --stat`: solo los archivos listados arriba).

## 1. Impacto del cambio B1 (`Optional[T]`) en seguridad — análisis del nuevo camino

**Superficie**: `Optional[T]` se usa **solo** en los dos DTOs de entrada del PATCH
(`ScheduleItemPatch` y `WhatsappChannelPatch`; grep de `Optional[`: 14 coincidencias, todas en
`optional.go` y esos dos DTOs, siempre `Optional[string]`). Nunca aparece en un DTO de salida:
sin riesgo de fuga de estado interno al serializar (sus campos son no exportados; se usaría como
`{}` en el peor caso, y no ocurre).

1. **`DisallowUnknownFields` sigue operante (CWE-20).** El decodificador de `decodeAndValidate`
   (`handler.go:130-151`) aplica `DisallowUnknownFields` **al nivel del DTO**; como los DTOs no
   implementan `json.Unmarshaler` (solo lo implementan sus campos `Optional`), el decoder recorre
   las claves del objeto y sigue rechazando las desconocidas antes de llegar al campo.
   El `UnmarshalJSON` interno usa un `json.Unmarshal` fresco **sin** esa restricción, pero `T`
   es siempre `string` (un escalar no tiene campos desconocidos): no hay relajación posible.
   **Verificado en vivo**: `PATCH {"foo": null}` → **400 `invalid`** (P1, §2).
2. **Payloads malformados y de tipo incorrecto (CWE-20).** El `UnmarshalJSON` de-califica tipos
   no string vía `json.Unmarshal(data, &v)`: `{"endTime": 123}` y `{"endTime": {}}` → **400
   `invalid`** (P2/P3, con fila `failure` en auditoría). El `MaxBytesReader` del body, el chequeo
   de «un único objeto JSON» y `validate.Struct` del decoder siguen intactos (no tocados).
3. **`null`/ausente no permite saltarse validaciones (CWE-20/862).**
   - Los campos **obligatorios** (`nameEs`, `placeEs`, `startTime`, `dayOfWeek`, `kind`,
     `destination`, `network`, `url`, `publicationState`, `sortOrder`) siguen siendo `*T`: con
     `encoding/json`, `null` deja el puntero en nil → se trata como **ausente** → no se tocan.
     En vivo: `{"nameEs": null}` → **400 «No hay cambios que guardar»** y el nombre quedó intacto
     (P4); vaciar un obligatorio con `""` sigue en **400** con `details.nameEs` «obligatorio» (P5).
   - Los campos **anulables** (`endTime`, `nameEn`, `descriptionEs/En`, `placeEn`) con `null` →
     `Value()` = `""` → pasan por los mismos normalizadores de siempre:
     `normalizeEndTime("")` → nil (quita la hora de fin); `normalizeOptional("")` → limpia la
     traducción. **Un valor** (no `null`) sigue validándose igual: `"25:00"` → 400 formato (P6);
     la validación cruzada sobre la **entidad fusionada** (`updated.EndTime != nil &&
     *updated.EndTime <= updated.StartTime` → 400, `service_admin.go:788-793`) sigue intacta
     (P7: `endTime "09:00"` con `startTime 10:00` → 400 «posterior a la hora de inicio»).
     No hay vía para dejar un obligatorio vacío ni un horario inconsistente.
4. **`hasChanges` trata `null` como cambio.** `{"endTime": null}` ya no responde 400 «No hay
   cambios»: ahora entra al merge, aplica la limpieza y deja su fila `home.schedule.update`
   (`analyze` M5: vaciar es un cambio). **Verificado en vivo** (P12/§2): 200, `endTime` ausente
   persistido tras refetch, `nameEn` **intacto** (ausente = no tocado) y estado `published`
   intacto; fila `home.schedule.update` `result=success` presente.
5. **Autorización/CSRF/auditoría sin cambios.** `Optional` vive después de `decodeAndValidate`,
   ya dentro de la cadena `authn → guard → authz(portada) → CSRF`: no toca identidad, permisos
   ni transacción. `UpdateService`/`UpdateWhatsappChannel` conservan el registro transaccional y
   `recordContentFailure` en cada 400 de dominio.

**Impacto de I1/A1 (frontend)**: puramente de presentación. `formatClockTime` aplica una regex a
un dato del contrato (`HH:MM` validado en servidor) y lo renderiza como **texto de React**
(escapado); las claves i18n son estáticas; A1 son clases CSS (`min-h-11`, `cursor-pointer`). Sin
nuevas llamadas de red, sin `dangerouslySetInnerHTML`/`innerHTML` (grep: **0**), sin manejo de URLs.
**Sin impacto en seguridad.**

## 2. Pruebas dinámicas (stack real, 2026-10-10: `docker ps` sano, `/healthz` 200)

Cuenta ADMIN de e2e (`ana@ejemplo.com`); servicio de prueba creado publicado con `endTime 12:00`
y `nameEn`, limpiado al final (DELETE 204).

| # | Prueba | Resultado |
|---|---|---|
| P1 | `PATCH {"foo": null}` (campo desconocido) | **400 `invalid`** — `DisallowUnknownFields` intacto con `Optional` ✓ |
| P2 | `PATCH {"endTime": 123}` | **400 `invalid`** — tipo rechazado dentro de `Optional.UnmarshalJSON` ✓ |
| P3 | `PATCH {"endTime": {}}` | **400 `invalid`** ✓ |
| P4 | `PATCH {"nameEs": null}` (obligatorio, no anulable) | **400 «No hay cambios que guardar»**; el nombre quedó **intacto** — `null` = ausente, sin bypass ✓ |
| P5 | `PATCH {"nameEs": ""}` | **400** con `details.nameEs` «obligatorio» — no se puede vaciar un obligatorio ✓ |
| P6 | `PATCH {"endTime": "25:00"}` | **400** formato HH:MM — `null` no relaja la validación de valores ✓ |
| P7 | `PATCH {"endTime": "09:00"}` (inicio 10:00) | **400** «posterior a la hora de inicio» — cruce intacto sobre la entidad fusionada ✓ |
| P8 | `PATCH {"endTime": null}` sobre id inexistente | **404 `not_found`** ✓ |
| P9 | `PATCH` sin sesión | **401 `unauthenticated`** (A01) ✓ |
| P10 | `PATCH` con sesión, sin `X-CSRF-Token` | **403 `forbidden`** «El token de seguridad…» (CSRF) ✓ |
| P11 | `PATCH` con cuenta sin permiso `portada` (`sara-auditoria-seguridad@ejemplo.com`) | **403** genérico «No tienes permiso para acceder a este módulo» + fila `denied` en auditoría (A01) ✓ |
| P12 | `PATCH {"endTime": null}` (camino real de B1) | **200**; `endTime` **ausente y persistido** tras refetch; `nameEn` y `publicationState` **intactos**; fila `home.schedule.update` `success` ✓ |
| P13 | `GET /api/v1/portada?lang=es` | el servicio publicado aparece **sin** `endTime` y el DTO público solo expone `dayOfWeek/id/name/place/startTime` — sin `publicationState` ni pares crudos (FR-013) ✓ |
| P14 | Trazabilidad | cada una de mis 11 peticiones dejó su fila: 8 `failure` (P1–P7 más el 404; decodificación con etiqueta UUID vía `RecordRejectedBestEffort`, dominio con nombre vía `recordContentFailure`), 1 `denied` (P11), 1 `success` (P12), 1 `create` — cuadre exacto ✓ |

## 3. Verificaciones estáticas y de herramientas

| Comando/chequeo | Resultado |
|---|---|
| `make security` → `govulncheck ./...` | **0 vulnerabilidades que afecten al código** (1 en módulos requeridos no llamada — idéntico al ciclo 1, fuera del criterio §IV) ✓ |
| `make security` → `npm audit --audit-level=high` | **0 vulnerabilidades** ✓ |
| `go test ./internal/portada/...` y `go test -count=1 -run 'TestPatch\|TestUpdate…'` | todo `ok` (incluye las 4 de regresión de B1: `service_patch_null_test.go` y `TestPatchScheduleNullClearsAndAudits`) ✓ |
| `grep dangerouslySetInnerHTML\|innerHTML frontend/src` | **0** ✓ |
| `grep` SQL dinámico (`Sprintf`+`SELECT/INSERT/UPDATE/DELETE`) en `backend/` | **0** ✓ |
| `grep` secretos `password\|secret\|api_key\|token` con valor literal | **0 reales**: las 70 coincidencias son fixtures de prueba, nombres de variables de entorno y el password local de testcontainers (`testutil/containers.go`) — todo preexistente y sin cambios ✓ |
| `Optional[` en backend | solo `optional.go` + los 2 DTOs de PATCH, siempre `Optional[string]` ✓ |
| Diferencial de archivos desde mi auditoría | solo `optional.go`, `service_admin.go` (+tests), frontend de presentación y e2e — **cero cambios** en rutas, middleware, storage, validadores, auditoría, migraciones, compose, `.env.example` ✓ |

## 4. Estado de los hallazgos del ciclo 1 (S1–S5)

| ID | Estado | Nota |
|---|---|---|
| **S3** — PATCH no distinguía `null` de ausente | **RESUELTO** por `d5d4610` (B1) | Verificado en vivo (P12) y con pruebas de regresión; la limpieza ahora aplica, se persiste y se audita (`home.schedule.update`). Sin abrir ningún hueco (§1). |
| **S1** — `logoFile`/`coverImageFile` sin patrón de nombre generado | **Sigue abierta (MENOR)** | `identityImageRules` (`service_admin.go:908-929`) sin cambios: longitud ≤120 + texto alternativo, sin `storage.ValidName`. El diff no la tocó; sin nueva exposición (la descarga valida el patrón y los no referenciados dan 404, D9/D10/D15 del ciclo 1). |
| **S2** — subidas rechazadas sin fila `failure` | **Sigue abierta (MENOR)** | `UploadImage` (`service_admin.go:322-360`) sin cambios: archivo vacío, firma y tamaño devuelven 400 sin `recordContentFailure`. (Los rechazos de **decodificación** sí dejan fila vía `RecordRejectedBestEffort`, verificado en P1–P3.) |
| **S4** — `nosniff` solo en la descarga de imágenes | **Sigue abierta (MENOR)** | Sin cambios; hardening diferible. |
| **S5** — huérfanos acumulables en `uploads_data` | **Sigue abierta (MENOR)** | Sin cambios; diseño aceptado (R3-8), limpieza anotada para F4+. |

## 5. Hallazgos nuevos

**BLOQUEANTE**: ninguno. **IMPORTANTE**: ninguno.

**MENOR (nota, sin acción requerida para F3)**: un PATCH `{"nameEn": null}` sobre un elemento
cuya traducción **ya estaba vacía** marca `dataChanged` y deja fila `home.schedule.update`
aunque el valor efectivo no cambia (ruido de auditoría, no confusión: el `targetLabel` y la
acción son correctos). Es la decisión explícita de `analyze` M5 («vaciar es un cambio») y el
coste de implementarla con presencia real; sin impacto de seguridad (superficie ya autorizada,
acción exacta, `result` verídicamente `success`). Se registra como **S6 (MENOR, opcional)**:
si se desea, distinguir «null que ya estaba vacío» de «null que vacía» en una deuda futura.
**No bloquea.**

## 6. Residuos de esta re-validación

- Filas de auditoría de mis 11 peticiones sobre el servicio de prueba «Seg-C2 null»
  (8 `failure`, 1 `success`, 1 `denied`, 1 `create`) + 1 `home.schedule.delete` de la limpieza
  (DELETE 204). Mismo patrón que dejan los e2e.
- El servicio de prueba **fue eliminado**; no quedan imágenes huérfanas nuevas (no hubo subidas).
- El contenido público quedó idéntico al inicio (el servicio creado se borró; la portada pública
  nunca mostró nada de borrador).

## 7. Veredicto

**APROBADO** — 0 hallazgos BLOQUEANTES y 0 IMPORTANTE. El fix B1 (`Optional[T]`) cierra S3 **sin
abrir ningún hueco**: `DisallowUnknownFields` operante (P1), tipos malformados rechazados (P2/P3),
`null` no vacía obligatorios ni salta validaciones de formato, rango ni cruce (P4–P7),
autorización/CSRF/404 intactos (P8–P11), limpieza persistida y auditada (P12) y vía pública sin
filtraciones (P13). I1/A1 son presentación sin superficie nueva. `make security` en verde y
pruebas de regresión en verde.

**Estado de menores**: S3 resuelta; S1, S2, S4 y S5 siguen abiertas como deuda registrada (S1/S2
recomendadas para el mismo PR, como en el ciclo 1); S6 nueva, opcional.

**Nota para el orquestador**: el APROBADO de seguridad coincide ahora con los APROBADO de QA y
código del ciclo 2. Desde la dimensión de seguridad, la rama está en condiciones de abrir el PR;
S1/S2 (pequeñas, de higiene dentro del perímetro autorizado) quedan a criterio del PR o como
deuda explícita, y S4/S5/S6 como deuda registrada.
