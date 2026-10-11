# Revisión consolidada final — F10 (2026-10-11)

**Rama:** `010-sistema-diseno-movimiento` · **Alcance:** `git diff main...HEAD` (todo F10) · **Ciclos:** 3/3

Reportes: `revision-2026-10-11-{qa,codigo,seguridad}.md` (ciclo 1), `…-ciclo2.md`, `…-ciclo3.md`.

## Veredicto global: **APROBADO** ✅

| Rol | Ciclo 1 | Ciclo 2 | Ciclo 3 (final) |
|---|---|---|---|
| `qa-tester` | APROBADO | APROBADO | **APROBADO** |
| `revisor-codigo` | RECHAZADO (C1, C2) | RECHAZADO (N1) | **APROBADO** |
| `seguridad` | APROBADO | APROBADO | **APROBADO** |

## Bucle de corrección (3 ciclos)

| Ciclo | Hallazgos | Corrección (`dev-frontend`) |
|---|---|---|
| 1 | **C1** (revelado auto-disparado fuera de pantalla), **C2** (contraste hover botón primario en oscuro), **IMP-1** (selector más grande que F3), **I1** (CTA «Ver horario» roto en borrador) | `48dd0e9`, `f240ea0`, `dd4c5ce`, `1339f3e` (+ `00bf59f` M1-seg, `71f9729` M2-rev) |
| 2 | **N1** (regresión de contraste de la opción activa del selector, 4,05:1 en claro) | `880f840` (+ `386be2e` cierra NIT) |
| 3 | ninguno | — |

## Estado final verificado

- Unit **409/409**, e2e F10 **25/25**, e2e congeladas F2/F3 **8/8**; lint/typecheck/build ✔.
- **SC-008** intacto tras los 3 ciclos: 0 pruebas congeladas editadas; contrato de F2/F3 intacto.
- **0 cambios** en backend/migraciones/API/infra/`.env.example`; **0 dependencias nuevas**; `npm audit` sin vulnerabilidades altas/críticas; sin XSS ni secretos.
- Contrastes AA verificados (hover del botón y opción activa del selector, claro y oscuro).

## Pendientes de verificación humana (no automatizables)

- **SC-002** («moderno» y fiel a la marca), **SC-013** (propuesta B, «0 copias» de las referencias), **SC-014** (carácter del movimiento), **SC-003/SC-010** (pruebas con personas), **SC-001/SC-011** (visual) y **SC-009** (contraste fino en tema oscuro): se cierran en la **revisión de aceptación con el cliente** (puerta de PR).
- **M-4** heredado de F2 (rate-limit de login → e2e en dos tandas) y endurecimiento CSP futuro (documentado, no bloqueante).

**F10 lista para la puerta de entrega (PR).**
