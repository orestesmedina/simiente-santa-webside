# Tasks: Portada e información general (F3)

**Branch**: `003-portada-info-general` | **Date**: 2026-10-09 | **Spec**: [spec.md](./spec.md) (APROBADA, US1–US5 / FR-001…FR-019) · **Plan**: [plan.md](./plan.md) (P3-1…P3-19) · **UX**: [ux.md](./ux.md)

**Input**: plan aprobado (`plan.md`, P3-1…P3-19) · `research.md` (R3-1…R3-18) · `data-model.md`
(migraciones `000005`/`000006`, tablas `home_*`, ampliación de `admin_actions`) ·
`contracts/openapi.yaml` (delta de diseño, **inmutable**) · `ux.md` (pantallas, componentes, textos,
estados) · `quickstart.md` (§0…§12, verificación) · constitución (`.specify/memory/constitution.md`)
· skills `go-backend`, `postgres-db`, `react-frontend` · F1 y F2 ya integradas
(`specs/001-estructura-base/tasks.md` T001…, `specs/002-acceso-gestion-usuarios/tasks.md` T201…T254).

> **Regla de ejecución**: **un commit por tarea** (Conventional Commits; los hooks del kit lo
> validan). Toda tarea de código **incluye sus pruebas** en el mismo commit (constitución §III):
> backend con `go test` y, donde toca persistencia real, `go test -tags=integration`
> (`//go:build integration`, PostgreSQL real levantado con `testcontainers-go` — R19 de F2);
> frontend con Vitest + Testing Library + MSW; flujos críticos con Playwright (`make e2e`). La
> cobertura exigida es **≥ 80 % en `internal/portada/service*.go`** (`go test -cover`). Quien escribe
> no aprueba: cada tarea pasa por `qa-tester`, `revisor-codigo` y `seguridad` antes de integrarse
> (flujo del orquestador). **El CI del kit (`.github/workflows/ci.yml`) no se toca**.

## Decisiones del plan que estas tareas respetan (sin reabrir)

- **P3-2/R3-1**: contenido bilingüe por columnas `*_es` (obligatorio) / `*_en` (opcional); el
  fallback `en → es` lo resuelve el service público por campo (nunca campo vacío, FR-009).
- **P3-3/R3-2**: `GET /api/v1/portada?lang=es|en` devuelve los contenidos ya localizados
  (`Cache-Control: no-store`); el registro de `error.code` **no crece** (una subida inválida es
  `400 invalid` con `details`).
- **P3-4/R3-3**: `publication_state` **por elemento** —también cada singleton: identidad, «quiénes
  somos» y contacto publican y retiran **por separado**— con consultas públicas `…Published`
  separadas; las secciones sin elementos publicados se **omiten**; se publica al guardar (FR-014).
- **P3-5/R3-4**: singletons con `singleton BOOLEAN` + `UNIQUE` y upsert transaccional con
  `pg_advisory_xact_lock`.
- **P3-6/P3-7/P3-8**: horario estructurado (`day_of_week` 0–6 con **selector localizado**, nunca
  texto libre; `start_time` "HH:MM"; **`end_time` opcional** para rangos, `analyze` C2); WhatsApp
  `kind=direct|group` con destino normalizado y duplicado exacto → `409`; redes con catálogo fijo en
  el `CHECK`, `UNIQUE (network)` y hosts oficiales validados.
- **P3-9/R3-8**: imágenes en **disco local** (`UPLOAD_DIR`, volumen `uploads_data`), firma binaria
  (JPEG/PNG/WebP; **sin SVG ni GIF**), nombre generado `img_<uuid>.<ext>`, ≤ 8 MB
  (`UPLOAD_MAX_BYTES`), descarga con `nosniff`/`inline` y **`Cache-Control: no-store`**, que **solo**
  sirve archivos referenciados por contenido **publicado** (`404` en otro caso; `analyze` C4/M6) y
  cuya subida se audita con `home.image.upload` (`analyze` I8).
- **P3-10/R3-9**: i18n propio con diccionarios **tipados** es/en y memoria en **`localStorage`**
  (primera visita sin preferencia → español; **sin** auto-detección del navegador; FR-010 ajustado
  por el humano el 2026-10-09).
- **P3-11/R3-10**: se **activa** el permiso `portada` ya sembrado por F2 (catálogo sin cambios):
  `AdminChain("portada")` en el backend y `RequirePermission`/menú filtrado en el frontend.
- **P3-12/R3-11**: toda mutación registra `admin_actions` **en su misma transacción** (fail-closed)
  con los 15 códigos `home.*` (incl. `home.image.upload`) y `target_kind='content'` (migración
  `000006`); regla exacta de códigos (`analyze` M5): el **alta** usa `home.*.create` aunque nazca
  publicada, `home.publish`/`home.unpublish` **solo** registran cambios de estado y un `PATCH` que
  cambia datos y estado deja **dos filas**; las denegaciones y rechazos se registran best-effort
  resolviendo la acción en el dominio `portada`.
- **P3-14**: el delta de `contracts/openapi.yaml` se fusiona **aditivamente** en
  `backend/api/openapi.yaml` (`info.version` 0.3.0 → **0.4.0**) con los dos cambios documentados del
  `AdminActionItem` (enum `action` + `targetKind: content`); el snapshot de `specs/` no se edita.
- **P3-15/R3-14**: etiqueta `url` nueva en `platform/validate`; límites del service («quiénes somos»
  1.000 caracteres; ≤ 50 servicios; ≤ 20 canales; 1 enlace por red).
- **P3-16/R3-15**: `/` es la portada pública y la pantalla «Estado del sistema» de F1 se mueve a
  **`/health`** (ruta de la SPA; el endpoint `/healthz` **no** se toca).
- **P3-18/R3-18**: el *chrome* público (cabecera/pie) usa la identidad publicada y, si no hay,
  fallback estático (nombre del diccionario + logo versionado en el repo) — nunca datos en borrador.

## Alineación de nomenclatura (decidida por el orquestador/`ux.md`; no se reabre)

1. **Rutas SPA**: el módulo del panel es **`/panel/informacion`** (como `ux.md` §3; el plan lo llamaba
   `/panel/portada`) y la pantalla de estado es **`/health`** (decisión humana del 2026-10-09).
   `/` queda para la portada pública.
2. **Carpetas de features**: **`frontend/src/features/publico/`** y
   **`frontend/src/features/informacion/`** (nomenclatura de `ux.md` §6; el plan las llamaba
   `features/public`/`features/portada`). Los nombres de **componentes y hooks** son los de `ux.md`
   §6.2/§6.3, con nombres de código en inglés.
3. **i18n** vive en **`frontend/src/features/publico/i18n/`** (`LanguageProvider`, `useLanguage`,
   diccionarios `es`/`en`) con su `messages.ts`, como `ux.md` §6.2; lo reutilizarán F4–F9.
4. **Rutas de API**: **no cambian** — son las del contrato (`/api/v1/portada`,
   `/api/v1/media/{fileName}`, `/api/v1/admin/portada/*`); el snapshot `contracts/openapi.yaml` es
   inmutable (P3-14).
5. **Componentes compartidos de F2**: solo dos extensiones menores —`Button` con variante `accent`
   y `StatusPill` con `draft`/`published` («Borrador»/«Publicado»), `ux.md` §6.1—; `Pagination` **no**
   se usa en F3 (listados cortos). Todo lo demás reutiliza el inventario único de F2 sin duplicar
   markup (lo revisa `revisor-codigo`).
6. **Numeración T301…T340**: patrón por funcionalidad de F1 (T0xx) y F2 (T2xx); los IDs `T001…` ya
   están usados por F1.

## Decisiones del `analyze` (2026-10-09) que estas tareas aplican

Cierran los hallazgos de `analyze.md` sin reabrir la spec; los que afectan a `ux.md` quedan aquí como
decisión **para que `disenador-ux` alinee su documento**:

| Hallazgo | Decisión aplicada |
|---|---|
| **C1** | Nueva tarea **T340** (service + handler + pruebas de `GET /api/v1/admin/portada`, el agregado del panel) |
| **C2** | Horario **estructurado**: `dayOfWeek` 0–6 en **selector localizado** (nunca texto libre), `startTime` "HH:MM" y **`endTime` "HH:MM" opcional** (rango «10:00 a. m. − 12:00 m.»); el formato a.m./p.m. es presentación. `ux.md` §4.6/D-3 se alinea a esto |
| **C3** | La memoria de idioma es `localStorage`: **solo la primera visita sin preferencia entra en español**; una re-visita con preferencia **la respeta**. `ux.md` §2.2.4/§8.2 se corrigen y T328/T337 lo prueban |
| **C4** | `GET /api/v1/media/{fileName}` **solo** sirve archivos referenciados por contenido **publicado** (`404` en otro caso) con `Cache-Control: no-store` (RG3-8) |
| **I2** | **Sin fallback de idioma en el cliente** para el sitio público (el servidor resuelve `en → es`); si hace falta un helper para pares `{es, en}`, es solo para los formularios del panel y vive en `features/informacion/`. `ux.md` retira `textContentFor` del sitio público |
| **I3** | **No se añade** regla de servicios duplicados (la spec solo define duplicados de WhatsApp): `ux.md` §4.6 retira el texto "Duplicado exacto: igual día+hora+nombre+lugar → rechazado" |
| **I4** | La frase del pie es **cadena de interfaz i18n** (`footer.welcome`), **no** contenido editable: `ux.md` retira "editable en panel". No se añade campo |
| **I5** | Nomenclatura única de tokens: **la de `ux.md` §1** (`navy`, `navy-soft`, `teal`, `teal-strong`, `white`, `cream`, `coral`, `leaf`; `--font-display`/`--font-sans`/`--font-emotiva`) |
| **I6** | Todo campo `*En`: `""`/espacios → **`NULL`** (trim en el service); documentado en contrato y quickstart y probado en T317/T319/T320/T334–T336 |
| **I7** | El ancho mínimo probado es **320 px** (el que promete `ux.md`): e2e y checklist en 320/768/1280. **Sin** dependencia de axe (§II): la accesibilidad se valida con la checklist manual de QA (SC-008); un barrido automatizado queda como ampliación solo con aprobación humana |
| **I8** | La subida de imagen **se audita**: código nuevo `home.image.upload` en `000006` (aún no aplicada) + registro fail-closed en T323 |
| **M4** | El diccionario i18n añade la clave `brand.name` (nombre de marca del *chrome* de fallback, R3-18); `ux.md` §7.2 la incluye en su catálogo |

## Leyenda

| Marca | Significado |
|---|---|
| `[backend]` `[frontend]` `[db]` `[infra]` | Capa y subagente responsable (`[backend]`/`[db]` → `dev-backend`; `[frontend]` → `dev-frontend`; `[infra]` → `devops`) |
| `[P1]`…`[P7]` | **Paralelizable** solo dentro de su grupo (archivos distintos y sin dependencia previa entre ellas; ver "Grupos de paralelismo") |
| *(sin `[P]`)* | Secuencial: tiene dependencias previas que deben estar integradas |
| *(FR-…, US…)* | Requisitos de `spec.md` que cubre la tarea |

> **Orden de capas**: primero el contrato (Fase 1), luego `[db]`, luego `platform/`, y dentro del
> dominio `portada` **repository → service → handler/routes** (Fases 5–7): cada capa se construye y
> prueba con lo que necesita la siguiente (el service con fakes, el handler con service falso),
> manteniendo la dirección de dependencias **`handler → service → repository`** de §II. La
> **auditoría** se escribe junto a cada mutación (Fases 5–6) porque FR-017 exige que edición y
> registro sean atómicos: una tarea de "registro" al final obligaría a reabrir tareas cerradas.

