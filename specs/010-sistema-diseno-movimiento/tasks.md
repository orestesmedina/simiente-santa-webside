# Tasks: Sistema de diseño y movimiento (F10)

**Input**: Design documents from `/specs/010-sistema-diseno-movimiento/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/, ux.md (propuesta **B «Marea Turquesa»** aprobada), quickstart.md

**Tests**: SÍ — obligatorios (constitución §III). Cada tarea incluye sus pruebas (Vitest para componentes/lógica, Playwright para e2e). Las pruebas se escriben **antes** de la implementación y deben fallar antes de ella.

**Excepción declarada a §III (N4)**: T1001–T1004 (capas de tokens sin lógica), T1035 (documentación) y T1036 (verificación de cierre) **no** llevan pruebas automatizadas propias: su verificación es `build`/`typecheck`/`grep`/checklist manual, tal como indica cada tarea. Es la **única** excepción y queda anotada en cada una de esas tareas para que la revisión de implementación no la vuelva a plantear.

**Organization**: Fases por capa de cimiento y por historia de usuario (US1…US6), para implementar y validar incrementalmente. **F10 es 100 % frontend** (presentación): **0 tareas de backend, datos, migraciones o API** (`data-model.md`, `contracts/README.md`).

---

## Nota de alcance (no se reabre)

- F10 toca **solo** `frontend/` (sitio público). El **panel de administración no se toca en absoluto** (F10B, FR-020/SC-008): ni `features/panel|usuarios|roles|auditoria|auth|informacion`, ni el inventario compartido `frontend/src/components/` (R10-8), ni el layout del panel.
- El sistema visual público vive en una capa nueva `frontend/src/components/publico/` precisamente para no alterar el inventario compartido que pinta el panel.
- **No se incluye el panel F10B** en ninguna tarea.
- El contrato semántico observable queda **congelado** (R10-9): rutas, ids/anclas (`quienes-somos`, `horario`, `whatsapp`, `contacto`, `redes`), roles y nombres accesibles existentes (`nav.label` = «Secciones de la portada», `a11y.skip`, `lang.*`, `brand.name`…), `localStorage['ss.lang']` y comportamiento bilingüe **no cambian**. Si una prueba congelada de F3 exigiera un cambio visual, se corrige el código o **se escala al humano**: nunca se edita la prueba (RG10-6/RG10-18, SC-008).

## Decisiones de cierre y del plan que estas tareas respetan (sin reabrir)

| # | Decisión | Fuente |
|---|---|---|
| **D1** | Propuesta visual **B «Marea Turquesa»**: más aire entre secciones, mapa del sitio en el pie, menú hamburguesa en móvil, barra fija en escritorio, blobs/glass **moderados**, paleta oscura dominada por navy | `ux.md` (ronda 3), FR-026 |
| **D2** | El **selector de idioma NO se repite dentro del panel del menú móvil**: vive solo en la **cabecera** (compacto) y en el **pie** | Cierre puerta 2 (2026-10-10), `ux.md` §2.3/§0.6 |
| **D3** | Rótulos del mapa del sitio: **Explorar / Contacto / Síguenos** | Cierre puerta 2 (2026-10-10), `ux.md` §2.4 |
| **D4** | Navegación **por dispositivo**: escritorio ≥ 861 px = barra fija con anclas visibles (sin hamburguesa); móvil ≤ 860 px = **menú hamburguesa accesible** (patrón *disclosure*). El corte se hace con **CSS (media queries)**, no con JS de resize | P10-13/R10-14, `ux.md` §0.5 |
| **D5** | Movimiento **marcado/elegante con revelado al scroll**, con suelo de perceptibilidad (≥ 400 ms y ≥ 16 px por entrada) y techo (≤ 800 ms por elemento y por secuencia, stagger ≤ 240 ms). Sin parallax, sin scroll-scrub, sin librería | FR-005, P10-4/P10-12, R10-4/R10-13 |
| **D6** | **`prefers-reduced-motion` es incondicional** (doble capa CSS + JS en vivo): 0 revelados ni transiciones de entrada, contenido completo al instante | FR-006/SC-005, P10-5 |
| **D7** | Tema: `data-theme` en `<html>` + variables semánticas **aditivas** (tokens de F3 intactos); `localStorage['ss.theme']` (`'light'/'dark'`, ausente = sistema); `matchMedia` en vivo; script anti-FOUC. **El logotipo nunca se recolorea** (en oscuro se apoya en su pastilla crema) | P10-1…P10-3, P10-11, FR-023/FR-024 |
| **D8** | La cabecera al bajar transiciona **solo su superficie** (`data-scrolled`), **sin cambio de geometría** (la «compactación» de `ux.md` es de superficie: reducir altura rompería anclas y `scroll-padding-top`) | P10-6/P10-12, R10-13 (prevalece sobre `ux.md` §5.3 en este punto mecánico) |
| **D9** | Tokens de espaciado de sección: `--space-section`, `--space-section-tight`, `--space-block` (más `--pad-panel`, `--pad-text`, `--space-fluid` de `ux.md` §0.3) **por breakpoint** en `index.css` (tabla de plan P10-15/R10-16); la tabla `clamp()` de `ux.md` §0.3 es la **dirección visual** y la calibración final se cierra en QA (RG10-22, SC-002) | P10-15/R10-16 |
| **D10** | **Nomenclatura de componentes** (ver tabla siguiente): `MobileMenu`, `FooterSiteMap`, `useSiteLinks`, `Reveal`/`RevealGroup`, `useTheme`/`ThemeToggle` (como fijan `ux.md` y el encargo) + `SiteHeader`/`PublicNav` de plan P10-6. `ux.md` llama `TransitionHeader`/`SiteMapFooter` a dos de ellos: **mismo componente, nombre de plan/encargo** | Este documento |
| **D11** | Claves i18n nuevas con la convención **namespaced** existente de `messages.ts` (`nav.*`, `theme.*`, `footer.*`); los textos son los del catálogo de `ux.md` §7. Las claves **existentes no cambian ni de nombre ni de texto** | FR-019/SC-010 + contrato congelado R10-9 |
| **D12** | Las aserciones de pruebas de F3 **no se modifican**; solo se **añaden** casos nuevos (claves nuevas, formato compacto, menú, tema, movimiento) | SC-008 |
| **D13** | Sin dependencias nuevas (React 19, Tailwind 4 y APIs del navegador: `matchMedia`, `IntersectionObserver`, `ResizeObserver`) | R10-10, §II |
| **D14** | Compatibilidad por construcción con las pruebas congeladas: el `<nav aria-label>` de la cabecera **permanece visible** también a 320/160 px (dentro vive el botón hamburguesa, que da tamaño al landmark); los enlaces **siguen siempre en el DOM** (sin atributos `hidden`/`aria-hidden`/`inert` de React: jsdom debe seguir viéndolos) y el ciclo de tabulación congelado de `portada-accesibilidad.spec.ts` (`tabCycle` **< 30 pasos** a 320 px) se **recuenta** con el inventario real en T1013/T1015 (ver D16/D18 y el recuento de T1015) | RG10-18/RG10-21 |
| **D15** | **Una sola lista de enlaces en el DOM** (corrección C4): dentro del `<nav aria-label={t('nav.label')}>` vive **una única** lista (`<ul>`) alimentada por `useSiteLinks`. `MobileMenu` **no monta una lista duplicada**: es el controlador (estado `open`, ARIA, foco, scrim, scroll-lock, cruce de breakpoint) de **esa misma lista**, que en móvil se presenta como panel. Así las pruebas congeladas siguen resolviendo **una** coincidencia por enlace dentro del landmark nombrado (`homePage.test.tsx`, `portada-publica.spec.ts`) | Corrección de coherencia (analyze C4) |
| **D16** | **Colapso móvil que saca los enlaces del orden de tabulación** (corrección I2): a ≤ 860 px con el panel cerrado, la lista se oculta con **`display:none` bajo la media query** (clase tipo `hidden …`): los enlaces ocultos **no son enfocables** ni aparecen en el árbol de accesibilidad (FR-013 «el foco nunca queda sobre algo oculto»), pero **siguen en el DOM** para jsdom. Al abrir (o ≥ 861 px) la misma lista vuelve a ser visible y enfocable. Con el fallback sin JS (sin el marcador `js`) la lista se muestra siempre como filas compactas | Corrección de coherencia (analyze I2) |
| **D17** | **Menú móvil = patrón *disclosure* de APG, NO diálogo** (corrección I1): botón con `aria-expanded`/`aria-controls`; al abrir, el foco entra en el primer enlace de la lista; `Escape`, scrim, control visible o activar un enlace cierran y **devuelven el foco al botón**; **sin trampa de foco** (`Tab`/`Shift+Tab` pueden salir del panel hacia la barra de controles y el resto de la página: es una lista de enlaces normal); **sin** `role="dialog"` **ni** `aria-modal`. Se conservan el bloqueo del scroll de fondo (`data-nav-open`) y el cierre por scrim/enlace | Corrección de coherencia (analyze I1) |
| **D18** | **Mapa del sitio sin grupos vacíos** (corrección C1): un grupo **no se renderiza si no tiene ítems** (Explorar solo con al menos una sección publicada —«Inicio» forma parte de ese grupo—; Contacto solo con datos de contacto; Síguenos solo con redes) y **el mapa entero no se dibuja si no hay nada que listar**; los rótulos de grupo son **etiquetas no-heading** (`<p>`/`<span>`, **nunca** `h2`/`h3` ni ningún rol `heading`): `queryByRole('heading', { name: 'Contacto' })` de `homePage.test.tsx` resuelve por nombre a **cualquier** nivel de encabezado y con datos completos exige esa coincidencia **única** (el `h2` de la sección) | Corrección de coherencia (analyze C1) |
| **D19** | **Nombres accesibles únicos** (corrección C2, ampliada al hero): los enlaces del pie **nunca** repiten el nombre de un enlace de sección — Contacto con prefijo de contexto **«Contacto: …»** («Contacto: Correo» `mailto:`, «Contacto: Teléfono» `tel:`, «Contacto: WhatsApp» = **un solo** enlace, el primer canal; la dirección es **texto**, no enlace), Síguenos con etiquetas **planas** de `SOCIAL_NETWORK_LABELS` («Facebook», «Instagram», «YouTube»), **nunca** `social.action` ni `whatsapp.action.*` —; y el CTA de WhatsApp del hero usa un nombre propio que **no** contiene «Escribir por WhatsApp» ni «Entrar al grupo» (esos nombres pertenecen a los canales de `WhatsAppSection` y las pruebas congeladas los resuelven con **coincidencia única**) | Corrección de coherencia (analyze C2 + hallazgo propio) |
| **D20** | **«Simiente» siempre alcanzable** (corrección C3): el enlace de marca conserva su nombre completo (identidad publicada, contiene «Simiente») en **todos** los anchos; a < 480 px la palabra puede quedar solo visualmente oculta con técnica **`sr-only`** (nunca `display:none`: la sacaría del `innerText` que exige el `tabCycle` congelado a 320 px). La marca del pie sigue siendo **texto** (no añade paradas al ciclo) | Corrección de coherencia (analyze C3) |
| **D21** | **Números de movimiento: una sola fuente** (corrección M3/N2): manda la tabla de tokens de **T1004 / plan P10-12** — `--motion-duration-reveal` **600 ms**, distancia 28 px (20 px ≤ 480), stagger 80 ms con tope 240 ms, micro 180 ms, hero 700 ms, cabecera 220 ms, techo **≤ 800 ms** por elemento y por secuencia —; `ux.md` §5 es **dirección visual** (sus ±56–60 px / 0,7–0,8 s se sustituyen por los tokens) y la calibración final la cierra QA (RG10-22). El e2e mide el **mismo** techo (≤ 800 ms, sin tolerancia de 900 ms) | Corrección de coherencia (analyze M3/N2) |

## Nomenclatura (alineada con `ux.md` y `plan.md`; no se reabre)

| Nombre en código | Qué es | Equivalente en `ux.md` |
|---|---|---|
| `SiteHeader` | Cabecera fija (marca, clúster de controles, nav) | `TransitionHeader` |
| `PublicNav` | Anclas de sección en filas `flex-wrap` (≥ 861 px) + botón hamburguesa (≤ 860 px) | nav de la barra |
| `MobileMenu` | Panel del menú hamburguesa accesible (≤ 860 px) | `MobileMenu` |
| `FooterSiteMap` | Mapa del sitio del pie (Explorar / Contacto / Síguenos) | `SiteMapFooter` |
| `useSiteLinks` | Fuente única de enlaces visibles (sobre `sections.ts`/`presentSections`) | `useSiteLinks()` |
| `Reveal` / `useReveal` / `RevealGroup` | Revelado al scroll y entrada escalonada | idem |
| `usePrefersReducedMotion` / `useScrolled` | Preferencia de movimiento en vivo / sentinela de cabecera | — |
| `useTheme` / `ThemeProvider` / `ThemeToggle` | Tema claro/oscuro con persistencia | idem |
| `PublicButton` / `SectionHeading` / `Card` | Primitivas visuales del sistema público | botones / títulos / paneles de vidrio |

Claves i18n nuevas — **nomenclatura única (corrección M1)**: la tabla siguiente son las claves **finales** (T1006 las usa tal cual; `plan.md` y `ux.md` quedan alineados con ella). Las filas de `ux.md` §7 son «voz de diseño» y se **mapean** a claves existentes que **no se tocan**: `nav.groupLabel` → `nav.label` («Secciones de la portada», congelada), `skipToContent` → `a11y.skip` («Saltar al contenido principal», congelada), `lang.groupLabel` → `lang.label` (congelada). Textos es/en de `ux.md` §7:

| Clave | Español | English |
|---|---|---|
| `nav.menu.open` | Abrir menú de secciones | Open sections menu |
| `nav.menu.close` | Cerrar menú de secciones | Close sections menu |
| `nav.menu.label` | Menú de secciones | Sections menu |
| `footer.sitemap.label` | Mapa del sitio | Site map |
| `footer.sitemap.explore` | Explorar | Explore |
| `footer.sitemap.contact` | Contacto | Contact |
| `footer.sitemap.social` | Síguenos | Follow us |
| `footer.sitemap.home` | Inicio | Home |
| `theme.label` | Tema | Theme |
| `theme.toDark` | Cambiar a tema oscuro | Switch to dark theme |
| `theme.toLight` | Cambiar a tema claro | Switch to light theme |

Los nombres del grupo Contacto del pie **no** llevan clave nueva: se componen con las claves existentes + el dato («Correo: {email}», «Teléfono: {phone}», «WhatsApp: {canal}», D19) y los de Síguenos con `SOCIAL_NETWORK_LABELS` plano.

## Leyenda

| Marca | Significado |
|---|---|
| `[frontend]` `[infra]` | Capa y subagente responsable (`[frontend]` → `dev-frontend`; `[infra]` → `devops` con QA) |
| `[P1]`…`[P7]` | **Paralelizable** solo dentro de su grupo (archivos distintos, sin dependencia previa entre ellas; ver «Grupos de paralelismo») |
| *(sin `[P]`)* | Secuencial: tiene dependencias previas que deben estar integradas |
| *(FR-…, SC-…, US…)* | Requisitos de `spec.md` que cubre la tarea |

> **Orden de capas**: tokens y tema (`index.css`/`index.html`) → hooks y primitivas → navegación/pie → movimiento → aplicación de la propuesta B a las secciones → selector/tema de cabecera → e2e → cierre. Cada tarea se construye y prueba con lo que necesita la siguiente; las unitarias viajan **dentro** de cada tarea (§III) y los e2e son tareas propias por spec (Fase 8).

## Tabla resumen

| ID | Tarea | Capa | Fase | Paralelo | Depende de |
|---|---|---|---|---|---|
| T1001 | Tokens semánticos de tema claro/oscuro en `index.css` | `[frontend]` | 1 | P1 | — |
| T1002 | Script anti-FOUC + marcador `js` en `index.html` | `[frontend]` | 1 | P1 | — |
| T1003 | Tokens de espaciado de sección (`--space-section*`) en `index.css` | `[frontend]` | 1 | — | T1001 |
| T1004 | Tokens de movimiento + `prefers-reduced-motion` global + `scroll-*` en `index.css` | `[frontend]` | 1 | — | T1003 |
| T1005 | `useTheme` + `ThemeProvider` (persistencia `ss.theme`, `matchMedia` en vivo) | `[frontend]` | 1 | P2 | T1001, T1002 |
| T1006 | Claves i18n nuevas es/en (`messages.ts`, `es.ts`, `en.ts`) | `[frontend]` | 1 | P2 | — |
| T1007 | `useSiteLinks` (fuente única de enlaces visibles) | `[frontend]` | 1 | P2 | — |
| T1008 | `PublicButton` (variantes de marca + microinteracciones) | `[frontend]` | 2 | P3 | T1004 |
| T1009 | `SectionHeading` (patrón único de título de sección + acento) | `[frontend]` | 2 | P3 | T1004 |
| T1010 | `Card` (panel de vidrio por tema) | `[frontend]` | 2 | P3 | T1004 |
| T1011 | `SiteHeader` (cabecera fija, marca, clúster de controles, `<nav>` visible) | `[frontend]` | 3 | — | T1004, T1007 |
| T1012 | `PublicNav` de escritorio (anclas `flex-wrap`, sin scroll de enlaces) | `[frontend]` | 3 | — | T1011 |
| T1013 | `MobileMenu` + botón hamburguesa (≤ 860 px, patrón *disclosure*) | `[frontend]` | 3 | — | T1011, T1006 |
| T1014 | `--header-offset` (ResizeObserver) + `scroll-padding-top` efectivo + integración en `PublicLayout` | `[frontend]` | 3 | — | T1012, T1013 |
| T1015 | `FooterSiteMap` (Explorar / Contacto / Síguenos, sin enlaces rotos) | `[frontend]` | 3 | — | T1007, T1006 |
| T1016 | `PublicFooter` con la propuesta B (monta `FooterSiteMap` + franja lema/idioma) | `[frontend]` | 3 | — | T1015, T1010 |
| T1017 | `usePrefersReducedMotion` (preferencia de movimiento en vivo) | `[frontend]` | 4 | — | T1004 |
| T1018 | `Reveal` / `useReveal` (revelado progresivo, observer compartido) | `[frontend]` | 4 | — | T1017 |
| T1019 | `RevealGroup` (entrada escalonada con tope) | `[frontend]` | 4 | — | T1018 |
| T1020 | `useScrolled` + cabecera `data-scrolled` (superficie, sin cambio de geometría) | `[frontend]` | 4 | — | T1011, T1017 |
| T1021 | `IdentityHero` aplicado a la propuesta B (con entrada de hero) | `[frontend]` | 5 | — | T1008, T1009, T1010 |
| T1022 | `WhoWeAreSection` + `ScheduleSection` aplicadas (tokens, aire, jerarquía) | `[frontend]` | 5 | P4 | T1008, T1009, T1010 |
| T1023 | `WhatsAppSection` + `ContactSection` aplicadas | `[frontend]` | 5 | P4 | T1008, T1009, T1010 |
| T1024 | `SocialSection` + `SectionSkeleton` aplicadas | `[frontend]` | 5 | P4 | T1008, T1009, T1010 |
| T1025 | Capa decorativa B (blobs `aria-hidden`, glass moderado, degradable) | `[frontend]` | 5 | P4 | T1004 |
| T1026 | `HomePage`: composición con `Reveal`/`RevealGroup` (sin re-animaciones) | `[frontend]` | 5 | — | T1018, T1019, T1021…T1025 |
| T1027 | `LanguageSwitcher` compacto (comportamiento F3 intacto) | `[frontend]` | 6 | P5 | T1006 |
| T1028 | `ThemeToggle` + integración en el clúster de la cabecera | `[frontend]` | 7 | P5 | T1005, T1011 |
| T1029 | E2E `diseno-movimiento.spec.ts` (barra fija, anchos, teclado, selector) | `[frontend]` | 8 | P6 | T1026, T1027, T1028 |
| T1030 | E2E `navegacion-movil.spec.ts` (menú hamburguesa + mapa del sitio) | `[frontend]` | 8 | P6 | T1013, T1015 |
| T1031 | E2E `tema-oscuro.spec.ts` (3 caminos de US5, logo intacto) | `[frontend]` | 8 | P6 | T1005, T1028 |
| T1032 | E2E `movimiento.spec.ts` (perceptibilidad, stagger, cabecera, sin bloqueo) | `[frontend]` | 8 | P6 | T1020, T1026 |
| T1033 | E2E `movimiento-reducido.spec.ts` (0 revelados, contenido completo) | `[frontend]` | 8 | P6 | T1017, T1026 |
| T1034 | No-regresión F2/F3: suites **sin cambios** en verde + contrato congelado | `[frontend]` | 9 | — | todas |
| T1035 | Docs de cimientos para F4–F9 (`components/publico/README.md` + `quickstart.md`) | `[frontend]` | 9 | P7 | T1026 |
| T1036 | Verificación de cierre (`make ci`, `make e2e`, checklist `quickstart.md`, mapa SC) | `[infra]` | 9 | — | todas |

**Total: 36 tareas** — 35 `[frontend]` · 1 `[infra]` · **20 marcadas `[P]`** · 0 `[backend]` · 0 `[db]` (F10 es de presentación: 0 cambios de API ni de datos).

## Grupos de paralelismo

| Grupo | Tareas | Por qué pueden ir en paralelo |
|---|---|---|
| P1 | T1001 · T1002 | `frontend/src/index.css` y `frontend/index.html` son archivos distintos sin dependencias |
| P2 | T1005 · T1006 · T1007 | `features/publico/theme/`, `features/publico/i18n/` y `features/publico/hooks/useSiteLinks.ts` no se solapan |
| P3 | T1008 · T1009 · T1010 | Tres archivos nuevos en `frontend/src/components/publico/` sin dependencias entre sí |
| P4 | T1022 · T1023 · T1024 · T1025 | Componentes de sección distintos (`WhoWeAreSection`/`ScheduleSection`, `WhatsAppSection`/`ContactSection`, `SocialSection`/`SectionSkeleton`, capa decorativa); solo T1025 toca `index.css` y ya no lo comparte |
| P5 | T1027 · T1028 | `LanguageSwitcher.tsx` y `ThemeToggle.tsx`/`theme/` son archivos distintos (T1028 añade el toggle a `SiteHeader`, T1027 no lo toca) |
| P6 | T1029 · T1030 · T1031 · T1032 · T1033 | Cinco specs e2e independientes (`frontend/e2e/*.spec.ts`, un archivo cada una) |
| P7 | T1035 | Documentación, sin código de producción en el alcance de otras tareas |

---

## Fase 1 — Cimientos: tokens, tema, espaciado, movimiento, i18n y enlaces

**Purpose**: infraestructura compartida que DEBE existir antes de cualquier historia de usuario.

**⚠️ CRITICAL**: ninguna fase de US puede empezar sin esta fase completa.

- [X] T1001 · Tokens semánticos de tema claro/oscuro en `frontend/src/index.css` · `[frontend]` `[P1]`

  - **Archivos**: `frontend/src/index.css` (EDITADO). **Sin tocar** los tokens de F3 (`--color-navy`, `--color-teal`, `--color-cream`, `--color-coral`, `--color-green`, tipografías): la capa semántica es **aditiva** (P10-2, R10-2).
  - **Qué hace**: añade `@custom-variant dark (&:where([data-theme='dark'], [data-theme='dark'] *));` y el bloque de variables **semánticas** por tema bajo `html[data-theme='light']` / `html[data-theme='dark']`: `--surface`, `--surface-raised`, `--text`, `--text-muted`, `--border`, `--accent`, `--accent-strong` (turquesa de texto `#0e7a66` en claro / turquesa puro en oscuro), `--focus`, `--shadow`, `--logo-surface` (pastilla crema del logotipo) — todas derivadas de la paleta de marca respetando la proporción 70/20/10 y las reglas del logotipo (FR-002/FR-024). Las registra con `@theme inline` para que existan utilidades (`bg-surface`, `text-muted`, `border-subtle`…). Documenta la tabla de pares de contraste AA por tema en `specs/010-sistema-diseno-movimiento/quickstart.md` (referencia de QA, RG10-7). *Cubre FR-023/FR-024/FR-002*, P10-1/P10-2.
  - **Pruebas incluidas (§III)**: — (capa de tokens sin lógica; su verificación funcional llega con `theme.test.tsx` en T1005 y los e2e de tema; **excepción §III declarada**, N4). **Cómo se verifica**: `npm run build` y `npm run typecheck` en verde; con `grep` se comprueba que los tokens de F3 siguen intactos y que ninguna regla nueva afecta a `frontend/src/components/` (el panel no cambia: FR-020); las utilidades semánticas existen (`bg-surface` aparece en el CSS generado).
  - **Criterio de terminado**: la capa semántica es aditiva, el panel no cambia un solo píxel (revisión visual) y la tabla de contraste por par de tokens queda documentada.
  - **Commit sugerido**: `feat(frontend): tokens semanticos de tema claro/oscuro (F10)`

- [X] T1002 · Script anti-FOUC y marcador `js` en `frontend/index.html` · `[frontend]` `[P1]`

  - **Archivos**: `frontend/index.html` (EDITADO).
  - **Qué hace**: inserta un script **inline** (~10 líneas, sin dependencias) **antes de las hojas de estilo**: lee `localStorage['ss.theme']`; si no es `'light'|'dark'`, usa `matchMedia('(prefers-color-scheme: dark)')`; fija `document.documentElement.dataset.theme` y `style.colorScheme`; añade el marcador `js` en `<html>` (T1013 lo usa para el fallback sin JS del menú). Idempotente y silencioso (`try/catch` si `localStorage` falla → tema del sistema). *Cubre FR-023* y mitiga RG10-1 (parpadeo «antes y después» del tema al cargar).
  - **Pruebas incluidas (§III)**: sin pruebas propias (**excepción §III declarada**, N4): la comprobación automatizada vive en el e2e `tema-oscuro.spec.ts` (**T1031**: el primer atributo `data-theme` ya es el correcto en el primer pintado, sin flash) y en `movimiento-reducido.spec.ts`/`diseno-movimiento.spec.ts` (el marcador `js` está presente). **Cómo se verifica** ahora: `npm run build` en verde; con el DevTools «sin JS» el contenido se ve completo (el fallback de T1013 depende del marcador `js` ausente).
  - **Criterio de terminado**: al cargar con el sistema en modo oscuro y sin elección guardada no se ve el parpadeo del tema claro; sin JS, el sitio es usable.
  - **Commit sugerido**: `feat(frontend): script anti-FOUC del tema en index.html`

- [X] T1003 · Tokens de espaciado de sección (`--space-section*`) en `frontend/src/index.css` · `[frontend]`

  - **Archivos**: `frontend/src/index.css` (EDITADO, sobre T1001).
  - **Qué hace**: añade la capa de espaciado vertical de sección **por breakpoint** (D9, P10-15/R10-16): `--space-section` (3.25/4/5.5/7 rem), `--space-section-tight` (2/2.5/3.5/4.5 rem), `--space-block` (1.5/2/2.5/3 rem) en los cortes < 480 / 480–767 / 768–1279 / ≥ 1280 px, más los tokens de `ux.md` §0.3 (`--pad-panel`, `--pad-text`, `--space-fluid`). Los registra en `@theme inline` para que existan utilidades (`py-section`, `py-section-tight`, `gap-block`, `p-panel`). **Ninguna sección define su propio padding vertical mágico** (regla de composición para F4–F9). *Cubre el «más aire» de FR-001/FR-003/US1* (FR-026 propuesta B).
  - **Pruebas incluidas (§III)**: — (capa de tokens; **excepción §III declarada**, N4). **Cómo se verifica**: `npm run build` en verde; las utilidades existen en el CSS generado; se documenta en el README de `components/publico/` (T1035) la regla «toda sección usa estos tokens».
  - **Criterio de terminado**: el aire entre secciones se puede calibrar en **un solo sitio** (los tokens); nada del panel usa estos tokens (R10-8).
  - **Commit sugerido**: `feat(frontend): tokens de espaciado vertical de seccion (F10)`

- [X] T1004 · Tokens de movimiento + `prefers-reduced-motion` global + `scroll-*` · `[frontend]`

  - **Archivos**: `frontend/src/index.css` (EDITADO, sobre T1003).
  - **Qué hace**: (1) tokens de movimiento de la tabla de plan P10-12/R10-13 (**fuente única de números**, D21/M3) — `--motion-duration-micro` (180 ms), `--motion-duration-header` (220 ms), `--motion-duration-reveal` (**600 ms**, dentro del techo de 800 ms), `--motion-duration-hero` (700 ms), `--motion-distance` (28 px; 20 px ≤ 480 px), `--motion-stagger-step` (80 ms, tope 3 saltos = 240 máx.), `--motion-ease-out` y `--motion-ease-standard` — con la regla dura de **suelo de perceptibilidad** (toda entrada ≥ 400 ms y ≥ 16 px) documentada junto a ellos; (2) `html { scroll-padding-top: var(--header-offset, 0px); scroll-behavior: smooth }` con el suave **solo** bajo `prefers-reduced-motion: no-preference`; (3) el bloque global `@media (prefers-reduced-motion: reduce)` que anula **todo** el sistema de movimiento (transiciones y animaciones a 0 s, sin delays, sin transforms de revelado, `scroll-behavior: auto`). *Cubre FR-005/FR-006/FR-014* (D5/D6).
  - **Pruebas incluidas (§III)**: sin pruebas propias (**excepción §III declarada**, N4): las aserciones automatizadas viven en `motion.test.tsx` (T1018/T1019: suelo ≥ 400 ms / ≥ 16 px) y en los e2e **T1032** (`transition-duration` en rango **≥ 400 ms y ≤ 800 ms** —D21—, distancia ≥ 16 px) y **T1033** (con `emulateMedia({ reducedMotion: 'reduce' })`: opacidad 1, `transform: none`, sin delays, `scroll-behavior: auto`). **Cómo se verifica** ahora: `npm run build` en verde; con la preferencia de movimiento reducido activa en el navegador, ningún elemento transiciona (inspección manual rápida).
  - **Criterio de terminado**: los parámetros del sistema de movimiento están en **una sola capa calibrable** (sin valores sueltos por componente) y el movimiento reducido anula todo por CSS.
  - **Commit sugerido**: `feat(frontend): tokens de movimiento y respeto global de reduced-motion`

- [X] T1005 · `useTheme` + `ThemeProvider` (persistencia `ss.theme`, `matchMedia` en vivo) · `[frontend]` `[P2]`

  - **Archivos**: `frontend/src/features/publico/theme/ThemeProvider.tsx`, `frontend/src/features/publico/theme/useTheme.ts`, `frontend/src/features/publico/theme/theme.test.tsx` (NUEVOS). El cableado en `PublicLayout` llega con T1011/T1028.
  - **Qué hace**: contexto de tema siguiendo el patrón `LanguageProvider` de F3 (P10-3/R10-3): estado del **tema efectivo** (`'light' | 'dark'`) y de la **preferencia guardada** (`preference: 'light' | 'dark' | null`); `setTheme('light'|'dark')` aplica de inmediato `document.documentElement.dataset.theme` y `style.colorScheme` y persiste en `localStorage['ss.theme']`; **sin elección guardada** el tema es el del sistema y el listener `change` de `matchMedia('(prefers-color-scheme: dark)')` lo respeta **en caliente**; **con elección manual** prevalece la manual hasta que se cambie (edge case de la spec). Persistencia con tolerancia de F3: sin `localStorage` → tema del sistema, sin errores ni avisos. El panel **no** se envuelve (P10-11: el atributo es global pero el efecto es solo del sitio público). *Cubre FR-023/SC-012*.
  - **Pruebas incluidas (§III)**: `theme.test.tsx` (Vitest + Testing Library, con `matchMedia` simulado): primera visita sin elección → sistema (`matchMedia`); elección manual → cambio inmediato + persistida en `ss.theme`; re-visita con elección guardada → manda la manual; cambio en caliente del sistema **solo** sin elección manual; `localStorage` caído → tema del sistema sin excepciones; `data-theme`/`color-scheme` aplicados en `<html>`. **Test-first**: los casos se escriben antes y fallan contra el hook inexistente.
  - **Criterio de terminado**: los 3 caminos de US5 (sistema → manual → persistencia) pasan en unitarias; el hook es la única fuente de verdad del tema.
  - **Commit sugerido**: `feat(frontend): ThemeProvider y useTheme con persistencia y matchMedia en vivo`

- [X] T1006 · Claves i18n nuevas es/en (`messages.ts`, `es.ts`, `en.ts`) · `[frontend]` `[P2]`

  - **Archivos**: `frontend/src/features/publico/i18n/messages.ts`, `es.ts`, `en.ts` (EDITADOS), `frontend/src/features/publico/i18n/i18n.test.tsx` (EDITADO **solo añadiendo** casos: D12).
  - **Qué hace**: añade al catálogo tipado `publicMessageKeys` las claves nuevas de la tabla de nomenclatura (**nomenclatura única, M1/D11**): `nav.menu.open/close`, `nav.menu.label`, `footer.sitemap.label/explore/contact/social/home`, `theme.label`, `theme.toDark`, `theme.toLight` — con los textos es/en de `ux.md` §7. Las claves **existentes no se renombran ni cambian de texto** (contrato congelado R10-9: `nav.label` sigue siendo «Secciones de la portada», `lang.es`/`lang.en`, `lang.label`, `a11y.skip` = «Saltar al contenido principal», `brand.name`, `footer.welcome`…). **Nada de claves nuevas que choquen con el contrato**: `nav.groupLabel`, `skipToContent` y `lang.groupLabel` de `ux.md` §7 son «voz de diseño» y se resuelven con las claves congeladas `nav.label`, `a11y.skip` y `lang.label`. *Cubre FR-019/SC-010*.
  - **Pruebas incluidas (§III)**: `i18n.test.tsx` (ampliado **aditivamente**): cada clave nueva existe en `es` y `en` (el tipado `Record<PublicMessageKey, string>` ya lo garantiza por compilación; el test cubre la paridad de contenido y que ningún texto queda vacío). **Test-first**: el caso de exhaustividad se escribe antes de añadir las claves (falla por clave inexistente).
  - **Criterio de terminado**: `npm run typecheck` impide dejar una clave sin traducir; las aserciones de F3 del test no se han tocado (D12).
  - **Commit sugerido**: `feat(frontend): claves i18n nuevas de F10 (menu, mapa del sitio, tema)`

- [X] T1007 · `useSiteLinks` (fuente única de enlaces visibles) · `[frontend]` `[P2]`

  - **Archivos**: `frontend/src/features/publico/hooks/useSiteLinks.ts`, `frontend/src/features/publico/hooks/useSiteLinks.test.ts` (NUEVOS). `sections.ts` **sin cambios** (solo se tocaría si la propuesta reordena secciones; aquí no lo hace).
  - **Qué hace**: hook único que devuelve los enlaces navegables a partir de `presentSections(data)` de `features/publico/sections.ts` (mismo criterio borrador de F3: una sección sin contenido publicado **no** es enlace) con su `anchor`, su etiqueta (misma clave i18n de `PUBLIC_SECTIONS`) y, para el futuro F4–F9, un campo opcional `href` de **ruta ya implementada** (hoy vacío: no se inventan enlaces). Es la única fuente de la cabecera (`SiteHeader`/`PublicNav`/`MobileMenu`) y del pie (`FooterSiteMap`): nav y pie nunca se desincronizan ni en nombres ni en borradores (FR-012/FR-027). Puede exponer también los grupos de contacto/redes del pie (datos de `PortadaPublica`) para `FooterSiteMap`.
  - **Pruebas incluidas (§III)**: `useSiteLinks.test.ts` (Vitest): con secciones en borrador desaparecen los enlaces correspondientes (0 enlaces rotos, edge case de la spec); el orden es el de la página; los nombres salen de las mismas claves i18n que usa la cabecera; ninguna entrada F4–F9 aparece antes de que su ruta exista. **Test-first**: los casos se escriben antes que el hook.
  - **Criterio de terminado**: una sola función alimenta cabecera y pie; cambiar `sections.ts` actualiza ambos sin tocar componentes.
  - **Commit sugerido**: `feat(frontend): useSiteLinks como fuente unica de navegacion y mapa`

**Checkpoint**: cimientos listos — tokens (tema, espaciado, movimiento), tema con persistencia, claves i18n y modelo de enlaces. Puede empezar cualquier historia.

---

## Fase 2 — Primitivas del sistema visual público (US1/US6)

**Purpose**: componentes base reutilizables en `frontend/src/components/publico/` (nunca en el inventario compartido del panel, R10-8) para que **no haya dos estilos distintos para la misma función** (FR-003).

- [X] T1008 · `PublicButton` (variantes de marca + microinteracciones) · `[frontend]` `[P3]`

  - **Archivos**: `frontend/src/components/publico/PublicButton.tsx`, `frontend/src/components/publico/PublicButton.test.tsx` (NUEVOS).
  - **Qué hace**: botón/enlace de acción con variantes `primary` (turquesa), `accent` (coral, solo gráfico) y `secondary` (contorno), tipografía Poppins, objetivo táctil ≥ 44 px, foco visible (anillo turquesa 2 px) y microinteracción de `--motion-duration-micro` (hover/`:focus-visible`/`:active` con `transform: translateY(-2px)`/`scale(1.02)` máximo, **solo** `transform`/`opacity`/colores: FR-008 sin desplazar el maquetado). Usa utilidades semánticas (funciona en ambos temas). **Excluye** cualquier efecto sobre el logotipo (FR-002). *Cubre FR-003/FR-008/FR-004*.
  - **Pruebas incluidas (§III)**: `PublicButton.test.tsx`: renderiza cada variante; objetivo ≥ 44 px (clases de tamaño); clases de foco visibles; sin reglas de `filter`/`transform` sobre imágenes de marca; los estados hover/activo no cambian layout (solo clases de transición). **Test-first**.
  - **Criterio de terminado**: es el **único** estilo de botón de acción del sitio público (SC-011); heredable por F4–F9.
  - **Commit sugerido**: `feat(frontend): PublicButton, boton base del sistema publico`

- [X] T1009 · `SectionHeading` (patrón único de título de sección + acento) · `[frontend]` `[P3]`

  - **Archivos**: `frontend/src/components/publico/SectionHeading.tsx`, `frontend/src/components/publico/SectionHeading.test.tsx` (NUEVOS).
  - **Qué hace**: encabezado de sección único (kicker/número 01–05 en Bebas Neue + título en Bebas + acento turquesa animable `scaleX` 0→1, 500 ms, **deshabilitado** con movimiento reducido vía CSS de T1004). Recibe el texto por props/children (sin i18n interna: las claves las resuelve cada sección). Jerarquía de encabezado correcta (`h2` bajo el `h1` del hero). *Cubre FR-003/FR-004/FR-005*.
  - **Pruebas incluidas (§III)**: `SectionHeading.test.tsx`: renderiza kicker y título con el nivel de encabezado correcto; el acento es decorativo (`aria-hidden`); las clases de animación usan los tokens de T1004 (no valores sueltos). **Test-first**.
  - **Criterio de terminado**: todas las secciones de la portada y futuras usan este componente para sus títulos (0 jerarquías distintas).
  - **Commit sugerido**: `feat(frontend): SectionHeading, patron unico de titulo de seccion`

- [X] T1010 · `Card` (panel de vidrio por tema) · `[frontend]` `[P3]`

  - **Archivos**: `frontend/src/components/publico/Card.tsx`, `frontend/src/components/publico/Card.test.tsx` (NUEVOS).
  - **Qué hace**: superficie de contenido («panel de vidrio» de la propuesta B): radio, borde `--border`, sombra `--shadow`, fondo `--surface-raised` (translúcido en la variante `glass`), padding con `--pad-panel`, microinteracción de elevación en hover **solo** donde sea interactivo (una Card clicable) y nunca sobre imágenes de marca. Usa tokens de tema (ambos temas correctos por construcción). *Cubre FR-003/FR-024*.
  - **Pruebas incluidas (§III)**: `Card.test.tsx`: variantes `solid`/`glass`; padding desde los tokens (clases); sin `filter` sobre hijos de marca; sin eventos de ratón cuando no es interactiva. **Test-first**.
  - **Criterio de terminado**: es la única superficie de contenido del sitio público; heredable por F4–F9.
  - **Commit sugerido**: `feat(frontend): Card, superficie de contenido del sistema publico`

**Checkpoint**: primitivas listas — la portada puede vestirse con un solo sistema (US1/US6).

---

## Fase 3 — Navegación y pie (US2, Priority: P1)

**Goal**: navegar por las secciones sin esfuerzo en cualquier dispositivo: barra fija en escritorio, menú hamburguesa accesible en móvil, mapa del sitio en el pie, sin scroll incómodo.

**Independent Test**: en escritorio la barra es visible y fija y las anclas llegan a cada sección sin tapar el título; a 320 px el menú hamburguesa se abre/cierra con ratón y teclado (foco gestionado, Escape, scrim, sin scroll de fondo) y el pie lista las secciones visibles con enlaces que funcionan; 0 desplazamientos horizontales.

- [X] T1011 · `SiteHeader` (cabecera fija, marca, clúster de controles, `<nav>` visible) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/components/SiteHeader.tsx`, `frontend/src/features/publico/components/SiteHeader.test.tsx` (NUEVOS); `frontend/src/features/publico/components/PublicLayout.tsx` (EDITADO: delega la cabecera a `SiteHeader`).
  - **Qué hace**: extrae la cabecera de `PublicLayout` a un componente propio: `position: sticky; top: 0` (siempre fija, ambos modos, FR-014), fila de marca (`BrandLogo` **con sus colores originales** + nombre completo de la identidad publicada como **enlace a `/`**) y clúster de controles (slot del selector de idioma + slot del control de tema de T1028) + `<nav aria-label={t('nav.label')}>` que **permanece visible en todos los anchos** (D14: la lista de anclas colapsa, el `<nav>` nunca; dentro del `<nav>` viven el botón de T1013 y la **única** lista de enlaces, D15). **Marca a < 480 px (C3/D20)**: puede mostrar solo el monograma **visualmente**, pero el enlace de marca conserva siempre el nombre completo —que contiene «Simiente»— como texto accesible y en su `innerText`, con técnica **`sr-only`** (**nunca** `display:none`: el `tabCycle` congelado de `portada-accesibilidad.spec.ts` exige un `<a>` cuyo `innerText` incluya «Simiente» a 320 px). Mantiene el contrato congelado (R10-9): mismos `aria-label`, mismos enlaces `href="#…"`, mismas claves i18n que hoy. *Cubre FR-010/FR-014/FR-025*.
  - **Pruebas incluidas (§III)**: `SiteHeader.test.tsx`: el `<nav>` existe con el nombre accesible congelado y **está visible**; las anclas de las secciones presentes están en el DOM (y las ausentes, no: hereda F3); la cabecera renderiza el clúster de controles sin taparse (estructura de orden: marca → controles); **el enlace de marca conserva su texto completo con «Simiente»** (aserción sobre el nombre accesible y sobre el contenido de texto del enlace, sin `display:none`, D20/C3). **Test-first**.
  - **Criterio de terminado**: `PublicLayout` ya no dibuja la cabecera a mano; las e2e congeladas de F3 (`portada-publica.spec.ts`, `portada-accesibilidad.spec.ts`) siguen en verde **sin cambios**.
  - **Commit sugerido**: `refactor(frontend): SiteHeader, cabecera fija del sitio publico`

- [X] T1012 · `PublicNav` de escritorio (anclas `flex-wrap`, sin scroll de enlaces) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/components/PublicNav.tsx`, `frontend/src/features/publico/components/PublicNav.test.tsx` (NUEVOS).
  - **Qué hace**: lista de anclas de sección (de `useSiteLinks`) en filas `flex-wrap` compactas **sin `overflow-x-auto`** (fin del caso (a) del «scroll incómodo»), objetivos ≥ 44 px, foco visible, subrayado turquesa que se despliega en hover/foco (FR-008). Escala a F4–F9 (~8–10 enlaces) en una o varias filas compactas (FR-012). Es **la única lista de enlaces del `<nav>`** (D15/C4): `MobileMenu` (T1013) reutiliza estos mismos nodos y no monta una segunda lista. Debajo de 861 px la lista se colapsa con **`display:none` bajo la media query** (D16/I2, clase tipo `hidden …`): los enlaces ocultos **no son enfocables** ni están en el árbol de accesibilidad (FR-013) pero **siguen en el DOM** (sin atributos `hidden`/`aria-hidden`/`inert` de React: jsdom debe seguir viéndolos). La lógica del botón vive en T1013; con el fallback sin JS la lista se muestra siempre como filas compactas. *Cubre FR-011/FR-012/SC-004*.
  - **Pruebas incluidas (§III)**: `PublicNav.test.tsx`: todas las anclas de `useSiteLinks` presentes con sus nombres accesibles (**una sola** por sección en todo el DOM, D15); los enlaces **siguen siempre en el DOM** (sin atributos `hidden`/`aria-hidden`/`inert`, D14) y en estado cerrado llevan la clase de colapso que los oculta con `display:none` a ≤ 860 px (D16: la comprobación de que **no reciben `Tab`** es e2e, T1030); sin clases de `overflow-x-auto`; objetivos ≥ 44 px. **Test-first**.
  - **Criterio de terminado**: en escritorio no existe ningún carrusel/scroll de enlaces; con la previsión de secciones F4–F9 (inyectando entradas extra en el test) la lista sigue íntegra.
  - **Commit sugerido**: `feat(frontend): PublicNav con anclas en filas flex-wrap`

- [X] T1013 · `MobileMenu` + botón hamburguesa (≤ 860 px, patrón *disclosure*) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/components/MobileMenu.tsx`, `frontend/src/features/publico/components/MobileMenu.test.tsx` (NUEVOS); `PublicNav.tsx` (EDITADO: monta el botón **dentro** del `<nav>`, D14).
  - **Qué hace**: menú hamburguesa accesible para móvil ≤ 860 px (D4, P10-13/R10-14), patrón *disclosure* de las guías APG, sin librerías: (1) **botón** `<button>` real **dentro del `<nav>`** (es `<button>`: no compite por nombre con los enlaces, D14) con nombre accesible localizado (`nav.menu.open`/`nav.menu.close` dinámico), `aria-expanded` sincronizado, `aria-controls` al id del panel/lista, objetivo **≥ 44 px**, icono ☰→✕ **decorativo** (el nombre accesible es el texto localizado); (2) **panel = la única lista de enlaces del `<nav>`** (D15/C4): `MobileMenu` **no monta una lista duplicada**; es el **controlador** de la lista de `useSiteLinks` (mismos nodos, mismas claves i18n → mismos nombres que en escritorio), que al abrir se presenta como panel con superficie sólida por tokens, altura máxima `calc(100dvh - var(--header-offset))` con scroll interno propio, **control visible de cierre** y **scrim** que cierra al tocar fuera; cerrada a ≤ 860 px la lista está con `display:none` (D16: fuera del orden de tabulación y del árbol de accesibilidad, FR-013) y al abrir (o ≥ 861 px) vuelve a ser visible y enfocable; (3) **foco — patrón *disclosure* APG, SIN trampa de foco** (D17/I1): al abrir, el foco entra al primer enlace de la lista; `Escape` / control visible / scrim / botón cierran y **devuelven el foco al botón**; activar una sección cierra el panel y el foco continúa en el destino; `Tab`/`Shift+Tab` **pueden salir del panel** hacia la barra de controles y el resto de la página (es una lista de enlaces normal, **no** un diálogo: **sin** `role="dialog"`, **sin** `aria-modal`); (4) **sin scroll de fondo**: `html[data-nav-open='true'] { overflow: hidden }` mientras está abierto (rueda y teclado); (5) **cruce de breakpoint** (≥ 861 px): el panel se cierra solo, `aria-expanded` vuelve a `false`, reaparecen las filas de anclas, sin estados residuales; (6) **sin gestos** (solo clic/toque y teclado); (7) **fallback sin JS** (sin el marcador `js` de T1002): las anclas se pintan como filas compactas visibles a cualquier ancho y el botón no se renderiza; (8) transiciones del panel con `--motion-duration-micro` y **apertura/cierre instantáneos** con movimiento reducido. **D2: el panel NO contiene el selector de idioma** (solo enlaces de secciones). *Cubre FR-010/FR-013/FR-025/US2 esc. 2–3–7, SC-003/SC-004/SC-009*.
  - **Pruebas incluidas (§III)**: `MobileMenu.test.tsx` (Vitest + TL): `aria-expanded` sincronizado y `aria-controls` apuntando al id del panel/lista; nombre accesible localizado del botón; al abrir, foco en el primer enlace y **sin trampa de foco** (`Tab` puede salir del panel —D17—; aserción de que **no** hay `role="dialog"` ni `aria-modal`); `Escape`, control visible y scrim cierran y devuelven el foco al botón; cierre al activar una sección; cierre al superar 860 px (evento simulado); `data-nav-open` bloquea el scroll de fondo; con `prefers-reduced-motion` el panel no anima; el panel **no** contiene el selector de idioma (D2); **no se renderiza una segunda lista de enlaces** (D15: `within(nav)` sigue resolviendo una única coincidencia por enlace) y con el menú cerrado la lista lleva la clase de `display:none` (D16; la exclusión del orden de tabulación la comprueba el e2e **T1030**). **Test-first** (los casos se escriben antes que `MobileMenu`).
  - **Criterio de terminado**: el ciclo completo abrir→navegar→cerrar funciona con ratón y con teclado; se **recuenta** el ciclo de tabulación de `portada-accesibilidad.spec.ts` (`tabCycle` < 30 pasos a 320 px, RG10-21 — el inventario de paradas está en **T1015**, corrección I3) y sigue cumpliendo **sin editar la prueba**.
  - **Commit sugerido**: `feat(frontend): MobileMenu, menu hamburguesa accesible en movil`

- [X] T1014 · `--header-offset` (ResizeObserver) + `scroll-padding-top` efectivo + integración · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/components/SiteHeader.tsx` (EDITADO), `frontend/src/features/publico/components/PublicLayout.tsx` (EDITADO), tests de ambos (EDITADOS **aditivamente**).
  - **Qué hace**: mide la altura real de la cabecera con `ResizeObserver` (con el menú cerrado: el panel es overlay y no altera la altura) y la publica como `--header-offset` en `<html>` para que el `scroll-padding-top` de T1004 impida que la barra tape el destino de un ancla (título de sección siempre visible tras el salto: FR-014/RG10-3). Termina la integración: `PublicLayout` monta `ThemeProvider` (T1005) + `SiteHeader` + `main` + `PublicFooter`, pasando las secciones presentes. *Cubre FR-014/FR-025*, edge case «la barra fija no tapa el destino».
  - **Pruebas incluidas (§III)**: tests unitarios: `--header-offset` se actualiza cuando cambia la altura medida (observer simulado) y no se recalcula con el menú abierto; la integración de `PublicLayout` conserva el skip-link, el `main id="contenido"` y el orden de landmarks (revisión de `layout.test.tsx`/`homePage.test.tsx` existentes **sin cambios**, D12). El e2e «título visible tras el salto» lo cubre **T1029**.
  - **Criterio de terminado**: saltar a una sección deja su título visible (no tapado) con y sin zoom 200 %; las suites congeladas siguen en verde.
  - **Commit sugerido**: `feat(frontend): scroll-padding-top medido desde la cabecera fija`

- [X] T1015 · `FooterSiteMap` (Explorar / Contacto / Síguenos, sin enlaces rotos) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/components/FooterSiteMap.tsx`, `frontend/src/features/publico/components/FooterSiteMap.test.tsx` (NUEVOS).
  - **Qué hace**: mapa del sitio dentro del `<footer>` existente (sin crear otro landmark `contentinfo`): `<nav aria-label={t('footer.sitemap.label')}>` con **nombre accesible propio y distinto** del de la cabecera (dos `navigation` sin nombres distintos romperían las consultas congeladas) y los **3 grupos con su rótulo visible** (D3): **Explorar** (`footer.sitemap.explore`: «Inicio» + las secciones de `useSiteLinks`), **Contacto** (`footer.sitemap.contact`: dirección, correo, teléfono, WhatsApp — mismos datos del bloque 04) y **Síguenos** (`footer.sitemap.social`: Facebook, Instagram, YouTube — mismos enlaces del bloque 05).
    - **Sin grupos vacíos (D18/C1)**: un grupo **no se renderiza si no tiene ítems** (Explorar solo con al menos una sección publicada —«Inicio» forma parte de ese grupo—, Contacto solo con datos de contacto, Síguenos solo con redes) y **el mapa entero no se dibuja si no hay nada que listar** (respuesta vacía de `homePage.test.tsx` no deja ni rastro del mapa).
    - **Rótulos de grupo = etiquetas no-heading (D18/C1)**: `<p>`/`<span>` en versalitas turquesa, **nunca** `h2`/`h3` — `queryByRole('heading', { name: 'Contacto' })` de `homePage.test.tsx` resuelve por nombre a **cualquier** nivel de encabezado y «Contacto» colisionaría con el `h2` de la sección 04 —.
    - **Nombres accesibles únicos (D19/C2)**: los enlaces del pie **nunca** repiten el nombre de un enlace de sección. Explorar reutiliza los nombres de sección (claves `nav.sections.*`, iguales que la cabecera, sin consultas congeladas por nombre global). Contacto: la dirección es **texto** (no enlace) y los enlaces llevan prefijo de contexto **«Contacto: …»** — «Contacto: Correo» (`mailto:`), «Contacto: Teléfono» (`tel:`), «Contacto: WhatsApp» (un **solo** enlace: el primer canal publicado) —; jamás el dato desnudo como nombre. Síguenos: etiquetas **planas** de `SOCIAL_NETWORK_LABELS` («Facebook», «Instagram», «YouTube»), **nunca** `social.action` («Ver en Facebook») ni `whatsapp.action.*`. Ningún nombre del pie contiene «Escribir por WhatsApp», «Entrar al grupo» ni «Ver en …» (consultas congeladas de coincidencia única en `homePage.test.tsx`).
    - Listas `<ul>` reales, enlaces reales `href`, objetivos ≥ 44 px, columnas en escritorio (grid) y **apilado** en móvil (orden marca → Explorar → Contacto → Síguenos → franja de idioma). **Regla «sin enlaces rotos» (FR-027/SC-015)**: solo destinos existentes — secciones con contenido publicado (mismo criterio que la cabecera) y, para F4–F9, entradas con `href` de **ruta ya implementada** (el componente no cambia: se añade una columna/grupo cuando la sección exista). *Cubre FR-027/FR-012/FR-013*.
  - **Recuento del ciclo de tabulación a 320 px, menú cerrado (corrección I3, RG10-21)**: el `tabCycle` congelado de `portada-accesibilidad.spec.ts` exige **< 30 pasos** (los pasos = paradas + 1 del cierre del ciclo). Inventario con la cabecera nueva (D15–D17), el pie nuevo y los CTAs del hero (D19):

    | Zona | Paradas | Detalle |
    |---|---|---|
    | Cabecera | 5 | skip-link, enlace de marca (con «Simiente», D20), radio `es` del selector (el grupo de radios = 1 parada), `ThemeToggle`, botón hamburguesa |
    | Hero | 2 | CTA «Ver horario» + CTA de WhatsApp con nombre propio (D19) |
    | Secciones | 1–6 | WhatsApp: 1 canal (el seed de la prueba congelada borra los canales antiguos); Contacto: `mailto` + `tel` (la dirección es texto); Redes: 0–3 chips |
    | Pie · franja | 1 | radio `es` del segundo selector |
    | Pie · mapa | 10–13 | Explorar: «Inicio» + hasta 5 secciones (6); Contacto: correo + teléfono + WhatsApp (3); Síguenos: 0–3 |
    | **Total** | **~20–26** | **~21–27 pasos** → **< 30 ✓** (margen ≥ 3) |

    El saldo neto frente a F3 es de **+6…+8 paradas** (el pie nuevo y los controles nuevos **sustituyen** a las 5 anclas de cabecera, que a ≤ 860 px dejan de ser tabulables — D16). Si QA midiera ≥ 28 pasos con datos residuales, se aligera el pie en este orden: (1) «Inicio» del grupo Explorar como texto/no renderizado, (2) el WhatsApp del pie como texto — la dirección ya es texto —. **Nunca** se edita la prueba (RG10-6).
  - **Pruebas incluidas (§III)**: `FooterSiteMap.test.tsx`: lista exactamente las secciones presentes con los **mismos nombres** que la cabecera (mismas claves i18n, D11); con una sección en borrador no aparece su enlace (ni en el pie ni en la cabecera); no renderiza ninguna ruta F4–F9 inexistente (dada una entrada con `href` no implementado, se omite); landmark con nombre accesible propio (`footer.sitemap.label`); rótulos de grupo **Explorar / Contacto / Síguenos** (D3) **sin rol `heading`** (D18/C1); con datos vacíos **no** se renderiza el mapa y un grupo sin ítems no deja rastro (D18/C1); **ningún nombre accesible del pie duplica** el de otro enlace del documento (prefijos en Contacto, planos en Síguenos; dirección como texto) (D19/C2). **Test-first**.
  - **Criterio de terminado**: el mapa lista el 100 % de las secciones visibles y 0 secciones sin publicar; con la previsión F4–F9 (entradas extra en el test) sigue sin enlaces rotos.
  - **Commit sugerido**: `feat(frontend): FooterSiteMap, mapa del sitio en el pie`

- [X] T1016 · `PublicFooter` con la propuesta B (monta `FooterSiteMap` + franja lema/idioma) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/components/PublicFooter.tsx` (EDITADO), `frontend/src/features/publico/components/PublicFooter.test.tsx` (NUEVO o ampliado **aditivamente**).
  - **Qué hace**: aplica el sistema visual al pie (tokens semánticos, `--pad-panel`, `--space-section-tight`): marca + lema (`footer.welcome`, Playfair itálica), monta `FooterSiteMap` (T1015) y conserva la franja inferior con el lema y el **selector de idioma** (comportamiento F3 intacto, US4 esc. 5). El `<footer>`/landmark `contentinfo` de F3 se conserva tal cual (R10-9). **La marca del pie sigue siendo texto, no enlace** (D20/C3): el enlace cuyo texto incluye «Simiente» lo garantiza la cabecera (T1011) y así el pie no añade paradas al ciclo de tabulación (I3). *Cubre FR-027/FR-003/US2 esc. 8*.
  - **Pruebas incluidas (§III)**: `PublicFooter.test.tsx`: renderiza el mapa del sitio dentro del `footer` existente (un único `contentinfo`); la franja inferior contiene el selector de idioma; contraste/objetivos ≥ 44 px por clases (la auditoría congelada del `footer` la cubre el e2e). **Test-first**.
  - **Criterio de terminado**: el pie completo cumple la maqueta B (columnas en escritorio, apilado en móvil) y las pruebas congeladas del footer siguen en verde.
  - **Commit sugerido**: `feat(frontend): PublicFooter con mapa del sitio y franja de idioma`

**Checkpoint**: US2 navegable y testeable por sí sola — barra fija (escritorio), menú hamburguesa accesible (móvil), mapa del sitio, sin scroll incómodo.

---

## Fase 4 — Movimiento (US3, Priority: P2)

**Goal**: movimiento **percibido de verdad** —revelado al scroll, entrada escalonada, microinteracciones claras— sin molestar y con `prefers-reduced-motion` incondicional.

**Independent Test**: en modo normal las secciones se revelan claramente al entrar en pantalla (duración y distancia sobre el suelo de perceptibilidad); con `prefers-reduced-motion` no se revela nada y todo funciona igual.

- [X] T1017 · `usePrefersReducedMotion` (preferencia de movimiento en vivo) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/motion/usePrefersReducedMotion.ts`, `frontend/src/features/publico/motion/motion.test.tsx` (NUEVOS).
  - **Qué hace**: hook que expone `prefers-reduced-motion` vía `matchMedia('(prefers-reduced-motion: reduce)')` con listener `change` (se respeta **en caliente**, sin recargar: edge case de la spec). Es la **segunda capa** de la doble capa de D6 (la primera es el CSS global de T1004). *Cubre FR-006/SC-005*.
  - **Pruebas incluidas (§III)**: `motion.test.tsx` (primeros casos): el valor inicial sigue la media; un cambio de la media se refleja sin recargar; sin `matchMedia` (entornos antiguos/jsdom) degrada a «con movimiento» sin excepciones. **Test-first**.
  - **Criterio de terminado**: `Reveal`/`RevealGroup`/`MobileMenu`/cabecera pueden consultar la preferencia en vivo.
  - **Commit sugerido**: `feat(frontend): usePrefersReducedMotion con matchMedia en vivo`

- [X] T1018 · `Reveal` / `useReveal` (revelado progresivo, observer compartido) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/motion/Reveal.tsx`, `frontend/src/features/publico/motion/useReveal.ts` (NUEVOS), `motion.test.tsx` (EDITADO **aditivamente**).
  - **Qué hace**: envoltorio de revelado al scroll (P10-4/R10-5): `IntersectionObserver` **compartido** (`rootMargin: 0px 0px -10% 0px` o umbral equivalente), entradas con `opacity` + `translateY(var(--motion-distance))` (o lateral según `from`), duración `--motion-duration-reveal` (≥ 400 ms, ≤ 800 ms), curva `--motion-ease-out`, **una sola vez por elemento y por carga**. **Progresivo (R10-5)**: el contenido **nace visible** por defecto; el estado pendiente solo se aplica cuando JS está montado y no hay movimiento reducido; con JS caído, con movimiento reducido o si el observer no dispara (salvavidas de ~1 s) el contenido se ve igualmente; `focusin` sobre un elemento pendiente lo **revela al instante** (el foco nunca queda sobre algo oculto). Solo `opacity`/`transform` (compositor, FR-007), `will-change` temporal. *Cubre FR-005/FR-007/FR-009/SC-014* y los edge cases «JS caído», «foco temprano» y «animación fallida».
  - **Pruebas incluidas (§III)**: `motion.test.tsx`: al observar, el elemento pasa de pendiente a revelado con fade + `translateY` y **parámetros ≥ suelo** (duración ≥ 400 ms, distancia ≥ 16 px — aserción sobre clases/estilos, SC-014); una sola vez (re-renderizar no re-anima); con `prefers-reduced-motion` el contenido **nace visible** y no se observa; `focusin` revela al instante; si el observer no dispara, el salvavidas revela todo. **Test-first**.
  - **Criterio de terminado**: el revelado es perceptible y verificable (parametrizado, no «a ojo») y nunca oculta contenido necesario (FR-009).
  - **Commit sugerido**: `feat(frontend): Reveal, revelado progresivo al scroll`

- [X] T1019 · `RevealGroup` (entrada escalonada con tope) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/motion/RevealGroup.tsx` (NUEVO), `motion.test.tsx` (EDITADO **aditivamente**).
  - **Qué hace**: reparte `--reveal-index` entre los hijos (2–4 elementos: tarjetas de una fila, ítems de una lista) → `transition-delay` de `--motion-stagger-step` (**80 ms**, D21) por salto, **tope 3 saltos (240 ms máx.)** y secuencia total del bloque ≤ 800 ms (D5). Con movimiento reducido, sin delays. *Cubre FR-005/SC-014* («entrada escalonada cuando convenga»).
  - **Pruebas incluidas (§III)**: `motion.test.tsx`: los hijos reciben índices 0,1,2,3(tope) con delays acotados; más de 4 hijos no amplía el tope; con movimiento reducido no hay delays. **Test-first**.
  - **Criterio de terminado**: el escalonado es perceptible pero acotado (nunca «secuencia larga»).
  - **Commit sugerido**: `feat(frontend): RevealGroup, entrada escalonada acotada`

- [X] T1020 · `useScrolled` + cabecera `data-scrolled` (superficie, sin cambio de geometría) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/motion/useScrolled.ts` (NUEVO), `frontend/src/features/publico/components/SiteHeader.tsx` (EDITADO), `motion.test.tsx` y `SiteHeader.test.tsx` (EDITADOS **aditivamente**).
  - **Qué hace**: sentinela observada con `IntersectionObserver` (sin listener de `scroll`) que activa `data-scrolled` en la cabecera al bajar y lo quita al volver arriba. La cabecera transiciona **solo su superficie** —fondo translúcido → vidrio (`backdrop-filter` **acotado a la cabecera**, borde hairline turquesa, sombra suave), `--motion-duration-header` (220 ms)— **sin cambio de geometría** (D8: misma altura → cero riesgo para anclas y `scroll-padding-top`). Con movimiento reducido, cambio instantáneo. Si DevTools muestra coste de pintura del blur en el dispositivo modesto de referencia, se degrada a superficie sólida sin `backdrop-filter` (FR-007/RG10-9). *Cubre FR-014/FR-005* («cabecera con transición al bajar»).
  - **Pruebas incluidas (§III)**: `motion.test.tsx`: `data-scrolled` se activa al bajar y se quita al volver (sentinela simulada) y **sin listeners de `scroll`**; con movimiento reducido el cambio es instantáneo; `SiteHeader.test.tsx`: la altura/clases de layout no cambian entre estados (solo superficie). **Test-first**.
  - **Criterio de terminado**: la transición de cabecera es perceptible (SC-014) y el `scroll-padding-top` sigue correcto (geometría constante).
  - **Commit sugerido**: `feat(frontend): cabecera con transicion de superficie al bajar`

**Checkpoint**: el sistema de movimiento completo y calibrable (parámetros en tokens, doble capa reduced-motion, sin librería).

---

## Fase 5 — Propuesta B «Marea Turquesa» aplicada a la portada (US1, Priority: P1) 🎯 MVP

**Goal**: la portada completa (todas las secciones de F3) vestida con el sistema: paleta oscura navy dominante, paneles de vidrio moderados, aire entre secciones, tipografía de marca y microinteracciones; ambos temas.

**Independent Test**: recorrer la portada publicada muestra el mismo sistema en las 8 secciones (cabecera, hero, quiénes somos, horario, WhatsApp, contacto, redes, pie), legible con contenidos reales de distinta forma, en 320 px / 768 px / 1280 px, en tema claro y oscuro, y con logotipo intacto.

- [X] T1021 · `IdentityHero` aplicado a la propuesta B (con entrada de hero) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/components/IdentityHero.tsx` (EDITADO), `frontend/src/features/publico/components/BrandLogo.tsx` (**sin cambios**), `IdentityHero.test.tsx` (NUEVO o ampliado aditivamente).
  - **Qué hace**: hero de identidad con la composición B: titular en Bebas Neue, lema en Playfair itálica, CTAs con `PublicButton` — «Ver horario» (ancla a `#horario`) y un CTA hacia WhatsApp con **nombre propio que no contiene «Escribir por WhatsApp» ni «Entrar al grupo»** (p. ej. «Escríbenos por WhatsApp»): esos dos nombres pertenecen a los canales de `WhatsAppSection` y `homePage.test.tsx` los resuelve con **coincidencia única** (D19/C2; microcopy ajustada respecto a la maqueta B: manda el contrato congelado, SC-008) —, imagen de portada en marco de vidrio (`Card glass`), texto izquierda / imagen derecha en escritorio y columna en móvil, padding con `--space-section-tight`/`--pad-panel`, y **entrada de hero** de `--motion-duration-hero` (solo la primera carga real: claves React estables, FR-008). El logotipo se muestra **con sus colores originales** sobre `--logo-surface` (pastilla crema) en ambos temas, sin `filter`/inversión (FR-002/FR-024; la prueba `BrandLogo.test.tsx` que lo prohíbe se conserva intacta). Sin JS/animación, el hero se ve completo desde el primer instante (FR-007/FR-009). *Cubre FR-001/FR-002/FR-004/US1*.
  - **Pruebas incluidas (§III)**: tests del hero: estructura y textos sin cambios respecto a F3 (contrato congelado: `a11y.hero`, ids); CTAs con `PublicButton` y **sin colisión de nombres accesibles** con la sección WhatsApp (aserción: ningún enlace del hero contiene «Escribir por WhatsApp» ni «Entrar al grupo» en su nombre accesible, D19/C2); clases de tokens (no colores explícitos `bg-white`/`text-navy`); el logotipo no recibe clases de animación ni de tema. **Test-first** (aserciones nuevas).
  - **Criterio de terminado**: el hero cumple la maqueta B en ambos temas; 0 reglas de recoloreo sobre el logotipo.
  - **Commit sugerido**: `feat(frontend): IdentityHero con la propuesta B (F10)`

- [X] T1022 · `WhoWeAreSection` + `ScheduleSection` aplicadas · `[frontend]` `[P4]`

  - **Archivos**: `frontend/src/features/publico/components/WhoWeAreSection.tsx`, `ScheduleSection.tsx` (EDITADOS) + sus tests (NUEVOS/ampliados aditivamente).
  - **Qué hace**: aplica el sistema (tokens semánticos en vez de colores explícitos, `Card`/`SectionHeading`, `--space-section` entre secciones, `--space-block` título→contenido, jerarquía y legibilidad para todas las edades FR-004) a «quiénes somos» (panel de vidrio) y al horario (hora en Bebas turquesa grande, filas legibles, densidad útil conservada). **Contrato congelado del horario (corrección M2)**: se conservan **ambas** representaciones con su breakpoint actual — la `table` (`hidden … sm:table`) con `caption` `sr-only` cuyo **nombre accesible sigue siendo «Horario de servicios»** (lo exige `portada-publica.spec.ts`) y la lista de tarjetas `ul` (`sm:hidden`, consultada como `#horario li` por `portada-accesibilidad.spec.ts` a 160 px) —; solo cambia el vestido (tokens, tipografía, espaciado). Contenidos reales de distinta forma (texto largo, un servicio suelto) sin huecos ni desbordes. *Cubre FR-001/FR-003/FR-004/SC-001*.
  - **Pruebas incluidas (§III)**: tests de cada sección: estructura/contenido sin cambios (contrato congelado) —incluida la **representación dual** del horario: la `table` conserva el nombre accesible «Horario de servicios» y siguen existiendo la `ul`/`li` de la representación móvil (M2)—; clases de tokens; sin estilos propios de padding vertical mágico (usan `py-section`/`py-section-tight`); texto largo y lista de 1 servicio renderizan sin romper la composición (clases de layout). **Test-first**.
  - **Criterio de terminado**: ambas secciones aplican el sistema (0 excepciones: SC-001) y sus tests de F3 siguen en verde sin cambios.
  - **Commit sugerido**: `feat(frontend): Quienes somos y Horario con el sistema visual (F10)`

- [X] T1023 · `WhatsAppSection` + `ContactSection` aplicadas · `[frontend]` `[P4]`

  - **Archivos**: `WhatsAppSection.tsx`, `ContactSection.tsx` (EDITADOS) + tests.
  - **Qué hace**: sistema aplicado a WhatsApp (CTA verde de marca con `PublicButton`, microinteracción marcada) y contacto (filas-pastilla: dirección/correo/teléfono con desplazamiento de 6 px en hover, `transform` only, FR-008; datos de F3 intactos). `--space-section-tight` si se encadenan como bloque denso (D9). Mismos contratos y textos de F3. *Cubre FR-001/FR-003/FR-004/SC-001*.
  - **Pruebas incluidas (§III)**: tests: enlaces `tel:`/`mailto:`/WhatsApp sin cambios; clases de tokens; filas con objetivo ≥ 44 px; microinteracción sin desplazamiento de layout. **Test-first**.
  - **Criterio de terminado**: mismas funciones y datos que F3, aspecto B, ambos temas correctos.
  - **Commit sugerido**: `feat(frontend): WhatsApp y Contacto con el sistema visual (F10)`

- [X] T1024 · `SocialSection` + `SectionSkeleton` aplicadas · `[frontend]` `[P4]`

  - **Archivos**: `SocialSection.tsx`, `SectionSkeleton.tsx`, `SocialIcon.tsx` (EDITADOS solo si hace falta) + tests.
  - **Qué hace**: redes como chips con elevación/relleno en hover (`transform`/`opacity`), `SocialIcon` conservando su contrato (los iconos de marca no se recolorean fuera de la paleta: FR-002); `SectionSkeleton` re-pensada para B: paneles de vidrio fantasmas con las medidas y separaciones de la sección destino y pulso sutil **solo si hay movimiento** (el estado de carga nunca depende de la animación). *Cubre FR-001/FR-003/FR-009/SC-001*.
  - **Pruebas incluidas (§III)**: tests: chips ≥ 44 px con nombres accesibles de F3 intactos; skeleton con las mismas clases de espaciado que las secciones destino (no «brinca» al cargar); el pulso usa tokens de movimiento y queda anulado con movimiento reducido. **Test-first**.
  - **Criterio de terminado**: redes y carga con el sistema; 0 cambios de contenido.
  - **Commit sugerido**: `feat(frontend): Redes y skeleton con el sistema visual (F10)`

- [X] T1025 · Capa decorativa B (blobs `aria-hidden`, glass moderado, degradable) · `[frontend]` `[P4]`

  - **Archivos**: `frontend/src/features/publico/components/GradientBlobs.tsx` (NUEVO), `frontend/src/index.css` (EDITADO: estilos de blobs/glass si no caben como utilidades), `GradientBlobs.test.tsx` (NUEVO).
  - **Qué hace**: los 2 blobs de gradiente de la propuesta B en deriva continua al fondo (solo `transform`/`opacity` animados, 26–32 s alternos), **decorativos** (`aria-hidden`, nunca tras texto esencial con contraste insuficiente), con blur **estático moderado y acotado** (es el mayor coste de B) y regla de degradación (RG10-9/`ux.md` §9.4): si el dispositivo de referencia modesto muestra tirones o tareas largas > 50 ms atribuibles a los blobs, se reduce el blur o se retiran **antes** que entorpecer la lectura (FR-007). Con `prefers-reduced-motion`: **sin blobs animados** (estáticos o ausentes). Sin parallax ni scroll-scrub (D5). *Cubre FR-001/FR-005/FR-007* (dirección visual de B).
  - **Pruebas incluidas (§III)**: `GradientBlobs.test.tsx`: los blobs son decorativos (`aria-hidden`), no interceptan eventos (`pointer-events: none`), no animan con movimiento reducido y no aplican `filter` sobre contenido de marca. **Test-first**.
  - **Criterio de terminado**: la atmósfera B se percibe sin coste de fluidez en el dispositivo de referencia (checklist `quickstart.md`) y degrada limpio.
  - **Commit sugerido**: `feat(frontend): capa decorativa de la propuesta B (blobs y glass)`

- [X] T1026 · `HomePage`: composición con `Reveal`/`RevealGroup` (sin re-animaciones) · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/pages/HomePage.tsx` (EDITADO), `homePage.test.tsx` (EDITADO **aditivamente**).
  - **Qué hace**: envuelve los bloques de cada sección con `Reveal` (variante lateral alternada en escritorio / vertical en móvil) y sus componentes internos con `RevealGroup` (stagger acotado), aplica el acento animado de `SectionHeading`, y garantiza **claves React estables por sección** (`key` = ancla): cambiar de idioma o de tema re-renderiza **sin re-montar**, las secciones vistas no se re-animan y el hero solo entra en la primera carga (FR-008/SC-014 «0 rebotes ni parpadeos»). El salto por ancla combina desplazamiento suave + revelado del destino. *Cubre FR-005/FR-008/SC-014* y el edge case «cambio de idioma/tema no re-anima».
  - **Pruebas incluidas (§III)**: `homePage.test.tsx` (ampliado): las secciones montan con `Reveal`; cambiar el idioma o el tema **no** re-monta las secciones (mismas instancias/`key` estables: se verifica con contadores de render o claves); las secciones ausentes siguen sin pintarse (SC-012 de F3). **Test-first**.
  - **Criterio de terminado**: el recorrido de la portada se revela de forma perceptible y sin repeticiones molestas; con movimiento reducido la composición es la misma sin animación.
  - **Commit sugerido**: `feat(frontend): HomePage con revelado al scroll y entrada escalonada`

**Checkpoint**: **MVP** — la portada completa con la propuesta B, ambos temas, navegación, pie con mapa y movimiento. Validable con el cliente (SC-002/SC-014 en revisión).

---

## Fase 6 — Selector de idioma compacto (US4, Priority: P2)

- [X] T1027 · `LanguageSwitcher` compacto (comportamiento F3 intacto) · `[frontend]` `[P5]`

  - **Archivos**: `frontend/src/features/publico/i18n/LanguageSwitcher.tsx` (EDITADO), `frontend/src/features/publico/i18n/i18n.test.tsx` (EDITADO **solo añadiendo** aserciones de formato, D12).
  - **Qué hace**: **solo formato y huella visual** (Q5/FR-015): textos completos actuales (`lang.es`/`lang.en`) en 0.82–0.85 rem con espaciado/disposición más ajustados (pastilla fina con borde translúcido, activo con fondo turquesa suave y peso 700), acomodo en el clúster de controles de la cabecera y en la franja del pie. **Sin** píldoras «ES/EN», **sin** menú desplegable, **sin** banderas ni siglas como identificación principal. El cuerpo de texto no baja del umbral de legibilidad (FR-004) y cada opción compensa con padding hasta ≥ 44 px de objetivo. **Comportamiento idéntico a F3** (FR-016): radios nativos, `useId`, `ss.lang`, cambio inmediato, memoria entre visitas, primera visita en español, sin auto-detección. **D2: no se añade al panel del menú móvil**. *Cubre FR-015/FR-016/FR-017/SC-007*.
  - **Pruebas incluidas (§III)**: las pruebas **existentes** de `i18n.test.tsx` pasan **sin editarlas** (comportamiento intacto: D12); se **añaden**: los textos visibles son los nombres completos (sin siglas), cada opción tiene nombre accesible completo, el elemento ocupa menos espacio que antes (aserción de clases de espaciado — **no** medición de píxeles, P10-10), y no provoca desplazamiento horizontal (clases de layout). **Test-first** (aserciones nuevas primero).
  - **Criterio de terminado**: menos huella visual, textos completos, comportamiento F3 probado y sin cambios; no aparece dentro del menú móvil.
  - **Commit sugerido**: `feat(frontend): LanguageSwitcher compacto (formato, sin cambios de comportamiento)`

---

## Fase 7 — Modo oscuro aplicado (US5, Priority: P2)

- [X] T1028 · `ThemeToggle` + integración en el clúster de la cabecera · `[frontend]` `[P5]`

  - **Archivos**: `frontend/src/features/publico/theme/ThemeToggle.tsx` (NUEVO), `frontend/src/features/publico/components/SiteHeader.tsx` (EDITADO: monta el toggle), `theme.test.tsx` (EDITADO **aditivamente**).
  - **Qué hace**: interruptor claro/oscuro **compacto** (objetivo 44 px, iconos decorativos 🌙/☀️) en el clúster de controles de la cabecera (junto al selector de idioma): `aria-pressed` con el estado y `aria-label` dinámico localizado (`theme.toDark`/`theme.toLight`, FR-023), cambio **inmediato sin recargar** vía `setTheme` (T1005), foco visible y sin perder el foco al cambiar de tema, sin ocupar espacio innecesario (misma queja que motivó compactar el idioma). En móvil ≤ 860 px convive con la hamburguesa a 320 px sin taparse (FR-025). **D7**: con elección guardada manda la manual; sin ella, sigue al sistema. *Cubre FR-023/FR-025/SC-012*.
  - **Pruebas incluidas (§III)**: `theme.test.tsx` (ampliado): el toggle alterna el tema efectivo y persiste `ss.theme`; nombre accesible dinámico y estado anunciado (`aria-pressed`); el cambio no mueve el foco; el control es operable con teclado. **Test-first**.
  - **Criterio de terminado**: los 3 caminos de US5 funcionan desde la cabecera; a 320 px caben marca (monograma) + idioma + tema + hamburguesa sin overflow.
  - **Commit sugerido**: `feat(frontend): ThemeToggle accesible en la cabecera del sitio publico`

---

## Fase 8 — E2E (Playwright, `frontend/e2e/`)

**Purpose**: flujos críticos de punta a punta y verificación de los requisitos medibles. Son tareas `[P6]` (un archivo cada una). Las suites existentes de F2/F3 **no se editan** (D12/SC-008): estas specs son **nuevas**.

- [X] T1029 · E2E `diseno-movimiento.spec.ts` (barra fija, anchos, teclado, selector) · `[frontend]` `[P6]`

  - **Archivos**: `frontend/e2e/diseno-movimiento.spec.ts` (NUEVO).
  - **Qué hace**: escritorio (≥ 861 px): barra fija visible en el 100 % del recorrido con **todas las anclas visibles** (sin menús ocultos); clic en ancla → título de la sección destino visible (no tapado: `scroll-padding-top`); 320/360/768/1024/1280 px: `document.documentElement.scrollWidth <= clientWidth` (0 desplazamiento horizontal, SC-004) y 0 enlaces que requieran scroll lateral (el menú móvil lo cubre T1030); recorrido con teclado y foco visible; **selector de idioma**: cambia al instante, persiste al navegar y entre visitas (contexto persistente), primera visita en español, textos completos accesibles; **cabecera a 320 px**: marca + idioma + tema + hamburguesa alcanzables y sin taparse (FR-025) y el enlace de marca conservando un texto con «Simiente» (nombre completo accesible, D20/C3).
  - **Pruebas incluidas (§III)**: la propia spec es la prueba (Playwright); **criterio de terminado**: `make e2e` la ejecuta en verde y cubre US1/US2/US4 (SC-004/SC-007).
  - **Commit sugerido**: `test(frontend): e2e de diseno, navegacion fija y selector compacto`

- [X] T1030 · E2E `navegacion-movil.spec.ts` (menú hamburguesa + mapa del sitio) · `[frontend]` `[P6]`

  - **Archivos**: `frontend/e2e/navegacion-movil.spec.ts` (NUEVO).
  - **Qué hace**: a 320 px: botón hamburguesa ≥ 44 px con nombre accesible y `aria-expanded`/`aria-controls`; abrir/cerrar con ratón **y** con teclado (Enter/Espacio); al abrir, el foco entra a la lista (primer enlace); **Escape**, control visible y toque fuera (scrim) cierran y **devuelven el foco al botón**; **sin trampa de foco** (D17/I1): `Tab` puede salir del panel con normalidad y **no** hay `role="dialog"`/`aria-modal`; **una sola lista de enlaces en el DOM** (D15/C4: cada sección resuelve a **una** única coincidencia dentro del `navigation` nombrado, también con el menú abierto); con el menú **cerrado**, los enlaces de la lista **no** están en el orden de tabulación (D16/I2: ocultos con `display:none`, no solo invisibles) y al abrir vuelven a ser enfocables; al activar una sección, el panel se cierra y el título del destino queda visible; **sin scroll de fondo** con el panel abierto (rueda y teclado); sin scroll horizontal; al ensanchar a ≥ 861 px el panel se cierra y reaparecen las anclas; cabecera con `data-scrolled` al bajar en ambos modos; y **mapa del sitio del pie**: lista las secciones publicadas (apilado a 320 px, columnas en escritorio), sus enlaces son alcanzables por teclado y llevan a la sección, con una sección en borrador no queda enlace (ni pie ni cabecera), el landmark del pie tiene nombre accesible propio (diferente del de la cabecera), **los grupos sin ítems no se renderizan y el mapa no se dibuja si no hay nada que listar** (D18/C1), los rótulos de grupo **no** tienen rol `heading` (D18) y **los nombres accesibles del pie son únicos** (prefijos «Correo:…/Teléfono:…/WhatsApp:…», planos en Síguenos; D19/C2). El **recuento del ciclo de tabulación** (< 30 pasos, el inventario está en T1015) se verifica aquí sin editar la prueba congelada.
  - **Pruebas incluidas (§III)**: la propia spec es la prueba; **criterio de terminado**: cubre US2 completo y los edge cases del menú (SC-003/SC-004/SC-009/SC-015).
  - **Commit sugerido**: `test(frontend): e2e del menu hamburguesa movil y el mapa del sitio`

- [X] T1031 · E2E `tema-oscuro.spec.ts` (3 caminos de US5, logo intacto) · `[frontend]` `[P6]`

  - **Archivos**: `frontend/e2e/tema-oscuro.spec.ts` (NUEVO).
  - **Qué hace**: (1) sistema en oscuro sin elección guardada → sitio en oscuro desde el **primer pintado** (sin flash: el `data-theme` inicial es correcto, anti-FOUC de T1002); (2) elección manual → cambio inmediato sin recarga y **persistente** tras recarga (y en visita posterior); (3) primera visita sin elección → sigue la preferencia del sistema; cambio en caliente del sistema **solo** sin elección manual; logotipo con sus colores originales en oscuro (sin `filter`/inversión); mismo contenido y mismas funciones en ambos temas (incluidos navegación, menú y mapa); control de tema operable por teclado con nombre accesible y estado.
  - **Pruebas incluidas (§III)**: la propia spec es la prueba; **criterio de terminado**: cubre US5 y los caminos de error (SC-012).
  - **Commit sugerido**: `test(frontend): e2e de tema claro/oscuro con persistencia`

- [X] T1032 · E2E `movimiento.spec.ts` (perceptibilidad, stagger, cabecera, sin bloqueo) · `[frontend]` `[P6]`

  - **Archivos**: `frontend/e2e/movimiento.spec.ts` (NUEVO).
  - **Qué hace**: en modo normal (sin `reducedMotion`): una sección fuera de pantalla nace con opacidad < 1 y desplazamiento y, al entrar en pantalla, termina en opacidad 1 / `transform: none` con `transition-duration` en rango (**≥ 400 ms y ≤ 800 ms** — el mismo techo de D5/D21, sin tolerancia de 900 ms —) y distancia **≥ 16 px** (suelo anti «casi imperceptible», SC-014/RG10-14) y ≤ 800 ms por elemento (techo); componentes hermanos con `transition-delay` escalonado (tope 240 ms); cabecera gana `data-scrolled` al bajar (superficie con transición, **sin cambio de altura**); clic en ancla → desplazamiento suave y título del destino visible; **la animación no bloquea**: un clic durante la animación funciona al instante y el contenido pendiente es clicable (FR-007/SC-006); al cambiar de idioma o de tema no se re-anima lo ya visto (0 rebotes, SC-014).
  - **Pruebas incluidas (§III)**: la propia spec es la prueba (aserciones de estilos computados, no capturas de píxeles, P10-10); **criterio de terminado**: el movimiento es claramente perceptible y acotado (SC-014).
  - **Commit sugerido**: `test(frontend): e2e de movimiento perceptible y sin bloqueos`

- [X] T1033 · E2E `movimiento-reducido.spec.ts` (0 revelados, contenido completo) · `[frontend]` `[P6]`

  - **Archivos**: `frontend/e2e/movimiento-reducido.spec.ts` (NUEVO).
  - **Qué hace**: `emulateMedia({ reducedMotion: 'reduce' })`: **0** revelados ni transiciones de entrada perceptibles —todos los bloques con opacidad 1 y `transform: none` desde el primer instante, sin `transition-delay`, `scroll-behavior: auto`, cabecera sin transición, blobs sin animar, menú móvil con apertura/cierre instantáneos— y **100 % del contenido y las funciones disponibles**, idénticos al modo con movimiento (SC-005); se compara con el modo normal de T1032 (ahí sí hay revelado, aquí ninguno).
  - **Pruebas incluidas (§III)**: la propia spec es la prueba; **criterio de terminado**: SC-005 verificado de forma determinista.
  - **Commit sugerido**: `test(frontend): e2e de movimiento reducido (SC-005)`

---

## Fase 9 — Cierre

- [X] T1034 · No-regresión F2/F3: suites **sin cambios** en verde + contrato congelado · `[frontend]`

  - **Archivos**: **ninguno de producción**; solo lectura/ejecución. (Si algo falla, se corrige el **código** de F10 o se **escala al humano**: D12/RG10-6.)
  - **Qué hace**: ejecuta `npm test -- --run` y `make e2e` completos y verifica que **todas** las suites de F2/F3 pasan **sin haberlas editado**: unitarias (`BrandLogo.test.tsx`, `homePage.test.tsx`, `i18n.test.tsx`, `layout.test.tsx`, `sections.test.ts`, `router.test.tsx`, `guards.test.tsx`, `providers.test.tsx`, las de panel/usuarios/roles/auditoría/auth/información) y e2e (`portada-publica`, `portada-accesibilidad` —incluido `tabCycle` < 30 pasos a 320 px, RG10-21—, `portada-panel`, `portada-imagenes`, `portada-patch-null`, `acceso`, `auditoria`, `status`). Verifica además el **contrato congelado** (R10-9): rutas, ids/anclas, `aria-label`/roles/nombres accesibles existentes, `ss.lang`, comportamiento bilingüe y de borrador/publicado. *Cubre FR-020/FR-021/SC-008*.
  - **Pruebas incluidas (§III)**: la ejecución completa de ambas suites **es** la prueba (§III + SC-008); **criterio de terminado**: 100 % en verde con `git diff` de las suites de F2/F3 **vacío**.
  - **Commit sugerido**: `test(frontend): verificacion de no-regresion de F2/F3 (SC-008)`

- [X] T1035 · Docs de cimientos para F4–F9 (`components/publico/README.md` + `quickstart.md`) · `[frontend]` `[P7]`

  - **Archivos**: `frontend/src/components/publico/README.md` (NUEVO), `specs/010-sistema-diseno-movimiento/quickstart.md` (EDITADO: resultados de la checklist), `specs/010-sistema-diseno-movimiento/ux.md` (solo si QA calibra valores de tokens: D9/RG10-22).
  - **Qué hace**: documenta los cimientos heredables para que F4–F9 nazcan con el sistema **sin rediseño** (FR-018/SC-011): catálogo de tokens (tema claro/oscuro con su tabla de contraste, espaciado de sección, movimiento con suelo/techo), primitivas (`PublicButton`/`SectionHeading`/`Card`), patrones (`Reveal`/`RevealGroup`, `MobileMenu`/`SiteHeader`, `FooterSiteMap`/`useSiteLinks`, `ThemeToggle`), reglas de composición (toda sección usa `--space-section*`; un solo estilo por función; el logotipo nunca se anima ni se recolorea) y cómo añadir una sección nueva (nueva entrada en `sections.ts` + ruta → la navegación y el mapa del sitio se actualizan solos, sin enlaces rotos). Deja escrito el **criterio de ubicación de componentes** (N3): `frontend/src/components/publico/` es la capa de **primitivas del sistema** (reutilizables por cualquier feature) y `frontend/src/features/publico/components/` la capa de **secciones/composición** de la portada; un componente nuevo se coloca según si es primitiva de sistema o pieza de una sección concreta (no por «no alterar el inventario compartido», que aplica a los dos). Registra la calibración final de espaciado/movimiento decidida en QA. *Cubre FR-018/US6/SC-011*.
  - **Pruebas incluidas (§III)**: — (documentación; **excepción §III declarada**, N4). **Cómo se verifica**: una persona ajena a F10 puede construir una sección de ejemplo F4 siguiendo el README sin inventar estilos base (revisión en QA); el README referencia solo mecanismos que existen en el código.
  - **Criterio de terminado**: los cimientos están entregados y documentados (SC-011) y `quickstart.md` refleja la calibración real.
  - **Commit sugerido**: `docs(frontend): cimientos del sistema de diseño para F4-F9`

- [X] T1036 · Verificación de cierre (`make ci`, `make e2e`, checklist `quickstart.md`, mapa SC) · `[infra]`

  - **Archivos**: **ninguno** (verificación); acta de resultados en `specs/010-sistema-diseno-movimiento/quickstart.md` (EDITADO con los resultados).
  - **Qué hace**: ejecuta `make ci` y `make e2e` completos (unitarias + typecheck + lint + e2e, incluida la no-regresión) y recorre la checklist manual de `quickstart.md`: «moderno» y fidelidad al Manual (SC-002), **movimiento marcado/elegante aprobado como perceptible y sin marear** (SC-014), contraste AA por tema y estado incluidos hover/foco/activo (SC-012/SC-009), selector con «menos espacio» y reconocible (SC-007), logotipo sin recolorear en ambos temas y en estados interactivos (FR-002/FR-024), panel **sin cambios** (revisión visual + suites), zoom 200 % y 320 px, rendimiento en el dispositivo de referencia (SC-006/RG10-9) y 0 textos sin traducir (SC-010). Cruza el **mapa de cobertura** de abajo (FR-001…FR-027, SC-001…SC-015) y deja listo el acta para la **aprobación humana** (SC-002/SC-013/SC-014 la requieren: no se auto-aprueba).
  - **Pruebas incluidas (§III)**: la ejecución completa + checklist **son** la verificación (sin pruebas propias: **excepción §III declarada**, N4); **criterio de terminado**: `make ci` y `make e2e` en verde y checklist completa con resultados registrados, pendiente solo de la revisión del cliente.
  - **Commit sugerido**: `chore: verificacion de cierre de F10 (checklist y mapa SC)`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Fase 1 (Cimientos)**: sin dependencias previas — **bloquea todas** las demás fases.
- **Fase 2 (Primitivas)**: depende de Fase 1 (tokens de T1004).
- **Fase 3 (Navegación/pie)**: depende de Fase 1 (tokens + `useSiteLinks` + claves i18n); T1016 depende además de T1010 (`Card`).
- **Fase 4 (Movimiento)**: depende de Fase 1 (tokens de movimiento); T1020 depende además de T1011 (`SiteHeader`).
- **Fase 5 (Propuesta B)**: depende de Fases 1–2; T1026 depende además de Fase 4 y de T1021…T1025.
- **Fase 6 (Selector)** y **Fase 7 (Tema)**: dependen de Fase 1 y de T1011 (clúster de la cabecera); son independientes entre sí (P5).
- **Fase 8 (E2E)**: depende de las fases que cubre cada spec (ver tabla).
- **Fase 9 (Cierre)**: depende de todo lo anterior.

### Orden de implementación sugerido (incremental y testeable)

1. **T1001 → T1003 → T1004** (tokens en `index.css`) **en paralelo con T1002** (anti-FOUC): el sistema de capas queda listo y el panel sigue intacto.
2. **T1005 · T1006 · T1007 en paralelo** (tema, i18n, enlaces): tres módulos independientes con sus unitarias en verde.
3. **T1008 · T1009 · T1010 en paralelo** (primitivas).
4. **T1011 → T1012 → T1013 → T1014** (cabecera → anclas de escritorio → menú móvil → integración): al terminar, US2 es navegable y se **recuenta** el ciclo de tabulación congelado (RG10-21).
5. **T1015 → T1016** (mapa del sitio → pie).
6. **T1017 → T1018 → T1019**, y **T1020** en cuanto exista `SiteHeader`: el sistema de movimiento queda calibrable.
7. **T1021** y **T1022 · T1023 · T1024 · T1025 en paralelo**, y **T1026** al final de la fase: **MVP** — portada completa con la propuesta B. *Punto de validación con el cliente (SC-002/SC-014).*
8. **T1027 · T1028 en paralelo** (selector compacto y control de tema).
9. **T1029 · T1030 · T1031 · T1032 · T1033 en paralelo** (e2e).
10. **T1034 → T1035 → T1036** (no-regresión, docs, verificación de cierre → aprobación humana).

### Within each phase

- Las pruebas **se escriben primero** y deben fallar antes de la implementación (§III).
- Dentro de una tarea: componente/hook → su test → integración mínima → commit.
- Cada tarea termina con `npm test -- --run`, `npm run typecheck` y `npm run lint` en verde (y `make e2e` cuando la tarea toca navegación/pie, para las suites congeladas: RG10-18).

### Parallel Opportunities

- Grupos **P1…P7** de la tabla «Grupos de paralelismo» (20 tareas `[P]`).
- Las 5 specs e2e (P6) pueden escribirse y ejecutarse en paralelo entre sí.

---

## Estrategia MVP (entrega incremental)

1. **Cimientos** (Fase 1–2): tokens + tema + primitivas — aún sin cambio visible grande, pero toda base verificada.
2. **MVP = Fases 1–5** (T1001…T1026): portada completa con la propuesta B, ambos temas, navegación por dispositivo y mapa del sitio con movimiento — **validable con el cliente** (SC-002).
3. **Fases 6–7**: selector compacto y control de tema (US4/US5).
4. **Fases 8–9**: e2e completos, no-regresión, docs y verificación de cierre → revisión humana.

Cada fase deja `make ci` en verde y no rompe ninguna historia anterior.

---

## Cobertura de requisitos (FR-001…FR-027 / SC-001…SC-015)

| Requisito | Tareas | Se verifica con |
|---|---|---|
| FR-001 (sistema visual 2026 sobre el Manual) | T1001, T1003, T1008–T1010, T1021–T1025 | Revisión visual (SC-001/SC-002), checklist `quickstart.md` |
| FR-002 (reglas del logotipo, proporción 70/20/10, iconografía nueva ok) | T1001, T1021, T1024 | `BrandLogo.test.tsx` intacta, checklist (SC-012) |
| FR-003 (componentes base, un solo estilo por función) | T1008–T1010, T1021–T1024, T1035 | Unitarias de primitivas, revisión SC-001/SC-011 |
| FR-004 (legibilidad para todas las edades) | T1008, T1021–T1024, T1027 | Revisión QA, e2e 320 px/zoom 200 % (T1029) |
| FR-005 (movimiento marcado con revelado al scroll) | T1004, T1018–T1020, T1025, T1026 | `motion.test.tsx`, e2e `movimiento.spec.ts` (T1032) |
| FR-006 (`prefers-reduced-motion` incondicional) | T1004, T1017–T1019, T1013 | `motion.test.tsx`, e2e `movimiento-reducido.spec.ts` (T1033, SC-005) |
| FR-007 (sin degradar carga ni fluidez) | T1018, T1025, T1026 | e2e T1032 (contenido operable durante la animación), checklist SC-006 |
| FR-008 (retroalimentación clara, sin repeticiones molestas) | T1008, T1021, T1023, T1026 | Unitarias + e2e T1032 (cambio de idioma/tema no re-anima) |
| FR-009 (nada depende de ver una animación) | T1018, T1024, T1026 | `motion.test.tsx` (fallbacks, foco), e2e T1032/T1033 |
| FR-010 (navegación por dispositivo: barra fija / hamburguesa) | T1011–T1013 | e2e T1029 (escritorio) y T1030 (móvil) |
| FR-011 (fin del «scroll incómodo», 0 scroll horizontal) | T1012, T1013, T1014 | e2e T1029 (`scrollWidth <= clientWidth` desde 320 px) |
| FR-012 (escala a F4–F9) | T1007, T1012, T1015, T1035 | Unitarias con entradas F4–F9 simuladas, docs |
| FR-013 (accesibilidad de la navegación + foco del menú) | T1012, T1013, T1015 | `MobileMenu.test.tsx`, e2e T1030, `portada-accesibilidad.spec.ts` sin cambios |
| FR-014 (navegación siempre alcanzable, sin tapar destinos) | T1011, T1014, T1020 | e2e T1029/T1030 (título visible tras el salto, `data-scrolled`) |
| FR-015/FR-016/FR-017 (selector compacto, comportamiento F3, sin desplazamientos) | T1027 | `i18n.test.tsx` (existentes sin cambios + nuevos), e2e T1029 (SC-007) |
| FR-018 (aplicación a toda la portada + cimientos F4–F9) | T1021–T1026, T1035 | Revisión SC-001/SC-011 |
| FR-019 (textos nuevos es/en) | T1006 | `i18n.test.tsx` (SC-010) |
| FR-020/FR-021 (no-regresión funcional, sin datos nuevos) | T1034 (y la frontera F10B de todas) | Suites F2/F3 sin cambios (SC-008), `data-model.md` |
| FR-022 (accesibilidad en ambos temas) | T1013, T1015, T1028, T1029, T1030 | e2e de teclado, checklist SC-009 |
| FR-023 (modo oscuro + control manual + persistencia) | T1001, T1002, T1005, T1028 | `theme.test.tsx`, e2e `tema-oscuro.spec.ts` (T1031, SC-012) |
| FR-024 (contraste AA en ambos temas, logo sin recolorear) | T1001, T1010, T1021, T1036 | Checklist por tema y estado (SC-012) |
| FR-025 (cabecera completa a 320 px) | T1011, T1013, T1028 | e2e T1029 (320 px: marca + idioma + tema + menú) |
| FR-026 (propuesta B aprobada antes de código) | Proceso (ya cumplido) + T1021–T1026, T1036 | Acta de aceptación humana (SC-013) |
| FR-027 (mapa del sitio sin enlaces rotos, preparado F4–F9) | T1007, T1015, T1016 | `FooterSiteMap.test.tsx`, e2e T1030 (SC-015) |
| SC-001/SC-011 (coherencia y cimientos) | Fase 2 + Fase 5 + T1035 | Revisión visual + docs |
| SC-002/SC-013/SC-014 («moderno», propuestas, movimiento perceptible) | T1026, T1032, T1036 | **Aprobación humana** + checklist (no auto-aprobable) |
| SC-003/SC-004 (navegación hallable, 0 scrolls) | T1012–T1015, T1029, T1030 | e2e + pruebas con personas (aceptación humana) |
| SC-005/SC-006 (reduced-motion, sin bloqueos) | T1004, T1017–T1019, T1032, T1033 | e2e deterministas |
| SC-007 (selector con menos espacio, reconocible) | T1027, T1029, T1036 | e2e + checklist humana |
| SC-008 (0 regresiones F2/F3) | T1034 | Suites sin cambios en verde |
| SC-009 (accesibilidad, ambos temas) | T1013, T1015, T1028–T1030, T1036 | e2e teclado + checklist lector de pantalla |
| SC-010 (0 textos sin traducir) | T1006, T1036 | `i18n.test.tsx` (tipado) |
| SC-012 (tema: sistema/manual/persistencia, contraste, logo) | T1005, T1028, T1031, T1036 | e2e tema + checklist de contraste |
| SC-015 (mapa del sitio sin rotos) | T1015, T1030 | Unitarias + e2e |

---

## Notes

- `[P]` = archivos distintos, sin dependencias previas entre ellas (ver «Grupos de paralelismo»).
- Cada tarea se cierra con **commit pequeño** (Conventional Commits; los mensajes sugeridos son orientativos) y con sus pruebas en verde.
- **F10B (panel) queda fuera** de este documento en su totalidad.
- Cualquier prueba congelada de F2/F3 que «falle» por el rediseño se corrige en el **código** de F10 o se **escala al humano**; jamás se edita la prueba para hacerla pasar (RG10-6/RG10-18, SC-008).
