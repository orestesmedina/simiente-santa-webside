# Research — F7 Ofrendas y diezmos (Donaciones)

**Fecha**: 2026-10-14 · **Spec**: [spec.md](./spec.md) (aprobada v2 por el humano el 2026-10-14;
Q1–Q11 resueltas) · **Plan**: [plan.md](./plan.md)

> **Notas de partida.** (1) La spec v2 ya cerró Q1–Q11 (decisión 22 del roadmap) y este documento
> las da por firmes: agrupación por tipo con orden manual (Q1), campos y obligatoriedad (Q2),
> IBAN (Q3), SINPE (Q4), uso de las donaciones como texto libre (Q5), sin imágenes (Q6), bloque de
> portada (Q7), nombre y ruta «Ofrendas y diezmos»/`/ofrendas` con módulo y permiso «donaciones»
> (Q8), ocultar y redirigir sin nada publicado (Q9), granularidad de publicación (Q10) y texto
> introductorio (Q11). **Ninguna de ellas se reabre aquí.** (2) Donde se reutiliza una decisión ya
> tomada se cita (R3-x de F3, R4-x/P4-x de F4, R6-x/P6-x de F6) y **no se reabre**. (3) Los puntos
> donde la spec **se contradice** o **no llega** están marcados como tales (R7-5, R7-9, R7-12) y
> recogidos en «Preguntas para el humano» de [plan.md](./plan.md): se aplican los defaults indicados
> y se elevan al humano, **no se inventan reglas de negocio**.

Este documento resuelve las decisiones técnicas abiertas de F7 con alternativas y motivos. La
numeración es **R7-1…R7-16**, propia de esta funcionalidad, para no colisionar con R6-1…R6-17 de
F6, R4-1…R4-16 de F4, R3-1…R3-18 de F3 ni R1–R23 de F2.

Contexto reutilizado sin cambios: capas `handler → service → repository` y reglas R1–R8 de
`docs/tecnico/arquitectura.md`; sobre de éxito = DTO directo / `{items}` en listados (§8.1.2);
paginación `platform/paginate` (§8.1.3); `ORDER BY` determinista (§8.1.7); validación
`platform/validate` (tags `required`, `omitempty`, `min`, `max`, `oneof`, `email`, `phone`, `url`);
autorización `middleware.AdminChain`/`AuthzByModule`; auditoría en `admin_actions` con
`InsertAdminAction` transaccional (R3-11/R4-10/R6-12); bilingüismo por columnas `*_es`/`*_en` con
resolución `?lang=` en el servidor (R3-1/R3-2); `publication_state` por elemento con consultas
`…Published` (R3-3); `sort_order INTEGER NOT NULL DEFAULT 0` con desempate por `id` (patrón de
`000005`/`000007`/`000010`); sección sin contenido → **redirección a la portada** (R6-13/P6-21);
bloque de portada con endpoint propio y tope fijo en el servidor (R6-14).

---

## R7-1. Modelo de datos: **dos tablas** — `donation_methods` (el listado administrable) y `donation_texts` (los dos textos, un renglón por clave)

**Pregunta.** Key Entities define tres cosas editables: el **método de donación** (cuenta IBAN o
número de SINPE), el **texto introductorio** (Q11) y el **texto de uso de las donaciones** (Q5),
cada uno con su propio estado de publicación (Q10). ¿Cómo se persiste?

**Decisión.** Dos tablas (ver [data-model.md](./data-model.md)):

1. **`donation_methods`** — una fila por método de donación (varias cuentas y varios números, Q1):
   `kind`, `number`, `institution`, `holder_es`/`holder_en` (opcional), `currency` (solo IBAN),
   `publication_state` y `sort_order` (por grupo). Es el "listado administrable agrupado por tipo"
   de FR-001/FR-004: la agrupación es por `kind` y el orden manual vive **dentro** de cada grupo.
