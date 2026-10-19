# Protocolo de prueba de usabilidad SC-012 (y SC-001) — F9 Medios: prédicas y podcasts

Criterio **SC-012**: «En pruebas con personas de distintas edades, incluidas con poca experiencia en internet, al menos el **90 %** encuentra una prédica o un podcast y **accede a la grabación** (el enlace lo lleva a YouTube o a Spotify) sin ayuda, y al menos el **90 %** identifica **que el sitio lo saca hacia YouTube o Spotify, en una pestaña nueva,** antes de activar el enlace.» Incluye el **cronómetro de los 30 s** de **SC-001** («llega a la grabación en YouTube en menos de 30 segundos desde que abre la sección»). Complementa **SC-008** (revisión con **lector de pantalla real**).

**Estado: PENDIENTE — se ejecuta DESPUÉS del despliegue** (así lo sitúa la spec, SC-008/SC-012; corrección B3 de la fase 5: aquí **no se dan por hechos**). Responsable: el equipo; la sesión la **observa `qa-tester` junto con el humano** y el informe **lo firma el humano**. No bloquea abrir el PR, pero **sí el cierre de la funcionalidad**, igual que en F4 (SC-010), F6 y F7 (SC-012).

Este protocolo es una prueba con el producto terminado y real, no de laboratorio:

- Con el sitio **ya desplegado** (o, como preparación, con el stack local `make up`) y contenido **real publicado**: el equipo pulsa **«Buscar episodios nuevos»** en `/panel/medios` y comprueba el resumen (con los feeds del 2026-10-17 la referencia es **10 fichas**: 8 unidas y 2 con un solo enlace).
- Dejar visible: al menos un episodio **con los dos enlaces** («Ver en YouTube» + «Escuchar en Spotify»), uno **con un solo formato**, uno con **descripción corta** y otro **sin descripción** (para ver que no queda un hueco), un episodio **oculto** que **no debe verse**, y —si los feeds lo permiten— uno con la marca **«Ya no aparece en el feed»** (que **sí** debe verse).
- En el dispositivo que la persona use a diario y, si se puede, una segunda ronda en el otro extremo (computadora o teléfono).
- En español; el inglés solo si la persona quiere cambiarlo por su cuenta (no se fuerza).

## Muestra

- **≥ 6 personas**, **2 por franja**: **18–35**, **36–59** y **60+**, con poca experiencia en internet incluida («adultos mucho más mayores que saben lo básico de internet», `idea.md` §7).
- Idealmente 1–2 que nunca hayan usado el panel (solo visitantes).

## Guion de la sesión (~10 min por persona)

1. Dejar abierta la portada del sitio real. **Sin dar contexto**: «Bienvenido/a; te pido que uses el sitio como lo harías normalmente».
2. **Cronómetro**: arranca cuando la persona abre `/medios` (inicio de la tarea 1) y se detiene cuando la grabación se abre en la pestaña nueva (final de la tarea 3). Se anota también el parcial de la tarea 1.
3. Tareas en este orden, **sin decir dónde está nada**:
   1. **Encontrar una prédica**: «¿Hay prédicas o podcasts de la iglesia en este sitio? ¿Encontrá uno?» — solo cuenta como éxito si llega **por su cuenta** (menú, bloque de la portada o pie) y **elige una ficha** sin que se le indique.
   2. **Saber que el enlace sale del sitio, ANTES de tocarlo**: «¿Qué crees que pasa si tocás “Ver en YouTube”?» — solo cuenta como éxito si dice, **antes de activarlo**, que lo va a sacar del sitio / va a YouTube o Spotify / se abre en **otra pestaña**. Si lo activa primero y solo entonces se da cuenta, se anota como **no logrado**.
   3. **Abrir la grabación y volver**: «Abrila.» — se comprueba que la grabación se abre en una **pestaña nueva**, que la pestaña del sitio **sigue abierta** y que la persona **vuelve al sitio** sin perderse. Aquí se cierra el cronómetro de **SC-001 (< 30 s)**; si se frustra, detener a los 60 s y anotar «no alcanzado».
4. Libre (solo si la persona **quiere**): cambiar el idioma y comprobar que los nombres, fechas y enlaces **no cambian** y que la descripción sin traducción aparece en español.
5. Si la persona usa **lector de pantalla** o navegación por teclado: recorrer `/medios`, las fichas y los enlaces externos, y el bloque de la portada (esto complementa **SC-008**).

