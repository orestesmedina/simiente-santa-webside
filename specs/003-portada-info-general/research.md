# Research — F3 Portada e información general

**Fecha**: 2026-10-09 · **Spec**: [spec.md](./spec.md) (aprobada por el humano el 2026-10-09, con las
10 aclaraciones resueltas) · **Plan**: [plan.md](./plan.md)

Este documento resuelve las decisiones técnicas abiertas de F3 con alternativas y motivos. La
numeración es **R3-1…R3-18**, propia de esta funcionalidad, para no colisionar con R1–R23 de
`specs/002-acceso-gestion-usuarios/research.md` (F2) ni con D1–D23/D-A1…D-A9 de
`docs/tecnico/decisiones.md`. Donde se reutiliza una decisión de F1/F2 se cita y **no se reabre**.

Contexto reutilizado sin cambios: capas `handler → service → repository` y reglas R1–R8 de
`docs/tecnico/arquitectura.md`; sobre de éxito = DTO directo, sobre de error = `ErrorEnvelope`
(§8.1.2); validación con `platform/validate`; autorización `middleware.AuthzByModule`; auditoría en
`admin_actions` (F2, P20); sesión, CSRF, rate-limit y paginación de F2.

---

## R3-1. Patrón de contenido bilingüe: **columnas `*_es` / `*_en` por campo traducible**

**Pregunta.** Decisión 8 del roadmap: español es el idioma base (obligatorio) e inglés opcional por
contenido. ¿Cómo se guarda? F4–F9 reutilizarán el patrón.

**Decisión.** Una columna por idioma y por campo traducible: `name_es TEXT NOT NULL`,
`name_en TEXT NULL`, `text_es`, `text_en`… Los campos **no traducibles** (URLs, teléfono, correo,
destino de WhatsApp, `day_of_week`, `start_time`/`end_time`, archivos de imagen) van **sin** sufijo.
La regla "español obligatorio, inglés opcional" (FR-008, Decisión 8) es una propiedad de esquema:
`*_es NOT NULL` + `*_en NULL`. **Normalización de lo opcional** *(fijada por el `analyze` I6)*: todo
campo `*_en` llega como `""` o solo espacios → se guarda como **`NULL`** (trim en el service; nunca
`''`, que los `CHECK` de longitud —`BETWEEN 1 AND …`— rechazan). Es la única forma de representar
"sin traducción" y la usan a la vez el panel (US2 esc. 5: dejar el campo en inglés vacío y poder
guardar), el contrato (`""` aceptado y normalizado a `null`) y el público (un `NULL` dispara el
fallback `en → es`, R3-2).

**Alternativas descartadas.**

| Alternativa | Por qué no |
|---|---|
| Tabla de traducciones tipo EAV (`translations(entity_type, entity_id, locale, field, value)`) | Flexible para N idiomas, pero pierde tipos y `CHECK` de longitud por campo en la BD, obliga a un join/merge por cada lectura, hace ilegibles las consultas sqlc y complica la validación "es obligatorio". Para 2 idiomas fijos la flexibilidad no se paga. |
| Tabla hija por idioma (`home_service_translations(service_id, locale, …)`, `UNIQUE(service_id, locale)`) | Bien normalizada y preparada para N idiomas, pero duplica tablas y joins para resolver un caso de exactamente dos idiomas donde el segundo es opcional; el fallback exigiría un `LEFT JOIN`/`COALESCE` en cada consulta. |
| Guardar JSON `{"es": …, "en": …}` en una columna | Sin validación por idioma en BD, sin índices ni `CHECK` útiles, y sqlc no lo tipa. |

**Por qué la decisión.** La Decisión 8 fija el conjunto de idiomas (es + en) y que el inglés sea
**opcional por contenido**: con columnas, `es` obligatorio y `en` opcional quedan garantizados por la
BD (`NOT NULL` / `NULL`), el fallback es un `COALESCE`/paso en el service, las consultas sqlc siguen
siendo planas y la validación de F2 (`platform/validate`) se reutiliza tal cual sobre DTOs planos
(`nameEs`, `nameEn`). **Riesgo aceptado**: si algún día entra un tercer idioma (decisión de roadmap,
no de F3) hay que migrar a tabla hija por idioma; el cambio queda acotado a las tablas de contenido
y a los DTOs, sin tocar la API pública (que ya resuelve idioma en el servidor, R3-2).

---

## R3-2. La API pública resuelve el idioma (`GET /api/v1/portada?lang=…`) con fallback en el service

**Pregunta.** ¿La API devuelve ambos idiomas y el frontend elige, o el servidor resuelve el idioma
pedido?

**Decisión.** El servidor resuelve. `GET /api/v1/portada?lang=es|en` (por defecto `es`) devuelve los
contenidos **ya localizados**: cada campo traducible se resuelve con la regla
`en si existe, si no es` (FR-009) en el service, de modo que **nunca** sale un campo vacío por falta
de traducción (SC-006) sin que el frontend tenga que reimplementar nada. El cambio de idioma del
selector vuelve a pedir la portada con el otro `lang` (TanStack Query, clave de caché por idioma).

**Alternativas descartadas.** Devolver `{es: …, en: …}` y resolver en cliente (duplica la regla de
fallback en cada consumidor — F4–F9 la repetirían —, dobla el tamaño de la respuesta y permite que
una vista muestre huecos si el cliente falla); cabecera `Accept-Language` (opaca para el selector
explícito y difícil de cachear; además la spec dice que **no** se analiza el idioma del navegador);
un endpoint por idioma (`/portada/en`, `/portada/es`) (más rutas para lo mismo).

