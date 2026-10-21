---
nombre: ci
descripcion: Redacta o actualiza la integración continua del proyecto (su workflow de GitHub) a partir del perfil, y la propone para aprobación.
argumentos: [vacío para crear o actualizar; "ver" solo muestra qué se ejecuta hoy]
---
La integración continua es del proyecto: instalar un lenguaje o levantar una base de datos depende de su tecnología. El kit solo trae `.github/workflows/bowser.yml`, con los controles que no dependen de ella. Este comando escribe el workflow del proyecto (`.github/workflows/ci.yml`), que ejecuta en GitHub los mismos verbos que el equipo ejecuta con `make ci`. Aplica la skill `ci-proyecto`.

Si la persona escribió "ver", ejecuta `make check-ci`, explica el resultado en lenguaje simple y termina.

En cualquier otro caso:

1. Ejecuta `make profile`. Si el proyecto no tiene perfil, detente y propón crearlo primero con el comando `bowser-profile`: la integración continua ejecuta los verbos, y los verbos los dice el perfil.
2. Ejecuta `make check-ci`: dice qué verbos de qué partes ejecuta ya algún workflow y cuáles no ejecuta nadie.
3. Delega en `devops` la redacción, pasándole las dos salidas. Si tu herramienta no puede lanzar subagentes, asume tú ese rol. Si el proyecto ya tiene `ci.yml`, se actualiza: no se empieza de cero ni se pierden sus pasos propios.
4. Lo que no se pueda deducir del proyecto (qué versión de una herramienta usar, qué servicios necesitan las pruebas) **se pregunta a la persona**. No inventes versiones.
5. Muestra la propuesta en lenguaje simple: qué se revisa en cada parte, qué se instala y en qué versión, qué servicios se levantan, qué partes quedan sin revisar porque no tienen verbos, y qué secretos o variables hay que crear en GitHub. Espera la aprobación explícita de la persona.
6. Con su "sí", escribe el archivo y ejecuta `make check-ci`. Si queda algún verbo sin ejecutar, corrígelo o explica por qué se deja fuera.
7. Dile a la persona que un workflow solo se prueba de verdad en GitHub: el primer Pull Request dirá si funciona. Nunca edites `bowser.yml`: es del kit.
