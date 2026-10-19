# Revisión de seguridad (ciclo 2) — F9 Medios (prédicas y podcasts)

- **Rama**: `009-medios-predicas-y-podcasts` · **HEAD**: `c65e67b` (repo limpio) · diff de corrección `8ee1fcf..f393ddb` (los 11 commits de código, `b1b01fe` y `df76506` entre ellos)
- **Fecha**: 2026-10-18 · **Revisor**: `seguridad` (Fase 7, revalidación del ciclo 1 de corrección)
- **Referencias**: informe anterior `revision-2026-10-18-seguridad.md` (veredicto APRUEBA con B1/B2/B3), `estado.md` bitácora del ciclo 1, OWASP Top 10 (2021).
- **Método**: lectura de los 11 commits de corrección y del estado final de los archivos que tocan (no solo los diffs); reejecución de la tabla compartida de barreras y del paquete `internal/medios` + `cmd/api` con `-count=1`; `govulncheck ./...`, `npm audit --audit-level=high`, `go vet`; greps de secretos y patrones peligrosos sobre el diff de corrección y sobre el árbol; **verificación en vivo de B1 contra el binario de HEAD** (instancia propia aparte, misma base de datos, solo GET públicos de lectura).

## Veredicto: **APRUEBA**

B1 y B2 quedan **cerrados** (verificados en código, por prueba y —B1— en vivo), B3 se mantiene como **falso positivo** (el diff de corrección no toca e2e), y **ninguna corrección abrió nada nuevo**: ninguna de las cuatro barreras de la red saliente se debilitó, la fábrica sigue siendo el único punto que construye un `http.Client`, las escrituras por columnas y la consulta por lote reducen alcance (no se saltan comprobaciones ni aceptan identificadores del cliente), y las transacciones nuevas mantienen la regla de auditoría atómica (P9-14). Sin críticas, altas, medias ni bajas nuevas; dos informativos (uno de ellos una nota de entorno que el equipo debe atender antes de la puerta 3).

---

## Estado de los hallazgos del ciclo 1

### B1 — `offset` sin tope superior en la ruta pública → **CERRADO** (commits `b1b01fe`)

**En código:**
- `service.go` `normalizarPaginacionPublica` ahora acota `offset > paginate.MaxOffset` → `MaxOffset` (misma regla que `limit`: el tope nunca es un error; la página está vacía igual).
- `platform/paginate/paginate.go` (`parseOffset`, líneas 93-95) acota igual, de modo que **también quedan cerrados los espejos de panel de F4/F6/F7/F9** que usan `FromValues` (la deuda señalada en la recomendación 1 del informe anterior).
- El contrato no cambia (`offset` ya declaraba `minimum: 0`); el sobre devuelve el valor acotado, como ya hacía con `limit`.

**Por prueba (reproducidas por el revisor, `-count=1`):** unitarias de `paginate`, del service y del handler público (`?offset=2147483648` → 200 con `offset = MaxOffset` y página vacía), y de integración contra PostgreSQL real (el offset en el tope es una página vacía; el valor sin acotar sigue siendo error de la base — la razón del tope). Paquete `internal/medios` + `cmd/api` en verde.

**En vivo (binario de `c65e67b` compilado por el revisor, instancia aparte en `:8091`, misma base de datos, GET públicos de lectura):**

| Petición a `GET /api/v1/medios` | Resultado |
|---|---|
| `?offset=2147483648&limit=5` (el caso del 500) | **200** `{"lang":"es","items":[],"total":9,"limit":5,"offset":2147483647}` |
| `?offset=2147483647&limit=5` (tope exacto) | 200, ídem (acotado al tope, sin cambio) |
| `?offset=9999999999999&limit=5` | 200 con `offset: 2147483647` (antes devolvía el valor sin acotar) |
| `?offset=4294967296&limit=5` | 200 con `offset: 2147483647` |
| `?offset=-5` | 400 `invalid` con `details.offset` |
| `?offset=9223372036854775808` (> int64) | 400 `invalid` «debe ser un número entero» |
| `?offset=abc` | 400 `invalid` con `details.offset` |
| `?offset=0&limit=2` (control) | 200 con items y `total` correctos |

Controles colaterales intactos: `Cache-Control: no-store` y `Content-Type: application/json` en la ruta pública, `?lang=fr` → 400, y el espejo de panel `GET /api/v1/admin/medios?offset=2147483648` sin sesión → 401.

### B2 — la allowlist fijaba el host pero no el puerto → **CERRADO** (commit `df76506`)

