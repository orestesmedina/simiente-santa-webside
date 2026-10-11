# Revisión de seguridad — F10 Sistema de diseño y movimiento

**Fecha**: 2026-10-11 · **Rama**: `010-sistema-diseno-movimiento` · **Auditor**: `seguridad`
(solo lectura; este reporte es la única escritura).
**Alcance**: `git diff main...HEAD` (49 commits, `e081ec0…54cbe60`), 82 archivos
(9.739+/317−). F10 es de presentación: **todo el código es `frontend/`** (56 fuentes y pruebas
unitarias, 8 e2e, `index.html`, `index.css`) más `specs/` y `docs/`. **Sin `backend/`, sin
migraciones, sin `openapi.yaml`, sin endpoints, sin formularios** (verificado con
`git diff main...HEAD --name-only`): la afirmación de la spec de «sin superficie nueva» se
confirma.

---

## Hallazgos

### CRÍTICO — ninguno

### IMPORTANTE — ninguno

### MENOR

#### M1 · El e2e escribe la sesión de administrador (cookies `session` + `csrf_token`) en el `tmpdir` compartido, sin permisos restrictivos ni limpieza

- **Archivo**: `frontend/e2e/publicSeed.ts:55-72` (nuevo en F10; patrón inexistente en el
  `helpers.ts` de F2/F3 — `grep storageState helpers.ts` → 0).
- **Evidencia**:
  ```ts
  const stateFile = join(tmpdir(), 'simiente-e2e-f10-admin.json');   // línea 55
  …
  await context.storageState({ path: stateFile });                  // línea 67
  return browser.newContext({ storageState: stateFile });           // línea 72
  ```
  El JSON de `storageState` contiene las cookies de sesión del administrador sembrado.
  Playwright lo escribe con el modo por defecto (`0644` con umask habitual) directamente en
  `/tmp`: legible por otros usuarios del mismo equipo/runner, y **nunca se borra** entre
  corridas (la condicionante es `existsSync`, no frescura).
- **Impacto**: bajo y acotado a desarrollo/CI — son credenciales **semilla** contra el stack
  local, no de producción. En un runner compartido, una sesión viva de una corrida previa
  quedaría reutilizable por otro proceso/usuario. CWE-522/312 (exposición de datos en reposo),
  aplicado a infraestructura de pruebas.
- **Corrección sugerida**: (a) escribir el archivo con permisos restrictivos — p. ej.
  `writeFileSync(path, data, { mode: 0o600 })` — o guardarlo en un directorio del proyecto
  ignorado por git (`.playwright/.tmp/`) en vez del `tmpdir()` compartido; (b) borrarlo en un
  `globalTeardown` o al inicio de cada corrida, antes del `existsSync`; (c) opcional: nombre
  único por corrida (PID/timestamp) para no compartir sesión entre ejecuciones concurrentes.
- **No bloqueante**: no toca producción.

#### M2 · Script inline anti-FOUC: seguro hoy, pero fija el patrón «inline» que un CSP estricto futuro deberá whitelistear (hoy no hay CSP en todo el repo)

- **Archivo**: `frontend/index.html:15-35` (nuevo, T1002).
- **Evidencia**: el bloque solo lee `localStorage.getItem('ss.theme')`, acepta
  exclusivamente `'light' | 'dark'` (lista cerrada) y escribe `dataset.theme`,
  `style.colorScheme` y `classList.add('js')` — sin `document.write`, sin `innerHTML`, sin
  reflejar en HTML nada leído (verificado línea a línea; véase también §Verificaciones, punto
  11). El repo, sin embargo, **no define Content-Security-Policy en ningún punto**
  (`index.html` sin meta; sin middleware ni config de proxy) — condición preexistente, no
  introducida por F10.
- **Impacto**: ninguno explotable hoy (no hay CSP que relajar y no hay dato atacable en el
  bloque); es deuda de endurecimiento: cuando se adopte `script-src 'self'`, este bloque
  quedará bloqueado y habrá que permitirlo por hash.
