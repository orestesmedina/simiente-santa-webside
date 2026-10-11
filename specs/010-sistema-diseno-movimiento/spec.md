# Feature Specification: Sistema de diseño y movimiento (F10)

**Feature Branch**: `010-sistema-diseno-movimiento`

**Created**: 2026-10-10

**Status**: Draft — aclaraciones Q1–Q8 resueltas con el cliente el 2026-10-10; pendiente de aprobación de la spec

**Input**: User description: "F10 — Sistema de diseño y movimiento (roadmap §3, `docs/producto/roadmap.md`): el sitio público deja de verse «de 2010»: sistema visual moderno (tendencias 2026) construido sobre el Manual de Identidad, con animaciones y microinteracciones agradables (respetando `prefers-reduced-motion`), y el rediseño de la navegación (menú sin scroll incómodo) y del selector de idioma (sin ocupar espacio innecesario). Es transversal: se aplica a la portada (F3) y a las secciones F4–F9." **Decisión 11 (2026-10-10)**: el cliente revisó la portada de F3, la encontró poco moderna (pide tendencias 2026 y animaciones agradables), con el menú con scroll incómodo y el selector de idioma ocupando espacio; se decidió construir el rediseño como **funcionalidad propia y transversal (F10)** en vez de reabrir F3: **F3 se mantiene como base funcional** y F10 aplica el sistema visual y de movimiento a la portada **y** a las secciones F4–F9. Restricciones del encargo (`idea.md` intro y §7): sitio **moderno y limpio**, «que no sea un diseño típico y común de IA», apto para **personas de todas las edades** (incluidas mayores con lo básico de internet), **sin menús ocultos ni gestos desconocidos**, con animaciones modernas, adaptable a cualquier dispositivo. **Identidad de marca**: el diseño se construye sobre el **Manual de Identidad** (`resources/MANUAL DE MARCA.pdf`, logotipo `resources/simiente.jpeg`) y sobre los **design tokens ya definidos en F3** (`specs/003-portada-info-general/ux.md` §1: paleta `#1a2b4a`/`#00c9a7`/`#F5F2EC`/`#ff6b3d`/`#217638`, tipografías Bebas Neue / Poppins / Playfair Display); F10 los aplica y los consolida, no los reinventa. **Dependencia**: F3 (terminada) provee la portada pública funcional (contenido, bilingüismo, publicación borrador/publicado, accesibilidad) y F2 (terminada) provee el panel de administración, cuyo comportamiento NO cambia. Dos decisiones de UX de F3 quedan marcadas para revisión en esta funcionalidad: la **navegación** (D-5: enlaces de ancla en fila desplazable, sin menú hamburguesa) y el **selector de idioma** (D-1: radios nativos con ambos idiomas escritos completos). **Aclaraciones resueltas con el cliente el 2026-10-10 (Q1–Q8)**: el alcance **incluye también el rediseño visual del panel de administración y el modo oscuro** del dispositivo (sitio y panel); el «scroll incómodo» del menú cubría **dos** casos (desplazamiento lateral de los enlaces en móvil y pérdida de la navegación al bajar) y se resuelve con **barra de navegación fija siempre visible**; el nivel de animación es **moderado**; el selector de idioma **solo se compacta** (sin píldoras ni menú); se permite **crear iconografía e ilustraciones nuevas** respetando la marca; y antes de implementar se presentan **3 propuestas de diseño distintas** de las que el humano aprueba una (referencias de inspiración del cliente: `https://iglesiaoasis.com/` y `https://www.river-arena.org/`, no para copiar). Detalle en *Aclaraciones resueltas (2026-10-10)*.

> **Nota de alcance y terminología**: F10 es una funcionalidad **de presentación, movimiento, navegación, formato del selector de idioma y tema (claro/oscuro)**. No cambia qué información muestra el sitio, cómo se publica ni quién puede editarla: eso sigue siendo F3, y el panel de administración conserva íntegro su comportamiento de F2 y F3. El **visitante** y el **usuario del panel** son los mismos de F3. F10 entrega tres cosas: (1) el **rediseño aplicado** a lo que hoy existe —la portada de F3 **y el panel de administración completo** (decisión Q1 del 2026-10-10)—, (2) los **cimientos reutilizables** del sistema (visual, de movimiento, de navegación y del selector) para que las secciones futuras F4–F9 (eventos, grupos de conexión, ministerios, donaciones, noticias/galería, medios) nazcan ya con él, y (3) el **modo oscuro** del sitio y del panel, acoplado a la preferencia del dispositivo (decisión Q7). F10 **no construye** las secciones F4–F9. Antes de implementar, la fase de diseño DEBE presentar **3 propuestas visuales distintas** y el humano aprueba una (FR-027). Las 8 preguntas que cambiaban alcance o experiencia quedaron **resueltas con el cliente el 2026-10-10** (ver *Aclaraciones resueltas (2026-10-10)*); no queda ninguna `[NECESITA ACLARACIÓN]` abierta.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Ver un sitio público moderno y cuidado, fiel a la marca (Priority: P1)

Como visitante (miembro de la iglesia o persona nueva), quiero que el sitio público se vea actual y bien diseñado —tipografía, color, espaciado y composición agradables y fieles a la identidad de la iglesia— para percibir una iglesia cercana y de confianza, y no un sitio anticuado ni una plantilla genérica.

**Why this priority**: Es el corazón de F10 y la razón de la Decisión 11: el cliente vio la portada «poco moderna» y pidió un sitio de tendencias 2026 que además «no sea un diseño típico y común de IA» (`idea.md` intro). La primera impresión de toda la presencia de la iglesia depende de esta historia: sin ella, el resto (movimiento, navegación, selector) se aplica sobre un aspecto que el cliente ya rechazó.

**Independent Test**: Se prueba de forma independiente con la portada publicada: se recorre visualmente cada sección verificando que aplica el sistema (tipografías de marca, paleta con su proporción, espaciado, jerarquía, componentes), que respeta las reglas del logotipo y el contraste, y que el conjunto se ve coherente y actual en teléfono, tableta y computadora; se compara con el diseño previo y la diferencia es evidente. Entrega valor por sí sola: el sitio ya no se ve «de 2010».

**Acceptance Scenarios**:

1. **Dado** el sitio público rediseñado, **cuando** un visitante abre la portada en cualquier dispositivo, **entonces** todas sus secciones (cabecera, identidad/portada, «quiénes somos», horario de servicios, WhatsApp, contacto, redes sociales y pie) se presentan con el mismo sistema visual —tipografías, color, espaciado y jerarquía— y se perciben como un conjunto coherente y actual.
2. **Dado** el rediseño, **cuando** se revisa frente a la identidad de la iglesia, **entonces** se reconocen la paleta, las tipografías y el logotipo del Manual de Identidad, y el logotipo se muestra sin deformar, sin cambiar de colores, sin girar y sin efectos.
3. **Dado** contenidos reales de distinta forma (textos largos, un servicio suelto en el horario, contenido solo en español navegando en inglés, varias redes publicadas), **cuando** el visitante los ve, **entonces** el diseño se mantiene legible y cuidado, sin huecos, sin composiciones rotas ni textos apretados o desbordados.
4. **Dado** un teléfono con pantalla de 320 px de ancho o un navegador con el texto ampliado al 200 %, **cuando** el visitante usa el sitio, **entonces** la composición conserva su jerarquía y legibilidad, sin desbordes ni pérdida de información.
5. **Dado** cualquier elemento interactivo (enlace, botón, selector de idioma), **cuando** el visitante pasa el cursor, lo enfoca con teclado o lo activa, **entonces** se percibe claramente como interactivo y el contraste sigue cumpliendo WCAG 2.1 AA en todos esos estados.

