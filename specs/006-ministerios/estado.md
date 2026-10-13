# Estado: 006-ministerios

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 006-ministerios |
| Flujo | equipo-feature |
| Fase | 2/9 · Planificar — documentos escritos; **puerta 2 pendiente** |
| Ciclo de corrección | 0/3 |
| Próximo paso | Puerta 2: el humano responde (a) el **default del interruptor `leaderVisible`**, (b) si `/ministerios` sin nada publicado muestra **aviso sin redirigir** (patrón de F4) y (c) **aprueba el plan**; después, Fase 3 (`tasks` + `analyze`) |
| Bloqueado por | — |
| Actualizado | 2026-10-13 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | **aprobado** | humano | 2026-10-13 | v1: «Apruebo la spec» respondiendo las aclaraciones Q1–Q7 (la v2 amplió el alcance y se pidió re-aprobación, como en F4). v2: **«Re-apruebo la spec v2»**, respondiendo las aclaraciones nuevas Q8 (el medio de contacto es opcional aunque haya encargado) y Q9 (el interruptor de visibilidad oculta también el nombre del encargado) |
| Plan | pendiente | | | |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno)

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-13 — **Arranque de F6 (decisión del humano en sesión).** El humano pidió «continuemos con el f6» en lugar de F5 (el siguiente pendiente del roadmap). F4, F5, F6 y F7 son independientes entre sí según el roadmap §3, así que se arranca F6 — Ministerios por el flujo `equipo-feature`. Numeración de la rama: **`006-ministerios`**, alineada con el identificador de roadmap (precedente: F10 → `010-sistema-diseno-movimiento`), dejando el número 005 libre para F5 (grupos de conexión).

- 2026-10-13 — **Aclaraciones de F6 resueltas por el humano (Q1–Q7)**, en su mayoría **distintas de la opción recomendada**, con **ampliación de alcance** en tres puntos. Respuestas literales: **Q1** encargado **opcional** con nombre y un medio de contacto (celular o correo) y un **interruptor para decidir si el contacto se muestra**; **Q2** enlace de WhatsApp **opcional** (un ministerio puede publicarse sin enlace); **Q3** se acepta **cualquier `https` válido y seguro** (sin restringir a dominios de WhatsApp); **Q4** imagen **opcional** cargada por el equipo (única respuesta igual a la recomendación); **Q5** orden **manual con una columna numérica** de orden —precedente verificado en el repositorio: `sort_order INTEGER NOT NULL DEFAULT 0` con `ORDER BY sort_order, id` en `home_services`/`home_whatsapp_channels` (migración `000005`) y `agenda_categories` (migración `000007`)—; **Q6** **se añaden horario y lugar** con la particularidad de que el horario varía por ministerio y a veces no es un día fijo («alabanza: todos los viernes 7 p. m.» vs. «matrimonios: una vez al mes, fecha variable»), interpretado como **horario en texto libre opcional, sin recurrencia calculable ni calendario**, señalado para corrección en la puerta 2 si el humano esperaba otra cosa; **Q7** **sí**: **bloque de ministerios en la portada** con el patrón del bloque «Lo que viene» de F4, ubicado **después del bloque de eventos**. Consecuencia: la spec pasa a **v2** (6 US, 29 FR, 17 SC) y **requiere re-aprobación explícita** por el cambio de alcance. Nuevos puntos abiertos derivados: **Q8** (¿un solo medio de contacto y obligatorio si hay encargado?) y **Q9** (¿el interruptor oculta solo el contacto o también el nombre del encargado?).
Nuevos puntos abiertos derivados: **Q8** (¿un solo medio de contacto y obligatorio si hay encargado?) y **Q9** (¿el interruptor oculta solo el contacto o también el nombre del encargado?), **resueltos el mismo 2026-10-13**: **Q8** el medio de contacto es **opcional aunque haya encargado** («si se pone encargado, es opcional el correo o celular»); **Q9** el interruptor **oculta también el nombre del encargado** («Oculta también el nombre»), de modo que con la visibilidad apagada **ningún dato del encargado —ni el nombre— sale del servidor**. El humano **re-aprobó la spec v2** con esas respuestas. Además confirmó **Q6** como texto libre, el **tope de 5 ítems** en el bloque de la portada y la **re-aprobación de la v2**.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-13 — **Fase 2 (planificar): documentos escritos.** `plan.md`, `research.md`, `data-model.md`, `contracts/openapi.yaml` y `quickstart.md` por `arquitecto`; `ux.md` por `disenador-ux`, en paralelo y coherentes entre sí. Claves: una tabla `ministries` (encargado como columnas del ministerio, sin tabla de personas), migraciones **`000010_create_ministries`** y **`000011_extend_admin_actions_for_ministries`** (**sin** migración de permisos: «ministerios» ya existe), `sort_order` + `ORDER BY sort_order, id`, contrato **0.6.0 → 0.7.0**, endpoints `GET /api/v1/ministerios`, `GET /api/v1/portada/ministerios` (tope 5) y el módulo del panel tras `AdminChain("ministerios")`, **privacidad del encargado en tres capas** (DTO, consulta pública y bloque de portada) con pruebas de «0 apariciones», **0 dependencias nuevas**. Puerta 2 pendiente.
- 2026-10-13 — **Puerta 1 superada: spec v2 aprobada.** El humano resolvió Q1–Q9 (Q8: el medio de contacto es opcional aunque haya encargado; Q9: el interruptor de visibilidad oculta también el nombre), confirmó el horario como texto libre y el tope de 5 ítems en el bloque de la portada, y dijo: **«Re-apruebo la spec v2»**. La spec queda cerrada con 6 US, 29 FR y 17 SC y **sin marcas `[NECESITA ACLARACIÓN]`**. Roadmap: decisión 19. Arranca la Fase 2 (plan técnico). Commit `4e4fa0f` (v2) y el de este ajuste.
- 2026-10-13 — Inicio de la funcionalidad F6 (ministerios): rama `006-ministerios` creada desde `main` (`0797d3d`), `spec.md` desde la plantilla de Spec Kit. Fase 1 (especificar). Roadmap: F6 → **en curso**.
