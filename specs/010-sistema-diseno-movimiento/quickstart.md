# Quickstart: F10 — Sistema de diseño y movimiento

**Feature**: F10 · **Branch**: `010-sistema-diseno-movimiento`

Cómo **levantar** el sitio público rediseñado y **validar** cada parte del alcance: sistema visual
(con la propuesta aprobada **B «Marea Turquesa»** y los **tokens de espaciado de sección**), tema
claro/oscuro (sistema → interruptor → persistencia), movimiento **marcado/elegante con revelado al
hacer scroll** y movimiento reducido, navegación **responsive** (barra fija en escritorio + **menú
hamburguesa accesible en móvil**), **mapa del sitio en el pie** y selector de idioma compacto. Sirve
como guía de QA manual (SC-002, SC-007, SC-009, SC-012, SC-014) y de verificación de
**no-regresión** (SC-008).

> Nota: este documento se escribe en la fase de plan; los comandos y rutas son los del stack
> vigente (F1–F3). Si la implementación mueve algo, se actualiza aquí.

---

## 1. Levantar el stack

```bash
make up            # docker compose: PostgreSQL + backend + frontend en http://localhost:5173
make e2e           # Playwright local contra el stack real (frontend/e2e/)
```

Para iterar solo en frontend:

```bash
cd frontend
npm install
npm run dev        # http://localhost:5173
npm test -- --run  # unitarias (Vitest + Testing Library)
npm run typecheck && npm run lint
```

Apertura de referencia: la portada publicada en `/` (con contenido real cargado desde el panel, que
**no se ha tocado**). Útil también `/health` para comprobar que F1 sigue igual.

## 2. Validar el sistema visual (US1, SC-001/SC-002)

En **teléfono (320 px), tableta y computadora**, recorrer la portada completa (cabecera,
identidad/portada, «quiénes somos», horario, WhatsApp, contacto, redes, pie):

- [ ] Todas las secciones aplican el **mismo** sistema: tipografías de marca (Bebas Neue / Poppins /
      Playfair Display), proporción de color 70/20/10, espaciado y jerarquía coherentes.
- [ ] El conjunto se percibe **moderno y actual**, distinto del diseño previo y **no** como una
      plantilla genérica (comparación visual lado a lado; calibración con las referencias del
      cliente solo como inspiración — nada copiado).
- [ ] El **logotipo**: sin deformar, sin recolorear, sin girar, sin efectos, **en ambos temas y en
      todos los estados** (reposo, hover, foco, activo).
- [ ] Contenido real de distinta forma (textos largos, un servicio suelto, contenido solo en español
      navegando en inglés, varias redes): legible, sin huecos ni desbordes.
- [ ] Todo elemento interactivo se percibe como tal (hover, foco, activo) con contraste AA.
- [ ] **Aire entre secciones** (tokens `--space-section*`, ajuste del cliente): las secciones se
      separan con un espaciado vertical **homogéneo y perceptiblemente mayor** que antes («más
      aire»), que **crece con el ancho** de la pantalla (móvil → escritorio) y **no** resulta en
      huecos vacíos ni en un scroll interminable; encabezado de sección y contenido quedan bien
      relacionados (`--space-block`). Calibración final en esta revisión (si «vacía» o «apretada»,
      se ajustan los tokens, no cada componente).
- [ ] Zoom al **200 %** y 320 px: sin desbordes ni pérdida de información.

## 3. Validar el tema claro/oscuro (US5, SC-012)

**Tres caminos** (en cada uno: contraste AA, logo con colores originales, mismo contenido y mismas
funciones que en claro):

1. **Sistema**: poner el dispositivo/ navegador en **modo oscuro** y abrir el sitio con una pestaña
   limpia (sin `ss.theme`) → el sitio se muestra **en oscuro**. (En Chrome DevTools:
   ⋮ → More tools → Rendering → *Emulate CSS prefers-color-scheme*.)
2. **Interruptor manual**: con el sitio abierto, usar el control de tema de la cabecera → el tema
   cambia **de inmediato, sin recargar**; **recargar** y **cerrar y reabrir** el navegador → la
   elección **persiste** (`localStorage['ss.theme']`).
3. **Primera visita sin elección**: borrar `ss.theme` (DevTools → Application → Local Storage) y
   alternar el modo del sistema → el sitio **sigue al sistema** (claro y oscuro). Si el sistema
   cambia **con el sitio abierto** y no hay elección manual, la interfaz lo sigue **sin recargar**.

Además:

- [ ] La elección manual **manda** hasta que se cambia (aunque el sistema cambie después).
- [ ] El control de tema es **compacto** (no ocupa espacio innecesario), accesible por teclado y con
      lector de pantalla (nombre claro del control y de su estado), foco visible.