---

### User Story 2 - Navegar por las secciones sin esfuerzo ni scroll incómodo (Priority: P1)

Como visitante de cualquier edad y nivel de experiencia con internet, quiero una navegación clara y cómoda hacia las secciones del sitio, para llegar a lo que busco sin pelearme con el menú ni con desplazamientos incómodos.

**Why this priority**: Es una de las dos quejas nombradas del cliente (Decisión 11: «menú con scroll incómodo») y choca con una restricción dura del encargo: apto para personas de todas las edades, sin menús ocultos ni gestos desconocidos (`idea.md` intro y §7). Sin navegación cómoda, el resto del sitio no sirve a su público —especialmente a las personas mayores con lo básico de internet—.

**Independent Test**: Se prueba de forma independiente sobre la portada publicada: en teléfono, tableta y computadora se comprueba que todos los enlaces de la navegación están a la vista y llevan a su sección, sin desplazamiento lateral de los enlaces ni desplazamiento horizontal de la página; que la navegación sigue visible y accesible al bajar (barra fija); se recorre con teclado y lector de pantalla; y se prueba con personas de poca experiencia para ver si encuentran y usan la navegación sin ayuda. Entrega valor por sí sola: cualquiera llega a cualquier sección sin esfuerzo.

**Acceptance Scenarios**:

1. **Dado** el sitio público, **cuando** un visitante abre la portada, **entonces** ve una navegación clara con los enlaces a las secciones visibles a primera vista, sin menús ocultos ni gestos desconocidos.
2. **Dado** un teléfono con pantalla estrecha, **cuando** el visitante usa la navegación, **entonces** llega a cualquier enlace sin tener que desplazar los enlaces de lado (caso (a) del «scroll incómodo», aclaración del 2026-10-10) y la página nunca provoca desplazamiento horizontal.
3. **Dado** una persona con poca experiencia en internet, **cuando** busca una sección concreta (p. ej., el horario de servicios), **entonces** la encuentra con títulos claros y la activa con un solo clic o toque, sin instrucciones adicionales ni prueba y error.
4. **Dado** una persona que usa teclado o lector de pantalla, **cuando** recorre la navegación, **entonces** alcanza y activa cada enlace en un orden comprensible, siempre con foco visible, y cada enlace indica claramente a qué sección lleva.
5. **Dado** el patrón de navegación rediseñado, **cuando** se prueba con la cantidad prevista de secciones del sitio completo (portada + F4–F9), **entonces** sigue cumpliendo los escenarios anteriores sin volver al problema del scroll, sin menús hamburguesa ni desplegables ocultos.
6. **Dado** el visitante que baja por la portada, **cuando** se desplaza hacia abajo, **entonces** la navegación permanece visible y accesible en todo momento (barra fija, caso (b) del «scroll incómodo», aclaración del 2026-10-10), sin tapar el contenido ni descolocar el maquetado.

---

### User Story 3 - Recorrer el sitio con movimiento agradable y sin molestias (Priority: P2)

Como visitante, quiero que el sitio responda de forma viva y agradable —microinteracciones al tocar, transiciones suaves, contenido que aparece con calma— para que navegar sea placentero; y como persona sensible al movimiento (o que prefiere menos movimiento), quiero que el sitio respete mi preferencia del sistema, para usarlo sin mareos ni distracciones.

**Why this priority**: Las animaciones y microinteracciones agradables son petición explícita del cliente (Decisión 11, `idea.md` intro) y mitad del nombre de la funcionalidad. Va como P2 porque un sitio moderno y cómodo de navegar ya es un MVP valioso con US1 + US2, y el movimiento lo enriquece sin ser condición de uso: la información y las funciones nunca pueden depender de ver una animación. El nivel fijado con el cliente el 2026-10-10 es **moderado**: microinteracciones en hover/foco/clic y revelados suaves y breves, sin llegar a un estilo expresivo.

**Independent Test**: Se prueba de forma independiente en dos configuraciones: con movimiento normal (se observan microinteracciones y transiciones agradables de nivel moderado que no estorban la lectura ni la interacción) y con la preferencia de movimiento reducido del dispositivo activa (no hay animaciones de entrada, revelado ni desplazamiento; el contenido aparece completo de inmediato y todo funciona igual). Además se comprueba en un dispositivo modesto que las animaciones no retrasan la lectura ni bloquean las acciones. Entrega valor por sí sola: navegar el sitio resulta agradable y también cómodo para quien prefiere menos movimiento.

**Acceptance Scenarios**:

1. **Dado** el sitio rediseñado, **cuando** el visitante pasa el cursor, enfoca con teclado o toca un elemento interactivo, **entonces** recibe una retroalimentación clara y agradable, sin saltos ni desplazamientos del maquetado.
2. **Dado** el visitante que baja por la página o cambia de sección, **cuando** el contenido entra en pantalla, **entonces** aparece con una transición suave y breve que no retrasa la lectura (nivel moderado: sin secuencias llamativas ni largas).
3. **Dado** un dispositivo con la preferencia de movimiento reducido activa (`prefers-reduced-motion`), **cuando** el visitante usa el sitio, **entonces** no hay animaciones de entrada, revelado ni desplazamiento: el contenido se muestra completo de inmediato y el 100 % de las funciones sigue disponible.
4. **Dado** cualquier animación, **cuando** se reproduce, **entonces** nunca oculta información necesaria ni bloquea una acción del visitante; el contenido útil es visible e interactivo desde el primer instante.
5. **Dado** el visitante que cambia de idioma o vuelve a una sección ya vista, **cuando** ocurre, **entonces** el movimiento no se repite de forma molesta (sin «rebotar» ni parpadeos) y la página no se descoloca.
6. **Dado** un dispositivo modesto o con batería baja, **cuando** el visitante navega, **entonces** la experiencia sigue fluida, sin tirones perceptibles ni esperas para interactuar.

---

### User Story 4 - Cambiar de idioma con un selector compacto y claro (Priority: P2)

Como visitante, quiero que el selector de idioma —los mismos textos «Español» / «English»— ocupe menos espacio sin perder claridad, para que la cabecera respire y el sitio se vea más cuidado, sin perder la facilidad de cambiar de idioma.

**Why this priority**: Es la segunda queja nombrada del cliente (Decisión 11: «el selector de idioma ocupando espacio»). Va como P2 porque el comportamiento del selector ya funciona y está probado en F3; aquí solo se **compacta** (decisión Q5 del 2026-10-10: reducir tamaño y espacio de los textos actuales, sin píldoras ni menú), sin tocar el comportamiento.

**Independent Test**: Se prueba de forma independiente con el selector rediseñado: se compara con el actual y ocupa menos espacio conservando los textos completos de cada idioma; personas de poca experiencia lo identifican como selector de idioma; cambia el idioma de inmediato, la elección se mantiene al navegar y se recuerda entre visitas; un dispositivo sin preferencia guardada arranca en español; y se opera por completo con teclado y lector de pantalla. Entrega valor por sí sola: la cabecera respira y el cambio de idioma sigue al alcance de un toque.

