# Feature Specification: Estructura base del proyecto (F1)

**Feature Branch**: `001-estructura-base`

**Created**: 2026-09-29

**Status**: Approved (2026-09-30) — alcance ampliado con la plataforma interna y la receta para agregar áreas de negocio; aprobada por el humano tras revisar el delta del 2026-09-30

**Input**: User description: "F1 — Estructura base del proyecto (roadmap §3, `docs/producto/roadmap.md`): la base técnica sobre la que se construye todo lo demás. Backend en Go con un endpoint de estado (`/healthz`) que verifica la conexión a PostgreSQL; frontend en React que muestra el estado del backend; todo (backend, frontend, base de datos) se levanta con Docker Compose con un solo comando; y un pipeline de CI (GitHub Actions) que valida cada cambio (pruebas, linters) antes de integrarse. Para el visitante final aún no hay funcionalidad visible: es el esqueleto del proyecto." · Ampliación de alcance (decisión del 2026-09-30): "la base técnica de F1 debe incluir una plataforma interna que estandarice lo transversal (configuración, registro de eventos, manejo uniforme de errores, acceso a datos y salud) para que ninguna área de negocio la reimplemente, y una receta documentada y verificable para agregar un área de negocio nueva como unidad aislada, de modo que F2–F9 nazcan ya con ese esqueleto."

> **Nota de alcance**: esta funcionalidad es infraestructura. Sus "usuarios" son el **equipo del proyecto** (quienes desarrollan) y **quien opera el sistema**. Para el visitante final aún no hay funcionalidad de negocio visible; la única página existente muestra el estado del sistema. El stack oficial (Go, React + TypeScript, PostgreSQL, Docker Compose, GitHub Actions) es una **restricción del proyecto** (AGENTS.md y constitución), no una decisión de diseño de esta spec. Con la ampliación del 2026-09-30, F1 entrega además la base técnica común (plataforma interna) y la receta para agregar áreas de negocio: siguen siendo infraestructura interna y no agregan funcionalidad visible para el visitante.

## Cambios respecto de la versión aprobada (2026-09-30)

- **Motivo**: el humano amplió el alcance de F1 (decisión del 2026-09-30). Además de entorno local, estado del sistema, validación automática de cada cambio y configuración por entorno, F1 debe dejar una **base técnica común (plataforma interna)** para lo transversal y una **receta para agregar áreas de negocio**, de modo que F2–F9 nazcan sobre ese esqueleto y ninguna área reimplemente lo transversal.
- **Agregados — historias de usuario** (con sus escenarios de aceptación y prueba independiente): **US4** base técnica común reutilizable por todas las áreas (P1); **US5** receta documentada y verificable para agregar un área de negocio nueva como unidad aislada (P2); **US6** respuestas de la API con formato uniforme de éxito y de error, sin exponer información interna (P2); **US7** proyecto sobre versiones con soporte de seguridad vigente (P3).
- **Agregados — requisitos funcionales**: **FR-010** y **FR-011** (base común provista de forma compartida y prohibición de reimplementarla por área); **FR-012** y **FR-013** (formato uniforme de éxito y de error; los errores inesperados no exponen información interna y su detalle queda registrado para diagnóstico); **FR-014** y **FR-015** (receta documentada, seguible sin decisiones de arquitectura y con pasos de verificación; el área nueva es una unidad aislada que no modifica las existentes); **FR-016** (versiones con soporte de seguridad vigente).
- **Agregados — criterios de éxito**: **SC-006** (las capacidades transversales existen una sola vez, en la base común); **SC-007** (la receta se aplica al menos una vez de principio a fin, sin modificar áreas existentes y con las validaciones en verde); **SC-008** (formato uniforme en el 100% de las respuestas de la API); **SC-009** (cero información interna en la respuesta ante errores inesperados, con el detalle registrado internamente); **SC-010** (100% de tecnologías y versiones en uso con soporte de seguridad vigente).
- **Agregados — casos límites**: necesidad de un área que la base común no ofrece; riesgo de que un área nueva rompa las existentes; fallo del registro de eventos; pérdida del soporte de seguridad de una versión en uso.
- **Ajustados**: la *Nota de alcance* menciona ahora que F1 entrega también la base común y la receta; el campo *Input* incorpora la descripción de la ampliación; *Out of Scope* aclara que F1 no construye áreas de negocio (solo la receta y su verificación con un ejercicio de práctica); *Assumptions* documenta los nuevos defaults (qué significa "unidad aislada", cómo se resuelve una necesidad transversal y cómo se verifica la receta).
- **Sin cambios**: US1–US3, FR-001–FR-009, SC-001–SC-005 y los casos límites previos se mantienen tal cual. Tampoco se decide ninguna tecnología ni detalle de implementación: eso sigue siendo materia de la fase `plan`, que se rehará tras la re-aprobación.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Levantar el entorno completo con un solo comando (Priority: P1)

