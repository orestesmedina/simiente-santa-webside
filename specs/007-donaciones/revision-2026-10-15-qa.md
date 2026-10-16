# Revisión QA — F7 Ofrendas y diezmos (2026-10-15/16)

**Rol**: `qa-tester` · **Rama**: `007-donaciones` · **Alcance**: `git diff main...HEAD` (T701–T737, ~133 archivos ±25.000 líneas) contra la **spec v2 aprobada** (5 US, 25 FR, 14 SC).

**Veredicto: APROBADO** — todo lo automatizable pasa, incluida una **batería independiente con datos reales** que sembré y comprobé por mi cuenta (~36 verificaciones V contra la API en vivo): números íntegros y exactos, 0 borradores expuestos, redirección silenciosa, bloque de portada debajo de ministerios con tope 5 resuelto en el servidor, permiso «donaciones» verificado en el servidor con auditoría de intentos, validación IBAN/SINPE, moneda solo en IBAN, titular opcional y bilingüismo con respaldo. Los 2 únicos fallos e2e de mi corrida se debieron al **entorno** (base reconstruida en vacío, sin siembra de portada), no al producto, y quedaron en verde al restaurar la siembra. Quedan como pendientes manuales el protocolo de usabilidad SC-012, el lector de pantalla real y una pasada en dispositivo físico, y como hallazgos menores el residuo de datos de prueba del entorno y el acoplamiento e2e de F7 a contenido de portada sembrado por otras suites.

## Entorno y ejecución de suites (hecho por mí, no repetido del estado.md)

| Suite | Comando | Resultado |
|---|---|---|
| Stack de la rama | `make up` + `migrate up` (13/13, incl. **000012/000013**; `schema_migrations=13, dirty=f`) + `docker compose build backend frontend` (tras el último commit de código) | ✅ imágenes reconstruidas 2026-10-16 |
| Backend unitarias | `go test ./...` | ✅ 22 paquetes ok (`donaciones`: **94** funciones de prueba) |
| Backend integración (PostgreSQL real) | `DATABASE_URL_TEST=… go test -tags=integration -count=1 ./...` | ✅ todos los paquetes ok; `donaciones`: **117** pruebas (migraciones UpDownUp de 000012/000013, checks a nivel de base, transacción atómica con auditoría, catálogo de permisos sin cambios) |
| Cobertura del dominio | `go test -tags=integration -coverprofile` sobre `internal/donaciones` | **89,0 %** por sentencias del paquete; **`service*.go` ≈ 92,1 %** (media de 46 funciones), coherente con el 89,6 % reportado por desarrollo |
| Frontend unitarias | `npm test -- --run` (2 corridas completas) | ✅ **627/627** en 100 archivos, ambas en verde |
| E2e de F7 — Docker `mcr.microsoft.com/playwright:v1.63.0-jammy`, `--network host`, `--workers=1`, repo montado rw | `npx playwright test --config e2e/playwright.config.ts --workers=1 ofrendas- portada-ofrendas` (+ lote pendientes tras sembrar portada) | ✅ **23/23** de los 4 specs (`ofrendas-publica` 7, `ofrendas-panel` 9, `ofrendas-accesibilidad` 3, `portada-ofrendas` 4); corrida total 5,7 min + 20,5 s |
| Batería API en vivo (`curl`/psql, ~36 verificaciones propias, marcadas V-n) | sección pública, bloque de portada, panel, permisos, auditoría | ✅ (tabla siguiente) |

**Estado del stack**: `backend`/`frontend` reconstruidos con las imágenes de la rama; base **reconstruida en limpio** por mí (`down -v` + 13 migraciones + `setup/initialize` con las credenciales del e2e) tras descubrir que una corrida previa de mis pruebas de integración había dejado la cuenta de administración con otra contraseña (motivo del primer fallo e2e, con bloqueo de 15 min de F2 respetado: 3 inicios de sesión en total en toda la sesión).

