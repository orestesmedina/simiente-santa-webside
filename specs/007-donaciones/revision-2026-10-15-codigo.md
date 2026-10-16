# Revisión de código — F7 Ofrendas y diezmos (Donaciones)

**Fecha**: 2026-10-15 · **Fase**: 5 (validar) · **Revisor**: `revisor-codigo` (solo lectura; este informe es la única escritura)
**Alcance**: `git diff main...HEAD` (rama `007-donaciones`, 51 commits desde `26d2893`, 133 archivos, +25 000/−43) contrastado con `spec.md` (FR-001…FR-025, SC-001…SC-014, Q1–Q11), `plan.md` (P7-1…P7-21, R7-1…R7-15), `ux.md`, `tasks.md` (T701–T738), `data-model.md`, `contracts/openapi.yaml` (0.8.0) y `.specify/memory/constitution.md`.

## Veredicto

**APROBADO CON CAMBIOS MENORES** — sin hallazgos CRÍTICOS ni ALTOS. La implementación corresponde a lo aprobado en las 21 decisiones del plan: IBAN por **estructura + longitud exacta por país sin mod-97** (`numero.go` con registro embebido, `CR`=22), SINPE **6/7/8 con prefijo normalizado a `+506`**, `institution` **no traducible** (columna única), duplicados por **tipo+número normalizado con `exclude_id`** → 409, publicación **por elemento y por texto** (estado propio de cada uno), orden agrupado `CASE kind → sort_order → id` idéntico en público y panel, **tope 5 en el bloque / 200 en la sección** resueltos en el servidor, y presencia de la navegación **desde el documento de la sección** (P7-8), con redirección silenciosa de `/ofrendas` en el frontend (P7-21). Auditoría transaccional con la regla exacta de códigos, **0 dependencias nuevas**, `schema.d.ts` regenerado desde el contrato y **0 `any`** en `frontend/src`. Quedan **2 MEDIO/BAJO sustantivos** (uno de auditoría: `PutTexto` sin registro cuando no cambia nada) y **7 BAJO/SUGERENCIA**.

## Evidencia verificada en esta revisión

| Comprobación | Resultado |
|---|---|
| `gofmt -l`, `go vet ./...`, `go build ./...` (backend) | limpios |
| `go test ./...` (unitarias, 22 paquetes) | en verde |
| Cobertura de `internal/donaciones/service*.go` (perfil propio, no bitácora) | **89,58 %** (318/355 statements): `service.go` 92,7 %, `service_admin.go` 83,3 %, `service_audit.go` 100 %, `service_public.go` 97,5 % — la cifra reportada (89,6 %) es exacta y supera el ≥80 % de §III |
| `npx tsc --noEmit`, `npx eslint .` (frontend) | limpios |
| `npx vitest --run` (frontend) | **777/777 en verde** (113 archivos). Con **2 «unhandled errors» de arranque de workers** (`vitest-pool: Timeout waiting for worker to respond`) reproducidos también en un subconjunto: es **flake de infraestructura** de este entorno (WSL lento), no de las pruebas; ver Observaciones |
| Pruebas desactivadas (`skip`/`only`/`xit`/`todo`) | ninguna (backend ni frontend ni e2e) |
| `any` injustificado en `frontend/src` (fuera de tests y `schema.d.ts`) | ninguno |
| Diff de `go.mod`/`package.json` | **vacío** (0 dependencias nuevas, cumplido) |
| `backend/api/openapi.yaml` v0.8.0 vs snapshot `specs/007-donaciones/contracts/openapi.yaml` | **coherentes**: las 14 definiciones F7 comparadas por script; solo diferencias de comilla YAML y una línea en blanco |
| `error.code` del contrato | sin crecer (solo prosa nueva); `AdminActionItem.action` +8 códigos `donation.*` |
| Migraciones previas | **no editadas**: solo se añaden `000012`/`000013` (§VI) |
| Tareas T701–T737 | todas `[X]` con código y prueba; T738 (docs+PR) pendiente a propósito; commits por tandas con marcadores de tareas coherentes con `tasks.md` |

