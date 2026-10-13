# Quickstart — Cómo probar F6 (ministerios) en local

**Fecha**: 2026-10-13 · **Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md) ·
**Contrato**: [contracts/openapi.yaml](./contracts/openapi.yaml)

Recorrido manual de las US1–US6 y sus criterios de aceptación (el bloque de la portada está en
§14), con `curl` para la API y pasos de navegador para lo visual. Mismo estilo que el quickstart de
F4. Al final, §15 ejecuta las pruebas automatizadas y §16 mapea cada criterio a su sección.

Los dos casos que la spec hace críticos tienen sección propia: **la ficha del encargado oculta**
(§7, SC-014) y **el ministerio sin enlace y sin encargado** (§7, SC-017).

## 0. Preparar el entorno

```bash
make up                      # docker compose up -d (db, redis, backend, frontend)
make db-migrate              # aplica 000010 y 000011 sobre la BD levantada
cd backend  && make sqlc     # solo si cambiaste queries/ (código generado commiteado)
cd frontend && npm install && npm run api-gen && npm test -- --run
```

**Sin variables de entorno nuevas** (F6 no añade ninguna; `APP_TZ` de F4 sigue como está).
Tras cambiar código: `docker compose up -d --build` (§8.1.8).

Variables de trabajo:

```bash
export API=http://localhost:8080
export WEB=http://localhost:5173
```

Comprobación de siempre: `curl -i $API/healthz` → `200` (intacto, F1).

## 1. La sección pública sin cuenta (US1, FR-001…FR-006, SC-001)

```bash
curl -s "$API/api/v1/ministerios?lang=es" | jq
```

- `200` con `{"lang":"es", ...}` **sin autenticación** y con `Cache-Control: no-store`.
- Con ministerios publicados aparece `items`, en el **orden manual del módulo** (`sortOrder`, `id`;
  FR-004), con `name` y `description` siempre, y con `schedule`, `place`, `image`, `whatsappUrl` y
  `leader` **solo cuando existen** (FR-002: las claves opcionales se **omiten**, nunca vacías).
- Con la base recién migrada verás `{"lang":"es"}`: es correcto (el equipo carga el contenido desde
  el panel) y así se comprueba la regla de **clave omitida** (SC-010).
- **Ninguna respuesta pública contiene** `publicationState`, borradores, traducciones sin resolver
  ni datos del encargado oculto (SC-002/SC-006/SC-014).

En el navegador: **`$WEB/ministerios`** es la sección (una entrada de navegación «Ministerios» en
barra, menú hamburguesa y mapa del sitio, FR-006). Con la base vacía **no verás la entrada** y la
URL escrita a mano muestra la página con el **aviso de sección vacía**, **sin redirigir** (patrón
firme de F4, research R6-13): es el comportamiento esperado.

## 2. Rol y cuentas con el permiso «ministerios» (US2, FR-011)

El permiso `ministerios` ("Ministerios") **ya existe** en el catálogo de F2 (migración `000002`):
F6 **no** crea permisos nuevos y el catálogo sigue en **10** permisos. El administrador inicial ya
lo tiene (el init asigna todo el catálogo) y el formulario de roles lo ofrece sin el sello
«Disponible más adelante». Con una cuenta de administración de usuarios y roles:

```bash
# Login y cookies (flujo de F2, quickstart §2 de F2).
# Credenciales del ejemplo: la cuenta de las pruebas e2e (frontend/e2e/helpers.ts).
curl -s -c /tmp/ss.jar -X POST "$API/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@ejemplo.com","password":"Semilla.2026"}' | jq
export CSRF=$(grep csrf_token /tmp/ss.jar | awk '{print $7}')

# Dos cuentas de prueba: solo ministerios y sin permisos de F6.
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"gestores-ministerios","permissions":["ministerios"]}' | jq
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"sin-acceso-f6","permissions":[]}' | jq
```

Crea una cuenta por rol con `POST /api/v1/admin/usuarios` (ver quickstart de F3 §2) y guarda sus
cookies como `/tmp/mi.jar` y `/tmp/no.jar` (con su `CSRF`).

## 3. Crear un ministerio (US2, FR-010, FR-014, FR-019, SC-004)

Con la cuenta de `ministerios` (`/tmp/mi.jar`):

