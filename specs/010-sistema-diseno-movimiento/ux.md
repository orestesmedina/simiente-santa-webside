# UX: F10 — Sistema de diseño y movimiento (sitio público + modo oscuro)

> Documento de diseño para `disenador-ux`. Fuentes: `spec.md` de F10 (aprobada 2026-10-10, decisiones Q1–Q8), §0/§1/§9 de `specs/003-portada-info-general/ux.md` (tokens de F3) y el Manual de Identidad (logotipo con sus reglas; si algo del manual está resumido aquí, manda el PDF original: `resources/MANUAL DE MARCA.pdf`).
> **Entregable central (FR-026)**: 3 propuestas visuales distintas e innovadoras, con maquetas estáticas autocontenidas en `propuestas/`, para que el humano apruebe UNA antes de implementar.
> Sin panel de administración (F10B, fuera). No toca estado.md ni el roadmap; no implementa la app.

---

## 0. Base compartida (heredada, no reinventada)

Tomada de F3 y de la spec de F10 — las 3 propuestas se construyen sobre esto:

- **Paleta y proporción 70/20/10**: `#1a2b4a` azul marino (70 %), `#F5F2EC` crema / blanco (20 %), `#00c9a7` turquesa + `#ff6b3d` naranja + `#217638` verde (10 %). Colores derivados solo para contraste (AA). Reglas de contraste de F3 §1.1 siguen aplicando: turquesa `#00c9a7` nunca como texto pequeño sobre claro (usar `#0e7a66`); lectura principal siempre navy sobre claro o blanco/crema sobre navy.
- **Tipografías**: Bebas Neue (títulos/H1/H2), Poppins (cuerpo 18 px, botones), Playfair Display itálica (lema, frases emocionales). Fallbacks de sistema en maquetas; en producción self-hosted (`@fontsource`), como recomendó F3.
- **Reglas del logotipo**: nunca deformar, recolorear, girar ni aplicar efectos; siempre `object-contain` con dimensiones explícitas; el `alt` lo gestiona el equipo en el panel.
- **Restricciones duras**: sin menús ocultos ni gestos desconocidos; cuerpo ≥ 18 px; objetivos táctiles ≥ 44 px; cero desplazamiento horizontal desde 320 px; `prefers-reduced-motion` incondicional; comportamiento funcional de F3 intacto (solo presentación/navegación/selector/tema).
- **Se permiten iconografía e ilustraciones nuevas** (Q6) respetando paleta, tipografías y reglas del manual.
- Coordina con lo resuelto: Q2/Q3 barra fija siempre visible (elimina scroll lateral de enlaces y la pérdida de navegación al bajar); Q4 movimiento moderado; Q5 selector compactado (textos completos «Español/English», sin píldoras, sin menú); Q7 modo oscuro con preferencia del sistema por defecto + interruptor manual recordado; Q8 3 propuestas.

**Movimiento: patrón único compartido por las 3 propuestas (nivel moderado, FR-005–FR-009)**

| Interacción | Comportamiento |
|---|---|
| Hover / foco sobre enlaces, tarjetas y botones | Transición corta (150–250 ms) de color/borde/elevación leve; nunca desplaza el maquetado |
| Revelado al entrar en pantalla | Fade + subida breve (~10–14 px, ≤ 0,5 s), una sola vez por sesión de página, vía `IntersectionObserver`; el foco de teclado fuerza visibilidad inmediata (edge case de la spec) |
| Cambio de idioma / tema | Sin animación de reentrada; solo el cambio instantáneo de valores |
| `prefers-reduced-motion` | Todo desactivado: contenido completo visible de inmediato, 100 % de funciones (CSS: `transition/animation: none`) |
| Dispositivo modesto | Animaciones solo en `transform`/`opacity` (compositor), sin efectos sobre scroll continuo ni parallax |

**Modo oscuro: patrón de control (Q7, compartido)**

