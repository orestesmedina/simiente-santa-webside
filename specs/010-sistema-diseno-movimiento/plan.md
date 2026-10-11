# Implementation Plan: Sistema de diseño y movimiento (F10)

**Feature Branch**: `010-sistema-diseno-movimiento` · **Fecha**: 2026-10-10
**Spec**: `spec.md` — **APROBADA** (2026-10-10, con ajuste de puerta 1: panel → F10B; modo oscuro con
interruptor manual)

> **Relación con `ux.md` (fase de diseño, `disenador-ux`, en paralelo)**: la spec (FR-026/SC-013)
> exige **3 propuestas visuales distintas** aplicadas al sitio público y al modo oscuro, y el humano
> aprueba una **antes de escribir código**. La **dirección visual** la fija `ux.md`; este plan fija la
> **mecánica** —tokens, tema, movimiento, navegación, formato del selector—. La propuesta elegida
> puede cambiar composición, acentos, iconografía e ilustraciones **sin alterar** ninguna decisión de
> este plan (ver R10 y la sección «Frontera con el diseño»).

---

## Summary

F10 aplica un **sistema visual moderno y un sistema de movimiento moderado** al sitio público
existente (la portada de F3, con todas sus secciones), añade **modo oscuro con interruptor manual**
(respetando por defecto `prefers-color-scheme`, con persistencia entre visitas), rediseña la
**navegación como barra fija siempre visible** sin scroll incómodo (ni lateral de enlaces ni pérdida
al bajar) y **compacta el selector de idioma** sin tocar su comportamiento. Entrega además los
**cimientos reutilizables** (tokens con ambos temas, primitivas de componente, patrones de
movimiento, patrón de navegación y formato del selector) para que F4–F9 nazcan con el sistema.

Es un cambio **exclusivamente de frontend y de presentación**: **0 cambios** en backend, datos,
migraciones, contratos, rutas y URLs (`data-model.md`, `contracts/README.md`). El panel de
administración **no se toca** (F10B). El contrato semántico observable (ids, anclas, roles, nombres
accesibles, `localStorage['ss.lang']`, comportamiento bilingüe) queda **congelado** (R10-9) para que
las suites de F2/F3 pasen **sin cambios** (SC-008).

**Dependencias nuevas: NINGUNA** (R10-10): todo se resuelve con React, Tailwind 4 y las APIs del
navegador (`matchMedia`, `IntersectionObserver`, `ResizeObserver`).

## Technical Context

| Aspecto | Decisión |
|---|---|
| Frontend | React 19 + TypeScript `strict` + Vite + **Tailwind 4** (tokens en `frontend/src/index.css`) |
| Routing | `react-router-dom` 7 — **rutas sin cambios** (`/`, `/health`, `/panel/...`) |
| Datos | `@tanstack/react-query` — mismas queries de F3 (`usePublicHomeData`, `getPortada(lang)`) |
| i18n | Propio de F3 (`features/publico/i18n/`); solo se añaden claves nuevas (es/en) |
| Tema | `data-theme` en `<html>` + variables semánticas + variante `dark` de Tailwind 4 (R10-1) |
| Movimiento | CSS + `IntersectionObserver`; sin librería (R10-4) |
| Pruebas | Vitest + Testing Library (unitarias); Playwright (e2e local, `frontend/e2e/`) |
| Backend/BD/API | **Sin cambios** (F10 es de presentación; FR-020/FR-021) |
| Panel admin | **Sin cambios** (fuera de alcance: F10B) |

## Constitution Check

| Principio | Cumplimiento |
|---|---|
| §I La especificación manda | Plan fiel a la spec aprobada (FR-001…FR-026); huecos = 0 (todas las preguntas resueltas 2026-10-10). Los defaults revisables se marcan donde la spec los deja abiertos. |
| §II Arquitectura / dependencias | Solo frontend; **0 dependencias nuevas** (justificación negativa en R10-10). Sin cambios de contrato → `contracts/README.md` lo deja constatado («el contrato se escribe antes que el código» se cumple sin necesidad de delta). |
| §III Pruebas (test-first) | Cada tarea traerá sus pruebas: unitarias (Vitest) de tema, movimiento, navegación y selector; e2e (Playwright) de navegación fija, tema, movimiento reducido y selector; no-regresión con las suites de F2/F3 sin tocar. |
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

