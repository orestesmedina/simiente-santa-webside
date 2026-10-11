# Análisis de coherencia (analyze) — F10 Sistema de diseño y movimiento

**Fase**: 3/9 (Coherencia) · **Revisor**: `revisor-codigo` · **Fecha**: 2026-10-10
**Re-análisis (ronda 2)** tras las correcciones del `arquitecto` y el `disenador-ux` sobre
`tasks.md`, `plan.md`, `research.md`, `quickstart.md` y `ux.md`.
**Artefactos revisados**: `spec.md` (FR-001…FR-027, SC-001…SC-015, US1–US6), `plan.md`,
`research.md` (R10-1…R10-16), `data-model.md`, `contracts/README.md`, `quickstart.md`,
`ux.md` (diseño final B + correcciones post-analyze), `tasks.md` (T1001…T1036, D1–D21),
`.specify/memory/constitution.md`, `AGENTS.md` y el estado real de `frontend/` (suites
congeladas F2/F3 re-verificadas: `homePage.test.tsx`, `layout.test.tsx`, `i18n.test.tsx`,
`portada-publica.spec.ts`, `portada-accesibilidad.spec.ts`, `LanguageSwitcher.tsx`,
`WhatsAppSection.tsx`, `sections.ts`, `es.ts`/`en.ts`).

---

## Veredicto: **APROBADO** ✅

Los **4 CRÍTICOS y los 3 IMPORTANTES del primer análisis quedan cerrados** de forma
correcta y verificable contra las pruebas congeladas. La cobertura de requisitos sigue
completa (todos los FR/SC con tareas; 0 huérfanas), la arquitectura se mantiene limpia
(0 dependencias nuevas, 0 migraciones, 0 cambios de API, frontera F10B intacta) y no
aparecen contradicciones nuevas. Quedan **1 MENOR abierto por acuerdo** (M4, calibración
de QA) y **2 NIT** cosméticos; **ninguno bloquea** la Fase 4 (implementar).

**Resumen de hallazgos restantes**: **0 CRÍTICO · 0 IMPORTANTE · 1 MENOR (no bloqueante) · 2 NIT**.

---

## Cierre de hallazgos de la ronda 1

