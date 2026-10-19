# PR — F9 Medios: prédicas y podcasts

> Descripción para el Pull Request de la rama `009-medios-predicas-y-podcasts` (preparada por
> `documentador` el 2026-10-18 y **actualizada el 2026-10-19 en el cierre documental**, tras la
> fase 7 de validación; T945). **El merge y el despliegue requieren aprobación humana explícita**
> (fase 9/9 del flujo `equipo-feature`; puertas «PR / merge» y «Despliegue» en `estado.md`). Esta
> sesión del `documentador` no tenía shell ni `gh`: quien abre el PR es `devops`/el humano con esta
> descripción.

## feat: F9 — Medios: prédicas y podcasts (sección `/medios`, importación por feeds y bloque en la portada)

### Resumen

La cuarta sección de contenido dinámico del sitio: **las prédicas y podcasts de la iglesia**, visibles
sin cuenta, con el módulo del panel que **los importa desde los dos feeds con un botón** y su bloque
**al final** de la portada. Cubre las **47/47 tareas** de [`tasks.md`](./tasks.md) (T901–T946 +
T907A; US1–US6 / FR-001…FR-030 / SC-001…SC-017 de la **spec v2**) y cierra la fase 7 con **tres
rondas de validación** (QA, código y seguridad) y **14 commits de corrección**, **sin hallazgos
bloqueantes abiertos**.

**Modelo (decisión 27 del roadmap): importación semiautomática + enlaces que salen del sitio.**
El contenido **se trae de los feeds** —el Atom del canal de YouTube y el RSS del podcast— al pulsar
**«Buscar episodios nuevos»**; el sitio **no reproduce nada dentro de sí ni pide nada a terceros**
(0 recursos de YouTube/Spotify/Anchor al abrir la sección o la portada) y **enlaza** a cada servicio
**en pestaña nueva anunciada**. **Sin cron, worker, cola ni despliegue previo.**

**Veredicto de la fase 7**: `qa-tester` **APRUEBA**, `seguridad` **APRUEBA** y `revisor-codigo`
**APRUEBA CON CAMBIOS con todos sus hallazgos resueltos y confirmados** (detalles abajo).

### Qué se construyó

**Backend (Go + PostgreSQL, módulo `internal/medios`)**
- **7 endpoints** (contrato **0.8.0 → 0.9.0** en [`backend/api/openapi.yaml`](backend/api/openapi.yaml),
  fusión aditiva del delta `contracts/openapi.yaml`, que no se edita; `make api-gen` **sin deriva**):
  públicos **`GET /api/v1/medios?lang=es|en`** (tanda de 12 con «Cargar más», `no-store`, `lang`
  cerrado) y **`GET /api/v1/portada/medios?lang=es|en`** (tope **5 resuelto en el servidor**); de panel
  `/api/v1/admin/medios` (lista con filtros de vista), `POST …/buscar-nuevos`, `GET`/`PATCH …/{id}`,
  `POST …/{id}/unir` y `POST …/{id}/marcas/resolver`, tras `AdminChain("medios")` + `RequirePermission`.
- **Migraciones `000014_create_media_content`** (tablas **`media_items`**, **`media_episode_sources`**,
  **`media_sync_runs`**) y **`000015_extend_admin_actions_for_media`** (los 6 códigos del `CHECK` de
  `admin_actions`), ambas con `up`/`down` y verificadas **up → down → up**; consultas sqlc nuevas
  `media_episodes.sql`, `media_episode_sources.sql` y `media_sync_runs.sql` (`sqlc-verify` sin deriva).
- **Sin permisos nuevos**: F9 **activa** el permiso **«Medios (prédicas y podcasts)»** (`medios`),
  reservado por F2; **el catálogo sigue en 10** y el nombre **no cambia**.
- **Auditoría**: **6 códigos `media.*`** —`sync`, `description.update`, `hide`, `show`, `merge`,
  `mark.resolve`—, registro **atómico** con la mutación, **con rechazos y denegaciones**.
