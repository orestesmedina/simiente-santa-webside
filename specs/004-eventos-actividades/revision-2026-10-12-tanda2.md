# Revisión consolidada (tanda 2) — F4 correcciones (2026-10-12)

Alcance: commits `9aef868` (reorden del bloque de portada), `2a70591` (fondo claro fuera del layout del panel), `f179daa` (e2e de regresión + `ux.md`), `310cd1c` (e2e `/cambiar-contrasena`). Reportes:
`revision-2026-10-12-{qa,codigo,seguridad}-tanda2.md`.

## Veredicto global

**APROBADO** — los tres roles aprueban.

| Rol | Veredicto | Nota |
|---|---|---|
| `qa-tester` | APROBADO | 46 e2e en 4 tandas, 530 unitarias frontend y suites Go (unit+integración, cobertura 87,0 %) en verde |
| `revisor-codigo` | APROBADO | arreglo del login correcto y sin tocar el sitio público; reorden cumple D-24; 1 IMPORTANTE (documentación) + 5 sugerencias |
| `seguridad` | APROBADO | solo presentación/orden de componentes; autorización intacta; sin secretos ni hallazgos |

## Cambios

- **Bug (login en negro)**: pantallas **fuera** de `PanelLayout`/`AppLayout` (`/login`, `/cambiar-contrasena`, `/sin-permiso`, 404) ahora se envuelven en `PanelSurfaceLayout` (`bg-white`, **sin** depender del tema). Causa raíz: el anti-FOUC de F10 fija `color-scheme: dark` en `<html>` y esas rutas no declaraban fondo (hueco previo de F10, no regresión de F4). Sin fondo global en `body` para no provocar destello en la portada oscura. Prueba de regresión con `colorScheme: 'dark'` en las 4 pantallas.
- **Ajuste (Q8)**: el bloque «Lo que viene» pasa a ser la **primera sección tras el hero**, **antes de «quiénes somos»** (D-24 confirmada por el humano el 2026-10-12).

## Hallazgos corregidos / pendientes

- **IMPORTANTE (documentación)** — `plan.md` P4-19 contradecía D-24: **ya alineado** (bloque tras el hero, antes de «quiénes somos»).
- **Sugerencias (deuda menor, no bloquean)**: aserción de posición de WhatsApp retirada en `homePage.test.tsx`; selector muerto `#root > div` en `layout.test.tsx`; token `--color-panel-ink` fuera del sistema F10 y aplicado a medias junto a `text-slate-900/700`; posible destello previo al montaje en `/login` con tema oscuro.

## Pendientes manuales (heredados)

SC-010 (usabilidad con ≥6 personas) y lectura con lector de pantalla físico.

## Próximo paso

CHANGELOG (Corregido + Cambiado), actualizar `pr.md` y abrir el PR (lo hace el humano).
