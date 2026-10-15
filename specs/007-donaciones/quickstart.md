# Quickstart — Cómo probar F7 (ofrendas y diezmos) en local

**Fecha**: 2026-10-14 · **Spec**: [spec.md](./spec.md) · **Plan**: [plan.md](./plan.md) ·
**Contrato**: [contracts/openapi.yaml](./contracts/openapi.yaml)

Recorrido manual de las US1–US6 y sus criterios de aceptación (el bloque de la portada está en
§14), con `curl` para la API y pasos de navegador para lo visual. Mismo estilo que el quickstart de
F6. Al final, §15 ejecuta las pruebas automatizadas y §16 mapea cada criterio a su sección.

Los casos que la spec hace críticos tienen sección propia: **los números exactos** —IBAN validado
por estructura y longitud y agrupado en bloques (§4, SC-009) y SINPE normalizado (§5, SC-009)—, **la
ficha sin titular y el SINPE sin moneda** (§3, SC-010/SC-013), los **dos textos con su propio
estado** (§7, US3 esc. 7) y **la redirección de `/ofrendas` sin nada publicado** (§10, SC-010).

## 0. Preparar el entorno

```bash
make up                      # docker compose up -d (db, redis, backend, frontend)
make db-migrate              # aplica 000012 y 000013 sobre la BD levantada
cd backend  && make sqlc     # solo si cambiaste queries/ (código generado commiteado)
cd frontend && npm install && npm run api:gen && npm test -- --run
```

**Sin variables de entorno nuevas** (F7 no añade ninguna). Tras cambiar código:
`docker compose up -d --build` (§8.1.8).

Variables de trabajo:

```bash
export API=http://localhost:8080
export WEB=http://localhost:5173
```

Comprobación de siempre: `curl -i $API/healthz` → `200` (intacto, F1).

## 1. La sección pública sin cuenta (US1, FR-001…FR-006, SC-001)

```bash
curl -s "$API/api/v1/ofrendas?lang=es" | jq
```

- `200` con `{"lang":"es", ...}` **sin autenticación** y con `Cache-Control: no-store`.
- Con contenido publicado aparecen `items` (los métodos, en el **orden de la sección**: primero los
  `iban` y después los `sinpe`, cada grupo por `sortOrder`, `id` — FR-004), `intro` (el texto
  introductorio) y `usage` (el texto de uso), **cada uno solo si está publicado** (Q10).
- Cada método trae `kind`, `number` (íntegro y exacto), `institution` siempre, y `currency`
  (**solo si es IBAN**) y `holder` **solo cuando existe** (Q2: las claves opcionales se **omiten**,
  nunca vacías; US1 esc. 5).
- Con la base recién migrada verás `{"lang":"es"}`: es correcto (el equipo carga el contenido desde
  el panel) y así se comprueba la regla de **clave omitida** (SC-010). La API **no** redirige:
  `GET /api/v1/ofrendas` sigue siendo **`200` con las claves omitidas** (el contrato no cambia; la
  redirección de la ruta SPA es una decisión del frontend, P7-21).
- **Ninguna respuesta pública contiene** `publicationState`, borradores ni traducciones sin
  resolver (SC-002/SC-006).

En el navegador: **`$WEB/ofrendas`** es la sección «Ofrendas y diezmos» (una entrada de navegación
en barra, menú hamburguesa y mapa del sitio, FR-006). Con la base vacía **no verás la entrada** y
la URL escrita a mano **redirige a la portada** (`$WEB/`) —decisión del cliente del 2026-10-14 (Q9),
**distinta a propósito** de F4 en `/eventos`, que muestra su aviso sin redirigir (no se unifican;
P7-21)—: verás la portada completa, **sin ningún destello** de la sección vacía (estado de carga →
redirección limpia) y el botón «Atrás» no te devuelve a la URL muerta. Con algo publicado (métodos
**o** textos — R7-12 —), la URL muestra la sección con normalidad.

## 2. Rol y cuentas con el permiso «donaciones» (US2, FR-011)

