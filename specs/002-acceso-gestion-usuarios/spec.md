# Feature Specification: Acceso y gestión de usuarios (F2)

**Feature Branch**: `002-acceso-gestion-usuarios`

**Created**: 2026-10-04

**Status**: Draft — pendiente de aprobación humana (5 aclaraciones abiertas, ver *Aclaraciones pendientes*)

**Input**: User description: "F2 — Acceso y gestión de usuarios (roadmap §3, `docs/producto/roadmap.md`): El equipo de la iglesia inicia sesión en un panel de administración; los administradores crean usuarios, los activan o desactivan, y crean roles con los permisos por módulo que necesiten (p. ej., un rol de contenido, un rol de ministerios)." Decisiones aplicables del roadmap: **Decisión 4** (usuarios activo/inactivo) y **Decisión 5** (permisos por módulo agrupados en roles, sin catálogo fijo de roles; el administrador crea roles nuevos y les asigna permisos de forma independiente; **la spec de F2 define cómo queda garantizado el arranque con un administrador inicial**). Restricción de `idea.md` §2: los usuarios comunes no se registran; solo el administrador crea cuentas.

> **Nota de alcance y terminología**: esta funcionalidad es la puerta del **panel de administración** que usarán las funcionalidades F3–F9. Se distinguen dos personas distintas: el **visitante** (el "usuario común" de `idea.md` §2: ve el sitio público, no tiene cuenta y no se registra) y el **usuario del panel** (el "usuario del sistema" y el "usuario administrador" de `idea.md` §2: tiene cuenta, inicia sesión y gestiona el contenido). En esta spec, "usuario" significa siempre **usuario del panel** salvo que se diga lo contrario. F2 entrega: acceso autenticado al panel, gestión de cuentas (crear, editar, activar/desactivar) y gestión de roles con permisos por módulo. El contenido de los módulos (portada, eventos, actividades, grupos, ministerios, donaciones, noticias, medios) llega en F3–F9; aquí solo se definen los permisos que esos módulos usarán.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Iniciar sesión en el panel de administración (Priority: P1)

Como miembro del equipo de la iglesia, quiero iniciar sesión en el panel de administración con mi correo y mi contraseña, para acceder a las herramientas de gestión que me corresponden.

**Why this priority**: Es la puerta de entrada de F2 y de todo el panel (F3–F9 dependen de ella). Sin acceso autenticado no se puede proteger ninguna gestión y todos los demás requisitos de esta funcionalidad quedan sin sentido.

**Independent Test**: Se prueba de forma independiente con el panel disponible y al menos una cuenta activa: se inicia sesión con credenciales válidas y se llega al panel; se intenta con credenciales incorrectas, con una cuenta inactiva y sin sesión, y en los tres casos se bloquea el acceso. Entrega valor por sí sola: el equipo ya tiene una puerta segura al panel.

**Acceptance Scenarios**:

1. **Dado** una cuenta activa con credenciales válidas, **cuando** inicio sesión en la pantalla de acceso, **entonces** entro al panel de administración y veo las secciones que mis permisos autorizan.
2. **Dado** una cuenta activa, **cuando** envío un correo o una contraseña incorrectos, **entonces** el sistema muestra un mensaje de error genérico, no me deja entrar y no revela si la cuenta existe.
3. **Dado** una cuenta inactiva, **cuando** intento iniciar sesión con credenciales correctas, **entonces** el sistema impide el acceso con un mensaje comprensible.
4. **Dado** que no he iniciado sesión, **cuando** intento abrir cualquier sección del panel, **entonces** soy llevado a la pantalla de acceso.
5. **Dado** que inicié sesión, **cuando** elijo cerrar sesión, **entonces** mi acceso termina y cualquier acción posterior exige iniciar sesión de nuevo.
6. **Dado** una sesión abierta sin actividad, **cuando** transcurre el periodo de inactividad definido, **entonces** la sesión termina y debo iniciar sesión de nuevo para continuar.

---

### User Story 2 - Garantizar el arranque con un administrador inicial (Priority: P1)

Como responsable del sistema, quiero que todo arranque del sistema deje garantizado al menos un administrador inicial con permisos completos, para que siempre exista alguien que pueda crear usuarios y roles sin depender de nadie más.

