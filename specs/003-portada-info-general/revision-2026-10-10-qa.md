# Validación QA — F3 Portada e información general

**Fecha**: 2026-10-10 · **Rama**: `003-portada-info-general` · **Validador**: `qa-tester`
(solo escribe pruebas y reportes; no toca código de producción).

**Contraste**: criterios de `spec.md` (US1–US5, FR-001…FR-019, SC-001…SC-013) contra
**evidencia ejecutada en esta sesión**: stack local completo (`docker compose` + gb migración
`000006`), pruebas unitarias e integración del backend, suite Vitest del frontend, **6 e2e con
Playwright** (los 5 existentes + 1 nuevo) y trazas `curl` contra el stack real.

## Ejecuciones de esta sesión

| Comando | Resultado |
|---|---|
| `cd backend && go test ./...` | verde en 17 paquetes; `internal/portada` **rojo por las 3 pruebas nuevas de B1** (fallan a propósito hasta corregir el defecto, ver Hallazgos) |
| `cd backend && go test -tags=integration ./...` | verde ✓ (repositorio + auditoría atómica + migración 000006 up/down/up; no depende de B1) |
| `cd frontend && npm test -- --run` | **45 archivos, 269 pruebas, verde** ✓ |
| `golangci-lint run ./internal/portada/` | `0 issues` ✓ · `gofmt -l` vacío ✓ · `go vet` ✓ · `tsc` y `eslint` ✓ |
| `npx playwright test --config e2e/playwright.config.ts --workers=1` | **5/5 existentes en verde** (30,2 s): acceso, auditoría, portada-pública, portada-panel, estado ✓ |
| Nuevo `e2e/portada-accesibilidad.spec.ts` | **pasa salvo 1 aserción** que documenta el hallazgo A1 (objetivo puntable del selector de idioma, 20 px < 24 px WCAG 2.5.8) |

El entorno del navegador tuvo que repararse primero: el Chromium de Playwright carecía de
`libnspr4/libnss3/libasound2`; se descargaron los paquetes del repositorio de Ubuntu y se
extrajeron en `/tmp/opencode/pwlibs/usr/lib/x86_64-linux-gnu` (mismo `LD_LIBRARY_PATH` que
documenta `estado.md`). Con ello la suite completa de e2e corre de nuevo contra el stack real.

## Matriz de cobertura — criterios → prueba(s) → resultado

