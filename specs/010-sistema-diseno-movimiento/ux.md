# UX: F10 — Sistema de diseño y movimiento — **Diseño final: Propuesta B «Marea Turquesa»**

> **Estado (ronda 3, 2026-10-10, cierre de fase de diseño):** el cliente **eligió la Propuesta B «Marea Turquesa» y aprobó el plan, con tres cambios** que este documento ya integra: (1) **más aire** — nueva escala de espaciado vertical con tokens de sección; (2) **mapa del sitio en el pie** — todas las opciones existentes del sitio público, en columnas/apilado, accesible y sin inventar enlaces; (3) **menú hamburguesa en móvil** (escritorio conserva la barra fija sin hamburguesa). Este documento **reemplaza a las 3 opciones** de la ronda 2 como **documento de diseño final** listo para `dev-frontend`. La maqueta de referencia vigente es `propuestas/propuesta-b.html` (HTML+CSS+JS autocontenido, con movimiento real, tema claro/oscuro, selector de idioma demo, hamburguesa y mapa del sitio).
>
> Fuentes: `spec.md` de F10 (US1–US6; FR-005/006/007/009/026), tokens de F3 (`specs/003-portada-info-general/ux.md`), Manual de Identidad (`resources/MANUAL DE MARCA.pdf` manda), referencias de inspiración del cliente (apple.com, microsoft.com — **inspiración, nunca copia**). Sin panel (F10B fuera). No toca `estado.md` ni el roadmap; no implementa la app.
>
> ⚠️ **Punto a confirmar para el analista:** `spec.md` FR-012 y US2 esc. 5 aún dicen «sin menús hamburguesa» (decisión ronda 1). El cliente en ronda 3 pidió explícitamente hamburguesa **en móvil**; el diseño final la aplica como **excepción móvil** y mantiene FR-014 (barra fija sin hamburguesa, todos los enlaces visibles) **en escritorio**. Conviene actualizar la redacción de FR-012/US2 antes de implementar.

---

## 0. Sistema compartido (base de F3, blanda, no reinventada)

### 0.1 Paleta y logotipo
- **70/20/10**: navy `#1a2b4a` domina (el sitio «vive» sobre navy), crema/blanco `#F5F2EC` 20 %, turquesa `#00c9a7` + coral `#ff6b3d` + verde `#217638` 10 %.
- **Regla de turquesa como texto**: sobre fondo claro usa `#0e7a66`; sobre navy/oscuro el turquesa puro es acento directo (≈ 8:1). Coral solo como detalle gráfico (nunca texto). Verde vive casi solo en el CTA de WhatsApp.
- **Logotipo**: intacto SIEMPRE (sin deformar, recolorear, girar ni efectos, incluido hover/foco). En cualquier tema se apoya sobre su **pastilla crema**. En las maquetas, monograma placeholder «SS».

### 0.2 Tipografías
- **Bebas Neue**: titulares de sección y momentos de impacto (hora del horario, números 01–05).
- **Poppins**: cuerpo de texto 18 px, botones, navegación.
- **Playfair Display itálica**: lema y frases emocionales.
- Producción: self-hosted (`@fontsource`); maqueta: fallbacks de sistema.

### 0.3 Escala de espaciado (nueva — cambio 1 del cliente)
El sitio debe sentirse **amplio y ordenado, con respiración entre bloques**, sin perder la energía de la marea. Tokens (definidos en `frontend/src/index.css` vía `@theme` + variables de tema):

| Token | Valor | Uso |
|---|---|---|
| `--space-section` | `clamp(5rem, 11vw, 8.5rem)` | separación **entre secciones** (padding-top de cada sección; el pie lo hereda como margen) |
| `--space-block` | `clamp(1.4rem, 3vw, 2.25rem)` | separación **entre bloques** dentro de una sección (título → texto → CTA; paneles de la misma fila) |
| `--space-fluid` | `clamp(.8rem, 2vw, 1.25rem)` | separación fina (filas de contacto, chips de redes, CTA del hero, grupos del pie) |
| `--pad-panel` | `clamp(2rem, 5vw, 3.5rem)` | **márgenes internos generosos** de los paneles de vidrio y del pie |
| `--pad-text` | `clamp(1rem, 2.4vw, 1.6rem)` | aire interno de texto y detalles |

