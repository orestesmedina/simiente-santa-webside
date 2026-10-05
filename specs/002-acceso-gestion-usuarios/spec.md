# Feature Specification: Acceso y gestión de usuarios (F2)

**Feature Branch**: `002-acceso-gestion-usuarios`

**Created**: 2026-10-04

**Status**: Aprobada con aclaraciones resueltas (2026-10-04); **cambio de alcance: auditoría** incorporado por decisión del humano (2026-10-04) — pendiente de aprobación del añadido

**Input**: User description: "F2 — Acceso y gestión de usuarios (roadmap §3, `docs/producto/roadmap.md`): El equipo de la iglesia inicia sesión en un panel de administración; los administradores crean usuarios, los activan o desactivan, y crean roles con los permisos por módulo que necesiten (p. ej., un rol de contenido, un rol de ministerios)." Decisiones aplicables del roadmap: **Decisión 4** (usuarios activo/inactivo) y **Decisión 5** (permisos por módulo agrupados en roles, sin catálogo fijo de roles; el administrador crea roles nuevos y les asigna permisos de forma independiente; **la spec de F2 define cómo queda garantizado el arranque con un administrador inicial**). Restricción de `idea.md` §2: los usuarios comunes no se registran; solo el administrador crea cuentas. Aclaraciones Q1–Q5 resueltas con el humano el 2026-10-04 (ver *Aclaraciones (resueltas)*). **Cambio de alcance aprobado por el humano el 2026-10-04**: se añade la **auditoría de F2** —último acceso por cuenta, historial de inicios de sesión (exitosos y fallidos) y registro de acciones administrativas, consultables en el panel de solo lectura— (ver *Decisiones adicionales confirmadas por el humano el 2026-10-04*).

> **Nota de alcance y terminología**: esta funcionalidad es la puerta del **panel de administración** que usarán las funcionalidades F3–F9. Se distinguen dos personas distintas: el **visitante** (el "usuario común" de `idea.md` §2: ve el sitio público, no tiene cuenta y no se registra) y el **usuario del panel** (el "usuario del sistema" y el "usuario administrador" de `idea.md` §2: tiene cuenta, inicia sesión y gestiona el contenido). En esta spec, "usuario" significa siempre **usuario del panel** salvo que se diga lo contrario. F2 entrega: acceso autenticado al panel, gestión de cuentas (crear, editar, activar/desactivar), gestión de roles con permisos por módulo y la **auditoría de F2** (último acceso por cuenta, historial de inicios de sesión y registro de acciones administrativas de gestión, consultables en el panel). El registro de auditoría cubre solo el ámbito de F2 (accesos al panel y gestión de cuentas y roles); "origen" significa el **origen del acceso**, la dirección desde la que se hizo el intento. El contenido de los módulos (portada, eventos, actividades, grupos, ministerios, donaciones, noticias, medios) llega en F3–F9; aquí solo se definen los permisos que esos módulos usarán.

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
6. **Dado** una sesión abierta sin actividad, **cuando** transcurren 30 minutos de inactividad, **entonces** la sesión termina y debo iniciar sesión de nuevo para continuar.
7. **Dado** una sesión abierta con actividad continua, **cuando** alcanza 1 hora desde que se inició, **entonces** la sesión termina igualmente y debo iniciar sesión de nuevo para continuar.

---

### User Story 2 - Garantizar el arranque con un administrador inicial (Priority: P1)

Como responsable del sistema, quiero que el arranque del sistema deje garantizado al menos un administrador inicial con permisos completos mediante una acción de inicialización única, para que siempre exista alguien que pueda crear usuarios y roles sin depender de nadie más y sin que esa puesta en marcha pueda repetirse ni abusarse.

**Why this priority**: La Decisión 5 del roadmap deja expresamente a esta spec cómo se garantiza el arranque. Sin esa garantía, el panel podría nacer sin nadie capaz de administrarlo y toda la gestión de usuarios y roles quedaría inaccesible o tendría que resolverse fuera del producto.

**Independent Test**: Se prueba de forma independiente en una instalación nueva (sin cuentas): se ejecuta la acción de inicialización única y se comprueba que existe un administrador inicial capaz de entrar y crear usuarios y roles; se intenta ejecutar esa acción de nuevo y se comprueba que el sistema la impide; después se intenta dejar el panel sin administración activa y se comprueba que el sistema lo impide. Entrega valor por sí sola: el panel siempre es administrable desde el primer minuto.

**Acceptance Scenarios**:

1. **Dado** una instalación nueva sin cuentas de usuario, **cuando** se ejecuta la acción de inicialización única de puesta en marcha, **entonces** queda creado el administrador inicial con permisos completos.
2. **Dado** que la acción de inicialización ya se ejecutó, **cuando** alguien intenta ejecutarla de nuevo, **entonces** el sistema la impide y no crea ninguna cuenta adicional.
3. **Dado** el administrador inicial, **cuando** inicia sesión por primera vez, **entonces** puede crear usuarios y roles de inmediato, sin pasos adicionales.
4. **Dado** el panel con un único administrador activo, **cuando** intento desactivarlo, **entonces** el sistema lo impide con un mensaje que explica que debe quedar al menos un administrador activo.
5. **Dado** el panel con varios administradores activos, **cuando** desactivo uno de ellos, **entonces** la operación se completa con normalidad mientras quede al menos otro administrador activo.
6. **Dado** el panel con un único administrador activo, **cuando** intento cambiarle el rol de forma que pierda el permiso de administrar usuarios y roles, **entonces** el sistema lo impide por la misma regla.

---

### User Story 3 - Crear cuentas para el equipo (Priority: P1)

Como administrador, quiero crear cuentas para las personas del equipo indicando su nombre, sus apellidos, su correo, su número de teléfono, el rol que les corresponde y su contraseña inicial, para que cada quien trabaje con su propia cuenta y nadie comparta credenciales.

**Why this priority**: La gestión de cuentas es una de las dos mitades del alcance de F2 (`idea.md` §2 y Decisión 4). Sin cuentas propias no hay roles que asignar ni acceso diferenciado, y seguiríamos compartiendo un único acceso.

**Independent Test**: Se prueba de forma independiente con un administrador autenticado: crea una cuenta válida y aparece en el listado en estado activo; repite la operación con un correo ya en uso (también escrito con otras mayúsculas o con espacios sobrantes) y con datos inválidos, y el sistema la impide en ambos casos; comprueba que el titular debe cambiar la contraseña inicial al entrar. Entrega valor por sí sola: el equipo ya puede tener cuentas individuales.

**Acceptance Scenarios**:

1. **Dado** un administrador autenticado, **cuando** creo una cuenta con datos válidos y un rol asignado, **entonces** la cuenta queda creada en estado activo y aparece en el listado de usuarios.
2. **Dado** un administrador, **cuando** intento crear una cuenta con un correo que ya está en uso —aunque lo escriba con otras mayúsculas o con espacios sobrantes—, **entonces** el sistema lo impide con un mensaje claro y no queda ninguna cuenta duplicada.
3. **Dado** un formulario de creación con datos inválidos o incompletos (por ejemplo, un correo mal formado, un teléfono con un formato que no es telefónico o un campo obligatorio vacío —nombre, apellidos, correo o teléfono—), **cuando** envío el formulario, **entonces** el sistema indica qué corregir y no crea la cuenta.
4. **Dado** una persona sin el permiso de administrar usuarios y roles, **cuando** intenta crear una cuenta, **entonces** el sistema impide la operación.
5. **Dado** un administrador, **cuando** creo una cuenta indicando un rol que no existe, **entonces** el sistema lo impide indicando que debe elegir un rol válido.
6. **Dado** un administrador, **cuando** creo una cuenta y defino su contraseña inicial, **entonces** su titular debe cambiar esa contraseña al entrar por primera vez antes de usar el panel.

---

### User Story 4 - Crear roles con permisos por módulo (Priority: P1)

Como administrador, quiero crear roles con un nombre propio y los permisos por módulo que necesiten (p. ej., un rol de contenido con eventos y actividades, sin ministerios), para que cada persona del equipo solo acceda a lo que le corresponde.

**Why this priority**: Es la segunda mitad del alcance de F2 (Decisión 5): sin roles con permisos por módulo, todo el equipo tendría el mismo acceso y no se podría delegar trabajo con seguridad. Es independiente de la gestión de cuentas y se puede construir y probar por separado.

**Independent Test**: Se prueba de forma independiente con un administrador autenticado: crea un rol con dos permisos de módulo combinados libremente, lo asigna a una cuenta y se comprueba que esa cuenta solo puede gestionar esos módulos; se comprueba además que un nombre de rol repetido se rechaza (también si solo difiere en mayúsculas o espacios) y que un rol sin permisos no se puede crear. Entrega valor por sí sola: ya existe delegación de acceso por módulo.

**Acceptance Scenarios**:

1. **Dado** un administrador autenticado, **cuando** creo un rol con un nombre único y uno o más permisos por módulo, **entonces** el rol queda disponible para asignarse a cuentas.
2. **Dado** un administrador, **cuando** creo un rol combinando permisos de forma libre (p. ej., eventos y actividades, sin ministerios), **entonces** el sistema acepta esa combinación sin exigir un catálogo previo de roles.
3. **Dado** un administrador, **cuando** intento crear un rol con un nombre que ya existe —aunque difiera solo en mayúsculas o en espacios sobrantes—, **entonces** el sistema lo impide con un mensaje claro.
4. **Dado** un administrador, **cuando** intento crear un rol sin ningún permiso, **entonces** el sistema lo impide indicando que un rol debe tener al menos un permiso.
5. **Dado** una cuenta cuyo rol no incluye el permiso de administrar usuarios y roles, **cuando** intenta gestionar usuarios o roles, **entonces** el sistema impide la operación.
6. **Dado** una cuenta cuyo rol otorga permisos sobre un módulo, **cuando** esa persona trabaja en el panel, **entonces** solo puede usar los módulos de sus permisos: los demás no le aparecen ni puede forzar su acceso.

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

Como administrador, quiero actualizar un rol existente (su nombre y sus permisos) y eliminar los roles que ya no se usan, para que el acceso del equipo siga reflejando el trabajo real.

**Why this priority**: Las responsabilidades del equipo cambian con el tiempo; sin mantenimiento de roles, los permisos quedan viejos y el panel se vuelve inseguro o incómodo. Depende de la creación de roles (US4), por eso va después.

**Independent Test**: Se prueba de forma independiente con un rol asignado a una cuenta: se le quita un permiso y se comprueba que la cuenta deja de poder usar ese módulo; se le agrega otro y la cuenta lo gana; se le cambia el nombre y el rol conserva sus cuentas. La eliminación se prueba con un rol sin cuentas asignadas (se completa) y con un rol en uso (se impide). Entrega valor por sí sola: los permisos se mantienen al día sin recrear roles.

**Acceptance Scenarios**:

1. **Dado** un rol existente con cuentas asignadas, **cuando** un administrador le quita un permiso, **entonces** las cuentas con ese rol dejan de tener ese acceso a partir de ese momento.
2. **Dado** un rol existente con cuentas asignadas, **cuando** un administrador le agrega un permiso, **entonces** las cuentas con ese rol ganan ese acceso sin necesidad de pasos adicionales.
3. **Dado** un rol existente, **cuando** un administrador le cambia el nombre, **entonces** el rol conserva sus permisos y sus cuentas asignadas y aparece en todo el panel con el nuevo nombre.
4. **Dado** un rol que ya no está asignado a ninguna cuenta, **cuando** un administrador lo elimina, **entonces** el rol desaparece del catálogo sin afectar a ninguna cuenta.
5. **Dado** un rol con cuentas asignadas, **cuando** un administrador intenta eliminarlo, **entonces** el sistema lo impide con un mensaje que explica que primero debe reasignar esas cuentas a otro rol.
6. **Dado** un rol existente, **cuando** un administrador intenta dejarlo sin ningún permiso, **entonces** el sistema lo impide indicando que un rol debe conservar al menos un permiso.
7. **Dado** una persona sin el permiso de administrar usuarios y roles, **cuando** intenta actualizar o eliminar un rol, **entonces** el sistema impide la operación.

---

### User Story 7 - Cambiar mi propia contraseña (Priority: P3)

Como usuario del panel, quiero cambiar mi contraseña desde mi cuenta, para mantener el control de mi acceso.

**Why this priority**: Refuerza la seguridad de las cuentas individuales y completa el ciclo de credenciales, pero no bloquea la gestión de usuarios ni de roles; por eso va al final.

**Independent Test**: Se prueba de forma independiente con una cuenta autenticada: cambia su contraseña indicando la actual y una nueva válida y luego entra con la nueva; con la contraseña actual incorrecta y con una nueva que no cumple la política, el sistema lo impide. Se prueba además la entrada con una contraseña definida por un administrador, que exige cambio inmediato. Entrega valor por sí sola: cada persona mantiene su propia credencial.

**Acceptance Scenarios**:

1. **Dado** una cuenta autenticada, **cuando** cambia su contraseña indicando la actual y una nueva válida, **entonces** el cambio se aplica y puede iniciar sesión con la nueva contraseña.
2. **Dado** una cuenta autenticada, **cuando** indica una contraseña actual incorrecta, **entonces** el sistema impide el cambio con un mensaje claro.
3. **Dado** una cuenta autenticada, **cuando** propone una contraseña que no cumple la política vigente —de menos de 8 o más de 64 caracteres, sin combinar mayúsculas, minúsculas, números y caracteres especiales, o igual a su nombre, a sus apellidos o a su correo (comparación normalizada)—, **entonces** el sistema indica el requisito incumplido y no aplica el cambio.
4. **Dado** una cuenta con la contraseña definida o restablecida por un administrador, **cuando** inicia sesión con ella, **entonces** el sistema le exige cambiarla antes de usar el panel.
5. **Dado** una persona que olvidó su contraseña, **cuando** un administrador le define una contraseña nueva en su cuenta, **entonces** puede entrar con ella y el sistema le exige cambiarla de inmediato.

---

### User Story 8 - Consultar el registro de accesos y acciones administrativas (Priority: P3)

Como administrador, quiero consultar el registro de accesos al panel y de acciones administrativas de gestión —con filtros por cuenta y por rango de fechas— y ver el último acceso de cada cuenta, para saber quién entró, desde dónde, qué se hizo sobre las cuentas y los roles, y cuándo.

**Why this priority**: Es valor de control y acompañamiento: da confianza y permite revisar qué pasó, pero no bloquea el acceso ni la gestión de usuarios y roles, que ya entregan su valor sin él (US1–US7). Por eso va al final, como P3.

**Independent Test**: Se prueba de forma independiente con una cuenta administradora y otra sin ese permiso, tras generar algunos accesos (exitosos y fallidos) y algunas acciones de gestión: la sección muestra ambos historiales con fecha y hora, resultado e IP de origen; los filtros por cuenta y por rango de fechas y la paginación funcionan; la ficha de cada cuenta muestra su último acceso; un intento fallido contra un correo que no existe queda registrado sin asociarse a ninguna cuenta; los registros no se pueden editar ni borrar; y la cuenta sin permiso no puede abrir la sección. Entrega valor por sí sola: el administrador ve qué pasó en el panel.

**Acceptance Scenarios**:

1. **Dado** un administrador autenticado, **cuando** abre la sección de registro, **entonces** ve el historial de inicios de sesión (exitosos y fallidos) y el de acciones administrativas, con fecha y hora, resultado y los datos propios de cada registro.
2. **Dado** la sección de registro, **cuando** filtro por cuenta y por rango de fechas, **entonces** veo únicamente los registros de esa cuenta dentro de ese periodo.
3. **Dado** un periodo con muchos registros, **cuando** recorro la sección, **entonces** los registros se muestran paginados y puedo pasar de una página a otra sin perder los filtros aplicados.
4. **Dado** la sección de registro, **cuando** intento editar o borrar un registro, **entonces** no puedo: el registro es de solo lectura.
5. **Dado** una cuenta sin el permiso de administrar usuarios y roles, **cuando** intenta abrir la sección de registro, **entonces** el sistema impide el acceso con un mensaje claro.
6. **Dado** un administrador consultando una cuenta, **cuando** mira su ficha, **entonces** ve la fecha y el origen de su último acceso exitoso, o una indicación de que esa cuenta aún no ha iniciado sesión.
7. **Dado** un inicio de sesión fallido contra un correo que no corresponde a ninguna cuenta, **cuando** se registra el intento, **entonces** queda registrado con su fecha y hora, su resultado y su IP de origen, sin asociarse a ninguna cuenta y sin que ningún mensaje a la persona revele si la cuenta existe.
8. **Dado** un administrador que realiza una acción sensible de gestión (crear, editar, activar o desactivar una cuenta, restablecer la contraseña de una cuenta, o crear, editar o eliminar un rol), **cuando** la realiza, **entonces** la acción queda registrada con quién la hizo, qué hizo, sobre qué y cuándo.

---

### Edge Cases

