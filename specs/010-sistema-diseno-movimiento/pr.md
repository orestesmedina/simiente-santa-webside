# PR — F10 Sistema de diseño y movimiento

> Descripción para el Pull Request de la rama `010-sistema-diseno-movimiento` (preparada por
> `devops`, 2026-10-11). **El merge y el despliegue requieren aprobación humana explícita**
> (fases 9/9 del flujo equipo-feature; puertas «PR / merge» y «Despliegue» pendientes en
> `estado.md`).

## feat: F10 — Sistema de diseño y movimiento

### Resumen

El sistema visual y de movimiento que recae sobre todo el sitio público (y que las secciones
futuras del roadmap heredarán), construido **exclusivamente en frontend + specs: sin cambios de
backend, migraciones, datos ni API**. Diseño final **B «Marea Turquesa»** elegido por el cliente
entre 3 propuestas (iteración de la puerta 2). Cierra las **36 tareas (T1001–T1036)** de
`tasks.md`.

**Frontend (React + TypeScript + Vite, sin dependencias nuevas)**
- **Sistema de diseño**: tokens de color/tipografía/espaciado con variables semánticas; identidad
  según el Manual de la marca (paleta turquesa `#00c9a7` sobre azul profundo `#1a2b4a`,
  Bebas Neue/Poppins/Playfair Display, logo); iconografía e ilustraciones nuevas dentro de la marca
  (Q6), con 0 copias de las referencias (iglesiaoasis.com, river-arena.org, apple.com,
  microsoft.com — solo inspiración, SC-013).
- **Movimiento marcado pero elegante** (revisado con el cliente en la puerta 2): revelado de
  secciones al hacer scroll con `IntersectionObserver` (600 ms, desplazamiento 20–28 px, *stagger*
  de 80 ms, suelo de perceptibilidad ≥ 400 ms/16 px, techo ≤ 800 ms), cabecera que transiciona al
  bajar y `scroll-behavior: smooth` — **todo apagado por completo con `prefers-reduced-motion`**
  (conmutación en caliente verificada: un bloque ya revelado no se oculta al apagarla, M2).
- **Navegación por dispositivo** (FR-010…FR-014): barra fija siempre visible en escritorio;
  en móvil ≤ 860 px, **menú hamburguesa accesible** (patrón *disclosure*, colapsado fuera del
  orden de tabulación, sin trampa de foco, scroll de fondo bloqueado), una sola lista de secciones
  en ambos; orden y textos de F3 intactos.
- **Selector de idioma compacto** (FR-015…, US4): solo compacta los textos actuales de F3
  (Q5), 152×44 px de objetivo táctil (menor que el de F3, IMP-1), opción activa con par de
  contraste AA en claro y oscuro (N1).
- **Modo oscuro con interruptor manual** (US5): `data-theme` con variables semánticas, preferencia
  en `localStorage`, contraste AA por tema (hover del botón primario 5,3:1 claro / 6,7:1 oscuro,
  C2); el sistema respeta la preferencia del SO como inicial.
- **Mapa del sitio en el pie** (FR-027, SC-015): grupos **Explorar / Contacto / Síguenos** con
  nombres accesibles únicos, sin grupos vacíos; marca «Simiente» presente con `sr-only`.
- CTA del hero «Escríbenos por WhatsApp» **condicional a la sección publicada** (I1) y con texto
  propio distinto del accesible congelado de F3 (decisión 2026-10-10 en `estado.md`).
- 7 e2e nuevas de Playwright (`diseno-movimiento`, `navegacion-movil`, `tema-oscuro`, `movimiento`,
  `movimiento-reducido`, `qa-medidas-f10`, helpers) + pruebas unitarias y componentes nuevas con
  *fakes* propios (`intersectionObserver.ts`, `matchMedia.ts`) — **0 pruebas congeladas de F2/F3
  editadas** (SC-008/RG10-6).

**Backend / datos / API / infraestructura**: **sin cambios.** Verificado contra la base de la rama
(`17094ce`): `git diff` vacío en `backend/` (incl. `migrations/` — la última sigue siendo
`000006_*`), `backend/api`, `frontend/src/api/schema.d.ts` (`api:gen` idempotente, sin deriva de
contrato), `.env.example` (F10 no añade variables), `docker-compose.yml`, `Dockerfile*` y
`.github/`. 51 archivos de `frontend/` + specs + 1 línea de roadmap.

### Cómo probarlo (quickstart)

Recorrido completo con evidencia en [`quickstart.md`](./quickstart.md) (§8.1 y §9). Resumen:

```bash
make up                          # db, redis, backend, frontend (sin migraciones nuevas)
cd frontend && npm install
# Sitio público: http://localhost:5173/  ·  estado: /health
# - bajar: revelado progresivo de secciones · barra fija · mapa del sitio en el pie
# - interruptor de tema claro/oscuro (persistente) · selector de idioma compacto
# - móvil <= 860 px: menú hamburguesa accesible
# con prefers-reduced-motion activado: el movimiento desaparece y el contenido es visible de una vez
cd frontend && npm run ci.sh     # lint + typecheck + unit + build + audit
# e2e (dos tandas, ver nota M-4):
npx playwright test --workers=1  # requiere stack levantado; en WSL con imagen Docker de Playwright
```

