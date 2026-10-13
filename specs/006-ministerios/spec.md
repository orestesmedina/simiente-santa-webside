# Feature Specification: Ministerios (F6)

**Feature Branch**: `006-ministerios`

**Created**: 2026-10-13

**Status**: Draft

**Input**: User description: "F6 — Ministerios (roadmap §3, `docs/producto/roadmap.md`): El visitante conoce cada ministerio, su encargado y el enlace al grupo de WhatsApp para unirse; el equipo administra el listado desde el panel." Y de `idea.md` §1: "nuestros diferentes ministerios (sus encargados y grupos de distribucion por whatsapps)", §2 (usuarios del sistema: "Crear, modificar, eliminar ministerios"), §4 ("los eventos, grupos de conexión, ministerios, videos y podcast sean dinamicos no algo estático") y §7 (el sitio debe adaptarse a cualquier dispositivo y manejar permisos, roles y estados). Decisiones aplicables del roadmap: **Decisión 2** (2026-09-28: los ministerios entran al MVP, F6), **Decisión 4** (2026-09-28: contenido con estados borrador/publicado; solo lo publicado es visible al visitante), **Decisión 5** (2026-09-28: permisos por módulo agrupados en roles; el catálogo de roles es libre y cada permiso autoriza la gestión completa de su módulo), **Decisión 6** (2026-09-28: el sitio público es bilingüe, español e inglés, con selección de idioma), **Decisión 8** (2026-09-28: cada contenido puede ingresarse en ambos idiomas o solo en español; el español es el idioma base y el inglés es opcional por contenido; la interfaz siempre es bilingüe) y **Decisión 9** (2026-10-04: el modelo de cuentas y roles de F2 —cuentas que no se eliminan, un solo rol por cuenta, roles editables/eliminables solo si no están en uso— se usa tal cual y F6 no lo cambia). **Dependencias**: F2 y F3 están terminadas; F6 es **independiente de F4, F5 y F7**. F2 provee el acceso autenticado al panel, los roles con permisos por módulo, el patrón de autorización verificado en el servidor y el registro de auditoría (Decisión 10, ampliado en F3 y F4); su catálogo fijo de permisos **ya incluye el permiso «ministerios»** (etiqueta «Ministerios»), verificado en FR-015 de F2 y en la siembra del catálogo de F2 (`('ministerios', 'Ministerios')`), por lo que **F6 NO crea permisos nuevos** —a diferencia de F4, que añadió «categorías»—: F6 **activa** un permiso ya reservado. F3 provee el sitio público con bilingüismo (fallback al español), estados borrador/publicado y la regla de ocultar secciones vacías. **Restricción del cliente** (`idea.md` §7 y su introducción): el sitio debe adaptarse a cualquier dispositivo y ser apto para personas de todas las edades, incluidas adultos mayores con lo básico de internet.

> **Nota de alcance y terminología**: esta funcionalidad tiene **dos partes**: (a) la **sección pública** del sitio donde el visitante conoce cada ministerio —nombre, descripción, encargado y enlace al grupo de WhatsApp para unirse— y (b) el **panel**, donde el equipo con el permiso **«ministerios»** administra el listado. Se reutiliza la terminología de F2 y F3: el **visitante** (el "usuario común" de `idea.md` §2: ve el sitio público, no tiene cuenta y no se registra) y el **usuario del panel** (el "usuario del sistema" y el "usuario administrador" de `idea.md` §2: tiene cuenta, inicia sesión y gestiona el contenido). **Qué es un ministerio en este producto**: un **área de servicio de la iglesia** (p. ej., alabanza, sonido, ujieres, acción social, escuela dominical) con un **encargado responsable** y un **grupo de WhatsApp de distribución** al que la persona se une (Decisión 2, `idea.md` §1). **Un ministerio NO es un grupo de conexión** (F5: talleres y clases como maquillaje, fotografía, inglés o asados, con horario, lugar y contacto para inscribirse) ni un evento o una actividad (F4): son conceptos distintos y **F6 no construye F5**. F6 entrega: la presentación pública del listado de ministerios, en español e inglés (español base, inglés opcional con fallback), y la gestión completa del listado desde el panel con el permiso **«ministerios»** —ya existente en el catálogo de F2, sin permisos nuevos—, estados borrador/publicado y registro de auditoría que **sigue el patrón ya construido en F2/F3/F4** (quién, qué, sobre qué y cuándo, solo lectura; no se re-especifica aquí). El sitio nace con el **sistema de diseño y movimiento de F10** ya entregado (se aplica, no se re-especifica).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Conocer los ministerios y unirse a uno (Priority: P1)

Como visitante (miembro de la iglesia o persona nueva), quiero ver cada ministerio de la iglesia —su nombre, su descripción, quién lo encarga y el enlace al grupo de WhatsApp para unirme— para conocer las áreas de servicio, elegir en cuál participar y unirme sin buscar en ningún otro sitio.

**Why this priority**: Es el corazón de F6 y de su fila del roadmap: "el visitante conoce cada ministerio, su encargado y el enlace al grupo de WhatsApp para unirse". `idea.md` §1 sitúa los ministerios "con sus encargados y grupos de distribución por whatsapps" entre lo que el sitio debe dar a conocer, y §4 exige que sean dinámicos. Sin esta historia no hay valor visible para el público y la gestión del panel (US2–US3) no tendría destinatario. Es la traducción del criterio de éxito del roadmap §8: "un canal unificado donde ver y conocer las actividades y grupos de la iglesia" y "una puerta para acercarse a la congregación".

**Independent Test**: Se prueba de forma independiente con ministerios publicados: un visitante sin sesión abre la sección y encuentra cada ministerio con su nombre, su descripción, su encargado y su enlace al grupo de WhatsApp; activa el enlace y llega al grupo para unirse con un solo clic o toque; al cambiar y publicar un dato desde el panel y recargar, el visitante ve el dato actualizado; el listado respeta el orden definido por el equipo; y nada del listado se confunde con un taller de F5 ni con un evento de F4. Entrega valor por sí sola: el público ya conoce los ministerios y puede unirse.

**Acceptance Scenarios**:

1. **Dado** ministerios publicados, **cuando** un visitante abre la sección sin iniciar sesión, **entonces** ve cada uno con su nombre, su descripción, su encargado y su enlace al grupo de WhatsApp para unirse.
2. **Dado** un ministerio publicado, **cuando** el visitante activa el enlace al grupo de WhatsApp (clic o toque), **entonces** se abre WhatsApp o la página de invitación del grupo para que pueda unirse, en un solo paso.
3. **Dado** varios ministerios publicados, **cuando** el visitante recorre la sección, **entonces** los ve en el orden estable definido por el equipo (Q5), sin tener que ordenar, filtrar ni buscar para verlos todos.
4. **Dado** la sección publicada, **cuando** una persona nueva la recorre, **entonces** identifica de qué se trata cada ministerio, quién lo encarga y cómo unirse, sin buscar en WhatsApp, en Facebook ni en ningún otro sitio.
5. **Dado** un ministerio publicado con imagen y otro sin imagen, **cuando** el visitante los ve, **entonces** ambos se presentan completos y coherentes, sin huecos ni campos en blanco.
6. **Dado** el listado de ministerios, **cuando** el visitante lo recorre, **entonces** no confunde un ministerio con un taller o clase (grupo de conexión de F5) ni con un evento o actividad (F4): la sección solo presenta ministerios (FR-001).
7. **Dado** varios ministerios que comparten encargado, **cuando** el visitante los consulta, **entonces** cada uno se muestra con su propio encargado y su propio enlace, sin mezclar datos entre ministerios.

---

### User Story 2 - Gestionar los ministerios desde el panel (Priority: P1)

Como usuario del panel con el permiso «ministerios», quiero crear, editar y eliminar los ministerios de la iglesia —con su nombre, su descripción, su encargado y su enlace al grupo de WhatsApp, en español y, si quiero, también en inglés— para mantener al día el listado sin depender de un desarrollador.

**Why this priority**: Es la mitad administrativa de F6 y el criterio de éxito de `idea.md` §4: que el contenido sea "dinamico no algo estático que no se pueden administrar", y la necesidad explícita de `idea.md` §2: los usuarios del sistema deben "crear, modificar, eliminar ministerios". Es construible y testeable por separado de la sección pública: reutiliza el patrón probado de F3 y F4 y activa un permiso ya existente en F2, sin tocar el modelo de permisos.

**Independent Test**: Se prueba de forma independiente con una cuenta con el permiso «ministerios» y otra sin él: la primera crea un ministerio con datos válidos y el sistema lo conserva; con campos obligatorios vacíos o un enlace de WhatsApp mal formado o no seguro, el sistema indica qué corregir y no guarda nada a medias; edita, reordena y elimina un ministerio (con confirmación) y la eliminación queda registrada en la auditoría; la segunda cuenta recibe una denegación clara en toda operación del módulo, incluidos los intentos forzados por la interfaz. Entrega valor por sí sola: el equipo ya administra sus ministerios.

**Acceptance Scenarios**:

1. **Dado** un usuario del panel con el permiso «ministerios», **cuando** crea un ministerio con datos válidos (nombre, descripción y encargado en español, y enlace de WhatsApp válido y seguro) y guarda, **entonces** el sistema confirma la operación y conserva el ministerio.
2. **Dado** una cuenta sin el permiso «ministerios», **cuando** intenta crear, editar o eliminar un ministerio, **entonces** el sistema impide la operación con un mensaje claro y no cambia nada.
3. **Dado** un formulario de ministerio con datos inválidos o incompletos (nombre vacío, descripción en español vacía, encargado vacío, enlace de WhatsApp mal formado o con un destino no seguro), **cuando** envío el formulario, **entonces** el sistema indica qué corregir y no guarda ningún cambio parcial.
4. **Dado** un ministerio existente, **cuando** edito sus datos y guardo, **entonces** el sistema conserva los cambios, y si el ministerio estaba publicado se hacen visibles de inmediato para el visitante (patrón de F3: se publica al guardar).
5. **Dado** un ministerio existente, **cuando** lo elimino y confirmo la eliminación, **entonces** el ministerio desaparece del panel y del sitio público, y la eliminación queda registrada en el registro de auditoría de F2 (quién, qué y cuándo).
6. **Dado** el formulario de ministerio, **cuando** la persona ingresa la versión en inglés, **entonces** el sistema conserva esa versión junto a la versión en español.
7. **Dado** el formulario de ministerio, **cuando** la persona deja vacío el campo en inglés, **entonces** puede guardar igualmente, porque el inglés es opcional por contenido (Decisión 8).
8. **Dado** una cuenta con el permiso «ministerios» pero sin otros permisos de módulo, **cuando** intenta forzar el acceso a la gestión de otros módulos (p. ej., «eventos» o «grupos»), **entonces** el sistema lo impide: los permisos son independientes entre sí (Decisión 5).
9. **Dado** el módulo con varios ministerios, **cuando** el usuario cambia su orden de presentación y guarda (Q5), **entonces** el nuevo orden se conserva y es exactamente el que ve el visitante.

---

### User Story 3 - Controlar la publicación de cada ministerio (Priority: P2)

Como usuario del panel con el permiso «ministerios», quiero dejar cada ministerio en borrador o publicarlo por separado, para que el público solo vea información revisada y nunca cambios a medias.

**Why this priority**: La Decisión 4 (contenido borrador/publicado) es parte explícita del alcance y protege la imagen pública de la iglesia: sin control de publicación, cada borrador sería público de inmediato. Va como P2 porque US2 ya entrega un MVP usable (crear y conservar el listado), mientras que esta historia añade el control de visibilidad elemento a elemento. Reutiliza literalmente el patrón probado en F3 (FR-013 de F3) y F4.

**Independent Test**: Se prueba de forma independiente con un ministerio en borrador y otro publicado: el borrador no aparece para el visitante y el resto de la sección se ve coherente; al publicarlo aparece; al devolverlo a borrador desaparece; una sección sin ningún ministerio publicado se oculta por completo (navegación y mapa del sitio incluidos, sin enlaces rotos); y ninguna respuesta del sistema hacia el público expone un borrador. Entrega valor por sí sola: el equipo decide qué se hace público y cuándo.

**Acceptance Scenarios**:

1. **Dado** un ministerio en estado borrador, **cuando** un visitante abre la sección, **entonces** ese ministerio no se le muestra y el resto de la sección se ve coherente, sin huecos ni campos en blanco.
2. **Dado** que no hay ningún ministerio publicado (o todos están en borrador), **cuando** un visitante abre el sitio público, **entonces** la sección se oculta por completo: no aparece vacía, y la navegación (barra o menú hamburguesa) y el mapa del sitio del pie no dejan enlaces rotos ni huérfanos (FR-005, patrón de F3 y F10).
3. **Dado** un ministerio en borrador, **cuando** el usuario del panel lo publica, **entonces** el ministerio aparece en el sitio público.
4. **Dado** un ministerio publicado, **cuando** el usuario del panel lo devuelve a borrador, **entonces** deja de ser visible para el visitante.
5. **Dado** un ministerio publicado, **cuando** el usuario del panel edita sus datos y guarda, **entonces** los cambios se hacen visibles de inmediato para el visitante (patrón de F3: se publica al guardar).
6. **Dado** cualquier ministerio en borrador, **cuando** se consulta el sitio público por cualquier vía, **entonces** el sistema NUNCA expone ese contenido ni sus datos al público.
7. **Dado** el módulo en el panel, **cuando** el usuario revisa los ministerios, **entonces** ve el estado (borrador/publicado) de cada uno para saber qué es público hoy.

