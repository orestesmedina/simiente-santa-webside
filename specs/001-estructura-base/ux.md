# Diseño UX: Estructura base del proyecto (F1)

**Feature Branch**: `001-estructura-base`
**Created**: 2026-09-29
**Status**: Diseño de referencia para la fase `plan`
**Input**: [spec.md](./spec.md) — FR-005, User Stories 2 y 3, escenarios de aceptación 2.1–2.4 y casos límite.

> **Nota de alcance**: este diseño es **meramente funcional**. La identidad de marca, los estilos, las animaciones y el bilingüismo están fuera de alcance en F1 (llegan en F3, según "Out of Scope" de la spec). Aquí se define estructura, contenido, estados y textos. Donde el desarrollador deba decidir algo visual, se indica; el estado **nunca** se comunica solo con color ni con animaciones.

## 1. Flujo de usuario

Página única, pública, sin navegación ni autenticación (ambas llegan en F2/F3). Los "usuarios" reales de esta página son el equipo del proyecto y quien opera el sistema, pero los textos son comprensibles para cualquier persona (FR-005).

**Flujo principal:**

1. La persona abre la URL raíz del frontend en el navegador (móvil o escritorio).
2. La página carga y **consulta automáticamente** el estado del backend. Mientras no haya respuesta, se muestra el estado *Consultando*.
3. Si hay respuesta interpretable, la página muestra el **estado actual** (veredicto + detalle de los dos componentes + hora de la última consulta). Fin del flujo: la persona ya sabe si el sistema está sano, sin realizar ninguna acción (escenario 2.3, SC-005).
4. Si la persona quiere comprobar de nuevo (por ejemplo, tras levantar la base de datos), pulsa **"Volver a consultar el estado"**: la página reconsulta y **reemplaza** lo mostrado por el resultado nuevo, nunca un estado memorizado (FR-003, escenario 2.4).

**Caminos de error:**

- **Error A — respuesta recibida, base de datos no conectada**: la página muestra el estado *Base de datos no conectada* (escenario 2.2, FR-004: el backend sigue respondiendo).
- **Error B — sin respuesta del backend**: la página muestra el estado *No se pudo consultar*, distinguiéndolo explícitamente del error A en el propio texto: aquí el problema es el servidor o la conexión con él, y el estado de la base de datos **no se puede saber** (User Story 2: un fallo de conexión a la BD no debe confundirse con un fallo del backend).

**Transición entre estados (escenario 2.4 y caso límite "BD se cae en ejecución"):** cada consulta nueva sustituye al resultado anterior, en cualquier dirección (conectada → no conectada y al revés). El área de estado es una región `aria-live`, así que las personas con lector de pantalla oyen el cambio sin recargar la página.

## 2. Pantalla única: "Estado del sistema"

**Propósito**: mostrar de un vistazo si el sistema del sitio está funcionando, distinguiendo "base de datos conectada" de "no conectada" y de "no se pudo consultar".

**Contenido** (de arriba abajo, una sola columna; el orden del documento es el orden de lectura):

| # | Elemento | Contenido |
|---|---|---|
| 1 | Título de página (`<title>`) | "Estado del sistema · Iglesia Simiente Santa" |
| 2 | `<h1>` | "Estado del sistema" |
| 3 | Párrafo introductorio | Qué es esta página y nota de sitio en construcción (texto en §6) |
| 4 | Encabezado `<h2>` | "Estado actual" |
| 5 | **Región de estado** (en vivo, `aria-live`) | Según el estado de la vista (§3): veredicto, detalle de los dos componentes (servidor, base de datos) y hora de la última consulta |
| 6 | Botón | "Volver a consultar el estado" |

**Detalle de los dos componentes** (visible en los estados con respuesta, A y éxito): una lista de definición (`<dl>`) con dos entradas — *Servidor* y *Base de datos* —, cada una con su palabra de estado ("en marcha", "conectada", "no conectada", "sin respuesta", "no se pudo comprobar"). En el estado *No se pudo consultar*, la base de datos figura como **"no se pudo comprobar"**, nunca como "no conectada": sin backend no hay forma de saberlo, y la spec exige distinguir ambos fallos.

**Hora de la última consulta**: se muestra siempre que hay un resultado (éxito o error), con la hora exacta. Hace visible que lo mostrado es el estado *del momento de esa consulta* (FR-003), y sirve al equipo para saber si lo que ve es fresco o antiguo.

**Acciones disponibles**: exactamente una — volver a consultar. No hay navegación, enlaces, formularios ni autenticación.

