# Analyze — Coherencia cruzada de F7 (Ofrendas y diezmos / Donaciones)

**Fecha**: 2026-10-14 · **Fase**: `/speckit.analyze` · **Revisor**: `revisor-codigo` (solo lectura; este
reporte es la única escritura).

**Artefactos contrastados**: `spec.md` (aprobada v2, 25 FR / 14 SC / US1–US6, Q1–Q11 resueltas) ·
`plan.md` (P7-1…P7-21, RG7-1…RG7-12, puerta 2 aprobada) · `research.md` (R7-1…R7-16) ·
`data-model.md` (tablas `donation_methods`/`donation_texts`, migraciones `000012`/`000013`) ·
`contracts/openapi.yaml` (delta, 9 operaciones, 0.8.0) · `ux.md` (D-1…D-16) · `quickstart.md`
(§0–§16) · `tasks.md` (T701–T738) · `estado.md` · código real de F2/F3/F4/F6/F10
(`backend/migrations/`, `frontend/e2e/{helpers,agendaSeed,ministeriosSeed}.ts`,
`frontend/src/features/{roles,panel,publico}`, `frontend/src/lib/permissions.ts`) y skills
`go-backend`/`postgres-db`/`react-frontend`.

**Método**: matriz US1–US6 / FR-001…FR-025 / SC-001…SC-014 ↔ plan ↔ tareas ↔ endpoints ↔ tablas;
contraste punta a punta de las decisiones firmes (ruta `/ofrendas`, nombre «Ofrendas y diezmos»,
permiso `donaciones`, endpoint del bloque, orden por tipo, campos/obligatoriedad, IBAN/SINPE,
moneda solo IBAN, titular opcional, entidad no traducible, `publication_state` por elemento y por
texto, textos `intro`/`usage`); verificación de las cuatro contradicciones conocidas (a)–(d) en
`tasks.md`; revisión de constitución tarea a tarea; cálculo de longitudes de los fixtures IBAN
(ejecutado, no estimado) y verificación de los hechos citados contra el código real.

---

## Resumen ejecutivo

**VEREDICTO: APROBADO CON CAMBIOS MENORES** — sin hallazgos CRÍTICOS. La cobertura es **completa**:
los 25 FR, los 14 SC y las 6 US tienen fila en `plan.md` y al menos una tarea en `tasks.md` con
verificación asociada (sin requisito sin tarea ni tarea sin requisito); las 9 operaciones del
contrato tienen implementador; y las **cuatro contradicciones conocidas quedan arbitradas
explícitamente** en `tasks.md` §9 con la lectura correcta (navegación desde el documento de la
sección, endpoint `portada/ofrendas`, entidad **no** traducible, duplicado por tipo+número). Quedan
**3 hallazgos ALTOS** y 7 MEDIOS. Los altos son acotados pero duele ignorarlos: (A1) los **fixtures
del IBAN de Costa Rica contradicen la propia regla de 22 caracteres** en T707 y el quickstart —el
caso "aceptado" de la tabla de pruebas normaliza a 21 y el "normalizado" prometido tiene 24—, lo
que enmarraría el test-first de la regla más crítica del módulo; (A2) `ExistsDonationMethodDuplicate`
tal como está firmado (`:kind`, `:number`) **no excluye la fila que se edita**, con lo que toda
edición chocaría consigo misma en un `409` y ninguna prueba del plan lo detectaría; y (A3) `ux.md`
sigue fijando la **presencia de la navegación desde el bloque** (y dice literalmente que
`usePublicOfrendas` "no alimenta la presencia"), justo lo contrario de P7-8/T727. Los MEDIOS son
textos de `ux.md` sin realinear, una mis-cita de la spec en research/plan/tasks y un par de huecos
de trazabilidad/orden en `tasks.md`. Ninguno cambia arquitectura, contrato ni modelo.

---

## Hallazgos

### CRÍTICOS (bloquean implementar)

*(ninguno)*

### ALTOS (resolver antes de ejecutar las tareas citadas)

