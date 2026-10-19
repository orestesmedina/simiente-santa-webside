# Quickstart — Cómo probar F9 (Medios: prédicas y podcasts) en local

**Fecha**: 2026-10-17 · **Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md) ·
**Contrato**: [contracts/openapi.yaml](./contracts/openapi.yaml)

Recorrido manual de las US1–US6 y sus criterios de aceptación (el bloque de la portada está en
§10), con `curl` para la API y pasos de navegador para lo visual. Mismo estilo que el quickstart
de F7. Al final, §13 ejecuta las pruebas automatizadas y §14 mapea cada criterio a su sección.

Los casos que la spec hace críticos tienen sección propia: **forzar una búsqueda** (§3, SC-004),
**la unión de duplicados sin depender de la red** (§5, SC-003), **la regla del «estante de 15»**
(§8, SC-014: lo que sale del feed no se oculta, y solo un enlace confirmado muerto oculta), **la
protección de lo del equipo** (§6–§7, SC-007) y **la redirección de `/medios` sin nada publicable**
(§11, SC-016).

## 0. Preparar el entorno

```bash
make up                      # docker compose up -d (db, redis, backend, frontend)
make db-migrate              # aplica 000014 y 000015 sobre la BD levantada
make api-gen                 # regenera frontend/src/api/schema.d.ts (contrato 0.9.0)
cd frontend && npm install && npm test -- --run
```

**Sin variables de entorno nuevas** (F9 no añade ninguna). Tras cambiar código:
`docker compose up -d --build` (§8.1.8).

Comprobación de siempre: `curl -i $API/healthz` → `200` (intacto, F1).

```bash
export API=http://localhost:8080
export WEB=http://localhost:5173
```

### 0.1 Capturar los fixtures de los feeds (una vez, antes de las pruebas de backend)

El parser se prueba contra **fixtures XML reales** (research R9-3/R9-5). Se capturan desde un
entorno donde los feeds respondan (el feed de YouTube puede devolver 404 desde IPs de datacenter:
research R9-1) y se commitean como datos de prueba:

```bash
curl -s 'https://www.youtube.com/feeds/videos.xml?channel_id=UCXXJbgowcGnjI7BDsPnKOTw' \
  > backend/internal/medios/testdata/youtube_feed.xml
curl -s 'https://anchor.fm/s/1116351e8/podcast/rss' \
  > backend/internal/medios/testdata/podcast_feed.xml
cd backend && go test ./internal/medios/...
```

Con los feeds del cliente tal como estaban el 2026-10-17 (9 y 9 elementos, 8 en común) estos
fixtures contienen el par de títulos que la unión debe resolver («Potencial sin Resultados La
Lección de la Higuera» / «Potencial sin Resultados: La Lección de la Higuera»).

## 1. La sección pública sin cuenta (US1, FR-002…FR-006, SC-001/SC-005/SC-006)

```bash
curl -s "$API/api/v1/medios?lang=es&limit=12&offset=0" | jq
```

- `200` con `{"lang":"es","items":[…],"total":N,"limit":12,"offset":0}` **sin autenticación** y
  con `Cache-Control: no-store`.
- Cada ficha trae `id`, `title` (el del feed), `publishedAt`, `description` (corta) y, según
  existan, `coverUrl` (ruta relativa `/api/v1/media/<file>`), `videoUrl` (YouTube) y `audioUrl`
  (Spotify for Creators/Anchor). **Las claves que no aplican se omiten** (una ficha de solo audio
  no trae `videoUrl`: sin líneas vacías, SC-010).
- Con la base recién migrada verás `items: []` y `total: 0`: es correcto (el contenido entra por
  la búsqueda de §3) y así se comprueba la regla de clave omitida.
- **Ninguna respuesta pública contiene** `visibility`, `reviewMark`, `sources` ni fichas ocultas,
  absorbidas o con «posible duplicado» sin resolver (SC-002).

En el navegador: **`$WEB/medios`** es la sección «Medios» (una entrada de navegación en barra,
menú hamburguesa y mapa del sitio, FR-007). Con 0 publicables **no verás la entrada** y la URL
escrita a mano **redirige a la portada** (`$WEB/`) sin destello ni aviso (Q10, patrón de F6/F7):
verás la portada completa y «Atrás» no te devuelve a la URL muerta.