- **La barrera 1 exige el puerto**: `verificarURL` rechaza `u.Port()` distinto de vacío/`443` (`cliente_http.go:282-284`) — aplica a la petición inicial **y a cada redirect** (`checkRedirect` → `verificarURL`).
- **El dialer también lo exige** (defensa en profundidad): `dialSeguro` rechaza `puerto != "443"` (`:199-201`) y **conecta al 443 fijo, nunca al puerto que traiga la petición** (`:212`, `net.JoinHostPort(ips[0].String(), puertoPermitido)`).
- **Está en la tabla compartida de barreras**: el subtest «puerto distinto de 443 rechazado (barrera 1)» vive dentro de `ejecutarBarrerasCompartidas` y la corren los **cuatro** consumidores — verificado en verde con `-count=1`: `TestClienteSalienteBarreras` (sintético), `TestLectorFeedsBarreras` (feeds), `TestProbadorBarreras` (probador) y `TestFetchCoverDescargadorBarreras` (carátulas). El subtest además ejercita el dialer con puerto ajeno (espera el rechazo) y confirma que el 443 explícito sí sale, y que ninguna petición se envía cuando se rechaza.
- **Ninguna barrera se debilitó**: el commit solo añade (un error tipado, una constante, dos comprobaciones); quedan intactas la allowlist exacta por host y por consumidor (mapa sin comodines, con el test por consumidor), el rechazo de IPs privadas/loopback/link-local/metadata en cada conexión con conexión a la IP ya validada (sin rebinding), el tope de 10 saltos con `CheckRedirect` que ahora también valida el **puerto** del `Location` (más fuerte, no más débil), el `io.LimitReader` en todo cuerpo con el `cuerpoInfinito` que falla si alguien lo rebasa, `Proxy: nil`, `User-Agent` propio y los tiempos de espera. `nuevoClienteSaliente` sigue fallando sin allowlist/tope/tiempo.
- **La fábrica sigue siendo el único punto**: grep global — el único `http.Client` de producción de todo el backend sigue en `cliente_http.go:162`; las únicas apariciones de `http.Get` fuera son las pruebas del healthcheck de `platform/httpserver` (contra el servidor local, no red saliente); la única puerta exportada que arma la red sigue siendo `NewRedSaliente` (`red_saliente.go:41`).

### B3 — contraseñas de e2e hard-coded → **SE MANTIENE FALSO POSITIVO**

El diff de corrección (`8ee1fcf..f393ddb`) **no toca `frontend/e2e/`** (verificado con `git diff --stat`). Las cadenas siguen exactamente donde estaban (`mediosSeed.ts:1001`, `medios-panel.spec.ts:578,624,642`): cuentas creadas en caliente contra el stack local con cambio obligatorio de contraseña inmediato, sin secretos reales en el diff (grep de secretos sobre el diff de corrección: 0). Deuda heredada de F4/F6/F7; opcional su centralización, como ya se recomendó.

---

## Regresión de seguridad sobre los otros 9 commits — **sin hallazgos nuevos**

- **`3a2a380` (no pisar la carátula al unir; una descarga por ficha)**: la descarga sigue pasando por la fábrica única (ahora con el 443 exigido); `UpdateMediaItemCoverIfNull` pone la condición en el propio `WHERE` (`cover_file IS NULL`), de modo que dos búsquedas simultáneas no se pisan ni borran archivos ajenos; la elección de una carátula por ficha (la del video) reduce descargas y escrituras. Menos superficie, no más.
- **`779f82c` (alta atómica)**: `InsertMediaItemWithSource` y `MarkDuplicatePair` en una sola transacción que se revierte al conflictuar; el `INSERT` escribe `duplicate_of_id` como parámetro `narg`. Todo parametrizado por sqlc; mejora de integridad, sin SQL dinámico.
- **`55ec8a7` (pareja recíproca)**: `candidato()` descarta fichas ya emparejadas y la marca escribe con `duplicate_of_id IS NULL` dentro de transacción. Cierra huecos de integridad; nada de seguridad tocado.
- **`54e45d5` (resolver el par en una transacción)**: `ResolveDuplicatePair` escribe las dos filas **y la auditoría dentro de la misma transacción** (`insertActions` en `withTx`): la regla P9-14 (mutación + auditoría atómicas) no solo se mantiene, se refuerza — desaparece la ventana en la que la pareja quedaba cambiada sin registro.
- **`c860a45` (escritura por columnas)**: `UpdateMediaItemAvailability` toca solo el eje de disponibilidad y `visibility = COALESCE(narg, visibility)` (solo la indisponibilidad **confirmada** impone visibilidad); `UpdateMediaItemJoin` toca solo `title/published_at/origin`. **No permite saltarse ninguna comprobación**: la validación de dominio sigue aplicándose en el service y la ficha se relee tras las comprobaciones antes de decidir la marca; el alcance de escritura del sync se **reduce** (la descripción, su marca de edición, la carátula y el eje de duplicados del equipo quedan fuera). Ambas por `WHERE id =` parametrizado.
- **`c5e7a0a` (procedencia por lote + código de auditoría del PATCH)**: `ListMediaItemSourcesByItems` es `item_id = ANY($1::uuid[])` parametrizado y **los ids nunca vienen de la petición**: salen de las filas de la página que el propio servidor acaba de leer (`fuentesDeLaPagina`), tipadas `[]uuid.UUID`. El código de auditoría del `PATCH` se resuelve **después** de decodificar el cuerpo con el mismo criterio del camino de éxito (`media.description.update` / `media.hide` / `media.show`), corrigiendo la imprecisión del rechazo.
- **`0f778a5` (tiempo de espera propio de 120 s)**: es un tiempo de espera **del navegador del equipo**, por encima del presupuesto del servidor (90 s) para no abortar a media importación. **No amplía la ventana de trabajo del servidor**: el `POST /buscar-nuevos` sigue sosteniendo como máximo 90 s con sus constantes (30 elementos/30 comprobaciones/30 carátulas por búsqueda, 4 en vuelo, 10/5 s por petición saliente, cuerpos ≤2 MiB/64 KiB/8 MiB) y sigue tras sesión + permiso `medios` + CSRF. Sin nuevo vector de abuso.
- **`05e3a4b` (contar runas)**: validación del frontend alineada con la del servidor; sin superficie nueva.
- **`f393ddb` (aviso de truncado + `findMedio`)**: la pareja se busca recorriendo las páginas del listado contra el servidor (GET autorizados, ids que ya están en el cliente); el aviso del truncado es UI. Sin entrada nueva sin validar (ver I2).

