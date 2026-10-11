# Research: F10 — Sistema de diseño y movimiento

**Feature**: F10 — Sistema de diseño y movimiento · **Branch**: `010-sistema-diseno-movimiento`
**Fase**: 0 de `/speckit.plan` (decisiones técnicas; numeración **R10-1…R10-16**, propia de F10)
**Fecha**: 2026-10-10 · **Revisión (puerta 2, 2026-10-10)**: el cliente rechazó las 3 propuestas por
«muy pobres, casi imperceptibles» y elevó el nivel de movimiento a **marcado/elegante con revelado al
hacer scroll** (FR-005). Se reevaluó la estrategia de movimiento completa —CSS puro / scroll-driven vs.
librería—: **R10-4** se reescribe, **R10-5** y **R10-10** se actualizan y se añade **R10-13**
(parámetros, cabecera al bajar, desplazamiento suave y rendimiento).
**Revisión 2 (cierre del plan, 2026-10-10)**: el humano aprobó la propuesta visual **B «Marea
Turquesa»** con **3 ajustes** que suman mecánica: navegación responsive con **menú hamburguesa
accesible en móvil** (sustituye en pantallas estrechas la restricción «sin menús ocultos»),
**mapa del sitio en el pie** y **espaciado vertical de sección**. Se actualiza **R10-6** (navegación
responsive) y se añaden **R10-14** (menú hamburguesa accesible), **R10-15** (mapa del sitio del pie) y
**R10-16** (tokens de espaciado de sección). Sin dependencias nuevas ni cambios de contrato.

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

## R10-4 — Movimiento: `IntersectionObserver` + CSS con stagger; **sin librería de animación** (reevaluado en la puerta 2)

> **Revisión 2026-10-10 (puerta 2)**: el cliente rechazó las 3 propuestas por «muy pobres, casi
> imperceptibles» y elevó el nivel a **marcado/elegante con revelado al hacer scroll** (FR-005). La
> estrategia se reevaluó completa —CSS puro / scroll-driven vs. librería— con ese nivel nuevo. El
> problema detectado eran los **parámetros** (duración y distancia casi imperceptibles), no la
> capacidad del mecanismo; aun así, la comparativa se rehace abajo de forma explícita. Los valores
> concretos (duraciones, curvas, distancia, stagger) viven en **R10-13**.

**Decisión**: el sistema de movimiento se construye con **transiciones y keyframes CSS** orquestados
por un **`IntersectionObserver` compartido** (componente `Reveal` + `RevealGroup` con entrada
escalonada por variable CSS), más los tokens de movimiento de R10-13. **Sin dependencias nuevas.**

### Comparativa con el nivel nuevo (marcado/elegante + revelado al scroll + stagger)