El permiso `donaciones` ("Donaciones") **ya existe** en el catálogo de F2 (migración `000002`,
línea 25): F7 **no** crea permisos nuevos, **no lo renombra** (Q8) y el catálogo sigue en **10**
permisos. El administrador inicial ya lo tiene (el init asigna todo el catálogo) y el formulario de
roles lo ofrece sin el sello «Disponible más adelante». Con una cuenta de administración de usuarios
y roles:

```bash
# Login y cookies (flujo de F2, quickstart §2 de F2).
# Credenciales del ejemplo: la cuenta de las pruebas e2e (frontend/e2e/helpers.ts).
curl -s -c /tmp/ss.jar -X POST "$API/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@ejemplo.com","password":"Semilla.2026"}' | jq
export CSRF=$(grep csrf_token /tmp/ss.jar | awk '{print $7}')

# Dos cuentas de prueba: solo donaciones y sin permisos de F7.
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"gestores-donaciones","permissions":["donaciones"]}' | jq
curl -s -b /tmp/ss.jar -X POST "$API/api/v1/admin/roles" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"name":"sin-acceso-f7","permissions":[]}' | jq
```

Crea una cuenta por rol con `POST /api/v1/admin/usuarios` (ver quickstart de F3 §2) y guarda sus
cookies como `/tmp/mi.jar` y `/tmp/no.jar` (con su `CSRF`).

## 3. Crear un método de donación (US2, FR-010, FR-014, FR-019, SC-004)

Con la cuenta de `donaciones` (`/tmp/mi.jar`):

```bash
# Obligatorios: kind, number e institution (Q2) y, además, currency si es IBAN (FR-014).
# holderEs/holderEn son OPCIONALES ("si se rellena, se publica"; Q2).
curl -s -b /tmp/mi.jar -X POST "$API/api/v1/admin/donaciones" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{
    "kind": "iban",
    "number": "CR01 0000-0000-0000-0000-0",
    "institution": "Banco Nacional de Costa Rica",
    "holderEs": "Iglesia Simiente Santa",
    "currency": "CRC",
    "publicationState": "published",
    "sortOrder": 1
  }' | jq
```

- `201` con el método guardado; `number` viene **normalizado** (`"CR0100000000000000000000"`:
  los espacios y guiones desaparecen y todo va en mayúsculas — Q3/FR-015 —).
- **Cuenta IBAN sin titular** (caso **válido**, Q2): envía el mismo alta sin `holderEs` → `201` y
  la ficha pública **no** mostrará línea de titular (ni vacía ni con marcador; US1 esc. 5).
- **SINPE Móvil** (sin moneda — Q2 —):

```bash
curl -s -b /tmp/mi.jar -X POST "$API/api/v1/admin/donaciones" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{
    "kind": "sinpe",
    "number": "8888 8888",
    "institution": "Kolbi",
    "publicationState": "published",
    "sortOrder": 1
  }' | jq
```

- `201` y `number` guardado como `"+50688888888"` (prefijo normalizado — Q4/FR-016 —).

**Errores que no guardan nada** (400 `invalid` con `details` por campo; SC-004/FR-014):

| Envío | Resultado |
|---|---|
| Sin `kind`, `number` o `institution` | 400 `details.kind` / `details.number` / `details.institution` |
| `"kind": "iban"` **sin** `currency` | 400 `details.currency` (la moneda es obligatoria en IBAN, FR-014) |
| `"kind": "sinpe"` **con** `currency` | 400 `details.currency` (el SINPE no lleva moneda, Q2) |
| `holderEn` sin `holderEs` | 400 `details.holderEs` (español base por campo, Decisión 8) |
| Cualquier error | **no se guarda nada a medias** (la transacción es toda o nada) |

## 4. Cuentas IBAN: aceptadas y rechazadas (US1, FR-015/Q3, SC-009)

