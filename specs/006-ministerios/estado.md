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
| Fase | 7/9 · Entregar — implementación (T601–T637), validación y convergencia **completas**; **T638 hecha** (docs y borrador del PR). **PR pendiente de abrir** |
| Ciclo de corrección | 1/3 (cerrado: los tres roles aprobaron el 2026-10-13) |
| Próximo paso | El humano abre el Pull Request con `specs/006-ministerios/pr.md` (`gh` no está disponible en este entorno, igual que en F4). Cuando **apruebe el PR**: `make costos CERRAR=1` + commit de `estado.md`/`costos.json` en la rama; al confirmar el **merge**: roadmap → `terminada`. Queda manual SC-012 (usabilidad) y SC-008 (lector de pantalla) |
| Bloqueado por | — |
| Actualizado | 2026-10-13 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | **aprobado** | humano | 2026-10-13 | v1: «Apruebo la spec» respondiendo las aclaraciones Q1–Q7 (la v2 amplió el alcance y se pidió re-aprobación, como en F4). v2: **«Re-apruebo la spec v2»**, respondiendo las aclaraciones nuevas Q8 (el medio de contacto es opcional aunque haya encargado) y Q9 (el interruptor de visibilidad oculta también el nombre del encargado) |
| Plan | **aprobado** | humano | 2026-10-13 | **«Apruebo el plan»**, con dos decisiones propias en la puerta 2: el interruptor de la ficha del encargado nace **«Oculto por defecto»** y `/ministerios` sin publicados **«Redirigir a la portada»** |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-2026-10-13.md`).

- **(ninguno)** — el único hallazgo importante (**I1**, contrato de `whatsappUrl`) y la mejora **M1** (`sortOrder` acotado a `int32`) se corrigieron en el **ciclo 1** (`1e0d9e5`, `6328cff`, `2d75367`) y los **tres roles los re-validaron como cerrados**.
- **Deuda aceptada (no bloquea el merge)**: M2 (`RecordAction` en `service.go`, precedente de F4) · M3 (`mergeInvalid` conserva el último motivo) · M4 (`setTimeout` en una prueba de portada) · M5 (el listado del panel se corta en 100 sin aviso de total) · M6 (promover `Optional[T]` a `platform/` con el cuarto consumidor) · S1 (redacción en la prosa del contrato) · S2 (los 3 fragmentos del snapshot de F6 superados por el ciclo 1: registrar, no editar) · B1 de seguridad (`imageFile` sin comprobar existencia en el almacén; imagen rota, no fuga) · B2 (contraseñas de prueba en e2e) · informativos I1–I3 de seguridad.
- **Manual (no automatizable)**: **SC-012** (usabilidad con ≥ 6 personas y el cronómetro de los 30 s de SC-001), con protocolo en `pruebas-usabilidad-SC-012.md`, y **SC-008** (lector de pantalla real y pantalla táctil física).
- **Ajeno a F6**: `frontend/e2e/qa-bugs-004.spec.ts` («Bug 1», scroll a `#whatsapp` desde `/eventos`) es **flaky preexistente de F3/F4** (~25 %). **Pendiente de decisión del humano**: abrir el flujo `equipo-bug` para F4 o dejarlo registrado.

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-13 — **Arranque de F6 (decisión del humano en sesión).** El humano pidió «continuemos con el f6» en lugar de F5 (el siguiente pendiente del roadmap). F4, F5, F6 y F7 son independientes entre sí según el roadmap §3, así que se arranca F6 — Ministerios por el flujo `equipo-feature`. Numeración de la rama: **`006-ministerios`**, alineada con el identificador de roadmap (precedente: F10 → `010-sistema-diseno-movimiento`), dejando el número 005 libre para F5 (grupos de conexión).
- 2026-10-13 — **Aclaraciones de F6 resueltas por el humano (Q1–Q7)**, en su mayoría **distintas de la opción recomendada**, con **ampliación de alcance** en tres puntos: **Q1** encargado **opcional** con nombre y un medio de contacto (celular o correo) y un **interruptor para decidir si el contacto se muestra**; **Q2** enlace de WhatsApp **opcional**; **Q3** se acepta **cualquier `https` válido y seguro** (sin restringir a dominios de WhatsApp); **Q4** imagen **opcional** cargada por el equipo; **Q5** orden **manual con una columna numérica** (precedente verificado: `sort_order` + `ORDER BY sort_order, id` en F3/F4); **Q6** **se añaden horario y lugar**, con el horario como **texto libre opcional sin recurrencia** (varía por ministerio y a veces no es un día fijo); **Q7** **sí**: **bloque de ministerios en la portada** con el patrón de F4, ubicado **después del bloque de eventos**. La spec pasó a **v2** (6 US, 29 FR, 17 SC).
- 2026-10-13 — **Q8 y Q9 resueltas**: **Q8** el medio de contacto es **opcional aunque haya encargado** («si se pone encargado, es opcional el correo o celular»); **Q9** el interruptor **oculta también el nombre del encargado** («Oculta también el nombre»). El humano **re-aprobó la spec v2**, confirmó **Q6** como texto libre y el **tope de 5 ítems** del bloque de la portada.
- 2026-10-13 — **Decisiones de la puerta 2**: interruptor **«Oculto por defecto»** (privacidad por defecto) y `/ministerios` sin publicados **«Redirigir a la portada»**; después, que la redirección sea **silenciosa** y que **no** se añada anuncio accesible para lectores de pantalla (queda como deuda anotada en el plan).

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-13 — **Fase 7 (entregar) — documentación lista y PR pendiente de abrir.** `documentador` (T638): `CHANGELOG.md` **`[0.6.0]`**, `README.md`, notas de entrega `docs/entrega/F6-ministerios.md` y el borrador `specs/006-ministerios/pr.md`. **`gh` no está disponible** en el entorno, así que el PR lo abre el humano con ese borrador (igual que en F4). Commit del orquestador por el trabajo del documentador.
- 2026-10-13 — **Fases 5 y 6 completas: validación APROBADA y convergencia.** Validación en paralelo: `qa-tester` **APROBADO**, `revisor-codigo` **APROBADO CON CAMBIOS MENORES** (1 importante I1 + 6 menores), `seguridad` **APROBADO** (2 bajos, 4 informativos). Consolidado en `revision-2026-10-13.md`. **Ciclo de corrección 1** por **I1** (contrato de `whatsappUrl`: ahora admite `null` y `MinisterioPatch` exige `minProperties: 1`) y **M1** (`sortOrder` acotado a `int32`): corregido en `1e0d9e5`, `6328cff` y `2d75367`, y **re-validado como cerrado por los tres roles**. Convergencia: las 38 tareas `[X]`, sin tareas nuevas → **Converged**.
- 2026-10-13 — **Fase 4 (implementar) completa.** **T601–T637**: migraciones `000010`/`000011` (sin migración de permisos: «ministerios» ya existe, catálogo en 10), contrato vivo **0.7.0** sin `$ref` colgantes, paquete `backend/internal/ministerios/` completo (repositorio, servicios y handlers) con **privacidad del encargado en tres capas**, `media.IsFilePublished` ampliado sin regresión, frontend completo (`/ministerios`, navegación, bloque de portada, panel), **e2e 19/19** y verificación de cierre de `devops`: `make ci` verde, cobertura `service*.go` **96,83 %**, migraciones `up→down→up`, `sqlc`/contrato sin deriva, humo del despliegue y **sin cambios de infraestructura**.
- 2026-10-13 — **Fase 3 (tareas y coherencia) completa.** `tasks.md` con **38 tareas** (T601–T638: 3 `[db]`, 17 `[backend]`, 16 `[frontend]`, 2 `[infra]`) y el `analyze` del `revisor-codigo`: **COHERENTE CON OBSERVACIONES**, **0 críticos**, 5 IMPORTANTES (A6-1 navegación desde el documento del bloque, A6-2 `$ref` `Limit`/`Offset` en la fusión del contrato, A6-3 el `PATCH` fusiona campos presentes, A6-4 lectura de SC-017 en el bloque, A6-5 T628 depende de T624) y 8 menores. **Todos corregidos** y verificados por el orquestador; ningún hallazgo abierto.
- 2026-10-13 — **Puerta 2 superada: plan aprobado** con **dos decisiones propias del humano**: (a) el interruptor de la ficha del encargado nace **«Oculto por defecto»** (privacidad por defecto: publicar es un acto expreso; `leader_visible BOOLEAN NOT NULL DEFAULT FALSE`) y (b) **`/ministerios` sin publicados redirige a la portada**, **distinto a propósito** del aviso sin redirigir de F4 en `/eventos` (no unificar). Aplicado en `spec.md`, `plan.md` (P6-21 y RG6-13), `research.md`, `data-model.md`, `contracts/openapi.yaml`, `quickstart.md` y `ux.md` (D-13, D-14). Puntos señalados por el arquitecto y **no** implementados por no estar en la spec: anuncio `aria-live` al redirigir (deuda anotada).
- 2026-10-13 — **Fase 2 (planificar): documentos escritos.** `plan.md`, `research.md`, `data-model.md`, `contracts/openapi.yaml` y `quickstart.md` por `arquitecto`; `ux.md` por `disenador-ux`. Claves: una tabla `ministries` (encargado como columnas del ministerio), migraciones **`000010`** y **`000011`** (**sin** migración de permisos), `sort_order` + `ORDER BY sort_order, id`, contrato **0.6.0 → 0.7.0**, endpoints públicos y de panel tras `AdminChain("ministerios")`, privacidad en tres capas, **0 dependencias nuevas**.
- 2026-10-13 — **Puerta 1 superada: spec v2 aprobada.** El humano resolvió Q1–Q9, confirmó el horario como texto libre y el tope de 5 ítems, y dijo: **«Re-apruebo la spec v2»**. La spec queda cerrada con 6 US, 29 FR y 17 SC y **sin marcas `[NECESITA ACLARACIÓN]`**. Roadmap: decisión 19.
- 2026-10-13 — Inicio de la funcionalidad F6 (ministerios): rama `006-ministerios` creada desde `main` (`0797d3d`), `spec.md` desde la plantilla de Spec Kit. Fase 1 (especificar). Roadmap: F6 → **en curso**.
