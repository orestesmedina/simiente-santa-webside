# Revisión QA — F10 Sistema de diseño y movimiento (2026-10-11)

**Validador:** `qa-tester` · **Rama:** `010-sistema-diseno-movimiento` · **Alcance:** `git diff main...HEAD` (F10 completo) contra `specs/010-sistema-diseno-movimiento/spec.md` (FR-001–FR-027, SC-001–SC-015).

## Resumen ejecutivo

| Veredicto | Bloqueantes | Importantes | Menores |
|---|---|---|---|
| **APROBADO** | **0** | **1** | **2** |

- **26 de 28 criterios FR/SC con cobertura automatizada al menos parcial**; 2 quedan enteramente pendientes de verificación humana (SC-002, SC-003, que por definición requieren cliente/personas).
- Suite completa en verde: unit **402/402**, e2e congeladas **8/8**, e2e nuevas de F10 **22/22**, e2e de QA añadidas en esta revisión **2/2**. Lint, typecheck y build en verde.
- **El IMPORTANTE es medible y reproducible: el selector de idioma compacto ocupa MÁS espacio que el diseño F3 que debía compactar** (FR-015/SC-007). Se devuelve a `dev-frontend` (arreglo trivial de espaciado) o a decisión del cliente si la pastilla con borde se acepta como tal.

## Ejecuciones (evidencia)

| Ejecución | Resultado |
|---|---|
| `npm test -- --run` (unit, 65 archivos) | ✅ **402/402** |
| `npm run lint` / `npm run typecheck` / `npm run build` | ✅ en verde |
| e2e nuevas F10 (`diseno-movimiento`, `movimiento`, `movimiento-reducido`, `navegacion-movil`, `tema-oscuro`) vía imagen Docker `mcr.microsoft.com/playwright:v1.63.0-jammy` contra el stack real, `--workers=1` | ✅ **22/22** |
| e2e congeladas F2/F3 (`acceso`, `auditoria`, `portada-publica`, `portada-accesibilidad`, `portada-panel`, `portada-imagenes`, `portada-patch-null`, `status`), segunda tanda por el rate-limit de login M-4 | ✅ **8/8** |
| e2e de QA añadidas en esta revisión (`e2e/qa-medidas-f10.spec.ts`) | ✅ **2/2** |

Notas de entorno (no regresiones): Chromium no arranca en WSL (como en F1–F3, se usó la imagen Docker de Playwright); en la primera tanda el directorio montado omitía `resources/` (error de invocación, `ENOENT /resources/simiente.jpeg`, corregido montando la raíz del repo); el rate-limit M-4 de login (20/min por IP) obligó a correr en dos tandas y no afectó al resultado.

**SC-008 (no regresión):** `git diff main...HEAD` sobre `e2e/` congeladas, `backend/`, `docker-compose.yml`, `Makefile`, `.env.example`, `App.tsx` y el panel es **vacío** (ni una línea tocada), y las 8 e2e congeladas **pasan intactas** contra el sitio rediseñado — incluida la de accesibilidad de F3, que ahora verifica **el nuevo diseño** (contraste ≥ 4.5:1 de `main` y pie, teclado, foco, 44 px, 200 %). 0 pruebas congeladas editadas.

## Pruebas añadidas por QA (solo archivos de prueba, 0 cambios de producción)

- `frontend/e2e/qa-medidas-f10.spec.ts` (2 nuevas):
  1. **Cierre de la calibración M4 de `analyze.md`** («se cierra en QA»): el selector compacto medido sobre estilos **computados reales** a 320 px — texto ≥ 14 px (medido: 14 px), altura de línea ≥ 20 px, objetivo táctil ≥ 44 px (medido: 44 × ~93 px por opción), huella de la pastilla ≤ 56 px (medido: 50 px).
  2. **FR-014 completado**: el botón del menú hamburguesa queda visible y operable **al final de la página** en móvil (las e2e previas comprobaban la cabecera al inicio y la barra en escritorio); y en escritorio los enlaces de la barra siguen a la vista en el fondo. Cubre el hueco que quedaba de «la navegación permanece alcanzable en el 100 % de la página mientras se baja» (SC-004).

## Matriz de cobertura (criterio → prueba → resultado)

✔ automatizado y pasa · ◐ parcial (parte automatizada pasa; el resto requiere humano) · ○ pendiente humano