Mecánica **fija por este plan** (la propuesta elegida no la cambia):
tokens semánticos con dos temas (R10-1/R10-2), resolución y persistencia del tema (R10-3), reglas de
movimiento y `prefers-reduced-motion` (R10-4/R10-5), barra fija con `scroll-padding-top` medido y
anclas en filas (R10-6), selector con radios nativos y textos completos (R10-7).

Lo que **sí decide `ux.md`**: la dirección visual concreta de la propuesta aprobada (composición de
la portada, tratamiento tipográfico fino, dosis de color, iconografía/ilustraciones nuevas —permitidas
por Q6—, forma exacta del control de tema y la disposición compacta final del selector, ubicación de
secciones si se reordena la portada). Restricción dura: el Manual de Identidad manda sobre las
tendencias (FR-001) y el logotipo nunca se recolorea ni se deforma (FR-002, también en oscuro: FR-024).

## Decisiones técnicas (resumen; análisis completo en `research.md`)

Numeración **P10-1…P10-11** (de plan) con su **R10-x** de `research.md`.

| # | Decisión | Por qué (una línea) | Alternativa descartada |
|---|---|---|---|
| **P10-1** (R10-1) | Tema como **`data-theme` en `<html>`** + variables CSS **semánticas** (`--surface`, `--text`, `--border`, `--accent`…) registradas con `@theme inline` + `@custom-variant dark` de Tailwind 4 | El tema se resuelve en una capa de tokens: la mayoría de componentes ni menciona el tema; atributo autoexplicativo y extensible | `dark:` solo con `prefers-color-scheme` (sin interruptor); clase `.dark`; librería de temas; re-tematizar los tokens de F3 |
| **P10-2** (R10-2) | Tokens de F3 (`--color-navy`…) **intactos**; capa semántica **aditiva**; sitio público usa utilidades semánticas | Garantiza por construcción que el panel no cambia de aspecto (FR-020) | Re-tematizar tokens existentes (recolorearía el panel); dos hojas de estilo por tema |
| **P10-3** (R10-3) | **`ThemeProvider`** (patrón `LanguageProvider`): `localStorage['ss.theme']` (`'light'/'dark'`, ausente = sistema), `matchMedia` **en vivo**, script **inline anti-FOUC** en `index.html` | FR-023 completo: sistema por defecto, manual persistente, cambio en caliente, sin parpadeos al cargar | Resolver el tema solo en React (flash); cookie; `sessionStorage`; guardar `'system'` como valor |
| **P10-4** (R10-4) | **Movimiento CSS puro** + `Reveal` con un `IntersectionObserver` compartido; solo `opacity`/`transform` y colores | Nivel moderado (Q4) sin orquestación: **0 dependencias**, sin riesgo de descolocar maquetado | `framer-motion`/`GSAP` (dependencias, JS por frame); scroll-driven animations como base (soporte desigual) |
| **P10-5** (R10-5) | `prefers-reduced-motion` en **doble capa** (CSS global + hook `usePrefersReducedMotion` en vivo); contenido **visible por defecto**, revelado progresivo; `focusin` revela al instante | FR-006 incondicional y edge cases: JS caído → contenido visible; foco nunca sobre algo oculto | Ocultar con CSS y confiar en que JS llega (zona en blanco si falla) |
| **P10-6** (R10-6) | Navegación: **`position: sticky`** en la cabecera + `html { scroll-padding-top: var(--header-offset) }` medido con `ResizeObserver` + anclas en **`flex-wrap`** (fin del `overflow-x-auto`) | Elimina los dos casos del «scroll incómodo» (Q2) sin tapar contenido ni romper anclas; escala a F4–F9 en filas compactas | `fixed` + padding (se superpone); hamburguesa (prohibida); carrusel de enlaces (es el caso (a)) |
| **P10-7** (R10-7) | Selector: **solo formato compacto** (espaciado/disposición); radios nativos, textos completos, `useId`, `ss.lang` intactos | Decisión Q5 literal + SC-008: las pruebas de F3 pasan sin cambios | Píldoras «ES/EN», menú, banderas (rechazados por el cliente) |
| **P10-8** (R10-8) | Sistema visual público en **`frontend/src/components/publico/`**; inventario compartido **sin tocar** | El panel pinta con el inventario compartido: no se le puede cambiar un solo estilo (FR-020) | Extender variantes de los componentes compartidos (alteraría el panel) |
| **P10-9** (R10-9) | **Contrato semántico congelado**: rutas, ids/anclas, roles y nombres accesibles, `ss.lang`, comportamiento; el rediseño va «por dentro» | SC-008 exige las suites de F3 **sin cambios**: ellas son la red de no-regresión | Cambiar ids/roles y «arreglar» las pruebas (viola SC-008) |
| **P10-10** (R10-12) | Verificación de «moderno», de «menos espacio» (SC-007) y de contraste AA por **revisión humana + checklist** (`quickstart.md`), con e2e para comportamiento | La spec descarta instrumentación; las mediciones de píxeles en e2e son frágiles | e2e que mida anchos de píxel del selector; axe en runtime (fuera del patrón de F3) |
| **P10-11** (R10-11) | `data-theme` en `<html>` (anti-FOUC) pero efectivo solo en el sitio público; `ThemeProvider`/`ThemeToggle` solo en `PublicLayout` | Mantiene el panel inalterado aunque el atributo sea global | Tema por contexto React sin tocar `<html>` (no funciona el anti-FOUC) |

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
     solo bajo `prefers-reduced-motion: no-preference`).
   - Bloque global `@media (prefers-reduced-motion: reduce)` que anula transiciones/animaciones.
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

