# Analyze — Coherencia cruzada de F4 (Eventos y actividades)

**Fecha**: 2026-10-11 · **Fase**: `/speckit.analyze` · **Revisor**: `revisor-codigo` (solo lectura; este
reporte es la única escritura).

**Artefactos contrastados**: `spec.md` (re-aprobada, 24 FR / 13 SC / US1–US6, Q1–Q7 resueltas) ·
`plan.md` (P4-1…P4-17, RG4-1…RG4-10, puerta 2 aplicada) · `research.md` (R4-1…R4-15) ·
`data-model.md` (tablas `agenda_*`, migraciones `000007`/`000008`/`000009`) ·
`contracts/openapi.yaml` (delta, 17 operaciones) · `ux.md` (D-1…D-23) · `quickstart.md` (§0–§14) ·
`tasks.md` (T401–T441) · código real de F2/F3/F10 (`backend/migrations/`, `internal/usuarios`,
`platform/middleware`, `internal/portada`, `frontend/src/features/{auditoria,informacion,publico,
roles,panel}`, `frontend/e2e/helpers.ts`, `.env.example`) · constitución y skills
`go-backend`/`postgres-db`/`react-frontend`.

**Método**: matriz US1–US6 / FR-001…FR-024 ↔ plan ↔ tareas ↔ endpoints ↔ tablas; rastreo de cada
camino público ("lo que viene", imágenes, borradores, vencidos, sección vacía); contraste de
permisos (`eventos`/`actividades`/`categorias`) de punta a punta; verificación de la excepción
aprobada de F2 contra el código real; y colisión de cada tarea con F2/F3/F10.

---

## Resumen ejecutivo

**VEREDICTO: APROBADO** — sin hallazgos CRÍTICOS. La cobertura es completa: los 24 FR y las 6 US
tienen fila en el plan y al menos una tarea con verificación; las 17 operaciones del contrato tienen
quién las construye; el permiso nuevo `categorias` está cubierto de punta a punta (migración
`000009` → init → `AdminChain("categorias")` → visibilidad de panel → contrato → e2e); y la
excepción aprobada de F2 (5 aserciones + fixture `sampleCatalog()`, 9 → 10 permisos) es exacta,
coherente entre plan y tareas, está marcada como excepción y **verificada contra el código** (las
líneas citadas existen y fijan 9). Quedan **5 hallazgos IMPORTANTE** y 6 MENORES: casi todos son
puntos en los que `ux.md` (redactado antes de la puerta 2) quedó **sin realinear** con las
decisiones firmes y dos promesas del plan/tablas de cobertura que ninguna tarea implementa (la
constante de seguridad "200 por bloque" de RG4-8 y las accesos rápidos de `InicioPage`). Ninguno
impide implementar: en todos los casos existe una autoridad clara (spec + contrato + research +
tasks), pero conviene corregirlos antes de ejecutar las tareas de frontend para no copiar textos ni
reglas obsoletas.

---

## Hallazgos

### CRÍTICOS (bloquean implementar)

*(ninguno)*

### IMPORTANTES (resolver antes o durante la implementación)

