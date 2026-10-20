---
nombre: news
descripcion: Muestra qué cambió en el kit, versión por versión (make news).
argumentos: [versión desde la que mostrar, ej. 1.9.0]
---
Ejecuta `make news` en la raíz del proyecto. Si la persona indicó una versión, usa `make news SINCE=<versión>`.

Este comando solo lee. Resume cada versión en una o dos líneas, en lenguaje simple, y destaca aparte los pasos de **Al actualizar**, que son los que alguien tiene que hacer a mano.
