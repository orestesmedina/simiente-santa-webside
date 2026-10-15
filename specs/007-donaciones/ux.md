# UX: F7 — Ofrendas y diezmos (donaciones)

> Documento de diseño para `disenador-ux`. Fuentes: `spec.md` de F7 (v2, aclaraciones Q1–Q11 resueltas por el humano el 2026-10-14), `ux.md` de F6 (bloque de portada, orden manual, redirección de sección oculta, textos del panel), `ux.md` de F4 (modelo de formato, bloque «Lo que viene», patrones de panel) y `ux.md` de F3/F10 (fallback, estados, textos), más el código actual de `frontend/src/` (React 19 + Tailwind 4; componentes de F2/F3/F4/F6/F10 ya construidos y verificados en este repositorio).
> No es código: lo implementa el `dev-frontend`. Rutas y contratos los fija el `arquitecto`; aquí se proponen y los supuestos van marcados.
> **Decisiones de UX a revisar por el humano**: ver §9 (resumen). Cada una lleva su marca ⟲. Reutilizo los patrones ya aprobados de F4 (D-25/D-28: extracto en fila sin carrusel), F6 (D-14: redirección a la portada; D-6: orden por campo numérico) y aquí solo marco lo nuevo de F7.
> F7 **aplica** el sistema de F10 y los patrones de F2/F3/F4/F6; **no redefine** tokens, navegación, selector de idioma, modo oscuro ni el rediseño del panel (F10B).
> **Distinción de nombres (Q8, no negociable)**: la **sección pública** es «Ofrendas y diezmos» en `/ofrendas`; el **módulo y el permiso del panel** siguen llamándose **«donaciones»** (diferencia deliberada: en la UI se usan los dos nombres según superficie, nunca mezclados).

---

## 0. Principios de diseño (heredados, aplicados a F7)

1. **La exactitud del número es sagrada.** Una cuenta mal leída hace perder un aporte: el número se muestra **íntegro, agrupado para leerlo y copiable en un toque** (o transcribible en bloques); jamás se recorta con elipsis, se parte en dos líneas de forma arbitraria ni se «embellece» con transformaciones (FR-003, SC-009). Copiar getValue el número **normalizado** (tal como se guardó, FR-015/FR-016), nunca la versión con espacios de presentación.
2. **Datos y textos, cero imágenes — y se nota.** F7 no lleva ninguna imagen (Q6, FR-020): ni encabezado, ni logos de bancos, ni íconos decorativos de catálogo. Las fichas son tipografía y estructura; la ausencia de imagen **no queda como hueco**: la ficha nace sin zona de imagen (patrón «sin hueco, sin reserva» de F4/F6).
3. **Dos tipos, cero confusión.** Cuentas IBAN y números de SINPE Móvil viven en **dos grupos separados y rotulados** de la misma página, IBAN primero (Q1); dentro de cada grupo, el orden manual del equipo. El visitante distingue el tipo por el **grupo y (en la portada) por el rótulo de la tarjeta**, no por deducción (FR-004, SC-013).
4. **«Sin dato» no deja huella.** Un método sin titular no muestra línea de titular — ni vacía ni con marcador «sin datos» (Q2) —; un texto (introductorio o de uso) en borrador simplemente no se dibuja y la sección queda coherente (US3 esc. 7). Cero campos en blanco en el 100 % de los casos.
5. **Aporte seguro, never a medias.** Cada número y cada texto tiene su **propio** estado borrador/publicado (Q10, FR-012): nada llega al público revisado a medias; y lo ya publicado que se edita se corrige **de inmediato** (FR-013) — corregir un número erróneo no puede esperar.
6. **Apto para todas las edades.** Cuerpo 18 px, objetivos ≥ 44 px, iconos siempre con texto, copia con texto visible (no solo icono), `prefers-reduced-motion` intacto, 0 desplazamiento horizontal a 320 px (FR-024, FR-025; `idea.md` §7: personas mayores que saben lo básico de internet — aquí transcribir un número es la operación crítica).
7. **Solo lo publicado existe.** Sin ningún método publicado (y sin textos publicados), la sección **no existe** — ni en la barra, ni en el menú, ni en el mapa del sitio ni en la portada — y si alguien abre `/ofrendas` a mano es **redirigido a la portada** de forma limpia, sin aviso y sin parpadeo (Q9, patrón de F6, decisión 20b; **distinto a propósito** de F4, que mantenía `HiddenNotice`). Cero secciones vacías, cero enlaces huérfanos.
8. **El panel hereda F2/F3/F4/F6 tal cual.** Título + banda de ayuda, `Table`/tarjetas, `StatusPill`, `Dialog`, `ConfirmDialog`, `Notice`, `Tabs` de idioma, validación con Zod y mensajes en español: cero estilos nuevos (FR-011, FR-018).

---

## 1. Tokens aplicados (la verdad vive en F3/F10; solo se citan las decisiones de F7)

| Tema | Reutilizado de | Nota específica de F7 |
|---|---|---|
| Paleta 70/20/10 | F3 §1.1 / F10 §0.1 | Navy dominante, tarjetas en crema/vidrio, acento turquesa en rótulos y foco. Sin colores nuevos. |
| Tipografías | F3 §1.2 / F10 §0.2 | H1 de página y rótulos de grupos en Bebas; cuerpo, fichas y números en Poppins. **El número de cuenta en Poppins 600 con `letter-spacing` amplio y tabular.** F7 no usa Playfair. |
| Espaciado | F10 §0.3 | El introductorio, cada grupo y el texto de uso se separan con `--space-section`; las fichas dentro de un grupo con `--space-block`. |
| Superficies | `Card` (`solid`/`glass`, F10) | Cada método es una `Card solid` **sin zona de imagen** (la tarjeta nace textual); el bloque de portada vive como el de F6 (fila de `Card solid` compactas). |
| Movimiento | F10 §5 (`Reveal`/`RevealGroup`, 600 ms / micro 180 ms) | Revelado lateral una sola vez por grupo (stagger suave, tope 240 ms). **El botón de copiado usa el feedback de micro (180 ms), sin animaciones del número**: un número que «late» o se corroe visualmente invita a dudar de su exactitud. Con `prefers-reduced-motion`: todo íntegro e instantáneo. |
| Foco visible | F3 §1.4 / F10 §6 | Anillo 2 px: navy sobre claro, crema sobre navy; en los botones de copia y en los títulos-enlace del extracto. |
| Chip «Cuenta IBAN» / «SINPE Móvil» | patrón del chip neutro de F4 (D-2/D-25) y F6 (D-11) | Píldora con borde y texto grises, Poppins 0.875 rem, **sin color semántico**: identifica el tipo en el extracto de portada (en la sección lo nombra el rótulo del grupo, y la ficha no repite chip). |
| Orden manual | precedente `sort_order, id` de F3/F4/F6 (FR-004) | Valor numérico editable **dentro de cada grupo** en el panel; en público el orden llega listo del servidor. Sin drag & drop. |

**Nuevos tokens: ninguno.** F7 usa exclusivamente los de F3/F10.

**Del inventario de F4/F6 NO se usa** (con razón de negocio, no de estilo): `DateBadge` y `formatoFechas.ts` (no hay fechas: la donación no vence), `CategoryChip` (el tipo se nombra con el grupo/rótulo, no con catálogo), `MapLink` (no hay lugares ni mapas), `ImageUploader`/`CardImage` (Q6: sin imágenes, y esto es regla de negocio, no de estilo — el formulario no ofrece carga alguna), `WhatsappJoinButton` (sin enlaces externos: la persona dona por sus propios medios; Out of Scope), la marca «Pasado» (nada vence). F7 tampoco usa `Pagination` (sin paginación, defaults de la spec) ni `DateRangeFilter` (no hay rango temporal).

---

## 2. Flujo de usuario

### 2.1 Visitante: sabe cómo apoyar (US1, camino feliz)

1. Abre el sitio (`/`, idioma según preferencia guardada; primera visita: español). En la portada, **después del bloque de ministerios de F6**, ve el bloque **«Ofrendas y diezmos»** (§4.9): hasta **5** fichas compactas sin imagen — cada una con su tipo (Cuenta IBAN / SINPE Móvil), entidad y número — y el botón **«Ver formas de ofrendar»** hacia la sección completa (FR-022).
2. Toca «Ofrendas y diezmos» (barra en escritorio; ítem del menú hamburguesa en móvil; también en el mapa del sitio del pie) o el botón del bloque → llega a **`/ofrendas`**: página de contenido real con **H1 «Ofrendas y diezmos»**.
3. Lee primero el **texto introductorio** («cómo donar», Q11): uno o varios párrafos que explican cómo puede aportar y remarcan que los números son los de la iglesia — tal como el equipo lo escribió, completo y con sus saltos (FR-021).
4. Baja al grupo **«Cuentas IBAN»** (rótulo de bloque): cada ficha muestra tipo, **número íntegro agrupado en bloques** (p. ej. `CR12 BANK 0000 0000 0000 0000` — el agrupado es presentación; copia lleva el número normalizado), entidad financiera, titular (solo si existe) y moneda (**colones o dólares** — línea propia de la ficha, Q2). Botón **«Copiar número»** junto al número (§4.1).
5. Baja al grupo **«SINPE Móvil»**: cada ficha muestra el número en su **formato normalizado** (con prefijo, agrupado para lectura), entidad y titular (solo si existe) — **sin línea de moneda** (Q2): la ficha SINPE simplemente no la lleva, sin marcas de «sin moneda».
6. Termina con el **texto de uso de las donaciones** (Q5): uno o varios párrafos que explican en qué se invierten los aportes (obra social, mantenimiento, misiones…) — texto libre, no un listado — tal como el equipo lo escribió.
7. Quiere transferir: copia el número (un toque; si lo transcribe, los bloques de 4 le sirven de guía de lectura, US5 esc. 1) y lo pega en su app bancaria fuera del sitio — el sitio no cobra ni procesa nada (Out of Scope, FR-001).
8. Termina y recorre el resto: pie de F10 intacto — el mapa del sitio muestra «Ofrendas y diezmos» en el grupo **Explorar** (junto a «Eventos y actividades» y «Ministerios»), solo si hay algo publicado (FR-005/FR-006).

### 2.2 Visitante: cambia a inglés (US4)

1. En la barra (o el pie), toca **English** en el selector de F10.
2. La página cambia de inmediato: la **interfaz** («Offerings & tithes», «Cuentas bancarias / Bank accounts», «Copiar número / Copy number», «Cómo usamos las donaciones / How we use the offerings») va siempre del diccionario es/en; el **contenido** con versión en inglés (texto introductorio, texto de uso, titular, entidad si el equipo lo tradujo allí — FR-007 aplica a los campos de texto que el equipo llenó en inglés) se muestra en inglés.
3. Un campo **sin inglés** se muestra en español, idéntico diseño, **sin hueco ni marca de «sin traducir»** (FR-008; el fallback lo resuelve el servidor).
4. **El número, el tipo, la entidad y la moneda no cambian en absoluto como dato** (US4 esc. 5 · FR-007): no son contenido traducible; el número sigue siendo copiable con exactitud en los dos idiomas. La moneda es el mismo dato («colones»/«dólares»); su **etiqueta de interfaz** sí se traduce (⟲ D-6, §8.2).
5. La elección persiste al navegar (portada incluida) y entre visitas (FR-009); `html lang` pasa a `en`.

### 2.3 Equipo: gestiona la información (US2, camino feliz)