## Hallazgos

### MEDIO

**M1 — `PutTexto` aplica la escritura sin ninguna fila de auditoría cuando no cambia nada (borde de FR-017/SC-011).**
- `backend/internal/donaciones/service_admin.go:293-312` — con el texto existente y un PUT idéntico, `dataChanged` y `stateChanged` son ambos `false` y `actions` queda vacío, pero **igual se llama** `UpsertDonationText` (`:308`), cuya consulta hace `DO UPDATE SET … updated_at = now()` (`backend/internal/db/queries/donation_texts.sql`, `UpsertDonationText`): la fila se reescribe —`updatedAt` cambia— con **cero registros** de auditoría. El edge case de la spec es literal: «nunca debe aplicarse una operación sin registro, ni registrarse una operación que no se aplicó» (FR-017).
- Es además **inconsistente con el hermano**: `UpdateDonacion` (`service_admin.go:146-148`) hace *early return* sin escribir ni auditar en el mismo supuesto, fijado por `TestUpdateDonacionNoChangesReturnsCurrent` (`service_admin_test.go:293-310`, exige `len(repo.committed) == 0`). `PutTexto` no tiene ese caso de prueba.
- **Corrección propuesta** (`dev-backend` + su prueba): en `PutTexto`, si `found && !dataChanged && !stateChanged`, devolver `donacionTextoAdminFrom(current)` sin escribir (misma mecánica que `UpdateDonacion`), y añadir el test espejo; o, si se prefiere que todo PUT deje huella, registrar siempre `donation.text.update`. Impacto real hoy: bajo (solo `updated_at`), pero es la única rendija del régimen de auditoría que la revisión encontró.

### BAJO / SUGERENCIA

**B1 — `RecordAction`, un método del repository, vive en el archivo del service (se repite M2 de F6).**
`backend/internal/donaciones/service.go:480-482` implementa `func (r *repository) RecordAction` en el archivo cuyo propio comentario dice que «la persistencia y la auditoría atómica viven en el repository». En F6 ya se señaló (M2) y el patrón se copió. **Corrección** (`dev-backend`): moverlo a `repository.go`/`repository_donaciones.go`. Es el cuarto módulo que lo arrastra: conviene corregirlo ya que el precedente está escrito.

**B2 — `PATCH {}` es 200 no-op y `DonacionPatch` no declara `minProperties: 1`, en contraste con F3/F4/F6.**
`backend/api/openapi.yaml:6057-6097` (`DonacionPatch`, sin `minProperties`) frente a `:4345`, `:4456`, `:4530`, `:5722` (`ScheduleItemPatch`, `WhatsappChannelPatch`, `SocialLinkPatch`, `MinisterioPatch` **sí** lo declaran) y al comportamiento de F6 (PATCH vacío → 400). Aquí es **deliberado y está fijado por prueba** (`TestUpdateDonacionNoChangesReturnsCurrent`, que documenta «el contrato no declara minProperties»), pero el visitante/mismo equipo recibe un comportamiento distinto según el módulo. **Corrección** (`dev-backend`, contrato + service + test): alinear a `minProperties: 1` → 400 como el resto, o dejar constancia explícita de la excepción en la prosa del contrato. Es coherencia de plataforma, no un defecto funcional.

**B3 — `null` explícito en campos no anulables del PATCH: dos comportamientos distintos.**
`backend/internal/donaciones/service_admin.go:326-331` (`kind`, `number`), `:355-357` (`sortOrder`) y `:361` (`publicationState`) **ignoran** el `null` en silencio (`.Set() && !.Null()`), mientras `:334-337` (`institution`) lo convierte en `""` y responde 400. El contrato los declara no anulables (`type: string`/`integer`), así que un cliente conforme nunca lo envía; pero la API forzada recibe *no-op 200* donde podría recibir un 400 explícito. **Corrección** (`dev-backend`): rechazar con `400 invalid` + `details` el `null` sobre campos no anulables (o ignorarlo en todos por igual, con prueba que lo fije).

