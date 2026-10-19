# Cuestionario para el cliente — F9 «Medios: prédicas y podcasts»

**Para:** equipo de la iglesia (quien administra el sitio y quien publica el contenido)
**De parte de:** equipo de desarrollo · **Fecha:** 2026-10-17
**Funcionalidad:** F9 del roadmap — *Medios: prédicas y podcasts* (fase 2)

---

## Cómo usar este documento

No hace falta que sea un documento técnico. Léanlo, contesten lo que sepan y donde no estén seguros
escriban **«ustedes deciden»**: propondremos la opción que creemos mejor y se la mostraremos antes de
construir nada.

**Plazo sugerido:** mientras más rápido lleguen las respuestas, más rápido empezamos a construir.
Nada se construye hasta que ustedes aprueben la especificación (el documento que describe qué se hará).

Las respuestas se registran como decisiones del proyecto y quedan en el roadmap.

---

## 0. De qué se trata F9

Que cualquier persona que entre al sitio pueda **ver o escuchar las prédicas y los podcasts de la
iglesia sin instalar YouTube ni Spotify**, y que el equipo pueda **administrarlos desde el panel**,
igual que ya administra eventos, ministerios y ofrendas.

Ya existe la base: el panel con permisos, la portada, el sistema visual con modo claro/oscuro y la
navegación. F9 agrega **una sección nueva** y su módulo en el panel.

**El permiso «Medios (prédicas y podcasts)» ya existe** desde la funcionalidad 2, así que F9 **no
crea permisos nuevos**: solo lo activa.

---

## 1. Lo que necesitamos de ustedes (1 minuto, sin trámites)

No hay que pedir permisos a nadie ni pagar nada. Solo necesitamos saber **dónde publican hoy**:

| Lo que necesitamos | Cómo se obtiene | ¿Lo tenemos? |
|---|---|---|
| **El canal de YouTube de la iglesia** (la dirección, p. ej. `youtube.com/@nombre`) | Es público. Nosotros deducimos el identificador interno | https://www.youtube.com/@IglesiaSimienteSanta | 
| **La dirección del feed del podcast**, si ya tienen podcast en Spotify/Apple (p. ej. `https://anchor.fm/s/xxxx/podcast/rss`) | Se copia de su servicio de alojamiento: **Spotify for Creators → Ajustes → Availability → RSS Distribution**. Si usan otro servicio, se la piden a ellos | https://anchor.fm/s/1116351e8/podcast/rss | 
| **Quién administra esas cuentas** (nombre y correo) | — | Orestes Medina |

> Un **feed** es una lista que la plataforma publica sola con lo último que han subido. No hay que
> crearlo ni solicitarlo: ya existe. Lo único que hace falta es **conocer su dirección**.

---

## 2. La decisión grande: cómo se ve y se reproduce

Esta es la única pregunta que cambia de verdad el alcance y el costo. Elijan una opción.

### Opción A — Solo un enlace (lo más barato)
El equipo pega la dirección del episodio y el visitante ve un enlace que lo lleva a YouTube o Spotify.
Tiene que **salir del sitio** y, si no tiene la aplicación, el teléfono se la pide.

❌ No cumple lo que pide el documento de la idea: *«sin necesidad de tener instaladas esas aplicaciones»*.
**No la recomendamos.**

### Opción B — Enlace + reproductor dentro del sitio ⭐ **recomendada**
El equipo **pega la dirección** del video o del episodio y el sitio hace el resto:
- rellena solo el **título, el autor y la imagen** (así nadie escribe nada dos veces ni se equivoca),
- el visitante **reproduce ahí mismo**, sin salir del sitio y sin instalar nada.

**No requiere ninguna cuenta, ninguna clave y ningún pago**, ni mensualidades: se apoya en servicios
públicos de YouTube y Spotify que existen para esto.
👉 **Esta es nuestra recomendación para F9.**

### Opción C — Que se publique solo (sincronización automática)
El sitio revisa por su cuenta el canal y el podcast cada cierto tiempo y **agrega los episodios nuevos
sin que nadie los pegue**.

Es la más cómoda a largo plazo, pero:
- es una **funcionalidad aparte**, no parte de F9: necesita un proceso que corre solo, control de
  errores y, si se hiciera por las interfaces oficiales, cuentas y claves que hay que mantener;
- **Spotify hoy exige una suscripción de pago** al dueño de la aplicación que use su interfaz oficial
  (desde marzo de 2026) y limita las aplicaciones pequeñas a 5 usuarios de prueba: **no conviene
  depender de eso**. La vía barata para hacerlo algún día son los **feeds** (las listas automáticas),
  que son gratuitos y sin claves;
- **no tiene dónde correr todavía**: el sitio aún no está desplegado en un servidor, y este proceso
  necesita vivir en uno.

👉 **Nuestra recomendación:** hacer F9 con la **Opción B** ahora, y dejar la Opción C como
funcionalidad propia para después del despliegue (ya está anotada en el roadmap como idea futura).

**Respuesta:** ☐ A · ☐ B (recomendada) · ☐ C · X «ustedes deciden»

---

## 3. Preguntas de contenido

**Q1. ¿Dónde publican hoy las prédicas y los podcasts?**
☐ Solo en YouTube · ☐ Solo en Spotify · X En YouTube y en Spotify (el mismo contenido) ·
☐ En otro servicio: ______ · ☐ Todavía no publicamos nada

> Si es **el mismo contenido en los dos lados**, díganlo: hay que decidir si se muestra como **una
> sola prédica** o como dos entradas (una de video y otra de audio).

**Q2. ¿«Prédicas» y «podcasts» son una sola lista con un tipo, o dos secciones separadas?**
- X **Una sola sección** «Medios» donde cada elemento se marca como *prédica* o *podcast* y se puede
  ver por bloques (**recomendada**: menos trabajo, un solo lugar donde buscar; es el mismo patrón que
  ya usamos en Eventos y actividades).
