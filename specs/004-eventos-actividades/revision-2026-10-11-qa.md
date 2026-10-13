# Revisión QA — F4 Eventos y actividades (2026-10-11)

**Rol**: `qa-tester` · **Rama**: `004-eventos-actividades` · **Alcance**: `git diff main...HEAD`
**Veredicto: APROBADO** (con pendientes manuales y 2 hallazgos menores que no bloquean)

## Entorno y ejecución de suites

| Suite | Comando | Resultado |
|---|---|---|
| Backend unitarias | `go test ./...` | ✅ 20 paquetes ok |
| Backend integración (PostgreSQL real, `DATABASE_URL_TEST`) | `go test -count=1 -tags=integration ./...` | ✅ ok en todos los paquetes (agenda: 320 subpruebas, 17 `TestIntegration*` en verde; usuarios 4 s… 80 s; portada con testcontainers) |
| Cobertura de `internal/agenda/service*.go` | `go test -tags=integration -count=1 -cover ./internal/agenda/` | ✅ **86.8 %** (≥ 80 % de la constitución §III) |
| Frontend unitarias | `npm test -- --run` | ✅ 74 archivos / 505 pruebas |
| Lint + typecheck | `npm run lint && npm run typecheck` | ✅ sin errores |
| E2e tanda 2 (F4 agenda, Docker Playwright) | `npx playwright test agenda- --workers=1` | ✅ 5 casos (30 s) |
| E2e tanda 1 (F2: `acceso` ajustado por F4, `auditoria`) | `npx playwright test acceso auditoria --workers=1` | ✅ 2 casos (14 s) |
| E2e tanda 3 (F3/F10: portada, diseño, tema, nav, status) | `npx playwright test portada- diseno- movimiento navegacion- tema- qa- status --workers=1` | ✅ 31 casos (47 s) — incluye contraseña/idioma/tema persistentes entre visitas y contraste/zoom 200 % automatizados de F10 |
| API en vivo (quickstart §1) | `curl /api/v1/proximamente` | ✅ 200 + `Cache-Control: no-store`; `lang=fr` → 400 invalid claro; con agenda vacía responde `{"lang":"es"}` (bloques omitidos) |

## Matriz de cobertura (criterio → prueba → resultado)

### US1 — Descubrir lo que viene (P1)

| Criterio | Prueba | Estado |
|---|---|---|
| AS-1 visitante sin sesión ve título/fecha/lugar/detalle, dos bloques distintos | `e2e/agenda-publica.spec.ts` (H1, dos `h2` en orden, tarjetas con `<time>`, lugar, mapa, categoría, detalle `pre-line`); `service_public_test.go` | ✅ |
| AS-2 próxima fecha del evento + texto de frecuencia | e2e pública (`time[datetime=…T19:00]`, `Más fechas`, `frequency`); `TestGetProximamenteDatesAreOnlyUpcoming` | ✅ |
| AS-3 actividad puntual con fecha y horario (inicio/fin opcionales) | e2e pública (actividad 8:00−11:00 a. m.); `TestGetProximamenteAppliesTheNextRule`, `service_admin_test::TestCreateActividad…` | ✅ |
| AS-4 orden cronológico de lo más próximo a lo más lejano | `TestGetProximamenteOrdersChronologicallyWithStableTieBreak`; `TestIntegrationListUpcomingEventsOrdersByNextDate`; e2e pública | ✅ |
| AS-5 retiro de lo vencido | e2e pública (vencidos con count 0); `next rule` unitaria/integrales; e2e siembra `pastEventTitle` / `todayFinishedTitle` | ✅ |
| AS-6 sabe qué/cuándo/dónde sin buscar en otro sitio | e2e pública (elementos completos + elemento interactivo de mapa `target=_blank rel=noopener`) | ✅ |
| AS-7 evento y actividad misma fecha: ambos en su bloque | `TestGetProximamenteOrdersChronologicallyWithStableTieBreak` (eventos y actividades de la misma fecha, cada uno en su bloque, desempate estable `title,id`) | ✅ |

### US2 — Gestión de eventos (P1)