- Control compacto en la cabecera, junto al selector de idioma: botón circular de 44 px con icono (🌙/☀ según estado), `aria-label` que anuncia la acción («Cambiar a tema oscuro» / «Cambiar a tema claro») y `aria-pressed` del estado. La primera visita sigue la preferencia del sistema (`prefers-color-scheme`); la elección manual se recuerda (mismo mecanismo de persistencia que el idioma) y manda sobre el sistema hasta que se cambie. El cambio se aplica sin recargar y sin animación de reentrada. Cadenas es/en: `theme.label` «Tema / Theme», `theme.toDark` «Cambiar a tema oscuro / Switch to dark theme», `theme.toLight` «Cambiar a tema claro / Switch to light theme».
- **Paleta en tema oscuro (derivada por contraste, logotipo intacto)**: superficies `#0d1626` / `#16233c` (azul noche, deriva del navy), texto `#edf0f6` / blanco (≈ 13:1 sobre `#0d1626`), acento turquesa `#00c9a7` permitido como texto/acento sobre oscuro (≈ 8:1), crema reservado a superficies puntuales, verde y naranja solo como relleno gráfico puntual con texto navy/oscuro encima, bordes translúcidos claros `rgba(238,241,246,.14)` y foco en turquesa con anillo 2 px. El logo conserva SIEMPRE sus colores originales: se apoya en una pastilla/superficie crema o blanca donde el fondo sea oscuro (nunca se invierte).

---

## 1. Propuesta A — «Raíz Viva»

**Dirección visual:** un editorial orgánico y cálido. Sobre lienzo crema/blanco, titulares Bebas de gran tamaño con una palabra destacada en turquesa dentro del mismo titular, tarjetas 2xl con esquinas acogedoras y una ilustración nueva de brote en turquesa/naranja. La marca mira hacia su nombre: la simiente que germina; el tono es cercano y luminoso.

- **Paleta y tipografías**: 70 % crema/blanco como lienzo, navy en texto y superficies, turquesa en la palabra-clave del H1, en numeración de sección y foco; coral solo en un detalle mínimo (punto de acento de la ilustración). Bebas para H1/H2 (con *énfasis* de palabras coloreadas — recurso distintivo), Poppins cuerpo, Playfair el lema y las citas.
- **Composición de la portada**: hero con 2 columnas (texto + imagen de portada con ilustración sobrepuesta, nunca sobre el logo), misión/visión como 2 bloques gemelos en tarjetas; luego secciones numeradas 01–05 con cabecera de regla fina y número turquesa; tarjetas horizontales de 2 columnas con la hora como elemento visual en Bebas; contacto en una gran tarjeta redondeada 3 columnas; redes como píldoras; pie navy-redondeado con frase Playfair.
- **Barra fija**: fondo del propio lienzo con blur al bajar (nunca tapa con color opaco en exceso), logo + nombre a la izquierda, enlaces **con wrap a 2 filas compactas si hace falta** (nunca scroll lateral), idioma «Español · English» compacto y botón circular de tema a la derecha. En móvil: 2 filas (marca+controles / enlaces). Escala a F4–F9 porque el wrap convive con 8–10 enlaces sin romper.
- **Selector compacto**: `Español · English` en 0,85 rem, el activo subrayado turquesa + negrita; literal completo accesible en ambos temas; en el pie igual.
- **Modo oscuro**: el navy se profundiza a azul noche `#0d1626` con superficies `#16233c`; el navy del texto pasa a claro; el turquesa `#00c9a7` pasa a ser texto/acento directo (AA sobre oscuro). El logo conserva sus colores originales — sobre superficies oscuras se apoya en su propia pastilla crema para seguir legible, sin recolorear.
- **Movimiento**: tarjetas se elevan ~4 px en hover con borde turquesa; reveal de fade+subida breve; botón CTA escala 1,03. Nada de parallax ni secuencias.
- **Wireframe**:

```
[≡ logo SIMIENTE SANTA | Quiénes somos·Horario·WhatsApp·Contacto·Redes | Idioma E·🌙]  ← fija
─────────────────────────────────────────────────────────────
KICKER ─────────────────                                  ┌────────────┐
TÚ CASA EN CRISTO,                                        │  IMAGEN    │
TU RAÍZ VIVA ←(palabra turquesa)                          │ (portada)  │
«Brotamos para dar vida.»                                 │   +brote   │
[ Nuestra misión ][ Nuestra visión ]                      └────────────┘
[ ESCRIBIR POR WHATSAPP ]                                 01–05 numeradas
─────────────────────────────────────────────────────────────
Móvil (320 px): 1 columna; mismo orden; enlaces wrap en 2 filas.
```

## 2. Propuesta B — «Cielo Abierto»

**Dirección visual:** navy como «cielo» siempre presente (70 % literal: fondo de página), con capas flotantes ligeramente más claras como paneles que contienen el contenido; hero a sangre completa centrado con logo, tipografía masiva y un haz sutil de turquesa (radial suave, nunca sobre el logo); píldoras/estrellas mínimas de ilustración nueva. Es la más inmersiva y «hecha» de las 3 — la experiencia más de primera marca.

- **Paleta y tipografías**: fondo navy dominante; texto blanco; capas `#22355c` (derivado); turquesa como luz: hora de servicios en Bebas turquesa, numeración, encabezados de datos y foco; coral solo un punto junto al CTA secundario. Bebas masiva en hero (portrait de 6,5 rem), Poppins cuerpo, Playfair el subtítulo del hero y la frase del pie.
- **Composición de la portada**: hero a pantalla completa (min 72 vh) todo centrado: logo, titular masivo «BIENVENIDO A TU CASA», subtítulo Playfair, 2 CTAs. Después cada sección es una «capa flotante»: panel redondeado de navy-más-claro con su contenido (números ◆ 01–05). Horario con hora en Bebas grande. Redes como píldoras. Pie flotante en una línea simple.
- **Barra fija**: banda delgada que empieza opaca y se vuelve translúcida (blur) al bajar; enlaces redondos con fondo `panel` en hover e idioma dentro de una pastilla con borde; botón de tema circular con borde fino. Móvil: mismo layout compacto, wrap natural; apta para 320 px porque usa tamaños pequeños sin sacrificar 44 px de target.
- **Selector compacto**: pastilla con borde `🌐 Español / English` (el emoji es decorativo + aria-label), texto completo en ambas opciones y el activo en turquesa con fondo. Sin píldora «ES/EN» como identificativo único; el 🔭 emoji es decoración (con `aria-label` «Idioma»; el grupo lo anuncia).
- **Modo oscuro**: se va a un azul noche más hondo `#0d1626` con capas `#131f36`; la interfaz ya era clara-sobre-marca en claro, así que el cambio es sutil y elegante. Las estrellas de ilustración solo titilan si el movimiento está permitido; nunca ocultan información. El logo conserva sus colores originales sobre su pastilla crema.
- **Movimiento**: transición del fondo de la barra al bajar; hover de tarjetas/botones con elevación breve; píldoras de estrella con parpadeo suave (solo con movimiento permitido). El foco nunca queda oculto tras un efecto.
- **Wireframe**:

```
[🌱 logo | Quiénes somos·Horario·WhatsApp·Contacto·Redes | 🌐 Español·English | 🌙] ← fija, delgada
══════════════════════════════════════════════════════════════════
        (cielo navy a sangre completa)
              ◦      ·        °   (estrellas ilustrativas)
                  [ LOGO ]
            BIENVENIDO A TU CASA
         «Esta siempre será tu casa.»
        [ Ver horario ]  [ Escribir por WhatsApp ]
══════════════════════════════════════════════════════════════════
   ┌ capa: Quiénes somos ────────┐  ┌ capa: Horario ───────────┐
   │ ◆ 01  texto                 │  │ ◆ 02  10:00 a. m. (Bebas)│
   └─────────────────────────────┘  └──────────────────────────┘
   (así cada sección, luego Redes como píldoras y pie simple)
─────────────────────────────────────────────────────────────
Móvil: hero centrado; capas columna; cabecera wraps.
```

## 3. Propuesta C — «Luz Serena»