| Criterio | Pruebas que lo cubren | Resultado |
|---|---|---|
| **FR-001** sistema visual moderno sobre tokens F3 + Manual | unit de componentes (`PublicButton`, `Card`, `SectionHeading`, `GradientBlobs`…), `index.css` con tokens; «moderno» → SC-002 | ◐ |
| **FR-002** logotipo sin deformar/recolorear/girar/efectos; proporción de marca | e2e `tema-oscuro` «logotipo conserva sus colores originales» (sin filtros, pastilla crema); congelada `portada-imagenes` | ✔ |
| **FR-003** componentes base reutilizables, sin dos estilos | unit de base en `components/publico` + README de cimientos; «sin dos estilos» → revisión visual (SC-001) | ◐ |
| **FR-004** legibilidad para todas las edades | congelada `portada-accesibilidad` (cuerpo grande, contraste, 200 %); **M4 cerrado** con `qa-medidas-f10` (14 px / 44 px medidos) | ✔ |
| **FR-005** movimiento marcado/elegante con revelado al scroll | `movimiento.spec`: suelo de perceptibilidad ≥ 400 ms/16 px, techo ≤ 800 ms, escalonado ≤ 240 ms, cabecera que transiciona, ancla que desliza con título visible; «elegante» final → SC-014 humano | ✔ ◐ |
| **FR-006** `prefers-reduced-motion` total (incondicional) | `movimiento-reducido.spec`: 0 revelados/delays/transiciones, scroll instantáneo, cabecera sin transición, blobs sin animar; cambio en caliente sin recargar | ✔ |
| **FR-007** movimiento no degrada carga/fluidez | `movimiento.spec` «la animación no bloquea» + `movimiento-reducido` (100 % disponible); dispositivo modesto no instrumentado → humano | ◐ |
| **FR-008** retroalimentación sin saltos; sin repeticiones molestas | `movimiento.spec` «cambiar de idioma o de tema no re-anima»; estados hover/foco en unit; calidad perceptual → humano | ✔ ◐ |
| **FR-009** nada depende de ver una animación | `movimiento-reducido.spec` «100 % del contenido y funciones disponibles» + Edge del foco (bloques revelados al recibir foco, e2e T1030) | ✔ |
| **FR-010** barra fija escritorio / hamburguesa accesible móvil | `diseno-movimiento` (barra, anclas a la vista, sin hamburguesa en escritorio); `navegacion-movil` (44 px, `aria-expanded`, foco al entrar, Escape→foco al botón, scrim, sin scroll de fondo, sin trampa, anuncio de estado) | ✔ |
| **FR-011** fin del «scroll incómodo» (lateral + al bajar) | `diseno-movimiento` «sin desplazamiento horizontal a 320/360/768/1024/1280», `navegacion-movil`, y FR-014 abajo | ✔ |
| **FR-012** previsión patrón F4–F9 (~8–10 enlaces) | no verificable hoy: el modelo solo tiene 5 secciones; nav/mapa derivan de datos («modelo de secciones») según plan | ○ (diseño) |
| **FR-013** navegación accesible (teclado/lector, foco visible) | congelada `portada-accesibilidad` + ciclo de foco del menú (`navegacion-movil`) | ✔ |
| **FR-014** alcanzable al bajar: barra fija / botón siempre visible; no tapa | `diseno-movimiento` (barra al bajar, ancla que no tapa el título); `navegacion-movil` (data-scrolled en ambos modos); **completado por QA** (`qa-medidas-f10`: hamburguesa operable al fondo en móvil) | ✔ |
| **FR-015** selector compacto, textos completos, sin píldoras/desplegable | unit `i18n.test` (radios «Español/Inglés», sin siglas ni combobox); e2e `diseno-movimiento`; **«menos espacio» NO se cumple** → hallazgo IMP-1 | ◐ ❗ |
| **FR-016** comportamiento de F3 intacto (inmediato, memoria, primera visita es) | e2e `diseno-movimiento` (cambio inmediato, `ss.lang`, recarga, re-visita); unit «sin auto-detección»; congelada `portada-publica` (fallback es) | ✔ |
| **FR-017** selector sin desplazamientos ni tapar en ningún estado | unit (inline-flex, sin anchos forzados); e2e coexistencia a 320 px sin solapes; 0 scroll horizontal | ✔ |
| **FR-018** aplicado a toda la portada + cimientos documentados | e2e recorren todas las secciones del sistema; cimientos: `frontend/src/components/publico/README.md` + `ux.md`; homogeneidad exacta → SC-001 visual | ◐ |
| **FR-019** cero textos de interfaz sin traducir | unit `i18n.test` «claves i18n nuevas de F10 (FR-019/SC-010)»: es/en, no vacías, catálogo de `ux.md` §7 | ✔ |
| **FR-020** no regresión F3 / panel intacto | e2e congeladas 8/8 + diff vacío de backend/panel/App/`App.tsx`/`LanguageProvider`/`sections.ts` | ✔ |
| **FR-021** sin datos nuevos | diff vacío en `backend/` y migraciones | ✔ |
| **FR-022** accesibilidad AA en ambos temas | claro: congelada (contraste/44 px/foco/200 %); oscuro: control operable + funciones idénticas; **auditoría completa en oscuro** → SC-009 humano | ◐ |
| **FR-023** oscuro: sistema + interruptor + memoria | `tema-oscuro.spec`: los tres caminos (sistema, manual, primera visita) + persistencia | ✔ |
| **FR-024** contraste AA ambos temas; logotipo original | logotipo ✔ (ambos temas); **0 fallos de contraste en oscuro** sin prueba automática → humano | ◐ |
| **FR-025** cabecera convive a 320 px sin taparse | `diseno-movimiento` «marca, idioma, tema y hamburguesa conviven sin taparse» (sin solapes) | ✔ |
| **FR-026** proceso 3 propuestas, elegida B, sin copias | registro en spec/propuestas + aprobación puerta 2; «0 copias» → comparación visual humana | ✔ ◐ |
| **FR-027** mapa del sitio del pie sin rotos ni vacíos | `navegacion-movil` (grupos Explorar/Contacto/Síguenos, nombres únicos, enlace que funciona sin taparse; WhatsApp en borrador desaparece de nav **y** mapa) | ✔ |
| **SC-001** 100 % secciones con el sistema, 0 excepciones | recorridos e2e por sección + unit; «0 excepciones de estilo» es revisión visual | ◐ |
| **SC-002** cliente: «moderno», no típico de IA; ≥ 80 % personas | no automatizable | ○ humano |
| **SC-003** ≥ 90 % personas llegan sin ayuda | no automatizable | ○ humano |
| **SC-004** 0 scroll horizontal y nav 100 % alcanzable | `diseno-movimiento` (320–1280), `navegacion-movil`, `qa-medidas-f10` (al fondo) | ✔ |
| **SC-005** con reduced-motion 0 animaciones perceptibles | `movimiento-reducido.spec` (medido 0 delays/duraciones/opacity<1) | ✔ |
| **SC-006** interactivo desde el primer instante; fluido | e2e «animación no bloquea»; «dispositivo de referencia más modesto» → humano | ◐ |
| **SC-007** selector más compacto y 100 % identificación | textos/nombres/siglas ✔ (unit + e2e); **«menos espacio que el actual» falla** (IMP-1); identificación por personas → humano | ◐ ❗ |
| **SC-008** 0 regresiones F2/F3 | e2e congeladas 8/8 + 0 líneas tocadas (ver arriba) | ✔ |
| **SC-009** 0 hallazgos críticos accesibilidad, claro y oscuro | automatizado: claro✔, oscuro parcial; auditoría humana de cierre | ◐ |
| **SC-010** 0 cadenas nuevas sin traducir | unit `i18n.test` (es/en) + e2e inglés («Page sections», radios «English») | ✔ |
| **SC-011** coherencia transversal + cimientos documentados | componentes base unit + README/`ux.md`; «0 doble estilo» revisión visual | ◐ |
| **SC-012** oscuro: sistema/persistencia/logotipo; 0 diferencias | `tema-oscuro.spec` (100 % de los caminos, contenido idéntico, logotipo sin filtros); contraste AA oscuro → humano | ✔ ◐ |
| **SC-013** 3 propuestas, 1 aprobada antes de codificar; 0 copias | registro en puerta 2 con fecha; «0 copias» comparación visual humana | ✔ ◐ |
| **SC-014** movimiento claramente perceptible; 0 rebotes | `movimiento.spec` (suelo de perceptibilidad, no re-anima); sensación «marcado y elegante» → humano | ✔ ◐ |
| **SC-015** mapa del sitio: 100 % visibles, 0 rotos | `navegacion-movil` «mapa del sitio del pie» | ✔ |

