# Feature Specification: Eventos y actividades (F4)

**Feature Branch**: `004-eventos-actividades`

**Created**: 2026-10-11

**Status**: Cambio de alcance aprobado (2026-10-11) — permiso «categorías» añadido; pendiente de re-aprobación humana

**Input**: User description: "F4 — Eventos y actividades (roadmap §3, `docs/producto/roadmap.md`): El visitante descubre lo que viene: **eventos** periódicos de la iglesia (noche de aposentos, vigilia) y **actividades** puntuales (repartir comida, predicar en tal lugar), con fecha, lugar y detalle; el equipo administra ambos por separado. Es el corazón del problema: que nadie se quede sin enterarse." Decisiones aplicables del roadmap: **Decisión 3** (2026-09-28: **eventos** = encuentros periódicos de la iglesia; **actividades** = acciones puntuales; son conceptos distintos, modelados por separado dentro de F4), **Decisión 4** (2026-09-28: contenido con estados borrador/publicado; solo lo publicado es visible al visitante), **Decisión 5** (2026-09-28: permisos por módulo agrupados en roles; el catálogo de roles es libre y cada permiso autoriza la gestión completa de su módulo), **Decisión 6** (2026-09-28: el sitio público es bilingüe, español e inglés, con selección de idioma) y **Decisión 8** (2026-09-28: cada contenido puede ingresarse en ambos idiomas o solo en español; el español es el idioma base y el inglés es opcional por contenido; la interfaz siempre es bilingüe). Restricción de `idea.md` §7 y del encargo (`idea.md` intro): el sitio debe adaptarse a cualquier dispositivo y ser apto para personas de todas las edades, incluidas adultos mayores con lo básico de internet. **Problema que resuelve** (`idea.md` §1 y §3): hoy las actividades "se olvida[n]" porque solo se comparte un afiche de vez en cuando en WhatsApp o Facebook, y así "no asisten personas nuevas porque no saben nada". **Dependencias**: F2 (terminada) provee el panel de administración, los roles con permisos por módulo, el patrón de autorización y el registro de auditoría; la spec de F2 **ya reservó los permisos «eventos» y «actividades»** —por separado— dentro de su catálogo (FR-015 de F2) y esta funcionalidad **los activa**; además, F4 **añade el permiso nuevo «categorías»** al catálogo de F2 —ampliándolo de **9 a 10 permisos**— por **decisión explícita del humano del 2026-10-11** (cambio de alcance autorizado que sustituye la premisa previa de que F4 no creaba permisos nuevos). F4 activa así **tres** permisos del catálogo: «eventos», «actividades» y «categorías». F3 (terminada) provee el sitio público, el bilingüismo con fallback al español, los estados borrador/publicado y el patrón de ocultar por completo las secciones sin contenido publicado (FR-013 de F3). F10 (terminada) provee el sistema de diseño y movimiento, la navegación fija (barra visible en escritorio, menú hamburguesa accesible en móvil), el selector de idioma compacto, el modo oscuro y el mapa del sitio del pie: las secciones nuevas de F4 **nacen ya con ese sistema aplicado** y este documento **no lo re-especifica** (ver Out of Scope). **Aclaraciones resueltas con el humano el 2026-10-11**: las 6 preguntas abiertas de la primera versión (Q1–Q6) quedan resueltas por la opción recomendada en cada caso (ver *Aclaraciones resueltas (2026-10-11)*).

> **Nota de alcance y terminología**: esta funcionalidad es la **primera sección de contenido dinámico del sitio público** después de la portada. Se reutiliza la terminología de F2 y F3: el **visitante** (el "usuario común" de `idea.md` §2: ve el sitio público, no tiene cuenta y no se registra) y el **usuario del panel** (el "usuario del sistema" y el "usuario administrador" de `idea.md` §2: tiene cuenta, inicia sesión y gestiona el contenido). F4 entrega: la presentación pública de **lo que viene** —**eventos** (encuentros periódicos: noche de aposentos, vigilia…) y **actividades** (acciones puntuales: repartir comida, predicar en tal lugar…)— con título, fecha(s), lugar y detalle, en español e inglés (español base, inglés opcional con fallback), y la gestión **por separado** de ambos conceptos desde el panel, con los permisos por módulo «eventos» y «actividades» ya reservados en F2 y el **permiso nuevo «categorías»** —añadido al catálogo de F2 (9→10 permisos) por decisión explícita del humano del 2026-10-11—, que autoriza la administración del **catálogo de categorías** (no «eventos» ni «actividades»), estados borrador/publicado y registro de auditoría. F4 activa así **tres permisos del catálogo: «eventos», «actividades» y «categorías»**, independientes entre sí. Los eventos y las actividades son **modelos distintos** (Decisión 3): comparten presentación pública pero se administran en módulos separados y nunca se mezclan en un solo listado sin identificación. La **ubicación** de la sección quedó fijada en las aclaraciones del 2026-10-11: una única entrada de navegación del sitio público —«Eventos y actividades»— con dos bloques separados, sin tocar la portada de F3 (Q5, FR-021); la **presentación visual** de cada elemento (lista, ficha, tarjetas) corresponde a la fase de UX (ver Out of Scope). Los demás módulos del sitio (grupos de conexión, ministerios, donaciones, noticias, medios) siguen siendo F5–F9 y **no** forman parte de F4.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Descubrir lo que viene: próximos eventos y actividades (Priority: P1)

Como visitante (miembro de la iglesia o persona nueva), quiero ver en el sitio público los próximos eventos y las próximas actividades de la iglesia —con título, fecha, lugar y detalle—, para enterarme de lo que viene, saber cuándo y dónde asistir y no perderme nada.

**Why this priority**: Es el corazón de F4 y de todo el producto según `idea.md` §1 y §3: "la gran mayoría del tiempo se olvida [las actividades] y no asisten personas nuevas porque no saben nada". Sin esta historia no hay valor visible para el público y la gestión del panel (US2–US4) no tendría destinatario. Es la traducción directa del criterio de éxito del roadmap §8: "un canal unificado donde ver y conocer las actividades".

**Independent Test**: Se prueba de forma independiente con eventos y actividades publicados con fecha futura: un visitante sin sesión abre la sección y encuentra cada elemento con su título, su fecha, su lugar y su detalle, con los eventos (periódicos) y las actividades (puntuales) claramente identificados como conceptos distintos; al cambiar y publicar un dato desde el panel y recargar, el visitante ve el dato actualizado; y un elemento cuya fecha ya pasó no aparece entre "lo que viene". Entrega valor por sí sola: el público ya sabe qué va a pasar, cuándo y dónde.

**Acceptance Scenarios**:

1. **Dado** eventos y actividades publicados con fecha futura, **cuando** un visitante abre la sección sin iniciar sesión, **entonces** ve cada uno con su título, su fecha, su lugar y su detalle, y distingue sin confusión qué es un evento (encuentro periódico) y qué es una actividad (acción puntual) (Decisión 3).
2. **Dado** un evento periódico publicado (p. ej., una vigilia), **cuando** el visitante lo consulta, **entonces** identifica cuándo es la próxima vez que ocurre (la próxima fecha concreta cargada) y, de existir, el texto de frecuencia del evento (p. ej., «todos los viernes») (Q1).
3. **Dado** una actividad puntual publicada (p. ej., repartir comida), **cuando** el visitante la consulta, **entonces** identifica la fecha, el horario (hora de inicio y, si se indicó, hora de fin) y el lugar exactos de esa acción concreta (Q2).
4. **Dado** varios eventos y actividades publicados, **cuando** el visitante recorre la sección, **entonces** los ve en orden cronológico de lo más próximo a lo más lejano, sin tener que buscar cuál viene primero (Q3).
5. **Dado** un elemento publicado cuya fecha ya pasó, **cuando** el visitante abre la sección, **entonces** ese elemento no aparece entre "lo que viene": se retira automáticamente de la vista pública (Q3 y Q4).
6. **Dado** la sección publicada, **cuando** una persona nueva la recorre, **entonces** sabe qué va a pasar, cuándo, a qué hora y dónde, sin buscar en WhatsApp, en Facebook ni en ningún otro sitio.
7. **Dado** un evento y una actividad con la misma fecha, **cuando** el visitante los ve, **entonces** ambos aparecen, cada uno en su bloque de concepto, sin que uno oculte al otro.

---

### User Story 2 - Gestionar los eventos desde el panel (Priority: P1)

Como usuario del panel con el permiso de «eventos», quiero crear, editar y eliminar los eventos periódicos de la iglesia —con su título, sus fechas, su lugar y su detalle, en español y, si quiero, también en inglés— para mantener al día lo que anunciamos sin depender de un desarrollador ni de un afiche suelto.

**Why this priority**: Es una de las dos mitades de gestión de F4 y el criterio de éxito de `idea.md` §4: que el contenido sea dinámico, no estático. Los eventos (noche de aposentos, vigilia) son los encuentros periódicos que sostienen la vida de la congregación (Decisión 3) y su gestión es un módulo independiente con su propio permiso ya reservado en F2 (FR-015 de F2), por lo que es construible y testeable por separado de las actividades.

**Independent Test**: Se prueba de forma independiente con una cuenta con el permiso «eventos» y otra sin él: la primera crea un evento con datos válidos y el sistema lo conserva; con campos obligatorios vacíos, fechas mal formadas o inválidas o sin versión en español, el sistema indica qué corregir y no guarda nada a medias; edita y elimina un evento (con confirmación) y la eliminación queda registrada en la auditoría; la segunda cuenta recibe una denegación clara en toda operación del módulo, incluidos los intentos forzados por la interfaz. Entrega valor por sí sola: el equipo ya administra sus eventos.

**Acceptance Scenarios**:

1. **Dado** un usuario del panel con el permiso «eventos», **cuando** crea un evento con datos válidos (título, fecha(s), lugar y detalle en español) y guarda, **entonces** el sistema confirma la operación y conserva el evento.
2. **Dado** una cuenta sin el permiso «eventos», **cuando** intenta crear, editar o eliminar un evento, **entonces** el sistema impide la operación con un mensaje claro y no cambia nada.
3. **Dado** un formulario de evento con datos inválidos o incompletos (título vacío, sin ninguna fecha, fecha mal formada, detalle en español vacío, enlace de mapa mal formado), **cuando** envío el formulario, **entonces** el sistema indica qué corregir y no guarda ningún cambio parcial.
4. **Dado** un evento existente, **cuando** edito sus datos y guardo, **entonces** el sistema conserva los cambios, y si el evento estaba publicado se hacen visibles de inmediato para el visitante (patrón de F3: se publica al guardar).
5. **Dado** un evento existente, **cuando** lo elimino y confirmo la eliminación, **entonces** el evento desaparece del panel y del sitio público, y la eliminación queda registrada en el registro de auditoría de F2 (quién, qué y cuándo).
6. **Dado** el formulario de evento, **cuando** la persona ingresa la versión en inglés, **entonces** el sistema conserva esa versión junto a la versión en español.
7. **Dado** el formulario de evento, **cuando** la persona deja vacío el campo en inglés, **entonces** puede guardar igualmente, porque el inglés es opcional por contenido (Decisión 8).
8. **Dado** una cuenta con el permiso «eventos» pero sin el permiso «actividades», **cuando** intenta forzar el acceso a la gestión de actividades, **entonces** el sistema lo impide: los permisos de F4 son independientes entre sí (Decisión 5).

---

### User Story 3 - Gestionar las actividades desde el panel (Priority: P1)

Como usuario del panel con el permiso de «actividades», quiero crear, editar y eliminar las actividades puntuales de la iglesia —repartir comida, predicar en tal lugar, una jornada de servicio…— con su título, su fecha, su lugar y su detalle, en español y opcionalmente en inglés, para anunciar cada acción concreta sin depender de nadie más.

**Why this priority**: Es la segunda mitad de gestión de F4 y la contraparte obligada de US2 por la Decisión 3: las actividades son acciones puntuales con modelo propio (fecha única frente a periodicidad) y permiso propio («actividades», ya reservado en F2). Va junto a US2 en P1 porque el valor del producto exige que **ambos** conceptos se administren: un sitio con solo eventos no cumple el encargo de `idea.md` §1 ("acción social, prédicas, adoraciones, entrevistas, noches de oración").

**Independent Test**: Se prueba de forma independiente con una cuenta con el permiso «actividades» y otra sin él: la primera crea una actividad puntual válida y el sistema la conserva con su fecha única; las validaciones rechazan datos inválidos sin guardar nada a medias; edita y elimina una actividad (con confirmación) con registro en auditoría; la segunda cuenta recibe denegación clara en toda operación del módulo y una cuenta solo con «eventos» tampoco puede gestionar actividades. Entrega valor por sí sola: el equipo ya administra sus actividades puntuales.

**Acceptance Scenarios**:

1. **Dado** un usuario del panel con el permiso «actividades», **cuando** crea una actividad con datos válidos (título, fecha, lugar y detalle en español) y guarda, **entonces** el sistema confirma la operación y conserva la actividad.
2. **Dado** una cuenta sin el permiso «actividades», **cuando** intenta crear, editar o eliminar una actividad, **entonces** el sistema impide la operación con un mensaje claro y no cambia nada.
3. **Dado** un formulario de actividad con datos inválidos o incompletos (título vacío, fecha vacía o mal formada, detalle en español vacío, enlace de mapa mal formado), **cuando** envío el formulario, **entonces** el sistema indica qué corregir y no guarda ningún cambio parcial.
4. **Dado** una actividad existente, **cuando** edito sus datos y guardo, **entonces** el sistema conserva los cambios, y si estaba publicada se hacen visibles de inmediato para el visitante (patrón de F3: se publica al guardar).
5. **Dado** una actividad existente, **cuando** la elimino y confirmo la eliminación, **entonces** desaparece del panel y del sitio público, y la eliminación queda registrada en la auditoría de F2.
6. **Dado** el formulario de actividad, **cuando** la persona deja vacío el campo en inglés, **entonces** puede guardar igualmente (Decisión 8); si lo llena, el sistema conserva ambas versiones.
7. **Dado** una cuenta con el permiso «actividades» pero sin el permiso «eventos», **cuando** intenta forzar el acceso a la gestión de eventos, **entonces** el sistema lo impide.

---

### User Story 4 - Controlar la publicación de cada evento y actividad (Priority: P2)

Como usuario del panel con el permiso del módulo correspondiente, quiero publicar o retirar cada evento y cada actividad por separado —dejarlos en borrador hasta que estén revisados—, para que el público solo vea información confirmada y nunca cambios a medias.