### Incidencias de entorno de mi corrida e2e (no son defectos del producto)

1. **Corrida A falló en el login de administrador** (el seed de e2e logueó contra una base a la que mis pruebas de integración habían vaciado tablas): 4 fallos/19 sin correr por `429 rate_limited` de F2. Resolución: `make down -v` + migraciones + inicialización + reinicio de Redis (contadores); tras ello `login` 200 y el resto de la sesión respetó el límite.
2. **Corrida B: 19 ✓ / 2 ✗ / 2 sin correr**. Los 2 fallos (`ofrendas-publica:462` nav oculta y `ofrendas-accesibilidad:210` espera de «Quiénes somos» con timeout) provienen de una base recién creada **sin ninguna portada publicada**: el snapshot muestra la `nav` con `list` vacía. Tras ejecutar el seed de F3 (`portada-publica`, 1 ✓), **corrida C: 14/14 ✓** (incluye los 2 fallados y los 2 relegados). Comprobado: no hay assert relajado; el producto se comporta como la spec. → Hallazgo H3.

## Verificación independiente con datos reales (API viva, resultados por códigos)

Sembré por la API con la cuenta de administración real y comprobé por mi cuenta:

| # | Verificación | Prueba concreta ejecutada por mí | Resultado |
|---|---|---|---|
| V-1 | **IBAN con espacios y guiones → guardado normalizado** (FR-015/Q3) | `POST kind=iban number="CR05 0152-0200 1234 5678 90"` → **201**, DTO guarda `CR05015202001234567890` (**22** caracteres) | ✅ |
| V-2 | IBAN internacional válido (ES-24) aceptado (Q3) | `ES91 2100 0418 4502 0005 1332` → **201** `ES9121000418450200051332` | ✅ |
| V-3–V-6 | **0 números mal formados aceptados** (SC-009/Q3) | IBAN CR de 21 y de 23, país `XX`, y **emoji en el número** → **4×400** con `details.number` "El número no tiene el formato de su tipo", sin guardar nada | ✅ |
| V-7–V-8 | **SINPE con o sin `+506`, normalizado** (FR-016/Q4) | `"8888 8888"` → `+50688888888`; `"+506 8765-4321"` → `+50687654321` (ambos **201**) | ✅ |
| V-9–V-11 | SINPE no válido rechazado (default aprobado puerta 2: primer dígito **6/7/8**) | `51112233` (no celular) → 400; 7 dígitos → 400; 9 dígitos → 400 | ✅ |
| V-12–V-13 | **Moneda solo en IBAN** (Q2) | IBAN sin moneda → 400 `details.currency` "obligatoria en las cuentas IBAN"; SINPE con moneda → 400 "el SINPE Móvil no lleva moneda" | ✅ |
| V-14 | Moneda fuera de catálogo → 400 | `currency:"EUR"` → 400 "Elige una de las opciones válidas: CRC, USD." | ✅ |
| V-15 | Vacíos → 400 por campo, sin parciales (FR-014) | número e institución vacíos → 400 con `details.number` **y** `details.institution` | ✅ |
| V-16 | **Titular opcional: sin dato, la clave pública se OMITE** (US1 esc. 5) | ficha publicada sin titular en `GET /ofrendas`: **no existe la clave `holder`** (nunca vacía ni marcador) | ✅ |
| V-17 | Español base por campo (FR-007) | `holderEn` sin `holderEs` → **400** indicando qué falta | ✅ |
| V-18 | Límites de campo | institución 161 → 400; (rango probado también en textos) | ✅ |
| V-19 | Duplicado → 409 (P7-10; regla tipo+número de puerta 2) | crear de nuevo el mismo IBAN escrito con **otra separación** (se normaliza igual) → **409 conflict** | ✅ |
| V-20 | **Editar sin cambiar el número → 200** (analyze A2 corregido) | `PATCH` solo institución sobre la fila → 200 (la consulta de duplicados excluye `id <> propio`) | ✅ |
| V-21 | Cambiar el número al de OTRO método → 409 | `PATCH number=ES9121000418450200051332` sobre la fila ya ocupada → **409** | ✅ |
| V-22 | El formato valida POR TIPO | un número IBAN como SINPE → 400 (formatos disjuntos) | ✅ |
| V-23 | **0 borradores expuestos por cualquier vía** (SC-002/FR-012) | borrador con marca única (institución+titular) y texto introductorio en borrador: `grep` en `GET /ofrendas` y `/portada/ofrendas` en **es y en** (**4/4** respuestas) → **0 apariciones** de la marca, de `publicationState` ni de `"draft"`; con nada publicado la API responde **solo `{"lang":"es"}`** (claves `items`/`intro`/`usage` omitidas, nunca `null`, no hay 3xx) | ✅ |
| V-24 | `Cache-Control: no-store` (SC-003) | cabecera de la sección pública | ✅ |
| V-25–V-27 | **Publicar/retirar/editar visible de inmediato** (SC-003/FR-013) | publicar → aparece en la primera lectura; editar el número del publicado (`CR11 0001…` → `CR11 9999…`) → el **nuevo** número en la sección **y en el bloque de portada** en la primera lectura posterior | ✅ |
| V-28 | Devolver a borrador → desaparece de inmediato de sección **y** bloque | lectura tras el `PATCH` → ausente en ambas superficies | ✅ |
| V-29 | **Tope 5 del bloque resuelto en el servidor** (FR-022/SC-014) | 9 publicados (6 IBAN `sortOrder` 1-6 + 3 SINPE 1-3): sección **9** ítems en orden `iban…sinpe` (por `sort_order, id`); bloque devuelve **exactamente los 5 primeros** (`== sección[:5]`); con `?limit=50&page=2&offset=9` **sigue 5** (parámetros ignorados, no hay API de paginación) | ✅ |
| V-30 | **Textos bilingües con respaldo** (FR-008/SC-006) | `intro` es+en publicado: es «Introducción con tildes ñ ¿e¡ y 🙏» / en «Intro with tildes…»; `usage` solo español: en inglés muestra el **español** (nunca vacío) | ✅ |
| V-31 | Número íntegro idéntico entre idiomas | `items[0].number` es vs en → mismo valor exacto | ✅ |
| V-32 | **Texto en borrador no deja hueco** (Q10/US3 esc. 7) | intro → borrador con métodos publicados: la clave `intro` desaparece, `usage` sigue, sección coherente | ✅ |
| V-33–V-34 | Límite de texto 2.000 (default revisable) | 2.001 → **400**; exactamente 2.000 → **200** | ✅ |
| V-35–V-36 | textEs obligatorio y `textKey` del catálogo | vacío → 400; clave inválida `clave-invalida` → 400 | ✅ |
| V-37 | `lang` fuera de registro → 400 (nunca 3xx) | `?lang=fr` → 400 patrón "no es válido" | ✅ |
| V-38 | `lang` en mayúsculas se normaliza | `?lang=EN` → 200 | ✅ |
| V-39 | **Permiso verificado en el servidor** (SC-005): 6 llamadas forzadas → 6×403 | cuenta real con rol **solo `eventos`**: GET lista, POST crear, GET textos, PUT texto, PATCH editar, DELETE → **6×403** con mensaje claro | ✅ |
| V-40 | Los intentos forzados no cambian nada | «Hack» ausente del sitio público tras las 6 denegaciones | ✅ |
| V-41 | **Permisos independientes** (Decisión 5) | rol solo-`donaciones`: `GET admin/donaciones` → **200**, `GET admin/eventos` y `admin/ministerios` → **403**; la nav del panel muestra solo su módulo (e2e) | ✅ |
| V-42 | Auditoría de intentos denegados con el actor correcto | `admin_actions`: `result='denied'` + `action` `donation.*` con **actor = la cuenta restringida** (no la admin) | ✅ |
| V-43–V-44 | **Auditoría completa y legible** (SC-011/FR-017) | las **8 acciones** (`donation.create/update/delete/publish/unpublish/text.update/text.publish/text.unpublish`) presentes durante la sesión; validaciones fallidas → `result='failure'` (16 en mi sesión); etiqueta objetivo legible: **«Donaciones · T-num pub · CR11999999999999999977»** | ✅ |
| V-45 | **Solo textos publicados: sección viva, bloque ausente** (P7-13, decisión puerta 2) | 0 métodos + `usage` publicado: `GET /ofrendas` = `{lang,usage}` (sin `items`), bloque de portada **0 items** | ✅ |
| — | Auditoría de solo lectura | las rutas de auditoría son **GET** (F2, T240); no existe endpoint de écritura | ✅ |
| — | Catálogo de permisos sin cambios (FR-011) | `SELECT count(*) FROM permissions` = **10**; `donaciones` ya existía; `git diff` de migraciones no añade INSERT de permisos; e2e de F2 (`acceso.spec.ts`) ajustado acotadamente (RG7-6: usa `noticias`/`grupos` como módulos sin construir) | ✅ |
| — | Sin imágenes en F7 (FR-020) | el formulario no ofrece carga; e2e aserta `main img` = 0 | ✅ |