---

### User Story 4 - Ver los ministerios en inglés (Priority: P2)

Como visitante que prefiere el inglés, quiero cambiar el idioma del sitio con el selector y ver los ministerios en inglés cuando exista traducción, para conocer las áreas de servicio de la iglesia y unirme aunque mi idioma base no sea el español.

**Why this priority**: Las decisiones 6 y 8 hacen del bilingüismo un compromiso del sitio, y el encargo de `idea.md` §1 busca dar a conocer la iglesia "en nuestra provicia y por qué no internacionalmente". Va como P2: el valor central (los ministerios, en español) ya está en US1, y el inglés amplía el público sin bloquear el resto. Reutiliza íntegramente el comportamiento del selector de idioma de F3/F10.

**Independent Test**: Se prueba de forma independiente con ministerios publicados en ambos idiomas y con ministerios que solo tienen español: al elegir inglés, la interfaz y los contenidos traducidos se muestran en inglés, los contenidos sin inglés se muestran en español y no aparece ningún campo vacío; al volver a español, todo se muestra en español; la elección persiste al navegar y en una visita posterior desde el mismo dispositivo; y un dispositivo sin preferencia guardada se muestra en español. Entrega valor por sí sola: el público anglófono conoce los ministerios.

**Acceptance Scenarios**:

1. **Dado** la sección en español, **cuando** el visitante elige inglés en el selector de idioma, **entonces** la interfaz y los contenidos con versión en inglés se muestran en inglés, y la elección se mantiene al navegar por el resto del sitio y se recuerda entre visitas en el mismo dispositivo (FR-010 de F3).
2. **Dado** un ministerio que solo tiene versión en español, **cuando** el visitante navega en inglés, **entonces** ese ministerio se muestra en español (idioma base), nunca vacío ni traducido automáticamente (FR-009 de F3).
3. **Dado** un dispositivo sin ninguna preferencia de idioma guardada (primera visita), **cuando** el visitante abre la sección, **entonces** se muestra en español, que es el idioma base (Decisión 8).
4. **Dado** la sección en inglés, **cuando** el visitante vuelve a elegir español, **entonces** toda la página se muestra de nuevo en español sin pérdida de información.
5. **Dado** el enlace al grupo de WhatsApp de un ministerio, **cuando** el visitante cambia de idioma, **entonces** el destino del enlace no cambia (no es contenido traducible) y sigue llevando al mismo grupo.

---

### User Story 5 - Recorrer la sección en cualquier dispositivo y de forma accesible (Priority: P2)

Como visitante de cualquier edad, dispositivo y nivel de experiencia con internet, quiero que la sección de ministerios se adapte a mi pantalla y sea fácil de usar, para conocer los ministerios y unirme sin dificultades.

**Why this priority**: Es una restricción explícita del cliente (`idea.md` §7 y su introducción: personas de todas las edades, "adultos mucho más mayores que saben lo básico de internet"). Va como P2 porque es una capa transversal sobre US1 y US4 verificable por separado, pero no puede postergarse más allá del MVP sin incumplir el encargo. La sección nace con el sistema de diseño y movimiento de F10 ya aplicado (sin re-especificarlo aquí).

**Independent Test**: Se prueba de forma independiente sobre la sección publicada: se abre en teléfono, tableta y computadora y se comprueba que todos los ministerios, encargados y enlaces están disponibles sin desplazamiento horizontal; se recorre con teclado y con lector de pantalla; se comprueba legibilidad y contraste, también con el modo oscuro de F10 activo y con texto ampliado. Entrega valor por sí sola: cualquier persona puede conocer los ministerios y unirse desde cualquier dispositivo.

**Acceptance Scenarios**:

1. **Dado** un teléfono con pantalla estrecha, **cuando** el visitante abre la sección, **entonces** ve todos los ministerios sin desplazamiento horizontal y puede activar el enlace de WhatsApp con comodidad (elementos de toque suficientemente grandes).
2. **Dado** una persona con poca experiencia en internet, **cuando** recorre la sección, **entonces** encuentra los ministerios con títulos claros y una navegación simple (barra visible en escritorio, menú hamburguesa claramente etiquetado en móvil, según F10), sin gestos desconocidos.
3. **Dado** una persona que usa el teclado o un lector de pantalla, **cuando** navega la sección, **entonces** puede recorrer y activar todos los contenidos y enlaces en un orden comprensible, con foco visible.
4. **Dado** una persona con baja visión, **cuando** usa la sección (en tema claro u oscuro), **entonces** el texto se lee con contraste suficiente y puede ampliarse hasta el 200 % sin que se pierda información.
5. **Dado** cualquier dispositivo, **cuando** el visitante abre la sección, **entonces** el contenido y las funciones disponibles son los mismos que en los demás dispositivos.

---

### Edge Cases

