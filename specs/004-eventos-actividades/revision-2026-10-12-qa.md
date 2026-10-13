# Revisión QA — F4 Eventos y actividades, re-validación (2026-10-12)

**Rol**: `qa-tester` · **Rama**: `004-eventos-actividades` · **Alcance**: commits desde `ef93c77` (3 bugs corregidos: `e1f6fed`/`4348875`, `72665e2`, `66c8472` + bloque «Lo que viene» de la portada: `3619436`, `3ebb183`, `e74d086`, `907d6cd`).
**Veredicto: APROBADO** (0 bloqueantes; 2 pendientes manuales heredados sin cambios + 1 nota de entorno)

## Entorno y ejecución de suites

| Suite | Comando | Resultado |
|---|---|---|
| Backend unitarias | `go test ./...` | ✅ 20 paquetes ok |
| Backend integración (PostgreSQL real, `DATABASE_URL_TEST` del `.env`) | `go test -count=1 -tags=integration -p 1 ./...` | ✅ 21/21 paquetes ok (agenda 4,6 s con **139 subpruebas** corriendo sin ningún SKIP; `portada` 17,8 s; `session`/testcontainers contención conocida como infra, no de F4) |
| Cobertura `internal/agenda` | `-tags=integration -cover ./internal/agenda/` | ✅ **87,0 %** (≥ 80 % de la constitución §III) |
| Frontend unitarias | `npm test -- --run` | ✅ **78 archivos / 528 pruebas** (antes 74/505: +4 archivos y +23 pruebas nuevas, todas de este alcance) |
| Lint + typecheck | `npm run lint && npm run typecheck` | ✅ sin errores |
| E2e tanda A (F3/F10 + bloque: `portada- diseno- movimiento navegacion- tema- status`) | Docker `mcr.microsoft.com/playwright:v1.63.0-jammy`, `--workers=1` | ✅ **33 passed (54,6 s)** — incluye los 4 casos de `portada-proximamente.spec.ts` (T449) |
| E2e tanda B (F4: `agenda- qa-bugs`) | ídem | ✅ **8 passed (22,9 s)** — incluye los 3 casos de `qa-bugs-004.spec.ts` |
| E2e tanda C (F2: `acceso auditoria`) | ídem | ✅ **2 passed (14,2 s)** — sin regresión tras los arreglos |

Las tandas se repartieron por el rate-limit de `POST /api/v1/auth/login` (20/min por IP, heredado del 2026-10-11); no hubo `429` en ninguna corrida.

## Matriz de cobertura — Bugs corregidos

| Bug | Prueba(s) de reproducción | Estado |
|---|---|---|
| **Bug 1** — desde `/eventos` el menú no navega a las anclas de la portada (`href` relativos) | **Unidad**: `qa-b1-siteLinks-relativas.test.tsx` (cada ancla es `/#ancla`; ningún enlace queda relativo; el ítem de agenda sigue siendo `/eventos` sin doble barra). **E2e**: `qa-bugs-004.spec.ts` «Bug 1» (click «WhatsApp» desde `/eventos` → URL `/#whatsapp`, sección visible y en viewport). **E2e heredado**: `navegacion-movil.spec.ts` usa anclas absolutas y honra el ancla al llegar desde otra ruta (fixes `e1f6fed`/`4348875`) | ✅ pasa |
| **Bug 2** — menú sin hover ni opción actual | **Unidad**: `qa-b2-PublicNav-activo.test.tsx` (`aria-current="page"` en la ruta actual en `/eventos`, la portada no marca nada, cada enlace lleva clases `hover:`). **E2e**: `qa-bugs-004.spec.ts` «Bug 2» (repinto real: `background-color` cambia al `hover()`) | ✅ pasa |
| **Bug 3** — modal de «Agregar evento» pequeño y con doble scroll | **E2e**: `qa-bugs-004.spec.ts` «Bug 3» (agrega 2 fechas, mide `width ≥ 640 px` y `scrollWidth ≤ clientWidth`: sin scroll horizontal). Fix `66c8472` (`size="wide"`, `overflow-x-hidden`, `minmax(0,1fr)`), sin debilitar `Dialog.test.tsx` (rol, Escape, Cerrar intactos) | ✅ pasa |
| Los arreglos no rompen `/eventos`, la sección ni el panel | e2e tanda B íntegra: `agenda-publica` (visitante/borradores/inglés/anchos/aviso), `agenda-panel` (3 casos: gestión, categorías, permisos cruzados + auditoría), `agenda-accesibilidad` (teclado, foco, claro/oscuro); tanda C sin rotos de auditoría | ✅ pasa |

## Matriz de cobertura — Bloque «Lo que viene» de la portada (US7, FR-025–FR-027, SC-014/SC-015)

