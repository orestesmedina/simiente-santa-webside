# Feature Specification: Portada e información general (F3)

**Feature Branch**: `003-portada-info-general`

**Created**: 2026-10-09

**Status**: Aprobada (2026-10-09) — ajuste de la memoria del idioma incorporado por decisión del humano del 2026-10-09

**Input**: User description: "F3 — Portada e información general (roadmap §3, `docs/producto/roadmap.md`): Cualquier visitante ve una portada con la identidad de la iglesia, quiénes somos, horario de servicios, canales de WhatsApp y redes sociales, en español e inglés; el equipo edita esa información desde el panel." Decisiones aplicables del roadmap: **Decisión 4** (contenido con estados borrador/publicado; solo lo publicado es visible al visitante), **Decisión 6** (el sitio público es bilingüe, español e inglés, con selección de idioma) y **Decisión 8** (cada contenido puede ingresarse en ambos idiomas o solo en español: el español es el idioma base y el inglés es opcional por contenido; la interfaz, en cambio, siempre es bilingüe). Restricción de `idea.md` §7 y del encargo (`idea.md` intro): el sitio debe adaptarse a cualquier dispositivo y ser apto para personas de todas las edades, incluidas con poca experiencia de internet. **Dependencia**: F2 (terminada) provee el acceso autenticado al panel, los roles con permisos por módulo, el patrón de autorización y el registro de auditoría; la spec de F2 ya reservó el permiso **«portada e información general»** dentro de su catálogo (FR-015 de F2) y esta funcionalidad lo activa. **Identidad de marca**: el cliente entregó el **Manual de Identidad de la Iglesia Simiente Santa** (`resources/MANUAL DE MARCA.pdf`, logotipo `resources/simiente.jpeg`), que es la guía visual que la portada DEBE respetar (FR-002); sus datos de marca y contenido real se recogen como referencia en *Identidad de marca — referencia para la fase de plan y UX*. **Aclaraciones resueltas con el humano el 2026-10-09**: las 10 preguntas abiertas de la primera versión quedan resueltas (ver *Aclaraciones resueltas (2026-10-09)*).

> **Nota de alcance y terminología**: esta funcionalidad es la **primera cara pública del sitio**: la portada (página de inicio) que cualquier persona puede ver sin cuenta ni registro. Se reutiliza la terminología de F2: el **visitante** (el "usuario común" de `idea.md` §2: ve el sitio público, no tiene cuenta y no se registra) y el **usuario del panel** (el "usuario del sistema" y el "usuario administrador" de `idea.md` §2: tiene cuenta, inicia sesión y gestiona el contenido). F3 entrega: la portada pública con identidad de la iglesia, «quiénes somos», horario de servicios, canales de WhatsApp, redes sociales y datos de contacto, bilingüe (español base, inglés opcional con fallback al español), con estados borrador/publicado por elemento, y la gestión de todo ese contenido desde el panel con el permiso del módulo y registro de auditoría. Las demás secciones del sitio (eventos, actividades, grupos, ministerios, donaciones, noticias, medios) llegan en F4–F9 y **no** forman parte de F3.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ver la portada con la información general de la iglesia (Priority: P1)

Como visitante (miembro de la iglesia o persona nueva), quiero abrir el sitio y ver de inmediato quiénes son, cuándo se reúnen y cómo contactarlos —WhatsApp, redes sociales, dirección, correo y teléfono—, para conocer la iglesia, saber cuándo asistir y poder escribirles o visitarlos.

**Why this priority**: Es el corazón de F3 y la primera cara pública de todo el sitio: resuelve el problema central de `idea.md` §1 y §3 (las personas no saben que la iglesia existe ni cuándo asiste). Sin esta historia no hay valor visible para el público y todo lo demás de F3 queda sin destinatario.

**Independent Test**: Se prueba de forma independiente con la información general publicada: un visitante sin sesión abre la portada y encuentra la identidad de la iglesia, «quiénes somos», el horario de servicios, los canales de WhatsApp, las redes sociales y los datos de contacto; al cambiar un dato publicado y recargar, el visitante ve el dato actualizado. Entrega valor por sí sola: el público ya sabe quiénes son, cuándo reunirse y cómo contactarlos.

**Acceptance Scenarios**:

1. **Dado** la portada con información publicada, **cuando** un visitante abre el sitio sin iniciar sesión, **entonces** ve la identidad de la iglesia, la sección «quiénes somos», el horario de servicios, los canales de WhatsApp, las redes sociales y los datos de contacto (dirección física, correo electrónico y teléfono), sin necesidad de ninguna cuenta.
2. **Dado** que el equipo actualiza y publica un cambio en el horario de servicios, **cuando** el visitante recarga la portada, **entonces** ve el horario actualizado.
3. **Dado** un canal de WhatsApp publicado, **cuando** el visitante lo activa (clic o toque), **entonces** se abre ese canal para escribir a la iglesia.
4. **Dado** una red social publicada, **cuando** el visitante la elige, **entonces** se abre el perfil de la iglesia en esa red.
5. **Dado** la portada publicada, **cuando** una persona nueva la recorre, **entonces** puede saber quién es la iglesia, en qué creen, cuándo reunirse, dónde están y cómo contactarlos sin buscar en ningún otro sitio.

---

### User Story 2 - Editar la información general desde el panel (Priority: P1)