**Dirección visual:** la contraria: un folio blanco-limpio con tinta navy, tipo editorial-suizo de gran aire: casi sin tarjetas, listas de líneas finas (hairlines), titulares Playfair itálica + Poppins mayúsculas-espaciadas, y toques de turquesa en las líneas de enfasis. Es la más «limpia, contemporánea y editorial» — la que parece más cara y más seria, apta para todas las edades por ser de lectura muy clara.

- **Paleta y tipografías**: 70 % blanco, 20 % crema en fichas puntuales, 10 % turquesa (subrayados, numeración, hover); navy como tinta del texto y como bloque sólido del PIE (único momento 100 % de marca). Tipografías: Poppins mayúsculas con tracking amplio para rótulos de sección, Playfair Display regular/itálica para el H1 (más humano y editorial), Bebas para horas y para la marca del pie.
- **Composición de la portada**: hero 2 columnas con imagen en recuadro fino (no a sangre), título «Brotamos para dar vida» en Playfair y «SIMIENTE SANTA» en mayúsculas Poppins-tracked debajo; cada sección con cabecera `numero — título-en-SMALL-CAPS — nota-Po periche itálica a la derecha` (01/02/03…5) y **listas con líneas finas**: horario como filas día→hora, contacto como tabla dirección/correo/teléfono, redes como filas de lista completas «✦ Facebook de la iglesia — Ver en Facebook →». Pie navy sólido con logotipo y frase Playfair.
- **Barra fija**: fina, de una sola línea, fondo blanco con subrayado turquesa en hover y en activo,logo a la izquierda seguido de nombre en tracking; herramientas a la derecha en texto pequeño. Móvil: 2 filas (marca y enlaces; controles en la línea de arriba). La barra fija es lo más leve de las tres.
- **Selector compacto**: la variante la más mínima: `ES· Español / English` — letra pequeña (0,82 rem) con el activo en negrita + subrayado turquesa; «ES·» es solo hint visual decorado, el nombre completo vive en el enlace accesible.
- **Modo oscuro**: papel que se vuelve tinta: fondo `#10192b` (azul-noche cálida), texto crema `#edf0f5`, líneas de las tablas `#2c3a58`; turquesa pasa a ser texto/acento directo (contrase AA sobre oscuro); el bloque del PIE se invierte a tinta turquesa sobre azul-noche — el logo siempre con sus colores originales.
- **Movimiento**: la más contenida: subrayado de líneas que se despliega (escala X) en nav; reveal fade+subida; las líneas del hover de horario se tiñen turquesa con 8 % de fondo. Nada de elevación ni sombras.
- **Wireframe**:

```
[logo SIMIENTE SANTA | Quiénes somos·Horario·WhatsApp·Contacto·Redes | ES· Español/English | ☾]  ← fija fina
─────────────────────────────────────────────────────────────
─────────────
Brotamos para dar vida          │
SIMIENTE SANTA                  │  [FICHAS/IMAGEN]
texto/bienes …                  │
─────────────────────────────────────────────────────────────
01 QUIÉNES SOMOS          casa · fe · vida
  texto editorial con fichita lateral…
02 HORARIO                te esperamos
  DOMINGO    10:00 a. m. — Culto…
  MIÉRCOLES  7:00 p. m.  — Estudio…
03 WHATSAPP   fichas con botones
04 CONTACTO   dirección / correo / teléfono (líneas)
05 REDES      filas ☐ Facebook — Ver en Facebook →
─────────────────────────────────────────────────────────────
PIE: bloque navy sólido con logo, frase y (fila) selector.
```

---

## 4. Comparativa

