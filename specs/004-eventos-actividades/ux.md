# UX: F4 — Eventos y actividades

> Documento de diseño para `disenador-ux`. Fuentes: `spec.md` de F4 (aclaraciones Q1–Q6 resueltas el 2026-10-11), `ux.md` de F3 (modelo de formato, textos y patrones de panel), `ux.md` de F10 (sistema de diseño y movimiento **ya entregado**: «Marea Turquesa»), el Manual de Identidad (`resources/MANUAL DE MARCA.pdf`) y el código actual de `frontend/src/` (React 19 + Tailwind 4; componentes de F2/F3/F10 ya construidos).
> No es código: lo implementa el `dev-frontend`. Rutas y contratos los fija el `arquitecto`; aquí se proponen.
> **Decisiones de UX a revisar por el humano**: ver §9 (resumen). Cada una lleva su marca ⟲; las marcadas ✅ quedaron **resueltas por el humano el 2026-10-11** (D-1: ruta `/eventos` · D-6: permiso nuevo `categorias` · D-13: aviso sin redirección) **o cerradas por decisión firme realineada el 2026-10-11** (D-16: 409 categoría en uso · D-22: fechas sin conversión de zona).
> F4 **aplica** el sistema de F10 y los patrones de F2/F3; **no redefine** tokens, navegación, selector de idioma, modo oscuro ni el rediseño del panel (F10B).

---

## 0. Principios de diseño (heredados, aplicados a F4)

1. **Cada elemento es un dato, no decoración.** Título, fecha, hora, lugar y detalle son los protagonistas; la imagen y la categoría son accesorios opcionales. Nada se manufactura para «llenar»: un elemento sin imagen y sin categoría se ve completo (FR-020).
2. **Dos conceptos, cero confusión.** Los eventos (periódicos) y las actividades (puntuales) viven en **dos bloques separados** de la misma página (Q5); nunca se mezclan en un solo listado sin identificación (FR-001). El visitante distingue qué es periódico y qué es puntual por el bloque y el rótulo, no por deducción.
3. **La fecha es la información número uno.** Fecha escrita en palabras, localizada por idioma («viernes 17 de octubre, 7:00 p. m.» / «Friday, October 17, 7:00 PM»), en `<time datetime>` para lectores de pantalla, grande y legible. Nada de «17/10» a solas ni de formatos ambiguos (FR-009).
4. **Cercano, no institucional.** Los rótulos hablan como la iglesia habla: «Lo que viene», «Ver ubicación», «Todos los viernes». El tono del pie («Esta siempre será tu casa») es el techo, no el estándar: F4 es agenda, no emocional — su voz es clara, directa y amable.
5. **Apto para todas las edades.** Cuerpo 18 px, objetivos ≥ 44 px, iconos siempre con texto, `prefers-reduced-motion` intacto, 0 desplazamiento horizontal a 320 px (FR-023, FR-024).
6. **Solo lo publicado existe.** Sin ningún evento ni actividad publicados, la sección **no existe**: ni en la navegación, ni en el mapa del sitio, ni como página vacía (FR-013, FR-022). Un bloque sin elementos no se dibuja: no queda título sin contenido.
7. **El panel hereda F2/F3 tal cual.** Módulos, `Tabs`, `Table`, `StatusPill`, `Dialog`, `ConfirmDialog`, `Notice`, validaciones con Zod y mensajes en español: cero estilos nuevos (FR-016; el repintado es F10B y no entra aquí).

---

## 1. Tokens aplicados (la verdad vive en F3/F10; solo se citan las decisiones de F4)

| Tema | Reutilizado de | Nota específica de F4 |
|---|---|---|
| Paleta 70/20/10 | F3 §1.1 / F10 §0.1 | La página vive sobre navy con Keynotes en crema; acento turquesa para la sección activa, chips y subrayados. Turquesa pequeño como texto siempre `#0e7a66` sobre claro. |
| Tipografías | F3 §1.2 / F10 §0.2 | **Bebas** para el rótulo grande de fecha (momento de impacto, como la hora del horario); **Poppins** para todo el cuerpo; **Playfair itálica** solo si el equipo quiere elevar un texto de frecuencia... no: la frecuencia va in Poppins por claridad (⟲ D-3). F4 no usa Playfair. |
| Espaciado | F10 §0.3 (`--space-section/block/fluid`, `--pad-panel/text`) | Dos bloques separados por `--space-section`; los elementos dentro de un bloque por `--space-block`. |
| Superficies | `Card` (`solid`/`glass`, F10) | Cada evento/actividad es una `Card solid`; la imagen va a sangre dentro del marco con `padded={false}` en su zona superior. |
| Movimiento | F10 §5 (`Reveal`, tokens 600 ms / micro 180 ms) | Revelado lateral alternado por **bloque**, no por cada tarjeta (los cards usan stagger suave, tope 240 ms). Con `prefers-reduced-motion`: sin revelado, todo visible al instante. |
| Foco visible | F3 §1.4 / F10 §6 | Anillo 2 px: navy sobre claro, crema sobre navy. Nunca se oculta foco en foco-inicial de revelado (guarda de `Reveal`). |
| Chip de categoría | tokens existentes | Píldora pequeña con borde `navy-soft`, texto Poppins 0.875 rem; **sin color semántico** — la categoría no es un estado (⟲ D-2, más info en §4.1). |
| Chip de concepto | tokens existentes | La tarjeta NO repite su concepto «Evento/Actividad»: lo nombra el bloque en el que vive (§4.1). |

**Nuevos tokens: ninguno.** F4 usa exclusivamente los de F3/F10.

---

## 2. Flujo de usuario

### 2.1 Visitante: descubre lo que viene (US1, camino feliz)

1. Abre el sitio (`/`, idioma según su preferencia guardada; primera visita: español). Ve la barra de F10 con la lista de secciones de la portada… **y una entrada nueva: «Eventos y actividades»** — se dibuja solo si hay al menos un evento o actividad publicados con fecha desde hoy (FR-022; con secciones de portada vacías el pie/menú no dejan huérfanos).
2. Toca «Eventos y actividades» (barra en escritorio; item del panel del menú móvil con el mismo nombre) → desplazamiento a **`/eventos`**: página de contenido real con **H1 «Eventos y actividades»** y dos bloques: **«Próximos eventos»** y **«Próximas actividades»**.
3. En **Próximos eventos** ve cada periodicidad: título, la **próxima fecha** grande y escrita, el texto de frecuencia si existe («todos los viernes»), horario del día, lugar con enlace **«Ver ubicación»** (abre el mapa en la app o en la web, solo si el equipo cargó enlace), categoría (si hay) y detalle.
4. En **Próximas actividades** ve cada acción puntual: título, fecha y horario exactos, lugar, categoría (si hay), imagen (si hay) y detalle.
5. Toca «Ver ubicación» en un evento → abre `https://maps…` en pestaña nueva con `rel="noopener"`; el sitio queda donde estaba.
6. Termina y navega: pie de la página (mapa del sitio con el grupo **Explorar** que ahora incluye «Eventos y actividades»), selector de idioma e idioma, como en toda página de F10.
7. **Lo que ya pasó no está** (Q3): un evento cuya vigilia fue el fin de semana pasado no aparece; el elemento del día en curso sí, hasta que termine su fecha/hora (Q4).

### 2.2 Visitante: cambia a inglés (US5)

1. En la barra (o el pie), toca **English** en el selector de F10.
2. La página cambia de inmediato: la **interfaz** («Upcoming events», «See location», «Loading…») va siempre del diccionario es/en; los títulos, lugares, detalles y frecuencias **con versión en inglés** se muestran en inglés.
3. Un contenido **sin inglés** se muestra en español, con idéntico diseño y **sin ningún hueco** (patrón de F3, resuelto por el servidor; el frontend no reimplementa el fallback).
4. Las **fechas** se reescriben en formato inglés (`Intl`): "Friday, October 17, 7:00 PM"; **nunca cambia el día ni la hora reales** (el dato es el mismo; cambia la presentación).
5. La elección persiste y se recuerda entre visitas (F3/F10); `html lang` cambia a `en`.

### 2.3 Equipo: gestiona eventos (US2, camino feliz)

