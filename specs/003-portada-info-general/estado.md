# Estado: 003-portada-info-general

<!--
Lo mantiene el ORQUESTADOR (no los subagentes). Se copia a specs/NNN-nombre/estado.md
justo después de crear la spec, y se actualiza en cada cambio de fase, en cada puerta
de aprobación, en cada ciclo de corrección y al cerrar la sesión.
`make estado` lee la tabla "Resumen": conserva los nombres de sus campos.
-->

## Resumen

| Campo | Valor |
|---|---|
| Rama | 003-portada-info-general |
| Flujo | equipo-feature |
| Fase | 4/9 · Implementar |
| Ciclo de corrección | 0/3 |
| Próximo paso | Implementar tareas de `tasks.md` por capa (empezando por `[db]`: T303–T308); un commit por tarea |
| Bloqueado por | — (nada) |
| Actualizado | 2026-10-09 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano | 2026-10-09 | "apruebo la spec" |
| Plan | aprobado | humano | 2026-10-09 | "si" (respuesta a «¿Apruebas el plan de F3?») |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-<fecha>.md`).

- (ninguno)

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-09 — **Aclaraciones de F3 resueltas por el humano** (registradas en `spec.md`, sección «Aclaraciones resueltas»): identidad con nombre + lema/misión/visión + logo e imagen de portada; «quiénes somos» en texto plano con límite (1.000 caracteres, revisable); horario como lista de servicios (día, hora, nombre, lugar); WhatsApp con números y enlaces de grupo; redes con catálogo fijo y un enlace por red; contacto con ubicación + correo + teléfono; publicación por secciones/elementos; editar lo publicado se ve al guardar; edición auditada en F2; sección sin contenido publicado se oculta.
- 2026-10-09 — **Manual de Identidad del cliente** (`resources/MANUAL DE MARCA.pdf`, logo `resources/simiente.jpeg`): el cliente pidió que el diseño del sitio siga su manual de marca. Datos de marca (misión, visión, valores, personalidad, voz, paleta `#1a2b4a`/#00c9a7/#ffffff/#F5F2EC/#ff6b3d/#217638, tipografías Bebas Neue/Poppins/Playfair Display y reglas del logotipo) recogidos en `spec.md` como referencia para plan y UX.
- 2026-10-09 — **Ajustes del plan (decisiones del humano, puerta 2):** (a) **idioma con `localStorage`** — se recuerda la elección entre visitas y una primera visita sin preferencia se muestra en español; se actualiza **FR-010** de la spec (y US4/Assumptions/Q11) y se unifica con el plan y `ux.md`; (b) la portada pública pasa a `/` y la página «Estado del sistema» de F1 se mueve a **`/health`** (no `/estado`); (c) publicación/retiro **por sección, cada una por separado** (identidad, quiénes somos y contacto son cada uno un elemento publicable propio), coherente con FR-013.

## Bitácora

Una línea por sesión o hito, la más reciente arriba.

- 2026-10-09 — Inicio de F3: rama `003-portada-info-general` y spec redactada (19 FR, 13 SC), tras resolver 10 aclaraciones con el humano y leer su Manual de Identidad.
- 2026-10-09 — **Puerta 1 superada**: el humano aprobó la spec («apruebo la spec»). Arranca Fase 2 (plan + UX).
- 2026-10-09 — Fase 2 completada: `plan.md`, `research.md`, `data-model.md`, `contracts/openapi.yaml`, `quickstart.md` y `ux.md`. Pendiente la puerta 2 (aprobación del plan).
- 2026-10-09 — **Puerta 2 superada**: el humano aprobó el plan («si») y pidió versionar los recursos de marca (`resources/`) en el repo. Arranca Fase 3 (tareas + coherencia).
- 2026-10-09 — **Recursos de marca versionados**: `resources/MANUAL DE MARCA.pdf` y `resources/simiente.jpeg` entran al repo; el logo se publicará además en `frontend/public/brand/` durante la implementación.
- 2026-10-09 — Fase 3 completada: `tasks.md` (40 tareas T301–T340) y `analyze.md`. El primer `analyze` **rechazó** con 4 críticos (C1 agregado admin sin tarea; C2 horario; C3 memoria de idioma; C4 descarga de imágenes y borradores); se corrigieron con arquitecto y UX y el re-análisis dio **APROBADO** (0 críticos), cerrando además N1–N4.
