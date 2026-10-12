# Estado: 004-eventos-actividades

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 004-eventos-actividades |
| Flujo | equipo-feature |
| Fase | 9/9 · Entregar (documentación lista; PR pendiente de abrir por el humano) |
| Ciclo de corrección | 1/3 (cerrado: los tres roles aprobaron en el ciclo 2) |
| Próximo paso | **El humano** hace `git push` y abre el PR (el entorno no tiene `gh`); al aprobarlo, `make costos CERRAR=1` y merge |
| Bloqueado por | — |
| Actualizado | 2026-10-11 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano | 2026-10-11 | v1: "apruebo la spec"; v2 (permiso «categorías»): "Re-apruebo la spec" |
| Plan | aprobado | humano | 2026-10-11 | "Apruebo el plan" (con las decisiones firmes: APP_TZ=America/Costa_Rica, ruta /eventos, aviso sin redirigir, permiso «categorías», endTime exige startTime, borrado de categoría en uso 409) |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-2026-10-11-ciclo2.md`).

- **(ninguno)** — los bloqueantes del ciclo 1 (I1, M1–M3) están corregidos y re-validados; los **tres roles APROBARON** en el ciclo 2.
- **Deuda aceptada (no bloquea)**: deriva documental **S1** (se corrige en convergencia), **S2** (falta una prueba de la rama de conflicto de `UpdateEvento`), **S3** (estilo); seguridad **B1/B2/B3**; revisor **M5** (`Optional[T]`); **QA**: SC-010 y lector de pantalla quedan como verificación **manual** con el humano.

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-11 — **Arranque de F4 (decisión del humano en sesión).** El humano indicó «continuemos con F4», el siguiente pendiente del roadmap (F1, F2, F3 y F10 ya terminadas). Se arranca el flujo `equipo-feature` en la rama `004-eventos-actividades`.
- 2026-10-11 — **Aclaraciones de F4 resueltas por el humano** (registradas en `spec.md`, sección «Aclaraciones resueltas (2026-10-11)»; también decisión 16 del roadmap): las 6 preguntas se resolvieron por la opción recomendada — **Q1** fechas concretas + texto de frecuencia; **Q2** hora de inicio/fin opcionales; **Q3** solo lo próximo; **Q4** retiro automático de la vista pública al pasar la fecha; **Q5** una entrada «Eventos y actividades» con dos bloques; **Q6** categoría opcional con catálogo administrable.
- 2026-10-11 — **Decisiones de la puerta 2 (plan) y cambio de alcance** (registradas en `plan.md` §"Puntos resueltos" y decisión 17 del roadmap): **`APP_TZ = America/Costa_Rica`**; **ruta pública `/eventos`**; **página con aviso sin redirigir** si no hay nada publicado; **permiso nuevo «categorías»** para el catálogo (amplía el catálogo de F2 de 9→10; el humano asumió el impacto: migración que siembra el permiso y actualización de las pruebas de F2 que fijan 9); `endTime` exige `startTime`; borrar categoría en uso → 409; imagen de elemento vencido no se sirve.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-11 — **Fase 5 (Validar) — ciclo 2: APROBADO por los tres roles.** El ciclo de corrección 1 (`1714e2b`, `3ec0b2d`, `53c7007`, `8b03db3`) cerró I1 (consultas por lote, sin N+1) y M1–M3. Re-validación: `qa-tester` APROBADO, `revisor-codigo` APROBADO, `seguridad` APROBADO. Consolidado en `revision-2026-10-11-ciclo2.md`. Quedan solo sugerencias no bloqueantes (S1 deriva documental, S2, S3) y deuda aceptada. Arranca la convergencia (Fase 8).

- 2026-10-11 — **Fase 5 (Validar) — ciclo 1**: `qa-tester` **APROBADO** (US1–US6 y SC-001…SC-013 cubiertos; SC-010 y lector de pantalla quedan manuales), `seguridad` **APROBADO** (0 críticos/altos/medios; 3 bajos + 3 informativos), `revisor-codigo` **APROBADO con cambios menores** (1 IMPORTANTE: N+1 en `GetProximamente`/`ListEventos`/`ListActividades`; 5 menores). Consolidado en `revision-2026-10-11.md`. El orquestador abre **ciclo de corrección 1** para I1 + M1–M3.

- 2026-10-11 — **T410 y T440 completadas** (`devops`): T410 documenta **`APP_TZ=America/Costa_Rica`** en `.env.example` (consumidor `platform/config`; compose no lo inyecta y el contenedor usa el default idéntico; commit `chore(infra): documentar APP_TZ en .env.example (F4 T410)`). **T440 — verificación de cierre en verde**: **`make ci`** completa (lint, go test unitarias + `-tags=integration` contra PostgreSQL real, vitest **74 archivos / 505 pruebas**, govulncheck 0, npm audit 0); **e2e de agenda** en 3 tandas (pública, panel, accesibilidad): **5 casos pasados** con `LD_LIBRARY_PATH` local contra el rate-limit de login; **cobertura `internal/agenda/service*.go` = 89,4 %** (≥ 80 % exigido); **`sqlc-verify` y `api-gen` sin deriva** (1 commit de diff: 0); **migraciones `000008`/`000009` up→down→up** (T416 en verde) y BD dev en versión 9; **`docker compose up`** con los 4 servicios arriba, `/healthz` 200, ruta inexistente `404` con `ErrorEnvelope`, `GET /api/v1/proximamente?lang=es` `200 no-store` con base recién migrada (`{"lang":"es"}` vacío, SC-012) y `lang=fr` → `400 invalid`; catálogo en la BD = **10 permisos con `categorias`**; **sin cambios de F4 en docker-compose.yml, Dockerfiles ni `.github/`** (árbol limpio). Pendiente registrado en hallazgos: SC-010 (manual con personas) y el fix local de librerías del navegador.
- 2026-10-11 — **Ajuste autorizado de un e2e de F2 por el cambio de producto de F4** (`dev-frontend`): `acceso.spec.ts` creaba un rol con «Eventos» + «Actividades» para comprobar el escenario «cuenta con módulos **reservados** → solo Inicio y el aviso». Con F4 esos dos permisos ya son módulos reales, así que el fixture pasa a dos módulos de F5–F9 («Noticias y galería» y «Donaciones»): el escenario y sus aserciones se conservan intactos (no se debilita nada). Misma naturaleza que la excepción aprobada RG4-10/T405 sobre las pruebas de F2 del catálogo 9 → 10.
- 2026-10-11 — **Fase 9 (e2e) completa** (`dev-frontend`): **T442** completada (`agenda-accesibilidad.spec.ts`: enlace de salto con el primer `Tab`, travesía con foco visible en cada paso, sin trampas de foco, radios del idioma con flechas y `Espacio`, menú móvil a 320 px y aviso de sección vacía alcanzable sin redirigir, en tema claro y oscuro). Los cinco casos de la agenda (T438/T439/T442) corren en verde juntos (`--workers=1`, 18 s). Los e2e se ejecutan con la imagen Docker oficial de Playwright (`mcr.microsoft.com/playwright:v1.63.0-jammy`, `--network host`), porque el host no tiene las librerías del navegador (`libnspr4`), y la suite completa se reparte en **tres tandas** —F2 (`acceso auditoria`), F4 (`agenda-`) y F3/F10 (`portada- diseno- movimiento navegacion- tema- qa- status`)— porque `POST /api/v1/auth/login` limita a 20 intentos/min por IP y con todo en una sola corrida los últimos specs reciben `429`. Documentado en la cabecera de los tres specs de F4. **T439** completada (`agenda-panel.spec.ts`: alta+publicación de un evento con fechas y una actividad desde la interfaz con el público viéndolo en la primera carga, edición con «ya visibles», retirar, eliminar con confirmación, catálogo de categorías con el `409` en uso que solo deja cerrar, permisos cruzados de los tres módulos —API 403— y auditoría con etiquetas legibles; 3 casos, 13 s). **T438** completada (`agenda-publica.spec.ts`: visitante, dos bloques, borradores/vencidos invisibles, imagen de borrador 404, inglés con fallback, 320/768/1280 sin scroll horizontal y aviso sin redirigir con el ítem oculto en barra y pie) + `e2e/agendaSeed.ts` (siembra y limpieza determinista de la agenda). El e2e destapó y se corrigió un defecto real de T433: el revelado lateral de los bloques no estaba dentro de un `RevealGroup` y la página ganaba 28 px de scroll horizontal (commit `fix(frontend): recortar el revelado lateral de /eventos`). Los e2e se ejecutan con la imagen Docker de Playwright (`--workers=1`): el host no tiene las librerías del navegador.
- 2026-10-11 — **Fase 8 (Frontend) en curso** (`dev-frontend`): **T429** completada (`api/agenda.ts` + hooks de eventos, actividades, categorías, subida de imagen y consulta pública `proximamente`; 27 pruebas nuevas) **T431** completada (promoción de `ImageUploader`/`PublishControls` al inventario único `src/components/`, sin cambios de comportamiento) **T432** completada (`formatoFechas.ts` con `Intl` sin conversión de zona + 16 claves i18n es/en de agenda) **T433** completada (página pública `/eventos` con los dos bloques, tarjetas, `DateBadge`/`MapLink`/`CategoryChip` y `HiddenNotice` sin redirigir) y **T434** completada (ítem «Eventos y actividades» con `href /eventos` en barra, menú móvil y mapa del pie, oculto sin contenido publicado y próximo; handler MSW por defecto para la consulta pública compartida). **T435** completada (panel de eventos: `EventsPage`, `EventsList`, `EventForm`, `DatesEditor`, marca «Pasado», `409` de duplicado y validaciones de fechas; 12 pruebas). Suite completa: 69 archivos / 461 pruebas en verde. **T436** completada (panel de actividades con fecha única, sin frecuencia, reutilizando `ItemFormFields`; 8 pruebas). **T437** completada (catálogo de categorías con alta/edición y borrado con `409` en uso que solo permite cerrar; 6 pruebas). **T430** completada (permisos `eventos`/`actividades`/`categorias`, rutas `/eventos` y `/panel/*` con `RequirePermission`, menú y `AccesosRapidos` filtrados por permiso y los 15 códigos de `ACTION_LABELS`). **Fase 8 completa**: 74 archivos / 505 pruebas en verde, `npm run lint`, `npm run typecheck` y `npm run build` sin errores.
- 2026-10-11 — **Puertas 1 (re-aprobación) y 2 superadas**: el humano **re-aprobó la spec** ("Re-apruebo la spec") y **aprobó el plan** ("Apruebo el plan"). Arranca la Fase 3 (tareas + coherencia).
- 2026-10-11 — **Fase 2 (Planificar)**: `plan.md`, `research.md`, `data-model.md`, `contracts/openapi.yaml` y `quickstart.md` redactados por `arquitecto`; `ux.md` por `disenador-ux`. 4 tablas `agenda_*`, 3 migraciones (`000007` contenido, `000008` auditoría, `000009` permiso «categorías»), 0 dependencias nuevas, reutiliza el sistema de F10 y la subida de imágenes de F3. **Cambio de alcance aprobado** en la puerta 2 (permiso «categorías»); spec actualizada → **re-aprobación pendiente**. **Puerta 2 pendiente** de aprobación humana del plan.
- 2026-10-11 — **Puerta 1 superada**: el humano **aprobó la spec** ("apruebo la spec"). Arranca la Fase 2 (plan técnico).
- 2026-10-11 — **Fase 1 (Especificar)**: `spec.md` redactada por `analista-producto` y **aclarada** (Q1–Q6 resueltas por el humano, todas por la opción recomendada). 6 user stories, FR-001–FR-024, SC-001–SC-013; sin marcas `[NECESITA ACLARACIÓN]`. En la **puerta 1**, pendiente la aprobación humana de la spec. Rama `004-eventos-actividades` creada; roadmap: F4 → **en curso**.
- 2026-10-11 — Inicio de la funcionalidad F4 (eventos y actividades): rama `004-eventos-actividades` creada, `spec.md` desde plantilla. Fase 1 (especificar).