**Why this priority**: La Decisión 5 del roadmap deja expresamente a esta spec cómo se garantiza el arranque. Sin esa garantía, el panel podría nacer sin nadie capaz de administrarlo y toda la gestión de usuarios y roles quedaría inaccesible o tendría que resolverse fuera del producto.

**Independent Test**: Se prueba de forma independiente en una instalación nueva (sin cuentas): se pone en marcha el sistema con el mecanismo definido y se comprueba que existe un administrador inicial capaz de entrar y crear usuarios y roles; después se intenta dejar el panel sin administración activa y se comprueba que el sistema lo impide. Entrega valor por sí sola: el panel siempre es administrable desde el primer minuto.

**Acceptance Scenarios**:

1. **Dado** una instalación nueva sin cuentas de usuario, **cuando** se pone en marcha el sistema por primera vez con el mecanismo de arranque definido (aclaración Q1), **entonces** queda garantizado un administrador inicial con permisos completos.
2. **Dado** el administrador inicial, **cuando** inicia sesión por primera vez, **entonces** puede crear usuarios y roles de inmediato, sin pasos adicionales.
3. **Dado** el panel con un único administrador activo, **cuando** intento desactivarlo (o eliminarlo, según la aclaración Q3), **entonces** el sistema lo impide con un mensaje que explica que debe quedar al menos un administrador activo.
4. **Dado** el panel con varios administradores activos, **cuando** desactivo uno de ellos, **entonces** la operación se completa con normalidad mientras quede al menos otro administrador activo.
5. **Dado** el panel con un único administrador activo, **cuando** intento quitarle el permiso de administrar usuarios y roles (cambiando su rol), **entonces** el sistema lo impide por la misma regla.

---

### User Story 3 - Crear cuentas para el equipo (Priority: P1)

Como administrador, quiero crear cuentas para las personas del equipo indicando su identificación, su correo y el rol que les corresponde, para que cada quien trabaje con su propia cuenta y nadie comparta credenciales.

**Why this priority**: La gestión de cuentas es una de las dos mitades del alcance de F2 (`idea.md` §2 y Decisión 4). Sin cuentas propias no hay roles que asignar ni acceso diferenciado, y seguiríamos compartiendo un único acceso.

**Independent Test**: Se prueba de forma independiente con un administrador autenticado: crea una cuenta válida y aparece en el listado en estado activo; repite la operación con un correo ya en uso y con datos inválidos, y el sistema la impide en ambos casos. Entrega valor por sí sola: el equipo ya puede tener cuentas individuales.

**Acceptance Scenarios**:

1. **Dado** un administrador autenticado, **cuando** creo una cuenta con datos válidos y un rol asignado, **entonces** la cuenta queda creada en estado activo y aparece en el listado de usuarios.
2. **Dado** un administrador, **cuando** intento crear una cuenta con un correo que ya está en uso, **entonces** el sistema lo impide con un mensaje claro y no queda ninguna cuenta duplicada.
3. **Dado** un formulario de creación con datos inválidos o incompletos (por ejemplo, un correo mal formado o un campo obligatorio vacío), **cuando** envío el formulario, **entonces** el sistema indica qué corregir y no crea la cuenta.
4. **Dado** una persona sin el permiso de administrar usuarios y roles, **cuando** intenta crear una cuenta, **entonces** el sistema impide la operación.
5. **Dado** un administrador, **cuando** creo una cuenta indicando un rol que no existe, **entonces** el sistema lo impide indicando que debe elegir un rol válido.

---

### User Story 4 - Crear roles con permisos por módulo (Priority: P1)

Como administrador, quiero crear roles con un nombre propio y los permisos por módulo que necesiten (p. ej., un rol de contenido con eventos y actividades, sin ministerios), para que cada persona del equipo solo acceda a lo que le corresponde.

**Why this priority**: Es la segunda mitad del alcance de F2 (Decisión 5): sin roles con permisos por módulo, todo el equipo tendría el mismo acceso y no se podría delegar trabajo con seguridad. Es independiente de la gestión de cuentas y se puede construir y probar por separado.

