# Revisión QA — F9 Medios, revalidación del ciclo 1 (2026-10-18/19)

**Rol**: `qa-tester` · **Rama**: `009-medios-predicas-y-podcasts` · **Alcance**: comprobar que los **11 commits de corrección** del ciclo 1 (tras el informe de `revisor-codigo` con 3 ALTOS) no rompieron el comportamiento que aprobé en `revision-2026-10-18-qa.md`, más una re-verificación en vivo de las aserciones clave del ciclo 1.

## Veredicto: APRUEBA

Todas las suites pasan en mi propia corrida (unitarias de Go, integración de Go con PostgreSQL real, Vitest **946/946** —eran 930: **+16 de las pruebas nuevas de las correcciones**— y los **25/25 e2e de F9 con `--workers=1`**). Las **5 correcciones necesarias quedan comprobadas** (2 en vivo contra la API, todas con su prueba automatizada sin aserciones debilitadas). Encontré **un fallo reproducible**, pero su causa es una **carrera en el helper e2e `portadaContiene`** (infraestructura de pruebas), no de producto: el backend y el frontend se comportan correctamente — demostrado en vivo y con una spec de diagnóstico (`frontend/e2e/qa-ciclo2-debug.spec.ts`, creada por mí). Corregí el helper para esperar a que la consulta del bloque termine (sin debilitar la aserción: sigue exigiendo el texto visible tras cargar) y toda la batería quedó en verde.

## Suites ejecutadas por mí (encima del ciclo 1)

| Suite | Comando | Resultado |
|---|---|---|
| Backend unitarias | `cd backend && go test ./...` | ✅ ok |
| Backend integración (PostgreSQL real) | `go test -tags=integration -count=1 ./internal/medios/...` | ✅ ok (0,5 s) |
| Frontend unitarias | `npx vitest run` | ✅ **946/946** en 131 archivos (131 vs 129: 2 archivos nuevos) |
| E2e de F9 | `npx playwright test --config e2e/playwright.config.ts --workers=1 medios- portada-medios` | ✅ **25/25** tras corregir el helper |
| Lint | `make lint` | ✅ (eslint + tsc limpios) |

## Comprobaciones del ciclo 1, repetidas sobre el código corregido (en vivo)

| # | Verificación | Evidencia | Resultado |
|---|---|---|---|
| R1 | `POST /admin/medios/buscar-nuevos` sin sesión | `401` `{"error":{"code":"unauthenticated",…}}` | ✅ FR-021 |
| R2 | 403 con cuenta sin `medios` (`qa-sin-medios-oct18@ejemplo.com`, contraseña restablecida por mí a efectos de prueba) | `403` `{"code":"forbidden","message":"No tienes permiso para acceder a este módulo"}` | ✅ FR-021/SC-009 |
| R3 | El 403 (y cada operación) queda auditado | `admin_actions`: `media.sync|denied=5` (incluye el mío de hoy; antes 4), `media.sync|success=3` (incluye mi búsqueda en vivo), `media.hide/show/merge/description.update/mark.resolve` con `success` **y** `denied` | ✅ FR-022/SC-010 |
| R4 | Catálogo de permisos intacto | `count(permissions)=10`, `medios` presente | ✅ SC-009 |
| R5 | **Idempotencia real (FR-012/SC-004)** — 2ª búsqueda en vivo | `{"importedCount":0,"joinedCount":0,"markedCount":0,"hiddenCount":0,…,"youtube":"failed","podcast":"ok"}` en 0,35 s | ✅ |
| R6 | **Fail-safe feed caído en vivo (FR-016/SC-007)**: en la misma corrida `youtube:"failed"` (404 desde esta IP) | `hiddenCount:0, markedCount:0` — nada se oculta ni marca por un feed caído | ✅ |
| R7 | Predicado público: sección `GET /api/v1/medios` | 9 visibles / 9 totales de los reales; **0 claves proscritas** (`visibility`/`reviewMark`/`mergedIntoId`…) en ninguna ficha; claves: `audioUrl, coverUrl, description, id, publishedAt, title`; **0 fichas sin enlace**; **0 títulos repetidos**; descripciones ≤ ~300 con «…» | ✅ FR-002/FR-009/FR-020/SC-002/SC-003 |
| R8 | **Tope 5 de la portada (FR-027/SC-015)** | `GET /api/v1/portada/medios?lang=es` → exactamente **5** con 9 visibles, en orden de fecha | ✅ |

## Estado de las correcciones comprobadas (lo nuevo del ciclo 2)

