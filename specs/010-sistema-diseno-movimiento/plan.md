# Implementation Plan: Sistema de diseño y movimiento (F10)

**Feature Branch**: `010-sistema-diseno-movimiento` · **Fecha**: 2026-10-10 · **Revisión 2** (cierre
del plan): propuesta visual **B «Marea Turquesa» APROBADA** con 3 ajustes del cliente
(navegación responsive con menú hamburguesa accesible en móvil, mapa del sitio en el pie y espaciado
vertical de sección) → decisiones **P10-13/P10-14/P10-15** (research **R10-14/R10-15/R10-16**).
**Spec**: `spec.md` — **APROBADA** (2026-10-10, con ajuste de puerta 1: panel → F10B; modo oscuro con
interruptor manual; **iteración de la puerta 2**: nivel de movimiento → **marcado/elegante con
revelado al hacer scroll**, FR-005/SC-014)

> **Relación con `ux.md` (fase de diseño, `disenador-ux`, en paralelo)**: la spec (FR-026/SC-013)
> exige **3 propuestas visuales distintas** aplicadas al sitio público y al modo oscuro, y el humano
> aprueba una **antes de escribir código**. **Resuelto en la revisión 2 (2026-10-10)**: el humano
> aprobó la **propuesta B «Marea Turquesa»** con **3 ajustes** que amplían la mecánica de este plan
> —navegación responsive (barra fija en escritorio + **menú hamburguesa accesible en móvil**),
> **mapa del sitio en el pie** y **espaciado vertical de sección** (P10-13/P10-14/P10-15)—. La
> **dirección visual** la fija `ux.md`; este plan fija la **mecánica** —tokens, tema, movimiento,
> navegación, espaciado, formato del selector—. La propuesta elegida puede cambiar composición,
> acentos, iconografía e ilustraciones **sin alterar** ninguna decisión de este plan (ver R10-4/R10-13
> y la sección «Frontera con el diseño»); los 3 ajustes del cliente **sí** cambian la mecánica y ya
> quedan incorporados aquí.

---

## Summary

F10 aplica un **sistema visual moderno y un sistema de movimiento marcado/elegante con revelado al
hacer scroll** (nivel elevado en la puerta 2 tras el rechazo de las propuestas «casi imperceptibles»:
secciones y componentes que se revelan al entrar en pantalla con entrada escalonada, transiciones
perceptibles y cabecera que transiciona al bajar — R10-4/R10-13) al sitio público
existente (la portada de F3, con todas sus secciones), añade **modo oscuro con interruptor manual**
(respetando por defecto `prefers-color-scheme`, con persistencia entre visitas), rediseña la
**navegación de forma responsive: barra fija siempre visible en escritorio (≥ 861 px) y menú
hamburguesa accesible en móvil (≤ 860 px, patrón *disclosure* con gestión de foco)** sin scroll
incómodo (ni lateral de enlaces ni pérdida al bajar), añade un **mapa del sitio en el pie**
(secciones de la portada de hoy, preparado para F4–F9 sin enlaces rotos), da **más aire entre
secciones** con tokens de espaciado vertical responsive y **compacta el selector de idioma** sin
tocar su comportamiento. Entrega además los **cimientos reutilizables** (tokens con ambos temas y de
espaciado, primitivas de componente, patrones de movimiento, patrón de navegación responsive y
formato del selector) para que F4–F9 nazcan con el sistema.

Es un cambio **exclusivamente de frontend y de presentación**: **0 cambios** en backend, datos,
migraciones, contratos, rutas y URLs (`data-model.md`, `contracts/README.md`). El panel de
administración **no se toca** (F10B). El contrato semántico observable (ids, anclas, roles, nombres
accesibles, `localStorage['ss.lang']`, comportamiento bilingüe) queda **congelado** (R10-9) para que
las suites de F2/F3 pasen **sin cambios** (SC-008).

**Dependencias nuevas: NINGUNA** (R10-10): todo se resuelve con React, Tailwind 4 y las APIs del
navegador (`matchMedia`, `IntersectionObserver`, `ResizeObserver`). La reevaluación con el nivel de
movimiento marcado (puerta 2) comparó `motion`/`GSAP` y las CSS scroll-driven animations y
**confirmó** la decisión sin dependencias (comparativa en R10-4).

## Technical Context

| Aspecto | Decisión |
|---|---|
| Frontend | React 19 + TypeScript `strict` + Vite + **Tailwind 4** (tokens en `frontend/src/index.css`) |
| Routing | `react-router-dom` 7 — **rutas sin cambios** (`/`, `/health`, `/panel/...`) |
| Datos | `@tanstack/react-query` — mismas queries de F3 (`usePublicHomeData`, `getPortada(lang)`) |
| i18n | Propio de F3 (`features/publico/i18n/`); solo se añaden claves nuevas (es/en) |
| Tema | `data-theme` en `<html>` + variables semánticas + variante `dark` de Tailwind 4 (R10-1) |
| Movimiento | **CSS (transiciones/keyframes + tokens de movimiento) + `IntersectionObserver` compartido con entrada escalonada**; sin librería (R10-4/R10-13; reevaluado en la puerta 2 con el nivel marcado) |
| Navegación y pie | Responsive con breakpoint **860 px** (móvil ≤ 860 / escritorio ≥ 861 — el corte de la maqueta aprobada `propuesta-b.html` y `ux.md`): barra fija con anclas en filas `flex-wrap` en escritorio / **menú hamburguesa accesible** (patrón *disclosure*) en móvil (P10-6/P10-13, R10-6/R10-14); **mapa del sitio en el pie** sobre el mismo modelo de secciones (P10-14, R10-15) |
| Espaciado | Tokens de espaciado vertical de sección (`--space-section*`) con valores por breakpoint en `index.css` (P10-15, R10-16) |
| Pruebas | Vitest + Testing Library (unitarias); Playwright (e2e local, `frontend/e2e/`) |
| Backend/BD/API | **Sin cambios** (F10 es de presentación; FR-020/FR-021) |
| Panel admin | **Sin cambios** (fuera de alcance: F10B) |

## Constitution Check

| Principio | Cumplimiento |
|---|---|
| §I La especificación manda | Plan fiel a la spec aprobada (FR-001…FR-026); huecos = 0 (todas las preguntas resueltas 2026-10-10). Los defaults revisables se marcan donde la spec los deja abiertos. Los **3 ajustes del cliente al aprobar el plan** (2026-10-10) se incorporan como decisiones P10-13…P10-15; el de navegación **sustituye en pantallas estrechas** la restricción «sin menús ocultos» de FR-010/FR-012 (decisión humana posterior y de mayor rango; ver nota en «Cobertura de requisitos»). |
| §II Arquitectura / dependencias | Solo frontend; **0 dependencias nuevas** (justificación negativa en R10-10, **reevaluada** con el nivel marcado de movimiento en R10-4: `motion`/`GSAP`/scroll-driven comparadas y descartadas). Sin cambios de contrato → `contracts/README.md` lo deja constatado («el contrato se escribe antes que el código» se cumple sin necesidad de delta). |
| §III Pruebas (test-first) | Cada tarea traerá sus pruebas: unitarias (Vitest) de tema, movimiento, navegación (incluido el menú hamburguesa móvil y el mapa del sitio del pie) y selector; e2e (Playwright) de navegación fija y responsive, menú móvil accesible, tema, movimiento reducido y selector; no-regresión con las suites de F2/F3 sin tocar. |
| §IV Seguridad | Sin superficie nueva (sin endpoints, sin formularios, sin `dangerouslySetInnerHTML`); `localStorage` solo guarda preferencias de presentación (`ss.lang`, `ss.theme`); sin secretos. |
| §V Calidad | TypeScript `strict`, sin `any`; funciones pequeñas; nombres en inglés en código. |
| §VI–VII BD / operación | No aplican (sin backend ni datos). |
| §VIII Gobierno | Este plan **requiere aprobación humana** antes de implementar; adicionalmente, FR-026 exige aprobar la propuesta visual (de `ux.md`) **antes de escribir código**. |

## Frontera con F10B (panel de administración)

- F10 **no toca el panel en absoluto** (FR-020, SC-008): ni `features/panel`, `features/usuarios`,
  `features/roles`, `features/auditoria`, `features/auth`, `features/informacion`, ni los componentes
  compartidos de `frontend/src/components/` (R10-8), ni el layout del panel.
- El sistema visual público vive en una capa nueva (`components/publico/`) precisamente para no
  alterar el inventario compartido que pinta el panel.
- El panel **no muestra** control de tema ni adopta el tema oscuro: un visitante con tema oscuro que
  abra `/panel` ve el panel con su aspecto actual (consecuencia conocida, R10-11). La reconciliación
  visual de ambos catálogos es tarea de **F10B**.
- Verificación: las suites e2e de panel (`portada-panel.spec.ts`, `acceso.spec.ts`,
  `auditoria.spec.ts`) y las unitarias de panel pasan **sin cambios**; revisión visual del panel en
  QA como comprobación extra.

## Frontera con el diseño (`ux.md`, FR-026)