**Why this priority**: La Decisión 4 (contenido borrador/publicado) es parte explícita del alcance y protege la imagen pública de la iglesia: sin control de publicación, cada borrador sería público de inmediato. Va como P2 porque US2 y US3 solas ya entregan un MVP usable (crear y conservar contenido), mientras que esta historia añade el control de visibilidad elemento a elemento. Reutiliza literalmente el patrón probado en F3 (FR-013 y FR-014 de F3).

**Independent Test**: Se prueba de forma independiente con un evento en borrador, una actividad en borrador y un elemento publicado: los borradores no aparecen para el visitante y el resto de la sección se ve coherente; al publicarlos aparecen; al devolverlos a borrador desaparecen; una sección cuyo contenido está todo en borrador se oculta por completo (navegación y mapa del sitio incluidos, sin enlaces rotos); y ninguna respuesta del sistema hacia el público expone un borrador. Entrega valor por sí sola: el equipo decide qué se hace público y cuándo.

**Acceptance Scenarios**:

1. **Dado** un evento o una actividad en estado borrador, **cuando** un visitante abre la sección, **entonces** ese elemento no se le muestra y el resto de la sección se ve coherente, sin huecos ni campos en blanco.
2. **Dado** que no hay ningún evento ni actividad publicados (o todos están en borrador), **cuando** un visitante abre el sitio público, **entonces** la sección se oculta por completo: no aparece vacía, y la navegación (barra o menú hamburguesa) y el mapa del sitio del pie no dejan enlaces rotos ni huérfanos (FR-013 de F3 y FR-027 de F10).
3. **Dado** un elemento en borrador, **cuando** el usuario del panel lo publica, **entonces** el elemento aparece en el sitio público.
4. **Dado** un elemento publicado, **cuando** el usuario del panel lo devuelve a borrador, **entonces** deja de ser visible para el visitante.
5. **Dado** un elemento publicado, **cuando** el usuario del panel edita sus datos y guarda, **entonces** los cambios se hacen visibles de inmediato para el visitante (patrón de F3: se publica al guardar).
6. **Dado** cualquier evento o actividad en borrador, **cuando** se consulta el sitio público por cualquier vía, **entonces** el sistema NUNCA expone ese contenido ni sus datos al público.
7. **Dado** el módulo en el panel, **cuando** el usuario revisa los eventos o las actividades, **entonces** ve el estado (borrador/publicado) de cada elemento para saber qué es público hoy.

---

### User Story 5 - Ver los eventos y las actividades en inglés (Priority: P2)

Como visitante que prefiere el inglés, quiero cambiar el idioma del sitio con el selector y ver los eventos y las actividades en inglés cuando exista traducción, para enterarme de lo que viene aunque mi idioma base no sea el español.

**Why this priority**: Las decisiones 6 y 8 hacen del bilingüismo un compromiso del sitio, y el encargo de `idea.md` §1 busca dar a conocer la iglesia "en nuestra provincia y por qué no internacionalmente". Va como P2: el valor central (lo que viene, en español) ya está en US1, y el inglés amplía el público sin bloquear el resto. Reutiliza íntegramente el comportamiento del selector de idioma de F3/F10.

**Independent Test**: Se prueba de forma independiente con eventos y actividades publicados en ambos idiomas y con elementos que solo tienen español: al elegir inglés, la interfaz y los contenidos traducidos se muestran en inglés, los contenidos sin inglés se muestran en español y no aparece ningún campo vacío; al volver a español, todo se muestra en español; la elección persiste al navegar y en una visita posterior desde el mismo dispositivo; y un dispositivo sin preferencia guardada se muestra en español. Entrega valor por sí sola: el público anglófono puede enterarse de lo que viene.

**Acceptance Scenarios**:

1. **Dado** la sección en español, **cuando** el visitante elige inglés en el selector de idioma, **entonces** la interfaz y los contenidos con versión en inglés se muestran en inglés, y la elección se mantiene al navegar por el resto del sitio y se recuerda entre visitas en el mismo dispositivo (FR-010 de F3).
2. **Dado** un evento o una actividad que solo tiene versión en español, **cuando** el visitante navega en inglés, **entonces** ese contenido se muestra en español (idioma base), nunca vacío ni traducido automáticamente (FR-009 de F3).
3. **Dado** un dispositivo sin ninguna preferencia de idioma guardada (primera visita), **cuando** el visitante abre la sección, **entonces** se muestra en español, que es el idioma base (Decisión 8).
4. **Dado** la sección en inglés, **cuando** el visitante vuelve a elegir español, **entonces** toda la página se muestra de nuevo en español sin pérdida de información.
5. **Dado** las fechas y los horarios de los elementos, **cuando** el visitante cambia de idioma, **entonces** cada fecha y cada horario se presentan de forma clara y comprensible en el idioma elegido (p. ej., "viernes 17 de octubre, 7:00 p. m."), no solo con números ambiguos, y sin cambiar el día ni la hora reales del elemento.

---

### User Story 6 - Recorrer la sección en cualquier dispositivo y de forma accesible (Priority: P2)

Como visitante de cualquier edad, dispositivo y nivel de experiencia con internet, quiero que la sección de eventos y actividades se adapte a mi pantalla y sea fácil de usar, para enterarme de lo que viene sin dificultades.

**Why this priority**: Es una restricción explícita del cliente (`idea.md` §7 y su introducción: personas de todas las edades, "adultos mucho más mayores que saben lo básico de internet"). Va como P2 porque es una capa transversal sobre US1 y US5 verificable por separado, pero no puede postergarse más allá del MVP sin incumplir el encargo. La sección nace con el sistema de diseño y movimiento de F10 ya aplicado (sin re-especificarlo aquí).

**Independent Test**: Se prueba de forma independiente sobre la sección publicada: se abre en teléfono, tableta y computadora y se comprueba que todos los elementos, fechas, lugares y enlaces están disponibles sin desplazamiento horizontal; se recorre con teclado y con lector de pantalla; se comprueba legibilidad y contraste, también con el modo oscuro de F10 activo y con texto ampliado. Entrega valor por sí sola: cualquier persona puede enterarse de lo que viene desde cualquier dispositivo.

**Acceptance Scenarios**:

1. **Dado** un teléfono con pantalla estrecha, **cuando** el visitante abre la sección, **entonces** ve todos los elementos sin desplazamiento horizontal y puede activar cada enlace con comodidad (elementos de toque suficientemente grandes).
2. **Dado** una persona con poca experiencia en internet, **cuando** recorre la sección, **entonces** encuentra lo que viene con títulos claros y una navegación simple (barra visible en escritorio, menú hamburguesa claramente etiquetado en móvil, según F10), sin gestos desconocidos.
3. **Dado** una persona que usa el teclado o un lector de pantalla, **cuando** navega la sección, **entonces** puede recorrer y activar todos los contenidos y enlaces en un orden comprensible, con foco visible.
4. **Dado** una persona con baja visión, **cuando** usa la sección (en tema claro u oscuro), **entonces** el texto se lee con contraste suficiente y puede ampliarse hasta el 200 % sin que se pierda información.
5. **Dado** cualquier dispositivo, **cuando** el visitante abre la sección, **entonces** el contenido y las funciones disponibles son los mismos que en los demás dispositivos.

---

### Edge Cases

