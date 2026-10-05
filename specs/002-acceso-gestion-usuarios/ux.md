# UX: F2 — Acceso y gestión de usuarios

> Documento de diseño para `disenador-ux`. Fuente: `spec.md` (aprobada 2026-10-04, con cambio de alcance auditoría US8/FR-021–FR-026 incorporado el 2026-10-04) y skill `react-frontend`.
> No es código: el `dev-frontend` lo implementa. Las decisiones de rutas/contratos concretos las fija el `arquitecto` en `plan.md`; aquí se proponen.

## 0. Principios de diseño

- **Mobile-first**: una sola columna; el listado pasa de tarjetas (móvil) a tabla (tableta en adelante).
- **Usable por personas mayores**: cuerpo a **18 px** mínimo, interlineado ≥ 1.5, botones y campos con objetivo táctil de **mínimo 44×44 px**, nada de iconografía sola (siempre icono + texto), lenguaje llano sin jerga ("desactivar la cuenta" y no "suspender el acceso").
- **Contraste WCAG 2.1 AA**: texto normal ≥ 4.5:1, texto grande ≥ 3:1. Base: `slate-900` sobre `white`, errores en texto `#B42318`-equivalente AA sobre blanco, enlaces subrayados sobre fondo de lectura.
- **Errores que ayudan**: cada mensaje dice qué pasó, qué significa y qué hacer; nunca expone detalles internos ni confirma si una cuenta existe (FR-003).
- **El backend manda**: la UI oculta secciones sin permiso por comodidad, pero toda operación sensible se vuelve a verificar en el servidor (FR-016). La sesión vive en cookie `HttpOnly`; nada de tokens en `localStorage`.
- **Español**: todo el panel en español (el bilingüismo es del sitio público, F3+).

---

## 1. Flujo de usuario

### 1.1 Camino feliz: entrar y trabajar

1. La persona abre `/login` (o cualquier URL del panel sin sesión → redirige a `/login`, FR-001).
2. Escribe **correo** y **contraseña**, pulsa "Entrar" (o `Enter`).
3. Éxito → el sistema consulta "mi sesión" (usuario, rol, permisos, `cambiar contraseña obligatoria`).
   - Si la sesión pide cambio de contraseña obligatorio (contraseña definida/restablecida por un administrador, FR-010):
     → primero `/cambiar-contrasena` en modo "obligatorio": no puede navegar ni cerrar el aviso hasta cambiarla (US3.6, US7.4–5). Después → panel.
   - Si no → entra directo al panel y ve solo los módulos que su rol autoriza (US1.1).
4. Trabaja. "Salir" siempre visible (FR-004).

### 1.2 Caminos de error en el acceso

| Situación | Camino |
|---|---|
| Credenciales incorrectas | Mensaje genérico arriba del formulario, el campo incorrecto no se revela; foco en el aviso; puede reintentar. **No se muestra "intentos restantes"** (revelaría si la cuenta existe). FR-003, SC-008 |
| Cuenta inactiva con credenciales correctas | Mensaje dedicado: el acceso está desactivado y qué hacer (p. ej. "Pide a un administrador que la reactive."). US1.3 |
| 5 fallos → bloqueo | Formulario deshabilitado con aviso y **hora calculada** de reintento ("a partir de las 16:45"); cuenta atrás visible; sin poder enviar desde la UI hasta que expire. FR-006 |
| Sesión expirada | El aviso 2 minutos antes del cierre por inactividad ("¿Sigues ahí?") con botón "Sigo aquí"; si no hay respuesta → `/login` con notificación "Tu sesión terminó por inactividad." La sesión también caduca al cumplir el techo absoluto de 1 hora aunque haya actividad: → `/login` con "Tu sesión terminó." FR-005 |
| Caduca con formulario a medias | Al volver a entrar se le informa: "Tu sesión terminó. La última acción no se aplicó; vuelve a hacerla." Nada se aplica a medias (Edge Case). El borrador **no se recupera automáticamente** en el MVP (mejora futura); se comunica con claridad. |
| Cuenta desactivada con sesión abierta | La siguiente petición responde 401 → se cierra la sesión y se redirige a `/login` con aviso ("Tu cuenta se desactivó. Pide a un administrador que la active."). US5.2 |
| URL sin sesión | Cualquier ruta del panel → `/login?destino=<url>`; al iniciar sesión, si el origen era del panel, vuelve a esa página. |

### 1.3 Flujos de gestión (solo quien tiene el permiso "Administrar usuarios y roles")