1. Inicia sesión → **Inicio** del panel: un acceso nuevo en `AccesosRapidos` — **«Gestión de donaciones»** (con permiso «donaciones») → `/panel/donaciones`. Cuenta sin el permiso: no ve el acceso y la ruta está protegida.
2. En la página (título «Gestión de donaciones» + banda de ayuda) ve **dos zonas**: **«Textos de la sección»** arriba (dos tarjetas: introductorio y uso, cada una con su píldora Borrador/Publicado y su botón Editar) y debajo el **listado de métodos agrupado por tipo**: bloque «Cuentas IBAN» y bloque «SINPE Móvil» (Q1), cada uno con su tabla (desktop) / tarjetas (móvil).
3. Cada fila muestra: **Orden** (el número que definió, dentro de su grupo), **Número** (normalizado), **Entidad**, **Titular** («—» si vacío), **Moneda** (solo fila IBAN; «—» no aplica en SINPE), **Estado** (`Borrador`/`Publicado`) y acciones.
4. Toca **«Agregar cuenta»** (en el bloque IBAN) o **«Agregar número»** (en el bloque SINPE) → `Dialog` con `DonacionForm` (§4.4): pestañas de idioma, tipo fijado por el bloque de origen, campos obligatorios/opcionales según el tipo, validación del número con mensajes en claro y Orden numérico.
5. Guarda → queda **Borrador**: visible en el panel, invisible al público; aviso «Quedó en borrador…».
6. Toca **«Publicar»** en su fila (sin confirmación) → píldora «Publicado» + aviso. El visitante YA lo ve (sección y, si entra en los primeros 5, bloque de portada).
7. **Edita un método publicado** y guarda → aviso **«Cambios guardados: ya visibles en la sección.»** (FR-013). Corregir un número de cuenta con error es minutos, no días (motivo de US2).
8. **Reordena** cambiando el campo Orden dentro de su grupo y guardando → el nuevo orden es exactamente el que ve el visitante dentro de ese grupo (US2 esc. 10, FR-004).
9. Edita los **textos** → `Dialog` con `TextoForm` (§4.5): pestañas de idioma, textarea de uno o varios párrafos, y el guardado respeta el estado propio de cada texto (Q10); si estaba publicado, «Cambios guardados: ya visibles en la sección.»
10. **Retira** («Retirar de la sección» con `ConfirmDialog` suave) o **elimina** («Eliminar», confirmación explícita de no-recuperable) — ambas quedan en la auditoría de F2 (FR-017).

### 2.4 Equipo: publica o retira por elemento (US3)

- **Borrador** → nadie lo ve fuera del panel (FR-012); en la portada tampoco (US6 esc. 6).
- **Publicado** → visible en `/ofrendas` en su grupo y orden, y (si entra en los primeros 5 del orden de la sección) en el bloque de la portada.
- Texto (introductorio o uso) con **su propio** estado: un texto en borrador no se dibuja en público aunque haya métodos publicados (Q10); los tres estados son independientes.
- Sin **nada** publicado (ni métodos ni textos) → la sección entera desaparece de la navegación (barra, menú, mapa), el bloque de la portada no se dibuja (FR-005, FR-022) **y** `/ofrendas` redirige a la portada (Q9).
- El listado del panel muestra SIEMPRE el estado real al público hoy (US3 esc. 6); no hay estados intermedios ni marcas de tiempo («Pasado» de F4 no existe aquí: nada vence).

### 2.5 Errores del equipo

| Situación | Camino |
|---|---|
| Campo obligatorio vacío (número o entidad; tipo viene fijado) | Nada se guarda: error **junto al campo** + resumen arriba del formulario; el foco va al primer campo invalidado (FR-014). |
| IBAN con estructura o longitud no válida | «Ese número no parece una cuenta IBAN válida. Revisa la cantidad de letras y dígitos: vale tanto el IBAN de Costa Rica como uno internacional.» Sin guardar nada (FR-015). |
| SINPE que no es un celular costarricense de 8 dígitos | «Ese número no parece un celular de Costa Rica: son 8 dígitos, con o sin el prefijo +506. Revisa que no falte o que sobre ningún dígito.» Sin guardar nada (FR-016). |
| IBAN/SINPE escrito con espacios o guiones | **Válido si su estructura es válida**: se guarda normalizado (FR-015/FR-016); el formulario lo avisa con texto suave bajo el campo («Se guardará sin espacios ni guiones — así lo verá el público agrupado en bloques.» ⟲). |
| Versión en inglés de un campo de texto sin su versión en español | Se bloquea al guardar indicando qué campo falta en español (regla por campo, Decisión 8); **un campo vacío no obliga a nada** (Q2, US2 esc. 8). |
| Límites de extensión | Contador vivo al escribir; al exceder, «Recorta X para poder guardar» (con el límite vigente — default 2.000 para los dos textos, revisable: Assumptions). |
| Moneda sin elegir en una cuenta IBAN | «Elige la moneda de la cuenta: colones o dólares.» Sin guardar (FR-014: obligatoria solo dentro de IBAN). |
| Moneda en un SINPE Móvil | El formulario no ofrece el campo (Q2); no hay error posible. |
| Duplicado (mismo tipo, misma entidad y mismo número) | **Aviso de duplicado; nada se guarda** («Ya existe esa misma cuenta/número.» — default revisable de la spec, Edge Cases). |
| Caracteres especiales (tildes, ñ, ¿ ¡, emojis) en los textos | Se conservan íntegros (FR-019/Edge Cases); los **números** limitan su entrada a lo que cada formato admite (estructura IBAN / dígitos del SINPE) — la validación lo explica en palabras, no mensajes técnicos (FR-015/FR-016). |
| Error del sistema al guardar | Aviso genérico, sin detalles internos, **sin perder lo escrito** (los valores permanecen en el formulario). |
| Operación sin permiso «donaciones» | Ruta protegida (`RequirePermission`) + página «sin permiso» de F2; si el servidor lo rechaza en caliente: mensaje sin detalles (FR-011). Tener otro permiso no da acceso a este (probado cruzado en US2 esc. 9). |
| Dos personas editando el mismo elemento | Se conserva la última versión completa guardada (default de la spec, como F3). |
| Borrador por cualquier vía | Jamás se expone: ni en HTML, ni en JSON, ni en el bloque de portada (SC-002). |

### 2.6 Persona sin permiso (el caso desfavorable)

- Sin permiso «donaciones»: no existe el acceso «Gestión de donaciones» (ni en `AccesosRapidos` ni en la navegación del panel); si fuerza la URL `/panel/donaciones`: `RequirePermission` → `/sin-permiso` (F2, texto intacto).
- Tener «donaciones» NO da los otros módulos, y ningún otro permiso da acceso a este: los permisos son independientes (Decisión 5; probado en US2 esc. 9). F7 no crea permisos nuevos ni cambia el nombre del permiso (Q8, FR-011).
- El sitio público no pide sesión: no opera el caso «visitante sin permiso».

### 2.7 Visitante: descubre cómo apoyar desde la portada (US6, Q7)

1. Abre el sitio y, **sin navegar más que el scroll**, ve en la portada el bloque **«Ofrendas y diezmos»** — **inmediatamente debajo del bloque de ministerios de F6** (Q7, FR-022) —: hasta **5** fichas compactas **sin imagen**, en el orden de la sección (IBAN primero, luego SINPE; orden manual de cada grupo), **cada una identificada con su rótulo de tipo** («Cuenta IBAN» / «SINPE Móvil») porque en la portada no hay grupos rotulados.
2. Cada ficha compacta muestra: rótulo de tipo, **entidad financiera** (el «título» de la ficha), **número** en su presentación de lectura (sin botón de copiar en el extracto: la copia vive en la sección) y titular si existe; moneda solo en IBAN.
3. Quiere copiar, ver todos o leer en qué se usan los aportes → toca el botón **«Ver formas de ofrendar»** del bloque → `/ofrendas` en un solo paso (US6 esc. 5).
4. Con más de 5 publicados: exactamente los primeros 5 según el orden de la sección, **sin paginación, carrusel ni scroll lateral** (US6 esc. 2; SC-014); con 1–4: exactos, sin huecos ni marcadores vacíos (US6 esc. 3).
5. Con **ninguno** publicado: el bloque **no se dibuja** — sin título, sin mensaje, sin aviso, sin reservar espacio — y la portada queda coherente (US6 esc. 4, FR-022). **Diferencia deliberada con F4**: el bloque «Lo que viene» sí muestra su estado vacío con mensaje; este, como el de ministerios, no. Un borrador jamás aparece (US6 esc. 6).
6. En inglés: interfaz del bloque traducida y contenido con fallback al español según FR-008, sin marcas; el número no cambia (US4 esc. 5).

---

## 3. Rutas propuestas (React Router)

```
Sitio público (sin sesión):
/                       Portada de F3 + bloque F4 + bloque F6 + bloque F7 (OfrendasBlock)
/ofrendas               «Ofrendas y diezmos» de F7 (OfrendasPage)
/health                 Estado del sistema (F1), fuera de la navegación pública

Panel (patrón F2/F3/F4/F6, sin rediseño):
/panel                  Inicio (acceso nuevo por permiso)
  /panel/donaciones     RequirePermission(DONACIONES)  Gestión de donaciones
  /panel/ministerios /panel/eventos /panel/… (F4/F6 y demás, intactos)
```

- **`/ofrendas` es una ruta real**, como `/eventos` y `/ministerios` (mismo patrón aprobado por el humano en F4/F6). Se añade como **ruta** al final de la **única lista de enlaces** de F10 — después de «Ministerios» — en `useSiteLinks.ts`: `SiteLinkId` se extiende con el id nuevo (propuesta: `'offerings'`; el nombre exacto lo fija el arquitecto) y el vocabulario con la clave i18n `nav.sections.offerings`. El `<nav>` sigue siendo una sola lista en DOM (la de F10/MobileMenu funciona sin cambios); no se toca `nav.label`. ⟲ D-1
- **Presencia en la navegación**: la entrada «Ofrendas y diezmos» existe **solo** si la consulta del **bloque de la portada** (`usePortadaOfrendas`) trae al menos un método publicado — **mismo criterio puerta 2 de F6** (dos consultas: el bloque alimenta barra, menú y pie por queryClient; `/ofrendas` se alimenta por su cuenta con `usePublicOfrendas`). Sin publicados: ni barra, ni `MobileMenu`, ni pie, ni bloque (FR-005).
- **Mapa del sitio del pie**: «Ofrendas y diezmos» entra en el grupo **Explorar** (no se abre columna nueva), mismo criterio D-7 de F4.
- **Sección oculta abierta a mano**: `/ofrendas` con 0 publicados **redirige a la portada** (Q9, decisión del humano 2026-10-14). Es **distinto a propósito** de `/eventos` de F4 (que muestra `HiddenNotice` sin redirigir) e **igual por decisión** a `/ministerios` de F6: no unificar los comportamientos. Mecánica sin parpadeo: ver la fila «Sección oculta» de §4.1.
- **Permission code**: el módulo se autoriza con el permiso existente **`donaciones`** de F2 (`RequirePermission`); su **nombre no cambia** (Q8) y el código exacto lo confirma el arquitecto frente al catálogo sembrado de F2. No hay permiso nuevo.
- **Portada**: el bloque se inserta en la secuencia de `HomePage` **inmediatamente después del bloque de ministerios de F6** (Q7, FR-022) y antes de la siguiente sección de F3. Dirección de entrada con `Reveal`: la de F6, para no romper la alternancia entre bloques contiguos (decisión menor de implementación; ⟲ supuesta, ver §9).

---

## 4. Pantallas

### 4.1 Sitio público — `/ofrendas` (visitante; US1, US4, US5)

**Propósito**: en una lectura (SC-001: menos de 30 s) el visitante sabe **a qué cuenta o número aportar** — con el número exacto listo para copiar o transcribir — y **en qué se usan las donaciones**, sin preguntar por WhatsApp ni Facebook (SC-012).

**Estructura** (página de contenido bajo `PublicLayout` de F10):