**Propuesta aprobada (revisión 2, 2026-10-10): B «Marea Turquesa»**, con 3 ajustes del cliente que
**amplían la mecánica** de este plan y ya están incorporados en ella: navegación responsive con menú
hamburguesa accesible en móvil (P10-13/R10-14), mapa del sitio en el pie (P10-14/R10-15) y tokens de
espaciado vertical de sección (P10-15/R10-16). `ux.md` ajusta su propuesta B a esa mecánica (el pie
con mapa del sitio y el aire entre secciones forman parte de la dirección visual).

Mecánica **fija por este plan**:
tokens semánticos con dos temas (R10-1/R10-2), resolución y persistencia del tema (R10-3), reglas de
movimiento marcado/elegante con revelado al scroll y `prefers-reduced-motion` (R10-4/R10-5/R10-13),
navegación responsive —barra fija con `scroll-padding-top` medido y anclas en filas en escritorio;
menú hamburguesa accesible en móvil— (R10-6/R10-14), mapa del sitio en el pie alimentado por el
modelo de secciones (R10-15), espaciado de sección con tokens responsive (R10-16), selector con
radios nativos y textos completos (R10-7).

Lo que **sí decide `ux.md`**: la dirección visual concreta de la propuesta aprobada (composición de
la portada, tratamiento tipográfico fino, dosis de color, iconografía/ilustraciones nuevas —permitidas
por Q6—, forma exacta del control de tema y la disposición compacta final del selector, ubicación de
secciones si se reordena la portada). Restricción dura: el Manual de Identidad manda sobre las
tendencias (FR-001) y el logotipo nunca se recolorea ni se deforma (FR-002, también en oscuro: FR-024).

## Decisiones técnicas (resumen; análisis completo en `research.md`)

Numeración **P10-1…P10-15** (de plan) con su **R10-x** de `research.md` (R10-1…R10-16). Las
decisiones **P10-13/P10-14/P10-15** nacen de los 3 ajustes del cliente al aprobar el plan
(2026-10-10).

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| **P10-1** (R10-1) | Tema como **`data-theme` en `<html>`** + variables CSS **semánticas** (`--surface`, `--text`, `--border`, `--accent`…) registradas con `@theme inline` + `@custom-variant dark` de Tailwind 4 | El tema se resuelve en una capa de tokens: la mayoría de componentes ni menciona el tema; atributo autoexplicativo y extensible | `dark:` solo con `prefers-color-scheme` (sin interruptor); clase `.dark`; librería de temas; re-tematizar los tokens de F3 |
| **P10-2** (R10-2) | Tokens de F3 (`--color-navy`…) **intactos**; capa semántica **aditiva**; sitio público usa utilidades semánticas | Garantiza por construcción que el panel no cambia de aspecto (FR-020) | Re-tematizar tokens existentes (recolorearía el panel); dos hojas de estilo por tema |
| **P10-3** (R10-3) | **`ThemeProvider`** (patrón `LanguageProvider`): `localStorage['ss.theme']` (`'light'/'dark'`, ausente = sistema), `matchMedia` **en vivo**, script **inline anti-FOUC** en `index.html` | FR-023 completo: sistema por defecto, manual persistente, cambio en caliente, sin parpadeos al cargar | Resolver el tema solo en React (flash); cookie; `sessionStorage`; guardar `'system'` como valor |
| **P10-4** (R10-4) | **Movimiento CSS + `IntersectionObserver` compartido** (`Reveal`/`RevealGroup` con entrada escalonada); solo `opacity`/`transform` y colores | Nivel **marcado/elegante** (FR-005) sin librería: revelado + stagger claramente perceptibles con **0 dependencias** y sin riesgo de descolocar el maquetado | `motion`/`GSAP` (dependencias de ~20–50 KB, JS por frame); CSS scroll-driven animations como base (no Baseline: Firefox sin soporte por defecto, riesgo de contenido clavado en el keyframe inicial) |
| **P10-12** (R10-13) | **Tokens de movimiento** con suelo de perceptibilidad (≥ 400 ms, ≥ 16 px) y techo (≤ 800 ms, stagger ≤ 240 ms) + **cabecera con transición de superficie al bajar** (sentinela + `data-scrolled`, sin cambio de geometría) + `scroll-behavior: smooth` solo bajo `no-preference` | Corrige el «casi imperceptible» rechazado sin caer en lo estridente: parámetros únicos y verificables (e2e + revisión humana SC-014), heredables por F4–F9 | Ajustar valores «a ojo» por componente (deriva); cabecera que reduce altura al bajar (saltos de anclas); `blur`/parallax (mareo y coste de pintura) |
| **P10-5** (R10-5) | `prefers-reduced-motion` en **doble capa** (CSS global + hook `usePrefersReducedMotion` en vivo); contenido **visible por defecto**, revelado progresivo; `focusin` revela al instante | FR-006 incondicional y edge cases: JS caído → contenido visible; foco nunca sobre algo oculto | Ocultar con CSS y confiar en que JS llega (zona en blanco si falla) |
| **P10-6** (R10-6) | Navegación **en escritorio (≥ 861 px)**: **`position: sticky`** en la cabecera + `html { scroll-padding-top: var(--header-offset) }` medido con `ResizeObserver` + anclas en **`flex-wrap`** (fin del `overflow-x-auto`); en móvil (≤ 860 px) el patrón es el menú hamburguesa de **P10-13** | Elimina los dos casos del «scroll incómodo» (Q2) sin tapar contenido ni romper anclas; escala a F4–F9 en filas compactas | `fixed` + padding (se superpone); hamburguesa **en escritorio** (innecesaria: caben filas compactas); carrusel de enlaces (es el caso (a)) |
| **P10-7** (R10-7) | Selector: **solo formato compacto** (espaciado/disposición); radios nativos, textos completos, `useId`, `ss.lang` intactos | Decisión Q5 literal + SC-008: las pruebas de F3 pasan sin cambios | Píldoras «ES/EN», menú, banderas (rechazados por el cliente) |
| **P10-8** (R10-8) | Sistema visual público en **`frontend/src/components/publico/`**; inventario compartido **sin tocar** | El panel pinta con el inventario compartido: no se le puede cambiar un solo estilo (FR-020) | Extender variantes de los componentes compartidos (alteraría el panel) |
| **P10-9** (R10-9) | **Contrato semántico congelado**: rutas, ids/anclas, roles y nombres accesibles, `ss.lang`, comportamiento; el rediseño va «por dentro» | SC-008 exige las suites de F3 **sin cambios**: ellas son la red de no-regresión | Cambiar ids/roles y «arreglar» las pruebas (viola SC-008) |
| **P10-10** (R10-12) | Verificación de «moderno», de «menos espacio» (SC-007) y de contraste AA por **revisión humana + checklist** (`quickstart.md`), con e2e para comportamiento | La spec descarta instrumentación; las mediciones de píxeles en e2e son frágiles | e2e que mida anchos de píxel del selector; axe en runtime (fuera del patrón de F3) |
| **P10-11** (R10-11) | `data-theme` en `<html>` (anti-FOUC) pero efectivo solo en el sitio público; `ThemeProvider`/`ThemeToggle` solo en `PublicLayout` | Mantiene el panel inalterado aunque el atributo sea global | Tema por contexto React sin tocar `<html>` (no funciona el anti-FOUC) |
| **P10-13** (R10-14) | **Navegación responsive**: escritorio (≥ 861 px) = barra fija de P10-6 con anclas siempre visibles; móvil (≤ 860 px) = **menú hamburguesa accesible** (patrón *disclosure*: botón con nombre accesible, `aria-expanded`/`aria-controls`, ≥ 44 px; panel con las secciones; foco al primer enlace al abrir; Escape, control visible y toque fuera cierran con foco devuelto al botón; Tab contenido en el panel; sin scroll de fondo; sin gestos) | Decisión del cliente al aprobar el plan (2026-10-10): en pantallas estrechas un menú accesible a un toque da más aire y comodidad que filas de enlaces, sin perder «siempre visible y a un toque» (FR-014) ni accesibilidad (FR-013) | Filas de anclas también en móvil (estorbosas con 8–10 enlaces F4–F9); hamburguesa en escritorio (innecesaria y más oculta que las filas); librerías de menú/focus-trap (0 dependencias, R10-10) |
| **P10-14** (R10-15) | **Mapa del sitio en el pie**: `FooterSiteMap` dentro de `PublicFooter`, alimentado por el mismo modelo de secciones (`sections.ts`/`presentSections`) que la navegación; columnas en escritorio, apilado en móvil; solo enlaza destinos existentes | Índice visible y familiar del sitio público que refuerza FR-010/FR-013 y hereda de una sola fuente de verdad (nav y pie nunca divergen, ni en nombres ni en borradores) | Pie con menús desplegados/megamenu; mapa en JSON aparte (duplicaría el modelo de secciones); enlaces a rutas F4–F9 aún inexistentes (enlaces rotos) |
| **P10-15** (R10-16) | **Tokens de espaciado vertical de sección** (`--space-section` y acompañantes con valores responsive por breakpoint) aplicados por las secciones de la portada | «Más aire» entre secciones pedido por el cliente con un único control calibrable (no `py-*` sueltos por componente) y heredable por F4–F9 | Márgenes arbitrarios por sección (deriva de espaciado); `clamp()` continuo (valores por breakpoint son más predecibles de verificar) |

