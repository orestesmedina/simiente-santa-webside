# Diseño UX: Estructura base del proyecto (F1)

**Feature Branch**: `001-estructura-base`
**Created**: 2026-09-29
**Status**: Diseño de referencia para la fase `plan` — **actualizado el 2026-09-30** al formato uniforme de respuestas del plan aprobado (D7, D13, D20) · **actualizado el 2026-10-03** con ajustes menores marcados «Ajuste (2026-10-03)» (hallazgos M3 y desviación 9 de la revisión de código; sin cambios de comportamiento).
**Input**: [spec.md](./spec.md) — FR-005, User Stories 2 y 6, escenarios de aceptación 2.1–2.4 y casos límite · [contracts/openapi.yaml](./contracts/openapi.yaml) (contrato de `/healthz`).

> **Nota de alcance**: este diseño es **meramente funcional**. La identidad de marca, los estilos, las animaciones y el bilingüismo están fuera de alcance en F1 (llegan en F3, según "Out of Scope" de la spec). Aquí se define estructura, contenido, estados y textos. Donde el desarrollador deba decidir algo visual, se indica; el estado **nunca** se comunica solo con color ni con animaciones.

## Cambios respecto de la versión anterior (2026-09-30)

**Motivo**: el plan de F1 fue aprobado el 2026-09-30 con un **formato uniforme de respuestas** (D13): el éxito es el DTO documentado y todo error 4xx/5xx es un sobre `{"error":{"code","message","details"?}}` (D7/D13). El diseño de la página — botón manual "Volver a consultar el estado", **sin auto-refresco**, botón deshabilitado durante la reconsulta, indicador "Actualizando…", hora de la última consulta y SC-005 (<3 s) — **no cambia**. Lo que se ajusta es de dónde viene cada estado y cómo se decide (delta para la revisión):

1. **Nuevo §3.1 — mapeo respuesta → estado**: tabla exacta de qué respuesta HTTP produce qué estado, con las reglas de validación del sobre. Es lo que pide el plan ("Ajustes pendientes en ux.md" §1).
2. **Corrección del origen del error A**: antes se suponía "respuesta OK con la base de datos no conectada"; con el sobre uniforme, la BD no conectada llega como un **503** con `ErrorEnvelope` `database_unavailable` (D7). El estado, su veredicto y sus textos **siguen igual**: cambia solo la vía.
3. **El error B se concreta**: timeout de **5 s** con `AbortController` (D20), error de red, respuesta que no cumple el sobre o sobre válido pero con un código no esperado → mismo veredicto "No se pudo consultar", ahora con dos *motivaciones internas* (`sin-respuesta`, `respuesta-inesperada`) que ajustan una línea del detalle para no afirmar algo falso (un servidor que respondió un error no está "sin respuesta").
4. **Defensa ante códigos desconocidos** dentro de un sobre válido: mensaje genérico, de modo que una versión futura del backend no pueda romper ni confundir la página.
5. **Un 200 con cuerpo no conforme** (vacío, incompleto o con `database` no reconocida) **nunca** muestra "conectado": el éxito exige validar el sobre de éxito completo.
6. **Decisiones de contenido sobre el sobre de error** (ajustes 2 y 3 del plan): `error.message` y `details` del 503 **no se muestran** en la interfaz; los literales de §6 mandan (FR-005: lenguaje llano y controlado). Reversible — ver Dudas abiertas.
7. **Componentes alineados con el plan**: la consulta vive en la capa `api/` (`apiFetch` + `ApiError`, D13) y en el hook TanStack Query (consulta al montar + `refetch` manual, D14); el botón se deshabilita mientras **cualquier** consulta está en curso, también la primera.

## 1. Flujo de usuario

Página única, pública, sin navegación ni autenticación (ambas llegan en F2/F3). Los "usuarios" reales de esta página son el equipo del proyecto y quien opera el sistema, pero los textos son comprensibles para cualquier persona (FR-005).