- **Contenido en borrador**: nunca debe ser visible para el visitante (US3, FR-012), ni en la página ni en ninguna respuesta del sistema hacia el público; un borrador no debe "filtrarse" por ninguna vía.
- **Sección sin ningún ministerio publicado** (o con todos en borrador): la sección **se oculta por completo** (patrón de F3 y F4); la navegación (barra, menú hamburguesa) y el mapa del sitio del pie se actualizan sin dejar enlaces rotos ni huérfanos.
- **Ministerio publicado que se edita**: los cambios guardados se hacen **visibles de inmediato** para el visitante (patrón de F3: se publica al guardar); no hay paso intermedio de borrador para lo que ya está publicado.
- **Ministerio sin imagen**: se presenta completo y coherente, sin huecos (FR-020); la imagen nunca es imprescindible para entender el ministerio ni para unirse.
- **Enlace de WhatsApp mal formado o no seguro** (esquema `javascript:`, `data:`, `vbscript:` u otro distinto de `https`, o un destino que no pasa la validación de FR-015): se rechaza con indicación de cómo corregirlo y no se guarda nada; el sistema nunca debe renderizar ni ejecutar un enlace no seguro.
- **Enlace de WhatsApp que deja de funcionar** (la invitación del grupo caduca o el grupo se renueva): el sitio solo enlaza y **no verifica la vigencia del destino**; la corrección la hace el equipo editando el enlace (mismo criterio de F3).
- **Ministerio sin enlace de WhatsApp o sin encargado**: no debe poder publicarse si se confirman las opciones recomendadas de Q1 y Q2; el sistema indica qué falta. Hasta que se resuelvan Q1 y Q2, este comportamiento es provisional.
- **Varios ministerios con el mismo encargado**: es válido y cada ministerio conserva sus propios datos; no hay ficha de encargado compartida (el encargado es un campo de cada ministerio, ver Key Entities).
- **Ministerio duplicado exacto** (mismo nombre, misma descripción, mismo encargado y mismo enlace): se trata como duplicado y se rechaza o se señala (default revisable, coherente con F2 y F3).
- **Dos ministerios con el mismo enlace de grupo de WhatsApp** (p. ej., un grupo de distribución compartido): se señala como aviso pero se permite (default revisable, ver Assumptions); un duplicado exacto sí se rechaza.
- **Nombre, descripción o encargado con caracteres especiales del español** (tildes, ñ, ¿ ¡) **y con emojis**: se conservan íntegros en el panel y en el sitio público, sin corromperse ni perderse.
- **Descripción más larga que el límite de extensión** (2.000 caracteres por defecto, default revisable): el sistema lo indica al excederse y no guarda hasta corregirlo; los textos largos deben acomodarse sin romper la presentación en ningún dispositivo.
- **Imagen opcional con archivo inválido o carga fallida**: se indica el problema con el archivo y el ministerio se conserva sin imagen; la sección pública no se rompe ni muestra huecos. Toda imagen cargada lleva texto alternativo accesible.
- **Contenido en inglés sin versión en español**: no debe poder guardarse; el español es el idioma base y obligatorio (Decisión 8, FR-014), y el sistema lo impide indicando qué falta.
- **Dos personas editando el mismo ministerio al mismo tiempo**: se conserva la última versión guardada completa (default revisable, coherente con F3); no debe quedar una mezcla incoherente de datos sin aviso.
- **Eliminación de un ministerio publicado**: desaparece de lo público de inmediato, la eliminación queda registrada en la auditoría y no hay papelera ni recuperación (default revisable, coherente con "sin historial de versiones" de F3).
- **Muchos ministerios publicados**: se muestran todos en el orden definido, sin límite artificial (default revisable); si en el futuro fueran demasiados, el agrupamiento o el "ver más" serían un ajuste menor acordado con el humano.
- **Operación sin permiso «ministerios»** (lectura de la gestión, creación, edición, cambio de estado, reordenamiento o eliminación): denegada con mensaje claro, igual que en F2 (FR-011); tener otro permiso de módulo no da acceso a este.
- **Cualquier operación del módulo que no se completa**: la operación no se aplica y el intento queda registrado en la auditoría con su resultado; nunca debe aplicarse una operación sin registro, ni registrarse una operación que no se aplicó (FR-017, coherente con F2).

### Errores esperados

Todos los mensajes deben ser comprensibles para personas no técnicas y nunca deben exponer información interna del sistema (coherente con F1, F2 y F3).

| Situación | Qué debe ver la persona |
|---|---|
| Campo obligatorio vacío al crear o editar un ministerio (p. ej., nombre, descripción o encargado en español) | Indicación de qué campo corregir, sin guardar nada |
| Enlace de WhatsApp mal formado | Indicación de cómo corregir el enlace, sin guardar nada |
| Enlace de WhatsApp con un destino no seguro (p. ej., `javascript:`) u otro formato no admitido | Indicación de que el enlace no es válido o seguro y cómo corregirlo, sin guardar nada |
| Intento de guardar un ministerio sin versión en español (idioma base obligatorio) | Mensaje que indica que falta el contenido en español |
| Nombre, descripción o encargado que excede su límite de extensión | Mensaje con el límite vigente y lo que falta recortar |
| Imagen con archivo inválido o carga fallida | Mensaje que indica el problema con el archivo; el ministerio se conserva sin imagen y la sección no se rompe |
| Operación del módulo sin el permiso «ministerios» | Mensaje claro de acceso denegado, sin detalles internos |
| Intento de eliminar un ministerio | Confirmación explícita antes de eliminar, explicando que la operación no se puede deshacer |
| Sección sin ningún ministerio publicado | La sección se oculta; en ningún caso se muestra vacía ni con campos en blanco |
| Error inesperado del sistema al crear, editar, publicar o eliminar | Mensaje genérico sin información interna, sin perder lo que la persona ya haya escrito sin aviso |

## Requirements *(mandatory)*

### Functional Requirements

**Definición del concepto (Ministerios ≠ Grupos de conexión)**

- **FR-001**: Un **ministerio** DEBE representar un **área de servicio de la iglesia** con un **encargado responsable** y un **grupo de WhatsApp de distribución** al que la persona se une (Decisión 2, `idea.md` §1). Un ministerio NO DEBE confundirse ni mezclarse con los **grupos de conexión** de F5 (talleres y clases con horario, lugar y contacto para inscribirse) ni con los eventos o las actividades de F4: la sección pública y el módulo del panel DEBEN presentar y gestionar **únicamente ministerios**, y F6 NO DEBE construir ninguna funcionalidad de F5.

**Descubrimiento público (visitante, sin autenticación)**

- **FR-002**: El sitio DEBE mostrar una sección pública accesible sin autenticación con el **listado de ministerios**, donde cada ministerio presente, como mínimo, su **nombre**, su **descripción**, su **encargado** y su **enlace al grupo de WhatsApp para unirse**. Un visitante DEBE poder identificar de qué se trata cada ministerio, quién lo encarga y cómo unirse sin buscar en ningún otro sitio.
- **FR-003**: El enlace al grupo de WhatsApp DEBE poder activarse **con un solo clic o toque** desde la ficha del ministerio y DEBE abrir el destino del enlace (WhatsApp o la página de invitación del grupo) para que la persona pueda unirse. La acción DEBE presentarse con un texto claramente identificado como unirse al grupo (la redacción concreta la define la fase de UX). El sitio solo enlaza: NO DEBE haber inscripción, registro ni formulario en F6 (ver Out of Scope).
- **FR-004**: Los ministerios DEBEN presentarse en un **orden estable definido por el equipo** (Q5), sin exigir al visitante ninguna acción extra (ordenar, filtrar ni buscar) para ver el listado completo.
- **FR-005**: Cuando la sección no tiene **ningún ministerio publicado** (o todos están en borrador), DEBE **ocultarse por completo** (patrón de F3, FR-013 de F3, y de F4): la navegación (barra, menú hamburguesa) y el mapa del sitio del pie DEBEN quedar sin enlaces rotos ni huérfanos, y el sitio NUNCA DEBE mostrar secciones vacías ni campos en blanco.
- **FR-006**: La sección DEBE vivir bajo **una única entrada de navegación del sitio público —«Ministerios»—** en la ruta pública **`/ministerios`** (coherente con el patrón de F4, cuya ruta es `/eventos`), presente en la barra o menú hamburguesa y en el mapa del sitio del pie (FR-021).

