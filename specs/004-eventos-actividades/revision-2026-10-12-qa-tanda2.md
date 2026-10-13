# Revisión QA — F4 Eventos y actividades, re-validación tanda 2 (2026-10-12)

**Rol**: `qa-tester` · **Rama**: `004-eventos-actividades` · **Alcance**: `9aef868` (reorden del bloque «Lo que viene»), `2a70591` (fondo claro fuera de los layouts), `f179daa` (e2e de regresión).
**Veredicto: APROBADO** (0 defectos; sin pruebas adicionales necesarias).

## Entorno y ejecución de suites

| Suite | Comando | Resultado |
|---|---|---|
| E2e bug fondo oscuro (nuevo, este alcance) | Docker `mcr.microsoft.com/playwright:v1.63.0-jammy`, `--workers=1`, `qa-bug-fondo-oscuro` | ✅ **3 passed (7,5 s)** |
| E2e tanda A (F3/F10 + bloque, reorden) | ídem, `portada- diseno- movimiento navegacion- tema- status` | ✅ **33 passed (1,2 m)** |
| E2e tanda B (F4) | ídem, `agenda- qa-bugs-004` | ✅ **8 passed (28,5 s)** |
| E2e tanda C (F2) | ídem, `acceso auditoria` | ✅ **2 passed (18,1 s)** |
| Frontend unitarias | `npm test -- --run` | ✅ **78 archivos / 530 pruebas** (antes 528: +2 del shell claro en `layout.test.tsx`) |
| Lint + typecheck | `npm run lint && npm run typecheck` | ✅ sin errores |
| Backend unitarias | `go test ./...` | ✅ 20 paquetes ok |
| Backend integración (`DATABASE_URL_TEST` exportada) | `go test -count=1 -tags=integration -p 1 ./...` | ✅ 21/21 paquetes ok |
| Cobertura `internal/agenda` | `-tags=integration -cover` | ✅ **87,0 %** (≥ 80 % constitución §III) |

Tandas separadas por el rate-limit de `POST /api/v1/auth/login` (20/min); ningún `429`.

## Matriz de cobertura — Arreglo del fondo negro fuera de los layouts (2a70591 + f179daa)

| Criterio | Prueba(s) | Estado |
|---|---|---|
| `/login` con tema oscuro del dispositivo: fondo claro no transparente y título oscuro (legible) | e2e `qa-bug-fondo-oscuro-fuera-layouts.spec.ts` caso 1 (`colorScheme: 'dark'` sin `ss.theme`; `colorScheme=dark` en `<html>`; primer fondo pintado de la cadena ancestros = `rgb(255,255,255)`; luminancia del título < 0,5) | ✅ |
| `/sin-permiso` ídem | ídem, caso 2 (`section`, «No tienes acceso a esta sección») | ✅ |
| 404 (`/esta-ruta-no-existe-404`) ídem | ídem, caso 3 (`main`, «Página no encontrada») | ✅ |
| `/cambiar-contrasena` ídem (no automatizable en e2e sin sesión previa) | unidad `layout.test.tsx` «pinta la superficie clara del panel» (`bg-white` + `text-panel-ink`) y «acepta contenido explícito … (404 fuera del panel)»; verificación estructural: `router.tsx` envuelve las 4 rutas en `PanelSurfaceLayout` (`/login`, `/sin-permiso`, `/cambiar-contrasena` como layout route bajo `RequireAuth`; 404 con contenido explícito en el catch-all `*`) | ✅ |
| No se debilitó nada: el e2e mide repintado real (no grep de clases), y la aserción resuelve la cadena de ancestros, no un elemento concreto | e2e propio + `lint`/`typecheck` limpios | ✅ |

## Matriz de cobertura — Reorden del bloque «Lo que viene» (9aef868)

| Criterio | Prueba(s) | Estado |
|---|---|---|
| Orden nuevo: hero → bloque → «quiénes somos» → «horario»; el resto de F3 intacto | unidad `homePage.test.tsx` (comparación `compareDocumentPosition` entre los 4 encabezados, aserciones previas sin quitar) | ✅ |
| La portada sigue pintando / el reorden no rompe F3 ni el bloque | e2e tanda A íntegra: `portada-publica`, `portada-proximamente` (los 4 casos, T449), `portada-imagenes`, `portada-accesibilidad`, `portada-panel`, `portada-patch-null`, `theme`-oscuro, movimiento/movimiento-reducido, navegación-movil, `diseno-`, `status` — **33/33** | ✅ |
| Sin regresión en el resto (panel, accesibilidad F4, auditoría F2) | e2e tandas B (8) y C (2) en verde | ✅ |

## Defectos de este alcance

**Ninguno.** No fue necesario escribir pruebas adicionales: la de regresión del fondo (`f179daa`) cubre las tres rutas automatizables y la unitaria de `layout.test.tsx` cubre la cuarta; el reorden ya tenía su aserción de orden en la unitaria y la portada queda cubierta por la tanda A.

Notas heredadas (sin cambios): `session`/`testcontainers` mantienen la contención conocida como infra (73,7 s); pendientes manuales SC-010/SC-015 con personas y lector de pantalla físico.

## Veredicto final

**APROBADO.** Las 46 pruebas e2e de las 4 tandas, 530 unitarias de frontend y las suites Go (unitarias + integración, cobertura 87,0 %) están en verde, con lint y typecheck limpios. El fondo claro fuera de los layouts está probado con repintado real en tema oscuro y el reorden del bloque «Lo que viene» no rompe la portada ni ninguna suite. Puede avanzar a entrega (`documentador`/`devops`).