- **Cuenta desactivada con sesión abierta**: pierde el acceso de inmediato (US5, escenario 2); no debe poder seguir actuando en el panel con esa sesión.
- **Último administrador activo**: ninguna operación (desactivarlo o cambiarle el rol de forma que pierda el permiso de administración) puede dejar el panel sin administración; el sistema lo impide con un mensaje que explica la restricción.
- **Permisos sobre módulos que aún no existen** (portada e información general, eventos, actividades, grupos, ministerios, donaciones, noticias, medios: F3–F9): quedan reservados desde F2 y no dan acceso a nada hasta que esos módulos se construyan.
- **Cuenta sin ningún permiso de módulo** (solo con acceso al panel): entra pero no ve ninguna sección de gestión; esto es válido y no debe romper el panel.
- **Dos personas creando a la vez dos cuentas con el mismo correo**: solo debe quedar creada una; la segunda operación se rechaza como duplicado.
- **Un correo o un nombre de rol que solo difiere en mayúsculas o en espacios sobrantes**: se normaliza, se trata como el mismo y se rechaza como duplicado (decisión Q5); el sistema debe indicar el problema en lugar de guardar una cuenta o un rol prácticamente duplicado.
- **Asignar un rol a una cuenta que ya tiene otro**: el nuevo rol reemplaza al anterior —una cuenta tiene un solo rol (decisión Q4)— y los permisos efectivos son exactamente los del rol asignado, nunca la suma de los dos.
- **Eliminar un rol con cuentas asignadas**: el sistema lo impide y explica que primero deben reasignarse esas cuentas; un rol solo se elimina cuando ninguna cuenta lo tiene asignado (decisión Q4), de modo que ninguna cuenta queda nunca sin rol por una eliminación.
- **Dejar un rol sin ningún permiso** (al crearlo o editándolo): no se permite; un rol debe conservar al menos un permiso (decisión Q4).
- **Repetir la acción de inicialización del administrador**: se impide; la acción es única y no debe poder repetirse ni abusarse de ella (decisión Q1).
- **Cuenta con contraseña definida o restablecida por un administrador**: la persona debe cambiarla al entrar antes de usar el panel (decisión Q2); esa contraseña inicial o restablecida también debe cumplir la política de contraseñas (FR-010).
- **Contraseña que no cumple la política** (de menos de 8 o más de 64 caracteres, sin combinar mayúsculas, minúsculas, números y caracteres especiales, o igual al nombre, a los apellidos o al correo de la persona, con comparación normalizada): se rechaza indicando el requisito incumplido, tanto cuando la define un administrador como cuando la cambia la persona (FR-010).
- **Cadena de intentos fallidos de inicio de sesión** (alguien probando contraseñas): el contador de fallos se incrementa con cada intento fallido. El **5.º fallo** recibe el mismo mensaje de error genérico de siempre —sin revelar si la cuenta existe— y **activa el bloqueo temporal** de **15 minutos**; desde el **6.º intento** y durante esos 15 minutos el acceso queda bloqueado y cada intento recibe el mensaje de bloqueo temporal que indica esos 15 minutos (FR-006, decisión confirmada el 2026-10-04). Pasados esos 15 minutos la persona puede volver a intentar iniciar sesión.
- **Sesión que termina mientras se llena un formulario** (por 30 minutos de inactividad o por alcanzar la hora máxima desde su inicio): la acción se interrumpe y el sistema pide iniciar sesión de nuevo; no debe aplicarse a medias ni perderse sin aviso.
- **Sesión con actividad continua que alcanza la hora máxima**: 1 hora desde el inicio es el techo absoluto; aunque la persona siga trabajando sin parar, la sesión termina y debe iniciar sesión de nuevo. La inactividad (30 minutos) solo puede cerrarla antes, nunca extenderla.
- **Pérdida de acceso de la única persona que sabe su contraseña**: un administrador le define una contraseña nueva en su cuenta y la persona la cambia al entrar (decisión Q2). No hay recuperación por auto-servicio con correo en el MVP.
- **Intento de inicio de sesión contra un correo que no existe**: se registra el intento como fallido, con su fecha y hora y su IP de origen, **sin asociarlo a ninguna cuenta y sin crear nada**; el mensaje a la persona sigue sin revelar si la cuenta existe (coherente con FR-003). El registro no debe generar "cuentas fantasma" a partir de correos inventados.
- **Acción administrativa que no se completa** (falla o es denegada por falta de permiso): el intento queda registrado con su resultado y la operación no se aplica; para la auditoría importa saber qué se intentó, quién, sobre qué y cuándo.
- **Intentos de inicio de sesión durante el bloqueo temporal** (FR-006, desde el 6.º intento y durante 15 minutos): cada intento recibe el mensaje de bloqueo temporal y también se registra como intento fallido; el bloqueo no suspende ni borra el registro.
- **Cuenta que nunca ha iniciado sesión**: su ficha no muestra ningún último acceso e indica que aún no hay accesos; no debe aparecer una fecha vacía ni inventada.
- **Cuenta desactivada**: su historial de accesos y las acciones registradas sobre ella se conservan íntegros (coherente con FR-012 y FR-013); desactivar no borra nada del registro.
- **Contraseñas y credenciales**: nunca aparecen en ningún registro; de un restablecimiento se registra quién lo hizo, sobre qué cuenta y cuándo, nunca la contraseña definida (FR-010, constitución §IV).
- **Gran volumen de registros**: la consulta se pagina y se filtra por cuenta y por rango de fechas; el registro no se edita ni se borra desde el panel, y no hay purga en el MVP (ver Out of Scope y Assumptions).
- **Cambio de la propia contraseña (US7)**: no es una acción administrativa y no se registra en el historial de acciones; los accesos de esa persona sí quedan en el historial de accesos.

### Errores esperados

Todos los mensajes deben ser comprensibles para personas no técnicas y nunca deben exponer información interna del sistema (coherente con FR-012 y FR-013 de F1).