Como usuario del panel con el permiso de «portada e información general», quiero editar la identidad de la iglesia (nombre oficial, lema/misión/visión, logotipo e imagen de portada), «quiénes somos», el horario de servicios, los canales de WhatsApp, las redes sociales y los datos de contacto —en español y, si quiero, también en inglés— para mantener la información al día sin depender de un desarrollador.

**Why this priority**: Es la segunda mitad del alcance de F3 y el criterio de éxito de `idea.md` §4: que el contenido sea dinámico, no estático. Sin edición desde el panel, la portada sería un cartel fijo y el equipo volvería a depender de terceros para cada cambio. Reutiliza el patrón de F2 (permisos por módulo ya reservados), por lo que es construible y testeable de forma independiente.

**Independent Test**: Se prueba de forma independiente con una cuenta con el permiso del módulo y otra sin él: la primera edita un dato válido, guarda y el sistema lo conserva; con campos obligatorios vacíos, enlaces mal formados, correos o teléfonos inválidos o una red social repetida, el sistema indica qué corregir y no guarda nada a medias; la segunda recibe una denegación clara en toda operación del módulo. Entrega valor por sí sola: el equipo ya mantiene su propia información.

**Acceptance Scenarios**:

1. **Dado** un usuario del panel con el permiso de «portada e información general», **cuando** edita la información general con datos válidos y guarda, **entonces** el sistema confirma la operación y conserva los cambios.
2. **Dado** una cuenta sin ese permiso, **cuando** intenta editar la información general, **entonces** el sistema impide la operación con un mensaje claro y no cambia nada.
3. **Dado** un formulario con datos inválidos o incompletos (campo obligatorio vacío, enlace de WhatsApp o de red social mal formado, correo o teléfono con formato inválido, contenido en español vacío), **cuando** envío el formulario, **entonces** el sistema indica qué corregir y no guarda ningún cambio parcial.
4. **Dado** el formulario de edición, **cuando** la persona ingresa la versión en inglés de un contenido, **entonces** el sistema conserva esa versión junto a la versión en español.
5. **Dado** el formulario de edición, **cuando** la persona deja vacío el campo en inglés de un contenido, **entonces** puede guardar igualmente, porque el inglés es opcional por contenido (Decisión 8).
6. **Dado** el formulario de identidad, **cuando** la persona sube el logotipo o la imagen de portada, **entonces** el sistema los conserva y la portada los muestra respetando las reglas de uso del logotipo del Manual de Identidad (no deformar, no cambiar los colores, no girar, no agregar efectos).
7. **Dado** una cuenta sin el permiso del módulo, **cuando** intenta forzar el acceso a la gestión de la información general por cualquier vía de la interfaz, **entonces** el sistema lo impide con un mensaje claro y sin revelar detalles internos.

---

### User Story 3 - Controlar la publicación de cada elemento de la portada (Priority: P2)

Como usuario del panel con el permiso del módulo, quiero publicar o retirar cada elemento de la portada por separado —un servicio del horario, un canal de WhatsApp, una red social, un texto—, para que el público solo vea información revisada y nunca cambios a medias.

**Why this priority**: La Decisión 4 (contenido borrador/publicado) es parte explícita del alcance y protege la imagen pública de la iglesia: sin control de publicación, cada edición se haría visible de inmediato y un error cualquiera sería público. Va como P2 porque US2 sola ya entrega un MVP usable (editar y conservar la información), mientras que esta historia añade el control de visibilidad elemento a elemento.

**Independent Test**: Se prueba de forma independiente con un elemento en borrador y otro publicado: el borrador no aparece para el visitante y el resto de la portada se ve coherente; al publicarlo aparece; al devolverlo a borrador desaparece; una sección cuyos elementos están todos en borrador se oculta por completo. Todo ello sin que el visitante necesite cuenta ni sesión. Entrega valor por sí sola: el equipo decide qué se hace público y cuándo, elemento a elemento.

**Acceptance Scenarios**:

1. **Dado** un elemento de la portada en estado borrador, **cuando** un visitante abre la portada, **entonces** ese elemento no se le muestra y el resto de la portada se ve coherente, sin huecos ni campos en blanco.
2. **Dado** una sección cuyos elementos están todos en borrador (o que aún no tiene ningún elemento), **cuando** un visitante abre la portada, **entonces** la sección se oculta por completo.
3. **Dado** un elemento en borrador, **cuando** el usuario del panel lo publica, **entonces** el elemento aparece en la portada pública.
4. **Dado** un elemento publicado, **cuando** el usuario del panel lo devuelve a borrador, **entonces** deja de ser visible para el visitante.
5. **Dado** un elemento publicado, **cuando** el usuario del panel edita sus datos y guarda, **entonces** los cambios se hacen visibles de inmediato para el visitante (decisión del humano del 2026-10-09: se publica al guardar).
6. **Dado** cualquier contenido en borrador, **cuando** se consulta el sitio público por cualquier vía, **entonces** el sistema NUNCA expone ese contenido ni sus datos al público.
7. **Dado** el módulo en el panel, **cuando** el usuario revisa los contenidos, **entonces** ve el estado (borrador/publicado) de cada elemento para saber qué es público hoy.

---

### User Story 4 - Ver la portada en inglés (Priority: P2)

Como visitante que prefiere el inglés, quiero cambiar el idioma de la portada con un selector, para entender la información de la iglesia aunque mi idioma base no sea el español.

**Why this priority**: La Decisión 6 y la Decisión 8 hacen del bilingüismo un compromiso del sitio: el objetivo de `idea.md` es dar a conocer la iglesia "en nuestra provincia y por qué no internacionalmente". Va como P2: el valor central (la información visible) ya está en US1, y el inglés amplía el alcance del público sin bloquear el resto.

