# Research — F9 Medios: prédicas y podcasts

**Fecha**: 2026-10-17 · **Spec**: [spec.md](./spec.md) (v2, aprobada por el humano el 2026-10-17;
decisión 27 del roadmap) · **Plan**: [plan.md](./plan.md)

> **Notas de partida.** (1) La spec v2 ya cerró el modelo: importación desde dos feeds con botón
> «Buscar episodios nuevos», una entrada por sermón (video + audio unidos), enlaces externos sin
> contenido de terceros, panel reducido y estado visible/oculto. **Nada de eso se reabre aquí.**
> (2) Donde se reutiliza una decisión ya tomada se cita (R3-x de F3, R4-x de F4, R6-x de F6,
> R7-x de F7) y **no se reabre**. (3) Los puntos donde la spec **se contradice** o **no llega**
> están marcados como tales (R9-7, R9-8, R9-14) y recogidos en «Preguntas para el humano» de
> [plan.md](./plan.md): se aplican los defaults indicados y se elevan al humano, **no se inventan
> reglas de negocio**. (4) Toda afirmación de este documento lleva **evidencia**: XML real
> capturado el 2026-10-17, pruebas HTTP contra los servicios reales, código del repo o fuentes
> públicas con enlace.

Este documento resuelve las decisiones técnicas abiertas de F9 con alternativas y motivos. La
numeración es **R9-1…R9-14**, propia de esta funcionalidad, para no colisionar con R7-1…R7-16 de
F7, R6-1… de F6, R4-1… de F4 ni R3-1… de F3.

Contexto reutilizado sin cambios: capas `handler → service → repository` y reglas R1–R8 de
`docs/tecnico/arquitectura.md`; sobre de éxito = DTO directo / `{items}` en listados (§8.1.2);
paginación `platform/paginate` (§8.1.3); `ORDER BY` determinista (§8.1.7); validación
`platform/validate`; autorización `middleware.AdminChain`/`AuthzByModule`; auditoría en
`admin_actions` con `InsertAdminAction` transaccional (R7-10); bilingüismo por columnas
`*_es`/`*_en` con resolución `?lang=` en el servidor (R3-1/R3-2, P7-9); `sort_order`/desempate por
`id` (patrón `000005`/`000007`/`000010`); sección sin contenido → **redirección a la portada**
(R6-13/P6-21/P7-21); bloque de portada con endpoint propio y tope fijo en el servidor (R6-14,
P7-14); imágenes servidas por el dominio `media` con `IsFilePublished` (R4-6/R6-7).

---

## R9-1. ¿Los videos «no listados» (unlisted) de YouTube aparecen en el feed Atom del canal?

**Pregunta.** Si un video *unlisted* entrara en `youtube.com/feeds/videos.xml?channel_id=…`, F9 lo
importaría y lo publicaría en el sitio, con lo que un enlace pensado para circulación restringida
acabaría en la web de la iglesia. Hay que saber si el feed lo filtra.

**Evidencia.** (a) **Fuentes públicas consultadas el 2026-10-17** (3 de 3 independientes entre sí
coinciden):

- *WPRSS Aggregator — «YouTube RSS Feed: URL Format, Examples & Import Notes»*
  (finder.wprssaggregator.com/rss-feeds/platform/youtube): «Private and unlisted videos never
  appear in the feed»; los feeds exponen «roughly the most recent 15 uploads».
- *HushFlow — «YouTube Channel RSS Feeds (2026)»* (hushflowai.com/youtube-channel-rss-feed): el
  feed trae «the channel's 15 most recent **public** uploads… It does not include members-only or
  unlisted videos».
- *LinkMyPrompt — «Convert YouTube Channel to RSS Feed»* (2026-06-12): «Public videos only.
  Unlisted and private videos never show up».

(b) **Una fuente discrepa**: *Kompozy — «YouTube RSS automation»* (2026-05-06) afirma que «un
video puede pasar brevemente por el estado unlisted durante la edición y el feed puede dispararse
con esa publicación», y recomienda filtrar por estado público. Es la única afirmación en este
sentido encontrada; no está respaldada por documentación de YouTube ni por otras fuentes.

(c) **No se pudo comprobar con el feed real desde este entorno**: las 3 peticiones a
`https://www.youtube.com/feeds/videos.xml?channel_id=UCXXJbgowcGnjI7BDsPnKOTw` (2026-10-17)
devolvieron **404**, cosa esperable —la misma fuente WPRSS advierte que «YouTube can still return
500/404 to server-side requests from datacenter IPs (Vercel, AWS) regardless of channel state» y
que el endpoint tuvo «intermittent 404s starting December 2025»—. El equipo **sí** verificó el
feed el 2026-10-17 desde su entorno (9 entradas, verificado en la spec, §Datos de origen), y el
feed del podcast respondió bien desde aquí (R9-5), de modo que el 404 es del origen, no del
procedimiento.

**Decisión.** **El feed es la fuente de verdad**: F9 importa lo que el feed publique y lo trata
como contenido público (es literal FR-010/FR-003: «los datos publicados DEBEN venir de los feeds»).
Se da por **cerrado** que el feed del canal no incluye *unlisted* (evidencia 3:1 + interés
comercial de YouTube: el feed es la superficie pública de suscripción). **Pero no queda cerrado
como imposible** —la fuente discrepante no se pudo refutar experimentalmente desde aquí—, así que
la mitigación es de diseño y cuesta cero:

1. Todo lo importado entra **visible pero reversible**: el panel tiene «ocultar» con un clic
   (FR-015/FR-019) y la sincronización nunca revierte esa decisión.