Como miembro del equipo del proyecto, quiero levantar todo el entorno de desarrollo local (backend, frontend y base de datos) con un único comando, para empezar a trabajar sin instalar ni configurar servicios a mano.

**Why this priority**: Sin un entorno local reproducible no se puede desarrollar nada de lo demás del roadmap. Es la capacidad fundacional: reduce la incorporación de personas nuevas a minutos y garantiza que todo el equipo trabaja sobre el mismo entorno.

**Independent Test**: Se prueba de forma independiente clonando el repositorio en una máquina limpia (solo con Docker instalado), siguiendo la documentación y ejecutando el comando único de arranque: los tres servicios (backend, frontend, base de datos) deben quedar en ejecución y accesibles. Entrega valor por sí sola: cualquier persona del equipo puede trabajar localmente.

**Acceptance Scenarios**:

1. **Dado** un clon limpio del repositorio en una máquina con Docker, **cuando** ejecuto el comando único de arranque documentado, **entonces** el backend, el frontend y la base de datos quedan en ejecución y accesibles, sin pasos manuales adicionales.
2. **Dado** el entorno en ejecución, **cuando** lo detengo por completo y vuelvo a ejecutar el comando de arranque, **entonces** el entorno vuelve a levantarse igual que la primera vez (arranque repetible, sin estado residual que lo impida).
3. **Dado** una persona nueva en el equipo, **cuando** sigue únicamente la documentación del repositorio, **entonces** logra levantar el entorno completo sin ayuda de otra persona.

---

### User Story 2 - Consultar el estado del sistema (Priority: P2)

Como miembro del equipo o quien opera el sistema, quiero consultar en cualquier momento el estado del backend y de su conexión a la base de datos, para saber si el sistema está sano y detectar problemas rápidamente, distinguiendo "conectado" de "no conectado".

**Why this priority**: Es la única capacidad visible que entrega F1 y la que permite verificar que todo lo demás funciona. Además, la operación del sistema exige un punto de consulta de salud (constitución §VII). Sin ella, un fallo de conexión a la base de datos sería indistinguible de un fallo del backend.

**Independent Test**: Se prueba de forma independiente con el entorno levantado: se consulta el endpoint de estado con la base de datos activa y luego con la base de datos detenida, comprobando que la respuesta distingue ambos casos; y se abre la página inicial del frontend para ver el estado reflejado. Entrega valor por sí sola: el sistema es observable desde el primer día.

**Acceptance Scenarios**:

1. **Dado** el backend en ejecución y la base de datos accesible, **cuando** consulto el endpoint de estado, **entonces** la respuesta indica que el servicio está vivo y que la base de datos está conectada.
2. **Dado** el backend en ejecución y la base de datos detenida o inaccesible, **cuando** consulto el endpoint de estado, **entonces** la respuesta indica que la base de datos no está conectada (distinguiéndolo claramente del estado "conectada") y el backend sigue respondiendo.
3. **Dado** el entorno en ejecución, **cuando** abro la página inicial del frontend en el navegador, **entonces** veo el estado actual del backend y de su conexión a la base de datos, sin realizar ninguna acción adicional.
4. **Dado** que la base de datos cambia de "conectada" a "no conectada" (o al revés) con el sistema en ejecución, **cuando** vuelvo a consultar el estado, **entonces** la respuesta refleja el estado actual, no un estado memorizado del arranque.

