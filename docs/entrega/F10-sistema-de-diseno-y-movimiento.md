# Entrega F10 — Sistema de diseño y movimiento

**Fecha:** 2026-10-11 · **Versión:** 0.4.0 · **Rama:** `010-sistema-diseno-movimiento`

Nota de entrega en lenguaje sencillo. Si querés la parte técnica, está en el [`README`](../../README.md), en el [`CHANGELOG`](../../CHANGELOG.md) y en la [spec de la funcionalidad](../../specs/010-sistema-diseno-movimiento/spec.md).

## Qué se entrega

El sitio público deja de verse «de 2010»: es el **rediseño moderno** que pediste (la propuesta visual **B «Marea Turquesa»**, la que elegiste), aplicado a la portada de hoy y dejado como base para las secciones que vienen después.

- **Un aspecto moderno y cuidado, fiel a la marca**: los mismos colores, tipografías y reglas del Manual de Identidad, pero con composición actual, más aire entre secciones y componentes iguales en todo el sitio (los mismos botones, títulos y tarjetas para lo mismo). El logotipo sigue intacto: sin deformar, sin recolorear, sin girar y sin efectos.
- **Animaciones agradables**: las secciones se revelan al bajar la página y los botones y enlaces responden al toque, al pasar el cursor y al enfocarlos con teclado. Nada bloquea ni retrasa la lectura: la animación es una capa visual sobre el contenido, que siempre está disponible desde el primer instante. Y si el dispositivo tiene activada la preferencia de **«reducir movimiento»**, el sitio la respeta: todo aparece completo, sin animaciones.
- **Menú en el móvil**: en la computadora la navegación sigue siendo una **barra visible y fija**; en el teléfono aparece un **botón «Menú»** que abre la lista completa de secciones. Es un botón claro y grande, se cierra con su propia «X» o con la tecla Escape, y desaparece para siempre el «scroll incómodo»: ningún enlace se sale de lado y la navegación nunca se pierde al bajar.
- **Cambio de idioma más compacto**: el mismo selector «Español / English», con los nombres completos de cada idioma, pero más pequeño y sin ocupar espacio de más.
- **Tema claro y oscuro**: el sitio sigue por defecto el modo del dispositivo (si el teléfono está en oscuro, el sitio se ve oscuro); además hay un **interruptor en la cabecera** para elegir claro u oscuro a gusto. La elección se recuerda en ese dispositivo para las próximas visitas. En los dos temas el texto se lee bien (contraste alto) y el logotipo conserva siempre sus colores originales.
- **Mapa del sitio en el pie**: con todas las secciones visibles, agrupadas en **Explorar / Contacto / Síguenos**, para llegar a cualquier parte desde abajo.
- **Pensado para todas las edades**: se usa igual con dedo, mouse y teclado; funciona con lector de pantalla; se ve bien desde los 320 px de ancho (teléfonos pequeños) y con el texto ampliado al 200 %.
- **Base para lo que viene**: este sistema es el que usarán las próximas secciones (eventos, grupos, ministerios, donaciones, noticias y medios), así que todo el sitio se sentirá como una sola experiencia.

## Qué no cambia

- **El contenido y su gestión siguen igual**: la portada muestra lo mismo de antes (identidad, «quiénes somos», horario de servicios, WhatsApp, redes y contacto), en español e inglés, con la misma publicación por elemento (borrador/publicado) desde `/panel/informacion`, con los mismos permisos y con la misma auditoría. **Solo cambia cómo se ve, no qué muestra ni cómo se administra.**
- **El panel de administración no se toca**: mantiene su aspecto actual y todas sus funciones de F2/F3.
- **No hay datos, formularios ni direcciones web (URLs) nuevas**: nada de lo que ya existe se movió ni se perdió.

## Cómo verlo

1. **Levantá el sitio**: `make up` (y `make db-migrate` la primera vez). Abrí <http://localhost:5173/> sin iniciar sesión.
2. **Compará con antes**: con contenido cargado y publicado, bajá por la portada completo y mirá cada sección; se nota el cambio de aspecto y el «aire» nuevo entre secciones.
3. **En el teléfono** (o achicando la ventana): aparece el botón **«Menú»**; abrilo, comprobá que todas las secciones están ahí y cerralo con la «X» o con Escape.
4. **Bajá despacio por la página**: las secciones se revelan al entrar en pantalla.
5. **Tema**: usá el interruptor de la cabecera para pasar de claro a oscuro (y recargá para ver que se recuerda). Si tu dispositivo ya está en modo oscuro, entrá con una pestaña limpia y el sitio se mostrará oscuro.
6. **Idioma**: el selector «Español / English» de la cabecera sigue cambiando todo de inmediato y recordando la elección.
7. **El pie**: abajo está el mapa del sitio con las secciones visibles.

La guía completa de revisión, con todos los detalles y la checklist de aceptación, está en [`specs/010-sistema-diseno-movimiento/quickstart.md`](../../specs/010-sistema-diseno-movimiento/quickstart.md) (§9 es la lista para revisar juntos).

## Límites conocidos de esta entrega

- **El panel de administración conserva su aspecto actual**: su rediseño con este mismo sistema visual quedó como **F10B**, una funcionalidad aparte.
- **Todavía no hay eventos, grupos, ministerios ni donaciones**: el sitio muestra lo mismo de antes (la portada de F3); las secciones nuevas llegan con F4–F9 y ya nacerán con este sistema.
- **Quedan cosas solo para revisión con personas**, que no se prueban con computadora: que el sitio se sienta **moderno y fiel a la marca** a su gusto, que el movimiento se note y no maree, que no se haya copiado nada de las referencias y cómo lo usan personas de distintas edades. Es la revisión de aceptación que hacemos con ustedes antes del merge.
- **Si alguien tiene activado «reducir movimiento»** en su dispositivo (o en el navegador), no verá las animaciones: es a propósito, y el sitio funciona completo.
- **Sin despliegue a internet**: por ahora todo funciona en local, en tu equipo.

## Qué sigue

1. **La revisión de aceptación con ustedes** (puntos anteriores) y el merge del Pull Request de la rama `010-sistema-diseno-movimiento`.
2. **F4 — Eventos y actividades**: que nadie se quede sin enterarse de lo que viene, ya con el sistema visual nuevo.
3. **F10B** — el rediseño del panel de administración con este mismo lenguaje visual (sin cambiar ninguna de sus funciones), en el orden que decidan.

## Enlaces

- [`README.md`](../../README.md) — instalación, comandos y variables de entorno.
- [`CHANGELOG.md`](../../CHANGELOG.md) — qué cambió en la versión 0.4.0.
- [`specs/010-sistema-diseno-movimiento/spec.md`](../../specs/010-sistema-diseno-movimiento/spec.md) — la especificación aprobada.
- [`docs/producto/roadmap.md`](../producto/roadmap.md) — las funcionalidades del proyecto y su estado.
