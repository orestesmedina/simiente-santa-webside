# UX: F6 — Ministerios

> Documento de diseño para `disenador-ux`. Fuentes: `spec.md` de F6 (v2 aprobada el 2026-10-13: 6 US, 29 FR, 17 SC, aclaraciones Q1–Q9 cerradas), `ux.md` de F4 (modelo de formato, tarjetas, bloque de portada y patrones de panel), `ux.md` de F3 (fallback, estados, textos) y `ux.md` de F10 («Marea Turquesa», sistema ya entregado), más el código actual de `frontend/src/` (React 19 + Tailwind 4; componentes de F2/F3/F4/F10 ya construidos y verificados en este repositorio).
> No es código: lo implementa el `dev-frontend`. Rutas y contratos los fija el `arquitecto`; aquí se proponen y los supuestos van marcados.
> **Decisiones de UX a revisar por el humano**: ver §9 (resumen). Cada una lleva su marca ⟲ (pendiente). Reutilizo los patrones ya aprobados de F4 (D-1 ruta propia, D-25 extracto con chip, D-28 sin carrusel) y aquí solo marco lo nuevo de F6.
> **Ajuste 2026-10-13 (decisiones del humano, ya resueltas)**: (a) el interruptor «mostrar la ficha del encargado» inicia **APAGADO por defecto** — «Oculto por defecto» (D-13); (b) `/ministerios` sin ningún ministerio publicado **redirige a la portada**, no muestra el aviso de F4 — «Redirigir a la portada» (D-14). Ambas están aplicadas en todo el documento. D-3 también quedó resuelta por el humano ese día.
> F6 **aplica** el sistema de F10 y los patrones de F2/F3/F4; **no redefine** tokens, navegación, selector de idioma, modo oscuro ni el rediseño del panel (F10B).

---

## 0. Principios de diseño (heredados, aplicados a F6)

1. **Cada elemento es un dato, no decoración.** Nombre, descripción, horario y lugar son los protagonistas; encargado, imagen y enlace de WhatsApp son accesorios opcionales. Nada se manufactura para «llenar»: un ministerio sin imagen, sin encargado, sin horario y sin enlace **se ve completo** (FR-002, SC-017).
2. **Un ministerio no es un evento ni un taller — y se nota.** El horario es **texto libre** del equipo («Todos los viernes, 7:00 p. m.», «Una vez al mes, fecha variable»), **no una fecha** (Q6): F6 no usa `DateBadge`, ni `<time datetime>`, ni calendarios, ni retiro automático (FR-001, FR-024). La sección tampoco mezcla conceptos: solo ministerios, con una sola lista y un solo rótulo (nada de bloques por «tipo»).
3. **La ficha se sostiene sola.** Sin enlace → **sin botón de unirse** (ni vacío ni roto, FR-003); sin encargado publicado → **sin línea de encargado ni de contacto** (ni vacía, ni con solo el nombre, Q9); sin imagen → sin hueco ni reserva; sin horario o lugar → sin esa línea (FR-002, FR-024). Cero campos en blanco en el 100 % de los casos.
4. **Unirse es un paso.** Con enlace publicado, el visitante abre WhatsApp o la invitación del grupo con **un solo toque**, desde el mismo lugar de siempre: el botón verde del CTA de WhatsApp del sitio (FR-003). Sin enlace, la puerta es el encargado **cuando el equipo decide publicar su contacto** (FR-025).
5. **Privacidad de alguien real, no de un campo.** El encargado es una persona: su ficha (nombre + contacto) solo es pública si el equipo enciende el interruptor — que además **inicia apagado**: guardar un encargado nunca lo publica por sí solo (D-13); y al apagarlo **ni el nombre ni el celular/correo salen del servidor**, no solo se tapan en la pantalla (FR-026, SC-014). El panel SIEMPRE explica, en palabras del equipo, qué se hará público.
6. **Apto para todas las edades.** Cuerpo 18 px, objetivos ≥ 44 px, iconos siempre con texto, `prefers-reduced-motion` intacto, 0 desplazamiento horizontal a 320 px (FR-022, FR-023, idea.md §7: adultos mayores que saben lo básico de internet).
7. **Solo lo publicado existe.** Sin ningún ministerio publicado, la sección **no existe** — ni en la barra, ni en el menú, ni en el mapa del sitio — y si alguien abre `/ministerios` a mano es **redirigido a la portada** de forma limpia, sin aviso y sin parpadeo (D-14; **distinto a propósito** de F4, que mantenía al visitante en una página mínima); **el bloque de la portada desaparece por completo, sin mensaje** (FR-005, FR-028). Cero secciones vacías, cero enlaces huérfanos.
8. **El panel hereda F2/F3/F4 tal cual.** Título + banda de ayuda, `Table`/tarjetas, `StatusPill`, `Dialog`, `ConfirmDialog`, `Notice`, validación con Zod y mensajes en español: cero estilos nuevos (FR-011, FR-018).

---

## 1. Tokens aplicados (la verdad vive en F3/F10; solo se citan las decisiones de F6)

| Tema | Reutilizado de | Nota específica de F6 |
|---|---|---|
| Paleta 70/20/10 | F3 §1.1 / F10 §0.1 | Navy dominante con tarjeta en crema/vidrio, acento turquesa en rótulos y foco. Sin colores nuevos. |
| Tipografías | F3 §1.2 / F10 §0.2 | H1 de página y rótulos en Bebas; cuerpo, botones y fichas en Poppins 18 px. F6 no usa Playfair. |
| Espaciado | F10 §0.3 (`--space-section/block/fluid`, `--pad-panel/text`) | La lista es «un plato más» del sitio: H1 → `--space-block` → grilla → botón/pie. |
| Superficies | `Card` (`solid`/`glass`, F10) | Cada ministerio es una `Card solid`; la imagen (si hay) va a sangre 16:9 con `padded={false}`. El bloque de portada vive dentro de una `Card glass` como `ProximamenteBlock`. |
| Movimiento | F10 §5 (`Reveal`/`RevealGroup`, 600 ms / micro 180 ms) | Revelado lateral **una sola vez** para el listado (stagger suave, tope 240 ms), como las secciones de F3. Con `prefers-reduced-motion`: todo visible al instante. |
| Foco visible | F3 §1.4 / F10 §6 | Anillo 2 px: navy sobre claro, crema sobre navy; nunca oculto tras el revelado. |
| Botón de unirse | tokens de WhatsApp de F10 (verde, CTA) | Mismo verde y gramática del CTA de WhatsApp del sitio: icono + texto, 48 px de alto en móvil. |
| Chip «Ministerio» | patrón del chip F4 (D-2/D-25) | Píldora neutra con borde y texto grises, Poppins 0.875 rem, **sin color semántico**: identifica el concepto en el extracto de portada, no un estado. |
| Orden manual | precedente `sort_order, id` de F3/F4 (FR-004) | Columna numérica editable en el panel; en público, el orden llega listo del servidor. Sin drag & drop. |

**Nuevos tokens: ninguno.** F6 usa exclusivamente los de F3/F10.

**Del inventario de F4 NO se usa** (con razón de negocio, no de estilo): `DateBadge` y `formatoFechas.ts` (no hay fechas), `CategoryChip` (no hay categorías), `DatesEditor` (no hay fechas), `MapLink` (no hay enlaces de mapa), la marca «Pasado» (nada vence: no hay fechas). F6 no compite con la agenda: su horario es texto, su orden es el manual del equipo.

---

## 2. Flujo de usuario

### 2.1 Visitante: conoce los ministerios y se une (US1, camino feliz)

1. Abre el sitio (`/`, idioma según preferencia guardada; primera visita: español). En la portada, después del bloque «Lo que viene» de F4 y antes de «quiénes somos», ve el bloque **«Ministerios»** (§4.9): hasta 5 tarjetas con imagen, título y una línea de horario — y el botón **«Ver más ministerios»** (FR-027).
2. Toca «Ministerios» (barra en escritorio; panel del menú hamburguesa en móvil) o el botón del bloque → llega a **`/ministerios`**: página de contenido real con **H1 «Ministerios»** y una sola grilla de fichas en el orden manual del equipo (Q5; FR-004).
3. Recorre las fichas: cada una muestra su **nombre**, su **descripción**, su **horario** y su **lugar** (líneas de texto común, tal como el equipo los escribió — Q6), y — si el equipo lo publicó — su **encargado** con nombre y, si existe, su medio de contacto (celular o correo, uno solo: Q8).
4. Quiere unirse a un ministerio con enlace: activa el botón verde **«Unirse al grupo de WhatsApp»** → se abre WhatsApp o la página de invitación del grupo en un solo paso (FR-003); el sitio queda donde estaba (pestaña nueva con `rel="noopener"`).
5. Un ministerio **sin enlace** se ve igual de completo: no hay botón alguno, y la ficha aporta nombre, descripción y horario/lugar si existen (Q2, SC-017). Su puerta es el encargado **solo si su contacto está publicado** (Q1); si no, es información pura del área — válido y coherente.
6. Termina y recorre el resto del sitio: pie de F10 intacto — el mapa del sitio muestra **«Ministerios»** en el grupo **Explorar** (junto a «Eventos y actividades»), solo si hay algo publicado (FR-021).

### 2.2 Visitante: cambia a inglés (US4)

1. En la barra (o el pie), toca **English** en el selector de F10.
2. La página cambia de inmediato: la **interfaz** («Ministries», «Join the WhatsApp group», «Loading…») va siempre del diccionario es/en; el **contenido** con versión en inglés (nombre, descripción, encargado, horario, lugar) se muestra en inglés.
3. Un campo **sin inglés** se muestra en español, con idéntico diseño y **sin ningún hueco ni marca** (FR-008; el fallback lo resuelve el servidor).
4. El **enlace de WhatsApp no cambia**: es el mismo destino en los dos idiomas (US4 escenario 5 · FR-008; el enlace no es contenido traducible).
5. La elección persiste y se recuerda entre visitas (F3/F10); `html lang` pasa a `en`.

### 2.3 Equipo: gestiona los ministerios (US2, camino feliz)