- **Contenido en borrador**: nunca debe ser visible para el visitante (US4, FR-013), ni en la página ni en ninguna respuesta del sistema hacia el público; un borrador no debe "filtrarse" por ninguna vía.
- **Sección sin ningún elemento publicado** (o con todo en borrador): la sección **se oculta por completo** (patrón de F3, FR-013 de F3); la navegación y el mapa del sitio del pie se actualizan sin dejar enlaces rotos ni huérfanos (FR-027 de F10).
- **Elemento publicado que se edita**: los cambios guardados se hacen **visibles de inmediato** para el visitante (patrón de F3: se publica al guardar); no hay paso intermedio de borrador para lo que ya está publicado.
- **Evento periódico cuyas fechas cargadas ya pasaron todas** (p. ej., el equipo aún no carga la próxima vigilia): el elemento no aparece como "lo que viene" sin fecha futura; se retira de la vista pública hasta que se cargue una próxima fecha, sin borrarlo del panel (Q1 y Q4; el panel señala al equipo que falta cargar la próxima fecha).
- **Elemento del día en curso** (hoy): se muestra entre "lo que viene" durante ese día y se retira automáticamente de la vista pública al terminar su fecha/hora (Q3 y Q4).
- **Elemento con fecha pasada al crearlo o editarlo**: no se rechaza por sí solo el registro de una fecha pasada (el equipo puede cargar lo que ya ocurrió y el panel lo muestra como pasado, Q4), pero el elemento no aparece en la vista pública de "lo que viene" (Q3).
- **Hora de fin anterior a la de inicio** en una fecha (cada fecha corresponde a un día concreto con hora de inicio y hora de fin opcionales, Q2): se rechaza con indicación de cómo corregirlo y no se guarda nada (FR-015).
- **Dos elementos en la misma fecha y hora**: ambos se muestran, cada uno en su bloque de concepto, en un orden estable y reproducible (default revisable).
- **Evento y actividad con datos idénticos**: son conceptos distintos (Decisión 3) y ambos se conservan; no se detectan como duplicados entre sí. Un duplicado exacto **dentro** del mismo módulo (mismo título, misma fecha, mismo lugar) se trata como duplicado y se rechaza o se señala (default revisable, coherente con F2 y F3).
- **Contenido sin versión en inglés navegando en inglés**: se muestra la versión en español (FR-008); nunca un campo vacío, un texto en otro idioma sin aviso ni una traducción automática (no existe traducción automática).
- **Contenido en inglés sin versión en español**: no debe poder guardarse; el español es el idioma base y obligatorio (Decisión 8, FR-015), y el sistema lo impide indicando qué falta.
- **Título, lugar, detalle o categoría con caracteres especiales del español** (tildes, ñ, ¿ ¡) **y con emojis**: se conservan íntegros en el panel y en el sitio público, sin corromperse ni perderse.
- **Detalle más largo que el límite de extensión** (2.000 caracteres por defecto, default revisable): el sistema lo indica al excederse y no guarda hasta corregirlo; los textos largos deben acomodarse sin romper la presentación en ningún dispositivo.
- **Imagen opcional con archivo inválido o carga fallida**: se indica el problema con el archivo y el elemento se conserva sin imagen; la sección pública no se rompe ni muestra huecos (FR-020). Toda imagen cargada lleva texto alternativo accesible.
- **Enlace de mapa/lugar mal formado**: se rechaza con indicación de cómo corregirlo y no se guarda el dato (FR-015); la vigencia del destino no se verifica (ver Out of Scope).
- **Dos personas editando el mismo evento o actividad al mismo tiempo**: se conserva la última versión guardada completa (default revisable, coherente con F3); no debe quedar una mezcla incoherente de datos sin aviso.
- **Eliminación de un elemento publicado**: desaparece de lo público de inmediato, la eliminación queda registrada en la auditoría y no hay papelera ni recuperación (default revisable, coherente con "sin historial de versiones" de F3).
- **Muchos elementos próximos**: se muestran todos los próximos en orden cronológico sin límite artificial (default revisable); si en el futuro fueran demasiados, la paginación o el "ver más" serían un ajuste menor acordado con el humano.
- **Operación sin permiso sobre «eventos», «actividades» o «categorías»** (lectura de la gestión, creación, edición, publicación o eliminación, incluida la administración del catálogo de categorías): denegada con mensaje claro, igual que en F2 (FR-012); tener un permiso no da acceso a los otros.
- **Cualquier operación del módulo (crear, editar, publicar/retirar, eliminar) que no se completa**: la operación no se aplica y el intento queda registrado en la auditoría con su resultado; nunca debe aplicarse una operación sin registro, ni registrarse una operación que no se aplicó (FR-017, coherente con F2).

### Errores esperados

Todos los mensajes deben ser comprensibles para personas no técnicas y nunca deben exponer información interna del sistema (coherente con F1, F2 y F3).

| Situación | Qué debe ver la persona |
|---|---|
| Campo obligatorio vacío al crear o editar un evento o una actividad (p. ej., título, fecha, lugar o detalle en español) | Indicación de qué campo corregir, sin guardar nada |
| Fecha mal formada o inválida (p. ej., "32/13/2026") | Indicación de cómo corregir la fecha, sin guardar nada |
| Hora de fin anterior a la de inicio en una fecha (Q2) | Mensaje que indica el horario correcto, sin guardar nada |
| Intento de guardar un evento o una actividad sin versión en español (idioma base obligatorio) | Mensaje que indica que falta el contenido en español |
| Enlace de mapa/lugar mal formado | Indicación de cómo corregir el enlace, sin guardar nada |
| Título, lugar o detalle que excede su límite de extensión | Mensaje con el límite vigente y lo que falta recortar |
| Imagen con archivo inválido o carga fallida | Mensaje que indica el problema con el archivo; el elemento se conserva sin imagen y la sección no se rompe |
| Operación del módulo sin el permiso correspondiente («eventos», «actividades» o «categorías») | Mensaje claro de acceso denegado, sin detalles internos |
| Intento de eliminar un elemento | Confirmación explícita antes de eliminar, explicando que la operación no se puede deshacer |
| Sección sin ningún elemento publicado | La sección se oculta; en ningún caso se muestra vacía ni con campos en blanco |
| Evento periódico sin ninguna fecha futura cargada | No aparece en "lo que viene"; el panel indica que falta cargar la próxima fecha (default revisable) |
| Error inesperado del sistema al crear, editar, publicar o eliminar | Mensaje genérico sin información interna, sin perder lo que la persona ya haya escrito sin aviso |

## Requirements *(mandatory)*

### Functional Requirements

**Descubrimiento público (visitante, sin autenticación)**

