# Specification Quality Checklist: Acceso y gestión de usuarios (F2)

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-04
**Last Validated**: 2026-10-04 (iteración 2 — aclaraciones Q1–Q5 resueltas en la fase `clarify`)
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notes

- **[NECESITA ACLARACIÓN] resueltas (2026-10-04)**: las 5 marcas (Q1–Q5) quedaron resueltas con el humano en la fase `clarify`. Decisiones: **Q1** acción de inicialización única que crea el primer administrador, no repetible ni abusable (FR-007); **Q2** restablecimiento de contraseña hecho por un administrador, sin servicio de correo, con cambio obligatorio al entrar, política de 8 caracteres mínimo, y recuperación por auto-servicio con correo fuera del MVP (FR-010, Out of Scope); **Q3** no se eliminan cuentas, solo se desactivan conservando datos (FR-013); **Q4** una cuenta tiene un solo rol, los roles se editan (nombre y permisos) y se eliminan solo si no están en uso (FR-014, FR-017); **Q5** correos y nombres de rol se comparan normalizados y se rechazan como duplicados (FR-009, FR-014). Cero marcas abiertas en la spec.
- **Ítem "Requirements are testable and unambiguous"** e **"All functional requirements have clear acceptance criteria"**: ahora marcados; los FR que dependían de aclaraciones (FR-007, FR-008, FR-010, FR-013, FR-017) quedaron definidos con comportamiento verificable, con escenarios de aceptación en US2, US3, US6 y US7.
- **Content Quality (implementation details)**: la spec no decide tecnologías, ni estructura, ni mecanismos concretos. Dos menciones deliberadas y por restricción ajena a esta spec: la constitución §IV (las contraseñas nunca en claro y la verificación de autorización en cada operación) y la nota pendiente de F1 sobre la sesión (`estado.md`, D-A7), citada solo como pendiente heredada. Nada de eso se diseña aquí.
- **Interpretación de la Decisión 5** (documentada en Assumptions): "permisos por módulo" se toma literalmente (un permiso = gestión completa de un módulo) y "no hay catálogo fijo" se refiere a los roles, no a los permisos. Q4 confirmó la asignación de un solo rol por cuenta; la granularidad por acción dentro de un módulo sigue fuera de alcance salvo ampliación.
- **Garantía del administrador inicial** (lo que la Decisión 5 deja a esta spec): garantizada por FR-007 (acción de inicialización única, no repetible) y FR-008 (regla anti-bloqueo confirmada: siempre al menos una cuenta activa con permiso de administración).
- **Casos límites y errores**: 14 casos límites y 13 situaciones de error esperado, todos con comportamiento definido (incluidos los derivados de Q1–Q5: rol en uso al eliminar, rol sin permisos, repetir la inicialización, contraseña definida por administrador, normalización de duplicados).
- **Fuera de alcance**: se declaró explícitamente el registro público de usuarios, la eliminación de cuentas (decisión Q3), la recuperación de contraseña por auto-servicio con correo (idea futura/backlog), la gestión de contenido de F3–F9, la identificación con cuentas externas, la auditoría de acciones, el bilingüismo del panel y el despliegue a producción.

**Resultado de la validación (iteración 2, 2026-10-04)**: 16/16 ítems aprobados; no queda ninguna marca `[NECESITA ACLARACIÓN]`. La spec queda **aprobada en aclaraciones y a la espera de la aprobación humana** para pasar a la planificación técnica (`plan`).