**B4 — `checkDetails` se apoya en los nombres autogenerados por PostgreSQL para los `CHECK` sin nombre.**
`backend/internal/donaciones/repository.go:292-339` mapea `donation_methods_check`, `donation_methods_check1`, `donation_methods_check2`, `donation_methods_institution_check1`… El comentario es honesto («el nombre no es bonito pero sí estable, porque ninguna migración aplicada se edita») y la integración real **fija el mapping disparando cada CHECK** (`repository_migrations_integration_test.go:334+`, `TestIntegrationDonationChecksFireInTheDatabase` con los nombres exactos). El fallo si una migración futura reordena o reemplaza CHECKs es **fail-safe**: vuelve al mensaje genérico sin `details`. **Corrección** (`dev-backend`, cuando valga la pena): nombrar explícitamente los CHECK en una migración nueva (sin tocar `000012`).

**B5 — Duplicados: carrera check-then-insert y sonda sin índice (trade-off aprobado en P7-10).**
`ExistsDonationMethodDuplicate` (`repository_donaciones.go:151-161` + `queries/donation_methods.sql`) se consulta antes de insertar, pero `donation_methods` **no tiene `UNIQUE (kind, number)`** —decisión expresa de P7-10 para conservar el default «se señala»—: dos altas simultáneas idénticas pueden ambas pasar el chequeo y ambas insertar (el `23505`→409 de `classify` nunca saltará en esta tabla). La sonda además corre sin índice sobre `(kind, number)` (solo hay `(kind, sort_order, id)` y `(publication_state)`). Impacto práctico mínimo (lista corta, panel con pocos actores), pero conviene dejarlo registrado en `estado.md` como default revisable y añadir el índice —o el `UNIQUE`— si el default cambia.

**B6 — Validación de `lang` triplicada entre handler y service.**
`backend/internal/donaciones/handler_public.go:36-46` y `:83-93` reimplementan la normalización y el mensaje de `normalizeLang` (`service.go:181-193`); el mismo texto vive en tres lugares. **Corrección** (`dev-backend`): que el handler pase el `lang` en crudo y deje la validación (única) en el service, que ya responde `400 invalid` con `details.lang`.

**B7 — Nomenclatura de identificadores (deuda transversal, no de F7).**
El dominio usa identificadores en español (`MetodoDonacion`, `DonacionAdmin`, `OfrendasPublicas`, `NormalizarIban`, `agruparIban`, `presentarNumero`…) frente a §V («Nombres descriptivos en inglés dentro del código»). Es **idéntico al precedente ya aprobado** de F3/F4/F6 (`PortadaPublica`, `MinisterioPublico`, `ProximamentePublica`…), con los helpers técnicos en inglés (`validateMethod`, `mergeDonacion`, `localizeMetodo`). No es un hallazgo contra F7: si se quiere alinear a la constitución, es una decisión de dirección técnica que afecta a todos los dominios.

## Puntos delicados de F7: dictamen