**Acceptance Scenarios**:

1. **Dado** el sitio rediseñado, **cuando** el visitante mira la cabecera, **entonces** el selector de idioma ocupa menos espacio que el actual conservando los textos completos «Español» y «English» (sin píldoras «ES / EN» ni menú desplegable), y sigue siendo claramente identificable como selector de idioma, tanto para personas de poca experiencia como para lectores de pantalla.
2. **Dado** el selector, **cuando** el visitante elige el otro idioma, **entonces** la página cambia de inmediato, la elección se mantiene al navegar por el resto del sitio y se recuerda entre visitas en el mismo dispositivo (comportamiento de F3, FR-010 de F3, sin cambios).
3. **Dado** un dispositivo sin ninguna preferencia de idioma guardada (primera visita), **cuando** abre el sitio, **entonces** se muestra en español, que es el idioma base, igual que en F3.
4. **Dado** el selector compacto, **cuando** se usa con teclado o lector de pantalla, **entonces** cada opción está claramente identificada —el nombre completo del idioma es accesible, no solo siglas visuales— y se activa sin esfuerzo, con foco visible.
5. **Dado** el selector en la cabecera y, si se conserva, en el pie, **cuando** el visitante lo usa en cualquiera de los dos, **entonces** el comportamiento es idéntico en ambos y no provoca desplazamientos ni tapa contenido en ningún ancho de pantalla.

---

### User Story 5 - Usar el panel de administración con el mismo diseño moderno (Priority: P2)

Como usuario del panel (equipo de la iglesia), quiero que el panel de administración tenga el mismo sistema visual y el mismo movimiento que el sitio público —login, inicio, usuarios, roles, auditoría, gestión de la portada y pantallas de apoyo— para trabajar con una herramienta clara, actual y agradable, sin perder ninguna función.

**Why this priority**: La Decisión Q1 del 2026-10-10 amplió el alcance de F10 al panel de administración: el equipo trabaja ahí a diario y hoy tiene un aspecto distinto y más frío que el del sitio público. Va como P2 porque el panel ya es funcional y usable (F2, F3): el rediseño mejora la experiencia de trabajo sin ser condición para el público, y debe hacerse sin tocar el comportamiento para no poner en riesgo lo ya probado.

**Independent Test**: Se prueba de forma independiente recorriendo todas las pantallas del panel con la cuenta de un administrador y comparándolas con el sistema visual definido; se repite el recorrido con una cuenta con permisos limitados; se ejecutan las pruebas de F2 y F3 completas y todas pasan sin cambios; y se opera el panel con teclado y lector de pantalla. Entrega valor por sí sola: el equipo trabaja con una herramienta moderna y coherente con su marca.

**Acceptance Scenarios**:

1. **Dado** el panel rediseñado, **cuando** el usuario del panel inicia sesión y recorre sus pantallas (inicio de sesión, inicio del panel, usuarios, roles, auditoría, portada e información, cambiar contraseña y pantalla de «sin permiso»), **entonces** todas aplican el mismo sistema visual —tipografías, color, espaciado, jerarquía, componentes— y el mismo movimiento moderado del sitio público, como un conjunto coherente.
2. **Dado** el rediseño del panel, **cuando** el usuario realiza cualquier operación existente (crear o desactivar un usuario, crear un rol y sus permisos, consultar la auditoría, editar la información general, publicar o retirar un elemento), **entonces** el comportamiento es idéntico al actual y el 100 % de las pruebas de F2 y de F3 sigue pasando sin cambios.
3. **Dado** el panel con pantallas densas (tablas de usuarios y auditoría, formularios de edición, listados con estados), **cuando** el equipo trabaja con ellas, **entonces** siguen siendo claras y legibles, sin perder información, estados ni acciones.
4. **Dado** el panel rediseñado, **cuando** se usa con teclado o lector de pantalla, **entonces** se puede recorrer y operar por completo, con foco visible y contraste AA, igual que antes del rediseño.
5. **Dado** el panel en su idioma único (español), **cuando** se rediseña, **entonces** sigue íntegramente en español y con el tono cercano ya definido; el rediseño no introduce textos en otro idioma ni cambia el idioma del panel.

---

### User Story 6 - Ver el sitio y el panel en modo oscuro (Priority: P2)

Como visitante o usuario del panel que usa el modo oscuro de su dispositivo, quiero que el sitio público y el panel se adapten a ese modo, para leer con comodidad —por la noche o en ambientes oscuros— sin deslumbramientos y con la misma calidad de diseño.

**Why this priority**: La Decisión Q7 del 2026-10-10 incluyó el modo oscuro en F10: es una expectativa habitual en el diseño actual y mejora la comodidad de lectura. Va como P2 porque el sitio y el panel son plenamente funcionales en tema claro; el tema oscuro es una capa de adaptación sobre el sistema visual ya rediseñado (US1 y US5) y se puede verificar por separado.

**Independent Test**: Se prueba de forma independiente poniendo el dispositivo en modo oscuro: el sitio público y todas las pantallas del panel se presentan en tema oscuro, con contraste AA verificado, el logotipo con sus colores originales y las mismas funciones que en tema claro; se pone el dispositivo en modo claro y todo vuelve sin pérdida; se repite la prueba con movimiento reducido y con teclado/lector de pantalla. Entrega valor por sí sola: quien prefiere el modo oscuro usa el sitio y el panel con total comodidad.

**Acceptance Scenarios**:

1. **Dado** un dispositivo con modo oscuro activo, **cuando** el visitante abre el sitio público, **entonces** se muestra en tema oscuro, legible y fiel a la marca, con el mismo contenido y las mismas funciones que en tema claro.
2. **Dado** un dispositivo con modo oscuro activo, **cuando** el usuario del panel inicia sesión y trabaja en cualquier pantalla, **entonces** el panel se muestra en tema oscuro sin perder ninguna función, estado ni acción.
3. **Dado** cualquier tema (claro u oscuro), **cuando** se revisa el contraste de textos, botones, estados interactivos y foco, **entonces** cumple WCAG 2.1 nivel AA en ambos temas.
4. **Dado** el logotipo y las imágenes de marca en tema oscuro, **cuando** se muestran, **entonces** conservan sus colores originales y sus reglas de uso (sin deformar, sin recolorear, sin girar, sin efectos), sobre superficies que los hagan legibles.
5. **Dado** el dispositivo que cambia de tema con el sitio o el panel abiertos, **cuando** cambia, **entonces** la interfaz lo respeta sin necesidad de recargar (default revisable) y sin parpadeos molestos.
6. **Dado** el modo oscuro combinado con la preferencia de movimiento reducido, **cuando** se usan a la vez, **entonces** todo sigue funcionando igual: ambos requisitos se cumplen sin conflicto.

---

### User Story 7 - Un sitio coherente hoy y listo para las secciones futuras (Priority: P3)

Como visitante (hoy, y cuando lleguen las secciones futuras de eventos, grupos de conexión, ministerios, donaciones, noticias y medios), quiero que todo el sitio público se sienta una sola experiencia —mismo estilo, mismo movimiento, misma navegación— para reconocerlo y usarlo sin aprender nada nuevo.