- **Unión por título normalizado + mismo día** en `America/Costa_Rica` **sin margen** (lo dudoso →
  entradas separadas, la adicional oculta, ambas marcadas «posible duplicado»); **idempotencia** de la
  búsqueda (0 reimportaciones, 0 borrados, 0 cambios de visibilidad/descripción/uniones) y
  **resumen** con entraron/unieron/marcaron/**feed fallido**.
- **Disponibilidad a prueba de fallos (FR-017, regla del «estante de 15»)**: solo **404/410
  confirmados** ocultan la ficha completa; el que sale del feed pero sigue vivo queda **visible** con
  la marca «ya no aparece en el feed»; lo inconcluso **nunca oculta**; una marca resuelta **no se
  reabre por la misma causa**.
- **Fábrica única de cliente HTTP saliente** (corrección A1 de `analyze.md`) con las cuatro barreras
  —solo `https` en allowlist, sin IPs privadas/metadata, sin redirecciones fuera, `io.LimitReader` y
  firma binaria en imágenes— aplicada a los **tres consumidores** (feeds, disponibilidad, carátulas),
  con la misma tabla de pruebas sobre los tres.
- **Carátulas copiadas al sitio** (descarga acotada y validada; 0 dependencias de terceros en el
  navegador del visitante). **0 dependencias nuevas** (`golang.org/x/text` pasa de indirecta a
  directa, **misma versión**; `go.mod` y `package.json` sin altas).

**Frontend (React + TypeScript + Vite)**
- **Sección pública `/medios`** (`MediosPage` en `features/publico/medios/`) con fichas **carátula,
  nombre, fecha y descripción corta** —**la descripción vacía no se pinta** (decisión 28)—, orden del
  más reciente al más antiguo, **«Cargar más»** (tanda de 12), enlaces externos **anunciados**
  (servicio, salida del sitio y pestaña nueva) y **marcador neutro decorativo** cuando no hay carátula.
  **Sin nada visible → redirección silenciosa a la portada** y entrada de navegación oculta.
- **Bloque de la portada al final, después del de ofrendas y diezmos** (F7): exactamente los
  **últimos 5** visibles, **«Ver todo»** → `/medios`, oculto por completo con 0 visibles
  (`GET /api/v1/portada/medios`).
- **Panel `/panel/medios`** (entrada «Gestión de medios» en Inicio): zona **«Buscar episodios
  nuevos»** con botón desactivado durante la búsqueda y **resumen** del resultado (con aviso propio
  si el tiempo se agota), listado con estado/origen/formatos y **filtro de vista** «Todo / Solo lo
  pendiente», **ocultar/mostrar**, **unir duplicados** con vista previa y doble confirmación,
  **resolver marcas** («Mostrar de nuevo» con confirmación) y editor de **descripción corta es/en**;
  `RequirePermission(MEDIOS)` en las rutas y las **6 etiquetas de auditoría** `media.*`.
- **i18n `es`/`en`** con respaldo al español (el contenido llega en español; solo la descripción
  corta admite inglés a mano) y **sistema de F10** aplicado: tema claro/oscuro, `prefers-reduced-motion`,
  foco visible, 320 px sin desplazamiento horizontal.
- **Vitest 946 pruebas / 131 archivos** en verde al cierre (930/129 en T944 + 16 de las correcciones),
  con `lint` y `typecheck` limpios.

**Infraestructura / Docker (devops)**
- **Sin cambios**: `docker-compose.yml`, Dockerfiles, `.github/`, `Makefile`/`proyecto.mk`, kit
  compartido y **`.env.example` sin cambios** (F9 no introduce ninguna variable ni secreto: los dos
  feeds se leen sin claves).

### Validación del equipo (fase 7, 2026-10-18/19 — tres rondas, 14 commits de corrección)

**Ronda 1 (2026-10-18)**
- **`qa-tester` — APRUEBA** ([informe](./revision-2026-10-18-qa.md)): verificación **independiente en
  vivo** contra la API real —búsqueda contra los feeds reales a demanda e idempotente, feed caído sin
  ocultar ni marcar nada, público sin filtraciones, bloque de la portada con tope 5 en el servidor,
  `401`/`403` auditados (`denied`) y catálogo de permisos intacto en 10—; suites propias en verde
  (Go, integración, Vitest **930/930**, e2e de F9 **25/25** con `--workers=1`).
- **`seguridad` — APRUEBA** ([informe](./revision-2026-10-18-seguridad.md)): sin bloqueantes, altos
  ni medios; **3 bajos** (B1: `offset` sin tope en la ruta pública daba `500` con un valor enorme;
  B2 y B3 de defensa en profundidad/falso positivo) y 6 informativos. El punto sensible —la primera
  red saliente del backend— **está cerrado por construcción** (la fábrica única es el único punto que
  construye un `http.Client`).
- **`revisor-codigo` — APRUEBA CON CAMBIOS** ([informe](./revision-2026-10-18-codigo.md)): **3 ALTOS**
  —**H1**: «Buscar episodios nuevos» se cortaba a los **5 s** del cliente con el servidor teniendo
  **90 s** de presupuesto (y el abort cancelaba el trabajo del servidor); **H2**: la sincronización
  **pisaba la carátula** al unir y, en la primera importación, el ganador era **aleatorio**; **H3**:
  el alta dejaba una **ficha huérfana sin enlaces** en la carrera entre dos búsquedas—, **3 MEDIOS**
  (**M1**: huecos en la pareja recíproca de duplicados; **M2**: `ResolveMark` del par no atómico;
  **M3**: escritura de fila completa que podía revertir una edición del equipo) y **3 BAJOS**
  (**B1**: el límite de 300 caracteres se contaba en UTF-16; **B2**: el fallo de un `PATCH` de solo
  visibilidad se auditaba mal; **B3**: el panel no avisaba del truncado y la unión dependía de la
  página cargada).

**Ciclo 1 (11 commits)** — se corrigieron **los 3 altos, los 3 medios y los bajos**: backend (8) —
la carátula **ya no se pisa** al unir y se encola **una sola por ficha** prefiriendo la del video; el
**alta es atómica** (una transacción ficha + fuente que revierte al conflictuar); cerrados los
**huecos de la pareja recíproca**; **`ResolveMark` del par en una transacción**; la búsqueda
**escribe solo sus columnas** (ya no puede revertir una edición del equipo); el **`offset` público
queda acotado**; la **red saliente exige puerto 443**; y el **código de auditoría** del `PATCH` de
visibilidad más la **procedencia por lote** (sin N+1)— y frontend (3) —**tiempo de espera propio de
120 s** para la búsqueda (antes se cortaba a los 5 s y el equipo no veía el resumen), con mensaje
claro si se agota; el **límite de 300 cuenta runas**; y el panel **avisa del truncado** y la unión
**busca la pareja contra el servidor**—. Cada corrección con su prueba **en rojo antes del arreglo**.

**Ronda 2** — **`qa-tester` APRUEBA** ([informe](./revision-2026-10-18-qa-ciclo2.md); Vitest
**946/946**, e2e **25/25**; corrigió además una carrera del helper e2e `portadaContiene` —
infraestructura de pruebas, sin debilitar aserciones—) y **`seguridad` APRUEBA**
([informe](./revision-2026-10-18-seguridad-ciclo2.md); sus 3 bajos **cerrados y verificados en vivo**,
sin nada nuevo). **`revisor-codigo` — APRUEBA CON CAMBIOS** ([informe](./revision-2026-10-18-codigo-ciclo2.md)):
los 9 hallazgos del ciclo anterior **resueltos**, y aparece **un hallazgo nuevo MEDIO (N1)**: la
corrección de la pareja recíproca (M1) **bloqueaba la unión automática segura** con una ficha de par
resuelto, de modo que el mismo sermón podía verse **dos veces**; más dos sugerencias menores (**N2**:
micro-ventana de la marca de disponibilidad; **N3**: carátula huérfana en el almacén).

**Ciclo 2 (3 commits)** — **N1 corregido** (`32a9619`: la exclusión de la pareja queda **solo en la
rama dudosa**; la unión segura vuelve a funcionar con par resuelto) y de paso **N2** (una resolución
de marca **ya no puede barrerse**) y **N3** (la carátula huérfana **se retira solo si no está en uso**).

**Ronda 3** ([informe](./revision-2026-10-18-codigo-ciclo3.md)) — **N1/N2/N3 confirmados resueltos y
sin regresiones** en ninguna garantía de la spec. Aparecen **N4** (IMPORTANTE, **preexistente**): el
doble de pruebas `fakeRepository` **no soportaba concurrencia** y **`go test -race` fallaba de forma
determinista**; como **el CI ejecuta `-race`** (puerta en `.github/workflows/ci.yml`), el PR habría
salido **en rojo**. Corregido (`b69571d`: mutex en el doble, **sin relajar aserciones**) y **S1**
(`7dbd8b2`: no re-escribir una marca de la misma clase). Hoy **`go test -race ./...` está en verde
en los 22 paquetes del backend**.

**Total: 14 commits de corrección.** **Veredicto final**: **QA aprueba, seguridad aprueba y código
aprueba con todos sus hallazgos resueltos y confirmados.**

### Decisiones de producto registradas

- **Decisión 26 (roadmap §6, 2026-10-17)**: F9 se construye **antes que F5 y F8** (depende solo de
  F2 y F3, ambas terminadas) y no necesita permisos nuevos; el cuestionario se le entrega al cliente
  **antes** de aprobar la spec. *(Su punto de la Opción B quedó corregido por la 27.)*
- **Decisión 27 (2026-10-17)**: tras las respuestas del cliente al cuestionario
  `docs/producto/cuestionario-F9-medios.md`, **giro de alcance** —importación desde los **feeds**
  con botón **a demanda** y **enlaces que salen del sitio** (Q15: «No» a contenido de terceros)—;
  con ello el panel pasa a **buscar/revisar/ocultar/unir/describir**, el estado pasa de
  «borrador/publicado» a **«visible/oculto»**, **sin borrado** y con **una entrada por sermón** (Q18).
  Queda registrado que **F9 no cumple la promesa de `idea.md` §1** de ver las grabaciones «sin
  necesidad de tener instaladas esas aplicaciones»: es una decisión consciente del cliente.
- **Decisión 28 (2026-10-18, T946)**: la **descripción corta puede estar vacía** (feed o panel) y
  entonces la ficha **omite esa línea**; se retiró el `minLength: 1` del delta y del contrato vivo
  (sigue en **0.9.0**).
- **Cinco precisiones de la puerta 2** («Aplica las 5 y apruebo el plan», 2026-10-17), firmadas en
  `spec.md` y anotadas en `plan.md`: (1) **solo un «posible duplicado» sin resolver** mantiene la
  ficha fuera del público —una marcada «ya no aparece en el feed» **sigue visible**—; (2) un enlace
  **confirmado como no disponible oculta la ficha completa**, también la que junta video y audio;
  (3) un episodio **sin descripción** se importa igual y la ficha **omite** esa línea; (4) el
  **«mismo día»** de la unión se calcula en `America/Costa_Rica` **sin margen** y la duda se **marca**
  en vez de unir; (5) al unir **manda el nombre del feed del video** (default #7).
- **Ajustes aprobados por el humano antes de la puerta 1**: regla del **«estante de 15»**
  con comprobación de disponibilidad (FR-017) y **apertura de los enlaces externos en pestaña nueva
  anunciada** (FR-004/FR-005).
- **Aprobaciones humanas**: spec **v2** («Apruebo la spec v2») y plan («Aplica las 5 y apruebo el
  plan»), ambos el 2026-10-17; decisión 28 el 2026-10-18. **No queda nada pendiente de re-aprobación.**

### Desviaciones y cambios acotados (registrados, no debilitamiento)

| ID | Qué es | Por qué |
|---|---|---|
| **Unión de duplicados** (P9-10 / R9-12) | La unión **no borra** la fila: la ficha absorbida queda con puntero (`merged_into_id`), se le **reparten sus fuentes** a la resultante y sale de todas las consultas (**0 `DELETE`**), todo **en una transacción** | Decisión fijada en `plan.md`/`research.md` antes de implementar; conserva el identificador para que la siguiente búsqueda **no reimporte** lo unido. Requiere que el predicado público excluya absorbidas (probado: «0 ocultos, 0 absorbidas, 0 duplicados abiertos expuestos»). Dictamen del revisor: «bien hecha y justificada» |
| **Cliente HTTP único con barreras** (corrección A1 de `analyze.md`) | Las cuatro barreras SSRF (allowlist `https`, IPs privadas/metadata, redirecciones y límite de cuerpo) se centralizan en **una fábrica única** del dominio usada por **los tres consumidores** | En el plan original estaban solo en la descarga de carátulas; el **lector de feeds** y el **probador de disponibilidad** hacían `GET` sin política. Es la **primera red saliente del backend** (RG9-4): `seguridad` la aprobó y el revisor la calificó como «la mejor pieza del cambio» |
| **Siembra e2e por SQL** (RG9-7, `T939`) | Nuevo patrón: `frontend/e2e/mediosSeed.ts` siembra **por SQL**, no por API | F9 **no tiene alta manual** (FR-019), así que no puede crear fichas por API sin disparar la red real. Declarado en `plan.md` y en `tasks.md` como patrón nuevo, reversible y acotado; los e2e **no pulsan «Buscar episodios nuevos»** |
| **Topes de teclado de F3 y F6** (hallazgo A1) | Se suben los **topes de pasos** de `portada-accesibilidad.spec.ts` (30 → con el bloque de medios) y `ministerios-accesibilidad.spec.ts` (45) | El bloque añade ~11 elementos enfocables: el ciclo se recorre **entero**, solo que con más pasos. **Es un tope de esas pruebas, no un defecto del sitio**; ninguna aserción de comportamiento se relajó |
| **Aserción de `alt` para decoración** (hallazgo A2) | La aserción de F3 que exige `alt` no vacío en `main` **omite los subárboles `aria-hidden`** | El marcador neutro de carátula es **decoración** (el nombre va justo debajo) y está en un contenedor `aria-hidden`: `alt=""` es **la opción correcta** según WCAG. Se refina la aserción en vez de inventarle un texto |
| **RG9-6 (T932)** | Actualización **acotada** de las pruebas de frontend de F2/F4/F6/F7 que fijan la lista de permisos y los accesos del panel | Mismo gesto autorizado a F4/F6/F7 al activarse un módulo: solo recuentos, listas y sellos. **El backend no toca ninguna prueba de F2** y el catálogo sigue en 10 |
| **Tandas e2e** | La suite pasa a **seis tandas** y la tanda 3 usa la lista explícita `'portada-(?!medios)'` | El patrón `portada-` de la tanda 3 habría arrastrado `portada-medios.spec.ts` y la habría corrido dos veces (hallazgo A3) |

### Evidencia de cierre (T944, 2026-10-18)

| Verificación | Resultado |
|---|---|
| `make ci` | **en verde**: lint backend (gofmt/vet/golangci-lint) y frontend (ESLint/tsc), backend **unitarias + integración contra PostgreSQL real — 24 paquetes ok**, frontend, **`govulncheck`: 0 vulnerabilidades** y **`npm audit`: 0** |
| `go test -race ./...` | **en verde en los 22 paquetes del backend** (hallazgo N4 de la ronda 3; el CI del kit ejecuta `-race`) |
| Cobertura `backend/internal/medios/` | **81,1 %** (umbral ≥ 80 %) |
| Frontend | **Vitest 946 pruebas / 131 archivos** al cierre (930/129 en T944 + 16 de las correcciones) |
| e2e Playwright | **131 pruebas, 131 en verde** en las **6 tandas** (T944); la tanda de F9 (`medios-` + `portada-medios`) **25/25** y re-verificada tras cada ciclo |
| Migraciones | `000014`/`000015` con ciclo **up → down → up** verificado (incluido el `down` destructivo de `admin_actions`) y catálogo de permisos **intacto en 10** |
| Deriva de generados | **`make api-gen` sin deriva**; contrato vivo en **0.9.0** (7 endpoints: **2 públicos y 5 de panel**) |
| Entorno | **`.env.example` sin cambios** (F9 no introduce ninguna variable ni secreto) y **0 dependencias nuevas** (`golang.org/x/text` de indirecta a directa, misma versión; `package.json` intacto) |

*Nota de método (T945)*: las cifras de esta tabla provienen de la evidencia de cierre **T944**
entregada por el orquestador y de los informes de la fase 7 (misma fuente que el registro en
[`estado.md`](./estado.md)). Esta sesión del `documentador` **no tuvo acceso a shell ni a git** y
**no las re-ejecutó**.

### Lo conocido (aceptado y documentado; no bloquea el merge)

- **Tercer elemento dudoso contra una ficha ya emparejada (borde de spec; decisión del humano del
  2026-10-19: «se acepta y se documenta»)**: la pareja de «posible duplicado» es **de dos** (el
  modelo aprobado **no cambia**). Si una **tercera** coincidencia **dudosa** (solo título o solo
  día) cae sobre una ficha que ya tiene pareja, hoy **entra suelto y visible**, sin marca; solo es
  un problema si resulta ser el mismo sermón, y la mitigación es que **el equipo lo oculta a mano**.
  Queda como **mejora futura** extender el eje de duplicados a grupos. La **unión automática
  segura** (título normalizado + mismo día, FR-013) **sí** funciona contra una ficha emparejada
  (arreglo N1). Registrado en [`spec.md`](./spec.md) (nota bajo FR-014) y en
  [`data-model.md`](./data-model.md), y explicado al cliente en
  [`docs/entrega/F9-medios-predicas-y-podcasts.md`](../../docs/entrega/F9-medios-predicas-y-podcasts.md).
- **Flake preexistente de F3 en el e2e** (no es de F9): `frontend/e2e/portada-accesibilidad.spec.ts`
  mide un botón de **44 px** como **43,98** (redondeo interno de Chromium, del orden de **1/64 px**)
  de forma intermitente y **pasa al relanzar**. Anterior a esta rama; no afecta al producto.
- **Inestabilidad preexistente de la suite de integración completa** (`go test -tags=integration
  ./...`): interferencia entre suites sobre la base compartida (los `TRUNCATE` de `users`/`roles`).
  Verificada también en el commit base, así que **no la introduce F9**; la validación corrió la
  integración por paquetes contra PostgreSQL real.
- **`make ci` local no ejecuta `-race` y el CI sí** (el CI corre `go test -race` en
  `.github/workflows/ci.yml`): es una diferencia del kit, preexistente; el hallazgo N4 existía
  precisamente por ella. Conviene alinearlo (quedó registrado).

### Pendientes (no bloquean abrir el PR)

- **SC-008** — revisión con **lector de pantalla real** (NVDA/VoiceOver/TalkBack) de `/medios` y del
  bloque, en tema claro y oscuro: **0 hallazgos críticos**, 0 controles sin etiqueta, 0 trampas de
  foco y anuncio comprensible de cada enlace externo. **Post-despliegue.**
- **SC-012** — usabilidad con **≥ 6 personas** de distintas edades (2 por franja 18–35 / 36–59 / 60+,
  con poca experiencia incluidas): **≥ 90 %** encuentra una prédica y llega a la grabación sin ayuda,
  **≥ 90 %** entiende **antes** de activar el enlace que el sitio lo saca a YouTube/Spotify **en otra
  pestaña**, y el **cronómetro de SC-001 (< 30 s)**. **Post-despliegue**, con firma del humano.
  Protocolo listo en [`pruebas-usabilidad-SC-012.md`](./pruebas-usabilidad-SC-012.md).
- **Despliegue a producción** y **merge** del PR: solo con aprobación humana explícita; después,
  `make costos CERRAR=1`.

### Cómo probarlo (quickstart)

Recorrido completo con evidencia por sección en [`quickstart.md`](./quickstart.md) (§0–§14).
Resumen:

```bash
make up                 # db, redis, backend, frontend
make db-migrate         # 000014 y 000015
cd frontend && npm install
# Portada (visitante): http://localhost:5173/  -> bloque de medios al final
curl -s 'http://localhost:8080/api/v1/portada/medios?lang=es' | jq
# Sección completa:    http://localhost:5173/medios
curl -s 'http://localhost:8080/api/v1/medios?lang=es' | jq
# Panel: login con la cuenta sembrada de los e2e (frontend/e2e/helpers.ts)
# -> /panel/medios con el permiso «Medios (prédicas y podcasts)»
# -> botón «Buscar episodios nuevos» (10 fichas con los feeds del 2026-10-17)
```

Pruebas automatizadas: `make ci` (lint + backend unit/integración + frontend + security) y las e2e
**en seis tandas** con `--workers=1` (comandos exactos en la cabecera de
`frontend/e2e/medios-publica.spec.ts`). La suite de integración completa es inestable por
interferencia (preexistente, ver «Lo conocido»); correr por paquetes contra PostgreSQL real.

### Notas de entrega

- **Tareas**: **47/47 cerradas** (T901–T946 + T907A), incluidas T944 (verificación) y T945 (esta
  documentación). *Cubre el cierre de US1–US6 / FR-001…FR-030 / SC-001…SC-017 como entregable*, con
  los dos pendientes manuales post-despliegue arriba.
- **Documentación actualizada por `documentador`** (cierre del 2026-10-19): `CHANGELOG.md`
  (**0.8.0** con el resultado real de la validación y su «Conocido»), `README.md` (sección «Qué es
  F9», `/medios` y el botón en el quickstart, migraciones `000014`/`000015`, consultas sqlc, tandas
  e2e y enlaces), `docs/entrega/F9-medios-predicas-y-podcasts.md` (nota para el cliente, con el
  borde del tercer elemento en lenguaje no técnico), `pruebas-usabilidad-SC-012.md` (protocolo),
  `spec.md` (nota única de limitación conocida bajo FR-014), `data-model.md` (nota junto al eje de
  duplicados) y este `pr.md`. **No se tocó** código, pruebas, `plan.md`, `ux.md`, `tasks.md`,
  `estado.md`, `contracts/`, el roadmap ni `costos.json`.
- **Informes de la fase 7**: [`revision-2026-10-18-qa.md`](./revision-2026-10-18-qa.md),
  [`revision-2026-10-18-codigo.md`](./revision-2026-10-18-codigo.md),
  [`revision-2026-10-18-seguridad.md`](./revision-2026-10-18-seguridad.md),
  [`revision-2026-10-18-qa-ciclo2.md`](./revision-2026-10-18-qa-ciclo2.md),
  [`revision-2026-10-18-codigo-ciclo2.md`](./revision-2026-10-18-codigo-ciclo2.md),
  [`revision-2026-10-18-seguridad-ciclo2.md`](./revision-2026-10-18-seguridad-ciclo2.md) y
  [`revision-2026-10-18-codigo-ciclo3.md`](./revision-2026-10-18-codigo-ciclo3.md).
- **Contrato**: `backend/api/openapi.yaml` en **0.9.0** y sin deriva de `make api-gen` /
  `make sqlc-verify`.
- **Avisos para el cliente**: RG9-1 (el feed de YouTube puede devolver 404 desde servidores: el
  resumen lo dice y nada se oculta por esa causa), RG9-12 (el podcast trae **la misma carátula**
  para todos sus episodios de solo audio: es el arte del programa) y el borde del **tercer elemento
  dudoso** (ver «Lo conocido»).
- **Merge y despliegue**: solo con aprobación humana explícita (puertas pendientes en `estado.md`).
