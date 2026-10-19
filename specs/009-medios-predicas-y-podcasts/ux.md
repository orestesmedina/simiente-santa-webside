# UX: F9 — Medios: prédicas y podcasts

> Documento de diseño para `disenador-ux`. Fuente: `spec.md` de F9 **v2 (2026-10-17, modelo aprobado por el humano)** —importación desde los feeds de YouTube y del podcast, enlaces externos en pestaña nueva anunciada, panel reducido—; patrones de `specs/007-donaciones/ux.md` (zona de textos/listados del panel, bloque textual de portada, redirección de sección oculta, decisiones de «editar» vs. «publicar»), `specs/004-eventos-actividades/ux.md` (ficha con imagen, bloque «Lo que viene», estados del panel), `specs/010-sistema-diseno-movimiento/ux.md` (sistema visual «Marea Turquesa», navegación, temas y movimiento) y el código actual de `frontend/src/` (React 19 + Tailwind 4; componentes de F2–F7/F10 ya construidos y verificados en este repositorio: `Card`, `CardImage`, `SectionHeading`, `PublicButton` con `external`, `StatusPill`, `PublishControls`, `Dialog`, `ConfirmDialog`, `EmptyState`, `Notice`, `Table`, `Tabs`, `Reveal`/`RevealGroup`, `SectionSkeleton`, `useSiteLinks`, `MinistriesPage` con su redirección verificada).
> No es código: lo implementa el `dev-frontend`. Rutas y contratos los fija el `arquitecto`; aquí se proponen y los supuestos van marcados.
> **Decisiones de UX a revisar por el humano en la puerta 2**: ver §9 (resumen). Cada una lleva su marca ⟲. Reutilizo los patrones ya aprobados (ruta real al final de la lista de navegación de F4/F6/F7, redirección a la portada de F6/F7, fila de bloque de portada de F4/F6/F7, tabla/tarjetas del panel) y aquí solo marco lo nuevo de F9.
> F9 **aplica** el sistema de F10 y los patrones de F2–F7; **no redefine** tokens, navegación, selector de idioma, modo oscuro ni el rediseño del panel (F10B).
> **Distinción de nombres (Q7/Q13)**: la **sección pública** es **«Medios»** en `/medios`; el **módulo y el permiso del panel** se llaman **«medios» / «Medios»** (aquí los dos nombres coinciden; convención de F7/F6 aplicada igual).
> **Regla de oro del diseño (Q15/FR-004)**: **el sitio nunca carga nada de terceros y nunca reproduce nada dentro de sí.** La carátula es una copia servida por el propio sitio; el video y el audio viven en YouTube y Spotify y el visitante **sale del sitio** por un enlace **anunciado** que abre **pestaña nueva**. Ninguna pieza de este diseño depende de un recurso de YouTube, Spotify ni Anchor (SC-005: 0 cargas de terceros).

---

## 0. Principios de diseño (heredados, aplicados a F9)