**Why this priority**: La transversalidad es parte explícita del encargo (Decisión 11: F10 aplica a la portada y a F4–F9) y es la razón de construir F10 **antes** que F4 según el roadmap: que las secciones posteriores nazcan ya con el sistema y no haya que rediseñarlas después. Va como P3 porque hoy solo existe la portada: la experiencia «coherente entre secciones» se puede mostrar primero en la propia portada y en los cimientos que F10 deja listos.

**Independent Test**: Se prueba de forma independiente verificando que toda la portada usa los componentes y reglas del sistema (sin dos estilos distintos para la misma función) y que los cimientos del sistema quedan entregados y documentados para que una sección futura los aplique sin rediseño (una sección nueva construida con ellos debe encajar sin tocar el sistema). Entrega valor por sí sola: el sitio actual ya es coherente y las secciones futuras saldrán más rápido y sin rediseños.

**Acceptance Scenarios**:

1. **Dado** el sitio público actual, **cuando** el visitante recorre la portada, **entonces** todas sus secciones aplican los mismos componentes y reglas del sistema (no hay dos estilos distintos para la misma función, p. ej., dos tipos de botón de acción o dos jerarquías de título de sección).
2. **Dado** los cimientos del sistema (tipografía, color con sus temas claro y oscuro, espaciado, componentes base, patrones de movimiento, navegación y selector de idioma), **cuando** el equipo diseña una sección futura (F4–F9), **entonces** puede aplicarlos sin rediseñar el sistema ni inventar estilos base nuevos.
3. **Dado** el rediseño, **cuando** en el futuro llega una sección nueva, **entonces** el visitante la percibe como parte del mismo sitio (mismo aspecto, mismo movimiento, misma navegación), sin diferencias llamativas respecto a la portada.
4. **Dado** el cambio de idioma, **cuando** se aplica, **entonces** los textos de interfaz nuevos que introduce el rediseño están disponibles en español e inglés, sin textos sin traducir.

---

### Edge Cases

- **Preferencia de movimiento reducido activa** (`prefers-reduced-motion`): sin animaciones de entrada, revelado ni desplazamiento; el contenido aparece completo de inmediato; el 100 % de las funciones y la información sigue disponible; nada depende de haber visto una animación (US3, FR-006).
- **La preferencia de movimiento cambia con el sitio abierto** (el visitante activa «reducir movimiento» en su dispositivo): el sitio la respeta sin necesidad de recargar (default revisable, ver Assumptions).
- **Animación o elemento visual que falla al cargar**: el contenido aparece igualmente; nunca una zona en blanco, un texto oculto ni un elemento inoperable.
- **Dispositivo muy modesto, con batería baja o con ahorro de datos**: el contenido es legible e interactivo desde el primer instante; las animaciones nunca bloquean ni retrasan una acción; si el movimiento degrada la experiencia, se reduce o se elimina antes que entorpecer la lectura (FR-007).
- **Foco de teclado sobre un elemento que aún no se ha «revelado»** (p. ej., contenido con transición de entrada al bajar): al recibir foco, el elemento DEBE hacerse visible de inmediato; el foco nunca debe quedar sobre algo oculto o transparente por una animación pendiente.
- **Barra de navegación fija** (FR-014): no tapa el destino de un ancla al saltar a una sección (el título de la sección queda visible), no tapa contenido con texto ampliado al 200 % ni en pantallas de 320 px, y no desplaza el contenido de forma que se recorte.
- **Zoom al 200 % y texto ampliado**: la composición aguanta sin pérdida de información ni desplazamiento horizontal (hereda FR-018 de F3).
- **Pantallas de 320 px de ancho**: navegación, selector y todas las secciones siguen siendo usables, sin desbordes (SC-007 de F3 sigue aplicando).
- **Navegación con muchas secciones** (previsión F4–F9, ~8–10 enlaces): todos los enlaces siguen alcanzables sin scroll incómodo, sin menús ocultos ni desplazamiento horizontal de la página (US2 esc. 5, FR-012).
- **Cambio de idioma con el diseño nuevo**: el texto en inglés y el texto de fallback (en español) ocupan longitudes distintas; el cambio no debe descolocar la composición ni repetir animaciones de entrada de forma molesta (US3 esc. 5).
- **Sección que desaparece por completo** (todo su contenido en borrador, comportamiento de F3 FR-013): la navegación actualiza sus enlaces sin dejar enlaces rotos, enlaces huérfanos ni animaciones apuntando a algo que ya no existe.
- **Modo oscuro activo en el dispositivo** (o cambio de tema en caliente): el sitio y el panel se presentan en el tema del dispositivo con contraste AA y las mismas funciones y contenido; el cambio de tema con la aplicación abierta se respeta sin recargar (default revisable) y sin parpadeos molestos (US6, FR-025, FR-026).
- **Logotipo e imágenes en modo oscuro**: el logotipo conserva SIEMPRE sus colores originales —nunca se recolorea ni se invierte para adaptarlo al tema oscuro— y se muestra sobre una superficie que lo haga legible; las imágenes se adaptan sin deformarse ni perder información (FR-026).
- **Tablas y formularios densos del panel en modo oscuro** (usuarios, roles, auditoría, gestión de la portada): siguen legibles, con estados claros (borrador/publicado, éxito/error, activo/inactivo) que no dependan solo del color (FR-026).
- **Logotipo en estados interactivos** (hover, foco, activo): las microinteracciones NUNCA deforman, recolorean, giran ni añaden efectos al logotipo (reglas del Manual de Identidad, FR-002 de F3).
- **Carga con caché o recarga**: sin parpadeos que hagan parecer el sitio roto (estilos o tipografías que aparecen a destiempo); el visitante no debe ver el «antes y después» del rediseño montándose.
- **Lector de pantalla y movimiento**: los cambios visuales por animación no deben generar anuncios molestos ni repetidos; el contenido es accesible desde el inicio, aunque su revelado visual sea gradual.

### Errores esperados

F10 no introduce nuevas operaciones que fallen, pero sí situaciones degradadas que DEBEN resolverse sin romper la experiencia. Todos los casos conservan la información y las funciones completas.

| Situación | Qué debe ver la persona |
|---|---|
| Visitante con `prefers-reduced-motion` activo (o «reducir movimiento» en el dispositivo) | Sitio completo con el mismo contenido y las mismas funciones, sin animaciones de entrada, revelado ni desplazamiento |
| Animación o elemento visual que no carga | El contenido aparece igualmente; nunca una zona en blanco ni un elemento inoperable |
| Dispositivo lento o con batería baja | Contenido legible e interactivo desde el primer instante; las animaciones no bloquean ni retrasan las acciones |
| Navegación con más secciones de las actuales (previsión F4–F9) | Todos los enlaces alcanzables sin scroll incómodo ni desplazamiento horizontal de la página |
| Teclado o lector de pantalla con el diseño nuevo (sitio y panel) | Foco visible siempre, orden comprensible y todo operable; las animaciones nunca ocultan el foco ni el contenido |
| Cambio de idioma con el selector compacto | Comportamiento idéntico a F3: cambio inmediato, se mantiene al navegar, se recuerda entre visitas; primera visita en español |
| Texto ampliado al 200 % o pantalla de 320 px | Composición legible sin desbordes, sin pérdida de información ni acciones inaccesibles; la barra fija no tapa contenido |
| Contenido en español mostrado navegando en inglés (fallback de F3) | Se integra en el nuevo diseño sin huecos ni descolocaciones, con el mismo aspecto que el resto |
| Dispositivo con modo oscuro (o cambio de tema en caliente) | Sitio y panel adaptados al tema del dispositivo, con contraste AA y el mismo contenido y funciones que en tema claro |
| Logotipo o imágenes en modo oscuro | Logotipo con sus colores originales sobre una superficie legible, nunca recoloreado; imágenes sin deformar |