| Opción | ¿Logra revelado + stagger **perceptible**? | Soporte (2026) | Coste | Riesgo | Veredicto |
|---|---|---|---|---|---|
| **A. `IntersectionObserver` + CSS** (transiciones/keyframes + stagger por variable CSS) | **Sí, total**: entrada con opacidad 0→1 + `translateY` de 20–28 px durante ~600 ms, y escalonado de 80 ms entre componentes hermanos | 100 % (API del navegador, sin flags, en todo el objetivo) | **0 KB** de bundle; animación en el compositor | **Bajo**: determinista, parámetros fijos, verificable en Playwright (estilos computados) | **ELEGIDA** |
| **B. CSS scroll-driven animations** (`animation-timeline: view()/scroll()`) | Sí (incluso atadas al progreso del scroll), pero solo donde haya soporte | **No es Baseline**: Chrome/Edge 115+, Safari 17.4+, pero **Firefox estable aún no lo soporta por defecto** (solo con flag `layout.css.scroll-driven-animations.enabled`) → un 15–25 % de usuarios según audiencia sin el efecto | 0 KB | **Alto**: (1) en navegadores sin soporte el contenido puede quedarse clavado en el keyframe inicial (oculto) si se usa `fill-mode: both` sin guardia correcta, y `@supports (animation-timeline: scroll())` puede dar **falso positivo** en Firefox (acepta la sintaxis con el flag apagado); (2) `prefers-reduced-motion` **no** anula la línea de tiempo: hay que resetear `animation-timeline` a mano; (3) progreso dependiente de la posición de scroll → difícil de verificar de forma determinista en e2e; (4) peligros conocidos de repintado en iOS Safari | **Descartada como base** (y como mejora progresiva de F10: un efecto distinto por navegador contradice «0 rebotes/0 diferencias» de SC-014). Reevaluable en el futuro con guardias de `@supports` verificadas en Firefox real |
| **C. Librería de animación** (`motion` —antes `framer-motion`—, `GSAP` + ScrollTrigger) | Sí, con orquestación rica (stagger declarativo, springs, timelines, pin/scrub) | Total | `motion` ≈ 18–34 KB gzip (+ ≈ 15 KB API React); `GSAP` core ≈ 27 KB + ≈ 10 KB ScrollTrigger | **Medio**: JS en cada frame, gestión de ciclo de vida (limpieza de triggers, StrictMode doble-invocación), más superficie para descolocar el maquetado; su beneficio real es **layout animations, `AnimatePresence`, gestos y pin/scrub**, que F10 **no usa** (y el scrub/parallax queda desaconsejado por la propia spec: «nada mareante») | **Descartada** (ver R10-10). Candidata natural para F10B o si una sección futura necesita transiciones de layout compartido (`layoutId`), salidas animadas o timelines complejas — con justificación §II y aprobación humana |

**Por qué A alcanza el nivel marcado sin librería**: lo que el cliente percibió como «casi
imperceptible» se corrige con **parámetros**, no con más motor — distancia de entrada 20–28 px en vez
de «pocos píxeles», 600 ms en vez de 200–300 ms, stagger visible de 80 ms, curva con salida rápida y
asentamiento elegante — todo lo cual CSS resuelve nativamente sobre el compositor. Una librería
aportaría orquestación que este alcance no necesita (sin gestos, sin animación de layout, sin
scroll-scrub) a cambio de ~20–50 KB y JS por frame, en un sitio cuyo público incluye dispositivos
modestos y personas con poca experiencia (FR-007/SC-006). Constitución §II: **0 dependencias nuevas**
sigue siendo la decisión, ahora **tras reevaluar explícitamente la alternativa de librería** con el
nivel alto.

**Alternativas descartadas** (detalle en la tabla): (B) scroll-driven animations como mecanismo base
—soporte no Baseline con riesgo de contenido oculto y de `@supports` engañoso en Firefox—;
(C) `motion`/framer-motion y `GSAP`+ScrollTrigger —dependencia nueva + JS-driven para efectos que
CSS resuelve; sus ventajas reales quedan fuera de F10—; `tailwindcss-animate` (sin valor frente a los
tokens de movimiento propios); librerías pequeñas de «reveal on scroll» (envuelven exactamente el
patrón A con dependencia añadida).

**Rendimiento (FR-007/SC-006)**: solo `opacity`/`transform`, `will-change` temporal y medido, un solo
observer compartido, sin `filter`/`blur`, sin listeners de `scroll` — ver detalle en R10-13.

## R10-5 — `prefers-reduced-motion` incondicional + revelado progresivo a prueba de fallos

**Decisión (doble capa; incondicional incluso con el nivel marcado de FR-005)**:
1. **CSS**: bloque global `@media (prefers-reduced-motion: reduce)` que anula **por completo** el
   sistema de movimiento del sitio público: `transition-duration: 0s`, `animation-duration: 0s`,
   `animation-delay: 0s`, sin desplazamientos de entrada (`transform: none` en los estados de
   revelado) y `html { scroll-behavior: auto }` (el suave solo vive bajo `no-preference`). Con
   movimiento reducido hay **0 revelados y 0 transiciones de entrada perceptibles** (SC-005), sea
   cual sea el nivel de FR-005; también la cabecera cambia de estado al bajar **sin transición**.
