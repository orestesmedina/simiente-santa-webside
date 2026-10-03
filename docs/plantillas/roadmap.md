# Roadmap: [nombre del producto]

<!--
Lo propone el orquestador a partir de docs/producto/idea.md; el humano decide orden y alcance.
Se copia a docs/producto/roadmap.md. El orquestador actualiza la columna Estado:
  pendiente → en curso (al crear la spec) → en revisión (PR abierto) → terminada (merge confirmado por un humano)
  pausada: se empezó y se dejó a propósito (anota el motivo en su estado.md)
El cambio de estado se hace en la rama de la funcionalidad y llega a main con el merge;
mientras tanto, `make estado` muestra desde main el trabajo en curso de las otras ramas.
`make estado` lee esta tabla: conserva las columnas.
-->

Objetivo del MVP: [una frase]

| # | Funcionalidad | Qué incluye | Depende de | Estado | Rama / PR |
|---|---|---|---|---|---|
| 1 | Estructura base | Backend Go con /healthz, frontend React, PostgreSQL, Docker Compose y CI | — | pendiente | |
| 2 | [Ej.: Registro de usuarios] | [Alcance en una línea] | 1 | pendiente | |
| 3 | [Ej.: Inicio de sesión] | [Alcance en una línea] | 2 | pendiente | |

## Fuera del MVP

- [Lo que se deja para después]

## Cambios al roadmap

- AAAA-MM-DD — [Qué cambió y quién lo decidió]