| Criterio | Prueba / evidencia | Resultado |
|---|---|---|
| **US1** esc. 1–5 (SC-001) | e2e `portada-publica.spec.ts` (identidad, quiénes somos, horario, WhatsApp, redes, contacto sin sesión; enlaces con `href` correcto) + traza viva `curl /api/v1/portada?lang=es` 200 con las 6 secciones | ✅ verificado |
| **US2** esc. 1 (guardar y conservar) | e2e `portada-panel.spec.ts` («Cambios guardados.», «Publicado en la portada.») + unitarias `TestSaveIdentity*`, `TestSaveAbout`, `TestSaveContact` | ✅ |
| **US2** esc. 2 y 7 (permiso, mensaje claro) | e2e panel: sin enlace del módulo, `/panel/informacion` → `/sin-permiso`, API forzada `403 forbidden`; traza viva `401 unauthenticated` sin sesión | ✅ |
| **US2** esc. 3 (validación sin guardado parcial) | trazas vivas FR-015 (abajo), e2e panel («Escribe el nombre.» junto al campo), unitarias `TestSaveContactHandlerInvalidEmail`, `TestCreateServiceEndBeforeStart`, `TestCreateWhatsappDirectAndDuplicate`, `TestCreateSocialLinkValidation` | ✅ |
| **US2** esc. 4 ( inglés conservado) | unitarias `TestSaveIdentityNormalizesAndAudits` (`nameEn` persiste); e2e panel guarda versión en inglés; integración `TestIntegrationUpsertHomeIdentitySingleton` | ✅ |
| **US2** esc. 5 (inglés opcional) | unitarias de normalización (analyze I6): `""`→NULL; traza viva: `PUT quienes-somos` con `textEn ""` → 200 | ✅ |
| **US2** esc. 6 (logo/imagen de portada) | e2e panel (subida 201 + alt obligatorio) + trazas vivas de rechazo (.svg renombrado → 400; 9 MB → 400; `logoFile` sin `logoAltEs` → 400) + `BrandLogo` usa `object-contain` y dimensiones fijas (D-9: no se deforma) | ✅ |
| **US3** esc. 1, 2, 6 y 7 (borradores ocultos, sección oculta, estado en el panel) | e2e pública (borrador con `toHaveCount(0)`, `#whatsapp` desaparece con solo borradores, badges «Borrador/Publicado»), handler tests `OmitsEmptySections`, payload público sin `publicationState` (traza viva: 0 apariciones) | ✅ |
| **US3** esc. 3–4 (publicar/retirar por elemento) | e2e panel (`Publicar`/`Retirar de la portada` por servicio, canal y los 3 singletons) + integración `TestIntegrationSingletonsPublishPerSection` | ✅ |
| **US3** esc. 5 / FR-014 / SC-003 (visible al guardar) | e2e panel (`PATCH` publicado → la portada pública muestra el cambio en la primera recarga; `Cache-Control: no-store` verificado vivo en `/api/v1/portada`) | ✅ |
| **US4** esc. 1–5 / FR-010 / SC-006 | e2e pública: primer visita → es; cambio → `lang=en` + interfaz traducida + contenido sin inglés en español (sin huecos); persiste al recargar y en `localStorage['ss.lang']`; re-visita respeta; contexto `en-US` sin preferencia → español | ✅ |
| **US5** esc. 1 / FR-018 / SC-007 | e2e pública y nuevo `portada-accesibilidad.spec.ts`: 320/360/768/1280 sin scroll horizontal, mismo h1 y funciones | ✅ |
| **US5** esc. 2 (persona con poca experiencia) | navegación visible siempre, sin menú oculto (sin hamburguesa), títulos `h1`/`h2` claros, selector con radios nativos y textos completos | ✅ (parte manual SC-010 pendiente, abajo) |
| **US5** esc. 3 (teclado / lector) | nuevo e2e de accesibilidad: el primer `Tab` foca el enlace de salto, `Enter` acciona, foco **visible** en cada paso del ciclo completo (sin trampa), radios operables con flechas | ✅ teclado · ⚠️ lector de pantalla real queda manual |
| **US5** esc. 4 (contraste, ampliación) | nuevo e2e: contraste WCAG (4.5:1 / 3:1 texto grande) calculado con luminancias en todo `main`+pie, 0 incumplimientos; texto 200 % conservado íntegro | ✅ |
| **FR-001…FR-007** (secciones y estructura) | e2e publica + trazas vivas + `service_public_test.go` (`FallbackPerField`, `OrderingAndLinks`, `OmitsEmptySections`, `OpenMediaPolicy`) | ✅ |
| **FR-003** (texto plano, ≤1.000) | traza viva 1.001 caracteres → `400` con límite; caracteres especiales (`ñ ¿ ¡ 😀`) preservados (unitaria `service_test.go:601`); React renderiza datos nunca HTML activo | ✅ |
| **FR-008/FR-009** (bilingüismo + fallback) | unitaria `TestGetPortadaFallbackPerField`, e2e pública en inglés, traza viva `lang=en` devuelve contenido resuelto | ✅ |
| **FR-012 / SC-005 / SC-011** | middleware de permiso `portada` + e2e panel completo con cuenta sin permiso; denegación auditada (`result='denied'`) | ✅ |
| **FR-013** (nunca exponer borradores) | e2e pública SC-002 + integración `IsHomeFilePublished` sigue el estado (imagen de identidad borrador → `404`, publicada → `200`, en vivo y e2e) | ✅ |
| **FR-015** (tabla de errores) | trazas vivas de esta sesión: correo inválido 400 · teléfono corto (unitaria) · grupo WhatsApp ajeno 400 · día 0–6 fuera 400 · `25:99` 400 · fin ≤ inicio 400 · texto 1.001 400 · solo inglés 400 · estado inválido 400 · red fuera de catálogo 400 · segundo enlace de red **409** · canal duplicado **409** (e2e panel lo ve en pantalla) · dominio de red ajeno 400 · sin guardar nada en todos | ✅ |
| **FR-017 / SC-013** (auditoría) | e2e panel: `home.image.upload`, `home.identity/about/contact/schedule/whatsapp.create/update`, `home.publish/unpublish`, `home.social.*`, `denied`; **solo lectura** confirmada: únicamente `GET /auditoria/acciones` y `GET /auditoria/accesos` registrados (`handler_audit.go:104-105`) | ✅ |
| **SC-002** | e2e pública + payload vivo sin rastro de borradores + `Cache-Control: no-store` en media y portada | ✅ |
| **SC-004** (< 2 min) | flujo e2e del panel completo (sembrar–publicar–ver) corre en ~8 s自动化 en máquina local | ✅ |
| **SC-007/SC-008** | e2e en 4 anchos + nuevo e2e de accesibilidad (teclado, contraste, alts, 200 %) · checklist §10 consignada abajo | ✅ con 1 hallazgo menor (A1) |
| **SC-009** | e2e pública: `wa.me/\d+` con `target=_blank rel=noopener`, `mailto:`, `tel:`, URL de Instagram exacta | ✅ |
| **SC-012** | e2e pública (`#whatsapp` y su ancla desaparecen sin publicados) + `TestGetPortadaOmitsEmptySections` | ✅ |
| **SC-010** (usabilidad con ≥6 personas) | **NO EJECUTABLE por automatización ni por esta sesión**: protocolo en `quickstart.md` §10.7; coordina el **humano + qa-tester** y se reporta en `pruebas-usabilidad-SC-010.md` | ⚠️ pendiente de ejecutar (no bloqueante para el código; bloqueante para cierre de F3) |

