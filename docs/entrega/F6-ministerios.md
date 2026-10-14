# Entrega F6 — Ministerios

**Fecha:** 2026-10-13 · **Versión:** 0.6.0 · **Rama:** `006-ministerios`

Nota de entrega en lenguaje sencillo. Si querés la parte técnica, está en el [`README`](../../README.md), en el [`CHANGELOG`](../../CHANGELOG.md) y en la [spec de la funcionalidad](../../specs/006-ministerios/spec.md).

## Qué se entrega

Ahora el sitio **muestra los ministerios de la iglesia** y el equipo los administra solo, desde el panel. Nada de esto existía antes.

- **La sección «Ministerios»** (en <http://localhost:5173/ministerios>), visible **sin cuenta ni registro**: una ficha por ministerio con su **nombre**, su **descripción**, su **horario** y su **lugar** escritos tal cual los ponga el equipo («todos los viernes 7 p. m.», «una vez al mes, fecha variable»), una **imagen** si quieren cargarla, el **encargado** cuando deciden mostrarlo, y el botón **«Unirse al grupo de WhatsApp»** cuando el ministerio tiene enlace.
- **El enlace es opcional**: un ministerio puede publicarse perfectamente **sin enlace** (la ficha se ve completa, sin botones vacíos). Y cuando sí hay enlace, se acepta **cualquier dirección segura que empiece por `https`** —no hace falta que sea de WhatsApp—; lo que no sea seguro (`http://`, direcciones raras) se rechaza con un aviso claro y **no se guarda nada a medias**.
- **El encargado también es opcional** y su ficha está **«oculta por defecto»**: aunque carguen el nombre y el contacto del encargado, el sitio público **no los muestra hasta que alguien encienda el interruptor** de forma expresa. Ver el apartado de privacidad más abajo.
- **El orden lo deciden ustedes**: cada ministerio tiene un número de orden y con él se reordena el listado público (el que va primero se ve primero). No depende de la fecha de creación.
- **En español y en inglés**, igual que la portada y la sección de eventos: el selector de la cabecera cambia todo de inmediato y lo que no esté traducido se muestra en español (nunca un hueco).
- **Aparece solo cuando hay contenido**: mientras no haya **ningún ministerio publicado**, la entrada «Ministerios» no se muestra en la barra, en el menú ni en el pie, **y si alguien abre la dirección directamente lo manda a la portada** (sin avisos ni pantallas raras). Esta es una diferencia a propósito con la sección de eventos: ahí aparecía un aviso y aquí se redirige — así lo decidiste en la puerta 2.
- **El bloque «Ministerios» en la portada**, colocado **justo después del bloque de eventos y antes de «Quiénes somos»**: hasta **5 fichas** (los primeros según el orden que ustedes fijen) con su imagen, su nombre y su horario/lugar, con el botón **«Ver más ministerios»** para llegar a la sección completa. **Si no hay nada publicado, el bloque no aparece** (la portada queda limpia, como siempre).
- **El panel para administrarla**: **`/panel/ministerios`**, con alta, edición, publicación, retirada y eliminación con confirmación; subida de imagen; reordenación; y el **interruptor de visibilidad del encargado**. En Inicio aparece el acceso directo **«Gestión de ministerios»**.
- **Un permiso existente, sin permisos nuevos**: se usa **«Ministerios»**, el permiso que ya estaba reservado desde la gestión de usuarios (F2). Quien no lo tiene, **no ve ni puede tocar el módulo** —el servidor lo comprueba en cada operación y responde con un mensaje claro—. El catálogo de permisos **sigue teniendo 10**: no se agregó ninguno nuevo.
- **Todo queda registrado**: cada alta, edición, publicación, retirada, cambio de visibilidad del encargado, subida de imagen, eliminación y denegación aparece en `/panel/auditoria` con texto legible («Creó el ministerio», «Publicó la ficha del encargado del ministerio», «Retiró el ministerio»…) y el nombre del ministerio al lado, igual que en F2, F3 y F4.

## Qué pasa con los datos del encargado (lo importante)

El sitio puede publicar el **nombre** de una persona encargada y **un** medio de contacto: **celular o correo, nunca los dos** (y también se puede guardar solo el nombre, sin contacto).

- **Por defecto está todo oculto**: el interruptor **«Mostrar la ficha del encargado en el sitio público»** nace **apagado**. Mientras siga apagado, **ni el nombre ni el contacto salen del servidor**: no aparecen en la sección, ni en el bloque de la portada, ni en los registros internos. Lo que pasa en el panel (donde solo entra quien tiene el permiso) sí conserva el dato completo.
- **Publicar es un acto expreso**: para que se vea la ficha del encargado hay que encender el interruptor ministerio por ministerio.
- **Se puede apagar y encender las veces que haga falta** sin perder lo que se escribió: al apagar, la ficha desaparece del sitio en el acto; al encender, vuelve con todos los datos intactos.
- **El bloque de la portada nunca muestra ni el encargado ni el enlace**, aunque la ficha esté publicada: en la portada solo van imagen, nombre y horario/lugar.

## Cómo se usa (para el equipo de la iglesia)

### Visitante (sin cuenta)

1. **Desde la portada** (<http://localhost:5173/>): el bloque **«Ministerios»**, después de los eventos y antes de «Quiénes somos», muestra los primeros ministerios; tocando cualquiera (o el botón **«Ver más ministerios»**) se llega a la sección completa.
2. Abrir <http://localhost:5173/ministerios> o llegar desde el menú («Ministerios»).
3. Cada ficha dice qué es el ministerio, cómo participar, a qué hora y en qué lugar; si el equipo publicó el encargado, se ve su nombre y su contacto; si hay enlace, el botón **«Unirse al grupo de WhatsApp»** lleva al grupo en un paso.
4. Para cambiar a inglés, se usa el selector de la cabecera.

### Panel (con cuenta)

1. **Entrar**: <http://localhost:5173/login> y elegir **Inicio → Gestión de ministerios** (o el menú del panel). Si la persona no tiene el permiso «Ministerios», la sección no aparece y el sitio no deja entrar, con un mensaje claro.
2. **Crear un ministerio**: nombre y descripción en español son lo único obligatorio; encargado (nombre + celular **o** correo), enlace, horario, lugar e imagen son **opcionales**. Al guardar queda en **borrador** hasta que alguien toque **Publicar**.
3. **Decidir si se muestra el encargado**: el interruptor **«Mostrar la ficha del encargado en el sitio público»** nace **apagado**; enciéndanlo solo cuando quieran que el nombre y el contacto se vean afuera.
4. **Ordenar**: con el campo de orden colocan cada ministerio donde quieran en la lista pública.
5. **Publicar o retirar**: como en la portada y en eventos, cada ministerio se publica y se retira por separado; los cambios se ven de inmediato en el sitio.
6. **Idioma de cada texto**: cada formulario tiene la pestaña **English** al lado de la española. Es opcional: dejen el inglés vacío y el sitio mostrará el español.
7. **Eliminar**: con confirmación; **borrar es borrar para siempre** (si no quieren perderlo, usen «Retirar»).
8. **Auditoría**: todo lo anterior queda en `/panel/auditoria`, con filtros por cuenta y rango de fechas.

Si algo no se puede guardar, el formulario avisa **junto al campo** qué corregir y no se guarda nada a medias.

## Cómo probarlo

La guía completa está en [`specs/006-ministerios/quickstart.md`](../../specs/006-ministerios/quickstart.md):

- **§0 · Preparar el entorno**: `make up` y `make db-migrate` (aplica las migraciones nuevas `000010` y `000011`).
- **§1 · La sección sin cuenta** y **§7 · Privacidad del encargado** (el punto más importante).
- **§2 · Permisos**: crear un rol con «Ministerios» y comprobar que no se mezcla con eventos o portada.
- **§3–§6 · Panel**: crear ministerios, encargado opcional, enlaces aceptados/rechazados y orden manual.
- **§8 · Publicar y retirar**, **§9 · idiomas**, **§10 · imágenes**, **§11 · dispositivos y accesibilidad**, **§12 · sin permiso**, **§13 · auditoría**, **§14 · bloque de la portada**, **§15 · pruebas automatizadas** (Go + 652 pruebas de frontend + 19 casos extremo a extremo de F6 en verde) y **§16 · mapa de criterios**.

Nota para quien ejecute las pruebas extremo a extremo: se corren **en cuatro tandas separadas** (usuarios y auditoría, eventos, portada/diseño, ministerios) y en serie, por el límite de intentos de inicio de sesión por minuto; el detalle está en la cabecera de los propios archivos de prueba.

## Decisiones tuyas que quedaron aplicadas

- **Encargado opcional y con visibilidad configurable** (y el interruptor **apagado por defecto**).
- **Un solo medio de contacto: celular o correo** —y se puede guardar solo el nombre, sin contacto.
- **Enlace opcional** y **cualquier `https` seguro** servido (sin restringir a WhatsApp).
- **Imagen opcional**, cargada por el equipo.
- **Orden manual** con un número de orden.
- **Horario y lugar como textos libres** (sin calendario ni fechas calculadas).
- **Bloque de ministerios en la portada**, después de los eventos y antes de «Quiénes somos», con **tope de 5**.
- **`/ministerios` sin publicados redirige a la portada**, de forma silenciosa (distinto a `/eventos`).

## Límites conocidos de esta entrega

- **Un encargado por ministerio**: no hay listas de miembros, ni inscripciones, ni cobros, ni asistencia, ni avisos automáticos (correo/WhatsApp); «unirse» es el enlace del grupo.
- **Sin calendario**: el sistema no calcula la próxima fecha ni manda recordatorios; el horario es el texto que ustedes escriban.
- **Sin traducción automática**: el inglés lo escribe el equipo, contenido por contenido.
- **El panel sigue en español** (como F2, F3 y F4).
- **Borrar es borrar para siempre**: si algo no quieren perder, usen «Retirar» o apaguen el interruptor del encargado.
- **Quedan dos revisiones solo con personas**, ya preparadas: la **prueba de usabilidad** (6 personas de distintas edades completan 3 tareas sin ayuda, con cronómetro para el «en menos de 30 s») y la **lectura con lector de pantalla** (NVDA/VoiceOver) de la sección y del bloque. Lo automatizable de accesibilidad —teclado, foco, contraste, textos al 200 %, 320 px— ya está en verde. Protocolo en [`specs/006-ministerios/pruebas-usabilidad-SC-012.md`](../../specs/006-ministerios/pruebas-usabilidad-SC-012.md).
- **Sin despliegue a internet**: por ahora todo funciona en local, en tu equipo.

## Qué sigue

1. **La revisión de aceptación con ustedes** (usabilidad y lector de pantalla, arriba) y el **merge del Pull Request** de la rama `006-ministerios` — el merge lo aprueba una persona.
2. **F5 — Grupos de conexión**, y siguen donaciones, noticias y medios, en el orden del [roadmap](../producto/roadmap.md).
3. **F10B** — el rediseño del panel de administración con el sistema visual de F10 (sin cambiar ninguna función).

## Enlaces

- [`README.md`](../../README.md) — instalación, comandos y variables de entorno.
- [`CHANGELOG.md`](../../CHANGELOG.md) — qué cambió en la versión 0.6.0.
- [`specs/006-ministerios/spec.md`](../../specs/006-ministerios/spec.md) — la especificación aprobada.
- [`docs/producto/roadmap.md`](../producto/roadmap.md) — las funcionalidades del proyecto y su estado.
