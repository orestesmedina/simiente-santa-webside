# Re-validación QA (ciclo 2) — F3 Portada e información general

**Fecha**: 2026-10-10 · **Rama**: `003-portada-info-general` · **Validador**: `qa-tester`
(solo escribe pruebas y reportes; no toca código de producción).

**Objeto**: re-validación del bucle de corrección 2 contra el informe anterior
(`revision-2026-10-10-qa.md`, veredicto anterior: **RECHAZADO** por 1 BLOQUEANTE B1).
Correcciones evaluadas: B1 `fix(portada): distinguir null de ausente en los PATCH`
(`d5d4610`), A1 `fix(frontend): objetivo táctil del selector de idioma` (`2b72138`),
I1 `fix(frontend): horario en a.m./p.m. localizado` (`fc96185`).

## Ejecuciones de esta sesión

| Comando | Resultado |
|---|---|
| `cd backend && go test -count=1 ./internal/portada/...` | **verde** — las 3 pruebas de regresión de B1 (`service_patch_null_test.go`) pasan con la corrección (`0.022 s`, sin caché) ✓ |
| `cd backend && go test -count=1 -tags=integration ./...` | **verde en 19 paquetes** ✓ (repositorio + auditoría atómica + migración 000006; Docker arriba) |
| `cd frontend && npm test -- --run` | **45 archivos, 271 pruebas, verde** ✓ (2 nuevas vs ciclo 1: `schedule.test.ts` y `homePage.test.tsx` del I1) |
| `npx playwright test --config e2e/playwright.config.ts --workers=1` | **7/7 en verde (26,6 s)**: acceso, auditoría, portada-accesibilidad, portada-panel, portada-patch-null (nueva, abajo), portada-pública, status ✓ |
| `npx eslint e2e/portada-patch-null.spec.ts` + `tsc --noEmit` + `prettier` | ✓ |

## Estado de los defectos del ciclo 1

### B1 (BLOQUEANTE → **CERRADO**)

**`null` distingue limpiar de no tocar, verificado en las tres capas:**

1. **Unitarias de regresión**: las 3 pruebas que decodifican el cuerpo JSON exacto del
   panel (`{"endTime": null}`, `{"nameEn": null}`) pasan a verde con la corrección
   `Optional[T]`: `null` limpia, ausente no toca y PATCH solo-`null` ya no responde
   `400 «No hay cambios que guardar»`.
2. **E2e en el STACK REAL** (prueba nueva `portada-patch-null.spec.ts`, 2,8 s):
   servicio publicado con hora de fin y traducción →
   - `PATCH {"endTime": null}` → **200** (antes 400) y la respuesta omite `endTime`;
   - campos ausentes intactos: `nameEn` y `publicationState: 'published'` persisten;
   - segundo `PATCH {"nameEn": null}` → **200**, traducción vaciada y sigue publicado;
   - la limpieza se ve en el refetch del panel: `GET /admin/portada` ya no devuelve `endTime`;
   - la portada pública (`GET /api/v1/portada?lang=es`) refleja la limpieza al recargar
     (FR-014/SC-003 intacto);
   - **auditoría viva**: `GET /admin/auditoria/acciones` contiene
     `home.schedule.update` con `targetLabel «Portada · Horario · QA-B1c2 …»` por cada
     limpieza (FR-017), confirmada también contra la tabla `admin_actions` de PostgreSQL.
   - contraprueba en WhatsApp: `PATCH {"nameEn": null}` en un canal → 200 y limpiado.
3. **No se rompió nada**: backend unit + integración, 271 unitarias de frontend y 7/7 e2e en verde.

Sobre la corrección: es local a la capa de administración del servicio de portada (tipo
`Optional[T]` en los PATCH, `d5d4610`); la matriz del ciclo 1 sigue cubierta: los mismos
5 e2e + el de accesibilidad pasan sin cambios, más el nuevo de B1.

### A1 (MENOR → **CERRADO**)

- `2b72138` eleva el objetivo de toque del selector de idioma de 20 px a **min-h-11 (44 px)**
  en cabecera y pie, sin cambiar el diseño.
- La aserción del e2e de accesibilidad (`toque ≥ 44 px` sobre enlaces y botones a 320 px)
  pasa; suite completa en verde. Registro: además del mínimo WCAG 2.5.8 (24 px) queda por
  encima de la regla de la skill (44 px), relevante para la tarea 3 del protocolo SC-010.

### I1 (IMPORTANTE del revisor → **CERRADO**)

