# Quickstart — Cómo probar F3 (portada e información general) en local

**Fecha**: 2026-10-09 · **Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md) ·
**Contrato**: [contracts/openapi.yaml](./contracts/openapi.yaml)

Recorrido manual de las US1–US5 y sus criterios de aceptación, con `curl` para el panel y la portada
pública y pasos de navegador para lo visual. Mismo estilo que el quickstart de F2. Al final,
§11 ejecuta las pruebas automatizadas y §12 mapea cada criterio a su sección.

## 0. Preparar el entorno

```bash
make up                      # docker compose up -d (db, redis, backend, frontend)
make db-migrate              # aplica 000005 y 000006 sobre la BD levantada
cd backend  && make sqlc     # solo si cambiaste queries/ (código generado commiteado)
cd frontend && npm install && npm run api-gen && npm test -- --run
```

Variables nuevas (`.env.example`): `UPLOAD_DIR` (por defecto `./uploads` localmente;
`/var/lib/simiente/uploads` en el contenedor, volumen `uploads_data`) y `UPLOAD_MAX_BYTES`
(8 MB). Tras cambiar código: `docker compose up -d --build` (§8.1.8).

Variables de trabajo:

```bash
export API=http://localhost:8080
export WEB=http://localhost:5173
```

Comprobación de siempre: `curl -i $API/healthz` → `200` (intacto, F1).

## 1. Portada pública sin cuenta (US1, FR-001…FR-007, SC-001)

```bash
curl -s "$API/api/v1/portada?lang=es" | jq
```

- `200` con `{"lang":"es", ...}` **sin autenticación** (no hace falta cookie ni token).
- Con contenido publicado aparecen las secciones con datos: `identity` (nombre, lema, misión, visión,
  logo e imagen de portada), `about` (quiénes somos), `schedule` (horario), `whatsapp`, `socials` y
  `contact` (dirección, correo, teléfono).
- **Secciones sin elementos publicados NO aparecen** (SC-012): si aún no hay nada publicado, la
  respuesta es `{"lang":"es"}` y el navegador muestra la portada sin secciones vacías.
- Ninguna respuesta pública contiene `publicationState`, traducciones sin resolver ni borradores
  (SC-002). Con la base recién migrada verás `{"lang":"es"}`: es correcto (el equipo carga el
  contenido desde el panel; no hay contenido sembrado).
- Cabeceras: `Cache-Control: no-store` (SC-003).

En el navegador: `$WEB/` es la portada (FR-001). La pantalla «Estado del sistema» de F1 sigue viva
en **`$WEB/health`** (R3-15; ruta de la SPA, distinta del endpoint `$API/healthz`).

## 2. Rol y cuenta con el permiso del módulo (US2, FR-012)

El permiso `portada` ("Portada e información general") **ya existe** en el catálogo de F2
(migración `000002`): no se crea nada nuevo. Con una cuenta de administración de usuarios y roles:

```bash
# Login y cookies (reutiliza el flujo de F2, quickstart §2).
# Credenciales del ejemplo: la cuenta del administrador sembrada por las pruebas e2e
# de este repo (frontend/e2e/helpers.ts). En otro entorno, usa el correo y la
# contraseña de tu propio administrador (creado con la inicialización de F2).
curl -s -c /tmp/ss.jar -X POST "$API/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@ejemplo.com","password":"Semilla.2026"}' | jq
export CSRF=$(grep csrf_token /tmp/ss.jar | awk '{print $7}')
```

```bash
# Rol con el permiso de portada y su cuenta de editor
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"contenido","permissions":["portada"]}' | jq
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/usuarios" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"firstName":"Ana","lastName":"Editora","email":"editora@ejemplo.com",
       "phone":"8888-8888","roleId":"<id-del-rol>","password":"OtraClave1!"}' | jq
```

## 3. Editar identidad, «quiénes somos» y contacto (US2, FR-002, FR-003, FR-007, FR-011, FR-015)

Con la cuenta editora (o cualquier cuenta con permiso `portada`):