1. **Enlace de salto** (_skip link_) — de F10, intacto; `<main>` de `/ofrendas` es su destino aquí.
2. **H1 «Ofrendas y diezmos»** + subtítulo fijo de una línea (cadena de interfaz, no campo del panel — mismo criterio D-2/D-10 de F6/F4): «Las formas de apoyar económicamente a nuestra iglesia.» / («The ways to support our church financially.») ⟲ D-2
3. **Texto introductorio** (si está publicado; Q11): párrafo(s) a todo el ancho de lectura (columna cómoda, máx. ~65ch desktop), Poppins 18 px, saltos preservados, tal como el equipo lo escribió (FR-021). En borrador: **no se dibuja** — sin hueco ni «-placeholder», la sección sigue coherente (US3 esc. 7).
4. **Grupo «Cuentas IBAN»** (`SectionHeading` con kicker `01` del sistema, título en Bebas): todas las cuentas IBAN publicadas en el orden manual, con `RevealGroup` y stagger suave. Si no hay ninguna publicada, **el rótulo no se dibuja** (cero rótulo huérfano, mismo criterio D-5 de F4).
5. **Grupo «Números de SINPE Móvil»** (`SectionHeading`, kicker `02`): igual, tras el grupo IBAN (Q1).
6. **Texto de uso de las donaciones** (si está publicado; Q5): mismo tratamiento del introductorio, separado por `--space-section`, precedido por un **título de bloque menor** «¿En qué usamos las donaciones?» / («How we use the offerings») — cadena de interfaz fija ⟲ D-3 (el contenido del texto es del panel; solo el rótulo es interfaz).
7. **Pie de F10 intacto**: mapa del sitio con «Ofrendas y diezmos» en **Explorar** (solo si hay algo publicado), selector de idioma, tema claro/oscuro por tokens.

**Composición de una ficha (`DonacionCard`, `Card solid` del sistema, sin zona de imagen)** — de arriba a abajo, **cada elemento solo existe si tiene dato**:

- **Tipo: no repite chip en la sección** — lo nombra el grupo en el que vive (mismo criterio de F4 §4.1 con «Evento/Actividad»; el chip de tipo es solo para el extracto de portada, donde no hay grupos).
- **Entidad financiera** (h3, Poppins 600 1.25rem como el h3 de F3/F6; **no** Bebas grande: no compite con el H1): «Banco Nacional de Costa Rica».
- **Número** (el protagonista; bloque `NumeroDonacion`, §6.3):
  - Presentación de lectura: **IBAN agrupado en bloques de 4** separados por espacios (`CR12 0151 0000 0011 2233 44`), Poppins 600, letra tabular, un tamaño que priorice la exactitud antes que el ancho — **nunca recortado con elipsis ni partido por guiones propios**; `overflow-wrap: anywhere` tolera el reacomodo en 320 px **sin perder ni un carácter a la vista** (FR-024, SC-009).
  - **SINPE Móvil** en su formato normalizado presentado agrupado para lectura (propuesta: `+506 8888 8888`; la cadena exacta normalizada la fija el plan — Assumptions de la spec). ⟲ D-4
  - **Botón «Copiar número» / («Copy number»)** junto al número: `PublicButton` variante secundaria pequeño (≥ 44 px de alto), icono de copiar **+ texto** (el icono nunca solo), al pulsarlo copia al portapapeles **el número tal como se guardó — normalizado, sin espacios de presentación** (FR-003: el valor del número no se altera) y muestra la confirmación `role="status"` «Número copiado.» / («Number copied.») por 2 s (revisable). Si el portapapeles no está disponible (navegador o discapacidad), la ficha muestra la ayuda de transcripción: «Cópialo a mano por bloques, de izquierda a derecha.» — que siempre acompaña al número como texto pequeño (§ abajo). ⟲ D-5
  - **Ayuda de transcripción** (siempre visible, Poppins pequeña bajo el número): por ejemplo «Puedes copiarlo en un toque o escribirlo tal cual, bloque por bloque.» / («You can copy it with one tap or type it out block by block.») — pensada para personas mayores (US5; el paralelo con «sin gestos desconocidos» de US1). ⟲ D-5
  - El número como texto lleva `aria-label` completito con el número entero sin separadores (ver §8.1) para que el lector lo lea de corrido sin «cuatro, cuatro» confuso.
- **Moneda** (solo IBAN, Q2): línea con palabra explícita, no un icono de moneda a solas: **«Moneda: colones»** / «Moneda: dólares» / («Currency: colones (CRC)» / «dollars (USD)» — ⟲ D-6, texto de interfaz fijo + dato; los dos idiomas muestran el mismo dato, la palabra sí es de interfaz).
- **Titular** (solo si existe, Q2): línea con icono 👤 + nombre tal como el equipo lo escribió («Cuenta en dollars», llamada «Titular» — el texto fijo «Titular» / («Account holder») más el nombre). Sin titular: **el bloque no existe** — nunca una línea vacía ni «—» en público (US1 esc. 5).
- La `Card` **no es un enlace en bloque**: el único interactivo de la ficha es el botón de copiar (regla de F4/F6 §4.1).

**Propuesta wireframe desktop y móvil** (esquema de referencia, no maqueta):

```
DESKTOP
[ BARRA F10: 🌱 Simiente Santa · Quiénes somos · Horario · WhatsApp · Contacto · Redes
                                      · Eventos y actividades · Ministerios · Ofrendas y diezmos | Español/English | 🌙 ]
══════════════════════════════════════════════════════════════════
  OFRENDAS Y DIEZMOS  (h1, Bebas)
  Las formas de apoyar económicamente a nuestra iglesia.
        ┐ --space-block
  Tu aporte va directo a la obra: las transferencias y el SINPE
  Móvil son la vía más simple para ofrendar. (introductorio del panel)
        ┐ --space-section
  01 CUENTAS IBAN (SectionHeading + acento turquesa)
  ┌─ Card ─────────────────────────┐  ┌─ Card ─────────────────────────┐
  │ Banco Nacional de Costa Rica   │  │ BAC San José                   │
  │ CR12 0151 0000 0011 2233 44    │  │ CR01 0102 2200 3344 5566 77    │
  │ [⧉ Copiar número]              │  │ [⧉ Copiar número]              │
  │ puedes copiarlo… o escribirlo  │  │ titular: Simiente Santa        │
  │ por bloques                    │  │ Moneda: dólares                │
  │ Moneda: colones                │  └────────────────────────────────┘
  │ (sin titular: no línea)        │
  └────────────────────────────────┘
        ┐ --space-section
  02 NÚMEROS DE SINPE MÓVIL
  ┌─ Card ─────────────────────────┐
  │ Banco Promerica                │
  │ +506 8888 8888                 │   (sin línea de moneda: no aplica)
  │ [⧉ Copiar número]              │
  │ titular: Tesorería Simiente Santa │
  └────────────────────────────────┘
        ┐ --space-section
  ¿EN QUÉ USAMOS LAS DONACIONES? (rótulo menor)
  Tu aporte sostiene la obra social, el mantenimiento del templo
  y las misiones. (texto de uso del panel, completo, saltos intactos)
        ┐ --space-section
  PIE DE F10 (mapa del sitio: Explorar incluye «Ofrendas y diezmos»)
══════════════════════════════════════════════════════════════════

MOBILE (≤ 640 px) (columna única, según F10):
[ SS Simiente Santa · Español/English · 🌙 · ☰ ]
→ ☰ abre panel (la MISMA lista, con Ofrendas y diezmos)
grilla 1 col: número a todo el ancho, botón de copiar a todo el ancho (≥44 px)
```

**Estados de la página pública**:

| Estado | Qué ve el visitante |
|---|---|
| Carga (primera) | Esqueleto con medidas reales: H1 + bloques de `Card` fantasma **sin zona de imagen** (rotulado como F10), patrón `SectionSkeleton`; la barra, el idioma y el tema ya visibles; cero brincos de layout. |
| Éxito completa | Introductorio + grupo IBAN + grupo SINPE + texto de uso, tal como el equipo los dejó publicados. |
| Éxito parcial (un grupo vacío) | Solo se dibuja el grupo con contenido; el rótulo del grupo vacío no existe (cero rotulo huérfano, D-5 de F4). |
| Un texto en borrador | Ese texto no se dibuja; el resto de la sección queda coherente, sin huecos ni marcadores (US3 esc. 7). |
| **Sección oculta** (0 publicados entre métodos y textos; Q9) | La entrada **no existe en la barra, ni en el menú, ni en el mapa del pie** (FR-005). Si alguien abre `/ofrendas` a mano — enlace guardado o URL tecleada — **redirige a la portada** con `<Navigate to="/" replace>`: **nunca se pinta una sección vacía**, jamás `HiddenNotice`, jamás mensaje, jamás página en blanco previa. Mecánica sin parpadeo (patrón F6 §4.1): la decisión de redirigir la toma la página con **su propia consulta** (`usePublicOfrendas`); si el dato ya está en caché, la redirección ocurre **en el primer render**, sin esqueleto; si aún no llegó, la página muestra **solo el esqueleto estándar de carga** mientras se decide; al confirmarse 0 publicados, la redirección ocurre **en el mismo ciclo de render** en que llega el dato, con `replace` para no dejar `/ofrendas` en el historial. La persona llega a la portada **sin sentir que algo falló**: sin errores, sin texto, sin saltos. **Distinto a propósito de `/eventos` de F4**, que mantiene al visitante con `HiddenNotice`: no unificar. |
| Error de carga sin nada | H1 + mensaje bilingüe claro: «No pudimos cargar la información de ofrendas. Revisa tu conexión y vuelve a intentarlo.» + botón «Volver a intentar» (patrón F3/F10). |
| Error con caché previa | Se mantiene el último contenido válido y un aviso discreto: «No pudimos actualizar; ves la última versión disponible.» (F3 §4.1). |
| Fallback es→en (US4) | El campo sin inglés se muestra en español, idéntico diseño, **sin** marca de «sin traducir» (FR-008). El número no cambia. |
| Orden estable | Fijo de punta a punta: IBAN primero, luego SINPE; dentro de cada grupo, `sort_order, id` (FR-004); igual resultado con cualquier ancho de pantalla o idioma. |
| Duplicado de concepto | La página solo contiene la información de donaciones; nada de eventos (F4), grupos (F5) ni ministerios (F6) (FR-001, SC-013). |
| `prefers-reduced-motion` | Sin revelado ni stagger; contenido íntegro al instante (FR-023/F10). |
| Borrador | Jamás visible en esta página ni en ninguna respuesta pública, por ninguna vía (FR-012, SC-002). |

**Responsivo**:

- **Móvil (< 640 px)**: grilla de 1 columna, fichas a todo ancho; el número a todo el ancho con agrupado(?) — con bloques de 4 y `overflow-wrap: anywhere` el IBAN completo se lee sin desplazamiento horizontal a 320 px (probado con el IBAN más largo del tipo CR, 22 caracteres); botón «Copiar número» a todo el ancho con ≥ 44 px de alto.
- **Tableta (≥ 640 px)**: 2 columnas de fichas; H1 y rótulos crecen con `clamp`.
- **Escritorio (≥ 1024 px)**: `max-w-6xl`; fichas con número destacado y botón a la mitad superior; los textos (intro y uso) en columna de lectura cómoda (~65–70ch), no a todo el ancho de página.

### 4.2 Panel — Inicio (cambio menor en F2)

`AccesosRapidos` añade **una** tarjeta con el patrón ícono + texto existente:

- **«Gestión de donaciones»** (con permiso `donaciones` → `/panel/donaciones`).

Una cuenta sin el permiso no ve la tarjeta (mismo criterio de F4/F6). Ni «donaciones» implica otro módulo, ni otro implica a este (Decisión 5). El área del módulo en auditoría es «donaciones» (el nombre no cambia, Q8; la registra F2 con el patrón existente; sin UI nueva).

### 4.3 Panel — Gestión de donaciones — `/panel/donaciones` (US2, US3)

