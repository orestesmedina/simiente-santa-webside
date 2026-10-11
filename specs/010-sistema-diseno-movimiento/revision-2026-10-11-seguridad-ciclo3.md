# Revisión de seguridad — F10 · ciclo 3 (re-validación final)

**Fecha**: 2026-10-11 · **Rama**: `010-sistema-diseno-movimiento` · **Auditor**: `seguridad`
(solo lectura; este reporte es la única escritura).
**Objeto**: `git diff abb28a2..HEAD` — los **2 commits de código del ciclo 3**
(`880f840` «opción de idioma activa con par de contraste sólido AA» y `386be2e`
«cierra los NIT del ciclo 2: TDZ del timer, latch de reduced-motion, hover en
oscuro») más docs (`cfaaa71`, que versiona los reportes del ciclo 2). Encargo:
confirmar que **no hay riesgos nuevos**.

---

## 1. Qué cambió (verificado con `git diff abb28a2..HEAD --stat`)

**Solo 4 archivos de producción**, todos presentacionales:

| Archivo | Cambio | Evaluación |
|---|---|---|
| `i18n/LanguageSwitcher.tsx` (1 línea) | opción activa: `'bg-accent/15 … text-accent'` → `'bg-accent … text-navy'` | clases **estáticas** de Tailwind; sin datos externos, sin superficie |
| `index.css` (1 línea) | token `--accent-hover` del tema oscuro: `#00c9a7` → `#00b598` | valor **hex estático**; nada dinámico |
| `motion/Reveal.tsx` | latch «una vez mostrado, nunca se re-oculta»: ajuste de estado durante el render (patrón React de estado derivado) | solo renombra estado y añade `setRevealed(true)`; `data-reveal` sigue siendo un union **estático** (`'in' | 'up' | 'left' | 'right'`); sin DOM dinámico |
| `motion/useReveal.ts` (6 líneas) | TDZ: `let timer` declarado **antes** de los callbacks que lo consultan | reordenación de una declaración local; sin cambios de comportamiento observable de seguridad |

El resto: pruebas (`PublicButton.test`, `i18n.test`, `motion.test`,
`IdentityHero.test`), el **nuevo helper de pruebas `src/test/contrast.ts`**
(44 líneas), `components/publico/README.md`, `specs/` y `estado.md`.

**`frontend/index.html`, `package.json`, `package-lock.json`, `e2e/` y todo el
backend: sin cambios** en el ciclo (stat del diff) → la superficie auditada en
los ciclos 1-2 permanece intacta; el hash del script anti-FOUC documentado para
el futuro CSP (M2) sigue siendo válido.

## 2. Nuevo archivo `src/test/contrast.ts` — sin riesgo

Helper **solo de pruebas** (no entra en el build de producción): lee
`src/index.css` con `readFileSync` por **rutas fijas** relativas al cwd
(`src/index.css` / `frontend/src/index.css`), extrae tokens con una regex sobre
nombres pasados como literales por las pruebas y calcula el contraste WCAG 2.1
sobre hexes. Sin entrada externa, sin red, sin `eval`.

## 3. Barridos re-ejecutados sobre el diff del ciclo

- Grep anclado sobre las **309 líneas** del diff de `frontend/` del ciclo 3 por
  `href|src=|url(|http|<script|password|secret|api_key|Bearer|innerHTML|dangerously|window.open|eval(`
  → **0 coincidencias** (salida 1 de grep, sin tuberías).
- `npm audit --audit-level=high` re-ejecutado → **0 vulnerabilidades**
  (`package-lock.json` sin cambios; 0 dependencias nuevas).
- 0 cambios de backend/migraciones/API/infra/e2e en el ciclo.

## 4. Hallazgos

### CRÍTICO — ninguno · IMPORTANTE — ninguno · MENOR — ninguno · NIT — ninguno (nuevos)

### Heredados (estado, sin cambios en este ciclo)

- **M1** (sesión del e2e en tmp) — **cerrado** en el ciclo 2 (dir único `0700`
  por proceso, archivo `0600`, borrado al salir).
- **M2** (script inline anti-FOUC + futuro CSP) — abierta **por diseño**, no
  bloqueante; `index.html` intacto en este ciclo.
- **N1** (semillas e2e en `helpers.ts`, preexistente de F3) y **N2** (maquetas
  estáticas) — sin cambios.
- **N-c2-1** (limpieza del storageState solo en `'exit'`) — opcional, sin
  cambios; no bloqueante.

---

## Veredicto

**APROBADO** — 0 CRÍTICO · 0 IMPORTANTE · 0 MENOR · 0 NIT nuevos.

El ciclo 3 toca solo clases/tokens CSS estáticos, un latch de estado React y una
reordenación de declaración, más pruebas y un helper de contraste de solo-lectura
local: no introduce secretos, permisos, URLs dinámicas ni vías de inyección.
La superficie auditada en los ciclos 1-2 está intacta y `npm audit` sigue en 0.