```bash
# Identidad: nombre oficial + lema/misión/visión (es obligatorio, en opcional)
curl -s -b /tmp/ss.jar -X PUT "$API/api/v1/admin/portada/identidad" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{
    "nameEs": "Iglesia Simiente Santa",
    "nameEn": "Simiente Santa Church",
    "taglineEs": "Un espacio para encontrarse con Dios",
    "missionEs": "Simiente Santa es un espacio donde las personas pueden encontrarse con Dios, vivir en libertad, construir relaciones genuinas y descubrir su propósito para generar un impacto en su entorno.",
    "visionEs": "consolidarse como una iglesia que transforma vidas y comunidades, formando personas libres, seguras y con propósito, que viven una fe auténtica y generan un impacto en el mundo.",
    "publicationState": "published"
  }' | jq

# Quiénes somos: texto plano, ≤1.000 caracteres por idioma
curl -s -b /tmp/ss.jar -X PUT "$API/api/v1/admin/portada/quienes-somos" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"textEs":"Simiente Santa es un espacio donde las personas pueden encontrarse con Dios de forma real, cercana y transformadora.","textEn":"","publicationState":"published"}' | jq

# Contacto: dirección, correo y teléfono son obligatorios (FR-015)
curl -s -b /tmp/ss.jar -X PUT "$API/api/v1/admin/portada/contacto" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"addressEs":"San José, Costa Rica","email":"hola@simientesanta.org","phone":"+506 8888-8888","publicationState":"published"}' | jq
```

Errores esperados (todos `400 invalid` con `details` por campo; **no se guarda nada**, FR-015):

| Qué enviar | Qué se ve |
|---|---|
| `{"nameEs":"","publicationState":"draft"}` | `details.nameEs` — campo obligatorio |
| `textEs` de 1.001 caracteres | `details.textEs` con el límite (FR-003) |
| `"email":"no-es-correo"` | `details.email` |
| `"phone":"12"` | `details.phone` (≥7 dígitos, criterio de F2) |
| `{"textEn":"Solo inglés"}` sin `textEs` | `details.textEs` — falta la versión en español (Decisión 8) |
| `{"nameEs":"x","publicationState":"publicado"}` | `details.publicationState` — solo `draft`/`published` |

Con caracteres especiales: prueba `tildes, ñ, ¿¡ y emojis 🙌` en `textEs` → se conservan íntegros en
la respuesta y en la portada (edge case de la spec).

**Campos en inglés (`*En`) — normalización** *(analyze I6)*: `""` o solo espacios en cualquier campo
`*En` se guarda como `null` ("sin traducción"): el ejemplo de «quiénes somos» de arriba envía
`"textEn":""` y guarda **sin error** (US2 esc. 5). En la respuesta del panel verás `textEn: null`.
Un `*En` con contenido real se guarda tal cual.

## 4. Horario, WhatsApp y redes (US2, FR-004, FR-005, FR-006)

```bash
# Servicio del horario (día 0=domingo…6=sábado, hora "HH:MM")
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/portada/horario" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"dayOfWeek":0,"startTime":"10:00","endTime":"12:00","nameEs":"Culto dominical","placeEs":"Templo principal","publicationState":"published"}' | jq

# Canal de WhatsApp: mensaje directo (teléfono) y grupo (URL)
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/portada/whatsapp" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"nameEs":"Escríbenos","kind":"direct","destination":"+506 8888-8888","publicationState":"published"}' | jq
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/portada/whatsapp" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"nameEs":"Grupo de la iglesia","kind":"group","destination":"https://chat.whatsapp.com/XXXXXXXX","publicationState":"published"}' | jq

# Red social (catálogo fijo: facebook, instagram, youtube, tiktok, spotify)
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/portada/redes" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"network":"instagram","url":"https://www.instagram.com/simientesanta","publicationState":"published"}' | jq
```

Errores esperados:

| Qué enviar | Qué se ve |
|---|---|
| WhatsApp `{"kind":"direct","destination":"abc"}` | `400` con `details.destination` (teléfono inválido) |
| WhatsApp `{"kind":"group","destination":"http://otro-sitio.com/x"}` | `400` con `details.destination` (debe ser https de `chat.whatsapp.com`/`wa.me`) |
| El mismo canal otra vez | `409 conflict` — canal duplicado exacto (kind + destino + nombre) |
| Red `{"network":"x","url":"https://x.com/iglesia"}` | `400` con `details.network` — red fuera del catálogo |
| Segundo enlace para `instagram` | `409 conflict` — un solo enlace por red (Q5) |
| `{"network":"facebook","url":"https://otro-sitio.com"}` | `400` con `details.url` — dominio que no es de esa red |
| Horario con `startTime":"25:99"` | `400` con `details.startTime` |
| Horario con `endTime` ≤ `startTime` (p. ej. `10:00`–`09:00`) | `400` con `details.endTime` (el fin es opcional pero debe ser posterior, `analyze` C2) |

Estado completo del módulo (incluye borradores; solo con permiso): `curl -s -b /tmp/ss.jar
"$API/api/v1/admin/portada" | jq`.

## 5. Publicar y retirar por elemento (US3, FR-013, FR-014, SC-002, SC-003, SC-012)

```bash
# Retirar un solo elemento (no se borra: sigue en el panel)
curl -s -b /tmp/ss.jar -X PATCH "$API/api/v1/admin/portada/horario/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"publicationState":"draft"}' | jq

# Editar un elemento publicado → el cambio es visible de inmediato (FR-014)
curl -s -b /tmp/ss.jar -X PATCH "$API/api/v1/admin/portada/horario/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"startTime":"11:00","publicationState":"published"}' | jq
```

Comprobación en la portada pública:

1. Con el elemento en `draft` → `curl -s "$API/api/v1/portada" | jq` **no lo contiene** y el resto se
   ve coherente (sin huecos); si esa sección queda sin publicados, **desaparece por completo** (SC-012).
2. Al publicarlo → aparece en la primera carga posterior (SC-003; `no-store`).
3. Al editar un elemento publicado y guardar → el visitante ve el dato nuevo recargando (FR-014).
4. `DELETE` de un elemento → `204` y desaparece también del panel (borrado físico; sin historial).

La publicación es **por elemento**, también en los *singletons*: identidad, «quiénes somos» y contacto
son cada uno un elemento con su **propio** `publicationState` y se publican/retiran por separado (el
mismo `PUT` de §3 con `publicationState: draft|published`; retirar uno no oculta los demás). Solo lo
publicado llega al visitante (FR-013). **Y también aplica a las imágenes** *(analyze C4)*: al retirar
la identidad, su logo/imagen de portada deja de servirse (`GET /api/v1/media/…` → `404`, §6) sin
borrarse; al publicarla de nuevo, la misma URL vuelve a responder `200`.

## 6. Logo e imagen de portada (US2, FR-002, FR-019)

```bash
# Subir el logo (solo JPEG/PNG/WebP por firma binaria, ≤8 MB)
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/portada/imagenes" \
  -H "X-CSRF-Token: $CSRF" -F "file=@resources/simiente.jpeg" | jq
# → {"fileName":"img_<uuid>.jpg","url":"/api/v1/media/img_<uuid>.jpg","mimeType":"image/jpeg","sizeBytes":123456}

# Referenciarlo desde la identidad (con texto alternativo obligatorio, FR-019)
curl -s -b /tmp/ss.jar -X PUT "$API/api/v1/admin/portada/identidad" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"nameEs":"Iglesia Simiente Santa","logoFile":"img_<uuid>.jpg","logoAltEs":"Logotipo de Iglesia Simiente Santa","publicationState":"published"}' | jq

# Descarga pública (la portada la usa en <img>)
curl -i "$API/api/v1/media/img_<uuid>.jpg"   # 200, Content-Type image/jpeg, nosniff, inline, no-store
```

**Política de descarga (FR-013/SC-002, `analyze` C4)**: la descarga **solo** sirve archivos
**referenciados por contenido publicado** (en F3, el logo/imagen de portada de una identidad
`published`). `Cache-Control: no-store` (la respuesta depende del estado de publicación).

Errores esperados: archivo `.svg` o `.html` renombrado → `400` con `details.file` (tipo no permitido);
archivo de 9 MB → `400` con `details.fileSize` (límite de 8 MB); `logoFile` sin `logoAltEs` → `400`
con `details.logoAltEs`; nombre que **no cumple el patrón** en `/api/v1/media/…` → **`400 invalid`**
y nombre válido pero **inexistente o no publicado** → **`404 not_found`** (criterio fijo, `analyze`
M6). La portada **no se rompe** si la imagen falta: se muestra sin ella (edge case).