1. Inicia sesión → **Inicio** del panel: dos accesos nuevos en `AccesosRapidos` — **«Gestión de eventos»** (con permiso «eventos») y **«Gestión de actividades»** (con permiso «actividades»), cada uno **solo** si la cuenta lo tiene.
2. Entra a **`/panel/eventos`** (título «Gestión de eventos» + banda de ayuda): listado con tabla (desktop) / tarjetas (móvil): título, **próxima fecha**, frecuencia y número de próximas fechas, estado (`Borrador`/`Publicado`) y, si corresponde, marca **«Pasado»**.
3. Toca **«Agregar evento»** → `Dialog` con `Tabs` de idioma **Español / English (opcional)** (patrón F2/F3) y campos: título (es obligatorio), **fechas** (lista de filas fecha + hora inicio + hora fin opcionales, con «Agregar fecha» y quitar por fila), **frecuencia** (texto opcional), lugar (texto obligatorio) + **enlace de mapa** opcional, detalle (textarea, contador 2.000), organizador (opcional), categoría (`Select`, opcional) e imagen con **alt** (opcional).
4. Guarda → el evento queda **Borrador**: fila visible en el panel, invisible al público; aviso «Evento guardado en borrador…».
5. Toca **«Publicar»** en su fila (sin confirmación) → píldora verde «Publicado» + aviso «Publicado en "Eventos y actividades"». El visitante YA lo ve.
6. **Edita un evento publicado** y guarda → aviso **«Cambios guardados: ya visibles en la sección.»** (patrón F3: se publica al guardar, FR-014).
7. **Retira** («Retirar de la sección» con `ConfirmDialog`) o **elimina** («Eliminar», confirmación explícita de no-recuperable, FR-015) — en ambos casos queda registro en la auditoría de F2.

### 2.4 Equipo: gestiona actividades (US3)

La misma mecánica de §2.3, bajo **`/panel/actividades`** y permiso «actividades», con tres diferencias (Q1/Q2):
- En el formulario NO hay lista de fechas: **una sola fecha** con hora inicio y hora fin opcionales; **sin campo de frecuencia** (una actividad es puntual, Decisión 3).
- Al guardar sin escribir nada en inglés, se guarda igual (Decisión 8); el aviso inicial explica «Quedó en borrador…».
- Los rótulos dicen «Actividad» («Actividad guardada en borrador», «Publicado en la sección», etc.), nunca «it» mezclado.

### 2.5 Equipo: publica o retira por elemento (US4)

El listado del panel muestra SIEMPRE el estado real al público hoy:
- **Borrador** → nadie lo ve fuera del panel.
- **Publicado** con fecha futura → el visitante lo ve.
- **Publicado + todas sus fechas pasadas** → el visitante **no** lo ve (se retira de la vista pública, Q4); la píldora «Publicado» se mantiene en el panel y una marca de texto **«Pasado»** con leyenda explica: «Sus fechas ya pasaron: no se muestra en la sección hasta que cargues una próxima fecha.» (Q4; el panel señala el caso, no lo borra).
- La sección pública desaparece **solo** cuando **ningún** evento ni actividad está visible (todos en borrador o pasados): nav y mapa deben actualizar sin enlaces rotos (FR-022).

### 2.6 Errores del equipo

| Situación | Camino |
|---|---|
| Campo obligatorio vacío (título, fecha, lugar, detalle en español) | Nada se guarda: error **junto al campo** + resumen arriba del formulario; el foco va al primer campo invalidado (FR-015). |
| Fecha mal formada o inválida | Indicación de cómo corregirla; el input nativo de fecha evita casi todo formato manual (⟲ D-4). |
| Hora de fin anterior a la de inicio | Error en la fila de la fecha, con el horario correcto sugerido («debe ser posterior a la de inicio»); sin guardar nada (Q2). |
| Sin versión en español (idioma base) | Se bloquea al guardar y se explica el principio del español base (Decisión 8). |
| Enlace de mapa mal formado | Error con formato esperado y ejemplo; sin guardar (FR-015). |
| Límites de extensión (título, lugar, detalle) | Contador visible al escribir; al exceder el límite, mensaje «Recorta X para poder guardar». |
| Imagen inválida o carga fallida | Aviso del problema con el archivo; el formulario conserva el resto y el elemento queda sin imagen (FR-020; la sección pública no se rompe). |
| Error del sistema al guardar | Aviso genérico, sin detalles internos, **sin perder lo escrito** (los valores permanecen). |
| Operación sin permiso | Ruta protegida (`RequirePermission` con el código del módulo) + `ForbiddenPage` de F2; si el servidor lo rechaza en caliente: el mismo mensaje sin detalles («No tienes acceso a esta sección…»). |
| Duplicado exacto dentro del módulo (mismo título + fecha + lugar) | Aviso de duplicado; nada se guarda. |
| Dos personas editando el mismo elemento | Se conserva la última versión completa guardada; sin aviso de concurrencia (default de la spec, como F3). |
| Borrador por cualquier vía | Jamás se expone: ni en HTML público ni en la URL de una imagen (404 si el contenido no está publicado). |

### 2.7 Persona sin permiso (el caso desfavorable)

- Sin permiso «eventos»: no existe el acceso «Gestión de eventos» (ni en `AccesosRapidos` ni en navegación del panel); si fuerza la URL `/panel/eventos`: `RequirePermission` → `/sin-permiso` (F2, texto intacto).
- Tener «eventos» NO da la gestión de «actividades» ni al revés (FR-012, Decisión 5; confirmado por prueba cruzada en US2 esc. 8 y US3 esc. 7). Tener «eventos» o «actividades» tampoco da el acceso a «Categorías»: ese catálogo se autoriza con el permiso propio `categorias` (✅ D-6).
- La banda de ayuda nos recuerda «Solo lo publicado es visible al público» (coherente con F3 §4.3).

---

## 3. Rutas propuestas (React Router)

```
Sitio público (sin sesión):
/                       Portada de F3/F10 (intacta; F4 NO la toca)
/eventos                «Eventos y actividades» de F4 (EventsPage)
/health                 Estado del sistema (F1), fuera de la navegación pública

Panel (patrón F2/F3, sin rediseño):
/panel                  Inicio (accesos nuevos por permiso)
  /panel/eventos        RequirePermission(EVENTOS)   Módulo de eventos
  /panel/actividades    RequirePermission(ACTIVIDADES) Módulo de actividades
  /panel/categorias     RequirePermission(CATEGORIAS)  Catálogo de categorías (✅ D-6, resuelta: permiso nuevo `categorias`)
  /panel/informacion /panel/usuarios /panel/roles /panel/auditoria … (F2/F3, intactos)
```

- **`/eventos` es una ruta real, no un ancla de la portada** (Q5: la sección no pertenece a la portada). Los enlaces de la portada a `/eventos` deben incluir la base y volver a las anclas `/#quienes-somos`, `/#horario`, etc. (nota para el `dev-frontend`: la **única** fuente de enlaces `useSiteLinks()` se extiende con un ítem «Eventos y actividades» que es una **ruta**, mientras que los anclas de portada se construyen con `/#<ancla>` fuera de `/`). La lista del `<nav>` sigue siendo **una sola** en DOM y el patrón `MobileMenu` (disclosure APG, la MISMA lista reubicada) funciona sin cambios; la etiqueta congelada `nav.label` («Secciones de la portada») se conserva, no se renombra. ⟲ D-1
- **Nav de la barra**: «Eventos y actividades» se añade al final (después de las señales de portada); la barra wrap si en escritorio el vocabulario excede una fila (F10 lo prevé).
- **Mapa del sitio del pie**: el elemento entra en el grupo **Explorar** cuando hay algo publicado. No se abre Columna nueva (⟲ D-7, para no dejar un grupo de un solo item).
- **Sección oculta**: si la consulta pública de eventos/actividades no devuelve nada, el ítem no se renderiza en barra, ni en `MobileMenu`, ni en el pie (FR-022) — la misma regla de «una sola lista dinámica» de F10 (§2.3): `useSiteLinks()` decide con el dato en la caché de la consulta pública (⟲ D-23: la presencia se deduce de la respuesta pública de eventos compartida por queryClient, sin pedir el módulo dos veces).
- `RequirePermission` codes: los códigos técnicos exactos los fija el `arquitecto` con los de F2 (`eventos`, `actividades` reservados en FR-015 de F2). **El panel de categorías se autoriza con el permiso nuevo `categorias`** (décimo permiso del catálogo de F2, cambio de alcance aprobado por el humano el 2026-10-11; área "categorías"): **no** con `eventos` ni con `actividades`, que conservan sus permisos independientes (✅ D-6, resuelta).

