# Revisión de seguridad — F2: parche de toolchain Go 1.27.2 (2026-10-08)

**Rama**: `002-acceso-gestion-usuarios` · **Alcance**: únicamente el commit `b0cefda` («fix(infra): fija Go 1.27.2 para cerrar las vulnerabilidades de la stdlib (GO-2026-6603…6617)») — **no se re-audita F2** (completa, auditada el 2026-10-05 en `revision-2026-10-05-seguridad.md`)
**Auditor**: agente `seguridad` (solo lectura)
**Motivo**: el CI falló en el paso «Vulnerabilidades» (`govulncheck`): 9 vulnerabilidades de la stdlib de Go **1.27.1**, corregidas en **1.27.2** (GO-2026-6603, -6605, -6607, -6608, -6610, -6611, -6612, -6613, -6617: HTTP/2, net/http, net/textproto, crypto/tls). El fix fija `go 1.27.2` en `backend/go.mod` y `golang:1.27.2` en `backend/Dockerfile`.

## Veredicto: **APROBADO**

El pin cierra las 9 vulnerabilidades detectadas (evidencia: `govulncheck` v1.8.0 corriendo contra la stdlib **go1.27.2** → exit 0, cero hallazgos en código llamado e importado); la versión fijada es el parche más reciente de la rama 1.27, la única política de soporte vigente que importa (ramas 1.26/1.27); el commit no introduce secretos ni dependencias; el residual GO-2026-5932 sigue siendo solo a nivel de módulo, sin llamadas alcanzables; y el único hallazgo de imagen (distroless `:latest`) es **preexistente** (M2 de la revisión de F1, 2026-10-03) y no fue introducido ni agravado por este commit.

---

## Evidencia (comandos y exit codes)

1. **Toolchain efectiva**: en `backend/`, `GOTOOLCHAIN=auto go env GOVERSION` → **`go1.27.2`**. Dato relevante: la toolchain instalada localmente es **1.27.1** (`go version` → `go1.27.1 linux/amd64`); con `GOTOOLCHAIN=auto` (el valor por defecto) la directiva `go 1.27.2` de `go.mod` **eleva automáticamente** la toolchain a 1.27.2. El pin no solo se cumple en el CI: cualquier entorno con una toolchain más vieja la sube solo.
2. **Escaneo**: `GOTOOLCHAIN=auto go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...` → **exit 0**. Salida: «No vulnerabilities found. Your code is affected by 0 vulnerabilities. This scan also found 0 vulnerabilities in packages you import and 1 vulnerability in modules you require, but your code doesn't appear to call these». Con `-show verbose` confirma: «Govulncheck scanned the following 13 modules **and the go1.27.2 standard library**» — el escaneo se hizo contra la stdlib parcheada, no solo con un binario nuevo.
3. **Contraste (las 9 CVEs venían de la toolchain sin parchear, y ya no se puede usar)**: forzando `GOTOOLCHAIN=go1.27.1` tanto `go build ./...` como `govulncheck ./...` → **exit 1**: «go: go.mod requires go >= 1.27.2 (running go 1.27.1; GOTOOLCHAIN=go1.27.1)». La directiva `go 1.27.2` es un **piso duro**, no una sugerencia: el módulo no se puede compilar ni escanar con 1.27.1 a menos que alguien fuerce `GOTOOLCHAIN=local` *y* edite `go.mod` — dos decisiones deliberadas y visibles en el diff.
4. **Diff del commit**: `git show b0cefda --stat` → 4 archivos (`README.md`, `backend/Dockerfile`, `backend/go.mod`, `docs/tecnico/decisiones.md`; +9/−5). Único cambio en `go.mod`: línea 1 `go 1.27` → `go 1.27.2`. **Sin cambios en `require`, `go.sum` intacto, sin secretos** (solo cadenas de versión). Sin archivos de frontend → el resultado de `npm audit` (0 hallazgos altos, 2026-10-05) no cambia.
5. **CI**: `.github/workflows/ci.yml:97` — `go-version-file: backend/go.mod` → el CI instala exactamente 1.27.2; `ci.yml:19` — `GOVULNCHECK_VERSION: v1.8.0` (mismo binario que esta auditoría); `ci.yml:113-116` — paso «Vulnerabilidades» (`govulncheck ./...`). El mismo comando que falló en el CI sale en verde localmente con la misma versión de escáner.
6. **Soporte de la rama**: go1.27.2 se publicó el **2026-10-08** (junto a go1.26.9) — es el **parche más reciente de la rama 1.27**; no existe 1.27.3 (historial de versiones de Go). Solo las ramas 1.26 y 1.27 reciben parches de seguridad.

## Comprobación por punto solicitado