**Independent Test**: Se prueba de forma independiente con la portada publicada en ambos idiomas y con un contenido que solo tiene español: al elegir inglés, la interfaz y los contenidos traducidos se muestran en inglés, el contenido sin inglés se muestra en español y no aparece ningún campo vacío; al volver a español, todo se muestra en español. Se comprueba además que la elección persiste al navegar y en una visita posterior desde el mismo dispositivo, y que un dispositivo sin preferencia guardada se muestra en español. Entrega valor por sí sola: el público anglófono puede usar la portada.

**Acceptance Scenarios**:

1. **Dado** la portada en español, **cuando** el visitante elige inglés en el selector de idioma, **entonces** la interfaz y los contenidos con versión en inglés se muestran en inglés, y la elección se mantiene al navegar por el resto del sitio y se recuerda entre visitas en el mismo dispositivo.
2. **Dado** un contenido que solo tiene versión en español, **cuando** el visitante navega en inglés, **entonces** ese contenido se muestra en español (idioma base), nunca vacío ni traducido automáticamente.
3. **Dado** un dispositivo sin ninguna preferencia de idioma guardada (primera visita), **cuando** el visitante abre la portada, **entonces** se muestra en español, que es el idioma base (Decisión 8).
4. **Dado** la portada en inglés, **cuando** el visitante vuelve a elegir español en el selector, **entonces** toda la página se muestra de nuevo en español sin pérdida de información.
5. **Dado** un visitante que eligió un idioma en una visita anterior desde el mismo dispositivo, **cuando** vuelve a abrir el sitio, **entonces** el idioma se muestra según su preferencia guardada, sin tener que elegirlo de nuevo.

---

### User Story 5 - Navegar la portada en cualquier dispositivo y de forma accesible (Priority: P2)

Como visitante de cualquier edad, dispositivo y nivel de experiencia con internet, quiero que la portada se adapte a mi pantalla y sea fácil de usar, para encontrar la información sin dificultades.

**Why this priority**: Es una restricción explícita del cliente (`idea.md` §7 y su introducción: personas de todas las edades, "adultos mucho más mayores que saben lo básico de internet"). Va como P2 porque es una capa transversal sobre US1 y US4 que se puede verificar por separado, pero no puede postergarse más allá del MVP sin incumplir el encargo.

**Independent Test**: Se prueba de forma independiente sobre la portada publicada: se abre en teléfono, tableta y computadora y se comprueba que todas las secciones y enlaces están disponibles sin desplazamiento horizontal; se recorre con teclado y con lector de pantalla; se comprueba legibilidad y contraste. Entrega valor por sí sola: cualquier persona puede usar la portada desde cualquier dispositivo.

**Acceptance Scenarios**:

1. **Dado** un teléfono con pantalla estrecha, **cuando** el visitante abre la portada, **entonces** ve todas las secciones sin desplazamiento horizontal y puede activar cada enlace con comodidad (elementos de toque suficientemente grandes).
2. **Dado** una persona con poca experiencia en internet, **cuando** recorre la portada, **entonces** encuentra cada sección con títulos claros y una navegación simple y sin menús ocultos ni gestos desconocidos.
3. **Dado** una persona que usa el teclado o un lector de pantalla, **cuando** navega la portada, **entonces** puede recorrer y activar todos los contenidos y enlaces en un orden comprensible.
4. **Dado** una persona con baja visión, **cuando** usa la portada, **entonces** el texto se lee con contraste suficiente y puede ampliarse sin que se pierda información.
5. **Dado** cualquier dispositivo, **cuando** el visitante abre la portada, **entonces** el contenido y las funciones disponibles son los mismos que en los demás dispositivos.

---

### Edge Cases

