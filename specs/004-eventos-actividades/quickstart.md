# Quickstart — Cómo probar F4 (eventos y actividades) en local

**Fecha**: 2026-10-11 · **Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md) ·
**Contrato**: [contracts/openapi.yaml](./contracts/openapi.yaml)

Recorrido manual de las US1–US6 y sus criterios de aceptación, con `curl` para la API y pasos de
navegador para lo visual. Mismo estilo que el quickstart de F3. Al final, §13 ejecuta las pruebas
automatizadas y §14 mapea cada criterio a su sección.

## 0. Preparar el entorno

```bash
make up                      # docker compose up -d (db, redis, backend, frontend)
make db-migrate              # aplica 000007, 000008 y 000009 sobre la BD levantada
cd backend  && make sqlc     # solo si cambiaste queries/ (código generado commiteado)
cd frontend && npm install && npm run api-gen && npm test -- --run
```

Variables nuevas (`.env.example`): **`APP_TZ=America/Costa_Rica`** — zona horaria del lugar para el
corte de "hoy" del retiro automático (research R4-2; **valor fijado por el humano en la puerta 2**,
no es un ejemplo: es el valor del proyecto). **Si no se define una zona válida, el backend falla con
error claro** (mejor un corte correcto que uno silencioso).
Tras cambiar código: `docker compose up -d --build` (§8.1.8).

Variables de trabajo:

```bash
export API=http://localhost:8080
export WEB=http://localhost:5173
```

Comprobación de siempre: `curl -i $API/healthz` → `200` (intacto, F1).

## 1. "Lo que viene" sin cuenta (US1, FR-001…FR-005, SC-001)

```bash
curl -s "$API/api/v1/proximamente?lang=es" | jq
```

- `200` con `{"lang":"es", ...}` **sin autenticación** y con `Cache-Control: no-store`.
- Con contenido publicado y próximo aparecen los dos bloques: `events` (cada evento con `dates` —
  su primera entrada es **la próxima fecha** (FR-004) — y `frequency` si se cargó) y `activities`
  (cada una con su `date`, `startTime`/`endTime` opcionales). Cada elemento trae `title`, `place`,
  `detail`, y si se cargaron también `mapUrl`, `organizer`, `category` e `image` (FR-018/FR-019).
- **Orden cronológico** de lo más próximo a lo más lejano dentro de cada bloque (FR-003), sin
  filtros ni paginación (FR-003). Eventos y actividades son **bloques distintos** (Decisión 3): no
  se mezclan nunca en un solo listado.
- **Ninguna respuesta pública contiene** `publicationState`, borradores, fechas pasadas ni
  traducciones sin resolver (SC-002/SC-009). Con la base recién migrada verás `{"lang":"es"}`: es
  correcto (el equipo carga el contenido desde el panel).
- **Bloques omitidos cuando están vacíos** (SC-012): si no hay eventos próximos, la clave `events`
  no existe.

En el navegador: **`$WEB/eventos`** es la sección (una entrada de navegación
«Eventos y actividades» con los bloques «Próximos eventos» y «Próximas actividades», FR-021). Con la
base vacía **no verás la entrada en la navegación** y la URL escrita a mano muestra la página con el
**aviso** «Por ahora no hay eventos ni actividades» — **sin redirigir** (research R4-13): es el
comportamiento esperado.

## 2. Rol y cuentas con los permisos de los módulos (US2/US3, FR-012)

Los permisos `eventos` ("Eventos") y `actividades` ("Actividades") **ya existen** en el catálogo de
F2 (migración `000002`). El permiso **`categorias`** ("Categorías de agenda") es **nuevo** en F4
(migración `000009`, 10.º del catálogo — cambio de alcance aprobado en la puerta 2): la gestión del
catálogo de categorías **solo** se autoriza con él (no basta `eventos` ni `actividades`; research
R4-4). El catálogo lo lee F2 dinámicamente: el **administrador inicial recibe `categorias` solo** y
el formulario de roles lo ofrece (con su sello retirado, como los otros dos). Con una cuenta de
administración de usuarios y roles:

```bash
# Login y cookies (flujo de F2, quickstart §2 de F2).
# Credenciales del ejemplo: la cuenta de las pruebas e2e (frontend/e2e/helpers.ts).
curl -s -c /tmp/ss.jar -X POST "$API/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@ejemplo.com","password":"Semilla.2026"}' | jq
export CSRF=$(grep csrf_token /tmp/ss.jar | awk '{print $7}')

# Cuatro cuentas de prueba: solo eventos, solo actividades, solo categorías y
# sin permisos de F4
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"editores-eventos","permissions":["eventos"]}' | jq
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"editores-actividades","permissions":["actividades"]}' | jq
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"gestores-categorias","permissions":["categorias"]}' | jq
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"sin-acceso-f4","permissions":[]}' | jq
```

Crea una cuenta por rol con `POST /api/v1/admin/usuarios` (ver quickstart de F3 §2) y guarda sus
cookies como `/tmp/ev.jar`, `/tmp/ac.jar`, `/tmp/cat.jar`, `/tmp/no.jar` (con su `CSRF`).

## 3. Crear un evento con fechas (US2, FR-010, FR-018, SC-004)

Con la cuenta de `eventos` (`/tmp/ev.jar`):

```bash
# Evento periódico: título, fechas concretas, texto de frecuencia opcional,
# lugar + enlace de mapa opcional, detalle, organizador y categoría opcional.
curl -s -b /tmp/ev.jar -X POST "$API/api/v1/admin/eventos" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{
    "titleEs": "Vigilia de oración",
    "titleEn": "Prayer vigil",
    "frequencyEs": "Todos los viernes",
    "detailEs": "Una noche para orar y adorar juntos. Trae tu Biblia.",
    "placeEs": "Templo principal",
    "mapUrl": "https://maps.google.com/?q=Iglesia+Simiente+Santa",
    "organizerEs": "Equipo de intercesión",
    "categoryId": "<id-de-categoria>",
    "publicationState": "published",
    "dates": [
      {"date": "2026-10-16", "startTime": "19:00", "endTime": "21:00"},
      {"date": "2026-10-23", "startTime": "19:00"},
      {"date": "2026-10-30", "startTime": "19:00"}
    ]
  }' | jq
```

- El evento queda con sus **fechas concretas** (Q1: el sistema **no** calcula recurrencias; la
  periodicidad la comunica `frequencyEs` + las fechas cargadas).
- `dates` exige **al menos una** fecha y admite hasta 50; cada fecha es un día concreto con
  `startTime`/`endTime` opcionales (Q2). Una actividad puede tener solo `startTime`, ninguna hora o
  el rango completo.
- Con `publicationState: "published"` el evento es público **si tiene fecha próxima**; la primera
  carga de `GET /api/v1/proximamente` ya lo muestra (SC-003).

Errores esperados (todos `400 invalid` con `details` por campo; **no se guarda nada**, FR-015):

| Qué enviar | Qué se ve |
|---|---|
| `{"titleEs":"","detailEs":"x","placeEs":"y","dates":[{"date":"2026-10-16"}]}` | `details.titleEs` — campo obligatorio |
| `dates: []` (sin ninguna fecha) | `details.dates` — el evento necesita al menos una fecha |
| `{"date":"2026-13-32"}` o `{"date":"32/10/2026"}` | `details.dates[i].date` — fecha inválida (formato `YYYY-MM-DD`) |
| `{"date":"2026-10-16","startTime":"19:00","endTime":"18:00"}` | `details.dates[i].endTime` — la fin debe ser posterior al inicio |
| `{"date":"2026-10-16","endTime":"21:00"}` | `details.dates[i].endTime` — el fin exige el inicio (decisión firme) |
| `{"date":"2026-10-16","startTime":"25:99"}` | `details.dates[i].startTime` — hora fuera de `HH:MM` |
| Dos fechas `2026-10-16` en el mismo evento | `details.dates` — fecha repetida |
| `detailEs` de 2.001 caracteres | `details.detailEs` con el límite (FR-018) |
| `{"mapUrl":"no-es-una-url"}` | `details.mapUrl` — enlace mal formado |
| `{"titleEn":"Solo inglés"}` sin `titleEs` | `details.titleEs` — falta la versión en español (Decisión 8) |
| `{"imageFile":"img_x.jpg"}` sin `imageAltEs` | `details.imageAltEs` — la imagen necesita texto alternativo (FR-020) |
| `{"categoryId":"<uuid-inexistente>"}` | `details.categoryId` — categoría desconocida |