1. Inicia sesión → **Inicio** del panel: un acceso nuevo en `AccesosRapidos` — **«Gestión de ministerios»** (con permiso «ministerios») → `/panel/ministerios`. Cuenta sin el permiso: no ve el acceso y la ruta está protegida.
2. En la página (título «Gestión de ministerios» + banda de ayuda) ve el listado: tabla (desktop) / tarjetas (móvil) con **Orden** (el número que definió), **Nombre**, **Horario / Lugar**, **Encargado** (con marca «Ficha oculta» si el interruptor está apagado), **Estado** (`Borrador`/`Publicado`) y acciones.
3. Toca **«Agregar ministerio»** → `Dialog` con `Tabs` de idioma **Español / English (opcional)**: nombre y descripción (obligatorios en español), horario y lugar (textos libres opcionales), encargado (opcional) con **un solo medio de contacto** (celular o correo, opcional: Q8) y su **interruptor de visibilidad** — «Mostrar la ficha del encargado en el sitio público» — que inicia **apagado**: la ficha no se publica hasta que el equipo la encienda expresamente (D-13), enlace de WhatsApp (opcional, `https` válido), imagen con texto alternativo (opcional) y **Orden** (número con 0 por defecto, Q5).
4. Guarda → queda **Borrador**: visible en el panel, invisible al público; aviso «Quedó en borrador…».
5. Toca **«Publicar»** en su fila (sin confirmación) → píldora «Publicado» + aviso. El visitante YA lo ve (sección y, si entra en los primeros 5, bloque de portada).
6. **Edita un ministerio publicado** y guarda → aviso **«Cambios guardados: ya visibles en la sección.»** (patrón F3, FR-013). No hay paso intermedio.
7. Cambia el interruptor de visibilidad y guarda → aviso que dice en claro qué pasó: «Ficha del encargado ocultada: el público ya no ve su nombre ni su contacto. Los datos quedan aquí.» o «Ficha del encargado publicada.» (Q9; FR-025). Volver a activar repone la ficha sin reescribir nada (US2 esc. 10–11).
8. **Reordena** cambiando el campo Orden y guardando → el nuevo orden es exactamente el que ve el visitante (FR-016).
9. **Retira** («Retirar de la sección», con `ConfirmDialog` suave) o **elimina** («Eliminar», confirmación explícita de no-recuperable, FR-017) — ambas quedan en la auditoría de F2.

### 2.4 Equipo: publica o retira por elemento (US3)

- **Borrador** → nadie lo ve fuera del panel (FR-012).
- **Publicado** → visible en `/ministerios` en el orden definido y, si está entre los primeros 5, en el bloque de la portada (FR-027).
- Sin **ningún** publicado → la sección entera desaparece de la navegación (barra, menú, mapa), el bloque de la portada no se dibuja (FR-005, FR-028) **y** `/ministerios` redirige a la portada (D-14).
- El listado del panel muestra SIEMPRE el estado real al público hoy (US3 esc. 7); no hay estados intermedios («Pasado» de F4 no existe aquí: nada vence).

### 2.5 Errores del equipo

| Situación | Camino |
|---|---|
| Campo obligatorio vacío (nombre o descripción en español) | Nada se guarda: error **junto al campo** + resumen arriba del formulario; el foco va al primer campo invalidado (FR-014). |
| Enlace mal formado | Error con formato esperado y ejemplo completo; sin guardar (FR-015). También puede dejarse vacío (Q2). |
| Enlace con esquema no seguro (`javascript:`, `data:`, `vbscript:` o distinto de `https`) | «Ese enlace no es válido o no es seguro.» Se explica que debe empezar por `https://`; sin guardar (FR-015, Q3). |
| Medio de contacto con formato no válido | Mensaje según el tipo (celular o correo); sin guardar. Solo se valida formato, no existencia (default revisable, Q8). |
| Versión en inglés de un campo sin su versión en español | Se bloquea al guardar indicando qué campo falta en español (regla por campo); **un campo vacío no obliga a nada** (Q6, Decisión 8). |
| Límites de extensión | Contador vivo al escribir; al exceder, «Recorta X para poder guardar» (FR-019). |
| Imagen inválida o carga fallida | Aviso del problema con el archivo; el formulario conserva el resto y el ministerio queda sin imagen (FR-020). |
| Error del sistema al guardar | Aviso genérico, sin detalles internos, **sin perder lo escrito** (los valores permanecen). |
| Operación sin permiso «ministerios» | Ruta protegida (`RequirePermission`) + página «sin permiso» de F2; si el servidor lo rechaza en caliente: mensaje sin detalles (FR-011). |
| Duplicado exacto (mismo nombre, descripción, encargado y enlace) | «Ya existe un ministerio igual.» Nada se guarda (default revisable). |
| Dos ministerios con el mismo enlace de WhatsApp | **Aviso no bloqueante**: «Otro ministerio ya usa este enlace; revísalo si no es intencional.» El guardado sale adelante (default revisable). ⟲ |
| Dos personas editando el mismo ministerio | Se conserva la última versión completa guardada (default de la spec, como F3). |
| Borrador por cualquier vía | Jamás se expone: ni en HTML, ni en JSON, ni en la URL de su imagen (404, patrón F3). |
| Encargado con ficha no publicada por cualquier vía pública | **0 apariciones** del nombre o del contacto en respuestas públicas (FR-026): el servidor lo resuelve y el frontend solo presenta lo que llega (ver §8.3). |

### 2.6 Persona sin permiso (el caso desfavorable)

- Sin permiso «ministerios»: no existe el acceso «Gestión de ministerios» (ni en `AccesosRapidos` ni en la navegación del panel); si fuerza la URL `/panel/ministerios`: `RequirePermission` → `/sin-permiso` (F2, texto intacto).
- Tener «ministerios» NO da los otros módulos, y ningún otro permiso da acceso a este: los permisos son independientes (Decisión 5; probado en US2 esc. 8). F6 no crea permisos nuevos: usa el permiso «ministerios» reservado por F2 (FR-011).
- El sitio público no pide sesión: no opera el caso «visitante sin permiso».

### 2.7 Visitante: descubre los ministerios desde la portada (US6, Q7)

1. Abre el sitio y, sin navegar más que el scroll, ve en la portada el bloque **«Ministerios»** — **inmediatamente después del bloque «Lo que viene» de F4 y antes de «quiénes somos»** (FR-027): hasta **5** ministerios publicados en tarjetas con imagen, en el orden manual del módulo, cada uno identificado con su rótulo «Ministerio».
2. Toca el **título de una tarjeta** → llega a `/ministerios` en un solo paso (US6 esc. 5).
3. Quiere ver más → toca el botón **«Ver más ministerios»** del bloque → misma sección. Con más de 5, la portada muestra exactamente los primeros 5 (`sort_order, id`), sin paginación ni carrusel (FR-027).
4. Con 1–4 publicados: se muestran exactos, sin huecos ni marcadores vacíos (US6 esc. 3).
5. Con **ninguno** publicado: el bloque **no se dibuja** — sin título, sin mensaje, sin aviso; la portada queda coherente, como cualquier sección vacía de F3 (FR-028). **Diferencia deliberada con F4**: el bloque «Lo que viene» sí muestra su mensaje de estado vacío; este no (decisión del humano, Q7). Un borrador jamás aparece (US6 esc. 6).
6. En inglés: interfaz del bloque traducida y contenido con fallback al español según FR-008, sin marcas.

---

## 3. Rutas propuestas (React Router)

```
Sitio público (sin sesión):
/                       Portada de F3 + bloque de F4 + bloque de F6 (MinisteriosBlock)
/ministerios            «Ministerios» de F6 (MinisteriosPage)
/health                 Estado del sistema (F1), fuera de la navegación pública

Panel (patrón F2/F3/F4, sin rediseño):
/panel                  Inicio (acceso nuevo por permiso)
  /panel/ministerios    RequirePermission(MINISTERIOS)  Gestión de ministerios
  /panel/eventos /panel/actividades /panel/informacion /panel/usuarios /panel/roles /panel/auditoria … (F2/F3/F4, intactos)
```

- **`/ministerios` es una ruta real, como `/eventos`** (mismo patrón aprobado por el humano en F4, D-1). Se añade como **ruta** al final de la **única lista de enlaces** de F10 — después de «Eventos y actividades» — en `useSiteLinks.ts`: `SiteLinkId` se extiende con el id nuevo (propuesta: `'ministries'`; el nombre exacto lo fija el arquitecto) y el vocabulario con la clave i18n `nav.sections.ministries`. El `<nav>` sigue siendo una sola lista en DOM (la congelada de F10/MobileMenu funciona sin cambios); no se toca `nav.label`. ⟲ D-1
- **Presencia en la navegación**: la entrada «Ministerios» existe **solo** si la consulta del **bloque de la portada** (`usePortadaMinisterios` — la misma que comparte `MinisteriosBlock` por `queryClient`) trae al menos un ministerio publicado. *(Ajuste 2026-10-13, puerta 2 — autoridad: T628, P6-9, R6-8: la propuesta D-23 del documento de F4 — presenciar la entrada con la consulta pública de la sección, «una sola petición paga página + barra + pie» — quedó superada por el criterio técnico confirmado por el humano. El resultado observable no cambia — misma regla «hay al menos un ministerio publicado» — pero son **dos consultas**: el bloque alimenta barra, menú y pie, y `/ministerios` se alimenta por su cuenta.)* Sin publicados: ni barra, ni `MobileMenu`, ni pie (FR-005).
- **Mapa del sitio del pie**: «Ministerios» entra en el grupo **Explorar** (no se abre columna nueva), mismo criterio D-7 de F4.
- **Sección oculta abierta a mano**: `/ministerios` con 0 publicados **redirige a la portada** (D-14, decisión del humano 2026-10-13). Es **distinto a propósito** de `/eventos` de F4 (que muestra `HiddenNotice` sin redirigir): no unificar los dos comportamientos. Mecánica sin parpadeo: ver la fila «Sección oculta» de §4.1.
- **Permission code**: el módulo se autoriza con el permiso existente **`ministerios`** de F2 (`RequirePermission`); el código técnico exacto lo confirma el arquitecto frente al catálogo sembrado de F2 (FR-011). No hay permiso nuevo.
- **Portada**: el bloque se inserta en la secuencia de `HomePage` entre `<Reveal key="proximamente">` y `<Reveal key="quienes-somos">` (FR-027). Dirección de entrada propuesta: `from="left"` (mantener la alternancia hacia «quiénes somos» puede forzar dos «left» seguidos: decisión menor de implementación, sin tocar las demás secciones).

---

## 4. Pantallas

### 4.1 Sitio público — `/ministerios` (visitante; US1, US4, US5)

**Propósito**: en una lectura (SC-001, SC-012) el visitante sabe qué áreas de servicio hay, en qué consisten, dónde y cuándo sirven, quién encarga cada una **cuando el equipo decidió publicarlo** y cómo unirse cuando existe el enlace.

**Estructura** (página de contenido bajo `PublicLayout` de F10):

