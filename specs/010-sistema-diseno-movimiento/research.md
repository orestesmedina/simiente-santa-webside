# Research: F10 — Sistema de diseño y movimiento

**Feature**: F10 — Sistema de diseño y movimiento · **Branch**: `010-sistema-diseno-movimiento`
**Fase**: 0 de `/speckit.plan` (decisiones técnicas; numeración **R10-1…R10-12**, propia de F10)
**Fecha**: 2026-10-10

F10 es una funcionalidad **100 % de presentación del sitio público** (frontend). No hay trabajo de
backend, ni datos, ni contratos (ver `data-model.md` y `contracts/README.md`). Las decisiones de
arquitectura son, por tanto, de **estrategia visual, tema, movimiento, navegación y presentación del
selector**, todas sobre el stack vigente (React 19 + TS strict + Vite + Tailwind 4 + React Router 7 +
TanStack Query; Vitest + Testing Library; Playwright).

**Restricción de proceso (FR-026)**: la **dirección visual** concreta la fija `ux.md` (3 propuestas y
el humano aprueba una, en paralelo a este plan). Todo lo que aquí se decide es **mecánica** —tokens,
tema, movimiento, navegación, formato del selector— y está deliberadamente aislado de la dirección
visual: la propuesta elegida puede cambiar composición, acentos e ilustraciones **sin alterar** estas
decisiones.

---

## R10-1 — Mecanismo de tema: atributo `data-theme` en `<html>` + variables semánticas + variante `dark` de Tailwind 4

**Decisión**: el tema se expresa como atributo **`data-theme="light" | "dark"` en `<html>`** (el mismo
nodo donde F3 ya pone `document.documentElement.lang`). Sobre él:

1. Una capa de **variables CSS semánticas** de superficie/texto/borde/acento (`--surface`,
   `--surface-raised`, `--text`, `--text-muted`, `--border`, `--accent-contrast`…), con valores
   distintos bajo `html[data-theme='light']` y `html[data-theme='dark']`, registradas en Tailwind 4
   con `@theme inline` para que existan utilidades `bg-surface`, `text-muted`… que cambian solas con
   el tema.
2. Una **variante `dark` personalizada** de Tailwind 4
   (`@custom-variant dark (&:where([data-theme='dark'], [data-theme='dark'] *))`) para los pocos casos
   puntuales que no se resuelven con tokens (p. ej. la superficie de respaldo bajo el logotipo).

**Por qué**: el atributo es autoexplicativo, extensible (si en el futuro se añade un tercer estado
explícito «seguir sistema»), accesible desde CSS y JS, y sigue el patrón ya usado por F3 con `lang`.
Las variables semánticas hacen que el 90 % del rediseño ni siquiera mencione el tema (los colores
concretos viven en una sola tabla por tema), lo que reduce el riesgo de olvidos de contraste.

**Alternativas descartadas**:
- `dark:` de Tailwind 4 con `prefers-color-scheme` **solo**: no admite interruptor manual (FR-023).
- Clase `.dark` en `<html>`: equivalente técnico; se prefiere el atributo por legibilidad y
  extensibilidad (y porque `class` puede ser compartida con otras clases de layout sin querer).
- Re-tematizar **los tokens existentes** de F3 (`--color-navy`, `--color-cream`…) por tema:
  **descartado porque esos tokens también los consume el panel** (F2/F3); redefinirlos cambiaría la
  apariencia del panel, que F10 **no puede tocar** (FR-020). Ver R10-2.
- Librería de temas (`next-themes` y similares): dependencia nueva para ~40 líneas de lógica (§II).

## R10-2 — Capa de tokens semánticos **aditiva**; tokens de marca de F3 intactos

**Decisión**: `frontend/src/index.css` **conserva sin modificar** el `@theme` actual (paleta
`navy/teal/cream/coral/leaf`, tipografías, `--shadow-brand`). F10 **añade** un bloque de tokens
semánticos (`--surface`, `--text`, `--border`, `--accent`…) con sus dos valores por tema. El sitio
público migra sus componentes a las utilidades semánticas; el panel sigue usando los tokens y clases
explícitas de hoy.

**Por qué**: garantiza «cero cambios visuales en el panel» (FR-020/SC-008) por construcción: si una
clase del panel no cambia, su pintura no cambia. Además, los tokens semánticos son el «cimiento» que
F4–F9 hereda: una sección nueva usa `bg-surface` y queda automáticamente correcta en ambos temas.

**Alternativa descartada**: duplicar la hoja de estilos por tema (dos fuentes de verdad, deriva
garantizada).

## R10-3 — Resolución, persistencia y anti-parpadeo del tema

**Decisión**: `ThemeProvider` en el sitio público (mismo patrón que `LanguageProvider` de F3):