| Punto (decisión del plan) | Dictamen |
|---|---|
| **1. IBAN estructura + longitud, sin mod-97 (P7-3/Q3)** | **Correcto.** `numero.go:24-69`: regex de estructura + registro IBAN embebido por país (`CR`=22), guardado normalizado a los caracteres del número, sin mod-97 en ninguna parte (el comentario lo fija). Presentación en bloques **solo en el frontend** (`format.ts:agruparIban`, 4+4 con trim; quitar espacios devuelve el valor exacto) y `NumeroDonacion` mantiene la lectura de lector de pantalla = valor copiable (`aria-hidden` + `sr-only`). El espejo del cliente (`validation.ts:69-168`) replica la tabla y la regla; la autoridad es el servidor. Tabla de casos en `numero_test.go` y e2e `ofrendas-publica.spec.ts` con IBAN internacional. |
| **2. SINPE 6/7/8 con prefijo normalizado (P7-4/Q4)** | **Correcto.** `numero.go:82-99`: acepta con/sin `+506` (y `506…` de 11 chars), exige 8 dígitos con primer dígito 6/7/8 y publica `+506XXXXXXXX`. El `CHECK` de la BD (`'^\+506[0-9]{8}$'`) refuerza la forma. Coherente en cliente (`normalizarSinpe`). |
| **3. `institution` no traducible (P7-5/RG7-1)** | **Correcto y estructural**: columna única `institution`; solo `holder_es`/`holder_en` y los textos llevan par de idiomas (`000012`, `model.go`, `MetodoDonacionPublico`). `localizeMetodo` (`service_public.go:99-113`) solo resuelve `holder`; `kind`/`number`/`institution`/`currency` viajan idénticos (SC-006). El e2e de idioma comprueba la identidad de los datos. |
| **4. Duplicados tipo+número con `exclude_id` (P7-10/R7-9)** | **Correcto** (salvo la carrera sin `UNIQUE`, B5): la consulta filtra `kind = ? AND number = ?` sobre el número **normalizado** y excluye la fila editada (`AND (exclude_id IS NULL OR id <> exclude_id)`). Probado: `service_admin_test.go:167-232` («editar sin cambiar número no confía; cambiarlo a otro sí»), `repository_donaciones_integration_test.go:301` (`…DuplicateNormalizes`) y e2e `ofrendas-panel.spec.ts:315`. |
| **5. Publicación por elemento y por texto (P7-12/Q10)** | **Correcto**: `publication_state` en ambas tablas; `ListPublishedDonationMethods`/`ListPublishedDonationTexts` son las **únicas** vías públicas (`WHERE publication_state = 'published'`); el DTO público **no tiene campo de estado** (defensa estructural como la del encargado en F6). «Se publica al guardar»: `UpdateDonacion` con estado cambiado registra **una fila** `donation.publish`/`unpublish` además de la de datos, ambas en el mismo tx (`service_admin.go:161-176`). Textos con su estado propio (`PutTexto`, `textStateCode`). e2e `ofrendas-panel.spec.ts:374` y `:417`. |
| **6. Orden agrupado (P7-20/FR-004)** | **Correcto**: `ORDER BY CASE kind WHEN 'iban' THEN 0 ELSE 1 END, sort_order, id` en público **y** en panel (mismo orden para el visitante y el equipo, US2 esc. 10). `sortOrder` acotado a int32 en validación (`MinSortOrder`/`MaxSortOrder`, `service.go:282-284`) — el hallazgo M1 de F6 (truncamiento silencioso) quedó **corregido aquí**. Integración `…ListDonationMethodsOrdersAndFilters` y `…RespectsLimit`. |
| **7. Tope 5/200 en el servidor (FR-022/RG7-12)** | **Correcto**: `maxPortadaItems = 5` y `maxPublicItems = 200` como constantes del service (`service.go:53-60`), `LIMIT` en SQL, el contrato **no** expone parámetro de límite en el bloque y `GetPortadaOfrendas` comparte `localizeMetodo` (mismo DTO de ítem). Probado en service (`…CapsAtFiveAndKeepsOrder`, `…OmitsItemsWithoutMethods`) e integración (`…RespectsLimit`); e2e `portada-ofrendas.spec.ts:145` con **8 publicados → exactamente 5**, `:225` (menos → exactos) y `:256` (0 → bloque inexistente). Posición en portada: **una inserción** tras `MinisteriosBlock` y antes de `WhoWeAreSection` (`HomePage.tsx:103-110`), Q7. |
| **8. Navegación desde el documento de la sección (P7-8)** | **Correcto**: `useSiteLinks` consume `usePublicOfrendas` (compartido por `queryClient` con la página) y `hasPublicOfrendasContent` = ≥1 método **o** ≥1 texto (`usePublicOfrendas.ts:54-56`): la sección puede vivir con solo textos (P7-13) sin dejar la entrada huérfana, y el bloque no decide presencia. El patrón F4/F6 queda documentado como divergencia deliberada en el propio hook. |
| **9. Redirección silenciosa de `/ofrendas` (P7-21/Q9)** | **Correcta y mecánicamente igual que F6**: `OfrendasPage.tsx:42-49` — `SectionSkeleton` mientras `isPending`, `<Navigate to="/" replace />` **solo** con éxito y sin contenido, **ante error no redirige** (H1 + «Volver a intentar», `:67-76`, y con caché previa se conserva lo último válido, `:78-82`). El API responde 200 con claves omitidas (nunca `null`), como exige el plan. e2e `ofrendas-publica.spec.ts:393` y `ofrendas-accesibilidad.spec.ts:359` (foco tras la redirección). |
| **10. Auditoría (P7-11/R7-10)** | **Correcta salvo M1**: alta = **solo** `donation.create` (también naciendo publicado, `service_admin_test.go:52` exige `len==1` y código exacto); reordenar = `donation.update`; textos sin `create` (singleton) → `donation.text.update` incluso en el primer guardado; mutación + filas en la **misma transacción** y fail-closed (`insertActions`, `repository.go:71-78`; integración `…RollsBackWhenAuditFails`/`…WhenAuditCodeIsUnknown`); denegaciones → `RecordDenied` (`result='denied'`) y rechazos JSON/DTO → `RecordRejectedBestEffort` **sin cuerpo de petición** (`handler_admin.go:236-257`); `auditActionFromRoute` solo deja fila en escrituras (los GET no). Los 8 códigos viven en `platform/audit` y en el `CHECK` de `000013`, con up→down→up probado. |
| **11. Permisos sin tocar F2 (P7-6/Q8)** | **Correcto**: `AdminChain("donaciones")` (`routes.go:58`), permiso `donaciones` ya sembrado por `000002`; `TestIntegrationPermissionsCatalogUnchanged` fija el catálogo en 10 **también tras los down**. Frontend: `DONACIONES` en `lib/permissions`, `RequirePermission` en `/panel/donaciones`, y RG7-6 (actualización acotada de tests) es **legítimo**: `permissions.test.ts` sigue exigiendo la lista **exacta** con `toEqual` (más estricto), `roles.test.tsx` conserva su aserción de sello, `panel.test.tsx` **añade** casos sin tocar los existentes, `test/server.ts` añade handlers neutros documentados y `acceso.spec.ts` cambia «Donaciones» por «Grupos de conexión» como módulo aún sin construir (el escenario de F2 sigue vivo). |
| **12. Sin imágenes (P7-16/Q6)** | **Correcto**: 0 cambios en `media`, `platform/storage` e `IsFilePublished`; ninguna ruta de subida nueva; las fichas públicas no tienen zona de imagen (`DonacionCard`). |
| **13. Contrato 0.8.0 (P7-19)** | **Coherente**: delta → fusión aditiva, 9 operaciones (2 públicas + 7 de panel), snapshot = vivo (14 esquemas comparados por script), `error.code` sin crecer, `AdminActionItem.action` +8. Los únicos reparos son B2/B3 (`minProperties` y `null` en no-anulables). |