Reglas de respiración: cada sección es un «plato» separado del anterior por `--space-section` (nunca menos de 5rem); dentro del panel, `* + *` respeta `--space-block` — la transición de bloque tiene aire, no es contenido pegado al borde; scrolled de la barra no reduce la respiración del contenido (`scroll-margin-top: 96px` compensa). La densidad se conserva donde es útil (filas del horario), el aire se gana **entre** secciones y bloques.

### 0.4 Temas (claro/oscuro con interruptor)
- «Claro» de B = navy `#1a2b4a` (reposo natural de la propuesta); «oscuro» profundiza a `#0b1424`, paneles `rgba(16,27,48,.7)`, blobs más sutiles, turquesa como acento directo AA sobre oscuro.
- Interruptor circular 44 px en la barra: `aria-pressed` + `aria-label` dinámico; cambio inmediato sin recargar; elección recordada (localStorage); primera visita según `prefers-color-scheme`. Logo intacto en ambos.

### 0.5 Navegación (con el cambio 3 del cliente)
- **Escritorio (> 860 px)**: barra fija siempre visible, enlaces todos a la vista (wrap si hiciera falta), SIN hamburguesa. Al bajar: la barra se vuelve **banda de vidrio** (`rgba(12,22,38,.72)` + blur + borde inferior turquesa de 1 px) y se compacta — transición perceptible.
- **Móvil (≤ 860 px)**: la barra sigue fija y visible (marca + idioma + tema + botón hamburguesa); los enlaces viven en un **panel de menú accesible** descrito en §2.3 y §6.

### 0.6 Selector de idioma compacto
«Español / English» en 0,82–0,85 rem, texto completo, en pastilla fina con borde translúcido; activo con fondo turquesa suave y peso 700; sin píldoras de siglas ni menú. Comportamiento de F3 intacto (memoria entre visitas, primera visita en español, fallback es→en). Se muestra en barra y en la franja inferior del pie (como en F3).

### 0.7 `prefers-reduced-motion` (incondicional)
Con esa preferencia: sin blobs animados, sin revelados, sin suavidad de scroll, sin transiciones de entrada — el contenido aparece completo de inmediato y todo funciona igual. Media query CSS que anula transiciones/animaciones + guarda en el JS del observador. Referencia SC-005.

---

## 1. Flujo de usuario

1. **Llegada**: el visitante abre la portada publicada → barra fija arriba (navy; en móvil con botón hamburguesa), hero/identidad revelándose con entrada lateral.
2. **Recorrido**: al hacer scroll cada sección entra por un lado alternando, con stagger interno → contenido siempre presente («capa visual», nunca bloquea lectura ni acción).
3. **Navegación — escritorio**: clic en un enlace de la barra → desplazamiento suave a la ancla con `scroll-margin-top` (la barra jamás tapa el título); barra pasa a vidrio.
4. **Navegación — móvil**: toque en hamburguesa → panel desde el borde superior con bloques de enlaces en stagger, scroll de fondo bloqueado, foco entra al primer enlace → toque en enlace → cierra el panel y salta a la sección; Escape, toque fuera o el propio botón también cierran; al cerrar el foco vuelve al botón.
5. **Idioma**: clic en «Español» o «English» → cambio inmediato de todos los textos e `lang` del documento; sin repetir animaciones de entrada (FR-008).
6. **Tema**: clic en el círculo 🌙/☀️ → cambia de inmediato; se recuerda entre visitas.
7. **Pie de página**: al terminar el recorrido, el visitante encuentra el **mapa del sitio** (grupos con encabezados y enlaces a todo lo existente), datos de contacto y redes, y el selector de idioma.
8. **Caminos de error/degradación**: ver §3 (estados) — claros: movimiento reducido, contenido en borrador que oculta la sección y actualiza nav/mapa sin enlaces rotos, contenido solo en español navegando en inglés (fallback), 320 px y zoom 200 %. F10 no agrega operaciones que fallen.