| Criterio | Prueba | Estado |
|---|---|---|
| AS-1 crear con datos válidos y conservarse | `e2e/agenda-panel.spec.ts` (alta por UI + «Evento creado.»); `TestCreateEventoStoresNormalizedContent` (incl. tildes/ñ/¿¡/emojis intactos, `*En` vacío → NULL) | ✅ |
| AS-2 sin permiso: mensaje claro y nada cambia | `TestAuthzByModuleDeniesAndRecordsWithoutPermission` (middleware); e2e `agenda-panel` (403 por API forzada por los tres permisos); `service_audit_test::TestRecordDeniedPersistsTheDenial` | ✅ |
| AS-3 validaciones sin guardar nada a medias | `TestCreateEventoValidatesEveryField` (19 casos: título/lugar/detalle obligatorios y sus límites, mapa inválido, sin fechas, fecha `2026-13-32`, fecha repetida, `endTime` exige `startTime` y fin≥inicio, hora mal formada, imagen sin `alt`/nombre inválido, categoría inexistente/no-UUID, estado inválido) + auditoría del rechazo `result=failure`; `TestIntegrationEventDatesUniqueRejectsRepeatedDate` | ✅ |
| AS-4 edición de publicado visible de inmediato | e2e panel («Cambios guardados: ya visibles en la sección», visitante recarga y ve el título editado); `TestUpdateEventoReplacesDatesAndPublishesOnSave` | ✅ |
| AS-5 eliminar con confirmación + auditoría | e2e panel (diálogo «se eliminará para siempre…» y «Eliminado.»); `TestDeleteEventoAuditsAndFinds`; `TestIntegrationWriteTxCommitsMutationAndAudit` (mutación+auditoría atómicas; rollback si falla el registro) | ✅ |
| AS-6 versión en inglés conservada | e2e siembra (`titleEn`/`frequencyEn`/`placeEn` visibles en inglés, e2e pública); `TestGetProximamenteFallsBackPerField` | ✅ |
| AS-7 inglés vacío se puede guardar (Decisión 8) | `TestCreateEventoStoresNormalizedContent` (título en inglés de solo espacios → NULL sin error) | ✅ |
| AS-8 «eventos» no da acceso a «actividades» | e2e panel `agenda-panel` (cuenta solo-eventos → `403` en categorías/actividades; IDEM actividades/eventos/categorías mutuas) | ✅ |

### US3 — Gestión de actividades (P1)

| Criterio | Prueba | Estado |
|---|---|---|
| AS-1 alta válida conservada (fecha única) | e2e panel (alta de actividad) ; `TestCreateActividadStoresNormalizedContent`; `TestUpdateActividadKeepsModulesApart` | ✅ |
| AS-2 denegación clara | e2e panel (cruces de los 3 permisos) | ✅ |
| AS-3 validaciones | `TestCreateActividadValidatesEveryField` (misma batería + hora `25:00`, fin anterior, fin sin inicio, mapa `ftp://`, 2001 chars) ; `TestIntegrationActivityScheduleChecks` | ✅ |
| AS-4 edición publicada visible de inmediato | e2e panel + FR-014 en eventos (mismo patrón(panel) probado por U.S.); `TestUpdateActividad…` | ✅ |
| AS-5 eliminar con confirmación + auditoría | e2e panel; `TestDeleteGetAndListActividades` | ✅ |
| AS-6 inglés opcional / ambas versiones | `TestCreateActividadStoresNormalizedContent` (normalización I6) + fallback e2e pública (actividad solo español en inglés → español) | ✅ |
| AS-7 «actividades» no da acceso a «eventos» | e2e panel | ✅ |
| Duplicado exacto intra-módulo 409; evento≠actividad no duplicado | `TestCreateEventoDuplicateIsConflict`, `TestCreateActividadDuplicateIsConflict` (verifica que un EVENTO idéntico no colisiona), `TestIntegrationActivityDuplicateIsConflict`, `TestIntegrationExistsEventDuplicateDetectsSameModuleDuplicate` | ✅ |

### US4 — Publicar/retirar por elemento (P2)

| Criterio | Prueba | Estado |
|---|---|---|
| AS-1 borrador no aparece, resto coherente | e2e pública (`draftEventTitle` count 0 con la sección poblada) + panel («Quedó en borrador…») | ✅ |
| AS-2 nada publicado ⇒ sección oculta, nav y sitemap sin huérfanos, aviso sin redirigir | e2e pública (URL `/eventos` intacta, `h1`, 0 `h2`, 0 enlaces «Eventos y actividades» en barra/menú/sitemap, portada coherente) + `TestGetProximamenteHandlerOmitsEmptyBlocks` (claves vacías omitidas) | ✅ |
| AS-3 publicar → aparece | e2e panel | ✅ |
| AS-4 retirar a borrador → desaparece (sin borrarse) | e2e panel (Retirar con confirmación; «Retirado») | ✅ |
| AS-5 editar lo publicado → visible de inmediato | e2e panel + `no-store` verificado en vivo | ✅ |
| AS-6 NUNCA exponer borrador por ninguna vía (incl. imagen) | e2e pública: `GET /api/v1/media/<img-de-borrador>` → **404**; servicio público solo filtra `published`+próximo; el estado `isPast` conserva `published` (**no se convierte en borrador**) según `nextDate/isPast` del DTO | ✅ |
| AS-7 el equipo ve el estado de cada elemento | Vitest `events.test.tsx`/`activities.test.tsx` (estado **Borrador/Publicado/Pasado** en la tabla, «Pasado» como marca aparte) | ✅ |