| ID | Severidad | Dónde | Qué | Recomendación |
|---|---|---|---|---|
| **A1** | ALTO | `tasks.md:411` (T707, casos **aceptados**: `CR01 0000-0000-0000-0000-0`), `quickstart.md:103,112` (§3: envía ese número y promete `"CR0100000000000000000000"`), `quickstart.md:146` (§4: "ACEPTADO: IBAN costarricense (22 caracteres) … → `"CR0100000000000000000000"`"), `quickstart.md:152-154` (rechazado "CR exige 22 caracteres: **faltaba uno**"), `quickstart.md:159` (presentación `CR01 0000 0000 0000 0000 0000`), `quickstart.md:239` (PATCH con `CR01 0000-0000-0000-0000-1`), `research.md:82` (mismo ejemplo) ↔ regla `CR` = 22 (`plan.md:148` P7-3, `research.md:92` R7-2, `data-model.md:316`, `spec.md` FR-015/SC-009) | **Los fixtures del IBAN costarricense contradicen la regla que implementan** (longitudes medidas, no estimadas): `CR01 0000-0000-0000-0000-0` normaliza a **21** caracteres (4+17) y el "normalizado" prometido `"CR0100000000000000000000"` tiene **24**; el caso rechazado de `quickstart.md:152` tiene **20** ("faltaban dos", no "faltaba uno"); la presentación agrupada de `quickstart.md:159` tiene 24 caracteres (un CR real agrupa `CR01 0000 0000 0000 0000 00`). ES/DE/GB sí cuadran (24/22/22). Consecuencia real: el test-first de T707 escribiría primero una tabla de casos **rota** —el "aceptado" fallaría contra la regla `CR`=22— y el camino de menor resistencia sería **relajar la longitud** (rompiendo SC-009: "0 IBAN fuera de la estructura/longitud válida") en vez de corregir el fixture. Además el quickstart es el guion de QA de §4/SC-009. | Fijar un único ejemplo canónico de 22 caracteres (p. ej. entrada `CR05 0152 0200 1234 5678 90` → normalizado `CR05015202001234567890`, 22) y usarlo en T707, `research.md:82`, `quickstart.md` §3/§4/§8; corregir el rechazado a 21/23 reales y la presentación agrupada a `CR05 0152 0200 1234 5678 90`. Añadir a T707 un caso explícito "CR de 21 y de 23 → rechazado" ya cuadrado con la regla. |
| **A2** | ALTO | `research.md:313-329` (R7-9: comparación "igualdad simple sobre columnas"), `data-model.md:278` (`ExistsDonationMethodDuplicate` (`:kind`, `:number`)), `tasks.md:451` (T709: misma firma), `tasks.md:594,613` (T715: "duplicado exacto → `409`" sobre el resultado **fusionado**), `contracts/openapi.yaml:178,501-502` (PATCH → `409` "duplicado") ↔ `research.md:499-503` (R7-15: la transacción escribe **la fila completa ya fusionada**) | **La consulta de duplicados no excluye la fila que se está editando.** Bajo R7-15 todo `PATCH` revalida el conjunto fusionado y devuelve `409` si existe un par `kind`+`number` igual: pero ese par es **el propio registro** en cualquier edición que no cambie el número (editar la entidad, el titular, el orden…), es decir, **toda edición devolvería `409` contra sí misma**. Ninguna prueba del plan lo detecta: las unitarias usan fakes (el fake devolvería `false` si se le programa así) y el test de integración de T709 prueba el duplicado en alta y el `Update` aislado, no "editar sin cambiar el número". Si el implementador improvisa (saltarse el chequeo si `number` no cambia), se pierde el caso "cambio el número al de otro método", que sí exige `409`. | Fijar la firma `ExistsDonationMethodDuplicate(:kind, :number, :exclude_id)` con `AND id <> :exclude_id` (o comparar contra la fila previa) en `data-model.md`, T703/T709/T715 y la prosa del contrato; añadir a T715/T709 la prueba **"editar un método sin cambiar el número no devuelve `409`"** y **"PATCH que cambia el número al de otro método devuelve `409`"** (service + integración). |
| **A3** | ALTO | `ux.md:136` (§3: "la entrada existe **solo** si la consulta del **bloque** (`usePortadaOfrendas`)…"), `ux.md:448` (`usePublicOfrendas`: "**No** alimenta la presencia en `useSiteLinks`"), `ux.md:449` (`usePortadaOfrendas`: "compartida por `queryClient` con `useSiteLinks` para deducir la presencia"), `ux.md:600` (§10: "La presencia en la navegación **sale del bloque**, no de `/ofrendas`") ↔ `plan.md:153` (P7-8: presencia **desde el documento de la sección**), `plan.md:504-507` (Complexity 1), `tasks.md:139-144` (§9.1: "manda el plan"), `tasks.md:894-906` (T727) | La contradicción (a) está **bien dirimida en `tasks.md`** con la única lectura coherente con P7-13 (la sección vive con ≥1 método **o** ≥1 texto: con la regla de `ux.md`, un introductorio publicado sin métodos dejaría la sección viva **sin entrada de navegación**). Pero `ux.md` **no quedó realineado** y sus renglones son la fuente de "responsabilidades de hooks" que copia el `dev-frontend` (Fase 7 manda seguir `ux.md` §6): quien construya `usePublicOfrendas` desde `ux.md:448` escribiría literalmente lo contrario de T727, y `ux.md:136,449,600` describen un cableado (`usePortadaOfrendas` → `useSiteLinks`) que T727 prohíbe. Es el mismo patrón de error que en F4 (texto viejo copiado, no arquitecto equivocado). | Anotar en `ux.md` §3/§6.3/§10 que los renglones de presencia están **superados por P7-8** (o dejar en T722/T727 constancia explícita de que `usePublicOfrendas` **sí** alimenta `useSiteLinks` y `usePortadaOfrendas` **no**), y fijar en T727 su prueba de que con solo `intro`/`usage` el ítem existe (ya está) **y** que el bloque no es fuente de navegación. |

