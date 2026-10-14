# Feature Specification: Donaciones (F7)

**Feature Branch**: `007-donaciones`

**Created**: 2026-10-14

**Status**: Draft — con **11 preguntas `[NECESITA ACLARACIÓN]` abiertas (Q1–Q11)**, marcadas en el punto exacto del texto y resumidas al final. Pendiente de la fase `clarify` y de la aprobación humana (puerta 1). Ninguna de las preguntas abiertas bloquea la lectura del documento: cada una señala qué cambia según la respuesta.

**Input**: User description: "F7 — Donaciones (roadmap §3, `docs/producto/roadmap.md`): El visitante ve cómo apoyar a la iglesia: cuentas IBAN y SINPE Móvil, y en qué se usan las donaciones; el equipo administra esa información desde el panel." Y de `idea.md` §1: "una sección de cómo pueden ayudarnos con donaciones, numeros de cuentas IBAN y SINPE Movil y en qué utilizamos esas donaciones", §1 ("donde hacer donaciones"), §2 (usuarios comunes: "Navegar por el sitio y ver la información de actividades, eventos, grupos de conexiones, ministerios y dónde o cómo donar"), §5 ("seccion de donaciones" dentro del MVP) y §7 (el sitio debe adaptarse a cualquier dispositivo y manejar permisos, roles y estados). Decisiones aplicables del roadmap: **Decisión 2** (2026-09-28: las donaciones entran al MVP, F7), **Decisión 4** (2026-09-28: contenido con estados borrador/publicado; solo lo publicado es visible al visitante), **Decisión 5** (2026-09-28: permisos por módulo agrupados en roles; el catálogo de roles es libre y cada permiso autoriza la gestión completa de su módulo), **Decisión 6** (2026-09-28: el sitio público es bilingüe, español e inglés, con selección de idioma), **Decisión 7** (2026-09-28: la información de donaciones —IBAN/SINPE y su uso— se administra desde el panel; F7 depende de F2 y F3), **Decisión 8** (2026-09-28: cada contenido puede ingresarse en ambos idiomas o solo en español; el español es el idioma base y el inglés es opcional por contenido; la interfaz siempre es bilingüe) y **Decisión 9** (2026-10-04: el modelo de cuentas y roles de F2 —cuentas que no se eliminan, un solo rol por cuenta, roles editables/eliminables solo si no están en uso— se usa tal cual y F7 no lo cambia).

> **Nota de alcance y terminología**: esta funcionalidad tiene **dos partes confirmadas** y **una tercera por confirmar**: (a) la **sección pública** donde el visitante ve **cómo apoyar a la iglesia** —las **cuentas bancarias (IBAN)** y los **números de SINPE Móvil** por los que puede donar, y **en qué se usan las donaciones**—, (b) el **panel**, donde el equipo con el permiso **«donaciones»** administra esa información, y (c) una posible **presencia en la portada** que aún no está decidida `[NECESITA ACLARACIÓN: Q7 — ¿F7 incluye un bloque en la portada de F3 (patrón de F4 y F6) o la donación vive solo en su sección?]`. Se reutiliza la terminología de F2 y F3: el **visitante** (el "usuario común" de `idea.md` §2: ve el sitio público, no tiene cuenta y no se registra) y el **usuario del panel** (el "usuario del sistema" y el "usuario administrador" de `idea.md` §2: tiene cuenta, inicia sesión y gestiona el contenido). **Qué es la información de donaciones en este producto**: los **datos públicos que la iglesia ofrece para que una persona pueda aportar económicamente por los canales habituales del país** —una o varias **cuentas bancarias (IBAN)** y uno o varios **números de SINPE Móvil** `[NECESITA ACLARACIÓN: Q1 — ¿es un listado administrable de varias cuentas y varios números, o una sola cuenta IBAN y un solo SINPE Móvil?]`— junto con el contenido que explica **en qué se utilizan esas donaciones** `[NECESITA ACLARACIÓN: Q5 — ¿texto libre editable o listado administrable de destinos/usos?]`. **F7 NO procesa donaciones**: no hay cobro ni pago en línea, montos, metas, comprobantes ni registro de quién donó (ver Out of Scope); el sitio **publica la información** y la persona dona por sus propios medios (transferencia bancaria o SINPE Móvil). **Una donación NO es un evento, una actividad, un grupo de conexión ni un ministerio**: son conceptos distintos y F7 no construye ninguna de esas funcionalidades. Tampoco se re-especifica lo que ya resuelven **F2** (inicio de sesión, cuentas, roles, permisos y el mecanismo de auditoría), **F3** (estructura del sitio público, bilingüismo con respaldo, ocultar secciones vacías) y **F10** (sistema de diseño y movimiento, navegación, selector de idioma, modo oscuro, mapa del sitio).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ver cómo apoyar a la iglesia (Priority: P1)

Como visitante (miembro de la iglesia o persona nueva que quiere apoyar), quiero ver en un solo lugar las cuentas bancarias (IBAN) y los números de SINPE Móvil por los que puedo donar y en qué se utilizan las donaciones, para aportar económicamente a la iglesia con confianza y sin tener que preguntar por WhatsApp ni en Facebook.

**Why this priority**: Es el corazón de F7 y de su fila del roadmap: "el visitante ve cómo apoyar a la iglesia: cuentas IBAN y SINPE Móvil, y en qué se usan las donaciones". `idea.md` §1 sitúa la donación entre los fines del sitio ("cómo pueden ayudarnos con donaciones… y en qué utilizamos esas donaciones") y `idea.md` §2 lo resume como "dónde o cómo donar". Sin esta historia no hay valor visible para el público y la gestión del panel (US2–US3) no tendría destinatario. Traduce además el criterio de éxito del roadmap §8 ("una puerta para acercarse a la congregación"): aquí la puerta es el apoyo económico con transparencia sobre el uso de los aportes.

**Independent Test**: Se prueba de forma independiente con información de donaciones publicada: un visitante sin sesión abre la sección y encuentra cada cuenta o número de donación con su tipo (IBAN o SINPE Móvil), la entidad y el número tal como el equipo lo registró, íntegro y exacto, y lee en qué se usan las donaciones; al cambiar y publicar un dato desde el panel y recargar, el visitante ve el dato actualizado; el listado respeta el orden definido por el equipo; y nada de lo mostrado se confunde con un evento, un grupo de conexión ni un ministerio. Entrega valor por sí sola: el público ya sabe cómo donar y en qué se invierten sus aportes.

**Acceptance Scenarios**:

1. **Dado** información de donaciones publicada, **cuando** un visitante abre la sección sin iniciar sesión, **entonces** ve cada método de donación con su tipo (cuenta IBAN o SINPE Móvil), la entidad y el número tal como se registró, y el contenido sobre **en qué se usan las donaciones** `[NECESITA ACLARACIÓN: Q5 — ¿texto libre editable o listado administrable de destinos/usos?]`.
2. **Dado** una cuenta bancaria (IBAN) publicada, **cuando** el visitante la consulta, **entonces** el número aparece **completo y exacto**, sin cortes ni caracteres perdidos, de modo que pueda copiarlo o transcribirlo sin error para hacer su transferencia (FR-003).
3. **Dado** un número de SINPE Móvil publicado, **cuando** el visitante lo consulta, **entonces** identifica que ese es el número al que puede enviar el aporte por SINPE Móvil desde su banco en el celular, con el número tal como se registró `[NECESITA ACLARACIÓN: Q4 — reglas de formato y presentación del SINPE Móvil]`.
4. **Dado** varios métodos de donación publicados (una o varias cuentas IBAN y/o uno o varios SINPE Móvil), **cuando** el visitante recorre la sección, **entonces** los ve en el **orden estable definido por el equipo**, sin tener que ordenar, filtrar ni buscar para verlos todos `[NECESITA ACLARACIÓN: Q1 — ¿listado administrable con orden manual o agrupado por tipo (cuentas IBAN y SINPE Móvil)?]`.
5. **Dado** la sección publicada, **cuando** una persona nueva la recorre, **entonces** identifica cómo apoyar a la iglesia (a qué cuenta o número donar) y **en qué se usan las donaciones**, sin buscar en WhatsApp, en Facebook ni preguntando a nadie.
6. **Dado** que **no hay ninguna información de donaciones publicada** (o todo está en borrador), **cuando** un visitante intenta ver la sección, **entonces** el sistema responde según lo que decida el cliente `[NECESITA ACLARACIÓN: Q9 — ¿ocultar la sección por completo (patrón F3/F6), mostrar un aviso sin redirigir (patrón F4) o redirigir a la portada (patrón F6)?]`; en ningún caso se muestra una sección vacía, con campos en blanco ni con datos de ejemplo.
7. **Dado** un método de donación publicado **sin** algún dato opcional (p. ej., sin nombre del titular o sin moneda, según Q2), **cuando** el visitante ve su ficha, **entonces** la ficha se presenta completa y coherente: **sin líneas vacías ni marcadores de "sin datos"**.
8. **Dado** la sección de donaciones, **cuando** el visitante la recorre, **entonces** no la confunde con un evento, un grupo de conexión ni un ministerio: solo presenta cómo donar y en qué se usan las donaciones (FR-001, SC-013).

---

### User Story 2 - Gestionar la información de donaciones desde el panel (Priority: P1)

Como usuario del panel con el permiso «donaciones», quiero crear, editar y eliminar las cuentas y números por los que se recibe el aporte —con su tipo (IBAN o SINPE Móvil), su entidad, su número y los demás datos acordados— y mantener el contenido que explica en qué se usan las donaciones, todo en español y, si quiero, también en inglés, para mantener al día esta información sin depender de un desarrollador.

**Why this priority**: Es la mitad administrativa de F7 y la Decisión 7 del roadmap ("la información de donaciones (IBAN/SINPE y su uso) se administra desde el panel"), además del criterio de éxito de `idea.md` §4 (contenido dinámico, no estático). Es construible y testeable por separado de la sección pública: reutiliza el patrón probado de F3, F4 y F6 y activa un permiso **ya existente** en F2 —**F7 no crea permisos nuevos**—, sin tocar el modelo de permisos. Es además crítica para la confianza: si un número de cuenta cambia (el banco reemplaza la cuenta, se abre un SINPE nuevo), el equipo debe poder corregirlo en minutos.

**Independent Test**: Se prueba de forma independiente con una cuenta con el permiso «donaciones» y otra sin él: la primera crea un método de donación con datos válidos y el sistema lo conserva; con un número mal formado o incompleto, el sistema indica qué corregir y no guarda nada a medias; edita, reordena y elimina un método (con confirmación) y la eliminación queda registrada en la auditoría; edita el contenido de "en qué se usan las donaciones" y el cambio se conserva; la segunda cuenta recibe una denegación clara en toda operación del módulo, incluidos los intentos forzados por la interfaz. Entrega valor por sí sola: el equipo ya administra su información de donaciones.

**Acceptance Scenarios**:

1. **Dado** un usuario del panel con el permiso «donaciones», **cuando** crea un método de donación con datos válidos —**obligatoriedad de cada campo según lo que decida el cliente** `[NECESITA ACLARACIÓN: Q2 — ¿qué campos son obligatorios: nombre del banco/entidad, titular, número, moneda?]`— y guarda, **entonces** el sistema confirma la operación y conserva el método de donación con todos sus datos.
2. **Dado** una cuenta sin el permiso «donaciones», **cuando** intenta crear, editar o eliminar un método de donación o modificar el contenido de uso de las donaciones, **entonces** el sistema impide la operación con un mensaje claro y no cambia nada.
3. **Dado** un formulario de método de donación con datos inválidos o incompletos (campo obligatorio vacío, número con formato no válido según Q3/Q4), **cuando** envío el formulario, **entonces** el sistema indica qué corregir y **no guarda ningún cambio parcial** (FR-014).
4. **Dado** un método de donación publicado, **cuando** edito sus datos (p. ej., el número de cuenta) y guardo, **entonces** el sistema conserva los cambios y se hacen visibles de inmediato para el visitante (patrón de F3: se publica al guardar).
5. **Dado** un método de donación existente, **cuando** lo elimino y confirmo la eliminación, **entonces** desaparece del panel y del sitio público, y la eliminación queda registrada en el registro de auditoría de F2 (quién, qué y cuándo).
6. **Dado** el formulario del módulo, **cuando** la persona ingresa la versión en inglés de un campo de texto relleno, **entonces** el sistema conserva esa versión junto a la versión en español.
7. **Dado** el formulario del módulo, **cuando** la persona deja vacío un campo en inglés (o un campo entero opcional), **entonces** puede guardar igualmente: el inglés es opcional por contenido (Decisión 8) y **un campo vacío no obliga a nada**.
8. **Dado** una cuenta con el permiso «donaciones» pero sin otros permisos de módulo, **cuando** intenta forzar el acceso a la gestión de otros módulos (p. ej., «eventos» o «ministerios»), **entonces** el sistema lo impide: los permisos son independientes entre sí (Decisión 5).
9. **Dado** el módulo con varios métodos de donación, **cuando** el usuario cambia su **valor de orden** y guarda (patrón de F3/F4/F6), **entonces** el nuevo orden se conserva y es exactamente el que ve el visitante (FR-004) `[NECESITA ACLARACIÓN: Q1 — ¿el orden es un único listado o por grupos de tipo?]`.
10. **Dado** el contenido de "en qué se usan las donaciones" (y el texto introductorio, si existe — Q11), **cuando** el usuario lo edita y guarda, **entonces** el sistema conserva el cambio y se hace visible de inmediato si está publicado.

---

### User Story 3 - Controlar la publicación de la información de donaciones (Priority: P2)

Como usuario del panel con el permiso «donaciones», quiero dejar la información de donaciones en borrador o publicarla por separado, para que el público solo vea datos revisados y nunca cambios a medias —algo especialmente sensible cuando se trata de números de cuenta.

**Why this priority**: La Decisión 4 (contenido borrador/publicado) es parte explícita del alcance y protege la imagen pública de la iglesia: un número de cuenta a medias o equivocado puede hacer que un aporte se pierda. Va como P2 porque US2 ya entrega un MVP usable (crear y conservar la información), mientras que esta historia añade el control de visibilidad. Reutiliza literalmente el patrón probado en F3, F4 y F6.

**Independent Test**: Se prueba de forma independiente con un método de donación en borrador y otro publicado: el borrador no aparece para el visitante y el resto de la sección se ve coherente; al publicarlo aparece; al devolverlo a borrador desaparece; una sección sin nada publicado responde según Q9; y ninguna respuesta del sistema hacia el público expone un borrador. Entrega valor por sí sola: el equipo decide qué se hace público y cuándo.

**Acceptance Scenarios**:

1. **Dado** un método de donación en estado borrador, **cuando** un visitante abre la sección, **entonces** ese método no se le muestra y el resto de la sección se ve coherente, sin huecos ni campos en blanco.
2. **Dado** que no hay nada publicado (o todo está en borrador), **cuando** un visitante intenta abrir la sección, **entonces** el sistema responde según la decisión del cliente `[NECESITA ACLARACIÓN: Q9 — ocultar por completo, aviso sin redirigir o redirección a la portada]`, sin mostrar la sección vacía ni dejar enlaces rotos ni huérfanos en la navegación (barra, menú hamburguesa y mapa del sitio del pie).
3. **Dado** un método de donación en borrador, **cuando** el usuario del panel lo publica, **entonces** aparece en el sitio público.
4. **Dado** un método de donación publicado, **cuando** el usuario del panel lo devuelve a borrador, **entonces** deja de ser visible para el visitante en toda superficie pública.
5. **Dado** cualquier información de donaciones en borrador, **cuando** se consulta el sitio público por cualquier vía, **entonces** el sistema **NUNCA** expone ese contenido ni sus datos al público.
6. **Dado** el módulo en el panel, **cuando** el usuario revisa la información de donaciones, **entonces** ve el estado (borrador/publicado) de cada elemento para saber qué es público hoy.
7. **Dado** la pregunta abierta sobre granularidad `[NECESITA ACLARACIÓN: Q10 — ¿la publicación es por cada cuenta/número (patrón F3/F4/F6) o la sección se publica como un todo?]`, **cuando** el humano la responda, **entonces** el control de publicación adopta esa granularidad sin cambiar el resto de la spec.

---

### User Story 4 - Ver la información de donaciones en inglés (Priority: P2)

Como visitante que prefiere el inglés, quiero cambiar el idioma del sitio con el selector y ver la información de donaciones en inglés cuando exista traducción, para saber cómo apoyar a la iglesia aunque mi idioma base no sea el español.

**Why this priority**: Las decisiones 6 y 8 hacen del bilingüismo un compromiso del sitio, y `idea.md` §1 busca dar a conocer la iglesia "en nuestra provicia y por qué no internacionalmente": quien dona puede estar fuera de Costa Rica. Va como P2: el valor central (la información, en español) ya está en US1 y el inglés amplía el público sin bloquear el resto. Reutiliza íntegramente el comportamiento del selector de idioma de F3/F10.

**Independent Test**: Se prueba de forma independiente con textos traducidos al inglés y con textos que solo tienen español: al elegir inglés, la interfaz y los contenidos traducidos se muestran en inglés, los contenidos sin inglés se muestran en español y no aparece ningún campo vacío; los números de cuenta y de SINPE Móvil son idénticos en ambos idiomas; al volver a español, todo se muestra en español; la elección persiste al navegar y entre visitas desde el mismo dispositivo; y un dispositivo sin preferencia guardada se muestra en español. Entrega valor por sí sola: el público anglófono sabe cómo donar.

**Acceptance Scenarios**:

1. **Dado** la sección en español, **cuando** el visitante elige inglés en el selector de idioma, **entonces** la interfaz y los contenidos con versión en inglés se muestran en inglés, y la elección se mantiene al navegar por el resto del sitio y se recuerda entre visitas en el mismo dispositivo (FR-009 de F3).
2. **Dado** un contenido que solo tiene versión en español (p. ej., el texto de "en qué se usan las donaciones" o una nota de un método), **cuando** el visitante navega en inglés, **entonces** ese contenido se muestra en español (idioma base), nunca vacío ni traducido automáticamente.
3. **Dado** un dispositivo sin ninguna preferencia de idioma guardada (primera visita), **cuando** el visitante abre la sección, **entonces** se muestra en español, que es el idioma base (Decisión 8).
4. **Dado** la sección en inglés, **cuando** el visitante vuelve a elegir español, **entonces** toda la página se muestra de nuevo en español sin pérdida de información.
5. **Dado** un número de cuenta (IBAN) o de SINPE Móvil publicado, **cuando** el visitante cambia de idioma, **entonces** el número **no cambia en absoluto**: los números y los identificadores de la cuenta no son contenido traducible (FR-007) y siguen siendo copiables o transcribibles con exactitud.

---

### User Story 5 - Recorrer la sección en cualquier dispositivo y de forma accesible (Priority: P2)

Como visitante de cualquier edad, dispositivo y nivel de experiencia con internet, quiero que la sección de donaciones se adapte a mi pantalla y sea fácil de usar, para poder leer y copiar los números de cuenta sin dificultades.

**Why this priority**: Es una restricción explícita del cliente (`idea.md` §7 y su introducción: personas de todas las edades, "adultos mucho más mayores que saben lo básico de internet"), y aquí adquiere peso especial: **transcribir un número de cuenta es una operación delicada** donde un error de lectura hace perder el aporte. Va como P2 porque es una capa transversal sobre US1 y US4 verificable por separado, pero no puede postergarse más allá del MVP sin incumplir el encargo. La sección nace con el sistema de diseño y movimiento de F10 ya aplicado (sin re-especificarlo aquí).

**Independent Test**: Se prueba de forma independiente sobre la sección publicada: se abre en teléfono, tableta y computadora y se comprueba que todos los números y textos están disponibles sin desplazamiento horizontal y legibles en su totalidad; se recorre con teclado y con lector de pantalla; se comprueba legibilidad y contraste, también con el modo oscuro de F10 activo y con texto ampliado al 200 %. Entrega valor por sí sola: cualquier persona puede leer y copiar los datos de donación desde cualquier dispositivo.

**Acceptance Scenarios**:

1. **Dado** un teléfono con pantalla estrecha, **cuando** el visitante abre la sección, **entonces** ve todos los métodos de donación sin desplazamiento horizontal y los números se leen completos (sin cortes que impidan copiarlos o transcribirlos), con elementos de toque suficientemente grandes.
2. **Dado** una persona con poca experiencia en internet, **cuando** recorre la sección, **entonces** encuentra la información con títulos claros y una navegación simple (barra visible en escritorio, menú hamburguesa claramente etiquetado en móvil, según F10), sin gestos desconocidos.
3. **Dado** una persona que usa el teclado o un lector de pantalla, **cuando** navega la sección, **entonces** puede recorrer y activar todos los contenidos y enlaces (incluida la copia del número, si la fase de UX la ofrece) en un orden comprensible, con foco visible.
4. **Dado** una persona con baja visión, **cuando** usa la sección (en tema claro u oscuro), **entonces** el texto —incluidos los números de cuenta— se lee con contraste suficiente y puede ampliarse hasta el 200 % sin que se pierda información.
5. **Dado** cualquier dispositivo, **cuando** el visitante abre la sección, **entonces** el contenido y las funciones disponibles son los mismos que en los demás dispositivos.

---

### Edge Cases