## Estrategia de tema (detalle de implementación)

1. **`frontend/index.html`**: script inline (~10 líneas, sin dependencias) antes de las hojas de
   estilo: lee `localStorage['ss.theme']`; si no es `'light'|'dark'`, usa
   `matchMedia('(prefers-color-scheme: dark)')`; fija `document.documentElement.dataset.theme` y
   `style.colorScheme`. Idempotente y silencioso (try/catch si `localStorage` falla).
2. **`frontend/src/index.css`**:
   - `@custom-variant dark (&:where([data-theme='dark'], [data-theme='dark'] *));`
   - Bloque de tokens semánticos: valores por tema bajo `html[data-theme='light']` /
     `html[data-theme='dark']` (superficies, texto, texto atenuado, bordes, acento, foco, sombra),
     derivados de la paleta de marca (el oscuro se deriva de `navy`/`cream`, sin romper la proporción
     70/20/10 ni las reglas del logotipo: FR-024) y registrados en `@theme inline` para las
     utilidades (`bg-surface`, `text-muted`…).
   - `html { scroll-padding-top: var(--header-offset, 0px); scroll-behavior: smooth }` (este último
     solo bajo `prefers-reduced-motion: no-preference`; con movimiento reducido, `auto` — R10-13).
   - Bloque global `@media (prefers-reduced-motion: reduce)` que anula **todo** el sistema de
     movimiento (entradas, revelados, stagger, transición de cabecera y desplazamiento suave:
     `transition-duration: 0s`, `animation-delay: 0s`, `transform: none` en estados de revelado —
     FR-006/SC-005, R10-5).
3. **`frontend/src/features/publico/theme/`**: `ThemeProvider.tsx` (estado del tema efectivo; aplica
   `data-theme`/`color-scheme` en `<html>`; listener `change` de `matchMedia` cuando no hay elección
   manual), `useTheme.ts` (contexto: `theme` efectivo, `preference` guardada, `setTheme('light'|'dark')`),
   `ThemeToggle.tsx` (interruptor claro/oscuro compacto de la cabecera; nombre accesible localizado
   y estado anunciado, FR-023). Persistencia con el mismo criterio de tolerancia que F3 (sin
   `localStorage` → tema del sistema, sin avisos).
4. **Superficie del logotipo**: `BrandLogo` se muestra **con sus colores originales siempre**; en
   tema oscuro se apoya en una superficie clara/crema definida como token (nunca `filter`, inversión
   ni recoloreo — la prueba `BrandLogo.test.tsx` que lo prohíbe se conserva intacta).
5. **Temas en el resto de componentes**: las secciones de la portada pasan de colores explícitos
   (`bg-white`, `text-navy`…) a utilidades semánticas; la tabla de contraste AA por par de tokens se
   documenta en `ux.md` (fase de diseño) y se verifica en QA en ambos temas.

## Estrategia visual: espaciado y respiración de sección (detalle) — P10-15/R10-16

El cliente pidió **«más aire» entre secciones**. En lugar de `py-*` sueltos por componente, el
espaciado vertical vive en **tokens** (capa aditiva de `index.css`, junto a los semánticos de tema),
registrados en `@theme inline` para que existan utilidades (`py-section`, `py-section-tight`,
`gap-block`) heredables por F4–F9:

| Token | < 480 px | 480–767 px | 768–1279 px | ≥ 1280 px | Uso |
|---|---|---|---|---|---|
| `--space-section` | `3.25rem` | `4rem` | `5.5rem` | `7rem` | padding vertical de cada sección de la portada |
| `--space-section-tight` | `2rem` | `2.5rem` | `3.5rem` | `4.5rem` | secciones densas o encadenadas del mismo bloque (p. ej. contacto + redes) |
| `--space-block` | `1.5rem` | `2rem` | `2.5rem` | `3rem` | separación entre el encabezado de sección (`SectionHeading`) y su contenido |

- Valores por **breakpoint** (no `clamp()` continuo): predecibles de verificar y de ajustar en un solo
  sitio. Son **defaults revisables**: la calibración final («moderno y con aire, no vacío ni con
  scroll eterno») se cierra en QA con la propuesta B (checklist de `quickstart.md`, SC-002/SC-011);
  el e2e **no** mide píxeles de espaciado (coherente con P10-10/R10-12).
- Regla de composición: **toda** sección de la portada y del futuro F4–F9 usa estos tokens (ninguna
  define su propio padding vertical mágico); el hero de identidad puede usar `--space-section-tight`
  o superior según `ux.md`.
- No se tocan tokens ni estilos del panel (R10-2/R10-8): la capa de espaciado es solo del sitio
  público.

## Estrategia de movimiento (detalle) — nivel **marcado/elegante con revelado al hacer scroll** (FR-005)

> **Revisión de la puerta 2 (2026-10-10)**: el cliente rechazó el nivel «moderado» por
> «casi imperceptible» y pidió más animación. El sistema pasa a revelado de secciones y componentes al
> entrar en pantalla (con entrada escalonada), transiciones perceptibles, microinteracciones marcadas
> y cabecera que transiciona al bajar. Sigue **sin parallax, sin scroll-scrub y sin secuencias
> largas** («elegante, no estridente»). Las garantías incondicionales **no se relajan**:
> `prefers-reduced-motion` (FR-006), contenido y funciones sin depender de la animación (FR-009) y
> lectura/interacción desde el primer instante (FR-007).

### Opción recomendada y alternativas (comparativa completa en `research.md` R10-4)

- **RECOMENDADA — `IntersectionObserver` + CSS con stagger**: `Reveal`/`RevealGroup`
  (`features/publico/motion/`) sobre un único observer compartido; entradas con `opacity` +
  `transform`; escalonado con `transition-delay` por variable `--reveal-index`. Logra el nivel
  marcado con **0 dependencias**, en el 100 % de los navegadores objetivo, de forma determinista y
  verificable en e2e. Lo «casi imperceptible» se corrige con **parámetros** (tabla inferior), no con
  un motor más potente.
- **Descartada — CSS scroll-driven animations** (`animation-timeline: view()/scroll()`): poderosas
  como scroll-scrub nativo, pero **no son Baseline** en 2026 (Chrome/Edge 115+, Safari 17.4+;
  **Firefox estable solo con flag** → 15–25 % de usuarios sin efecto), con riesgo real de que el
  contenido quede **clavado en el keyframe inicial (oculto)** en navegadores sin soporte, con
  `@supports` que puede dar falsos positivos en Firefox, con `prefers-reduced-motion` que **no**
  anula la línea de tiempo (hay que resetearla a mano) y con verificación no determinista en e2e. No
  puede ser ni el mecanismo base ni una mejora progresiva de F10 (un efecto distinto por navegador
  choca con SC-014).
- **Descartada — librería (`motion`/framer-motion, `GSAP` + ScrollTrigger)**: ≈ 20–50 KB gzip + JS
  por frame; sus ventajas reales (layout animations, `AnimatePresence`, gestos, pin/scrub) quedan
  **fuera de F10**, y el scrub/parallax está desaconsejado por la propia spec. Se reevaluó por
  exigencia del nivel alto y se volvió a descartar (constitución §II); sería la candidata de F10B o
  de secciones futuras con transiciones de layout — siempre con justificación y aprobación humana.

### Tokens de movimiento (`index.css`; cimiento heredable por F4–F9 — R10-13)

| Token | Valor | Uso |
|---|---|---|
| `--motion-duration-micro` | `180 ms` | hover/foco/activo |
| `--motion-duration-header` | `220 ms` | cabecera al bajar |
| `--motion-duration-reveal` | `600 ms` | revelado de secciones/bloques |
| `--motion-duration-hero` | `700 ms` | entrada de la portada (solo 1ª carga) |
| `--motion-distance` | `28 px` (`20 px` ≤ 480 px) | desplazamiento de entrada |
| `--motion-stagger-step` | `80 ms` | escalonado entre hermanos (tope 3 saltos = 240 máx.) |
| `--motion-ease-out` | `cubic-bezier(0.22, 1, 0.36, 1)` | entradas y revelados |
| `--motion-ease-standard` | `cubic-bezier(0.4, 0, 0.2, 1)` | microinteracciones y cabecera |

Reglas duras: **suelo de perceptibilidad** — todo revelado/transición de entrada ≥ 400 ms y ≥ 16 px
(o fade completo 0→1) — y **techo** — ≤ 800 ms por elemento y ≤ 800 ms la secuencia total de un
bloque — (FR-005 «hasta ~0,8 s» + «nunca casi imperceptibles»; SC-014).

### Reglas por efecto