- **Crear cuenta**: Inicio → "Usuarios" → botón "Crear usuario" → formulario (nombre, apellidos, correo, teléfono, rol, contraseña inicial) → éxito → vuelve al listado con aviso "usuario creado" y la fila nueva arriba.
- **Editar cuenta / restablecer contraseña**: fila → "Editar" → formulario precargado (nombre, apellidos, correo, teléfono, rol; estado; botón aparte "Definir una contraseña nueva") → éxito → aviso y listado actualizado.
- **Activar / desactivar**: en el listado, acción por fila "Desactivar"/"Activar" con confirmación solo al desactivar. Intentos que violan la regla del último administrador activo resultan en error explicado (FR-008).
- **Crear/editar rol**: Inicio → "Roles" → "Crear rol" o "Editar" → nombre + casillas de permisos → éxito → listado.
- **Eliminar rol**: fila → "Eliminar" → confirmación → si hay cuentas asignadas, se impide con mensaje "primero reasígalas" (no se deshabilita a ciegas; el intento siempre da una explicación). US6.5
- **Cambiar mi contraseña**: menú de cuenta → "Cambiar contraseña" (actual + nueva) o la versión obligatoria tras restablecimiento.
- **Consultar la auditoría** (solo con permiso de administración, US8): Inicio → "Auditoría" → pestaña "Historial de accesos" o "Acciones administrativas" → filtrar por cuenta y por rango de fechas → leer registros paginados. No hay nada que crear ni modificar: es una consulta.

---

## 2. Rutas propuestas (React Router)

```
/                   (existente: la página "Estado del sistema" de F1; sin módulo de gestión)
/login             Pantalla de acceso
/sin-permiso        Acceso denegado
/panel              Layout autenticado (RequireAuth) con barra de navegación
  /panel            → "Inicio" (Mi cuenta: datos, cambiar contraseña, salir)
  /panel/usuarios   RequirePermiso(admin_usuarios_roles). Listado + crear/editar
  /panel/roles      RequirePermiso(admin_usuarios_roles). Listado + crear/editar/eliminar
  /panel/auditoria  RequirePermiso(admin_usuarios_roles). Registro de solo lectura
/cambiar-contrasena Cambiar contraseña (autenticado, sin permiso extra)
```

- `RequireAuth`: si no hay sesión o la API responde 401 → `/login` con motivo ("sin-sesion", "expirada", "cuenta-desactivada") y `destino` para volver.
- `RequirePermiso(permiso)`: si `miSesion` no incluye el permiso → muestra `SinPermiso` (página `/sin-permiso` con retorno al Inicio). No adivina: solo muestran/permiten lo que el rol autoriza.
- Un catch-all sencillo del router lleva a una página "Página no encontrada" con enlace al panel (o a `/login` si no hay sesión).

---

## 3. Pantallas

### 3.1 Pantalla de acceso — `/login`

- **Propósito**: identificar a la persona y abrirla la sesión (US1, P1).
- **Contenido**: título "Entrar al panel", campos **Correo** y **Contraseña** (con ojo de mostrar/ocultar, `autocomplete` correcto), botón "Entrar", bloque de avisos/noticias del sistema (sesión expirada, cuenta desactivada, etc.) y sin más ornamentos. No hay "regístrate" ni "olvidé mi contraseña" (auto-servicio fuera de alcance: FR-010, Out of Scope).
- **Acciones**: Entrar; datos precargados por el navegador (`autocomplete="email"` / `"current-password"`).
- Nota: si el sistema aún no está inicializado (instalación nueva), no hay pantalla dedicada de puesta en marcha: la inicialización es por API, según el plan (quickstart §1). La pantalla de acceso se muestra y funciona con normalidad.

### 3.2 Layout del panel — `/panel/*`

- **Propósito**: marco común autenticado con navegación filtrada por permisos (US4.6, FR-016).
- **Contenido móvil primero**: barra superior con título de sección, botón de menú (hamburguesa) que abre un panel lateral con la navegación; en ≥ tablet, barra lateral fija con el mismo menú. Pie de barra: mi nombre + rol, "Salir". Navegación (solo módulos existentes y autorizados):
  - "Inicio" (todos).
  - "Usuarios" y "Roles" (solo con permiso de administración de usuarios y roles).
  - "Auditoría" (mismo permiso de administración de usuarios y roles; no existe un permiso propio de auditoría, FR-024).
  - Sin permiso adicional: no aparece ninguna otra sección; una cuenta sin permisos ve solo Inicio (Edge Case válido, FR-016).
- Del panel de navegación quedan fuera (no se pintan ni se habilitan) los módulos F3–F9; sus permisos ya se eligen al crear roles, marcados como "disponible más adelante" (§4.d).