**Bilingüismo (Decisión 6 y Decisión 8)**

- **FR-007**: El contenido de cada ministerio DEBE poder ingresarse en español (idioma base, obligatorio) y opcionalmente en inglés, y la interfaz del sitio público DEBE estar siempre disponible en ambos idiomas (Decisión 6 y Decisión 8).
- **FR-008**: Cuando un ministerio no tiene versión en inglés y el visitante navega en inglés, el sistema DEBE mostrar la versión en español (idioma base); NUNCA DEBE mostrar un campo vacío ni una traducción automática de un contenido que no fue traducido (patrón de F3, FR-009 de F3). El **enlace al grupo de WhatsApp NO es contenido traducible**: es un único destino igual en ambos idiomas.
- **FR-009**: El idioma se gestiona con el selector ya existente (FR-010 de F3 y F10): el cambio se aplica de inmediato, se mantiene al navegar por el sitio y se recuerda entre visitas en el mismo dispositivo; un dispositivo sin preferencia guardada se muestra en español.

**Administración desde el panel (patrón de F2 y F3)**

- **FR-010**: El panel DEBE permitir **crear, editar y eliminar ministerios** con su versión en español y su versión opcional en inglés (FR-007), y DEBE permitir mantener cada ministerio en estado borrador o publicado (FR-012).
- **FR-011**: La gestión de ministerios DEBE verificar el permiso **«ministerios»**, ya **existente en el catálogo fijo de permisos de F2** (FR-015 de F2; verificado en la siembra del catálogo de F2: `('ministerios', 'Ministerios')`). **F6 NO crea permisos nuevos ni modifica el modelo de permisos de F2** (a diferencia de F4, que añadió «categorías» por decisión del 2026-10-11): F6 **activa** el permiso reservado y nada más. La verificación DEBE seguir el patrón de autorización de F2, **verificada en el servidor** según el rol de la cuenta (Decisión 5, constitución §IV), y el permiso es **independiente de los demás**: tener «ministerios» no da acceso a otros módulos ni otros permisos dan acceso a este. Toda operación no autorizada DEBE denegarse con un mensaje claro.
- **FR-012**: Cada ministerio DEBE poder mantenerse en estado **borrador** o **publicado** y DEBE poder publicarse o retirarse **por separado** (Decisión 4, patrón de F3). Solo lo publicado DEBE ser visible para el visitante; el sistema NUNCA DEBE exponer al público ningún ministerio en borrador ni sus datos, por ninguna vía. Cuando no haya ningún ministerio publicado, la sección DEBE ocultarse por completo (FR-005).
- **FR-013**: Cuando se guarda un cambio sobre un ministerio **ya publicado**, los cambios DEBEN hacerse visibles de inmediato para el visitante (patrón de F3: se publica al guardar); no hay paso intermedio de borrador para lo que ya está publicado.
- **FR-014**: El sistema DEBE validar toda entrada del módulo en el panel: campos obligatorios completos (nombre, descripción y encargado en español, y enlace de WhatsApp, según Q1 y Q2), versión en español siempre presente y versión en inglés opcional (Decisión 8), enlace de WhatsApp que cumpla FR-015 y textos dentro de sus límites de extensión (FR-019). Ante datos inválidos o incompletos DEBE indicar qué corregir y NUNCA DEBE guardar cambios parciales. Todo contenido ingresado DEBE tratarse como datos y NUNCA DEBE ejecutarse como código ni insertar elementos activos (constitución §IV).
- **FR-015**: El **enlace al grupo de WhatsApp** DEBE ser una dirección **válida y segura**: solo DEBE aceptarse un enlace `https` con un destino que cumpla la validación acordada (Q3), y el sistema DEBE **rechazar** cualquier enlace con esquema o destino no seguro —`javascript:`, `data:`, `vbscript:` o cualquier otro distinto de `https`—, así como cualquier entrada que al renderizarse pueda ejecutarse como código. El enlace mostrado al visitante DEBE ser exactamente el enlace guardado tras la validación, sin transformaciones que alteren su destino. La **vigencia** del destino no se verifica (ver Out of Scope).
- **FR-016**: El panel DEBE permitir **definir el orden de presentación** de los ministerios (Q5) y el orden elegido DEBE ser el que ve el visitante (FR-004).
- **FR-017**: Cada operación de este módulo —crear, editar, cambiar el estado de publicación, reordenar (Q5) y eliminar un ministerio— DEBE quedar registrada en el **registro de auditoría** siguiendo el **patrón ya construido en F2/F3/F4** (quién la hizo, qué hizo, sobre qué elemento y cuándo, incluidos los intentos que no se completan; registro de solo lectura que no se edita ni se borra). No se re-especifica ese mecanismo aquí.
- **FR-018**: La interfaz del panel para este módulo se entrega en español, como el resto del panel definido en F2 (Decisión 6 aplicada al sitio público).

**Contenido de cada ministerio**

- **FR-019**: Cada ministerio DEBE poder mantener: **nombre**, **descripción**, **encargado** (el nombre de la persona o equipo responsable; ver Q1) y **enlace al grupo de WhatsApp** para unirse (ver Q2), más **imagen opcional** (FR-020); sus textos DEBEN ser traducibles de forma opcional al inglés (FR-007), salvo el enlace (FR-008). La descripción DEBE admitir texto multilínea con los caracteres propios del español (tildes, ñ, ¿ ¡) y emojis sin transformación, dentro del límite de extensión de FR-014.
- **FR-020**: La imagen de un ministerio DEBE ser opcional: un ministerio sin imagen DEBE presentarse completo y coherente. Toda imagen cargada DEBE llevar texto alternativo accesible y DEBE respetar las reglas de identidad vigentes (no deformar el logotipo ni recolorearlo, FR-002 de F3 y FR-024 de F10).

**Relación con el sitio público (F3 y F10), responsividad y accesibilidad**

