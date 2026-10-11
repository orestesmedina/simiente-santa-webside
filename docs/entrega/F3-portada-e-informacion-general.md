# Entrega F3 — Portada e información general

**Fecha:** 2026-10-10 · **Versión:** 0.3.0 · **Rama:** `003-portada-info-general`

Nota de entrega en lenguaje sencillo. Si querés la parte técnica, está en el [`README`](../../README.md), en el [`CHANGELOG`](../../CHANGELOG.md) y en la [spec de la funcionalidad](../../specs/003-portada-info-general/spec.md).

## Qué se entrega

El sitio ya tiene **cara pública**: cualquiera puede abrirlo y conocer la iglesia sin registrarse ni pedir permiso. Y el equipo ya puede mantener esa información al día **sin depender de un desarrollador**.

- **La portada de la iglesia** (ahora es la página de inicio del sitio): el nombre y la identidad de Simiente Santa con su logotipo e imagen, «quiénes somos», el horario de servicios, los canales de WhatsApp, las redes sociales y el contacto (dirección, correo y teléfono). Cada enlace de WhatsApp y de redes abre el chat o el perfil con un solo toque.
- **En español y en inglés**: un selector arriba cambia todo el sitio de inmediato. Lo que ustedes no traduzcan se muestra en español —nunca un hueco ni una traducción automática—, y el sitio recuerda el idioma elegido en ese dispositivo.
- **Pensada para todos**: se ve bien en teléfono, tableta y computadora (desde las pantallas más estrechas), se usa con el dedo o con el teclado, y respeta el Manual de Identidad de la iglesia (los colores, las tipografías y las reglas del logotipo: no se deforma, no se gira, no se le cambian los colores ni se le agregan efectos).
- **El panel para administrarla**, en `/panel/informacion`: seis pestañas —Identidad, Quiénes somos, Horario de servicios, Canales de WhatsApp, Redes sociales y Contacto—, con la opción de **subir el logotipo y la imagen de portada** y con un estado visible de cada elemento: **Borrador** o **Publicado**.
- **Control de lo que es público**: nada se publica solo. Cada pieza se publica o se retira por separado, y una sección que está toda en borrador desaparece por completo de la portada (sin dejar huecos ni campos vacíos). Si editan algo que ya estaba publicado y guardan, el cambio se ve de inmediato.
- **Permiso propio y registro de cambios**: para entrar hacen falta los permisos de F2; este módulo usa el permiso «Portada e información general». Quien no lo tenga no ve ni puede tocar nada de esta sección. Y **cada edición queda registrada** en `/panel/auditoria`: quién lo hizo, qué cambió y cuándo.

## Cómo se usa (para el equipo de la iglesia)

1. **Entrar**: abran <http://localhost:5173/login> e inicien sesión con su cuenta de panel. En el menú lateral elijan **«Portada e información general»** (también hay acceso directo en la pantalla de Inicio). Si esa persona no tiene el permiso del módulo, la sección no aparece y el sitio no deja entrar, con un mensaje claro.
2. **Identidad**: nombre oficial de la iglesia, lema, misión y visión. Ahí mismo se **suben el logotipo y la imagen de portada** (fotos en JPG, PNG o WebP, de hasta 8 MB) y se escribe el texto alternativo —esa descripción corta que leen las personas que usan lector de pantalla—.
3. **Quiénes somos**: el texto con el que la iglesia se presenta, hasta 1.000 caracteres.
4. **Horario de servicios**: se agrega un servicio por cada reunión (día, hora, nombre y lugar) y cada uno se publica o retira por su cuenta.
5. **Canales de WhatsApp**: puede haber varios; cada uno con su nombre (p. ej. «Escríbenos» o «Grupo de la iglesia») y su destino: un número para mensaje directo o el enlace del grupo.
6. **Redes sociales**: Facebook, Instagram, YouTube, TikTok y Spotify, con un solo enlace por red.
7. **Contacto**: dirección, correo y teléfono; los tres son obligatorios para poder guardar.
8. **Idioma de cada texto**: cada formulario tiene la pestaña **English** al lado de la española. Es opcional: si dejan el inglés vacío y guardan, no pasa nada (el sitio mostrará ese texto en español).
9. **Borrador o publicado**: al guardar algo nuevo queda en **borrador** —solo el equipo lo ve— hasta que alguien toca **«Publicar»**. **«Retirar de la portada»** devuelve el elemento a borrador sin borrarlo. Si editan algo ya publicado y guardan, el cambio se ve de inmediato en la portada.
10. **Cambiar el idioma del sitio**: en la portada, arriba a la derecha, está el selector «Español / English». El sitio recuerda la elección en ese dispositivo; la primera visita siempre entra en español.

Si algo no se puede guardar, el formulario avisa **junto al campo** qué corregir (un correo mal escrito, un enlace incompleto, un campo obligatorio vacío…) y no se guarda nada a medias.

## Cómo probarlo

La guía completa está en [`specs/003-portada-info-general/quickstart.md`](../../specs/003-portada-info-general/quickstart.md):

- **§0 · Preparar el entorno**: `make up` y `make db-migrate`.
- **§1 · La portada sin cuenta**: abrir <http://localhost:5173/> sin iniciar sesión.
- **§2–§4 · Panel**: entrar con una cuenta con el permiso y editar identidad, «quiénes somos», contacto, horario, WhatsApp y redes.
- **§5 · Borrador y publicación**, **§6 · logo e imagen**, **§7 · idiomas**, **§8 · sin permiso**, **§9 · auditoría**, **§10 · dispositivos**, **§11 · pruebas automatizadas** (7 end-to-end en verde).

## Límites conocidos de esta entrega

- **Todavía no hay eventos, grupos, ministerios ni donaciones** en el sitio: eso llega con las siguientes funcionalidades (F4–F9, en el orden del [roadmap](../producto/roadmap.md)).
- **No hay formularios de contacto ni inscripción**: la gente les escribe por WhatsApp, redes, correo o teléfono. La dirección se muestra como texto, sin mapa dentro del sitio.
- **No hay traducción automática**: el inglés lo escriben ustedes, contenido por contenido.
- **Borrar un elemento lo borra para siempre** (sin historial de versiones ni recuperación). Si algo no quieren perder, usen «Retirar de la portada» en vez de borrarlo.
- **El sitio no comprueba que los enlaces sigan vivos**: si caduca un grupo de WhatsApp o cambia un perfil, hay que corregir el enlace desde el panel.
- **Quedan dos revisiones con personas pendientes**, ya coordinadas para esta fase de entrega: una **prueba de usabilidad** con personas de distintas edades (encontrar el horario y un canal de WhatsApp, cambiar de idioma) y la **revisión manual de accesibilidad** (teclado, lector de pantalla, contraste). La parte automatizada ya está en verde.
- **Sin despliegue a internet**: por ahora todo funciona en local, en tu equipo.

## Qué sigue

**F4 — Eventos y actividades**: que nadie se quede sin enterarse de lo que viene. Después siguen grupos de conexión, ministerios y donaciones, en el orden del [roadmap](../producto/roadmap.md).

## Enlaces

- [`README.md`](../../README.md) — instalación, comandos y variables de entorno.
- [`CHANGELOG.md`](../../CHANGELOG.md) — qué cambió en la versión 0.3.0.
- [`specs/003-portada-info-general/spec.md`](../../specs/003-portada-info-general/spec.md) — la especificación aprobada.
- [`docs/producto/roadmap.md`](../producto/roadmap.md) — las funcionalidades F1–F9 y su estado.
