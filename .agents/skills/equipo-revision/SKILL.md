---
name: equipo-revision
description: Validación de los cambios actuales con QA, revisión de código y seguridad en paralelo, con un veredicto consolidado. Usar antes de abrir o aprobar un Pull Request, o cuando el usuario pide revisar cambios.
---
# Flujo: validar cambios

Alcance: lo que indique el usuario, o `git diff main...HEAD` si no indica nada.

Lanza **en paralelo** (o en secuencia si tu herramienta no permite paralelo):
1. `qa-tester` — cobertura de los criterios de aceptación de la spec correspondiente en `specs/`.
2. `revisor-codigo` — calidad, apego al plan y a la constitución.
3. `seguridad` — vulnerabilidades, secretos y dependencias.

Consolida los tres reportes en uno:
- **Veredicto global:** APROBADO solo si los tres aprueban.
- **Bloqueantes** (deben corregirse antes del merge), con archivo, línea y responsable (`dev-backend` o `dev-frontend`).
- **Mejoras sugeridas** (no bloquean).

Guarda el reporte en `specs/<feature>/revision-<fecha>.md`. No modifiques código en este flujo.

Si existe `specs/<feature>/estado.md`, el orquestador actualiza ahí el `Ciclo de corrección` y los **Hallazgos abiertos** (solo los bloqueantes, con enlace al reporte).