- **Información en borrador**: nunca debe ser visible para el visitante (US3, FR-012), ni en la página ni en ninguna respuesta del sistema hacia el público; un borrador no debe "filtrarse" por ninguna vía —aquí con más razón, porque un número de cuenta sin revisar puede llevar a un aporte perdido—.
- **Sección sin nada publicado** (o con todo en borrador): el comportamiento público queda **pendiente de decisión** `[NECESITA ACLARACIÓN: Q9 — ocultar la sección por completo (patrón F3 y F6), mostrar un aviso sin redirigir (patrón de F4 en `/eventos`) o redirigir a la portada (patrón de F6 en `/ministerios`)]`. Sea cual sea la decisión, la navegación (barra, menú hamburguesa) y el mapa del sitio del pie DEBEN quedar **sin enlaces rotos ni huérfanos** y el sitio NUNCA DEBE mostrar secciones vacías ni campos en blanco.
- **Dato publicado que se edita**: los cambios guardados se hacen **visibles de inmediato** para el visitante (patrón de F3: se publica al guardar); no hay paso intermedio de borrador para lo que ya está publicado. Es crítico para la corrección de un número de cuenta erróneo.
- **Número de cuenta (IBAN) mal formado o con estructura no válida**: se rechaza con indicación de cómo corregirlo y no se guarda nada `[NECESITA ACLARACIÓN: Q3 — reglas de validación y formato del IBAN: ¿estructura costarricense (CR + dígitos de control), IBAN internacional, cómo se normaliza y cómo se presenta?]`.
- **Número de SINPE Móvil mal formado**: se rechaza con indicación de cómo corregirlo y no se guarda nada `[NECESITA ACLARACIÓN: Q4 — ¿celular costarricense de 8 dígitos? ¿con o sin `+506`? ¿se aceptan otros identificadores SINPE además del número? ¿la presentación es tal como el equipo lo escribe?]`.
- **Número con espacios, guiones o caracteres escritos por el equipo** (p. ej., el IBAN agrupado en bloques): el sistema DEBE conservar el valor exacto del número y su presentación pública DEBE permitir copiarlo o transcribirlo sin error `[NECESITA ACLARACIÓN: Q3 — ¿se normaliza guardando solo los caracteres del número y se presenta agrupado, o se conserva tal como se escribió?]`.
- **Dos métodos de donación con el mismo número** (mismo IBAN o mismo SINPE, o una cuenta repetida): se trata como duplicado y se rechaza o se señala (default revisable, coherente con F2, F3 y F6).
- **Método de donación sin algún dato opcional** (p. ej., sin nombre del titular o sin moneda, según Q2): caso **válido**; la ficha pública no muestra línea vacía alguna y el método se publica igualmente.
- **Cambio de cuenta bancaria** (el banco reemplaza el IBAN o se abre un nuevo SINPE): es una edición normal sobre un elemento publicado; el dato viejo deja de mostrarse de inmediato y el sistema no conserva historial de números anteriores (default revisable, coherente con "sin historial de versiones" de F3). El visitante que copió un número antiguo y donó a él queda fuera del alcance del sistema (ver Out of Scope).
- **Texto "en qué se usan las donaciones" largo o con varias secciones**: se conserva íntegro como texto y la presentación debe acomodarlo sin romperse; el límite de extensión se indica al excederse (default revisable).
- **Caracteres especiales del español** (tildes, ñ, ¿ ¡) **y emojis** en los textos: se conservan íntegros en el panel y en el sitio público, sin corromperse. En los **números de cuenta no deben admitirse emojis ni letras fuera de la estructura válida** del formato (Q3, Q4).
- **Versión en inglés sin versión en español de un campo de texto relleno**: no debe poder guardarse; la regla del español base aplica **por campo** (Decisión 8) y el sistema lo impide indicando qué falta.
- **Idioma en inglés con números de cuenta**: los números, el tipo de método y la entidad financiera **no cambian** entre idiomas (FR-007); solo se traducen los textos.
- **Imagen (solo si el cliente decide que la hay)** `[NECESITA ACLARACIÓN: Q6 — ¿la sección y/o cada método de donación llevan imagen (encabezado, logos de banco) o no hay imágenes?]`: si existe, es opcional, un elemento sin imagen se presenta completo y coherente, toda imagen lleva texto alternativo accesible y un archivo inválido o una carga fallida se indica sin romper la sección.
- **Operación sin permiso «donaciones»** (lectura de la gestión, creación, edición, cambio de estado, reordenamiento o eliminación): denegada con mensaje claro, igual que en F2; tener otro permiso de módulo no da acceso a este.
- **Cualquier operación del módulo que no se completa**: la operación no se aplica y el intento queda registrado en la auditoría con su resultado; nunca debe aplicarse una operación sin registro, ni registrarse una operación que no se aplicó (FR-017, coherente con F2).
- **Dos personas editando el mismo elemento al mismo tiempo**: se conserva la última versión guardada completa (default revisable, coherente con F3).
- **Eliminación de un método de donación publicado**: desaparece de lo público de inmediato, la eliminación queda registrada en la auditoría y no hay papelera ni recuperación (default revisable, coherente con F3).
- **Muchos métodos de donación publicados**: se muestran todos en el orden definido, sin límite artificial (default revisable) ni paginación en el MVP.

### Errores esperados

Todos los mensajes deben ser comprensibles para personas no técnicas y nunca deben exponer información interna del sistema (coherente con F1, F2 y F3).

| Situación | Qué debe ver la persona |
|---|---|
| Campo obligatorio vacío al crear o editar un método de donación | Indicación de qué campo corregir, sin guardar nada (la obligatoriedad concreta depende de Q2) |
| Número de cuenta (IBAN) con estructura o formato no válido | Indicación de cómo corregir el número, sin guardar nada (reglas según Q3) |
| Número de SINPE Móvil no válido | Indicación de cómo corregir el número, sin guardar nada (reglas según Q4) |
| Intento de guardar una versión en inglés de un campo de texto sin su versión en español | Mensaje que indica que falta el contenido en español de ese campo |
| Texto que excede su límite de extensión | Mensaje con el límite vigente y lo que falta recortar |
| Imagen con archivo inválido o carga fallida (solo si aplica según Q6) | Mensaje que indica el problema con el archivo; el elemento se conserva sin imagen y la sección no se rompe |
| Operación del módulo sin el permiso «donaciones» | Mensaje claro de acceso denegado, sin detalles internos |
| Intento de eliminar un método de donación | Confirmación explícita antes de eliminar, explicando que la operación no se puede deshacer |
| Sección sin nada publicado | Comportamiento público por confirmar (Q9); en ningún caso sección vacía ni campos en blanco, y la navegación sin enlaces rotos |
| Error inesperado del sistema al crear, editar, publicar o eliminar | Mensaje genérico sin información interna, sin perder lo que la persona ya haya escrito sin aviso |

## Requirements *(mandatory)*

### Functional Requirements

**Definición del concepto (la información de donaciones no es cobro ni contenido de otros módulos)**

- **FR-001**: La información de donaciones DEBE representar **los datos públicos para que una persona pueda aportar económicamente a la iglesia por los canales habituales del país** —cuentas bancarias (IBAN) y números de SINPE Móvil `[NECESITA ACLARACIÓN: Q1 — ¿listado administrable de varias cuentas y varios números, o una sola cuenta IBAN y un solo SINPE Móvil fijos?]`— junto con el contenido que explica **en qué se utilizan esas donaciones** `[NECESITA ACLARACIÓN: Q5]`. F7 **NO DEBE procesar, cobrar ni registrar donaciones** (ver Out of Scope) y NO DEBE confundirse ni mezclarse con eventos, actividades (F4), grupos de conexión (F5) ni ministerios (F6): la sección pública y el módulo del panel DEBEN presentar y gestionar **únicamente información de donaciones**, y F7 NO DEBE construir ninguna parte de otras funcionalidades.

**Descubrimiento público (visitante, sin autenticación)**