- **Contenido en borrador**: nunca debe ser visible para el visitante (US3, FR-013), ni en la página ni en ninguna respuesta del sistema hacia el público; un borrador no debe "filtrarse" por ninguna vía.
- **Sección sin ningún elemento publicado** (p. ej., aún no hay horario publicado, o todos sus elementos están en borrador): la sección **se oculta por completo** (decisión del humano del 2026-10-09); la portada se muestra coherente, sin secciones vacías ni campos en blanco.
- **Contenido publicado que se edita**: los cambios guardados se hacen **visibles de inmediato** para el visitante (decisión del humano del 2026-10-09: se publica al guardar); no hay paso intermedio de borrador para lo que ya está publicado.
- **Contenido sin versión en inglés navegando en inglés**: se muestra la versión en español (FR-009); nunca un campo vacío, un texto en otro idioma sin aviso ni una traducción automática (no existe traducción automática).
- **Contenido en inglés sin versión en español**: no debe poder guardarse; el español es el idioma base y obligatorio (Decisión 8, FR-015), y el sistema lo impide indicando qué falta.
- **Enlace de WhatsApp o de red social mal formado**: se rechaza con indicación de cómo corregirlo y no se guarda nada (FR-015).
- **Red social repetida o fuera del catálogo**: el catálogo de redes es fijo (Facebook, Instagram, YouTube, TikTok, Spotify y equivalentes) y admite **un solo enlace por red** (decisión del humano del 2026-10-09); el sistema rechaza un segundo enlace para la misma red y una red fuera del catálogo, indicando el problema.
- **Canales de WhatsApp**: puede haber varios (número para mensaje directo y/o enlace de grupo), cada uno con nombre/propósito y un solo destino; un canal duplicado exacto debe rechazarse o señalarse como duplicado (default revisable, ver Assumptions).
- **Correo electrónico o teléfono de contacto con formato inválido**: se rechaza con indicación de cómo corregirlo; no se guarda el dato (FR-015).
- **Texto con caracteres especiales del español** (tildes, ñ, ¿ ¡) **y con emojis**: se conserva íntegro en el panel y en la portada, sin corromperse ni perderse.
- **«Quiénes somos» más largo que el límite de extensión** (1.000 caracteres por defecto, default revisable): el sistema lo indica al excederse y no guarda hasta corregirlo. Los demás textos largos deben acomodarse sin romper la portada.
- **Dos personas editando la información general al mismo tiempo**: se conserva la última versión guardada completa (default revisable; ver Assumptions); no debe perderse el conjunto de datos ni quedar una mezcla incoherente sin aviso.
- **Logotipo e imagen de portada**: se suben desde el panel; deben respetar las reglas del Manual de Identidad (el logotipo no se deforma, no se le cambian los colores, no se gira, no se le agregan efectos) y llevar texto alternativo accesible (FR-019); si la carga o el archivo falla, la portada no se rompe y se muestra sin esa imagen con un aviso al usuario del panel.
- **Enlaces que dejan de funcionar** (una red cambia de perfil, un grupo de WhatsApp caduca): el sitio solo enlaza; no verifica la vigencia de los destinos (ver Out of Scope). La corrección la hace el equipo editando el enlace.
- **Operación sin permiso sobre el módulo** (lectura de la gestión o edición): denegada con mensaje claro, igual que en F2 (FR-012).
- **Edición de la información general**: queda registrada en la auditoría de F2 (FR-017); un fallo de la edición no debe dejar registro de una operación que se aplicó, ni aplicarse sin registro.

### Errores esperados

Todos los mensajes deben ser comprensibles para personas no técnicas y nunca deben exponer información interna del sistema (coherente con F1 y F2).

| Situación | Qué debe ver la persona |
|---|---|
| Campo obligatorio vacío al editar la información general (p. ej., nombre de la iglesia, texto de «quiénes somos» en español, dirección, correo o teléfono) | Indicación de qué campo corregir, sin guardar nada |
| Correo electrónico o teléfono de contacto con formato inválido | Indicación de cómo corregir el dato, sin guardar nada |
| Enlace de WhatsApp o de red social mal formado | Indicación de cómo corregir el enlace, sin guardar nada |
| Red social repetida (ya hay un enlace para esa red) o fuera del catálogo | Mensaje que indica que cada red del catálogo admite un solo enlace y cuál es el problema |
| «Quiénes somos» que excede el límite de extensión | Mensaje con el límite vigente y lo que falta recortar |
| Intento de guardar un contenido sin versión en español (idioma base obligatorio) | Mensaje que indica que falta el contenido en español |
| Operación del módulo sin el permiso correspondiente | Mensaje claro de acceso denegado, sin detalles internos |
| Sección de la portada sin ningún elemento publicado | La sección se oculta; en ningún caso se muestra vacía ni con campos en blanco |
| Logotipo o imagen de portada con archivo inválido o carga fallida | Mensaje que indica el problema con el archivo; la portada no se rompe |
| Error inesperado del sistema al editar o publicar | Mensaje genérico sin información interna, sin perder lo que la persona ya haya escrito sin aviso |

## Requirements *(mandatory)*

### Functional Requirements

**Portada pública (visitante, sin autenticación)**

- **FR-001**: El sitio DEBE mostrar una portada pública accesible sin autenticación, con la identidad de la iglesia, la sección «quiénes somos», el horario de servicios, los canales de WhatsApp, las redes sociales y los datos de contacto (dirección física o ubicación, correo electrónico y teléfono). La portada es la página de inicio del sitio público.
- **FR-002**: La portada DEBE mostrar la identidad de la iglesia de forma destacada: el nombre oficial, el lema/misión/visión, el logotipo y la imagen de portada, todo ello gestionable desde el panel. El diseño de la portada DEBE respetar la identidad visual del **Manual de Identidad de la Iglesia Simiente Santa** (`resources/MANUAL DE MARCA.pdf`) y sus reglas de uso del logotipo: **no deformar, no cambiar los colores, no girar, no agregar efectos**. Los datos de marca y el contenido real del manual se recogen como referencia en *Identidad de marca — referencia para la fase de plan y UX*.
- **FR-003**: La portada DEBE incluir la sección «quiénes somos» como **texto plano** con límite de extensión (1.000 caracteres por defecto, valor revisable; ver Assumptions). Todo el contenido DEBE mostrarse como contenido y NUNCA DEBE ejecutarse como código ni insertar elementos activos (constitución §IV).
- **FR-004**: La portada DEBE incluir el horario de servicios como **lista de varios servicios**; cada servicio DEBE tener día, hora, nombre/descripción y lugar, y sus textos DEBEN poder traducirse opcionalmente al inglés.
- **FR-005**: La portada DEBE incluir los canales de WhatsApp de la iglesia, que pueden ser **números para mensaje directo y/o enlaces de grupo**; DEBE poder haber varios canales y cada uno DEBE identificarse con nombre/propósito y DEBE poder abrirse directamente desde la portada con un clic o toque.
- **FR-006**: La portada DEBE incluir las redes sociales de la iglesia dentro de un **catálogo fijo de redes** (Facebook, Instagram, YouTube, TikTok, Spotify y otras equivalentes), con **un solo enlace por red**; cada red DEBE abrir el perfil de la iglesia en esa red con un clic o toque.
- **FR-007**: Además de los canales de WhatsApp y las redes sociales, la portada DEBE incluir los **datos de contacto**: dirección física o ubicación de la iglesia, correo electrónico y teléfono.

