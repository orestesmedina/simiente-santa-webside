# Research — F6 Ministerios

**Fecha**: 2026-10-13 · **Spec**: [spec.md](./spec.md) (aprobada v2 por el humano el 2026-10-13;
Q1–Q9 resueltas) · **Plan**: [plan.md](./plan.md)

> **Notas de partida.** (1) La spec v2 ya cerró Q1–Q9 y el humano **confirmó además, el
> 2026-10-13** (estado.md), la lectura de Q6 (horario y lugar como **textos libres**, sin
> recurrencia ni calendario) y el **tope de 5** del bloque de la portada: este documento las da por
> firmes y solo las reseña. (2) El único punto que este research **no** decide es el **default del
> interruptor de visibilidad del encargado** al crear (R6-3): se eleva a la puerta 2.

Este documento resuelve las decisiones técnicas abiertas de F6 con alternativas y motivos. La
numeración es **R6-1…R6-17**, propia de esta funcionalidad, para no colisionar con R4-1…R4-16 de
F4, R3-1…R3-18 de F3 ni R1–R23 de F2. Donde se reutiliza una decisión ya tomada se cita (R3-x,
R4-x, P4-x) y **no se reabre**.

Contexto reutilizado sin cambios: capas `handler → service → repository` y reglas R1–R8 de
`docs/tecnico/arquitectura.md`; sobre de éxito = DTO directo / `{items}` en listados (§8.1.2);
paginación `platform/paginate` (§8.1.3); `ORDER BY` determinista (§8.1.7); validación
`platform/validate`; autorización `middleware.AdminChain`/`AuthzByModule`; auditoría en
`admin_actions` con `InsertAdminAction` transaccional (R3-11); imágenes en disco local con
`platform/storage` (R3-8) y descarga en `internal/media/` con `IsFilePublished` (R4-6);
bilingüismo por columnas `*_es`/`*_en` con resolución `?lang=` en el servidor (R3-1/R3-2);
`publication_state` por elemento con consultas `…Published` (R3-3); `sort_order INTEGER NOT NULL
DEFAULT 0` con índice y ordenación `sort_order, id` (Q5, precedente `000005`/`000007`).

---

## R6-1. Modelo de datos: **una** tabla `ministries`, con el encargado como campos del ministerio

**Pregunta.** FR-001/FR-019/Key Entities: el ministerio tiene nombre, descripción, encargado
opcional (nombre + un medio de contacto opcional + visibilidad), enlace opcional, horario y lugar
opcionales e imagen opcional. ¿Cómo se modela, sobre todo el encargado?