Con caracteres especiales: prueba `tildes, ñ, ¿¡ y emojis 🙌` en `titleEs`/`detailEs` → se conservan
íntegros (edge case). **Las fechas pasadas están permitidas** al crear/editar (el equipo puede cargar
lo que ya ocurrió) pero ese elemento no aparece en "lo que viene" (edge case, FR-002).

**Campos en inglés (`*En`)**: `""` o solo espacios se guarda como `null` ("sin traducción", patrón
I6 de F3): puedes enviar `"titleEn":""` sin error y ver `titleEn: null` en la respuesta.

## 4. Crear una actividad puntual (US3, FR-011, FR-019, SC-004)

Con la cuenta de `actividades` (`/tmp/ac.jar`):

```bash
curl -s -b /tmp/ac.jar -X POST "$API/api/v1/admin/actividades" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{
    "titleEs": "Reparto de comida",
    "detailEs": "Vamos al barrio San José a repartir almuerzos. Puedes ayudar a cargar.",
    "placeEs": "Barrio San José",
    "organizerEs": "Ministerio de acción social",
    "categoryId": "<id-de-accion-social>",
    "date": "2026-10-18",
    "startTime": "09:00",
    "endTime": "13:00",
    "publicationState": "published"
  }' | jq
```

- Una actividad tiene **una** fecha (Decisión 3): sin periodicidad ni lista de fechas; su horario es
  opcional (solo inicio, rango o sin hora — Q2). Las mismas validaciones de §3 aplican.
- **Duplicado exacto del módulo** (misma actividad con el mismo título, fecha y lugar): `409
  conflict` (edge case). Un evento y una actividad con los mismos datos **no** son duplicados entre
  sí (Decisión 3): ambos se conservan.

## 5. Categorías del catálogo (Q6, FR-018/FR-019)

El catálogo se gestiona **solo con el permiso `categorias`** (research R4-4): usa la cuenta
`/tmp/cat.jar`. Con `eventos` o `actividades` **no** alcanza (pruébalo: responde `403`).

```bash
# El catálogo inicial viene sembrado por 000007 (acción social, predicación,
# adoración, entrevista, noche de oración, otro).
curl -s -b /tmp/cat.jar "$API/api/v1/admin/categorias" | jq

curl -s -b /tmp/cat.jar -X POST "$API/api/v1/admin/categorias" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"nameEs": "Jovenes", "nameEn": "Youth", "sortOrder": 60}' | jq

# Cruce de permisos (SC-005/SC-011): cuenta solo `eventos` frente al catálogo
curl -s -b /tmp/ev.jar "$API/api/v1/admin/categorias" | jq   # 403 forbidden
```

Errores esperados: nombre duplicado → `409 conflict`; `DELETE` de una categoría que usan eventos o
actividades → `409 conflict` ("está en uso"): primero quítasela a esos elementos (research R4-5,
decisión firme). La categoría es **opcional**: un elemento sin categoría es válido (Q6).

## 6. Imágenes (FR-020)

```bash
# Subida tras el permiso del módulo (JPEG/PNG/WebP por firma binaria, ≤8 MB)
curl -s -b /tmp/ev.jar -X POST "$API/api/v1/admin/eventos/imagenes" \
  -H "X-CSRF-Token: $CSRF" -F "file=@resources/simiente.jpeg" | jq
# → {"fileName":"img_<uuid>.jpg","url":"/api/v1/media/img_<uuid>.jpg", ...}

# Referenciarla desde el evento (con texto alternativo obligatorio)
curl -s -b /tmp/ev.jar -X PATCH "$API/api/v1/admin/eventos/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"imageFile":"img_<uuid>.jpg","imageAltEs":"Sala de la vigilia iluminada"}' | jq

# Descarga pública
curl -i "$API/api/v1/media/img_<uuid>.jpg"   # 200, image/jpeg, nosniff, inline, no-store
```

