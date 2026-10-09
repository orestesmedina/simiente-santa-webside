# Revisión de código — commit b0cefda «fix(infra): fija Go 1.27.2»

**Rama:** `002-acceso-gestion-usuarios` · **Fecha:** 2026-10-08 · **Revisor:** `revisor-codigo`
**Alcance:** solo el commit `b0cefda8ab8f1b9679076b0af38d48cbb8b9dc30` (el resto de F2 ya está revisado en `revision-2026-10-05-codigo.md`). Verificado con `git show b0cefda --stat` y `git show b0cefda`.

## Veredicto: APROBADO CON OBSERVACIONES

El cambio es exactamente el mínimo prometido (4 archivos, 9 inserciones / 5 borrados), no toca el kit, no añade dependencias ni código de producción, y la palanca elegida (`backend/go.mod`) es la correcta. El único hallazgo es de exactitud documental en la nota añadida a D-A5.

## Puntos correctos (contrastados)

- **Cambio mínimo, sin desviación de alcance.** `git show b0cefda --stat`: solo `README.md` (+2/−2), `backend/Dockerfile` (+2/−2), `backend/go.mod` (+1/−1), `docs/tecnico/decisiones.md` (+4/−0). `go.sum` intacto (sin dependencias nuevas), cero archivos de código de producción.
- **Ningún archivo del kit tocado.** Contrastado con `.kit-manifest.json`: el diff no incluye `.github/workflows/ci.yml`, `.specify/memory/constitution.md`, `Makefile`, `AGENTS.md` ni `.agents/skills/**` (ni ningún otro de los 37 archivos del manifiesto). Regla 10 respetada.
- **La palanca correcta es el `go.mod`.** `.github/workflows/ci.yml:97` usa `go-version-file: backend/go.mod`; con `go 1.27.2` en la directiva `go`, `setup-go` instala 1.27.2 (o superior de la rama) sin tocar el CI del kit. Un `toolchain go1.27.2` solo no habría movido la instalación del CI.
- **Pin exacto y reproducible.** `backend/Dockerfile:11` — `FROM golang:1.27.2 AS build`: parche fijado, sin `:latest`, coherente con D-A5 ("Flotar con `golang:latest`: rechazado — builds no reproducibles"). El comentario del Dockerfile (línea 3) se actualizó en el mismo commit, sin dejar versión obsoleta.
- **Coherencia README ↔ realidad.** `README.md:171` (tabla de herramientas) y `README.md:194` (tabla de soporte) reflejan `1.27.2` y mantienen la rama 1.27 como la decisión de fondo ("parche de la rama 1.27"), sin contradecir D-A5.
- **Reclamo de `govulncheck` verificado.** Ejecutado `govulncheck ./...` (v1.8.0) sobre el árbol actual: sale en verde, 0 vulnerabilidades en el código y en los paquetes importados. La afirmación de la nota es cierta.
- **Semántica Go correcta.** `go 1.27.2` como directiva `go` es válida (patch permitido desde Go 1.21) y sube el mínimo del toolchain de forma intencional. Comprobado en la máquina: fuera del módulo el toolchain local es go1.27.1, pero dentro de `backend/` el comando `go` conmuta a **go1.27.2** (`GOTOOLCHAIN=auto`), así que builds/tests locales ya corren sobre el parche.
- **Sin huellas de `golang:1.27` (sin parche) operativas fuera de `specs/`.** Búsqueda global de `golang:1.27`: fuera de artefactos históricos de `specs/` solo queda el cuerpo histórico de D-A5 (`docs/tecnico/decisiones.md:152`), que la nota del 2026-10-08 de la misma sección deja claramente superado.

## Hallazgos

### Bloqueantes
Ninguno.

### Importantes