| Hallazgo | Dónde quedó corregido | Verificación contra las suites congeladas |
|---|---|---|
| **C1** Mapa del pie vs `queryByRole('heading')` | `tasks.md` **D18** + T1015 («Sin grupos vacíos», «Rótulos no-heading»); `ux.md` §2.4 (reglas de renderizado 1–3, etiqueta no-heading); `plan.md` estrategia de navegación (6) y tabla de pruebas; `research.md` R10-15; `quickstart.md` §5.3 | ✅ Con datos vacíos el mapa **no se dibuja** → `queryByRole('heading', { level: 2 })).toBeNull()` pasa; con datos parciales el grupo «Contacto» no se renderiza y su rótulo **no** es `heading` (ni `h2` ni `h3`) → `queryByRole('heading', { name: 'Contacto' })).toBeNull()` resuelve solo el `h2` de la sección, como exige la prueba |
| **C2** Nombres accesibles duplicados (pie ↔ secciones ↔ CTA hero) | `tasks.md` **D19** + T1015 («Nombres accesibles únicos») + T1021 (CTA del hero con nombre propio); `ux.md` §2.1/§2.4; `plan.md` (5), RG10-18; `research.md` R10-15; `quickstart.md` §5.3 | ✅ Verificado contra `homePage.test.tsx`: «Contacto: Correo/Teléfono/WhatsApp» o «Correo: {email}…» ≠ `getByRole('link', { name: 'hola@ejemplo.com' })` (match exacto) ni `'+506 8888 8888'`; Síguenos con `SOCIAL_NETWORK_LABELS` plano («Facebook») ≠ regex `/Ver en Facebook/`; el CTA del hero (p. ej. «Escríbenos por WhatsApp») no contiene «Escribir por WhatsApp»/«Entrar al grupo», que pertenecen a `WhatsAppSection` (`whatsapp.action.dm/group`, verificado en `es.ts` e `IdentityHero.tsx` — hoy el hero **no** tiene enlace de WhatsApp, así que el nombre nuevo no desplaza ninguna aserción); dirección del pie como texto |
| **C3** Marca «Simiente» a <480 px | `tasks.md` **D20** + T1011/T1016 (enlace de marca con nombre completo, `sr-only`, **nunca** `display:none`; marca del pie = texto); `ux.md` §2.2/§3; `plan.md` FR-025 y RG10-18; `research.md` (4) | ✅ El `tabCycle` congelado exige un `<a>` con «Simiente» en el `innerText` a 320 px: el enlace de cabecera conserva el nombre completo (la técnica `sr-only` mantiene el texto en el `innerText` — `display:none` lo expulsaría); con el nombre de identidad sembrado («Iglesia Simiente …») «Simiente» cae dentro del `slice(0, 40)` |
| **C4** Lista duplicada de enlaces dentro del `<nav>` nombrado | `tasks.md` **D15** + T1012/T1013 (`MobileMenu` **controla la misma lista**; aserción «no se renderiza una segunda lista»); `ux.md` §2.3 («Una sola lista — inviolable»), §4.2; `plan.md` P10-13 y (3); `research.md` (3) | ✅ `within(nav).getByRole('link', …)` (`homePage.test.tsx`) y `sectionsNav.getByRole(…)` (`portada-publica.spec.ts`) siguen resolviendo **una** coincidencia por sección, con menú abierto y cerrado |
| **I1** Trampa de foco vs barra operable / `role="dialog"` | `tasks.md` **D17** + T1013/T1030 (disclosure APG: sin trampa, sin `role="dialog"`/`aria-modal`; aserciones explícitas en `MobileMenu.test.tsx`); `plan.md` P10-13 y gestión de foco; `research.md` R10-14; `quickstart.md` §5.2; `ux.md` §2.3/§6 | ✅ Un único patrón en los 5 documentos (todas las menciones de `dialog`/`aria-modal`/trampa son negativas); la barra (idioma/tema) sigue operable, coherente con FR-010/FR-023/FR-025 y con la spec (que no exige trampa) |
| **I2** Colapso de anclas fuera del tab order | `tasks.md` **D16** + T1012/T1013 (`display:none` bajo la media query: enfocables solo al abrir); `plan.md` (2), RG10-18; `research.md` (2); `ux.md` §2.3; `quickstart.md` §5.2 | ✅ Sin foco invisible (FR-013/edge case de la spec) y sin atributos React `hidden`/`aria-hidden`/`inert`: jsdom sigue viendo los nodos (las clases Tailwind no aplican estilos en jsdom) → unitarias congeladas intactas |
| **I3** Recuento del `tabCycle` < 30 | `tasks.md` T1015 (**tabla de inventario** por zonas: ~20–26 paradas → ~21–27 pasos) + recuento adelantado a T1013/T1030; `plan.md` RG10-21 (inventario real, +6…+8 netos, escalera de aligeramiento si QA ≥ 28); `research.md` (4) | ✅ Inventario contrastado con el seed de `portada-accesibilidad.spec.ts` (identidad, about, contacto, 1 servicio, 1 canal; redes solo residuales) y con los componentes actuales: cifras plausibles y **< 30** con el margen declarado; el aligeramiento y la prohibición de editar la prueba (RG10-6) están fijados |

Además quedaron cerrados los **MENORES/NIT de la ronda 1**: **M1** (claves i18n — `ux.md` §7
mapeado al contrato congelado `nav.label`/`a11y.skip`/`lang.label`, nomenclatura única
`nav.menu.open/close` + `footer.sitemap.*` incluido `footer.sitemap.home` en plan y tasks),
**M2** (`ScheduleSection` con representación dual congelada — tabla `sm:` con nombre accesible
«Horario de servicios» + lista `sm:hidden` — en `ux.md` §4.1 y T1022), **M3/N2** (números de
movimiento unificados en **D21**: tokens de T1004/plan como única fuente, techo **≤ 800 ms**
sin tolerancia de 900 ms, `ux.md` §5 como dirección visual; verificado en T1004/T1032),
**N1** (advertencias obsoletas de `ux.md` marcadas como resueltas), **N3** (criterio de
ubicación de componentes documentado en T1035) y **N4** (excepción declarada a §III para
T1001–T1004/T1035/T1036).

---

## Hallazgos restantes

### 🟡 MENOR (no bloqueante; queda abierto por acuerdo)

#### M4 · 0.82–0.85 rem (~13 px) en el selector vs FR-004, sin umbral de legibilidad definido