**Independent Test**: Se prueba de forma independiente con un administrador autenticado: crea un rol con dos permisos de módulo combinados libremente, lo asigna a una cuenta y se comprueba que esa cuenta solo puede gestionar esos módulos; se comprueba además que un nombre de rol repetido se rechaza. Entrega valor por sí sola: ya existe delegación de acceso por módulo.

**Acceptance Scenarios**:

1. **Dado** un administrador autenticado, **cuando** creo un rol con un nombre único y uno o más permisos por módulo, **entonces** el rol queda disponible para asignarse a cuentas.
2. **Dado** un administrador, **cuando** creo un rol combinando permisos de forma libre (p. ej., eventos y actividades, sin ministerios), **entonces** el sistema acepta esa combinación sin exigir un catálogo previo de roles.
3. **Dado** un administrador, **cuando** intento crear un rol con un nombre que ya existe, **entonces** el sistema lo impide con un mensaje claro.
4. **Dado** una cuenta cuyo rol no incluye el permiso de administrar usuarios y roles, **cuando** intenta gestionar usuarios o roles, **entonces** el sistema impide la operación.
5. **Dado** una cuenta cuyo rol otorga permisos sobre un módulo, **cuando** esa persona trabaja en el panel, **entonces** solo puede usar los módulos de sus permisos: los demás no le aparecen ni puede forzar su acceso.

---

### User Story 5 - Activar y desactivar cuentas (Priority: P2)

Como administrador, quiero activar o desactivar las cuentas del equipo, para controlar quién puede entrar al panel sin perder los datos de lo que ya se hizo.

**Why this priority**: La Decisión 4 (usuarios activo/inactivo) es parte explícita del alcance. Va después de crear cuentas y roles porque solo tiene sentido sobre ellos, pero es necesaria para responder a las altas y bajas del equipo.

**Independent Test**: Se prueba de forma independiente con una cuenta activa: se desactiva y se comprueba que ya no puede entrar (ni con sesión abierta) y que sus datos siguen disponibles; se vuelve a activar y recupera el acceso. Entrega valor por sí sola: el control de quién entra al panel queda en manos del administrador.

**Acceptance Scenarios**:

1. **Dado** una cuenta activa, **cuando** un administrador la desactiva, **entonces** su titular ya no puede iniciar sesión con credenciales correctas y los datos de la cuenta se conservan.
2. **Dado** una cuenta desactivada con una sesión abierta, **cuando** la desactivación se hace efectiva, **entonces** pierde el acceso al panel de inmediato.
3. **Dado** una cuenta inactiva, **cuando** un administrador la vuelve a activar, **entonces** su titular recupera el acceso con normalidad.
4. **Dado** el listado de cuentas, **cuando** un administrador lo consulta, **entonces** ve cada cuenta con su estado (activo/inactivo), su correo y su rol, para saber quién tiene acceso hoy.
5. **Dado** una persona sin el permiso de administrar usuarios y roles, **cuando** intenta activar o desactivar cuentas, **entonces** el sistema impide la operación.

---

### User Story 6 - Mantener los roles al día (Priority: P2)

Como administrador, quiero actualizar los permisos de un rol existente y eliminar los roles que ya no se usan, para que el acceso del equipo siga reflejando el trabajo real.

**Why this priority**: Las responsabilidades del equipo cambian con el tiempo; sin mantenimiento de roles, los permisos quedan viejos y el panel se vuelve inseguro o incómodo. Depende de la creación de roles (US4), por eso va después.

**Independent Test**: Se prueba de forma independiente con un rol asignado a una cuenta: se le quita un permiso y se comprueba que la cuenta deja de poder usar ese módulo; se le agrega otro y la cuenta lo gana. La eliminación de roles se prueba según lo que defina la aclaración Q4. Entrega valor por sí sola: los permisos se mantienen al día sin recrear roles.

**Acceptance Scenarios**:

1. **Dado** un rol existente con cuentas asignadas, **cuando** un administrador le quita un permiso, **entonces** las cuentas con ese rol dejan de tener ese acceso a partir de ese momento.
2. **Dado** un rol existente con cuentas asignadas, **cuando** un administrador le agrega un permiso, **entonces** las cuentas con ese rol ganan ese acceso sin necesidad de pasos adicionales.
3. **Dado** un rol que ya no se usa, **cuando** un administrador lo elimina, **entonces** el sistema aplica el comportamiento que defina la aclaración Q4 (y en todo caso nunca deja a una cuenta en una situación que pueda bloquear la administración del panel).
4. **Dado** una persona sin el permiso de administrar usuarios y roles, **cuando** intenta actualizar o eliminar un rol, **entonces** el sistema impide la operación.

