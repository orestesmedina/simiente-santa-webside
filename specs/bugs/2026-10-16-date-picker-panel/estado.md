# Estado: 2026-10-16-date-picker-panel

<!-- Bug: los campos de fecha/hora del panel no muestran su selector nativo. -->

## Resumen

| Campo | Valor |
|---|---|
| Rama | fix/date-picker-panel |
| Flujo | equipo-bug |
| Fase | 1/5 · Diagnóstico cerrado (paso 2: prueba que falla) |
| Ciclo de corrección | 0/3 |
| Próximo paso | `qa-tester` escribe una prueba que falle; luego `dev-frontend` aplica el arreglo |
| Bloqueado por | — |
| Actualizado | 2026-10-16 |

## Aprobaciones

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | n/a (bug) | | | |
| Plan | n/a (bug) | | | |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

- **Bug:** en el panel (p. ej. crear evento), los campos de fecha/hora se ven como un input de texto y obligan a escribir la fecha a mano; el usuario recuerda que antes había un *date picker*.

## Diagnóstico (causa raíz)

- Los campos **siguen siendo nativos** `<input type="date">` / `<input type="time">` (decisión de UX **D-4** de F4; el código no cambió desde F4). Verificado en el bundle desplegado: hay `type:"date"` (4) y `type:"time"` (6).
- **F10** introdujo el modo oscuro y fija `color-scheme` **globalmente en `<html>`** (`index.html` anti-FOUC y `ThemeProvider`).
- El **panel es siempre claro** (`bg-white` en sus layouts) pero **no resetea `color-scheme`**; los `input` son transparentes (preflight de Tailwind), así que se ven blancos.
- Con el tema público en **oscuro**, el navegador pinta el **indicador nativo del selector de fecha/hora en color claro** sobre el fondo blanco del panel → **invisible**. El texto se lee porque lleva color explícito. Resultado: parece un input de texto y hay que teclear la fecha.
- **Arreglo previsto:** forzar `color-scheme: light` en los contenedores del panel (`AppLayout`, `PanelLayout`, `PanelSurfaceLayout`), sin tocar el sitio público. Cubrirlo con una prueba que falle antes del arreglo.

## Decisiones y aclaraciones

- 2026-10-16 (humano): reportado como bug; "solo sale un input, hay que ingresar la fecha manualmente". Pide que vuelva a ser un *date picker*. Aprobado arreglarlo.

## Bitácora

- 2026-10-16 — Diagnóstico cerrado; rama `fix/date-picker-panel` creada.