**Ajuste (2026-10-03) — encabezado con navegación mínima**: existe un encabezado de aplicación (`<header>`) con un `<nav aria-label="Navegación principal">` y un único enlace **«Inicio»** (`frontend/src/app/layout.tsx`), **fuera** del área de estado. Es el shell de la aplicación (semántico y accesible) pensado para F2/F3; en F1 no hay más rutas ni secciones, así que el flujo de arriba no cambia. La regla de §2 se mantiene: dentro del área de estado, las acciones disponibles siguen siendo **exactamente una** (volver a consultar). La autenticación sigue sin existir.

**Flujo principal:**

1. La persona abre la URL raíz del frontend en el navegador (móvil o escritorio).
2. La página carga y **consulta automáticamente** el estado del backend **una sola vez al cargar** (consulta al montar con TanStack Query; **sin auto-refresco** en F1, D14). Mientras no haya respuesta, se muestra el estado *Consultando*.
3. Si hay respuesta interpretable, la página muestra el **estado actual** (veredicto + detalle de los dos componentes + hora de la última consulta). Fin del flujo: la persona ya sabe si el sistema está sano, sin realizar ninguna acción (escenario 2.3, SC-005: estado visible en <3 s desde que la página termina de cargar).
4. Si la persona quiere comprobar de nuevo (por ejemplo, tras levantar la base de datos), pulsa **"Volver a consultar el estado"**: la página reconsulta y **reemplaza** lo mostrado por el resultado nuevo, nunca un estado memorizado (FR-003, escenario 2.4).

**Caminos de error** (los dos previstos; el mapeo exacto está en §3.1):

- **Error A — respuesta recibida, base de datos no conectada** (`503` + `error.code = "database_unavailable"`, D7): la página muestra el estado *Base de datos no conectada* (escenario 2.2, FR-004: el backend sigue respondiendo). El sobre llega en ≤2 s por el timeout del ping del backend (D7/D8), antes del timeout de 5 s del frontend (D20): nunca se confunde con "sin respuesta".
- **Error B — sin respuesta interpretable**: timeout de 5 s (D20), error de red o conexión, respuesta que no cumple el sobre, o sobre con un código no esperado → la página muestra el estado *No se pudo consultar*, distinguiéndolo explícitamente del error A en el propio texto: aquí el estado de la base de datos **no se puede saber** (User Story 2: un fallo de conexión a la BD no debe confundirse con un fallo del backend).

**Transición entre estados (escenario 2.4 y caso límite "BD se cae en ejecución"):** cada consulta nueva sustituye al resultado anterior, en cualquier dirección (conectada → no conectada y al revés). El área de estado es una región `aria-live`, así que las personas con lector de pantalla oyen el cambio sin recargar la página.

## 2. Pantalla única: "Estado del sistema"

**Propósito**: mostrar de un vistazo si el sistema del sitio está funcionando, distinguiendo "base de datos conectada" de "no conectada" y de "no se pudo consultar".

**Contenido** (de arriba abajo, una sola columna; el orden del documento es el orden de lectura):

| # | Elemento | Contenido |
|---|---|---|
| 1 | Título de página (`<title>`) | "Estado del sistema · Iglesia Simiente Santa" |
| 2 | Encabezado `<header>` con `<nav aria-label="Navegación principal">` | Enlace único **«Inicio»** — **Ajuste (2026-10-03)**: navegación mínima, **fuera** del área de estado (ver el ajuste de §1) |
| 3 | `<h1>` | "Estado del sistema" |
| 4 | Párrafo introductorio | Qué es esta página y nota de sitio en construcción (texto en §6) |
| 5 | Encabezado `<h2>` | "Estado actual" |
| 6 | **Región de estado** (en vivo, `aria-live`) | Según el estado de la vista (§3): veredicto, detalle de los dos componentes (servidor, base de datos) y hora de la última consulta |
| 7 | Botón | "Volver a consultar el estado" |

