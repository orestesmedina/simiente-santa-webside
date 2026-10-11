# UX: F10 — Sistema de diseño y movimiento — **Diseño final: Propuesta B «Marea Turquesa»**

> **Estado (ronda 3, 2026-10-10, cierre de fase de diseño):** el cliente **eligió la Propuesta B «Marea Turquesa» y aprobó el plan, con tres cambios** que este documento ya integra: (1) **más aire** — nueva escala de espaciado vertical con tokens de sección; (2) **mapa del sitio en el pie** — todas las opciones existentes del sitio público, en columnas/apilado, accesible y sin inventar enlaces; (3) **menú hamburguesa en móvil** (escritorio conserva la barra fija sin hamburguesa). Este documento **reemplaza a las 3 opciones** de la ronda 2 como **documento de diseño final** listo para `dev-frontend`. La maqueta de referencia vigente es `propuestas/propuesta-b.html` (HTML+CSS+JS autocontenido, con movimiento real, tema claro/oscuro, selector de idioma demo, hamburguesa y mapa del sitio).
>
> **Cierre de la puerta 2 (2026-10-10) — dos decisiones del cliente ya integradas:** (4) el **selector de idioma NO se repite dentro del panel del menú hamburguesa** — vive solo en la **cabecera** (compacto) y en el **pie**; (5) los **rótulos del mapa del sitio** del pie son **Explorar / Contacto / Síguenos**.
>
> Fuentes: `spec.md` de F10 (US1–US6; FR-005/006/007/009/026), tokens de F3 (`specs/003-portada-info-general/ux.md`), Manual de Identidad (`resources/MANUAL DE MARCA.pdf` manda), referencias de inspiración del cliente (apple.com, microsoft.com — **inspiración, nunca copia**). Sin panel (F10B fuera). No toca `estado.md` ni el roadmap; no implementa la app.
>
> **Correcciones de coherencia (post-analyze, 2026-10-10), ya integradas en este documento:** (A) el panel del menú móvil **reutiliza la misma lista** del `<nav>` «Secciones de la portada» — una sola lista en el DOM, cero ítems extra dentro del landmark con nombre congelado; (B) menú con **patrón *disclosure* (APG)**: sin `role="dialog"`/`aria-modal`, sin trampa de foco — la barra sigue operable; (C) **colapso explícito**: la lista cerrada queda fuera del orden de tabulación (`display:none` bajo la media query); (D) mapa del pie: **grupo sin ítems no se renderiza**, **mapa sin grupos no se renderiza**, rótulos de grupo **sin rol `heading`** — etiqueta no-heading, ni `h2` ni `h3`: las consultas congeladas de `homePage.test.tsx` resuelven encabezados por nombre a cualquier nivel — y **nombres accesibles únicos** frente a los enlaces de las secciones; (E) marca < 480 px: la palabra se oculta **solo visualmente** (sr-only) — el nombre accesible conserva «Simiente» a 320 px; (F) claves i18n alineadas al contrato congelado de F3 (§7) y `ScheduleSection` conserva su representación dual (§4.1); (G) números de movimiento unificados con el plan: **revelado 600 ms, techo absoluto ≤ 800 ms** (§5).
>
> ~~⚠️ **Punto a confirmar para el analista:** `spec.md` FR-012 y US2 esc. 5 aún dicen «sin menús hamburguesa» (decisión ronda 1).~~ ✅ **Ya resuelto**: la spec quedó actualizada en la puerta 2 (FR-010…FR-014 describen la navegación **por dispositivo**: barra fija en escritorio, menú hamburguesa accesible en móvil). **No** se vuelve a «corregir» la spec hacia atrás.

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
«Español / English» en 0,82–0,85 rem, texto completo, en pastilla fina con borde translúcido; activo con fondo turquesa suave y peso 700; sin píldoras de siglas ni menú. Comportamiento de F3 intacto (memoria entre visitas, primera visita en español, fallback es→en). Se muestra en la **cabecera** (compacto) y en la **franja inferior del pie** (como en F3) — son sus **únicos** lugares: por decisión del cliente (cierre de puerta 2, 2026-10-10) el selector **no se repite dentro del panel del menú móvil**.

### 0.7 `prefers-reduced-motion` (incondicional)
Con esa preferencia: sin blobs animados, sin revelados, sin suavidad de scroll, sin transiciones de entrada — el contenido aparece completo de inmediato y todo funciona igual. Media query CSS que anula transiciones/animaciones + guarda en el JS del observador. Referencia SC-005.