- **Clave de memoria**: `localStorage['ss.theme']` con valores `'light' | 'dark'` (nomenclatura
  coherente con `ss.lang`). **Ausente = seguir al sistema** (no se guarda un valor «system»).
- **Precedencia**: elección manual guardada → si no la hay, `matchMedia('(prefers-color-scheme: dark)')`
  **en vivo** (listener `change`: si el dispositivo cambia de tema y no hay elección manual, la
  interfaz lo sigue sin recargar — default revisable de la spec, cumplido así).
- **Persistencia**: una elección manual se aplica de inmediato, se guarda y manda en visitas
  siguientes hasta que se cambie (FR-023). Si `localStorage` falla (navegación privada), se usa el
  tema del sistema sin avisos (mismo criterio de tolerancia de F3).
- **Anti-flash (FOUC)**: script **inline mínimo en `frontend/index.html`** que se ejecuta antes del
  primer pintado, lee `ss.theme` (o el sistema) y fija `data-theme` + `color-scheme` en `<html>`.
  Después, React solo mantiene el estado en sincronía. Alternativa descartada: resolver el tema solo
  en React (el primer pintado llegaría en claro y «parpadearía» el oscuro — edge case explícito de la
  spec: «sin parpadeos que hagan parecer el sitio roto»).
- `ThemeToggle` (cabecera) alterna claro/oscuro; su estado accesible describe el tema **efectivo**
  (FR-023: nombre claro del control y de su estado).

**Alternativas descartadas**: cookie (innecesaria); guardar `'system'` como valor explícito (el valor
ausente ya lo expresa; si UX pide un tercer botón «Sistema», el mecanismo lo admite sin migración);
`sessionStorage` (no persiste entre visitas, contradice FR-023).

## R10-4 — Movimiento: CSS puro + `IntersectionObserver`; **sin librería de animación**

**Decisión**: el nivel **moderado** (Q4) se construye con:
- **Transiciones CSS** en estados hover/foco/activo (solo `transform`, `opacity`, `color`,
  `background-color`, `box-shadow`, `border-color`: nunca propiedades que muevan el maquetado).
- **Revelado suave y breve** del contenido con un componente `Reveal` ligero basado en
  `IntersectionObserver` (un observer compartido por página), con `duration` corta (≈200–300 ms) y
  desplazamiento mínimo o nulo (Q4: nada expresivo).
- Keyframes CSS para detalles puntuales (p. ej. el acento de sección).

**Por qué sin librería** (constitución §II — dependencias nuevas requieren justificación y se
minimizan): el nivel moderado no necesita orquestación de timelines, ni animación de layout, ni
gestos; `framer-motion`/`GSAP` aportan peso de bundle, JS en cada frame y riesgo de descolocar el
maquetado para un efecto que CSS hace nativamente. **0 dependencias nuevas** en F10.

**Alternativas descartadas**: `framer-motion` (dependencia + JS-driven; innecesaria para «moderado»);
`tailwindcss-animate` (plugin sin valor frente a utilidades propias); **scroll-driven animations /
`@starting-style`** de CSS como base (soporte de navegador aún desigual en 2026: se podrían usar solo
como mejora progresiva, nunca como único mecanismo).

**Rendimiento (FR-007/SC-006)**: el contenido es visible e interactivo **desde el primer instante**
(ver R10-5); las animaciones solo tocan `opacity`/`transform` (compositor, sin reflow); un
`IntersectionObserver` compartido (no uno por elemento); sin animaciones que bloqueen clics
(`pointer-events` intactos).

## R10-5 — `prefers-reduced-motion` incondicional + revelado progresivo a prueba de fallos

**Decisión (doble capa)**:
1. **CSS**: bloque global `@media (prefers-reduced-motion: reduce)` que desactiva transiciones y
   animaciones (o las reduce a instantáneas) en todo el sitio público, y `html { scroll-behavior }`
   suave **solo** bajo `@media (prefers-reduced-motion: no-preference)`.
2. **JS**: el hook `usePrefersReducedMotion` (matchMedia con listener en vivo — la preferencia puede
   cambiar con el sitio abierto sin recargar, edge case de la spec) hace que `Reveal` **nunca oculte**
   contenido cuando hay movimiento reducido: se renderiza ya visible.

**Diseño a prueba de fallos (FR-009, edge cases de la spec)**:
- El estado por defecto del contenido es **visible**; el estado oculto se aplica **solo si** JS montó
  `Reveal` **y** el movimiento está permitido. Si el JS falla o la animación no carga: contenido
  visible, nunca zona en blanco.