---

### User Story 7 - Cambiar mi propia contraseña (Priority: P3)

Como usuario del panel, quiero cambiar mi contraseña desde mi cuenta, para mantener el control de mi acceso.

**Why this priority**: Refuerza la seguridad de las cuentas individuales y completa el ciclo de credenciales, pero no bloquea la gestión de usuarios ni de roles; por eso va al final.

**Independent Test**: Se prueba de forma independiente con una cuenta autenticada: cambia su contraseña indicando la actual y una nueva válida y luego entra con la nueva; con la contraseña actual incorrecta y con una nueva que no cumple la política, el sistema lo impide. Entrega valor por sí sola: cada persona mantiene su propia credencial.

**Acceptance Scenarios**:

1. **Dado** una cuenta autenticada, **cuando** cambia su contraseña indicando la actual y una nueva válida, **entonces** el cambio se aplica y puede iniciar sesión con la nueva contraseña.
2. **Dado** una cuenta autenticada, **cuando** indica una contraseña actual incorrecta, **entonces** el sistema impide el cambio con un mensaje claro.
3. **Dado** una cuenta autenticada, **cuando** propone una contraseña que no cumple la política vigente, **entonces** el sistema indica el requisito incumplido y no aplica el cambio.

---

### Edge Cases

- **Cuenta desactivada con sesión abierta**: pierde el acceso de inmediato (US5, escenario 2); no debe poder seguir actuando en el panel con esa sesión.
- **Último administrador activo**: ninguna operación (desactivar, eliminar o quitarle el permiso de administración) puede dejar el panel sin administración; el sistema lo impide con un mensaje que explica la restricción.
- **Permisos sobre módulos que aún no existen** (eventos, actividades, grupos, ministerios, donaciones, noticias, medios: F4–F9): quedan reservados desde F2 y no dan acceso a nada hasta que esos módulos se construyan.
- **Cuenta sin ningún permiso de módulo** (solo con acceso al panel): entra pero no ve ninguna sección de gestión; esto es válido y no debe romper el panel.
- **Dos personas creando a la vez dos cuentas con el mismo correo**: solo debe quedar creada una; la segunda operación se rechaza como duplicado.
- **Un rol eliminado que está asignado a cuentas**: su comportamiento exacto queda definido por la aclaración Q4; en ningún caso debe poder bloquear la administración del panel.
- **Una cuenta con varios roles** (si la aclaración Q4 lo permite): los permisos efectivos deben ser la suma de los permisos de sus roles, sin contradicciones ni accesos de más.
- **Cadena de intentos fallidos de inicio de sesión** (alguien probando contraseñas): el sistema limita los intentos fallidos para impedir la prueba masiva de contraseñas.
- **Sesión que expira mientras se llena un formulario**: la acción se interrumpe y el sistema pide iniciar sesión de nuevo; no debe aplicarse a medias ni perderse sin aviso.
- **Pérdida de acceso de la única persona que sabe su contraseña**: resolución según la aclaración Q2 (auto-servicio o restablecimiento por un administrador).
- **Intento de crear una cuenta o un rol cuyo correo o nombre solo difiere en mayúsculas o en espacios sobrantes**: el sistema debe indicar el problema en lugar de guardar una cuenta o un rol prácticamente duplicado. [NECESITA ACLARACIÓN: ¿un correo o un nombre de rol que solo difiere en mayúsculas o en espacios sobrantes debe tratarse como el mismo (y rechazarse como duplicado) o como distinto?]

### Errores esperados

Todos los mensajes deben ser comprensibles para personas no técnicas y nunca deben exponer información interna del sistema (coherente con FR-012 y FR-013 de F1).