**Detalle de los dos componentes** (visible en todos los estados salvo *Consultando* inicial): una lista de definición (`<dl>`) con dos entradas — *Servidor* y *Base de datos* —, cada una con su palabra de estado ("en marcha", "conectada", "no conectada", "sin respuesta", "no se pudo comprobar"). Dos reglas fijas:

- En el estado *No se pudo consultar*, la base de datos figura como **"no se pudo comprobar"**, nunca como "no conectada": sin una respuesta interpretable no hay forma de saberlo, y la spec exige distinguir ambos fallos.
- En la variante *respuesta inesperada* de ese mismo estado (§3.3), el **servidor** también figura como "no se pudo comprobar": respondió, pero no con un estado utilizable.

**Hora de la última consulta**: se muestra siempre que hay un resultado (éxito o error), con la hora exacta **del intento que produjo ese resultado** — la marca del éxito o del fallo que se está viendo, nunca la del último éxito anterior. Hace visible que lo mostrado es el estado *del momento de esa consulta* (FR-003), y sirve al equipo para saber si lo que ve es fresco o antiguo.

**Acciones disponibles**: exactamente una — volver a consultar. No hay navegación, enlaces, formularios ni autenticación.

**Ajuste (2026-10-03)**: el encabezado de la aplicación sí incluye una navegación mínima (el enlace **«Inicio»** descrito en §1, fuera del área de estado y de su región `aria-live`). **Dentro** del área de estado no cambia nada: las acciones siguen siendo exactamente una (volver a consultar) y no hay enlaces, formularios ni autenticación.

**Nota de sitio en construcción**: el párrafo introductorio incluye una frase que explica que el sitio está en construcción y que por ahora solo existe esta página. Evita confusión si un miembro de la iglesia llega aquí por error. Es una decisión de contenido del diseño (no está pedida literalmente en la spec); debe ser trivial de quitar o reemplazar en F3.

## 3. Estados de la vista

### 3.1 Mapeo respuesta → estado (formato uniforme de respuestas, D7/D13/D20)

La respuesta de `/healthz` pasa por **una sola puerta de datos**: `getSystemStatus()` en la capa `api/` (mecanismo único de `apiFetch<T>` + `ApiError`, D13), con `AbortController` y timeout de **5 s** (D20). Esa función valida el cuerpo de éxito (`status === "ok"` y `database === "connected"`) y devuelve el DTO `SystemStatus` tipado, o **lanza** `ApiError` (sobre de error, red, timeout o cuerpo no conforme): no traduce a estados ni devuelve nunca un cuerpo crudo.

**Ajuste (2026-10-03) — dónde vive la traducción**: la traducción de ese resultado a uno de los estados de `EstadoSistema` (§4) la hace la feature `status`, en `toEstadoSistema()`, invocada por el hook `useSystemStatus()`. El contrato de datos vive en `api/` y la traducción a estados en la feature; el invariante es el mismo de siempre: **el resto de la página no ve nunca el cuerpo crudo, los códigos HTTP ni los códigos de error; solo estados.** La tabla de mapeo de abajo describe el resultado final de esa cadena, sea cual sea el punto donde se decida.

| Lo que llega | Condición | Estado resultante |
|---|---|---|
| `200` + `SystemStatus` | Cuerpo válido: objeto con `status` y `database` y `database === "connected"` | **Conectado** (éxito) |
| `503` + sobre de error | Sobre válido con `error.code === "database_unavailable"` | **Base de datos no conectada** (error A) |
| Sobre de error válido, otro código | `code = internal`, `not_found`, `method_not_allowed` o **un código desconocido** (versión futura del backend) | **No se pudo consultar** — variante *respuesta inesperada* (defensiva, mensaje genérico) |
| **Sin respuesta** | Timeout de 5 s (`AbortController`, D20) o error de red / conexión | **No se pudo consultar** — variante *sin respuesta* |
| **Respuesta no conforme** | Cualquier estado HTTP (`200` incluido) con cuerpo que no cumple el sobre: HTML o texto plano, JSON malformado, `SystemStatus` incompleto o con un valor de `database` no reconocido | **No se pudo consultar** — variante *respuesta inesperada* |

