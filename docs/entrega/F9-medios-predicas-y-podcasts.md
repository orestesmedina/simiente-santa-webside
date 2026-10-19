# Entrega F9 — Medios: prédicas y podcasts

**Fecha:** 2026-10-18 · **Versión:** 0.8.0 · **Rama:** `009-medios-predicas-y-podcasts`

Nota de entrega en lenguaje sencillo. Si querés la parte técnica, está en el [`README`](../../README.md), en el [`CHANGELOG`](../../CHANGELOG.md) y en la [spec de la funcionalidad](../../specs/009-medios-predicas-y-podcasts/spec.md).

## Qué se entrega

Ahora el sitio **reúne las prédicas y los podcasts de la iglesia** en un solo lugar, y el equipo los publica **con un botón**, sin escribir ni copiar ningún dato. Nada de esto existía antes.

- **La sección «Medios»** (en <http://localhost:5173/medios>), visible **sin cuenta ni registro**: una ficha por sermón con **carátula, nombre, fecha y una descripción corta**, ordenada **del más reciente al más antiguo**, con el botón **«Cargar más»** cuando hay muchas. Cada ficha lleva sus enlaces: **«Ver en YouTube»** y, cuando el sermón también está en el podcast, **«Escuchar en Spotify»**.
- **El bloque de medios al final de la portada**: los **últimos 5** episodios (los últimos que se publicaron) con el botón **«Ver todo»**, colocado **debajo del bloque de ofrendas y diezmos**. Si no hay nada visible, el bloque no aparece y la portada queda limpia.
- **Los enlaces salen del sitio, en otra pestaña**: al tocar «Ver en YouTube» o «Escuchar en Spotify», el navegador abre YouTube o Spotify en una **pestaña nueva** y **la página de la iglesia sigue abierta**. El enlace **avisa** de eso (a qué servicio va y que se abre en otra pestaña), para que nadie se pierda al volver.
- **El panel para administrarla**: **`/panel/medios`** (o el acceso directo **«Gestión de medios»** en Inicio). Ahí el equipo **busca los episodios nuevos, los revisa, oculta o muestra, une los duplicados y ajusta la descripción corta**. Todo queda registrado en `/panel/auditoria`.
- **Un permiso existente, sin permisos nuevos**: se usa **«Medios (prédicas y podcasts)»**, el permiso que ya estaba reservado desde la gestión de usuarios (F2). Quien no lo tiene, **no ve ni puede tocar el módulo**. El catálogo de permisos **sigue teniendo 10**.
- **Se ve bien en cualquier dispositivo**: teléfono, tableta y computadora, en tema claro y oscuro, con teclado y con lector de pantalla.

## Cómo se usa (lo importante)

### 1. Publicar un sermón nuevo: un botón

1. Entra al panel con una cuenta que tenga el permiso **«Medios (prédicas y podcasts)»**.
2. Abre **`/panel/medios`** y pulsa **«Buscar episodios nuevos»**.
3. El sistema **lee el canal de YouTube y el feed del podcast** —los dos que ustedes ya usan— y trae **solo lo que falte**: carátula, nombre, fecha, descripción y enlaces, **sin que ustedes escriban nada**.
4. Al terminar aparece un **resumen**: cuántos episodios entraron, cuántos se unieron, cuántos quedaron marcados y, si algún feed no respondió, **cuál falló**.

Es seguro pulsarlo muchas veces: **solo agrega lo que falta, nunca borra nada y nunca cambia lo que ustedes ya decidieron** (una ficha que ustedes ocultaron **sigue oculta** después de sincronizar, y una descripción que corrigieron **no se pierde**). **No hay nada automático ni programado**: el sitio busca cuando ustedes lo piden.

**Referencia:** con los feeds de hoy, la primera búsqueda trajo **10 fichas** —8 unidas y 2 con un solo enlace—. Es un dato de comprobación, no una regla.

### 2. Una entrada por sermón (video y audio juntos)

Cuando el mismo sermón está **en YouTube y en el podcast**, el sistema lo reconoce —por el título (ignorando mayúsculas, puntos y espacios) y por ser del **mismo día**— y lo deja en **una sola ficha con los dos enlaces**. Lo que no puede unir con seguridad queda en dos fichas y **el visitante nunca ve el mismo sermón dos veces**: una queda visible y la otra oculta, **marcada** para que ustedes decidan.

### 3. Las tres marcas (y qué hacer con cada una)

El panel marca lo que necesita su atención. Las tres se ven como una etiqueta con palabra y color:

| Marca | Qué significa | Qué decide ustedes |
|---|---|---|
| **«Posible duplicado»** | El sistema no pudo estar seguro de que dos entradas sean el mismo sermón. | **«Unir»** (si es el mismo sermón: queda una sola ficha con los dos enlaces; hay confirmación porque **no se puede deshacer**) o confirmar que **son distintos** (ambas quedan visibles y sin marca). |
| **«No disponible»** | Se comprobó que el enlace de esa grabación **ya no funciona** (fue borrado o quedó privado). El público **no la ve**. | **«Mostrar de nuevo»** (si creen que ya vuelve a funcionar) o **«Mantener oculto»**. |
| **«Ya no aparece en el feed»** | El episodio salió de la lista que publica YouTube o el podcast —pasó de los ~15 últimos videos—, **pero su enlace sigue funcionando**: **sigue visible** para el público. | **«Mantener visible»** (la marca se cierra) u **ocultarlo**. |

Dos reglas que conviene saber:

- **Salir de la lista de YouTube o del podcast NO oculta nada.** Si el sistema no puede comprobar que el enlace dejó de funcionar, **tampoco oculta**: solo deja la marca para que ustedes revisen. Un episodio solo se oculta cuando se **confirma** que su enlace ya no está.
- **Una marca resuelta no se vuelve a abrir por la misma causa.**

### 4. Ocultar, mostrar y ajustar la descripción

- **Ocultar** es la forma de quitar algo del sitio (no hay borrado: si se borrara, volvería a entrar en la próxima búsqueda). Lo oculto **desaparece de la sección y del bloque de la portada** en cuanto se guarda.
- La **descripción corta** es lo único que ustedes pueden corregir —el nombre, la fecha, la carátula y los enlaces los trae el feed—. Si la dejan **vacía**, la ficha **no pinta esa línea** (no deja huecos). También pueden escribir su versión en inglés, si quieren.

## Qué NO hace (y por qué)

- **No reproduce nada dentro del sitio**: no hay video ni audio reproduciéndose en la página; el visitante **abre YouTube o Spotify en otra pestaña**. Es una **decisión del cliente** («No» a contenido de terceros, para no depender de esos servicios dentro de la iglesia); si algún día se quiere reproducir dentro, es un cambio aparte que habrá que decidir.
- **No crea episodios a mano ni los edita por completo**: no hay formulario para pegar una dirección ni para corregir el nombre, la carátula, la fecha o los enlaces. **El contenido viene de lo que ustedes publican en YouTube y en el podcast.**
- **No borra episodios**: la acción es **ocultar**.
- **No busca solo**: no hay proceso automático ni programado. Cada vez que quieran incorporar lo nuevo, pulsan **«Buscar episodios nuevos»**.
- **No administra sus cuentas de YouTube ni de Spotify**: solo lee lo que esas plataformas publican en sus feeds (son públicos, gratuitos y no llevan claves). Publicar allí sigue siendo con sus herramientas de siempre.
- **No trae todo el archivo histórico**: solo lo que los feeds ofrezcan hoy (YouTube conserva sus últimos ~15 videos). Lo que ya se importó **se queda**.
- **Sin filtros ni búsqueda en la sección**, sin descargas, sin comentarios, sin estadísticas de reproducción ni notificaciones.

## Avisos y limitaciones conocidas

- **El feed de YouTube puede no responder desde un servidor**: desde una computadora normal responde bien, pero desde direcciones de servidor devuelve error. Si eso pasa en producción, **el resumen del botón dirá cuál falló**, se importará lo del otro feed y **no se oculta ni se marca nada** por esa causa. Conviene vigilar las primeras búsquedas.
- **El podcast trae la misma carátula** (el arte del programa) para todos sus episodios de solo audio: es así en el origen, no es un error del sitio.
- **Prueba local:** en el entorno de pruebas las carátulas de los episodios reales no estaban en la carpeta de imágenes, así que ahí se ven **marcadores neutros**. Se arregla limpiando las tablas de medios y volviendo a pulsar «Buscar episodios nuevos» con el sitio levantado.
- **Defecto de las pruebas automáticas (no del producto)**: una prueba de accesibilidad de la portada (de F3, anterior a esta entrega) falla de vez en cuando porque el navegador mide un botón de 44 píxeles como 43,98; **al relanzarla pasa**. Queda registrado en el [roadmap](../../docs/producto/roadmap.md) y no afecta a lo que ven los visitantes.

## Pruebas que se harán con el sitio ya en producción

Dos comprobaciones **no se pueden automatizar** y quedan para después del despliegue, igual que en funcionalidades anteriores: la **prueba de usabilidad con personas de distintas edades** —que al menos 9 de cada 10 encuentren una prédica y lleguen a la grabación sin ayuda, que entiendan **antes de tocar el enlace** que los va a sacar del sitio hacia YouTube o Spotify en otra pestaña, y que lo logren en menos de 30 segundos— y la **revisión con un lector de pantalla real** (NVDA/VoiceOver/TalkBack). El protocolo completo está en [`specs/009-medios-predicas-y-podcasts/pruebas-usabilidad-SC-012.md`](../../specs/009-medios-predicas-y-podcasts/pruebas-usabilidad-SC-012.md).

## Cómo probarlo en dos minutos

1. Levantá el sitio (`make up`) y entrá al panel con una cuenta que tenga el permiso **«Medios (prédicas y podcasts)»** (o con el administrador inicial).
2. En **`/panel/medios`**, pulsá **«Buscar episodios nuevos»** y leé el resumen: con los feeds actuales deberían entrar **10 fichas** (8 unidas y 2 con un solo enlace).
3. Revisá el listado: probá a **ocultar** una ficha y comprobá que desaparece de la sección y del bloque; volvé a **mostrarla**.
4. Abrí <http://localhost:5173/medios> sin sesión y tocá un enlace: se abre YouTube o Spotify **en otra pestaña** y la pestaña del sitio **sigue abierta**. Probá también **«Cargar más»** y cambiá de idioma.
5. Volvé a la portada (<http://localhost:5173/>) y verás el bloque **al final, debajo del de ofrendas y diezmos**, con los últimos 5 y **«Ver todo»**.
6. Ocultá todos los episodios desde el panel (aquí **no hay borradores**: el estado es visible/oculto): la entrada **«Medios»** desaparece del menú, el bloque de la portada no aparece y abrir `/medios` te manda a la portada, sin avisos.