**Balance:** 18 ✔ · 10 ◐ · 2 ○ (ambos por definición humanos: SC-002, SC-003). Ningún criterio automatizable quedó sin prueba ni falló en su cobertura automatizada salvo la cláusula «menos espacio» de FR-015/SC-007 (IMP-1).

## Hallazgos

### IMP-1 · IMPORTANTE — El selector de idioma ocupa MÁS espacio que el diseño que debía compactar (FR-015 / SC-007)

- **Qué exige la spec:** FR-015 «DEBE ocupar menos espacio que el actual, compactando los textos actuales mediante una reducción de tamaño y espacio»; SC-007 «ocupa menos espacio en cabecera que el actual (comparación directa en la misma vista de referencia)».
- **Cómo se midió (comparación directa en la misma vista de referencia):** render real del header contra el stack, mismo navegador y viewport (320 px y 1280 px):
  - diseño **actual (main, F3)**: **159 × 44 px** (medido desde fin `main` con `vite preview` en worktree);
  - diseño **F10 (rama)**: **188 × 50 px** (medido igual).
  Resultado: **+29 px de ancho (+18 %) y +6 px de alto**. Causa: las opciones pasaron de `px-1` a `px-2`, el grupo añadió borde + `px-1.5 py-0.5`, aunque el hueco entre opciones se redujo (`gap-3 → gap-1`); el texto mantiene `text-sm` (14 px), tamaño que no se redujo.