## Estrategia de movimiento (detalle)

- **Microinteracciones (FR-005/FR-008)**: transiciones cortas (≈150–200 ms) en `hover`, `:focus-visible`
  y `:active` de enlaces, botones, tarjetas y controles (color, borde, sombra, `transform:
  translateY(-1px)` como máximo). Nunca propiedades de layout → 0 desplazamientos del maquetado. El
  logotipo **queda excluido** de toda microinteracción (FR-002).
- **Revelado (FR-005/US3)**: componente `Reveal` (`features/publico/motion/`) que envuelve bloques de
  sección: fade + desplazamiento de pocos píxeles, 200–300 ms, una sola vez por elemento. Progresivo
  (R10-5): si no hay JS o hay movimiento reducido, el contenido nace visible.
- **Sin repeticiones molestas (FR-008)**: claves React estables por sección (`key` = ancla) → el cambio
  de idioma o de tema re-renderiza sin re-montar, y las secciones vistas no se re-animan.
- **Rendimiento (FR-007/SC-006)**: un `IntersectionObserver` compartido; animaciones solo del
  compositor; contenido visible e interactivo desde el primer instante (la animación nunca retrasa la
  lectura más de lo imperceptible; objetivo SC-006: ≤ 0,5 s).

## Estrategia de navegación (detalle)

- **Cabecera fija**: `SiteHeader` (extraído de `PublicLayout`/`PublicChrome`) con
  `position: sticky; top: 0; z-index` adecuado. Composición: fila de marca (logo + nombre) y controles
  (selector de idioma compacto + control de tema) + fila(s) de anclas.
- **Sin scroll incómodo**: `PublicNav` sin `overflow-x-auto`: `flex-wrap` con filas compactas y
  objetivos ≥ 44 px; a 320 px cabecera en dos filas, todo visible, sin scroll horizontal de enlaces ni
  de página (FR-011, SC-004). El e2e comprueba `scrollWidth <= clientWidth` a 320 px.
- **Anclas que no se tapan**: `--header-offset` (altura real medida) alimenta `scroll-padding-top`;
  el título de la sección destino queda visible al saltar y al bajar (FR-014).
- **Accesibilidad (FR-013)**: enlaces reales `<a href="#…">` con texto claro, orden de foco natural,
  foco visible (utilidad de foco por contexto, ya patrón de F3), `aria-label` del `nav` localizado;
  el enlace de salto «Saltar al contenido» se conserva.