**Decisión.** Una **sola tabla** `ministries` (ver [data-model.md](./data-model.md)). El encargado
**no es una entidad**: son **columnas del ministerio** (`leader_name_es`/`leader_name_en`,
`leader_phone`, `leader_email`, `leader_visible`), tal como fija Key Entities ("no es un directorio
de personas… es un campo del ministerio; varios ministerios pueden compartir encargado **sin ficha
compartida**"). No hay tabla `leaders` ni tabla de traducciones: patrón R3-1 literal.

Por qué una tabla y no un agregado de dos: el encargado no tiene vida propia (no se lista, no se
busca, no se edita por separado, no tiene auditoría propia); separarlo crearía una FK y un CRUD que
la spec prohíbe conceptualmente (Out of Scope: "directorios de encargados ni fichas de personas").

**Alternativas descartadas.**

| Alternativa | Por qué no |
|---|---|
| Tabla `leaders` + FK desde `ministries` | Contradice Key Entities ("sin ficha compartida"): dos ministerios con el mismo nombre de encargado compartirían fila y un edit los cambiaría a ambos |
| EAV / tabla de traducciones (`ministry_translations`) | Rompe el patrón R3-1 probado en F3/F4 (columnas `*_es`/`*_en`) y complica sqlc sin ganancia |
| JSON (`leader JSONB`) | Pierde los `CHECK` (exclusión celular/correo, longitudes, "contacto exige nombre") y el filtrado; `pgtype`/sqlc se complica |
| Encargado obligatorio con contacto | Descartado por Q1/Q8: el encargado es opcional y su contacto también |

---

## R6-2. Privacidad de la ficha del encargado (FR-026/SC-014): garantía en la **capa de lectura pública**, no en la interfaz

**Pregunta.** FR-026 es el requisito crítico y no negociable de F6: con `leaderVisible` apagado
—nombre **y** contacto— **ningún dato del encargado puede salir del servidor hacia ninguna ruta
pública** (API, HTML, metadatos, sitemap, bloque de portada, respuestas forzadas sin sesión).
"No basta con ocultarlo en la interfaz".

**Decisión.** La regla se implementa en **tres capas independientes**, de modo que un fallo en una
no filtra (defensa en profundidad) y que la prueba sea mecánica ("0 apariciones"):

1. **DTO público sin campos para el dato oculto.** El tipo `MinisterioPublico` del contrato público
   (y el struct Go equivalente de `service_public.go`) tiene un **objeto `leader` opcional**
   (`LiderPublico`) que **solo se construye** cuando `leader_visible = TRUE` **y** existe
   `leader_name_es`. Cuando no, la clave `leader` **no existe** en la respuesta (ni vacía ni nula).
   No existe ningún otro campo público que transporte nombre, celular ni correo: el DTO público no
   tiene `publicationState`, ni `leaderPhone`, ni `leaderEmail`. El **bloque de la portada**
   (`ItemMinisterioPortada`) solo lleva `id`, `name` e `image`: **ni siquiera tiene dónde filtrar**.
2. **La consulta pública no devuelve el dato.** Las consultas `…Published` de
   `queries/ministries.sql` seleccionan los campos del encargado con
   `CASE WHEN leader_visible THEN leader_name_es END`, `…leader_phone`, `…leader_email`
   (equivalente a `NULLIF` condicional): con el interruptor apagado **la fila que sale de
   PostgreSQL ya no contiene** el nombre ni el contacto. Un error futuro de mapeo en el service
   recibiría `NULL`, no el dato. Las consultas del **panel** (tras `AdminChain("ministerios")`) sí
   seleccionan todo: dato guardado ≠ dato publicado (FR-025/FR-026).
3. **Pruebas "0 apariciones"** en las tres capas: service (el DTO público de un ministerio oculto
   **no contiene** el nombre ni el contacto, ni como campo ni como cadena), handler (el cuerpo JSON
   serializado no los contiene, en `lang=es` y `lang=en`, ni en la sección ni en el bloque),
   repository (la fila pública viene con `NULL`) y **e2e** (con `leaderVisible: false` y datos
   reales cargados, se recorre la página, se hace `page.content()` / respuesta de red y se comprueba
   que **0 cadenas** del nombre y del celular/correo aparecen — mismo método que pide SC-014:
   inspección de tráfico y de HTML—).

**Por qué no basta con ocultar en la interfaz.** La SPA de React no renderiza HTML del servidor
(no hay SSR), pero la API pública **sí** es una vía de salida: si el DTO público llevara el dato,
cualquier visitante lo leería con `curl` aunque la interfaz no lo pintara (exactamente el caso que
`seguridad` audita). Por eso la garantía vive en el DTO + la consulta, no en el componente.

**Cadenas de salida públicas revisadas una a una** (inventario para `seguridad`):

| Vía pública | ¿Puede contener datos del encargado? | Por qué |
|---|---|---|
| `GET /api/v1/ministerios?lang=` | **No** | `leader` solo si `leaderVisible`; la consulta ya devuelve `NULL` si no (R6-2.1/2) |
| `GET /api/v1/portada/ministerios?lang=` | **No** | DTO sin campos de encargado (solo `id`, `name`, `schedule`/`place` e `image`) |
| `GET /api/v1/media/{fileName}` | **No** | Solo binarios de imagen referenciados por contenido publicado |
| HTML inicial / metadatos / sitemap del pie | **No** | SPA sin SSR: el HTML no incrusta contenido; el pie y el sitemap se pintan desde las respuestas anteriores |
| Respuestas de error públicas | **No** | `ErrorEnvelope` genérico (F1); los logs no registran cuerpos de petición (F2) |
| Rutas de panel `/api/v1/admin/ministerios` | Sí (correcto) | Tras `AdminChain("ministerios")`: el dato guardado vive para quien tiene el permiso |

**Alternativas descartadas.**

| Alternativa | Por qué no |
|---|---|
| Ocultar solo en el componente React | Contradice literalmente FR-026 ("no basta"); el dato saldría por la API |
| Enviar el dato y borrarlo en el handler con un `if` | Está más cerca del fallo: cualquier DTO nuevo o serialización directa lo expone; sin garantía por consulta |
| Cifrar / ofuscar el contacto en la BD | La spec pide conservarlo íntegro y editable en el panel; el problema es la **salida**, no el almacenamiento |
| Dos endpoints públicos (uno "con ficha" y otro sin) | Duplica superficie pública para ocultar un campo; el riesgo de filtración se multiplica |
| Devolver `leader: null` explícito | La clave omitida es el patrón de F3/F4 (SC-012: sin líneas vacías) y evita que el cliente pinte "Encargado: —" |

---

## R6-3. El interruptor `leaderVisible` (Q9/FR-025): dato guardado ≠ dato publicado

**Pregunta.** Q1/Q9: el equipo decide con un interruptor si **la ficha entera del encargado
(nombre y contacto)** se publica. ¿Cómo se llama, cómo se guarda y **con qué valor por defecto**
nace un ministerio con encargado?

**Decisión.**

- **Nombre**: `leaderVisible` (columna `leader_visible BOOLEAN NOT NULL`), con etiqueta de
  interfaz «Mostrar la ficha del encargado al público». La spec llama al interruptor `showContact`
  "o el nombre que fije el arquitecto" (FR-025); se elige `leaderVisible` porque desde Q9 el
  interruptor gobierna **nombre y contacto** (la ficha entera), y `showContact` sería engañoso.
  **Equivalencia documentada**: `showContact` de la spec = `leaderVisible` del contrato y del
  código (FR-026/SC-014 se leen con esa clave).
- **Semántica**: el interruptor gobierna **solo la publicación**, nunca el guardado (FR-025/US2
  esc. 10–11): apagarlo y guardar conserva nombre y contacto en el panel; activarlo de nuevo los
  publica **sin reescribir nada**. No hay borrado derivado, ni cascada, ni doble fila: es un
  booleano más del guardado completo (R6-16).
- **Reglas con el resto del modelo**: `leader_visible` sin datos de encargado es válido y no hace
  nada; con encargado pero sin contacto, la ficha pública muestra **solo el nombre** (Q8: no es un
  error); con encargado oculto, el ministerio sigue publicándose con nombre, descripción, imagen y
  horario/lugar (FR-002/SC-017).
- **Auditoría**: un guardado que solo cambia este booleano registra `ministry.leader.show` /
  `ministry.leader.hide` (R6-12) — así el registro de solo lectura responde "quién ocultó o mostró
  la ficha del encargado", que es exactamente la operación que FR-017 enumera aparte.

**Default al crear — PUNTO ABIERTO para la puerta 2 (no lo decido yo).** La spec no fija si un
ministerio nuevo con encargado nace **visible** u **oculto** y es la única decisión de este plan que
queda al humano (ver [plan.md](./plan.md) → "Puntos que necesitaban aprobación humana"):

- **(a) `TRUE` (visible) por defecto** — recomendado: el equipo pone el encargado para que el
  visitante sepa a quién acercarse (US1/FR-002) y el patrón del sitio es "se publica al guardar";
  el interruptor está en el formulario **desde la creación** (Q1), así que ocultar es una decisión
  consciente de una sola acción.
- **(b) `FALSE` (oculto) por defecto** — privacidad por defecto: la publicación de un celular o
  correo personal exige un acto afirmativo; protege al equipo que llena el formulario rápido.
- **(c) Sin default**: el formulario exige elegir («¿Mostrar la ficha del encargado al público?
  Sí/No») y el DTO requiere `leaderVisible` explícito cuando hay encargado.

La migración `000010` escribe de momento `DEFAULT TRUE` con un comentario que remite a esta puerta:
si el humano elige (b) o (c), cambia **una línea** de la migración (todavía no aplicada en ningún
entorno) y la validación del DTO.

---

## R6-4. Medio de contacto del encargado (Q8): celular **o** correo, nunca ambos

**Pregunta.** Q8: el contacto es **opcional aunque haya encargado** y es **un solo medio**
(celular o correo). ¿Cómo se guarda y cómo se valida?

**Decisión.** Dos columnas anulables y **mutuamente excluyentes**, `leader_phone` y `leader_email`,
con `CHECK` que impide que ambas tengan valor y con la regla "contacto exige nombre"
(`leader_name_es` presente). Se validan con los **tags ya existentes** de `platform/validate`
(`phone` y `email`, usados por F2/F3): **cero código de validación nuevo**, solo reglas de dominio.

- `leader_phone`: tag `phone` (formato de teléfono con `+` opcional, separadores habituales y ≥ 7
  dígitos; `validPhone` de `platform/validate`), longitud 7–32 (mismo `CHECK` que `home_contact`).
- `leader_email`: tag `email` (formato de correo), longitud 3–254 (mismo `CHECK` que
  `home_contact`).
- **La validación es de formato, no de existencia ni vigencia** (default revisable de la spec:
  "que el celular responda o que el correo exista" está en Out of Scope).
- En la ficha pública (solo con `leaderVisible`): `phone` **o** `email` viaja en `LiderPublico`
  (`phone?` / `email?`, nunca los dos — la BD ya lo impide); si no hay contacto, solo el nombre.

**Alternativas descartadas.** Una sola columna `leader_contact` con un `kind` (obliga a validar el
valor contra el tipo en el service y pierde los tags `phone`/`email`); guardar celular y correo en
dos campos simultáneos (contradice Q8: "nunca ambos"); normalizar el teléfono a dígitos (alteraría
lo que el equipo escribió, igual que F3 no normaliza el horario).

---

## R6-5. Validación del enlace (FR-015/Q3): cualquier `https` válido y seguro, sin restricción de dominios

**Pregunta.** FR-015/Q3: el enlace al grupo de WhatsApp es opcional; si existe debe ser "cualquier
`https` válido y seguro", rechazando `javascript:`, `data:`, `vbscript:` y cualquier esquema
distinto de `https`. ¿Dónde vive la regla y cómo se refuerza?

**Decisión.** La regla ya existe y es exactamente la que pide la spec: el tag **`url`** de
`platform/validate` (F3, R3-14) llama a `validURL`, que acepta **solo** `net/url` con
`Scheme == "https"` y `Host != ""`. Eso rechaza por construcción `javascript:`, `data:`,
`vbscript:`, `http:`, rutas relativas y cadenas sin host, **sin mirar el dominio** (Q3: no se
exige que sea de WhatsApp). Tres capas:

1. **Service** (autoridad): tag `url` + reglas de dominio — longitud ≤ 500, sin caracteres de
   espacio ni de control (`strings.ContainsFunc` con `unicode.IsSpace`/`IsControl` → `400 invalid`
   con `details.whatsappUrl`), y **sin transformar** el valor: lo guardado se sirve tal cual
   (FR-015: "el enlace mostrado DEBE ser exactamente el enlace guardado tras la validación").
2. **Base de datos**: `CHECK (whatsapp_url IS NULL OR whatsapp_url ~ '^https://[^[:space:]]+$')`
   (defensa en profundidad: ni un `INSERT` forzado guarda un esquema no `https`).
3. **Render**: React pinta `href` como atributo (sin `dangerouslySetInnerHTML`, constitución §IV);
   el cliente además valida con Zod antes de enviar, pero **el servidor es la autoridad**.

La **vigencia** del destino no se verifica (Out of Scope): el sitio solo enlaza.

**Prueba exigida** (SC-009): tabla del service con los casos `https://chat.whatsapp.com/…`,
`https://ejemplo.com/grupo` (dominio **no** de WhatsApp: **aceptado**, Q3),
`javascript:alert(1)`, `JAVASCRIPT:alert(1)`, `data:text/html,…`, `vbscript:…`, `http://…`,
`//evil.example`, `https://` (sin host), `https://a b` (espacio), string de 501 caracteres →
aceptados solo los `https` válidos; e2e del panel que intenta guardar `javascript:alert(1)` y ve
el error sin guardar nada, y e2e público que comprueba que el `href` renderizado es el valor
guardado.

**Alternativas descartadas.** Restringir a dominios de WhatsApp (descartado por el humano en Q3);
una lista blanca de hosts (lo mismo, más mantenimiento); librería de validación de URLs (la stdlib
basta y la regla ya está probada); guardar el `href` reescrito/normalizado (FR-015 prohíbe
transformaciones que alteren el destino); validar solo en el frontend (§IV).

---

## R6-6. Permisos: se **activa** `ministerios`; F6 **no** crea permisos ni toca F2

**Pregunta.** FR-011: la gestión exige el permiso **«ministerios»**, "ya existente en el catálogo
fijo de F2… F6 NO crea permisos nuevos ni modifica el modelo de permisos ni las migraciones de F2".
¿Qué hay que tocar?

**Decisión.** **Nada del modelo de permisos.** Verificado en
`000002_create_roles_and_permissions`: el catálogo sembrado incluye
`('ministerios', 'Ministerios')` (junto a `grupos`, `donaciones`…). F6 solo:

- monta **un** grupo de panel con `middleware.AdminChain("ministerios", deps)` (el patrón de F2
  verificado en el servidor: authn → guard de cambio de contraseña → authz(`ministerios`) → CSRF),
  con el `Recorder` del propio dominio `ministerios` para registrar denegaciones (`result='denied'`);
- en el frontend añade `MINISTERIOS = 'ministerios'` a `lib/permissions.ts` (y a
  `MODULE_PERMISSIONS`, con lo que `isPermissionAvailable` deja de marcarlo como «Disponible más
  adelante») y protege `/panel/ministerios` con `RequirePermission(MINISTERIOS)`;
- **no** hay migración de permisos (a diferencia de F4, que tuvo que sembrar `categorias` en
  `000009`): el catálogo de `permissions` **no cambia** y sigue en 10 filas;
- el administrador inicial ya tiene `ministerios` porque el init de F2 asigna **todo** el catálogo
  dinámicamente (sin cambios en F2).

Permisos independientes (Decisión 5): tener `ministerios` no da acceso a `eventos`, `portada` ni
`admin_usuarios_roles`, y al revés (SC-005). Toda operación no autorizada → `403 forbidden`
genérico y registrada.

**Alternativas descartadas.** Crear un permiso nuevo por cada sub-acción (el catálogo de F2 es por
módulo, no por operación); derivar el acceso de `portada` o de `eventos` (mezclaría módulos
independientes); comprobar permisos solo en el cliente (CWE-862).

---

## R6-7. Imágenes: se reutiliza `platform/storage` + `internal/media` (F3/F4) y `IsFilePublished` crece con `ministries`

**Pregunta.** FR-020/Q4: imagen opcional por ministerio, con texto alternativo. F3/F4 ya tienen
subida multipart validada (firma binaria, JPEG/PNG/WebP, ≤ 8 MB, nombre generado
`img_<uuid>.<ext>`, disco local con volumen `uploads_data`) y descarga pública
`GET /api/v1/media/{fileName}` con la regla "solo archivos referenciados por contenido publicado"
(`IsFilePublished`, en `internal/media/` desde F4). ¿Cómo entra F6 sin duplicar nada?

**Decisión.** Reutilización literal (P4-7 de F4):

1. **Subida por módulo**: `POST /api/v1/admin/ministerios/imagenes` (multipart) tras
   `AdminChain("ministerios")`, con `platform/storage` tal cual; se audita con
   `ministry.image.upload` (fail-closed, como en F3/F4).
2. **Descarga**: `GET /api/v1/media/{fileName}` **sin cambios de ruta ni de cabeceras**; la
   consulta `IsFilePublished` (`queries/media.sql`) **crece** con una tercera rama:
   `ministries.image_file = :file_name AND publication_state = 'published'`. La imagen de un
   ministerio en borrador **no se sirve** (FR-012: "por ninguna vía"), y vuelve a servirse al
   publicarlo (mismo comportamiento que F3/F4 ya definen para su contenido).
3. `image_alt_es` es **obligatorio si hay imagen** (`CHECK` en la BD + validación condicional en el
   DTO) y `image_alt_en` opcional con fallback; quitar la imagen = `imageFile: null` (y se limpia el
   alt en la misma validación de dominio).

**Alternativas descartadas.** Un servicio de archivos nuevo para F6 (duplicaría la regla "solo
contenido publicado" y el volumen); servir todas las imágenes subidas (rompe FR-012); un endpoint
de descarga por módulo (el HTML tendría que saber de qué módulo es cada archivo); exigir imagen
(obligatoria en la spec v1, opcional desde Q4).

---

## R6-8. Superficie HTTP y caché: dos rutas públicas, seis de panel, y la navegación bebe del bloque

**Pregunta.** ¿Qué rutas públicas y de panel, qué política de caché y **con qué respuesta se
decide si la sección existe** (entrada de navegación y mapa del sitio)?

**Decisión.** Convención de F3/F4 (R3-13/R4-7), segmentos en español donde el dominio lo pide
(§5.1):

| Superficie | Rutas | Protección |
|---|---|---|
| Pública (visitante) | `GET /api/v1/ministerios?lang=es\|en` (sección completa) | ninguna (solo lectura); **`Cache-Control: no-store`** |
| Pública (portada) | `GET /api/v1/portada/ministerios?lang=es\|en` (bloque, tope 5) | ninguna; **`no-store`** |
| Pública (archivos) | `GET /api/v1/media/{fileName}` | ninguna; solo archivos de contenido publicado (R6-7) |
| Panel — ministerios | `GET/POST /api/v1/admin/ministerios`, `GET/PATCH/DELETE /api/v1/admin/ministerios/{id}`, `POST /api/v1/admin/ministerios/imagenes` | `AdminChain("ministerios")` (sesión + guard + CSRF + permiso `ministerios`) |

- **Ruta del SPA**: **`/ministerios`** (FR-006, fijada por la spec; coherente con `/eventos` de
  F4), con una única entrada de navegación «Ministerios».
- **Un documento público por superficie**, ambos con `{lang, items?}` (la clave `items` se **omite
  cuando está vacía**, patrón de F3/F4: SC-012/FR-005 —sin claves vacías— y nunca `null`).
- **Presencia de la navegación = el documento del bloque** (`GET /api/v1/portada/ministerios`):
  `useSiteLinks` consume esa misma consulta (compartida con el bloque de la portada vía TanStack
  Query) y la entrada «Ministerios» existe **solo si `items` trae algo** (barra, menú hamburguesa y
  mapa del sitio a la vez, fuente única). Es **equivalente** a mirar la sección porque ambos
  miran el **mismo predicado "publicado"** (el bloque es el mismo listado recortado a 5): 0
  peticiones extra en la portada —la página más visitada— y ninguna forma de que la navegación
  prometa una sección vacía (RG6-2 documenta la coherencia).
- **Sin paginación ni filtros públicos** (FR-004: el visitante no hace nada para verlo todo); el
  público se acota con la constante de seguridad del service `maxPublicItems = 200` (misma idea que
  `maxBlockItems` de F4, RG6-12). El panel **sí** pagina (`limit`/`offset` de `platform/paginate`,
  §8.1.3) con filtro opcional `publicationState`.
- `no-store` en ambos públicos garantiza SC-003 (un cambio guardado se ve en la primera carga
  posterior). El registro de `error.code` **no crece**: `invalid`, `unauthenticated`, `forbidden`,
  `not_found`, `conflict`, `internal`, todos ya en el enum.

**Alternativas descartadas.**

| Alternativa | Por qué no |
|---|---|
| Un solo endpoint público para sección y bloque | La sección necesita todos los ministerios y el bloque solo 5 **resueltos en el servidor** (FR-027: el tope no es configurable ni del cliente) |
| Presencia de navegación desde el documento completo (`GET /api/v1/ministerios`) | Cada página pública descargaría la lista entera (descripciones ≤ 2.000 caracteres) solo para decidir si pintar un enlace; el bloque ya está en caché en la portada |
| Un endpoint "presencia" (`GET /api/v1/ministerios/exists`) | Una tercera ruta pública para un booleano que ya devuelven las otras dos |
| `GET /api/v1/portada/ministerios` dentro de `GET /api/v1/portada` (F3) | Mezclaría dominios y tocaría el contrato cerrado de F3 (mismo criterio que P4-18 de F4) |
| Derivar los 5 del bloque en el cliente desde la lista completa | FR-027 exige el tope **en el servidor**; además la portada pagaría la lista completa |

---

## R6-9. Bilingüismo: `*_es`/`*_en` por campo, `?lang` con fallback, y el enlace que no se traduce

Sin cambios sobre F3/F4 (R3-1/R3-2/R4-3):

- **Guardado**: `*_es` obligatorio **solo si el campo se rellena** (FR-007/Q6: "un campo vacío no
  obliga a nada"), `*_en` opcional; `""`/espacios → `NULL` (normalización I6, trim en el service).
  **Traducibles**: `name`, `description`, `leader_name`, `schedule`, `place` e `image_alt`.
  **No traducibles**: `whatsapp_url` (FR-008: "es un único destino igual en ambos idiomas"),
  `image_file`, `leader_phone`, `leader_email` y `leader_visible`.
- **Regla por campo**: si un campo opcional se rellena en **algún** idioma, su versión **en español
  debe existir** (`schedule_en` sin `schedule_es` → `400 invalid` con `details.scheduleEs`; mismo
  `CHECK` en la BD). `description_en`/`name_en` dependen de `*_es`, que son obligatorios.
- **Lectura pública**: `?lang=es|en` (por defecto `es`, Decisión 8) resuelve cada campo en el
  servidor con `en si existe, si no es` (R3-2): **nunca** un campo vacío por falta de traducción
  (SC-006). El cliente público no aplica fallback (analyze I2 de F3).
- El **panel va en español** (FR-018, Decisión 6 aplicada al sitio público): sin `i18n/` en las
  pantallas de gestión.

---

## R6-10. Validación de entrada y límites (FR-014)

Toda validación en el **backend** (§IV, CWE-20) aunque el formulario también valide:

- **Tags existentes** de `platform/validate` (sin dependencias): `required`, `omitempty`, `max`,
  `email`, `phone`, `url`. **Ningún tag nuevo**.
- **Reglas de dominio** en el service (constantes revisables, en línea con F4):

  | Campo | Límite | Origen del límite |
  |---|---|---|
  | `nameEs`/`nameEn` | 160 | como `title` de F4 y `name` de F3 |
  | `descriptionEs`/`descriptionEn` | **2.000** | default revisable fijado por la spec |
  | `leaderNameEs`/`leaderNameEn` | 160 | como `organizer` de F4 |
  | `leaderPhone` | 7–32 | mismo `CHECK` que `home_contact.phone` |
  | `leaderEmail` | 3–254 | mismo `CHECK` que `home_contact.email` |
  | `scheduleEs`/`scheduleEn` | 200 | texto libre corto («Todos los viernes, 7:00 p. m.») |
  | `placeEs`/`placeEn` | 200 | como `place` de F4 |
  | `whatsappUrl` | 500 | como `map_url` de F4 |
  | `imageAltEs`/`imageAltEn` | 300 | como `image_alt` de F3/F4 |
  | `sortOrder` | `INTEGER` (rango int32) | precedente `sort_order` de F3/F4 |

- **Obligatorios**: solo `nameEs` y `descriptionEs` (FR-014: "nada más es obligatorio").
- **Coherencia del encargado**: `leaderPhone` y `leaderEmail` nunca juntos (R6-4); cualquiera de
  los dos exige `leaderNameEs`; `leaderNameEn` exige `leaderNameEs`.
- **Imagen**: `imageFile` con el patrón del nombre generado; `imageAltEs` obligatorio si hay
  imagen (y al quitar la imagen se limpia el alt).
- **Ante datos inválidos**: `400 invalid` con `details` por campo y **sin guardado parcial** (el
  `PATCH` es un solo `UPDATE` en una transacción con su auditoría).
- Contenido tratado **siempre como datos** (§IV, CWE-79): texto plano multilínea, emojis y tildes
  intactos, sin HTML interpretado.

---

## R6-11. Duplicados exactos → `409`; enlace de WhatsApp compartido → **aviso** no bloqueante

**Pregunta.** Edge cases de la spec: "un duplicado exacto (mismo nombre, misma descripción, mismo
encargado y mismo enlace) se rechaza o se señala (default revisable)"; "dos ministerios con el
mismo enlace de grupo (grupo compartido) se señala como aviso pero se permite (default revisable)".

**Decisión** (defaults aplicados, revisables por el humano):

- **Duplicado exacto → `409 conflict`** ("se rechaza", la opción que la spec pone primero y la que
  usan F2/F3/F4, R4-9): comparación de valores **normalizados** (`btrim` + minúsculas) de
  `name_es` + `description_es` + `leader_name_es` + `whatsapp_url` (con `IS NOT DISTINCT FROM` para
  los opcionales) vía `ExistsMinistryDuplicate` en la transacción del guardado. No hay índice único
  (la normalización es de capa de service; un `UNIQUE` sobre expresiones no aporta al volumen real).
- **Enlace compartido → se permite y se señala**: el DTO del panel `MinisterioAdmin` lleva
  **`whatsappUrlShared: boolean`** (calculado en la consulta con `EXISTS` sobre el mismo
  `whatsapp_url` en otra fila) y el panel muestra un `Notice` **no bloqueante** («Este enlace ya lo
  usa otro ministerio»). Así el "aviso" vive en el servidor —visible también en la lista y en la
  edición, no solo en el momento de guardar— y la API forzada sigue permitiendo el guardado, tal
  como pide el default.

**Alternativas descartadas.** `409` también para el enlace compartido (la spec dice "se permite");
solo señalizar en el formulario antes de guardar (la API forzada quedaría sin señal); un campo
`warnings[]` en la respuesta de escritura (rompería el sobre de éxito = DTO directo y solo se vería
una vez); `UNIQUE (whatsapp_url)` (prohibiría el grupo compartido, contra la spec).

---

## R6-12. Auditoría: 8 códigos `ministry.*` sobre el registro de F2, transaccional y cerrado

**Pregunta.** FR-017: crear, editar, **cambiar la visibilidad de la ficha del encargado**, cambiar
el estado, **reordenar** y eliminar deben quedar en la auditoría de F2 (solo lectura).

**Decisión.** Se repite literalmente el patrón de F3/F4 (R3-11/R4-10): `InsertAdminAction`
**dentro de la transacción** de cada mutación (fail-closed: si el registro falla, la operación **no
se aplica**), `target_kind='content'`, `target_label` descriptivo y registro **cerrado** por
`CHECK` en la BD (migración `000011`). Los **8 códigos** nuevos en `platform/audit`:

| Situación | Código | Etiqueta del panel (`ACTION_LABELS`) |
|---|---|---|
| Alta (nazca en `draft` o `published`) | `ministry.create` | Creó un ministerio |
| Edición de datos (incluye `sortOrder`) | `ministry.update` | Modificó un ministerio |
| Solo cambio de estado | `ministry.publish` / `ministry.unpublish` | Publicó / Retiró un ministerio |
| Solo cambio de visibilidad de la ficha del encargado | `ministry.leader.show` / `ministry.leader.hide` | Mostró / Ocultó la ficha del encargado |
| Borrado | `ministry.delete` | Eliminó un ministerio |
| Subida de imagen | `ministry.image.upload` | Subió la imagen de un ministerio |

Regla exacta (misma que F4, analyze M5): **datos y estado/visibilidad en el mismo guardado → dos
filas** (`ministry.update` + `ministry.publish`/`unpublish` y/o `ministry.leader.show`/`hide`);
**reordenar solo** se registra como `ministry.update` (el orden es un dato del elemento, igual que
en F3/F4 — `home.schedule.update` no tenía código de reorden—; si el humano prefiere un
`ministry.reorder` propio, es un código más en la misma lista). `target_label` =
`Ministerios · <nameEs>` / `Ministerios · Imagen · <fileName>` (máx. 254, `CHECK` de F2). Éxito en
la transacción; `failure` (DTO inválido, duplicado) y `denied` (sin permiso) best-effort. La
resolución método+ruta → acción vive en el dominio `ministerios` (implementa `audit.Recorder`).

**Alternativas descartadas.** Registro de auditoría propio para F6 (FR-017 pide el de F2);
registrar tras el commit (rompe "no aplicarse sin registro"); un solo código `ministry.update`
para todo (no distinguiría quién **ocultó** la ficha del encargado, que es la operación
sensible); `target_kind` nuevo (los elementos de contenido ya son `content`).

---

## R6-13. Publicación por elemento y sección sin contenido: aviso, **sin redirigir**

Patrón R3-3/R4-12 sin cambios: `publication_state TEXT CHECK IN ('draft','published')` por
ministerio; consultas `…Published` separadas para el público (la única función que ve borradores es
la del panel, protegida); el `PATCH` sobre un ministerio publicado guarda y **conserva** el estado
(FR-013: "se publica al guardar"); el paso por borrador solo existe al crear.

**Sección sin nada publicado** (FR-005/SC-010): la entrada «Ministerios» **no existe** en la barra,
el menú hamburguesa ni el mapa del sitio (R6-8, fuente única `useSiteLinks`) y el bloque de la
portada **no se renderiza** (FR-028). **La URL directa `/ministerios`** aplicará el patrón firme
que F4 ya cerró para `/eventos` (R4-13, puerta 2 de F4: **aviso de sección vacía sin redirigir**,
`HiddenNotice`, con el sitio completo alrededor): es la red de seguridad para una URL tecleada o
guardada en favoritos y evita "pantallas sin explicación". *Es un punto que se reseña al humano en
la puerta 2 porque la spec de F6 no lo describe literalmente; la alternativa (404 o redirección a
`/`) quedó descartada en F4 por el mismo motivo.*

**Alternativas descartadas.** Redirigir a `/` (descartado en F4: una URL que "no lleva a ninguna
parte" sin explicación confunde, sobre todo a adultos mayores); página vacía sin mensaje
(contradice SC-010: 0 pantallas sin explicación); renderizar borradores para previsualizar
(FR-012: "por ninguna vía").

---

## R6-14. Bloque de ministerios en la portada: endpoint propio, tope 5, oculto sin publicados

**Contexto.** FR-027…FR-029 (Q7): la portada muestra un bloque con el patrón del bloque «Lo que
viene» de F4, **inmediatamente después del bloque de eventos y antes de «quiénes somos»**, con
**hasta 5** ministerios publicados en el orden manual (`sort_order, id`), en tarjetas con su
imagen y con un botón a `/ministerios`. El tope **5** y el orden de presentación son firmes; la
composición visual la fija la fase de UX (`ux.md`).

**Decisión.** Endpoint público propio **`GET /api/v1/portada/ministerios?lang=es|en`**, registrado
por el dominio `ministerios` (mismo gesto que F4 registró `/api/v1/portada/proximamente`:
Complexity 8 de F4 — el segmento `portada` es el agregado de F3 pero el recurso es de este módulo
—, sin tocar ninguna operación de F3):

- **Tope 5 firme en el servidor** (FR-027): la consulta `ListPublishedMinistries` con
  `ORDER BY sort_order, id` y `LIMIT 5` (constante `maxPortadaItems = 5` del service). **Sin
  parámetro de límite**: el 5 es decisión de la spec, no un default configurable.
- **Solo publicado** (mismo predicado que la sección: no hay fechas en F6 — R6-1/Q6 —, así que no
  hay regla "próximo" que aplicar) y **mismo orden manual** que la sección.
- **DTO mínimo** (`ItemMinisterioPortada`): `id`, `name`, `schedule`/`place`
  opcionales e `image` (`ImagenPublica` con `url` + `alt`) opcional — la tarjeta compacta del bloque
  es "imagen + nombre (+ horario/lugar cuando existen), con el detalle en `/ministerios`"
  (composición de `ux.md` §4.9 sobre la idea de Q7: "las tarjetas de ministerio son más ligeras que
  las fichas completas"). Sin descripción, encargado ni enlace: esos datos viven en la sección
  completa. **El DTO del bloque no tiene campos de encargado**, por lo que la privacidad de FR-026
  es estructural en esta ruta (R6-2).
- **Sin ministerios publicados el bloque NO se renderiza** (FR-028/SC-016): la respuesta es
  `{"lang":"…"}` (`items` omitida) y la portada **no pinta nada** — ni bloque vacío ni mensaje.
  **Diferencia explícita y consciente con el bloque «Lo que viene» de F4**, que sí muestra un
  aviso: aquí manda la regla general de F3 de ocultar secciones vacías (decisión del humano, Q7).
- **Ubicación**: `MinisteriosBlock` se inserta en `HomePage` **inmediatamente después de
  `ProximamenteBlock` y antes de `WhoWeAreSection`** (el orden acordado). Es **una inserción**: el
  resto de la portada de F3 no se toca (FR-021/FR-029).
- Con 1–4 ítems se muestran exactos (sin huecos ni marcadores); con más de 5, los 5 primeros sin
  paginación ni carrusel (SC-016).

**Alternativas descartadas.**

1. **Añadir el bloque a `GET /api/v1/portada` de F3**: tocaría el contrato cerrado de F3 y mezclaría
   dominios (mismo criterio que P4-18).
2. **Derivar el bloque en el cliente** recortando la lista de la sección: FR-027 exige el tope
   **en el servidor**; la portada además descargaría la lista completa para pintar 5.
3. **Mostrar el bloque vacío con un mensaje** como F4: contradice FR-028/SC-016, que exige
   ocultarlo por completo (el humano lo decidió explícitamente en Q7).
4. **Tope configurable desde el panel** (Out of Scope: "montarlo en carrusel, configurarlo desde el
   panel (tope, contenido o posición)").

---

## R6-15. Frontend: F10 aplicado sin cambios; inventario compartido que crece un poco

F6 **aplica** el sistema de diseño y movimiento de F10 (tokens, modo claro/oscuro, movimiento con
`prefers-reduced-motion`, `PublicLayout`, `PublicNav`, `MobileMenu`, `FooterSiteMap`,
`LanguageSwitcher`, `ThemeToggle`) sin redefinirlo, y los patrones de panel de F2/F3/F4:

- **Sección pública `/ministerios`** en `features/publico/ministerios/` (mismo reparto que
  `features/publico/eventos/` de F4), con los nombres de pieza de `ux.md` §6.3:
  `MinisteriosPage` (H1 + grilla de fichas), `MinistryCard` (imagen con `CardImage`, nombre,
  descripción completa, líneas de horario/lugar/encargado **solo con dato**, ficha del encargado
  **solo si viene en la respuesta**), `EncargadoLine` (nombre + contacto cuando existen),
  `WhatsappJoinButton` (**solo si hay `whatsappUrl`**, FR-003), `MinisteriosBlock` (bloque de la
  portada) + `MinistryItem` (tarjeta compacta del bloque: imagen, chip «Ministerio», título-enlace
  a `/ministerios`, horario/lugar si existen — **sin** encargado ni botón de unión) y hooks
  `usePublicMinistries` / `usePortadaMinisterios` (TanStack Query) con claves i18n tipadas nuevas.
- **Navegación**: `useSiteLinks` crece con la entrada **«Ministerios»** (`SiteLinkId` nuevo
  `'ministries'`, `href: '/ministerios'`, clave `nav.sections.ministries`) presente **solo si**
  el documento del bloque trae ítems (R6-8; ver la nota sobre `ux.md` §3 en el plan); la consumen
  a la vez `PublicNav`, `MobileMenu` y `FooterSiteMap` (grupo «Explorar» del pie).
- **Panel** (en español): `features/ministerios/` con los nombres de `ux.md` §6.4:
  `MinistriesPage` (listado con `StatusPill` de estado, `sortOrder` visible y acciones),
  `MinistriesList`, `MinistryForm` (en `Dialog`, pestañas es/en, campos es/en, bloque de encargado
  con `ContactoEncargado` —select «Sin contacto / Celular / Correo» + un solo campo— y
  `ShowContactSwitch` (`role="switch"` con leyendas de privacidad), enlace, horario, lugar, imagen
  con `ImageUploader`, `PublishControls`), hooks `useMinistries`/`useCreateMinistry`/
  `useUpdateMinistry`/`useDeleteMinistry`/`useUploadMinistryImage`, `messages.ts`, `validation.ts`
  (Zod) y `errors.ts` (mapeo de `details` por campo). Sin `EstadoBadge`, `DatesEditor`,
  `MapLink`, `CategoryChip` ni `Pagination` (F6 no tiene fechas, mapas, categorías ni
  paginación visible).
- **Inventario único** (regla de F2/T242): F6 **reutiliza** `Field`, `TextArea`,
  `Select`, `Button`, `Notice`, `ConfirmDialog`, `Dialog`, `EmptyState`, `Table`, `Tabs`,
  `StatusPill`, `ImageUploader`, `PublishControls` y hace **dos promociones** de piezas que hoy
  viven en `features/publico/eventos/` y que F6 comparte: **`CardImage`** (de `ItemCardParts.tsx`)
  e **`HiddenNotice`**, ambas a `src/components/publico/` (mismo gesto que F4 hizo con
  `ImageUploader`/`PublishControls`). `ShowContactSwitch` nace en la feature (si F3/F4/F7 lo
  necesitan algún día, se promueve entonces).
- **Permisos de F2 en el frontend**: `MINISTERIOS` en `lib/permissions.ts` (+ `MODULE_PERMISSIONS`),
  `RequirePermission(MINISTERIOS)` en `/panel/ministerios`, tarjeta «Gestión de ministerios» en
  `AccesosRapidos` de `InicioPage` y `ACTION_LABELS` con las 8 etiquetas de auditoría (R6-12).
- **Tests existentes que se actualizan por añadir el permiso** (mismo gesto que F4 hizo con los
  suyos): `lib/permissions.test.ts` (lista de `MODULE_PERMISSIONS`),
  `features/roles/permissions.test.ts` (`ministerios` deja de ser "Disponible más adelante"),
  `features/roles/roles.test.tsx` (el sello) y `features/panel/panel.test.tsx` (accesos rápidos).

**Alternativas descartadas.** Copiar `ImageUploader`/`PublishControls`/`HiddenNotice` en la feature
(duplicado, contra la regla del inventario); un layout público nuevo (F10 ya lo entregó); resolver
la visibilidad de la navegación en cada componente (fuente única: `useSiteLinks`); un segundo
catálogo de i18n (las claves van en `features/publico/i18n/`, tipadas).

---

## R6-16. Concurrencia: gana la última escritura completa

Igual que F3/F4 (R3-17/R4-11): cada guardado escribe el ministerio entero (incluidos `sortOrder` y
`leaderVisible`) en una sola transacción con su auditoría; dos guardados simultáneos se serializan
en la BD y queda el conjunto completo de un mismo formulario. Sin control optimista (`version`/
`If-Match`) ni historial (fuera de alcance; si el humano lo pide, es una columna + `409`). Un
`PATCH` solo cambia los campos presentes y `null` explícito borra lo opcional (patrón F4).

---

## R6-17. Dependencias nuevas: **ninguna**

Constitución §II: dependencias minimizadas y justificadas. F6 **no añade ninguna dependencia** ni
en backend (`net/http` para el multipart, `net/url` para la validación del enlace, `pgx`/`sqlc`,
`testcontainers-go` ya están) ni en frontend (React 19, React Router, TanStack Query, React Hook
Form + Zod, Tailwind, Vitest + Testing Library + MSW, Playwright).

**Descartado explícitamente.** Librería de validación de URLs o de "link preview" (la stdlib basta:
R6-5); librería de teléfono/correo (los tags `phone`/`email` de `platform/validate` ya existen);
librería de expresiones regulares para el `https` (el `CHECK` de la BD y `net/url` cubren la regla);
cualquier dependencia de UI nueva (el inventario de F10/F2/F3/F4 cubre la ficha y el bloque).

---

## Resumen de decisiones

| # | Decisión | Un vistazo |
|---|---|---|
| R6-1 | Modelo | Una tabla `ministries`; el encargado son **campos del ministerio** (sin tabla `leaders`, sin fichas compartidas) |
| R6-2 | **Privacidad (FR-026)** | Tres capas: DTO público sin campos del encargado oculto + consultas públicas que devuelven `NULL` (`CASE WHEN leader_visible`) + pruebas "0 apariciones" (service/handler/repository/e2e); inventario de vías públicas para `seguridad` |
| R6-3 | Interruptor | `leaderVisible` (equivale al `showContact` de la spec): gobierna solo la publicación, no el guardado; códigos de auditoría propios; **default al crear → puerta 2** |
| R6-4 | Contacto (Q8) | `leader_phone` **o** `leader_email` (nunca ambos: `CHECK` + service), tags `phone`/`email` existentes, contacto exige nombre, validación solo de formato |
| R6-5 | Enlace (FR-015/Q3) | Tag `url` (https + host, sin dominios) + sin espacios/control + `CHECK '^https://[^[:space:]]+$'`; el valor se sirve tal cual |
| R6-6 | Permisos | `AdminChain("ministerios")`; permiso `ministerios` **ya sembrado** en `000002`; **sin migración de permisos y sin tocar F2** |
| R6-7 | Imágenes | `platform/storage` + `internal/media` reutilizados; subida `POST /api/v1/admin/ministerios/imagenes`; `IsFilePublished` crece con `ministries` (el borrador no sirve su imagen) |
| R6-8 | Superficie HTTP | `GET /api/v1/ministerios?lang=` + `GET /api/v1/portada/ministerios?lang=` (tope 5), `no-store`; panel `/api/v1/admin/ministerios` (limit/offset + filtro); la navegación bebe del documento del bloque |
| R6-9 | Bilingüe | `*_es`/`*_en` por campo (es solo si se rellena el campo), `?lang` con fallback en el servidor; `whatsapp_url` no traducible |
| R6-10 | Validación | Tags existentes + reglas de dominio; límites: nombre/encargado 160, descripción 2.000, horario/lugar 200, enlace 500, alt 300; `400 invalid` con `details`, sin guardado parcial |
| R6-11 | Duplicados | Exacto (nombre+descripción+encargado+enlace normalizados) → `409`; enlace compartido → **aviso** (`whatsappUrlShared`) sin bloquear |
| R6-12 | Auditoría | 8 códigos `ministry.*` (migración `000011`), transaccional fail-closed, `target_kind='content'`; visibilidad = `ministry.leader.show/hide`; reordenar = `ministry.update` |
| R6-13 | Publicación | `publication_state` por elemento, "se publica al guardar"; sección vacía → entrada oculta + `/ministerios` con **aviso sin redirigir** (patrón F4 R4-13) |
| R6-14 | Bloque de portada | `GET /api/v1/portada/ministerios` con **tope 5 en el servidor**, orden `sort_order, id`, tarjetas imagen+nombre (+horario/lugar según `ux.md` §4.9), botón a `/ministerios`, **oculto por completo sin publicados** (a diferencia del bloque de F4) |
| R6-15 | Frontend | F10 aplicado; `/ministerios` + `/panel/ministerios` (nombres de pieza según `ux.md`); `MINISTERIOS` en permisos; `CardImage` y `HiddenNotice` promovidos al inventario; `ACTION_LABELS` +8 |
| R6-16 | Concurrencia | Última escritura completa gana (transacción única + auditoría) |
| R6-17 | Dependencias | **Ninguna nueva** en backend ni frontend |