## Pruebas añadidas (solo archivos de prueba)

1. `backend/internal/portada/service_patch_null_test.go` — **3 pruebas** que decodifican el
   cuerpo JSON **exacto** del panel (`{"endTime": null}`, `{"nameEn": null}`) con
   `encoding/json` y ejercitan el merge: documentan B1. **FALLAN hoy** (así es como deben
   hundirse el defecto); pasan a verde al corregirlo y quedan como guardia de regresión:
   - `TestPatchScheduleJSONNullClearsEndTime` — `endTime: null` debe quitar la hora de fin.
   - `TestPatchWhatsappJSONNullClearsNameEn` — `nameEn: null` debe vaciar la traducción.
   - `TestPatchScheduleNullOnlyIsAccepted` — un PATCH de solo `null` no debe responder
     `400 «No hay cambios que guardar»` (el contrato exige `minProperties: 1`).
2. `frontend/e2e/portada-accesibilidad.spec.ts` — la parte automatizable de la checklist
   WCAG 2.1 AA del `quickstart` §10 sobre el stack real, sin dependencias nuevas: un solo
   `h1` + títulos `h2` por sección, `lang` correcto, todas las `<img>` con `alt` no vacío
   (logo con su alt del panel), área de toque ≥44 px de enlaces y botones a 320 px, recorrido
   de teclado completo con foco visible UnityEditor y cierre del ciclo (sin trampa), radios
   del idioma operables con flechas, contraste WCAG calculado en todo el texto de `main` y
   pie, texto al 200 % sin perder contenido. El resto de la suite e2e sigue en verde con él.

## Checklist WCAG 2.1 AA (SC-008, quickstart §10)

| Punto | Cómo se comprobó | Resultado |
|---|---|---|
| §10.1 teléfono/tableta/escritorio | e2e en 320/360/768/1280: secciones, enlaces y h1 completos; overflow ≤ 1 px | ✅ |
| §10.2 toques ≥ 44 px | nuevo e2e mide caja de cada `a`/`button` visible | ✅ (con hallazgo A1 en el selector de idioma) |
| §10.3 teclado y lector | nuevo e2e (foco visible en los 14 objetivos del ciclo, activación con Enter/flechas) | ✅ teclado · ⚠️ lector de pantalla real: queda para la revisión manual de cierre (no hay ninguna evidencia de fallo por diseño: landmarks, `aria-labelledby`, enlaces de salto, `sr-only`, alt) |
| §10.4 títulos, alt, avisos | nuevo e2e (estructura + alt) | ✅ |
| §10.5 contraste y ampliación | cálculo WCAG en vivo: 0 incumplimientos en `main` y pie; 200 % conservado | ✅ (con hallazgo M-a11y en reflow fino, abajo) |
| §10.6 sin menús ocultos | revisión del código `PublicLayout`: sin hamburguesa, anclas visibles siempre | ✅ |
| Exclusión consciente | el texto del hero sobre la imagen de portada (foto + capa `bg-navy/80`) se computa contra el fondo plano de `bg-navy` (cumple 14:1), pero la cardinalidad exacta depende de cada foto: se deja a la inspección visual del cierre | ⚠️ manual |

## Hallazgos

### BLOQUEANTE

**B1 — La semántica `null` del contrato en los PATCH no está implementada: el panel pierde en
silencio la limpieza de campos** *(coincide con `revision-2026-10-10-codigo.md`; reproducido y
evidenciado en vivo por QA)*.

- **Evidencia en vivo (stack real de esta sesión)**: con el servicio `QA-B1 Serv` publicado
  (`19:00–20:30`, id `d2aee50a-…`):
  1. `PATCH /api/v1/admin/portada/horario/{id}` con `{"endTime": null}` → **400 «No hay cambios
     que guardar»** (el contrato admite el objeto; limpiar ES un cambio).
  2. El estado tras el PATCH conserva `endTime: 20:30` (nunca se limpia).
  3. `PATCH` con `{"nameEn": null}` → 400 y `nameEn` intacto.