```bash
# Lo único obligatorio: nameEs y descriptionEs (FR-014). Todo lo demás es opcional.
curl -s -b /tmp/mi.jar -X POST "$API/api/v1/admin/ministerios" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{
    "nameEs": "Ministerio de alabanza",
    "nameEn": "Worship ministry",
    "descriptionEs": "Servimos a la iglesia con música y adoración en los cultos y eventos.",
    "leaderNameEs": "María Pérez",
    "leaderPhone": "+506 8888-8888",
    "leaderVisible": true,
    "whatsappUrl": "https://chat.whatsapp.com/AbCdEf123456",
    "scheduleEs": "Todos los viernes, 7:00 p. m. para ensayar",
    "placeEs": "Salón principal",
    "sortOrder": 10,
    "publicationState": "published"
  }' | jq
```

- `201` con el DTO del panel (`MinisterioAdmin`) y `whatsappUrlShared: false` (§5 explica ese
  aviso). El ministerio publicado **ya es visible** en la sección (SC-004: crear y publicar en
  menos de 2 minutos).
- **Campos inválidos → `400 invalid` con `details` por campo y sin guardar nada** (FR-014):

| Qué envías | Qué ves |
|---|---|
| sin `nameEs` o sin `descriptionEs` | `details.nameEs` / `details.descriptionEs` |
| `nameEn` sin `nameEs` (o `scheduleEn` sin `scheduleEs`, etc.) | `details.<campo>Es`: falta la versión en español **de ese campo** |
| `leaderPhone` **y** `leaderEmail` a la vez | `details.leaderPhone` (celular **o** correo, nunca ambos; Q8) |
| `leaderPhone`/`leaderEmail` sin `leaderNameEs` | `details.leaderNameEs`: el contacto necesita encargado con nombre |
| `descriptionEs` de 2.001 caracteres | `details.descriptionEs` con el límite vigente |
| `imageFile` sin `imageAltEs` | `details.imageAltEs` (FR-020) |

También en el navegador: **`$WEB/panel/ministerios`** con el formulario completo (en español,
FR-018), que muestra el interruptor «Mostrar la ficha del encargado al público» desde la creación
(Q1).

## 4. Campos opcionales: encargado, horario y lugar (US2, Q1/Q6/Q8, FR-024/FR-025)

```bash
# Encargado SOLO con nombre (sin contacto): caso válido (Q8), no es un error.
curl -s -b /tmp/mi.jar -X POST "$API/api/v1/admin/ministerios" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{
    "nameEs": "Ministerio de ujieres",
    "descriptionEs": "Recibimos y acompañamos a cada persona que llega a la iglesia.",
    "leaderNameEs": "Carlos Ruiz",
    "leaderVisible": true,
    "publicationState": "published"
  }' | jq

# Sin encargado, sin enlace, con horario variable y lugar (Q2/Q6): caso válido.
curl -s -b /tmp/mi.jar -X POST "$API/api/v1/admin/ministerios" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{
    "nameEs": "Ministerio de matrimonios",
    "descriptionEs": "Encuentros para fortalecer las familias de la iglesia.",
    "scheduleEs": "Una vez al mes, fecha variable",
    "scheduleEn": "Once a month, date varies",
    "placeEs": "En la casa de una familia de la iglesia",
    "sortOrder": 20,
    "publicationState": "published"
  }' | jq
```

- El horario y el lugar son **textos libres** (Q6/FR-024): el sistema **no** calcula fechas, ni
  recurrencias, ni calendarios a partir de ellos (SC-015). Un texto largo como «Una vez al mes,
  fecha variable» se conserva íntegro y la ficha lo muestra tal cual.
- **Un campo vacío no obliga a nada** (FR-007): puedes guardar sin encargado, sin contacto, sin
  enlace, sin horario, sin lugar y sin imagen.

## 5. Enlace de WhatsApp: aceptados y rechazados (US1, FR-015/Q3, SC-009)

```bash
# ACEPTADO: cualquier https válido — sin restricción de dominios (Q3).
curl -s -b /tmp/mi.jar -X PATCH "$API/api/v1/admin/ministerios/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"whatsappUrl": "https://ejemplo.com/grupo-del-ministerio"}' | jq   # 200 (¡no es de WhatsApp!)

# RECHAZADO (400 invalid, details.whatsappUrl; no guarda nada):
for bad in 'javascript:alert(1)' 'JAVASCRIPT:alert(1)' 'data:text/html,<script>x</script>' \
           'vbscript:msgbox(1)' 'http://chat.whatsapp.com/x' '//chat.whatsapp.com/x' \
           'https://' 'https://un dos' ; do
  echo "== $bad"
  curl -s -b /tmp/mi.jar -X PATCH "$API/api/v1/admin/ministerios/<id>" \
    -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
    -d "{\"whatsappUrl\": \"$bad\"}" | jq
done
```