| ID | Severidad | Dónde | Qué | Recomendación |
|---|---|---|---|---|
| **I1** | IMPORTANTE | `ux.md:479` (§8.2: "con **zona horaria del navegador** (coherente con `lib/format.ts`)") y `ux.md:509` (D-22) ↔ `research.md:125-126` (R4-3: "El valor nunca se recalcula de zona horaria: es la fecha y la hora **del lugar**"), `contracts/openapi.yaml:815-824` (`Date`/`ClockTime`: "No es un timestamp: no se convierte de zona horaria"), `spec.md:110,191` (FR-009/US5 esc. 5: "sin cambiar el día ni la hora reales"), `tasks.md` T432 (no fija la regla de parseo) | Contradicción sobre la zona horaria de **presentación**: `ux.md` manda formatear en la TZ del navegador apuntando a `lib/format.ts`, y research/contrato dicen que el dato es local del lugar y **no se convierte**. La referencia a `lib/format.ts` es además una trampa real: ese helper (`frontend/src/lib/format.ts:15-23`) hace `new Date(value)` sobre **timestamps** (`createdAt`, última conexión) y formatea en la TZ del navegador; aplicado a un `YYYY-MM-DD` (que `new Date` parsea como medianoche UTC) produce un **desfase de un día en cualquier TZ negativa — incluida `America/Costa_Rica` (UTC−6)**: el 17 se mostraría como "viernes 16 de octubre" para los propios visitantes de la iglesia. Es exactamente lo que FR-009 prohíbe. | Fijar en T432 que las fechas `YYYY-MM-DD`/`HH:MM` se formatean **sin conversión de zona** (parsear las partes, o `Intl` con `timeZone: 'UTC'`/modo *date-only*), con prueba de que `formatLongDate('2026-10-17','es')` → «viernes 17 de octubre» en cualquier TZ; y corregir/anotar `ux.md` §8.2/D-22 para que diga "presentación del dato local del lugar, sin conversión de zona" (coherente con R4-3 y el contrato inmutable). |
| **I2** | IMPORTANTE | `ux.md:208` (§4.1 estados: "Elemento del día en curso … **incluso si una hora ya pasó (el criterio es por día**, Q3/Q4)") ↔ `spec.md:138` (edge case: "se retira automáticamente de la vista pública **al terminar su fecha/hora**"), `research.md:64-73` (R4-2: `event_date = hoy` **con hora de fin** → se retira al terminar su hora), `quickstart.md:285-287` (§9.3), `tasks.md` T416/T417 (los tres casos de R4-2) y T438 ("con `endTime` de hoy ya terminado desaparece solo") | `ux.md` contradice la regla de retiro de la spec y de R4-2: su "criterio es por día" dejaría visible un elemento de hoy con `endTime` ya terminado, incumpliendo FR-006/SC-009. La implementación sí está bien fijada en research/tasks/pruebas (los tres casos de R4-2); el texto de estados de `ux.md` está mal y es la referencia que sigue el `dev-frontend`. | Corregir la fila de `ux.md` §4.1 con la regla exacta de R4-2: "hoy sin hora de fin → visible todo el día; hoy con hora de fin ya terminado → fuera; mañana en adelante → visible". La prueba de T432/T433 no debe replicar el "criterio es por día". |
| **I3** | IMPORTANTE | `ux.md:265` (§4.6: «Al eliminarla, esos elementos **quedarán sin categoría**», D-16), `ux.md:419` (§7.1 texto «…Al eliminarla quedarán sin categoría. ¿Quieres continuar?»), `ux.md:503` (D-16) ↔ decisión firme `409` en `plan.md:134,543-544`, `tasks.md:46-48` y T437 ("el diálogo lo explica y **no ofrece desligarla**"), `contracts/openapi.yaml:749-759` (deleteCategoria → `409`), `quickstart.md:200-202` | `ux.md` describe el borrado de categoría en uso **desligándola** de los elementos, pero la decisión firme de la puerta 2 es `409 conflict` (FK `RESTRICT`): la UI debe decir "quítala primero" y no ofrecer continuar. El plan lo lista como "pendiente menor" y T437 afirma "ux.md D-16 alineado a esto", pero **`ux.md` no quedó alineado** y su §7.1 es la fuente de textos que copian T435–T437 ("los textos siguen `ux.md`"): un dev que copie §7.1 prometería un comportamiento que la API rechaza. | Anotar `ux.md` §4.6/§7.1/D-16 como superados por la decisión firme (texto único: «Está en uso por N elementos: quítala primero») o dejar en `tasks.md` constancia explícita de que esos renglones de `ux.md` están superados. |
| **I4** | IMPORTANTE | `ux.md:61-62` (§2.3 paso 1: "dos accesos nuevos en `AccesosRapidos`"), `ux.md:221-225` (§4.2: tres tarjetas «Gestión de eventos»/«Gestión de actividades»/«Categorías» filtradas por permiso), `ux.md:533` (§11 traza US2 → "InicioPage accesos") ↔ `plan.md:250-256` (solo edita `router.tsx` y `layout.tsx`) y `tasks.md` T430:764-784 (archivos: `lib/permissions.ts`, `features/roles/permissions.ts`, `router.tsx`, `layout.tsx`) — ninguna tarea toca `frontend/src/features/panel/components/AccesosRapidos.tsx` ni `InicioPage.tsx` | La UI de accesos rápidos del inicio del panel (tres tarjetas nuevas filtradas por permiso) que `ux.md` promete en el camino feliz de US2 **no tiene tarea asignada**: T430 solo cubre menú, rutas y guards. Los FR se siguen cumpliendo (autorización en servidor, rutas y menú), pero el flujo de `ux.md` §2.3 quedaría incompleto y QA lo vería como hueco. | Ampliar el alcance de T430 con `features/panel/components/AccesosRapidos.tsx` (+ test de que cada tarjeta se muestra solo con su permiso) y añadir los archivos a su lista; o declarar explícitamente fuera de alcance las tarjetas de `InicioPage` en `tasks.md` y ajustar `ux.md` §4.2. |
| **I5** | IMPORTANTE | `plan.md:449` (RG4-8: "El service acota con una constante de seguridad (**200 por bloque**)"), `plan.md:461-463` (Complexity 2: "se acota con una constante de seguridad del service (200 por bloque, RG4-8) **y el contrato lo documenta**") ↔ `contracts/openapi.yaml:168-172` (`getProximamente`: "sin filtros ni paginación"; **sin ningún tope documentado**) y `tasks.md` T417:518-521 (constantes: límites de campos, ≤50 fechas, ≤100 categorías — **sin el tope 200/bloque**) | La única mitigación del riesgo RG4-8 (respuestas grandes en móvil) no está en ninguna tarea ni en el contrato: la constante no existe en el plan de implementación y la afirmación de Complexity 2 ("el contrato lo documenta") es falsa frente al delta inmutable. Si nadie lo advierte, el tope se construiría por improvisión o no se construiría, y el plan quedaría desmentido por sus propios entregables. | Añadir la constante "≤ 200 elementos por bloque" a la lista de T417 (con su prueba) y documentarla en la prosa del `GET /api/v1/proximamente` al fusionar en T406 (o, si el equipo prefiere el "sin límite artificial" literal de la spec, retirar el tope de RG4-8/Complexity 2 y dejar constancia de que la revisión de volumen queda como default revisable). |