- **Dónde**: `tasks.md` T1027 («0.82–0.85 rem… El cuerpo de texto no baja del umbral de
  legibilidad (FR-004)»), `ux.md` §0.6. El «umbral» sigue sin valor numérico en ningún artefacto.
- **Estado**: abierto **como calibración de QA**, tal como acordaron orquestador y arquitecto.
  No contradice ninguna prueba congelada ni otro artefacto; SC-007 («menos espacio») y FR-004
  («legibilidad para todas las edades») se cierran en la revisión humana.
- **Corrección sugerida (en QA, no bloquea implementar)**: fijar el umbral (p. ej. ≥ 0.875 rem /
  14 px para el texto identificable, con el nombre completo siempre accesible) o registrar la
  decisión en el acta de T1036 si se mantiene 0.82 rem.

### ⚪ NITS (cosméticos; pueden corregirse de paso)

- **N5 · Microcopy del prefijo de Contacto en dos formas**: D19/T1015 escriben «Contacto:
  Correo» / «Contacto: Teléfono» / «Contacto: WhatsApp», mientras la nota de i18n de
  `tasks.md` y `plan.md` (5) escriben «Correo: {email}» / «Teléfono: {phone}» / «WhatsApp:
  {canal}». Ambas evitan las colisiones (ninguna coincide con el dato desnudo), pero
  conviene dejar **una** forma en T1006/T1015 para que el test de nombres únicos se escriba
  contra un microcopy fijo.
- **N6 · Margen del ciclo de tabulación**: la tabla de T1015 dice «~21–27 pasos → < 30 ✓
  (margen ≥ 3)»; en el extremo superior el margen es 2 y con datos residuales RG10-21 admite
  hasta ~28. La escalera de aligeramiento y el umbral de re-cuento (≥ 28) ya están fijados;
  solo conviene ajustar la frase para no prometer «margen ≥ 3».

---

## Cobertura FR/SC → tareas (sin cambios respecto a la ronda 1)

**Sin FR/SC sin tarea y sin tareas huérfanas** (T1001…T1036 referencian requisitos
existentes; las correcciones no alteraron la cobertura).

| Requisito | Tareas | Verificación | Estado |
|---|---|---|---|
| FR-001/FR-002/FR-003/FR-004 (sistema visual, logo, componentes, legibilidad) | T1001, T1003, T1008–T1010, T1021–T1025, T1027, T1035 | unitarias de primitivas, `BrandLogo.test.tsx` intacta, revisión SC-001/SC-002, quickstart §2 | ✅ |
| FR-005…FR-009 (movimiento marcado, reduced-motion, fluidez, feedback, independencia) | T1004, T1013, T1017–T1020, T1024–T1026 | `motion.test.tsx`, e2e T1032/T1033 (techo unificado ≤ 800 ms, D21) | ✅ |
| FR-010…FR-014 (navegación por dispositivo, sin scroll incómodo, escala F4–F9, accesible, siempre alcanzable) | T1011–T1014, T1020 | `MobileMenu.test.tsx`, e2e T1029/T1030, congeladas sin cambios | ✅ (D15–D17) |
| FR-015/FR-016/FR-017 (selector compacto, comportamiento F3, sin desplazamientos) | T1027 | `i18n.test.tsx` (existentes sin cambios + nuevos), e2e T1029 | ✅ (ver M4) |
| FR-018/FR-019 (portada + cimientos F4–F9, textos es/en) | T1006, T1021–T1026, T1035 | `i18n.test.tsx` (SC-010), revisión SC-001/SC-011 | ✅ (M1 cerrado) |
| FR-020/FR-021 (no-regresión, sin datos nuevos) | T1034 + frontera F10B | suites F2/F3 sin cambios | ✅ (C1–C4 cerrados) |
| FR-022 (accesibilidad en ambos temas) | T1013, T1015, T1028–T1030 | e2e teclado + checklist SC-009 | ✅ (I1/I2 cerrados) |
| FR-023/FR-024/FR-025 (tema con interruptor, contraste, cabecera 320 px) | T1001, T1002, T1005, T1010, T1011, T1021, T1028, T1031, T1036 | `theme.test.tsx`, e2e T1031, checklist SC-012 | ✅ (C3 cerrado) |
| FR-026 (3 propuestas, B aprobada antes de código) | proceso cumplido + T1021–T1026, T1036 | acta SC-013 | ✅ |
| FR-027 (mapa del sitio sin rotos) | T1007, T1015, T1016 | `FooterSiteMap.test.tsx`, e2e T1030 | ✅ (C1/C2 cerrados) |
| SC-001…SC-015 | ver tabla de `tasks.md` (líneas ~560–600) | e2e, unitarias, checklist y **aprobación humana** (SC-002/SC-013/SC-014 no auto-aprobables) | ✅ (SC-008 sin riesgos abiertos) |