## Tabla resumen

| ID | Tarea | Capa | Fase | Paralelo | Depende de |
|---|---|---|---|---|---|
| T301 | Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.4.0) | `[backend]` | 1 | P1 | — |
| T302 | Dependencias npm nuevas (@fontsource ×3) | `[frontend]` | 1 | P1 | — |
| T303 | Regenerar `frontend/src/api/schema.d.ts` desde el contrato fusionado | `[frontend]` | 1 | — | T301, T302 |
| T304 | `docker-compose.yml`: volumen `uploads_data` + variables `UPLOAD_*` del backend | `[infra]` | 2 | P2 | — |
| T305 | `.env.example`: `UPLOAD_DIR` y `UPLOAD_MAX_BYTES` | `[infra]` | 2 | P2 | — |
| T306 | Migración `000005_create_home_content` (6 tablas, up/down) | `[db]` | 3 | — | — |
| T307 | Migración `000006_extend_admin_actions_for_home_content` (up/down) | `[db]` | 3 | — | T306 |
| T308 | Consultas sqlc `home.sql` + generar y commitear `internal/db` | `[db]` | 3 | — | T306, T307 |
| T309 | `platform/storage`: interfaz `Store` + `LocalStore` seguro | `[backend]` | 4 | P3 | — |
| T310 | `platform/validate`: etiqueta `url` (https + host) | `[backend]` | 4 | P3 | — |
| T311 | `platform/audit`: +15 códigos `home.*` y `TargetKindContent` | `[backend]` | 4 | P3 | — |
| T312 | `platform/config`: `UPLOAD_DIR`, `UPLOAD_MAX_BYTES` | `[backend]` | 4 | P3 | — |
| T313 | `portada`: `model.go` + `repository.go` (mapeo pgtype, `withTx`, auditoría sobre el tx) | `[backend]` | 5 | — | T308, T311 |
| T314 | `portada`: repository de singletons (upsert + advisory lock) + integración | `[backend]` | 5 | P4 | T313 |
| T315 | `portada`: repository de listas (horario/WhatsApp/redes, `…Published`) + integración | `[backend]` | 5 | P4 | T313 |
| T316 | Integración: migración `000006` up→down→up y auditoría con objetivo `content` atómica | `[backend]` | 5 | — | T314, T315 |
| T317 | `portada`: `service.go` (tipos comunes, invariantes y límites) | `[backend]` | 6 | — | T313 |
| T318 | `portada`: `service_public.go` — portada localizada con fallback y solo publicado | `[backend]` | 6 | — | T317 |
| T319 | `portada`: `service_admin.go` — identidad, quiénes somos y contacto (con auditoría) | `[backend]` | 6 | — | T317, T314 |
| T320 | `portada`: `service_admin.go` — horario, WhatsApp y redes (con auditoría) | `[backend]` | 6 | — | T317, T315 |
| T321 | `portada`: `service_audit.go` — denegaciones y rechazos (`audit.Recorder`) | `[backend]` | 6 | — | T317 |
| T322 | `portada`: `handler_public.go` — `GET /api/v1/portada` | `[backend]` | 7 | — | T318 |
| T323 | `portada`: `handler_images.go` + `handler_media.go` — subida y descarga de imágenes | `[backend]` | 7 | — | T309, T312, T322 |
| T324 | `portada`: `handler_admin.go` — identidad, quiénes somos y contacto | `[backend]` | 7 | — | T319, T322 |
| T325 | `portada`: `handler_admin.go` — horario, WhatsApp y redes | `[backend]` | 7 | — | T320, T322 |
| T340 | `portada`: agregado del panel `GET /api/v1/admin/portada` (service + handler + pruebas) *(analyze C1)* | `[backend]` | 7 | — | T314, T315, T319, T320, T322 |
| T326 | `portada`: `routes.go` + `cmd/api/main.go` (permiso `portada` cableado) + humo | `[backend]` | 7 | — | T321, T322, T323, T324, T325, T340 |
| T327 | Marca: tokens del Manual, tipografías `@fontsource`, logo en `public/brand/`, `BrandLogo` | `[frontend]` | 8 | P5 | T302 |
| T328 | `features/publico/i18n`: `LanguageProvider`, diccionarios tipados es/en, `localStorage` | `[frontend]` | 8 | P5 | — |
| T329 | Componentes compartidos: `Button` (`accent`) y `StatusPill` (`draft`/`published`) | `[frontend]` | 8 | P5 | — |
| T330 | `api/portada.ts`: cliente público y de panel + `uploadImage` | `[frontend]` | 8 | — | T303 |
| T331 | Permisos y rutas: `PORTADA`, router (`/`, `/health`, `/panel/informacion`), menú del panel | `[frontend]` | 8 | — | T330 |
| T332 | `features/publico`: `PublicLayout`, `HomePage` y secciones de la portada | `[frontend]` | 8 | — | T327, T328, T329, T330, T331 |
| T333 | `features/informacion`: `InformationPage` (pestañas de las 6 piezas) + `messages.ts` | `[frontend]` | 8 | — | T331 |
| T334 | `features/informacion`: `IdentityForm` + `ImageUploader` + `PublishControls` | `[frontend]` | 8 | — | T333 |
| T335 | `features/informacion`: `WhoWeAreForm` + `ContactForm` (publicar/retirar por sección) | `[frontend]` | 8 | — | T334 |
| T336 | `features/informacion`: horario, WhatsApp y redes (listas + formularios + publicar/retirar) | `[frontend]` | 8 | — | T334 |
| T337 | E2E Playwright `portada-publica.spec.ts` (US1, US3 público, US4, US5) | `[frontend]` | 9 | P7 | T332 |
| T338 | E2E Playwright `portada-panel.spec.ts` (US2, US3 panel, permisos) | `[frontend]` | 9 | P7 | T334, T335, T336 |
| T339 | Verificación de cierre: `make ci`, `make e2e`, `quickstart.md` §0–§12 y mapa SC | `[infra]` | 9 | — | todas |

**Total: 40 tareas** — 20 `[backend]` · 14 `[frontend]` · 3 `[db]` · 3 `[infra]` · 15 marcadas `[P]`.

## Grupos de paralelismo

| Grupo | Tareas | Por qué pueden ir en paralelo |
|---|---|---|
| P1 | T301 · T302 | `backend/api/openapi.yaml` y `frontend/package.json` son archivos distintos sin dependencias |
| P2 | T304 · T305 | `docker-compose.yml` y `.env.example` son archivos distintos |
| P3 | T309 · T310 · T311 · T312 | Paquetes `platform` distintos (`storage`, `validate`, `audit`, `config`) sin dependencias entre sí |
| P4 | T314 · T315 | `repository_home.go` se divide por responsabilidad (singletons vs. listas): archivos/funciones distintas que parten de T313 |
| P5 | T327 · T328 · T329 | `features/publico/` (marca), `features/publico/i18n/` y `frontend/src/components/` no se solapan |
| P7 | T337 · T338 | Dos specs e2e independientes (`e2e/portada-publica.spec.ts` y `e2e/portada-panel.spec.ts`) |

---

## Fase 1 — Contrato OpenAPI, dependencias y tipos

- [ ] T301 · Fusionar el delta OpenAPI en `backend/api/openapi.yaml` (0.4.0) · `[backend]` `[P1]`

  - **Archivos**: `backend/api/openapi.yaml` (EDITADO). *No* se toca
    `specs/003-portada-info-general/contracts/openapi.yaml` (snapshot inmutable, P3-14).
  - **Qué hace**: funde **aditivamente** el delta aprobado en el contrato vivo: las **16**
    operaciones nuevas (`GET /api/v1/portada`, `GET /api/v1/media/{fileName}`,
    `GET /api/v1/admin/portada` y las 13 mutaciones del panel) y sus schemas (`PortadaPublica`,
    `PortadaAdmin`, `Identity*`, `About*`, `Contact*`, `ScheduleItem*`, `WhatsappChannel*`,
    `SocialLink*`, `ImageUploadResult`, `PublicationState`, parámetro `PortadaItemId`), reutilizando
    `ErrorEnvelope`/`ErrorBody`, las respuestas reutilizables y `sessionCookie`. Aplica además los
    **dos cambios documentados del delta** en `AdminActionItem`: enum `action` **+ 15 códigos
    `home.*`** y enum `targetKind` **+ `content`** (con su prosa: `targetId` en `null` y `targetLabel`
    identificando el elemento). Sube `info.version` a **0.4.0**. El registro de `error.code` **no
    cambia** (P3-14): las operaciones públicas declaran `security: []` y las del panel `sessionCookie`.
    *Cubre el contrato de todos los FR*; garantiza **FR-013** (ninguna operación pública devuelve
    borradores) y **FR-017** (los códigos `home.*` en el registro cerrado).
  - **Pruebas incluidas** (§III): — (artefacto de contrato). **Cómo se verifica**: el YAML parsea sin
    errores; las rutas públicas (`/api/v1/portada`, `/api/v1/media/{fileName}`) tienen `security: []`
    y las de `/api/v1/admin/portada` `sessionCookie`; ningún schema público contiene
    `publicationState` ni campos `*En` sin resolver (FR-013/FR-009);
    `grep -n "version:" backend/api/openapi.yaml` → `0.4.0`; `make api-gen` completa sin errores.
  - **Criterio de terminado**: el contrato vivo contiene F1 + F2 + F3 y es la única fuente de tipos
    del frontend (T303).
  - **Commit sugerido**: `docs(api): fusionar delta de F3 en el contrato vivo (0.4.0)`

- [ ] T302 · Dependencias npm nuevas (tipografías del Manual) · `[frontend]` `[P1]`

  - **Archivos**: `frontend/package.json`, `frontend/package-lock.json` (EDITADOS).
  - **Qué hace**: añade **solo** `@fontsource/bebas-neue`, `@fontsource/poppins` y
    `@fontsource/playfair-display` (R3-12/R3-16: tipografías del Manual de Identidad, autoalojadas,
    sin CDN de terceros). Nada más: no se instala ninguna otra librería (sin `react-i18next`, sin
    procesadores de imagen: P3-10/P3-9).
  - **Pruebas incluidas** (§III): — (manifiesto). **Cómo se verifica**:
    `cd frontend && npm ci && npm run lint && npm run typecheck && npm test -- --run && npm run build`
    en verde; `npm audit` sin altas ni críticas (§IV); `go list` no aplica (sin dependencias Go
    nuevas: R3-16).
  - **Criterio de terminado**: solo las tres dependencias nuevas (justificadas en `research.md` R3-16);
    ningún archivo del kit modificado.
  - **Commit sugerido**: `chore(frontend): agregar tipografías del Manual (@fontsource)`

- [ ] T303 · Regenerar `frontend/src/api/schema.d.ts` · `[frontend]`

  - **Archivos**: `frontend/src/api/schema.d.ts` (GENERADO y commiteado).
  - **Qué hace**: ejecuta `make api-gen` (o `npm run api:gen`) contra **`backend/api/openapi.yaml`**
    (nunca contra el snapshot de `specs/`, P3-14) y commitea los tipos de F3 (`PortadaPublica`,
    `PortadaAdmin`, `IdentityInput`/`IdentityAdmin`, `ScheduleItem*`, `WhatsappChannel*`,
    `SocialLink*`, `ImageUploadResult`, `PublicationState`, `AdminActionItem` con los enums ampliados).
  - **Pruebas incluidas** (§III): — (artefacto generado). **Cómo se verifica**: regenerar no produce
    `git diff`; `npm run typecheck` pasa con los tipos nuevos.
  - **Criterio de terminado**: `schema.d.ts` reproducible desde el contrato y listo para
    `api/portada.ts` (T330; regla anti-deriva de §8.1.5).
  - **Commit sugerido**: `build(frontend): regenerar schema.d.ts con los tipos de F3`