- [ ] El sitio en oscuro **no muestra parpadeos** al cargar (ni «claro → oscuro» al abrir:
      anti-FOUC).
- [ ] 0 fallos de contraste AA en textos, botones, estados interactivos y foco **en ambos temas**
      (revisión por estado: reposo, hover, foco, activo, éxito/error si los hay).

## 4. Validar el movimiento y el movimiento reducido (US3, SC-005/SC-006/SC-014)

> Nivel esperado (**marcado/elegante con revelado al hacer scroll**, FR-005): el movimiento se
> **percibe con claridad** —nada «casi imperceptible»— pero sigue siendo elegante: sin parallax, sin
> mareos y sin secuencias largas. Parámetros de referencia (tokens de `research.md` R10-13):
> revelados ≈ 600 ms con desplazamiento de 20–28 px, escalonado de 80 ms entre componentes,
> microinteracciones ≈ 180 ms, techo de 800 ms por secuencia.

**Con movimiento normal** (sistema sin «reducir movimiento»):

- [ ] Al **bajar** por la página, cada sección y sus componentes **se revelan al entrar en pantalla**
      de forma claramente perceptible (fade + desplazamiento), con **entrada escalonada** en grupos
      (tarjetas de una fila, ítems de una lista): se ve la cascada, no una entrada única plana.
- [ ] Las transiciones entre secciones (salto por ancla) y las microinteracciones (hover, foco,
      activación) se **notan** y son agradables; **sin desplazar** el maquetado y **sin** efectos
      mareantes (sin parallax, sin rebotes elásticos, sin desenfoques).
- [ ] Ninguna animación dura lo que una espera: la secuencia más larga de un bloque termina en
      **≤ ~0,8 s**; el contenido se lee con normalidad mientras las animaciones se reproducen.
- [ ] **El contenido nunca queda bloqueado**: lo que aún se está revelando es **clicable** y operable
      (una acción durante la animación responde al instante); el foco de teclado sobre un elemento
      pendiente lo **muestra de inmediato**.
- [ ] **Cabecera al bajar**: al desplazarse, la cabecera **transiciona** a su estado compacto de
      superficie (fondo más sólido, borde/sombra suaves) de forma perceptible **sin cambiar de
      altura** ni descolocar el contenido; al volver arriba, regresa.
- [ ] **Anclas con desplazamiento suave**: un clic en un enlace de navegación desliza hasta la
      sección (no salta en seco) y el título del destino queda visible (la barra no lo tapa).
- [ ] Cambiar de idioma o de tema **no re-anima** secciones vistas ni provoca parpadeos, rebotes ni
      descolocaciones (SC-014).
- [ ] El contenido útil es visible e interactivo **desde el primer instante** (la animación es una
      capa visual); fluidez en un dispositivo modesto (checklist de rendimiento: sin tirones al bajar;
      en DevTools → Performance no hay tareas largas atribuibles al movimiento).

**Con movimiento reducido** (sistema con «reducir movimiento»; DevTools → Rendering → *Emulate
prefers-reduced-motion: reduce*, o la opción del sistema; en e2e:
`page.emulateMedia({ reducedMotion: 'reduce' })`):

- [ ] **0 animaciones** de entrada, revelado, escalonado o desplazamiento: el contenido aparece
      **completo de inmediato** (sin fades ni desplazamientos al bajar), y la cabecera cambia de
      estado **sin transición**.
- [ ] El salto a las anclas es **instantáneo** (sin desplazamiento suave).
- [ ] **100 % de las funciones y la información** disponibles, idénticas al modo con movimiento.
- [ ] Activar «reducir movimiento» **con el sitio abierto** surte efecto sin recargar.
- [ ] El **menú hamburguesa móvil** se abre/cierra **sin animación** (instantáneo) y el **mapa del
      sitio del pie** no añade transiciones de entrada.

**Comparación directa (SC-005/SC-014)**: recorrer la portada en ambas configuraciones, una detrás de
otra: en modo normal el revelado al scroll y las transiciones se **perciben claramente**; con
movimiento reducido están **por completo ausentes** y todo funciona igual.

## 5. Validar la navegación (US2, SC-003/SC-004)

> **Patrón por dispositivo** (decisión del cliente 2026-10-10): **escritorio (≥ 861 px)** — barra
> fija con **todas** las anclas siempre a la vista (sin menús ocultos); **móvil (≤ 860 px)** — las
> anclas viven en un **menú hamburguesa accesible** detrás de un botón siempre visible (el corte de
> 860 px es el de la maqueta aprobada `propuesta-b.html`). En ambos modos: a un clic o toque, sin
> gestos, sin scroll incómodo.