2. **`donation_texts`** — **un renglón por clave de texto** (`text_key IN ('intro','usage')`,
   `UNIQUE`, catálogo cerrado con `CHECK`): `body_es`/`body_en`, `publication_state` y nada más.
   Los dos textos tienen **exactamente la misma forma** (Key Entities los describe en paralelo:
   "texto libre editable de uno o varios párrafos, con su propio estado de publicación y traducible
   al inglés de forma opcional"), así que comparten tabla; la clave los distingue y su `UNIQUE`
   garantiza ≤1 fila por clave (el patrón "singleton por clave", generalización del singleton de
   `000005_create_home_content`). Las filas **se crean en el primer guardado** (upsert del `PUT`),
   como los singletons de F3 ("son `null` hasta el primer guardado"); **no se siembra contenido**
   en la migración (el contenido lo escribe el equipo, no el esquema).

Por qué no una tabla más de métodos ni una de traducciones: cada método es un elemento con orden,
estado y auditoría propios (merece fila); los textos no tienen vida propia más allá de su clave
(no se listan, no se ordenan, no se reordenan): separarlos en dos tablas idénticas duplicaría
esquema, consultas y validación sin ganancia.

**Alternativas descartadas.**

| Alternativa | Por qué no |
|---|---|
| Dos tablas singleton (`donation_intro`, `donation_usage`) al estilo `home_about` | Misma forma, dos esquemas, dos juegos de consultas y de `CHECK`; el catálogo cerrado `text_key` da las mismas garantías con la mitad de superficie |
| Una sola tabla para todo (métodos **y** textos con columnas anulables por uso) | Mezcla dos agregados con ciclos de vida distintos (listado ordenado vs. singleton) y obliga a `CHECK` cruzados que la BD no puede expresar con claridad |
| Tabla de traducciones / EAV (`*_translations`) | Rompe el patrón R3-1 probado en F3/F4/F6 (columnas `*_es`/`*_en`) y complica sqlc sin ganancia |
| JSONB para los textos | Pierde los `CHECK` de longitud, el `publication_state` por texto y la consulta simple "solo lo publicado" |
| Siembra de las dos filas en la migración | El esquema no debe contener contenido del cliente; además exigiría `body_es` vacío, que contradice `char_length ≥ 1` y abriría el hueco "texto publicado vacío" |

---

## R7-2. Número de cuenta IBAN (Q3/FR-015): estructura y longitud, guardado normalizado, presentación en bloques

**Pregunta.** Q3/FR-015: validar **por estructura y longitud** (costarricense e internacional
válido según la estructura IBAN), guardar "el número exacto normalizado a los caracteres del
número" y presentarlo "agrupado en bloques". ¿Dónde vive la regla y qué significa cada parte?

**Decisión.** Tres piezas, todas sin dependencias nuevas:

1. **Normalización al guardar** (la que exige Q3): se acepta el número tal como lo escribe el
   equipo (con espacios o guiones: `CR01 0000-0000-0000-0000-0`), se **recorta a los caracteres
   del número** (se eliminan espacios y guiones, y se pasa a mayúsculas) y **ese valor normalizado
   es el que se guarda y el que se publica** (FR-015: "el número publicado DEBE ser el número
   válido guardado, sin transformaciones que lo alteren"). La normalización es **determinista** y
   por eso la regla de duplicados (R7-9) puede comparar números con igualdad simple.
2. **Validación por estructura y longitud** (literal de Q3/FR-015):
   - estructura: `^[A-Z]{2}[0-9]{2}[A-Z0-9]{10,30}$` (2 letras de país + 2 dígitos de control +
     el BBAN alfanumérico);
   - longitud: **exacta por país**, según la tabla del registro IBAN (ISO 13616 / SWIFT),
     embebida como constante Go (~90 países; Costa Rica `CR` = 22, España `ES` = 24, Alemania
     `DE` = 22, Reino Unido `GB` = 22, Portugal `PT` = 25…). Un país **fuera del registro** o con
     longitud distinta a la de su país se **rechaza** con `400 invalid` y `details.number`
     ("estructura o longitud no válidas"), que es exactamente la tabla de errores de la spec.
3. **Presentación agrupada en bloques** (Q3/FR-003): la agrupación en bloques de 4 es
   **presentación**, no dato: la hace el **frontend** con un helper puro (`formatIbanDisplay`,
   `features/publico/ofrendas/format.ts`, con pruebas unitarias) que inserta separadores visuales
   cada 4 caracteres **sin alterar el valor**; el **botón de copiar copia `number` tal cual**
   (el valor exacto guardado). Así SC-009 ("íntegro, exacto y copiable… 0 caracteres alterados")
   es verificable campo a campo y el valor copiado coincide siempre con el de la BD.

La regla completa vive en el **dominio** (`internal/donaciones/numero.go`, funciones puras con su
tabla de casos), **no** en `platform/validate`: los tags de `platform/validate` son formatos
genéricos (correo, teléfono, URL) y el IBAN es un formato de negocio de este módulo (mismo criterio
que F6/R6-5 para las reglas de dominio que los tags no cubren). La BD **refuerza** la forma con un
`CHECK` por tipo (defensa en profundidad, R7-1/data-model): un `INSERT` forzado tampoco guarda un
IBAN con letras de más o espacios.

**Sobre el dígito de control (mod-97) — no se implementa.** Q3 y FR-015 dicen literalmente
"estructura y longitud" **dos veces** (decisión Q3: "validar estructura y longitud") y el edge case
de la spec define lo aceptado como "es válido si su estructura y longitud lo son". Añadir la
verificación mod-97 sería una **regla adicional no pedida** que rechazaría números estructuralmente
válidos (p. ej. si el equipo transcribe mal un dígito, recibiría un error que la spec no describe,
y si la tabla de longitudes fallara, un rechazo injusto). Por eso la validación es
**estructura + longitud** y el mod-97 queda como **endurecimiento revisable** (si el humano lo
pide: una función más en `numero.go` y un caso en la tabla de pruebas; sin cambios de modelo).
**Esto se eleva como pregunta al humano** (ver plan.md, "Preguntas para el humano").

**Alternativas descartadas.**

| Alternativa | Por qué no |
|---|---|
| Librería de validación de IBAN (p. ej. `github.com/...iban`) | **0 dependencias nuevas**: la regla pedida (estructura + longitud) es una regex y una tabla de constantes; una librería traería además mod-97 (regla no pedida) y un punto de mantenimiento externo |
| Solo longitud en rango 15–34 sin tabla por país | Aceptaría `CR` con 30 caracteres (longitud inválida para Costa Rica): SC-009 pide "0 IBAN fuera de la estructura/longitud válida" |
| Aceptar países fuera del registro con rango 15–34 | No se puede validar la longitud de un país sin registro; un IBAN de país desconocido no puede demostrarse válido |
| Guardar el número con espacios/guiones tal como se escribió | Q3 manda guardar "normalizado a los caracteres del número"; además rompería la comparación de duplicados |
| Agrupar en bloques en el servidor (devolver el número ya agrupado) | FR-015 prohíbe transformaciones sobre el valor publicado; la agrupación es presentación (como el formato a.m./p.m. de F3, que también lo hace el cliente) y así `number` es siempre exacto |
| Mod-97 obligatorio | Regla no pedida (estructura + longitud), riesgo de falsos rechazos; queda como revisable |

---

## R7-3. Número de SINPE Móvil (Q4/FR-016): celular costarricense de 8 dígitos, con o sin `+506`

**Pregunta.** Q4/FR-016: debe corresponder a un "celular costarricense de 8 dígitos", admitido "con
o sin prefijo `+506`", y el sistema "normaliza el prefijo al guardar". ¿Cuál es la forma guardada
y cómo se distingue un celular de cualquier otra cifra de 8 dígitos?

**Decisión.**

- **Aceptación**: se admite `88888888`, `8888-8888`, `8888 8888`, `+506 8888 8888`,
  `+50688888888` (espacios y guiones se eliminan al normalizar; el prefijo es opcional).
- **Validación**: exactamente **8 dígitos** y **primer dígito 6, 7 u 8** —los prefijos de móvil en
  la numeración de Costa Rica—, que es lo que distingue un **celular** de una línea fija (Q4 dice
  "celular costarricense", no "cualquier número de 8 dígitos"). Un `2xxx-xxxx` (fija) o un número
  de otra longitud se rechaza con `400 invalid` y `details.number`.
  **Default revisable**: si el equipo prefiere aceptar cualquier 8 dígitos, es quitar la condición
  del primer dígito en `numero.go` (una línea y sus casos de prueba). **Se eleva como pregunta**.
- **Normalización al guardar**: **`+506XXXXXXXX`** (13 caracteres: prefijo `+506` siempre presente
  y los 8 dígitos juntos). La spec deja "la representación exacta (`+506` presente o ausente en el
  dato guardado)" como detalle del plan; se elige el prefijo **presente** porque es la forma
  autónoma (el número se entiende fuera de Costa Rica — US4: quien ofrenda puede estar en el
  extranjero —, es un único valor copiable y hace la comparación de duplicados trivial). Ese valor
  es el que se publica ("su formato normalizado", FR-003) y el que copia el botón de copiar.
- **`CHECK` en la BD**: `number ~ '^\+506[0-9]{8}$'` para `kind = 'sinpe'` (defensa en profundidad).

**Alternativas descartadas.** Guardar solo los 8 dígitos (pierde la señal de país y obliga al
cliente a inventar el prefijo al mostrar); guardarlo como el usuario lo escribió (rompe
"normalizado al guardar" y la comparación de duplicados); presentarlo agrupado (`+506 8888 8888`)
como valor de copia (el valor copado debe ser exacto; la agrupación visual es decisión de `ux.md`
sobre el mismo string); librería de teléfonos (los tags `phone` de F2/F3 no aplican: aquí la regla
es de numeración costarricense y vive en el dominio).

---

## R7-4. Moneda (Q2/FR-019): solo cuentas IBAN, dos valores, etiqueta localizada en el cliente

**Decisión.** Columna `currency TEXT NULL CHECK (currency IN ('CRC','USD'))` con dos `CHECK`
cruzados sobre `kind`: **IBAN ⇒ `currency IS NOT NULL`** (FR-014: "la moneda solo aplica y es
obligatoria dentro de las cuentas IBAN") y **SINPE ⇒ `currency IS NULL`** (Q2: "el SINPE Móvil no
lleva moneda"; el formulario ni ofrece el campo y la ficha nunca lo muestra). Se guardan los
**códigos** `CRC`/`USD` y la etiqueta visible («Colones (₡)» / «Dólares (US$)») la localiza el
**frontend** con i18n, exactamente como F3 hace con `day_of_week` ("selector localizado por i18n,
NUNCA texto libre", `000005`): la moneda no es contenido traducible (FR-007/SC-006: idéntica en
ambos idiomas) y no se guarda texto libre que cada formulario deletrearía distinto.

**Alternativas descartadas.** Guardar «colones»/«dólares» como texto (sin catálogo, sin control);
ampliar el catálogo a más monedas (Q2 fija dos: colones y dólares; ampliarlo sería una decisión
aparte, mismo criterio que el catálogo fijo de redes de F3); moneda obligatoria para todos (Q2 la
limita a IBAN); moneda opcional también en IBAN (FR-014 la hace obligatoria ahí).

---

## R7-5. Qué es traducible y qué no (FR-007/FR-019/US4/SC-006) — **contradicción de la spec, resuelta por el peso de los criterios medibles**

**Pregunta.** ¿La **entidad financiera** (nombre del banco) es traducible al inglés?

**Contradicción detectada en la spec (no resuelta aquí por gusto: se señala al humano).**

| Fuente | Dice |
|---|---|
| **FR-007** | "el número, el tipo de método, **la entidad financiera** y la moneda **NO son contenido traducible**: son un único valor idéntico en ambos idiomas" |
| **US4, escenario 5** | "el número, el tipo, **la entidad financiera** y la moneda **no cambian en absoluto**: no son contenido traducible" |
| **SC-006** (medible) | "el tipo, **la entidad financiera** y la moneda son **idénticos en ambos idiomas** (0 alteraciones por cambio de idioma)" |
| **FR-019** | "los textos (**entidad y titular**) DEBEN ser traducibles de forma opcional al inglés **salvo los datos no traducibles de FR-007**" |
| **Key Entities** | "Los textos que lo acompañen (**entidad, titular**) son traducibles… **el número, el tipo, la entidad y la moneda no lo son**" (la misma frase la lista en ambos lados: artefacto de edición) |

**Decisión aplicada (default, revisable por el humano): la entidad financiera NO es traducible** y
el **titular SÍ** lo es, de forma opcional. Motivo: tres de las fuentes —incluidos **dos criterios
verificables** (US4 esc. 5 y SC-006) y la norma de bilingüismo (FR-007)— dicen que no lo es, y solo
una mención incidental en FR-019/Key Entities dice que sí (y la propia FR-019 contiene la salvedad
"salvo los datos no traducibles de FR-007", que la deja sin efecto). Además es la única lectura que
**garantiza SC-006 por construcción**: si `institution_en` existiera y se rellenara, el nombre del
cambiaría con el idioma y SC-006 fallaría literalmente.

**Cómo se implementa**: `institution` es **una sola columna** (no un par `*_es`/`*_en`) y el único
campo traducible del método es `holder_es`/`holder_en` (titular, opcional: si se rellena, su
versión en español debe existir —"un campo vacío no obliga a nada", Decisión 8—). Si el humano
decide lo contrario, el cambio es acotado y documentado: `institution` → `institution_es`/
`institution_en` (migración + `CHECK` de español por campo + fallback `en → es` + contrato) y
habría que reescribir la línea de SC-006/US4 esc. 5 (hoy imposible de cumplir con traducción).
**Se eleva como pregunta #1 al humano.**

El resto de la regla bilingüe es la de F3/F4/F6 sin cambios (R7-8): traducibles los **textos**
(introductorio, de uso y titular); no traducibles `kind`, `number`, `institution` y `currency`.

---

## R7-6. Permisos: se **activa** `donaciones`; F7 **no** crea permisos ni toca F2

Sin cambios sobre R6-6. Verificado en `000002_create_roles_and_permissions`: el catálogo sembrado
incluye `('donaciones', 'Donaciones')` (línea 25). F7 solo:

- monta **un** grupo de panel con `middleware.AdminChain("donaciones", deps)` (authn → guard de
  cambio de contraseña → authz(`donaciones`) → CSRF), con el `Recorder` del dominio `donaciones`
  para registrar denegaciones (`result='denied'`);
- añade `DONACIONES = 'donaciones'` a `frontend/src/lib/permissions.ts` (y a
  `MODULE_PERMISSIONS`, con lo que `isPermissionAvailable` deja de marcarlo como «Disponible más
  adelante» — hoy sale de `features/roles/permissions.test.ts:24` como `false`) y protege
  `/panel/donaciones` con `RequirePermission(DONACIONES)`;
- **no** hay migración de permisos: el catálogo de `permissions` **no cambia** (sigue en 10 filas,
  como en F6) y `000002` no se toca (§VI);
- el administrador inicial ya tiene `donaciones` porque el init de F2 asigna **todo** el catálogo.

Permisos independientes (Decisión 5): tener `donaciones` no da acceso a `eventos`, `ministerios`,
`portada` ni `admin_usuarios_roles`, ni al revés (SC-005, US2 esc. 9). Toda operación no
autorizada → `403 forbidden` genérico y registrada.

**Alternativas descartadas.** Permiso nuevo por sub-acción (el catálogo de F2 es por módulo);
renombrar el permiso a `ofrendas` (Q8 y FR-015 de F2 lo prohíben: el permiso **no cambia de
nombre**); comprobar permisos solo en el cliente (CWE-862).

---

## R7-7. Superficie HTTP: el público nombra la **sección** (`ofrendas`), el panel nombra el **módulo** (`donaciones`)

**Pregunta.** Q8 separa los nombres a propósito: la sección pública es «Ofrendas y diezmos» en
`/ofrendas` y el módulo/permiso del panel es «donaciones». ¿Cómo se nombran las rutas de la API?

**Decisión.** La misma distinción de Q8, aplicada a los segmentos:

| Superficie | Rutas | Protección |
|---|---|---|
| Pública (sección) | `GET /api/v1/ofrendas?lang=es\|en` | ninguna (solo lectura); **`Cache-Control: no-store`** |
| Pública (portada) | `GET /api/v1/portada/ofrendas?lang=es\|en` (bloque, tope 5) | ninguna; **`no-store`** |
| Panel — métodos | `GET/POST /api/v1/admin/donaciones`, `GET/PATCH/DELETE /api/v1/admin/donaciones/{id}` | `AdminChain("donaciones")` |
| Panel — textos | `GET /api/v1/admin/donaciones/textos`, `PUT /api/v1/admin/donaciones/textos/{textKey}` | `AdminChain("donaciones")` |

El segmento **público** es `ofrendas` porque sirve a la **sección pública** (espejo de la ruta del
SPA `/ofrendas`, como `/api/v1/ministerios` espejó `/ministerios` y `/api/v1/proximamente` sirvió
`/eventos`); el segmento **de panel** es `donaciones` porque sirve al **módulo** del panel (Q8:
"el módulo y el permiso del panel siguen llamándose «donaciones»"), y `/api/v1/portada/ofrendas`
sigue el patrón exacto de `/api/v1/portada/ministerios` y `/api/v1/portada/proximamente`.

- **Un documento público por superficie**, con el sobre `{lang, items?…}` de F3/F4/F6 (§8.1.2:
  `items` se **omite** cuando está vacía, nunca `null`); la sección añade además `intro`/`usage`
  (los dos textos publicados, omitidos si no lo están).
- **Presencia de la navegación = documento de la sección** (`GET /api/v1/ofrendas`), **no** del
  bloque. Aquí F7 se aparta a propósito de F6/P6-9 (que la dedujo del bloque) y vuelve al patrón
  de F4/D-23 ("la presencia se deduce de la misma respuesta que pinta la sección"). Motivo
  estructural: el bloque de F7 solo lleva **métodos** (FR-022) y la sección puede existir con
  **solo textos** publicados (R7-12): el documento del bloque **no puede** decidir la presencia sin
  contradecir FR-005. `useSiteLinks` (que ya consume tres fuentes: portada, `useUpcomingEvents` y
  `usePortadaMinisterios`) suma una cuarta consulta —`usePublicOfrendas`— compartida por
  `queryClient` con la página `/ofrendas`. Coste: una petición pública pequeña más por página
  (el documento trae números y, como mucho, dos textos de 2.000 caracteres; acotado por
  `maxPublicItems = 200`, R7-14).
- **Sin paginación ni filtros públicos** (FR-004: el visitante no hace nada para verlo todo); el
  panel **sí** pagina (`limit`/`offset`) con filtro opcional `publicationState`.
- `no-store` en ambos públicos garantiza SC-003 (un cambio guardado se ve en la primera carga).
- El registro de `error.code` **no crece**: `invalid`, `unauthenticated`, `forbidden`, `not_found`,
  `conflict`, `internal`, todos ya en el enum.

**Alternativas descartadas.**

| Alternativa | Por qué no |
|---|---|
| Nombrar todo `donaciones` (`/api/v1/donaciones`, `/api/v1/portada/donaciones`) | La superficie pública serviría a la sección «Ofrendas y diezmos» con un nombre que la spec reserva al módulo del panel (Q8); además rompería el espejo ruta SPA ↔ API de F4/F6 |
| Nombrar todo `ofrendas` (`/api/v1/admin/ofrendas`) | El módulo y el permiso del panel **se llaman** `donaciones` (Q8, FR-011) y `AdminChain` toma ese código |
| Un solo endpoint para sección y bloque | FR-022 exige el tope **resuelto en el servidor** para el bloque y la sección necesita todos |
| Presencia de navegación desde el bloque (P6-9 de F6) | Con solo textos publicados el bloque no existe (FR-022) pero la sección sí (R7-12): la navegación quedaría huérfana respecto a la página |
| Un endpoint "presencia" (`/exists`) | Una tercera ruta pública para un booleano que ya devuelven las otras dos |
| Añadir el bloque a `GET /api/v1/portada` de F3 | Mezclaría dominios y tocaría el contrato cerrado de F3 (mismo criterio que P4-18/P6-16) |

---

## R7-8. Bilingüismo: `*_es`/`*_en` solo donde hay texto traducible, `?lang` con fallback en el servidor

Sin cambios sobre R3-1/R3-2/R4-3/R6-9, aplicados al inventario reducido de F7:

- **Traducibles**: `holder` (titular del método), `body` de los dos textos (introductorio y de uso).
- **No traducibles**: `kind`, `number`, `institution` (R7-5), `currency` (R7-4).
- **Guardado**: `*_en` opcional; `""`/espacios → `NULL` (normalización I6: trim en el service);
  si un campo opcional se rellena en algún idioma, su versión **en español debe existir**
  (`holder_en` sin `holder_es` → `400 invalid` con `details.holderEs`; mismo `CHECK` en la BD;
  `body_en` exige `body_es`, que además es obligatorio).
- **Lectura pública**: `?lang=es|en` (por defecto `es`, Decisión 8) resuelve cada campo traducible
  en el servidor con `en si existe, si no es` (R3-2): **nunca** un campo vacío por falta de
  traducción (SC-006). El cliente público no aplica fallback (analyze I2 de F3).
- El **panel va en español** (FR-018): sin `i18n/` en las pantallas de gestión.

---

## R7-9. Duplicados: mismo **tipo y mismo número** (normalizado) → `409 conflict`

**Pregunta.** La spec trae dos redacciones del mismo default revisable: el edge case dice "dos
métodos de donación con el **mismo número** (mismo IBAN o mismo SINPE, o una cuenta repetida): se
trata como duplicado y se rechaza o se señala" y Assumptions dice "un duplicado exacto (**mismo
tipo, misma entidad y mismo número**) se rechaza o se señala como duplicado".

**Decisión aplicada (default, revisable).** "Se rechaza" (la opción que la spec pone primero y la
que usan F2/F3/F4/F6) con la regla **más fuerte de las dos**: **`kind` + `number` normalizados**
(`ExistsDonationMethodDuplicate`), que cubre el caso de Assumptions y además el del edge case
("una cuenta repetida" bajo un nombre de banco escrito distinto). Como la normalización de F7 es
determinista y estructural (R7-2/R7-3: el número se guarda ya normalizado), la comparación es una
igualdad simple sobre columnas, sin `btrim`/`lower` en SQL. **Sin índice `UNIQUE` en la BD** (mismo
gesto que F6/R6-11: la regla vive en el service y su transacción; el volumen real —pocas cuentas—
no justifica el índice y así el default "se rechaza o se señala" sigue siendo una línea que
cambiar). Si el humano quiere blindarlo frente a dos altas simultáneas, `UNIQUE (kind, number)` es
una línea en la migración siguiente (documentado en data-model.md).

**Alternativas descartadas.** `409` solo con la tripleta tipo+entidad+número (dejaría pasar la
misma cuenta con el banco escrito de otra forma); solo señalizar en el formulario (la API forzada
quedaría sin señal, como advierte R6-11); `UNIQUE` desde ya (bloquearía el default "se señala" y
haría falta una migración para revertirlo).

---

## R7-10. Auditoría: 8 códigos `donation.*` sobre el registro de F2 (FR-017)

Mismo patrón literal que R3-11/R4-10/R6-12: `InsertAdminAction` **dentro de la transacción** de
cada mutación (fail-closed: si el registro falla, la operación **no se aplica**),
`target_kind='content'`, `target_label` descriptivo y registro **cerrado** por `CHECK` en la BD
(migración `000013`). Los **8 códigos** nuevos en `platform/audit`:

| Situación | Código | Etiqueta del panel (`ACTION_LABELS`) |
|---|---|---|
| Alta de un método (nazca en `draft` o `published`) | `donation.create` | Creó un método de donación |
| Edición de datos de un método (incluye `sortOrder`) | `donation.update` | Modificó un método de donación |
| Solo cambio de estado de un método | `donation.publish` / `donation.unpublish` | Publicó / Retiró un método de donación |
| Borrado de un método | `donation.delete` | Eliminó un método de donación |
| Guardado de un texto (introductorio o de uso; alta o edición) | `donation.text.update` | Modificó un texto de ofrendas y diezmos |
| Solo cambio de estado de un texto | `donation.text.publish` / `donation.text.unpublish` | Publicó / Retiró un texto de ofrendas y diezmos |

Regla exacta (misma que F4/F6, analyze M5 y A6-10): **el alta registra solo `donation.create`**
—una sola fila, también con `publicationState: published`—; en la **edición**, datos y estado en el
mismo guardado → **dos filas** (`donation.update` + `donation.publish`/`unpublish`; ídem para
textos); **reordenar** se registra como `donation.update` (el orden es un dato del elemento, igual
que en F3/F4/F6; si el humano prefiere `donation.reorder`, es un código más en la misma lista).
`target_label` = `Donaciones · <institution> · <number>` para métodos y
`Donaciones · Texto introductorio` / `Donaciones · Texto de uso` para los textos (máx. 254,
`CHECK` de F2). Los textos **no** tienen código `donation.text.create`: son singletons y su
primer guardado se audita como `donation.text.update` (mismo criterio que `home.about.update` de
F3, que cubre alta y edición del singleton). Éxito en la transacción; `failure` (DTO inválido,
duplicado) y `denied` (sin permiso) best-effort. La resolución método+ruta → acción vive en el
dominio `donaciones` (implementa `audit.Recorder`).

**Alternativas descartadas.** Registro propio para F7 (FR-017 pide el de F2); registrar tras el
commit (rompe "no aplicarse sin registro"); un solo código `donation.update` para todo (no
distinguiría quién **retiró** un número de cuenta del público, que es la operación sensible);
`target_kind` nuevo (los elementos de contenido ya son `content`).

---

## R7-11. Publicación: por método **y** por texto, con su propio estado, "se publica al guardar"

Patrón R3-3/R4-12/R6-13 sin cambios, aplicado a **dos niveles** (Q10):

- **Cada método** lleva `publication_state` y se publica/retira por separado (FR-012).
- **Cada texto** (introductorio y de uso) lleva **su propio** `publication_state` (Q10/FR-012):
  un texto en borrador simplemente no se muestra y el resto de la sección se sostiene (US3 esc. 7:
  "sin huecos ni marcadores vacíos").
- Consultas `…Published` separadas para el público (la única función que ve borradores es la del
  panel, protegida); el `PATCH`/`PUT` sobre un elemento publicado guarda y **conserva** el estado
  (FR-013: "se publica al guardar"); el paso por borrador solo existe al crear.
- Un borrador **nunca** sale hacia el público por ninguna vía (SC-002): ni la sección, ni el
  bloque, ni ningún otro DTO público (F7 no tiene imágenes, así que `IsFilePublished` **no se
  toca** — Q6/FR-020 —).

---

## R7-12. Sección sin nada publicado: ocultar y redirigir `/ofrendas` a la portada — y **qué significa "nada publicado"**

**Regla de mecánica** (la decisión del cliente es Q9, no se reabre): idéntica a R6-13/P6-21 de F6
—**distinta a propósito** del aviso sin redirigir de `/eventos` (F4/R4-13), que no se toca—:

1. La entrada «Ofrendas y diezmos» **no existe** en la barra, el menú hamburguesa ni el mapa del
   sitio (fuente única `useSiteLinks`, R7-7) y el bloque de la portada **no se renderiza** (FR-022).
2. La URL directa `/ofrendas` **redirige a la portada (`/`)**, sin parpadeo y **sin aviso**: la
   decisión vive en el **frontend**, en el elemento de ruta de `/ofrendas` —`SectionSkeleton`
   mientras `usePublicOfrendas` está en vuelo y, **solo con éxito y sin contenido**,
   `<Navigate to="/" replace />`; `replace` evita bucles con «Atrás»— y **ante error NO se
   redirige** (estado de error con «Volver a intentar», patrón de `EventsPage`).
3. **El contrato no cambia**: `GET /api/v1/ofrendas` sigue respondiendo `200` con las claves de
   contenido **omitidas**. Un `3xx` del API rompería el sobre `{lang, …}` y a sus consumidores; y un
   `404` daría un código de error para algo normal (una sección aún sin contenido). La
   "redirección/404" del encargo queda, por tanto, como **redirección del SPA** documentada en la
   prosa del contrato (igual que F6).

**Qué es "no tener nada publicado" — punto que la spec deja en el aire y se decide aquí.**
FR-005 dice "cuando la sección **no tiene nada publicado** (o todo está en borrador —**métodos y
textos incluidos**—)" y US1/US3/SC-010 repiten "sin nada publicado". **Ningún escenario de
aceptación** cubre el caso "textos publicados y 0 métodos". **Decisión aplicada (default,
revisable)**: la sección —y con ella la entrada de navegación y la vida de `/ofrendas`— existe si
hay **al menos un método publicado o al menos un texto publicado** (lectura literal del paréntesis:
"todo está en borrador —métodos y textos incluidos—" es el estado en el que se oculta). Consecuencias
del caso frontera: con **solo textos** publicados, `/ofrendas` muestra los textos (sin grupos de
métodos, sin huecos), la navegación luce el enlace y **el bloque de la portada sigue oculto**
(FR-022 es explícito: el bloque vive de los **métodos** publicados). La alternativa —exigir ≥1
método publicado para que la sección exista— es igual de defendible y es **un solo predicado**
cambiable en `hasOfrendasContent`; **se eleva como pregunta al humano** (ver plan.md).

**Alternativas descartadas.** Aviso de sección vacía sin redirigir (`HiddenNotice`, patrón F4:
**descartado por Q9** para esta sección y conservado en `/eventos`); `404` (un error para algo
normal); página vacía sin mensaje (contradice SC-010); redirigir desde el backend o con `302` en
el contrato (rompe el sobre JSON); redirigir también ante **error** de la API (una caída transitoria
expulsaría al visitante).

---

## R7-13. Bloque de la portada: endpoint propio, tope 5 en el servidor, **mismo DTO de ítem**, oculto sin métodos publicados

**Contexto.** FR-022/US6/SC-014 (Q7): el bloque va **inmediatamente debajo del bloque de
ministerios (F6)** —es decir, entre `MinisteriosBlock` y «quiénes somos» en `HomePage`—, con hasta
**5** métodos publicados en el orden de la sección y un enlace/botón a `/ofrendas`.

**Decisión.** Endpoint público propio **`GET /api/v1/portada/ofrendas?lang=es|en`**, registrado por
el dominio `donaciones` (mismo gesto que F4 con `proximamente` y F6 con `ministerios`: el segmento
`portada` es el agregado de F3 pero el recurso es de este módulo):

- **Tope 5 firme en el servidor** (FR-022): `maxPortadaItems = 5` (constante del service) +
  `LIMIT 5` en la **misma** consulta de la sección (`ListPublishedDonationMethods`, con límite
  distinto). **Sin parámetro de límite**: el 5 es decisión de la spec.
- **Mismo predicado "publicado" y mismo orden** que la sección: `ORDER BY CASE kind … END,
  sort_order, id` (iban primero, sinpe después) — los "5 primeros según el orden de la sección" de
  SC-014 salen de recortar la propia lista.
- **Mismo DTO de ítem que la sección** (`MetodoDonacionPublico`), literalmente: FR-022 dice "el
  bloque es un **extracto de los mismos métodos de donación, no un modelo nuevo**". La tarjeta
  compacta del bloque es composición de `ux.md` (pinta lo que quiera de esos campos: tipo, número,
  entidad, moneda, titular); **no hace falta un DTO más pequeño** y así imposible que el bloque y
  la sección se desalineen. Si `ux.md` fijara una tarjeta con menos campos, recortar en el cliente
  no toca la API.
- **Sin métodos publicados el bloque NO se renderiza** (FR-022/SC-014): la respuesta es
  `{"lang":"…"}` (`items` omitida) y la portada **no pinta nada** — ni bloque vacío ni mensaje
  (coherente con el bloque de F6 y distinto a propósito del «Lo que viene» de F4).
- **Ubicación**: `OfrendasBlock` se inserta en `HomePage` **inmediatamente después de
  `MinisteriosBlock`** y antes de `WhoWeAreSection`. Es **una inserción**: el resto de la portada
  de F3 no se toca (FR-023).

**Alternativas descartadas.** Un `ItemOfrendaPortada` más pequeño (FR-022 pide "no un modelo
nuevo"; duplicaría esquema y pruebas); derivar los 5 en el cliente (FR-022 exige el tope en el
servidor); añadir el bloque a `GET /api/v1/portada` de F3 (mezcla dominios, toca su contrato);
mostrar el bloque vacío con mensaje como F4 (contradice FR-022/SC-014); carrusel/paginación
(Out of Scope).

---

## R7-14. Validación de entrada y límites (FR-014)

Toda validación en el **backend** (§IV, CWE-20) aunque el formulario también valide:

- **Tags existentes** de `platform/validate` (sin tags nuevos): `required`, `omitempty`, `max`,
  `oneof` (para `kind`, `currency`, `publicationState` y `textKey`).
- **Reglas de dominio** en `internal/donaciones/numero.go` + constantes del service (R7-2/R7-3):

  | Campo | Regla / límite | Origen |
  |---|---|---|
  | `kind` | `iban` \| `sinpe` (obligatorio) | Q1/FR-019 |
  | `number` (IBAN) | estructura `^[A-Z]{2}[0-9]{2}[A-Z0-9]+$` + longitud exacta por país (registro IBAN); se guarda normalizado | Q3/FR-015 (R7-2) |
  | `number` (SINPE) | celular CR de 8 dígitos (primer dígito 6/7/8) con o sin `+506`; se guarda `+506XXXXXXXX` | Q4/FR-016 (R7-3) |
  | `institution` | obligatorio, 1…160, no traducible | Q2/FR-019 (R7-5) |
  | `holderEs`/`holderEn` | opcional, 1…160; `holderEn` exige `holderEs` | Q2/FR-007 |
  | `currency` | obligatoria **solo** en IBAN; prohibida en SINPE; `CRC` \| `USD` | Q2/FR-014 (R7-4) |
  | `sortOrder` | `INTEGER` (rango int32), por defecto 0 | FR-004 (patrón F3/F4/F6) |
  | `textEs` | obligatorio, 1…**2.000** | default revisable de la spec ("límite del orden de 2.000") |
  | `textEn` | opcional, 1…2.000; exige `textEs` | Decisión 8 |

- **Obligatorios de un método**: `kind`, `number`, `institution` (+ `currency` si es IBAN). **Lo
  demás es opcional** (Q2: el titular es opcional y "un campo vacío no obliga a nada").
- **Ante datos inválidos**: `400 invalid` con `details` por campo y **sin guardado parcial** (la
  transacción escribe la fila completa ya fusionada — R7-15 — o no escribe nada).
- Contenido tratado **siempre como datos** (§IV, CWE-79): texto plano multilínea, emojis, tildes,
  `ñ`, `¿¡` intactos; **en los números no se admiten emojis ni letras fuera de la estructura**
  (la validación de R7-2/R7-3 ya lo impide); sin `dangerouslySetInnerHTML`.

---

## R7-15. Concurrencia: gana la última escritura completa

Igual que R3-17/R4-11/R6-16. El `PATCH` del método **fusiona los campos presentes** sobre la fila
actual (los ausentes no cambian; `null` explícito borra lo opcional, p. ej. `holderEs`, `currency`
solo con cambio de tipo, `sortOrder` no es anulable) y la transacción escribe **la fila completa ya
fusionada** —normalización del número incluida— con su auditoría. El `PUT` de los textos es
**reemplazo completo** (patrón F3 de los singletons: `PUT /…/quienes-somos` "crea o reemplaza el
singleton"): el formulario del panel envía siempre el texto entero. Sin control optimista
(`version`/`If-Match`) ni historial (fuera de alcance; si el humano lo pide, es una columna +
`409`).

---

## R7-16. Dependencias nuevas: **ninguna**

Constitución §II. F7 **no añade** ninguna dependencia ni en backend (`pgx`/`sqlc`,
`testcontainers-go`, `net/http` ya están; la validación del IBAN es una regex y una tabla de
constantes, la del SINPE otra regex) ni en frontend (React 19, React Router, TanStack Query, React
Hook Form + Zod, Tailwind, Vitest + Testing Library + MSW, Playwright). **Descartado
explícitamente**: librería de IBAN (traería mod-97, regla no pedida, y un punto de mantenimiento
externo para una tabla de ~90 constantes), librería de teléfonos (la regla es de numeración
costarricense y vive en el dominio), librería de "copiar al portapapeles" (el API
`navigator.clipboard` del navegador basta y la pieza de botón ya existe en el inventario de F10),
cualquier dependencia de UI nueva (el inventario de F2/F3/F4/F6/F10 cubre la ficha, la lista, los
textos y el bloque).

---

## Resumen de decisiones

| # | Decisión | Un vistazo |
|---|---|---|
| R7-1 | Modelo | Dos tablas: `donation_methods` (listado administrable) y `donation_texts` (singleton por clave `intro`/`usage`); sin tablas de traducciones ni siembra de contenido |
| R7-2 | IBAN (Q3) | Estructura + longitud exacta por país (registro IBAN embebido, CR = 22); guardado **normalizado a los caracteres del número**; agrupación en bloques **presentación** del frontend; el botón copia el valor exacto; **mod-97 no** (regla no pedida; revisable) |
| R7-3 | SINPE (Q4) | Celular CR de 8 dígitos (primer dígito 6/7/8) con o sin `+506`; guardado `+506XXXXXXXX` |
| R7-4 | Moneda (Q2) | `currency IN ('CRC','USD')` **obligatoria en IBAN y prohibida en SINPE** (2 `CHECK` cruzados); etiqueta localizada en el cliente (patrón `day_of_week`) |
| R7-5 | Traducibilidad | `holder` sí; `institution`, `kind`, `number`, `currency` **no** (FR-007/US4/SC-006) — **contradicción de la spec señalada al humano** |
| R7-6 | Permisos | `AdminChain("donaciones")` con el permiso **ya sembrado** en `000002`; **sin migración de permisos y sin tocar F2** |
| R7-7 | Superficie HTTP | `GET /api/v1/ofrendas` + `GET /api/v1/portada/ofrendas` (público = **sección**) y `/api/v1/admin/donaciones…` (panel = **módulo**), `no-store`; **la presencia de la navegación bebe del documento de la sección** (patrón F4/D-23, no F6/P6-9: los textos solo viven en la sección) |
| R7-8 | Bilingüe | `*_es`/`*_en` solo en titular y textos; `?lang` con fallback en el servidor; panel en español |
| R7-9 | Duplicados | Mismo **tipo + número normalizado** → `409` (más fuerte que el default de Assumptions; cubre el edge case); sin `UNIQUE` en BD (patrón F6) |
| R7-10 | Auditoría | 8 códigos `donation.*` (migración `000013`), transaccional fail-closed, `target_kind='content'`; textos sin código `create` (singleton) |
| R7-11 | Publicación | `publication_state` por método **y por texto** (Q10); "se publica al guardar"; consultas `…Published`; F7 **no toca** `IsFilePublished` (sin imágenes) |
| R7-12 | Sin nada publicado | Entrada oculta + **`/ofrendas` redirige a la portada** (P7-21; decisión del cliente Q9, distinta a propósito de F4); **"nada publicado" = sin métodos NI textos** (default sobre un caso que la spec no cubre; pregunta al humano) |
| R7-13 | Bloque de portada | `GET /api/v1/portada/ofrendas` con **tope 5 en el servidor**, mismo orden y **mismo DTO de ítem** que la sección ("no un modelo nuevo"), botón a `/ofrendas`, **oculto sin métodos publicados**, debajo del bloque de ministerios |
| R7-14 | Validación | Tags existentes + reglas de dominio en `numero.go`; límites: entidad/titular 160, textos 2.000; `400 invalid` con `details`, sin guardado parcial |
| R7-15 | Concurrencia | `PATCH` fusiona y escribe la fila completa; `PUT` de textos reemplaza (singleton de F3); última escritura completa gana |
| R7-16 | Dependencias | **Ninguna nueva** en backend ni frontend |