Comprobación de privacidad (SC-005) en las DevTools del navegador, pestaña **Red**, abriendo
`$WEB/medios` y la portada: **0 peticiones** a `youtube.com`, `spotify.com`, `anchor.fm`,
`i.ytimg.com` ni `cloudfront.net`. Las carátulas que veas son `/api/v1/media/…` (del sitio).

Comprobación de enlaces (SC-006): cada ficha ofrece «Ver en YouTube» y/o «Escuchar en Spotify» con
texto que **anuncia la salida del sitio y la pestaña nueva** (FR-005); al activarlos se abre el
servicio externo en pestaña nueva y la del sitio sigue abierta.

## 2. Rol y cuentas con el permiso «medios» (US2, FR-021, SC-009)

El permiso `medios` ("Medios (prédicas y podcasts)") **ya existe** en el catálogo de F2
(migración `000002`): F9 **no** crea permisos nuevos y el catálogo sigue en **10**. El
administrador inicial ya lo tiene y el formulario de roles lo ofrece sin el sello «Disponible más
adelante». Con una cuenta de administración de usuarios y roles:

```bash
# Login y cookies (flujo de F2). Credenciales de las pruebas e2e (frontend/e2e/helpers.ts).
curl -s -c /tmp/ss.jar -X POST "$API/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@ejemplo.com","password":"Semilla.2026"}' | jq
export CSRF=$(grep csrf_token /tmp/ss.jar | awk '{print $7}')

# Dos cuentas de prueba: solo medios y sin permisos de F9.
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"gestores-medios","permissions":["medios"]}' | jq
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"sin-acceso-f9","permissions":[]}' | jq
```

Crea una cuenta por rol con `POST /api/v1/admin/usuarios` (ver quickstart de F3 §2) y guarda sus
cookies como `/tmp/mi.jar` y `/tmp/no.jar` (con su `CSRF`). La cuenta **sin** `medios` debe
recibir `403 forbidden` en **toda** operación del módulo (incluido buscar), con mensaje claro en
la interfaz y el intento registrado como `denied` en la auditoría (§12).

## 3. «Buscar episodios nuevos»: forzar la búsqueda (US2, FR-010…FR-012, FR-018, SC-004)

Con la cuenta de `medios` (`/tmp/mi.jar`) y en el navegador en **`$WEB/panel/medios`**:

1. Pulsa **«Buscar episodios nuevos»** (el botón se desactiva y muestra «Buscando…» mientras
   corre: la búsqueda ocurre **en la propia petición**, sin cron, worker ni cola — FR-011).
2. Al terminar verás el **resumen** (FR-018): cuántos entraron, cuántos se unieron, cuántos
   quedaron marcados, cuántos se ocultaron por enlace confirmado no disponible, las comprobaciones
   de disponibilidad y **qué feeds fallaron** si alguno falló.

Por API:

```bash
curl -s -b /tmp/mi.jar -X POST "$API/api/v1/admin/medios/buscar-nuevos" \
  -H "X-CSRF-Token: $CSRF" | jq
```

- **Primera búsqueda con los feeds del cliente** (9 y 9 elementos, 8 en común, verificados el
  2026-10-17): quedan **10 fichas** —8 con `origin: both` (los dos enlaces) y 2 con un solo
  enlace— y el resumen dice `importedCount: 10`, `joinedCount: 8`, `markedCount: 0` (es el dato
  verificado de la spec, **no** una regla de negocio). Puede tardar unos segundos (la primera vez
  copia ~10 carátulas).
- **Repetir la búsqueda sin novedades**: `importedCount: 0`, `joinedCount: 0`, `markedCount: 0` y
  **el catálogo no cambia en nada** (FR-012: ni duplicados, ni pisadas; SC-007).
- **Un feed que no responde** (§8): el resumen lo marca como `failed`, se importa lo del otro y
  **no se oculta ni se marca nada** por esa causa (FR-016).
- Los enlaces y títulos nuevos no se escriben a mano: **0 campos completados por el equipo**
  (SC-004).

Comprobación de idempotencia a fondo: pulsa el botón 3 veces seguidas; el listado del panel y la
sección pública son idénticos (mismos `id`).

## 4. Una entrada por sermón (US2, FR-001/FR-013, SC-003)

Tras la primera búsqueda, en **`$WEB/medios`** y en el panel:

- Las fichas unidas muestran **los dos enlaces** («Ver en YouTube» y «Escuchar en Spotify») y
  `origin: both` en el panel; el nombre publicado es el del feed del video (default revisable #7).
- Las dos entradas exclusivas —«A Mayor Fe, Mayor Presión» (solo video) y «Fe como fruto del
  Espíritu» (solo audio)— aparecen con **un solo enlace**, sin marcas y sin huecos en la ficha.
- En el panel, la lista de fuentes de una ficha unida muestra **las dos procedencias** (proveedor,
  identificador externo y el título de **cada** feed: el del podcast se conserva como dato de
  origen).

## 5. Duplicado dudoso y unión con un clic — SIN RED (US3, FR-014/FR-019, SC-003)

La regla de marcado la cubren las pruebas Go (fixtures de §0.1); para probar **el flujo de unión
del panel sin depender de la red** se siembra un par marcado directamente en la BD (mismo SQL que
usa `frontend/e2e/mediosSeed.ts`):

```bash
docker compose exec -T db psql -U app -d app <<'SQL'
-- Par «posible duplicado»: mismo día, títulos que NO coinciden ni normalizados.
-- La primera entrada (video) queda visible; la adicional (audio), oculta. Ambas con la marca
-- ABIERTA (duplicate_of_id recíproco y duplicate_mark_resolved_at NULL) — modelo de dos ejes de
-- la fase 5 (P9-19): la marca de duplicado vive en su eje y no compite con la de disponibilidad.
INSERT INTO media_items (id, title, published_at, description_es, origin, visibility,
                         duplicate_of_id)
VALUES
  ('11111111-1111-4111-8111-111111111111', 'Sermón de prueba (video)', '2026-05-10T15:00:00Z',
   'Ficha base del par dudoso.', 'video', 'visible',
   '22222222-2222-4222-8222-222222222222'),
  ('22222222-2222-4222-8222-222222222222', 'Otro título para el mismo día', '2026-05-10T18:00:00Z',
   'Ficha adicional del par dudoso.', 'audio', 'hidden',
   '11111111-1111-4111-8111-111111111111');

INSERT INTO media_episode_sources (item_id, provider, external_id, feed_title, url)
VALUES
  ('11111111-1111-4111-8111-111111111111', 'youtube', 'aaaaaaaaaaa', 'Sermón de prueba (video)',
   'https://www.youtube.com/watch?v=aaaaaaaaaaa'),
  ('22222222-2222-4222-8222-222222222222', 'podcast', 'guid-de-prueba-1', 'Otro título para el mismo día',
   'https://podcasters.spotify.com/pod/show/iglesia-simiente-santa/episodes/otro-titulo-e3zzzzz');
SQL
```

Ahora, en **`$WEB/panel/medios`**:

1. Ambas fichas llevan la marca **«posible duplicado»**; la adicional además está oculta. **Ninguna
   de las dos aparece** en `$WEB/medios` ni en el bloque de la portada mientras la marca esté
   abierta (FR-009/SC-002).
2. **Unir**: el botón «Unir» abre un diálogo con el **resultado** (nombre, carátula y enlaces
   finales) y exige **confirmación explícita** (la unión es definitiva, default revisable #8).
   Al confirmar: queda **una sola ficha** con `origin: both` y **los dos enlaces**, **visible**, sin
   marca, y la ficha absorbida desaparece de todas las listas (queda con `mergedIntoId`, no se
   borra: FR-019). La operación aparece en la auditoría (§12).
3. **Resolver «no son duplicados»** (variante para el caso de sermones de verdad distintos): en
   lugar de «Unir», usa «Resolver marca» y elige `visible` (la llamada lleva **`mark:
   "possible_duplicate"`**: cada resolución indica la marca que resuelve): **ambas fichas quedan
   visibles y sin marca activa** (US3 esc. 2) y pasan a publicarse normalmente. La marca queda como
   «Resuelta» y **las búsquedas no la reabren** (§8).

Por API (con `/tmp/mi.jar`):

```bash
# Unir: {id} se conserva, otherId queda absorbida.
curl -s -b /tmp/mi.jar -X POST "$API/api/v1/admin/medios/11111111-1111-4111-8111-111111111111/unir" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"otherId":"22222222-2222-4222-8222-222222222222"}' | jq

# Resolver una marca sin unir (ambas visibles, sin marca activa). `mark` indica qué marca se
# resuelve (modelo de dos ejes de la fase 5: aquí 'possible_duplicate'):
curl -s -b /tmp/mi.jar -X POST \
  "$API/api/v1/admin/medios/22222222-2222-4222-8222-222222222222/marcas/resolver" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"mark":"possible_duplicate","visibility":"visible"}' | jq
```

## 6. Ocultar y mostrar (US3, FR-009/FR-015, SC-002/SC-007)

- En el panel, **oculta** una ficha: desaparece de `$WEB/medios`, del bloque de la portada y de
  cualquier respuesta pública; **vuelve a mostrarla** y reaparece en su lugar del orden.
- **Pulsa «Buscar episodios nuevos» con la ficha oculta**: **sigue oculta** (FR-015/SC-007: la
  sincronización nunca cambia la visibilidad decidida por el equipo).
- Comprobación por API: `GET /api/v1/medios` nunca incluye la ficha oculta ni sus datos.

## 7. Descripción corta (US3, FR-020/FR-023, SC-007)

- La descripción que ves al importar es el **recorte de la descripción del feed** (primeras
  frases, ≤ 300 caracteres; la descripción completa del feed **no se publica**).
- **Ajústala** desde el panel (campo es y campo en opcional; `descriptionEn` **exige**
  `descriptionEs` → si envías solo el inglés, `400 invalid` con `details.descriptionEs`): el texto
  del equipo se publica al guardar y **sobrevive a las búsquedas posteriores** (pulsa «Buscar
  episodios nuevos» y compruébalo: SC-007). En el panel verás que la ficha quedó marcada como
  «descripción editada por el equipo».
- Límites: 1…300 caracteres por idioma; más allá, el mensaje indica el límite y lo que falta
  recortar. Tildes, `ñ`, `¿¡` y emojis se conservan íntegros (pruébalo escribiendo «¡Qué día! 🙌»).

## 8. La regla del «estante de 15»: marcas de disponibilidad (FR-016/FR-017, SC-014)

Los feeds solo publican sus últimos elementos (YouTube conserva ~15), así que **salir de la
ventana del feed NO significa que la grabación haya desaparecido**. Cuando un elemento importado
deja de aparecer en su feed, la búsqueda **comprueba su enlace** (gratis y sin claves) y aplica:
sigue disponible → **sigue visible** y se marca «ya no aparece en el feed»; confirmado no
disponible (404/410) → **se oculta** y se marca «no disponible»; la comprobación falla → **no se
oculta** (solo marca; regla a prueba de fallos).

**Alcance exacto del disparador (matiz de la fase 5)**: la comprobación **solo** mira lo que **sale
de la ventana del feed**. Un enlace borrado que **aún** está dentro de la ventana se sigue publicando
hasta que salga de ella (con los feeds congelados puede tardar indefinidamente): **no es un fallo**,
es el alcance acordado. Solo se comprueba, además, lo que pertenece a fichas **visibles**: una ficha
que el equipo ocultó no se re-comprueba.

**Fichas con dos enlaces (regla de agregación de la fase 5)**: se comprueba **cada fuente por
separado** y el resultado de la ficha es **el peor** de los dos — un enlace confirmado muerto oculta
**la ficha entera** aunque el otro responda 200 —. La marca «ya no aparece en el feed» solo se pone
cuando **todas** las fuentes de la ficha salieron de sus ventanas: si solo salió la del podcast y el
video sigue en el suyo, la ficha queda **sin marca** (decir «ya no aparece en el feed» sería falso).

Para verlo **de forma determinista**, siembra dos fichas con fuentes que **no** están en los feeds
reales: al pulsar «Buscar episodios nuevos», el sistema detecta que salieron de la ventana y las
comprueba:

```bash
docker compose exec -T db psql -U app -d app <<'SQL'
-- (a) Enlace VIVO (página real del podcast): debe seguir VISIBLE y marcarse
--     «ya no aparece en el feed».
INSERT INTO media_items (id, title, published_at, description_es, origin, visibility)
VALUES ('33333333-3333-4333-8333-333333333333', 'Vivo pero fuera del feed',
        '2026-01-11T15:00:00Z', 'Su enlace sigue disponible.', 'audio', 'visible');
INSERT INTO media_episode_sources (item_id, provider, external_id, feed_title, url)
VALUES ('33333333-3333-4333-8333-333333333333', 'podcast', 'guid-vivo-1',
        'Vivo pero fuera del feed',
        'https://podcasters.spotify.com/pod/show/iglesia-simiente-santa/episodes/Potencial-sin-Resultados-La-Leccin-de-la-Higuera-e3jsj3n');

-- (b) Enlace CONFIRMADO INEXISTENTE (404 verificado): debe OCULTARSE y marcarse
--     «no disponible».
INSERT INTO media_items (id, title, published_at, description_es, origin, visibility)
VALUES ('44444444-4444-4444-8444-444444444444', 'Borrado de verdad',
        '2026-01-12T15:00:00Z', 'Su enlace ya no existe.', 'audio', 'visible');
INSERT INTO media_episode_sources (item_id, provider, external_id, feed_title, url)
VALUES ('44444444-4444-4444-8444-444444444444', 'podcast', 'guid-muerto-1',
        'Borrado de verdad',
        'https://podcasters.spotify.com/pod/show/iglesia-simiente-santa/episodes/episodio-que-no-existe-abcdef123');
SQL
```

Pulsa «Buscar episodios nuevos» y comprueba en el panel:

- «Vivo pero fuera del feed»: **sigue visible** para el visitante, con la marca «ya no aparece en
  el feed» (`availabilityMark: missing_from_feed`) y `availabilityResult: available` (FR-017).
- «Borrado de verdad»: **oculta** para el visitante, con la marca «no disponible»
  (`availabilityMark: unavailable`) y
  `availabilityResult: unavailable`; el resumen la cuenta en `hiddenCount` (Q16/SC-014).
- **La comprobación que no se completa** (timeout, 5xx, caída de red) **nunca** oculta: solo marca
  y se cuenta en `availabilityInconclusiveCount`. Ese caso se prueba de forma automatizada con un
  probador falso (`internal/medios/disponibilidad_test.go`): para reproducirlo a mano bastaría
  cortar la salida a internet del backend durante la búsqueda.
- **Feed caído** (FR-016): si un feed no responde, el resumen lo marca `failed` y **0 fichas se
  ocultan o se marcan** por esa causa. Se simula bloqueando la URL del feed (p. ej. apuntando el
  host a `127.0.0.1` en `/etc/hosts` del contenedor backend) o desconectando la red: lo importante
  es comprobar que **el resto del catálogo queda intacto** y que, al repetir la búsqueda cuando el
  feed responda, se recupera lo que faltaba.
- **No reabrir marcas resueltas**: resuelve la marca de «Vivo pero fuera del feed» (elige `visible`;
  la llamada lleva `mark: "missing_from_feed"`) y vuelve a pulsar «Buscar episodios nuevos»: la marca
  **no se reabre por esa misma causa** (default revisable #6). **Pero la indisponibilidad confirmada
  manda (precisión aprobada en la fase 5)**: si una búsqueda posterior **confirma** que el enlace ya
  no está (404/410), la ficha **se oculta de nuevo** y queda con la marca «no disponible» **abierta**
  — es lo que promete el aviso de «Mostrar de nuevo» — y el equipo puede volver a mostrarla. Para
  ver ese caso sin esperar: siembra una ficha ya resuelta como «mantener visible» y apunta su enlace
  a un episodio inexistente del podcast (404): al buscar, queda **oculta + «no disponible»**.

## 9. Bilingüe (US4, FR-025/FR-026, SC-011)

- Con el selector de idioma en **inglés**: la interfaz de `$WEB/medios` y del bloque está en
  inglés; las descripciones **con versión en inglés** se muestran en inglés y las demás **en
  español** (nunca vacías ni traducidas automáticamente).
- **`title`, `publishedAt`, `coverUrl`, `videoUrl` y `audioUrl` no cambian** al cambiar de idioma
  (compruébalo comparando dos capturas; SC-011).
- Por API: `curl -s "$API/api/v1/medios?lang=en&limit=12"` vs `?lang=es`: solo `description` puede
  diferir. La elección de idioma persiste al navegar (incluida la portada) y entre visitas.
- El panel se entrega **en español** (FR-024), como F2–F7.

## 10. Bloque de la portada (US6, FR-027, SC-015)

```bash
curl -s "$API/api/v1/portada/medios?lang=es" | jq
```

- `200` con **hasta 5** fichas (tope **del servidor**: no hay parámetro de límite) en el mismo
  orden que la sección; con menos de 5 publicables, exactas las que hay (**sin huecos**); con 0,
  la respuesta viene **sin `items`** y la portada **no pinta el bloque** (sin mensaje).
- En `$WEB/`: el bloque está **al final de la portada, después del bloque de ofrendas y diezmos**
  (Q12), con las últimas 5 fichas (carátula, nombre, descripción corta) y el botón **«Ver todo»**
  que lleva a `/medios` en un solo paso. Los enlaces de las fichas del bloque **sacan del sitio**
  igual que en la sección.
- Con más de 5 publicables se ven exactamente los 5 más recientes (sin paginación ni carrusel).

## 11. Sin nada publicable → redirección limpia (US1 esc. 7, FR-008, SC-016)

Oculta **todas** las fichas (manualmente o con `UPDATE media_items SET visibility='hidden'`) y
comprueba en `$WEB`:

- La entrada **«Medios» desaparece** de la barra, del menú hamburguesa y del mapa del pie (sin
  enlaces rotos ni huérfanos).
- El bloque de la portada **no se muestra**.
- La visita directa a **`$WEB/medios` redirige a la portada** de forma limpia: sin destello de
  sección vacía (el esqueleto cubre la carga), sin aviso, sin pantalla intermedia y «Atrás» no
  entra en bucles.
- **Ante un error del servidor** (pruébalo parando el backend): `/medios` muestra su estado de
  error con «Volver a intentar» y **no** redirige (una caída transitoria no expulsa al visitante).

## 12. Auditoría (US2/US3, FR-022, SC-010)

En **`$WEB/panel/auditoria`** (o `GET /api/v1/admin/auditoria/acciones`), tras los pasos
anteriores verás filas con etiquetas humanas para los **6 códigos `media.*`**:

| Código | Cuándo |
|---|---|
| `media.sync` | cada «Buscar episodios nuevos» completada (con extracto del resumen en el detalle) |
| `media.description.update` | al guardar la descripción corta |
| `media.hide` / `media.show` | al ocultar/mostrar (un guardado que cambia datos **y** visibilidad deja **dos** filas) |
| `media.merge` | al confirmar una unión (la ficha absorbida se cita en el detalle) |
| `media.mark.resolve` | al resolver una marca (con la marca resuelta y la visibilidad elegida) |

Cada fila indica **quién, qué, sobre qué y cuándo**; los intentos denegados (`denied`) y los
rechazados (`failure`) también quedan. El registro es de **solo lectura** (0 operaciones de edición
o borrado sobre él) y ninguna operación se aplica sin registro.

## 13. Dispositivos y accesibilidad (US5, FR-028…FR-030, SC-008/SC-017)

Checklist manual (la verificación con **lector de pantalla real** es SC-008 y se ejecuta después
del despliegue, con el protocolo de F6/F7):

- [ ] 320 px (teléfono), 768 px (tableta) y 1280 px (escritorio): sin desplazamiento horizontal,
      con la ficha completa (carátula, nombre, descripción y **enlaces**) en todos los anchos.
- [ ] Recorrido **solo con teclado** por `/medios`, las fichas, los enlaces externos y el bloque:
      foco visible en cada paso, sin trampas de foco, orden comprensible.
- [ ] Los enlaces externos **se anuncian** (a qué servicio llevan, que se sale del sitio y que se
      abre en una pestaña nueva) también para el lector de pantalla (texto sr-only).
- [ ] Modo claro **y** oscuro: contraste suficiente; texto ampliado al 200 % sin perder
      información.
- [ ] Movimiento de F10 respetando `prefers-reduced-motion` (revelado al hacer scroll de la
      sección y del bloque).

## 14. Pruebas automatizadas y mapeo de criterios

```bash
cd backend  && go test ./... && go test -tags=integration ./...
cd frontend && npm test -- --run
make e2e                      # Playwright (requiere `make up` levantado)
make ci                       # lint + test + security
```

| Criterio | Dónde se comprueba |
|---|---|
| SC-001/SC-006 | §1 · `medios-publica.spec.ts` |
| SC-002 | §1, §5, §6 · pruebas «0 no publicables expuestos» · e2e |
| SC-003 | §4, §5 · tabla de casos de unión · e2e panel |
| SC-004 | §3 · pruebas del sync |
| SC-005 | §1 (DevTools: 0 peticiones a terceros) · e2e con interceptor de red |
| SC-007 | §6, §7, §8 · pruebas de protección |
| SC-008/SC-012 | §13 (manuales; SC-012 con personas, protocolo de F6/F7) |
| SC-009/SC-010 | §2, §12 · pruebas de handler y transaccionales · e2e |
| SC-011 | §9 · pruebas de i18n · e2e |
| SC-013 | §1 (solo fichas de medios en la sección y el bloque) |
| SC-014 | §8 · `disponibilidad_test.go` · e2e |
| SC-015 | §10 · `portada-medios.spec.ts` |
| SC-016 | §11 · e2e (redirección sin destello) |
| SC-017 | §13 · e2e en 3 anchos |