### 3.3 Inicio del panel — `/panel` (índice)

- **Propósito**: aterrizaje con lo mínimo útil: cuenta propia y accesos rápidos; aquí aterriza quien no tiene permiso de administración.
- **Contenido**: tarjeta "Mi cuenta" (nombre, correo, rol, estado del sistema de F1 como mini-resumen si procede), tarjeta "Gestión" con los accesos autorizados (Usuarios, Roles, Auditoría) si procede, botón "Cambiar contraseña", botón "Salir".
- Un usuario sin ningún permiso de módulo ve solo "Mi cuenta" y un texto propio: "Tu cuenta está activa. Aún no tienes secciones asignadas; si necesitas acceso, habla con tu administrador."

### 3.4 Usuarios — `/panel/usuarios` (solo con permiso de administración)

- **Propósito**: ver quién tiene acceso hoy y gestionar cuentas (US3, US5; FR-019).
- **Móvil**: tarjetas apiladas; **tableta en adelante**: tabla.
- **Contenido por fila**: nombre y apellidos, correo, rol (texto plano), estado (`Activo`/`Inactivo` en una "píldora" con color Y texto), y acciones: **Editar**, **Desactivar** o **Activar**. Sin acción de eliminar (FR-013): no existe y no aparece ningún botón gris que confunda.
- **Encabezado**: contador ("8 usuarios"), botón principal "Crear usuario". (Sin buscador de texto libre en el MVP; el listado se pagina, ver §8.)
- **Detalle del formulario**: ver `FormularioUsuario` (§4.c).
- **Último acceso (FR-021)**: en la tarjeta móvil y en la tabla se muestra, bajo el correo o como última columna, el **último acceso exitoso** con fecha, hora e IP de origen ("Último acceso: 03/10/2026 a las 17:42, desde 189.2.4.15"). Si la cuenta **aún no ha entrado nunca**: texto "Nunca ha entrado al panel." — nunca una fecha inventada ni un guion ambiguo.

### 3.5 Crear/editar usuario (modal o página, solo con permiso de administración)

- **Propósito**: crear cuenta con contraseña inicial o editar datos de una cuenta (FR-009, FR-011).
- **Campos** (móvil en columna única, todos obligatorios): **Nombre** (texto), **Apellidos** (texto), **Correo** (texto con validación de formato, con aviso de que se compara ignorando mayúsculas y espacios), **Teléfono** (texto con validación de formato; ayuda "Con código del país si procede, p. ej. 612 345 678 o +34 612 345 678"), Rol (selector con los roles existentes; obligatorio), Estado (solo en edición; conmutador con explicación "Desactivar bloquea el acceso de inmediato pero conserva todo."), y:
  - **crear**: "Contraseña inicial" con toggle mostrar/ocultar + **checklist en vivo de la política** (mín. 8, mayúsculas, minúsculas, números, caracteres especiales, distinta del nombre, los apellidos y el correo, máx. 64) que se va marcando verde al cumplirse; botón "Definir una contraseña nueva" (edición, nunca muestra ni devuelve la actual).
  - Al crear una cuenta se envía también el aviso: "La persona deberá cambiar esta contraseña la primera vez que entre." (FR-010).
- **Validaciones visibles**: junto a cada campo; resumen de errores arriba, no solo CSS rojo.

### 3.6 Roles — `/panel/roles` (solo con permiso de administración)

- **Propósito**: construir y mantener los roles con permisos por módulo (US4, US6; FR-014/015/017).
- **Contenido por fila**: nombre, resumen de permisos (chips con nombres de módulo, con sello "pendiente de módulo" para los reservados), **número de cuentas asignadas**, acciones Editar / Eliminar.
- **Eliminar**: botón siempre activo; confirmación destructiva solo si el rol no está en uso; si está en uso, se impide con mensaje que explica la reasignación previa (FR-017). El botón nunca desaparece: el intento siempre da una respuesta comprensible.

### 3.7 Crear/editar rol (modal o página, solo con permiso de administración)

- **Propósito**: crear o editar nombre y permisos de un rol (FR-014, FR-017).
- **Contenido**: Nombre (+ aviso de normalización) y **casillas por módulo** de FR-015: Portada e información general; Eventos; Actividades; Grupos de conexión; Ministerios; Donaciones; Noticias y galería; Medios; Administración de usuarios y roles. Los módulos F3–F9 (que aún no existen) aparecen con el sello "disponible más adelante" y sin efecto de acceso hasta que construyan: es válido y esperado.
- **Reglas visibles**: al menos un permiso (botón "Guardar" deshabilitado + texto explicativo "Un rol necesita al menos un permiso."); sin catálogo previo (Decisión 5). Al editar el rol "Administración de usuarios y roles" aplican las mismas reglas del último administrador (FR-008).

