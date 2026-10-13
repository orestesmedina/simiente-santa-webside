# Revisión de seguridad — F4: re-auditoría (3 bugs + bloque de portada Q8)

- **Rama**: `004-eventos-actividades` · **Rango revisado**: `5cd6da0..HEAD` (todo lo posterior al ciclo 2 aprobado: contrato/SQL/service/handler del extracto —`7e46d6b`, `b9717f2`, `1a39c4e`, `e9b9401`—, cliente/hook/bloque/e2e —`3619436`, `3ebb183`, `e74d086`, `907d6cd`— y los arreglos de los 3 bugs —`ef93c77`, `72665e2`, `e1f6fed`, `66c8472`, `4348875`—)
- **Fecha**: 2026-10-12 · **Revisor**: `seguridad` (re-validación)
- **Referencias**: `revision-2026-10-11-seguridad.md` (ciclo 1, B1–B3), `revision-2026-10-11-seguridad-ciclo2.md` (APROBADO), `.specify/memory/constitution.md` §IV, `spec.md` FR-013/FR-025…FR-027/SC-014.
- **Método**: lectura completa del diff y de los archivos vigentes (routes, handler_public, service_public, repository_public, `queries/agenda.sql`, sqlc generado, componentes del bloque), greps de SQL dinámico/`dangerouslySetInnerHTML`/`innerHTML`/`eval`/`window.open`/secretos sobre el rango, y ejecución de `govulncheck ./...` y `npm audit --audit-level=high`.

## Veredicto: **APROBADO**

Sin críticos, altos, medios ni bajos; un informativo preexistente. El endpoint público y el bloque no amplían la superficie de ataque y los 3 arreglos de bugs son de presentación/navegación sin riesgo de seguridad.

---

## Verificación de los puntos pedidos

### 1. El endpoint público no expone borradores ni vencidos — ✅

- `GET /api/v1/portada/proximamente` se registra en `RegisterPublic` (`routes.go:42`): solo lectura, **sin** middlewares de sesión (solo los globales), igual que `GET /api/v1/proximamente`.
- `ListNextUpcomingItems` (`queries/agenda.sql:371-432`) exige `publication_state = 'published'` en **ambas** ramas (eventos `:397`, actividades `:411`) y la regla R4-2 de corte (`fecha > today` o `= today` con `end_time >= now_time`, `:375-379` y `:413-417`): ni borradores ni vencidos —ni siquiera los de hoy ya terminados— cruzan al público.
- La ventana de 2 meses con respaldo (`in_window = EXISTS(…)`, `:427-429`) solo **reordena qué** próximos se muestran; nunca relaja el filtro publicado/próximo.
- Probado en integración: `repository_public_integration_test.go:150-187` siembra borrador, vencido, actividad borrador, actividad vencida y evento de hoy ya terminado, y afirma que ninguno aparece (FR-013/SC-014).
- `handler_public.go:72-99`: `lang` validado por lista cerrada `es|en` (400 con `details.lang`), `Cache-Control: no-store`, `serviceConfigured` falla cerrado (`handler.go:209-216`).

### 2. SQL parametrizado — ✅ (CWE-89)

- Todo va por `sqlc.arg()`: el código generado (`db/agenda.sql.go:748+`) usa `$2::date`, `$3::time`, `$4::date` (window_end) y `$5` (lim). Cero concatenación o SQL dinámico en el rango (greps vacíos).
- `repository_public.go:23-38` pasa valores tipados (`pgDate`/`pgTime`); `lim` es la constante del servidor `maxPortadaItems = 5` (`service_public.go:371`), no un parámetro de la petición.
- El corte `today/nowTime` y `windowEnd` se calculan en el service en `APP_TZ` (`cutoff`/`portadaWindowEnd`), no desde la entrada.

### 3. Sin fuga de datos ni información interna — ✅

- `ItemProximoPortada` (`service_public.go:390-399`) solo lleva `kind`, `id`, `title`, `date`, `startTime`, `endTime`, `place` e `image`{url,alt}: **sin** `publicationState`, `detail`, lista de fechas, organizador ni categoría (coincide con el contrato 0.6.0, `additionalProperties: false`).
- Sin `items` la clave se omite (nunca `null`); errores vía `apperr` + `httpserver.WriteError` (los 500 no exponen la causa; `serviceConfigured` → Internal genérico).
- La imagen viaja como ruta relativa `/api/v1/media/<file>` (`localizeImage`), la misma superficie ya auditada de F3/F4.