- **FR-021**: La sección DEBE integrarse en el sitio público **naciendo con el sistema de diseño y movimiento de F10** (mismos componentes, tipografía, color con modo claro/oscuro, espaciado y movimiento respetando `prefers-reduced-motion`) y con la navegación de F10 (barra fija visible en escritorio, menú hamburguesa accesible en móvil, mapa del sitio del pie), **sin rediseñar ni re-especificar ese sistema**. La sección DEBE vivir bajo la única entrada de navegación «Ministerios» (FR-006) y NO DEBE tocar ni alterar la portada de F3 ni ninguna otra sección (ver Out of Scope).
- **FR-022**: La sección DEBE adaptarse a cualquier dispositivo (teléfono, tableta y computadora) sin pérdida de contenido ni de funciones, y DEBE ser navegable por personas de todas las edades y niveles de experiencia: títulos claros, navegación simple, texto legible y elementos interactivos de tamaño suficiente (`idea.md` §7).
- **FR-023**: La sección DEBE ser accesible: navegación completa por teclado, compatibilidad con lectores de pantalla, contraste suficiente (en tema claro y oscuro), posibilidad de ampliar el texto hasta el 200 % sin perder información y texto alternativo en las imágenes. El nivel de referencia es WCAG 2.1 nivel AA (default revisable, ver Assumptions).

### Key Entities *(include if feature involves data)*

- **Ministerio** *(área de servicio de la iglesia, Decisión 2)*: uno o varios. Cada ministerio tiene **nombre**, **descripción**, **encargado** (el nombre de la persona o equipo responsable del área —ver Q1; no es un directorio de personas: es un campo del ministerio—) y **enlace al grupo de WhatsApp** para unirse (ver Q2 y FR-015), más **imagen opcional** con texto alternativo (FR-020), **orden de presentación** definido por el equipo (Q5, FR-004, FR-016), **estado de publicación** (borrador/publicado) y versión en español (obligatoria) e inglés (opcional) de sus textos. Un ministerio **no** es un grupo de conexión (F5) ni un evento o una actividad (F4) (FR-001).
- **Enlace al grupo de WhatsApp** *(destino único, no traducible)*: el medio por el que la persona se une al grupo de distribución del ministerio (FR-003, FR-008). DEBE ser un enlace `https` válido y seguro (FR-015, Q3); su vigencia no se verifica. El sitio solo enlaza: no administra grupos ni envía mensajes.
- **Estado de publicación** (borrador/publicado, por ministerio): cada ministerio se publica o retira por separado (Decisión 4) y decide si el visitante lo ve o no; la sección sin ministerios publicados se oculta por completo (FR-005).

Relaciones: cada ministerio se identifica por su nombre, tiene un encargado responsable y un único enlace de unión por grupo de WhatsApp (Q2), y presenta sus textos en español (obligatorio) e inglés (opcional con fallback al español, FR-008). Lo gestiona una cuenta del panel con el permiso **«ministerios»** (F2, ya existente en el catálogo: F6 no crea permisos nuevos) y cada operación sobre él queda registrada en la auditoría siguiendo el patrón de F2/F3/F4 (FR-017). El visitante solo ve lo publicado (FR-012), en el orden definido por el equipo (FR-004), y se une al grupo activando el enlace (FR-003). La composición concreta de la presentación pública (listado, tarjetas, ficha de detalle) la define la fase de UX.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de los visitantes puede ver los ministerios publicados (nombre, descripción, encargado y enlace al grupo de WhatsApp) sin autenticación, y encuentra cómo unirse en menos de 30 segundos desde que abre la sección.
- **SC-002**: 0 ministerios en borrador visibles para el público en las pruebas de aceptación (incluida cualquier respuesta del sistema hacia el visitante).
- **SC-003**: Todo cambio guardado sobre un ministerio publicado es visible en el sitio público desde la primera carga posterior al guardado (se publica al guardar), sin pasos adicionales.
- **SC-004**: Un miembro del equipo crea y publica un ministerio nuevo en menos de 2 minutos desde el panel.
- **SC-005**: El 100% de las operaciones del módulo verifican el permiso **«ministerios»** en el servidor: 0 operaciones completadas sin permiso en las pruebas de aceptación (incluidos los intentos forzados y el cruce con otros permisos) y 100% denegadas con mensaje claro.
- **SC-006**: Con el idioma en inglés, el 100% de los textos de la interfaz está traducido y 0 campos aparecen vacíos por falta de traducción: cada ministerio sin versión en inglés se muestra íntegramente en español.
- **SC-007**: La sección es usable en teléfono, tableta y computadora: 0 pérdida de elementos o funciones entre dispositivos y 0 desplazamiento horizontal en el ancho de pantalla más estrecho probado (320 px).
- **SC-008**: La sección supera una revisión de accesibilidad con 0 hallazgos críticos (referencia WCAG 2.1 nivel AA, en tema claro y oscuro) y es recorrerla y operarla por completo con teclado y con lector de pantalla.
- **SC-009**: El 100% de los enlaces de WhatsApp publicados se abren con un solo clic o toque y llevan al grupo correcto, y **0 enlaces no seguros** (esquemas distintos de `https`, `javascript:`, `data:`…) son aceptados al guardar ni renderizados al visitante en las pruebas de aceptación.
- **SC-010**: 0 secciones vacías ni campos en blanco: la sección sin ministerios publicados se oculta por completo y la navegación (barra, menú hamburguesa y mapa del sitio) queda sin enlaces rotos ni huérfanos.
- **SC-011**: El 100% de las operaciones del módulo de las pruebas de aceptación (crear, editar, publicar/retirar, reordenar y eliminar) queda registrado en la auditoría siguiendo el patrón de F2/F3/F4 (quién, qué, sobre qué y cuándo), y 0 registros de auditoría pueden editarse ni borrarse.
- **SC-012**: En pruebas con personas de distintas edades, incluidas con poca experiencia en internet, al menos el 90% identifica qué ministerio es, quién lo encarga y cómo unirse, sin ayuda.
- **SC-013**: 0 grupos de conexión (F5) ni eventos o actividades (F4) aparecen en la sección de ministerios en las pruebas de aceptación: la distinción conceptual de FR-001 se observa en el 100% de los casos probados.

## Out of Scope

Queda explícitamente **fuera del alcance** de F6:

- **Grupos de conexión (F5)**: los talleres y clases con horario, lugar y contacto para inscribirse (maquillaje, fotografía, inglés, asados…, `idea.md` §5) son el catálogo de F5, **no** ministerios. F6 no los muestra, no los administra y **no construye ninguna parte de F5** (FR-001).
- **Eventos y actividades (F4)**: los encuentros periódicos y las acciones puntuales y su agenda son de F4; F6 no los muestra ni los administra. Un ministerio **no** es un evento: no tiene fechas, horarios ni retiro automático por fecha.
- **Formularios de inscripción, registro o solicitud de unión**: la unión al ministerio se resuelve con el enlace al grupo de WhatsApp (FR-003); cualquier formulario de inscripción es un candidato de fase 3 (patrón de la Decisión 1 para grupos de conexión).
- **Administración de grupos de WhatsApp, envío de mensajes o publicación automática en redes sociales** (`idea.md` §8): el sitio solo presenta y enlaza.
- **Verificación de la vigencia del enlace de WhatsApp**: el sistema solo valida el formato y la seguridad del enlace (FR-015); no comprueba que el grupo siga activo ni que la invitación siga vigente. La corrección es manual, editando el enlace.
- **Directorio de encargados ni fichas de personas**: el encargado es un campo de cada ministerio (ver Key Entities y Q1); no hay perfiles de encargados, sus horarios ni sus datos personales de contacto más allá de lo que decida Q1.
- **Contenido adicional del ministerio** (horario y lugar de reunión del área, redes propias del ministerio, número de miembros, fotos del equipo): fuera salvo que el cliente lo amplíe en Q6.
- **Bloque de ministerios en la portada de F3**: la sección vive solo en `/ministerios`; la portada de F3 no se toca (ver Q7 y FR-021).
- **Traducción automática de contenidos**: el inglés se ingresa manualmente; si un ministerio no tiene inglés, se muestra en español (FR-008).
- **Historial de versiones del contenido, papelera ni recuperación de ministerios eliminados, y programación de publicación con fecha futura** (default revisable, coherente con F3).
- **Donaciones (F7), noticias y galería (F8) y medios (F9)**: conceptos y módulos distintos; F6 no los muestra ni los administra.
- **Gestión de usuarios, roles y permisos** (F2): F6 solo **activa** el permiso **«ministerios»** ya existente en el catálogo de F2 (FR-015 de F2): **no crea permisos nuevos** (a diferencia de F4, que añadió «categorías») y queda fuera cualquier otro cambio al modelo de permisos o a la gestión de cuentas y roles (Decisión 9: esa gestión sigue siendo F2, intacta).
- **Rediseño o sistema de diseño** (F10 y F10B): F6 **aplica** el sistema de diseño y movimiento de F10 ya entregado (sitio público) y nace en el panel con los patrones actuales de F2/F3/F4; el rediseño visual del panel es F10B y no forma parte de F6. Tampoco se re-especifica la navegación, el selector de idioma ni el modo oscuro (F10).
- **Bilingüismo del panel de administración**: el panel se entrega en español (como F2, F3 y F4); la Decisión 6 aplica al sitio público.
- **Despliegue a producción** y cambios de infraestructura: sigue pendiente a nivel de proyecto.

## Assumptions

Los puntos que cambian el alcance o la experiencia se formulan como preguntas abiertas en *Aclaraciones* (Q1–Q7), **cada una con su opción recomendada** para que el humano solo tenga que confirmarla. Lo demás son defaults razonables tomados donde la descripción no fija una regla; **son provisionales y revisables** por el humano.

**Decisiones ya establecidas que F6 aplica sin reabrirlas**

- **Estados borrador/publicado** (Decisión 4, patrón de F3 y F4): publicación por ministerio; solo lo publicado es visible; se publica al guardar; la sección sin publicados se oculta por completo (FR-005, FR-012, FR-013).
- **Permisos por módulo agrupados en roles** (Decisión 5): F6 activa el permiso **«ministerios»** ya existente en el catálogo fijo de F2 (verificado en FR-015 de F2 y en la siembra del catálogo de F2); **F6 no crea permisos nuevos** (a diferencia de F4) y verifica la autorización en el servidor (FR-011, constitución §IV).
- **Bilingüismo** (Decisiones 6 y 8): español base obligatorio, inglés opcional por contenido con fallback al español; la interfaz del sitio público siempre bilingüe; el panel en español (FR-007 a FR-009, FR-018).
- **Auditoría**: F6 sigue el **patrón ya construido en F2/F3/F4** (registro de solo lectura con quién, qué, sobre qué y cuándo, incluidos los intentos fallidos) sin re-especificarlo (FR-017).
- **Sistema de diseño y movimiento**: la sección nace con el sistema de F10 aplicado (modo claro/oscuro, movimiento con `prefers-reduced-motion`, navegación y mapa del sitio); no se re-especifica aquí (FR-021).

**Defaults revisables**

- **Campos obligatorios**: nombre, descripción, encargado y enlace de WhatsApp en español (salvo lo que decidan Q1 y Q2); imagen opcional; inglés opcional. La obligatoriedad concreta de encargado y enlace queda sujeta a Q1 y Q2.
- **Límites de extensión**: descripción de **2.000 caracteres** por defecto (coherente con F4); nombre y encargado con límites cortos de texto (default revisable, el arquitecto fija el detalle).
- **Un solo enlace de WhatsApp por ministerio** (Q2): el enlace es un único destino de unión; si el cliente necesita varios, se amplía la spec.
- **Orden de presentación**: hasta que se resuelva Q5, el orden es estable y definible por el equipo (FR-016); el desempate por nombre es un default revisable.
- **Duplicados**: un duplicado exacto (mismo nombre, misma descripción, mismo encargado y mismo enlace) se rechaza o se señala como duplicado (default revisable, coherente con F2 y F3); dos ministerios con el mismo enlace de grupo se señalan como aviso pero se permiten (default revisable).
- **Eliminación de ministerios**: definitiva, con confirmación explícita y registrada en la auditoría; no hay papelera ni recuperación (default revisable, coherente con F3).
- **Ediciones simultáneas**: si dos personas editan el mismo ministerio al mismo tiempo, se conserva la última versión guardada completa (default revisable, coherente con F3).
- **Sin límite artificial de ministerios** ni filtros ni búsqueda en el MVP: se muestran todos en el orden definido (default revisable).
- **Validación de enlaces**: además de la seguridad (FR-015), se valida el formato; no se comprueba que el destino exista ni que el grupo siga activo.
- **Idioma del panel**: se entrega en español, como F2, F3 y F4; el bilingüismo del roadmap aplica al sitio público (Decisión 6).
- **Nivel de accesibilidad**: se toma **WCAG 2.1 nivel AA** como referencia (default revisable); el requisito de fondo es el de `idea.md` §7: apto para personas de todas las edades y dispositivos.
- **Dependencias**: F2 terminada (panel, roles con permisos por módulo, **permiso «ministerios» ya reservado en FR-015 de F2 y sembrado en su catálogo** —F6 no crea permisos nuevos—, patrón de autorización y registro de auditoría), F3 terminada (sitio público, bilingüismo con fallback, estados borrador/publicado, ocultar secciones vacías) y F10 terminada (sistema de diseño y movimiento, navegación, selector de idioma, modo oscuro y mapa del sitio del pie). F6 es independiente de F4, F5 y F7.