| Situación | Qué debe ver la persona |
|---|---|
| Correo o contraseña incorrectos | Mensaje de error genérico (sin revelar si la cuenta existe) |
| Cuenta inactiva | Mensaje que indica que ese acceso está desactivado |
| Datos inválidos o incompletos en un formulario | Indicación de qué campo corregir, sin crear nada |
| Teléfono con un formato que no es telefónico (o con menos de 7 dígitos) | Mensaje que indica cómo corregir el número de teléfono, sin crear la cuenta |
| Correo de cuenta ya en uso (también si solo difiere en mayúsculas o espacios) | Mensaje claro de duplicado, sin crear la cuenta |
| Nombre de rol ya en uso (también si solo difiere en mayúsculas o espacios) | Mensaje claro de duplicado, sin crear el rol |
| Rol sin ningún permiso al crearlo o editarlo | Mensaje que indica que un rol debe tener al menos un permiso |
| Intento de eliminar un rol con cuentas asignadas | Mensaje que explica que primero deben reasignarse esas cuentas |
| Intento de repetir la acción de inicialización del administrador | Mensaje que indica que esa acción ya se hizo y no puede repetirse |
| Operación sin el permiso correspondiente | Mensaje claro de acceso denegado, sin detalles internos |
| Operación que dejaría el panel sin administrador | Mensaje que explica la restricción y cómo proceder |
| Sesión terminada (por cerrarla, por 30 minutos de inactividad o por alcanzar la hora máxima) | Mensaje que pide iniciar sesión de nuevo |
| Contraseña nueva que no cumple la política (de menos de 8 o más de 64 caracteres, sin combinar mayúsculas, minúsculas, números y caracteres especiales, o igual al nombre, a los apellidos o al correo, con comparación normalizada) | Mensaje con el requisito incumplido |
| 5.º intento fallido de inicio de sesión (quinto fallo) | Mensaje de error genérico igual que en cualquier fallo (sin revelar si la cuenta existe); ese fallo activa el bloqueo temporal |
| Intento de inicio de sesión durante el bloqueo temporal (desde el 6.º intento, durante 15 minutos) | Mensaje que indica que el acceso queda bloqueado temporalmente durante 15 minutos |
| Filtros de consulta del registro sin resultados | Indicación clara de que no hay registros que coincidan con esos filtros, sin romper la pantalla |
| Error inesperado del sistema | Mensaje genérico sin información interna |

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema DEBE exponer un panel de administración cuyas secciones solo son accesibles con una sesión iniciada; todo intento de acceso sin sesión DEBE dirigir a la pantalla de acceso.
- **FR-002**: El sistema DEBE autenticar a las personas del equipo con su correo y su contraseña, y DEBE rechazar el acceso de las cuentas inactivas aunque las credenciales sean correctas.
- **FR-003**: El mensaje de un inicio de sesión fallido DEBE ser genérico y NUNCA DEBE revelar si la cuenta existe; la contraseña NUNCA DEBE mostrarse ni devolverse en ninguna respuesta.
- **FR-004**: El sistema DEBE permitir cerrar sesión en cualquier momento desde el panel, terminando el acceso de esa sesión.
- **FR-005**: La sesión DEBE terminar como máximo **1 hora** después de iniciarse —incluso con actividad continua— y, además, DEBE terminar tras **30 minutos de inactividad** —que puede acortarla antes—; en cualquiera de los dos casos, todo acceso posterior DEBE exigir un nuevo inicio de sesión (decisión confirmada por el humano el 2026-10-04).
- **FR-006**: El sistema DEBE limitar los intentos fallidos de inicio de sesión para impedir la prueba masiva de contraseñas. El contador de fallos DEBE incrementarse con cada inicio de sesión fallido. El **5.º fallo** DEBE responder con el mensaje de error genérico de FR-003 —sin revelar si la cuenta existe— y DEBE **activar el bloqueo temporal** de **15 minutos** (decisión confirmada por el humano el 2026-10-04). Desde el **6.º intento** y durante esos 15 minutos, todo intento DEBE ser rechazado con el mensaje de bloqueo temporal que indica esos 15 minutos, también sin revelar si la cuenta existe (coherente con FR-003). Pasados esos 15 minutos el sistema DEBE permitir de nuevo el inicio de sesión.
- **FR-007**: El primer administrador (cuenta con permisos completos, capaz de crear usuarios y roles de inmediato) DEBE crearse mediante una acción de inicialización de puesta en marcha que solo puede ejecutarse una sola vez: el sistema DEBE impedir que se repita y DEBE impedir cualquier uso abusivo de ella (decisión Q1).
- **FR-008**: El sistema DEBE impedir que el panel se quede sin al menos una cuenta activa con permiso de administrar usuarios y roles: ninguna operación (desactivar una cuenta, cambiar su rol o quitarle ese permiso) DEBE poder dejarlo sin administración (regla confirmada).
- **FR-009**: Los administradores DEBEN poder crear cuentas del equipo indicando nombre, apellidos, correo, número de teléfono, rol asignado y contraseña inicial (datos de la cuenta confirmados por el humano el 2026-10-04). El nombre y los apellidos son campos obligatorios separados; el correo y el teléfono también son obligatorios, y el teléfono DEBE tener un formato telefónico razonable: dígitos, con espacios, guiones o paréntesis como separadores habituales, un prefijo internacional opcional y al menos **7 dígitos**. El sistema DEBE rechazar correos ya en uso, teléfonos con un formato que no es telefónico, roles inexistentes y datos inválidos o incompletos, indicando qué corregir. Para detectar duplicados, los correos DEBEN compararse normalizados: uno que solo difiere en mayúsculas o en espacios sobrantes se trata como el mismo (decisión Q5).
- **FR-010**: La contraseña inicial de una cuenta nueva DEBE ser definida por un administrador al crearla, y la persona DEBE cambiarla al entrar por primera vez antes de usar el panel. Cuando alguien olvida su contraseña o pierde el acceso, la recuperación DEBE resolverse con un restablecimiento hecho por un administrador (que define una contraseña nueva que la persona cambia al entrar); la recuperación por auto-servicio con correo queda fuera del MVP. Toda contraseña —la inicial definida por un administrador, las restablecidas por un administrador y las nuevas elegidas por la persona— DEBE cumplir la política vigente (decisión Q2, confirmada por el humano el 2026-10-04): DEBE tener entre **8 y 64 caracteres** (mínimo 8, máximo 64), DEBE combinar **mayúsculas, minúsculas, números y caracteres especiales**, y DEBE ser **distinta del nombre, de los apellidos y del correo** de la persona. La regla es de **igualdad con comparación normalizada** (como en la decisión Q5), no de contenido: la contraseña DEBE ser distinta de esos datos, pero puede contenerlos; lo que se rechaza es que sea igual a cualquiera de ellos.
- **FR-011**: Los administradores DEBEN poder editar los datos de una cuenta (nombre, apellidos, correo, número de teléfono y rol asignado) y su estado (activo/inactivo), con las mismas validaciones que al crearla.
- **FR-012**: Al desactivar una cuenta, el sistema DEBE bloquear su acceso de inmediato —también las sesiones ya abiertas— y DEBE conservar todos sus datos.
- **FR-013**: El sistema NUNCA DEBE eliminar cuentas: la única forma de retirar el acceso es desactivarlas (FR-012) y todos sus datos DEBEN conservarse siempre. Lo que `idea.md` §2 llama "eliminar" se resuelve como desactivación en el MVP (decisión Q3).
- **FR-014**: Los administradores DEBEN poder crear roles compuestos por un nombre único y uno o más permisos por módulo, combinados libremente (no existe un catálogo fijo de roles); un rol NUNCA DEBE quedar sin al menos un permiso. Los nombres de rol DEBEN compararse normalizados: uno que solo difiere en mayúsculas o en espacios sobrantes se trata como el mismo (decisión Q5).
- **FR-015**: Los permisos DEBEN corresponder a los módulos del producto: portada e información general, eventos, actividades, grupos de conexión, ministerios, donaciones, noticias y galería, medios, y administración de usuarios y roles; cada permiso autoriza la gestión completa de su módulo.
- **FR-016**: El sistema DEBE verificar el permiso correspondiente en cada operación del panel según el rol de la cuenta, y DEBE denegar con un mensaje claro toda operación no autorizada.
- **FR-017**: Una cuenta DEBE tener un solo rol asignado (decisión Q4). Los roles DEBEN poder editarse (nombre y permisos) y DEBEN poder eliminarse únicamente cuando ninguna cuenta los tiene asignados: si hay cuentas con ese rol, el sistema DEBE impedir la eliminación y explicar que primero deben reasignarse esas cuentas.
- **FR-018**: Todo cambio en los permisos de un rol DEBE reflejarse de inmediato en el acceso de las cuentas que lo tienen, sin pasos adicionales.
- **FR-019**: El sistema DEBE mostrar el listado de cuentas con su estado (activo/inactivo), su correo y su rol asignado.
- **FR-020**: Las personas con cuenta en el panel DEBEN poder cambiar su propia contraseña.
- **FR-021** *(cambio de alcance: auditoría, 2026-10-04)*: La ficha de cada cuenta DEBE mostrar la fecha y el origen (IP) de su **último acceso exitoso** al panel; una cuenta que aún no ha iniciado sesión DEBE indicarlo sin mostrar ningún acceso inventado.
- **FR-022**: El sistema DEBE registrar cada **intento de inicio de sesión** —exitoso y fallido— con fecha y hora, resultado e **IP de origen**, asociado a la cuenta cuando se pueda identificar. Un intento con un correo que no corresponde a ninguna cuenta DEBE registrarse igualmente como fallido, sin asociarse a ninguna cuenta y sin crear nada, y el mensaje a la persona NUNCA DEBE revelar si la cuenta existe (coherente con FR-003).
- **FR-023**: El sistema DEBE registrar cada **acción administrativa sensible de gestión** con **quién la hizo, qué hizo, sobre qué** (cuenta o rol) **y cuándo**, y su resultado: crear, editar, activar o desactivar una cuenta; restablecer la contraseña de una cuenta; crear, editar o eliminar un rol. La creación del administrador inicial (FR-007) cuenta como creación de cuenta y DEBE registrarse. Si la operación no se completa (falla o se deniega), el intento DEBE registrarse igualmente con su resultado. El cambio de la propia contraseña (FR-020) no es una acción administrativa y no se registra en este historial.
- **FR-024**: El panel DEBE exponer una sección de registro —historial de accesos y de acciones administrativas— accesible **solo a las cuentas con el permiso de administrar usuarios y roles** (decisión de permiso documentada en *Decisiones adicionales*): DEBE mostrar ambos historiales, DEBE permitir filtrar por cuenta y por rango de fechas y DEBE paginar los resultados. Las cuentas sin ese permiso DEBEN ver denegado el acceso con un mensaje claro.
- **FR-025**: El registro DEBE ser de **solo lectura** desde el panel: el sistema NUNCA DEBE permitir editar ni borrar registros por ninguna vía de la interfaz, y los registros DEBEN conservarse aunque la cuenta se desactive, cambie de rol o se le editen los datos (coherente con FR-012 y FR-013).
- **FR-026**: NINGÚN registro DEBE contener contraseñas ni otros datos de credenciales: de un inicio de sesión solo su resultado, y de un restablecimiento de contraseña quién lo hizo, sobre qué cuenta y cuándo (FR-010, constitución §IV).