- El valor guardado se sirve **tal cual** (FR-015: sin transformar el destino): compara el
  `whatsappUrl` de `GET /api/v1/admin/ministerios/<id>` con el del público.
- **Enlace compartido** (dos ministerios con el mismo enlace): se **permite** y el panel lo señala
  con un aviso —`whatsappUrlShared: true` en `MinisterioAdmin` (R6-11)—; un **duplicado exacto**
  (mismo nombre, descripción, encargado y enlace) → `409 conflict`.
- El sitio solo enlaza: **no** verifica que el grupo siga vigente (Out of Scope).

## 6. Orden manual del módulo (US2, FR-004/FR-016, SC-003)

```bash
# Cambiar el valor de orden y ver que el visitante ve exactamente ese orden.
curl -s -b /tmp/mi.jar -X PATCH "$API/api/v1/admin/ministerios/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"sortOrder": 5}' | jq
curl -s "$API/api/v1/ministerios?lang=es" | jq '.items[].name'   # el orden de esta lista ES el del visitante
```

- El orden es `sortOrder`, `id` (desempate por identificador, **no** por nombre) en la sección y en
  el bloque de la portada (FR-027), y el panel lista igual (para que se vea lo que verá el
  visitante).

## 7. Privacidad del encargado, ficha oculta y ministerio sin enlace (FR-025/FR-026, SC-014, SC-017)

**El requisito crítico de F6.** Con el interruptor apagado, **ni el nombre ni el celular/correo**
del encargado pueden aparecer en **ninguna** respuesta pública (SC-014).

```bash
# 1) Con la ficha VISIBLE, el dato sale (como debe ser):
curl -s "$API/api/v1/ministerios?lang=es" | jq '.items[] | select(.name=="Ministerio de alabanza") | .leader'
#    → {"name":"María Pérez","phone":"+506 8888-8888"}

# 2) El equipo OCULTA la ficha (sin tocar los datos):
curl -s -b /tmp/mi.jar -X PATCH "$API/api/v1/admin/ministerios/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"leaderVisible": false}' | jq
#    → 200 y el panel SIGUE viendo leaderNameEs/leaderPhone (dato guardado ≠ dato publicado)

# 3) Verificación de "0 apariciones" (SC-014) — API pública:
curl -s "$API/api/v1/ministerios?lang=es" | grep -c "María Pérez"      # → 0
curl -s "$API/api/v1/ministerios?lang=es" | grep -c "8888-8888"       # → 0
curl -s "$API/api/v1/ministerios?lang=en" | grep -c "María Pérez"      # → 0 (tampoco en inglés)
curl -s "$API/api/v1/portada/ministerios?lang=es" | grep -c "8888"     # → 0 (el bloque no lleva encargado)
curl -s "$API/api/v1/ministerios?lang=es" | jq '.items[] | select(.name=="Ministerio de alabanza") | has("leader")'
#    → false  (la clave `leader` no existe: no es una línea vacía, FR-026)

# 4) Verificación en el HTML y en el tráfico del navegador (SC-014 pide inspección):
#    - Abre $WEB/ministerios con las DevTools abiertas, pestaña Network: busca el nombre y el
#      celular en la respuesta JSON → 0 coincidencias.
#    - En Elements/Inspector: busca el nombre y el celular en el DOM → 0 coincidencias.
#    - La ficha del ministerio se ve completa SIN línea de encargado (ni vacía ni solo con nombre).
```

- **Al volver a activar el interruptor** (`"leaderVisible": true`), la ficha se publica de nuevo
  **sin reescribir nada** (US2 esc. 11).
- **Ministerio sin enlace** (§3: «Ministerio de matrimonios»): en `$WEB/ministerios` **no hay botón
  de unirse** (ni vacío ni roto, SC-009) y la ficha se ve completa con su horario y lugar.
- **Ministerio sin encargado mostrable y sin enlace** (el caso SC-017): la ficha muestra nombre +
  descripción (+ horario/lugar si existen), **0 botones vacíos, 0 líneas vacías, 0 huecos**, y el
  bloque de la portada lo presenta igual.

## 8. Publicar y retirar por elemento (US3, FR-012/FR-013, SC-002, SC-003, SC-010)

