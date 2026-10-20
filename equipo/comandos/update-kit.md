---
nombre: update-kit
descripcion: Trae la última versión del kit, la instala y explica qué cambió (make update-kit).
---
Este comando cambia archivos del proyecto. Antes de empezar, ejecuta `git status --short`: si hay cambios sin commit, avisa y pregunta si se sigue.

Ejecuta `make update-kit` en la raíz del proyecto.

- **Si se detiene porque un archivo del kit fue modificado en el proyecto,** no fuerces nada. Explica las tres salidas y espera la decisión de la persona: llevar el cambio al kit, excluir el archivo (`kit.excluir` en `equipo/config.json`) o pisarlo con `make install-kit FORCE=1`, que guarda un respaldo en `.kit-respaldo/`.
- **Si termina bien,** resume las novedades en lenguaje simple y lista aparte, sin omitir ninguno, los pasos de **Al actualizar** que haya que hacer a mano.
- **Si avisa que el kit recomienda otros modelos,** dilo y menciona `make update-models`. Aplicarlos es decisión de la persona.

No hagas el commit por tu cuenta: muestra qué cambió (`git status --short`) y pregunta. Recuerda que el commit debe incluir `.kit-manifest.json`. Como los agentes y los comandos pueden haber cambiado, recomienda abrir una sesión nueva después.