## Pruebas F2/F4/F6 modificadas por F7 (T725, «cambio acotado RG7-6»)

Revisadas con `git diff` archivo por archivo (`lib/permissions.test.ts`, `features/roles/roles.test.tsx`, `features/panel/panel.test.tsx`, `app/layout.test.tsx`, `features/panel/components/AccesosRapidos.test.tsx`, `test/server.ts`, `frontend/e2e/acceso.spec.ts`): **ninguna aserción se relajó ni se borró** — son actualizaciones de catálogo (recuentos/listas exactas que ahora incluyen `DONACIONES`, casos nuevos añadidos) y el cambio de `acceso.spec.ts` mantiene el escenario de F2 usando módulos F5/F9. **Dictamen: legítimo.** Recomendación (ya hecha en F4/F6): dejar constancia de RG7-6 en el registro del PR/`estado.md`.

## Pruebas: calidad

- **Backend (117 `Test*` en `internal/donaciones`)**: unitarias sobre un repository falso en memoria con aserciones de **comportamiento real** (números normalizados byte a byte, códigos de auditoría exactos, `PATCH` fusionando y `null` explícito **desde JSON**, orden y topes) e integración contra PostgreSQL real para lo que solo la BD puede probar (up→down→up de `000012`/`000013`, `CHECK`s disparándose con su nombre, permisos intactos, atomicidad y rollback, solo publicado en las vías públicas). Cubren también errores (`…RepositoryError`, `…RejectsInvalid`, `TestMutationFailureDoesNotRecordSuccess`).
- **Frontend (777)**: `validation.test.ts` con tabla de casos IBAN/SINPE espejo del dominio, hooks con handlers MSW, `OfrendasPage.test.tsx` fijando la redirección (REPLACE, esqueleto, **no** redirigir ante error), `useSiteLinks.test.tsx` con el predicado de presencia, i18n `es`/`en` con paridad de claves.
- **E2E (4 specs + seed)**: mapean US1–US6 y los SC duros (números exactos, tope 5, redirección sin destello, permisos, auditoría con etiquetas, 320 px, teclado). **No se ejecutaron en esta revisión** (requieren el stack levantado); su revisión es estática y queda para `qa-tester`/T737.

