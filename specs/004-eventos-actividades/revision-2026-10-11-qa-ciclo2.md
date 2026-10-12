# Revisión QA — F4 Eventos y actividades, ciclo 2 (2026-10-11)

**Rol**: `qa-tester` · **Rama**: `004-eventos-actividades` · **Alcance**: re-validación tras el ciclo de corrección 1 (I1 + M1–M3, commits `1714e2b`, `3ec0b2d`, `53c7007`, `8b03db3`).
**Veredicto: APROBADO**

## Suporte del ciclo de corrección

| Hallazgo del ciclo 1 | Fix implementado | Cobertura nueva |
|---|---|---|
| I1 (N+1 en fechas de agenda) | `1714e2b`: `ListUpcomingEventDatesForEvents` / `ListEventDatesForEvents` (batch `= ANY($1::uuid[])`); `GetProximamente` agrupa en memoria; `ListEventos` por página; `ListActividades` corte único | `TestGetProximamenteReadsDatesInOneBatchQuery` (25 eventos → 1 llamada al lote; sin eventos → 0), `TestListEventosReadsDatesInOneBatchQuery` (12 → 1 vacia → 0), `TestIntegrationListUpcomingEventDatesForEvents`, `TestIntegrationListEventDatesForEvents` — todas en verde |
| M1 (código muerto en test de portada) | `3ec0b2d`: retirado `IsHomeFilePublished` del fake | suite de `portada` íntegra |
| M2 (tope de categorías con `ListCategories` + carrera) | `53c7007`: `CountCategories` (CONTEO) y contrato de solo-lectura | `service_admin_test` verfica que el tope **no** lista el catálogo (`repo.listCategoriesCalls = 0`); integración pasa |
| M3 (`targetLabel` con UUID en `fail` de edición) | `8b03db3`: etiqueta legible del título previo | suite de `audit` íntegra |

**Nada más cambió**: `git diff d83e74b..HEAD` toca solo los ficheros listados arriba (agenda SQL/repository/service + sus pruebas). Ningún cambio en handlers públicos, frontend, migraciones ni compose.

## Suites ejecutadas (re-run completo)

| Suite | Comando | Resultado |
|---|---|---|
| Backend unitarias | `go test ./...` | ✅ ok en todos los paquetes |
| Backend integración (PostgreSQL real, `DATABASE_URL_TEST` de `.env`) | `go test -count=1 -tags=integration -p 1 ./...` | ✅ **20/20 paquetes ok, 0 FAIL**. Nota: en la primera pasada sin `-p 1`, `internal/platform/session` (testcontainers Redis) volvió a fallar por contención de paquetes en paralelo — es el mismo hallazgo menor 1 del ciclo 1, infra de pruebas, no de F4 |
| Cobertura `internal/agenda` | `-cover -tags=integration` | ✅ **86.6 %** (≥ 80 % de la constitución §III) |
| Frontend unitarias | `npm test -- --run` | ✅ **74 archivos / 505 pruebas** pasadas |
| E2e tanda 2 (F4 agenda, imagen Docker `mcr.microsoft.com/playwright:v1.63.0-jammy`) | `npx playwright test --config e2e/playwright.config.ts agenda- --workers=1` | ✅ **5 passed (18 s)** |
| E2e tanda 1 (F2) | `acceso auditoria --workers=1` | ✅ **2 passed (13 s)** |
| E2e tanda 3 (F3/F10) | `portada- diseno- movimiento navegacion- tema- qa- status --workers=1` | ✅ **31 passed (45 s)**. En una corrida intermedia falló 1 caso con un `toHaveURL` — al re-lanzar la tanda completa salió 31/31 en verde dos veces; se registra como flake aislado (no se identificó defecto de F4; probable contención con las otras tandas del mismo minuto) |

## Contrato de la vista pública sin cambios (focus del pedido)

Verificado por pruebas + en vivo contra el stack levantado (`docker compose` con los 4 servicios arriba):