- **FR-002**: El sitio DEBE mostrar una sección pública accesible sin autenticación con **cómo apoyar a la iglesia**: los **métodos de donación publicados** (cuentas IBAN y números de SINPE Móvil) y el contenido sobre **en qué se usan las donaciones** `[NECESITA ACLARACIÓN: Q5 — ¿texto libre editable o listado administrable de destinos/usos?; y Q11 — ¿lleva además un texto introductorio editable tipo «cómo donar»?]`. Un visitante DEBE poder identificar **a qué cuenta o número donar** y **en qué se invierten sus aportes** sin buscar en ningún otro sitio (SC-001, SC-012).
- **FR-003**: Cada método de donación publicado DEBE mostrar su **tipo** (cuenta IBAN o SINPE Móvil), su **entidad financiera** y su **número tal como se registró: íntegro y exacto**, sin cortes ni transformaciones que lo alteren, de modo que la persona pueda **copiarlo o transcribirlo sin error** para hacer su aporte. La presentación concreta (agrupación de dígitos, moneda, datos opcionales como el titular) queda sujeta a `[NECESITA ACLARACIÓN: Q2 — campos y obligatoriedad, incluidos titular y moneda]` y a `[NECESITA ACLARACIÓN: Q3 y Q4 — formato y presentación de IBAN y SINPE Móvil]`; la composición visual la define la fase de UX sin alterar la exactitud del número.
- **FR-004**: Los métodos de donación DEBEN presentarse en un **orden estable definido por el equipo** mediante un **valor numérico de orden** (el patrón ya usado en F3, F4 y F6: valor numérico de orden con desempate por identificador de inserción), sin exigir al visitante ninguna acción extra (ordenar, filtrar ni buscar) para ver la lista completa. `[NECESITA ACLARACIÓN: Q1 — ¿el orden es un único listado manual o la sección agrupa por tipo (cuentas IBAN y, aparte, SINPE Móvil), con orden manual dentro de cada grupo?]`
- **FR-005**: Cuando la sección **no tiene nada publicado** (o todo está en borrador), DEBE responder según la decisión del cliente `[NECESITA ACLARACIÓN: Q9 — (a) ocultarse por completo, patrón de F3 y F6; (b) mostrar un aviso sin redirigir, patrón de `/eventos` (F4); o (c) redirigir a la portada, patrón de `/ministerios` (F6)]`. Sea cual sea la opción, la navegación (barra, menú hamburguesa) y el mapa del sitio del pie DEBEN quedar **sin enlaces rotos ni huérfanos** y el sitio NUNCA DEBE mostrar secciones vacías ni campos en blanco.
- **FR-006**: La sección DEBE vivir bajo **una única entrada de navegación del sitio público** y una ruta pública propia `[NECESITA ACLARACIÓN: Q8 — ¿«Donaciones» en la ruta `/donaciones`, u otro nombre y/u otra ruta (p. ej., «Cómo apoyarnos» u «Ofrendas y diezmos»)?]`, presente en la barra o menú hamburguesa y en el mapa del sitio del pie, coherente con el patrón de F4 (`/eventos`) y F6 (`/ministerios`).

**Bilingüismo (Decisión 6 y Decisión 8)**

- **FR-007**: Los textos de la información de donaciones DEBEN poder ingresarse en español (idioma base) y opcionalmente en inglés, y la interfaz del sitio público DEBE estar siempre disponible en ambos idiomas (Decisión 6 y Decisión 8). La regla del español base aplica **por campo de texto**: si un campo de texto se rellena, su versión en español DEBE existir; **un campo vacío no obliga a nada**. **Los números de cuenta (IBAN), los números de SINPE Móvil, el tipo de método y la entidad financiera NO son contenido traducible**: son un único valor idéntico en ambos idiomas (US4, escenario 5).
- **FR-008**: Cuando un campo de texto no tiene versión en inglés y el visitante navega en inglés, el sistema DEBE mostrar la versión en español (idioma base); NUNCA DEBE mostrar un campo vacío ni una traducción automática de un contenido que no fue traducido (patrón de F3, FR-009 de F3).
- **FR-009**: El idioma se gestiona con el selector ya existente (FR-010 de F3 y F10): el cambio se aplica de inmediato, se mantiene al navegar por el sitio y se recuerda entre visitas en el mismo dispositivo; un dispositivo sin preferencia guardada se muestra en español.

**Administración desde el panel (patrón de F2 y F3)**

- **FR-010**: El panel DEBE permitir **crear, editar y eliminar métodos de donación** (cuentas IBAN y números de SINPE Móvil) con sus datos (FR-019) y sus versiones de texto en español e inglés opcional (FR-007), y DEBE permitir **mantener el contenido sobre en qué se usan las donaciones** `[NECESITA ACLARACIÓN: Q5]` y, si el cliente lo decide, el **texto introductorio** de la sección `[NECESITA ACLARACIÓN: Q11]`; el panel DEBE permitir además mantener cada elemento en estado borrador o publicado (FR-012, con la granularidad de `[NECESITA ACLARACIÓN: Q10]`).
- **FR-011**: La gestión de la información de donaciones DEBE verificar el permiso **«donaciones»**, ya **existente en el catálogo fijo de permisos de F2** (FR-015 de F2). **F7 NO crea permisos nuevos ni modifica el modelo de permisos ni nada de F2**: solo **activa** el permiso reservado. La verificación DEBE seguir el patrón de autorización de F2, **verificada en el servidor** según el rol de la cuenta (Decisión 5, constitución §IV), y el permiso es **independiente de los demás**: tener «donaciones» no da acceso a otros módulos ni otros permisos dan acceso a este. Toda operación no autorizada DEBE denegarse con un mensaje claro.
- **FR-012**: La información de donaciones DEBE poder mantenerse en estado **borrador** o **publicado** y DEBE poder publicarse o retirarse según la granularidad que decida el cliente `[NECESITA ACLARACIÓN: Q10 — ¿por cada cuenta/número (patrón de F3, F4 y F6) o la sección como un todo?]`. Solo lo publicado DEBE ser visible para el visitante; el sistema NUNCA DEBE exponer al público ningún borrador ni sus datos, por ninguna vía.
- **FR-013**: Cuando se guarda un cambio sobre un elemento **ya publicado**, los cambios DEBEN hacerse visibles de inmediato para el visitante (patrón de F3: se publica al guardar); no hay paso intermedio de borrador para lo que ya está publicado.
- **FR-014**: El sistema DEBE validar toda entrada del módulo en el panel: **campos obligatorios completos** `[NECESITA ACLARACIÓN: Q2 — ¿cuáles son obligatorios: entidad, titular, número, moneda?]`, **números con estructura y formato válidos** según FR-015 y FR-016 `[NECESITA ACLARACIÓN: Q3 y Q4]`, versión en español presente por campo de texto cuando el campo se rellena y versión en inglés opcional (Decisión 8), y textos dentro de sus límites de extensión. Ante datos inválidos o incompletos DEBE indicar qué corregir y NUNCA DEBE guardar cambios parciales. Todo contenido ingresado DEBE tratarse como datos y NUNCA DEBE ejecutarse como código ni insertar elementos activos (constitución §IV).
- **FR-015**: El número de cuenta (IBAN) DEBE **validarse** con reglas de estructura y formato antes de guardarse `[NECESITA ACLARACIÓN: Q3 — ¿qué reglas?: ¿estructura IBAN costarricense (CR + dígitos de control y longitud fija)? ¿se aceptan IBAN de otros países? ¿se admiten espacios o guiones al escribirlo y cómo se normaliza? ¿cómo se presenta al público (agrupado o tal cual)?]`. Ante un número que no cumple las reglas acordadas, el sistema DEBE rechazarlo con indicación de cómo corregirlo y NO DEBE guardarlo. El número publicado DEBE ser **exactamente el número válido guardado**, sin transformaciones que lo alteren (FR-003).
- **FR-016**: El número de SINPE Móvil DEBE **validarse** con reglas de estructura y formato antes de guardarse `[NECESITA ACLARACIÓN: Q4 — ¿qué reglas?: ¿celular costarricense de 8 dígitos? ¿con o sin prefijo `+506`? ¿se admite cualquier formato de identificador SINPE además del número de celular? ¿cómo se presenta al público?]`. Ante un número que no cumple las reglas acordadas, el sistema DEBE rechazarlo con indicación de cómo corregirlo y NO DEBE guardarlo.
- **FR-017**: Cada operación de este módulo —crear, editar, cambiar el estado de publicación, reordenar y eliminar un método de donación, y editar el contenido de uso de las donaciones— DEBE quedar registrada en el **registro de auditoría** siguiendo el **patrón ya construido en F2/F3/F4** (quién la hizo, qué hizo, sobre qué elemento y cuándo, incluidos los intentos que no se completan; registro de solo lectura que no se edita ni se borra). No se re-especifica ese mecanismo aquí.
- **FR-018**: La interfaz del panel para este módulo se entrega en español, como el resto del panel definido en F2 (Decisión 6 aplicada al sitio público).