---

## 1. Flujo de usuario

1. **Llegada**: el visitante abre la portada publicada → barra fija arriba (navy; en móvil con botón hamburguesa), hero/identidad revelándose con entrada lateral.
2. **Recorrido**: al hacer scroll cada sección entra por un lado alternando, con stagger interno → contenido siempre presente («capa visual», nunca bloquea lectura ni acción).
3. **Navegación — escritorio**: clic en un enlace de la barra → desplazamiento suave a la ancla con `scroll-margin-top` (la barra jamás tapa el título); barra pasa a vidrio.
4. **Navegación — móvil**: toque en hamburguesa → la **misma lista del nav** se recoloca como panel desde el borde superior (patrón *disclosure*: sin diálogo, sin trampa de foco; la barra sigue operable), con bloques de enlaces en stagger corto, scroll de fondo bloqueado y colapso explícito al cerrar (lista fuera del tab order) → toque en enlace → cierra el panel y salta a la sección; Escape o el propio botón cierran y devuelven el foco al botón; toque fuera (scrim) también cierra.
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
  [ Ver horario ] [ Escribir por WhatsApp ]      ← el CTA de WhatsApp usa un nombre
        ┐ aire = --space-section (≥ 5rem, clamp responsive)     propio (p. ej. «Escríbenos por
                                                                  WhatsApp»): «Escribir por WhatsApp»
                                                                  pertenece al canal de la sección 03
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
[ SS SIMIENTE SANTA        Español/English  🌙  ☰ ]  ← fija, 1 fila, sin wrap; enlace de marca con nombre accesible completo
──────────────────────────────────────────────
  BIENVENIDO A TU CASA EN CRISTO
  «Brotamos para dar vida.»
  [ imagen portada en columna ]
  [ Ver horario ] [ WhatsApp ]
        ┐ sección por sección, mismas separaciones (clamp responsivo)
  … 01→05 igual, llegando desde arriba en un solo eje …
  PIE apilado: marca+lema abajo ▸ Explorar ▸ Contacto ▸ Síguenos ▸ franja idioma