2. **JS**: el hook `usePrefersReducedMotion` (matchMedia con listener en vivo — la preferencia puede
   cambiar con el sitio abierto sin recargar, edge case de la spec) hace que `Reveal` **nunca oculte**
   contenido cuando hay movimiento reducido: se renderiza ya visible.

**Diseño a prueba de fallos (FR-007/FR-009, edge cases de la spec)**:
- **Estado por defecto del contenido: visible**. El estado «pendiente» (opacidad 0 + desplazamiento
  de entrada) se aplica **solo si** se dan las tres condiciones: (a) JS montó `Reveal` (marcador `js`
  en `<html>`, puesto por el script inline anti-FOUC junto a `data-theme`), (b) el movimiento está
  permitido y (c) el elemento aún no se ha revelado. Si el JS falla o la animación no carga:
  contenido visible, nunca una zona en blanco. El estado pendiente usa **solo `opacity`/`transform`**
  —nunca `visibility: hidden`, `display: none` ni `content-visibility`— de modo que el contenido
  conserva su sitio, sigue siendo clicable (`pointer-events` intactos) y las aserciones de visibilidad
  de las suites de F3 pasan **sin cambios** (un `opacity: 0` sigue siendo «visible» para Playwright).
- **Salvavidas global**: un timeout de seguridad (≈ 1 s tras el montaje) que revela todo lo pendiente
  si el observer no llegó a disparar (JS a medias, observer roto); los elementos ya en pantalla al
  cargar se revelan de inmediato con su stagger corto —no esperan a un scroll.
- **Foco sobre contenido pendiente de revelado**: manejador `focusin` (y regla CSS equivalente) que
  hace visible el elemento **de inmediato**; el foco nunca queda sobre algo transparente. Igual al
  saltar a un ancla: el destino se revela al instante.
- **Sin repeticiones molestas**: el revelado se ejecuta **una sola vez por elemento y por carga**
  (registro de revelados persistente por elemento); cambiar de idioma o de tema re-renderiza sin
  re-montar (claves React estables, `key` por sección = ancla), de modo que no se re-animan secciones
  ya vistas ni hay «rebotes» ni parpadeos (US3 esc. 6, SC-014). La entrada de la portada (hero) solo
  ocurre en el primer montaje de cada carga real.

## R10-6 — Navegación: **responsive** — barra fija con `position: sticky` + `scroll-padding-top` medido + filas que envuelven (escritorio) / menú hamburguesa accesible (móvil)

> **Actualización 2026-10-10 (cierre del plan)**: el cliente, al aprobar el plan, decidió que en
> **pantallas estrechas** la navegación se colapse en un **menú hamburguesa accesible** (detalle en
> R10-14). La restricción «sin menús ocultos» (FR-010/FR-012, del encargo original) queda **solo para
> escritorio** (≥ 861 px), donde las filas de anclas siempre visibles siguen siendo la solución. La
> barra fija (FR-014) no cambia en ningún modo.

**Decisión**:
- **Fijeza (ambos modos)**: `position: sticky; top: 0` en el `<header>` del sitio público (se queda en
  el flujo del documento). Acompaña al visitante al bajar (caso (b) del «scroll incómodo», FR-014)
  **sin tapar contenido**: al estar en flujo, nunca se superpone al texto. La transición de superficie
  al bajar (`data-scrolled`, R10-13) es igual en escritorio y en móvil y no cambia la geometría.
- **Destino de anclas visible (ambos modos)**: `html { scroll-padding-top: var(--header-offset) }`,
  donde `--header-offset` se mide con `ResizeObserver` (altura real de la cabecera **con el menú
  cerrado**: el panel móvil es un overlay y no altera la medida; también cambia con el zoom al 200 %
  y el texto ampliado). Así, al saltar a una sección, su título queda por debajo de la barra, no
  debajo de ella (edge case «barra fija no tapa el destino»).
