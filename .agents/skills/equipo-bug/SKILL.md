---
name: equipo-bug
description: Flujo para corregir un bug con diagnóstico de causa raíz, prueba que lo reproduce, arreglo mínimo y validación del equipo. Usar cuando el usuario reporta un error o comportamiento incorrecto.
---
# Flujo: corregir un bug

El bug es el que describió el usuario en su mensaje.

1. **Diagnóstico:** investiga la causa raíz leyendo el código y los logs. Explica la causa antes de tocar nada.
2. **Reproducción:** pide a `qa-tester` una prueba automatizada que falle por este bug.
3. **Arreglo:** delega en `dev-backend` o `dev-frontend` según la capa. Cambio mínimo que haga pasar la prueba sin romper otras.
4. **Validación:** aplica la skill `equipo-revision` sobre el cambio.
5. **Registro:** `documentador` agrega la entrada en CHANGELOG (sección Corregido).

Si el bug revela que la spec estaba mal o incompleta, detente y propón el cambio a la spec antes de arreglar el código.

Alternativa: si el proyecto tiene instalada la extensión de bugs de Spec Kit (`specify extension add bug`), puedes usar sus fases `bug-assess` → `bug-fix` → `bug-test` para los pasos 1 a 3.
