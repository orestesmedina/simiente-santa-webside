# Revisión de código — 2026-10-12 (tanda 2): reorden de portada + fondo claro fuera del panel

- **Revisor:** `revisor-codigo` · **Rama:** `004-eventos-actividades`
- **Alcance:** `9aef868` (reorden del bloque «Lo que viene»), `2a70591` (fondo claro en las
  pantallas fuera del layout del panel, bug-login), `f179daa` (e2e de regresión + `ux.md` D-24 ✅).
- **Verificaciones ejecutadas:** `tsc --noEmit` ✓ · `eslint` sobre los archivos tocados ✓ ·
  `vitest run` **530/530 ✓** (78 archivos; +2 pruebas nuevas de `PanelSurfaceLayout`). El e2e
  Playwright no se ejecuta en esta revisión (requiere stack levantado); queda para `qa-tester`.

## Veredicto: **APROBADO**

Sin hallazgos bloqueantes. El arreglo del login es correcto y **no toca el sitio público**: no hay
fondo en `body` ni en `<html>`; `PanelSurfaceLayout` solo envuelve `/login`, `/sin-permiso`,
`/cambiar-contrasena` y el 404 (`frontend/src/app/router.tsx:86-92, 96-101, 139-146`), y las rutas
públicas siguen bajo `PublicLayout`, que pinta `bg-surface` con `min-h-screen`
(`PublicLayout.tsx:75`) sobre el lienzo oscuro → sin destello blanco en tema oscuro. El reorden
cumple D-24 ✅ y no rompe la portada: las secciones de F3 conservan contenido, anclas y direcciones
de entrada; la alternancia del `RevealGroup` queda ahora perfecta
(`HomePage.tsx:82-91, 92-117`: right→left→right→left→right→left).

## Bloqueante

Ninguno.

## Importante

1. **`plan.md` (P4-19, líneas 168, 311 y 456) vuelve a contradecir `ux.md` D-24 ✅ y el código.**
   El plan aún fija la ubicación «entre «Horario de servicios» y los canales de WhatsApp» y rechaza
   explícitamente «tras el `IdentityHero` (descartado por `ux.md` D-24: aplasta la identidad)» —
   que es **justo** lo que el humano confirmó el 2026-10-12 y lo que implementa
   `HomePage.tsx:82-91`. Es el espejo del hallazgo 1 de `revision-2026-10-12-codigo.md`: `ux.md` se
   actualizó en `f179daa`, pero no el plan. Un ciclo futuro que siga `plan.md` "corregiría" la
   posición hacia atrás. **Acción:** realinear P4-19 (y las líneas 311/456) con D-24 confirmada.

## Sugerencia

2. **`frontend/src/features/publico/homePage.test.tsx:165-173` — se retiró una aserción.** Antes se
   comprobaba horario→bloque→WhatsApp; ahora solo hero→bloque→quiénes→horario, con lo que
   «sin alterar F3» queda cubierto a medias (la posición de WhatsApp ya no se verifica). Propuesta:
   añadir `expect(horario.compareDocumentPosition(whatsapp)).toBe(Node.DOCUMENT_POSITION_FOLLOWING)`
   o una aserción del orden completo de secciones.
3. **`frontend/src/app/layout.test.tsx:43` — selector muerto.** `container.querySelector('#root > div')`
   nunca matchea en RTL (no hay `#root`): siempre cae en `container.firstElementChild`. Borrar el
   selector y dejar el `firstElementChild`, que es el que realmente prueba.
4. **`frontend/e2e/qa-bug-fondo-oscuro-fuera-layouts.spec.ts:56` — cobertura por debajo del docstring.**
   El comentario promete las cuatro pantallas, pero el bucle cubre tres: falta `/cambiar-contrasena`
   (requiere sesión). Añadir un caso con sesión (storageState o login en el test) o matizar el
   docstring a las tres rutas públicas.
5. **`--color-panel-ink` es un token nuevo fuera del sistema de F10 y se aplica a medias.**
   `frontend/src/index.css:27-35` duplica `slate-900` (#0f172a) para que el e2e lea `rgb()`, pero
   `ux.md` dice que F4 «no redefine tokens» y dentro de las mismas pantallas se mezcla con
   `text-slate-900`/`text-slate-700` (`router.tsx:38, 43, 55`; párrafos de `LoginPage`/
   `ChangePasswordPage`), que siguen emitiendo `oklch()`. Es defendible, pero conviene: (a) usar el
   token en toda la tinta de esas cuatro pantallas o (b) documentarlo en `ux.md` §1 como excepción
   técnica. Además el docstring de `PanelSurfaceLayout` (`layout.tsx:46`) dice `text-slate-900`
   mientras el código usa `text-panel-ink` (`layout.tsx:56`): alinear el comentario.
6. **Flash previo al montaje en `/login` (tema oscuro).** El arreglo es post-montaje de React: el
   lienzo negro sigue visible entre el HTML y el primer render. Si se quiere eliminar del todo,
   plantear un fondo estático inicial para esas rutas; no urgente.

## Notas de corrección (verificadas, no son hallazgos)

- No se debilitaron pruebas: `layout.test.tsx` solo **añade** (2 casos) y el cambio de
  `homePage.test.tsx` actualiza la aserción de orden al nuevo D-24 (ver sugerencia 2).
- `PanelSurfaceLayout` mantiene la capa de presentación en `app/layout.tsx`, sin lógica de negocio
  ni dependencias nuevas; `fondoPintado`/`luminancia` del e2e resuelven la cadena de ancestros y no
  atan el arreglo a un elemento concreto — buena prueba de regresión.