### MENORES (estilo / documentación / precisión)

| ID | Severidad | Dónde | Qué | Recomendación |
|---|---|---|---|---|
| **M1** | MENOR | `ux.md:203,444,500` (D-13 y clave `events.empty.hidden`: «Ahora mismo no hay nada publicado en esta sección. Vuelve pronto.») ↔ `tasks.md:79-80` y T433 («Por ahora no hay eventos ni actividades», copia del humano; "si su copia difiere, gana la del humano") | La copia del aviso de sección vacía difiere entre `ux.md` y la decisión del humano. Está arbitrado en `tasks.md`, pero T432 pide "añadir las claves tipadas **nuevas de `ux.md`**" sin precisar que el valor de `events.empty.hidden` es el de la copia del humano. | Fijar en T432/T433 que el valor de `events.empty.hidden` (y el de `HiddenNotice`) es «Por ahora no hay eventos ni actividades» / equivalente en inglés; anotar la copia superada en `ux.md` §7.2. |
| **M2** | MENOR | `plan.md:415` (FR-024: "…· e2e de teclado (SC-008)") y `tasks.md:1043` (misma cita) ↔ tareas: T438 (925-940) no incluye recorrido por teclado y T440 solo es checklist manual de `quickstart` §10 | Las tablas de cobertura prometen un "e2e de teclado" que ninguna tarea implementa (el precedente existe: `portada-accesibilidad.spec.ts` de F3). | O añadir el recorrido por teclado/foco a T438 (o a una spec `agenda-accesibilidad.spec.ts`), o quitar "e2e de teclado" de las filas de FR-024/SC-008 y dejar la verificación manual explícita. |
| **M3** | MENOR | `contracts/openapi.yaml:1113-1117` (`EventoAdmin.nextDate` **opcional**) ↔ `plan.md:397` y `tasks.md` T424:663-666 ("el panel marca `isPast`/`nextDate: null`") | El plan trata `nextDate: null` como señal siempre presente ("falta cargar la próxima fecha"), pero el contrato lo deja opcional: el cliente no distinguiría "sin próxima fecha" (`null`) de "campo ausente". | Añadir `nextDate` a `required` de `EventoAdmin` en la fusión de T406 (nullable) y exigirlo en la prueba de T424. |
| **M4** | MENOR | `ux.md:287` (§4.9: la auditoría muestra «Creó/Modificó/Publicó/Retiró/Eliminó el {evento/actividad} "…"») ↔ `tasks.md` T439:952-954 (verifica "los **códigos** `event.*`/`activity.*`/`category.*`") y `frontend/src/features/auditoria/messages.ts:22-37` (`ACTION_LABELS` solo etiqueta los 8 códigos de F2; los `home.*` de F3 ya salen como código crudo) | `ux.md` promete textos amigables en el módulo de auditoría que la UI actual no produce para las 15 acciones nuevas (y ya hoy no produce los de F3), y ninguna tarea extiende `ACTION_LABELS`. T439 ya acepta el código crudo, así que la contradicción está arbitrada; queda como deuda de UX. | Extender `ACTION_LABELS` en una tarea frontend (p. ej. dentro de T430: `features/auditoria/messages.ts` + test) para los 15 códigos, o corregir `ux.md` §4.9 para aceptar el código y anotar la deuda heredada de F3. |
| **M5** | MENOR | `spec.md:7` ("Status: … **pendiente de re-aprobación humana**") ↔ `estado.md:28,51` (el humano re-aprobó la spec: "Re-apruebo la spec") y `tasks.md:4` ("APROBADA v2") | La línea de estado del `spec.md` congelado quedó desactualizada respecto al registro vivo de aprobaciones; puede confundir en fases posteriores. | No editar la spec (congelada): dejar la constancia aquí y en `estado.md` de que la re-aprobación quedó registrada el 2026-10-11. |
| **M6** | MENOR | `spec.md:198` (FR-013: "Cuando **no haya ningún evento ni actividad publicados**, la sección DEBE ocultarse por completo") ↔ implementación "publicado **y próximo**" (`plan.md:424-426` R4-12, `ux.md:82`, `tasks.md` T434) | La condición literal de la spec ("nada publicado") es más estrecha que la implementación ("nada publicado y próximo"): con todo publicado pero vencido, la spec literal diría "no ocultar" pero la sección quedaría vacía (lo prohíbe la misma FR-013 y SC-012). La lectura del plan es la única coherente. | Sin cambio de código: dejar constancia de la interpretación (ya implícita en R4-12/`ux.md` §2.5) para que la validación de SC-012 la exija tal cual. |

