# Analyze — Coherencia cruzada de F3 (Portada e información general)

**Fecha**: 2026-10-09 · **Fase**: `/speckit.analyze` · **Revisor**: `revisor-codigo` (solo lectura; este
reporte es la única escritura).

**Artefactos contrastados**: `spec.md` (APROBADA, 19 FR / 13 SC / US1–US5) · `plan.md` (P3-1…P3-19) ·
`research.md` (R3-1…R3-18) · `data-model.md` (migraciones `000005`/`000006`) ·
`contracts/openapi.yaml` (16 operaciones) · `ux.md` · `quickstart.md` · `tasks.md` (T301–T339) ·
código real de F1/F2 (`backend/migrations/`, `platform/middleware`, `platform/audit`, router de
frontend) · constitución y skills `go-backend`/`postgres-db`/`react-frontend`.

**Método**: matriz FR-001…FR-019 ↔ tareas ↔ endpoints ↔ tablas; rastreo de cada camino público;
contraste bilingüismo, permiso/auditoría, imágenes, accesibilidad y rutas SPA entre los 8 documentos.

---

## Hallazgos

### CRÍTICOS (bloquean implementar)

| ID | Severidad | Dónde | Qué | Recomendación |
|---|---|---|---|---|
| **C1** | CRÍTICO | `contracts/openapi.yaml:189-218` (`getPortadaAdmin`) ↔ `tasks.md` T322/T324/T325/T326 (ninguna lo implementa); consumido por T330:702 (`getPortadaAdmin()`) y por T333–T336 | El contrato declara **16** operaciones, pero las tareas backend solo construyen **15**: falta el agregado `GET /api/v1/admin/portada` (lectura del panel "en ambos idiomas y con `publicationState`", incl. borradores). No hay service (`GetPortadaAdmin`) ni handler que lo sirva: T322 cubre el GET público, T324 los `PUT` de singletons, T325 los `POST/PATCH/DELETE` de listas, T326 solo registra rutas. Sin él, `features/informacion` no tiene de dónde precargar los formularios ni mostrar el estado de cada elemento (US3 esc. 7, FR-011). | Crear una tarea explícita (o ampliar T324/T325) para el service + handler + pruebas de `GET /api/v1/admin/portada` (`PortadaAdmin`, singletons `null` hasta el primer guardado, sobres `{items}`), y añadirla a la matriz de cobertura de FR-011/US3. |
| **C2** | CRÍTICO | `ux.md:227` (§4.6) y `ux.md:439` (D-3) ↔ `data-model.md:117-137`, `contracts/openapi.yaml:1101-1107,1149-1155`, `research.md` R3-5, `tasks.md` T320/T332 | Contradicción de comportamiento en el horario. `ux.md` pide **Día como texto libre** ("Ej.: Domingos") y **Hora como texto libre** ("10:00 a. m. − 12:00 m."), y afirma "el **día** es el texto traducible; la **hora** es igual en ambos idiomas". El modelo/contrato/tareas fijan `day_of_week SMALLINT 0–6` + `start_time "HH:MM"` (regex en BD y contrato), **no traducibles**, con el nombre del día localizado por i18n. El formulario de T336 (`ServiceForm`) no puede alimentar `dayOfWeek: integer` ni `startTime: "HH:MM"` con campos de texto libre, y el día traducible no tendría columna. | Decidir antes de implementar T336/T337/T338: (a) alinear `ux.md` al horario estructurado (select de día 0–6 localizado + campo hora `HH:MM`, con ayuda de formato), o (b) si el cliente exige texto libre, reabrir spec/plan/data-model/contrato/tareas (incluida migración) y re-aprobar. D-3 está marcada ⟲ pero ya fue decidida en el plan (P3-6) y en el contrato inmutable. |
| **C3** | CRÍTICO | `ux.md:103` (§2.2.4) y `ux.md:426` (§8.2) ↔ `spec.md:163` (FR-010) y `spec.md:307-309` (Q11), `research.md:248-260`, `tasks.md` T328 | Contradicción de comportamiento en la memoria del idioma: `ux.md` dice a la vez que la elección persiste "entre visitas (`localStorage`, confirmado por el humano el 2026-10-09)" **y** que "una visita **nueva** siempre entra en español" / "visita nueva → español". FR-010/Q11 (decisión explícita del humano que **ajustó la spec**) exigen que, tras una elección previa, la preferencia **se respete en las visitas siguientes**; solo la primera visita sin preferencia guardada se muestra en español. Implementado literalmente, se incumple FR-010 y SC-006/US4 esc. 5. | Corregir `ux.md` §2.2.4 y §8.2 a: "primera visita sin preferencia guardada → español; las visitas siguientes respetan `localStorage`". Verificar que T337 (e2e) prueba explícitamente **re-visita con preferencia guardada → idioma respetado**, no solo "carga limpia → español". |
| **C4** | CRÍTICO | `contracts/openapi.yaml:145-187` (`GET /api/v1/media/{fileName}`, pública) ↔ `spec.md:169` (FR-013 "por ninguna vía"), `spec.md:211` (SC-002 "cualquier respuesta del sistema"), `plan.md:379` (RG3-8: "un solo camino público hace imposible filtrar un borrador"), `research.md` R3-8/R3-13, `tasks.md` T323 | El endpoint de descarga de imágenes es público, sirve **cualquier** archivo almacenado por nombre y no consulta el estado de publicación. Un logo/imagen de un elemento **retirado a borrador** (o de identidad en borrador) sigue accesible por su URL: el nombre es inmutable y la caché es `immutable`, así que quien conoció la URL mientras estaba publicado sigue viendo datos de un elemento ya en borrador. El borrado solo ocurre al **reemplazar** la imagen (best-effort), no al retirar. Esto contradice FR-013/SC-002 literalmente y falsa la afirmación de RG3-8; además, borrar al retirar **no** es la solución (rompería FR-014 al republicar). | Decidir y documentar antes de T323: o (a) la descarga verifica que el archivo esté referenciado por contenido **publicado** (consulta por nombre, con caché corta/invalidación al publicar/retirar), o (b) se acepta el trade-off de "URL opaca inmutable" y se declara **explícitamente** como excepción acotada de FR-013 en `spec`/`plan`/contrato (aprobación humana), con prueba e2e de que el borrador nunca es **descubrible** (sin URL en el HTML/JSON público). Sea cual sea la opción, añadir la prueba en T323/T318 y en `quickstart` §5/§6. |

