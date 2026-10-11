# Revisión QA — ciclo 2 — F10 Sistema de diseño y movimiento (2026-10-11)

**Validador:** `qa-tester` · **Rama:** `010-sistema-diseno-movimiento` · **Ciclo:** 2 (re-validación tras la corrección de IMP-1 de `revision-2026-10-11-qa.md`) · **Nuevos commits revisados:** `1339f3e` (I1), `00bf59f` (M1-seguridad), `71f9729` (M2-revisor), `dfa03bf` (pruebas), `abb28a2` (C2).

## Resumen ejecutivo

| Veredicto | Bloqueantes | Importantes | Menores |
|---|---|---|---|
| **APROBADO** | **0** | **0** | **2** |

- **IMP-1 (selector de idioma «menos espacio que el actual») queda CERRADO**: medido sobre el render real contra el stack, la huella pasó de **188 × 50 px** a **152 × 44 px**, por debajo de la referencia de F3 (**159 × 44 px**) en **−7 px de ancho y 0 px de alto**, sin perder legibilidad ni objetivo táctil. La corrección quita el borde y reduce el relleno (`gap-0.5`, `px-0.5` en pastilla y opciones); el texto sigue en **14 px** y cada opción mide **81 × 44 px**.
- La suite completa queda en verde: unit **407/407** (+5 nuevas), lint, typecheck y build ✔, e2e de F10 **25/25**, e2e congeladas de F2/F3 **8/8** (SC-008).
- El equipo además **reforzó las pruebas con lo que pedí**: aserción dura `ancho ≤ 159 / alto ≤ 44 px` contra la referencia F3 en `diseno-movimiento` («el selector de idioma ocupa menos que el de F3 y sigue legible (IMP-1)»), y las pruebas de QA del ciclo 1 (`e2e/qa-medidas-f10.spec.ts`) quedaron **intactas** (verificadas frente a mi entrega: ni una aserción debilitada).

## Ejecuciones (evidencia)

| Ejecución | Resultado |
|---|---|
| `npm test -- --run` (unit, 65 archivos) | ✅ **407/407** |
| `npm run lint` / `npm run typecheck` / `npm run build` | ✅ en verde |
| Medida del selector en vivo (Chromium en imagen Docker, 320 px y 1280 px) | ✅ **152 × 44 px** (pastilla), borde 0, texto 14 px, opción 81 × 44 px |
| e2e nuevas F10 + pruebas de QA del ciclo 1 (`diseno-movimiento`, `movimiento`, `movimiento-reducido`, `navegacion-movil`, `tema-oscuro`, `qa-medidas-f10`), tanda 1, `--workers=1`, imagen `mcr.microsoft.com/playwright:v1.63.0-jammy` contra el stack real | ✅ **25/25** |
| e2e congeladas F2/F3 (`acceso`, `auditoria`, `portada-publica`, `portada-accesibilidad`, `portada-panel`, `portada-imagenes`, `portada-patch-null`, `status`), tanda 2 | ✅ **8/8** |

Nota de entorno: igual que en el ciclo 1, rate-limit M-4 → dos tandas; sin errores (tanda 1 usa 6 suites ≤ 20 logins/min).

## Cambios revisados desde el ciclo 1 (diff `54cbe60..HEAD`)

| Cambio | Verificación de QA |
|---|---|
| **fix I1** (`IdentityHero`/`HomePage`): el CTA «Ver horario» del hero es condicional a la sección publicada | e2e `navegacion-movil` reforzada: con horario y WhatsApp en borrador, 0 `a[href="#horario"]` y 0 `a[href="#whatsapp"]` ni en nav ni en pie (sin enlaces rotos, FR-027) — ✔ pasa |
| **fix M2-revisor** (`Reveal.tsx`/`useReveal.ts`): un bloque ya revelado o nacido mostrado no se vuelve a ocultar al apagar `reduced-motion` en caliente; el salvavidas de 1 s se cancela con la primera entrada del observer y un bloque fuera de pantalla **sigue pendiente** hasta entrar | `movimiento.spec` reforzada: tras 1,5 s el bloque fuera de pantalla sigue pendiente (no se auto-revela) y al entrar revela con las métricas de FR-005 — ✔ pasa; `movimiento-reducido.spec` (FR-006) sigue ✔ |
| **fix M1-seguridad** (`publicSeed.ts`): el `storageState` del e2e se escribe en directorio temporal propio `0700`/archivo `0600` y se borra al salir | solo pruebas; verifico que no toca las congeladas (no las usa) — ✔ |
| **fix C2** (`PublicButton.tsx`): hover del botón primario con tokens `accent-hover`/`on-accent-hover`; «selectores de tema» documentados en README | unit `PublicButton.test.tsx` +42 líneas — ✔; la congelada de contraste sigue ✔ |
| **pruebas** (`dfa03bf`): 5 unit nuevas (402→407) + e2e de medida IMP-1 y refuerzos | revisadas una a una: **ninguna prueba congelada editada** y las mías intactas — ✔ |

