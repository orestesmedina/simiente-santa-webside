# UX: F3 — Portada e información general

> Documento de diseño para `disenador-ux`. Fuentes: `spec.md` (F3, aclaraciones del 2026-10-09), `ux.md` de F2 (plantilla de pantallas, estados y componentes), **Manual de Identidad de la Iglesia Simiente Santa** (`resources/MANUAL DE MARCA.pdf`) y el código actual de `frontend/` (React 19 + Tailwind 4, componentes de F2 ya construidos).
> No es código: el `dev-frontend` lo implementa. Rutas y contratos los fija el `arquitecto`; aquí se proponen.
> **Decisiones de UX a revisar por el humano**: ver §9 (resumen). Cada una lleva su marca ⟲.

---

## 0. Principios de diseño

La marca manda. Estos principios salen del Manual de Identidad y del encargo del cliente («nada de plantilla genérica», adaptable a personas de todas las edades):

1. **Cercano, no institucional.** La portada habla como una persona: «Bienvenido a tu casa», no «Le extendemos una cordial bienvenida». Títulos claros y frases esperanzadoras; cero formalismos. La voz se aplica a la portada **y al panel** (los mensajes del panel heredan el tono: «El servicio quedó en borrador; publícalo cuando quieras que el público lo vea.»).
2. **Moderno y limpio.** Mucho aire, una paleta aplicada con proporción (70% azul marino, 20% blanco/crema, 10% turquesa/acentos), títulos en Bebas Neue de gran tamaño y cuerpo en Poppins. Nada de sombras fuertes, gradientes decorativos ni iconografía suelta; el logo nunca lleva efectos (regla del manual).
3. **Apto para todas las edades.** Cuerpo mínimo 18 px, objetivos táctiles ≥ 44×44 px, iconos siempre **con** texto, navegación de secciones con enlaces de ancla siempre visibles — **no hay menú hamburguesa ni acordeones** en la portada ⟲ (el cliente pidió «sin menús ocultos ni gestos desconocidos»). Todo lo importante se ve a primera vista.
4. **Español base, inglés al alcance de un toque.** El selector de idioma vive en la cabecera, visible siempre con los dos idiomas escritos completos («Español» / «English»), sin gestos raros ni suposiciones sobre el navegador.
5. **Solo lo publicado existe.** Si una sección no tiene nada publicado, no aparece: ni título, ni «próximamente», ni huecos. La página siempre se ve completa y cuidada.
6. **El backend manda** (heredado de F2): la UI oculta lo que el rol no permite, y toda operación se vuelve a verificar en el servidor.
7. **Panel consistente, portada con marca.** El panel de F3 reutiliza el sistema de F2 (`Button`, `Field`, `Notice`, `Table`, `Tabs`… con su look slate) para que el equipo no reaprenda nada; la identidad de marca (paleta, tipografías) se aplica al **sitio público**. Mantener el panel como está en F3 es decisión ⟲ revisable (repintar el panel con la paleta de marca podría hacerse después como mejora transversal).

---

## 1. Design tokens

### 1.1 Colores

Definirlos como variables CSS de Tailwind 4 en `frontend/src/index.css` (`@theme { ... }`). Nombres de token en inglés (coherente con el código), etiqueta en español para el equipo ⟲.

| Token | Valor | Español | Uso y proporción (70/20/10) |
|---|---|---|---|
| `navy` | `#1a2b4a` | azul marino | Fondos grandes (cabecera, hero, pie, secciones alternas), texto sobre claro. **70 %.** |
| `navy-soft` | derivado, p. ej. `#e3e8f0` | azul empolvado | Superficies suaves y bordes finos de tarjetas sobre claro. |
| `teal` | `#00c9a7` | turquesa | Detalle, subrayados, chips, icono activo, hover de acento. **10 %.** |
| `teal-strong` | derivado, p. ej. `#0e7a66` | turquesa profundo | **Único turquesa usado como texto o icono pequeño sobre claro** (por contraste, ver abajo). |
| `cream` | `#F5F2EC` | crema | Fondo de secciones de lectura y tarjetas sobre claro. **20 %.** |
| `coral` | `#ff6b3d` | naranja | Acento cálido puntual (p. ej. foco decorativo del CTA de WhatsApp). Muy dosificado. |
| `leaf` | `#217638` | verde | Éxitos (hereda del rol de verde del `Notice` de éxito), elemento de la hoja del logo. |
| `white` / `black` | `#ffffff` / `#000000` | blanco / negro | Texto sobre navy, fondos. |

**Reglas de contraste (verificadas con la paleta):**

- `navy` sobre `white` ≈ 12.6:1 y sobre `cream` ≈ 11.8:1 → texto normal AA sobrado; **toda la lectura va en navy sobre blanco/crema.**
- `teal` sobre blanco ≈ 2.3:1 → **prohibido como texto o icono pequeño sobre claro**; solo como fondo (con texto `navy` encima ≈ 6.7:1, AA), borde decorativo o elemento grande gráfico.
- Fondo `coral` con texto `navy` ≈ 5.0:1 → AA; `coral` nunca como texto sobre claro (2.8:1).
- `leaf` como texto sobre blanco ≈ 5.7:1 → AA válido para estados de éxito (permite mantener el `Notice` verde de F2 en el panel).
- Sobre fondos `navy`, el texto es AA como `white` (o `cream`); el **foco visible** en esos contextos usa anillo `cream`/`white` (ver §1.4).

### 1.2 Tipografías

| Familia | Rol | Mapeo de código |
|---|---|---|
| **Bebas Neue** | Títulos y nombre de la iglesia (H1/H2 del sitio público) | `--font-display` → `font-display` |
| **Poppins** (normal/itálica) | Todo el cuerpo, botones y formularios | `--font-sans` → fuente por defecto |
| **Playfair Display** | Énfasis emocional: lema, citas de misión/visión, frases de bienvenida | `--font-emotiva` → `font-emotiva` |