```bash
# Retirar un solo ministerio (no se borra: sigue en el panel)
curl -s -b /tmp/mi.jar -X PATCH "$API/api/v1/admin/ministerios/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"publicationState": "draft"}' | jq

# Editar un ministerio publicado → el cambio es visible de inmediato (FR-013, "se publica al guardar")
curl -s -b /tmp/mi.jar -X PATCH "$API/api/v1/admin/ministerios/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"descriptionEs": "Texto corregido."}' | jq
curl -s "$API/api/v1/ministerios?lang=es" | jq   # primera carga posterior: ya cambió (no-store)
```

- **0 borradores en público** (SC-002): `grep`/`jq` sobre `GET /api/v1/ministerios` y sobre
  `GET /api/v1/portada/ministerios` no encuentra el borrador; su imagen tampoco se sirve
  (`GET /api/v1/media/<file>` → `404`, R6-7).
- **Sin ningún publicado**: la entrada «Ministerios» desaparece de barra, menú hamburguesa y mapa
  del sitio, `/ministerios` muestra el **aviso** (sin redirigir) y la portada **no pinta el bloque**
  (SC-010/SC-016).

## 9. Bilingüe: selector es/en y fallback (US4, FR-007…FR-009, SC-006)

```bash
curl -s "$API/api/v1/ministerios?lang=en" | jq   # contenidos con inglés → inglés; sin inglés → español
curl -s "$API/api/v1/ministerios?lang=es" | jq
curl -s "$API/api/v1/ministerios?lang=fr" | jq   # 400 invalid (details.lang)
```

- En el navegador: el selector de idioma de F10 cambia la interfaz **y** los contenidos; un campo
  sin traducción se muestra **en español** (nunca vacío ni traducido automáticamente); la elección
  se mantiene al navegar (incluida la portada) y se recuerda entre visitas.
- El **enlace de WhatsApp no cambia** con el idioma (no es contenido traducible, FR-008).

## 10. Imágenes (FR-020)

```bash
# Subida tras el permiso del módulo (JPEG/PNG/WebP por firma binaria, ≤8 MB)
curl -s -b /tmp/mi.jar -X POST "$API/api/v1/admin/ministerios/imagenes" \
  -H "X-CSRF-Token: $CSRF" -F "file=@/tmp/alabanza.jpg" | jq
#   → {"fileName":"img_<uuid>.jpg","url":"/api/v1/media/img_<uuid>.jpg", ...}

# Referenciarla desde el ministerio (con texto alternativo obligatorio en español)
curl -s -b /tmp/mi.jar -X PATCH "$API/api/v1/admin/ministerios/<id>" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"imageFile":"img_<uuid>.jpg","imageAltEs":"Equipo de alabanza ensayando"}' | jq

# Descarga pública (solo con el ministerio publicado)
curl -i "$API/api/v1/media/img_<uuid>.jpg"        # 200 con la imagen
```

- Un ministerio **sin imagen** se presenta completo (SC-017: sin huecos); un archivo inválido (GIF,
  SVG, > 8 MB) → `400` con `details.file` y **el ministerio se conserva sin imagen**.

## 11. Responsividad y accesibilidad (US5, FR-022/FR-023, SC-007/SC-008)

Checklist manual sobre `$WEB/ministerios` **y** el bloque de la portada (mismo nivel que F3/F4):

- [ ] Teléfono (320 px), tableta (768 px) y computadora (1280 px): todo el contenido disponible y
      **sin desplazamiento horizontal** (SC-007).
- [ ] Recorrido completo **solo con teclado**, con foco visible en cada paso y sin trampas de foco.
- [ ] Lector de pantalla: cada ficha se entiende (títulos claros, `alt` de las imágenes).
- [ ] Contraste suficiente y texto ampliable al **200 %** sin perder información, en **tema claro y
      oscuro** (F10) (SC-008).
- [ ] Botón «Unirse al grupo» y enlaces tocables con tamaño suficiente (≥ 44 px).

## 12. Sin permiso: todo denegado (US2, FR-011, SC-005)

Con la cuenta `/tmp/no.jar` (sin `ministerios`):

```bash
curl -s -b /tmp/no.jar -X POST "$API/api/v1/admin/ministerios" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"nameEs":"Hack","descriptionEs":"Intento"}' | jq
#   → {"error":{"code":"forbidden","message":"No tienes permiso para acceder a este módulo"}}
```

- Igual en `PATCH`, `DELETE`, `GET /{id}` y la subida de imágenes: **0 operaciones sin permiso**
  (SC-005) y el intento queda registrado (`result='denied'`).