---

## 2. Pantallas

### 2.1 Portada (única pantalla del sitio hoy; orden de F3 intacto)

Barra fija → hero/identidad → 01 Quiénes somos → 02 Horario → 03 WhatsApp → 04 Contacto → 05 Redes → pie con mapa del sitio.

```
DESKTOP
[🌱 SIMIENTE SANTA | Quiénes somos·Horario·WhatsApp·Contacto·Redes | Español / English | 🌙]  ← fija; al bajar: vidrio + blur + borde turquesa
══════════════════════════════════════════════════════════════════
        (navy; blob turquesa→navy detrás del titular)
  BIENVENIDO A TU            ┌─ panel de vidrio ─┐
  CASA EN CRISTO  (Bebas)    │  imagen portada    │      ← hero, texto izq / imagen der
  «Brotamos para dar vida.»  └────────────────────┘
  [ Ver horario ] [ Escribir por WhatsApp ]
        ┐ aire = --space-section (≥ 5rem, clamp responsive)
  ← 01 QUIÉNES SOMOS   (panel de vidrio, padding generoso)
        ┐
  → 02 HORARIO  (dígitos Bebas turquesa grandes)
        ┐
  ← 03 WHATSAPP  (CTA verde)
        ┐
  → 04 CONTACTO (filas-pastilla: dirección/correo/teléfono)
        ┐
  ← 05 REDES    (chips con hover elevado)
        ┐
  PIE · MAPA DEL SITIO (columnas):
  | marca+lema | Explorar          | Contacto              | Síguenos      |
  |             | Inicio            | Dirección · Correo    | FB · IG · YT  |
  |             | Quiénes somos     | Teléfono · WhatsApp   |               |
  |             | Horario·WhatsApp  |                       |               |
  |             | Contacto·Redes    |                       |               |
  ─ franja inferior: lema ············ [ Español / English ]
══════════════════════════════════════════════════════════════════
```

### 2.2 Portada — MÓVIL (menú cerrado)
```
MOBILE (≤ 860 px), menú CERRADO
[ SS SIMIENTE SANTA        Español/English  🌙  ☰ ]  ← fija, 1 fila, sin wrap
──────────────────────────────────────────────
  BIENVENIDO A TU CASA EN CRISTO
  «Brotamos para dar vida.»
  [ imagen portada en columna ]
  [ Ver horario ] [ WhatsApp ]
        ┐ sección por sección, mismas separaciones (clamp responsivo)
  … 01→05 igual, llegando desde arriba en un solo eje …
  PIE apilado: marca+lema abajo ▸ Explorar ▸ Contacto ▸ Síguenos ▸ franja idioma
```
En pantallas < 480 px la palabra de la marca se oculta (queda el monograma) para que idioma + tema + hamburguesa quepan en 320 px sin overflow; objetivos táctiles ≥ 44 px (botón de tema 44, hamburguesa 44, y el chip de idioma compensa con padding hasta 44).