### IMPORTANTES (resolver antes o durante la implementación)

| ID | Severidad | Dónde | Qué | Recomendación |
|---|---|---|---|---|
| **I1** | IMPORTANTE | `ux.md:214-218` (§4.4) ↔ `spec.md:161,171` (FR-008/FR-015), `data-model.md:64-77`, `contracts/openapi.yaml:949-1004` | `ux.md` §4.4 define la identidad con **"Nombre oficial (no traducible)"** (el modelo/contrato tienen `name_en`; FR-008 hace traducibles los textos), **"Lema (es obligatorio)"** (es `tagline_es NULL` y el contrato no lo exige) y **"texto alternativo obligatorio (es/en)"** (solo `*_alt_es` es obligatorio con imagen; `en` es opcional). | Alinear `ux.md` §4.4 con el contrato: nombre es/en (es obligatorio), lema opcional, `alt` obligatorio solo en español con `en` opcional (como ya dice `ux.md` §8.1). |
| **I2** | IMPORTANTE | `ux.md:320,427` (`textContentFor({es,en}, lang)`) y `tasks.md` T328:670, T332:750 ↔ `plan.md:117` (P3-3) y `research.md:52-56` (R3-2), `contracts/openapi.yaml:684-716` | Se define un **fallback cliente** `textContentFor` sobre pares `{es,en}`, pero el DTO público devuelve strings **ya resueltos por el servidor** (P3-3/R3-2: "sin que el frontend tenga que reimplementar nada"). El helper no tendría qué resolver sobre `PortadaPublica` y sus pruebas no probarían nada real; la única fuente de fallback quedaría duplicada. | O el helper se elimina del sitio público (el fallback lo garantiza el service y se prueba ahí), o se restringe a los formularios del panel (que sí manejan pares es/en). Ajustar T328/T332 para no prometer lógica de fallback cliente sobre datos ya resueltos. |
| **I3** | IMPORTANTE | `ux.md:228` (§4.6: "Duplicado exacto: igual día+hora+nombre+lugar → rechazado") ↔ `data-model.md:120-137` (sin `UNIQUE` en `home_services`), `contracts/openapi.yaml:327-361` (`createScheduleItem` sin `409`), `tasks.md` T320 (sin regla de duplicado de servicios) | `ux.md` promete un rechazo de servicios duplicados ("por default de spec") que **no existe** en la spec (el duplicado solo está definido para canales de WhatsApp), ni en la BD, ni en el contrato, ni en las tareas. | O se quita el texto de `ux.md` §4.6, o se añade la regla de verdad (migración con `UNIQUE`, `409` en el contrato — que es inmutable: sería nueva iteración — y caso de prueba en T320/T336). Hoy el usuario vería un error que el sistema nunca emitirá. |
| **I4** | IMPORTANTE | `ux.md:175` (§4.1 pie: "frase corta de voz de marca si el equipo la carga (campo opcional)"), `ux.md:321,395` (`footer.welcome` "frase editable en panel") ↔ `spec.md:182-188` (Key Entities), `data-model.md`, `contracts/openapi.yaml`, `tasks.md` | Contenido "editable desde el panel" sin soporte en ningún otro artefacto: no hay campo en `home_identity`, ni DTO, ni endpoint, ni tarea. Además `ux.md` lo trata a la vez como cadena de diccionario (`footer.welcome`) y como contenido editable (dos orígenes para el mismo texto). | Decidir: o es contenido → añadir campo (p. ej. `footer_phrase_es/en` en `home_identity` + contrato + tareas), o es cadena de interfaz → quitar "editable en panel" y dejarlo en el diccionario i18n. Cualquier cambio de contrato exige iteración del delta. |
| **I5** | IMPORTANTE | `plan.md:120` (P3-13: `--color-azul`, `--color-turquesa`…, `--font-titulos`/`--font-texto`/`--font-enfasis`) y `research.md:359-367` (R3-12) ↔ `ux.md:27-54` (`navy`, `teal`, `teal-strong`, `navy-soft`, `coral`, `leaf`, `--font-display`/`--font-sans`/`--font-emotiva`) ↔ `tasks.md` T327:649-660 (lista los nombres de `ux.md` pero exige "los nombres de token que fija este plan") | Tres nomenclaturas de tokens de marca y una tarea autocontradictoria: T327 no puede cumplir a la vez su descripción (nombres de `ux.md`) y su criterio de terminado (nombres del plan). El plan dice literalmente que fija los nombres "para que el diseño no invente una paleta paralela". | Fijar **una** nomenclatura (recomendado: la de `ux.md`/T327, más rica y ya contrastada: `navy`, `teal`, `teal-strong`…, mapeando los tokens del plan) y actualizar P3-13/R3-12 para que T327 tenga un único nombre por token. |
| **I6** | IMPORTANTE | `quickstart.md:96` (ejemplo con `"textEn":""`) y `quickstart.md:213` ("prueba dejando `textEn:""`") ↔ `data-model.md:91-92` (`CHECK (char_length(text_en) BETWEEN 1 AND 1000)` rechaza `''`), `contracts/openapi.yaml:1036-1038` (`[string,"null"]`) | El recorrido de verificación usa cadena vacía como "sin inglés", pero el `CHECK` de la BD rechaza `''` (solo `NULL` o 1..1000) y ni el contrato ni las tareas definen la normalización `"" → NULL`. El tester que siga §3 obtendrá `400`/error de restricción donde el documento promete éxito, y el formulario del panel ("campo en inglés vacío", US2 esc. 5) no sabe qué enviar. | Definir la regla en un solo sitio y propagarla: el service normaliza `""`/espacios → `NULL` en todo `*_en` (T317 ya habla de "trim") y se documenta en el contrato (`AboutInput` etc.) y en `quickstart` §3/§7; añadir caso de prueba en T319/T320 y en T334–T336. |
| **I7** | IMPORTANTE | `ux.md:197` ("sin desplazamiento horizontal en el ancho mínimo probado de **320 px**") ↔ `plan.md:357` (SC-007: e2e 360/768/1280), `quickstart.md:258` (§10: 360/768/1280), `tasks.md` T337:848 (360/768/1280) · SC-008 solo con checklist manual (`plan.md:358`, `quickstart.md:256`, RG3-10) | Hueco de verificación: la promesa de diseño (320 px) no se prueba (solo 360 px), y SC-008 ("0 hallazgos críticos de accesibilidad") depende 100 % de una checklist manual sin automatización de apoyo (axe/contraste). Es legible que sea manual, pero el alcance real debe coincidir con lo prometido. | Alinear el ancho (o `ux.md` a 360 px o e2e/checklist a 320 px) y añadir, si el equipo lo acepta, un barrido automatizado de accesibilidad (p. ej. `@axe-core/playwright`) en T337 como respaldo de la checklist, dejando explícito en `quickstart` §10 qué verifica la máquina y qué verifica QA. |
| **I8** | IMPORTANTE | `contracts/openapi.yaml:622-632` (la subida "no deja huella visible"), `research.md:333` (`home.identity.update` "incluye subir/reemplazar logo"), `research.md:317-321` (éxito transaccional; fallo/denegación best-effort), `tasks.md` T323/T321 ↔ `spec.md:173` (FR-017 "cada edición… queda registrada"), `spec.md:222` (SC-013 "100 % de las ediciones registradas") | `POST /api/v1/admin/portada/imagenes` escribe en disco (y el reemplazo/borrado de archivos ocurre best-effort) **sin dejar registro en `admin_actions`**: no hay código `home.*` para subida y T323 no registra nada. `quickstart` §9 ("Tras los cambios de §3–§6… cada edición aparece") promete un registro que §6 no producirá. | Decidir: añadir un código de acción (p. ej. `home.image.upload`, ampliando `000006` **antes** de aplicarse — hoy solo está en diseño) y registrarlo en la transacción correspondiente, **o** dejarlo explícito en FR-017/`quickstart` que la subida se audita solo al guardar la identidad que la referencia. Cualquiera de las dos debe reflejarse en contrato + tareas + §9. |
| **I9** | IMPORTANTE | `quickstart.md:224` (`/panel/portada`) ↔ `ux.md:148,203` (`/panel/informacion`), `tasks.md:59-61,714-732` (alineación 1: `/panel/informacion`) | El paso de verificación de SC-005/SC-011 (§8: "forzar la ruta") cita la ruta **anterior** `/panel/portada`, que ya no existirá: T326/T331 exigen que `quickstart` §8 sea ejecutable tal cual. | Actualizar `quickstart` §8 a `/panel/informacion` (y revisar que no queden más referencias a `/panel/portada` fuera de `plan.md`, cuyo drift está documentado en M1). |

