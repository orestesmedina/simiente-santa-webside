# Specification Quality Checklist: Acceso y gestión de usuarios (F2)

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-04
**Last Validated**: 2026-10-04 (iteración 5 — correcciones del `analyze`: semántica del 5.º intento, política de contraseñas y módulos reservados)
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

- **[NECESITA ACLARACIÓN] resueltas (2026-10-04)**: las 5 marcas (Q1–Q5) quedaron resueltas con el humano en la fase `clarify`. Decisiones: **Q1** acción de inicialización única que crea el primer administrador, no repetible ni abusable (FR-007); **Q2** restablecimiento de contraseña hecho por un administrador, sin servicio de correo, con cambio obligatorio al entrar, y recuperación por auto-servicio con correo fuera del MVP (FR-010, Out of Scope); **Q3** no se eliminan cuentas, solo se desactivan conservando datos (FR-013); **Q4** una cuenta tiene un solo rol, los roles se editan (nombre y permisos) y se eliminan solo si no están en uso (FR-014, FR-017); **Q5** correos y nombres de rol se comparan normalizados y se rechazan como duplicados (FR-009, FR-014). Cero marcas abiertas en la spec. **Decisiones adicionales confirmadas por el humano el 2026-10-04**: política de contraseñas de **8 a 64 caracteres** (mínimo 8, máximo 64), combinando mayúsculas, minúsculas, números y caracteres especiales, y **distinta del nombre, de los apellidos y del correo** de la persona (igualdad con comparación normalizada, no de contenido) (FR-010, US7.3); límite de intentos fallidos de inicio de sesión en el que el **5.º fallo** responde el error genérico (sin revelar si la cuenta existe) y activa el bloqueo temporal del acceso durante 15 minutos, con mensaje de bloqueo desde el **6.º intento** (FR-006, Edge Cases).
- **Decisiones nuevas del humano (2026-10-04, iteración 3)**: **datos de la cuenta** = nombre, apellidos, correo y número de teléfono, todos obligatorios y con nombre y apellidos como campos separados (FR-009, FR-011, US3; el formato concreto del teléfono es una decisión menor anotada en la spec); **política de contraseñas** ampliada a distinta del nombre, de los apellidos y del correo (FR-010, US7.3); **tiempos de sesión** = 1 hora absoluta como techo desde el inicio y cierre anticipado a los 30 minutos de inactividad (FR-005, US1 escenarios 6–7). Cero marcas `[NECESITA ACLARACIÓN]` en la spec.
- **Ítem "Requirements are testable and unambiguous"** e **"All functional requirements have clear acceptance criteria"**: ahora marcados; los FR que dependían de aclaraciones (FR-007, FR-008, FR-010, FR-013, FR-017) quedaron definidos con comportamiento verificable, con escenarios de aceptación en US2, US3, US6 y US7.
- **Content Quality (implementation details)**: la spec no decide tecnologías, ni estructura, ni mecanismos concretos. Dos menciones deliberadas y por restricción ajena a esta spec: la constitución §IV (las contraseñas nunca en claro y la verificación de autorización en cada operación) y la nota pendiente de F1 sobre la sesión (`estado.md`, D-A7), citada solo como pendiente heredada. Nada de eso se diseña aquí.
- **Interpretación de la Decisión 5** (documentada en Assumptions): "permisos por módulo" se toma literalmente (un permiso = gestión completa de un módulo) y "no hay catálogo fijo" se refiere a los roles, no a los permisos. Q4 confirmó la asignación de un solo rol por cuenta; la granularidad por acción dentro de un módulo sigue fuera de alcance salvo ampliación.
- **Garantía del administrador inicial** (lo que la Decisión 5 deja a esta spec): garantizada por FR-007 (acción de inicialización única, no repetible) y FR-008 (regla anti-bloqueo confirmada: siempre al menos una cuenta activa con permiso de administración).
- **Casos límites y errores**: 24 casos límites y 16 situaciones de error esperado, todos con comportamiento definido (incluidos los derivados de Q1–Q5 y de las decisiones confirmadas el 2026-10-04: rol en uso al eliminar, rol sin permisos, repetir la inicialización, contraseña definida por administrador, contraseña que incumple la política (8–64 caracteres, 4 clases de caracteres o igualdad con el nombre, los apellidos o el correo), bloqueo activado por el 5.º intento fallido —con error genérico— y mensaje de bloqueo desde el 6.º intento, normalización de duplicados, teléfono con formato inválido y los dos cortes de sesión —30 minutos de inactividad y hora máxima—; y los derivados del cambio de alcance de auditoría: intento fallido contra un correo inexistente registrado sin revelar existencia, acción administrativa registrada aunque no se complete, intentos durante el bloqueo temporal también registrados, cuenta que nunca entró sin último acceso, historial conservado en cuentas desactivadas, ausencia total de credenciales en los registros, consulta paginada sin purga y cambio de la propia contraseña fuera del historial administrativo).
- **Fuera de alcance**: se declaró explícitamente el registro público de usuarios, la eliminación de cuentas (decisión Q3), la recuperación de contraseña por auto-servicio con correo (idea futura/backlog), la gestión de contenido de F3–F9, la identificación con cuentas externas, el bilingüismo del panel y el despliegue a producción. La **auditoría de acciones ya NO está fuera de alcance**: pasó a alcance por decisión del humano del 2026-10-04 (US8, FR-021–FR-026); quedan fuera sus límites: la auditoría del contenido de F3–F9, y la exportación, alertas y retención/purga del registro.

