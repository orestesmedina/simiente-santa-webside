# Feature Specification: Estructura base del proyecto (F1)

**Feature Branch**: `001-estructura-base`

**Created**: 2026-09-29

**Status**: Draft

**Input**: User description: "F1 — Estructura base del proyecto (roadmap §3, `docs/producto/roadmap.md`): la base técnica sobre la que se construye todo lo demás. Backend en Go con un endpoint de estado (`/healthz`) que verifica la conexión a PostgreSQL; frontend en React que muestra el estado del backend; todo (backend, frontend, base de datos) se levanta con Docker Compose con un solo comando; y un pipeline de CI (GitHub Actions) que valida cada cambio (pruebas, linters) antes de integrarse. Para el visitante final aún no hay funcionalidad visible: es el esqueleto del proyecto."

> **Nota de alcance**: esta funcionalidad es infraestructura. Sus "usuarios" son el **equipo del proyecto** (quienes desarrollan) y **quien opera el sistema**. Para el visitante final aún no hay funcionalidad de negocio visible; la única página existente muestra el estado del sistema. El stack oficial (Go, React + TypeScript, PostgreSQL, Docker Compose, GitHub Actions) es una **restricción del proyecto** (AGENTS.md y constitución), no una decisión de diseño de esta spec.

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

### Edge Cases

- **La base de datos tarda más en arrancar que el backend** (típico en el primer arranque): el endpoint de estado responde "no conectada" hasta que la conexión se establece; el backend no falla ni se cae por ello.
- **La base de datos se cae con el sistema ya en ejecución**: el endpoint de estado pasa a indicar "no conectada" y el backend sigue respondiendo; cuando la base de datos se recupera, el estado vuelve a "conectada" sin reiniciar ningún servicio.
- **Un puerto necesario ya está ocupado** en la máquina de desarrollo: el arranque falla con un error identificable y la documentación indica qué puertos utiliza el proyecto.
- **Falta una variable de entorno**: el entorno local funciona con los valores de ejemplo documentados; no se requiere ningún secreto real para desarrollo local.
- **El pipeline falla por un problema de la infraestructura de CI** (no del cambio): la ejecución se puede reintentar sin modificar el cambio.

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

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Una persona nueva en el equipo levanta el entorno local completo en menos de 15 minutos, siguiendo únicamente la documentación del repositorio y ejecutando un único comando de arranque.
- **SC-002**: El estado del sistema distingue correctamente los escenarios "base de datos conectada" y "base de datos no conectada" en el 100% de las verificaciones de aceptación (probado en ambos escenarios, incluida la recuperación tras una caída).
- **SC-003**: El 100% de los cambios propuestos recibe un veredicto automático (apto / no apto) antes de integrarse; ningún cambio con validaciones fallidas queda integrado.
- **SC-004**: El veredicto automático de un cambio está disponible en menos de 10 minutos desde que se propone.
- **SC-005**: Cualquier persona que abra la página inicial ve el estado actual del sistema sin acciones adicionales, en menos de 3 segundos desde que la página termina de cargar.

## Out of Scope

Queda explícitamente **fuera del alcance** de F1:

- **Toda funcionalidad de negocio visible para el visitante final**: portada e información general (F3), eventos y actividades (F4), grupos de conexión (F5), ministerios (F6), donaciones (F7), noticias y galería (F8) y medios (F9) del roadmap.
- **Acceso y gestión de usuarios, roles y permisos** (F2): el endpoint de estado y la página inicial de F1 son públicos y no requieren autenticación.
- **Despliegue a producción** ni a ningún entorno distinto de la máquina local de desarrollo y el pipeline de integración continua.
- **Diseño visual de la página inicial** (identidad de marca, estilos, animaciones): la página de F1 es meramente funcional (muestra el estado); el diseño llega con F3.
- **Bilingüismo de la interfaz** (decisión 6 del roadmap): se introduce con el sitio público en F3.
- **Esquema de base de datos de negocio**: en F1 la base de datos solo necesita existir y aceptar conexiones; no hay tablas de negocio.

## Assumptions

- El stack oficial (Go, React + TypeScript, PostgreSQL, Docker Compose, GitHub Actions) es una **restricción del proyecto** definida en AGENTS.md y la constitución, no una decisión de diseño de esta spec. Los detalles internos de implementación (estructura de carpetas, librerías, herramientas de prueba, gestión de migraciones) los define el plan técnico en la fase `plan`.
- El endpoint de estado y la página inicial son **públicos** (sin autenticación): no exponen información sensible y el sistema de usuarios llega en F2.
- Las validaciones del pipeline son las exigidas por los estándares del proyecto (constitución §III y §V): pruebas automatizadas y análisis estático de backend y frontend. Las herramientas concretas las elige el plan técnico.
- El entorno local **no requiere secretos reales**: se usan valores de ejemplo (documentados en `.env.example`) válidos solo para desarrollo.
- "Un único comando" se refiere al arranque del entorno completo con el repositorio ya clonado; la instalación de Docker es prerequisito de la máquina, no parte de esta funcionalidad.
- F1 **no incluye despliegue a producción**; cuando el proyecto lo necesite se definirá como funcionalidad posterior.
- La base de datos de F1 no contiene tablas de negocio; basta con que acepte conexiones para verificar el estado.