- **Revelado al hacer scroll (FR-005/US3)**: `Reveal` envuelve bloques de sección; al entrar en
  pantalla (`rootMargin: 0px 0px -10% 0px`) entran con fade + `translateY(var(--motion-distance))`,
  600 ms, curva `--motion-ease-out`, **una sola vez por elemento y por carga**. Progresivo (R10-5):
  sin JS, con movimiento reducido o si algo falla, el contenido **nace visible**.
- **Entrada escalonada (stagger)**: `RevealGroup` reparte `--reveal-index` entre componentes
  hermanos (tarjetas de una fila, ítems de una lista) → `transition-delay` de 80 ms por salto, tope
  3 saltos (240 máx.); la secuencia total de un bloque nunca supera 800 ms («cuando convenga»:
  grupos de 2–4 elementos, no listas enteras).
- **Transiciones entre secciones/componentes**: el salto por ancla combina desplazamiento suave +
  revelado del destino; el encabezado de sección anima su acento (`scaleX` 0→1, 500 ms) con el
  título; sin cross-fade de página ni orquestación larga entre secciones.
- **Cabecera con transición al bajar (FR-014)**: `SiteHeader` pasa a `data-scrolled` (sentinela
  observada con `IntersectionObserver`, sin listener de `scroll`) y transiciona **solo su
  superficie** —fondo translúcido → sólido, borde hairline y sombra suave, 220 ms— **sin cambio de
  geometría** (misma altura → cero riesgo para anclas y `scroll-padding-top`). Con movimiento
  reducido, cambio instantáneo.
- **Desplazamiento suave a las anclas**: `html { scroll-behavior: smooth }` **solo** bajo
  `prefers-reduced-motion: no-preference`; con movimiento reducido, `auto` (salto instantáneo). El
  destino queda visible gracias al `scroll-padding-top` medido (R10-6).
- **Microinteracciones marcadas (FR-008)**: 180 ms en `hover`, `:focus-visible` y `:active` de
  enlaces, botones, tarjetas y controles (color, borde, sombra, `transform: translateY(-2px)` /
  `scale(1.02)` como máximo). Nunca propiedades de layout → 0 desplazamientos del maquetado. El
  logotipo **queda excluido** de todo movimiento (FR-002).
- **Sin repeticiones molestas (FR-008/SC-014)**: claves React estables por sección (`key` = ancla) →
  el cambio de idioma o de tema re-renderiza sin re-montar, las secciones vistas no se re-animan y el
  hero solo entra en la primera carga real.
- **Rendimiento (FR-007/SC-006)**: solo `opacity`/`transform` (compositor); **sin `filter`/`blur`**;
  `will-change` **temporal** (se añade al empezar la animación y se retira en `transitionend`) y
  medido en DevTools; un `IntersectionObserver` compartido; sin listeners de `scroll`; salvavidas de
  1 s que revela todo lo pendiente si el observer no dispara; contenido visible e interactivo desde
  el primer instante (la animación es una capa visual; los elementos pendientes siguen siendo
  clicables y el foco los revela al instante). Objetivo: 0 tareas largas > 50 ms atribuibles al
  movimiento en el dispositivo de referencia más modesto; si degrada, se reduce o se elimina.

## Estrategia de navegación (detalle) — **responsive** (P10-6 + P10-13/R10-14)

> **Decisión del cliente (aprobación del plan, 2026-10-10)**: en escritorio la navegación sigue
> siendo barra fija con todas las anclas visibles; en **móvil** se colapsa en un **menú hamburguesa
> accesible**. Esto **sustituye solo en pantallas estrechas** la restricción previa de «sin menús
> ocultos» (FR-010/FR-012); la restricción sigue viva en escritorio y el espíritu general —a un toque
> o clic, sin gestos desconocidos— se mantiene en ambos modos.

- **Breakpoint de cambio**: **860 px** —móvil **≤ 860 px**, escritorio **≥ 861 px**—: el mismo corte
  de la maqueta aprobada `propuestas/propuesta-b.html` y de `ux.md` (`TransitionHeader`, breakpoint
  860), para que diseño e implementación no diverjan. En escritorio/tableta anclas siempre visibles
  en filas `flex-wrap`; en móvil la cabecera compone fila de marca + controles (idioma, tema) +
  **botón hamburguesa** y las anclas viven en el panel. El corte se hace **con CSS (media queries)**,
  no con JS de resize: sin sorpresas de hidratación y con el comportamiento correcto desde el primer
  pintado. Alternativas descartadas: 768 px (el `md` de Tailwind 4; divergería de la maqueta
  aprobada sin ganancia real) y 1024 px (ocultaría innecesariamente la navegación en tabletas
  apaisadas, donde sí caben las filas de anclas).
- **Cabecera fija (ambos modos)**: `SiteHeader` con `position: sticky; top: 0; z-index` adecuado.
  Composición: fila de marca (logo + nombre) y controles (selector de idioma compacto + control de
  tema) + fila(s) de anclas (escritorio) o botón hamburguesa (móvil). Al bajar, la cabecera
  transiciona su **superficie** a `data-scrolled` (sin cambio de geometría; R10-13/P10-12) **en los
  dos modos**; con movimiento reducido, el cambio es instantáneo. La transición adopta la «banda de
  vidrio» de la propuesta B —superficie translúcida + `backdrop-filter` **acotado a la cabecera**,
  borde hairline y sombra suave, 220 ms— manteniendo la **geometría constante**: la «compactación»
  de `ux.md` es de superficie, **no de altura** (reducir la altura movería anclas y haría inestable
  `scroll-padding-top`, R10-13); si DevTools muestra coste de pintura del blur, se degrada a
  superficie sólida sin `backdrop-filter` (FR-007/RG10-9). `--header-offset` se re-mide con
  `ResizeObserver` **con el menú cerrado** (el panel es un overlay y no altera la altura de la
  cabecera, por lo que `scroll-padding-top` es estable).
- **Escritorio (≥ 861 px) — sin scroll incómodo**: `PublicNav` sin `overflow-x-auto`: `flex-wrap`
  con filas compactas y objetivos ≥ 44 px; sin scroll horizontal de enlaces ni de página (FR-011,
  SC-004). El e2e comprueba `scrollWidth <= clientWidth` en 320/360/768/1280 px.
- **Móvil (≤ 860 px) — menú hamburguesa accesible** (patrón *disclosure* de las guías APG, alineado
  con `ux.md` §2.3/§6; sin librerías, R10-10):
  - **Botón**: `<button>` real con **nombre accesible localizado** (`nav.menu.toggle`), `aria-expanded`
    sincronizado con el estado y `aria-controls` apuntando al id del panel; objetivo **≥ 44 px**;
    siempre visible (es el equivalente móvil de «la navegación siempre está a un toque», FR-014).
  - **Panel**: lista de **las mismas secciones** (mismo modelo `sections.ts`, mismas claves i18n →
    mismos nombres accesibles que en escritorio) bajo la cabecera, con superficie sólida por tokens
    (coherente con `--surface` de cada tema), altura máxima `calc(100dvh - var(--header-offset))` con
    scroll interno propio si hiciera falta, y un **control visible de cierre**.
  - **Gestión de foco** (igual que `ux.md` y la maqueta B): al abrir, el foco **entra en el panel**
    (primer enlace de la lista; el control visible de cierre también está presente); **`Escape`**,
    el **control visible de cierre**, el **toque/clic fuera** (scrim) y el propio botón lo cierran y
    **devuelven el foco al botón**; activar una sección cierra el panel y el foco continúa en el
    destino (ancla); mientras está abierto, `Tab`/`Shift+Tab` quedan **contenidos en el panel y su
    botón** (el fondo va con scrim y sin scroll, así que no hay nada operable detrás); el foco nunca
    queda sobre contenido oculto.
  - **Sin scroll de fondo**: con el panel abierto, `html[data-nav-open='true'] { overflow: hidden }`
    (la página de detrás no se desplaza ni con rueda ni con teclado); al cerrar, se restaura.
  - **Sin gestos desconocidos**: solo clic/toque y teclado (`Enter`/`Espacio` abren y cierran); ni
    swipe ni deslizamientos.
  - **Transiciones**: el panel entra/sale con fade de `--motion-duration-micro` (180 ms, solo
    `opacity`/`transform`, sin animación de altura ni de layout); con `prefers-reduced-motion`,
    apertura/cierre **instantáneos** (FR-006).
  - **Cruce de breakpoint** (rotar el dispositivo o redimensionar): el panel se **cierra solo** y
    `aria-expanded` vuelve a `false`; al pasar a escritorio reaparecen las filas de anclas sin
    recarga y sin estado residual.
  - **Sin JS (fallback)**: los enlaces se pintan como filas compactas visibles a cualquier ancho
    (el botón no se renderiza/está oculto sin el marcador `js` de `index.html`): la navegación es
    usable aunque el JS falle (misma filosofía de R10-5: el contenido nunca depende de la animación
    ni del script). Con JS, el colapso lo resuelve CSS bajo la media query + `data-nav-open`.
