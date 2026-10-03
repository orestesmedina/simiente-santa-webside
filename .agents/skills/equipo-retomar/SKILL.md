---
name: equipo-retomar
description: Reconstruye dónde quedó el trabajo (qué se hizo, en qué fase va, qué falta y cuál es el próximo paso) a partir del roadmap, estado.md, los artefactos de Spec Kit y git. Usar al iniciar cada sesión y cuando el usuario pregunta por dónde iban, qué falta o quiere retomar.
---
# Flujo: retomar el trabajo

Objetivo: que el usuario sepa en segundos dónde quedó todo, sin depender de la memoria de ninguna sesión anterior.
Este flujo **solo lee**; no cambia código. La única escritura permitida es corregir `estado.md` o el roadmap (paso 4).

## Versión corta (al iniciar cada sesión)

Ejecuta `make estado` y abre la conversación con 2 a 4 líneas:

```
Quedamos en: 003-registro-usuarios · Fase 6/9 (implementar) · 7/12 tareas
Aprobados: spec y plan · Hallazgos abiertos: 0
Próximo paso: T008 [frontend] formulario de registro → dev-frontend. ¿Sigo?
```

Si `make estado` muestra avisos (estado desactualizado, cambios sin commit, aprobaciones sin registrar), menciónalos en una línea. Si no hay trabajo en curso, dilo y sugiere la siguiente funcionalidad `pendiente` del roadmap.

## Versión completa (cuando el usuario pide retomar o algo no cuadra)

1. **Panorama del proyecto.** Lee `docs/producto/roadmap.md`: terminadas, en curso, en revisión, pausadas, y la siguiente pendiente.
2. **Funcionalidad actual.** Rama actual (`git branch --show-current`). Si es `main`, `make estado` lista en "Trabajo en otras ramas" las funcionalidades en curso (el roadmap de main solo se actualiza con cada merge): pregunta en cuál seguir y cámbiate a esa rama (`git checkout <rama>` y `git submodule update`).
3. **Contrasta el estado con la evidencia.** Lee `specs/<rama>/estado.md` y compáralo con:
   - Qué artefactos existen: `spec.md` (fase 1), `plan.md` (3), `tasks.md` (4), reportes `revision-*.md` (7).
   - Tareas `[X]` y pendientes en `tasks.md`.
   - `git log` de la rama (últimos commits y fechas) y `git status` (cambios sin commit).
   - Bloqueantes del último `revision-*.md`.
   - PR abierto, si puedes consultarlo (`gh pr list --head <rama>`).
4. **Si no coinciden**, mandan los archivos y git: corrige `estado.md` y explícale al usuario qué ajustaste.
   **Excepción: las aprobaciones.** No se deducen de que exista un archivo. Si `estado.md` no registra una aprobación (o no existe `estado.md`), pregunta: "¿Apruebas la spec / el plan?" antes de avanzar a una fase que la requiera.
5. **Si no existe `estado.md`** (funcionalidad empezada antes de usar este flujo), créalo desde `docs/plantillas/estado.md` con lo reconstruido, marca como `pendiente de confirmar` las aprobaciones que no consten y anota en la bitácora "Estado reconstruido el <fecha>".
6. **Cambios sin commit.** Si `git status` muestra cambios, no los descartes ni los commitees por tu cuenta: muéstralos y pregunta qué hacer.
7. **Informe al usuario** (máximo 10 líneas):
   - **Hecho:** fases completadas y aprobaciones (con fecha).
   - **En curso:** fase actual, tareas hechas/total, ciclo de corrección.
   - **Falta:** tareas pendientes o fases restantes, hallazgos abiertos, decisiones pendientes.
   - **Próximo paso** concreto, y la pregunta "¿Sigo?".
8. Cuando el usuario confirme, continúa con el flujo que corresponda (`equipo-feature`, `equipo-bug`) **desde esa fase**, sin repetir las ya cerradas.