- **Escala a F4–F9 (FR-012)**: el modelo de navegación sigue siendo **datos** (patrón
  `sections.ts`); las secciones futuras añaden entradas y sus rutas cuelgan de `PublicLayout`; con
  ~8–10 enlaces caben en filas compactas sin menús ocultos.

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
├── research.md          # EDITABLE (fase 0: R10-1…R10-12)
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
│   ├── diseno-movimiento.spec.ts        # NUEVO: navegación fija, sin scroll horizontal (320 px),
│   │                                    #   anclas visibles, selector compacto con comportamiento
│   │                                    #   intacto, foco visible (US1/US2/US4)
│   ├── tema-oscuro.spec.ts              # NUEVO: sistema→manual→persistencia, cambio en caliente,
│   │                                    #   logo sin recolorear, mismas funciones en ambos temas (US5)
│   ├── movimiento-reducido.spec.ts      # NUEVO (o sección del anterior): emulateMedia
│   │                                    #   reducedMotion='reduce' → contenido completo, sin
│   │                                    #   animaciones de entrada/revelado (US3, SC-005)
│   └── (portada-*.spec.ts, acceso…)     # SIN CAMBIOS — red de no-regresión de F2/F3 (SC-008)
└── src/
    ├── index.css                        # EDITADO: @custom-variant dark, tokens semánticos claro/
    │                                    #   oscuro (@theme inline), scroll-padding-top,
    │                                    #   scroll-behavior condicional, bloque global de
    │                                    #   prefers-reduced-motion, utilidades de foco y de motion
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
    │   ├── motion/                      # NUEVO — R10-4/R10-5
    │   │   ├── Reveal.tsx               #   revelado progresivo (IntersectionObserver compartido)
    │   │   ├── usePrefersReducedMotion.ts
    │   │   └── motion.test.tsx
    │   ├── i18n/
    │   │   ├── messages.ts              # EDITADO: + claves 'theme.label', 'theme.light', 'theme.dark'
    │   │   ├── es.ts / en.ts            # EDITADOS: traducciones de las claves nuevas (FR-019)
    │   │   ├── LanguageSwitcher.tsx     # EDITADO: solo formato compacto (comportamiento intacto)
    │   │   └── i18n.test.tsx            # EDITADO: + claves nuevas (exhaustividad es/en)
    │   ├── components/
    │   │   ├── PublicLayout.tsx         # EDITADO: monta ThemeProvider + SiteHeader sticky
    │   │   ├── SiteHeader.tsx           # NUEVO (extraído): cabecera fija (marca, nav, idioma, tema)
    │   │   ├── PublicNav.tsx            # NUEVO: anclas en filas flex-wrap (patrón escalable F4–F9)
    │   │   ├── PublicFooter.tsx         # EDITADO: aplica tokens semánticos; selector compacto
    │   │   ├── IdentityHero.tsx         # EDITADO: aplica sistema visual + motion moderado
    │   │   ├── (WhoWeAreSection, ScheduleSection, WhatsAppSection, ContactSection,
    │   │   │    SocialSection, SectionSkeleton, BrandLogo, SocialIcon)
    │   │   │                            # EDITADOS: tokens semánticos + reglas de composición;
    │   │   │                            #   BrandLogo SIN cambios de reglas (nunca se recolorea)
    │   │   └── *.test.tsx               # EDITADOS/NUEVOS según toque
    │   ├── pages/HomePage.tsx           # EDITADO: composición con Reveal y sistema visual
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
| Movimiento | Unitaria | `Reveal` muestra el contenido al observar; con `prefers-reduced-motion` el contenido nace visible (0 animación); foco (`focusin`) sobre pendiente → visible al instante; JS sin montar → contenido visible (fallback); sin re-animación al re-renderizar | `features/publico/motion/motion.test.tsx` |
| Navegación | Unitaria | `SiteHeader`/`PublicNav`: todas las anclas presentes, sin `overflow-x-auto`, `scroll-padding-top` actualizado con la altura medida; secciones ausentes (borrador) no dejan enlaces rotos (hereda F3) | tests de `SiteHeader`/`PublicNav` |
| Selector | Unitaria | Comportamiento idéntico a F3 (las pruebas existentes de `i18n.test.tsx` pasan **sin cambios**); formato compacto: textos completos, nombre accesible completo, sin siglas | `i18n.test.tsx` (existente) + assertions de formato |
| i18n | Unitaria | Claves nuevas (`theme.*`) presentes en `es` y `en` (exhaustividad tipada) | `i18n.test.tsx` |
| Primitivas públicas | Unitaria | `PublicButton`/`SectionHeading`/`Card`: estados hover/foco/activo accesibles (clases de foco), sin reglas que afecten al logo | `components/publico/*.test.tsx` |
| No-regresión | Unitaria | **Todas** las suites de F2/F3 pasan sin editar (incluida `BrandLogo.test.tsx`: logo sin `filter/transform/...`) | `npm test -- --run` |
| E2E navegación | Playwright | Navegación visible al bajar (barra fija); clic en ancla → título de la sección visible (no tapado); 320 px: sin scroll horizontal de página ni de enlaces, todos los enlaces alcanzables; recorrido con teclado y foco visible; selector: cambia idioma al instante, persiste al navegar y entre visitas, primera visita en español | `e2e/diseno-movimiento.spec.ts` |
| E2E tema | Playwright | Tres caminos de US5: sistema oscuro sin elección → oscuro; manual → inmediato y persistente tras recarga; primera visita → sistema. Logo con colores originales en oscuro; mismas funciones/contenido en ambos temas; control accesible por teclado | `e2e/tema-oscuro.spec.ts` |
| E2E movimiento reducido | Playwright | `emulateMedia({ reducedMotion: 'reduce' })`: contenido completo e interactivo desde el primer instante, sin animaciones de entrada/revelado (SC-005); con movimiento normal: microinteracciones presentes | `e2e/movimiento-reducido.spec.ts` |
| E2E no-regresión | Playwright | Suites de F3 (`portada-publica`, `portada-accesibilidad`, `portada-panel`, `portada-imagenes`, `portada-patch-null`) y de F2 (`acceso`, `auditoria`, `status`) pasan **sin cambios** (SC-008) | `make e2e` |
| Manual (QA + humano) | Checklist `quickstart.md` | «Moderno» y fidelidad al Manual (SC-002), 3 propuestas (SC-013), contraste AA por tema y estado (SC-012), selector con «menos espacio» (SC-007), logo sin recolorear, panel sin cambios, zoom 200 % | `quickstart.md` |