---

## Fase 2 — Infraestructura local (`[infra]`, la aplica `devops`)

- [ ] T304 · `docker-compose.yml`: volumen de imágenes y variables del backend · `[infra]` `[P2]`

  - **Archivos**: `docker-compose.yml` (EDITABLE: no está en `.kit-manifest.json`). *No se toca*
    `.github/workflows/ci.yml` (archivo del kit).
  - **Qué hace**: añade el **volumen con nombre `uploads_data:/var/lib/simiente/uploads`** al
    servicio `backend` (los archivos subidos sobreviven a `down`/rebuild como `pgdata`, R3-8/RG3-1) y
    las variables `UPLOAD_DIR` (en el contenedor `/var/lib/simiente/uploads`) y `UPLOAD_MAX_BYTES`
    (`8388608`) al servicio `backend`. El resto de servicios no cambia (`db`, `redis`, `frontend`).
    *Soporta FR-002/FR-011* (logo e imagen de portada) y §VII (`make up` levanta todo).
  - **Pruebas incluidas** (§III): — (orquestación). **Cómo se verifica**: `docker compose config`
    declara el volumen y las variables; `make up` levanta los cuatro servicios; un
    `docker compose down && make up` **conserva** los archivos de `/var/lib/simiente/uploads`;
    `curl -i http://localhost:8080/healthz` sigue respondiendo como en F1 (§8.1.9).
  - **Criterio de terminado**: los archivos subidos persisten entre reinicios del contenedor; ningún
    archivo del kit modificado.
  - **Commit sugerido**: `build(compose): volumen uploads_data y variables de imágenes`

- [ ] T305 · `.env.example`: variables de imágenes · `[infra]` `[P2]`

  - **Archivos**: `.env.example` (EDITABLE).
  - **Qué hace**: documenta `UPLOAD_DIR` (por defecto `./uploads` fuera de Docker) y
    `UPLOAD_MAX_BYTES` (`8388608` = 8 MB) con su consumidor (`platform/config`, T312) y su uso en
    `quickstart.md` §0/§6. Sin secretos (§IV): no hay credenciales nuevas en F3.
  - **Pruebas incluidas** (§III): — (documento de configuración). **Cómo se verifica**: ninguna
    variable sin consumidor y ningún consumidor sin su variable; `cp .env.example .env` es usable tal
    cual; lo audita `seguridad`.
  - **Criterio de terminado**: la lista documentada coincide exactamente con la que lee
    `platform/config` (T312) y con la que inyecta compose (T304).
  - **Commit sugerido**: `chore(env): documentar UPLOAD_DIR y UPLOAD_MAX_BYTES`

---

## Fase 3 — Migraciones y capa de datos (`[db]`)

> Convenciones de la skill `postgres-db` y de `data-model.md` (fuente de verdad): `id UUID`,
> `created_at`/`updated_at`, índices y `CHECK`/`UNIQUE` **en la base**, `up`/`down` completos, y
> **nunca** se edita una migración aplicada. Numeración sin huecos desde `000001` (F2 terminó en
> `000004`).

- [ ] T306 · Migración `000005_create_home_content` · `[db]`

  - **Archivos**: `backend/migrations/000005_create_home_content.up.sql` y
    `000005_create_home_content.down.sql` (NUEVOS).
  - **Qué hace**: crea las **6 tablas de contenido** tal como `data-model.md`: `home_identity`,
    `home_about`, `home_contact` (singletons con `singleton BOOLEAN CHECK (singleton)` +
    `UNIQUE (singleton)`, R3-4) y `home_services`, `home_whatsapp_channels`, `home_social_links`
    (colecciones). Patrón bilingüe `*_es NOT NULL` / `*_en NULL` (FR-008, R3-1);
    `publication_state` **en cada tabla** (FR-013: cada singleton publica por separado);
    `CHECK` de límites (quiénes somos ≤ 1.000, FR-003), de `day_of_week`/`start_time`/`end_time`
    (patrón "HH:MM" y `end_time > start_time`; R3-5/`analyze` C2), de
    alt obligatorio con imagen (FR-019), `UNIQUE (network)` (FR-006/Q5) y
    `UNIQUE (kind, destination, name_es)` (FR-005, duplicado exacto). Índices
    `(sort_order, id)` en las colecciones ordenables. `down` elimina las 6 tablas en orden inverso.
    *Cubre FR-001…FR-008 y FR-013* a nivel de datos.
  - **Pruebas incluidas** (§III): — (SQL). **Cómo se verifica**: `make db-migrate` aplica sin
    errores sobre la BD de F2 (`000004` → `000005`); `migrate … down` revierte sin errores; los
    `CHECK`/`UNIQUE` de `data-model.md` están literalmente en el SQL; ninguna tabla sin
    `id`/`created_at`/`updated_at`.
  - **Criterio de terminado**: el esquema coincide con `data-model.md` (lo revisa `revisor-codigo`
    contra la skill `postgres-db`).
  - **Commit sugerido**: `feat(db): tablas de contenido de la portada (000005)`

- [ ] T307 · Migración `000006_extend_admin_actions_for_home_content` · `[db]`

  - **Archivos**: `backend/migrations/000006_extend_admin_actions_for_home_content.up.sql` y
    `.down.sql` (NUEVOS).
  - **Qué hace**: amplía el registro de auditoría de F2 (FR-017, R3-11): extiende el `CHECK` de
    `admin_actions.action` con los **15 códigos `home.*`** (incl. `home.image.upload`, `analyze` I8),
    el de `target_kind` con **`content`** y
    sustituye el CHECK de coherencia de objetivo por `admin_actions_target_coherence_check` (para
    `content`: ambas FK en `NULL` y `target_label` obligatorio). El `down` restaura los CHECK
    originales (destructivo a propósito: elimina las filas `target_kind='content'`, como el down de
    `000004` tira sus tablas; documentado en `data-model.md`). *Cubre FR-017* a nivel de datos.
  - **Pruebas incluidas** (§III): — (SQL; su verificación ejecutable es T316). **Cómo se verifica**:
    `make db-migrate` aplica; los nombres de restricción que referencia (`admin_actions_action_check`,
    `admin_actions_target_kind_check`, `admin_actions_check1`) son los que creó `000004`; el registro
    de códigos queda cerrado en la BD.
  - **Criterio de terminado**: la BD garantiza el registro cerrado de acciones y de tipos de objetivo
    ampliado a F3, sin tocar ninguna migración anterior (§VI).
  - **Commit sugerido**: `feat(db): ampliar admin_actions a la portada (000006)`

- [ ] T308 · Consultas sqlc `home.sql` + generar · `[db]`

  - **Archivos**: `backend/internal/db/queries/home.sql` (NUEVO) y `backend/internal/db/*.go`
    (GENERADOS y commiteados: `make sqlc`).
  - **Qué hace**: escribe las consultas nombradas de `data-model.md`: `UpsertHomeIdentity`,
    `GetHomeIdentity`, `GetHomeIdentityPublished` (e ídem `HomeAbout`/`HomeContact`),
    `Insert/Update/Delete/Get/ListHomeServices` + `ListHomeServicesPublished` (ídem
    `HomeWhatsappChannels` y `HomeSocialLinks`, con `ORDER BY sort_order, id` / `network`) e
    `IsHomeFilePublished` (descarga de imágenes: `TRUE` solo si el nombre está referenciado por
    contenido **publicado** — `analyze` C4 —; es la **única definición** de esta consulta y la
    consumen T314/T318/T323; F4–F9 la amplían con sus tablas).
    **Sin** `SELECT *`, listas con `ORDER BY` determinista (§8.1.7) y **las variantes `…Published`
    como única vía pública** (P3-4). La auditoría **reutiliza** `InsertAdminAction` de
    `queries/audit.sql` (no se duplica, R3-11). *Cubre FR-001…FR-007, FR-013 y el camino de archivos
    de FR-013* a nivel de acceso a datos.
  - **Pruebas incluidas** (§III): — (SQL anotado; su ejecución es T314–T316). **Cómo se verifica**:
    `make sqlc` genera sin errores y `make sqlc-verify` no detecta deriva; los tipos generados
    (`internal/db`) quedan commiteados; ninguna consulta de escritura sobre `admin_actions` más allá
    de la reutilizada (FR-025 de F2).
  - **Criterio de terminado**: la capa generada compila y es la única fuente de SQL del dominio
    `portada` (T313–T315).
  - **Commit sugerido**: `feat(db): consultas sqlc del contenido de la portada`

---

## Fase 4 — Núcleo de plataforma (`backend/internal/platform/`)

- [ ] T309 · `platform/storage`: interfaz `Store` + `LocalStore` seguro · `[backend]` `[P3]`

  - **Archivos**: `backend/internal/platform/storage/storage.go`,
    `local.go`, `storage_test.go`, `local_test.go` (NUEVOS).
  - **Qué hace**: plumbing de archivos (R3-8): interfaz `Store` con `Save`/`Open`/`Delete` e
    implementación `LocalStore` sobre `os` en `UPLOAD_DIR`. **Seguridad**: el nombre lo genera el
    servidor (`img_<uuid>.<ext>`), `Open` solo acepta nombres que cumplen
    `^img_[0-9a-f-]{36}\.(jpg|png|webp)$` (rechaza `../`, absolutos y nombres de cliente → sin path
    traversal), `Delete` es idempotente y los errores van envueltos con `%w` (R7). Sin SQL, sin HTTP,
    sin estado global (R6: el directorio entra por constructor). *Soporta FR-002/FR-011/FR-019 y el
    edge case "archivo inválido o carga fallida"*.
  - **Pruebas incluidas** (§III): unitarias con `t.TempDir()`. **Cómo se verifica**:
    `go test ./internal/platform/storage/` en verde: `Save` genera nombres únicos y cumple la regex;
    `Open` rechaza `../../etc/passwd`, `img_abc.jpg` (nombre de cliente) y rutas absolutas;
    `Delete` de un archivo inexistente no falla; los errores son identificables (`ErrNotFound`).
  - **Criterio de terminado**: ningún camino del servidor depende de un nombre controlado por el
    usuario (lo audita `seguridad`).
  - **Commit sugerido**: `feat(storage): almacén local de imágenes con nombres validados`