### 5.1 Escritorio y tableta apaisada (≥ 861 px)

- [ ] **Todos** los enlaces de navegación están **a la vista** (sin menús ocultos, sin hamburguesa,
      sin desplegables, sin gestos) y caben en filas compactas **sin scroll lateral**.
- [ ] Al **bajar** por la página, la navegación **permanece visible y accesible** (barra fija), sin
      tapar contenido ni descolocar el maquetado; la cabecera **transiciona su superficie** al bajar
      (estado `data-scrolled`, ver §4) sin cambiar de altura.
- [ ] Clic en un enlace → la **sección destino y su título quedan visibles** (la barra no los tapa),
      con desplazamiento **suave** (instantáneo con movimiento reducido, ver §4).
- [ ] La página **nunca** provoca desplazamiento horizontal (comprobar a 861 px, 1024 px y 1280 px,
      y con ~8–10 enlaces si se simula la previsión de F4–F9).

### 5.2 Móvil (≤ 860 px): menú hamburguesa accesible (a 320 px)

- [ ] La cabecera muestra marca, selector de idioma, control de tema y el **botón hamburguesa**
      (objetivo **≥ 44 px**, con **nombre accesible** claro —lector de pantalla lo anuncia como menú
      de secciones— y estado anunciado: **`aria-expanded`** / **`aria-controls`** correctos).
- [ ] **Abrir/cerrar con ratón** (clic) **y con teclado** (`Enter`/`Espacio` sobre el botón): al
      **abrir, el foco entra en la lista** (primer enlace); **`Escape`**, el **control visible de
      cierre**, el **toque/clic fuera** (scrim) y el propio botón lo cierran y **devuelven el foco al
      botón**; **sin trampa de foco** (patrón *disclosure* de APG, **no** un diálogo: sin
      `role="dialog"` ni `aria-modal`): `Tab`/`Shift+Tab` **pueden salir del panel** hacia la barra
      (el idioma y el tema siguen operables) y el resto de la página; al elegir una sección,
      el panel se cierra y el foco sigue en el destino.
- [ ] El panel es **la misma lista de enlaces de la navegación** (una sola lista en el DOM, sin
      duplicados): con el menú **cerrado** a ≤ 860 px esos enlaces **no son alcanzables con `Tab`**
      (quedan **fuera del orden de tabulación**, no solo invisibles) y al abrir vuelven a ser
      visibles y enfocables.
- [ ] Con el panel abierto, la página de fondo **no se desplaza** (ni con rueda ni con teclado) y el
      panel se adapta a pantallas bajas (scroll interno propio si hace falta).
- [ ] **Sin gestos**: todo se opera con toques, clics y teclado (ni swipe ni deslizamientos).
- [ ] Se llega a cualquier enlace **sin desplazar enlaces de lado** y la página **nunca** provoca
      desplazamiento horizontal a 320 px; todos los enlaces visibles del sitio cumplen **≥ 44 px**.
- [ ] Al ensanchar la ventana más allá del breakpoint (**860 px**, o al rotar el dispositivo) el panel
      **se cierra solo** y reaparecen las filas de anclas de escritorio, sin recargas ni restos de
      estado.
- [ ] Clic en un enlace del panel → sección destino y título visibles (barra no tapa), desplazamiento
      suave (instantáneo con movimiento reducido).
- [ ] Teclado y lector de pantalla: orden comprensible, **foco visible siempre**, estado del menú
      anunciado, etiquetas correctas; una persona con poca experiencia abre el menú y llega a su
      sección con un solo toque (SC-003).

### 5.3 Mapa del sitio en el pie (ajuste del cliente)

- [ ] El **pie** lista las secciones del sitio público (las de la portada de hoy) en **columnas en
      escritorio** y **apilado en móvil**, con el mismo lenguaje y los mismos nombres que la cabecera.
- [ ] **0 enlaces rotos**: las secciones en borrador/ausentes **no** dejan enlaces (mismo criterio que
      la cabecera) y **no** se enlazan rutas de F4–F9 que aún no existen (cuando existan, entran al
      modelo y el pie las muestra sin tocar el componente).
- [ ] Accesible: landmark `nav` con **nombre accesible propio** (distinto del de la cabecera), lista
      real de enlaces, foco visible, objetivos ≥ 44 px, contraste AA **en ambos temas** (la auditoría
      de contraste de los e2e ya incluye el pie).
- [ ] **Sin grupos vacíos**: un grupo sin ítems **no se renderiza** y, si no hay nada que listar (0
      secciones, sin contacto ni redes), **no se dibuja el mapa**; los rótulos de grupo son
      etiquetas planas (**no** encabezados).