**Propósito**: administrar los números y los dos textos sin depender de nadie, y ver de un vistazo qué es público hoy.

- Título «Gestión de donaciones» + banda de ayuda: «Aquí administras las cuentas y números donde la iglesia recibe ofrendas, y los textos que explican cómo donar y en qué se usan. Solo lo publicado es visible al público.» ⟲ D-7
- **Zona «Textos de la sección»** (arriba; dos bloques lado a lado en desktop, apilados en móvil):
  - **Introductorio** y **¿En qué usamos las donaciones?**: cada uno es una tarjeta con **fragmento del contenido guardado** (primeras líneas visibles), **`StatusPill`** (Borrador/Publicado, estado **propio** de cada texto, Q10) y botón **Editar** → `Dialog` con `TextoForm` (§4.5).
  - Si un texto **nunca se ha escrito** (no existe aún): el bloque muestra el estado vacío de su tarjeta: «Todavía sin escribir.» + el mismo botón Editar (no es un hueco: tiene su acción). ⟲ D-8
- **Listado de métodos agrupado por tipo** (Q1; dos bloques con `SectionHeading` interno del panel, mismo lenguaje que la sección pública):
  - **Bloque «Cuentas IBAN»** + botón **«Agregar cuenta»**.
  - **Bloque «Números de SINPE Móvil»** + botón **«Agregar número»**.
  - Cada bloque es una `Table` (desktop) / tarjetas apiladas (móvil), patrón F2/F4/F6. Columnas:
    - **Orden** — el valor numérico definido por el equipo, dentro de su grupo (US2 esc. 10).
    - **Número** — el normalizado (tal como quedó guardado; en IBAN puede mostrarse también agrupado para lectura del equipo — misma pieza `NumeroDonacion` con su botón de copiar reutilizada ⟲ D-9: al copiar desde el panel, el equipo copia el número exacto igual que lo haría quien dona).
    - **Entidad** (español si existe — es texto traducible es/en: se muestra la del idioma del panel, español).
    - **Titular** — el nombre guardado, o «—» (aquí SÍ se usa «—»: es un listado del panel, como en F6; en público nunca aparece).
    - **Moneda** — «colones»/«dólares» en IBAN; **columna no existe en el bloque SINPE** (Q2).
    - **Estado** — `StatusPill` (`Borrador`/`Publicado`), como en F3/F4/F6.
    - **Acciones** — `Editar` · `Publicar`/`Retirar de la sección` · `Eliminar`.
- **Eliminar**: `ConfirmDialog` danger: «"{tipo}": se eliminará para siempre «{entidad} {número}». Esto no se puede deshacer.» (patrón F2/F4/F6; se usa entidad+número porque el método no tiene título). Queda en auditoría (FR-017). No hay papelera.
- El listado completo (sin paginación, default) va ordenado SIEMPRE por grupo (IBAN antes, SINPE después) y `sort_order, id` dentro de cada grupo, así el equipo ve el orden que consume el público. Los dos bloques del listado usan un **rango de orden independiente por grupo** (Q1: la orden manual es «dentro de cada grupo»; el «mover una cuenta del grupo IBAN a SINPE» no es reordenar, es editar el tipo — ver §4.4).

### 4.4 Panel — Formulario de método de donación (`DonacionForm`, en `Dialog`)

Patrón F3/F4/F6 (pestañas de idioma + Zod + avisos junto al campo) con las reglas propias de F7:

- **Tabs de idioma**: **Español** (obligatorio para el texto que se rellene) / **English (opcional)** — con la ayuda de F3 («Si lo dejas vacío, el público verá este contenido en su versión en español.»). Regla **por campo**: entidad y titular son textos traducibles; **el número no se traduce** (FR-007) y **no aparece en las pestañas** (es un único campo no lingüístico). **Un campo vacío no obliga a nada** (titular opcional; US2 esc. 8).
- Pestaña **Español**:
  - **Entidad financiera** (`Field`, obligatorio; contador vivo, límite corto del contrato — el arquitecto fija el número; ayuda: «El nombre del banco; por ejemplo: Banco Nacional de Costa Rica.»).
  - **Titular de la cuenta** (`Field`, opcional; ayuda: «Solo si quieres que el público vea a nombre de quién está la cuenta.»).
- Pestaña **English (opcional)**: entidad y titular en inglés, el mismo par de campos con marca opcional (frase de ayuda de arriba). El número no aparece aquí.
- **Campos no lingüísticos (después de las pestañas, siempre visibles)**:
  - **Número de cuenta / número de teléfono** (un solo `Field`, obligatorio, etiqueta dinámica según el tipo fijo del formulario). Nada en las pestañas.
    - Tipo **Cuenta IBAN**: validación Zod de estructura y longitud (IBAN costarricense o internacional; FR-015). Ayuda: «Puedes pegarlo tal como te lo dieron, con espacios o guiones: se guardará limpio (sin espacios) y el público lo verá agrupado en bloques para leerlo sin error.» Campo con autolimpieza visual de espacios/guiones pegados (solo present dbHelper, la cadena la conserva tal cual hasta guardar).
    - Tipo **SINPE Móvil**: validación de celular costarricense de 8 dígitos, con o sin `+506` (FR-016). Ayuda: «El número de celular a donde llega la donación; acepta +506 o no — se guarda con su prefijo normalizado.»
    - **El tipo no es un campo de edición**: cambiar una cuenta IBAN por un SINPE Móvil no tiene sentido como edición in-place (el número cambia de semántica); la propuesta ⟲ D-9 es que el formulario **muestra el tipo como etiqueta fija al editar** (con ayuda: «El tipo no se cambia aquí: si necesitas el número como cuenta o como celular, crea el nuevo y elimina el otro.»); al **crear**, el tipo viene fijado por el bloque de origen (Agregar cuenta / Agregar número) y el formulario no ofrece selector. Cero errores de tipo «heredados» de la edición. (Si el humano prefiere un selector editable al editar, es una revisión de §4.4, no del negocio — default revisable del plan.)
  - **Moneda** (solo formulario de tipo IBAN; `Select`): **«Colones»** / **«Dólares»**, **obligatoria** (US2 esc. 3, FR-014; primera opción no vale como vacía). En el formulario de SINPE **el campo no existe** (Q2).
  - **Orden** (`Field` numérico, con valor 0 por defecto; ayuda: «Las cuentas (o los números) se muestran en orden: menor número primero. Con empate, el más antiguo va antes.»). Es el mecanismo de reordenamiento (Q1): **editor numérico, sin drag & drop ni flechas** — mismo precedente que la columna de `sort_order` de F3/F4/F6 (⟲ D-10).
- **Guardar**: `Button` loading («Guardando…»); el modal no se cierra hasta terminar; `Esc` no descarta lo escrito sin aviso (mismo criterio D-18 de F4, formulario largo). Al guardar con el mismo tipo/entidad/número que otro método: aviso de duplicado que impide el guardado (§2.5).

**Nota de reutilización**: `ItemFormFields` de F4 no sirve tal cual (lugar, organizador, fechas). `DonacionForm` compone directamente las piezas compartidas (`Tabs`, `Field`, `Select`, `Dialog`) del mismo patrón. Sin `ImageUploader` nunca (Q6).

### 4.5 Panel — Edición de los textos (`TextoForm`, en `Dialog`)

Un solo componente para los dos textos, parametrizado con el título del texto («texto introductorio» / «texto de uso de las donaciones»), con el patrón F6 §4.4/F4:

- **Tabs de idioma**: **Español** (obligatorio si el campo se va a guardar con algo — es el texto base del público, FR-007) / **English (opcional)** con la misma ayuda de F3. Regla **por campo** (aunque aquí el «campo» es un solo textarea por idioma).
- **Texto** (`TextArea` grande, contador vivo `{n}/2.000` — default revisable, Assumptions; saltos preservados; admite tildes, ñ, ¿ ¡ y emojis íntegros; ayuda: «Uno o varios párrafos. Aparecerá tal como lo escribas, con sus saltos.»).
- **Estado del texto**: **no se cambia en el formulario** — se cambia con `PublishControls` en su tarjeta del listado (§4.3), como todo el módulo (Q10: mismo mecanismo que un método). El `StatusPill` en la tarjeta del formulario de edición no existe: solo se muestra en la zona de listado, para que «editar» y «publicar» sean dos decisiones separadas (patrón del módulo). ⟲ D-11
- **Guardar**: `Button` loading; el éxito cierra y `Notice` en la zona de textos: «Cambios guardados: ya visibles en la sección.» (si estaba publicado) o «Cambios guardados.» (si seguía borrador).

### 4.6 Pantallas de estado vacío (módulo del panel)

- **Listado de métodos vacío por total** (grupo IBAN sin ninguno): «Todavía no hay cuentas IBAN. Agrega la primera para que el público la vea.» + `EmptyState` con botón «Agregar cuenta». Ídem SINPE («Agrega el primer número»).
- Si hay métodos pero **ninguno publicado**: aviso destacado «Nada está publicado ahora: el público no ve esta sección ni el bloque de la portada, y `/ofrendas` redirige a la portada, hasta que publiques al menos una cuenta o un número.» (FR-005 + FR-022 + Q9).
- **Textos sin escribir / ambos en borrador**: cada tarjeta muestra su estado («Todavía sin escribir.» o píldora Borrador) y su botón Editar; si **ningún texto está publicado** pero hay métodos publicados, el grupo sigue visible — solo los textos faltan en el público (coherencia de §4.1).

### 4.7 Texto largo, sin datos, muchos métodos

| Caso | Cómo se ve |
|---|---|
| Introductorio o uso con 2.000 caracteres / varios párrafos | En el panel el contador lo acota; en público se muestra completo con saltos preservados, en columna de lectura (~65–70ch), sin romperse en móvil (Poppins 18, ancho fluido). |
| Titular largo (p. ej. «Asociación Cristiana Simiente Santa, por medio de su Junta Directiva») | Se envuelve en dos renglones si hace falta (`text-wrap: balance`), conservando tildes, ñ y demás caracteres íntegros; se muestra completo, sin elipsis (FR-003: cada campo íntegro). |
| Número IBAN de longitud internacional larga | Bloques de 4 con `overflow-wrap: anywhere`: se reacomoda en 320 px **sin quitarle espacio al número completo** — la ficha crece en altura en vez de recortar el número (SC-009, SC-007). |
| Cuenta IBAN sin titular | La ficha se presenta completa sin línea de titular y sin marcador (Q2, US1 esc. 5). |
| SINPE y moneda | La ficha SINPE no dibuja línea de moneda; el formulario SINPE no la ofrece (Q2). |
| Muchos métodos publicados (10+) | La sección los lista **todos** en el orden definido (2 grupos; 1 col móvil / 2 col tablet+; sin paginación, default de la spec). El extracto de portada queda fijo en 5 (FR-022). |
| Duplicado exacto | Rechazado en el formulario (aviso; nada guarda). |
| Texto nunca escrito | En el panel: estado vacío de su tarjeta («Todavía sin escribir.») con acción Editar; en público: no se dibuja. |

### 4.8 Cross-check con F2 (área de auditoría)

- La auditoría de F2 registra cada operación con textos como «Creó/Modificó/Publicó/Retiró/Eliminó la cuenta "…"» / «el número de SINPE "…"» / «el texto introductorio» / «el texto de uso» + resultado, incluidos los intentos que no se completan (FR-017). El módulo de auditoría no cambia su UI.
- El registro es solo lectura, como en F2/F3/F4/F6 (SC-011).

### 4.9 Portada — bloque «Ofrendas y diezmos» (`/`; US6, Q7, FR-022, SC-014)