**Nota de sitio en construcción**: el párrafo introductorio incluye una frase que explica que el sitio está en construcción y que por ahora solo existe esta página. Evita confusión si un miembro de la iglesia llega aquí por error. Es una decisión de contenido del diseño (no está pedida literalmente en la spec); debe ser trivial de quitar o reemplazar en F3.

## 3. Estados de la vista

Máquina de estados: `consultando` → (respuesta interpretable) → `conectado` **o** `bd-no-conectada` · (sin respuesta o ininterpretable) → `inaccesible`. Desde cualquier estado, el botón vuelve a `consultando`; el resultado de cada consulta **reemplaza** siempre al anterior.

| Estado | Cuándo | Qué se muestra exactamente |
|---|---|---|
| **Consultando** | Carga inicial (consulta automática en curso, sin dato aún); también al reconsultar, *conservando visible el resultado anterior* marcado como "Actualizando…" | Veredicto: "Consultando el estado del sistema…" — texto estático, **sin** animación (fuera de alcance). En reconsulta: el resultado anterior sigue visible y el botón queda desactivado para evitar consultas solapadas. |
| **Todo funcionando** (conectado) | Respuesta interpretable: servidor vivo y BD conectada (escenario 2.1) | Veredicto: "El sistema está funcionando". Detalle: Servidor → *en marcha*; Base de datos → *conectada*. Hora de la última consulta. |
| **Base de datos no conectada** (error A) | Respuesta interpretable: servidor vivo, BD no conectada (escenario 2.2; caso límite "BD tarda en arrancar") | Veredicto: "El sistema está en marcha, pero la base de datos no está conectada". Detalle: Servidor → *en marcha*; Base de datos → *no conectada*. Explicación breve (incluye la pista de que la BD puede estar detenida o aún arrancando) y hora de la consulta. |
| **No se pudo consultar** (error B) | Sin respuesta del backend (caído, inalcanzable, sin responder a tiempo) o respuesta no interpretable | Veredicto: "No se pudo consultar el estado del sistema". Detalle: Servidor → *sin respuesta*; Base de datos → *no se pudo comprobar*. Explicación que distingue explícitamente este caso del error A, sugerencia de reintentar y hora del intento. |

**Estados que no aplican** (se documentan para no dejar huecos):

- **Vacío**: no existe. La vista muestra un único dato (el estado), no colecciones; la ausencia de dato es el estado *Consultando*.
- **Sin permisos**: no existe. La página es pública (spec, Assumptions); la autenticación llega en F2.
- **Parcial/caché**: prohibido. Nunca se muestra un estado anterior como si fuera actual: cada consulta reemplaza al resultado previo (FR-003, escenario 2.4).

## 4. Componentes React

No hay componentes existentes: este es el primer código del frontend. Lista mínima (mantenerla corta; no crear componentes de botón ni de tarjeta de diseño — eso llega con el sistema de diseño de F3):

| Componente | Responsabilidad | Props principales |
|---|---|---|
| `PaginaEstado` | Página completa: título, intro, región de estado y botón. Posee la lógica: consulta al montar, consulta al pulsar el botón, timeout de la consulta, mapeo respuesta → estado | `endpoint: string` (URL de `/healthz`, leída de la variable de entorno de Vite; valor de ejemplo documentado en `.env.example` — FR-008) |
| `ResultadoEstado` | Renderiza el estado actual: veredicto, detalle de los dos componentes, hora de última consulta y explicación del error si lo hay | `estado: EstadoSistema`, `fechaConsulta?: Date`, `actualizando: boolean` |
| `ItemVerificacion` | Una fila del detalle: término + palabra de estado | `nombre: string`, `valor: 'ok' \| 'error' \| 'sin-respuesta' \| 'no-comprobable'` |

**Tipos compartidos** (los define `dev-frontend` junto al contrato que acuerde el arquitecto):

```ts
// Discriminated union: la respuesta de /healthz se mapea SIEMPRE a uno de estos.
type EstadoSistema =
  | { kind: 'conectado' }        // respuesta OK, BD conectada
  | { kind: 'bd-no-conectada' }  // respuesta OK, BD no conectada
  | { kind: 'inaccesible', motivo?: string }; // sin respuesta, timeout o respuesta ininterpretable
```

El botón es un `<button>` nativo dentro de `PaginaEstado`; no se justifica un componente propio en F1.

## 5. Accesibilidad (WCAG 2.1 AA, en lo aplicable a una página de texto)

