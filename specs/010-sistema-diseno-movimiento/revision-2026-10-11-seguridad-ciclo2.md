# Revisión de seguridad — F10 · ciclo 2 (re-validación)

**Fecha**: 2026-10-11 · **Rama**: `010-sistema-diseno-movimiento` · **Auditor**: `seguridad`
(solo lectura; este reporte es la única escritura).
**Objeto**: `git diff 54cbe60..HEAD` — los **9 commits del ciclo 2**
(`cca21dd…abb28a2`), esto es, las correcciones del ciclo 1 (C1/C2/I1/IMP-1 de
QA/revisor, **M1-seguridad** mía) más sus pruebas y docs. Encargo: verificar que
M1 quedó resuelto y que las correcciones **no introducen riesgos nuevos**
(secretos, permisos, inyección).

---

## 1. M1 (sesión de admin del e2e en `tmpdir` compartido) — **CERRADO**

Fix: `00bf59f` («el storageState del e2e usa un directorio propio 0o700/0o600
y se borra al salir»), `frontend/e2e/publicSeed.ts`. Verificado el código final:

```ts
stateDir = mkdtempSync(join(tmpdir(), 'simiente-e2e-f10-'));   // línea 84
chmodSync(stateDir, 0o700);                                    // línea 85
process.once('exit', () => { rmSync(stateDir, { recursive: true, force: true }); });
…
await context.storageState({ path: stateFile });              // dentro del dir 0700
chmodSync(stateFile, 0o600);                                  // línea 71
```

Las cuatro condiciones de mi hallazgo original quedan cubiertas:

1. **Directorio propio con permisos restrictivos**: `mkdtempSync` crea el
   directorio **ya con modo `0700`** — verificado empíricamente en este entorno
   (`stat` → `40700`) — y el `chmodSync(stateDir, 0o700)` explícito lo refrenda.
   Resultado: **ningún otro usuario puede atravesar el directorio en ningún
   instante**, de modo que el momento en que Playwright escribe el archivo (con
   su modo por defecto, brevemente `0644`) **no expone nada**: la lectura
   requiere `+x` del directorio, que solo tiene el dueño. La ventana
   chmod-después-de-escribir que habría sido TOCTOU en un directorio compartido
   queda estructuralmente moot aquí; el `chmodSync(stateFile, 0o600)` es
   defensa en profundidad (p. ej. si el directorio se relajara algún día).
2. **Sin path predecible ni compartido**: el sufijo aleatorio de `mkdtemp` da un
   directorio **único por proceso del runner**: desaparece tanto la lectura
   cruzada entre usuarios como la reutilización de una sesión de una corrida
   anterior por otra (la condicionante `existsSync` ahora solo ve archivos del
   propio proceso).
3. **Limpieza**: `process.once('exit')` (registrado una sola vez, protegido por
   la guarda `if (!stateDir)`) borra el directorio completo con
   `rmSync(recursive, force)` al terminar el proceso.
4. **Sin secretos nuevos**: el único par de credenciales del e2e sigue siendo la
   semilla preexistente `ADMIN` de `helpers.ts` (N1 del ciclo 1, no introducida
   por F10 ni tocada en este ciclo).

Residual (véase N-c2-1): la limpieza solo cubre la salida «normal» del proceso.

## 2. Riesgos nuevos introducidos por las correcciones — **ninguno**

Revisado commit a commit el diff del ciclo 2 (sin contar mis/docs de revisión):

