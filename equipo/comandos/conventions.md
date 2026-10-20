---
nombre: conventions
descripcion: Redacta las convenciones propias del proyecto (nombres, estructura, patrones, bibliotecas) leyendo su código, las compara con el estándar y las propone para aprobación.
argumentos: [vacío para todas las partes; el nombre de una parte para solo esa]
---
Las skills del catálogo del kit dicen el estándar de cada tecnología. Este comando escribe lo que es propio de **este** proyecto: cómo nombra las cosas, cómo se organiza, qué patrones y bibliotecas usa, y en qué se aparta del estándar. Queda en una skill del proyecto por cada parte (`.agents/skills/convenciones-de-<parte>/SKILL.md`). Aplica la skill `convenciones-proyecto`.

1. Ejecuta `make profile`. Si el proyecto no tiene perfil, detente y propón crearlo primero con el comando `bowser-profile`: las convenciones se redactan por parte, y las partes las dice el perfil.
2. Si la persona nombró una parte, trabaja solo esa. Si no, todas las que tienen roles, una a la vez.
3. Delega en el `arquitecto` la lectura del código y la redacción, pasándole la parte, su carpeta, sus carpetas de terceros (no se leen) y las skills del catálogo que el perfil le asigna (son el estándar contra el que compara). Si tu herramienta no puede lanzar subagentes, asume tú ese rol. Si la parte ya tiene convenciones, se actualizan: no se empieza de cero.
4. Donde el código hace lo mismo de dos formas, **pregunta a la persona** cuál vale para el código nuevo. No elijas tú.
5. Muestra la propuesta en lenguaje simple: lo más característico de cada parte, en qué se aparta del estándar, y aparte la **deuda de seguridad** encontrada (lo que el estándar exige y el proyecto no cumple). Esa deuda nunca se escribe como convención. Espera la aprobación explícita de la persona.
6. Con su "sí", escribe los archivos y agrega cada skill a las `skills` de su parte en `equipo/perfil.json`, sin quitar las que ya tenía.
7. Ejecuta `make profile` y `make install-kit`, para que los roles reciban la skill nueva.
8. Dile a la persona que el commit lo confirma ella, con `APPROVED_PROFILE=1 git commit ...`, porque cambió el perfil. Nunca uses esa variable por tu cuenta.

En un proyecto que todavía no tiene código no hay nada que leer: dilo, y propón ejecutar este comando cuando la primera funcionalidad esté terminada.