- **Corrección sugerida**: documentar en el propio comentario (o en `docs/`) el hash SHA-256
  del contenido exacto entre `<script>` y `</script>`, listo para el futuro CSP. Calculado en
  esta revisión sobre `frontend/index.html`:
  `'sha256-GZeV77WoDBkmduXUY6nsyTE+oIcRVD0FGvQEGLGu5zk='`
  (alternativa: extraer el bloque a un `.js` externo bloqueante en `<head>`).
- **No bloqueante**.

### NIT

#### N1 · Credenciales semilla del e2e hardcodeadas — preexistente de F3, fuera del diff de F10

- **Archivo**: `frontend/e2e/helpers.ts:25-31` (`ana@ejemplo.com` / `Semilla.2026`). F10 las
  **reutiliza** (`publicSeed.ts:59`) pero no las añade. Solo valen contra un stack local sin
  inicializar (o sembrado por el propio e2e); el login de producción no las acepta (el
  administrador real nace del `BOOTSTRAP_TOKEN` de `.env`, no versionado). Sugerencia a futuro:
  moverlas a entorno de e2e (`.env.e2e` fuera de git) y documentarlas como semillas. **No
  cuenta contra F10.**

#### N2 · Maquetas estáticas con `<script>` inline en `specs/010…/propuestas/*.html`

- Las tres propuestas incluyen un `<script>` de demo de tema. Son artefactos de diseño **no
  publicados** (no entran en el build de Vite), sin URLs externas ni secretos (verificado).
  Solo recordatorio: no copiar/pegar sus patrones a producción cuando exista CSP. Informativo.

---

## Verificaciones sin hallazgo (evidencia)

1. **Sin superficie de servidor nueva**: diff acotado a `frontend/` + `specs/` + `docs/`; el
   «contracts» nuevo de F10 (`specs/010…/contracts/README.md`) es documentación, no código.
2. **XSS / inyección de HTML (CWE-79)**: `grep -rn` de
   `dangerouslySetInnerHTML|innerHTML|insertAdjacentHTML|document.write` en `frontend/src`,
   `frontend/e2e` y `frontend/index.html` → **0 coincidencias**. Todo el texto proveniente de
   la API se renderiza por JSX (escapado por React) — `SocialSection`, `FooterSiteMap`,
   `IdentityHero`, `ContactSection`, `ScheduleSection`, etc.
3. **Enlaces externos `target="_blank"`**: los 4 puntos del diff llevan
   `rel="noopener noreferrer"` — `PublicButton.tsx:66` (rama `external`), `FooterSiteMap.tsx:124`
   (WhatsApp) y `:143` (redes), `SocialSection.tsx:34` (patrón preexistente de F3 conservado).
   Sin «reverse tabnabbing» (CWE-1021).
4. **URLs dinámicas de la API como `href`** (riesgo de esquema `javascript:`): las URLs de
   redes y WhatsApp llegan **ya validadas por el servidor** (defensa correcta: validar en el
   emisor, no en cada consumidor nuevo): `validURL` (`backend/internal/platform/validate/validate.go:275-283`)
   exige `https` + host (rechaza `javascript:`, `data:`, `http:`), registrado como tag `url`
   (`validate.go:153`) y aplicado a `SocialLinkInput.URL` (`model.go:203`,
   `validate:"required,url,max=500"`); los grupos de WhatsApp además pasan una whitelist de
   hosts (`service.go:80` y `:295`: solo `chat.whatsapp.com` / `wa.me`) y el canal `direct` se
   construye **en servidor** como `https://wa.me/<dígitos>` (`service.go:194,210`). El nuevo
   `FooterSiteMap` no introduce, pues, ninguna URL cruda sin validar.
5. **Anclas de navegación**: `useSiteLinks.ts` genera los `href` solo como `#${section.anchor}`
   a partir de la lista estática de secciones (`sections.ts`); sin input de usuario ni enlaces
   inventados.