| Criterio | A — «Raíz Viva» | B — «Cielo Abierto» | C — «Luz Serena» |
|---|---|---|---|
| Dirección | Orgánica, cálida, editorial-orgánica | Inmersiva, masiva, el «cielo» azul | Mínima, editorial, limpio y aérea |
| Protagonismo del color | Crema/blanco con toques turquesa | Navy por todas partes (70 % literal) | Blanco con tinta fina |
| Primera impresión | Acogedora y moderna | Impactante y actual | Serena y editorial |
| Riesgo de «plantilla IA» | Bajo (recuso de palabra-turquesa ilustración) | Bajo-medio (hero centrado es visitado, se diferencia con luces+bordes) | Muy bajo (menos común en webs de iglesia) |
| Legibilidad para todas las edades | Alta (18 px everywhere, tarjetas redondeadas) | Media-alta (mucho texto blanco sobre navy, tamaños grandes; verificar densidad de capas) | Muy alta (contraste máximo, sin cajas) |
| Mejor en modo oscuro | Buena | Muy buena (ya es oscura de esencia, el cambio es sutil) | Muy buena (papel→tinta, cambio elegante) |
| Escalabilidad F4–F9 (navegación) | Buena (wrap de fila compacta) | Buena (píldora redonda de enlaces, compacta) | Muy buena (líneas finas casi no ocupan) |
| Peso/rendimiento | Bajo | Bajo (blur y 1 gradiente de radio, cuidado) | Mínimo (sin blur ni tarjetas) |
| Esfuerzo de implementación | Medio | Medio-alto (capas/gradiente/columnas flor) | Bajo-medio |
| Movimiento característico | Elevación de tarjetas | Parpadeo flotante + transición de barra | Despliegue de subrayado |

## 5. Recomendación

**Recomiendo A — «Raíz Viva».** Es la que mejor equilibra los tres polos del encargo: (1) «moderna 2026» sin ser fría — la ilustración de brote y la palabra-turquesa dentro del titular le dan carácter propio no-genérico; (2) fiel al Manual: usa la proporción 70/20/10 clásica y el logo intacto, con los colores acogedores de la marca como protagonistas; (3) apta para todas las edades y accesible por defecto (mucho contraste, tarjetas generosas, cuerpo claro). B es la más impactante pero pone el foco en la marca antes que en el contenido del visitante (el mucho navy puede restar calidez al «cercano» que pide la voz del manual, y suma un gradiente suave que roza la contención de F3); C es la más elegante y ligera, pero para una iglesia comunitaria corre más riesgo de parecer «distante», y su lenguaje editorial asume lectores revista-más-que-teléfono. Si el humano quisiera la modernidad absoluta casi sin riesgo, la alternativa sería B.

**Propuesta aprobada por el humano**: ☐ A ☐ B ☐ C — (marcar antes de implementar; FR-026/SC-013).

## 6. Componentes React (para `dev-frontend`; plan base, nombres finales los fija el arquitecto)

### 6.1 Reutilizados (de `frontend/src/components/`)

- `Button` — extender con variante `accent` (fondo teal / texto navy) si la propuesta lo pide (A y B la usan; C usa enlaces subrayados y «btn-sol» invertido). NUNCA aplicar efectos sobre el logo.
- `Notice`, `Dialog`, `ConfirmDialog`, `Field`, `Table`, `Tabs`, `StatusPill`, `EmptyState`, `Pagination` — del panel: SIN cambios (el panel no se toca; F10B lo rediseña después).

### 6.2 De `frontend/src/features/publico/` (F3: reutilizar/estender)

| Componente | Extensión |
|---|---|
| `PublicLayout` | Cambia el fondo/estilo de la cabecera según la propuesta; añade slots `Toolbar` (idioma + tema); mantiene el pie. |
| `LanguageSwitcher` | Cambia de radios-pastilla a formato compacto (texto completo 0,82–0,85 rem, subrayado/fondo en el activo). Comportamiento de F3 intacto. |
| `HomePage`, `IdentityHero`, `WhoWeAreSection`, `ScheduleSection`, `WhatsAppSection`, `SocialSection`, `ContactSection`, `PublicFooter` | Cambian el maquetado (clases/estructura visual) según la propuesta elegida; el _data-flow_, contenidos y fallback no cambian. |
| `SectionSkeleton` | Estilo adaptado a la propuesta (líneas blanco/crema/noche según tema). |
| `SocialIcon` | Está bien para la iconografía social (SVG inline propio, sin dependencias). «ⓘ» Q6 permite ilustración nueva: va en `IdentityHero` como SVG decorativo con `aria-hidden`, nunca sobre el logo. |

