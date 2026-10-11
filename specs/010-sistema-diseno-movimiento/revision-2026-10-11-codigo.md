# Revisión de código — F10 «Sistema de diseño y movimiento»

- **Fecha**: 2026-10-11 · **Revisor**: `revisor-codigo` (Fase 5 · Validación)
- **Alcance**: `git diff main...HEAD` (rama `010-sistema-diseno-movimiento`, 82 archivos, +9 739/−317)
- **Referencias**: `spec.md`, `plan.md`, `ux.md` (diseño final B), `tasks.md` (D15–D21), `.specify/memory/constitution.md`

## Resumen de verificaciones base

| Comprobación | Resultado |
|---|---|
| SC-008 · pruebas congeladas F2/F3 sin editar | ✅ 0 borrados en todos los `*.test.tsx`/`*.spec.ts` (solo `describe` nuevos anexados a `homePage.test.tsx` e `i18n.test.tsx`; las 8 suites e2e congeladas ni se tocan). Suite unit re-ejecutada por mí: **402/402 en verde** (65 archivos). |
| Sin cambios de backend / datos / API | ✅ Fuera de `frontend/` y `specs/` solo cambia `docs/producto/roadmap.md`. Sin migraciones, sin contratos, sin `.env`. |
| 0 dependencias nuevas | ✅ `package.json` intacto (sin librerías de animación/menú). |
| TS `strict` sin `any` injustificado | ✅ Sin `any`/`as any` en el código nuevo; casts acotados y comentados. |
| Sin `dangerouslySetInnerHTML` | ✅ |
| `prefers-reduced-motion` en doble capa | ✅ bloque global de `index.css` (T1004) + `usePrefersReducedMotion` en vivo (T1017); blobs, revelados, hero, delays y `scroll-behavior` anulados. |
| Decisiones D15–D21 | ✅ verificadas en código (una sola lista en el `<nav>`; *disclosure* sin `role="dialog"`/`aria-modal` ni trampa de foco; colapso `display:none` bajo media query sin atributos `hidden`/`aria-hidden`/`inert`; mapa sin grupos vacíos y rótulos `<p>` sin rol heading; nombres del pie con prefijo/planos; marca «Simiente» con `sr-only` a < 480 px; tokens de movimiento únicos — 600 ms / 28 px (20 px ≤ 480) / stagger 80 ms tope 240 ms / micro 180 / hero 700 / cabecera 220 — y e2e con techo ≤ 800 ms). |
| Estructura y documentación | ✅ primitivas en `components/publico/` con `README.md` heredable para F4–F9; componentes cortos y con nombres claros; i18n es/en completa (13 claves nuevas, verificadas en `i18n.test.tsx`). |

## Hallazgos

### CRÍTICO

**C1 · El «salvavidas» de 1 s revela todo el contenido fuera del viewport: el revelado al scroll solo funciona en el primer segundo**
`frontend/src/features/publico/motion/useReveal.ts:66-78` (temporizador incondicional en la línea 78; el callback compartido de `getObserver()` líneas 39-52 ignora las entradas `isIntersecting: false` sin cancelarlo).
`observeReveal()` arma `window.setTimeout(reveal, REVEAL_FALLBACK_MS)` **siempre**, aunque el `IntersectionObserver` sí haya reportado el elemento (fuera de pantalla). Como todos los `Reveal` de `HomePage` montan a la vez, **1 s después de la carga cada bloque pasa a `data-reveal='in'` esté o no visible**: el visitante que baja la página ya no ve ningún revelado (FR-005/SC-014, el corazón de la funcionalidad: «el 100 % de las secciones… se revela al entrar en pantalla»). Además, los e2e que deberían detectarlo no pueden: `movimiento.spec.ts` («arranca pendiente») es una carrera contra ese segundo, y las aserciones `toHaveAttribute('data-reveal','in')` tras hacer scroll (`navegacion-movil.spec.ts:263`, `movimiento.spec.ts:73-80`) pasan trivialmente aunque el revelado ya se haya producido fuera de pantalla.
**Corrección**: que el salvavidas solo cubra «observador que nunca reporta». P. ej., ampliar el mapa compartido a `{ reveal, seen }` y, en el callback del observer, llamar `seen()` (limpia el temporizador) para **toda** entrada recibida —intersecte o no— revelando solo si `isIntersecting`. Prueba que lo fije: renderizar un bloque sin disparar intersección y avanzar `REVEAL_FALLBACK_MS` con el observer **habiendo** reportado `isIntersecting: false` → debe seguir **pendiente**; y e2e que espere > 1 s antes de exigir el estado pendiente y compruebe que el paso a `in` ocurre al entrar en pantalla.

**C2 · Hover del botón primario incumple AA en tema oscuro (≈ 2,1:1)**
`frontend/src/components/publico/PublicButton.tsx:32` — `primary: 'bg-accent text-navy hover:bg-accent-strong hover:text-white'`. En tema oscuro `--accent-strong` vale `#00c9a7` (igual que `--accent`, `index.css` bloque `html[data-theme='dark']`), así que el hover deja **texto blanco sobre turquesa** (contraste ≈ 2,1:1; ni siquiera AA-large 3:1). FR-024 exige AA «en todos los estados (reposo, hover, foco, activo…)» **en ambos temas** y SC-012 pide «0 fallos de contraste AA en los elementos revisados en ambos temas»; la constitución hace de WCAG 2.1 AA una regla no negociable. La auditoría de contraste del e2e congelado solo recorre el estado de reposo, por lo que no lo detecta (y el propio `README.md` §2.2 advierte de los falsos negativos con alfa).
**Corrección**: mantener `text-navy` en hover dentro del tema oscuro (p. ej. `hover:bg-accent-strong hover:text-white dark:hover:text-navy`, o un token `--accent-hover` + `--on-accent-hover` por tema). Añadir una aserción del par de colores del hover en ambos temas (unitaria sobre tokens o e2e de contraste con `page.hover()`).