| Situación | Qué debe ver la persona |
|---|---|
| Correo o contraseña incorrectos | Mensaje de error genérico (sin revelar si la cuenta existe) |
| Cuenta inactiva | Mensaje que indica que ese acceso está desactivado |
| Datos inválidos o incompletos en un formulario | Indicación de qué campo corregir, sin crear nada |
| Correo de cuenta ya en uso | Mensaje claro de duplicado, sin crear la cuenta |
| Nombre de rol ya en uso | Mensaje claro de duplicado, sin crear el rol |
| Operación sin el permiso correspondiente | Mensaje claro de acceso denegado, sin detalles internos |
| Operación que dejaría el panel sin administrador | Mensaje que explica la restricción y cómo proceder |
| Sesión expirada o cerrada | Mensaje que pide iniciar sesión de nuevo |
| Contraseña nueva que no cumple la política | Mensaje con el requisito incumplido |
| Error inesperado del sistema | Mensaje genérico sin información interna |

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema DEBE exponer un panel de administración cuyas secciones solo son accesibles con una sesión iniciada; todo intento de acceso sin sesión DEBE dirigir a la pantalla de acceso.
- **FR-002**: El sistema DEBE autenticar a las personas del equipo con su correo y su contraseña, y DEBE rechazar el acceso de las cuentas inactivas aunque las credenciales sean correctas.
- **FR-003**: El mensaje de un inicio de sesión fallido DEBE ser genérico y NUNCA DEBE revelar si la cuenta existe; la contraseña NUNCA DEBE mostrarse ni devolverse en ninguna respuesta.
- **FR-004**: El sistema DEBE permitir cerrar sesión en cualquier momento desde el panel, terminando el acceso de esa sesión.
- **FR-005**: La sesión DEBE terminar tras un periodo de inactividad, y todo acceso posterior DEBE exigir un nuevo inicio de sesión.
- **FR-006**: El sistema DEBE limitar los intentos fallidos de inicio de sesión para impedir la prueba masiva de contraseñas.
- **FR-007**: Tras la instalación, el sistema DEBE garantizar un administrador inicial con permisos completos (capaz de crear usuarios y roles de inmediato) mediante un mecanismo de arranque definido. [NECESITA ACLARACIÓN: ¿cuál es el mecanismo de arranque del administrador inicial: credenciales iniciales definidas por quien instala el sistema, una acción de inicialización ejecutada una sola vez, o un asistente de primer acceso en la propia pantalla de acceso?]
- **FR-008**: El sistema DEBE impedir que el panel se quede sin al menos una cuenta activa con permiso de administrar usuarios y roles: ninguna operación (desactivar, eliminar, cambiar su rol o quitarle ese permiso) DEBE poder dejarlo sin administración.
- **FR-009**: Los administradores DEBEN poder crear cuentas del equipo indicando identificación, correo y rol asignado; el sistema DEBE rechazar correos ya en uso, roles inexistentes y datos inválidos o incompletos, indicando qué corregir.
- **FR-010**: El sistema DEBE definir cómo recibe su contraseña inicial una cuenta nueva y cómo recupera el acceso quien la olvida, y DEBE exigir una política de contraseñas. [NECESITA ACLARACIÓN: ¿la contraseña inicial y la recuperación de acceso se resuelven por auto-servicio del usuario con un mensaje a su correo (lo que exige un servicio de correo) o mediante un restablecimiento hecho por un administrador? ¿Y qué reglas debe cumplir la contraseña: longitud mínima y complejidad?]
- **FR-011**: Los administradores DEBEN poder editar los datos de una cuenta (identificación, correo y rol asignado) y su estado (activo/inactivo).
- **FR-012**: Al desactivar una cuenta, el sistema DEBE bloquear su acceso de inmediato —también las sesiones ya abiertas— y DEBE conservar todos sus datos.
- **FR-013**: El sistema DEBE poder eliminar cuentas además de desactivarlas. [NECESITA ACLARACIÓN: `idea.md` §2 dice que el administrador puede "crear más usuarios o eliminarlos o desactivarlos", pero el roadmap (Decisión 4 y la fila F2) solo define activo/inactivo: ¿está en alcance eliminar cuentas y, si lo está, qué ocurre con sus datos y con el contenido que esa persona haya creado?]
- **FR-014**: Los administradores DEBEN poder crear roles compuestos por un nombre único y un conjunto de permisos por módulo, combinados libremente (no existe un catálogo fijo de roles).
- **FR-015**: Los permisos DEBEN corresponder a los módulos del producto: portada e información general, eventos, actividades, grupos de conexión, ministerios, donaciones, noticias y galería, medios, y administración de usuarios y roles; cada permiso autoriza la gestión completa de su módulo.
- **FR-016**: El sistema DEBE verificar el permiso correspondiente en cada operación del panel según el rol de la cuenta, y DEBE denegar con un mensaje claro toda operación no autorizada.
- **FR-017**: La asignación de roles a una cuenta y el ciclo de vida de los roles (si se pueden editar y eliminar y qué ocurre con las cuentas que ya tienen un rol eliminado) DEBEN responder a lo que se defina en la aclaración correspondiente. [NECESITA ACLARACIÓN: ¿una cuenta puede tener varios roles a la vez (con permisos que se suman) o solo un rol? ¿Y un rol se puede editar y eliminar, y qué ocurre con las cuentas que ya tienen asignado un rol eliminado?]
- **FR-018**: Todo cambio en los permisos de un rol DEBE reflejarse de inmediato en el acceso de las cuentas que lo tienen, sin pasos adicionales.
- **FR-019**: El sistema DEBE mostrar el listado de cuentas con su estado (activo/inactivo), su correo y su rol asignado.
- **FR-020**: Las personas con cuenta en el panel DEBEN poder cambiar su propia contraseña.