## Requirements *(mandatory)*

### Functional Requirements

**Sistema visual moderno (sobre el Manual de Identidad y los tokens de F3)**

- **FR-001**: El sitio público DEBE aplicar un sistema visual moderno (referencia de tendencias 2026) construido sobre el **Manual de Identidad** y los design tokens ya definidos en F3 (paleta, tipografías Bebas Neue / Poppins / Playfair Display, espaciado, radios), de modo que el sitio se perciba actual —no «de 2010»— y distinto de una plantilla genérica, sin contradecir la marca. Si una tendencia entra en conflicto con el Manual de Identidad, DEBE prevalecer el Manual. El mismo sistema se aplica al panel según FR-023.
- **FR-002**: El sistema visual DEBE respetar las reglas del logotipo del Manual (no deformar, no cambiar los colores, no girar, no agregar efectos) en todas sus aplicaciones, incluidos los estados interactivos, y DEBE conservar la proporción de color de la marca (70 % azul marino, 20 % blanco/crema, 10 % turquesa y acentos); los colores derivados solo se admiten para garantizar contraste y legibilidad (como ya se hizo en F3). **SÍ se permite crear iconografía e ilustraciones nuevas** para el rediseño (decisión del cliente del 2026-10-10), siempre que respeten la paleta, las tipografías y las reglas del Manual y del logotipo; el logotipo y el Manual NO se modifican.
- **FR-003**: El sistema DEBE definir y usar componentes base reutilizables —tipografía, color, espaciado, jerarquía de títulos, botones y acciones, tarjetas, encabezados de sección, estados de foco/hover/activo, imágenes y pie— aplicados de forma consistente en todas las secciones del sitio público y en el panel (FR-023); NUNCA DEBE haber dos estilos distintos para la misma función.
- **FR-004**: El sistema visual DEBE conservar la legibilidad para todas las edades heredada de F3 (cuerpo de texto suficientemente grande, contraste AA, títulos claros): la modernidad NO DEBE lograrse a costa de la legibilidad ni de la claridad.

**Movimiento y microinteracciones**

- **FR-005**: El sitio público DEBE incorporar animaciones y microinteracciones agradables con **nivel moderado** (decisión del cliente del 2026-10-10): retroalimentación al pasar el cursor, enfocar y activar (hover/foco/clic), transiciones suaves entre estados y revelado suave y breve del contenido al navegar; sin llegar a un estilo expresivo, sin secuencias largas ni llamativas — carácter cercano y calmado, acorde a la marca, nada estridente ni infantil.
- **FR-006**: Todo movimiento DEBE respetar la preferencia de movimiento reducido del sistema (`prefers-reduced-motion`): cuando esté activa, las animaciones de entrada, revelado y desplazamiento DEBEN desactivarse por completo (o reducirse a cambios instantáneos), el contenido DEBE mostrarse completo de inmediato y el 100 % de las funciones DEBE seguir disponible. Este requisito es incondicional.
- **FR-007**: El movimiento NO DEBE degradar la carga ni la fluidez: el contenido útil DEBE ser visible e interactivo desde el primer instante (ninguna animación bloquea la lectura ni retrasa una acción más de lo mínimo perceptible) y la navegación DEBE mantenerse fluida incluso en dispositivos modestos.
- **FR-008**: Cada elemento interactivo DEBE dar retroalimentación clara en hover, foco y activación, sin desplazar el maquetado al cambiar de estado; el movimiento NO DEBE repetirse de forma molesta (p. ej., al cambiar de idioma o al volver a una sección ya vista) ni ocultar información necesaria en ningún momento.
- **FR-009**: La información y la funcionalidad NUNCA DEBEN depender de ver una animación: todo contenido y toda acción DEBEN estar disponibles también con el movimiento desactivado, con animaciones fallidas o en dispositivos que las limiten.

**Navegación (rediseño)**

- **FR-010**: El sitio público DEBE ofrecer una navegación clara y siempre visible hacia las secciones del contenido, apta para personas de todas las edades y niveles de experiencia: sin menús ocultos ni desplegables ocultos, sin gestos desconocidos, con títulos claros y enlaces activables con un solo clic o toque (restricción de `idea.md` intro y §7).
- **FR-011**: La navegación DEBE eliminar por completo el «scroll incómodo» del menú reportado por el cliente, que la aclaración del 2026-10-10 define como **dos casos**: (a) los enlaces se desplazan lateralmente en la cabecera en móvil —ningún enlace DEBE requerir desplazamiento lateral del menú para ser alcanzado— y (b) la navegación se pierde al bajar la página —DEBE seguir accesible mientras se recorre el sitio (FR-014)—. Además, la página NUNCA DEBE provocar desplazamiento horizontal en ningún ancho probado (mínimo 320 px).
- **FR-012**: El patrón de navegación DEBE quedar preparado para alojar las secciones futuras (F4–F9): con la cantidad prevista de secciones del sitio completo DEBE seguir cumpliendo FR-010 y FR-011, sin menús hamburguesa ni desplegables ocultos. (Default revisable sobre cómo acomodar muchos enlaces: siempre visibles, en una o varias filas compactas; ver Assumptions.)
- **FR-013**: La navegación DEBE ser accesible: operable por completo con teclado y lector de pantalla, en un orden comprensible, con foco visible en todo momento y con enlaces reales que indican claramente a qué sección llevan (hereda FR-019 de F3).
- **FR-014**: La navegación del sitio público DEBE ser una **barra fija siempre visible** (decisión del cliente del 2026-10-10): acompaña al visitante mientras se desplaza por la página, mantiene todos los enlaces alcanzables en todo momento (incluida la previsión de F4–F9) y NUNCA tapa el contenido ni descoloca el maquetado: el destino de un ancla, los títulos de sección y el texto ampliado DEBEN quedar siempre visibles.

**Selector de idioma (rediseño)**

- **FR-015**: El selector de idioma DEBE ocupar menos espacio que el actual **compactando los textos actuales** («Español» / «English») mediante una reducción de tamaño y espacio (decisión del cliente del 2026-10-10): **SIN píldoras «ES / EN», SIN menú desplegable y SIN siglas como identificación principal**. El nombre completo de cada idioma DEBE seguir siendo claro y accesible (para personas de poca experiencia con internet y para lectores de pantalla) y el selector DEBE poder usarse con un solo clic o toque, sin menús ocultos ni gestos desconocidos.
- **FR-016**: El comportamiento del selector DEBE ser idéntico al definido en F3 (FR-010 de F3): el cambio se aplica de inmediato a la página visible, se mantiene al navegar por el resto del sitio, se recuerda entre visitas en el mismo dispositivo y la primera visita sin preferencia guardada se muestra en español (idioma base); sin auto-detección del idioma del navegador. Solo cambia su **formato y huella visual**, nunca su comportamiento.
- **FR-017**: El selector, allá donde se muestre (cabecera y pie en F3), NO DEBE provocar desplazamiento horizontal, tapar contenido ni descolocar el maquetado en ningún ancho de pantalla ni en ningún estado (reposo, foco, activo, idioma seleccionado).

