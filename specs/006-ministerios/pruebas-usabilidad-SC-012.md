# Protocolo de prueba de usabilidad SC-012 (y SC-001) — F6 Ministerios

Criterio **SC-012**: «En pruebas con personas de distintas edades, incluidas con poca experiencia en internet, al menos el **90 %** identifica qué ministerio es, de qué se trata y cómo participar —unirse por el enlace cuando existe, o acercarse al encargado cuando su contacto está publicado—, sin ayuda.» Incluye el **cronómetro de los 30 s** de **SC-001** («para los ministerios con enlace, encuentra cómo unirse en menos de 30 segundos desde que abre la sección»).

**Estado: PENDIENTE** (queda abierto al cerrar la fase de validación; no es automatizable). Responsable: el equipo (o la persona que designe el humano). No bloquea el APROBADO de QA, pero **sí el cierre de la funcionalidad**, igual que en F4 (SC-010).

Este protocolo es una prueba con el producto terminado y real, no de laboratorio:

- Con el stack levantado (`make up`) y contenido **real publicable** por el equipo (no «Lorem ipsum»): 5–8 ministerios, al menos uno **con enlace de WhatsApp**, al menos uno **sin enlace y sin encargado** (SC-017), uno con encargado visible (nombre + celular o correo), uno con horario/lugar y un borrador que no debe verse.
- En el dispositivo que la persona use a diario y, si se puede, una segunda ronda en el otro extremo (computadora o teléfono).
- En español; el inglés solo si la persona quiere cambiarlo por su cuenta (no se fuerza).

## Muestra

- **≥ 6 personas**, distintas edades («adultos mucho más mayores que saben lo básico de internet» incluidos, `idea.md` §7), mixto de experiencia con internet.
- Idealmente 1–2 que nunca hayan usado el panel (solo visitantes).

## Guion de la sesión (~10 min por persona)

1. Dejar abierta la portada del sitio real. **Sin dar contexto**: «Bienvenido/a; te pido que uses el sitio como lo harías normalmente».
2. Tareas en este orden (sin decir dónde está nada):
   1. "¿Hay algún ministerio en esta iglesia? ¿Me dirías uno de los que ves?" — solo cuenta como éxito si lo identifica **por su cuenta** (por el menú, por el bloque de la portada o por el pie).
   2. "¿De qué se trata ese ministerio? ¿cómo participarías?" — leer la descripción y señalar si el horario/lugar le sirve.
   3. Con el ministerio **que tiene enlace de WhatsApp**: "Únete al grupo de ese ministerio" — aquí corre el **cronómetro de SC-001**: desde que abre la sección hasta que toca/click en el enlace. Si se frustra, detener a los 60 s y anotar «no alcanzado».
   4. Con el ministerio con encargado publicado y con el **sin enlace y sin encargado** (SC-017): "¿Cómo harías para acercarte sin el enlace?" — la ficha debe bastar.
   5. Libre: si la persona **quiere**, cambia el idioma (registrar si el contenido quedó bien sin traducción).
3. Cerrar con tres preguntas abiertas: «¿Qué te parece que está esta sección?», «¿Hay algo que te confunda o te haría desconfiar?» (privacidad de los encargados, fichas sin botón de unirse, la redirección a la portada cuando no hay publicados).

## Qué registrar (una fila por persona y tarea)

| Dato | Cómo se anota |
|---|---|
| Tarea 1: encuentra los ministerios | sí/no y **de dónde** (menú, bloque de la portada, pie) |
| Tarea 1: sabe de qué trata | sí/no — con qué elementos de la ficha (descripción, horario, lugar…) |
| Tarea 2 (SC-001): tiempo hasta tocar el enlace | **segundos** (objetivo: < 30 s); «no alcanzado» al terminar 60 s |
| Tarea 2: tocó el enlace (clic o toque único, llega al grupo) | sí/no |
| Tarea 3: sabe cómo acercarse con lo que hay en la ficha | sí/no por ministerio (con encargado visible y sin encargado) |
| Ayudas dadas por quien modera | nº de pistas (0 = éxito puro) |
| Sorpresas, confusión o desconfianza | textual (especial: si denuncia ver datos de borradores o encargados no publicados → **defecto**, avisar QA al instante) |

**Criterio de aprobación**: en la tarea de identificación, al menos el **90 %** de personas lo logra sin ayuda (SC-012); y como mínimo 5 de cada 6 logran unirse/acercarse sin ayuda; el promedio de tiempo hasta el clic **< 30 s** (SC-001).

## Cómo tratar los hallazgos

1. Lo que no sea un fallo de la spec (p. ej., «no encontré el menú») va a la **lista de hallazgos** de la revisión de QA con severidad y, si aplica, ajuste menor de UX (sin cambiar la spec).
2. Si el hallazgo **toca el comportamiento descrito en la spec** (aparece un borrador, un encargado oculto o un enlace no seguro → privacidad/seguridad), se convierte en **defecto bloqueante** con pasos de reproducción y se escala al orquestador.

---

*Preparado por `qa-tester`, F6, 2026-10-13. Formato del precedente de F4: protocolo de SC-010 de `specs/004-eventos-actividades/`.*