- **Compatibilidad con las pruebas congeladas de F3 (R10-9/SC-008)** — restricciones que **fijan** el
  diseño (verificado leyendo las suites): (1) el elemento `<nav aria-label="Secciones de la portada">`
  **permanece visible** incluso a 320 px y 160 px (zoom 200 %) → el botón hamburguesa va **dentro**
  del `<nav>` y solo la lista de anclas se colapsa; (2) los enlaces de sección **siguen en el DOM**
  siempre (el colapso es CSS puro, sin `hidden`/`aria-hidden`/`inert` de React) → las unitarias de
  jsdom pasan sin cambios y las e2e de escritorio (1280 px) ven las anclas visibles; (3) la travesía
  por teclado congelada a 320 px exige que el ciclo contenga «Quiénes somos… Redes sociales» y cierre
  en **< 30 pasos** → con el menú cerrado en móvil esas etiquetas las aporta el **mapa del sitio del
  pie** (P10-14, mismos nombres) y el saldo neto de paradas de tabulación respecto a hoy es mínimo
  (ver RG10-21); (4) todos los objetivos visibles ≥ 44 px y el contraste del `footer` pasan por las
  auditorías congeladas → botón, panel y enlaces del pie se diseñan a esos umbrales.
- **Anclas que no se tapan**: `--header-offset` (altura real medida) alimenta `scroll-padding-top`;
  el título de la sección destino queda visible al saltar y al bajar (FR-014), también al navegar
  desde el panel móvil (que se cierra antes del desplazamiento suave).
- **Accesibilidad (FR-013/FR-022)**: enlaces reales `<a href="#…">` con texto claro, orden de foco
  natural, foco visible (utilidad de foco por contexto, ya patrón de F3), `aria-label` del `nav`
  localizado; el enlace de salto «Saltar al contenido» se conserva; estado del menú anunciado
  (`aria-expanded`) y cierre anunciado de forma no intrusiva.
- **Escala a F4–F9 (FR-012)**: el modelo de navegación sigue siendo **datos** (patrón
  `sections.ts`); las secciones futuras añaden entradas y sus rutas cuelgan de `PublicLayout`. En
  escritorio, ~8–10 enlaces caben en filas compactas; en móvil, el panel es una lista que escala sin
  coste (sin menús anidados).

## Mapa del sitio en el pie (detalle) — P10-14/R10-15

- **Componente nuevo**: `FooterSiteMap` en `frontend/src/features/publico/components/FooterSiteMap.tsx`,
  montado dentro de `PublicFooter` (el `<footer>`/landmark `contentinfo` de F3 se conserva: el mapa
  vive dentro de él, no crea otro `footer`).
- **Fuente de datos única**: se alimenta del **mismo modelo de secciones** que la navegación
  (`features/publico/sections.ts`: `PUBLIC_SECTIONS` + `presentSections(data)`, con `PublicLayout`
  pasando las secciones presentes a `PublicFooter`). Mismas claves i18n → **mismos nombres de enlace**
  que la cabecera (coherencia nav ↔ pie, y las etiquetas congeladas de la travesía por teclado se
  cumplen por construcción). Hoy lista las secciones de la portada (`quienes-somos`, `horario`,
  `whatsapp`, `contacto`, `redes`) más el enlace a la portada (`/`) si `ux.md` lo incluye.
- **Regla «sin enlaces rotos»**: **solo se renderizan destinos existentes** — secciones con contenido
  publicado (mismo criterio borrador de F3: una sección ausente no deja enlace muerto, ni en el pie ni
  en la cabecera) y, para F4–F9, entradas del modelo con `href` de **ruta ya implementada**. El modelo
  se extiende entonces con un campo `href` (hoy solo `anchor`) y el componente **no cambia**: no se
  enlaza nunca a rutas F4–F9 inexistentes.
- **Disposición**: columnas en escritorio (grid con `auto-fit`/2–3 columnas) y **apilado** en móvil;
  secciones de la propuesta B aplicadas (tokens semánticos, `--space-section-tight`/`--space-block`).
- **Accesibilidad**: landmark `<nav>` propio con **nombre accesible propio y distinto** del de la
  cabecera (`footer.sitemap.label`, es/en; dos `navigation` sin nombres distintos romperían las
  consultas congeladas), lista `<ul>` real, foco visible, objetivos ≥ 44 px, contraste AA en ambos
  temas (el e2e de contraste congelado ya audita `footer :is(a,p,span)`).

## Selector de idioma (detalle)

- `LanguageSwitcher` **conserva** radios nativos, textos completos, `useId`, `ss.lang` y el cambio
  inmediato (FR-016). La compactación es de **espaciado y disposición** (gaps/padding más ajustados,
  acomodo en el clúster de controles), sin bajar el cuerpo de texto del umbral de legibilidad (FR-004)
  y sin siglas, píldoras, menús ni banderas (FR-015).
- Accesibilidad: cada opción se identifica por el **nombre completo del idioma** (visible y como
  nombre accesible); `legend` localizado; operable con un clic o toque y con teclado, foco visible.
- Se muestra en cabecera y pie (como en F3) con comportamiento idéntico en ambos (US4 esc. 5).

## Project Structure

### Documentation (esta funcionalidad)

```text
specs/010-sistema-diseno-movimiento/
├── plan.md              # EDITABLE (este archivo, /speckit.plan)
├── research.md          # EDITABLE (fase 0: R10-1…R10-13)
├── data-model.md        # EDITABLE (fase 1): SIN CAMBIOS de datos, documentado
├── quickstart.md        # EDITABLE (fase 1: cómo levantar y validar el rediseño)
├── spec.md              # APROBADA — no se edita
├── contracts/
│   └── README.md        # EDITABLE (fase 1): SIN CAMBIOS de API, documentado
├── ux.md                # DEL disenador-ux (3 propuestas; el humano aprueba una) — paralelo
└── tasks.md             # Fase 2 (/speckit.tasks — NO lo crea este comando)
```

### Source Code — frontend (lo único que se toca)

