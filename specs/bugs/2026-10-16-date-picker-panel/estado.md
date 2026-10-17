# Estado: 2026-10-16-date-picker-panel

<!-- Bug: los campos de fecha/hora del panel no muestran su selector nativo. -->

## Resumen

| Campo | Valor |
|---|---|
| Rama | fix/date-picker-panel |
| Flujo | equipo-bug |
| Fase | 5/5 · Validado (2.ª iteración); pendiente la aprobación humana del merge del PR #11 |
| Ciclo de corrección | 1/3 · cerrado (la 2.ª iteración se validó sin bloqueantes) |
| Próximo paso | Merge del **PR #11** (lo hace el humano). Si el CI vuelve a caer por el flaky preexistente, relanzar el job de backend |
| Bloqueado por | — |
| Actualizado | 2026-10-16 |

## Aprobaciones

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | n/a (bug) | | | |
| Plan | n/a (bug) | | | |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

- **2.ª iteración (2026-10-17, humano):** tras restaurar el `color-scheme` claro, los campos siguen **sin un icono visible** que indique dónde pulsar para abrir el selector. Se añadirá un **botón con icono** (calendario/reloj) junto a cada campo de fecha/hora del panel, que abre el selector nativo (`showPicker()`), además del arreglo de `color-scheme`.

## Verificación y validación (1.ª iteración)

- **Prueba de regresión:** `frontend/e2e/qa-bug-color-scheme-panel.spec.ts` — falla **antes** (`color-scheme` resuelto `dark` en el panel con tema oscuro) y pasa **después** (panel e inputs resuelven `light`, los campos siguen siendo nativos).
- **QA:** 12/12 e2e relacionadas en verde (incluye `tema-oscuro`, `qa-bug-fondo-oscuro-fuera-layouts`, `qa-medidas-f10`), `npm test` 782/782, lint/typecheck limpios. **APRUEBA**.
- **Revisión de código:** cambio mínimo en las tres superficies del panel; el sitio público no se toca; la prueba falla sin el arreglo y pasa con él. **APRUEBA** (condiciones: actualizar este estado —hecho— y registrar el fallo preexistente de `qa-bugs-004` Bug 1 —abajo—).
- **Seguridad:** sin impacto (solo CSS + spec). **APRUEBA**.
- **No atribuible a este arreglo:** el e2e `frontend/e2e/qa-bugs-004.spec.ts` (Bug 1, scroll a `#whatsapp` desde `/`) falla de forma determinista en este entorno; es el **e2e flaky preexistente de F3/F4** ya registrado en el roadmap (decisión 21). La página afectada no usa las clases del arreglo ni `PublicLayout` importa `app/layout.tsx`.

## Validación de la 2.ª iteración (iconos visibles)

- **QA:** `npm test` **793/793**, spec del bug **2/2**, e2e nuevas/relacionadas **8/8** (`qa-bug-color-scheme-panel`, `qa-date-picker-panel-indicador-unico`, `agenda-panel`, `agenda-accesibilidad`, `auditoria`), lint/typecheck limpios. **APRUEBA**.
- **Revisión de código:** botón accesible, un solo icono por campo, `ref` reenviado intacto, pruebas reales, sitio público intacto. **APRUEBA**, con una condición a confirmar visualmente: que el valor de la fecha no se recorte en la rejilla a 640–768 px (`pr-11`); dev y QA lo revisaron con capturas y sin desbordes. Sugerencias no bloqueantes: posible doble icono en Firefox, `aria-label` repetidos entre filas.
- **Seguridad:** sin impacto (UI + pruebas, sin dependencias ni datos). **APRUEBA**.

## Diagnóstico (causa raíz)

- Los campos **siguen siendo nativos** `<input type="date">` / `<input type="time">` (decisión de UX **D-4** de F4; el código no cambió desde F4). Verificado en el bundle desplegado: hay `type:"date"` (4) y `type:"time"` (6).
- **F10** introdujo el modo oscuro y fija `color-scheme` **globalmente en `<html>`** (`index.html` anti-FOUC y `ThemeProvider`).
- El **panel es siempre claro** (`bg-white` en sus layouts) pero **no resetea `color-scheme`**; los `input` son transparentes (preflight de Tailwind), así que se ven blancos.
- Con el tema público en **oscuro**, el navegador pinta el **indicador nativo del selector de fecha/hora en color claro** sobre el fondo blanco del panel → **invisible**. El texto se lee porque lleva color explícito. Resultado: parece un input de texto y hay que teclear la fecha.
- **Arreglo previsto:** forzar `color-scheme: light` en los contenedores del panel (`AppLayout`, `PanelLayout`, `PanelSurfaceLayout`), sin tocar el sitio público. Cubrirlo con una prueba que falle antes del arreglo.

## Decisiones y aclaraciones

- 2026-10-16 (humano): reportado como bug; "solo sale un input, hay que ingresar la fecha manualmente". Pide que vuelva a ser un *date picker*. Aprobado arreglarlo.

## Bitácora

- 2026-10-17 — **2.ª iteración** (pedida por el humano: «faltan los iconos para saber dónde dar clic»): botón con icono de calendario/reloj en `Field` para `date`/`time`, que abre `showPicker()`, con un solo icono por campo. Test-first (`47566e5`, `a5c64e9`, `fa5d029`, spec de QA `e23e49a`). Validada por QA/código/seguridad (793/793 unitarias, 8/8 e2e). Pendiente el merge del PR #11.
- 2026-10-16 — Validación: QA, revisión de código y seguridad **aprueban** (12/12 e2e relacionadas, 782/782 unitarias). Pendiente el PR.
- 2026-10-16 — Arreglo aplicado (`70c7012`): `[color-scheme:light]` en `AppLayout`, `PanelSurfaceLayout` y `PanelLayout`; la spec del bug pasa.
- 2026-10-16 — Prueba que falla (`e19973c`): `frontend/e2e/qa-bug-color-scheme-panel.spec.ts`.
- 2026-10-16 — Diagnóstico cerrado; rama `fix/date-picker-panel` creada.