- [ ] **Nombres accesibles únicos**: los enlaces del grupo Contacto llevan prefijo de contexto
      («Contacto: …»; la dirección es **texto**, no enlace) y los de Síguenos etiquetas planas
      («Facebook», «Instagram», «YouTube»): ningún nombre del pie duplica el de un enlace de sección
      ni el de un CTA.
- [ ] Clic en un enlace del pie → el título de la sección destino queda visible (la barra fija no lo
      tapa).

## 6. Validar el selector de idioma (US4, SC-007)

- [ ] **Menos espacio** que el actual (comparación directa en la misma vista: cabecera de F3 vs.
      cabecera de F10) conservando los textos completos **«Español»** y **«English»** (sin siglas,
      sin píldoras, sin menú, sin banderas).
- [ ] Identificable como selector de idioma por personas de poca experiencia y por lector de pantalla
      (nombre accesible = nombre completo del idioma).
- [ ] Comportamiento **idéntico a F3**: cambio inmediato; se mantiene al navegar; se recuerda entre
      visitas (`localStorage['ss.lang']`); primera visita sin preferencia → **español**.
- [ ] Igual en cabecera y en pie; sin desplazamientos ni tapados en ningún ancho.

## 7. Validar textos, accesibilidad y fronteras (SC-008/SC-009/SC-010)

- [ ] Cambiar a inglés: **0 textos de interfaz nuevos sin traducir** (incluidos los del control de
      tema, los del **menú hamburguesa** —nombre del botón y del control de cierre— y los del **mapa
      del sitio**); el contenido sin inglés muestra su fallback en español integrado en el diseño.
- [ ] Recorrido completo **por teclado y con lector de pantalla** en **ambos temas**: foco visible,
      orden comprensible, todo operable (referencia WCAG 2.1 AA; checklist manual de QA). **En móvil
      (320 px) incluye el menú hamburguesa**: abrir/cerrar con teclado, `Escape`, foco devuelto al
      botón, estado del menú anunciado, y recorrido del **mapa del sitio del pie**.
- [ ] **Panel de administración SIN CAMBIOS**: login, `/panel/informacion`, usuarios, roles y
      auditoría lucen y funcionan **exactamente igual que antes de F10** (F10B no existe aún);
      verificación con las suites e2e de panel y una pasada visual.
- [ ] `/health` sin cambios; rutas y URLs sin cambios.

## 8. Pruebas automatizadas (no-regresión incluida)

```bash
cd frontend
npm test -- --run        # unitarias: tema, movimiento, navegación (menú hamburguesa móvil y mapa del
                         # sitio del pie), selector + TODAS las de F2/F3
npm run typecheck
npm run lint
make e2e                 # Playwright: diseno-movimiento, navegacion-movil, tema-oscuro, movimiento,
                         #   movimiento-reducido
                         # + portada-publica, portada-accesibilidad, portada-panel, portada-imagenes,
                         #   portada-patch-null, acceso, auditoria, status  (SIN CAMBIOS, SC-008)
make ci                  # pipeline completo
```

**Criterio de cierre**: todas las suites pasan **sin editar las pruebas existentes** (SC-008). Si una
prueba de F3 exigiera un cambio visual que la propuesta aprobada provoca, se escala al humano: no se
modifica la prueba para hacerla pasar.

## 9. Revisión de aceptación (humano)

- [ ] La propuesta visual elegida —**B «Marea Turquesa»** (aprobada 2026-10-10, FR-026)— se reconoce
      en la implementación y el cliente la califica como **moderna y acorde a su marca** («no parece
      un diseño típico y común de IA») (SC-002).
- [ ] Los **3 ajustes del cliente** quedan verificados: (1) navegación responsive —barra fija con
      todas las anclas en escritorio, **menú hamburguesa accesible** en móvil (§5.2)—; (2) **mapa del
      sitio en el pie** sin enlaces rotos (§5.3); (3) **más aire entre secciones** con tokens de
      espaciado (§2).
- [ ] El **movimiento** se percibe como **marcado y elegante**: el cliente confirma que el revelado
      al scroll, el escalonado y las transiciones **se notan de verdad** (nada «casi imperceptible»)
      y que el conjunto **no resulta estridente ni mareante** (SC-014).
- [ ] 0 elementos copiados de las referencias (`iglesiaoasis.com`, `river-arena.org`, `apple.com`,
      `microsoft.com/es-cr`) (SC-013).
- [ ] Cimientos para F4–F9 entregados y documentados (`ux.md` + este plan): una sección nueva
      construida con ellos encaja sin rediseño (SC-011), incluidos los **tokens de movimiento**, los
      **tokens de espaciado** y el modelo de secciones que alimenta nav **y** pie.