### 3.8 Cambiar contraseña — `/cambiar-contrasena`

- **Propósito**: cambiar mi propia contraseña (US7; FR-020) y el cambio obligatorio tras una contraseña definida/restablecida por un administrador (FR-010).
- **Contenido**: Contraseña actual (no se pide en el modo "obligatorio" del primer cambio), Contraseña nueva y Confirmar nueva, **check-list en vivo de la política** igual que en 3.5 (mín. 8, mayúsculas, minúsculas, números, caracteres especiales, distinta del nombre, los apellidos y el correo, máx. 64), botón "Guardar" (deshabilitado hasta que todo cumpla).
- **Modo obligatorio**: idéntico pero sin navegación (ni menú ni enlaces) y con texto "Por seguridad, cambia esta contraseña antes de continuar."; al terminar → `/panel` con aviso de éxito.

### 3.9 Sin permiso — `/sin-permiso`

- **Propósito**: de forma transparente al "acceso denegado" de FR-016/US4.5–7: al abrir una URL sin permiso, o tras una operación que el servidor rechaza.
- **Contenido**: título "No tienes acceso a esta sección", texto "Si necesitas entrar aquí, pide a un administrador de la iglesia que actualice tu rol.", botón "Volver al inicio del panel" y "Cerrar sesión". Sin detalles internos del rechazo.
- También la navegación evita llegar: los módulos sin permiso no se muestran en el menú (§3.2).

### 3.10 Auditoría — `/panel/auditoria` (solo con permiso de administración; US8, FR-021–FR-026)

- **Propósito**: consulta **de solo lectura** de los registros de la administración: quién entró, quién no pudo y qué cambios se hicieron sobre cuentas y roles (US8). Nada se crea, edita ni borra aquí: **no hay botones de "Editar", "Eliminar" ni acciones por fila** (FR-025) — la vista no los pinta siquiera gris.
- **Estructura**: título "Auditoría" + **dos pestañas** grandes y etiquetadas con texto (nunca solo iconos):
  - **"Historial de accesos"**: intentos de inicio de sesión, exitosos y fallidos.
    - Móvil: tarjetas; tableta en adelante: tabla. Por fila: **fecha y hora** ("03/10/2026, 17:42"), **cuenta** (el **nombre de la cuenta**, `userName` del contrato; para un intento sin cuenta identificada: **"Intento sin cuenta asociada"** — no se muestra el correo del intento ni se afirma que se guardó uno), **resultado** en píldora con texto ("Exitoso" / "Fallido", más que color) e **IP de origen**. Sin contraseñas ni credenciales (FR-026).
  - **"Acciones administrativas"**: lo que hicieron los administradores.
    - Por fila: **quién** (el **nombre** de quien hizo la acción, `actorName` del contrato), **qué hizo** ("Creó una cuenta", "Restableció la contraseña", "Desactivó una cuenta", "Editó un rol", "Eliminó un rol"…), **sobre qué** (la cuenta o el rol afectado, con su **nombre** si aplica), **cuándo** (fecha y hora) y **resultado** ("Completada" / "No completada" — el intento que falló o fue denegado también aparece, FR-023). Sin credenciales (FR-026).
- **Filtros** (encima del listado, en línea en tablet, apilados en móvil):
  - **Cuenta**: selector con las cuentas existentes ("Todas" por defecto); en "Acciones administrativas", el filtro se aplica a "quién la hizo" y a "sobre qué cuenta" según la pestaña (lo concreto lo fija el contrato del `arquitecto`).
  - **Rango de fechas**: campo "Desde" y campo "Hasta" con selector de date; por defecto vacíos ("todo el histórico").
  - Botón "Filtrar" (aplica) y enlace/botón "Quitar filtros".
- **Paginación** (obligatoria, FR-024): al pie, "Anterior" / "Siguiente" con texto "Mostrando {a}–{b} de {total}"; página actual no repetible como hipervínculo. La paginación se reinicia a la página 1 al aplicar filtros.
- **Nota**: igual que Usuarios y Roles (§8), aquí también se pagina: es requisito de la spec (FR-024), no una mejora futura.

### 3.11 Aviso de sesión por caducar (superposición global en el layout del panel)