1. **La salida del sitio se anuncia antes de darla.** Todo enlace externo se lee antes de activarlo: dice **a qué servicio lleva** y **que se abre en una pestaña nueva** (visible y para lector de pantalla). La pestaña del sitio **sigue abierta**: una persona mayor nunca «pierde» el sitio de la iglesia al tocar el enlace (FR-005, default revisable #9, SC-012).
2. **Cero contenido de terceros — y se nota.** Sin reproductores incrustados, sin peticiones a YouTube/Spotify/Anchor, sin avisos de cookies de terceros. Las carátulas son **copias servidas por el sitio** (FR-004). La ficha es 100 % propia: carátula (`<img>` local), texto y botones; el único punto de salida es el `href` del enlace que la persona decide activar.
3. **Una ficha por sermón.** El mismo sermón en video y en audio es **una sola ficha con los dos enlaces**; el visitante jamás ve el mismo sermón dos veces (Q18, SC-003). Con un solo formato → solo ese enlace, sin líneas vacías ni «sin datos» (US1 esc. 3).
4. **El contenido entra solo; el equipo solo revisa.** El equipo **no llena campos** (Q3): pulsa un botón y corrige a lo sumo la descripción corta. El panel no ofrece alta manual, ni edición del nombre/carápula/fecha/enlaces, ni borrado (FR-019): un episodio que no debe verse **se oculta**.
5. **Lo decidido por el equipo es intocable.** Ocultar, mostrar, unir y ajustar la descripción sobreviven a cualquier búsqueda (FR-015, SC-007). El diseño lo hace visible: el panel dice siempre el efecto de cada acción y el resumen deja claro qué cambió y qué no.
6. **Nunca se oculta por falta de certeza.** El episodio que sale de la ventana del feed **sigue visible** si su enlace sigue disponible (o si la comprobación no se pudo completar) y solo queda **marcado para revisión** (FR-017, regla aprobada antes de la puerta 1). Las tres marcas se distinguen por **nombre y explicación**, nunca por color solo.
7. **Apto para todas las edades.** Cuerpo 18 px, objetivos ≥ 44 px, iconos siempre con texto, lenguaje sin tecnicismos («canal», «feed», «idempotente» **no aparecen** en la UI: se dice «YouTube», «el podcast», «es seguro pulsarlo otra vez»), 0 desplazamiento horizontal a 320 px (`idea.md` §7).
8. **Solo lo visible existe.** Sin ningún episodio visible, la sección **no existe** — ni barra, ni menú, ni mapa del pie —, el bloque de la portada **no se dibuja** y `/medios` **redirige a la portada** limpia y silenciosamente (Q10, patrón de `/ministerios` y `/ofrendas`). Cero secciones vacías, cero enlaces huérfanos.
9. **El panel hereda F2/F3/F4/F6/F7 tal cual.** Título + banda de ayuda, `Table`/tarjetas, `StatusPill`, `Dialog`, `ConfirmDialog`, `Notice`, `Tabs` de idioma, validación con Zod y mensajes en español (FR-024): cero estilos nuevos. F9 no rediseña el panel (F10B).

---

## 1. Tokens y componentes aplicados (la verdad vive en F3/F10; solo se citan las decisiones de F9)

| Tema | Reutilizado de | Nota específica de F9 |
|---|---|---|
| Paleta 70/20/10 | F3 §1.1 / F10 §0.1 | Navy dominante, `Card solid` por ficha, acento turquesa en rótulos y foco. Sin colores nuevos. |
| Tipografías | F3 §1.2 / F10 §0.2 | H1 de página y rótulos de bloque en Bebas; nombre del episodio en Poppins 600; descripción y fecha en Poppins regular. |
| Espaciado | F10 §0.3 | `--space-section` entre bloques vecinos de la portada y del pie; `--space-block` entre fichas; el bloque de portada dentro de un `Card glass` con `--pad-panel`, como F6/F7. |
| Superficies | `Card` (`solid`/`glass`) | Ficha pública: `Card solid` `as="li"`; bloque de portada: `Card glass`, igual que `MinisteriosBlock`/`OfrendasBlock`. |
| Carátula | `CardImage` (inventario único) | 16:9 a sangre en el marco superior. **Ajuste**: la imagen siempre es local (copia servida por el sitio). Si el episodio **no tiene carátula**, el servidor lo indica y se muestra el **marcador neutro** (§6.3): mismo marco 16:9 con la identidad de la iglesia — nunca un hueco y nunca una petición a terceros. El `onError` de `CardImage` se mantiene como red de seguridad: si la copia falla, se retira y entra el marcador neutro. |
| Movimiento | F10 §5 (`Reveal`/`RevealGroup`) | Revelado lateral una sola vez por página/bloque, stagger suave (tope 240 ms). Hover/active de los enlaces con el token micro. Con `prefers-reduced-motion`: todo íntegro e instantáneo. |
| Foco visible | F3 §1.4 / F10 §6 | Anillo 2 px: enlaces externos, «Cargar más», «Ver todo» y todo control del panel. |
| Fechas | `formatoFechas.ts` de F4 (`formatLongDate`) | «17 de mayo de 2026» / («May 17, 2026»); la interpretación en `America/Costa_Rica` la resuelve el servidor (FR-006). Sin `DateBadge` ni marca «Pasado» (nada vence). |
| Orden | nuevo en este catálogo | **Automático por fecha descendente** (más reciente primero, FR-006): **sin `sort_order` ni columna de orden en el panel** — primera funcionalidad de listado sin orden manual (Q9) ⟲ D-11. |
| Píldoras | `StatusPill` con `label` | Estado: **«Visible»** (verde) / **«Oculto»** (grafito). Marcas: **«Posible duplicado»** (ámbar), **«No disponible»** (rojo), **«Ya no aparece en el feed»** (grafito). Siempre texto + color (nunca color solo). |
| «Cargar más» | `PublicButton` | **No se usa `Pagination`** (es por páginas): «Cargar más» agrega la siguiente tanda a la misma lista (Q11). |

**Nuevos tokens: ninguno.** F9 usa exclusivamente los de F3/F10.

**Del inventario NO se usa** (con razón de negocio): `CategoryChip` (no hay catálogo), `DatesEditor`/`DateRangeFilter` (las fechas vienen del feed y no se editan), `MapLink` (no hay lugares), `ImageUploader` (la carátula viene del feed: 0 carga manual de imagen — Q22 sin objeto), `NumeroDonacion`/`CopyAccountButton` (nada que copiar aquí), `PublishControls` (publicar/retirar es vocabulario de portada; aquí el estado es visible/oculto con otra mecánica ⟲ D-8), `WhatsappJoinButton` (sin grupos), `HiddenNotice` (la sección oculta **redirige**, no avisa — patrón de F6/F7), `Pagination` (Q11 eligió «Cargar más»).

---

## 2. Flujo de usuario

### 2.1 Visitante: encuentra las grabaciones y las abre donde están (US1, US5 — camino feliz)

1. Abre el sitio. En la barra de escritorio (o el menú hamburguesa en móvil, o el mapa del pie) está **«Medios»** — al final de la lista, después de «Ofrendas y diezmos» (FR-007) — solo cuando hay algo visible (FR-008).
2. La toca → llega a **`/medios`**: **H1 «Medios»** y un subtítulo fijo que ya **anuncia la salida del sitio**: «Las prédicas y los podcasts de nuestra iglesia; se abren en YouTube y Spotify, en una pestaña nueva.» (US5 esc. 2/3: la persona entiende qué es la sección y qué harán los enlaces antes de activar nada).
3. Recorre la lista de fichas — **una columna** en móvil, **dos** en tableta/escritorio — del **más reciente al más antiguo** (FR-006): cada una con **carátula** (servida por el propio sitio), **nombre** tal como lo trae el feed, **fecha** en letra pequeña y la **descripción corta** (FR-020).
4. En cada ficha ve los **botones de enlace** según sus formatos: **«Ver en YouTube»** (si hay video), **«Escuchar en Spotify»** (si hay audio), o los dos (Q2/Q8: los formatos se presentan con los enlaces, no en bloques separados). Cada botón lleva indicación visible de pestaña nueva y su nombre accesible lo anuncia completo (FR-005).
5. Toca el botón → **se abre una pestaña nueva** con YouTube/Spotify; la del sitio **sigue abierta en el mismo lugar** de la lista. El sitio no cargó nada de esos servicios en ningún momento (SC-005).
6. Si hay más episodios, al final ve **«Cargar más»** (tanda de 12, default revisable #5): cada toque agrega la siguiente tanda al final de la misma lista, sin recargar la página y sin perder el lugar.
7. Termina y recorre el resto: pie de F10 intacto; el mapa del sitio muestra «Medios» en el grupo **Explorar** (solo si hay algo visible).

### 2.2 Visitante: cambia a inglés (US4)

1. En la barra (o el pie), toca **English** en el selector de F10.
2. La **interfaz** cambia de inmediato: «Media», «Watch on YouTube», «Listen on Spotify», «Load more» van del diccionario es/en.
3. El **contenido** se muestra con el respaldo de la Decisión 8: si la descripción corta tiene **versión en inglés** (traducida a mano en el panel), se muestra en inglés; si no, se muestra **en español** — el nombre es el del feed y va tal cual — **sin hueco, sin marca de «sin traducir», sin traducción automática** (FR-025). Lo demás **no cambia en absoluto**: carátula, nombre, fecha, enlaces y formatos son idénticos en ambos idiomas (FR-026, US4 esc. 5).
4. La elección persiste al navegar (portada incluida) y entre visitas; `html lang` pasa a `en`.

### 2.3 Equipo: publica el sermón de la semana sin escribir nada (US2 — «Buscar episodios nuevos»)

> Flujo completo (antes / durante / después) en §4.4. Resumen:

1. Inicia sesión → **Inicio** del panel: un acceso en `AccesosRapidos` — **«Medios»** (con permiso «medios») → `/panel/medios`. Sin el permiso: no ve el acceso y la ruta está protegida (FR-021).
2. En la página ve, arriba, la **zona de importación**: título de la zona + botón grande **«Buscar episodios nuevos»** + línea de apoyo. Debajo, la **lista del catálogo** (la primera vez: vacía, con el mismo botón como acción principal).
3. Pulsa el botón → **durante**: el botón queda deshabilitado en «Buscando…» con un mensaje de contexto (`role="status"`); la lista sigue visible. La búsqueda corre **dentro de la petición** (sin proceso de fondo, FR-011).
4. **Después**: un **resumen** aparece entre la zona de búsqueda y la lista (`Notice`, §4.4): cuántos episodios entraron, cuántos se unieron (video + audio en una ficha), cuántos quedaron **marcados como posible duplicado** y qué feed no respondió, si alguno. La lista se refresca con lo nuevo en su lugar por fecha.
5. Si todo quedó unido, no hay más que hacer — el contenido ya es visible en `/medios` y, si entra entre los últimos, en el bloque de la portada. Si hay marcas de revisión, las resuelve (§2.4).

### 2.4 Equipo: mantiene el catálogo (US3)

- **Ocultar / mostrar**: botón en cada fila. *Ocultar* desaparece el episodio de `/medios` y del bloque de la portada al momento; *Mostrar* lo devuelve a **su lugar por fecha** (FR-015). Un episodio oculto **sigue oculto** tras cada búsqueda (SC-007) — el panel lo dice en la propia interfaz (§4.7).
- **Unir duplicados**: en cada fila con marca «posible duplicado» hay **«Revisar duplicado»** → un `Dialog` muestra las dos candidatas lado a lado y **la ficha resultante** de la unión (§4.6); el equipo confirma **definitivamente** (doble paso con `ConfirmDialog`) o resuelve que **no son el mismo sermón** → las dos quedan visibles y sin marca. Sin duplicados pendientes, la acción sencillamente no existe en ninguna fila (§4.6).
- **Ajustar la descripción corta**: **«Editar descripción»** en cada fila → `Dialog` con pestañas de idioma (Español / English opcional), el valor vigente **precargado** y editable, contador con el límite (§4.5). Al guardar, **el texto del equipo manda** y sobrevive a las búsquedas (FR-020, SC-007): aviso inmediato «Cambios guardados: ya visibles en la sección.» (FR-015: se publica al guardar, patrón F3).
- **Marcas resolubles** (§4.7): «posible duplicado» se resuelve **uniendo** o confirmando que son distintos; «no disponible» (el episodio quedó oculto) se resuelve **mostrándolo de nuevo** (con confirmación suave) o **dejándolo oculto**; «ya no aparece en el feed» (el episodio **sigue visible**) se resuelve **manteniéndolo visible** (la marca se cierra) u **ocultándolo**. Una marca resuelta **no se reabre por la misma causa**, y la interfaz del panel lo dice al resolver (US3 esc. 6).
- Toda operación queda en la **auditoría** de F2 sin UI nueva (FR-022; §4.10).

### 2.5 Errores del equipo

| Situación | Camino |
|---|---|
| «Buscar episodios nuevos» y un feed no responde | El resumen informa **qué feed falló**, lo que sí se importó del otro, y avisa de volver a intentarlo. **Nada se oculta ni se marca por esa causa** (FR-016); repetir la búsqueda recupera lo que faltaba. |
| «Buscar episodios nuevos» sin novedades | Resumen neutro: «No hay episodios nuevos: nada cambió.» La lista queda igual, sin duplicar nada (FR-012). |
| Búsqueda interrumpida (se corta la red') | Nada importado a medias (FR-023: nunca cambios parciales); la lista queda como estaba; `Notice` de error genérico para reintentar. |
| Dos elementos del mismo día que ni normalizando coinciden | Entran **separados**: la primera importada queda visible, la adicional **oculta**, y ambos quedan **marcados «posible duplicado»** (FR-014); el equipo los une o los resuelve como distintos (§4.6). |
| El enlace se comprueba **ya no disponible** | El episodio se **oculta del público** al momento y queda **marcado «No disponible»** con su explicación (Q16/FR-017). Nunca se muestra un enlace roto como si funcionara (SC-014). |
| La comprobación de disponibilidad no se puede completar | El episodio **sigue visible**, queda **marcado «Ya no aparece en el feed»**, y el panel aclara: «no se pudo comprobar si su enlace sigue funcionando» (FR-017, regla a prueba de fallos). |
| Descripción corta por encima del límite | El contador bloquea y avisa: «{n} de {L} caracteres. Recorta {m} para poder guardar.» (límite del default revisable #1). |
| Versión en inglés rellena **sin su español** | **No permite guardar**: «El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de la descripción corta.» (FR-023/FR-025). |
| Episodio sin carátula en su feed | Marcador neutro en la ficha pública y en el panel (cero huecos, patrón F6/F7). |
| Error inesperado del sistema al importar o gestionar | Aviso genérico sin detalles internos; **ni la lista ni lo importado se pierden ni se alteran**. |
| Operación sin permiso «medios» | Ruta protegida (`RequirePermission` de F2) + página «sin permiso»; rechazo en caliente: mensaje claro sin detalles (FR-021). |
| Dos personas usando el panel a la vez | Se conserva la **última versión guardada completa** en cada operación (default revisable #12). |

### 2.6 Persona sin permiso (el caso desfavorable)

- Sin permiso «medios»: no existe el acceso «Medios» (ni en `AccesosRapidos` ni en la navegación del panel); si fuerza la URL `/panel/medios`: `RequirePermission` → `/sin-permiso` (F2, texto intacto).
- Tener «medios» no da los otros módulos, y ningún otro permiso da acceso a este (Decisión 5). F9 no crea permisos nuevos ni toca el catálogo (sigue en 10) ni el nombre de este (Q13/FR-021).
- El sitio público no pide sesión: no opera el caso «visitante sin permiso».

### 2.7 Visitante: descubre las grabaciones desde la portada (US6, Q12, FR-027)

1. Recorre la portada; **al final, después del bloque «Ofrendas y diezmos» de F7** (Q12), ve el bloque **«Medios»**: panel de vidrio rotulado con **hasta 5 fichas compactas** — carátula arriba, nombre y descripción corta —, en el orden de la sección.
2. **Cada ficha del bloque lleva sus enlaces externos** (FR-027: los mismos de la sección): «Ver en YouTube» y/o «Escuchar en Spotify» como pastillas compactas con el anuncio de pestaña nueva igual que en la sección.
3. Quiere ver todo → toca el botón **«Ver todo»** del bloque → `/medios` en **un solo paso** (US6 esc. 5).
4. Con más de 5 visibles: exactamente los **últimos 5** del orden de la sección, sin paginación, carrusel ni scroll lateral (US6 esc. 2). Con 1–4: exactos, sin huecos ni marcadores vacíos (US6 esc. 3).
5. Con **ningún episodio visible**: el bloque **no se muestra en absoluto** — sin título, sin mensaje, sin reservar espacio — (US6 esc. 4; patrón de F6/F7; **diferencia deliberada con «Lo que viene» de F4**, que sí muestra su vacío con mensaje).
6. Un episodio oculto o una entrada pendiente de «posible duplicado» jamás aparece en el bloque (US6 esc. 6, FR-009).
7. En inglés: interfaz por diccionario; contenido con respaldo (FR-025), sin marcas; datos no traducibles idénticos (US4 esc. 5).

---

## 3. Rutas propuestas (React Router)

```
Sitio público (sin sesión):
/                      Portada de F3 + bloque F4 + bloque F6 + bloque F7 + bloque F9 (MediosBlock)
/medios                «Medios» de F9 (MediosPage)
/health                Estado del sistema (F1), fuera de la navegación pública

Panel (patrón F2/F3/F4/F6/F7, sin rediseño):
/panel                 Inicio (acceso nuevo por permiso)
  /panel/medios        RequirePermission(MEDIOS)  Gestión de medios
  /panel/{demás}       (F2–F7, intactos)
```

- **`/medios` es una ruta real** (mismo patrón aprobado por el humano en F4/F6/F7 — ⟲ D-1). Se añade al final de la **única lista de enlaces** de F10 — después de «Ofrendas y diezmos» — en `useSiteLinks.ts`: `SiteLinkId` se extiende con el id nuevo (propuesta: `'media'`; el nombre exacto lo fija el arquitecto) y el vocabulario con la clave i18n `nav.sections.media`. El `<nav>` sigue siendo **una sola lista en DOM**; no se toca `nav.label`.
- **Presencia en la navegación (FR-008/FR-009)**: la entrada «Medios» existe **solo** si el documento público (`usePublicMedios`) trae **al menos un episodio visible**. La presencia sale del **documento de la sección**, compartida por `queryClient` con la página. El **bloque de la portada** (`usePortadaMedios`) es una consulta propia (tope 5 resuelto en servidor) y **no es fuente de navegación** (misma separación de responsabilidades de F7).
- **Mapa del sitio del pie**: «Medios» entra en el grupo **Explorar** (no se abre columna nueva), después de «Ofrendas y diezmos» — mismo criterio de F4/F6/F7.
- **Sección oculta abierta a mano**: `/medios` con 0 visibles **redirige a la portada** (Q10): **sin página vacía, sin aviso, sin parpadeo** (mecánica exacta en §4.1, patrón de `MinisteriosPage` — ya construido y verificado). **Distinto a propósito** de `/eventos` de F4 (que conserva `HiddenNotice`): no unificar.
- **Permission code**: el módulo se autoriza con el permiso existente **`medios`** (`RequirePermission`, constante a añadir en `lib/permissions.ts` — el nombre del code lo confirma el arquitecto frente a F2); no cambia de nombre (Q13). No hay permiso nuevo.
- **Portada**: el bloque se inserta en `HomePage` **inmediatamente después de `<OfrendasBlock />`** (Q12/FR-027) — verificado en el código actual — y antes de la siguiente sección de F3. Dirección de `Reveal`: la que conserve la alternancia contigua (⟲ D-14; decisión menor).

---

## 4. Pantallas

### 4.1 Sitio público — `/medios` (visitante; US1, US4, US5)

**Propósito**: en una lectura (SC-001: menos de 30 s hasta la grabación en YouTube) el visitante sabe qué prédicas y podcasts hay, elige una y la abre **donde está publicada**, sabiendo que sale del sitio.

**Estructura** (página de contenido bajo `PublicLayout` de F10):

1. **Enlace de salto** (_skip link_) — de F10, intacto; `<main>` de `/medios` es su destino aquí.
2. **H1 «Medios»** (`media.page.title`) + subtítulo fijo (cadena de interfaz, no campo del panel — criterio D-2 de F6/F7): «Las prédicas y los podcasts de nuestra iglesia; se abren en YouTube y Spotify, en una pestaña nueva.» / («Watch and listen to our church's sermons and podcast; they open in YouTube and Spotify in a new tab.») ⟲ D-2 — primera oportunidad del 90 % de SC-012.
3. **Lista de fichas** (una sola, FR-006): `<ul>` con grilla de 1 columna móvil → 2 columnas tableta/escritorio, con `RevealGroup` y stagger suave. Solo episodios **visibles**, del más reciente al más antiguo. Cero filtros, cero buscador (Q11: **lista simple**).
4. **«Cargar más»** — al final de la lista, **solo si quedan episodios** por mostrar: `PublicButton` secundaria a todo el ancho en móvil, ≥ 44 px. Una tanda = **12** fichas (default revisable #5). Cuando no queda más que mostrar, el botón **desaparece** (sin mensaje «no hay más»: la lista simplemente se acaba). ⟲ D-3
5. **Pie de F10 intacto**: mapa del sitio con «Medios» en **Explorar** (solo si hay algo visible), selector de idioma, tema claro/oscuro.

**Composición de la ficha (`MedioCard`, `Card solid` `as="li"`)** — de arriba abajo, **cada elemento solo existe si tiene dato**:

- **Carátula** a sangre en el marco superior (`CardImage`, 16:9), es la copia local del feed. Episodio **sin carátula**: `CaratulaNeutra` (§6.3) — **mismo marco 16:9** con la identidad de la iglesia y el lema; la ficha queda completa, sin hueco ni rotura (US1 esc. 8; ⟲ D-4). `alt` de la carátula: «Carátula del episodio {nombre}» / («Cover of the episode {name}») ⟲ D-31.
- **Nombre** (h3, Poppins 600 1.25rem, igual tratamiento que el h3 de F3/F6; **no Bebas grande**): el título del feed **tal cual** («Potencial sin Resultados La Lección de la Higuera» — con su puntuación; tildes, ñ, ¿ ¡ y emojis íntegros). **No es un enlace** (⟲ D-5): la ficha no es clicable en bloque.
- **Fecha de publicación** (texto pequeño, gris del tema): «17 de mayo de 2026» vía `formatLongDate`; interpretada en `America/Costa_Rica` en el servidor (FR-006); **el dato es único entre idiomas** (FR-026) y la palabra del formato sigue el locale de la interfaz.
- **Descripción corta**: Poppins 18 px, visible hasta ~3–4 líneas en la ficha (`line-clamp` como presentación); **el texto completo está en el DOM** para lector de pantalla — el recorte real del dato lo fija el límite de FR-020 en el servidor, no la vista (⟲ D-6). Con la versión en inglés del equipo cuando existe (§2.2).
- **Botones de enlace— los únicos interactivos de la ficha** (uno o dos según los formatos):
  - **«Ver en YouTube»** (`PublicButton` con `external`): icono ↗ visible + microtexto visible «pestaña nueva» dentro del propio botón (⟲ D-7).
  - **«Escuchar en Spotify»** (ídem, si hay audio).
  - En móvil van **apilados a todo el ancho** (cada botón ≥ 44 px); en escritorio, en fila con wrap.
  - El **anuncio completo vive en `aria-label`**: «Abre YouTube en una pestaña nueva para ver {nombre}; sales del sitio de la iglesia.» — y la variante Spotify para escuchar (§7.2). El texto visible da el servicio; el anuncio completo viaja con el enlace para las dos audiencias (SC-012, ⟲ D-7).
  - Apertura siempre `target="_blank" rel="noopener noreferrer"` — ya resuelto por `PublicButton` con `external`; nada que inventar.
  - Un episodio con **un solo formato** muestra **solo su botón**, sin líneas ni marcadores vacíos (US1 esc. 3).

**Propuesta wireframe desktop y móvil** (esquema de referencia, no maqueta):

```
DESKTOP
[ BARRA F10: 🌱 Simiente Santa · Quiénes somos · Horario · WhatsApp · Contacto · Redes
                          · Eventos y actividades · Ministerios · Ofrendas y diezmos · Medios | Español/English | 🌙 ]
══════════════════════════════════════════════════════════════════
  MEDIOS  (H1, Bebas)
  Las prédicas y los podcasts de nuestra iglesia; se abren en
  YouTube y Spotify, en una pestaña nueva.
        ┐ --space-block
  ┌─ Card ────────────────────┐  ┌─ Card ────────────────────┐
  │  [CARÁTULA local, 16:9]   │  │  [CARÁTULA local, 16:9]   │
  │  Potencial sin Resultados │  │  A Mayor Fe, Mayor Presión│
  │  17 de mayo de 2026       │  │  17 de mayo de 2026       │
  │  La lección de la higuera…│  │  …descripción corta…      │
  │  [Ver en YouTube ↗]       │  │  [Ver en YouTube ↗]       │
  │  [Escuchar en Spotify ↗]  │  └───────────────────────────┘
  └───────────────────────────┘   (una sola ficha: video + audio)
        ┐
  [ Cargar más ]  (solo si quedan)
        ┐ --space-section
  PIE DE F10 (mapa del sitio: Explorar incluye «Medios»)
══════════════════════════════════════════════════════════════════

MOBILE (≤ 640 px), columna única:
[ SS Simiente Santa · Español/English · 🌙 · ☰ ]
→ ☰ abre panel (la MISMA lista, con «Medios»)
grilla 1 col: carátula a todo el ancho, nombre, fecha, descripción,
botones de enlace a TODO el ancho y ≥ 44 px, apilados.
```

**Estados de la página pública**:

| Estado | Qué ve el visitante |
|---|---|
| Carga (primera) | Esqueleto con medidas reales: H1 + grilla fantasma de `Card` **con área de imagen 16:9** y líneas de texto (como los fantasmas de eventos/ministerios); barra, idioma y tema ya visibles; cero brincos de layout. |
| Éxito | Fichas del más reciente al más antiguo; enlaces anunciados; «Cargar más» agregando una tanda nueva por toque. |
| Éxito con ≤ 12 visibles | Lista completa sin «Cargar más» (⟲ D-3). |
| Episodio con un solo formato | Solo su botón; ficha coherente, sin líneas vacías ni «sin datos» (US1 esc. 3). |
| Episodio sin carátula | Marcador neutro con la identidad de la iglesia (FR-002, US1 esc. 8). |
| **Sección sin nada visible** (Q10/FR-008) | La entrada **no existe en barra, menú ni mapa del pie** (nav sin enlaces rotos). Si alguien abre `/medios` a mano: **redirección silenciosa a la portada** con `<Navigate to="/" replace />` (patrón construido de `MinisteriosPage`): **nunca se pinta la sección vacía**, jamás aviso; con la consulta en caché redirige en el primer render; sin caché, **solo el esqueleto estándar** mientras se decide; con `replace` el historial queda limpio. `MediosBlock` también devuelve `null`. Cero pantallas intermedias (SC-016). |
| Error de carga sin nada en caché | H1 + mensaje bilingüe claro: «No pudimos cargar los medios. Revisa tu conexión y vuelve a intentarlo.» + botón «Volver a intentar». **El error nunca actúa como «sin contenido»**: no redirige (el público no debe perder la sección por una falla pasajera; ⟲ D-12). |
| Error con caché previa | Último contenido válido + aviso discreto: «No pudimos actualizar; ves la última versión disponible.» (F3). |
| Episodio oculto | No aparece en público; la búsqueda nunca lo vuelve a publicar (FR-015). |
| Texto solo en español navegando en inglés | Contenido en español, idéntico diseño, sin marca de «sin traducir» (FR-025); el nombre no cambia. |
| Duplicado pendiente de resolución | La entrada adicional está **oculta** y no aparece por ninguna vía (FR-009/FR-014; US1 esc. 6). |
| Orden estable | `fecha de publicación desc.` con desempate por orden de importación (default revisable #11); igual en cualquier ancho o idioma. |
| Duplicado de concepto | Solo prédicas y podcasts: nada de eventos (F4), galería (F8), grupos (F5) ni ministerios (F6) (FR-001/SC-013). |
| `prefers-reduced-motion` | Sin revelado ni stagger; contenido íntegro al instante (F10). |

**Responsivo**:

- **Móvil (< 640 px)**: 1 columna; carátula a todo el ancho; nombres y descripciones con `overflow-wrap` sin desbordes; **botones de enlace apilados a todo el ancho** (≥ 44 px); 0 scroll horizontal a 320 px (probado con títulos largos).
- **Tableta (≥ 640 px)**: 2 columnas; H1 con `clamp`.
- **Escritorio (≥ 1024 px)**: `max-w-6xl`, **sigue en 2 columnas** (la ficha es alta por la carátula; 3+ columnas mínificaría nombre y descripción; ⟲ D-13).
- **Zoom 200 %**: diseño fluido en `rem` con `clamp`; la ficha reacomoda sin perder información (US5 esc. 4).

### 4.2 Panel — Inicio (cambio menor en F2)

`AccesosRapidos` añade **una** tarjeta con el patrón ícono + texto existente:

- **«Medios»** (con permiso `medios` → `/panel/medios`).

Una cuenta sin el permiso no ve la tarjeta (patrón F4/F6/F7). La auditoría registra el área «medios» sin UI nueva (FR-022).

### 4.3 Panel — Gestión de medios — `/panel/medios` (US2, US3)

**Propósito**: mantener el catálogo de episodios importados de los dos feeds con las únicas acciones del módulo (buscar, ocultar/mostrar, unir, descripción), y ver de un vistazo **qué es visible hoy al público** y **qué está pendiente de revisión**.

- Título «Medios» + banda de ayuda (⟲ D-8): «Aquí miras y mantienes las grabaciones que el sitio trae desde el canal de YouTube y el feed del podcast. Pulsa “Buscar episodios nuevos” cuando se publique algo nuevo y revisa con calma lo que quede marcado; solo lo visible lo verá el público» (textos completos en §7.1).
- **Estructura de la página (arriba → abajo)**:
  1. **Zona «Buscar episodios nuevos»** (fija arriba del listado): título pequeño de la zona «Desde YouTube y el podcast», el **botón primario** «Buscar episodios nuevos» y una **línea de apoyo** debajo — «Es seguro pulsarlo todas las veces que quieras: solo agrega lo que falte, nunca borra y nunca cambia lo que ya decidiste.» (⟲ D-9: la idempotencia **se dice** en la UI; §4.4).
  2. **Resumen de la última búsqueda** (transitorio, §4.4): `Notice` con conteos y estado de los feeds; visible hasta la próxima búsqueda o recarga.
  3. **Filtro rápido** (⟲ D-10): fila de pastillas **«Todo» / «Solo lo pendiente de revisión»** (con conteo cuando hay pendientes: «Posible duplicado (2)», «No disponible (1)», «Ya no aparece en el feed (3)»). Es un **filtro de vista**, no de datos; ayuda a resolver marcas rápido. La spec no lo exige: es propuesta revisable.
  4. **Listado del catálogo**: **una sola lista** (no hay grupos como IBAN/SINPE en F7), `Table` (desktop) / tarjetas apiladas (móvil), patrón F2/F4/F6. Por fila:
     - **Carátula**: miniatura 48 px; marcador neutro si no hay.
     - **Nombre** (el del feed, **no editable**, texto principal de la fila) + **fecha** debajo en gris pequeño.
     - **Formatos**: chips con palabra **«Video»** / **«Audio»** según los enlaces resultantes (⟲ D-15).
     - **Estado**: `StatusPill` **«Visible»** / **«Oculto»**.
     - **Marca** (solo si la hay): píldora con palabra + **explicador de una línea** debajo (§4.7).
     - **Acciones**: **«Editar descripción»** · **«Ocultar»**/**«Mostrar»** · **«Revisar duplicado»** (solo en filas con marca de duplicado) · resolución de marca según su tipo (§4.7). Ninguna acción es solo icono.
  5. **Todo lo importado está en la lista** (visibles, ocultos y marcados): el equipo **siempre ve lo que el público ve hoy**. Sin estados intermedios. El **origen** (de qué feed vino y cuándo se importó) se muestra **dentro de los diálogos** (§4.5/§4.6), no como columna de la lista (⟲ D-32): la lista queda concentrada en estado/marca/acciones.

**Nota — por qué la lista no se agrupa por estado**: agrupar visible/oculto haría que un episodio «salte de bloque» al ocultarlo (movimiento innecesario) y separaría lo marcado del resto. Con el filtro «Solo lo pendiente de revisión» (⟲ D-10), la revisión queda a un clic. Si el humano prefiere una sección «Pendientes» separada, es revisión de diseño, no de negocio.

### 4.4 Flujo «Buscar episodios nuevos» (antes / durante / después)

**Antes (estado estable)**:

- El botón primario queda siempre disponible en su zona exclusiva, **sin confirmación previa** (⟲ D-16): la búsqueda es **idempotente y no destructiva** (FR-012: solo agrega; nunca borra ni pisa lo decidido) — pedir confirmación cada vez no protege de nada y suma una parada a una acción **semanal**.
- Con **nada importado todavía** (primera vez): el listado muestra su estado vacío y remite al mismo botón (§4.8).
- Sin sesión válida o sin permiso: no se llega a operar (§2.6).

**Durante (mientras corre la búsqueda, dentro de la misma petición)**:

- El botón pasa a **`Button loading`** («Buscando…») **deshabilitado**: no se puede pulsar de nuevo mientras corre (una sola búsqueda activa a la vez; patrón de F3; ⟲ D-16).
- **`role="status"`** junto al botón: «Buscando episodios nuevos en YouTube y en el podcast…» (lo anuncia el lector de pantalla automáticamente; visible igualmente para quien no lo usa). ⟲ D-17
- El listado queda **visible y sin bloquear** hasta terminar; los controles de fila **se deshabilitan** durante la búsqueda (ningún ocultar/uni/editar concurrente que deje datos a medias; FR-023). ⟲ D-16
- **Sin barra de progreso ni porcentajes** (el proceso es una lectura + comparación de segundos) ⟲ D-18; si el servidor informa por feed, el texto del estado podría avanzar («Leyendo el podcast…»), pero la propuesta mínima es un solo mensaje.
- El corte de tiempo de cada feed lo fija el servidor (decisión del arquitecto): la UI solo queda a la espera del resumen.

**Después (el resumen de la búsqueda, FR-018)**:

El **resumen** aparece entre la zona de búsqueda y el listado refrescado. Un mensaje por caso (texto exacto en §7.1):

| Resultado | Tono del `Notice` | Mensaje |
|---|---|---|
| Nuevos importados, todo unido, sin marcas | Éxito (verde) | «Se agregaron {N} episodios nuevos; {A} se unieron con su video y su audio. Ya están visibles en “Medios”.» |
| Nuevos importados **y** algunos marcados | Éxito (verde) | «Se agregaron {N} episodios nuevos; {A} se unieron con su video y su audio; {D} quedaron marcados como posible duplicado. Revísalos abajo.» ⟲ D-19: un segmento solo se muestra si su número es ≥ 1 (nunca «0 se unieron»). |
| **Sin novedades** (FR-012) | Neutro | «No hay episodios nuevos: nada cambió. Cuando la iglesia publique algo más, vuelve a pulsar el botón.» (⟲ D-20) |
| **Un feed falló** (FR-016) | Aviso (ámbar) | «El feed de {YouTube / el podcast} no respondió. Se agregó lo que trajo {el otro}; nada se ocultó ni se marcó por esa causa. Vuelve a pulsar la búsqueda en un rato para recuperar lo que falte.» (⟲ D-21: se nombra por servicio, nunca «Atom/RSS/Anchor») |
| **Ambos feeds fallaron** | Error (rojo) | «Ninguno de los dos feeds respondió. No se agregó nada y no se cambió nada: todo lo que tenías está a salvo. Prueba más tarde.» |
| **Cambios de disponibilidad** (FR-017) | Aviso/info | «La búsqueda revisó {M} episodios para ver si sus enlaces siguen funcionando: mira las marcas.» (y las filas afectadas quedan con su marca y explicador, §4.7) |

- El resumen **permanece visible** hasta la próxima búsqueda o recarga — no es un toast de 3 s (⟲ D-22: el equipo necesita poder comunicar el resultado a otros).
- La **lista se refresca** sola (`invalidateQueries` sobre el catálogo; el documento público y el bloque también se invalidan). Lo nuevo entra en su lugar por fecha; nunca se reordena a mano.
- Con todas las marcas resueltas y sin novedades, el panel queda listo para la próxima semana sin verificación manual del público: la visibilidad ya se aplicó al guardar (FR-015).
- La **auditoría** registró la operación (quién, qué, resultado — FR-022), sin UI nueva.

**Comprobación de importación con los datos entregados** (dato de partida, no regla): la primera búsqueda con los feeds del 2026-10-17 produce **10 entradas** — 8 sermones unidos (con los dos enlaces) + 2 con un solo enlace (de YouTube «A Mayor Fe, Mayor Presión»; del podcast «Fe como fruto del Espíritu»). Es el resultado de referencia que QA puede replicar; no es un requisito de negocio.

### 4.5 Panel — Dialog «Editar descripción» (`DescripcionForm`, en `Dialog`; patrón F3/F6/F7)

- **Contexto arriba del formulario**: miniatura de la ficha (carátula pequeña + nombre + fecha, en solo lectura) para que el equipo sepa **qué episodio** está ajustando sin leer identificadores (⟲ D-23; los nombres y fechas contiguas se parecen — el contexto evita editar la ficha equivocada).
- **Tabs de idioma** (`Tabs` existente): **Español** (obligatorio **si el campo se rellena** — FR-023) / **English (opcional)** con la ayuda de F3: «Si lo dejas vacío, el público verá este contenido en su versión en español.»
- **Campo de texto** (`TextArea` grande, contador vivo **{n}/{L}**; límite del default revisable #1, propuesta **L = 300**):
  - Precargado con la **descripción corta vigente** (la generada por el recorte automático llega como punto de partida para retocar; **una vez guardada por el equipo manda** y no se regenera ni se pierde — FR-015/FR-020/SC-007).
  - Al guardarse se cambia **solo** el texto: la descripción completa del feed no se publica nunca (FR-020); ayuda bajo el campo: «Se muestra tal como la escribas, en la ficha del episodio.»
  - **Límite excedido**: el contador pasa a error y «Guardar» queda deshabilitado (aviso exacto en §7.1). F9 nunca guarda a medias (FR-023).
- **La visibilidad no se cambia aquí** (⟲ D-24): se cambia con «Ocultar»/«Mostrar» en la fila — «editar texto» y «decir si se ve» son decisiones separadas (patrón de F7 D-11).
- **Guardar**: `Button` loading («Guardando…»); el modal no se cierra hasta terminar; `Esc` no descarta lo escrito sin aviso (criterio D-18 de F4).
- **Al guardar**: cierre + `Notice` verde «Cambios guardados: ya visibles en la sección.» (FR-015: se publica al guardar). Si el episodio estaba oculto: «Cambios guardados.» — el público no lo ve hasta «Mostrar»; la auditoría registra el cambio igual (FR-022).

### 4.6 Panel — Dialog «Revisar duplicado» (`UnirDuplicadosDialog` + `ConfirmDialog`)

**Flujo — decisión antes que sorpresa**:

1. El equipo activa **«Revisar duplicado»** en una fila marcada → el `Dialog` muestra **las dos candidatas lado a lado** (columna A y B): cada una con **carátula, nombre completo, fecha, enlace y origen** («De YouTube» / «Del podcast»), con su `StatusPill` de visibilidad actual («Visible» la primera importada, «Oculto» la adicional — la regla de FR-014 **expresada en pantalla**: el equipo ve por qué una está escondida).
2. Debajo, la zona **«Así quedaría la ficha si las unes»** — **vista previa del resultado**: carátula, **el nombre de la ficha de YouTube** (default revisable #7, declarado con el microcopy «Al unir, el nombre publicado es el de YouTube.» ⟲ D-25), la fecha y **los dos enlaces juntos**. Antes de confirmar, el equipo ya vio **qué quedará** — la confirmación es informada (FR-014/US3 esc. 1).
3. **Tres acciones al pie**:
   - **«Unir»** → abre un `ConfirmDialog` **danger**: «Se unirán las dos entradas en una sola ficha. **Esto no se puede deshacer.**» Botones **«Unir definitivamente»** / «Cancelar» (⟲ D-26: doble paso porque la unión es **definitiva** — default revisable #8; es la única operación irreversible del módulo, merece su barrera). Al confirmar: `Notice` verde «Unidas en una sola ficha, visible en “Medios”.»; la fila única queda con los dos enlaces y **sin marca**; y la operación queda en auditoría (FR-022).
   - **«No son el mismo sermón»** → resuelve la marca en su lugar: ambas quedan **visibles y sin marca**; `Notice` verde «Listo: quedaron las dos como episodios distintos.» (US3 esc. 2). Sin confirmación extra: no destruye nada y es reversible ocultando después (⟲ D-27).
   - **«Cancelar»** (o `Esc` / clic fuera): cierra sin cambios, con devolución del foco a la fila origen (trampa de foco estándar del `Dialog`).
4. **Caso sin nada que revisar**: si no hay ninguna marca de duplicado, **ninguna fila ofrece el botón y no existe diálogo de unión** (patrón «sin dato, sin huella»); el último resumen de búsqueda ya lo informó (⟲ D-28: no hay «botón de unir» general para unir fichas cualquiera a mano — **unir solo opera sobre pares que el sistema marcó**; fuera de eso, la acción no existe en la UI).

### 4.7 Panel — Marcas de revisión y su resolución

Cada marca entra con **palabra + explicación** en la fila (§4.3) y se resuelve desde ahí (FR-017 + US3 esc. 6):

| Marca (píldora) | Qué significa (línea que acompaña a la marca) | Estado del episodio | Acciones | Al resolver |
|---|---|---|---|---|
| **Posible duplicado** | «El sistema no pudo estar seguro de si estas dos entradas son el mismo sermón.» | Primera importada **Visible**; adicional **Oculta** (FR-014) | «Revisar duplicado» (§4.6) | Unión → una ficha visible con los dos enlaces; «no son el mismo» → ambas visibles sin marca. |
| **No disponible** | «Se comprobó: el enlace de esta grabación ya no funciona (borrada o privada). El público no la ve.» | **Oculta** (Q16/FR-017) | **«Mostrar de nuevo»** (con `ConfirmDialog` suave: «Si el enlace de {nombre} sigue sin funcionar, volverá a quedar oculto la próxima vez que se revise.» ⟲ D-28) · **«Mantener oculto»** (cierra la marca y queda oculto, sin pendiente). | Mostrar → visible y **sin marca**; «Mantener oculto» → sigue oculto con la marca cerrada. La misma causa no reabre la marca (FR-017). |
| **Ya no aparece en el feed** | «Salió de la ventana del feed de {YouTube / el podcast} —eso es normal, la ventana es corta—. Su enlace {“sigue funcionando” / “no se pudo comprobar (se revisó sin respuesta)”}: sigue visible para el público.» La incertidumbre se **nombra**, no se esconde. | **Visible** (regla a prueba de fallos: nunca oculto por falta de certeza) | **«Mantener visible»** (cierra la marca; el episodio queda como está) · **«Ocultar»** (decisión del equipo, reversible) ⟲ D-29. | Mantener visible → sin marca; Ocultar → oculto. La misma causa no reabre. |

> **No solo color** (FR-030): cada marca lleva **palabra** en la píldora y, si el humano lo aprueba (⟲ D-30), un **icono** distinto: dos fichas apiladas (duplicado), enlace roto (no disponible), reloj (ya no aparece en el feed). La píldora con palabra ya cumple el «no solo color» por sí sola; el icono es refuerzo propuesto.

> **El episodio con «Ya no aparece en el feed» sigue visible para el público**: el visitante lo ve normal — solo el panel lo distingue como pendiente. En ninguna superficie pública aparecen marcas, píldoras ni frases sobre feeds: el lenguaje del visitante es solo «prédicas y podcasts».

### 4.8 Estados vacíos del panel

- **Catálogo vacío (nada importado nunca)**: el listado muestra `EmptyState` — «Todavía no hay episodios. Pulsa “Buscar episodios nuevos” para que el sitio lea por primera vez el canal de YouTube y el feed del podcast.» — **sin duplicar el botón**: el estado vacío **apunta a la zona de importación**, donde ya vive la única acción de la página (⟲ D-33; FR-011).
- **Filtro «Solo lo pendiente de revisión» sin pendientes**: «No hay nada pendiente de revisión: todo está al día.» (es un logro del equipo, no un hueco).
- **Tras una búsqueda que no trajo nada** (ambos feeds caídos o sin novedades): la lista conserva lo que ya había **sin duplicar ni alterar nada** (FR-016/FR-012) y el resumen lo explica (§4.4).

### 4.9 Portada — bloque «Medios» (`/`; US6, Q12, FR-027, SC-015)

> Único contenido de F9 en la portada de F3; **tercer bloque añadido por el patrón de bloques** (después de ofrendas — decisiones 18/20/22 de F4/F6/F7). Usa la **misma gramática** aprobada (`SectionHeading` + fila + botón). Datos por **endpoint público propio** — propuesta `GET /api/v1/portada/medios?lang=es|en` vía `usePortadaMedios`: tope **5** resuelto en servidor, orden de la sección (fecha desc.), **solo lo visible** (FR-009), DTO mínimo, `Cache-Control: no-store` — una petición para el bloque; el contrato exacto lo fija el arquitecto. La **presencia en la navegación no sale de este endpoint** (misma separación P7-8 de F7; sale del documento de la sección `usePublicMedios`).

**Propósito**: quien entra descubre solo con la portada que hay prédicas y podcasts — y llega a una grabación en **un toque** (enlace) o a la sección en **un toque más** («Ver todo»).

**Ubicación (Q12, ya decidida)**: inmediatamente después del bloque «Ofrendas y diezmos» de F7 — insertado en `HomePage` tras `<OfrendasBlock />` (verificado). Sin alterar ninguna otra sección de F3.

| Estado | Qué ve el visitante (portada) |
|---|---|
| Con 5 visibles | `SectionHeading` «Medios» + fila de **5 fichas compactas** (carátula, nombre, descripción corta y enlaces externos por ficha, FR-027; ⟲ D-5): 1 col móvil, 2 tableta, 5 escritorio (⟲ D-15 — la ficha con carátula es alta; a 5 columnas la descripción queda en 2–3 líneas). |
| Con menos de 5 (1–4) | Exactas las que hay, sin huecos ni marcadores vacíos (US6 esc. 3). |
| **Con 0 visibles** | El bloque **no se dibuja**: sin título, sin mensaje, sin aviso, sin reservar espacio (FR-027; US6 esc. 4 — como `MinisteriosBlock`/`OfrendasBlock`, **a diferencia** de «Lo que viene» de F4). |
| Carga (primera) | Esqueleto propio: encabezado + fila de fichas fantasma **con área de carátula 16:9** y líneas de texto. |
| Error sin caché | El bloque **no se dibuja** (no deja un título huérfano sin contenido) — patrón de F6/F7. |
| Error con ítems en caché | Último contenido válido + aviso discreto: «No pudimos actualizar; ves la última versión disponible.» (F3). |
| Idiomas (US6 esc. 7) | Interfaz por diccionario; contenido con respaldo al español (FR-025) sin marcas; carátula, nombre, fecha, enlaces y formatos idénticos (FR-026). |
| `prefers-reduced-motion` | Sin `Reveal` ni transiciones: ítems visibles de inmediato (FR-028). |

**Composición de la ficha compacta (`MedioItem`)** — variante más pequeña de `MedioCard`, con los enlaces conservados:

- **Carátula** a sangre del marco superior (16:9).
- **Nombre del episodio** (h3, Poppins 600, máx. 2 líneas; el nombre completo vive en el DOM, el corte visible es presentación). Tal como lo trae el feed (no se traduce, FR-026).
- **Descripción corta** de máximo 2 líneas visibles (el texto completo en el DOM, ⟲ D-6).
- **Enlaces externos por ficha**: pastillas compactas «Ver en YouTube ↗» / «Escuchar en Spotify ↗» con el mismo anuncio completo en `aria-label` (solo las que existen).
- **Fecha omitida en el bloque** (⟲ D-32): el orden ya comunica recencia y el espacio de la ficha es corto; la fecha vive en la sección.
- **Botón del bloque** — debajo de la fila, alineado al final (a todo el ancho en móvil): `PublicButton` secundaria **«Ver todo»** / («See all») → `/medios` en un solo paso (US6 esc. 5). Con 0 ítems no existe (el bloque entero tampoco).

**Responsivo** — misma gramática D-28: móvil columna única; tableta 2 col; escritorio 5 col; **nunca carrusel ni scroll lateral** (0 scroll a 320 px).

**Accesibilidad** — nivel exacto del bloque de F7 (§4.9 de su ux.md) con los enlaces por ficha encima: `section` con `aria-labelledby` al `h2` «Medios» (al nivel de los títulos de la portada); fichas `li` con `h3` (`<ul>` conserva el rol de lista); orden de lectura = orden del DOM = orden por fecha; carátulas con `alt` completo (⟲ D-31); **cada enlace externo con su anuncio completo** (FR-005 igual a la sección); objetivos ≥ 44 px; fecha como texto (sin `<time>`), `prefers-reduced-motion` sin revelado.

---

## 5. Estados por vista con datos (panel)

### a. Zona de búsqueda (`/panel/medios`, arriba)

| Estado | Qué ve |
|---|---|
| Carga de la página | La zona de búsqueda **ya está disponible** (no depende del listado): título, botón y línea de apoyo; el listado carga con su propio `EmptyState` loading. |
| Reposo | Título de la zona + botón primario + línea de apoyo + (si la hay) el último resumen. |
| Buscando | Botón en loading «Buscando…» deshabilitado + `role="status"` «Buscando episodios nuevos en YouTube y en el podcast…»; acciones de fila deshabilitadas (⟲ D-16). |
| Éxito con importación | `Notice` verde con los conteos (§4.4); listado refrescado. |
| Sin novedades | `Notice` neutro «No hay episodios nuevos: nada cambió…» |
| Un feed falló | `Notice` ámbar: qué feed, que nada se ocultó/marcó, invitación a reintentar (FR-016). |
| Ambos feeds fallaron | `Notice` rojo: sin cambios en el catálogo. |
| Error del sistema al importar | `Notice` de error genérico; la lista quedó como estaba (nada a medias) — «tus episodios están a salvo» (§7.1). |

### b. Listado del catálogo

| Estado | Qué ve |
|---|---|
| Cargando | `EmptyState` kind=loading («Cargando los episodios…»). |
| Vacío (nada importado nunca) | `EmptyState` con el texto de §4.8/§7.1, remitiendo al botón de la zona de importación (⟲ D-33). |
| Éxito completo | Lista con carátula, nombre+fecha, chips de formatos, píldora de estado, marca con explicador y acciones por fila, en orden por fecha. |
| Filtro «Pendiente» sin pendientes | «No hay nada pendiente de revisión: todo está al día.» |
| Error de carga sin caché | «No se pudo cargar los episodios.» + `Button` «Reintentar» (`Notice` error, `role="alert"`). |
| Error con caché previa | Último listado válido + `Notice` discreto (patrón F3). |
| Sin permiso | No llega (ruta protegida); rechazo en caliente: mensaje de F2, sin cambiar nada. |
| Éxito de operación de fila | `Notice` temporal verde del catálogo de §7.1 (oculto/mostrar/descripción/unión/resolución de marca). |

### c. Los formularios (`DescripcionForm`, `UnirDuplicadosDialog` — ambos en `Dialog`)

- Precargados desde la fila; confirmaciones con los textos de §7.1.
- Guardando/unir: `Button` loading; el modal no se cierra hasta terminar; `Esc` respeta el criterio D-18 de F4 (sin descartar cambios sin aviso).
- Éxito: cierre + `Notice` en la página + fila refrescada.
- Errores de campo (español/inglés, límite) junto al campo y resumen arriba, foco al primer error (FR-023).
- Trampa de foco estándar del `Dialog` y devolución del foco a la fila que lo abrió.

### d. Ocultar / mostrar / unir / resolver marca

- **Ocultar**: directo, **sin confirmación** (⟲ D-8 — es totalmente reversible con «Mostrar»); la fila pasa a «Oculto» y el aviso dice el efecto real: «Quedó oculto: el público no lo ve.» Desaparece de `/medios` **y del bloque** (FR-009).
- **Mostrar**: directo; el episodio **reaparece en su lugar por fecha** (no al final); aviso: «Ahora está visible en “Medios”.»
- **Unir**: doble paso (§4.6 ⟲ D-26) con `ConfirmDialog`; definitiva; una ficha única visible sin marca + auditoría.
- **«Mostrar de nuevo»** (desde la marca «No disponible»): **con `ConfirmDialog` suave** (⟲ D-28) — es la única acción que puede exponer de nuevo al público un enlace posiblemente defectuoso si el equipo se equivoca; merece la barrera. «Mantener oculto» y «Mantener visible» son directas (no cambian nada ante el público).
- **Auto-publicación al guardar** (episodio visible): «Cambios guardados: ya visibles en la sección.» (FR-015).
- **Toda acción de fila queda deshabilitada durante una búsqueda en curso** (FR-023; ⟲ D-16).

---

## 6. Componentes React

> Nombres en inglés (coherente con el repo): la **feature pública** en `frontend/src/features/publico/medios/`; la del panel en `frontend/src/features/medios/` (módulo/permiso, Q13). El listado final de tareas lo fija el arquitecto; aquí la propuesta con responsabilidad. **Reutilizar antes de crear**: casi todo existe (verificado en el código actual).

### 6.1 Reutilizados del inventario único (sin cambios)

| Componente | Uso en F9 |
|---|---|
| `Button` / `PublicButton` | Acciones del panel; «Cargar más», «Ver todo»; y con **`external`** los enlaces externos de las fichas (`target`/`rel` resueltos por el componente). |
| `Field`, `TextArea` (de `features/informacion/components/`), `Notice` | Formulario de descripción; avisos de panel y públicos. |
| `Dialog` / `ConfirmDialog` | Descripción; unir duplicados y confirmaciones de resolución de marca. |
| `EmptyState` | Carga/vacío/error de listado y zonas del panel. |
| `StatusPill` | Estado «Visible/Oculto» y **marcas** (con `label` sobrescrito). |
| `Table` | Listado del catálogo (desktop); tarjetas en móvil (patrón F4/F6/F7). |
| `Tabs` | Pestañas de idioma es/en del dialog de descripción (patrón F3/F4/F6/F7). |
| `CardImage` | Carátula a sangre en fichas (cuadrado del tipo de imagen del DTO lo define el arquitecto; patrón de imagen local idéntico). |
| `Card` / `SectionHeading` | Superficie por ficha y rótulos (sección y bloque). |
| `Reveal` / `RevealGroup` | Entrada de la lista pública y del bloque de portada (stagger suave). |
| `SectionSkeleton` | Carga de la página pública y del bloque (fantasmas con área de imagen). |
| `PublicLayout`, `SiteHeader`, `PublicNav`, `MobileMenu`, `useSiteLinks` | Navegación; lista única extendida con «Medios» (tras «Ofrendas y diezmos»). |
| `LanguageSwitcher`, `LanguageProvider`/`useLanguage`, i18n (`es.ts`/`en.ts`) | Selector y cadenas; respaldo de contenido según FR-025. |
| `ThemeToggle`/`useTheme` | Modo claro/oscuro (F10, intacto). |
| `FooterSiteMap`, `BrandLogo`, `GradientBlobs`, `SocialIcon` | Pie, chrome y fondo intactos. |
| `AccesosRapidos` | Acceso nuevo del panel por permiso (F2). |
| `formatoFechas.ts` (de F4) | Presentación de la fecha de publicación es/en. |
| `RequirePermission` / `guards` | Protección de `/panel/medios` con el código «medios» (patrón construido). |

### 6.2 Nuevos del sitio público (`frontend/src/features/publico/medios/`)

| Componente / hook | Responsabilidad |
|---|---|
| `MediosPage` | H1 + subtítulo + grilla de fichas por fecha desc + «Cargar más», a partir de `usePublicMedios`; **decide la redirección a la portada con éxito y 0 visibles** (Q10; mecánica sin parpadeo, patrón `MinisteriosPage`); **ante error no redirige** (⟲ D-12). |
| `MediosGrid` | Grilla de `MedioCard` (1/2 columnas), `<ul>` semántica, `RevealGroup`. |
| `MedioCard` | Ficha pública: carátula (con marcador neutro), nombre (h3, no enlace), fecha, descripción corta, fila de enlaces externos. Los únicos interactivos son los enlaces (⟲ D-5). |
| `CaratulaNeutra` | Marcador 16:9 con la identidad de la iglesia para episodio sin carátula (recursos del repo, nunca peticiones a terceros; ⟲ D-4). |
| `MedioLinks` | Fila de botones «Ver en YouTube» / «Escuchar en Spotify» — `PublicButton external` con icono ↗ + «pestaña nueva» visible y `aria-label` completo (FR-005). |
| `CargarMasButton` | «Cargar más»: agrega la siguiente tanda; desaparece cuando no quedan (FR-006/Q11, ⟲ D-3). |
| `MediosBlock` (portada) | Bloque de portada: `SectionHeading` «Medios» + fila de hasta 5 `MedioItem` + «Ver todo», a partir de `usePortadaMedios` (tope 5 en servidor); **devuelve `null` sin ítems** y en error sin caché (como `MinisteriosBlock`/`OfrendasBlock`); esqueleto con áreas de carátula; aviso discreto con caché y error. **No es fuente de navegación**. |
| `MedioItem` (portada) | Ficha compacta: carátula, nombre (h3, máx. 2 líneas), descripción corta (máx. 2 líneas, DOM completo), **enlaces externos por ficha** (pastillas con anuncio completo); fecha omitida (⟲ D-32). |
| `usePublicMedios` | TanStack Query del documento público (contrato del arquitecto; propuesta `GET /api/v1/medios?lang=es|en` — fichas visibles ya ordenadas e idiomas resueltos + marca de «quedan más»): alimenta la página **y la presencia en `useSiteLinks`** (compartida por `queryClient`). |
| `usePortadaMedios` | TanStack Query para el bloque: propuesta `GET /api/v1/portada/medios?lang=es|en` (tope 5 en servidor, `no-store`); alimenta solo `MediosBlock`. Son **dos consultas** deliberadas (nunca consume la de `/medios`). |

### 6.3 Nuevos del panel (`frontend/src/features/medios/`)

| Componente / hook | Responsabilidad |
|---|---|
| `pages/MediosPage.tsx` | Página del módulo: título + banda de ayuda + zona de búsqueda + resumen + filtro de marcas (⟲ D-10) + listado + estados. |
| `components/BuscarEpisodios.tsx` | Zona de importación: botón primario + línea de apoyo + texto de contexto con `role="status"` durante la búsqueda. |
| `components/ResumenImportacion.tsx` | `Notice` del resumen con los conteos (solo menciona segmentos ≥ 1, ⟲ D-19) por los seis casos de §4.4. |
| `components/FiltroMarcas.tsx` | Pastillas «Todo / Solo lo pendiente de revisión» (⟲ D-10; filtro de vista, con conteos). |
| `components/EpisodiosList.tsx` | `Table`/tarjetas del catálogo: carátula miniatura, nombre+fecha, chips de formatos, `StatusPill`, marcades marca con explicador y acciones por fila. |
| `components/DescripcionForm.tsx` (`Dialog`) | Edición de la descripción corta (§4.5): miniatura de contexto (⟲ D-23), pestañas de idioma, contador/límite, guardar con auto-publicación. |
| `components/UnirDuplicadosDialog.tsx` | Flujo de unión (§4.6): candidatas lado a lado + **vista previa del resultado** + «Unir» (→ `ConfirmDialog` definitive) / «No son el mismo sermón» / Cancelar. |
| `components/MarcaDeRevision.tsx` | Píldora de marca con palabra (+ icono si se aprueba, ⟲ D-30) y línea de explicador por tipo (§4.7). |
| `validation.ts` | Zod: descripción es/en dentro del límite; español presente si se rellenó inglés (FR-023/FR-025). **No valida datos del feed** (nunca pasan por un formulario). |
| `messages.ts` | Textos del módulo en español (FR-024) — catálogo del §7.1. |
| Hooks TanStack Query/mutations | `useMedios` (catálogo), `useBuscarEpisodios` (mutation: dispara, deshabilita la UI, devuelve el resumen y refresca), `useOcultarMostrarEpisodio`, `useUnirDuplicados`, `useActualizarDescripcion`, `useResolverMarca` (mutations que invalidan el catálogo). `usePublicMedios`/`usePortadaMedios` viven en la capa pública (§6.2). Ningún `fetch` en componentes (regla F2). |

**Pruebas esperadas** (Vitest + Testing Library + MSW): página pública con 0 visibles → **redirección silenciosa** (con caché en el primer render; sin caché solo el esqueleto); 1/12/13+ visibles → «Cargar más» acumulando tandas y desapareciendo al final; un formato → un solo botón; sin carátula → marcador neutro; orden por fecha con desempate por importación; **0 recursos de terceros** (0 `iframe`, 0 `src` de youtube/spotify/anchor — SC-005); **cada enlace externo con `target=_blank` + `rel="noopener noreferrer"` + `aria-label` que anuncia servicio, pestaña nueva y salida del sitio** (SC-006/SC-008); bloque de portada con 0 (ausente, sin mensaje), 1–5 y >5 (últimos 5 exactos); fallback de descripción sin inglés (por campo) y datos no traducibles idénticos. Panel: búsqueda con éxito, sin novedades, un feed caído, ambos caídos y comprobación de disponibilidad con sus tres salidas; **idempotencia** (doble búsqueda: nada duplicado, visibilidad/descripciones/uniones intactas — SC-007); unir duplicados con doble confirmación y «no son el mismo sermón»; ocultar/mostrar (desaparece de todo lo público; reaparece por fecha; **sigue oculto tras la próxima búsqueda**); marcas con sus resoluciones y su no reapertura; límite de la descripción y español obligatorio cuando se rellena el inglés; permisos cruzados (solo «medios»); textos de §7 literales; 320 px sin desplazamiento horizontal.

---

## 7. Catálogo de textos

### 7.1 Panel (español; FR-024) — añadidos a los de F2/F3/F4/F6/F7

| Situación | Texto |
|---|---|
| Banda de ayuda | «Aquí miras y mantienes las grabaciones que el sitio trae desde el canal de YouTube y el feed del podcast. Pulsa “Buscar episodios nuevos” cuando se publique algo nuevo y revisa con calma lo que quede marcado; solo lo visible lo verá el público.» |
| Título de la zona de búsqueda | «Desde YouTube y el podcast» |
| Botón principal | «Buscar episodios nuevos» |
| Línea de apoyo de la zona | «Es seguro pulsarlo todas las veces que quieras: solo agrega lo que falte, nunca borra y nunca cambia lo que ya decidiste.» |
| Buscando (botón) | «Buscando…» |
| Buscando (`role="status"`) | «Buscando episodios nuevos en YouTube y en el podcast…» |
| Resumen: nuevos sin marcas | «Se agregaron {N} episodios nuevos; {A} se unieron con su video y su audio. Ya están visibles en “Medios”.» |
| Resumen: nuevos con marcas | «Se agregaron {N} episodios nuevos; {A} se unieron con su video y su audio; {D} quedaron marcados como posible duplicado. Revísalos abajo.» |
| Resumen: sin novedades | «No hay episodios nuevos: nada cambió. Cuando la iglesia publique algo más, vuelve a pulsar el botón.» |
| Resumen: un feed falló | «El feed de {YouTube o el podcast} no respondió. Se agregó lo que trajo {el otro}; nada se ocultó ni se marcó por esa causa. Vuelve a pulsar la búsqueda en un rato para recuperar lo que falte.» |
| Resumen: ambos lamentos | «Ninguno de los dos feeds respondió. No se agregó nada y no se cambió nada: todo lo que tenías está a salvo. Prueba más tarde.» |
| Resumen: revisión de enlaces | «La búsqueda revisó {M} episodios para ver si sus enlaces siguen funcionando: mira las marcas.» |
| Estado vacío del listado | «Todavía no hay episodios.» (+ segunda línea) «Pulsa “Buscar episodios nuevos” para que el sitio lea por primera vez el canal de YouTube y el feed del podcast.» |
| Filtro sin pendientes | «No hay nada pendiente de revisión: todo está al día.» |
| Ayuda del filtro | «Las marcas las pone el sistema cuando algo necesita tu revisión; se resuelven en cada fila.» |
| Ocultar | «Quedó oculto: el público no lo ve.» |
| Mostrar | «Ahora está visible en “Medios”.» |
| Mostrar de nuevo (confirmación) | «Si el enlace de “{nombre}” sigue sin funcionar, volverá a quedar oculto la próxima vez que se revise.» Botones «Mostrar» / «Dejarlo como está» |
| Descripción guardada (visible) | «Cambios guardados: ya visibles en la sección.» |
| Descripción guardada (oculto) | «Cambios guardados.» |
| Límite de la descripción | «{n} de {L} caracteres. Recorta {m} para poder guardar.» |
| Inglés sin español | «El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de la descripción corta.» |
| Ayuda del campo descripción | «Se muestra tal como la escribas, en la ficha del episodio. La descripción completa del feed no se publica.» |
| Ayuda «English (opcional)» | «Si lo dejas vacío, el público verá este contenido en su versión en español.» |
| Contexto del dialog de descripción | «Estás ajustando la descripción de:» |
| Título de la vista previa | «Así quedaría la ficha si las unes.» |
| Nota de la vista previa | «Al unir, el nombre publicado es el de YouTube.» |
| Confirmar unión | «Se unirán las dos entradas en una sola ficha. Esto no se puede deshacer.» Botones «Unir definitivamente» / «Cancelar» |
| Unión hecha | «Unidas en una sola ficha, visible en “Medios”.» |
| «No son el mismo sermón» | Acción «No son el mismo sermón» → «Listo: quedaron las dos como episodios distintos, visibles sin marca.» |
| Marca: posible duplicado (explicador) | «El sistema no pudo estar seguro de si estas dos entradas son el mismo sermón.» |
| Marca: no disponible (explicador) | «Se comprobó: el enlace de esta grabación ya no funciona (borrada o privada). El público no la ve.» |
| Marca: ya no aparece — funciona | «Salió de la ventana del feed de {YouTube o el podcast} (así es: la ventana es corta). Su enlace sigue funcionando y sigue visible para el público.» |
| Marca: ya no aparece — sin comprobar | «Salió de la ventana del feed de {YouTube o el podcast}. No se pudo comprobar si su enlace sigue funcionando; mientras tanto, sigue visible para el público.» |
| Resolución cerrada (común) | «La marca se cerró y no se vuelve a abrir por esta misma causa.» |
| Error del sistema al guardar/unir | «No se pudo completar la operación. Tus episodios están a salvo: vuelve a intentarlo en unos minutos.» |
| Sin permiso (servidor) | «No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.» (F2) |

> El resumen no menciona segmentos con número 0 (⟲ D-19). «Un feed falló» se nombra por su servicio cotidiano (⟲ D-21).

### 7.2 Sitio público — cadenas de interfaz es/en

> El **contenido** (nombre, descripción con su inglés si el equipo lo tradujo) llega del panel/feeds con respaldo (FR-025). Estas son las palabras de la interfaz.

| Clave | Español | English |
|---|---|---|
| `nav.sections.media` | Medios | Media |
| `media.page.title` | Medios | Media |
| `media.page.subtitle` | Las prédicas y los podcasts de nuestra iglesia; se abren en YouTube y Spotify, en una pestaña nueva. | Watch and listen to our church's sermons and podcast; they open in YouTube and Spotify in a new tab. |
| `media.loadMore` | Cargar más | Load more |
| `media.loading` | Cargando… | Loading… |
| `media.error.load` | No pudimos cargar los medios. Revisa tu conexión y vuelve a intentarlo. | We couldn't load the media. Check your connection and try again. |
| `media.error.retry` | Volver a intentar | Try again |
| `media.error.partial` | No pudimos actualizar; ves la última versión disponible. | We couldn't refresh; you're seeing the latest available version. |
| `media.link.youtube` | Ver en YouTube | Watch on YouTube |
| `media.link.spotify` | Escuchar en Spotify | Listen on Spotify |
| `media.link.newTab` | pestaña nueva | new tab |
| `a11y.media.link.youtube` | Abre YouTube en una pestaña nueva para ver {nombre}; sales del sitio de la iglesia. | Opens YouTube in a new tab to watch {name}; leaves the church's website. |
| `a11y.media.link.spotify` | Abre Spotify en una pestaña nueva para escuchar {nombre}; sales del sitio de la iglesia. | Opens Spotify in a new tab to listen to {name}; leaves the church's website. |
| `a11y.media.image` | Carátula del episodio {nombre} | Cover of the episode {name} |
| `a11y.media.card.heading` | Lista de episodios de medios | List of media episodes |
| `a11y.media.card.newTab` | (pestaña nueva) | (new tab) |
| `home.media.title` | Medios | Media |
| `home.media.subtitle` | Las últimas prédicas y podcasts | The latest sermons and episodes |
| `home.media.seeAll` | Ver todo | See all |
| `a11y.home.media.heading` | Medios: mira los episodios recientes y entra a la sección completa | Media: quick look at recent episodes and link to the full section |

> La redirección de la sección oculta (Q10) **no añade cadena alguna**: no hay aviso que mostrar; la única palabra posible es `media.loading` sobre el esqueleto. El anuncio de pestaña nueva va visible dentro del propio enlace (D-7) y completo en el `aria-label`.

---

## 8. Accesibilidad (WCAG 2.1 AA) e internacionalización

### 8.1 Accesibilidad

- **Teclado**: skip link → logo → navegación (barra; menú móvil *disclosure* de F10) → idioma → tema → contenido → pie. En la sección, el recorrido por ficha es: nombre (no interactivo) → **botón YouTube** → (botón Spotify, si hay) → ficha siguiente. «Cargar más», «Ver todo» y todos los controles reciben el anillo de foco de F10; nada depende de un gesto; sin trampas de foco.
- **Foco visible** (anillo 2 px por tema) en todo interactivo, incluidos los **enlaces externos en página y bloque**; un elemento no revelado que recibe foco se hace visible de inmediato (guarda de `Reveal`, F10).
- **Enlaces externos (FR-005/FR-030 — crítico)**: cada enlace — en ficha y en bloque — lleva (1) **servicio con palabra visible** («Ver en YouTube», «Escuchar en Spotify»), (2) **indicador visible de salida** (icono ↗ + «pestaña nueva» dentro del botón, ⟲ D-7), (3) `aria-label` completo con las **tres** informaciones (servicio, sale del sitio, pestaña nueva — §7.2), (4) `target="_blank" rel="noopener noreferrer"`. El anuncio viaja con el botón: nadie depende del recuerdo del subtítulo.
- **Semántica**: un `<h1>` por página («Medios»; el bloque usa `h2` vía `SectionHeading`); cada ficha `li` con `h3` (rol de lista conservado con `<ul>`); orden de lectura = orden del DOM = orden por fecha; los estados y marcas del panel siempre con palabra (`StatusPill`); el listado del panel es `Table` con encabezados `<th scope>` y tarjetas con etiquetas en móvil.
- **No solo color**: formatos por **chips con palabra** («Video»/«Audio»); visibilidad y marcas por **píldoras con palabra**; las marcas además con icono propuesto (⟲ D-30) y su **línea de explicación**.
- **Carátulas**: `<img>` con `alt` completo («Carátula del episodio {nombre}», ⟲ D-31); marcador neutro nunca hueco; **0 recursos de terceros** (SC-005: QA lo verifica por DOM/HTML — 0 `iframe`, 0 `src` de terceros).
- **Contraste y letra**: cuerpo 18 px (17 px en fichas compactas del bloque); Poppins 600 en el nombre con contraste AA en ambos temas (SC-008); el turquesa pequeño solo en su clase de contraste `#0e7a66` sobre claro; Bebas de rótulos igual que los demás bloques.
- **Zoom 200 % y 320 px**: diseño fluido en `rem` con `clamp`; 0 desplazamiento horizontal (SC-017); botones apilados y grilla a columna única; `overflow-wrap` sin romper el layout; recortes visuales siempre con el texto completo en el DOM.
- **Objetivos táctiles ≥ 44 px**: enlaces externos (a todo el ancho en móvil), «Cargar más», «Ver todo», ítems del menú móvil, botones del panel (el `PublicButton`/`Button` existente ya cumple).
- **Avisos**: estado de la búsqueda con `role="status"`; el resumen como `Notice` con su rol correcto; ocultar/mostrar con `Notice` temporal del efecto; `EmptyState` error con `role="alert"`.
- **`prefers-reduced-motion`**: sin revelados ni stagger; todo íntegro al instante (F10 incondicional); el estado «Buscando…» conserva su mensaje de texto (el spinner es lo único que se quita — ⟲ D-18).
- **Idioma del documento**: `html lang` por F3/F10; el contenido en español dentro de una página en inglés lleva `lang="es"` en su envoltorio (práctica heredada de F3/F4/F6).

### 8.2 Internacionalización (Decisiones 6 y 8, acotadas por Q4)

- **Interfaz** (siempre bilingüe): diccionarios `es.ts`/`en.ts`; las claves nuevas de F9 (§7.2) se suman; `LanguageProvider` y `localStorage` intactos; primera visita en español, sin auto-detección. El panel no cambia de idioma (FR-024).
- **Contenido** (por campo): **solo la descripción corta es traducible** (FR-025) — español base (obligatorio si el campo lleva contenido; FR-023) y English opcional desde el panel; **el fallback es→en lo resuelve el servidor** (el frontend nunca muestra el campo vacío ni pone marcas); el **nombre** es el del feed y no se traduce (FR-026); carátula, fecha (dato único con presentación por locale), enlaces y formatos **idénticos entre idiomas** (US4 esc. 5).
- El visitante en inglés ve el catálogo con lo no traducido **en español**, idéntico diseño y sin marca alguna (US4 esc. 2; caso normal de la Decisión 8).
- Caracteres especiales del feed (tildes, ñ, ¿ ¡, «:`, emojis) se conservan íntegros en panel y público, sin limpiezas ni transformaciones (Edge Case de la spec).

---

## 9. Decisiones de UX para revisar por el humano (resumen — puerta 2)

| # | Decisión | Alternativa |
|---|---|---|
| D-1 ⟲ | «Medios» es una **ruta real** al final de la lista única de F10, **después de «Ofrendas y diezmos»** (patrón F4/F6/F7; FR-007/Q8). | Ancla de portada (mezclaría el orden de lectura): no. |
| D-2 ⟲ | Subtítulo fijo de `/medios` que **anuncia salida del sitio y pestaña nueva** en la primera lectura (base del 90 % de SC-012). | Subtítulo neutro (el anuncio quedaría solo en los botones). |
| D-3 ⟲ | «Cargar más» **desaparece** cuando ya no queda nada por mostrar (sin mensaje «no hay más»). | Un cierre tipo «Eso es todo» (ruido extra). |
| D-4 ⟲ | Sin carátula → **marcador neutro** en el mismo marco 16:9 (la grilla no «salta» de forma; cero huecos). | Sacar la imagen del layout (la grilla temblaría entre fichas). |
| D-5 ⟲ | La ficha **no es clicable en bloque** (ni el nombre): los únicos interactivos son **los botones de enlace externo** (una ficha-enlace única sería ambigua con doble formato y permitiría salir del sitio sin querer). | Ficha-enlace al enlace principal (menos toques, ambiguo con dos formatos). |
| D-6 ⟲ | En las fichas la descripción **se recorta solo visualmente** (líneas tope por superficie) con el **texto completo en el DOM**; el recorte del dato real vive en el límite del panel (FR-020). | Nunca recortar visualmente (la grilla rompe con textos largos inusuales). |
| D-7 ⟲ | El anuncio de pestaña nueva es **visible dentro del botón** (icono ↗ + microtexto «pestaña nueva») **además** del `aria-label` completo. | Solo `aria-label` completo (menos ruido visual; pero SC-012 pide que el visitante común lo perciba: lo visible ayuda). |
| D-8 ⟲ | **Ocultar sin confirmación** (reversible); **«Mostrar de nuevo» desde la marca «No disponible» sí con ConfirmDialog** suave (es la única acción que puede exponer un enlace defectuoso). | Confirmar también el ocultar (una parada más semanal sin protección real). |
| D-9 ⟲ | **Línea de apoyo** bajo el botón de búsqueda que dice la idempotencia («es seguro pulsarlo…»): confianza para la primera vez. | Sin línea (se explica solo en la banda general). |
| D-10 ⟲ | **Filtro rápido de marcas** en el panel («Todo» / «Solo lo pendiente de revisión» con conteos): revisar marcas sin recorrer el listado. | Sin filtro (la spec no lo exige; con pocas marcas basta). |
| D-11 ⟲ | **Sin orden manual** (Q9): el listado del panel refleja el mismo orden por fecha del público; sin columna Orden ni drag & drop. | Orden manual (chocaría con Q9 y las fechas vienen del feed). |
| D-12 ⟲ | Un **error de carga** en `/medios` muestra mensaje + «Volver a intentar» y **no redirige**: solo el éxito con 0 visibles redirige. | Tratar error como «sin contenido» (el público perdería la sección por una falla pasajera). |
| D-13 ⟲ | Grilla pública de **máximo 2 columnas** en `/medios` (ficha alta; preserva la lectura; coherente con «más aire», decisión 15). | 3 columnas en pantallas grandes (más denso, menos aire). |
| D-14 ⟲ | Posición del `Reveal` del bloque en portada: conservar la alternancia con el bloque de ofrendas (decisión menor de implementación, como F7 D-15). | Fija por definición. |
| D-15 ⟲ | **Chips de formato por fila del panel** («Video»/«Audio», por palabra) y enlaces por ficha en el extracto: los formatos se ven, no se deducen contando botones. | Solo los botones (menos claro en el panel). |
| D-16 ⟲ | **Búsqueda: sin cola y sin confirmación previa**; botón en loading mientras corre y acciones de fila deshabilitadas (FR-012/FR-023 son el seguro). | Cola de búsquedas concurrentes (resúmenes confusos; nada lo pide). |
| D-17 ⟲ | El estado «Buscando» nombra los servicios como los llama la iglesia («en YouTube y en el podcast»), nunca tecnicismos. | Texto técnico (prohibido por FR-018). |
| D-18 ⟲ | Sin barra de progreso; un solo mensaje de estado (mejora opcional: avanzar por feed). El spinner cede con `prefers-reduced-motion`. | Barra de progreso (no cabe: la importación es atómica y rápida). |
| D-19 ⟲ | El resumen **no menciona segmentos en 0** (nunca «0 se unieron»). | Todos los conteos siempre (ruido de «0»). |
| D-20 ⟲ | El mensaje «sin novedades» es **neutro con expectativa** («cuando la iglesia publique algo más…»), no suena a error. | Mensaje seco «No hay episodios nuevos.» |
| D-21 ⟲ | Los feeds se nombran por su servicio cotidiano («el feed de YouTube», «el podcast»), nunca «Atom/RSS/Anchor» (FR-018: nada interno expuesto). | Nombres técnicos (FR-018 los excluye). |
| D-22 ⟲ | El resumen **permanece visible** hasta la próxima búsqueda o recarga (no es un toast): para comunicar el resultado a otros. | Toast de 3 s (el resultado se olvidaría). |
| D-23 ⟲ | El `Dialog` de descripción lleva **miniatura de contexto** (carátula + nombre + fecha) arriba: evitar ajustar el episodio equivocado (nombres parecidos, fechas contiguas). | Sin miniatura (dialog más corto; riesgo de error de ficha). |
| D-24 ⟲ | La **visibilidad no se cambia desde el formulario de descripción** (solo en la fila): «editar texto» y «decir si se ve» separadas (patrón F7 D-11). | Interruptor de visibilidad en el formulario (dos vías para lo mismo). |
| D-25 ⟲ | La **vista previa de la unión** muestra el nombre que ganaría (el de YouTube, default revisable #7) con la frase declaratoria — si el humano prefiere el del podcast, cambia solo esa línea y la vista previa. | Sin vista previa del nombre (revelación sorpresiva tras la unión). |
| D-26 ⟲ | **Unir: doble confirmación** (Dialog de detalle + `ConfirmDialog` danger «Esto no se puede deshacer.»): es la única operación irreversible del módulo. | Un solo paso «Unir» (deficiente contra un error no deshacible). |
| D-27 ⟲ | **«No son el mismo sermón» sin confirmación extra** (no destruye nada; reversible). | Confirmarlo también (uniformidad contraproducente). |
| D-28 ⟲ | **«Mostrar de nuevo» con ConfirmDialog** suave (consecuencia semanal explicada). | Directo (más rápido, riesgo público). |
| D-29 ⟲ | **La resolución de marcas es directa** («Mantener visible», «Mantener oculto»), con el mensaje que cierra explicando la no-reapertura por la misma causa; lo único peligroso («Mostrar de nuevo») lleva su barrera (D-28). | Confirmación para toda resolución (uniforme, más pasos para operaciones inocuas). |
| D-30 ⟲ | **Iconos por marca** (duplicado= dos fichas; no disponible= enlace roto; ya no aparece= reloj) además de la palabra; si el humano prefiere solo palabra, la píldora con `label` ya cumple «no solo color». | Solo palabra (más sobrio). |
| D-31 ⟲ | **`alt` completo en las carátulas** («Carátula del episodio {nombre}»). | Decorativa (`alt=""` — la carátula es información real de la ficha). |
| D-32 ⟲ | En el listado del panel, el **origen del episodio** («De YouTube»/«Del podcast») va dentro de los diálogos, no como columna; y en la ficha compacta del bloque, la **fecha se omite** (el orden ya la comunica; la ficha de la sección la da). | Columna «Origen» y fecha en el extracto (más info, más ruido/altura). |
| D-33 ⟲ | **Estado vacío del listado remite a la zona de importación** (un solo botón en toda la página): el FR-011 necesita una sola puerta. | Botón duplicado dentro del estado vacío (dos puertas para la misma acción). |

---

## 10. Notas para el `dev-frontend` / `qa-tester`

- **El bloque de portada usa su propio hook/endpoint** (`usePortadaMedios` → propuesta `GET /api/v1/portada/medios?lang=es|en`): tope 5 en servidor, orden de la sección, solo visibles, DTO mínimo, `no-store`. Pruebas contadas: 1, 2, 3, 4, 5 y **0 ítems** (bloque ausente, sin mensaje ni título — SC-015/SC-016).
- **La presencia en la navegación sale del documento de la sección** (misma separación de F6/F7): `usePublicMedios` alimenta la página, decide la **redirección con éxito y 0 visibles** (Q10) y, compartida por `queryClient`, alimenta `useSiteLinks` (entrada «Medios» solo con ≥ 1 visible). `usePortadaMedios` **no** es fuente de navegación.
- **Q10 (redirección, crítica)**: `/medios` con 0 visibles → `<Navigate to="/" replace>` **sin parpadeo** (con la consulta ya en caché, en el primer render; sin caché, solo el esqueleto mientras se decide), sin pintar H1 ni grilla vacía; el historial queda limpio. QA confirma que `/eventos` conserva su `HiddenNotice` y que `/ministerios` y `/ofrendas` también redirigen (patrones deliberadamente distintos, ya probados en el repo).
- **FR-004 (0 terceros, crítico)**: QA verifica por DOM/HTML que la sección y el bloque **jamás** piden o incrustan nada de YouTube/Spotify/Anchor: 0 `iframe`, 0 `src` de terceros, 0 reproductores. Las carátulas son rutas del propio sitio (copiadas al importar). El único contacto con esos servicios es el `href` que el visitante activa.
- **FR-005 (anuncio, crítico)**: cada enlace externo lleva `target="_blank"` + `rel="noopener noreferrer"` + `aria-label` con servicio/salida/pestaña nueva **más el indicador visible** (⟲ D-7). Prueba: el `aria-label` del enlace contiene el nombre del servicio, «pestaña nueva» y «sales del sitio» (y su equivalente en inglés).
- **FR-015/SC-007 (nada se pisa, crítico)**: tras una búsqueda: la visibilidad del equipo **no cambia**; la descripción ajustada **no se pierde**; la unión confirmada **no se deshace**; y un feed caído **no oculta ni marca nada**. La garantía está en el servidor; QA lo prueba con segunda búsqueda y feeds simulados.
- **FR-017 (regla del «estante de 15»)**: verificar las **tres salidas**: sigue disponible → **sigue visible** con marca «Ya no parece en el feed» y la línea «funciona»; ya no disponible → **oculto** con «No disponible»; comprobación falla → **visible** con la marca y la línea de «no se pudo comprobar». En ningún caso se oculta solo por dejar de aparecer en el feed.
- **Marcas y público (FR-009)**: nada de una entrada marcada sin resolver se publica por ninguna vía — sección, bloque, respuestas públicas. El panel es el único lugar donde existen las marcas.
- **D-8/D-28 (visibilidad)**: ocultar directo; «Mostrar de nuevo» con confirmación; QA prueba el ciclo ocultar → mostrar → ocultar → buscar → **sigue como el equipo lo dejó**.
- **D-26 (unión, definitiva)**: doble confirmación; tras unir queda una ficha con **nombre de YouTube** (default #7), carátula, fecha y los dos enlaces, sin marca; la entrada absorbida desaparece del listado; la auditoría registra la unión.
- **El diseño no cambia el sistema de F10**: `Card`, `SectionHeading`, `Reveal`, `useSiteLinks`, i18n, temas y tokens tal cual; 0 tokens nuevos; la página nueva es «otra sección», no una marca propia. La única extensión de lo existente es el id de navegación (⟲ D-1).
- **El panel no se rediseña** (F10B): componentes de F2/F3/F4/F6/F7 tal cual; un acceso en `AccesosRapidos` con `hasPermission('medios')`.
- **Datos del panel con Zod** espejo del contrato; **NUNCA `fetch` en componentes** (hooks en `features/medios/` y `features/publico/medios/`); tipos desde el contrato del arquitecto (sin `any`).
- **Textos SIEMPRE del diccionario en el sitio público** (§7.2), nunca JSX suelto (SC-006).
- **Idempotencia y concurrencia (D-16)**: botón y acciones deshabilitados durante la búsqueda; el servidor no aplica nada a medias (FR-023); una búsqueda con error deja el catálogo tal como estaba.
- **SC-001 (30 s)** y **SC-012 (90 %)** se verifican con el protocolo manual consolidado de F6/F7 tras el despliegue (incluida la lectura del anuncio de salida del sitio antes de activar el enlace).

---

## 11. Trazabilidad — qué pantalla cubre qué US/FR

| US / FR | Pantalla / componente | § |
|---|---|---|
| US1 · FR-002 · FR-003 · FR-004 · FR-005 · FR-006 | `/medios` (MediosPage, MediosGrid, MedioCard, CaratulaNeutra, MedioLinks, CargarMasButton) | §4.1, §2.1 |
| US1 esc. 7 · FR-008 · Q10 | Redirección silenciosa a la portada + navegación sin huérfanos (barra, menú, pie) | §4.1, §3 |
| US2 · FR-010 · FR-011 · FR-012 · FR-013 · FR-014 · FR-016 · FR-018 · FR-021 | `/panel/medios` (MediosPage, BuscarEpisodios, ResumenImportacion, EpisodiosList) + acceso en `AccesosRapidos` | §4.2, §4.3, §4.4, §5 |
| US3 · FR-015 · FR-017 · FR-020 · FR-021 · FR-022 · FR-023 | DescripcionForm, UnirDuplicadosDialog + ConfirmDialog, MarcaDeRevision y sus resoluciones | §4.5, §4.6, §4.7, §5 |
| US4 · FR-024 · FR-025 · FR-026 | LanguageSwitcher (F10), respaldo por campo (solo la descripción), datos no traducibles idénticos | §2.2, §7.2, §8.2 |
| US5 · FR-028 · FR-029 · FR-030 | Sistema F10 en sección y bloque: teclado/lector, 320 px sin scroll, 200 %, contraste claro/oscuro, `reduce-motion`, 44 px, anuncio completo de enlaces externos | §4.1, §4.9, §8.1 |
| **US6 · FR-027 · SC-015 (Q12)** | Bloque «Medios» en la portada (MediosBlock/MedioItem vía `usePortadaMedios`, tope 5, **después del bloque de ofrendas**), bloque ausente con 0 visibles, «Ver todo» | §2.7, §4.9 |
| FR-017 (regla del «estante de 15» aprobada) | Las tres marcas con explicador y resolución; el público jamás ve marcas | §4.7 |
| SC-005 (0 terceros) | 0 iframe/0 recursos externos; carátula copiada y servida por el propio sitio | §4.1, §4.9, §10 |
| SC-006 (enlaces externos correctos) | `PublicButton external` con `target`/`rel`/`aria-label` estándar en fichas y bloque | §4.1, §4.9, §8.1 |
| Auditoría (FR-022) | Módulo de Auditoría de F2 intacto; registra todas las operaciones de «medios», sin UI nueva | §4.10, §4.4 |
| Permisos (FR-021) | `RequirePermission(MEDIOS)`; el catálogo de F2 sigue en 10; sin permisos nuevos | §3, §2.6 |
