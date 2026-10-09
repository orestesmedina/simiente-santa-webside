# Entrega F2 — Acceso y gestión de usuarios

**Fecha:** 2026-10-05 · **Versión:** 0.2.0 · **Rama:** `002-acceso-gestion-usuarios`

Nota de entrega en lenguaje sencillo. Si querés la parte técnica, está en el [`README`](../../README.md), en el [`CHANGELOG`](../../CHANGELOG.md) y en la [spec de la funcionalidad](../../specs/002-acceso-gestion-usuarios/spec.md).

## Qué se entrega

F2 abre el **panel de administración**: la puerta por la que el equipo de la iglesia va a trabajar en todo lo que viene (eventos, grupos, ministerios, donaciones…). Quien entra, ya lo hace con su propia cuenta y solo puede tocar lo que su rol le permite.

- **Inicio de sesión** con correo y contraseña en <http://localhost:5173/login>. Sin sesión iniciada no se llega a ninguna sección del panel.
- **Cuentas individuales para el equipo**: un administrador crea, edita, activa y desactiva cuentas, y les restablece la contraseña si alguien pierde el acceso. **Las cuentas no se borran**: se desactivan y todo lo que tienen se conserva.
- **Roles con permisos por módulo**: cada administrador crea los roles que necesite (p. ej. «Contenido» con eventos y actividades, sin ministerios) y les asigna los permisos que correspondan. Cada persona solo ve y puede usar las secciones de su rol.
- **Auditoría (solo lectura)**: una sección del panel con el **último acceso** de cada cuenta, el **historial de accesos** (quién entró y quién no pudo, con fecha, hora y dirección desde la que se intentó) y el **historial de acciones administrativas** (quién creó, editó o desactivó qué, y cuándo). Se filtra por cuenta y por rango de fechas; desde ahí **no se puede editar ni borrar** ningún registro.
- **Seguridad del acceso**: los mensajes de error nunca revelan si una cuenta existe; el quinto intento fallido bloquea el acceso 15 minutos; la sesión se cierra a los **30 minutos de inactividad** o a la **hora máxima** desde que se inició, lo que pase primero; y quien recibe una contraseña de un administrador **debe cambiarla al entrar**.
- **Puesta en marcha garantizada**: una instalación nueva crea su primer administrador con una **acción única que no se puede repetir**, de modo que siempre hay alguien capaz de administrar el panel y ninguna operación puede dejarlo sin administración.

## Cómo probarlo

La guía completa está en [`specs/002-acceso-gestion-usuarios/quickstart.md`](../../specs/002-acceso-gestion-usuarios/quickstart.md), secciones **§1 a §3**:

- **§0 · Preparar el entorno**: `make up` y `make db-migrate`. Con `docker compose ps`, `db` y `redis` deben estar en `healthy`.
- **§1 · Inicializar el administrador**: poner un token en `BOOTSTRAP_TOKEN` dentro de `.env` (cómo, en el [`README` → «Puesta en marcha del panel»](../../README.md)) y llamar a `POST /api/v1/setup/initialize` con la cabecera `X-Setup-Token`. Esperado: `201` con la cuenta creada; repetir el mismo comando → `409` («ya se hizo y no puede repetirse»).
- **§2 · Iniciar sesión**: entrar en <http://localhost:5173/login> con esa cuenta. Con correo o contraseña equivocados aparece **siempre el mismo mensaje genérico**, sin decir si la cuenta existe.
- **§3 · Comprobar la sesión**: la sesión se mantiene mientras se usa, cerrar sesión la termina de verdad y el panel pide volver a entrar cuando pasa el tiempo.

El resto del quickstart (§4 en adelante) recorre la creación de roles y cuentas, los permisos, los cambios de contraseña, el bloqueo por intentos fallidos y la auditoría; el §11 ejecuta todas las pruebas automatizadas.

## Recuperación de contraseña (decisión de esta entrega)

**No hay envío de correos en esta versión.** Si alguien olvida su contraseña o pierde el acceso, **la restablece un administrador** desde el panel: le define una contraseña nueva y la persona debe cambiarla en cuanto entra. Así queda resuelto sin depender de un servicio de correo, que el MVP todavía no tiene. La recuperación por auto-servicio con enlace al correo quedó anotada como idea futura en el [roadmap](../producto/roadmap.md) (§5) y registrada como decisión 9 (§6).

## Límites conocidos de esta entrega

- **Las pruebas end-to-end no corren en este equipo (WSL)**: faltan las librerías de sistema de Chromium (`libnspr4`, `libnss3`, `libasound2`) y no hay permisos para instalarlas. Se ejecutaron con la **imagen Docker oficial de Playwright** (`mcr.microsoft.com/playwright:v1.63.0-jammy`) contra el stack real: **3 de 3 en verde**. En un equipo con esas librerías instaladas, el comando habitual es `make e2e`.
- **Redis no guarda nada entre reinicios** (decisión del proyecto): si se reinicia el servicio, hay que volver a iniciar sesión y los contadores de intentos fallidos empiezan de cero. El registro de auditoría **no** se pierde: vive en la base de datos.
- **Los permisos de los módulos que aún no existen** (eventos, actividades, grupos, ministerios, donaciones, noticias, medios…) ya aparecen en el formulario de roles, pero hoy no abren nada: están reservados para F3–F9.
- **La auditoría cubre solo este alcance**: accesos al panel y gestión de cuentas y roles. No exporta los registros, no manda avisos y no los borra.
- **El panel está en español**; el bilingüismo (español/inglés) es del sitio público y llega con F3.
- **Sin despliegue a internet**: todo funciona en local, en tu equipo.

## Qué sigue

**F3 — Portada e información general**: el sitio público con la identidad de la iglesia (horario, quiénes somos, redes), que el equipo edita desde este panel. Después siguen eventos, grupos, ministerios y donaciones, en el orden del [roadmap](../producto/roadmap.md).

## Enlaces

- [`README.md`](../../README.md) — instalación, comandos y variables de entorno.
- [`CHANGELOG.md`](../../CHANGELOG.md) — qué cambió en la versión 0.2.0.
- [`specs/002-acceso-gestion-usuarios/spec.md`](../../specs/002-acceso-gestion-usuarios/spec.md) — la especificación aprobada.
- [`docs/producto/roadmap.md`](../producto/roadmap.md) — las funcionalidades F1–F9 y su estado.
