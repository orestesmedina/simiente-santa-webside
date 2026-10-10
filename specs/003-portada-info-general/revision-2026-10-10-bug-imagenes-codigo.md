# Revisión de código — bug «imágenes de la portada no cargan» (F3)

**Fecha**: 2026-10-10 · **Rama**: `003-portada-info-general` · **Auditor**: `revisor-codigo`
(solo lectura; este reporte es la única escritura). Validación del bucle de bug (paso 4/5).
**Objeto**:
- `ec2ae76` — `fix(frontend): F3 resuelve las URLs de media contra la API` (`mediaUrl()` en
  `frontend/src/api/client.ts` + uso en `IdentityHero`, `PublicLayout`, `PublicFooter`, y
  aserciones de `src` resuelto en `homePage.test.tsx` / `client.test.ts`).
- `b257e0b` — prueba de regresión `frontend/e2e/portada-imagenes.spec.ts` (con el ajuste del
  helper `browserBase` y del locator del logo incluido en `ec2ae76`).

**Método**: `git show` de ambos commits; lectura de los ficheros tocados y de sus vecinos
(`BrandLogo`, `helpers.ts`, `playwright.config.ts`, contrato `schema.d.ts`, backend
`portada/service.go`); grep de todos los usos de `logoUrl`/`coverImageUrl`/`mediaUrl`; y
verificación ejecutada en local: `tsc --noEmit` limpio, `vitest run` **277/277** (45 ficheros)
y `eslint` limpio en los ficheros tocados. Árbol limpio (sin cambios pendientes).

---

## Veredicto: **APROBADO**

El arreglo es **correcto, mínimo y bien ubicado**: `mediaUrl()` tiene el comportamiento
esperado en todos los casos del contrato, se reutiliza `API_BASE_URL` sin duplicar lógica, no
rompe nada (typecheck, lint y suite completa en verde) y la prueba de regresión **no es
tautológica** (distingue el bug del arreglo). Queda **un hallazgo IMPORTANTE de cobertura de
pruebas** (I-1: 2 de los 3 call sites del fix no tienen aserción que los proteja de una
regresión parcial) y dos menores opcionales. **Ninguno requiere cambios en el código de
producción**; se recomienda cerrar I-1 en este mismo ciclo (es un cambio de una línea en tests).

---

## 1. Corrección del helper (`client.ts:10-31`)

Caso por caso frente a la lista de revisión:

| Caso | Entrada | Resultado | ¿Correcto? |
|---|---|---|---|
| Relativo | `/api/v1/media/img_logo.jpg` | `API_BASE_URL + path` | ✅ (`client.test.ts:277-281`) |
| Absoluto http/https | `http(s)://cdn…/logo.png` | intacto | ✅ (`client.test.ts:283-290`) |
| `data:` | `data:image/png;base64,AAAA` | intacto | ✅ (misma rama del regex) |
| `//host` | `//cdn.example.com/logo.png` | intacto | ✅ (`client.test.ts:292-294`); correcto por semántica HTML: `//host` es protocolo-relativa, no relativa al origen |
| `undefined` | `undefined` | `undefined` | ✅ conserva el fallback (`BrandLogo`: `src = DEFAULT_BRAND_LOGO_SRC` solo actúa con `undefined`) |
| `''` | `''` | `''` | ✅ mismo comportamiento que antes del fix; el hero lo sigue excluyendo con `showCover` |
| Doble barra | `//…` como path | rama absoluta, intacto | ✅ correcto; y `/api//v1…` no lo produce el backend (prefijo fijo `MediaPathPrefix = "/api/v1/media/"`, `backend/internal/portada/service.go:40`) |
| Base con barra final | `VITE_API_URL=http://host:8080/` | `…8080//api/…` | ⚠️ sin normalizar — ver **M-1** (misma suposición que `apiFetch`) |

- El regex `ABSOLUTE_URL_PATTERN = /^(?:[a-z][a-z0-9+.-]*:|\/\/)/i` (`client.ts:11`) cubre
  cualquier esquema RFC 3986 y `//host`, con bandera `i` (esquemas en mayúsculas). No
  clasifica mal paths que lleven `:` más adelante (`/api/v1/x:y` no matchea porque solo se
  mira el inicio). ✅
- El tipo `string | undefined` coincide exactamente con el contrato (`schema.d.ts:1251,1255`:
  `logoUrl?`, `coverImageUrl?` opcionales). Sin `any`, TS estricto. ✅
- **No rompe nada**: 277/277 unitarias en verde, `tsc --noEmit` limpio, `eslint` limpio.

## 2. Cambio mínimo y bien ubicado

- `mediaUrl` vive en `src/api/client.ts`, junto a `API_BASE_URL` que **reutiliza** (definido en
  `client.ts:8`, usado en `client.ts:29`): sin recopiar la URL base ni inventar otra fuente. ✅
- Sin duplicar lógica: una única función, usada desde los tres puntos que pintan media
  (`IdentityHero.tsx:22,38`, `PublicLayout.tsx:51`, `PublicFooter.tsx:30`). El grep confirma que
  son **todos** los usos de `logoUrl`/`coverImageUrl` en render: no hay call site olvidado. ✅
- No toca `nginx.conf`, backend ni contrato, coherente con la causa raíz declarada (el navegador
  pedía la media al origen de la SPA). ✅
- Observación: la plantilla `` `${API_BASE_URL}${path}` `` replica lo que ya hace `apiFetch`
  (`client.ts:136`). Es trivial y hoy consistente; ver **M-2**.

