# Revisión QA — F6 Ministerios (2026-10-13)

**Rol**: `qa-tester` · **Rama**: `006-ministerios` · **Alcance**: `git diff main...HEAD` (~40 commits desde `d114cac`, T601–T637) contra la **spec v2 aprobada** (6 US, 29 FR, 17 SC).

**Veredicto: APROBADO** — todo lo automatizable pasa, incluida la verificación independiente con datos reales que sembré y comprobé por mi cuenta (privacidad, permisos, validaciones, orden, tope de 5, redirección silenciosa); quedan como pendientes manuales el protocolo de usabilidad SC-012 (con el cronómetro de 30 s de SC-001) y el lector de pantalla real, y como hallazgo el flaky preexistente de F3/F4, que no es de F6.

## Entorno y ejecución de suites (hecho por mí, no repetido del estado.md)

| Suite | Comando | Resultado |
|---|---|---|
| Backend unitarias | `go test ./...` | ✅ todos los paquetes ok (ministerios incluido) |
| Backend integración (PostgreSQL real) | `go test -tags=integration ./internal/ministerios/...` | ✅ ok (17 `TestIntegration*`, incl. `UpDownUp` de 000010/000011, catálogo de permisos sin cambios, checks a nivel de base) |
| Cobertura del dominio | `go test -tags=integration -coverprofile` sobre `internal/ministerios` | cobertura del paquete **82,1 %** por sentencias; **`service*.go` ≈ 96 %** (media por función 95,7 %), coherente con lo reportado por desarrollo |
| Frontend unitarias | `npm test -- --run` | ✅ **649/649** |
| E2e de F6 (tanda 4) — Docker Playwright 1.63.0-jammy, `--network host`, `--workers=1` | `npx playwright test ministerios- portada-ministerios` | ✅ **19/19** (26,5 s) |
| Sondeo del flaky de F3/F4 | `npx playwright test qa-bugs-004` | ❌ Bug 1 falló en mi corrida (preexistente de F4, ver abajo) |
| API en vivo (`curl` sin seed de e2e) | sección pública, bloque de portada, panel, permisos, auditoría | ✅ recorrido completo (ver matriz) |
| DOM en vivo (script Playwright ad-hoc, fuera del repo) | portada, `/ministerios`, estado vacío | ✅ (ver matriz; el script no se agregó a la suite) |

**Estado del stack**: imágenes de `backend`/`frontend` reconstruidas el 2026-10-13 18:17 (después del último commit de código de F6, `33cb3a5`), así que lo verificado en vivo es el binario de la rama.

**Nota de sesión**: la validación corrió con `revisor-codigo` y `seguridad` en paralelo, con actividad concurrente sobre el mismo stack (logins de la misma cuenta admin y siembras de otros specs e2e). Re-autenticé cuando la sesión me quedó invalidada y releí el estado de la base tras cada aserción; nada de lo que informo aquí depende de lecturas posiblemente interferidas.

## Verificación independiente con datos reales

Sembré yo los casos: **M1** (con encargado «Roberto Teléfono» / «Robert Phone» y celular `+506 8888-8888`, interruptor **omitido** en la creación, publicado, con enlace de WhatsApp, horario y lugar en ambos idiomas), **M2** (encargado visible con solo correo, sin enlace), **M3** (borrador con encargado visible), **https de dominio que NO es WhatsApp**, **Enlace compartido** (el mismo enlace de M1) y varios casos de frontiera, todo por la API con cuenta de administración real:

| # | Verificación | Prueba concreta ejecutada por mí | Resultado |
|---|---|---|---|
| 1 | **Interruptor apagado por defecto** (US2.1, FR-025, puerta 2-A) | `POST` **sin** la clave `leaderVisible` con encargado y contacto: el DTO del panel devolvió `"leaderVisible":false`; la fila nació con `false` | ✅ |
| 2 | **Privacidad del encargado: 0 apariciones en toda respuesta pública** (FR-026/SC-014, con datos reales; bilingüe) | `GET /api/v1/ministerios?lang=es` y `?lang=en` sin sesión: para M1 **no existe la clave `leader`** (omitida); 0 coincidencias por `grep` de nombre (es/en) y celular en ambos endpoints (sección y portada) en ambos idiomas (**8/8 grep = 0**) | ✅ |
| 3 | Idem en el **HTML renderizado y en el tráfico de red** | Playwright: portada y `/ministerios` renderizadas; `outerHTML` de ambos documentos **sin** «Roberto Teléfono», «Robert Phone» ni «8888-8888»; de las respuestas API capturadas por la SPA (5), **0 contienen datos del encargado oculto** | ✅ |
| 4 | **La prueba no es vacua**: con el interruptor **encendido** el dato SÍ aparece | `PATCH leaderVisible=true` → la sección pública muestra `"leader":{"name":"Roberto Teléfono","phone":"+506 8888-8888"}`; en HTML aparece el correo de M2 visible | ✅ Y el bloque de la portada nunca transporta encargado ni siquiera con la ficha publicada (por diseño del DTO del bloque) |
| 5 | **Ciclo del interruptor sin perder datos** (US2.10–11, FR-025) | apagar → encender → apagar por `PATCH`: la ficha aparece/desaparece del público en cada lectura inmediata; los datos del panel (nombre es/en + celular) se conservan íntegros con `false` | ✅ |
| 6 | **Enlace opcional** (FR-003, Q2) | M2 sin enlace: en la respuesta pública la clave `whatsappUrl` **se omite** (no va vacía); en el DOM la ficha de M2 no tiene botón (los 2 botones existentes corresponden a las únicas 2 fichas con enlace: M1 y «Enlace compartido») | ✅ |
| 7 | **Esquemas no seguros rechazados** (FR-015, Q3) | `POST` con `javascript:`, `data:`, `vbscript:`, `http://` y sin esquema → **5×400 `invalid`**, cada uno con `details.whatsappUrl` («empiece por https://») y **sin guardar nada** (0 filas nuevas) | ✅ |
| 8 | **`https` de dominio NO de WhatsApp aceptado** (Q3) | `POST` con `https://t.me/ministerio` → **201** (la primera corrida dio 400 por un error MÍO: se me coló un tab en medio de la URL; repetida limpia: aceptado) | ✅ |
| 9 | **Contacto: celular O correo, nunca ambos** (Q8) | `POST` con `leaderPhone`+`leaderEmail` → 400 (además el check `ministries_check` lo impone a nivel de base, verificado con `\d ministries`); correo y celular mal formados → 400 | ✅ |
| 10 | **Encargado solo con nombre, sin línea de contacto** (Q8) | caso cubierto por `TestIntegrationListPublishedMinistriesShowsThePublishedLeader` (ministry «Niños», solo nombre, se exige que NO llegue ningún contacto) y `TestGetMinisteriosLeaderWithNameOnly` | ✅ |
| 11 | **Obligatorios y límites** (FR-014) | `nameEs:""`, `descriptionEs:""`, 2001 chars → 400 con detalle por campo; descripción de **exactamente 2000** chars → 201 | ✅ |
| 12 | **Duplicado exacto / enlace compartido** | repetir M1 clave a clave → **409 `conflict`** con mensaje claro; mismo enlace con datos distintos → **201 con `whatsappUrlShared:true`** (aviso, se permite) | ✅ |
| 13 | **Español base por campo** (FR-007/FR-008) | `scheduleEn` sin `scheduleEs` → 400; `nameEn` sin `nameEs` → 400; y en público `lang=en`: M1 muestra `schedule`/`place` en inglés y su descripción (sin traducción) **en español**, nunca vacía | ✅ |
| 14 | **Borrador jamás expuesto** (US3.6, FR-012, SC-002) | M3 (borrador, con encargado) ausente de las 2 respuestas públicas en 2 idiomas (4/4), del HTML de portada y sección, y de ningún intento forzado | ✅ |
| 15 | **Orden manual `sort_order, id`** (FR-004/FR-016) | M2 reordenada de `20` a `1` → la sección pública la adelanta a M1 (10); el índice `ministries_sort_order_id_idx` existe; desempate por `id` cubierto por `assertOrderedBySortOrderThenID` | ✅ |
| 16 | **Cambios visibles de inmediato** (FR-013/SC-003) | M1 despublicado → desaparece de sección y portada en la primera lectura posterior; republished → reaparece (2×) | ✅ |
| 17 | **Bloque de portada: tope 5 resuelto en servidor** (FR-027/SC-016) | 4 publicados + 3 sembrados más (`Tope2–Tope4`): el endpoint devolvió **exactamente los 5 primeros por `sort_order`**, nunca más; `?limit` ignorado (`TestGetPortadaMinisteriosHandlerIgnoresLimit`); `Cache-Control: no-store` | ✅ |
| 18 | **Posición del bloque en la portada** (Q7/FR-027) | DOM por posición: «Lo que viene» (956) → tarjetas de eventos (1157–1246) → **«Ministerios» (1854) + 5 tarjetas (2055)** → «Quiénes somos» (2652): exactamente **entre el bloque de eventos y «quiénes somos»** | ✅ |
| 19 | **Bloque sin publicados: oculto por completo** (FR-028) | con 0 publicados: la portada no muestra el bloque (el título no está y «Quiénes somos» queda coherente); `GET /api/v1/portada/ministerios` → **`200 {"lang":"es"}` sin `items`**, `no-store` | ✅ |
| 20 | **`/ministerios` sin publicados: la API NO redirige** (contrato estable) | `curl` sin navegador → **200** sin `items` (no 30x) | ✅ |
| 21 | **Redirección silenciosa sin destello** (SC-010, puerta 2-B) | visita directa `goto('/ministerios')` con `MutationObserver` + polling de 4 s: **0 apariciones** del contenido de la sección (ni H1 ni grilla ni aviso), URL final `/`, 0 mensajes de «sin ministerios», portada completa; ídem con `history.pushState` (SPA); el botón **Atrás no vuelve a la URL muerta** | ✅ |
| 22 | **Nav sin enlaces rotos** (FR-005) | **0** `<a>` hacia `/ministerios` (por href o texto) en toda la portada renderizada — barra, menú y mapa del pie ocultos | ✅ |
| 23 | **Permisos en el servidor** (SC-005, FR-011) | rol `sin-f6-qa` con **solo `eventos`** + cuenta de prueba real: **9 llamadas forzadas → 9×403** (GET lista, GET uno, POST crear, PATCH editar/publicar/leaderVisible/reordenar, POST imagen multipart, DELETE) con mensaje claro («No tienes permiso…») y **el ministerio sigue intacto** después | ✅ |
| 24 | **F6 no crea permisos nuevos** (FR-011) | la sesión del init responde con el catálogo **exacto de 10 permisos**; `TestIntegrationPermissionsCatalogUnchanged` en la base | ✅ |
| 25 | **Auditoría: quién/qué/sobre qué/cuándo, incl. intentos** (FR-017/SC-011) | en `admin_actions` (con actor, `result` y `created_at`): durante esta sesión aparecieron los **8 códigos** (`create`, `update`, `delete`, `publish`, `unpublish`, `leader.show`, `leader.hide`, `image.upload`); los **9 intentos forzados quedaron `result=denied` con el actor correcto** (la cuenta sin permiso, no la de admin) y las validaciones fallidas como `result=failure` | ✅ |
| 26 | **El alta registra SOLO `ministry.create`** (T615) | crear un ministerio con `leaderVisible:true` y releer el intervalo exacto en `admin_actions`: **una sola fila** (`ministry.create`), sin `leader.show` anidado | ✅ |
| 27 | **Imagen opcional + alt** (FR-020) | multipart de un PNG de 1 px con `imageAltEs` → 201 (`img_<uuid>.png`); el público lo expone como  `"image":{"url","alt"}`. El alt no se exige en la **subida** sino al **guardar** (semántica de fusión A6-3: el `PATCH` sin `imageAltEs` conserva el alt guardado → 200 correcto) | ✅ |
| 28 | **Horario/lugar: texto libre, 0 fechas derivadas** (FR-024/SC-015) | el código de las fichas no usa `new Date`/`<time>` y las pruebas unitarias `MinistryCard`/`MinisteriosPage`/`MinistryItem` lo asientan explícitamente (aserciones «sin `<time>`»); en el DOM el horario es texto tal cual | ✅ |

## Matriz US × escenario → cumple → evidencia

Referencias: los números `V-n` remiten a la tabla de verificación independiente; `e2e` a la tanda 4 que ejecuté (número entre paréntesis = orden del listado de la corrida); `go-int` a las pruebas de integración de Go; `vitest` a las unitarias de frontend.

### US1 — Conocer y unirse (P1) → **cumple**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 datos completos cuando existen y se publican | ✅ | V-1/4/6/10; e2e (11–15); `go-int` público |
| 2 unirse en un paso al grupo | ✅ | V-4; e2e (12) (botón con el href tal cual) |
| 3 sin enlace/encargado: ficha coherente (SC-017) | ✅ | V-3/6 (M2 sin botón ni líneas vacías); e2e (15/17) |
| 4 orden estable sin esfuerzo (Q5) | ✅ | V-15; `go-int` `OrdersAndLimits` |
| 5 horario/lugar como texto (Q6) | ✅ | V-13/28; vitest «sin `<time>`» |
| 6 identifica y sabe cómo participar (estructura) | ✅ | V-3/4/10; el «90 % de personas» es **SC-012, manual** |
| 7 con y sin imagen, coherente | ✅ | V-27; e2e (15–19); alt accesible |
| 8 varios con el mismo encargado, ficha no compartida | ✅ | encargado por columnas del ministerio (sin tabla de personas); V-4; `go-int` por nombre de los 3 perfiles |
| 9 nada se confunde con F4/F5 | ✅ | solo `ministries` en todas las vías; sin `event_dates`/recurrencia (V-28); modelo separado (`data-model.md`) |

### US2 — Gestión del panel (P1) → **cumple**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 crear con válidos e interruptor nace apagado | ✅ | V-1; e2e panel (5); `go-int` `TestIntegrationHiddenByDefault` |
| 2 sin permiso: mensaje claro, nada cambia | ✅ | V-23; e2e panel (9) cruza los 3 permisos de F4 |
| 3 datos inválidos → indicar qué corregir, sin medias | ✅ | V-9/11/13 + 19 casos de `TestValidateMinistryReportsEachField` |
| 4 edición del publicado visible al instante | ✅ | V-16; e2e panel (4) |
| 5 eliminar con confirmación + auditoría | ✅ | V-25/26 (204 y fila fuera + auditoría); e2e panel (8) con diálogo |
| 6 versión en inglés conservada (por campo) | ✅ | V-5 (leader_name_{es,en} conservados tras el ciclo); `go-int` `PersistEveryColumn` |
| 7 inglés vacío permitido (Decisión 8) | ✅ | V-13; `go-int` valor vacío → NULL sin error |
| 8 cruce con otros módulos (independencia) | ✅ | V-23 (rol con `eventos` no accede); viceversa en e2e panel (9) |
| 9 reordenar conserva el orden visible | ✅ | V-15 |
| 10 apagar el interruptor deja la ficha sin publicar | ✅ | V-5 |
| 11 volver a encender restaura la publicación | ✅ | V-5 |

### US3 — Control de la publicación (P2) → **cumple**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 borrador oculto, resto coherente | ✅ | V-14 |
| 2 sin publicados: sección, bloque y nav ocultos + `/ministerios` redirige | ✅ | V-19/20/21/22; e2e (15/18) |
| 3 publicar → aparece (sección y bloque) | ✅ | V-16; e2e panel (4) |
| 4 devolver a borrador → desaparece | ✅ | V-16 |
| 5 cambios visibles al guardar | ✅ | V-16 |
| 6 borrador nunca expuesto por cualquier vía | ✅ | V-14 |
| 7 panel muestra el estado | ✅ | e2e panel (7) |
| 8 visita directa: redirección limpia, sin parpadeo, silenciosa | ✅ | V-21; e2e (15) con aserciones de «destello» |

### US4 — Ver en inglés (P2) → **cumple**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 selector: interfaz y contenidos traducidos, persistente | ✅ | e2e (13) (selector por radios); persistencia entre visitas del selector de F10 (tanda 3) |
| 2 fallback por campo sin vacíos | ✅ | V-13; e2e (13) |
| 3 sin preferencia → español | ✅ | selector de F10 reutilizado (tanda 3) |
| 4 volver a español sin pérdida | ✅ | e2e (13) |
| 5 enlace no cambia con el idioma | ✅ | mismo `whatsappUrl` en es/en (V-2, dos `curl`) |

### US5 — Responsividad y accesibilidad (P2) → **cumple lo automatizable; manual: lector real**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 sin scroll horizontal a 320 px (y más) | ✅ | e2e (14/19): 320/768/1280 px |
| 2 navegación simple, hamburguesa etiquetada | ✅ | e2e (15/19) + sistema F10 reutilizado |
| 3 teclado con foco visible y orden correcto | ✅ | e2e ministerios-accesibilidad (1–3, 19): recorrido completo, CTA del bloque, sin trampas de foco, claro/oscuro |
| 4 contraste y texto al 200 % sin pérdida | ✅ | tokens y medidas de F10 automatizadas (tanda 3 de F4); e2e F6 en ambos temas |
| 5 mismas funciones en cualquier dispositivo | ✅ | e2e 3 anchos |

**Pendiente manual**: lector de pantalla real (NVDA/VoiceOver) sobre `/ministerios` y el bloque. La estructura semántica, jerarquía y etiquetas ya están automatizadas y en verde. (El «sin anuncio accesible al redirigir» es decisión del humano — Out of Scope de la spec.)

### US6 — Bloque de la portada (P2) → **cumple**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 entre «Lo que viene» y «Quiénes somos», tarjetas con imagen | ✅ | V-18; e2e (16) posición por data-testid |
| 2 >5 → los 5 primeros por orden manual, sin paginación | ✅ | V-17; e2e (16) con 8 sembrados |
| 3 <5 → exactos, sin marcadores vacíos | ✅ | V-3 (2 items); e2e (17) |
| 4 0 publicados → bloque oculto por completo | ✅ | V-19; e2e (18) |
| 5 botón abre `/ministerios` en un paso | ✅ | e2e (16/19) click y teclado |
| 6 sin borradores en el bloque | ✅ | V-14; e2e (16) |
| 7 bilingüismo + responsividad sin alterar el resto de la portada | ✅ | V-2 (es/en); e2e (19) 320 px; la posición exacta demuestra que no altera lo demás (V-18) |

## SC — resumen

| SC | Estado | Evidencia principal |
|---|---|---|
| SC-001 100 % sin sesión + <30 s para unirse | ✅ la parte sin sesión y estructura / ⚠️ crono manual | todas las verificaciones V sin cookies; el crono queda en el protocolo SC-012 |
| SC-002 0 borradores | ✅ | V-14 (todas las vías) |
| SC-003 visible en la primera carga posterior | ✅ | V-16, `no-store`, e2e panel (4) |
| SC-004 crear+publicar <2 min | ✅ | e2e panel (4) (assert temporal <120 s) |
| SC-005 permiso en el 100 % de las operaciones | ✅ | V-23/24 |
| SC-006 inglés completo, 0 vacíos | ✅ | V-13, e2e (13) |
| SC-007 0 scroll a 320 | ✅ | e2e (14/19) |
| SC-008 0 hallazgos críticos AA | ✅ automatizable / ⚠️ lector real | e2e accesibilidad claro/oscuro + manual |
| SC-009 0 enlaces no seguros · 0 botones vacíos | ✅ | V-6/7/8; e2e panel (7) |
| SC-010 0 secciones vacías + redirección limpia | ✅ | V-19/20/21/22 |
| SC-011 auditoría completa, no editable | ✅ | V-25/26; e2e panel (10) etiquetas legibles; check de acción permitida en la base |
| SC-012 usabilidad ≥90 % | ⚠️ **PENDIENTE manual** | protocolo en `pruebas-usabilidad-SC-012.md` |
| SC-013 0 grupos de F5 / eventos F4 en estas vistas | ✅ | modelo separado; V-28 (sin recurrencia ni `<time>`) |
| SC-014 0 apariciones del encargado oculto | ✅ | V-2/3/4 (JSON 8/8 greps en 4 endpoints-idiomas, HTML, red) |
| SC-015 horario/lugar tal cual, 0 vacíos, 0 calendario | ✅ | V-28 |
| SC-016 bloque exacto: 5 o los que haya | ✅ | V-17/19 |
| SC-017 sin enlace + sin encargado coherente | ✅ | V-3; e2e (15/17) |

## Hallazgos

### Bloqueantes

**Ninguno.**

### Observaciones / mejoras (no correctivas ahora)

1. **Cobertura del paquete `ministerios` 82,1 %** (por sentencias; `service*.go` ≈ 96 %). Quedan ramas sin cubrir en errores de IO/escritura de imágenes y algunos paths de repositorio. No rebaja el estándar de F4 (service 86,8 %) ni del plan (≥ 80 % en service); es una mejora opcional si se retoca ese código.
2. **Residuo local de pruebas** en la base de desarrollo: el rol `sin-f6-qa` y la cuenta desactivada `otros-qa-…@ejemplo.com` (F2 no elimina cuentas y el rol quedó en uso). Los 11 ministerios de prueba y las 2 imágenes QA **sí** se eliminaron por la API. Un `make down -v` limpia todo; no afecta a otros entornos.
3. **Actividad concurrente en la sesión** (revisor/seguridad/e2e en paralelo): si se bloquea de nuevo contra esta base, tener en cuenta entradas de auditoría con actores de e2e que no pertenecen a ninguna suite en curso.

### Flaky preexistente (F3/F4, `frontend/e2e/qa-bugs-004.spec.ts` «Bug 1»)

En mi corrida falló «Bug 1» (scroll a `#whatsapp` desde `/eventos`; el `scroll-behavior: smooth` se cancela por un cambio de layout) y los otros 2 tests no llegaron a correr. Coincide con la tasa de ~25 % reportada. **No es de F6**: F6 no cambió ese spec (`git diff main...HEAD -- frontend/e2e/qa-bugs-004.spec.ts` = vacío) ni el comportamiento que prueba. Se reporta como hallazgo **preexistente, no bloqueante de F6**.

## Pendiente manual (protocolo y responsable)

- **SC-012** — protocolo completo (tareas, guion, registro y criterio de aprobación, incluido el **cronómetro de 30 s de SC-001**) en **`specs/006-ministerios/pruebas-usabilidad-SC-012.md`**. Responsable: el equipo (la persona que designe el humano). Mínimo 6 personas de distinta edad y experiencia. Debe ejecutarse **antes del cierre de la funcionalidad** (precedente SC-010 de F4).
- **Lector de pantalla real** (NVDA/VoiceOver) sobre `/ministerios` y el bloque de la portada, y una pasada en pantalla táctil real. Automatizable ya en verde; verificado por personas del protocolo manual.

## Recorrido del quickstart

En vivo verifiqué §0 (stack de la rama), §1 (200 sin sesión + `no-store`, `items` omitida, sin publicados `{"lang":"es"}`, API no redirige), §2 (roles y cuentas con permisos), §3 (alta y reglas de campos), §5 (duplicado 409 / enlace compartido con aviso), §7 (privacidad con 0 apariciones) y §9 (fallback en/en → es). El resto de las secciones está cubierto por las suites que ejecuté (tanda 4 e2e, integración, unitarias) y por el bloque `portada-ministerios` del e2e.

## Confirmación

**No toqué ningún código de producción ni pruebas existentes** — ni la spec, plan, ux, estado, roadmap ni plantillas. El único archivo añadido es este informe y el protocolo de usabilidad, con un commit `docs(qa)`.

## Veredicto final

**APROBADO** para la puerta de validación: emitir PR con el `documentador`, con los dos pendientes manuales (SC-012 y lector real) y el flaky preexistente de F4 asentados como acciones **antes del cierre** de la funcionalidad.

---

## Ciclo 1 — re-validación (2026-10-13, tras las correcciones I1 y M1)

**Alcance**: los commits `1e0d9e5` (contrato: `whatsappUrl` anulable + `minProperties: 1`), `6328cff` (formulario manda `null`, nunca `""`) y `2d75367` (`sortOrder` acotado a int32). Re-verificación **acotada a lo que el cambio toca** + regresión de privacidad, en vivo contra el stack y con las suites.

### Entorno

- El stack local (`backend`/`frontend`) seguía con las imágenes de las 18:16, **anteriores** a las correcciones (20:52–21:06): reconstruí ambas imágenes (`docker compose build backend frontend`) y las levanté de nuevo. Todo lo verificado en vivo corre contra el binario corregido.
- La base quedó limpia tras las verificaciones (0 filas en `ministries`); los 3 ministerios QA que sembré y los del e2e se eliminaron solos.

### Verificación en vivo con datos reales (API + DOM)

Sembré por la API con la cuenta admin: **MIN-A** (creado **con `whatsappUrl: null` explícito**, encargado visible con celular, completo en ambos idiomas, publicado), **MIN-B** (creado **con** enlace, publicado) y **MIN-C** (con enlace, encargado con celular, **interruptor omitido** → nació `"leaderVisible": false`).

| # | Verificación | Evidencia concreta | Resultado |
|---|---|---|---|
| R1 | **Crear sin enlace** (FR-003/SC-009, SC-017) | `POST` con `"whatsappUrl": null` → 201; DTO del panel `whatsappUrl: null`; la sección pública **omite la clave** (nunca vacía); DOM: ficha de MIN-A sin botón de unirse, con encargado «Roberto Teléfono» + `+506 8888-8888`, sin líneas de enlace vacías ni `href=""` | ✅ |
| R2 | **Quitar el enlace del publicado** — el riesgo real (que el `PATCH null` borre algo más) | MIN-B con enlace visible en sección (`whatsappUrl` presente) → `PATCH {"whatsappUrl": null}` → **200**; el DTO del panel conserva **íntegros** `nameEs/En`, `descriptionEs/En`, `leaderNameEs/En`, `leaderEmail`, `leaderVisible`, `scheduleEs`, `placeEs`, `sortOrder: 20`, `publicationState: published` y deja `whatsappUrl: null`; sección es/en muestran todo lo demás (fallback en intacto) y **sin** `whatsappUrl`; DOM: sin botón en la ficha, encargado/correo/horario/lugar intactos | ✅ |
| R3 | **Bloque de portada tras el cambio** | el bloque volvió a servir a MIN-B (`name/schedule/place`); su DTO **nunca transporta** el enlace (FR-027, por diseño) y el HTML de la portada no contiene el enlace quitado (`QaRvTest999` = 0) ni el encargado oculto | ✅ |
| R4 | **`PATCH {}` sigue siendo 400** (`minProperties: 1`) | `{"error":{"code":"invalid","message":"No hay cambios que guardar"}}`, HTTP 400 | ✅ |
| R5 | **`sortOrder` fuera de int32 → 400 por campo, sin truncar** | POST y PATCH con `4294967296` y `-2147483649` → **4×400** con `details.sortOrder` («el orden admite valores entre -2147483648 y 2147483647»); extremos `2147483647` y `-2147483648` → **200** y la ficha quedó exacta en la columna `INTEGER` (verificado en la base: 10 y 20 tras restaurar) | ✅ |
| R6 | **Regresión de privacidad del encargado oculto** (SC-014) | MIN-C: 12/12 greps = **0** apariciones de «Sigfrido Oculto» / «Sigfrid Hidden» / «7777-7777» en los 4 endpoints públicos (sección y portada, es y en); clave `leader` **omitida** en es/en; HTML de `/ministerios` y de la portada con 0 apariciones (script Playwright fuera del repo). El interruptor nace apagado (`"leaderVisible": false` con la clave omitida en el alta) y **el dato sí existe en el panel** (no se descartó en el guardado) | ✅ |
| R7 | El renderizado del enlace **no se rompió** con el cambio | la ficha de MIN-C (sí tiene enlace) muestra su botón con el `href` exacto; la prueba `TestContractPatchAllowsNullWhatsappURL` y las 3 de Vitest del formulario quedan en verde | ✅ |

### Suites ejecutadas por mí en esta re-validación

| Suite | Comando | Resultado |
|---|---|---|
| Frontend unitarias | `npm test -- --run` | ✅ **652/652** (94 archivos; 649 anteriores + 3 nuevas de `6328cff`) |
| Backend unitarias | `go test ./...` | ✅ 20 paquetes ok |
| Backend integración (PostgreSQL real, `count=1`) | `DATABASE_URL_TEST=… go test -tags=integration -count=1 -v ./internal/ministerios/...` | ✅ **125 PASS / 0 FAIL** |
| Pruebas nuevas de contrato/dominio (explícitas) | `go test -run TestContractPatchAllowsNullWhatsappURL\|TestContractSortOrderIsBounded\|TestSortOrderOutOfRangeIsRejected` | ✅ todas PASS (con los subtest `orden_por_encima/debajo_del_rango_de_la_columna`) |
| E2e de F6 — Docker Playwright 1.63.0-jammy, `--network host`, `--workers=1` | `npx playwright test --config e2e/playwright.config.ts ministerios- portada-ministerios` | ✅ **19/19** (27,0 s) |
| DOM en vivo (script ad-hoc fuera del repo) | portada + `/ministerios` | ✅ (tabla R1–R7) |

### Notas de la corrida

1. **Las pruebas de integración se omiten en silencio** si `DATABASE_URL_TEST` no está definida (`t.Skipf` en `internal/platform/testutil/db.go`): una corrida sin la variable da `ok` en <1 s sin ejecutar nada. Con la variable puesta contra el PostgreSQL local del compose, las 125 corren y pasan. (Observación, no defecto: el patrón es el mismo de F2/F3/F4.)
2. La primera tanda e2e falló 4 pruebas por `ENOENT /resources/simiente.jpeg`: **error del método** (monté solo `frontend/` en el contenedor y el fixture vive en `<repo>/resources/`); montando el repo raíz, 19/19. No es de F6.
3. Las correcciones no tocan ninguna vía por la que puediera filtrarse el encargado (solo `whatsappUrl`, `minProperties` y `sortOrder`): la regresión R6 lo confirma en vivo.

### Cobertura del caso corregido

La cobertura automatizada del ciclo ya existía y es suficiente; **no añadí pruebas nuevas**:

- Backend: `TestUpdateMinisterioNullExplicitFromJSON` (PATCH `null` borra el enlace y conserva el resto), `TestContractPatchAllowsNullWhatsappURL` (esquema anulable + `minProperties`), `TestUpdateMinisterioHandlerRejections/sin campos` y service (PATCH vacío → 400), `TestSortOrderOutOfRangeIsRejected` + subtest del dominio + HTTP (rango int32, extremos válidos, sin truncar).
- Frontend: `ministryForm.test.tsx` (I1): crear sin enlace manda `null` y no `""`; quitar el enlace al editar manda `null`; editar sin tocar lo conserva.

### Confirmación

**No toqué ningún código de producción ni ninguna prueba existente.** Lo único que añadí en el repo es esta sección del informe. En el stack local reconstruí las imágenes de `backend`/`frontend` (estado del entorno, no del repositorio) y ejecuté las suites; la base quedó sin filas de prueba.

### Veredicto del ciclo 1

**APROBADO** — las actuaciones de corrección I1/M1 hacen lo que dicen, el `PATCH null` no borra nada más, el 400 del `PATCH {}` y del `sortOrder` fuera de int32 se comportan como el contrato, la privacidad del encargado sigue intacta y las suites completas están en verde. Sin bloqueantes. Los pendientes manuales de la validación original (SC-012 y lector real) siguen abiertos para el cierre.