```text
frontend/
├── index.html                           # EDITADO: script anti-FOUC (data-theme + color-scheme)
├── e2e/
│   ├── diseno-movimiento.spec.ts        # NUEVO: navegación fija (escritorio) y sin scroll horizontal
│   │                                    #   (320/360/768/1280 px), anclas visibles en escritorio,
│   │                                    #   selector compacto con comportamiento intacto, foco
│   │                                    #   visible (US1/US2/US4)
│   ├── navegacion-movil.spec.ts         # NUEVO: menú hamburguesa a 320 px — abrir/cerrar con ratón
│   │                                    #   y teclado (Enter/Espacio), foco entra al panel, Escape y
│   │                                    #   control visible cierran y devuelven foco al botón,
│   │                                    #   aria-expanded/aria-controls, ≥ 44 px, sin scroll de
│   │                                    #   fondo, cruce de breakpoint — y mapa del sitio del pie
│   │                                    #   (enlaces alcanzables, sin destinos rotos) (US2/US4)
│   ├── tema-oscuro.spec.ts              # NUEVO: sistema→manual→persistencia, cambio en caliente,
│   │                                    #   logo sin recolorear, mismas funciones en ambos temas (US5)
│   ├── movimiento.spec.ts               # NUEVO: revelado PERCEPTIBLE al scroll (opacidad/distancia
│   │                                    #   ≥ suelo, duración en rango), stagger entre hermanos,
│   │                                    #   cabecera con estado data-scrolled al bajar, anclas con
│   │                                    #   desplazamiento suave y contenido operable durante la
│   │                                    #   animación (US3 modo normal, SC-014)
│   ├── movimiento-reducido.spec.ts      # NUEVO: emulateMedia({ reducedMotion: 'reduce' }) → 0
│   │                                    #   revelados/transiciones de entrada, contenido completo
│   │                                    #   desde el primer instante, scroll-behavior auto (US3,
│   │                                    #   SC-005)
│   └── (portada-*.spec.ts, acceso…)     # SIN CAMBIOS — red de no-regresión de F2/F3 (SC-008)
└── src/
    ├── index.css                        # EDITADO: @custom-variant dark, tokens semánticos claro/
    │                                    #   oscuro (@theme inline), tokens de espaciado de sección
    │                                    #   (--space-section*, R10-16), scroll-padding-top,
    │                                    #   scroll-behavior condicional, bloque global de
    │                                    #   prefers-reduced-motion, tokens de movimiento (R10-13),
    │                                    #   utilidades de foco y de motion
    ├── components/
    │   ├── (Button, Field, …)           # SIN CAMBIOS — inventario del panel (FR-020, R10-8)
    │   └── publico/                     # NUEVO — primitivas del sistema visual público
    │       ├── PublicButton.tsx         #   variantes primary/accent/secondary (marca)
    │       ├── SectionHeading.tsx       #   H2 de sección con acento (patrón único de título)
    │       ├── Card.tsx                 #   tarjeta de contenido (radio, borde, sombra por tema)
    │       └── *.test.tsx
    ├── features/publico/
    │   ├── theme/                       # NUEVO — R10-3
    │   │   ├── ThemeProvider.tsx        #   resolución sistema→manual, matchMedia en vivo, data-theme
    │   │   ├── useTheme.ts              #   contexto: theme efectivo, preference, setTheme
    │   │   ├── ThemeToggle.tsx          #   interruptor claro/oscuro de la cabecera (accesible)
    │   │   └── theme.test.tsx
    │   ├── motion/                      # NUEVO — R10-4/R10-5/R10-13
    │   │   ├── Reveal.tsx               #   revelado progresivo (IntersectionObserver compartido)
    │   │   ├── RevealGroup.tsx          #   entrada escalonada (--reveal-index, tope 3 saltos)
    │   │   ├── usePrefersReducedMotion.ts
    │   │   ├── useScrolled.ts           #   sentinela → estado data-scrolled de la cabecera
    │   │   └── motion.test.tsx
    │   ├── i18n/
    │   │   ├── messages.ts              # EDITADO: + claves 'theme.*', 'nav.menu.toggle',
    │   │   │                            #   'nav.menu.close', 'footer.sitemap.label'
    │   │   ├── es.ts / en.ts            # EDITADOS: traducciones de las claves nuevas (FR-019)
    │   │   ├── LanguageSwitcher.tsx     # EDITADO: solo formato compacto (comportamiento intacto)
    │   │   └── i18n.test.tsx            # EDITADO: + claves nuevas (exhaustividad es/en)
    │   ├── components/
    │   │   ├── PublicLayout.tsx         # EDITADO: monta ThemeProvider + SiteHeader sticky y pasa
    │   │   │                            #   las secciones presentes al pie
    │   │   ├── SiteHeader.tsx           # NUEVO (extraído): cabecera fija (marca, nav, idioma, tema)
    │   │   ├── PublicNav.tsx            # NUEVO: anclas en filas flex-wrap (≥ 861 px) + botón
    │   │   │                            #   hamburguesa y panel accesible (≤ 860 px, R10-14);
    │   │   │                            #   patrón escalable F4–F9
    │   │   ├── PublicNav.test.tsx       # NUEVO: estado/ARIA/foco/Escape/scroll bloqueado/fallback
    │   │   ├── FooterSiteMap.tsx        # NUEVO: mapa del sitio del pie (columnas en escritorio,
    │   │   │                            #   apilado en móvil; solo destinos existentes, R10-15)
    │   │   ├── FooterSiteMap.test.tsx   # NUEVO: mismos nombres que la cabecera, sin enlaces a
    │   │   │                            #   secciones ausentes, landmark con nombre propio
    │   │   ├── PublicFooter.tsx         # EDITADO: aplica tokens semánticos + espaciado; selector
    │   │   │                            #   compacto; monta FooterSiteMap
    │   │   ├── IdentityHero.tsx         # EDITADO: aplica sistema visual + entrada de hero (motion)
    │   │   ├── (WhoWeAreSection, ScheduleSection, WhatsAppSection, ContactSection,
    │   │   │    SocialSection, SectionSkeleton, BrandLogo, SocialIcon)
    │   │   │                            # EDITADOS: tokens semánticos + reglas de composición;
    │   │   │                            #   BrandLogo SIN cambios de reglas (nunca se recolorea)
    │   │   └── *.test.tsx               # EDITADOS/NUEVOS según toque
    │   ├── pages/HomePage.tsx           # EDITADO: composición con Reveal/RevealGroup y sistema visual
    │   └── sections.ts                  # SIN CAMBIOS (o mínimo: si ux.md reordena, solo el orden)
    ├── features/(panel|usuarios|roles|auditoria|auth|informacion|status)/
    │                                    # SIN CAMBIOS — frontera F10B (FR-020)
    └── app/ (router, layout, guards…)   # SIN CAMBIOS — rutas intactas
```

> **Nota de alcance de edición**: `sections.ts` y las rutas solo se tocarían si la propuesta aprobada
> reordena secciones de la portada (permitido por la spec, «composición de la portada», sin cambiar
> contenido ni comportamiento); en ese caso el cambio es de orden/registro, sin lógica nueva.

## Estrategia de pruebas

Cobertura: la que exige §III para frontend (componentes con lógica → Vitest; flujos críticos →
Playwright) **y** la no-regresión (SC-008). Comandos: `npm test -- --run` · `npm run typecheck` ·
`npm run lint` · `make e2e` · `make ci`.

| Capa | Tipo | Qué verifica F10 | Dónde |
|---|---|---|---|
| Tema | Unitaria (Vitest + TL) | Primera visita sin elección → sistema (`matchMedia`); elección manual → inmediata y persistida en `ss.theme`; re-visita → manda la manual; sistema cambia en caliente **solo** sin elección manual; `localStorage` caído → sistema sin errores; `data-theme`/`color-scheme` aplicados en `<html>`; `ThemeToggle` alterna, tiene nombre accesible y estado | `features/publico/theme/theme.test.tsx` |
| Movimiento | Unitaria | `Reveal`/`RevealGroup`: el contenido pendiente se revela al observar con fade + `translateY` (parámetros ≥ suelo: ≥ 400 ms, ≥ 16 px — SC-014); escalonado de 80 ms con tope de 3 saltos; una sola vez por elemento (sin re-animación al re-renderizar); con `prefers-reduced-motion` el contenido nace visible (0 animación) y la preferencia en vivo se respeta sin recargar; foco (`focusin`) sobre pendiente → visible al instante; JS sin montar o salvavidas (timeout) → contenido visible (fallback) | `features/publico/motion/motion.test.tsx` |
| Cabecera al bajar | Unitaria | `useScrolled` (sentinela): `data-scrolled` se activa al bajar y se quita al volver arriba; con movimiento reducido el cambio es instantáneo (sin transición); sin listeners de `scroll` | `features/publico/motion/motion.test.tsx` |
| Navegación | Unitaria | `SiteHeader`/`PublicNav`: todas las anclas presentes, sin `overflow-x-auto` en escritorio, `scroll-padding-top` actualizado con la altura medida; secciones ausentes (borrador) no dejan enlaces rotos (hereda F3); los enlaces **siempre están en el DOM** (colapso solo CSS) | tests de `SiteHeader`/`PublicNav` |
| Menú hamburguesa (móvil) | Unitaria | Estado abierto/cerrado: `aria-expanded` sincronizado y `aria-controls` apuntando al panel; nombre accesible del botón localizado; al abrir, el foco entra al panel (primer enlace) y `Tab` queda contenido en panel + botón; `Escape`, el control visible de cierre y el scrim (toque fuera) cierran y **devuelven el foco al botón**; cierre al activar una sección; cierre al superar el breakpoint (860 px); scroll de fondo bloqueado con el panel abierto (`data-nav-open`); apertura/cierre instantáneos con `prefers-reduced-motion`; fallback sin JS → enlaces visibles | `features/publico/components/PublicNav.test.tsx` |
| Mapa del sitio (pie) | Unitaria | `FooterSiteMap`: lista exactamente las secciones presentes con los **mismos nombres** que la cabecera (mismas claves i18n); sin enlaces a secciones ausentes/en borrador ni a rutas inexistentes; landmark `<nav>` con nombre accesible propio; enlaces reales `href="#…"`; columnas/apilado por breakpoint | `features/publico/components/FooterSiteMap.test.tsx` |
| Selector | Unitaria | Comportamiento idéntico a F3 (las pruebas existentes de `i18n.test.tsx` pasan **sin cambios**); formato compacto: textos completos, nombre accesible completo, sin siglas | `i18n.test.tsx` (existente) + assertions de formato |
| i18n | Unitaria | Claves nuevas (`theme.*`) presentes en `es` y `en` (exhaustividad tipada) | `i18n.test.tsx` |
| Primitivas públicas | Unitaria | `PublicButton`/`SectionHeading`/`Card`: estados hover/foco/activo accesibles (clases de foco), sin reglas que afecten al logo | `components/publico/*.test.tsx` |
| No-regresión | Unitaria | **Todas** las suites de F2/F3 pasan sin editar (incluida `BrandLogo.test.tsx`: logo sin `filter/transform/...`) | `npm test -- --run` |
| E2E navegación (escritorio) | Playwright | Barra fija visible al bajar (todas las anclas a la vista ≥ 861 px, sin menús ocultos); clic en ancla → título de la sección visible (no tapado); 320/360/768/1024/1280 px: sin scroll horizontal de página ni de enlaces; recorrido con teclado y foco visible; selector: cambia idioma al instante, persiste al navegar y entre visitas, primera visita en español | `e2e/diseno-movimiento.spec.ts` |
| E2E navegación móvil | Playwright | **320 px, menú hamburguesa**: botón ≥ 44 px con nombre accesible y `aria-expanded`/`aria-controls`; **abrir/cerrar con ratón y con teclado** (Enter/Espacio); al abrir, el foco entra al panel (primer enlace); **Escape**, control visible y toque fuera cierran y devuelven el foco al botón; `Tab` contenido en el panel; al activar una sección, el panel se cierra y el título del destino queda visible; **sin scroll de fondo** con el panel abierto (rueda y teclado); sin scroll horizontal; al ensanchar a ≥ 861 px el panel se cierra y reaparecen las filas de anclas; cabecera con `data-scrolled` al bajar en ambos modos | `e2e/navegacion-movil.spec.ts` |
| E2E mapa del sitio | Playwright | El pie lista las secciones publicadas (columnas en escritorio, apilado a 320 px); sus enlaces son alcanzables por teclado y llevan a la sección (título visible tras el salto); con una sección en borrador no queda enlace (ni en el pie ni en la cabecera); landmark del pie con nombre accesible propio | `e2e/navegacion-movil.spec.ts` |
| E2E tema | Playwright | Tres caminos de US5: sistema oscuro sin elección → oscuro; manual → inmediato y persistente tras recarga; primera visita → sistema. Logo con colores originales en oscuro; mismas funciones/contenido en ambos temas; control accesible por teclado | `e2e/tema-oscuro.spec.ts` |
| E2E movimiento (modo normal) | Playwright | El revelado es **claramente perceptible**: una sección fuera de pantalla nace con opacidad < 1 y desplazamiento, y al entrar en pantalla termina en opacidad 1 / `transform: none` con `transition-duration` en rango (≥ 400 ms y ≤ 900 ms) y distancia ≥ 16 px (suelo anti «casi imperceptible»); componentes hermanos con `transition-delay` escalonado; cabecera gana `data-scrolled` al bajar (superficie con transición, sin cambio de altura); clic en ancla → desplazamiento suave y título del destino visible; la animación **no bloquea**: el contenido pendiente es clicable y un clic durante la animación funciona al instante (US3, SC-014) | `e2e/movimiento.spec.ts` |
| E2E movimiento reducido | Playwright | `emulateMedia({ reducedMotion: 'reduce' })`: **0** revelados ni transiciones de entrada perceptibles — todos los bloques con opacidad 1 y `transform: none` desde el primer instante, sin `transition-delay`, `scroll-behavior: auto`, cabecera sin transición — y 100 % del contenido y las funciones disponibles, idénticos al modo con movimiento (SC-005); contraste con el modo normal del spec anterior | `e2e/movimiento-reducido.spec.ts` |
| E2E no-regresión | Playwright | Suites de F3 (`portada-publica`, `portada-accesibilidad`, `portada-panel`, `portada-imagenes`, `portada-patch-null`) y de F2 (`acceso`, `auditoria`, `status`) pasan **sin cambios** (SC-008) | `make e2e` |
| Manual (QA + humano) | Checklist `quickstart.md` | «Moderno» y fidelidad al Manual (SC-002), 3 propuestas (SC-013), **movimiento marcado/elegante aprobado como perceptible y sin marear** (SC-014), contraste AA por tema y estado (SC-012), selector con «menos espacio» (SC-007), logo sin recolorear, panel sin cambios, zoom 200 % | `quickstart.md` |

