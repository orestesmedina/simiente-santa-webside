# Sistema de diseño público (cimientos para F4–F9)

> F10 · «Marea Turquesa» — catálogo de los **cimientos heredables** del sitio público. Una sección
> nueva (F4–F9) se construye con esto **sin rediseñar**: tokens, primitivas, patrones de movimiento y
> de navegación, tema y reglas de composición. Fuente de verdad del diseño: `ux.md`; de la mecánica:
> `plan.md`/`tasks.md` de `specs/010-sistema-diseno-movimiento/`.

---

## 1. Dónde vive cada componente (criterio de ubicación, N3)

| Capa                        | Ruta                               | Qué contiene                                                                                                                                                                                                              |
| --------------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Primitivas del sistema**  | `frontend/src/components/publico/` | Piezas **reutilizables por cualquier feature** del sitio público: `PublicButton`, `SectionHeading`, `Card`. Sin datos de negocio ni i18n interna.                                                                         |
| **Secciones y composición** | `frontend/src/features/publico/`   | Piezas **de la portada concreta**: `components/` (SiteHeader, PublicNav, MobileMenu, IdentityHero, WhoWeAreSection, …), `motion/`, `theme/`, `i18n/`, `hooks/`, `pages/`.                                                 |
| **Inventario del panel**    | `frontend/src/components/` (raíz)  | Componentes del **panel de administración** (Button, Table, Dialog, Field…). **No se tocan** desde el sitio público (F10B los reconciliará): el sistema público es una capa aparte precisamente para no alterar el panel. |

**Regla**: un componente nuevo va a `components/publico/` si es una **primitiva de sistema**
(reutilizable, sin depender de la portada) y a `features/publico/…` si es una **pieza de una sección**
concreta. El criterio es _qué es_, no _dónde queda más cómodo_.

---

## 2. Tokens

Todos viven en `frontend/src/index.css`. La capa semántica es **aditiva**: no toca los tokens de marca
de F3 (`--color-navy`, `--color-teal`, `--color-cream`…), que siguen siendo los del panel.

### 2.1 Tema (claro/oscuro)

El tema se resuelve por el atributo `data-theme` de `<html>`: lo fija el **script anti-FOUC** de
`index.html` antes del primer pintado y lo mantiene `ThemeProvider` en caliente. `dark:` de Tailwind 4
está ligado al atributo (`@custom-variant dark`), no a `prefers-color-scheme`.

| Variable            | Tema claro (`light`)             | Tema oscuro (`dark`)          | Utilidad Tailwind          |
| ------------------- | -------------------------------- | ----------------------------- | -------------------------- |
| `--surface`         | `#1a2b4a` (navy)                 | `#0b1424`                     | `bg-surface`               |
| `--surface-raised`  | `rgb(255 255 255 / .08)`         | `rgb(16 27 48 / .7)`          | `bg-surface-raised`        |
| `--text`            | `#eef3f8`                        | `#eef3f8`                     | `text-ink`                 |
| `--text-muted`      | `#aebdd6`                        | `#aebdd6`                     | `text-muted`               |
| `--border`          | `rgb(238 243 248 / .16)`         | `rgb(238 243 248 / .14)`      | `border-line`              |
| `--accent`          | `#00c9a7`                        | `#00c9a7`                     | `text-accent`, `bg-accent` |
| `--accent-strong`   | `#0e7a66`                        | `#00c9a7`                     | `text-accent-strong`       |
| `--accent-hover`    | `#0e7a66`                        | `#00b598`                     | `bg-accent-hover`          |
| `--on-accent-hover` | `#ffffff`                        | `#1a2b4a`                     | `text-on-accent-hover`     |
| `--focus`           | `#00c9a7`                        | `#00c9a7`                     | `outline-focus`            |
| `--shadow`          | `0 18px 44px rgb(5 10 20 / .35)` | `0 18px 44px rgb(0 0 0 / .5)` | `shadow-surface`           |
| `--logo-surface`    | `#f5f2ec`                        | `#f5f2ec`                     | `bg-logo-surface`          |

**Reglas del color (FR-002/FR-024):**

