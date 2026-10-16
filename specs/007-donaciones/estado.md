# Estado: 007-donaciones

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 007-donaciones (eliminada tras el merge) |
| Flujo | equipo-feature |
| Fase | 9/9 · Terminada (PR #10 **fusionado en `main`** el 2026-10-16; rama eliminada) |
| Ciclo de corrección | 2/3 · cerrado (M1 resuelto; C2 → defecto conocido 23 del roadmap) |
| Próximo paso | Nada para F7. Siguiente funcionalidad del roadmap: **F5 (grupos de conexión)** |
| Bloqueado por | — (C2 reclasificado como defecto conocido, no bloqueante, por decisión del humano el 2026-10-16) |
| Bloqueado por | — |
| Actualizado | 2026-10-16 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano (cliente) | 2026-10-14 | "si" (aprueba la spec v2 de F7) |
| Plan | aprobado | humano (cliente) | 2026-10-14 | "apruebo" (aprueba el plan de F7; acepta los defaults presentados: IBAN solo estructura/longitud, SINPE 6/7/8, sección visible con solo textos) |
| PR / merge | aprobado | humano (cliente) | 2026-10-16 | "has el cierre, yo hago el merge" — el humano hará el merge |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno: M1 resuelto y confirmado en el ciclo 1; ver [revision-2026-10-15.md](./revision-2026-10-15.md))

## Defectos conocidos (no bloqueantes)

- **C2 — flaky del CI en las pruebas de migración destructivas (preexistente).** En CI las pruebas de integración corren **en paralelo contra una sola base** (`DATABASE_URL_TEST`), y varias `TestIntegrationMigration*UpDownUp` ejecutan un `down` con `DELETE FROM admin_actions` (RowExclusive) + `ALTER TABLE admin_actions` (AccessExclusive) en la misma transacción; dos de ellas se solapan y producen `deadlock detected (40P01)` (o, en menor medida, `23514`). Es **anterior a F7** (F3/F4/F6 ya tenían 3 ciclos; F7 añadió el cuarto) y **se decidió no corregirlo en F7** (opción A del humano, 2026-10-16): el PR se cierra con el flaky registrado. Evidencia: el código pasa 15/15 y 3/3 suites completas en base fresca. Arreglos propuestos (no aplicados): **aislar cada ciclo en su propio esquema** o **correr la integración en serie (`-p 1`) en el CI del kit** (`ci.yml` es archivo del kit y no se edita aquí). Intentos revertidos: lock de aviso entre ciclos; `ACCESS EXCLUSIVE` + borrado; base de datos aislada por prueba.

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-14 (humano): se construye **F7 (donaciones) antes que F5 (grupos de conexión)**. F5 queda pendiente; ambas son independientes (roadmap §3).
- 2026-10-14 (humano): aclaraciones de F7 (Q1–Q11) resueltas; registradas como **decisión 22 del roadmap**. Cambios de alcance: sección «Ofrendas y diezmos» en `/ofrendas` (Q8) y bloque en la portada debajo del de ministerios (Q7).
- 2026-10-14 (humano): al aprobar el plan (puerta 2), acepta los defaults del arquitecto: **IBAN se valida solo por estructura y longitud** (sin mod-97), **SINPE Móvil exige primer dígito 6/7/8** (celular, no línea fija), **la sección se muestra aunque solo haya textos publicados y 0 métodos** (bloque de portada oculto), y **duplicados por tipo + número** (`409`); los textos no se borran, solo se retiran a borrador.
- 2026-10-16 (humano): ante el flaky del CI (C2) y tras los 3 intentos de arreglo revertidos, elige la **opción A**: relanzar el CI y **dejar C2 registrado como defecto conocido** (no bloquear el cierre de F7). El arreglo de raíz (aislar por esquema o `-p 1` en el CI del kit) queda para después.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-16 — **F7 cerrada**: PR #10 fusionado (squash) en `main` (`5f313f0`) y rama eliminada. Roadmap → terminada.
- 2026-10-16 — **Puerta 3: PR aprobado** por el humano ("has el cierre, yo hago el merge"). Costo de la tarea **cerrado** (225,6 M tokens ≈ $4,98). El merge lo hace el humano.
- 2026-10-16 — CI del PR #10 **en verde** (corrida 38976039295): los 6 trabajos pasan y el flaky 40P01 no reapareció. Pendiente solo la aprobación humana del merge.
- 2026-10-16 — Decisión del humano (opción A): C2 queda como **defecto conocido** y se relanza el CI. Ciclo 2 cerrado.
- 2026-10-16 — C2 (flaky del CI): 3 intentos de arreglo revertidos (lock de aviso; AccessExclusive; base aislada por prueba) — el backend vuelve al estado validado del PR. Escalado al humano.
- 2026-10-16 — Fase 7 (entrega): documentación (CHANGELOG 0.7.0, README, nota de entrega y protocolo SC-012) y **PR #10 abierto**; pendiente la aprobación humana del merge.
- 2026-10-16 — Ciclo de corrección 1 cerrado (M1 resuelto en `7f05e40`, confirmado por el revisor). Fase 6 (convergencia): no hay tareas nuevas. Inicio de la entrega (T738).
- 2026-10-15 — Fase 5: validación con QA/código/seguridad (los tres aprueban). Fase 4: implementadas T701–T737; `make ci` en verde; cobertura `service*.go` 89,6 %.
- 2026-10-14 — Fase 3: `tasks.md` (38 tareas) y `analyze` (veredicto: sin críticos; 3 altos A1/A2/A3 corregidos) — fase cerrada. Fase 4 (implementar) iniciada.
- 2026-10-14 — Puerta 2: **plan aprobado** por el humano (frase: "apruebo"); defaults aceptados. Fase 3 (tareas y coherencia) iniciada.
- 2026-10-14 — Fase 2: `arquitecto` produjo plan/research/data-model/contracts/quickstart y `disenador-ux` produjo `ux.md`. Corregida una contradicción interna de la spec (entidad financiera no traducible). Pendiente la puerta 2 (aprobación del plan).
- 2026-10-14 — Puerta 1: **spec v2 aprobada** por el humano. Fase 2 (planificar) iniciada.
- 2026-10-14 — Fase 1 (clarify): las 11 preguntas resueltas y aplicadas; `spec.md` pasa a v2 sin marcas abiertas. Pendiente la puerta 1 (aprobación de la spec).
- 2026-10-14 — Fase 1: rama `007-donaciones` creada y `spec.md` redactada por `analista-producto` (25 FR, 13 SC, 5 historias). 11 preguntas abiertas (Q1–Q11) para `clarify`.
- 2026-10-14 — Inicio de la funcionalidad.