**Nota de servicio**: la Agrupación del IBAN en bloques es **solo presentación** (`agruparIban` añade espacios; la versión agrupada va `aria-hidden` y el número **íntegro sin separadores** en `sr-only`): el valor copiado por el botón y leído por pantalla es el normalizado de la API (e2e «copiar el valor normalizado exacto» ✓). El SINPE se publica tal cual el servicio lo normaliza (`+506` + 8 dígitos).

## Matriz US × escenario → cumple → evidencia

### US1 — Ver cómo apoyar a la iglesia (P1) → **cumple**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 intro + grupos por tipo + textos | ✅ | e2e `ofrendas-publica` (1) (grupos «01 Cuentas IBAN»/«02 Números de SINPE Móvil» + intro + uso); V-30 |
| 2 IBAN íntegro, agrupado, entidad, titular, moneda | ✅ | V-1/V-16; `NumeroDonacion` (agrupado decorativo + exacto en sr-only); e2e grupo (1) |
| 3 SINPE sin moneda, normalizado | ✅ | V-7/V-8; DTO público: `currency` solo si `kind==='iban'` (`localizeMetodo`) |
| 4 orden: IBAN primero y manual dentro del grupo | ✅ | V-29 (9 ítems, `sort_order,id`); `ORDER BY CASE kind…, sort_order, id` + índice (go-int `OrdersAndFilters`) |
| 5 sin titular: ficha coherente, sin línea vacía | ✅ | V-16 (clave omitida); e2e con los dos casos «sin titular» del seed |
| 6 identifica cómo donar (estructura) | ✅ | V-30 (intro+grupos+uso); el «90 %» es **SC-012, manual** |
| 7 sin nada publicado: redirección limpia sin aviso | ✅ | e2e (7) `393` con **detector de destello** (MutationObserver + control positivo) + `replace` (no añade historial) + **0 enlaces `/ofrendas`** en barra, hamburguesa y mapa; V-23 la API no redirige (200 `{lang}`) |
| 8 no se confunde con F4/F5/F6 | ✅ | modelo y endpoints separados (`donation_*`); e2e sin contenido ajeno (SC-013) |

