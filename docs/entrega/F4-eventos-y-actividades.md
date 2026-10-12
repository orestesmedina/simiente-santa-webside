# Entrega F4 — Eventos y actividades

**Fecha:** 2026-10-11 · **Versión:** 0.5.0 · **Rama:** `004-eventos-actividades`

Nota de entrega en lenguaje sencillo. Si querés la parte técnica, está en el [`README`](../../README.md), en el [`CHANGELOG`](../../CHANGELOG.md) y en la [spec de la funcionalidad](../../specs/004-eventos-actividades/spec.md).

## Qué se entrega

Ya está la sección que resuelve el problema principal: **que nadie se quede sin enterarse de lo que viene**. El sitio público muestra los próximos eventos y actividades de la iglesia, y el equipo los administra desde el panel, sin depender de nadie más.

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

1. Abrir <http://localhost:5173/eventos> o llegar desde el menú («Eventos y actividades»).
2. Ver los dos bloques en orden: primero lo más próximo. Cada tarjeta dice qué es (evento o actividad), cuándo, a qué hora y dónde; el enlace de lugar abre la ubicación en el mapa.
3. Para el detalle completo, se toca la tarjeta. Para cambiar a inglés, se usa el selector de la cabecera.

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
- **§7 · Publicar y retirar**, **§8 · idiomas**, **§9 · retiro automático al pasar la fecha**, **§10 · dispositivos y accesibilidad**, **§11 · sin permiso**, **§12 · auditoría**, **§13 · pruebas automatizadas** (Go + 505 pruebas de frontend + 38 casos extremo a extremo en verde) y **§14 · mapa de criterios**.

Nota para quien ejecute las pruebas extremo a extremo: se corren **en tres tandas separadas** (F2, F4, F3/F10) y en serie, por el límite de intentos de inicio de sesión por minuto; el detalle está en la cabecera de los propios archivos de prueba.

## Límites conocidos de esta entrega

- **No hay inscripciones ni avisos**: la gente se entera entrando al sitio; no se envían correos, mensajes ni recordatorios, y no hay botón «agregar a mi calendario».
- **No se calculan fechas automáticamente**: el equipo carga las fechas concretas de cada evento (y opcionalmente un texto de frecuencia). El sistema no genera la próxima fecha por periodicidad.
- **No hay «Lo que viene» en la portada**: la sección vive bajo su propia entrada de menú; la portada de F3 no se tocó. Si lo quieren en la portada, sería un pedido aparte.
- **Borrar es borrar para siempre** (igual que en la portada): si algo no quieren perder, usen «Retirar» en vez de eliminarlo.
- **Sin traducción automática**: el inglés lo escribe el equipo, contenido por contenido.
- **El panel sigue en español** (como F2 y F3): el bilingüismo es del sitio público.
- **Quedan dos revisiones solo con personas**, ya preparadas para esta entrega: la **prueba de usabilidad** (SC-010: personas de distintas edades identifican qué viene, cuándo y dónde sin ayuda) y la **lectura con lector de pantalla** (NVDA/VoiceOver) de la sección. Lo automatizable de accesibilidad —teclado, foco, jerarquía, contraste, textos al 200 %, 320 px— ya está en verde.
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