- **FR-001**: El sitio DEBE mostrar una sección pública accesible sin autenticación que presente **lo que viene**: los próximos **eventos** (encuentros periódicos de la iglesia) y las próximas **actividades** (acciones puntuales), cada uno con título, fecha(s), lugar y detalle. Eventos y actividades DEBEN presentarse como **conceptos distintos y claramente diferenciados**, modelados por separado (Decisión 3), nunca mezclados en un solo listado sin identificación.
- **FR-002**: La vista pública DEBE mostrar únicamente lo que está por venir: los elementos con fecha **desde hoy en adelante**, en orden cronológico (Q3). Un elemento cuya fecha ya pasó NO DEBE aparecer en el sitio público; su recuento y su memoria corresponden a F8 (noticias y galería), fuera de F4.
- **FR-003**: Los elementos DEBEN presentarse en **orden cronológico por la próxima fecha**, de lo más próximo a lo más lejano (Q3), sin exigir al visitante ninguna acción extra (filtros, búsqueda ni paginación) para ver lo inmediato.
- **FR-004**: Un **evento** DEBE representar un encuentro periódico mediante una **lista explícita de fechas concretas** cargadas por el equipo (una o varias próximas ocurrencias), más un **texto opcional de frecuencia** (p. ej., «todos los viernes») (Q1). El sistema NO DEBE calcular fechas de recurrencia automáticamente. La vista pública DEBE mostrar la **próxima fecha** del evento para que el visitante identifique cuándo es la próxima vez que ocurre y, de existir, el texto de frecuencia.
- **FR-005**: Una **actividad** DEBE representar una acción puntual con **una fecha concreta** (un día), que DEBE poder incluir **hora de inicio y hora de fin opcionales** (rango dentro del mismo día); una actividad sin hora también es válida (Q2). El mismo criterio de horario DEBE aplicar a cada fecha concreta de un evento (Q2). Una actividad NO DEBE exigir periodicidad (Decisión 3).
- **FR-006**: Cuando la fecha de un evento o de una actividad publicado pasa, el sistema DEBE **retirarlo automáticamente de la vista pública** (Q4), sin exigir intervención del equipo: la vista pública nunca DEBE mostrar como "próximo" algo que ya ocurrió. El elemento DEBE **conservar su estado publicado en el panel** —donde el equipo lo ve como pasado— y NO DEBE borrarse ni convertirse en borrador.

**Bilingüismo (Decisión 6 y Decisión 8)**

- **FR-007**: Todo el contenido de eventos y actividades DEBE poder ingresarse en español (idioma base, obligatorio) y opcionalmente en inglés, y la interfaz del sitio público DEBE estar siempre disponible en ambos idiomas (Decisión 6 y Decisión 8).
- **FR-008**: Cuando un contenido no tiene versión en inglés y el visitante navega en inglés, el sistema DEBE mostrar la versión en español de ese contenido (idioma base); NUNCA DEBE mostrar un campo vacío ni una traducción automática de un contenido que no fue traducido (patrón de F3, FR-009 de F3).
- **FR-009**: El idioma se gestiona con el selector ya existente (FR-010 de F3 y F10): el cambio se aplica de inmediato, se mantiene al navegar por el sitio y se recuerda entre visitas en el mismo dispositivo; un dispositivo sin preferencia guardada se muestra en español. Las fechas y los horarios DEBEN presentarse de forma clara y comprensible en el idioma elegido (fecha escrita, no solo numérica), sin alterar el día ni la hora reales.

**Administración desde el panel (patrón de F2 y F3)**

- **FR-010**: El panel DEBE permitir **crear, editar y eliminar eventos** (encuentros periódicos) con su versión en español y su versión opcional en inglés (FR-007), y DEBE permitir mantener cada evento en estado borrador o publicado (FR-013).
- **FR-011**: El panel DEBE permitir **crear, editar y eliminar actividades** (acciones puntuales) con su versión en español y su versión opcional en inglés (FR-007), y DEBE permitir mantener cada actividad en estado borrador o publicado (FR-013).
- **FR-012**: La gestión de eventos DEBE verificar el permiso **«eventos»**, la gestión de actividades DEBE verificar el permiso **«actividades»** y la administración del **catálogo de categorías** (crear, editar y eliminar categorías) DEBE verificar el permiso **«categorías»**. Los permisos «eventos» y «actividades» están **ya reservados en el catálogo de F2 (FR-015 de F2)**; el permiso **«categorías»** es **nuevo y se añade al catálogo de F2** por decisión explícita del humano del 2026-10-11 (cambio de alcance autorizado: el catálogo pasa de 9 a 10 permisos). La verificación sigue el patrón de autorización de F2 según el rol de la cuenta; cada permiso autoriza la gestión completa de su módulo y **los tres son independientes entre sí**: tener uno no da acceso a los otros (Decisión 5). Toda operación no autorizada DEBE denegarse con un mensaje claro.
- **FR-013**: Cada evento y cada actividad DEBE poder mantenerse en estado **borrador** o **publicado** y DEBE poder publicarse o retirarse **por separado** (Decisión 4, patrón de F3). Solo lo publicado DEBE ser visible para el visitante; el sistema NUNCA DEBE exponer al público ningún contenido en borrador ni sus datos, por ninguna vía. Cuando no haya ningún evento ni actividad publicados, la sección DEBE ocultarse por completo (navegación y mapa del sitio del pie incluidos, sin enlaces rotos): el sitio NUNCA DEBE mostrar secciones vacías ni campos en blanco.
- **FR-014**: Cuando se guarda un cambio sobre un elemento **ya publicado**, los cambios DEBEN hacerse visibles de inmediato para el visitante (patrón de F3: se publica al guardar); no hay paso intermedio de borrador para lo que ya está publicado.
- **FR-015**: El sistema DEBE validar toda entrada de ambos módulos en el panel: campos obligatorios completos (título, fecha(s), lugar y detalle en español), fechas y horas con formato válido y coherentes entre sí (cada fecha corresponde a un día concreto, con hora de inicio y hora de fin opcionales —Q2—; la hora de fin NUNCA DEBE ser anterior a la de inicio), enlaces de mapa/lugar con formato válido, versión en español siempre presente y versión en inglés opcional (Decisión 8), y textos dentro de sus límites de extensión (FR-018 y FR-019). Ante datos inválidos o incompletos DEBE indicar qué corregir y NUNCA DEBE guardar cambios parciales. Todo contenido ingresado DEBE tratarse como datos y NUNCA DEBE ejecutarse como código ni insertar elementos activos (constitución §IV).
- **FR-016**: La interfaz del panel para estos módulos se entrega en español, como el resto del panel definido en F2 (Decisión 6 aplicada al sitio público).
- **FR-017**: Cada operación de estos módulos —crear, editar, cambiar el estado de publicación y eliminar un evento o una actividad— DEBE quedar registrada en el **registro de auditoría de F2** con quién la hizo, qué hizo, sobre qué elemento y cuándo —incluidos los intentos que no se completan— (patrón de F3, FR-017 de F3, ampliado a estas operaciones como en F2); el registro DEBE conservar el mismo carácter de **solo lectura** de F2 (FR-025 de F2): no se edita ni se borra.

**Contenido de cada concepto (Decisión 3)**

- **FR-018**: Cada **evento** DEBE poder mantener: título, su **lista de fechas concretas** con el horario de cada una (hora de inicio y hora de fin opcionales) y su **texto opcional de frecuencia** (Q1 y Q2), lugar (texto) con enlace de mapa o ubicación **opcional**, detalle/descripción, imagen **opcional**, organizador/encargado **opcional** y **categoría opcional** elegida del catálogo administrable del panel con el permiso **«categorías»** (Q6 y Q7); sus textos DEBEN ser traducibles de forma opcional al inglés.
- **FR-019**: Cada **actividad** DEBE poder mantener: título, su **fecha concreta** con su horario (hora de inicio y hora de fin opcionales, Q2), lugar (texto) con enlace de mapa o ubicación **opcional**, detalle/descripción, imagen **opcional**, organizador/encargado **opcional** y **categoría opcional** elegida del catálogo administrable del panel con el permiso **«categorías»** (Q6 y Q7); sus textos DEBEN ser traducibles de forma opcional al inglés. Una actividad NO DEBE exigir periodicidad: su modelo es la fecha de una acción puntual (Decisión 3).
- **FR-020**: La imagen de un evento o de una actividad DEBE ser opcional: un elemento sin imagen DEBE presentarse completo y coherente. Toda imagen cargada DEBE llevar texto alternativo accesible y DEBE respetar las reglas de identidad vigentes (no deformar el logotipo ni recolorearlo, FR-002 de F3 y FR-024 de F10).