- [ ] T310 · `platform/validate`: etiqueta `url` · `[backend]` `[P3]`

  - **Archivos**: `backend/internal/platform/validate/validate.go` y `validate_test.go` (EDITADOS).
  - **Qué hace**: añade la etiqueta `url` al validador de F2 (R3-14): acepta solo esquema `https`
    con host no vacío (vía `net/url`) y máximo 500 caracteres; rechaza `http://`, `javascript:`,
    `data:` y cadenas sin host. *Cubre FR-015* (enlaces de WhatsApp y de redes con formato válido).
    Las demás etiquetas (`required`, `min`, `max`, `email`, `oneof`, `phone`) no cambian.
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/platform/validate/` en verde con los casos anteriores y con que un DTO
    válido pasa; el `details` por campo se mantiene.
  - **Criterio de terminado**: `url` es reutilizable por F4–F9 igual que `email`/`phone`; sin
    dependencias nuevas (§II).
  - **Commit sugerido**: `feat(validate): etiqueta url (https + host)`

- [ ] T311 · `platform/audit`: códigos `home.*` y objetivo `content` · `[backend]` `[P3]`

  - **Archivos**: `backend/internal/platform/audit/audit.go` y `audit_test.go` (EDITADOS).
  - **Qué hace**: amplía el registro cerrado del plumbing de F2 (R3-11.1): los **15 códigos**
    `home.identity.update`, `home.about.update`, `home.contact.update`,
    `home.schedule.create|update|delete`, `home.whatsapp.create|update|delete`,
    `home.social.create|update|delete`, `home.image.upload` (subida de imágenes, `analyze` I8),
    `home.publish`, `home.unpublish` (en `ActionCodes` y
    `ValidActionCode`) y el `TargetKindContent = "content"` con su validación en `Action.Validate()`
    (ambas FK en nil y `TargetLabel` obligatorio). *Cubre FR-017*.
  - **Pruebas incluidas** (§III): unitarias (tabla de casos). **Cómo se verifica**:
    `go test ./internal/platform/audit/` en verde: cada código `home.*` pasa `Validate`; un objetivo
    `content` con una FK rellena o sin `TargetLabel` falla; los códigos y tipos de F2 siguen
    intactos (regresión).
  - **Criterio de terminado**: la lista cerrada de código coincide literalmente con el `CHECK` de la
    migración `000006` (T307) — lo comprueba la integración de T316.
  - **Commit sugerido**: `feat(audit): códigos home.* y objetivo content para F3`

- [ ] T312 · `platform/config`: variables de imágenes · `[backend]` `[P3]`

  - **Archivos**: `backend/internal/platform/config/config.go` y `config_test.go` (EDITADOS).
  - **Qué hace**: lee `UPLOAD_DIR` (por defecto `./uploads`) y `UPLOAD_MAX_BYTES` (por defecto
    `8388608`) con la misma disciplina de F2 (valores por defecto usables en local, validación al
    arrancar, mensaje claro si un valor es inválido). *§VII (config por variables de entorno)*.
  - **Pruebas incluidas** (§III): unitarias. **Cómo se verifica**:
    `go test ./internal/platform/config/` en verde: valores por defecto, sobrescritura por
    entorno y rechazo de `UPLOAD_MAX_BYTES` no numérico o ≤ 0.
  - **Criterio de terminado**: ninguna constante de configuración vive en el código (R6/§VII) y la
    lista coincide con `.env.example` (T305).
  - **Commit sugerido**: `feat(config): UPLOAD_DIR y UPLOAD_MAX_BYTES`

---

## Fase 5 — Dominio `portada`: modelo y repository

- [ ] T313 · `portada`: `model.go` + `repository.go` · `[backend]`

  - **Archivos**: `backend/internal/portada/model.go`, `repository.go`,
    `repository_test.go` (NUEVOS).
  - **Qué hace**: entidades del dominio (`Identity`, `About`, `Contact`, `Service`,
    `WhatsappChannel`, `SocialLink` + `PublicationState`), DTOs con etiquetas `validate` y sufijos
    `Es`/`En` (R3-1), y la base del repository: constructor sobre `*pgxpool.Pool`, mapeos
    `pgtype` → dominio (§8.1.6), `withTx` (patrón de F2) y la **inserción de `admin_actions` dentro
    del `tx`** vía la consulta compartida `InsertAdminAction` de `internal/db` (mapeo
    `audit.Action` → `InsertAdminActionParams`; R3-11: sin que un dominio importe a otro, R2).
    *Soporta FR-008 (patrón bilingüe) y FR-017 (registro atómico)*.
  - **Pruebas incluidas** (§III): unitarias de los mapeos (`pgtype` ↔ dominio, nil/`Valid`) y del
    mapeo de acciones. **Cómo se verifica**: `go test ./internal/portada/` en verde; `gofmt`/`go vet`
    limpios; ningún `pgtype` sale de `repository.go` (R4).
  - **Criterio de terminado**: el paquete compila con sus tipos y el tx de escritura queda listo para
    las mutaciones de T314/T315.
  - **Commit sugerido**: `feat(portada): modelo y base del repository con auditoría transaccional`

- [ ] T314 · `portada`: repository de singletons + integración · `[backend]` `[P4]`

  - **Archivos**: `backend/internal/portada/repository_home.go` (singletons),
    `repository_home_integration_test.go` (NUEVOS; `//go:build integration`).
  - **Qué hace**: `UpsertHomeIdentity`/`UpsertHomeAbout`/`UpsertHomeContact` (upsert
    `ON CONFLICT (singleton) DO UPDATE` **en transacción con `pg_advisory_xact_lock`**, R3-4), sus
    lecturas `Get…`/`Get…Published` e `IsHomeFilePublished` (la consulta de T308; nombre de archivo →
    `TRUE` solo si una
    identidad **publicada** lo referencia; `analyze` C4). *Cubre FR-002, FR-003, FR-007 y FR-013 (los
    singletons publican por separado y la descarga de imágenes respeta el estado)*.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real** (helpers de
    `platform/testutil` de F2). **Cómo se verifica**:
    `go test -tags=integration ./internal/portada/` en verde: dos `Upsert` concurrentes dejan **una**
    fila con la última escritura completa; `Get…Published` nunca devuelve la fila en `draft`; los
    `CHECK` de longitud y de alt-con-imagen se disparan con datos inválidos; el upsert actualiza
    `updated_at`; `IsHomeFilePublished` es `FALSE` con la identidad en borrador, `TRUE` al publicarla y
    otra vez `FALSE` al retirarla (C4/SC-002).
  - **Criterio de terminado**: singletons garantizados por la BD (una fila) y publicación por
    sección verificada (FR-013).
  - **Commit sugerido**: `feat(portada): singletons con upsert serializado y lecturas publicadas`

- [ ] T315 · `portada`: repository de listas + integración · `[backend]` `[P4]`

  - **Archivos**: `backend/internal/portada/repository_home.go` (listas; EDITADO),
    `repository_home_integration_test.go` (AMPLIADO).
  - **Qué hace**: CRUD de `home_services`, `home_whatsapp_channels` y `home_social_links`
    (`Insert/Update/Delete/Get/List` + `List…Published`) con el `ORDER BY` del contrato
    (`sort_order, id` / `network`) y las traducciones de restricción a `apperr` (`conflict` para
    `UNIQUE`). *Cubre FR-004, FR-005, FR-006 y FR-013*.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/portada/` en verde: `UNIQUE (network)` → error traducido a
    `conflict`; `UNIQUE (kind, destination, name_es)` solo en el duplicado exacto; `CHECK` de
    `publication_state`, `day_of_week` y `start_time`/`end_time`; `List…Published` sin filas `draft`; el orden
    es estable entre páginas (aunque F3 no pagina).
  - **Criterio de terminado**: las tres colecciones quedan cubiertas por persistencia real con sus
    invariantes en la BD.
  - **Commit sugerido**: `feat(portada): persistencia de horario, WhatsApp y redes`

- [ ] T316 · Integración: migración `000006` y auditoría `content` atómica · `[backend]`

  - **Archivos**: `backend/internal/portada/repository_audit_integration_test.go` (NUEVO) y/o
    `repository_migrations_integration_test.go` (NUEVO).
  - **Qué hace**: la verificación ejecutable de `data-model.md` §"Validación del modelo": (1) la
    migración `000006` aplica **up → down → up** sin errores (evidencia de RG3-2, incluida la
    restauración de los CHECK de `admin_actions`); (2) una fila de `admin_actions` con
    `target_kind='content'`, ambas FK en `NULL` y `target_label` se inserta y el `CHECK` de
    coherencia rechaza una FK rellena o un `target_label` ausente; (3) una mutación de contenido y su
    registro son **atómicos**: al forzar un fallo en la inserción del registro, la mutación **no se
    aplica** (edge case "no aplicarse sin registro", FR-017). *Cubre FR-017 y FR-013*.
  - **Pruebas incluidas** (§III): **integración contra PostgreSQL real**. **Cómo se verifica**:
    `go test -tags=integration ./internal/portada/` en verde con los tres casos; los 15 códigos
    `home.*` del `CHECK` de la BD coinciden con `platform/audit` (T311).
  - **Criterio de terminado**: FR-017 verificado en la base (éxito atómico; nada aplicado sin
    registro) y la migración reversible.
  - **Commit sugerido**: `test(portada): integración de migración 000006 y auditoría content`

---

## Fase 6 — Dominio `portada`: servicios

- [ ] T317 · `portada`: `service.go` — tipos comunes, invariantes y límites · `[backend]`

  - **Archivos**: `backend/internal/portada/service.go`, `service_test.go` (NUEVOS).
  - **Qué hace**: interfaz `Repository` (definida por quien la consume, R3), constantes de negocio
    revisables («quiénes somos» **1.000** caracteres; máx. **50** servicios y **20** canales;
    catálogo de redes y hosts oficiales por red — R3-7/R3-14), normalización (trim + colapso de
    espacios; teléfono a dígitos con `+`; destino de grupo tal cual; **`""`/espacios → `NULL` en todo
    campo `*_en`**, `analyze` I6 — nunca `''`, que los `CHECK` rechazan) e invariantes compartidas
    ("es obligatorio, en opcional"; `publicationState` válido). *Soporta FR-003 y FR-015*.
  - **Pruebas incluidas** (§III): unitarias (tabla de casos de normalización y de límites).
    **Cómo se verifica**: `go test ./internal/portada/` en verde; las constantes son el único punto
    de verdad de los límites (los usan T319/T320 y los cita el contrato); `""`, `"  "` y `"\t"` en un
    `*En` salen como `NULL` y `" x "` como `"x"`.
  - **Criterio de terminado**: los servicios de T318–T320 solo expresan reglas, sin constantes
    dispersas.
  - **Commit sugerido**: `feat(portada): invariantes y constantes del dominio`

- [ ] T318 · `portada`: `service_public.go` — portada localizada · `[backend]`

  - **Archivos**: `backend/internal/portada/service_public.go`, `service_public_test.go` (NUEVOS).
  - **Qué hace**: `GetPortada(lang)` (FR-001/FR-008/FR-009): resuelve cada campo traducible con
    fallback **`en` si existe, si no `es`** (nunca vacío, SC-006; el cliente **no** reimplementa el
    fallback — `analyze` I2 —), lee **solo** las variantes
    `…Published` (P3-4: ninguna vía de acceso público a borradores, SC-002), **omite por completo**
    las secciones sin elementos publicados (Q10/SC-012) y arma los enlaces públicos (WhatsApp →
    `https://wa.me/<dígitos>` para `direct`, URL del grupo para `group`; SC-009). `lang` fuera de
    `es|en` → `apperr.Invalid`. Ordena las colecciones según el contrato. Además la **política de
    descarga de imágenes** *(analyze C4)*: `OpenMedia(fileName)` valida el nombre (`400 invalid` si
    no cumple el patrón, `404 not_found` si es válido pero no existe o **no está referenciado por
    contenido publicado** — `IsHomeFilePublished`, la consulta de T308 expuesta por el repository en
    T314, `analyze` M6 —) y abre el archivo del `storage.Store`.
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos de fallback por campo y de la
    política de descarga). **Cómo se verifica**: `go test ./internal/portada/` en verde: contenido
    solo-en-es se muestra en
    español pidiendo `lang=en` (jamás hueco); con todo en borrador la respuesta solo lleva `lang`;
    una sección mixta publica solo lo publicado; `lang=fr` → `invalid`; el orden es el esperado;
    la descarga devuelve `ErrNotFound` con la identidad en borrador o retirada y el archivo con la
    identidad publicada (C4); un nombre fuera del patrón → `ErrInvalid` (M6).
  - **Criterio de terminado**: SC-002 y SC-006 verificables en el service; cobertura del service
    medida con `go test -cover` (≥ 80 % al cierre).
  - **Commit sugerido**: `feat(portada): portada pública localizada con fallback en→es`