## Informativos (no requieren cambio para aprobar)

- **I1 (nota de entorno — atender antes de la puerta 3)**: el stack compartido del compose (`:8080`) corre una imagen **construida a las 19:43 CST, anterior a los 11 commits del ciclo 1**: contra él, `?offset=2147483648` **todavía responde 500** (verificado en vivo). No es un defecto del código de la rama — contra el binario de `c65e67b` responde 200 acotado (tabla de arriba) — sino un stack local sin reconstruir. El equipo debe reconstruir (`docker compose up -d --build`) antes de los e2e de la puerta 3 para que las verificaciones locales se hagan contra el binario corregido.
- **I2 (vigilable)**: `findMedio` recorre las páginas del listado (una por cada 100 fichas del catálogo) al pulsar «Revisar duplicado» si la pareja no está en la primera: son GET autorizados equivalentes a paginar a mano, acotados por el tope del servidor por petición; con un catálogo muy grande conviene vigilar el volumen (mismo espíritu que I4 del informe anterior).
- **I3 (constancia)**: `govulncheck ./...`: **0 vulnerabilidades que afecten al código** (1 informativa en módulo requerido que el código no llama — idéntica a la del ciclo 1); `npm audit --audit-level=high`: **0**; `go vet ./internal/medios/... ./cmd/api/...`: ok. Greps sobre el diff de corrección y sobre el árbol: sin `dangerouslySetInnerHTML`/`innerHTML`/`eval`/`document.write`/`InsecureSkipVerify`/`http.Client` nuevos en producción (las dos apariciones de `http.get` son dobles MSW de **tests** del frontend), sin SQL construido por concatenación/`Sprintf` en producción (0), sin secretos nuevos (`.env.example` y `package.json` sin cambios en el diff de corrección).

---

## Pruebas ejecutadas por el revisor

- `go test -count=1 ./internal/medios/... ./cmd/api/...`: **ok** (0,43 s / 0,08 s). Tabla compartida de barreras con el caso de puerto 443, en verde por los cuatro consumidores: `TestClienteSalienteBarreras`, `TestLectorFeedsBarreras`, `TestProbadorBarreras`, `TestFetchCoverDescargadorBarreras`.
- `govulncheck ./...`: 0 alcanzables (1 informativa no llamada). `npm audit --audit-level=high`: 0. `go vet`: ok.
- **En vivo, con el binario de HEAD compilado por el revisor** (instancia propia en `:8091`, misma base de datos, GET públicos de lectura — sin tocar el stack compartido ni disparar `buscar-nuevos`): la batería de B1 de la tabla de arriba + controles (`no-store`, `lang=fr` → 400, panel espejo sin sesión → 401). La instancia se detuvo y eliminó al terminar.
- Greps: secretos y patrones peligrosos sobre `8ee1fcf..f393ddb` (0 relevantes), SQL dinámico en producción (0), `http.Client`/dialers fuera de la fábrica (0).

## Nota de alcance

- El revisor **no modificó código ni artefactos del proyecto**: el único archivo escrito es este informe (el binario de verificación se compiló y borró en `/tmp/opencode`; la instancia de verificación fue aparte, en `:8091`, y ya está detenida).
- No se disparó `POST /buscar-nuevos` con sesión (mutaría la base compartida y saldría a internet): las barreras de red quedaron fijadas por la tabla compartida sobre los tres consumidores reales, reejecutada en verde por el revisor.