**Reglas de validación (defensivas):**

- **El éxito exige el sobre completo**: un `200` con cuerpo vacío, incompleto o con una `database` distinta de `"connected"` **no** se muestra como "conectado" → va a *No se pudo consultar* (variante *respuesta inesperada*). Si una versión futura del contrato amplía los valores de `SystemStatus.database`, este mapeo se revisa entonces (ver Dudas abiertas); hoy no se inventa un estado "degradado".
- **Solo `database_unavailable` es error A**: ningún otro código del registro — presente o futuro — representa "BD no conectada".
- **Cero ramas sin estado**: si la validación falla por la causa que sea (cuerpo raro, código raro, red lenta), el resultado es siempre uno de los estados de la tabla; jamás una página en blanco ni un error crudo.
- **Código y `error.message` se leen, no se muestran**: sirven para elegir el estado; el texto frente a la persona lo pone §6.

### 3.2 Máquina de estados

`consultando` → (`200` + `SystemStatus` válido) → `conectado` · (`503` + `database_unavailable`) → `bd-no-conectada` · (cualquier otra cosa: timeout, red, sobre inesperado o no conforme) → `inaccesible`.

Desde cualquier estado **con resultado**, el botón vuelve a `consultando`; el resultado de cada consulta **reemplaza** siempre al anterior. El botón está deshabilitado mientras hay una consulta en curso (primera o de reconsulta), para evitar consultas solapadas.

### 3.3 Estados

| Estado | Cuándo (vía el mapeo de §3.1) | Qué se muestra exactamente |
|---|---|---|
| **Consultando** | Consulta automática al cargar (sin dato aún) **o** reconsulta | Veredicto: "Consultando el estado del sistema…" — texto estático, **sin** animación (fuera de alcance). **Primera consulta**: solo el veredicto; no hay detalle ni hora, porque todavía no se ha comprobado nada. **Reconsulta**: el resultado anterior sigue visible marcado como "Actualizando…" y el botón queda deshabilitado. |
| **Todo funcionando** (conectado) | `200` + `SystemStatus` válido (escenario 2.1) | Veredicto: "El sistema está funcionando". Detalle: Servidor → *en marcha*; Base de datos → *conectada*. Hora de la última consulta. |
| **Base de datos no conectada** (error A) | `503` + `error.code = "database_unavailable"` (escenario 2.2; caso límite "BD tarda en arrancar") | Veredicto: "El sistema está en marcha, pero la base de datos no está conectada". Detalle: Servidor → *en marcha*; Base de datos → *no conectada*. Explicación breve (incluye la pista de que la BD puede estar detenida o aún arrancando) y hora de la consulta. El `message` y `details` del sobre **no se muestran** (decisión en §7; los literales de §6 mandan). |
| **No se pudo consultar** (error B) | Sin respuesta interpretable: dos motivaciones internas, **mismo veredicto**. *(a) sin respuesta*: timeout de 5 s o error de red. *(b) respuesta inesperada*: sobre válido con código no esperado (p. ej. `internal`), código desconocido, o cuerpo que no cumple el sobre. | Veredicto: "No se pudo consultar el estado del sistema". Detalle: Base de datos → *no se pudo comprobar* siempre; Servidor → *sin respuesta* en (a) y *no se pudo comprobar* en (b). Explicación que distingue explícitamente este estado del error A (aquí el problema es el servidor o la conexión con él, y de la base de datos no se sabe nada), sugerencia de reintentar y hora del intento. Textos exactos en §6. |

### 3.4 Estados que no aplican (se documentan para no dejar huecos)