Escalas (móvil → escritorio):

- **H1 (nombre oficial en el hero):** `clamp(2.5rem, 7vw, 4.75rem)`, interlineado 1.0, en `white` sobre imagen o `navy` sobre claro.
- **H2 (título de sección):** `clamp(1.75rem, 4.5vw, 2.5rem)` Bebas, con acento `teal` (línea o columna) como marca de sección.
- **H3 (nombre de un servicio/canal):** Poppins 600 `1.25rem`.
- **Cuerpo:** Poppins `1.125rem` (18 px) / 1.6 mínimo (clientela de todas las edades, herencia de F2).
- **Énfasis emocional:** Playfair Display itálica `clamp(1.25rem, 3vw, 1.75rem)`.
- **Pequeño (pie, etiquetas auxiliares):** ≥ `1rem` en el sitio público; nada por debajo en portada.

**Carga de fuentes** ⟲: self-hosted recomendado (`@fontsource/bebas-neue`, `@fontsource/poppins`, `@fontsource/playfair-display`) por rendimiento y para evitar dependencias externas en un sitio de iglesia; la alternativa (Google Fonts) la fija el `arquitecto`.

### 1.3 Espaciado, radios y sombras

- Contenedor de portada: `max-w-6xl`, márgenes laterales `px-4` (móvil) / `px-6` (tableta en adelante); secciones `py-12` móvil → `py-20` escritorio.
- Lectura larga («quiénes somos»): columna `max-w-3xl`.
- Radios: campos y botones `rounded-lg` (8 px, hereda coherencia con el panel); **tarjetas de portada** `rounded-2xl` (acogedor, cercano).
- Sombras: una sola, suave y cálida — `0 2px 8px rgb(26 43 74 / 0.10)` para tarjetas flotantes; el resto por bordes finos `navy-soft`. La marca pide cercanía sin adornos: sin neomorfismo, sin gradientes sobre el logo, sin animaciones de entrada por defecto (respeto de `prefers-reduced-motion`, §8).

### 1.4 Botones, formularios y foco

- **Botones** (extender `Button` de F2 con una variante nueva, sin romper las existentes) ⟲:
  - `primary` → fondo `navy`, texto `white` (equivalente al slate-900 de hoy; se re-tiñe con `navy`).
  - `accent` (nueva, portada) → fondo `teal`, texto `navy`; para los CTAs de WhatsApp y acciones cálidas.
  - `secondary` → borde `navy-soft`, texto `navy`, fondo `white`/`cream`.
  - `danger` y `link` → sin cambios (F2).
- **Formularios**: solo existen en el panel y reutilizan tal cual `Field`, `Select`, `Notice`, `Dialog` y `ConfirmDialog` de F2 (misma apariencia slate). La edición es/en no introduce estilos nuevos: párrafo de ayuda + contador, ver §4.5. En el sitio público NO hay formularios (su comunicación es solo por enlaces, Out of Scope de la spec).
- **Foco visible**: anillo de 2 px con `outline-offset: 2px`; color por contexto — `navy` sobre fondos claros, `cream`/`white` sobre fondos `navy`/imagen (implementar como utilería `focus-ring-clara` / `focus-ring-oscura` ⟲; lo fija el `arquitecto` en el plan).

---

## 2. Flujo de usuario

### 2.1 Visitante: primera vez (camino feliz, US1)

1. Abre `https://<sitio>` → carga la **portada** (`/`), siempre en español de inicio (Decisión 8). Ve la cabecera con logo + nombre y los enlaces de sección.
2. Baja naturalmente o usa los enlaces de ancla: **Quiénes somos · Horario · WhatsApp · Contacto · Redes** (solo aparecen las secciones con contenido publicado).
3. En **Hero** conoce el nombre, la misión y la visión de la iglesia con su imagen.
4. En **Horario** encuentra día, hora, nombre y lugar de cada servicio — sin entrar a ningún lado: la lista es la información.
5. En **WhatsApp** toca un canal («Escribir por WhatsApp» o «Entrar al grupo») → se abre el chat en la app del teléfono o en la web.
6. En **Redes** toca la red («Ver en Facebook», etc.) → se abre el perfil de la iglesia.
7. En **Contacto** ve dirección, correo y teléfono, con enlaces cómodos (`tel:`, `mailto:`); la dirección es texto y se puede copiar a mano ⟲ (sin mapa incrustado en MVP, ver decisiones).

### 2.2 Visitante: cambia a inglés (US4)

1. En la cabecera toca el selector «Español / English» → toca **English**.
2. La página entera cambia de inmediato (FR-010): interfaz y contenidos traducidos.
3. Todo contenido **sin versión en inglés** se muestra en español, con su texto completo; **nunca** un hueco ni una traducción automática (FR-009). Sin aviso por elemento: el fallback es el comportamiento normal del sitio, no una incidencia que el visitante deba conocer ⟲.
4. La elección se mantiene al navegar (las F4+ heredan el mismo contexto React de idioma) y entre visitas en el dispositivo (`localStorage`, confirmado por el humano el 2026-10-09); una visita **nueva** siempre entra en español.
5. Volver a «Español» restaura todo al instante, sin pérdida de información (US4 esc. 4).

Errores del idioma: si el guardado del idioma falla (p. ej. navegación privada), no pasa nada grave — solo no persiste entre visitas; no se avisa.

### 2.3 Equipo: edita y publica (US2, US3)