**Comprobación con borradores (C4)**: con la identidad retirada a borrador (§5),
`curl -i "$API/api/v1/media/img_<uuid>.jpg"` → **`404`** (el archivo sigue en disco, pero no se
sirve); al volver a publicar la identidad, la misma URL responde `200` otra vez.

## 7. Bilingüe: selector es/en y fallback (US4, FR-008, FR-009, FR-010, SC-006)

```bash
curl -s "$API/api/v1/portada?lang=en" | jq    # contenidos en inglés donde existen
curl -s "$API/api/v1/portada?lang=es" | jq    # todo en español
curl -s "$API/api/v1/portada?lang=fr" | jq    # 400 invalid (solo es|en)
```

- Un contenido **sin** versión en inglés aparece **en español** (idioma base): nunca un campo vacío ni
  una traducción automática (FR-009). Prueba dejando `textEn:""` en «quiénes somos» y pidiendo `lang=en`
  (el `""` se guarda como `null`, §3, y el público muestra el texto **en español**, jamás un hueco).
- En el navegador (`$WEB/`): el selector visible cambia la página **de inmediato** (interfaz y
  contenidos), se mantiene al navegar durante la visita **y entre visitas en el mismo dispositivo**
  (`localStorage`, R3-9: la primera visita sin preferencia guardada entra en español; si el guardado
  falla, solo no persiste entre visitas, sin avisos). **No** se auto-detecta el idioma del navegador.
  La interfaz pública está **100 % traducida** en inglés (claves tipadas).

## 8. Sin permiso: todo denegado (US2 esc. 2 y 7, FR-012, SC-005, SC-011)

Con una cuenta que **no** tenga el permiso `portada`:

- El panel no muestra la entrada "Portada" ni se puede forzar la ruta (`/panel/informacion` redirige a
  "No tienes acceso a esta sección").
- Por API forzada, todo responde `403 forbidden` con mensaje claro y sin detalles internos:

```bash
curl -s -b /tmp/sin-permiso.jar -X PUT "$API/api/v1/admin/portada/contacto" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"addressEs":"x","email":"a@b.co","phone":"8888888","publicationState":"draft"}' | jq
# → {"error":{"code":"forbidden","message":"No tienes permiso para acceder a este módulo"}}
```

- La denegación **queda registrada** en la auditoría (`result='denied'`, FR-017).
- **Orden real de guards** (el de F2, verificado en el cierre): en la **primera sesión** de una
  cuenta con `mustChangePassword`, el guard de contraseña responde `403 password_change_required`
  **antes** de evaluarse el permiso del módulo; **sin sesión**, `401 unauthenticated`. Solo una
  cuenta sin esas condiciones llega al `403 forbidden` de permisos del enunciado.

## 9. Auditoría de las ediciones (FR-017, SC-013)

Tras los cambios de §3–§6, con la cuenta de administración de usuarios y roles:

```bash
curl -s -b /tmp/ss.jar "$API/api/v1/admin/auditoria/acciones?limit=20" | jq
```

- Cada edición aparece con **quién** (`actorName`/`actorEmail`), **qué** (`action` `home.*` +
  `targetLabel` como `Portada · Horario · Culto dominical`) y **cuándo** (`createdAt`).
- **Regla de códigos** *(analyze M5)*: un **alta** registra `home.*.create` aunque el elemento nazca
  publicado; `home.publish`/`home.unpublish` **solo** aparecen en cambios de estado; una edición que
  cambia datos y estado deja dos filas. **Las subidas de imagen (§6) también se auditan** con
  `home.image.upload` (`targetLabel` = `Portada · Imagen · <fileName>`, `analyze` I8).
- También se registran los intentos fallidos (`failure`) y las denegaciones (`denied`).
- El registro sigue siendo de **solo lectura**: no existe ninguna operación de escritura sobre él y la
  sección de auditoría de F2 no muestra controles de edición (FR-025 de F2).