- **Foco sobre contenido pendiente de revelado**: manejador `focusin` (y regla CSS
  `:focus-visible` equivalente) que hace visible el elemento **de inmediato**; el foco nunca queda
  sobre algo transparente.
- **Sin repeticiones molestas**: el revelado se ejecuta **una vez por elemento y por carga**; cambiar
  de idioma o de tema re-renderiza sin re-montar (claves React estables, `key` por sección = ancla),
  de modo que no se re-animan secciones ya vistas (US3 esc. 5).

## R10-6 — Navegación: barra **fija** con `position: sticky` + `scroll-padding-top` medido + filas que envuelven

**Decisión**:
- **Fijeza**: `position: sticky; top: 0` en el `<header>` del sitio público (se queda en el flujo del
  documento). Acompaña al visitante al bajar (caso (b) del «scroll incómodo», FR-014) **sin tapar
  contenido**: al estar en flujo, nunca se superpone al texto.
- **Destino de anclas visible**: `html { scroll-padding-top: var(--header-offset) }`, donde
  `--header-offset` se mide con `ResizeObserver` (altura real de la cabecera: cambia con las filas que
  envuelven, el zoom al 200 % y el texto ampliado). Así, al saltar a una sección, su título queda
  por debajo de la barra, no debajo de ella (edge case «barra fija no tapa el destino»).
- **Sin desplazamiento lateral de enlaces (caso (a))**: la lista de anclas pasa de
  `overflow-x-auto` (la causa actual del problema) a **`flex-wrap` con filas compactas**: todos los
  enlaces visibles siempre, sin scroll lateral del menú ni de la página. En 320 px la cabecera se
  organiza en dos filas (fila 1: marca + controles; fila 2: anclas), sin hamburguesa ni desplegables
  (FR-010/FR-011).
- **Escala a F4–F9 (FR-012)**: los enlaces siguen siendo datos (`sections.ts` → modelo de navegación
  reutilizable); con ~8–10 secciones se acomodan en una o varias filas compactas (default revisable de
  la spec) — sin menús ocultos. Las rutas futuras F4–F9 ya cuelgan de `PublicLayout` (patrón de
  router de F3), así que la cabecera fija, el tema y el selector acompañan a las secciones nuevas sin
  cambios.

**Alternativas descartadas**:
- `position: fixed` + padding compensatorio en `body`: se superpone al contenido y acentúa los
  problemas de zoom 200 %/320 px (FR-025 exige que nada se tape).
- Menú hamburguesa / desplegable: **prohibido** por la restricción del encargo (sin menús ocultos ni
  gestos desconocidos).
- Carrusel horizontal de enlaces: es exactamente el caso (a) del «scroll incómodo».
- `scroll-margin-top` por sección: equivalente válido; se centraliza en `scroll-padding-top` (una sola
  fuente de la altura de cabecera).

## R10-7 — Selector de idioma: solo formato compacto; comportamiento de F3 congelado

**Decisión**: se conserva **intacto** el mecanismo de F3 —dos **radios nativos** con los textos
completos «Español»/«English», grupo único por instancia (`useId`), cambio inmediato, persistencia
en `ss.lang`, primera visita en español, sin auto-detección (FR-016). Solo cambia su **huella
visual**: se compacta con espaciado y disposición (reducción de gaps/padding y acomodo dentro del
clúster de controles de la cabecera), **sin píldoras «ES/EN», sin menú, sin banderas y sin siglas**
(FR-015, decisión Q5). El texto de cada idioma **no se reduce por debajo de la legibilidad heredada**
(FR-004: nada por debajo de `1rem` en el sitio público — la compacción es horizontal, no de cuerpo de
texto). Nombre accesible de cada opción = **nombre completo del idioma** (US4 esc. 4).

**Por qué**: minimiza el riesgo de regresión (las pruebas de F3 del selector pasan sin cambios: el
contrato semántico —roles, nombres accesibles, `ss.lang`— no cambia) y respeta literalmente la
decisión del cliente (solo compactar).

**Alternativas descartadas** (todas explícitamente rechazadas por el cliente en Q5): píldoras
«ES/EN», `select`/menú, banderas, siglas como identificación principal.

## R10-8 — Dónde vive el sistema visual: capa pública nueva; inventario compartido intacto

**Decisión**: las primitivas del sistema visual del sitio público viven en
**`frontend/src/components/publico/`** (p. ej. `PublicButton`, `SectionHeading`, `Card`, `Reveal` no
va aquí sino en `features/publico/motion/`… ver `plan.md`). El inventario compartido de
`frontend/src/components/` (`Button`, `Field`, `Notice`, `Dialog`…) **no se modifica**: lo siguen
usando el panel de F2/F3, que F10 no toca (FR-020). Regla de composición documentada: **el sitio
público usa la capa pública; el panel usa el inventario compartido; ninguna función tiene dos estilos
distintos dentro de cada ámbito** (FR-003/SC-011 — el ámbito del «sin dos estilos» es el **sitio
público**, según la spec).