- [ ] T319 · `portada`: `service_admin.go` — identidad, quiénes somos y contacto · `[backend]`

  - **Archivos**: `backend/internal/portada/service_admin.go`, `service_admin_test.go` (NUEVOS).
  - **Qué hace**: `SaveIdentity`/`SaveAbout`/`SaveContact` (FR-002/FR-003/FR-007/FR-011): reemplazo
    completo del singleton con validación FR-015 (obligatorios, es obligatorio, 1.000 caracteres,
    correo con `email`, teléfono con `phone`, `url` para los campos de imagen referencia), **alt
    obligatorio en español cuando hay imagen** (FR-019), `publicationState` por sección (FR-013:
    cada singleton publica **por separado**), publicación al guardar (FR-014) y **registro
    transaccional** `home.identity.update`/`home.about.update`/`home.contact.update` —más una fila
    adicional `home.publish`/`home.unpublish` **si cambia el estado** (`analyze` M5: dos filas, de
    modo que `home.publish`/`home.unpublish` solo significan cambio de estado)— (FR-017,
    fail-closed). Al reemplazar una imagen la anterior se borra (best-effort con log, R3-8).
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/portada/` en verde: cada campo inválido → `400` con `details` y **sin
    guardar nada** (FR-015); guardar sobre un singleton publicado deja los cambios visibles
    (FR-014); retirar una sección no afecta a las demás; cada guardado deja su fila de auditoría en
    la misma transacción (el fake del repository lo observa) —con **dos filas** si cambia el estado y
    `home.publish`/`home.unpublish` **solo** en ese caso (M5)— y si el registro falla no se guarda
    (edge case); `""`/espacios en un campo `*En` se guarda como `NULL` y el español vacío se rechaza
    (I6/FR-015); texto con tildes/`ñ`/`¿¡`/emojis intacto.
  - **Criterio de terminado**: FR-002/FR-003/FR-007/FR-011/FR-013/FR-014/FR-015/FR-017 cubiertos en
    el service.
  - **Commit sugerido**: `feat(portada): guardado de identidad, quiénes somos y contacto`

- [ ] T320 · `portada`: `service_admin.go` — horario, WhatsApp y redes · `[backend]`

  - **Archivos**: `backend/internal/portada/service_admin.go` (EDITADO) y `service_admin_test.go`.
  - **Qué hace**: alta/edición/borrado de servicios (día 0–6 válidos —**selector localizado, nunca
    texto libre**—, `start_time` "HH:MM" y **`end_time` opcional posterior al inicio** para rangos
    «10:00–12:00», `analyze` C2; nombre y lugar
    en español obligatorios), canales (nombre en español, `kind` coherente con el destino:
    teléfono con `phone` si `direct`; URL https de `chat.whatsapp.com`/`wa.me` si `group`; **duplicado
    exacto → `409`**, R3-6) y redes (red del catálogo → `400 details.network`; host oficial de la red
    → `400 details.url`; **segundo enlace para la misma red → `409`**, R3-7). Límites de colección
    (≤50 servicios, ≤20 canales), `publicationState` por elemento (FR-013), publicación al guardar
    (FR-014) y registro transaccional `home.schedule.*`/`home.whatsapp.*`/`home.social.*` con la
    regla exacta de códigos (`analyze` M5): el **alta** siempre `home.*.create` aunque nazca
    publicada, `home.publish`/`home.unpublish` **solo** para cambios de estado (y una edición que
    cambia datos y estado deja **dos filas**) (FR-017).
  - **Pruebas incluidas** (§III): unitarias con fakes (tabla de casos). **Cómo se verifica**:
    `go test ./internal/portada/` en verde: cada rechazo con su `details` y sin guardar nada
    (FR-015: nada de cambios parciales); `endTime` ≤ `startTime` o mal formado → `400 details.endTime`;
    duplicados → `conflict`; **sin regla de duplicados de servicios** (no está en la spec: `analyze`
    I3 — ninguna prueba la exige); borrado físico del elemento con su
    `targetLabel` conservado en la auditoría; publicar/retirar un elemento no toca los demás; un alta
    publicada registra `home.*.create` (no `home.publish`) y un cambio de estado registra solo
    `home.publish`/`home.unpublish` (M5); `""`/espacios en `*En` → `NULL` (I6);
    cobertura del service sostenida (≥ 80 %).
  - **Criterio de terminado**: FR-004/FR-005/FR-006/FR-011/FR-013/FR-014/FR-015/FR-017 cubiertos en
    el service.
  - **Commit sugerido**: `feat(portada): gestión de horario, WhatsApp y redes con validación`

- [ ] T321 · `portada`: `service_audit.go` — denegaciones y rechazos · `[backend]`

  - **Archivos**: `backend/internal/portada/service_audit.go`, `service_audit_test.go` (NUEVOS).
  - **Qué hace**: implementa `audit.Recorder` para el dominio (R3-11.4): `RecordDenied` resuelve
    **método+ruta → acción `home.*`** de las rutas de F3 (tabla del dominio, como `actionFromRoute`
    de F2 pero solo con `/api/v1/admin/portada/...`, incluida la subida de imágenes →
    `home.image.upload`) y la persiste con `result='denied'`
    best-effort; el helper de rechazo de DTO inválido hace lo propio con `result='failure'`. Nunca
    transporta el cuerpo de la petición (FR-026 de F2). *Cubre FR-017 (y el patrón de denegación de
    FR-012)*.
  - **Pruebas incluidas** (§III): unitarias con fakes. **Cómo se verifica**:
    `go test ./internal/portada/` en verde: cada ruta del módulo resuelve a su acción y
    `targetLabel` (`Portada · <Sección> · <elemento>`); una ruta desconocida no deja fila; un fallo
    de registro no cambia la respuesta (best-effort, R23 de F2).
  - **Criterio de terminado**: las denegaciones de permiso sobre las rutas de F3 quedan registradas
    igual que las de F2 (SC-013).
  - **Commit sugerido**: `feat(portada): registro de denegaciones y rechazos del módulo`

---

## Fase 7 — Dominio `portada`: handlers, rutas y cableado

- [ ] T322 · `portada`: `handler_public.go` — `GET /api/v1/portada` · `[backend]`

  - **Archivos**: `backend/internal/portada/handler.go`, `handler_public.go`,
    `handler_public_test.go` (NUEVOS).
  - **Qué hace**: publica la operación del contrato: valida `lang` (`es|en`, por defecto `es`),
    llama a `GetPortada` (T318) y responde `200` con `PortadaPublica` **y cabecera
    `Cache-Control: no-store`** (SC-003). Sin autenticación (FR-001). El DTO público **nunca**
    contiene `publicationState` ni campos sin resolver (FR-013/FR-009). Helpers HTTP del dominio
    (decodificar/validar/responder con `apperr` + `WriteError`, patrón de F2).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/portada/` en verde: `200` con el sobre correcto y `no-store`; `lang=fr` →
    `400 invalid` con `details.lang`; secciones ausentes cuando el service no las devuelve (SC-012);
    error del service → sobre `ErrorEnvelope` genérico.
  - **Criterio de terminado**: `quickstart.md` §1 es ejecutable tal cual (US1).
  - **Commit sugerido**: `feat(portada): endpoint público de la portada`

- [ ] T323 · `portada`: `handler_images.go` + `handler_media.go` — imágenes · `[backend]`

  - **Archivos**: `backend/internal/portada/handler_images.go`, `handler_media.go`,
    `handler_images_test.go`, `handler_media_test.go` (NUEVOS).
  - **Qué hace**: `POST /api/v1/admin/portada/imagenes` (multipart del contrato, R3-8): tope con
    `http.MaxBytesReader` (`UPLOAD_MAX_BYTES`), tipo por **firma binaria** (`http.DetectContentType`;
    solo JPEG/PNG/WebP — **sin SVG ni GIF**), nombre generado `img_<uuid>.<ext>`, guarda con
    `storage.Store` (T309), **registra `home.image.upload`** (`targetLabel` =
    `Portada · Imagen · <fileName>`, `analyze` I8; fail-closed: si el registro falla, se aborta y se
    elimina el archivo) y responde `201` con `ImageUploadResult` (la subida **no** cambia contenido
    visible). Y `GET /api/v1/media/{fileName}` (público): valida el nombre por regex antes
    de tocar el disco y **solo sirve archivos referenciados por contenido publicado**
    (`OpenMedia` de T318 sobre `IsHomeFilePublished` — escrita en T308 y expuesta por el repository en
    T314 —; `analyze` C4) con
    `X-Content-Type-Options: nosniff`, `Content-Disposition: inline` y `Cache-Control: no-store`.
    **Criterio de códigos (M6)**: nombre que **no cumple el patrón** → `400 invalid`; nombre válido
    pero **inexistente o no publicado** → `404 not_found`. *Cubre
    FR-002, FR-011, FR-013 (camino de archivos) y FR-019 + el edge case "archivo inválido o carga
    fallida"*.
  - **Pruebas incluidas** (§III): `httptest` con `Store` falso y `t.TempDir()` para la subida real.
    **Cómo se verifica**: `go test ./internal/portada/` en verde: sin archivo o campo erróneo → `400
    invalid` con `details.file`; archivo de más de `UPLOAD_MAX_BYTES` → `400` con `details.fileSize`;
    un `.svg` o un HTML renombrado → `400` (la firma manda); cada subida deja su fila
    `home.image.upload` y un fallo de registro aborta sin dejar archivo (I8);
    nombre fuera del patrón → **`400`** y nombre válido inexistente o de identidad en borrador →
    **`404`** (M6/C4: un archivo retirado deja de servirse pero no se borra, y al republicar vuelve);
    descarga correcta con `nosniff`/`inline`/`no-store` y sin tocar el disco en los rechazos
    (path traversal bloqueado).
  - **Criterio de terminado**: `quickstart.md` §6 ejecutable; la portada no se rompe si falta la
    imagen (lo resuelve el frontend, T332) y `seguridad` aprueba las cabeceras.
  - **Commit sugerido**: `feat(portada): subida y descarga segura de imágenes`