### MEDIOS (resolver durante la implementación / antes de QA)

| ID | Severidad | Dónde | Qué | Recomendación |
|---|---|---|---|---|
| **M1** | MEDIO | `research.md:194-195` (cita de FR-019: "los textos (**entidad y titular**) DEBEN ser traducibles…" y de Key Entities: "Los textos que lo acompañen (**entidad, titular**) son traducibles…"), `plan.md:62-63,150,487,590-596` ("traducible **según FR-019/Key Entities**", RG7-1, "Preguntas para el humano" #1), `tasks.md:48-49,1345-1348` ("La contradicción de FR-019/Key Entities… pregunta #1 al humano") ↔ `spec.md:213` (FR-019: "**solo el titular** DEBE ser traducible… el número, el tipo, **la entidad financiera** y la moneda NO son traducibles") y `spec.md:229` (Key Entities: "**Solo el titular** es traducible") | **Los artefactos citan una contradicción de la spec que ya no existe.** La spec aprobada (v2) fue corregida en fase 2 (bitácora de `estado.md:52`) y hoy FR-019/Key Entities coinciden con FR-007/US4 esc. 5/SC-006: la entidad **no** es traducible. Los textos citados por `research.md` **no están** en `spec.md`: la "contradicción de la spec" y la "pregunta #1 al humano" quedan como deuda de redacción. El default aplicado (`institution` de una sola columna) es exactamente lo que dice la spec, así que **nada se rompe**, pero tres artefactos presentan como abierto algo cerrado y citan enunciados inexistentes — un revisor o el humano podrían reabrir una decisión ya tomada. | Corregir las citas de `research.md:194-195` (o marcarlas como "redacción v1, superada") y re-cuadrar `plan.md` RG7-1/"Preguntas" #1 y `tasks.md:48-49,1345-1348` como "**ya resuelto por la spec v2** (solo el titular es traducible)", dejando el default P7-5 como implementación literal del FR-007/FR-019 vigentes. |
| **M2** | MEDIO | `ux.md:59` (§2.2: contenido en inglés incluye "entidad si el equipo lo tradujo allí"), `ux.md:264` (§4.3: "Entidad (español si existe — **es texto traducible es/en**)"), `ux.md:280` (§4.4: "Pestaña English: **entidad y titular** en inglés"), `ux.md:562-563` (§8.2: "los dos textos, **la entidad** y el titular son textos traducibles… la entidad puede llevar versión en inglés") ↔ `plan.md:150` (P7-5), `tasks.md:147-150` (§9.3: las pestañas "solo contienen el titular"), `tasks.md:988-990` (T730), `contracts/openapi.yaml:388-390` | La contradicción (c) quedó dirimida (entidad **no** traducible; `tasks.md` §9.3/T730 lo fijan bien), pero `ux.md` conserva **cuatro** renglones que dan la entidad por traducible y `ux.md §7.1/§4.4` son la fuente de textos y de composición del formulario que copia T730: un dev que siga `ux.md:280` pondría "Entidad financiera" en la pestaña English, contra FR-007/SC-006 y el contrato (sin `institutionEn`). | Alinear `ux.md` §2.2/§4.3/§4.4/§8.2 a "solo el titular es traducible; `institution` es campo no lingüístico" (o declararlos superados en T730), y añadir a T730 la prueba "la pestaña English **no** ofrece campo de entidad". |
| **M3** | MEDIO | `ux.md:327` (§4.9: "propuesta `GET /api/v1/portada/donaciones?lang=es|en`… **DTO mínimo**"), `ux.md:449` (igual), `ux.md:599` (§10: igual) ↔ `contracts/openapi.yaml:273-325` (`GET /api/v1/portada/ofrendas`, `PortadaOfrendas` → **mismo** `MetodoDonacionPublico`), `plan.md:159` (P7-14: "DTO de bloque más pequeño (duplicaría esquema)" descartado), `tasks.md:145-146` (§9.2) | La contradicción (b) quedó dirimida (manda el contrato, `tasks.md` §9.2), pero `ux.md` conserva la ruta alternativa **y** el "DTO mínimo", que contradice la regla explícita de FR-022/R7-13 ("el bloque es un extracto de los mismos métodos, **no un modelo nuevo**"). El "DTO mínimo" induciría a crear un segundo esquema de ítem justo lo que el plan prohíbe. | Sustituir en `ux.md:327,449,599` la ruta por `/api/v1/portada/ofrendas` y "DTO mínimo" por "mismo `MetodoDonacionPublico`; la tarjeta compacta **recorta campos en la composición**" (texto de `tasks.md` T728/T714). |
| **M4** | MEDIO | `ux.md:97` (§2.5: "Duplicado (**mismo tipo, misma entidad y mismo número**)"), `ux.md:492` (§7.1: «Ya existe esa misma cuenta o número (**mismo tipo, entidad y número**). Edita el que ya tienes o **cambia algún dato**.») ↔ `plan.md:155` (P7-10), `research.md:313-329` (R7-9), `tasks.md:151-152` (§9.4), `contracts/openapi.yaml:178` | La contradicción (d) quedó dirimida (regla firme: **tipo + número normalizado**), pero el texto de error de `ux.md` describe la regla triple de Assumptions y, peor, aconseja "cambia algún dato": bajo la regla firme, cambiar la entidad **no** resuelve el duplicado (solo cambiar el número o editar el existente). Es un mensaje activamente engañoso para el equipo del panel. | Cambiar el mensaje (p. ej. «Ya existe un método con ese mismo tipo y número. Edita el que ya tienes o cambia el número.») en `ux.md` §7.1 y §2.5, y fijar en T730/T715 que el aviso refleja la regla tipo+número. |
| **M5** | MEDIO | `tasks.md:1339` ("Defaults revisables aplicados (**preguntas abiertas para el humano**; nada se implementa como regla cerrada sin esto)") y `tasks.md:1345-1362` ↔ `estado.md:29,45` (el humano **aceptó** en la puerta 2 los defaults: "IBAN solo estructura/longitud (sin mod-97), SINPE 6/7/8, sección visible con solo textos publicados y 0 métodos… duplicados por tipo + número (409); los textos no se borran"), `plan.md:584-611` | Los puntos 2–6 de la lista de `tasks.md` ya **fueron aceptados por el humano** al aprobar el plan (y el punto 1, la entidad, lo cierra la propia spec — M1). Solo queda genuinamente abierto el punto 7 (presentación del SINPE, `ux.md` D-4 ⟲). Presentarlos como "preguntas abiertas" contradice el registro de aprobaciones y puede provocar que QA/`seguridad` re-abran defaults ya firmes, o que el PR los presente como pendientes. | Re-etiquetar la sección como "**defaults aceptados en la puerta 2 (2026-10-14)**" citando `estado.md:45`, y dejar solo D-4 como pendiente del humano; ajustar la coletilla de `plan.md:584` ("nada se implementa… sin esto") con la constancia de que el humano ya respondió. |
| **M6** | MEDIO | `tasks.md:115-119` (§9.2: "Las utilidades de presentación viven en `features/publico/ofrendas/format.ts` + `features/donaciones/lib/numero.ts`"), `ux.md:460` (§6.4: `lib/numero.ts`), `tasks.md:156` (D-4 sería "una línea en `lib/numero.ts`"), `tasks.md:1365` ("T726/`lib/numero.ts`") ↔ `tasks.md:857` (T726 crea `format.ts`) y `tasks.md:985-986` (T730 crea `NumeroDisplay.tsx`, `validation.ts`, `errors.ts` — **sin `lib/numero.ts`**) | Dos problemas de trazabilidad: (1) **`features/donaciones/lib/numero.ts` no figura en el listado de "Archivos" de ninguna tarea** (T730, su dueño natural, no lo lista), aunque §9.2/§6.4 lo dan por existente; (2) la lógica de agrupado del IBAN queda **duplicada** en `format.ts` (público) y `lib/numero.ts` (panel) con el invariante RG7-7 ("quitar los separadores devuelve `number`") a mantener en **dos** implementaciones, y §9.5:156 vs 1365 ubican el cambio de D-4 en archivos distintos (T726 trabaja con `format.ts`, no con `lib/numero.ts`). | Añadir `lib/numero.ts` (+ su test) a los archivos de T730, y decidir una sola fuente de presentación (p. ej. `format.ts` compartido o `lib/numero.ts` reexportado) con la prueba RG7-7 en el lugar único; corregir la referencia cruzada de §9.5/§1365. |
| **M7** | MEDIO | `tasks.md:213` (T732 "Depende de: **T729–T731**") ↔ `tasks.md:1044` (preámbulo de Fase 8: "Toda la Fase 8 exige el **backend completo (T721)** integrado y la base de F7 migrada") y `frontend/e2e/ministeriosSeed.ts:4` / `agendaSeed.ts:4` (los seeds del patrón crean datos **por API**: `apiSend` sobre `API_URL`, no por la UI del panel) | La dependencia de T732 sobre las tres tareas de UI del panel es **artificial** e internamente contradictoria con el preámbulo de Fase 8: las fixtures (patrón F4/F6) se siembran por API y solo necesitan T721 (y las cuentas de login). El efecto es serializar toda la Fase 8 detrás de la cadena T729→T730→T731 y retrasar T733/T736, que ni siquiera usan el panel. | Cambiar la dependencia de T732 a **T721** (backend + migraciones + helpers de login), dejando T729–T731 como dependencia **solo** de T734; o justificar explícitamente por qué las fixtures necesitan las pantallas del panel. |

### BAJOS (estilo / documentación / precisión)

| ID | Severidad | Dónde | Qué | Recomendación |
|---|---|---|---|---|
| **B1** | BAJO | `tasks.md:328` (T704: "las **4** de panel `GET/POST /api/v1/admin/donaciones` y `GET/PATCH/DELETE …/{id}`") ↔ `contracts/openapi.yaml` (5 operaciones de panel: GET, POST, GET/{id}, PATCH/{id}, DELETE/{id}) y `plan.md:112` ("2 públicas · **7** de panel") | El recuento "4 de panel" es erróneo (son 5; más las 2 de textos = 7 de panel), aunque el total **9** es correcto en `tasks.md:14` y `plan.md:112`. | Corregir a "las 5 de CRUD de métodos y las 2 de textos". |
| **B2** | BAJO | `tasks.md:988` ("(D-9 de la alineación §9: `institution` no es traducible)") vs `tasks.md:964,993,1010` (D-9 = "tipo no editable", `ux.md:586`) y `ux.md:263` (D-9 = reutilización de `NumeroDisplay`) | La referencia "D-9" apunta a dos decisiones distintas (la 3ª de la alineación §9 y la D-9 de `ux.md`), y en `ux.md` D-9 se usa además para otra cosa (T729: "copia exacta — D-9 —"). Confuso para quien implemente o revise. | Citas explícitas: "alineación §9.3" para la entidad y "ux.md D-9" para el tipo fijo; renumerar o nombrar los puntos de la alineación (A9.1…A9.5). |
| **B3** | BAJO | `tasks.md:8` ("D-1…D-16; **pendientes de confirmación humana**") y `ux.md:5` (cada decisión lleva su marca ⟲) ↔ tareas que las implementan como firmes (T726 con D-2/D-3/D-5/D-6, T728 con D-12/D-14/D-15, T730 con D-9, T731 con D-11…); solo D-4 queda diferida (`tasks.md:153-157`) | Las 16 decisiones de UX están marcadas para revisión humana pero las tareas las ejecutan sin que conste su confirmación (salvo lo arbitrado en §9). Funcionalmente inocuo — son composición, no API — pero el registro de aprobaciones no las cubre. | Recoger la confirmación de D-1…D-16 (menos D-4) en la próxima puerta, o declarar en `tasks.md` que se ejecutan como **default revisable** de composición. |
| **B4** | BAJO | `spec.md:7` ("Pendiente de la aprobación humana (puerta 1) antes de la planificación técnica") ↔ `estado.md:28` (spec **aprobada** el 2026-10-14) y `tasks.md:5` ("APROBADA v2") | La línea de estado del documento congelado quedó desactualizada (mismo caso que F4/M5). | No editar la spec (congelada): dejar la constancia aquí y en `estado.md`. |
| **B5** | BAJO | `quickstart.md:303-305` (§11: guion SC-012 con tarea 3 = "**cambia el idioma** y comprueba que los números siguen iguales") ↔ `tasks.md:1177-1180` (T737: tarea 3 = "**copiar o transcribir** un número de cuenta sin error") | El guion de la prueba de usabilidad de SC-012 difiere en su tercera tarea entre el protocolo manual y la tarea de cierre (SC-001, cronometrado sobre la tarea 1, sí coincide). | Unificar el guion (propongo el de `quickstart.md`, que además cubre SC-006/SC-009) en uno de los dos documentos. |
| **B6** | BAJO | `spec.md:158` (edge case: "se muestran **todos**… **sin límite artificial** (default revisable) ni paginación en el MVP") ↔ `contracts/openapi.yaml:230-232,676-679` (`maxPublicItems`: "**máximo 200** métodos"), `plan.md:497` (RG7-11), `tasks.md:517-518` | El tope de seguridad 200 (constante del service, documentada **en el contrato** — a diferencia del caso análogo de F4) coexiste con el "sin límite artificial" de la spec: es un default revisable bien propagado (plan/contrato/tareas/pruebas), pero figura en las "Preguntas para el humano" y en el edge case de formas opuestas. | Sin cambio de código: listar `maxPublicItems = 200` entre los defaults revisables de la entrega (T738) para que el humano lo tenga a la vista. |
| **B7** | BAJO | `ux.md:586` (D-9 ⟲: "el **tipo no es editable** al editar") y `tasks.md:993,1010` (T730: sin selector, "el guardado conserva el tipo") ↔ `contracts/openapi.yaml:899-900` (`DonacionPatch`: "Cambiar `kind` **revalida** el número y la moneda del conjunto fusionado") | La UI nunca envía `kind` al editar pero la API lo admite y el contrato documenta ese caso: no es contradicción funcional (la restricción es de composición), pero la divergencia no está declarada y el test de "el guardado conserva el tipo" solo la cubre desde la UI. | Una línea en T730/T715: "la UI no edita `kind` (D-9); la API lo permite por contrato — caso cubierto por la revalidación del `PATCH`" (y su prueba en T715 ya existe: "cambiar `kind` revalida…"). |

---

## Lo que sí está coherente (comprobado)

- **Cobertura US/FR/SC ↔ tareas**: US1–US6, FR-001…FR-025 y SC-001…SC-014 tienen fila en
  `plan.md` ("Cobertura de requisitos" y "Métricas") y en `tasks.md` (1254-1308) con verificación
  asociada; **sin FR/SC sin cobertura y sin tarea sin trazabilidad** (T704 "contrato de todos los
  FR", T705 artefacto generado, T732 fixtures, T737/T738 cierre). Las 9 operaciones del contrato
  tienen implementador (T713/T714/T717/T718 públicas; T719/T720 de panel) y los edge cases sin FR
  propio están recogidos en el bloque final de `tasks.md` (1310-1327).
- **Decisiones firmes de punta a punta** (contrastadas en los siete documentos): ruta SPA
  **`/ofrendas`** y nombre **«Ofrendas y diezmos»**; módulo/permiso **`donaciones`** sin cambio de
  nombre y **sin permisos nuevos**; endpoint del bloque **`GET /api/v1/portada/ofrendas`** con
  tope 5 del servidor y **mismo `MetodoDonacionPublico`**; orden `ORDER BY CASE kind WHEN 'iban'
  THEN 0 ELSE 1 END, sort_order, id` (iban → sinpe, `sortOrder, id` por grupo) idéntico en
  data-model, contrato, plan, tasks y quickstart; campos y obligatoriedad (`kind`/`number`/
  `institution` obligatorios, `holder` opcional, `currency` **solo y obligatoria** en IBAN y
  **prohibida** en SINPE, sin notas); IBAN estructura+longitud por país con guardado normalizado y
  agrupado en el cliente (copia = valor exacto); SINPE `+506XXXXXXXX`; entidad **no traducible**
  (una columna) y titular/`body` `*_es`/`*_en` con español por campo; `publication_state` **por
  método y por texto** con consultas `…Published`; textos `intro`/`usage` con `UNIQUE (text_key)`,
  `PUT` singleton y sin borrado.
- **Cuatro contradicciones conocidas (a)–(d) resueltas coherentemente en `tasks.md` §9**, cada una
  con la lectura que garantiza los SC: (a) navegación desde el documento de la sección (P7-8, única
  compatible con "la sección vive con solo textos"); (b) manda el contrato `portada/ofrendas`;
  (c) entidad no traducible (única lectura que hace cumplible SC-006); (d) duplicado `409` por
  **tipo + número normalizado**. Lo que resta es texto de `ux.md` sin realinear (A3, M2–M4).
- **Hechos verificados contra el código** (los que los artefactos afirman): `000002` siembra
  `('donaciones', 'Donaciones')` (`000002_create_roles_and_permissions.up.sql:25`) y el catálogo
  queda en **10** tras `categorias` (F4/`000009`); las migraciones terminan en **`000011`** (F6), de
  modo que `000012`/`000013` continúan sin huecos y ninguna se edita; `admin_actions_action_check`
  existe **con nombre propio desde `000006`**; `isPermissionAvailable('donaciones')` es hoy `false`
  (`features/roles/permissions.test.ts:24`, tal como cita R7-6); las credenciales del quickstart §2
  coinciden con `frontend/e2e/helpers.ts:28-30`; `HomePage` renderiza `MinisteriosBlock`
  (`HomePage.tsx:100`) antes de `WhoWeAreSection` (`:104`), exactamente la posición del bloque;
  `AccesosRapidos.tsx`/`InicioPage.tsx` existen (T724 tiene dueño — el hueco que en F4 fue I4 no se
  repite); `useSiteLinks` ya consume las tres fuentes que describe R7-7.
- **Constitución**: capas `handler → service → repository` intactas (T707→T712→T717, sin atajos);
  **contrato antes del código** (T704 en fase 2, snapshot inmutable, fusión aditiva 0.7.0→0.8.0,
  `error.code` sin cambios); migraciones versionadas con `up`/`down` completos (el `destructivo a
  propósito` de `000013` sigue el precedente de `000006`/`000008`/`000011`) y con
  `id`/`created_at`/`updated_at`; SQL parametrizado vía sqlc; validación de toda entrada en backend
  con `details` y **sin guardado parcial**; cobertura **≥ 80 % en `service*.go`** exigida y medida
  en T737; **pruebas de repository contra PostgreSQL real** (T709–T711, con up→down→up y la regla
  "0 borradores"); **0 dependencias nuevas** (R7-16, sin librerías de IBAN/teléfonos); sin tocar el
  kit, `media`/`IsFilePublished` (Q6) ni el CI.
- **Auditoría**: los 8 códigos `donation.*` son idénticos en `research.md` R7-10, el `CHECK` de
  `000013` (up y down), el enum del contrato, `platform/audit` (T706) y `ACTION_LABELS` (T724); la
  regla "el alta deja **solo** `donation.create`; datos + estado en edición → **dos filas**;
  textos sin `create`; fail-closed transaccional" es la misma en research, data-model, contrato,
  tasks y quickstart §13. El conteo del `down` (8+15+15+8) cuadra con las migraciones reales.
- **Pruebas sin huecos estructurales**: unitarias (numero, service con fakes, handler `httptest`,
  audit, mapeos), integración (repository + migraciones + atomicidad), frontend (Vitest/MSW por
  tarea), **4 specs e2e** (pública, panel, accesibilidad **por teclado** — el "e2e de teclado" que en
  F4 quedó sin dueño aquí es T735 — y portada), y manuales (checklist WCAG §11, usabilidad SC-012
  con cronómetro para SC-001). Test-first exigido tarea a tarea; el único cambio sobre pruebas
  ajenas (T725/RG7-6) está acotado, documentado y **sin** tocar backend.

---

## Verdades incómodas / riesgos residuales

1. **Los fixtures de datos son el punto más frágil de la fase** (A1): el módulo entero se juega la
   confianza en "0 números mal formados" y la tabla de casos del IBAN ya nace con un CR imposible.
   Si A1 no se corrige antes de T707, el test-first empuja —con toda su buena intención— hacia un
   `CR` de 21 caracteres aceptado, es decir, hacia relajar la regla para que el test pase.
2. **`ExistsDonationMethodDuplicate` es un diseño incompleto, no un detalle** (A2): firmado así,
   cualquier implementación literal rompe la edición completa del módulo (US2 esc. 6, SC-004) y
   ninguna prueba planeada lo vería. Es el típico bug que aparece en QA end-to-end como "no puedo
   editar nada" y cuya causa está en el diseño, no en el código.
3. **`ux.md` vuelve a ser el artefacto más débil** (A3, M2–M4): cuatro puntos quedaron sin realinear
   tras la puerta 2 y `tasks.md` los arbitra bien —pero `ux.md` es la referencia que copia el
   `dev-frontend`—. El riesgo no es de arquitectura: es **copiar texto viejo** (navegación por el
   bloque, entidad traducible, endpoint y DTO "mínimo", mensaje de duplicado engañoso).
4. **Los "defaults abiertos" ya no están abiertos** (M5): el humano aceptó los seis defaults
   sustantivos en la puerta 2; presentarlos como preguntas pendientes diluye la responsabilidad del
   registro de aprobaciones y puede generar ruido en QA/seguridad. Queda uno solo: la presentación
   del SINPE (D-4).
5. **Duplicación silenciosa de presentación** (M6): `format.ts` y `lib/numero.ts` con la misma
   responsabilidad y el invariante RG7-7 por duplicar, más un archivo (`lib/numero.ts`) que ninguna
   tarea crea. Pequeño hoy, molesto cuando D-4 se apruebe y haya que tocar "una línea" en dos
   sitios distintos según el documento que se lea.
6. **Tres defaults "revisables" viajarán a la entrega sin dueño humano** si nadie los lista:
   `maxPublicItems = 200` (B6), la presentación SINPE (D-4) y el guion de SC-012 (B5). Conviene que
   el `documentador` los saque a la luz en el PR.