- **Breakpoint de cambio: 860 px** (móvil **≤ 860 px**; escritorio **≥ 861 px**): el mismo corte de
  la maqueta aprobada `propuestas/propuesta-b.html` y de `ux.md` (`TransitionHeader`, breakpoint 860)
  — diseño e implementación sin divergencias. **≥ 861 px (escritorio y tableta apaisada)**: la lista
  de anclas pasa de `overflow-x-auto` (la causa actual del problema) a **`flex-wrap` con filas
  compactas**: todos los enlaces visibles siempre, sin scroll lateral del menú ni de la página,
  objetivos ≥ 44 px. **≤ 860 px (móvil y tableta vertical)**: la cabecera compone fila de marca +
  controles (idioma, tema) + **botón hamburguesa** que abre el panel con las secciones (R10-14); sin
  scroll horizontal en ningún caso (FR-011). El corte es **CSS (media queries)**, no JS de resize.
- **Escala a F4–F9 (FR-012)**: los enlaces siguen siendo datos (`sections.ts` → modelo de navegación
  reutilizable); con ~8–10 secciones se acomodan en una o varias filas compactas en escritorio y en la
  lista del panel en móvil. Las rutas futuras F4–F9 ya cuelgan de `PublicLayout` (patrón de
  router de F3), así que la cabecera fija, el tema y el selector acompañan a las secciones nuevas sin
  cambios.

**Alternativas descartadas**:
- `position: fixed` + padding compensatorio en `body`: se superpone al contenido y acentúa los
  problemas de zoom 200 %/320 px (FR-025 exige que nada se tape).
- Menú hamburguesa **en escritorio** (≥ 861 px): innecesaria — caben filas compactas siempre
  visibles, que son más accesibles para el público del proyecto (FR-010). La hamburguesa **solo en
  móvil** se adopta por decisión del cliente (R10-14), no se descarta.
- Filas de anclas también **en móvil**: con 5 enlaces y controles aprietan la cabecera a 320 px y con
  la previsión F4–F9 (~8–10) invadirían varias filas la pantalla pequeña; el cliente eligió el menú.
- Carrusel horizontal de enlaces: es exactamente el caso (a) del «scroll incómodo».
- `scroll-margin-top` por sección: equivalente válido; se centraliza en `scroll-padding-top` (una sola
  fuente de la altura de cabecera).
- Librerías de menú/mega-menú o de focus-trap: 0 dependencias (R10-10); el patrón *disclosure* cabe
  en unas pocas líneas de React + CSS.

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

## R10-10 — Sin dependencias nuevas (reevaluado con el nivel marcado de la puerta 2)

**Decisión**: F10 añade **0 dependencias** (ni runtime ni dev). Todo se resuelve con React, CSS
(Tailwind 4 + tokens de movimiento) y las APIs del navegador (`matchMedia`, `IntersectionObserver`,
`ResizeObserver`), ya soportadas por el objetivo de navegadores. Constitución §II cumplida sin
excepciones que justificar.

**Justificación explícita tras la reevaluación** (nivel marcado de FR-005): se compararon de nuevo
`motion` (antes `framer-motion`; ≈ 18–34 KB gzip, + ≈ 15 KB con la API de React), `GSAP` +
ScrollTrigger (≈ 27 KB core + ≈ 10 KB del plugin) y las CSS scroll-driven animations (R10-4). Ninguna
aporta algo que este alcance necesite: F10 **no** incluye gestos, animación de layout compartido,
salidas animadas de componentes ni escenas con scroll crucificado (el scrub y el parallax quedan
además fuera por diseño: «nada mareante»). El nivel «marcado» se logra con **parámetros de CSS**
(R10-13), no con un motor más potente. Si en el futuro F10B o una sección F4–F9 necesitara
transiciones de layout (`layoutId`), `AnimatePresence` o timelines con ScrollTrigger, `motion` sería
la candidata recomendada — pero eso exige justificación §II y **aprobación humana explícita**.

**Justificación negativa del resto** (por si la revisión pregunta): `next-themes` (la lógica de R10-3
es pequeña y propia), `tailwindcss-animate` (sin valor frente a los tokens de movimiento propios),
librerías «reveal on scroll» (envuelven el patrón de R10-4 con dependencia añadida), librerías de
contraste/axe en runtime (la accesibilidad se audita en QA, como en F3).

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

