# Revisión QA — ciclo 3 (final) — F10 Sistema de diseño y movimiento (2026-10-11)

**Validador:** `qa-tester` · **Rama:** `010-sistema-diseno-movimiento` · **Ciclo:** 3 (re-validación tras corregir los NIT del ciclo 2) · **Nuevos commits revisados:** `880f840` (N1: par sólido de la opción activa), `386be2e` (NIT: TDZ del timer, latch de `reduced-motion`, hover en oscuro), `cfaaa71` (docs).

## Resumen ejecutivo

| Veredicto | Bloqueantes | Importantes | Menores |
|---|---|---|---|
| **APROBADO** | **0** | **0** | **2** (heredados, sin cambios) |

- **IMP-1 sigue cerrado:** huella del selector medida en vivo en las 4 combinaciones (claro/oscuro × 320/1280 px): **152 × 44 px** en todas, sin borde, por debajo de la referencia F3 (159 × 44 px), texto 14 px, opción 81 × 44 px.
- **El nuevo par sólido de la opción activa cumple AA con margen, verificado por contraste calculado sobre el render real: 6,66:1** (fondo `rgb(0, 201, 167)` = `--accent`, texto `rgb(26, 43, 74)` = navy), igual en ambos temas (el acento no cambia de valor entre claro y oscuro) y en los dos anchos. Reproduce el valor declarado por `dev-frontend`.
- Suite completa en verde: unit **409/409** (+2), lint/typecheck/build ✔, e2e F10+QA **25/25**, e2e congeladas F2/F3 **8/8** (SC-008).

## Ejecuciones (evidencia)

| Ejecución | Resultado |
|---|---|
| `npm test -- --run` (unit, 65 archivos) | ✅ **409/409** |
| `npm run lint` / `npm run typecheck` / `npm run build` | ✅ en verde |
| Medida in vivo del selector (claro/oscuro × 320/1280 px, Chromium en imagen Docker contra el stack reconstruido) | ✅ huella **152 × 44 px**; contraste opción activa **6,66:1** (AA ≥ 4,5) en los 4 combos |
| e2e F10 + QA (`diseno-movimiento`, `movimiento`, `movimiento-reducido`, `navegacion-movil`, `tema-oscuro`, `qa-medidas-f10`), tanda 1, `--workers=1` | ✅ **25/25** |
| e2e congeladas F2/F3 (`acceso`, `auditoria`, `portada-publica`, `portada-accesibilidad`, `portada-panel`, `portada-imagenes`, `portada-patch-null`, `status`), tanda 2 | ✅ **8/8** |

Nota de entorno: rate-limit M-4 → dos tandas, sin errores.

## Cambios revisados desde el ciclo 2 (diff `abb28a2..HEAD`)

| Cambio | Verificación de QA |
|---|---|
| **N1** (`LanguageSwitcher.tsx`: activa pasa de `bg-accent/15 text-accent` a **`bg-accent text-navy`** sólido) | medición in vivo (arriba): **6,66:1** en claro y oscuro — ✔; unit nueva «la opción activa usa un par de contraste sólido y cumple AA en ambos temas (N1)» con el helper `src/test/contrast.ts` — ✔. La identidad de marca se respeta: usa el mismo par que el botón primario (accent → navy), dentro de la regla «colores derivados solo por contraste» del FR-002 |
| **helpers de contraste** (`PublicButton.test.tsx` −30 líneas → `src/test/contrast.ts`) | refactorización pura: las aserciones de AA del hover (`accent-hover`/`on-accent-hover` ≥ 4,5:1 en ambos temas) quedan **idénticas**, solo cambia el origen del helper — ✔ sin debilitamiento |
| **TDZ del timer** (`useReveal.ts`): `timer` declarado antes de los callbacks | unit + e2e del movimiento en verde; el salvavidas ya no puede morir en la zona muerta temporal — ✔ |
| **latch de «una vez mostrado, no se re-oculta»** (`Reveal.tsx`): cierre de estado durante el render (patrón de estado derivado de React); apagar `reduced-motion` en caliente no re-oculta ni re-anima lo ya visto | `motion.test.tsx` +14 líneas (nueva prueba del latch) — ✔; e2e `movimiento-reducido` (FR-006) y `movimiento` (FR-008/SC-014) ✔ |
| **hover en oscuro** (`index.css`, 1 token) | la congelada de contraste de F3 (que corre sobre el sitio rediseñado) ✔ y la unit del par hover en ambos temas ✔ |

