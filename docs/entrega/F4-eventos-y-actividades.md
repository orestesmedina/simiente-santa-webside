# Entrega F4 — Eventos y actividades

**Fecha:** 2026-10-11 · **Actualizada:** 2026-10-12 · **Versión:** 0.5.0 · **Rama:** `004-eventos-actividades`

Nota de entrega en lenguaje sencillo. Si querés la parte técnica, está en el [`README`](../../README.md), en el [`CHANGELOG`](../../CHANGELOG.md) y en la [spec de la funcionalidad](../../specs/004-eventos-actividades/spec.md).

## Qué se entrega

Ya está la sección que resuelve el problema principal: **que nadie se quede sin enterarse de lo que viene**. El sitio público muestra los próximos eventos y actividades de la iglesia, y el equipo los administra desde el panel, sin depender de nadie más.

- **El bloque «Lo que viene» en la portada**, colocado **justo después de la imagen de bienvenida y antes de «Quiénes somos»** (así lo decidiste el 2026-10-12: lo que viene es lo primero que se lee junto a la identidad): hasta **5 elementos** próximos —eventos y actividades juntos, los más cercanos por fecha— en una fila de tarjetas pequeñas con su imagen, su título, su fecha y su lugar, cada uno marcado como evento o actividad. Se prefieren los de los **próximos 2 meses** y, si en ese plazo no hay nada, se muestran los siguientes próximos. Con el botón **«Ver más eventos»** se llega a la sección completa en un solo paso. Si todavía no hay nada publicado, el bloque se queda visible con el mensaje **«Por el momento no hay eventos ni actividades»** (es la única sección que no se oculta cuando está vacía).
- **La sección «Eventos y actividades»** (en <http://localhost:5173/eventos>), con dos bloques bien separados:
  - **Próximos eventos**: los encuentros periódicos de la iglesia (noche de aposentos, vigilia…) con sus fechas concretas y, si quieren, un texto de frecuencia («todos los viernes»).
  - **Próximas actividades**: las acciones puntuales (repartir comida, predicar en tal lugar…) con su fecha y su horario.
  - Cada uno se ve con **título, fecha, hora, lugar y detalle** (más una imagen si se cargó), ordenados de lo más próximo a lo más lejano.
- **En español y en inglés**, igual que la portada: el selector de la cabecera cambia todo de inmediato, lo que no traduzcan se muestra en español (nunca un hueco) y las fechas y horas se adaptan al idioma elegido.
- **Se retira solo al pasar la fecha**: cuando un evento o actividad ya pasó, deja de aparecer en la sección pública automáticamente —sin que nadie tenga que tocar nada— y en el panel sigue ahí, marcado como pasado. Lo que está en borrador nunca se ve afuera.
- **Aparece solo cuando hay contenido**: mientras no haya nada publicado y próximo, la entrada «Eventos y actividades» no se muestra en el menú ni en el pie, y si alguien abre la dirección directa ve un aviso amable en lugar de una página vacía o de ser redirigido.
- **Pensada para todos**: nace con el sistema visual de F10 (tipografías, colores, modo claro/oscuro y animaciones que respetan «reducir movimiento»), se ve bien en teléfono, tableta y computadora, y se usa con dedo, mouse y teclado.
- **El panel para administrarla**, con **tres secciones nuevas y tres permisos independientes**: **Eventos** (`/panel/eventos`), **Actividades** (`/panel/actividades`) y **Categorías** (`/panel/categorias`). Cada contenido tiene estado **Borrador/Publicado** (se publica al guardar, como en la portada), se le puede subir imagen y se puede eliminar con confirmación.
- **Un permiso nuevo, «Categorías de agenda»**: el catálogo de permisos de F2 pasó de 9 a 10. Quien administra categorías **no** por eso administra eventos ni actividades: los tres permisos son independientes y el sistema lo comprueba en el servidor en cada operación.
- **Todo queda registrado**: cada alta, edición, publicación, retirada, eliminación y denegación aparece en `/panel/auditoria` con un texto legible («Creó el evento "…"», «Editó la categoría "…"»), igual que en F2 y F3.

## Cómo se usa (para el equipo de la iglesia)

### Visitante (sin cuenta)

1. **Desde la portada** (<http://localhost:5173/>): el bloque **«Lo que viene»**, justo después de la imagen de bienvenida y antes de «Quiénes somos», ya muestra los próximos 5 eventos y actividades con su imagen; tocando cualquiera (o el botón **«Ver más eventos»**) se llega a la sección completa.
2. Abrir <http://localhost:5173/eventos> o llegar desde el menú («Eventos y actividades»).
3. Ver los dos bloques en orden: primero lo más próximo. Cada tarjeta dice qué es (evento o actividad), cuándo, a qué hora y dónde; el enlace de lugar abre la ubicación en el mapa.
4. Para el detalle completo, se toca la tarjeta. Para cambiar a inglés, se usa el selector de la cabecera.

### Panel (con cuenta)

1. **Entrar**: <http://localhost:5173/login> y elegir en el menú **Eventos y actividades → Eventos**, **Actividades** o **Categorías** (también hay accesos directos en la pantalla de Inicio). Si la persona no tiene el permiso del módulo, la sección no aparece y el sitio no deja entrar, con un mensaje claro.
2. **Crear un evento**: título, fechas (se agregan las que hagan falta, con hora de inicio y fin opcionales), lugar, detalle e imagen opcional; si quieren, el texto de frecuencia. Al guardar queda en **borrador** hasta que alguien toca **Publicar**.
3. **Crear una actividad**: lo mismo, pero con **una sola fecha** y su horario opcional.
4. **Categorías**: catálogo opcional (acción social, predicación, adoración, entrevista, noche de oración, otro…). Si intentan borrar una categoría que se está usando, el sistema lo impide y avisa «quítala primero».
5. **Publicar o retirar**: como en la portada, cada elemento se publica y se retira por separado; si editan algo ya publicado y guardan, el cambio se ve de inmediato en el sitio.
6. **Idioma de cada texto**: cada formulario tiene la pestaña **English** al lado de la española. Es opcional: dejen el inglés vacío si quieren y el sitio mostrará el español.
7. **Pasado**: un evento o actividad cuya fecha ya venció no se borra: en el panel aparece con la marca «Pasado» y simplemente ya no se muestra en la sección pública.
8. **Auditoría**: todo lo anterior queda en `/panel/auditoria`, con filtros por cuenta y rango de fechas.

Si algo no se puede guardar, el formulario avisa **junto al campo** qué corregir y no se guarda nada a medias.

## Cómo probarlo

La guía completa está en [`specs/004-eventos-actividades/quickstart.md`](../../specs/004-eventos-actividades/quickstart.md):

- **§0 · Preparar el entorno**: `make up` y `make db-migrate` (aplica las migraciones nuevas).
- **§1 · La sección sin cuenta**: abrir <http://localhost:5173/eventos> sin iniciar sesión.
- **§2 · Permisos**: crear un rol con «Eventos», «Actividades» y/o «Categorías de agenda» y comprobar que no se mezclan.
- **§3–§6 · Panel**: crear eventos con fechas, actividades puntuales, categorías e imágenes.
- **§7 · Publicar y retirar**, **§8 · idiomas**, **§9 · retiro automático al pasar la fecha**, **§10 · dispositivos y accesibilidad**, **§11 · sin permiso**, **§12 · auditoría**, **§13 · pruebas automatizadas** (Go + 530 pruebas de frontend + 46 casos extremo a extremo en verde) y **§14 · mapa de criterios**.
- **§15 · Bloque «Lo que viene» de la portada**: los 5 ítems, la ventana de 2 meses con respaldo, el mensaje cuando no hay nada y el botón «Ver más eventos».

Nota para quien ejecute las pruebas extremo a extremo: se corren **en cuatro tandas separadas** (fondo oscuro de las pantallas de acceso, F3/F10 + bloque, F4 + bugs, F2) y en serie, por el límite de intentos de inicio de sesión por minuto; el detalle está en la cabecera de los propios archivos de prueba.

## Correcciones incluidas

Además de lo anterior, esta entrega incluye **cuatro arreglos de errores detectados en la revisión**:

- **Desde la sección `/eventos`, el menú no llevaba a las secciones de la portada**: los enlaces de las secciones se construían mal al venir de otra página. Ya navegan bien y, al llegar, la página se coloca en la sección elegida.
- **El menú público no marcaba la sección actual ni respondía al pasar el cursor**: ahora la sección en la que estás parado aparece resaltada (y lo anuncia el lector de pantalla) y el efecto al pasar el cursor se ve con el contraste adecuado.
- **El diálogo para agregar un evento era angosto y con desplazamiento lateral**: ahora se abre más ancho y se pueden agregar varias fechas sin desplazarse a los costados.
- **Las pantallas de acceso salían con fondo negro sobre fondo negro**: si el teléfono o la computadora estaba en **modo oscuro**, en <http://localhost:5173/login>, en el cambio de contraseña, en la pantalla «No tienes permiso» y en el «Página no encontrada» (404) el texto casi no se leía. Ahora esas cuatro pantallas se pintan siempre con **fondo claro**, sin importar el tema del dispositivo. Era un descuido del rediseño anterior (F10), no de esta funcionalidad, y ya tiene su prueba de regresión.

Se aprovechó además para cumplir con el requisito de objetivos táctiles grandes (≥ 44 px) en las tarjetas del bloque nuevo.

## Ajuste de posición del bloque (decisión tuya, 2026-10-12)

El bloque **«Lo que viene»** pasó de estar entre el horario y los canales de WhatsApp a ser **la primera sección de la portada, después de la imagen de bienvenida y antes de «Quiénes somos»**. Solo cambió su lugar: el resto de la portada se recorre hacia abajo sin perder nada (mismo contenido, mismos enlaces y misma entrada desde el menú), y la portada se ve completa en todas las pantallas.

## Límites conocidos de esta entrega

- **No hay inscripciones ni avisos**: la gente se entera entrando al sitio; no se envían correos, mensajes ni recordatorios, y no hay botón «agregar a mi calendario».
- **No se calculan fechas automáticamente**: el equipo carga las fechas concretas de cada evento (y opcionalmente un texto de frecuencia). El sistema no genera la próxima fecha por periodicidad.
- **El bloque de la portada no se puede configurar**: no hay más de 5 ítems, ni paginación, ni filtros, ni ajuste de fechas desde el panel, y F4 **no rediseña el resto de la portada de F3** (solo añade ese bloque). Todo lo demás de la portada sigue siendo de F3.
- **Borrar es borrar para siempre** (igual que en la portada): si algo no quieren perder, usen «Retirar» en vez de eliminarlo.
- **Sin traducción automática**: el inglés lo escribe el equipo, contenido por contenido.
- **El panel sigue en español** (como F2 y F3): el bilingüismo es del sitio público.
- **Quedan dos revisiones solo con personas**, ya preparadas para esta entrega: la **prueba de usabilidad** (SC-010/SC-015: personas de distintas edades identifican qué viene, cuándo y dónde sin ayuda, ahora también desde la portada) y la **lectura con lector de pantalla** (NVDA/VoiceOver) de la sección y del bloque nuevo. Lo automatizable de accesibilidad —teclado, foco, jerarquía, contraste, textos al 200 %, 320 px— ya está en verde.
- **Sin despliegue a internet**: por ahora todo funciona en local, en tu equipo.

## Qué sigue

1. **La revisión de aceptación con ustedes** (usabilidad y lector de pantalla, arriba) y el **merge del Pull Request** de la rama `004-eventos-actividades` — el merge lo aprueba una persona.
2. **F5 — Grupos de conexión**, y siguen ministerios, donaciones, noticias y medios, en el orden del [roadmap](../producto/roadmap.md).
3. **F10B** — el rediseño del panel de administración con el sistema visual de F10 (sin cambiar ninguna función).

## Enlaces

- [`README.md`](../../README.md) — instalación, comandos y variables de entorno.
- [`CHANGELOG.md`](../../CHANGELOG.md) — qué cambió en la versión 0.5.0.
- [`specs/004-eventos-actividades/spec.md`](../../specs/004-eventos-actividades/spec.md) — la especificación aprobada.
- [`docs/producto/roadmap.md`](../producto/roadmap.md) — las funcionalidades del proyecto y su estado.
