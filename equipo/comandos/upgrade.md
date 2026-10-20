---
nombre: upgrade
descripcion: Revisa qué le falta al proyecto para aprovechar la versión 2.0 del kit, arregla lo mecánico y acompaña el resto paso a paso (make upgrade).
argumentos: [vacío para revisar y acompañar; "ver" solo muestra lo pendiente]
---
Un proyecto que viene de la 1.x del kit queda con pendientes que la actualización no resuelve sola: el perfil, su integración continua, las reglas heredadas de la constitución, nombres de comandos anteriores, scripts que pasaron a ser suyos. `make upgrade` los junta en una lista, separando lo que resuelve un agente, lo que resuelve una persona y lo que es mecánico.

Primero ejecuta `make upgrade`: desde la herramienta solo muestra la lista, no aplica nada. Explica cada pendiente en lenguaje simple, del que más bloquea al que es solo un aviso.

Si la persona escribió "ver", termina ahí.

En cualquier otro caso, sigue en este orden y espera el "sí" de la persona antes de cada paso:

1. **Arreglos mecánicos** (nombres anteriores de comandos y opciones, comandos retirados con reemplazo, scripts que pasan de `scripts/` a `tools/`). Muéstrale los cambios que la lista propone y, con su "sí", aplícalos con `make upgrade YES=1`. Antes, ejecuta `git status --short`: si hay cambios sin commit, avisa y pregunta si se sigue.
2. **Lo que resuelve un agente**, uno a la vez y en este orden: el perfil (comando `bowser-profile`), la integración continua (comando `bowser-ci`) y, si la lista lo recomienda, las convenciones (comando `bowser-conventions`). Cada uno tiene su propio paso de aprobación: síguelo.
3. **Lo que resuelve una persona** (las reglas de `equipo/reglas.md`, las reglas heredadas de `.gitignore`): explícale qué revisar y por qué, y ofrécete a proponer el texto. No edites `equipo/reglas.md`: lo aplica ella, y su commit pide `APPROVED_CONSTITUTION=1`.
4. Al final, ejecuta otra vez `make upgrade` y dile qué quedó pendiente.

No hagas commits por tu cuenta, ni uses `APPROVED_PROFILE=1` o `APPROVED_CONSTITUTION=1`: muestra qué cambió (`git status --short`) y deja que la persona los confirme. Si cambió el perfil, recuérdale que su commit pide `APPROVED_PROFILE=1`.