**Por qué**: cualquier retoque visual de `Button`/`Notice`/… cambiaría el panel (violación directa de
FR-020/SC-008). La reconciliación de ambos catálogos queda para **F10B**, que sí rediseña el panel y
puede unificarlos.

**Alternativa descartada**: extender con variantes los componentes compartidos «sin tocar las
variantes existentes» — en la práctica cualquier cambio de clases base o de tokens compartidos altera
el panel; demasiado riesgo para un objetivo puramente visual.

## R10-9 — Contrato semántico congelado para garantizar cero regresión (SC-008)

**Decisión**: F10 congela todo lo que las pruebas de F2/F3 observan: **rutas** (`/`, `/health`,
`/panel/...`), **ids y anclas** (`#contenido`, `#quienes-somos`, `#horario`, `#whatsapp`, `#contacto`,
`#redes`), **roles y nombres accesibles** (nav `aria-label`, `legend` del selector, textos
«Español»/«English», `alt` del logo), **claves de `localStorage`** (`ss.lang`), **comportamiento**
(bilingüismo, fallback, borrador/publicado) y la **API** sin cambios. El rediseño se hace **por
dentro** de esos contratos (clases CSS, composición, tokens).

**Por qué**: SC-008 exige que el 100 % de las pruebas de F3 (y las de F2) pasen **sin cambios**. Con
este congelamiento, las suites existentes son exactamente la red de no-regresión de F10.

**Regla de escalado**: si alguna prueba existente codificara un detalle visual que la propuesta
elegida obligara a cambiar, **no se edita la prueba a espaldas del criterio** — se escala al humano
(SC-008 lo prohíbe implícitamente). La única prueba actual que mira clases es `BrandLogo.test.tsx`
(`object-contain`, sin `filter|rotate|transform|shadow|mix-blend`): esa regla del logotipo se
**conserva** (FR-002), no choca con el rediseño.

## R10-10 — Sin dependencias nuevas

**Decisión**: F10 añade **0 dependencias** (ni runtime ni dev). Todo se resuelve con React, CSS
(Tailwind 4) y las APIs del navegador (`matchMedia`, `IntersectionObserver`, `ResizeObserver`), ya
soportadas por el objetivo de navegadores. Constitución §II cumplida sin excepciones que justificar.

**Justificación negativa explícita** (por si la revisión pregunta): `framer-motion`/`GSAP`
(innecesarias para nivel moderado, ver R10-4), `next-themes` (la lógica de R10-3 es pequeña y propia),
`tailwindcss-animate` (sin valor), librerías de contraste/axe en runtime (la accesibilidad se audita
en QA, como en F3).

## R10-11 — Alcance del `data-theme`: global en el documento, efectivo solo en el sitio público

**Decisión**: `data-theme` se fija en `<html>` (necesario para que el anti-flash funcione antes de
React y para que `color-scheme` sea correcto), pero **solo el sitio público consume los tokens
semánticos**. El panel mantiene clases y colores explícitos de hoy → pinta igual esté el atributo que
esté. `ThemeProvider` y `ThemeToggle` se montan **solo** en `PublicLayout`: el panel no ofrece
control de tema (queda para F10B).

**Consecuencia conocida (frontera F10B, documentada)**: un visitante con tema oscuro que navegue a
`/panel` verá el panel en su aspecto actual (claro, sin cambios) — correcto según FR-020: el panel no
se toca. Cuando F10B rediseñe el panel, decidirá si adopta el tema.

## R10-12 — Verificación de «moderno», espacio del selector y contraste

**Decisión**: sin instrumentación (Out of Scope de la spec):
- **«Moderno» (SC-002)** y **3 propuestas (SC-013)**: revisión humana sobre la portada en claro y
  oscuro, comparada con el diseño previo y calibrada con las referencias (inspiración, no copia).
- **«Menos espacio» del selector (SC-007)**: comparación directa antes/después en la misma vista de
  referencia (captura de la cabecera con el selector actual de F3 vs. el nuevo); el e2e comprueba
  comportamiento y presencia, no píxeles (mediciones de ancho en e2e son frágiles: se dejan al
  checklist de `quickstart.md`).
- **Contraste AA en ambos temas (FR-024/SC-012)**: tabla de ratios por par de tokens (como la de
  `ux.md` §1.1 de F3) verificada en la fase de diseño y con checklist manual de QA por tema y estado
  (reposo/hover/foco/activo); el e2e cubre presencia, foco visible y `color-scheme`, no cálculo de
  contraste.
