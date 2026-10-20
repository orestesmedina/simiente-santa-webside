---
nombre: status
descripcion: Muestra por dónde va el proyecto y cuál es el próximo paso (make status).
argumentos: [todo]
---
Ejecuta `make status` en la raíz del proyecto. Si la persona pidió `todo`, usa `make status ALL=1` para incluir las funcionalidades ya terminadas.

Este comando solo lee: no cambies ningún archivo.

Responde en 2 a 4 líneas, en lenguaje simple:

1. En qué funcionalidad y fase estamos, y cuántas tareas van hechas.
2. Qué está aprobado y si hay hallazgos abiertos.
3. El próximo paso.

Si la salida trae avisos (estado desactualizado, cambios sin commit, aprobaciones sin registrar), menciónalos en una línea cada uno. Si el comando falla, muestra el mensaje exacto y sugiere `make doctor`.
