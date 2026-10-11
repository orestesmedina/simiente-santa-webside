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
| Fase | **Terminada** (F3 fusionada en `main`; PR [#6](https://github.com/orestesmedina/simiente-santa-webside/pull/6) aprobado y mergeado el 2026-10-10) |
| Ciclo de corrección | 2/3 (cerrado) + 1 flujo de bug (imágenes) |
| Próximo paso | F3 cerrada y **costo congelado** (263.2 M ≈ $8.67). Siguiente: **F10 — Sistema de diseño y movimiento** (nueva spec con `equipo-feature`), recomendado antes de F4. Pendiente de despliegue a producción (aprobación humana). |
| Bloqueado por | — (Docker operativo; e2e ejecutables con `LD_LIBRARY_PATH=/tmp/opencode/pwlibs/usr/lib/x86_64-linux-gnu`) |
| Actualizado | 2026-10-09 |

## Aprobaciones

Solo se marca "aprobado" cuando el humano lo dijo explícitamente; se anota quién, cuándo y la frase.

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | aprobado | humano | 2026-10-09 | "apruebo la spec" |
| Plan | aprobado | humano | 2026-10-09 | "si" (respuesta a «¿Apruebas el plan de F3?») |
| PR / merge | aprobado | humano | 2026-10-10 | "ya aprobé el PR y ya hice merge de la rama" (PR [#6](https://github.com/orestesmedina/simiente-santa-webside/pull/6) fusionado en `main`) |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

Bloqueantes de la última validación que aún no se corrigen (ver `revision-2026-10-10-*.md`).

- **(ninguno)** — los 4 hallazgos del ciclo 2 están corregidos y re-validados; solo queda deuda menor aceptada (abajo).

**Bug detectado por el cliente tras la entrega (flujo `equipo-bug`, 2026-10-10):**
- **Imágenes de la portada no cargan.** Causa raíz: la API devuelve URLs de media **relativas** (`/api/v1/media/…`) y el frontend las pinta como `src` en su propio origen; `frontend/nginx.conf` **no proxya `/api/`**, así que `http://localhost:5173/api/v1/media/…` responde `index.html` (200 `text/html`) en vez de la imagen. Verificado: directo al backend da `200 image/jpeg`. Fix: resolver las URLs de media contra `API_BASE_URL` al renderizar (`IdentityHero`, `PublicLayout`, `PublicFooter`) + e2e que compruebe que el `<img>` carga de verdad. El «scroll del menú» y el selector es/en **no son defectos**: son decisiones de UX (D-5 y D-1) que abordará F10.
  - **Resuelto y validado (paso 4/5 ✅):** prueba de regresión `portada-imagenes.spec.ts` (roja sin el fix, verde con él), arreglo `ec2ae76` (`mediaUrl`), refuerzo del test de cabecera/pie `1d60acc` (cierra I-1 del revisor). **Los tres roles APROBARON** (0 bloqueantes): `revision-2026-10-10-bug-imagenes-{qa,codigo,seguridad}.md`. CHANGELOG `0.3.0` actualizado (`d3c1187`). Hallazgo de entorno H1 (flakiness por rate-limit de login al correr la suite e2e completa; no es defecto).

**Historial del ciclo 2 (corregidos y re-validados):**
- **B1 (BLOQUEANTE, código + QA)** — `PATCH` no vaciaba campos opcionales con `null` → `fix(portada): distinguir null de ausente` (`d5d4610`, tipo `Optional[T]`).
- **I1 (IMPORTANTE, código)** — horario en 24 h → a.m./p.m. localizado (`fc96185`).
- **A1 (MENOR, QA)** — objetivo táctil del selector 20 px → 44 px (`2b72138`).
- **N1 (MENOR, revisor)** — «m.» a las 12 → solo mediodía exacto (`db83a61`).

**Re-validación (ciclo 2): APROBADO por los tres roles** (`revision-2026-10-10-{codigo,qa,seguridad}-ciclo2.md`), 0 bloqueantes. Suites tras los fixes: backend unit+integración ✅, frontend **273** pruebas ✅, e2e **7/7** ✅.

**Deuda aceptada (no bloquea):** M-a11y (reflow a zoom 200 %, QA), S1/S2/S4/S5/S6 (seguridad: nombre de archivo de imagen, auditoría de subidas rechazadas, `nosniff` solo en media, huérfanos, fila de auditoría de `null` sin cambio) y M1/M2/M3 del revisor. Registradas para una futura funcionalidad.

## Decisiones y aclaraciones

Lo que se decidió en el chat y no está en spec.md ni plan.md (con fecha y quién decidió).

- 2026-10-09 — **Aclaraciones de F3 resueltas por el humano** (registradas en `spec.md`, sección «Aclaraciones resueltas»): identidad con nombre + lema/misión/visión + logo e imagen de portada; «quiénes somos» en texto plano con límite (1.000 caracteres, revisable); horario como lista de servicios (día, hora, nombre, lugar); WhatsApp con números y enlaces de grupo; redes con catálogo fijo y un enlace por red; contacto con ubicación + correo + teléfono; publicación por secciones/elementos; editar lo publicado se ve al guardar; edición auditada en F2; sección sin contenido publicado se oculta.
- 2026-10-09 — **Manual de Identidad del cliente** (`resources/MANUAL DE MARCA.pdf`, logo `resources/simiente.jpeg`): el cliente pidió que el diseño del sitio siga su manual de marca. Datos de marca (misión, visión, valores, personalidad, voz, paleta `#1a2b4a`/#00c9a7/#ffffff/#F5F2EC/#ff6b3d/#217638, tipografías Bebas Neue/Poppins/Playfair Display y reglas del logotipo) recogidos en `spec.md` como referencia para plan y UX.
- 2026-10-10 — **Rediseño como F10 (decisión del humano).** Tras revisar la portada, el cliente pidió un sitio más moderno (2026) con animaciones, sin el scroll del menú y con el selector de idioma compacto. Se decide NO reabrir F3: el rediseño va como **F10 «Sistema de diseño y movimiento»** (transversal, antes de F4), y los defectos de F3 (imágenes) se corrigen por `equipo-bug`. Registrado en `roadmap.md` (F10 + decisión 11).
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
- 2026-10-10 — **Validación (Fase 7) y corrección (ciclo 2) cerradas**: ciclo 1 → B1 (bloqueante: `PATCH` `null`), I1 (importante) y A1 (menor) corregidos y **re-validados con APROBADO de los tres roles**; N1 (menor) también corregido. Fase 8 (convergencia): 40/40 tareas, 0 pendientes → **Converged**. Arranca la Fase 9 (entrega).
- 2026-10-10 — **Fase 9 (entrega) preparada**: `documentador` actualizó `CHANGELOG.md` (0.3.0), `README.md`, `quickstart.md` y creó `docs/entrega/F3-portada-e-informacion-general.md`; `devops` verificó la entrega (sin deriva, compose/`.env.example` ok, `.github/` intacto) y redactó `pr.md`. El push y la apertura del PR quedan pendientes (este entorno no tiene `gh` ni credenciales de GitHub). Roadmap: F3 → **en revisión**.
- 2026-10-10 — **Bug de imágenes (equipo-bug) corregido y validado**: `qa-tester` lo reprodujo (`b257e0b`), `dev-frontend` lo arregló (`ec2ae76`, `mediaUrl`), se cerró el importante I-1 (`1d60acc`) y los tres roles aprobaron; CHANGELOG actualizado. El rediseño moderno queda como **F10** en el roadmap (pendiente de arrancar).
- 2026-10-10 — **F3 TERMINADA**: el humano aprobó y fusionó el PR [#6](https://github.com/orestesmedina/simiente-santa-webside/pull/6) en `main`. **Costo cerrado** (`make costos CERRAR=1`): 263.2 M tokens ≈ **$8.67**. Roadmap: F3 → *terminada*. Siguiente: **F10 — Sistema de diseño y movimiento**.