---

### User Story 3 - Validación automática de cada cambio (Priority: P3)

Como miembro del equipo, quiero que cada cambio propuesto pase automáticamente por pruebas y análisis estático antes de poder integrarse, para que la calidad mínima del proyecto esté garantizada desde el primer cambio.

**Why this priority**: Protege todo lo construido: la constitución del proyecto (§III) exige que ningún cambio se integre con pruebas fallando, así que la validación automática debe existir desde F1. Es P3 porque no bloquea el arranque del entorno ni la consulta de estado, pero es indispensable antes de que el proyecto crezca.

**Independent Test**: Se prueba de forma independiente proponiendo dos cambios: uno que supera las validaciones y otro que las falla (por ejemplo, una prueba rota a propósito). El pipeline debe marcar el primero como apto y el segundo como no apto, con el motivo visible. Entrega valor por sí sola: la calidad queda protegida automáticamente.

**Acceptance Scenarios**:

1. **Dado** un cambio propuesto en el repositorio, **cuando** se envía para integración, **entonces** se ejecutan automáticamente las pruebas automatizadas y los análisis estáticos (linters) del backend y del frontend definidos por los estándares del proyecto.
2. **Dado** un cambio que no supera alguna validación, **cuando** el pipeline termina, **entonces** el cambio queda marcado como no apto para integrarse, con el motivo del fallo visible.
3. **Dado** un cambio que supera todas las validaciones, **cuando** el pipeline termina, **entonces** el cambio queda marcado como apto para integrarse.

---

### User Story 4 - Contar con una base técnica común para todas las áreas de negocio (Priority: P1)

Como miembro del equipo del proyecto, quiero que exista una base técnica común (plataforma interna) —configuración, registro de eventos, manejo uniforme de errores, acceso a datos, tratamiento transversal de cada petición y salud— que todas las áreas de negocio reutilicen, para que ninguna tenga que reimplementar lo transversal y todo el sistema se comporte de forma consistente.

**Why this priority**: Es el corazón de la ampliación de alcance del 2026-09-30 y la razón de que F1 exista para F2–F9. Sin una base común, cada área de negocio inventaría su propia configuración, su forma de registrar eventos, su manejo de errores y su acceso a datos, multiplicando el trabajo y las inconsistencias. Debe existir antes de que nazca cualquier área.

**Independent Test**: Se prueba de forma independiente sin ninguna área de negocio: se comprueba que la base común provee las capacidades listadas y que son realmente compartidas (un mismo cambio de configuración aplica para todo el sistema; un error provocado en cualquier punto se registra y se responde del mismo modo), y se verifica que no existan implementaciones propias de esas capacidades fuera de la base común. Entrega valor por sí sola: el equipo trabaja sobre un cimiento común desde el primer día.

**Acceptance Scenarios**:

1. **Dado** el sistema en ejecución, **cuando** cualquier parte del sistema (el ejercicio de la receta o, más adelante, un área de negocio) necesita configuración, registro de eventos, manejo de errores, acceso a datos, tratamiento transversal de cada petición o salud, **entonces** utiliza las capacidades provistas por la base técnica común, sin implementarlas por su cuenta.
2. **Dado** una capacidad de la base común utilizada desde varios lugares, **cuando** se cambia su configuración, **entonces** el cambio aplica para todos los que la usan, sin tener que modificarlos uno por uno.
3. **Dado** cualquier operación del sistema, **cuando** ocurre algo que debe quedar registrado (inicio, resultado o error), **entonces** el evento se registra con el mismo formato, los mismos niveles de severidad y la misma trazabilidad en todas las partes del sistema.
4. **Dado** un error inesperado en cualquier parte del sistema, **cuando** se produce, **entonces** queda registrado con su detalle interno de diagnóstico de la misma forma, sin importar en qué parte del sistema ocurrió.

