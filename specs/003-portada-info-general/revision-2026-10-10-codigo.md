# Revisión de código — F3 Portada e información general

**Fecha**: 2026-10-10 · **Rama**: `003-portada-info-general` · **Revisor**: `revisor-codigo` (solo
lectura; este reporte es la única escritura). **Alcance**: `git diff main...HEAD` (código backend,
frontend, migraciones y contratos), contrastado con `spec.md`, `plan.md`, `research.md`,
`data-model.md`, `contracts/openapi.yaml`, `ux.md`, `tasks.md` (40/40), `analyze.md` y `cierre.md`,
y con la constitución y las skills `go-backend`/`react-frontend`/`postgres-db`.

## Verificaciones ejecutadas (sin Docker)

| Comando | Resultado |
|---|---|
| `cd backend && gofmt -l .` | sin salida ✓ |
| `go vet ./...` | ✓ |
| `golangci-lint run ./...` | `0 issues` ✓ |
| `go test ./...` | todo `ok` ✓ |
| `cd frontend && npm run lint` | limpio ✓ |
| `npm run typecheck` (TS strict) | ✓ |
| `npm test -- --run` | 45 archivos, 269 pruebas, todo en verde ✓ |
| Cobertura ponderada de `service*.go` (`internal/portada`) | **84.7 %** (525/620 sentencias) ≥ 80 % ✓ |

Arquitectura por capas verificada: handlers (`handler*.go`) solo decodifican/validan y delegan en el
puerto `PortadaService`; el service no conoce `net/http` ni `pgx`; el repository es el único que
habla con sqlc/PostgreSQL y mete la auditoría en la MISMA transacción. El contrato OpenAPI de la
aplicación (`backend/api/openapi.yaml` 0.4.0) contiene el delta de F3 idéntico al de
`contracts/openapi.yaml` (solo diferencias de comillas YAML). Las 16 operaciones están registradas
(RegisterPublic: 2 + RegisterAdmin: 14). `/healthz` de F1 intacto. Las migraciones `000001`–`000004`
no se tocan (solo se añaden `000005`/`000006`, con `up`/`down` completos y prueba up→down→up).

## Hallazgos

### BLOQUEANTE

**B1 — Los PATCH no vacían campos opcionales con `null`: cambio de usuario perdido en silencio.**
- **Dónde**: `backend/internal/portada/model.go:153-182` (los `Patch` usan `*string`), junto con
  `backend/internal/portada/service_admin.go:732-777` (`mergeSchedule`), `:822-825`
  (`mergeWhatsappChannel`) y `frontend/src/features/informacion/components/ServiceForm.tsx:87-89,131-137`
  y `WhatsAppForm.tsx:123-128` (envían `null` con `nullable()`).
- **Qué**: el contrato (`backend/api/openapi.yaml`, `ScheduleItemPatch`) promete literalmente:
  «Un campo `*En` con `null` o `""`/espacios vacía su traducción al inglés (analyze I6); `endTime:
  null` quita la hora de fin». Pero en Go, JSON `null` y campo **ausente** dejan el mismo valor
  (`*string == nil`; verificado con `encoding/json`), y los `merge*` solo actúan `if patch.X != nil`.
  El frontend —que cumple el contrato— envía `null` para limpiar (nunca `""`). Consecuencia real: el
  usuario quita la «hora de fin» o vacía una traducción al inglés de un servicio/canal, pulsa
  Guardar, ve «Guardado», el panel invalida `PORTADA_ADMIN_QUERY_KEY`
  (`useUpdateService.ts:17-19`) y tras el refetch **vuelve el valor anterior**. Además, un PATCH
  cuyo único campo es `{"endTime": null}` choca contra `hasChanges()`
  (`service_admin.go:690-705`) y responde `400 "No hay cambios que guardar"`, cuando el contrato
  admite el objeto con `minProperties: 1`.
- **Recomendación**: dar presencia real al campo en el backend (wrapper `optional[string]` con
  `UnmarshalJSON` propio o `json.RawMessage` que distinga ausente / `null` / `""`), **o** cambiar el
  contrato para que la limpieza sea `""` y ajustar el frontend (`nullable()` → enviar `""`) — pero
  siempre en un solo criterio y propagado. Añadir pruebas que ejerciten el camino real: hoy
  `service_admin_test.go:435-448` limpia con `&empty` (es decir `""`, que sí funciona) y las pruebas
  de frontend asumen que `null` funciona contra mocks MSW; por eso ningún test capturó la brecha.

### IMPORTANTE

**I1 — Formato de presentación del horario sin implementar (desviación de `ux.md`).**
- **Dónde**: `frontend/src/features/publico/schedule.ts:22-29` (`formatTimeRange` conserva «el
  formato de 24 h del dato») ↔ `ux.md:227` (§4.6) y `ux.md:441` (D-3): «el formato de presentación
  («10:00 a. m. − 12:00 m.») lo resuelve la interfaz a partir del dato de 24 h».
- **Qué**: la portada muestra `10:00 – 12:00` en 24 h, no el formato a.m./p.m. localizado que la UX
  aprobada promete para el público objetivo (personas de todas las edades, SC-010). La desviación se
  consigna en un comentario pero no está registrada como decisión ni en `tasks.md` ni en `cierre.md`.