```
En pantallas < 480 px la palabra de la marca se oculta **solo visualmente** con técnica `sr-only`/visually-hidden — **nunca `display:none`**, que la expulsaría del `innerText` — para que idioma + tema + hamburguesa quepan en 320 px sin overflow. El **enlace de marca de la cabecera conserva su nombre accesible completo con «Simiente»** (visible o sr-only); las pruebas congeladas de accesibilidad exigen a 320 px un enlace enfocable cuyo texto renderizado incluya «Simiente» y ese requisito lo garantiza **el enlace de la cabecera** (la marca del pie se mantiene como **texto**, sin enlace propio: no consume paradas del ciclo de tabulación — recuento en `tasks.md` T1015). Objetivos táctiles ≥ 44 px (botón de tema 44, hamburguesa 44, y el chip de idioma compensa con padding hasta 44).

### 2.3 Portada — MÓVIL (menú ABIERTO, patrón *disclosure*)
```
MOBILE (≤ 860 px), menú ABIERTO (disclosure)      foco → primer enlace de la lista
┌─ barra fija (sigue visible y operable) ──────────┐
│  SS SIMIENTE…      Español/English   🌙   ✕      │  ← ☰ pasa a ✕
├──────────────────────────────────────────────────┤
│   ▓▓ EL MISMO <nav> de la barra, recolocado…▓▓   │
│   por CSS: la MISMA lista, sin ítems añadidos    │
│                                                  │
│   Quiénes somos                       ›          │  ← stagger corto por ítem
│   Horario de servicios                ›          │
│   WhatsApp                            ›          │
│   Contacto                            ›          │
│   Redes sociales                      ›          │
│   (los ítems son EXACTAMENTE la lista dinámica    │
│    de secciones publicadas compartida con la     │
│    nav; nada hardcodeado, nada duplicado)         │        ← resto: scrim oscuro
└──────────────────────────────────────────────────┘  (scroll de fondo bloqueado)
```
- **Una sola lista (inviolable)**: el panel **reutiliza la misma lista** del `<nav aria-label={t('nav.label')}>` «Secciones de la portada» — dentro de ese landmark nombrado hay **un único** juego de nodos, ni en escritorio ni en móvil. El panel no monta una segunda lista dentro del `<nav>`: o reubica visualmente los mismos nodos con CSS, o su contenedor vive fuera del landmark. El botón hamburguesa sí va dentro del `<nav>` (es `<button>`, no compite por nombres de enlace). Los ítems son los que emite la fuente dinámica de `useSiteLinks()` — **sin enlaces fijados a mano** (las pruebas congeladas exigen coincidencia única por nombre dentro del `<nav>` y, con secciones en borrador, que «WhatsApp»/«Contacto» **no** existan ahí).
- **Colapso explícito**: con el panel cerrado a ≤ 860 px, la lista se oculta con una técnica que la **saca del orden de tabulación y del árbol de accesibilidad** (clase que aplica `display:none` bajo la media query, p. ej. `hidden`/equivalente Tailwind). Los enlaces nunca reciben foco en estado cerrado; nunca se oculta solo con `max-height`/`opacity`/`transform` (foco invisible). Los nodos siguen en el DOM para jsdom (las pruebas unitarias congeladas los ven igual). Al abrir: visible y enfocable.
- Botón: 44×44 px, icono 3 barras que se cruzan a ✕, controla el panel con **`aria-expanded` + `aria-controls`** (apunta a la lista del nav) y nombre accesible dinámico («Abrir/Cerrar menú de secciones»).
- Panel: cubre desde debajo de la barra, de vidrio con blur, con scrim oscuro detrás; enlaces a ancho completo, min-height 56 px cada uno.
- **Gestión del foco — patrón *disclosure* (APG), sin diálogo modal**: al abrir el foco va al **primer enlace**; **sin trampa de foco** y **sin `role="dialog"`/`aria-modal`** — la página sigue operable: `Tab`/`Shift+Tab` continúan por la barra (selector de idioma e interruptor de tema siguen al alcance por teclado); **Escape cierra el panel y devuelve el foco al botón**; clic en un enlace o en el scrim también cierra; sin gestos desconocidos (nada de swipe); al agrandar la ventana a escritorio el menú se cierra solo.
- El scroll de la página se bloquea mientras está abierto; el tema oscuro de la barra se mantiene.
- **Sin selector de idioma dentro del panel** (decisión del cliente, cierre de puerta 2): el panel contiene solo los enlaces de secciones; el idioma sigue al alcance en la **barra** (visible y operable arriba del panel, alcanzable también por teclado por no haber trampa de foco) y en el **pie**.

### 2.4 Pie — mapa del sitio (detalle)
- `nav aria-label="Mapa del sitio"`. Cada grupo con su **rótulo de grupo como etiqueta NO-heading** (`<p>`/`<span>` en versalitas turquesa) — **nunca `h2`, `h3` ni ningún rol `heading`**: `homePage.test.tsx` (congelada) resuelve encabezados **por nombre a cualquier nivel** — con datos completos exige `getByRole('heading', { name: 'Contacto' })` **único** (el `h2` de la sección 04) y con la respuesta vacía que **no exista ningún `heading` de nivel 2** en la página completa (cabecera + pie) —; con las reglas de abajo ningún rótulo puede colisionar con los nombres de sección.
- **Reglas de renderizado (inviolables contra las pruebas congeladas):**
  1. **Un grupo sin ítems no se renderiza**: «Contacto» solo aparece si hay datos de contacto; «Síguenos» solo si hay redes. El grupo «Explorar» solo aparece si hay al menos una sección publicada.
  2. **El mapa del sitio no se renderiza si no hay nada** (0 secciones publicadas, sin contacto ni redes): el pie queda con marca/lema + franja de idioma, sin `mapa` vacío ni rótulos huérfanos.
  3. Los 3 grupos («Explorar / Contacto / Síguenos») describen el caso **con datos completos**, no un requisito fijo de renderizado.
- **Nombres accesibles únicos (inviolable)**: ningún enlace del pie repite el nombre accesible de un enlace de sección ni el texto de `social.action` («Ver en Facebook»):
  - Grupo **Contacto**: nombres **prefijados de contexto** — «Contacto: Dirección», «Contacto: Correo», «Contacto: Teléfono», «Contacto: WhatsApp» (o `aria-label` localizado equivalente). La **dirección es texto**, no enlace; «Contacto: WhatsApp» es **un solo enlace** (el primer canal publicado). Jamás el dato desnudo como nombre (un `mailto:`/`tel:` duplicado haría fallar las aserciones de coincidencia única de `homePage.test.tsx`) ni los textos CTA de la sección 03 («Escribir por WhatsApp»/«Entrar al grupo» son de `WhatsAppSection`).
  - Grupo **Síguenos**: etiquetas **planas** de `SOCIAL_NETWORK_LABELS` («Facebook», «Instagram», «YouTube») — nunca el `social.action` de las secciones ni un CTA repetido con ese texto.
  - Aspa propia en `FooterSiteMap.test.tsx`: ningún nombre accesible del pie duplica el de un enlace de sección de la página.
- Grupos con datos completos:
  - **Explorar**: Inicio, Quiénes somos, Horario de servicios, WhatsApp, Contacto, Redes sociales (anclas de la portada — todo lo que hoy existe).
  - **Contacto**: «Contacto: Dirección» (texto, **no** enlace), «Contacto: Correo» (`mailto:`), «Contacto: Teléfono» (`tel:`), «Contacto: WhatsApp» (un solo enlace: el primer canal publicado) — mismos datos del bloque 04.
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
| **320 px / zoom 200 % / lento** | Sin desbordes horizontales ni overflow; fila única de barra con la palabra de la marca en sr-only (<480 px) — nombre accesible «Simiente» íntegro; objetivos ≥ 44 px; movimiento siempre en `transform/opacity` (nunca bloquea la lectura). |

---

## 4. Componentes React (para `dev-frontend`)

### 4.1 Reutilizados (F3 — comportamiento intacto)
- `PublicLayout`, `LanguageSwitcher`, `HomePage`, `IdentityHero` (hoy `Hero`/`WhoWeAreSection`), `ScheduleSection`, `WhatsAppSection`, `ContactSection`, `SocialSection`, `SectionSkeleton`, `SocialIcon`, `Button` (se le añaden variantes con feedback táctil `active:scale`).
- Ninguno cambia su contrato de datos ni su lógica; solo se les aplica la maqueta B nueva.
- **`ScheduleSection` (contrato congelado, M2)**: conserva su **representación dual de F3** — tabla (visible desde `sm`, nombre accesible «Horario de servicios» que exige el e2e congelado) **y** lista de tarjetas (visible bajo `sm`, la que apunta la e2e de accesibilidad a 160 px) — con su breakpoint intacto. F10 solo cambia el vestido (tokens, tipografía en Bebas): no sustituye una estructura por la otra.

### 4.2 Nuevos (cimientos del sistema)

| Componente / hook | Props principales | Responsabilidad |
|---|---|---|
| `Reveal` + `useReveal` | `as`, `from: 'left'|'right'|'up'|'in'`, `delay?: ms`, `once: true` | Envoltorio de revelado (IntersectionObserver umbral .15, rootMargin -8%); dispara UNA VEZ; si un elemento aun no revelado recibe foco se hace visible de inmediato; con `prefers-reduced-motion` no observa y muestra de inmediato. Nunca oculta foco ni bloquea lectura. |
| `TransitionHeader` | `links: {id,label}[]`, `breakpoint?: 860`, slots `lang`, `theme`, `menu` (móvil) | Barra fija: clase `scrolled` al bajar (vidrio+blur+borde turquesa+altura compacta). **Usa un único juego de enlaces**: en escritorio el `<nav>` va en línea; en móvil ese MISMO `<nav>` (nombrado con `t('nav.label')`, congelado) se convierte en el panel del menú vía CSS — sin segunda lista dentro del landmark. El botón hamburguesa lo controla con `aria-expanded`/`aria-controls`. Cumple FR-014 (fija) y FR-025 (320 px). |
| `MobileMenu` ⭐new | `open: boolean`, `onClose?: () => void`, `lists: {id,label}[]` (la MISMA lista del nav), `labels: {open, close}` | Panel de menú para pantallas estrechas con **patrón *disclosure* (APG)**: **reutiliza la lista del `<nav>`** (un solo juego de nodos dentro del landmark nombrado; nunca una lista duplicada); botón con `aria-expanded`/`aria-controls`; `.open` anima ~180 ms (token micro) con stagger corto por ítem. Foco: entra al **primer enlace** al abrir, **Escape cierra y devuelve el foco al botón**, **sin trampa de foco** (la barra — idioma/tema — sigue operable por teclado), **sin `role="dialog"`/`aria-modal`**, **sin `tabIndex` negociados**: el colapso del panel cerrado usa una clase que lo saca del tab order/apagado (p. ej. `hidden` bajo la media query); scrim opcional (`onScrimClick`); bloqueo de scroll de fondo; cierre automático si el breakpoint crece a escritorio. Componente **controlado**: no decide abrirse por sí mismo. |
| `useMobileMenu` | — | Hook con la lógica del menú (open/close, `aria-expanded`, etiqueta dinámica). Opcional si `MobileMenu` controlado basta. |
| `SiteMapFooter` ⭐new | `groups: {label: string, items: {href, label}[]}[]` (solo con ítems), `brand`, `lema`, `contact`, `social?`, `langSlot?: ReactNode`, `columnsDesktop: 1.3fr 1fr 1.1fr 1.1fr` | Pie con mapa del sitio: recibe ya los grupos de enlaces existentes (array fuente compartido con la nav) y los pinta en columnas (desktop) / apilado (móvil); franja inferior con lema + slot para el selector de idioma. `<nav aria-label={t('footer.sitemap.label')}>`; **rótulos de grupo como etiquetas no-heading** (nunca un rol `heading`: colisionaría con las consultas congeladas de `homePage.test.tsx`); **un grupo sin ítems no se renderiza**; **si ningún grupo tiene ítems, el nav del mapa no se renderiza**. **Nombres accesibles únicos** (contacto prefijado «Contacto: …»; redes planas — ver §2.4); nunca un nombre que duplique un enlace de sección. Enlaces ≥ 44 px. Para F4–F9 basta añadir un grupo más; no inventa enlaces. |
| `ThemeToggle` + `useTheme` + `ThemeScript` | — | Interruptor 44 px con `aria-pressed`/`aria-label` dinámicos; persistencia; script inline anti-destello; observación en vivo de `prefers-color-scheme` si no hay elección manual. |
| `tokens` en `frontend/src/index.css` (`@theme`) + vars de tema oscuro | — | Espaciado (`--space-section/block/fluid`, `--pad-panel/text`); **movimiento con los valores del plan/T1004 (única fuente numérica — ver nota M3 en §5)**: `--motion-distance` 28 px (20 px a ≤480), `--motion-duration-reveal` **600 ms**, `--motion-duration-micro` 180 ms, `--motion-ease-out` `cubic-bezier(0.22,1,0.36,1)`, stagger 80 ms (tope 240 ms), techo absoluto **≤ 800 ms**; paleta y temas; documentados para F4–F9. |
| `features/publico/messages.ts` | — | Cadenas nuevas es/en (ver §7). |

### 4.3 Flujo de datos del nav/mapa
Una única fuente `useSiteLinks()` (o equivalente) devuelve los enlaces visibles según el contenido publicado (las secciones en borrador se excluyen). De ella se alimentan `TransitionHeader` (escritorio y móvil vía `MobileMenu`) y `SiteMapFooter` — la cabecera y el pie nunca se desincronizan y actualizar la fuente es suficiente para ambas.

---

## 5. Sistema de movimiento (nivel marcado/elegante con revelado al scroll — FR-005)

> **Fuente de verdad de los números (M3):** la tabla de tokens del plan (`P10-12` / tarea T1004) es la única referencia numérica; esta sección es **dirección visual** para percibir la intención. Lo que sigue ya usa esos valores: revelado **600 ms**, microinteracciones en el token micro (**180 ms** para el menú móvil), techo absoluto **≤ 800 ms** (no hay un «900 ms» en ningún artefacto; el e2e debe validar ≤ 800 ms). QA puede calibrar dentro de la tabla (RG10-22).

1. **Revelado de secciones**: cada panel entra desde un lado alternado con distancia del token `--motion-distance` (28 px; 20 px a ≤480 px) + `opacity`, duración **600 ms** con curva expansiva `cubic-bezier(0.22,1,0.36,1)`; **stagger 80 ms** (tope combinado 240 ms) entre componentes internos (kicker → título → texto → CTA); en móvil todo entra desde un solo eje (arriba).
2. **Blobs de gradiente**: 2 blobs en deriva continua al fondo, solo `transform` a 60 fps (26/32 s, alternos). Solo si hay movimiento; en dispositivos modestos el blur se modera (es el mayor costo de B — se vigila el framerate en las pruebas).
3. **Transición de cabecera**: al bajar, fondo pasa a vidrio (`rgba(12,22,38,.72)` + `backdrop-filter: blur(12px)`), altura compacta y borde inferior turquesa de 1 px que aparece — perceptible, no un «brinco».
4. **Microinteracciones**: hover de paneles = elevación 6 px + borde turquesa; botones escalan 1,02–1,04 al activarse; filas de contacto se desplazan 6 px a la derecha; chips de redes suben y se rellenan levemente; subrayado turquesa se despliega en la nav. Todo en `transform/opacity`, sin desplazar el maquetado.
5. **Menú móvil**: panel entra con `translateY(-14px) → 0` + fade con el token micro (**~180 ms**); ítems con stagger corto (≤ 240 ms combinado); icono ☰ → ✕ (cruce de barras). Con movimiento reducido: apertura/cierre instantáneos.
6. **`prefers-reduced-motion`** (FR-006, incondicional): sin blobs, revelado, suavidad ni transiciones de entrada — contenido completo al instante; el sitio funciona igual para teclado y lector de pantalla. Media query CSS global + guarda en JS.
7. **Nunca**: bloqueo de lectura o acción (FR-007), repetición molesta al cambiar idioma/tema o al revisitar una sección (FR-008); el contenido y las funciones jamás dependen de ver una animación (FR-009).

---

## 6. Accesibilidad (WCAG 2.1 AA, ambos temas)

- **Orden de teclado**: skip link → logo → hamburguesa (móvil) → enlaces de nav (escritorio) → selector de idioma → interruptor de tema → contenido → … → mapa del sitio → idioma del pie. Con el menú móvil abierto la lista oculta (cerrada) **no existe en el tab order** (colapso explícito, §2.3).
- **Menú móvil (patrón *disclosure* APG, sin diálogo)**: botón 44 px con nombre accesible dinámico, `aria-expanded` y `aria-controls` apuntando a la lista del nav (que es el propio panel — una sola lista, ver §2.3); **sin `role="dialog"` ni `aria-modal`** y **sin trampa de foco**: al abrir el foco entra al primer enlace, `Tab`/`Shift+Tab` siguen por la página (la barra sigue operable), **Escape cierra y devuelve el foco al botón**, clic en enlace o en el scrim cierra; scroll de fondo bloqueado; sin gestos desconocidos.
- **Tema**: botón 44 px, `aria-pressed` dinámica + `aria-label` "Cambiar a tema claro/oscuro"; cambio de tema sin perder foco.
- **Foco visible**: anillo turquesa 2 px; contraste del outline verificado sobre navy claro/oscuro y sobre vidrio en estado `scrolled` y menú.
- **Contraste AA en ambos temas**: revisión con script (llamada prevista en implementación): texto ≥ 4,5:1, grande ≥ 3:1, incluidos `hover/focus/active` y turquesa pequeño solo `#0e7a66` sobre claro. Blobs: son decorativos (`aria-hidden`), nunca tras texto esencial con contraste insuficiente.
- **Objetivos ≥ 44 px**: navegación, selector de idioma, tema, hamburguesa, CTAs, filas de contacto, chips de redes, enlaces del pie, ítems del menú móvil.
- **Semántica**: `<nav aria-label={t('nav.label')}>` («Secciones de la portada»/«Page sections» — **clave congelada de F3, única lista de enlaces dentro del landmark**); `role="group" aria-label={t('lang.label')}`; `<nav aria-label={t('footer.sitemap.label')}>` con **rótulos de grupo como etiquetas no-heading** (nunca un rol `heading` en el pie), listas, grupos condicionales y mapa ausente si no hay grupos (§2.4); `lang` en `<html>` según idioma (comportamiento F3). Landmarks limpios: banner (header), main, contentinfo (footer).
- **Movimiento**: nunca condición para ver ni usar contenido (FR-009); esc. de foco temprano resuelto (elemento a revelar se visibiliza si recibe foco).
- **320 px y zoom 200 %**: sin overflow, sin tapar nada, sin scroll lateral (FR-025, heredado F3).

