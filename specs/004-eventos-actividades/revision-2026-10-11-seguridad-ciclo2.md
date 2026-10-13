# Revisión de seguridad — F4 Eventos y actividades (ciclo 2)

- **Rama**: `004-eventos-actividades` · **Rango revisado**: `d83e74b..HEAD` (ciclo de corrección del backend: `1714e2b`, `3ec0b2d`, `53c7007`, `8b03db3`; 13 archivos, +598/−96, sin cambios en `go.mod`, frontend ni infraestructura)
- **Fecha**: 2026-10-11 · **Revisor**: `seguridad` (re-validación tras el ciclo de corrección)
- **Referencias**: `revision-2026-10-11-seguridad.md` (ciclo 1, APROBADO con B1–B3), `.specify/memory/constitution.md` §IV (CWE-89), `spec.md` R4-2/R4-9/R4-10.
- **Método**: lectura completa del diff de los 4 commits (SQL sqlc, repository, services, pruebas), greps de SQL dinámico/concatenación y de secretos sobre el rango y el árbol, verificación del filtro público en `queries/agenda.sql` vigente, y ejecución de `govulncheck ./...` y `npm audit --audit-level=high`.

## Veredicto: **APROBADO**

El ciclo de corrección no introduce ningún hallazgo de seguridad. Sin críticos, altos, medios ni bajos; dos informativos. El veredicto APROBADO del ciclo 1 se mantiene.

---

## Verificación de los puntos del ciclo

### 1. Consultas por lote parametrizadas — `1714e2b` (CWE-89)

- `ListEventDatesForEvents` (`backend/internal/db/queries/agenda.sql:206-213`): `WHERE event_id = ANY(sqlc.arg('event_ids')::uuid[])`, generada por sqlc como `q.db.Query(ctx, ..., eventIds)` con placeholder `$1`. Sin concatenación.
- `ListUpcomingEventDatesForEvents` (ídem, `:215-227`): `ANY($1::uuid[])` y regla R4-2 con `today`/`now_time` como `$2::date`/`$3::time` parametrizados (corte calculado por el service en `APP_TZ`). Sin concatenación.
- El helper `pgUUIDs` (`backend/internal/agenda/repository.go:133-139`) solo mapea `uuid.UUID → pgtype.UUID`; no construye SQL.
- La variante por evento `ListUpcomingEventDates` se retiró por completo: sin referencias residuales en `internal/`.
- Grep de `fmt.Sprintf`/concatenación con SQL en producción del backend: sin coincidencias.

### 2. Filtro «publicado y próximo» y separación público/panel — intactos

- `ListUpcomingEvents` (única vía pública de eventos) no fue tocada y sigue exigiendo `WHERE e.publication_state = 'published'` (`queries/agenda.sql:160`).
- Público (`service_public.go` `GetProximamente`): lee **solo fechas próximas** del bloque vía `ListUpcomingEventDatesForEvents` y mantiene la defensa de omitir eventos sin fecha próxima (FR-002/FR-013). Sin fechas pasadas ni borradores.
- Panel (`service_admin.go` `ListEventos`): usa `ListEventDatesForEvents` (todas las fechas) — correcto, es superficie autenticada del panel; ningún dato interno cruza al servicio público.
- `CountCategories` — `53c7007`: SQL estático `SELECT count(*)::bigint FROM agenda_categories`, sin entrada de usuario. Sin riesgo.

### 3. Etiquetas de auditoría — `8b03db3`

Cambia `target_label` de UUID a título/nombre **actual** del elemento en los rechazos (y el nombre actual, no el nuevo, en los conflictos de la base). No toca SQL, permisos ni exposición: el título ya se registraba en las filas de éxito, la tabla de auditoría es de acceso restringido al panel y el renderizado de React escapa por defecto. Sin fuga nueva. Pruebas: `TestRejectionLabelsUseTheTitle` (5 casos, todos `result='failure'`).

### 4. Código muerto — `3ec0b2d`

Solo elimina `IsHomeFilePublished` y `publishedFiles` del falso de pruebas de portada. Sin impacto en producción ni en la política de publicación (cubierta por `internal/media`).

### 5. Fugas de datos e información interna

Ninguna: el diff no añade campos a los DTO públicos, no expone errores internos nuevos (`wrap`/`classify` sin cambios) y no toca CORS, sesiones ni cabeceras.

### 6. Dependencias

- `govulncheck ./...`: **0 vulnerabilidades que afectan al código** (symbol results limpio).
- `npm audit --audit-level=high`: **0 vulnerabilidades**.

### 7. Secretos

Grep del diff `d83e74b..HEAD` y del árbol (`password`, `secret`, `api_key`, `token`, claves privadas, URLs con credenciales): **sin coincidencias**. Árbol de trabajo limpio.

---

## Hallazgos

### Críticos / Altos / Medios / Bajos

Ninguno.

### Informativos

**I1 — GO-2026-5932 preexistente en `golang.org/x/crypto@v0.57.0` (paquete `openpgp`, unmaintained).**
- `govulncheck` la reporta a nivel de módulo pero confirma que el código no lo llama (0 en símbolos). No fue introducida por este ciclo (ningún cambio en `go.mod`/`go.sum` en `d83e74b..HEAD`). Deuda de dependencia indirecta, sin acción en F4.

**I2 — Las etiquetas de rechazo ahora registran el título/nombre del contenido en `target_label`.**
- Superficie ya existente (las filas de éxito ya lo hacían) y de acceso restringido. Solo recordatorio: el panel de auditoría debe seguir renderizándolo con escape estándar de React (ya lo hace); no introducir `dangerouslySetInnerHTML` ahí en el futuro.

## Conclusión

El ciclo de corrección fue de rendimiento (N+1 → lote) y legibilidad de auditoría; no amplía la superficie de ataque. Todas las consultas nuevas son sqlc parametrizadas, el filtro público R4-2 y la separación público/panel quedan intactos, las dependencias siguen sin altas/críticas y no aparecen secretos. **APROBADO** (mantiene el veredicto del ciclo 1).
