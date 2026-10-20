---
name: react-frontend
description: Convenciones para escribir interfaces en React con TypeScript según la documentación oficial (react.dev: las Reglas de React, cómo estructurar el estado y cuándo usar efectos). Usar siempre que se cree o modifique un componente, un hook o cualquier archivo .tsx o .jsx.
metadata:
  madurez: redactada
---
# React — convenciones

Fuentes: react.dev (Rules of React, Thinking in React, Choosing the State Structure, You Might Not Need an Effect, Using TypeScript) y, para la accesibilidad, las pautas WCAG 2.2 del W3C. Revisado el 2026-10-09. Ante una duda que esta página no resuelva, se consulta ahí.

Qué herramienta de compilación, de rutas, de datos o de estilos usa el proyecto no lo decide React ni esta página: es del proyecto.

## Versión y punto de partida
- La última versión mayor estable de React (react.dev/versions), con TypeScript en modo `strict`.
- react.dev recomienda empezar con un framework o con una herramienta de compilación como Vite. Create React App está retirado.
- Componentes de función y hooks. No se escriben componentes de clase.

## Componentes
- Un componente es una función con nombre en `PascalCase` que devuelve JSX. Se declara en el nivel superior del archivo: **nunca dentro de otro componente**, porque se crearía de nuevo en cada render y perdería su estado.
- Pequeños y con una sola responsabilidad. Cuando uno crece, se parte.
- Las props se tipan con un `interface` o un `type`; nada de `any`. Los hijos se tipan como `React.ReactNode`.
- Cada elemento de una lista lleva una `key` estable que viene de los datos (un identificador). No el índice del arreglo ni un valor aleatorio.
- Se muestra u oculta con condiciones en el JSX; no se manipula el documento a mano (`document.querySelector`). Para llegar a un elemento, `useRef`.

## Las reglas de React
- **Los componentes y los hooks son puros**: con las mismas props, estado y contexto devuelven lo mismo. Durante el render no se hacen llamadas al servidor, no se modifican variables de fuera ni se toca el documento.
- **Las props y el estado no se modifican**: se crea un valor nuevo (`setItems([...items, nuevo])`, no `items.push(nuevo)`). Tampoco se modifica un valor después de pasarlo a un hook o al JSX.
- Un componente se usa en JSX (`<Perfil />`); no se llama como función (`Perfil()`).
- **Los hooks se llaman solo en el nivel superior** de un componente o de otro hook: nunca dentro de condiciones, ciclos, funciones anidadas ni después de un `return`.
- Un hook propio empieza con `use` y solo se llama desde componentes u otros hooks.
- Estas reglas las comprueba `eslint-plugin-react-hooks`, el complemento oficial: sus avisos no se silencian.

## Estado
- El mínimo necesario. **Lo que se puede calcular de las props o de otro estado no es estado**: se calcula durante el render.
- Sin duplicar: un dato vive en un solo lugar. No se copia una prop a un estado.
- El estado vive en el componente más cercano que lo necesita; si lo necesitan dos, se sube a su ancestro común.
- Varios valores que cambian juntos van en un solo estado. Cuando la lógica de actualización crece, `useReducer`. Para lo que muchos componentes lejanos necesitan, contexto.
- Para reiniciar el estado de un componente cuando cambia de qué se trata, se le cambia la `key`.

## Efectos
- Un efecto sirve para **sincronizar con algo de fuera de React** (el navegador, una suscripción, una biblioteca que no es de React). Si no hay nada externo, no hace falta un efecto.
- **No** se usa un efecto para transformar datos para mostrarlos (se calcula en el render) ni para reaccionar a una acción de la persona (eso va en el manejador del evento).
- Todo efecto que se suscribe o inicia algo devuelve su función de limpieza.
- La lista de dependencias incluye todo valor reactivo que el efecto usa. No se quita una para "arreglar" un ciclo: se arregla el efecto.
- Pedir datos dentro de un efecto, escrito a mano, tiene problemas conocidos (respuestas que llegan fuera de orden, sin caché): react.dev recomienda usar el mecanismo de datos del framework o una biblioteca de caché. Si aun así se hace a mano, la limpieza descarta la respuesta vieja.

## Interfaz y accesibilidad
- Cada vista con datos del servidor contempla y muestra cuatro estados: cargando, sin resultados, error y éxito.
- Elementos HTML por lo que significan (`button`, `a`, `nav`, `main`, `label`). Un `div` con `onClick` no es un botón.
- Cada campo con su `label` (`htmlFor` + `id`; el `id`, con `useId`). Toda imagen con `alt`.
- Todo se puede usar con el teclado y el foco se ve. Contraste de al menos 4.5:1 en el texto.

## Seguridad
- JSX escapa lo que se muestra entre llaves. **`dangerouslySetInnerHTML` no se usa** con datos; si el plan lo exige, el HTML se limpia antes con una biblioteca hecha para eso.
- Una dirección que viene de datos se comprueba antes de ponerla en `href` o `src`: solo `http:`, `https:` o rutas del propio sitio.
- Ninguna clave ni token en el código de la interfaz ni en `localStorage`: todo lo que llega al navegador es público. La sesión va en una cookie `HttpOnly` que maneja el servidor.
- Lo que la interfaz oculta o desactiva no protege nada: el permiso se comprueba en el servidor.

## Pruebas
- Cada archivo de prueba va junto a lo que prueba: `Perfil.tsx` y `Perfil.test.tsx`.
- Se prueba lo que la persona ve y hace (el texto, el rol de un elemento, un clic), no los detalles internos del componente. React no fija una herramienta: lo habitual es Testing Library sobre el ejecutor de pruebas del proyecto.
- La lógica que no depende de React (cálculos, formato, validaciones) va en funciones puras, que se prueban solas.

## Comandos
Los pone el proyecto en su `package.json` y en su perfil. Lo que no puede faltar: la comprobación de tipos (`tsc --noEmit`) y ESLint con `eslint-plugin-react-hooks`.

## Lo que no va en git
- `node_modules/`, la carpeta que produce la construcción (según la herramienta: `dist/` o `build/`) y los informes de pruebas y de cobertura (`coverage/`).
- `package.json` y el archivo de bloqueo de versiones (`package-lock.json` o el del gestor que use el proyecto) sí van en git.

## Si el proyecto tiene sus propias convenciones
- Esta página es el estándar. Si el proyecto tiene una skill de convenciones propias, **en nombres, estructura, patrones y bibliotecas manda esa**, para que el código nuevo se parezca al que ya hay. Lo que ella no diga se hace como dice esta página.
- **Seguridad** y **Las reglas de React** no admiten excepción. Si cumplirlas exige cambiar algo que tu tarea no cubre, detente y avisa.
- No reformatees ni reorganices lo que tu tarea no toca. Lo que encuentres fuera del estándar va en la entrega, como deuda.