- ☐ Dos secciones separadas con dos entradas de navegación.

**Q3. ¿Qué datos quiere llenar el equipo por cada episodio?**
Marquen los que usen. **Título y fecha** son obligatorios en cualquier caso.

X Título · X Fecha de publicación · X Descripción o resumen · X Quién predica / autor ·
X Serie o tema · ☐ Invitado · X Imagen (si no se llena, se usa la del video) · ☐ Duración ·
X Tipo (prédica / podcast) · ☐ Otro: ______

> Nota honesta: la **duración no la entrega el servicio de YouTube** en el modo que recomendamos, así
> que sería un dato que el equipo escribe a mano. Si prefieren, se omite.

**Q4. ¿El contenido se ingresa en español e inglés, o solo en español?**
- X **Español obligatorio e inglés opcional**, como el resto del sitio (**recomendada**: es la regla
  que ya rige para todo el contenido, decisión 8 del roadmap).
- ☐ Solo español.

**Q5. ¿Cada episodio se puede dejar en borrador y publicar después?**
- X **Sí, borrador o publicado, por episodio** (**recomendada**: es la regla que ya usan eventos,
  ministerios y ofrendas).
- ☐ No: lo que se carga se ve al instante.

**Q6. ¿Se puede eliminar un episodio?**
X Sí, con confirmación (**recomendada**) · ☐ No, solo se retira a borrador

---

## 4. Preguntas de la página pública

**Q7. ¿Cómo se llama la sección y en qué dirección vive?**
Nombre: X **«Medios»** · ☐ «Prédicas y podcasts» · ☐ Otro: ______
Dirección: X **`/medios`** (recomendada) · ☐ `/predicas` · ☐ `/podcasts` · ☐ Otro: ______

**Q8. ¿Lleva su propia entrada en el menú?**
X **Sí**, junto a las demás secciones (barra de escritorio, menú móvil y mapa del pie) ·
☐ No, solo se llega desde el bloque de la portada

**Q9. ¿En qué orden aparecen los episodios?**
- X **Del más reciente al más antiguo** (**recomendada**: en contenido grabado, lo nuevo es lo que
  interesa).
- ☐ Orden manual, que el equipo decide arrastrando o con un número (como ya se hace en ministerios y
  ofrendas).

**Q10. Si todavía no hay nada publicado, ¿qué pasa al abrir la sección?**
- X **Se oculta la entrada del menú y la dirección redirige a la portada** (**recomendada**; es lo que
  ya hacen Ministerios y Ofrendas, para no mostrar páginas vacías).
- ☐ Se muestra la página con un aviso («todavía no hay prédicas publicadas»), sin redirigir (es lo que
  hace Eventos).

**Q11. ¿Hacen falta filtros o búsqueda?**
X No, una lista simple con **«Cargar más»** (**recomendada** para empezar) ·
☐ Sí, filtrar por tipo (prédica / podcast) · ☐ Sí, buscar por predicador o tema · ☐ Otro: ______

**Q12. ¿Quieren un bloque de «Prédicas y podcasts» en la portada?**
La portada ya tiene: identidad, «Lo que viene» (eventos), quiénes somos, ministerios y ofrendas.
- X **Sí, al final, después de ofrendas**, con los últimos 5 y un botón «Ver todo» (**recomendada**).
- ☐ No, solo la sección propia.

---

## 5. Preguntas del panel y de los enlaces

**Q13. El módulo del panel se llamará «Medios» y usará el permiso que ya existe. ¿De acuerdo?**
X Sí (**recomendada**) · ☐ Otro nombre: ______

**Q14. ¿Quién puede publicar episodios?**
X Los roles que ustedes creen con el permiso «Medios» (es como funcionan los demás módulos) ·
☐ Solo el rol de administrador

**Q15. Al reproducir dentro del sitio se carga contenido de YouTube o Spotify, que son servicios de
terceros (usan sus propias cookies). ¿Están de acuerdo?**
- ☐ **Sí, con el modo de privacidad de YouTube** (que no deja cookies hasta que se reproduce) y un
  aviso claro (**recomendada**).
- X No: preferimos que el visitante abra el video en YouTube/Spotify con un enlace (sería la Opción A).

**Q16. Si un video se borra o se vuelve privado en YouTube, ¿qué debería pasar?**
- X **Se oculta solo de la página pública y el panel lo marca para que el equipo lo revise**
  (**recomendada**).
- ☐ Se sigue mostrando igual, con un aviso de que no está disponible.

**Q17. ¿Solo enlaces de YouTube y Spotify, o también de otros servicios (Vimeo, Facebook, Instagram)?**
X **YouTube y Spotify se reproducen dentro del sitio; los demás se muestran como enlace**
(**recomendada**) · ☐ Solo YouTube y Spotify, nada más

---

## 6. Lo que NO incluye F9 (para que no haya sorpresas)

- **Publicación automática** (Opción C): queda como funcionalidad futura, después del despliegue.
- **Formulario de inscripción** o comentarios en los episodios.
- **Descargar** el audio o el video, y **transcripciones** o subtítulos.
- **Noticias y galería** de eventos pasados: eso es F8, otra funcionalidad.

---

## 7. Qué pasa después de responder

1. Actualizamos la **especificación** con sus respuestas y se las enviamos para que la **aprueben**
   (ahí pueden corregir cualquier punto).
2. Con la especificación aprobada, el equipo diseña la solución técnica y se la presentamos para una
   **segunda aprobación**.
3. Se construye, se prueba y se les entrega un Pull Request con las pruebas y los resultados.
4. **Nada se integra ni se publica sin su aprobación.**