### Evidencia de pruebas (2026-10-11, HEAD `216208b`)

| Verificación | Resultado |
|---|---|
| `make ci` (lint + backend + frontend + security) | **EXIT=0** · `gofmt`/`go vet`/`golangci-lint` 0 issues · backend unit+integración ok (takedown: `backend/` sin cambios) · frontend lint/typecheck/unit en verde · `build de producción` ok (548 kB, aviso de chunk preexistente, no de F10) · `govulncheck`/`npm audit` sin vulnerabilidades altas/críticas |
| Unitarias de frontend | **409 pruebas en verde** (incluye las de F1–F3 sin deriva) |
| e2e Playwright (`--workers=1`, stack Docker real) | **25/25 F10** + **8/8 congeladas de F2/F3**, en verde (dos tandas por el rate-limit de login, M-4) |
| No-regresión F2/F3 | **0 pruebas congeladas editadas**; anclas congeladas y contrato semántico (R10-9) intactos: `ss.lang`, `tabCycle < 30` a 320 px, claves `nav.label`/`a11y.skip`/`lang.*` |
| Sin deriva | `api:gen` + `git diff --exit-code` en `schema.d.ts` → sin diff · `sqlc`/migraciones intactas · `.env.example` de 57 líneas de F2/F3 · `docker compose ps` sano (`healthz` 200, SPA 200) |

### Validación del equipo (ciclo 3 — APROBADO por los tres roles)

Ciclo 1 halló 4 hallazgos (C1, C2, IMP-1, I1) + 2 menores (M1-seguridad, M2-revisor); el ciclo 2
detectó la regresión de contraste **N1** introducida por IMP-1; el ciclo 3 quedó limpio.
Veredictos en [`revision-2026-10-11-final.md`](./revision-2026-10-11-final.md):

| Ciclo | Hallazgos | Corrección |
|---|---|---|
| 1 | **C1** (revelado auto-disparado fuera de pantalla), **C2** (hover del botón primario en oscuro), **IMP-1** (selector más grande que F3), **I1** (CTA «Ver horario» roto en borrador) | `48dd0e9`, `f240ea0`, `dd4c5ce`, `1339f3e` (+ `00bf59f`, `71f9729`) |
| 2 | **N1** (contraste de la opción activa del selector, 4,05:1 en claro) | `880f840` (+ NIT `386be2e`) |
| 3 | ninguno | — |

Reportes por ciclo y rol: `revision-2026-10-11-{qa,codigo,seguridad}.md`, `…-ciclo2.md`,
`…-ciclo3.md` (ciclo 1), y el consolidado
[`revision-2026-10-11-final.md`](./revision-2026-10-11-final.md). La verificación de cierre de
`devops` está en [`cierre.md`](./cierre.md).

### Pendientes de coordinación humana (no bloquean el merge)

La parte automatizable quedó en verde; estos criterios se cierran en la **revisión de aceptación
con el cliente** (checklist lista en `quickstart.md` §9):

- **SC-002 / SC-014** (subjetivos y de marca) — «moderno, acorde a la marca» y carácter del
  movimiento «perceptible y no mareante»: impresión visual del cliente.
- **SC-013** — 0 elementos copiados: revisión de originalidad de la propuesta B frente a las
  4 referencias.
- **SC-001 / SC-003 / SC-010 / SC-011** — impresión visual y **pruebas con personas** (≥ 6
  participantes, ≥ 80 % de tareas completadas, preferencia de diseño ≥ 70 %) — coordinación
  humana, no automatizable.
- **SC-009** — contraste fino (con lector / arranque de brillo) en tema oscuro: revisión fina del
  cliente.
- **M-4** (heredado de F2) — rate-limit de login de 20/min por IP: los e2e se ejecutan en dos
  tandas; el endurecimiento futuro del CSP está documentado y no es bloqueante.

### Notas de entrega

- Estado: **36/36 tareas, Converged** (fase 8); esta preparación del PR cierra la parte `devops`
  de la fase 9 (Entregar).
- La rama `010-sistema-diseno-movimiento` está lista para push; el **push y el PR los lanza el
  humano** (como en F3, el entorno de `devops` no tiene `gh`/credenciales). Comando sugerido:

```bash
git push -u origin 010-sistema-diseno-movimiento
gh pr create --base main --head 010-sistema-diseno-movimiento \
  --title "feat: F10 — Sistema de diseño y movimiento" \
  --body-file specs/010-sistema-diseno-movimiento/pr.md
```

- Tras el merge: `documentador` (CHANGELOG/README/quickstart) y cierre de costos con
  `make costos CERRAR=1`.
- **Sin despliegue a producción** en esta feature (no existe workflow de CD y el entorno de
  producción no está aprobado; observación B de `cierre.md`).