2. El flujo de F9 es **«el equipo pulsa y revisa»** (US2/US3): la búsqueda deja un resumen y las
   marcas, y el listado del panel muestra origen y enlaces de cada entrada para que la revisión sea
   inmediata.
3. Si el cliente detectara un video que no quiere publicar: ocultar (2 segundos) y queda resuelto
   para siempre (no se reimporta: su identificador externo ya existe).

**Alternativas descartadas.** Filtrar por estado de privacidad vía YouTube Data API (exige clave y
gestión de cuota: viola «gratuito, sin claves ni pagos» de FR-010); comprobar el video contra la
página pública del canal (HTML de terceros, no es un contrato); no importar videos más nuevos que
X minutos (heurística sin garantía). **Señalado a la puerta 2** como incógnita no cerrable desde
el equipo.

---

## R9-2. Comprobación de disponibilidad de un enlace: gratis, sin claves y sin cuota

**Pregunta.** FR-017 exige comprobar si «el enlace concreto sigue disponible» para cada elemento
que sale de la ventana del feed, «una comprobación gratuita y sin claves», y aplicar tres casos
(disponible / no disponible / no se pudo comprobar) con regla **a prueba de fallos** (nunca se
oculta por falta de certeza).

**Evidencia (pruebas HTTP reales el 2026-10-17).**

| Comprobación | URL de prueba | Resultado |
|---|---|---|
| YouTube oEmbed, video **vivo** | `https://www.youtube.com/oembed?url=https://www.youtube.com/watch?v=dQw4w9WgXcQ&format=json` | **200** con JSON (`title`, `author_name`, `thumbnail_url`, `html`…) |
| YouTube oEmbed, video **inexistente** (formato correcto, 11 caracteres) | `…watch?v=abcdefghijk…` | **404** |
| YouTube oEmbed, id mal formado | `…watch?v=zzzzzzzzzzzzz…` | **400** |
| Spotify oEmbed, episodio **inexistente** | `https://open.spotify.com/oembed?url=https://open.spotify.com/episode/0000000000000000000000` | **404** |
| Spotify oEmbed, enlace **real de nuestros episodios** (`podcasters.spotify.com/…`) | `…oembed?url=https://podcasters.spotify.com/pod/show/iglesia-simiente-santa/episodes/Potencial-sin-Resultados-…-e3jsj3n` | **504** en los dos intentos: el oEmbed de Spotify **no resuelve** enlaces `podcasters.spotify.com` |
| Página del episodio **real** (`podcasters.spotify.com`) | `…/episodes/Potencial-sin-Resultados-La-Leccin-de-la-Higuera-e3jsj3n` | **200** (HTML del episodio) |
| Página del episodio **inexistente** (`podcasters.spotify.com`) | `…/episodes/episodio-que-no-existe-abcdef123` | **404** |

**Decisión.** Dos probadores por proveedor, sin claves ni cuota (son los mismos endpoints que
usa cualquier lector de feeds):

- **Video (YouTube)**: `GET https://www.youtube.com/oembed?url=<url-del-video>&format=json`.
  `200` → **disponible**; `404` → **no disponible** (confirmado); **cualquier otra cosa**
  (`400`, `401`, `403`, `429`, `5xx`, timeout, error de red) → **no se pudo comprobar**.
- **Audio (Spotify for Creators / Anchor)**: `GET` (o `HEAD`) sobre el **enlace del episodio**
  devuelto por el feed (`podcasters.spotify.com/pod/show/…/episodes/…`). `200` → disponible;
  `404`/`410` → no disponible (confirmado); todo lo demás → no se pudo comprobar. *El oEmbed de
  Spotify NO sirve para estos enlaces (504 verificado arriba): se descarta como probador.*

**Coste**: **una petición HTTP por elemento afectado y por búsqueda** —solo a los elementos
importados cuyo identificador ya no aparece en la ventana de su feed y cuyo feed **sí** respondió
(en un catálogo de ~15 elementos de ventana, lo habitual son 0–2 comprobaciones por búsqueda;
tope duro de 30 por búsqueda, constante revisable). Nunca se comprueba lo que sigue en el feed.

**Comportamiento ante error** (la regla a prueba de fallos de FR-017): solo el resultado
**«no disponible confirmado»** (`404`/`410`) oculta la ficha y la marca «no disponible»;
**todo lo demás** —incluido timeout, `5xx`, `429` o error de red— deja la ficha **visible** y
marca para revisión, y el resumen informa de que la comprobación no se pudo completar
(`availabilityInconclusive`). Un probador caído **nunca** vacía el sitio.

**Alternativas descartadas.** `HEAD` sobre `youtube.com/watch?v=…` (YouTube responde 200 con
página de error cuando el video no está: no distingue); oEmbed de Spotify para el audio (504
verificado con nuestros enlaces); YouTube Data API `videos.list` (clave + cuota: viola FR-010);
interfaz de Spotify para creadores (desde el 9-mar-2026 exige suscripción de pago al dueño de la
aplicación —decisión 26 del roadmap—); no comprobar nada y ocultar al salir del feed (la regla
del «estante de 15» que el humano **rechazó** antes de la puerta 1: vaciaría el sitio solo).

---

## R9-3. Parser RSS/Atom en Go: **propio con `encoding/xml`**, 0 dependencias nuevas, detrás de una interfaz

**Pregunta.** ¿Qué parser usar para el feed Atom de YouTube y el RSS 2.0 del podcast: una
librería nueva (p. ej. `github.com/mmcdole/gofeed`) o parsing propio con la librería estándar?