## Aclaraciones

Puntos abiertos que no se pueden resolver sin inventar reglas de negocio. Cada uno incluye la **opción recomendada** para que el humano solo tenga que confirmarla o corregirla. Hasta que se resuelvan, el alcance afectado queda marcado como provisional en los requisitos correspondientes.

### Q1 — Encargado: contenido del campo y obligatoriedad (FR-002, FR-019, Key Entities)

**Pregunta**: ¿el encargado de un ministerio es solo un **campo de texto con el nombre** de la persona (o equipo) responsable, o incluye también un **medio de contacto** (teléfono, correo)? ¿Puede un ministerio **publicarse sin encargado**?

**Opción recomendada**: campo de texto único **obligatorio** con el nombre de la persona o equipo responsable, **sin datos personales de contacto** (el canal de unión es el grupo de WhatsApp, FR-003); un ministerio sin encargado **no se publica**. Con esto se cumple "su encargado" del roadmap sin exponer datos personales ni construir un directorio de personas. *(Alternativa: encargado opcional, o encargado con contacto público.)*

### Q2 — Enlace al grupo de WhatsApp: obligatoriedad y cantidad (FR-003, FR-019, Key Entities)

**Pregunta**: ¿el enlace al grupo de WhatsApp es **obligatorio** para publicar un ministerio? ¿Puede un ministerio tener **varios** enlaces (p. ej., grupo general y grupo de jóvenes)?

**Opción recomendada**: **un único enlace obligatorio** por ministerio (sin enlace no se publica, porque "el enlace al grupo de WhatsApp para unirse" es el valor central de F6). *(Alternativa: enlace opcional para ministerios sin grupo aún, o varios enlaces por ministerio; cualquiera de las dos amplía el alcance y se documentaría en esta spec.)*

### Q3 — Alcance de la validación del enlace de WhatsApp (FR-015)

**Pregunta**: además de exigir un enlace `https` **válido y seguro** (rechazando `javascript:`, `data:`, `vbscript:` y cualquier esquema distinto de `https`), ¿se **restringe el destino a dominios de WhatsApp** (`chat.whatsapp.com` para invitaciones de grupo, `wa.me` / `whatsapp.com`) o se acepta **cualquier dirección `https` válida**?

**Opción recomendada**: **restringir a dominios de WhatsApp** (`chat.whatsapp.com`, `wa.me`, `whatsapp.com`), rechazando cualquier otro destino con indicación de cómo corregirlo: es un "grupo de WhatsApp para unirse" y así se evita que un campo llamado "enlace de WhatsApp" lleve a cualquier parte. Si el equipo usa enlaces acortados o de terceros, la restricción se amplía con esos dominios o se baja a "cualquier `https`" con confirmación del humano. *(Alternativa: aceptar cualquier `https` válido y seguro; más flexible, menos estricto.)*

### Q4 — Imagen o ícono del ministerio (FR-019, FR-020)

**Pregunta**: ¿el ministerio lleva una **imagen cargada por el equipo** (patrón de F3 y F4), un **ícono elegido de un catálogo** con color de la marca, o **sin elemento visual**?

**Opción recomendada**: **imagen opcional cargada por el equipo**, con texto alternativo accesible y reglas de identidad (patrón de F3 y F4); sin catálogo de íconos en F6. Un ministerio sin imagen se presenta completo y coherente (FR-020). *(Alternativas: catálogo de íconos con color de marca —más trabajo de UX—; o sección sin imágenes —más austera—.)*

### Q5 — Orden de presentación de los ministerios (FR-004, FR-016)

**Pregunta**: ¿en qué orden ve el visitante los ministerios: **orden manual definido por el equipo** (p. ej., un campo de posición o reordenamiento en el panel), **orden alfabético por nombre** u **orden de creación**?

**Opción recomendada**: **orden manual definido desde el panel** (FR-016), con el nombre como desempate estable: suele interesar destacar primero áreas como alabanza o acción social sin depender del alfabeto. *(Alternativas: alfabético —más simple y sin mantenimiento—; orden de creación —predecible pero poco expresivo—.)*

### Q6 — Contenido adicional del ministerio (FR-019, Out of Scope)

**Pregunta**: ¿el ministerio muestra además **horario y lugar de reunión del área**, **redes sociales propias**, **número de miembros**, **fotos del equipo** u otros datos?

**Opción recomendada**: **no en F6**: solo nombre, descripción, encargado, enlace de WhatsApp e imagen opcional (los campos de la fila del roadmap y de `idea.md`). Si el cliente quiere alguno de estos datos, se amplía esta spec **antes** de la fase de plan para no reabrir el alcance después. *(Alternativa: incluir desde ya, p. ej., horario/lugar de reunión del ministerio como campos opcionales.)*

### Q7 — ¿Bloque de ministerios en la portada de F3? (FR-021, Out of Scope)

**Pregunta**: como F4 añadió el bloque «Lo que viene» a la portada (2026-10-12), ¿quiere el cliente un **bloque resumen de ministerios en la portada** (p. ej., "Conoce nuestros ministerios" con algunos ítems), o la sección vive **solo** en `/ministerios`?

**Opción recomendada**: **no**: la sección vive solo en `/ministerios` con su entrada de navegación, y **la portada de F3 no se toca** (FR-021). Un bloque en la portada sería un cambio de alcance como el de F4 (con su impacto en UX y pruebas) y hoy no está pedido en el roadmap ni en `idea.md`. *(Alternativa: bloque resumen en la portada, que se documentaría como cambio de alcance.)*

### Sobre el resto del documento

Fuera de Q1–Q7 no hay ambigüedades abiertas: los estados borrador/publicado (Decisión 4), los permisos por módulo —con el permiso **«ministerios» ya existente en el catálogo de F2 y sin permisos nuevos para F6 (Decisión 5, FR-015 de F2)—, el bilingüismo español base / inglés opcional con fallback (Decisiones 6 y 8), la auditoría de solo lectura (patrón de F2/F3/F4), la ruta pública `/ministerios` y la entrada única de navegación (patrón de F4 `/eventos`), la regla de ocultar la sección vacía (patrón de F3 y F4), la validación de enlaces `https` seguros sin `javascript:` (FR-015, constitución §IV), la responsividad y accesibilidad (`idea.md` §7) y la aplicación del sistema de diseño de F10 están definidos por decisiones o specs ya aprobadas y se recogen aquí sin reabrirlas.
