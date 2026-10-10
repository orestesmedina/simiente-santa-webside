# QA — Validación del bug «imágenes de la portada no cargan» (ciclo 4, 2026-10-10)

**Rama:** `003-portada-info-general` · HEAD validado: `ec2ae76` (fix) sobre `b257e0b` (regresión). Árbol limpio al cerrar la sesión (sin cambios pendientes).
**Objetos validados:** prueba de regresión `frontend/e2e/portada-imagenes.spec.ts` (incluidos los ajustes del `dev-frontend` al helper `browserBase` y al locator del logo) y fix `mediaUrl()` en `src/api/client.ts` + usos en `IdentityHero`/`PublicLayout`/`PublicFooter`.

---

## Veredicto: **APROBADO**

La prueba de regresión es **válida** (falla sin el fix, pasa con él, sin aserciones debilitadas), el fix resuelve el criterio afectado y el resto de la suite queda en verde (e2e 8/8 y unitarias 277/277). Queda **un hallazgo menor** de estabilidad del entorno e2e (interacción con el control de seguridad R12), que no corresponde al fix y no bloquea.

---

## 1. La regresión discrimina el bug (evidencia roja/verde real)

Método: revertí **solo los 4 archivos de producción** del fix (`client.ts`, `IdentityHero.tsx`, `PublicLayout.tsx`, `PublicFooter.tsx` a `ec2ae76^`), **manteniendo intacta la prueba ajustada**, reconstruí la imagen del frontend (`docker compose build frontend`) y volví a ejecutar la regresión contra el stack real. Luego restauré el fix desde HEAD, reconstruí de nuevo y repetí. Sin dejarme cambios (`git checkout HEAD --` tras cada fase; `git status` limpio).

| Escenario | Resultado | Evidencia |
|---|---|---|
| Fix **activo** (HEAD `ec2ae76`, imagen reconstruida) | ✅ **PASS** — 1 test passed | `/tmp/opencode/qa-evidencia/verde-con-fix.txt` |
| Fix **revertido** (producción sin `mediaUrl`, prueba nueva intacta) | ❌ **FAIL** — el navegador nunca pidió `http://localhost:8080/api/v1/media/img_….jpg` (punto exacto del bug) | `/tmp/opencode/qa-evidencia/rojo-sin-fix.txt` |
| Fix **activo** de nuevo (restaurado desde HEAD, imagen reconstruida) | ✅ **PASS** | `verde-con-fix.txt` (re-ejecutado) |

El fallo en rojo se produjo en la aserción de red (a) — `expect.poll(… requestsPorUrl…).toBeDefined()` con timeout: sin `mediaUrl` el `src` se resolvía contra `:5173` y la petición a `:8080` no ocurría nunca. Es decir, la prueba **no pasa en falso**: distingue ambas implementaciones.

## 2. ¿Se debilitó la prueba al ajustarla? **No.**

Comparación de la versión de `b257e0b` frente a la ajustada en `ec2ae76`:

- **`browserBase` ahora resuelve contra el origen de la API (`API_URL`)** en lugar del origen de la SPA (`:5173`). Es ajuste obligatorio: con el fix el `src` apunta al origen de la API, y precisamente eso es lo que la regresión debe verificar. La comprobación en rojo del punto 1 demuestra que el ajuste no escondió el bug: sigue fallando sin el fix. Además la prueba mantiene el **control negativo** (la media servida directamente por el backend responde `200 image/*`), así que un fallo del siembra no puede confundirse con un fallo de la SPA.
- **Aserciones conservadas:** por logo y por portada, `status = 200` + `content-type ≅ image/*` de la respuesta de red real del navegador, y `naturalWidth > 0` con `expect.poll` (descodificación real).
- **Locator acotado al hero** (`page.getByRole('region', { name: identityName })`): es **más estricto**, no menos. Con `getByRole('img', …)` a nivel de página, el `alt` del logo repetido en cabecera/hero/pie provoca violación de *strict mode* (excepción, no aserción). Acotar fija el sujeto real del bug (las imágenes del hero) que la prueba verifica por `naturalWidth`.
- **Unitarias del fix (añaden cobertura):** `client.test.ts` cubre `mediaUrl` en 4 casos (relativa, absoluta http/https/data, protocolo-relativa `//host`, `undefined`/`''`); `homePage.test.tsx` **añade** `toHaveAttribute('src', \`${API_BASE_URL}/…\`)` al logo y a la portada (antes solo existían). No se eliminó ninguna aserción.