## SC-008 — revisión con lector de pantalla real (aparte de las personas)

Se hace con **VoiceOver, NVDA o TalkBack**, en **tema claro y tema oscuro**, sobre `/medios` y sobre el bloque de la portada, con la checklist de [`quickstart.md`](./quickstart.md) §13:

- [ ] 320 px / 768 px / 1280 px: sin desplazamiento horizontal y con la ficha completa (carátula, nombre, descripción y enlaces).
- [ ] Recorrido **solo con teclado**: foco visible en cada paso, sin trampas de foco, orden comprensible.
- [ ] Los enlaces externos **se anuncian** (a qué servicio llevan, que se sale del sitio y que se abre en una pestaña nueva) también para el lector de pantalla.
- [ ] Contraste suficiente en claro y oscuro; texto ampliado al 200 % sin perder información.
- [ ] Movimiento de F10 respetando `prefers-reduced-motion`.

Referencia: **WCAG 2.1 nivel AA** (default revisable de la spec). **0 hallazgos críticos.**

## Qué registrar (una fila por persona y tarea)

| Dato | Cómo se anota |
|---|---|
| Persona: franja y experiencia con internet | 18–35 / 36–59 / 60+ · poca experiencia sí/no · dispositivo |
| Tarea 1: encuentra una prédica sin ayuda | sí/no y **de dónde** (menú, bloque de la portada, pie) |
| Tarea 1: parcial (hasta identificar la ficha) | **segundos** |
| Tarea 2: **sabe que sale del sitio ANTES de tocar el enlace** | sí/no · textual de la persona |
| Tarea 3: abre la grabación y vuelve al sitio | sí/no · ¿la pestaña del sitio siguió abierta? |
| **Tarea 3: tiempo total (SC-001)** | **segundos** (objetivo: < 30 s); «no alcanzado» a los 60 s |
| Ayudas dadas por quien modera | nº de pistas (0 = éxito puro) |
| Sorpresas, confusión o desconfianza | textual (especial: si denuncia ver un episodio **oculto** o una ficha con «posible duplicado» sin resolver → **defecto**, avisar a QA al instante) |
| SC-008 (si aplica a la persona) | recorrido por teclado/lector: hallazgos y dónde |

## Criterio de aceptación (SC-012 / SC-001 / SC-008)

- **≥ 90 %** de las personas (redondeando hacia arriba: con 6 personas, **las 6**) **encuentra una prédica y accede a la grabación sin ayuda**.
- **≥ 90 %** identifica **antes de activar el enlace** que el sitio lo lleva a YouTube/Spotify **en una pestaña nueva**.
- El **tiempo mediano hasta abrir la grabación es < 30 s** (SC-001), con **0** enlaces equivocados (cada ficha lleva a su video/episodio correcto).
- **0 hallazgos críticos** de accesibilidad con **lector de pantalla real** y recorrido completo por teclado (SC-008), en tema claro y oscuro.
- **0** episodios ocultos ni «posibles duplicados» sin resolver vistos por las personas (SC-002).

## Cómo se registra y cuándo

- **Cuándo**: **después del despliegue a producción** (y solo entonces se da el criterio por cumplido). Como preparación, puede repetirse el guion en local con el stack levantado, pero **el informe de cierre es el de producción**.
- **Quién**: el equipo con personas reales; **observado por `qa-tester` junto con el humano**.
- **Dónde**: el informe (tabla con los tiempos reales y las firmas) se añade **al final de este archivo** o se registra en `estado.md` como cierre de SC-008/SC-012.
- **Firma**: **el humano** aprueba el resultado; sin esa firma, los dos criterios siguen abiertos.

## Cómo tratar los hallazgos

1. Lo que no sea un fallo de la spec (p. ej., «no encontré el menú») va a la **lista de hallazgos** de la revisión de QA con severidad y, si aplica, ajuste menor de UX (sin cambiar la spec).
2. Si el hallazgo **toca el comportamiento descrito en la spec** (se ve un episodio oculto, un duplicado sin resolver, o un enlace roto presentado como disponible → privacidad/seguridad), se convierte en **defecto bloqueante** con pasos de reproducción y se escala al orquestador.

---

*Preparado por `documentador`, F9, 2026-10-18 (T944/T945). Formato del precedente de F6/F7: `specs/006-ministerios/pruebas-usabilidad-SC-012.md` y `specs/007-donaciones/pruebas-usabilidad-SC-012.md`.*