**Además.** `lang` solo admite `es|en` (cualquier otro valor → `400 invalid`); la respuesta lleva
`lang` para que el cliente sepa qué resolvió y `Cache-Control: no-store` (R3-13). **Sin fallback en
el cliente** *(fijado por el `analyze` I2)*: el sitio público consume strings ya resueltos y **no**
vuelve a aplicar la regla `en → es` (una segunda fuente de fallback podría divergir y sus pruebas no
probarían nada real). Si hace falta un helper para mostrar pares `{es, en}` —p. ej. la previsualización
de los formularios del panel, que sí editan ambos idiomas— vive en `features/informacion/` y solo se
usa ahí.

---

## R3-3. Estado de publicación **por elemento** y secciones que se ocultan

**Pregunta.** Decisión 4 + Q7/Q8/Q10: borrador/publicado por elemento, publicación al guardar y
secciones sin elementos publicados ocultas. ¿Cómo se modela y dónde se decide la ocultación?

**Decisión.** Cada fila de contenido lleva
`publication_state TEXT NOT NULL DEFAULT 'draft' CHECK (publication_state IN ('draft','published'))`
(identidad, quiénes somos, contacto, servicios, canales de WhatsApp y redes: **todos**, FR-013). En
particular, **identidad, «quiénes somos» y contacto son cada uno un elemento publicable por separado**:
una sección = un elemento *singleton* con su propio estado (no existe un estado agrupado de "sección"
ni de página); retirar uno no afecta a los demás. Las consultas **públicas** filtran
`publication_state = 'published'` en SQL (una sola vía de acceso a los datos → el borrador no puede
filtrarse por ninguna otra ruta, FR-013/SC-002). El service público **omite por completo** las
secciones sin elementos publicados (FR-013/Q10) y el frontend solo renderiza las secciones presentes
(SC-012: 0 secciones vacías). `PATCH` sobre un elemento ya publicado guarda los datos y **conserva**
el estado: los cambios son visibles desde la primera carga posterior (FR-014/SC-003, "se publica al
guardar"); el paso por borrador solo existe al **crear**.

**Alternativas descartadas.** Estado por sección (contradice Q7, que pide elemento a elemento; y los
singletons de identidad/quiénes somos/contacto son elementos, cada uno con su estado); columna
`published_at` + programación (fuera de alcance: sin calendario de publicación); estado en el
frontend (la autoridad de qué es público debe estar en el servidor); devolver el borrador con un flag
y que el cliente lo oculte (filtraría contenido en borrador hacia el público, prohibido por FR-013).

---

## R3-4. Singleton de identidad / quiénes somos / contacto: columna `singleton` + upsert serializado

**Pregunta.** Identidad, «quiénes somos» y contacto son únicos (Key Entities). ¿Cómo se garantiza una
sola fila?

**Decisión.** Tablas `home_identity`, `home_about` y `home_contact` con
`singleton BOOLEAN NOT NULL DEFAULT TRUE CHECK (singleton)` y `UNIQUE (singleton)`: el `CHECK`
obliga a que la fila (si existe) valga `TRUE` y el `UNIQUE` garantiza **como máximo una fila**. El
PUT del panel es un **upsert** en transacción: `INSERT … ON CONFLICT (singleton) DO UPDATE …
RETURNING`, precedido de `pg_advisory_xact_lock` por tabla (mismo patrón que el guard anti-bloqueo
de F2, P7) para que dos primeras escrituras simultáneas no compitan. Los campos de negocio conservan
sus `NOT NULL` y `CHECK` normales (a diferencia de sembrar una fila vacía, que los debilitaría).

**Alternativas descartadas.** Sembrar la fila en la migración con id fijo y solo `UPDATE` (una fila
con strings vacíos obligaría a relajar los `NOT NULL`/`CHECK` de los campos de negocio y dejaría un
estado "a medio crear" en la BD); índice parcial `UNIQUE ((true))` (funciona, pero el conflict target
`ON CONFLICT ((true))` es críptico y poco legible en sqlc); dejarlo al service sin guardia de BD
(dos peticiones simultáneas crearían dos filas).

---

## R3-5. Horario de servicios: campos estructurados (`day_of_week`, `start_time`, `end_time` opcional)

**Pregunta.** Q3: cada servicio tiene día, hora, nombre/descripción y lugar, y sus textos son
traducibles. ¿Día y hora como texto o como valores?

**Decisión** *(fijada tras el `analyze` C2 del 2026-10-09; `ux.md` se alinea a ella — D-3 queda
retirada)*. `day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6)` (0 = domingo …
6 = sábado), `start_time TEXT NOT NULL CHECK (start_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$')`
("HH:MM", 24 h) y **`end_time TEXT NULL`** con el mismo patrón y
`CHECK (end_time IS NULL OR end_time > start_time)`: permite rangos tipo «10:00 a. m. − 12:00 m.»
y el fin es **opcional** (un servicio puede no tener hora de cierre). Día y hora **no** son
traducibles: el día se elige en un **selector con las 7 opciones localizadas por i18n** (el número es
el dato; la traducción es de interfaz, nunca texto libre) y la hora se captura y se enseña como
valor "HH:MM" (el formato a.m./p.m. es presentación del cliente). Los textos que sí son traducibles
(nombre/descripción y lugar) van con el patrón `*_es`/`*_en` (R3-1). `sort_order INTEGER NOT NULL
DEFAULT 0` fija el orden de la lista (por defecto: `sort_order, id`).

**Alternativas descartadas.** `day_es`/`day_en` como texto libre ("Domingos", "dom"; era la
propuesta de `ux.md` D-3: permite el mismo día en filas distintas, imposible de ordenar ni de
traducir sin tocar datos y sin validación FR-015); hora como texto libre "10:00 a. m. − 12:00 m."
(el rango se modela mejor con `start_time` + `end_time` y el formato lo localiza la interfaz);
`start_time` solo, sin `end_time` (no admite el rango que la UX quiere mostrar); columna
`day_time TIMESTAMPTZ` (mezcla fecha y hora para un horario semanal recurrente); tipo `TIME` de
PostgreSQL (correcto semánticamente, pero obliga a convertir `pgtype.Time` ↔ cadena "HH:MM" en el
repository y en el contrato sin ganancia visible: el valor se muestra, no se calcula; el `CHECK` por
regex da la misma garantía de formato).

---

## R3-6. Canales de WhatsApp: `kind` (direct/group) + destino normalizado + duplicado literal

**Pregunta.** Q4: números para mensaje directo y/o enlaces de grupo, varios canales, nombre/propósito
y un único destino. Edge case: canal exactamente duplicado.

**Decisión.** Tabla `home_whatsapp_channels` con `kind TEXT CHECK (kind IN ('direct','group'))` y
`destination TEXT`:

- `kind='direct'` → el destino es un **teléfono** validado con el mismo criterio de F2 (etiqueta
  `phone` de `platform/validate`: dígitos con espacios/guiones/paréntesis, `+` opcional y ≥7
  dígitos, FR-015). Se normaliza a solo dígitos (`+` conservado) y el enlace público se arma como
  `https://wa.me/<dígitos>`.
- `kind='group'` → el destino es una **URL** `https` cuyo host es `chat.whatsapp.com` o `wa.me`
  (validación de dominio en el service, formato con el nuevo tag `url` de `platform/validate`,
  R3-14). El enlace público es esa URL.

**Duplicado** (edge case de la spec, "se rechaza o señaliza como duplicado — default revisable"):
se implementa **literalmente** lo que la spec define como "canal exactamente igual": mismo
`kind` + mismo destino normalizado + mismo nombre/propósito normalizado → `409 conflict`
(`UNIQUE (kind, destination, name_es)` en la BD sobre valores ya normalizados). Dos canales con el
mismo destino pero distinto nombre se permiten porque la spec solo exige rechazar el duplicado
exacto; si el humano prefiere la regla estricta de F2 (mismo destino = duplicado, sin mirar el
nombre) es un cambio de un `UNIQUE` — quedó anotado como default revisable.

**Alternativas descartadas.** Dos columnas `phone`/`group_url` (una tabla con dos destinos mutuamente
exclusivos y varios `CHECK` cruzados); guardar solo la URL final (se perdería la validación del
teléfono y el panel no podría mostrar el número); validar el destino con una petición HTTP a WhatsApp
(fuera de alcance: la spec no verifica la vigencia de enlaces).

---

## R3-7. Redes sociales: catálogo fijo en la BD, un enlace por red, hosts oficiales validados

**Pregunta.** Q5: catálogo fijo (Facebook, Instagram, YouTube, TikTok, Spotify y equivalentes) con
un solo enlace por red.

**Decisión.** Tabla `home_social_links` con `network TEXT NOT NULL CHECK (network IN
('facebook','instagram','youtube','tiktok','spotify'))` y **`UNIQUE (network)`** (un enlace por red,
garantizado por la BD). El `CHECK` es el catálogo fijo: una red fuera del catálogo se rechaza en la
BD y en el service (`400 invalid` con `details.network`), y un segundo enlace para la misma red
responde `409 conflict`. La URL se valida con el tag `url` (R3-14) **y** con la comprobación de host
del service: cada red admite sus dominios oficiales (`facebook.com`, `instagram.com`, `youtube.com`,
`youtu.be`, `tiktok.com`, `spotify.com`, con subdominios habituales `www.`/`m.`). Incorporar otra red
"equivalente" (p. ej. X) es un ajuste menor acordado con el humano: nueva fila del `CHECK`, un caso
en el mapa de hosts y una etiqueta — se documenta en `data-model.md` y en el contrato.

**Alternativas descartadas.** Catálogo solo como constantes de Go (la BD no impondría el `CHECK` y
dos instancias podrían divergir); tabla `social_networks` con FK (una tabla para 5 valores fijos que
cambian cada pocos años); permitir cualquier URL válida sin mirar el host (el edge case de la spec
pide rechazar "red social repetida o fuera del catálogo", y un enlace `facebook.com` apuntando a
otro sitio es un error frecuente que conviene avisar al guardar, no al navegar).

---

## R3-8. Imágenes (logo y portada): **disco local del backend** con volumen Docker, validación estricta y servicio de archivos propio

**Pregunta.** FR-002/FR-011: el equipo sube el logotipo y la imagen de portada desde el panel.
¿Dónde se guardan los archivos en el MVP (y qué pasa con Docker/CI)?

**Decisión.** **Disco local administrado por el backend**:

- Nuevo plumbing `internal/platform/storage/` (interfaz `Store` con `Save`/`Open`/`Delete` e
  implementación `LocalStore` sobre `os`): sin SQL, sin HTTP, reutilizable por F8 (galería) y F9.
- Directorio configurado con `UPLOAD_DIR` (por defecto `./uploads` en local; en el contenedor
  `/var/lib/simiente/uploads`). `docker-compose.yml` añade el **volumen con nombre**
  `uploads_data:/var/lib/simiente/uploads`, de modo que los archivos sobreviven a
  `docker compose down`/rebuild igual que `pgdata`. El CI no necesita volúmenes: las pruebas usan
  `t.TempDir()`.
- Subida con `POST /api/v1/admin/portada/imagenes` (multipart, `net/http` estándar) y descarga
  pública con `GET /api/v1/media/{fileName}` (R3-13).
- **Política de descarga (decisión del `analyze` C4, 2026-10-09)**: la descarga **solo** sirve
  archivos **referenciados por contenido publicado** — en F3, el logo o la imagen de portada de una
  **identidad `published`** —. Un archivo que no esté referenciado por contenido publicado
  (elemento retirado a borrador, identidad en borrador, subida huérfana o inexistente) responde
  **`404 not_found`**; un nombre que no cumple el patrón responde **`400 invalid`** (M6). Así
  FR-013/SC-002 ("por ninguna vía") se cumple también en el camino de archivos: retirar un elemento
  hace que su imagen deje de servirse sin borrarla (al republicar vuelve a verse, FR-014). La
  comprobación es una consulta por nombre sobre las tablas de contenido (`IsHomeFilePublished`, variante
  `…Published`); F4–F9 la amplían con sus tablas cuando tengan imágenes.
- **Caché: `Cache-Control: no-store` también en imágenes** (trade-off consciente, RG3-8): al
  depender la descarga del estado actual, una caché `immutable` de largo plazo volvería a mostrar la
  imagen de un elemento ya retirado (quien conoció la URL la seguiría viendo). Se sacrifica el
  rendimiento de la caché por cumplir FR-013; para un sitio con 1–2 imágenes el coste es nulo. Plan
  B (solo con aprobación humana): URL **versionada** por publicación + `immutable`, que eximiría de
  la comprobación por petición aceptando la excepción "URL opaca no descubrible".
- **Seguridad del archivo** (constitución §IV; "evitar ejecución de contenido", FR-015):
  - nombre **generado** por el servidor (`img_<uuid>.<ext>`): el nombre del cliente nunca se usa en
    el disco ni en la ruta (elimina *path traversal* y colisiones);
  - tipo detectado por **firma binaria** (`http.DetectContentType`), no por extensión ni por el
    `Content-Type` del cliente: solo `image/jpeg`, `image/png` e `image/webp`;
  - **SVG prohibido** (es XML con scripts: vector clásico de XSS almacenado) y también GIF
    (innecesario para logo/portada y pesado);
  - tope de tamaño con `http.MaxBytesReader` (**8 MB**, configurable `UPLOAD_MAX_BYTES`): `400`
    con el límite si se excede (el registro de códigos de error no crece: R3-13);
  - la descarga responde `X-Content-Type-Options: nosniff`, `Content-Disposition: inline` y
    `Content-Type` del tipo detectado en la subida; el nombre validado por regex
    `^img_[0-9a-f-]{36}\.(jpg|png|webp)$` antes de tocar el disco;
  - reemplazar una imagen borra el archivo anterior (best-effort, con log) para no acumular huérfanos.

**Alternativas descartadas.**

| Alternativa | Por qué no |
|---|---|
| Servicio externo (S3/Cloudinary/Imgix) | Dependencia nueva de runtime, credenciales y red para un MVP que corre en Docker Compose local; la spec deja "despliegue a producción" fuera de alcance. Queda como plan B si el despliegue real exige CDN: la interfaz `storage.Store` es exactamente el punto de cambio. |
| Guardar los bytes en PostgreSQL (`bytea`) | Hinchazón de la base, copias de seguridad caras y streaming pobre; la BD guarda la referencia, el archivo vive en disco. |
| Servirlo con nginx/estático directo desde el volumen | Requiere tocar el frontend/servidor estático y pierde el control de cabeceras y de validación de nombres; además el backend es quien conoce el tipo detectado. |
| Aceptar cualquier imagen que el navegador reproduzca | Un SVG o un HTML renombrado `.jpg` con `Content-Type` manipulado es ejecución de contenido (CWE-79/434); se exige firma binaria y tipos cerrados. |

---

## R3-9. i18n del sitio público: **provider propio con diccionarios tipados** + memoria de idioma en `localStorage`

**Pregunta.** FR-008/FR-010: interfaz pública siempre bilingüe, selector visible, el cambio aplica de
inmediato y se mantiene durante la visita; idioma por defecto español. ¿Librería o mecanismo propio?

**Decisión.** Mecanismo propio mínimo en `frontend/src/i18n/`:

- `LanguageProvider` (React context) + hook `useLanguage()` → `{ lang, setLang, t }`;
- diccionarios `es.ts`/`en.ts` con la **misma clave tipada**
  (`Record<PublicMessageKey, string>`): TypeScript no compila si falta una clave en un idioma —
  garantía en compilación de SC-006 ("100 % de los textos de la interfaz traducidos") que ninguna
  librería da por sí sola;
- `t()` resuelve la clave del idioma activo; **las cadenas de la interfaz** (rótulos, títulos de
  sección, aria-labels) viven ahí; **los contenidos** vienen ya localizados del backend (R3-2);
- el idioma del contenido y el de la interfaz viajan juntos: `setLang` cambia ambos (el mismo
  `lang` del context es el parámetro `?lang=` de la API);
- **memoria**: `localStorage['ss.lang']`. La elección se mantiene al navegar por el resto del
  sitio durante la visita **y entre visitas en el mismo dispositivo** (decisión del humano del
  2026-10-09). **La primera visita sin preferencia guardada se muestra en español** (idioma base,
  Decisión 8). Si el guardado falla (p. ej. navegación privada), el idioma simplemente no persiste
  entre visitas: sin avisos ni errores. **No** se auto-detecta el idioma del navegador.

**Sobre la memoria entre visitas.** La primera versión de este documento eligió `sessionStorage`
(memoria por visita) porque el FR-010 original pedía que "una visita nueva se muestra en español" y
el Assumptions pedía conservar la elección entre visitas: dos lecturas incompatibles. **El humano
resolvió la tensión el 2026-10-09**: la memoria es `localStorage` (la elección **se conserva entre
visitas**) y "visita nueva" se entiende como **primera visita sin preferencia guardada**, que sí se
muestra en español. FR-010 y el Assumptions de `spec.md` se actualizan con esa redacción (lo refleja
el orquestador/analista en la spec; el resto del diseño no cambia).

**Alternativas descartadas.** `react-i18next` + `i18next` (2 dependencias de runtime + ~40 kB para
2 idiomas y ~40 cadenas, sin aportar la exhaustividad tipada que SC-006 exige; la skill
`react-frontend` no la prescribe); JSON cargado por red (una petición más para algo que son 2 kB de
código); resolver el idioma con `navigator.language` (la spec lo descarta explícitamente).

---

## R3-10. Permisos y autorización: se **activa** el permiso `portada` ya sembrado por F2

**Pregunta.** FR-012: toda operación del módulo exige el permiso «portada e información general»
reservado por F2 (FR-015 de F2). ¿Cómo se conecta?

**Decisión (verificado en el código de F2).** La migración `000002_create_roles_and_permissions`
ya siembra `('portada', 'Portada e información general')` en `permissions`: **el catálogo no cambia**
(Decisión 5 del roadmap: catálogo fijo). F3 solo **activa** ese permiso:

- backend: subgrupo `/api/v1/admin/portada` montado con
  `middleware.AdminChain("portada", deps)` — la misma cadena aprobada de F2
  (`authn → guard de cambio de contraseña → authz('portada') → CSRF`), con lo que la denegación es
  idéntica a la de F2 (`403 forbidden`, mensaje genérico que no revela el permiso) y **queda
  registrada** en `admin_actions` con `result='denied'` (R3-11);
- frontend: `PORTADA = 'portada'` en `lib/permissions.ts`, guard `RequirePermission code={PORTADA}`
  en `/panel/informacion`, entrada nueva en el menú filtrado por permisos y **`isPermissionAvailable`**
  deja de marcar `portada` como "Disponible más adelante" (`features/roles/permissions.ts`).

**Alternativas descartadas.** Crear un permiso nuevo (rompería Decisión 5 y el catálogo de F2: la
spec de F3 lo prohíbe explícitamente en Out of Scope); un middleware propio de autorización (ya
existe `AuthzByModule`, probado en F2); comprobar permisos solo en el frontend (viola §IV/CWE-862).

---

## R3-11. Auditoría en el registro de F2: SQL compartido vía `internal/db`, registro cerrado ampliado y resolución de rutas en el dominio F3

**Pregunta.** FR-017: cada edición queda registrada en la auditoría de F2 (quién, qué, cuándo), y el
edge case exige que **una edición aplicada no pueda quedar sin registro ni viceversa** (atomicidad).
Pero las tablas de auditoría las "posee" el dominio `usuarios` y la regla R2 prohíbe que un dominio
importe a otro.

**Decisión.** Cuatro piezas:

1. **SQL compartido, no dominio compartido.** R2 dice literalmente: "si dos dominios necesitan lo
   mismo, eso sube a `internal/platform/` (o **a `internal/db` si es SQL compartido, vía consulta
   nombrada**)". La consulta `InsertAdminAction` ya existe generada en `internal/db` (F2). El
   repository del dominio `portada` la llama **dentro de la misma transacción** de su mutación
   (`withTx` + `gendb.New(tx).InsertAdminAction(...)`), de modo que contenido + registro son
   atómicos (o ambos o ninguno; R23 de F2). Ningún dominio importa a otro y el mapeo
   `audit.Action` → `InsertAdminActionParams` (~10 líneas) se repite en el repository de `portada`
   (ver Complexity Tracking del plan).
2. **El registro cerrado de acciones crece** en `platform/audit` (plumbing compartido, ya define
   `ActionCodes`): 15 códigos nuevos `home.*` (R3-11.1) y el `TargetKind` `content`. La migración
   `000006` extiende los `CHECK` de `admin_actions` (`action`, `target_kind` y la coherencia de
   objetivo) para que la BD siga siendo el registro cerrado. Los objetivos de contenido **no tienen
   FK** (son 6 tablas distintas): van con `target_kind='content'`, `target_user_id`/`target_role_id`
   en `NULL` y `target_label` = `Portada · <Sección> · <nombre del elemento>` (el "qué editó" de
   FR-017; `targetLabel` ya es nullable en el contrato).
3. **Éxito transaccional, fallo y denegación best-effort** (mismo criterio que P20 de F2): la
   mutación exitosa registra en su propia transacción (**fail-closed**: si el registro falla, no se
   aplica la edición — FR-017/edge case); los intentos rechazados (JSON/DTO inválido) y las
   denegaciones de permiso se registran best-effort (`result='failure'`/`'denied'`) sin cambiar la
   respuesta.
4. **La resolución método+ruta → acción vive en el dominio F3.** F2 resuelve sus rutas en
   `usuarios.actionFromRoute`, que no conoce (ni debe conocer) las de F3. El dominio `portada`
   implementa `audit.Recorder` (`RecordDenied`) con su propia tabla de rutas y lo recibe
   `middleware.AdminDeps.Recorder` en **su** subgrupo; la persistencia de esa denegación vuelve a
   `admin_actions` por la consulta compartida (punto 1). Así F4–F9 repiten el patrón sin tocar
   `usuarios`.

**R3-11.1 — Códigos de acción de F3** (registro cerrado; equivalen a "quién hizo qué"):

| Código | Cuándo |
|---|---|
| `home.identity.update` | guardar la identidad (referencias a logo/imagen de portada incluidas) |
| `home.about.update` | guardar «quiénes somos» |
| `home.contact.update` | guardar los datos de contacto |
| `home.schedule.create` / `.update` / `.delete` | alta, edición y borrado de un servicio del horario |
| `home.whatsapp.create` / `.update` / `.delete` | alta, edición y borrado de un canal de WhatsApp |
| `home.social.create` / `.update` / `.delete` | alta, edición y borrado de un enlace de red social |
| `home.image.upload` | subir una imagen (logo/portada) al panel (`analyze` I8: la subida escribe disco y se audita por sí misma) |
| `home.publish` / `home.unpublish` | **solo** un cambio de estado: publicar o retirar un elemento existente (el `targetLabel` dice cuál) |

**Regla exacta de qué código corresponde** *(fijada por el `analyze` M5)*:

- **Alta** (`create`, nazca en `draft` o `published`) → siempre `home.<sección>.create`
  (un alta ya publicada **no** registra `home.publish`).
- **Edición** de datos sin tocar el estado → `home.<sección>.update`.
- **Cambio de estado** sobre un elemento existente → `home.publish` / `home.unpublish`.
- Un `PATCH` que cambia **a la vez** datos y estado deja **dos filas** en la misma transacción
  (`home.<sección>.update` + `home.publish`/`home.unpublish`), de modo que "qué editó" queda
  completo y `home.publish`/`home.unpublish` **solo** significan cambio de estado.
- **Subida de imagen** → `home.image.upload` con `target_label` = `Portada · Imagen · <fileName>`;
  si el registro falla, la subida se aborta y el archivo se elimina (fail-closed, coherente con el
  edge case). La referencia final de la imagen a la identidad se registra aparte con
  `home.identity.update` al guardar.

El detalle valores antes/después **no** se registra: la spec lo deja como default deseable pero
revisable (Assumptions de F3, igual que en F2).

**Alternativas descartadas.** Que `portada` escriba `admin_actions` con SQL propio duplicado (dos
paquetes con SQL de la misma tabla, contra R2); que `usuarios` conozca las rutas de F3 en
`actionFromRoute` (acoplaría F2 a F3–F9 y obligaría a tocar F2 ante cada módulo nuevo); un registro
de auditoría nuevo para contenido (FR-017 pide **el registro de F2**, el mismo de solo lectura);
registrar la edición tras el commit sin transacción (el edge case prohíbe "aplicarse sin registro").

---

## R3-12. Identidad de marca: tokens en el tema del frontend, recursos versionados en el repo

**Pregunta.** FR-002 manda respetar el Manual de Identidad (paleta, tipografías, reglas del
logotipo). ¿Dónde viven los tokens y el logo?

**Decisión.** Sí, se versionan en el repo:

- **Tokens de tema** en frontend: variables CSS en `frontend/src/index.css` expuestas en el tema de
  Tailwind (`tailwind.config`) y familias tipográficas. **Nomenclatura única de tokens (fijada por
  el `analyze` I5: la de `ux.md` §1, la más rica; un solo nombre por token)**:

  | Token | Valor | Equivalente del Manual |
  |---|---|---|
  | `navy` | `#1a2b4a` | azul (70 %) |
  | `navy-soft` | derivado de `navy` | azul suave (soportes, esqueletos) |
  | `teal` | `#00c9a7` | turquesa |
  | `teal-strong` | derivado de `teal` | turquesa intenso (estados) |
  | `white` | `#ffffff` | blanco |
  | `cream` | `#F5F2EC` | crema |
  | `coral` | `#ff6b3d` | naranja (complementario) |
  | `leaf` | `#217638` | verde (complementario) |
  | `--font-display` | `'Bebas Neue'` | títulos |
  | `--font-sans` | `'Poppins'` | texto general |
  | `--font-emotiva` | `'Playfair Display'` | énfasis emocional |

  No existen alias (`--color-azul`, `--font-titulos`… quedan **retirados**): quien escriba CSS usa
  solo estos nombres. La aplicación concreta (composición, jerarquía) la define `disenador-ux` en
  `ux.md`; el plan **adopta** su nomenclatura para que el diseño no invente una paleta paralela.
- **Tipografías autoalojadas** con `@fontsource/bebas-neue`, `@fontsource/poppins` y
  `@fontsource/playfair-display` (dependencias npm **solo de build**): el sitio carga sin depender de
  un CDN de terceros (mejor para conexiones lentas —el público objetivo— y sin llamadas a terceros).
- **Logotipo versionado**: `resources/simiente.jpeg` (entregado por el cliente) se copia a
  `frontend/public/brand/simiente-logo.jpeg` como **recurso de marca del producto** (fallback del
  encabezado, R3-18) y se documentan sus reglas de uso en el componente: **no deformar** (sin
  `width`/`height` que rompan la proporción), **no cambiar colores** (sin `filter`/`mix-blend`),
  **no girar** (sin `rotate`/`transform`) y **sin efectos** (sin sombras sobre el logo). La
  comprobación es de revisión de código (CSS/`className` del componente `BrandLogo`), no de runtime.
- El PDF `resources/MANUAL DE MARCA.pdf` **no** se sirve en el sitio: queda como referencia del
  cliente.

**Alternativas descartadas.** Google Fonts por `<link>` (dependencia de red en cada carga y llamada
a un tercero); guardar los tokens solo en el PDF del manual (el código no puede leer un PDF); subir
el logo solo desde el panel (una instalación recién levantada quedaría sin ninguna marca visible;
el recurso del repo no es contenido editorial y por tanto no contradice FR-013 — ver R3-18).

---

## R3-13. Superficie HTTP: `/api/v1/portada` (pública), `/api/v1/media/{file}` y `/api/v1/admin/portada`; caché

**Pregunta.** ¿Qué rutas y qué política de caché?

**Decisión.** Convención de `arquitectura.md` §5.1 (segmentos en español donde el dominio lo pide):

| Superficie | Rutas | Protección |
|---|---|---|
| Pública (visitante) | `GET /api/v1/portada?lang=es\|en` | ninguna (solo lectura); **`Cache-Control: no-store`** |
| Pública (archivos) | `GET /api/v1/media/{fileName}` | ninguna; **solo** archivos referenciados por contenido **publicado** (R3-8, `analyze` C4) — el resto, `404`; `Cache-Control: no-store` (trade-off de RG3-8: la descarga depende del estado actual, así que no puede ser `immutable`) |
| Panel | `GET /api/v1/admin/portada` y las mutaciones `PUT/PATCH/POST/DELETE` de identidad, quiénes somos, contacto, horario, WhatsApp, redes e imágenes | `AdminChain('portada')` de F2 (sesión + CSRF + permiso `portada`) |

`no-store` en la portada pública garantiza SC-003 ("los cambios guardados son visibles desde la
primera carga posterior"): un navegador no puede atender una portada guardada desde caché. Las
imágenes, en cambio, son inmutables por nombre y sí se cachean agresivamente (rendimiento en móvil).

El registro de `error.code` **no crece** (R3-14): las subidas inválidas responden `400 invalid` con
`details.file`/`details.fileSize`, igual que cualquier validación.

**Alternativas descartadas.** `/api/v1/public/portada` (prefijo `public` no usado por F1/F2: lo
público ya es "lo que no está bajo `/admin`"); servir imágenes desde un bucket público con URL
absoluta (ataría el contrato al proveedor); `ETag`/`must-revalidate` en la portada (complejidad sin
beneficio hoy: el volumen de la portada es un JSON pequeño).

---

## R3-14. Validación de entrada: etiqueta `url` nueva en `platform/validate` + límites de negocio

**Pregunta.** FR-015 pide validar campos obligatorios, correo, teléfono, enlaces de WhatsApp y de
redes, español siempre presente y el límite de «quiénes somos». ¿Dónde vive cada regla?

**Decisión.** Todo se valida en el **backend** (§IV, CWE-20) aunque el frontend también valide:

- `platform/validate` **crece con la etiqueta `url`** (acepta solo `https` con host no vacío,
  máx. 500 caracteres): es la tercera vez que haría falta un enlace (F3: WhatsApp y redes; F4–F9:
  enlaces de eventos) y reutiliza el patrón de tags existente (`required`, `min`, `max`, `email`,
  `oneof`, `phone`). El host concreto (WhatsApp, red social) lo comprueba el **service** del dominio,
  que es quien conoce el negocio (R3-6/R3-7).
- Reutilización literal de los criterios de F2: `email` para el correo de contacto y `phone` para el
  teléfono (FR-015 los exige "con el mismo criterio de F2").
- Español obligatorio: `*_es` con `required` (y `NOT NULL` en la BD); inglés `omitempty` (Decisión 8).
- Límites (constantes del service, revisables por el humano): «quiénes somos» **1.000 caracteres**
  por idioma (FR-003); nombre/lema/misión/visión 160/300/600/600; lugar 200; nombre de canal 120;
  dirección 300; teléfono 32 (como F2); URL 500. Colecciones acotadas: máx. **50** servicios, **20**
  canales de WhatsApp y **1 por red** de redes (el agregado del panel es un documento, no un listado
  paginado — ver plan, Complexity Tracking).
- Texto plano (FR-003/FR-015, constitución §IV): los textos se guardan y se muestran como datos; el
  frontend **no** usa `dangerouslySetInnerHTML` en ninguna vista de F3 y el backend no transforma el
  texto (tildes, `ñ`, `¿¡` y emojis se conservan intactos — validado con casos de prueba).

**Alternativas descartadas.** Validar URLs solo en el frontend (§IV lo prohíbe); una librería de
validación de URLs (la stdlib `net/url` + el tag basta); límites en la BD solamente (el mensaje al
humano sale mejor del service con `details` por campo).

---

## R3-15. La portada reemplaza `/` como página de inicio; el estado del sistema pasa a `/health`

**Pregunta.** FR-001 dice que la portada **es** la página de inicio del sitio público. Hoy `/`
muestra `StatusPage` (la pantalla de F1 que comprueba que backend y BD están vivos).

**Decisión.** `/` pasa a ser la portada pública (layout público, F3) y `StatusPage` se mantiene
disponible en **`/health`** (misma pantalla, sin cambios; es una herramienta de diagnóstico útil en
local y para QA — ruta de la SPA, distinta del endpoint de backend `/healthz`, que no se toca:
constitución §VII). La prueba de humo de `cmd/api` y los e2e de F2 que apuntaban a `/` se ajustan en
sus tareas.

**Alternativas descartadas.** Borrar la pantalla de estado (pierde utilidad de diagnóstico y F1 la
certificó); dejar la portada en otra ruta (`/inicio`) y que `/` siga siendo estado (contradice
FR-001); `/estado` (primera elección de este plan; el humano la cambió a `/health` el 2026-10-09 por
coherencia con el vocabulario operativo del proyecto).

---

## R3-16. Dependencias nuevas: **ninguna de runtime en backend**; 3 de tipografías en frontend

Constitución §II: toda dependencia nueva se justifica. F3 **no añade ninguna dependencia de runtime
al backend**: `net/http` (multipart, `ServeContent`), `net/url`, `mime`/`http.DetectContentType`,
`os` y `pgx`/`sqlc` ya existen. Las pruebas reutilizan `testcontainers-go` de F2.

Frontend añade **solo** las tres tipografías del Manual de Identidad:
`@fontsource/bebas-neue`, `@fontsource/poppins`, `@fontsource/playfair-display` (R3-12) — son assets
de build, sin código de ejecución y sin llamadas de red a terceros. Todo lo demás se reutiliza:
React 19, React Router, TanStack Query, React Hook Form + Zod (formularios del panel), Tailwind,
Vitest + Testing Library + MSW y Playwright.

**Descartado.** `react-i18next`/`i18next` (R3-9); `sharp`/librerías de procesamiento de imágenes (el
MVP guarda el archivo tal cual: el navegador lo escala; no se generan derivados — fuera de alcance);
cliente S3 (R3-8); librería de validación de URLs (R3-14).

---

## R3-17. Ediciones simultáneas: **gana la última escritura completa** (sin mezclar datos)

**Pregunta.** Edge case: dos personas editando la información general a la vez; default revisable de
la spec: "se conserva la última versión guardada completa".

**Decisión.** Cada guardado es una escritura **completa del elemento** dentro de una transacción
(`PUT` de singletons: reemplazo completo; `PATCH` de un elemento de lista: campos presentes, una
sola fila): dos guardados simultáneos se serializan en la BD y el último commit es el que queda, con
el conjunto **completo** de un mismo formulario — nunca una mezcla de dos formularios. El singleton
va además precedido de `pg_advisory_xact_lock` (R3-4). No hay control de concurrencia optimista
(`version`/`If-Match`) ni historial (fuera de alcance, Assumptions de la spec): si el humano lo pide
más adelante, es una columna `version` + `409` — cambio acotado.

---

## R3-18. Identidad en borrador y el "chrome" del sitio público

**Pregunta.** Si la identidad (nombre, logo) está en borrador o aún no existe, ¿qué muestra el
encabezado del sitio público? (FR-013 prohíbe exponer **cualquier** contenido en borrador.)

**Decisión.** La identidad es un elemento con su `publication_state` como todos los demás (FR-013 es
literal: "cada elemento"). El **chrome** del layout público (encabezado y pie) usa, cuando la
identidad está publicada, su nombre y logotipo; cuando no, un **recurso estático del producto**: el
nombre "Iglesia Simiente Santa" del diccionario i18n y el logotipo versionado en el repo
(`frontend/public/brand/simiente-logo.jpeg`, R3-12). Ese fallback **no** es contenido editorial del
CMS ni procede de la base de datos: no hay riesgo de filtrar un borrador (SC-002). La sección
"identidad" completa (lema, misión, visión, imagen de portada) solo se renderiza si está publicada.

**Alternativas descartada.** Mostrar siempre la identidad aunque esté en borrador (viola FR-013/SC-002
de forma literal); un encabezado sin nombre ni logo cuando falta la identidad (una portada sin marca
es un error de producto evidente para el público objetivo).

---

## Resumen para la puerta de aprobación

| # | Decisión | Un vistazo |
|---|---|---|
| R3-1 | Patrón bilingüe | Columnas `*_es` (obligatorio) / `*_en` (opcional) por campo traducible; `""`/espacios en `*_en` → `NULL`; reutilizable en F4–F9 |
| R3-2 | Localización | El servidor resuelve `?lang=es\|en` con fallback `en → es` por campo |
| R3-3 | Publicación | `publication_state` por elemento (identidad, «quiénes somos» y contacto: cada singleton es un elemento publicable **por separado**); el público solo ve `published`; secciones vacías omitidas; se publica al guardar (FR-014) |
| R3-4 | Singletons | `singleton BOOLEAN` + `UNIQUE` + upsert con advisory lock |
| R3-5 | Horario | `day_of_week` 0–6 (selector localizado, no texto libre) + `start_time` "HH:MM" + `end_time` opcional (rango) + textos `es/en` + `sort_order` |
| R3-6 | WhatsApp | `kind` direct/group, destino normalizado, duplicado literal (kind+destino+nombre) → 409 |
| R3-7 | Redes | Catálogo fijo en `CHECK`, `UNIQUE(network)`, hosts oficiales validados |
| R3-8 | Imágenes | Disco local (`UPLOAD_DIR`) + volumen Docker `uploads_data`; firma binaria; sin SVG; 8 MB; nombre generado; `/api/v1/media/{file}` **solo** sirve archivos de contenido publicado (404 en otro caso) con `no-store` |
| R3-9 | i18n frontend | Provider propio con diccionarios tipados es/en; memoria en `localStorage` entre visitas (primera visita sin preferencia → español; sin auto-detección del navegador; FR-010 ajustado por el humano el 2026-10-09) |
| R3-10 | Permisos | Se activa `portada` ya sembrado por F2; `AdminChain('portada')`; guard `RequirePermission` |
| R3-11 | Auditoría | `InsertAdminAction` compartido en la transacción de la mutación; 15 códigos `home.*` (incl. `home.image.upload`) + `target_kind='content'` (migración 000006); `home.publish/unpublish` solo para cambios de estado; denegaciones resueltas por el dominio F3 |
| R3-12 | Marca | Tokens con la nomenclatura única de `ux.md` (`navy`, `teal`, `cream`, `coral`, `leaf`, `--font-display/sans/emotiva`), `@fontsource` ×3, logo versionado en el repo con sus reglas |
| R3-13 | Rutas y caché | `GET /api/v1/portada` y `GET /api/v1/media/{file}` (`no-store`; la descarga exige contenido publicado), panel en `/api/v1/admin/portada` |
| R3-14 | Validación | Tag `url` nuevo en `platform/validate`; límites de longitud y colecciones acotadas |
| R3-15 | Rutas SPA | `/` = portada; `StatusPage` → `/health` (ruta de la SPA; `/healthz` de backend intacto) |
| R3-16 | Dependencias | 0 nuevas en backend; 3 tipografías `@fontsource` en frontend |
| R3-17 | Concurrencia | Última escritura completa gana (transacción por elemento) |
| R3-18 | Chrome sin identidad publicada | Fallback estático (nombre del diccionario + logo del repo), nunca datos del CMS en borrador |
