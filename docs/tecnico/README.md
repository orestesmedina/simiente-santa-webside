# Documentación técnica

Guía de **cómo** se construye el proyecto (el **qué** vive en `specs/` y en `docs/producto/`).

- [arquitectura.md](./arquitectura.md) — principios y reglas de dependencia, árbol de carpetas del
  backend, la interfaz `httpserver.Registrar` que neutraliza el router, un ejemplo vertical completo
  del dominio `contacto`, el flujo de una petición, las pruebas por capa y la receta de 10 pasos
  para agregar un área de negocio.
- [decisiones.md](./decisiones.md) — registro de decisiones (ADR corto D-A1…D-A9): monorepo y
  contrato OpenAPI, plataforma interna, sqlc, router, Go 1.27, config/logs/DI, sesiones de F2
  *(propuesta pendiente)*, dependencias mínimas y diferidos, con las preguntas abiertas al final.

Para el **proceso** del equipo (roles, fases, aprobaciones) está `docs/GUIA-INICIO.md` (guía del
kit) y `equipo/orquestador.md`; para el rumbo del producto, `docs/producto/roadmap.md`. Las specs
aprobadas en `specs/<funcionalidad>/` mandan sobre cualquier documento de esta carpeta cuando
afecten al comportamiento (constitución §I).