**Pruebas que no faltan** (las críticas de la spec): **0 animaciones** con `prefers-reduced-motion`
(SC-005); **barra fija visible en el 100 % del recorrido** y **0 desplazamientos horizontales** desde
320 px (SC-004); **persistencia del tema manual** en el 100 % de re-visitas y **sistema** en el 100 %
de primeras visitas (SC-012); **0 recoloreados del logo** (SC-012); **0 textos sin traducir**
(SC-010); **0 regresiones** de F2/F3 (SC-008); foco nunca sobre contenido oculto por revelado (edge
case); contenido visible si una animación falla (edge case).

## Cobertura de requisitos

| Requisito | Dónde se satisface | Se verifica con |
|---|---|---|
| FR-001/FR-002/FR-004 (sistema visual sobre el Manual, reglas del logo, legibilidad) | Tokens semánticos + `components/publico/` + secciones rediseñadas; `BrandLogo` intacto | QA visual vs. Manual (`quickstart.md`), `BrandLogo.test.tsx` |
| FR-003 (componentes base, sin dos estilos por función) | `components/publico/` + reglas de composición documentadas | Revisión visual (SC-001/SC-011) |
| FR-005…FR-009 (movimiento moderado, reduced-motion, rendimiento, sin repeticiones, nunca dependiente) | P10-4/P10-5 (`motion/`, CSS) | `motion.test.tsx`, e2e movimiento reducido (SC-005/SC-006) |
| FR-010…FR-014 (navegación clara, sin scroll incómodo, escala F4–F9, accesible, barra fija) | P10-6 (`SiteHeader`, `PublicNav`, `scroll-padding-top`) | e2e navegación (SC-003/SC-004) |
| FR-015…FR-017 (selector compacto, comportamiento F3, sin desplazamientos) | P10-7 (`LanguageSwitcher`) | `i18n.test.tsx` + e2e (SC-007) |
| FR-018 (aplicación a toda la portada + cimientos para F4–F9) | Todas las secciones + documentación de cimientos en este plan y `ux.md` | Revisión visual (SC-001/SC-011) |
| FR-019 (textos nuevos es/en) | Claves `theme.*` en `messages.ts`/`es.ts`/`en.ts` | `i18n.test.tsx` (SC-010) |
| FR-020/FR-021 (no-regresión funcional, sin datos nuevos) | Frontera F10B; `data-model.md`; contrato congelado (R10-9) | Suites F2/F3 sin cambios (SC-008) |
| FR-022 (accesibilidad en ambos temas) | Foco por contexto, orden, objetivos ≥ 44 px, temas con contraste AA | QA manual + e2e teclado (SC-009) |
| FR-023/FR-024/FR-025 (modo oscuro + control manual, contraste AA, cabecera completa a 320 px) | P10-1…P10-3, P10-11 | `theme.test.tsx`, e2e tema (SC-012) |
| FR-026 (3 propuestas y aprobación antes de código) | Proceso: `ux.md` (disenador-ux) + puerta humana | Registro en `estado.md` (orquestador; SC-013) |