**Bilingüismo (Decisión 6 y Decisión 8)**

- **FR-008**: Todo contenido de la portada DEBE poder ingresarse en español (idioma base, obligatorio) y opcionalmente en inglés, y la interfaz del sitio público DEBE estar siempre disponible en ambos idiomas (Decisión 6 y Decisión 8).
- **FR-009**: Cuando un contenido no tiene versión en inglés y el visitante navega en inglés, el sistema DEBE mostrar la versión en español de ese contenido (idioma base); NUNCA DEBE mostrar un campo vacío ni una traducción automática de un contenido que no fue traducido.
- **FR-010**: La portada DEBE ofrecer un selector de idioma visible (español/inglés); el cambio DEBE aplicarse de inmediato a la página visible, DEBE mantenerse al navegar por el resto del sitio y DEBE recordarse entre visitas en el mismo dispositivo (decisión del humano del 2026-10-09). Cuando el dispositivo no tiene ninguna preferencia de idioma guardada (primera visita), el sitio DEBE mostrarse en español, que es el idioma base (Decisión 8).

**Administración desde el panel (patrón de F2)**

- **FR-011**: El panel DEBE permitir crear y editar la información general de la portada —identidad (nombre oficial, lema/misión/visión, logotipo e imagen de portada), «quiénes somos», horario de servicios, canales de WhatsApp, redes sociales y datos de contacto— con su versión en español y su versión opcional en inglés (FR-008).
- **FR-012**: Toda operación de gestión de este módulo DEBE verificar el permiso **«portada e información general»** —ya reservado en el catálogo de F2 (FR-015 de F2)— según el rol de la cuenta, siguiendo el patrón de autorización de F2, y DEBE denegar con un mensaje claro toda operación no autorizada.
- **FR-013**: Cada elemento de la portada DEBE poder mantenerse en estado **borrador** o **publicado** y DEBE poder publicarse o retirarse **por separado** (decisión del humano del 2026-10-09: publicación por secciones/elementos). Solo lo publicado DEBE ser visible para el visitante; el sistema NUNCA DEBE exponer al público ningún contenido en borrador ni sus datos, por ninguna vía. Cuando una sección no tiene ningún elemento publicado, la sección DEBE ocultarse por completo: la portada NUNCA DEBE mostrar secciones vacías ni campos en blanco.
- **FR-014**: Cuando se guarda un cambio sobre un elemento **ya publicado**, los cambios DEBEN hacerse visibles de inmediato para el visitante (decisión del humano del 2026-10-09: se publica al guardar); no hay paso intermedio de borrador para lo que ya está publicado.
- **FR-015**: El sistema DEBE validar toda entrada del módulo en el panel: campos obligatorios completos (nombre de la iglesia, texto de «quiénes somos» en español, dirección física, correo electrónico y teléfono de contacto), correo y teléfono con formato válido (el teléfono con el mismo criterio telefónico razonable de F2: dígitos, con espacios, guiones o paréntesis como separadores habituales, un prefijo internacional opcional y al menos 7 dígitos), enlaces de WhatsApp y de redes sociales con formato válido, una sola red del catálogo con un enlace (rechazando repeticiones y redes fuera del catálogo), versión en español siempre presente, versión en inglés opcional (Decisión 8) y «quiénes somos» dentro del límite de extensión (FR-003). Ante datos inválidos o incompletos DEBE indicar qué corregir y NUNCA DEBE guardar cambios parciales. Todo contenido ingresado DEBE tratarse como datos y NUNCA DEBE ejecutarse como código ni insertar elementos activos (constitución §IV).
- **FR-016**: La interfaz del panel para este módulo se entrega en español, como el resto del panel definido en F2 (Decisión 6 aplicada al sitio público).
- **FR-017**: Cada edición de la información general DEBE quedar registrada en el **registro de auditoría de F2** con quién la hizo, qué editó y cuándo (decisión del humano del 2026-10-09), ampliando el alcance de la auditoría de F2 a este módulo; el registro DEBE conservar el mismo carácter de solo lectura de F2 (FR-025 de F2).

**Responsividad y accesibilidad (restricción del cliente, `idea.md` §7)**

- **FR-018**: La portada DEBE adaptarse a cualquier dispositivo (teléfono, tableta y computadora) sin pérdida de contenido ni de funciones, y DEBE ser navegable por personas de todas las edades y niveles de experiencia: títulos claros, navegación simple, texto legible y elementos interactivos de tamaño suficiente.
- **FR-019**: La portada DEBE ser accesible: navegación completa por teclado, compatibilidad con lectores de pantalla, contraste suficiente, posibilidad de ampliar el texto sin perder información y texto alternativo en las imágenes (logotipo e imagen de portada). El nivel de referencia es WCAG 2.1 nivel AA (default revisable, ver Assumptions).

### Key Entities *(include if feature involves data)*