### Key Entities

- **Cuenta del panel**: persona del equipo de la iglesia con acceso autenticado al panel. Se identifica por su correo, tiene datos de identificación, un estado (activo/inactivo) y roles asignados. No debe confundirse con el "usuario común" de `idea.md` §2: el visitante del sitio público no tiene cuenta ni se registra.
- **Rol**: nombre único creado por el administrador que agrupa permisos; define qué puede hacer la cuenta que lo tiene. No hay un catálogo fijo de roles (Decisión 5).
- **Permiso**: autorización sobre un módulo del producto (p. ej., gestionar eventos). Es la unidad con la que se arman los roles; su catálogo corresponde a los módulos del producto más la administración de usuarios y roles.
- **Sesión**: acceso en curso de una cuenta autenticada. Nace al iniciar sesión y termina al cerrar sesión, por inactividad o cuando la cuenta se desactiva.

Relaciones: un rol agrupa varios permisos y un permiso puede estar en varios roles; una cuenta tiene roles asignados (uno o varios según la aclaración Q4) y, a través de ellos, permisos efectivos.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de las secciones del panel exige sesión iniciada: 0 accesos logrados a secciones del panel sin iniciar sesión en las pruebas de aceptación.
- **SC-002**: Un miembro del equipo con credenciales válidas inicia sesión y llega al panel en menos de 30 segundos desde que abre la pantalla de acceso.
- **SC-003**: En el 100% de instalaciones nuevas queda un administrador inicial listo para crear usuarios y roles sin ningún paso adicional ajeno al mecanismo definido.
- **SC-004**: 0 escenarios en los que el panel se queda sin al menos una cuenta activa con permiso de administración: toda operación que lo intenta es bloqueada y explicada.
- **SC-005**: Un administrador crea una cuenta, crea un rol y los asocia en menos de 3 minutos.
- **SC-006**: Una cuenta desactivada no logra acceder en el 100% de los intentos, ni con credenciales correctas ni con una sesión que tuviera abierta.
- **SC-007**: El 100% de las operaciones de gestión del panel verifican el permiso correspondiente: 0 operaciones realizadas sin permiso en las pruebas de aceptación.
- **SC-008**: Ningún mensaje de error de acceso revela si una cuenta existe (verificado en el 100% de los mensajes de error de inicio de sesión de las pruebas).
- **SC-009**: Un cambio en los permisos de un rol se refleja en el acceso de sus cuentas desde la primera acción posterior, sin pasos adicionales.
- **SC-010**: Una persona con cuenta cambia su contraseña y vuelve a entrar con la nueva en menos de 1 minuto.

## Out of Scope

Queda explícitamente **fuera del alcance** de F2:

- **Registro público de usuarios**: los usuarios comunes **no crean cuentas ni inician sesión**; solo un administrador crea cuentas (`idea.md` §2). Tampoco hay autoalta, invitaciones abiertas ni cuentas de visitante.
- **Gestión del contenido de los módulos** (portada e información general, eventos, actividades, grupos de conexión, ministerios, donaciones, noticias y galería, medios): pertenece a F3–F9. F2 solo define los permisos que esos módulos usarán después.
- **Identificación con cuentas externas** (redes sociales, cuentas de correo como único acceso sin contraseña, etc.) y cualquier acceso sin contraseña.
- **Recuperación o notificaciones por SMS, WhatsApp u otros canales**: la resolución de credenciales se limita a lo que defina la aclaración Q2.
- **Registro de auditoría de acciones** (quién hizo qué cambio y cuándo) más allá de lo que el sistema registre para su propio diagnóstico.
- **Bilingüismo del panel de administración**: la Decisión 6 aplica al sitio público (F3–F9); el panel se entrega en español (ver Assumptions). De quererse el panel bilingüe, sería una ampliación.
- **Despliegue a producción** ni a entornos distintos de los ya manejados por el proyecto: sigue pendiente a nivel de proyecto.
- **Diseño visual del sitio público** (identidad de marca, estilos, animaciones): llega con F3.

## Aclaraciones pendientes

Cinco preguntas quedan abiertas porque no tienen una respuesta única razonable y su respuesta cambia el alcance o el comportamiento (Q1–Q4); Q5 es un caso límite menor que ya tiene default documentado. Se resuelven con el humano (fase `clarify`) **antes** de la planificación técnica. Las marcas `[NECESITA ACLARACIÓN]` están en los requisitos correspondientes; aquí se listan con su contexto y sus opciones.

### Q1 — Mecanismo del administrador inicial (FR-007)

**Contexto**: la Decisión 5 del roadmap deja a esta spec "cómo queda garantizado el arranque con un administrador inicial". La garantía en sí está en FR-007 y FR-008; lo que falta es el mecanismo concreto.

**Opciones**:

| Opción | Mecanismo | Implicación |
|---|---|---|
| A | Credenciales iniciales definidas por quien instala el sistema | Requiere que quien instala elija y proteja esas credenciales; el primer acceso es manual |
| B | Una acción de inicialización ejecutada una sola vez | Requiere que alguien ejecute esa acción y que no pueda repetirse ni abusarse de ella |
| C | Asistente de primer acceso en la propia pantalla de acceso | La pantalla de acceso debe poder distinguir "sin administradores" y proteger ese primer alta |

**También confirmar**: la regla anti-bloqueo de FR-008 (siempre al menos una cuenta activa con permiso de administración).

### Q2 — Contraseñas: entrega inicial, recuperación y política (FR-010)

**Contexto**: un administrador crea las cuentas (US3), pero la spec no define cómo recibe su contraseña la persona ni qué pasa si la olvida; tampoco define la política de contraseñas.

**Opciones**:

| Opción | Contraseña inicial y recuperación | Implicación |
|---|---|---|
| A | Auto-servicio con mensaje al correo de la persona | Requiere un servicio de correo y que la persona tenga acceso a ese correo |
| B | Restablecimiento hecho por un administrador | Sin servicio de correo; la persona depende de un administrador para recuperar el acceso |
| C | Mezcla: auto-servicio para la recuperación y administrador para la alta | Compone ambas; sigue requiriendo servicio de correo |

**Y además**: reglas de la contraseña (por ejemplo, longitud mínima y si exige letras, números o símbolos). Mientras se resuelve, se toma como mínimo provisional al menos 8 caracteres (ver Assumptions).

### Q3 — Ciclo de vida de las cuentas: desactivar y eliminar (FR-013)

**Contexto**: `idea.md` §2 dice que el administrador puede "crear más usuarios o **eliminarlos** o desactivarlos"; el roadmap (Decisión 4 y fila F2) solo define el estado activo/inactivo. La spec define "desactivar" (bloquear acceso, conservar datos, FR-012), pero no puede definir "eliminar" sin decidir qué pasa con lo que esa persona creó.

**Opciones**:

| Opción | Alcance de la eliminación | Implicación |
|---|---|---|
| A | No se eliminan cuentas: solo se desactivan | Lo más simple; los datos nunca se pierden |
| B | Se eliminan, conservando el contenido que la persona creó | Hay que decidir a quién se atribuye ese contenido |
| C | Se eliminan junto con su contenido | Pérdida de información; riesgo de borrar material del sitio |

**Nota**: lo que ocurre con el contenido creado por una cuenta eliminada afecta sobre todo a F4–F9 (todavía no construidos).

### Q4 — Roles: asignación y ciclo de vida (FR-017)

**Contexto**: la Decisión 5 permite crear roles con permisos combinados libremente, pero no define si una cuenta puede tener varios roles ni qué se puede hacer con un rol ya creado.

**Preguntas concretas**:

1. ¿Una cuenta puede tener **varios roles a la vez** (con permisos que se suman) o **solo uno**?
2. ¿Un rol se puede **editar** (cambiarle nombre y permisos) y **eliminar**?
3. Si un rol se puede eliminar, ¿qué ocurre con las cuentas que ya lo tenían asignado?

### Q5 — Normalización de nombres y correos (caso límite)

**Contexto**: el caso límite de cuentas o roles que "solo cambian en mayúsculas o espacios accidentales" toca una regla de datos que no está definida en ningún documento.

**Pregunta**: ¿un correo o un nombre de rol que solo difiere en mayúsculas o en espacios sobrantes debe tratarse como el mismo (y rechazarse como duplicado) o como distinto?

> Nota: Q5 se incluye por ser una ambigüedad real detectada al escribir los casos límites, aunque su impacto es menor que el de Q1–Q4. Si el humano prefiere no decidirla, se aplicará el default de Assumptions.

## Assumptions

Defaults razonables tomados donde la descripción no fija una regla; **son provisionales y revisables** por el humano. Los puntos que cambian el alcance no se asumen: están marcados como `[NECESITA ACLARACIÓN]`.

- **Identificación de la cuenta**: la cuenta se identifica con el correo electrónico de la persona (práctica estándar); no hay "nombres de usuario" aparte. Cada correo pertenece a una sola cuenta.
- **Significado de "desactivar"**: bloquear el acceso de inmediato (también las sesiones ya abiertas) y conservar todos los datos de la cuenta. Desactivar no borra nada. La eliminación, si existe, se decide en Q3.
- **Política de contraseñas (provisional, hasta Q2)**: al menos 8 caracteres. La contraseña nunca se muestra ni se devuelve; su resguardo sigue las reglas de seguridad de la constitución §IV.
- **Sesión**: existe cierre de sesión manual (US1) y expiración por inactividad (FR-005); el periodo provisional de inactividad es de 30 minutos, ajustable por el cliente. El detalle de cómo se materializa la sesión tiene una decisión pendiente heredada de F1 (`specs/001-estructura-base/estado.md`, decisión D-A7); esta spec no la decide.
- **Unidad del permiso**: un permiso corresponde a un módulo completo (lectura literal de la Decisión 5: "permisos por módulo"), no a acciones sueltas dentro del módulo (crear, editar, borrar). Si hiciera falta esa distinción, sería una ampliación.
- **Catálogo de permisos de F2**: los módulos del producto del roadmap más la administración de usuarios y roles (FR-015). Los permisos de módulos que aún no existen quedan reservados y no producen acceso a nada hasta que esos módulos se construyan (F4–F9).
- **"Administrador" no es un rol fijo**: es toda cuenta con el permiso de administrar usuarios y roles; este permiso es el que define quién gestiona cuentas y roles. El administrador inicial nace con todos los permisos.
- **Un rol tiene al menos un permiso** y su nombre es único (un rol vacío o duplicado no aporta nada y confunde).
- **Idioma del panel**: se entrega en español, que es el idioma del equipo que lo usa. El bilingüismo del roadmap (Decisión 6) aplica al sitio público de F3–F9. De quererse el panel en ambos idiomas, sería una ampliación.
- **Límite de intentos fallidos** (FR-006): medida estándar contra la prueba masiva de contraseñas; su configuración concreta (cuántos intentos y por cuánto tiempo) la define la planificación técnica dentro de una práctica estándar.
- **Sin cuentas compartidas**: cada persona del equipo debe tener su propia cuenta; el uso compartido de credenciales queda fuera del modelo.