### 2.3 Portada — MÓVIL (menú ABIERTO)
```
MOBILE (≤ 860 px), menú ABIERTO                  foco → «Inicio»
┌─ barra fija (sigue visible) ─────────────────────┐
│  SS SIMIENTE…      Español/English   🌙   ✕      │  ← ☰ pasa a ✕
├──────────────────────────────────────────────────┤
│   ▓▓ vidrio: rgba(13,22,38,.92) + blur ▓▓        │
│                                                  │
│   Inicio                              ›          │  ← stagger 60 ms por bloques
│   Quiénes somos                       ›          │
│   Horario                             ›          │
│   WhatsApp                            ›          │
│   Contacto                            ›          │
│   Redes                               ›          │
│                                                  │        ← resto: scrim oscuro
└──────────────────────────────────────────────────┘  (scroll de fondo bloqueado)
```
- Botón: 44×44 px, icono 3 barras que se cruzan a ✕, `aria-expanded` + `aria-controls` + nombre accesible dinámico («Abrir/Cerrar menú de secciones»).
- Panel: cubre desde debajo de la barra, de vidrio con blur, con scrim oscuro detrás; enlaces a ancho completo, min-height 56 px cada uno.
- Gestión del foco: al abrir va al primer enlace; Tab/Shift+Tab cicla solo entre botón y panel; Escape cierra y **devuelve el foco al botón**; clic en un enlace o en el scrim también cierra; sin gestos desconocidos (nada de swipe); al agrandar la ventana a escritorio el menú se cierra solo.
- El scroll de la página se bloquea mientras está abierto; el tema oscuro de la barra se mantiene.

### 2.4 Pie — mapa del sitio (detalle)
- `nav aria-label="Mapa del sitio"` con 3 grupos, cada uno con **encabezado de grupo** (h2 en versalitas turquesa) y lista de enlaces reales ≥ 44 px de alto:
  - **Explorar**: Inicio, Quiénes somos, Horario de servicios, WhatsApp, Contacto, Redes sociales (anclas de la portada — todo lo que hoy existe).
  - **Contacto**: Dirección del templo, Correo, Teléfono, WhatsApp (mismos datos del bloque 04).
  - **Síguenos**: Facebook, Instagram, YouTube (mismos enlaces del bloque 05).
- **Preparado para F4–F9**: el patrón es «una columna = un grupo»; para una sección futura se AÑADE una columna SOLO cuando la sección exista y esté publicada (p. ej. «Eventos» → F4); no se inventan enlaces. La navegación de la cabecera y el mapa comparten la misma lista de origen (un único array `links` en front, para que cabecera y pie jamás se desincronicen).
- Franja inferior: lema a la izquierda y selector de idioma a la derecha (comportamiento de F3).
- Escritorio: 4 columnas en grilla (marca 1,3fr + 3 grupos). Móvil: **apilado**, orden marca → Explorar → Contacto → Síguenos → franja de idioma.

---

## 3. Estados (por vista con datos)

F10 no cambia datos ni estados de publicación de F3 — presenta con el sistema. Vista por vista:

| Vista/estado | Qué muestra |
|---|---|
| **Cargando** (fetch de portada en curso) | `SectionSkeleton` re-pensada para B: paneles de vidrio fantasmas con las mismas medidas y separaciones de la sección destino, pulso sutil (solo si hay movimiento). La barra y el hero estructural ya visibles → nada «brinca» al cargar. |
| **Éxito** | Secciones publicadas con estilo final (fotos en el marco de vidrio del hero; hora en Bebas turquesa en horario; filas-pastilla en contacto; chips en redes). |
| **Error (falla la carga del contenido)** | Igual que F3: mensaje claro bilingüe en un panel de vidrio con acción de reintento; el maquetado no se rompe; el contenido nunca depende de una animación cargada (FR-009). |
| **Vacío / sección en borrador** | La sección NO se dibuja; la navegación y el mapa del sitio se actualizan del mismo array de enlaces → 0 enlaces rotos o huérfanos (edge case de F3). |
| **Solo español navegando en inglés (fallback)** | Las piezas en español se integran con el mismo aspecto; sin huecos ni recolocaciones por la flexibilidad del espaciado. |
| **`prefers-reduced-motion`** | Todo el contenido visible al instante; sin blobs, revelados, suavidad ni transiciones; el menú móvil funciona sin animación; selector de idioma e idioma y tema intactos. SC-005. |
| **Sin permisos** | N/A en el sitio público; el panel está fuera (F10B). |
| **320 px / zoom 200 % / lento** | Sin desbordes horizontales ni overflow; fila única de barra con monograma; objetivos ≥ 44 px; movimiento siempre en `transform/opacity` (nunca bloquea la lectura). |