### 4. Secretos — ✅ (CWE-798)

- Grep del rango `password|secret|api_key|token|Bearer`: las coincidencias `"tokens"` son `specs/…/costos.json` (métricas de IA, no credenciales); `"Semilla.2026"` es la **credencial demo sembrada y documentada** de F1 (patrón preexistente, deuda B3 ya aceptada); el e2e nuevo `portada-proximamente.spec.ts` ni siquiera hace login (flujo de visitante). Sin secretos nuevos.

### 5. Dependencias — ✅

- `govulncheck ./...`: **0 vulnerabilidades que afectan al código** (1 a nivel de módulo no llamada: GO-2026-5932 `openpgp`, el informativo I1 preexistente del ciclo 2).
- `npm audit --audit-level=high`: **0 vulnerabilidades**.

### 6. Los arreglos de bugs no introducen problemas — ✅

- **Bug 1** (`useSiteLinks.ts:73`): `href = "/#" + anchor`. El prefijo `/#` hace **imposible** un esquema inyectable (`javascript:`/`data:`): el resultado siempre es ruta relativa al origen + fragmento. El `anchor` viene de la API pública de portada (contenido publicado), igual que antes; el cambio reduce riesgo (relativo → absoluto). El scroll al ancla (`PublicLayout.tsx:31-52`) usa `getElementById` + `scrollIntoView`: sin vector de inyección.
- **Bug 2** (`PublicNav.tsx`): `aria-current` se calcula del `pathname` del router contra `href` estáticos; solo presentación/accesibilidad. Sin cambios de superficie.
- **Bug 3** (`Dialog.tsx`): prop `size` + `overflow-x-hidden` — solo CSS/layout. `EventForm`/`ActivityForm` pasan `size="wide"`; `ImageUploader` un ajuste de clase.
- **Bloque**: `ProximamenteBlock`/`UpcomingItem` renderizan el DTO del servidor como texto (React escapa; cero `dangerouslySetInnerHTML`/`innerHTML`/`eval`/`window.open` en el rango —la única mención es una celda de tabla de documentación—); `Link to` con constantes `/eventos#eventos`/`/eventos#actividades`; botón «Ver más» con `href="/eventos"` estático; `CardImage` usa `mediaUrl` sobre la URL del servidor; `buildUrl` codifica con `URLSearchParams`.
- El handler MSW de pruebas (`test/server.ts`) devuelve una agenda vacía: sin datos sensibles.

---

## Hallazgos

### Críticos / Altos / Medios / Bajos

Ninguno.

### Informativos

**I1 (preexistente) — GO-2026-5932 en `golang.org/x/crypto@v0.57.0` (paquete `openpgp`, unmaintained).**
`govulncheck` la reporta a nivel de módulo con 0 símbolos afectados. Sin cambios de `go.mod` en el rango; deuda de dependencia indirecta ya registrada en el ciclo 2, sin acción en F4.

**I2 — El extracto repite una lectura pública independiente de `/api/v1/proximamente`.**
Dos endpoints públicos de solo lectura con el mismo filtro R4-2; superficie coherente y `no-store` en ambos. Solo recordatorio de hygiene: si en el futuro se añade rate-limit a las rutas públicas, incluir `/api/v1/portada/proximamente` (hoy ninguna ruta pública lo tiene y no se introdujo nada nuevo por este cambio).

## Conclusión

El extracto de la portada hereda intacta la garantía de F4 —solo publicado y próximo, SQL 100 % parametrizado por sqlc, DTO mínimo sin datos internos— y está respaldado por pruebas de integración que siembran borradores y vencidos y comprueban su ausencia. Los 3 arreglos de bugs son de navegación/presentación (href prefijado a `/#`, `aria-current`, CSS del modal) sin vectores de inyección ni de control de acceso. Sin secretos nuevos y dependencias sin altas/críticas. **APROBADO.**