### US5 — Inglés (P2)

| Criterio | Prueba | Estado |
|---|---|---|
| AS-1 selector aplica de inmediato, se mantiene al navegar y entre visitas | Vitest + e2e pública (radio Inglés → `html lang=en`, headings, contenidos; F3 la persistencia — heredada, fuera de F4) | ✅ (persistencia entre visitas: heredada de F3) |
| AS-2 fallback es por campo, nunca vacío | `TestGetProximamenteFallsBackPerField` (cada campo traducible) | ✅ |
| AS-3 primera visita sin preferencia → español | comportamiento del selector de F10/F3 (heredado); máxima visita e2e ángulo sin preferencia cargada arranca en es | ✅ (heredado) |
| AS-4 volver a español sin pérdida | e2e pública (radio Spanish → `lang=es`, H1 en español) | ✅ |
| AS-5 fechas escritas y comprensibles por idioma, sin cambiar día/hora | e2e pública («17 de octubre», «7:00 p. m. − 9:00 p. m.» / «October 17», «7:00 PM − 9:00 PM» con `<time datetime>` real); `formatoFechas.ts` con `Intl` sin conversión de zona (16 pruebas) | ✅ |

### US6 — Responsividad y accesibilidad (P2)

| Criterio | Prueba | Estado |
|---|---|---|
| AS-1 / SC-007 320/768/1280 sin scroll horizontal, mismos elementos | e2e pública (`horizontalOverflow ≤ 1 px` en los tres anchos con toda la agenda sembrada) | ✅ |
| AS-2 títulos claros y navegación simple (h1/h2, menú etiquetado) | e2e pública + `agenda-accesibilidad` (menú móvil accesible por nombre) | ✅ |
| AS-3 / SC-008 teclado + lector: recorrido completo, foco visible, sin trampas | `agenda-accesibilidad.spec.ts`: salto al contenido con 1.er `Tab`, travesía completa con indicador de foco por paso, `Shift+Tab` sin trampa, flechas + `Espacio` en radios, claro y oscuro, aviso alcanzable vacío sin redirigir | ✅ (lector de pantalla real: manual, pendiente) |
| AS-4 contraste claro/oscuro y textos hasta 200 % | tokens y medidas de F10 aplicados a la agenda (mismo sistema: `qa-medidas-f10` texto ≥14 px y toque ≥44 px); contraste y texto al 200 % **automatizados** por F10 en `portada-accesibilidad.spec.ts` (✅ tanda 3); el e2e de agenda verifica el tema claro/oscuro de `/eventos` | ✅ (aplicado el sistema F10 verificado) |
| AS-5 mismas funciones en cualquier dispositivo | e2e 3 anchos + menú móvil por teclado | ✅ |

### Edge cases y errores esperados

| Caso | Prueba | Estado |
|---|---|---|
| Elemento del día: se ve durante su día, se retira al terminar (`endTime` pasado) | `TestGetProximamenteAppliesTheNextRule` (hoy en curso vs hoy terminado) + `TestIntegrationListUpcomingEventsFollowsTheNextRule`; e2e `todayFinishedTitle` invisible | ✅ |
| Cargar fechas pasadas está permitido; panel las muestra como pasado | `TestCreateEventoAllowsPastDates`; `nextDate null` + `isPast` + marca «Pasado» (vitest) | ✅ |
| Evento sin próxima fecha: no aparece; aviso en panel | `TestIntegrationListUpcomingEventsFollowsTheNextRule` + vitest («Pasado» con `nextDate: null`) | ✅ |
| Caracteres especiales y emojis íntegros | `TestCreateEventoStoresNormalizedContent` («¿aposentos? 🎉», «ñandú») | ✅ |
| Detalle > 2.000 se rechaza con el límite | tablas de validación (`detalle largo`) | ✅ |
| Imagen inválida / carga fallida ⇒ conservar el elemento sin romper | `TestUploadImageHandlerRejectsUnsupportedTypes` (`.svg/.html` renombrado, firma binaria), `TooLarge` (9 MB), `FailClosed`; elemento con imagen inválida nunca se guarda | ✅ |
| `lang=fr` rechazado con mensaje claro | en vivo: 400 `{"error":{"code":"invalid",...}}` + `validLang` | ✅ |
| Auditoría atómica + best-effort en fallas/denegaciones | `TestIntegrationWriteTxRollsBackWhenAuditFails/RollsBackWhenAuditCodeIsUnknown`, `TestRecordDenied…`, `TestRecordRejectedBestEffort`, `TestUpdateEventoDoesNotSaveWhenAuditFails` | ✅ |

### SC (resumen)

