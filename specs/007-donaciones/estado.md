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
| Rama | 007-donaciones |
| Flujo | equipo-feature |
| Fase | 7/9 · Entregar (PR #10 abierto; CI en rojo por un flaky de infraestructura de pruebas — decisión pendiente del humano) |
| Ciclo de corrección | 2/3 · bloqueado y escalado (C1 cerrado; C2: 3 intentos de arreglo revertidos) |
| Próximo paso | Decidir con el humano cómo cerrar C2: volver a lanzar el CI (y registrar el flaky como defecto conocido) o implementar el aislamiento por esquema / proponer al kit el `-p 1` en el CI |
| Bloqueado por | Flaky preexistente del CI (C2): las pruebas de migración destructivas compiten en la base compartida |
| Bloqueado por | — |
| Actualizado | 2026-10-16 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano (cliente) | 2026-10-14 | "si" (aprueba la spec v2 de F7) |
| Plan | aprobado | humano (cliente) | 2026-10-14 | "apruebo" (aprueba el plan de F7; acepta los defaults presentados: IBAN solo estructura/longitud, SINPE 6/7/8, sección visible con solo textos) |
| PR / merge | abierto (pendiente de aprobación humana) | — | 2026-10-16 | PR #10 abierto: https://github.com/orestesmedina/simiente-santa-webside/pull/10 |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno: M1 resuelto y confirmado en el ciclo 1; ver [revision-2026-10-15.md](./revision-2026-10-15.md))
- **C2 (bloqueante, CI; escalado al humano)** — el CI del PR #10 falla de forma intermitente con `deadlock detected (40P01)` en `TestIntegrationMigration000008UpDownUp` (`internal/agenda`). Causa raíz: en CI las pruebas de integración corren **en paralelo contra una sola base** (`DATABASE_URL_TEST`) y varias `TestIntegrationMigration*UpDownUp` (portada 000006, agenda 000008/000009, ministerios 000010/000011, donaciones 000012/000013) ejecutan un `down` con `DELETE FROM admin_actions` (RowExclusive) + `ALTER TABLE admin_actions` (AccessExclusive) en la misma transacción: dos de ellas se bloquean en ciclo. Es **preexistente** (F3/F4/F6 ya tenían 3 ciclos); F7 añadió el cuarto y la hizo probable. Modo de fallo secundario: `23514` cuando el `down` de 000006 restaura el CHECK de F2 y otra prueba confirmó filas con códigos de F3–F7. Evidencia medida: el código del PR pasa **15/15** y **3/3** suites completas en una base **fresca** (el fallo es raro y depende de la concurrencia). Intentos de arreglo **revertidos** (el backend quedó idéntico a `119ffb3`): (1) lock de aviso entre ciclos → elimina el 40P01 pero no el 23514; (2) `ACCESS EXCLUSIVE` + borrado → **peor** (timeouts en cascada por bloquear `admin_actions` todo el ciclo); (3) base aislada por prueba (`CREATE DATABASE`) → **peor** (el `CREATE/DROP DATABASE` por prueba satura el servidor). Arreglo propuesto: **aislar cada ciclo en su propio esquema** (barato) o **cambiar el CI del kit para correr la integración en serie (`-p 1`)** — `ci.yml` es archivo del kit y no se edita en el proyecto.

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-14 (humano): se construye **F7 (donaciones) antes que F5 (grupos de conexión)**. F5 queda pendiente; ambas son independientes (roadmap §3).
- 2026-10-14 (humano): aclaraciones de F7 (Q1–Q11) resueltas; registradas como **decisión 22 del roadmap**. Cambios de alcance: sección «Ofrendas y diezmos» en `/ofrendas` (Q8) y bloque en la portada debajo del de ministerios (Q7).
- 2026-10-14 (humano): al aprobar el plan (puerta 2), acepta los defaults del arquitecto: **IBAN se valida solo por estructura y longitud** (sin mod-97), **SINPE Móvil exige primer dígito 6/7/8** (celular, no línea fija), **la sección se muestra aunque solo haya textos publicados y 0 métodos** (bloque de portada oculto), y **duplicados por tipo + número** (`409`); los textos no se borran, solo se retiran a borrador.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

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