### Key Entities

- **Cuenta del panel**: persona del equipo de la iglesia con acceso autenticado al panel. Se identifica por su correo y tiene como datos de la cuenta su nombre, sus apellidos, su correo y su número de teléfono (todos obligatorios), un estado (activo/inactivo) y un rol asignado. Su ficha muestra además la fecha y el origen de su último acceso exitoso. No debe confundirse con el "usuario común" de `idea.md` §2: el visitante del sitio público no tiene cuenta ni se registra.
- **Rol**: nombre único creado por el administrador que agrupa al menos un permiso; define qué puede hacer la cuenta que lo tiene. Puede editarse (nombre y permisos) y eliminarse solo cuando no está en uso. No hay un catálogo fijo de roles (Decisión 5).
- **Permiso**: autorización sobre un módulo del producto (p. ej., gestionar eventos). Es la unidad con la que se arman los roles; su catálogo corresponde a los módulos del producto más la administración de usuarios y roles.
- **Sesión**: acceso en curso de una cuenta autenticada. Nace al iniciar sesión y termina al cerrar sesión, a los **30 minutos de inactividad**, al alcanzar la duración máxima de **1 hora** desde su inicio o cuando la cuenta se desactiva.
- **Registro de acceso** *(auditoría)*: entrada del historial de inicios de sesión con fecha y hora, resultado (exitoso/fallido) e IP de origen, asociada a la cuenta cuando se puede identificar. Un intento contra un correo que no corresponde a ninguna cuenta queda registrado sin asociación. De los accesos exitosos se deriva el último acceso que muestra la ficha de la cuenta. Es de solo lectura y no contiene credenciales.
- **Registro de acción administrativa** *(auditoría)*: entrada del historial de acciones sensibles de gestión con quién la hizo, qué hizo, sobre qué cuenta o rol, cuándo y con qué resultado —incluidos los intentos que no se completan—. Es de solo lectura y no contiene credenciales.

Relaciones: un rol agrupa varios permisos y un permiso puede estar en varios roles; una cuenta tiene un solo rol y, a través de él, permisos efectivos; un rol solo puede eliminarse cuando ninguna cuenta lo tiene asignado. Cada registro de acceso puede estar asociado a una cuenta (salvo los intentos no identificables, que quedan sin asociación), y cada registro de acción administrativa nombra a la cuenta que lo hizo y al objeto (cuenta o rol) sobre el que se hizo.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de las secciones del panel exige sesión iniciada: 0 accesos logrados a secciones del panel sin iniciar sesión en las pruebas de aceptación.
- **SC-002**: Un miembro del equipo con credenciales válidas inicia sesión y llega al panel en menos de 30 segundos desde que abre la pantalla de acceso.
- **SC-003**: En el 100% de instalaciones nuevas, la única acción de inicialización deja un administrador inicial listo para crear usuarios y roles sin ningún paso adicional, y todo intento de repetirla es bloqueado.
- **SC-004**: 0 escenarios en los que el panel se queda sin al menos una cuenta activa con permiso de administración: toda operación que lo intenta es bloqueada y explicada.
- **SC-005**: Un administrador crea una cuenta, crea un rol y los asocia en menos de 3 minutos.
- **SC-006**: Una cuenta desactivada no logra acceder en el 100% de los intentos, ni con credenciales correctas ni con una sesión que tuviera abierta.
- **SC-007**: El 100% de las operaciones de gestión del panel verifican el permiso correspondiente: 0 operaciones realizadas sin permiso en las pruebas de aceptación.
- **SC-008**: Ningún mensaje de error de acceso revela si una cuenta existe (verificado en el 100% de los mensajes de error de inicio de sesión de las pruebas).
- **SC-009**: Un cambio en los permisos de un rol se refleja en el acceso de sus cuentas desde la primera acción posterior, sin pasos adicionales.
- **SC-010**: Una persona con cuenta cambia su contraseña y vuelve a entrar con la nueva en menos de 1 minuto.
- **SC-011**: 0 intentos de eliminar un rol en uso ni de repetir la inicialización tienen éxito en las pruebas de aceptación, y ningún correo o nombre de rol casi duplicado (solo por mayúsculas o espacios) queda registrado.
- **SC-012** *(auditoría)*: Un administrador encuentra los accesos y las acciones administrativas de una cuenta concreta en menos de 1 minuto, filtrando por cuenta y por rango de fechas, y ve el último acceso de cada cuenta en su ficha.
- **SC-013** *(auditoría)*: El 100% de los inicios de sesión (exitosos y fallidos) y de las acciones administrativas sensibles de las pruebas de aceptación queda registrado con sus datos (fecha y hora, resultado y, en los accesos, IP de origen), y 0 registros pueden editarse o borrarse desde el panel.

## Out of Scope

Queda explícitamente **fuera del alcance** de F2:

- **Registro público de usuarios**: los usuarios comunes **no crean cuentas ni inician sesión**; solo un administrador crea cuentas (`idea.md` §2). Tampoco hay autoalta, invitaciones abiertas ni cuentas de visitante.
- **Eliminación de cuentas**: en el MVP **no se eliminan cuentas**; solo se desactivan y sus datos se conservan siempre (decisión Q3). Aunque `idea.md` §2 menciona "eliminar", la decisión es no eliminar; una eventual eliminación con destino de sus datos sería una ampliación futura.
- **Recuperación de contraseña por auto-servicio con correo** (y cualquier flujo de credenciales que exija un servicio de correo): queda para más adelante como idea futura/backlog. En el MVP la recuperación se resuelve con restablecimiento hecho por un administrador (decisión Q2).
- **Gestión del contenido de los módulos** (portada e información general, eventos, actividades, grupos de conexión, ministerios, donaciones, noticias y galería, medios): pertenece a F3–F9. F2 solo define los permisos que esos módulos usarán después.
- **Identificación con cuentas externas** (redes sociales, cuentas de correo como único acceso sin contraseña, etc.) y cualquier acceso sin contraseña.
- **Recuperación o notificaciones por SMS, WhatsApp u otros canales**: la resolución de credenciales se limita al restablecimiento por un administrador (decisión Q2).
- **Auditoría más allá del ámbito de F2**: el registro cubre los accesos al panel y las acciones administrativas de gestión de cuentas y roles (cambio de alcance del 2026-10-04). Auditar la gestión del contenido de los módulos (portada, eventos, actividades, grupos, ministerios, donaciones, noticias y galería, medios: F3–F9) queda para cuando esos módulos existan.
- **Exportación, alertas y políticas de retención o purga del registro**: no hay exportación del registro, avisos automáticos ni borrado programado de entradas en el MVP; los registros se conservan mientras el sistema funciona (defaults revisables, ver Assumptions).
- **Bilingüismo del panel de administración**: la Decisión 6 aplica al sitio público (F3–F9); el panel se entrega en español (ver Assumptions). De quererse el panel bilingüe, sería una ampliación.
- **Despliegue a producción** ni a entornos distintos de los ya manejados por el proyecto: sigue pendiente a nivel de proyecto.
- **Diseño visual del sitio público** (identidad de marca, estilos, animaciones): llega con F3.

## Aclaraciones (resueltas el 2026-10-04)

**Ninguna aclaración pendiente.** Las cinco preguntas (Q1–Q5) se resolvieron con el humano en la fase `clarify` del 2026-10-04, y ese mismo día el humano confirmó además los datos de la cuenta, la política de contraseñas, los tiempos de sesión y el límite de intentos fallidos (ver *Decisiones adicionales confirmadas por el humano el 2026-10-04*). Queda aquí el registro de la decisión tomada en cada una; el detalle de opciones consideradas se conserva en el historial de esta spec.

### Q1 — Mecanismo del administrador inicial (FR-007) — **Resuelta**

**Decisión (opción B)**: acción de inicialización de puesta en marcha ejecutada **una sola vez**, que crea el primer administrador con permisos completos. No puede repetirse ni abusarse de ella (FR-007, US2 escenarios 1–2). **Confirmada además** la regla anti-bloqueo de FR-008: siempre al menos una cuenta activa con permiso de administrar usuarios y roles.

### Q2 — Contraseñas: entrega inicial, recuperación y política (FR-010) — **Resuelta**

**Decisión (opción B)**: para el MVP, restablecimiento hecho por un administrador, **sin servicio de correo**: el administrador define la contraseña inicial al crear la cuenta y puede reiniciarla cuando alguien pierde el acceso; la persona la cambia al entrar (FR-010, US3 escenario 6, US7 escenarios 4–5). **Política para el MVP (confirmada por el humano el 2026-10-04)**: entre **8 y 64 caracteres** (mínimo 8, máximo 64), la contraseña debe combinar **mayúsculas, minúsculas, números y caracteres especiales** y debe ser **distinta del nombre, de los apellidos y del correo** de la persona (igualdad con comparación normalizada, no de contenido). La **recuperación por auto-servicio con correo queda fuera del MVP** y se anota como idea futura/backlog (ver Out of Scope).

### Q3 — Ciclo de vida de las cuentas: desactivar y eliminar (FR-013) — **Resuelta**

**Decisión (opción A)**: en el MVP **no se eliminan cuentas**; solo se desactivan y todos sus datos se conservan (FR-012, FR-013). Aunque `idea.md` §2 menciona "eliminar", la decisión es no eliminar: toda mención a eliminar cuentas se retira del alcance (US2 escenario 4, US6, Edge Cases).

### Q4 — Roles: asignación y ciclo de vida (FR-017) — **Resuelta**

**Decisión**: una cuenta tiene **un solo rol** (sus permisos efectivos son los de ese rol; al asignar otro, el nuevo reemplaza al anterior). Los roles se pueden **editar** (nombre y permisos) y **eliminar solo si no están en uso**: si hay cuentas asignadas, primero deben reasignarse y el sistema no permite la eliminación. Un rol tiene **al menos un permiso** y **nombre único** (FR-014, FR-017, US6).

### Q5 — Normalización de nombres y correos (caso límite) — **Resuelta**

**Decisión**: un correo o un nombre de rol que solo difiere en mayúsculas o en espacios sobrantes **se trata como el mismo** (se normaliza) y se rechaza como duplicado (FR-009, FR-014, Edge Cases).

### Decisiones adicionales confirmadas por el humano el 2026-10-04