**Cobertura de US**: US1 → Fase 2+5 · US2 → Fase 3 · US3 → Fase 4 · US4 → T1027 ·
US5 → T1005/T1028/T1031 · US6 → T1035. ✅

## Coherencia transversal (spec ↔ plan ↔ ux ↔ tasks)

| Tema | Estado |
|---|---|
| Navegación por dispositivo (barra ≥861 / hamburguesa ≤860, breakpoint 860) | ✅ único patrón *disclosure* APG en los 5 documentos (D4/D15–D17) |
| Selector compacto (textos completos, sin píldoras/menú, solo cabecera+pie) | ✅ (M4 abierto solo como calibración) |
| Modo oscuro con interruptor + persistencia + anti-FOUC | ✅ (D7; FR-023/FR-024) |
| Mapa del sitio (Explorar/Contacto/Síguenos, sin grupos vacíos, nombres únicos) | ✅ (D3/D18/D19) |
| Movimiento marcado + `prefers-reduced-motion` incondicional, números únicos | ✅ (D5/D6/D21) |
| Marca «Simiente» y contrato congelado en todos los anchos | ✅ (D14/D20) |
| No-panel / F10B · 0 cambios backend/datos/API/migraciones | ✅ (Out of Scope; `data-model.md`; `contracts/README.md`) |

## Constitución

| Principio | Estado |
|---|---|
| §III test-first | ✅ (excepción declarada para T1001–T1004/T1035/T1036: N4) |
| §V TypeScript strict, sin `any` | ✅ |
| §IV sin `dangerouslySetInnerHTML` | ✅ |
| §II 0 dependencias nuevas (justificadas en R10-10/R10-4) | ✅ |
| §VI 0 migraciones | ✅ |
| §VIII aprobación humana (SC-002/SC-013/SC-014) | ✅ |

## Riesgos residuales conocidos (no bloqueantes)

1. **M4** (tamaño del selector): se cierra en QA con SC-007/SC-002/FR-004.
2. El `tabCycle` depende de datos residuales de la base compartida: el recuento de T1013/T1015
   debe hacerse con el estado real que dejan las suites; la escalera de aligeramiento de
   RG10-21 cubre el caso.
3. La calibración de espaciado/movimiento (RG10-22) y el rendimiento de blobs+blur (`ux.md`
   §9.4) se cierran en QA, tal como prevén los artefactos.

---

# 📜 Histórico — Análisis ronda 1 (2026-10-10, superado)

> Veredicto de la ronda 1: **RECHAZADO** (4 CRÍTICO · 3 IMPORTANTE · 4 MENOR · 4 NIT).
> **Todos los CRÍTICOS e IMPORTANTES quedaron corregidos** por `arquitecto`/`disenador-ux`
> en esta ronda (ver tabla de cierre arriba). Se conserva el resumen por trazabilidad:

- **C1** (cerrado): encabezados de grupo del mapa del pie (h2 «Explorar»/«Contacto») vs
  `homePage.test.tsx` (`queryByRole('heading', …)` con datos vacíos y parciales).
- **C2** (cerrado): nombres accesibles duplicados del pie (`hola@ejemplo.com`, teléfono,
  `/Ver en Facebook/`, `/Escribir por WhatsApp/`) vs `getByRole` de coincidencia única.
- **C3** (cerrado): monograma de marca <480 px vs la aserción «Simiente» del `tabCycle`.
- **C4** (cerrado): lista duplicada del menú móvil dentro del `<nav>` «Secciones de la portada».
- **I1** (cerrado): trampa de foco vs barra operable / `role="dialog"` sin resolver.
- **I2** (cerrado): técnica de colapso de anclas sin fijar (tab order / árbol de accesibilidad).
- **I3** (cerrado): estimación incompleta del saldo de paradas del `tabCycle` (RG10-21).
- **M1–M3** (cerrados) · **M4** (abierto, ver arriba) · **N1–N4** (cerrados) · **N5–N6** (nuevos, NIT).