### IMPORTANTE

**I1 · El CTA «Ver horario» del hero apunta a `#horario` aunque la sección no exista (enlace roto con horario en borrador)**
`frontend/src/features/publico/components/IdentityHero.tsx:85` — `<PublicButton href="#horario">` se renderiza **siempre**, pero `HomePage.tsx` solo monta `ScheduleSection` cuando hay servicios publicados (y `useSiteLinks`/`FooterSiteMap` excluyen el ancla). Con todo el horario en borrador queda un enlace muerto a `#horario`, en contradicción con el edge case de la spec («sin dejar enlaces rotos, enlaces huérfanos ni animaciones apuntando a algo que ya no existe», FR-027/SC-015) y con el propio criterio ya aplicado al CTA de WhatsApp (`whatsappHref` condicional, con la prueba «sin canal de WhatsApp publicado no pinta el CTA (0 enlaces rotos)», `IdentityHero.test.tsx:59`).
**Corrección**: hacer el CTA condicional igual que el de WhatsApp —p. ej. `scheduleHref?: string` que `HomePage` solo pase cuando `schedule?.length`— y añadir la prueba espejo («sin horario publicado no pinta el CTA»).

### MENOR

**M1 · `RevealGroup` inyecta `revealIndex` a cualquier hijo React**
`frontend/src/features/publico/motion/RevealGroup.tsx:29-35` — `isValidElement` acepta **cualquier** elemento; un hijo que no sea `<Reveal>` (un `<div>` plano) recibe `revealIndex` como atributo desconocido en el DOM (warning de React) sin efecto. Hoy los usos de `HomePage` son correctos, pero F4–F9 va a reutilizarlo.
**Corrección**: clonar solo cuando `child.type === Reveal` (o tipar `children: ReactElement<RevealProps>[]` con guard) y devolver el resto tal cual.

**M2 · Al apagarse `prefers-reduced-motion` en caliente, un bloque no visto se oculta y re-anima**
`frontend/src/features/publico/motion/Reveal.tsx:43-46,73` — `revealed = observed || prefersReduced || !supportsObserver`: si la preferencia pasa de `reduce` a `no-preference` con el sitio abierto, un bloque aún no observado vuelve a `data-reveal='left|up'` (el CSS `html.js` lo pinta a opacidad 0) y luego re-anima. El contenido visible desaparece un instante (FR-008: sin parpadeos ni ocultar información; edge case «la preferencia de movimiento cambia con el sitio abierto»).
**Corrección**: no volver a estado pendiente una vez que el bloque ya se mostró — memorizar en el componente que la preferencia reducida estuvo activa (o fijar `observed = true` al desactivarse) — y cubrirlo con una prueba del cambio en caliente en esa dirección.

### NIT

**N1 · Título de prueba que sobrepromete**: `frontend/src/features/publico/components/FooterSiteMap.test.tsx:115` — «…no duplican los de las secciones» solo se refiere a Contacto/Síguenos (los de «Explorar» repiten a propósito los nombres de sección, D19/tasks.md línea 290). Renombrar para que describa lo que mide (unicidad dentro del pie + prohibiciones de D19).

**N2 · `preventDefault()` global de Escape**: `frontend/src/features/publico/components/useMobileMenu.ts:72-78` — con un *disclosure* basta con cerrar; no cancelar el `Escape` del documento entero para no interferir con otros usos futuros.

**N3 · `phoneHref` duplicado** en `ContactSection.tsx:11` y `FooterSiteMap.tsx:19` (y constantes `LINK` casi idénticas en `ContactSection.tsx:21`/`FooterSiteMap.tsx:24`): extraer a un módulo común del feature (p. ej. `features/publico/contact.ts`).

**N4 · Coste sostenido de los blobs**: `frontend/src/index.css` (`.public-blob`, `blur(70px)` + `animation: … 26s/32s infinite alternate` sobre elementos de 46vw/38vw) — pintura continua en dispositivos modestos (FR-007/RG10-9 ya prevé degradar). Verificar en el dispositivo de referencia y, si degrada, bajar el blur o congelar la deriva manteniendo la capa decorativa.

## Veredicto

**RECHAZADO** — 2 bloqueantes:

1. **C1** (`useReveal.ts:78`): el salvavidas destruye el revelado al scroll más allá del primer segundo (FR-005/SC-014), y la e2e correspondiente no lo detecta por diseño de sus aserciones.
2. **C2** (`PublicButton.tsx:32`): fallo de contraste AA en hover del botón primario en tema oscuro (FR-024/SC-012, constitución).

Con C1 y C2 corregidos (más su prueba de regresión) y los IMPORTANTE/MENOR atendidos o aceptados por el orquestador, la rama está en condiciones de aprobar: la no-regresión de F2/F3 está certificada (SC-008), el alcance se cumple y la base para F4–F9 es sólida y coherente con D15–D21.