## 7. Textos de interfaz nuevos (es/en, FR-019)

> ⚠️ **Contrato congelado de F3 (inviolable):** la implementación **reutiliza** las claves y textos de F3 sin cambiarlos — `nav.label` = «Secciones de la portada»/«Page sections», `a11y.skip` = «Saltar al contenido principal»/«Skip to main content», `lang.*`, `nav.sections.*`, `social.action`, `footer.welcome` — porque las pruebas congeladas los resuelven literalmente. **Este documento no propone sustitutos**: no existe `nav.groupLabel` ni `skipToContent` como claves nuevas, y la clave del botón de menú es **`nav.menu.open`/`nav.menu.close`** (mismo nombre en plan y tasks — nunca «toggle»). Las claves `footer.sitemap.*` y `theme.*` son las únicas agregadas a los diccionarios (el maquetado demo puede usar textos equivalentes, pero la implementación usa estas filas).

| Clave | Español | English |
|---|---|---|
| `nav.menu.open` | Abrir menú de secciones | Open sections menu |
| `nav.menu.close` | Cerrar menú de secciones | Close sections menu |
| `nav.menu.label` | Menú de secciones | Sections menu |
| `footer.sitemap.label` | Mapa del sitio | Site map |
| `footer.sitemap.explore` | Explorar | Explore |
| `footer.sitemap.home` | Inicio | Home |
| `footer.sitemap.contact` | Contacto | Contact |
| `footer.sitemap.social` | Síguenos | Follow us |
| `theme.label` | Tema | Theme |
| `theme.toDark` | Cambiar a tema oscuro | Switch to dark theme |
| `theme.toLight` | Cambiar a tema claro | Switch to light theme |
| `nav.label` *(congelada F3)* | Secciones de la portada | Page sections |
| `a11y.skip` *(congelada F3)* | Saltar al contenido principal | Skip to main content |
| `lang.label` *(congelada F3)* | Idioma | Language |
| `content.error.*` / `content.retry` | (de F3, conservados y vestidos con B) | |

