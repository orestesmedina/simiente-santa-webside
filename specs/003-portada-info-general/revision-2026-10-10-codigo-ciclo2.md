# Re-validación de código (ciclo 2) — F3 Portada e información general

**Fecha**: 2026-10-10 · **Rama**: `003-portada-info-general` · **Revisor**: `revisor-codigo` (solo
lectura; este reporte es la única escritura). **Alcance**: cierre de los hallazgos de
`revision-2026-10-10-codigo.md` (ciclo 1) sobre las correcciones `d5d4610` (B1) y `fc96185` (I1),
más regresiones de esas dos correcciones. No se re-revisa el resto de F3 ya validado en el ciclo 1.

## Verificaciones ejecutadas

| Comando | Resultado |
|---|---|
| `cd backend && gofmt -l .` | sin salida ✓ |
| `go vet ./...` | ✓ |
| `golangci-lint run ./...` | `0 issues` ✓ |
| `go test ./...` | todo `ok` ✓ |
| `go test -tags=integration ./internal/portada/...` (Docker + `DATABASE_URL_TEST` del `.env`) | `ok 17.8s` ✓ (sin skips) |
| Cobertura ponderada de `service*.go` (`internal/portada`) | **85.3 %** (529/620 sentencias) ≥ 80 % ✓ |
| `cd frontend && npm run lint` | limpio ✓ |
| `npm run typecheck` (TS strict) | ✓ |
| `npm test -- --run` | 271 pruebas, todo en verde ✓ |

## Estado de los hallazgos del ciclo 1

### B1 — CERRADO ✅

Los PATCH ahora distinguen `null` de campo ausente y la limpieza persiste y se audita.

- **Presencia real del campo**: `backend/internal/portada/optional.go:21-56` define
  `Optional[T]` con `UnmarshalJSON` propio que separa los tres estados (ausente → `Set()=false`,
  valor, `null` explícito → `Set()=true, Null()=true`). Como `encoding/json` solo invoca
  `UnmarshalJSON` cuando la clave existe, la ausencia deja el valor cero ("ausente") y no toca el
  campo — comportamiento correcto.
- **Aplicado exactamente a los campos anulables del contrato**: `service_admin.go:156-178`
  (`endTime`, `nameEn`, `descriptionEs`, `descriptionEn`, `placeEn` en `ScheduleItemPatch`;
  `nameEn` en `WhatsappChannelPatch`), que son exactamente los `type: [string, "null"]` de
  `backend/api/openapi.yaml:2608-2633,2713-2715`. Los no anulables siguen como `*T`.
- **`null` vacía el campo**: `mergeSchedule` (`service_admin.go:736-744` para `endTime` y
  `:756-781` para los `*En`) y `mergeWhatsappChannel` (`:827-829`) pasan
  `patch.X.Value()` ("" cuando vino `null`) a `normalizeEndTime` (`:964-976` → nil) y
  `normalizeOptional` (`service.go:185-191` → nil), es decir, `null` ⇒ `NULL` en BD. El
  `repository` lo lleva a NULL con `pgTextPtr` (`repository_home.go:284`).
- **Ausente no toca**: solo los campos con `Set()`/puntero no-nil entran en los `merge*` ✓.
- **`{"endTime": null}` ya no es 400**: `hasChanges` (`service_admin.go:694-704`) cuenta `.Set()`,
  y `TestPatchScheduleJSONNullClearsEndTime` y `TestPatchScheduleNullOnlyIsAccepted`
  (`service_patch_null_test.go:43-81,109-124`) verifican que el cuerpo exacto del panel
  decodifica, no responde «No hay cambios» y limpia `endTime`/`nameEn` — igual en WhatsApp
  (`:85-104`). Las pruebas decodifican el JSON real (`encoding/json`), no structs construidos a
  mano, que era el vacío del ciclo 1.
- **Auditoría M5**: `TestPatchScheduleNullClearsAndAudits` (`service_admin_test.go:478-498`)
  verifica que un PATCH de solo `null` deja **una** fila `home.schedule.update` (limpiar es cambio
  de datos, no de estado) ✓.
- **Sin regresiones**: el contrato no cambió (`d5d4610` no toca `backend/api/openapi.yaml` ni
  `contracts/`); `DisallowUnknownFields` sigue activo (`handler.go:135`) y operante, porque
  `Optional` implementa `UnmarshalJSON` solo en los campos, no en el struct
  (`service_admin.go:151-153` lo documenta); capas intactas (`optional.go` solo importa
  `encoding/json`); `golangci-lint` 0 issues; el diff de `service_admin.go` en `d5d4610` (41+/36−)
  es mínimo y solo toca la semántica de presencia.

### I1 — CERRADO ✅

- `frontend/src/features/publico/schedule.ts:33-47` (`formatClockTime`) presenta la hora del dato
  24 h en formato localizado con claves i18n (`schedule.time.am/pm/noon` en
  `i18n/es.ts:36-38`, `i18n/en.ts:36-38`, catálogo `i18n/messages.ts:42-44`), y
  `formatTimeRange` (`:54-61`) compone el rango. El resultado es literalmente el ejemplo de
  `ux.md:227`/`ux.md:441` (D-3): `formatTimeRange('10:00','12:00')` → «10:00 a. m. − 12:00 m.»
  (`schedule.test.ts:29`), con medianoche «12:00 a. m.» y en «10:00 AM − 12:00 PM».