---

### User Story 5 - Agregar un área de negocio nueva siguiendo una receta (Priority: P2)

Como miembro del equipo del proyecto, quiero que exista una receta documentada y verificable para agregar un área de negocio nueva como una unidad aislada, para que cada funcionalidad del roadmap (F2–F9) nazca sobre el mismo esqueleto, sin reabrir decisiones de arquitectura y sin tocar las áreas ya construidas.

**Why this priority**: Es la segunda mitad de la ampliación de alcance: la base común evita reimplementar lo transversal y la receta evita que cada área nueva se invente su propia forma de integrarse. Depende de la base común (US4), por eso va después; debe estar lista antes de que arranque F2.

**Independent Test**: Se prueba de forma independiente pidiendo a una persona del equipo que no participó en la creación de la receta que la siga de principio a fin para agregar un área nueva de práctica: debe lograrlo sin consultar decisiones de arquitectura, sin modificar las áreas existentes y con las validaciones automáticas en verde al final. Entrega valor por sí sola: el camino para crecer está escrito y probado.

**Acceptance Scenarios**:

1. **Dado** la receta documentada, **cuando** una persona del equipo la sigue para agregar un área de negocio nueva, **entonces** completa todo el proceso sin necesidad de tomar decisiones de arquitectura ni consultar a otra persona: cada paso está indicado.
2. **Dado** un sistema con áreas de negocio existentes, **cuando** se agrega un área nueva siguiendo la receta, **entonces** las áreas existentes no se modifican ni cambian de comportamiento: todo lo agregado pertenece a la nueva unidad.
3. **Dado** un área nueva agregada siguiendo la receta, **cuando** se ejecutan las validaciones automáticas del proyecto, **entonces** pasan sin necesidad de ajustar las áreas existentes.
4. **Dado** la receta documentada, **cuando** la sigo hasta el final, **entonces** incluye los pasos para comprobar por mí mismo que el área quedó correctamente integrada y que las demás áreas siguen intactas.

---

### User Story 6 - Respuestas de la API con formato uniforme y sin información interna (Priority: P2)

Como miembro del equipo que consume la API (por ejemplo, la interfaz del sitio), quiero que toda respuesta de la API —de éxito y de error— siga un único formato documentado y que un error inesperado nunca revele información interna del sistema, para manejar cualquier respuesta con un solo mecanismo y no exponer el interior del sistema a quien no debe verlo.

**Why this priority**: Es la cara visible del manejo uniforme de errores de la base común. Sin un formato único, cada área respondería a su manera y quien consume la API tendría que tratar cada caso aparte; y un error que delata el interior del sistema es una puerta abierta a quien busca atacarlo. Va junto con la base común (US4) y debe estar resuelto antes de que el primer área de negocio exponga operaciones.

**Independent Test**: Se prueba de forma independiente consultando operaciones que terminan bien, que fallan por datos inválidos y que fallan por un error inesperado provocado a propósito: las tres respuestas deben compartir el mismo formato; la de error inesperado no debe contener nombres internos, trazas ni datos de infraestructura; y el detalle del error debe quedar registrado internamente para diagnóstico. Entrega valor por sí sola: la API ya es predecible y segura para quien la consume.

**Acceptance Scenarios**:

1. **Dado** una operación de la API que termina bien, **cuando** recibo la respuesta, **entonces** sigue el formato de éxito uniforme documentado para toda la API.
2. **Dado** una operación que falla por un problema previsto (por ejemplo, datos inválidos), **cuando** recibo la respuesta, **entonces** sigue el formato de error uniforme e indica de forma comprensible qué se esperaba.
3. **Dado** una operación que falla por un problema inesperado, **cuando** recibo la respuesta, **entonces** sigue el formato de error uniforme con un mensaje genérico que no revela información interna del sistema (nombres internos, trazas de error, datos de infraestructura ni de la base de datos).
4. **Dado** dos operaciones distintas de la API, **cuando** comparo sus respuestas (de éxito o de error), **entonces** ambas usan el mismo formato: quien consume la API necesita implementar el manejo de respuestas una sola vez.
5. **Dado** un error inesperado respondido con mensaje genérico, **cuando** lo diagnostico desde el sistema, **entonces** el detalle interno del error está registrado para poder investigarlo.