- **Propósito**: advertir el cierre inminente por inactividad y permitir continuar con "Sigo aquí" (renueva la actividad); si no se pulsa, expira (FR-005). Dos límites fijados por la spec: **30 minutos de inactividad** con aviso 2 minutos antes, y un **techo absoluto de 1 hora** de sesión aunque haya actividad; al llegar el techo, la sesión se cierra sin aviso previo posible (el usuario ve "Tu sesión terminó." al hacer la siguiente acción).
- Doble capa: **banda superior** fina amarilla 2 minutos antes del cierre por inactividad + **modal** de confirmación con botones "Sigo aquí" (primario) / "Salir". En el aviso no se anuncia "cada segundo" (algo ruidoso para lectores de pantalla): se anuncia al abrirse y con la cuenta de tiempo redondeada ("quedan menos de 2 minutos").

---

## 4. Estados por vista con datos

### a. Pantalla de acceso

| Estado | Qué ve la persona |
|---|---|
| Vacío/inicial | Formulario vacío con etiquetas visibles (nunca solo "placeholder"). |
| Enviando | Botón "Comprobando…" deshabilitado, espiral. Interrupciones: sin doble envío. |
| Éxito | Deja de verse (navega a `/panel` o al `destino`). |
| Error de credenciales | Bloque rojo con `role="alert"`: "Correo o contraseña incorrectos." (genérico, FR-003). |
| Cuenta inactiva | Bloque amarillo-rojo: "Ese acceso está desactivado. Pide a un administrador de la iglesia que lo reactive." |
| Cuenta bloqueada (5 fallos) | Formulario oculto/deshabilitado + temporizador: "Por seguridad, has superado los intentos permitidos. Podrás intentarlo de nuevo a partir de las {HH:MM} (quedan {M} minutos)." |
| Error del sistema | "No se pudo conectar con el sistema. Vuelve a intentarlo en unos minutos." (sin detalles técnicos internos). |

### b. Listado de usuarios

| Estado | Qué ve |
|---|---|
| Cargando | Esqueletos/indicador "Cargando usuarios…" en la zona de resultados (cabecera y botón quedan interactivos). |
| Vacío | "Aún no hay cuentas. Crea la primera para tu equipo." + botón "Crear usuario". |
| Error | "No se pudo cargar el listado." + botón "Reintentar". |
| Éxito | Tabla/tarjetas con datos y accesos por fila. |
| Sin permiso | Ruta protegida: no se llega; si algo sale raro → `SinPermiso` (§3.9). |
| Actualizando (acción de fila) | La fila muestra acción en curso; el listado queda accesible. |
| Confirmación de éxito | Aviso verde superior: "Cuenta creada." / "Cuenta activada." / "Cuenta desactivada." |

### c. Formulario de usuario (crear/editar)

- Vacío (crear) o precargado (editar). Cada validación vive junto a su campo, aparece al enviar (no en vivo para no asustar) salvo la **política de contraseña**, que se marca en vivo.
- Enviando: "Guardando…" deshabilitado, cierre del modal bloqueado.
- Éxito: cierre + aviso en el listado.
- Errores: campo rojo con texto (obligatorios: "Escribe el nombre.", "Escribe los apellidos.", "Escribe un correo con este formato: nombre@dominio.com", "Escribe un número de teléfono válido."; "Elige un rol de la lista.") y errores de servidor mapeados: correo duplicado ("Ya existe una cuenta con ese correo; prueba con otro."; funciona también si solo cambia mayúsculas/espacios, Q5), rol inexistente ("El rol elegido ya no existe; elige otro."), regla del último administrador ("No puedes dejar el panel sin un administrador activo. Activa otra cuenta con permiso de administración primero.").

### d. Listado y formulario de rol

- Igual patrón que b/c con sus textos propios: cargando, vacío ("Aún no hay roles. Crea el primero para organizar los accesos de tu equipo."), error con reintento, éxito.
- Errores propios: nombre duplicado ("Ya existe un rol con ese nombre; elige otro." Q5), sin permisos ("Un rol necesita al menos un permiso."), en uso al eliminar ("Este rol lo están usando {N} cuentas. Primero cámbiales el rol y después podrás eliminarlo.").
- El sello "disponible más adelante" en los permisos de módulos F3–F9 es **información**, no error.

### e. Cambiar contraseña

- Éxito: "Contraseña cambiada. La próxima vez que entres, usa la nueva." (modos obligatorio y normal) y vuelve al panel.
- Errores: la actual no es correcta ("Tu contraseña actual no coincide. Vuelve a escribirla."), la nueva no cumple un requisito (el checklist lo señala en rojo, con texto que nombra el requisito, no un texto genérico).

### g. Auditoría (§3.10)