---

## 4. Componentes React (para `dev-frontend`)

### 4.1 Reutilizados (F3 — comportamiento intacto)
- `PublicLayout`, `LanguageSwitcher`, `HomePage`, `IdentityHero` (hoy `Hero`/`WhoWeAreSection`), `ScheduleSection`, `WhatsAppSection`, `ContactSection`, `SocialSection`, `SectionSkeleton`, `SocialIcon`, `Button` (se le añaden variantes con feedback táctil `active:scale`).
- Ninguno cambia su contrato de datos ni su lógica; solo se les aplica la maqueta B nueva.

### 4.2 Nuevos (cimientos del sistema)

| Componente / hook | Props principales | Responsabilidad |
|---|---|---|
| `Reveal` + `useReveal` | `as`, `from: 'left'|'right'|'up'|'in'`, `delay?: ms`, `once: true` | Envoltorio de revelado (IntersectionObserver umbral .15, rootMargin -8%); dispara UNA VEZ; si un elemento aun no revelado recibe foco se hace visible de inmediato; con `prefers-reduced-motion` no observa y muestra de inmediato. Nunca oculta foco ni bloquea lectura. |
| `TransitionHeader` | `links: {id,label}[]`, `breakpoint?: 860`, slots `lang`, `theme`, `menu` (móvil) | Barra fija: clase `scrolled` al bajar (vidrio+blur+borde turquesa+altura compacta). En escritorio renderiza nav en línea; en móvil oculta la nav y muestra el botón hamburguesa que controla `MobileMenu`. Cumple FR-014 (fija) y FR-025 (320 px). |
| `MobileMenu` ⭐new | `open: boolean`, `onClose?: () => void`, `links: {id,label}[]`, `labels: {open, close}` | Panel de menú para pantallas estrechas: `<nav>` con los enlaces de `links`; `.open` anima entrada 0,35 s con stagger 60 ms por ítem; maneja foco (entra al abrir, Escape cierra y devuelve foco, Tab encapsula), scrim opcional (`onScrimClick`), bloqueo de scroll de fondo, cierre automático si el breakpoint crece a escritorio. Componente **controlado**: no decide abrirse por sí mismo. |
| `useMobileMenu` | — | Hook con la lógica del menú (open/close, `aria-expanded`, etiqueta dinámica). Opcional si `MobileMenu` controlado basta. |
| `SiteMapFooter` ⭐new | `groups: {heading: string, label: string, items: {href, label}[]}[]`, `brand`, `lema`, `contact`, `social?`, `langSlot?: ReactNode`, `columnsDesktop: 1.3fr 1fr 1.1fr 1.1fr` | Pie con mapa del sitio: recibe ya los grupos de enlaces existentes (array fuente compartido con la nav) y los pinta en columnas (desktop) / apilado (móvil); franja inferior con lema + slot para el selector de idioma. `<nav aria-label="Mapa del sitio">` con encabezados de grupo visibles. Para F4–F9 basta añadir un grupo más; no inventa enlaces. |
| `ThemeToggle` + `useTheme` + `ThemeScript` | — | Interruptor 44 px con `aria-pressed`/`aria-label` dinámicos; persistencia; script inline anti-destello; observación en vivo de `prefers-color-scheme` si no hay elección manual. |
| `tokens` en `frontend/src/index.css` (`@theme`) + vars de tema oscuro | — | Espaciado (`--space-section/block/fluid`, `--pad-panel/text`), duraciones (`--dur-reveal: .75s`, `--dur-micro: .3s`, `--ease`, stagger 60–140 ms), paleta y temas; documentados para F4–F9. |
| `features/publico/messages.ts` | — | Cadenas nuevas es/en (ver §7). |