> Único contenido de F7 en la portada de F3 (junto al de ministerios, el segundo bloque añadido a la portada; Q7). Sigue **el patrón ya construido de los bloques de F4/F6** (misma gramática de `SectionHeading` + fila + botón) con una diferencia de forma necesaria: **sin imágenes** (Q6) — la ficha compacta es textual: tipo + entidad + número. Datos por **endpoint público propio** — propuesta `GET /api/v1/portada/donaciones?lang=es|en` vía `usePortadaOfrendas`: tope **5** resuelto en servidor, orden de la sección (IBAN, luego SINPE, `sort_order, id` en cada grupo), solo lo publicado, DTO mínimo, `Cache-Control: no-store` — una sola petición para el bloque; el contrato exacto lo fija el arquitecto (FR-022). El mismo resultado alimenta la **presencia en la navegación** (una sola petición paga bloque + barra + menú + pie; mismo criterio de F6).

**Propósito**: quien entra al sitio sabe **solo con la portada** cómo apoyar a la iglesia, y de ahí salta a `/ofrendas` en un paso para copiar el número exacto y leer el uso de las donaciones (SC-014).

**Ubicación (Q7, ya decidida por el humano)**: **inmediatamente debajo del bloque de ministerios de F6**. Sin reordenar ni alterar ninguna otra sección de F3.

| Estado | Qué ve el visitante (portada) |
|---|---|
| Con 5 publicados | Encabezado `SectionHeading` «Ofrendas y diezmos» + **una fila horizontal de 5 tarjetas compactas de texto** (misma rejilla D-28 de F4/F6: 1 col móvil, 2 tableta, 5 escritorio), en el orden de la sección. |
| Con menos de 5 (1–4) | Exactamente esas fichas en la fila; sin huecos ni marcadores vacíos, sin estirar (US6 esc. 3). |
| **Con 0 publicados** | **El bloque no se dibuja**: sin título, sin mensaje, sin aviso, sin reservar espacio — regla general de F3 de ocultar secciones vacías (Q7, Q9). La portada queda coherente; **diferencia deliberada** con «Lo que viene» (que sí muestra su estado vacío). |
| Carga (primera) | Esqueleto **propio y de altura acotada**: encabezado + fila de fichas fantasma **de texto** (sin zona de imagen — más bajas que las de eventos/ministerios: el alto lo da el número), mismo patrón del esqueleto de `ProximamenteBlock`; si el resultado es 0, el bloque desaparece al llegar el dato (brinco de layout aceptado: el caso 0 es raro en un módulo que se administra con contenido; mismo ⟲ D-9 de F6). |
| Error sin nada en caché | El bloque **no se dibuja** (no deja un título huérfano sin contenido), coherente con la regla general de F3 y con la decisión de F6 (D-10). Sin «Volver a intentar» en el bloque: el usuario lo abre desde `/ofrendas` o recargando. |
| Error con ítems en caché | Se muestra el último contenido válido + aviso discreto: «No pudimos actualizar; ves la última versión disponible.» (patrón F3). |
| Fallback es→en (US6 esc. 7) | Interfaz del bloque por diccionario; contenido (entidad, titular…) según FR-008 con fallback del servidor, sin marcas; **el número no cambia** (US4 esc. 5). |
| Borrador | Jamás aparece como ficha (US6 esc. 6, FR-012, SC-002). |
| `prefers-reduced-motion` | Sin `Reveal` ni transiciones: ítems visibles de inmediato. |

**Composición de una ficha compacta (`OfrendasItem`, «Card solid» de texto)** — misma gramática que `UpcomingItem`/`MinistryItem` de F4/F6, **sin imagen y sin título-enlace individual**: en el extracto, la ficha presenta datos; **la acción es la del bloque** (ir a la sección); no se copia desde la portada.

- **Rótulo de tipo**: chip pequeño fijo **«Cuenta IBAN»** o **«SINPE Móvil»** / («Bank account (IBAN)» / «SINPE Móvil number», diccionario) — necesario porque en la portada no hay rótulos de grupo y el tipo es la primera decisión que la persona debe tomar (¿a cuál transferir?). Píldora neutra con borde, sin color semántico. ⟲ D-12
- **Entidad financiera** (Poppins 600 1.0625rem, máx. **2 líneas** con elipsis): es lo que la persona reconoce («el del Banco Nacional»); **no es un enlace**: en la portada no se copia ni navega por ficha (ver regla de abajo).
- **Número**: una línea con la presentación de lectura (bloques de 4 / SINPE normalizado); si no cabe: **se envuelve a 2 líneas máx.** con `overflow-wrap: anywhere` — **nunca se recorta a la mitad con elipsis, nunca un número incompleto a la vista** (la exactitud manda sobre la compacidad; si el ancho lo exige, la ficha crece). En móvil normalmente cabe en 1 línea; si no, se usa la envoltura disponible.
- **Moneda** (solo IBAN): línea pequeña «Colones» / «Dólares» / («Colones» / «Dollars» — palabra de interfaz; ⟲ D-13).
- **Titular** (solo si existe): línea pequeña con el nombre (si no cabe: se recorta con elipsis visual y el nombre completo para lectores de pantalla; en el extracto es contextual, la ficha completa de la sección lo muestra íntegro).
- **Sin botón de copiar por ficha**: el extracto invita a la sección, donde la copia es explícita (FR-003 vive allí; en la portada, un botón de copiar sin la promesa exacta — bloques, confirmación — distraería sin llevar al aporte). ⟲ D-14

**Botón del bloque** — debajo de la fila, alineado al final (a todo el ancho en móvil): `PublicButton` variante secundaria «Ver formas de ofrendar» / («See ways to give»): lleva a `/ofrendas` (44 px de alto, en móvil ≥ 44 px de alto como todo interactivo). Único paso 2 del bloque: sin paginación, sin pestañas, sin ítems 6+ (FR-022). Solo existe con 1–5 ítems — con 0, el bloque entero no existe.

**Responsivo** — mismo D-28 de F4/F6, sin reapertura: móvil columna única a todo ancho; tableta grilla envuelta de 2 columnas; escritorio `grid-cols-5` fijo sin `overflow`; **nunca carrusel ni scroll lateral** (0 scroll a 320 px, SC-009/SC-007).

**Accesibilidad** — nivel exacto de F4 §4.10 / F6 §4.9: `section` con `aria-labelledby` al `h2` «Ofrendas y diezmos» (mismo nivel que los títulos de la portada); fichas como `li` con `h3` (rol list conservado, `<ul>`); orden de lectura = orden del DOM = orden de la sección; el chip lleva el tipo con palabra (no solo color); el número con `aria-label` completo; hit targets ≥ 44 px (el botón del bloque); imágenes: no hay (FR-020); sin `<time>` (nada vence); `prefers-reduced-motion` sin revelado. En esta ficha, **h3 no es un enlace** (a diferencia de F4/F6, donde el título-enlace es el único interactivo por ficha): aquí ninguna ficha es clicable — el único paso es el botón del bloque (mismo criterio de F6 con el botón de unirse huérfano en el extracto; ⟲ D-14).

---

## 5. Estados por vista con datos (panel)

### a. Zona de textos (`/panel/donaciones`, arriba)

| Estado | Qué ve |
|---|---|
| Cargando | `EmptyState` kind=loading («Cargando los textos…») — junto con el del listado. |
| Texto no escrito nunca | Tarjeta con «Todavía sin escribir.» (sin píldora: el estado aún no existe) + botón Editar (§4.6). ⟲ D-8 |
| Texto en borrador | Fragmento del contenido + `StatusPill` «Borrador» + botón Editar. |
| Texto publicado | Fragmento + píldora «Publicado». |
| Error de carga | «No se pudo cargar los textos.» + `Button` «Reintentar» (`Notice` error, `role="alert"`). |
| Éxito de guardado | `Notice` temporal verde «Cambios guardados.» / «Cambios guardados: ya visibles en la sección.» (según el estado del texto). |

### b. Vista general del listado (`/panel/donaciones`, abajo)

| Estado | Qué ve |
|---|---|
| Cargando | `EmptyState` kind=loading («Cargando las cuentas…») por bloque; barra de acciones («Agregar cuenta» / «Agregar número») ya visible. |
| Vacío (sin ninguno) | `EmptyState` con texto y botón de agregar por bloque (§4.6). |
| Vacío con todo borrador | Aviso destacado «Nada está publicado ahora…» (§4.6) + listado por grupos. |
| Error de carga sin nada | «No se pudo cargar las cuentas y números.» + `Button` «Reintentar» (`Notice` error, `role="alert"`). |
| Error con contenido en caché | Último listado válido + `Notice` discreto sobre la actualización (patrón F3). |
| Éxito | Dos bloques (IBAN, SINPE) con sus filas: Orden, Número, Entidad, Titular («—»), Moneda (IBAN), píldoras y acciones. |
| Sin permiso | No llega (ruta protegida); si el servidor lo rechaza en caliente: `ForbiddenPage`/mensaje de F2. |
| Éxito de operación | `Notice` temporal verde: «Cuenta creada.» / «Número de SINPE creado.» / «Cambios guardados: ya visibles en la sección.» / «Publicado en "Ofrendas y diezmos".» / «Retirado de la sección. Estará disponible como borrador para publicarlo cuando quieras.» / «Eliminado.» (§7.1). |

### c. Los formularios (`DonacionForm`, `TextoForm` en `Dialog`)

- Precargado al editar; vacío al crear (con Orden 0 y la moneda sin elegir en IBAN).
- Guardando: `Button` loading; modal sin cerrarse; `Esc` con aviso si hay cambios (D-18 de F4).
- Éxito: cierre + `Notice` en la página (con los textos de §7.1).
- Errores de campo con los textos del §7, junto al campo y resumen arriba (patrón F2), foco al primer error.
- Trampa de foco y devolución del foco a la fila que abrió el `Dialog`.

### d. Publicar / retirar / eliminar

- **Publicar / retirar**: mismas reglas de F6 (publicar directo sin confirmación; retirar con `ConfirmDialog` suave: «"{entidad}" dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.» Botones «Retirar» / «Dejar como está.»). El estado de un texto se cambia igual desde su tarjeta (Q10).
- **Eliminar**: `ConfirmDialog` danger con el texto del §4.3 (entidad+número, no hay título). Botones «Eliminar» / «Cancelar».
- **Auto-publicar al guardar** (elemento ya publicado): «Cambios guardados: ya visibles en la sección.» (FR-013; crítico para corregir un número error de cuenta el mismo día).
- **Cambio de orden**: guardar el `Dialog` de edición actualiza el orden; el `Notice` confirma el cambio y la fila aparece en su nueva posición dentro del grupo.

---

## 6. Componentes React

> Nombres en inglés (coherente con el repo), en la carpeta de feature **`donaciones`** (el módulo/permiso, Q8; no «ofrendas», que es solo la ruta pública). El listado final de tareas lo fija el arquitecto; aquí es la propuesta con responsabilidad. **Reutilizar antes de crear**: casi todo existe (verificado en `frontend/src/`).

### 6.1 Reutilizados de F2/F3/F4/F6 (sin cambios)

| Componente | Uso en F7 |
|---|---|
| `Button` / `PublicButton` | Acciones del panel; botón «Copiar número» y «Ver formas de ofrendar» en público. |
| `Field`, `TextArea` (de `features/informacion/components/`), `Select`, `Notice` | Formularios; avisos del panel y avisos públicos. |
| `Dialog` / `ConfirmDialog` | Alta/edición (método y textos); retirar y eliminar. |
| `EmptyState` | Cargando, vacío y avisos del módulo. |
| `StatusPill` | Estado borrador/publicado de métodos y de cada texto. |
| `Table` | Listado de cada bloque (desktop); tarjetas en móvil como F4/F6. |
| `Tabs` | Pestañas de idioma es/en de los dos formularios (patrón F3/F4/F6). |
| `PublishControls` | Publicar / retirar por método y por texto (textos propios del módulo con sus props). |
| `Pagination` / `DateRangeFilter` / `DatesEditor` / `MapLink` / `DateBadge` / `CategoryChip` / `ImageUploader` / `CardImage` / `HiddenNotice` | **Ninguno se usa** (sin paginación, sin fechas, sin mapa, sin catálogo, **sin imágenes por decisión del humano Q6**, y la sección oculta redirige — no usa el aviso de F4). |