| Estado | Qué ve |
|---|---|
| Cargando | "Cargando registros…" en la zona de resultados; los filtros siguen interactivos. Al cambiar de pestaña o página, la zona se recarga sin perder los filtros. |
| Vacío | "Todavía no hay registros." Con filtros aplicados, se añade: "Prueba a quitar los filtros o usar un rango de fechas más amplio." Sin filtros (sistema recién instalado): "Todavía no hay registros de actividad. Aparecerán cuando alguien entre al panel o se gestione una cuenta o un rol." |
| Error | "No se pudo cargar el registro." + botón "Reintentar". |
| Éxito | Tabla/tarjetas paginadas con resumen "Mostrando {a}–{b} de {total} registros" (`role="status"` para anunciar el total al aplicar filtros). |
| Sin permiso | Ruta protegida con el mismo permiso que Usuarios y Roles (§3.9). |
| Pestañas y filtros | Cada pestaña mantiene su propia búsqueda y paginación al alternar; aplicar un filtro vuelve a la página 1. |

- La vista no tiene estados de acción en curso ni confirmaciones: no hay acciones (FR-025).

### f. Aviso de sesión

- Cargando sesión: al entrar al panel se pide una vez "mi sesión". Error → banda "No se pudo verificar tu acceso" con "Reintentar" y, si vuelve a fallar con 401, salida limpia.

---

## 5. Componentes

**Existentes que se reutilizan tal cual**: `AppLayout` (cabecera simple para `/` y `/login`), `StatusItem`/`StatusResult` dentro del Inicio (con adaptación menor si el `arquitecto` lo ve). No existe aún biblioteca genérica: F2 crea la base en `frontend/src/components/`.

### Componentes genéricos nuevos (`frontend/src/components/`)

> Nombres de código definitivos (en inglés). El inventario definitivo vive en el plan: `tasks.md` (T242, componentes compartidos); `ux.md` solo da el mapeo a sus etiquetas en español.

| Componente (código) | Props clave | Uso |
|---|---|---|
| `Button` (antes "Boton") | `variant` ('primary' \| 'secondary' \| 'danger' \| 'link'), `loading`, `size` (mín. 44 px) | Todos los botones. |
| `Field` (antes "Campo") | `label`, `type`, `error?`, `help?`, `required`, `autocomplete?` | Envoltorio accesible label + `aria-describedby` con error/ayuda. |
| `PasswordField` (antes "CampoContrasena") | igual que `Field` + `showToggle`, `policy?: string[]` (checklist en vivo, incluye "distinta del nombre, los apellidos y el correo" y "máximo 64 caracteres"), `autocomplete` (p. ej. `new-password`) | Acceso, formularios de contraseña. |
| `Select` | `label`, `options`, `error?` | Rol en formulario de usuario. |
| `Notice` | `variant` ('success' \| 'error' \| 'info' \| 'warning'), `children`, `temporal?` | Mensajes de sistema; `role="alert"` en error/warning, `role="status"` en éxito/info. |
| `ConfirmDialog` (antes "Confirmacion") | `title`, `description`, `confirmText`, `danger?`, `onConfirm`, `onClose` | Desactivar cuenta, eliminar rol. Trampa de foco + `Esc` + devolución del foco. |
| `EmptyState` (junto a estados de carga y error) | `message`, `action?` | Estados vacío, cargando y error. |
| `StatusPill` | `value` ('active' \| 'inactive') | Color + texto (más que color). |
| `Table` | cabeceras + filas; en móvil se compone como lista de tarjetas | Usuarios, roles y auditoría (móvil primero). |
| `Tabs` (antes "Pestannas") | `tabs[{id, label}]`, `active`, `onChange` | Las dos vistas de la Auditoría; texto visible + `aria-current`. |
| `Pagination` (antes "Paginacion") | `page`, `totalPages`, `onChange`, `summaryText` | Pie de la Auditoría y listados: "Anterior" / "Siguiente" deshabilitados en los extremos. |
| `DateRangeFilter` | envuelve dos `Field` tipo date + botones "Filtrar" / "Quitar filtros" | Rango de fechas de la Auditoría. |
| `Dialog` | `title`, `description?`, `onClose` | Formularios en modal (Crear/editar usuario/rol). |

### De funcionalidad (`frontend/src/features/auth/`, `features/usuarios/`, `features/roles/`, `features/auditoria/`)

