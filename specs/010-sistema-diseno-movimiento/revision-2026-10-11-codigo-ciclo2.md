# Revisión de código — F10 · Ciclo 2 (re-validación de correcciones)

- **Fecha**: 2026-10-11 · **Revisor**: `revisor-codigo` (Fase 5, ciclo 2)
- **Alcance**: commits de corrección `48dd0e9` (C1), `f240ea0` (C2), `dd4c5ce` (IMP-1 de QA), `1339f3e` (I1), `00bf59f` (M1-seguridad), `71f9729` (M2), `dfa03bf` (pruebas QA) y `abb28a2` (docs), contra `revision-2026-10-11-codigo.md`.
- **Base verificada**: suite unitaria re-ejecutada por mí: **407/407 en verde** (+5 pruebas nuevas frente al ciclo 1: C1, C2, I1, M2 e IMP-1). SC-008 intacto: **0 borrados** en cualquier `*.test.tsx`/`*.spec.ts` respecto a `main` (las 2 líneas eliminadas por `dd4c5ce` en `i18n.test.tsx` pertenecen al `describe` añadido por F10 en esta rama; ninguna aserción congelada de F2/F3 se tocó).

## Cierre de los hallazgos del ciclo 1

| Hallazgo | Estado | Evidencia |
|---|---|---|
| **C1** salvavidas revelaba todo a los 1 s | ✅ **CERRADO** (no mitigado) | `useReveal.ts` ahora usa `handlers = Map<Element, {seen, reveal}>`: **toda** entrada del observer —intersecte o no— llama `seen()` y cancela el temporizador; el salvavidas solo cubre un observer que nunca reporta. Prueba unitaria que fija el caso exacto (bloque fuera del viewport sigue pendiente tras `REVEAL_FALLBACK_MS * 3` y se revela al entrar) y guarda e2e en `movimiento.spec.ts` (espera 1500 ms y exige estado pendiente) — ya no es una carrera. |
| **C2** hover del primario AA en oscuro | ✅ **CERRADO** | Tokens por tema `--accent-hover`/`--on-accent-hover` (claro `#0e7a66`/`#ffffff` = 5,26:1; oscuro `#00c9a7`/`#1a2b4a` = 6,66:1 — verificado por mí), `PublicButton` usa `hover:bg-accent-hover hover:text-on-accent-hover`, y `PublicButton.test.tsx` calcula la razón de contraste de **ambos** pares desde los tokens (≥ 4,5:1) con aserción de «un valor por tema». |
| **I1** CTA «Ver horario» a sección ausente | ✅ **CERRADO** | `scheduleHref?: string` condicional en `IdentityHero`, provisto por `HomePage` solo con servicios publicados; prueba unitaria espejo («sin horario publicado no pinta el CTA») y e2e que, con horario en borrador, exige `a[href="#horario"]` y `a[href="#whatsapp"]` con conteo 0 (0 enlaces rotos, FR-027/SC-015). |
| **M2** re-ocultar al apagar reduced-motion | ✅ cerrado en el caso citado (con residual mínimo) | `observed` nace `true` si hay movimiento reducido o no hay `IntersectionObserver` (latch al montar), con prueba. Ver N4 para el residual. |

Correcciones colaterales revisadas y correctas: `dd4c5ce` (huella del selector, con e2e midiendo ≤ 159 × 44 px frente a F3 y ≥ 44 px táctil/≥ 14 px de texto), `00bf59f` (storageState del e2e en directorio `0o700`/archivo `0o600`, borrado en `exit`), `dfa03bf` (e2e de medidas de QA).

## Hallazgos

### CRÍTICO