1. **Enlace de salto** (_skip link_) — de F10, intacto; `<main>` de `/ministerios` es su destino aquí.
2. **H1 «Ministerios»** + subtítulo fijo de una línea (cadena de interfaz, no campo del panel — mismo criterio D-10 de F4): «Las áreas de servicio de nuestra iglesia: conoce cada una y cómo participar.» / («The service areas of our church: get to know each one and how to take part.») ⟲ D-2
3. **Grilla única de fichas de ministerios** (un solo bloque, sin rótulos de concepto internos — solo hay ministerios, FR-001): todos los **publicados**, en el orden manual del servidor (`sort_order, id`), con `RevealGroup` y stagger suave.
4. **Pie de F10 intacto**: mapa del sitio con «Ministerios» en **Explorar** (solo si hay publicados), selector de idioma en la franja, tema claro/oscuro por tokens.

**Composición de una ficha (`MinistryCard`, `Card solid` del sistema)** — de arriba a abajo, **cada elemento solo existe si tiene dato**:

- **Imagen** (si hay): a sangre del marco superior, 16:9, `object-cover` al centro, con el `alt` del equipo (obligatorio en español, en opcional) — **misma pieza** `CardImage` de F4 (`features/publico/eventos/ItemCardParts.tsx`): sin imagen no hay hueco ni reserva, y si falla un `onError` la retira sin quebrar el layout (FR-020). ✅ D-3 (resuelta por el humano 2026-10-13; promoción de la pieza, ver §6.3)
- **Nombre** (h3, Poppins 600 1.25rem — como el h3 de F3/F4; **no** es Bebas grande: no compite con el H1).
- **Descripción**: texto plano multilínea con saltos preservados, cuerpo Poppins 18 px; se muestra **completa** (sin «Ver más», mismo criterio D-3 de F4); hasta 2.000 caracteres acomodados sin romper (FR-019).
- **Horario** (si hay): línea con icono 🕒 + el texto tal como el equipo lo escribió («Todos los viernes, 7:00 p. m.», «Una vez al mes, fecha variable»), en Poppins normal — **no** hay `DateBadge`, ni `<time>`, ni recalcular nada (FR-024). Texto largo: se envuelve con `text-wrap: balance`, conservando tildes, ñ, ¿ ¡ y emojis.
- **Lugar** (si hay): línea con icono 📍 + texto libre («Salón principal», «En la casa de la familia Pérez»). Sin enlace de mapa: F6 no lo tiene (fuera de alcance; no se reusa `MapLink`).
- **Encargado** (solo si el interruptor está activado y existe nombre): línea con icono 👤 + **nombre**; y **solo si existe y hay medio de contacto guardado**, el mismo bloque añade la línea de contacto: 📱 celular o ✉️ correo (el que el equipo guardó, un solo medio). Con ficha publicada y encargado sin contacto: **solo el nombre**, sin línea de contacto, y **no es un error** (Q8). Con interruptor apagado, o sin encargado: **el bloque no existe** — nunca una línea vacía ni con nombre huérfano (FR-026). ⟲ D-4 (contacto como texto, ver §9)
- **Botón de unirse** (solo si hay enlace): `PublicButton` en su variante verde (el verde de WhatsApp del sistema), icono + **texto** «Unirse al grupo de WhatsApp» / («Join the WhatsApp group»), con `target="_blank"` + `rel="noopener"` y nombre accesible completo «Unirse al grupo de WhatsApp de {nombre}» / («Join the WhatsApp group of {name}»). Un toque, destino exacto (FR-003). **Sin enlace: el botón no existe** — ni vacío ni deshabilitado (Q2). ⟲ D-5
- La `Card` **no es un enlace en bloque**: solo lo interactivo explícito es clicable (regla de F4 §4.1) — el lector de pantalla recibe una ficha legible, no un «objeto enlace».

**Propuesta wireframe desktop y móvil** (esquema de referencia, no maqueta):

```
DESKTOP
[ BARRA F10: 🌱 Simiente Santa · Quiénes somos · Horario · WhatsApp · Contacto · Redes
                                      · Eventos y actividades · Ministerios | Español/English | 🌙 ]
══════════════════════════════════════════════════════════════════
  MINISTERIOS  (h1, Bebas)
  Las áreas de servicio de nuestra iglesia: conoce cada una y cómo participar.
        ┐ --space-block
  ┌─ Card ─────────────────┐  ┌─ Card ─────────────────┐
  │ [imagen 16:9 a sangre] │  │ (sin imagen: no hueco) │
  │ Alabanza               │  │ Escuela dominical      │
  │ Llevamos la adoración… │  │ Cada sábado los niños… │
  │ 🕒 Todos los viernes,   │  │ 🕒 Domingos, 9:00 a. m. │
  │    7:00 p. m.          │  │ 📍 Salón 2             │
  │ 📍 Templo central      │  │ 👤 Ana Rosales         │
  │ 👤 Josué Ramírez       │  │    ✉️ escuela@…        │
  │    📱 +504 9999-9999   │  │ (sin enlace: no botón) │
  │ [🟩 Unirse al grupo     │  └────────────────────────┘
  │    de WhatsApp]        │
  └────────────────────────┘
  … grilla 2 col desktop, 1 col móvil; sin paginar (default: todos) …
  PIE DE F10 (mapa del sitio: Explorar incluye «Ministerios»)
══════════════════════════════════════════════════════════════════

Casos «sin dato» que se resuelven solos (misma grilla, sin huecos):
  · el ministerio de matrimonios (sin enlace, sin encargado publicado,
    sin imagen): nombre + descripción + «🕒 Una vez al mes, fecha variable» — nada más, y está completo.
  · un ministerio con ficha oculta (interruptor apagado): sin línea de encargado;
    el visitante no sabe que existe ese dato (FR-026).
```

**Estados de la página pública**:

| Estado | Qué ve el visitante |
|---|---|
| Carga (primera) | Esqueleto con medidas reales: H1 + grilla de `Card` fantasma (imagen 16:9 + tres líneas), patrón `SectionSkeleton` de F10; la barra, el idioma y el tema ya visibles; cero brincos de layout. |
| Éxito | Grilla completa con las fichas en el orden manual; cada ficha muestra solo lo que tiene. |
| **Sección oculta** (0 publicados; Q7) | La entrada **no existe en la barra, ni en el menú, ni en el mapa del pie** (FR-005). Si alguien abre `/ministerios` a mano — enlace guardado o URL tecleada — **redirige a la portada** con `<Navigate to="/" replace>` (D-14): **nunca se pinta una sección vacía**. Mecánica sin parpadeo: la decisión de redirigir la toma la página con **su propia consulta** (`usePublicMinistries`, ver §6.3); la presencia en la navegación se deduce aparte del bloque (`usePortadaMinisterios` — dos consultas en F6, criterio puerta 2 / T628); si el dato de la página ya está en caché (caso típico), la redirección ocurre **en el primer render**, sin esqueleto, sin H1 y sin grilla; si aún no llegó, la página muestra **solo el esqueleto estándar de carga** mientras se decide — jamás `HiddenNotice`, jamás la grilla vacía, jamás un mensaje de error. Al confirmarse 0 publicados, la redirección ocurre **en el mismo ciclo de render** en que llega el dato, con `replace` para no dejar `/ministerios` en el historial. La persona llega a la portada **sin sentir que algo falló**: sin errores, sin texto de aviso, sin saltos de contenido. **Distinto a propósito de `/eventos` de F4**, que mantiene al visitante con `HiddenNotice`: no unificar. |
| Error de carga sin nada | H1 + mensaje bilingüe claro: «No pudimos cargar los ministerios. Revisa tu conexión y vuelve a intentarlo.» + botón «Volver a intentar» (patrón F3/F10). |
| Error con caché previa | Se mantiene el último contenido válido y un aviso discreto: «No pudimos actualizar; ves la última versión disponible.» (F3 §4.1). |
| Imagen rota o sin imagen | La ficha conserva nombre, descripción, horario/lugar, encargado y botón; `onError` retira la imagen sin mover el layout. |
| Fallback es→en (US4) | El campo sin inglés se muestra en español, idéntico diseño, **sin** marca de «sin traducir» (FR-008). El enlace no cambia de destino. |
| Orden estable | Fijo de punta a punta: siempre `sort_order, id` (FR-004); igual resultado con cualquier ancho de pantalla o idioma. |
| Duplicado de concepto | La página solo contiene ministerios; nada de talleres (F5) ni eventos (F4) (FR-001, SC-013). |
| `prefers-reduced-motion` | Sin revelado ni stagger; contenido íntegro al instante (FR-023). |
| Borrador | Jamás visible en esta página ni en ninguna respuesta pública, por ninguna vía (FR-012, SC-002). |

**Responsivo**:

- **Móvil (< 640 px)**: grilla de 1 columna, fichas a todo ancho; imagen 16:9 arriba; botón de unirse a todo el ancho con 48 px de alto. Sin desplazamiento horizontal a 320 px (SC-007).
- **Tableta (≥ 640 px)**: 2 columnas de fichas; H1 y subtítulo crecen con `clamp`.
- **Escritorio (≥ 1024 px)**: `max-w-6xl`, fichas con más aire (`--space-block`); el botón de unirse va al final de la ficha, alineado a la izquierda; el bloque de encargado y contacto se lee en dos líneas cómodas.

### 4.2 Panel — Inicio (cambio menor en F2)

`AccesosRapidos` añade **una** tarjeta con el patrón ícono + texto existente:

- **«Gestión de ministerios»** (con permiso `ministerios` → `/panel/ministerios`).

Una cuenta sin el permiso no ve la tarjeta (mismo criterio de F4). Ni «ministerios» implica otro módulo, ni otro implica a este (Decisión 5). El área del módulo en auditoría es «ministerios» (la registra F2 con el patrón existente; sin UI nueva).

### 4.3 Panel — Gestión de ministerios — `/panel/ministerios` (US2, US3)

**Propósito**: administrar el listado sin depender de nadie, y ver de un vistazo qué es público hoy y cómo está compuesto cada ministerio (incluida la decisión de privacidad del encargado).

- Título «Gestión de ministerios» + banda de ayuda: «Aquí administras las áreas de servicio de la iglesia (p. ej., alabanza, sonido, acción social). Solo lo publicado es visible al público.»
- **Listado** (`Table` desktop / tarjetas apiladas en móvil, patrón F2/F4): columnas
  - **Orden** — el valor numérico definido por el equipo (Q5; ordenación `sort_order, id`).
  - **Nombre** (español).
  - **Horario / Lugar** — los textos tal cual («Todos los viernes, 7:00 p. m. · Templo central»); con nada: «—».
  - **Encargado** — el nombre guardado, o «—»; si el interruptor está apagado se añade la etiqueta de texto **«Ficha oculta»** (contexto: en la fila; significado en un `title`/leyenda, ver §4.5).
  - **Estado** — `StatusPill` (`Borrador`/`Publicado`), como en F3/F4. **No** existe la marca «Pasado»: nada vence (no hay fechas).
  - **Acciones** — `Editar` · `Publicar`/`Retirar de la sección` · `Eliminar`.