- Atomicidad: si el registro fallara, la edición **no se aplica** (prueba de integración; edge case
  "no aplicarse sin registro").

## 10. Responsividad y accesibilidad (US5, FR-018, FR-019, SC-007, SC-008)

Checklist manual sobre `$WEB/` (la ejecuta `qa-tester` en la validación; referencia WCAG 2.1 AA):

1. **Teléfono (320 px), tableta (768 px) y escritorio (1280 px)** (el mínimo **320 px** es el que
   promete `ux.md`, `analyze` I7): todas las secciones visibles, sin
   desplazamiento horizontal y con los mismos enlaces y funciones en los tres (SC-007).
2. **Toques**: cada enlace/botón con área de al menos 44 px (regla de la skill `react-frontend`).
3. **Teclado**: recorrido completo con `Tab` en orden comprensible, foco visible y activación con
   `Enter`/`Espacio` (incluido el selector de idioma y los enlaces de WhatsApp/redes).
4. **Lector de pantalla**: títulos de sección claros (`h1`/`h2`), imágenes con `alt` (el logo y la
   imagen de portada usan los textos alternativos del panel, FR-019) y avisos con `aria-live`.
5. **Baja visión**: contraste suficiente con la paleta del Manual (azul `#1a2b4a` sobre blanco/crema;
   turquesa solo como acento con texto de contraste) y ampliación de texto al 200 % sin pérdida.
6. **Sin menús ocultos ni gestos**: navegación simple con títulos claros (personas con poca
   experiencia en internet).
7. **Usabilidad SC-010 (prueba manual con personas; protocolo mínimo, `analyze` M7)**: ≥ **6
   personas** (2 por franja 18–35 / 36–59 / 60+, incluidas con poca experiencia en internet), guion
   de **3 tareas sin ayuda** sobre la portada real —(1) encontrar el horario de servicios, (2)
   encontrar un canal de WhatsApp, (3) cambiar a inglés y volver a español—, sesión breve observada
   por `qa-tester` con el humano. Resultado por tarea y persona documentado en
   `specs/003-portada-info-general/pruebas-usabilidad-SC-010.md` (umbral: ≥ 90 % de tareas
   completadas). **No automatizable**: queda fuera de `make ci`/`make e2e` y su alcance real es el
   que este protocolo describe.

## 11. Pruebas automatizadas y veredicto completo

```bash
cd backend
go test ./...                          # unitarias (service, handler, platform/storage, validate, audit)
go test -tags=integration ./...        # repository contra PostgreSQL real + migración 000006 up/down/up
go test ./... -cover                   # ≥80 % en internal/portada/service*.go (§III)

cd ../frontend
npm run lint && npm run typecheck
npm test -- --run                      # Vitest + MSW (portada, panel, i18n)
npx playwright test e2e/portada-publica.spec.ts e2e/portada-panel.spec.ts   # e2e (local)

make ci                                # lint + test + security (govulncheck, npm audit)
```

## 12. Mapa de criterios → secciones

| Criterio | Sección |
|---|---|
| SC-001 (información visible sin cuenta) | §1, §11 (e2e pública) |
| SC-002 (0 borradores expuestos) | §5, §11 (pruebas "0 borradores expuestos") |
| SC-003 (cambio visible desde la primera carga) | §5, §11 (e2e panel) |
| SC-004 (actualizar y publicar en <2 min) | §3–§5 (observación en e2e panel) |
| SC-005 / SC-011 (permiso verificado en servidor) | §8, §11 |
| SC-006 (inglés completo, fallback a español) | §7, §11 |
| SC-007 / SC-008 (dispositivos y accesibilidad) | §10, §11 (e2e en 3 anchos) |
| SC-009 (enlaces con un clic) | §4, §11 (e2e pública) |
| SC-010 (personas de distintas edades) | §10.7 (protocolo de usabilidad con personas; informe `pruebas-usabilidad-SC-010.md`) |
| SC-012 (0 secciones vacías) | §1, §5 |
| SC-013 (ediciones registradas; registro de solo lectura) | §9, §11 |
| FR-015 (validación de entradas) | §3, §4 (tablas de errores) |
| FR-018 / FR-019 (responsividad y accesibilidad) | §6, §10 |