**Transversalidad (aplicación y cimientos)**

- **FR-018**: El sistema visual, el movimiento, la navegación y el selector DEBEN aplicarse a la portada completa de F3 (todas sus secciones: cabecera, identidad/portada, «quiénes somos», horario de servicios, WhatsApp, contacto, redes sociales y pie) de forma homogénea, y DEBEN quedar definidos como cimientos reutilizables y documentados para que las secciones F4–F9 los apliquen sin rediseño. F10 NO construye las secciones F4–F9. El panel se rediseña con el mismo sistema según FR-023.
- **FR-019**: Todo texto de interfaz que introduzca el rediseño **del sitio público** DEBE estar disponible en español e inglés (hereda FR-008 y SC-006 de F3): cero textos de interfaz sin traducir en ninguno de los dos idiomas. El panel sigue íntegramente en español (FR-016 de F3, FR-024).

**No regresión funcional, datos y accesibilidad**

- **FR-020**: F10 NO DEBE cambiar el comportamiento funcional de F3: el contenido y su edición desde el panel, el bilingüismo y su fallback (inglés con contenido solo en español → se muestra en español), los estados borrador/publicado, los permisos, la auditoría, las rutas y las URLs del sitio se conservan íntegros; solo cambian la presentación, la navegación, el movimiento, el formato del selector de idioma y el tema (claro/oscuro).
- **FR-021**: F10 NO DEBE introducir datos nuevos ni cambiar los existentes: no agrega entidades, campos ni reglas de publicación; toda la información mostrada sigue siendo la de F3 y F2, gestionada igual desde el panel.
- **FR-022**: El sitio público y el panel DEBEN mantener la accesibilidad heredada (FR-019 de F3 y los requisitos de accesibilidad de F2; referencia WCAG 2.1 nivel AA — default revisable) **en ambos temas (claro y oscuro)**: navegación completa por teclado, compatibilidad con lectores de pantalla, contraste suficiente en todos los estados (incluidos los nuevos estados interactivos y las animaciones), ampliación de texto al 200 % sin pérdida de información, objetivos táctiles ≥ 44 px y foco visible siempre. La barra fija y el movimiento NO DEBEN dificultar la navegación por teclado ni por lector de pantalla.

**Panel de administración — rediseño (decisión Q1 del 2026-10-10)**

- **FR-023**: El sistema visual y de movimiento definidos en FR-001 a FR-009 DEBE aplicarse también al **panel de administración** en todas sus pantallas (inicio de sesión, inicio del panel, usuarios, roles, auditoría, portada e información, cambiar contraseña y pantalla de «sin permiso») y en su layout (navegación y cabecera del panel), con la misma coherencia y calidad que el sitio público; la navegación del panel DEBE ser clara y cómoda al desplazarse por pantallas densas, siguiendo el criterio de accesibilidad de FR-022.
- **FR-024**: El rediseño del panel NO DEBE cambiar su funcionalidad ni su comportamiento: inicio de sesión, usuarios (activar/desactivar), roles y permisos, auditoría, gestión de la portada (F3), validaciones, mensajes y pantallas de error se conservan íntegros; el panel sigue íntegramente en español y con su tono cercano (FR-016 de F3); el 100 % de las pruebas de F2 y de F3 DEBE seguir pasando sin cambios. Solo cambian su presentación, su movimiento y su tema (claro/oscuro).

**Modo oscuro (decisión Q7 del 2026-10-10)**

- **FR-025**: El sitio público y el panel DEBEN adaptarse al **modo oscuro del dispositivo**: cuando el dispositivo lo solicita, ambas superficies se presentan en tema oscuro, y en tema claro cuando el dispositivo lo solicita; el contenido, las funciones y el comportamiento son idénticos en ambos temas.
- **FR-026**: En ambos temas (claro y oscuro) y en todos los estados (reposo, hover, foco, activo, error, éxito, borrador/publicado), el contraste DEBE cumplir WCAG 2.1 nivel AA; el tema oscuro DEBE derivarse de la paleta de la marca sin romper sus reglas —el logotipo conserva SIEMPRE sus colores originales, NUNCA se recolorea ni se invierte para adaptarlo al tema oscuro— y sin perder legibilidad en tablas, formularios, estados y elementos densos del panel.

**Proceso de diseño (decisión Q8 del 2026-10-10)**

- **FR-027**: Antes de implementar, la fase de diseño DEBE presentar al humano **3 propuestas visuales distintas e innovadoras** (direcciones visuales diferentes entre sí, aplicadas a las mismas pantallas de referencia: la portada y una pantalla del panel) y el humano DEBE aprobar una propuesta antes de escribir código. Las referencias del cliente —`https://iglesiaoasis.com/` y `https://www.river-arena.org/`— son de **inspiración** (para captar el tipo de modernidad que busca) y NUNCA de copia: no se replican sus composiciones, textos, imágenes ni elementos identificables.

### Key Entities *(include if feature involves data)*

F10 **no introduce entidades ni datos nuevos** (FR-021): las entidades de F3 (identidad, «quiénes somos», horario de servicios, canales de WhatsApp, redes sociales, datos de contacto, estado de publicación) y las de F2 (cuentas, roles, permisos, auditoría) se conservan sin cambios y solo cambian de presentación. Los artefactos que F10 sí crea son conceptuales y de diseño:

- **Sistema visual (design tokens y componentes base)**: la definición única de tipografía, color, espaciado, jerarquía y componentes reutilizables del sitio público y del panel, construida sobre el Manual de Identidad y los tokens de F3; es el activo que heredan las secciones F4–F9 (FR-003, FR-018, FR-023).
- **Temas claro y oscuro**: las variantes de color del sistema para tema claro y tema oscuro, derivadas de la marca y verificadas en contraste AA, aplicadas al sitio y al panel según la preferencia del dispositivo (FR-025, FR-026).
- **Patrones de movimiento**: las reglas de animación y microinteracción de nivel moderado (qué se anima, cómo, cuándo y qué ocurre con movimiento reducido) aplicables al sitio y al panel (FR-005 a FR-009).
- **Patrón de navegación**: el modo en que el sitio presenta sus secciones —barra fija siempre visible— y cómo escala a las secciones futuras (FR-010 a FR-014).
- **Selector de idioma (formato)**: la presentación compacta del selector (mismos textos completos, menos tamaño y espacio), con su comportamiento heredado intacto de F3 (FR-015 a FR-017).