---

## Lo que sí está coherente (comprobado)

- **Cobertura US/FR ↔ tareas**: US1–US6 y FR-001…FR-024 tienen fila en `plan.md` ("Cobertura de
  requisitos", 379-415) y en `tasks.md` (1010-1043) con verificación asociada; **sin requisito sin
  tarea ni tarea sin requisito**. Las 17 operaciones del contrato tienen implementador (1 pública +
  12 de eventos/actividades + 4 de categorías = T423–T428). Los edge cases de la spec sin FR propio
  están recogidos en el bloque final de `tasks.md` (1045-1054).
- **Rutas**: idénticas en los ocho documentos — SPA `/eventos` y `/panel/{eventos,actividades,
  categorias}`; API `/api/v1/proximamente`, `/api/v1/media/{fileName}`,
  `/api/v1/admin/{eventos,actividades,categorias}`. Sin variantes (`/eventos-actividades` solo
  aparece como alternativa descartada).
- **Permiso `categorias` de punta a punta**: migración `000009` (T403, `INSERT` sin tocar `000002`)
  → init de F2 lo otorga al administrador (T405; verificado que `Initialize` lee el catálogo
  dinámico) → autorización `AdminChain("categorias")` en servidor (T426/T428; `AdminChain(module,
  deps)` existe en `platform/middleware/chain.go:51`, sin helpers nuevos) → visibilidad en panel
  (T430: `CATEGORIAS` en `lib/permissions.ts`, `RequirePermission`, menú filtrado, sello «Disponible
  más adelante» retirado en `features/roles/permissions.ts`) → contrato (T406: prosa 9 → 10 y las
  operaciones de categorías citan el permiso; delta 642-679) → e2e de cruce (T439) → quickstart
  §2/§5/§11 → SC-005/SC-011. Los tres permisos se declaran independientes en los mismos términos en
  spec/plan/research/tasks/contrato/ux.
- **Excepción aprobada de F2 (5 aserciones + fixture)**: coherente entre plan (RG4-10, Complexity 4,
  Propuesta 4, tabla de pruebas: "3 señaladas + 2 detectadas") y `tasks.md` (T405: "las 5
  aserciones"; cabecera de Fase 2, T440, T441 y nota de cierre), marcada como **excepción** en todos
  los sitios, y **verificada contra el código**: `repository_roles_integration_test.go:52`,
  `service_init_integration_test.go:57`, `service_init_test.go:199`, `service_roles_integration_test.go:69`,
  `service_roles_test.go:464`, fixture `sampleCatalog()` en `service_init_test.go:145` y comentario
  `service_init.go:66` existen y fijan el catálogo en 9, tal como citan plan y tareas. Solo cambia
  el recuento (ninguna aserción de comportamiento se relaja).
- **Constitución**: capas `handler → service → repository` intactas; delta OpenAPI redactado e
  inmutable **antes** del código (fusión T406 antes de repository/service/handler); test-first
  exigido tarea a tarea con capas de prueba por nivel; migraciones versionadas con `down` completos
  (el `destructivo a propósito` de `000008`/`000009` sigue el precedente de `000006`); SQL
  parametrizado vía sqlc (la regla "próximo" viaja como `sqlc.narg`, nunca concatenada); validación
  de toda entrada en backend (`details` por campo, sin guardado parcial); accesibilidad FR-024 con
  checklist WCAG 2.1 AA en `quickstart` §10; 0 dependencias nuevas.
- **Colisiones con F2/F3/F10**: ninguna tarea rompe lo existente. Verificado en el código: las
  migraciones terminan en `000006` y `000002` ya siembra `eventos`/`actividades`
  (`000002_create_roles_and_permissions.up.sql:21-22`); `admin_actions_action_check` existe con
  nombre propio desde `000006`; `IsHomeFilePublished`/`handler_media.go` viven hoy en `portada`
  (`repository_home.go:219-223`, `home.sql:288`) y T415 los **mueve** sin cambiar comportamiento
  (con la regresión `portada-imagenes` como evidencia); `ImageUploader`/`PublishControls` existen en
  `features/informacion/components/` y T431 los mueve con sus pruebas; `useSiteLinks` ya está
  preparado para rutas propias (docstring `useSiteLinks.ts:30-32`) y T434 conserva la única lista
  (prueba congelada de `MobileMenu`); `.env.example` no tiene `APP_TZ` (T410 lo añade sin tocar el
  kit); CI, Makefile y docker-compose sin cambios; credenciales del quickstart §2 coinciden con
  `frontend/e2e/helpers.ts`.
- **Riesgos RG4-1…RG4-7, RG4-9 y RG4-10** tienen tarea o prueba asociada (RG4-8 es la excepción: I5).
- **Retiro automático / `endTime`**: la regla R4-2 (tres casos) es idéntica en research, data-model,
  contrato, quickstart §9 y tareas (T416/T417/T438), y `endTime` exige `startTime` está declarado de
  forma uniforme en contrato (`OcurrenciaPublica`/`FechaInput`), quickstart §3 (tabla de errores),
  T417/T435 y la migración (`CHECK (end_time IS NULL OR start_time IS NOT NULL)`).

---

## Verdades incómodas / riesgos residuales

1. **`ux.md` es el artefacto más débil de la fase.** Fue redactado antes de la puerta 2 y solo se
   realineó parcialmente: quedan **cuatro** puntos divergentes (D-13 copia del aviso — M1; D-16
   borrado de categoría — I3; D-22 zona horaria — I1; "el criterio es por día" — I2), y el plan solo
   registró dos de ellos como "pendiente menor". Antes de ejecutar Fase 8 conviene un pase de
   alineación de `ux.md` (o una sección explícita en `tasks.md` que declare qué renglones quedan
   superados por decisiones firmes); si no, el riesgo no es de arquitectura sino de **copiar texto
   viejo**.
2. **La zona horaria es un campo minado sin arbitraje escrito** (I1): cuatro documentos dicen "no
   convertir" y `ux.md` apunta al helper que convertiría. Nadie lo había señalado (ni el "pendiente
   menor" del plan). Si I1 no se corrige, el bug es silencioso y sistemático (un día menos en toda
   fecha mostrada desde zonas UTC−x) y solo lo delataría una prueba con el string exacto.
3. **Dos promesas sin dueño** (I5, I4): la constante "200 por bloque" que el plan dice que "el
   contrato documenta" (no es cierto) y las accesos rápidos de `InicioPage`. Ambas se caerían
   silenciosamente de la implementación sin que ninguna prueba fallara.
4. **La excepción sobre pruebas de F2 es legítima pero será examinada** (RG4-10): está perfectamente
   acotada (5 aserciones de recuento + fixture + comentarios; verificado línea a línea), pero el PR
   debe presentarla como la primera entrada y con el diff a la vista, porque toca una funcionalidad
   cerrada. El riesgo real no es el diff sino la percepción: sin la frase "cambio de producto
   autorizado, no debilitamiento" alguien podría leerlo como aflojar pruebas.
5. **F3 ya arrastra una deuda que F4 hereda**: el módulo de auditoría muestra los códigos `home.*`
   sin etiquetar (M4). Si F4 añade 15 códigos más sin etiquetar, la tabla de auditoría del panel
   será cada vez menos legible para el equipo no técnico (el público objetivo del panel). Merece una
   decisión explícita, no arrastrarse.
6. **Límites "revisables" por todas partes** (≤50 fechas, ≤2000 caracteres, ≤100 categorías, 200 por
   bloque, WCAG AA de referencia): son defaults que la spec deja al humano. No bloquean, pero el
   `documentador` debería listarlos en la entrega para que sean revisables de verdad.