- **I1 — La nota D-A5 describe mal el rango de advisories (exactitud documental).** `docs/tecnico/decisiones.md:165-168` dice «las 9 vulnerabilidades de la stdlib (GO-2026-6603…GO-2026-6617, corregidas en go1.27.2)». Consultada la base `vuln.go.dev` ID por ID (2026-10-08): el rango tiene **15 IDs**, de los cuales **no existen** GO-2026-6606 y GO-2026-6614 (404); GO-2026-6615 y GO-2026-6616 son de **módulos OpenTelemetry** (no stdlib y no corregidas por go1.27.2, sino por versiones de esos módulos); y los **11** restantes (6603, 6604, 6605, 6607–6613, 6617) sí son stdlib corregidos en 1.27.2. Ni «9» ni «todo el rango corregido en go1.27.2» se sostienen. **Propuesta:** reformular la nota con los IDs exactos, p. ej. «las 11 vulnerabilidades de la stdlib del rango GO-2026-6603…GO-2026-6617 (sin contar GO-2026-6606/6614, no publicadas, ni GO-2026-6615/6616, de OpenTelemetry)», o citar el número exacto que arrojó el `govulncheck` previo con la fecha del escaneo. No afecta al código ni al pin; es la trazabilidad de la auditoría de seguridad lo que queda impreciso.

### Sugerencias (no bloqueantes)

- **S1 — Cuerpo histórico de D-A5 con pins obsoletos a simple vista.** `docs/tecnico/decisiones.md:152-153` aún dice literalmente «imagen `golang:1.27` en `backend/Dockerfile` y `go 1.27` en `backend/go.mod`». La nota del 2026-10-08 lo supera, pero un lector que copie la sección «Decisión» sin leer las consecuencias se lleva los pins viejos. Un «(pins actualizados a 1.27.2: ver nota 2026-10-08)» al final del punto lo resolvería sin reescribir la historia.
- **S2 — `make doctor` reporta go1.27.1 mientras el README fija 1.27.2.** El toolchain instalado en la máquina es 1.27.1 (`scripts/doctor.sh:32`, archivo del kit, solo exige mínimo 1.23 → sin conflicto); `README.md:168` dice «Verificables con `make doctor`». Con `GOTOOLCHAIN=auto` los builds ya usan 1.27.2, pero convendría actualizar el Go local a 1.27.2 o añadir una nota al pie de la tabla explicando que el toolchain local puede ir por detrás y que `go` conmuta automáticamente.
- **S3 — `govulncheck` reporta 1 vulnerabilidad en módulos requeridos sin llamadas alcanzables** (detalle que no contradice el «en verde»). Interesante dejar constancia del `govulncheck -show verbose` correspondiente cuando se cierre F2, para que la auditoría quede completa.
- **S4 — `CHANGELOG.md:55`** («Toolchain fijada: Go 1.27…») es una entrada histórica de F1 y no era alcance de este commit; al cerrar F2, las notas de entrega deberían mencionar el pin 1.27.2 (o añadir una entrada `fix(infra)` en «Unreleased»).
- **S5 — Digest en el FROM.** `golang:1.27.2` fija major.minor.patch pero la etiqueta sigue siendo mutable (el hallazgo de `revision-2026-10-03-seguridad.md` sobre tags flotantes aplica en menor medida). Fijar el digest (`golang:1.27.2@sha256:…`) quedaría alineado con el «minor/digest exacto» que el propio `README.md:196-197` declara pendiente para las demás imágenes.

## Constitución (`.specify/memory/constitution.md`)

Ninguna regla queda infringida por este cambio:
- **§I** — no cambia comportamiento; es un pin de toolchain coherente con lo ya decidido (D-A5/D5).
- **§II** — cero dependencias nuevas (`go.sum` intacto).
- **§III–VI** — sin código, ni migraciones, ni interfaz; las pruebas existentes no se tocan.
- **§VII** — la configuración sigue por variables de entorno y `docker compose up` funciona igual; el pin endurece la reproducibilidad del build.
- **§VIII** — es un `fix(infra)` sobre la rama de F2; el merge sigue requiriendo aprobación humana.

## Recomendaciones de cierre

1. Corregir **I1** (reformular el conteo/IDs de la nota D-A5) — cambio documental de 1 línea, sin tocar el pin.
2. S1–S2 junto con la revisión documental de cierre de F2; S4 con el `documentador` en la fase de entrega; S5 queda como parte del pendiente ya registrado de digest para imágenes.

---

## Nota de cierre — 2026-10-08 (sesión de continuación)