---

### User Story 7 - Mantener el proyecto sobre versiones con soporte de seguridad vigente (Priority: P3)

Como quien opera el sistema, quiero que el proyecto se mantenga sobre versiones de sus tecnologías y dependencias dentro del periodo de soporte de seguridad vigente, para que la base sobre la que se construye todo el roadmap no acumule riesgos conocidos y ya sin corrección posible.

**Why this priority**: F1 es el esqueleto que copiarán F2–F9: si arranca sobre versiones sin soporte de seguridad, todo el proyecto hereda el riesgo. No bloquea el arranque del entorno ni la consulta de estado, pero debe quedar garantizado desde el primer día y mantenerse en el tiempo.

**Independent Test**: Se prueba de forma independiente listando las tecnologías y versiones en uso del proyecto y contrastándolas con el periodo de soporte de seguridad declarado por quien las mantiene: todas deben estar vigentes. Entrega valor por sí sola: el proyecto arranca sobre una base sostenible.

**Acceptance Scenarios**:

1. **Dado** el proyecto, **cuando** se define o actualiza su base técnica, **entonces** las tecnologías y versiones en uso están dentro de su periodo de soporte de seguridad vigente.
2. **Dado** una tecnología o versión en uso que deja de tener soporte de seguridad vigente, **cuando** se detecta, **entonces** queda identificada como pendiente de actualización.

---

### Edge Cases

