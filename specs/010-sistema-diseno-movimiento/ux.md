# UX: F10 — Sistema de diseño y movimiento (sitio público + modo oscuro) — **Ronda 2**

> **Ronda 2 de la puerta 2 (2026-10-10):** el cliente rechazó las 3 propuestas de la ronda 1 («Raíz Viva», «Cielo Abierto», «Luz Serena») por resultarle **«muy pobres, casi imperceptibles»** en animación, y pidió explícitamente transición entre componentes y **revelado de las secciones al hacer scroll hacia abajo**. La spec se actualizó: FR-005 exige nivel **marcado/elegante con revelado al hacer scroll**; FR-006/FR-009 siguen siendo incondicionales (`prefers-reduced-motion` suprime TODO; el contenido nunca depende de la animación). Estas 3 propuestas reemplazan por completo a las anteriores; nada de «Raíz Viva» / «Cielo Abierto» / «Luz Serena» sobrevive.
>
> Fuentes: `spec.md` de F10 (US1–US6, esp. US3 y FR-005/006/007/009/026), tokens de F3, Manual de Identidad (`resources/MANUAL DE MARCA.pdf` manda), referencias de inspiración del cliente (apple.com, microsoft.com/es-cr, river-arena.org, iglesiaoasis.com — **inspiración, nunca copia**).
> Sin panel (F10B fuera). No toca `estado.md` ni el roadmap; no implementa la app.

---

## 0. Base compartida (heredada, no reinventada)

- **Paleta 70/20/10**: `#1a2b4a` navy (70 %), `#F5F2EC` crema/blanco (20 %), `#00c9a7` turquesa + `#ff6b3d` coral + `#217638` verde (10 %). Derivados solo para contraste AA. Turquesa nunca como texto pequeño sobre claro → usar `#0e7a66`; sobre oscuro turquesa puro es texto/acento directo (≈ 8:1).
- **Tipografías**: Bebas Neue (titulares y usos de impacto), Poppins (cuerpo 18 px, botones), Playfair Display itálica (lema, frases emocionales). Fallbacks de sistema en las maquetas; en producción self-hosted (`@fontsource`).
- **Logotipo**: sin deformar, sin recolorear, sin girar, sin efectos — en TODOS los estados, incluido hover/foco. En tema oscuro se apoya sobre pastilla/superficie crema o blanca; nunca se invierte.
- **Compartido por las 3 propuestas (las rechazadas no lo lograron; aquí ya por defecto)**:
  - **Navegación: barra fija** con transición de cabecera al hacer scroll (la barra se compacta, gana fondo/elevación/sombra: se *ve* la transición) — sin hamburguesa, sin menús ocultos, sin scroll lateral de enlaces (wrap a 2 filas si hace falta); apta para F4–F9 (~8–10 enlaces).
  - **Revelado al hacer scroll de nivel marcado**: `IntersectionObserver` + CSS; desplazamientos de entrada grandes (28–48 px) o barrido/máscara, duración 0,55–0,8 s con curva de salida expansiva (`cubic-bezier(.16,1,.3,1)`), **stagger real de 80–140 ms** entre componentes de una misma sección. Cada sección reancla; el foco de teclado fuerza visibilidad inmediata; una vez revelado no se repite.
  - **Desplazamiento suave a anclas** (`scroll-behavior: smooth` + `scroll-margin-top` para que la barra fija jamás tape el título del destino).
  - **Microinteracciones perceptibles**: hover de tarjetas con elevación y brillo/sheen; botones que escalan ≈ 1,04 al pulsarse; enlaces con indicador que se despliega; todo sin desplazar el maquetado, todo en `transform`/`opacity` (compositor, fluido en dispositivos modestos).
  - **Selector de idioma compacto**: «Español · English» en 0,82–0,85 rem, texto completo, sin píldoras ni menú; activo en negrita con subrayado turquesa; comportamiento de F3 intacto (memoria entre visitas, primera visita en español, fallback es→en).
  - **Interruptor de tema + modo oscuro**: botón circular 44 px junto al idioma, `aria-pressed` + `aria-label` dinámico, apli­cado sin recargar, elección recordada (localStorage), primera visita según `prefers-color-scheme`. Paletas oscuras por propuesta, contrastes AA, logo intacto.
  - **`prefers-reduced-motion` (incondicional)**: en las maquetas y en el sistema, con esa preferencia NO hay revelado, ni desplazamiento suave, ni entradas: el contenido aparece completo de inmediato y todo funciona igual. Implementado con media query CSS que anula `transition/animation` de entrada y un guarda en el JS del observador.