- **Vacío**: no existe. La vista muestra un único dato (el estado), no colecciones; la ausencia de dato es el estado *Consultando*.
- **Sin permisos**: no existe. La página es pública (spec, Assumptions); la autenticación llega en F2.
- **Parcial/caché**: prohibido. Nunca se muestra un estado anterior como si fuera actual: cada consulta reemplaza al resultado previo (FR-003, escenario 2.4). La única excepción visible es la *reconsulta*, donde el resultado anterior **sigue marcado** como "Actualizando…" hasta que llega el nuevo.
- **Estado degradado**: no existe en F1. Un `200` con una `database` no reconocida no es "algo intermedio": es *respuesta inesperada* → *No se pudo consultar* (§3.1). Si el contrato futuro define valores intermedios, se rediseña entonces.

## 4. Componentes React

No hay componentes existentes: este es el primer código del frontend. Lista mínima (mantenerla corta; no crear componentes de botón ni de tarjeta de diseño — eso llega con el sistema de diseño de F3). Alineada con la estructura del plan (D14): la página vive en `features/status/pages/StatusPage.tsx`, el hook en `features/status/hooks/useSystemStatus.ts`, el fetch en `api/status.ts` y la traducción a estados en `features/status/toEstadoSistema.ts` (**Ajuste 2026-10-03**: ver fila `toEstadoSistema`).

| Componente / unidad | Responsabilidad | Props principales |
|---|---|---|
| `AppLayout` (`app/layout.tsx`, shell de la aplicación) | **Ajuste (2026-10-03)**: encabezado con `<nav aria-label="Navegación principal">` y enlace único **«Inicio»**, más la región `<main>` donde renderiza la ruta activa. Vive **fuera** del área de estado y de su región `aria-live`; no añade acciones a la pantalla de §2 | — (renderiza `<Outlet>`) |
| `PaginaEstado` (será `StatusPage.tsx`) | Página completa: título, intro, región de estado en vivo y el único botón. **No hace `fetch`** (prohibido por la skill): consulta al montar vía el hook y al pulsar el botón vía `refetch`. Deshabilita el botón mientras hay consulta en curso (primera o de reconsulta) | — (la URL la resuelve la capa `api/`; no es prop) |
| `useSystemStatus` (hook) | TanStack Query sobre `getSystemStatus()`: **consulta al montar + refetch manual, sin auto-refresco** (D14); traduce cada resultado con `toEstadoSistema` y expone resultado (`EstadoSistema`), hora de **ese intento** (éxito o error) y bandera de consulta en curso | — |
| `getSystemStatus` (función, `api/status.ts`) | Único punto de entrada del fetch: `AbortController` con timeout de **5 s** (D20) vía `apiFetch`; valida el cuerpo de éxito y devuelve el DTO `SystemStatus` tipado **o lanza** `ApiError` (**Ajuste 2026-10-03**: no traduce a estados; nunca devuelve un cuerpo crudo) | — |
| `toEstadoSistema` (función, `features/status/toEstadoSistema.ts`) | **Ajuste (2026-10-03)**: traduce el resultado de `getSystemStatus()` (éxito o `ApiError`) a **siempre** un `EstadoSistema` según la tabla §3.1; cero ramas sin estado | `result: { ok: true; value: SystemStatus } \| { ok: false; error: unknown }` |
| `ResultadoEstado` | Renderiza el estado actual: veredicto, detalle de los dos componentes, hora de última consulta y explicación del error si lo hay | `estado: EstadoSistema`, `fechaConsulta?: Date`, `actualizando: boolean` |
| `ItemVerificacion` | Una fila del detalle: término + palabra de estado | `nombre: string`, `valor: 'ok' \| 'error' \| 'sin-respuesta' \| 'no-comprobable'` |

**Tipos compartidos** (los define `dev-frontend` junto al contrato; el mapeo respuesta → estado se registra aquí, como pide el plan):