### US2 — Gestionar la información desde el panel (P1) → **cumple**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 crear con válidos (titular opcional, moneda solo IBAN, sin notas) | ✅ | V-1 (201); formulario: campo moneda solo con `kind==='iban'`; e2e panel (2) |
| 2 sin permiso: mensaje claro, nada cambia | ✅ | V-39/40; e2e panel (9) con 7 llamadas forzadas por la UI/API |
| 3 inválidos → indicar qué corregir, sin medias | ✅ | V-3–V-6, V-9–V-15, V-17, V-18; `validateMethod`/`validateTexto` con `details` por campo; revalidación del PATCH fusionado (R7-15) |
| 4 IBAN con espacios → normalizado guardado + agrupado publicado | ✅ | V-1 + `agruparIban` + e2e de copia |
| 5 SINPE con/sin `+506` → normalizado al guardar | ✅ | V-7/V-8 |
| 6 editar → visible de inmediato | ✅ | V-26/27 (live); «se publica al guardar» (FR-013) |
| 7 eliminar con confirmación + auditoría | ✅ | e2e panel (8) con diálogo; `donation.delete` en `admin_actions` (V-43) |
| 8 Titular: campos vacíos no obligan; inglés opcional | ✅ | V-16/V-17; `holderEn` sin `holderEs` → 400 qué falta |
| 9 cruce con otros módulos | ✅ | V-41 (solo-donaciones → eventos/ministerios 403); e2e panel (9) en ambas direcciones |
| 10 reordenar conserva y es el orden visible | ✅ | V-29 (mismo orden en panel listado y público); go-int; e2e panel (7) |
| 11 textos editados y conservados | ✅ | V-30 (PUT intro/usage 200, visibles al momento); textos sin borrado (se retiran a borrador, decisión P7-17) |