**Contenido de la información de donaciones**

- **FR-019**: Cada método de donación DEBE poder mantener, como mínimo, su **tipo** (cuenta IBAN o SINPE Móvil), su **entidad financiera** y su **número**; los demás campos —**nombre del titular**, **moneda** (colones, dólares u otra), notas o instrucciones adicionales y cualquier otro dato— quedan sujetos a `[NECESITA ACLARACIÓN: Q2 — composición exacta, obligatoriedad de cada campo y si el nombre del titular y la moneda se publican]`. Los campos de texto DEBEN admitir los caracteres propios del español (tildes, ñ, ¿ ¡) sin transformación; los números DEBEN limitarse a lo que su formato admite (FR-015, FR-016) y sus textos DEBEN ser traducibles de forma opcional al inglés salvo los datos no traducibles de FR-007.
- **FR-020**: El uso de imagen en la sección o en cada método de donación queda **por decidir** `[NECESITA ACLARACIÓN: Q6 — ¿imagen de encabezado de la sección, logos de las entidades financieras, sin imágenes?]`. Si el cliente decide que hay imagen, DEBE ser **opcional**, un elemento sin imagen DEBE presentarse completo y coherente, toda imagen cargada DEBE llevar texto alternativo accesible y DEBE respetar las reglas de identidad vigentes (no deformar el logotipo ni recolorearlo, FR-002 de F3 y FR-024 de F10). Si decide que no hay, F7 no incluye ninguna imagen.
- **FR-021**: La sección DEBE presentar el contenido sobre **en qué se utilizan las donaciones** `[NECESITA ACLARACIÓN: Q5 — ¿un texto libre editable (uno o varios párrafos) o un listado administrable de destinos/usos (p. ej., obra social, mantenimiento, misiones), cada uno con título y descripción?]`, con sus textos traducibles de forma opcional al inglés (FR-007) y con la misma regla de publicación que el resto del módulo (FR-012). Si el cliente lo decide, la sección DEBE poder presentar además un **texto introductorio editable** (p. ej., cómo donar paso a paso) `[NECESITA ACLARACIÓN: Q11]`. La composición concreta de la presentación la define la fase de UX.

**Presencia en la portada (por confirmar)**

- **FR-022**: `[NECESITA ACLARACIÓN: Q7 — ¿F7 incluye un bloque de donaciones en la portada de F3 (patrón del bloque «Lo que viene» de F4 y del bloque de ministerios de F6) o la donación vive solo en su sección?]`. Si el cliente lo aprueba, ese bloque se especificará en la fase `clarify` con el patrón ya probado (extracto de lo publicado, tope resuelto en el servidor, botón hacia la sección completa y ocultamiento o aviso cuando no hay nada publicado), sin rediseñar el resto de la portada; si no, F7 **no toca la portada** y este requisito queda sin efecto. Mientras la pregunta esté abierta, **no se construye ningún bloque de portada**.

**Relación con el sitio público (F3 y F10), responsividad y accesibilidad**

- **FR-023**: La sección DEBE integrarse en el sitio público **naciendo con el sistema de diseño y movimiento de F10** (mismos componentes, tipografía, color con modo claro/oscuro, espaciado y movimiento respetando `prefers-reduced-motion`) y con la navegación de F10 (barra fija visible en escritorio, menú hamburguesa accesible en móvil, mapa del sitio del pie), **sin rediseñar ni re-especificar ese sistema**. Fuera de lo que responda Q7, F7 NO DEBE tocar ni alterar la portada ni ninguna otra sección (ver Out of Scope).
- **FR-024**: La sección DEBE adaptarse a cualquier dispositivo (teléfono, tableta y computadora) sin pérdida de contenido ni de funciones, y DEBE ser navegable por personas de todas las edades y niveles de experiencia: títulos claros, navegación simple, texto legible y elementos interactivos de tamaño suficiente (`idea.md` §7). Los números de cuenta DEBEN ser legibles y copiables o transcribibles en su totalidad en el ancho más estrecho probado (FR-003, SC-009).
- **FR-025**: La sección DEBE ser accesible: navegación completa por teclado, compatibilidad con lectores de pantalla, contraste suficiente (en tema claro y oscuro), posibilidad de ampliar el texto hasta el 200 % sin perder información y texto alternativo en las imágenes (si las hay, Q6). El nivel de referencia es WCAG 2.1 nivel AA (default revisable, ver Assumptions).

### Key Entities *(include if feature involves data)*

- **Método de donación** *(la forma en que una persona puede aportar: una cuenta bancaria (IBAN) o un número de SINPE Móvil)*: uno o varios `[NECESITA ACLARACIÓN: Q1]`. Cada uno tiene **tipo** (cuenta IBAN o SINPE Móvil), **entidad financiera** (el banco o entidad de SINPE) y **número** —los tres datos centrales, exactos e íntegros (FR-003)—; los demás campos (**titular**, **moneda**, notas) y su obligatoriedad quedan por confirmar `[NECESITA ACLARACIÓN: Q2]`. Lleva además **valor numérico de orden** definido por el equipo (FR-004) y **estado de publicación** (borrador/publicado, FR-012). Los textos que lo acompañen son traducibles al inglés de forma opcional (FR-007); **el número, el tipo y la entidad no lo son**.
- **Información sobre el uso de las donaciones** *(en qué se utilizan los aportes: obra social, mantenimiento del templo, misiones, etc.)*: el contenido que da transparencia al propósito de la donación. `[NECESITA ACLARACIÓN: Q5 — ¿es un texto libre editable o un listado administrable de destinos/usos, cada uno con título y descripción?]`. Si fuera un listado, cada destino tendría su propio orden y su propia publicación; si fuera un texto, sería un contenido único con su estado de publicación. Sus textos son traducibles al inglés de forma opcional (FR-007).
- **Texto introductorio de la sección** *(opcional, por confirmar)*: el mensaje de "cómo pueden ayudarnos" o los pasos para donar que enmarcan las cuentas `[NECESITA ACLARACIÓN: Q11 — ¿existe este texto editable o basta con las cuentas y el uso de las donaciones?]`.
- **Estado de publicación** (borrador/publicado): decide si el visitante ve o no cada elemento de la información de donaciones (Decisión 4); su granularidad —por elemento o por sección— queda por confirmar `[NECESITA ACLARACIÓN: Q10]`.

