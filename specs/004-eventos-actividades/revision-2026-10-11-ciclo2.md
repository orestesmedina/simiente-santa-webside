# Revisión consolidada (ciclo 2) — F4 Eventos y actividades (2026-10-11)

Re-validación tras el **ciclo de corrección 1**. Reportes:
`revision-2026-10-11-{qa,codigo,seguridad}-ciclo2.md`.

## Veredicto global

**APROBADO** — los tres roles aprueban. El ciclo de corrección 1 cerró todos los bloqueantes.

| Rol | Veredicto | Nota |
|---|---|---|
| `qa-tester` | APROBADO | Go unit + integración 20/20 (`-p 1`), Vitest 505/505, e2e 5/5 F4 (+ F2 2/2, F3/F10 31/31); vista pública intacta |
| `revisor-codigo` | APROBADO | I1 cerrado (consultas por lote, 1 por bloque, sin SQL muerto); M1–M3 cerrados; sin pruebas congeladas tocadas ni debilitadas |
| `seguridad` | APROBADO | SQL parametrizado, sin fuga, `govulncheck`/`npm audit` sin altas/críticas, sin secretos |

## Bloqueantes

Ninguno. Los hallazgos del ciclo 1 quedan cerrados:

- **I1** — `ListEventDatesForEvents`/`ListUpcomingEventDatesForEvents` con `event_id = ANY($1::uuid[])`: 1 consulta por bloque/página (25 eventos → 1 llamada), sin N+1 ni SQL huérfano; comportamiento (orden, filtro «próximo», claves omitidas, fallback es) sin cambios. Commit `1714e2b`.
- **M1** — retirado el código muerto `IsHomeFilePublished` del fake de portada. `3ec0b2d`.
- **M2** — `CountCategories` + tope por conteo (1 conteo / 0 listados). `53c7007`.
- **M3** — etiquetas de auditoría con el título/nombre actual en los `fail`. `8b03db3`.

## Sugerencias (no bloquean el merge)

- **S1 — Deriva documental**: `data-model.md` y `tasks.md` aún nombran `ListUpcomingEventDates` (retirada) y no citan `CountCategories`. Se corrige en la convergencia.
- **S2** — la nueva rama de conflicto de `UpdateEvento` (`service_admin.go:263-266`) no tiene una prueba que la fije. Deuda menor.
- **S3** — comentario de sección sin línea en blanco (estilo).

## Deuda aceptada (de los ciclos 1–2)

- Seguridad **B1** (PATCH no revalida longitud de campos opcionales; la BD corta con 400 genérico), **B2** (`imageFile` por patrón, UUID no enumerable), **B3** (contraseña de prueba en e2e).
- Revisor **M5** (`Optional[T]` duplicado).
- **QA**: SC-010 (usabilidad con ≥6 personas) y lectura con lector de pantalla físico → verificación **manual** con el humano.

## Próximo paso

Convergencia: 42/42 tareas hechas; corregir la deriva documental (S1) y pasar a la entrega
(documentación + PR). El merge y el despliegue los aprueba el humano.