1. Inicia sesión → panel → en el **Inicio** aparece la tarjeta de gestión **«Portada e información»** (solo con el permiso del módulo).
2. Entra a **`/panel/informacion`**: pestañas con las 6 piezas (Identidad · Quiénes somos · Horario de servicios · WhatsApp · Redes sociales · Contacto). Ve de un vistazo el **estado** (Borrador/Publicado) de cada elemento.
3. **Crear un servicio** (ejemplo): pestaña Horario → «Agregar servicio» → formulario (día, hora, nombre, lugar; pestaña «English» opcional) → Guardar → el servicio queda en **Borrador**: visible en el listado del panel, invisible en la portada. Aviso: «Servicio guardado en borrador. Publícalo cuando esté listo.»
4. **Publicar**: fila → «Publicar» (sin confirmación; es la acción esperada) → píldora cambia a «Publicado» → el elemento ya es visible al visitante. Aviso: «Publicado en la portada.»
5. **Editar un elemento publicado**: formulario → Guardar → aviso **«Cambios guardados: ya visibles en la portada.»** (el sistema recuerda que se publica al guardar, FR-014) ⟲.
6. **Retirar**: fila → «Retirar de la portada» → confirmación clara → desaparece del sitio público, sigue en el panel como borrador.

### 2.4 Errores del equipo

| Situación | Camino |
|---|---|
| Datos inválidos o incompletos | Nada se guarda; el error vive **junto al campo** y hay resumen arriba del formulario; el foco va al primer campo invalidado (FR-015). |
| Enlace mal formado (wa.me / red social) | Mensaje con el formato esperado y ejemplo real del catálogo. |
| Contenido sin versión en español | Se impide guardar con mensaje que explica el rol del español base (Decisión 8). |
| Límite de «quiénes somos» | Contador en rojo con cuántos caracteres hay que recortar. |
| Imagen inválida / carga fallida | Aviso del problema con el archivo; el formulario conserva el resto de los datos; la portada no se rompe. |
| Error del sistema al guardar | Aviso genérico sin detalles internos **y sin perder lo escrito** (el formulario conserva los valores). |
| Operación sin permiso | Ruta protegida con `RequirePermission` → `ForbiddenPage` de F2; y en caso de rechazo del servidor, el mismo mensaje de F2 («No tienes acceso a esta sección…»). |
| Edición simultánea (dos personas) | Se conserva la última versión completa guardada (default de la spec); la UI no detecta conflicto (sin aviso de concurrencia en MVP) ⟲. |

### 2.5 Persona sin permiso (el caso desfavorable)

- La entrada no existe para ella: el menú del panel no muestra «Portada e información» y el Inicio no muestra su tarjeta de acceso.
- Si fuerza la URL (`/panel/informacion`): `RequirePermission` la lleva a `/sin-permiso` (pantalla de F2, texto sin cambios). Nada de la información general se revela.
- Desde el sitio público no hay nada que forzar: no existe acción de edición ahí; el HTML público jamás contiene borradores (SC-002).

---

## 3. Rutas propuestas (React Router)

```
Sitio público (sin sesión):
/                      Portada de la iglesia (HomePage)
/health                «Estado del sistema» de F1, movida fuera de la portada (decisión D-4)

Panel (F2, sin cambios):
/login  /sin-permiso  /cambiar-contrasena
/panel                 Inicio
  /panel/informacion   RequirePermission(portada). Módulo F3 (pestañas)
  /panel/usuarios  /panel/roles  /panel/auditoria   (F2)
```

- La portada **reemplaza** la página de F1 como inicio de `/` (decisión del humano del 2026-10-09); el «Estado del sistema» se conserva como página de diagnóstico en **`/health`** y **no aparece en la navegación pública** (no es contenido de la iglesia).
- `<AppLayout>` de F1 se reemplaza por el nuevo `PublicLayout` (con marca e i18n); el `AppLayout` actual solo quedará para `/health`.
- El idioma **no crea rutas** (`/en/…` no existe en MVP): el cambio es inmediato en la misma URL ⟲.
- `RequirePermission.code` para este módulo ⟲: código técnico exacto lo fija el `arquitecto` (el permiso existe en el catálogo de F2; en las pruebas de F2 aparece como `portada` — «Portada e información general»).

---

## 4. Pantallas

### 4.1 Portada pública — `/` (visitante, sin sesión; US1/US4/US5)

**Propósito**: primera cara del sitio. En menos de 30 s (SC-001) cualquier persona sabe quién es la iglesia, cuándo se reúne y cómo contactarla.

**Estructura** (una sola página, scroll natural; los anclas apuntan a cada sección):

1. **Enlace de salto** (_skip link_): «Saltar al contenido principal» — primer `Tab` de la página.
2. **Cabecera** (fondo `navy`): logo pequeño + nombre «Simiente Santa» (enlace al inicio), enlaces de ancla a las secciones visibles, selector de idioma (radios nativos «Español»/«English») ⟲.
3. **Hero — identidad**: imagen de portada como fondo con capa `navy`; sobre ella: logo en `object-contain` (nunca se deforma), H1 con el nombre oficial, misión y visión como bloques etiquetados («Nuestra misión», «Nuestra visión») y el lema en Playfair si existe. Si no hay imagen publicada o falla su carga: fondo `navy` limpio, la sección sigue completa (nunca hueco) FR-002/Edge.
4. **Quiénes somos**: título de sección + texto plano (con saltos de línea simples preservados), columna de lectura `max-w-3xl`.
5. **Horario de servicios**: lista de tarjetas (móvil) / tabla ligera (≥ tablet): **Día, Hora, Servicio, Lugar**; cada servicio una tarjeta clara con esos 4 datos visibles juntos.
6. **Canales de WhatsApp**: tarjetas con nombre/propósito como título y botón de acción — «Escribir por WhatsApp» (número) o «Entrar al grupo» (enlace de grupo). Icono de WhatsApp siempre con texto.
7. **Contacto**: tres datos con etiqueta y valor: **Dirección**, **Correo**, **Teléfono**; correo y teléfono como enlaces (`mailto:`, `tel:`); la dirección como texto plano.
8. **Redes sociales** (en el contenido, no solo en el pie): una tarjeta/fila por red del catálogo publicada, con icono **+ nombre** («Facebook de la iglesia» y no solo un icono).
9. **Pie** (`navy`): logo, nombre, redes sociales en versión compacta, selector de idioma y una frase corta de voz de marca si el equipo la carga (campo opcional, ejemplo: «Esta siempre será tu casa.») ⟲.