**Relación con el sitio público (F3 y F10)**

- **FR-021**: La sección DEBE integrarse en el sitio público **naciendo con el sistema de diseño y movimiento de F10** (mismos componentes, tipografía, color con modo claro/oscuro, espaciado y movimiento respetando `prefers-reduced-motion`) y con la navegación de F10 (barra fija visible en escritorio, menú hamburguesa accesible en móvil, mapa del sitio del pie), sin rediseñar ni re-especificar ese sistema. La sección DEBE vivir bajo **una única entrada de navegación del sitio público —«Eventos y actividades»— con dos bloques claramente separados: «Próximos eventos» y «Próximas actividades»** (Q5), y **NUNCA DEBE tocar ni alterar la portada de F3** (ver Out of Scope).
- **FR-022**: Cuando la sección no tiene contenido publicado, DEBE ocultarse de forma coherente en toda la navegación (barra, menú hamburguesa y mapa del sitio del pie), sin dejar enlaces rotos ni huérfanos (FR-013 de F3 y FR-027 de F10).

**Responsividad y accesibilidad (restricción del cliente, `idea.md` §7)**

- **FR-023**: La sección DEBE adaptarse a cualquier dispositivo (teléfono, tableta y computadora) sin pérdida de contenido ni de funciones, y DEBE ser navegable por personas de todas las edades y niveles de experiencia: títulos claros, navegación simple, texto legible y elementos interactivos de tamaño suficiente.
- **FR-024**: La sección DEBE ser accesible: navegación completa por teclado, compatibilidad con lectores de pantalla, contraste suficiente (en tema claro y oscuro), posibilidad de ampliar el texto hasta el 200 % sin perder información y texto alternativo en las imágenes. El nivel de referencia es WCAG 2.1 nivel AA (default revisable, ver Assumptions). Las fechas y los horarios DEBEN presentarse en un formato claro y comprensible para personas con poca experiencia en internet (FR-009).

### Key Entities *(include if feature involves data)*

- **Evento** *(encuentro periódico, Decisión 3)*: uno o varios. Cada evento tiene título, una **lista explícita de fechas concretas** (una o varias próximas ocurrencias) con el horario de cada fecha (hora de inicio y hora de fin opcionales) y un **texto opcional de frecuencia** (p. ej., «todos los viernes») (Q1 y Q2); lugar (texto) con enlace de mapa opcional, detalle/descripción, imagen opcional, organizador/encargado opcional y categoría opcional del catálogo administrable con el permiso «categorías» (Q6 y Q7), con versión en español (obligatoria) e inglés (opcional) de sus textos. El sistema no calcula recurrencias: la periodicidad se comunica con las fechas cargadas y el texto de frecuencia, y la vista pública muestra siempre la próxima fecha (Q1).
- **Actividad** *(acción puntual, Decisión 3)*: una o varias. Cada actividad tiene título, **una fecha concreta** (un día) con su horario (hora de inicio y hora de fin opcionales) (Q2), lugar (texto) con enlace de mapa opcional, detalle/descripción, imagen opcional, organizador/encargado opcional y categoría opcional del catálogo administrable con el permiso «categorías» (Q6 y Q7), con versión en español (obligatoria) e inglés (opcional) de sus textos. No tiene periodicidad: describe una acción puntual (repartir comida, predicar en tal lugar).
- **Estado de publicación** (borrador/publicado, por elemento): cada evento y cada actividad se publica o retira por separado (Decisión 4) y decide si el visitante lo ve o no; la sección sin elementos publicados se oculta por completo (FR-022). Un elemento publicado cuya fecha pasa **se retira automáticamente de la vista pública y conserva su estado publicado en el panel**, donde el equipo lo ve como pasado; no se borra ni se convierte en borrador (Q4).
- **Categoría** *(opcional, Q6)*: etiqueta de tipo de evento o actividad (p. ej., acción social, predicación, adoración, entrevista, noche de oración, otro) que ayuda al visitante a reconocer de qué se trata. Se elige de un **catálogo administrable desde el panel con el permiso «categorías»** (Q6 y Q7) —permiso nuevo de F4, distinto e independiente de «eventos» y «actividades»— (catálogo inicial: acción social, predicación, adoración, entrevista, noche de oración, otro); un elemento sin categoría es válido.

Relaciones: los eventos y las actividades son **modelos distintos que comparten presentación pública** (Decisión 3) y se gestionan en **módulos separados con permisos separados** («eventos» y «actividades»), y el catálogo de categorías se administra con el permiso **«categorías»** (FR-012); los tres permisos son independientes entre sí. Cada elemento lleva su estado de publicación y sus versiones en español (obligatoria) e inglés (opcional con fallback al español, FR-008); lo gestiona una cuenta del panel con el permiso del módulo correspondiente (F2) y cada operación sobre él queda registrada en la auditoría de F2 (FR-017). El visitante solo ve lo publicado con fecha desde hoy en adelante (FR-002, Q3 y Q4); lo pasado queda en el panel y su recuento público corresponde a F8.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de los visitantes puede ver lo que viene (próximos eventos y actividades con título, fecha, lugar y detalle) sin autenticación, y lo encuentra en menos de 30 segundos desde que abre la sección.
- **SC-002**: 0 contenidos en borrador visibles para el público en las pruebas de aceptación (incluida cualquier respuesta del sistema hacia el visitante).
- **SC-003**: Todo cambio guardado sobre un elemento publicado es visible en el sitio público desde la primera carga posterior al guardado (se publica al guardar), sin pasos adicionales.
- **SC-004**: Un miembro del equipo crea y publica un evento o una actividad nueva en menos de 2 minutos desde el panel.
- **SC-005**: El 100% de las operaciones de los módulos verifican el permiso correspondiente («eventos», «actividades» o «categorías»): 0 operaciones completadas sin permiso en las pruebas de aceptación —incluido el cruce entre los tres permisos— y 100% denegadas con mensaje claro.
- **SC-006**: Con el idioma en inglés, el 100% de los textos de la interfaz está traducido y 0 campos aparecen vacíos por falta de traducción: cada contenido sin versión en inglés se muestra íntegramente en español.
- **SC-007**: La sección es usable en teléfono, tableta y computadora: 0 pérdida de elementos o funciones entre dispositivos y 0 desplazamiento horizontal en el ancho de pantalla más estrecho probado (320 px).
- **SC-008**: La sección supera una revisión de accesibilidad con 0 hallazgos críticos (referencia WCAG 2.1 nivel AA, en tema claro y oscuro) y es recorrerla y operarla por completo con teclado y con lector de pantalla.
- **SC-009**: 0 elementos cuya fecha ya pasó aparecen entre "lo que viene" en las pruebas de aceptación: el retiro automático de la vista pública (Q3 y Q4) se observa en el 100% de los casos probados, y un evento sin próxima fecha nunca aparece vacío ni sin fecha.
- **SC-010**: En pruebas con personas de distintas edades, incluidas con poca experiencia en internet, al menos el 90% identifica qué actividades y eventos vienen, cuándo son y dónde, sin ayuda.
- **SC-011**: Una persona del equipo sin el permiso del módulo correspondiente no logra ver ni modificar ese módulo en el 100% de los intentos de las pruebas de aceptación.
- **SC-012**: 0 secciones vacías ni campos en blanco: la sección sin contenido publicado se oculta por completo y la navegación (barra, menú hamburguesa y mapa del sitio) queda sin enlaces rotos ni huérfanos.
- **SC-013**: El 100% de las operaciones de los módulos de eventos y actividades de las pruebas de aceptación (crear, editar, publicar/retirar y eliminar) queda registrado en la auditoría de F2 (quién, qué, sobre qué y cuándo), y 0 registros de auditoría pueden editarse ni borrarse.