## R10-13 — Parámetros de movimiento, cabecera al bajar y desplazamiento suave (nivel «marcado/elegante»)

**Decisión**: un conjunto único de **tokens de movimiento** en `frontend/src/index.css` (stitución que
F4–F9 hereda), con **suelo de perceptibilidad** y **techo de duración** (FR-005: «hasta ~0,8 s» y
«NUNCA casi imperceptibles»):

| Token | Valor | Uso |
|---|---|---|
| `--motion-duration-micro` | `180 ms` | hover/foco/activo de controles, enlaces y tarjetas |
| `--motion-duration-header` | `220 ms` | transición de la cabecera al bajar |
| `--motion-duration-reveal` | `600 ms` | revelado de secciones y bloques |
| `--motion-duration-hero` | `700 ms` | entrada de la portada (solo primera carga) |
| `--motion-distance` | `28 px` (`20 px` en pantallas ≤ 480 px) | desplazamiento de entrada |
| `--motion-stagger-step` | `80 ms` | escalonado entre componentes hermanos (tope 3 saltos = 240 máx.) |
| `--motion-ease-out` | `cubic-bezier(0.22, 1, 0.36, 1)` | entradas y revelados (salida rápida, asentamiento elegante) |
| `--motion-ease-standard` | `cubic-bezier(0.4, 0, 0.2, 1)` | microinteracciones y cabecera |

- **Suelo de perceptibilidad (anti «casi imperceptible»)**: todo revelado o transición de entrada dura
  **≥ 400 ms** y recorre **≥ 16 px** (o pasa de opacidad 0 a 1 de forma evidente); techo **≤ 800 ms**
  por elemento y **≤ 800 ms** la secuencia total de un bloque (stagger incluido): se percibe con
  claridad, pero nunca se convierte en espera ni en «secuencia larga» (SC-014).
- **Revelado de secciones y componentes**: al entrar en pantalla (`IntersectionObserver` con
  `rootMargin: 0px 0px -10% 0px`, un solo observer compartido), el bloque entra con fade +
  `translateY(var(--motion-distance))` y sus componentes hijos **escalonados** (`RevealGroup` +
  `transition-delay: calc(var(--reveal-index) * var(--motion-stagger-step))`, tope de 3 saltos). Una
  sola vez por elemento y por carga (R10-5).
- **Transiciones entre secciones**: al navegar por las anclas, el desplazamiento suave + el revelado
  del destino forman la transición; el encabezado de sección anima su acento (`transform: scaleX` de
  0→1, 500 ms) junto al título, y las tarjetas de una fila entran escalonadas. Sin cross-fade de
  página ni orquestación entre secciones.
- **Cabecera con transición al bajar (FR-014/FR-005)**: `SiteHeader` gana el estado `data-scrolled`
  cuando la página se desplaza, detectado con una **sentinela** `<div>` al inicio del documento
  observada por `IntersectionObserver` (sin listener de `scroll` en el hilo principal). La transición
  es solo de **superficie** —`background-color` de translúcido a sólido (`--surface`), `border-color`
  hairline y `box-shadow` suave, 220 ms— **sin cambio de geometría** (misma altura): cero riesgo de
  descolocar anclas o de saltos de maquetado. Con movimiento reducido, el cambio de estado es
  **instantáneo**.
- **Desplazamiento suave a las anclas**: `html { scroll-behavior: smooth }` **solo** bajo
  `@media (prefers-reduced-motion: no-preference)`; con movimiento reducido, `auto` (salto
  instantáneo). El destino sigue protegido por `scroll-padding-top` medido (R10-6): la transición
  nunca tapa el título de la sección destino.
- **Microinteracciones más marcadas (FR-008)**: 180 ms en `hover`/`:focus-visible`/`:active` con
  cambios de color/borde/sombra y `transform: translateY(-2px)` / `scale(1.02)` como máximo en
  tarjetas y botones (nunca propiedades de layout → sin desplazamientos del maquetado). El logotipo
  **queda excluido** de todo movimiento (FR-002).