- El **dato sigue siendo 24 h**: el patrón del contrato (`openapi.yaml:2607,2610`) no cambió y
  solo se altera la presentación (`ScheduleSection.tsx:53,73`); el panel sigue editando «HH:MM»
  (`ServiceForm.tsx` con `TIME_PATTERN`).
- Pruebas unitarias cubren es/en, `endTime` null/ausente y valores fuera de contrato
  (`schedule.test.ts:26-51`); el e2e quedó alineado (`e2e/portada-publica.spec.ts:195`).
- Único defectillo residual → ver **N1** abajo (no era parte del alcance literal de I1).

### M1 — ABIERTO ⚠️ (menor)

Sin cambios: `UploadImage` (`service_admin.go:325-352`) sigue devolviendo error en archivo vacío
(`:329-334`), formato inválido y demasiado grande (`:338-351`) sin `recordContentFailure`, frente
al patrón del resto de mutaciones (p. ej. `:426-441`). La excepción sigue sin documentarse.

### M2 — ABIERTO ⚠️ (menor)

Sin cambios esenciales: los `merge*` (`service_admin.go:714-905`) marcan `dataChanged` por la mera
presencia del campo, no por diferencia con el valor actual. Nota: tras B1 la decisión de contar
`null` como cambio es **correcta** (vaciar es un cambio real, exigido por M5 y verificado por
`TestPatchScheduleNullClearsAndAudits`); lo que permanece abierto es el caso «Guardar sin tocar
nada» (PATCH con valores idénticos → fila `home.*.update`).

### M3 — ABIERTO ⚠️ (menor, aceptado por diseño)

Sin cambios: `SaveIdentity` lee `previous` fuera de la transacción (`service_admin.go:223-226`) y el
borrado de la imagen reemplazada es best-effort (R3-8). Sigue siendo deuda acotada documentada.

## Hallazgos nuevos (todos MENORES; ninguno bloquea)

**N1 — El español marca «m.» cualquier hora de la franja de las 12, no solo el mediodía exacto.**
- **Dónde**: `frontend/src/features/publico/schedule.ts:44-45`
  (`hours === 12 ? 'schedule.time.noon' : …`).
- **Qué**: «12:45» se presenta como «12:45 m.»; por norma (RAE) y por el propio ejemplo de
  `ux.md` («12:00 m.»), «m.» corresponde al mediodía **exacto** y «12:45 p. m.» a la franja
  posterior. En inglés es correcto («12:45 PM»).
- **Recomendación**: `hours === 12 && match[2] === '00' ? 'schedule.time.noon' : hours < 12 ?
  'schedule.time.am' : 'schedule.time.pm'`, más un par de casos en `schedule.test.ts`
  («12:45» → «12:45 p. m.», «12:00» → «12:00 m.»).

**N2 — `Optional[T]` no implementa `MarshalJSON` (trampa latente, sin impacto hoy).**
- **Dónde**: `backend/internal/portada/optional.go:21-56`.
- **Qué**: los campos son no exportados, así que si algún día se serializa un `*Patch` cada campo
  `Optional` saldría `{}` y, al redondear, se interpretaría como "vaciar". Verificado que hoy nada
  serializa los PATCH (solo se decodifican).
- **Recomendación**: añadir `MarshalJSON` (ausente → omitir; `null` → `null`; valor → valor) o un
  test que congele el formato de cable del tipo.

**N3 — Los campos de los PATCH no validan longitud en el servicio (pre-existente, no regresión).**
- **Dónde**: `backend/internal/portada/service_admin.go:156-185` (los `Patch` no llevan tags
  `validate` ni hay chequeo manual de longitud en los `merge*`).
- **Qué**: un `nameEn`/`description*`/`place*` más largo que el `maxLength` del contrato no se
  rechaza con `details` por campo (como sí hacen los `Input` de POST/PUT con `validate:"max=…"` en
  `model.go:143-193`): llega al `CHECK` de la BD y responde 400 genérico «Algún dato no cumple las
  reglas de la base de datos» (`repository.go:263-267`). Nota relacionada: un `null` en un campo no
  anulable (p. ej. `{"nameEs": null}`, que el contrato declara `type: string`) se ignora como
  ausente en lugar de responder 400 — comportamiento pre-existente que el comentario de
  `service_admin.go:151-152` documenta.
- **Recomendación**: tags `validate:"omitempty,max=…"` en los campos de los `Patch` (y `*int` con
  `min`/`max`) o validación manual equivalente, para que los errores de longitud sean 400 con
  `details` por campo.

## Sugerencias (sin severidad)

- Prueba JSON-explícita de "clave **ausente** no toca el valor" (hoy se ejercita a nivel de struct
  con `OptionalOf`); cerraría el triángulo ausente/valor/`null` de B1.
- Caso de integración de `repository` que actualice `end_time` a NULL y lo relea (el camino
  `pgTextPtr` → `sqlc.arg('end_time')` está cubierto solo por la prueba unitaria del service).

## Veredicto

**APROBADO** — 0 BLOQUEANTES. B1 (ciclo 1) cerrado con un único criterio y prueba de regresión que
ejercita el cuerpo real del panel, incluida la auditoría M5; I1 cerrado con la presentación
a.m./p.m. localizada prometida en `ux.md` y el dato intacto en 24 h; sin regresiones (contrato sin
cambios, `DisallowUnknownFields` operante, capas, `golangci-lint`, TS strict, integración real en
verde). Quedan 3 menores previos (M1–M3) y 3 menores nuevos (N1–N3) como deuda acotada; N1 es el
candidato natural a resolverse en este mismo PR por ser el único con efecto visible para el
usuario.