**I1 (Importante, documental) → RESUELTO en `a2ab81a`** (`docs(tecnico): precisa los IDs de las CVEs de la stdlib en la nota de D-A5`). Verificado con `git show a2ab81a` (solo `docs/tecnico/decisiones.md`, +4/−3):

- La nota ya **no** usa el rango engañoso `GO-2026-6603…GO-2026-6617` ni afirma que todos los IDs del rango son stdlib corregidos en go1.27.2.
- Cita los **9 IDs exactos** del escaneo —GO-2026-6603, -6605, -6607, -6608, -6610, -6611, -6612, -6613 y -6617— y los **acota a su fuente**: «las 9 vulnerabilidades de la stdlib que `govulncheck` v1.8.0 reportó con go1.27.1 y que go1.27.2 corrige». Es exactamente la propuesta de corrección de I1 (IDs exactos + atribución al escaneo con fecha/toolchain).
- Contrastado ID por ID contra `vuln.go.dev` (2026-10-08): los 9 existen, son `stdlib` y están corregidos en 1.27.2; las áreas citadas (net/http, HTTP/2, net/textproto, crypto/tls) coinciden con sus resúmenes. La acotación «que `govulncheck` reportó» es además necesaria: go1.27.1 tiene **11** advisories de stdlib corregidos en 1.27.2 (faltan de la lista GO-2026-6604 y GO-2026-6609, no alcanzables/no aplicables según el escaneo), y el texto nuevo ya no contradice ese hecho.

**`bf999e7` verificado** (`docs(entrega): registra el arreglo del CI (Go 1.27.2)`, solo `CHANGELOG.md`, +1):

- Único bullet nuevo en `[0.2.0]` → `### Corregido` (tras «Rutas con parámetros» y «`Retry-After`»), antes de «No entra en esta versión». `[0.1.0]` y el resto del archivo intactos.
- Redacción veraz en lo verificable: `backend/go.mod:1` = `go 1.27.2` ✓; `backend/Dockerfile:11` = `FROM golang:1.27.2 AS build` ✓; el CI instala Go con `go-version-file: backend/go.mod` (`.github/workflows/ci.yml:97`) ✓; `govulncheck` v1.8.0 en verde sobre el árbol actual ✓ (re-ejecutado). Los 9 IDs y áreas coinciden con la nota de D-A5.
- No duplica indebidamente D-A5: el bullet es el registro de entrega (qué estaba roto y cómo se corrigió) y es autocontenido; D-A5 sigue siendo el registro de decisión.
- **Límites de verificación (no bloqueantes):** (a) «el CI … resolvía go1.27.1» es coherente con el fallo observado y con el toolchain local (go1.27.1), pero no verificable desde este entorno (sin acceso a los logs de Actions); el humano puede confirmarlo con el enlace del run al persistir `estado.md`. (b) Micro-redacción: «cuya stdlib tiene 9 vulnerabilidades corregidas en go1.27.2» — estrictamente el stdlib de go1.27.1 tiene 11 advisories corregidos en 1.27.2; los 9 son los que el escaneo reportó como alcanzables (D-A5 lo formula ya bien: «que `govulncheck` … reportó»). Alinear el bullet con esa misma fórmula lo dejaría exacto.

**Kit y código de producción:** ninguno de los dos commits toca archivos de `.kit-manifest.json` (`a2ab81a` → solo `docs/tecnico/decisiones.md`; `bf999e7` → solo `CHANGELOG.md`, ninguno de los dos está en el manifiesto) ni código de producción.

**Veredicto final del arreglo completo (`b0cefda` + `a2ab81a` + `bf999e7`): APROBADO.** El único hallazgo «Importante» (I1) queda cerrado; permanecen como sugerencias no bloqueantes S1 (puntero «ver nota 2026-10-08» en el cuerpo histórico de D-A5, `decisiones.md:152-153`), S2 (`make doctor` reporta go1.27.1 vs README 1.27.2), S3 (constancia del `govulncheck -show verbose` al cerrar F2), S5 (digest para `golang:1.27.2`) y la micro-redacción del bullet de CHANGELOG apuntada arriba.