## Out of Scope

Queda explícitamente **fuera del alcance** de F4:

- **Inscripción, registro o inscripción por formulario a eventos o actividades**: `idea.md` §6 deja fuera "que los usuarios comunes se puedan registrar para llevar cursos o charlas"; la comunicación es por los canales publicados (organizador/encargado, WhatsApp y redes de F3). La inscripción es un candidato de fase 3 (patrón de la Decisión 1 para grupos).
- **Recordatorios, avisos, suscripciones o notificaciones** (por correo, WhatsApp o el navegador) sobre eventos próximos.
- **Exportación o suscripción de calendario** (p. ej., "agregar a mi calendario", archivos `.ics`, calendarios ICS públicos).
- **Noticias, galería y archivo visual de lo que ya pasó** (F8): F4 anuncia lo que viene; el recuento de eventos recientes con imágenes y videos pertenece a F8 (Q3: lo pasado no se muestra en el sitio público dentro de F4).
- **Cálculo automático de recurrencia** (próximas fechas generadas por frecuencia, excepciones, feriados): la periodicidad se comunica con fechas concretas cargadas por el equipo y un texto de frecuencia (Q1).
- **Bloque de resumen «Lo que viene» en la portada de F3**: fuera de alcance por la decisión Q5 (la sección vive bajo su única entrada de navegación «Eventos y actividades» y la portada no se toca); solo entraría como ampliación acotada si el cliente lo pide explícitamente.
- **Creación de banners o imágenes para redes sociales** (`idea.md` §6) y cualquier generación de material gráfico.
- **Publicación automática en redes sociales ni envío de mensajes por WhatsApp** (`idea.md` §8): el sitio solo presenta y enlaza.
- **Traducción automática de contenidos**: el inglés se ingresa manualmente; si un contenido no tiene inglés, se muestra en español (FR-008).
- **Historial de versiones del contenido, papelera ni recuperación de elementos eliminados, y programación de publicación con fecha futura** en el MVP (default revisable, coherente con F3).
- **Verificación de la vigencia de enlaces de mapa o de lugares**: el sistema solo valida el formato; no comprueba que el destino exista ni que el lugar siga vigente.
- **Grupos de conexión (F5), ministerios (F6) y donaciones (F7)**: conceptos y módulos distintos; F4 no los muestra ni los administra. Los "grupos de conexión" con horario semanal que crea una persona (talleres, `idea.md` §5) NO son eventos de F4: son el catálogo de F5.
- **Gestión de usuarios, roles y permisos** (F2): F4 activa los permisos «eventos» y «actividades» ya reservados en el catálogo de F2 (FR-015 de F2) y **añade el permiso nuevo «categorías»** a ese catálogo (9→10 permisos) por decisión explícita del humano del 2026-10-11: es el único cambio autorizado al modelo de permisos y queda registrado como cambio de alcance (Q7). La administración del catálogo de categorías se hace con el permiso «categorías» (no con «eventos» ni «actividades»). Queda fuera cualquier otro cambio al modelo de permisos: **no se añaden más permisos de módulos de F5–F9** ni se altera la gestión de usuarios y roles (esa gestión sigue siendo F2).
- **Rediseño o sistema de diseño** (F10 y F10B): F4 **aplica** el sistema de diseño y movimiento de F10 ya entregado (sitio público) y nace en el panel con los patrones actuales de F2/F3; el rediseño visual del panel es F10B y no forma parte de F4. Tampoco se re-especifica la navegación, el selector de idioma ni el modo oscuro (F10).
- **Bilingüismo del panel de administración**: el panel se entrega en español (como F2 y F3); la Decisión 6 aplica al sitio público.
- **Despliegue a producción** y cambios de infraestructura: sigue pendiente a nivel de proyecto.

## Assumptions

Los puntos que cambiaban el alcance o la experiencia fueron resueltos explícitamente en las aclaraciones del 2026-10-11 (ver *Aclaraciones resueltas (2026-10-11)*) y aquí aparecen ya como **decisiones firmes**. Lo demás son defaults razonables tomados donde la descripción no fija una regla; **son provisionales y revisables** por el humano.

**Decisiones firmes (resueltas el 2026-10-11)**

- **Alcance temporal de lo público (Q3)**: la vista pública muestra **solo lo próximo** —lo que viene desde hoy en adelante—, en orden cronológico ascendente por la próxima fecha; lo pasado no se muestra en el sitio público (su recuento corresponde a F8).
- **Periodicidad de los eventos (Q1)**: **lista explícita de fechas concretas cargadas por el equipo**, más un **texto opcional de frecuencia** ("todos los viernes"); el sistema **no** calcula fechas de recurrencia automáticamente.
- **Horario (Q2)**: **fecha con hora de inicio y hora de fin opcionales** (rango dentro del mismo día), aplicable a cada fecha de un evento y a la fecha de una actividad; una actividad puede publicarse con solo hora de inicio o sin hora.
- **Destino de lo que pasa de fecha (Q4)**: un elemento publicado cuya fecha pasa **se retira automáticamente de la vista pública** y **conserva su estado publicado en el panel**, donde el equipo lo ve como pasado; no se borra ni se convierte en borrador.
- **Fechas pasadas al crear o editar (derivado de Q3 y Q4)**: se permiten (el equipo puede cargar lo que ya ocurrió y el panel lo muestra como pasado), pero no aparecen en la vista pública de "lo que viene".
- **Ubicación en el sitio (Q5)**: **una única entrada de navegación del sitio público —«Eventos y actividades»— con dos bloques separados** («Próximos eventos» y «Próximas actividades»), sin tocar la portada de F3; la composición concreta (lista, tarjetas, ficha de detalle) la define la fase de UX.
- **Categorías (Q6 y Q7)**: **categoría opcional con catálogo administrable desde el panel con el permiso «categorías»** (permiso nuevo de F4, añadido al catálogo de F2 por decisión del humano del 2026-10-11) (catálogo inicial: acción social, predicación, adoración, entrevista, noche de oración, otro); un elemento sin categoría es válido.

**Defaults revisables**