**N1 · Regresión de contraste AA introducida por `dd4c5ce`: la opción de idioma ACTIVA queda a 4,05:1 en tema claro**
`frontend/src/features/publico/i18n/LanguageSwitcher.tsx:35-41` + `frontend/src/index.css` (tokens de tema claro).
Al quitar el borde y poner `bg-surface-raised` en el `<fieldset>`, la pastilla activa `bg-accent/15 text-accent` (14 px `font-bold` → **no** es texto grande; umbral 4,5:1) se compone ahora sobre el triple de capas —`--surface` → `bg-surface-raised` (8 % blanco) → `bg-accent/15`— y el resultado es **≈ 4,05:1** (`text-accent` `#00c9a7` sobre un fondo efectivo ≈ `#255164`). Antes de `dd4c5ce` era ≈ 5,03:1 (la pastilla caía sobre `--surface`). Viola FR-024 (AA «en todos los estados… en ambos temas»), SC-012 («0 fallos de contraste AA») y la constitución (WCAG 2.1 AA no negociable). El auditor de contraste del e2e congelado **no lo detecta** (no resuelve fondos con alfa — el propio `README.md` §2.2 lo advierte), así que la suite sigue verde. En tema oscuro es 6,33:1 (OK) y la opción inactiva (`text-muted`) 5,83:1 (OK).
**Corrección (≈ 1 línea, preferir par sólido según el criterio del propio README)**:
- Opción A (más fiel a `ux.md` §0.6 «fondo turquesa suave y peso 700»): `text-ink` en la pastilla activa (quedaría ≈ 7,7:1), o
- Opción B (misma gramática que `PublicButton primary`, SC-011): activo `bg-accent text-navy font-bold` (≈ 6,7:1).
Y añadir el par de la pastilla activa (fondo compuesto vs texto) al test de contraste estilo `PublicButton.test.tsx` (C2), para que ninguna corrección futura vuelva a abrir el hueco.

### NIT

**N1a · Fragilidad TDZ en el salvavidas**: `frontend/src/features/publico/motion/useReveal.ts` — `seen` hace `window.clearTimeout(timer)` y `timer` se declara con `const` **después** de `io.observe(element)`. Un observador con callback síncrono (polyfill no spec) dispararía un `ReferenceError` en la zona muerta temporal. Declarar `let timer: number | undefined` antes (o guardarlo). Con el `IntersectionObserver` real es seguro.

**N2 · Hover del primario sin cambio de color en tema oscuro**: `PublicButton.tsx:32` — en oscuro, `--accent-hover`/`--on-accent-hover` (`#00c9a7`/`#1a2b4a`) son idénticos al estado de reposo (`bg-accent text-navy`): la retroalimentación de hover queda solo en el lift `hover:-translate-y-0.5` (FR-008 pide retroalimentación clara; el lift la da, pero conviene un matiz de color distinto en oscuro, p. ej. un borde o un tono ligeramente distinto de turquesa).

**N3 · Prueba dependiente de `process.cwd()`**: `frontend/src/components/publico/PublicButton.test.tsx` (lectura de `src/index.css`). Funciona con `make test-frontend`/CI (`cd frontend && npm test`), pero revienta si vitest se lanza desde la raíz. Preferir una ruta relativa al propio módulo (`new URL('../index.css', import.meta.url)`) o el helper `css` compartido que ya usa `motion.test.tsx`.

**N4 · Residual mínimo de M2**: con el sitio abierto, la secuencia `prefers-reduced-motion` OFF → ON → **OFF** vuelve a poner un bloque no observado en estado pendiente (`Reveal.tsx:46-52`): lo que se mostró estáticamente durante la fase ON se oculta un frame y re-anima al apagarse. Latch completo: en el efecto, cuando `prefersReduced` sea `true`, fijar `setObserved(true)` (así «una vez mostrado, nunca se re-oculta»).

**N5 · Aserción vaga en la prueba de I1**: `IdentityHero.test.tsx` — `queryByRole('link', { name: '#horario' })` no comprueba nada real (ningún enlace se llama así); la aserción efectiva es la de `'Ver horario'`. Renombrar o eliminar.

## Veredicto

**RECHAZADO** — 1 bloqueante:

- **N1**: la corrección de IMP-1 (`dd4c5ce`) introdujo una regresión de contraste AA en la opción de idioma activa en tema claro (4,05:1 < 4,5:1, estado permanente y visible; FR-024/SC-012/constitución).

**C1, C2 e I1 quedan cerrados de forma real** (con sus pruebas de regresión unitarias y e2e). El resto del ciclo está limpio: 407/407 unitarias, SC-008 intacto y sin cambios de alcance. Con la corrección de N1 (una línea + un par de contraste en pruebas) y los NIT aceptables o atendidos, la rama merece aprobación.
