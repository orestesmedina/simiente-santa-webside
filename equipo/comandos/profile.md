---
nombre: profile
descripcion: Crea o actualiza el perfil del proyecto (partes, tecnologías y comandos) mirando el código, y lo propone para aprobación.
argumentos: [vacío para crear o actualizar; "ver" solo lo muestra]
---
El perfil (`equipo/perfil.json`) le dice al kit qué partes tiene el proyecto, en qué carpeta está cada una, qué comando ejecuta cada verbo (formato, revisar, probar, cobertura, auditar, generar) y qué programas necesita una máquina para trabajar en él. Aplica la skill `perfil-proyecto`.

Si la persona escribió "ver", ejecuta `make profile`, explica el resultado en lenguaje simple y termina.

En cualquier otro caso:

1. Ejecuta `make profile DETECT=1` (solo lee: muestra en JSON lo que encontró en el proyecto) y, si ya hay perfil, `make profile`.
2. Delega en el `arquitecto` la redacción del perfil, pasándole esa salida. Si tu herramienta no puede lanzar subagentes, asume tú ese rol.
3. Lo que no se pueda deducir del código (cómo se prueba, qué versión se usa, si una carpeta es de terceros) **se pregunta a la persona**. Si el proyecto no tiene algo todavía (por ejemplo, pruebas), el verbo queda en `null`: no inventes un comando.
4. Muestra la propuesta en lenguaje simple: qué partes encontraste, con qué tecnología, qué agente trabaja cada una, qué comando corre cada verbo, cuáles quedan sin definir y qué programas debe tener instalados cada máquina (`requisitos`, lo que revisa `make doctor`). Espera la aprobación explícita de la persona.
5. Con su "sí", escribe `equipo/perfil.json` y ejecuta `make profile`. Si marca errores, corrígelos y vuelve a mostrar el resultado.
6. Ejecuta `make install-kit`: trae del catálogo del kit las skills que el perfil nombra y retira las que ya no nombra, y regenera los roles para que cada agente sepa qué partes, skills y comandos son suyos. Sin ese paso, el commit del perfil se rechaza. Si avisa que una skill no existe o que está solo redactada, díselo a la persona.
7. Propón a la persona las reglas de `.gitignore` de las tecnologías del perfil (las da cada skill en "Lo que no va en git") y, con su "sí", escríbelas fuera del bloque del kit. Si el proyecto no tiene entorno local (`make up` lo avisa), dilo como pendiente para `devops`.
8. Si el proyecto ya tiene código, dile a la persona que el paso siguiente es el comando `bowser-conventions`: las skills del catálogo traen el estándar de cada tecnología, y ese comando redacta lo propio de este proyecto (nombres, estructura, patrones).
9. Dile a la persona que el commit del perfil lo confirma ella, con `APPROVED_PROFILE=1 git commit ...`. Nunca uses esa variable por tu cuenta: los comandos del perfil se ejecutan en su máquina y en la integración continua.