**Secciones de portada (orden, igual en las 3, contenido de F3 intacto)**: cabecera fija → hero/identidad → quiénes somos → horario de servicios → WhatsApp → contacto → redes sociales → pie.

---

## 1. Propuesta A — «Bento Vivo»

**Diferencia clave en una línea**: la portada como una **cuadrícula bento** de azulejos de colores de la marca que se descargan al hacer scroll con stagger — la más juguetona, estructuralmente novedosa y la referencia mais directa a Microsoft.

- **Dirección visual**: movimiento a base de panel ordenado. La página se apoya en crema `#F5F2EC`; los bloques de información viven en tarjetas de distinto tamaño y color (navy, turquesa, naranja, verde, crema-blanca) como una pizarra moderna. Tipografía Bebas gigante como hilo conductor, con las palabras-clave en Playfair itálico turquesa.
- **Paleta y uso de marca**: 70 % crema/blanco de fondo, navy para bloques de peso y texto; 20 % tarjetas navy/blancas cruzando; 10 % turquesa/coral/verde repartidos por azulejos de detalle (icono, hora, punto de acento). Logo intacto en cabecera, siempre legible.
- **Composición de portada**: hero = 1 tarjeta-navy grande (texto + CTA) + tarjeta de imagen + azulejos decorativos con iconografía nueva (brotes, campana, reloj — todo SVG propio, `aria-hidden`); luego cada sección es una fila o mosaico bento (horario como tarjeta-reloj con la hora en Bebas; WhatsApp como tarjeta verde con CTA; contacto como 3 azulejos gemelos; redes como fila de píldoras).
- **Navegación en barra fija**: barra crema translúcida de arranque, al bajar se oscurece a navy sólido con sombra y reduce su altura (transición perceptible de color+altura+elevación). Enlaces con indicador inferior turquesa que se crece de 0 a 100 % en hover/foco. Móvil: línea 1 marca + idioma + tema, línea 2 enlaces wrap — nada lateral.
- **Selector de idioma**: `Español · English` muy compacto, unidos por un punto medio fino; el activo subrayado turquesa, 0,85 rem.
- **Modo oscuro**: los azulejos se reparten sobre `#0d1626` (azulejos navy-más-claros `#1d2c4c`, tarjetas turquesa mantienen texto navy), texto claro; la tarjetas crema se mantienen (logo y legibilidad); la barra pasa de oscuro-suave a sólido, pero mantiene el depth de `card shadows`. Turquesa pasa a directo; coral/verde como breaks no-texto.
- **Sistema de movimiento (lo más marcado de las 3 en sensación «descarga»)**:
  1. **Sección al entrar en pantalla**: cada bloque mosaico sube desde 48 px con opacity y **scale 0,94→1**, con stagger de 90 ms entre azulejos de la misma fila (hasta ~0,8 s por azulejo).
  2. **Azulejos decorativos**: rotación leve base que queda tras el reveal (no animación continua).
  3. **Hover/foco**: elevación 6–8 px + borde turquesa; sheen de luz (pseudo elemento) barriendo la tarjeta en hover; CTA escala 1,04 al pulsarse.
  4. **Barra fija**: color de fondo, altura y sombra transicionan al cruzar el hero (clase `scrolled`).
  5. **`prefers-reduced-motion`**: sin reveal, sin sheen, sin desplazamiento suave (todo con HTML estático de inmediato).
- **Wireframes** (`≤`/`>` = dirección de entrada de la animación; `sX` = segunda entrada staggered):

```
[logo SIMIENTE SANTA | Quiénes somos·Horario·WhatsApp·Contacto·Redes | Español · English | 🌙] ← fija (crema→navy al bajar)
═════════════════════════════════════════════════════════════════
┌───────────── BENTO HERO ────────────────┐  ┌── imagen ──┐
│ TÚ CASA EN CRISTO, TU RAÍZ VIVA         │  │            │
│ «Brotamos para dar vida.»               │  │            │┐
│ [ Nuestro horario ][ WhatsApp ]         │  └────────────┘
├── azulejos decorativos (rotación leve) ─┴──┘
── 01 QUIÉNES SOMOS ── (2 azulejos gemelos, hover sube 6px)
── 02 HORARIO ── (tarjeta reloj: 10:00 a. m. Bengal. grande turquesa)
── 03 WHATSAPP ── (tarjeta verde, CTA con tap-response)
── 04 CONTACTO ── (3 azulejos: dirección / correo / teléfono)
── 05 REDES ── (tirillas horizontales + hover sheen)
── PIE navy: logo + lema
═════════════════════════════════════════════════════════════════
Móvil: bento pasa a columna; mismo orden; 320 px sin desbordes.
```