- **La base de datos tarda más en arrancar que el backend** (típico en el primer arranque): el endpoint de estado responde "no conectada" hasta que la conexión se establece; el backend no falla ni se cae por ello.
- **La base de datos se cae con el sistema ya en ejecución**: el endpoint de estado pasa a indicar "no conectada" y el backend sigue respondiendo; cuando la base de datos se recupera, el estado vuelve a "conectada" sin reiniciar ningún servicio.
- **Un puerto necesario ya está ocupado** en la máquina de desarrollo: el arranque falla con un error identificable y la documentación indica qué puertos utiliza el proyecto.
- **Falta una variable de entorno**: el entorno local funciona con los valores de ejemplo documentados; no se requiere ningún secreto real para desarrollo local.
- **El pipeline falla por un problema de la infraestructura de CI** (no del cambio): la ejecución se puede reintentar sin modificar el cambio.
- **Un área de negocio necesita algo que la base común no ofrece**: si la necesidad es propia del área, se resuelve dentro de su unidad aislada; si resulta ser transversal (del mismo tipo que las capacidades de la base común), se incorpora a la base común como capacidad compartida antes de usarla. En ningún caso el área reimplementa de forma aislada algo transversal.
- **Un área nueva podría romper las áreas existentes**: lo impiden el aislamiento (unidad aparte, sin editar las existentes) y las validaciones automáticas; si el agregado del área hace fallar algo de las existentes, el cambio queda marcado como no apto para integrarse hasta corregirlo.
- **El registro de eventos falla o no está disponible**: el sistema no deja de funcionar ni de responder por ello; el registro es una ayuda para el diagnóstico, no parte del resultado de cada operación.
- **Una tecnología o versión en uso pierde su soporte de seguridad vigente** durante el desarrollo: queda identificada como pendiente de actualización (US7) para que el riesgo no se herede al resto del roadmap.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema DEBE poder levantarse por completo en local (backend, frontend y base de datos) con un único comando documentado, teniendo Docker como único prerequisito instalado en la máquina.
- **FR-002**: El backend DEBE exponer un endpoint de estado (`/healthz`, nombre exigido por la constitución §VII) que informe si el servicio está vivo y si la conexión con la base de datos está establecida, distinguiendo "conectada" de "no conectada".
- **FR-003**: El endpoint de estado DEBE reflejar el estado real de la conexión a la base de datos en el momento de cada consulta, no un estado memorizado del arranque.
- **FR-004**: El backend DEBE seguir respondiendo en el endpoint de estado aunque la base de datos no esté disponible.
- **FR-005**: El frontend DEBE mostrar una página inicial que consulte el estado del backend y lo presente de forma comprensible para cualquier persona (al menos: conectado / no conectado).
- **FR-006**: Todo cambio propuesto DEBE disparar automáticamente un pipeline de integración continua que ejecute las pruebas automatizadas y los análisis estáticos (linters) de backend y frontend exigidos por los estándares del proyecto (constitución §III y §V).
- **FR-007**: El pipeline DEBE marcar cada cambio como apto o no apto para integrarse según el resultado de las validaciones, mostrando el motivo del fallo cuando lo haya.
- **FR-008**: La configuración DEBE provenir de variables de entorno; el repositorio DEBE incluir valores de ejemplo documentados para el entorno local y NUNCA debe contener secretos reales.
- **FR-009**: El repositorio DEBE incluir documentación que permita a una persona nueva levantar el entorno local desde cero sin ayuda de otro miembro del equipo.
- **FR-010**: El proyecto DEBE proveer una base técnica común (plataforma interna), compartida por todas las áreas de negocio, que estandarice al menos: configuración, registro de eventos, manejo uniforme de errores, acceso a datos, tratamiento transversal de cada petición y salud.
- **FR-011**: Ninguna área de negocio DEBE reimplementar las capacidades de la base técnica común: todas DEBEN utilizar las que ella provee. Toda necesidad resultante transversal DEBE incorporarse a la base común como capacidad compartida, nunca dentro de un área.
- **FR-012**: Toda respuesta de la API DEBE seguir un único formato documentado, tanto para éxitos como para errores, de modo que quien la consume pueda manejar cualquier respuesta con un solo mecanismo.
- **FR-013**: Los errores inesperados NUNCA DEBEN exponer información interna del sistema en la respuesta (nombres internos, trazas de error, datos de infraestructura o de la base de datos): la respuesta DEBE contener un mensaje genérico y el detalle interno DEBE quedar registrado para diagnóstico.
- **FR-014**: El proyecto DEBE incluir una receta documentada para agregar un área de negocio nueva que pueda seguirse de principio a fin sin necesidad de tomar decisiones de arquitectura, e incluya los pasos de verificación del resultado.
- **FR-015**: El agregado de un área de negocio nueva DEBE ser una unidad aislada: su incorporación NUNCA DEBE modificar las áreas de negocio existentes ni su comportamiento.
- **FR-016**: El proyecto DEBE mantenerse sobre versiones de sus tecnologías y dependencias dentro del periodo de soporte de seguridad vigente; lo que deje de estarlo DEBE quedar identificado como pendiente de actualización.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Una persona nueva en el equipo levanta el entorno local completo en menos de 15 minutos, siguiendo únicamente la documentación del repositorio y ejecutando un único comando de arranque.
- **SC-002**: El estado del sistema distingue correctamente los escenarios "base de datos conectada" y "base de datos no conectada" en el 100% de las verificaciones de aceptación (probado en ambos escenarios, incluida la recuperación tras una caída).
- **SC-003**: El 100% de los cambios propuestos recibe un veredicto automático (apto / no apto) antes de integrarse; ningún cambio con validaciones fallidas queda integrado.
- **SC-004**: El veredicto automático de un cambio está disponible en menos de 10 minutos desde que se propone.
- **SC-005**: Cualquier persona que abra la página inicial ve el estado actual del sistema sin acciones adicionales, en menos de 3 segundos desde que la página termina de cargar.
- **SC-006**: Las capacidades transversales (configuración, registro de eventos, manejo de errores, acceso a datos, tratamiento de peticiones y salud) están implementadas una sola vez, en la base común: 0 reimplantaciones propias en otras partes del sistema (verificado en F1).
- **SC-007**: Al menos una vez antes del cierre de F1, una persona del equipo sigue la receta de principio a fin y agrega un área nueva de práctica sin tomar decisiones de arquitectura y sin modificar las áreas existentes, con las validaciones automáticas en verde al terminar.
- **SC-008**: El 100% de las respuestas de la API observadas en las pruebas de aceptación (de éxito y de error) sigue el formato uniforme documentado.
- **SC-009**: En el 100% de los errores inesperados provocados durante las pruebas de aceptación, la respuesta no contiene información interna del sistema y el detalle del error queda registrado internamente para diagnóstico.
- **SC-010**: En cada verificación de F1, el 100% de las tecnologías y versiones en uso está dentro de su periodo de soporte de seguridad vigente.