**Contenido, no decoración**: nada de poema en el hero: el texto del hero son campos reales del panel (nombre, lema, misión, visión). Las frases de voz de marca **no las inventa el diseño**: son contenido editable.

**Estados de la portada**:

| Estado | Qué ve el visitante |
|---|---|
| Carga (primera) | Esqueletos por sección (títulos + líneas grises `navy-soft`) con `<p role="status">Cargando…</p>`; la cabecera y el idioma ya visibles ⟲. |
| Con contenido | Secciones completas con su marca. |
| Sección sin nada publicado | **La sección no existe en la página** (ni título ni rastro); su ancla desaparece de la cabecera (SC-012, FR-013). |
| Contenido sin inglés (en inglés) | El texto se muestra en español, integrado en idéntico diseño; sin campo vacío ni marca de «sin traducir» (FR-009). |
| Error de carga sin nada en caché | Composición hero + mensaje «No pudimos cargar la información de la iglesia. Revisa tu conexión y vuelve a intentarlo.» + botón «Volver a intentar» — nunca detalles técnicos. |
| Error de carga con contenido previo | Se mantiene el último contenido y se muestra abajo un aviso discreto: «No pudimos actualizar la información; lo que ves es la última versión disponible.» ⟲ |
| Imagen rota o sin imagen | El hero conserva el fondo `navy` y el texto; la portada no se rompe (Edge). |
| Borrador por cualquier vía | Jamás se pintan borradores, ni sus textos, ni sus imágenes (SC-002). |

**Responsivo** (móvil primero):

- **Móvil (< 640 px)**: columna única; hero ~70 vh con imagen y textos legibles (contraste garantizado sobre la banda); enlaces de ancla en una fila con desplazamiento horizontal (sin menú oculto) ⟲; tarjetas de WhatsApp/red a todo lo ancho con botones de 48 px de alto; contacto en tarjetas; secciones separadas por aire `py-12`.
- **Tableta (≥ 640 px)**: horario en tabla con encabezados; contacto en 3 columnas; hero con textos mayores.
- **Escritorio (≥ 1024 px)**: hero a sangre completa; contacto y WhatsApp en tarjetas con mayor aire; contenido centrado `max-w-6xl`; el texto se amplía sin romper (diseño fluido en `rem`).
- **Sin desplazamiento horizontal** en el ancho mínimo probado de 320 px (SC-007).

### 4.2 Panel — Inicio (cambio menor en F2)

La tarjeta «Gestión» de `InicioPage` añade el acceso **«Portada e información»** (icono + texto, mismo patrón que Usuarios/Roles/Auditoría) **solo** si la sesión tiene el permiso del módulo. Cuentas sin el permiso no ven nada nuevo (heredado del patrón).

### 4.3 Panel — Portada e información, vista general — `/panel/informacion`

**Propósito**: administrar todo el contenido de la portada desde un lugar único (US2, US3; FR-011/012).

- **Estructura**: título «Portada e información general» + una banda inicial de ayuda («Lo que guardes aquí se muestra en la portada pública de la iglesia. Solo lo publicado en cada sección es visible al público.») + **`Tabs`** (componente de F2): **Identidad · Quiénes somos · Horario de servicios · WhatsApp · Redes sociales · Contacto**.
- Cada pestaña carga bajo demanda (TanStack Query) con skeleton.
- Reglas transversales: todo elemento tiene **píldora de estado** — `StatusPill` con valores `draft`/`published` («Borrador» en gris, «Publicado» en verde) ⟲; el formulario siempre guarda primero, y **publicar/retirar es una acción dedicada por cada sección** (identidad, quiénes somos, contacto) y **por elemento** en los listados (servicios, canales, redes): ninguna acción agrupa varias secciones (FR-013, decisión D-2 del 2026-10-09).
- Cross-check con F2: la edición queda registrada en la auditoría de F2 (FR-017) **sin UI distinta visible** para el usuario (lo muestra el módulo de auditoría).

### 4.4 Panel — Identidad (pestaña)

**Contenido**: **Nombre oficial** (no traducible, obligatorio) · **Lema** (es obligatorio; en opcional) ⟲ · **Misión** (es/en, opcional el inglés) · **Visión** (es/en) · **Logotipo**: subir/ver imagen con previsualización y **texto alternativo** obligatorio (es/en; se usa en el logo de cabecera y pie) · **Imagen de portada**: subir/ver con previsualización y texto alternativo obligatorio. Decisiones clave:
- Recordatorio visible bajo el campo del logo con las reglas del manual: «El logo no se deforma, no cambia de color, no se gira ni lleva efectos. Sube el archivo tal como lo entregó el diseñador.» El sistema **no** valida (ni puede) el cumplimiento de marca: es guía ⟲.
- El logo SIEMPRE se muestra respetando su proporción original (`object-contain`) en portada y cabecera; **la imagen de portada** sí se recorta visualmente al centro (`object-cover`) por diseño responsivo, con aviso en la ayuda: «La imagen se recorta ligeramente según el tamaño de pantalla.» ⟲
- Un solo campo `alt` es/en por imagen (FR-019) ⟲.
- Estado de publicación: la **sección Identidad** se publica/retira por su cuenta (FR-013, decisión D-2 del 2026-10-09): sus textos, logo e imagen de portada van juntos como un único elemento gestionable.

### 4.5 Panel — Quiénes somos (pestaña)

**Contenido**: un `textarea` **texto plano** (sin formato, sin HTML) con contador en vivo `{n}/1.000` y pestañas «Español» (obligatorio) / «English (opcional)»; ayuda: «Escríbelo con tus palabras; se muestran los saltos de línea tal cual.» Al superar 1.000 caracteres el contador se torna visible y el guardado se bloquea con mensaje claro (FR-003/FR-015). Se publica/retira por su cuenta (FR-013).

### 4.6 Panel — Horario de servicios (pestaña)