### US3 — Controlar la publicación (P2) → **cumple** (detalle P7-13 por decisión aprobada puerta 2)

| AS | Cumple | Evidencia |
|---|---|---|
| 1 borrador oculto, resto coherente | ✅ | V-23 (4/4); e2e getDrafts |
| 2 sin nada publicado: nav/bloque ocultos + `/ofrendas` redirige limpia | ✅ | e2e (7) completo; V-23 (API responde `{lang}`; la decisión es del frontend, P7-21) |
| 3 publicar → aparece en grupo y orden | ✅ | V-25; e2e panel (1) |
| 4 devolver a borrador → desaparece (sección + bloque) | ✅ | V-28 |
| 5 NUNCA expone borrador por cualquier vía | ✅ | V-23; DTO público **sin campos de estado** (imposible filtrarlo capas arriba); filtro `WHERE publication_state='published'` en banco (go-int `ExcludesDrafts`) |
| 6 panel muestra estado de cada método y texto | ✅ | `StatusPill` por fila y por texto en `DonacionesPage`/`DonacionesListsByType` |
| 7 texto borrador con métodos → sin hueco | ✅ | V-32; e2e (5) `343`. **Nota**: la sección vive con **≥1 método o ≥1 texto** publicados (decisión del humano puerta 2, aplicada en `useSiteLinks`; con solo textos el **bloque** no aparece — V-45) |

### US4 — Ver la información en inglés (P2) → **cumple**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 selector: interfaz y contenidos traducidos, persistente | ✅ | e2e (3) `263` (selector de F10 reutilizado, radios) |
| 2 fallback por campo, nunca vacío | ✅ | V-30 (`usage` en→es); `pickString` en servidor |
| 3 sin preferencia → español | ✅ | selector de F10 (tanda propia de F10 en verde) |
| 4 volver a español sin pérdida | ✅ | e2e (3) |
| 5 número/tipo/entidad/moneda idénticos | ✅ | V-31 (es=en), V-2 (moneda invariante); FR-007 «no traducible» |

### US5 — Dispositivos y accesibilidad (P2) → **cumple lo automatizable; manual: lector real**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 sin scroll horizontal, números completos a 320 px | ✅ | e2e (4) `306` y accesibilidad (289): `overflow-wrap:anywhere`, **launching sin elipsis**, 320/768/1280 |
| 2 navegación simple (F10) | ✅ | barra/hamburguesa/mapa reutilizados (suites de F10); e2e accesibilidad |
| 3 teclado con foco visible, en claro y oscuro | ✅ | e2e accesibilidad (1)(302): recorrido completo, copia con `Enter`, sin trampa de foco (`Shift+Tab`), CTA del bloque |
| 4 contraste y texto al 200 % | ✅ | `contrastFailures` = [] en claro y oscuro; tokens de F10; (200 % en pantalla real ⚠ manual) |
| 5 mismas funciones en todo dispositivo | ✅ | e2e 3 anchos + bloque a 320 |