**Con el resto de criterios:** diff sin tocar `backend/`, panel, `App.tsx`, e2e congeladas ni `.env.example` (0 archivos); **IMP-1 se mantiene cerrado** (aserción dura ≤159×44 en e2e `diseno-movimiento` «IMP-1», ✔, y mis 2 pruebas de `qa-medidas-f10` ✔). Las demás correcciones de ciclos previos (I1, M1-seguridad, M2-revisor, C2) siguen en verde.

## Estado FR/SC (solo deltas del ciclo 3)

La matriz completa está en `revision-2026-10-11-qa.md` y el delta del ciclo 2 en `revision-2026-10-11-qa-ciclo2.md`. En este ciclo solo actualizan:

| Criterio | Ciclo 2 | Ciclo 3 | Evidencia nueva |
|---|---|---|---|
| **FR-022 / FR-024** (contraste AA en ambos estados/temas de los controles nuevos) | ◐ | **✔ (la pieza nueva)** | opción activa del selector 6,66:1 en ambos temas (medido en render real + unit); el hover del botón primario seguía ✔ |
| **FR-008** (no re-animar lo ya visto; sin repeticiones molestas) | ✔ | ✔ reforzado | latch de `Reveal.tsx`: misma regla que la del ciclo 2, ahora también reactivada correctamente tras el TDZ fix; e2e y unit nuevas pasan |
| **FR-005 / SC-014** (revelado al scroll como capa visual estable) | ✔ | ✔ | TDZ del timer corregido; revelado con métricas de perceptibilidad intacto (25/25) |
| Resto | como en ciclo 2 | sin cambios | misma cobertura: 8/8 congeladas + 25/25 nuevas |

Permanecen pendientes de **verificación humana** (no bloquean lo técnico, sin cambios desde el ciclo 1): SC-002 «moderno», SC-003/SC-007 con personas, SC-001/SC-011 revisión visual, SC-009 auditoría fina de contraste en oscuro, SC-013 «0 copias», SC-014 sensación final, y FR-012 (previsión F4–F9) cuando existan secciones.

## Hallazgos del ciclo 3

- **BLOQUEANTES:** ninguno.
- **IMPORTANTES:** ninguno.
- **MENORES (heredados, sin cambios respecto al ciclo 2):**
  1. Previsión F4–F9 (FR-012) no probable hasta que existan secciones reales (patrón correcto por diseño y fuente única nav/mapa/pie).
  2. Rate-limit M-4 de login de F2 (20/min por IP): e2e en dos tandas; sin acción de F10.

## Veredicto

**APROBADO** — 0 bloqueantes, 0 importantes.

El selector de idioma mantiene la huella compacta que cerró IMP-1 (152 × 44 px < 159 × 44 de F3) y su opción activa usa ahora un par sólido con **6,66:1** de contraste en ambos temas, por encima del AA requerido. Los NIT del ciclo 2 (TDZ del timer, latch de `reduced-motion`, hover en oscuro) están corregidos y cubiertos por pruebas nuevas sin debilitar ninguna existente. Toda la suite en verde: unit 409, e2e 25 + 8 congeladas, lint/typecheck/build. F10 queda lista para la puerta de PR; los criterios humanos (SC-002, SC-003, SC-009 en oscuro, SC-013, SC-014) quedan anotados para la revisión de aceptación.