- **Listado**: tarjetas (móvil) / tabla (tableta+): día, hora, nombre, lugar con píldora de estado y acciones **Editar / Publicar / Retirar / Agregar**.
- **Agregar servicio** → `Dialog` con formulario (estilo F2): **Día** (texto libre con ayuda «Ej.: Domingos»), **Hora** (texto libre con ayuda «Ej.: 10:00 a. m. − 12:00 m.») ⟲ (sin selector de hora: las iglesias escriben horas con rangos y variantes locales), **Nombre/descripción** (texto), **Lugar** (texto con ayuda «Ej.: Templo central»), y pestaña «English (opcional)» con los mismos 4 campos (FR-004). Nada se traduce por sí: el **día** es el texto traducible; la **hora** es igual en ambos idiomas ⟲.
- Duplicado exacto: igual día+hora+nombre+lugar → rechazado con mensaje («Ya existe un servicio exactamente igual; puedes publicarlo una sola vez.») ⟲ por default de spec.

### 4.7 Panel — Canales de WhatsApp (pestaña)

- **Listado** tarjetas/tabla: nombre/propósito, **tipo** (Mensaje directo / Grupo), destino legible (número con formato, o dominio del enlace), estado y acciones.
- **Agregar canal** → formulario: **Nombre/propósito** (es/en) · **Tipo** (radio: «Número para mensaje directo» / «Enlace de grupo») · destino único:
  - **Número**: campo con criterio telefónico de F2; ayuda «Con código de país: +506 8888 8888». El sistema construye el enlace `wa.me` al guardar ⟲.
  - **Grupo**: campo de URL; ayuda «Pega la invitación del grupo: empieza por https://chat.whatsapp.com».
- Duplicado exacto (mismo destino + mismo nombre) → rechazado con mensaje del catálogo. Cada canal = un elemento publicable individual.

### 4.8 Panel — Redes sociales (pestaña)

- **Listado fijo del catálogo**: Facebook, Instagram, YouTube, TikTok y Spotify — **una fila por red** (aunque no tenga enlace), cada una con **URL única**, estado y acciones (cada red publicada es un elemento publicable) ⟲.
- Si una red no tiene enlace: la fila se muestra con el texto «Sin enlace todavía» y el botón Editar; el catálogo SIEMPRE se muestra completo; lo que cambia por red es si está publicada o no.
- Solo las filas **publicadas** aparecen en la portada.

### 4.9 Panel — Contacto (pestaña)

- Formulario: **Dirección** (obligatoria, es/en), **Correo** (obligatorio, formato correo; un solo valor), **Teléfono** (obligatorio, criterio telefónico de F2; un solo valor). La **sección Contacto** se publica/retira por su cuenta (FR-013, decisión D-2 del 2026-10-09); sus tres datos van juntos dentro de la sección, de modo que no quede medio publicada (p. ej. dirección sin teléfono).

### 4.10 Pantalla de estado de sección vacía (panel, cada pestaña de listado)

- «Todavía no hay {servicios / canales}. Agrega el primero para que aparezca en la portada.» + botón de agregar.
- La sección visible en portada ocurre con al menos un elemento **publicado**; en el **borrador** se muestra como fila normal en el panel.

---

## 5. Estados por vista con datos (panel)

### a. Vista general del módulo (pestañas)

| Estado | Qué ve |
|---|---|
| Cargando | Esqueletos en la zona de datos del listado; la banda de pestañas ya visible e interactiva. |
| Vacío | EmptyState por pestaña con acción «Agregar …». |
| Error | «No se pudo cargar la información.» + «Reintentar» (Notice de error, `role="alert"`). |
| Éxito | Listado con píldoras de estado y acciones. |
| Sin permiso | Ruta protegida: no se llega; si algo pasa → ForbiddenPage de F2. |
| Éxito de operación | Notice temporal verde: «Cambios guardados.» / «Publicado en la portada.» / «Retirado de la portada.» |

### b. Formularios (identidad, quiénes somos, servicio, canal, redes, contacto)

- Precargado al editar; vacío al crear. Los campos de «English» van en su pestaña con la marca «opcional».
- Guardando: `Button loading` («Guardando…»), cierre del modal bloqueado.
- Éxito: cierre + Notice en el listado («Cambios guardados.», o «Cambios guardados: ya visibles en la portada.» si era publicado).
- Errores de campo (con los textos del §7): junto al campo y resumen arriba (patrón F2): obligatorios, formatos (correo, teléfono, WhatsApp, redes), falta del español base, límite de 1.000, imagen inválida.
- Modal de edición con `Dialog` (F2): `Esc` cierra, trampa de foco, devolución del foco a la fila que lo abrió.

### c. Publicar / retirar

- **Publicar** acción de fila, sin confirmación. La píldora cambia a «Publicado»; el foco se mantiene en la fila tras la operación.
- **Retirar**: `ConfirmDialog` suave: título «Retirar de la portada», descripción «"{nombre}" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.», botones «Retirar» (danger) / «Dejar como está» ⟲.
- Los elementos ya publicados se **auto-publican** al guardar (FR-014), con un aviso que lo dice en palabras del usuario (no jerga técnica).

---

## 6. Componentes React

> Nombres definitivos en inglés (coherente con F2). El listado final de tareas lo fija el `arquitecto`; aquí es propuesta con la responsabilidad de cada componente.

### 6.1 Reutilizados de F2 (sin cambios o con extensión menor)