**SC-008 re-verificado:** diff desde el ciclo 1 sin tocar `backend/`, panel, `App.tsx`, e2e congeladas ni `.env.example`; las 8 e2e congeladas pasan intactas.

## Matriz FR/SC (estado del ciclo 2)

La matriz completa sigue en `revision-2026-10-11-qa.md`; aquí el **delta de los criterios afectados por las correcciones del ciclo 2**:

| Criterio | Ciclo 1 | Ciclo 2 | Evidencia nueva |
|---|---|---|---|
| **FR-015 / SC-007** (primera cláusula «menos espacio que el actual») | ❗ IMP-1 fallaba | **✔** | medida en vivo 152 × 44 < 159 × 44 (F3); aserción dura en e2e `diseno-movimiento` IMP-1, ✔; umbrales M4 de `qa-medidas-f10` ✔ (14 px texto, 44 px táctil, huella ≤ 56 px → 44 px) |
| **FR-005 / SC-014** (revelado al scroll, además del refuerzo C1) | ✔ | ✔ reforzado | `movimiento.spec`: bloque fuera de pantalla **sigue pendiente** tras el salvavidas y se revela solo al entrar (el revelado de SC-014 ya no se auto-anula a los 1 s) |
| **FR-008** (sin re-animaciones molestas) | ✔ | ✔ | fix M2: apagar `reduced-motion` en caliente no re-oculta ni re-anima lo ya visto |
| **FR-027 / SC-015** (0 enlaces rotos) | ✔ | ✔ reforzado | e2e con horario y WhatsApp en borrador: 0 enlaces a secciones ausentes en nav, pie y CTA del hero |
| **FR-003 / SC-011** (tokens de hover documentados) | ◐ | ✔ sube un peldaño | tokens `accent-hover` + unidad + README de cimientos actualizado |
| Resto de FR/SC (FR-001–FR-027, SC-001–SC-015 no listados arriba) | como en el ciclo 1 | sin cambios | misma cobertura: 8/8 congeladas + 25/25 nuevas |

Estados acumulados de las partes humanas de los criterios (sin cambios respecto al ciclo 1): SC-001/SC-002/SC-003/SC-006 (fluidez)/SC-009 (contraste fino en oscuro)/SC-013 «0 copias»/SC-014 (sensación final) y FR-012 (previsión F4–F9) siguen **pendientes de verificación humana o de secciones reales** — ninguno impide el aprobar técnico.

## Hallazgos del ciclo 2

- **BLOQUEANTES:** ninguno.
- **IMPORTANTE:** ninguno (IMP-1 cerrado y verificado por medida).
- **MENORES:**
  1. **MEN-3 · Menor — El CTA «Ver horario» del hero depende ahora de la publicación de la sección**: la presencia del texto varía con el estado del contenido; cubierto por e2e y unit (`IdentityHero.test.tsx` actualizado). Informativo: si en F10B alguna pantalla del panel muestra los CTA, tendrá que respetar la misma condición.
  2. **MEN-2 (arrastrado de ciclo 1) · Menor — Rate-limit M-4 de login** (`POST /api/v1/auth/login`, 20/min por IP): sin cambios; se corre en dos tandas. Volverá a importar cuando la suite crezca más.

## Veredicto

**APROBADO** — 0 bloqueantes, 0 importantes.

IMP-1 queda **cerrado**: el selector de idioma de F10 (152 × 44 px) es ahora más compacto que el de F3 (159 × 44 px) manteniendo el texto a 14 px y el objetivo táctil a 44 px, verificado por medida en vivo y por la aserción dura añadida al e2e. Las demás correcciones del ciclo (I1, M2-revisor, M1-seguridad, C2) están cubiertas por pruebas reforzadas y no introducen regresión ninguna: toda la suite (unit 407, e2e 25 + 8 congeladas, lint/typecheck/build) en verde.

Queda pendiente para la revisión de aceptación del humano (no bloquea la parte técnica): SC-002 «moderno», SC-003/SC-007 con personas, SC-001/SC-011 visual, SC-009 contraste fino en oscuro, SC-013 «0 copias», SC-014 sensación final y FR-012 con las secciones F4–F9 cuando existan.
