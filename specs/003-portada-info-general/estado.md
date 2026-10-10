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
| Fase | 7/9 · Validar (implementación cerrada, 40/40 tareas) |
| Ciclo de corrección | 2/3 (ciclo 1: 4 defectos de F4; ciclo 2: B1 bloqueante de validación + I1/A1) |
| Próximo paso | Validación en paralelo con `qa-tester`, `revisor-codigo` y `seguridad` → `revision-2026-10-09-*.md` |
| Bloqueado por | — (Docker operativo; e2e ejecutables con `LD_LIBRARY_PATH=/tmp/opencode/pwlibs/usr/lib/x86_64-linux-gnu`) |
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

Bloqueantes de la última validación que aún no se corrigen (ver `revision-2026-10-10-*.md`).

- **B1 (BLOQUEANTE, código + QA)** — Los `PATCH` del panel no vacían campos opcionales con `null`: en Go `null` y campo ausente son indistinguibles, y el frontend envía `null` para quitar `endTime`/una traducción. Se pierde en silencio (la UI dice «Guardado» y el valor reaparece); `{"endTime":null}` responde `400`. Reproducido en vivo por QA. **→ `dev-backend` (ciclo 2).**
- **I1 (IMPORTANTE, código)** — `formatTimeRange` muestra el horario en 24 h, pero `ux.md` §4.6/D-3 promete presentación a.m./p.m. localizada. **→ `dev-frontend` (ciclo 2).**
- **A1 (MENOR, QA)** — Objetivo de toque del selector de idioma 20 px (< 24 px recomendado). **→ `dev-frontend` (ciclo 2).**
- **M-a11y (MENOR, QA)** — Reflow fino a zoom 200 % (56 px de desborde, sin pérdida de contenido). Aceptado como deuda.
- Seguridad: **APROBADO**, 0 bloqueantes (5 menores S1–S5, incluida la semántica `null` vista desde seguridad ≡ B1).

**Correcciones del ciclo 2 (aplicadas, pendientes de re-validación):**
- **B1** ✅ `fix(portada): distinguir null de ausente en los PATCH` (`d5d4610`; tipo `Optional[T]`; las 3 pruebas de regresión de QA en verde).
- **I1** ✅ `fix(frontend): horario en a.m./p.m. localizado` (`fc96185`).
- **A1** ✅ `fix(frontend): objetivo táctil del selector de idioma ≥44 px` (`2b72138`).
- Verificado tras el fix: backend unit+integración ✅, frontend **271** pruebas ✅, e2e **6/6** ✅ (incluye `portada-accesibilidad.spec.ts`).

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
- 2026-10-09 — **Fase 4 (Implementar) en curso**: T301–T316 hechas y comprometidas; HEAD `f6ca31d`. La verificación de integración de T316 (`go test -tags=integration ./internal/portada/`) quedó **sin confirmar** (el entorno Docker local se cayó). Fase corregida de `4/9` a `6/9` (numeración canónica de las 9 fases) al retomar.
- 2026-10-09 — **Retomar**: el stack Docker local (backend, frontend, `db`, redis) había quedado detenido; `make estado` marcaba `tasks.md` más nuevo que `estado.md`. Estado actualizado y roadmap corregido (F3 sin comillas en la columna Rama).
- 2026-10-09 — **Fase 4 avanzada**: T301–T336 implementadas y comprometidas (contrato 0.4.0, migraciones `000005`/`000006`, sqlc, `platform/*`, repositorio y servicios de `portada` con 84.7 % de cobertura, handlers y rutas con el permiso `portada` cableado, y todo el frontend: marca/i18n/portada pública/panel de información). Backend y frontend en verde. Faltan T337–T338 (e2e) y T339 (cierre).
- 2026-10-09 — **T337/T338 (e2e) escritas pero sin ejecutar**: `frontend/e2e/portada-publica.spec.ts` y `portada-panel.spec.ts` (commits `844f570`, `04be06c`), sin marcar `[X]`. **Bloqueo de entorno**: Docker Desktop no tiene la integración WSL activa (`docker` no encuentra el daemon), y el Chromium de Playwright carece de librerías de sistema. Verificado lo posible sin Docker: `gofmt`/`go vet`/`go test ./...` (unit) ✅, frontend lint/typecheck/268 pruebas ✅, `make security` ✅ (0 vulnerabilidades).
- 2026-10-09 — **Docker activado por el humano y validación de F3 ejecutada**: `make up` + migraciones (`000001`–`000006`) sobre base limpia; **integración backend ✅** (`go test -tags=integration ./...`), **5 e2e ✅** (acceso, auditoría, portada-pública, portada-panel, estado). Al ejecutar por primera vez los e2e se encontraron y **corrigieron 3 defectos** (bucle de corrección 1): (a) permiso de escritura del volumen `uploads_data` para el usuario no-root → `500` en la subida (fix `a93a233`); (b) el selector de idioma compartía `name` entre cabecera y pie (radios sin marcar) → `useId()`; (c) la `nav` generaba scroll horizontal a 320 px → `min-w-0`. Además se actualizó el e2e de estado de F1 a la ruta `/health`. T337/T338 marcadas `[X]`.
- 2026-10-09 — **Fase 4 cerrada (40/40 tareas)**: T339 (cierre) en verde — `make ci` EXIT=0 (lint + unit + integración + 269 pruebas frontend + security), 5 e2e, sin deriva (`sqlc`/`api:gen`), cobertura `service*.go` 84.7 %. El cierre destapó 5 hallazgos de lint en pruebas del backend (corregidos en `dda8120`, bucle de corrección 1). Pendientes manuales: SC-010 (usabilidad, humano+QA) y checklist WCAG de SC-008 (QA). Informe: `cierre.md`.