- **Rendimiento (FR-007/SC-006)**: solo `opacity` y `transform` (compositor, sin reflow); **sin
  `filter`/`blur`** (no son del compositor y pueden marear); `will-change` **temporal y medido** —se
  añade al empezar la animación y se retira en `transitionend`— (nada de `will-change` permanente en
  decenas de nodos; se verifica en DevTools → Layers/Performance); un `IntersectionObserver`
  compartido (no uno por elemento); sin listeners de `scroll`; contenido visible e interactivo desde
  el primer instante y objetivo **0 tareas largas > 50 ms** atribuibles al movimiento en el
  dispositivo de referencia más modesto (checklist de `quickstart.md`). Si el movimiento degrada la
  experiencia, se reduce o se elimina antes que entorpecer la lectura (FR-007).

**Alternativas descartadas**: scroll-scrubbed/parallax (la spec lo desaconseja: «nada mareante»);
`filter: blur()` en las entradas (coste de pintura y mareo potencial); cabecera que **reduce su
altura** al bajar (riesgo de saltos en anclas y de `scroll-padding-top` inestable; la transición de
superficie ya es claramente perceptible); transiciones de entrada por encima de 800 ms o con
rebote/curvas elásticas tipo `back.out` («nada estridente ni infantil»).

## R10-14 — Menú hamburguesa accesible en móvil (patrón *disclosure*) — ajuste del cliente 2026-10-10

**Decisión**: en pantallas **≤ 860 px** (breakpoint 860 — ver R10-6, alineado con `ux.md` y la
maqueta `propuesta-b.html`), la lista de anclas se colapsa tras un **botón hamburguesa** con el
patrón de **divulgación** de las guías APG (WAI-ARIA Authoring Practices): un *disclosure* con
contención de foco, **no un diálogo modal** — sin `aria-modal`, sin librerías de *focus-trap*, sin
gestos.

- **Botón**: `<button>` real (no `div`, no `summary`), **nombre accesible localizado**
  (`nav.menu.toggle`: «Menú de secciones» / «Section menu»), **`aria-expanded`** sincronizado con el
  estado y **`aria-controls`** con el id del panel; área de toque **≥ 44 px**; siempre visible en la
  cabecera (es lo que mantiene FR-014 «alcanzable en todo momento» en móvil: los enlaces están a **un
  toque**).
- **Panel**: la lista de secciones **del mismo modelo `sections.ts`** (mismas claves i18n → mismos
  nombres accesibles que en escritorio; R10-15 comparte la fuente), con superficie sólida por tokens
  (`--surface`, correcta en claro y oscuro), altura máxima `calc(100dvh - var(--header-offset))` con
  **scroll interno** si la lista es larga (F4–F9), y **control visible de cierre**.
- **Gestión de foco** (el punto más delicado; RG10-17 — igual que `ux.md` §2.3/§6): al abrir, el
  foco **entra en el panel** (primer enlace de la lista; el control visible de cierre está
  presente); **`Escape`**, el **control visible** de cierre, el **toque/clic fuera** (scrim) y el
  propio botón lo cierran y **devuelven el foco al botón**; activar una sección cierra el panel y el
  foco continúa en el destino del ancla; mientras está abierto, `Tab`/`Shift+Tab` quedan
  **contenidos en el panel y su botón** (el fondo va con scrim y sin scroll: no hay nada operable
  detrás). El foco nunca queda sobre contenido oculto (coherente con R10-5).
- **Sin scroll de fondo**: `html[data-nav-open='true'] { overflow: hidden }` mientras el panel está
  abierto (rueda y teclado); se restaura al cerrar (RG10-19; prueba en iOS/Android en QA).
- **Sin gestos desconocidos**: solo clic/toque y teclado (`Enter`/`Espacio`); ni swipe ni
  deslizamientos (hereda la restricción del encargo).