Errores esperados: `.svg`/`.html` renombrado → `400` con `details.file`; archivo de 9 MB → `400`
con `details.fileSize`; nombre fuera del patrón en `/api/v1/media/…` → `400 invalid`; nombre válido
pero inexistente, **en borrador o de un elemento cuya fecha ya pasó** → `404 not_found` (la imagen
sigue en disco y vuelve a servirse al publicar o al cargar una fecha próxima — research R4-6,
decisión firme). Un
elemento sin imagen se muestra completo (FR-020) y la sección no se rompe.

## 7. Publicar y retirar por elemento (US4, FR-013, FR-014, SC-002, SC-003, SC-012)

```bash
# Retirar un solo elemento (no se borra: sigue en el panel)
curl -s -b /tmp/ev.jar -X PATCH "$API/api/v1/admin/eventos/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"publicationState":"draft"}' | jq

# Editar un elemento publicado → el cambio es visible de inmediato (FR-014)
curl -s -b /tmp/ev.jar -X PATCH "$API/api/v1/admin/eventos/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"placeEs": "Salón de actos", "publicationState":"published"}' | jq
```

Comprobación en `$API/api/v1/proximamente` y en `$WEB/eventos`:

1. Con el elemento en `draft` → **no aparece** y el resto se ve coherente (sin huecos); si **todo**
   queda en borrador, los bloques desaparecen, **la entrada de navegación desaparece** (barra, menú
   hamburguesa y mapa del sitio del pie) y la URL directa `/eventos` muestra la **página con el
   aviso** de sección vacía («Por ahora no hay eventos ni actividades»), **sin redirigir** (SC-012,
   R4-13).
2. Al publicar → aparece en la primera carga posterior (SC-003; `no-store`).
3. Al editar un elemento publicado y guardar → el visitante ve el dato nuevo recargando (FR-014).
4. `DELETE` (con confirmación en la interfaz) → `204`, desaparece del panel y del público, y queda
   registrado en la auditoría (FR-017).

La publicación es **por elemento** (Decisión 4): retirar un evento no oculta las actividades ni los
demás eventos.

## 8. Bilingüe: selector es/en, fallback y fechas localizadas (US5, FR-007…FR-009, SC-006)

```bash
curl -s "$API/api/v1/proximamente?lang=en" | jq   # contenidos en inglés donde existen
curl -s "$API/api/v1/proximamente?lang=es" | jq   # todo en español
curl -s "$API/api/v1/proximamente?lang=fr" | jq   # 400 invalid (solo es|en)
```

- Un contenido **sin** versión en inglés aparece **en español** (idioma base): nunca un campo vacío
  ni una traducción automática (FR-008). Prueba dejando `frequencyEn` en `null` y pidiendo
  `lang=en`: verás el texto de frecuencia **en español**.
- En el navegador: el selector cambia la página de inmediato (interfaz y contenidos), se mantiene
  al navegar **y entre visitas** en el mismo dispositivo (`localStorage`, R3-9 de F3; primera visita
  sin preferencia → español). **No** se auto-detecta el idioma del navegador.
- **Fechas y horas** (FR-009): con `lang=es` verás algo como «viernes 17 de octubre, 7:00 p. m. −
  9:00 p. m.» y con `lang=en` «Friday, October 17, 7:00 PM − 9:00 PM» (research R4-3): fecha
  escrita, no solo numérica, sin cambiar el día ni la hora reales. Los eventos muestran su **próxima
  fecha** y, si cargaste más, las siguientes; el texto de frecuencia acompaña.

## 9. Retiro automático al pasar la fecha (FR-002, FR-006, Q3/Q4, SC-009)

Sin cron ni intervención del equipo: la vista pública solo muestra lo de **hoy en adelante**.

1. Crea/publica una **actividad** con `date` = **ayer** → `GET /api/v1/proximamente` **no la
   contiene** (y la sección se oculta si es lo único).
2. Crea/publica un **evento** cuyas fechas sean todas pasadas → no aparece (evento sin fecha
   próxima), pero **sigue en el panel** con `nextDate: null` e `isPast: true` y con el aviso de que
   falta cargar la próxima fecha (SC-009).