- [ ] T324 · `portada`: `handler_admin.go` — identidad, quiénes somos y contacto · `[backend]`

  - **Archivos**: `backend/internal/portada/handler_admin.go` (NUEVO),
    `handler_admin_test.go` (NUEVO).
  - **Qué hace**: `PUT /api/v1/admin/portada/identidad`, `/quienes-somos` y `/contacto` según el
    contrato: decodifica el DTO, valida con `platform/validate` (errores con `details` por campo),
    llama al service (T319) y responde `200` con el `*Admin` guardado. Nada de SQL en el handler
    (R4) y nada de credenciales en respuestas (FR-026 de F2). *Cubre FR-011, FR-015* (y sobre la
    validación, FR-002/FR-003/FR-007).
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/portada/` en verde: `200` con el sobre correcto; DTO inválido → `400` con
    `details` del campo (también JSON mal formado); sin sesión → `401` y sin permiso → `403` (los
    monta la cadena, T326); el servicio recibe exactamente el DTO validado.
  - **Criterio de terminado**: `quickstart.md` §3 ejecutable (US2, errores de FR-015 visibles).
  - **Commit sugerido**: `feat(portada): endpoints de identidad, quiénes somos y contacto`

- [ ] T325 · `portada`: `handler_admin.go` — horario, WhatsApp y redes · `[backend]`

  - **Archivos**: `backend/internal/portada/handler_admin.go` (EDITADO), `handler_admin_test.go`.
  - **Qué hace**: `POST`/`PATCH`/`DELETE` de `/api/v1/admin/portada/horario[/{id}]`,
    `/whatsapp[/{id}]` y `/redes[/{id}]` según el contrato (`201`/`200`/`204`; `PATCH` con al menos
    un campo; `404 not_found` para `{id}` inexistente; `409 conflict` para duplicados). El
    `publicationState` viaja en el cuerpo (publicar/retirar por elemento, FR-013).
    *Cubre FR-004, FR-005, FR-006, FR-011 y FR-015*.
  - **Pruebas incluidas** (§III): `httptest` con service falso. **Cómo se verifica**:
    `go test ./internal/portada/` en verde: cada código del contrato; `PATCH` vacío → `400`;
    duplicado → `409`; `DELETE` → `204` sin cuerpo; los errores usan `ErrorEnvelope`.
  - **Criterio de terminado**: `quickstart.md` §4 y §5 ejecutables (US2/US3).
  - **Commit sugerido**: `feat(portada): endpoints de horario, WhatsApp y redes`

- [ ] T340 · `portada`: agregado del panel `GET /api/v1/admin/portada` (service + handler + pruebas) · `[backend]` *(añadida por el `analyze` C1)*

  - **Archivos**: `backend/internal/portada/service_admin.go` (EDITADO: `GetPortadaAdmin`),
    `service_admin_test.go`, `handler_admin.go` (EDITADO), `handler_admin_test.go` (AMPLIADOS).
  - **Qué hace**: implementa la operación que el contrato ya declara y que **ninguna tarea cubría**
    (`analyze` C1): `GET /api/v1/admin/portada` devuelve el **estado completo del módulo** para
    precargar el panel —identidad, «quiénes somos», contacto, horario, WhatsApp y redes— **en ambos
    idiomas y con `publicationState` por elemento** (incluidos los borradores: esta es la única vista
    que los ve, y exige el permiso `portada`). Service `GetPortadaAdmin`: singletons **`null` hasta
    el primer guardado**, colecciones en su sobre `{items}` (§8.1.2) con el `ORDER BY` del contrato y
    acotadas (≤50 servicios, ≤20 canales; redes ≤ catálogo) — sin paginación: es un documento de
    estado (Complexity Tracking 3 del plan). Handler: `200` con `PortadaAdmin`, sin resolver idiomas
    (aquí se devuelven los pares `Es`/`En` crudos, con `*En` en `null` cuando no hay traducción).
    *Cubre FR-011 y US3 esc. 7 (ver el estado de cada elemento)*.
  - **Pruebas incluidas** (§III): unitarias del service con fakes + `httptest` del handler.
    **Cómo se verifica**: `go test ./internal/portada/` en verde: singletons `null` con el módulo
    vacío; con datos, los seis sectores presentes con sus `publicationState` reales (borradores
    incluidos); sobres `{items: []}` (nunca `null`) en colecciones vacías; `*En` en `null` cuando no
    hay traducción; sin sesión → `401` y sin permiso → `403` (cadena de T326).
  - **Criterio de terminado**: `features/informacion` (T333–T336) puede precargar formularios y
    mostrar la píldora de estado de cada elemento con esta única llamada.
  - **Commit sugerido**: `feat(portada): agregado del panel GET /api/v1/admin/portada`

- [ ] T326 · `portada`: `routes.go` + `cmd/api/main.go` (permiso `portada` cableado) · `[backend]`

  - **Archivos**: `backend/internal/portada/routes.go` (NUEVO), `backend/cmd/api/main.go` y
    `main_test.go` (EDITADOS).
  - **Qué hace**: `RegisterPublic` publica `/api/v1/portada` y `/api/v1/media/{fileName}` **sin**
    middlewares de sesión (solo los globales; superficie pública de solo lectura) y `RegisterAdmin`
    publica el grupo **`/api/v1/admin/portada` con `middleware.AdminChain("portada", deps)`**
    (authn → guard de contraseña → **authz(`portada`)** → CSRF), **activando y cableando** el permiso
    del catálogo de F2 (FR-012/P3-11): `PermissionPortada = "portada"` como constante del dominio y
    el `Recorder` de la cadena = el de F3 (T321). El DI de `main.go` construye repository, `Store`,
    servicios y handlers con `platform/config` (T312). Prueba de humo: rutas nuevas publicadas con
    su sobre y `/healthz` intacto (§8.1.9).
  - **Pruebas incluidas** (§III): unitarias de `routes.go` (registro de rutas) + humo de
    `cmd/api/main_test.go`. **Cómo se verifica**: `go test ./...` en verde; por API forzada, una
    cuenta sin permiso recibe `403 forbidden` en **toda** operación del módulo (SC-005/SC-011) y su
    denegación queda registrada (T321); `/healthz` responde como en F1 y una ruta inexistente
    `404 not_found`.
  - **Criterio de terminado**: `quickstart.md` §8 ejecutable; `make up` publica la superficie
    completa de F3.
  - **Commit sugerido**: `feat(portada): rutas del módulo con permiso portada y cableado en main`

---

## Fase 8 — Frontend (marca, i18n, portada pública y panel)

- [ ] T327 · Marca: tokens, tipografías, logo y `BrandLogo` · `[frontend]` `[P5]`

  - **Archivos**: `frontend/src/index.css`, `frontend/tailwind.config.*` (EDITADOS),
    `frontend/public/brand/simiente-logo.jpeg` (NUEVO, copia de `resources/simiente.jpeg`),
    `frontend/src/features/publico/components/BrandLogo.tsx` y su prueba (NUEVOS).
  - **Qué hace**: tokens del Manual (R3-12/`ux.md` §1) con la **nomenclatura única adoptada por el
    `analyze` I5 (la de `ux.md`; sin alias)**: variables CSS + tema Tailwind `navy #1a2b4a`,
    `navy-soft`, `teal #00c9a7`, `teal-strong`, `white`, `cream #F5F2EC`, `coral #ff6b3d`,
    `leaf #217638`, y familias `--font-display` (Bebas Neue) / `--font-sans` (Poppins) /
    `--font-emotiva` (Playfair Display), cargadas con `@fontsource` (T302). `BrandLogo` muestra el
    logo con `object-contain`, **sin deformar, sin
    cambiar colores, sin girar y sin efectos** (FR-002: nada de `transform`/`filter`/sombras sobre
    el logo) y con dimensiones explícitas. *Cubre FR-002 y los criterios de contraste de
    FR-018/FR-019*.
  - **Pruebas incluidas** (§III): Vitest (el logo carga, lleva `alt` y conserva su proporción).
    **Cómo se verifica**: `npm test -- --run` en verde; `npm run build` carga las tres tipografías
    sin CDN externo; `grep` de los nombres de token en el CSS: solo `navy`/`teal`/… y
    `--font-display/sans/emotiva` (cero alias tipo `--color-azul`, I5); revisión de `revisor-codigo`
    sobre las reglas del logo (CSS del componente).
  - **Criterio de terminado**: la nomenclatura de tokens es **una sola** (la de `ux.md` §1, mapeada
    al Manual en `research.md` R3-12): ni `ux.md` ni el código usan nombres paralelos.
  - **Commit sugerido**: `feat(frontend): tokens de marca, tipografías y BrandLogo`

- [ ] T328 · `features/publico/i18n`: provider, diccionarios tipados y `localStorage` · `[frontend]` `[P5]`

  - **Archivos**: `frontend/src/features/publico/i18n/LanguageProvider.tsx`, `useLanguage.ts`,
    `messages.ts`, `es.ts`, `en.ts`, `i18n.test.tsx` (NUEVOS).
  - **Qué hace**: i18n propio sin librerías (R3-9): `LanguageProvider` con `lang` (`es|en`),
    `setLang` y `t()`; diccionarios con **la misma clave tipada**
    (`Record<PublicMessageKey, string>` — SC-006 garantizado en compilación), incluida la clave
    `brand.name` del *chrome* de fallback (R3-18/`analyze` M4). **Sin fallback de idioma en el
    cliente** *(analyze I2)*: el sitio público consume strings ya resueltos por el servidor (R3-2);
    si el panel necesita un helper para pares `{es, en}` (previsualización), vive en
    `features/informacion/` y solo se usa ahí. Memoria en **`localStorage['ss.lang']`** (entre
    visitas; **primera visita sin preferencia guardada → español** y **una re-visita con preferencia
    la respeta**, FR-010 ajustado el 2026-10-09/`analyze` C3); si el guardado falla, no persiste sin
    avisos. Actualiza el atributo `lang` del `html`. **Sin** auto-detección del idioma del navegador.
    *Cubre FR-008 y FR-010*.
  - **Pruebas incluidas** (§III): Vitest. **Cómo se verifica**: `npm test -- --run` en verde:
    ambos diccionarios tienen exactamente las mismas claves (la prueba itera el tipo); `setLang`
    cambia `t()` al instante y persiste en `localStorage`; sin preferencia guardada el idioma es
    `es`, y **con preferencia guardada (re-visita) se respeta** (C3); ninguna prueba exige fallback
    cliente de contenidos (I2: el fallback es del service, T318).
  - **Criterio de terminado**: el módulo es reutilizable por F4–F9 (exporta `LanguageProvider` y
    `useLanguage`) y `ux.md` §7.2 queda implementado tal cual (con `brand.name`).
  - **Commit sugerido**: `feat(frontend): i18n tipado es/en con memoria en localStorage`

- [ ] T329 · Componentes compartidos: `Button` (`accent`) y `StatusPill` (`draft`/`published`) · `[frontend]` `[P5]`

  - **Archivos**: `frontend/src/components/Button.tsx`, `StatusPill.tsx` y sus pruebas (EDITADOS).
  - **Qué hace**: las dos extensiones menores autorizadas (`ux.md` §6.1 / alineación 5): `Button`
    con variante **`accent`** (fondo `teal`, texto `navy`) para los CTAs de la portada y `StatusPill`
    con los estados **`draft`/`published`** («Borrador»/«Publicado», texto además de color — US3
    esc. 7). **Sin romper el panel de F2** (regresión de sus pruebas). *Cubre FR-013 (estado visible
    en el panel) y el estilo de FR-002*.
  - **Pruebas incluidas** (§III): Vitest de ambos componentes. **Cómo se verifica**:
    `npm test -- --run` en verde **incluida la regresión** de los componentes de F2; ninguna otra
    pieza del inventario único cambia.
  - **Criterio de terminado**: el inventario único de F2 queda ampliado solo con estas dos
    extensiones (documentadas aquí; lo revisa `revisor-codigo`).
  - **Commit sugerido**: `feat(frontend): Button accent y StatusPill draft/published`