| Componente | Extensión pequeña ⟲ | Uso |
|---|---|---|
| `Button` | añadir variante `accent` (teal/navy); colores `primary/secondary` re-tintados a `navy` en el sitio público (sin romper el panel) | CTAs de la portada y botones del panel. |
| `Field` | — | Formularios del módulo. |
| `Select` | — | Tipo de canal en el formulario (si el arquitecto lo modela como lista), no aplica al sitio público. |
| `Notice` | — | Avisos del módulo y mensajes públicos de la portada. |
| `Dialog` / `ConfirmDialog` | — | Formularios en modal, retirar publicación. |
| `EmptyState` | — | Estados vacíos del panel. |
| `StatusPill` | añadir `draft`/`published` con textos «Borrador»/«Publicado» ⟲ | Estados de publicación. |
| `Table` | — | Listados del panel (servicios, canales, redes). |
| `Tabs` | — | Pestañas del módulo y pestaña de idioma es/en del formulario. |
| `Pagination` | **no se usa en F3** ⟲: los listados de portada son cortos por naturaleza (sin paginación; revisable si alguna sección creciera). | — |

### 6.2 Nuevos del sitio público (`frontend/src/features/publico/`)

| Componente / hook | Responsabilidad |
|---|---|
| `PublicLayout` | Cabecera con marca (logo, nombre), navegación de anclas, slot del selector de idioma, pie. |
| `LanguageSwitcher` | Radios nativos «Español/English» estilizados; cambia el `LanguageContext` y persiste (es/en interfaz). |
| `HomePage` | Composición de secciones a partir de `usePublicHomeData`; decide qué secciones se muestran (solo publicadas). |
| `IdentityHero` | Imagen de fondo + logo con `object-contain` + nombre en H1 + bloques Misión/Visión/Lema; cubre el caso de imagen fallida o ausente. |
| `WhoWeAreSection` | Texto plano con saltos, máximo ancho de lectura. |
| `ScheduleSection` | Lista/tarjetas de servicios con `Table` adaptada público (día, hora, nombre, lugar). |
| `WhatsAppSection` | Tarjetas de canal con CTA correcto por tipo. |
| `SocialSection` | Fila/tarjeta por red del catálogo con icono + nombre — enlaces `target="_blank" rel="noopener"`. |
| `ContactSection` | Dirección/correo/teléfono con enlaces nativos. |
| `PublicFooter` | Logo, frase opcional, redes, idioma. |
| `SectionSkeleton` | Esqueletos por sección (estilo brand: líneas `navy-soft`). |
| `SocialIcon` | SVG inline propio (facebook, instagram, youtube, tiktok, spotify) — sin dependencias nuevas ⟲. |
| `usePublicHomeData` | TanStack Query: una consulta pública (contrato del `arquitecto`); respeta caché y estructura por secciones. |
| `LanguageProvider` / `useLanguage` / módulo i18n (`features/publico/i18n`) | Contexto de idioma de interfaz es/en; diccionario de cadenas; actualiza el `html lang`; `textContentFor(content, lang)` — helper de fallback al español (nunca campo vacío). |
| `features/publico/messages.ts` | Catálogo de cadenas de interfaz es/en (texto plano, sin frameworks de i18n pesados) ⟲. |

### 6.3 Nuevos del panel (`frontend/src/features/informacion/`)

| Componente / hook | Responsabilidad |
|---|---|
| `InformationPage` | Pestañas de las 6 piezas, banda de permiso (RequirePermission), errores comunes. |
| `IdentityForm` | Nombre + lema/misión/visión es/en + subida **logo/portada** con previsualización y `alt` es/en + estado de publicación. |
| `WhoWeAreForm` | Texto plano + contador 1.000 + pestañas de idioma + publicar/retirar. |
| `ServiceForm` (en `Dialog`) | Alta/edición de servicio con pestaña «English (opcional)». |
| `ServicesList` | Tarjetas/tabla con píldora y acciones publicar/retirar por servicio. |
| `WhatsAppForm` (en `Dialog`) | Nombre + tipo + destino único (teléfono validado, o URL de grupo) con ayuda contextual. |
| `WhatsAppList` | Tarjetas/tabla por canal. |
| `SocialsList` | Filas fijas del catálogo con URL por red y estado. |
| `ContactForm` | Dirección (es/en) + correo + teléfono; la sección se publica/retira por su cuenta. |
| `ImageUploader` | Field de archivo + previsualización inmediata + texto alternativo (`alt`) es/en + aviso de reglas del manual para el logo. ⟲ |
| `PublishControls` | Botones «Publicar»/«Retirar de la portada» reutilizando `ConfirmDialog`. |
| `features/informacion/messages.ts` | Textos de error y confirmación del módulo (§7). |
| Hooks TanStack Query | `usePublicHome` (pública), `useIdentity`, `useUpdateIdentity`, `useWhoWeAre`, `useServices`, `useCreateService`, `useUpdateService`, `useWhatsappChannels`, `useCreateWhatsappChannel`, `useUpdateWhatsappChannel`, `useSocials`, `useContact`, y `usePublish`/`useUnpublish` por elemento ⟲ (nombres finales los fija el `arquitecto` en los contratos). Ningún `fetch` en componentes (regla de F2). |

**Pruebas esperadas** (Vitest + Testing Library + MSW, lógica visible): composición de la portada con secciones ocultas por estado (0 secciones vacías), fallback es→en para contenido sin inglés, selector de idioma es↔en persistente, validaciones de formularios (formatos de enlace, correo, teléfono, límite 1.000, español obligatorio, imagen inválida) y ciclo de publicación (borrador → publicar → visible; retirar → invisible), y guardas de permiso. Detalles en §10.

---

## 7. Catálogo de textos

### 7.1 Panel (español; FR-016) — añadidos a los de F2