| Componente / hook (código) | Responsabilidad |
|---|---|
| `features/auth`: `LoginPage`, `useSession`, `useLogin`, `useLogout` | Login y sus errores (§3.1, §4.a). |
| `PanelLayout` | Layout del panel con navegación y `SessionWarning`. |
| `features/auth`: `RequireAuth` / `RequirePermiso` / `RequirePasswordChange` (en `app/guards.tsx`) | Sesión, redirección y permisos de ruta. Fuente: endpoint "mi sesión" del backend. |
| `features/auth`: `SessionWarning` | Caducidad con "Sigo aquí". Temporizador con pruebas (se congelan tiempos). |
| `features/usuarios`: `UsersPage`, `UserForm` | Listado paginado (nombre y apellidos, correo, teléfono, rol, estado, último acceso) + crear/editar; validación con Zod (obligatorios, correo, teléfono, rol, política "distinta de" nombre/apellidos/correo, máx. 64). |
| `features/roles`: `RolesPage`, `RoleForm`, `PermissionsGroup` | Casillas por módulo con estados (sellos "más adelante"), errores de uso y duplicado. |
| `features/auth`: `ForbiddenPage` | §3.9 (sin permiso). |
| `features/auth`: `ChangePasswordPage`, `useChangePassword` | Cambio propio y modo obligatorio. |
| Hooks de datos | `useUsers`, `useCreateUser`, `useUpdateUser`, `useSetUserStatus`, `useRoles`, `useCreateRole`, `useUpdateRole`, `useDeleteRole`, `useChangeMyPassword` sobre TanStack Query; ningún `fetch` en componentes. |
| `features/auditoria`: `AuditPage`, `AccessEventsList`, `AdminActionsList`, `AuditFilters`, `useAccessEvents` | Auditoría (§3.10): dos pestañas reutilizando `Table`, filtros y `Pagination`; sin botones de edición ni borrado. |

**Con pruebas** (Vitest + Testing Library, lógica visible): `LoginPage` (estados de error y bloqueo), `SessionWarning` (cuenta atrás), `RequirePermiso`, formularios de usuario y rol (Zod y errores del servidor), `PermissionsGroup`. El `dev-frontend` las escribe con MSW. En la **Auditoría**: `useAccessEvents` (filtros + página, el cambio de filtro reinicia a la página 1), cambio de pestaña conservando búsqueda, y el último acceso en la ficha de usuario (con y sin accesos, FR-021). También: que la vista de auditoría no ofrece ninguna acción de edición o borrado (FR-025).

---

## 6. Accesibilidad (WCAG 2.1 AA)

- **Teclado**: todo interactivo alcanzable y visible; orden de tabulación lógico (título → campos → botón). `Enter` envía el formulario de acceso. En listados, las acciones por fila son botones reales.
- **Foco**: visible siempre (`focus-visible` con anillo de 2 px `slate-900`); al aparecer un error, el foco va al `Notice` y, si es de campo, al primer campo inválido; tras iniciar sesión, foco en el `h1` del panel.
- **Semántica**: `<h1>` único por página; formularios con `<label>` explícito; errores con `aria-invalid` + `aria-describedby`; navegación con `<nav aria-label="Navegación del panel">`; listados con tablas semánticas (th con ámbito) y tarjetas equivalentes en móvil. En la Auditoría: las pestañas son botones con estado `aria-current` (o patrón de pestañas ARIA con flechas); la paginación se anuncia con el resumen "Mostrando {a}–{b} de {total}" tras cada cambio (`role="status"`).
- **Avisos**: error/alerta como `role="alert"`, éxito como `role="status"`. El aviso de sesión usa `role="alertdialog"` y trampa de foco.
- **Más que color**: activo/inactivo y permisos llevan texto explícito, nunca solo rojo/verde. Los errores no se comunican solo con color del borde.
- **Contraste y letra**: la base del panel es 18 px y 1.5 de interlineado; botones ≥ 44 px; sin texto por debajo de 12 px. La marca visual es el texto real, no `title` ni `placeholder`.
- **Contraseñas**: `type=password` con toggle accesible (`aria-pressed` + texto "Mostrar/ocultar"; no cambia la palabra). Nada de autocompletar por defecto fuera de lo correcto (`new-password` en formularios de nueva contraseña).
- **Movimiento**: nada animado de forma obligatoria; respeta `prefers-reduced-motion`.
- **Idioma**: `lang="es"` del documento; el texto de `role=` es coherente.

---

## 7. Catálogo de textos (español, llano)

> Reglas: nada técnico, nada de estados internos ("401", "token"), sin confirmar si una cuenta existe cuando no procede. `X` es un dato real de contexto.