- `GET /api/v1/proximamente` (sección vacía) → **`200`, `Cache-Control: no-store`, `{"lang":"es"}`** (bloques omitidos — igual que en el ciclo 1) ✅
- `lang=fr` → **400 `invalid`** con mensaje claro (`solo se admite es o en`) ✅
- Cobertura del comportamiento público (unitarias + integración, todas en verde): `TestGetProximamenteOnlyReturnsPublishedAndUpcoming` (solo `published` + próximo), `TestGetProximamenteAppliesTheNextRule`, `TestGetProximamenteDatesAreOnlyUpcoming`, `TestGetProximamenteOrdersChronologicallyWithStableTieBreak` (orden cronológico + desempate estable), `TestGetProximamenteFallsBackPerField` (fallback es por campo), `TestGetProximamenteCapsBlocksAtMaxBlockItems` (tope 200), handler de bloques vacíos (claves omitidas), `TestGetProximamenteNeverExposesPublicationState` (sin filtrar datos de publicación) ✅
- El fix I1 solo cambió **cómo** se leen las fechas (1 consulta por bloque en vez de una por evento), no **qué** se devuelve: la agrupación conservan el orden ascendente de la consulta y la defensa "evento sin fecha próxima no aparece" sigue activa ✅

## Panel (eventos / actividades / categorías) sin cambio de comportamiento

- Las únicas diferencias funcionales en el panel son las aprobadas por el ciclo: M2 cambia el **conteo** del tope (mismo 409 y mismo mensaje; nueva prueba asegura que el catálogo ya no se lista entero) y M3 cambia solo la **etiqueta legible** en la auditoría de los `fail` de edición (nada en la API ni en los campos).
- El e2e `agenda-panel` (3 casos) y las 12 pruebas de `EventsPage` / 8 de `ActivitiesPage` / 6 de categorías pasan íntegras: altas, `409` de duplicado y de categoría en uso, permiso `categorias`, auditoría.
- `sqlc-verify`/migraciones: sin cambios en `migrations/` en este ciclo (diff vacío), por lo que la migración 9 se mantiene tal cual está.

## Matriz de cobertura del ciclo 1 (mantiene el estado del ciclo 1)

La matriz completa del ciclo 1 (US1–US6, SC-001…SC-013) se mantiene: ninguna prueba se eliminó o debilitó (el diff solo **añade** pruebas nuevas) y todas las suites que las cubren pasan. Los 2 pendientes manuales persisten sin cambios:
- **SC-010** — protocolo de usabilidad con ≥6 personas (manual, fuera de `make ci`) ⚠️
- **Lector de pantalla físico** (`NVDA`/`VoiceOver` sobre `/eventos`) — lo automatizable (teclado, foco, jerarquía, `<time>`, `alt`, contraste y 200 % desde F10) está en verde; el criterio completo de SC-008 requiere la lectura manual ⚠️

## Hallazgos del ciclo 2

- **Menor (infra de pruebas, no de F4, ya conocido)** — contención de `testcontainers` entre paquetes en la corriente completa sin `-p 1`: `session` falló tras 167 s y en aislado pasa. Recomendación heredada: serializar las pruebas con testcontainers o seguir documentando la ejecución secuencial de `make ci`.
- **Menor (flake aislado de e2e)** — un caso de la tanda 3 falló con `toHaveURL` en la 1.ª de las 3 corridas y pasó en las 2 repeticiones completas. Sin reproducibilidad estable; se sugiere observar en el siguiente ciclo antes de abrir bug formal.

## Veredicto final

**APROBADO.** El ciclo de corrección cierra los 4 hallazgos (I1 con consultas por lote comprobadas y pruebas nuevas por lote; M1–M3) sin cambiar la API pública ni el panel, sin romper ninguna suite y manteniendo la cobertura ≥ 80 % y las 505 pruebas del frontend. Puede emitirse el PR vía `documentador` con los pendientes manuales anotados.