- **Movimiento**: el panel entra/sale con fade corto (180 ms, solo `opacity`/`transform`, sin animar
  altura ni layout) y con `prefers-reduced-motion` la apertura/cierre es **instantánea** (FR-006
  incondicional).
- **Cruce de breakpoint** (rotar/girar el dispositivo): el panel se cierra solo y `aria-expanded`
  vuelve a `false`; sin estado residual ni overlays colgados.
- **Fallback sin JS**: el marcador `js` de `index.html` gobierna el colapso: **sin JS los enlaces se
  pintan como filas compactas visibles** a cualquier ancho (y el botón no se renderiza/queda oculto):
  la navegación nunca depende del script (misma filosofía de «contenido visible sin JS» de R10-5).
  Con JS, el colapso lo resuelve **CSS puro** bajo la media query + `data-nav-open`, sin quitar los
  enlaces del DOM.

**Compatibilidad con las pruebas congeladas de F3 (R10-9/SC-008)** — restricciones verificadas
leyendo las suites, que **fijan** el diseño:
1. `portada-accesibilidad.spec.ts` exige el landmark `navigation` **visible** a 320 px y a 160 px
   (zoom 200 %) → el `<nav aria-label="Secciones de la portada">` **nunca se oculta**: el botón va
   dentro del `<nav>` y solo la lista colapsa.
2. Las unitarias de jsdom (`homePage.test.tsx`) consultan los enlaces del `<nav>` → los enlaces
   **permanecen en el DOM** y el colapso es CSS (sin `hidden`/`aria-hidden`/`inert` de React; en
   jsdom no se aplican media queries → las consultas siguen pasando sin cambios).
3. La travesía por teclado congelada a 320 px exige que el ciclo contenga las 5 etiquetas de sección
   y cierre en **< 30 pasos** (`tabCycle`) → con el menú cerrado, esas etiquetas las aporta el
   **mapa del sitio del pie** (mismas claves i18n) y el saldo de paradas de tabulación es mínimo
   (botón hamburguesa y control de tema se suman; las anclas de cabecera dejan de ser tabulables en
   móvil) — recuento verificado en implementación (RG10-21).
4. La auditoría de objetivos ≥ 44 px a 320 px incluye `a[href], button` visibles → botón y panel se
   diseñan a ese umbral.

**Alternativas descartadas**:
- `<details>`/`<summary>` nativo: funcionaría sin JS, pero el control de foco/Escape es menos
  predecible entre navegadores y el estado `aria-expanded` explícito se vuelve opaco; con el fallback
  `js` se gana lo mismo con menos rarezas.
- «Checkbox hack» (`:checked`) o `:target`: comportamiento de foco errático y nombres accesibles
  frágiles.
- Diálogo modal a pantalla completa (`<dialog>`, focus-trap): más peso del necesario para una lista
  de 5–10 enlaces; el *disclosure* cumple con menos riesgo (aunque conserva el bloqueo de scroll).
- Renderizar los enlaces condicionalmente según estado (React): rompería las unitarias congeladas y
  el fallback sin JS (RG10-18).
- Librerías de menú o `focus-trap`: 0 dependencias (R10-10).

## R10-15 — Mapa del sitio en el pie — ajuste del cliente 2026-10-10

**Decisión**: nuevo componente **`FooterSiteMap`** en
`frontend/src/features/publico/components/FooterSiteMap.tsx`, montado dentro de `PublicFooter` (el
`<footer>`/landmark `contentinfo` de F3 se conserva intacto: el mapa vive **dentro** de él).

- **Fuente única de verdad**: el **mismo modelo de secciones** que la navegación
  (`features/publico/sections.ts`: `PUBLIC_SECTIONS`, `presentSections(data)`); `PublicLayout` pasa
  las secciones presentes a `PublicFooter`. Mismas claves i18n (`nav.sections.*`) → **mismos nombres
  de enlace** que la cabecera: nav y pie nunca divergen (ni en nombres ni en reglas de borrador), y
  las etiquetas congeladas de la travesía por teclado se cumplen por construcción (ver R10-14).