| Commit | Cambio | Evaluación de seguridad |
|---|---|---|
| `48dd0e9` (C1) | `useReveal.ts`/`Reveal.tsx`: el salvavidas de 1 s se cancela cuando el observer reporta; mapa `Element → {seen, reveal}` | Lógica de ciclo de vida (`IntersectionObserver`, `setTimeout`); **0** `innerHTML`/inyección; sin datos externos |
| `f240ea0` (C2) | `PublicButton.tsx` (1 línea de clases) + 6 líneas en `index.css` (`--accent-hover`, `--on-accent-hover` por tema) | Tokens **estáticos** (hex) y utilidades de Tailwind; sin superficie de ataque |
| `dd4c5ce` (IMP-1) | `LanguageSwitcher.tsx`: solo cambios de clases | Presentacional |
| `1339f3e` (I1) | `IdentityHero`/`HomePage`: el CTA «Ver horario» pasa a `scheduleHref` **condicional** | El `href` es la **ancla estática** `'#horario'` (literal en `HomePage.tsx:77`) o nada: sin input externo, sin URL dinámica |
| `71f9729` (M2-revisor) | `Reveal.tsx`: estado inicial `observed` si nace mostrado | Estado React; sin superficie |
| `00bf59f` (M1) | §1 | Cerrado, sin riesgos nuevos |
| `dfa03bf`/`abb28a2` | Pruebas e2e (`qa-medidas-f10.spec.ts` + medidas) y docs | La spec nueva **reutiliza** `createAdminContext`/`seedPublicPortada`; 0 credenciales nuevas (grep de `password|secret|api_key|Bearer` en el diff → solo la semilla preexistente y texto de docs) |

Barridos re-ejecutados sobre el árbol final:

- `dangerouslySetInnerHTML|innerHTML|insertAdjacentHTML|document.write|window.open|javascript:|eval(|new Function`
  en `frontend/src` + `frontend/e2e` + `index.html` → **0**.
- Los 4 `target="_blank"` siguen llevando `rel="noopener noreferrer"`
  (`PublicButton.tsx:66`, `FooterSiteMap.tsx:124/143`, `SocialSection.tsx:34`).
- **`index.html` sin cambios** en el ciclo → el hash documentado para el futuro
  CSP (M2, ciclo 1) sigue siendo válido:
  `'sha256-GZeV77WoDBkmduXUY6nsyTE+oIcRVD0FGvQEGLGu5zk='`.
- **`package.json`/`package-lock.json` sin cambios** → 0 dependencias nuevas;
  `npm audit --audit-level=high` re-ejecutado → **0 vulnerabilidades**.
- **0 cambios de backend/migraciones/API/infra** en el ciclo (diff acotado a
  `frontend/` + `specs/`).

---

## Hallazgos del ciclo 2

### CRÍTICO — ninguno · IMPORTANTE — ninguno · MENOR — ninguno (nuevo)

### NIT

#### N-c2-1 · La limpieza del `storageState` solo cubre la salida normal del proceso

- **Archivo**: `frontend/e2e/publicSeed.ts:86-92` (`process.once('exit', …)`).
- **Evidencia**: `SIGINT` (Ctrl-C local), `SIGTERM` (timeout de CI) o un cuelgue
  duro no disparan `'exit'`: el directorio `0700` con el `admin-state.json`
  queda huérfano en `/tmp` hasta que el SO lo retire (systemd-tmpfiles, ~10 días
  en distribuciones habituales).
- **Impacto**: mínimo — el directorio es `0700` (solo el dueño o root pueden
  leerlo) y la sesión caduca servidor-side; es higiene, no exposición.
- **Corrección sugerida** (opcional, deuda de pruebas): registrar también
  `SIGINT`/`SIGTERM` (o barrer al inicio los `simiente-e2e-f10-*` con mtime >
  1 h, seguro frente a corridas concurrentes por el sufijo único).
- **No bloqueante**.

### Heredados del ciclo 1 (estado)

- **M1** — **cerrado** en este ciclo (§1).
- **M2** (script inline anti-FOUC + futuro CSP) — **abierta por diseño**, no
  bloqueante, sin cambios; el hash sigue documentado en el reporte del ciclo 1.
- **N1** (semillas e2e hardcodeadas en `helpers.ts`, preexistente de F3) — sin
  cambios.
- **N2** (maquetas estáticas con script inline) — sin cambios.

---

## Veredicto

**APROBADO** — 0 CRÍTICO · 0 IMPORTANTE · 0 MENOR · 1 NIT (nuevos en el ciclo).

M1 queda **cerrado** con la solución correcta (directorio único `0700` por
proceso que hace inaccesible el archivo a otros usuarios desde su creación,
archivo `0600`, borrado al salir). Las correcciones del ciclo 2 (QA, revisor y
M1) no introducen secretos, permisos débiles ni vías de inyección: todo lo
nuevo es lógica de ciclo de vida, tokens CSS estáticos y pruebas; `npm audit`
sigue en 0 y el backend, `package.json` e `index.html` no se tocan.