Relaciones: la información de donaciones se identifica por el conjunto de sus métodos (cuentas IBAN y números de SINPE) y su contenido de uso; lo gestiona una cuenta del panel con el permiso **«donaciones»** (F2, ya existente en el catálogo: F7 no crea permisos nuevos ni toca nada de F2) y cada operación sobre él queda registrada en la auditoría siguiendo el patrón de F2/F3/F4 (FR-017). El visitante solo ve lo publicado (FR-012), en el orden manual definido por el equipo (FR-004), y dona por sus propios medios (transferencia bancaria o SINPE Móvil) con los datos publicados: el sitio **no procesa ni registra la donación** (ver Out of Scope). La composición concreta de la presentación pública (listado, fichas, agrupación visual) la define la fase de UX.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de los visitantes puede ver los métodos de donación publicados y el uso de las donaciones sin autenticación, y encuentra **cómo donar** (a qué cuenta o número) en menos de 30 segundos desde que abre la sección.
- **SC-002**: 0 información de donaciones en borrador visible para el público en las pruebas de aceptación (incluida cualquier respuesta del sistema hacia el visitante).
- **SC-003**: Todo cambio guardado sobre un elemento publicado (incluida la corrección de un número de cuenta) es visible en el sitio público desde la primera carga posterior al guardado, sin pasos adicionales.
- **SC-004**: Un miembro del equipo crea y publica un método de donación nuevo en menos de 2 minutos desde el panel.
- **SC-005**: El 100% de las operaciones del módulo verifican el permiso **«donaciones»** en el servidor: 0 operaciones completadas sin permiso en las pruebas de aceptación (incluidos los intentos forzados y el cruce con otros permisos) y 100% denegadas con mensaje claro.
- **SC-006**: Con el idioma en inglés, el 100% de los textos de la interfaz y de los contenidos está traducido y 0 campos aparecen vacíos por falta de traducción; y en el 100% de los casos los números de cuenta (IBAN) y de SINPE Móvil son **idénticos** en ambos idiomas (0 alteraciones por cambio de idioma).
- **SC-007**: La sección es usable en teléfono, tableta y computadora: 0 pérdida de elementos o funciones entre dispositivos y 0 desplazamiento horizontal en el ancho de pantalla más estrecho probado (320 px).
- **SC-008**: La sección supera una revisión de accesibilidad con 0 hallazgos críticos (referencia WCAG 2.1 nivel AA, en tema claro y oscuro) y es recorrerla y operarla por completo con teclado y con lector de pantalla.
- **SC-009**: El 100% de los números publicados se muestra **íntegro, exacto y copiable o transcribible** (0 cortes, 0 caracteres alterados), y **0 números mal formados** son aceptados al guardar ni renderizados al visitante en las pruebas de aceptación (las reglas concretas dependen de Q3 y Q4).
- **SC-010**: 0 secciones vacías ni campos en blanco: la sección sin nada publicado responde según Q9 sin mostrar vacíos, y la navegación (barra, menú hamburguesa y mapa del sitio) queda sin enlaces rotos ni huérfanos en el 100% de los estados probados.
- **SC-011**: El 100% de las operaciones del módulo de las pruebas de aceptación (crear, editar, publicar/retirar, reordenar, eliminar y editar el contenido de uso) queda registrado en la auditoría siguiendo el patrón de F2/F3/F4 (quién, qué, sobre qué y cuándo), y 0 registros de auditoría pueden editarse ni borrarse.
- **SC-012**: En pruebas con personas de distintas edades, incluidas con poca experiencia en internet, al menos el 90% identifica **cómo puede donar** (a qué cuenta o número) y **en qué se usan las donaciones**, sin ayuda.
- **SC-013**: 0 eventos, actividades, grupos de conexión ni ministerios aparecen en la sección de donaciones en las pruebas de aceptación: la distinción conceptual de FR-001 se observa en el 100% de los casos probados.

## Out of Scope

Queda explícitamente **fuera del alcance** de F7:

- **Procesar donaciones en línea**: pasarelas de pago, cobro con tarjeta, PayPal, enlaces de cobro SINPE, apps bancarias, botones de "donar ahora" con monto, o cualquier flujo en el que el sitio reciba, intermediar o registre el dinero. La persona dona por sus propios medios con los datos publicados.
- **Montos, metas, campañas de recaudación, termómetros de progreso y estadísticas de donaciones**.
- **Registro de donantes y de donaciones**: historial, comprobantes, recibos, cartas de agradecimiento, declaraciones tributarias y cualquier dato personal de quien dona (el sitio ni siquiera sabe quién donó).
- **Verificación de la existencia, vigencia o titularidad de las cuentas**: el sistema solo valida la estructura y el formato de los números (FR-015, FR-016, según Q3 y Q4); no comprueba que la cuenta exista, que siga activa ni que pertenezca a la iglesia. La corrección es manual, editando el dato.
- **Avisos de cuentas antiguas**: si una cuenta se reemplaza o se elimina, el sistema no mantiene ni muestra números anteriores ni avisa a quien ya los copió (default revisable, coherente con "sin historial de versiones" de F3).
- **QR de SINPE, códigos de pago, enlaces profundos a apps bancarias o plantillas de mensaje de WhatsApp para donar**: cualquier elemento de este tipo es una ampliación que se decidiría aparte.
- **Donaciones en especie, voluntariado, apadrinamiento o "otras formas de ayudar"**: acercarse a servir corresponde a los ministerios (F6) y a los grupos de conexión (F5); F7 cubre el apoyo económico.
- **Eventos y actividades (F4), grupos de conexión (F5) y ministerios (F6)**: conceptos y módulos distintos; F7 no los muestra ni los administra.
- **Gestión de usuarios, roles y permisos** (F2): F7 solo **activa** el permiso **«donaciones»** ya existente en el catálogo de F2 (FR-015 de F2): **no crea permisos nuevos** y **no toca el modelo de permisos ni nada de F2** (Decisión 9: esa gestión sigue intacta).
- **Bloque de la portada**: mientras Q7 esté abierta, F7 **no toca la portada**; si el cliente lo aprueba, el bloque entra al alcance con su propia especificación en `clarify` y todo lo demás de este aparte se mantiene.
- **Rediseño o sistema de diseño** (F10 y F10B): F7 **aplica** el sistema de diseño y movimiento de F10 ya entregado (sitio público) y nace en el panel con los patrones actuales de F2/F3; el rediseño visual del panel es F10B y no forma parte de F7. Tampoco se re-especifican la navegación, el selector de idioma ni el modo oscuro (F10).
- **Bilingüismo del panel de administración**: el panel se entrega en español (como F2, F3 y F4); la Decisión 6 aplica al sitio público.
- **Traducción automática de contenidos**: el inglés se ingresa manualmente; si un campo no tiene inglés, se muestra en español (FR-008).
- **Historial de versiones, papelera ni recuperación de elementos eliminados, y programación de publicación con fecha futura** (default revisable, coherente con F3).
- **Despliegue a producción** y cambios de infraestructura: sigue pendiente a nivel de proyecto.

## Assumptions

Ninguna de las reglas de negocio señaladas con `[NECESITA ACLARACIÓN]` se da por supuesta: están abiertas y se resuelven con el cliente en la fase `clarify` (ver *Preguntas abiertas (fase `clarify`)*). Lo que sigue son las decisiones ya establecidas que F7 aplica sin reabrirlas y los defaults razonables tomados donde la descripción no fija una regla; **los defaults son provisionales y revisables** por el humano.

**Decisiones ya establecidas que F7 aplica sin reabrirlas**

- **Las donaciones entran al MVP** (Decisión 2) y **su información —IBAN/SINPE y su uso— se administra desde el panel** (Decisión 7); F7 depende de F2 y F3.
- **Estados borrador/publicado** (Decisión 4, patrón de F3, F4 y F6): solo lo publicado es visible; se publica al guardar (FR-012, FR-013). La **granularidad** de la publicación queda abierta en Q10.
- **Permisos por módulo agrupados en roles** (Decisión 5): F7 activa el permiso **«donaciones»** ya existente en el catálogo fijo de F2 (FR-015 de F2); **F7 no crea permisos nuevos ni toca el modelo de permisos ni nada de F2** y verifica la autorización en el servidor (FR-011, constitución §IV).
- **Bilingüismo** (Decisiones 6 y 8): español base por campo de texto (obligatorio solo si el campo se rellena), inglés opcional por contenido con respaldo al español; la interfaz del sitio público siempre bilingüe; el panel en español (FR-007 a FR-009, FR-018). Los números y datos de la cuenta no son traducibles (FR-007).
- **Auditoría**: F7 sigue el **patrón ya construido en F2/F3/F4** (registro de solo lectura con quién, qué, sobre qué y cuándo, incluidos los intentos fallidos) sin re-especificarlo (FR-017).
- **Modelo de cuentas y roles de F2** (Decisión 9): intacto; F7 no lo cambia.
- **Sistema de diseño y movimiento**: la sección nace con el sistema de F10 aplicado (modo claro/oscuro, movimiento con `prefers-reduced-motion`, navegación y mapa del sitio); no se re-especifica aquí (FR-023).