### 4.3 Flujo de datos del nav/mapa
Una única fuente `useSiteLinks()` (o equivalente) devuelve los enlaces visibles según el contenido publicado (las secciones en borrador se excluyen). De ella se alimentan `TransitionHeader` (escritorio y móvil vía `MobileMenu`) y `SiteMapFooter` — la cabecera y el pie nunca se desincronizan y actualizar la fuente es suficiente para ambas.

---

## 5. Sistema de movimiento (nivel marcado/elegante con revelado al scroll — FR-005)

1. **Revelado de secciones**: cada panel entra desde un lado alternado (± 56–60 px) + `opacity`, `0,7–0,8 s`, curva expansiva `cubic-bezier(.16,1,.3,1)`; **stagger 110 ms** entre componentes internos (kicker → título → texto → CTA); en móvil todo entra desde un solo eje (arriba).
2. **Blobs de gradiente**: 2 blobs en deriva continua al fondo, solo `transform` a 60 fps (26/32 s, alternos). Solo si hay movimiento; en dispositivos modestos el blur se modera (es el mayor costo de B — se vigila el framerate en las pruebas).
3. **Transición de cabecera**: al bajar, fondo pasa a vidrio (`rgba(12,22,38,.72)` + `backdrop-filter: blur(12px)`), altura compacta y borde inferior turquesa de 1 px que aparece — perceptible, no un «brinco».
4. **Microinteracciones**: hover de paneles = elevación 6 px + borde turquesa; botones escalan 1,02–1,04 al activarse; filas de contacto se desplazan 6 px a la derecha; chips de redes suben y se rellenan levemente; subrayado turquesa se despliega en la nav. Todo en `transform/opacity`, sin desplazar el maquetado.
5. **Menú móvil**: panel entra con `translateY(-14px) → 0` + fade 0,35 s; ítems con stagger 60 ms; icono ☰ → ✕ (cruce de barras). Con movimiento reducido: apertura/cierre instantáneos.
6. **`prefers-reduced-motion`** (FR-006, incondicional): sin blobs, revelado, suavidad ni transiciones de entrada — contenido completo al instante; el sitio funciona igual para teclado y lector de pantalla. Media query CSS global + guarda en JS.
7. **Nunca**: bloqueo de lectura o acción (FR-007), repetición molesta al cambiar idioma/tema o al revisitar una sección (FR-008); el contenido y las funciones jamás dependen de ver una animación (FR-009).

---

## 6. Accesibilidad (WCAG 2.1 AA, ambos temas)

- **Orden de teclado**: skip link → logo → enlaces de nav (móvil: primero hamburguesa) → selector de idioma → interruptor de tema → contenido → … → mapa del sitio → idioma del pie.
- **Menú móvil (patrón accesible completo)**: botón 44 px con nombre accesible dinámico y `aria-expanded`; `aria-controls` apuntando al panel (con `role="dialog"` ligero); foco entra al abrir (`primer enlace`), Tab/Shift+Tab encapsula (botón↔panel), Escape cierra y devuelve foco al botón, cierre «explícito» desde el propio control; scrim con clic cierra; scroll de fondo bloqueado; sin gestos desconocidos; `aria-modal="false"` (la página sigue operable por lector de pantalla en la barra).
- **Tema**: botón 44 px, `aria-pressed` dinámica + `aria-label` "Cambiar a tema claro/oscuro"; cambio de tema sin perder foco.
- **Foco visible**: anillo turquesa 2 px; contraste del outline verificado sobre navy claro/oscuro y sobre vidrio en estado `scrolled` y menú.
- **Contraste AA en ambos temas**: revisión con script (llamada prevista en implementación): texto ≥ 4,5:1, grande ≥ 3:1, incluidos `hover/focus/active` y turquesa pequeño solo `#0e7a66` sobre claro. Blobs: son decorativos (`aria-hidden`), nunca tras texto esencial con contraste insuficiente.
- **Objetivos ≥ 44 px**: navegación, selector de idioma, tema, hamburguesa, CTAs, filas de contacto, chips de redes, enlaces del pie, ítems del menú móvil.
- **Semántica**: `<nav aria-label="Secciones del sitio">`; `role="group" aria-label="Idioma"`; `<nav aria-label="Mapa del sitio">` con encabezados de grupo (h2) y listas; `lang` en `<html>` según idioma (comportamiento F3). Landmarks limpios: banner (header), main, contentinfo (footer).
- **Movimiento**: nunca condición para ver ni usar contenido (FR-009); esc. de foco temprano resuelto (elemento a revelar se visibiliza si recibe foco).
- **320 px y zoom 200 %**: sin overflow, sin tapar nada, sin scroll lateral (FR-025, heredado F3).

