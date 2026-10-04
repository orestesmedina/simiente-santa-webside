# Entrega F1 — Estructura base

**Fecha:** 2026-10-03 · **Versión:** 0.1.0 · **Rama / PR:** `001-estructura-base` · [PR #1](https://github.com/orestesmedina/simiente-santa-webside/pull/1)

Nota de entrega en lenguaje sencillo. Si querés la parte técnica, está en el [`README`](../../README.md) y en el [`CHANGELOG`](../../CHANGELOG.md).

## Qué se entrega

Es la **primera pieza del sitio**: todavía no hay contenido visible para la iglesia. Lo que se entrega es todo el cimiento sobre el que se construye lo demás, para que a partir de ahora cada entrega sea rápida y segura.

- **El sitio completo arranca con un solo comando**: base de datos, servidor interno y página web, todo junto con Docker.
- **Una página de estado** en <http://localhost:5173> que comprueba y muestra en castellano si el servidor y la base de datos están funcionando.
- **Un verificador público** en <http://localhost:8080/healthz>: contesta `ok` cuando todo está bien y avisa si la base de datos no responde.
- **Base de datos PostgreSQL 16** con sus migraciones guardadas y versionadas (hoy sin tablas de contenido: eso llega con las siguientes entregas).
- **Revisiones automáticas en cada cambio** (CI): formato, pruebas, vulnerabilidades y búsqueda de datos sensibles. Seis comprobaciones, todas en verde.
- **Una receta probada** (`docs/tecnico/arquitectura.md`, §8) para agregar las áreas de negocio de la iglesia —eventos, grupos, ministerios, donaciones—. Se aplicó de principio a fin antes de cerrar esta entrega y la última corrida añadió un área nueva sin decisiones técnicas y sin tocar lo ya existente.

## Cómo levantarlo

1. Instalar [Docker](https://www.docker.com/) (es el único requisito).
2. Desde la carpeta del proyecto:

   ```bash
   make instalar-hooks   # una sola vez, activa las revisiones de git
   make up               # enciende base de datos + servidor + página web
   ```

3. Abrir <http://localhost:5173>.

Para detener: `make down` (los datos se conservan; volver a levantar es `make up`).

## Cómo verificar que todo está bien

- **Página web:** abrir <http://localhost:5173> y ver el mensaje «El sistema está funcionando.», con el detalle del servidor y de la base de datos.
- **Verificador directo:** abrir <http://localhost:8080/healthz>. Si todo está bien responde `200` con `{"status":"ok","database":"connected"}`; si la base de datos no responde, avisa con un error claro (`database_unavailable`) en lugar de fallar en silencio.

## Qué sigue

**F2 — Acceso y gestión de usuarios**: inicio de sesión en un panel de administración, gestión de usuarios y roles con permisos. Después siguen portada, eventos, grupos, ministerios y donaciones, en el orden del [roadmap](../producto/roadmap.md).

## Límites conocidos de esta entrega

- **No hay contenido de la iglesia todavía**: no se pueden crear ni ver eventos, grupos, ministerios ni donaciones (llegan en F3–F7).
- **No hay inicio de sesión** (llega en F2).
- **No hay despliegue a internet**: todo funciona en local, en tu equipo.

## Enlaces

- [`README.md`](../../README.md) — instalación, comandos y variables de entorno.
- [`CHANGELOG.md`](../../CHANGELOG.md) — qué cambió en la versión 0.1.0.
- [`docs/producto/roadmap.md`](../producto/roadmap.md) — las funcionalidades F1–F9 y su estado.