### US6 — Bloque de la portada (P2) → **cumple**

| AS | Cumple | Evidencia |
|---|---|---|
| 1 bloque debajo de ministerios, ítems con tipo | ✅ | e2e portada (1): posición **exacta por DOM** tras `MinisteriosBlock` (HomePage: ministerios → ofrendas → quiénes somos); chip de tipo |
| 2 >5 → primeros 5 por orden de sección, sin paginación | ✅ | V-29; e2e (1) con 8 sembrados |
| 3 <5 → exactos, sin huecos | ✅ | e2e (2) `225`; variante del seed con `{publishedCount: 2}` |
| 4 0 publicados → bloque no existe | ✅ | e2e (3) `256`; `return null` sin título ni mensaje |
| 5 botón abre `/ofrendas` en un paso | ✅ | e2e (1)(292) click y teclado |
| 6 sin borradores en el bloque | ✅ | V-23 (portada es/en) + e2e (1) |
| 7 bilingüismo + responsividad sin tocar el resto de la portada | ✅ | e2e (4) a 320 px; posición exacta prueba que solo se inserta; bloque resuelto en es/en (V-23) |

## SC — resumen

| SC | Estado | Evidencia principal |
|---|---|---|
| SC-001 100 % sin sesión | ✅ estructura / ⚠ crono manual | todas las V y e2e públicas sin cookies; el «<30 s» queda en el protocolo SC-012 |
| SC-002 0 borradores | ✅ | V-23 (4 endpoints, 2 semillas), DTO sin estado, e2e |
| SC-003 primera carga posterior | ✅ | V-24–V-28, `no-store`, e2e panel (1) |
| SC-004 crear+publicar <2 min | ✅ | e2e panel (1) con assert temporal (pasado) |
| SC-005 permiso en 100 % operaciones | ✅ | V-39 (6×403), V-41, V-42; e2e panel (9); catálogo 10 intacto (go-int) |
| SC-006 inglés completo, 0 vacíos, invariantes | ✅ | V-30/31, e2e (3) |
| SC-007 0 scroll a 320 | ✅ | e2e (4) accesibilidad + portada (4) |
| SC-008 0 hallazgos críticos AA | ✅ automatizable / ⚠ lector real | accesibilidad claro/oscuro verde; lector real manual |
| SC-009 0 números mal formados; íntegros | ✅ | V-3–V-6, V-9–V-11 (rechazos), V-1/2/7/8 (exactos), e2e copia, checks SQL 000012 |
| SC-010 0 secciones vacías + redirección limpia | ✅ | e2e (7) (destello, replace, 0 enlaces muertos), V-23/45 |
| SC-011 auditoría completa, no editable | ✅ | V-43/44 (8 códigos, failure y denied con actor), transacción atómica (go-int rollback), GET-only |
| SC-012 usabilidad ≥90 % | ⚠ **PENDIENTE manual** | protocolo aún no creado para F7 (ver Pendiente manual) |
| SC-013 0 F4/F5/F6 en F7 | ✅ | modelo/endpoints separados; e2e sin mezcla |
| SC-014 bloque exacto (5/en su orden/posición/oculto) | ✅ | V-29/45 + e2e portada (1)–(4) |

## Hallazgos

### Bloqueantes

**Ninguno.**

### Observaciones / mejoras (no bloqueantes)

