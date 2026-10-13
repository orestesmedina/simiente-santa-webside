# Revisión de seguridad — F4: re-auditoría, tanda 2 (fondo claro + reorden del bloque de portada)

- **Rama**: `004-eventos-actividades` · **Rango revisado**: `3955864..f179daa` (commits `9aef868` reorden del bloque «Lo que viene», `2a70591` fondo claro fuera de los layouts del panel, `f179daa` e2e de regresión + ux.md)
- **Fecha**: 2026-10-12 · **Revisor**: `seguridad` (segunda tanda de re-validación)
- **Referencias**: `revision-2026-10-12-seguridad.md` (tanda 1, APROBADO), `revision-2026-10-11-seguridad-ciclo2.md` (APROBADO), `.specify/memory/constitution.md` §IV.
- **Método**: lectura completa del diff de los 3 commits y de los archivos vigentes (`layout.tsx`, `router.tsx`, `HomePage.tsx`, páginas de auth, `index.css`), grep del rango por `dangerouslySetInnerHTML|innerHTML|eval|password|secret|token|api_key|localStorage|document.cookie|authorization`, y `govulncheck ./...` + `npm audit --audit-level=high`.

## Veredicto: **APROBADO**

Sin hallazgos de ninguna severidad. Los tres commits son **presentación y orden de componentes**: no tocan backend, contratos, SQL ni lógica de autorización; no exponen datos nuevos; no introducen secretos.

---

## Verificación por commit

### `9aef868` — reorden del bloque «Lo que viene» — ✅ sin impacto

- `HomePage.tsx`: único cambio es mover el `<Reveal key="proximamente">` existente (mismo componente `ProximamenteBlock`, misma data, mismo `from="right"`) de después de «Horario» a antes de «Quiénes somos». Consumo de datos, hooks y render idénticos.
- `homePage.test.tsx`: solo cambian las aserciones de orden (`compareDocumentPosition`).
- Nada de datos, autenticación ni superficie de ataque.

### `2a70591` — fondo claro en `/login`, `/cambiar-contrasena`, `/sin-permiso` y 404 — ✅ sin impacto

- **Autorización intacta (CWE-862)**: el único cambio en `router.tsx` es envolver rutas en el nuevo layout. Verificado en HEAD (`router.tsx:85-146`): `/cambiar-contrasena` sigue **dentro** de `RequireAuth` (`:92-98`, el guard queda por fuera del layout); `/login` y `/sin-permiso` siguen públicas (como antes); el 404 (`*`) sigue siendo catch-all público; todos los `RequirePermission` de `/panel/*` (usuarios/roles/auditoría/portada/eventos/actividades/categorías) quedan intactos. No se añade ni se quita ningún guard.
- **Nuevo componente `PanelSurfaceLayout`** (`layout.tsx:47-53`): puramente presentacional, un `div` con clases (`bg-white text-panel-ink`) y `{children ?? <Outlet />}`; `children` es `ReactNode` interno del router, nunca entrada de usuario. Sin `dangerouslySetInnerHTML` (grep vacío), sin `href` nuevos.
- **`index.css`**: añade el token `--color-panel-ink: #0f172a` (un valor de color sRGB, no un secreto).
- **Páginas de auth** (`LoginPage.tsx`, `ChangePasswordPage.tsx`): solo cambio de clase `text-slate-900` → `text-panel-ink` en los `<h1>`. Cero cambios en formularios, manejo de sesión, mensajes de error ni peticiones.
- **Tests** (`layout.test.tsx`): aserciones de clases/contenido.

### `f179daa` — e2e de regresión + ux.md — ✅ sin impacto

- `qa-bug-fondo-oscuro-fuera-layouts.spec.ts`: prueba Playwright nueva que lee **estilos computados** (fondo/título) en las rutas públicas con `colorScheme: 'dark'`. Sin credenciales, sin login, sin datos de usuario; no modifica código de producción.
- `ux.md`: documentación de producto (decisión D-24 confirmada).

## Barridos del rango

- **Secretos (CWE-798)**: grep del rango sin coincidencias nuevas; las únicas apariciones de `password`/`token` son el import de guards, la ruta `/cambiar-contrasena`, la constante preexistente `PASSWORD_CHANGED_MESSAGE` y «tokens» del sistema de diseño en ux.md — todas benignas y preexistentes.
- **XSS (CWE-79) / inyección (CWE-89)**: grep vacío de `dangerouslySetInnerHTML|innerHTML|eval`; no hay backend ni SQL en el rango (10 archivos: 8 frontend, 2 docs de spec).
- **Dependencias**: ningún `go.mod`/`package.json` tocado; `govulncheck ./...` → 0 vulnerabilidades que afecten al código; `npm audit --audit-level=high` → 0 vulnerabilidades. Estado idéntico a la tanda 1 del mismo día.

## Nota final

El ajuste de `color-scheme`/fondos no interactúa con `HttpOnly` de cookies ni con el flujo de sesión: el script anti-FOUC de `index.html` no se modificó en este rango.