### 6.3 Nuevos (cimientos del sistema, utilizable por F4–F9)

| Componente / hook | Responsabilidad |
|---|---|
| `ThemeToggle` | Botón circular de 44 px en la cabecera: lee/rota el tema, `aria-pressed` + `aria-label` dinámico; guarda la elección (misma persistencia que el idioma); en primera visita respeta `prefers-color-scheme`. |
| `useTheme` | Contexto/hook de tema: inicializa (elección guardada → sistema), observa `prefers-color-scheme` en vivo si no hay elección manual, conmuta `data-theme` en `<html>` sin recargar. |
| `ThemeScript` | Script inline mínimo que aplica el tema ANTES del primer render para evitar destello (edge case «carga con caché/recarga»). |
| `Reveal` (o hook `useReveal`) | Wrap para el fade+subida de la propuesta: `IntersectionObserver` + atributo `prefers-reduced-motion`; además, al recibir foco el elemento se hace visible de inmediato. |
| `tokens` (`frontend/src/index.css` `@theme`) | Consolidar los design tokens (colores, tipografías, espaciado, radios, breakpoints) y añadir las variables de **tema oscuro** y de **movimiento** (duraciones). Documento para F4–F9. |
| `features/publico/messages.ts` | Añadir cadenas es/en: `theme.*` (ver §0), `nav.*` (si cambia algún rótulo), y la descripción accesible de controles (`a11y.themeLabel`, etc.). |

### 6.4 Cimientos para F4–F9 (FR-018 / SC-011)

Documentar en el `plan.md` (el arquitecto): tokens con ambos temas, patrón de `ThemeToggle`/`Reveal`, el patrón de cabecera fija con slots (navegación + idioma + tema) y el formato del selector — para que cualquier sección nueva se construya con esto sin rediseñar el sistema.

## 7. Accesibilidad (WCAG 2.1 AA) — resumen compartido

- **Teclado**: `Tab` orden lógico — skip link → logo → enlaces de navegación (todos visibles) → selector de idioma (dos enlaces reales) → interruptor de tema → contenido de secciones. Sin trampas de foco; el foco nunca queda sobre un elemento aún no «revelado».
- **Foco visible**: anillo 2 px turquesa sobre claro/oscuro (o color adaptado por contexto); visible en scroll de la barra fija también.
- **Contraste AA en ambos temas**: plan contrast-check por las combinaciones del sistema (F3 ya dejó las reglas; la variante oscura las hereda con las sustituciones del §0); 0 texto por debajo de 4,5:1 normal, 3:1 grande.
- **Objetivos táctiles**: ≥ 44×44 px (botón tema, enlaces de nav, las opciones de idioma y las filas/CTA de tarjetas).
- **Semántica**: `<nav aria-label="Secciones de la portada">`, `role="group" aria-label="Idioma"` para el selector, `aria-pressed` y `aria-label` dinámico en `ThemeToggle`; `html lang` cambian entre es/en igual que en F3.
- **Todo lo no-movimiento-céntrico**: cualquier función funciona sin animación; los «click» que llevan información nunca dependen del parpadeo de ilustraciones; reducción de movimiento en vivo (sin recargar).

## 8. Textos de interfaz nuevos (es/en, FR-019)

| Clave | Español | English |
|---|---|---|
| `theme.label` | Tema | Theme |
| `theme.toDark` | Cambiar a tema oscuro | Switch to dark theme |
| `theme.toLight` | Cambiar a tema claro | Switch to light theme |
| `a11y.themeState` | (anunciado por aria-pressed, sin cadena nueva) | (idem) |

El resto de rótulos y frases de interfaz ya existen en F3 (`nav.sections.*`, `lang.*`, `footer.welcome`, `whatsapp.action.*`, etc.) y **no cambian de texto ni de clave**: solo de presentación. Las propuestas NO introducen textos de contenido nuevos (el contenido viene del panel, intacto).