1. **H1 — Residuo de datos de prueba en la base de desarrollo** (menor): 2 roles (`QA sin F7 …`, `QA solo F7 …`) + 2 cuentas **desactivadas** creadas para V-39/V-41. F2 no permite eliminar un rol con cuentas asignadas (regla propia, verificado: `409 conflict`), así que quedan hasta `make down -v` (que también limpia los roles/cuentas del e2e de la tanda de permisos). No afecta a otros entornos.
2. **H2 — Errores «unhandled» de infra en vitest** (menor): en las 2 corridas completas de `npm test` (627/627 ✓) vitest reportó entre 1 y 15 «Unhandled Error» de tipo `Pool.schedule`/timers (teardown de workers en WSL); el exit code fue 0. Es ruido del entorno, no del producto; conviene vigilarlo en CI (la CI no corre e2e ni debiera cambiar por esto).
3. **H3 — Los e2e de F7 asumen contenido de portada sembrado por otras suites** (menor, robustez): 2 pruebas (`ofrendas-publica` «sin nada publicado» y `ofrendas-accesibilidad` «la sección se recorre con teclado») fallan en una **base limpia** porque esperan enlaces de secciones de portada («Quiénes somos») y una `nav` no vacía, que hasta hoy siembran specs de F3/F10 en otra matriz. El producto es correcto (verificado: al sembrar portada, 14/14 ✓). Sugerencia opcional: que `ofrendasSeed.ts` (o ese par de specs) siembre el mínimo de portada (o deje de depender de él) para que F7 sea verde en cualquier base; **no corregido en esta revisión** (es código de prueba de F7 y lo decide el equipo).
4. **H4 — Cobertura del paquete 89,0 %** (`service*.go` ≈ 92,1 %): in line with the annoncé 89,6 % (variación de corrida); cumple el estándar del proyecto (service ≥ 80 %). Quedan ramas de errores de IO no cubiertas; mejora opcional.
5. **H5 — El tope interno de la sección es 200** (configuración interna RG7-12): la spec dice «sin límite artificial» y el plan fija un techo de seguridad técnico de 200 ítems en la consulta pública; documentado en el código, sin impacto esperado. SiImplOptions se lo quiere levantar, es cambio de una constante.

### Flaky preexistente (F3/F4, `qa-bugs-004.spec.ts`)

No lo ejecuté en esta revisión (no corresponde a F7 y consume logins). `git diff main...HEAD -- frontend/e2e/qa-bugs-004.spec.ts` = vacío (F7 no tocó ese spec ni su comportamiento). Se mantiene la observación de F4/F6.

## Pendiente manual (protocolo y responsable)

- **SC-012** — protocolo de usabilidad (mínimo 6 personas de distintas edades y experiencia; incluye **cronómetro de 30 s de SC-001**). No existe aún para F7; precedentes F4 (SC-010) y F6 (SC-012) lo exigen **antes del cierre** de la funcionalidad. Responsable: el equipo (la persona que designe el humano).
- **Lector de pantalla real** (NVDA/VoiceOver) sobre `/ofrendas` y el bloque de la portada, y una pasada en pantalla táctil física. La estructura semántica, jerarquía, etiquetas, foco y contraste ya están automatizados y en verde.
- **SC-001** (cómo donar «en menos de 30 segundos») se mide dentro del protocolo SC-012.

## Confirmación

**No toqué ningún código de producción ni pruebas existentes** ni ninguna spec/plan/ux/estado. Lo creado: este informe + los datos de prueba de la base (limpiados los métodos/textos; roles/cuentas QA residuales documentados en H1). En el entorno reconstruí el stack (`down -v`, 13 migraciones, `setup/initialize`), apliqué las migraciones y ejecuté las suites y la batería en vivo; el módulo quedó **vacío (0 métodos, los dos textos en borrador)** y el sitio público sin contenido de F7.

## Veredicto final

**APROBADO** para la puerta de validación: emitir PR con el `documentador`, con **tres pendientes manuales** (protocolo SC-012 con cronómetro de 30 s, lector de pantalla real y pasada física) y los hallazgos H1–H5 (ninguno bloqueante) asentados; ejecutar el protocolo SC-012 **antes del cierre** de la funcionalidad, como en F4 y F6.