**Decisión.** **Parsing propio con `encoding/xml`** (librería estándar de Go), con **dos
estructuras pequeñas —una por formato—** detrás de la interfaz del dominio `FeedSource`, y con
fixtures XML reales en las pruebas (capturados de los feeds del cliente, ver R9-5 y el paso 0 del
[quickstart](./quickstart.md)). Motivos:

1. **Son dos formatos fijos y conocidos** (no un agregador universal): Atom de YouTube (etiquetas
   `yt:videoId`, `media:group`/`media:thumbnail`/`media:description`) y RSS 2.0 de Anchor
   (`item/title`, `description`, `link`, `guid`, `pubDate`, `enclosure`, `itunes:image`). Las
   estructuras necesarias caben en ~80 líneas de structs con tags.
2. **0 dependencias nuevas** (constitución §II, precedente literal de F7/R7-2 con el registro
   IBAN): `gofeed` trae arrastrados `goquery` y `golang.org/x/net` y ofrece universalidad (RSS
   0.9x/1.0/2.0, Atom, JSON Feed, detección de extensiones) que F9 **no** necesita.
3. **Control total de los bordes peligrosos**: `encoding/xml` descodifica entidades y CDATA de
   forma segura, **no ejecuta nada**, y nos permite (a) **no conservar HTML** (las descripciones
   del podcast vienen con etiquetas `<p>…` dentro de CDATA — verificado en R9-5 — y F9 solo
   publica texto plano, FR-023/constitución §IV), (b) validar los enlaces contra los hosts
   esperados antes de guardarlos (R9-8) y (c) parsear fechas con las formas que los feeds usan de
   verdad (`GMT` y desplazamientos numéricos: R9-5).
4. **Aislado tras una interfaz, probable sin red**: `FeedSource` (con `FetchYouTube(ctx)` y
   `FetchPodcast(ctx)`) y su fake `fake_feed_source_test.go` replican el patrón que el proyecto ya
   usa —`backend/internal/agenda/fake_repository_test.go` y
   `backend/internal/ministerios/fake_repository_test.go`: fakes en memoria en el mismo paquete de
   pruebas—. El resto del dominio (importación, unión, marcas, disponibilidad) se prueba a ciegas
   del parseo, y el parser se prueba **solo contra fixtures XML** (los reales capturados + casos
   patológicos: CDATA, emojis, tildes, campos ausentes, fechas `GMT` y `+0000`, XML roto → error
   claro, nunca importación parcial de un item corrupto).

**Detalles de implementación que la evidencia fija:**

