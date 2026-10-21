---
description: "Muestra cuánto ha costado en IA la tarea actual, por agente y modelo (make costs)."
---
<!-- GENERADO por scripts/sincronizar.py desde equipo/comandos/costs.md. No editar: cambia la fuente y ejecuta `make sync`. -->

Ejecuta `make costs` en la raíz del proyecto. Registra el consumo nuevo de la funcionalidad de la rama actual y muestra su costo.

- Si la persona pidió `todo`: `make costs ALL=1` (resumen de todo el proyecto; no registra nada).
- Si pidió `hoy`: `make costs PRICES=today` (cuánto costaría a precios actuales; no guarda nada).

**Nunca ejecutes `make costs CLOSE=1` desde este comando.** El costo se cierra solo después de que una persona aprueba el Pull Request.

Responde con el total, los dos o tres agentes o modelos que más pesan y, si lo hay, cualquier aviso de la salida. Aclara que es el costo equivalente a precio de API: con una suscripción, el gasto real es el que muestra la consola del proveedor.

Argumentos que escribió la persona junto al comando ([todo | hoy]; puede venir vacío): $ARGUMENTS