- **Recomendación**: implementar la presentación localizada (es: «10:00 a. m. – 12:00 m.», en:
  «10:00 AM – 12:00 PM») en `formatTimeRange`, **o** registrar la decisión de mantener 24 h y
  ajustar `ux.md` §4.6/D-3 con el humano. Es presentación; no bloquea por sí solo.

### MENOR

**M1 — Las subidas de imagen rechazadas no dejan fila `result='failure'`.**
- **Dónde**: `backend/internal/portada/service_admin.go:326-348` (`UploadImage`): archivo vacío,
  formato inválido o demasiado grande devuelven error sin `recordContentFailure`, mientras el resto
  de mutaciones del módulo sí deja la fila de fallo best-effort (p. ej. `service_admin.go:193-199`).
- **Recomendación**: registrar `home.image.upload` con `result='failure'` en los tres rechazos (o
  documentar la excepción en R3-11), para que el patrón de auditoría sea uniforme.

**M2 — Ediciones sin cambio real de valor sí generan fila `home.X.update`.**
- **Dónde**: `mergeSchedule`/`mergeWhatsappChannel`/`mergeSocialLink` (`service_admin.go:710-900`)
  marcan `dataChanged` por la mera presencia del campo, no por diferencia con el valor actual.
- **Qué**: pulsar «Guardar» sin tocar nada escribe un `home.*.update` en la auditoría. No viola
  SC-013 (sobre-registro, no pérdida), pero la semántica del contrato («un **cambio de datos**
  registra…») es más estricta que la implementación.
- **Recomendación**: comparar con el valor actual antes de marcar `dataChanged` (y aprovechar para
  cubrir el caso `null` de B1).

**M3 — Posible imagen huérfana ante guardos concurrentes de identidad.**
- **Dónde**: `SaveIdentity` (`service_admin.go:220-239`) lee `previous` fuera de la transacción y
  borra la imagen reemplazada best-effort; dos guardos concurrentes que reemplazan la misma imagen
  pueden dejar el archivo intermedio sin borrar.
- **Recomendación**: aceptado por el diseño best-effort (R3-8); basta con dejar constancia o añadir
  una rutina de limpieza de huérfanos en una funcionalidad futura (F4+).

## Coherencia con `analyze.md` (segunda pasada) — verificada en el código

- **C1** cerrado: `GetPortadaAdmin` (`service_admin.go:1043-1102`, `handler_admin.go:108-123`,
  `GET /api/v1/admin/portada`) con singletons `null` al inicio y sobres `{items}` nunca `null` ✓.
- **C2** cerrado: horario estructurado (`day_of_week` 0–6, `start_time`/`end_time` con
  `CHECK (end_time IS NULL OR end_time > start_time)`, validación `400 details.endTime`) coherente en
  BD/servicio/contrato/frontend ✓.
- **C3** cerrado: `LanguageProvider` solo cae a español sin preferencia guardada; la re-visita
  respeta `localStorage['ss.lang']`, con prueba e2e explícita (`e2e/portada-publica.spec.ts:257`) ✓.
- **C4** cerrado: `GET /api/v1/media/{fileName}` solo sirve archivos referenciados por identidad
  **publicada** (`IsHomeFilePublished`, `internal/db/queries/home.sql:288-296`), `400` por patrón /
  `404` por no publicado, `Cache-Control: no-store` + `nosniff` + `inline` ✓.
- **N1/N2/N3/N4** cerrados en los artefactos y respetados por el código (prosa de auditoría «dos
  filas» en el contrato; `IsHomeFilePublished` único nombre; misión/visión opcionales en `ux.md:214`
  y en los DTOs; frase del pie como cadena i18n) ✓.
- **I2** cerrado: el fallback `en → es` vive solo en el service (`pickString`/`pickOptional`) y el
  frontend renderiza lo resuelto; sin helper de fallback duplicado ✓.
- **I6** cerrado: `""`/espacios → NULL en el service (`normalizeOptional`), salvo la semántica de
  `null` que queda en **B1**.
- **I8** cerrado: `home.image.upload` en `000006`, registro fail-closed con borrado del archivo si
  falla el registro (`service_admin.go:351-356`) ✓.

## Lo positivo (resumen)

Auditoría atómica fail-closed (contenido + filas en la misma transacción, deletes sin fila si el id
desaparece), advisory locks para los singletons, validación de dominio con `details` por campo,
catálogo de redes y hosts de WhatsApp acotados, imágenes por firma binaria (sin SVG/GIF) con nombre
generado por el servidor, superficie pública sin rastro de borradores (consultas `…Published`
separadas y DTOs públicos sin `publicationState`), i18n tipado que impide traducciones faltantes en
compilación, formularios de panel sobre componentes compartidos de F2 (markup no duplicado) y
pruebas significativas (MSW + componentes reales, integración de repositorio y e2e con
precondiciones deterministas).

## Veredicto

**RECHAZADO** — 1 BLOQUEANTE (B1: la semántica `null` del contrato en los PATCH no está implementada
y el panel pierde en silencio los cambios de «quitar hora de fin»/«vaciar traducción»), 1 IMPORTANTE
(I1: formato de presentación del horario prometido en `ux.md`) y 3 MENORES (M1–M3).

Requisito para volver a revisar: cerrar B1 con un único criterio (backend que distinga `null` de
ausente, o `""` como valor de limpieza en contrato + frontend) y pruebas unitarias que ejerciten
exactamente el cuerpo que envía el panel; cerrar o documentar I1. Los menores pueden resolverse en
el mismo PR o quedar registrados como deuda acotada.