---

## 2. Propuesta B — «Marea Turquesa»

**Diferencia clave en una línea**: luz y **gradientes turquesa** a base de banda sobre navy profundo con paneles glass que entran alternando desde los lados — cinematografía suave (Apple/oasis vibes) con movimiento «le fluye».

- **Dirección visual**: el sitio respira como una marea cálida: fondo navy (70 %) tocado por **blobs de gradiente turquesa** suaves que se mueven de forma perceptible y elegante y paneles de vidrio (`backdrop-filter`) con contenido. Hero no centrado de todo (diferencia clave con la propuesta rechazada «Cielo Abierto»): texto a la izquierda, imagen con marco de vidrio a la derecha, y blobs animados detrás.
- **Paleta y uso de marca**: 70 % navy `#1a2b4a` a sangre (hereda solidez del Manual a sin hacerla «oscura»); 20 % paneles/vidrio del propio navy más o menos claro; 10 % turquesa dominante (gradiente, subrayados, horas, foco) con coral solo un punto; el verde lo aporta el CTA de WhatsApp. Logo intacto.
- **Composición de portada**: hero con blob gradiente Turquesa→navy tras el título; secciones como paneles de vidrio redondeados sobre el fondo navy con blobs por detrás, entrando **alternando desde la izquierda y la derecha** (perceptible en ese eje). Horario con dígitos en Bebas turquesa de 4 rem; WhatsApp con btn verde; contacto como lista de pastillas; redes como ramos de baldosas.
- **Navegación en barra fija**: al bajar la barra se vuelve banda de vidrio (fondo `rgba(13,22,38,.72)` + blur + borde inferior turquesa de 1 px que aparece con el scroll); enlaces marcan con subrayado turquesa animado; en móvil 2 filas wrap, controllable a 320 px.
- **Selector de idioma**: pastilla muy fina con borde translúcido: `Español / English`, activo con fondo turquesa suave (la pastilla es la contención, no un menú).
- **Modo oscuro**: el tema «claro» de esta propuesta ya es el navy `#1a2b4a`; su oscuro profundiza a `#0b1424` con paneles `#101b30` y el mismo lenguaje de blobs, solo más sutil; turquesa como acento directo (AA sobre oscuro); coral solo como detalle gráfico. El logotipo se apoya siempre sobre su pastilla crema.
- **Sistema de movimiento (el más cinematográfico)**:
  1. **Revelado de secciones**: paneles entran desde los lados (± 56 px) + opacity, con **ease expansivo 0,75 s** y stagger 110 ms entre sus componentes internos (título → texto → CTA).
  2. **Blobs gradiente**: deriva lenta y continua (solo si movimiento permitido; a 60 fps no bloquea, `transform only`).
  3. **Barra fija**: cambio de fondo a vidrio + blur + borde inferior turquesa al hacer scroll (transición perceptible).
  4. **Hover/foco**: paneles ganan luminosidad del borde; botones elevan y el gradiente de fondo se intensifica; enlaces de red con swipe flecha.
  5. **`prefers-reduced-motion`**: sin blobs animados, sin reveal, sin desplazamiento suave: páginas estáticas, mismo contenido.
- **Wireframes**:

```
[🌱 SIMIENTE SANTA | Quiénes somos · Horario · WhatsApp · Contacto · Redes | Español / English | 🌙] ← fija
════════════════════════════════════════════════════════════════
  (navy a sangre, blob Turquesa → navy detrás del titular)
  BIENVENIDO A TU           ┌─ panel de vidrio ─┐
  CASA EN CRISTO  (Bebas)   │   imagen portada   │
  «Brotamos para dar vida.» └────────────────────┘
  [ Ver horario ]  [ Escribir por WhatsApp ]
── (paneles de vidrio sobre el manto navy, alternando ← y →) ──
  ← 01 Quiénes somos   |   02 Horario (10:00 turquesa grande) →
  ← 03 WhatsApp  |  04 Contacto  |  05 Redes ─ ramos
── PIE vidrio con logo ──
════════════════════════════════════════════════════════════════
Móvil: hero pasa columna; paneles entran desde el lado único arriba;
nada lateral; 320 px sin desbordes.
```