- **Identidad de la iglesia** (única): nombre oficial, lema/misión/visión, logotipo e imagen de portada, con versión en español (obligatoria) e inglés (opcional) de sus textos. Su presentación DEBE respetar el Manual de Identidad y las reglas de uso del logotipo (FR-002).
- **Quiénes somos** (única): texto **plano** descriptivo de la iglesia, con límite de extensión (1.000 caracteres por defecto, revisable), versión en español (obligatoria) y en inglés (opcional).
- **Horario de servicio** (uno o varios): cada servicio de la iglesia con día, hora, nombre/descripción y lugar; sus textos son traducibles de forma opcional al inglés.
- **Canal de WhatsApp** (uno o varios): cada canal de contacto por WhatsApp, con nombre/propósito y un único destino —número para mensaje directo o enlace de grupo—; el destino no es contenido traducible, el nombre/propósito sí puede serlo.
- **Red social** (una por red del catálogo): perfil de la iglesia en una red del catálogo fijo (Facebook, Instagram, YouTube, TikTok, Spotify y equivalentes), con un solo enlace por red.
- **Datos de contacto** (sección): dirección física o ubicación de la iglesia, correo electrónico y teléfono.
- **Estado de publicación** (borrador/publicado, por elemento): cada elemento de la portada se publica o retira por separado (decisión del humano del 2026-10-09) y decide si el visitante lo ve o no; una sección sin elementos publicados se oculta por completo.

Relaciones: la portada presenta una identidad, un «quiénes somos», uno o varios servicios del horario, uno o varios canales de WhatsApp, una red por cada red del catálogo usada y los datos de contacto. Todo ese contenido lo gestiona una cuenta del panel con el permiso «portada e información general» (F2), cada elemento lleva su estado de publicación y sus versiones en español (obligatoria) e inglés (opcional con fallback al español, FR-009), y cada edición queda registrada en la auditoría de F2 (FR-017).

### Identidad de marca — referencia para la fase de plan y UX

Fuente: **Manual de Identidad de la Iglesia Simiente Santa** entregado por el cliente (`resources/MANUAL DE MARCA.pdf`; logotipo `resources/simiente.jpeg`). La portada DEBE respetar esta identidad (FR-002); los detalles siguientes son **referencia para la fase de plan y para el diseño UX** —no son requisitos de código detallados— y su aplicación concreta (paleta, tipografías, composición) se define en UX y se verifica en la implementación frontend (ver Out of Scope). Los textos marcados como contenido real se respetan **literalmente**.

- **Misión (literal)**: «Simiente Santa es un espacio donde las personas pueden encontrarse con Dios, vivir en libertad, construir relaciones genuinas y descubrir su propósito para generar un impacto en su entorno.»
- **Visión (literal)**: «consolidarse como una iglesia que transforma vidas y comunidades, formando personas libres, seguras y con propósito, que viven una fe auténtica y generan un impacto en el mundo.»
- **Valores (literal)**: pertenencia, seguridad, esperanza, amor, alegría, propósito, unidad, autenticidad, compromiso, impacto.
- **Quiénes somos (texto base literal)**: «Simiente Santa es un espacio donde las personas pueden encontrarse con Dios de forma real, cercana y transformadora.»
- **Personalidad de marca**: **somos** cercanos, bíblicos, esperanzadores, familiares, relevantes, inspiradores; **no somos** religiosos, fríos, complicados, institucionales, tradicionalistas.
- **Voz de marca**: habla como una persona real, cercana, que ama a las personas; evita lenguaje religioso complejo, formalismos innecesarios y frases vacías.
- **Paleta**: azul `#1a2b4a` (70%), turquesa `#00c9a7`, blanco `#ffffff`, crema `#F5F2EC`; complementarios naranja `#ff6b3d` y verde `#217638`.
- **Tipografías**: títulos **Bebas Neue** · texto general **Poppins** (normal/itálica) · énfasis emocional **Playfair Display**.
- **Logotipo — reglas del manual**: no deformar, no cambiar los colores, no girar, no agregar efectos.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de los visitantes puede ver la información general (identidad, «quiénes somos», horario de servicios, canales de WhatsApp, redes sociales y datos de contacto) sin autenticación, y la encuentra en menos de 30 segundos desde que abre la portada.
- **SC-002**: 0 contenidos en borrador visibles para el público en las pruebas de aceptación (incluida cualquier respuesta del sistema hacia el visitante).
- **SC-003**: Todo cambio guardado sobre un elemento publicado es visible en la portada pública desde la primera carga posterior al guardado (se publica al guardar), sin pasos adicionales.
- **SC-004**: Un miembro del equipo actualiza y publica un dato concreto (p. ej., un horario de servicio) en menos de 2 minutos desde el panel.
- **SC-005**: El 100% de las operaciones del módulo verifican el permiso «portada e información general»: 0 operaciones completadas sin permiso en las pruebas de aceptación y 100% denegadas con mensaje claro.
- **SC-006**: Con el idioma en inglés, el 100% de los textos de la interfaz está traducido y 0 campos aparecen vacíos por falta de traducción: cada contenido sin versión en inglés se muestra íntegramente en español.
- **SC-007**: La portada es usable en teléfono, tableta y computadora: 0 pérdida de secciones o funciones entre dispositivos y 0 desplazamiento horizontal en el ancho de pantalla más estrecho probado.
- **SC-008**: La portada supera una revisión de accesibilidad con 0 hallazgos críticos (referencia WCAG 2.1 nivel AA) y es recorrerla y operarla por completo con teclado y con lector de pantalla.
- **SC-009**: El 100% de los canales de WhatsApp y de las redes sociales publicados se abren con un solo clic o toque desde la portada y llevan al destino correcto.
- **SC-010**: En pruebas con personas de distintas edades, incluidas con poca experiencia en internet, al menos el 90% encuentra el horario de servicios y los canales de contacto sin ayuda.
- **SC-011**: Una persona del equipo sin el permiso del módulo no logra ver ni modificar la información general en el 100% de los intentos de las pruebas de aceptación.
- **SC-012**: 0 secciones vacías ni campos en blanco en la portada: toda sección sin elementos publicados se oculta por completo (decisión del humano del 2026-10-09).
- **SC-013**: El 100% de las ediciones de la información general de las pruebas de aceptación queda registrado en la auditoría de F2 (quién editó, qué y cuándo), y 0 registros de auditoría pueden editarse ni borrarse.