- **Idioma y semántica**: `<html lang="es">`; `<title>` descriptivo; un único `<h1>`; detalle como `<dl>`; hora de consulta como `<time>`; `<meta name="viewport" content="width=device-width, initial-scale=1">` (imprescindible para móvil).
- **Región en vivo**: el área de estado usa `role="status"` + `aria-live="polite"` para que los cambios de estado (incluida la transición conectada → no conectada) se anuncien a lectores de pantalla sin recargar. El botón queda **fuera** de esa región para no ensuciar los anuncios.
- **Teclado**: la página es navegable por teclado con el orden natural del documento; el único control es el botón. No interceptar teclas ni capturar foco. Mantener el contorno de foco por defecto del navegador.
- **Estado ≠ color** (WCAG 1.4.1): el estado se comunica con **texto** (veredicto + palabras de estado). Si el desarrollo añade color o iconos como refuerzo, nunca serán el único medio, y en F1 no se piden.
- **Contraste** (WCAG 1.4.3): con los estilos por defecto del navegador (texto negro sobre blanco) se cumple AA. Si se toca algo, mínimo 4.5:1 para texto.
- **Zoom y reflow** (WCAG 1.4.4 / 1.4.10): sin estilos que fijen tamaños: la página debe verse y usarse a 200 % de zoom y en un viewport de 320 px de ancho (un móvil pequeño), con reflujo natural en una columna.
- **Animaciones**: ninguna (además de estar fuera de alcance, beneficia a personas con sensibilidad al movimiento y a conexiones lentas).
- **Objetivo táctil**: recomendación mínima funcional (no diseño): botón con alto ≥ 44 px para pulgar. Si el equipo decide cero CSS en F1, el botón nativo es aceptable y este punto queda para F3.
- **Textos**: español claro y oraciones cortas, pensado también para personas mayores (idea.md: públicos de todas las edades); el detalle técnico (si lo hay) va siempre después del mensaje principal, nunca en su lugar.

## 6. Textos (literales, español)

**Introducción estática:**

> Este sitio web está en construcción. Por ahora, esta página muestra si el sistema de la Iglesia Simiente Santa está funcionando correctamente.

**Estado *Consultando* (inicial):**

> Consultando el estado del sistema…

**Durante una reconsulta (resultado anterior visible):**

> Actualizando el estado…

**Estado *Todo funcionando*:**

> **El sistema está funcionando.**
> Servidor: en marcha.
> Base de datos: conectada.
> Última consulta: [hora exacta, p. ej. "14:32:05"].

**Estado *Base de datos no conectada*:**

> **El sistema está en marcha, pero la base de datos no está conectada.**
> Servidor: en marcha.
> Base de datos: no conectada.
> Última consulta: [hora exacta].
> El sitio web está corriendo, pero no puede leer ni guardar información por ahora. Esto suele ocurrir cuando la base de datos está detenida o todavía está arrancando.

**Estado *No se pudo consultar*:**

> **No se pudo consultar el estado del sistema.**
> Servidor: sin respuesta.
> Base de datos: no se pudo comprobar.
> Última consulta: [hora exacta del intento].
> No hubo respuesta del servidor, así que no sabemos si el sistema está funcionando. Esto no significa que la base de datos esté desconectada: aquí el problema está en el propio servidor o en la conexión con él. Espere unos momentos y vuelva a consultar.

**Botón:**

> Volver a consultar el estado

## 7. Decisiones de diseño y huecos detectados en la spec

Decisiones tomadas aquí (revisables en `plan`), y huecos que la spec no define — **se reportan, no se suponen resueltos**:

1. **Auto-refresco / polling**: la spec exige que la consulta refleje el estado real (FR-003) pero no dice si la página debe refrescarse sola. Este diseño opta por lo mínimo verificable: consulta automática al cargar + botón manual. Si se quiere refresco periódico, decidirlo en `plan`.
2. **Timeout de la consulta**: la spec no cubre el caso "el backend cuelga sin responder" (ni éxito ni error de red). Este diseño trata la ausencia de respuesta a tiempo como estado *No se pudo consultar*; el valor concreto del timeout (sugerencia: 5 s) y el contrato exacto de `/healthz` (HTTP/JSON por caso) los define el arquitecto en `contracts/`.
3. **Respuesta ininterpretable** (backend responde pero el cuerpo no se entiende): la agrupo con el estado *No se pudo consultar* con el mismo texto; el contrato del arquitecto debería imposibilitar este caso.
4. **Nota "sitio en construcción"** en la intro: añadido de contenido no pedido literalmente por la spec (evita confusión en visitantes reales); decidido aquí, trivial de retirar en F3.
5. **URL del backend configurable** por variable de entorno (`endpoint`): necesario para que dev/prod apunten a hosts distintos (FR-008); la variable concreta la define el plan junto a `devops`.