---

## 3. Propuesta C — «Huella Tipográfica»

**Diferencia clave en una línea**: **la tipografía ES la animación** — titulares Bebas gigantes que se despliegan línea a línea con máscara al hacer scroll, bandas cremanas y navy que se alternan con barrido, la más editorial-kinética (Apple }}

- **Dirección visual**: revolución tipográfica (revista 2026). Mucho aire, tipografía a escala «póster»; cada titular se despliega con máscara (línea a línea) al llegar al scroll, como si la página se imprimiera. Una base crema con **bandas navy a sangre completa para alternar el ritmo**
- **Paleta y uso de marca**: 70 % crema/blanco con mucho aire; 20 % bandas navy completas (una sección y el pie); 10 % turquesa en subrayados que se despliegan y punto rojo coral. El color vive DENTRO de la tipografía (palabra en turquesa/navy en medio de un titular), no en cajas.
- **Composición de portada**: hero tipográfico apilado en 3 líneas Bebas de hasta 5 rem con la 2ª línea en Playfair itálica turquesa («TÚ CASA / *en Cristo,* / TU RAÍZ VIVA»), imagen en rectángulo grande; cada sección abre con numeración gigante (01–05) en outline que se rellena al entrar; horario como tabla de Bebas; contacto en filas; redes en fila de enlaces-texto; pie banda navy con logo.
- **Navegación en barra fija**: la más ligera: la barra crema se compacta al bajar (menor altura, fondo blanco con hairline y sombra) y la marca se reduce con una transición suave; el subrayado turquesa de los enlaces se despliega de izquierda a derecha al hover/foco.
- **Selector de idioma**: `Español / English` en 0,82 rem; activo negrita y subrayado turquesa desplegado.
- **Modo oscuro**: «tinta que se vuelce papel»: fondo `#101a2c`, texto crema; las bandas navy se vuelven bandas crema (invertida elegancia) con texto navy dentro; turquesa puro disponible como acento; el logo sigue con sus colores (sobre pasilla crema/panel navy). La barra también cambia.
- **Sistema de movimiento (el más «poblado por la marca»; tipografía-kinética)**:
  1. **Titulares**: despliegue línea a línea dentro de máscara (`overflow:hidden` + `translateY(110%)→0`), 0,6 s por línea, **stagger 120 ms entre líneas** — el efecto Apple-keynote que la iteración pide.
  2. **Secciones**: bloque entero con desplazamiento de 32 px + opacity 0,7 s; los **números de sección (01–05)** llegan primero, con 120 ms de adelanto respecto al resto de la sección (se percibe que la numeración abre el paso).
  3. **Bandas navy**: barrido horizontal (scaleX 0→1 en 0,5 s) cuando la banda entra y el contenido revela después.
  4. **Hover/foco**: subrayados que se despliegan; filas de horario/contacto se tiñen turquesa 6–8 %; botones que se hunden levemente al activar (scale 0,97).
  5. **Barra fija**: shrink de altura + sombra + estilo compacto con el scroll.
  6. **`prefers-reduced-motion`**: titulares ya visibles completos; sin despliegues ni barridos ni suave; páginas sin anim, mismas funciones.
- **Wireframes**:

```
[logo SIMIENTE SANTA | Quiénes somos·Horario·WhatsApp·Contacto·Redes | Español / English | 🌙] ← fija (se compacta al bajar)
════════════════════════════════════════════════════════════════
  TÚ CASA          ← línea 1 (máscara despliega)
  en Cristo,       ← línea 2 Playfair itálica turquesa
  TU RAÍZ VIVA     ← línea 3
  [imagen grande del hero]
── 01 QUIÉNES SOMOS: texto editorial con subrayados desplegados
── banda navy ▨ a sangre: 02 HORARIO (filas Bebas
      DOMINGO       10:00 a. m. — Culto
      MIÉRCOLES      7:00 p. m. — Estudio)
── 03 WHATSAPP (botones fila)
── 04 CONTACTO / 05 REDES (filas + flechas)
── PIE banda navy con logo + lema Playfair
════════════════════════════════════════════════════════════════
Móvil: tipografía escala con clamp(); máscaras iguales; 320 px ok.
```