Relaciones: los cinco artefactos se aplican hoy a la portada de F3 y al panel de administración, y quedan como base para F4–F9; ninguno de ellos cambia las entidades ni el comportamiento funcional de F3 ni de F2.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100 % de las secciones de la portada (cabecera, identidad/portada, «quiénes somos», horario, WhatsApp, contacto, redes, pie) aplica el sistema visual, con 0 excepciones de tipografía, color, espaciado o componente en la revisión visual contra el sistema definido.
- **SC-002**: En la revisión de aceptación, el cliente da su visto bueno al rediseño como «moderno» y acorde a su marca («no parece un diseño típico y común de IA»), y en pruebas con personas el ≥ 80 % lo califica como «moderno o actual» frente al diseño anterior.
- **SC-003**: En pruebas con personas de distintas edades, incluidas con poca experiencia en internet, el ≥ 90 % encuentra y llega a la sección que busca desde la navegación sin ayuda y sin pelearse con el menú.
- **SC-004**: 0 desplazamientos horizontales de la página en los anchos probados (desde 320 px), 0 enlaces de navegación que requieran desplazar el menú de lado, y la navegación permanece visible y accesible en el 100 % de la página mientras se baja (barra fija), con la cantidad de secciones actual y con la previsión de F4–F9.
- **SC-005**: Con `prefers-reduced-motion` activo: 0 animaciones de entrada, revelado ni desplazamiento perceptibles en las pruebas, y 100 % del contenido y las funciones disponibles, idénticos a los del modo con movimiento.
- **SC-006**: En el 100 % de las cargas probadas, el contenido útil es visible e interactivo desde el primer instante (las animaciones no retrasan la lectura más de 0,5 s ni bloquean ninguna acción) y la navegación se percibe fluida, sin tirones, en el dispositivo de referencia más modesto.
- **SC-007**: El selector de idioma ocupa menos espacio en cabecera que el actual (comparación directa en la misma vista de referencia, conservando los textos completos «Español» y «English») y el 100 % de los participantes en pruebas lo identifica correctamente como selector de idioma, incluidas las pruebas con lector de pantalla.
- **SC-008**: 0 regresiones funcionales: el 100 % de las pruebas de aceptación de F3 (portada pública, bilingüismo y fallback, borrador/publicado, permisos, auditoría) y de F2 (inicio de sesión, usuarios, roles, permisos, auditoría) sigue pasando sin cambios.
- **SC-009**: La revisión de accesibilidad del sitio y del panel rediseñados, **en tema claro y en tema oscuro**, cierra con 0 hallazgos críticos (referencia WCAG 2.1 nivel AA), incluidos los nuevos estados interactivos, las animaciones y la barra fija, y es recorrerlos y operarlos por completo con teclado y con lector de pantalla.
- **SC-010**: 0 textos de interfaz nuevos sin traducir en el sitio público: el 100 % de las cadenas de interfaz que introduce el rediseño existe en español y en inglés (el panel sigue íntegramente en español).
- **SC-011**: Coherencia transversal: 0 elementos visuales con dos estilos distintos para la misma función (en el sitio y en el panel), y los cimientos del sistema (visual con ambos temas, movimiento, navegación, selector) quedan entregados y documentados de modo que la primera sección futura (F4) pueda construirse aplicándolos sin rediseñar el sistema.
- **SC-012**: El panel aplica el sistema visual y de movimiento en el 100 % de sus pantallas (inicio de sesión, inicio, usuarios, roles, auditoría, portada e información, cambiar contraseña y sin permiso), con 0 excepciones en la revisión visual, y **0 funciones perdidas**: verificado por las pruebas de F2 y F3 (SC-008) y por el recorrido de las operaciones del equipo.
- **SC-013**: Modo oscuro: el sitio y el panel se adaptan al tema del dispositivo con **0 fallos de contraste AA** en los elementos revisados en ambos temas, 0 diferencias de contenido o funciones entre temas, y el logotipo siempre con sus colores originales (0 recoloreados ni inversiones).
- **SC-014**: Proceso de diseño: **3 propuestas visuales distintas e innovadoras** presentadas al humano y **1 aprobada antes de escribir código**; 0 elementos copiados de las referencias (`iglesiaoasis.com`, `river-arena.org`) en la propuesta aprobada (revisión visual comparativa).

## Out of Scope

Queda explícitamente **fuera del alcance** de F10:

- **Construir las secciones F4–F9** (eventos y actividades, grupos de conexión, ministerios, donaciones, noticias/galería y medios): F10 solo entrega los cimientos para que nazcan con el sistema y aplica el rediseño donde hoy existe contenido y pantallas (la portada de F3 y el panel de F2/F3).
- **Cualquier cambio funcional de F3 o de F2**: contenido, datos, edición desde el panel, bilingüismo y fallback, estados borrador/publicado, usuarios, roles, permisos, auditoría, validaciones, mensajes, rutas y URLs no cambian (FR-020, FR-024). Si un cambio funcional se necesitara, sería una ampliación que se decide aparte.
- **Cambios en el backend, los datos o los contratos de información**: F10 es de presentación; no introduce datos nuevos ni cambia los existentes (FR-021).
- **Cambios al Manual de Identidad, al logotipo o a la marca**: F10 se construye **sobre** el manual y solo crea material gráfico nuevo de apoyo (iconografía e ilustraciones, decisión Q6); no modifica el manual ni el logotipo, y no crea banners ni material gráfico para redes sociales (`idea.md` §6).
- **`/health` («Estado del sistema» de F1)**: se conserva como está; es una página de diagnóstico de F1, no una pantalla del panel, y queda fuera de la navegación pública (decisión D-4 de F3). El rediseño del panel **sí** alcanza sus pantallas propias (inicio de sesión, usuarios, roles, auditoría, portada e información, cambiar contraseña y «sin permiso»), pero no `/health`.
- **Un control manual de tema (claro/oscuro) en la interfaz**: el modo oscuro se acopla a la preferencia del dispositivo (FR-025); un interruptor propio sería un ajuste aparte si el cliente lo pide (ver Assumptions).
- **Formularios de cualquier tipo** (contacto, inscripción, suscripción): sigue fuera de alcance (`idea.md` §6); la comunicación sigue siendo por enlaces y datos de contacto publicados.
- **Nuevas funciones de interacción del visitante ni del usuario del panel** (cuentas de visitante, registro, comentarios, buscador, mapas incrustados, nuevos módulos del panel, etc.): F10 no agrega comportamientos nuevos.
- **Traducción automática de contenidos** y cualquier cambio en el modelo de idiomas de contenido (se mantiene el de F3) o del idioma del panel (sigue en español).
- **Analítica web, medición en producción o pruebas A/B**: las métricas de éxito se verifican en las pruebas de aceptación, no con instrumentación en producción.
- **Despliegue a producción** y cambios de infraestructura: siguen pendientes a nivel de proyecto.

## Assumptions

Defaults razonables tomados donde la descripción no fija una regla; **son provisionales y revisables** por el humano. Los puntos que cambiaban alcance o experiencia quedaron **resueltas con el cliente el 2026-10-10** (ver *Aclaraciones resueltas (2026-10-10)*); los marcados «decisión firme» son el resultado de esa aclaración y solo cambian con un nuevo acuerdo del humano.