```ts
// La respuesta de /healthz se mapea SIEMPRE a uno de estos (tabla §3.1),
// mediante `toEstadoSistema` (feature `status`; ver Ajuste 2026-10-03 en §3.1).
// El `motivo` de `inaccesible` es interno: guía pruebas y diagnóstico;
// frente a la persona solo cambia la palabra del detalle del servidor,
// nunca aparecen códigos HTTP ni códigos de error (§3.1, §6).
export type EstadoSistema =
  | { kind: 'conectado' }                    // 200 + SystemStatus válido
  | { kind: 'bd-no-conectada' }              // 503 + error.code "database_unavailable" (sobre de error, D7)
  | { kind: 'inaccesible'; motivo: 'sin-respuesta' | 'respuesta-inesperada' };
```

> Corrección respecto de la versión anterior: `bd-no-conectada` ya **no** es "respuesta OK con BD no conectada" — con el sobre uniforme es un `503` con `ErrorEnvelope` (D7). Y `motivo` pasa de opcional a obligatorio: siempre se sabe por qué no hubo respuesta interpretable.

El botón es un `<button>` nativo dentro de `PaginaEstado`; no se justifica un componente propio en F1.

## 5. Accesibilidad (WCAG 2.1 AA, en lo aplicable a una página de texto)

- **Idioma y semántica**: `<html lang="es">`; `<title>` descriptivo; un único `<h1>`; detalle como `<dl>`; hora de consulta como `<time>`; `<meta name="viewport" content="width=device-width, initial-scale=1">` (imprescindible para móvil).
- **Región en vivo**: el área de estado usa `role="status"` + `aria-live="polite"` para que los cambios de estado — incluidas las transiciones conectado → no conectada, conectado → *No se pudo consultar* y entre las dos variantes del error B — se anuncien a lectores de pantalla sin recargar. El botón queda **fuera** de esa región para no ensuciar los anuncios.
- **Estructura de lectura constante**: los cuatro estados comparten el mismo orden — veredicto → detalle → hora → explicación —, de modo que la lectura visual y el anuncio del lector de pantalla sean previsibles en cualquier estado.
- **Teclado**: la página es navegable por teclado con el orden natural del documento; dentro del área de estado el único control es el botón, y antes en el orden aparece el enlace **«Inicio»** del encabezado (**Ajuste 2026-10-03**; también queda fuera de la región en vivo, igual que el botón). No interceptar teclas ni capturar foco. Mantener el contorno de foco por defecto del navegador. Con el botón deshabilitado durante la consulta, el foco no se pierde: la deshabilitación es de *pulsación*, no de visibilidad.
- **Estado ≠ color** (WCAG 1.4.1): el estado se comunica con **texto** (veredicto + palabras de estado). Si el desarrollo añade color o iconos como refuerzo, nunca serán el único medio, y en F1 no se piden.
- **Contraste** (WCAG 1.4.3): con los estilos por defecto del navegador (texto negro sobre blanco) se cumple AA. Si se toca algo, mínimo 4.5:1 para texto.
- **Zoom y reflow** (WCAG 1.4.4 / 1.4.10): sin estilos que fijen tamaños: la página debe verse y usarse a 200 % de zoom y en un viewport de 320 px de ancho (un móvil pequeño), con reflujo natural en una columna.
- **Lenguaje llano**: español claro y oraciones cortas, pensado para personas de todas las edades (idea.md: públicos de toda la iglesia). Los términos del sobre (`code`, `details`, HTTP, "503") **nunca** aparecen frente a la persona; todo se dice con las palabras de §6.
- **Animaciones**: ninguna (además de estar fuera de alcance, beneficia a personas con sensibilidad al movimiento y a conexiones lentas).
- **Objetivo táctil**: recomendación mínima funcional (no diseño): botón con alto ≥ 44 px para pulgar. Si el equipo decide cero CSS en F1, el botón nativo es aceptable y este punto queda para F3.

## 6. Textos (literales, español)

**Introducción estática:**

> Este sitio web está en construcción. Por ahora, esta página muestra si el sistema de la Iglesia Simiente Santa está funcionando correctamente.

**Estado *Consultando* (primera consulta, en curso):**

> Consultando el estado del sistema…