---

## 4. Tabla comparativa (ronda 2)

| Criterio | A — «Bento Vivo» | B — «Marea Turquesa» | C — «Huella Tipográfica» |
|---|---|---|---|
| Dirección visual | Cuadrícula bento de azulejos de marca | Gradientes turquesa y vidrio sobre navy | Editorial kinético, tipografía a escala póster |
| Sensación del movimiento | «Descarga ordenada»: los azulejos caen en fila | «Fluye»: paneles llegan alternando desde los lados | «Prensa viva»: las palabras se despliegan con máscara |
| Movimiento más llamativo | Reveal con scale + stagger por azulejos; sheen hover | Blobs de gradiente continuos + entrada lateral alterna | Despliegue titular línea a línea + barrido de bandas |
| Protagonismo del color | Crema con mosaico de marca (70/20/10 visible) | Navy dominante con turquesa como luz | Crema con bandas navy rítmicas; color dentro del texto |
| Primera impresión | Moderna, viva, contemporánea | Cinematográfica, elegante e inmersiva | Sorprendente, editorial, «no típica de IA» |
| Riesgo de «plantilla IA» | Bajo (el bento es tendencia muy conocida; la se diferencia con la paleta, la iconografía propia y las ilustraciones de la marca) | Medio (gradiente + vidrio es un terreno visitado; la marea turquesa en navy de marca lo separa) | Muy bajo (tipografía kinética casi no se ve en webs de iglesia) |
| Legibilidad para todas las edades | Muy alta (todo en tarjetas con cuerpo Poppins 18 px) | Alta (mucho texto claro sobre navy; contraste verificado) | Muy alta (contraste máximo; sin cajas; tamaños grandes) |
| Mejor en modo oscuro | Muy buena (mosaico sobre azul noche) | Muy buena (ya vive en la temperatura; el oscuro solo profundiza) | Muy buena (tinta↔papel, cambio notable) |
| Escalabilidad F4–F9 (navegación) | Muy buena (2 filas wrap) | Buena (pastilla de enlaces wrap) | Muy buena (la nav más mínima) |
| Riesgo con `prefers-reduced-motion` | Bajo (reveal simple, fácil de anular) | Medio (blobs continuos que hay que silenciar del todo) | Bajo (despliegues lineales, fáciles de anular) |
| Peso/rendimiento | Bajo (reveals + sheen; `transform`/`opacity` puros) | Medio (gradientes animados + blur: moderar en dispositivos modestos) | Bajo (máscaras y scaleX; sin blur salvo nada) |
| Esfuerzo de implementación | Medio | Medio-alto | Medio |

## 5. Recomendación

**Recomiendo A — «Bento Vivo»**, con **C — «Huella Tipográfica» como primera alternativa**.

- El pedido duro del cliente era «más animación, que se sienta»: las 3 lo entregan, pero **A es la que mejor equilibra animación marcada + acogida para todas las edades** — los azulejos con stagger de 90 ms y revelado grande se sienten de inmediato, y el patrón bento escala de forma natural a F4–F9 (cada sección futura entra como un nuevo conjunto de azulejos sin inventar estilo). El riesgo de «plantilla IA» se neutraliza con la paleta de la marca, la ilustración propia y la palabra-clave en Playfair turquesa.
- **C** es la más innovadora y probablemente la que más sorprende al cliente, pero su lenguaje editorial-kinético exige textos de bienvenida cuidados y quizá más afinación por sección; es la apuesta fuerte si el cliente quiere «nada típico» por encima de todo.
- **B** es la más impresionante visualmente pero con más costo de rendimiento (blur y gradientes animados) y más riesgo de marear a personas mayores respecto a `prefers-reduced-motion` e inteligibilidad con todo ese fondo.

**Propuesta aprobada por el humano**: ☐ A «Bento Vivo» ☐ B «Marea Turquesa» ☐ C «Huella Tipográfica» — (marcar antes de implementar; FR-026 / SC-013).

## 6. Componentes React (para dev-frontend)

### 6.1 Reutilizados de `frontend/src/components/`
- `Button` — añadir variantes con feedback táctil (`active:scale`), nunca sobre el logo.

