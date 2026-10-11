# Estado: 010-sistema-diseno-movimiento

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 010-sistema-diseno-movimiento |
| Flujo | equipo-feature |
| Fase | 5/9 · Validar |
| Ciclo de corrección | 3/3 |
| Próximo paso | Corregir la regresión N1 (contraste de la opción activa del selector en tema claro) y re-validar (ciclo 3); si aprueba, convergencia y entrega |
| Bloqueado por | — |
| Actualizado | 2026-10-10 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano | 2026-10-10 | "Apruebo la spec" (con ajuste de alcance: F10 = sitio público + modo oscuro con interruptor manual; panel → F10B) |
| Plan | aprobado | humano | 2026-10-10 | "Apruebo el plan" (con los ajustes de diseño pedidos: más aire entre paneles, mapa del sitio en el pie y menú hamburguesa en móvil) |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- **N1 (BLOQUEANTE, revisor, ciclo 2)** — la corrección de IMP-1 (`dd4c5ce`) puso `bg-surface-raised` en el `<fieldset>` y la opción **activa** del selector (`bg-accent/15 text-accent`) queda a **4,05:1** en tema claro (< AA). Ver `revision-2026-10-11-codigo-ciclo2.md`. Corregir con un par sólido.
- **Cerrados en el ciclo 2**: C1, C2, I1, IMP-1 (confirmados por `revisor-codigo` y `qa-tester`). `seguridad` sigue APROBADO.

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-10 — **Arranque de F10 (decisión del humano en sesión).** Ante la pregunta «¿qué funcionalidad arrancamos?», el humano eligió **F10 — Sistema de diseño y movimiento** antes que F4, coherente con la Decisión 11 del roadmap (construirlo antes de F4 para que las secciones futuras nazcan ya con el sistema). Se arranca el flujo `equipo-feature` en la rama `010-sistema-diseno-movimiento`.
- 2026-10-10 — **Aclaraciones de F10 resueltas por el cliente** (registradas en `spec.md`, sección «Aclaraciones resueltas (2026-10-10)»; también decisión 12 del roadmap): **Q1** F10 incluye también el **panel de administración** (no solo el sitio público); **Q2** el «scroll incómodo» eran **ambos** casos (enlaces que se desplazan de lado en móvil y navegación que se pierde al bajar); **Q3** navegación en **barra fija siempre visible**; **Q4** animación **moderada** (respetando `prefers-reduced-motion`); **Q5** el selector **solo compacta los textos actuales** (sin píldoras ni desplegables, comportamiento de F3 intacto); **Q6** **se permite crear iconografía/ilustraciones nuevas** dentro de la marca; **Q7** **se incluye el modo oscuro** (sitio y panel); **Q8** referencias de inspiración (no copiar) `iglesiaoasis.com` y `river-arena.org`, y el cliente quiere **3 propuestas de diseño distintas e innovadoras** antes de implementar.
- 2026-10-10 — **Cierre de la puerta 2 (confirmación del cliente).** El cliente **confirmó** el ajuste de spec (menú hamburguesa en móvil) y el **diseño final B**; además decidió: el **selector de idioma no se repite dentro del panel del menú móvil** (solo cabecera y pie) y los **rótulos del mapa del sitio** son **Explorar / Contacto / Síguenos**. Fase 2 cerrada; arranca la Fase 3 (tareas + coherencia).
- 2026-10-10 — **Microcopy del CTA del hero (F10).** Para no duplicar el nombre accesible «Escribir por WhatsApp» (que pertenece a la sección de WhatsApp y está **congelado** en las pruebas de F3), el CTA del hero usa un texto propio («Escríbenos por WhatsApp»). Es una divergencia menor frente a la maqueta B; si el cliente exige el texto exacto, se escala (SC-008 prohíbe editar pruebas existentes).

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-11 — **Ciclo 2 (re-validación)**: `qa-tester` **APROBADO** (IMP-1 cerrado: selector 152×44 px < F3), `seguridad` **APROBADO** (M1 cerrado), `revisor-codigo` **RECHAZADO** por **N1**: la corrección de IMP-1 introdujo una regresión de contraste (opción activa del selector a 4,05:1 en tema claro, < AA). Arranca el ciclo 3.
- 2026-10-11 — **Ciclo de corrección 1 aplicado**: `dev-frontend` corrigió **C1** (el salvavidas del revelado solo cubre un observador que no reporta; el bloque fuera de pantalla permanece pendiente), **C2** (hover del botón primario con par AA por tema: 5,3:1 claro / 6,7:1 oscuro), **IMP-1** (selector compacto 152×44 px, menor que F3) e **I1** (CTA «Ver horario» condicional); además M1-seguridad y M2-revisor. Suites: **407** unit + e2e **8/8** congeladas + **25/25** F10. Re-validación en curso.
- 2026-10-11 — **Fase 4 (Implementar) cerrada — 36/36 tareas**: T1001–T1036 en verde. Unit **402** pruebas; e2e **8/8** congeladas + **22/22** nuevas; lint/typecheck/build en verde; **0 pruebas congeladas editadas** (SC-008); **0 cambios** en backend/migraciones/API/infra/`.env.example`; cierre en `cierre.md`. Bugs reales corregidos por los e2e (overflow lateral 28 px a 1024 px, scroll de fondo del menú móvil, contraste del selector). Notas de entorno: rate-limit de login (M-4, e2e en dos tandas) y Playwright vía imagen Docker. Arranca la Fase 5 (validar).
- 2026-10-10 — **Fase 3 cerrada**: `tasks.md` (36 tareas T1001–T1036, 20 `[P]`) y `analyze.md`. El primer `analyze` **rechazó** (4 CRÍTICOS + 3 IMPORTANTES: colisiones con pruebas congeladas de F2/F3); corregidos por `arquitecto` y `disenador-ux` (una sola lista en el nav, patrón *disclosure* sin trampa de foco, colapso fuera del tab order, mapa sin grupos vacíos, nombres accesibles únicos, marca «Simiente» con `sr-only`, recuento de tabulación < 30). Re-análisis: **APROBADO** (0 críticos). Queda **M4** (umbral de tamaño del selector) como MENOR para QA. Arranca la Fase 4 (implementar).
- 2026-10-10 — **Diseño final B aplicado**: `ux.md` reescrito como diseño final de **B «Marea Turquesa»** (más aire entre secciones con tokens de espaciado, **mapa del sitio** en el pie con grupos accesibles, **menú hamburguesa accesible** en móvil ≤ 860 px) y `propuesta-b.html` reescrita; `plan.md`/`research.md`/`quickstart.md` alineados (navegación por dispositivo, componente de mapa del sitio, tokens de espaciado, e2e del menú). Spec actualizada: FR-010 a FR-014 por dispositivo, **FR-027** (mapa del sitio) y **SC-015** nuevos. Pendiente: confirmar el ajuste de spec.
- 2026-10-10 — **Puerta 2 (iteración 2): plan aprobado y propuesta elegida (B).** El cliente aprobó el plan («Apruebo el plan») y eligió la propuesta **B «Marea Turquesa»** con tres cambios: (1) **más aire entre paneles** (hoy se sienten pegados); (2) **mapa del sitio en el pie** con todas las opciones (referencia Apple/Microsoft); (3) en **móvil, la navegación se oculta en un menú hamburguesa** (en escritorio sigue visible). Mantiene el orden actual de secciones. El punto (3) **cambia FR-010 de la spec** («sin menús ocultos»): se actualiza la spec y se re-confirma con el humano.
- 2026-10-10 — **Iteración de diseño 2 completada**: `arquitecto` refuerza el plan de movimiento (revelados de **600 ms** con desplazamiento 20–28 px y *stagger* de 80 ms; suelo de perceptibilidad ≥ 400 ms/16 px y techo ≤ 800 ms; cabecera que transiciona al bajar; `scroll-behavior: smooth` solo sin reduced-motion; **0 dependencias nuevas**); `disenador-ux` entrega **3 propuestas nuevas** —A «Bento Vivo», B «Marea Turquesa», C «Huella Tipográfica»— y reescribe `ux.md` y las maquetas con revelado real por scroll. Pendiente: **puerta 2** (iteración 2).
- 2026-10-10 — **Puerta 2 — iteración de diseño 1 (feedback del cliente).** El cliente **no aprobó el plan** y **rechazó las 3 propuestas** (A/B/C). Pidió **más animación**: transiciones entre componentes y **revelado de las secciones al hacer scroll** (las actuales le parecieron «muy pobres, casi imperceptibles»). Añadió referencias de inspiración: `apple.com`, `microsoft.com/es-cr` y `river-arena.org` (además de `iglesiaoasis.com`). Implica: subir el nivel de animación (de moderado a **marcado/elegante con revelado al hacer scroll**), ajustar el plan de movimiento y **rehacer las 3 propuestas** con una dirección nueva. Sigue vigente: respetar `prefers-reduced-motion`, no bloquear contenido y no regresar F2/F3.
- 2026-10-10 — **Fase 2 completada**: `plan.md`, `research.md`, `data-model.md` (sin cambios de datos), `contracts/README.md` (sin cambios de API) y `quickstart.md` (`arquitecto`); `ux.md` con **3 propuestas** —A «Raíz Viva», B «Cielo Abierto», C «Luz Serena»— más maquetas HTML navegables en `propuestas/` (`disenador-ux`). Sin dependencias nuevas (movimiento con CSS puro); tema con `data-theme` + variables semánticas. Pendiente: **puerta 2** (aprobar el plan y elegir una propuesta).
- 2026-10-10 — **Puerta 1 superada:** el humano aprobó la spec (**«Apruebo la spec»**) con dos ajustes de alcance: **F10 = sitio público + modo oscuro** (con **interruptor manual claro/oscuro**) y el **panel se separa a F10B**. Aprobación registrada en «Aprobaciones». Roadmap: F10 ajustada y **F10B** añadida (decisión 13). Arranca la Fase 2 (plan + 3 propuestas de diseño).
- 2026-10-10 — **8 aclaraciones resueltas con el cliente** (Q1–Q8): F10 se amplía a **panel de administración + modo oscuro**, navegación en **barra fija siempre visible**, animación **moderada**, selector de idioma **solo compactado**, **iconografía/ilustraciones nuevas permitidas**, referencias de inspiración y **3 propuestas de diseño a presentar antes de implementar**. Spec actualizada por `analista-producto`. Pendiente: aprobación de la spec (puerta 1).
- 2026-10-10 — Inicio de F10: rama `010-sistema-diseno-movimiento` y spec redactada por `analista-producto` (5 historias US1–US5, 25 escenarios, 22 FR, 11 SC, 13 Edge Cases). Quedan **8 preguntas abiertas (Q1–Q8)** que cambian alcance o experiencia; se aclaran con el humano antes de aprobar la spec.