**Defaults revisables**

- **Orden de presentación**: valor numérico de orden con desempate por identificador de inserción (el patrón ya usado en F3, F4 y F6); la pregunta abierta de Q1 es solo si además hay agrupación por tipo.
- **Límites de extensión**: los textos largos (p. ej., el uso de las donaciones) con un límite del orden de **2.000 caracteres** por campo (coherente con F4 y F6); los números y nombres cortos con límites cortos de texto (el detalle lo fija el arquitecto en el plan). Los límites son default revisable, no regla de negocio.
- **Duplicados**: un duplicado exacto (mismo tipo, misma entidad y mismo número) se rechaza o se señala como duplicado (default revisable, coherente con F2, F3 y F6).
- **Sin límite artificial de métodos de donación** en la sección, y sin filtros ni búsqueda en el MVP: se muestran todos en el orden definido (default revisable).
- **Eliminación**: definitiva, con confirmación explícita y registrada en la auditoría; no hay papelera ni recuperación (default revisable, coherente con F3).
- **Ediciones simultáneas**: si dos personas editan el mismo elemento al mismo tiempo, se conserva la última versión guardada completa (default revisable, coherente con F3).
- **Idioma del panel**: se entrega en español, como F2, F3 y F4 (Decisión 6 aplica al sitio público).
- **Nivel de accesibilidad**: se toma **WCAG 2.1 nivel AA** como referencia (default revisable); el requisito de fondo es el de `idea.md` §7: apto para personas de todas las edades y dispositivos.
- **Composición visual de la sección y de las fichas**: la define la fase de UX sin cambiar el contenido, los datos exactos de FR-003 ni las decisiones de esta spec.

## Dependencias

- **F2 — Acceso y gestión de usuarios (terminada)**: panel con inicio de sesión, cuentas y roles con permisos por módulo, **permiso «donaciones» ya reservado y existente en el catálogo desde F2** (FR-015 de F2: F7 no crea permisos nuevos ni toca nada de F2), patrón de autorización en el servidor y registro de auditoría.
- **F3 — Portada e información general (terminada)**: sitio público, bilingüismo con respaldo al español (Decisiones 6 y 8), estados borrador/publicado y la regla de no mostrar secciones vacías, que F7 hereda salvo la decisión que responda Q9.
- **F10 — Sistema de diseño y movimiento (terminada)**: sistema visual y de movimiento, navegación (barra fija en escritorio, menú hamburguesa en móvil), selector de idioma, modo oscuro y mapa del sitio del pie; F7 nace aplicándolo (FR-023).
- **Precedentes de patrón**: F4 y F6 (orden manual con valor numérico de orden, estados borrador/publicado, ruta pública bajo una única entrada de navegación, bloque de portada si Q7 lo aprueba). F7 **no depende funcionalmente** de F4, F5 ni F6 y puede construirse en paralelo a ellas (roadmap §3).
- **Pendiente antes de planificar**: las preguntas Q1–Q11 deben resolverse en la fase `clarify`; varias de ellas (Q1, Q2, Q3, Q4, Q5, Q9 y Q10) definen reglas de negocio sin las cuales la implementación quedaría a interpretación.

## Preguntas abiertas (fase `clarify`)

Marcadas con `[NECESITA ACLARACIÓN]` en el punto exacto del texto. Ninguna se resuelve por suposición. Se indica la opción más simple o coherente con lo ya construido, **solo como referencia para el cliente**: no se aplica hasta que él decida.

| # | Pregunta (enunciado corto) | Dónde aplica | Referencia para el cliente |
|---|---|---|---|
| Q1 | ¿Listado administrable de varias cuentas IBAN y varios SINPE Móvil, o una sola cuenta y un solo número? ¿Y un único listado o agrupado por tipo? | FR-001, FR-004, FR-019, Key Entities | El patrón del proyecto es listado administrable con orden manual |
| Q2 | ¿Qué campos lleva cada cuenta y cuáles son obligatorios (entidad, titular, número, moneda)? ¿Se publica el titular? ¿La moneda es campo aparte? | FR-003, FR-014, FR-019 | Mínimo: tipo, entidad y número; el resto opcional |
| Q3 | ¿Qué validación y formato se exige al IBAN (estructura costarricense, IBAN internacional, espacios, presentación)? | FR-015, FR-003, Edge Cases | Validar estructura y longitud, guardar el número exacto |
| Q4 | ¿Qué validación y formato se exige al SINPE Móvil (8 dígitos, `+506`, otros identificadores SINPE, presentación)? | FR-016, FR-003, Edge Cases | Celular costarricense de 8 dígitos, con o sin prefijo |
| Q5 | «En qué se usan las donaciones»: ¿texto libre editable o listado administrable de destinos/usos? | FR-002, FR-010, FR-021, Key Entities | Texto libre editable es lo mínimo que pide `idea.md` |
| Q6 | ¿Hay imágenes (encabezado de la sección, logos de banco) o la sección es solo datos y texto? | FR-020, Edge Cases | Sin imágenes es lo más simple; imagen opcional de encabezado es la alternativa |
| Q7 | ¿F7 incluye un bloque de donaciones en la portada de F3 o solo la sección? | FR-022, Nota de alcance | `idea.md` solo pide una sección; F4 y F6 sí tienen bloque |
| Q8 | ¿Nombre y ruta de la sección: «Donaciones» en `/donaciones`, u otro? | FR-006 | El patrón es «Donaciones» en `/donaciones` |
| Q9 | Sin nada publicado: ¿ocultar la sección (F3/F6), aviso sin redirigir (F4 en `/eventos`) o redirección a la portada (F6 en `/ministerios`)? | FR-005, US1/US3, Edge Cases | Ocultar por completo y redirigir a la portada (última decisión del cliente en F6) |
| Q10 | ¿La publicación es por cada cuenta/número o la sección se publica como un todo? | FR-010, FR-012, US3 | Por cada elemento (patrón de F3, F4 y F6) |
| Q11 | ¿La sección lleva además un texto introductorio editable («cómo donar») o basta con las cuentas y el uso? | FR-002, FR-010, FR-021, Key Entities | `idea.md` habla de "cómo pueden ayudarnos": un intro editable es coherente |

### Sobre el resto del documento

Los puntos que no son reglas de negocio —estados borrador/publicado (Decisión 4), permisos por módulo con el permiso **«donaciones» ya existente en F2 y sin permisos nuevos (Decisión 5, FR-015 de F2), bilingüismo español base / inglés opcional con respaldo (Decisiones 6 y 8), auditoría de solo lectura (patrón de F2/F3/F4), navegación y sistema de diseño (F10), responsividad y accesibilidad (`idea.md` §7) y que el sitio **no procesa donaciones**— están definidos por decisiones o specs ya aprobadas y se recogen aquí sin reabrirlas. Todo lo marcado `[NECESITA ACLARACIÓN]` (Q1–Q11) queda para la fase `clarify`; una vez resueltas, esta spec se actualiza y pasa a la puerta 1 (aprobación humana) antes de cualquier plan técnico.