1. **Las 9 CVEs quedan cerradas con el pin** — ✓. `govulncheck` v1.8.0 contra la stdlib go1.27.2: exit 0, 0 vulnerabilidades en símbolos llamados y 0 en paquetes importados (evidencia 1-3). El CI, que lee `backend/go.mod` (`ci.yml:97`), instalará 1.27.2 y repetirá exactamente este resultado.
2. **Versión fijada en soporte vigente** — ✓. Rama 1.27 (la mayor vigente; 1.26 es la otra con parches), parche 1.27.2 publicado hoy 2026-10-08: **no hay versión parche más nueva de la rama**; no existe 1.27.3. El proyecto queda en la punta de la rama con mayor ventana de soporte; no conviene (ni es posible) subir más dentro de 1.27.
3. **Imagen base del Dockerfile** — build: `FROM golang:1.27.2 AS build` (`backend/Dockerfile:11`) ✓ etiqueta con parche exacto, coincide con `go.mod`. Runtime: `FROM gcr.io/distroless/static-debian12:latest` (`backend/Dockerfile:29`): riesgo de etiqueta mutable, **hallazgo preexistente M2 de la revisión de F1 del 2026-10-03** (`specs/001-estructura-base/revision-2026-10-03-seguridad.md`), no introducido por este commit (solo tocó la etapa de build). **No es bloqueante para este PR**: es un riesgo de reproducibilidad/cadena de suministro (dos builds del mismo commit pueden diferir), no una vulnerabilidad explotable conocida, y el binario que ejecuta la imagen es estático y no toma nada de la stdlib de la base (ver punto 5). Queda abierto para `devops` con la corrección ya propuesta en M2: fijar por digest (`gcr.io/distroless/static-debian12@sha256:…`) y ajustar el comentario de `Dockerfile:8-10` («La etiqueta va fijada (nada de :latest)»), que sigue siendo inexacto para la línea 29. Nota menor heredada de M2: `golang:1.27.2` también es etiqueta y no digest — garantiza la versión de Go, pero la base Debian puede reconstruirse; mismo tratamiento (digest) cuando se atienda M2.
4. **Secretos y dependencias** — ✓. El diff no toca `require`, `go.sum`, ni ningún archivo de frontend; cero secretos (solo cadenas de versión en 4 archivos). **GO-2026-5932 residual** (`golang.org/x/crypto@v0.57.0`, paquete `openpgp` sin mantener, «Fixed in: N/A»): confirmado con `-show verbose` que aparece **solo** en la sección «Module Results» («1 vulnerability in modules you require, but your code doesn't appear to call these»); símbolos y paquetes importados limpios. El único uso de `x/crypto` en el proyecto es `bcrypt` (verificado en la revisión del 2026-10-05). **Dictamen: no bloqueante** — no hay llamada alcanzable, no existe fix publicado (N/A), no es accionable hoy (no se puede «no requerir» un paquete de un módulo que sí se necesita), y ya estaba conocido e informado el 2026-10-05. Re-evaluar cuando `x/crypto` publique una versión con remediación.
5. **Binario de producción con la stdlib parcheada** — ✓. La compilación ocurre **dentro** de la imagen `golang:1.27.2` (`Dockerfile:11`) con `CGO_ENABLED=0 GOOS=linux go build` (`Dockerfile:22-25`): net/http, crypto/tls, net/textproto y HTTP/2 son código de stdlib que se **compila dentro del binario** en ese momento. La imagen de ejecución (distroless, sin toolchain ni nada de Go) solo copia el binario (`COPY --from=build /out/api /api`, `Dockerfile:38`): no hay stdlib «en runtime» que parchear — el artefacto **lleva incorporada** la stdlib 1.27.2 con la que se compiló. Al subir el `FROM`, la garantía alcanza al binario de producción, no solo al escáner del CI. Verificado: `Dockerfile:11` fija `golang:1.27.2` (y el comentario de la línea 3 ya lo refleja). Refuerzo adicional: la directiva `go 1.27.2` de `go.mod` es un piso (evidencia 3), así que ni el CI ni un build local pueden bajar de 1.27.2 sin editar archivos.

---

## Hallazgos

### Críticos
Ninguno.

### Altos
Ninguno.

### Medios
Ninguno (en el alcance de este commit).

### Bajos / observaciones no bloqueantes

**O-1 · Runtime distroless con `:latest` (preexistente, no introducido aquí)** — `backend/Dockerfile:29`. Etiqueta mutable: riesgo de reproducibilidad y cadena de suministro (registry comprometido/repoblado entrega contenido distinto sin señal en el repo), más el comentario inexacto de `Dockerfile:8-10`. Ya reportado como **M2** en `specs/001-estructura-base/revision-2026-10-03-seguridad.md`; este commit no lo tocó (solo elevó la etapa de build, línea 11). Para este PR **no es bloqueante**; sigue abierto para `devops`: fijar por digest y corregir el comentario (idem `golang:1.27.2` y las etiquetas de las imágenes de frontend cuando se atienda en bloque).

**O-2 · GO-2026-5932 residual a nivel de módulo (`golang.org/x/crypto@v0.57.0`, `openpgp` sin mantener, sin fix disponible)** — solo «modules you require», sin llamadas alcanzables (uso real: `bcrypt`). Preexistente e informado el 2026-10-05; sin cambio en este commit. **No bloqueante**; re-evaluar con la próxima bump de `x/crypto` (responsable del seguimiento: `dev-backend`).

---

## Bloqueantes

**Ninguno.**

## Resumen

El commit hace exactamente lo que dice y nada más: fija el parche de la rama en los dos puntos que deciden la versión (D-A5 de `docs/tecnico/decisiones.md`: `go.mod` y `FROM` del Dockerfile), deja constancia en decisiones y README, y el resultado es verificable y verificado: stdlib 1.27.2 → las 9 GO-2026-66xx cerradas, sin coste en secretos ni dependencias. Los dos únicos residuos (distroless `:latest`, GO-2026-5932) son preexistentes, conocidos y no bloqueantes. El paso «Vulnerabilidades» del CI saldrá en verde.