- [ ] T330 · `api/portada.ts`: cliente público y de panel · `[frontend]`

  - **Archivos**: `frontend/src/api/portada.ts` y `portada.test.ts` (NUEVOS).
  - **Qué hace**: cliente sobre el tipo generado (T303) con el `client.ts` de F2 (credenciales +
    `X-CSRF-Token`): `getPortada(lang)`, `getPortadaAdmin()`, `updateIdentity/About/Contact`,
    `create/update/deleteService`, `create/update/deleteWhatsappChannel`,
    `create/update/deleteSocialLink` y `uploadImage(file)` (multipart). **Ningún `fetch` fuera de
    aquí** (regla de la skill). Tipos `PublicationState` y DTOs del contrato; nada de `any`.
    *Soporta FR-001 y FR-011*.
  - **Pruebas incluidas** (§III): Vitest + MSW. **Cómo se verifica**: `npm test -- --run` en verde:
    cada función llama a su ruta/método del contrato con su cuerpo; los errores se propagan como
    `ApiError` con `code`/`details` (para los mensajes por campo); `uploadImage` envía `FormData` con
    el campo `file`.
  - **Criterio de terminado**: los hooks de T332–T336 solo consumen este módulo.
  - **Commit sugerido**: `feat(frontend): cliente de API de la portada`

- [ ] T331 · Permisos y rutas: `PORTADA`, `/`, `/health` y `/panel/informacion` · `[frontend]`

  - **Archivos**: `frontend/src/lib/permissions.ts`, `frontend/src/features/roles/permissions.ts`,
    `frontend/src/app/router.tsx`, `frontend/src/app/layout.tsx` y sus pruebas (EDITADOS).
  - **Qué hace**: **activa el permiso `portada`** en el frontend (P3-11): constante
    `PORTADA = 'portada'` en `lib/permissions.ts` y `isPermissionAvailable` deja de marcarlo como
    "Disponible más adelante" (`features/roles/permissions.ts`, US2 esc. 7). Rutas (alineación 1):
    **`/` → portada pública** (`PublicLayout` + `HomePage`, T332), **`/health` → `StatusPage`**
    (la pantalla de F1, sin cambios; fuera de la navegación pública) y **`/panel/informacion`** bajo
    `RequirePermission code={PORTADA}`; entrada "Portada e información general" en el menú del panel
    **solo con el permiso**; `/sin-permiso` se reutiliza tal cual (F2). *Cubre FR-001, FR-012 y
    FR-016 (panel en español)*.
  - **Pruebas incluidas** (§III): Vitest (router y guards). **Cómo se verifica**:
    `npm test -- --run` en verde: `/` renderiza la portada; `/health` muestra «Estado del sistema»;
    `/panel/informacion` sin permiso va a `/sin-permiso` y sin sesión a `/login`; el menú no muestra
    la entrada sin permiso; la regresión de las rutas de F2 pasa.
  - **Criterio de terminado**: `quickstart.md` §1 y §8 navegables; ningún e2e de F2 queda apuntando
    a `/` como pantalla de estado (los ajusta T337/T338 si hiciera falta, RG3-11).
  - **Commit sugerido**: `feat(frontend): rutas de la portada, /health y /panel/informacion con permiso`

- [ ] T332 · `features/publico`: `PublicLayout`, `HomePage` y secciones · `[frontend]`

  - **Archivos**: `frontend/src/features/publico/` (`pages/HomePage.tsx`,
    `components/PublicLayout.tsx`, `LanguageSwitcher.tsx`, `IdentityHero.tsx`,
    `WhoWeAreSection.tsx`, `ScheduleSection.tsx`, `WhatsAppSection.tsx`, `SocialSection.tsx`,
    `SocialIcon.tsx`, `ContactSection.tsx`, `PublicFooter.tsx`, `SectionSkeleton.tsx`,
    `hooks/usePublicHomeData.ts`, pruebas) (NUEVOS).
  - **Qué hace**: la portada pública de `ux.md` §4.1 (US1/US4/US5): `PublicLayout` (cabecera con
    marca + navegación de anclas visibles —**sin menú hamburguesa**, D-5— + `LanguageSwitcher` +
    pie), `HomePage` que compone **solo las secciones presentes** en la respuesta (SC-012: cero
    secciones vacías), `IdentityHero` (logo `object-contain`, H1 con el nombre, misión/visión/lema;
    sin imagen → fondo `navy` limpio, nunca hueco), `WhoWeAreSection` (texto plano con saltos),
    `ScheduleSection` (día localizado desde `dayOfWeek` en **selector/etiqueta de i18n, nunca texto
    libre**; rango «10:00–12:00» cuando hay `endTime`, `analyze` C2; nombre y lugar), `WhatsAppSection`
    (CTA con un clic: `wa.me`/grupo, SC-009), `SocialSection` (`target="_blank" rel="noopener"`,
    `SocialIcon` SVG propio sin dependencias), `ContactSection` (enlaces nativos de
    dirección/correo/teléfono), `PublicFooter` (frase `footer.welcome` del diccionario i18n, **no**
    editable — `analyze` I4) y `SectionSkeleton`. Contenidos renderizados **tal como los resolvió el
    servidor** (strings ya localizados; **sin** fallback cliente — `analyze` I2 —); `alt` de imágenes
    desde los datos (FR-019);
    responsivo ≥44 px y semántica accesible (FR-018/FR-019, `ux.md` §8.1; ancho mínimo **320 px**,
    `analyze` I7). *Cubre US1, US4, US5 y
    FR-001…FR-007, FR-009, FR-010, FR-013, FR-018, FR-019*.
  - **Pruebas incluidas** (§III): Vitest + Testing Library + MSW. **Cómo se verifica**:
    `npm test -- --run` en verde: secciones ausentes no se renderizan (SC-012); en inglés, un
    contenido sin versión en inglés llega ya resuelto en español desde el MSW y se renderiza tal cual
    (SC-006; la prueba **no** exige fallback en el componente, I2); `LanguageSwitcher` cambia
    la página al instante y persiste (FR-010); los enlaces de WhatsApp/red apuntan al destino
    correcto (SC-009); las imágenes llevan `alt`; un servicio con `endTime` muestra el rango y sin él
    solo la hora de inicio (C2); el estado cargando usa `SectionSkeleton` y el error
    se muestra sin romper la página.
  - **Criterio de terminado**: `quickstart.md` §1, §7 y §10 navegables a mano (US1/US4/US5).
  - **Commit sugerido**: `feat(frontend): portada pública con secciones, marca e i18n`

- [ ] T333 · `features/informacion`: `InformationPage` (pestañas del módulo) · `[frontend]`

  - **Archivos**: `frontend/src/features/informacion/pages/InformationPage.tsx`,
    `messages.ts`, `information.test.tsx` (NUEVOS).
  - **Qué hace**: la vista del módulo en el panel (`ux.md` §4.3, **en español**, FR-016): título
    «Portada e información general», banda de ayuda, **`Tabs`** de F2 con las 6 piezas
    (**Identidad · Quiénes somos · Horario de servicios · WhatsApp · Redes sociales · Contacto**) y
    el estado («Borrador»/«Publicado») de cada elemento a la vista (US3 esc. 7). Estados de
    `ux.md` §5 (cargando, vacío, error, éxito con `Notice` + `aria-live`) y `messages.ts` con el
    catálogo de textos de `ux.md` §7.1. *Cubre FR-011, FR-012 (interfaz solo con permiso) y FR-016*.
  - **Pruebas incluidas** (§III): Vitest + MSW. **Cómo se verifica**: `npm test -- --run` en verde:
    las 6 pestañas navegan; estado cargando/vacío/error correctos; los textos son los de `ux.md`
    §7.1 (voz de marca del Manual).
  - **Criterio de terminado**: el esqueleto del módulo listo para los formularios de T334–T336.
  - **Commit sugerido**: `feat(frontend): InformationPage con pestañas del módulo`

- [ ] T334 · `features/informacion`: `IdentityForm` + `ImageUploader` + `PublishControls` · `[frontend]`

  - **Archivos**: `frontend/src/features/informacion/components/IdentityForm.tsx`,
    `ImageUploader.tsx`, `PublishControls.tsx` y sus pruebas (NUEVOS); hooks
    `useIdentity.ts`, `useUpdateIdentity.ts`, `useUploadImage.ts` (NUEVOS).
  - **Qué hace**: formulario de identidad (`ux.md` §4.4: nombre oficial, lema/misión/visión con
    pestaña «English (opcional)», logo e imagen de portada) con **`ImageUploader`** (campo de archivo
    + previsualización inmediata + `alt` es/en obligatorio en español + aviso de reglas del manual y
    del recorte de la imagen de portada) y **`PublishControls`** («Publicar»/«Retirar de la portada»
    con `ConfirmDialog`: **la sección identidad publica por su cuenta**, D-2/FR-013). Validación
    cliente con Zod (espejo de FR-015; la autoridad sigue siendo el servidor) y errores junto al
    campo. *Cubre FR-002, FR-011, FR-013, FR-015 y FR-019*.
  - **Pruebas incluidas** (§III): Vitest + MSW. **Cómo se verifica**: `npm test -- --run` en verde:
    subida inválida muestra el aviso sin romper el formulario y la portada no se ve afectada;
    guardar sin `alt` con imagen → error del servidor mostrado junto al campo; publicar/retirar
    cambia la píldora y avisa («Publicado en la portada.» / «Retirado de la portada…»); al editar un
    elemento publicado el aviso recuerda que ya es visible (FR-014, `ux.md` §7.1).
  - **Criterio de terminado**: `quickstart.md` §3 (identidad) y §6 (imágenes) navegables.
  - **Commit sugerido**: `feat(frontend): identidad con ImageUploader y PublishControls`

- [ ] T335 · `features/informacion`: `WhoWeAreForm` + `ContactForm` · `[frontend]`

  - **Archivos**: `frontend/src/features/informacion/components/WhoWeAreForm.tsx`,
    `ContactForm.tsx` y pruebas (NUEVOS); hooks `useWhoWeAre.ts`, `useUpdateWhoWeAre.ts`,
    `useContact.ts`, `useUpdateContact.ts` (NUEVOS).
  - **Qué hace**: `WhoWeAreForm` (`ux.md` §4.5: texto plano con contador de **1.000**, pestañas
    es/en, inglés opcional) y `ContactForm` (`ux.md` §4.9: dirección es/en, correo y teléfono
    obligatorios), ambos con **publicar/retirar por sección** (`PublishControls`: cada singleton
    publica **por separado**, FR-013/D-2) y validación cliente espejo de FR-015. *Cubre FR-003,
    FR-007, FR-011, FR-013 y FR-015*.
  - **Pruebas incluidas** (§III): Vitest + MSW. **Cómo se verifica**: `npm test -- --run` en verde:
    el contador avisa al superar 1.000 y el guardado se bloquea con el mensaje del servidor;
    guardar sin español → error junto al campo; teléfono/correo inválidos → mensajes de `ux.md`
    §7.1; un campo «English» vacío o con espacios guarda sin error (US2 esc. 5/I6); retirar una
    sección no afecta a las demás.
  - **Criterio de terminado**: `quickstart.md` §3 (quiénes somos y contacto) navegables.
  - **Commit sugerido**: `feat(frontend): quiénes somos y contacto con publicación por sección`