| Situación | Texto |
|---|---|
| Credenciales incorrectas | "Correo o contraseña incorrectos." |
| Cuenta inactiva | "Ese acceso está desactivado. Pide a un administrador de la iglesia que lo reactive para poder entrar." |
| Bloqueo temporal | "Por seguridad, se han superado los intentos permitidos. Podrás volver a intentarlo a partir de las {hora}." |
| Sesión por caducar | "¿Sigues ahí? Tu sesión se cerrará en {m} minutos si no la usamos." |
| Sesión cerrada por techo absoluto | "Tu sesión terminó (cada sesión dura como máximo 1 hora). Vuelve a entrar para continuar." |
| Contraseña con datos personales | "La contraseña debe ser distinta del nombre, los apellidos y el correo." (+ máx. 64 caracteres) |
| Sesión terminada | "Tu sesión terminó. Vuelve a entrar para continuar." (+, si estaba con formulario: "La última acción no llegó a guardarse; vuelve a hacerla.") |
| Cuenta desactivada con sesión abierta | "Tu cuenta se desactivó. Para entrar de nuevo, pide a un administrador que la reactive." |
| Creado con éxito | "Cuenta creada." / "Rol creado." |
| Editado con éxito | "Cambios guardados." |
| Cambio de estado | "Cuenta activada." / "Cuenta desactivada." |
| Contraseña cambiada | "Contraseña cambiada. La próxima vez que entres, usa la nueva." |
| Correo duplicado | "Ya existe una cuenta con ese correo. Prueba otro correo." |
| Rol duplicado | "Ya existe un rol con ese nombre. Elige otro nombre." |
| Rol sin permisos | "Un rol necesita al menos un permiso. Marca al menos uno." |
| Rol en uso (eliminar) | "No se puede eliminar: {N} cuentas usan este rol. Primero cámbiales el rol y después podrás eliminarlo." |
| Último administrador | "No se puede {desactivar la cuenta / quitarle el permiso}: dejaría el panel sin ningún administrador activo. Activa o cambia a otra cuenta de administración y repite la acción." |
| Sin permiso | "No tienes acceso a esta sección. Pide a quien administra el panel que revise tu rol." |
| Datos con formato incorrecto | "Escribe el nombre." / "Escribe los apellidos." / "Este correo no tiene el formato correcto." / "Escribe un número de teléfono válido." / "Escribe la contraseña con las reglas indicadas abajo." |
| Sesión expirada al intentar acción | "Tu sesión terminó. Entra de nuevo y vuelve a hacer la acción; no se guardó a medias." |
| Error del sistema | "No se pudo completar la operación. Vuelve a intentarlo en unos minutos; si sigue, avísanos." |
| Confirmaciones destructivas | Desactivar: "La persona perderá el acceso de inmediato, aunque tenga sesión abierta; sus datos se conservan." · Eliminar rol: "Se borrará el rol '{nombre}' con sus {N} permisos. Las cuentas que lo usen no se verán afectadas si antes les asignas otro." |
| Auditoría vacía | "Todavía no hay registros." (+ con filtros: "Prueba a quitar los filtros o usar un rango de fechas más amplio.") |
| Último acceso de la cuenta | "{fecha}, a las {hora}, desde {IP}" · Sin accesos: "Nunca ha entrado al panel." |
| Resultado del acceso (píldora) | "Exitoso" / "Fallido" · Cuenta identificada: "{nombre de la cuenta}" · Intento sin cuenta identificada: fila con "Intento sin cuenta asociada". |
| Resultado de la acción (píldora) | "Completada" / "No completada" |
| Error de la auditoría | "No se pudo cargar el registro. Vuelve a intentarlo en unos minutos; si sigue, avísanos." |

---

## 8. Notas para el `dev-frontend` / `qa-tester`

- La **lógica con pruebas** está en hooks y formularios: política de contraseñas en Zod (compartida con el backend vía contrato: mín. 8, mayúsculas/minúsculas/números/especiales, **distinta de** nombre/apellidos/correo, máx. 64), validators de obligatorios (nombre, apellidos, correo, teléfono), cuenta atrás del bloqueo y de la sesión (30 min inactividad + techo absoluto de 1 hora, se congelan tiempos), y guardas de ruta.
- Cualquier hora ("a partir de las HH:MM") se calcula en cliente con la zona del navegador; el backend aporta la marca de tiempo.
- Alineado con el contrato: los listados (usuarios, roles y auditoría) usan **paginación** (`limit`/`offset`); sin filtros de texto libres en el MVP.
- La Auditoría es de solo lectura en toda la pila: el `dev-frontend` no implementa ninguna acción sobre los registros (FR-025) y el `qa-tester` verifica que intentar editar/borrar desde el panel no existe en la interfaz.
- Verificar también el **último acceso** en el listado de usuarios: cuenta nueva sin entrada → "Nunca ha entrado al panel." (FR-021).
- El diseño móvil primero se verifica a 320 px: tarjetas, sin desbordes, targets de 44 px.