- **Contenido**: las secciones de la portada de hoy (`quienes-somos`, `horario`, `whatsapp`,
  `contacto`, `redes`); opcionalmente un enlace a la portada (`/`) si `ux.md` lo decide.
- **Regla «sin enlaces rotos»**: **solo se renderizan destinos existentes** — secciones con contenido
  publicado (mismo criterio borrador/publicado de F3: una sección ausente no deja enlace muerto) y,
  para **F4–F9**, entradas del modelo con `href` de **ruta ya implementada**. El modelo se extiende
  entonces con `href` (hoy solo `anchor`) y el componente **no cambia**: nunca se enlaza a rutas
  futuras inexistentes.
- **Disposición**: **columnas en escritorio** (grid `auto-fit`/2–3 columnas) y **apilado** en móvil;
  tokens semánticos y de espaciado (R10-16).
- **Accesibilidad (FR-013/FR-022)**: landmark `<nav>` con **nombre accesible propio y distinto** del
  de la cabecera (`footer.sitemap.label`, es/en) — dos landmarks `navigation` con el mismo nombre
  romperían tanto la accesibilidad como las consultas congeladas —, lista `<ul>` real de enlaces
  `<a href="#…">`, foco visible, objetivos ≥ 44 px y contraste AA en ambos temas (el e2e de contraste
  congelado ya audita `footer :is(a,p,span)`).

**Alternativas descartadas**:
- Pie con menús desplegados/megamenu o «nube de etiquetas»: más oculto y menos familiar para el
  público del proyecto.
- Modelo de datos del pie **aparte** (un `footerSections.ts`): duplicaría la fuente de secciones →
  deriva garantizada (una sección se publica y solo aparece en un sitio).
- Enlazar ya las rutas F4–F9 previstas: **enlaces rotos** hoy; la regla es «solo destinos
  existentes».
- Componente en `frontend/src/components/` (inventario compartido): el pie es del sitio público;
  tocar el inventario compartido tocaría el panel (R10-8/FR-020).

## R10-16 — Tokens de espaciado vertical de sección — ajuste del cliente 2026-10-10

**Decisión**: el «más aire» entre secciones pedido por el cliente se expresa como **tokens de
espaciado vertical** en `frontend/src/index.css` (capa aditiva, junto a los semánticos de tema),
registrados en `@theme inline` para generar utilidades (`py-section`, `py-section-tight`,
`gap-block`) heredables por F4–F9:

| Token | < 480 px | 480–767 px | 768–1279 px | ≥ 1280 px | Uso |
|---|---|---|---|---|---|
| `--space-section` | `3.25rem` | `4rem` | `5.5rem` | `7rem` | padding vertical de cada sección |
| `--space-section-tight` | `2rem` | `2.5rem` | `3.5rem` | `4.5rem` | secciones densas o encadenadas del mismo bloque |
| `--space-block` | `1.5rem` | `2rem` | `2.5rem` | `3rem` | separación entre el encabezado de sección y su contenido |

- **Valores por breakpoint** (no `clamp()` continuo): predecibles de verificar y de ajustar en un solo
  sitio. Son **defaults revisables**: la calibración final se cierra en QA con la propuesta B
  (checklist `quickstart.md`, SC-002/SC-011); el e2e **no** mide píxeles de espaciado (R10-12).
- **Regla de composición**: toda sección de la portada y del futuro F4–F9 usa estos tokens; ninguna
  define su propio padding vertical. El aislamiento del panel se mantiene (R10-2/R10-8).

**Alternativas descartadas**:
- `padding`/`margin` arbitrarios por componente (`py-16`, `py-24`…): deriva de espaciado imposible de
  recalibrar de una vez y sin valor heredable.
- `clamp()` continuo sobre el ancho: elegante, pero dificulta la verificación y el ajuste discreto por
  breakpoint.
- Espaciado por márgenes colapsables entre secciones: impredecible con los revelados de R10-4
  (bordes/transform) — el padding por token es determinista.