## Sin hallazgos (verificado y correcto)

- **Capas (§II)**: handlers sin SQL ni lógica de negocio; service sin `net/http` ni `pgx`; repository único que conoce pgx y traduce errores con `%w` (`wrap`/`classify`, `repository.go:240-353`). Rutas públicas sin sesión; panel con `AdminChain("donaciones")`.
- **Contrato antes del código (§II)**: el delta (commit `f80faa0`) precede a la implementación; `schema.d.ts` regenerado (`97b09fd`) y el cliente tipado sale solo del contrato.
- **SQL parametrizado (§IV)**: todo vía sqlc (`sqlc.arg`/`sqlc.narg`), sin concatenación; entrada validada también en el servidor (`validate.Struct` + reglas de dominio); `DisallowUnknownFields` y cuerpo acotado a 1 MiB.
- **Migraciones (§VI)**: `000012`/`000013` nuevas con up/down completos; `id/created_at/updated_at` en ambas tablas; índices explícitos; el down de `000013` es destructivo a propósito y documentado (patrón de `000006`/`000008`/`000011`).
- **Observabilidad (§VII)**: `log/slog` con logger por petición (`request_id`) en los fallos best-effort; `/healthz` preexistente sin cambios; cableado por `cmd/api/main.go` con `Recorder: donacionesSvc`.
- **Errores con contexto**: todos los `fmt.Errorf("…: %w", err)` con operación legible; mensajes al público sin información interna.
- **Tipos sin `any`** y TypeScript strict; funciones cortas; el inventario UI se reutiliza (`Table`, `PublishControls`, `Card`, `SectionSkeleton`…) sin copiar piezas (P7-18).

## Observaciones (no son hallazgos de código)

- **O1 — Flake de infraestructura en Vitest**: la suite completa (777/777) y un subconjunto (127/127) terminan en verde pero con 1–2 «unhandled errors» del pool de workers (`Timeout waiting for worker to respond`) en este entorno WSL. Conviene vigilarlo en CI (ajustar `pool`/`fileParallelism` si aparece); no afecta a la corrección del código.
- **O2 — `Optional[T]` llega a la 4.ª copia** (`internal/{portada,agenda,ministerios,donaciones}/optional.go`): F6 ya lo registró como M6 con «regla de tres cumplida»; el plan de F7 decidió no promoverlo a `platform/`. Queda como deuda conocida y documentada.

## Veredicto final

**APROBADO CON CAMBIOS MENORES** (→ **APRUEBA**). Sin bloqueantes. Antes de la entrega (T738/PR) se recomienda cerrar **M1** (auditoría de `PutTexto` sin cambios) y, si se quiere, B1–B3 como limpieza; B4–B7 pueden quedar registrados como deuda en `estado.md`.