| SC | Cobertura |
|---|---|
| SC-001 visitante 100 % lo ve sin cuenta | e2e pública (contexto sin sesión) ✅ |
| SC-002 0 borradores expuestos | e2e pública +媒体 404 + servicio filtra `published` ✅ |
| SC-003 cambio visible en 1.ª carga | e2e panel + `no-store` en vivo ✅ |
| SC-004 crear+publicar <2 min | aserción temporal en e2e panel (`elapsed < 120`) ✅ |
| SC-005 / SC-011 permiso en el 100 % de operaciones, cruce de los 3 | `AdminChain(permiso)` propio por grupo + middleware test + e2e cruzado completo ✅ |
| SC-006 inglés completo y sin campos vacíos | e2e pública + fallback por campo ✅ |
| SC-007 0 scroll horizontal a 320 px | e2e 3 anchos con `≤ 1 px` ✅ |
| SC-008 0 hallazgos críticos AA | e2e teclado/foco de agenda en claro y oscuro + las medidas AA de F10 (contraste, toque, zoom 200 %) en la tanda 3 ✅; lector de pantalla real con lector físico: manual ⚠️ | ✅ (lectura física pendiente) |
| SC-009 0 vencidos en «lo que viene»; evento sin próxima fecha nunca aparece vacío | next-rule (unitaria+integración) + e2e ✅ |
| SC-010 90 % de personas identifican lo que viene | **protocolo manual (≥ 6 personas, `pruebas-usabilidad-SC-010.md`) — PENDIENTE, fuera de make ci** ⚠️ |
| SC-012 0 secciones vacías / enlaces rotos | e2e pública (aviso, URL estable, 0 enlaces en barra/menú/sitemap; portada sin huérfanos) ✅ |
| SC-013 auditoría completa y solo lectura | unitarias por operación + dos filas al cambiar datos+estado + `denied`/`failure`; e2e panel muestra etiquetas humanas y el elemento; F2 auditoria e2e (tanda 1) verifica «no editables ni borrables» ✅ |

## Hallazgos

### Bloqueantes

Ninguno.

### Menores / mejoras

1. **Inestabilidad de la suite de integración por contención (`internal/platform/session`)** — severidad: menor (infra de pruebas, no de F4).
   Reproducción: `export DATABASE_URL_TEST=... && go test -count=1 -tags=integration ./internal/...` — en la primera corrida completa, `session` falló después de 167 s; ejecutado **aislado** pasa (`exit=0`). Causa probable: los `testcontainers` que giran en paralelo por paquete contra el mismo daemon compiten y el paquete de sesión agota tiempos.
   Recomendación: serializar las pruebas que usan `testcontainers` (`-p 1` o un semáforo de paquete) o documentar que `make ci` las corre secuencialmente (comportamiento actual del Makefile, que va paquete a paquete en dos comandos separados).
2. **Cobertura de accesibilidad desigual**: lector de pantalla real, contraste formal en tema claro/oscuro y ampliación de texto al 200 % no son automatizables en los e2e actuales. SC-008 exige «0 hallazgos críticos»; la parte automatizada (teclado, foco, nombres accesibles, jerarquía, `<time>`, `alt`) está completa y en verde, pero el criterio completo **no** puede darse por cerrado sin las verificaciones manuales anotadas.
3. **Persistencia del idioma entre visitas** (FR-009): el e2e de F4 comprueba el cambio inmediato y al recargar; la persistencia entre visitas estrenada por F10/F3 se reutiliza sin una prueba específica desde `/eventos`. Riesgo bajo (mismo `localStorage`); se puede añadir una aserción a `agenda-publica.spec.ts` si se quiere cierre formal desde F4.

## Pendiente manual (no verificable automatizada)

- **SC-010** — prueba de usabilidad con personas (protocolo en quickstart §10.7; informe `pruebas-usabilidad-SC-010.md`).
- **Lector de pantalla con lector físico** (`NVDA`/`VoiceOver`) sobre `/eventos` — la estructura (h1/h2/`<time>`/`alt`/nombres accesibles) está automatizada y en verde, pero SC-008 pide operar la sección «con lector de pantalla», que no se automatiza.
- Validación visual en dispositivos físicos (táctil real; los e2e usan viewports emulados).

## Recorrido del quickstart

Se verificó en vivo el §0 (stack + migraciones), §1 (`proximamente` sin sesión, `no-store`, bloques omitidos), §8.1 (`lang=fr` → 400) y §13 (suites). Los §2–§7 y §9–§12 están cubiertos por las pruebas de integración/e2e enumeradas arriba (los mismos recorridos con semillas deterministas) y por las 3 tandas e2e ejecutadas después de `make up`, incluida la regeneración de la imagen de `backend`/`frontend` ya reconstruida (estado en §T440 del estado.md).

## Veredicto final

**APROBADO** para la puerta de validación (emitir PR con el `documentador`), con los pendientes manuales de SC-010 y las verificaciones de accesibilidad anotadas explícitamente antes del cierre de la funcionalidad.