- **Causa**: `*string` no distingue `null` de campo ausente (`service_admin.go:153-182`); el
  panel SÍ envía `null` (`ServiceForm.tsx:87-89,131-137`, `nullable()`), cumple el contrato y
  tras «Guardado» el refetch devuelve el valor anterior.
- **Prueba añadida**: las 3 de `service_patch_null_test.go` fallan ahora y van a verde con la
  corrección. El repositorio queda con `go test ./internal/portada/` en rojo por estas pruebas
  hasta corregir B1 (es su propósito); el resto de la suite sigue en verde.
- **Corrección**: la del revisor (wrapper con `UnmarshalJSON` que distinga ausente/`null`/`""`,
  o cambio de contrato a `""` + ajuste del frontend, en un solo criterio).

### IMPORTANTE

- *(ninguno nuevo; el I1 del revisor de código — formato de presentación del horario prometido
  en `ux.md` §4.6/D-3 — sigue abierto y es del ciclo del revisor, no reevaluado aquí)*

Nota: QA no replica I1 como criterio de la spec (la spec no lo fija); queda registrado para el
humano.

### MENOR

- **A1 — Objetivo puntable del selector de idioma por debajo del mínimo**: los radios nativos
  son de 13 px y su `label` (el objetivo real de puntero) mide **20 px de alto**, por debajo del
  24 px de WCAG 2.5.8 «Target Size (Minimum)» (WCAG 2.2 AA) y muy lejos de los 44 px de la regla
  de la skill (quickstart §10.2). No incumple WCAG 2.1 AA (el mínimo de tamaño en 2.1 es nivel
  AAA), por eso es menor — pero es el control clave de la tarea 3 del protocolo SC-010
  («cambiar a inglés») para personas mayores. Corrección sugerida: p. ej. `className="size-6"`
  en el `input` o `min-h-6`/`py` en el `label` (`LanguageSwitcher.tsx`). La prueba de
  accesibilidad añadida falla en esta aserción hasta corregirlo.
- **M-a11y — Reflow fino al 200 % de zoom**: con el texto al 200 % sobre un teléfono de 320 px
  (viewport efectivo de 160 px), las tarjetas de contacto/enlaces (correo sin puntos de corte,
  `dl` con `dt/dd` largos) desbordan ~56 px y exigen scroll horizontal para llegar a su borde.
  No perdía contenido ni funcionalidad (WCAG 1.4.4 cumple) y el ancho prometido de 320 px no
  desborda (1.4.10 cumple), pero se sugiere `break-words`/`min-w-0` en esas tarjetas para una
  experiencia mejor a zoom alto. Medido con un probe Playwright de esta sesión.
- *(M1–M3 del revisor de código quedan como estaban: subidas de imagen rechazadas sin fila
  `failure`, ediciones sin cambio que registran `update`, huérfanos de imagen en guardados
  concurrentes; no se reevalúan como criterios de spec.)*

### Pendientes no ejecutables por esta sesión (no son defectos de código)

- **SC-010**: prueba manual con ≥6 personas (2 por franja de edad) — coordina el **humano**
  con `qa-tester`, protocolo en `quickstart.md` §10.7, informe en
  `specs/003-portada-info-general/pruebas-usabilidad-SC-010.md`. **La F3 no puede cerrar sin
  ella**; no la ejecuté por instrucción explícita.
- **Lector de pantalla real** (SC-008): pasada manual con NVDA/VoiceOver recomendada en el
  cierre; el código ya reúne landmarks, etiquetas, alt y foco visible.

## Lo positivo

Auditoría content-plan en la misma transacción (prueba de integración de atomicidad), catálogo
de redes cerrado y un enlace por red (409 vivo), imágenes por firma binaria con nombre
generado, la descarga pública dependiente del estado de publicación (`404`↔`200` vivo),
fallback `en→es` sin huecos, payload público irreconocible, i18n tipado, sesiones CSRF/cookies
correctas (401 vivo sin sesión) y 5 e2e + 269 unitarias de frontend totalmente repetibles en el
stack local.

## Veredicto

**RECHAZADO** — **1 BLOQUEANTE** (B1, reproducido en vivo y con pruebas de regresión añadidas,
rojas hasta corregirlo), **0 IMPORTANTE nuevos** (I1 del revisor sigue abierto), **2 MENORES
nuevos** (A1 del tamaño del selector de idioma y M-a11y de reflow al 200 %), más 1 pendiente
manual (SC-010) y 1 pendiente manual (lector de pantalla).

Ruta sugerida al orquestador: bucle de corrección 2 con `dev-backend` (B1) y `dev-frontend`
(A1/M-a11y si el humano lo aprueba); al estar en verde `service_patch_null_test.go` y la suite
e2e completa, vuelta de QA para el veredicto; SC-010 y el lector de pantalla real quedan para la
validación del cierre con el humano.