*(Mientras la primera consulta está en curso no se muestra detalle ni hora: todavía no se ha comprobado nada.)*

**Durante una reconsulta (resultado anterior visible):**

> Actualizando el estado…

**Estado *Todo funcionando* (`200` + `SystemStatus`):**

> **El sistema está funcionando.**
> Servidor: en marcha.
> Base de datos: conectada.
> Última consulta: [hora exacta, p. ej. "14:32:05"].

**Estado *Base de datos no conectada* (`503` + `database_unavailable`):**

> **El sistema está en marcha, pero la base de datos no está conectada.**
> Servidor: en marcha.
> Base de datos: no conectada.
> Última consulta: [hora exacta del intento].
> El sitio web está corriendo, pero no puede leer ni guardar información por ahora. Esto suele ocurrir cuando la base de datos está detenida o todavía está arrancando.

*(El texto de `error.message` del sobre no se muestra aquí: este literal propio ya lo dice en llano y mantiene el tono de toda la página — decisión en §7.)*

**Estado *No se pudo consultar* — variante *sin respuesta* (timeout de 5 s, error de red):**

> **No se pudo consultar el estado del sistema.**
> Servidor: sin respuesta.
> Base de datos: no se pudo comprobar.
> Última consulta: [hora exacta del intento].
> No hubo respuesta del servidor, así que no sabemos si el sistema está funcionando. Esto no significa que la base de datos esté desconectada: aquí el problema está en el propio servidor o en la conexión con él. Espere unos momentos y vuelva a consultar.

**Estado *No se pudo consultar* — variante *respuesta inesperada* (sobre con código no esperado o desconocido, y respuesta que no cumple el sobre):**

> **No se pudo consultar el estado del sistema.**
> Servidor: no se pudo comprobar.
> Base de datos: no se pudo comprobar.
> Última consulta: [hora exacta del intento].
> El servidor respondió, pero con algo que no pudimos interpretar, así que no sabemos si el sistema está funcionando. Esto no significa que la base de datos esté desconectada: aquí el problema está en el propio servidor o en la conexión con él. Espere unos momentos y vuelva a consultar.

**Botón:**

> Volver a consultar el estado

## 7. Decisiones de diseño, dudas abiertas y huecos

### 7.1 Decisiones

Decisiones tomadas aquí o resueltas por el plan aprobado (con referencia):