- **Agregar ministerio** → `Dialog` con `MinistryForm` (§4.4).
- **Eliminar**: `ConfirmDialog` danger: «"{nombre}" se eliminará para siempre. Esto no se puede deshacer.» (patrón F2/F4). Queda en auditoría (FR-017). No hay papelera.
- El listado completo (sin paginación, default) va ordenado SIEMPRE `sort_order, id`, así el equipo ve el orden que consume el público.

### 4.4 Panel — Formulario de ministerio (`MinistryForm`, en `Dialog`)

Patrón F3/F4 (pestañas de idioma + Zod + avisos junto al campo) con los campos propios de F6:

- **Tabs de idioma**: **Español** (obligatorio) / **English (opcional)** — con la ayuda de F3 («Si lo dejas vacío, el público verá este contenido en su versión en español.»). Regla **por campo**: un campo opcional (horario, lugar, encargado) rellenado en inglés sin su versión española bloquea el guardado indicando qué falta; **un campo vacío no obliga a nada** (Q6, Decisión 8).
- Pestaña **Español**:
  - **Nombre** (`Field`, obligatorio; contador vivo, límite corto del contrato — el arquitecto fija el número, default de F3/F4: 160).
  - **Descripción** (`TextArea`, obligatoria, contador vivo `{n}/2.000`, saltos preservados; admite tildes, ñ, ¿ ¡ y emojis íntegros).
  - **Horario** (`Field`, opcional; ayuda: «P. ej. «Todos los viernes, 7:00 p. m.» o «Una vez al mes, fecha variable». Se muestra tal como lo escribas — no se calculan fechas.»).
  - **Lugar** (`Field`, opcional; ayuda: «P. ej. «Salón principal» o «En la casa de la familia Pérez».»).
  - **Encargado** (bloque opcional, ver §4.5): nombre + **un** medio de contacto (opcional) + interruptor de visibilidad.
