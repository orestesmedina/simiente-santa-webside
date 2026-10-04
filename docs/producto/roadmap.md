# Roadmap — Sitio web de la Iglesia Simiente Santa

**Fuente:** `docs/producto/idea.md` · **Estado:** v1.0 — aprobado por el cliente el 2026-09-29 · **Fecha:** 2026-09-28

Este documento divide el producto en funcionalidades independientes, en orden de dependencia. No es una especificación: cada funcionalidad tendrá su propia spec en `specs/` cuando se apruebe.

## 1. Problema principal

La iglesia no tiene un canal unificado de difusión: hoy solo se comparte un afiche de vez en cuando en WhatsApp o Facebook y "la gran mayoría del tiempo se olvida". Resultado: los miembros y las personas nuevas no se enteran de las actividades, eventos y grupos de conexión, y las personas nuevas no asisten porque no saben que existen (`idea.md` §1 y §3).

## 2. El MVP más pequeño que lo resuelve

Un **sitio público** moderno, responsivo y navegable por personas de todas las edades, en español e inglés, donde cualquier persona descubre **los próximos eventos/actividades**, **los grupos de conexión**, **los ministerios** y **cómo donar**, además de la información general de la iglesia (quiénes somos, horario de servicios, canales de WhatsApp y redes sociales). Y un **panel de administración** con inicio de sesión, roles con permisos por módulo y gestión de usuarios, para que el equipo de la iglesia mantenga ese contenido dinámico sin depender de un desarrollador.

Con eso se cumple lo esencial de los criterios de éxito (`idea.md` §4): canal unificado para los miembros, puerta de entrada para personas nuevas y contenido administrable.

**Fuera del MVP (decisión 2):** las noticias/galería de eventos recientes y los medios (prédicas y podcasts) quedan para la **fase 2**; ministerios y donaciones entran al MVP (F6 y F7) porque el cliente así lo decidió.

## 3. Funcionalidades comprometidas (F1–F9)

La columna **Fase** indica el orden de construcción: `1` = MVP núcleo, `2` = después del MVP. La columna **Estado** la mantiene el orquestador (`pendiente → en curso → en revisión → terminada`; `make estado` lee esta tabla) y **Rama / PR** enlaza la funcionalidad en curso.