**Resultado de la validación (iteración 2, 2026-10-04)**: 16/16 ítems aprobados; no queda ninguna marca `[NECESITA ACLARACIÓN]`. La spec queda **aprobada en aclaraciones y a la espera de la aprobación humana** para pasar a la planificación técnica (`plan`).

**Resultado de la validación (iteración 3, 2026-10-04)**: 16/16 ítems siguen aprobados tras incorporar las tres decisiones nuevas del humano (datos de cuenta, política de contraseñas y tiempos de sesión); la spec se mantiene sin marcas `[NECESITA ACLARACIÓN]` y sin detalles de implementación, a la espera de la aprobación humana.

- **Cambio de alcance: auditoría (2026-10-04, iteración 4)**: el humano aprobó añadir a F2 la auditoría (US8, FR-021–FR-026, entidades *Registro de acceso* y *Registro de acción administrativa*, SC-012–SC-013). La sección de registro usa **el mismo permiso de administrar usuarios y roles** (decisión documentada: el catálogo de permisos corresponde a los módulos del producto y la auditoría no es un módulo, sino un instrumento de la administración de usuarios y roles). La auditoría se retiró de *Out of Scope* y se acotaron sus límites (sin auditoría de contenido de F3–F9, sin exportación, alertas ni retención/purga). Cero marcas `[NECESITA ACLARACIÓN]` y sin detalles de implementación.

**Resultado de la validación (iteración 4, 2026-10-04)**: 16/16 ítems siguen aprobados tras incorporar el cambio de alcance de auditoría; la spec se mantiene sin marcas `[NECESITA ACLARACIÓN]` y sin detalles de implementación, a la espera de la aprobación humana del añadido.

- **Correcciones del `analyze` (2026-10-04, iteración 5)**: se ajustó la spec para que sea inequívoca en tres puntos sin cambiar su alcance. **F-03 (FR-006)**: el contador de fallos se incrementa con cada intento fallido; el **5.º fallo** responde el error genérico (sin revelar si la cuenta existe) y activa el bloqueo temporal, y desde el **6.º intento** y durante 15 minutos el acceso queda bloqueado con el mensaje de bloqueo (coherente en FR-006, Edge Cases, tabla de errores, *Decisiones adicionales* y *Assumptions*). **F-09 (FR-010)**: la regla de la contraseña es **distinta del nombre, de los apellidos y del correo** (igualdad con comparación normalizada, no de contenido) y la política queda en **8–64 caracteres** (mínimo 8, máximo 64), alineada con el plan (FR-010, US7.3, Edge Cases, tabla de errores, Q2, *Decisiones adicionales* y *Assumptions*). **F-14**: los módulos reservados que aún no existen pasan a **F3–F9** (la portada F3 tampoco existe en F2). Cero marcas `[NECESITA ACLARACIÓN]` y sin detalles de implementación.

**Resultado de la validación (iteración 5, 2026-10-04)**: 16/16 ítems siguen aprobados tras las correcciones del `analyze`; la spec se mantiene sin marcas `[NECESITA ACLARACIÓN]`, sin detalles de implementación y sin cambios de alcance.