| Corrección | Cómo la comprobé | Estado |
|---|---|---|
| **H1 — La búsqueda ya no se corta a los 5 s y el equipo ve el resumen (FR-018)** | `frontend/src/api/medios.ts`: `SYNC_TIMEOUT_MS = 120_000` y `timeoutMs` pasado explícito; fijado por pruebas (`api/medios.test.ts`, `mediosHooks.test.tsx`, nuevas). En vivo mi búsqueda devolvió el **resumen completo** con todos los campos (R5) | ✅ corregida |
| **H2 — La carátula no se pisa al unir y una sola por ficha (la del video)** | `TestBuscarNuevosNoPisaLaCaratula` (dos sub-casos: unir no pisa ni borra el archivo del video; **las dos carátulas de la misma búsqueda → gana la del video**, se descarga solo una) y `TestBuscarNuevosCaratulas`; estructuralmente, `media_items.cover_file` es **una sola columna** por ficha. Verificado de código y suite en verde | ✅ corregida |
| **H3 — Alta atómica: sin fichas huérfanas en carrera** | `TestBuscarNuevosAltaAtomica` (ficha+fuente en una transacción: el que pierde la carrera **no escribe nada**; también el caso del par dudoso), `TestIntegrationInsertMediaItemWithSourceEsAtomico` (PostgreSQL real). Mi verificación en vivo R7: **0 fichas visibles sin enlaces** | ✅ corregida |
| **A4 — El `offset` enorme en la ruta pública ya no da 500** | En vivo: `GET /api/v1/medios?limit=20&offset=999999999` → **`200`** con `{"items":[],"total":9,…}` (antes 500); fijado por `paginate.TestFromValuesCapsOffset` y `TestIntegrationPaginacionConOffsetEnElTope` | ✅ corregida |
| **B3 — El panel avisa si hay más de 100 fichas y la unión encuentra a la pareja fuera de la primera página** | `MediosPage.test.tsx`: «con más de 100 fichas la lista avisa de que hay más (**no trunca en silencio**)» — texto literal «Esta lista muestra 100 de 101 episodios. »— y «la pareja no está en las primeras 100 fichas» → consulta con `offset=100`; en verde en Vitest (y en la corrida completa 946/946) | ✅ corregida |

Junto con ellas, aplicadas y con prueba: pareja recíproca única (`TestBuscarNuevosParejaReciprocaUnica` + 3 pruebas de integración de disciplina/mark/resolve **atómicos en transacción**, M1/M2), el sync ya **no revierte una edición durante la búsqueda** (`TestBuscarNuevosNoRevierteUnaEdicionDuranteLaBusqueda`, M3), PATCH de solo-visibilidad ya se audita con su código propio (B2, `TestValidarMarcaYVisibilidad`), límite de descripción contado en **runas** con emoji (B1, `validation.test`/`DescripcionForm.test`), y la fábrica saliente exige **puerto 443** (`red_saliente.go` + pruebas).

## Hallazgos nuevos

| # | Severidad | Hallazgo | Evidencia |
|---|---|---|---|
| N1 | **menor (infraestructura de pruebas)** — corregida en esta sesión | El helper e2e `portadaContiene` tenía una **carrera**: el bloque `#medios` se dibuja desde la carga con su esqueleto (sin fichas todavía) y el helper comprobaba el texto al instante sin esperar la consulta → la aserción «true» (el título vuelve a la portada tras «Mostrar») podía perder la carrera de forma reproducible; la de «false» pasaba vacacionalmente. La prueba fallaba dos veces seguidas; el backend/Frontend correctos. Corregí el helper (espera activa con `expect(...).toBeVisible({timeout:5s})`, sin cambiar qué se espera) y el diagnóstico quedó en `frontend/e2e/qa-ciclo2-debug.spec.ts`. La batería entera: **25/25**. | Falso rojo en la fila 347 de `medios-panel`; reproducido 2/2; luego verde al esperar el bloque. Mi spec de diagnóstico muestra: portada después del PATCH de «Mostrar» **contiene** la ficha (API y navegador) |
| N2 | menor (entorno/sesión anterior, se repite) | Residuo de esta revalidación: restablecí la contraseña de la cuenta de prueba `qa-sin-medios-oct18@ejemplo.com` (fines de prueba); mis búsquedas en vivo añadieron filas honestas a `admin_actions` (`media.sync` success/denied) y mi intento de 403. No alteran el comportamiento. | R1–R5 |

**No encontré ninguna regresión de producto**: ni el tope de 5, ni el reordenado por fecha, ni el predicado público, ni la idempotencia, ni el fail-safe se vieron alterados por los 11 commits de corrección; solo el helper tenía el problema (N1).

## Pruebas nuevas añadidas

- `frontend/e2e/qa-ciclo2-debug.spec.ts` — spec de diagnóstico del QA: siembra, oculta/muestra por la interfaz y captura la respuesta real de `/api/v1/portada/medios` en cada paso (incluye what the portada browser request returns). Es prueba, no producción, y documente dónde se perdía.
- Ajuste del helper `portadaContiene` en `frontend/e2e/medios-panel.spec.ts`: espera a que el bloque cargue antes de comparar el texto (corrección de la prueba; ninguna aserción se debilitó). **No hice commit** de nada.

## Otro nivel

No repetí `make test` completo (toda la agenda de backend de otros módulos) ni `make security`: el alcance de esta revalidación es F9; en el ciclo 1 todo eso estaba en verde y los toques de corrección se circunscriben a `medios`, `paginate` y `frontend/src/{api,features/medios}` (diff verificado: solo archivos de esos paquetes + pruebas). La integración completa `go test -tags=integration ./...` conserva su inestabilidad declarada preexistente sobre la base compartida; corrí la de `internal/medios` en verde.

---
*Preparado por `qa-tester`, 2026-10-18/19, F9 fase 7, revalidación del ciclo 1. Sin commit.*
