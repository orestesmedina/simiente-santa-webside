# Revisión de código — 2026-10-12 (ciclo 3): 3 bugs + bloque «Lo que viene» (Q8)

- **Revisor:** `revisor-codigo` · **Rama:** `004-eventos-actividades`
- **Alcance:** commits `ef93c77..4250838` (3 arreglos de bug + F4 T447–T449) y, por encargo, el
  backend del extracto (`b9717f2`, `1a39c4e`, `e9b9401` — T443–T446) y el contrato 0.6.0.
- **Verificaciones ejecutadas:** `tsc --noEmit` ✓ · `eslint .` ✓ · `vitest run` 528/528 ✓ ·
  `go test ./internal/agenda/...` ✓. El e2e Playwright **no** se ejecuta en esta revisión
  (requiere servicios); queda para `qa-tester`.

## Veredicto: **APROBADO**

Sin hallazgos bloqueantes. Los dos hallazgos «Importante» son de coherencia **entre documentos**
(plan vs ux); el código es consistente con `ux.md` y sus pruebas.

## Bloqueante

Ninguno.

## Importante

1. **`plan.md:168` (P4-19) contradice `ux.md:579` (D-24) y `ux.md:307` — ubicación del bloque.**
   El plan dice «insertado tras el `IdentityHero` como primera pieza del `RevealGroup`» (y rechaza
   «bloque al final»); `ux.md` fija «entre «Horario de servicios» y «Canales de WhatsApp»» y
   rechaza explícitamente «tras el hero (aplasta la identidad)». La implementación
   (`frontend/src/features/publico/pages/HomePage.tsx:93-101`) y su prueba
   (`frontend/src/features/publico/homePage.test.tsx:134`) siguen D-24, que es la decisión ⟲ del
   alcance y lo que el encargo fija como autoridad: correcto. Pero el artefacto aprobado queda
   divergente y un ciclo futuro que siga el plan "corregiría" la ubicación. **Acción:** alinear
   `plan.md` P4-19 con D-24 (o llevar la discrepancia al humano).

2. **`plan.md:525` (RG4-11) vs `ux.md:317` (fila «Error sin nada en caché»): estado de error.**
   RG4-11 dice «el bloque se omite con el mismo mensaje del estado vacío»; `ux.md` y el código
   (`frontend/src/features/publico/eventos/ProximamenteBlock.tsx:80-91`) muestran el bloque con
   `Notice` de error + «Volver a intentar» (patrón F3). La implementación sigue `ux.md`, que es lo
   correcto; alinear la mitigación de RG4-11 para no dejar la semilla de un "arreglo" indebido.

## Sugerencia

3. `frontend/src/features/publico/components/PublicNav.tsx:15-21` — `useActivePath` llama al hook
   `useLocation` dentro de un `try/catch` para tolerar pruebas sin `Router`. Funciona y el linter
   pasa, pero es una llamada condicional encubierta y traga cualquier error. Más limpio: montar las
   pruebas de presentación en `MemoryRouter` y usar `useLocation()` directo.
4. `frontend/src/features/publico/eventos/hooks/usePortadaProximamente.ts:35-43` duplica el bloque
   de "retained" de `useUpcomingEvents.ts:36-43`. Extraer un `useRetainedQuery` compartido.
5. `frontend/src/features/publico/components/PublicLayout.tsx:50-68` (`useHashScrollOnArrival`):
   `handled` resuelve cada ruta+fragmento una sola vez por vida del layout; un regreso SPA a la
   misma ruta+hash no re-enrolla. Hoy no se manifiesta (los enlaces de sección son `<a>` nativos que
   recargan), pero conviene documentarlo o resetear `handled` al cambiar de `pathname`.

## Notas de verificación (sin hallazgo)

- **Pruebas congeladas conservan su intención:** `useSiteLinks.test.tsx:47,66,77-78,139-144` afirma
  ahora el ancla **absoluta** con las mismas aserciones (misma forma, valor actualizado);
  `e2e/navegacion-movil.spec.ts:253-256` solo cambia el selector y sigue exigiendo **0** enlaces a
  secciones ausentes; `qa-b1-siteLinks-relativas.test.tsx:50-57` solo completa el fixture con
  `detail` (requerido por el tipo), sin tocar lo que se afirma.
- **Bugs:** bug1 queda cubierto por dos capas (enlaces absolutos `/#ancla` en
  `useSiteLinks.ts:70-74` + salto al ancla tras cargar datos en `PublicLayout.tsx:50-68`, con
  reintento hasta encontrar el destino); bug2 (`aria-current` + `hover` AA, `PublicNav.tsx:30-43`);
  bug3 (`Dialog` `size="wide"` + `overflow-x-hidden`, `Dialog.tsx:18-29,105`, y columnas
  `minmax(0,1fr)`+`min-w-0` en `DatesEditor.tsx:42-76`).
- **Q8/FR-025–FR-027:** tope 5 y ventana de 2 meses con respaldo en **servidor**
  (`backend/internal/db/queries/agenda.sql` ListNextUpcomingItems: `in_window = EXISTS (…)` — las
  ramas ventana/respaldo nunca se mezclan —; `maxPortadaItems=5`, `portadaWindowMonths=2` en
  `service_public.go:369-375`); cliente solo presenta (`agenda.ts:81-85`); excepción a SC-012 con
  mensaje `role="status"` y botón oculto sin ítems (`ProximamenteBlock.tsx:93-110`); e2e
  `portada-proximamente.spec.ts` cubre RG4-12 (mensaje ⇔ entrada de nav oculta ⇔ sin botón) y
  RG4-14 (ventana y respaldo).
- **Capas/contrato:** handler público sin sesión, errores envueltos y 500 genérico
  (`handler_public.go`); contrato 0.6.0 coherente en `backend/api/openapi.yaml:37`,
  `contracts/openapi.yaml:107` y `schema.d.ts` regenerado; i18n completa en `es.ts`/`en.ts`.
- **Rendimiento:** `LIMIT 5` en SQL; una petición propia y en paralelo a la de F3 con
  `keepPreviousData` (RG4-11); sin consultas N+1 ni renders redundantes.