```bash
# ACEPTADO: IBAN costarricense (22 caracteres) escrito con espacios/guiones → se normaliza.
#   "CR01 0000-0000-0000-0000-0"  →  "CR0100000000000000000000"
# ACEPTADO: IBAN internacional válido según su país (longitud exacta del registro IBAN):
#   "ES91 2100 0418 4502 0005 1332" (ES = 24) · "DE89 3704 0044 0532 0130 00" (DE = 22)
# RECHAZADO (400 invalid, details.number; no guarda nada):
curl -s -b /tmp/mi.jar -X POST "$API/api/v1/admin/donaciones" -H 'Content-Type: application/json' \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"kind":"iban","number":"CR01 0000-0000-0000-0000","institution":"Banco Nacional","currency":"CRC"}' | jq
#   → "estructura o longitud no válidas" (CR exige 22 caracteres: faltaba uno)
#   "CR01 0000-0000-0000-0000-00" (23 para CR) · "ZZ01 0000000000000000" (país fuera del registro)
#   "CR01-0000-0000-0000-0000-😀" · "hola" (ni letras ni emojis fuera de la estructura)
```

- El número **guardado y publicado** es el exacto normalizado (FR-015: "sin transformaciones que lo
  alteren"); la presentación **agrupada en bloques** (`CR01 0000 0000 0000 0000 0000`) es del
  cliente y el **botón de copiar copia el valor exacto** (`number`).
- Nota: la validación es **estructura y longitud** (literal de Q3/FR-015); **no** se verifica el
  dígito de control mod-97 (pregunta abierta al humano, plan.md §"Preguntas").

## 5. SINPE Móvil: aceptados y rechazados (US1, FR-016/Q4, SC-009)

```bash
# ACEPTADOS (todos se guardan como "+50688888888"):
#   "88888888" · "8888-8888" · "8888 8888" · "+506 8888 8888" · "+50688888888"
# RECHAZADOS (400 invalid, details.number):
curl -s -b /tmp/mi.jar -X POST "$API/api/v1/admin/donaciones" -H 'Content-Type: application/json' \
  -H "X-CSRF-Token: $CSRF" \
  -d '{"kind":"sinpe","number":"8888 888","institution":"Kolbi"}' | jq
#   → "celular costarricense de 8 dígitos" (falta uno)
#   "2222-2222" (línea fija, no celular: primer dígito 6/7/8) · "888888888" (9 dígitos)
#   "+507 8888 8888" · "ocho ocho…" (letras) · "8888-8888-88"
```

- El número publicado está **en su formato normalizado** (`+50688888888`), listo para copiarse o
  transcribirse sin error (SC-009). El prefijo exacto del guardado es un default del plan
  (`+506` presente; R7-3) sin impacto de negocio.

## 6. Orden manual dentro de cada grupo (US2, FR-004, SC-003)

```bash
# Cambiar el valor de orden de un método dentro de su grupo (iban o sinpe):
curl -s -b /tmp/mi.jar -X PATCH "$API/api/v1/admin/donaciones/$ID" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"sortOrder": 3}' | jq
```

- En el panel verás el mismo orden que el visitante: **primero el grupo de cuentas IBAN y después
  el de SINPE Móvil** y, dentro de cada grupo, `sortOrder`, `id` (US2 esc. 10).
- Recarga `$WEB/ofrendas`: el orden es exactamente el que acabas de fijar (SC-003).

## 7. Los dos textos: introductorio y de uso (US2, FR-021, Q5/Q11)

```bash
# GET: los dos textos, cada uno con SU PROPIO estado; null hasta el primer guardado (patrón F3).
curl -s -b /tmp/mi.jar "$API/api/v1/admin/donaciones/textos" | jq

# PUT del texto introductorio («cómo donar»): crea o reemplaza el singleton.
curl -s -b /tmp/mi.jar -X PUT "$API/api/v1/admin/donaciones/textos/intro" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{
    "textEs": "Puedes apoyar la obra de la iglesia con tu ofrenda o diezmo por transferencia bancaria o por SINPE Móvil. Elige la cuenta o el número de la lista y usa los datos tal como aparecen.",
    "textEn": "You can support the church work with your offering or tithe via bank transfer or SINPE Móvil.",
    "publicationState": "published"
  }' | jq

# PUT del texto de uso de las donaciones (Q5: texto libre, NO un listado de destinos).
curl -s -b /tmp/mi.jar -X PUT "$API/api/v1/admin/donaciones/textos/usage" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{
    "textEs": "Tus aportes se usan en la obra social, el mantenimiento del templo y las misiones.",
    "publicationState": "published"
  }' | jq
```

- `textEs` obligatorio (1…2.000 caracteres), `textEn` opcional (exige `textEs`) y
  `publicationState` por defecto `draft`. Texto plano de uno o varios párrafos: tildes, `ñ`, `¿¡` y
  emojis se conservan **íntegros** (constitución §IV).
- Cualquier `textKey` fuera de `intro`/`usage` → 400 (catálogo cerrado).
- **No hay operación de borrado**: un texto se retira poniéndolo en `draft` (FR-010 habla de
  "mantenerlos"; ver pregunta 6 de plan.md).
- En el navegador, `$WEB/ofrendas` muestra cada texto **solo si está publicado** y sin huecos si no
  lo está (US3 esc. 7).

## 8. Publicar y retirar por elemento y por texto (US3, FR-012/FR-013, SC-002, SC-003, SC-010)

```bash
# Retirar un solo método (no se borra: sigue en el panel)
curl -s -b /tmp/mi.jar -X PATCH "$API/api/v1/admin/donaciones/$ID" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"publicationState":"draft"}' | jq

# Editar un método publicado → el cambio es visible de inmediato (FR-013, "se publica al guardar")
curl -s -b /tmp/mi.jar -X PATCH "$API/api/v1/admin/donaciones/$ID" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"number":"CR01 0000-0000-0000-0000-1"}' | jq

# Retirar el texto de uso (su propio estado, Q10) sin tocar el introductorio
curl -s -b /tmp/mi.jar -X PUT "$API/api/v1/admin/donaciones/textos/usage" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"textEs":"Tus aportes se usan en la obra social, el mantenimiento del templo y las misiones.","publicationState":"draft"}' | jq
```

Comprueba con `curl -s "$API/api/v1/ofrendas?lang=es" | jq` tras cada paso: el borrador **no
aparece** (ni método ni texto) y el resto de la sección se sostiene **sin huecos ni marcadores**
(US3 esc. 1/esc. 7, SC-002).

## 9. Bilingüe: selector es/en y fallback (US4, FR-007…FR-009, SC-006)

```bash
curl -s "$API/api/v1/ofrendas?lang=en" | jq
```

- Con `lang=en`: `holder` y los textos salen en inglés **si existe traducción** y en español si no
  (FR-008: **nunca** un campo vacío; p. ej. el `usage` del paso anterior solo tiene español).
- **Los números, el tipo, la entidad financiera y la moneda NO cambian** entre idiomas (FR-007/US4
  esc. 5/SC-006): compara la respuesta `es` y la `en` campo a campo (idénticos).
- En el navegador: el selector de idioma (F10) se mantiene al navegar y entre visitas; sin
  preferencia guardada, todo se muestra en español (Decisión 8).

## 10. Sin nada publicado: ocultación y redirección (US3, FR-005/FR-006, SC-010)

Con **todo** en borrador (métodos **y** textos):

```bash
curl -s "$API/api/v1/ofrendas?lang=es" | jq     # → {"lang":"es"}   (claves omitidas)
curl -s "$API/api/v1/portada/ofrendas?lang=es" | jq   # → {"lang":"es"}  (sin items)
```

En el navegador, con las DevTools abiertas:

- **No** aparece la entrada «Ofrendas y diezmos» ni en la barra, ni en el menú hamburguesa, ni en el
  mapa del sitio del pie (sin enlaces rotos ni huérfanos).
- `$WEB/ofrendas` escrito a mano **redirige a la portada** (`$WEB/`) de forma limpia: **sin
  destello** de sección vacía (solo verás el estado de carga → portada), **sin aviso ni mensaje**
  y con «Atrás» sin bucles (`replace`). 0 pantallas intermedias (SC-010).
- El bloque de la portada **no se pinta** (§14).

> Regla de "nada publicado" = sin métodos **ni textos** publicados (R7-12, default revisable):
> con **solo textos** publicados la sección y su navegación siguen vivas (aunque el bloque siga
> oculto, §14). La pregunta está abierta al humano (plan.md §"Preguntas", punto 2).

## 11. Responsividad y accesibilidad (US5, FR-024/FR-025, SC-007/SC-008)

Checklist manual (la ejecuta `qa-tester` y se registra en la revisión):

- [ ] **320 px** (teléfono): la sección y el bloque sin desplazamiento horizontal; los números se
      leen completos (IBAN en bloques, SINPE normalizado) y se pueden copiar con el botón.
- [ ] **768 px** (tableta) y **1280 px** (computadora): mismos contenidos y funciones que en móvil
      (0 pérdida entre dispositivos).
- [ ] **Teclado**: recorrido completo con `Tab`/`Enter` (fichas, botón de copiar, enlaces, selector
      de idioma) con **foco visible** en cada paso y sin trampas de foco.
- [ ] **Lector de pantalla**: las fichas anuncian tipo, número, entidad, moneda y titular en un
      orden comprensible; los grupos tienen encabezados reales.
- [ ] **Contraste** suficiente en tema claro y oscuro y **zoom 200 %** sin perder información.
- [ ] Con poco contenido (1 método, sin textos) la página se compone **sin huecos** ni marcadores.

**Usabilidad con personas (SC-012, protocolo manual estilo F3/F4/F6)**: ≥ 6 personas (2 por franja
18–35 / 36–59 / 60+, incluidas con poca experiencia en internet), sin ayuda del equipo, con el
guion de 3 tareas: (1) encuentra **a qué cuenta o número donar** y cópialo/transcríbelo; (2) di
**en qué se usan las donaciones**; (3) cambia el idioma y comprueba que los números siguen
iguales. Se cronometra la tarea 1 (umbral **< 30 s** de SC-001) y se registra el resultado por
participante en `pruebas-usabilidad-SC-012.md` (umbral: **≥ 90 %** de aciertos sin ayuda).

## 12. Sin permiso: todo denegado (US2, FR-011, SC-005)

Con `/tmp/no.jar` (cuenta sin `donaciones`) y con una cuenta que tenga **otros** permisos de módulo
(p. ej. solo `ministerios`):

```bash
curl -s -b /tmp/no.jar -X POST "$API/api/v1/admin/donaciones" \
  -H 'Content-Type: application/json' -H "X-CSRF-Token: $CSRF" \
  -d '{"kind":"sinpe","number":"88888888","institution":"Kolbi"}' | jq
#   → 403 {"error":{"code":"forbidden","message":"No tienes permiso para acceder a este módulo"}}
```

- Lo mismo para `GET` de la lista, `PATCH`, `DELETE` y los dos `PUT` de textos: **toda** operación
  del módulo está denegada con mensaje claro (US2 esc. 2) y **tener otros permisos no da acceso**
  (US2 esc. 9: los permisos son independientes).
- Un intento forzado por la API queda **registrada** en la auditoría como `denied` (§13).
- En el panel, la entrada «Gestión de donaciones» y la ruta `/panel/donaciones` no existen para esa
  cuenta (`RequirePermission`).

## 13. Auditoría de las operaciones (FR-017, SC-011)

```bash
curl -s -b /tmp/mi.jar "$API/api/v1/admin/auditoria/acciones?limit=50" | jq
```

- Cada operación deja su fila con **quién, qué, sobre qué y cuándo**: `donation.create` (alta),
  `donation.update` (edición y reordenamiento), `donation.publish`/`donation.unpublish` (cambios de
  estado de un método), `donation.delete` (borrado), `donation.text.update` (guardado de un texto)
  y `donation.text.publish`/`donation.text.unpublish` (cambios de estado de un texto).
- Un guardado que cambia **datos y estado** deja **dos filas** (una por operación) y el alta deja
  **solo** `donation.create`. Los rechazos (`failure`) y las denegaciones (`denied`) también se
  registran (best-effort).
- `targetLabel`: `Donaciones · <entidad> · <número>` para métodos y `Donaciones · Texto
  introductorio` / `Donaciones · Texto de uso` para los textos.
- El registro es de **solo lectura**: no hay ninguna operación que lo edite ni lo borre.

## 14. Bloque de ofrendas y diezmos de la portada (US6, FR-022, SC-014)

```bash
curl -s "$API/api/v1/portada/ofrendas?lang=es" | jq
#   → {"lang":"es","items":[ … ]}   (máximo 5, orden de la sección: iban → sinpe, sortOrder, id)
```

- Con más de 5 métodos publicados aparecen **exactamente los 5 primeros** del orden de la sección
  (sin paginación ni carrusel); con menos, exactos los que hay (**sin huecos** ni marcadores).
- Con **0 métodos publicados** la respuesta viene **sin `items`** y la portada **no renderiza el
  bloque** (ni vacío ni con mensaje: diferencia consciente con el bloque «Lo que viene» de F4).
- En el navegador, en `$WEB/`: el bloque está **inmediatamente debajo del bloque de ministerios**
  (F6) y antes de «Quiénes somos», cada ítem identificado con su tipo (cuenta IBAN o SINPE Móvil) y
  con un enlace/botón que abre `/ofrendas` en un solo paso. El resto de la portada de F3 no cambia.
- Idioma, responsividad y accesibilidad del bloque: mismos criterios que la sección (§9, §11).

## 15. Pruebas automatizadas y veredicto completo

```bash
cd backend  && go test ./... && go test -tags=integration ./...
cd frontend && npm test -- --run
make e2e          # Playwright: ofrendas-publica / ofrendas-panel / ofrendas-accesibilidad / portada-ofrendas
make ci           # lint + tests + govulncheck + npm audit
```

Veredicto de F7 (para `qa-tester`): **0 borradores expuestos** (SC-002) · **cambio guardado visible
en la primera carga** (SC-003) · **< 2 min** para crear y publicar un método (SC-004) · **0
operaciones sin permiso** (SC-005) · **fallback sin huecos y datos idénticos entre idiomas**
(SC-006) · **320 px sin scroll horizontal** (SC-007) · **0 hallazgos críticos de accesibilidad**
(SC-008) · **0 números mal formados y 0 caracteres alterados** (SC-009) · **0 secciones vacías y
redirección limpia** (SC-010) · **100 % de operaciones registradas** (SC-011) · **cómo donar
identificado sin ayuda** (SC-012, usabilidad con personas) · **0 elementos de otros módulos**
(SC-013) · **bloque con exactamente 5 y oculto sin métodos** (SC-014).

## 16. Mapa de criterios → secciones

| Criterio | Sección |
|---|---|
| SC-001 (ve la información y encuentra cómo donar) | §1, §3, §7 |
| SC-002 (0 borradores) | §8, §15 |
| SC-003 (cambio visible en la primera carga) | §6, §8 |
| SC-004 (< 2 min desde el panel) | §3 |
| SC-005 (permiso verificado en servidor) | §2, §12 |
| SC-006 (bilingüe sin huecos; datos idénticos) | §9 |
| SC-007 (320 px sin scroll) | §11 |
| SC-008 (accesibilidad) | §11 |
| SC-009 (números íntegros; 0 mal formados) | §4, §5 |
| SC-010 (0 vacíos; redirección de `/ofrendas`) | §1, §8, §10 |
| SC-011 (auditoría completa y de solo lectura) | §13 |
| SC-012 (usabilidad con personas) | §11 (protocolo) |
| SC-013 (sin eventos/grupos/ministerios en la sección) | §1, §14 |
| SC-014 (bloque: 5 exactos, debajo de ministerios, oculto sin métodos) | §14 |
| US1–US6 | §1 · §2–§7 · §8 · §9 · §11 · §14 |