- **Cruce de permisos** (Decisión 5): una cuenta con `eventos`, `portada` o `admin_usuarios_roles`
  **no** gestiona ministerios, y una cuenta con `ministerios` no accede a los otros módulos.
- En la interfaz: `/panel/ministerios` no aparece en el menú y la ruta directa cae en el guard
  (`RequirePermission`).

## 13. Auditoría de las operaciones (FR-017, SC-011)

Con la cuenta de administración (`/tmp/ss.jar`):

```bash
curl -s -b /tmp/ss.jar "$API/api/v1/admin/auditoria/acciones?limit=20" | jq
```

- Verás una fila por operación (`targetKind: "content"`, `targetLabel: "Ministerios · Ministerio de
  alabanza"`): `ministry.create`, `ministry.update`, `ministry.publish`/`ministry.unpublish`,
  `ministry.leader.show`/`ministry.leader.hide` (solo cambios del interruptor),
  `ministry.image.upload` y `ministry.delete`.
- Datos **y** estado/visibilidad en el mismo guardado → **una fila por operación** (R6-12).
- El registro es de **solo lectura**: ninguna operación del API lo edita ni lo borra (SC-011).

## 14. Bloque de ministerios de la portada (US6, FR-027…FR-029, SC-016)

```bash
curl -s "$API/api/v1/portada/ministerios?lang=es" | jq
#   → {"lang":"es","items":[{...},{...},{...},{...},{...}]}   (máximo 5, orden sortOrder,id)
```

- **Tope 5 firme del servidor**: con 8 ministerios publicados salen **exactamente los 5 primeros**
  del orden manual; con menos, exactos los que hay (**sin huecos ni marcadores**); **sin parámetro
  de límite** ni paginación (FR-027).
- Cada ítem lleva `id`, `name`, `schedule`/`place` cuando existen e `image` (tarjeta compacta de
  imagen + nombre, con el detalle en `/ministerios`; `ux.md` §4.9). **Sin campos de encargado** en
  este DTO (FR-026).
- **Sin ningún publicado**: la respuesta es `{"lang":"es"}` (sin `items`) y la portada **no
  renderiza el bloque en absoluto** — ni vacío ni con mensaje (FR-028, a diferencia del bloque
  «Lo que viene» de F4) (SC-016).
- En el navegador (`$WEB/`): el bloque está **inmediatamente después del bloque de eventos
  («Lo que viene») y antes de «Quiénes somos»**; sus tarjetas y su botón abren `/ministerios`; en
  inglés respeta el fallback y a 320 px se pila sin scroll lateral (US6/FR-029).

## 15. Pruebas automatizadas y veredicto completo

```bash
cd backend  && go test ./... && go test -tags=integration ./...
cd frontend && npm run lint && npm run typecheck && npm test -- --run
make e2e          # incluye ministerios-publica/panel/accesibilidad y portada-ministerios
make ci           # todo junto (sqlc-verify, api-gen sin drift, govulncheck, npm audit)
```

Veredicto de F6 = checklist de §11 + §7 (privacidad) + estas pruebas en verde. La revisión final
corresponde a `qa-tester`, `revisor-codigo` y `seguridad` (quien escribe no aprueba).

## 16. Mapa de criterios → secciones

| Criterio | Sección |
|---|---|
| SC-001 (ver todo sin cuenta, unirse en < 30 s) | §1, §5, §10 |
| SC-002 (0 borradores en público) | §8 |
| SC-003 (cambio visible en la primera carga) | §6, §8 |
| SC-004 (crear y publicar en < 2 min) | §3 |
| SC-005 (permiso verificado en servidor, cruce) | §12 |
| SC-006 (inglés con fallback, 0 huecos) | §9 |
| SC-007 (320 px sin scroll) | §11 |
| SC-008 (accesibilidad, claro y oscuro) | §11 |
| SC-009 (enlaces seguros; 0 botones vacíos) | §5, §7 |
| SC-010 (0 secciones vacías; nav sin enlaces rotos) | §1, §8 |
| SC-011 (auditoría completa, solo lectura) | §13 |
| SC-012 (usabilidad con personas ≥ 90 %) | §11 (protocolo; informe `pruebas-usabilidad-SC-012.md`) |
| SC-013 (solo ministerios; horario sin fechas) | §1, §4 |
| **SC-014 (0 apariciones del encargado oculto)** | **§7** |
| SC-015 (horario/lugar tal como se escribieron) | §4 |
| SC-016 (bloque: 5 exactos, oculto sin publicados) | §14 |
| SC-017 (sin enlace y sin encargado: ficha coherente) | §7 |