### 6.2 De `frontend/src/features/publico/` (reutilizar/estender, comportamiento intacto)
| Componente | Extensión |
|---|---|
| `PublicLayout` | Header fija con transición de scroll (compactado/vidrio/sólido), slots para idioma+tema, pie según propuesta. |
| `LanguageSwitcher` | Formato compacto (0,82–0,85 rem, texto completo) según propuesta; comportamiento F3 intacto. |
| `HomePage`, `IdentityHero` (hoy `Hero`/`WhoWeAreSection`), `ScheduleSection`, `WhatsAppSection`, `ContactSection`, `SocialSection`, `PublicFooter`, `SectionSkeleton`, `SocialIcon` | Maquetado nuevo según la propuesta aprobada; flujos de datos, fallback y publicación intactos. Cada bloque se envuelve en `Reveal` con stagger. |

### 6.3 Nuevos (cimientos para F4–F9)
| Componente / hook | Responsabilidad |
|---|---|
| `Reveal` + hook `useReveal` | Envoltorio de revelado: `IntersectionObserver` con umbral 0,15, variante de entrada (`up` / `left` / `right` / `lines` / `scale`), delay de stagger, DISPARA UNA VEZ; al recibir foco se hace visible de inmediato (regla de edge case); con `prefers-reduced-motion` no observa y muestra de inmediato. |
| `TransitionHeader` | Barra fija con clase `scrolled` (sombra + fondo + altura) al cruzar el umbral del scroll. |
| `ThemeToggle` + `useTheme` + `ThemeScript` | Interruptor 44 px con `aria-pressed`/`aria-label` dinámico; persistencia de la elección; script inline pre-render anti-destello; observación en vivo de `prefers-color-scheme` si no hay elección manual. |
| Tokens de movimiento en `frontend/src/index.css` (`@theme`) + variables de tema oscuro | Duraciones (‑‑dur‑reveal: 0,7 s; ‑‑dur‑micro: 0,25 s; ‑‑stagger: 90 ms), curvas; documentadas para F4–F9. |
| `features/publico/messages.ts` | Cadenas nuevas es/en: `theme.toDark`, `theme.toLight`, `theme.state.*`, `a11y.revealNote` si se necesita. |

## 7. Accesibilidad (WCAG 2.1 AA, en ambos temas)

- **Teclado**: skip link → logo → enlaces de nav → selector de idioma → tema → contenido. Sin trampas; el foco de teclado nunca queda sobre algo sin revelar (los componentes `Reveal` se visibilizan si reciben foco antes de su revelado).
- **Foco visible**: anillo 2 px turquesa (o alternante según fondo) sobre claro y sobre oscuro, incluido el estado `scrolled` de la barra.
- **Contraste AA**: Revisión (script) prevista de las combinaciones del sistema en ambos temas: texto ≥ 4,5:1, grande/avalo ≥ 3:1, incluidos hover/foco/activo y fuentes “derived” (turquesa pequeño sobre claro solo con #0e7a66).
- **Objetivos ≥ 44 px**: botón tema, enlaces de nav, selector de idioma (opciones de etiqueta completadas con padding), filas de redes u contactos, CTAs.
- **Orden y semántica**: `<nav aria-label="Secciones del sitio">`, grupo de idioma con `role="group" aria-label="Idioma"`, `aria-pressed` en tema; `lang` en `<html>` cambia es/en como en F3.
- **Movimiento**: nunca bloquea lectura ni acción (FC-007); mover solo `transform`/`opacity`; el scroll nunca es condición para ver contenido (con lector de pantalla todo se anuncia igual); con `prefers-reduced-motion` cero revelados, cero suave, cero blobs — check igual en la maqueta (usar emulador DevTools o forzando la media query).
- **320 px y zoom 200 %**: la barra de las 3 propuestas conserva un tamaño objetivo, sin overflows ni superpositions.

## 8. Textos de interfaz nuevos (es/en, FR-019)

| Clave | Español | English |
|---|---|---|
| `theme.label` | Tema | Theme |
| `theme.toDark` | Cambiar a tema oscuro | Switch to dark theme |
| `theme.toLight` | Cambiar a tema claro | Switch to light theme |
| `lang.groupLabel` | Idioma | Language |
| `nav.groupLabel` | Secciones del sitio | Site sections |
| `skipToContent` | Ir al contenido | Skip to content |

El resto de rótulos ya existe en F3 y no cambia de texto ni de clave; las maquetas contienen demo bilingüe interactivo (clic en «Español/English») solo para que el cliente lo sienta: en el sitio real esa lógica sigue siendo la de F3.