## 7. Textos de interfaz nuevos (es/en, FR-019)

| Clave | Español | English |
|---|---|---|
| `menu.open` | Abrir menú de secciones | Open sections menu |
| `menu.close` | Cerrar menú de secciones | Close sections menu |
| `menu.groupLabel` | Menú de secciones | Sections menu |
| `map.groupLabel` | Mapa del sitio | Site map |
| `map.explore` | Explorar | Explore |
| `map.contact` | Contacto | Contact |
| `map.social` | Síguenos | Follow us |
| `nav.groupLabel` | Secciones del sitio | Site sections |
| `theme.label` | Tema | Theme |
| `theme.toDark` | Cambiar a tema oscuro | Switch to dark theme |
| `theme.toLight` | Cambiar a tema claro | Switch to light theme |
| `lang.groupLabel` | Idioma | Language |
| `skipToContent` | Ir al contenido | Skip to content |
| `content.error.*` / `content.retry` | (de F3, conservados y vestidos con B) | |

Los rótulos de contenido ya existen en F3 y no cambian de texto ni de clave. La maqueta demo bilingüe es solo para percibir el resultado.

## 8. Preparado para F4–F9 (sin inventar enlaces)

- **Nav**: nuevo enlace = nuevo ítem en la única fuente de enlaces; escritorio wrap/dos filas; móvil entra automáticamente al panel; el mapa del sitio recibe un nuevo grupo o ítem **solo cuando la sección exista**.
- **Secciones nuevas** heredan: `Reveal` (variante lateral), `.glass` + tokens `--space-block/--pad-panel`, botones existentes, headings del sistema, sin rediseño.
- No se muestran enlaces de secciones futuras en ninguna superficie (nada muerto, nada «pronto»).

## 9. Puntos a confirmar

1. **spec.md**: FR-012 («sin menús hamburguesa») y US2 esc. 5 quedan sustituidos en móvil por la decisión de ronda 3 del cliente (hamburguesa en móvil; escritorio sin hamburguesa, barra fija). Aprobar la redacción de esta excepción con el analista antes de implementar (este documento no edita `spec.md`).
2. **Grupos del mapa del sitio**: se propuso «Explorar / Contacto / Síguenos» (con «Inicio»). Ajustar rotulación si el cliente quiere otros nombres.
3. **Selector de idioma del pie**: se conserva (F3 lo tenía) en la franja inferior. Confirmar si además debe mostrarse dentro del panel del menú móvil (hoy en la maqueta solo queda en la barra).
4. **Rendimiento de blobs+blur**: en dispositivos muy modestos los gradientes animados y el `backdrop-filter` son el mayor costo de B; si degradan la fluidez se moderan o se reducen antes que entorpecer la lectura (FR-007). Se resuelve en implementación.
5. **Orden de los grupos del pie en móvil**: confirmar que se apilan en este orden: marca → Explorar → Contacto → Síguenos → franja de idioma (así está en la maqueta).