- `fc96185`: la presentación del horario sigue `ux.md` §4.6/D-3 (es: «10:00 a. m. − 12:00 m.»,
  en: «10:00 AM − 12:00 PM»); el dato del contrato sigue siendo HH:MM (sin tocar el backend).
- Cubierto con pruebas nuevas: `schedule.test.ts` (mediodía `m.`, medianoche, rangos) y
  `homePage.test.tsx:151-152` sobre la portada montada; los 7 e2e pasan.

### M-a11y (MENOR → sigue **abierto**, deuda aceptada)

- No se tocaron las tarjetas de contacto (`break-words`/`min-w-0` solo en `MiCuentaCard`
  y la `nav`): a 200 % de zoom sobre un teléfono de 320 px (viewport efectivo 160 px) el
  desborde fino persiste. No incumple ningún criterio de la spec (WCAG 1.4.4 y 1.4.10
  siguen cubiertos por el e2e de accesibilidad; el texto al 200 % conserva todo el contenido).
- Estado: deuda de experiencia sin criterio roto; que el humano decida si entra en cierre o
  backlog.

## Resultado de las suites (veredicto por componente)

| Suite | Resultado | Sesión anterior |
|---|---|---|
| Backend unit (`go test ./internal/portada/...`) | ✅ (incluye las 3 de B1) | ❌ por B1 (esperado) |
| Backend integración (`-tags=integration`) | ✅ 19 paquetes | ✅ |
| Frontend unitarias (`npm test -- --run`) | ✅ 45 archivos / 271 pruebas | ✅ 269 (ccl 1: +2 del I1) |
| e2e Playwright (incl. `portada-accesibilidad.spec.ts`) | ✅ **7/7** | ✅ 5/5 + 1 aserción roja de A1 |
| Tipos / lint / formato de la prueba nueva | ✅ `tsc`, `eslint`, `prettier`, hooks git | — |

## Prueba añadida (versión de la traza viva del ciclo 2)

- `frontend/e2e/portada-patch-null.spec.ts` (nueva, commiteada `d98df37`). La traza temporal
  `qa-temp-ciclo2-b1.spec.ts` del intento anterior quedó **terminada y renombrada** a nombre
  permanente: no queda ninguna copia suelta `qa-temp-*` en el repo (eliminada).
  Justificación para versionarla: cierra en vivo el único bloqueante del ciclo 1 y queda como
  guardia de regresión del contrato (`minProperties: 1` con solo-`null`, distinción
  `null`/ausente, limpieza auditada y visible en portada). **Sin archivos temporales sueltos.**
- Corrección sobre el borrador temporal: su aserción de auditoría buscaba `entityId` en el
  payload paginado, y el contrato real (`AdminActionItem`, `model.go:352-363`) entrega
  `targetId` para cuentas/roles y `targetLabel` para contenido — la versión permanente
  empareja por `targetLabel` + nombre único de servicio. El fallo era de la propia prueba,
  no del producto.

## Hallazgos de esta sesión

- **BLOQUEANTE**: **0**.
- **IMPORTANTE**: **0** (I1 cerrado).
- **MENOR**: 1 — M-a11y reflow a 200 % de zoom (no es defecto de criterio; ver arriba).
- Nada que añadir: no se reprodujo ningún SC/FR nuevo roto. La corrección de B1 (`d5d4610`)
  es local a la capa de administración del servicio de portada (tipo `Optional[T]` en los
  PATCH), sin tocar la portada pública, la publicación, permisos ni auditoría (los 7 e2e en
  verde lo confirman).

## Pendientes no ejecutables por esta sesión (no son defectos de código)

- **SC-010**: prueba manual con ≥6 personas (2 por franja de edad) — coordina el **humano**
  con `qa-tester`, protocolo en `quickstart.md` §10.7, informe en
  `pruebas-usabilidad-SC-010.md`. Sigue siendo **requisito del cierre de F3** con el humano.
- **Lector de pantalla real** (SC-008): pasada manual con NVDA/VoiceOver recomendada en el
  cierre (sin evidencia de fallo por diseño).

## Veredicto

**APROBADO** — **0 BLOQUEANTES, 0 IMPORTANTE, 1 menor (M-a11y, deuda aceptada)**.

B1 cerrado en las tres capas (unitaria, integración y stack real con auditoría), A1 e I1
cerrados con sus pruebas, suite completo en verde (backend unit + integración, 271 unitarias
de frontend, 7/7 e2e incluidos accesibilidad y la nueva guardia de B1). Con esto el bucle de
corrección 2 queda saldado desde QA; quedan para el cierre con el humano las verificaciones
no automatizables (SC-010 y lector de pantalla real).