- **Pasos para reproducir:** abrir `/` (320 px u 1280 px), inspeccionar `header fieldset` → 188×50; contra `main` → 159×44.
- **Esperado según spec / obtenido:** huella ≤ 159×44 → **188×50**.
- **Impacto:** la primera cláusula de SC-007 no se cumple y el objetivo de la queja del cliente (US4, «cabecera respira») queda a medias: la pastilla con borde sí añade identificación y no rompe coexistencia ni legibilidad (M4 cerrado: 14 px de texto, 44 px táctil). No es bloqueante de seguridad ni de datos; es desviación de un criterio medible explícito.
- **Propuesta de arreglo (a `dev-frontend`):** márgenes de la pastilla (`px-2`→`px-1` o sin `border` + `py-0.5`), o escalar al cliente si la pastilla con borde es lo que quiere (decisión revisable de Q5 «default revisable»). Tras el arreglo, `e2e/qa-medidas-f10.spec.ts` puede añadir la aserción dura `ancho ≤ 159` (no la añadí ahora para no acoplar la prueba excesivamente al diseño).

### MEN-1 · MENOR — La previsión F4–F9 (FR-012 / parte de SC-004) no es probable hasta que existan secciones

Con ~8–10 secciones el patrón debería seguir cumpliéndose, pero hoy el modelo solo tiene 5 secciones y no hay modo de sembrar más. La nav y el mapa derivan del modelo de secciones (una fuente para ambos), lo que respalda el diseño; queda verificado functional cuando existan secciones reales. Sin acción antes del PR.

### MEN-2 · MENOR — M-4 de F2 heredado: rate-limit de login para los e2e

`post /api/v1/auth/login` limita 20/min por IP (incidencia conocida de F2, F10 no la introduce ni agrava más allá del número de suites). Los corre en dos tandas. No requiere acción de F10.

## Pendientes de verificación humana (no automatizables)

1. **SC-002**: calificación del cliente («moderno», «no típico de IA») — checklist de revisión de aceptación en `quickstart.md` §9.
2. **SC-003 / SC-007 (personas)** y **SC-006 (fluidez en dispositivo modesto)**: pruebas de control del público con personas.
3. **SC-001 / SC-011 (visual)** y **SC-013 (0 copias de las referencias)**: revisión visual comparativa de la implementación contra B «Marea Turquesa».
4. **SC-009 / FR-024 en tema oscuro**: auditoría de contraste AA en oscuro de cierre (el automatizado cubre logotipo, funciones y operabilidad en oscuro; el contraste fino de cada superficie en oscuro no tiene prueba automática).
5. **SC-014**: sensación final del nivel «marcado/elegante» (ni estridente ni mareante) con el cliente.
6. **FR-012**: re-verificar el patrón con las secciones reales F4–F9 cuando nazcan.

## Veredicto

**APROBADO** — 0 bloqueantes.

La implementación de F10 está en verde en toda la suite (unit 402, e2e 22 nuevas + 8 congeladas + 2 de QA, lint/typecheck/build) y cumple las historias automatizables: navegación por dispositivo con menú hamburguesa accesible completo, tema claro/oscuro con los tres caminos y sin flash, movimiento perceptible acotado y totalmente suprimido con `prefers-reduced-motion`, mapa del sitio sin enlaces rotos, bilingüismo completo de las cadenas nuevas y cero regresiones de F2/F3 (SC-008, 0 pruebas congeladas tocadas).

El hallazgo **IMP-1 (FR-015/SC-007: la pastilla del selector es más grande que el diseño que compacta)** debe devolverse a `dev-frontend` para corrección o escalar al cliente antes de la puerta de PR; es un cambio de una línea de espaciado, no de arquitectura. Los pendientes de verificación humana se muestran como tales para la revisión de aceptación.