## 3. Sin regresiones en los componentes tocados ni sus pruebas

- `IdentityHero`: `showCover` pasa de `Boolean(identity.coverImageUrl)` a `Boolean(coverSrc)`
  con semántica idéntica (`mediaUrl` es inyectiva sobre `''`/`undefined`); el `onError` y el
  respaldo de fondo `navy` intactos. ✅
- `BrandLogo` no cambia y su contrato (`src?: string`) acepta `string | undefined`. ✅
- `homePage.test.tsx:139-145` se **fortalece** (aserciones del `src` resuelto) sin perder las
  previas; `client.test.ts` añade 4 casos del helper sin tocar los existentes. ✅
- Suite completa en verde (277/277), lint limpio: sin regresiones.

## 4. La prueba de regresión es significativa (no tautológica)

- Reproduce el bug con navegador real: siembra identidad publicada con logo y portada, mantiene
  el **control negativo** (la media directa del backend responde `200 image/*`,
  `portada-imagenes.spec.ts:113-117`, para no confundir fallo de siembra con fallo de la SPA) y
  comprueba (a) `content-type: image/*` —no `text/html` del fallback de SPA— y (b) descodificación
  real (`naturalWidth > 0`). ✅
- **Distingue bug de arreglo**: sin `mediaUrl`, el `src` se resolvería contra `:5173` y la
  petición a `browserBase(...)` (= origen de la API, `portada-imagenes.spec.ts:200-202`) nunca
  ocurriría → el `expect.poll(...).toBeDefined()` falla por timeout. La QA ya lo demostró con
  evidencia rojo/verde real (revertido = FAIL, con fix = PASS); lo confirmo por lectura. ✅
- Los ajustes de `ec2ae76` sobre la prueba original son **legítimos y no la debilitan**:
  - `browserBase` resuelve ahora contra `API_URL`: era obligatorio, porque el arreglo apunta el
    `src` al origen de la API y eso es exactamente lo que la regresión debe vigilar.
  - El logo se acota al hero (`page.getByRole('region', { name: identityName })`): el `alt` del
    logo se repite en cabecera/hero/pie (mismo `identity.logoAlt`), así que sin acotar el locator
    fallaría por modo estricto. La intención de (b) queda intacta.

---

## Hallazgos

### BLOQUEANTE
Ninguno.

### IMPORTANTE

**I-1. Las llamadas de `PublicLayout` y `PublicFooter` quedan sin protección de regresión.**
Las aserciones del `src` resuelto se acotan al hero (`homePage.test.tsx:139-145`, `within(hero)`)
y el e2e solo verifica descodificación de las imágenes del hero
(`portada-imagenes.spec.ts:163-182`). Pero `homePage.test.tsx` renderiza el `PublicLayout` completo
(línea 71: cabecera y pie con `BrandLogo` usando el **mismo** `logoUrl`) y nadie afirma su `src`.
Si se elimina `mediaUrl(...)` de `PublicLayout.tsx:51` o `PublicFooter.tsx:30` —regresión parcial
del bug: logo roto en cabecera y pie— **toda la suite seguiría en verde**. Es justo la clase de
regresión que esta entrega debería impedir.
*Propuesta concreta (una de):*
- `homePage.test.tsx`, tras las aserciones del hero, cubrir los tres usos de una vez:
  ```ts
  for (const img of screen.getAllByRole('img', { name: 'Logotipo de la iglesia' })) {
    expect(img).toHaveAttribute('src', `${API_BASE_URL}/api/v1/media/img_logo.png`);
  }
  ```
  (hoy hay 3 coincidencias: cabecera, hero y pie); o
- en el e2e, replicar el `poll` de `naturalWidth > 0` sobre los `img` de la cabecera y del pie
  antes de acotar al hero.

### MENOR

**M-1. `mediaUrl` no normaliza una base con barra final.** `client.ts:29` concatena en crudo:
con `VITE_API_URL=http://host:8080/` produciría `http://host:8080//api/…`. No es regresión —
`apiFetch` (`client.ts:136`) asume lo mismo y los valores documentados no llevan barra final
(`.env.example:57`, `frontend/Dockerfile:18`, `docker-compose.yml:109`)— pero es el único caso de
la lista sin cubrir. *Propuesta opcional*: normalizar una vez en la definición
(`.replace(/\/+$/, '')`) y beneficiar también a `apiFetch`.

**M-2. La concatenación `base + path` está duplicada en dos sitios** (`client.ts:29` y
`client.ts:136`). Hoy es consistente; si se aplica M-1, conviene extraer un `joinApiUrl(path)`
interno y usarlo en ambos para que no se desalineen. No bloquea.

**M-3. La aserción de red (a) del e2e acopla la prueba a la estrategia de este arreglo.**
`portada-imagenes.spec.ts:139-158` exige que el navegador pida la media al origen de la API; un
arreglo alternativo (p. ej. proxy nginx de `/api/` en el origen de la SPA) haría fallar la prueba
con el bug resuelto. Aceptable como regresión de *este* fix (y ya está documentado en el
comentario de `browserBase`, líneas 193-199); las aserciones (b) de descodificación son agnósticas
a la estrategia. Sin cambio obligatorio.

---

## Cierre

El fix ataca la causa raíz en el lugar correcto, con una abstracción única, tipada y reutilizada,
y con pruebas en los tres niveles (unitarias del helper, integración del `src` resuelto, e2e
red + descodificación). **APROBADO**; se recomienda cerrar I-1 (una línea de test) antes del merge
del ciclo del bug.
