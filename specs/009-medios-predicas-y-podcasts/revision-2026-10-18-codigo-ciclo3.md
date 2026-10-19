# Revalidación de código (ciclo 3) — F9 Medios: prédicas y podcasts

**Fecha**: 2026-10-18 · **Fase**: 7 (validar, ciclo de corrección 2 → revalidación) · **Revisor**: `revisor-codigo` (solo lectura; este informe es la única escritura)
**Rama**: `009-medios-predicas-y-podcasts` · **Alcance**: `git diff 2556f00..32a9619` (1 commit: `32a9619` «fix(medios): permitir la union automatica segura con una ficha de par resuelto», 7 archivos, +339/−52, todos de backend) contrastado contra `spec.md` v2 (FR-013, FR-014, FR-015, FR-017, SC-003, SC-007, A3), `data-model.md` (invariante de la pareja recíproca, C1/default #6), el informe anterior (`revision-2026-10-18-codigo-ciclo2.md`, hallazgos N1–N3) y las garantías de los ciclos previos.

## Veredicto

**APRUEBA CON CAMBIOS** — **N1 queda confirmado como resuelto**: la unión automática segura (FR-013) vuelve a funcionar contra una ficha de par **resuelto** (una ficha con los dos enlaces, puntero y resolución intactos) y la exclusión de fichas con pareja sobrevive **solo** en la rama dudosa, con la invariante M1 doblemente protegida (guardia en `candidato()` + guarda `duplicate_of_id IS NULL` en la SQL). **N2 y N3 quedan cerrados correctamente** —N3 incluso mejor que la propuesta: no puede borrar la carátula que el sitio está sirviendo—, con pruebas que fijan comportamiento y fallan contra el código anterior. **Cero regresiones** en las garantías de la spec. Los cambios que faltan **no son de este commit** sino de un defecto latente que este ciclo descubrió (**N4, IMPORTANTE, preexistente**): el doble `fakeRepository` no es seguro para concurrencia y `go test -race ./internal/medios/...` —la propia puerta del CI (`.github/workflows/ci.yml:112`)— **falla de forma determinista** (5/5 corridas, y también 5/5 en el commit padre: no es regresión de `32a9619`). Con el mutex del doble —diez líneas— el cambio es mergeable. Quedan además una micro-ventana residual de N2 (SUGERENCIA) y el **borde de spec** del tercer elemento dudoso, que en mi opinión **exige decisión del humano** (ver sección propia).

## Evidencia verificada en esta revalidación

| Comprobación | Resultado |
|---|---|
| `gofmt -l .`, `go vet ./internal/medios/...`, `go build ./...` (backend) | limpios |
| `go test -count=1 ./internal/medios/...` (unitarias) | en verde (0,4 s) |
| `go test -count=1 -tags integration ./internal/medios/...` con **PostgreSQL real** (`DATABASE_URL_TEST` → BD `app_test` del compose, migraciones al día) | **en verde — EJECUTADAS de verdad** (11,2 s) |
| Pruebas nuevas del commit, con `-v` | **todas se ejecutan y pasan** (sin `skip`): `TestBuscarNuevosParejaReciprocaUnica` (3 subtests, incl. «un match seguro con una ficha emparejada se une»), `TestBuscarNuevosNoRevierteUnaResolucionDeMarca`, `TestBuscarNuevosNoDejaCaratulasHuerfanas` (2 subtests), `TestIntegrationMarcaDeDisponibilidadNoSeDevuelve`, más las de regresión (`PrecedenciaA3`, `NoPisaLaCaratula`, `NoPisaLoDelEquipo`, `Merge…DosEnlaces`, `EsIdempotente`, `SyncWritesOnlyItsColumns`) |
| **`go test -count=1 -race ./internal/medios/...`** | **FALLA** — carrera de datos en el doble de pruebas (ver **N4**). Reproducida 5/5 en el commit y **5/5 en el padre** (`2556f00`, worktree temporal ya retirado): preexistente, no introducida por `32a9619` |
| `npx vitest run` sobre `features/medios`, `features/publico/medios`, `api/medios.test.ts` | **127/127 en verde** (13 archivos) |
| Pruebas desactivadas (`skip`/`only`/`xit`) | ninguna (solo un `t.Skipf` legítimo por datos de zona horaria, `service_sync_test.go:1706`) |
| Diff de `backend/go.mod`, `frontend/package.json`, `backend/migrations/`, `contracts/`, `routes.go` | **0 cambios**: sin dependencias ni migraciones nuevas, contrato 0.9.0 intacto, misma superficie HTTP |
| Frontend | **0 archivos tocados** por `32a9619` |

## Hallazgo por hallazgo

| # | Hallazgo (informe anterior) | Estado | Evidencia |
|---|---|---|---|
| **N1** | La exclusión de fichas con pareja en `candidato()` bloqueaba la unión automática segura (FR-013) y dejaba el mismo sermón visible dos veces (SC-003) | **RESUELTO y confirmado** | Ver análisis detallado abajo. |
| **N2** | Micro-ventana residual: el sync devolvía a la fila las marcas que leyó | **RESUELTO** (con una micro-ventana aún más estrecha: ver S1) | Ver análisis detallado abajo. |
| **N3** | La descarga que no se usaba en la carrera de la carátula quedaba huérfana | **RESUELTO (mejor que lo propuesto)** | Ver análisis detallado abajo. |

### N1 — Confirmado: la unión segura funciona con par resuelto y M1 sigue protegida

**La unión automática segura vuelve a funcionar con una ficha de par resuelto.** `candidato()` (`service_sync.go:722-758`) ya no descarta por `DuplicadoDeID` antes de comparar: el descarte general quedó en `!encontrada || ficha.UnidoAID != nil` (`:736`) y el match **seguro** (título normalizado **y** día) devuelve la ficha aunque esté emparejada (`:742-746`). La unión es inocua para la pareja por construcción: `unirEntrada` (`:770-793`) escribe solo la fuente (`insertarFuente`, `ON CONFLICT` = idempotencia) y `UpdateMediaItemJoin`, cuya SQL escribe **únicamente** `title`/`published_at`/`origin` (`media_episodes.sql:311-321`): `duplicate_of_id`, `duplicate_mark_resolved_at`, la visibilidad y la carátula quedan intactos. **Prueba** (`service_sync_test.go:528-578`, «un match seguro con una ficha emparejada se une»): un podcast que re-publica el audio de un sermón cuyo video está en un par **resuelto** → `Unidos=1`, **2 fichas** (no 3: la prueba falla contra el código anterior, que creaba una tercera), la ficha unida con **los dos enlaces** y `origin='both'`, el puntero recíproco **en los dos sentidos** sin romperse y **las dos resoluciones del par sin reabrir**. Exactamente lo pedido en la corrección propuesta.

**La exclusión sigue vigente solo en la rama dudosa y M1 queda protegida.** El match dudoso (solo título o solo día) exige ahora candidato libre explícitamente (`service_sync.go:747`, `ficha.DuplicadoDeID == nil`), porque es la única rama que escribe el puntero recíproco con `MarkMediaItemDuplicate`. La invariante queda **doble**: la guarda del service y la de la SQL (`media_episodes.sql:346`, `AND duplicate_of_id IS NULL`, comentada como invariante de la pareja en `:337`). **Pruebas**: «un match solo por título no roba el puntero de una ficha emparejada» (`service_sync_test.go:580-619`, ahora con el caso reforzado: la ficha emparejada conserva su pareja, la fuente del elemento nuevo no se le cuela y la entrada suelta nace aparte) y el primer subcaso intacto «dos elementos del mismo feed no comparten candidato»; en integración, `TestIntegrationMarkDuplicatePairEsAtomico` sigue en verde. `marcarDudoso` (`:804-835`) no cambió.

**Matiz menor (no es hallazgo)**: con este arreglo, una ficha con par **abierto** también puede recibir una unión segura (el código no distingue abierto/resuelto en la rama segura). Es coherente: la unión no toca el eje de duplicados, el par queda para «Revisar duplicado» y la adicional sigue oculta; el caso de prueba cubierto usa par resuelto, que era el borde de FR-013.

### N2 — Cerrado correctamente (S1: micro-ventana residual aún más estrecha)

El eje de la marca viaja ahora como `narg`: `CambioDisponibilidad.MarcaNueva` (`repository_episodios.go:200-207`, nil = **no tocar el eje**) y la SQL escribe `availability_mark = COALESCE(narg, availability_mark)` con `availability_mark_resolved_at = CASE WHEN narg IS NOT NULL THEN NULL ELSE … END` (`media_episodes.sql:298-301`): una marca nueva entra **abierta** y sin decisión el eje **no se escribe**. `aplicarDisponibilidad` (`service_sync.go:404-440`) ya no parte de «devolver lo leído»: `cambio` nace sin marca y solo las ramas del `switch` la fijan. La precedencia A3 se conserva tal cual (`:415-422`: 404/410 confirmado → marca nueva abierta **y** oculta, aunque el equipo hubiera resuelto). **Pruebas**: `TestBuscarNuevosNoRevierteUnaResolucionDeMarca` (`service_sync_test.go:761-805`, con un gancho que hace resolver la marca al equipo justo entre la lectura y la escritura —falla contra el código anterior, que la borraba—) e `TestIntegrationMarcaDeDisponibilidadNoSeDevuelve` (`repository_episodios_integration_test.go:1218-1270`, PostgreSQL real: sin marca nueva la fila conserva marca y resolución; con marca nueva se escribe y entra abierta). Es exactamente la propuesta. El residuo que queda es **S1** (abajo), más estrecho que el original.

### N3 — Cerrado, y mejor que lo propuesto

En el error de `UpdateMediaItemCoverIfNull`, `copiarCaratula` (`service_sync.go:577-586`) relee la fila viva (`GetMediaItemByID`, lectura directa sin caché, `repository_episodios.go:133-142`) y retira su copia **solo si la fila confirma que no la usa** (`viva.Caratula == nil || *viva.Caratula != archivo`), delegando en `borrarCaratulasSinUso` (`service_admin.go:556-571`, mejor esfuerzo con log). Esto es **más seguro que mi propuesta original** (borrar a ciegas en el `return`): un error de respuesta de una escritura que sí se aplicó ya no puede borrar el archivo que el sitio está sirviendo —justo el riesgo que mi propuesta abría—. Si la relectura falla o no encuentra la fila, **no se borra** (dirección segura: queda un huérfano que nunca se publica, igual que antes). **Pruebas**: `TestBuscarNuevosNoDejaCaratulasHuerfanas` (`service_sync_test.go:1445-1499`) —«la escritura que no se aplicó retira su copia» (falla contra el código anterior: 0 borrados) y «un error de respuesta no borra la carátula que sí se guardó» (guardia contra el sobre-borrado: fijaría el fallo de una corrección a ciegas)—. Límite conocido y aceptable (mejor esfuerzo): el caso teórico en que el `UPDATE` se aplica **después** de la relectura (pérdida de conexión a mitad de sentencia) dejaría la fila apuntando a un archivo borrado; requiere una ventana sub-milisegundo y un fallo de red concreto, y el camino por defecto (ficha sin carátula, marcador neutro) sigue siendo el previsto por P9-7.

## Regresiones buscadas (ninguna encontrada)

| Garantía de la spec | Comprobación |
|---|---|
| Unión irreversible y atómica (FR-013/FR-019) | `RepointMediaItemSources` y `Merge` **intactos** en este commit; `TestMergeDejaUnaFichaConLosDosEnlaces` e `TestIntegrationRepointMediaItemSourcesMovesSources` en verde |
| Invariante de la pareja recíproca (M1/default #6) | Doble guarda (service `service_sync.go:747` + SQL `media_episodes.sql:346`); la rama segura no escribe el eje; pruebas citadas en N1 en verde |
| El sync no pisa lo del equipo (FR-015/SC-007) | Las escrituras del sync siguen por columnas y el eje de marca es ahora **más** estrecho; `UpdateMediaItemJoin` solo `title`/`published_at`/`origin`; `TestBuscarNuevosNoPisaLoDelEquipo`, `TestBuscarNuevosNoRevierteUnaEdicionDuranteLaBusqueda` y `TestIntegrationSyncWritesOnlyItsColumns` en verde |
| Precedencia de la indisponibilidad confirmada (A3) | Conservada sin cambios en la lógica del `switch`; `TestBuscarNuevosPrecedenciaA3` (2 subtests) en verde |
| Fail-safe del «estante de 15» (FR-017/SC-014) | `comprobar`/`peorQue`/`todasFueraDeVentana` intactos |
| Idempotencia de la búsqueda (R9-11/P9-11) | `ON CONFLICT (provider, external_id) DO NOTHING` intacto; `TestBuscarNuevosEsIdempotente` e integración en verde |
| Predicado público (FR-009/SC-002) | Consultas públicas sin tocar en el diff |
| Contrato, migraciones, dependencias, superficie HTTP | 0 cambios |
| Carátulas (SC-005, defaults #7/#10) | Elección determinista por ficha y escritura condicional intactas; `TestBuscarNuevosNoPisaLaCaratula` (2 subtests) en verde |

## Hallazgos nuevos

### IMPORTANTE

**N4 — `go test -race` falla de forma determinista: el doble `fakeRepository` no es seguro para concurrencia (PREEXISTENTE; no lo introduce `32a9619`).**
- La carrera está entre las goroutines reales de `copiarCaratulas` (`service_sync.go:514-524`, `maxEnVuelo` en vuelo): una goroutine escribe `f.episodios[i].Caratula` en `UpdateMediaItemCoverIfNull` (`service_test.go:1028-1029`) mientras otra lee el slice en `buscar()` (`service_test.go:652-653`) desde el mismo método. `fakeRepository` **no tiene mutex** (`service_test.go:565`) y modela mal al repository real, que sí es seguro (pgxpool).
- **No es regresión de este commit**: reproducida 5/5 con `go test -count=1 -race -run TestBuscarNuevosCaratulas ./internal/medios/` sobre `32a9619` **y 5/5 sobre el padre `2556f00`** (worktree temporal, ya retirado). El fallo se concentra en `TestBuscarNuevosCaratulas/el_tope_de_carátulas_acota_las_descargas`. Las pruebas sin `-race` pasan, por eso los ciclos anteriores (que no corrieron `-race`) lo vieron verde.
- **Por qué importa ahora**: la puerta del CI corre `go test -race -coverprofile=coverage.out ./...` (`.github/workflows/ci.yml:112`); en máquinas con paralelismo suficiente la carrera se manifiesta casi siempre (aquí 5/5). O el job de backend está roto o el runner lo enmascara (fallos intermitentes). Y este commit toca exactamente ese doble y esa ruta (`copiarCaratula`).
- **Corrección propuesta** (`dev-backend`, test-only, ~10 líneas): un `sync.Mutex` (o `sync.RWMutex`) en `fakeRepository` que proteja **todos** los accesos a `episodios`/`fuentes` —empezando por `buscar`, `UpdateMediaItemCoverIfNull`, `GetMediaItemByID` y `UpdateMediaItemAvailability`—. Alternativa inferior: serializar descargas en pruebas, que dejaría de probar la concurrencia real de `copiarCaratulas`.

### SUGERENCIA

**S1 — Micro-ventana residual de N2 (más estrecha que la original): la rama `todasFuera` re-escribe una marca de la misma clase aunque no cambie.**
`aplicarDisponibilidad` (`service_sync.go:423-430`) fija `MarcaNueva = FueraDelFeed` siempre que `!yaResuelta`, incluso cuando la fila **ya tiene** esa marca abierta —el valor no cambia, pero la SQL limpia `availability_mark_resolved_at` (`media_episodes.sql:299-301`)**. Si el equipo resuelve la marca justo entre la lectura (`service_sync.go:261`) y ese `UPDATE`, la resolución se pierde igual que en el N2 original (solo en este subcaso: misma clase abierta + todas las fuentes fuera de ventana; la resolución debe llegar en la ventana de milisegundos). El commit dice «una resolución hecha entre la lectura y la escritura no se revierte», que es cierto salvo aquí. **Propuesta** (una línea, service): en la rama `todasFuera`, escribir `MarcaNueva` solo cuando la clase **cambia** (`ficha.MarcaDisponibilidad == nil || *ficha.MarcaDisponibilidad != MarcaDisponibilidadFueraDelFeed`); la rama A3 debe seguir sobrescribiendo siempre (precisión aprobada). Nota: la prueba `TestBuscarNuevosNoRevierteUnaResolucionDeMarca` cubre el caso «sin decisión de marca»; este subcaso no está cubierto.

## Borde de spec: un tercer elemento dudoso contra una ficha ya emparejada

**Situación** (la que el equipo dejó señalada): un elemento cuya correspondencia es **dudosa** (solo título o solo día) contra una ficha que **ya tiene pareja** no puede entrar en un par nuevo (M1: la pareja recíproca es de dos). Hoy `candidato()` lo ignora silenciosamente y el elemento entra por `crearFicha`: **visible, suelto y sin marca** (así lo fija el subcaso `service_sync_test.go:613-616`).

**Mi opinión: exige decisión del humano** —no lo doy por defecto de implementación ni por limitación ya aceptada—, por esta colisión:
- **Por la letra de la spec es una desviación**: FR-014 es un DEBE —«entradas separadas, la adicional **oculta** y **marcadas** «posible duplicado»»— y SC-003 es criterio medible: «el 100% de los posibles duplicados que el sistema no pudo unir con seguridad **quedan marcados**». Aquí la entrada queda **visible y sin marca**, y si resulta ser el mismo sermón el visitante lo ve dos veces (SC-003 «0 repeticiones»), sin camino de revisión en el panel («Revisar duplicado» solo con par abierto).
- **Pero cerrarlo toca el modelo aprobado**: el eje de duplicados es un **par recíproco de dos** (`data-model.md`, C1/default #6, sin `CHECK` de simetría a propósito). «Marcar contra una ficha ya emparejada» exigiría par de N, marca unilateral o un segundo eje —cambios de `data-model.md`/`spec.md` que solo aprueba el humano—. Y la alternativa barata (ocultar la entrada suelta) es peor: contenido invisible **sin** camino de resolución.

**Riesgo real acotado**: el disparador es que el feed publique un elemento que difiere en título normalizado **o** día de una ficha cuyo contraparte ya está en par (p. ej. borde de día + par ya resuelto). Bajo de probabilidad, pero los feeds re-publican y el catálogo crece durante años. Mitigación manual existente: ocultar una a mano; una re-publicación posterior con título **y** día ya se une sola (N1).

**Opciones que veo para el humano**: (a) **aceptar la limitación y documentarla** —nota en `data-model.md` (borde del par de dos) y en las «Consecuencias asumidas» de `spec.md`, con la regla operativa «si aparece una entrada suelta que el equipo sospecha duplicada, ocultarla a mano»—; (b) aprobar una extensión del eje de duplicados para marcas unilaterales (cierre completo de FR-014/SC-003 en ese borde, coste: modelo + panel); (c) dejarlo como está sin documentar —lo desaconsejo—. **Mi recomendación**: (a) ahora, con (b) como mejora futura si el caso se presenta en la operación. Hasta que el humano decida, el comportamiento actual queda como **borde abierto conocido**, no bloqueante para este merge (el código es consistente con el modelo aprobado; lo que falta es la decisión de producto).

## Lo que se revisó y lo que no

**Revisado**: el diff completo de `32a9619` (SQL sqlc + generada, repository, service de sincronización y las pruebas de los dos lados); la correspondencia con N1/N2/N3; las garantías de la tabla de regresiones con el código y las pruebas delante; unitarias y de integración con PostgreSQL real **ejecutadas** (incluidas las nuevas, verificadas con `-v`); frontend (sin cambios en el commit) **ejecutado**; `go vet`, `gofmt`, `build` **ejecutados**; `-race` **ejecutado** (origen de N4, con verificación en el commit padre).

**No revisado (fuera de este informe)**: e2e de Playwright (lane de `qa-tester`); pruebas manuales post-despliegue; la documentación de entrega (`pr.md`, CHANGELOG); los módulos F1–F8; el comportamiento de la red saliente contra servicios reales.

## Cierre

| Ítem | Estado |
|---|---|
| N1 (MEDIO) | **RESUELTO y confirmado** — unión segura con par resuelto operativa; M1 protegida (guardia + guarda) |
| N2 (SUGERENCIA) | **RESUELTO** — micro-residuo S1 (SUGERENCIA, una línea) |
| N3 (SUGERENCIA) | **RESUELTO** — y más seguro que lo propuesto |
| N4 (IMPORTANTE) | **ABIERTO** — carrera en el doble de pruebas (preexistente): mutex en `fakeRepository` |
| S1 (SUGERENCIA) | **ABIERTA** — no re-escribir la marca cuando su clase no cambia |
| Borde de spec (tercer elemento dudoso) | **ABIERTO — decisión del humano** (recomendación: limitación aceptable documentada en `data-model.md` + `spec.md`) |

**Veredicto final: APRUEBA CON CAMBIOS** — cambios: (1) el mutex de N4 (IMPORTANTE, preexistente, test-only, ~10 líneas; es lo único que puede mantener rojo el gate `-race` del CI), (2) opcional pero barato: S1 y la documentación del borde según decida el humano. `32a9619` en sí está correcto, cierra N1/N2/N3 con pruebas de comportamiento y sin regresiones.