- [ ] T336 · `features/informacion`: horario, WhatsApp y redes · `[frontend]`

  - **Archivos**: `frontend/src/features/informacion/components/ServicesList.tsx`,
    `ServiceForm.tsx`, `WhatsAppList.tsx`, `WhatsAppForm.tsx`, `SocialsList.tsx` y pruebas
    (NUEVOS); hooks `useServices.ts`, `useCreateService.ts`, `useUpdateService.ts`,
    `useWhatsappChannels.ts`, `useCreateWhatsappChannel.ts`, `useUpdateWhatsappChannel.ts`,
    `useSocials.ts`, `useUpdateSocial.ts` y los de borrado/publicación (NUEVOS).
  - **Qué hace**: los tres listados con sus formularios (`ux.md` §4.6–§4.8): servicios (**día en
    `Select` con las 7 opciones localizadas por i18n —nunca texto libre—**, `startTime` "HH:MM" y
    **`endTime` opcional** para el rango «10:00 a. m. − 12:00 m.» — `analyze` C2 —, nombre, lugar +
    «English (opcional)»), canales (nombre, tipo directo/grupo con ayuda contextual
    del destino) y redes (**filas fijas del catálogo** con URL por red y estado), cada fila con
    píldora de estado y **publicar/retirar por elemento** (`PublishControls`) + borrado con
    confirmación. Estados vacíos de `ux.md` §4.10 («Todavía no hay servicios…»). Los campos
    «English» vacíos se envían como `""` y el servidor los guarda como `null` (I6). **Sin** mensaje
    de "servicio duplicado" (no existe esa regla: `analyze` I3). *Cubre FR-004,
    FR-005, FR-006, FR-011, FR-013, FR-014 y FR-015*.
  - **Pruebas incluidas** (§III): Vitest + MSW. **Cómo se verifica**: `npm test -- --run` en verde:
    alta válida deja el elemento en «Borrador» y el aviso de `ux.md` §7.1; el `Select` de día ofrece
    las 7 opciones localizadas y `endTime` es opcional (con fin anterior al inicio muestra el error
    del servidor, C2); publicar → visible en el
    ciclo (SC-003/SC-004); enlace mal formado, red repetida o fuera del catálogo → errores del
    servidor junto al campo («cada red admite un solo enlace…»); un campo «English» vacío guarda sin
    error (US2 esc. 5/I6); retirar → confirmación clara y el
    elemento sigue en el panel.
  - **Criterio de terminado**: `quickstart.md` §4 y §5 navegables (US2/US3 completos).
  - **Commit sugerido**: `feat(frontend): horario, WhatsApp y redes con publicación por elemento`

---

## Fase 9 — E2E y cierre

- [ ] T337 · E2E Playwright `portada-publica.spec.ts` · `[frontend]` `[P7]`

  - **Archivos**: `frontend/e2e/portada-publica.spec.ts` (NUEVO).
  - **Qué hace**: el recorrido del visitante (US1, US3 público, US4, US5) sobre
    `quickstart.md` §1/§5/§7/§10: sin sesión ve las secciones publicadas y **ningún borrador**
    (SC-001/SC-002), tampoco sus imágenes (**la descarga de un elemento retirado responde `404`**,
    `analyze` C4); las secciones sin publicados no aparecen (SC-012); el selector cambia a inglés
    (interfaz + contenidos resueltos por el servidor sin huecos, SC-006), persiste al navegar, una
    **re-visita con preferencia guardada respeta el idioma** y solo una carga limpia sin preferencia
    entra en español (FR-010/`analyze` C3); los enlaces de WhatsApp/red abren el destino correcto
    con un clic (SC-009); recorrido por **teclado** y en **3 anchos** (320/768/1280 — el mínimo
    320 px es el que promete `ux.md`, `analyze` I7 —) sin
    desplazamiento horizontal (SC-007/SC-008).
  - **Pruebas incluidas** (§III): Playwright (`make e2e` / `npx playwright test`). **Cómo se
    verifica**: la spec pasa en verde localmente; valida SC-001, SC-002, SC-006, SC-007, SC-008,
    SC-009 y SC-012.
  - **Criterio de terminado**: US1, US3 (lado público), US4 y US5 cubiertas de extremo a extremo.
  - **Commit sugerido**: `test(e2e): portada pública (visitante, idioma, accesibilidad)`

- [ ] T338 · E2E Playwright `portada-panel.spec.ts` · `[frontend]` `[P7]`

  - **Archivos**: `frontend/e2e/portada-panel.spec.ts` (NUEVO).
  - **Qué hace**: el recorrido del equipo (US2, US3 en el panel) sobre `quickstart.md` §2–§6,
    §8 y §9: login con permiso `portada`, **precargar el módulo con `GET /api/v1/admin/portada`**
    (agregado de T340: singletons `null` al inicio, píldoras de estado), editar identidad/quiénes
    somos/contacto/horario (con rango `endTime` opcional, `analyze` C2)/WhatsApp/
    redes, subir el logo (y comprobar que queda registrado `home.image.upload`, I8), **publicar/retirar
    por sección y por elemento** y comprobar que el cambio
    publicado es visible en la portada **desde la primera carga** (SC-003/SC-004) y que al retirar la
    identidad su imagen deja de servirse (`404`, C4); errores de
    FR-015 visibles junto al campo; cuenta **sin permiso**: sin entrada en el menú, `/panel/informacion`
    forzada → `/sin-permiso` y API forzada → `403` (SC-005/SC-011); todo queda en la **auditoría de
    F2** con quién/qué/cuándo (SC-013).
  - **Pruebas incluidas** (§III): Playwright. **Cómo se verifica**: la spec pasa en verde; valida
    SC-003, SC-004, SC-005, SC-011 y SC-013.
  - **Criterio de terminado**: US2 y US3 cubiertas de extremo a extremo.
  - **Commit sugerido**: `test(e2e): gestión de la portada desde el panel (publicar/retirar)`

- [ ] T339 · Verificación de cierre · `[infra]`

  - **Archivos**: ninguno (evidencia en el PR). *No se toca* `.github/workflows/ci.yml` ni ningún
    archivo del kit.
  - **Qué hace**: ejecuta la verificación completa de F3 y la registra como evidencia: `make ci`
    (**lint + test + security**; `make db-migrate` es un comando aparte y se ejecuta al preparar el
    entorno, §0), `make e2e`, `make sqlc-verify` y `make api-gen` sin deriva, y el recorrido manual
    de `quickstart.md` §0–§12 (mapa §12: SC-001…SC-013). Comprueba además que `GET /healthz` responde
    como en F1, que una ruta inexistente responde `404` con `ErrorEnvelope` (§8.1.9), que la
    cobertura de `internal/portada/service*.go` es **≥ 80 %** (`go test -cover`) y que el
    `quickstart` §10 (accesibilidad) queda recorrido por QA. **Prueba de usabilidad SC-010**
    *(protocolo mínimo, `analyze` M7)*: **≥ 6 personas** (2 por franja 18–35 / 36–59 / 60+, incluidas
    con poca experiencia en internet), guion de **3 tareas sin ayuda** sobre la portada real (encontrar
    el horario de servicios, encontrar un canal de WhatsApp, cambiar a inglés y volver a español),
    sesión breve observada por `qa-tester` con el humano; resultado por tarea y persona documentado
    en **`specs/003-portada-info-general/pruebas-usabilidad-SC-010.md`** adjunto al cierre (umbral:
    ≥ 90 % de tareas completadas). Es una prueba **manual con personas** (no automatizable) y queda
    fuera de `make ci`/`make e2e`.
  - **Pruebas incluidas** (§III): — (es la ejecución de todas las anteriores). **Cómo se verifica**:
    cada comando anterior en verde y su salida adjunta al PR; el mapa de criterios → secciones queda
    recorrido completo y el informe de usabilidad de SC-010 adjunto.
  - **Criterio de terminado**: `make ci` y `make e2e` en verde; SC-001…SC-013 evidenciados. La
    revisión final (`qa-tester`, `revisor-codigo`, `seguridad` en paralelo) y el PR/CHANGELOG/README
    los cierra el orquestador con `documentador` y `devops` en la fase de entrega (no son tareas de
    este documento).
  - **Commit sugerido**: — (sin cambios de código)

---

## Cobertura de requisitos (FR-001…FR-019)

| FR | Tareas | Verificación principal |
|---|---|---|
| FR-001 (portada pública sin autenticación) | T301, T308, T318, T322, T331, T332 | quickstart §1 · e2e T337 (SC-001) |
| FR-002 (identidad + Manual de Identidad) | T302, T306, T314, T319, T323, T327, T334 | quickstart §3/§6 · revisión de `BrandLogo` (SC-001) |
| FR-003 («quiénes somos» plano ≤1.000) | T306, T317, T319, T335 | pruebas del service/handler · quickstart §3 |
| FR-004 (horario de servicios) | T306, T308, T315, T320, T325, T336 (día en selector localizado + `startTime`/`endTime` opcional, C2) | pruebas de service/repository · quickstart §4 |
| FR-005 (canales de WhatsApp) | T306, T308, T315, T320, T325, T336 | pruebas de service · e2e T337 (SC-009) |
| FR-006 (redes, catálogo y un enlace por red) | T306, T315, T320, T325, T336 | integración (`UNIQUE (network)`) · quickstart §4 |
| FR-007 (datos de contacto) | T306, T314, T319, T335 | quickstart §3 |
| FR-008 (contenido es/en + interfaz bilingüe) | T306, T313, T328, T332, T334–T336 | pruebas de i18n y de service (SC-006) |
| FR-009 (fallback `en → es`, nunca vacío) | T318, T328, T332 | tabla de casos del service · e2e T337 (SC-006) |
| FR-010 (selector de idioma; primera visita en español) | T328, T332 | pruebas de i18n · e2e T337 |
| FR-011 (gestión completa desde el panel) | **T340** (agregado `GET /api/v1/admin/portada`, `analyze` C1: US3 esc. 7 — estado de cada elemento —), T319, T320, T324, T325, T330, T333–T336 | quickstart §2–§6 · e2e T338 (SC-004) |
| FR-012 (permiso «portada e información general») | T321, T326, T331 | pruebas de handler (403 por API forzada) · e2e T338 (SC-005, SC-011) |
| FR-013 (borrador/publicado por elemento; nada de borradores al público) | T306, T308, T314–T316, T318, T319, T320, T322, T323 (**descarga de imágenes solo de contenido publicado**, `analyze` C4), T329, T332, T334–T336 | pruebas "0 borradores expuestos" · e2e T337 (SC-002, SC-012) |
| FR-014 (se publica al guardar) | T318, T319, T320, T336 | e2e T338 (SC-003) |
| FR-015 (validación completa de entradas) | T310, T317, T319, T320, T324, T325, T334–T336 | tablas de casos (service/handler) · quickstart §3/§4 |
| FR-016 (panel en español) | T333–T336 | revisión de `revisor-codigo` (0 cadenas del panel pasadas por `t()`) |
| FR-017 (auditoría de F2 ampliada, atómica) | T307, T311, T313, T316, T319–T321, T323 (`home.image.upload` en la subida, `analyze` I8) | integración (mutación+registro) · quickstart §9 (SC-013) |
| FR-018 (responsividad y navegación simple) | T327, T332 | e2e T337 en 3 anchos · quickstart §10 (SC-007) |
| FR-019 (accesibilidad y `alt` de imágenes) | T306, T309, T323, T327, T332, T334 | checklist WCAG de quickstart §10 · e2e T337 (SC-008) |

También quedan cubiertos los **Edge Cases** de la spec sin FR propio: **secciones vacías ocultas**
(T318, T332, e2e T337) · **texto con tildes/`ñ`/`¿¡`/emojis intacto** (T319, T334–T336) · **imagen
inválida o fallida sin romper la portada** (T323, T332, T334) · **ediciones simultáneas** (última
escritura completa gana: T314/T319) · **canal duplicado exacto** y **red repetida/fuera de catálogo**
(T320, T336) · **sin permiso por cualquier vía** (T326, T331, e2e T338) · **un fallo de la edición no
deja registro aplicado ni se aplica sin registro** (T316, T319/T320).