## Riesgos y mitigación

| # | Riesgo | Mitigación |
|---|---|---|
| RG10-1 | Flash del tema al cargar («antes y después») | Script anti-FOUC en `index.html` (R10-3) + e2e de primer atributo correcto |
| RG10-2 | Cambio visual **del panel** por efecto colateral de tokens/estilos globales | Tokens nuevos **aditivos** (P10-2); inventario compartido sin tocar (P10-8); revisión visual del panel en QA; suites de panel sin cambios |
| RG10-3 | La barra fija tapa el destino de un ancla o contenido ampliado al 200 % | `scroll-padding-top` con altura medida en vivo (P10-6) + e2e «título visible tras el salto» |
| RG10-4 | A 320 px la cabecera (nav + idioma + tema) desborda o se tapa | Filas `flex-wrap` sin scroll lateral; composición de dos filas a 320 px; e2e con `scrollWidth <= clientWidth` y elementos alcanzables |
| RG10-5 | Contenido oculto por el revelado (JS caído, foco temprano, reduced-motion) | Revelado progresivo con contenido visible por defecto + `focusin` inmediato + doble capa reduced-motion (P10-5) |
| RG10-6 | Pruebas de F3 que codifican detalles visuales del diseño viejo | Contrato semántico congelado (R10-9); si una prueba exigiera un cambio visual, **escala al humano** (SC-008 prohíbe editarla para «hacerla pasar») |
| RG10-7 | Fallos de contraste AA en tema oscuro (estados densos, tarjetas, foco) | Tabla de contraste por par de tokens en la fase de diseño; checklist QA por tema y estado; regla: colores derivados solo para contraste (herencia de F3) |
| RG10-8 | La propuesta visual elegida desborda la mecánica del plan | La mecánica es independiente de la dirección (ver «Frontera con el diseño»); `ux.md` solo fija dirección dentro de ella |
| RG10-9 | Tirones en dispositivos modestos por demasiadas animaciones | Un `IntersectionObserver` compartido; solo `opacity`/`transform`; nivel moderado; SC-006 en el dispositivo de referencia más modesto |
| RG10-10 | El logotipo se recolorea/deforma en oscuro o en microinteracciones | `BrandLogo` excluido de motion y de temas; prueba existente intacta; checklist QA (SC-012) |
| RG10-11 | Textos nuevos sin traducir (SC-010) | Catálogo tipado `PublicMessageKey`: falta una traducción → **no compila** (patrón de F3) |
| RG10-12 | Cambio de idioma re-anima o descoloca la composición | Claves React estables (sin re-montaje); longitudes distintas es/en probadas en e2e (edge case de la spec) |

## Complexity Tracking

Sin complejidades que compensar: 0 dependencias nuevas, 0 cambios de contrato, 0 migraciones. Los dos
puntos con más intrínseca —el anti-FOUC del tema y el revelado progresivo a prueba de fallos— se
resuelven con mecanismos pequeños y probados (script inline; `IntersectionObserver` + visibilidad por
defecto), documentados en `research.md` R10-3 y R10-5.