## Out of Scope

Queda explícitamente **fuera del alcance** de F3:

- **Las demás secciones del sitio público**: eventos y actividades (F4), grupos de conexión (F5), ministerios (F6), donaciones (F7), noticias y galería (F8) y medios (F9). La portada los alojará en futuras funcionalidades, pero F3 ni los muestra ni los administra.
- **Formularios de cualquier tipo** (contacto, inscripción, suscripción): la comunicación es solo mediante enlaces a WhatsApp y redes sociales y los datos de contacto publicados (FR-007); `idea.md` §6 deja los formularios de inscripción fuera del alcance.
- **Creación de banners o imágenes para redes sociales** (`idea.md` §6) y cualquier generación de material gráfico de marca.
- **Publicación automática en redes sociales ni envío de mensajes por WhatsApp**: el sitio solo enlaza hacia esos canales; las preguntas abiertas de `idea.md` §8 sobre automatización siguen fuera del MVP.
- **Traducción automática de contenidos**: el inglés se ingresa manualmente; si un contenido no tiene inglés, se muestra en español (FR-009).
- **Historial de versiones del contenido**: no hay versionado, ni recuperación de versiones anteriores, ni calendario de publicación programada en el MVP (default revisable, ver Assumptions).
- **Verificación de la vigencia de los enlaces**: el sistema no comprueba que un grupo de WhatsApp siga activo o que un perfil de red social siga existiendo; la corrección es manual editando el enlace.
- **Verificación detallada del cumplimiento del Manual de Identidad pieza a pieza** (paleta, tipografías, composición, cada uso del logotipo): la portada DEBE respetar el manual y sus reglas de logotipo (FR-002), pero la definición visual concreta corresponde a la fase de diseño UX y su verificación detallada a la implementación frontend. La spec no convierte los detalles del manual en requisitos de código.
- **Bilingüismo del panel de administración**: el panel se entrega en español (como F2); la Decisión 6 aplica al sitio público.
- **Gestión de usuarios, roles y permisos** (F2): F3 solo usa el permiso «portada e información general» ya reservado en el catálogo de F2; cualquier cambio en el modelo de permisos sería una ampliación.
- **Páginas de destino de canales y redes**: el sitio enlaza hacia WhatsApp y las redes; el contenido de esos destinos no se gestiona desde el sitio.
- **Despliegue a producción** y cambios de infraestructura: sigue pendiente a nivel de proyecto.

## Assumptions

Defaults razonables tomados donde la descripción no fija una regla; **son provisionales y revisables** por el humano. Los puntos que cambiaban el alcance fueron resueltos explícitamente en las aclaraciones del 2026-10-09 (ver *Aclaraciones resueltas*).

- **Memoria del idioma (decisión firme del humano del 2026-10-09)**: la elección del selector de idioma se mantiene al navegar por el sitio y **se recuerda entre visitas en el mismo dispositivo** (en la implementación se guarda en el propio dispositivo del visitante; está previsto `localStorage`). Un dispositivo sin ninguna preferencia de idioma guardada se muestra en **español** (idioma base, Decisión 8). No se analiza el idioma del navegador para proponer otro idioma (default revisable).
- **Nivel de accesibilidad**: se toma **WCAG 2.1 nivel AA** como referencia de accesibilidad (default revisable); el requisito de fondo es el de `idea.md` §7: apto para personas de todas las edades y dispositivos.
- **Idioma del panel**: se entrega en español, como el panel de F2; el bilingüismo del roadmap aplica al sitio público (Decisión 6).
- **La portada es única y es la página de inicio** del sitio público: la información general se administra desde un único lugar del panel; no existen varias "portadas" ni micrositios.
- **Traducción manual**: no hay traducción automática ni sugerencias automáticas de traducción; el inglés opcional lo ingresa el equipo.
- **Sin historial de versiones**: al editar se sobrescribe lo guardado; no hay versión anterior recuperable ni borrado diferido (default revisable).
- **Ediciones simultáneas**: si dos personas editan la información general al mismo tiempo, se conserva la última versión guardada completa (default revisable).
- **Validación de enlaces**: se valida el formato del enlace (que sea una dirección reconocible), no que el destino exista o siga activo.
- **Contenido textual**: admite los caracteres propios del español (tildes, ñ, ¿ ¡) y emojis sin transformación.
- **Límite de extensión de «quiénes somos»** (resolución del 2026-10-09: texto plano con límite): el límite concreto se fija en **1.000 caracteres** como default revisable (FR-003).
- **Datos de contacto obligatorios** (resolución del 2026-10-09: la portada incluye dirección, correo y teléfono): los tres son campos obligatorios de la sección de contacto (default revisable); el correo se valida con formato de correo y el teléfono con el criterio telefónico razonable de F2 (FR-015).
- **Catálogo de redes** (resolución del 2026-10-09: catálogo fijo con un enlace por red): el catálogo inicial es **Facebook, Instagram, YouTube, TikTok y Spotify**; incorporar otras redes equivalentes (p. ej., X) es un ajuste menor que se acuerda con el humano (default revisable).
- **Canales de WhatsApp duplicados**: un canal exactamente igual a otro (mismo destino y mismo nombre/propósito) se trata como duplicado y se rechaza (default revisable, coherente con la detección de duplicados de F2).
- **Contenido inicial de identidad**: los textos reales de la identidad (misión, visión, valores, «quiénes somos» base) provienen del Manual de Identidad y se respetan literalmente (ver *Identidad de marca — referencia para la fase de plan y UX*); el equipo los carga y completa desde el panel.
- **Manual de Identidad**: es la guía visual de la portada; la spec exige respetarlo y sus reglas de logotipo (FR-002), y deja la aplicación detallada de paleta, tipografías y composición a la fase de UX y a su verificación en frontend (ver Out of Scope).
- **Auditoría de F3** (resolución del 2026-10-09): las ediciones de la información general se registran en el registro de auditoría de F2 (quién, qué y cuándo), ampliando su alcance a este módulo con el mismo carácter de solo lectura; el detalle fino de cada cambio (valores antes/después) sigue siendo un default deseable pero revisable, igual que en F2.
- **Dependencias**: F2 terminada (acceso al panel, roles con permisos por módulo, permiso «portada e información general» ya reservado, patrón de autorización y registro de auditoría) y el Manual de Identidad entregado por el cliente (`resources/MANUAL DE MARCA.pdf`).