### MENORES (estilo / documentación)

| ID | Severidad | Dónde | Qué | Recomendación |
|---|---|---|---|---|
| **M1** | MENOR | `plan.md:123,206-225` (`/panel/portada`, `features/public`/`features/portada`, `PortadaPage`/`PortadaAdminPage`, `i18n/` en raíz) ↔ `tasks.md:57-67` (alineación de nomenclatura, "no se reabre") | El plan aprobado conserva rutas/carpetas/componentes antiguos; la renombración quedó bien documentada y decidida en `tasks.md`, pero el plan queda como fuente divergente. | Añadir una nota en `plan.md` remitiendo a la "Alineación de nomenclatura" de `tasks.md` (sin reabrir decisiones). |
| **M2** | MENOR | `data-model.md:199` y `data-model.md:322-323` | Cita "Riesgo **RG3-6** del plan" donde corresponde **RG3-2** (el de los nombres de restricción de `000006`; RG3-6 es XSS). `tasks.md:444` sí cita RG3-2 correctamente. | Corregir las dos referencias. (Verificado en el código: los nombres `admin_actions_action_check`, `admin_actions_target_kind_check` y `admin_actions_check1` **sí** son los que genera `000004`, por orden de definición; el riesgo está bien identificado, solo la cita está mal.) |
| **M3** | MENOR | `ux.md:366` ("Usa JPG o PNG de menos de {tamaño}") ↔ `contracts/openapi.yaml:1388` y `research.md` R3-8 (JPEG/PNG/**WebP**) | El mensaje de error de imagen inválida omite WebP, que el servidor sí admite. | Incluir WebP en el texto de `ux.md` §7.1. |
| **M4** | MENOR | `research.md:500-505` (R3-18: fallback con "el nombre 'Iglesia Simiente Santa' del diccionario i18n") ↔ `ux.md:377-403` (catálogo §7.2, sin esa clave) | La clave del nombre de marca para el *chrome* de fallback no está en el catálogo de cadenas que T328 implementa literal. | Añadir la clave (p. ej. `brand.name`) al catálogo §7.2 y a los diccionarios es/en. |
| **M5** | MENOR | `research.md:341-343` (R3-11.1: al cambiar datos y estado se registra `home.publish`/`home.unpublish`) ↔ `contracts/openapi.yaml:337-338` (`createScheduleItem` "Nace en `draft` salvo que se envíe `publicationState: published`") | No se define el código de auditoría de un **alta ya publicada** (`home.schedule.create` o `home.publish`). | Precisarlo en R3-11.1 y en las pruebas de T320 (recomendado: `home.*.create` en el alta, `home.publish/unpublish` solo en cambios de estado). |
| **M6** | MENOR | `quickstart.md:201` ("nombre inventado en `/api/v1/media/…` → `404 not_found` (o `400` si no cumple el formato)") ↔ `contracts/openapi.yaml:182-187` (400 y 404 declarados) y `tasks.md` T323 ("`400`/`404`") | Ambigüedad menor: dos códigos posibles para el mismo caso sin criterio fijo; la prueba e2e/unitaria no sabría cuál exigir. | Fijar: nombre que no cumple el patrón → `400 invalid`; nombre válido pero inexistente → `404 not_found` (coherente con el *pattern* del contrato). |
| **M7** | MENOR | `quickstart.md:298` (SC-010 → "prueba con personas; se documenta al cierre"), `plan.md:360` ↔ `tasks.md` T339 (sin protocolo ni responsable de esa prueba) | SC-010 (≥90 % de personas de distintas edades encuentran horario y contacto) no tiene protocolo, muestra ni responsable explícito. | Añadir a T339/`quickstart` §10 el protocolo mínimo (nº de personas, guion, dónde se documenta el resultado) o declararlo explícitamente fuera de la validación automática con aprobación humana. |

---

## Lo que sí está coherente (comprobado)

- **Cobertura FR ↔ tareas**: FR-001…FR-019 tienen fila en `tasks.md` y en `plan.md` con verificación
  asociada; salvo C1, todos los endpoints del contrato tienen tarea (`getPortada` T322, `getMediaFile`
  y `uploadPortadaImage` T323, `PUT` de singletons T324, `POST/PATCH/DELETE` de listas T325).
- **FR-013 (borradores) y publicación por sección** están reflejados en contrato (sin
  `publicationState` en `PortadaPublica`; secciones omitidas), modelo (`publication_state` en las 6
  tablas, consultas `…Published` separadas) y tareas (T308/T314/T315/T318/T322 + pruebas "0
  borradores expuestos" y e2e) — salvo el camino de archivos de **C4**.
- **Migraciones**: `000005`/`000006` con `up`/`down` completos, `id`/`created_at`/`updated_at`,
  `CHECK`/`UNIQUE`/índices explícitos, ninguna migración existente se modifica; los nombres de
  restricción que `000006` referencia coinciden con `000004` (verificado en el SQL real) y hay prueba
  up→down→up (T316).
- **Permiso `portada` cableado y probado**: sembrado por F2 (`000002`, verificado),
  `AdminChain("portada", deps)` con `Recorder` existe en `platform/middleware` (verificado),
  `RequirePermission`/menú en frontend (T331), 403 por API forzada y denegaciones auditadas
  (T321/T326), e2e de SC-005/SC-011 (T338).
- **Auditoría atómica** (`InsertAdminAction` sobre el `tx`, fail-closed, `target_kind='content'`) en
  T313/T316/T319/T320, con migración `000006` y pruebas de integración — salvo el hueco de subidas
  (**I8**).
- **Bilingüismo de datos**: patrón `*_es NOT NULL`/`*_en NULL` coherente en modelo, contrato y tareas;
  fallback `en → es` en el service con pruebas de tabla de casos (T318) y e2e en inglés (T337) — salvo
  las ambigüedades de **C3** e **I6**.
- **Seguridad**: SQL parametrizado (sqlc), validación en backend (tag `url` nuevo con pruebas), sin
  `dangerouslySetInnerHTML`, imágenes sin SVG/GIF por firma binaria con nombre generado, `nosniff`/
  `inline`/`immutable`, secretos solo por entorno (`UPLOAD_*` no son secretos), dependencias nuevas
  justificadas (0 en backend, 3 `@fontsource` en frontend).
- **Rutas SPA** `/`, `/health` (SPA, distinta de `/healthz`, intacto) coherentes en `ux.md`, `tasks.md`
  y `quickstart` §1; la divergencia `/panel/informacion` está **decidida y documentada** en
  `tasks.md` (solo queda el residuo de **I9** y el drift de **M1**).

---

## Veredicto

**RECHAZADO** — hay **4 CRÍTICOS** (C1: `GET /api/v1/admin/portada` sin tarea de implementación; C2:
horario día/hora contradictorio entre `ux.md` y modelo/contrato/tareas; C3: `ux.md` reintroduce la
contradicción de memoria de idioma que el humano ya resolvió; C4: `/api/v1/media/{fileName}` ignora el
estado de publicación y compromete FR-013/SC-002 "por ninguna vía").

Requisito para volver a analizar: cerrar C1–C4 con decisión documentada (C1 y C2 requieren tocar
`tareas.md`; C2 puede requerir iteración del delta del contrato; C3 es corrección de `ux.md` + prueba
e2e de re-visita; C4 requiere decisión de diseño sobre la descarga de imágenes y su reflejo en
contrato/plan/tareas). Los 9 IMPORTANTES pueden resolverse durante la implementación, pero se
recomienda cerrar I1–I4 e I6 antes de T332–T336 (son los que más retrabajo generan en frontend).

---
---

# Re-análisis 2026-10-09 (segunda pasada)

**Ejecutado por**: `revisor-codigo`, fase `analyze` (re-ejecución tras correcciones). **Alcance**:
los 8 artefactos de `specs/003-portada-info-general/` contrastados de nuevo contra la primera pasada
de este reporte y la constitución. **Diferencias revisadas** (`git diff` sobre
`contracts/openapi.yaml`, `data-model.md`, `plan.md`, `quickstart.md`, `research.md`, `ux.md`;
`tareas.md` reescrito completo, T301–T340): ~300 líneas cambiadas. `spec.md` **no** cambió (sigue
aprobada e intacta ✓).

## Estado por hallazgo de la primera pasada

| ID | Estado | Evidencia de cierre |
|---|---|---|
| **C1** (agregado `GET /api/v1/admin/portada` sin tarea) | **CERRADO** | Nueva **T340** (`tasks.md:683-704`): service `GetPortadaAdmin` + handler + pruebas (singletons `null` al inicio, pares `Es`/`En` crudos, sobres `{items}`, borradores incluidos); en la matriz **FR-011** y **US3** (`tasks.md:1030`, `plan.md:331,348`), en el conteo (40 tareas, 20 `[backend]`) y en las dependencias (T326 depende de T340; T330/T333–T336 la consumen). |
| **C2** (horario día/hora contradictorio) | **CERRADO** | Horario estructurado y coherente en las 5 capas: `day_of_week` 0–6 + `start_time` + **`end_time` opcional** con `CHECK (end_time IS NULL OR end_time > start_time)` (`data-model.md:120-145`), `endTime` en los 4 schemas del contrato (`ScheduleItemPublic/Admin/Input/Patch`), `research.md` R3-5, `plan.md` P3-6, `ux.md` §4.6/D-3 (D-3 **retirada**: `Select` de día localizado + horas "HH:MM"; formato a.m./p.m. es presentación), `tasks.md` T306/T308/T315/T320/T332/T336 y `quickstart.md` §4 (ejemplo con `endTime` y error `400 details.endTime`). Sin texto libre y sin día traducible en ningún artefacto. |
| **C3** (memoria del idioma) | **CERRADO** | `ux.md` §2.1/§2.2.4/§8.2 corregidas: "**solo la primera visita sin preferencia** entra en español", "una visita posterior **respeta la preferencia guardada**"; sin auto-detección del navegador. Probado en T328 ("con preferencia guardada (re-visita) se respeta") y en el e2e T337 ("**re-visita con preferencia guardada respeta el idioma**"), coherente con FR-010/Q11, `research.md` R3-9 y `quickstart.md` §7. |
| **C4** (media expone borradores) | **CERRADO** | Opción (a) implementada: `GET /api/v1/media/{fileName}` **solo** sirve archivos referenciados por contenido **publicado** (`404 not_found` en otro caso; `400 invalid` si el nombre no cumple el patrón) y responde **`Cache-Control: no-store`** (contrato `:150-190`; RG3-8/RG3-9 y P3-9 del plan; R3-8/R3-13). Consulta `IsHomeFilePublished`/`IsFilePublished` con prueba de integración ("al retirar la identidad pasa a `FALSE`"), T323 con caso "identidad en borrador → `404`" y e2e T337/T338; `quickstart.md` §5/§6 documentan retirar→`404`, republicar→`200`. El plan B (URL versionada + `immutable`) queda **cerrado como excepción a FR-013 con aprobación humana**: no aplicado ✓. |
| **I1** (identidad en `ux.md` §4.4) | **CERRADO con regresión** → ver **N4** | Nombre con par es/en y `alt` obligatorio solo en español quedaron alineados (D-8 corregida), pero la corrección hizo **misión/visión "es obligatorio"**, que contradice el contrato (ver N4). |
| **I2** (`textContentFor` en cliente) | **CERRADO** | Retirado del sitio público: `ux.md` §6.2/§8.2/§10, `research.md` R3-2, `plan.md` P3-10 y `tasks.md` T328 (un helper de pares `{es,en}`, si hiciera falta, vive solo en `features/informacion/` para formularios). El fallback es único: el service (P3-3). |
| **I3** (duplicado de servicios inventado) | **CERRADO** | `ux.md` §4.6 retira el rechazo de duplicados de servicios; `tasks.md` T320 lo hace explícito ("**sin regla de duplicados de servicios** (no está en la spec) — ninguna prueba la exige"). Decisión consciente y sin promesa al usuario. |
| **I4** (frase de pie "editable en panel") | **CERRADO con residuo** → ver **N3** | `footer.welcome` queda como **cadena de interfaz i18n**, no contenido del panel (`ux.md` §4.1.9, §6.2, §7.2). |
| **I5** (tokens de marca) | **CERRADO** | Nomenclatura **única** (la de `ux.md` §1): `plan.md` P3-13 y `research.md` R3-12 la adoptan y declaran **retirados** los alias `--color-azul`/`--font-titulos`; T327 ya no es autocontradictorio (exige "cero alias" y verifica con `grep`). |
| **I6** (`""` vs `NULL` en `*En`) | **CERRADO** | Regla única "`""`/espacios → `NULL` (trim en el service)" documentada en el contrato (cabecera + `IdentityInput`/`AboutInput`/`ContactInput`/`WhatsappChannel*`/`ScheduleItemPatch`), `data-model.md` (patrón + invariante), `research.md` R3-1, `quickstart.md` §3/§7 y probada en T317/T319/T320/T334–T336. |
| **I7** (320 px y accesibilidad) | **CERRADO (diferimiento consciente)** | Ancho mínimo **320 px** alineado en `plan.md` SC-007, `quickstart.md` §10.1, `tasks.md` T332/T337; el barrido automatizado de accesibilidad (axe) queda **explícitamente fuera** por dependencia nueva (§II) y SC-008 se valida con la checklist manual de QA: alcance declarado, no hueco. |
| **I8** (subidas sin auditoría) | **CERRADO** | Código **15.º**: `home.image.upload` en `000006` (los **15** códigos son coherentes en contrato, `data-model.md`, `plan.md`, `research.md` R3-11.1, `quickstart.md` §9 y T307/T311/T316/T321/T323), registro **fail-closed** en T323 ("si el registro falla, se aborta y se elimina el archivo") con prueba, y en la matriz FR-017. |
| **I9** (ruta `/panel/portada` en quickstart) | **CERRADO** | `quickstart.md` §8 usa `/panel/informacion`. Única referencia residual en `plan.md`, ya cubierta por M1. |
| **M1** (drift de `plan.md`) | **CERRADO** | Nota de nomenclatura en `plan.md:104-109` que remite a "Alineación de nomenclatura" de `tasks.md` ("ante cualquier duda manda `tasks.md`/`ux.md`"). |
| **M2** (cita RG3-6) | **CERRADO** | `data-model.md` cita **RG3-2** en ambos sitios. |
| **M3** (WebP en el texto de error) | **CERRADO** | `ux.md` §7.1: "JPG, PNG o WebP". |
| **M4** (clave del nombre de marca) | **CERRADO** | `brand.name` en el catálogo §7.2 de `ux.md` y en la nota de fallback del *chrome*. |
| **M5** (código de auditoría del alta publicada) | **CERRADO con regresión** → ver **N1** | Regla exacta definida en `research.md` R3-11.1 (`create` siempre en el alta; `home.publish/unpublish` **solo** cambio de estado; combinado → **dos filas**), `quickstart.md` §9 y T319/T320; pero la prosa por operación del contrato quedó desalineada (ver N1). |
| **M6** (400/404 ambiguos en media) | **CERRADO** | Criterio fijo: patrón → `400 invalid`; válido pero inexistente/no publicado → `404 not_found` (contrato, `quickstart.md` §6, T323). |
| **M7** (SC-010 sin protocolo) | **CERRADO** | Protocolo mínimo en `quickstart.md` §10.7 y T339 (≥6 personas por franjas 18–35/36–59/60+, 3 tareas sin ayuda, informe `pruebas-usabilidad-SC-010.md`, umbral ≥90 %), reflejado en `plan.md` SC-010 y en el mapa §12. |

## Nuevos hallazgos (introducidos o revelados por las correcciones)

| ID | Severidad | Dónde | Qué | Recomendación |
|---|---|---|---|---|
| **N1** | IMPORTANTE | `contracts/openapi.yaml:256-257` (`updateIdentity`), `:290-291` (`updateAbout`), `:324-325` (`updateContact`), `:398` (`updateScheduleItem`), `:497` (`updateWhatsappChannel`), `:592` (`updateSocialLink`) ↔ `research.md:381` (R3-11.1: "deja **dos filas**"), `quickstart.md:268`, `tasks.md:540-548` (T319) y `:567-570` (T320) | La corrección de M5 fijó la regla "una edición que cambia datos **y** estado deja **dos filas** (`home.*.update` + `home.publish/unpublish`)" en research/quickstart/tasks, pero la prosa de las **seis operaciones de edición** del contrato sigue diciendo "Auditoría: `home.X.update` **o** `home.publish`/`home.unpublish` si cambia el estado" (una sola fila). El contrato es la fuente de la que T319/T320/T324/T325 escriben sus pruebas y de la que se afirma SC-013: dos lecturas distintas dan número de filas distinto. | Actualizar las seis descripciones del contrato a la regla de R3-11.1 (p. ej. "registra `home.X.update` y, si cambia el estado, **además** `home.publish`/`home.unpublish`, en la misma transacción"). Es prosa del delta (aún sin fusionar): cambio sin impacto en schemas. |
| **N2** | MENOR | `data-model.md:289,326,345` y `plan.md:390` / `tasks.md:328` (`IsHomeFilePublished`) ↔ `research.md:232`, `tasks.md:438,446,520,634` (`IsFilePublished`) | La consulta que resuelve C4 tiene **dos nombres** (`IsHomeFilePublished` vs `IsFilePublished`) y su atribución varía (T323 dice "`IsFilePublished` de T314" cuando la escribe T308 en `home.sql`). | Un solo nombre (recomendado: `IsHomeFilePublished`, el del `data-model.md` fuente de verdad) y citar la tarea que la escribe (T308) y las que la consumen (T314/T318/T323). |
| **N3** | MENOR | `ux.md:178` (párrafo "Contenido, no decoración") ↔ `ux.md:175` (I4: `footer.welcome` "no es campo del panel; no es contenido gestionado") | Queda la frase residual "Las frases de voz de marca **no las inventa el diseño**: son contenido editable", que contradice la resolución de I4 (la frase del pie es cadena de interfaz, no contenido editable). | Reescribir el párrafo: los **campos del hero** (nombre, lema, misión, visión) sí son contenido editable; la frase del pie es interfaz. |
| **N4** | IMPORTANTE | `ux.md:214` (§4.4: "**Misión** (es obligatorio, en opcional) · **Visión** (es obligatorio, en opcional)") ↔ `contracts/openapi.yaml:986` (`IdentityInput.required: [nameEs, publicationState]`; `missionEs`/`visionEs`/`taglineEs` opcionales), `data-model.md:68-71` (`mission_es`/`vision_es NULL`), `spec.md:171` (FR-015: obligatorios = nombre, «quiénes somos» es, dirección, correo, teléfono) | Regresión de la corrección de I1: `ux.md` ahora exige misión y visión en español como **obligatorias**, un requisito que **no** está en FR-015 (lista cerrada de obligatorios), ni en el contrato (`required` solo `nameEs`+`publicationState`), ni en la BD (columnas `NULL`). El formulario de T334 bloquearía guardados que el servidor acepta, y los mensajes de error prometidos no coincidirían con `details` del backend. | Alinear `ux.md` §4.4 al contrato: lema, misión y visión **opcionales** (es opcional, en opcional), tal como ya quedó el lema. Si el cliente quiere misión/visión obligatorias, es cambio de spec (FR-015) + contrato + `data-model.md` con aprobación humana; no hacerlo solo en la UX. |

## Comprobaciones transversales de esta pasada (sin hallazgo)

- **`endTime` propagado sin contradicciones**: patrón `^([01][0-9]|2[0-3]):[0-5][0-9]$` y
  `CHECK (end_time IS NULL OR end_time > start_time)` idénticos en BD/contrato; validación de dominio
  con `400 details.endTime` (T320) y mensaje de UX ("La hora de fin debe ser posterior…"); opcional en
  los cuatro schemas (no figura en ningún `required`); `endTime: null` lo quita (PATCH).
- **15 códigos de auditoría** consistentes en las 6 fuentes (no quedan "14 códigos" en ningún
  artefacto) y `000006` los contiene literalmente; la regla de códigos es la misma en
  `research.md`/`quickstart.md`/`tasks.md` (salvo la prosa del contrato: N1).
- **`Cache-Control: no-store`** unificado: portada pública y descarga de media; ningún artefacto
  promete ya `immutable` (solo aparece como alternativa descartada/plan B). RG3-9 y el rendimiento
  móvil quedan cubiertos por la justificación de `plan.md` Performance Goals.
- **Tokens**: `navy/navy-soft/teal/teal-strong/white/cream/coral/leaf` + `--font-display/sans/emotiva`
  sin alias vivos en plan, research, ux y T327.
- **Conteos de `tasks.md`**: 40 tareas (20+14+3+3), 15 `[P]`, numeración T301–T340 y dependencias
  (T326 ← T340) coherentes; matriz FR-001…FR-019 completa con T340 en FR-011/US3 y T323 en
  FR-013/FR-017.
- **Rutas SPA**: `/`, `/health`, `/panel/informacion` coherentes en ux/tasks/quickstart; solo el
  drift documentado de `plan.md` (M1, cerrado con nota).
- **Constitución**: migraciones versionadas con up/down e inmutables ✓; capas y contrato antes que
  código ✓; permiso `portada` en servidor + CSRF ✓; validación backend y SQL parametrizado ✓;
  dependencias nuevas: 0 en backend, 3 `@fontsource` justificadas, axe **no** se añade (I7) ✓.

---

## Veredicto (segunda pasada, 2026-10-09)

**APROBADO** — los **4 CRÍTICOS** de la primera pasada quedan **cerrados** y no hay críticos nuevos.

Quedan **2 IMPORTANTES** (N1: prosa de auditoría del contrato desalineada con la regla "dos filas";
N4: `ux.md` vuelve a declarar misión/visión obligatorias, contra FR-015/contrato/BD — regresión de la
corrección de I1) y **2 MENORES** (N2: dos nombres para la consulta de media; N3: frase residual en
`ux.md` §4.1). **Recomendación**: corregir N4 y N1 **antes** de T334/T319–T325 (son cambios de una
línea en `ux.md` y de prosa en el delta del contrato, sin tocar schemas ni migraciones); N2 y N3
pueden resolverse durante la implementación.

Conteo acumulado de la segunda pasada: **0 CRÍTICOS** · **2 IMPORTANTES** · **2 MENORES** abiertos;
20 hallazgos de la primera pasada (4 C + 9 I + 7 M) verificados como cerrados, 2 de ellos con el
residuo anotado (I1→N4, M5→N1).