- **Datos de la cuenta (FR-009, FR-011, US3)**: al crear o editar una cuenta se piden **nombre, apellidos, correo y número de teléfono**. El nombre y los apellidos son obligatorios y se piden como campos separados; el correo es obligatorio y sigue siendo la identificación de acceso; el teléfono es **obligatorio** y se valida con un formato telefónico razonable. *[Decisión menor: la regla concreta de formato del teléfono —longitud mínima, separadores y prefijo— es un default revisable; se fijó en FR-009 como al menos 7 dígitos, con separadores habituales y prefijo internacional opcional.]*
- **Política de contraseñas (FR-010, US7.3)**: entre **8 y 64 caracteres** (mínimo 8, máximo 64), combinando **mayúsculas, minúsculas, números y caracteres especiales**, y **distinta del nombre, de los apellidos y del correo** de la persona (igualdad con comparación normalizada, no de contenido). Complementa la decisión Q2 (la contraseña inicial y los restablecimientos los hace un administrador; el auto-servicio por correo sigue fuera del MVP).
- **Tiempos de sesión (FR-005, US1 escenarios 6–7)**: la sesión dura **como máximo 1 hora** desde que se inicia —techo absoluto, incluso con actividad continua— y además **se cierra tras 30 minutos de inactividad**, que puede acortarla antes. En cualquiera de los dos casos se exige iniciar sesión de nuevo.
- **Límite de intentos fallidos (FR-006)**: el contador de fallos se incrementa con cada intento fallido; el **5.º fallo** responde el error genérico (sin revelar si la cuenta existe) y **activa el bloqueo temporal** de **15 minutos**; desde el **6.º intento** y durante esos 15 minutos el acceso queda bloqueado con el mensaje de bloqueo temporal. Pasados esos 15 minutos se puede volver a intentar iniciar sesión.
- **Cambio de alcance: auditoría de F2 (decisión del humano, 2026-10-04)**: la auditoría pasa de Out of Scope a **alcance de F2** (US8, FR-021–FR-026). Se añade: el **último acceso exitoso** por cuenta (fecha y origen), el **historial de inicios de sesión** —exitosos y fallidos— con fecha/hora, resultado e IP de origen, asociado a la cuenta cuando se pueda identificar; el **registro de acciones administrativas sensibles** (crear, editar, activar o desactivar una cuenta, restablecer la contraseña de una cuenta, crear, editar o eliminar un rol) con quién la hizo, qué hizo, sobre qué y cuándo —también cuando la operación no se completa—; y una **sección del panel de solo lectura**, accesible a quien administra usuarios y roles, con filtros por cuenta y por rango de fechas y paginación. Prioridad **P3**: es valor de control y acompañamiento y no bloquea el acceso ni la gestión.
- **Permiso de la sección de registro (decisión de producto documentada, 2026-10-04)**: la sección de auditoría usa **el mismo permiso de administrar usuarios y roles**, sin un permiso propio de "auditoría". Por qué: (1) el catálogo de permisos corresponde a los **módulos del producto** (Decisión 5, FR-015) y la auditoría no es un módulo del producto, sino un instrumento de la administración de usuarios y roles; (2) todo lo que el registro muestra —accesos y cambios sobre cuentas y roles— pertenece a ese ámbito, que quien tiene ese permiso ya puede ver y definir, de modo que un permiso aparte no aportaría separación real de funciones en un equipo pequeño; (3) mantener un único permiso evita ensanchar el catálogo con una entrada ajena a los módulos y simplifica los roles. Si en el futuro hiciera falta separar la consulta de la gestión (p. ej., una revisión externa al equipo), sería una ampliación con un permiso nuevo.

## Assumptions

Defaults razonables tomados donde la descripción no fija una regla; **son provisionales y revisables** por el humano. Los puntos que cambian el alcance no se asumen: fueron resueltos explícitamente en las aclaraciones Q1–Q5.

- **Identificación de la cuenta**: la cuenta se identifica con el correo electrónico de la persona (práctica estándar); no hay "nombres de usuario" aparte. Cada correo pertenece a una sola cuenta y se compara normalizado (decisión Q5). Los datos de la cuenta son nombre, apellidos, correo y número de teléfono, todos obligatorios y con nombre y apellidos como campos separados (confirmado por el humano el 2026-10-04); el formato concreto del teléfono es un default revisable (ver FR-009).
- **Significado de "desactivar"**: bloquear el acceso de inmediato (también las sesiones ya abiertas) y conservar todos los datos de la cuenta. Desactivar no borra nada y no existe eliminación de cuentas en el MVP (decisión Q3).
- **Política de contraseñas (confirmada en Q2 y por el humano el 2026-10-04)**: entre **8 y 64 caracteres** (mínimo 8, máximo 64), combinando **mayúsculas, minúsculas, números y caracteres especiales**, y **distinta del nombre, de los apellidos y del correo** de la persona (igualdad con comparación normalizada, no de contenido). Aplica a la contraseña inicial, a las restablecidas por un administrador y a los cambios que hace la persona. La contraseña nunca se muestra ni se devuelve; su resguardo sigue las reglas de seguridad de la constitución §IV.
- **Entrega de credenciales fuera del sistema**: al no haber servicio de correo en el MVP, cómo se comunica a la persona su contraseña inicial o restablecida queda a criterio del equipo (p. ej., en persona); el sistema solo exige el cambio al entrar (decisión Q2).
- **Sesión**: existe cierre de sesión manual (US1), cierre tras **30 minutos de inactividad** y un límite absoluto de **1 hora** desde el inicio de la sesión (FR-005, US1 escenarios 6–7, confirmado por el humano el 2026-10-04): la inactividad puede terminar la sesión antes, y 1 hora es el techo incluso con actividad continua. El detalle de cómo se materializa la sesión tiene una decisión pendiente heredada de F1 (`specs/001-estructura-base/estado.md`, decisión D-A7); esta spec no la decide.
- **Unidad del permiso**: un permiso corresponde a un módulo completo (lectura literal de la Decisión 5: "permisos por módulo"), no a acciones sueltas dentro del módulo (crear, editar, borrar). Si hiciera falta esa distinción, sería una ampliación.
- **Catálogo de permisos de F2**: los módulos del producto del roadmap más la administración de usuarios y roles (FR-015). Los permisos de módulos que aún no existen quedan reservados y no producen acceso a nada hasta que esos módulos se construyan (F3–F9).
- **"Administrador" no es un rol fijo**: es toda cuenta con el permiso de administrar usuarios y roles; este permiso es el que define quién gestiona cuentas y roles. El administrador inicial nace con todos los permisos.
- **Idioma del panel**: se entrega en español, que es el idioma del equipo que lo usa. El bilingüismo del roadmap (Decisión 6) aplica al sitio público de F3–F9. De quererse el panel en ambos idiomas, sería una ampliación.
- **Límite de intentos fallidos** (FR-006, confirmado por el humano el 2026-10-04): el bloqueo temporal lo activa el **5.º intento fallido** —que recibe el error genérico, sin revelar si la cuenta existe— y dura **15 minutos**, durante los que el acceso queda bloqueado con el mensaje de bloqueo temporal desde el **6.º intento**; pasados esos 15 minutos se puede volver a intentar iniciar sesión. Medida estándar contra la prueba masiva de contraseñas.
- **Sin cuentas compartidas**: cada persona del equipo debe tener su propia cuenta; el uso compartido de credenciales queda fuera del modelo.
- **Alcance y datos de la auditoría** *(cambio aprobado el 2026-10-04)*: el registro cubre los accesos al panel y las acciones administrativas de gestión de cuentas y roles (US8). "Origen" es la IP de origen del intento de acceso. Los registros se conservan mientras el sistema funciona: no hay retención, purga ni exportación en el MVP (defaults revisables, ver Out of Scope).
- **Nivel de detalle de las acciones registradas**: se registra quién, qué acción, sobre qué cuenta o rol, cuándo y con qué resultado; conservar además el detalle concreto de los cambios (qué datos o permisos se modificaron, antes y después) es un default deseable pero revisable, no un requisito del MVP.
- **Permiso de la auditoría**: la sección de registro se gobierna con el permiso de administrar usuarios y roles (decisión documentada en *Decisiones adicionales*); no existe un permiso propio de "auditoría" en el catálogo.