## Aclaraciones resueltas (2026-10-09)

**Ninguna aclaración pendiente.** Las diez preguntas abiertas de la primera versión se resolvieron con el humano el 2026-10-09 (quien decide: humano/cliente). Queda aquí el registro de la decisión tomada en cada una, más una decisión adicional del mismo día sobre la memoria del idioma (Q11).

### Q1 — Identidad de la iglesia y recursos gráficos (FR-002) — **Resuelta**

**Decisión (opción 1)**: el equipo edita desde el panel el **nombre oficial, el lema/misión/visión** y **sube el logotipo y la imagen de portada**. Además, el diseño del sitio DEBE seguir el **Manual de Identidad de la Iglesia Simiente Santa** (`resources/MANUAL DE MARCA.pdf`, logotipo `resources/simiente.jpeg`), cuyos datos de marca y contenido real se recogen literalmente en *Identidad de marca — referencia para la fase de plan y UX* (misión, visión, valores, «quiénes somos» base, personalidad y voz de marca, paleta, tipografías y reglas del logotipo). Los detalles de diseño/marca quedan como referencia para la fase de plan y UX: la spec exige respetar la identidad visual del manual y las reglas de uso del logotipo, sin convertirlos en requisitos de código detallados.

### Q2 — Formato de «quiénes somos» (FR-003) — **Resuelta**

**Decisión**: **texto plano** con límite de extensión; el límite concreto se fija en **1.000 caracteres** como valor revisable (ver Assumptions).

### Q3 — Estructura del horario de servicios (FR-004) — **Resuelta**

**Decisión**: **lista de varios servicios**; cada servicio tiene **día, hora, nombre/descripción y lugar**, y sus textos son traducibles de forma opcional al inglés.

### Q4 — Canales de WhatsApp (FR-005) — **Resuelta**

**Decisión**: **ambos tipos**: números para mensaje directo y enlaces de grupo. Puede haber **varios canales** y cada uno se identifica con **nombre/propósito** (y un único destino).

### Q5 — Redes sociales (FR-006) — **Resuelta**

**Decisión**: **catálogo fijo de redes** (Facebook, Instagram, YouTube, TikTok, Spotify u otras equivalentes) con **un enlace por red**.

### Q6 — Contacto adicional (FR-007) — **Resuelta**

**Decisión**: además de WhatsApp y redes sociales, la portada incluye **ubicación/dirección física, correo electrónico y teléfono**.

### Q7 — Granularidad de publicación (FR-013) — **Resuelta**

**Decisión**: **por secciones/elementos**: cada elemento de la portada se publica o retira por separado.

### Q8 — Edición de contenido publicado (FR-014) — **Resuelta**

**Decisión**: los cambios guardados sobre contenido publicado **se hacen visibles de inmediato** (se publica al guardar); no hay paso intermedio de borrador para lo ya publicado.

### Q9 — Auditoría de ediciones (FR-017) — **Resuelta**

**Decisión**: **sí**, la edición de la información general se registra en el registro de auditoría de F2 (quién editó, qué y cuándo), ampliando su alcance a este módulo.

### Q10 — Secciones sin contenido publicado (Edge Cases) — **Resuelta**

**Decisión**: **se oculta la sección vacía**; la portada se muestra coherente, sin secciones vacías ni campos en blanco.

### Q11 — Memoria de la elección de idioma (FR-010, US4) — **Resuelta (decisión adicional del 2026-10-09)**

**Decisión (humano, 2026-10-09)**: el sitio **recuerda la elección de idioma entre visitas en el mismo dispositivo** (persiste; en la implementación se guardará en `localStorage`). La **primera vez** que un dispositivo accede sin preferencia guardada, el sitio se muestra en **español** (idioma base, Decisión 8); a partir de una elección previa, esa preferencia se respeta en las visitas siguientes. El selector de idioma sigue visible y el cambio sigue aplicándose de inmediato y manteniéndose al navegar por el sitio.