## Out of Scope

Queda explícitamente **fuera del alcance** de F1:

- **Toda funcionalidad de negocio visible para el visitante final**: portada e información general (F3), eventos y actividades (F4), grupos de conexión (F5), ministerios (F6), donaciones (F7), noticias y galería (F8) y medios (F9) del roadmap.
- **Acceso y gestión de usuarios, roles y permisos** (F2): el endpoint de estado y la página inicial de F1 son públicos y no requieren autenticación.
- **Despliegue a producción** ni a ningún entorno distinto de la máquina local de desarrollo y el pipeline de integración continua.
- **Diseño visual de la página inicial** (identidad de marca, estilos, animaciones): la página de F1 es meramente funcional (muestra el estado); el diseño llega con F3.
- **Bilingüismo de la interfaz** (decisión 6 del roadmap): se introduce con el sitio público en F3.
- **Esquema de base de datos de negocio**: en F1 la base de datos solo necesita existir y aceptar conexiones; no hay tablas de negocio.
- **Las áreas de negocio del roadmap (F2–F9)**: F1 no construye ninguna. Entrega la base técnica común y la receta para que esas áreas nazcan sobre ellas; la receta se verifica con un ejercicio de práctica, no con una funcionalidad de negocio real.

## Assumptions

- El stack oficial (Go, React + TypeScript, PostgreSQL, Docker Compose, GitHub Actions) es una **restricción del proyecto** definida en AGENTS.md y la constitución, no una decisión de diseño de esta spec. Los detalles internos de implementación (estructura de carpetas, librerías, herramientas de prueba, gestión de migraciones) los define el plan técnico en la fase `plan`.
- El endpoint de estado y la página inicial son **públicos** (sin autenticación): no exponen información sensible y el sistema de usuarios llega en F2.
- Las validaciones del pipeline son las exigidas por los estándares del proyecto (constitución §III y §V): pruebas automatizadas y análisis estático de backend y frontend. Las herramientas concretas las elige el plan técnico.
- El entorno local **no requiere secretos reales**: se usan valores de ejemplo (documentados en `.env.example`) válidos solo para desarrollo.
- "Un único comando" se refiere al arranque del entorno completo con el repositorio ya clonado; la instalación de Docker es prerequisito de la máquina, no parte de esta funcionalidad.
- F1 **no incluye despliegue a producción**; cuando el proyecto lo necesite se definirá como funcionalidad posterior.
- La base de datos de F1 no contiene tablas de negocio; basta con que acepte conexiones para verificar el estado.
- La base técnica común estandariza las capacidades transversales listadas (configuración, registro de eventos, manejo uniforme de errores, acceso a datos, tratamiento transversal de cada petición y salud); su diseño interno y el formato concreto de las respuestas de la API los define el plan técnico.
- "Unidad aislada" significa que agregar un área de negocio nueva solo agrega lo propio de esa área y su conexión a la base común: no se editan las áreas ya construidas. La base común sí puede admitir áreas nuevas.
- La verificación de la receta se hace con un **ejercicio de práctica**; F1 no deja detrás áreas de negocio reales ni funcionalidad visible para el visitante.
- Cuando una necesidad detectada al construir un área resulta transversal, se incorpora a la base común como capacidad compartida; lo exclusivo del área vive dentro de su unidad. Lo transversal se identifica por su naturaleza (configuración, registro, errores, acceso a datos, tratamiento de peticiones, salud), no por la frecuencia de uso.