---

## 4. Pantallas

### 4.1 Sitio público — `/eventos` (visitante; US1, US5, US6)

**Propósito**: en menos de 30 s (SC-001) el visitante sabe qué viene, cuándo es y dónde: agenda única con dos bloques separados por concepto.

**Estructura** (página de contenido bajo `PublicLayout` de F10):

1. **Enlace de salto** (_skip link_) — de F10, intacto; `<main>` de `/eventos` es su destino en esta página.
2. **H1 «Eventos y actividades»** + subtítulo de una línea (frase de interfaz): «Lo que viene: reuniones de la iglesia y acciones concretas para participar.» / («What's coming: church gatherings and concrete ways to take part.») — cadena i18n del diccionario, no campo del panel. ⟲ D-10
3. **Bloque «Próximos eventos»** (`SectionHeading` con el kicker `01` del sistema): título de bloque en Bebas. Si el bloque de actividades no tiene elementos publicados, NO se dibuja su título: la página muestra solo eventos, sin rótulo huérfano (⟲ D-5).
   Cada evento es una **`Card`** del sistema con:
   - **Categoría** (si hay): chip pequeño arriba del título (`CategoryChip`, Poppins, borde y texto grises; sin color semántico).
   - **Título** (h3, bebas en las tarjetas… Poppins 600 1.25rem como h3 de F3; el título del bloque es el Bebas jerárquico). El título del elemento NO compite con el H1 — no es bebas grande.
   - **Próxima fecha** (la carga real más cercana, escrita completa): rótulo destacado con `<time datetime="2026-10-17T19:00">` (`DateBadge`): fecha en Poppins 600 1.125rem, horario al lado en Bebas como «momento de impacto» del sistema («7:00 p. m.»). Con hora de fin: «7:00 – 9:00 p. m.».
   - **Frecuencia** (si hay): texto «Todos los viernes» en Poppins itálica gris, junto a la fecha (resuelto por el par es/en con fallback al español).
   - **Más próximas fechas** (si el evento tiene más de una fecha futura cargada): las fechas restantes (solo las futuras) en una lista compacta con `<time>` por cada una, mismas reglas de formato; todas a la vista, sin desplegar ni plegar (nada queda escondido en la tarjeta). ⟲ D-11
   - **Lugar**: texto del lugar (obligatorio) + si hay enlace de mapa: `PublicButton` variante secundaria/text-link con icono y **texto** «Ver ubicación» (juntos, con `target="_blank"` + `rel="noopener"`).
   - **Detalle**: texto plano multilínea con saltos preservados, cuerpo Poppins 18 px; límite paneled de 2.000 no requiere `Ver más` (⟲ D-3, diseño sin recortes en la sección pública).
   - **Imagen** (si hay): en la parte superior a sangre del marco (16:9, `object-cover`, recorte al centro, con `alt` del equipo obligatorio en español y en inglés si el contenido tiene EN; ⟲ D-9). Sin imagen: la tarjeta no muestra hueco (`Card` sin placeholder de imagen, sin «reserva visual»).
4. **Bloque «Próximas actividades»** (`SectionHeading` con el kicker `02`): mismas reglas de `Card`, con tres diferencias: frecuencia NO existe; el rótulo de fecha muestra **una única fecha**; el horario puede omitirse por completo (actividad sin hora). ⟲ D-12
5. **Pie de F10 intacto**: mapa del sitio incluye «Eventos y actividades» en **Explorar** (solo si la sección está publicada); selector de idioma en la franja; tema oscuro funciona en ambos bloques.

**Propuesta wireframe desktop y móvil** (esquema de referencia, no maqueta):

```
DESKTOP
[ BARRA F10: 🌱 Simiente Santa · Quiénes somos · Horario · WhatsApp · Contacto · Redes · Eventos y actividades | Español/English | 🌙 ]
/self scrolled → vidrio* + blur
══════════════════════════════════════════════════════════════════
 EVENTOS Y ACTIVIDADES  (h1, Bebas; panel de vidrio largo)
 Subtítulo de una línea (gris-claro Poppins)
        ┐ --space-section
  01 PRÓXIMOS EVENTOS (SectionHeading + acento turquesa)
  ┌─ Card ─────────────────┐  ┌─ Card ─────────────────┐
  │ [imagen a sangre 16:9] │  │ (sin imagen: no hueco) │
  │ acción social          │  │ noche de oración       │
  │ Importación de comida  │  │ Vigilia de viernes     │
  │ viernes 17 de oct.     │  │ viernes 24 de oct.     │
  │ 7:00 – 9:00 p. m.      │  │ 9:00 p. m.  ▸ más fechas│
  │ «todos los viernes»    │  │ …                      │
  │ 📍 Templo central · Ver ubicación · 🗂 Detalle    │
  │ (detalle Poppins 18px, saltos preservados)      │
  └────────────────────────┘  └────────────────────────┘
  … grilla 2 col desktop, 1 col móvil …
        ┐ --space-section
  02 PRÓXIMAS ACTIVIDADES (igual, sin frecuencia; fecha única)
  ┌─ Card ─────────────────┐
  │ repartir de comida ·   │
  │ sábado 18 de oct.      │
  │ 8:00 – 11:00 a. m.     │
  │ 📍 Comedor comunal · Ver ubicación · (sin categoría) │
  └────────────────────────┘
        ┐ --space-section
  PIE DE F10 (mapa del sitio: Explorar incluye «Eventos y actividades»)
══════════════════════════════════════════════════════════════════

MOBILE (≤ 860 px) (columna única, según F10: cards a lo ancho)
[ SS Simiente Santa · Español/English · 🌙 · ☰ ]
→ ☰ abre panel de menú (la MISMA lista, con Eventos y actividades)
grilla de cards 1 col, imagen arriba 16:9, todo el resto apilado
```

**Estados de la página pública**:

| Estado | Qué ve el visitante |
|---|---|
| Carga (primera) | Esqueletos: H1 + 2 bloques de `Card` fantasma con medidas reales (`SectionSkeleton`, F10); la barra, el idioma y el tema ya visibles; nunca un «brinco» de layout al llegar. |
| Éxito con ambos bloques | Sección completa con marca y cards del sistema. |
| Éxito con un solo concepto | Solo el bloque con contenido se dibuja: el otro rótulo no existe (cero hueco, cero «nada por aquí») ⟲ D-5. |
| **Sección oculta** (0 publicados, o todos pasados/en borrador) | La **ruta `/eventos` no se enlaza en ninguna parte** (barra, menú móvil y pie). Si alguien abre `/eventos` directamente, ve la **página mínima** de D-13: H1 + aviso estándar de interfaz («Ahora mismo no hay nada publicado en esta sección. Vuelve pronto.» / «There's nothing published here right now. Check back soon.») — nav y pie no enlazan a la sección, y jamás aparecen datos reales ni borradores (FR-022, SC-012). |
| Error de carga sin nada en caché | H1 + mensaje bilingüe claro en un panel de vidrio: «No pudimos cargar los eventos y actividades. Revisa tu conexión y vuelve a intentarlo.» + botón «Volver a intentar» (patrón F3/F10). |
| Error con caché previa | Se mantiene el último contenido válido y abajo del bloque un aviso discreto: «No pudimos actualizar; ves la última versión disponible.» (F3 §4.1). |
| Imagen rota o sin imagen | La tarjeta conserva título, fecha, lugar y detalle; `onError` retira la imagen de la tarjeta sin transformar el layout (sin «quiebra»). |
| Fallback es→en (US5) | El contenido sin inglés se muestra en español, idéntico diseño, SIN marca de «sin traducir» (FR-008). |
| Elemento del día en curso | Aparece entre lo próximo **mientras su hora no haya pasado**, con la regla de retiro exacta (R4-2/SC-009): hoy **sin** hora de fin → visible todo el día; hoy con hora de fin (`endTime`) **ya terminada** → **fuera** de «lo que viene» (se retira al terminar su fecha/hora); mañana en adelante → visible. |
| Fecha pasada en el contenido | El sistema puede guardarlo en el panel pero NO lo muestra aquí (Q3). Un evento con TODAS sus fechas pasadas no aparece, pero SÍ se conserva en el panel. |
| Duplicado de concepto | No existe duplicado de imagen ni de fecha: cada elemento es distinto; el orden dentro de un mismo bloque resuelve el empate de manera estable (ver §9, D-14). |
| `prefers-reduced-motion` | Sin revelado, todo visible al instante, contenido íntegro (FR-009 de F10, incondicional). |
| Borrador | Jamás visible en esta página ni en ninguna respuesta pública (SC-002, cualquier vía). |

**Responsivo**:
- **Móvil (< 640 px)**: grilla de 1 columna, cards a todo ancho; `Card padded` generoso; imagen 16:9 arriba; fecha en Poppins 600, horario Bebas (misma jerarquía que desktop, más compacta). Sin desplazamiento horizontal a 320 px (SC-007).
- **Tableta (≥ 640 px)**: 2 columnas de cards; el H1 y los rótulos de bloque crecen con `clamp`.
- **Escritorio (≥ 1024 px)**: `max-w-6xl`, cards con más aire (`--space-block`), los CTAs de «Ver ubicación» en línea junto al lugar; bloques separados con `--space-section`.

### 4.2 Panel — Inicio (cambio menor en F2)

`AccesosRapidos` añade dos tarjetas con el patrón ícono + texto existente:
- **«Gestión de eventos»** (con permiso `eventos`) → `/panel/eventos`
- **«Gestión de actividades»** (con permiso `actividades`) → `/panel/actividades`
- **«Categorías»** (con permiso `categorias`) → `/panel/categorias` (✅ D-6, resuelta: permiso nuevo `categorias`).
Una cuenta con solo «eventos» NO ve el acceso de actividades: los módulos son independientes (FR-012, Decisión 5). Ni «eventos» ni «actividades» implican el permiso «categorias» ni al revés.

> **Alcance de F4 (cierra I4):** estas tres tarjetas de `AccesosRapidos` en `InicioPage` —y la de «Categorías»— **sí forman parte de F4** y las cubre la tarea de panel (extiende su alcance a `features/panel/components/AccesosRapidos.tsx`, con prueba de que cada tarjeta se muestra solo con su permiso; `hasPermission`, patrón F10 — §10).

### 4.3 Panel — Gestión de eventos — `/panel/eventos` (US2, US4)

**Propósito**: administrar los encuentros periódicos sin depender de nadie; ver de un vistazo qué es público y qué no.

- Título «Gestión de eventos» + banda de ayuda: «Aquí administras los encuentros periódicos de la iglesia (p. ej., noche de aposentos, vigilia). Solo lo publicado y con fecha futura es visible al público.»
- **Listado** (`Table` desktop / tarjetas apiladas en móvil, patrón F2): columnas **Título**, **Próxima fecha** (escrita y con hora corta; si no tiene fecha futura: «—» con la marca Pasado), **Frecuencia / Fechas futuras** («Todos los viernes · 3 fechas futuras»), **Estado** (`StatusPill`), **Acciones** (`Editar` · `Publicar`/`Retirar de la sección` · `Eliminar`).
- **Píldoras**: `StatusPill` mantiene sus dos valores (`Borrador`/`Publicado`). El estado «Pasado» NO es un cuarto estado de publicación: cuando un elemento publicado tiene todas sus fechas pasadas, la fila conserva su píldora «Publicado» y añade una **etiqueta de texto «Pasado»** con la leyenda «Sus fechas ya pasaron: el público no lo ve hasta que cargues una próxima fecha.» (⟲ D-8: separa estado de publicación de visibilidad).
- **Agregar evento** → `Dialog` con `EventForm` (§4.4).
- **Eliminar**: `ConfirmDialog` danger con descripción explícita «"{título}" se eliminará para siempre. Esto no se puede deshacer.» (patrón F2). La eliminación queda en auditoría (FR-017).

### 4.4 Panel — Formulario de evento (`EventForm`, en `Dialog`)

Patrón F3 ServiceForm						   (TabItem + Zod). Campos:

- **Tabs de idioma**: **Español** (obligatorio) / **English (opcional)**. En el tab ES se rellena todo lo obligatorio; en EN, títulos, lugar, detalle, frecuencia y **alt de imagen**, todos en inglés; **el botón guarda si EN está vacío** y lo indica con la ayuda de F3 («Si lo dejas vacío, el público verá este contenido en su versión en español.»).
- **Título** (`Field`, es obligatorio / en opcional).
- **Frecuencia** (texto libre opcional: «Todos los viernes»).
- **Lista de fechas (`DatesEditor`)**: filas de `[fecha] [hora inicio opcional] [hora fin opcional]` con «Agregar» y quitar por fila. Inputs nativos `date` y `time` de navegador (⟲ D-4); Zod valida el formato y la coherencia de hora fin > hora inicio por fila (FR-015). Guardar exige **al menos una fila** (FR-015; puede ser pasada: no aparece al público).
- **Lugar** (`Field` obligatorio, es/en) y **enlace de mapa** (`Field` opcional, es/en; valida formato URL; ayuda: «Pega el enlace de Google Maps o similar; por ejemplo: https://maps.google.com/?q=…»). No se verifica el destino (Out of Scope).
- **Detalle** (`TextArea` reutilizable, contador vivo `{n}/2.000`, es/en; saltos de línea simples).
- **Organizador/encargado** (opcional, es/en): texto libre.
- **Categoría** (`Select`, opcional, con la primera opción «— Sin categoría —»; las opciones van en español siempre en el panel; ⟲ D-15: si se quiere en el catálogo un título en inglés, se agrega en la gestión de categorías, §4.6).
- **Imagen** (`ImageUploader` de F3): previsualización inmediata, `alt` obligatorio en español (en opcional) y reglas ya documentadas. Si falla carga: aviso y el resto del formulario queda íntegro.
- **Guardar**: `Button loading` («Guardando…»); el resultado envía el aviso del listado (§5.c, patrón F3). El modal no se cierra hasta terminar; `Esc` no descarta lo escrito sin aviso (D-18).

### 4.5 Panel — Gestión de actividades — `/panel/actividades` (US3, US4)

Simétrico a §4.3 con ActividadForm:
- La tabla/columnas cambian: **Título**, **Fecha** (solo una), **Horario** («8:00 – 11:00 a. m.» / «—» si no tiene), **Estado**, **Acciones**.
- El formulario no tiene ni lista de fechas ni campo de frecuencia: **Fecha única** + hora inicio / hora fin opcionales (Q1/Q2; Decisión 3).
- Los mensajes del módulo dicen «Actividad» en lugar de «Evento» (misma mecánica, textos distintos; viven en `messages.ts`).

### 4.6 Panel — Catálogo de categorías — `/panel/categorias` (Q6; permiso `categorias` — ✅ D-6, resuelta)

**Propósito**: administrar el catálogo compartido de etiquetas.
- Listado (`Table`): **Nombre** (es/en, con en pendiente si falta), **Cuántos elementos la usan** («3 eventos · 1 actividad» — ayuda para eliminar a conciencia), **Acciones**: `Editar` · `Eliminar`.
- **Agregar categoría** → `Dialog` con nombre es (obligatorio) + en (opcional). La creación/edición de una categoría no requiere publicar (el catálogo no es contenido público: solo etiqueta que viaja con cada elemento).
- **Los elementos guardan el nombre calculado en el momento**: si se renombra «Acción social» a «Ayuda social», los elementos la muestran con el nombre actual (relación por referencia).
- **Eliminar categorías**: `ConfirmDialog` danger. **Decisión firme (⟩ plan/contrato): una categoría en uso NO se puede borrar** — la API la rechaza con **409 (conflict)** por la FK `RESTRICT`; la interfaz nunca ofrece desligarla ni continuar: el diálogo lo explica con la cuenta («Está en uso por N elementos: quítala primero de los elementos») y su única acción es cerrar. ✅ D-16 (cerrada).
- El catálogo inicial sembrado: acción social, predicación, adoración, entrevista, noche de oración, otro.

### 4.7 Pantalla de estado vacío (cada módulo de panel)

- «Todavía no hay {eventos/actividades}. Agrega el primero para que el público lo vea.» + `Button` de «Agregar {evento/actividad}».
- Si hay elementos pero **ninguno publicado**: aviso destacado «Nada está publicado ahora: el público no ve esta sección hasta que publiques al menos un {evento/actividad}.»

### 4.8 Texto largo, muchos elementos, sin categoría, sin imagen, sin fecha futura

| Caso | Cómo se ve |
|---|---|
| Detalle de 2.000 caracteres | La página pública lo muestra completo (sin `Ver más`, ⟲ D-3); móvil y escritorio sin desbordes (Poppins 18, ancho fluido). |
| Muchos elementos próximos (10+) | La página los lista todos, 1 col móvil / 2 col tablet+ (sin paginación, default de la spec, ⟲ D-17). |
| Elemento sin imagen | Tarjeta con solo textos; no hueco, no reserva, no sombra distinta. |
| Elemento sin categoría | Se omite el `CategoryChip`; no queda línea vacía. |
| Evento sin fecha futura | No se muestra al público; en el panel, la fila lleva la marca **«Pasado»** con la leyenda (§4.3) — señala al equipo que falta cargar la próxima fecha, sin borrar nada (Q4). |
| Frecuencia sin fechas futuras | No aparece nada en lo público; la frecuencia sigue visible en el panel. |
| Texto de frecuencia largo | Se muestra en una línea con `text-wrap: balance`; conserva tal cual tildes, ñ, ¿ ¡ y emojis; el sistema no lo recorta ni lo transforma. |

### 4.9 Cross-check con F2 (área de auditoría)

- **Auditoría de F2** registra cada operación con los textos «Creó/Modificó/Publicó/Retiró/Eliminó el {evento/actividad} "{título}"» + resultado; el módulo de auditoría NO cambia su UI (FR-017). El `qa-tester` verifica las entradas para cada operación.
- El registro es solo lectura, como en F2 (0 ediciones/borrados; SC-013).

---

## 5. Estados por vista con datos (panel)

### a. Vista general del módulo (`/panel/eventos`, `/panel/actividades`)

| Estado | Qué ve |
|---|---|
| Cargando | `EmptyState` kind=loading («Cargando eventos…»); barra de acciones ya visible. |
| Vacío (sin ninguno) | `EmptyState` con texto y botón de agregar (§4.7). |
| Vacío con todo borrador | Aviso destacado «Nada está publicado…» + listado completo. |
| Vacío con todo pasado | Aviso «El público no ve esta sección porque sus fechas ya pasaron; carga la próxima fecha.» |
| Error de carga sin nada | «No se pudo cargar los eventos.» + `Button` «Reintentar» (Notice error, `role="alert"`). |
| Error con contenido en caché | Muestra el último listado válido + Notice discreto sobre la actualización. |
| Éxito | Listado con píldoras, marcas de pasado y acciones. |
| Sin permiso | Ruta protegida: no se llega; si el servidor lo rechaza, ForbiddenPage / mensaje de F2. |
| Éxito de operación | Notice temporal verde («Evento creado.» / «Publicado en la sección.» / «Retirado de la sección. Estará disponible como borrador…» / «Eliminado.»), patrón F3. |

### b. Los formularios (`EventForm`, `ActivityForm`, `CategoryForm`)

- Precargado al editar; vacío al crear. Los campos de «English» en su propia pestaña con la marca opcional.
- Guardando: `Button loading` («Guardando…»); modal sin cerrarse; el `Esc` no cierra sin aviso si hay cambios (patrón F2/F3 ⟲: en F3 el modal se cierra con `Esc` sin aviso; aquí, con lista de fechas, es mejor protegido — dev-frontend decide según `Dialog` actual; ⟲ D-18).
- Éxito: cierre + Notice en el listado («Evento creado.», o «Cambios guardados: ya visibles en la sección.» si estaba publicado).
- Errores de campo (con los textos del §7), junto al campo y resumen arriba (patrón F2).
- Modal con `Dialog` de F2: trampa de foco y devolución del foco a la fila que lo abrió; el `Esc` respeta D-18 (no descarta cambios sin aviso).

### c. Publicar / retirar / eliminar

- **Publicar**: acción de fila, sin confirmación; píldora cambia a «Publicado»; el foco se queda en la fila.
- **Retirar**: `ConfirmDialog` suave: «"{nombre}" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.» Botones «Retirar» (danger) / «Dejar como está».
- **Eliminar**: `ConfirmDialog` danger: «Eliminar "{nombre}", ¿verdad? Esto no se puede deshacer.» Botones «Eliminar» / «Cancelar» — texto más explícito que el de F3 porque no hay papelera.
- **Auto-publicar al guardar** (elemento ya publicado): Notice «Cambios guardados: ya visibles en la sección.»

---

## 6. Componentes React

> Nombres en inglés (coherente con el repo). El listado final de tareas lo fija el `arquitecto`; aquí es la propuesta con responsabilidad. **Reutilizar antes de crear**: casi todo existe.

### 6.1 Reutilizados de F2/F3 (sin cambios o con extensión mínima)

| Componente | Extensión pequeña ⟲ | Uso |
|---|---|---|
| `Button` / `PublicButton` | — | Acciones de panel y «Ver ubicación» público. |
| `Field`, `Select`, `TextArea`, `Notice` | — | Formularios y avisos. |
| `Dialog` / `ConfirmDialog` | — | Alta/edición, retirar, eliminar, eliminar categoría. |
| `EmptyState` | — | Estados vacíos y de carga. |
| `StatusPill` | — (sin cambios: la marca «Pasado» es una etiqueta aparte, D-8) | Estados borrador/publicado en el panel; junto a la píldora, la etiqueta «Pasado» cuando corresponda. |
| `Table` | — | Listados de eventos, actividades, categorías. |
| `Tabs` | — | Pestañas de idioma es/en del formulario (patrón F3). |
| `ImageUploader` | — | Imagen + `alt` es/en del elemento. |
| `PublishControls` | — | Cambiar publicar/retirar por elemento. |
| `Pagination` | **no se usa** ⟲ D-17: los listados públicos y de panel muestran todos los elementos; una paginación sería un ajuste menor si el humano la pide más adelante. | — |
| `DateRangeFilter` | **no se usa** en F4: el alcance temporal lo decide el sistema (desde hoy) — no hay filtro de rango en el público ni en el panel ⟲ D-19. | — |

### 6.2 Reutilizados de F10 (sin cambios)

| Componente | Uso |
|---|---|
| `PublicLayout`, `SiteHeader`, `PublicNav`, `MobileMenu`, `useSiteLinks` | Navegación y la **única lista de enlaces** del sitio; se extiende el array con un **ítem de ruta** hacia `/eventos` (⟲ D-1). |
| `LanguageSwitcher`, `LanguageProvider`/`useLanguage`, i18n (`es.ts`/`en.ts`/`messages.ts`) | Selector y cadenas; nada nuevo de idioma. |
| `ThemeToggle`/`useTheme` | Modo claro/oscuro en la página nueva. |
| `SectionHeading` | Título de ambos bloques (01/02). |
| `Card` | Superficie de cada evento/actividad (`variant="solid"`), la imagen en su marco superior con `padded={false}`. |
| `Reveal` / `RevealGroup` / `useReveal` | Entrada lateral alternada por bloques; guarda de foco (elemento a revelar con foco se hace visible). |
| `SectionSkeleton` | Cargando la página. |
| `FooterSiteMap` | Grupo **Explorar** incluye «Eventos y actividades» cuando existe; nunca lo inventa. |
| `GradientBlobs`, `BrandLogo`, `SocialIcon` | Fondo y chrome intactos. |

### 6.3 Nuevos del sitio público (`frontend/src/features/publico/eventos/`)

| Componente / hook | Responsabilidad |
|---|---|
| `EventsPage` | Composición de H1 + dos bloques a partir de `useUpcomingEvents`; decide qué bloques se dibujan (solo los con contenido), estados de carga/error/oculta. |
| `EventsSection` | Bloque «Próximos eventos»: título `SectionHeading` + grilla de tarjetas + `RevealGroup`. |
| `ActivitiesSection` | Bloque «Próximas actividades» (misma maqueta, sin frecuencia). |
| `EventCard` | Tarjeta de un evento: Imagen opcional, `CategoryChip`, título, `DateBadge`, lista de fechas restantes, frecuencia, lugar + `MapLink`, detalle. |
| `ActivityCard` | Tarjeta de una actividad: Imagen opcional, chip, título, `DateBadge` (una fecha), lugar + `MapLink`, detalle. |
| `CategoryChip` | Etiqueta de categoría (⟲ D-2). |
| `DateBadge` | Fecha + horario **localizados** con `<time datetime>` (formato por idioma con `Intl`; única implementación compartida; no inventa formatos a mano). |
| `MapLink` | Enlace exterior (`target="_blank" rel="noopener"`) con texto «Ver ubicación» (icono + texto) y `aria-label` completo (D-20). |
| `HiddenNotice` | Aviso de sección oculta (H1 + mensaje de interfaz, sin datos; D-13). |
| `useUpcomingEvents` | TanStack Query: consulta pública de eventos+actividades próximos (contrato del `arquitecto`); alimenta la página y la presencia en `useSiteLinks`. |
| `features/publico/eventos/formatoFechas.ts` | Escribir en palabras las fechas (`Intl.DateTimeFormat` por idioma): «viernes 17 de octubre, 7:00 p. m.» / «Friday, October 17, 7:00 PM»; horarios en 12 h según el locale. **Construye la fecha desde sus partes (`YYYY-MM-DD` + `HH:MM`), sin conversión de zona horaria** (nunca `new Date(string)`; D-22). |

### 6.4 Nuevos del panel (`frontend/src/features/eventos/`)

| Componente / hook | Responsabilidad |
|---|---|
| `EventsPage` (panel, en `features/eventos/pages/`) | Panel de gestión: título, banda de ayuda, listado + estados (§5). |
| `ActivitiesPage` | Simétrico para actividades. |
| `CategoriesPage` | Catálogo de categorías (listado + agregar/editar/eliminar, §4.6). |
| `EventForm` (`Dialog`) | Alta/edición de evento con lista de fechas, dos lenguas, todo el patrón F2/F3 (⟲ D-18). |
| `ActivityForm` (`Dialog`) | Alta/edición de actividad puntual (fecha única, sin frecuencia). |
| `CategoryForm` (`Dialog`) | Alta/edición de categoría (nombre es obligatorio / en opcional). |
| `DatesEditor` | Editor de lista de fechas del EventForm (agregar/quitar filas, inputs nativos, Zod por fila y cruce fin > inicio). |
| `EventsList` / `ActivitiesList` | Tablas de listado con `StatusPill`, marca «Pasado» y acciones (patrón `ServicesList`). |
| `PublishControls` reutilizable | Acciones publicar/retirar por elemento (F3). |
| `features/eventos/messages.ts` | Textos del módulo en español (FR-016) — catálogo del §7.1. |
| Hooks TanStack Query | `useEvents`, `useCreateEvent`, `useUpdateEvent`, `useDeleteEvent` · `useActivities`, `useCreateActivity`, `useUpdateActivity`, `useDeleteActivity` · `useEventCategories`, `useCreateCategory`, `useUpdateCategory`, `useDeleteCategory` · `usePublishEvent`/`usePublishActivity` (o reuso del patrón `useUpdate*`). Ningún `fetch` en componentes (regla F2). |

**Pruebas esperadas** (Vitest + Testing Library + MSW): composición pública con bloques según estado (0 bloques vacíos, sección oculta actualiza nav/mapa sin huérfanos), orden cronológico con empate estable (D-14), fallback de contenido sin inglés, formato de fecha por idioma sin cambiar el valor real, `prefers-reduced-motion`, validaciones de eventos (≥1 fecha, fin > inicio, español obligatorio, límites, enlace de mapa, duplicados), ciclos de publicación por elemento, permisos cruzados eventos/actividades/categorías, y **0 borradores filtrados en cualquier vía** (incluso imágenes: 404).

---

## 7. Catálogo de textos

### 7.1 Panel (español; FR-016) — añadidos a los de F2/F3

| Situación | Texto |
|---|---|
| Guardado | «Cambios guardados.» |
| Guardado con auto-publicación | «Cambios guardados: ya visibles en la sección.» |
| Publicar | «Publicado en "Eventos y actividades".» |
| Retirar | «Retirado de la sección. Estará disponible como borrador para publicarlo cuando quieras.» |
| Crear evento/actividad | «{Evento/Actividad} creado.» + aviso la primera vez: «Quedó en borrador: no lo verá el público hasta que lo publiques.» |
| Eliminar (confirmación) | «"{nombre}" se eliminará para siempre. Esto no se puede deshacer.» |
| Eliminado | «Eliminado.» |
| Campo obligatorio vacío | «Escribe {el título / el lugar / el detalle…}.» |
| Sin fecha (evento) | «Agrega al menos una fecha para el evento.» |
| Sin fecha (actividad) | «Escoge la fecha de la actividad.» |
| Fecha mal formada | «Ese día no existe o no está bien escrito. Escógelo del calendario.» |
| Hora fin no válida | «La hora de fin debe ser posterior a la hora de inicio.» |
| Sin versión en español | «El contenido en español es obligatorio: es el idioma base del sitio.» |
| Enlace de mapa mal formado | «Ese enlace no parece una dirección web. Pégalo completo; por ejemplo: https://maps.google.com/?q=…» |
| Límite de detalle | «{n} de 2.000 caracteres. Recorta {m} para poder guardar.» |
| Límite de título/lugar | «Ese texto es demasiado largo (máximo {límite}).» |
| Imagen inválida | «Ese archivo no es una imagen válida. Usa JPG, PNG o WebP de menos de {tamaño} y prueba de nuevo.» (F3/M3) |
| Duplicado dentro del módulo | «Ya existe un {evento/actividad} igual (mismo título, misma fecha y mismo lugar). Edita el que ya tienes o cambia algún dato.» |
| Categoría en uso al eliminarla | **`409 conflict`: no se puede borrar una categoría en uso.** Texto: «Está en uso por {n} elementos: quítala de los elementos primero.» (no se desliga; decisión firme del plan/contrato — D-16 cerrada) |
| Categoría duplicada | «Ya existe una categoría con ese nombre.» |
| Evento pasado (leyenda) | «Sus fechas ya pasaron: el público no lo ve hasta que cargues una próxima fecha.» |
| Sección sin publicados | «Nada está publicado ahora: el público no ve esta sección hasta que publiques al menos un {evento/actividad}.» |
| Error del sistema al guardar | «No se pudo completar la operación. Tus escritos están a salvo; vuelve a intentarlo en unos minutos.» |
| Sin permiso (servidor) | «No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.» (F2) |
| Ayuda «English (opcional)» | «Si lo dejas vacío, el público verá este contenido en su versión en español.» |
| Ayuda de fechas (evento) | «Cada fila es una fecha concreta en que se repite el encuentro. La sección siempre muestra la más cercana.» |
| Ayuda de imagen | «Opcional. Añade el texto alternativo para personas que no la pueden ver. El logo de la iglesia respeta las reglas de marca.» |

### 7.2 Sitio público — cadenas de interfaz es/en

> El contenido de la iglesia no va aquí: título, lugar, detalle, frecuencia, categoría y `alt` de imagen vienen del panel (es → fallback). Estas son las palabras de la interfaz.

| Clave | Español | English |
|---|---|---|
| `nav.sections.events` | Eventos y actividades | Events & activities |
| `events.page.title` | Eventos y actividades | Events & activities |
| `events.page.subtitle` | Lo que viene: reuniones de la iglesia y acciones concretas en las que puedes participar. | What's coming: church gatherings and concrete ways to take part. |
| `events.block.scheduled` | Próximos eventos | Upcoming events |
| `events.block.activities` | Próximas actividades | Upcoming activities |
| `events.location.label` | Lugar | Location |
| `events.location.link` | Ver ubicación | See location |
| `events.date.next` | Próxima fecha | Next date |
| `events.date.more` | Más fechas | More dates |
| `events.empty.hidden` | «Ahora mismo no hay nada publicado en esta sección. Vuelve pronto.» — **copia canónica en §4.1** (D-13); aquí solo se referencia. | «There's nothing published here right now. Check back soon.» |
| `events.category.aria` | Categoría: {nombre} | Category: {name} |
| `events.loading` | Cargando… | Loading… |
| `events.error.load` | No pudimos cargar los eventos y actividades. Revisa tu conexión y vuelve a intentarlo. | We couldn't load the events and activities. Check your connection and try again. |
| `events.error.retry` | Volver a intentar | Try again |
| `events.error.partial` | No pudimos actualizar; ves la última versión disponible. | We couldn't refresh; you're seeing the latest available version. |
| `a11y.events.heading` | Eventos y actividades de la iglesia | Church events and activities |

> El `alt` de las imágenes es obligatorio en el panel (español, inglés opcional): no existe una cadena de fallback automática en el sitio público.

Fechas: **se escriben desde el dato**, no desde el diccionario (`formatoFechas.ts`, §6.3): «viernes 17 de octubre, 7:00 p. m.» en español y «Friday, October 17, 7:00 PM» en inglés; siempre envueltas en `<time datetime="…">`; reloj de 12 h en español (`a. m.`/`p. m.` con espacio, `m.` para el mediodía) y AM/PM en inglés según `Intl`.

---

## 8. Accesibilidad (WCAG 2.1 AA) e internacionalización

### 8.1 Accesibilidad

- **Teclado**: skip link → logo → navegación (enlaces de barra; menú móvil con el patrón *disclosure* de F10) → idioma → tema → contenido → pie. Los «Ver ubicación» abren en pestaña nueva con **nombre accesible completo** («Ver ubicación de "{título}"» / «See location for "{title}"», D-20), para que un lector de pantalla los entienda sin depender del contexto del DOM. En el panel, `Esc` respeta D-18 (no descarta cambios sin aviso).
- **Foco visible** (anillo de 2 px contrastado según tema) en todo interactivo, incluidos los enlaces a mapa. El contenido nunca depende de una animación: un elemento aún no revelado que RECIBE foco se hace visible de inmediato (guarda de `Reveal`, F10).
- **Semántica**: un `<h1>` por página («Eventos y actividades»); los títulos de bloque `h2` por `SectionHeading`; cada elemento de tarjeta es un `article` (o `li` en la lista) con `h3`; las listas de fechas usan `<ul>` con `<time datetime>` por fecha; sin tablas de datos públicas (una página de agenda no es tabla — la diferencia de presentación la resuelven bloques+grilla).
- **Fechas para lectores de pantalla**: la fecha escrita se lee como texto (es/en), con el **complemento** de `<time datetime="YYYY-MM-DDTHH:MM">` para transportar/interpretar el dato; el orden del DOM es el cronológico que se promete (Q3).
- **Más que color**: lo que distingue un evento de una actividad no es el color — son dos bloques con títulos distintos; las categorías no poseen color semántico (⟲ D-2). Píldoras de estado en el panel llevan SIEMPRE texto.
- **Imágenes**: `alt` es/en gestionado por el equipo (obligatorio en español, en opcional); si la imagen falla, la tarjeta no se quiebra (un `onError` retira la imagen con el mínimo cambio de layout). El logo solo aparece en el chrome del sitio (F10) y respeta las reglas de marca del Manual.
- **Contraste y letra**: cuerpo 18px, rótulos de bloque Bebas con contraste AA en ambos temas; el turquesa pequeño como texto existe solo como `#0e7a66 sobre claro`; los chips con texto gris navy y borde visible.
- **Zoom/200 % y 320 px**: diseño fluido en `rem` con clamp; nada fija px para texto; 0 desplazamiento horizontal (SC-007).
- **Objetivos táctiles ≥ 44 px**: `PublicButton` de «Ver ubicación» (48 px de alto en móvil), chips de categoría, ítems del menú móvil, cambiador de idioma y tema; la tarjeta en sí no es clicable (solo lo son sus enlaces y botones): ningún control por debajo de 44 px.
- **Avisos**: el cambio de idioma mantiene el anuncio único con `role="status"` de F10; los del panel reutilizan `Notice` con su rol correcto.
- **`prefers-reduced-motion`**: sin revelado ni stagger — contenido visible al instante; F10 lo garantiza (§5.6), y F4 lo respeta (no añade movimiento nuevo).
- **Idioma del documento**: `html lang` conmutado por F3/F10; el contenido en fallback (escrita en español dentro de una página en inglés) va con `lang="es"` en su envoltorio (buena práctica heredada de F3) ⟲ D-21: el detalle de implementación lo fija el `arquitecto` según lo que entrega el servidor.

### 8.2 Internacionalización (Decisión 6 y Decisión 8)

- **Interfaz** (siempre bilingüe): diccionario `es.ts`/`en.ts` de F3, sin librería externa; `LanguageProvider` y `localStorage` intactos; primera visita en español, sin auto-detección.
- **Contenido** (por elemento): español obligatorio, inglés opcional; el fallback es→es lo resuelve el servidor en la entrega (FR-008); el frontend nunca muestra un campo vacío ni traducción automática.
- **Fechas**: mismo dato, distinta presentación por idioma; los helpers de `formatoFechas.ts` (`Intl.DateTimeFormat('es'|'en')`) construyen la fecha **a partir de sus partes (`YYYY-MM-DD` + hora), sin conversión de zona horaria** — el dato de fecha es el **del lugar** y **nunca se recalcula ni se convierte** (R4-3; contrato `Date`/`ClockTime` «no es un timestamp: no se convierte de zona horaria»). Prohibido `new Date("YYYY-MM-DD")` con `Intl` sobre la zona del navegador: lo parsea como UTC y desplaza un día en zonas UTC−x (incluida `America/Costa_Rica`, UTC−6), lo que FR-009 prohibe. Nota ✅ D-22 (cerrada): sin conversión, hoy y siempre; si algún día se quisiera fijar la zona horaria de la iglesia, sería un cambio de contrato del backend, no del diseño.
- **El panel NO cambia de idioma** (siempre en español, FR-016); solo la página pública tiene el `LanguageSwitcher`.

---

## 9. Decisiones de UX para revisar por el humano (resumen)

| # | Decisión | Alternativa |
|---|---|---|
| D-1 ✅ | `/eventos` es una **ruta real** (confirmada por el humano el 2026-10-11); la nav mezcla anclas de portada y un enlace de ruta en la **única** lista de enlaces de F10. | Dos entradas separadas («Próximos eventos» / «Próximas actividades», descartado Q5) o redirigir anclas internas. |
| D-2 ⟲ | **Chip de categoría neutro** (texto + borde, sin color semántico): la categoría orienta de qué se trata pero no es un estado. | Colores por categoría (confusos con estados y tema). |
| D-3 ⟲ | **Detalle completo en la tarjeta pública** (sin «Ver más»): la agenda es el corazón del sitio y recortar obliga a un tap. | Mostrar una primera porción con «Ver más» expandible. |
| D-4 ⟲ | **Inputs nativos `date`/`time` de navegador** para fechas y horas en el panel: 0 errores de parseo y compatible con móviles. | Selects/combos propios o datos de fecha con texto libre. |
| D-5 ⟲ | Un bloque **sin elementos no se dibuja** (ni rótulo): si solo hay eventos publicados, la página no muestra «Próximas actividades». | Mostrar siempre ambos rótulos con un «no hay nada» (violaría el principio «solo lo publicado existe»). |
| D-6 ✅ | **RESUELTA (humano, 2026-10-11, cambio de alcance aprobado):** el catálogo de categorías vive en `/panel/categorias` con **un permiso nuevo `categorias`** (décimo del catálogo de F2); «eventos» y «actividades» conservan sus permisos independientes y no autorizan el catálogo. | Gestión del catálogo dentro de cada módulo (duplicaría la pantalla) o reutilizar `eventos`/`actividades` como permiso (descartado: el humano aprobó el permiso nuevo). |
| D-7 ⟲ | «Eventos y actividades» entra en el grupo **Explorar** del pie (no abre columna nueva). | Columna nueva del pie con un solo item (como F10 §2.4 previó). |
| D-8 ⟲ | La marca **«Pasado»** en el panel es una etiqueta de texto junto al estado (no un cuarto estado de `StatusPill`): publicado + fechas pasadas = publicado que ya no se ve. | Cuarta píldora «Pasado» (fusionaría estado de publicación con visibilidad). |
| D-9 ⟲ | La imagen del elemento se recorta (`object-cover`) al centro en marco 16:9; nunca se deforma. | Mostrar la imagen original con barras (páginas irregulares). |
| D-10 ⟲ | Subtítulo de la página como **cadena de interfaz fija** (no campo del panel). | Que el equipo lo escriba en el panel (más contenido que administrar). |
| D-11 ⟲ | Las **fechas futuras restantes** del evento se muestran todas, en lista compacta (sin toggle). | Sólo la próxima y una nota «y más fechas» (menos transparente). |
| D-12 ⟲ | Las **actividades sin hora** se muestran sin bloque de horario (no un «Hora: —»). | Mostrar «Hora: —» (ruido). |
| D-13 ✅ | Abrir `/eventos` sin nada publicado muestra **página mínima** (H1 + aviso de interfaz) y nav sin el ítem; **sin redirigir** a la portada (confirmado por el humano el 2026-10-11). | Redirigir a `/` (descartado: URL extraña sin explicación). |
| D-14 ⟲ | El **empate de fecha y hora** se ordena por título (es) alfabéticamente: **estable y reproducible** (default de la spec, revisable). | Orden por creación. |
| D-15 ⟲ | Las **categorías** se administran con nombre es/en (en el panel muestra ES y en EN se usa en el público si existe). | Categorías en una sola lengua (la cara del público). |
| D-16 ✅ | **RESUELTA (decisión firme, contrato 2026-10-11):** borrar una categoría **en uso** se rechaza con **409 conflict** (FK `RESTRICT`); la UI lo dice («Está en uso por N elementos: quítala de los elementos primero») y **no ofrece desligarla ni continuar**. | Avisar la cuenta y permitir continuar desligando (superado: la API lo rechaza). |
| D-17 ⟲ | Sin paginación en la página pública (spec default); si algún día crecen demasiado, el «Ver más» sería un ajuste menor acordado con el humano. | Paginar a N por página. |
| D-18 ⟲ | `Esc` del `Dialog` de evento/actividad NO descarta sin aviso cuando hay cambios (mejor seguro con lista de fechas). | Igual que F3 (cierre directo). |
| D-19 ⟲ | **No hay filtro de rango de fechas** en ninguno de las vistas: el alcance temporal lo decide el sistema (desde hoy + próximo cronológico). | `DateRangeFilter` del panel en listados (aporte técnico, no UX del público). |
| D-20 ⟲ | El enlace «Ver ubicación» lleva `aria-label` completo («Ver ubicación de {título}» / «See location for {title}»). | Texto a solas con contexto del h3 cercano. |
| D-21 ⟲ | **Contenido en fallback dentro de página en inglés** con `lang="es"` en su envoltorio. | Sin marca de idioma en el nodo (menos correcto para lectores de pantalla). |
| D-22 ✅ | **RESUELTA (alineada con decisión firme, R4-3/contrato, 2026-10-11):** las fechas se presentan **sin conversión de zona horaria**: el helper construye la fecha desde sus partes (`YYYY-MM-DD` + `HH:MM`); el dato es el local del lugar; nada de `new Date("YYYY-MM-DD")` (los parsea como UTC y desplaza un día en zonas UTC−x). | Formatear con la zona del navegador vía `lib/format.ts` (superado: era una trampa de desplazamiento de día). Fijar zona de la iglesia (cambio de contrato; fuera de F4). |
| D-23 ⟲ | La **presencia de la sección en la navegación** se deduce de la respuesta pública de eventos compartida por queryClient (una sola petición alimenta página, barra y pie). | Consulta separada solo para saber si la sección existe (doble carga). |

---

## 10. Notas para el `dev-frontend` / `qa-tester`

- **La lógica con pruebas**: filtro «desde hoy en adelante» + orden cronológico por próxima fecha (empate estable, D-14), cálculo de «pasado» (evento con todas sus fechas pasadas no se muestra, y el panel lo marca), bloques según disponibilidad (D-5), formato de fecha por idioma sin alterar el dato (con `<time>` intacto), fallback es→en del contenido (resuelto por el servidor), validaciones del `DatesEditor` (≥1 fila, formato y fin > inicio por fila), español obligatorio, límites (2.000 detalle), enlace de mapa con formato válido, duplicado dentro del módulo, ciclo publicar/retirar/eliminar con auditoría, permisos cruzados eventos/actividades/categorías, y presencia de la nav/mapa **solo con contenido publicado**.
- **SC-002**: `qa-tester` verifica por cualquier vía que un borrador no se filtre (HTML, JSON público, URL de imagen → 404 si el contenido que la usa no está publicado, patrón de F3 §10), y que no se exponen datos de elementos pasados.
- **El diseño no cambia el sistema de F10**: usa `Card`, `SectionHeading`, `Reveal`, `useSiteLinks`, i18n, temas y tokens tal cual; la página nueva es «otra sección» y no una marca visual propia (0 nuevos tokens).
- **La única lista de enlaces** se extiende con el ítem de ruta; NO se duplica dentro de `PublicNav` ni `MobileMenu` (prueba congelada de F10 sobre la `MobileMenu` exige la MISMA lista). Los enlaces de anclas de portada construidos fuera de `/` van con la base `/#ancla`.
- **El panel no se rediseña** (F10B): usa los componentes de F2/F3 tal cual; los accesos de InicioPage van en `AccesosRapidos` con `hasPermission`.
- **La imagen del panel usa `ImageUploader` de F3**, con `alt` obligatorio en español y en opcional; previsualización local (`FileReader`); los archivos se sirven por `GET /api/v1/media/{fileName}` **solo si el contenido que los usa está publicado** (patrón F3). Al fallar, la tarjeta pública no se rompe.
- **Textos SIEMPRE del diccionario en la página pública** (`features/publico/i18n`), nunca JSX suelto: 100 % de la interfaz traducible (SC-006).
- **SC-008** se verifica con checklist manual de QA en ambos temas (claro y oscuro), teclado, lector de pantalla, 320 px y zoom 200 %: mismas bases de F3/F10.
- **Verificar contra el Manual de marca** en la revisión visual: fuentes correctas, paleta con proporción, sin recolorear el logo, `alt` definido en el panel. El acento turquesa de los rótulos de bloque y el borde del foco siguen el sistema.

---

## 11. Trazabilidad — qué pantalla cubre qué US/FR

| US / FR | Pantalla / componente | § |
|---|---|---|
| US1 · FR-001 · FR-002 · FR-003 · FR-004 · FR-005 · FR-006 | `/eventos` (EventsPage, EventsSection, ActivitiesSection, EventCard, ActivityCard, DateBadge) — dos bloques, solo lo próximo, cronológico, próxima fecha + frecuencia, retiro automático de lo pasado | §4.1, §2.1 |
| US2 · FR-010 · FR-012 («eventos») · FR-015 · FR-016 · FR-017 · FR-018 | `/panel/eventos` (EventsPage, EventsList, EventForm con `DatesEditor`), InicioPage accesos | §4.3, §4.4, §5, §2.3 |
| US3 · FR-011 · FR-012 («actividades») · FR-019 | `/panel/actividades` (ActivitiesPage, ActivitiesList, ActivityForm fecha única sin frecuencia) | §4.5, §2.4 |
| US4 · FR-013 · FR-014 · FR-022 | PublishControls por elemento, ConfirmDialog de retirar/eliminar, sección oculta (nav+mapa), sección sin publicados en panel | §4.1, §4.7, §5.c, §2.5 |
| US5 · FR-007 · FR-008 · FR-009 | LanguageSwitcher (F10), contenido con/es EN, `formatoFechas.ts` (fecha escrita por idioma) | §2.2, §4.1, §7.2, §8.2 |
| US6 · FR-020 · FR-021 · FR-023 · FR-024 | Página con sistema F10, `Card` (imagen opcional + alt), teclado/lector, 320 px, 200 %, `prefers-reduced-motion` | §4.1, §8.1 |
| Q4 (pasado) | Marca «Pasado» en panel + retiro automático del público (sin cambios de estado) | §4.3, §9 D-8 |
| Q6 (categorías) | `CategoryChip` (público), `/panel/categorias` (catálogo es/en, inicial sembrado) | §4.1, §4.6, §9 D-6/D-15/D-16 |
| Auditoría (FR-017) | Módulo Auditoría de F2 intacto; no añade UI nueva | §4.9 |