**Pruebas que no faltan** (las críticas de la spec): **revelado claramente perceptible en modo normal
y ausente del todo con `prefers-reduced-motion`** (SC-014/SC-005) — verificado con
`emulateMedia({ reducedMotion })` y con aserciones de estilos computados (suelo ≥ 400 ms / ≥ 16 px,
techo ≤ 800 ms) —; **cabecera con transición al bajar sin tapar anclas**; **contenido operable
durante la animación** (FR-007); **barra fija visible en el 100 % del recorrido** y **0
desplazamientos horizontales** desde 320 px (SC-004); **persistencia del tema manual** en el 100 % de
re-visitas y **sistema** en el 100 % de primeras visitas (SC-012); **0 recoloreados del logo**
(SC-012); **0 textos sin traducir** (SC-010); **0 regresiones** de F2/F3 (SC-008); **0 re-animaciones
ni rebotes** al cambiar de idioma o de tema (SC-014); foco nunca sobre contenido oculto por revelado
(edge case); contenido visible si una animación o el observer fallan (edge case + salvavidas);
**menú hamburguesa móvil accesible de punta a punta** — abrir/cerrar con ratón y teclado, foco al
abrir, Escape/control visible con foco devuelto al botón, `aria-expanded`/`aria-controls`, ≥ 44 px,
**sin scroll de fondo** — (FR-010/FR-013/FR-025 + ajuste del cliente); **mapa del sitio del pie con
0 enlaces rotos** (secciones en borrador ausentes; F4–F9 solo cuando sus rutas existan) y con los
mismos nombres que la cabecera (FR-012/FR-013 + ajuste del cliente); **espaciado de sección por
tokens** en toda la portada (FR-001/FR-003 + ajuste del cliente); y respeto de las **pruebas
congeladas** que fijan el diseño nuevo: `<nav>` visible a 320/160 px, ciclo de teclado < 30 pasos con
las etiquetas de sección presentes (RG10-21) y objetivos ≥ 44 px.

## Cobertura de requisitos

> **Nota de decisión (aprobación del plan, 2026-10-10)**: los 3 ajustes del cliente —navegación
> responsive con menú hamburguesa accesible en móvil, mapa del sitio en el pie y espaciado de
> sección— quedaron **reflejados en `spec.md`** (FR-010…FR-014 con el patrón por dispositivo,
> FR-025 y la nueva **FR-027** del mapa del sitio) y en `ux.md` (ronda 3, propuesta B ajustada). Este
> plan implementa esa redacción; el espaciado «más aire» se cubre con FR-001/FR-003/US1 («espaciado»
> del sistema visual) materializado en tokens (P10-15).

| Requisito | Dónde se satisface | Se verifica con |
|---|---|---|
| FR-001/FR-002/FR-004 (sistema visual sobre el Manual, reglas del logo, legibilidad) | Tokens semánticos + `components/publico/` + secciones rediseñadas + **tokens de espaciado `--space-section*` (P10-15)**; `BrandLogo` intacto | QA visual vs. Manual (`quickstart.md`), `BrandLogo.test.tsx` |
| FR-003 (componentes base, sin dos estilos por función) | `components/publico/` + reglas de composición documentadas (incluido el espaciado por tokens) | Revisión visual (SC-001/SC-011) |
| FR-005…FR-009 (movimiento marcado con revelado al scroll, reduced-motion, rendimiento, sin repeticiones, nunca dependiente) | P10-4/P10-5/P10-12 (`motion/`, tokens de movimiento, cabecera `data-scrolled`, CSS) | `motion.test.tsx`, `e2e/movimiento.spec.ts` (perceptibilidad, stagger, cabecera, sin bloqueo) y `e2e/movimiento-reducido.spec.ts` (SC-005/SC-006/SC-014) |
| FR-010…FR-014 (navegación clara, sin scroll incómodo, escala F4–F9, accesible, barra fija **por dispositivo**) | **P10-6 + P10-13**: escritorio ≥ 861 px = barra fija con anclas en `flex-wrap` y `scroll-padding-top` medido; móvil ≤ 860 px = **menú hamburguesa accesible** (patrón de FR-010): botón ≥ 44 px con `aria-expanded`/`aria-controls`, panel con las secciones, foco al abrir, Escape/control visible/toque fuera con foco devuelto al botón, sin scroll de fondo, sin gestos. **P10-14**: mapa del sitio en el pie (mismo modelo de secciones, sin enlaces rotos) | Unitarias `PublicNav`/`FooterSiteMap`; `e2e/diseno-movimiento.spec.ts` + `e2e/navegacion-movil.spec.ts` (SC-003/SC-004) |
| FR-015…FR-017 (selector compacto, comportamiento F3, sin desplazamientos) | P10-7 (`LanguageSwitcher`) | `i18n.test.tsx` + e2e (SC-007) |
| FR-018 (aplicación a toda la portada + cimientos para F4–F9) | Todas las secciones + **mapa del sitio del pie (`FooterSiteMap`) y tokens de espaciado** como cimientos heredables + documentación en este plan y `ux.md` | Revisión visual (SC-001/SC-011) |
| FR-019 (textos nuevos es/en) | Claves `theme.*` en `messages.ts`/`es.ts`/`en.ts` | `i18n.test.tsx` (SC-010) |
| FR-020/FR-021 (no-regresión funcional, sin datos nuevos) | Frontera F10B; `data-model.md`; contrato congelado (R10-9) | Suites F2/F3 sin cambios (SC-008) |
| FR-022 (accesibilidad en ambos temas) | Foco por contexto, orden, objetivos ≥ 44 px, temas con contraste AA; **gestión de foco del menú móvil** (abrir/cerrar/Escape) y **mapa del sitio accesible** (landmark con nombre propio) | QA manual + e2e teclado (SC-009), `e2e/navegacion-movil.spec.ts` |
| FR-023/FR-024/FR-025 (modo oscuro + control manual, contraste AA, cabecera completa a 320 px) | P10-1…P10-3, P10-11; a 320 px la cabecera alberga marca + selector + tema + **botón hamburguesa ≥ 44 px** (los enlaces, a un toque en el panel — P10-13); si hace falta, la palabra de marca se reduce a monograma por debajo de 480 px (permitido por FR-002: sin deformar ni recolorear; detalle de `ux.md`) | `theme.test.tsx`, e2e tema (SC-012), e2e móvil (320 px) |
| FR-027 (mapa del sitio del pie, sin enlaces rotos, preparado para F4–F9) | **P10-14** (`FooterSiteMap` sobre `sections.ts`/`presentSections`; solo destinos existentes; `href` de ruta solo cuando exista) | Unitarias `FooterSiteMap.test.tsx`, e2e mapa del sitio (SC-004/SC-011) |
| FR-026 (3 propuestas y aprobación antes de código) | Proceso: `ux.md` (disenador-ux) + puerta humana | Registro en `estado.md` (orquestador; SC-013) |