- **Evento periódico sin fecha futura**: no aparece en la vista pública hasta que se cargue una próxima fecha; el panel lo señala al equipo (default revisable).
- **Sin límite artificial de elementos ni filtros**: se muestran todos los próximos en orden cronológico; si en el futuro fueran demasiados, la paginación o el "ver más" serían un ajuste menor acordado con el humano (default revisable).
- **Lugar**: texto obligatorio (nombre/dirección del lugar) y **enlace de mapa o ubicación opcional** (URL con formato válido); no se incrustan mapas interactivos ni se verifica el destino (default revisable).
- **Imagen**: opcional por elemento; cuando se sube, lleva texto alternativo accesible. Un elemento sin imagen se presenta completo (FR-020).
- **Organizador/encargado**: campo de texto opcional con el nombre de la persona o grupo responsable; no hay directorio de personas ni fichas de encargados (eso pertenece a F5/F6). El contacto se resuelve con los canales publicados de F3 (default revisable).
- **Detalle/descripción**: texto plano multilínea con límite de extensión (**2.000 caracteres** por defecto, default revisable); admite los caracteres propios del español (tildes, ñ, ¿ ¡) y emojis sin transformación (coherente con F3).
- **Traducción manual**: no hay traducción automática ni sugerencias; el inglés opcional lo ingresa el equipo (Decisión 8).
- **Sin historial de versiones**: al editar se sobrescribe lo guardado; no hay versión anterior recuperable (default revisable, coherente con F3).
- **Eliminación de elementos**: definitiva, con confirmación explícita y registrada en la auditoría; no hay papelera ni recuperación (default revisable).
- **Ediciones simultáneas**: si dos personas editan el mismo elemento al mismo tiempo, se conserva la última versión guardada completa (default revisable, coherente con F3).
- **Duplicados**: un duplicado exacto dentro del mismo módulo (mismo título, misma fecha y mismo lugar) se rechaza o se señala como duplicado (default revisable, coherente con F2 y F3); un evento y una actividad con datos parecidos no se consideran duplicados (Decisión 3).
- **Elementos en la misma fecha y hora**: ambos se muestran, cada uno en su bloque de concepto, en un orden estable y reproducible (default revisable).
- **Validación de enlaces**: se valida el formato del enlace (que sea una dirección reconocible), no que el destino exista o siga activo.
- **Idioma del panel**: se entrega en español, como F2 y F3; el bilingüismo del roadmap aplica al sitio público (Decisión 6).
- **Nivel de accesibilidad**: se toma **WCAG 2.1 nivel AA** como referencia (default revisable); el requisito de fondo es el de `idea.md` §7: apto para personas de todas las edades y dispositivos.
- **Dependencias**: F2 terminada (panel, roles con permisos por módulo, **permisos «eventos» y «actividades» ya reservados** en FR-015 de F2 —más el **permiso nuevo «categorías»** que F4 añade al catálogo de F2 (9→10) por decisión del humano del 2026-10-11, Q7—, patrón de autorización y registro de auditoría), F3 terminada (sitio público, bilingüismo con fallback, estados borrador/publicado, ocultar secciones vacías) y F10 terminada (sistema de diseño y movimiento, navegación, selector de idioma, modo oscuro y mapa del sitio del pie).

## Aclaraciones resueltas (2026-10-11)

**Ninguna aclaración pendiente.** Las seis preguntas abiertas de la primera versión (Q1–Q6) se resolvieron con el humano el 2026-10-11 (quien decide: humano/cliente), en todos los casos por la **opción recomendada**. Además, el 2026-10-11 el humano tomó una **decisión adicional de cambio de alcance** (Q7): el catálogo de categorías se administra con un **permiso nuevo «categorías»**, ampliando el catálogo de permisos de F2. Queda aquí el registro de la decisión tomada en cada una.

### Q1 — Modelo de la periodicidad de un evento (FR-004, FR-018, Key Entities) — **Resuelta**

**Decisión (opción 1)**: **lista explícita de fechas concretas cargadas por el equipo** (una o varias próximas ocurrencias) **+ texto opcional de frecuencia** (p. ej., «todos los viernes»). **Sin cálculos automáticos de recurrencia**: el sistema no genera fechas. La vista pública muestra la próxima fecha cargada y, de existir, el texto de frecuencia. (Descartadas: patrón calculado, un evento por fecha y solo texto de frecuencia.)

### Q2 — Representación del horario (FR-005, FR-015, FR-018, FR-019) — **Resuelta**

**Decisión (opción 1)**: **fecha con hora de inicio y hora de fin opcionales** (rango dentro del mismo día), aplicable a cada fecha de un evento y a la fecha de una actividad; una actividad puede publicarse sin hora. La hora de fin nunca puede ser anterior a la de inicio (FR-015). (Descartadas: fecha y hora únicas; rango de fechas de varios días.)

### Q3 — Alcance temporal de la vista pública (FR-002, FR-003) — **Resuelta**

**Decisión (opción 1)**: **solo lo próximo** —desde hoy en adelante—, en orden cronológico de lo más próximo a lo más lejano. Lo pasado **no se muestra** en el sitio público: su recuento corresponde a F8 (noticias y galería). (Descartadas: bloque «Ya pasó» al final; todo agrupado por mes con lo pasado marcado.)

### Q4 — Destino automático de un elemento cuando pasa su fecha (FR-006) — **Resuelta**

**Decisión (opción 1)**: un elemento publicado cuya fecha pasa **se retira automáticamente de la vista pública** y **conserva su estado publicado en el panel**, donde el equipo lo ve como pasado; **no se borra ni se convierte en borrador**. El sitio muestra solo lo que viene sin intervención del equipo. (Descartadas: marca «realizado» visible para el visitante; retiro manual.)

### Q5 — Ubicación de la sección en el sitio público (FR-021) — **Resuelta**

**Decisión (opción 1)**: **una única entrada de navegación —«Eventos y actividades»— con dos bloques separados**: «Próximos eventos» y «Próximas actividades». Respeta la separación de conceptos (Decisión 3) y mantiene el menú corto (coherente con la previsión de ~8–10 enlaces de F10). **La portada de F3 no se toca** (el bloque de resumen «Lo que viene» queda fuera de alcance salvo petición explícita del cliente, ver Out of Scope). (Descartadas: dos entradas independientes; página «Lo que viene» con bloque en la portada.)

### Q6 — Categorías de eventos y actividades (FR-018, FR-019, Key Entities) — **Resuelta**

**Decisión (opción 1)**: **categoría opcional con catálogo administrable desde el panel** (catálogo inicial: acción social, predicación, adoración, entrevista, noche de oración, otro). Un elemento sin categoría es válido. (Descartadas: sin categoría en el MVP; etiquetas libres.)

### Q7 — Permiso para administrar el catálogo de categorías (cambio de alcance, FR-012, FR-018, FR-019) — **Decisión adicional del 2026-10-11**

**Decisión del humano (2026-10-11, cambio de alcance explícito)**: el **catálogo de categorías** de eventos y actividades se administra en el panel con un **permiso nuevo «categorías»**, y no con «eventos» ni «actividades». Este permiso **se añade al catálogo fijo de permisos que definió F2**, que pasa de **9 a 10 permisos**; F4 activa así **tres** permisos del catálogo —«eventos», «actividades» y «categorías»—, independientes entre sí (FR-012). Modifica la premisa previa de que F4 solo activaba «eventos» y «actividades» sin crear permisos nuevos; el humano tomó la decisión conociendo el impacto (actualizar pruebas existentes de F2 y una migración nueva) y la asume. Sigue fuera de alcance añadir más permisos de módulos de F5–F9 (ver Out of Scope).

### Sobre el resto del documento

Fuera de Q1–Q6 no hay ambigüedades abiertas: los estados borrador/publicado (Decisión 4), la separación de permisos por módulo —«eventos» y «actividades» ya reservados por F2 (Decisión 5, FR-015 de F2), más el permiso nuevo «categorías» añadido por el cambio de alcance del 2026-10-11 (Q7)—, el bilingüismo español base / inglés opcional con fallback (Decisiones 6 y 8), la auditoría de solo lectura (patrón de F2/F3), la responsividad y accesibilidad (`idea.md` §7) y la aplicación del sistema de diseño de F10 están definidos por decisiones o specs ya aprobadas y se recogen aquí sin reabrirlas.