6. **Ejecución dinámica / apertura de ventanas**: `grep -rnE` de
   `window.open|javascript:|eval\(|new Function|location.(href|assign|replace)` en
   `frontend/src` + `frontend/e2e` → **0 coincidencias**.
7. **Secretos (CWE-798)**: grep del diff completo por
   `password|secret|api_key|token|bearer|credential|private key` → solo las semillas e2e
   preexistentes (N1) y menciones literarias de «token» en `specs/`/`docs/` (los *tokens* de
   diseño). **Sin `.env` en el diff**; `package.json` / `package-lock.json` intactos.
8. **Almacenamiento local**: `localStorage` solo guarda preferencias de presentación:
   `ss.theme` (nuevo en F10) y `ss.lang` (F3). Valores validados contra lista cerrada antes de
   usarse — `isTheme` (`ThemeProvider.tsx:12-15`) y `stored === 'light' || stored === 'dark'`
   (`index.html:20`). Sin tokens de sesión, sin PII.
9. **Dependencias**: `cd frontend && npm audit --audit-level=high` → **0 vulnerabilidades**
   (y 0 dependencias nuevas: `package.json` sin cambios). Backend sin cambios en F10 y
   `govulncheck ./...` → **0 vulnerabilidades que afecten al código** (nota informativa,
   preexistente de `main` y fuera del alcance de F10: 1 vulnerabilidad en módulos requeridos
   pero no llamados — identificar con `govulncheck -show verbose` en el siguiente pase de
   backend).
10. **Control de acceso (CWE-862/639)**: F10 no añade rutas ni endpoints; las existentes
    (portada pública sin sesión, panel con permiso `portada`) no se tocan. Los consumidores
    nuevos (`useSiteLinks`, `FooterSiteMap`, `MobileMenu`) solo **leen** de la consulta
    pública ya existente (`getPortada`, `usePublicHomeData`); el único cambio del hook es
    `placeholderData: keepPreviousData` (presentación, sin efecto de seguridad).
11. **Script anti-FOUC (`index.html:15-35`)**: revisado línea a línea — solo asignaciones de
    `dataset`/`style.colorScheme`/`classList` con constantes o el valor validado
    `'light'|'dark'`; `try/catch` silencioso; sin `document.write`; nada de lo leído se
    refleja en HTML (solo atributos DOM por API).
12. **Menú móvil / foco (`useMobileMenu.ts`)**: patrón *disclosure* de APG sin trampa de foco
    ni `role="dialog"`; escribe `dataset.navOpen` (constante) y gestiona foco con
    referencias — sin superficie de ataque; el bloqueo de scroll es puramente CSS
    (`html[data-nav-open='true']`).
13. **`index.css`**: solo `@import` de paquetes locales de npm (`tailwindcss`, `@fontsource/*`);
    sin `url()` externas, sin `@import` remoto. Los tokens de movimiento/espaciado/tema son
    valores estáticos.
14. **`mediaUrl` (`client.ts:26`, passthrough de URLs absolutas incl. `data:`)**: fuera del
    diff de F10 y ya evaluado y aceptado con sus barreras en la revisión de seguridad de F3
    (`revision-2026-10-10-bug-imagenes-seguridad.md`); F10 solo **reubica** el `src` de la
    imagen en `IdentityHero` (mismo valor `coverSrc`, mismo helper). Sin cambio de riesgo.

---

## Veredicto

**APROBADO** — 0 CRÍTICO · 0 IMPORTANTE · 2 MENOR · 2 NIT.

F10 es de presentación y cumple lo prometido en su spec (§IV Seguridad): no añade superficie de
servidor, formulario ni inyección de HTML; todos los `target="_blank"` llevan `rel`; las URLs
dinámicas están validadas en el emisor (backend) y las nuevas rutas de render solo consumen la
consulta pública existente. Los dos MENORES son, respectivamente, de infraestructura de
pruebas (M1) y de endurecimiento futuro (M2): ninguno bloquea el merge; se sugiere atenderlos
en F11+ o en una tarea de deuda técnica.