Los rótulos de contenido ya existen en F3 y no cambian de texto ni de clave. La maqueta demo bilingüe es solo para percibir el resultado.

## 8. Preparado para F4–F9 (sin inventar enlaces)

- **Nav**: nuevo enlace = nuevo ítem en la única fuente de enlaces; escritorio wrap/dos filas; móvil entra automáticamente al panel; el mapa del sitio recibe un nuevo grupo o ítem **solo cuando la sección exista**.
- **Secciones nuevas** heredan: `Reveal` (variante lateral), `.glass` + tokens `--space-block/--pad-panel`, botones existentes, headings del sistema, sin rediseño.
- No se muestran enlaces de secciones futuras en ninguna superficie (nada muerto, nada «pronto»).

## 9. Puntos a confirmar

1. ~~**spec.md**: FR-012 («sin menús hamburguesa») y US2 esc. 5…~~ ✅ **Resuelto (puerta 2)**: `spec.md` ya describe la navegación **por dispositivo** (FR-010…FR-014: barra fija en escritorio, menú hamburguesa accesible en móvil). La excepción quedó aprobada y redactada; este documento **no** pide cambios en la spec ni debe «corregirla» hacia atrás.
2. ~~**Grupos del mapa del sitio**: se propuso «Explorar / Contacto / Síguenos» (con «Inicio»). Ajustar rotulación si el cliente quiere otros nombres.~~ ✅ **Resuelto por el cliente** (cierre de puerta 2, 2026-10-10): los rótulos del mapa del sitio son **Explorar / Contacto / Síguenos** (así están en la maqueta y en §7).
3. ~~**Selector de idioma del pie**: se conserva (F3 lo tenía) en la franja inferior. Confirmar si además debe mostrarse dentro del panel del menú móvil (hoy en la maqueta solo queda en la barra).~~ ✅ **Resuelto por el cliente** (cierre de puerta 2, 2026-10-10): el selector de idioma **NO se repite dentro del panel del menú móvil**; vive solo en la **cabecera** (compacto) y en el **pie** (así está en la maqueta).
4. **Rendimiento de blobs+blur**: en dispositivos muy modestos los gradientes animados y el `backdrop-filter` son el mayor costo de B; si degradan la fluidez se moderan o se reducen antes que entorpecer la lectura (FR-007). Se resuelve en implementación.
5. **Orden de los grupos del pie en móvil**: confirmar que se apilan en este orden: marca → Explorar → Contacto → Síguenos → franja de idioma (así está en la maqueta).