3. Un elemento **de hoy** aparece durante su día; si tiene `endTime` y la hora ya pasó, desaparece
   solo (la regla exacta está en research R4-2). Prueba con un `endTime` hace 5 minutos: no aparece;
   con un `endTime` en 5 minutos: sí.
4. Al cargarle una fecha futura al evento, **vuelve a aparecer** sin tocar su estado: sigue
   `published` en el panel (Q4), donde el equipo lo ve como pasado. Nada de esto se borra ni se
   convierte en borrador.
5. Lo pasado **no** se muestra en el sitio público (su recuento corresponde a F8, fuera de F4).

## 10. Responsividad y accesibilidad (US6, FR-023, FR-024, SC-007, SC-008)

Checklist manual sobre `$WEB/eventos` (la ejecuta `qa-tester`; referencia WCAG 2.1 AA):

1. **Teléfono (320 px), tableta (768 px) y escritorio (1280 px)**: ambos bloques completos, sin
   desplazamiento horizontal, mismos enlaces y funciones en los tres (SC-007).
2. **Toques**: cada enlace/botón con área de al menos 44 px (regla de la skill `react-frontend`);
   el enlace de mapa se activa con comodidad.
3. **Teclado**: recorrido completo con `Tab` en orden comprensible, foco visible y activación con
   `Enter`/`Espacio` (incluidos los enlaces de cada elemento y el selector de idioma).
4. **Lector de pantalla**: títulos de bloque claros (`h1`/`h2`), imagen con el `alt` cargado en el
   panel (FR-020), fechas como texto legible (no solo números) y `<time datetime>` semántico.
5. **Baja visión**: contraste suficiente **en tema claro y oscuro** (F10) y ampliación del texto al
   200 % sin pérdida de información.
6. **Sin menús ocultos ni gestos**: navegación simple (barra en escritorio, menú hamburguesa
   etiquetado en móvil, F10) para personas con poca experiencia en internet.
7. **Usabilidad SC-010 (prueba manual con personas; protocolo mínimo, estilo F3/M7)**: ≥ **6
   personas** (2 por franja 18–35 / 36–59 / 60+, incluidas con poca experiencia en internet), guion
   de **3 tareas sin ayuda** sobre la sección real —(1) decir qué eventos vienen y cuándo es el
   próximo, (2) encontrar dónde es una actividad concreta y a qué hora, (3) cambiar a inglés y
   volver—, sesión breve observada por `qa-tester` con el humano. Resultado documentado en
   `specs/004-eventos-actividades/pruebas-usabilidad-SC-010.md` (umbral: ≥ 90 % de tareas
   completadas). **No automatizable**: queda fuera de `make ci`/`make e2e`.

## 11. Sin permiso: todo denegado, incluido el cruce entre módulos (US2/US3, FR-012, SC-005, SC-011)

Con la cuenta **solo `eventos`** (`/tmp/ev.jar`) sobre actividades, y al revés; y con la cuenta sin
permisos (`/tmp/no.jar`) sobre todo:

- El panel no muestra las entradas ajenas y las rutas forzadas redirigen a "No tienes acceso a esta
  sección". La entrada «Categorías» **sí** aparece **solo** para las cuentas con el permiso
  `categorias` (research R4-4): ni `eventos` ni `actividades` la desbloquean.
- Por API forzada, todo responde `403 forbidden` con mensaje claro y sin detalles internos
  (incluido `/api/v1/admin/categorias` desde una cuenta solo `eventos` o solo `actividades`):

```bash
curl -s -b /tmp/ev.jar -X POST "$API/api/v1/admin/actividades" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"titleEs":"x","detailEs":"y","placeEs":"z","date":"2026-12-01"}' | jq
# → {"error":{"code":"forbidden","message":"No tienes permiso para acceder a este módulo"}}

curl -s -b /tmp/ev.jar "$API/api/v1/admin/categorias" | jq   # 403 forbidden (hace falta `categorias`)

curl -s -b /tmp/no.jar "$API/api/v1/admin/eventos" | jq   # 403 forbidden
```