### 6.2 Reutilizados de F10 (sin cambios)

| Componente | Uso |
|---|---|
| `PublicLayout`, `SiteHeader`, `PublicNav`, `MobileMenu`, `useSiteLinks` | Navegación; **única lista de enlaces** extendida con «Ofrendas y diezmos» (ruta, tras «Ministerios»). |
| `LanguageSwitcher`, `LanguageProvider`/`useLanguage`, i18n (`es.ts`/`en.ts`/`messages.ts`) | Selector y cadenas; nada nuevo de idioma. |
| `ThemeToggle`/`useTheme` | Modo claro/oscuro en las superficies nuevas. |
| `SectionHeading` | Rótulos de la sección (01 IBAN, 02 SINPE) y del bloque de portada. |
| `Card` | Superficie pública (`variant="solid"` por ficha, sin imagen). |
| `Reveal` / `RevealGroup` | Entrada de los grupos y del bloque de portada (stagger suave). |
| `SectionSkeleton` | Cargando la sección pública y el bloque. |
| `FooterSiteMap`, `BrandLogo`, `GradientBlobs`, `SocialIcon` | Pie, chrome y fondo intactos. |

### 6.3 Nuevos del sitio público (`frontend/src/features/publico/ofrendas/`)

| Componente / hook | Responsabilidad |
|---|---|
| `OfrendasPage` | Composición de H1 + subtítulo + introductorio + dos grupos + texto de uso a partir de `usePublicOfrendas`; decide carga/error/**redirección a la portada con 0 publicados** (Q9: `<Navigate to="/" replace>`, sin parpadeo — ver §4.1; **no usa `HiddenNotice`**, patrón de F6). |
| `IntroText` / `UsageText` | Párrafo(s) del texto que llega (columna de lectura, saltos preservados); **no se monta si el texto no viene** (texto en borrador o no escrito). |
| `DonacionesGroup` | Bloque rotulado (01/02): `SectionHeading` + grilla de `DonacionCard` + `RevealGroup`; no se dibuja si no hay fichas en el grupo. |
| `DonacionCard` | Ficha pública: entidad (h3), `NumeroDonacion`, moneda (solo IBAN, con la palabra de interfaz), titular solo con dato; sin imágenes, sin chip. |
| `NumeroDonacion` | Presentación exacta del número: IBAN **agrupado en bloques de 4** o SINPE en su formato normalizado (envuelto a dos líneas si hace falta), `aria-label` con el número completo para lectores, `overflow-wrap: anywhere` a 320 px, **sin elipsis nunca**. |
| `CopyAccountButton` | Botón «Copiar número» (icono + texto, ≥ 44 px): copia el número **normalizado** al portapapeles, cambia a confirmación «Número copiado.» (`role="status"`, 2 s) y ante fallo del portapapeles muestra la vía de transcripción; el botón es el único interactivo de la ficha. |
| `OfrendasBlock` (portada) | Bloque «Ofrendas y diezmos» de la portada: `SectionHeading` + fila/grilla de hasta 5 `OfrendasItem` a partir de `usePortadaOfrendas` (endpoint propio, tope 5 en servidor); **se oculta por completo sin ítems** (nada de mensaje vacío — como `MinisteriosBlock`, a diferencia de `ProximamenteBlock`) — devuelve `null` en éxito sin ítems y en error sin caché; esqueleto de fichas de texto mientras carga; aviso discreto con caché y error. |
| `OfrendasItem` (portada) | Variante compacta de texto: chip de tipo («Cuenta IBAN» / «SINPE Móvil»), entidad (h3, no enlace), número completo sin elipsis (mayormente envuelta a 2 líneas si hace falta), moneda (IBAN) y titular si existe; sin botón de copiar. |
| `usePublicOfrendas` | TanStack Query sobre el endpoint público del módulo (contrato del arquitecto; propuesta `GET /api/v1/ofrendas?lang=es|en`, DTO con textos + grupos + orden listos): alimenta **solo la página** `/ofrendas` y decide la redirección con 0 publicados (Q9). **No** alimenta la presencia en `useSiteLinks`. |
| `usePortadaOfrendas` | TanStack Query para el bloque de la portada: propuesta `GET /api/v1/portada/donaciones?lang=es|en` (FR-022), DTO mínimo, tope 5 en servidor, `no-store`; alimentada por `OfrendasBlock` y compartida por `queryClient` con `useSiteLinks` para deducir la presencia de la navegación (una sola petición paga bloque + barra + menú + pie; criterio F6). Son **dos consultas**: nunca consume la consulta de `/ofrendas`. |

### 6.4 Nuevos del panel (`frontend/src/features/donaciones/`)

| Componente / hook | Responsabilidad |
|---|---|
| `pages/DonacionesPage.tsx` | Panel de gestión: título, banda de ayuda, zona de textos + listado por grupos + estados (§4.3, §5). |
| `components/DonacionesListsByType.tsx` | Los dos bloques de listado con `SectionHeading` de panel, botón de agregar por bloque y `Table`/tarjetas móviles. |
| `components/DonacionForm.tsx` (`Dialog`) | Alta/edición de un método (§4.4) con el patrón de pestañas + Zod (validación en `validation.ts`, mirroring F6). |
| `components/TextoForm.tsx` (`Dialog`) | Edición del texto introductorio o del texto de uso (§4.5): pestañas + textarea con contador (el estado se cambia fuera, en su tarjeta). |
| `components/NumeroDisplay.tsx` | Presentación/copia del número en el panel (reutilizada del concepto de `NumeroDonacion`, apoyo de la columna Número del listado). |
| `lib/numero.ts` | Utilidades de presentación: agrupado en bloques de 4 del IBAN, formato de lectura del SINPE; **la normalización/copia es del dato que llega del servidor** (no se re-normaliza en pantalla). |
| `validation.ts` | Zod: IBAN por estructura/longitud (con espacios/guiones tolerados al escribir → normalizado al guardar es tarea del backend), celular costarricense de 8 dígitos con o sin `+506`, moneda obligatoria solo en IBAN, español por campo de texto, límites de extensión. |
| `messages.ts` | Textos del módulo en español (FR-018) — catálogo del §7.1. |
| Hooks TanStack Query | `useDonaciones`, `useCreateDonacion`, `useUpdateDonacion`, `useDeleteDonacion`, `useTextoIntroductorio`/`useUpdateTextoIntroductorio`, `useTextoUso`/`useUpdateTextoUso` (o un par único parametrizado por tipo de texto; lo fija el arquitecto). Ningún `fetch` en componentes (regla F2). |

**Pruebas esperadas** (Vitest + Testing Library + MSW): composición pública con casos 0/1/5/>5 publicados por grupo y por total, cada caso «sin dato» (sin titular → 0 líneas; sin moneda en SINPE → 0 líneas; textos en borrador → 0 huecos), número íntegro copiable (el valor del portapapeles = número normalizado, no la presentación con bloques), agrupación IBAN/SINPE y orden `sort_order, id` dentro de cada grupo, **La vista con 0 publicados: redirección a la portada sin parpadeo** (con la consulta en caché, redirige en el primer render sin pintar H1 ni fichas; sin caché, solo el esqueleto mientras se decide; y `/eventos` conserva su `HiddenNotice` — comportamientos deliberadamente distintos, Q9), bloque de portada con tope 5 resuelto en servidor (1…5, **0 ítems → bloque ausente**, sin mensaje ni título) y oculto en 0 líneas, fallback de contenido sin inglés (por campo) y **número/tipo/entidad/moneda idénticos en ambos idiomas**, validaciones del formulario (IBAN estructura/longitud, con o sin espacios/guiones→normalizado; SINPE 8 dígitos ± `+506`; moneda obligatoria solo en IBAN; español base por campo; límites; duplicado), ciclo publicar/retirar/eliminar con auditoría y textos también publicar/retirar (Q10), permisos cruzados con «donaciones», 320 px sin scroll y con IBAN legible completo, y copia accesible por teclado/lector.

---

## 7. Catálogo de textos

### 7.1 Panel (español; FR-018) — añadidos a los de F2/F3/F4/F6

| Situación | Texto |
|---|---|
| Banda de ayuda | «Aquí administras las cuentas y números donde la iglesia recibe ofrendas, y los textos que explican cómo donar y en qué se usan. Solo lo publicado es visible al público.» |
| Crear cuenta | «Cuenta creada.» + primera vez: «Quedó en borrador: no la verá el público hasta que la publiques.» |
| Crear SINPE | «Número de SINPE creado.» + primera vez: «Quedó en borrador: no lo verá el público hasta que lo publiques.» |
| Guardar con auto-publicación | «Cambios guardados: ya visibles en la sección.» |
| Guardar (borrador) | «Cambios guardados.» |
| Publicar | «Publicado en "Ofrendas y diezmos".» |
| Retirar | «Retirado de la sección. Estará disponible como borrador para publicarlo cuando quieras.» |
| Retirar (confirmación) | «{entidad} dejará de ser visible al público hasta que lo publiques de nuevo. Sus datos se conservan.» |
| Eliminar (confirmación) | «{tipo} {entidad} {número} se eliminará para siempre. Esto no se puede deshacer.» |
| Eliminado | «Eliminado.» |
| Entidad vacía | «Escribe el nombre del banco.» |
| Número vacío | «Escribe el número de la cuenta (o el número de celular para SINPE Móvil).» |
| IBAN no válido | «Ese número no parece una cuenta IBAN válida. Revisa la cantidad de letras y dígitos: vale tanto el IBAN de Costa Rica como uno internacional.» |
| SINPE no válido | «Ese número no parece un celular de Costa Rica: son 8 dígitos, con o sin el prefijo +506.» |
| Moneda sin elegir (IBAN) | «Elige la moneda de la cuenta: colones o dólares.» |
| Inglés sin español (por campo) | «El contenido en español es obligatorio: es el idioma base del sitio. Falta la versión en español de {campo}.» |
| Límite de los textos | «{n} de 2.000 caracteres. Recorta {m} para poder guardar.» |
| Duplicado | «Ya existe esa misma cuenta o número (mismo tipo, entidad y número). Edita el que ya tienes o cambia algún dato.» |
| Sección sin publicados | «Nada está publicado ahora: el público no ve esta sección ni el bloque de la portada, y `/ofrendas` redirige a la portada, hasta que publiques al menos una cuenta o un número.» |
| Error del sistema al guardar | «No se pudo completar la operación. Tus escritos están a salvo; vuelve a intentarlo en unos minutos.» |
| Sin permiso (servidor) | «No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol.» (F2) |
| Ayuda del orden | «Las cuentas (o los números, dentro de su grupo) se muestran en orden: menor número primero. Con empate, el más antiguo va antes.» |
| Ayuda del número (IBAN) | «Puedes pegarlo tal como te lo dieron, con espacios o guiones: se guardará limpio y el público lo verá agrupado en bloques.» |
| Ayuda del número (SINPE) | «El número de celular a donde llega la donación; acepta +506 o no — se guarda con su prefijo normalizado.» |
| Ayuda del titular | «Opcional. Solo si quieres que el público vea a nombre de quién está la cuenta.» |
| Ayuda de la entidad | «El nombre del banco; por ejemplo: Banco Nacional de Costa Rica.» |
| Ayuda del texto (ambos) | «Uno o varios párrafos. Aparecerá tal como lo escribas, con sus saltos.» |
| Ayuda «English (opcional)» | «Si lo dejas vacío, el público verá este contenido en su versión en español.» |
| Texto sin escribir | «Todavía sin escribir.» |

### 7.2 Sitio público — cadenas de interfaz es/en

> El contenido (textos, titular, entidad… con inglés si el equipo lo llenó) viene del panel con fallback del servidor (FR-008). Estas son las palabras de la interfaz.

| Clave | Español | English |
|---|---|---|
| `nav.sections.offerings` | Ofrendas y diezmos | Offerings & tithes |
| `offerings.page.title` | Ofrendas y diezmos | Offerings & tithes |
| `offerings.page.subtitle` | Las formas de apoyar económicamente a nuestra iglesia. | The ways to support our church financially. |
| `offerings.iban.heading` | Cuentas IBAN | Bank accounts (IBAN) |
| `offerings.sinpe.heading` | Números de SINPE Móvil | SINPE Móvil numbers |
| `offerings.holder.label` | Titular | Account holder |
| `offerings.currency.label` | Moneda | Currency |
| `offerings.currency.colones` | colones | colones (CRC) |
| `offerings.currency.dolares` | dólares | dollars (USD) |
| `offerings.copy` | Copiar número | Copy number |
| `offerings.copied` | Número copiado. | Number copied. |
| `offerings.copy.hint` | Puedes copiarlo en un toque o escribirlo tal cual, bloque por bloque. | You can copy it with one tap or type it out block by block. |
| `offerings.usage.heading` | ¿En qué usamos las donaciones? | How we use the offerings |
| `offerings.loading` | Cargando… | Loading… |
| `offerings.error.load` | No pudimos cargar la información de ofrendas. Revisa tu conexión y vuelve a intentarlo. | We couldn't load the offering info. Check your connection and try again. |
| `offerings.error.retry` | Volver a intentar | Try again |
| `offerings.error.partial` | No pudimos actualizar; ves la última versión disponible. | We couldn't refresh; you're seeing the latest available version. |
| `a11y.offerings.account` | Cuenta IBAN: {entidad}, número {número} | Bank account: {entity}, number {number} |
| `a11y.offerings.sinpe` | SINPE Móvil: {entidad}, número {número} | SINPE Móvil: {entity}, number {number} |
| `home.offerings.title` | Ofrendas y diezmos | Offerings & tithes |
| `home.offerings.subtitle` | Las formas de apoyar a la iglesia | The ways to support the church |
| `home.offerings.seeAll` | Ver formas de ofrendar | See ways to give |
| `home.offerings.badge.iban` | Cuenta IBAN | Bank account (IBAN) |
| `home.offerings.badge.sinpe` | SINPE Móvil | SINPE Móvil number |
| `a11y.home.offerings.heading` | Ofrendas y diezmos: mira el resumen y entra a la sección completa | Offerings & tithes: quick look and link to the full section |

> La redirección de la sección oculta (Q9) **no añade ninguna cadena**: no hay aviso que mostrar; si la consulta tarda, la única palabra en pantalla es `offerings.loading` («Cargando…» / «Loading…») sobre el esqueleto.
> La moneda se muestra con **palabra de interfaz + dato idéntico entre idiomas** (US4 esc. 5: la moneda no cambia como dato; su palabra de etiqueta sí se traduce — ⟲ D-6).

---

## 8. Accesibilidad (WCAG 2.1 AA) e internacionalización

### 8.1 Accesibilidad

- **Teclado**: skip link → logo → navegación (enlaces de barra; menú móvil disclosure de F10) → idioma → tema → contenido → pie. El botón de copiar, los títulos y el botón del bloque reciben el anillo de foco de F10; nada depende de un gesto. En el panel, `Esc` respeta D-18 (sin descartar cambios sin aviso).
- **Foco visible** (anillo 2 px por tema) en todo interactivo, incluidos «Copiar número» y el botón del bloque de portada; un elemento aún no revelado que recibe foco se hace visible de inmediato (guarda de `Reveal`, F10).
- **Semántica**: un `<h1>` por página («Ofrendas y diezmos» en la sección; el bloque de portada usa `h2` con `SectionHeading`, al nivel de los títulos de F3/F4/F6); cada ficha es un `article`/`li` con `h3`; el listado público es `<ul>` con orden de lectura = orden manual; los **grupos** rotulados con `SectionHeading` (01/02) y, si la estructura lo pide, `role` de grupo con nombre accesible; tablas solo en el panel. **Sin** `<time>` en F7 (nada vence). El número de cuenta es texto normal (no campo de formulario en el sitio público, salvo la copia).
- **No solo color**: lo que distingue una cuenta IBAN de un SINPE en la portada es el **bloque** en la sección y el **chip con palabra** en el extracto; los estados del panel llevan siempre texto (`StatusPill` con palabra).
- **Sin imágenes** (Q6/FR-020): 0 `img` en F7 — también se elimina esa clase de problema de `alt`/`onError`.
- **Copia accesible**: el botón es un `button` real con nombre accesible completo («Copiar número de {entidad}»); la confirmación de copiado se anuncia con `role="status"` (texto, no solo cambio de icono); el número presentado tiene `aria-label` con el número **completo y sin separadores de bloques** para una lectura corrida; la ayuda de transcripción es texto visible, no `title`.
- **Contraste y letra**: cuerpo 18 px; Poppins 600 para el número con contraste AA en los dos temas (claro y oscuro — SC-008); Bebas de rótulos igual; el turquesa como texto pequeño siempre en su clase de contraste del sistema.
- **Zoom/200 % y 320 px**: diseño fluido en `rem` con `clamp`; 0 desplazamiento horizontal (SC-007); el número se reacomoda por bloques sin perder caracteres; el extracto de la portada pasa de fila de 5 → grilla de 2 → columna única por breakpoint (D-28 de F4/F6).
- **Objetivos táctiles ≥ 44 px**: «Copiar número» (≥ 44 px de alto, a todo el ancho en móvil), botón del bloque de portada, ítems del menú móvil; la ficha no es un enlace en bloque.
- **Avisos**: cambio de idioma con `role="status"` de F10; confirmación de copia con `role="status"`; en el panel, `Notice` con su rol correcto.
- **`prefers-reduced-motion`**: sin revelado ni stagger; contenido íntegro al instante (F10 incondicional; F7 no añade movimiento nuevo — el feedback de copia es cambio de texto, no animación compleja).
- **Idioma del documento**: `html lang` por F3/F10; el contenido en fallback dentro de una página en inglés va con `lang="es"` en su envoltorio (práctica heredada de F3/F4/F6).

### 8.2 Internacionalización (Decisión 6 y Decisión 8)

- **Interfaz** (siempre bilingüe): diccionario `es.ts`/`en.ts` de F3; `LanguageProvider` y `localStorage` intactos; primera visita en español, sin auto-detección (FR-009). El panel **no** cambia de idioma (FR-018).
- **Contenido** (por campo): los dos textos, la entidad y el titular son textos traducibles con español base si el campo se rellena; inglés opcional; el fallback es→es lo resuelve el servidor (FR-008); el frontend nunca muestra un campo vacío ni traducción automática. **Un campo vacío no obliga a nada** (Q2, US2 esc. 8).
- **El número, el tipo, la entidad y la moneda no cambian entre idiomas como dato** (US4 esc. 5, FR-007): en concreto, número y tipo son un único valor; la entidad puede llevar versión en inglés (es texto del equipo) — si no la lleva, se muestra en español; la moneda es el mismo dato en los dos idiomas y su **palabra** viene del diccionario (estándar de §7.2). ⟲ D-6
- Los dos textos se muestran tal como el equipo los escribió: sin `Intl`, sin recortar párrafos, sin transformar tildes, ñ, ¿ ¡ ni emojis.

### 8.3 Exactitud del número — cómo se verifica y quién la garantiza (FR-003, SC-009)

- **La garantía es del servidor**: validación de estructura/longitud (IBAN) y de celular de 8 dígitos (SINPE) al guardar; el dato guardado ya viene normalizado; el DTO público trae el número y (si el diseño pide) ninguna versión «para mostrar» separada — la agrupación en bloques es **presentación del frontend** (`lib/numero.ts`), nunca un segundo valor que el backend intente mantener «híbrido» (dos verdades).
- El botón de copia transmite el valor tal como lo provee la API (número normalizado); si un día cambia el DTO, la copia y la lectura siguen atadas al mismo campo: **cero riesgo de que «se copie» algo distinto de lo que se lee por bloques** — la prueba de QA lo verifica explícitamente (cadenas copiadas = cadenas legibles sin espacios).
- **Criterio QA**: en las pruebas, el copiar por botón y la lectura a mano de los bloques producen **el mismo valor**, 0 cortes, 0 alteraciones (SC-009), y ningún borrador se filtra (SC-002).

---

## 9. Decisiones de UX para revisar por el humano (resumen)

| # | Decisión | Alternativa |
|---|---|---|
| D-1 ⟲ | «Ofrendas y diezmos» es una **ruta real** al final de la única lista de enlaces de F10, después de «Ministerios» (mismo patrón aprobado en F4 D-1 y F6 D-1). | Entrada entre anclas de portada (mezclaría el orden de lectura) o entrada dentro de otros módulos (no: una sola navegación). |
| D-2 ⟲ | Subtítulo de `/ofrendas` como **cadena de interfaz fija** (no campo del panel) — «Las formas de apoyar económicamente a nuestra iglesia.» | Que el equipo lo escriba en el panel (más contenido que administrar; el texto introductorio ya cumple ese rol). |
| D-3 ⟲ | El texto de uso lleva un **rótulo de interfaz fija** («¿En qué usamos las donaciones?») con el contenido del equipo debajo; el introductorio no lleva rótulo propio (arranca la lectura). | Sin rotulo (la persona adivina el borde) o mismo rótulo para los dos textos (poco claros). |
| D-4 ⟲ | El SINPE en público se presenta **con su prefijo y agrupado para lectura** (p. ej. «+506 8888 8888»); la cadena exacta normalizada la fija el plan (Assumptions de la spec). | Mostrarlo sin agrupación (un solo ladrillo de 8 dígitos es más difícil de transcribir) o inventar una presentación no definida (riesgo de que difiera lo leído de lo copiado). |
| D-5 ⟲ | **Botón «Copiar número» en cada ficha** de la sección (icono + texto, ≥ 44 px) con confirmación `role="status"` y **ayuda visible de transcripción por bloques** (US1 esc. 2 y US5 esc. 3 mencionan ambas vías; el FR-003 del humano habilita el «si la fase de UX la ofrece» — la ofrecemos). La razón: la persona sin app bancaria a mano también transcribe; el botón de copia le evita el error. | Sin botón (solo transcripción: FR-003 lo permite, pero pierde el «sin error» de gente con menos destreza) o botón solo de icono (menos comprensible). |
| D-6 ⟲ | La moneda se muestra como **etiqueta de interfaz + dato** («Moneda: colones» / «Currency: colones (CRC)»): el dato es único entre idiomas, la palabra sí se traduce. | Icono de moneda a solas (menos legible, SC-012) o símbolos no traducibles «$ / ₡» (a veces ambiguos). |
| D-7 ⟲ | Banda de ayuda del panel (§4.3 y §7.1): menciona los dos textos de la sección + la regla de «solo lo publicado es visible». | Banda más corta (menos completa) u omitir la mención de los textos de la sección (la persona la descubriría tarde). |
| D-8 ⟲ | Zona de textos con tarjetas siempre presentes: texto nunca escrito → «Todavía sin escribir.» con acción Editar (en vez de ocultar la tarjeta). | Ocultar la tarjeta (la zona de textos «aparece» desde el panel cuando se escribe el primero; menos claro para quién empieza de cero). |
| D-9 ⟲ | El **tipo no es editable** al editar un método (etiqueta fija; para un tipo distinto, crear nuevo y eliminar el anterior). Todas las operaciones van a auditoría igual. | Selector editable (permitiría «cambiar» la semántica del número con solo dos clics y sin pista de auditoría del porqué — riesgoso con números de cuenta). |
| D-10 ⟲ | Reordenamiento por **campo numérico «Orden»** en el formulario + columna Orden en el listado, **dentro de cada grupo** (precedente `sort_order` de F3/F4/F6; sin drag & drop). | Editor de flechas en la fila (requiere un patrón no existente) o drag & drop (innecesario para listas cortas; problema en táctil). |
| D-11 ⟲ | El estado de un **texto** se cambia con `PublishControls` **en su tarjeta** (como los métodos), no en su formulario de edición: «editar» y «publicar» quedan como decisiones separadas (mismo mecanismo para todo el módulo). | Interruptor de estado dentro del formulario (dos vías para lo mismo; más fácil por accidente). |
| D-12 ⟲ | Chip de tipo («Cuenta IBAN» / «SINPE Móvil») en **cada ficha del extracto de portada** (a diferencia de la sección, que lo nombra con el rótulo del grupo): en la portada no hay grupos y el tipo es la primera pregunta de la persona. | Sin chip (la ficha es ambigua en la portada) o elementos agrupados en dos mini-bloques (rompe la fila de 5 del patrón, SC-014). |
| D-13 ⟲ | En el extracto, moneda y titular van como **líneas aparte**, aceptando envoltura: **nada se recorta con elipsis salvo titular muy largo** (y con su texto completo para lectores); el número **nunca** se recorta. | Comprimir más la ficha al precio de elipses en el número (inaceptable: rompe SC-009) o caer en fichas muy altas (aceptado: la exactitud manda). |
| D-14 ⟲ | **Sin botón de copiar ni enlace por ficha en el extracto de portada** (la ficha compacta no es ni enlace ni botón; la acción única es el botón del bloque): la copia exacta y el uso del aporte viven en `/ofrendas`. | Botón de copiar por ficha en la portada (duplicaria la acción sin la confirmación/ayuda de la sección) o ficha-enlace (un «objeto enlace» con texto interno, peor para lectores de pantalla). |
| D-15 ⟲ | **Posición del `Reveal`**: la dirección de entrada del bloque F7 en la portada se fija en la implementación para no romper la alternancia con el bloque de ministerios (decisión menor). | Fijo por definición (sin flexibilidad si el bloque de F6 cambia). |
| D-16 ⟲ | El **listado del panel usa «—»** para titular vacío (y SINPE no lleva columna de moneda), aunque en público esos casos no dejan huella: es listado interno, como F6. | «—» también en público (prohibido por la spec: US1 esc. 5). |

---

## 10. Notas para el `dev-frontend` / `qa-tester`

- **El bloque de portada usa su propio hook/endpoint** (`usePortadaOfrendas` → propuesta `GET /api/v1/portada/donaciones?lang=es|en`): tope 5 en servidor, orden de la sección, DTO mínimo, `no-store`. Del cliente solo llegan la presentación, el esqueleto y el aviso parcial; el tope y el orden son lógica del servidor. Pruebas contadas: 1, 2, 3, 4, 5 y **0 ítems** (bloque ausente, sin mensaje ni título — SC-014/SC-010).
- **La presencia en la navegación sale del bloque, no de `/ofrendas`**: `usePortadaOfrendas` alimenta `OfrendasBlock` y, compartida por `queryClient`, también la entrada en `useSiteLinks` (una sola petición paga bloque + barra + menú + pie; criterio F6). Con 0 publicados **ninguna superficie enlaza la sección** y la ruta directa **redirige a la portada**: la decisión de redirigir la toma la página con su propia consulta, `usePublicOfrendas` (FR-005, FR-006, SC-010, Q9).
- **Q9 (redirección, crítica)**: `/ofrendas` con 0 publicados → redirección limpia a `/` con `replace`: **sin parpadeo** — nunca se pinta `HiddenNotice`, ni H1, ni grilla vacía, ni mensaje de error (con la consulta ya en caché, redirige en el primer render; sin caché, solo el esqueleto estándar mientras se decide); el historial queda limpio (no queda `/ofrendas` tras la portada en *Volver*); y QA confirma que **`/eventos` conserva su `HiddenNotice`** y que `/ministerios` también redirige (patrones deliberadamente distintos).
- **D-13/D-14/D-12 (extracto de portada, crítico)**: ficha de texto **sin** imagen, **sin** enlace por ficha y **sin** botón de copia — solo chip de tipo + entidad + número + (moneda IBAN) + (titular); la acción única es «Ver formas de ofrendar». Pruebas de a11y con teclado: el bloque aporta **1** Tab por el botón del bloque (no 5 Tabs por ficha — no hay interactivos por ficha).
- **D-5 (copia, crítica)**: el portapapeles recibe el **número normalizado** (lo que la API trae, sin espacios de presentación). Prueba: representación con bloques ≠ cadena copiada, y **la lectura de bloques sin espacios reproduzca exactamente la cadena copiada** (SC-009). Fallback del portapapeles → nunca un silencio: la confirmación es texto (`role="status"`).
- **D-9 (tipo no editable)**: al editar, el tipo es etiqueta fija con ayuda; las pruebas verifican que el formulario no ofrece selector y que el guardado conserva el tipo.
- **D-8 (texto sin escribir)**: la zona de textos nunca se vacía por completo; la tarjeta con «Todavía sin escribir.» es un estado válido, no un error.
- **SC-009 (número exacto, crítico)**: QA verifica copia por botón y lectura a mano de bloques (mismo valor), **0 cortes con elipsis** en ningún caso (se prueba con un IBAN internacional largo) — la ficha crece en vez de tomar atajos —, y 0 números mal formados renderizados (la validación es del servidor; el frontend muestra tal cual lo que llega).
- **SC-013 (concepto)**: la sección y el bloque no contienen nada del modelo de F4/F5/F6 (FR-001): ni fechas, ni lugares, ni encargados de ministerio; solo tipos + números + entidad + titular + moneda + los dos textos.
- **El diseño no cambia el sistema de F10**: `Card`, `SectionHeading`, `Reveal`, `useSiteLinks`, i18n, temas y tokens tal cual; 0 tokens nuevos; la página nueva es «otra sección» y no una marca propia.
- **Datos del panel con Zod** espejo del contrato; NUNCA `fetch` en componentes (hooks de `features/donaciones/`); tipos desde `backend/api/openapi.yaml` (sin `any`).
- **Q6 (imágenes)**: el módulo y las superficies públicas **jamás** renderizan `<img>` ni ofrecen `ImageUploader`; prueba de presencia: 0 elementos `<img>` en F7 por HTML y por DOM.
- **Textos SIEMPRE del diccionario en el sitio público** (`features/publico/i18n`), nunca JSX suelto (SC-006).
- **SC-008** se verifica con checklist manual de QA en ambos temas (claro/oscuro), teclado, lector de pantalla, 320 px y zoom 200 % — mismas bases de F3/F10/F4/F6.
- **El panel no se rediseña** (F10B): componentes de F2/F3/F4/F6 tal cual; el acceso de InicioPage en `AccesosRapidos` con `hasPermission('donaciones')`.

---

## 11. Trazabilidad — qué pantalla cubre qué US/FR

| US / FR | Pantalla / componente | § |
|---|---|---|
| US1 · FR-002 · FR-003 · FR-004 · FR-021 | `/ofrendas` (OfrendasPage, IntroText, DonacionesGroup, DonacionCard, NumeroDonacion, CopyAccountButton, UsageText) — textos + dos grupos + números exactos copiables/transcribibles | §4.1, §2.1 |
| US1 esc. 7 · FR-005 · FR-006 · Q9 (sección oculta) | Redirección silenciosa a la portada + navegación sin huérfanos (barra, menú, pie) | §4.1, §3, §2.7 |
| US2 · FR-010 · FR-011 («donaciones») · FR-014 · FR-015 · FR-016 · FR-018 · FR-019 | `/panel/donaciones` (DonacionesPage, DonacionesListsByType, DonacionForm, TextoForm), InicioPage accesos, orden numérico por grupo | §4.2, §4.3, §4.4, §4.5, §5, §2.3 |
| US3 · FR-012 · FR-013 · Q10 (granularidad) · FR-017 | PublishControls por método y por texto, `StatusPill` en las dos zonas, ConfirmDialog, retirar/eliminar + auditoría | §2.4, §4.3, §4.5, §4.8, §5.d |
| US4 · FR-007 · FR-008 · FR-009 | LanguageSwitcher (F10), content fallback por campo, número/tipo/entidad/moneda no traducibles, moneda con palabra de interfaz | §2.2, §7.2, §8.2 |
| US5 · FR-024 · FR-025 · FR-023 (sistema F10) | Sistema F10 completo en sección y bloque: teclado/lector, 320 px sin scroll, 200 %, contraste claro/oscuro, reduce-motion, 18 px y 44 px, copia accesible | §4.1, §4.9, §8.1 |
| **US6 · FR-022 · SC-014 (Q7)** | Bloque «Ofrendas y diezmos» en la portada (OfrendasBlock con OfrendasItem vía usePortadaOfrendas, tope 5 en servidor, **debajo del bloque de ministerios de F6**), oculto por completo sin publicados, botón a `/ofrendas` | §2.7, §4.9 |
| Q6 (sin imágenes) | 0 imágenes en toda F7 (sin `ImageUploader`, sin `CardImage`, sin logos, sin íconos de catálogo) | §0.2, §1, §4.1, §6.1, §10 |
| Auditoría (FR-017) | Módulo Auditoría de F2 intacto; no añade UI nueva; registra crear/editar/publicar/retirar/eliminar + intentos | §4.8 |

---

## 12. Dudas para el humano (no inventadas; fuera de las Q1–Q11 ya resueltas)

La spec es completa para diseñar; el diseño no abre ningún punto de negocio. Están destacados, para awareness de puertas 2:

1. **Aparición del SIMPE / palabra Colones nivel (FR-007 esc. 5)**: en la ficha propuesta, la moneda se muestra como palabra de interfaz con el dato único («Moneda: colones» / «Currency: colones (CRC)»). El FR-007 dice que «la moneda no es contenido traducible» — la propuesta respeta el **dato** único. Si el humano prefiere que también la palabra no se traduzca alguna vez, es un cambia de §7.2, no del negocio. D-6.
2. **Presentación del extracto de portada (entidad + número, D-14)**: la propuesta usa la **entidad como título de la ficha** (no es un enlace): la ficha presenta el tipo, la entidad y el número de lectura, y **la única acción del bloque es ir a `/ofrendas`** (donde están la copia exacta y los textos). Si el humano prefiriera poder copiar el número directamente desde la portada, es una revisión de D-14 (añadiría el botón de copia a `OfrendasItem`), no un cambio de negocio.
3. **El texto introductorio nunca escrito**: en público, si el introductorio está en borrador o sin escribir, la sección arranca directo en el grupo IBAN (sin rótulo de introductorio). Es lo que dicta «sin datos, sin huella» — no se pide nada al humano para ello; lo consigno aquí para que se note que la primera lectura de la sección podría no tener palabras del equipo si solo se publican cuentas.
4. **Separador del IBAN agrupado**: se propuso el agrupado estándar en bloques de 4 separados por espacios (FR-003 exige exactitud, no una segmentación concreta). Si el humano quiere otro separador (p. ej. guiones), es un cambio de presentación, no de datos.
5. **Validación en caliente del IBAN/SINPE mientras se escribe en el panel**: la propuesta valida al guardar (patrón F2/F4/F6). Validar en vivo en el `Field` es una mejora posible; la spec (FR-014) solo exige indicar qué corregir sin guardar nada a medias, y ambas la cumplen.

Fuera de esto, no hay decisiones abiertas que requieran la puerta de UX: todo lo demás aplica el sistema aprobado y los patrones probados de F4/F6.
