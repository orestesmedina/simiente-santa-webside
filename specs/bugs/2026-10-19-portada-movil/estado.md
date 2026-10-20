# Estado: 2026-10-19-portada-movil

<!-- Bugs: portada con tarjetas demasiado pequeñas, pantalla en blanco al navegar en móvil y menú hamburguesa transparente que no cierra al tocar fuera. -->

## Resumen

| Campo | Valor |
|---|---|
| Rama | fix/portada-movil |
| Flujo | equipo-bug |
| Fase | 4/4 · Implementado y validado (QA, revisión y seguridad aprueban) |
| Ciclo de corrección | 0/3 |
| Próximo paso | Push de la rama y abrir el PR; falta la aprobación humana del merge |
| Bloqueado por | Aprobación humana del PR |
| Actualizado | 2026-10-19 |

## Aprobaciones

| Puerta | Estado | Quién | Fecha | Frase |
|---|---|---|---|---|
| Spec | n/a (bug) | | | |
| Plan | n/a (bug) | | | |
| PR / merge | pendiente | | | |
| Despliegue | pendiente | | | |

## Hallazgos abiertos

- (ninguno) — los tres reportes se corrigieron y las validaciones de QA, revisión de código y seguridad aprueban (ver `revision-2026-10-19.md`).

## Diagnóstico (causa raíz)

- **Bug 1:** `GRID = '... lg:grid-cols-5'` en `ProximamenteBlock`, `MinisteriosBlock`, `MediosBlock` y `OfrendasBlock`. Con `max-w-6xl` y el padding de `Card`, a 1024–1440 px cada columna mide ~160–170 px.
- **Bug 2:** `REVEAL_THRESHOLD = 0.15` en `useReveal.ts`. `MediosGrid` envuelve **toda** la lista (12 fichas) en un único `Reveal`; en 390×664 el bloque mide 5.234 px y la intersección máxima posible es ~11 %, por debajo del umbral. El observador reporta «no intersecta», cancela el salvavidas y el bloque queda en `opacity: 0` de forma permanente.
- **Bug 3:** la cabecera con `data-scrolled='true'` aplica `backdrop-filter: blur(12px)`. Ese filtro crea un *containing block* para descendientes `fixed`: el scrim `inset: 0` se resuelve contra la cabecera (390×68 px) y no contra el viewport (390×844 px). Además `.public-nav-list[data-open]` usa `background: var(--surface-raised)` = `rgba(255 255 255 / 0.08)`.

## Validación (2026-10-19)

Reporte consolidado: `revision-2026-10-19.md`.

- **QA (APRUEBA):** 958 pruebas unitarias, typecheck y lint limpios; `qa-portada-movil` 2/2; 40 e2e relacionados en verde. Breakpoints medidos: 1/2/3 tarjetas a 390/768/1280; Ofrendas 471×288 px (1,64:1) a 1280 y todas visibles; `pageOverflowX = 0`. Único fallo intermitente: `movimiento.spec.ts:167`, preexistente y no atribuible.
- **Revisión de código (APRUEBA):** sin bloqueantes; semántica del umbral y apilamiento del portal verificados en Chromium real. Tres hallazgos importantes resueltos (JSDoc del e2e de Medios, nota de fragilidad del panel y este estado).
- **Seguridad (APRUEBA):** sin secretos, HTML crudo, URLs no confiables ni dependencias nuevas; `npm audit` con 0 vulnerabilidades.

## Decisiones y aclaraciones

- 2026-10-19 (humano): «Lo que viene», Ministerios y Medios pasan a **carrusel** con animación: **1 tarjeta en móvil y 3 en escritorio**. Ofrendas **no** lleva carrusel: todas las cuentas siempre visibles, en tarjetas rectangulares tipo *credit card*.
- 2026-10-19 (humano): los tres arreglos van en una sola rama `fix/portada-movil`, con un commit por arreglo y validación de QA + revisión + seguridad antes del PR.

## Bitácora

- 2026-10-19 — **Commits locales**: `8a0fc5e` (bug 3), `a430bbe` (bug 2), `d7256c5` (bug 1), `2cda2bf` (estado y validación), `5159176` (costos). Pendiente push y PR.
- 2026-10-19 — **Implementado y validado**: umbral 0 (bug 2), scrim en portal + panel opaco (bug 3), carrusel 1/2/3 para «Lo que viene», Ministerios y Medios y Ofrendas rectangulares a 2 columnas (bug 1). QA, revisión y seguridad **aprueban**. Pendiente commit y PR.
- 2026-10-19 — Diagnóstico cerrado con reproducción real (iPhone 13 emulado): scrim de 390×68, panel `rgba(255,255,255,0.08)` y bloque de 5.234 px que nunca se revela. Rama `fix/portada-movil` creada.