- Pestaña **English (opcional)**: nombre, descripción, horario, **lugar** y **encargado (nombre)** en inglés, y el alt de la imagen. El contacto NO se traduce (es un dato, no un texto).
- Campos no lingüísticos (después de las pestañas):
  - **Enlace del grupo de WhatsApp** (`Field`, opcional; ayuda: «Pega la invitación del grupo; por ejemplo: https://chat.whatsapp.com/…». Cualquier https válido, cualquier dominio (Q3)). Validación Zod: no vacío → debe ser URL absoluta con esquema **`https`**; se rechaza `javascript:`, `data:`, `vbscript:` y cualquier otro esquema (FR-015). No se verifica vigencia (Out of Scope).
  - **Orden** (`Field` numérico, opcional con valor 0 por defecto; ayuda: «Los ministerios se muestran en orden: menor número primero. Con empatía, el más antiguo va antes.»). Es el mecanismo de reordenamiento (Q5): **editor numérico, sin drag & drop ni flechas** — mismo precedente que la columna de `sort_order` de F3/F4. ⟲ D-6
  - **Imagen** (`ImageUploader` de F3): previsualización inmediata y `alt` obligatorio en español (en opcional); reglas de identidad y ayuda de F3. Si falla la carga: aviso y el resto del formulario queda íntegro (FR-020). Sin catálogo de íconos.
- **Guardar**: `Button` loading («Guardando…»); el modal no se cierra hasta terminar; `Esc` no descarta lo escrito sin aviso (mismo criterio D-18 de F4, con formulario largo). Al guardar con el mismo enlace de WhatsApp que otro ministerio: aviso no bloqueante (§2.5). ⟲ D-7

**Nota de reutilización**: `ItemFormFields` de F4 no sirve tal cual (lugar obligatorio, enlace de mapa, organizador distinto, sin bloque de encargado): `MinistryForm` compone directamente las piezas compartidas (`Tabs`, `Field`, `TextArea`, `ImageUploader`, `Dialog`) recogidas del mismo patrón; si el arquitecto prefiere parametrizar `ItemFormFields`, que decida, pero el diseño no lo exige.

### 4.5 Panel — Bloque «Encargado» en el formulario (privacidad in-UI, Q1 + Q8 + Q9, FR-025/FR-026)

Todo junto y con la explicación textual en el momento:

1. **Nombre del encargado** (`Field`, opcional, es/en con pestaña): la persona o equipo responsable del área. Vacío → el ministerio no tiene encargado (caso válido y completo).
2. **Medio de contacto — uno solo, opcional aunque haya encargado** (`Select` + `Field` dinámicos): opciones **«— Sin contacto —»** (por defecto) / **«Celular»** / **«Correo».** El Select cambia la etiqueta y la validación del único `Field` de debajo:
   - «Número de celular (opcional)» — dígitos y `+` opcional (formato por defecto, revisable; el detalle lo fija el arquitecto).
   - «Correo electrónico (opcional)» — formato de correo.
   - Al cambiar el tipo, el valor anterior se reemplaza por el nuevo campo vacío (ayuda si hay dato: «Al cambiar el tipo se borra el contacto anterior»), nunca se guardan los dos (Q8: «nunca ambos»). Ayuda general: «Si el ministerio ya tiene su grupo de WhatsApp, quizá te sea útil el enlace en lugar de tu celular.»
3. **Interruptor de visibilidad** (nuevo `ShowContactSwitch`, §6.4; `role="switch"`): etiqueta clara **«Mostrar la ficha del encargado en el sitio público»**, **APAGADO por defecto** al crear (D-13, decisión del humano 2026-10-13: «Oculto por defecto»): la ficha del encargado **no se publica hasta que el equipo la encienda expresamente** — guardar un encargado no lo publica por sí solo. Enciende y apaga **la ficha completa**: nombre y contacto juntos.
   - Ayuda fija del interruptor, en palabras del equipo: «La ficha del encargado está oculta hasta que actives este interruptor. Empieza apagado para que seas tú quien decida cuándo hacer pública su información.»
   - Con el interruptor apagado (el estado por defecto al crear) y encargado con datos: **`Notice` warning claramente visible** (no margen fino): «La ficha del encargado quedará oculta: el público **no** verá ni su nombre ni su contacto, en ningún lugar del sitio. Tus datos quedan guardados aquí; puedes activar el interruptor otra vez cuando quieras publicarlos.»
   - Con encargado y ficha visible (interruptor encendido expresamente): leyenda de tono verde «Se publicará: el nombre y (si existe) el contacto del encargado aparecerán en la página de Ministerios.»
   - Con el interruptor apagado y encargado sin datos: la leyenda avisa igualmente (el dato no existe, la ficha siempre estará vacía si se activa).
   - Si el encargado no tiene nombre y el interruptor está ENCENDIDO, ayuda breve: «Para publicar la ficha, escribe primero el nombre.» (la ficha pública en ese caso queda vacía de facto — válido, no es error).
4. El contacto se valida **solo por formato** cuando existe; nunca cambia la decisión de visibilidad: el interruptor es la única llave (US2 esc. 10–11; FR-026 aplica en el servidor, no aquí).

**En el listado** (§4.3), el encargado con ficha oculta lleva la etiqueta «Ficha oculta» y al ver el listado en móvil la marca va junto al nombre. El panel **siempre** muestra los datos guardados (dato guardado ≠ dato publicado): editar vuelve a cargar el interruptor con la última decisión; el default **apagado solo aplica a la creación** (un ministerio nuevo nunca publica su ficha sin gesto expreso: D-13). ⟲ D-8 (textos y ubicación de la leyenda)

### 4.6 Pantalla de estado vacío (módulo del panel)

- «Todavía no hay ministerios. Agrega el primero para que el público lo vea.» + `EmptyState` con botón «Agregar ministerio».
- Si hay ministerios pero **ninguno publicado**: aviso destacado «Nada está publicado ahora: el público no ve esta sección ni el bloque de la portada, y `/ministerios` redirige a la portada, hasta que publiques al menos un ministerio.» (FR-005 + FR-028 + D-14).

### 4.7 Texto largo, sin datos, muchos ministerios

| Caso | Cómo se ve |
|---|---|
| Descripción de 2.000 caracteres | La ficha pública lo muestra completo (sin «Ver más» — D-3 de F4); móvil y escritorio sin desbordes (Poppins 18, ancho fluido). |
| Horario largo o variable («Una vez al mes, fecha variable») | Se muestra en una línea con `text-wrap: balance`, sin romperse; **sin** fechas ni calendario inventado (Q6). |
| Sin imagen | Ficha con solo textos; no hueco, no reserva, no sombra distinta (FR-020). |
| Sin encargado / con ficha oculta | Sin línea de encargado en público; en panel sí se ve el estado («—» o nombre + «Ficha oculta»). |
| Con ficha visible y sin contacto | La ficha pública muestra **solo el nombre** — sin línea vacía de contacto (Q8; no es error). |
| Sin enlace | Sin botón de unirse (Q2; SC-009: 0 botones vacíos o rotos). |
| Sin horario o sin lugar | Sin esa línea; cero «—», cero «sin datos» en público (Q6). |
| Sin enlace Y sin encargado publicado | Caso válido SC-017: ficha con nombre + descripción + horario/lugar si existen; **completa y coherente**. |
| Muchos ministerios (10+) | La sección los lista **todos** en el orden definido (1 col móvil / 2 tablet+; sin paginación, default de la spec). El extracto de portada queda fijo en 5. |
| Mismo encargado en varios ministerios | Cada ministerio muestra su propia línea de encargado; sin ficha compartida (Key Entities). |
| Duplicado exacto | Rechazado en el formulario (aviso; nada guarda). Mismo enlace de WhatsApp en dos ministerios: aviso no bloqueante (§2.5). |

### 4.8 Cross-check con F2 (área de auditoría)

- La auditoría de F2 registra cada operación con textos «Creó/Modificó/Publicó/Retiró/Eliminó el ministerio "{nombre}"» + resultado, **y las operaciones del interruptor** («Publicó/ocultó la ficha del encargado del ministerio "…"») + los intentos que no se completaron (FR-017). El módulo de auditoría no cambia su UI.
- El registro es solo lectura, como en F2/F3/F4 (SC-011).

### 4.9 Portada — bloque «Ministerios» (`/`; US6, Q7, FR-027–FR-029)

> Único contenido de F6 en la portada de F3 (FR-021). Sigue **el patrón ya construido de `ProximamenteBlock` de F4** (misma maqueta, mismo esqueleto, misma fila) con una diferencia de negocio firmada por el humano: **sin publicados, el bloque no se dibuja** — sin mensaje de estado vacío (FR-028). Datos por **endpoint público propio** — propuesta `GET /api/v1/portada/ministerios?lang=es|en` vía `usePortadaMinisterios`: tope **5** resuelto en servidor, orden `sort_order, id`, solo lo publicado, DTO mínimo, `Cache-Control: no-store` — una sola petición para el bloque; el contrato exacto lo fija el arquitecto. **Supuesto pendiente de confirmar**: nombre del endpoint y del hook (FR-027 lo sugiere con ese ejemplo).

**Propósito**: quien entra al sitio conoce las áreas de servicio de la iglesia antes de navegar, y de ahí salta a la sección completa en un paso.

**Ubicación (Q7, ya decidida por el humano)**: **inmediatamente después del bloque «Lo que viene» de F4, antes de «quiénes somos»**. Razón: el humano pidió «similar a los eventos y actividades, solo que después de la sección de eventos»; dos bloques del mismo patrón, contiguos, en la zona «descubrimiento» de la portada; el resto de F3 queda intacto.

| Estado | Qué ve el visitante (portada) |
|---|---|
| Con 5 publicados | Encabezado `SectionHeading` «Ministerios» + **una fila horizontal de 5 tarjetas compactas con imagen** (misma rejilla D-28 de F4: 1 col móvil, 2 tableta, 5 escritorio), en el orden manual del módulo. |
| Con menos de 5 (1–4) | Exactamente esas tarjetas en la fila; sin huecos ni marcadores vacíos, sin estirar (US6 esc. 3). |
| **Con 0 publicados** | **El bloque no se dibuja**: sin título, sin mensaje, sin aviso, sin reservar espacio — regla general de F3 de ocultar secciones vacías (FR-028). La portada queda coherente; **diferencia deliberada** con «Lo que viene» (que sí muestra su estado vacío). |
| Carga (primera) | Esqueleto **propio y de altura acotada**: encabezado + fila de tarjetas fantasma con medidas reales (mismo patrón del esqueleto de `ProximamenteBlock`); si el resultado es 0, el bloque desaparece al llegar el dato (brinco de layout aceptado: el caso 0 es raro en un módulo que se administra con contenido). ⟲ D-9 |
| Error sin nada en caché | El bloque **no se dibuja** (no deja un título huérfano sin contenido), coherente con la regla general de F3 y con FR-028. Sin «Volver a intentar» en el bloque: el usuario lo abre desde `/ministerios` o recargando. ⟲ D-10 |
| Error con ítems en caché | Se muestra el último contenido válido + aviso discreto: «No pudimos actualizar; ves la última versión disponible.» (patrón F3, `error.partial`). |
| Fallback es→en (US6 esc. 7) | Interfaz del bloque por diccionario; contenido (nombre, horario, lugar) según FR-008 con fallback del servidor, sin marcas; `alt` en español si no hay inglés. |
| Borrador | Jamás aparece como tarjeta (US6 esc. 6, FR-012). |
| `prefers-reduced-motion` | Sin `Reveal` ni transiciones: ítems visibles de inmediato. |

**Composición de una tarjeta compacta (`MinistryItem`)** — misma gramática que `UpcomingItem` de F4, en el idioma del visitante:

- **Imagen**: 16:9 a sangre del marco de la `Card solid` (recorte al centro), con el `alt` del equipo; sin imagen o si falla: **se retira sin hueco ni reserva** (la tarjeta conserva su tamaño y la fila no se rompe).
- **Rótulo de concepto**: chip pequeño fijo **«Ministerio»** / («Ministry», texto con borde, sin color semántico — sirve porque en la portada no hay rótulos por concepto y el visitante distingue el bloque de ministerios del de eventos con idéntico vocabulario que usó F4). ⟲ D-11
- **Título** (Poppins 600 1.0625rem, máx. **2 líneas** con elipsis): **es el enlace** de la tarjeta, hacia `/ministerios`, con nombre accesible completo «Ministerio: {nombre}» / («Ministry: {name}»). Solo el título es clicable (nunca la `Card` entera).
- **Horario** (si existe): línea con icono 🕒 + texto truncado con elipsis visual y el texto completo para lectores de pantalla (`title` + texto nativo).
- **Lugar** (si existe): línea con icono 📍, igual que el horario.
- **Sin horario/lugar/enlace**: la tarjeta no dibuja esas líneas; notablemente **no** hay botón de unirse en el extracto — el botón del bloque es el único CTA (el paralelo con F4: el extracto invita a la sección, no ejecuta la acción). ⟲ D-12

**Botón del bloque** — debajo de la fila, alineado al final (a todo el ancho en móvil): `PublicButton` variante secundaria «Ver más ministerios» / («See more ministries»): lleva a `/ministerios` (44 px de alto). Único paso 2 del bloque: sin paginación, sin pestañas, sin ítems 6+ (FR-027). Solo existe con 1–5 ítems — con 0, el bloque entero no existe.

**Responsivo** — mismo D-28 de F4, sin reapertura: móvil columna única a todo ancho; tableta grilla envuelta de 2 columnas; escritorio `grid-cols-5` fijo sin `overflow`; **nunca carrusel ni scroll lateral** (0 scroll a 320 px, SC-007).

**Accesibilidad** — nivel exacto de F4 §4.10: `section` con `aria-labelledby` al `h2` «Ministerios» (mismo nivel que los títulos de la portada —no al nivel del H1 del hero—); ítems como `li` con `h3` (rol list conservado, `<ul>`); orden de lectura = orden del DOM = orden manual; 1 Tab por tarjeta (título-enlace) + el botón del bloque; anillo de foco 2 px; hit targets ≥ 44 px; el chip/rótulo lleva `aria-label` «Ministerio» como calificador del enlace («Ministerio: {nombre}»); horario/lugar son texto, **sin** `<time>`; imágenes con `alt` del panel; `onError` estable; el módulo no compite con las demás secciones de la portada (alturas acotadas del extracto).

---

## 5. Estados por vista con datos (panel)

### a. Vista general del módulo (`/panel/ministerios`)

| Estado | Qué ve |
|---|---|
| Cargando | `EmptyState` kind=loading («Cargando ministerios…»); barra de acciones ya visible. |
| Vacío (sin ninguno) | `EmptyState` con texto y botón de agregar (§4.6). |
| Vacío con todo borrador | Aviso destacado «Nada está publicado ahora…» (§4.6) + listado completo. |
| Error de carga sin nada | «No se pudo cargar los ministerios.» + `Button` «Reintentar» (`Notice` error, `role="alert"`). |
| Error con contenido en caché | Último listado válido + `Notice` discreto sobre la actualización (patrón F3). |
| Éxito | Listado con Orden, píldoras, marca «Ficha oculta» y acciones. |
| Sin permiso | No llega (ruta protegida); si el servidor lo rechaza en caliente: `ForbiddenPage`/mensaje de F2. |
| Éxito de operación | `Notice` temporal verde: «Ministerio creado.» / «Cambios guardados: ya visibles en la sección.» / «Publicado en "Ministerios".» / «Retirado de la sección…» / «Eliminado.» / «Ficha del encargado ocultada…» / «Ficha del encargado publicada.» (§7.1). |

### b. El formulario (`MinistryForm` en `Dialog`)

- Precargado al editar (incluida la última decisión del interruptor); vacío al crear con interruptor **apagado** por defecto (D-13, decisión del humano 2026-10-13: la ficha no se publica hasta que el equipo la encienda expresamente).
- Guardando: `Button` loading; modal sin cerrarse; `Esc` con aviso si hay cambios (D-18 de F4).
- Éxito: cierre + `Notice` en el listado (con el texto específico de ficha del encargado cuando cambió).
- Errores de campo con los textos del §7, junto al campo y resumen arriba (patrón F2), foco al primer error.
- Trampa de foco y devolución del foco a la fila que abrió el `Dialog`.

### c. Publicar / retirar / eliminar / visibilidad del encargado

- **Publicar y publicar la ficha**: acciones directas al guardar (sin confirmación adicional); el foco queda en la fila.
- **Retirar de la sección**: `ConfirmDialog` suave: «"{nombre}" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.» Botones «Retirar» / «Dejar como está».
- **Eliminar**: `ConfirmDialog` danger: «"{nombre}" se eliminará para siempre. Esto no se puede deshacer.» Botones «Eliminar» / «Cancelar».
- **Interruptor al guardar**: el `Notice` de §5.a dice qué cambió para el público; si el interruptor se apaga con encargado con datos, la leyenda warning del formulario (§4.5) ya lo anticipó antes de guardar.
- **Auto-publicar al guardar** (ministerio ya publicado): «Cambios guardados: ya visibles en la sección.»

---

## 6. Componentes React

> Nombres en inglés (coherente con el repo). El listado final de tareas lo fija el arquitecto; aquí es la propuesta con responsabilidad. **Reutilizar antes de crear**: casi todo existe (verificado en `frontend/src/`).

### 6.1 Reutilizados de F2/F3/F4 (sin cambios, o con una nota mínima)

| Componente | Uso en F6 |
|---|---|
| `Button` / `PublicButton` | Acciones del panel; botón de unirse (verde) y «Ver más ministerios» (secundaria) en público. |
| `Field`, `TextArea` (de `features/informacion/components/`), `Select`, `Notice` | Formulario; avisos del panel y avisos públicos. |
| `Dialog` / `ConfirmDialog` | Alta/edición; retirar y eliminar. |
| `EmptyState` | Cargando, vacío y avisos del módulo. |
| `StatusPill` | Estado borrador/publicado (F6 usa la píldora directamente; **no** monta `EstadoBadge` de F4 que lleva la marca «Pasado»). |
| `Table` | Listado del módulo (desktop); tarjetas en móvil como F4. |
| `Tabs` | Pestañas de idioma es/en del formulario (patrón F3/F4). |
| `ImageUploader` | Imagen + `alt` es/en del ministerio. |
| `PublishControls` | Publicar / retirar por ministerio (textos propios del módulo con sus props). |
| `Pagination` / `DateRangeFilter` / `DatesEditor` / `MapLink` / `DateBadge` / `CategoryChip` / `EstadoBadge` | **Ninguno se usa** (sin paginación, sin fechas, sin mapa, sin categorías, sin «Pasado»). |

### 6.2 Reutilizados de F10 (sin cambios)

| Componente | Uso |
|---|---|
| `PublicLayout`, `SiteHeader`, `PublicNav`, `MobileMenu`, `useSiteLinks` | Navegación; **única lista de enlaces** extendida con «Ministerios» (ruta, D-1 de F4). |
| `LanguageSwitcher`, `LanguageProvider`/`useLanguage`, i18n (`es.ts`/`en.ts`/`messages.ts`) | Selector y cadenas; nada nuevo de idioma. |
| `ThemeToggle`/`useTheme` | Modo claro/oscuro en las superficies nuevas. |
| `SectionHeading` | Título del bloque de portada. |
| `Card` | Superficie pública (`variant="solid"` por ficha; `glass` para el contenedor del bloque). |
| `Reveal` / `RevealGroup` | Entrada del listado y del bloque de portada (alternada, stagger suave). |
| `SectionSkeleton` | Cargando la sección pública. |
| `FooterSiteMap`, `BrandLogo`, `GradientBlobs`, `SocialIcon` | Pie, chrome y fondo intactos. |
| `HiddenNotice` (de `features/publico/eventos/`) | **No se usa en F6.** El humano decidió (D-14, 2026-10-13) que `/ministerios` sin publicados **redirige a la portada** en lugar de mostrar el aviso de F4. El componente queda solo para `/eventos`, tal cual. |

### 6.3 Nuevos del sitio público (`frontend/src/features/publico/ministerios/`)

| Componente / hook | Responsabilidad |
|---|---|
| `MinisteriosPage` | Composición de H1 + subtítulo + grilla a partir de la consulta pública; decide carga/error/**redirección a la portada con 0 publicados** (D-14: `<Navigate to="/" replace>`, sin parpadeo — ver §4.1; **no usa `HiddenNotice`**, diferencia deliberada con `EventsPage` de F4). |
| `MinistryCard` | Ficha pública completa: `CardImage` (imagen opcional), nombre, descripción, líneas de horario/lugar/encargado solo con dato, `WhatsappJoinButton` solo con enlace. |
| `WhatsappJoinButton` | Enlace externo en forma de botón verde con icono + texto («Unirse al grupo de WhatsApp»), `target="_blank" rel="noopener"`, nombre accesible completo por ministerio. |
| `EncargadoLine` | Bloque de ficha del encargado (nombre + contacto) **solo con datos**; presentación de celular/correo con su icono; sin lógica de visibilidad (el servidor ya no manda el dato si está oculto). |
| `MinisteriosBlock` (portada) | Bloque «Ministerios» de la portada: `SectionHeading` + fila/grilla de hasta 5 `MinistryItem` a partir de `usePortadaMinisterios` (endpoint propio, tope 5 en servidor); **se oculta por completo sin ítems** (nada de mensaje vacío — diferencia con `ProximamenteBlock`) — devuelve `null` en éxito sin ítems y en error sin caché; esqueleto mientras carga; aviso discreto con caché y error. |
| `MinistryItem` (portada) | Variante compacta: imagen opcional, chip «Ministerio», título-enlace a `/ministerios`, líneas de horario/lugar si existen (sin encargado ni botón de unirse en el extracto). |
| `badgeChip` / piezas de chip | Píldora «Ministerio» (patrón del chip neutro de F4; puede ser un span dentro de `MinistryItem` — componente propio solo si el arquitecto la quiere compartida). |
| `usePublicMinistries` | TanStack Query sobre el endpoint público del módulo (contrato del arquitecto): alimenta **solo la página** `/ministerios` (H1 + grilla, y con la respuesta propia decide la redirección con 0 publicados, D-14). **No** alimenta la presencia en `useSiteLinks`: esa se deduce de `usePortadaMinisterios` (criterio puerta 2 del 2026-10-13, T628/P6-9/R6-8; la propuesta D-23 de F4 de una sola consulta compartida quedó superada). **Supuesto**: nombre del hook y del endpoint (`GET /api/v1/ministerios?lang=es|en`) al contrato. |
| `usePortadaMinisterios` | TanStack Query para el bloque de la portada: `GET /api/v1/portada/ministerios?lang=es|en` (propuesta, FR-027), DTO mínimo, tope 5 en servidor, `no-store`; alimentada por `MinisteriosBlock` y compartida por `queryClient` con `useSiteLinks` para deducir la presencia de la navegación (una sola petición paga bloque + barra + menú + pie). Son **dos consultas** en F6: nunca consume la consulta de `/ministerios`. |

**Nota** (✅ D-3, resuelta 2026-10-13): `CardImage` vive hoy dentro de `features/publico/eventos/ItemCardParts.tsx`. F6 la reutiliza tal cual (import por feature o promoción a una pieza común de `features/publico/components/` — decisión del arquitecto; su comportamiento exacto ya está probado y no cambia).

### 6.4 Nuevos del panel (`frontend/src/features/ministerios/`)

| Componente / hook | Responsabilidad |
|---|---|
| `pages/MinistriesPage.tsx` | Panel de gestión: título, banda de ayuda, listado + estados (§5.a). |
| `components/MinistriesList.tsx` | Tabla/columnas (Orden, Nombre, Horario/Lugar, Encargado + «Ficha oculta», Estado, Acciones) y tarjetas móviles; patrón `EventsList`. |
| `components/MinistryForm.tsx` (`Dialog`) | Alta/edición completa (§4.4) con el patrón F3/F4 de pestañas + Zod (validación propia en `validation.ts`, espejo de la del service). |
| `components/ContactoEncargado.tsx` | Select de tipo («— Sin contacto —» / Celular / Correo) + único Field dinámico con validación por tipo; nunca dos medios (Q8). |
| `components/ShowContactSwitch.tsx` | Interruptor de visibilidad de la ficha del encargado con `role="switch"`, leyendas contextuales y `Notice` warning al apagar (§4.5; FR-025/FR-026). Único caso de interruptor: si F3/F4 lo necesitaran algún día, se promociona a `src/components/`. |
| `pages/…`/`components/…` | Sin página de catálogo: F6 no crea permisos ni catálogos nuevos (FR-011; diferencia con F4). |
| `messages.ts` | Textos del módulo en español (FR-018) — catálogo del §7.1. |
| Hooks TanStack Query | `useMinistries`, `useCreateMinistry`, `useUpdateMinistry`, `useDeleteMinistry`, y el par de edición cubre publicar/retirar/visibilidad (patrón `useUpdate*` de F3/F4). Ningún `fetch` en componentes (regla F2). |

**Pruebas esperadas** (Vitest + Testing Library + MSW): composición pública con casos 0/1/5/>5 publicados y con cada caso «sin dato» (0 botones vacíos, 0 líneas de encargado/horario vacías — SC-017), orden estable `sort_order, id`, **la vista con 0 publicados: redirección a la portada sin parpadeo** (con la consulta en caché, redirige en el primer render sin pintar H1 ni grilla; sin caché, solo el esqueleto mientras se decide; y `/eventos` conserva su `HiddenNotice` — comportamientos deliberadamente distintos, D-14), default del interruptor **apagado** al crear (D-13) y con él apagado 0 datos del encargado en público (SC-014), fallback de contenido sin inglés (por campo), bloque de portada oculto en 0 ítems y con 0 líneas de aviso, chip/no chip del extracto, validaciones del formulario (obligatorios, https estricto, contacto por tipo con un solo medio, inglés sin español por campo, límites, duplicado exacto, mismo enlace con aviso), ciclo publicar/retirar/eliminar con auditoría, permisos cruzados con «ministerios», y **0 apariciones del encargado oculto en cualquier vía** (SC-014: HTML y JSON público con el interruptor apagado, incluidas respuestas forzadas sin sesión).

---

## 7. Catálogo de textos

### 7.1 Panel (español; FR-018) — añadidos a los de F2/F3/F4

| Situación | Texto |
|---|---|
| Guardado | «Cambios guardados.» |
| Guardado con auto-publicación | «Cambios guardados: ya visibles en la sección.» |
| Crear | «Ministerio creado.» + aviso la primera vez: «Quedó en borrador: no lo verá el público hasta que lo publiques.» |
| Publicar | «Publicado en "Ministerios".» |
| Retirar | «Retirado de la sección. Estará disponible como borrador para publicarlo cuando quieras.» |
| Ficha del encargado apagada al guardar | «Cambios guardados. La ficha del encargado quedó oculta: el público no verá su nombre ni su contacto.» |
| Ficha del encargado encendida al guardar | «Cambios guardados. La ficha del encargado está publicada.» |
| Eliminar (confirmación) | «"{nombre}" se eliminará para siempre. Esto no se puede deshacer.» |
| Eliminado | «Eliminado.» |
| Retirar (confirmación) | «"{nombre}" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.» |
| Nombre vacío | «Escribe el nombre del ministerio.» |
| Descripción vacía | «Escribe la descripción del ministerio.» |
| Enlace mal formado | «Ese enlace no parece una dirección web. Pégalo completo; por ejemplo: https://chat.whatsapp.com/…» |
| Enlace no seguro (`javascript:`, `data:`, otro esquema) | «Ese enlace no es válido o no es seguro. Debe empezar por https:// y ser una dirección normal.» |
| Celular no válido | «Ese número no parece un celular. Revisa los dígitos (y el + opcional del país).» |
| Correo no válido | «Ese correo no parece un correo. Revísalo, por favor.» |
| Inglés sin español (por campo) | «El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de {campo}.» |
| Límite de descripción | «{n} de 2.000 caracteres. Recorta {m} para poder guardar.» |
| Límite de textos cortos | «Ese texto es demasiado largo (máximo {límite}).» |
| Imagen inválida | «Ese archivo no es una imagen válida. Usa JPG, PNG o WebP de menos de {tamaño} y prueba de nuevo.» (F3) |
| Duplicado exacto | «Ya existe un ministerio igual (mismo nombre, misma descripción, mismo encargado y mismo enlace). Edita el que ya tienes o cambia algún dato.» |
| Mismo enlace en dos ministerios (aviso, no bloqueante) | «Otro ministerio ya usa este mismo enlace de WhatsApp. Revísalo si no es intencional; el ministerio sí se guardó.» |
| Leyenda de la ficha visible | «Se publicará: el nombre y (si existe) el contacto del encargado aparecerán en la página de Ministerios.» |
| Leyenda de la ficha oculta (warning) | «La ficha del encargado quedará oculta: el público no verá ni su nombre ni su contacto, en ningún lugar del sitio. Tus datos quedan guardados aquí; puedes activar el interruptor otra vez cuando quieras publicarlos.» |
| Ayuda del interruptor de la ficha (siempre visible junto al control) | «La ficha del encargado está oculta hasta que actives este interruptor. Empieza apagado para que seas tú quien decida cuándo hacer pública su información.» |
| Ficha visible sin nombre | «Para publicar la ficha, escribe primero el nombre del encargado.» |
| Sección sin publicados | «Nada está publicado ahora: el público no ve esta sección ni el bloque de la portada, y `/ministerios` redirige a la portada, hasta que publiques al menos un ministerio.» |
| Error del sistema al guardar | «No se pudo completar la operación. Tus escritos están a salvo; vuelve a intentarlo en unos minutos.» |
| Sin permiso (servidor) | «No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.» (F2) |
| Ayuda del orden | «Los ministerios se muestran en orden: menor número primero. Con empate, el más antiguo va antes.» |
| Ayuda del horario | «Escribe el horario tal como quieres que lo lean; por ejemplo: «Todos los viernes, 7:00 p. m.». No se calculan fechas ni calendarios.» |
| Ayuda del lugar | «Opcional. P. ej. «Salón principal» o «En la casa de la familia Pérez».» |
| Ayuda del contacto (tipo) | «Puedes dejarlo sin contacto, o elegir celular o correo (uno solo). Al cambiar el tipo se borra el contacto anterior.» |
| Ayuda «English (opcional)» | «Si lo dejas vacío, el público verá este contenido en su versión en español.» |
| Ayuda de imagen | «Opcional. Añade el texto alternativo para personas que no la pueden ver. El logo de la iglesia respeta las reglas de marca.» |

### 7.2 Sitio público — cadenas de interfaz es/en

> El contenido (nombre, descripción, encargado, horario, lugar y `alt`) viene del panel con fallback del servidor (FR-008). Estas son las palabras de la interfaz.

| Clave | Español | English |
|---|---|---|
| `nav.sections.ministries` | Ministerios | Ministries |
| `ministries.page.title` | Ministerios | Ministries |
| `ministries.page.subtitle` | Las áreas de servicio de nuestra iglesia: conoce cada una y cómo participar. | The service areas of our church: get to know each one and how to take part. |
| `ministries.leader.label` | Encargado | Led by |
| `ministries.join` | Unirse al grupo de WhatsApp | Join the WhatsApp group |
| `a11y.ministries.join` | Unirse al grupo de WhatsApp de {nombre} | Join the WhatsApp group of {name} |
| `ministries.loading` | Cargando… | Loading… |
| `ministries.error.load` | No pudimos cargar los ministerios. Revisa tu conexión y vuelve a intentarlo. | We couldn't load the ministries. Check your connection and try again. |
| `ministries.error.retry` | Volver a intentar | Try again |
| `ministries.error.partial` | No pudimos actualizar; ves la última versión disponible. | We couldn't refresh; you're seeing the latest available version. |
| `a11y.ministries.heading` | Ministerios de la iglesia | Church ministries |
| `home.ministries.title` | Ministerios | Ministries |
| `home.ministries.subtitle` | Las áreas donde puedes servir | The areas where you can serve |
| `home.ministries.seeAll` | Ver más ministerios | See more ministries |
| `home.ministries.badge` | Ministerio | Ministry |
| `a11y.home.ministries.heading` | Ministerios de la iglesia: mira el resumen y entra a la sección completa | Church ministries: quick look and link to the full section |

> El `alt` de las imágenes es obligatorio en el panel (español, inglés opcional): no existe cadena de fallback automática en el público.
> La redirección de la sección oculta (D-14) **no añade ninguna cadena**: no hay aviso que mostrar; si la consulta tarda, la única palabra en pantalla es `ministries.loading` («Cargando…» / «Loading…») sobre el esqueleto. La antigua clave `ministries.empty.hidden` de F4 no se replica aquí.

---

## 8. Accesibilidad (WCAG 2.1 AA) e internacionalización

### 8.1 Accesibilidad

- **Teclado**: skip link → logo → navegación (enlaces de barra; menú móvil disclosure de F10) → idioma → tema → contenido → pie. El botón de unirse y el título-enlace de las tarjetas reciben el anillo de foco de F10; nada depende de un gesto. En el panel, `Esc` respeta D-18 (sin descartar cambios sin aviso).
- **Foco visible** (anillo 2 px por tema) en todo interactivo, incluidos el botón de unirse y el título-enlace del extracto; un elemento aún no revelado que recibe foco se hace visible de inmediato (guarda de `Reveal`, F10).
- **Semántica**: un `<h1>` por página («Ministerios» en la sección; el bloque de portada usa `h2` con `SectionHeading`, al nivel de los títulos de F3/F4); cada ficha es un `article`/`li` con `h3`; el listado público es `<ul>` con orden de lectura = orden manual; tablas solo en el panel. **Sin** `<time>` en F6: el horario es texto (FR-024) y se lee como texto.
- **No solo color**: lo que distingue un ministerio de un evento en la portada es el **bloque** y el chip de texto («Ministerio»), no un color; los estados del panel llevan siempre texto (`StatusPill` con palabra).
- **Imágenes**: `alt` es/en gestionado por el equipo (obligatorio en español) (FR-020, FR-023); si falla, la ficha no se quiebra (`onError`, mínimo cambio de layout). El logo solo en el chrome (F10) y según el Manual.
- **Enlaces externos con nombre completo**: «Unirse al grupo de WhatsApp de {nombre}» (lectores de pantalla); el contacto del encargado es texto, nunca un enlace que el equipo no aprobó. ⟲ D-4
- **Interruptor del panel**: `role="switch"` + `aria-checked` + etiqueta visible («Mostrar la ficha del encargado en el sitio público»); su estado es teclado-operable y su leyenda contextual se anuncia al cambiar (el cambio se hace visible como texto, no solo color).
- **Contraste y letra**: cuerpo 18 px; Bebas de rótulos con contraste AA en los dos temas; verde de WhatsApp con su par de contraste aceptado del sistema; el turquesa como texto pequeño siempre en su clase `#0e7a66` sobre claro.
- **Zoom/200 % y 320 px**: diseño fluido en `rem` con `clamp`; 0 desplazamiento horizontal (SC-007); el extracto de la portada pasa de fila de 5 → grilla de 2 → columna única por breakpoint (D-28 de F4).
- **Objetivos táctiles ≥ 44 px**: botón de unirse (48 px en móvil), título-enlace del extracto (caja de 44 px, como `UpcomingItem`), botón del bloque, ítems del menú móvil; la tarjeta no es un enlace en bloque.
- **Avisos**: cambio de idioma con `role="status"` de F10; en el panel, `Notice` con su rol correcto.
- **`prefers-reduced-motion`**: sin revelado ni stagger; contenido íntegro al instante (F10 incondicional; F6 no añade movimiento nuevo).
- **Idioma del documento**: `html lang` por F3/F10; el contenido en fallback dentro de una página en inglés va con `lang="es"` en su envoltorio (práctica heredada de F3, ⟲ mismo criterio D-21 de F4).

### 8.2 Internacionalización (Decisión 6 y Decisión 8)

- **Interfaz** (siempre bilingüe): diccionario `es.ts`/`en.ts` de F3; `LanguageProvider` y `localStorage` intactos; primera visita en español, sin auto-detección (FR-009). El panel **no** cambia de idioma (FR-018).
- **Contenido** (por campo): español obligatorio solo si el campo se rellena; inglés opcional; el fallback es→es lo resuelve el servidor (FR-008); el frontend nunca muestra un campo vacío ni traducción automática. «Un campo vacío no obliga a nada» (Q6): horario, lugar y encargado pueden no existir y su falta no deja huella.
- **El enlace de WhatsApp y el contacto no son contenido traducible** (US4 escenario 5 · FR-008): iguales en los dos idiomas.
- **Horario y lugar**: textos puros, sin `Intl` ni formateo — se muestran tal como el equipo los escribió, con su versión por idioma solo si el equipo cargó el inglés.

### 8.3 Privacidad del encargado — cómo se verifica y quién la garantiza (FR-026, SC-014)

- **La garantía es del servidor**: con el interruptor apagado, el DTO público del ministerio **no trae** nombre ni contacto (null + omisión de la clave), y el HTML SSR/JSON de cualquier vía (sección, bloque de portada, API forzada sin sesión, metadatos) no los contiene. **El frontend no implementa la ocultación ni como refuerzo de seguridad** — solo presenta lo que llega: si el DTO no tiene encargado, `EncargadoLine` no se monta y no hay línea.
- Lo que el panel hace es **explicarlo**: leyendas del §4.5 y etiqueta «Ficha oculta» del listado, de modo que el equipo nunca confunda «guardar un encargado» con «publicarlo». Con el default **apagado** (D-13) esta separación es además estructural: un encargado recién guardado nunca es público; publicarlo es siempre un acto expreso del equipo (reactivar el interruptor). Esa es la nota explícita de garantía que pide la spec (FR-025: dato guardado ≠ dato publicado).
- **Criterio QA/seguridad**: con el interruptor apagado, una búsqueda exacta del nombre y del celular/correo en el HTML y el JSON del sitio público debe dar **0 resultados** (inspección de red, respuestas forzadas). Está en §10 como punto de prueba.

---

## 9. Decisiones de UX para revisar por el humano (resumen)

| # | Decisión | Alternativa |
|---|---|---|
| D-1 ⟲ | «Ministerios» es una **ruta real** al final de la única lista de enlaces de F10, después de «Eventos y actividades» (mismo patrón aprobado de F4 D-1; la posición al final respeta la secuencia de la portada: bloques ruta al final). | Entrada junto a los anclas de portada (dejaría la mezcla en otro orden) o dos entradas (descartado: una sola navegación). |
| D-2 ⟲ | Subtítulo de `/ministerios` como **cadena de interfaz fija** (no campo del panel). | Que el equipo lo escriba en el panel (más contenido que administrar). |
| D-3 ✅ | Reutilizar `CardImage` de F4 (con posible **promoción a pieza común**) para la imagen 16:9 de fichas y extracto — **resuelta por el humano el 2026-10-13** (el detalle de la promoción sigue siendo del arquitecto). | Copiar la pieza en `features/publico/ministerios/` (duplicaría comportamiento ya probado). |
| D-4 ⟲ | El **contacto del encargado se muestra como texto** (celular/correo con su icono), sin `tel:`/`mailto:`. | Enlaces de llamada/correo (añaden semántica útil pero activan apps sin decisión del equipo; el texto es copiable igualmente). Reabrible si el humano lo quiere. |
| D-5 ⟲ | Botón de unirse en **verde de WhatsApp** (`PublicButton` variante propia del CTA verde del sistema) y **solo con enlace** (Q2). Siempre el mismo verde del CTA de WhatsApp del sitio. | Botón secundario neutro (más apagado) o enlace de texto (menos descubrible; FR-003 pide «claramente identificado»). |
| D-6 ⟲ | Reordenamiento por **campo numérico «Orden»** en el formulario + columna en el listado (precedente `sort_order` de F3/F4; sin drag & drop ni flechas). | Editor de flechas/bajar-subir en la fila (más toques, requiere otro patrón no existente) o drag & drop (innecesario para listas cortas). |
| D-7 ⟲ | Mismo enlace de WhatsApp en dos ministerios: **aviso no bloqueante** tras guardar (default revisable de la spec). | Bloquear el guardar (dura para un grupo de distribución compartido) o silencio (pierde la «señalización» que pide la spec). |
| D-8 ⟲ | Textos y ubicación de la leyenda de privacidad del interruptor (§4.5): leyenda verde cuando se publica, `Notice` warning al apagar y etiqueta «Ficha oculta» en el listado. | Solo el interruptor sin leyendas (cumple FR-025 pero deja a la persona sin entender qué se hace público). |
| D-9 ⟲ | El bloque de portada **sí muestra esqueleto mientras carga** (aun pudiendo desaparecer con 0 ítems): el caso 0 es raro y el brinco es aceptado. | Sin esqueleto (el bloque aparece/desaparece en seco; brinco cuando hay contenido — peor). |
| D-10 ⟲ | Error **sin caché** en el bloque de portada: el bloque **no se dibuja** (regla general de F3), sin estado de error propio — distinto de F4, cuyo estado vacío con mensaje es la excepción ya firmada. | Reproducir el estado de error de `ProximamenteBlock` (mostraría un título + aviso de error en la portada por un fallo de red: necesidad discutible y es la vía que FR-028 evita). |
| D-11 ⟲ | Chip «Ministerio» en el extracto de portada, neutral (patrón F4 D-25/D-2). | Sin chip (la tarjeta se confundiría con las de eventos en una portada con dos bloques contiguos del mismo patrón). |
| D-12 ⟲ | El extracto **no lleva ni encargado ni botón de unirse**: tarjeta con imagen + chip + título-enlace + horario/lugar; la acción de unirse vive en la sección. | Botón de unirse por tarjeta (más alegría inmediata, pero duplica el CTA y rompe la simetría con el bloque de F4, donde el extracto también solo enlaza). |
| D-13 ✅ | La ficha del encargado inicia **APAGADA por defecto**: no se publica en el sitio hasta que el equipo encienda el interruptor expresamente. Decisión literal del humano 2026-10-13 («Oculto por defecto»), que cierra el default de Q1 (antes propuesto encendido). El formulario lo dice con claridad: Notice warning al apagado (caso de fábrica) + ayuda fija del interruptor + leyenda verde solo al encender. | Encendido por defecto (propuesta original de este documento; descartada por el humano). |
| D-14 ✅ | `/ministerios` con 0 publicados **redirige a la portada** (`<Navigate to="/" replace>`, sin aviso, sin parpadeo — §4.1). Decisión literal del humano 2026-10-13 («Redirigir a la portada»). Es **distinto a propósito** de `/eventos` de F4 (que muestra `HiddenNotice` sin redirigir): no unificar los dos comportamientos. | Aviso sin redirigir (patrón D-13 de F4, lo propuesto originalmente en este documento; descartado por el humano). |

---

## 10. Notas para el `dev-frontend` / `qa-tester`

- **El bloque de portada usa su propio hook/endpoint** (`usePortadaMinisterios` → propuesta `GET /api/v1/portada/ministerios?lang=es|en`): tope 5 en servidor, orden manual, DTO mínimo, `no-store`. Del cliente solo llegan la presentación, el esqueleto y el aviso parcial; el tope y el corte `sort_order, id` son lógica del servidor. Pruebas contadas: 1, 2, 3, 4, 5 y **0 ítems** (bloque ausente, sin mensaje ni título — SC-016).
- **La presencia en la navegación sale del bloque, no de `/ministerios`**: `usePortadaMinisterios` alimenta `MinisteriosBlock` y, compartida por `queryClient`, también la entrada en `useSiteLinks` (una sola petición paga bloque + barra + menú + pie). *(Ajuste 2026-10-13, puerta 2 — autoridad: T628, P6-9, R6-8; criterio confirmado por el humano — la propuesta D-23 de F4, que hacía deducir la presencia de la consulta pública de la sección con una única petición para página + barra + pie, quedó superada. El observable no cambia: misma regla «hay al menos un ministerio publicado» — pero son **dos consultas**: bloque + sección; el nombre exacto de hooks/endpoint queda al contrato del arquitecto.)* Con 0 publicados **ninguna superficie enlaza la sección** y la ruta directa **redirige a la portada**: la decisión de redirigir la toma la página con su propia consulta, `usePublicMinistries` (FR-005, SC-010, D-14).
- **D-13 (default del interruptor, crítico)**: al crear un ministerio con encargado, el interruptor llega **apagado**; con él apagado el público no recibe nada del encargado (SC-014) y el `Notice` warning del §4.5 también es el primer texto que ve el equipo; al encenderlo y guardar se republica sin reescribir (US2 esc. 10–11); editar nunca cambia el default (la última decisión manda).
- **D-14 (redirección, crítica)**: `/ministerios` con 0 publicados → redirección limpia a `/` con `replace`: **sin parpadeo** — nunca se pinta `HiddenNotice`, ni H1, ni grilla vacía, ni mensaje de error (con la consulta ya en caché, redirige en el primer render; sin caché, solo el esqueleto estándar mientras se decide); el historial queda limpio (no queda `/ministerios` tras la portada en *Volver*); y QA confirma que **`/eventos` conserva su `HiddenNotice`**: los dos comportamientos son deliberadamente distintos (D-14), no se «unifican».
- **SC-014 (crítico)**: QA verifica **0 apariciones** del nombre y del contacto del encargado con el interruptor apagado: HTML de `/ministerios`, HTML/JSON de la portada, JSON público del módulo y del bloque, y una respuesta forzada sin sesión; también que el panel conserva los datos y que reactivar el interruptor los republica sin reescribirlos (US2 esc. 10–11). El frontend no oculta nada por su cuenta: si el DTO trae el dato, se muestra (así el diseño nunca esconde un bug del backend — el test de seguridad lo detecta).
- **SC-017 (casos sin dato)**: listado de combinaciones a probar en público — sin enlace (0 botones), sin encargado publicado (0 líneas), ficha visible sin contacto (solo nombre), sin imagen / imagen fallida (0 huecos, layout estable), sin horario/lugar (0 líneas), combinación total «solo nombre + descripción». Cero «—», cero líneas vacías, cero botones deshabilitados por vacío.
- **SC-013 (concepto)**: la sección y el bloque no contienen nada del modelo de F4/F5 (FR-001): ni `DateBadge`, ni fechas calculadas, ni contenido de agenda; el horario aparece exactamente como texto del equipo.
- **Orden**: prueba de `sort_order, id` (no por nombre) en público y en el panel; cambiar el valor del campo Orden y guardar reordena el público al recargar (FR-016).
- **El diseño no cambia el sistema de F10**: `Card`, `SectionHeading`, `Reveal`, `useSiteLinks`, i18n, temas y tokens tal cual; 0 tokens nuevos; la página nueva es «otra sección» y no una marca propia.
- **Datos del panel con Zod** espejo del service; NUNCA `fetch` en componentes (hooks de `features/ministerios/hooks/`); tipos desde `backend/api/openapi.yaml` (sin `any`).
- **Imágenes**: `ImageUploader` de F3 con `alt` obligatorio en español; archivos servidos por `GET /api/v1/media/{fileName}` solo si el ministerio que los usa está publicado (404 si no: patrón F3 §10); al fallar, la ficha pública no se rompe.
- **Textos SIEMPRE del diccionario en el sitio público** (`features/publico/i18n`), nunca JSX suelto (SC-006).
- **SC-008** se verifica con checklist manual de QA en ambos temas (claro/oscuro), teclado, lector de pantalla, 320 px y zoom 200 % — mismas bases de F3/F10/F4.
- **El panel no se rediseña** (F10B): componentes de F2/F3/F4 tal cual; el acceso de InicioPage en `AccesosRapidos` con `hasPermission('ministerios')`.
- **Verificar contra el Manual de marca** en la revisión visual: fuentes, paleta 70/20/10, logo intacto, verde de WhatsApp del sistema en el botón de unirse, foco con el anillo del sistema.

---

## 11. Trazabilidad — qué pantalla cubre qué US/FR

| US / FR | Pantalla / componente | § |
|---|---|---|
| US1 · FR-002 · FR-003 · FR-004 · FR-005 · FR-006 | `/ministerios` (MinisteriosPage, MinistryCard, WhatsappJoinButton, EncargadoLine) — fichas completas, orden manual, unirse en un toque; 0 publicados → redirección a la portada sin parpadeo (D-14) | §4.1, §2.1 |
| US2 · FR-010 · FR-011 («ministerios») · FR-014 · FR-016 · FR-018 · FR-019 | `/panel/ministerios` (MinistriesPage, MinistriesList, MinistryForm con ContactoEncargado y ShowContactSwitch), InicioPage accesos, orden numérico | §4.2, §4.3, §4.4, §4.5, §5, §2.3 |
| US3 · FR-012 · FR-013 · FR-017 | PublishControls por ministerio, ConfirmDialog, sin publicados: barra+mapa sin la entrada y `/ministerios` redirige a la portada (D-14), bloque de portada oculto, auditoría | §2.4, §4.3, §4.8, §5.c |
| US4 · FR-007 · FR-008 · FR-009 | LanguageSwitcher (F10), contenido con fallback por campo (encargado/horario/lugar incluidos), enlace no traducible | §2.2, §7.2, §8.2 |
| US5 · FR-021 · FR-022 · FR-023 | Sistema F10 completo en sección y bloque, teclado/lector, 320 px, 200 %, reduce-motion, 18 px y 44 px | §4.1, §4.9, §8.1 |
| FR-024 (horario/lugar texto) | Líneas de texto en ficha (sin `DateBadge`, sin calendario), formulario con ayuda y contadores | §4.1, §4.4 |
| FR-025/FR-026/SC-014 (encargado + privacidad) | Bloque Encargado del formulario (ContactoEncargado + ShowContactSwitch + leyendas), marca «Ficha oculta» en el listado, garantía server-side | §4.5, §8.3, §10 |
| **US6 · FR-027 · FR-028 · FR-029 (Q7)** | Bloque «Ministerios» en la portada (MinisteriosBlock con MinistryItem vía usePortadaMinisterios, tope 5 en servidor, tras «Lo que viene» y antes de «quiénes somos»), oculto por completo sin publicados, botón a `/ministerios` | §2.7, §4.9 |
| Q5 (orden manual) | Campo Orden + columna; ordenación `sort_order, id` en público | §4.3, §4.4, §1 |
| Auditoría (FR-017) | Módulo Auditoría de F2 intacto; no añade UI nueva; registra también el interruptor | §4.8 |