- **Los design tokens de F3 son la base**: F10 consolida y aplica la paleta, las tipografías y las reglas de espaciado/contraste ya definidas en `specs/003-portada-info-general/ux.md`; los ajustes de detalle del sistema (colores derivados por contraste, escala tipográfica, espaciado, radios) son parte del trabajo de F10 y se aprueban en la fase de UX, sin reinventar la identidad.
- **El Manual de Identidad manda**: ante conflicto entre las tendencias 2026 y el manual (paleta, tipografías, reglas del logotipo), prevalece el manual; la modernidad se logra con composición, jerarquía, espaciado, movimiento y componentes, no rompiendo la marca. Los materiales gráficos nuevos (iconografía e ilustraciones, Q6) también lo respetan.
- **Nivel de animación moderado (decisión firme, Q4)**: microinteracciones en hover/foco/clic y revelados suaves y breves del contenido; sin estilo expresivo, sin parallax obligatorio ni secuencias largas. `prefers-reduced-motion` sigue siendo obligatorio e incondicional (FR-006).
- **Barra de navegación fija siempre visible (decisión firme, Q2 y Q3)**: la navegación acompaña al visitante al desplazarse; el «scroll incómodo» se define como desplazamiento lateral de los enlaces en móvil y pérdida de la navegación al bajar, y la solución debe eliminar ambos casos.
- **Navegación sin menús ocultos (heredado de `idea.md`, sin cambios)**: todos los enlaces se muestran siempre; si la lista crece (previsión F4–F9), se acomoda en una o varias filas compactas y legibles, sin hamburguesa ni desplegables ocultos.
- **Selector de idioma: solo compactar los textos actuales (decisión firme, Q5)**: «Español» / «English» con menos tamaño y espacio; sin píldoras «ES / EN», sin menú desplegable y sin banderas; el nombre completo del idioma se conserva claro y accesible. El comportamiento de F3 no cambia (FR-016).
- **Material gráfico nuevo permitido (decisión firme, Q6)**: se pueden crear iconografía e ilustraciones propias respetando la paleta, las tipografías y las reglas del Manual y del logotipo; el logotipo y el Manual NO se modifican.
- **Modo oscuro acoplado al dispositivo (decisión firme, Q7)**: se sigue la preferencia del sistema (tema claro/oscuro) sin interruptor manual propio en la interfaz (default revisable: si el cliente quiere un control de tema en la interfaz, es un ajuste menor que se acuerda antes de implementar). Si el dispositivo cambia de tema con la aplicación abierta, la interfaz lo respeta sin recargar (default revisable).
- **Proceso de diseño con 3 propuestas (decisión firme, Q8)**: la fase de UX presenta 3 direcciones visuales distintas e innovadoras sobre las mismas pantallas de referencia (la portada y una pantalla del panel) y el humano aprueba una antes de implementar; las referencias `https://iglesiaoasis.com/` y `https://www.river-arena.org/` son de inspiración, nunca de copia.
- **Panel incluido en F10 (decisión firme, Q1)**: el rediseño alcanza a todas las pantallas y al layout del panel (inicio de sesión, inicio, usuarios, roles, auditoría, portada e información, cambiar contraseña y «sin permiso»), sin cambiar su funcionalidad; este alcance amplía el trabajo de diseño e implementación que inicialmente se preveía solo para el sitio público.
- **La preferencia de movimiento se respeta también si cambia con el sitio abierto** (default revisable): no se exige recargar la página para que surta efecto.
- **Composición de la portada**: F10 puede reordenar la presentación de las secciones existentes de la portada si mejora la experiencia, siempre que todas las secciones publicadas sigan presentes y sin cambiar su contenido ni su comportamiento; cualquier reordenación se muestra al humano en la revisión de las 3 propuestas antes de implementar.
- **Sin peso perceptible de carga**: se prioriza que el movimiento y el nuevo aspecto (en ambos temas) no agreguen espera perceptible; si para lograrlo hiciera falta una herramienta nueva de animación, su necesidad se justifica en el plan (constitución §II, minimizar dependencias).
- **Idiomas de la interfaz**: todo texto nuevo del sitio público es/en (FR-019); el panel sigue íntegramente en español (FR-016 de F3, FR-024).
- **Anchos y modos de prueba**: mínimo 320 px de ancho y zoom 200 % (hereda F3); pruebas en teléfono, tableta y computadora, con y sin `prefers-reduced-motion`, y en tema claro y tema oscuro.
- **Accesibilidad**: WCAG 2.1 nivel AA como referencia (default revisable, igual que F3); el requisito de fondo sigue siendo el de `idea.md` §7: apto para personas de todas las edades y dispositivos.
- **Verificación de «moderno» y de satisfacción**: se hace en la revisión de las 3 propuestas y en la revisión de aceptación del humano (antes de implementar y al cierre) y en pruebas de aceptación con personas reales de distintas edades; no existe instrumentación de medición en producción (ver Out of Scope). Las referencias del cliente sirven como calibración del tipo de modernidad buscada, sin copiarlas (FR-027).
- **Dependencias**: F3 terminada (portada funcional que no se rompe), F2 terminada (panel que no se rompe) y el Manual de Identidad entregado por el cliente. El orden recomendado por el roadmap (F10 antes de F4) se respeta salvo decisión distinta del cliente.
- **Las decisiones ⟲ de F3 sobre el menú (D-5) y el selector (D-1) se revisan aquí**: F10 puede sustituirlas —el menú se resuelve con barra fija sin scroll incómodo y el selector se compacta—, pero siempre dentro de las restricciones fijas: sin menús ocultos ni gestos desconocidos, claridad para todas las edades, textos completos de idioma accesibles y comportamiento de idioma idéntico al de F3.

## Aclaraciones resueltas (2026-10-10)

**Ninguna aclaración pendiente.** Las ocho preguntas abiertas de la primera versión se resolvieron con el cliente el 2026-10-10 (quien decide: cliente). Queda aquí el registro de la decisión tomada en cada una.

### Q1 — Alcance del rediseño: ¿solo el sitio público o también el panel? — **Resuelta**

**Decisión**: F10 **incluye también el rediseño visual del panel de administración**, no solo el sitio público. El panel deja de estar fuera de alcance: el rediseño alcanza el inicio de sesión, usuarios, roles, auditoría, portada e información, cambiar contraseña, la pantalla de «sin permiso» y el layout del panel (FR-023), sin cambiar su funcionalidad (FR-024).

### Q2 — Significado del «scroll incómodo» del menú — **Resuelta**

**Decisión**: eran **ambos** casos: (a) los enlaces se desplazan lateralmente en la cabecera en móvil y (b) la navegación se pierde/desaparece al bajar la página. La solución debe eliminar los dos (FR-011).

### Q3 — Navegación al bajar la página — **Resuelta**

**Decisión**: **barra fija siempre visible**: la navegación acompaña al usuario al desplazarse (FR-014).

### Q4 — Nivel de animación — **Resuelta**

**Decisión**: **moderado** (microinteracciones en hover/foco/clic y revelados suaves y breves), sin llegar a expresivo; se mantiene obligatorio respetar `prefers-reduced-motion` (FR-005, FR-006).

### Q5 — Formato del selector de idioma — **Resuelta**

**Decisión**: **solo compactar los textos actuales** («Español» / «English») reduciendo tamaño y espacio, **sin píldoras «ES / EN» y sin menú desplegable**. Se mantiene el requisito de que el nombre completo del idioma sea claro y accesible (no siglas) y el comportamiento de F3 no cambia (FR-015, FR-016).

### Q6 — Iconografía e ilustraciones nuevas — **Resuelta**

**Decisión**: **sí se permite crear material gráfico nuevo** (iconografía e ilustraciones) respetando la paleta y las reglas del Manual de Identidad y del logotipo (FR-002).

### Q7 — Modo oscuro — **Resuelta**

**Decisión**: **se incluye en F10**: el sitio público y el panel de administración se adaptan al modo oscuro del dispositivo, con contraste AA en ambos temas (FR-025, FR-026).

### Q8 — Referencias visuales y proceso de diseño — **Resuelta**

**Decisión**: referencias **de inspiración, no para copiar**: `https://iglesiaoasis.com/` y `https://www.river-arena.org/`. Además, decisión de proceso: el cliente quiere ver **3 propuestas de diseño distintas e innovadoras antes de implementar** y aprueba una antes de codificar (FR-027).
