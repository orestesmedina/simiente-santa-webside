# Revisión de código — F10 · Ciclo 3 (re-validación final)

- **Fecha**: 2026-10-11 · **Revisor**: `revisor-codigo` (Fase 5, ciclo 3)
- **Alcance**: commits `880f840` (N1) y `386be2e` (NIT del ciclo 2), contra `revision-2026-10-11-codigo-ciclo2.md`.
- **Base verificada**: suite unitaria re-ejecutada por mí: **409/409 en verde** (65 archivos; +2 pruebas: contraste de la opción activa y latch OFF→ON→OFF). **SC-008 intacto**: `git diff main...HEAD` sobre todos los `*.test.tsx`/`*.spec.ts` sigue con **0 borrados** (ninguna aserción congelada de F2/F3 tocada en ningún ciclo). Sin cambios de backend/datos/API ni de dependencias.

## Cierre del bloqueante y de los NIT del ciclo 2

| Hallazgo | Estado | Evidencia |
|---|---|---|
| **N1** (CRÍTICO) contraste AA de la opción activa del selector | ✅ **CERRADO** | `LanguageSwitcher.tsx:45` pasa a par **sólido** `bg-accent font-bold text-navy` (sin alfa): verificado por mí con la fórmula WCAG — **6,66:1** en claro y en oscuro (`--accent #00c9a7` en ambos temas vs `--color-navy #1a2b4a`), y además resoluble por el auditor de contraste del e2e congelado (el problema de los fondos con alfa queda eliminado de raíz). `i18n.test.tsx` fija el par (clases + razón ≥ 4,5:1 calculada desde los tokens, con aserción de «un valor de acento por tema»). Opción inactiva sin cambios: `text-muted` ≈ 5,83:1 ✓. |
| **Huella IMP-1** | ✅ **CONSERVADA** | El cambio es solo de color del estado activo; `min-h-11`, `px-0.5`, `gap-0.5`, `text-sm` intactos. Las medidas siguen cubiertas: e2e `diseno-movimiento.spec.ts` (≤ 159 × 44 px frente a F3) y `qa-medidas-f10.spec.ts` (táctil ≥ 44 px, texto ≥ 14 px, pastilla ≤ 56 px). |
| **N1a** TDZ del `timer` | ✅ | `useReveal.ts`: `let timer` declarado antes de los callbacks; `seen`/`reveal` ya no pueden caer en la zona muerta temporal. |
| **N2** hover oscuro indistinguible | ✅ | `--accent-hover` oscuro pasa a `#00b598` (par `#00b598`/`#1a2b4a` = **5,42:1** ✓ ≥ 4,5) con desplazamiento perceptible frente al reposo `#00c9a7`, más el lift; README actualizado. El test de contraste de `PublicButton.test.tsx` cubre el nuevo valor. |
| **N3** `process.cwd()` frágil | ✅ | Helper compartido `frontend/src/test/contrast.ts` (`readSystemCss` resuelve `frontend/src/index.css` desde `frontend/` o desde la raíz) reutilizado por `PublicButton.test.tsx` e `i18n.test.tsx`. |
| **N4** residual de M2 (OFF→ON→OFF) | ✅ | `Reveal.tsx` implementa latch completo («una vez mostrado, nunca se re-oculta») con estado derivado durante el render —patrón React correcto, condicionado, sin bucle— y prueba unitaria de la secuencia OFF→ON→OFF. El revelado normal al scroll no se ve afectado (el latch solo se cierra al mostrarse). |
| **N5** aserción vaga | ✅ | Eliminada de `IdentityHero.test.tsx`; la aserción efectiva («Ver horario» ausente) permanece. |

## Hallazgos

**Ninguno.** 0 CRÍTICO · 0 IMPORTANTE · 0 MENOR · 0 NIT.

## Notas de verificación (no son hallazgos)

- Verificado por mí en este ciclo: diff completo de los commits, contraste WCAG por fórmula de los pares nuevos/modificados, SC-008 a nivel de diff y suite unitaria completa (409/409). Los e2e (Playwright, incluida la guarda de IMP-1 y las de C1/I1) no se re-ejecutan en esta revisión por entorno; los corren QA/CI en la validación paralela — el código no cambia geometría ni contratos desde el ciclo 2, solo color y pruebas.
- Puntos de vigilancia aceptados (ya documentados en `README.md`/`tasks.md`): calibración final de la propuesta B con el cliente (SC-002/SC-013, humana) y degradado de los blobs en el dispositivo modesto (RG10-9).

## Veredicto

**APROBADO** — N1 cerrado con par sólido AA verificado (6,66:1 en ambos temas) y regresión de contraste de la huella imposible por diseño (par sólido + test de tokens); huella compacta de IMP-1 conservada; los 5 NIT del ciclo 2 cerrados con pruebas; 409/409 unitarias en verde; SC-008 intacto tras los tres ciclos. La rama está lista para la fase de entrega (PR), sujeta a las aprobaciones humanas de siempre.