## Riesgos y mitigación

| # | Riesgo | Mitigación |
|---|---|---|
| RG10-1 | Flash del tema al cargar («antes y después») | Script anti-FOUC en `index.html` (R10-3) + e2e de primer atributo correcto |
| RG10-2 | Cambio visual **del panel** por efecto colateral de tokens/estilos globales | Tokens nuevos **aditivos** (P10-2); inventario compartido sin tocar (P10-8); revisión visual del panel en QA; suites de panel sin cambios |
| RG10-3 | La barra fija tapa el destino de un ancla o contenido ampliado al 200 % | `scroll-padding-top` con altura medida en vivo (P10-6) + e2e «título visible tras el salto» |
| RG10-4 | A 320 px la cabecera (marca + idioma + tema + menú) desborda o se tapa | Composición compacta: fila de marca/controles + botón hamburguesa ≥ 44 px (P10-13); sin scroll lateral; e2e con `scrollWidth <= clientWidth` y elementos alcanzables |
| RG10-5 | Contenido oculto por el revelado (JS caído, foco temprano, reduced-motion) | Revelado progresivo con contenido visible por defecto + `focusin` inmediato + doble capa reduced-motion (P10-5) |
| RG10-6 | Pruebas de F3 que codifican detalles visuales del diseño viejo | Contrato semántico congelado (R10-9); si una prueba exigiera un cambio visual, **escala al humano** (SC-008 prohíbe editarla para «hacerla pasar») |
| RG10-7 | Fallos de contraste AA en tema oscuro (estados densos, tarjetas, foco) | Tabla de contraste por par de tokens en la fase de diseño; checklist QA por tema y estado; regla: colores derivados solo para contraste (herencia de F3) |
| RG10-8 | La propuesta visual elegida desborda la mecánica del plan | La mecánica es independiente de la dirección (ver «Frontera con el diseño»); `ux.md` solo fija dirección dentro de ella |
| RG10-9 | Tirones o sobrecoste de rendimiento al revelar muchas animaciones en dispositivos modestos (nivel marcado) | Solo `opacity`/`transform` (compositor); un `IntersectionObserver` compartido; sin `filter`/`blur`; `will-change` temporal y medido; objetivo 0 tareas largas > 50 ms en el dispositivo de referencia (checklist `quickstart.md`); si degrada, se reduce o se elimina (FR-007) |
| RG10-10 | El logotipo se recolorea/deforma en oscuro o en microinteracciones | `BrandLogo` excluido de motion y de temas; prueba existente intacta; checklist QA (SC-012) |
| RG10-11 | Textos nuevos sin traducir (SC-010) | Catálogo tipado `PublicMessageKey`: falta una traducción → **no compila** (patrón de F3) |
| RG10-12 | Cambio de idioma re-anima o descoloca la composición | Claves React estables (sin re-montaje); longitudes distintas es/en probadas en e2e (edge case de la spec) |
| RG10-13 | **Movimiento excesivo, estridente o mareante** (rechazo inverso del cliente: pidió «más animación», pero la spec exige «elegante, nada estridente ni mareante») | Parámetros acotados en tokens (≤ 800 ms por elemento y por secuencia, stagger ≤ 240 ms, distancias ≤ 28 px, sin parallax/scrub/blur, sin curvas elásticas); revisión humana en la aceptación (SC-014: «perceptible y sin molestar»); ajuste fino de los tokens, no de cada componente |
| RG10-14 | **Volver a un movimiento «casi imperceptible»** por miedo a lo estridente (el error de la primera ronda) | Suelo de perceptibilidad codificado y verificado: revelados ≥ 400 ms y ≥ 16 px (aserciones de estilos computados en `e2e/movimiento.spec.ts`) + checklist de aceptación humana (SC-014) |
| RG10-15 | Presión por añadir una **librería de animación** nueva («más animación» ≠ más motor) | Evaluada y descartada con datos (R10-4/R10-10: ~20–50 KB, JS por frame, ventajas fuera de alcance); cualquier adición futura exige justificación §II y **aprobación humana explícita** |
| RG10-16 | Las suites e2e de F3 se vuelven frágiles con el movimiento nuevo (smooth scroll, estados de opacidad) | El estado pendiente usa solo `opacity`/`transform` (Playwright `toBeVisible` no lo ve afectado) y `pointer-events` intactos; el desplazamiento suave nunca roba eventos; si una suite falla por timing del movimiento se corrige **el código**, nunca la prueba (R10-9); en caso dudoso, escala al humano (SC-008) |
| RG10-17 | **Menú hamburguesa inaccesible** (foco perdido al abrir/cerrar, Escape sin funcionar, estado no anunciado, objetivo < 44 px) — riesgo propio de los menús móviles | Patrón *disclosure* completo y único (P10-13): `aria-expanded`/`aria-controls`, foco al abrir y devuelto al cerrar, Escape + control visible, objetivos ≥ 44 px; unitarias de `PublicNav.test.tsx` + e2e `navegacion-movil.spec.ts` (ratón y teclado) + checklist de lector de pantalla en QA |
| RG10-18 | El menú móvil o el pie nuevo **rompen las pruebas congeladas de F3** (SC-008): p. ej. el `<nav>` debe seguir visible a 320/160 px, los enlaces consultados en jsdom deben seguir en el DOM, la travesía por teclado debe seguir conteniendo las etiquetas de sección | Compatibilidad **por construcción** (ver «Estrategia de navegación»): el `<nav>` nunca se oculta (solo la lista colapsa con CSS), los enlaces no se quitan del DOM (`hidden`/`aria-hidden` prohibidos), el pie usa **las mismas claves i18n** que la cabecera y un landmark con **nombre propio**; verificación temprana de `portada-accesibilidad.spec.ts` y `portada-publica.spec.ts` en cada tarea de navegación/pie; si algo falla se corrige el código o se **escala al humano**, nunca la prueba (RG10-6) |
| RG10-19 | Con el panel abierto, el **fondo se desplaza** (rueda/teclado) o el panel desborda en pantallas bajas (320×480, iOS Safari) | `html[data-nav-open='true'] { overflow: hidden }`; panel con altura máxima `calc(100dvh - var(--header-offset))` y scroll interno; e2e explícito «sin scroll de fondo»; prueba manual en iOS/Android en QA |
| RG10-20 | **Enlaces rotos** en el mapa del sitio (secciones en borrador o rutas F4–F9 aún inexistentes) | El pie se alimenta de `presentSections` (mismo criterio borrador de F3) y la regla «solo destinos existentes»: las entradas F4–F9 se añaden al modelo **cuando su ruta exista** (campo `href` futuro); unitaria con secciones ausentes + e2e «sin destinos rotos» |
| RG10-21 | El **ciclo de foco congelado** de `portada-accesibilidad.spec.ts` (`tabCycle`, **< 30 pasos** a 320 px) se desborda con los elementos nuevos (control de tema, botón hamburguesa, mapa del sitio) | Saldo neto acotado: en móvil el pie **sustituye** a las anclas de cabecera como paradas del ciclo (mismas 5 etiquetas) y solo se suman el botón hamburguesa y el control de tema; se **recuenta** al implementar cada tarea; si aun así se excede, se ajusta la interfaz (p. ej. redundancia de controles) o se escala al humano — la prueba no se toca |
| RG10-22 | El **espaciado «más aire»** se pasa: scroll largo y sensación de vacío en pantallas grandes | Tokens acotados por breakpoint (`--space-section*`, P10-15) con calibración humana en QA (SC-002/SC-011) y ajuste en **un solo sitio**; variante `--space-section-tight` para secciones densas; sin mediciones de píxeles en e2e (P10-10) |

## Complexity Tracking

Sin complejidades que compensar: 0 dependencias nuevas, 0 cambios de contrato, 0 migraciones. Los
puntos con más intrínseca —el anti-FOUC del tema, el revelado progresivo a prueba de fallos, la
**calibración del movimiento** para que sea perceptible sin ser estridente (suelo/techo de
R10-13), la **gestión de foco del menú hamburguesa móvil** y su **compatibilidad por construcción
con las pruebas congeladas de F3** (RG10-18/RG10-21)— se resuelven con mecanismos pequeños y
probados (script inline; `IntersectionObserver` + visibilidad por defecto; tokens de movimiento con
parámetros fijos y verificados en e2e; patrón *disclosure* estándar con colapso CSS y enlaces siempre
en el DOM; mapa del sitio alimentado por el modelo de secciones), documentados en `research.md`
R10-3, R10-5, R10-13, R10-14 y R10-15.