## 3. Suite completa

- **e2e completa** (`npx playwright test --config e2e/playwright.config.ts --workers=1`): **8/8 passed (27.0 s)** en la ejecución limpia final — incluida `portada-imagenes.spec.ts` y todo el resto intacto (acceso, auditoría, accesibilidad, panel, patch-null, portada pública, status). Evidencia: `/tmp/opencode/qa-evidencia/suite-e2e-completa.txt`.
- **Unitarias de frontend** (`npm test -- --run`): **45 archivos / 277 tests passed**, sin SC roto.
- El fix no toca backend, contrato ni nginx; no procede re-ejecutar Go por este bug (se mantiene el `make ci` verde reportado en la fase de entrega, sin deriva en la rama).

## 4. Hallazgos

### H1 — Menor (estabilidad del entorno e2e, no del fix): la suite completa roza el límite 20/min por IP del rate-limit (R12) y puede fallar intermitentemente en `portada-publica`

- **Descubrimiento durante la validación:** dos ejecuciones de la suite completa fallaron 7/8, siempre en `portada-publica.spec.ts` (el último spec que hace login por orden alfabético) con `429 rate_limited` en el login del siembra.
- **Causa raíz (verificada con evidencia, no conjetura):**
  - El backend aplica un limitador por IP **en memoria** de 20 peticiones/min sobre `POST /api/v1/auth/login` y `/setup/initialize` (`backend/internal/platform/middleware/ratelimit.go`, P17/R12, ventana deslizante de 1 min; los 429 anunciaban «quedan 1 minutos» en la UI).
  - Toda la suite comparte la IP del host (`172.19.0.1`) y acumula **~20 logins en <35 s**, justo el umbral. Los logs del backend muestran `POST /auth/login → 429 code="rate_limited"` en cada ejecución fallida (20:40:35, 20:49:38…).
  - El bloqueo FR-006 **no** interviene: la BD registra **0 fallos de login** para `ana@ejemplo.com` en 3 h (el bloqueo/desbloqueo de `carlos` y `auditor` que sí aparece es deliberado de `acceso.spec.ts` y `auditoria.spec.ts`). Un arranque limpio de Redis (`login:*`) más la espera de la ventana deslizante eliminó el síntoma.
- **Pasos para reproducir:** ejecutar la suite completa mientras otra actividad e2e (otra suite en paralelo, o re-ejecuciones manuales en los 60 s previos) acumula hits sobre el mismo IP+ruta; el spec que loguea en último lugar recibe 429 en su siembra. Aislada o con ventana limpia pasa.
- **Resultado esperado:** verde siempre · **Obtenido:** rojo intermitente por interferencia con un control de seguridad, no por defecto del fix ni de los criterios de F3.
- **Sugerencia (fuera de esta rama, decide el dueño del harness):** reducir logins reutilizando sesión `storageState` compartida entre specs, o reordenar/reintentar; o documentar en el harness que la suite requiere 60 s de respiro entre ejecuciones. Nota independiente del mismo hallazgo: con la nueva prueba el total queda clavado **en exactamente 20 = límite**, margen cero; cualquier login adicional futuro romperá siempre la suite.
- **Severidad:** menor para este PR (no afecta al criterio del bug; la suite en verde se alcanza en ejecución limpia), pero merece seguimiento propio: es una bomba de flakiness para CI.

### Nota de entorno durante la validación
Por momentos, mis re-ejecuciones y las de agentes en paralelo compartieron el backend/Redis (stack único): por eso conviene seguir la recomendación de H1 antes de fiar la suite a ejecuciones concurrentes.

---

## Matriz de cobertura del hallazgo

| Criterio afectado (bug) | Prueba | Estado |
|---|---|---|
| SC-F3: las imágenes publicadas del hero cargan en la portada pública | `e2e/portada-imagenes.spec.ts` | ✅ PASS con fix · ❌ FAIL (probado) sin fix |
| `src` resuelto contra la base de la API | `homePage.test.tsx` (nueva aserción de atributo) + e2e red | ✅ |
| Regresión sin debilitar aserciones | comparación b257e0b ↔ ec2ae76 | ✅ (mantiene red/CT/naturalWidth; control negativo intacto) |
| Resto de la suite sin roturas | e2e 8/8 · vitest 45/277 | ✅ |