| Situación | Texto |
|---|---|
| Guardado | «Cambios guardados.» |
| Guardado con auto-publicación | «Cambios guardados: ya visibles en la portada.» |
| Publicar | «Publicado en la portada.» |
| Retirar | «Retirado de la portada. Estará disponible como borrador para publicarlo cuando quieras.» |
| Guardar (crear servicio/canal/red) | «Servicio creado.» / «Canal de WhatsApp creado.» / «Enlace de {red} guardado.» + aviso informativo la primera vez: «Quedó en borrador: no lo verá el público hasta que lo publiques.» ⟲ |
| Retirar (confirmación) | «"{nombre}" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.» |
| Campo obligatorio vacío | «Escribe {el nombre / la dirección / la misión…}.» |
| Correo inválido | «Este correo no tiene el formato correcto.» |
| Teléfono inválido | «Escribe un número de teléfono válido con código del país si corresponde.» |
| Enlace de WhatsApp (grupo) mal formado | «Ese enlace no parece de WhatsApp. Pega la invitación del grupo: empieza por https://chat.whatsapp.com.» |
| Enlace de red social mal formado / fuera de catálogo | «Ese enlace no corresponde a la red seleccionada. Pega el perfil de la iglesia en {red}, por ejemplo {ejemplo}.» |
| Red social repetida | «Cada red admite un solo enlace. Ya existe un enlace para {red}: edítalo o retíralo antes de cambiarlo.» |
| Canal duplicado | «Ya existe un canal igual (mismo nombre y mismo destino). Edita el que ya tienes o cámbiale el nombre.» |
| Límite de «quiénes somos» | «{n} de 1.000 caracteres. Recorta {m} para poder guardar.» |
| Sin versión en español | «El contenido en español es obligatorio: es el idioma base de la portada.» |
| Imagen inválida | «Ese archivo no es una imagen válida. Usa JPG o PNG de menos de {tamaño} y prueba de nuevo.» ⟲ |
| Error del sistema al guardar | «No se pudo completar la operación. Tus escritos están a salvo; vuelve a intentarlo en unos minutos.» |
| Sin permiso desde el servidor | «No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.» (igual que F2) |
| Ayuda del logo | «El logotipo no se deforma, no cambia de color, no se gira y no lleva efectos: sube el archivo original.» |
| Ayuda imagen portada | «Aparece detrás del título en la portada. Se recorta ligeramente según la pantalla; usa una imagen horizontal.» |
| Ayuda «English (opcional)» | «Si lo dejas vacío, el público verá este contenido en su versión en español.» |

### 7.2 Sitio público — cadenas de interfaz es/en

> El contenido de la iglesia **no** va aquí: viene del panel (es → fallback). Son solo las palabras de la interfaz.

| Clave | Español | English |
|---|---|---|
| `nav.sections.who` | Quiénes somos | Who we are |
| `nav.sections.schedule` | Horario de servicios | Service times |
| `nav.sections.whatsapp` | WhatsApp | WhatsApp |
| `nav.sections.contact` | Contacto | Contact |
| `nav.sections.social` | Redes sociales | Social media |
| `lang.label` | Idioma | Language |
| `lang.es` / `lang.en` | Español / Inglés | Spanish / English |
| `identity.mission` | Nuestra misión | Our mission |
| `identity.vision` | Nuestra visión | Our vision |
| `identity.lema` | Nuestro lema | Our motto ⟲ |
| `contact.phone` | Teléfono | Phone |
| `contact.email` | Correo | Email |
| `contact.address` | Dirección | Address |
| `whatsapp.action.dm` | Escribir por WhatsApp | Chat on WhatsApp |
| `whatsapp.action.group` | Entrar al grupo | Join the group |
| `social.action` | Ver en {red} | Visit us on {network} |
| `footer.welcome` | «Esta siempre será tu casa.» | «This will always be your home.» ⟲ (frase editable en panel) |
| `error.load` | No pudimos cargar la información de la iglesia. Revisa tu conexión y vuelve a intentarlo. | We couldn't load the church information. Check your connection and try again. |
| `error.retry` | Volver a intentar | Try again |
| `error.partial` | No pudimos actualizar la información; estás viendo la última versión disponible. | We couldn't refresh the information; you're seeing the latest available version. |
| `status.loading` | Cargando… | Loading… |
| `a11y.skip` | Saltar al contenido principal | Skip to main content |
| `a11y.logo` | Logotipo de la iglesia Simiente Santa | Simiente Santa church logo |
| `a11y.hero` | Imagen de la portada de la iglesia | Church cover image |

Texto alternativo del logo y de la imagen de portada: **lo carga el equipo en el panel** (FR-019); las cadenas de arriba son el default si el equipo no definió un `alt` especializado.

---

## 8. Accesibilidad (WCAG 2.1 AA) e internacionalización

### 8.1 Accesibilidad

- **Teclado**: todo interactivo (enlaces de WhatsApp/redes, selector de idioma, anclas, pestañas y filas del panel) alcanzable y operativo; los enlaces de ancla son `a[href]` reales; el selector de idioma son radios nativos (flechas naturales). `Esc` cierra los modales (F2). Sin trampas de foco.
- **Foco visible** siempre (anillo de 2 px; color según fondo — §1.4). El foco inicial no salta al cargar la portada; el `h1` es el punto de referencia al usar el «skip link».
- **Semántica**: skip link; `<header>` / `<nav aria-label="Secciones de la portada">` / `<main>` / `<footer>`; un `<h1>` (nombre oficial) y jerarquía propia de sección (título Bebas + sección); tablas con `th` de ámbito en el listado de horarios (tableta+); tarjetas equivalentes con listas semánticas en móvil.
- **Más que color**: estados Borrador/Publicado con texto siempre; iconos de redes y WhatsApp con **texto** visible junto al icono (principio F2).
- **Imágenes**: el `alt` lo gestiona el equipo (obligatorio en es y para el logo y la portada, con en opcional); si una imagen falla, la sección conserva su texto y el diseño no se rompe; el logo lleva dimensiones explícitas para no desplazar el maquetado al cargar.
- **Contraste y letra**: reglas del §1.1; cuerpo 18 px; contraste AA en botones `accent` (texto navy sobre teal 6.7:1) y en focos sobre navy (anillo crema 11:1 aprox).
- **Zoom / texto ampliado**: soportar 200 % con reflow sin pérdida (SC-008/FR-018); nada fija ancho en `px` para texto.
- **Objetivos táctiles**: ≥ 44 px en todos los controles, incluidos los botones de canal y de redes (los CTAs van a 48 px de alto en móvil).
- **Avisos**: los del panel reutilizan `Notice` (rol correcto); el cambio de idioma del sitio público se anuncia una sola vez con `role="status"` cortés («La página se muestra ahora en inglés.»), sin anuncios por cada cambio de sección ⟲.
- **Prefers-reduced-motion**: sin parallax ni transiciones obligatorias; el skeleton es estático (`animate-pulse` desactivado si el visitante lo prefiere).
- **Idioma del documento**: `<html lang>` conmutado dinámicamente entre `es` y `en` según el selector; el contenido en fallback se envuelve con `lang="es"` cuando la página está en inglés (buena práctica para lectores de pantalla) ⟲.