| Criterio | Prueba | Estado |
|---|---|---|
| AS-1: hasta 5 ítems, eventos y actividades juntos, mezclados y etiquetados, tarjetas con imagen + `alt` | e2e `portada-proximamente` caso 1 (8 ítems alternados → exactamente 5 listitems, los 5 más cercanos en orden; «Evento»/«Actividad» visibles; 2 imágenes con su `alt` en viewport); backend `TestGetPortadaProximamenteFixesLimitAndWindow`, `TestIntegrationListNextUpcomingItemsWindowAndBackup`; unidad `ProximamenteBlock.test.tsx` (etiqueta de concepto, orden) | ✅ |
| AS-2 (ventana de 2 meses preferente) | e2e caso 2 (cercano a +10 días dentro, lejano a +100 días fuera → **solo** «Cercano»); backend unitarios `TestGetPortadaProximamenteFixesLimitAndWindow`, `TestPortadaWindowEndAddsCalendarMonths` (2 meses de calendario) + integración | ✅ |
| AS-3 (respaldo fuera de la ventana) | e2e caso 2 (eliminado el cercano, entra «Lejano» aunque caiga fuera); `TestGetPortadaProximamenteWindowWithBackup` + integración (nunca se mezclan) | ✅ |
| AS-4 (menos de 5: exactos los que hay, sin huecos) | unidad «con menos de 5 no deja huecos ni marcadores vacíos»; e2e caso 1 (retirado un ítem, el siguiente próximo ocupa su lugar y siguen siendo 5 listitems exactos) | ✅ |
| AS-5 (sin nada: bloque con mensaje y **sin botón**) | e2e caso 3 (`role=status` «Por el momento no hay eventos ni actividades.», 0 `listitem`, 0 enlace «Ver más», y borrador próximo + actividad vencida invisibles con la nav de la sección oculta); unidad «sin ningún ítem muestra el mensaje»; backend `TestGetPortadaProximamenteWithoutItems` + handler omite `items` vacía | ✅ |
| AS-6 (botón «Ver más eventos» → `/eventos` en un paso) | e2e casos 1 y 4 (URL `/eventos`, `h1` «Eventos y actividades»; enlace con `href=/eventos` en ambos idiomas) | ✅ |
| AS-7 (vencidos/borrador/sin fecha futura no aparecen) | e2e caso 3 (draft próximo y vencido a −3 días invisibles); backend `TestGetPortadaProximamenteFiltersDraftsExpiredAndWithoutNextDate`; `TestGetProximamenteAppliesTheNextRule` heredado | ✅ |
| AS-8 (inglés con fallback por contenido, nada vacío) | e2e caso 4 («What's coming», «See more events»; «Friday vigil» traducido y «Repartir comida» pintado en español sin campo vacío; `html lang=en` ↔ `es` sin pérdida); backend `TestGetPortadaProximamenteLocalizesAndResolvesImage`; unidad de i18n/en del bloque | ✅ |
| AS-9 (responsivo a 320 px en columna sin scroll horizontal; teclado con foco visible; toque ≥ 44 px) | e2e caso 4 (bounding boxes: segunda tarjeta apilada bajo la primera en columna, sin scroll ≥ 1 px; anillo de foco en el título-enlace; click del título → `/eventos#actividades` visible en viewport = SC-015); `portada-accesibilidad` (F10: **mide todos los enlaces de `main`** y detectó el defecto → fix `e74d086` `min-h-11`) | ✅ |
| FR-027 sin rediseñar la portada de F3 | las 33 pruebas de la tanda A pasan íntegras (hero, quiénes somos, contacto, tema oscuro, movimiento, navegación móvil, sitemap); `HomePage.test.tsx` antiguo actualizado sin quitar aserciones | ✅ |

**Defectos de este alcance: 0.** El e2e del bloque, además, ya destapó y corrigió el objetivo táctil de 44 px (T448) antes de esta re-validación.

## Nota de entorno (no producto)

La primera pasada de integración **sin** `DATABASE_URL_TEST` exportada saltó silenciosamente las pruebas de repositorio (`t.Skip` de `testutil`): `agenda` "pasó" en 0,08 s. Con la variable exportada (como documentan `quickstart.md` y la bitácora) corrió completo. No es defecto de F4, pero conviene que `make test`/`make ci` la exporte o falle de forma explícita si falta (utilidad, para otro ciclo).

## Pendientes manuales (sin cambios, heredados)

- **SC-010/SC-015 (encuesta con personas)**: la parte automatizable de SC-015 (un solo paso desde portada a `/eventos` con vídeo y navegación coherente) está en verde; el protocolo de personas queda para el humano.
- **Lector de pantalla físico** (NVDA/VoiceOver): lo automatizable (foco visible, `aria-` correctos, jerarquía `<h2>/<h3>`, `<time>`, `alt`) en verde; la lectura real queda manual, ahora también del bloque de la portada (nueva superficie).

## Veredicto final

**APROBADO.** Los 3 bugs reportados están corregidos y cada uno tiene su prueba de reproducción en verde a nivel unidad y/o e2e, con repintado/mediciones reales y sin debilitar ninguna prueba existente. El bloque «Lo que viene» de la portada cumple US7/FR-025–FR-027 y SC-014/SC-015 con cobertura end-to-end y en las 3 capas (cliente API, hook, componente + service/repository sobre PostgreSQL), con la ventana de 2 meses con respaldo, el tope firme de 5 y el mensaje de «sin contenido». Las suites completas pasan: Go unitarias + integración (cobertura 87,0 %), Vitest 528 pruebas y 43 casos e2e en 3 tandas sin `429`. Puede emitirse el PR vía `documentador` con los pendientes manuales anotados y re-aprobar el alcance Q8 cuando el humano lo decida.