1. **Auto-refresco / polling** — *resuelto en el plan (D14)*: consulta al montar + botón manual; `refetchInterval` queda como nota para F3+ si el cliente lo pide. Sin cambios en el diseño.
2. **Timeout de la consulta** — *resuelto en el plan (D20)*: `AbortController` con timeout de **5 s**; timeout, error de red y respuesta no conforme → *No se pudo consultar*. Coherencia con SC-005: el backend responde en ≤2 s aunque la BD esté caída (ping con timeout, D7/D8), así que el caso normal cumple el <3 s; los 5 s solo acotan el caso patológico (backend colgado), y D20 descarta timeouts ≤2 s porque producirían falsos "No se pudo consultar" en BD lentas.
3. **Respuesta ininterpretable** — *resuelto y ampliado aquí*: la versión anterior lo agrupaba con el estado B "con el mismo texto" y esperaba que el contrato lo imposibilitara. Con el formato uniforme el caso es **posible y legítimo** (sobre válido con `internal`, fallbacks del router 404/405, proxies intermedios) y se trata con motivo propio (*respuesta inesperada*) y texto factual: un servidor que respondió un error no aparece como "sin respuesta".
4. **Nota "sitio en construcción"** en la intro: añadido de contenido no pedido literalmente por la spec (evita confusión en visitantes reales); decidido aquí, trivial de retirar en F3. Sin cambios.
5. **URL del backend configurable** (`VITE_API_URL`, FR-008): necesaria para que dev/prod apunten a hosts distintos. Sin cambios; la variable la documenta `.env.example`.
6. **`error.message` del sobre no se muestra** (ajuste 2 del plan): el único uso del mensaje del sobre es ayudar a elegir el estado; frente a la persona mandan los literales de §6, que son llanos, constantes y del tono de toda la página. Un texto proveniente del servidor podría variar de tono en versiones futuras; con esta decisión la página nunca habla con jerga ni de forma impredecible.
7. **`details` del 503 (`{"database":"disconnected"}`) no se muestra** (ajuste 3 del plan): no aporta una acción nueva (la acción es la misma: volver a consultar) y repite lo que el veredicto ya dice en llano. Es una decisión de contenido reversible: si quien opera quisiera verla, cabría un detalle técnico plegable en F2+.
8. **Defensa activa ante versiones futuras**: el éxito exige validar el sobre de éxito completo, y solo `database_unavailable` es error A; cualquier código desconocido o cuerpo no conforme degrada a *No se pudo consultar* con mensaje genérico. Así, una versión futura del backend no puede hacer que la página muestre un estado falso ni deje de funcionar.
9. **Navegación mínima en el encabezado** (**Ajuste 2026-10-03**, hallazgo M3 de la revisión de código): el shell (`AppLayout`, `app/layout.tsx`) sí lleva `<header>` + `<nav aria-label="Navegación principal">` con el enlace «Inicio», **fuera** del área de estado y de su región `aria-live`. Decisión: se documenta, no se quita (es semántico, accesible y prepara F2/F3). Dentro del área de estado no cambia nada: una sola acción (volver a consultar). Reflejado en §1, §2, §4 y §5.
10. **Dónde se traduce la respuesta a estados** (**Ajuste 2026-10-03**, desviación 9 de la revisión de código): el contrato de datos vive en `api/` (`getSystemStatus()` devuelve `SystemStatus` o lanza `ApiError`) y la traducción a `EstadoSistema` en la feature (`toEstadoSistema()`, invocada por el hook). Decisión: se acepta la separación porque el invariante de §3.1 se conserva — la página solo ve estados, nunca cuerpos crudos, códigos HTTP ni `error.code`. Reflejado en §3.1 y §4.

### 7.2 Dudas abiertas (se reportan, no se suponen resueltas)

1. **Usar `error.message` como explicación de apoyo en el error A**: el plan lo permite ("los literales de §6 siguen mandando si difieren"). Decisión actual: no mostrarlo (7.1.6). Pedir criterio al humano en la revisión; si se prefiere usar el mensaje del servidor, el cambio es solo de contenido de §6.
2. **Visibilidad de `details` para quien opera**: hoy oculto (7.1.7). Si la operación pide el valor `disconnected` a la vista, decidir dónde (plegable técnico) en F2+, no en F1.
3. **Valores futuros de `SystemStatus.database`**: el contrato lo fija como `const "connected"`, así que hoy no hay nada que decidir; si una versión futura añade valores intermedios (p. ej. un "degradado"), este diseño no los tiene: cada valor nuevo requerirá revisar la tabla de §3.1 y, previsiblemente, un estado nuevo de `EstadoSistema`.
4. **Códigos válidos del registro pero inaplicables a `/healthz`** (`not_found`, `method_not_allowed`): solo deberían aparecer por configuración errónea (mal `VITE_API_URL`); reciben el texto genérico de la variante *respuesta inesperada*. No se les da microcopy específico (no se documenta ese escenario de configuración en la spec).
5. **Distinguir las dos variantes del error B de forma aún más explícita** (p. ej. un texto distinto para "el servidor está caído" frente a "respondió algo raro"): hoy comparten veredicto y solo difiere la línea del servidor y una frase de la explicación; suficiente para F1. Si quien opera necesita más granularidad, se decide en una funcionalidad posterior.
6. **Respuesta que llega después del timeout**: cuando `AbortController` corta a los 5 s, cualquier respuesta tardía se ignora y el estado ya mostrado manda hasta la siguiente consulta manual. No hay auto-refresco que la recoja (D14); se anota para que el comportamiento no se interprete como un bug en pruebas.