### 8.2 Internacionalización (Decisión 6 y Decisión 8)

- **Interfaz** (siempre bilingüe): diccionario propio `es`/`en` sin librería externa; el idioma vive en un `LanguageProvider` (contexto React) y `localStorage` recuerda la elección del dispositivo entre visitas (confirmado por el humano el 2026-10-09); visita nueva → español; **no** se auto-detecta el idioma del navegador.
- **Contenido** (por elemento): español obligatorio, inglés opcional. La función `textContentFor({ es, en }, lang)` devuelve `en` si existe y no está vacío; si no, `es`; **nunca** muestra vacío. El mismo helper se usa en el sitio público y en las pruebas del fallback.
- **El panel NO cambia de idioma** (siempre español, FR-016); solo el sitio público tiene el `LanguageSwitcher`.
- Las cadenas de interfaz del sitio público salen SIEMPRE del diccionario; **nunca** texto literal dentro del `JSX` (habilita la traducción de interfaz al 100 %, SC-006).

---

## 9. Decisiones de UX para revisar por el humano (resumen)

| # | Decisión | Alternativa |
|---|---|---|
| D-1 ⟲ | Selector de idioma como **radios nativos** con ambos textos completos en la cabecera, sin auto-detección del navegador. | Un `select` compacto, o detección del idioma. |
| D-2 | Publicación y retiro **por sección, cada una por separado** (decidido por el humano el 2026-10-09): identidad, quiénes somos y contacto cada uno por su cuenta, igual que los servicios, canales y redes por elemento; ninguna acción agrupa varias secciones. | Un solo «Publicar la portada» que grupa varias o todas las secciones de una vez. |
| D-3 | **Hora de servicio** como texto libre con ayuda (permite rangos «10:00 a. m. − 12:00 m.»). | Selector de hora estructurado. |
| D-4 | La portada sustituye a la página de F1 en `/`; «Estado del sistema» (F1) se mueve a **`/health`**, fuera de la navegación pública (**decidido por el humano el 2026-10-09**). | Mantener el estado como tarjeta integrada dentro de la portada. |
| D-5 | **Sin menú hamburguesa** en la portada: anclas visibles en una fila desplazable de cabecera. | Menú plegable de cabecera. |
| D-6 | Sin CTA especial si no hay canal de WhatsApp publicado (el hero no inventa envíos de datos que no existen). | Reservar espacio de CTA igualmente. |
| D-7 | Un solo valor de **dirección/correo/teléfono** (sin sucursales ni variantes), según spec (FR-007, tres datos); los tres viven dentro de la sección Contacto (D-2). | Varios correos/teléfonos por fila. |
| D-8 | **Alt obligatorio** de las imágenes (logo/portada) con versión es/en. | Alt solo en español. |
| D-9 | La imagen de portada se recorta (`object-fit: cover`); el logo siempre se conserva con `object-contain` (nunca se deforma). | Mostrar la imagen de portada sin recorte con barras. |
| D-10 | El panel **no se repinta** con la paleta de marca en F3 (consistencia con F2); la identidad visual se aplica al sitio público. | Re-tintar completo el panel con la marca. |
| D-11 | i18n **sin dependencias externas** (diccionario propio minimalista). | react-i18next y amigos. |
| D-12 | Crear un elemento lo deja **en borrador por defecto** (coherente con el control de US3). | Crear en estado publicado u otra convención. |

---

## 10. Notas para el `dev-frontend` / `qa-tester`

- **La lógica con pruebas** está en: el helper de fallback `textContentFor` (sin inglés → español, nulo-seguro), el mapeo de secciones con publicación (ocultas si no hay nada publicado), validaciones de formularios en Zod (formatos de correo/teléfono/enlaces, límite 1.000, español obligatorio), el selector de idiomas (persistencia y cambio), los CTAs (construcción de `wa.me` a partir del número), y las anclas que responden al estado de publicación de cada sección.
- **SC-006/FR-009** es central en pruebas: en inglés, con contenido sin inglés, la portada muestra el texto en español **en el mismo diseño**, y jamás campos vacíos. Caso de prueba con las cinco secciones y distintas mezclas idioma/estado.
- **SC-002**: el `qa-tester` verifica **por cualquier vía** que no se exponga un borrador: el HTML público no contiene su texto, ni su imagen, ni su `alt`.
- El **panel no se toca visualmente**: reutiliza los componentes de F2 tal cual; F3 solo añade el acceso con `hasPermission` en `PanelLayout` y la tarjeta en `InicioPage`.
- El contexto de idioma debe alcanzar el HTML: `document.documentElement.lang` cambia entre `es` y `en`.
- Subida de imágenes: previsualización con `FileReader` + validación de tipo y tamaño **antes** de enviar; el tamaño máximo lo fija el `arquitecto` en el contrato ⟲.
- La granularidad de publicación (por sección/por elemento, decisión D-2 del 2026-10-09) debe coincidir con la del contrato del backend.
- **Verificar contra el Manual de marca** en la revisión visual: fuentes correctas (Bebas/Poppins/Playfair), proporción de color (70/20/10), logo sin efectos ni deformaciones, y el `alt` definido por el equipo en el panel.
- La portada no incluye mapa: si el equipo lo quiere en el futuro, es una mejora aparte (fuera del alcance actual).