| # | Funcionalidad | Qué hace para el usuario | Fase | Depende de | Estado | Rama / PR |
|---|---|---|---|---|---|---|
| F1 | Estructura base | Para el visitante aún nada visible: monta el proyecto (Go + React + PostgreSQL) con Docker Compose y CI, y una primera página muestra que el backend y la base de datos están vivos. Es el terreno sobre el que se construye todo lo demás. | 1 · MVP núcleo | — | terminada | [#1](https://github.com/orestesmedina/simiente-santa-webside/pull/1) fusionado (2026-10-03) |
| F2 | Acceso y gestión de usuarios | El equipo de la iglesia inicia sesión en un panel de administración; los administradores crean usuarios, los activan o desactivan, y crean roles con los permisos por módulo que necesiten (p. ej., un rol de contenido, un rol de ministerios). | 1 · MVP núcleo | F1 | en curso | `002-acceso-gestion-usuarios` |
| F3 | Portada e información general | Cualquier visitante ve una portada con la identidad de la iglesia, quiénes somos, horario de servicios, canales de WhatsApp y redes sociales, en español e inglés; el equipo edita esa información desde el panel. | 1 · MVP núcleo | F2 | pendiente | |
| F4 | Eventos y actividades | El visitante descubre lo que viene: **eventos** periódicos de la iglesia (noche de aposentos, vigilia) y **actividades** puntuales (repartir comida, predicar en tal lugar), con fecha, lugar y detalle; el equipo administra ambos por separado. Es el corazón del problema: que nadie se quede sin enterarse. | 1 · MVP núcleo | F2, F3 | pendiente | |
| F5 | Grupos de conexión | El visitante conoce el catálogo de talleres y clases (maquillaje, fotografía, inglés, asados…) con horario, lugar y el contacto del encargado para inscribirse; el equipo administra el catálogo desde el panel. | 1 · MVP núcleo | F2, F3 | pendiente | |
| F6 | Ministerios | El visitante conoce cada ministerio, su encargado y el enlace al grupo de WhatsApp para unirse; el equipo administra el listado desde el panel. | 1 · MVP núcleo | F2, F3 | pendiente | |
| F7 | Donaciones | El visitante ve cómo apoyar a la iglesia: cuentas IBAN y SINPE Móvil, y en qué se usan las donaciones; el equipo administra esa información desde el panel. | 1 · MVP núcleo | F2, F3 | pendiente | |
| F8 | Noticias y galería | El visitante ve qué ha pasado últimamente: noticias de eventos recientes con imágenes y videos, administradas por el equipo. | 2 · Después del MVP | F2, F3 | pendiente | |
| F9 | Medios: prédicas y podcasts | El visitante escucha o ve las grabaciones publicadas en YouTube y Spotify desde el propio sitio, sin instalar esas aplicaciones; el equipo administra los episodios. | 2 · Después del MVP | F2, F3 | pendiente | |

F4, F5, F6 y F7 son independientes entre sí y pueden trabajarse en paralelo. F8 y F9 también son paralelas entre sí.

Aplica a todo el sitio público (fases 1 y 2): es bilingüe, español e inglés, con selección de idioma desde la portada, y cada contenido puede ingresarse en ambos idiomas o solo en español (decisiones 6 y 8); usuarios y contenido tienen estados (decisión 4): usuarios activo/inactivo en F2, contenido borrador/publicado en las funcionalidades de contenido.

## 4. Fase 2 — después del MVP

Completa el alcance del MVP original de `idea.md` §5 con lo que quedó fuera por la decisión 2. F8 y F9 dependen de que existan el panel y el sitio público (F2, F3) y son paralelas entre sí (ver tabla de §3).

## 5. Fase 3 — ideas futuras (ninguna confirmada)

Vienen de las preguntas abiertas de `idea.md` §8 y de la sección de fuera de alcance; ninguna entra al roadmap comprometido hasta que el cliente la apruebe.

| Candidata | Qué haría | Depende de |
|---|---|---|
| Recuperación de contraseña por auto-servicio | El usuario del panel restablece su propia contraseña con un enlace enviado a su correo, sin depender de un administrador. En el MVP la contraseña la restablece un administrador (ver decisión 9); esto llegaría cuando el proyecto madure (requiere servicio de correo). | F2 |
| Formulario de inscripción a grupos | El visitante se inscribe a un grupo de conexión llenando un formulario; en el MVP solo se muestra la información y el contacto del encargado (decisión 1). | F5 |
| Sincronización de episodios | Publicar en YouTube/Spotify agrega el episodio al sitio automáticamente. | F8 |
| Publicación y banners para redes | Generar imágenes de eventos y publicar automáticamente en Instagram/WhatsApp/Facebook (los banners están fuera de alcance en `idea.md` §6). | F4 |
| Cursos virtuales | Plataforma donde los miembros se registran y llevan cursos impartidos por la iglesia (también listada fuera de alcance en §6). | F5, y una exploración propia |

## 6. Decisiones tomadas

1. **Inscripción a grupos de conexión (2026-09-28).** El MVP solo muestra la información de cada grupo y el contacto de su encargado; el formulario de inscripción queda para la fase 3.
2. **Alcance del MVP (2026-09-28).** Ministerios y donaciones entran al MVP (F6 y F7); noticias/galería y medios (prédicas y podcasts) quedan para la fase 2.
3. **Eventos y actividades (2026-09-28).** Son conceptos distintos, modelados por separado dentro de F4: **eventos** = encuentros periódicos de la iglesia (noche de aposentos, vigilia); **actividades** = acciones puntuales (repartir comida, predicar en tal lugar).
4. **Estados (2026-09-28).** Ambos: usuarios activo/inactivo y contenido borrador/publicado.
5. **Permisos y roles (2026-09-28).** Permisos por módulo (eventos, actividades, ministerios, grupos…) que se agrupan en roles; no hay catálogo fijo: el administrador crea roles nuevos y les asigna permisos de forma independiente (p. ej., un rol de contenido con eventos y actividades, sin ministerios). La spec de F2 define cómo queda garantizado el arranque con un administrador inicial.
6. **Idioma (2026-09-28).** El sitio es bilingüe, español e inglés, para que personas de otros países también lo vean.
7. **Donaciones (2026-09-28).** La información de donaciones (IBAN/SINPE y su uso) se administra desde el panel; F7 depende de F2 y F3.
8. **Bilingüismo del contenido (2026-09-28).** Todo el contenido puede ingresarse en ambos idiomas o solo en español: el español es el idioma base y el inglés es opcional por contenido (la interfaz, en cambio, siempre es bilingüe — decisión 6).
9. **Gestión de usuarios y roles de F2 (2026-10-04).** Se refinan las decisiones 4 y 5 con lo aprobado en la spec de F2: las cuentas **no se eliminan**, solo se activan o desactivan (los datos se conservan); cada cuenta tiene **un solo rol**; los roles se editan y se eliminan **solo si no están en uso**; el administrador inicial se crea con una **acción de inicialización única** (no repetible); y la recuperación de contraseña por auto-servicio con correo queda para el backlog (en el MVP la contraseña la restablece un administrador).

## 7. Preguntas abiertas

Ninguna pendiente: todas las que fueron surgiendo se resolvieron y están registradas en Decisiones tomadas. Las dudas nuevas que aparezcan se agregan aquí antes de suponer nada.

## 8. Criterios de éxito del MVP (de `idea.md` §4)

- Los miembros tienen un canal unificado donde ver y conocer las actividades y grupos de la iglesia.
- Las personas nuevas conocen las actividades y tienen una puerta para acercarse a la congregación.
- Eventos, grupos y demás contenido son dinámicos: administrables por el equipo, no estáticos.

## 9. Próximos pasos

1. Aprobado por el cliente el 2026-09-29 ✅
2. Empezar F1 (estructura base) con el flujo de la skill `equipo-feature`, y seguir el roadmap una funcionalidad a la vez, integrando cada una antes de empezar la siguiente.