- La denegación **queda registrada** en la auditoría (`result='denied'`, FR-017).
- Orden de guards (como en F2/F3): sin sesión → `401 unauthenticated`; con `mustChangePassword` →
  `403 password_change_required` **antes** del permiso; el `403 forbidden` de permisos solo llega
  en una cuenta con sesión válida.

## 12. Auditoría de las operaciones (FR-017, SC-013)

Tras los cambios de §3–§7, con la cuenta de administración de usuarios y roles:

```bash
curl -s -b /tmp/ss.jar "$API/api/v1/admin/auditoria/acciones?limit=20" | jq
```

- Cada operación aparece con **quién** (`actorName`/`actorEmail`), **qué** (código `event.*`,
  `activity.*` o `category.*` + `targetLabel` como `Eventos · Vigilia de oración`) y **cuándo**.
- **Regla de códigos** (research R4-10): un **alta** registra `*.create` aunque nazca publicada;
  `*.publish`/`*.unpublish` **solo** aparecen en cambios de estado; una edición que cambia datos y
  estado deja **dos filas**; las subidas de imagen (§6) se auditan con `event.image.upload` /
  `activity.image.upload`; los borrados con `*.delete`.
- También se registran los intentos fallidos (`failure`: validación, duplicado) y las denegaciones
  (`denied`).
- El registro sigue siendo de **solo lectura**: ninguna operación de escritura sobre él (FR-025 de
  F2). Atomicidad: si el registro fallara, la operación **no se aplica** (prueba de integración).

## 13. Pruebas automatizadas y veredicto completo

```bash
cd backend
go test ./...                          # unitarias (service, handler, audit + pruebas de F2 actualizadas a 10 permisos)
go test -tags=integration ./...        # repository contra PostgreSQL real + migraciones 000008/000009 up/down/up
go test ./... -cover                   # ≥80 % en internal/agenda/service*.go (§III)

cd ../frontend
npm run lint && npm run typecheck
npm test -- --run                      # Vitest + MSW (página pública, panel, i18n)
npx playwright test e2e/agenda-publica.spec.ts e2e/agenda-panel.spec.ts   # e2e (local)

make ci                                # lint + test + security (govulncheck, npm audit)
```

> Nota: las pruebas de F2 que fijaban el catálogo de permisos en 9 se **actualizaron deliberadamente
> a 10** como parte de F4 (cambio de producto autorizado, excepción registrada — RG4-10 del plan):
> ver `backend/internal/usuarios/`. Si `go test ./...` falla ahí, no es una regresión de F2.

## 14. Mapa de criterios → secciones

| Criterio | Sección |
|---|---|
| SC-001 (lo que viene visible sin cuenta) | §1, §13 (e2e pública) |
| SC-002 (0 borradores expuestos, tampoco imágenes) | §7, §6, §13 (pruebas "0 borradores/vencidos expuestos") |
| SC-003 (cambio visible desde la primera carga) | §7, §13 (e2e panel) |
| SC-004 (crear y publicar en <2 min) | §3–§5 (observación en e2e panel) |
| SC-005 / SC-011 (permiso verificado, cruce de módulos) | §11, §13 |
| SC-006 (inglés completo, fallback a español) | §8, §13 |
| SC-007 / SC-008 (dispositivos y accesibilidad) | §10, §13 (e2e en 3 anchos) |
| SC-009 (retiro automático; 0 vencidos en "lo que viene") | §9, §13 (fixtures con fechas fijas) |
| SC-010 (personas de distintas edades) | §10.7 (protocolo de usabilidad; informe `pruebas-usabilidad-SC-010.md`) |
| SC-012 (0 secciones vacías; navegación sin enlaces rotos) | §1, §7 |
| SC-013 (operaciones registradas; registro de solo lectura) | §12, §13 |
| FR-015 (validación de entradas) | §3, §4 (tablas de errores) |
| FR-018 / FR-019 (contenido de cada concepto) | §3, §4, §5 |
| FR-020 (imagen opcional con alt) | §6 |
| FR-021 / FR-022 (ubicación en el sitio y ocultación) | §1, §7 |
| FR-023 / FR-024 (responsividad y accesibilidad) | §10 |