- `encoding/xml` casa por **nombre local** (`xml:"videoId"` con el espacio de nombres de YouTube)
  y tolera la ausencia de campos opcionales (`media:description` puede no venir; `itunes:image`
  puede no venir → ficha sin carátula → marcador neutro, default revisable #4).
- Fechas: `time.Parse` con `time.RFC1123` (acepta `GMT` como zona) y, si falla, `time.RFC1123Z`
  (desplazamiento numérico); el RSS del podcast trae `pubDate` en `GMT` (verificado) y el Atom de
  YouTube trae `published` RFC3339 (`2026-05-25T…Z`). Una fecha imposible de parsear **descarta el
  item con aviso en el resumen** (sin fecha no hay orden ni unión segura), nunca inventa fecha.
- El HTML de la descripción se convierte a **texto plano** en el importador: quitar etiquetas,
  decodificar entidades (`&quot;`…), colapsar espacios. Se guarda y publica **solo texto plano**
  (FR-023: los datos nunca se ejecutan ni insertan elementos activos; el frontend no usa
  `dangerouslySetInnerHTML`).

**Alternativas descartadas.** `gofeed` (motivos 1–2); `encoding/xml` + XPath/XQuery (no existe en
la stdlib); regex sobre el XML (frágil y con riesgo de ReDoS); parsear el HTML de las páginas de
YouTube/Spotify (no son contratos, cambian sin aviso).

---

## R9-4. Cómo se normaliza y se une el mismo sermón (título + fecha)

**Pregunta.** FR-013 pide unir automáticamente el video y el audio del mismo sermón cuando «los
títulos coinciden al normalizar mayúsculas, puntuación y espacios» y «fueron publicados el mismo
día», y FR-014 manda marcar lo dudoso. Los títulos reales **no** coinciden carácter por carácter.

**Evidencia (datos reales).** YouTube: «Potencial sin Resultados La Lección de la Higuera» ·
podcast: «Potencial sin Resultados: La Lección de la Higuera» (capturado del RSS el 2026-10-17,
`<item>…<guid>bc42e0cd-d585-40ab-ac37-b67695a1d545</guid><pubDate>Mon, 25 May 2026 17:58:50
GMT</pubDate>`). Difieren en los dos puntos y en el espacio: se resuelven con **quitar puntuación,
colapsar espacios y pasar a mayúsculas**.

**Decisión — normalización (`internal/medios/titulo.go`, funciones puras).** Para cada título:

1. **NFKD + quitar marcas diacríticas** con `golang.org/x/text/unicode/norm` (**sin dependencia
   nueva**: `golang.org/x/text v0.42.0` ya está en `backend/go.mod` como indirecta de `pgx`;
   importarla directa solo la promueve de indirecta a directa — verificado en `go.mod`).
   Así «Lección» y «Leccion» coinciden si un feed pierde las tildes.
2. **Mayúsculas** Unicode (`strings.ToUpper`).
3. **Puntuación y símbolos → espacio** (`unicode.IsPunct`, `unicode.IsSymbol`: `: , . ; " ' ¿ ¡
   ! ? - – — ( ) [ ] …`), incluidos los signos del español.
4. **Colapsar espacios** en blanco y recortar.

Resultado verificado con el par real: ambos títulos → `POTENCIAL SIN RESULTADOS LA LECCION DE LA
HIGUERA`. El texto original **nunca se altera** (se publica el título del feed tal cual, FR-003);
la normalización es solo la llave de comparación.

**Decisión — regla de emparejamiento (importación).** Dos elementos de **proveedores distintos**
son **candidatos** si se cumple **cualquiera** de: (a) **mismo día** de publicación interpretado
en **`America/Costa_Rica`** (FR-006/decisión 17 de F4), o (b) **mismo título normalizado** (para
atrapar el caso frontera de R9-14). Entonces:

- **Unión automática (FR-013)**: candidatos con **título normalizado igual** y **mismo día** →
  una sola ficha con los dos enlaces (la del video aporta nombre, carátula y fecha — default
  revisable #7 —).
- **Dudoso (FR-014)**: candidatos que **no** cumplen ambos —p. ej. mismo día con títulos que no
  coinciden, o mismo título con días distintos— → **entradas separadas**, la adicional queda
  **oculta**, **ambas** se marcan «posible duplicado» y se enlazan entre sí (`duplicate_of_id`)
  para que el equipo una con un clic (o resuelva «no son duplicados»: US3 esc. 2).
- **Sin candidatura** → entradas independientes, normales, sin marca.

**Por qué esta regla y no otra.** Es literal de FR-013/FR-014 (la spec fija *exactamente* esta
comprobación: título normalizado + mismo día) y es **conservadora en la dirección correcta**: lo
que no se puede afirmar «con seguridad» no se une y queda **fuera del público** (el único camino
para cumplir SC-003 «0 repeticiones del mismo sermón» sin que el visitante vea dos veces un
sermón). Los falsos positivos de la candidatura por día (dos sermones distintos el mismo día) son
exactamente el caso que la spec prevé en US3 esc. 2: el equipo resuelve la marca y ambos quedan
visibles. Con los feeds reales la regla produce el resultado verificado en la spec: **8 uniones
automáticas y 2 entradas con un solo enlace** en la primera búsqueda (dato de comprobación, no
regla).

**Alternativas descartadas.** Comparar títulos exactos (los reales no coinciden: verificado);
similaridad difusa tipo Levenshtein/Jaccard (umbral arbitrario, falsas uniones que el visitante
nota como «un sermón que cambió de nombre», y la spec no la pide); margen ±1 día en la fecha
(la spec dice «mismo día»; queda como pregunta menor —ver plan §Preguntas— porque un desfase de
frontera entre feeds se resuelve hoy como dudoso + marca, que es seguro); identificador cruzado
entre plataformas (no existe en los feeds: YouTube trae `yt:videoId`, el podcast su `guid` de
Anchor; no hay campo común); unión por descripción (más larga y menos estable que el título).

---

## R9-5. Los dos feeds: estructura real verificada el 2026-10-17

**Pregunta.** ¿Qué trae exactamente cada feed (campos, formatos, identificadores) y qué
sorpresas de forma hay que asumir en el parser y el modelo?

**Evidencia.** **Podcast**: XML completo capturado el 2026-10-17 de
`https://anchor.fm/s/1116351e8/podcast/rss` (RSS 2.0 con espacios de nombres `dc`, `content`,
`atom`, `anchor`, `podcast`, `itunes`, `psc`; título del canal «Servicio dominical Iglesia
Simiente Santa», `generator: Anchor Podcasts`, `language: es-cr`, **9 `<item>`**). Por cada item:

| Campo | Ejemplo real | Observación |
|---|---|---|
| `title` | `<![CDATA[Potencial sin Resultados: La Lección de la Higuera]]>` | CDATA; tildes y `ñ` intactas |
| `description` | `<![CDATA[<p>¿Es posible tener mucha fe…</p><p>…</p>]]>` | **HTML dentro de CDATA**, hasta ~1.500 caracteres con `<p>`, `<ul>`, `<strong>`, `&quot;`… → se convierte a texto plano (R9-3) |
| `link` | `https://podcasters.spotify.com/pod/show/iglesia-simiente-santa/episodes/Potencial-sin-Resultados-La-Leccin-de-la-Higuera-e3jsj3n` | **es el enlace del episodio** («Escuchar en Spotify»); la página responde 200 y la inexistente 404 (R9-2) |
| `guid` | `bc42e0cd-d585-40ab-ac37-b67695a1d545` (`isPermaLink="false"`) | **identificador externo** del elemento de podcast |
| `pubDate` | `Mon, 25 May 2026 17:58:50 GMT` | RFC1123 con **`GMT`** (no desplazamiento numérico) |
| `enclosure` | `https://anchor.fm/s/1116351e8/podcast/play/…mp3` (`audio/mpeg`, ~150 MB) | el audio en sí; F9 **no lo usa** (no aloja ni transmite: FR-001) |
| `itunes:image href` | `https://d3t3ozftmdmh3i.cloudfront.net/staging/podcast_uploaded_nologo/45766890/45766890-1776097995866-18d4d70c5410a.jpg` | **la misma carátula en los 9 episodios**: es el arte del programa, no del episodio (verificado: todos los items repiten `href`) |
| `itunes:duration`, `itunes:summary`, `dc:creator`, `itunes:episodeType` | `01:00:02`, … | no se necesitan; se ignoran |

**YouTube**: el feed **no respondió desde este entorno (404, R9-1evidencia c)**, pero su formato
es público y estándar y la spec verificó su contenido el 2026-10-17 («por cada video aporta
identificador del video, título, enlace a `youtube.com/watch?v=…`, fecha de publicación, carátula
y descripción completa (muy larga)»), que coincide campo a campo con el Atom de YouTube:
`<entry>` con `yt:videoId` (identificador externo), `title`, `link` (`watch?v=…`), `published`
(RFC3339 UTC), `media:group` con `media:thumbnail url` (carátula en `i.ytimg.com/vi/<id>/…jpg`) y
`media:description` (la descripción completa, hasta ~1.500 caracteres). **Consecuencia operativa
para la implementación**: la primera tarea de backend **capturará fixtures reales** desde un
entorno donde el feed responda (el del equipo) y los cometerá en `backend/internal/medios/testdata/`;
el parser se prueba contra esos fixtures y el quickstart §0 documenta el paso. Hasta entonces, el
mapa de campos del feed de YouTube es el formato documentado + los datos verificados en la spec.

**Otras consecuencias verificadas.**

- **Fechas en el límite del día** (edge case de la spec, ahora con caso real): el episodio «La Fe
  y la confianza en el Abba» trae `pubDate = Fri, 03 Jul 2026 01:22:48 GMT`, que en
  `America/Costa_Rica` (UTC−6) es **2 de julio a las 19:22**. Por eso la comparación de «mismo
  día» de R9-4 y el orden de FR-006 se calculan **siempre** en la zona de la iglesia, no en UTC
  (decisión 17 de F4, `APP_TZ`): un «mismo día» calculado en UTC habría separado un par que en
  Costa Rica es del mismo día.
- **Carátula compartida**: las fichas de **solo audio** mostrarán la carátula del programa (misma
  imagen). Es un dato del feed (FR-003) y así se publica; si el cliente quisiera arte por
  episodio, tendría que publicarlo en Spotify (fuera de alcance de F9).
- Los feeds están **congelados desde el 3 de julio de 2026** (dato de la spec, no requisito) y
  ambos son **públicos, gratuitos y sin claves** (cuestionario §2).

---

## R9-6. Qué se guarda del feed y cuándo: **instantánea al importar**

**Pregunta.** FR-012 dice que la búsqueda «solo agrega lo que falte» y «nunca modifica lo que el
equipo decidió»; el default revisable #1 añade que la descripción corta sin ajustar «puede
regenerarse si cambia la descripción de origen»; el default #10 fija que la carátula copiada **no
se actualiza**. ¿La búsqueda refresca datos de los elementos ya importados?

**Decisión.** **Instantánea al importar y nada más**: los datos del feed —título, fecha, carátula,
enlaces y la descripción corta generada— se toman **una sola vez**, en la importación del
elemento, y las búsquedas posteriores **no los tocan** (solo agregan lo que falta y aplican las
marcas/visibilidad de FR-017, que son decisiones del sistema ya fijadas por la spec). Es la
lectura literal de FR-012 («agregar solo lo que falte», «pulsar sin novedades no cambia el
catálogo»), es consistente con el default #10 (carátula congelada) y hace que **SC-007** («0
descripciones que se pierdan, 0 uniones que se deshagan») sea verificable campo a campo.

Consecuencia: el default #1 en su frase «puede regenerarse si cambia la descripción de origen»
queda **sin efecto práctico** (no hay regeneración: la descripción corta solo nace al importar y
después solo la edita el equipo). Lo registramos como **default decidido**, no como olvido: la
alternativa —refrescar los campos no editados en cada búsqueda— existe (es un `UPDATE`
condicionado por `description_edited = false` / `cover_file IS NULL`) pero rompe la «instantánea»,
haría que un título publicado cambiara sin que el equipo lo supiera y complica el resumen y las
pruebas sin que la spec lo pida. Si el humano prefiere refrescar, es un cambio acotado.

**Alternativas descartadas.** Refrescar todo (pisa decisiones del equipo: prohibido por FR-015);
refrescar solo lo no editado (descrito arriba); guardar la descripción completa del feed para
regenerar (FR-020 prohíbe publicarla y sin regeneración no hace falta: no se guarda).

---

## R9-7. Marcas de revisión y su resolución (síntesis de una contradicción de la spec)

**Pregunta.** La spec dice tres cosas que leen en direcciones distintas: (i) FR-009 —«una entrada
pendiente de resolver un “posible duplicado” NO DEBE aparecer en ninguna superficie pública»— y
SC-002 —«0 entradas con marcas sin resolver visibles para el público»—; (ii) el default revisable
#2 —«de un par marcado, **solo la primera entrada importada queda visible**»—; y (iii) FR-017 —«el
enlace sigue disponible → el episodio **sigue visible** para el público y queda una marca “ya no
aparece en el feed”»—, que es explícitamente una entrada **con marca sin resolver y visible**.

**Decisión (única lectura que las cumple las tres).** La marca **no es un estado de publicación**:
la publicación la decide `visibility` (visible/oculto, decisión del equipo) **y** una única marca
bloquea por sí sola la publicidad:

| Marca | ¿Bloquea al público? | Visibilidad |
|---|---|---|
| `possible_duplicate` **sin resolver** | **Sí** (FR-009/SC-002): ni sección, ni portada, ni respuestas | la entrada «primera» conserva `visibility='visible'` (default #2: su **flag** queda visible) y la adicional entra `hidden`; **las dos quedan fuera del público** hasta que se resuelva |
| `unavailable` (no disponible confirmado) | ya está `hidden` (FR-017/Q16: el sistema la ocultó) | el equipo resuelve y decide |
| `missing_from_feed` («ya no aparece en el feed») | **No** (FR-017 literal: «sigue visible») | `visible`; solo es un aviso para el panel |

Es decir: **el predicado público es `visibility='visible'` Y (sin `possible_duplicate` abierto)**.
Así default #2 se cumple literalmente (solo la primera conserva su flag de visibilidad), SC-002 se
cumple (0 marcas «posible duplicado» abiertas en público) y FR-017 se cumple («ya no aparece en el
feed» sí se publica). La alternativa —que **toda** marca abierta bloquee— contradice FR-017 de
forma inequívoca y queda descartada; la alternativa —que el par marcado se publique parcialmente—
incumple SC-002 y FR-009.

**Resolución y no-reapertura.** La resolución es una acción del equipo (`POST
…/{id}/marcas/resolver` con la visibilidad elegida, US3 esc. 6) que fija `mark_resolved_at` y deja
la marca **en el registro** (`review_mark` se conserva, `mark_resolved_at IS NOT NULL`): la lista
del panel puede mostrar «Resuelta · posible duplicado» como dato histórico y, sobre todo, la
sincronización **no reabre una marca de la misma clase** cuyo `mark_resolved_at` no sea nulo
(FR-017/default revisable #6). Si surge una clase de marca **distinta** sobre la misma entrada,
manda la nueva (estado abierto) —son casos de negocio distintos: un duplicado dudoso resuelto
puede después volverse no disponible—. `unir` (US3 esc. 1) **despeja** las marcas de las dos
entradas («la marca desaparece») sobre la entrada resultante y marca la absorbida como resuelta.

**Alternativas descartadas.** Borrar la marca al resolver (se pierde el dato y no se puede
garantizar la no-reapertura: habría que recordarlo en otra tabla); marcas múltiples por entrada
(catalogo de estados compuestos sin pedido en la spec); estado `draft/published` de F3/F4/F6
(la spec lo excluye explícitamente: Consecuencia 5, «F9 no usa borrador/publicado»).

---

## R9-8. La carátula se copia al sitio: descarga saliente segura (primera del backend)

**Pregunta.** SC-005 exige 0 cargas de terceros: la carátula **se sirve desde el sitio**
(«se copia al importar»). Copiarla significa que el backend **descarga bytes de una URL que trae
el feed** —la primera llamada HTTP saliente de todo el backend (verificado: solo las pruebas usan
`http.Get`, `backend/internal/platform/httpserver/server_test.go`)—. ¿Cómo se hace sin abrir
SSRF ni contenido activo?

**Decisión.** Descarga **acotada** en `internal/medios` (función pura sobre un `http.Client`
propio, inyectable para pruebas) que **solo acepta**:

1. **`https`** y **host de allowlist**: `i.ytimg.com` (carátulas de YouTube) y
   `d3t3ozftmdmh3i.cloudfront.net` (carátula del podcast, verificada en R9-5). Lista de
   constantes documentada y fácil de ampliar si el cliente cambia de CDN (si un feed trae otra
   cosa: **sin carátula**, marcador neutro; nunca se descarga de un host no listado).
2. **Rechazo de direcciones privadas/loopback** tras resolver DNS (defensa en profundidad frente a
   un feed comprometido que apunte a `169.254.169.254` o a la red interna) y **sin seguir
   redirecciones** fuera de la allowlist (se comprueba cada `Location`).
3. **Tamaño acotado** con `io.LimitReader` (tope: el mismo `UPLOAD_MAX_BYTES` que ya usa
   `platform/storage`, configuración existente) y **validación de firma** con
   `storage.Detect` (solo JPEG/PNG/WebP: SVG, GIF, HTML renombrado → rechazados) **dentro de
   `Store.Save`** —mismo `platform/storage` de F3/F4/F6: nombres `img_<uuid>.<ext>` generados por
   el servidor, escritura atómica en disco—.
4. **Timeout propio** (10 s por descarga) y máximo 30 carátulas por búsqueda.

Un fallo de descarga **no** tira la importación: el elemento entra **sin carátula** (marcador
neutro, default revisable #4) y el resumen lo cuenta. El archivo copiado se sirve por el dominio
`media` existente (`GET /api/v1/media/{fileName}`), que solo entrega archivos referenciados por
contenido **publicado** (`IsFilePublished`, ampliado con la rama de F9 — ver data-model).

**Alternativas descartadas.** Enlazar la carátula directamente a `i.ytimg.com`/CloudFront
(viola SC-005: cargaría recursos de terceros en el navegador del visitante, que es exactamente lo
que Q15 prohíbe); copiar con `http.Get` a pelo (sin allowlist, sin límites: SSRF y denegación de
servicio); servir el SVG/HTML que devuelva el feed (contenido activo: prohibido, constitución
§IV); bibliotecas de descarga con caché (dependencia nueva sin justificación).

---

## R9-9. La búsqueda cabe en el ciclo petición/respuesta (sin proceso en segundo plano)

**Pregunta.** Constitución y spec: **nada de cron, worker, cola ni tarea de fondo** (FR-011, Out
of Scope). Pero una búsqueda real toca la red: 2 feeds + hasta ~30 carátulas + hasta ~30
comprobaciones de disponibilidad. ¿Cómo se garantiza que la petición HTTP del panel termine?

**Decisión.** La búsqueda es **síncrona** dentro de `POST /api/v1/admin/medios/buscar-nuevos`, con:

- **Presupuesto total** de **90 s** sobre el `context` de la petición
  (`context.WithTimeout`), constante revisable. Al agotarse, la búsqueda **se cierra con lo
  hecho** y el resumen lo dice (`availabilityInconclusive`, conteos parciales): jamás se
  bloquea la respuesta ni se deja trabajo a medias en la BD (cada elemento se guarda en su propia
  transacción corta).
- **Timeouts por petición saliente**: feeds 10 s · carátulas 10 s · probadores 5 s.
- **Paralelismo acotado**: carátulas y probadores con máximo **4 en vuelo** (`golang.org/x/sync`
  ya está en `go.mod` como indirecta de `testcontainers`; promoverla a directa **no** es
  dependencia nueva). Los 2 feeds se leen en paralelo.
- **Topes duros por búsqueda** (constantes revisables): 30 elementos nuevos, 30 carátulas, 30
  comprobaciones de disponibilidad. Con las ventanas reales (~15 + ~9 elementos) una búsqueda
  normal hace 2 peticiones de feed + 0–2 comprobaciones y, solo la primera vez, ~15 descargas de
  carátula: del orden de **5–20 s**.
- **UI**: el botón se desactiva mientras corre, muestra «Buscando…» y al terminar pinta el
  resumen (FR-018); el contrato documenta que la operación puede tardar.

Un feed que no responde **no** espera al presupuesto completo: falla rápido (timeout propio) y el
otro se procesa igual (FR-016). **Alternativa descartada**: cualquier forma de proceso en segundo
plano (prohibido por la spec y por la constitución, y exigiría despliegue/operación que F9 no
tiene); bajar los topes para «entrar siempre en 10 s» (empeoraría la primera importación sin
necesidad: el presupuesto actual es holgado).

---

## R9-10. Sección pública: lista con «Cargar más» servida por el servidor

**Pregunta.** FR-006/Q11 piden «lista simple con «Cargar más»» y el default revisable #5 fija
tandas de 12. ¿El endpoint devuelve todo (como F7, `maxPublicItems = 200`) y el cliente trocea, o
hay paginación real?

**Decisión.** **Paginación real por `limit`/`offset`** (reutilizando los parámetros estándar del
contrato y `platform/paginate`, §8.1.3): `GET /api/v1/medios?lang=&limit=&offset=` con
**`limit` por defecto 12** (la tanda del default #5) y tope 50; el sobre devuelve
`{lang, items, total, limit, offset}` y el botón «Cargar más» pide la siguiente tanda y **suma**
al listado. Motivos: es literal el «Cargar más» de Q11; cumple la convención de casa «todo listado
se acota» sin inventar una constante que **recortaría** el catálogo (el default «sin límite
artificial de episodios» de la spec choca con cualquier `maxPublicItems`); y con ~15–50 episodios
año no hay problema de rendimiento (índice por `published_at`).

El bloque de la portada mantiene el patrón de F6/F7: **tope 5 firme del servidor**, sin parámetro
de límite, mismo esquema de ítem, clave `items` **omitida** si no hay nada visible.

**Alternativas descartadas.** Todo en una respuesta con tope 200 (recorta el catálogo: la spec
pide sin límite artificial y «Cargar más» implica tandas); infinite scroll con cursor (la spec dice
botón «Cargar más» y el patrón del sitio es explícito); paginación con `page` (el contrato de casa
usa `limit`/`offset`, §8.1.3).

---

## R9-11. Identificador externo por proveedor e idempotencia de la búsqueda

**Pregunta.** Cómo garantizar que repetir «Buscar episodios nuevos» nunca importa dos veces el
mismo elemento (FR-012) ni rompe una unión ya hecha.

**Decisión.** Tabla **`media_episode_sources`** con **una fila por elemento de feed importado** y
**`UNIQUE (provider, external_id)`** (R9-5 fija los identificadores: `yt:videoId` para YouTube y
`guid` para el podcast). La importación de cada elemento es `INSERT … ON CONFLICT
(provider, external_id) DO NOTHING`: lo ya visto se **ignora** (no actualiza nada: R9-6) y lo
nuevo entra. La ficha (`media_items`) agrupa 1–2 fuentes y su `origin` (`video`/`audio`/`both`)
se mantiene en la misma transacción. La unión mueve **fuentes** entre fichas, de modo que el
identificador del elemento absorbido queda registrado en la ficha resultante y **no puede
reimportarse** jamás. La base de datos —no el código— es la garantía final de unicidad, incluso
con dos búsquedas simultáneas (default revisable #12: dos personas pulsando a la vez; el
`UNIQUE` hace que la segunda no duplique).

**Alternativas descartadas.** Recordar lo importado en memoria o en un JSON (se pierde al
reiniciar); comparar por título/fecha (los títulos cambian: verificado); unicidad solo en el
servicio con `SELECT` previo (condición de carrera entre dos búsquedas); identificador compuesto
dentro de `media_items` (columnas `youtube_id`/`podcast_id` con `UNIQUE` cada una: viable pero
fuerza `NULL`s cruzados y hace que la unión reescriba la ficha en vez de mover una fila con su
procedencia — `feed_title`, fecha de importación—, que el panel necesita para revisar).

---

## R9-12. Unir duplicados: **ficha absorbida con puntero**, sin borrados

**Pregunta.** FR-019/Out of Scope: «**NO DEBE** haber borrado de episodios… un elemento importado
que no debe verse se oculta». Pero unir dos entradas deja una sola ficha: ¿qué pasa con la fila
sobrante?

**Decisión.** Unión **por puntero**: la entrada absorbida queda con
`merged_into_id = <id de la ficha resultante>`, `visibility='hidden'` y sus marcas resuelta, y
**desaparece de todas las consultas** (públicas y de panel: `merged_into_id IS NULL`); sus
`media_episode_sources` **se repuntan** a la ficha resultante (que queda `origin='both'` y con los
dos enlaces). Así F9 **no tiene ni un solo `DELETE`** —literal FR-019—, la unión conserva la
procedencia completa (los dos títulos del feed quedan en las fuentes, para siempre visibles en el
panel) y la unicidad de los identificadores externos sigue garantizando que nada se reimporta. La
unión es **definitiva** (default revisable #8: confirmación explícita y sin deshacer; la
interfaz no ofrece deshacer, aunque el puntero haga teóricamente posible una recuperación manual).
La carátula propia de la ficha absorbida, si la tiene y no se usa, se **borra del almacén** (un
archivo huérfano no es «un episodio»; `storage.Delete` es idempotente) con la transacción en
mejor esfuerzo.

**Regla de la unión** (default #7 + coherencia): la ficha **resultante** conserva nombre,
carátula, fecha y descripción de la **entrada del video** (que es la que trae nombre y carátula
según el default #7) y adopta el enlace de audio de la otra; si **solo una** de las dos tenía la
descripción ajustada por el equipo, manda la ajustada (protección de FR-015/SC-007). El diálogo
de confirmación muestra el resultado antes de aplicarlo (nombre, carátula y enlaces finales).

**Alternativas descartadas.** `DELETE` de la fila vacía (incumple literalmente «no hay borrado»;
el propio Out of Scope lo explica: aquí **no** reaparecería en la siguiente búsqueda, pero la
regla es la regla); copiar el audio a la ficha base y dejar la otra «muerta» sin puntero (se
pierde la procedencia y el panel no podría mostrar qué se unió); unión reversible con historial
(default #8: la unión es definitiva; el historial está fuera de alcance).

---

## R9-13. Dependencias: **cero módulos nuevos**

**Pregunta.** Qué necesita F9 que no esté ya en `backend/go.mod` y en `frontend/package.json`.

**Decisión y verificación sobre el código real** (`backend/go.mod`, `frontend/`, revisado el
2026-10-17):

| Necesidad | Solución | ¿Dependencia nueva? |
|---|---|---|
| Parsear Atom/RSS | `encoding/xml` (stdlib) — R9-3 | **No** |
| Normalizar títulos (NFKD) | `golang.org/x/text/unicode/norm`; `golang.org/x/text v0.42.0` **ya está** en `go.mod` como indirecta de `pgx` | **No** (pasa de indirecta a directa) |
| Paralelismo acotado | `golang.org/x/sync` (ya en `go.mod` como indirecta); o goroutines + `sync.WaitGroup` de la stdlib | **No** |
| HTTP saliente (feeds, carátulas, probadores) | `net/http` con `http.Client` propio | **No** |
| Guardar carátulas | `platform/storage` + `storage.Store` de F3/F4/F6 | **No** |
| Fechas | `time` (RFC1123/RFC1123Z/RFC3339) + `APP_TZ` | **No** |
| Tipos/contratos frontend | generación existente (`make api-gen`) | **No** |

**Ningún módulo nuevo ni en backend ni en frontend.** Las únicas «altas» son dos constantes
promovidas de indirectas a directas, sin cambiar versiones ni `go.sum` más allá del acuse.

---

## R9-14. Huecos de la spec detectados y cómo se cubren (resumen)

1. **Marcas visibles/ocultas para el público** (contradicción FR-009/SC-002 ↔ default #2 ↔
   FR-017): resuelta en R9-7; **se eleva a la puerta 2**.
2. **Q16 con una ficha de dos enlaces**: la spec dice que un enlace confirmado no disponible
   oculta **el episodio** («el episodio desaparece de la página pública»). Si solo uno de los dos
   enlaces murió, **se oculta la ficha entera** (lectura literal) y la marca permite al equipo
   reactivarla. La alternativa —quitar solo el enlace roto y dejar la ficha con el otro— es un
   cambio de alcance (tocaría FR-003/FR-017/SC-014): **pregunta al humano**.
3. **Descripción corta cuando el feed no trae descripción** (FR-020 dice «cada episodio DEBE
   tener»): default = el campo queda **ausente** y la ficha **omite la línea** (patrón US1 esc. 3/
   SC-010 «sin líneas vacías»); el equipo puede escribirla después. Con los feeds reales no ocurre
   (todos los items traen descripción: verificado). **Pregunta menor al humano**.
4. **Frontera del «mismo día»** entre feeds: si el video cae a las 23:50 UTC de un día y el audio a
   las 01:22 UTC del siguiente (mismo día en Costa Rica, o días distintos según el cálculo), la
   unión automática **no** se dispara y el par entra como dudoso + marcas (seguro y reversible).
   Si el humano quiere margen ±1 día, es una constante. **Pregunta menor**.
5. **Desempate del orden con misma fecha** (default revisable #11): se aplica literal «orden de
   importación» = **FIFO** (`created_at` ascendente, luego `id`), con `ORDER BY published_at DESC,
   created_at ASC, id`. Cambiar la dirección del desempate es una línea.
6. **Carátula del podcast compartida entre episodios** (R9-5): es lo que trae el feed; no es un
   defecto de F9. Se informa al cliente.