- **`--accent-strong` es para texto turquesa sobre superficies CLARAS.** Sobre navy/oscuro el acento
  directo es `--accent` (≈ 6,7:1 en claro, ≈ 8,8:1 en oscuro). No usar `text-accent-strong` sobre
  `bg-surface`: queda ≈ 2,7:1 (falla AA).
- **El logotipo nunca se recolorea, deforma, gira ni anima**; en ambos temas se apoya en su pastilla
  crema (`bg-logo-surface`). Ver `BrandLogo.test.tsx` (congelada).
- **Coral solo como detalle gráfico**, nunca como texto.

### 2.2 Pares de contraste AA por tema (referencia para QA)

| Par                                                                  | Claro    | Oscuro   | AA texto normal (≥ 4,5:1) |
| -------------------------------------------------------------------- | -------- | -------- | ------------------------- |
| `--text` sobre `--surface`                                           | ≈ 12,6:1 | ≈ 16,6:1 | ✅                        |
| `--text-muted` sobre `--surface`                                     | ≈ 7,4:1  | ≈ 9,7:1  | ✅                        |
| `--accent` sobre `--surface`                                         | ≈ 6,7:1  | ≈ 8,8:1  | ✅                        |
| `bg-accent` + `text-navy` (PublicButton primary)                     | ≈ 6,7:1  | ≈ 8,8:1  | ✅                        |
| `bg-accent-hover` + `text-on-accent-hover` (hover del primary)       | ≈ 5,3:1  | ≈ 5,4:1  | ✅                        |
| `bg-accent` + `text-navy` (opción **activa** del selector de idioma) | ≈ 6,7:1  | ≈ 6,7:1  | ✅                        |
| `bg-leaf` (#217638) + `text-white` (CTA WhatsApp)                    | ≈ 5,7:1  | ≈ 5,7:1  | ✅                        |
| `bg-coral` (#ff6b3d) + `text-navy` (accent)                          | ≈ 5,0:1  | ≈ 5,0:1  | ✅                        |
| `--accent-strong` sobre `--surface`                                  | ≈ 2,7:1  | n/a      | ❌ **no usar sobre navy** |

> La auditoría automática de contraste del e2e congelado (`portada-accesibilidad.spec.ts`) recorre todo
> el texto de `main` y `footer`. **Ojo**: colores con alfa o `color-mix()` (p. ej. aplicar una opacidad
> al color del texto) hacen que esa auditoría no pueda resolver el color y los marque como negro (falso
> positivo de contraste). Para texto del sitio público, usar **colores sólidos** (`text-ink`,
> `text-muted`, `text-accent`).

### 2.3 Espaciado vertical de sección

Valores **por breakpoint** (un solo punto de calibración; la calibración final la cierra QA, RG10-22).

| Token                   | < 480 px                     | 480–767  | 768–1279 | ≥ 1280   | Utilidad           |
| ----------------------- | ---------------------------- | -------- | -------- | -------- | ------------------ |
| `--space-section`       | `3.25rem`                    | `4rem`   | `5.5rem` | `7rem`   | `py-section`       |
| `--space-section-tight` | `2rem`                       | `2.5rem` | `3.5rem` | `4.5rem` | `py-section-tight` |
| `--space-block`         | `1.5rem`                     | `2rem`   | `2.5rem` | `3rem`   | `gap-block`        |
| `--space-fluid`         | `clamp(.8rem, 2vw, 1.25rem)` |          |          |          | `gap-fluid`        |
| `--pad-panel`           | `clamp(2rem, 5vw, 3.5rem)`   |          |          |          | `p-panel`          |
| `--pad-text`            | `clamp(1rem, 2.4vw, 1.6rem)` |          |          |          | `p-text`           |

### 2.4 Movimiento (única fuente numérica)

| Token                      | Valor                           | Uso                                 |
| -------------------------- | ------------------------------- | ----------------------------------- |
| `--motion-duration-micro`  | `180ms`                         | hover / foco / activo               |
| `--motion-duration-header` | `220ms`                         | cabecera al bajar (`data-scrolled`) |
| `--motion-duration-reveal` | `600ms`                         | revelado de secciones/bloques       |
| `--motion-duration-hero`   | `700ms`                         | entrada del hero (solo 1.ª carga)   |
| `--motion-distance`        | `28px` (`20px` a ≤ 480 px)      | desplazamiento de entrada           |
| `--motion-stagger-step`    | `80ms` (tope 3 saltos = 240 ms) | escalonado entre hermanos           |
| `--motion-ease-out`        | `cubic-bezier(.22, 1, .36, 1)`  | entradas y revelados                |
| `--motion-ease-standard`   | `cubic-bezier(.4, 0, .2, 1)`    | microinteracciones y cabecera       |

**Reglas duras (SC-014):** todo revelado/transición de entrada **≥ 400 ms y ≥ 16 px** (suelo de
perceptibilidad) y **≤ 800 ms** por elemento y por secuencia (techo). Solo `opacity`/`transform`/colores
(nunca propiedades de layout): 0 desplazamientos del maquetado. **`prefers-reduced-motion` es
incondicional** y se aplica en **doble capa**: bloque CSS global (`index.css`) + hook
`usePrefersReducedMotion` en vivo.

---

## 3. Primitivas (`components/publico/`)

### `PublicButton` — `PublicButton.tsx`

```tsx
<PublicButton variant="primary">Ver horario</PublicButton>
<PublicButton href="#horario" variant="whatsapp" external>Escríbenos</PublicButton>
```

| Prop                                       | Valores                                                                                        | Notas                                              |
| ------------------------------------------ | ---------------------------------------------------------------------------------------------- | -------------------------------------------------- |
| `variant`                                  | `primary` (turquesa) · `accent` (coral) · `secondary` (contorno) · `whatsapp` (verde de marca) | por defecto `primary`                              |
| `href`                                     | `string`                                                                                       | si se indica, renderiza `<a>` en vez de `<button>` |
| `external`                                 | `boolean`                                                                                      | añade `target="_blank" rel="noopener noreferrer"`  |
| `onClick`, `type`, `className`, `children` |                                                                                                |                                                    |

Objetivo táctil ≥ 44 px, foco visible (`outline-focus`) y microinteracción de
`--motion-duration-micro` (solo `transform`/colores). **Es el único estilo de botón de acción del
sitio público** (SC-011).

### `SectionHeading` — `SectionHeading.tsx`

```tsx
<SectionHeading id="horario-title" kicker="02">
  Horario de servicios
</SectionHeading>
```

`kicker` (número/kicker 01–05), `level` (`2` por defecto; `3` para subtítulos), `id` (para el
`aria-labelledby` de la sección) y `className`. Incluye el **acento turquesa** decorativo
(`aria-hidden`, `scaleX` con el token de revelado; se despliega con el bloque `Reveal`). Todas las
secciones usan este componente para su título: **0 jerarquías distintas**.

### `Card` — `Card.tsx`

```tsx
<Card variant="glass" className="flex flex-col gap-block">…</Card>
<Card variant="glass" padded={false} className="overflow-hidden">imagen</Card>
```

`variant` (`solid` | `glass`), `interactive` (añade elevación en hover **solo** si es clicable),
`padded` (`false` para un marco de imagen a sangre), `as` (`div` | `section` | `article` | `li`).
Es la **única superficie de contenido** del sitio público (radio, borde `--border`, sombra
`--shadow`, fondo `--surface-raised`, padding `--pad-panel`).

---

## 4. Patrones

### 4.1 Movimiento — `features/publico/motion/`

| Pieza                       | API                                                                        | Para qué                                                                                   |
| --------------------------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| `Reveal`                    | `as`, `from: 'up'\|'left'\|'right'`, `delay?`, `revealIndex?`, `className` | Envuelve un bloque para que se revele al entrar en pantalla (600 ms, 28 px, una sola vez). |
| `RevealGroup`               | `className`                                                                | Reparte `--reveal-index` (0–3, tope) entre sus hijos `Reveal` → escalonado de 80 ms.       |
| `usePrefersReducedMotion()` | `boolean`                                                                  | Preferencia en vivo (segunda capa de `prefers-reduced-motion`).                            |
| `useScrolled()`             | `{ sentinelRef, scrolled }`                                                | Sentinela observada (sin listener de `scroll`) → `data-scrolled` de la cabecera.           |

**Progresivo por construcción**: el contenido **nace visible**; el estado pendiente solo lo pinta
`html.js`, y con `prefers-reduced-motion`, sin JS o si el observer no dispara (salvavidas de 1 s) el
contenido se ve completo. El **foco revela al instante** un bloque pendiente (el foco nunca queda
sobre algo oculto).

```tsx
<RevealGroup>
  <Reveal key="horario" from="right">
    <ScheduleSection items={schedule} />
  </Reveal>
  <Reveal key="contacto" from="right">
    <ContactSection contact={contact} />
  </Reveal>
</RevealGroup>
```

> En móvil (≤ 860 px) las variantes laterales se convierten en verticales por CSS. El grupo recorta el
> desborde horizontal del desplazamiento (`overflow-x: clip`) para no ganar scroll lateral (SC-004).

### 4.2 Navegación y pie

| Pieza                            | Rol                                                                                                                                                                                                                                                                                          |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `useSiteLinks()`                 | **Fuente única** de los enlaces navegables: secciones publicadas (`presentSections`), en el orden de la página y con las claves i18n de la cabecera. Alimenta cabecera **y** pie.                                                                                                            |
| `SiteHeader`                     | Cabecera fija (`sticky`): marca (enlace a `/`, con el nombre completo siempre accesible), `<nav>` y clúster de controles (idioma + tema). Publica `--header-offset` (altura real) para el `scroll-padding-top`.                                                                              |
| `PublicNav`                      | La **única lista** de enlaces del `<nav>`. En escritorio (≥ 861 px) filas `flex-wrap`; en móvil se colapsa (`display:none`, fuera del orden de tabulación) salvo con el panel abierto.                                                                                                       |
| `useMobileMenu()` + `MobileMenu` | Menú hamburguesa ≤ 860 px (**patrón _disclosure_ de APG**, sin diálogo ni trampa de foco): `aria-expanded`/`aria-controls`, foco al primer enlace al abrir, `Escape`/scrim/control cierran devolviendo el foco, scroll de fondo bloqueado (`html[data-nav-open='true']`).                    |
| `FooterSiteMap`                  | Mapa del sitio (`Explorar` / `Contacto` / `Síguenos`) dentro del `<footer>`: grupos **condicionales** (sin ítems no se renderiza; sin grupos no se dibuja), rótulos **no-heading** y **nombres accesibles únicos** (Contacto con prefijo «Contacto: …», dirección como texto, redes planas). |
| `PublicFooter`                   | Marca (texto, no enlace) + `FooterSiteMap` + franja con el lema y el selector de idioma.                                                                                                                                                                                                     |

**Contrato congelado (R10-9)**: rutas, anclas (`quienes-somos`, `horario`, `whatsapp`, `contacto`,
`redes`), `aria-label`/roles/nombres accesibles existentes, `localStorage['ss.lang']` y el
comportamiento bilingüe y de borrador/publicado **no cambian**. Las pruebas de F2/F3 son la red de
no-regresión: no se editan.

### 4.3 Tema — `features/publico/theme/`

| Pieza           | API                                                                                                       |
| --------------- | --------------------------------------------------------------------------------------------------------- |
| `ThemeProvider` | Envuelve el sitio público (`PublicLayout`). Aplica `data-theme`/`color-scheme` en `<html>`.               |
| `useTheme()`    | `{ theme, preference, setTheme }` (`preference: 'light'\|'dark'\|null`; `null` = sistema).                |
| `ThemeToggle`   | Interruptor 44 px con `aria-pressed` y nombre accesible dinámico; persiste en `localStorage['ss.theme']`. |

Sin elección guardada manda `prefers-color-scheme` (y sus cambios en caliente); con elección manual,
manda la manual. **El panel no se envuelve**: el atributo es global pero el efecto es solo del sitio
público (sus componentes usan los tokens semánticos).

### 4.4 Decoración — `GradientBlobs`

Capa de atmósfera (`aria-hidden`, `pointer-events: none`, `fixed`, recortada por `overflow-hidden`).
Solo `transform`, estática con `prefers-reduced-motion`. Nunca detrás de texto esencial.

---

## 5. Reglas de composición (obligatorias)

1. **Toda sección** usa `py-section` (o `py-section-tight` si es densa/encadenada) y `gap-block` entre
   encabezado y contenido. **Ninguna define su propio padding vertical.**
2. **Un solo estilo por función**: botones → `PublicButton`; títulos → `SectionHeading`; superficies →
   `Card`. No se duplican estilos «parecidos».
3. **Colores por tokens semánticos** (`bg-surface`, `text-ink`, `text-muted`, `border-line`,
   `text-accent`), nunca colores de F3 ni valores sueltos. Para texto, **colores sólidos**.
4. **Movimiento** solo con los tokens y solo `opacity`/`transform`/colores; el contenido nunca depende
   de ver una animación (FR-009) y el foco nunca queda sobre algo oculto.
5. **El logotipo no se recolorea, deforma, gira ni anima** (en ningún estado ni tema).
6. **Accesibilidad**: objetivo táctil ≥ 44 px, foco visible, contraste AA en ambos temas, jerarquía de
   encabezados correcta (un `h1` en el hero; `h2` por sección), listas y landmarks reales.
7. **El panel no se toca** (FR-020): nada de `components/` (raíz), `app/layout.tsx` ni las features del
   panel.

---

## 6. Cómo añadir una sección nueva (receta F4–F9)

Ejemplo: «Eventos» (F4).

1. **Modelo** — en `features/publico/sections.ts`: añadir el `id`, el `anchor` y la `messageKey` a
   `PUBLIC_SECTIONS`, y la regla de presencia en `isSectionPresent` (qué hace que la sección esté
   publicada).
2. **Textos** — en `features/publico/i18n/messages.ts` (+ `es.ts`/`en.ts`) añadir la clave del título.
   El tipado `Record<PublicMessageKey, string>` impide dejarla sin traducir.
3. **Componente** — crear `features/publico/components/EventsSection.tsx` usando `Card` +
   `SectionHeading` + `py-section`, con los tokens semánticos. Si necesitas una pieza reutilizable,
   va a `components/publico/`.
4. **Composición** — en `pages/HomePage.tsx`, renderizarla dentro del `RevealGroup` y envuelta en
   `<Reveal key="<anchor>" from="left|right">`.
5. **Nada más**: la **cabecera** (`PublicNav`/`MobileMenu`) y el **mapa del sitio** (`FooterSiteMap`)
   se alimentan de `useSiteLinks()`, así que aparecen solos en ambos y **sin enlaces rotos** (una
   sección sin publicados no genera enlace). No hay que tocar la navegación ni el pie.
6. **Pruebas** — unitarias de la sección (contrato, tokens, sin padding propio) y, si aporta flujo,
   e2e. Reutilizar los helpers de `frontend/e2e/` (`publicSeed.ts`, `publicUi.ts`).

---

## 7. Pruebas y verificación

- Unitarias: `cd frontend && npm test -- --run` · tipos y estilo: `npm run typecheck && npm run lint`.
- E2E (stack real): `make e2e` (requiere `make up`). En este WSL, Chromium no arranca por librerías de
  sistema; se usa la imagen oficial de Playwright contra el stack. Por el **rate-limit** de login
  (20/min por IP, M-4 de F2), conviene ejecutar en dos tandas (congeladas / nuevas).
- Al añadir una sección, comprobar la checklist manual de
  `specs/010-sistema-diseno-movimiento/quickstart.md` (contraste AA por tema y estado, 320 px, zoom
  200 %, foco visible, movimiento perceptible y su ausencia con `prefers-reduced-motion`).

## 8. Estado de calibración

Los valores de espaciado y movimiento son los **implementados** y están centralizados en
`index.css`. La **calibración final** («moderno y con aire, no vacío ni con scroll eterno»; movimiento
perceptible y no estridente) la cierra QA sobre la propuesta B (RG10-22/SC-002/SC-014): si algo se
ajusta, se cambia **el token**, nunca el componente.
